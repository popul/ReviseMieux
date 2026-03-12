//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/db"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/validation"
	apphttp "github.com/popul/revisemieux/internal/http"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
	"github.com/popul/revisemieux/internal/infra/eventbus"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

const jwtSecret = "test-secret-for-integration-tests"

// testApp holds the full wired application for HTTP integration tests.
type testApp struct {
	pool        *pgxpool.Pool
	router      *gin.Engine
	chapterRepo *postgres.ChapterRepository
	t           *testing.T
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := db.Migrate(context.Background(), pool, migrationsDir()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Repos
	chapterRepo := postgres.NewChapterRepository(pool)
	masteryRepo := postgres.NewMasteryRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	valRepo := postgres.NewValidationRepository(pool)
	publisher := eventbus.NewLogPublisher()
	clock := event.RealClock{}
	idGen := event.UUIDv7Generator{}

	// Services
	chapterSvc := app.NewChapterService(chapterRepo, masteryRepo)
	masterySvc := app.NewMasteryService(masteryRepo, publisher, clock)
	sessionSvc := app.NewSessionService(sessionRepo, chapterRepo, masteryRepo, publisher, clock, idGen)
	valSvc := app.NewValidationService(valRepo, chapterRepo, publisher, clock, idGen)
	onboardingSvc := app.NewOnboardingService(chapterRepo, masteryRepo, clock, idGen)

	// Handlers
	chapterHandler := handler.NewChapter(chapterSvc, chapterRepo, idGen, clock)
	masteryHandler := handler.NewMastery(masterySvc)
	sessionHandler := handler.NewSession(sessionSvc)
	valHandler := handler.NewValidation(valSvc)
	onboardingHandler := handler.NewOnboarding(onboardingSvc)

	// Router
	router := apphttp.NewRouter(apphttp.RouterConfig{
		JWTSecret:         jwtSecret,
		Version:           "test",
		ChapterHandler:    chapterHandler,
		MasteryHandler:    masteryHandler,
		SessionHandler:    sessionHandler,
		ValidationHandler: valHandler,
		OnboardingHandler: onboardingHandler,
	})

	ta := &testApp{pool: pool, router: router, chapterRepo: chapterRepo, t: t}
	// Truncate at the START to ensure clean state — avoids races with t.Cleanup from previous tests.
	ta.truncateAll()
	t.Cleanup(func() { ta.truncateAll() })
	return ta
}

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

func (ta *testApp) truncateAll() {
	ctx := context.Background()
	_, _ = ta.pool.Exec(ctx, `TRUNCATE users, exams, templates CASCADE`)
}

// seedUser creates a user and returns the ID and a valid JWT token.
func (ta *testApp) seedUser(role string) (uuid.UUID, string) {
	ta.t.Helper()
	id := uuid.Must(uuid.NewV7())
	_, err := ta.pool.Exec(context.Background(),
		`INSERT INTO users (id, role, display_name) VALUES ($1, $2, $3)`,
		id, role, "Test User")
	if err != nil {
		ta.t.Fatalf("seedUser: %v", err)
	}
	token, err := middleware.GenerateToken(jwtSecret, id, role, 24*time.Hour, time.Now())
	if err != nil {
		ta.t.Fatalf("generateToken: %v", err)
	}
	return id, token
}

func (ta *testApp) seedChapter(userID uuid.UUID) *chapter.Chapter {
	ta.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	ch := &chapter.Chapter{
		ID: uuid.Must(uuid.NewV7()), UserID: userID,
		Subject: "Physique-Chimie", ClassLevel: "4e",
		Name: "Densité", CreatedAt: now, UpdatedAt: now,
	}
	_, err := ta.pool.Exec(context.Background(),
		`INSERT INTO chapters (id, user_id, subject, class_level, name, archived, is_demo, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,false,false,$6,$7)`,
		ch.ID, ch.UserID, ch.Subject, ch.ClassLevel, ch.Name, ch.CreatedAt, ch.UpdatedAt)
	if err != nil {
		ta.t.Fatalf("seedChapter: %v", err)
	}
	return ch
}

func (ta *testApp) seedRevision(chapterID uuid.UUID) *chapter.Revision {
	ta.t.Helper()
	rev := &chapter.Revision{
		ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID,
		RevisionNumber: 1, Status: chapter.RevisionReady,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	_, err := ta.pool.Exec(context.Background(),
		`INSERT INTO chapter_revisions (id, chapter_id, revision_number, status, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		rev.ID, rev.ChapterID, rev.RevisionNumber, rev.Status, rev.CreatedAt)
	if err != nil {
		ta.t.Fatalf("seedRevision: %v", err)
	}
	_, _ = ta.pool.Exec(context.Background(),
		`UPDATE chapters SET current_revision_id = $1 WHERE id = $2`, rev.ID, chapterID)
	return rev
}

func (ta *testApp) seedItem(chapterID, revisionID uuid.UUID, term string) *chapter.Item {
	ta.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	item := &chapter.Item{
		ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID,
		RevisionID: revisionID, ItemType: chapter.ItemKnowledge,
		Term: &term, Confidence: 0.9,
		CreatedAt: now, UpdatedAt: now,
	}
	_, err := ta.pool.Exec(context.Background(),
		`INSERT INTO items (id, chapter_id, revision_id, item_type, term,
		                    confidence, validation_required, archived, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,false,false,$7,$8)`,
		item.ID, item.ChapterID, item.RevisionID, item.ItemType, item.Term,
		item.Confidence, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		ta.t.Fatalf("seedItem: %v", err)
	}
	return item
}

func (ta *testApp) seedValidationTask(itemID uuid.UUID) *validation.ValidationTask {
	ta.t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	task := &validation.ValidationTask{
		ID: uuid.Must(uuid.NewV7()), ItemID: itemID,
		Priority: 5, Status: validation.StatusPending,
		Source: validation.SourceUncertainty,
		CreatedAt: now, UpdatedAt: now,
	}
	_, err := ta.pool.Exec(context.Background(),
		`INSERT INTO validation_tasks (id, item_id, priority, status, source, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		task.ID, task.ItemID, task.Priority, task.Status, task.Source, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		ta.t.Fatalf("seedValidationTask: %v", err)
	}
	return task
}

func doRequest(router *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ============================================================
// Tests
// ============================================================

func TestHealth(t *testing.T) {
	ta := setupTestApp(t)
	w := doRequest(ta.router, "GET", "/health", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	ta := setupTestApp(t)
	w := doRequest(ta.router, "GET", "/api/v1/chapters", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestChapter_CreateAndList(t *testing.T) {
	ta := setupTestApp(t)
	_, token := ta.seedUser("student")

	// Create
	body := map[string]string{
		"subject":     "Mathématiques",
		"class_level": "3e",
		"name":        "Pythagore",
	}
	w := doRequest(ta.router, "POST", "/api/v1/chapters", token, body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	if created["name"] != "Pythagore" {
		t.Errorf("name = %v, want Pythagore", created["name"])
	}

	// List
	w = doRequest(ta.router, "GET", "/api/v1/chapters", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	var chapters []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &chapters)
	if len(chapters) != 1 {
		t.Fatalf("count = %d, want 1", len(chapters))
	}
}

func TestChapter_CreateValidation(t *testing.T) {
	ta := setupTestApp(t)
	_, token := ta.seedUser("student")

	// Missing required field
	body := map[string]string{"subject": "Maths"}
	w := doRequest(ta.router, "POST", "/api/v1/chapters", token, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestChapter_LessonCard(t *testing.T) {
	ta := setupTestApp(t)
	userID, token := ta.seedUser("student")
	ch := ta.seedChapter(userID)
	rev := ta.seedRevision(ch.ID)
	ta.seedItem(ch.ID, rev.ID, "Densité")
	ta.seedItem(ch.ID, rev.ID, "Masse volumique")

	w := doRequest(ta.router, "GET", "/api/v1/chapters/"+ch.ID.String()+"/lesson-card", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var card map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &card)
	items, ok := card["items"].([]interface{})
	if !ok {
		t.Fatal("items not an array")
	}
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}
}

func TestChapter_LessonCard_NotFound(t *testing.T) {
	ta := setupTestApp(t)
	_, token := ta.seedUser("student")

	w := doRequest(ta.router, "GET", "/api/v1/chapters/"+uuid.New().String()+"/lesson-card", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestOnboarding_Status(t *testing.T) {
	ta := setupTestApp(t)
	_, token := ta.seedUser("student")

	w := doRequest(ta.router, "GET", "/api/v1/onboarding/status", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var status map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &status)
	if status["account_created"] != true {
		t.Errorf("account_created = %v, want true", status["account_created"])
	}
}

func TestOnboarding_SeedDemo(t *testing.T) {
	ta := setupTestApp(t)
	_, token := ta.seedUser("student")

	w := doRequest(ta.router, "POST", "/api/v1/onboarding/seed-demo", token, nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["item_count"].(float64) != 8 {
		t.Errorf("item_count = %v, want 8", resp["item_count"])
	}

	// Idempotent — second call should also succeed
	w2 := doRequest(ta.router, "POST", "/api/v1/onboarding/seed-demo", token, nil)
	if w2.Code != http.StatusCreated {
		t.Fatalf("second seed status = %d", w2.Code)
	}
}

func TestValidation_ListPending(t *testing.T) {
	ta := setupTestApp(t)
	userID, token := ta.seedUser("student")
	ch := ta.seedChapter(userID)
	rev := ta.seedRevision(ch.ID)
	item := ta.seedItem(ch.ID, rev.ID, "Densité")
	ta.seedValidationTask(item.ID)
	ta.seedValidationTask(item.ID)

	w := doRequest(ta.router, "GET", "/api/v1/validations", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var tasks []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &tasks)
	if len(tasks) != 2 {
		t.Fatalf("tasks count = %d, want 2", len(tasks))
	}
}

func TestValidation_Resolve_Confirm(t *testing.T) {
	ta := setupTestApp(t)
	userID, token := ta.seedUser("parent")
	ch := ta.seedChapter(userID)
	rev := ta.seedRevision(ch.ID)
	item := ta.seedItem(ch.ID, rev.ID, "Densité")
	task := ta.seedValidationTask(item.ID)

	body := map[string]string{"action": "confirm"}
	w := doRequest(ta.router, "POST", "/api/v1/validations/"+task.ID.String()+"/resolve", token, body)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	// Verify item validation_required is now false and confidence >= 0.85
	gotItem, _ := ta.chapterRepo.FindItemByID(context.Background(), item.ID)
	if gotItem.ValidationRequired {
		t.Error("ValidationRequired should be false after confirm")
	}
	if gotItem.Confidence < 0.85 {
		t.Errorf("Confidence = %v, should be >= 0.85", gotItem.Confidence)
	}
}

func TestValidation_Resolve_AlreadyResolved(t *testing.T) {
	ta := setupTestApp(t)
	userID, token := ta.seedUser("parent")
	ch := ta.seedChapter(userID)
	rev := ta.seedRevision(ch.ID)
	item := ta.seedItem(ch.ID, rev.ID, "Densité")
	task := ta.seedValidationTask(item.ID)

	// Confirm first time
	body := map[string]string{"action": "confirm"}
	doRequest(ta.router, "POST", "/api/v1/validations/"+task.ID.String()+"/resolve", token, body)

	// Try again → 409
	w := doRequest(ta.router, "POST", "/api/v1/validations/"+task.ID.String()+"/resolve", token, body)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

func TestValidation_Resolve_InvalidAction(t *testing.T) {
	ta := setupTestApp(t)
	userID, token := ta.seedUser("parent")
	ch := ta.seedChapter(userID)
	rev := ta.seedRevision(ch.ID)
	item := ta.seedItem(ch.ID, rev.ID, "Densité")
	task := ta.seedValidationTask(item.ID)

	body := map[string]string{"action": "invalid"}
	w := doRequest(ta.router, "POST", "/api/v1/validations/"+task.ID.String()+"/resolve", token, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
