package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// Chapter handles chapter-related HTTP endpoints.
type Chapter struct {
	svc         *app.ChapterService
	chapterRepo chapter.Repository
	idGen       event.IDGenerator
	clock       event.Clock
}

// NewChapter creates a new Chapter handler.
func NewChapter(svc *app.ChapterService, chapterRepo chapter.Repository, idGen event.IDGenerator, clock event.Clock) *Chapter {
	return &Chapter{svc: svc, chapterRepo: chapterRepo, idGen: idGen, clock: clock}
}

// List godoc
//
//	@Summary		List chapters for authenticated user
//	@Description	Returns all non-archived chapters
//	@Tags			chapters
//	@Produce		json
//	@Success		200	{array}		dto.ChapterResponse
//	@Router			/api/v1/chapters [get]
func (h *Chapter) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	chapters, err := h.chapterRepo.FindByUser(c.Request.Context(), userID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list chapters"})
		return
	}

	result := make([]dto.ChapterResponse, len(chapters))
	for i, ch := range chapters {
		result[i] = toChapterDTO(ch)
	}
	c.JSON(http.StatusOK, result)
}

// Create godoc
//
//	@Summary		Create a new chapter
//	@Tags			chapters
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateChapterRequest	true	"Chapter data"
//	@Success		201		{object}	dto.ChapterResponse
//	@Router			/api/v1/chapters [post]
func (h *Chapter) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}

	ch := chapter.NewChapter(h.idGen, userID, req.Subject, req.ClassLevel, req.Name, h.clock.Now())
	if err := h.chapterRepo.Save(c.Request.Context(), ch); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create chapter"})
		return
	}

	c.JSON(http.StatusCreated, toChapterDTO(ch))
}

// GetLessonCard godoc
//
//	@Summary		Get lesson card for a chapter
//	@Description	Returns items and notions for the current revision (Z5-AC04)
//	@Tags			chapters
//	@Produce		json
//	@Param			chapter_id	path		string	true	"Chapter ID"
//	@Success		200			{object}	dto.LessonCardResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/api/v1/chapters/{chapter_id}/lesson-card [get]
func (h *Chapter) GetLessonCard(c *gin.Context) {
	chapterID, err := uuid.Parse(c.Param("chapter_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid chapter id"})
		return
	}

	ctx := c.Request.Context()

	ch, err := h.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		if errors.Is(err, chapter.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch chapter"})
		return
	}

	items, err := h.svc.GetLessonCardItems(ctx, chapterID)
	if err != nil {
		if errors.Is(err, chapter.ErrNoRevision) {
			c.JSON(http.StatusOK, dto.LessonCardResponse{
				Chapter: toChapterDTO(ch),
				Items:   []dto.ItemResponse{},
				Notions: []dto.NotionResponse{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch items"})
		return
	}

	notions, err := h.chapterRepo.FindNotionsByChapter(ctx, chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch notions"})
		return
	}

	itemDTOs := make([]dto.ItemResponse, len(items))
	for i, item := range items {
		itemDTOs[i] = toItemDTO(item)
	}

	notionDTOs := make([]dto.NotionResponse, len(notions))
	for i, n := range notions {
		notionDTOs[i] = toNotionDTO(n)
	}

	c.JSON(http.StatusOK, dto.LessonCardResponse{
		Chapter: toChapterDTO(ch),
		Items:   itemDTOs,
		Notions: notionDTOs,
	})
}

func toChapterDTO(ch *chapter.Chapter) dto.ChapterResponse {
	var revID *string
	if ch.CurrentRevisionID != nil {
		s := ch.CurrentRevisionID.String()
		revID = &s
	}
	return dto.ChapterResponse{
		ID:                ch.ID.String(),
		UserID:            ch.UserID.String(),
		Subject:           ch.Subject,
		ClassLevel:        ch.ClassLevel,
		Name:              ch.Name,
		CurrentRevisionID: revID,
		Archived:          ch.Archived,
		IsDemo:            ch.IsDemo,
	}
}

func toItemDTO(item *chapter.Item) dto.ItemResponse {
	var notionID *string
	if item.NotionID != nil {
		s := item.NotionID.String()
		notionID = &s
	}
	return dto.ItemResponse{
		ID:                 item.ID.String(),
		ChapterID:          item.ChapterID.String(),
		NotionID:           notionID,
		ItemType:           string(item.ItemType),
		Term:               item.Term,
		Confidence:         item.Confidence,
		ValidationRequired: item.ValidationRequired,
		Archived:           item.Archived,
		Keywords:           item.Keywords,
	}
}

func toNotionDTO(n *chapter.Notion) dto.NotionResponse {
	return dto.NotionResponse{
		ID:        n.ID.String(),
		ChapterID: n.ChapterID.String(),
		Name:      n.Name,
		SortOrder: n.SortOrder,
	}
}
