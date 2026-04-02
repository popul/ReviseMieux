package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/popul/revisemieux/internal/domain/session"
)

const scoringSystemPrompt = `Tu es un correcteur bienveillant pour collégiens. Tu compares la réponse de l'élève avec la réponse attendue.
Réponds UNIQUEMENT en JSON: {"score": 0.0-1.0, "is_correct": true/false, "explanation": "..."}

Règles de scoring:
- 1.0 : réponse complète et correcte (même si formulée différemment)
- 0.7-0.9 : réponse partiellement correcte (idée principale présente, détails manquants)
- 0.3-0.6 : réponse vague ou très incomplète
- 0.0 : réponse fausse, hors-sujet, ou vide
- Sois tolérant sur l'orthographe et la formulation, juge le FOND pas la FORME`

// AnswerScorer implements session.Scorer using an OpenAI-compatible LLM.
type AnswerScorer struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
}

// NewAnswerScorer creates a new AnswerScorer for an OpenAI-compatible API.
func NewAnswerScorer(baseURL, apiKey, model string) *AnswerScorer {
	return &AnswerScorer{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
	}
}

// Compile-time check that AnswerScorer implements session.Scorer.
var _ session.Scorer = (*AnswerScorer)(nil)

// ScoreAnswer sends the student answer to the LLM for evaluation.
func (s *AnswerScorer) ScoreAnswer(ctx context.Context, prompt string, expectedAnswer string, studentAnswer string) (*session.LLMScoreResult, error) {
	userPrompt := fmt.Sprintf("Question: %s\nRéponse attendue: %s\nRéponse de l'élève: %s", prompt, expectedAnswer, studentAnswer)

	reqBody := chatRequest{
		Model: s.model,
		Messages: []chatMessage{
			{Role: "system", Content: scoringSystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   256,
		Temperature: ptrFloat(0.0),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: marshal request: %w", err)
	}

	url := s.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: no choices in response")
	}

	text := chatResp.Choices[0].Message.Content
	if text == "" {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: empty response")
	}

	// Parse JSON response (strip markdown fences if present)
	var raw struct {
		Score       float64 `json:"score"`
		IsCorrect   bool    `json:"is_correct"`
		Explanation string  `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(stripMarkdownFences(text)), &raw); err != nil {
		return nil, fmt.Errorf("openaicompat.ScoreAnswer: parse response JSON: %w", err)
	}

	return &session.LLMScoreResult{
		Score:       raw.Score,
		IsCorrect:   raw.IsCorrect,
		Explanation: raw.Explanation,
	}, nil
}
