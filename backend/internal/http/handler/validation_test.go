package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/validation"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// --- Mock validation repo ---

type mockValidationRepo struct {
	tasks []*validation.ValidationTask
}

func (m *mockValidationRepo) FindByID(_ context.Context, id uuid.UUID) (*validation.ValidationTask, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, validation.ErrNotFound
}

func (m *mockValidationRepo) FindPendingByItem(_ context.Context, _ uuid.UUID) ([]*validation.ValidationTask, error) {
	return nil, nil
}

func (m *mockValidationRepo) FindPendingAll(_ context.Context, _ int) ([]*validation.ValidationTask, error) {
	var pending []*validation.ValidationTask
	for _, t := range m.tasks {
		if t.Status == validation.StatusPending {
			pending = append(pending, t)
		}
	}
	return pending, nil
}

func (m *mockValidationRepo) Save(_ context.Context, t *validation.ValidationTask) error {
	for i, existing := range m.tasks {
		if existing.ID == t.ID {
			m.tasks[i] = t
			return nil
		}
	}
	m.tasks = append(m.tasks, t)
	return nil
}

func setupValidationRouter(t *testing.T, valRepo validation.Repository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	chRepo := &mockChapterRepoWithItem{}
	svc := app.NewValidationService(valRepo, chRepo, &stubPublisher{}, stubClock{now: time.Now()}, stubIDGen{})
	h := handler.NewValidation(svc, chRepo)

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, testUserID)
		c.Next()
	})
	api.GET("/validations", h.ListPending)
	api.POST("/validations/:validation_id/resolve", h.Resolve)
	return r
}

func TestValidation_ListPending_Empty(t *testing.T) {
	valRepo := &mockValidationRepo{}
	r := setupValidationRouter(t, valRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/validations", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result []dto.ValidationTaskResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(result))
	}
}

func TestValidation_ListPending_WithTasks(t *testing.T) {
	taskID := uuid.Must(uuid.NewV7())
	valRepo := &mockValidationRepo{
		tasks: []*validation.ValidationTask{
			{
				ID:        taskID,
				ItemID:    uuid.Must(uuid.NewV7()),
				Priority:  1,
				Status:    validation.StatusPending,
				Source:    validation.SourceUncertainty,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
	r := setupValidationRouter(t, valRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/validations", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []dto.ValidationTaskResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}
	if result[0].Status != "PENDING" {
		t.Errorf("expected status=PENDING, got %q", result[0].Status)
	}
}

func TestValidation_Resolve_Confirm(t *testing.T) {
	taskID := uuid.Must(uuid.NewV7())
	valRepo := &mockValidationRepo{
		tasks: []*validation.ValidationTask{
			{
				ID:        taskID,
				ItemID:    uuid.Must(uuid.NewV7()),
				Priority:  1,
				Status:    validation.StatusPending,
				Source:    validation.SourceUncertainty,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
	r := setupValidationRouter(t, valRepo)

	body, _ := json.Marshal(dto.ResolveValidationRequest{Action: "confirm"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validations/"+taskID.String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidation_Resolve_InvalidAction(t *testing.T) {
	taskID := uuid.Must(uuid.NewV7())
	valRepo := &mockValidationRepo{
		tasks: []*validation.ValidationTask{
			{ID: taskID, Status: validation.StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	r := setupValidationRouter(t, valRepo)

	body, _ := json.Marshal(map[string]string{"action": "invalid"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validations/"+taskID.String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidation_Resolve_NotFound(t *testing.T) {
	valRepo := &mockValidationRepo{}
	r := setupValidationRouter(t, valRepo)

	body, _ := json.Marshal(dto.ResolveValidationRequest{Action: "confirm"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validations/"+uuid.Must(uuid.NewV7()).String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidation_Resolve_InvalidTaskID(t *testing.T) {
	valRepo := &mockValidationRepo{}
	r := setupValidationRouter(t, valRepo)

	body, _ := json.Marshal(dto.ResolveValidationRequest{Action: "confirm"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validations/not-a-uuid/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestValidation_Resolve_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	valRepo := &mockValidationRepo{}
	chRepo := &mockChapterRepoWithItem{}
	svc := app.NewValidationService(valRepo, chRepo, &stubPublisher{}, stubClock{now: time.Now()}, stubIDGen{})
	h := handler.NewValidation(svc, chRepo)

	r := gin.New()
	r.POST("/api/v1/validations/:validation_id/resolve", h.Resolve) // no auth

	body, _ := json.Marshal(dto.ResolveValidationRequest{Action: "confirm"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validations/"+uuid.Must(uuid.NewV7()).String()+"/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
