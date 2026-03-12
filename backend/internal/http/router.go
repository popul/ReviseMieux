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
	JWTSecret         string
	Version           string
	PipelineHandler   *handler.Pipeline
	MasteryHandler    *handler.Mastery
	ChapterHandler    *handler.Chapter
	SessionHandler    *handler.Session
	ValidationHandler *handler.Validation
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
		// Chapters
		if cfg.ChapterHandler != nil {
			api.GET("/chapters", cfg.ChapterHandler.List)
			api.POST("/chapters", cfg.ChapterHandler.Create)
			api.GET("/chapters/:chapter_id/lesson-card", cfg.ChapterHandler.GetLessonCard)
		}

		// Pipeline (upload is under chapters)
		if cfg.PipelineHandler != nil {
			api.POST("/chapters/:chapter_id/upload", cfg.PipelineHandler.Upload)
		}

		// Masteries
		if cfg.MasteryHandler != nil {
			api.GET("/masteries", cfg.MasteryHandler.GetByUser)
			api.GET("/masteries/:item_id", cfg.MasteryHandler.GetByItem)
			api.POST("/masteries/attempt", cfg.MasteryHandler.RecordAttempt)
		}

		// Sessions
		if cfg.SessionHandler != nil {
			api.POST("/sessions/daily", cfg.SessionHandler.ComposeDaily)
			api.GET("/sessions/:id", cfg.SessionHandler.GetByID)
			api.POST("/sessions/:id/resume", cfg.SessionHandler.Resume)
			api.GET("/sessions/:id/questions", cfg.SessionHandler.GetQuestions)
			api.POST("/sessions/:id/answer", cfg.SessionHandler.SubmitAnswer)
		}

		// Validation
		if cfg.ValidationHandler != nil {
			api.GET("/validations", cfg.ValidationHandler.ListPending)
			api.POST("/validations/:id/resolve", cfg.ValidationHandler.Resolve)
		}
	}

	return r
}
