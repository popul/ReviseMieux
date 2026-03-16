// Package anthropic implements the LLM adapters using the Anthropic Claude API.
package anthropic

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Config holds configuration for Anthropic LLM clients.
type Config struct {
	APIKey             string
	StructurationModel string // e.g., "claude-sonnet-4-6"
	FidelityModel      string // e.g., "claude-haiku-4-5"
	QuestionGenModel   string // e.g., "claude-haiku-4-5"
}

// newClient creates an Anthropic API client with the given API key.
func newClient(apiKey string) anthropic.Client {
	return anthropic.NewClient(option.WithAPIKey(apiKey))
}
