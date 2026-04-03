package session

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
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

// --- Scoring edge cases ---

func TestScoreRubric_ZeroTotal_Failure(t *testing.T) {
	r := ScoreRubric(0, 0)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want %q for 0/0", r.Class, ClassFailure)
	}
	if r.Score != 0 {
		t.Errorf("score: got %f, want 0 for 0/0", r.Score)
	}
}

func TestScoreNumeric_NaN_Failure(t *testing.T) {
	r := ScoreNumeric(math.NaN(), 42, 0.1, "m", "m", false, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure for NaN expected", r.Class)
	}
}

func TestScoreNumeric_Inf_Failure(t *testing.T) {
	r := ScoreNumeric(math.Inf(1), 42, 0.1, "m", "m", false, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure for Inf expected", r.Class)
	}
}

func TestScoreNumeric_ExpectedZero_ExactMatch(t *testing.T) {
	r := ScoreNumeric(0, 0, 0.1, "m", "m", false, false)
	if r.Class != ClassSuccess {
		t.Errorf("class: got %q, want success for 0==0", r.Class)
	}
}

func TestScoreNumeric_ExpectedZero_NonZeroActual(t *testing.T) {
	r := ScoreNumeric(0, 5, 0.1, "m", "m", false, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure for 0 vs 5", r.Class)
	}
}

func TestScoreNumeric_ValueOutOfTolerance_Failure(t *testing.T) {
	// 20% off with 10% tolerance
	r := ScoreNumeric(100, 120, 0.10, "m", "m", false, false)
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure (20%% > 10%% tolerance)", r.Class)
	}
}

func TestScoreKeywords_EmptyExpected_Failure(t *testing.T) {
	r := ScoreKeywords([]string{}, []string{"anything"})
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure for empty expected", r.Class)
	}
}

func TestScoreKeywords_EmptyFound_Failure(t *testing.T) {
	r := ScoreKeywords([]string{"chloroplaste"}, []string{})
	if r.Class != ClassFailure {
		t.Errorf("class: got %q, want failure (0/1)", r.Class)
	}
	if r.Score != 0 {
		t.Errorf("score: got %f, want 0", r.Score)
	}
}

func TestScoreKeywords_WhitespaceTrimmed(t *testing.T) {
	expected := []string{"  chloroplaste  "}
	found := []string{"chloroplaste"}
	r := ScoreKeywords(expected, found)
	if r.Score != 1.0 {
		t.Errorf("score: got %f, want 1.0 (whitespace should be trimmed)", r.Score)
	}
}

// --- SessionType Value Object ---

func TestSessionType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		st    SessionType
		valid bool
	}{
		{"daily", TypeDaily, true},
		{"diagnostic", TypeDiagnostic, true},
		{"mock_exam", TypeMockExam, true},
		{"evening_first", TypeEveningFirst, true},
		{"pre_class", TypePreClass, true},
		{"invalid empty", SessionType(""), false},
		{"invalid unknown", SessionType("UNKNOWN"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.Valid(); got != tt.valid {
				t.Errorf("SessionType(%q).Valid() = %v, want %v", tt.st, got, tt.valid)
			}
		})
	}
}

func TestParseSessionType_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  SessionType
	}{
		{"daily", TypeDaily},
		{"diagnostic", TypeDiagnostic},
		{"mock_exam", TypeMockExam},
		{"evening_first", TypeEveningFirst},
		{"pre_class", TypePreClass},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSessionType(tt.input)
			if err != nil {
				t.Fatalf("ParseSessionType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseSessionType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSessionType_Invalid(t *testing.T) {
	tests := []string{"", "UNKNOWN", "Daily", "invalid"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseSessionType(input)
			if err == nil {
				t.Errorf("ParseSessionType(%q) expected error, got nil", input)
			}
		})
	}
}

// --- SessionStatus Value Object ---

func TestSessionStatus_Valid(t *testing.T) {
	tests := []struct {
		name  string
		ss    SessionStatus
		valid bool
	}{
		{"COMPOSING", StatusComposing, true},
		{"IN_PROGRESS", StatusInProgress, true},
		{"COMPLETED", StatusCompleted, true},
		{"EXPIRED", StatusExpired, true},
		{"ABANDONED", StatusAbandoned, true},
		{"invalid empty", SessionStatus(""), false},
		{"invalid unknown", SessionStatus("DRAFT"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Valid(); got != tt.valid {
				t.Errorf("SessionStatus(%q).Valid() = %v, want %v", tt.ss, got, tt.valid)
			}
		})
	}
}

// --- NewSession constructor ---

var testIDGen = event.UUIDv7Generator{}

func TestNewSession_FieldsInitialized(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	now := time.Now()

	s := NewSession(testIDGen, userID, TypeDaily, TriggerManual, now)

	if s.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if s.UserID != userID {
		t.Errorf("UserID = %v, want %v", s.UserID, userID)
	}
	if s.SessionType != TypeDaily {
		t.Errorf("SessionType = %q, want %q", s.SessionType, TypeDaily)
	}
	if s.Status != StatusComposing {
		t.Errorf("Status = %q, want %q", s.Status, StatusComposing)
	}
	if s.TriggerType != TriggerManual {
		t.Errorf("TriggerType = %q, want %q", s.TriggerType, TriggerManual)
	}
	if s.StartedAt != nil {
		t.Error("StartedAt should be nil for new session")
	}
	if s.CompletedAt != nil {
		t.Error("CompletedAt should be nil for new session")
	}
	if s.CurrentQuestionIndex != 0 {
		t.Errorf("CurrentQuestionIndex = %d, want 0", s.CurrentQuestionIndex)
	}
	if !s.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", s.CreatedAt, now)
	}
}

// --- Session state transitions ---

func TestSession_Start_FromComposing(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	now := time.Now()

	err := s.Start(now)

	if err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
	if s.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", s.Status, StatusInProgress)
	}
	if s.StartedAt == nil || !s.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v, want %v", s.StartedAt, now)
	}
}

func TestSession_Start_FromInProgress_Error(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())

	err := s.Start(time.Now())

	if err != ErrSessionNotResumable {
		t.Errorf("Start() from IN_PROGRESS should return ErrSessionNotResumable, got %v", err)
	}
}

func TestSession_Start_FromCompleted_Error(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())
	_ = s.Complete(time.Now())

	err := s.Start(time.Now())

	if err != ErrSessionNotResumable {
		t.Errorf("Start() from COMPLETED should return ErrSessionNotResumable, got %v", err)
	}
}

func TestSession_Complete_FromInProgress(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())
	now := time.Now()

	err := s.Complete(now)

	if err != nil {
		t.Fatalf("Complete() unexpected error: %v", err)
	}
	if s.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", s.Status, StatusCompleted)
	}
	if s.CompletedAt == nil || !s.CompletedAt.Equal(now) {
		t.Errorf("CompletedAt = %v, want %v", s.CompletedAt, now)
	}
}

func TestSession_Complete_FromComposing_Error(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())

	err := s.Complete(time.Now())

	if err != ErrSessionNotCompletable {
		t.Errorf("Complete() from COMPOSING should return ErrSessionNotCompletable, got %v", err)
	}
}

func TestSession_Resume_FromInProgress(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())

	err := s.Resume()

	if err != nil {
		t.Fatalf("Resume() unexpected error: %v", err)
	}
}

func TestSession_Resume_FromComposing_Error(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())

	err := s.Resume()

	if err != ErrSessionNotResumable {
		t.Errorf("Resume() from COMPOSING should return ErrSessionNotResumable, got %v", err)
	}
}

func TestSession_Resume_FromCompleted_Error(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())
	_ = s.Complete(time.Now())

	err := s.Resume()

	if err != ErrSessionNotResumable {
		t.Errorf("Resume() from COMPLETED should return ErrSessionNotResumable, got %v", err)
	}
}

func TestSession_Abandon(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	_ = s.Start(time.Now())
	now := time.Now()

	s.Abandon(now)

	if s.Status != StatusAbandoned {
		t.Errorf("Status = %q, want %q", s.Status, StatusAbandoned)
	}
	if !s.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", s.UpdatedAt, now)
	}
}

func TestSession_Abandon_FromComposing(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDaily, TriggerManual, time.Now())
	now := time.Now()

	s.Abandon(now)

	if s.Status != StatusAbandoned {
		t.Errorf("Status = %q, want %q", s.Status, StatusAbandoned)
	}
}

func TestSession_FullLifecycle_Composing_InProgress_Completed(t *testing.T) {
	s := NewSession(testIDGen, uuid.Must(uuid.NewV7()), TypeDiagnostic, TriggerScheduled, time.Now())
	if s.Status != StatusComposing {
		t.Fatalf("initial status = %q, want COMPOSING", s.Status)
	}

	if err := s.Start(time.Now()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if s.Status != StatusInProgress {
		t.Fatalf("after Start, status = %q, want IN_PROGRESS", s.Status)
	}

	if err := s.Complete(time.Now()); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if s.Status != StatusCompleted {
		t.Fatalf("after Complete, status = %q, want COMPLETED", s.Status)
	}
}

// --- NewAttempt constructor ---

func TestNewAttempt_FieldsInitialized(t *testing.T) {
	sessionID := uuid.Must(uuid.NewV7())
	questionID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	answer := []byte(`{"choice":"A"}`)
	now := time.Now()

	a := NewAttempt(testIDGen, sessionID, questionID, userID, answer, 1.0, now)

	if a.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if a.SessionID != sessionID {
		t.Errorf("SessionID = %v, want %v", a.SessionID, sessionID)
	}
	if a.QuestionID != questionID {
		t.Errorf("QuestionID = %v, want %v", a.QuestionID, questionID)
	}
	if a.UserID != userID {
		t.Errorf("UserID = %v, want %v", a.UserID, userID)
	}
	if a.Score != 1.0 {
		t.Errorf("Score = %f, want 1.0", a.Score)
	}
	if a.Source != AttemptInteractive {
		t.Errorf("Source = %q, want %q", a.Source, AttemptInteractive)
	}
	if !a.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", a.CreatedAt, now)
	}
	if a.HintUsed {
		t.Error("HintUsed should default to false")
	}
	if a.RapidResponse {
		t.Error("RapidResponse should default to false")
	}
}

// --- PackConstraints ---

func TestDefaultPackConstraints(t *testing.T) {
	pc := DefaultPackConstraints()
	if pc.MaxWritingPerSession != 1 {
		t.Errorf("MaxWritingPerSession = %d, want 1", pc.MaxWritingPerSession)
	}
	if !pc.SessionMustIncludeDoc {
		t.Error("SessionMustIncludeDoc should be true")
	}
}
