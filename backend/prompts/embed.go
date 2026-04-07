// Package prompts is the single source of truth for LLM prompts shared
// across providers (anthropic, openaicompat) and the study-guide skill.
//
// Each prompt lives as a Markdown file in this directory tree and is embedded
// at compile time via //go:embed. The skill references the same files via
// ${CLAUDE_PROJECT_DIR}/backend/prompts/... so prototyping and production
// stay in sync.
//
// Lifecycle (see README.md): prototype (skill only) → validated → frozen
// (file in this tree, embedded by Go, hashed for drift detection).
package prompts

import _ "embed"

// OCRSystem is the system prompt for image → OCR blocks extraction.
//
//go:embed ocr/system.md
var OCRSystem string

// StructurationSystem is the system prompt for OCR → Items structuration.
//
//go:embed structuration/system.md
var StructurationSystem string

// ScoringSystem is the system prompt for student answer scoring.
//
//go:embed scoring/system.md
var ScoringSystem string
