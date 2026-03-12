// Command benchmark runs the LLM benchmark suite against configured providers.
//
// Usage:
//
//	go run ./cmd/benchmark/ --all
//	go run ./cmd/benchmark/ --provider=anthropic
//	go run ./cmd/benchmark/ --case=01_physique_densite
//	go run ./cmd/benchmark/ --all --runs=3 --output=csv
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
)

func main() {
	all := flag.Bool("all", false, "Run all providers")
	provider := flag.String("provider", "", "Run specific provider (anthropic, openai, google, mistral, deepseek)")
	caseID := flag.String("case", "", "Run specific test case")
	runs := flag.Int("runs", 1, "Number of runs per case (for variance measurement)")
	output := flag.String("output", "console", "Output format: console, json, csv")
	flag.Parse()

	if !*all && *provider == "" {
		fmt.Fprintln(os.Stderr, "Usage: benchmark --all or benchmark --provider=<name>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Providers: anthropic, openai, google, mistral, deepseek")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Set API keys via environment variables:")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_API_KEY, OPENAI_API_KEY, GOOGLE_AI_API_KEY, MISTRAL_API_KEY, DEEPSEEK_API_KEY")
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
	providers := buildProviders(*all, *provider)
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

	// Save results
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

func buildProviders(all bool, single string) []benchmark.Provider {
	var providers []benchmark.Provider

	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && (all || single == "anthropic") {
		providers = append(providers,
			llmanthro.NewBenchmarkProvider(key, "claude-sonnet-4-6", 3.00, 15.00),
			llmanthro.NewBenchmarkProvider(key, "claude-haiku-4-5", 1.00, 5.00),
		)
	}

	// TODO: add OpenAI, Google, Mistral, DeepSeek providers
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && (all || single == "openai") {
		fmt.Println("  [skip] OpenAI provider not yet implemented")
	}
	if key := os.Getenv("GOOGLE_AI_API_KEY"); key != "" && (all || single == "google") {
		fmt.Println("  [skip] Google provider not yet implemented")
	}
	if key := os.Getenv("MISTRAL_API_KEY"); key != "" && (all || single == "mistral") {
		fmt.Println("  [skip] Mistral provider not yet implemented")
	}
	if key := os.Getenv("DEEPSEEK_API_KEY"); key != "" && (all || single == "deepseek") {
		fmt.Println("  [skip] DeepSeek provider not yet implemented")
	}

	return providers
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
