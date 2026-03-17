package anthropic

import "github.com/popul/revisemieux/internal/infra/llm"

// Re-export from the shared llm package for backward compatibility.
// New code should import internal/infra/llm directly.

var (
	StructurationSystemPrompt = llm.StructurationSystemPrompt
	OCRSystemPrompt           = llm.OCRSystemPrompt
)

const (
	StructurationPromptVersion = llm.StructurationPromptVersion
	OCRPromptVersion           = llm.OCRPromptVersion
	FidelityPromptVersion      = llm.FidelityPromptVersion
)

var (
	StructurationPromptHash = llm.StructurationPromptHash
	OCRPromptHash           = llm.OCRPromptHash
	BuildUserPrompt         = llm.BuildUserPrompt
	BuildOCRUserPrompt      = llm.BuildOCRUserPrompt
)
