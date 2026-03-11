//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func newMasteryFixture(userID, itemID uuid.UUID, now time.Time) *mastery.Mastery {
	return &mastery.Mastery{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		ItemID:    itemID,
		State:     mastery.Unknown,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMasteryRepository_SaveAndFindByID(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())
	itemID := uuid.Must(uuid.NewV7())
	m := newMasteryFixture(userID, itemID, now)

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
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasteryRepository_FindByUserAndItem(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())
	itemID := uuid.Must(uuid.NewV7())
	m := newMasteryFixture(userID, itemID, now)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByUserAndItem(ctx, userID, itemID)
	if err != nil {
		t.Fatalf("FindByUserAndItem: %v", err)
	}
	if got.ID != m.ID {
		t.Errorf("ID = %v, want %v", got.ID, m.ID)
	}
}

func TestMasteryRepository_FindDueByUser(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())
	dueTime := now.Add(-1 * time.Hour)

	// Create a mastery that is due
	m := newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now)
	m.NextDueAt = &dueTime
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Create a mastery that is NOT due
	m2 := newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now)
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
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())

	m1 := newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now)
	m1.State = mastery.Fragile
	if err := repo.Save(ctx, m1); err != nil {
		t.Fatalf("Save m1: %v", err)
	}

	m2 := newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now)
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
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())
	masteries := []*mastery.Mastery{
		newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now),
		newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now),
		newMasteryFixture(userID, uuid.Must(uuid.NewV7()), now),
	}

	if err := repo.SaveAll(ctx, masteries); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	for _, m := range masteries {
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
	pool := setupTestPool(t)
	repo := postgres.NewMasteryRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := uuid.Must(uuid.NewV7())
	itemID := uuid.Must(uuid.NewV7())
	m := newMasteryFixture(userID, itemID, now)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save (insert): %v", err)
	}

	// Update state and save again (upsert)
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
