//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/validation"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func TestValidationRepository_SaveAndFindByID(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")

	task := tdb.seedValidationTask(item.ID, validation.StatusPending)

	got, err := repo.FindByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("ID = %v, want %v", got.ID, task.ID)
	}
	if got.Status != validation.StatusPending {
		t.Errorf("Status = %v, want PENDING", got.Status)
	}
	if got.Priority != 5 {
		t.Errorf("Priority = %d, want 5", got.Priority)
	}
}

func TestValidationRepository_FindByID_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidationRepository_FindPendingByItem(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")

	// 2 pending + 1 confirmed
	tdb.seedValidationTask(item.ID, validation.StatusPending)
	tdb.seedValidationTask(item.ID, validation.StatusPending)
	tdb.seedValidationTask(item.ID, validation.StatusConfirmed)

	pending, err := repo.FindPendingByItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("FindPendingByItem: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending count = %d, want 2", len(pending))
	}
}

func TestValidationRepository_FindPendingAll(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	// Create 5 pending tasks across 2 items
	item1 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 1")
	item2 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 2")
	for i := 0; i < 3; i++ {
		tdb.seedValidationTask(item1.ID, validation.StatusPending)
	}
	for i := 0; i < 2; i++ {
		tdb.seedValidationTask(item2.ID, validation.StatusPending)
	}

	// Limit 3
	tasks, err := repo.FindPendingAll(ctx, 3)
	if err != nil {
		t.Fatalf("FindPendingAll: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("count = %d, want 3", len(tasks))
	}

	// Limit 100 (get all)
	all, err := repo.FindPendingAll(ctx, 100)
	if err != nil {
		t.Fatalf("FindPendingAll (all): %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("all count = %d, want 5", len(all))
	}
}

func TestValidationRepository_SaveUpsert(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")

	task := tdb.seedValidationTask(item.ID, validation.StatusPending)

	// Resolve it
	resolverID := tdb.seedUser("parent")
	got, _ := repo.FindByID(ctx, task.ID)
	got.Status = validation.StatusConfirmed
	got.ResolvedBy = &resolverID

	if err := repo.Save(ctx, got); err != nil {
		t.Fatalf("Save (update): %v", err)
	}

	updated, _ := repo.FindByID(ctx, task.ID)
	if updated.Status != validation.StatusConfirmed {
		t.Errorf("Status = %v, want CONFIRMED", updated.Status)
	}
	if *updated.ResolvedBy != resolverID {
		t.Errorf("ResolvedBy = %v, want %v", *updated.ResolvedBy, resolverID)
	}
}

func TestValidationRepository_PriorityOrdering(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewValidationRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")

	// Create tasks with different priorities
	low := tdb.seedValidationTask(item.ID, validation.StatusPending)
	high := tdb.seedValidationTask(item.ID, validation.StatusPending)

	// Update priorities directly
	tdb.pool.Exec(ctx, `UPDATE validation_tasks SET priority = 1 WHERE id = $1`, low.ID)
	tdb.pool.Exec(ctx, `UPDATE validation_tasks SET priority = 10 WHERE id = $1`, high.ID)

	tasks, err := repo.FindPendingByItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("FindPendingByItem: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("count = %d, want 2", len(tasks))
	}
	// High priority should come first (ORDER BY priority DESC)
	if tasks[0].Priority != 10 {
		t.Errorf("first task priority = %d, want 10 (highest first)", tasks[0].Priority)
	}
}
