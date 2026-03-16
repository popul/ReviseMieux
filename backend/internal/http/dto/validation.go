package dto

import "time"

// ValidationTaskResponse represents a validation task in API responses.
type ValidationTaskResponse struct {
	ID         string     `json:"id"`
	ItemID     string     `json:"item_id"`
	CropURL    *string    `json:"crop_url,omitempty"`
	Suggestion *string    `json:"suggestion,omitempty"`
	Priority   int        `json:"priority"`
	Status     string     `json:"status"`
	Source     string     `json:"source"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ResolveValidationRequest is the request body for resolving a validation task.
type ResolveValidationRequest struct {
	Action        string  `json:"action" binding:"required,oneof=confirm correct unknown ignore"`
	CorrectedTerm *string `json:"corrected_term,omitempty"`
}
