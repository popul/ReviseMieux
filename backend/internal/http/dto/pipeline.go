package dto

// PipelineUploadResponse is returned after a photo upload + pipeline run.
type PipelineUploadResponse struct {
	RevisionID     string `json:"revision_id" example:"01945b2c-1234-7000-8000-000000000001"`
	TotalPages     int    `json:"total_pages" example:"3"`
	ProcessedPages int    `json:"processed_pages" example:"2"`
	FailedPages    int    `json:"failed_pages" example:"1"`
	TotalItems     int    `json:"total_items" example:"8"`
}
