package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// SwitchableWriter is an io.Writer whose target can be swapped at runtime.
// Used to redirect Gin's logger output after the router is created.
type SwitchableWriter struct {
	mu sync.RWMutex
	w  io.Writer
}

// NewSwitchableWriter creates a SwitchableWriter with the given initial target.
func NewSwitchableWriter(w io.Writer) *SwitchableWriter {
	return &SwitchableWriter{w: w}
}

func (sw *SwitchableWriter) Write(p []byte) (n int, err error) {
	sw.mu.RLock()
	w := sw.w
	sw.mu.RUnlock()
	return w.Write(p)
}

// Switch changes the underlying writer.
func (sw *SwitchableWriter) Switch(w io.Writer) {
	sw.mu.Lock()
	sw.w = w
	sw.mu.Unlock()
}

// Dev handles development-only endpoints.
type Dev struct {
	jwtSecret     string
	pool          *pgxpool.Pool
	onboardingSvc *app.OnboardingService
	migrationsDir string
	logWriter     *SwitchableWriter

	logMu   sync.Mutex
	logFile *os.File
}

// NewDev creates a new Dev handler with the given JWT secret and DB pool.
func NewDev(jwtSecret string, pool *pgxpool.Pool, onboardingSvc *app.OnboardingService, migrationsDir string, logWriter *SwitchableWriter) *Dev {
	return &Dev{jwtSecret: jwtSecret, pool: pool, onboardingSvc: onboardingSvc, migrationsDir: migrationsDir, logWriter: logWriter}
}

// Token returns a JWT for a fixed dev user, creating the user in DB if needed.
// Only available when the server runs in debug mode.
func (h *Dev) Token(c *gin.Context) {
	devUserID := uuid.MustParse("00000000-0000-7000-8000-000000000001")

	// Ensure user exists in DB
	if h.pool != nil {
		h.ensureUser(c.Request.Context(), devUserID)
	}

	token, err := middleware.GenerateToken(h.jwtSecret, devUserID, "student", 24*time.Hour, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, dto.DevTokenResponse{
		Token:  token,
		UserID: devUserID.String(),
	})
}

// SeedScenario resets the database and loads a named E2E scenario.
// POST /e2e/seed/:scenario — only available in debug mode.
// Optional JSON body: {"log_path": "/absolute/path/to/backend.log"}
func (h *Dev) SeedScenario(c *gin.Context) {
	scenario := c.Param("scenario")

	// Handle optional log redirection
	var body struct {
		LogPath string `json:"log_path"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.LogPath != "" {
		if err := h.startLogCapture(body.LogPath); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("log capture failed: %v", err)})
			return
		}
	}

	switch scenario {
	case "onboarding":
		if err := h.seedOnboarding(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("seed onboarding failed: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"scenario": scenario, "status": "ready"})
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: fmt.Sprintf("unknown scenario: %s", scenario)})
	}
}

// startLogCapture redirects backend logs (Gin + stdlib log) to both stdout and the given file.
// Closes any previously opened log file. The file is created/truncated.
// Uses LogWriter (a SwitchableWriter) so the Gin logger picks up the change dynamically.
func (h *Dev) startLogCapture(logPath string) error {
	if h.logWriter == nil {
		return fmt.Errorf("log capture not available: no switchable writer configured")
	}

	h.logMu.Lock()
	defer h.logMu.Unlock()

	// Close previous log file if any
	if h.logFile != nil {
		_ = h.logFile.Close()
		h.logFile = nil
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o750); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	f, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	h.logFile = f

	// Tee to both stdout and the file via the switchable writer
	w := io.MultiWriter(os.Stdout, f)
	h.logWriter.Switch(w)
	log.SetOutput(w)

	log.Printf("[E2E] Log capture started: %s", logPath)
	return nil
}

// truncateUserData removes all user-generated data while preserving reference tables.
// The category is defined by COMMENT ON TABLE in migration 006_table_categories.sql:
//   - 'user_data'  → truncated
//   - 'reference'  → preserved
//   - no comment   → preserved (e.g. schema_migrations)
func (h *Dev) truncateUserData(ctx context.Context) error {
	_, err := h.pool.Exec(ctx, `
		DO $$
		DECLARE t text;
		BEGIN
			FOR t IN
				SELECT c.relname
				FROM pg_class c
				JOIN pg_namespace n ON n.oid = c.relnamespace
				LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = 0
				WHERE n.nspname = 'public'
				  AND c.relkind = 'r'
				  AND d.description = 'user_data'
			LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(t) || ' CASCADE';
			END LOOP;
		END $$
	`)
	return err
}

// ensureReferenceData re-seeds reference tables by replaying seed migrations.
// Uses the same SQL files as the initial migrations — single source of truth.
func (h *Dev) ensureReferenceData(ctx context.Context) error {
	seedFile := filepath.Join(h.migrationsDir, "003_seed_templates.sql")
	sql, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed migration: %w", err)
	}
	_, err = h.pool.Exec(ctx, string(sql))
	return err
}

// seedOnboarding resets user data to a fresh state: user exists, no chapters.
// The demo chapter will be created via the UI (dashboard-seed-demo-btn) during the flow.
func (h *Dev) seedOnboarding(ctx context.Context) error {
	if err := h.truncateUserData(ctx); err != nil {
		return fmt.Errorf("truncate user data: %w", err)
	}

	if err := h.ensureReferenceData(ctx); err != nil {
		return fmt.Errorf("ensure reference data: %w", err)
	}

	devUserID := uuid.MustParse("00000000-0000-7000-8000-000000000001")
	h.ensureUser(ctx, devUserID)

	return nil
}

func (h *Dev) ensureUser(ctx context.Context, userID uuid.UUID) {
	_, _ = h.pool.Exec(ctx,
		`INSERT INTO users (id, role, display_name, email, timezone, created_at, updated_at)
		 VALUES ($1, 'student', 'Hugo (dev)', 'dev@revisemieux.local', 'Europe/Paris', now(), now())
		 ON CONFLICT (id) DO NOTHING`,
		userID,
	)
}
