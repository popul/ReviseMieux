package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Dev handles development-only endpoints.
type Dev struct {
	jwtSecret string
	pool      *pgxpool.Pool
}

// NewDev creates a new Dev handler with the given JWT secret and DB pool.
func NewDev(jwtSecret string, pool *pgxpool.Pool) *Dev {
	return &Dev{jwtSecret: jwtSecret, pool: pool}
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

func (h *Dev) ensureUser(ctx context.Context, userID uuid.UUID) {
	_, _ = h.pool.Exec(ctx,
		`INSERT INTO users (id, role, display_name, email, timezone, created_at, updated_at)
		 VALUES ($1, 'student', 'Hugo (dev)', 'dev@revisemieux.local', 'Europe/Paris', now(), now())
		 ON CONFLICT (id) DO NOTHING`,
		userID,
	)
}
