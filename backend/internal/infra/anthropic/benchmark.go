package anthropic

import (
	"context"
	"fmt"
	"time"

	sdkanthro "github.com/anthropics/anthropic-sdk-go"
	"github.com/popul/revisemieux/internal/benchmark"
)

// BenchmarkProvider implements benchmark.Provider using the Anthropic API.
type BenchmarkProvider struct {
	client   sdkanthro.Client
	name     string
	model    string
	priceIn  float64
	priceOut float64
}

// NewBenchmarkProvider creates a benchmark provider for a specific Anthropic model.
func NewBenchmarkProvider(apiKey, model string, priceIn, priceOut float64) *BenchmarkProvider {
	return &BenchmarkProvider{
		client:   newClient(apiKey),
		name:     "Anthropic",
		model:    model,
		priceIn:  priceIn,
		priceOut: priceOut,
	}
}

func (p *BenchmarkProvider) Name() string            { return p.name }
func (p *BenchmarkProvider) ModelID() string          { return p.model }
func (p *BenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *BenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// StructureBlocks sends the structuration prompt and returns the raw response.
func (p *BenchmarkProvider) StructureBlocks(ctx context.Context, systemPrompt, userPrompt string) (*benchmark.Response, error) {
	start := time.Now()

	msg, err := p.client.Messages.New(ctx, sdkanthro.MessageNewParams{
		Model:     sdkanthro.Model(p.model),
		MaxTokens: 4096,
		System: []sdkanthro.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []sdkanthro.MessageParam{
			sdkanthro.NewUserMessage(sdkanthro.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic benchmark: %w", err)
	}

	latency := time.Since(start).Milliseconds()

	text := extractText(msg)
	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  int(msg.Usage.InputTokens),
		TokensOutput: int(msg.Usage.OutputTokens),
		LatencyMs:    latency,
		ModelVersion: p.model,
	}, nil
}
