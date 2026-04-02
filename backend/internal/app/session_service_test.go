package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
)

// --- Session mocks ---

type mockSessionRepo struct {
	sessions  map[uuid.UUID]*session.Session
	questions map[uuid.UUID]*session.Question
	attempts  map[uuid.UUID]*session.Attempt
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:  make(map[uuid.UUID]*session.Session),
		questions: make(map[uuid.UUID]*session.Question),
		attempts:  make(map[uuid.UUID]*session.Attempt),
	}
}

func (m *mockSessionRepo) FindByID(_ context.Context, id uuid.UUID) (*session.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, session.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) FindActiveByUser(_ context.Context, userID uuid.UUID) (*session.Session, error) {
	for _, s := range m.sessions {
		if s.UserID == userID && s.Status == session.StatusInProgress {
			return s, nil
		}
	}
	return nil, session.ErrNotFound
}

func (m *mockSessionRepo) Save(_ context.Context, s *session.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) FindQuestionByID(_ context.Context, id uuid.UUID) (*session.Question, error) {
	q, ok := m.questions[id]
	if !ok {
		return nil, session.ErrNotFound
	}
	return q, nil
}

func (m *mockSessionRepo) FindQuestionsBySession(_ context.Context, sessionID uuid.UUID) ([]*session.Question, error) {
	var result []*session.Question
	for _, q := range m.questions {
		if q.SessionID == sessionID {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *mockSessionRepo) SaveQuestion(_ context.Context, q *session.Question) error {
	m.questions[q.ID] = q
	return nil
}

func (m *mockSessionRepo) SaveAttempt(_ context.Context, a *session.Attempt) error {
	m.attempts[a.ID] = a
	return nil
}

func (m *mockSessionRepo) FindAttemptsBySession(_ context.Context, sessionID uuid.UUID) ([]*session.Attempt, error) {
	var result []*session.Attempt
	for _, a := range m.attempts {
		if a.SessionID == sessionID {
			result = append(result, a)
		}
	}
	return result, nil
}

// --- Tests ---

// Z6-AC04: evening_first session proposed after pipeline
func TestZ6AC04_EveningFirstProposedAfterPipeline(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	chRepo.Save(context.Background(), ch)

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	// Simulate: pipeline produced 12 items, all UNKNOWN
	var itemIDs []uuid.UUID
	for i := 0; i < 12; i++ {
		itemID := uuid.Must(uuid.NewV7())
		itemIDs = append(itemIDs, itemID)
		m := mastery.NewMastery(idGen, ch.UserID, itemID, now)
		masteryRepo.saved = append(masteryRepo.saved, m)
	}

	sess, err := svc.ProposeEveningFirst(context.Background(), ch.UserID, ch.ID, itemIDs)
	if err != nil {
		t.Fatalf("ProposeEveningFirst: %v", err)
	}

	if sess.SessionType != session.TypeEveningFirst {
		t.Errorf("session type = %s, want evening_first", sess.SessionType)
	}
	if sess.TriggerType != session.TriggerScheduled {
		t.Errorf("trigger = %s, want scheduled", sess.TriggerType)
	}
	if sess.Status != session.StatusComposing {
		t.Errorf("status = %s, want COMPOSING", sess.Status)
	}
}

// Z6-AC05: evening_first = 100% UNKNOWN items, difficulty 1 templates
func TestZ6AC05_EveningFirstComposition(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	chRepo.Save(context.Background(), ch)

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	// 12 UNKNOWN items
	var itemIDs []uuid.UUID
	for i := 0; i < 12; i++ {
		itemID := uuid.Must(uuid.NewV7())
		itemIDs = append(itemIDs, itemID)
		m := mastery.NewMastery(idGen, ch.UserID, itemID, now)
		masteryRepo.saved = append(masteryRepo.saved, m)
	}

	sess, err := svc.ProposeEveningFirst(context.Background(), ch.UserID, ch.ID, itemIDs)
	if err != nil {
		t.Fatalf("ProposeEveningFirst: %v", err)
	}

	// Should have 6-10 questions (calibrated for 5-10 min)
	questions, _ := sessionRepo.FindQuestionsBySession(context.Background(), sess.ID)
	if len(questions) < 6 || len(questions) > 10 {
		t.Errorf("expected 6-10 questions, got %d", len(questions))
	}

	// All questions should use difficulty 1 templates
	for _, q := range questions {
		switch q.TemplateID {
		case "GEN.KNOW.FLASH_MCQ", "GEN.KNOW.DEF_SHORT", "GEN.KNOW.CLOZE_KEYWORDS":
			// OK — difficulty 1
		default:
			t.Errorf("unexpected template %q (should be difficulty 1 only)", q.TemplateID)
		}
	}
}

// Z4-AC04: resume interrupted session from last question
func TestZ4AC04_ResumeInterruptedSession(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	// Create a session with 10 questions, student answered 4
	userID := uuid.Must(uuid.NewV7())
	sess := session.NewSession(idGen, userID, session.TypeDaily, session.TriggerManual, now)
	sess.Start(now)
	sess.CurrentQuestionIndex = 4
	sessionRepo.Save(context.Background(), sess)

	// Resume
	resumed, err := svc.ResumeSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("ResumeSession: %v", err)
	}
	if resumed.CurrentQuestionIndex != 4 {
		t.Errorf("CurrentQuestionIndex = %d, want 4", resumed.CurrentQuestionIndex)
	}
	if resumed.Status != session.StatusInProgress {
		t.Errorf("status = %s, want IN_PROGRESS", resumed.Status)
	}
}

// Z4-AC04: cannot resume completed session
func TestZ4AC04_CannotResumeCompletedSession(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	userID := uuid.Must(uuid.NewV7())
	sess := session.NewSession(idGen, userID, session.TypeDaily, session.TriggerManual, now)
	sess.Start(now)
	sess.Complete(now)
	sessionRepo.Save(context.Background(), sess)

	_, err := svc.ResumeSession(context.Background(), sess.ID)
	if err == nil {
		t.Fatal("expected error resuming completed session")
	}
}

// Z4-AC06: pool empty → graceful error, no empty session created
func TestZ4AC06_PoolEmptyGracefulDegradation(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{} // no masteries = empty pool
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	chRepo.Save(context.Background(), ch)

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	_, err := svc.ComposeDaily(context.Background(), ch.UserID, ch.ID)
	if err == nil {
		t.Fatal("expected error for empty pool")
	}

	// No session should be created
	if len(sessionRepo.sessions) != 0 {
		t.Errorf("expected 0 sessions created, got %d", len(sessionRepo.sessions))
	}
}

// Z6-AC11: degraded mode without schedule — sessions still available
func TestZ6AC11_DegradedModeWithoutSchedule(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	chRepo.Save(context.Background(), ch)

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	// Without schedule, daily/evening_first/mock_exam should still be available
	types := svc.AvailableSessionTypes(false) // hasSchedule = false
	if !contains(types, session.TypeDaily) {
		t.Error("daily should be available without schedule")
	}
	if !contains(types, session.TypeEveningFirst) {
		t.Error("evening_first should be available without schedule")
	}
	if !contains(types, session.TypeMockExam) {
		t.Error("mock_exam should be available without schedule")
	}
	if contains(types, session.TypePreClass) {
		t.Error("pre_class should NOT be available without schedule")
	}
}

func contains(types []session.SessionType, target session.SessionType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

// Z4-AC05: pack constraints — max 1 writing, must include 1 document
func TestZ4AC05_PackConstraints(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	sessionRepo := newMockSessionRepo()
	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	userID := uuid.Must(uuid.NewV7())
	ch := chapter.NewChapter(idGen, userID, "HG", "4e", "Inegalites", now)
	chRepo.Save(context.Background(), ch)

	// Create 5 document items + 8 knowledge items + 3 writing items
	var docItemIDs, knowledgeItemIDs, writingItemIDs []uuid.UUID

	for i := 0; i < 5; i++ {
		item := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, ItemType: chapter.ItemDocument, CreatedAt: now, UpdatedAt: now}
		chRepo.SaveItem(context.Background(), item)
		docItemIDs = append(docItemIDs, item.ID)
	}
	for i := 0; i < 8; i++ {
		item := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now}
		chRepo.SaveItem(context.Background(), item)
		knowledgeItemIDs = append(knowledgeItemIDs, item.ID)
	}
	for i := 0; i < 3; i++ {
		item := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, ItemType: chapter.ItemWriting, CreatedAt: now, UpdatedAt: now}
		chRepo.SaveItem(context.Background(), item)
		writingItemIDs = append(writingItemIDs, item.ID)
	}

	// All items are due
	allItemIDs := append(append(docItemIDs, knowledgeItemIDs...), writingItemIDs...)
	for _, itemID := range allItemIDs {
		m := mastery.NewMastery(idGen, userID, itemID, now)
		// Make them due by setting NextDueAt in the past
		past := now.Add(-1 * time.Hour)
		m.NextDueAt = &past
		masteryRepo.saved = append(masteryRepo.saved, m)
	}

	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen, nil)

	sess, err := svc.ComposeDaily(context.Background(), userID, ch.ID)
	if err != nil {
		t.Fatalf("ComposeDaily: %v", err)
	}

	questions, _ := sessionRepo.FindQuestionsBySession(context.Background(), sess.ID)

	// Count item types in session
	docCount := 0
	writingCount := 0
	docSet := make(map[uuid.UUID]bool)
	for _, id := range docItemIDs {
		docSet[id] = true
	}
	writingSet := make(map[uuid.UUID]bool)
	for _, id := range writingItemIDs {
		writingSet[id] = true
	}

	for _, q := range questions {
		if docSet[q.ItemID] {
			docCount++
		}
		if writingSet[q.ItemID] {
			writingCount++
		}
	}

	// Z4-AC05: must include at least 1 document
	if docCount < 1 {
		t.Errorf("expected at least 1 document question, got %d", docCount)
	}

	// Z4-AC05: max 1 writing per session
	if writingCount > 1 {
		t.Errorf("expected at most 1 writing question, got %d", writingCount)
	}

	// Total should be ≤ 10
	if len(questions) > 10 {
		t.Errorf("expected at most 10 questions, got %d", len(questions))
	}
}

// Z4-AC09: feedback contains correct answer, what was missing, hint
func TestZ4AC09_EnrichedFeedback(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		expectedAnswer string
		studentAnswer  string
		wantCorrect    bool
		wantMissing    bool
		wantHint       bool
	}{
		{
			name:           "KEYWORDS missing terms",
			templateID:     "GEN.KNOW.CLOZE_KEYWORDS",
			expectedAnswer: `{"keywords":["chloroplaste","lumière","CO2"]}`,
			studentAnswer:  `{"answer":"les plantes font de la nourriture"}`,
			wantCorrect:    true,
			wantMissing:    true,
			wantHint:       true,
		},
		{
			name:           "MCQ wrong distractor",
			templateID:     "GEN.KNOW.FLASH_MCQ",
			expectedAnswer: `{"correct":"B","explanation":"La photosynthèse produit du glucose"}`,
			studentAnswer:  `{"answer":"A"}`,
			wantCorrect:    true,
			wantMissing:    true,
			wantHint:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := GenerateFeedback(tt.templateID, []byte(tt.expectedAnswer), []byte(tt.studentAnswer))

			if fb.CorrectAnswer == "" && tt.wantCorrect {
				t.Error("expected CorrectAnswer to be non-empty")
			}
			if fb.WhatWasMissing == "" && tt.wantMissing {
				t.Error("expected WhatWasMissing to be non-empty")
			}
			if fb.Hint == "" && tt.wantHint {
				t.Error("expected Hint to be non-empty")
			}
		})
	}
}
