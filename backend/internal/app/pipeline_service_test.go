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

func (m *mockChapterRepo) CountRevisionsByChapter(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
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
func (m *mockMasteryRepo) FindDueByUser(_ context.Context, userID uuid.UUID, before time.Time) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for _, ms := range m.saved {
		if ms.UserID == userID && ms.IsDue(before) {
			result = append(result, ms)
		}
	}
	return result, nil
}
func (m *mockMasteryRepo) FindByItem(_ context.Context, itemID uuid.UUID) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for _, ms := range m.saved {
		if ms.ItemID == itemID {
			result = append(result, ms)
		}
	}
	return result, nil
}
func (m *mockMasteryRepo) FindByUserAndState(_ context.Context, _ uuid.UUID, _ mastery.State) ([]*mastery.Mastery, error) {
	return nil, nil
}
func (m *mockMasteryRepo) FindAllByUser(_ context.Context, _ uuid.UUID) ([]*mastery.Mastery, error) {
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

type fixedIDGen struct {
	ids []uuid.UUID
	idx int
}

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

// Z2-AC10: progress callback is invoked per page
func TestPipelineService_Z2AC10_ProgressCallback(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique", "5e", "Chapitre", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	var progress []PageProgress
	svc.OnPageProgress(func(p PageProgress) {
		progress = append(progress, p)
	})

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		{FileName: "page2.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		{FileName: "page3.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	_, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	// Should have received 3 progress events
	if len(progress) != 3 {
		t.Fatalf("expected 3 progress events, got %d", len(progress))
	}

	// Check first event
	if progress[0].PageOrder != 1 {
		t.Errorf("progress[0].PageOrder = %d, want 1", progress[0].PageOrder)
	}
	if progress[0].TotalPages != 3 {
		t.Errorf("progress[0].TotalPages = %d, want 3", progress[0].TotalPages)
	}
	if progress[0].ItemsGenerated != 1 {
		t.Errorf("progress[0].ItemsGenerated = %d, want 1", progress[0].ItemsGenerated)
	}

	// Check last event
	if progress[2].PageOrder != 3 {
		t.Errorf("progress[2].PageOrder = %d, want 3", progress[2].PageOrder)
	}
}

// --- Fidelity check mock ---

type mockFidelityChecker struct {
	score float64
	err   error
}

func (m *mockFidelityChecker) CheckFidelity(_ context.Context, _ *chapter.Item, _ string) (*chapter.FidelityResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	flag := ""
	if m.score < 0.5 {
		flag = "low"
	}
	return &chapter.FidelityResult{Score: m.score, Flag: flag}, nil
}

// Z3-AC10 — Faithful item (score >= 0.7): no flag, no validation required
func TestZ3AC10_FidelityFaithful(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "SVT", "5e", "Photosynthese", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)
	svc.SetFidelityChecker(&mockFidelityChecker{score: 0.85})

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("fake")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}
	if result.TotalItems != 1 {
		t.Fatalf("expected 1 item, got %d", result.TotalItems)
	}

	// Find the item and check fidelity fields
	items, _ := chRepo.FindItemsByChapter(context.Background(), ch.ID, true)
	for _, item := range items {
		if item.ItemType == chapter.ItemKnowledge {
			if item.FidelityScore == nil || *item.FidelityScore != 0.85 {
				t.Errorf("fidelity_score: got %v, want 0.85", item.FidelityScore)
			}
			if item.FidelityFlag != nil {
				t.Errorf("fidelity_flag: got %v, want nil", item.FidelityFlag)
			}
			if item.ValidationRequired {
				t.Error("validation_required should be false for faithful item")
			}
		}
	}
}

// Z3-AC10 — Hallucinated item (score < 0.5): flag=low, validation required
func TestZ3AC10_FidelityHallucinated(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "SVT", "5e", "Photosynthese", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)
	svc.SetFidelityChecker(&mockFidelityChecker{score: 0.3})

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("fake")},
	}
	_, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	items, _ := chRepo.FindItemsByChapter(context.Background(), ch.ID, true)
	for _, item := range items {
		if item.ItemType == chapter.ItemKnowledge {
			if item.FidelityScore == nil || *item.FidelityScore != 0.3 {
				t.Errorf("fidelity_score: got %v, want 0.3", item.FidelityScore)
			}
			if item.FidelityFlag == nil || *item.FidelityFlag != "low" {
				t.Errorf("fidelity_flag: got %v, want 'low'", item.FidelityFlag)
			}
			if !item.ValidationRequired {
				t.Error("validation_required should be true for hallucinated item")
			}
		}
	}
}

// Z7-AC15 — Structuration LLM produces Items grouped by Notion (concept_tag)
//
// GIVEN: Le chapitre « Densité et masse volumique » contient des blocs OCR.
//
//	Le LLM attribue des concept_tags (NotionName) à chaque item.
//
// WHEN:  Le pipeline structure les blocs OCR via le LLM.
// THEN:  Les items sont créés avec des NotionIDs correspondant aux Notions,
//
//	et les Notions sont créées avec les bons noms.
func TestZ7AC15_StructurationLLM_ItemsEtNotions(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	// LLM returns multiple items grouped by distinct Notions (concept_tags)
	llm := &mockLLM{
		result: &chapter.StructurationResult{
			Items: []chapter.StructuredItem{
				{Type: chapter.ItemKnowledge, Term: "masse", Keywords: []string{"masse", "kg"}, NotionName: "Masse", Confidence: 0.9},
				{Type: chapter.ItemKnowledge, Term: "balance", Keywords: []string{"balance", "mesure"}, NotionName: "Masse", Confidence: 0.85},
				{Type: chapter.ItemKnowledge, Term: "volume", Keywords: []string{"volume", "litre"}, NotionName: "Volume", Confidence: 0.88},
				{Type: chapter.ItemProcedure, Term: "rho", Keywords: []string{"rho", "formule"}, NotionName: "Masse volumique (rho)", Confidence: 0.92, Steps: []string{"Identifier m", "Identifier V", "Calculer rho=m/V"}},
				{Type: chapter.ItemKnowledge, Term: "deplacement", Keywords: []string{"eau", "deplacement"}, NotionName: "Déplacement d'eau", Confidence: 0.87},
			},
			Notions: []string{"Masse", "Volume", "Masse volumique (rho)", "Déplacement d'eau"},
		},
	}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "Physique-Chimie", "5e", "Densité et masse volumique", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("fake-image-data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v", err)
	}

	// 5 items created
	if result.TotalItems != 5 {
		t.Errorf("TotalItems = %d, want 5", result.TotalItems)
	}

	// Verify 4 distinct Notions were created
	notions, err := chRepo.FindNotionsByChapter(context.Background(), ch.ID)
	if err != nil {
		t.Fatalf("FindNotionsByChapter: %v", err)
	}
	if len(notions) != 4 {
		t.Fatalf("expected 4 notions, got %d", len(notions))
	}

	notionNames := make(map[string]uuid.UUID)
	for _, n := range notions {
		notionNames[n.Name] = n.ID
	}

	expectedNotions := []string{"Masse", "Volume", "Masse volumique (rho)", "Déplacement d'eau"}
	for _, name := range expectedNotions {
		if _, ok := notionNames[name]; !ok {
			t.Errorf("missing notion %q", name)
		}
	}

	// Verify items are linked to correct Notions via NotionID
	items, err := chRepo.FindItemsByChapter(context.Background(), ch.ID, true)
	if err != nil {
		t.Fatalf("FindItemsByChapter: %v", err)
	}

	// Count items per notion
	itemsPerNotion := make(map[string]int)
	for _, item := range items {
		if item.NotionID != nil {
			for _, n := range notions {
				if n.ID == *item.NotionID {
					itemsPerNotion[n.Name]++
					break
				}
			}
		}
	}
	// "Masse" should have 2 items (masse + balance)
	if itemsPerNotion["Masse"] != 2 {
		t.Errorf("items for 'Masse' = %d, want 2", itemsPerNotion["Masse"])
	}
	// "Volume" should have 1 item
	if itemsPerNotion["Volume"] != 1 {
		t.Errorf("items for 'Volume' = %d, want 1", itemsPerNotion["Volume"])
	}

	// Verify PROCEDURE item has steps
	for _, item := range items {
		if item.ItemType == chapter.ItemProcedure {
			if len(item.Steps) == 0 {
				t.Error("PROCEDURE item should have steps")
			}
		}
	}

	// Verify mastery records created for all items
	if len(masteryRepo.saved) != 5 {
		t.Errorf("expected 5 mastery records, got %d", len(masteryRepo.saved))
	}
}

// Z8-AC03 — Recovery screen when first OCR fails
//
// GIVEN: L'élève vient de photographier son premier cours.
//
//	Le pipeline J0 échoue : 0 items valides générés.
//
// WHEN:  Le pipeline retourne un résultat vide.
// THEN:  Le résultat contient un RecoveryInfo avec :
//
//	(1) Message empathique
//	(3) CanRetry = true (reprendre les photos)
//	(5) HasDemoChapter = true si premier upload et demo disponible
func TestZ8AC03_RecoverySiPremierOCREchoue(t *testing.T) {

	t.Run("0 items sur premier upload avec demo → recovery complète", func(t *testing.T) {
		now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := fixedClock{t: now}
		idGen := &fixedIDGen{}
		userID := uuid.Must(uuid.NewV7())

		chRepo := newMockChapterRepo()
		storage := newMockStorage()
		ocr := newMockOCR()
		masteryRepo := &mockMasteryRepo{}
		publisher := &mockPublisher{}

		// LLM returns 0 items (OCR illisible)
		llm := &mockLLM{
			result: &chapter.StructurationResult{Items: nil, Notions: nil},
		}

		// Create a demo chapter (fallback available)
		demoChapter := chapter.NewChapter(idGen, userID, "Physique-Chimie", "5e", "Densité et masse volumique (démo)", now)
		demoChapter.IsDemo = true
		chRepo.Save(context.Background(), demoChapter)

		// Create the real chapter (first upload)
		realChapter := chapter.NewChapter(idGen, userID, "Physique-Chimie", "5e", "Mouvement et vitesse", now)
		chRepo.Save(context.Background(), realChapter)

		svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

		photos := []PageUpload{
			{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		}
		result, err := svc.UploadAndProcess(context.Background(), realChapter.ID, photos)
		if err != nil {
			t.Fatalf("UploadAndProcess should not fail globally: %v", err)
		}

		// THEN: Recovery info is populated
		if result.Recovery == nil {
			t.Fatal("expected Recovery info for 0 items")
		}
		if !result.Recovery.IsFirstUpload {
			t.Error("IsFirstUpload should be true")
		}
		if !result.Recovery.CanRetry {
			t.Error("CanRetry should be true (reprendre les photos)")
		}
		if result.Recovery.CanContinue {
			t.Error("CanContinue should be false (0 items)")
		}
		if !result.Recovery.HasDemoChapter {
			t.Error("HasDemoChapter should be true (fallback démo)")
		}
		if result.Recovery.Message == "" {
			t.Error("Message should not be empty (empathetic message)")
		}
	})

	t.Run("LLM failure sur premier upload → recovery avec message", func(t *testing.T) {
		now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := fixedClock{t: now}
		idGen := &fixedIDGen{}
		userID := uuid.Must(uuid.NewV7())

		chRepo := newMockChapterRepo()
		storage := newMockStorage()
		ocr := newMockOCR()
		masteryRepo := &mockMasteryRepo{}
		publisher := &mockPublisher{}

		// LLM fails (simulates timeout)
		llm := &mockLLM{err: fmt.Errorf("LLM timeout")}

		// Only one real chapter (first upload), no demo
		ch := chapter.NewChapter(idGen, userID, "Maths", "4e", "Pythagore", now)
		chRepo.Save(context.Background(), ch)

		svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

		photos := []PageUpload{
			{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		}
		result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
		if err != nil {
			t.Fatalf("UploadAndProcess should not fail globally: %v", err)
		}

		// THEN: Recovery info with no demo fallback
		if result.Recovery == nil {
			t.Fatal("expected Recovery info for LLM failure")
		}
		if !result.Recovery.IsFirstUpload {
			t.Error("IsFirstUpload should be true")
		}
		if result.Recovery.HasDemoChapter {
			t.Error("HasDemoChapter should be false (no demo chapter)")
		}
		if result.Recovery.CanRetry != true {
			t.Error("CanRetry should be true")
		}
	})

	t.Run("partial items sur premier upload → recovery avec CanContinue", func(t *testing.T) {
		now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := fixedClock{t: now}
		idGen := &fixedIDGen{}
		userID := uuid.Must(uuid.NewV7())

		chRepo := newMockChapterRepo()
		storage := newMockStorage()
		ocr := newMockOCR()
		masteryRepo := &mockMasteryRepo{}
		publisher := &mockPublisher{}

		// LLM: first page succeeds, second fails (partial)
		llm := &callCountLLM{
			results: []*chapter.StructurationResult{
				{
					Items:   []chapter.StructuredItem{{Type: chapter.ItemKnowledge, Term: "masse", Keywords: []string{"masse"}, NotionName: "Masse", Confidence: 0.9}},
					Notions: []string{"Masse"},
				},
				nil,
			},
			errs: []error{nil, fmt.Errorf("LLM error")},
		}

		ch := chapter.NewChapter(idGen, userID, "Physique", "5e", "Chapitre", now)
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

		// THEN: Recovery with CanContinue because some items were generated
		if result.Recovery == nil {
			t.Fatal("expected Recovery info for partial failure on first upload")
		}
		if !result.Recovery.CanContinue {
			t.Error("CanContinue should be true (some items generated)")
		}
		if !result.Recovery.CanRetry {
			t.Error("CanRetry should be true")
		}
	})

	t.Run("0 items sur upload non-premier → recovery sans IsFirstUpload", func(t *testing.T) {
		now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := fixedClock{t: now}
		idGen := &fixedIDGen{}
		userID := uuid.Must(uuid.NewV7())

		chRepo := newMockChapterRepo()
		storage := newMockStorage()
		ocr := newMockOCR()
		masteryRepo := &mockMasteryRepo{}
		publisher := &mockPublisher{}

		// LLM returns 0 items
		llm := &mockLLM{
			result: &chapter.StructurationResult{Items: nil, Notions: nil},
		}

		// Two real chapters → this is NOT the first upload
		ch1 := chapter.NewChapter(idGen, userID, "Maths", "5e", "Pythagore", now)
		chRepo.Save(context.Background(), ch1)
		ch2 := chapter.NewChapter(idGen, userID, "Physique", "5e", "Forces", now)
		chRepo.Save(context.Background(), ch2)

		svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

		photos := []PageUpload{
			{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
		}
		result, err := svc.UploadAndProcess(context.Background(), ch2.ID, photos)
		if err != nil {
			t.Fatalf("UploadAndProcess: %v", err)
		}

		// THEN: Recovery present but IsFirstUpload = false
		if result.Recovery == nil {
			t.Fatal("expected Recovery info for 0 items")
		}
		if result.Recovery.IsFirstUpload {
			t.Error("IsFirstUpload should be false (not first upload)")
		}
	})

	t.Run("retry illimité : pipeline réutilisable après échec", func(t *testing.T) {
		// Z8-AC03 : "Le nombre de retries n'est pas limité."
		now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := fixedClock{t: now}
		idGen := &fixedIDGen{}
		userID := uuid.Must(uuid.NewV7())

		chRepo := newMockChapterRepo()
		storage := newMockStorage()
		ocr := newMockOCR()
		masteryRepo := &mockMasteryRepo{}
		publisher := &mockPublisher{}

		llmRetry := &mockLLM{
			result: &chapter.StructurationResult{Items: nil, Notions: nil},
		}

		ch := chapter.NewChapter(idGen, userID, "Physique", "5e", "Chapitre", now)
		chRepo.Save(context.Background(), ch)

		svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llmRetry, publisher, clock, idGen)

		// Run pipeline 3 times (simulating retries) — should not block
		for i := 0; i < 3; i++ {
			photos := []PageUpload{
				{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
			}
			result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
			if err != nil {
				t.Fatalf("retry %d: UploadAndProcess error: %v", i, err)
			}
			if result.Recovery == nil {
				t.Fatalf("retry %d: expected Recovery info", i)
			}
		}
	})
}

// Z3-AC10 — Fidelity timeout: item kept with nil score, pipeline continues
func TestZ3AC10_FidelityTimeout(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.Must(uuid.NewV7()), "SVT", "5e", "Photosynthese", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)
	svc.SetFidelityChecker(&mockFidelityChecker{err: fmt.Errorf("timeout")})

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("fake")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("UploadAndProcess: %v (pipeline should continue on fidelity timeout)", err)
	}
	if result.TotalItems != 1 {
		t.Errorf("expected 1 item despite fidelity timeout, got %d", result.TotalItems)
	}

	// Item should have nil fidelity score
	items, _ := chRepo.FindItemsByChapter(context.Background(), ch.ID, true)
	for _, item := range items {
		if item.ItemType == chapter.ItemKnowledge {
			if item.FidelityScore != nil {
				t.Errorf("fidelity_score should be nil on timeout, got %v", *item.FidelityScore)
			}
		}
	}
}
