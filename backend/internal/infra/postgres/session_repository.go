package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/session"
)

var _ session.Repository = (*SessionRepository)(nil)

// SessionRepository implements session.Repository using PostgreSQL.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, session_type, status, trigger_type,
		       started_at, completed_at, current_question_index,
		       includes_pre_class, created_at, updated_at
		FROM sessions WHERE id = $1`, id)

	s := &session.Session{}
	if err := row.Scan(
		&s.ID, &s.UserID, &s.SessionType, &s.Status, &s.TriggerType,
		&s.StartedAt, &s.CompletedAt, &s.CurrentQuestionIndex,
		&s.IncludesPreClass, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("session_repo.FindByID: %w", session.ErrNotFound)
	}

	// Load chapter IDs
	rows, err := r.pool.Query(ctx, `SELECT chapter_id FROM session_chapters WHERE session_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("session_repo.FindByID: load chapters: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var chID uuid.UUID
		if err := rows.Scan(&chID); err != nil {
			return nil, fmt.Errorf("session_repo.FindByID: scan chapter: %w", err)
		}
		s.ChapterIDs = append(s.ChapterIDs, chID)
	}

	return s, nil
}

func (r *SessionRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) (*session.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, session_type, status, trigger_type,
		       started_at, completed_at, current_question_index,
		       includes_pre_class, created_at, updated_at
		FROM sessions WHERE user_id = $1 AND status = 'IN_PROGRESS'
		ORDER BY created_at DESC LIMIT 1`, userID)

	s := &session.Session{}
	if err := row.Scan(
		&s.ID, &s.UserID, &s.SessionType, &s.Status, &s.TriggerType,
		&s.StartedAt, &s.CompletedAt, &s.CurrentQuestionIndex,
		&s.IncludesPreClass, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("session_repo.FindActiveByUser: %w", session.ErrNotFound)
	}
	return s, nil
}

func (r *SessionRepository) Save(ctx context.Context, s *session.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, session_type, status, trigger_type,
		                      started_at, completed_at, current_question_index,
		                      includes_pre_class, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			current_question_index = EXCLUDED.current_question_index,
			updated_at = EXCLUDED.updated_at`,
		s.ID, s.UserID, s.SessionType, s.Status, s.TriggerType,
		s.StartedAt, s.CompletedAt, s.CurrentQuestionIndex,
		s.IncludesPreClass, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("session_repo.Save: %w", err)
	}

	// Upsert chapter links
	for _, chID := range s.ChapterIDs {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO session_chapters (session_id, chapter_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, s.ID, chID)
		if err != nil {
			return fmt.Errorf("session_repo.Save: chapter link: %w", err)
		}
	}

	return nil
}

func (r *SessionRepository) FindQuestionByID(ctx context.Context, id uuid.UUID) (*session.Question, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, session_id, template_id, item_id, visual_block_id,
		       rendered_prompt, rendered_visual_url, expected_answer,
		       grading_policy, clarification, llm_model_version,
		       prompt_template_version, times_seen, created_at
		FROM questions WHERE id = $1`, id)

	q := &session.Question{}
	if err := row.Scan(
		&q.ID, &q.SessionID, &q.TemplateID, &q.ItemID, &q.VisualBlockID,
		&q.RenderedPrompt, &q.RenderedVisualURL, &q.ExpectedAnswer,
		&q.GradingPolicy, &q.Clarification, &q.LLMModelVersion,
		&q.PromptTemplateVersion, &q.TimesSeen, &q.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("session_repo.FindQuestionByID: %w", session.ErrNotFound)
	}
	return q, nil
}

func (r *SessionRepository) FindQuestionsBySession(ctx context.Context, sessionID uuid.UUID) ([]*session.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, session_id, template_id, item_id, visual_block_id,
		       rendered_prompt, rendered_visual_url, expected_answer,
		       grading_policy, clarification, llm_model_version,
		       prompt_template_version, times_seen, created_at
		FROM questions WHERE session_id = $1 ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_repo.FindQuestionsBySession: %w", err)
	}
	defer rows.Close()

	var result []*session.Question
	for rows.Next() {
		q := &session.Question{}
		if err := rows.Scan(
			&q.ID, &q.SessionID, &q.TemplateID, &q.ItemID, &q.VisualBlockID,
			&q.RenderedPrompt, &q.RenderedVisualURL, &q.ExpectedAnswer,
			&q.GradingPolicy, &q.Clarification, &q.LLMModelVersion,
			&q.PromptTemplateVersion, &q.TimesSeen, &q.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("session_repo.FindQuestionsBySession: scan: %w", err)
		}
		result = append(result, q)
	}
	return result, nil
}

func (r *SessionRepository) SaveQuestion(ctx context.Context, q *session.Question) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO questions (id, session_id, template_id, item_id, visual_block_id,
		                       rendered_prompt, rendered_visual_url, expected_answer,
		                       grading_policy, clarification, llm_model_version,
		                       prompt_template_version, times_seen, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (id) DO UPDATE SET times_seen = EXCLUDED.times_seen`,
		q.ID, q.SessionID, q.TemplateID, q.ItemID, q.VisualBlockID,
		q.RenderedPrompt, q.RenderedVisualURL, q.ExpectedAnswer,
		q.GradingPolicy, q.Clarification, q.LLMModelVersion,
		q.PromptTemplateVersion, q.TimesSeen, q.CreatedAt)
	if err != nil {
		return fmt.Errorf("session_repo.SaveQuestion: %w", err)
	}
	return nil
}

func (r *SessionRepository) SaveAttempt(ctx context.Context, a *session.Attempt) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attempts (id, session_id, question_id, user_id, answer,
		                      score, feedback, source, rapid_response,
		                      response_time_ms, hint_used, clarification_used, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		a.ID, a.SessionID, a.QuestionID, a.UserID, a.Answer,
		a.Score, a.Feedback, a.Source, a.RapidResponse,
		a.ResponseTimeMs, a.HintUsed, a.ClarificationUsed, a.CreatedAt)
	if err != nil {
		return fmt.Errorf("session_repo.SaveAttempt: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindAttemptsBySession(ctx context.Context, sessionID uuid.UUID) ([]*session.Attempt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, session_id, question_id, user_id, answer,
		       score, feedback, source, rapid_response,
		       response_time_ms, hint_used, clarification_used, created_at
		FROM attempts WHERE session_id = $1 ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_repo.FindAttemptsBySession: %w", err)
	}
	defer rows.Close()

	var result []*session.Attempt
	for rows.Next() {
		a := &session.Attempt{}
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.QuestionID, &a.UserID, &a.Answer,
			&a.Score, &a.Feedback, &a.Source, &a.RapidResponse,
			&a.ResponseTimeMs, &a.HintUsed, &a.ClarificationUsed, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("session_repo.FindAttemptsBySession: scan: %w", err)
		}
		result = append(result, a)
	}
	return result, nil
}
