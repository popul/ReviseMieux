package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/validation"
)

var _ validation.Repository = (*ValidationRepository)(nil)

// ValidationRepository implements validation.Repository using PostgreSQL.
type ValidationRepository struct {
	pool *pgxpool.Pool
}

// NewValidationRepository creates a new ValidationRepository.
func NewValidationRepository(pool *pgxpool.Pool) *ValidationRepository {
	return &ValidationRepository{pool: pool}
}

func (r *ValidationRepository) FindByID(ctx context.Context, id uuid.UUID) (*validation.ValidationTask, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, item_id, crop_url, suggestion, priority,
		       status, resolved_by, source, student_note, created_at, updated_at
		FROM validation_tasks WHERE id = $1`, id)

	t := &validation.ValidationTask{}
	if err := row.Scan(
		&t.ID, &t.ItemID, &t.CropURL, &t.Suggestion, &t.Priority,
		&t.Status, &t.ResolvedBy, &t.Source, &t.StudentNote,
		&t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("validation_repo.FindByID: %w", validation.ErrNotFound)
	}
	return t, nil
}

func (r *ValidationRepository) FindPendingByItem(ctx context.Context, itemID uuid.UUID) ([]*validation.ValidationTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, item_id, crop_url, suggestion, priority,
		       status, resolved_by, source, student_note, created_at, updated_at
		FROM validation_tasks WHERE item_id = $1 AND status = 'PENDING'
		ORDER BY priority DESC, created_at`, itemID)
	if err != nil {
		return nil, fmt.Errorf("validation_repo.FindPendingByItem: %w", err)
	}
	defer rows.Close()

	var result []*validation.ValidationTask
	for rows.Next() {
		t := &validation.ValidationTask{}
		if err := rows.Scan(
			&t.ID, &t.ItemID, &t.CropURL, &t.Suggestion, &t.Priority,
			&t.Status, &t.ResolvedBy, &t.Source, &t.StudentNote,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("validation_repo.FindPendingByItem: scan: %w", err)
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *ValidationRepository) FindPendingAll(ctx context.Context, limit int) ([]*validation.ValidationTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, item_id, crop_url, suggestion, priority,
		       status, resolved_by, source, student_note, created_at, updated_at
		FROM validation_tasks WHERE status = 'PENDING'
		ORDER BY priority DESC, created_at
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("validation_repo.FindPendingAll: %w", err)
	}
	defer rows.Close()

	var result []*validation.ValidationTask
	for rows.Next() {
		t := &validation.ValidationTask{}
		if err := rows.Scan(
			&t.ID, &t.ItemID, &t.CropURL, &t.Suggestion, &t.Priority,
			&t.Status, &t.ResolvedBy, &t.Source, &t.StudentNote,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("validation_repo.FindPendingAll: scan: %w", err)
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *ValidationRepository) Save(ctx context.Context, t *validation.ValidationTask) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO validation_tasks (id, item_id, crop_url, suggestion, priority,
		                              status, resolved_by, source, student_note,
		                              created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			resolved_by = EXCLUDED.resolved_by,
			student_note = EXCLUDED.student_note,
			updated_at = EXCLUDED.updated_at`,
		t.ID, t.ItemID, t.CropURL, t.Suggestion, t.Priority,
		t.Status, t.ResolvedBy, t.Source, t.StudentNote,
		t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("validation_repo.Save: %w", err)
	}
	return nil
}
