//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/db"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
	"github.com/popul/revisemieux/internal/domain/validation"
)

// testDB holds a shared pool and provides helpers for integration tests.
type testDB struct {
	pool *pgxpool.Pool
	t    *testing.T
}

// setupTestDB creates a pool, runs migrations, and returns a testDB helper.
// Each test should call cleanup() to truncate tables.
func setupTestDB(t *testing.T) *testDB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL not set — integration tests require a dedicated test DB (see `make test-db-up`)")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Run migrations
	if err := db.Migrate(context.Background(), pool, migrationsDir()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tdb := &testDB{pool: pool, t: t}
	// Truncate at the START to ensure clean state — avoids races with t.Cleanup from previous tests.
	tdb.truncateAll()
	t.Cleanup(func() { tdb.truncateAll() })
	return tdb
}

func migrationsDir() string {
	// Go tests run from the package source directory.
	// From backend/internal/infra/postgres/ → backend/migrations/
	// Use runtime.Caller to get absolute path.
	_, filename, _, _ := runtimeCaller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

var runtimeCaller = runtime.Caller

// truncateAll removes all data from all tables using TRUNCATE CASCADE.
// Only root tables need to be listed — CASCADE handles FK-dependent tables.
func (tdb *testDB) truncateAll() {
	ctx := context.Background()
	_, _ = tdb.pool.Exec(ctx, `TRUNCATE users, exams, templates, llm_call_logs CASCADE`)
}

// seedUser inserts a test user and returns its ID.
func (tdb *testDB) seedUser(role string) uuid.UUID {
	tdb.t.Helper()
	id := uuid.Must(uuid.NewV7())
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO users (id, role, display_name) VALUES ($1, $2, $3)`,
		id, role, "Test User")
	if err != nil {
		tdb.t.Fatalf("seedUser: %v", err)
	}
	return id
}

// seedChapter inserts a chapter and returns it.
func (tdb *testDB) seedChapter(userID uuid.UUID) *chapter.Chapter {
	tdb.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	ch := &chapter.Chapter{
		ID:         uuid.Must(uuid.NewV7()),
		UserID:     userID,
		Subject:    "Physique-Chimie",
		ClassLevel: "4e",
		Name:       "Densité et masse volumique",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO chapters (id, user_id, subject, class_level, name, archived, is_demo, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,false,false,$6,$7)`,
		ch.ID, ch.UserID, ch.Subject, ch.ClassLevel, ch.Name, ch.CreatedAt, ch.UpdatedAt)
	if err != nil {
		tdb.t.Fatalf("seedChapter: %v", err)
	}
	return ch
}

// seedRevision inserts a revision for a chapter.
func (tdb *testDB) seedRevision(chapterID uuid.UUID) *chapter.Revision {
	tdb.t.Helper()
	rev := &chapter.Revision{
		ID:             uuid.Must(uuid.NewV7()),
		ChapterID:      chapterID,
		RevisionNumber: 1,
		Status:         chapter.RevisionProcessing,
		CreatedAt:      time.Now().UTC().Truncate(time.Microsecond),
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO chapter_revisions (id, chapter_id, revision_number, status, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		rev.ID, rev.ChapterID, rev.RevisionNumber, rev.Status, rev.CreatedAt)
	if err != nil {
		tdb.t.Fatalf("seedRevision: %v", err)
	}
	// Link as current revision
	_, err = tdb.pool.Exec(context.Background(),
		`UPDATE chapters SET current_revision_id = $1 WHERE id = $2`,
		rev.ID, chapterID)
	if err != nil {
		tdb.t.Fatalf("seedRevision link: %v", err)
	}
	return rev
}

// seedNotion inserts a notion for a chapter.
func (tdb *testDB) seedNotion(chapterID uuid.UUID, name string, order int) *chapter.Notion {
	tdb.t.Helper()
	n := &chapter.Notion{
		ID:        uuid.Must(uuid.NewV7()),
		ChapterID: chapterID,
		Name:      name,
		SortOrder: order,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO notions (id, chapter_id, name, sort_order, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		n.ID, n.ChapterID, n.Name, n.SortOrder, n.CreatedAt)
	if err != nil {
		tdb.t.Fatalf("seedNotion: %v", err)
	}
	return n
}

// seedItem inserts an item for a chapter+revision.
func (tdb *testDB) seedItem(chapterID, revisionID uuid.UUID, notionID *uuid.UUID, itemType chapter.ItemType, term string) *chapter.Item {
	tdb.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	item := &chapter.Item{
		ID:         uuid.Must(uuid.NewV7()),
		ChapterID:  chapterID,
		NotionID:   notionID,
		RevisionID: revisionID,
		ItemType:   itemType,
		Term:       &term,
		Confidence: 0.9,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO items (id, chapter_id, notion_id, revision_id, item_type, term,
		                    confidence, validation_required, archived, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		item.ID, item.ChapterID, item.NotionID, item.RevisionID,
		item.ItemType, item.Term, item.Confidence, false, false,
		item.CreatedAt, item.UpdatedAt)
	if err != nil {
		tdb.t.Fatalf("seedItem: %v", err)
	}
	return item
}

// seedTemplate inserts a question template.
func (tdb *testDB) seedTemplate(id string) {
	tdb.t.Helper()
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO templates (id, name, version, question_type, difficulty, prompt_template)
		 VALUES ($1, $2, 1, 'MCQ', 1, 'What is {{term}}?')
		 ON CONFLICT DO NOTHING`,
		id, id)
	if err != nil {
		tdb.t.Fatalf("seedTemplate: %v", err)
	}
}

// seedSession inserts a session.
func (tdb *testDB) seedSession(userID uuid.UUID, sType session.SessionType, status session.SessionStatus) *session.Session {
	tdb.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	s := &session.Session{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      userID,
		SessionType: sType,
		Status:      status,
		TriggerType: session.TriggerManual,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO sessions (id, user_id, session_type, status, trigger_type, current_question_index, includes_pre_class, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,0,false,$6,$7)`,
		s.ID, s.UserID, s.SessionType, s.Status, s.TriggerType, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		tdb.t.Fatalf("seedSession: %v", err)
	}
	return s
}

// seedQuestion inserts a question linked to a session, item and template.
func (tdb *testDB) seedQuestion(sessionID, itemID uuid.UUID, templateID string) *session.Question {
	tdb.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	q := &session.Question{
		ID:             uuid.Must(uuid.NewV7()),
		SessionID:      sessionID,
		TemplateID:     templateID,
		ItemID:         itemID,
		RenderedPrompt: "Qu'est-ce que la densité ?",
		ExpectedAnswer: mustJSON(map[string]string{"answer": "rapport masse/volume"}),
		GradingPolicy:  "KEYWORDS",
		CreatedAt:      now,
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO questions (id, session_id, template_id, item_id, rendered_prompt, expected_answer, grading_policy, times_seen, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8)`,
		q.ID, q.SessionID, q.TemplateID, q.ItemID, q.RenderedPrompt, q.ExpectedAnswer, q.GradingPolicy, q.CreatedAt)
	if err != nil {
		tdb.t.Fatalf("seedQuestion: %v", err)
	}
	return q
}

// seedValidationTask inserts a validation task.
func (tdb *testDB) seedValidationTask(itemID uuid.UUID, status validation.TaskStatus) *validation.ValidationTask {
	tdb.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	task := &validation.ValidationTask{
		ID:        uuid.Must(uuid.NewV7()),
		ItemID:    itemID,
		Priority:  5,
		Status:    status,
		Source:    validation.SourceUncertainty,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := tdb.pool.Exec(context.Background(),
		`INSERT INTO validation_tasks (id, item_id, priority, status, source, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		task.ID, task.ItemID, task.Priority, task.Status, task.Source, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		tdb.t.Fatalf("seedValidationTask: %v", err)
	}
	return task
}

// newMasteryFixture creates a mastery entity for testing.
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

func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
