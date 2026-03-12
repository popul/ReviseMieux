package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// Structurer implements chapter.LLMService using the Anthropic Claude API.
type Structurer struct {
	client anthropic.Client
	model  string
}

// NewStructurer creates a new Structurer with the given API key and model.
func NewStructurer(apiKey, model string) *Structurer {
	return &Structurer{
		client: newClient(apiKey),
		model:  model,
	}
}

// StructureBlocks takes OCR text blocks and produces structured items and notions.
func (s *Structurer) StructureBlocks(ctx context.Context, subject string, blocks []chapter.OCRBlock) (*chapter.StructurationResult, error) {
	// Serialize blocks for the prompt
	blocksJSON, err := json.Marshal(blocksToPromptFormat(blocks))
	if err != nil {
		return nil, fmt.Errorf("anthropic.StructureBlocks: marshal blocks: %w", err)
	}

	userPrompt := BuildUserPrompt(subject, string(blocksJSON))

	start := time.Now()

	msg, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: StructurationSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic.StructureBlocks: API call: %w", err)
	}

	_ = time.Since(start) // latency available if needed for logging

	// Extract text from response
	text := extractText(msg)
	if text == "" {
		return nil, fmt.Errorf("anthropic.StructureBlocks: empty response")
	}

	// Parse JSON response
	var raw rawStructurationResponse
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, fmt.Errorf("anthropic.StructureBlocks: parse response: %w", err)
	}

	// Map to domain types
	result := &chapter.StructurationResult{
		Notions: raw.Notions,
	}
	for _, item := range raw.Items {
		itemType, err := chapter.ParseItemType(item.Type)
		if err != nil {
			continue // skip unknown types
		}
		_ = itemType.Valid() // already validated by ParseItemType
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

// ModelVersion returns the model identifier for versioning (Z2-AC13).
func (s *Structurer) ModelVersion() string {
	return s.model
}

// PromptVersion returns the prompt template version for versioning (Z2-AC13).
func (s *Structurer) PromptVersion() string {
	return StructurationPromptVersion
}

// PromptHash returns the prompt content hash for drift detection.
func (s *Structurer) PromptHash() string {
	return StructurationPromptHash()
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

func extractText(msg *anthropic.Message) string {
	for _, block := range msg.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			return b.Text
		}
	}
	return ""
}
