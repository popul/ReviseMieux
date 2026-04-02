package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Onboarding handles onboarding HTTP endpoints (Z8).
type Onboarding struct {
	svc *app.OnboardingService
}

// NewOnboarding creates a new Onboarding handler.
func NewOnboarding(svc *app.OnboardingService) *Onboarding {
	return &Onboarding{svc: svc}
}

// GetStatus godoc
//
//	@Summary		Get onboarding status
//	@Description	Returns the current onboarding step (Z8-AC02)
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	dto.OnboardingStatusResponse
//	@Router			/api/v1/onboarding/status [get]
func (h *Onboarding) GetStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	status, err := h.svc.GetOnboardingStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get onboarding status"})
		return
	}

	resp := dto.OnboardingStatusResponse{
		AccountCreated:    status.AccountCreated,
		DemoSessionDone:   status.DemoSessionDone,
		FirstChapterReady: status.FirstChapterReady,
		HasDemoChapter:    status.HasDemoChapter,
		Step1Label:        status.Step1Label,
		Step2Label:        status.Step2Label,
		Step3Label:        status.Step3Label,
	}
	if status.DemoChapterID != nil {
		s := status.DemoChapterID.String()
		resp.DemoChapterID = &s
	}

	c.JSON(http.StatusOK, resp)
}

// SeedDemo godoc
//
//	@Summary		Seed demo chapter
//	@Description	Creates the demo chapter with 8 pre-generated items (Z8-AC01)
//	@Tags			onboarding
//	@Produce		json
//	@Success		201	{object}	dto.SeedDemoResponse
//	@Router			/api/v1/onboarding/seed-demo [post]
func (h *Onboarding) SeedDemo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	ch, items, err := h.svc.SeedDemoChapter(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to seed demo chapter", Details: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.SeedDemoResponse{
		ChapterID: ch.ID.String(),
		ItemCount: len(items),
		Message:   "Essaie une session de révision avec ce chapitre d'exemple — 3 min",
	})
}
