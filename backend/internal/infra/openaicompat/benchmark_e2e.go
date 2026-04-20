package openaicompat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"

	"github.com/popul/revisemieux/internal/benchmark"
	"github.com/popul/revisemieux/internal/infra/llm"
)

// maxTokenBudget is the maximum total token budget for all images in a single API request.
// Mistral rate-limits at 1M tokens/min. We target 800K to leave room for text prompt + OCR blocks.
const maxTokenBudget = 800_000

// approxTokensPerByte estimates ~1 token per 3 bytes of base64 content (empirical).
const approxTokensPerByte = 0.33

// encodeImagesForAPI encodes all images for an API call, adaptively reducing resolution
// to stay within the token budget. Returns base64-encoded images with their media types.
func encodeImagesForAPI(imagePaths []string) ([]string, []string, error) {
	// Start with max 2048px. If total exceeds budget, reduce to 1536, 1280, 1024.
	for _, maxDim := range []int{2048, 1536, 1280, 1024} {
		var encodedImages []string
		var mediaTypes []string
		var totalBytes int

		for _, imgPath := range imagePaths {
			encoded, mediaType, err := encodeImageWithMaxDim(imgPath, maxDim)
			if err != nil {
				return nil, nil, err
			}
			encodedImages = append(encodedImages, encoded)
			mediaTypes = append(mediaTypes, mediaType)
			totalBytes += len(encoded)
		}

		approxTokens := int(float64(totalBytes) * approxTokensPerByte)
		if approxTokens <= maxTokenBudget {
			return encodedImages, mediaTypes, nil
		}
	}

	// Fallback: 1024px (always use this if nothing fits)
	var encodedImages []string
	var mediaTypes []string
	for _, imgPath := range imagePaths {
		encoded, mediaType, err := encodeImageWithMaxDim(imgPath, 1024)
		if err != nil {
			return nil, nil, err
		}
		encodedImages = append(encodedImages, encoded)
		mediaTypes = append(mediaTypes, mediaType)
	}
	return encodedImages, mediaTypes, nil
}

// encodeImageWithMaxDim reads an image, resizes to fit within maxDim pixels, returns base64.
func encodeImageWithMaxDim(imgPath string, maxDim int) (string, string, error) {
	data, err := os.ReadFile(imgPath)
	if err != nil {
		return "", "", err
	}

	// Decode to check dimensions
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// Can't decode — send original
		return base64.StdEncoding.EncodeToString(data), detectMediaType(imgPath), nil
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// If already small enough, re-encode at quality 85 (still smaller than raw JPEG from phone)
	if w <= maxDim && h <= maxDim {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return base64.StdEncoding.EncodeToString(data), detectMediaType(imgPath), nil
		}
		return base64.StdEncoding.EncodeToString(buf.Bytes()), "image/jpeg", nil
	}

	// Resize: scale down to fit within maxDim, preserving aspect ratio
	scale := float64(maxDim) / float64(max(w, h))
	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)

	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	// Nearest-neighbor resize (fast, preserves text edges better than bilinear for OCR)
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			resized.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return base64.StdEncoding.EncodeToString(data), detectMediaType(imgPath), nil
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), "image/jpeg", nil
}

// Compile-time checks.
var _ benchmark.E2EProvider = (*E2EBenchmarkProvider)(nil)
var _ benchmark.HybridProvider = (*E2EBenchmarkProvider)(nil)

// E2EBenchmarkProvider implements benchmark.E2EProvider for end-to-end
// vision → structured items extraction via an OpenAI-compatible API.
type E2EBenchmarkProvider struct {
	ocr      *OCRProcessor // reuse vision call infrastructure
	name     string
	model    string
	priceIn  float64
	priceOut float64
}

// NewE2EBenchmarkProvider creates an E2E benchmark provider.
func NewE2EBenchmarkProvider(cfg Config) *E2EBenchmarkProvider {
	return &E2EBenchmarkProvider{
		ocr:      NewOCRProcessor(cfg),
		name:     cfg.Name,
		model:    cfg.Model,
		priceIn:  cfg.PriceIn,
		priceOut: cfg.PriceOut,
	}
}

// NewE2EFromIDP creates an E2E/Hybrid provider from a text-only IDP provider.
// Only works with --no-images (classic pipeline mode). StructureHybrid sends text only.
func NewE2EFromIDP(idpProv benchmark.Provider) *E2EBenchmarkProvider {
	bp, ok := idpProv.(*BenchmarkProvider)
	if !ok {
		return nil
	}
	// Create a minimal OCR processor (for doVisionRequest) using the same config
	proc := &OCRProcessor{
		httpClient: bp.httpClient,
		baseURL:    bp.baseURL,
		apiKey:     bp.apiKey,
		name:       bp.name,
		model:      bp.model,
	}
	return &E2EBenchmarkProvider{
		ocr:      proc,
		name:     bp.name,
		model:    bp.model,
		priceIn:  bp.priceIn,
		priceOut: bp.priceOut,
	}
}

// NewE2EFromOCR creates an E2E provider by wrapping an existing OCR provider.
// This allows auto-generating E2E builders from the OCR builder config.
func NewE2EFromOCR(ocrProv interface{}) *E2EBenchmarkProvider {
	switch p := ocrProv.(type) {
	case *OCRBenchmarkProvider:
		return &E2EBenchmarkProvider{
			ocr:     p.processor,
			name:    p.name,
			model:   p.model,
			priceIn: p.priceIn,
			priceOut: p.priceOut,
		}
	default:
		// Non-openaicompat providers (Anthropic, Mistral OCR) can't be wrapped — return nil.
		return nil
	}
}

func (p *E2EBenchmarkProvider) Name() string             { return p.name }
func (p *E2EBenchmarkProvider) ModelID() string          { return p.model }
func (p *E2EBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *E2EBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// StructureImages sends images + E2E structuration prompt to the VLM and returns structured items JSON.
func (p *E2EBenchmarkProvider) StructureImages(ctx context.Context, imagePaths []string, subject string) (*benchmark.Response, error) {
	encodedImages, mediaTypes, err := encodeImagesForAPI(imagePaths)
	if err != nil {
		return nil, fmt.Errorf("%s e2e: encode images: %w", p.name, err)
	}
	var contentParts []visionContent
	for i, encoded := range encodedImages {
		contentParts = append(contentParts, visionContent{
			Type: "image_url",
			ImageURL: &visionImageURL{
				URL: fmt.Sprintf("data:%s;base64,%s", mediaTypes[i], encoded),
			},
		})
	}
	contentParts = append(contentParts, visionContent{
		Type: "text",
		Text: llm.BuildE2EUserPrompt(subject),
	})

	userContent, _ := json.Marshal(contentParts)
	systemContent, _ := json.Marshal(llm.E2ESystemPrompt)

	reqBody := visionRequest{
		Model: p.ocr.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
	}
	if p.ocr.disableThinking {
		temp := 0.0
		reqBody.MaxTokens = 16384
		reqBody.Temperature = &temp
		reqBody.ReasoningEffort = "none"
	} else if p.ocr.reasoningModel {
		reqBody.MaxCompletionTokens = 16384
	} else {
		temp := 0.0
		reqBody.MaxTokens = 16384
		reqBody.Temperature = &temp
	}

	respBody, latency, err := p.ocr.doVisionRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s e2e: %w", p.name, err)
	}

	text := respBody.Choices[0].Message.Content
	if text == "" && respBody.Choices[0].Message.ReasoningContent != "" {
		text = respBody.Choices[0].Message.ReasoningContent
	}
	if text == "" {
		return nil, fmt.Errorf("%s e2e: empty content in response", p.name)
	}

	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  respBody.Usage.PromptTokens,
		TokensOutput: respBody.Usage.CompletionTokens,
		LatencyMs:    latency,
		ModelVersion: p.model,
	}, nil
}

// StructureHybrid sends images + OCR blocks to the VLM with the hybrid prompt.
// This is the hybrid pipeline: real OCR output + original images → structured items.
func (p *E2EBenchmarkProvider) StructureHybrid(ctx context.Context, imagePaths []string, blocksJSON string, subject string) (*benchmark.Response, error) {
	if p.ocr == nil {
		return nil, fmt.Errorf("%s hybrid: no vision processor configured", p.name)
	}

	var contentParts []visionContent
	// Only encode and send images if imagePaths is non-empty (hybrid mode).
	// When nil/empty, we're in classic pipeline mode (OCR text only, no images).
	if len(imagePaths) > 0 {
		encodedImages, mediaTypes, err := encodeImagesForAPI(imagePaths)
		if err != nil {
			return nil, fmt.Errorf("%s hybrid: encode images: %w", p.name, err)
		}
		for i, encoded := range encodedImages {
			contentParts = append(contentParts, visionContent{
				Type: "image_url",
				ImageURL: &visionImageURL{
					URL: fmt.Sprintf("data:%s;base64,%s", mediaTypes[i], encoded),
				},
			})
		}
	}
	contentParts = append(contentParts, visionContent{
		Type: "text",
		Text: llm.BuildHybridUserPrompt(subject, blocksJSON),
	})

	userContent, _ := json.Marshal(contentParts)
	systemContent, _ := json.Marshal(llm.HybridSystemPrompt)

	reqBody := visionRequest{
		Model: p.ocr.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
	}
	if p.ocr.disableThinking {
		temp := 0.0
		reqBody.MaxTokens = 16384
		reqBody.Temperature = &temp
		reqBody.ReasoningEffort = "none"
	} else if p.ocr.reasoningModel {
		reqBody.MaxCompletionTokens = 16384
	} else {
		temp := 0.0
		reqBody.MaxTokens = 16384
		reqBody.Temperature = &temp
	}

	respBody, latency, err := p.ocr.doVisionRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s hybrid: %w", p.name, err)
	}

	text := respBody.Choices[0].Message.Content
	if text == "" && respBody.Choices[0].Message.ReasoningContent != "" {
		text = respBody.Choices[0].Message.ReasoningContent
	}
	if text == "" {
		return nil, fmt.Errorf("%s hybrid: empty content in response", p.name)
	}

	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  respBody.Usage.PromptTokens,
		TokensOutput: respBody.Usage.CompletionTokens,
		LatencyMs:    latency,
		ModelVersion: p.model,
	}, nil
}
