package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// Compile-time check that MasteryRepository implements mastery.Repository.
var _ mastery.Repository = (*MasteryRepository)(nil)

// MasteryRepository implements mastery.Repository using PostgreSQL.
type MasteryRepository struct {
	pool *pgxpool.Pool
}

// NewMasteryRepository creates a new MasteryRepository.
func NewMasteryRepository(pool *pgxpool.Pool) *MasteryRepository {
	return &MasteryRepository{pool: pool}
}

func (r *MasteryRepository) FindByID(ctx context.Context, id uuid.UUID) (*mastery.Mastery, error) {
	const q = `
		SELECT id, user_id, item_id, state, next_due_at, last_review_at,
		       last_success_at, consecutive_successes, consecutive_failures,
		       current_difficulty, capped_at_ok, created_at, updated_at
		FROM masteries
		WHERE id = $1`

	m := &mastery.Mastery{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.UserID, &m.ItemID, &m.State, &m.NextDueAt, &m.LastReviewAt,
		&m.LastSuccessAt, &m.ConsecutiveSuccesses, &m.ConsecutiveFailures,
		&m.CurrentDifficulty, &m.CappedAtOK, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("mastery.Repository.FindByID: %w", mastery.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("mastery.Repository.FindByID: %w", err)
	}
	return m, nil
}

func (r *MasteryRepository) FindByUserAndItem(ctx context.Context, userID, itemID uuid.UUID) (*mastery.Mastery, error) {
	const q = `
		SELECT id, user_id, item_id, state, next_due_at, last_review_at,
		       last_success_at, consecutive_successes, consecutive_failures,
		       current_difficulty, capped_at_ok, created_at, updated_at
		FROM masteries
		WHERE user_id = $1 AND item_id = $2`

	m := &mastery.Mastery{}
	err := r.pool.QueryRow(ctx, q, userID, itemID).Scan(
		&m.ID, &m.UserID, &m.ItemID, &m.State, &m.NextDueAt, &m.LastReviewAt,
		&m.LastSuccessAt, &m.ConsecutiveSuccesses, &m.ConsecutiveFailures,
		&m.CurrentDifficulty, &m.CappedAtOK, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("mastery.Repository.FindByUserAndItem: %w", mastery.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("mastery.Repository.FindByUserAndItem: %w", err)
	}
	return m, nil
}

func (r *MasteryRepository) FindDueByUser(ctx context.Context, userID uuid.UUID, before time.Time) ([]*mastery.Mastery, error) {
	const q = `
		SELECT id, user_id, item_id, state, next_due_at, last_review_at,
		       last_success_at, consecutive_successes, consecutive_failures,
		       current_difficulty, capped_at_ok, created_at, updated_at
		FROM masteries
		WHERE user_id = $1 AND (next_due_at IS NULL OR next_due_at <= $2)
		ORDER BY next_due_at ASC NULLS FIRST`

	rows, err := r.pool.Query(ctx, q, userID, before)
	if err != nil {
		return nil, fmt.Errorf("mastery.Repository.FindDueByUser: %w", err)
	}
	defer rows.Close()

	return scanMasteries(rows)
}

func (r *MasteryRepository) FindByUserAndState(ctx context.Context, userID uuid.UUID, state mastery.State) ([]*mastery.Mastery, error) {
	const q = `
		SELECT id, user_id, item_id, state, next_due_at, last_review_at,
		       last_success_at, consecutive_successes, consecutive_failures,
		       current_difficulty, capped_at_ok, created_at, updated_at
		FROM masteries
		WHERE user_id = $1 AND state = $2
		ORDER BY updated_at DESC`

	rows, err := r.pool.Query(ctx, q, userID, state)
	if err != nil {
		return nil, fmt.Errorf("mastery.Repository.FindByUserAndState: %w", err)
	}
	defer rows.Close()

	return scanMasteries(rows)
}

func (r *MasteryRepository) FindByItem(ctx context.Context, itemID uuid.UUID) ([]*mastery.Mastery, error) {
	const q = `
		SELECT id, user_id, item_id, state, next_due_at, last_review_at,
		       last_success_at, consecutive_successes, consecutive_failures,
		       current_difficulty, capped_at_ok, created_at, updated_at
		FROM masteries
		WHERE item_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, itemID)
	if err != nil {
		return nil, fmt.Errorf("mastery.Repository.FindByItem: %w", err)
	}
	defer rows.Close()

	return scanMasteries(rows)
}

func (r *MasteryRepository) Save(ctx context.Context, m *mastery.Mastery) error {
	const q = `
		INSERT INTO masteries (
			id, user_id, item_id, state, next_due_at, last_review_at,
			last_success_at, consecutive_successes, consecutive_failures,
			current_difficulty, capped_at_ok, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (user_id, item_id) DO UPDATE SET
			state = EXCLUDED.state,
			next_due_at = EXCLUDED.next_due_at,
			last_review_at = EXCLUDED.last_review_at,
			last_success_at = EXCLUDED.last_success_at,
			consecutive_successes = EXCLUDED.consecutive_successes,
			consecutive_failures = EXCLUDED.consecutive_failures,
			current_difficulty = EXCLUDED.current_difficulty,
			capped_at_ok = EXCLUDED.capped_at_ok,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, q,
		m.ID, m.UserID, m.ItemID, m.State, m.NextDueAt, m.LastReviewAt,
		m.LastSuccessAt, m.ConsecutiveSuccesses, m.ConsecutiveFailures,
		m.CurrentDifficulty, m.CappedAtOK, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("mastery.Repository.Save: %w", err)
	}
	return nil
}

func (r *MasteryRepository) SaveAll(ctx context.Context, masteries []*mastery.Mastery) error {
	if len(masteries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const q = `
		INSERT INTO masteries (
			id, user_id, item_id, state, next_due_at, last_review_at,
			last_success_at, consecutive_successes, consecutive_failures,
			current_difficulty, capped_at_ok, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (user_id, item_id) DO UPDATE SET
			state = EXCLUDED.state,
			next_due_at = EXCLUDED.next_due_at,
			last_review_at = EXCLUDED.last_review_at,
			last_success_at = EXCLUDED.last_success_at,
			consecutive_successes = EXCLUDED.consecutive_successes,
			consecutive_failures = EXCLUDED.consecutive_failures,
			current_difficulty = EXCLUDED.current_difficulty,
			capped_at_ok = EXCLUDED.capped_at_ok,
			updated_at = EXCLUDED.updated_at`

	for _, m := range masteries {
		batch.Queue(q,
			m.ID, m.UserID, m.ItemID, m.State, m.NextDueAt, m.LastReviewAt,
			m.LastSuccessAt, m.ConsecutiveSuccesses, m.ConsecutiveFailures,
			m.CurrentDifficulty, m.CappedAtOK, m.CreatedAt, m.UpdatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range masteries {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("mastery.Repository.SaveAll: %w", err)
		}
	}
	return nil
}

func scanMasteries(rows pgx.Rows) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for rows.Next() {
		m := &mastery.Mastery{}
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.ItemID, &m.State, &m.NextDueAt, &m.LastReviewAt,
			&m.LastSuccessAt, &m.ConsecutiveSuccesses, &m.ConsecutiveFailures,
			&m.CurrentDifficulty, &m.CappedAtOK, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("mastery.Repository: scan: %w", err)
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mastery.Repository: rows: %w", err)
	}
	return result, nil
}
