package event

import (
	"time"

	"github.com/google/uuid"
)

// Event is the base interface for all domain events.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Publisher defines how domain events are dispatched.
type Publisher interface {
	Publish(events ...Event)
}

// ItemsGenerated is emitted when the pipeline produces items for a chapter.
type ItemsGenerated struct {
	ChapterID  uuid.UUID
	ItemIDs    []uuid.UUID
	OccurredOn time.Time
}

func (e ItemsGenerated) EventName() string    { return "items.generated" }
func (e ItemsGenerated) OccurredAt() time.Time { return e.OccurredOn }

// AttemptRecorded is emitted when a student records an attempt on a question.
type AttemptRecorded struct {
	UserID     uuid.UUID
	ItemID     uuid.UUID
	Score      float64
	OccurredOn time.Time
}

func (e AttemptRecorded) EventName() string    { return "attempt.recorded" }
func (e AttemptRecorded) OccurredAt() time.Time { return e.OccurredOn }

// ValidationResolved is emitted when a HITL validation task is resolved.
type ValidationResolved struct {
	ItemID     uuid.UUID
	OccurredOn time.Time
}

func (e ValidationResolved) EventName() string    { return "validation.resolved" }
func (e ValidationResolved) OccurredAt() time.Time { return e.OccurredOn }

// ExamCreated is emitted when a new exam is linked to chapters.
type ExamCreated struct {
	ExamID     uuid.UUID
	ChapterIDs []uuid.UUID
	OccurredOn time.Time
}

func (e ExamCreated) EventName() string    { return "exam.created" }
func (e ExamCreated) OccurredAt() time.Time { return e.OccurredOn }
