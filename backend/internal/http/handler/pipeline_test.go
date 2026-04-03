package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/http/handler"
)

func TestPipeline_Upload_InvalidChapterID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewPipeline(nil)

	r := gin.New()
	r.POST("/api/v1/chapters/:chapter_id/upload", h.Upload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chapters/not-a-uuid/upload", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPipeline_Upload_NoMultipartForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewPipeline(nil)

	r := gin.New()
	r.POST("/api/v1/chapters/:chapter_id/upload", h.Upload)

	chapterID := uuid.Must(uuid.NewV7()).String()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chapters/"+chapterID+"/upload", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPipeline_GetProgress_InvalidRevisionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewPipeline(nil)

	r := gin.New()
	r.GET("/api/v1/revisions/:revision_id/progress", h.GetProgress)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/revisions/not-a-uuid/progress", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
