package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
)

// --- Mocks for session service tests ---

type mockSessionRepo struct {
	sessions  map[uuid.UUID]*session.Session
	questions map[uuid.UUID][]*session.Question
	attempts  map[uuid.UUID][]*session.Attempt
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:  make(map[uuid.UUID]*session.Session),
		questions: make(map[uuid.UUID][]*session.Question),
		attempts:  make(map[uuid.UUID][]*session.Attempt),
	}
}

func (m *mockSessionRepo) FindByID(_ context.Context, id uuid.UUID) (*session.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, session.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) FindActiveByUser(_ context.Context, _ uuid.UUID) (*session.Session, error) {
	return nil, session.ErrNotFound
}

func (m *mockSessionRepo) Save(_ context.Context, s *session.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) FindQuestionByID(_ context.Context, id uuid.UUID) (*session.Question, error) {
	for _, qs := range m.questions {
		for _, q := range qs {
			if q.ID == id {
				return q, nil
			}
		}
	}
	return nil, session.ErrNotFound
}

func (m *mockSessionRepo) FindQuestionsBySession(_ context.Context, sessionID uuid.UUID) ([]*session.Question, error) {
	return m.questions[sessionID], nil
}

func (m *mockSessionRepo) SaveQuestion(_ context.Context, q *session.Question) error {
	m.questions[q.SessionID] = append(m.questions[q.SessionID], q)
	return nil
}

func (m *mockSessionRepo) SaveAttempt(_ context.Context, a *session.Attempt) error {
	m.attempts[a.SessionID] = append(m.attempts[a.SessionID], a)
	return nil
}

func (m *mockSessionRepo) FindAttemptsBySession(_ context.Context, sessionID uuid.UUID) ([]*session.Attempt, error) {
	return m.attempts[sessionID], nil
}

type mockChapterRepoForSession struct{}

func (m *mockChapterRepoForSession) FindByID(_ context.Context, _ uuid.UUID) (*chapter.Chapter, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) FindByUser(_ context.Context, _ uuid.UUID, _ bool) ([]*chapter.Chapter, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) Save(_ context.Context, _ *chapter.Chapter) error { return nil }
func (m *mockChapterRepoForSession) FindItemByID(_ context.Context, _ uuid.UUID) (*chapter.Item, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) FindItemsByChapter(_ context.Context, _ uuid.UUID, _ bool) ([]*chapter.Item, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) SaveItem(_ context.Context, _ *chapter.Item) error   { return nil }
func (m *mockChapterRepoForSession) SaveItems(_ context.Context, _ []*chapter.Item) error { return nil }
func (m *mockChapterRepoForSession) FindNotionsByChapter(_ context.Context, _ uuid.UUID) ([]*chapter.Notion, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) SaveNotion(_ context.Context, _ *chapter.Notion) error {
	return nil
}
func (m *mockChapterRepoForSession) FindRevisionByID(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) FindCurrentRevision(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) SaveRevision(_ context.Context, _ *chapter.Revision) error {
	return nil
}
func (m *mockChapterRepoForSession) FindPagesByRevision(_ context.Context, _ uuid.UUID) ([]*chapter.Page, error) {
	return nil, nil
}
func (m *mockChapterRepoForSession) SavePage(_ context.Context, _ *chapter.Page) error { return nil }
func (m *mockChapterRepoForSession) CountRevisionsByChapter(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

type mockMasteryRepoForSession struct{}

func (m *mockMasteryRepoForSession) FindByID(_ context.Context, _ uuid.UUID) (*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepoForSession) FindByUserAndItem(_ context.Context, _, _ uuid.UUID) (*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepoForSession) FindDueByUser(_ context.Context, _ uuid.UUID, _ time.Time) ([]*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepoForSession) FindByUserAndState(_ context.Context, _ uuid.UUID, _ mastery.State) ([]*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepoForSession) Save(_ context.Context, _ *mastery.Mastery) error      { return nil }
func (m *mockMasteryRepoForSession) SaveAll(_ context.Context, _ []*mastery.Mastery) error { return nil }

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _ ...event.Event) error { return nil }

type stubIDGenForSession struct{ seq int }

func (s *stubIDGenForSession) New() uuid.UUID {
	s.seq++
	return uuid.Must(uuid.NewV7())
}

type stubClockForSession struct{ now time.Time }

func (c stubClockForSession) Now() time.Time { return c.now }

func TestGetDebrief_CalculatesScoreAndPercentage(t *testing.T) {
	now := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	sessionID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	sessRepo := newMockSessionRepo()
	sessRepo.sessions[sessionID] = &session.Session{
		ID:     sessionID,
		UserID: userID,
		Status: session.StatusCompleted,
	}

	// 3 attempts: 1.0, 0.5, 1.0 => score=2.5, total=3, percentage=83.33...
	idGen := &stubIDGenForSession{}
	sessRepo.attempts[sessionID] = []*session.Attempt{
		{ID: idGen.New(), SessionID: sessionID, QuestionID: uuid.Must(uuid.NewV7()), UserID: userID, Score: 1.0, CreatedAt: now},
		{ID: idGen.New(), SessionID: sessionID, QuestionID: uuid.Must(uuid.NewV7()), UserID: userID, Score: 0.5, CreatedAt: now},
		{ID: idGen.New(), SessionID: sessionID, QuestionID: uuid.Must(uuid.NewV7()), UserID: userID, Score: 1.0, CreatedAt: now},
	}

	svc := app.NewSessionService(
		sessRepo,
		&mockChapterRepoForSession{},
		&mockMasteryRepoForSession{},
		&mockPublisher{},
		stubClockForSession{now: now},
		idGen,
		nil,
	)

	result, err := svc.GetDebrief(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total=3, got %d", result.Total)
	}

	if result.Score != 2.5 {
		t.Errorf("expected score=2.5, got %f", result.Score)
	}

	expectedPct := (2.5 / 3.0) * 100
	if result.Percentage < expectedPct-0.01 || result.Percentage > expectedPct+0.01 {
		t.Errorf("expected percentage~%.2f, got %.2f", expectedPct, result.Percentage)
	}

	if result.Transitions == nil {
		t.Error("expected transitions to be non-nil (empty slice)")
	}
}

func TestGetDebrief_NoAttempts_ReturnsZeros(t *testing.T) {
	now := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	sessionID := uuid.Must(uuid.NewV7())

	sessRepo := newMockSessionRepo()
	sessRepo.sessions[sessionID] = &session.Session{
		ID:     sessionID,
		Status: session.StatusCompleted,
	}

	svc := app.NewSessionService(
		sessRepo,
		&mockChapterRepoForSession{},
		&mockMasteryRepoForSession{},
		&mockPublisher{},
		stubClockForSession{now: now},
		&stubIDGenForSession{},
		nil,
	)

	result, err := svc.GetDebrief(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected total=0, got %d", result.Total)
	}
	if result.Score != 0 {
		t.Errorf("expected score=0, got %f", result.Score)
	}
	if result.Percentage != 0 {
		t.Errorf("expected percentage=0, got %f", result.Percentage)
	}
}

func TestGetDebrief_SessionNotFound_ReturnsError(t *testing.T) {
	now := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	sessRepo := newMockSessionRepo()

	svc := app.NewSessionService(
		sessRepo,
		&mockChapterRepoForSession{},
		&mockMasteryRepoForSession{},
		&mockPublisher{},
		stubClockForSession{now: now},
		&stubIDGenForSession{},
		nil,
	)

	_, err := svc.GetDebrief(context.Background(), uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
}

func TestGetDebrief_AllPerfectScores(t *testing.T) {
	now := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	sessionID := uuid.Must(uuid.NewV7())
	idGen := &stubIDGenForSession{}

	sessRepo := newMockSessionRepo()
	sessRepo.sessions[sessionID] = &session.Session{
		ID:     sessionID,
		Status: session.StatusCompleted,
	}
	sessRepo.attempts[sessionID] = []*session.Attempt{
		{ID: idGen.New(), SessionID: sessionID, Score: 1.0, CreatedAt: now},
		{ID: idGen.New(), SessionID: sessionID, Score: 1.0, CreatedAt: now},
	}

	svc := app.NewSessionService(
		sessRepo,
		&mockChapterRepoForSession{},
		&mockMasteryRepoForSession{},
		&mockPublisher{},
		stubClockForSession{now: now},
		idGen,
		nil,
	)

	result, err := svc.GetDebrief(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Percentage != 100 {
		t.Errorf("expected percentage=100, got %f", result.Percentage)
	}
}
