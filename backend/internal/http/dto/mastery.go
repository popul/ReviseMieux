package dto

import "time"

// MasteryResponse represents a mastery entry in API responses.
type MasteryResponse struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	ItemID               string     `json:"item_id"`
	State                string     `json:"state"`
	NextDueAt            *time.Time `json:"next_due_at,omitempty"`
	LastReviewAt         *time.Time `json:"last_review_at,omitempty"`
	ConsecutiveSuccesses int        `json:"consecutive_successes"`
	ConsecutiveFailures  int        `json:"consecutive_failures"`
	LastSuccessAt        *time.Time `json:"last_success_at,omitempty"`
	CappedAtOK           bool       `json:"capped_at_ok"`
}

// RecordAttemptRequest is the request body for recording an attempt score.
// Score uses *float64 so that Gin can distinguish "absent" from "zero" (0.0 is a valid score).
type RecordAttemptRequest struct {
	ItemID string   `json:"item_id" binding:"required"`
	Score  *float64 `json:"score" binding:"required,min=0,max=1"`
}

// RecordAttemptResponse is returned after recording an attempt.
type RecordAttemptResponse struct {
	Mastery  MasteryResponse `json:"mastery"`
	NewState string          `json:"new_state"`
}
