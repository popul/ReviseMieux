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

// Z1-AC02 — Progression FRAGILE → OK
// GIVEN: Un item en état FRAGILE avec consecutive_successes ≥ 1.
// WHEN:  L'élève répond correctement (score ≥ 0.7, sans aide).
// THEN:  État → OK, cs = 2, next_due_at = now + 3 jours.
func TestZ1AC02_FragileToOK(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	// First: UNKNOWN → FRAGILE (AC01)
	m.RecordAttempt(1.0, now)
	if m.State != Fragile {
		t.Fatalf("setup: expected FRAGILE after first success, got %q", m.State)
	}
	if m.ConsecutiveSuccesses != 1 {
		t.Fatalf("setup: expected cs=1, got %d", m.ConsecutiveSuccesses)
	}

	// Now: FRAGILE(cs=1) + success → OK
	now2 := now.Add(2 * time.Hour)
	err := m.RecordAttempt(0.8, now2)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q", m.State, OK)
	}
	if m.ConsecutiveSuccesses != 2 {
		t.Errorf("consecutive_successes: got %d, want 2", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now2.Add(3 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}

// Z1-AC07b — Récupération FRAGILE après régression (cs=0, réponse correcte)
// GIVEN: Un item en état FRAGILE avec cs=0 (suite à régression).
// WHEN:  L'élève répond correctement (score ≥ 0.7).
// THEN:  cs = 1, état reste FRAGILE, next_due_at = now + 1 jour.
func TestZ1AC07b_FragileRecoveryFromCS0(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	// Setup: put into FRAGILE with cs=0 (simulates post-regression state)
	m.State = Fragile
	m.ConsecutiveSuccesses = 0

	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Fragile {
		t.Errorf("state: got %q, want %q (should stay FRAGILE when cs was 0)", m.State, Fragile)
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
}

// Z1-AC07 — Régression FRAGILE sur échec (pas de descente sous FRAGILE)
// GIVEN: Un item en état FRAGILE.
// WHEN:  L'élève répond incorrectement.
// THEN:  État reste FRAGILE, cs = 0, next_due_at = now + 1 jour. Pas de retour à UNKNOWN.
func TestZ1AC07_FragileFailStaysFragile(t *testing.T) {
	m := newTestMastery(t)
	now := time.Date(2026, 3, 11, 18, 0, 0, 0, time.UTC)

	// Setup: FRAGILE with cs=1
	m.State = Fragile
	m.ConsecutiveSuccesses = 1

	err := m.RecordAttempt(0.2, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Fragile {
		t.Errorf("state: got %q, want %q (should not descend below FRAGILE)", m.State, Fragile)
	}
	if m.ConsecutiveSuccesses != 0 {
		t.Errorf("consecutive_successes: got %d, want 0", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDueAC07 := now.Add(24 * time.Hour)
	if !m.NextDueAt.Equal(wantDueAC07) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDueAC07)
	}
}

// Z1-AC06 — Régression OK → FRAGILE sur échec
// GIVEN: Un item en état OK.
// WHEN:  L'élève répond incorrectement.
// THEN:  État → FRAGILE, cs = 0, next_due_at = now + 1 jour.
func TestZ1AC06_OKToFragile(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK state with cs=2
	m.State = OK
	m.ConsecutiveSuccesses = 2

	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	err := m.RecordAttempt(0.3, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Fragile {
		t.Errorf("state: got %q, want %q", m.State, Fragile)
	}
	if m.ConsecutiveSuccesses != 0 {
		t.Errorf("consecutive_successes: got %d, want 0", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}

// Z1-AC05 — Régression SOLID → OK sur échec unique
// GIVEN: Un item en état SOLID avec cs ≥ 3.
// WHEN:  L'élève répond incorrectement.
// THEN:  État → OK (pas FRAGILE), cs = 0, next_due_at = now + 2 jours.
func TestZ1AC05_SolidToOK(t *testing.T) {
	m := newTestMastery(t)

	// Setup: SOLID state with cs=3
	m.State = Solid
	m.ConsecutiveSuccesses = 3

	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	err := m.RecordAttempt(0.2, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q (SOLID regresses to OK, not FRAGILE)", m.State, OK)
	}
	if m.ConsecutiveSuccesses != 0 {
		t.Errorf("consecutive_successes: got %d, want 0", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDueAC05 := now.Add(2 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDueAC05) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDueAC05)
	}
}

// Z1-AC07c — Récupération OK après régression (cs<2, réponse correcte)
// GIVEN: Un item en état OK avec cs = 0 (suite à régression SOLID→OK).
// WHEN:  L'élève répond correctement (score ≥ 0.7).
// THEN:  cs += 1, état reste OK, next_due_at = now + 3 jours.
func TestZ1AC07c_OKRecoveryFromCS0(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK with cs=0 (post-regression from SOLID→OK)
	m.State = OK
	m.ConsecutiveSuccesses = 0

	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q (should stay OK when cs < 2)", m.State, OK)
	}
	if m.ConsecutiveSuccesses != 1 {
		t.Errorf("consecutive_successes: got %d, want 1", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(3 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}

// Z1-AC07c + Z1-AC04 — Recovery path: OK(cs=1) + success with prior success 25h ago → cs=2, stays OK
func TestZ1AC07c_OKRecoveryPathCS1ToCS2(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK with cs=1, last success 25h ago (spacing met for accumulation)
	prevSuccess := time.Date(2026, 3, 11, 17, 0, 0, 0, time.UTC)
	m.State = OK
	m.ConsecutiveSuccesses = 1
	m.LastSuccessAt = &prevSuccess

	now := prevSuccess.Add(25 * time.Hour)
	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	// cs reaches 2 and spacing is met → transitions to SOLID (AC03)
	if m.State != Solid {
		t.Errorf("state: got %q, want %q", m.State, Solid)
	}
	if m.ConsecutiveSuccesses != 2 {
		t.Errorf("consecutive_successes: got %d, want 2", m.ConsecutiveSuccesses)
	}
}

// Z1-AC04 — OK(cs=1) + success without spacing → cs blocked at 1
func TestZ1AC04_OKRecoveryCS1BlockedBySpacing(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK with cs=1, last success 2h ago (spacing NOT met)
	recentSuccess := time.Date(2026, 3, 12, 16, 0, 0, 0, time.UTC)
	originalDue := time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)
	m.State = OK
	m.ConsecutiveSuccesses = 1
	m.LastSuccessAt = &recentSuccess
	m.NextDueAt = &originalDue

	now := recentSuccess.Add(2 * time.Hour)
	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q", m.State, OK)
	}
	// cs would reach 2, but spacing blocks → reverted to 1
	if m.ConsecutiveSuccesses != 1 {
		t.Errorf("consecutive_successes: got %d, want 1 (blocked by spacing)", m.ConsecutiveSuccesses)
	}
}

// Z1-AC12 — Maintien SOLID sur réussite successive
// GIVEN: Un item en état SOLID avec cs ≥ 3.
// WHEN:  L'élève répond correctement.
// THEN:  État reste SOLID, cs += 1, next_due_at = now + 7 jours.
func TestZ1AC12_SolidStaysSolid(t *testing.T) {
	m := newTestMastery(t)

	// Setup: SOLID with cs=3
	m.State = Solid
	m.ConsecutiveSuccesses = 3

	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	err := m.RecordAttempt(0.9, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Solid {
		t.Errorf("state: got %q, want %q", m.State, Solid)
	}
	if m.ConsecutiveSuccesses != 4 {
		t.Errorf("consecutive_successes: got %d, want 4", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(7 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}

// Z1-AC09 — Indépendance des Mastery states entre items
// GIVEN: Deux items A et B dans le même chapitre, A en SOLID, B en UNKNOWN.
// WHEN:  L'élève échoue sur B.
// THEN:  Le Mastery state de A n'est pas modifié.
func TestZ1AC09_MasteryIndependence(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)

	// Item A: SOLID
	mA := newTestMastery(t)
	mA.State = Solid
	mA.ConsecutiveSuccesses = 3

	// Item B: UNKNOWN
	mB := newTestMastery(t)

	// Fail on B
	mB.RecordAttempt(0.0, now)

	// A should be untouched
	if mA.State != Solid {
		t.Errorf("item A state: got %q, want %q (should be independent)", mA.State, Solid)
	}
	if mA.ConsecutiveSuccesses != 3 {
		t.Errorf("item A cs: got %d, want 3 (should be independent)", mA.ConsecutiveSuccesses)
	}
}

// Z1-AC03 — Progression OK → SOLID (espacement requis)
// GIVEN: Un item en état OK avec cs ≥ 2 et last_success_at ≥ 24h avant la tentative.
// WHEN:  L'élève répond correctement (score ≥ 0.7, sans aide).
// THEN:  État → SOLID, cs += 1, next_due_at = now + 7 jours.
func TestZ1AC03_OKToSolidWithSpacing(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK with cs=2, last success was 25h ago
	day1 := time.Date(2026, 3, 10, 18, 0, 0, 0, time.UTC)
	m.State = OK
	m.ConsecutiveSuccesses = 2
	m.LastSuccessAt = &day1

	// Attempt 25h later (spacing met)
	now := day1.Add(25 * time.Hour)
	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != Solid {
		t.Errorf("state: got %q, want %q", m.State, Solid)
	}
	if m.ConsecutiveSuccesses != 3 {
		t.Errorf("consecutive_successes: got %d, want 3", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(7 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}

// Z1-AC04 — Blocage OK → SOLID sans espacement
// GIVEN: Un item en état OK avec last_success_at < 24h.
// WHEN:  L'élève répond correctement dans la même session.
// THEN:  État reste OK. cs N'EST PAS incrémenté. next_due_at N'EST PAS modifié.
func TestZ1AC04_OKBlockedWithoutSpacing(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK with cs=2, last success was 2h ago (spacing NOT met)
	recentSuccess := time.Date(2026, 3, 11, 16, 0, 0, 0, time.UTC)
	originalDue := time.Date(2026, 3, 14, 18, 0, 0, 0, time.UTC)
	m.State = OK
	m.ConsecutiveSuccesses = 2
	m.LastSuccessAt = &recentSuccess
	m.NextDueAt = &originalDue

	// Attempt 2h later (spacing NOT met — only 2h, not 24h)
	now := recentSuccess.Add(2 * time.Hour)
	err := m.RecordAttempt(0.8, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q (should stay OK without spacing)", m.State, OK)
	}
	if m.ConsecutiveSuccesses != 2 {
		t.Errorf("consecutive_successes: got %d, want 2 (should NOT increment without spacing)", m.ConsecutiveSuccesses)
	}
	if !m.NextDueAt.Equal(originalDue) {
		t.Errorf("next_due_at: got %v, want %v (should NOT change without spacing)", *m.NextDueAt, originalDue)
	}
}

// Z1-AC13 — Plafond maîtrise OK pour items validation_required
// GIVEN: Un item avec CappedAtOK=true en état OK avec cs=2.
// WHEN:  L'élève répond correctement (gabarit simple).
// THEN:  État reste OK (plafonné), cs s'incrémente, pas de transition vers SOLID.
func TestZ1AC13_CappedAtOKBlocksSolid(t *testing.T) {
	m := newTestMastery(t)

	// Setup: OK, capped, cs=2, spacing would be met
	day1 := time.Date(2026, 3, 10, 18, 0, 0, 0, time.UTC)
	m.State = OK
	m.ConsecutiveSuccesses = 2
	m.CappedAtOK = true
	m.LastSuccessAt = &day1

	now := day1.Add(25 * time.Hour)
	err := m.RecordAttempt(0.9, now)
	if err != nil {
		t.Fatalf("RecordAttempt returned unexpected error: %v", err)
	}

	if m.State != OK {
		t.Errorf("state: got %q, want %q (CappedAtOK should block SOLID)", m.State, OK)
	}
	// cs still increments normally per AC spec
	if m.ConsecutiveSuccesses != 3 {
		t.Errorf("consecutive_successes: got %d, want 3", m.ConsecutiveSuccesses)
	}
	if m.NextDueAt == nil {
		t.Fatal("next_due_at: got nil, want non-nil")
	}
	wantDue := now.Add(3 * 24 * time.Hour)
	if !m.NextDueAt.Equal(wantDue) {
		t.Errorf("next_due_at: got %v, want %v", *m.NextDueAt, wantDue)
	}
}
