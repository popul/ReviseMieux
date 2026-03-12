package benchmark

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Quality indicator weights (must sum to 1.0).
const (
	weightCompleteness   = 0.25
	weightClassification = 0.15
	weightFidelity       = 0.20
	weightKeywords       = 0.10
	weightHallucination  = 0.20
	weightNotions        = 0.05
	weightSchema         = 0.05

	qualityWeight = 0.70
	costWeight    = 0.30
)

// Evaluate computes all quality and performance metrics for a single run.
func Evaluate(tc TestCase, parsed *ParsedOutput, resp *Response, provider Provider) EvalResult {
	result := EvalResult{
		CaseID:   tc.ID,
		Provider: provider.Name(),
		Model:    provider.ModelID(),
	}

	if resp != nil {
		result.LatencyMs = resp.LatencyMs
		result.TokensInput = resp.TokensInput
		result.TokensOutput = resp.TokensOutput
		result.CostUSD = float64(resp.TokensInput)*provider.PricePerMInput()/1_000_000 +
			float64(resp.TokensOutput)*provider.PricePerMOutput()/1_000_000
	}

	if parsed == nil {
		result.SchemaCompliance = false
		return result
	}

	result.SchemaCompliance = true
	result.ItemsFound = len(parsed.Items)

	if result.ItemsFound > 0 {
		result.CostPerItem = result.CostUSD / float64(result.ItemsFound)
	}

	// Source text for fidelity/hallucination checks
	sourceText := buildSourceText(tc.Blocks)

	// Q1 — Completeness
	matches := matchItems(tc.Golden.Items, parsed.Items)
	if len(tc.Golden.Items) > 0 {
		result.CompletenessScore = float64(len(matches)) / float64(len(tc.Golden.Items))
	}

	// Q2 — Classification accuracy
	correctTypes := 0
	for _, m := range matches {
		if strings.EqualFold(m.golden.Type, m.parsed.Type) {
			correctTypes++
		}
	}
	if len(matches) > 0 {
		result.ClassificationScore = float64(correctTypes) / float64(len(matches))
	}

	// Q3 — Fidelity to source
	traceable := 0
	for _, item := range parsed.Items {
		if isTraceable(item.Term, sourceText) {
			traceable++
		}
	}
	if len(parsed.Items) > 0 {
		result.FidelityScore = float64(traceable) / float64(len(parsed.Items))
	}

	// Q4 — Keyword quality (Jaccard)
	totalJaccard := 0.0
	jaccardCount := 0
	for _, m := range matches {
		j := jaccard(m.golden.Keywords, m.parsed.Keywords)
		totalJaccard += j
		jaccardCount++
	}
	if jaccardCount > 0 {
		result.KeywordScore = totalJaccard / float64(jaccardCount)
	}

	// Q5 — Hallucination rate
	hallucinated := 0
	for _, item := range parsed.Items {
		if !isTraceable(item.Term, sourceText) {
			hallucinated++
		}
	}
	if len(parsed.Items) > 0 {
		result.HallucinationRate = float64(hallucinated) / float64(len(parsed.Items))
	}

	// Q6 — Notion coherence
	if len(tc.Golden.Notions) > 0 {
		notionMatches := 0
		for _, gn := range tc.Golden.Notions {
			for _, pn := range parsed.Notions {
				if fuzzyMatch(gn, pn) {
					notionMatches++
					break
				}
			}
		}
		result.NotionScore = float64(notionMatches) / float64(len(tc.Golden.Notions))
	}

	// Composite scores
	result.QualityScore = result.CompletenessScore*weightCompleteness +
		result.ClassificationScore*weightClassification +
		result.FidelityScore*weightFidelity +
		result.KeywordScore*weightKeywords +
		(1-result.HallucinationRate)*weightHallucination +
		result.NotionScore*weightNotions
	if result.SchemaCompliance {
		result.QualityScore += weightSchema
	}

	return result
}

// ComputeCompositeScores sets the CompositeScore on each result using the max cost across all results.
func ComputeCompositeScores(results []EvalResult) {
	maxCost := 0.0
	for _, r := range results {
		if r.CostUSD > maxCost {
			maxCost = r.CostUSD
		}
	}
	for i := range results {
		costEfficiency := 0.0
		if maxCost > 0 {
			costEfficiency = 1 - (results[i].CostUSD / maxCost)
		}
		results[i].CompositeScore = results[i].QualityScore*qualityWeight + costEfficiency*costWeight
	}
}

// Summarize aggregates results for a single provider across all cases.
func Summarize(provider string, model string, results []EvalResult) RunSummary {
	s := RunSummary{
		Provider: provider,
		Model:    model,
		Results:  results,
	}
	if len(results) == 0 {
		return s
	}

	errors := 0
	for _, r := range results {
		if r.Error != "" {
			errors++
			continue
		}
		s.AvgCompletenessScore += r.CompletenessScore
		s.AvgClassificationScore += r.ClassificationScore
		s.AvgFidelityScore += r.FidelityScore
		s.AvgKeywordScore += r.KeywordScore
		s.AvgHallucinationRate += r.HallucinationRate
		s.AvgLatencyMs += r.LatencyMs
		s.TotalCostUSD += r.CostUSD
		s.AvgCostPerItem += r.CostPerItem
		s.AvgQualityScore += r.QualityScore
		s.AvgCompositeScore += r.CompositeScore
	}

	successful := len(results) - errors
	if successful > 0 {
		n := float64(successful)
		s.AvgCompletenessScore /= n
		s.AvgClassificationScore /= n
		s.AvgFidelityScore /= n
		s.AvgKeywordScore /= n
		s.AvgHallucinationRate /= n
		s.AvgLatencyMs /= int64(successful)
		s.AvgCostPerItem /= n
		s.AvgQualityScore /= n
		s.AvgCompositeScore /= n
	}
	s.ErrorRate = float64(errors) / float64(len(results))

	return s
}

// --- Helpers ---

type itemMatch struct {
	golden GoldenItem
	parsed ParsedItem
}

// matchItems finds the best matches between golden and parsed items using fuzzy term matching.
func matchItems(golden []GoldenItem, parsed []ParsedItem) []itemMatch {
	used := make(map[int]bool)
	var matches []itemMatch

	for _, g := range golden {
		bestIdx := -1
		bestScore := 0.0
		for i, p := range parsed {
			if used[i] {
				continue
			}
			score := termSimilarity(g.Term, p.Term)
			if score > bestScore && score > 0.4 {
				bestScore = score
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			used[bestIdx] = true
			matches = append(matches, itemMatch{golden: g, parsed: parsed[bestIdx]})
		}
	}
	return matches
}

// termSimilarity returns a similarity score between two terms (0-1).
func termSimilarity(a, b string) float64 {
	na := normalize(a)
	nb := normalize(b)

	// Exact match
	if na == nb {
		return 1.0
	}

	// Substring match
	if strings.Contains(na, nb) || strings.Contains(nb, na) {
		shorter := len(na)
		if len(nb) < shorter {
			shorter = len(nb)
		}
		longer := len(na)
		if len(nb) > longer {
			longer = len(nb)
		}
		if longer > 0 {
			return float64(shorter) / float64(longer)
		}
	}

	// Word overlap (Jaccard on words)
	wordsA := strings.Fields(na)
	wordsB := strings.Fields(nb)
	return jaccard(wordsA, wordsB)
}

// isTraceable checks if a term can be traced back to the source OCR text.
func isTraceable(term, sourceText string) bool {
	nTerm := normalize(term)
	nSource := normalize(sourceText)

	// Direct substring
	if strings.Contains(nSource, nTerm) {
		return true
	}

	// Check if significant words of the term appear in source
	words := strings.Fields(nTerm)
	significantWords := filterStopWords(words)
	if len(significantWords) == 0 {
		return true // only stop words, consider traceable
	}

	found := 0
	for _, w := range significantWords {
		if strings.Contains(nSource, w) {
			found++
		}
	}
	return float64(found)/float64(len(significantWords)) >= 0.6
}

// jaccard computes the Jaccard similarity between two string slices.
func jaccard(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	setA := make(map[string]bool, len(a))
	for _, s := range a {
		setA[normalize(s)] = true
	}
	setB := make(map[string]bool, len(b))
	for _, s := range b {
		setB[normalize(s)] = true
	}

	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 1.0
	}
	return float64(intersection) / float64(union)
}

// fuzzyMatch checks if two strings are similar enough.
func fuzzyMatch(a, b string) bool {
	return termSimilarity(a, b) > 0.5
}

// normalize lowercases, removes accents and extra whitespace.
func normalize(s string) string {
	s = strings.ToLower(s)
	// Remove accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return strings.Join(strings.Fields(result), " ")
}

// buildSourceText concatenates all OCR block texts into a single string.
func buildSourceText(blocks []OCRBlock) string {
	var parts []string
	for _, b := range blocks {
		parts = append(parts, b.Text)
	}
	return strings.Join(parts, " ")
}

// French stop words to exclude from traceability checks.
var stopWords = map[string]bool{
	"le": true, "la": true, "les": true, "un": true, "une": true, "des": true,
	"de": true, "du": true, "d": true, "l": true,
	"et": true, "ou": true, "est": true, "a": true, "en": true,
	"qui": true, "que": true, "quoi": true,
	"dans": true, "sur": true, "par": true, "pour": true, "avec": true,
	"son": true, "sa": true, "ses": true, "ce": true, "cette": true,
	"il": true, "elle": true, "on": true, "ne": true, "pas": true,
}

func filterStopWords(words []string) []string {
	var result []string
	for _, w := range words {
		if len(w) > 1 && !stopWords[w] {
			result = append(result, w)
		}
	}
	return result
}
