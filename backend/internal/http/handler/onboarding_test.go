package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/popul/revisemieux/internal/http/handler"
)

func TestOnboarding_GetStatus_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewOnboarding(nil)

	r := gin.New()
	r.GET("/api/v1/onboarding/status", h.GetStatus) // no auth

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestOnboarding_SeedDemo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewOnboarding(nil)

	r := gin.New()
	r.POST("/api/v1/onboarding/seed-demo", h.SeedDemo) // no auth

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/seed-demo", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
