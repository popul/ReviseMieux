package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// RouterConfig holds dependencies needed to build the router.
type RouterConfig struct {
	JWTSecret       string
	Version         string
	PipelineHandler *handler.Pipeline
}

// NewRouter creates and configures the Gin router with all routes.
func NewRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// --- Swagger ---
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Public routes ---
	health := handler.NewHealth(cfg.Version)
	r.GET("/health", health.Check)

	// --- Authenticated routes ---
	api := r.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWTSecret))
	{
		// Pipeline
		if cfg.PipelineHandler != nil {
			api.POST("/chapters/:chapter_id/upload", cfg.PipelineHandler.Upload)
		}
	}

	return r
}
