package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/http/dto"
)

// Pipeline handles photo upload and pipeline processing.
type Pipeline struct {
	svc *app.PipelineService
}

// NewPipeline creates a new Pipeline handler.
func NewPipeline(svc *app.PipelineService) *Pipeline {
	return &Pipeline{svc: svc}
}

// Upload godoc
//
//	@Summary		Upload photos to a chapter
//	@Description	Uploads one or more photos and triggers the J0 pipeline
//	@Tags			pipeline
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			chapter_id	path		string					true	"Chapter ID"
//	@Param			photos		formData	file					true	"Photo files"
//	@Success		200			{object}	dto.PipelineUploadResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Router			/api/v1/chapters/{chapter_id}/upload [post]
func (h *Pipeline) Upload(c *gin.Context) {
	chapterIDStr := c.Param("chapter_id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid chapter_id"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "expected multipart form"})
		return
	}

	files := form.File["photos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "no photos provided"})
		return
	}

	var uploads []app.PageUpload
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "cannot read file",
				Details: fh.Filename,
			})
			return
		}
		defer f.Close()

		uploads = append(uploads, app.PageUpload{
			FileName:    fh.Filename,
			ContentType: fh.Header.Get("Content-Type"),
			Body:        f,
		})
	}

	result, err := h.svc.UploadAndProcess(c.Request.Context(), chapterID, uploads)
	if err != nil {
		if errors.Is(err, chapter.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "pipeline error", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PipelineUploadResponse{
		RevisionID:     result.RevisionID.String(),
		TotalPages:     result.TotalPages,
		ProcessedPages: result.ProcessedPages,
		FailedPages:    result.FailedPages,
		TotalItems:     result.TotalItems,
	})
}

// GetProgress godoc
//
//	@Summary		Get pipeline progress for a revision
//	@Description	Returns the current processing status (Z8-AC04)
//	@Tags			pipeline
//	@Produce		json
//	@Param			revision_id	path		string	true	"Revision ID"
//	@Success		200			{object}	dto.PipelineProgressResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/api/v1/revisions/{revision_id}/progress [get]
func (h *Pipeline) GetProgress(c *gin.Context) {
	revisionID, err := uuid.Parse(c.Param("revision_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid revision_id"})
		return
	}

	progress, err := h.svc.GetRevisionProgress(c.Request.Context(), revisionID)
	if err != nil {
		if errors.Is(err, chapter.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "revision not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get progress"})
		return
	}

	c.JSON(http.StatusOK, dto.PipelineProgressResponse{
		RevisionID:     progress.RevisionID.String(),
		Status:         progress.Status,
		TotalPages:     progress.TotalPages,
		ProcessedPages: progress.ProcessedPages,
		FailedPages:    progress.FailedPages,
		TotalItems:     progress.TotalItems,
		Phase:          progress.Phase,
		PhaseMessage:   progress.PhaseMessage,
	})
}
