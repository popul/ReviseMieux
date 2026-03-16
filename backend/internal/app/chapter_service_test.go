package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// Z5-AC04: only items from current revision are returned
func TestZ5AC04_LessonCard_OnlyCurrentRevisionItems(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	ctx := context.Background()

	// Create chapter with R2 as current revision
	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	r1ID := uuid.Must(uuid.NewV7())
	r2ID := uuid.Must(uuid.NewV7())
	ch.CurrentRevisionID = &r2ID
	chRepo.Save(ctx, ch)

	chRepo.SaveRevision(ctx, &chapter.Revision{ID: r1ID, ChapterID: ch.ID, RevisionNumber: 1, Status: chapter.RevisionReady, CreatedAt: now})
	chRepo.SaveRevision(ctx, &chapter.Revision{ID: r2ID, ChapterID: ch.ID, RevisionNumber: 2, Status: chapter.RevisionReady, CreatedAt: now})

	// R1 item (archived — replaced by R2)
	term1 := "vitesse"
	item1 := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, RevisionID: r1ID,
		ItemType: chapter.ItemKnowledge, Term: &term1, Archived: true,
		CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SaveItem(ctx, item1)

	// R2 item (active)
	term2 := "vitesse"
	item2 := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, RevisionID: r2ID,
		ItemType: chapter.ItemKnowledge, Term: &term2, Archived: false,
		CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SaveItem(ctx, item2)

	svc := NewChapterService(chRepo)

	items, err := svc.GetLessonCardItems(ctx, ch.ID)
	if err != nil {
		t.Fatalf("GetLessonCardItems: %v", err)
	}

	// Only R2 item should be returned (not archived R1 item)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != item2.ID {
		t.Errorf("expected item2 (R2), got item %v", items[0].ID)
	}
}

// Z5-AC06: archived items not included in session composition
func TestZ5AC06_ArchivedItemsExcludedFromSession(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	ctx := context.Background()

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement", now)
	revID := uuid.Must(uuid.NewV7())
	ch.CurrentRevisionID = &revID
	chRepo.Save(ctx, ch)
	chRepo.SaveRevision(ctx, &chapter.Revision{ID: revID, ChapterID: ch.ID, RevisionNumber: 1, Status: chapter.RevisionReady, CreatedAt: now})

	// Active item
	term1 := "vitesse"
	activeItem := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, RevisionID: revID,
		ItemType: chapter.ItemKnowledge, Term: &term1, Archived: false,
		CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SaveItem(ctx, activeItem)

	// Archived item (from old revision, should be excluded)
	term2 := "masse"
	archivedItem := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID, RevisionID: revID,
		ItemType: chapter.ItemKnowledge, Term: &term2, Archived: true,
		CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SaveItem(ctx, archivedItem)

	svc := NewChapterService(chRepo)

	items, err := svc.GetSessionEligibleItems(ctx, ch.ID)
	if err != nil {
		t.Fatalf("GetSessionEligibleItems: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 eligible item, got %d", len(items))
	}
	if items[0].ID != activeItem.ID {
		t.Errorf("expected activeItem, got %v", items[0].ID)
	}
}
