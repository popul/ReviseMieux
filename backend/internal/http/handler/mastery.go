package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Mastery handles mastery-related HTTP endpoints.
type Mastery struct {
	svc *app.MasteryService
}

// NewMastery creates a new Mastery handler.
func NewMastery(svc *app.MasteryService) *Mastery {
	return &Mastery{svc: svc}
}

// GetByUser godoc
//
//	@Summary		Get masteries for authenticated user by state
//	@Description	Returns all masteries in the given state for the current user
//	@Tags			mastery
//	@Produce		json
//	@Param			state	query		string	true	"Mastery state (UNKNOWN, FRAGILE, OK, SOLID)"
//	@Success		200		{array}		dto.MasteryResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/api/v1/masteries [get]
func (h *Mastery) GetByUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	stateStr := c.Query("state")
	state, err := mastery.ParseState(stateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid state", Details: err.Error()})
		return
	}

	masteries, err := h.svc.GetByUser(c.Request.Context(), userID, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch masteries"})
		return
	}

	result := make([]dto.MasteryResponse, len(masteries))
	for i, m := range masteries {
		result[i] = toMasteryDTO(m)
	}
	c.JSON(http.StatusOK, result)
}

// GetByItem godoc
//
//	@Summary		Get mastery for a specific item
//	@Description	Returns the mastery for the authenticated user and given item
//	@Tags			mastery
//	@Produce		json
//	@Param			item_id	path		string	true	"Item ID"
//	@Success		200		{object}	dto.MasteryResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/api/v1/masteries/{item_id} [get]
func (h *Mastery) GetByItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid item_id"})
		return
	}

	m, err := h.svc.GetByUserAndItem(c.Request.Context(), userID, itemID)
	if err != nil {
		if errors.Is(err, mastery.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "mastery not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch mastery"})
		return
	}

	c.JSON(http.StatusOK, toMasteryDTO(m))
}

// RecordAttempt godoc
//
//	@Summary		Record an attempt on an item
//	@Description	Records a score and transitions mastery state
//	@Tags			mastery
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RecordAttemptRequest	true	"Attempt data"
//	@Success		200		{object}	dto.RecordAttemptResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/api/v1/masteries/attempt [post]
func (h *Mastery) RecordAttempt(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.RecordAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid item_id"})
		return
	}

	m, err := h.svc.RecordAttempt(c.Request.Context(), userID, itemID, *req.Score)
	if err != nil {
		if errors.Is(err, mastery.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "mastery not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to record attempt", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RecordAttemptResponse{
		Mastery:  toMasteryDTO(m),
		NewState: string(m.State),
	})
}

func toMasteryDTO(m *mastery.Mastery) dto.MasteryResponse {
	return dto.MasteryResponse{
		ID:                   m.ID.String(),
		UserID:               m.UserID.String(),
		ItemID:               m.ItemID.String(),
		State:                string(m.State),
		NextDueAt:            m.NextDueAt,
		LastReviewAt:         m.LastReviewAt,
		ConsecutiveSuccesses: m.ConsecutiveSuccesses,
		ConsecutiveFailures:  m.ConsecutiveFailures,
		LastSuccessAt:        m.LastSuccessAt,
		CappedAtOK:           m.CappedAtOK,
	}
}
