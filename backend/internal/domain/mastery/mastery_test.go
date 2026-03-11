package mastery

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// helper: create a fresh UNKNOWN mastery for testing.
func newTestMastery(t *testing.T) *Mastery {
	t.Helper()
	now := time.Date(2026, 3, 10, 18, 0, 0, 0, time.UTC)
	return &Mastery{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    uuid.Must(uuid.NewV7()),
		ItemID:    uuid.Must(uuid.NewV7()),
		State:     Unknown,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Z1-AC01 — Progression UNKNOWN → FRAGILE
// GIVEN: Un item en état UNKNOWN.
// WHEN:  L'élève répond correctement (score ≥ 0.7, sans aide).
// THEN:  État → FRAGILE, consecutive_successes = 1, next_due_at = now + 1 jour.
func TestZ1AC01_UnknownToFragile(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	err := m.RecordAttempt(0.7, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Fragile {
		t.Errorf("state: got %q, want %q", m.State, Fragile)
	}
	if m.ConsecutiveSuccesses != 1 {
		t.Errorf("consecutive_successes: got %d, want 1", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
	if m.LastReviewAt == nil || !m.LastReviewAt.Equal(now) {
		t.Errorf("last_review_at: got %v, want %v", m.LastReviewAt, now)
	}
	if m.LastSuccessAt == nil || !m.LastSuccessAt.Equal(now) {
		t.Errorf("last_success_at: got %v, want %v", m.LastSuccessAt, now)
	}
}

// Z1-AC01 — Edge case: score 0.69 is NOT a success (below 0.7 threshold).
func TestZ1AC01_ScoreBelowThreshold(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	err := m.RecordAttempt(0.69, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Unknown {
		t.Errorf("state: got %q, want %q (score 0.69 should not be success)", m.State, Unknown)
	}
	if m.ConsecutiveSuccesses != 0 {
		t.Errorf("consecutive_successes: got %d, want 0", m.ConsecutiveSuccesses)
	}
}

// Z1-AC01 — Edge case: score exactly 0.7 IS a success.
func TestZ1AC01_ScoreExactThreshold(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	err := m.RecordAttempt(0.7, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Fragile {
		t.Errorf("state: got %q, want %q (score 0.7 should be success)", m.State, Fragile)
	}
}

// Z1-AC11 — Échec sur item UNKNOWN (pas de descente sous UNKNOWN)
// GIVEN: Un item en état UNKNOWN avec consecutive_successes = 0.
// WHEN:  L'élève répond incorrectement.
// THEN:  État reste UNKNOWN, cs = 0, next_due_at = now + 1 jour.
func TestZ1AC11_UnknownFailStaysUnknown(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	err := m.RecordAttempt(0.0, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Unknown {
		t.Errorf("state: got %q, want %q", m.State, Unknown)
	}
	if m.ConsecutiveSuccesses != 0 {
		t.Errorf("consecutive_successes: got %d, want 0", m.ConsecutiveSuccesses)
	}
	if m.ConsecutiveFailures != 1 {
		t.Errorf("consecutive_failures: got %d, want 1", m.ConsecutiveFailures)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
	if m.LastReviewAt == nil || !m.LastReviewAt.Equal(now) {
		t.Errorf("last_review_at: got %v, want %v", m.LastReviewAt, now)
	}
	if m.LastSuccessAt != nil {
		t.Errorf("last_success_at: got %v, want nil (failure should not set it)", m.LastSuccessAt)
	}
}
