package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/handler"
)

func TestHealthCheck_Returns200WithStatusOk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHealth("1.2.3")

	r := gin.New()
	r.GET("/health", h.Check)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp dto.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status=ok, got %q", resp.Status)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected version=1.2.3, got %q", resp.Version)
	}
}

func TestHealthCheck_EmptyVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHealth("")

	r := gin.New()
	r.GET("/health", h.Check)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp dto.HealthResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "ok" {
		t.Errorf("expected status=ok, got %q", resp.Status)
	}
}
