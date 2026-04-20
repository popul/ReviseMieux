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

	"github.com/popul/revisemieux/internal/infra/llm"
)

// OCRProcessor implements OCR processing using an OpenAI-compatible vision API.
type OCRProcessor struct {
	httpClient      *http.Client
	baseURL         string
	apiKey          string
	name            string
	model           string
	reasoningModel  bool
	disableThinking bool
	perPage         bool
}

// NewOCRProcessor creates an OCR processor for an OpenAI-compatible vision API.
func NewOCRProcessor(cfg Config) *OCRProcessor {
	return &OCRProcessor{
		httpClient:      &http.Client{Timeout: 5 * time.Minute},
		baseURL:         cfg.BaseURL,
		apiKey:          cfg.APIKey,
		name:            cfg.Name,
		model:           cfg.Model,
		reasoningModel:  cfg.ReasoningModel,
		disableThinking: cfg.DisableThinking,
		perPage:         cfg.PerPage,
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
// If PerPage is true, processes images one at a time and merges JSON blocks.
func (p *OCRProcessor) ProcessPages(ctx context.Context, imagePaths []string, subject string) ([]byte, *RawOCRResponse, error) {
	if p.perPage && len(imagePaths) > 1 {
		return p.processPerPage(ctx, imagePaths, subject)
	}
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
		Text: llm.BuildOCRUserPrompt(subject),
	})

	userContent, _ := json.Marshal(contentParts)
	// System prompt from production source of truth
	systemContent, _ := json.Marshal(llm.OCRSystemPrompt)

	reqBody := visionRequest{
		Model: p.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
	}
	if p.disableThinking {
		temp := 0.0
		reqBody.MaxTokens = 8192
		reqBody.Temperature = &temp
		reqBody.ReasoningEffort = "none"
	} else if p.reasoningModel {
		reqBody.MaxCompletionTokens = 16384
	} else {
		temp := 0.0
		reqBody.MaxTokens = 8192
		reqBody.Temperature = &temp
	}

	chatResp, latency, err := p.doVisionRequest(ctx, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("%s ocr: %w", p.name, err)
	}

	text := chatResp.Choices[0].Message.Content
	// Reasoning models may put the answer in reasoning_content
	if text == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		text = chatResp.Choices[0].Message.ReasoningContent
	}

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
	Model               string          `json:"model"`
	Messages            []visionMessage `json:"messages"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`
}

// processPerPage processes images one at a time and merges the OCR JSON blocks.
func (p *OCRProcessor) processPerPage(ctx context.Context, imagePaths []string, subject string) ([]byte, *RawOCRResponse, error) {
	type ocrBlock struct {
		Text       string  `json:"text"`
		BlockType  string  `json:"block_type"`
		Confidence float64 `json:"confidence"`
	}

	var allBlocks []ocrBlock
	var totalTokensIn, totalTokensOut int
	var totalLatency int64

	for _, imgPath := range imagePaths {
		_, raw, err := p.ProcessPages(ctx, []string{imgPath}, subject)
		if err != nil {
			continue // skip failed pages
		}
		totalTokensIn += raw.TokensInput
		totalTokensOut += raw.TokensOutput
		totalLatency += raw.LatencyMs

		// Try to parse blocks from this page
		var wrapper struct {
			Blocks []ocrBlock `json:"blocks"`
		}
		cleaned := bytes.TrimSpace(raw.RawJSON)
		// Strip markdown fences if present
		if bytes.HasPrefix(cleaned, []byte("```")) {
			if idx := bytes.Index(cleaned[3:], []byte("\n")); idx >= 0 {
				cleaned = cleaned[3+idx+1:]
			}
			if bytes.HasSuffix(cleaned, []byte("```")) {
				cleaned = cleaned[:len(cleaned)-3]
			}
			cleaned = bytes.TrimSpace(cleaned)
		}
		if err := json.Unmarshal(cleaned, &wrapper); err == nil {
			allBlocks = append(allBlocks, wrapper.Blocks...)
		} else {
			// If not JSON, treat each line as a text block
			for _, line := range bytes.Split(raw.RawJSON, []byte("\n")) {
				line = bytes.TrimSpace(line)
				if len(line) > 0 {
					allBlocks = append(allBlocks, ocrBlock{
						Text:       string(line),
						BlockType:  "TEXT",
						Confidence: 0.8,
					})
				}
			}
		}
	}

	if len(allBlocks) == 0 {
		return nil, nil, fmt.Errorf("%s ocr per-page: no blocks extracted from %d images", p.name, len(imagePaths))
	}

	// Build merged JSON
	result, _ := json.Marshal(map[string]interface{}{"blocks": allBlocks})

	raw := &RawOCRResponse{
		RawJSON:      result,
		TokensInput:  totalTokensIn,
		TokensOutput: totalTokensOut,
		LatencyMs:    totalLatency,
	}
	return result, raw, nil
}

// doVisionRequest sends a vision request and returns the parsed response + latency.
func (p *OCRProcessor) doVisionRequest(ctx context.Context, reqBody visionRequest) (*chatResponse, int64, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal: %w", err)
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s call: %w", p.name, err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, 0, fmt.Errorf("parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, 0, fmt.Errorf("no choices in response")
	}

	return &chatResp, latency, nil
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
