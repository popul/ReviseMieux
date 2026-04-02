package session

import (
	"testing"
)

// Z1-AC10 — RUBRIC scoring

func TestZ1AC10_Rubric_3of4_Success(t *testing.T) {
	r := ScoreRubric(3, 4) // 75% = success
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
	if r.Score != 0.75 {
		t.Errorf("score: got %f, want 0.75", r.Score)
	}
}

func TestZ1AC10_Rubric_2of4_HalfSuccess(t *testing.T) {
	r := ScoreRubric(2, 4) // 50% = half-success
	if r.Class != ClassHalfSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassHalfSuccess)
	}
}

func TestZ1AC10_Rubric_1of4_Failure(t *testing.T) {
	r := ScoreRubric(1, 4) // 25% = failure
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want %q", r.Class, ClassFailure)
	}
}

func TestZ1AC10_Rubric_4of4_Success(t *testing.T) {
	r := ScoreRubric(4, 4)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
	if r.Score != 1.0 {
		t.Errorf("score: got %f, want 1.0", r.Score)
	}
}

// Z1-AC10 — NUMERIC scoring

func TestZ1AC10_Numeric_ExactValueCorrectUnit_Success(t *testing.T) {
	r := ScoreNumeric(1000, 1000, 0.10, "kg", "kg", false, false)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
}

func TestZ1AC10_Numeric_CorrectValueMissingUnit_Failure(t *testing.T) {
	// Faux négatif: value correct but unit absent
	r := ScoreNumeric(1000, 1000, 0.10, "kg", "", false, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want %q (faux négatif: unit missing)", r.Class, ClassFailure)
	}
}

func TestZ1AC10_Numeric_CloseValueCorrectUnit_HalfSuccess(t *testing.T) {
	r := ScoreNumeric(1000, 1050, 0.10, "kg", "kg", false, false) // 5% off, within 10% tolerance
	if r.Class != ClassHalfSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassHalfSuccess)
	}
}

func TestZ1AC10_Numeric_FormulaRequiredNotProvided_Failure(t *testing.T) {
	r := ScoreNumeric(1000, 1000, 0.10, "kg", "kg", true, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want %q (formula required but not provided)", r.Class, ClassFailure)
	}
}

func TestZ1AC10_Numeric_FormulaRequiredAndProvided_Success(t *testing.T) {
	r := ScoreNumeric(1000, 1000, 0.10, "kg", "kg", true, true)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
}

// Z1-AC10 — KEYWORDS scoring

func TestZ1AC10_Keywords_6of6_Success(t *testing.T) {
	expected := []string{"chloroplaste", "lumière", "CO2", "eau", "glucose", "oxygène"}
	found := []string{"chloroplaste", "lumière", "CO2", "eau", "glucose", "oxygène"}
	r := ScoreKeywords(expected, found)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
	if r.Score != 1.0 {
		t.Errorf("score: got %f, want 1.0", r.Score)
	}
}

func TestZ1AC10_Keywords_5of6_HalfSuccess(t *testing.T) {
	expected := []string{"chloroplaste", "lumière", "CO2", "eau", "glucose", "oxygène"}
	found := []string{"chloroplaste", "lumière", "CO2", "eau", "glucose"} // 83.3%
	r := ScoreKeywords(expected, found)
	if r.Class != ClassHalfSuccess {
		t.Errorf("class: got %q, want %q (83%% < 85%%)", r.Class, ClassHalfSuccess)
	}
}

func TestZ1AC10_Keywords_2of3_Failure(t *testing.T) {
	expected := []string{"chloroplaste", "lumière", "CO2"}
	found := []string{"chloroplaste", "lumière"} // 66.7%
	r := ScoreKeywords(expected, found)
	if r.Class != ClassHalfSuccess {
		t.Errorf("class: got %q, want %q (66%% is half-success)", r.Class, ClassHalfSuccess)
	}
}

func TestZ1AC10_Keywords_1of3_Failure(t *testing.T) {
	expected := []string{"chloroplaste", "lumière", "CO2"}
	found := []string{"chloroplaste"} // 33.3%
	r := ScoreKeywords(expected, found)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want %q (33%% < 50%%)", r.Class, ClassFailure)
	}
}

func TestZ1AC10_Keywords_CaseInsensitive(t *testing.T) {
	expected := []string{"Chloroplaste", "Lumière"}
	found := []string{"chloroplaste", "lumière"}
	r := ScoreKeywords(expected, found)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want %q", r.Class, ClassSuccess)
	}
}

// Z1-AC10 — Binary scoring (MCQ, CLOZE, SHORT_ANSWER)

func TestZ1AC10_Binary_Correct(t *testing.T) {
	r := ScoreBinary(true)
	if r.Score != 1.0 || r.Class != ClassSuccess {
		t.Errorf("got score=%f class=%q, want 1.0/success", r.Score, r.Class)
	}
}

func TestZ1AC10_Binary_Incorrect(t *testing.T) {
	r := ScoreBinary(false)
	if r.Score != 0.0 || r.Class != ClassFailure {
		t.Errorf("got score=%f class=%q, want 0.0/failure", r.Score, r.Class)
	}
}
