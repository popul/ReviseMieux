package mastery

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository is the port for mastery persistence.
type Repository interface {
	// FindByID returns a mastery by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Mastery, error)

	// FindByUserAndItem returns the mastery for a (user, item) pair.
	FindByUserAndItem(ctx context.Context, userID, itemID uuid.UUID) (*Mastery, error)

	// FindDueByUser returns all masteries due for review for a user.
	FindDueByUser(ctx context.Context, userID uuid.UUID, before time.Time) ([]*Mastery, error)

	// FindAllByUser returns all masteries for a user (all states).
	FindAllByUser(ctx context.Context, userID uuid.UUID) ([]*Mastery, error)

	// FindByUserAndState returns all masteries in a given state for a user.
	FindByUserAndState(ctx context.Context, userID uuid.UUID, state State) ([]*Mastery, error)

	// FindByItem returns all masteries for a given item (across all users).
	FindByItem(ctx context.Context, itemID uuid.UUID) ([]*Mastery, error)

	// Save persists a mastery (insert or update).
	Save(ctx context.Context, m *Mastery) error

	// SaveAll persists multiple masteries in a batch.
	SaveAll(ctx context.Context, masteries []*Mastery) error
}
