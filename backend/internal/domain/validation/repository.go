package validation

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the port for validation task persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ValidationTask, error)
	FindPendingByItem(ctx context.Context, itemID uuid.UUID) ([]*ValidationTask, error)
	FindPendingAll(ctx context.Context, limit int) ([]*ValidationTask, error)
	Save(ctx context.Context, t *ValidationTask) error
}
