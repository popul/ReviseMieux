// Package ocr implements the chapter.OCRService port using Gemini Flash VLM
// via its OpenAI-compatible endpoint.
package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/infra/llm"
)

const (
	geminiBaseURL      = "https://generativelanguage.googleapis.com/v1beta/openai"
	defaultModel       = "gemini-2.5-flash-preview-05-20"
	defaultTimeout     = 90 * time.Second
	defaultMaxTokens   = 8192
)

// GeminiOCR implements chapter.OCRService using Gemini Flash as a VLM (direct vision).
type GeminiOCR struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
}

// Config holds the configuration for the Gemini OCR adapter.
type Config struct {
	APIKey  string
	BaseURL string // Override (empty = default Gemini endpoint)
	Model   string // Override (empty = default model)
}

// NewGeminiOCR creates a new Gemini-based OCR processor.
func NewGeminiOCR(cfg Config) *GeminiOCR {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = geminiBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	return &GeminiOCR{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		model:      model,
	}
}

// ProcessPage implements chapter.OCRService.
// It downloads the image from imageURL, sends it to Gemini Vision, and parses
// the response into OCRResult blocks.
func (g *GeminiOCR) ProcessPage(ctx context.Context, imageURL string) (*chapter.OCRResult, error) {
	// 1. Download the image
	imgReq, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: build image request: %w", err)
	}
	imgResp, err := g.httpClient.Do(imgReq)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: download image %s: %w", imageURL, err)
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ocr.gemini: download image %s: HTTP %d", imageURL, imgResp.StatusCode)
	}

	imgData, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: read image body: %w", err)
	}

	// Detect content type from response or magic bytes
	contentType := imgResp.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(imgData)
	}

	// 2. Encode image as base64 data URL
	import_b64 := encodeBase64(imgData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, import_b64)

	// 3. Build the chat completion request with vision
	systemPrompt, _ := json.Marshal(llm.OCRSystemPrompt)
	userContent := []visionContent{
		{
			Type: "image_url",
			ImageURL: &visionImageURL{URL: dataURL},
		},
		{
			Type: "text",
			Text: llm.BuildOCRUserPrompt(""),
		},
	}
	userContentJSON, _ := json.Marshal(userContent)

	temp := 0.0
	reqBody := visionRequest{
		Model: g.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContentJSON},
		},
		MaxTokens:   defaultMaxTokens,
		Temperature: &temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: marshal request: %w", err)
	}

	// 4. Call the API
	url := g.baseURL + "/chat/completions"
	apiReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: build API request: %w", err)
	}
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	apiResp, err := g.httpClient.Do(apiReq)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: API call: %w", err)
	}
	defer apiResp.Body.Close()

	respBody, err := io.ReadAll(apiResp.Body)
	if err != nil {
		return nil, fmt.Errorf("ocr.gemini: read API response: %w", err)
	}

	if apiResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ocr.gemini: API HTTP %d: %s", apiResp.StatusCode, truncate(string(respBody), 500))
	}

	// 5. Parse the response
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("ocr.gemini: parse API response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("ocr.gemini: no choices in response")
	}

	text := chatResp.Choices[0].Message.Content

	// 6. Parse the OCR JSON output into domain blocks
	return parseOCRResponse(text)
}

// --- Response parsing ---

type ocrResponseJSON struct {
	Blocks []ocrBlockJSON `json:"blocks"`
}

type ocrBlockJSON struct {
	Text       string  `json:"text"`
	BlockType  string  `json:"block_type"`
	Confidence float64 `json:"confidence"`
}

func parseOCRResponse(text string) (*chapter.OCRResult, error) {
	var resp ocrResponseJSON
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		return nil, fmt.Errorf("ocr.gemini: parse OCR JSON: %w (raw: %s)", err, truncate(text, 200))
	}

	result := &chapter.OCRResult{
		Blocks: make([]chapter.OCRBlock, 0, len(resp.Blocks)),
	}
	for _, b := range resp.Blocks {
		bt := chapter.BlockType(b.BlockType)
		if !bt.Valid() {
			// Map non-standard types: DIAGRAM -> SCHEMA (the prompt uses DIAGRAM)
			switch b.BlockType {
			case "DIAGRAM":
				bt = chapter.BlockSchema
			default:
				bt = chapter.BlockText
			}
		}
		result.Blocks = append(result.Blocks, chapter.OCRBlock{
			Text:       b.Text,
			BlockType:  bt,
			Confidence: float32(b.Confidence),
		})
	}

	return result, nil
}

// --- API types (OpenAI-compatible) ---

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

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// --- Helpers ---

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
