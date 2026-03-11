package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// --- Mocks ---

type mockChapterRepo struct {
	chapters  map[uuid.UUID]*chapter.Chapter
	revisions map[uuid.UUID]*chapter.Revision
	pages     map[uuid.UUID]*chapter.Page
	items     map[uuid.UUID]*chapter.Item
	notions   map[uuid.UUID]*chapter.Notion
}

func newMockChapterRepo() *mockChapterRepo {
	return &mockChapterRepo{
		chapters:  make(map[uuid.UUID]*chapter.Chapter),
		revisions: make(map[uuid.UUID]*chapter.Revision),
		pages:     make(map[uuid.UUID]*chapter.Page),
		items:     make(map[uuid.UUID]*chapter.Item),
		notions:   make(map[uuid.UUID]*chapter.Notion),
	}
}

func (m *mockChapterRepo) FindByID(_ context.Context, id uuid.UUID) (*chapter.Chapter, error) {
	ch, ok := m.chapters[id]
	if !ok {
		return nil, chapter.ErrNotFound
	}
	return ch, nil
}

func (m *mockChapterRepo) FindByUser(_ context.Context, userID uuid.UUID, _ bool) ([]*chapter.Chapter, error) {
	var result []*chapter.Chapter
	for _, ch := range m.chapters {
		if ch.UserID == userID {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (m *mockChapterRepo) Save(_ context.Context, ch *chapter.Chapter) error {
	m.chapters[ch.ID] = ch
	return nil
}

func (m *mockChapterRepo) FindItemByID(_ context.Context, id uuid.UUID) (*chapter.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, chapter.ErrNotFound
	}
	return item, nil
}

func (m *mockChapterRepo) FindItemsByChapter(_ context.Context, chapterID uuid.UUID, includeArchived bool) ([]*chapter.Item, error) {
	var result []*chapter.Item
	for _, item := range m.items {
		if item.ChapterID == chapterID {
			if includeArchived || !item.Archived {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func (m *mockChapterRepo) SaveItem(_ context.Context, item *chapter.Item) error {
	m.items[item.ID] = item
	return nil
}

func (m *mockChapterRepo) SaveItems(_ context.Context, items []*chapter.Item) error {
	for _, item := range items {
		m.items[item.ID] = item
	}
	return nil
}

func (m *mockChapterRepo) FindNotionsByChapter(_ context.Context, chapterID uuid.UUID) ([]*chapter.Notion, error) {
	var result []*chapter.Notion
	for _, n := range m.notions {
		if n.ChapterID == chapterID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockChapterRepo) SaveNotion(_ context.Context, n *chapter.Notion) error {
	m.notions[n.ID] = n
	return nil
}

func (m *mockChapterRepo) FindRevisionByID(_ context.Context, id uuid.UUID) (*chapter.Revision, error) {
	rev, ok := m.revisions[id]
	if !ok {
		return nil, chapter.ErrNotFound
	}
	return rev, nil
}

func (m *mockChapterRepo) FindCurrentRevision(_ context.Context, chapterID uuid.UUID) (*chapter.Revision, error) {
	for _, ch := range m.chapters {
		if ch.ID == chapterID && ch.CurrentRevisionID != nil {
			rev, ok := m.revisions[*ch.CurrentRevisionID]
			if ok {
				return rev, nil
			}
		}
	}
	return nil, chapter.ErrNoRevision
}

func (m *mockChapterRepo) SaveRevision(_ context.Context, r *chapter.Revision) error {
	m.revisions[r.ID] = r
	return nil
}

func (m *mockChapterRepo) FindPagesByRevision(_ context.Context, revisionID uuid.UUID) ([]*chapter.Page, error) {
	var result []*chapter.Page
	for _, p := range m.pages {
		if p.RevisionID == revisionID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockChapterRepo) SavePage(_ context.Context, p *chapter.Page) error {
	m.pages[p.ID] = p
	return nil
}

type mockStorage struct {
	uploaded map[string]string // key -> url
}

func newMockStorage() *mockStorage {
	return &mockStorage{uploaded: make(map[string]string)}
}

func (m *mockStorage) Upload(_ context.Context, key string, _ string, _ io.Reader) (string, error) {
	url := "https://storage.example.com/" + key
	m.uploaded[key] = url
	return url, nil
}

type mockOCR struct {
	results map[string]*chapter.OCRResult // imageURL -> result
	err     error
}

func newMockOCR() *mockOCR {
	return &mockOCR{results: make(map[string]*chapter.OCRResult)}
}

func (m *mockOCR) ProcessPage(_ context.Context, imageURL string) (*chapter.OCRResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	r, ok := m.results[imageURL]
	if !ok {
		// Default: return a text block
		return &chapter.OCRResult{
			Blocks: []chapter.OCRBlock{
				{Text: "sample text from " + imageURL, BlockType: chapter.BlockText, Confidence: 0.9},
			},
		}, nil
	}
	return r, nil
}

type mockLLM struct {
	result *chapter.StructurationResult
	err    error
}

func newMockLLM() *mockLLM {
	return &mockLLM{
		result: &chapter.StructurationResult{
			Items: []chapter.StructuredItem{
				{
					Type:       chapter.ItemKnowledge,
					Term:       "vitesse",
					Keywords:   []string{"vitesse", "distance", "temps"},
					NotionName: "Mouvement et vitesse",
					Confidence: 0.85,
				},
			},
			Notions: []string{"Mouvement et vitesse"},
		},
	}
}

func (m *mockLLM) StructureBlocks(_ context.Context, _ string, _ []chapter.OCRBlock) (*chapter.StructurationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockMasteryRepo struct {
	saved []*mastery.Mastery
}

func (m *mockMasteryRepo) FindByID(_ context.Context, _ uuid.UUID) (*mastery.Mastery, error) {
	return nil, mastery.ErrNotFound
}
func (m *mockMasteryRepo) FindByUserAndItem(_ context.Context, _, _ uuid.UUID) (*mastery.Mastery, error) {
	return nil, mastery.ErrNotFound
}
func (m *mockMasteryRepo) FindDueByUser(_ context.Context, _ uuid.UUID, _ time.Time) ([]*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepo) FindByUserAndState(_ context.Context, _ uuid.UUID, _ mastery.State) ([]*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepo) Save(_ context.Context, ms *mastery.Mastery) error {
	m.saved = append(m.saved, ms)
	return nil
}
func (m *mockMasteryRepo) SaveAll(_ context.Context, ms []*mastery.Mastery) error {
	m.saved = append(m.saved, ms...)
	return nil
}

type mockPublisher struct {
	events []event.Event
}

func (m *mockPublisher) Publish(_ context.Context, events ...event.Event) error {
	m.events = append(m.events, events...)
	return nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type fixedIDGen struct{ ids []uuid.UUID; idx int }

func (g *fixedIDGen) New() uuid.UUID {
	if g.idx < len(g.ids) {
		id := g.ids[g.idx]
		g.idx++
		return id
	}
	return uuid.Must(uuid.NewV7())
}

// --- Tests ---

func TestPipelineService_UploadAndProcess_HappyPath(t *testing.T) {
	// GIVEN a chapter exists and user uploads 1 photo
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Mouvement et vitesse", now)
	if err := chRepo.Save(context.Background(), ch); err != nil {
		t.Fatal(err)
	}

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	// WHEN user uploads photos
	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("fake-image-data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)

	// THEN revision is created, pages are processed, items are generated
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}
	if result.RevisionID == uuid.Nil {
		t.Error("expected non-nil revision ID")
	}
	if result.TotalPages != 1 {
		t.Errorf("TotalPages = %d, want 1", result.TotalPages)
	}
	if result.ProcessedPages != 1 {
		t.Errorf("ProcessedPages = %d, want 1", result.ProcessedPages)
	}
	if result.TotalItems != 1 {
		t.Errorf("TotalItems = %d, want 1", result.TotalItems)
	}

	// Verify revision was saved
	rev, err := chRepo.FindRevisionByID(context.Background(), result.RevisionID)
	if err != nil {
		t.Fatalf("FindRevisionByID: %v", err)
	}
	if rev.Status != chapter.RevisionReady {
		t.Errorf("Revision status = %s, want READY", rev.Status)
	}

	// Verify chapter.CurrentRevisionID was set
	updatedCh, _ := chRepo.FindByID(context.Background(), ch.ID)
	if updatedCh.CurrentRevisionID == nil || *updatedCh.CurrentRevisionID != result.RevisionID {
		t.Error("chapter.CurrentRevisionID not set to new revision")
	}

	// Verify ItemsGenerated event was published
	if len(publisher.events) == 0 {
		t.Fatal("expected ItemsGenerated event")
	}
	igEvent, ok := publisher.events[0].(event.ItemsGenerated)
	if !ok {
		t.Fatalf("expected ItemsGenerated, got %T", publisher.events[0])
	}
	if igEvent.ChapterID != ch.ID {
		t.Error("ItemsGenerated.ChapterID mismatch")
	}

	// Verify mastery records were created (UNKNOWN state)
	if len(masteryRepo.saved) != 1 {
		t.Fatalf("expected 1 mastery saved, got %d", len(masteryRepo.saved))
	}
	if masteryRepo.saved[0].State != mastery.Unknown {
		t.Errorf("mastery state = %s, want UNKNOWN", masteryRepo.saved[0].State)
	}
}

// Z2-AC04: Aucun item généré sur une page
func TestPipelineService_Z2AC04_NoItemsExtracted(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := &mockLLM{
		result: &chapter.StructurationResult{Items: nil, Notions: nil},
	}
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Chapitre vide", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	// Page should be marked NO_ITEMS
	if result.TotalItems != 0 {
		t.Errorf("TotalItems = %d, want 0", result.TotalItems)
	}

	// Find the page and check status
	pages, _ := chRepo.FindPagesByRevision(context.Background(), result.RevisionID)
	if len(pages) != 1 {
		t.Fatal("expected 1 page")
	}
	if pages[0].OCRStatus != chapter.PageNoItems {
		t.Errorf("page status = %s, want NO_ITEMS", pages[0].OCRStatus)
	}
}

// Z2-AC05: LLM failure → page marked ITEMS_FAILED
func TestPipelineService_Z2AC05_LLMFailure(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := &mockLLM{err: fmt.Errorf("LLM quota exceeded")}
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Chapitre", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)

	// Pipeline should NOT fail globally — the page is marked failed, rest continues
	if err != nil {
		t.Fatalf("UploadAndProcess should not fail globally: %v", err)
	}
	if result.FailedPages != 1 {
		t.Errorf("FailedPages = %d, want 1", result.FailedPages)
	}

	// Page should be marked ITEMS_FAILED
	pages, _ := chRepo.FindPagesByRevision(context.Background(), result.RevisionID)
	if len(pages) != 1 {
		t.Fatal("expected 1 page")
	}
	if pages[0].OCRStatus != chapter.PageItemsFailed {
		t.Errorf("page status = %s, want ITEMS_FAILED", pages[0].OCRStatus)
	}
}

// Z2-AC07: 0 items total → revision status FAILED
func TestPipelineService_Z2AC07_ZeroItemsTotal(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	// All pages produce no items
	llm := &mockLLM{
		result: &chapter.StructurationResult{Items: nil, Notions: nil},
	}
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Chapitre", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		{FileName: "page2.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	// Revision should be FAILED since 0 items total
	rev, _ := chRepo.FindRevisionByID(context.Background(), result.RevisionID)
	if rev.Status != chapter.RevisionFailed {
		t.Errorf("revision status = %s, want FAILED", rev.Status)
	}

	// No ItemsGenerated event
	if len(publisher.events) != 0 {
		t.Errorf("expected no events, got %d", len(publisher.events))
	}
}

// Z2-AC01: Partial processing — some pages done, some failed
func TestPipelineService_Z2AC01_PartialProcessing(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()

	// OCR succeeds for both pages
	ocr := newMockOCR()

	// LLM: first call succeeds with 1 item, second call fails
	callCount := 0
	llm := &callCountLLM{
		results: []*chapter.StructurationResult{
			{
				Items:   []chapter.StructuredItem{{Type: chapter.ItemKnowledge, Term: "masse", Keywords: []string{"masse"}, NotionName: "Masse", Confidence: 0.9}},
				Notions: []string{"Masse"},
			},
			nil, // second call fails
		},
		errs: []error{nil, fmt.Errorf("LLM timeout")},
	}
	_ = callCount

	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Chapitre", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		{FileName: "page2.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	// Revision should be PARTIAL (some pages succeeded, some failed)
	rev, _ := chRepo.FindRevisionByID(context.Background(), result.RevisionID)
	if rev.Status != chapter.RevisionPartial {
		t.Errorf("revision status = %s, want PARTIAL", rev.Status)
	}
	if result.ProcessedPages != 1 {
		t.Errorf("ProcessedPages = %d, want 1", result.ProcessedPages)
	}
	if result.FailedPages != 1 {
		t.Errorf("FailedPages = %d, want 1", result.FailedPages)
	}
	if result.TotalItems != 1 {
		t.Errorf("TotalItems = %d, want 1", result.TotalItems)
	}
}

// callCountLLM returns different results for successive calls.
type callCountLLM struct {
	results []*chapter.StructurationResult
	errs    []error
	idx     int
}

func (m *callCountLLM) StructureBlocks(_ context.Context, _ string, _ []chapter.OCRBlock) (*chapter.StructurationResult, error) {
	i := m.idx
	m.idx++
	if i < len(m.errs) && m.errs[i] != nil {
		return nil, m.errs[i]
	}
	if i < len(m.results) {
		return m.results[i], nil
	}
	return &chapter.StructurationResult{}, nil
}
