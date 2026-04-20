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
	"strings"
	"time"

	"github.com/popul/revisemieux/internal/benchmark"
)

// Compile-time check.
var _ benchmark.OCRProvider = (*RawOCRBenchmarkProvider)(nil)

// RawOCRBenchmarkProvider wraps a lightweight VLM (like PaddleOCR-VL) that returns
// raw text instead of structured JSON blocks. It sends a simple prompt and wraps
// each paragraph of the response as a TEXT block with confidence 0.8.
type RawOCRBenchmarkProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	name       string
	model      string
	priceIn    float64
	priceOut   float64
}

// NewRawOCRBenchmarkProvider creates a benchmark OCR provider for lightweight VLMs
// that can't follow complex structured output prompts.
func NewRawOCRBenchmarkProvider(cfg Config) *RawOCRBenchmarkProvider {
	return &RawOCRBenchmarkProvider{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		name:       cfg.Name,
		model:      cfg.Model,
		priceIn:    cfg.PriceIn,
		priceOut:   cfg.PriceOut,
	}
}

func (p *RawOCRBenchmarkProvider) Name() string             { return p.name }
func (p *RawOCRBenchmarkProvider) ModelID() string          { return p.model }
func (p *RawOCRBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *RawOCRBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

func (p *RawOCRBenchmarkProvider) ExtractBlocks(ctx context.Context, imagePaths []string, _ string) (*benchmark.Response, error) {
	// Build multimodal message: images + simple text prompt
	var contentParts []visionContent
	for _, imgPath := range imagePaths {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, fmt.Errorf("%s ocr: read image %s: %w", p.name, imgPath, err)
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
		Text: "Extract all visible text blocks from these notebook photos. Return each block on a separate line.",
	})

	userContent, _ := json.Marshal(contentParts)

	temp := 0.0
	reqBody := visionRequest{
		Model: p.model,
		Messages: []visionMessage{
			{Role: "user", Content: userContent},
		},
		MaxTokens:   4096,
		Temperature: &temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s ocr marshal: %w", p.name, err)
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s ocr request: %w", p.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s ocr call: %w", p.name, err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s ocr read body: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s ocr HTTP %d: %s", p.name, resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("%s ocr parse response: %w", p.name, err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%s ocr: no choices in response", p.name)
	}

	rawText := chatResp.Choices[0].Message.Content
	if rawText == "" {
		return nil, fmt.Errorf("%s ocr: empty content in response", p.name)
	}

	// Convert raw text lines → JSON blocks format expected by the benchmark evaluator
	blocksJSON := rawTextToBlocks(rawText)

	return &benchmark.Response{
		RawJSON:      blocksJSON,
		TokensInput:  chatResp.Usage.PromptTokens,
		TokensOutput: chatResp.Usage.CompletionTokens,
		LatencyMs:    latency,
		ModelVersion: p.model,
	}, nil
}

// rawTextToBlocks converts raw text output into the JSON blocks format:
// {"blocks": [{"text": "...", "block_type": "TEXT", "confidence": 0.8}, ...]}
func rawTextToBlocks(rawText string) []byte {
	type block struct {
		Text       string  `json:"text"`
		BlockType  string  `json:"block_type"`
		Confidence float64 `json:"confidence"`
	}

	var blocks []block
	for _, line := range strings.Split(rawText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		blocks = append(blocks, block{
			Text:       line,
			BlockType:  "TEXT",
			Confidence: 0.8,
		})
	}

	result, _ := json.Marshal(map[string]any{"blocks": blocks})
	return result
}
