package openaicompat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	llmanthro "github.com/popul/revisemieux/internal/infra/anthropic"
)

// OCRProcessor implements OCR processing using an OpenAI-compatible vision API.
type OCRProcessor struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	name       string
	model      string
}

// NewOCRProcessor creates an OCR processor for an OpenAI-compatible vision API.
func NewOCRProcessor(cfg Config) *OCRProcessor {
	return &OCRProcessor{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		name:       cfg.Name,
		model:      cfg.Model,
	}
}

// RawOCRResponse holds raw API response metadata (for benchmark instrumentation).
type RawOCRResponse struct {
	RawJSON      []byte
	TokensInput  int
	TokensOutput int
	LatencyMs    int64
}

// ProcessPages sends images to the vision API and returns raw response.
// Uses the production prompts from infra/anthropic/prompts.go.
func (p *OCRProcessor) ProcessPages(ctx context.Context, imagePaths []string, subject string) ([]byte, *RawOCRResponse, error) {
	// Build user message content: images + text instruction
	var contentParts []visionContent
	for _, imgPath := range imagePaths {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, nil, fmt.Errorf("%s ocr: read image %s: %w", p.name, imgPath, err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		mediaType := detectMediaType(imgPath)
		contentParts = append(contentParts, visionContent{
			Type: "image_url",
			ImageURL: &visionImageURL{
				URL: fmt.Sprintf("data:%s;base64,%s", mediaType, encoded),
			},
		})
	}
	contentParts = append(contentParts, visionContent{
		Type: "text",
		Text: llmanthro.BuildOCRUserPrompt(subject),
	})

	userContent, _ := json.Marshal(contentParts)
	// System prompt from production source of truth
	systemContent, _ := json.Marshal(llmanthro.OCRSystemPrompt)

	temp := 0.0
	reqBody := visionRequest{
		Model: p.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   8192,
		Temperature: &temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("%s ocr marshal: %w", p.name, err)
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("%s ocr request: %w", p.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s ocr call: %w", p.name, err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("%s ocr read body: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("%s ocr HTTP %d: %s", p.name, resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, nil, fmt.Errorf("%s ocr parse response: %w", p.name, err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("%s ocr: no choices in response", p.name)
	}

	text := chatResp.Choices[0].Message.Content

	raw := &RawOCRResponse{
		RawJSON:      []byte(text),
		TokensInput:  chatResp.Usage.PromptTokens,
		TokensOutput: chatResp.Usage.CompletionTokens,
		LatencyMs:    latency,
	}

	return []byte(text), raw, nil
}

// --- Vision message types ---

type visionContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *visionImageURL `json:"image_url,omitempty"`
}

type visionImageURL struct {
	URL string `json:"url"`
}

type visionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type visionRequest struct {
	Model       string          `json:"model"`
	Messages    []visionMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
}

func detectMediaType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
