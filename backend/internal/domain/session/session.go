package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotResumable = errors.New("session cannot be resumed from current status")
	ErrSessionCompleted    = errors.New("session is already completed")
	ErrNoQuestions         = errors.New("session has no questions")
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

type SessionStatus string

const (
	StatusComposing  SessionStatus = "COMPOSING"
	StatusInProgress SessionStatus = "IN_PROGRESS"
	StatusCompleted  SessionStatus = "COMPLETED"
	StatusExpired    SessionStatus = "EXPIRED"
	StatusAbandoned  SessionStatus = "ABANDONED"
)

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
func NewSession(userID uuid.UUID, sessionType SessionType, trigger SessionTrigger, now time.Time) *Session {
	return &Session{
		ID:          uuid.Must(uuid.NewV7()),
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
		return ErrSessionCompleted
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
func NewAttempt(sessionID, questionID, userID uuid.UUID, answer []byte, score float64, now time.Time) *Attempt {
	return &Attempt{
		ID:         uuid.Must(uuid.NewV7()),
		SessionID:  sessionID,
		QuestionID: questionID,
		UserID:     userID,
		Answer:     answer,
		Score:      score,
		Source:     AttemptInteractive,
		CreatedAt:  now,
	}
}
