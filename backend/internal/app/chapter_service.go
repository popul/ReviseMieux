package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// NotionMastery represents aggregated mastery for a single notion (Z7-AC16).
type NotionMastery struct {
	Notion        *chapter.Notion
	TotalItems    int
	MasteryStates map[mastery.State]int // count per state
	DominantState mastery.State
}

// ChapterService handles chapter-related use cases.
type ChapterService struct {
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
}

// NewChapterService creates a new ChapterService.
func NewChapterService(chapterRepo chapter.Repository, masteryRepo ...mastery.Repository) *ChapterService {
	svc := &ChapterService{chapterRepo: chapterRepo}
	if len(masteryRepo) > 0 {
		svc.masteryRepo = masteryRepo[0]
	}
	return svc
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

// GetNotionMastery returns mastery aggregated by notion for a chapter (Z7-AC16).
// Sorted by revision urgency: FRAGILE first, then UNKNOWN, then OK, then SOLID.
func (s *ChapterService) GetNotionMastery(ctx context.Context, userID, chapterID uuid.UUID) ([]NotionMastery, error) {
	notions, err := s.chapterRepo.FindNotionsByChapter(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("chapter_service: find notions: %w", err)
	}

	items, err := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	if err != nil {
		return nil, fmt.Errorf("chapter_service: find items: %w", err)
	}

	// Build item→mastery map
	masteryMap := make(map[uuid.UUID]mastery.State)
	if s.masteryRepo != nil {
		for _, state := range []mastery.State{mastery.Unknown, mastery.Fragile, mastery.OK, mastery.Solid} {
			masteries, err := s.masteryRepo.FindByUserAndState(ctx, userID, state)
			if err != nil {
				continue
			}
			for _, m := range masteries {
				masteryMap[m.ItemID] = m.State
			}
		}
	}

	// Group items by notion
	notionItems := make(map[uuid.UUID][]*chapter.Item) // notionID → items
	for _, item := range items {
		if item.NotionID != nil {
			notionItems[*item.NotionID] = append(notionItems[*item.NotionID], item)
		}
	}

	var result []NotionMastery
	for _, notion := range notions {
		nm := NotionMastery{
			Notion:        notion,
			MasteryStates: make(map[mastery.State]int),
		}

		notItems := notionItems[notion.ID]
		nm.TotalItems = len(notItems)

		for _, item := range notItems {
			state, ok := masteryMap[item.ID]
			if !ok {
				state = mastery.Unknown
			}
			nm.MasteryStates[state]++
		}

		// Determine dominant state (urgency order: FRAGILE > UNKNOWN > OK > SOLID)
		if nm.MasteryStates[mastery.Fragile] > 0 {
			nm.DominantState = mastery.Fragile
		} else if nm.MasteryStates[mastery.Unknown] > 0 {
			nm.DominantState = mastery.Unknown
		} else if nm.MasteryStates[mastery.OK] > 0 {
			nm.DominantState = mastery.OK
		} else {
			nm.DominantState = mastery.Solid
		}

		result = append(result, nm)
	}

	// Sort by urgency: FRAGILE first, UNKNOWN, OK, SOLID
	stateOrder := map[mastery.State]int{
		mastery.Fragile: 0,
		mastery.Unknown: 1,
		mastery.OK:      2,
		mastery.Solid:   3,
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if stateOrder[result[j].DominantState] < stateOrder[result[i].DominantState] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}
