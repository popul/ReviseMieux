package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// ChapterService handles chapter-related use cases.
type ChapterService struct {
	chapterRepo chapter.Repository
}

// NewChapterService creates a new ChapterService.
func NewChapterService(chapterRepo chapter.Repository) *ChapterService {
	return &ChapterService{chapterRepo: chapterRepo}
}

// GetLessonCardItems returns the non-archived items for a chapter's current revision.
// Z5-AC04: only items from current_revision_id are shown.
func (s *ChapterService) GetLessonCardItems(ctx context.Context, chapterID uuid.UUID) ([]*chapter.Item, error) {
	ch, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("chapter_service: find chapter: %w", err)
	}
	if ch.CurrentRevisionID == nil {
		return nil, fmt.Errorf("chapter_service: %w", chapter.ErrNoRevision)
	}

	// Get non-archived items for this chapter
	items, err := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	if err != nil {
		return nil, fmt.Errorf("chapter_service: find items: %w", err)
	}

	// Filter to current revision only
	var result []*chapter.Item
	for _, item := range items {
		if item.RevisionID == *ch.CurrentRevisionID {
			result = append(result, item)
		}
	}

	return result, nil
}

// GetSessionEligibleItems returns items eligible for session composition.
// Z5-AC06: archived items are excluded.
func (s *ChapterService) GetSessionEligibleItems(ctx context.Context, chapterID uuid.UUID) ([]*chapter.Item, error) {
	// includeArchived=false ensures archived items are excluded
	return s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
}
