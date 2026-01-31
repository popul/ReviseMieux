// Package store contient les accès à la base de données
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Fiche représente une fiche de révision question/réponse
type Fiche struct {
	ID           string    `json:"id"`
	CoursID      string    `json:"coursId"`
	Question     string    `json:"question"`
	Reponse      string    `json:"reponse"`
	Difficulte   string    `json:"difficulte"` // "facile", "moyen", "difficile"
	Ordre        int       `json:"ordre"`
	DateCreation time.Time `json:"dateCreation"`
}

// FichesRepository définit les opérations pour les fiches de révision
type FichesRepository interface {
	CreerPlusieurs(ctx context.Context, fiches []*Fiche) error
	ListerParCours(ctx context.Context, coursID string) ([]*Fiche, error)
	Compter(ctx context.Context) (int, error)
	CompterParCours(ctx context.Context, coursID string) (int, error)
	Supprimer(ctx context.Context, id string) error
	SupprimerParCours(ctx context.Context, coursID string) error
}

// FichesRepo implémente FichesRepository avec PostgreSQL
type FichesRepo struct {
	db *sql.DB
}

// NouveauFichesRepo crée un nouveau repository pour les fiches
func NouveauFichesRepo(store *Store) *FichesRepo {
	return &FichesRepo{db: store.DB()}
}

// Vérification que FichesRepo implémente FichesRepository
var _ FichesRepository = (*FichesRepo)(nil)

// CreerPlusieurs insère plusieurs fiches dans une transaction
func (r *FichesRepo) CreerPlusieurs(ctx context.Context, fiches []*Fiche) error {
	if len(fiches) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO fiches (id, cours_id, question, reponse, difficulte, ordre, date_creation)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur préparation requête: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for i, fiche := range fiches {
		if fiche.ID == "" {
			fiche.ID = uuid.New().String()
		}
		fiche.DateCreation = now
		if fiche.Ordre == 0 {
			fiche.Ordre = i + 1
		}

		_, err := stmt.ExecContext(ctx,
			fiche.ID,
			fiche.CoursID,
			fiche.Question,
			fiche.Reponse,
			fiche.Difficulte,
			fiche.Ordre,
			fiche.DateCreation,
		)
		if err != nil {
			return fmt.Errorf("erreur insertion fiche %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit transaction: %w", err)
	}

	return nil
}

// ListerParCours récupère toutes les fiches d'un cours
func (r *FichesRepo) ListerParCours(ctx context.Context, coursID string) ([]*Fiche, error) {
	query := `
		SELECT id, cours_id, question, reponse, difficulte, ordre, date_creation
		FROM fiches
		WHERE cours_id = $1
		ORDER BY ordre ASC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête fiches: %w", err)
	}
	defer rows.Close()

	var fiches []*Fiche
	for rows.Next() {
		fiche := &Fiche{}
		err := rows.Scan(
			&fiche.ID,
			&fiche.CoursID,
			&fiche.Question,
			&fiche.Reponse,
			&fiche.Difficulte,
			&fiche.Ordre,
			&fiche.DateCreation,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan fiche: %w", err)
		}
		fiches = append(fiches, fiche)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération fiches: %w", err)
	}

	return fiches, nil
}

// Compter retourne le nombre total de fiches
func (r *FichesRepo) Compter(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM fiches`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage fiches: %w", err)
	}

	return count, nil
}

// CompterParCours retourne le nombre de fiches pour un cours
func (r *FichesRepo) CompterParCours(ctx context.Context, coursID string) (int, error) {
	query := `SELECT COUNT(*) FROM fiches WHERE cours_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, coursID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage fiches cours: %w", err)
	}

	return count, nil
}

// Supprimer supprime une fiche par son ID
func (r *FichesRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM fiches WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression fiche: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("fiche non trouvée: %s", id)
	}

	return nil
}

// SupprimerParCours supprime toutes les fiches d'un cours
func (r *FichesRepo) SupprimerParCours(ctx context.Context, coursID string) error {
	query := `DELETE FROM fiches WHERE cours_id = $1`

	_, err := r.db.ExecContext(ctx, query, coursID)
	if err != nil {
		return fmt.Errorf("erreur suppression fiches cours: %w", err)
	}

	return nil
}
