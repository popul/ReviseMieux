//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func TestMasteryRepository_SaveAndFindByID(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")
	m := newMasteryFixture(userID, item.ID, now)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != m.ID {
		t.Errorf("ID = %v, want %v", got.ID, m.ID)
	}
	if got.State != mastery.Unknown {
		t.Errorf("State = %v, want UNKNOWN", got.State)
	}
}

func TestMasteryRepository_FindByID_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasteryRepository_FindByUserAndItem(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")
	m := newMasteryFixture(userID, item.ID, now)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByUserAndItem(ctx, userID, item.ID)
	if err != nil {
		t.Fatalf("FindByUserAndItem: %v", err)
	}
	if got.ID != m.ID {
		t.Errorf("ID = %v, want %v", got.ID, m.ID)
	}
}

func TestMasteryRepository_FindDueByUser(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	dueTime := now.Add(-1 * time.Hour)

	item1 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 1")
	m := newMasteryFixture(userID, item1.ID, now)
	m.NextDueAt = &dueTime
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	item2 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 2")
	m2 := newMasteryFixture(userID, item2.ID, now)
	futureTime := now.Add(24 * time.Hour)
	m2.NextDueAt = &futureTime
	if err := repo.Save(ctx, m2); err != nil {
		t.Fatalf("Save m2: %v", err)
	}

	results, err := repo.FindDueByUser(ctx, userID, now)
	if err != nil {
		t.Fatalf("FindDueByUser: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].ID != m.ID {
		t.Errorf("got ID %v, want %v", results[0].ID, m.ID)
	}
}

func TestMasteryRepository_FindByUserAndState(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	item1 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 1")
	m1 := newMasteryFixture(userID, item1.ID, now)
	m1.State = mastery.Fragile
	if err := repo.Save(ctx, m1); err != nil {
		t.Fatalf("Save m1: %v", err)
	}

	item2 := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item 2")
	m2 := newMasteryFixture(userID, item2.ID, now)
	m2.State = mastery.OK
	if err := repo.Save(ctx, m2); err != nil {
		t.Fatalf("Save m2: %v", err)
	}

	results, err := repo.FindByUserAndState(ctx, userID, mastery.Fragile)
	if err != nil {
		t.Fatalf("FindByUserAndState: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].ID != m1.ID {
		t.Errorf("got ID %v, want %v", results[0].ID, m1.ID)
	}
}

func TestMasteryRepository_SaveAll(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)

	items := make([]*mastery.Mastery, 3)
	for i := range items {
		item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Item")
		items[i] = newMasteryFixture(userID, item.ID, now)
	}

	if err := repo.SaveAll(ctx, items); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	for _, m := range items {
		got, err := repo.FindByID(ctx, m.ID)
		if err != nil {
			t.Fatalf("FindByID(%v): %v", m.ID, err)
		}
		if got.State != mastery.Unknown {
			t.Errorf("State = %v, want UNKNOWN", got.State)
		}
	}
}

func TestMasteryRepository_SaveUpsert(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewMasteryRepository(tdb.pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, chapter.ItemKnowledge, "Densité")
	m := newMasteryFixture(userID, item.ID, now)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save (insert): %v", err)
	}

	m.State = mastery.Fragile
	m.ConsecutiveSuccesses = 1
	m.UpdatedAt = now.Add(time.Hour)
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save (update): %v", err)
	}

	got, err := repo.FindByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.State != mastery.Fragile {
		t.Errorf("State = %v, want FRAGILE", got.State)
	}
	if got.ConsecutiveSuccesses != 1 {
		t.Errorf("ConsecutiveSuccesses = %d, want 1", got.ConsecutiveSuccesses)
	}
}
