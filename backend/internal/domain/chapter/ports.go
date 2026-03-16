package chapter

import (
	"context"
	"io"
)

// Storage is the port for object storage (S3, etc.).
type Storage interface {
	// Upload stores a file and returns its URL.
	Upload(ctx context.Context, key string, contentType string, body io.Reader) (url string, err error)
}

// OCRBlock represents a single OCR-detected zone in an image.
type OCRBlock struct {
	Text       string
	BlockType  BlockType
	Confidence float32
	BBox       BBox
}

// BBox represents a bounding box.
type BBox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// OCRResult holds the OCR output for a single page.
type OCRResult struct {
	Blocks []OCRBlock
}

// OCRService is the port for OCR processing.
type OCRService interface {
	// ProcessPage runs OCR on a page image and returns detected blocks.
	ProcessPage(ctx context.Context, imageURL string) (*OCRResult, error)
}

// StructuredItem represents an item extracted by the LLM from OCR text.
type StructuredItem struct {
	Type       ItemType
	Term       string
	Keywords   []string
	Steps      []string // for PROCEDURE items
	NotionName string
	Confidence float32
}

// StructurationResult holds the LLM output for a set of OCR blocks.
type StructurationResult struct {
	Items   []StructuredItem
	Notions []string
}

// LLMService is the port for LLM-based structuration.
type LLMService interface {
	// StructureBlocks takes OCR text blocks and produces structured items and notions.
	StructureBlocks(ctx context.Context, subject string, blocks []OCRBlock) (*StructurationResult, error)
}
