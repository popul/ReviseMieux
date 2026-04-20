// Package openaicompat provides a benchmark provider for OpenAI-compatible APIs.
//
// This covers OpenAI, Google Gemini (via OpenAI compat), Mistral, and DeepSeek,
// which all expose the same chat completions endpoint format.
package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/popul/revisemieux/internal/benchmark"
)

// BenchmarkProvider implements benchmark.Provider for any OpenAI-compatible API.
type BenchmarkProvider struct {
	httpClient     *http.Client
	baseURL        string
	apiKey         string
	name           string
	model          string
	priceIn        float64
	priceOut       float64
	reasoningModel   bool
	disableThinking  bool
}

// Config holds the configuration for creating an OpenAI-compatible benchmark provider.
type Config struct {
	BaseURL         string
	APIKey          string
	Name            string
	Model           string
	PriceIn         float64 // USD per 1M input tokens
	PriceOut        float64 // USD per 1M output tokens
	ReasoningModel  bool    // Use max_completion_tokens instead of max_tokens (o3, deepseek-reasoner)
	DisableThinking bool    // Send reasoning_effort=none to skip thinking (Qwen3.6 on LM Studio)
	PerPage         bool    // Process images one at a time and merge results (for small VLMs like RolmOCR)
}

// NewBenchmarkProvider creates a benchmark provider for an OpenAI-compatible API.
func NewBenchmarkProvider(cfg Config) *BenchmarkProvider {
	return &BenchmarkProvider{
		httpClient:      &http.Client{Timeout: 5 * time.Minute},
		baseURL:         cfg.BaseURL,
		apiKey:          cfg.APIKey,
		name:            cfg.Name,
		model:           cfg.Model,
		priceIn:         cfg.PriceIn,
		priceOut:        cfg.PriceOut,
		reasoningModel:  cfg.ReasoningModel,
		disableThinking: cfg.DisableThinking,
	}
}

func (p *BenchmarkProvider) Name() string             { return p.name }
func (p *BenchmarkProvider) ModelID() string          { return p.model }
func (p *BenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *BenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// chatRequest is the OpenAI chat completions request body.
type chatRequest struct {
	Model               string        `json:"model"`
	Messages            []chatMessage `json:"messages"`
	MaxTokens           int           `json:"max_tokens,omitempty"`
	MaxCompletionTokens int           `json:"max_completion_tokens,omitempty"`
	Temperature         *float64      `json:"temperature,omitempty"`
	ReasoningEffort     string        `json:"reasoning_effort,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the OpenAI chat completions response body.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content          string `json:"content"`
			Reasoning        string `json:"reasoning"`         // OpenRouter reasoning models
			ReasoningContent string `json:"reasoning_content"` // DeepSeek reasoner
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// StructureBlocks sends the structuration prompt and returns the raw response.
func (p *BenchmarkProvider) StructureBlocks(ctx context.Context, systemPrompt, userPrompt string) (*benchmark.Response, error) {
	reqBody := chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
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
		reqBody.MaxTokens = 4096
		reqBody.Temperature = &temp
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s benchmark marshal: %w", p.name, err)
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s benchmark request: %w", p.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s benchmark call: %w", p.name, err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s benchmark read body: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s benchmark HTTP %d: %s", p.name, resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("%s benchmark parse response: %w", p.name, err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%s benchmark: no choices in response", p.name)
	}

	text := chatResp.Choices[0].Message.Content
	// Reasoning models may put the answer in reasoning_content (DeepSeek) or reasoning (OpenRouter)
	if text == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		text = chatResp.Choices[0].Message.ReasoningContent
	}
	if text == "" && chatResp.Choices[0].Message.Reasoning != "" {
		text = chatResp.Choices[0].Message.Reasoning
	}
	if text == "" {
		return nil, fmt.Errorf("%s benchmark: empty content in response (tokens_out=%d)", p.name, chatResp.Usage.CompletionTokens)
	}

	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  chatResp.Usage.PromptTokens,
		TokensOutput: chatResp.Usage.CompletionTokens,
		LatencyMs:    latency,
		ModelVersion: chatResp.Model,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
