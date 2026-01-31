// Package store contient les repositories pour l'accès aux données
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NouveauRessourcesRepo crée un nouveau repository pour les ressources
func NouveauRessourcesRepo(store *Store) *RessourcesRepo {
	return &RessourcesRepo{db: store.db}
}

// TypeRessource définit les types de ressources possibles
type TypeRessource string

const (
	TypeRessourceVideo   TypeRessource = "video"
	TypeRessourceArticle TypeRessource = "article"
	TypeRessourceSite    TypeRessource = "site"
)

// Ressource représente une ressource suggérée liée à un cours
type Ressource struct {
	ID           string        `json:"id"`
	CoursID      string        `json:"cours_id"`
	Titre        string        `json:"titre"`
	URL          string        `json:"url,omitempty"`
	Type         TypeRessource `json:"type"`
	Description  string        `json:"description,omitempty"`
	DateCreation time.Time     `json:"date_creation"`
}

// RessourcesRepository définit l'interface pour accéder aux ressources
type RessourcesRepository interface {
	// CreerPlusieurs crée plusieurs ressources en une transaction
	CreerPlusieurs(ctx context.Context, ressources []*Ressource) error
	// ListerParCours récupère toutes les ressources d'un cours
	ListerParCours(ctx context.Context, coursID string) ([]*Ressource, error)
	// Compter retourne le nombre total de ressources
	Compter(ctx context.Context) (int, error)
	// CompterParCours retourne le nombre de ressources d'un cours
	CompterParCours(ctx context.Context, coursID string) (int, error)
	// Supprimer supprime une ressource par son ID
	Supprimer(ctx context.Context, id string) error
	// SupprimerParCours supprime toutes les ressources d'un cours
	SupprimerParCours(ctx context.Context, coursID string) error
}

// RessourcesRepo implémente RessourcesRepository avec PostgreSQL
type RessourcesRepo struct {
	db *sql.DB
}

// Vérification que RessourcesRepo implémente RessourcesRepository
var _ RessourcesRepository = (*RessourcesRepo)(nil)

// CreerPlusieurs crée plusieurs ressources en une transaction
func (r *RessourcesRepo) CreerPlusieurs(ctx context.Context, ressources []*Ressource) error {
	if len(ressources) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO ressources (id, cours_id, titre, url, type, description, date_creation)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur préparation requête: %w", err)
	}
	defer stmt.Close()

	for _, ressource := range ressources {
		// Générer un ID si non fourni
		if ressource.ID == "" {
			ressource.ID = uuid.New().String()
		}
		// Définir la date de création si non fournie
		if ressource.DateCreation.IsZero() {
			ressource.DateCreation = time.Now()
		}

		_, err := stmt.ExecContext(ctx,
			ressource.ID,
			ressource.CoursID,
			ressource.Titre,
			nullString(ressource.URL),
			ressource.Type,
			nullString(ressource.Description),
			ressource.DateCreation,
		)
		if err != nil {
			return fmt.Errorf("erreur insertion ressource: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit transaction: %w", err)
	}

	return nil
}

// ListerParCours récupère toutes les ressources d'un cours
func (r *RessourcesRepo) ListerParCours(ctx context.Context, coursID string) ([]*Ressource, error) {
	query := `
		SELECT id, cours_id, titre, url, type, description, date_creation
		FROM ressources
		WHERE cours_id = $1
		ORDER BY date_creation ASC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête ressources: %w", err)
	}
	defer rows.Close()

	var ressources []*Ressource
	for rows.Next() {
		var res Ressource
		var url, description sql.NullString

		err := rows.Scan(
			&res.ID,
			&res.CoursID,
			&res.Titre,
			&url,
			&res.Type,
			&description,
			&res.DateCreation,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan ressource: %w", err)
		}

		if url.Valid {
			res.URL = url.String
		}
		if description.Valid {
			res.Description = description.String
		}

		ressources = append(ressources, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération ressources: %w", err)
	}

	return ressources, nil
}

// Compter retourne le nombre total de ressources
func (r *RessourcesRepo) Compter(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ressources").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage ressources: %w", err)
	}
	return count, nil
}

// CompterParCours retourne le nombre de ressources d'un cours
func (r *RessourcesRepo) CompterParCours(ctx context.Context, coursID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ressources WHERE cours_id = $1", coursID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage ressources cours: %w", err)
	}
	return count, nil
}

// Supprimer supprime une ressource par son ID
func (r *RessourcesRepo) Supprimer(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM ressources WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("erreur suppression ressource: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("ressource non trouvée")
	}

	return nil
}

// SupprimerParCours supprime toutes les ressources d'un cours
func (r *RessourcesRepo) SupprimerParCours(ctx context.Context, coursID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM ressources WHERE cours_id = $1", coursID)
	if err != nil {
		return fmt.Errorf("erreur suppression ressources cours: %w", err)
	}
	return nil
}
