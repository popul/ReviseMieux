package benchmark

import (
	"encoding/json"
	"strings"
)

// OCR quality indicator weights (must sum to 1.0).
const (
	ocrWeightDetection = 0.30
	ocrWeightText      = 0.50
	ocrWeightType      = 0.20

	ocrQualityWeight = 0.70
	ocrCostWeight    = 0.30
)

// EvaluateOCR computes all quality metrics for one OCR benchmark run.
func EvaluateOCR(tc TestCase, produced []OCRBlock, resp *Response, provider OCRProvider) OCREvalResult {
	result := OCREvalResult{
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

	if produced == nil {
		return result
	}

	golden := tc.Blocks
	result.BlocksExpected = len(golden)
	result.BlocksFound = len(produced)

	// Detection score: how well did it find the right number of blocks
	if result.BlocksExpected > 0 || result.BlocksFound > 0 {
		minB := result.BlocksExpected
		maxB := result.BlocksFound
		if result.BlocksFound < minB {
			minB = result.BlocksFound
		}
		if result.BlocksExpected > maxB {
			maxB = result.BlocksExpected
		}
		result.DetectionScore = float64(minB) / float64(maxB)
	} else {
		result.DetectionScore = 1.0
	}

	// Match golden blocks to produced blocks and compute text accuracy + type accuracy
	matches := matchBlocks(golden, produced)
	if len(matches) > 0 {
		totalTextAcc := 0.0
		correctTypes := 0
		for _, m := range matches {
			totalTextAcc += wordOverlap(m.golden.Text, m.produced.Text)
			if strings.EqualFold(m.golden.BlockType, m.produced.BlockType) {
				correctTypes++
			}
		}
		result.TextAccuracy = totalTextAcc / float64(len(matches))
		result.TypeAccuracy = float64(correctTypes) / float64(len(matches))
	}

	// Composite quality score
	result.QualityScore = result.DetectionScore*ocrWeightDetection +
		result.TextAccuracy*ocrWeightText +
		result.TypeAccuracy*ocrWeightType

	return result
}

// ComputeOCRCompositeScores sets the CompositeScore on each result using the max cost across all results.
func ComputeOCRCompositeScores(results []OCREvalResult) {
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
		results[i].CompositeScore = results[i].QualityScore*ocrQualityWeight + costEfficiency*ocrCostWeight
	}
}

// SummarizeOCR aggregates OCR results for a single provider across all cases.
func SummarizeOCR(provider string, model string, results []OCREvalResult) OCRRunSummary {
	s := OCRRunSummary{
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
		s.AvgDetectionScore += r.DetectionScore
		s.AvgTextAccuracy += r.TextAccuracy
		s.AvgTypeAccuracy += r.TypeAccuracy
		s.AvgLatencyMs += r.LatencyMs
		s.TotalCostUSD += r.CostUSD
		s.AvgQualityScore += r.QualityScore
		s.AvgCompositeScore += r.CompositeScore
	}

	successful := len(results) - errors
	if successful > 0 {
		n := float64(successful)
		s.AvgDetectionScore /= n
		s.AvgTextAccuracy /= n
		s.AvgTypeAccuracy /= n
		s.AvgLatencyMs /= int64(successful)
		s.AvgQualityScore /= n
		s.AvgCompositeScore /= n
	}
	s.ErrorRate = float64(errors) / float64(len(results))

	return s
}

// ParseOCROutput parses the raw JSON response from a vision LLM into OCR blocks.
func ParseOCROutput(rawJSON []byte) ([]OCRBlock, error) {
	cleaned := StripMarkdownFences(rawJSON)

	// Try parsing as {"blocks": [...]}
	var wrapper struct {
		Blocks []OCRBlock `json:"blocks"`
	}
	if err := json.Unmarshal(cleaned, &wrapper); err == nil && len(wrapper.Blocks) > 0 {
		return wrapper.Blocks, nil
	}

	// Try parsing as a direct array [...]
	var blocks []OCRBlock
	if err := json.Unmarshal(cleaned, &blocks); err == nil && len(blocks) > 0 {
		return blocks, nil
	}

	return nil, json.Unmarshal(cleaned, &wrapper) // return original error
}

// --- Helpers ---

type blockMatch struct {
	golden   OCRBlock
	produced OCRBlock
}

// matchBlocks finds the best matches between golden and produced blocks using text similarity.
func matchBlocks(golden, produced []OCRBlock) []blockMatch {
	used := make(map[int]bool)
	var matches []blockMatch

	for _, g := range golden {
		bestIdx := -1
		bestScore := 0.0
		for i, p := range produced {
			if used[i] {
				continue
			}
			score := wordOverlap(g.Text, p.Text)
			if score > bestScore && score > 0.2 {
				bestScore = score
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			used[bestIdx] = true
			matches = append(matches, blockMatch{golden: g, produced: produced[bestIdx]})
		}
	}
	return matches
}

// wordOverlap computes the fraction of significant words from the golden text
// that appear in the produced text (recall-oriented).
func wordOverlap(golden, produced string) float64 {
	goldenWords := significantWords(golden)
	if len(goldenWords) == 0 {
		return 1.0
	}

	producedNorm := normalize(produced)
	producedWordSet := make(map[string]bool)
	for _, w := range strings.Fields(producedNorm) {
		producedWordSet[w] = true
	}

	found := 0
	for _, w := range goldenWords {
		// Check exact word match or substring presence
		if producedWordSet[w] || strings.Contains(producedNorm, w) {
			found++
		}
	}
	return float64(found) / float64(len(goldenWords))
}

// significantWords extracts normalized, non-stop-word tokens from text.
func significantWords(text string) []string {
	words := strings.Fields(normalize(text))
	return filterStopWords(words)
}
