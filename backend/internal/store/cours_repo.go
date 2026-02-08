// Package store contient les accès à la base de données
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Cours représente un cours scanné avec son texte OCR
type Cours struct {
	ID                 string            `json:"id"`
	Titre              string            `json:"titre"`
	Matiere            string            `json:"matiere,omitempty"`
	TexteOCR           string            `json:"texteOCR"`
	TexteCorrige       string            `json:"texteCorrige,omitempty"`
	Confiance          float64           `json:"confiance"`
	ZonesIncertaines   []ZoneIncertaine  `json:"zonesIncertaines"`
	FichiersOriginaux  []string          `json:"fichiersOriginaux"`
	Images             []string          `json:"images"`
	DateCreation       time.Time         `json:"dateCreation"`
	DateModification   time.Time         `json:"dateModification"`
}

// ZoneIncertaine représente une zone du texte OCR avec faible confiance
type ZoneIncertaine struct {
	Debut  int    `json:"debut"`
	Fin    int    `json:"fin"`
	Texte  string `json:"texte"`
	Raison string `json:"raison"`
}

// CoursRepository définit les opérations CRUD pour les cours
type CoursRepository interface {
	Creer(ctx context.Context, cours *Cours) error
	ObtenirParID(ctx context.Context, id string) (*Cours, error)
	Lister(ctx context.Context, limite, offset int) ([]*Cours, error)
	MettreAJour(ctx context.Context, cours *Cours) error
	Supprimer(ctx context.Context, id string) error
	Compter(ctx context.Context) (int, error)
}

// CoursRepo implémente CoursRepository avec PostgreSQL
type CoursRepo struct {
	db *sql.DB
}

// NouveauCoursRepo crée un nouveau repository pour les cours
func NouveauCoursRepo(store *Store) *CoursRepo {
	return &CoursRepo{db: store.DB()}
}

// Vérification que CoursRepo implémente CoursRepository
var _ CoursRepository = (*CoursRepo)(nil)

// Creer insère un nouveau cours dans la base de données
func (r *CoursRepo) Creer(ctx context.Context, cours *Cours) error {
	if cours.ID == "" {
		cours.ID = uuid.New().String()
	}

	now := time.Now()
	cours.DateCreation = now
	cours.DateModification = now

	zonesJSON, err := json.Marshal(cours.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(cours.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	imagesJSON, err := json.Marshal(cours.Images)
	if err != nil {
		return fmt.Errorf("erreur sérialisation images: %w", err)
	}

	query := `
		INSERT INTO cours (id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, images, date_creation, date_modification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		cours.ID,
		cours.Titre,
		nullString(cours.Matiere),
		cours.TexteOCR,
		nullString(cours.TexteCorrige),
		cours.Confiance,
		zonesJSON,
		fichiersJSON,
		imagesJSON,
		cours.DateCreation,
		cours.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur création cours: %w", err)
	}

	return nil
}

// ObtenirParID récupère un cours par son identifiant
func (r *CoursRepo) ObtenirParID(ctx context.Context, id string) (*Cours, error) {
	query := `
		SELECT id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, COALESCE(images, '[]'::jsonb), date_creation, date_modification
		FROM cours
		WHERE id = $1
	`

	cours := &Cours{}
	var matiere, texteCorrige sql.NullString
	var zonesJSON, fichiersJSON, imagesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cours.ID,
		&cours.Titre,
		&matiere,
		&cours.TexteOCR,
		&texteCorrige,
		&cours.Confiance,
		&zonesJSON,
		&fichiersJSON,
		&imagesJSON,
		&cours.DateCreation,
		&cours.DateModification,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cours non trouvé: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération cours: %w", err)
	}

	cours.Matiere = matiere.String
	cours.TexteCorrige = texteCorrige.String

	if err := json.Unmarshal(zonesJSON, &cours.ZonesIncertaines); err != nil {
		return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
	}

	if err := json.Unmarshal(fichiersJSON, &cours.FichiersOriginaux); err != nil {
		return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &cours.Images); err != nil {
		return nil, fmt.Errorf("erreur désérialisation images: %w", err)
	}

	return cours, nil
}

// Lister récupère une liste de cours avec pagination
func (r *CoursRepo) Lister(ctx context.Context, limite, offset int) ([]*Cours, error) {
	query := `
		SELECT id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, COALESCE(images, '[]'::jsonb), date_creation, date_modification
		FROM cours
		ORDER BY date_creation DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limite, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur liste cours: %w", err)
	}
	defer rows.Close()

	var coursList []*Cours
	for rows.Next() {
		cours := &Cours{}
		var matiere, texteCorrige sql.NullString
		var zonesJSON, fichiersJSON, imagesJSON []byte

		err := rows.Scan(
			&cours.ID,
			&cours.Titre,
			&matiere,
			&cours.TexteOCR,
			&texteCorrige,
			&cours.Confiance,
			&zonesJSON,
			&fichiersJSON,
			&imagesJSON,
			&cours.DateCreation,
			&cours.DateModification,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan cours: %w", err)
		}

		cours.Matiere = matiere.String
		cours.TexteCorrige = texteCorrige.String

		if err := json.Unmarshal(zonesJSON, &cours.ZonesIncertaines); err != nil {
			return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
		}

		if err := json.Unmarshal(fichiersJSON, &cours.FichiersOriginaux); err != nil {
			return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
		}

		if err := json.Unmarshal(imagesJSON, &cours.Images); err != nil {
			return nil, fmt.Errorf("erreur désérialisation images: %w", err)
		}

		coursList = append(coursList, cours)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération cours: %w", err)
	}

	return coursList, nil
}

// MettreAJour met à jour un cours existant
func (r *CoursRepo) MettreAJour(ctx context.Context, cours *Cours) error {
	cours.DateModification = time.Now()

	zonesJSON, err := json.Marshal(cours.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(cours.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	imagesJSON, err := json.Marshal(cours.Images)
	if err != nil {
		return fmt.Errorf("erreur sérialisation images: %w", err)
	}

	query := `
		UPDATE cours
		SET titre = $2, matiere = $3, texte_ocr = $4, texte_corrige = $5, confiance = $6, zones_incertaines = $7, fichiers_originaux = $8, images = $9, date_modification = $10
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		cours.ID,
		cours.Titre,
		nullString(cours.Matiere),
		cours.TexteOCR,
		nullString(cours.TexteCorrige),
		cours.Confiance,
		zonesJSON,
		fichiersJSON,
		imagesJSON,
		cours.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour cours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cours non trouvé: %s", cours.ID)
	}

	return nil
}

// Supprimer supprime un cours par son identifiant
func (r *CoursRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM cours WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression cours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cours non trouvé: %s", id)
	}

	return nil
}

// Compter retourne le nombre total de cours
func (r *CoursRepo) Compter(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM cours`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage cours: %w", err)
	}

	return count, nil
}

// nullString convertit une chaîne en sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
