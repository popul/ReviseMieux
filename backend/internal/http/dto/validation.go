package dto

import "time"

// ValidationTaskResponse represents a validation task in API responses.
type ValidationTaskResponse struct {
	ID         string     `json:"id"`
	ItemID     string     `json:"item_id"`
	ChapterID  string     `json:"chapter_id,omitempty"`
	CropURL    *string    `json:"crop_url,omitempty"`
	Suggestion *string    `json:"suggestion,omitempty"`
	Priority   int        `json:"priority"`
	Status     string     `json:"status"`
	Source     string     `json:"source"`
	ItemTerm   *string    `json:"item_term,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// ResolveValidationRequest is the request body for resolving a validation task.
type ResolveValidationRequest struct {
	Action     string  `json:"action" binding:"required,oneof=confirm correct unknown ignore"`
	Correction *string `json:"correction,omitempty"`
	Note       *string `json:"note,omitempty"`
}
