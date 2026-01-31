// Package store contient les accès à la base de données
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Migration représente une migration de base de données
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

// MigrationRecord représente un enregistrement dans schema_migrations
type MigrationRecord struct {
	Version   int
	Name      string
	AppliedAt time.Time
}

// Migrer exécute toutes les migrations en attente
func (s *Store) Migrer(migrationsDir string) error {
	// Créer la table schema_migrations si elle n'existe pas
	if err := s.creerTableMigrations(); err != nil {
		return fmt.Errorf("erreur création table migrations: %w", err)
	}

	// Charger les migrations depuis le dossier
	migrations, err := chargerMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("erreur chargement migrations: %w", err)
	}

	// Récupérer les migrations déjà appliquées
	appliquees, err := s.migrationsAppliquees()
	if err != nil {
		return fmt.Errorf("erreur récupération migrations appliquées: %w", err)
	}

	// Appliquer les migrations en attente
	for _, m := range migrations {
		if appliquees[m.Version] {
			continue
		}

		if err := s.appliquerMigration(m); err != nil {
			return fmt.Errorf("erreur migration %d (%s): %w", m.Version, m.Name, err)
		}

		fmt.Printf("✓ Migration %03d_%s appliquée\n", m.Version, m.Name)
	}

	return nil
}

// Rollback annule la dernière migration
func (s *Store) Rollback(migrationsDir string) error {
	// Récupérer la dernière migration appliquée
	var version int
	var name string
	err := s.db.QueryRow(`
		SELECT version, name FROM schema_migrations
		ORDER BY version DESC LIMIT 1
	`).Scan(&version, &name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("aucune migration à annuler")
	}
	if err != nil {
		return fmt.Errorf("erreur récupération dernière migration: %w", err)
	}

	// Charger la migration correspondante
	migrations, err := chargerMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("erreur chargement migrations: %w", err)
	}

	var migration *Migration
	for _, m := range migrations {
		if m.Version == version {
			migration = &m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("fichier migration %d introuvable", version)
	}

	// Exécuter le rollback
	if err := s.annulerMigration(*migration); err != nil {
		return fmt.Errorf("erreur rollback migration %d: %w", version, err)
	}

	fmt.Printf("✓ Migration %03d_%s annulée\n", version, name)
	return nil
}

// Statut affiche le statut des migrations
func (s *Store) StatutMigrations(migrationsDir string) ([]MigrationRecord, []Migration, error) {
	// Créer la table si elle n'existe pas
	if err := s.creerTableMigrations(); err != nil {
		return nil, nil, err
	}

	// Récupérer les migrations appliquées
	rows, err := s.db.Query(`
		SELECT version, name, applied_at FROM schema_migrations ORDER BY version
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var appliquees []MigrationRecord
	appliqueeMap := make(map[int]bool)
	for rows.Next() {
		var r MigrationRecord
		if err := rows.Scan(&r.Version, &r.Name, &r.AppliedAt); err != nil {
			return nil, nil, err
		}
		appliquees = append(appliquees, r)
		appliqueeMap[r.Version] = true
	}

	// Charger toutes les migrations
	migrations, err := chargerMigrations(migrationsDir)
	if err != nil {
		return nil, nil, err
	}

	// Filtrer les migrations en attente
	var enAttente []Migration
	for _, m := range migrations {
		if !appliqueeMap[m.Version] {
			enAttente = append(enAttente, m)
		}
	}

	return appliquees, enAttente, nil
}

// creerTableMigrations crée la table de suivi des migrations
func (s *Store) creerTableMigrations() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	return err
}

// migrationsAppliquees retourne un map des versions déjà appliquées
func (s *Store) migrationsAppliquees() (map[int]bool, error) {
	rows, err := s.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliquees := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		appliquees[version] = true
	}

	return appliquees, rows.Err()
}

// appliquerMigration exécute une migration up dans une transaction
func (s *Store) appliquerMigration(m Migration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Exécuter le SQL de la migration
	if _, err := tx.Exec(m.UpSQL); err != nil {
		return fmt.Errorf("erreur SQL: %w", err)
	}

	// Enregistrer la migration
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
		m.Version, m.Name,
	)
	if err != nil {
		return fmt.Errorf("erreur enregistrement: %w", err)
	}

	return tx.Commit()
}

// annulerMigration exécute une migration down dans une transaction
func (s *Store) annulerMigration(m Migration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Exécuter le SQL de rollback
	if _, err := tx.Exec(m.DownSQL); err != nil {
		return fmt.Errorf("erreur SQL: %w", err)
	}

	// Supprimer l'enregistrement de la migration
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", m.Version)
	if err != nil {
		return fmt.Errorf("erreur suppression enregistrement: %w", err)
	}

	return tx.Commit()
}

// chargerMigrations lit tous les fichiers de migration depuis un dossier
func chargerMigrations(dir string) ([]Migration, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture dossier %s: %w", dir, err)
	}

	// Regex pour extraire version et nom: 001_create_cours.up.sql
	re := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

	migrationMap := make(map[int]*Migration)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(f.Name())
		if matches == nil {
			continue
		}

		var version int
		fmt.Sscanf(matches[1], "%d", &version)
		name := matches[2]
		direction := matches[3]

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("erreur lecture %s: %w", f.Name(), err)
		}

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		if direction == "up" {
			migrationMap[version].UpSQL = string(content)
		} else {
			migrationMap[version].DownSQL = string(content)
		}
	}

	// Convertir en slice et trier par version
	var migrations []Migration
	for _, m := range migrationMap {
		if m.UpSQL == "" {
			return nil, fmt.Errorf("migration %d: fichier .up.sql manquant", m.Version)
		}
		if m.DownSQL == "" {
			return nil, fmt.Errorf("migration %d: fichier .down.sql manquant", m.Version)
		}
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// ChargerMigrations est une fonction publique pour les tests
func ChargerMigrations(dir string) ([]Migration, error) {
	return chargerMigrations(dir)
}

// NettoyerNomMigration nettoie le nom pour l'affichage
func NettoyerNomMigration(name string) string {
	return strings.ReplaceAll(name, "_", " ")
}
