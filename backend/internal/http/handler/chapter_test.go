package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/app"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/http/dto"
	"github.com/popul/revisemieux/internal/http/handler"
	"github.com/popul/revisemieux/internal/http/middleware"
)

// --- Mocks ---

type mockChapterRepo struct {
	chapters []*chapter.Chapter
	items    map[uuid.UUID][]*chapter.Item // chapterID -> items
}

func (m *mockChapterRepo) FindByID(_ context.Context, id uuid.UUID) (*chapter.Chapter, error) {
	for _, ch := range m.chapters {
		if ch.ID == id {
			return ch, nil
		}
	}
	return nil, chapter.ErrNotFound
}

func (m *mockChapterRepo) FindByUser(_ context.Context, _ uuid.UUID, _ bool) ([]*chapter.Chapter, error) {
	return m.chapters, nil
}

func (m *mockChapterRepo) Save(_ context.Context, _ *chapter.Chapter) error { return nil }

func (m *mockChapterRepo) FindItemByID(_ context.Context, _ uuid.UUID) (*chapter.Item, error) {
	return nil, nil
}

func (m *mockChapterRepo) FindItemsByChapter(_ context.Context, chapterID uuid.UUID, _ bool) ([]*chapter.Item, error) {
	return m.items[chapterID], nil
}

func (m *mockChapterRepo) SaveItem(_ context.Context, _ *chapter.Item) error    { return nil }
func (m *mockChapterRepo) SaveItems(_ context.Context, _ []*chapter.Item) error { return nil }

func (m *mockChapterRepo) FindNotionsByChapter(_ context.Context, _ uuid.UUID) ([]*chapter.Notion, error) {
	return nil, nil
}

func (m *mockChapterRepo) SaveNotion(_ context.Context, _ *chapter.Notion) error { return nil }

func (m *mockChapterRepo) FindRevisionByID(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return nil, nil
}

func (m *mockChapterRepo) FindCurrentRevision(_ context.Context, _ uuid.UUID) (*chapter.Revision, error) {
	return nil, nil
}

func (m *mockChapterRepo) SaveRevision(_ context.Context, _ *chapter.Revision) error { return nil }

func (m *mockChapterRepo) FindPagesByRevision(_ context.Context, _ uuid.UUID) ([]*chapter.Page, error) {
	return nil, nil
}

func (m *mockChapterRepo) SavePage(_ context.Context, _ *chapter.Page) error { return nil }
func (m *mockChapterRepo) CountRevisionsByChapter(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

type mockMasteryRepo struct {
	masteries []*mastery.Mastery
}

func (m *mockMasteryRepo) FindByID(_ context.Context, _ uuid.UUID) (*mastery.Mastery, error) {
	return nil, nil
}

func (m *mockMasteryRepo) FindByUserAndItem(_ context.Context, userID, itemID uuid.UUID) (*mastery.Mastery, error) {
	for _, mst := range m.masteries {
		if mst.UserID == userID && mst.ItemID == itemID {
			return mst, nil
		}
	}
	return nil, mastery.ErrNotFound
}

func (m *mockMasteryRepo) FindDueByUser(_ context.Context, _ uuid.UUID, _ time.Time) ([]*mastery.Mastery, error) {
	return nil, nil
}

func (m *mockMasteryRepo) FindByItem(_ context.Context, itemID uuid.UUID) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for _, mst := range m.masteries {
		if mst.ItemID == itemID {
			result = append(result, mst)
		}
	}
	return result, nil
}

func (m *mockMasteryRepo) FindByUserAndState(_ context.Context, userID uuid.UUID, state mastery.State) ([]*mastery.Mastery, error) {
	var result []*mastery.Mastery
	for _, mst := range m.masteries {
		if mst.UserID == userID && mst.State == state {
			result = append(result, mst)
		}
	}
	return result, nil
}

func (m *mockMasteryRepo) FindAllByUser(_ context.Context, _ uuid.UUID) ([]*mastery.Mastery, error) {
	return nil, nil
}

func (m *mockMasteryRepo) Save(_ context.Context, _ *mastery.Mastery) error      { return nil }
func (m *mockMasteryRepo) SaveAll(_ context.Context, _ []*mastery.Mastery) error { return nil }

type stubIDGen struct{}

func (stubIDGen) New() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type stubClock struct{ now time.Time }

func (c stubClock) Now() time.Time { return c.now }

func setupChapterRouter(t *testing.T, chRepo chapter.Repository, mRepo mastery.Repository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	chapterSvc := app.NewChapterService(chRepo, mRepo)
	h := handler.NewChapter(chapterSvc, stubIDGen{}, stubClock{now: time.Now()})

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		// Inject a fake user ID for testing
		c.Set(middleware.ContextKeyUserID, testUserID)
		c.Next()
	})
	api.GET("/chapters", h.List)
	return r
}

var testUserID = uuid.MustParse("01234567-0123-0123-0123-0123456789ab")

func TestChapterList_WithItemCountAndMasteryBreakdown(t *testing.T) {
	chapterID := uuid.Must(uuid.NewV7())
	revisionID := uuid.Must(uuid.NewV7())

	item1 := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID, RevisionID: revisionID, ItemType: chapter.ItemKnowledge}
	item2 := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID, RevisionID: revisionID, ItemType: chapter.ItemKnowledge}
	item3 := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID, RevisionID: revisionID, ItemType: chapter.ItemProcedure}

	ch := &chapter.Chapter{
		ID:                chapterID,
		UserID:            testUserID,
		Subject:           "Maths",
		ClassLevel:        "6eme",
		Name:              "Fractions",
		CurrentRevisionID: &revisionID,
	}

	chRepo := &mockChapterRepo{
		chapters: []*chapter.Chapter{ch},
		items:    map[uuid.UUID][]*chapter.Item{chapterID: {item1, item2, item3}},
	}

	mRepo := &mockMasteryRepo{
		masteries: []*mastery.Mastery{
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: item1.ID, State: mastery.Unknown},
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: item2.ID, State: mastery.Fragile},
			{ID: uuid.Must(uuid.NewV7()), UserID: testUserID, ItemID: item3.ID, State: mastery.OK},
		},
	}

	r := setupChapterRouter(t, chRepo, mRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/chapters", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result []dto.ChapterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(result))
	}

	ch0 := result[0]
	if ch0.ItemCount != 3 {
		t.Errorf("expected item_count=3, got %d", ch0.ItemCount)
	}

	if ch0.MasteryBreakdown == nil {
		t.Fatal("expected mastery_breakdown to be present")
	}

	mb := ch0.MasteryBreakdown
	if mb.Unknown != 1 {
		t.Errorf("expected unknown=1, got %d", mb.Unknown)
	}
	if mb.Fragile != 1 {
		t.Errorf("expected fragile=1, got %d", mb.Fragile)
	}
	if mb.Ok != 1 {
		t.Errorf("expected ok=1, got %d", mb.Ok)
	}
	if mb.Solid != 0 {
		t.Errorf("expected solid=0, got %d", mb.Solid)
	}
}

func TestChapterList_NoItems_ReturnsZeroCountNilBreakdown(t *testing.T) {
	chapterID := uuid.Must(uuid.NewV7())

	ch := &chapter.Chapter{
		ID:         chapterID,
		UserID:     testUserID,
		Subject:    "Maths",
		ClassLevel: "6eme",
		Name:       "Fractions",
	}

	chRepo := &mockChapterRepo{
		chapters: []*chapter.Chapter{ch},
		items:    map[uuid.UUID][]*chapter.Item{},
	}
	mRepo := &mockMasteryRepo{}

	r := setupChapterRouter(t, chRepo, mRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/chapters", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []dto.ChapterResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if len(result) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(result))
	}

	if result[0].ItemCount != 0 {
		t.Errorf("expected item_count=0, got %d", result[0].ItemCount)
	}

	if result[0].MasteryBreakdown != nil {
		t.Error("expected mastery_breakdown to be nil for chapter with no items")
	}
}

func TestChapterList_ItemsWithNoMastery_AllCountAsUnknown(t *testing.T) {
	chapterID := uuid.Must(uuid.NewV7())
	revisionID := uuid.Must(uuid.NewV7())

	item1 := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID, RevisionID: revisionID}
	item2 := &chapter.Item{ID: uuid.Must(uuid.NewV7()), ChapterID: chapterID, RevisionID: revisionID}

	ch := &chapter.Chapter{
		ID:                chapterID,
		UserID:            testUserID,
		Subject:           "Maths",
		ClassLevel:        "6eme",
		Name:              "Fractions",
		CurrentRevisionID: &revisionID,
	}

	chRepo := &mockChapterRepo{
		chapters: []*chapter.Chapter{ch},
		items:    map[uuid.UUID][]*chapter.Item{chapterID: {item1, item2}},
	}
	// No masteries at all
	mRepo := &mockMasteryRepo{}

	r := setupChapterRouter(t, chRepo, mRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/chapters", nil)
	r.ServeHTTP(w, req)

	var result []dto.ChapterResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result[0].ItemCount != 2 {
		t.Errorf("expected item_count=2, got %d", result[0].ItemCount)
	}

	mb := result[0].MasteryBreakdown
	if mb == nil {
		t.Fatal("expected mastery_breakdown to be present")
	}
	// Items without a mastery record should count as UNKNOWN
	if mb.Unknown != 2 {
		t.Errorf("expected unknown=2, got %d", mb.Unknown)
	}
}
