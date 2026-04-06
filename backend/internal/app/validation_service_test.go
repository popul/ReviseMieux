package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/validation"
)

// --- Validation mocks ---

type mockValidationRepo struct {
	tasks map[uuid.UUID]*validation.ValidationTask
}

func newMockValidationRepo() *mockValidationRepo {
	return &mockValidationRepo{tasks: make(map[uuid.UUID]*validation.ValidationTask)}
}

func (m *mockValidationRepo) FindByID(_ context.Context, id uuid.UUID) (*validation.ValidationTask, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, validation.ErrNotFound
	}
	return t, nil
}

func (m *mockValidationRepo) FindPendingByItem(_ context.Context, itemID uuid.UUID) ([]*validation.ValidationTask, error) {
	var result []*validation.ValidationTask
	for _, t := range m.tasks {
		if t.ItemID == itemID && t.Status == validation.StatusPending {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockValidationRepo) FindPendingAll(_ context.Context, limit int) ([]*validation.ValidationTask, error) {
	var result []*validation.ValidationTask
	for _, t := range m.tasks {
		if t.Status == validation.StatusPending {
			result = append(result, t)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockValidationRepo) Save(_ context.Context, t *validation.ValidationTask) error {
	m.tasks[t.ID] = t
	return nil
}

// --- Tests ---

// Z3-AC06: chapter usable with 0 validations done
func TestZ3AC06_ChapterUsableWithZeroValidations(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	// Create chapter with items — some have validation_required
	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	revID := uuid.Must(uuid.NewV7())
	ch.CurrentRevisionID = &revID
	chRepo.Save(context.Background(), ch)
	chRepo.SaveRevision(context.Background(), &chapter.Revision{
		ID: revID, ChapterID: ch.ID, RevisionNumber: 1,
		Status: chapter.RevisionReady, CreatedAt: now,
	})

	// 6 items: 4 normal, 2 with validation_required
	for i := 0; i < 4; i++ {
		term := "item_normal_" + string(rune('A'+i))
		chRepo.SaveItem(context.Background(), &chapter.Item{
			ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, RevisionID: revID,
			ItemType: chapter.ItemKnowledge, Term: &term,
			ValidationRequired: false, CreatedAt: now, UpdatedAt: now,
		})
	}
	for i := 0; i < 2; i++ {
		term := "item_uncertain_" + string(rune('A'+i))
		itemID := uuid.Must(uuid.NewV7())
		chRepo.SaveItem(context.Background(), &chapter.Item{
			ID: itemID, ChapterID: ch.ID, RevisionID: revID,
			ItemType: chapter.ItemKnowledge, Term: &term,
			ValidationRequired: true, CreatedAt: now, UpdatedAt: now,
		})
		// Create PENDING validation task
		task := validation.NewValidationTask(idGen, itemID, validation.SourceUncertainty, now)
		valRepo.Save(context.Background(), task)
	}

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)

	// Chapter should be usable — diagnostic accessible
	usable, pendingCount, err := svc.IsChapterUsable(context.Background(), ch.ID)
	if err != nil {
		t.Fatalf("IsChapterUsable: %v", err)
	}
	if !usable {
		t.Error("chapter should be usable even with pending validations")
	}
	if pendingCount != 2 {
		t.Errorf("pendingCount = %d, want 2", pendingCount)
	}
}

// Z3-AC02: confirm action
func TestZ3AC02_ConfirmValidationTask(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Test", now)
	revID := uuid.Must(uuid.NewV7())
	ch.CurrentRevisionID = &revID
	chRepo.Save(context.Background(), ch)

	itemID := uuid.Must(uuid.NewV7())
	term := "ρ = m / V"
	chRepo.SaveItem(context.Background(), &chapter.Item{
		ID: itemID, ChapterID: ch.ID, RevisionID: revID,
		ItemType: chapter.ItemKnowledge, Term: &term,
		ValidationRequired: true, Confidence: 0.6,
		CreatedAt: now, UpdatedAt: now,
	})

	task := validation.NewValidationTask(idGen, itemID, validation.SourceUncertainty, now)
	suggestion := "ρ = m / V"
	task.Suggestion = &suggestion
	valRepo.Save(context.Background(), task)

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)
	resolverID := uuid.Must(uuid.NewV7())

	err := svc.Confirm(context.Background(), task.ID, resolverID)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}

	// Task should be CONFIRMED
	updated, _ := valRepo.FindByID(context.Background(), task.ID)
	if updated.Status != validation.StatusConfirmed {
		t.Errorf("task status = %s, want CONFIRMED", updated.Status)
	}

	// Item should have validation_required=false, confidence >= 0.85
	item, _ := chRepo.FindItemByID(context.Background(), itemID)
	if item.ValidationRequired {
		t.Error("item.ValidationRequired should be false after confirm")
	}
	if item.Confidence < 0.85 {
		t.Errorf("item.Confidence = %f, want >= 0.85", item.Confidence)
	}

	// ValidationResolved event should be published
	if len(publisher.events) == 0 {
		t.Fatal("expected ValidationResolved event")
	}
	if _, ok := publisher.events[0].(event.ValidationResolved); !ok {
		t.Errorf("expected ValidationResolved, got %T", publisher.events[0])
	}
}

// Z3-AC03: correct action
func TestZ3AC03_CorrectValidationTask(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Test", now)
	revID := uuid.Must(uuid.NewV7())
	ch.CurrentRevisionID = &revID
	chRepo.Save(context.Background(), ch)

	itemID := uuid.Must(uuid.NewV7())
	term := "wrong term"
	chRepo.SaveItem(context.Background(), &chapter.Item{
		ID: itemID, ChapterID: ch.ID, RevisionID: revID,
		ItemType: chapter.ItemKnowledge, Term: &term,
		ValidationRequired: true, Confidence: 0.5,
		CreatedAt: now, UpdatedAt: now,
	})

	task := validation.NewValidationTask(idGen, itemID, validation.SourceUncertainty, now)
	valRepo.Save(context.Background(), task)

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)
	resolverID := uuid.Must(uuid.NewV7())
	correctedTerm := "correct term"

	err := svc.Correct(context.Background(), task.ID, resolverID, &correctedTerm)
	if err != nil {
		t.Fatalf("Correct: %v", err)
	}

	// Task should be CORRECTED
	updated, _ := valRepo.FindByID(context.Background(), task.ID)
	if updated.Status != validation.StatusCorrected {
		t.Errorf("task status = %s, want CORRECTED", updated.Status)
	}

	// Item should have updated term, confidence=1.0, validation_required=false
	item, _ := chRepo.FindItemByID(context.Background(), itemID)
	if *item.Term != "correct term" {
		t.Errorf("item.Term = %q, want %q", *item.Term, "correct term")
	}
	if item.Confidence != 1.0 {
		t.Errorf("item.Confidence = %f, want 1.0", item.Confidence)
	}
	if item.ValidationRequired {
		t.Error("item.ValidationRequired should be false")
	}
}

// Z3-AC04: "je ne sais pas" action
func TestZ3AC04_MarkUnknownValidationTask(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	itemID := uuid.Must(uuid.NewV7())
	task := validation.NewValidationTask(idGen, itemID, validation.SourceUncertainty, now)
	valRepo.Save(context.Background(), task)

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)
	resolverID := uuid.Must(uuid.NewV7())

	err := svc.MarkUnknown(context.Background(), task.ID, resolverID)
	if err != nil {
		t.Fatalf("MarkUnknown: %v", err)
	}

	// Original task deferred
	updated, _ := valRepo.FindByID(context.Background(), task.ID)
	if updated.Status != validation.StatusDeferredByStudent {
		t.Errorf("task status = %s, want DEFERRED_BY_STUDENT", updated.Status)
	}

	// Admin task should be created
	var adminTasks []*validation.ValidationTask
	for _, tk := range valRepo.tasks {
		if tk.ID != task.ID && tk.ItemID == itemID {
			adminTasks = append(adminTasks, tk)
		}
	}
	if len(adminTasks) != 1 {
		t.Fatalf("expected 1 admin task, got %d", len(adminTasks))
	}
	if adminTasks[0].Source != validation.SourceStudentDeferred {
		t.Errorf("admin task source = %s, want student_deferred", adminTasks[0].Source)
	}
}

// Z3-AC05: ignore action
func TestZ3AC05_IgnoreValidationTask(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	itemID := uuid.Must(uuid.NewV7())
	task := validation.NewValidationTask(idGen, itemID, validation.SourceUncertainty, now)
	valRepo.Save(context.Background(), task)

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)
	resolverID := uuid.Must(uuid.NewV7())

	err := svc.Ignore(context.Background(), task.ID, resolverID)
	if err != nil {
		t.Fatalf("Ignore: %v", err)
	}

	updated, _ := valRepo.FindByID(context.Background(), task.ID)
	if updated.Status != validation.StatusIgnoredByStudent {
		t.Errorf("task status = %s, want IGNORED_BY_STUDENT", updated.Status)
	}
}

// Z3-AC09: no validation task if confidence > 0.85
func TestZ3AC09_NoValidationIfHighConfidence(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)

	// All items have high confidence
	items := []*chapter.Item{
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.90},
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.95},
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.85},
	}

	tasks, err := svc.CreateValidationTasksIfNeeded(context.Background(), items)
	if err != nil {
		t.Fatalf("CreateValidationTasksIfNeeded: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for high-confidence items, got %d", len(tasks))
	}
}

// Z3-AC09: validation tasks created for low confidence items
func TestZ3AC09_ValidationForLowConfidence(t *testing.T) {
	now := time.Date(2026, 3, 12, 18, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	valRepo := newMockValidationRepo()
	publisher := &mockPublisher{}

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)

	items := []*chapter.Item{
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.90},
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.60}, // below threshold
		{ID: uuid.Must(uuid.NewV7()), Confidence: 0.40}, // below threshold
	}

	tasks, err := svc.CreateValidationTasksIfNeeded(context.Background(), items)
	if err != nil {
		t.Fatalf("CreateValidationTasksIfNeeded: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for low-confidence items, got %d", len(tasks))
	}
}

// Z3-AC01: template restriction for validation_required items
func TestZ3AC01_TemplateRestrictionForValidationRequired(t *testing.T) {
	svc := &ValidationService{}

	// Item with validation_required
	eligible := svc.EligibleTemplates(true)
	for _, tmpl := range eligible {
		switch tmpl {
		case "GEN.KNOW.DEF_SHORT", "GEN.KNOW.FLASH_MCQ":
			// OK — allowed
		default:
			t.Errorf("template %q should not be eligible for validation_required items", tmpl)
		}
	}
	if len(eligible) != 2 {
		t.Errorf("expected 2 eligible templates, got %d", len(eligible))
	}

	// Item without validation_required
	allTemplates := svc.EligibleTemplates(false)
	if len(allTemplates) <= 2 {
		t.Error("non-restricted items should have more than 2 templates")
	}
}
