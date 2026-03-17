// Command benchmark runs the OCR and IDP benchmark suites against configured providers.
//
// Usage:
//
//	go run ./cmd/benchmark/ --type=idp --all
//	go run ./cmd/benchmark/ --type=ocr --all
//	go run ./cmd/benchmark/ --type=idp --provider=anthropic
//	go run ./cmd/benchmark/ --type=idp --models=gpt-4o,claude-sonnet-4-6,gemini-2.5-flash
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
	"time"

	"github.com/popul/revisemieux/internal/benchmark"
	llmanthro "github.com/popul/revisemieux/internal/infra/anthropic"
	"github.com/popul/revisemieux/internal/infra/llm"
	"github.com/popul/revisemieux/internal/infra/openaicompat"
)

// modelDef describes a benchmarkable LLM model.
type modelDef struct {
	ID         string // unique key used in --models flag
	Provider   string // provider name for display
	EnvKey     string // environment variable for API key
	IDPBuilder func(apiKey string) benchmark.Provider    // builder for IDP benchmark
	OCRBuilder func(apiKey string) benchmark.OCRProvider // builder for OCR benchmark (nil if not supported)
}

// modelCatalog lists all available models, grouped by provider.
var modelCatalog = []modelDef{
	// Anthropic
	{
		ID: "claude-sonnet-4-6", Provider: "anthropic", EnvKey: "ANTHROPIC_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return llmanthro.NewBenchmarkProvider(k, "claude-sonnet-4-6", 3.00, 15.00)
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return llmanthro.NewOCRBenchmarkProvider(k, "claude-sonnet-4-6", 3.00, 15.00)
		},
	},
	{
		ID: "claude-haiku-4-5", Provider: "anthropic", EnvKey: "ANTHROPIC_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return llmanthro.NewBenchmarkProvider(k, "claude-haiku-4-5", 1.00, 5.00)
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return llmanthro.NewOCRBenchmarkProvider(k, "claude-haiku-4-5", 1.00, 5.00)
		},
	},
	// OpenAI
	{
		ID: "gpt-4o", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o", PriceIn: 2.50, PriceOut: 10.00,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o", PriceIn: 2.50, PriceOut: 10.00,
			})
		},
	},
	{
		ID: "gpt-4o-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o-mini", PriceIn: 0.15, PriceOut: 0.60,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o-mini", PriceIn: 0.15, PriceOut: 0.60,
			})
		},
	},
	{
		ID: "o3-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "o3-mini", PriceIn: 1.10, PriceOut: 4.40,
				ReasoningModel: true,
			})
		},
		// o3-mini does not support vision
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
	{
		ID: "gemini-2.5-pro", Provider: "google", EnvKey: "GOOGLE_AI_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-pro", PriceIn: 1.25, PriceOut: 10.00,
				ReasoningModel: true,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-pro", PriceIn: 1.25, PriceOut: 10.00,
			})
		},
	},
	// Mistral
	{
		ID: "mistral-large-latest", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-large-latest", PriceIn: 2.00, PriceOut: 6.00,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-large-latest", PriceIn: 2.00, PriceOut: 6.00,
			})
		},
	},
	{
		ID: "mistral-small-latest", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
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
		ID: "mistral-small-3.2", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
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
		ID: "deepseek-chat", Provider: "deepseek", EnvKey: "DEEPSEEK_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.deepseek.com", APIKey: k, Name: "DeepSeek",
				Model: "deepseek-chat", PriceIn: 0.27, PriceOut: 1.10,
			})
		},
		// deepseek-chat does not support vision
	},
	{
		ID: "deepseek-reasoner", Provider: "deepseek", EnvKey: "DEEPSEEK_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.deepseek.com", APIKey: k, Name: "DeepSeek",
				Model: "deepseek-reasoner", PriceIn: 0.55, PriceOut: 2.19,
				ReasoningModel: true,
			})
		},
		// deepseek-reasoner does not support vision
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
	// MiniMax (via OpenRouter)
	{
		ID: "minimax-m2.5", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "minimax/minimax-m2.5", PriceIn: 0.25, PriceOut: 1.20,
			})
		},
		// no vision support
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
	// NVIDIA (via OpenRouter)
	{
		ID: "nemotron-nano-vl", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "nvidia/nemotron-nano-12b-v2-vl:free", PriceIn: 0.00, PriceOut: 0.00,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "nvidia/nemotron-nano-12b-v2-vl:free", PriceIn: 0.00, PriceOut: 0.00,
			})
		},
	},
	// Google Gemma 3 (via OpenRouter)
	{
		ID: "gemma3-27b", Provider: "openrouter", EnvKey: "OPENROUTER_API_KEY",
		IDPBuilder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "google/gemma-3-27b-it", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
		OCRBuilder: func(k string) benchmark.OCRProvider {
			return openaicompat.NewOCRBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://openrouter.ai/api/v1", APIKey: k, Name: "OpenRouter",
				Model: "google/gemma-3-27b-it", PriceIn: 0.10, PriceOut: 0.10,
			})
		},
	},
}

func main() {
	benchType := flag.String("type", "idp", "Benchmark type: ocr, idp")
	all := flag.Bool("all", false, "Run all providers with available API keys")
	provider := flag.String("provider", "", "Run all models for a provider (anthropic, openai, google, mistral, deepseek)")
	models := flag.String("models", "", "Comma-separated list of model IDs to run (e.g. claude-sonnet-4-6,gpt-4o,gemini-2.5-flash)")
	listModels := flag.Bool("list-models", false, "List all available models and exit")
	caseID := flag.String("case", "", "Run specific test case")
	runs := flag.Int("runs", 1, "Number of runs per case (for variance measurement)")
	output := flag.String("output", "console", "Output format: console, json, csv")
	report := flag.Bool("report", false, "Generate HTML report from latest results (no benchmark run)")
	reportRun := flag.String("report-run", "", "Generate report from a specific run directory")
	reportOutput := flag.String("report-output", "", "Output path for the HTML report")
	flag.Parse()

	// List models mode
	if *listModels {
		printModelCatalog()
		os.Exit(0)
	}

	// Report-only mode
	if *report || *reportRun != "" {
		resultsDir := filepath.Join(testdataDir(), "benchmark", "results")
		if err := generateReport(resultsDir, *reportRun, *reportOutput); err != nil {
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
		runOCRBenchmark(cases, *all, *provider, *models, *runs, *output)
	case "idp":
		runIDPBenchmark(cases, *all, *provider, *models, *runs, *output)
	default:
		fmt.Fprintf(os.Stderr, "Unknown benchmark type: %s (use 'ocr' or 'idp')\n", *benchType)
		os.Exit(1)
	}
}

// --- IDP Benchmark (structuration LLM) ---

func runIDPBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter string, runs int, output string) {
	fmt.Printf("[IDP] Loaded %d test case(s)\n", len(cases))

	providers := buildIDPProviders(all, single, modelFilter)
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No IDP providers configured (check API keys)")
		os.Exit(1)
	}
	fmt.Printf("[IDP] Running %d provider(s): %s\n", len(providers), idpProviderNames(providers))
	fmt.Printf("[IDP] Runs per case: %d\n\n", runs)

	var allResults []benchmark.EvalResult
	for _, p := range providers {
		fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
		for _, tc := range cases {
			for run := 0; run < runs; run++ {
				result := runSingleIDPCase(p, tc)
				if runs > 1 {
					fmt.Printf("  [%s] run %d/%d: Q=%.2f cost=$%.5f latency=%dms\n",
						tc.ID, run+1, runs, result.QualityScore, result.CostUSD, result.LatencyMs)
				} else {
					fmt.Printf("  [%s] Q=%.2f cost=$%.5f latency=%dms items=%d\n",
						tc.ID, result.QualityScore, result.CostUSD, result.LatencyMs, result.ItemsFound)
				}
				allResults = append(allResults, result)
			}
		}
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
	switch output {
	case "json":
		outputIDPJSON(summaries)
	case "csv":
		outputIDPCSV(summaries)
	default:
		outputIDPConsole(summaries)
	}

	saveIDPResults(summaries)
}

func runSingleIDPCase(p benchmark.Provider, tc benchmark.TestCase) benchmark.EvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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

func runOCRBenchmark(cases []benchmark.TestCase, all bool, single, modelFilter string, runs int, output string) {
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
	fmt.Printf("[OCR] Running %d provider(s): %s\n", len(providers), ocrProviderNames(providers))
	fmt.Printf("[OCR] Runs per case: %d\n\n", runs)

	var allResults []benchmark.OCREvalResult
	for _, p := range providers {
		fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
		for _, tc := range ocrCases {
			for run := 0; run < runs; run++ {
				result := runSingleOCRCase(p, tc)
				if runs > 1 {
					fmt.Printf("  [%s] run %d/%d: Q=%.2f cost=$%.5f latency=%dms blocks=%d/%d\n",
						tc.ID, run+1, runs, result.QualityScore, result.CostUSD, result.LatencyMs,
						result.BlocksFound, result.BlocksExpected)
				} else {
					fmt.Printf("  [%s] Q=%.2f text=%.2f detect=%.2f cost=$%.5f latency=%dms blocks=%d/%d\n",
						tc.ID, result.QualityScore, result.TextAccuracy, result.DetectionScore,
						result.CostUSD, result.LatencyMs, result.BlocksFound, result.BlocksExpected)
				}
				allResults = append(allResults, result)
			}
		}
	}

	benchmark.ComputeOCRCompositeScores(allResults)

	summaryMap := make(map[string][]benchmark.OCREvalResult)
	for _, r := range allResults {
		key := r.Provider + "|" + r.Model
		summaryMap[key] = append(summaryMap[key], r)
	}

	var summaries []benchmark.OCRRunSummary
	for key, results := range summaryMap {
		parts := strings.SplitN(key, "|", 2)
		s := benchmark.SummarizeOCR(parts[0], parts[1], results)
		summaries = append(summaries, s)
	}

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

	saveOCRResults(summaries)
}

func runSingleOCRCase(p benchmark.OCRProvider, tc benchmark.TestCase) benchmark.OCREvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

func buildIDPProviders(all bool, single, modelFilter string) []benchmark.Provider {
	selectedModels := parseModelFilter(modelFilter)
	var providers []benchmark.Provider
	for _, def := range modelCatalog {
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
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "idp")
	ts := time.Now().Format("2006-01-02_15h04")
	dir := filepath.Join(resultsDir, ts)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[IDP] Results saved to %s\n", path)

	reportPath := filepath.Join(dir, "report.html")
	if err := generateReport(resultsDir, ts, reportPath); err != nil {
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
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results", "ocr")
	ts := time.Now().Format("2006-01-02_15h04")
	dir := filepath.Join(resultsDir, ts)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\n[OCR] Results saved to %s\n", path)

	reportPath := filepath.Join(dir, "report.html")
	if err := generateOCRReport(resultsDir, ts, reportPath); err != nil {
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

		metaPath := filepath.Join(casesDir, e.Name(), "metadata.json")
		inputPath := filepath.Join(casesDir, e.Name(), "input.json")
		goldenPath := filepath.Join(casesDir, e.Name(), "golden_output.json")

		var meta struct {
			ID         string `json:"id"`
			Subject    string `json:"subject"`
			Level      string `json:"level"`
			Topic      string `json:"topic"`
			Difficulty string `json:"difficulty"`
			HasImages  bool   `json:"has_images"`
		}
		if err := readJSON(metaPath, &meta); err != nil {
			return nil, fmt.Errorf("read metadata %s: %w", e.Name(), err)
		}

		var input struct {
			Blocks []benchmark.OCRBlock `json:"blocks"`
		}
		if err := readJSON(inputPath, &input); err != nil {
			return nil, fmt.Errorf("read input %s: %w", e.Name(), err)
		}

		var golden benchmark.GoldenOutput
		if err := readJSON(goldenPath, &golden); err != nil {
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

func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
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
	fmt.Println("  benchmark --type=idp --models=gpt-4o,claude-sonnet-4-6    # IDP benchmark")
	fmt.Println("  benchmark --type=ocr --models=gpt-4o,claude-sonnet-4-6    # OCR benchmark")
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --all                              Run IDP (structuration) benchmark")
	fmt.Fprintln(os.Stderr, "  benchmark --type=ocr --all                              Run OCR (vision extraction) benchmark")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --provider=anthropic                Run all models for a provider")
	fmt.Fprintln(os.Stderr, "  benchmark --type=idp --models=gpt-4o,claude-sonnet-4-6   Pick specific models")
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
