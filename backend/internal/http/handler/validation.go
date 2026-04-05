package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/validation"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Validation handles validation task HTTP endpoints.
type Validation struct {
	svc *app.ValidationService
}

// NewValidation creates a new Validation handler.
func NewValidation(svc *app.ValidationService) *Validation {
	return &Validation{svc: svc}
}

// ListPending godoc
//
//	@Summary		List pending validation tasks
//	@Description	Returns pending validation tasks for HITL review
//	@Tags			validation
//	@Produce		json
//	@Param			limit	query		int	false	"Max results (default 20)"
//	@Success		200		{array}		dto.ValidationTaskResponse
//	@Router			/api/v1/validations [get]
func (h *Validation) ListPending(c *gin.Context) {
	limit := 20

	tasks, err := h.svc.ListPending(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list tasks"})
		return
	}

	result := make([]dto.ValidationTaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = h.toValidationTaskDTO(c.Request.Context(), t)
	}
	c.JSON(http.StatusOK, result)
}

// Resolve godoc
//
//	@Summary		Resolve a validation task
//	@Description	Performs confirm, correct, unknown, or ignore action
//	@Tags			validation
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Task ID"
//	@Param			body	body		dto.ResolveValidationRequest	true	"Resolution action"
//	@Success		200		{object}	dto.MessageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/api/v1/validations/{id}/resolve [post]
func (h *Validation) Resolve(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(c.Param("validation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task id"})
		return
	}

	var req dto.ResolveValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}

	switch req.Action {
	case "confirm":
		err = h.svc.Confirm(c.Request.Context(), taskID, userID)
	case "correct":
		err = h.svc.Correct(c.Request.Context(), taskID, userID, req.Correction)
	case "unknown":
		err = h.svc.MarkUnknown(c.Request.Context(), taskID, userID)
	case "ignore":
		err = h.svc.Ignore(c.Request.Context(), taskID, userID)
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid action"})
		return
	}

	if err != nil {
		if errors.Is(err, validation.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "task not found"})
			return
		}
		if errors.Is(err, validation.ErrAlreadyResolved) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "task already resolved"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to resolve task", Details: err.Error()})
		return
	}

	// Fetch updated task to return full representation
	updatedTask, fetchErr := h.svc.GetByID(c.Request.Context(), taskID)
	if fetchErr != nil {
		c.JSON(http.StatusOK, dto.MessageResponse{Message: "task resolved"})
		return
	}
	c.JSON(http.StatusOK, h.toValidationTaskDTO(c.Request.Context(), updatedTask))
}

func (h *Validation) toValidationTaskDTO(ctx context.Context, t *validation.ValidationTask) dto.ValidationTaskResponse {
	resp := dto.ValidationTaskResponse{
		ID:         t.ID.String(),
		ItemID:     t.ItemID.String(),
		CropURL:    t.CropURL,
		Suggestion: t.Suggestion,
		Priority:   t.Priority,
		Status:     string(t.Status),
		Source:     string(t.Source),
		CreatedAt:  t.CreatedAt,
	}
	// Enrich with item data (ChapterID + Term)
	if item, err := h.svc.GetItemByID(ctx, t.ItemID); err == nil {
		chapterID := item.ChapterID.String()
		resp.ChapterID = chapterID
		resp.ItemTerm = item.Term
	}
	// Use UpdatedAt as ResolvedAt when task is resolved
	if t.IsResolved() {
		resp.ResolvedAt = &t.UpdatedAt
	}
	return resp
}
