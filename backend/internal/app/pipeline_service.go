package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

const (
	// ocrMaxRetries is the max retries for OCR timeout (Z2-AC02).
	ocrMaxRetries = 3
	// ocrTimeout is the per-page OCR timeout.
	ocrTimeout = 10 * time.Second
	// blurryThreshold is the global confidence below which a page is marked blurry (Z2-AC03).
	blurryThreshold float32 = 0.3
	// schemaConfidenceThreshold: blocks below this are kept as visual documents (Z2-AC08).
	schemaConfidenceThreshold float32 = 0.5
)

// ErrOCRTimeout is returned when OCR times out after retries.
var ErrOCRTimeout = errors.New("pipeline: OCR timeout after retries")

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
	Recovery       *RecoveryInfo // Z8-AC03: non-nil when first upload fails
}

// RecoveryInfo provides context for the recovery screen (Z8-AC03).
type RecoveryInfo struct {
	IsFirstUpload   bool   // true if this was the user's first ever upload
	CanRetry        bool   // true: user can re-take photos
	CanContinue     bool   // true if some items were generated despite low confidence
	HasDemoChapter  bool   // true if demo chapter available as fallback
	Message         string // empathetic message
}

// PageProgress represents the progress after processing a single page.
type PageProgress struct {
	RevisionID     uuid.UUID
	PageOrder      int
	TotalPages     int
	Status         string // "done", "no_items", "failed"
	ItemsGenerated int
}

// PipelineService orchestrates the J0 pipeline: upload → OCR → structuration → items.
type PipelineService struct {
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
	storage     chapter.Storage
	ocr         chapter.OCRService
	llm         chapter.LLMService
	fidelity    chapter.FidelityChecker // Z3-AC10: optional fidelity checker
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
	onProgress  func(PageProgress)
}

// OnPageProgress sets a callback invoked after each page is processed.
// Used for SSE streaming (Z2-AC10).
func (s *PipelineService) OnPageProgress(fn func(PageProgress)) {
	s.onProgress = fn
}

// SetFidelityChecker sets an optional fidelity checker (Z3-AC10).
func (s *PipelineService) SetFidelityChecker(fc chapter.FidelityChecker) {
	s.fidelity = fc
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

	// 2. Create revision (increment number based on existing revisions)
	revCount, err := s.chapterRepo.CountRevisionsByChapter(ctx, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: count revisions: %w", err)
	}
	rev := &chapter.Revision{
		ID:             s.idGen.New(),
		ChapterID:      ch.ID,
		RevisionNumber: revCount + 1,
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
			s.notifyProgress(PageProgress{
				RevisionID: rev.ID, PageOrder: page.PageOrder,
				TotalPages: len(pages), Status: "failed",
			})
			continue
		}

		if len(items) == 0 {
			// Z2-AC04: no items extracted
			page.OCRStatus = chapter.PageNoItems
			page.UpdatedAt = now
			s.chapterRepo.SavePage(ctx, page)
			result.ProcessedPages++
			s.notifyProgress(PageProgress{
				RevisionID: rev.ID, PageOrder: page.PageOrder,
				TotalPages: len(pages), Status: "no_items",
			})
			continue
		}

		page.OCRStatus = chapter.PageDone
		page.UpdatedAt = now
		s.chapterRepo.SavePage(ctx, page)
		result.ProcessedPages++
		allItems = append(allItems, items...)
		s.notifyProgress(PageProgress{
			RevisionID: rev.ID, PageOrder: page.PageOrder,
			TotalPages: len(pages), Status: "done",
			ItemsGenerated: len(items),
		})
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

	// Z8-AC03: Build recovery info if pipeline failed or produced 0 items
	if rev.Status == chapter.RevisionFailed || result.TotalItems == 0 {
		isFirst := s.isFirstUpload(ctx, ch.UserID, ch.ID)
		hasDemoChapter := s.hasDemoChapter(ctx, ch.UserID)
		result.Recovery = &RecoveryInfo{
			IsFirstUpload:  isFirst,
			CanRetry:       true,
			CanContinue:    false,
			HasDemoChapter: hasDemoChapter,
			Message:        "Les photos sont un peu difficiles à lire. Pas de panique, ça arrive souvent au début !",
		}
	} else if rev.Status == chapter.RevisionPartial {
		// Some items generated despite issues — can continue in degraded mode
		isFirst := s.isFirstUpload(ctx, ch.UserID, ch.ID)
		if result.FailedPages > 0 && isFirst {
			result.Recovery = &RecoveryInfo{
				IsFirstUpload:  true,
				CanRetry:       true,
				CanContinue:    true,
				HasDemoChapter: s.hasDemoChapter(ctx, ch.UserID),
				Message:        "Certaines pages étaient difficiles à lire, mais on a quand même pu créer des questions !",
			}
		}
	}

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
	// OCR with retry (Z2-AC02)
	page.OCRStatus = chapter.PageOCRProcessing
	page.UpdatedAt = now
	s.chapterRepo.SavePage(ctx, page)

	ocrResult, err := s.ocrWithRetry(ctx, page.PhotoURL)
	if err != nil {
		reason := "ocr_timeout"
		page.FailReason = &reason
		return nil, fmt.Errorf("ocr: %w", err)
	}

	if len(ocrResult.Blocks) == 0 {
		return nil, nil
	}

	// Z2-AC03: Check global confidence for blurry detection
	globalConf := computeGlobalConfidence(ocrResult.Blocks)
	page.OCRConfidence = &globalConf
	if globalConf < blurryThreshold {
		page.OCRStatus = chapter.PageBlurry
		page.UpdatedAt = now
		s.chapterRepo.SavePage(ctx, page)
		// Still continue — items will be marked validation_required
	}

	// Z2-AC08: Separate SCHEMA/MAP blocks with low confidence → Document items
	var textBlocks []chapter.OCRBlock
	var visualDocItems []*chapter.Item
	for _, ob := range ocrResult.Blocks {
		if isVisualBlock(ob.BlockType) && ob.Confidence < schemaConfidenceThreshold {
			// Create Document item for visual block
			tag := strings.ToLower(string(ob.BlockType))
			cropURL := ob.Text // In real impl, this would be the crop URL
			item := &chapter.Item{
				ID:             s.idGen.New(),
				ChapterID:      ch.ID,
				RevisionID:     rev.ID,
				ItemType:       chapter.ItemDocument,
				Tags:           []string{tag},
				SourceImageURL: &cropURL,
				Confidence:     ob.Confidence,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			visualDocItems = append(visualDocItems, item)
		} else {
			textBlocks = append(textBlocks, ob)
		}
	}

	// Save visual document items
	for _, item := range visualDocItems {
		if err := s.chapterRepo.SaveItem(ctx, item); err != nil {
			return nil, fmt.Errorf("save doc item: %w", err)
		}
	}

	if len(textBlocks) == 0 {
		return visualDocItems, nil
	}

	// Structuration via LLM (only text blocks)
	page.OCRStatus = chapter.PageItemsGenerating
	page.UpdatedAt = now
	s.chapterRepo.SavePage(ctx, page)

	structResult, err := s.llm.StructureBlocks(ctx, ch.Subject, textBlocks)
	if err != nil {
		return visualDocItems, fmt.Errorf("llm: %w", err)
	}

	if len(structResult.Items) == 0 && len(visualDocItems) == 0 {
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

		// Z2-AC03: items from blurry pages need validation
		if globalConf < blurryThreshold {
			item.ValidationRequired = true
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

	// Z3-AC10: Fidelity check (step 7b) — verify items against source text
	if s.fidelity != nil {
		sourceText := collectSourceText(textBlocks)
		for _, item := range items {
			s.checkItemFidelity(ctx, item, sourceText, now)
		}
	}

	return append(visualDocItems, items...), nil
}

// checkItemFidelity runs the LLM fidelity check on a single item (Z3-AC10).
// On timeout, the item is kept with nil fidelity_score (degraded mode).
func (s *PipelineService) checkItemFidelity(ctx context.Context, item *chapter.Item, sourceText string, now time.Time) {
	result, err := s.fidelity.CheckFidelity(ctx, item, sourceText)
	if err != nil {
		// Timeout or error → degraded mode, continue normally
		return
	}

	item.FidelityScore = &result.Score

	if result.Score < 0.5 {
		// Hallucinated — flag and require validation
		flag := "low"
		item.FidelityFlag = &flag
		item.ValidationRequired = true
	}
	// Score >= 0.7 → faithful, no action needed
	// Score 0.5-0.7 → neutral zone, no flag

	item.UpdatedAt = now
	s.chapterRepo.SaveItem(ctx, item)
}

// isFirstUpload checks if this is the user's first non-demo chapter upload (Z8-AC03).
func (s *PipelineService) isFirstUpload(ctx context.Context, userID, chapterID uuid.UUID) bool {
	chapters, err := s.chapterRepo.FindByUser(ctx, userID, false)
	if err != nil {
		return true // assume first on error
	}
	realChapters := 0
	for _, ch := range chapters {
		if !ch.IsDemo {
			realChapters++
		}
	}
	return realChapters <= 1 // this chapter is the only real one
}

// hasDemoChapter checks if the user has an active demo chapter (Z8-AC03).
func (s *PipelineService) hasDemoChapter(ctx context.Context, userID uuid.UUID) bool {
	chapters, err := s.chapterRepo.FindByUser(ctx, userID, false)
	if err != nil {
		return false
	}
	for _, ch := range chapters {
		if ch.IsDemo {
			return true
		}
	}
	return false
}

// collectSourceText concatenates OCR block text for fidelity comparison.
func collectSourceText(blocks []chapter.OCRBlock) string {
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(b.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s *PipelineService) notifyProgress(p PageProgress) {
	if s.onProgress != nil {
		s.onProgress(p)
	}
}

// ocrWithRetry calls OCR with timeout and retries (Z2-AC02).
func (s *PipelineService) ocrWithRetry(ctx context.Context, imageURL string) (*chapter.OCRResult, error) {
	var lastErr error
	for attempt := 0; attempt <= ocrMaxRetries; attempt++ {
		ocrCtx, cancel := context.WithTimeout(ctx, ocrTimeout)
		result, err := s.ocr.ProcessPage(ocrCtx, imageURL)
		cancel()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%w: %v", ErrOCRTimeout, lastErr)
}

// computeGlobalConfidence returns the average confidence across all blocks (Z2-AC03).
func computeGlobalConfidence(blocks []chapter.OCRBlock) float32 {
	if len(blocks) == 0 {
		return 0
	}
	var total float32
	for _, b := range blocks {
		total += b.Confidence
	}
	return total / float32(len(blocks))
}

// isVisualBlock returns true for SCHEMA/MAP block types (Z2-AC08).
func isVisualBlock(bt chapter.BlockType) bool {
	switch bt {
	case chapter.BlockSchema, chapter.BlockMap:
		return true
	}
	return false
}

// ResumeRevision resumes an interrupted pipeline from incomplete steps (Z2-AC06).
// Reuses already-processed pages and only processes pending ones.
func (s *PipelineService) ResumeRevision(ctx context.Context, revisionID uuid.UUID) (*PipelineResult, error) {
	now := s.clock.Now()

	rev, err := s.chapterRepo.FindRevisionByID(ctx, revisionID)
	if err != nil {
		return nil, fmt.Errorf("pipeline.resume: find revision: %w", err)
	}
	if rev.Status != chapter.RevisionProcessing && rev.Status != chapter.RevisionPartial {
		return nil, fmt.Errorf("pipeline.resume: revision status %s is not resumable", rev.Status)
	}

	ch, err := s.chapterRepo.FindByID(ctx, rev.ChapterID)
	if err != nil {
		return nil, fmt.Errorf("pipeline.resume: find chapter: %w", err)
	}

	pages, err := s.chapterRepo.FindPagesByRevision(ctx, revisionID)
	if err != nil {
		return nil, fmt.Errorf("pipeline.resume: find pages: %w", err)
	}

	result := &PipelineResult{RevisionID: revisionID, TotalPages: len(pages)}
	var allItems []*chapter.Item
	notionCache := make(map[string]*chapter.Notion)

	// Load existing notions
	existingNotions, _ := s.chapterRepo.FindNotionsByChapter(ctx, ch.ID)
	for _, n := range existingNotions {
		notionCache[n.Name] = n
	}

	// Count already-done pages and items
	existingItems, _ := s.chapterRepo.FindItemsByChapter(ctx, ch.ID, true)
	for _, item := range existingItems {
		if item.RevisionID == revisionID {
			allItems = append(allItems, item)
		}
	}

	for _, page := range pages {
		switch page.OCRStatus {
		case chapter.PageDone, chapter.PageNoItems, chapter.PageBlurry:
			result.ProcessedPages++
			continue
		case chapter.PageFailed, chapter.PageItemsFailed:
			result.FailedPages++
			continue
		}

		// Re-process pending/processing pages
		items, err := s.processPage(ctx, ch, rev, page, notionCache, now)
		if err != nil {
			page.OCRStatus = chapter.PageItemsFailed
			page.UpdatedAt = now
			s.chapterRepo.SavePage(ctx, page)
			result.FailedPages++
			continue
		}
		if len(items) == 0 {
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

	// Update revision status
	if result.TotalItems == 0 {
		rev.Status = chapter.RevisionFailed
	} else if result.FailedPages > 0 {
		rev.Status = chapter.RevisionPartial
	} else {
		rev.Status = chapter.RevisionReady
	}
	s.chapterRepo.SaveRevision(ctx, rev)

	return result, nil
}

// RevisionProgress holds the current progress of a revision's pipeline (Z8-AC04).
type RevisionProgress struct {
	RevisionID     uuid.UUID
	Status         string
	TotalPages     int
	ProcessedPages int
	FailedPages    int
	TotalItems     int
	Phase          string // "reading", "generating", "done"
	PhaseMessage   string
}

// GetRevisionProgress returns the current pipeline progress for a revision (Z8-AC04).
func (s *PipelineService) GetRevisionProgress(ctx context.Context, revisionID uuid.UUID) (*RevisionProgress, error) {
	rev, err := s.chapterRepo.FindRevisionByID(ctx, revisionID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: find revision: %w", err)
	}

	pages, err := s.chapterRepo.FindPagesByRevision(ctx, revisionID)
	if err != nil {
		return nil, fmt.Errorf("pipeline: find pages: %w", err)
	}

	processed := 0
	failed := 0
	totalItems := 0
	for _, p := range pages {
		switch p.OCRStatus {
		case chapter.PageDone:
			processed++
		case chapter.PageNoItems:
			processed++
		case chapter.PageFailed, chapter.PageItemsFailed:
			failed++
		}
	}

	// Count items for this revision
	items, err := s.chapterRepo.FindItemsByChapter(ctx, rev.ChapterID, true)
	if err == nil {
		for _, item := range items {
			if item.RevisionID == revisionID {
				totalItems++
			}
		}
	}

	// Determine phase and message
	phase := "reading"
	msg := "Lecture de tes pages…"
	if processed > 0 || failed > 0 {
		phase = "generating"
		msg = "Création des questions…"
	}
	if rev.Status == chapter.RevisionReady || rev.Status == chapter.RevisionPartial || rev.Status == chapter.RevisionFailed {
		phase = "done"
		msg = "Terminé"
	}

	return &RevisionProgress{
		RevisionID:     revisionID,
		Status:         string(rev.Status),
		TotalPages:     len(pages),
		ProcessedPages: processed,
		FailedPages:    failed,
		TotalItems:     totalItems,
		Phase:          phase,
		PhaseMessage:   msg,
	}, nil
}
