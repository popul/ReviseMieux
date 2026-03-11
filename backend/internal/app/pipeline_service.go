package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// PageUpload represents a single photo to upload.
type PageUpload struct {
	FileName    string
	ContentType string
	Body        io.Reader
}

// PipelineResult holds the outcome of a pipeline run.
type PipelineResult struct {
	RevisionID     uuid.UUID
	TotalPages     int
	ProcessedPages int
	FailedPages    int
	TotalItems     int
}

// PipelineService orchestrates the J0 pipeline: upload → OCR → structuration → items.
type PipelineService struct {
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
	storage     chapter.Storage
	ocr         chapter.OCRService
	llm         chapter.LLMService
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
}

// NewPipelineService creates a new PipelineService.
func NewPipelineService(
	chapterRepo chapter.Repository,
	masteryRepo mastery.Repository,
	storage chapter.Storage,
	ocr chapter.OCRService,
	llm chapter.LLMService,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
) *PipelineService {
	return &PipelineService{
		chapterRepo: chapterRepo,
		masteryRepo: masteryRepo,
		storage:     storage,
		ocr:         ocr,
		llm:         llm,
		publisher:   publisher,
		clock:       clock,
		idGen:       idGen,
	}
}

// UploadAndProcess runs the full J0 pipeline for a set of photos on a chapter.
func (s *PipelineService) UploadAndProcess(ctx context.Context, chapterID uuid.UUID, photos []PageUpload) (*PipelineResult, error) {
	now := s.clock.Now()

	// 1. Fetch chapter
	ch, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: find chapter: %w", err)
	}

	// 2. Create revision
	rev := &chapter.Revision{
		ID:             s.idGen.New(),
		ChapterID:      ch.ID,
		RevisionNumber: 1, // TODO: increment based on existing revisions
		Status:         chapter.RevisionProcessing,
		CreatedAt:      now,
	}
	if err := s.chapterRepo.SaveRevision(ctx, rev); err != nil {
		return nil, fmt.Errorf("pipeline: save revision: %w", err)
	}

	// 3. Upload photos and create pages
	var pages []*chapter.Page
	for i, photo := range photos {
		key := fmt.Sprintf("chapters/%s/revisions/%s/pages/%d/%s",
			ch.ID, rev.ID, i+1, photo.FileName)

		url, err := s.storage.Upload(ctx, key, photo.ContentType, photo.Body)
		if err != nil {
			return nil, fmt.Errorf("pipeline: upload page %d: %w", i+1, err)
		}

		page := &chapter.Page{
			ID:         s.idGen.New(),
			RevisionID: rev.ID,
			PhotoURL:   url,
			PageOrder:  i + 1,
			OCRStatus:  chapter.PageOCRPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.chapterRepo.SavePage(ctx, page); err != nil {
			return nil, fmt.Errorf("pipeline: save page %d: %w", i+1, err)
		}
		pages = append(pages, page)
	}

	// 4. Process each page: OCR → structuration → items
	result := &PipelineResult{
		RevisionID: rev.ID,
		TotalPages: len(pages),
	}

	var allItems []*chapter.Item
	notionCache := make(map[string]*chapter.Notion) // name -> notion

	for _, page := range pages {
		items, err := s.processPage(ctx, ch, rev, page, notionCache, now)
		if err != nil {
			// Z2-AC05: page failure does not block the pipeline
			page.OCRStatus = chapter.PageItemsFailed
			page.UpdatedAt = now
			s.chapterRepo.SavePage(ctx, page)
			result.FailedPages++
			continue
		}

		if len(items) == 0 {
			// Z2-AC04: no items extracted
			page.OCRStatus = chapter.PageNoItems
			page.UpdatedAt = now
			s.chapterRepo.SavePage(ctx, page)
			result.ProcessedPages++
			continue
		}

		page.OCRStatus = chapter.PageDone
		page.UpdatedAt = now
		s.chapterRepo.SavePage(ctx, page)
		result.ProcessedPages++
		allItems = append(allItems, items...)
	}

	result.TotalItems = len(allItems)

	// 5. Determine revision final status
	if result.TotalItems == 0 {
		// Z2-AC07: 0 items total → FAILED
		rev.Status = chapter.RevisionFailed
	} else if result.FailedPages > 0 {
		// Z2-AC01: some pages failed → PARTIAL
		rev.Status = chapter.RevisionPartial
	} else {
		rev.Status = chapter.RevisionReady
	}
	s.chapterRepo.SaveRevision(ctx, rev)

	// 6. Set chapter's current revision
	ch.CurrentRevisionID = &rev.ID
	ch.UpdatedAt = now
	s.chapterRepo.Save(ctx, ch)

	// 7. Create UNKNOWN masteries for all new items and publish event
	if len(allItems) > 0 {
		var masteries []*mastery.Mastery
		var itemIDs []uuid.UUID
		for _, item := range allItems {
			m := mastery.NewMastery(s.idGen, ch.UserID, item.ID, now)
			masteries = append(masteries, m)
			itemIDs = append(itemIDs, item.ID)
		}
		if err := s.masteryRepo.SaveAll(ctx, masteries); err != nil {
			return nil, fmt.Errorf("pipeline: save masteries: %w", err)
		}

		s.publisher.Publish(ctx, event.ItemsGenerated{
			BaseEvent: event.BaseEvent{OccurredOn: now},
			ChapterID: ch.ID,
			ItemIDs:   itemIDs,
		})
	}

	return result, nil
}

func (s *PipelineService) processPage(
	ctx context.Context,
	ch *chapter.Chapter,
	rev *chapter.Revision,
	page *chapter.Page,
	notionCache map[string]*chapter.Notion,
	now time.Time,
) ([]*chapter.Item, error) {
	// OCR
	page.OCRStatus = chapter.PageOCRProcessing
	page.UpdatedAt = now
	s.chapterRepo.SavePage(ctx, page)

	ocrResult, err := s.ocr.ProcessPage(ctx, page.PhotoURL)
	if err != nil {
		return nil, fmt.Errorf("ocr: %w", err)
	}

	if len(ocrResult.Blocks) == 0 {
		return nil, nil
	}

	// Save blocks
	for _, ob := range ocrResult.Blocks {
		block := &chapter.Block{
			ID:         s.idGen.New(),
			PageID:     page.ID,
			BlockType:  ob.BlockType,
			Confidence: ob.Confidence,
			OCRText:    &ob.Text,
			CreatedAt:  now,
		}
		// We don't persist blocks via the chapter repo interface yet.
		// For now, blocks are transient — they feed the LLM step.
		_ = block
	}

	// Structuration via LLM
	page.OCRStatus = chapter.PageItemsGenerating
	page.UpdatedAt = now
	s.chapterRepo.SavePage(ctx, page)

	structResult, err := s.llm.StructureBlocks(ctx, ch.Subject, ocrResult.Blocks)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	if len(structResult.Items) == 0 {
		return nil, nil
	}

	// Create notions
	for _, notionName := range structResult.Notions {
		if _, exists := notionCache[notionName]; !exists {
			notion := &chapter.Notion{
				ID:        s.idGen.New(),
				ChapterID: ch.ID,
				Name:      notionName,
				SortOrder: len(notionCache),
				CreatedAt: now,
			}
			s.chapterRepo.SaveNotion(ctx, notion)
			notionCache[notionName] = notion
		}
	}

	// Create items
	var items []*chapter.Item
	for _, si := range structResult.Items {
		item := &chapter.Item{
			ID:         s.idGen.New(),
			ChapterID:  ch.ID,
			RevisionID: rev.ID,
			ItemType:   si.Type,
			Term:       &si.Term,
			Confidence: si.Confidence,
			Keywords:   si.Keywords,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		// Link to notion if exists
		if si.NotionName != "" {
			if notion, ok := notionCache[si.NotionName]; ok {
				item.NotionID = &notion.ID
			}
		}

		// Create steps for PROCEDURE items
		if si.Type == chapter.ItemProcedure {
			for j, stepContent := range si.Steps {
				item.Steps = append(item.Steps, chapter.ItemStep{
					ID:        s.idGen.New(),
					ItemID:    item.ID,
					StepOrder: j + 1,
					Content:   stepContent,
				})
			}
		}

		if err := s.chapterRepo.SaveItem(ctx, item); err != nil {
			return nil, fmt.Errorf("save item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
