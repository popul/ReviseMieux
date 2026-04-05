package validation

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

var (
	ErrAlreadyResolved = errors.New("validation task is already resolved")
	ErrNotFound        = errors.New("validation: not found")
)

// --- Value Objects ---

type TaskStatus string

const (
	StatusPending           TaskStatus = "PENDING"
	StatusConfirmed         TaskStatus = "CONFIRMED"
	StatusCorrected         TaskStatus = "CORRECTED"
	StatusUnknownAnswer     TaskStatus = "UNKNOWN_ANSWER"
	StatusIgnored           TaskStatus = "IGNORED"
	StatusDeferredByStudent TaskStatus = "DEFERRED_BY_STUDENT"
	StatusIgnoredByStudent  TaskStatus = "IGNORED_BY_STUDENT"
)

type TaskSource string

const (
	SourceUncertainty     TaskSource = "uncertainty_detection"
	SourceStudentReport   TaskSource = "student_report"
	SourceAnomaly         TaskSource = "anomaly_detection"
	SourceCoherence       TaskSource = "coherence_check"
	SourceFidelity        TaskSource = "fidelity_check"
	SourceStudentDeferred TaskSource = "student_deferred"
)

// --- Entity ---

// ValidationTask is the aggregate root for HITL validation.
type ValidationTask struct {
	ID          uuid.UUID
	ItemID      uuid.UUID
	CropURL     *string
	Suggestion  *string
	Priority    int
	Status      TaskStatus
	ResolvedBy  *uuid.UUID
	Source      TaskSource
	StudentNote *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewValidationTask creates a PENDING task.
func NewValidationTask(idGen event.IDGenerator, itemID uuid.UUID, source TaskSource, now time.Time) *ValidationTask {
	return &ValidationTask{
		ID:        idGen.New(),
		ItemID:    itemID,
		Status:    StatusPending,
		Source:    source,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Resolve transitions the task from PENDING to the given resolution status.
func (t *ValidationTask) Resolve(status TaskStatus, resolvedBy uuid.UUID, now time.Time) error {
	if t.Status != StatusPending {
		return ErrAlreadyResolved
	}
	if status == StatusPending {
		return errors.New("cannot resolve to PENDING status")
	}
	t.Status = status
	t.ResolvedBy = &resolvedBy
	t.UpdatedAt = now
	return nil
}

// Confirm marks the task as confirmed by a validator.
func (t *ValidationTask) Confirm(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusConfirmed, resolvedBy, now)
}

// Correct marks the task as corrected.
func (t *ValidationTask) Correct(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusCorrected, resolvedBy, now)
}

// MarkUnknown marks the validator as unsure (admin resolution).
func (t *ValidationTask) MarkUnknown(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusUnknownAnswer, resolvedBy, now)
}

// Ignore marks the task as ignored (admin resolution).
func (t *ValidationTask) Ignore(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusIgnored, resolvedBy, now)
}

// DeferByStudent marks the task as deferred by the student ("Je ne sais pas").
func (t *ValidationTask) DeferByStudent(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusDeferredByStudent, resolvedBy, now)
}

// IgnoreByStudent marks the task as ignored by the student.
func (t *ValidationTask) IgnoreByStudent(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusIgnoredByStudent, resolvedBy, now)
}

// IsResolved returns true if the task is no longer pending.
func (t *ValidationTask) IsResolved() bool {
	return t.Status != StatusPending
}
