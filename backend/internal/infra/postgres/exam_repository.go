package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// Compile-time check that ExamRepository implements chapter.ExamRepository.
var _ chapter.ExamRepository = (*ExamRepository)(nil)

// ExamRepository implements chapter.ExamRepository using PostgreSQL.
type ExamRepository struct {
	pool *pgxpool.Pool
}

// NewExamRepository creates a new ExamRepository.
func NewExamRepository(pool *pgxpool.Pool) *ExamRepository {
	return &ExamRepository{pool: pool}
}

func (r *ExamRepository) FindByID(ctx context.Context, id uuid.UUID) (*chapter.Exam, error) {
	const q = `
		SELECT id, user_id, name, exam_date, status, created_at, updated_at
		FROM exams
		WHERE id = $1`

	e := &chapter.Exam{}
	var examDate time.Time
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&e.ID, &e.UserID, &e.Title, &examDate, new(string), &e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("exam.Repository.FindByID: %w", chapter.ErrExamNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("exam.Repository.FindByID: %w", err)
	}
	e.ExamDate = examDate

	chapterIDs, err := r.findChapterIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	e.ChapterIDs = chapterIDs

	return e, nil
}

func (r *ExamRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*chapter.Exam, error) {
	const q = `
		SELECT id, user_id, name, exam_date, status, created_at, updated_at
		FROM exams
		WHERE user_id = $1
		ORDER BY exam_date ASC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("exam.Repository.FindByUser: %w", err)
	}
	defer rows.Close()

	var exams []*chapter.Exam
	for rows.Next() {
		e := &chapter.Exam{}
		var examDate time.Time
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Title, &examDate, new(string), &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("exam.Repository.FindByUser: scan: %w", err)
		}
		e.ExamDate = examDate
		exams = append(exams, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exam.Repository.FindByUser: rows: %w", err)
	}

	// Load chapter IDs for each exam
	for _, e := range exams {
		chapterIDs, err := r.findChapterIDs(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		e.ChapterIDs = chapterIDs
	}

	return exams, nil
}

func (r *ExamRepository) FindActiveByChapter(ctx context.Context, chapterID uuid.UUID) ([]*chapter.Exam, error) {
	const q = `
		SELECT e.id, e.user_id, e.name, e.exam_date, e.status, e.created_at, e.updated_at
		FROM exams e
		INNER JOIN chapter_exams ce ON ce.exam_id = e.id
		WHERE ce.chapter_id = $1 AND e.status = 'active'
		ORDER BY e.exam_date ASC`

	rows, err := r.pool.Query(ctx, q, chapterID)
	if err != nil {
		return nil, fmt.Errorf("exam.Repository.FindActiveByChapter: %w", err)
	}
	defer rows.Close()

	var exams []*chapter.Exam
	for rows.Next() {
		e := &chapter.Exam{}
		var examDate time.Time
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Title, &examDate, new(string), &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("exam.Repository.FindActiveByChapter: scan: %w", err)
		}
		e.ExamDate = examDate
		exams = append(exams, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exam.Repository.FindActiveByChapter: rows: %w", err)
	}

	for _, e := range exams {
		chapterIDs, err := r.findChapterIDs(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		e.ChapterIDs = chapterIDs
	}

	return exams, nil
}

func (r *ExamRepository) Save(ctx context.Context, exam *chapter.Exam) error {
	const upsertExam = `
		INSERT INTO exams (id, user_id, name, exam_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			exam_date = EXCLUDED.exam_date,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, upsertExam,
		exam.ID, exam.UserID, exam.Title, exam.ExamDate, exam.CreatedAt, exam.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exam.Repository.Save: %w", err)
	}

	// Replace chapter links
	const deleteLinks = `DELETE FROM chapter_exams WHERE exam_id = $1`
	if _, err := r.pool.Exec(ctx, deleteLinks, exam.ID); err != nil {
		return fmt.Errorf("exam.Repository.Save: delete links: %w", err)
	}

	if len(exam.ChapterIDs) > 0 {
		const insertLink = `INSERT INTO chapter_exams (chapter_id, exam_id) VALUES ($1, $2)`
		batch := &pgx.Batch{}
		for _, chID := range exam.ChapterIDs {
			batch.Queue(insertLink, chID, exam.ID)
		}
		br := r.pool.SendBatch(ctx, batch)
		defer br.Close()
		for range exam.ChapterIDs {
			if _, err := br.Exec(); err != nil {
				return fmt.Errorf("exam.Repository.Save: insert link: %w", err)
			}
		}
	}

	return nil
}

func (r *ExamRepository) findChapterIDs(ctx context.Context, examID uuid.UUID) ([]uuid.UUID, error) {
	const q = `SELECT chapter_id FROM chapter_exams WHERE exam_id = $1`
	rows, err := r.pool.Query(ctx, q, examID)
	if err != nil {
		return nil, fmt.Errorf("exam.Repository.findChapterIDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("exam.Repository.findChapterIDs: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
