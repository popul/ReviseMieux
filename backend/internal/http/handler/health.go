package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/popul/revisemieux/internal/http/dto"
)

// Health handles the health check endpoint.
type Health struct {
	version string
}

// NewHealth creates a new Health handler.
func NewHealth(version string) *Health {
	return &Health{version: version}
}

// Check godoc
//
//	@Summary		Health check
//	@Description	Returns the health status of the API
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	dto.HealthResponse
//	@Router			/health [get]
func (h *Health) Check(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status:  "ok",
		Version: h.version,
	})
}
