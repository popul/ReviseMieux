package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/session"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Session handles session-related HTTP endpoints.
type Session struct {
	svc *app.SessionService
}

// NewSession creates a new Session handler.
func NewSession(svc *app.SessionService) *Session {
	return &Session{svc: svc}
}

// ComposeDaily godoc
//
//	@Summary		Compose a daily review session
//	@Description	Creates a daily session from due items (Z4-AC06: empty pool returns error)
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.ComposeDailyRequest	true	"Session composition data"
//	@Success		201		{object}	dto.SessionResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/sessions/daily [post]
func (h *Session) ComposeDaily(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.ComposeDailyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}

	chapterID, err := uuid.Parse(req.ChapterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid chapter_id"})
		return
	}

	sess, err := h.svc.ComposeDaily(c.Request.Context(), userID, chapterID)
	if err != nil {
		if errors.Is(err, session.ErrEmptyPool) {
			c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "no items available for review"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to compose session", Details: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toSessionDTO(sess))
}

// Resume godoc
//
//	@Summary		Resume an interrupted session
//	@Description	Resumes from last question index (Z4-AC04)
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	dto.SessionResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/api/v1/sessions/{id}/resume [post]
func (h *Session) Resume(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session id"})
		return
	}

	sess, err := h.svc.ResumeSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "session not found"})
			return
		}
		if errors.Is(err, session.ErrSessionNotResumable) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "session cannot be resumed"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to resume session"})
		return
	}

	c.JSON(http.StatusOK, toSessionDTO(sess))
}

// GetByID godoc
//
//	@Summary		Get session by ID
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	dto.SessionResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/sessions/{id} [get]
func (h *Session) GetByID(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session id"})
		return
	}

	sess, err := h.svc.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch session"})
		return
	}

	c.JSON(http.StatusOK, toSessionDTO(sess))
}

// GetQuestions godoc
//
//	@Summary		Get questions for a session
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{array}		dto.QuestionResponse
//	@Router			/api/v1/sessions/{id}/questions [get]
func (h *Session) GetQuestions(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session id"})
		return
	}

	questions, err := h.svc.GetQuestions(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch questions"})
		return
	}

	result := make([]dto.QuestionResponse, len(questions))
	for i, q := range questions {
		result[i] = toQuestionDTO(q)
	}
	c.JSON(http.StatusOK, result)
}

// SubmitAnswer godoc
//
//	@Summary		Submit an answer to a question
//	@Description	Records attempt and returns enriched feedback (Z4-AC09)
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Session ID"
//	@Param			body	body		dto.SubmitAnswerRequest		true	"Answer data"
//	@Success		200		{object}	dto.SubmitAnswerResponse
//	@Router			/api/v1/sessions/{id}/answer [post]
func (h *Session) SubmitAnswer(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session id"})
		return
	}

	var req dto.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid question_id"})
		return
	}

	result, err := h.svc.SubmitAnswer(c.Request.Context(), sessionID, questionID, userID, req.Answer, req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to submit answer", Details: err.Error()})
		return
	}

	resp := dto.SubmitAnswerResponse{
		AttemptID: result.AttemptID.String(),
		Score:     result.Score,
	}
	if result.Feedback != nil {
		resp.Feedback = &dto.Feedback{
			CorrectAnswer:  result.Feedback.CorrectAnswer,
			WhatWasMissing: result.Feedback.WhatWasMissing,
			Hint:           result.Feedback.Hint,
		}
	}

	c.JSON(http.StatusOK, resp)
}

// Debrief godoc
//
//	@Summary		Get session debrief
//	@Description	Returns score, total, percentage after session completion
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	dto.DebriefResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/sessions/{id}/debrief [get]
func (h *Session) Debrief(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session id"})
		return
	}

	result, err := h.svc.GetDebrief(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get debrief"})
		return
	}

	transitions := make([]dto.MasteryTransition, len(result.Transitions))
	for i, t := range result.Transitions {
		transitions[i] = dto.MasteryTransition{
			ItemID:   t.ItemID.String(),
			ItemTerm: t.ItemTerm,
			From:     t.From,
			To:       t.To,
		}
	}

	c.JSON(http.StatusOK, dto.DebriefResponse{
		Score:       result.Score,
		Total:       result.Total,
		Percentage:  result.Percentage,
		Transitions: transitions,
	})
}

func toSessionDTO(s *session.Session) dto.SessionResponse {
	chapterIDs := make([]string, len(s.ChapterIDs))
	for i, id := range s.ChapterIDs {
		chapterIDs[i] = id.String()
	}
	return dto.SessionResponse{
		ID:                   s.ID.String(),
		UserID:               s.UserID.String(),
		SessionType:          string(s.SessionType),
		Status:               string(s.Status),
		TriggerType:          string(s.TriggerType),
		StartedAt:            s.StartedAt,
		CompletedAt:          s.CompletedAt,
		CurrentQuestionIndex: s.CurrentQuestionIndex,
		ChapterIDs:           chapterIDs,
	}
}

// templateQuestionType maps template IDs to their question type.
var templateQuestionType = map[string]string{
	"GEN.KNOW.FLASH_MCQ":      "MCQ",
	"GEN.KNOW.DEF_SHORT":      "SHORT_ANSWER",
	"GEN.KNOW.CLOZE_KEYWORDS": "CLOZE",
}

// templateChoices provides choices for MCQ templates.
var templateChoices = map[string][]string{
	"GEN.KNOW.FLASH_MCQ": {"Vrai", "Faux"},
}

func toQuestionDTO(q *session.Question) dto.QuestionResponse {
	qType := templateQuestionType[q.TemplateID]
	if qType == "" {
		qType = "SHORT_ANSWER"
	}
	return dto.QuestionResponse{
		ID:             q.ID.String(),
		TemplateID:     q.TemplateID,
		ItemID:         q.ItemID.String(),
		QuestionType:   qType,
		RenderedPrompt: q.RenderedPrompt,
		Choices:        templateChoices[q.TemplateID],
		VisualURL:      q.RenderedVisualURL,
	}
}
