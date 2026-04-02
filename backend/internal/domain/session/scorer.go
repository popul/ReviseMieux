package session

import "context"

// LLMScoreResult contains the LLM scoring output for text answers.
type LLMScoreResult struct {
	Score       float64 // 0.0 to 1.0
	IsCorrect   bool    // true if score >= 0.7
	Explanation string  // brief explanation of why
}

// Scorer evaluates a student answer against an expected answer using an LLM.
type Scorer interface {
	ScoreAnswer(ctx context.Context, prompt string, expectedAnswer string, studentAnswer string) (*LLMScoreResult, error)
}
