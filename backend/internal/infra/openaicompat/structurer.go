package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/infra/llm"
)

// Structurer implements chapter.LLMService using an OpenAI-compatible API.
// This covers Google Gemini, OpenAI, Mistral, and other compatible providers.
type Structurer struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
}

// NewStructurer creates a new Structurer for an OpenAI-compatible API.
func NewStructurer(baseURL, apiKey, model string) *Structurer {
	return &Structurer{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
	}
}

// StructureBlocks takes OCR text blocks and produces structured items and notions.
func (s *Structurer) StructureBlocks(ctx context.Context, subject string, blocks []chapter.OCRBlock) (*chapter.StructurationResult, error) {
	blocksJSON, err := json.Marshal(blocksToPromptFormat(blocks))
	if err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: marshal blocks: %w", err)
	}

	userPrompt := llm.BuildUserPrompt(subject, string(blocksJSON))

	reqBody := chatRequest{
		Model: s.model,
		Messages: []chatMessage{
			{Role: "system", Content: llm.StructurationSystemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   4096,
		Temperature: ptrFloat(0.0),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: marshal request: %w", err)
	}

	url := s.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: no choices in response")
	}

	text := chatResp.Choices[0].Message.Content
	if text == "" {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: empty response")
	}

	// Parse JSON response (strip markdown fences if present)
	var raw rawStructurationResponse
	if err := json.Unmarshal([]byte(stripMarkdownFences(text)), &raw); err != nil {
		return nil, fmt.Errorf("openaicompat.StructureBlocks: parse response JSON: %w", err)
	}

	result := &chapter.StructurationResult{
		Notions: raw.Notions,
	}
	for _, item := range raw.Items {
		itemType, err := chapter.ParseItemType(item.Type)
		if err != nil {
			continue // skip unknown types
		}
		result.Items = append(result.Items, chapter.StructuredItem{
			Type:       itemType,
			Term:       item.Term,
			Keywords:   item.Keywords,
			Steps:      item.Steps,
			NotionName: item.NotionName,
			Confidence: item.Confidence,
		})
	}

	return result, nil
}

// ModelVersion returns the model identifier for versioning.
func (s *Structurer) ModelVersion() string {
	return s.model
}

// PromptVersion returns the prompt template version for versioning.
func (s *Structurer) PromptVersion() string {
	return llm.StructurationPromptVersion
}

// PromptHash returns the prompt content hash for drift detection.
func (s *Structurer) PromptHash() string {
	return llm.StructurationPromptHash()
}

// --- internal types ---

type promptBlock struct {
	Text       string  `json:"text"`
	BlockType  string  `json:"block_type"`
	Confidence float32 `json:"confidence"`
}

func blocksToPromptFormat(blocks []chapter.OCRBlock) []promptBlock {
	result := make([]promptBlock, len(blocks))
	for i, b := range blocks {
		result[i] = promptBlock{
			Text:       b.Text,
			BlockType:  string(b.BlockType),
			Confidence: b.Confidence,
		}
	}
	return result
}

type rawStructurationResponse struct {
	Items   []rawItem `json:"items"`
	Notions []string  `json:"notions"`
}

type rawItem struct {
	Type       string   `json:"type"`
	Term       string   `json:"term"`
	Keywords   []string `json:"keywords"`
	Steps      []string `json:"steps"`
	NotionName string   `json:"notion_name"`
	Confidence float32  `json:"confidence"`
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) < 2 {
		return s
	}
	s = lines[1]
	if idx := strings.LastIndex(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

func ptrFloat(f float64) *float64 {
	return &f
}
