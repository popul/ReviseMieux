//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func TestChapterRepository_SaveAndFindByID(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	now := time.Now().UTC().Truncate(time.Microsecond)

	ch := &chapter.Chapter{
		ID:         uuid.Must(uuid.NewV7()),
		UserID:     userID,
		Subject:    "Mathématiques",
		ClassLevel: "3e",
		Name:       "Théorème de Pythagore",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repo.Save(ctx, ch); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, ch.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name != "Théorème de Pythagore" {
		t.Errorf("Name = %q, want %q", got.Name, "Théorème de Pythagore")
	}
	if got.Subject != "Mathématiques" {
		t.Errorf("Subject = %q, want %q", got.Subject, "Mathématiques")
	}
}

func TestChapterRepository_FindByID_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChapterRepository_FindByUser(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Active chapter
	ch1 := &chapter.Chapter{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		Subject: "PC", ClassLevel: "4e", Name: "Chapitre 1",
		CreatedAt: now, UpdatedAt: now,
	}
	// Archived chapter
	ch2 := &chapter.Chapter{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		Subject: "PC", ClassLevel: "4e", Name: "Chapitre 2",
		Archived: true, CreatedAt: now, UpdatedAt: now,
	}
	repo.Save(ctx, ch1)
	repo.Save(ctx, ch2)

	// Exclude archived
	active, err := repo.FindByUser(ctx, userID, false)
	if err != nil {
		t.Fatalf("FindByUser (active): %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active count = %d, want 1", len(active))
	}

	// Include archived
	all, err := repo.FindByUser(ctx, userID, true)
	if err != nil {
		t.Fatalf("FindByUser (all): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all count = %d, want 2", len(all))
	}
}

func TestChapterRepository_SaveUpsert(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	now := time.Now().UTC().Truncate(time.Microsecond)

	ch := &chapter.Chapter{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		Subject: "SVT", ClassLevel: "5e", Name: "Original",
		CreatedAt: now, UpdatedAt: now,
	}
	repo.Save(ctx, ch)

	ch.Name = "Updated"
	ch.UpdatedAt = now.Add(time.Hour)
	repo.Save(ctx, ch)

	got, _ := repo.FindByID(ctx, ch.ID)
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want Updated", got.Name)
	}
}

func TestChapterRepository_RevisionCRUD(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	rev := &chapter.Revision{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID,
		RevisionNumber: 1, Status: chapter.RevisionProcessing,
		CreatedAt: now,
	}

	if err := repo.SaveRevision(ctx, rev); err != nil {
		t.Fatalf("SaveRevision: %v", err)
	}

	got, err := repo.FindRevisionByID(ctx, rev.ID)
	if err != nil {
		t.Fatalf("FindRevisionByID: %v", err)
	}
	if got.Status != chapter.RevisionProcessing {
		t.Errorf("Status = %v, want PROCESSING", got.Status)
	}

	// Update status
	rev.Status = chapter.RevisionReady
	repo.SaveRevision(ctx, rev)

	got, _ = repo.FindRevisionByID(ctx, rev.ID)
	if got.Status != chapter.RevisionReady {
		t.Errorf("Status = %v, want READY", got.Status)
	}
}

func TestChapterRepository_CurrentRevision(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	got, err := repo.FindCurrentRevision(ctx, ch.ID)
	if err != nil {
		t.Fatalf("FindCurrentRevision: %v", err)
	}
	if got.ID != rev.ID {
		t.Errorf("ID = %v, want %v", got.ID, rev.ID)
	}
}

func TestChapterRepository_PageCRUD(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	page := &chapter.Page{
		ID: uuid.Must(uuid.NewV7()), RevisionID: rev.ID,
		PhotoURL: "s3://bucket/photo1.jpg", PageOrder: 0,
		OCRStatus: chapter.PageOCRPending, CreatedAt: now, UpdatedAt: now,
	}

	if err := repo.SavePage(ctx, page); err != nil {
		t.Fatalf("SavePage: %v", err)
	}

	pages, err := repo.FindPagesByRevision(ctx, rev.ID)
	if err != nil {
		t.Fatalf("FindPagesByRevision: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("pages count = %d, want 1", len(pages))
	}
	if pages[0].PhotoURL != "s3://bucket/photo1.jpg" {
		t.Errorf("PhotoURL = %q", pages[0].PhotoURL)
	}
}

func TestChapterRepository_ItemWithKeywordsAndSteps(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	notion := tdb.seedNotion(ch.ID, "Masse volumique", 0)

	now := time.Now().UTC().Truncate(time.Microsecond)
	term := "Densité"
	item := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID,
		NotionID: &notion.ID, RevisionID: rev.ID,
		ItemType: chapter.ItemKnowledge, Term: &term,
		Confidence: 0.9, Keywords: []string{"densité", "masse", "volume"},
		CreatedAt: now, UpdatedAt: now,
	}

	if err := repo.SaveItem(ctx, item); err != nil {
		t.Fatalf("SaveItem: %v", err)
	}

	got, err := repo.FindItemByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("FindItemByID: %v", err)
	}
	if *got.Term != "Densité" {
		t.Errorf("Term = %v, want Densité", *got.Term)
	}
	if len(got.Keywords) != 3 {
		t.Fatalf("Keywords count = %d, want 3", len(got.Keywords))
	}
	if got.Keywords[0] != "densité" {
		t.Errorf("Keywords[0] = %q, want densité", got.Keywords[0])
	}

	// Test PROCEDURE item with steps
	procTerm := "Calcul de densité"
	procItem := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID,
		NotionID: &notion.ID, RevisionID: rev.ID,
		ItemType: chapter.ItemProcedure, Term: &procTerm,
		Confidence: 0.85,
		Keywords:   []string{"calcul"},
		Steps: []chapter.ItemStep{
			{ID: uuid.Must(uuid.NewV7()), StepOrder: 0, Content: "Mesurer la masse"},
			{ID: uuid.Must(uuid.NewV7()), StepOrder: 1, Content: "Mesurer le volume"},
			{ID: uuid.Must(uuid.NewV7()), StepOrder: 2, Content: "Diviser masse par volume"},
		},
		CreatedAt: now, UpdatedAt: now,
	}
	// Set ItemID on steps
	for i := range procItem.Steps {
		procItem.Steps[i].ItemID = procItem.ID
	}

	if err := repo.SaveItem(ctx, procItem); err != nil {
		t.Fatalf("SaveItem (procedure): %v", err)
	}

	gotProc, err := repo.FindItemByID(ctx, procItem.ID)
	if err != nil {
		t.Fatalf("FindItemByID (procedure): %v", err)
	}
	if len(gotProc.Steps) != 3 {
		t.Fatalf("Steps count = %d, want 3", len(gotProc.Steps))
	}
	if gotProc.Steps[2].Content != "Diviser masse par volume" {
		t.Errorf("Steps[2].Content = %q", gotProc.Steps[2].Content)
	}
}

func TestChapterRepository_FindItemsByChapter(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	item1 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 1")
	item2 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 2")

	// Archive item2
	_, _ = tdb.pool.Exec(ctx, `UPDATE items SET archived = true WHERE id = $1`, item2.ID)

	// Without archived
	items, err := repo.FindItemsByChapter(ctx, ch.ID, false)
	if err != nil {
		t.Fatalf("FindItemsByChapter (active): %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("active count = %d, want 1", len(items))
	}
	if items[0].ID != item1.ID {
		t.Errorf("ID = %v, want %v", items[0].ID, item1.ID)
	}

	// With archived
	all, err := repo.FindItemsByChapter(ctx, ch.ID, true)
	if err != nil {
		t.Fatalf("FindItemsByChapter (all): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all count = %d, want 2", len(all))
	}
}

func TestChapterRepository_SaveItems_Batch(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	items := make([]*chapter.Item, 5)
	for i := range items {
		term := "Item " + string(rune('A'+i))
		items[i] = &chapter.Item{
			ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID,
			RevisionID: rev.ID, ItemType: chapter.ItemKnowledge,
			Term: &term, Confidence: 0.8,
			CreatedAt: now, UpdatedAt: now,
		}
	}

	if err := repo.SaveItems(ctx, items); err != nil {
		t.Fatalf("SaveItems: %v", err)
	}

	found, err := repo.FindItemsByChapter(ctx, ch.ID, false)
	if err != nil {
		t.Fatalf("FindItemsByChapter: %v", err)
	}
	if len(found) != 5 {
		t.Errorf("count = %d, want 5", len(found))
	}
}

func TestChapterRepository_NotionCRUD(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewChapterRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	n := &chapter.Notion{
		ID: uuid.Must(uuid.NewV7()), ChapterID: ch.ID,
		Name: "Masse volumique", SortOrder: 0, CreatedAt: now,
	}

	if err := repo.SaveNotion(ctx, n); err != nil {
		t.Fatalf("SaveNotion: %v", err)
	}

	notions, err := repo.FindNotionsByChapter(ctx, ch.ID)
	if err != nil {
		t.Fatalf("FindNotionsByChapter: %v", err)
	}
	if len(notions) != 1 {
		t.Fatalf("count = %d, want 1", len(notions))
	}
	if notions[0].Name != "Masse volumique" {
		t.Errorf("Name = %q", notions[0].Name)
	}

	// Upsert
	n.Name = "Masse volumique (updated)"
	repo.SaveNotion(ctx, n)

	notions, _ = repo.FindNotionsByChapter(ctx, ch.ID)
	if notions[0].Name != "Masse volumique (updated)" {
		t.Errorf("upserted Name = %q", notions[0].Name)
	}
}
