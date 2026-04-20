// Command benchmark runs the OCR and IDP benchmark suites against configured providers.
//
// Usage:
//
//	go run ./cmd/benchmark/ --type=idp --all
//	go run ./cmd/benchmark/ --type=ocr --all
//	go run ./cmd/benchmark/ --type=idp --provider=anthropic
//	go run ./cmd/benchmark/ --type=idp --models=gpt-4.1-mini,gemini-2.5-flash,qwen3.6-35b-a3b
//	go run ./cmd/benchmark/ --list-models
//	go run ./cmd/benchmark/ --type=idp --case=10_SVT_cours_louis
//	go run ./cmd/benchmark/ --type=idp --all --runs=3 --output=csv
//	go run ./cmd/benchmark/ --report
//	go run ./cmd/benchmark/ --report-run=2026-03-12_14h30
//
// Environment variables:
//
//	ANTHROPIC_API_KEY   — Anthropic Claude API key
//	OPENAI_API_KEY      — OpenAI GPT API key
//	GOOGLE_AI_API_KEY   — Google Gemini API key
//	MISTRAL_API_KEY     — Mistral API key
//	DEEPSEEK_API_KEY    — DeepSeek API key
//	OPENROUTER_API_KEY  — OpenRouter API key (for Qwen, Llama, MiniMax, GLM, InternVL, etc.)
//	LMSTUDIO_BASE_URL   — LM Studio local endpoint (default: http://localhost:1234/v1)
//	PADDLEOCR_BASE_URL  — llama-server endpoint for PaddleOCR-VL (default: http://localhost:1235/v1)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/popul/revisemieux/internal/benchmark"
	llmanthro "github.com/popul/revisemieux/internal/infra/anthropic"
	"github.com/popul/revisemieux/internal/infra/llm"
	llmmistral "github.com/popul/revisemieux/internal/infra/mistral"
	"github.com/popul/revisemieux/internal/infra/openaicompat"
	"gopkg.in/yaml.v3"
)

// modelDef describes a benchmarkable LLM model.
type modelDef struct {
	ID         string                                    // unique key used in --models flag
	Provider   string                                    // provider name for display
	EnvKey     string                                    // environment variable for API key
	IDPBuilder func(apiKey string) benchmark.Provider    // builder for IDP benchmark
	OCRBuilder func(apiKey string) benchmark.OCRProvider // builder for OCR benchmark (nil if not supported)
	E2EBuilder    func(apiKey string) benchmark.E2EProvider    // builder for E2E benchmark (nil if not supported)
	HybridBuilder func(apiKey string) benchmark.HybridProvider // builder for hybrid benchmark (nil if not supported)
}

// modelCatalog lists all available models, grouped by provider.
//
// Models removed on 2026-04-18 (based on benchmark run 2026-03-28):
//   - claude-sonnet-4-6:    $3/$15, worst OCR (0.52), IDP 0.73 — worst value of all models
//   - gpt-4o:               $2.5/$10, IDP 0.76, OCR 0.65 — superseded by gpt-4.1 ($2/$8)
//   - gpt-4o-mini:          $0.15/$0.60, IDP 0.77, OCR 0.67 — superseded by gpt-4.1-mini ($0.40/$1.60)
//   - o3-mini:              $1.10/$4.40, IDP 0.77, no vision — superseded by o4-mini (+ vision)
//   - deepseek-reasoner:    $0.55/$2.19, timeout 2/3 cas, no vision — instable
//   - gemini-2.5-pro:       $1.25/$10, IDP 0.79, OCR 0.74 — 10x plus cher que flash pour des résultats inférieurs
//   - mistral-large-latest: $2/$6, IDP 0.84, OCR 0.68 — cher, gain marginal vs mistral-small
//   - minimax-m2.5:         no vision — supersédé par minimax-m2.7
//   - nemotron-nano-vl:     free tier, résultats non fiables
//   - gemma3-27b:           $0.10/$0.10 — supersédé par gemma-4 (31B et 26B-A4B)
//   - internvl3-14b:        $0.03/$0.10 — redondant avec internvl3-78b, qualité insuffisante
//
// Modèles locaux LM Studio ajoutés (pricing = référence cloud pour le scoring composite) :
//   - qwen3.6-35b-a3b:     MoE 35B/3B actifs, ~$0.29/$1.65 (Alibaba Cloud ref)
//   - qwen3.5-35b-a3b:     MoE 35B/3B actifs, $0.16/$1.30 (OpenRouter ref)
//   - gemma4-26b-a4b:      MoE 26B/4B actifs, $0.07/$0.40 (OpenRouter ref) — remplace gemma3-27b
//   - paddleocr-vl-1.5:   testé puis retiré — 0.9B spécialisé docs imprimés, score OCR 0.40 sur
//                         cahiers manuscrits (vs 0.85 pour les VLM généralistes). Incompatible LM Studio,
//                         nécessite llama-server dédié. Pas adapté au use case.
//
// Nouveaux modèles cloud ajoutés :
//   - deepseek-v3.2:        $0.26/$0.38, remplace deepseek-chat ($0.27/$1.10), output 3x moins cher
//   - gemini-3.1-flash-lite-preview: $0.25/$1.50, le plus cheap multimodal chez Google, 1M context
//   - kimi-k2.5:            $0.38/$1.72 via OpenRouter, nativement multimodal (MoonshotAI)
//
// ERRATUM 2026-04-20 : "mistral-small-latest" résout vers Mistral Small 4 (119B MoE, juin 2026)
// et NON vers Mistral Small 3.1 (24B dense, mars 2025). Le GGUF local 3.1 ne peut pas atteindre
// la qualité cloud (0.58 vs 0.85). Renommé en "mistral-small-4" dans le catalogue pour clarté.
var modelCatalog = []modelDef{
	// Anthropic
	{
		ID: "claude-haiku-4-5", Provider: "anthropic", EnvKey: "ANTHROPIC_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return llmanthro.NewBenchmarkProvider(k, "claude-haiku-4-5", 1.00, 5.00)
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return llmanthro.NewOCRBenchmarkProvider(k, "claude-haiku-4-5", 1.00, 5.00)
		},
	},
	// OpenAI — GPT-4.1 family (replaces GPT-4o)
	{
		ID: "gpt-4.1", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1", PriceIn: 2.00, PriceOut: 8.00,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1", PriceIn: 2.00, PriceOut: 8.00,
			})
		},
	},
	{
		ID: "gpt-4.1-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1-mini", PriceIn: 0.40, PriceOut: 1.60,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1-mini", PriceIn: 0.40, PriceOut: 1.60,
			})
		},
	},
	{
		ID: "gpt-4.1-nano", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1-nano", PriceIn: 0.10, PriceOut: 0.40,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4.1-nano", PriceIn: 0.10, PriceOut: 0.40,
			})
		},
	},
	{
		ID: "o4-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "o4-mini", PriceIn: 1.10, PriceOut: 4.40,
				ReasoningModel: true,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "o4-mini", PriceIn: 1.10, PriceOut: 4.40,
			})
		},
	},
	// Google Gemini (OpenAI-compatible endpoint)
	{
		ID: "gemini-2.5-flash", Provider: "google", EnvKey: "GOOGLE_AI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-flash", PriceIn: 0.15, PriceOut: 0.60,
				ReasoningModel: true,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-flash", PriceIn: 0.15, PriceOut: 0.60,
			})
		},
	},
	// Google Gemini 3.1 Flash Lite (OpenAI-compatible endpoint)
	{
		ID: "gemini-3.1-flash-lite-preview", Provider: "google", EnvKey: "GOOGLE_AI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-3.1-flash-lite-preview", PriceIn: 0.25, PriceOut: 1.50,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-3.1-flash-lite-preview", PriceIn: 0.25, PriceOut: 1.50,
			})
		},
	},
	// Mistral
	// NOTE: "mistral-small-latest" résout vers Mistral Small 4 (119B MoE, 2603) — PAS le même modèle
	// que le GGUF local "Mistral-Small-3.1-24B" ! Voir incident bench 2026-04-20.
	{
		ID: "mistral-small-4", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-small-latest", PriceIn: 0.10, PriceOut: 0.30,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-small-latest", PriceIn: 0.10, PriceOut: 0.30,
			})
		},
	},
	{
		ID: "mistral-small-3.1-api", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-small-2503", PriceIn: 0.06, PriceOut: 0.18,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-small-2503", PriceIn: 0.06, PriceOut: 0.18,
			})
		},
	},
	{
		ID: "pixtral-12b", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "pixtral-12b-2409", PriceIn: 0.13, PriceOut: 0.13,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "pixtral-12b-2409", PriceIn: 0.13, PriceOut: 0.13,
			})
		},
	},
	// DeepSeek
	{
		ID: "deepseek-v3.2", Provider: "deepseek", EnvKey: "DEEPSEEK_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.deepseek.com", APIKey: k, Name: "DeepSeek",
				Model: "deepseek-chat", PriceIn: 0.26, PriceOut: 0.38,
			})
		},
		// deepseek-v3.2 does not support vision
	},
	// Local models via LM Studio — PriceIn/Out = 0 for fair composite scoring (no per-token cost).
	{
		ID: "mistral-small-3.1-bf16", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "bartowski/mistral-small-3.1-24b-instruct-2503", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "bartowski/mistral-small-3.1-24b-instruct-2503", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	{
		ID: "mistral-small-3.1-q8", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "mistral-small-3.1-24b-instruct-2503@q8_0", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "mistral-small-3.1-24b-instruct-2503@q8_0", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	{
		ID: "qwen3-vl-30b-a3b-local", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen3-vl-30b-a3b-instruct", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen3-vl-30b-a3b-instruct", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	{
		ID: "qwen3-vl-32b-local", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen3-vl-32b-instruct", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen3-vl-32b-instruct", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	{
		ID: "rolmocr-7b", Provider: "rolmocr", EnvKey: "ROLMOCR_BASE_URL",
		// OCR-only: RolmOCR is a Qwen2.5-VL-7B fine-tune specialized for OCR (92% handwriting accuracy).
		// Runs via llama-server. PerPage=true processes images one at a time (required for multi-page docs).
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "none", Name: "RolmOCR",
				Model: "rolmocr", PriceIn: 0, PriceOut: 0,
				PerPage: true,
			})
		},
	},
	{
		ID: "qwen3.6-35b-a3b", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen/qwen3.6-35b-a3b", PriceIn: 0, PriceOut: 0,
				DisableThinking: true,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen/qwen3.6-35b-a3b", PriceIn: 0, PriceOut: 0,
				DisableThinking: true,
			})
		},
	},
	{
		ID: "qwen3.5-35b-a3b", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen/qwen3.5-35b-a3b", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "qwen/qwen3.5-35b-a3b", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	{
		ID: "gemma4-26b-a4b", Provider: "lmstudio", EnvKey: "LMSTUDIO_BASE_URL",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "unsloth/gemma-4-26b-a4b-it", PriceIn: 0, PriceOut: 0,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: k, APIKey: "lm-studio", Name: "LMStudio",
				Model: "unsloth/gemma-4-26b-a4b-it", PriceIn: 0, PriceOut: 0,
			})
		},
	},
	// Qwen (via OpenRouter)
	{
		ID: "qwen3.5-397b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3.5-397b-a17b", PriceIn: 0.30, PriceOut: 0.30,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3.5-397b-a17b", PriceIn: 0.30, PriceOut: 0.30,
			})
		},
	},
	{
		ID: "qwen3.5-9b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3.5-9b", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3.5-9b", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
	},
	// Meta Llama 4 (via OpenRouter)
	{
		ID: "llama4-maverick", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "meta-llama/llama-4-maverick", PriceIn: 0.20, PriceOut: 0.20,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "meta-llama/llama-4-maverick", PriceIn: 0.20, PriceOut: 0.20,
			})
		},
	},
	{
		ID: "llama4-scout", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "meta-llama/llama-4-scout", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "meta-llama/llama-4-scout", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
	},
	// StepFun (via OpenRouter)
	{
		ID: "step3.5-flash", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "stepfun/step-3.5-flash", PriceIn: 0.10, PriceOut: 0.30,
				ReasoningModel: true,
			})
		},
		// no vision support
	},
	// Qwen3-VL (via OpenRouter)
	{
		ID: "qwen3-vl-30b-a3b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-30b-a3b-instruct", PriceIn: 0.13, PriceOut: 0.52,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-30b-a3b-instruct", PriceIn: 0.13, PriceOut: 0.52,
			})
		},
	},
	{
		ID: "qwen3-vl-235b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-235b-a22b-instruct", PriceIn: 0.20, PriceOut: 0.88,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-235b-a22b-instruct", PriceIn: 0.20, PriceOut: 0.88,
			})
		},
	},
	{
		ID: "qwen3-vl-32b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-32b-instruct", PriceIn: 0.10, PriceOut: 0.42,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "qwen/qwen3-vl-32b-instruct", PriceIn: 0.10, PriceOut: 0.42,
			})
		},
	},
	// Moonshot Kimi K2.5 (via OpenRouter) — natively multimodal, vision+text+video
	{
		ID: "kimi-k2.5", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "moonshotai/kimi-k2.5", PriceIn: 0.38, PriceOut: 1.72,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "moonshotai/kimi-k2.5", PriceIn: 0.38, PriceOut: 1.72,
			})
		},
	},
	// MiniMax M2.7 (via OpenRouter)
	{
		ID: "minimax-m2.7", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "minimax/minimax-m2.7", PriceIn: 0.30, PriceOut: 1.20,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "minimax/minimax-m2.7", PriceIn: 0.30, PriceOut: 1.20,
			})
		},
	},
	// Z.AI GLM-4.5V (via OpenRouter)
	{
		ID: "glm-4.5v", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "z-ai/glm-4.5v", PriceIn: 0.60, PriceOut: 1.80,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "z-ai/glm-4.5v", PriceIn: 0.60, PriceOut: 1.80,
			})
		},
	},
	// InternVL3 78B (via OpenRouter)
	{
		ID: "internvl3-78b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "opengvlab/internvl3-78b", PriceIn: 0.07, PriceOut: 0.26,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "opengvlab/internvl3-78b", PriceIn: 0.07, PriceOut: 0.26,
			})
		},
	},
	// NOTE: paddleocr-vl-1.5 retiré (bench 2026-04-18, score OCR 0.40) — modèle 0.9B spécialisé
	// documents imprimés (PDF, tableaux, formulaires), très mauvais sur photos de cahiers manuscrits
	// (détection 0.62, texte 0.37). Les VLM généralistes (qwen3.6-35b, gemini-2.5-flash) sont 2x meilleurs.
	// Mistral OCR 3 (dedicated OCR API — not chat completions)
	{
		ID: "mistral-ocr-3", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		// Mistral OCR is OCR-only (dedicated document processing endpoint)
		OCRBuilder: func(k string) benchmark.OCRProvider {
			// Pricing: ~$2 per 1000 pages. Token-based approximation: very cheap.
			return llmmistral.NewOCRBenchmarkProvider(k, "mistral-ocr-latest", 0.05, 0.05)
		},
	},
}

func init() {
	// Auto-generate E2EBuilder for every model that has an OCRBuilder.
	// E2E reuses the same config (baseURL, apiKey, model, pricing) but with the E2E provider.
	for i := range modelCatalog {
		// Auto-generate HybridBuilder for vision models (reuses E2E provider for structuration)
		if modelCatalog[i].OCRBuilder != nil && modelCatalog[i].HybridBuilder == nil {
			def := modelCatalog[i]
			testProv := openaicompat.NewE2EFromOCR(def.OCRBuilder("test"))
			if testProv != nil {
				modelCatalog[i].HybridBuilder = func(k string) benchmark.HybridProvider {
					return openaicompat.NewE2EFromOCR(def.OCRBuilder(k))
				}
			}
		}
		// Auto-generate HybridBuilder for text-only models (IDP only, no vision).
		// These only work with --no-images flag (classic pipeline mode).
		if modelCatalog[i].IDPBuilder != nil && modelCatalog[i].OCRBuilder == nil && modelCatalog[i].HybridBuilder == nil {
			def := modelCatalog[i]
			modelCatalog[i].HybridBuilder = func(k string) benchmark.HybridProvider {
				return openaicompat.NewE2EFromIDP(def.IDPBuilder(k))
			}
		}
		// Auto-generate E2EBuilder for vision models
		if modelCatalog[i].OCRBuilder != nil && modelCatalog[i].E2EBuilder == nil {
			def := modelCatalog[i] // capture
			// Test if the OCR provider can be wrapped into an E2E provider.
			// Non-openaicompat providers (Anthropic, Mistral OCR) return nil.
			testProv := openaicompat.NewE2EFromOCR(def.OCRBuilder("test"))
			if testProv != nil {
				modelCatalog[i].E2EBuilder = func(k string) benchmark.E2EProvider {
					return openaicompat.NewE2EFromOCR(def.OCRBuilder(k))
				}
			}
		}
	}
}

func main() {
	benchType := flag.String("type", "idp", "Benchmark type: ocr, idp, e2e, hybrid, pipeline")
	all := flag.Bool("all", false, "Run all providers with available API keys")
	provider := flag.String("provider", "", "Run all models for a provider (anthropic, openai, google, mistral, deepseek)")
	models := flag.String("models", "", "Comma-separated list of model IDs to run (e.g. gpt-4.1-mini,gemini-2.5-flash,qwen3.6-35b-a3b)")
	listModels := flag.Bool("list-models", false, "List all available models and exit")
	caseID := flag.String("case", "", "Run specific test case")
	runs := flag.Int("runs", 1, "Number of runs per case (for variance measurement)")
	output := flag.String("output", "console", "Output format: console, json, csv")
	report := flag.Bool("report", false, "Generate HTML report from latest results (no benchmark run)")
	reportRun := flag.String("report-run", "", "Generate report from a specific run directory")
	reportOutput := flag.String("report-output", "", "Output path for the HTML report")
	parallel := flag.Int("parallel", 8, "Max number of providers to run in parallel")
	appendTo := flag.String("append-to", "", "Append results to an existing run directory (e.g. 2026-03-27_14h30)")
	ocrModel := flag.String("ocr-model", "", "OCR model for hybrid mode (e.g. rolmocr-7b). The --models flag specifies the structuration model.")
	noImages := flag.Bool("no-images", false, "Hybrid mode: don't send images to the structurer (classic OCR→IDP pipeline)")
	flag.Parse()

	// List models mode
	if *listModels {
		printModelCatalog()
		os.Exit(0)
	}

	// Report-only mode
	if *report || *reportRun != "" {
		var resultsDir string
		var err error
		switch *benchType {
		case "ocr":
			resultsDir = filepath.Join(testdataDir(), "benchmark", "results", "ocr")
			err = generateOCRReport(resultsDir, *reportRun, *reportOutput)
		case "e2e":
			resultsDir = filepath.Join(testdataDir(), "benchmark", "results", "e2e")
			err = generateReport(resultsDir, *reportRun, *reportOutput)
		case "hybrid", "pipeline":
			resultsDir = filepath.Join(testdataDir(), "benchmark", "results", *benchType)
			err = generateReport(resultsDir, *reportRun, *reportOutput)
		default:
			resultsDir = filepath.Join(testdataDir(), "benchmark", "results")
			err = generateReport(resultsDir, *reportRun, *reportOutput)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !*all && *provider == "" && *models == "" {
		printUsage()
		os.Exit(1)
	}

	// Load test cases
	casesDir := filepath.Join(testdataDir(), "benchmark", "cases")
	cases, err := loadTestCases(casesDir, *caseID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading test cases: %v\n", err)
		os.Exit(1)
	}

	switch *benchType {
	case "ocr":
		runOCRBenchmark(cases, *all, *provider, *models, *runs, *output, *parallel, *appendTo)
	case "idp":
		runIDPBenchmark(cases, *all, *provider, *models, *runs, *output, *parallel, *appendTo)
	case "e2e":
		runE2EBenchmark(cases, *all, *provider, *models, *runs, *output, *parallel, *appendTo)
	case "hybrid":
		if *ocrModel == "" {
			fmt.Fprintln(os.Stderr, "Hybrid mode requires --ocr-model (e.g. --ocr-model=rolmocr-7b)")
			os.Exit(1)
		}
		runHybridBenchmark(cases, *all, *provider, *models, *ocrModel, *noImages, *runs, *output, *parallel, *appendTo)
	default:
		fmt.Fprintf(os.Stderr, "Unknown benchmark type: %s (use 'ocr' or 'idp')\n", *benchType)
		os.Exit(1)
	}
}

// --- IDP Benchmark (structuration LLM) ---

func runIDPBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter string, runs int, output string, parallel int, appendTo string) {
	fmt.Printf("[IDP] Loaded %d test case(s)\n", len(cases))

	providers := buildIDPProviders(all, single, modelFilter)
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No IDP providers configured (check API keys)")
		os.Exit(1)
	}

	// Load existing results and skip already-benchmarked models.
	var existingSummaries []benchmark.RunSummary
	if appendTo != "" {
		resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "idp")
		summaryPath := filepath.Join(resultsDir, appendTo, "summary.json")
		loaded, err := loadSummaries(summaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading existing results from %s: %v\n", summaryPath, err)
			os.Exit(1)
		}
		existingSummaries = loaded
		existing := make(map[string]bool)
		for _, s := range loaded {
			existing[s.Model] = true
		}
		var filtered []benchmark.Provider
		for _, p := range providers {
			if existing[p.ModelID()] {
				fmt.Printf("[IDP] Skipping %s (already in %s)\n", p.ModelID(), appendTo)
			} else {
				filtered = append(filtered, p)
			}
		}
		providers = filtered
		if len(providers) == 0 {
			fmt.Println("[IDP] All requested models already present, nothing to run.")
			return
		}
	}

	fmt.Printf("[IDP] Running %d provider(s): %s\n", len(providers), idpProviderNames(providers))
	fmt.Printf("[IDP] Runs per case: %d\n\n", runs)

	// Run providers in parallel (cases within a provider stay sequential to avoid rate limits).
	type providerResults struct {
		results []benchmark.EvalResult
	}
	perProvider := make([]providerResults, len(providers))
	var wg sync.WaitGroup
	var mu sync.Mutex // protects stdout
	sem := make(chan struct{}, parallel)

	for pi, p := range providers {
		wg.Add(1)
		go func(pi int, p benchmark.Provider) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var results []benchmark.EvalResult
			mu.Lock()
			fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
			mu.Unlock()
			for _, tc := range cases {
				for run := 0; run < runs; run++ {
					result := runSingleIDPCase(p, tc)
					mu.Lock()
					if runs > 1 {
						fmt.Printf("  [%s/%s] run %d/%d: Q=%.2f cost=$%.5f latency=%dms\n",
							p.ModelID(), tc.ID, run+1, runs, result.QualityScore, result.CostUSD, result.LatencyMs)
					} else {
						fmt.Printf("  [%s/%s] Q=%.2f cost=$%.5f latency=%dms items=%d\n",
							p.ModelID(), tc.ID, result.QualityScore, result.CostUSD, result.LatencyMs, result.ItemsFound)
					}
					mu.Unlock()
					results = append(results, result)
				}
			}
			perProvider[pi] = providerResults{results: results}
		}(pi, p)
	}
	wg.Wait()

	var allResults []benchmark.EvalResult
	for _, pr := range perProvider {
		allResults = append(allResults, pr.results...)
	}

	benchmark.ComputeCompositeScores(allResults)

	summaryMap := make(map[string][]benchmark.EvalResult)
	for _, r := range allResults {
		key := r.Provider + "|" + r.Model
		summaryMap[key] = append(summaryMap[key], r)
	}

	var newSummaries []benchmark.RunSummary
	for key, results := range summaryMap {
		parts := strings.SplitN(key, "|", 2)
		s := benchmark.Summarize(parts[0], parts[1], results)
		newSummaries = append(newSummaries, s)
	}

	// Merge with existing summaries when appending.
	summaries := append(existingSummaries, newSummaries...)

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	fmt.Println()
	switch output {
	case "json":
		outputIDPJSON(summaries)
	case "csv":
		outputIDPCSV(summaries)
	default:
		outputIDPConsole(summaries)
	}

	if appendTo != "" {
		saveIDPResultsTo(summaries, appendTo)
	} else {
		saveIDPResults(summaries)
	}
}

func runSingleIDPCase(p benchmark.Provider, tc benchmark.TestCase) benchmark.EvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	blocksJSON, _ := json.Marshal(tc.Blocks)
	systemPrompt, userPrompt := structurationPrompts(tc.Subject, string(blocksJSON))

	resp, err := p.StructureBlocks(ctx, systemPrompt, userPrompt)
	if err != nil {
		return benchmark.EvalResult{
			CaseID:    tc.ID,
			Provider:  p.Name(),
			Model:     p.ModelID(),
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	var parsed benchmark.ParsedOutput
	if err := json.Unmarshal(benchmark.StripMarkdownFences(resp.RawJSON), &parsed); err != nil {
		result := benchmark.Evaluate(tc, nil, resp, p)
		result.Timestamp = time.Now()
		result.Error = fmt.Sprintf("JSON parse error: %v", err)
		return result
	}

	result := benchmark.Evaluate(tc, &parsed, resp, p)
	result.Timestamp = time.Now()
	return result
}

// --- OCR Benchmark (vision/OCR extraction) ---

func runOCRBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter string, runs int, output string, parallel int, appendTo string) {
	// Filter to cases with images only
	var ocrCases []benchmark.TestCase
	for _, tc := range cases {
		if tc.HasImages && len(tc.ImagePaths) > 0 {
			ocrCases = append(ocrCases, tc)
		}
	}
	if len(ocrCases) == 0 {
		fmt.Fprintln(os.Stderr, "No test cases with images found for OCR benchmark")
		os.Exit(1)
	}
	fmt.Printf("[OCR] Loaded %d test case(s) with images\n", len(ocrCases))

	providers := buildOCRProviders(all, single, modelFilter)
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No OCR providers configured (check API keys and vision support)")
		os.Exit(1)
	}

	// Load existing results and skip already-benchmarked models.
	var existingSummaries []benchmark.OCRRunSummary
	if appendTo != "" {
		resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "ocr")
		summaryPath := filepath.Join(resultsDir, appendTo, "summary.json")
		loaded, err := loadOCRSummaries(summaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading existing results from %s: %v\n", summaryPath, err)
			os.Exit(1)
		}
		existingSummaries = loaded
		existing := make(map[string]bool)
		for _, s := range loaded {
			existing[s.Model] = true
		}
		var filtered []benchmark.OCRProvider
		for _, p := range providers {
			if existing[p.ModelID()] {
				fmt.Printf("[OCR] Skipping %s (already in %s)\n", p.ModelID(), appendTo)
			} else {
				filtered = append(filtered, p)
			}
		}
		providers = filtered
		if len(providers) == 0 {
			fmt.Println("[OCR] All requested models already present, nothing to run.")
			return
		}
	}

	fmt.Printf("[OCR] Running %d provider(s): %s\n", len(providers), ocrProviderNames(providers))
	fmt.Printf("[OCR] Runs per case: %d\n\n", runs)

	// Run providers in parallel (cases within a provider stay sequential to avoid rate limits).
	type ocrProviderResults struct {
		results []benchmark.OCREvalResult
	}
	perProvider := make([]ocrProviderResults, len(providers))
	var wg sync.WaitGroup
	var mu sync.Mutex // protects stdout
	sem := make(chan struct{}, parallel)

	for pi, p := range providers {
		wg.Add(1)
		go func(pi int, p benchmark.OCRProvider) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var results []benchmark.OCREvalResult
			mu.Lock()
			fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
			mu.Unlock()
			for _, tc := range ocrCases {
				for run := 0; run < runs; run++ {
					result := runSingleOCRCase(p, tc)
					mu.Lock()
					if runs > 1 {
						fmt.Printf("  [%s/%s] run %d/%d: Q=%.2f cost=$%.5f latency=%dms blocks=%d/%d\n",
							p.ModelID(), tc.ID, run+1, runs, result.QualityScore, result.CostUSD, result.LatencyMs,
							result.BlocksFound, result.BlocksExpected)
					} else {
						fmt.Printf("  [%s/%s] Q=%.2f text=%.2f detect=%.2f cost=$%.5f latency=%dms blocks=%d/%d\n",
							p.ModelID(), tc.ID, result.QualityScore, result.TextAccuracy, result.DetectionScore,
							result.CostUSD, result.LatencyMs, result.BlocksFound, result.BlocksExpected)
					}
					mu.Unlock()
					results = append(results, result)
				}
			}
			perProvider[pi] = ocrProviderResults{results: results}
		}(pi, p)
	}
	wg.Wait()

	var allResults []benchmark.OCREvalResult
	for _, pr := range perProvider {
		allResults = append(allResults, pr.results...)
	}

	benchmark.ComputeOCRCompositeScores(allResults)

	summaryMap := make(map[string][]benchmark.OCREvalResult)
	for _, r := range allResults {
		key := r.Provider + "|" + r.Model
		summaryMap[key] = append(summaryMap[key], r)
	}

	var newSummaries []benchmark.OCRRunSummary
	for key, results := range summaryMap {
		parts := strings.SplitN(key, "|", 2)
		s := benchmark.SummarizeOCR(parts[0], parts[1], results)
		newSummaries = append(newSummaries, s)
	}

	// Merge with existing summaries when appending.
	summaries := append(existingSummaries, newSummaries...)

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	fmt.Println()
	switch output {
	case "json":
		outputOCRJSON(summaries)
	case "csv":
		outputOCRCSV(summaries)
	default:
		outputOCRConsole(summaries)
	}

	if appendTo != "" {
		saveOCRResultsTo(summaries, appendTo)
	} else {
		saveOCRResults(summaries)
	}
}

func runSingleOCRCase(p benchmark.OCRProvider, tc benchmark.TestCase) benchmark.OCREvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := p.ExtractBlocks(ctx, tc.ImagePaths, tc.Subject)
	if err != nil {
		return benchmark.OCREvalResult{
			CaseID:    tc.ID,
			Provider:  p.Name(),
			Model:     p.ModelID(),
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	blocks, err := benchmark.ParseOCROutput(resp.RawJSON)
	if err != nil {
		result := benchmark.EvaluateOCR(tc, nil, resp, p)
		result.Timestamp = time.Now()
		result.Error = fmt.Sprintf("JSON parse error: %v", err)
		return result
	}

	result := benchmark.EvaluateOCR(tc, blocks, resp, p)
	result.Timestamp = time.Now()
	return result
}

// --- Provider builders ---

// --- E2E Benchmark (images → structured items, single VLM pass) ---

func runE2EBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter string, runs int, output string, parallel int, appendTo string) {
	// Filter to cases with images only
	var e2eCases []benchmark.TestCase
	for _, tc := range cases {
		if tc.HasImages && len(tc.ImagePaths) > 0 {
			e2eCases = append(e2eCases, tc)
		}
	}
	if len(e2eCases) == 0 {
		fmt.Fprintln(os.Stderr, "No test cases with images found for E2E benchmark")
		os.Exit(1)
	}
	fmt.Printf("[E2E] Loaded %d test case(s) with images\n", len(e2eCases))

	providers := buildE2EProviders(all, single, modelFilter)
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No E2E providers configured (check API keys and vision support)")
		os.Exit(1)
	}

	// Load existing results and skip already-benchmarked models.
	var existingSummaries []benchmark.RunSummary
	if appendTo != "" {
		resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "e2e")
		summaryPath := filepath.Join(resultsDir, appendTo, "summary.json")
		loaded, err := loadSummaries(summaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading existing results from %s: %v\n", summaryPath, err)
			os.Exit(1)
		}
		existingSummaries = loaded
		existing := make(map[string]bool)
		for _, s := range loaded {
			existing[s.Model] = true
		}
		var filtered []benchmark.E2EProvider
		for _, p := range providers {
			if existing[p.ModelID()] {
				fmt.Printf("[E2E] Skipping %s (already in %s)\n", p.ModelID(), appendTo)
			} else {
				filtered = append(filtered, p)
			}
		}
		providers = filtered
		if len(providers) == 0 {
			fmt.Println("[E2E] All requested models already present, nothing to run.")
			return
		}
	}

	fmt.Printf("[E2E] Running %d provider(s): %s\n", len(providers), e2eProviderNames(providers))
	fmt.Printf("[E2E] Runs per case: %d\n\n", runs)

	type providerResults struct {
		results []benchmark.EvalResult
	}
	perProvider := make([]providerResults, len(providers))
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, parallel)

	for pi, p := range providers {
		wg.Add(1)
		go func(pi int, p benchmark.E2EProvider) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var results []benchmark.EvalResult
			mu.Lock()
			fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
			mu.Unlock()
			for _, tc := range e2eCases {
				for run := 0; run < runs; run++ {
					result := runSingleE2ECase(p, tc)
					mu.Lock()
					fmt.Printf("  [%s/%s] Q=%.2f cost=$%.5f latency=%dms items=%d\n",
						p.ModelID(), tc.ID, result.QualityScore, result.CostUSD, result.LatencyMs, result.ItemsFound)
					mu.Unlock()
					results = append(results, result)
				}
			}
			perProvider[pi] = providerResults{results: results}
		}(pi, p)
	}
	wg.Wait()

	var allResults []benchmark.EvalResult
	for _, pr := range perProvider {
		allResults = append(allResults, pr.results...)
	}

	benchmark.ComputeCompositeScores(allResults)

	summaryMap := make(map[string][]benchmark.EvalResult)
	for _, r := range allResults {
		key := r.Provider + "|" + r.Model
		summaryMap[key] = append(summaryMap[key], r)
	}

	var newSummaries []benchmark.RunSummary
	for key, results := range summaryMap {
		parts := strings.SplitN(key, "|", 2)
		s := benchmark.Summarize(parts[0], parts[1], results)
		newSummaries = append(newSummaries, s)
	}

	summaries := append(existingSummaries, newSummaries...)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	fmt.Println()
	switch output {
	case "json":
		outputIDPJSON(summaries)
	case "csv":
		outputIDPCSV(summaries)
	default:
		outputIDPConsole(summaries)
	}

	if appendTo != "" {
		saveE2EResultsTo(summaries, appendTo)
	} else {
		saveE2EResults(summaries)
	}
}

func runSingleE2ECase(p benchmark.E2EProvider, tc benchmark.TestCase) benchmark.EvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := p.StructureImages(ctx, tc.ImagePaths, tc.Subject)
	if err != nil {
		return benchmark.EvalResult{
			CaseID:    tc.ID,
			Provider:  p.Name(),
			Model:     p.ModelID(),
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	var parsed benchmark.ParsedOutput
	if err := json.Unmarshal(benchmark.StripMarkdownFences(resp.RawJSON), &parsed); err != nil {
		result := benchmark.Evaluate(tc, nil, resp, p)
		result.Timestamp = time.Now()
		result.Error = fmt.Sprintf("JSON parse error: %v", err)
		return result
	}

	result := benchmark.Evaluate(tc, &parsed, resp, p)
	result.Timestamp = time.Now()
	return result
}

func buildE2EProviders(all bool, single, modelFilter string) []benchmark.E2EProvider {
	selectedModels := parseModelFilter(modelFilter)
	var providers []benchmark.E2EProvider
	for _, def := range modelCatalog {
		if def.E2EBuilder == nil {
			if shouldInclude(def, all, single, selectedModels) {
				fmt.Printf("  [skip] %s — no E2E support\n", def.ID)
			}
			continue
		}
		if !shouldInclude(def, all, single, selectedModels) {
			continue
		}
		key := os.Getenv(def.EnvKey)
		if key == "" {
			fmt.Printf("  [skip] %s — %s not set\n", def.ID, def.EnvKey)
			continue
		}
		providers = append(providers, def.E2EBuilder(key))
	}
	return providers
}

func e2eProviderNames(providers []benchmark.E2EProvider) string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name() + "/" + p.ModelID()
	}
	return strings.Join(names, ", ")
}

func saveE2EResults(summaries []benchmark.RunSummary) {
	ts := time.Now().Format("2006-01-02_15h04")
	saveE2EResultsTo(summaries, ts)
}

func saveE2EResultsTo(summaries []benchmark.RunSummary, runID string) {
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "e2e")
	dir := filepath.Join(resultsDir, runID)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[E2E] Results saved to %s\n", path)

	reportPath := filepath.Join(dir, "report.html")
	if err := generateReport(resultsDir, runID, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate HTML report: %v\n", err)
	}
}

// --- Hybrid Benchmark (real OCR → structuration with images + OCR blocks) ---

func runHybridBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter, ocrModelID string, noImages bool, runs int, output string, parallel int, appendTo string) {
	// Filter to cases with images
	var pipeCases []benchmark.TestCase
	for _, tc := range cases {
		if tc.HasImages && len(tc.ImagePaths) > 0 {
			pipeCases = append(pipeCases, tc)
		}
	}
	if len(pipeCases) == 0 {
		fmt.Fprintln(os.Stderr, "No test cases with images found for pipeline benchmark")
		os.Exit(1)
	}

	// Build the OCR provider
	var ocrProvider benchmark.OCRProvider
	for _, def := range modelCatalog {
		if def.ID == ocrModelID && def.OCRBuilder != nil {
			key := os.Getenv(def.EnvKey)
			if key == "" {
				fmt.Fprintf(os.Stderr, "OCR model %s requires %s to be set\n", ocrModelID, def.EnvKey)
				os.Exit(1)
			}
			ocrProvider = def.OCRBuilder(key)
			break
		}
	}
	if ocrProvider == nil {
		fmt.Fprintf(os.Stderr, "OCR model %s not found or has no OCR support\n", ocrModelID)
		os.Exit(1)
	}

	// Build the hybrid structuration providers
	hybridProviders := buildHybridProviders(all, single, modelFilter)
	if len(hybridProviders) == 0 {
		fmt.Fprintln(os.Stderr, "No hybrid providers configured for structuration")
		os.Exit(1)
	}

	fmt.Printf("[Hybrid] OCR: %s/%s → Structuration: %d model(s)\n", ocrProvider.Name(), ocrProvider.ModelID(), len(hybridProviders))
	fmt.Printf("[Hybrid] Loaded %d test case(s)\n", len(pipeCases))

	// Step 1: Run OCR on all cases (sequentially, once)
	fmt.Printf("\n[Hybrid] Step 1: Running OCR (%s)...\n", ocrProvider.ModelID())
	type ocrResult struct {
		blocksJSON string
		latencyMs  int64
		costUSD    float64
		tokensIn   int
		tokensOut  int
		err        error
	}
	ocrResults := make(map[string]ocrResult) // keyed by case ID
	for _, tc := range pipeCases {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		resp, err := ocrProvider.ExtractBlocks(ctx, tc.ImagePaths, tc.Subject)
		cancel()
		if err != nil {
			fmt.Printf("  [OCR/%s] ERROR: %v\n", tc.ID, err)
			ocrResults[tc.ID] = ocrResult{err: err}
			continue
		}
		blocks, parseErr := benchmark.ParseOCROutput(resp.RawJSON)
		if parseErr != nil {
			fmt.Printf("  [OCR/%s] PARSE ERROR: %v\n", tc.ID, parseErr)
			ocrResults[tc.ID] = ocrResult{err: parseErr, latencyMs: resp.LatencyMs}
			continue
		}
		blocksJSON, _ := json.Marshal(blocks)
		fmt.Printf("  [OCR/%s] %d blocks, %dms\n", tc.ID, len(blocks), resp.LatencyMs)
		costUSD := float64(resp.TokensInput)*ocrProvider.PricePerMInput()/1_000_000 +
			float64(resp.TokensOutput)*ocrProvider.PricePerMOutput()/1_000_000
		ocrResults[tc.ID] = ocrResult{
			blocksJSON: string(blocksJSON),
			latencyMs:  resp.LatencyMs,
			costUSD:    costUSD,
			tokensIn:   resp.TokensInput,
			tokensOut:  resp.TokensOutput,
		}
	}

	// Step 2: Run hybrid structuration with real OCR output
	fmt.Printf("\n[Hybrid] Step 2: Running hybrid structuration...\n")
	type providerResults struct {
		results []benchmark.EvalResult
	}
	perProvider := make([]providerResults, len(hybridProviders))
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, parallel)

	for pi, p := range hybridProviders {
		wg.Add(1)
		go func(pi int, p benchmark.HybridProvider) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var results []benchmark.EvalResult
			mu.Lock()
			fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
			mu.Unlock()
			for _, tc := range pipeCases {
				ocr := ocrResults[tc.ID]
				if ocr.err != nil {
					results = append(results, benchmark.EvalResult{
						CaseID:    tc.ID,
						Provider:  p.Name(),
						Model:     p.ModelID(),
						Timestamp: time.Now(),
						Error:     fmt.Sprintf("OCR failed: %v", ocr.err),
					})
					continue
				}
				for run := 0; run < runs; run++ {
					result := runSingleHybridCase(p, tc, ocr.blocksJSON, noImages, ocr.latencyMs, ocr.costUSD, ocr.tokensIn)
					mu.Lock()
					fmt.Printf("  [%s/%s] Q=%.2f cost=$%.5f latency=%dms items=%d\n",
						p.ModelID(), tc.ID, result.QualityScore, result.CostUSD, result.LatencyMs, result.ItemsFound)
					mu.Unlock()
					results = append(results, result)
					// Brief pause between cases to avoid rate limits
					time.Sleep(2 * time.Second)
				}
			}
			perProvider[pi] = providerResults{results: results}
		}(pi, p)
	}
	wg.Wait()

	var allResults []benchmark.EvalResult
	for _, pr := range perProvider {
		allResults = append(allResults, pr.results...)
	}

	benchmark.ComputeCompositeScores(allResults)

	summaryMap := make(map[string][]benchmark.EvalResult)
	for _, r := range allResults {
		key := r.Provider + "|" + r.Model
		summaryMap[key] = append(summaryMap[key], r)
	}

	var summaries []benchmark.RunSummary
	for key, results := range summaryMap {
		parts := strings.SplitN(key, "|", 2)
		s := benchmark.Summarize(parts[0], parts[1], results)
		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	fmt.Println()
	outputIDPConsole(summaries)

	ts := time.Now().Format("2006-01-02_15h04")
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "hybrid")
	dir := filepath.Join(resultsDir, ts)
	os.MkdirAll(dir, 0o755)
	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[Hybrid] Results saved to %s\n", path)
	fmt.Printf("[Hybrid] OCR model: %s\n", ocrProvider.ModelID())
	if err := generateReport(resultsDir, ts, filepath.Join(dir, "report.html")); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate HTML report: %v\n", err)
	}
}

func runSingleHybridCase(p benchmark.HybridProvider, tc benchmark.TestCase, blocksJSON string, noImages bool, ocrLatency int64, ocrCost float64, ocrTokensIn int) benchmark.EvalResult {
	// Retry with backoff on transient errors (429 rate limit, timeouts)
	var resp *benchmark.Response
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*60) * time.Second
			fmt.Printf("    [retry %d/%d after %s] %s/%s\n", attempt+1, 3, backoff, p.ModelID(), tc.ID)
			time.Sleep(backoff)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		imagePaths := tc.ImagePaths
		if noImages {
			imagePaths = nil // Classic pipeline: no images sent to structurer
		}
		resp, err = p.StructureHybrid(ctx, imagePaths, blocksJSON, tc.Subject)
		cancel()
		if err == nil {
			break
		}
		// Retry on rate limit (429) or timeout
		errStr := err.Error()
		if !strings.Contains(errStr, "429") && !strings.Contains(errStr, "deadline exceeded") {
			break // non-retryable error
		}
	}
	if err != nil {
		return benchmark.EvalResult{
			CaseID:    tc.ID,
			Provider:  p.Name(),
			Model:     p.ModelID(),
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	var parsed benchmark.ParsedOutput
	if err := json.Unmarshal(benchmark.StripMarkdownFences(resp.RawJSON), &parsed); err != nil {
		result := benchmark.Evaluate(tc, nil, resp, p)
		result.Timestamp = time.Now()
		result.Error = fmt.Sprintf("JSON parse error: %v", err)
		// Add OCR cost/latency to totals
		result.LatencyMs += ocrLatency
		result.CostUSD += ocrCost
		return result
	}

	result := benchmark.Evaluate(tc, &parsed, resp, p)
	result.Timestamp = time.Now()
	// Add OCR cost/latency to totals (pipeline = OCR + structuration)
	result.LatencyMs += ocrLatency
	result.CostUSD += ocrCost
	result.TokensInput += ocrTokensIn
	return result
}

func buildHybridProviders(all bool, single, modelFilter string) []benchmark.HybridProvider {
	selectedModels := parseModelFilter(modelFilter)
	var providers []benchmark.HybridProvider
	for _, def := range modelCatalog {
		if def.HybridBuilder == nil {
			if shouldInclude(def, all, single, selectedModels) {
				fmt.Printf("  [skip] %s — no hybrid support\n", def.ID)
			}
			continue
		}
		if !shouldInclude(def, all, single, selectedModels) {
			continue
		}
		key := os.Getenv(def.EnvKey)
		if key == "" {
			fmt.Printf("  [skip] %s — %s not set\n", def.ID, def.EnvKey)
			continue
		}
		providers = append(providers, def.HybridBuilder(key))
	}
	return providers
}

func buildIDPProviders(all bool, single, modelFilter string) []benchmark.Provider {
	selectedModels := parseModelFilter(modelFilter)
	var providers []benchmark.Provider
	for _, def := range modelCatalog {
		if def.IDPBuilder == nil {
			continue
		}
		if !shouldInclude(def, all, single, selectedModels) {
			continue
		}
		key := os.Getenv(def.EnvKey)
		if key == "" {
			fmt.Printf("  [skip] %s — %s not set\n", def.ID, def.EnvKey)
			continue
		}
		providers = append(providers, def.IDPBuilder(key))
	}
	return providers
}

func buildOCRProviders(all bool, single, modelFilter string) []benchmark.OCRProvider {
	selectedModels := parseModelFilter(modelFilter)
	var providers []benchmark.OCRProvider
	for _, def := range modelCatalog {
		if def.OCRBuilder == nil {
			if shouldInclude(def, all, single, selectedModels) {
				fmt.Printf("  [skip] %s — no vision support\n", def.ID)
			}
			continue
		}
		if !shouldInclude(def, all, single, selectedModels) {
			continue
		}
		key := os.Getenv(def.EnvKey)
		if key == "" {
			fmt.Printf("  [skip] %s — %s not set\n", def.ID, def.EnvKey)
			continue
		}
		providers = append(providers, def.OCRBuilder(key))
	}
	return providers
}

func parseModelFilter(modelFilter string) map[string]bool {
	selectedModels := make(map[string]bool)
	if modelFilter != "" {
		for _, m := range strings.Split(modelFilter, ",") {
			selectedModels[strings.TrimSpace(m)] = true
		}
	}
	return selectedModels
}

func shouldInclude(def modelDef, all bool, single string, selectedModels map[string]bool) bool {
	if len(selectedModels) > 0 {
		return selectedModels[def.ID]
	}
	if single != "" {
		return def.Provider == single
	}
	return all
}

// --- Output: IDP ---

func outputIDPConsole(summaries []benchmark.RunSummary) {
	fmt.Println("╔══════════════════════════╦═══════╦═══════╦═══════╦═════════╦════════╦═══════════╗")
	fmt.Println("║ Modèle                   ║ Compl.║ Fidél.║ Hallu.║ $/item  ║ Lat.ms ║ Score     ║")
	fmt.Println("╠══════════════════════════╬═══════╬═══════╬═══════╬═════════╬════════╬═══════════╣")
	for _, s := range summaries {
		name := s.Model
		if len(name) > 24 {
			name = name[:24]
		}
		fmt.Printf("║ %-24s ║ %5.2f ║ %5.2f ║ %5.2f ║ %7.5f ║ %6d ║ %9.4f ║\n",
			name,
			s.AvgCompletenessScore,
			s.AvgFidelityScore,
			s.AvgHallucinationRate,
			s.AvgCostPerItem,
			s.AvgLatencyMs,
			s.AvgCompositeScore,
		)
	}
	fmt.Println("╚══════════════════════════╩═══════╩═══════╩═══════╩═════════╩════════╩═══════════╝")
}

func outputIDPJSON(summaries []benchmark.RunSummary) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(summaries)
}

func outputIDPCSV(summaries []benchmark.RunSummary) {
	fmt.Println("provider,model,completeness,classification,fidelity,keywords,hallucination,latency_ms,total_cost,cost_per_item,quality,composite")
	for _, s := range summaries {
		fmt.Printf("%s,%s,%.4f,%.4f,%.4f,%.4f,%.4f,%d,%.6f,%.6f,%.4f,%.4f\n",
			s.Provider, s.Model,
			s.AvgCompletenessScore, s.AvgClassificationScore,
			s.AvgFidelityScore, s.AvgKeywordScore,
			s.AvgHallucinationRate, s.AvgLatencyMs,
			s.TotalCostUSD, s.AvgCostPerItem,
			s.AvgQualityScore, s.AvgCompositeScore,
		)
	}
}

func saveIDPResults(summaries []benchmark.RunSummary) {
	ts := time.Now().Format("2006-01-02_15h04")
	saveIDPResultsTo(summaries, ts)
}

func saveIDPResultsTo(summaries []benchmark.RunSummary, runID string) {
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "idp")
	dir := filepath.Join(resultsDir, runID)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[IDP] Results saved to %s\n", path)

	reportPath := filepath.Join(dir, "report.html")
	if err := generateReport(resultsDir, runID, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate HTML report: %v\n", err)
	}
}

// --- Output: OCR ---

func outputOCRConsole(summaries []benchmark.OCRRunSummary) {
	fmt.Println("╔══════════════════════════╦═══════╦═══════╦═══════╦═════════╦════════╦═══════════╗")
	fmt.Println("║ Modèle                   ║ Détec.║ Texte ║ Types ║ Coût $  ║ Lat.ms ║ Score     ║")
	fmt.Println("╠══════════════════════════╬═══════╬═══════╬═══════╬═════════╬════════╬═══════════╣")
	for _, s := range summaries {
		name := s.Model
		if len(name) > 24 {
			name = name[:24]
		}
		fmt.Printf("║ %-24s ║ %5.2f ║ %5.2f ║ %5.2f ║ %7.5f ║ %6d ║ %9.4f ║\n",
			name,
			s.AvgDetectionScore,
			s.AvgTextAccuracy,
			s.AvgTypeAccuracy,
			s.TotalCostUSD,
			s.AvgLatencyMs,
			s.AvgCompositeScore,
		)
	}
	fmt.Println("╚══════════════════════════╩═══════╩═══════╩═══════╩═════════╩════════╩═══════════╝")
}

func outputOCRJSON(summaries []benchmark.OCRRunSummary) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(summaries)
}

func outputOCRCSV(summaries []benchmark.OCRRunSummary) {
	fmt.Println("provider,model,detection,text_accuracy,type_accuracy,latency_ms,total_cost,quality,composite")
	for _, s := range summaries {
		fmt.Printf("%s,%s,%.4f,%.4f,%.4f,%d,%.6f,%.4f,%.4f\n",
			s.Provider, s.Model,
			s.AvgDetectionScore, s.AvgTextAccuracy, s.AvgTypeAccuracy,
			s.AvgLatencyMs, s.TotalCostUSD,
			s.AvgQualityScore, s.AvgCompositeScore,
		)
	}
}

func saveOCRResults(summaries []benchmark.OCRRunSummary) {
	ts := time.Now().Format("2006-01-02_15h04")
	saveOCRResultsTo(summaries, ts)
}

func saveOCRResultsTo(summaries []benchmark.OCRRunSummary, runID string) {
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "ocr")
	dir := filepath.Join(resultsDir, runID)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[OCR] Results saved to %s\n", path)

	reportPath := filepath.Join(dir, "report.html")
	if err := generateOCRReport(resultsDir, runID, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate OCR HTML report: %v\n", err)
	}
}

// --- Shared helpers ---

// structurationPrompts returns the production prompts from infra/anthropic/prompts.go.
// This ensures the benchmark always tests the same prompts used in production.
func structurationPrompts(subject string, blocksJSON string) (systemPrompt, userPrompt string) {
	return llm.StructurationSystemPrompt, llm.BuildUserPrompt(subject, blocksJSON)
}

func loadTestCases(casesDir, filterID string) ([]benchmark.TestCase, error) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		return nil, fmt.Errorf("read cases dir: %w", err)
	}

	var cases []benchmark.TestCase
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if filterID != "" && e.Name() != filterID {
			continue
		}

		metaPath := filepath.Join(casesDir, e.Name(), "metadata.yaml")
		inputPath := filepath.Join(casesDir, e.Name(), "input.yaml")
		goldenPath := filepath.Join(casesDir, e.Name(), "golden_output.yaml")

		// Skip incomplete case directories (missing required YAML files).
		if _, err := os.Stat(metaPath); err != nil {
			continue
		}
		if _, err := os.Stat(inputPath); err != nil {
			continue
		}
		if _, err := os.Stat(goldenPath); err != nil {
			continue
		}

		var meta struct {
			ID         string `yaml:"id"`
			Subject    string `yaml:"subject"`
			Level      string `yaml:"level"`
			Topic      string `yaml:"topic"`
			Difficulty string `yaml:"difficulty"`
			HasImages  bool   `yaml:"has_images"`
		}
		if err := readYAML(metaPath, &meta); err != nil {
			return nil, fmt.Errorf("read metadata %s: %w", e.Name(), err)
		}

		var input struct {
			Blocks []benchmark.OCRBlock `yaml:"blocks"`
		}
		if err := readYAML(inputPath, &input); err != nil {
			return nil, fmt.Errorf("read input %s: %w", e.Name(), err)
		}

		var golden benchmark.GoldenOutput
		if err := readYAML(goldenPath, &golden); err != nil {
			return nil, fmt.Errorf("read golden %s: %w", e.Name(), err)
		}

		tc := benchmark.TestCase{
			ID:         meta.ID,
			Subject:    meta.Subject,
			Level:      meta.Level,
			Topic:      meta.Topic,
			Difficulty: meta.Difficulty,
			Blocks:     input.Blocks,
			Golden:     golden,
			HasImages:  meta.HasImages,
		}

		// Discover image paths if has_images is true
		if meta.HasImages {
			imagesDir := filepath.Join(casesDir, e.Name(), "images")
			imgEntries, err := os.ReadDir(imagesDir)
			if err == nil {
				for _, img := range imgEntries {
					if img.IsDir() {
						continue
					}
					ext := strings.ToLower(filepath.Ext(img.Name()))
					if ext == ".jpeg" || ext == ".jpg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
						tc.ImagePaths = append(tc.ImagePaths, filepath.Join(imagesDir, img.Name()))
					}
				}
			}
		}

		cases = append(cases, tc)
	}
	return cases, nil
}

func readYAML(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

func printModelCatalog() {
	fmt.Println("Available models:")
	fmt.Println()
	currentProvider := ""
	for _, def := range modelCatalog {
		if def.Provider != currentProvider {
			currentProvider = def.Provider
			keySet := "not set"
			if os.Getenv(def.EnvKey) != "" {
				keySet = "OK"
			}
			fmt.Printf("  %s (%s: %s)\n", strings.ToUpper(def.Provider), def.EnvKey, keySet)
		}
		vision := ""
		if def.OCRBuilder != nil {
			vision = " [vision]"
		}
		fmt.Printf("    - %s%s\n", def.ID, vision)
	}
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  benchmark --type=idp --models=gpt-4.1-mini,gemini-2.5-flash    # IDP benchmark")
	fmt.Println("  benchmark --type=ocr --models=gpt-4.1-mini,gemini-2.5-flash    # OCR benchmark")
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --all                              Run IDP (structuration) benchmark")
	fmt.Fprintln(os.Stderr, "  benchmark --type=ocr --all                              Run OCR (vision extraction) benchmark")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --provider=anthropic                Run all models for a provider")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --models=gpt-4.1-mini,gemini-2.5-flash   Pick specific models")
	fmt.Fprintln(os.Stderr, "  benchmark --list-models                                 List available models")
	fmt.Fprintln(os.Stderr, "  benchmark --report                                      Generate HTML report")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Benchmark types:")
	fmt.Fprintln(os.Stderr, "  idp  — Structuration LLM: OCR blocks → items (default)")
	fmt.Fprintln(os.Stderr, "  ocr  — Vision/OCR: images → OCR blocks")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "API keys: ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_AI_API_KEY, MISTRAL_API_KEY, DEEPSEEK_API_KEY")
}

func idpProviderNames(providers []benchmark.Provider) string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name() + "/" + p.ModelID()
	}
	return strings.Join(names, ", ")
}

func ocrProviderNames(providers []benchmark.OCRProvider) string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name() + "/" + p.ModelID()
	}
	return strings.Join(names, ", ")
}

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}
