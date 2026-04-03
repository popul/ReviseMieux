package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LLMCallEntry represents a single LLM API call for logging and monitoring.
type LLMCallEntry struct {
	Timestamp     time.Time
	CallType      string // "structuration", "fidelity", "question_gen", "ocr", "scoring"
	Model         string
	PromptVersion string
	PromptHash    string
	TokensInput   int
	TokensOutput  int
	DurationMs    int64
	CostUSD       float64
	Status        string // "success", "error", "timeout"
	ItemsCount    *int
	AvgConfidence *float64
	CacheHit      bool
	Batch         bool
	UserID        *uuid.UUID
	ChapterID     *uuid.UUID
	Error         string
}

// LLMCallLogger defines the port for logging LLM API calls.
type LLMCallLogger interface {
	LogCall(ctx context.Context, entry LLMCallEntry) error
}
