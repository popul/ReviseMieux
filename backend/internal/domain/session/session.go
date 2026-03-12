package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

var (
	ErrNotFound              = errors.New("session: not found")
	ErrSessionNotResumable   = errors.New("session cannot be resumed from current status")
	ErrSessionNotCompletable = errors.New("session cannot be completed from current status")
	ErrNoQuestions           = errors.New("session has no questions")
	ErrEmptyPool             = errors.New("session: no items available for composition")
)

// --- Value Objects ---

type SessionType string

const (
	TypeDaily        SessionType = "daily"
	TypeDiagnostic   SessionType = "diagnostic"
	TypeMockExam     SessionType = "mock_exam"
	TypeEveningFirst SessionType = "evening_first"
	TypePreClass     SessionType = "pre_class"
)

// Valid returns true if the session type is one of the known values.
func (s SessionType) Valid() bool {
	switch s {
	case TypeDaily, TypeDiagnostic, TypeMockExam, TypeEveningFirst, TypePreClass:
		return true
	}
	return false
}

// ParseSessionType converts a string to a SessionType, returning an error if invalid.
func ParseSessionType(s string) (SessionType, error) {
	st := SessionType(s)
	if !st.Valid() {
		return "", fmt.Errorf("session.ParseSessionType: invalid type %q", s)
	}
	return st, nil
}

type SessionStatus string

const (
	StatusComposing  SessionStatus = "COMPOSING"
	StatusInProgress SessionStatus = "IN_PROGRESS"
	StatusCompleted  SessionStatus = "COMPLETED"
	StatusExpired    SessionStatus = "EXPIRED"
	StatusAbandoned  SessionStatus = "ABANDONED"
)

// Valid returns true if the session status is one of the known values.
func (s SessionStatus) Valid() bool {
	switch s {
	case StatusComposing, StatusInProgress, StatusCompleted, StatusExpired, StatusAbandoned:
		return true
	}
	return false
}

type SessionTrigger string

const (
	TriggerManual       SessionTrigger = "manual"
	TriggerScheduled    SessionTrigger = "scheduled"
	TriggerNotification SessionTrigger = "notification"
)

type QuestionType string

const (
	QuestionMCQ         QuestionType = "MCQ"
	QuestionShortAnswer QuestionType = "SHORT_ANSWER"
	QuestionNumeric     QuestionType = "NUMERIC"
	QuestionCloze       QuestionType = "CLOZE"
	QuestionRubric      QuestionType = "RUBRIC"
)

type AttemptSource string

const (
	AttemptInteractive  AttemptSource = "interactive"
	AttemptPaperReport  AttemptSource = "paper_report"
)

// --- Entities ---

// Session is the aggregate root for a review session.
type Session struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	SessionType          SessionType
	Status               SessionStatus
	TriggerType          SessionTrigger
	StartedAt            *time.Time
	CompletedAt          *time.Time
	CurrentQuestionIndex int
	IncludesPreClass     bool
	ChapterIDs           []uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewSession creates a session in COMPOSING status.
func NewSession(idGen event.IDGenerator, userID uuid.UUID, sessionType SessionType, trigger SessionTrigger, now time.Time) *Session {
	return &Session{
		ID:          idGen.New(),
		UserID:      userID,
		SessionType: sessionType,
		Status:      StatusComposing,
		TriggerType: trigger,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Start transitions from COMPOSING to IN_PROGRESS.
func (s *Session) Start(now time.Time) error {
	if s.Status != StatusComposing {
		return ErrSessionNotResumable
	}
	s.Status = StatusInProgress
	s.StartedAt = &now
	s.UpdatedAt = now
	return nil
}

// Resume allows continuing an IN_PROGRESS session.
func (s *Session) Resume() error {
	if s.Status != StatusInProgress {
		return ErrSessionNotResumable
	}
	return nil
}

// Complete marks the session as completed.
func (s *Session) Complete(now time.Time) error {
	if s.Status != StatusInProgress {
		return ErrSessionNotCompletable
	}
	s.Status = StatusCompleted
	s.CompletedAt = &now
	s.UpdatedAt = now
	return nil
}

// Abandon marks the session as abandoned.
func (s *Session) Abandon(now time.Time) {
	s.Status = StatusAbandoned
	s.UpdatedAt = now
}

// Question represents a generated question within a session.
type Question struct {
	ID                    uuid.UUID
	SessionID             uuid.UUID
	TemplateID            string
	ItemID                uuid.UUID
	VisualBlockID         *uuid.UUID
	RenderedPrompt        string
	RenderedVisualURL     *string
	ExpectedAnswer        []byte // JSONB
	GradingPolicy         string
	Clarification         []byte // JSONB
	LLMModelVersion       *string
	PromptTemplateVersion *string
	TimesSeen             int
	CreatedAt             time.Time
}

// Attempt represents a student's answer to a question.
type Attempt struct {
	ID                uuid.UUID
	SessionID         uuid.UUID
	QuestionID        uuid.UUID
	UserID            uuid.UUID
	Answer            []byte // JSONB
	Score             float64
	Feedback          *string
	Source            AttemptSource
	RapidResponse     bool
	ResponseTimeMs    *int
	HintUsed          bool
	ClarificationUsed bool
	CreatedAt         time.Time
}

// NewAttempt creates a new attempt.
func NewAttempt(idGen event.IDGenerator, sessionID, questionID, userID uuid.UUID, answer []byte, score float64, now time.Time) *Attempt {
	return &Attempt{
		ID:         idGen.New(),
		SessionID:  sessionID,
		QuestionID: questionID,
		UserID:     userID,
		Answer:     answer,
		Score:      score,
		Source:     AttemptInteractive,
		CreatedAt:  now,
	}
}
