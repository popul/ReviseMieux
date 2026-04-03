package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
)

func setupSessionRouterNoService(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Pass nil service - we only test input validation paths
	h := handler.NewSession(nil)

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, testUserID)
		c.Next()
	})
	api.POST("/sessions/daily", h.ComposeDaily)
	api.POST("/sessions/:session_id/resume", h.Resume)
	api.GET("/sessions/:session_id", h.GetByID)
	api.GET("/sessions/:session_id/questions", h.GetQuestions)
	api.POST("/sessions/:session_id/answer", h.SubmitAnswer)
	api.GET("/sessions/:session_id/debrief", h.Debrief)
	return r
}

func TestSession_ComposeDaily_InvalidBody(t *testing.T) {
	r := setupSessionRouterNoService(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/daily", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSession_ComposeDaily_InvalidChapterID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	body, _ := json.Marshal(dto.ComposeDailyRequest{ChapterID: "not-a-uuid"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/daily", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_ComposeDaily_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewSession(nil)
	r := gin.New()
	r.POST("/api/v1/sessions/daily", h.ComposeDaily)

	body, _ := json.Marshal(dto.ComposeDailyRequest{ChapterID: uuid.Must(uuid.NewV7()).String()})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/daily", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestSession_Resume_InvalidID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/not-a-uuid/resume", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_GetByID_InvalidID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_GetQuestions_InvalidID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/not-a-uuid/questions", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_SubmitAnswer_InvalidSessionID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	body, _ := json.Marshal(dto.SubmitAnswerRequest{QuestionID: uuid.Must(uuid.NewV7()).String(), Answer: "test"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/not-a-uuid/answer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_SubmitAnswer_InvalidBody(t *testing.T) {
	r := setupSessionRouterNoService(t)
	_ = time.Now() // silence unused import

	sessionID := uuid.Must(uuid.NewV7()).String()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/answer", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSession_SubmitAnswer_InvalidQuestionID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	body, _ := json.Marshal(dto.SubmitAnswerRequest{QuestionID: "not-a-uuid", Answer: "test"})
	sessionID := uuid.Must(uuid.NewV7()).String()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/answer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSession_SubmitAnswer_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewSession(nil)
	r := gin.New()
	r.POST("/api/v1/sessions/:session_id/answer", h.SubmitAnswer)

	body, _ := json.Marshal(dto.SubmitAnswerRequest{QuestionID: uuid.Must(uuid.NewV7()).String(), Answer: "test"})
	sessionID := uuid.Must(uuid.NewV7()).String()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/answer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestSession_Debrief_InvalidID(t *testing.T) {
	r := setupSessionRouterNoService(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/not-a-uuid/debrief", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
