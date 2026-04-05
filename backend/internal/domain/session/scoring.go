package session

import (
	"math"
	"strings"
)

// ScoreResult holds the computed score and classification.
type ScoreResult struct {
	Score float64 // 0.0 to 1.0
	Class ScoreClass
}

// ScoreClass categorizes a score.
type ScoreClass string

const (
	ClassSuccess     ScoreClass = "success"      // ≥ threshold
	ClassHalfSuccess ScoreClass = "half_success" // between half and full threshold
	ClassFailure     ScoreClass = "failure"      // below half threshold
)

// ScoreRubric computes a score for a RUBRIC question (Z1-AC10).
// validatedCriteria / totalCriteria determines the score.
// Success: ≥ 75%, Half-success: 50-74%, Failure: < 50%.
func ScoreRubric(validatedCriteria, totalCriteria int) ScoreResult {
	if totalCriteria == 0 {
		return ScoreResult{Score: 0, Class: ClassFailure}
	}
	score := float64(validatedCriteria) / float64(totalCriteria)
	// Floor to percentage: ⌊(validated/total) × 100⌋ / 100
	score = math.Floor(score*100) / 100

	var class ScoreClass
	switch {
	case score >= 0.75:
		class = ClassSuccess
	case score >= 0.50:
		class = ClassHalfSuccess
	default:
		class = ClassFailure
	}
	return ScoreResult{Score: score, Class: class}
}

// ScoreNumeric computes a score for a NUMERIC question (Z1-AC10).
// Checks value correctness (within tolerance), unit correctness, and optional formula.
func ScoreNumeric(expected, actual float64, tolerance float64, expectedUnit, actualUnit string, formulaRequired bool, formulaProvided bool) ScoreResult {
	// Edge case: expected value 0 or undefined → always failure
	if math.IsNaN(expected) || math.IsInf(expected, 0) {
		return ScoreResult{Score: 0, Class: ClassFailure}
	}

	valueCorrect := false
	if expected == 0 {
		valueCorrect = actual == 0
	} else {
		valueCorrect = math.Abs(actual-expected)/math.Abs(expected) <= tolerance
	}
	valueExact := actual == expected
	unitCorrect := strings.EqualFold(strings.TrimSpace(expectedUnit), strings.TrimSpace(actualUnit))

	// Formula check: if required and not provided → failure regardless
	if formulaRequired && !formulaProvided {
		return ScoreResult{Score: 0, Class: ClassFailure}
	}

	switch {
	case valueExact && unitCorrect:
		return ScoreResult{Score: 1.0, Class: ClassSuccess}
	case valueCorrect && unitCorrect:
		// Close value (within tolerance) + correct unit → half-success
		return ScoreResult{Score: 0.5, Class: ClassHalfSuccess}
	case valueCorrect && !unitCorrect:
		// Value correct but unit wrong/missing → failure (faux négatif)
		return ScoreResult{Score: 0, Class: ClassFailure}
	default:
		return ScoreResult{Score: 0, Class: ClassFailure}
	}
}

// ScoreKeywords computes a score for a KEYWORDS question (Z1-AC10).
// Score = found/total. Success: ≥ 85%, Half-success: 50-84%, Failure: < 50%.
func ScoreKeywords(expectedKeywords []string, foundKeywords []string) ScoreResult {
	if len(expectedKeywords) == 0 {
		return ScoreResult{Score: 0, Class: ClassFailure}
	}

	expectedSet := make(map[string]bool)
	for _, kw := range expectedKeywords {
		expectedSet[strings.ToLower(strings.TrimSpace(kw))] = true
	}

	found := 0
	for _, kw := range foundKeywords {
		if expectedSet[strings.ToLower(strings.TrimSpace(kw))] {
			found++
		}
	}

	score := float64(found) / float64(len(expectedKeywords))

	var class ScoreClass
	switch {
	case score >= 0.85:
		class = ClassSuccess
	case score >= 0.50:
		class = ClassHalfSuccess
	default:
		class = ClassFailure
	}
	return ScoreResult{Score: score, Class: class}
}

// ScoreBinary computes a binary score for MCQ, CLOZE, SHORT_ANSWER (Z1-AC10).
// Success = 1.0, Failure = 0.0.
func ScoreBinary(correct bool) ScoreResult {
	if correct {
		return ScoreResult{Score: 1.0, Class: ClassSuccess}
	}
	return ScoreResult{Score: 0, Class: ClassFailure}
}
