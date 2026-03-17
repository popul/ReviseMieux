// Package llm provides shared LLM prompt templates used across all providers.
//
// Prompts are stored as .txt files in the prompts/ directory and loaded via
// //go:embed. This ensures a single source of truth shared between production
// code (anthropic, openaicompat) and the benchmark generation script (Python).
package llm

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

// Prompt versions — increment when changing prompt content.
const (
	StructurationPromptVersion = "v1.1.0"
	OCRPromptVersion           = "v1.0.0"
	FidelityPromptVersion      = "v1.0.0"
)

// StructurationSystemPrompt is the system prompt for OCR → Items structuration.
//
//go:embed prompts/structuration_system.txt
var StructurationSystemPrompt string

// OCRSystemPrompt is the system prompt for image → OCR blocks extraction.
//
//go:embed prompts/ocr_system.txt
var OCRSystemPrompt string

// StructurationPromptHash returns the SHA-256 hash of the structuration prompt.
func StructurationPromptHash() string {
	h := sha256.Sum256([]byte(StructurationSystemPrompt))
	return fmt.Sprintf("sha256:%x", h[:8])
}

// OCRPromptHash returns the SHA-256 hash of the OCR prompt.
func OCRPromptHash() string {
	h := sha256.Sum256([]byte(OCRSystemPrompt))
	return fmt.Sprintf("sha256:%x", h[:8])
}

// BuildUserPrompt builds the user message for structuration.
func BuildUserPrompt(subject string, blocksJSON string) string {
	return fmt.Sprintf("Matière : %s\n\nBlocs OCR :\n%s", subject, blocksJSON)
}

// BuildOCRUserPrompt builds the user message for OCR extraction.
func BuildOCRUserPrompt(subject string) string {
	return fmt.Sprintf("Matière : %s\n\nExtrait tous les blocs de texte visibles sur ces photos de cahier.", subject)
}
