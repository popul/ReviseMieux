// Package store contient les acces a la base de donnees
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TermeLexique represente un terme du lexique extrait d un cours
type TermeLexique struct {
	ID         string    `json:"id"`
	CoursID    string    `json:"coursId"`
	Terme      string    `json:"terme"`
	Definition string    `json:"definition"`
	Contexte   string    `json:"contexte,omitempty"`
	Exemple    string    `json:"exemple,omitempty"`
	Categorie  string    `json:"categorie,omitempty"`
	Maitrise   int       `json:"maitrise"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// LexiqueRepository definit les operations pour les termes du lexique
type LexiqueRepository interface {
	Creer(ctx context.Context, terme *TermeLexique) error
	CreerPlusieurs(ctx context.Context, termes []*TermeLexique) error
	ListerParCours(ctx context.Context, coursID string) ([]*TermeLexique, error)
	ObtenirParID(ctx context.Context, id string) (*TermeLexique, error)
	MettreAJour(ctx context.Context, terme *TermeLexique) error
	MettreAJourMaitrise(ctx context.Context, id string, maitrise int) error
	SupprimerParCours(ctx context.Context, coursID string) error
}

// LexiqueRepo implemente LexiqueRepository avec PostgreSQL
type LexiqueRepo struct {
	db *sql.DB
}

// NouveauLexiqueRepo cree un nouveau repository pour le lexique
func NouveauLexiqueRepo(store *Store) *LexiqueRepo {
	return &LexiqueRepo{db: store.DB()}
}

// Verification que LexiqueRepo implemente LexiqueRepository
var _ LexiqueRepository = (*LexiqueRepo)(nil)

// Creer insere un nouveau terme dans la base de donnees
func (r *LexiqueRepo) Creer(ctx context.Context, terme *TermeLexique) error {
	if terme.ID == "" {
		terme.ID = uuid.New().String()
	}

	now := time.Now()
	terme.CreatedAt = now
	terme.UpdatedAt = now

	query := `
		INSERT INTO termes_lexique (id, cours_id, terme, definition, contexte, exemple, categorie, maitrise, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		terme.ID,
		terme.CoursID,
		terme.Terme,
		terme.Definition,
		nullableString(terme.Contexte),
		nullableString(terme.Exemple),
		nullableString(terme.Categorie),
		terme.Maitrise,
		terme.CreatedAt,
		terme.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("erreur creation terme lexique: %w", err)
	}

	return nil
}

// CreerPlusieurs insere plusieurs termes dans une transaction
func (r *LexiqueRepo) CreerPlusieurs(ctx context.Context, termes []*TermeLexique) error {
	if len(termes) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erreur debut transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO termes_lexique (id, cours_id, terme, definition, contexte, exemple, categorie, maitrise, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur preparation requete: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for i, terme := range termes {
		if terme.ID == "" {
			terme.ID = uuid.New().String()
		}
		terme.CreatedAt = now
		terme.UpdatedAt = now

		_, err = stmt.ExecContext(ctx,
			terme.ID,
			terme.CoursID,
			terme.Terme,
			terme.Definition,
			nullableString(terme.Contexte),
			nullableString(terme.Exemple),
			nullableString(terme.Categorie),
			terme.Maitrise,
			terme.CreatedAt,
			terme.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("erreur insertion terme %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit transaction: %w", err)
	}

	return nil
}

// ListerParCours recupere tous les termes d un cours
func (r *LexiqueRepo) ListerParCours(ctx context.Context, coursID string) ([]*TermeLexique, error) {
	query := `
		SELECT id, cours_id, terme, definition, COALESCE(contexte, ''), COALESCE(exemple, ''), COALESCE(categorie, ''), maitrise, created_at, updated_at
		FROM termes_lexique
		WHERE cours_id = $1
		ORDER BY terme ASC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requete termes lexique: %w", err)
	}
	defer rows.Close()

	var termes []*TermeLexique
	for rows.Next() {
		terme := &TermeLexique{}

		err := rows.Scan(
			&terme.ID,
			&terme.CoursID,
			&terme.Terme,
			&terme.Definition,
			&terme.Contexte,
			&terme.Exemple,
			&terme.Categorie,
			&terme.Maitrise,
			&terme.CreatedAt,
			&terme.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan terme lexique: %w", err)
		}

		termes = append(termes, terme)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur iteration termes lexique: %w", err)
	}

	return termes, nil
}

// ObtenirParID recupere un terme par son identifiant
func (r *LexiqueRepo) ObtenirParID(ctx context.Context, id string) (*TermeLexique, error) {
	query := `
		SELECT id, cours_id, terme, definition, COALESCE(contexte, ''), COALESCE(exemple, ''), COALESCE(categorie, ''), maitrise, created_at, updated_at
		FROM termes_lexique
		WHERE id = $1
	`

	terme := &TermeLexique{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&terme.ID,
		&terme.CoursID,
		&terme.Terme,
		&terme.Definition,
		&terme.Contexte,
		&terme.Exemple,
		&terme.Categorie,
		&terme.Maitrise,
		&terme.CreatedAt,
		&terme.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("terme lexique non trouve: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur recuperation terme lexique: %w", err)
	}

	return terme, nil
}

// MettreAJour met a jour un terme existant
func (r *LexiqueRepo) MettreAJour(ctx context.Context, terme *TermeLexique) error {
	terme.UpdatedAt = time.Now()

	query := `
		UPDATE termes_lexique
		SET terme = $1, definition = $2, contexte = $3, exemple = $4, categorie = $5, maitrise = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		terme.Terme,
		terme.Definition,
		nullableString(terme.Contexte),
		nullableString(terme.Exemple),
		nullableString(terme.Categorie),
		terme.Maitrise,
		terme.UpdatedAt,
		terme.ID,
	)
	if err != nil {
		return fmt.Errorf("erreur mise a jour terme lexique: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur verification mise a jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("terme lexique non trouve: %s", terme.ID)
	}

	return nil
}

// MettreAJourMaitrise met a jour uniquement le niveau de maitrise d un terme
func (r *LexiqueRepo) MettreAJourMaitrise(ctx context.Context, id string, maitrise int) error {
	if maitrise < 0 || maitrise > 5 {
		return fmt.Errorf("niveau de maitrise invalide: %d (doit etre entre 0 et 5)", maitrise)
	}

	query := `
		UPDATE termes_lexique
		SET maitrise = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, maitrise, time.Now(), id)
	if err != nil {
		return fmt.Errorf("erreur mise a jour maitrise: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur verification mise a jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("terme lexique non trouve: %s", id)
	}

	return nil
}

// SupprimerParCours supprime tous les termes d un cours
func (r *LexiqueRepo) SupprimerParCours(ctx context.Context, coursID string) error {
	query := `DELETE FROM termes_lexique WHERE cours_id = $1`

	_, err := r.db.ExecContext(ctx, query, coursID)
	if err != nil {
		return fmt.Errorf("erreur suppression termes lexique cours: %w", err)
	}

	return nil
}

// nullableString retourne nil si la chaine est vide, sinon la chaine
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
