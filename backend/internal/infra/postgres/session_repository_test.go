//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/session"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func TestSessionRepository_SaveAndFindByID(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")

	now := time.Now().UTC().Truncate(time.Microsecond)
	s := &session.Session{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		SessionType: session.TypeDaily, Status: session.StatusInProgress,
		TriggerType: session.TriggerManual,
		CreatedAt:   now, UpdatedAt: now,
	}

	if err := repo.Save(ctx, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.SessionType != session.TypeDaily {
		t.Errorf("SessionType = %v, want daily", got.SessionType)
	}
	if got.Status != session.StatusInProgress {
		t.Errorf("Status = %v, want IN_PROGRESS", got.Status)
	}
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSessionRepository_FindActiveByUser(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")

	// Create completed session
	completed := tdb.seedSession(userID, session.TypeDaily, session.StatusCompleted)
	// Create active session
	active := tdb.seedSession(userID, session.TypeDaily, session.StatusInProgress)
	_ = completed

	got, err := repo.FindActiveByUser(ctx, userID)
	if err != nil {
		t.Fatalf("FindActiveByUser: %v", err)
	}
	if got.ID != active.ID {
		t.Errorf("ID = %v, want %v", got.ID, active.ID)
	}
}

func TestSessionRepository_FindActiveByUser_NoActive(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")

	_, err := repo.FindActiveByUser(ctx, userID)
	if err == nil {
		t.Fatal("expected error for no active session, got nil")
	}
}

func TestSessionRepository_SessionWithChapterLinks(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch1 := tdb.seedChapter(userID)
	ch2 := tdb.seedChapter(userID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	s := &session.Session{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		SessionType: session.TypeMockExam, Status: session.StatusComposing,
		TriggerType: session.TriggerManual,
		ChapterIDs:  []uuid.UUID{ch1.ID, ch2.ID},
		CreatedAt:   now, UpdatedAt: now,
	}

	if err := repo.Save(ctx, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, s.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if len(got.ChapterIDs) != 2 {
		t.Fatalf("ChapterIDs count = %d, want 2", len(got.ChapterIDs))
	}
}

func TestSessionRepository_SaveUpsert(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")

	s := tdb.seedSession(userID, session.TypeDaily, session.StatusComposing)

	// Fetch and update
	got, _ := repo.FindByID(ctx, s.ID)
	got.Status = session.StatusInProgress
	now := time.Now().UTC().Truncate(time.Microsecond)
	got.StartedAt = &now
	got.UpdatedAt = now

	if err := repo.Save(ctx, got); err != nil {
		t.Fatalf("Save (update): %v", err)
	}

	updated, _ := repo.FindByID(ctx, s.ID)
	if updated.Status != session.StatusInProgress {
		t.Errorf("Status = %v, want IN_PROGRESS", updated.Status)
	}
}

func TestSessionRepository_QuestionCRUD(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, "KNOWLEDGE", "Densité")
	tdb.seedTemplate("GEN.KNOW.FLASH_MCQ")
	sess := tdb.seedSession(userID, session.TypeDaily, session.StatusInProgress)

	now := time.Now().UTC().Truncate(time.Microsecond)
	q := &session.Question{
		ID: uuid.Must(uuid.NewV7()), SessionID: sess.ID,
		TemplateID: "GEN.KNOW.FLASH_MCQ", ItemID: item.ID,
		RenderedPrompt: "Qu'est-ce que la densité ?",
		ExpectedAnswer: mustJSON(map[string]string{"answer": "rapport"}),
		GradingPolicy:  "KEYWORDS",
		CreatedAt:      now,
	}

	if err := repo.SaveQuestion(ctx, q); err != nil {
		t.Fatalf("SaveQuestion: %v", err)
	}

	gotQ, err := repo.FindQuestionByID(ctx, q.ID)
	if err != nil {
		t.Fatalf("FindQuestionByID: %v", err)
	}
	if gotQ.RenderedPrompt != "Qu'est-ce que la densité ?" {
		t.Errorf("RenderedPrompt = %q", gotQ.RenderedPrompt)
	}

	questions, err := repo.FindQuestionsBySession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("FindQuestionsBySession: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("questions count = %d, want 1", len(questions))
	}
}

func TestSessionRepository_AttemptCRUD(t *testing.T) {
	tdb := setupTestDB(t)
	repo := postgres.NewSessionRepository(tdb.pool)
	ctx := context.Background()
	userID := tdb.seedUser("student")
	ch := tdb.seedChapter(userID)
	rev := tdb.seedRevision(ch.ID)
	item := tdb.seedItem(ch.ID, rev.ID, nil, "KNOWLEDGE", "Densité")
	tdb.seedTemplate("GEN.KNOW.FLASH_MCQ")
	sess := tdb.seedSession(userID, session.TypeDaily, session.StatusInProgress)
	q := tdb.seedQuestion(sess.ID, item.ID, "GEN.KNOW.FLASH_MCQ")

	now := time.Now().UTC().Truncate(time.Microsecond)
	feedback := "Bonne réponse !"
	a := &session.Attempt{
		ID: uuid.Must(uuid.NewV7()), SessionID: sess.ID,
		QuestionID: q.ID, UserID: userID,
		Answer: mustJSON("rapport masse/volume"),
		Score:  1.0, Feedback: &feedback,
		Source: session.AttemptInteractive, CreatedAt: now,
	}

	if err := repo.SaveAttempt(ctx, a); err != nil {
		t.Fatalf("SaveAttempt: %v", err)
	}

	attempts, err := repo.FindAttemptsBySession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("FindAttemptsBySession: %v", err)
	}
	if len(attempts) != 1 {
		t.Fatalf("count = %d, want 1", len(attempts))
	}
	if attempts[0].Score != 1.0 {
		t.Errorf("Score = %v, want 1.0", attempts[0].Score)
	}
	if *attempts[0].Feedback != "Bonne réponse !" {
		t.Errorf("Feedback = %q", *attempts[0].Feedback)
	}
}
