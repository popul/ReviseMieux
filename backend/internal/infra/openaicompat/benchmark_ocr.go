package openaicompat

import (
	"context"
	"fmt"

	"github.com/popul/revisemieux/internal/benchmark"
)

// Compile-time check.
var _ benchmark.OCRProvider = (*OCRBenchmarkProvider)(nil)

// OCRBenchmarkProvider implements benchmark.OCRProvider using the production OCRProcessor.
type OCRBenchmarkProvider struct {
	processor *OCRProcessor
	name      string
	model     string
	priceIn   float64
	priceOut  float64
}

// NewOCRBenchmarkProvider creates an OCR benchmark provider that wraps the production OCR adapter.
func NewOCRBenchmarkProvider(cfg Config) *OCRBenchmarkProvider {
	return &OCRBenchmarkProvider{
		processor: NewOCRProcessor(cfg),
		name:      cfg.Name,
		model:     cfg.Model,
		priceIn:   cfg.PriceIn,
		priceOut:  cfg.PriceOut,
	}
}

func (p *OCRBenchmarkProvider) Name() string             { return p.name }
func (p *OCRBenchmarkProvider) ModelID() string          { return p.model }
func (p *OCRBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *OCRBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// ExtractBlocks delegates to the production OCRProcessor.ProcessPages and wraps the
// result as a benchmark.Response for instrumentation.
func (p *OCRBenchmarkProvider) ExtractBlocks(ctx context.Context, imagePaths []string, subject string) (*benchmark.Response, error) {
	_, raw, err := p.processor.ProcessPages(ctx, imagePaths, subject)
	if err != nil && raw == nil {
		return nil, fmt.Errorf("%s ocr benchmark: %w", p.name, err)
	}

	resp := &benchmark.Response{
		RawJSON:      raw.RawJSON,
		TokensInput:  raw.TokensInput,
		TokensOutput: raw.TokensOutput,
		LatencyMs:    raw.LatencyMs,
		ModelVersion: p.model,
	}

	if err != nil {
		return resp, fmt.Errorf("%s ocr benchmark: %w", p.name, err)
	}

	return resp, nil
}
