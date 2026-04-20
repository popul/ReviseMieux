// Package llm provides shared LLM prompt templates used across all providers.
//
// Prompts live in backend/prompts/ (single source of truth shared with the
// study-guide skill and benchmark scripts) and are embedded by the
// github.com/popul/revisemieux/prompts package via //go:embed.
// This file re-exports them with versioning + drift-detection helpers.
package llm

import (
	"crypto/sha256"
	"fmt"

	"github.com/popul/revisemieux/prompts"
)

// Prompt versions — increment when changing prompt content.
const (
	StructurationPromptVersion = "v1.1.0"
	OCRPromptVersion           = "v1.0.0"
	ScoringPromptVersion       = "v1.0.0"
	FidelityPromptVersion      = "v1.0.0"
)

// StructurationSystemPrompt is the system prompt for OCR → Items structuration.
var StructurationSystemPrompt = prompts.StructurationSystem

// OCRSystemPrompt is the system prompt for image → OCR blocks extraction.
var OCRSystemPrompt = prompts.OCRSystem

// ScoringSystemPrompt is the system prompt for student answer scoring.
var ScoringSystemPrompt = prompts.ScoringSystem

// E2ESystemPrompt is the system prompt for end-to-end vision → structured items.
var E2ESystemPrompt = prompts.E2ESystem

// StructurationPromptHash returns the SHA-256 hash of the structuration prompt.
func StructurationPromptHash() string {
	return shortHash(StructurationSystemPrompt)
}

// OCRPromptHash returns the SHA-256 hash of the OCR prompt.
func OCRPromptHash() string {
	return shortHash(OCRSystemPrompt)
}

// ScoringPromptHash returns the SHA-256 hash of the scoring prompt.
func ScoringPromptHash() string {
	return shortHash(ScoringSystemPrompt)
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
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

// BuildE2EUserPrompt builds the user message for E2E (images → items) extraction.
func BuildE2EUserPrompt(subject string) string {
	return fmt.Sprintf("Matière : %s\n\nAnalyse les photos de cahier ci-dessus et extrais les items de révision.", subject)
}

// HybridSystemPrompt is the system prompt for hybrid pipeline (OCR → images+blocks → items).
// Reuses the E2E prompt since the model receives both images and OCR text.
var HybridSystemPrompt = prompts.E2ESystem

// BuildHybridUserPrompt builds the user message for hybrid extraction (images + real OCR blocks).
func BuildHybridUserPrompt(subject string, blocksJSON string) string {
	return fmt.Sprintf("Matière : %s\n\nBlocs OCR pré-extraits :\n%s\n\nUtilise ces blocs ET les photos ci-dessus pour extraire les items de révision.", subject, blocksJSON)
}
