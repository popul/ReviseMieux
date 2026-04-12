package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/popul/revisemieux/internal/http/handler"
)

func TestDevToken_Returns200WithValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewDev("test-secret", nil, nil, "", nil)

	r := gin.New()
	r.GET("/dev/token", h.Token)

	req := httptest.NewRequest(http.MethodGet, "/dev/token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Check JSON content type
	ct := w.Header().Get("Content-Type")
	if ct == "" {
		t.Fatal("expected Content-Type header to be set")
	}
	if ct != "application/json; charset=utf-8" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}

	var body struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Token must be non-empty
	if body.Token == "" {
		t.Fatal("expected token to be non-empty")
	}

	// UserID must be a valid UUID
	if body.UserID == "" {
		t.Fatal("expected user_id to be non-empty")
	}
	if _, err := uuid.Parse(body.UserID); err != nil {
		t.Fatalf("expected user_id to be a valid UUID, got %q: %v", body.UserID, err)
	}
}

func TestDevToken_ReturnsDeterministicUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewDev("test-secret", nil, nil, "", nil)

	r := gin.New()
	r.GET("/dev/token", h.Token)

	// Call twice and ensure same user_id
	var userIDs [2]string
	for i := range userIDs {
		req := httptest.NewRequest(http.MethodGet, "/dev/token", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var body struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("call %d: decode error: %v", i, err)
		}
		userIDs[i] = body.UserID
	}

	if userIDs[0] != userIDs[1] {
		t.Fatalf("expected deterministic user_id, got %q and %q", userIDs[0], userIDs[1])
	}
}
