// Command benchmark runs the LLM benchmark suite against configured providers.
//
// Usage:
//
//	go run ./cmd/benchmark/ --all
//	go run ./cmd/benchmark/ --provider=anthropic
//	go run ./cmd/benchmark/ --models=gpt-4o,claude-sonnet-4-6,gemini-2.5-flash
//	go run ./cmd/benchmark/ --list-models
//	go run ./cmd/benchmark/ --case=01_physique_densite
//	go run ./cmd/benchmark/ --all --runs=3 --output=csv
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
	"github.com/popul/revisemieux/internal/infra/openaicompat"
)

// modelDef describes a benchmarkable LLM model.
type modelDef struct {
	ID       string // unique key used in --models flag
	Provider string // provider name for display
	EnvKey   string // environment variable for API key
	Builder  func(apiKey string) benchmark.Provider
}

// modelCatalog lists all available models, grouped by provider.
var modelCatalog = []modelDef{
	// Anthropic
	{
		ID: "claude-sonnet-4-6", Provider: "anthropic", EnvKey: "ANTHROPIC_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return llmanthro.NewBenchmarkProvider(k, "claude-sonnet-4-6", 3.00, 15.00)
		},
	},
	{
		ID: "claude-haiku-4-5", Provider: "anthropic", EnvKey: "ANTHROPIC_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return llmanthro.NewBenchmarkProvider(k, "claude-haiku-4-5", 1.00, 5.00)
		},
	},
	// OpenAI
	{
		ID: "gpt-4o", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o", PriceIn: 2.50, PriceOut: 10.00,
			})
		},
	},
	{
		ID: "gpt-4o-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "gpt-4o-mini", PriceIn: 0.15, PriceOut: 0.60,
			})
		},
	},
	{
		ID: "o3-mini", Provider: "openai", EnvKey: "OPENAI_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.openai.com/v1", APIKey: k, Name: "OpenAI",
				Model: "o3-mini", PriceIn: 1.10, PriceOut: 4.40,
			})
		},
	},
	// Google Gemini (OpenAI-compatible endpoint)
	{
		ID: "gemini-2.5-flash", Provider: "google", EnvKey: "GOOGLE_AI_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-flash", PriceIn: 0.15, PriceOut: 0.60,
			})
		},
	},
	{
		ID: "gemini-2.5-pro", Provider: "google", EnvKey: "GOOGLE_AI_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", APIKey: k, Name: "Google",
				Model: "gemini-2.5-pro", PriceIn: 1.25, PriceOut: 10.00,
			})
		},
	},
	// Mistral
	{
		ID: "mistral-large-latest", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-large-latest", PriceIn: 2.00, PriceOut: 6.00,
			})
		},
	},
	{
		ID: "mistral-small-latest", Provider: "mistral", EnvKey: "MISTRAL_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.mistral.ai/v1", APIKey: k, Name: "Mistral",
				Model: "mistral-small-latest", PriceIn: 0.10, PriceOut: 0.30,
			})
		},
	},
	// DeepSeek
	{
		ID: "deepseek-chat", Provider: "deepseek", EnvKey: "DEEPSEEK_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.deepseek.com", APIKey: k, Name: "DeepSeek",
				Model: "deepseek-chat", PriceIn: 0.27, PriceOut: 1.10,
			})
		},
	},
	{
		ID: "deepseek-reasoner", Provider: "deepseek", EnvKey: "DEEPSEEK_API_KEY",
		Builder: func(k string) benchmark.Provider {
			return openaicompat.NewBenchmarkProvider(openaicompat.Config{
				BaseURL: "https://api.deepseek.com", APIKey: k, Name: "DeepSeek",
				Model: "deepseek-reasoner", PriceIn: 0.55, PriceOut: 2.19,
			})
		},
	},
}

func main() {
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
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  benchmark --all                              Run all models (needs API keys)")
		fmt.Fprintln(os.Stderr, "  benchmark --provider=anthropic               Run all models for a provider")
		fmt.Fprintln(os.Stderr, "  benchmark --models=gpt-4o,claude-sonnet-4-6  Pick specific models")
		fmt.Fprintln(os.Stderr, "  benchmark --list-models                      List available models")
		fmt.Fprintln(os.Stderr, "  benchmark --report                           Generate HTML report")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "API keys: ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_AI_API_KEY, MISTRAL_API_KEY, DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	// Load test cases
	casesDir := filepath.Join(testdataDir(), "benchmark", "cases")
	cases, err := loadTestCases(casesDir, *caseID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading test cases: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d test case(s)\n", len(cases))

	// Build provider list
	providers := buildProviders(*all, *provider, *models)
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No providers configured (check API keys)")
		os.Exit(1)
	}
	fmt.Printf("Running %d provider(s): %s\n", len(providers), providerNames(providers))
	fmt.Printf("Runs per case: %d\n\n", *runs)

	// Run benchmark
	var allResults []benchmark.EvalResult
	for _, p := range providers {
		fmt.Printf("--- %s (%s) ---\n", p.Name(), p.ModelID())
		for _, tc := range cases {
			for run := 0; run < *runs; run++ {
				result := runSingleCase(p, tc)
				if *runs > 1 {
					fmt.Printf("  [%s] run %d/%d: Q=%.2f cost=$%.5f latency=%dms\n",
						tc.ID, run+1, *runs, result.QualityScore, result.CostUSD, result.LatencyMs)
				} else {
					fmt.Printf("  [%s] Q=%.2f cost=$%.5f latency=%dms items=%d\n",
						tc.ID, result.QualityScore, result.CostUSD, result.LatencyMs, result.ItemsFound)
				}
				allResults = append(allResults, result)
			}
		}
	}

	// Compute composite scores
	benchmark.ComputeCompositeScores(allResults)

	// Build summaries per provider
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

	// Sort by composite score
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	// Output
	fmt.Println()
	switch *output {
	case "json":
		outputJSON(summaries)
	case "csv":
		outputCSV(summaries)
	default:
		outputConsole(summaries)
	}

	// Save results and generate HTML report
	saveResults(summaries)
}

func runSingleCase(p benchmark.Provider, tc benchmark.TestCase) benchmark.EvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Build the prompt from test case blocks
	blocksJSON, _ := json.Marshal(tc.Blocks)
	systemPrompt := structurationSystemPrompt()
	userPrompt := fmt.Sprintf("Matière : %s\n\nBlocs OCR :\n%s", tc.Subject, string(blocksJSON))

	// Call the LLM
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

	// Parse the response JSON
	var parsed benchmark.ParsedOutput
	if err := json.Unmarshal(resp.RawJSON, &parsed); err != nil {
		result := benchmark.Evaluate(tc, nil, resp, p)
		result.Timestamp = time.Now()
		result.Error = fmt.Sprintf("JSON parse error: %v", err)
		return result
	}

	result := benchmark.Evaluate(tc, &parsed, resp, p)
	result.Timestamp = time.Now()
	return result
}

func structurationSystemPrompt() string {
	return `Tu es un assistant pédagogique spécialisé dans l'extraction de connaissances à partir de cours de collégiens français.

Ta tâche : à partir de blocs de texte OCR extraits d'une photo de cahier, tu dois produire des items de révision structurés.

## Types d'items

- KNOWLEDGE : fait, définition, propriété à mémoriser
- PROCEDURE : formule, méthode de calcul, étapes à suivre
- DOCUMENT : référence à un schéma, carte, tableau ou image
- WRITING : rédaction, argumentation, texte à produire

## Règles

1. Chaque item doit être FIDÈLE au texte source. Ne jamais inventer de contenu absent du texte OCR.
2. Le "term" est la phrase ou formule clé telle qu'elle apparaît dans le cours.
3. Les "keywords" sont les mots-clés qui serviront à générer des questions (cloze, QCM).
4. Les "steps" sont obligatoires pour les items PROCEDURE (étapes de la méthode).
5. Regroupe les items en "notions" (clusters sémantiques, 2-7 par chapitre).
6. Attribue un score de "confidence" (0-1) reflétant la certitude de l'extraction.
7. Confidence < 0.7 si le texte OCR est ambigu ou partiellement lisible.

## Format de sortie

Réponds UNIQUEMENT avec un JSON valide, sans markdown, sans commentaire :

{
  "items": [
    {
      "type": "KNOWLEDGE",
      "term": "phrase exacte du cours",
      "keywords": ["mot1", "mot2"],
      "steps": [],
      "notion_name": "Nom de la notion",
      "confidence": 0.92
    }
  ],
  "notions": ["Notion 1", "Notion 2"]
}`
}

func outputConsole(summaries []benchmark.RunSummary) {
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

func outputJSON(summaries []benchmark.RunSummary) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(summaries)
}

func outputCSV(summaries []benchmark.RunSummary) {
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

func saveResults(summaries []benchmark.RunSummary) {
	resultsDir := filepath.Join(testdataDir(), "benchmark", "results")
	ts := time.Now().Format("2006-01-02_15h04")
	dir := filepath.Join(resultsDir, ts)
	os.MkdirAll(dir, 0o755)

	data, _ := json.MarshalIndent(summaries, "", "  ")
	path := filepath.Join(dir, "summary.json")
	os.WriteFile(path, data, 0o644)
	fmt.Printf("\nResults saved to %s\n", path)

	// Auto-generate HTML report
	reportPath := filepath.Join(dir, "report.html")
	if err := generateReport(resultsDir, ts, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate HTML report: %v\n", err)
	}
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

		cases = append(cases, benchmark.TestCase{
			ID:         meta.ID,
			Subject:    meta.Subject,
			Level:      meta.Level,
			Topic:      meta.Topic,
			Difficulty: meta.Difficulty,
			Blocks:     input.Blocks,
			Golden:     golden,
		})
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

func buildProviders(all bool, single, modelFilter string) []benchmark.Provider {
	// Parse --models filter into a set
	selectedModels := make(map[string]bool)
	if modelFilter != "" {
		for _, m := range strings.Split(modelFilter, ",") {
			selectedModels[strings.TrimSpace(m)] = true
		}
	}

	var providers []benchmark.Provider
	for _, def := range modelCatalog {
		// Filter: --models takes priority, then --provider, then --all
		if len(selectedModels) > 0 {
			if !selectedModels[def.ID] {
				continue
			}
		} else if single != "" {
			if def.Provider != single {
				continue
			}
		} else if !all {
			continue
		}

		key := os.Getenv(def.EnvKey)
		if key == "" {
			fmt.Printf("  [skip] %s — %s not set\n", def.ID, def.EnvKey)
			continue
		}
		providers = append(providers, def.Builder(key))
	}
	return providers
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
		fmt.Printf("    - %s\n", def.ID)
	}
	fmt.Println()
	fmt.Println("Usage: benchmark --models=gpt-4o,claude-sonnet-4-6,gemini-2.5-flash")
}

func providerNames(providers []benchmark.Provider) string {
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
