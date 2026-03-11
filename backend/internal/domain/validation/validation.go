package validation

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyResolved = errors.New("validation task is already resolved")
)

// --- Value Objects ---

type TaskStatus string

const (
	StatusPending       TaskStatus = "PENDING"
	StatusConfirmed     TaskStatus = "CONFIRMED"
	StatusCorrected     TaskStatus = "CORRECTED"
	StatusUnknownAnswer TaskStatus = "UNKNOWN_ANSWER"
	StatusIgnored       TaskStatus = "IGNORED"
)

type TaskSource string

const (
	SourceUncertainty  TaskSource = "uncertainty_detection"
	SourceStudentReport TaskSource = "student_report"
	SourceAnomaly      TaskSource = "anomaly_detection"
	SourceCoherence    TaskSource = "coherence_check"
	SourceFidelity     TaskSource = "fidelity_check"
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
func NewValidationTask(itemID uuid.UUID, source TaskSource, now time.Time) *ValidationTask {
	return &ValidationTask{
		ID:        uuid.Must(uuid.NewV7()),
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

// MarkUnknown marks the validator as unsure.
func (t *ValidationTask) MarkUnknown(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusUnknownAnswer, resolvedBy, now)
}

// Ignore marks the task as ignored.
func (t *ValidationTask) Ignore(resolvedBy uuid.UUID, now time.Time) error {
	return t.Resolve(StatusIgnored, resolvedBy, now)
}

// IsResolved returns true if the task is no longer pending.
func (t *ValidationTask) IsResolved() bool {
	return t.Status != StatusPending
}
