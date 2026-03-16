package session

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the port for session aggregate persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)
	FindActiveByUser(ctx context.Context, userID uuid.UUID) (*Session, error)
	Save(ctx context.Context, s *Session) error

	// Questions
	FindQuestionByID(ctx context.Context, id uuid.UUID) (*Question, error)
	FindQuestionsBySession(ctx context.Context, sessionID uuid.UUID) ([]*Question, error)
	SaveQuestion(ctx context.Context, q *Question) error

	// Attempts
	SaveAttempt(ctx context.Context, a *Attempt) error
	FindAttemptsBySession(ctx context.Context, sessionID uuid.UUID) ([]*Attempt, error)
}
