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
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
)

func setupMasteryRouter(t *testing.T, mRepo mastery.Repository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svc := app.NewMasteryService(mRepo, &stubPublisher{}, stubClock{now: time.Now()})
	h := handler.NewMastery(svc)

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, testUserID)
		c.Next()
	})
	api.GET("/masteries", h.GetByUser)
	api.GET("/masteries/:item_id", h.GetByItem)
	api.POST("/masteries/attempt", h.RecordAttempt)
	return r
}

func TestMastery_GetByUser_ValidState(t *testing.T) {
	itemID := uuid.Must(uuid.NewV7())
	mRepo := &mockMasteryRepo{
		masteries: []*mastery.Mastery{
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: itemID, State: mastery.Fragile},
		},
	}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries?state=FRAGILE", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result []dto.MasteryResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 1 {
		t.Fatalf("expected 1 mastery, got %d", len(result))
	}
	if result[0].State != "FRAGILE" {
		t.Errorf("expected state=FRAGILE, got %q", result[0].State)
	}
}

func TestMastery_GetByUser_EmptyResult(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries?state=SOLID", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []dto.MasteryResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected 0 masteries, got %d", len(result))
	}
}

func TestMastery_GetByUser_InvalidState(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries?state=INVALID", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMastery_GetByUser_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mRepo := &mockMasteryRepo{}
	svc := app.NewMasteryService(mRepo, &stubPublisher{}, stubClock{now: time.Now()})
	h := handler.NewMastery(svc)

	r := gin.New()
	r.GET("/api/v1/masteries", h.GetByUser) // no auth middleware
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries?state=UNKNOWN", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMastery_GetByItem_Found(t *testing.T) {
	itemID := uuid.Must(uuid.NewV7())
	mRepo := &mockMasteryRepo{
		masteries: []*mastery.Mastery{
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: itemID, State: mastery.OK},
		},
	}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries/"+itemID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result dto.MasteryResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.State != "OK" {
		t.Errorf("expected state=OK, got %q", result.State)
	}
}

func TestMastery_GetByItem_NotFound(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries/"+uuid.Must(uuid.NewV7()).String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMastery_GetByItem_InvalidUUID(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/masteries/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMastery_RecordAttempt_Success(t *testing.T) {
	itemID := uuid.Must(uuid.NewV7())
	mRepo := &mockMasteryRepo{
		masteries: []*mastery.Mastery{
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: itemID, State: mastery.Unknown},
		},
	}
	r := setupMasteryRouter(t, mRepo)

	body, _ := json.Marshal(dto.RecordAttemptRequest{ItemID: itemID.String(), Score: 1.0})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/masteries/attempt", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result dto.RecordAttemptResponse
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.NewState != "FRAGILE" {
		t.Errorf("expected new_state=FRAGILE, got %q", result.NewState)
	}
}

func TestMastery_RecordAttempt_InvalidBody(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/masteries/attempt", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMastery_RecordAttempt_InvalidItemID(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	body, _ := json.Marshal(map[string]interface{}{"item_id": "not-a-uuid", "score": 0.5})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/masteries/attempt", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMastery_RecordAttempt_NotFound(t *testing.T) {
	mRepo := &mockMasteryRepo{}
	r := setupMasteryRouter(t, mRepo)

	body, _ := json.Marshal(dto.RecordAttemptRequest{ItemID: uuid.Must(uuid.NewV7()).String(), Score: 0.5})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/masteries/attempt", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
