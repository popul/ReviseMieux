package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/config"
)

func setupTestRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &Handlers{
		configFrontend: &ConfigFrontend{NombreMaxPages: cfg.NombreMaxPages},
	}
	r.GET("/api/config", h.ConfigFrontendHandler)
	return r
}

func TestConfigFrontendHandler(t *testing.T) {
	cfg := &config.Config{NombreMaxPages: 30}
	r := setupTestRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("statut = %d, attendu %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Succes bool           `json:"succes"`
		Config ConfigFrontend `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("erreur parsing JSON: %v", err)
	}

	if !resp.Succes {
		t.Error("succes devrait être true")
	}
	if resp.Config.NombreMaxPages != 30 {
		t.Errorf("nombreMaxPages = %d, attendu 30", resp.Config.NombreMaxPages)
	}
}

func TestConfigFrontendHandler_ValeurPerso(t *testing.T) {
	cfg := &config.Config{NombreMaxPages: 50}
	r := setupTestRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("statut = %d, attendu %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Succes bool           `json:"succes"`
		Config ConfigFrontend `json:"config"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("erreur parsing JSON: %v", err)
	}

	if resp.Config.NombreMaxPages != 50 {
		t.Errorf("nombreMaxPages = %d, attendu 50", resp.Config.NombreMaxPages)
	}
}
