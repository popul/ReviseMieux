// Package store contient les accès à la base de données
package store

import (
	"context"
	"database/sql"
	"time"
)

// QuotasJournaliers représente les quotas d'utilisation pour une journée
type QuotasJournaliers struct {
	ID          string
	Date        time.Time
	PagesOCR    int
	Generations int
}

// QuotasRepository définit les opérations sur les quotas
type QuotasRepository interface {
	// ObtenirOuCreer récupère les quotas du jour ou les crée s'ils n'existent pas
	ObtenirOuCreer(ctx context.Context, date time.Time) (*QuotasJournaliers, error)
	// IncrementerOCR incrémente le compteur OCR pour le jour donné
	IncrementerOCR(ctx context.Context, date time.Time, pages int) error
	// IncrementerGeneration incrémente le compteur de génération pour le jour donné
	IncrementerGeneration(ctx context.Context, date time.Time) error
}

// QuotasRepo implémente QuotasRepository avec PostgreSQL
type QuotasRepo struct {
	db *Store
}

// NouveauQuotasRepo crée un nouveau repository de quotas
func NouveauQuotasRepo(db *Store) *QuotasRepo {
	return &QuotasRepo{db: db}
}

// ObtenirOuCreer récupère ou crée les quotas du jour
func (r *QuotasRepo) ObtenirOuCreer(ctx context.Context, date time.Time) (*QuotasJournaliers, error) {
	dateJour := date.Truncate(24 * time.Hour)

	// Essayer d'abord de récupérer
	var quotas QuotasJournaliers
	err := r.db.DB().QueryRowContext(ctx, `
		SELECT id, date, pages_ocr, generations
		FROM quotas_journaliers
		WHERE date = $1
	`, dateJour).Scan(&quotas.ID, &quotas.Date, &quotas.PagesOCR, &quotas.Generations)

	if err == nil {
		return &quotas, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Créer si n'existe pas
	err = r.db.DB().QueryRowContext(ctx, `
		INSERT INTO quotas_journaliers (date, pages_ocr, generations)
		VALUES ($1, 0, 0)
		ON CONFLICT (date) DO UPDATE SET date = EXCLUDED.date
		RETURNING id, date, pages_ocr, generations
	`, dateJour).Scan(&quotas.ID, &quotas.Date, &quotas.PagesOCR, &quotas.Generations)

	if err != nil {
		return nil, err
	}

	return &quotas, nil
}

// IncrementerOCR incrémente le compteur de pages OCR
func (r *QuotasRepo) IncrementerOCR(ctx context.Context, date time.Time, pages int) error {
	dateJour := date.Truncate(24 * time.Hour)

	_, err := r.db.DB().ExecContext(ctx, `
		INSERT INTO quotas_journaliers (date, pages_ocr, generations)
		VALUES ($1, $2, 0)
		ON CONFLICT (date) DO UPDATE SET pages_ocr = quotas_journaliers.pages_ocr + $2
	`, dateJour, pages)

	return err
}

// IncrementerGeneration incrémente le compteur de générations
func (r *QuotasRepo) IncrementerGeneration(ctx context.Context, date time.Time) error {
	dateJour := date.Truncate(24 * time.Hour)

	_, err := r.db.DB().ExecContext(ctx, `
		INSERT INTO quotas_journaliers (date, pages_ocr, generations)
		VALUES ($1, 0, 1)
		ON CONFLICT (date) DO UPDATE SET generations = quotas_journaliers.generations + 1
	`, dateJour)

	return err
}
