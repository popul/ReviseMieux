package chapter

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrItemArchived = errors.New("item is archived")
	ErrNoRevision   = errors.New("chapter has no current revision")
)

// --- Value Objects ---

type BlockType string

const (
	BlockText      BlockType = "TEXT"
	BlockPhoto     BlockType = "PHOTO"
	BlockSchema    BlockType = "SCHEMA"
	BlockMap       BlockType = "MAP"
	BlockGraph     BlockType = "GRAPH"
	BlockTable     BlockType = "TABLE"
	BlockCircuit   BlockType = "CIRCUIT"
	BlockDecorative BlockType = "DECORATIVE"
)

type ItemType string

const (
	ItemKnowledge ItemType = "KNOWLEDGE"
	ItemProcedure ItemType = "PROCEDURE"
	ItemDocument  ItemType = "DOCUMENT"
	ItemWriting   ItemType = "WRITING"
)

type PageStatus string

const (
	PageUploading       PageStatus = "UPLOADING"
	PageOCRPending      PageStatus = "OCR_PENDING"
	PageOCRProcessing   PageStatus = "OCR_PROCESSING"
	PageProcessed       PageStatus = "PROCESSED"
	PageItemsGenerating PageStatus = "ITEMS_GENERATING"
	PageDone            PageStatus = "DONE"
	PageNoItems         PageStatus = "NO_ITEMS"
	PageItemsFailed     PageStatus = "ITEMS_FAILED"
	PageFailed          PageStatus = "FAILED"
)

type RevisionStatus string

const (
	RevisionProcessing RevisionStatus = "PROCESSING"
	RevisionReady      RevisionStatus = "READY"
	RevisionPartial    RevisionStatus = "PARTIAL"
	RevisionFailed     RevisionStatus = "FAILED"
)

// --- Entities ---

// Chapter is the aggregate root for capture context.
type Chapter struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Subject           string
	ClassLevel        string
	Name              string
	PackID            *string
	CurrentRevisionID *uuid.UUID
	Archived          bool
	IsDemo            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewChapter creates a new chapter.
func NewChapter(userID uuid.UUID, subject, classLevel, name string, now time.Time) *Chapter {
	return &Chapter{
		ID:         uuid.Must(uuid.NewV7()),
		UserID:     userID,
		Subject:    subject,
		ClassLevel: classLevel,
		Name:       name,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Revision represents a version of notes uploaded for a chapter.
type Revision struct {
	ID             uuid.UUID
	ChapterID      uuid.UUID
	RevisionNumber int
	Status         RevisionStatus
	CreatedAt      time.Time
}

// Page represents a single photo page within a revision.
type Page struct {
	ID         uuid.UUID
	RevisionID uuid.UUID
	PhotoURL   string
	PageOrder  int
	OCRStatus  PageStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Block represents an OCR-detected zone within a page.
type Block struct {
	ID         uuid.UUID
	PageID     uuid.UUID
	BlockType  BlockType
	CropURL    *string
	Confidence float32
	OCRText    *string
	CreatedAt  time.Time
}

// Notion represents a concept/topic extracted from a chapter.
type Notion struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	Name      string
	SortOrder int
	CreatedAt time.Time
}

// Item represents a reviewable knowledge unit.
type Item struct {
	ID                  uuid.UUID
	ChapterID           uuid.UUID
	NotionID            *uuid.UUID
	RevisionID          uuid.UUID
	ItemType            ItemType
	Term                *string
	Confidence          float32
	ValidationRequired  bool
	Archived            bool
	Keywords            []string
	Steps               []ItemStep
	LLMModelVersion     *string
	PromptTemplateVersion *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ItemStep represents an ordered step in a PROCEDURE item.
type ItemStep struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	StepOrder int
	Content   string
}

// Archive marks the item as archived so it won't be proposed in sessions.
func (i *Item) Archive(now time.Time) {
	i.Archived = true
	i.UpdatedAt = now
}

// Unarchive restores an archived item.
func (i *Item) Unarchive(now time.Time) {
	i.Archived = false
	i.UpdatedAt = now
}
