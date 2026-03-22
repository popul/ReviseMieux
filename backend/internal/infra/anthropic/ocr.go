package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdkanthro "github.com/anthropics/anthropic-sdk-go"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// OCRProcessor implements chapter.OCRService using the Anthropic vision API.
type OCRProcessor struct {
	client  sdkanthro.Client
	model   string
	subject string // default subject for context
}

// NewOCRProcessor creates an OCR processor using the Anthropic vision API.
func NewOCRProcessor(apiKey, model string) *OCRProcessor {
	return &OCRProcessor{
		client: newClient(apiKey),
		model:  model,
	}
}

// ProcessPage implements chapter.OCRService — extracts OCR blocks from a single image URL.
// In production, imageURL is an S3/storage URL. The adapter downloads the image and sends it
// to the vision API.
func (p *OCRProcessor) ProcessPage(ctx context.Context, imageURL string) (*chapter.OCRResult, error) {
	// Read image data — supports both local files and URLs
	imageData, mediaType, err := readImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("anthropic.ocr: read image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)

	contentBlocks := []sdkanthro.ContentBlockParamUnion{
		sdkanthro.NewImageBlockBase64(mediaType, encoded),
		sdkanthro.NewTextBlock(BuildOCRUserPrompt("")),
	}

	start := time.Now()

	msg, err := p.client.Messages.New(ctx, sdkanthro.MessageNewParams{
		Model:     sdkanthro.Model(p.model),
		MaxTokens: 8192,
		System: []sdkanthro.TextBlockParam{
			{Text: OCRSystemPrompt},
		},
		Messages: []sdkanthro.MessageParam{
			sdkanthro.NewUserMessage(contentBlocks...),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic.ocr: API call: %w", err)
	}

	_ = time.Since(start)

	text := extractText(msg)
	if text == "" {
		return nil, fmt.Errorf("anthropic.ocr: empty response")
	}

	// Parse response into OCR blocks
	return parseOCRResponse(text)
}

// ProcessPages extracts OCR blocks from multiple images in a single call.
// Returns raw response metadata for benchmark usage.
func (p *OCRProcessor) ProcessPages(ctx context.Context, imagePaths []string, subject string) (*chapter.OCRResult, *RawOCRResponse, error) {
	var contentBlocks []sdkanthro.ContentBlockParamUnion
	for _, imgPath := range imagePaths {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, nil, fmt.Errorf("anthropic.ocr: read image %s: %w", imgPath, err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		mediaType := detectMediaType(imgPath)
		contentBlocks = append(contentBlocks, sdkanthro.NewImageBlockBase64(mediaType, encoded))
	}
	contentBlocks = append(contentBlocks, sdkanthro.NewTextBlock(BuildOCRUserPrompt(subject)))

	start := time.Now()

	msg, err := p.client.Messages.New(ctx, sdkanthro.MessageNewParams{
		Model:     sdkanthro.Model(p.model),
		MaxTokens: 8192,
		System: []sdkanthro.TextBlockParam{
			{Text: OCRSystemPrompt},
		},
		Messages: []sdkanthro.MessageParam{
			sdkanthro.NewUserMessage(contentBlocks...),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic.ocr: API call: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	text := extractText(msg)
	if text == "" {
		return nil, nil, fmt.Errorf("anthropic.ocr: empty response")
	}

	ocrResult, err := parseOCRResponse(text)
	if err != nil {
		return nil, &RawOCRResponse{
			RawJSON:      []byte(text),
			TokensInput:  int(msg.Usage.InputTokens),
			TokensOutput: int(msg.Usage.OutputTokens),
			LatencyMs:    latency,
		}, err
	}

	return ocrResult, &RawOCRResponse{
		RawJSON:      []byte(text),
		TokensInput:  int(msg.Usage.InputTokens),
		TokensOutput: int(msg.Usage.OutputTokens),
		LatencyMs:    latency,
	}, nil
}

// RawOCRResponse holds raw API response metadata (for benchmark instrumentation).
type RawOCRResponse struct {
	RawJSON      []byte
	TokensInput  int
	TokensOutput int
	LatencyMs    int64
}

// ModelVersion returns the model identifier for versioning.
func (p *OCRProcessor) ModelVersion() string {
	return p.model
}

// PromptVersion returns the OCR prompt version.
func (p *OCRProcessor) PromptVersion() string {
	return OCRPromptVersion
}

// PromptHash returns the OCR prompt content hash.
func (p *OCRProcessor) PromptHash() string {
	return OCRPromptHash()
}

// --- internal helpers ---

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

// readImage reads image data from a local file path or URL.
func readImage(imageURL string) ([]byte, string, error) {
	// For local files, read directly
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		data, err := os.ReadFile(imageURL)
		if err != nil {
			return nil, "", err
		}
		return data, detectMediaType(imageURL), nil
	}

	// For remote URLs, we'd use http.Get — but for now return an error
	// since production will use pre-downloaded images via Storage adapter
	return nil, "", fmt.Errorf("remote URL download not yet implemented: %s", imageURL)
}

// rawOCRBlock is the JSON structure returned by the LLM for OCR.
type rawOCRBlock struct {
	Text       string  `json:"text"`
	BlockType  string  `json:"block_type"`
	Confidence float32 `json:"confidence"`
}

func parseOCRResponse(text string) (*chapter.OCRResult, error) {
	cleaned := stripMarkdownFences(text)

	// Try {"blocks": [...]}
	var wrapper struct {
		Blocks []rawOCRBlock `json:"blocks"`
	}
	if err := json.Unmarshal([]byte(cleaned), &wrapper); err == nil && len(wrapper.Blocks) > 0 {
		return toOCRResult(wrapper.Blocks), nil
	}

	// Try direct array [...]
	var blocks []rawOCRBlock
	if err := json.Unmarshal([]byte(cleaned), &blocks); err == nil && len(blocks) > 0 {
		return toOCRResult(blocks), nil
	}

	return nil, fmt.Errorf("anthropic.ocr: failed to parse OCR response")
}

// stripMarkdownFences removes ```json ... ``` fences from LLM responses.
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

func toOCRResult(raw []rawOCRBlock) *chapter.OCRResult {
	result := &chapter.OCRResult{}
	for _, b := range raw {
		bt := chapter.BlockText
		switch strings.ToUpper(b.BlockType) {
		case "DIAGRAM", "SCHEMA", "MAP":
			bt = chapter.BlockSchema
		}
		result.Blocks = append(result.Blocks, chapter.OCRBlock{
			Text:       b.Text,
			BlockType:  bt,
			Confidence: b.Confidence,
		})
	}
	return result
}
