package chapter

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

// Z5-AC01: canonical identity key for items
func TestZ5AC01_CanonicalKey_SameItem(t *testing.T) {
	// GIVEN two items with same term (different casing/accents), same type, same pack_id
	// WHEN we compute their canonical keys
	// THEN the keys are equal
	key1 := CanonicalKey("Vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_AccentsNormalized(t *testing.T) {
	// Accents are stripped: "énergie" == "energie"
	key1 := CanonicalKey("Énergie cinétique", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("energie cinetique", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match after accent removal: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_TrimWhitespace(t *testing.T) {
	key1 := CanonicalKey("  vitesse  ", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match after trim: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_DifferentType(t *testing.T) {
	// Same term, different type → different keys
	key1 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemProcedure, strPtr("pack1"))
	if key1 == key2 {
		t.Errorf("keys should differ for different types: %q == %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_DifferentPackID(t *testing.T) {
	// Same term and type, different pack_id → different keys
	key1 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack2"))
	if key1 == key2 {
		t.Errorf("keys should differ for different pack_id: %q == %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_NilPackID(t *testing.T) {
	// nil pack_id matches nil pack_id
	key1 := CanonicalKey("vitesse", ItemKnowledge, nil)
	key2 := CanonicalKey("vitesse", ItemKnowledge, nil)
	if key1 != key2 {
		t.Errorf("keys should match with nil pack_id: %q != %q", key1, key2)
	}
	// nil vs non-nil → different
	key3 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 == key3 {
		t.Error("nil vs non-nil pack_id should differ")
	}
}

func strPtr(s string) *string { return &s }

// --- Test helpers ---

var testIDGen = event.UUIDv7Generator{}

// --- BlockType Value Object ---

func TestBlockType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		bt    BlockType
		valid bool
	}{
		{"TEXT", BlockText, true},
		{"PHOTO", BlockPhoto, true},
		{"SCHEMA", BlockSchema, true},
		{"MAP", BlockMap, true},
		{"GRAPH", BlockGraph, true},
		{"TABLE", BlockTable, true},
		{"CIRCUIT", BlockCircuit, true},
		{"DECORATIVE", BlockDecorative, true},
		{"invalid empty", BlockType(""), false},
		{"invalid unknown", BlockType("UNKNOWN"), false},
		{"invalid lowercase", BlockType("text"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bt.Valid(); got != tt.valid {
				t.Errorf("BlockType(%q).Valid() = %v, want %v", tt.bt, got, tt.valid)
			}
		})
	}
}

// --- ItemType Value Object ---

func TestItemType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		it    ItemType
		valid bool
	}{
		{"KNOWLEDGE", ItemKnowledge, true},
		{"PROCEDURE", ItemProcedure, true},
		{"DOCUMENT", ItemDocument, true},
		{"WRITING", ItemWriting, true},
		{"invalid empty", ItemType(""), false},
		{"invalid unknown", ItemType("QUIZ"), false},
		{"invalid lowercase", ItemType("knowledge"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.Valid(); got != tt.valid {
				t.Errorf("ItemType(%q).Valid() = %v, want %v", tt.it, got, tt.valid)
			}
		})
	}
}

func TestParseItemType_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  ItemType
	}{
		{"KNOWLEDGE", ItemKnowledge},
		{"PROCEDURE", ItemProcedure},
		{"DOCUMENT", ItemDocument},
		{"WRITING", ItemWriting},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseItemType(tt.input)
			if err != nil {
				t.Fatalf("ParseItemType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseItemType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseItemType_Invalid(t *testing.T) {
	tests := []string{"", "QUIZ", "knowledge", "invalid"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseItemType(input)
			if err == nil {
				t.Errorf("ParseItemType(%q) expected error, got nil", input)
			}
		})
	}
}

// --- NewChapter constructor ---

func TestNewChapter_FieldsInitialized(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	now := time.Now()

	ch := NewChapter(testIDGen, userID, "Physique", "4e", "Chapitre 1 - Forces", now)

	if ch.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if ch.UserID != userID {
		t.Errorf("UserID = %v, want %v", ch.UserID, userID)
	}
	if ch.Subject != "Physique" {
		t.Errorf("Subject = %q, want %q", ch.Subject, "Physique")
	}
	if ch.ClassLevel != "4e" {
		t.Errorf("ClassLevel = %q, want %q", ch.ClassLevel, "4e")
	}
	if ch.Name != "Chapitre 1 - Forces" {
		t.Errorf("Name = %q, want %q", ch.Name, "Chapitre 1 - Forces")
	}
	if ch.Archived {
		t.Error("new chapter should not be archived")
	}
	if ch.IsDemo {
		t.Error("new chapter should not be demo")
	}
	if ch.CurrentRevisionID != nil {
		t.Error("new chapter should have no current revision")
	}
	if ch.PackID != nil {
		t.Error("new chapter should have no pack ID")
	}
	if !ch.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", ch.CreatedAt, now)
	}
	if !ch.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", ch.UpdatedAt, now)
	}
}

// --- Item Archive / Unarchive ---

func TestItem_Archive(t *testing.T) {
	item := &Item{
		ID:        uuid.Must(uuid.NewV7()),
		Archived:  false,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
	now := time.Now()

	item.Archive(now)

	if !item.Archived {
		t.Error("item should be archived after Archive()")
	}
	if !item.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt should be updated to %v, got %v", now, item.UpdatedAt)
	}
}

func TestItem_Unarchive(t *testing.T) {
	item := &Item{
		ID:        uuid.Must(uuid.NewV7()),
		Archived:  true,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
	now := time.Now()

	item.Unarchive(now)

	if item.Archived {
		t.Error("item should not be archived after Unarchive()")
	}
	if !item.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt should be updated to %v, got %v", now, item.UpdatedAt)
	}
}

func TestItem_ArchiveThenUnarchive(t *testing.T) {
	item := &Item{ID: uuid.Must(uuid.NewV7()), Archived: false}
	now := time.Now()

	item.Archive(now)
	if !item.Archived {
		t.Fatal("should be archived")
	}

	later := now.Add(time.Minute)
	item.Unarchive(later)
	if item.Archived {
		t.Fatal("should be unarchived")
	}
	if !item.UpdatedAt.Equal(later) {
		t.Errorf("UpdatedAt = %v, want %v", item.UpdatedAt, later)
	}
}

// --- Exam ---

func TestNewExam_FieldsInitialized(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	ch1 := uuid.Must(uuid.NewV7())
	ch2 := uuid.Must(uuid.NewV7())
	now := time.Now()
	examDate := now.Add(7 * 24 * time.Hour)

	exam := NewExam(testIDGen, userID, "Controle Forces", examDate, []uuid.UUID{ch1, ch2}, now)

	if exam.ID == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if exam.UserID != userID {
		t.Errorf("UserID = %v, want %v", exam.UserID, userID)
	}
	if exam.Title != "Controle Forces" {
		t.Errorf("Title = %q, want %q", exam.Title, "Controle Forces")
	}
	if !exam.ExamDate.Equal(examDate) {
		t.Errorf("ExamDate = %v, want %v", exam.ExamDate, examDate)
	}
	if len(exam.ChapterIDs) != 2 {
		t.Fatalf("ChapterIDs length = %d, want 2", len(exam.ChapterIDs))
	}
	if exam.ChapterIDs[0] != ch1 || exam.ChapterIDs[1] != ch2 {
		t.Errorf("ChapterIDs = %v, want [%v, %v]", exam.ChapterIDs, ch1, ch2)
	}
}

func TestExam_AddChapter(t *testing.T) {
	exam := &Exam{
		ID:         uuid.Must(uuid.NewV7()),
		ChapterIDs: []uuid.UUID{uuid.Must(uuid.NewV7())},
		UpdatedAt:  time.Now().Add(-time.Hour),
	}
	newCh := uuid.Must(uuid.NewV7())
	now := time.Now()

	exam.AddChapter(newCh, now)

	if len(exam.ChapterIDs) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(exam.ChapterIDs))
	}
	if exam.ChapterIDs[1] != newCh {
		t.Errorf("second chapter = %v, want %v", exam.ChapterIDs[1], newCh)
	}
	if !exam.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt should be updated")
	}
}

func TestExam_AddChapter_AlreadyLinked(t *testing.T) {
	ch := uuid.Must(uuid.NewV7())
	oldTime := time.Now().Add(-time.Hour)
	exam := &Exam{
		ID:         uuid.Must(uuid.NewV7()),
		ChapterIDs: []uuid.UUID{ch},
		UpdatedAt:  oldTime,
	}

	exam.AddChapter(ch, time.Now())

	if len(exam.ChapterIDs) != 1 {
		t.Errorf("should not duplicate: got %d chapters", len(exam.ChapterIDs))
	}
	// UpdatedAt should NOT change since nothing was added
	if !exam.UpdatedAt.Equal(oldTime) {
		t.Error("UpdatedAt should not change when chapter already linked")
	}
}

func TestExam_RemoveChapter(t *testing.T) {
	ch1 := uuid.Must(uuid.NewV7())
	ch2 := uuid.Must(uuid.NewV7())
	exam := &Exam{
		ID:         uuid.Must(uuid.NewV7()),
		ChapterIDs: []uuid.UUID{ch1, ch2},
		UpdatedAt:  time.Now().Add(-time.Hour),
	}
	now := time.Now()

	exam.RemoveChapter(ch1, now)

	if len(exam.ChapterIDs) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(exam.ChapterIDs))
	}
	if exam.ChapterIDs[0] != ch2 {
		t.Errorf("remaining chapter = %v, want %v", exam.ChapterIDs[0], ch2)
	}
	if !exam.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt should be updated")
	}
}

func TestExam_RemoveChapter_NotFound(t *testing.T) {
	ch := uuid.Must(uuid.NewV7())
	oldTime := time.Now().Add(-time.Hour)
	exam := &Exam{
		ID:         uuid.Must(uuid.NewV7()),
		ChapterIDs: []uuid.UUID{ch},
		UpdatedAt:  oldTime,
	}

	exam.RemoveChapter(uuid.Must(uuid.NewV7()), time.Now())

	if len(exam.ChapterIDs) != 1 {
		t.Errorf("should not remove anything: got %d chapters", len(exam.ChapterIDs))
	}
	if !exam.UpdatedAt.Equal(oldTime) {
		t.Error("UpdatedAt should not change when chapter not found")
	}
}

func TestExam_DaysUntil(t *testing.T) {
	tests := []struct {
		name     string
		examDate time.Time
		from     time.Time
		want     int
	}{
		{"7 days", time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), 7},
		{"0 days same day", time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC), time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), 0},
		{"past exam", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exam := &Exam{ExamDate: tt.examDate}
			got := exam.DaysUntil(tt.from)
			if got != tt.want {
				t.Errorf("DaysUntil() = %d, want %d", got, tt.want)
			}
		})
	}
}
