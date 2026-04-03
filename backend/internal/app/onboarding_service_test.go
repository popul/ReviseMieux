package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// --- mocks for onboarding ---

type mockChapterRepoOnboarding struct {
	chapters []*chapter.Chapter
	items    []*chapter.Item
	notions  []*chapter.Notion
	revision *chapter.Revision
}

func (m *mockChapterRepoOnboarding) FindByID(_ context.Context, id uuid.UUID) (*chapter.Chapter, error) {
	for _, ch := range m.chapters {
		if ch.ID == id {
			return ch, nil
		}
	}
	return nil, chapter.ErrNotFound
}

func (m *mockChapterRepoOnboarding) FindByUser(_ context.Context, _ uuid.UUID, _ bool) ([]*chapter.Chapter, error) {
	return m.chapters, nil
}

func (m *mockChapterRepoOnboarding) Save(_ context.Context, ch *chapter.Chapter) error {
	for i, existing := range m.chapters {
		if existing.ID == ch.ID {
			m.chapters[i] = ch
			return nil
		}
	}
	m.chapters = append(m.chapters, ch)
	return nil
}

func (m *mockChapterRepoOnboarding) FindItemByID(_ context.Context, id uuid.UUID) (*chapter.Item, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, chapter.ErrNotFound
}

func (m *mockChapterRepoOnboarding) FindItemsByChapter(_ context.Context, chapterID uuid.UUID, _ bool) ([]*chapter.Item, error) {
	var result []*chapter.Item
	for _, item := range m.items {
		if item.ChapterID == chapterID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockChapterRepoOnboarding) SaveItem(_ context.Context, item *chapter.Item) error {
	m.items = append(m.items, item)
	return nil
}

func (m *mockChapterRepoOnboarding) SaveItems(_ context.Context, items []*chapter.Item) error {
	m.items = append(m.items, items...)
	return nil
}

func (m *mockChapterRepoOnboarding) FindNotionsByChapter(_ context.Context, _ uuid.UUID) ([]*chapter.Notion, error) {
	return m.notions, nil
}

func (m *mockChapterRepoOnboarding) SaveNotion(_ context.Context, n *chapter.Notion) error {
	m.notions = append(m.notions, n)
	return nil
}

func (m *mockChapterRepoOnboarding) FindRevisionByID(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return m.revision, nil
}

func (m *mockChapterRepoOnboarding) FindCurrentRevision(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return m.revision, nil
}

func (m *mockChapterRepoOnboarding) SaveRevision(_ context.Context, r *chapter.Revision) error {
	m.revision = r
	return nil
}

func (m *mockChapterRepoOnboarding) FindPagesByRevision(_ context.Context, _ uuid.UUID) ([]*chapter.Page, error) {
	return nil, nil
}

func (m *mockChapterRepoOnboarding) SavePage(_ context.Context, _ *chapter.Page) error {
	return nil
}

func (m *mockChapterRepoOnboarding) CountRevisionsByChapter(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

type mockMasteryRepoOnboarding struct {
	masteries []*mastery.Mastery
}

func (m *mockMasteryRepoOnboarding) FindByID(_ context.Context, id uuid.UUID) (*mastery.Mastery, error) {
	for _, ms := range m.masteries {
		if ms.ID == id {
			return ms, nil
		}
	}
	return nil, mastery.ErrNotFound
}

func (m *mockMasteryRepoOnboarding) FindByUserAndItem(_ context.Context, userID, itemID uuid.UUID) (*mastery.Mastery, error) {
	for _, ms := range m.masteries {
		if ms.UserID == userID && ms.ItemID == itemID {
			return ms, nil
		}
	}
	return nil, mastery.ErrNotFound
}

func (m *mockMasteryRepoOnboarding) FindDueByUser(_ context.Context, _ uuid.UUID, _ time.Time) ([]*mastery.Mastery, error) {
	return nil, nil
}

func (m *mockMasteryRepoOnboarding) FindByUserAndState(_ context.Context, _ uuid.UUID, _ mastery.State) ([]*mastery.Mastery, error) {
	return nil, nil
}

func (m *mockMasteryRepoOnboarding) Save(_ context.Context, ms *mastery.Mastery) error {
	m.masteries = append(m.masteries, ms)
	return nil
}

func (m *mockMasteryRepoOnboarding) SaveAll(_ context.Context, ms []*mastery.Mastery) error {
	m.masteries = append(m.masteries, ms...)
	return nil
}

// --- Tests ---

// Z8-AC01: Demo chapter is created with 8 items in UNKNOWN state
func TestZ8AC01_SeedDemoChapter(t *testing.T) {
	chapterRepo := &mockChapterRepoOnboarding{}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)
	userID := uuid.New()

	ch, items, err := svc.SeedDemoChapter(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Chapter is marked as demo
	if !ch.IsDemo {
		t.Error("expected chapter to be is_demo=true")
	}
	if ch.Subject != "Physique-Chimie" {
		t.Errorf("expected subject Physique-Chimie, got %s", ch.Subject)
	}
	if ch.Name != "Densité et masse volumique (démo)" {
		t.Errorf("expected demo chapter name, got %s", ch.Name)
	}

	// 8 items created
	if len(items) != 8 {
		t.Fatalf("expected 8 items, got %d", len(items))
	}

	// Items have correct types (6 KNOWLEDGE + 2 PROCEDURE)
	knowledgeCount := 0
	procedureCount := 0
	for _, item := range items {
		switch item.ItemType {
		case chapter.ItemKnowledge:
			knowledgeCount++
		case chapter.ItemProcedure:
			procedureCount++
		}
		// All items should have confidence 1.0 (trusted)
		if item.Confidence != 1.0 {
			t.Errorf("expected confidence 1.0 for demo item, got %f", item.Confidence)
		}
		// All items should NOT require validation
		if item.ValidationRequired {
			t.Error("demo items should not require validation")
		}
	}
	if knowledgeCount != 6 {
		t.Errorf("expected 6 KNOWLEDGE items, got %d", knowledgeCount)
	}
	if procedureCount != 2 {
		t.Errorf("expected 2 PROCEDURE items, got %d", procedureCount)
	}

	// 8 masteries created in UNKNOWN state
	if len(masteryRepo.masteries) != 8 {
		t.Fatalf("expected 8 masteries, got %d", len(masteryRepo.masteries))
	}
	for _, m := range masteryRepo.masteries {
		if m.State != mastery.Unknown {
			t.Errorf("expected mastery state UNKNOWN, got %s", m.State)
		}
	}
}

// Z8-AC01: SeedDemoChapter is idempotent
func TestZ8AC01_SeedDemoIdempotent(t *testing.T) {
	chapterRepo := &mockChapterRepoOnboarding{}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)
	userID := uuid.New()

	ch1, _, err := svc.SeedDemoChapter(context.Background(), userID)
	if err != nil {
		t.Fatalf("first seed error: %v", err)
	}

	ch2, _, err := svc.SeedDemoChapter(context.Background(), userID)
	if err != nil {
		t.Fatalf("second seed error: %v", err)
	}

	if ch1.ID != ch2.ID {
		t.Error("expected same chapter on second call (idempotent)")
	}
}

// Z8-AC01: Demo chapter auto-archived when first real chapter is ready
func TestZ8AC01_ArchiveDemoOnRealChapter(t *testing.T) {
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	userID := uuid.New()
	demoID := uuid.New()
	realID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: demoID, UserID: userID, IsDemo: true, Archived: false},
			{ID: realID, UserID: userID, IsDemo: false, Archived: false},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	err := svc.ArchiveDemoIfNeeded(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Demo should be archived
	for _, ch := range chapterRepo.chapters {
		if ch.ID == demoID && !ch.Archived {
			t.Error("expected demo chapter to be archived")
		}
		if ch.ID == realID && ch.Archived {
			t.Error("real chapter should NOT be archived")
		}
	}
}

// Z8-AC01: Demo chapter NOT archived if no real chapter exists
func TestZ8AC01_NoArchiveWithoutRealChapter(t *testing.T) {
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	userID := uuid.New()
	demoID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: demoID, UserID: userID, IsDemo: true, Archived: false},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	err := svc.ArchiveDemoIfNeeded(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chapterRepo.chapters[0].Archived {
		t.Error("demo chapter should NOT be archived when no real chapter exists")
	}
}

// Z8-AC02: Empty state shows correct onboarding steps
func TestZ8AC02_OnboardingStatusEmptyState(t *testing.T) {
	userID := uuid.New()
	demoID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: demoID, UserID: userID, IsDemo: true, Archived: false},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	status, err := svc.GetOnboardingStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.AccountCreated {
		t.Error("account should be marked as created")
	}
	if !status.HasDemoChapter {
		t.Error("should have demo chapter")
	}
	if status.FirstChapterReady {
		t.Error("first chapter should NOT be ready yet")
	}
	if status.DemoChapterID == nil || *status.DemoChapterID != demoID {
		t.Error("demo chapter ID should be set")
	}
}

// Z8-AC08: Deterministic onboarding sequence — new user at demo_session step
func TestZ8AC08_DeterministicSequence_DemoAvailable(t *testing.T) {
	userID := uuid.New()
	demoID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: demoID, UserID: userID, IsDemo: true, Archived: false},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	status, err := svc.GetOnboardingStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.CurrentStep != StepDemoSession {
		t.Errorf("current_step: got %q, want %q", status.CurrentStep, StepDemoSession)
	}
	if len(status.CompletedSteps) < 2 {
		t.Fatalf("expected at least 2 completed steps, got %d", len(status.CompletedSteps))
	}
	if status.CompletedSteps[0] != StepAccountCreated {
		t.Errorf("completed_steps[0]: got %q, want %q", status.CompletedSteps[0], StepAccountCreated)
	}
	if status.CompletedSteps[1] != StepDemoAvailable {
		t.Errorf("completed_steps[1]: got %q, want %q", status.CompletedSteps[1], StepDemoAvailable)
	}
}

// Z8-AC08: After first real chapter with items → at first_session step
func TestZ8AC08_DeterministicSequence_FirstUploadDone(t *testing.T) {
	userID := uuid.New()
	realChID := uuid.New()
	revID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: uuid.New(), UserID: userID, IsDemo: true, Archived: true},
			{ID: realChID, UserID: userID, IsDemo: false, Archived: false, CurrentRevisionID: &revID},
		},
		items: []*chapter.Item{
			{ID: uuid.New(), ChapterID: realChID, ItemType: chapter.ItemKnowledge},
			{ID: uuid.New(), ChapterID: realChID, ItemType: chapter.ItemKnowledge},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	status, err := svc.GetOnboardingStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.CurrentStep != StepFirstSession {
		t.Errorf("current_step: got %q, want %q", status.CurrentStep, StepFirstSession)
	}
	if !status.FirstChapterReady {
		t.Error("first_chapter_ready should be true")
	}
}

// Z8-AC02: Empty state disappears after first real chapter
func TestZ8AC02_EmptyStateDisappearsWithRealChapter(t *testing.T) {
	userID := uuid.New()

	chapterRepo := &mockChapterRepoOnboarding{
		chapters: []*chapter.Chapter{
			{ID: uuid.New(), UserID: userID, IsDemo: true, Archived: true},
			{ID: uuid.New(), UserID: userID, IsDemo: false, Archived: false},
		},
	}
	masteryRepo := &mockMasteryRepoOnboarding{}
	now := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	svc := NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	status, err := svc.GetOnboardingStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.FirstChapterReady {
		t.Error("first chapter should be ready")
	}
	if status.HasDemoChapter {
		t.Error("demo chapter should no longer be active (archived)")
	}
}
