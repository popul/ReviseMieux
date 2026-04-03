package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
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

func TestOnboarding_GetStatus_WithAuth_ButNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewOnboarding(nil)

	r := gin.New()
	r.Use(gin.Recovery()) // recover from nil pointer panic
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uuid.Must(uuid.NewV7()))
		c.Next()
	})
	api.GET("/onboarding/status", h.GetStatus)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/status", nil)
	r.ServeHTTP(w, req)

	// With nil service, gin.Recovery catches the nil pointer panic and returns 500
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 (nil service panic), got %d", w.Code)
	}
}
