package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event is the base interface for all domain events.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// BaseEvent provides the shared OccurredOn field and OccurredAt() implementation.
type BaseEvent struct {
	OccurredOn time.Time
}

func (e BaseEvent) OccurredAt() time.Time { return e.OccurredOn }

// Publisher defines how domain events are dispatched.
type Publisher interface {
	Publish(ctx context.Context, events ...Event) error
}

// ItemsGenerated is emitted when the pipeline produces items for a chapter.
type ItemsGenerated struct {
	BaseEvent
	ChapterID uuid.UUID
	ItemIDs   []uuid.UUID
}

func (e ItemsGenerated) EventName() string { return "items.generated" }

// AttemptRecorded is emitted when a student records an attempt on a question.
type AttemptRecorded struct {
	BaseEvent
	UserID uuid.UUID
	ItemID uuid.UUID
	Score  float64
}

func (e AttemptRecorded) EventName() string { return "attempt.recorded" }

// ValidationResolved is emitted when a HITL validation task is resolved.
type ValidationResolved struct {
	BaseEvent
	ItemID uuid.UUID
}

func (e ValidationResolved) EventName() string { return "validation.resolved" }

// ExamCreated is emitted when a new exam is linked to chapters.
type ExamCreated struct {
	BaseEvent
	ExamID     uuid.UUID
	ChapterIDs []uuid.UUID
}

func (e ExamCreated) EventName() string { return "exam.created" }
