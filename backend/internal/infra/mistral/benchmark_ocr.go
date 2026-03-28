// Package mistral provides benchmark providers for Mistral's dedicated OCR API.
//
// Mistral OCR uses a separate /v1/ocr endpoint (not chat completions) and
// returns markdown pages. This provider converts the markdown output into
// OCR blocks for benchmark evaluation.
package mistral

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

	"github.com/popul/revisemieux/internal/benchmark"
)

// Compile-time check.
var _ benchmark.OCRProvider = (*OCRBenchmarkProvider)(nil)

// OCRBenchmarkProvider implements benchmark.OCRProvider using Mistral's dedicated OCR API.
type OCRBenchmarkProvider struct {
	httpClient *http.Client
	apiKey     string
	model      string
	priceIn    float64
	priceOut   float64
}

// NewOCRBenchmarkProvider creates a Mistral OCR benchmark provider.
func NewOCRBenchmarkProvider(apiKey, model string, priceIn, priceOut float64) *OCRBenchmarkProvider {
	return &OCRBenchmarkProvider{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		apiKey:     apiKey,
		model:      model,
		priceIn:    priceIn,
		priceOut:   priceOut,
	}
}

func (p *OCRBenchmarkProvider) Name() string            { return "Mistral" }
func (p *OCRBenchmarkProvider) ModelID() string          { return p.model }
func (p *OCRBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *OCRBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// ExtractBlocks sends images to the Mistral OCR API and converts the markdown
// response into OCR blocks for benchmark evaluation.
func (p *OCRBenchmarkProvider) ExtractBlocks(ctx context.Context, imagePaths []string, _ string) (*benchmark.Response, error) {
	var allPages []ocrPage
	totalIn, totalOut := 0, 0
	var totalLatency int64

	// Mistral OCR processes one document at a time.
	for _, imgPath := range imagePaths {
		pages, usage, latency, err := p.callOCR(ctx, imgPath)
		if err != nil {
			return nil, fmt.Errorf("mistral ocr %s: %w", filepath.Base(imgPath), err)
		}
		allPages = append(allPages, pages...)
		totalIn += usage.promptTokens
		totalOut += usage.completionTokens
		totalLatency += latency
	}

	blocks := markdownToBlocks(allPages)
	blocksJSON, err := json.Marshal(map[string]any{"blocks": blocks})
	if err != nil {
		return nil, fmt.Errorf("mistral ocr marshal blocks: %w", err)
	}

	return &benchmark.Response{
		RawJSON:      blocksJSON,
		TokensInput:  totalIn,
		TokensOutput: totalOut,
		LatencyMs:    totalLatency,
		ModelVersion: p.model,
	}, nil
}

// --- Mistral OCR API types ---

type ocrRequest struct {
	Model    string      `json:"model"`
	Document ocrDocument `json:"document"`
}

type ocrDocument struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url,omitempty"`
}

type ocrResponse struct {
	Pages []ocrPage `json:"pages"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

type ocrPage struct {
	Index    int    `json:"index"`
	Markdown string `json:"markdown"`
}

type tokenUsage struct {
	promptTokens     int
	completionTokens int
}

func (p *OCRBenchmarkProvider) callOCR(ctx context.Context, imgPath string) ([]ocrPage, tokenUsage, int64, error) {
	data, err := os.ReadFile(imgPath)
	if err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("read image: %w", err)
	}

	mediaType := detectMediaType(imgPath)
	encoded := base64.StdEncoding.EncodeToString(data)

	reqBody := ocrRequest{
		Model: p.model,
		Document: ocrDocument{
			Type:     "image_url",
			ImageURL: fmt.Sprintf("data:%s;base64,%s", mediaType, encoded),
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.mistral.ai/v1/ocr", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("call: %w", err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, tokenUsage{}, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var ocrResp ocrResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return nil, tokenUsage{}, 0, fmt.Errorf("parse response: %w", err)
	}

	usage := tokenUsage{}
	if ocrResp.Usage != nil {
		usage.promptTokens = ocrResp.Usage.PromptTokens
		usage.completionTokens = ocrResp.Usage.CompletionTokens
	}

	return ocrResp.Pages, usage, latency, nil
}

// markdownToBlocks converts Mistral OCR markdown pages into benchmark OCR blocks.
func markdownToBlocks(pages []ocrPage) []benchmark.OCRBlock {
	var blocks []benchmark.OCRBlock
	for _, page := range pages {
		for _, chunk := range splitMarkdownChunks(page.Markdown) {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			blocks = append(blocks, benchmark.OCRBlock{
				Text:       cleanMarkdown(chunk),
				BlockType:  classifyBlock(chunk),
				Confidence: 1.0,
			})
		}
	}
	return blocks
}

// splitMarkdownChunks splits markdown text into logical chunks, keeping tables together.
func splitMarkdownChunks(md string) []string {
	lines := strings.Split(md, "\n")
	var chunks []string
	var current []string
	inTable := false

	for _, line := range lines {
		isTableLine := strings.HasPrefix(strings.TrimSpace(line), "|")

		if isTableLine && !inTable {
			if text := strings.TrimSpace(strings.Join(current, "\n")); text != "" {
				chunks = append(chunks, text)
			}
			current = []string{line}
			inTable = true
		} else if !isTableLine && inTable {
			if text := strings.TrimSpace(strings.Join(current, "\n")); text != "" {
				chunks = append(chunks, text)
			}
			current = nil
			inTable = false
			if strings.TrimSpace(line) != "" {
				current = append(current, line)
			}
		} else if !inTable && strings.TrimSpace(line) == "" && len(current) > 0 {
			if text := strings.TrimSpace(strings.Join(current, "\n")); text != "" {
				chunks = append(chunks, text)
			}
			current = nil
		} else {
			current = append(current, line)
		}
	}

	if text := strings.TrimSpace(strings.Join(current, "\n")); text != "" {
		chunks = append(chunks, text)
	}
	return chunks
}

func classifyBlock(chunk string) string {
	lines := strings.Split(chunk, "\n")
	tableLines := 0
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "|") {
			tableLines++
		}
	}
	if tableLines > 1 {
		return "TABLE"
	}
	return "TEXT"
}

func cleanMarkdown(text string) string {
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimLeft(l, "# ")
	}
	return strings.Join(lines, "\n")
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
