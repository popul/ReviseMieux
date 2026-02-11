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

// PositionDansCours représente la position d'un concept dans le texte du cours
type PositionDansCours struct {
	Debut int `json:"debut,omitempty"`
	Fin   int `json:"fin,omitempty"`
}

// Concept représente un concept clé extrait d'un cours
type Concept struct {
	ID               string             `json:"id"`
	CoursID          string             `json:"coursId"`
	Nom              string             `json:"nom"`
	Definition       string             `json:"definition"`
	Importance       string             `json:"importance"` // "essentiel", "important", "secondaire"
	PositionDansCours *PositionDansCours `json:"positionDansCours,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

// ConceptsRepository définit les opérations pour les concepts
type ConceptsRepository interface {
	Creer(ctx context.Context, concept *Concept) error
	CreerPlusieurs(ctx context.Context, concepts []*Concept) error
	ListerParCours(ctx context.Context, coursID string) ([]*Concept, error)
	ObtenirParID(ctx context.Context, id string) (*Concept, error)
	MettreAJour(ctx context.Context, concept *Concept) error
	Supprimer(ctx context.Context, id string) error
	SupprimerParCours(ctx context.Context, coursID string) error
}

// ConceptsRepo implémente ConceptsRepository avec PostgreSQL
type ConceptsRepo struct {
	db *sql.DB
}

// NouveauConceptsRepo crée un nouveau repository pour les concepts
func NouveauConceptsRepo(store *Store) *ConceptsRepo {
	return &ConceptsRepo{db: store.DB()}
}

// Vérification que ConceptsRepo implémente ConceptsRepository
var _ ConceptsRepository = (*ConceptsRepo)(nil)

// Creer insère un nouveau concept dans la base de données
func (r *ConceptsRepo) Creer(ctx context.Context, concept *Concept) error {
	if concept.ID == "" {
		concept.ID = uuid.New().String()
	}

	now := time.Now()
	concept.CreatedAt = now
	concept.UpdatedAt = now

	positionJSON, err := json.Marshal(concept.PositionDansCours)
	if err != nil {
		return fmt.Errorf("erreur sérialisation position: %w", err)
	}

	query := `
		INSERT INTO concepts (id, cours_id, nom, definition, importance, position_dans_cours, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		concept.ID,
		concept.CoursID,
		concept.Nom,
		concept.Definition,
		concept.Importance,
		positionJSON,
		concept.CreatedAt,
		concept.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("erreur création concept: %w", err)
	}

	return nil
}

// CreerPlusieurs insère plusieurs concepts dans une transaction
func (r *ConceptsRepo) CreerPlusieurs(ctx context.Context, concepts []*Concept) error {
	if len(concepts) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO concepts (id, cours_id, nom, definition, importance, position_dans_cours, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur préparation requête: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for i, concept := range concepts {
		if concept.ID == "" {
			concept.ID = uuid.New().String()
		}
		concept.CreatedAt = now
		concept.UpdatedAt = now

		positionJSON, err := json.Marshal(concept.PositionDansCours)
		if err != nil {
			return fmt.Errorf("erreur sérialisation position concept %d: %w", i, err)
		}

		_, err = stmt.ExecContext(ctx,
			concept.ID,
			concept.CoursID,
			concept.Nom,
			concept.Definition,
			concept.Importance,
			positionJSON,
			concept.CreatedAt,
			concept.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("erreur insertion concept %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit transaction: %w", err)
	}

	return nil
}

// ListerParCours récupère tous les concepts d'un cours
func (r *ConceptsRepo) ListerParCours(ctx context.Context, coursID string) ([]*Concept, error) {
	query := `
		SELECT id, cours_id, nom, definition, importance, position_dans_cours, created_at, updated_at
		FROM concepts
		WHERE cours_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête concepts: %w", err)
	}
	defer rows.Close()

	var concepts []*Concept
	for rows.Next() {
		concept := &Concept{}
		var positionJSON []byte

		err := rows.Scan(
			&concept.ID,
			&concept.CoursID,
			&concept.Nom,
			&concept.Definition,
			&concept.Importance,
			&positionJSON,
			&concept.CreatedAt,
			&concept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan concept: %w", err)
		}

		if positionJSON != nil && string(positionJSON) != "null" {
			var position PositionDansCours
			if err := json.Unmarshal(positionJSON, &position); err == nil {
				concept.PositionDansCours = &position
			}
		}

		concepts = append(concepts, concept)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération concepts: %w", err)
	}

	return concepts, nil
}

// ObtenirParID récupère un concept par son identifiant
func (r *ConceptsRepo) ObtenirParID(ctx context.Context, id string) (*Concept, error) {
	query := `
		SELECT id, cours_id, nom, definition, importance, position_dans_cours, created_at, updated_at
		FROM concepts
		WHERE id = $1
	`

	concept := &Concept{}
	var positionJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&concept.ID,
		&concept.CoursID,
		&concept.Nom,
		&concept.Definition,
		&concept.Importance,
		&positionJSON,
		&concept.CreatedAt,
		&concept.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("concept non trouvé: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération concept: %w", err)
	}

	if positionJSON != nil && string(positionJSON) != "null" {
		var position PositionDansCours
		if err := json.Unmarshal(positionJSON, &position); err == nil {
			concept.PositionDansCours = &position
		}
	}

	return concept, nil
}

// MettreAJour met à jour un concept existant
func (r *ConceptsRepo) MettreAJour(ctx context.Context, concept *Concept) error {
	concept.UpdatedAt = time.Now()

	positionJSON, err := json.Marshal(concept.PositionDansCours)
	if err != nil {
		return fmt.Errorf("erreur sérialisation position: %w", err)
	}

	query := `
		UPDATE concepts
		SET nom = $2, definition = $3, importance = $4, position_dans_cours = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		concept.ID,
		concept.Nom,
		concept.Definition,
		concept.Importance,
		positionJSON,
		concept.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour concept: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("concept non trouvé: %s", concept.ID)
	}

	return nil
}

// Supprimer supprime un concept par son identifiant
func (r *ConceptsRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM concepts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression concept: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("concept non trouvé: %s", id)
	}

	return nil
}

// SupprimerParCours supprime tous les concepts d'un cours
func (r *ConceptsRepo) SupprimerParCours(ctx context.Context, coursID string) error {
	query := `DELETE FROM concepts WHERE cours_id = $1`

	_, err := r.db.ExecContext(ctx, query, coursID)
	if err != nil {
		return fmt.Errorf("erreur suppression concepts cours: %w", err)
	}

	return nil
}
