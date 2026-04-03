package validation

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

// fixedIDGen is a test stub for event.IDGenerator that returns a predetermined UUID.
type fixedIDGen struct {
	id uuid.UUID
}

func (f fixedIDGen) New() uuid.UUID { return f.id }

var _ event.IDGenerator = fixedIDGen{} // compile-time check

// helper: create a pending ValidationTask for testing.
func newTestTask(t *testing.T) *ValidationTask {
	t.Helper()
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	itemID := uuid.Must(uuid.NewV7())
	taskID := uuid.Must(uuid.NewV7())
	return NewValidationTask(fixedIDGen{id: taskID}, itemID, SourceUncertainty, now)
}

// --- NewValidationTask ---

func TestNewValidationTask_CreatesValidPendingTask(t *testing.T) {
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	itemID := uuid.Must(uuid.NewV7())
	taskID := uuid.Must(uuid.NewV7())
	idGen := fixedIDGen{id: taskID}

	task := NewValidationTask(idGen, itemID, SourceFidelity, now)

	if task.ID != taskID {
		t.Errorf("ID: got %v, want %v", task.ID, taskID)
	}
	if task.ItemID != itemID {
		t.Errorf("ItemID: got %v, want %v", task.ItemID, itemID)
	}
	if task.Status != StatusPending {
		t.Errorf("Status: got %q, want %q", task.Status, StatusPending)
	}
	if task.Source != SourceFidelity {
		t.Errorf("Source: got %q, want %q", task.Source, SourceFidelity)
	}
	if !task.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", task.CreatedAt, now)
	}
	if !task.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt: got %v, want %v", task.UpdatedAt, now)
	}
	if task.ResolvedBy != nil {
		t.Errorf("ResolvedBy: got %v, want nil", task.ResolvedBy)
	}
	if task.CropURL != nil {
		t.Errorf("CropURL: got %v, want nil", task.CropURL)
	}
	if task.Suggestion != nil {
		t.Errorf("Suggestion: got %v, want nil", task.Suggestion)
	}
}

func TestNewValidationTask_DifferentSources(t *testing.T) {
	sources := []TaskSource{
		SourceUncertainty,
		SourceStudentReport,
		SourceAnomaly,
		SourceCoherence,
		SourceFidelity,
		SourceStudentDeferred,
	}
	now := time.Now()
	for _, src := range sources {
		t.Run(string(src), func(t *testing.T) {
			task := NewValidationTask(event.UUIDv7Generator{}, uuid.Must(uuid.NewV7()), src, now)
			if task.Source != src {
				t.Errorf("Source: got %q, want %q", task.Source, src)
			}
			if task.Status != StatusPending {
				t.Errorf("Status: got %q, want %q", task.Status, StatusPending)
			}
		})
	}
}

// --- Resolve (table-driven) ---

func TestValidationTask_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		action     func(*ValidationTask, uuid.UUID, time.Time) error
		wantStatus TaskStatus
	}{
		{"confirm", (*ValidationTask).Confirm, StatusConfirmed},
		{"correct", (*ValidationTask).Correct, StatusCorrected},
		{"mark_unknown", (*ValidationTask).MarkUnknown, StatusUnknownAnswer},
		{"ignore", (*ValidationTask).Ignore, StatusIgnored},
		{"defer_by_student", (*ValidationTask).DeferByStudent, StatusDeferredByStudent},
		{"ignore_by_student", (*ValidationTask).IgnoreByStudent, StatusIgnoredByStudent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newTestTask(t)
			resolverID := uuid.Must(uuid.NewV7())
			resolveTime := time.Date(2026, 4, 2, 14, 0, 0, 0, time.UTC)

			err := tt.action(task, resolverID, resolveTime)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if task.Status != tt.wantStatus {
				t.Errorf("Status: got %q, want %q", task.Status, tt.wantStatus)
			}
			if task.ResolvedBy == nil {
				t.Fatal("ResolvedBy: got nil, want non-nil")
			}
			if *task.ResolvedBy != resolverID {
				t.Errorf("ResolvedBy: got %v, want %v", *task.ResolvedBy, resolverID)
			}
			if !task.UpdatedAt.Equal(resolveTime) {
				t.Errorf("UpdatedAt: got %v, want %v", task.UpdatedAt, resolveTime)
			}
		})
	}
}

// --- Resolve via raw Resolve method ---

func TestValidationTask_Resolve_DirectCall(t *testing.T) {
	task := newTestTask(t)
	resolverID := uuid.Must(uuid.NewV7())
	now := time.Now()

	err := task.Resolve(StatusCorrected, resolverID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != StatusCorrected {
		t.Errorf("Status: got %q, want %q", task.Status, StatusCorrected)
	}
}

func TestValidationTask_Resolve_CannotResolveToPending(t *testing.T) {
	task := newTestTask(t)
	resolverID := uuid.Must(uuid.NewV7())
	now := time.Now()

	err := task.Resolve(StatusPending, resolverID, now)
	if err == nil {
		t.Fatal("expected error when resolving to PENDING, got nil")
	}
	if task.Status != StatusPending {
		t.Errorf("Status should remain PENDING, got %q", task.Status)
	}
}

// --- Double resolution: ErrAlreadyResolved ---

func TestValidationTask_DoubleResolve_ReturnsErrAlreadyResolved(t *testing.T) {
	resolvedStatuses := []struct {
		name   string
		action func(*ValidationTask, uuid.UUID, time.Time) error
	}{
		{"confirmed", (*ValidationTask).Confirm},
		{"corrected", (*ValidationTask).Correct},
		{"unknown", (*ValidationTask).MarkUnknown},
		{"ignored", (*ValidationTask).Ignore},
		{"deferred_by_student", (*ValidationTask).DeferByStudent},
		{"ignored_by_student", (*ValidationTask).IgnoreByStudent},
	}

	for _, rs := range resolvedStatuses {
		t.Run(rs.name, func(t *testing.T) {
			task := newTestTask(t)
			resolverID := uuid.Must(uuid.NewV7())
			now := time.Now()

			// First resolution succeeds.
			if err := rs.action(task, resolverID, now); err != nil {
				t.Fatalf("first resolve: unexpected error: %v", err)
			}

			// Second resolution must fail with ErrAlreadyResolved.
			err := task.Confirm(uuid.Must(uuid.NewV7()), now.Add(time.Hour))
			if err != ErrAlreadyResolved {
				t.Errorf("second resolve: got %v, want ErrAlreadyResolved", err)
			}
		})
	}
}

// --- IsResolved ---

func TestValidationTask_IsResolved(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"pending is not resolved", StatusPending, false},
		{"confirmed is resolved", StatusConfirmed, true},
		{"corrected is resolved", StatusCorrected, true},
		{"unknown_answer is resolved", StatusUnknownAnswer, true},
		{"ignored is resolved", StatusIgnored, true},
		{"deferred_by_student is resolved", StatusDeferredByStudent, true},
		{"ignored_by_student is resolved", StatusIgnoredByStudent, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newTestTask(t)
			task.Status = tt.status

			got := task.IsResolved()
			if got != tt.want {
				t.Errorf("IsResolved(): got %v, want %v", got, tt.want)
			}
		})
	}
}

// --- State preservation: resolve does not alter other fields ---

func TestValidationTask_Resolve_PreservesImmutableFields(t *testing.T) {
	task := newTestTask(t)
	originalID := task.ID
	originalItemID := task.ItemID
	originalSource := task.Source
	originalCreatedAt := task.CreatedAt

	resolverID := uuid.Must(uuid.NewV7())
	now := time.Now()

	_ = task.Confirm(resolverID, now)

	if task.ID != originalID {
		t.Errorf("ID changed after resolve")
	}
	if task.ItemID != originalItemID {
		t.Errorf("ItemID changed after resolve")
	}
	if task.Source != originalSource {
		t.Errorf("Source changed after resolve")
	}
	if !task.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("CreatedAt changed after resolve")
	}
}
