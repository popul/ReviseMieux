package dto

// ChapterResponse represents a chapter in API responses.
type ChapterResponse struct {
	ID                string  `json:"id"`
	UserID            string  `json:"user_id"`
	Subject           string  `json:"subject"`
	ClassLevel        string  `json:"class_level"`
	Name              string  `json:"name"`
	CurrentRevisionID *string `json:"current_revision_id,omitempty"`
	Archived          bool    `json:"archived"`
	IsDemo            bool    `json:"is_demo"`
}

// CreateChapterRequest is the request body for creating a chapter.
type CreateChapterRequest struct {
	Subject    string `json:"subject" binding:"required"`
	ClassLevel string `json:"class_level" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

// ItemResponse represents an item in API responses.
type ItemResponse struct {
	ID                 string   `json:"id"`
	ChapterID          string   `json:"chapter_id"`
	NotionID           *string  `json:"notion_id,omitempty"`
	ItemType           string   `json:"item_type"`
	Term               *string  `json:"term,omitempty"`
	Confidence         float32  `json:"confidence"`
	ValidationRequired bool     `json:"validation_required"`
	Archived           bool     `json:"archived"`
	Keywords           []string `json:"keywords,omitempty"`
}

// NotionResponse represents a notion in API responses.
type NotionResponse struct {
	ID        string `json:"id"`
	ChapterID string `json:"chapter_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// LessonCardResponse is the full lesson card for a chapter.
type LessonCardResponse struct {
	Chapter ChapterResponse  `json:"chapter"`
	Items   []ItemResponse   `json:"items"`
	Notions []NotionResponse `json:"notions"`
}
