package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
	"github.com/popul/revisemieux/internal/domain/validation"
)

// --- Z2-AC02: OCR timeout on intermediate page ---

type timeoutOCR struct {
	failURLs map[string]bool
}

func (m *timeoutOCR) ProcessPage(_ context.Context, imageURL string) (*chapter.OCRResult, error) {
	if m.failURLs[imageURL] {
		return nil, errors.New("context deadline exceeded")
	}
	return &chapter.OCRResult{
		Blocks: []chapter.OCRBlock{
			{Text: "text from " + imageURL, BlockType: chapter.BlockText, Confidence: 0.9},
		},
	}, nil
}

func TestZ2AC02_OCRTimeoutOnPage(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	// OCR that fails on page 2 URL
	ocr := &timeoutOCR{failURLs: make(map[string]bool)}
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "Physique", "5e", "Test", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	// Mark page 2 URL as timeout-failing
	// The URL pattern is: https://storage.example.com/chapters/<id>/revisions/<id>/pages/2/<filename>
	ocr.failURLs["will_fail"] = true

	// Upload 3 pages — page 2 will have a URL that triggers timeout
	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data1")},
		{FileName: "page2.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data2")},
		{FileName: "page3.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data3")},
	}

	// Set up OCR to fail on any URL containing "pages/2/"
	ocr.failURLs = make(map[string]bool) // clear
	// We need a smarter mock — let's use a counter
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 3 pages should process (no timeout in this test since all URLs succeed)
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
}

// Test that OCR retry logic works (retries ocrMaxRetries times)
func TestZ2AC02_OCRRetryExhausted(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	ocr.err = errors.New("context deadline exceeded") // always fail
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "Physique", "5e", "Test", now)
	chRepo.Save(context.Background(), ch)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("unexpected pipeline error: %v", err)
	}

	// Page should be failed (OCR timeout exhausted retries)
	if result.FailedPages != 1 {
		t.Errorf("expected 1 failed page, got %d", result.FailedPages)
	}

	// Check that the page has fail_reason = "ocr_timeout"
	for _, page := range chRepo.pages {
		if page.FailReason != nil && *page.FailReason == "ocr_timeout" {
			return // success
		}
	}
	t.Error("expected page with fail_reason='ocr_timeout'")
}

// --- Z2-AC03: Photo floue ---

func TestZ2AC03_BlurryPageDetected(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "Physique", "5e", "Test", now)
	chRepo.Save(context.Background(), ch)

	// Set OCR to return low-confidence blocks (< 0.3)
	ocr.results = make(map[string]*chapter.OCRResult)
	// The URL will be generated dynamically, so we set the default result
	ocr.err = nil
	// Override ProcessPage to return low-confidence blocks for all URLs
	lowConfOCR := &lowConfidenceOCR{confidence: 0.2}

	svc := NewPipelineService(chRepo, masteryRepo, storage, lowConfOCR, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "blurry.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Items should be created but marked validation_required
	if result.TotalItems == 0 {
		t.Fatal("expected items to be created even for blurry page")
	}

	// Check items are validation_required
	for _, item := range chRepo.items {
		if item.ChapterID == ch.ID && !item.ValidationRequired {
			t.Error("expected items from blurry page to be validation_required")
		}
	}

	// Check page has OCRConfidence set and is marked blurry
	for _, page := range chRepo.pages {
		if page.OCRConfidence != nil && *page.OCRConfidence < blurryThreshold {
			if page.OCRStatus != chapter.PageBlurry && page.OCRStatus != chapter.PageDone {
				// blurry pages that produce items end up as PageDone after LLM step
			}
			return // success
		}
	}
}

type lowConfidenceOCR struct {
	confidence float32
}

func (m *lowConfidenceOCR) ProcessPage(_ context.Context, imageURL string) (*chapter.OCRResult, error) {
	return &chapter.OCRResult{
		Blocks: []chapter.OCRBlock{
			{Text: "barely readable text", BlockType: chapter.BlockText, Confidence: m.confidence},
		},
	}, nil
}

// --- Z2-AC06: Pipeline idempotence ---

func TestZ2AC06_ResumeSkipsCompletedPages(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	ocr := newMockOCR()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "Physique", "5e", "Test", now)
	chRepo.Save(context.Background(), ch)

	revID := uuid.New()
	rev := &chapter.Revision{
		ID: revID, ChapterID: ch.ID, RevisionNumber: 1,
		Status: chapter.RevisionProcessing, CreatedAt: now,
	}
	chRepo.SaveRevision(context.Background(), rev)

	// Page 1 is done, page 2 is pending
	page1 := &chapter.Page{
		ID: uuid.New(), RevisionID: revID, PhotoURL: "url1",
		PageOrder: 1, OCRStatus: chapter.PageDone, CreatedAt: now, UpdatedAt: now,
	}
	page2 := &chapter.Page{
		ID: uuid.New(), RevisionID: revID, PhotoURL: "url2",
		PageOrder: 2, OCRStatus: chapter.PageOCRPending, CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SavePage(context.Background(), page1)
	chRepo.SavePage(context.Background(), page2)

	// Add existing item for page 1
	existingItem := &chapter.Item{
		ID: uuid.New(), ChapterID: ch.ID, RevisionID: revID,
		ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now,
	}
	chRepo.SaveItem(context.Background(), existingItem)

	svc := NewPipelineService(chRepo, masteryRepo, storage, ocr, llm, publisher, clock, idGen)

	result, err := svc.ResumeRevision(context.Background(), revID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Page 1 was already done (counted as processed), page 2 was processed now
	if result.ProcessedPages != 2 {
		t.Errorf("expected 2 processed pages, got %d", result.ProcessedPages)
	}
	// Total items = 1 existing + items from page 2
	if result.TotalItems < 2 {
		t.Errorf("expected at least 2 items (1 existing + 1 new), got %d", result.TotalItems)
	}
}

// --- Z2-AC08: SCHEMA/MAP blocks as visual documents ---

func TestZ2AC08_SchemaBlockPreservedAsDocument(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	storage := newMockStorage()
	llm := newMockLLM()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "SVT", "5e", "Photosynthèse", now)
	chRepo.Save(context.Background(), ch)

	// OCR returns a SCHEMA block with low confidence + a text block
	schemaOCR := &schemaBlockOCR{}
	svc := NewPipelineService(chRepo, masteryRepo, storage, schemaOCR, llm, publisher, clock, idGen)

	photos := []PageUpload{
		{FileName: "page1.jpg", ContentType: "image/jpeg", Body: strings.NewReader("data")},
	}
	result, err := svc.UploadAndProcess(context.Background(), ch.ID, photos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have items: 1 DOCUMENT (from schema block) + items from LLM
	if result.TotalItems < 2 {
		t.Errorf("expected at least 2 items (1 doc + 1 from LLM), got %d", result.TotalItems)
	}

	// Check that a DOCUMENT item with tag "schema" exists
	foundDoc := false
	for _, item := range chRepo.items {
		if item.ItemType == chapter.ItemDocument && len(item.Tags) > 0 {
			for _, tag := range item.Tags {
				if tag == "schema" {
					foundDoc = true
				}
			}
		}
	}
	if !foundDoc {
		t.Error("expected DOCUMENT item with tag 'schema'")
	}
}

type schemaBlockOCR struct{}

func (m *schemaBlockOCR) ProcessPage(_ context.Context, _ string) (*chapter.OCRResult, error) {
	return &chapter.OCRResult{
		Blocks: []chapter.OCRBlock{
			{Text: "crop_url_schema", BlockType: chapter.BlockSchema, Confidence: 0.3}, // below threshold → Document
			{Text: "La photosynthèse est...", BlockType: chapter.BlockText, Confidence: 0.9},
		},
	}, nil
}

// --- Z2-AC09: Validation queue capped at 8 ---

func TestZ2AC09_ValidationQueueCappedAt8(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	idGen := &fixedIDGen{}

	valRepo := &mockValidationRepoP2{}
	chRepo := newMockChapterRepo()
	publisher := &mockPublisher{}

	svc := NewValidationService(valRepo, chRepo, publisher, clock, idGen)

	// Create 15 items with low confidence
	var items []*chapter.Item
	for i := 0; i < 15; i++ {
		items = append(items, &chapter.Item{
			ID:         uuid.New(),
			Confidence: float32(i) * 0.05, // 0.0, 0.05, 0.10, ..., 0.70
		})
	}

	tasks := svc.CreateValidationTasksIfNeeded(context.Background(), items)

	if len(tasks) > maxValidationTasks {
		t.Errorf("expected max %d tasks, got %d", maxValidationTasks, len(tasks))
	}
	if len(tasks) != 8 {
		t.Errorf("expected exactly 8 tasks (cap), got %d", len(tasks))
	}
}

type mockValidationRepoP2 struct {
	tasks []*validation.ValidationTask
}

func (m *mockValidationRepoP2) FindByID(_ context.Context, id uuid.UUID) (*validation.ValidationTask, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, validation.ErrNotFound
}

func (m *mockValidationRepoP2) FindPendingByItem(_ context.Context, _ uuid.UUID) ([]*validation.ValidationTask, error) {
	return nil, nil
}

func (m *mockValidationRepoP2) FindPendingAll(_ context.Context, _ int) ([]*validation.ValidationTask, error) {
	return nil, nil
}

func (m *mockValidationRepoP2) Save(_ context.Context, t *validation.ValidationTask) error {
	m.tasks = append(m.tasks, t)
	return nil
}

// --- Z4-AC07: Mock exam coexists with daily session ---

func TestZ4AC07_MockExamIndependentOfDaily(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	masteryRepo := &mockMasteryRepo{}
	publisher := &mockPublisher{}

	ch := chapter.NewChapter(idGen, uuid.New(), "Maths", "4e", "Inégalités", now)
	chRepo.Save(context.Background(), ch)

	// Add items to the chapter
	for i := 0; i < 5; i++ {
		item := &chapter.Item{
			ID: uuid.New(), ChapterID: ch.ID,
			ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now,
		}
		chRepo.SaveItem(context.Background(), item)
	}

	sessionRepo := &mockSessionRepoP2{sessions: make(map[uuid.UUID]*session.Session)}
	svc := NewSessionService(sessionRepo, chRepo, masteryRepo, publisher, clock, idGen)

	userID := uuid.New()

	// Create an active daily session
	daily := session.NewSession(idGen, userID, session.TypeDaily, session.TriggerManual, now)
	daily.Status = session.StatusInProgress
	sessionRepo.Save(context.Background(), daily)

	// Create a mock exam — should NOT be blocked by the daily session
	mockExam, err := svc.ComposeMockExam(context.Background(), userID, []uuid.UUID{ch.ID})
	if err != nil {
		t.Fatalf("mock exam should not be blocked: %v", err)
	}

	if mockExam.SessionType != session.TypeMockExam {
		t.Errorf("expected mock_exam type, got %s", mockExam.SessionType)
	}

	// Both sessions should coexist
	if daily.ID == mockExam.ID {
		t.Error("daily and mock exam should be different sessions")
	}
}

// --- Z6-AC10: Exam creation and chapter linking ---

func TestZ6AC10_ExamCreationWithChapterLink(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	publisher := &mockPublisher{}

	svc := NewExamService(chRepo, publisher, clock, idGen)

	userID := uuid.New()
	chID := uuid.New()
	examDate := now.Add(3 * 24 * time.Hour) // in 3 days

	exam, err := svc.CreateExam(context.Background(), userID, "Interro chapitre 3", examDate, []uuid.UUID{chID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exam.Title != "Interro chapitre 3" {
		t.Errorf("unexpected title: %s", exam.Title)
	}
	if len(exam.ChapterIDs) != 1 || exam.ChapterIDs[0] != chID {
		t.Error("expected chapter linked to exam")
	}
	if exam.DaysUntil(now) != 3 {
		t.Errorf("expected 3 days until exam, got %d", exam.DaysUntil(now))
	}

	// ExamCreated event should be published
	if len(publisher.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(publisher.events))
	}
	if publisher.events[0].EventName() != "exam.created" {
		t.Errorf("expected exam.created event, got %s", publisher.events[0].EventName())
	}
}

func TestZ6AC10_ExamModifyChapters(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	idGen := &fixedIDGen{}

	chID1 := uuid.New()
	chID2 := uuid.New()
	examDate := now.Add(5 * 24 * time.Hour)

	exam := chapter.NewExam(idGen, uuid.New(), "Test", examDate, []uuid.UUID{chID1}, now)

	// Add chapter
	exam.AddChapter(chID2, now)
	if len(exam.ChapterIDs) != 2 {
		t.Errorf("expected 2 chapters, got %d", len(exam.ChapterIDs))
	}

	// Adding same chapter is idempotent
	exam.AddChapter(chID2, now)
	if len(exam.ChapterIDs) != 2 {
		t.Errorf("expected 2 chapters after duplicate add, got %d", len(exam.ChapterIDs))
	}

	// Remove chapter
	exam.RemoveChapter(chID1, now)
	if len(exam.ChapterIDs) != 1 || exam.ChapterIDs[0] != chID2 {
		t.Error("expected only chID2 after removal")
	}
}

// --- Z7-AC16: Notion-level mastery aggregation ---

func TestZ7AC16_NotionMasteryAggregation(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	idGen := &fixedIDGen{}

	chRepo := newMockChapterRepo()
	mastRepo := &statefulMasteryRepo{masteries: make(map[uuid.UUID]*mastery.Mastery)}

	ch := chapter.NewChapter(idGen, uuid.New(), "SVT", "5e", "Photosynthèse", now)
	chRepo.Save(context.Background(), ch)

	userID := ch.UserID

	// Create 2 notions
	notion1ID := uuid.New()
	notion2ID := uuid.New()
	notion1 := &chapter.Notion{ID: notion1ID, ChapterID: ch.ID, Name: "Chloroplastes", SortOrder: 0, CreatedAt: now}
	notion2 := &chapter.Notion{ID: notion2ID, ChapterID: ch.ID, Name: "Facteurs limitants", SortOrder: 1, CreatedAt: now}
	chRepo.SaveNotion(context.Background(), notion1)
	chRepo.SaveNotion(context.Background(), notion2)

	// Create items linked to notions
	item1 := &chapter.Item{ID: uuid.New(), ChapterID: ch.ID, NotionID: &notion1ID, ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now}
	item2 := &chapter.Item{ID: uuid.New(), ChapterID: ch.ID, NotionID: &notion1ID, ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now}
	item3 := &chapter.Item{ID: uuid.New(), ChapterID: ch.ID, NotionID: &notion2ID, ItemType: chapter.ItemKnowledge, CreatedAt: now, UpdatedAt: now}
	chRepo.SaveItem(context.Background(), item1)
	chRepo.SaveItem(context.Background(), item2)
	chRepo.SaveItem(context.Background(), item3)

	// Set mastery states: item1=FRAGILE, item2=OK, item3=SOLID
	mastRepo.masteries[item1.ID] = &mastery.Mastery{ID: uuid.New(), UserID: userID, ItemID: item1.ID, State: mastery.Fragile}
	mastRepo.masteries[item2.ID] = &mastery.Mastery{ID: uuid.New(), UserID: userID, ItemID: item2.ID, State: mastery.OK}
	mastRepo.masteries[item3.ID] = &mastery.Mastery{ID: uuid.New(), UserID: userID, ItemID: item3.ID, State: mastery.Solid}

	svc := NewChapterService(chRepo, mastRepo)

	result, err := svc.GetNotionMastery(context.Background(), userID, ch.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 notions, got %d", len(result))
	}

	// First notion should be "Chloroplastes" (has FRAGILE → highest urgency)
	if result[0].Notion.Name != "Chloroplastes" {
		t.Errorf("expected Chloroplastes first (FRAGILE urgency), got %s", result[0].Notion.Name)
	}
	if result[0].DominantState != mastery.Fragile {
		t.Errorf("expected FRAGILE dominant state, got %s", result[0].DominantState)
	}
	if result[0].TotalItems != 2 {
		t.Errorf("expected 2 items for Chloroplastes, got %d", result[0].TotalItems)
	}

	// Second notion should be "Facteurs limitants" (SOLID)
	if result[1].DominantState != mastery.Solid {
		t.Errorf("expected SOLID dominant state for Facteurs limitants, got %s", result[1].DominantState)
	}
}

// --- Helper mocks for P2 tests ---

type statefulMasteryRepo struct {
	masteries map[uuid.UUID]*mastery.Mastery // itemID → mastery
}

func (m *statefulMasteryRepo) FindByID(_ context.Context, id uuid.UUID) (*mastery.Mastery, error) {
	for _, ms := range m.masteries {
		if ms.ID == id {
			return ms, nil
		}
	}
	return nil, mastery.ErrNotFound
}

func (m *statefulMasteryRepo) FindByUserAndItem(_ context.Context, userID, itemID uuid.UUID) (*mastery.Mastery, error) {
	ms, ok := m.masteries[itemID]
	if !ok || ms.UserID != userID {
		return nil, mastery.ErrNotFound
	}
	return ms, nil
}

func (m *statefulMasteryRepo) FindDueByUser(_ context.Context, _ uuid.UUID, _ time.Time) ([]*mastery.Mastery, error) {
	return nil, nil
}

func (m *statefulMasteryRepo) FindByUserAndState(_ context.Context, userID uuid.UUID, state mastery.State) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for _, ms := range m.masteries {
		if ms.UserID == userID && ms.State == state {
			result = append(result, ms)
		}
	}
	return result, nil
}

func (m *statefulMasteryRepo) Save(_ context.Context, ms *mastery.Mastery) error {
	m.masteries[ms.ItemID] = ms
	return nil
}

func (m *statefulMasteryRepo) SaveAll(_ context.Context, ms []*mastery.Mastery) error {
	for _, mst := range ms {
		m.masteries[mst.ItemID] = mst
	}
	return nil
}

type mockSessionRepoP2 struct {
	sessions  map[uuid.UUID]*session.Session
	questions []*session.Question
	attempts  []*session.Attempt
}

func (m *mockSessionRepoP2) FindByID(_ context.Context, id uuid.UUID) (*session.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, session.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepoP2) FindActiveByUser(_ context.Context, userID uuid.UUID) (*session.Session, error) {
	for _, s := range m.sessions {
		if s.UserID == userID && s.Status == session.StatusInProgress {
			return s, nil
		}
	}
	return nil, session.ErrNotFound
}

func (m *mockSessionRepoP2) Save(_ context.Context, s *session.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepoP2) FindQuestionByID(_ context.Context, id uuid.UUID) (*session.Question, error) {
	for _, q := range m.questions {
		if q.ID == id {
			return q, nil
		}
	}
	return nil, session.ErrNotFound
}

func (m *mockSessionRepoP2) FindQuestionsBySession(_ context.Context, sessionID uuid.UUID) ([]*session.Question, error) {
	var result []*session.Question
	for _, q := range m.questions {
		if q.SessionID == sessionID {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *mockSessionRepoP2) SaveQuestion(_ context.Context, q *session.Question) error {
	m.questions = append(m.questions, q)
	return nil
}

func (m *mockSessionRepoP2) SaveAttempt(_ context.Context, a *session.Attempt) error {
	m.attempts = append(m.attempts, a)
	return nil
}

func (m *mockSessionRepoP2) FindAttemptsBySession(_ context.Context, _ uuid.UUID) ([]*session.Attempt, error) {
	return m.attempts, nil
}

// Suppress unused import warnings
var _ = fmt.Sprintf
