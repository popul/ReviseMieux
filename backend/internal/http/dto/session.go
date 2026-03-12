package dto

import "time"

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	SessionType          string     `json:"session_type"`
	Status               string     `json:"status"`
	TriggerType          string     `json:"trigger_type"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	CurrentQuestionIndex int        `json:"current_question_index"`
	ChapterIDs           []string   `json:"chapter_ids,omitempty"`
}

// QuestionResponse represents a question in API responses.
type QuestionResponse struct {
	ID             string  `json:"id"`
	TemplateID     string  `json:"template_id"`
	ItemID         string  `json:"item_id"`
	RenderedPrompt string  `json:"rendered_prompt"`
	VisualURL      *string `json:"visual_url,omitempty"`
}

// SubmitAnswerRequest is the request body for submitting an answer.
type SubmitAnswerRequest struct {
	QuestionID string  `json:"question_id" binding:"required"`
	Answer     string  `json:"answer" binding:"required"`
	Score      float64 `json:"score" binding:"required,min=0,max=1"`
}

// SubmitAnswerResponse is returned after submitting an answer.
type SubmitAnswerResponse struct {
	AttemptID string    `json:"attempt_id"`
	Score     float64   `json:"score"`
	Feedback  *Feedback `json:"feedback,omitempty"`
}

// Feedback contains enriched feedback for incorrect answers (Z4-AC09).
type Feedback struct {
	CorrectAnswer  string `json:"correct_answer,omitempty"`
	WhatWasMissing string `json:"what_was_missing,omitempty"`
	Hint           string `json:"hint,omitempty"`
}

// ComposeDailyRequest is the request body for composing a daily session.
type ComposeDailyRequest struct {
	ChapterID string `json:"chapter_id" binding:"required"`
}

// ComposeEveningFirstRequest is the request body for proposing an evening_first session.
type ComposeEveningFirstRequest struct {
	ChapterID string   `json:"chapter_id" binding:"required"`
	ItemIDs   []string `json:"item_ids" binding:"required"`
}
