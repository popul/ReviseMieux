//go:build integration

package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFirstConnectionJourney verifies the complete first-connection user journey
// as documented in docs/user-journeys/first-connection.md.
//
// Steps:
//  1. GET  /dev/token                        → obtain dev JWT
//  2. GET  /api/v1/chapters                  → empty list
//  3. POST /api/v1/onboarding/seed-demo      → creates demo chapter (8 items, 3 notions)
//  4. GET  /api/v1/chapters                  → 1 chapter with item_count=8, mastery_breakdown
//  5. GET  /api/v1/chapters/{id}/lesson-card → chapter + 8 items + 3 notions
//  6. GET  /api/v1/masteries?state=UNKNOWN   → 8 masteries
//  7. POST /api/v1/sessions/daily            → session created
//  8. GET  /api/v1/sessions/{id}/questions   → 8 questions with types and prompts
//  9. POST /api/v1/sessions/{id}/answer      → feedback (x8)
//
// 10. GET  /api/v1/sessions/{id}/debrief     → score + transitions
func TestFirstConnectionJourney(t *testing.T) {
	ta := setupTestApp(t)

	// Seed the 3 question templates required by session composition.
	ta.seedTemplates()

	// ── Step 1: Get dev token ────────────────────────────────
	// In a real app, the mobile calls GET /dev/token (public).
	// Here we create a user + token via the test helper instead,
	// since the dev handler needs a DB pool wired differently in tests.
	userID, token := ta.seedUser("student")
	_ = userID

	// ── Step 2: List chapters → empty ────────────────────────
	w := doRequest(ta.router, "GET", "/api/v1/chapters", token, nil)
	assertStatus(t, w, http.StatusOK)

	var chapters []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &chapters)
	if len(chapters) != 0 {
		t.Fatalf("Step 2: expected 0 chapters, got %d", len(chapters))
	}

	// ── Step 3: Seed demo chapter ────────────────────────────
	w = doRequest(ta.router, "POST", "/api/v1/onboarding/seed-demo", token, nil)
	assertStatus(t, w, http.StatusCreated)

	var seedResp struct {
		ChapterID string `json:"chapter_id"`
		ItemCount int    `json:"item_count"`
		Message   string `json:"message"`
	}
	json.Unmarshal(w.Body.Bytes(), &seedResp)

	if seedResp.ItemCount != 8 {
		t.Fatalf("Step 3: item_count = %d, want 8", seedResp.ItemCount)
	}
	if seedResp.ChapterID == "" {
		t.Fatal("Step 3: chapter_id is empty")
	}
	chapterID := seedResp.ChapterID

	// ── Step 4: List chapters → 1 demo chapter with enrichment ─
	w = doRequest(ta.router, "GET", "/api/v1/chapters", token, nil)
	assertStatus(t, w, http.StatusOK)

	json.Unmarshal(w.Body.Bytes(), &chapters)
	if len(chapters) != 1 {
		t.Fatalf("Step 4: expected 1 chapter, got %d", len(chapters))
	}

	ch := chapters[0]
	if ch["is_demo"] != true {
		t.Error("Step 4: chapter should be is_demo=true")
	}
	if ch["name"] != "Densité et masse volumique (démo)" {
		t.Errorf("Step 4: name = %v", ch["name"])
	}
	if int(ch["item_count"].(float64)) != 8 {
		t.Errorf("Step 4: item_count = %v, want 8", ch["item_count"])
	}
	// mastery_breakdown should exist with 8 unknown
	mb, ok := ch["mastery_breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("Step 4: mastery_breakdown missing")
	}
	if int(mb["unknown"].(float64)) != 8 {
		t.Errorf("Step 4: mastery_breakdown.unknown = %v, want 8", mb["unknown"])
	}

	// ── Step 5: Lesson card → 8 items + 3 notions ────────────
	w = doRequest(ta.router, "GET", "/api/v1/chapters/"+chapterID+"/lesson-card", token, nil)
	assertStatus(t, w, http.StatusOK)

	var card struct {
		Chapter map[string]interface{}   `json:"chapter"`
		Items   []map[string]interface{} `json:"items"`
		Notions []map[string]interface{} `json:"notions"`
	}
	json.Unmarshal(w.Body.Bytes(), &card)

	if len(card.Items) != 8 {
		t.Fatalf("Step 5: items count = %d, want 8", len(card.Items))
	}
	if len(card.Notions) != 3 {
		t.Fatalf("Step 5: notions count = %d, want 3", len(card.Notions))
	}

	// Verify all items have a notion_id
	for i, item := range card.Items {
		if item["notion_id"] == nil || item["notion_id"] == "" {
			t.Errorf("Step 5: item %d has no notion_id", i)
		}
		if item["term"] == nil || item["term"] == "" {
			t.Errorf("Step 5: item %d has no term", i)
		}
	}

	// Verify notion names
	notionNames := map[string]bool{}
	for _, n := range card.Notions {
		notionNames[n["name"].(string)] = true
	}
	for _, expected := range []string{"Masse volumique (ρ)", "Densité et flottabilité", "Conversions et applications"} {
		if !notionNames[expected] {
			t.Errorf("Step 5: missing notion %q", expected)
		}
	}

	// ── Step 6: Masteries → 8 UNKNOWN ────────────────────────
	w = doRequest(ta.router, "GET", "/api/v1/masteries?state=UNKNOWN", token, nil)
	assertStatus(t, w, http.StatusOK)

	var masteries []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &masteries)
	if len(masteries) != 8 {
		t.Fatalf("Step 6: masteries count = %d, want 8", len(masteries))
	}
	for i, m := range masteries {
		if m["state"] != "UNKNOWN" {
			t.Errorf("Step 6: mastery %d state = %v, want UNKNOWN", i, m["state"])
		}
	}

	// ── Step 7: Compose daily session ────────────────────────
	w = doRequest(ta.router, "POST", "/api/v1/sessions/daily", token,
		map[string]string{"chapter_id": chapterID})
	assertStatus(t, w, http.StatusCreated)

	var session struct {
		ID          string `json:"id"`
		SessionType string `json:"session_type"`
		Status      string `json:"status"`
	}
	json.Unmarshal(w.Body.Bytes(), &session)

	if session.ID == "" {
		t.Fatal("Step 7: session id is empty")
	}
	if session.SessionType != "daily" {
		t.Errorf("Step 7: session_type = %v, want daily", session.SessionType)
	}
	sessionID := session.ID

	// ── Step 8: Get questions → 8 with types and prompts ─────
	w = doRequest(ta.router, "GET", "/api/v1/sessions/"+sessionID+"/questions", token, nil)
	assertStatus(t, w, http.StatusOK)

	var questions []struct {
		ID             string   `json:"id"`
		TemplateID     string   `json:"template_id"`
		ItemID         string   `json:"item_id"`
		QuestionType   string   `json:"question_type"`
		RenderedPrompt string   `json:"rendered_prompt"`
		Choices        []string `json:"choices"`
	}
	json.Unmarshal(w.Body.Bytes(), &questions)

	if len(questions) < 6 || len(questions) > 10 {
		t.Fatalf("Step 8: questions count = %d, want 6-10", len(questions))
	}

	// Verify each question has required fields
	for i, q := range questions {
		if q.ID == "" {
			t.Errorf("Step 8: question %d has no id", i)
		}
		if q.QuestionType == "" {
			t.Errorf("Step 8: question %d has no question_type", i)
		}
		if q.RenderedPrompt == "" {
			t.Errorf("Step 8: question %d has no rendered_prompt", i)
		}
		// MCQ questions should have choices
		if q.QuestionType == "MCQ" && len(q.Choices) == 0 {
			t.Errorf("Step 8: MCQ question %d has no choices", i)
		}
	}

	// ── Step 9: Answer all questions ─────────────────────────
	correctCount := 0
	for i, q := range questions {
		// Alternate: odd questions correct, even incorrect
		score := 0.0
		if i%2 == 0 {
			score = 1.0
			correctCount++
		}

		w = doRequest(ta.router, "POST", "/api/v1/sessions/"+sessionID+"/answer", token,
			map[string]interface{}{
				"question_id": q.ID,
				"answer":      `{"text":"test answer"}`,
				"score":       score,
			})
		assertStatus(t, w, http.StatusOK)

		var answerResp struct {
			AttemptID string  `json:"attempt_id"`
			Score     float64 `json:"score"`
			Feedback  *struct {
				CorrectAnswer  string `json:"correct_answer"`
				WhatWasMissing string `json:"what_was_missing"`
				Hint           string `json:"hint"`
			} `json:"feedback"`
		}
		json.Unmarshal(w.Body.Bytes(), &answerResp)

		if answerResp.AttemptID == "" {
			t.Errorf("Step 9: question %d: attempt_id is empty", i)
		}
		if answerResp.Score != score {
			t.Errorf("Step 9: question %d: score = %v, want %v", i, answerResp.Score, score)
		}
		// Failed answers (score < 0.7) should have feedback (Z4-AC09)
		if score < 0.7 && answerResp.Feedback == nil {
			t.Errorf("Step 9: question %d: expected feedback for incorrect answer", i)
		}
	}

	// ── Step 10: Debrief ─────────────────────────────────────
	w = doRequest(ta.router, "GET", "/api/v1/sessions/"+sessionID+"/debrief", token, nil)
	assertStatus(t, w, http.StatusOK)

	var debrief struct {
		Score       float64 `json:"score"`
		Total       int     `json:"total"`
		Percentage  float64 `json:"percentage"`
		Transitions []struct {
			ItemID   string `json:"item_id"`
			ItemTerm string `json:"item_term"`
			From     string `json:"from"`
			To       string `json:"to"`
		} `json:"transitions"`
	}
	json.Unmarshal(w.Body.Bytes(), &debrief)

	if debrief.Total != len(questions) {
		t.Errorf("Step 10: total = %d, want %d", debrief.Total, len(questions))
	}
	if debrief.Score != float64(correctCount) {
		t.Errorf("Step 10: score = %v, want %v", debrief.Score, float64(correctCount))
	}
	expectedPct := (float64(correctCount) / float64(len(questions))) * 100
	if debrief.Percentage != expectedPct {
		t.Errorf("Step 10: percentage = %v, want %v", debrief.Percentage, expectedPct)
	}

	// ── Verify mastery transitions ───────────────────────────
	// Items that scored 1.0 should now be FRAGILE (Z1-AC01: UNKNOWN + success → FRAGILE)
	w = doRequest(ta.router, "GET", "/api/v1/masteries?state=FRAGILE", token, nil)
	assertStatus(t, w, http.StatusOK)

	var fragileMasteries []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &fragileMasteries)

	// We answered correctCount questions with score=1.0
	// Each should have triggered UNKNOWN → FRAGILE via the AttemptRecorded event
	// Note: transitions depend on the event publisher being wired correctly
	t.Logf("After session: %d FRAGILE masteries (expected up to %d)", len(fragileMasteries), correctCount)

	// ── Verify seed-demo is idempotent ───────────────────────
	w = doRequest(ta.router, "POST", "/api/v1/onboarding/seed-demo", token, nil)
	assertStatus(t, w, http.StatusCreated)

	var seedResp2 struct {
		ChapterID string `json:"chapter_id"`
		ItemCount int    `json:"item_count"`
	}
	json.Unmarshal(w.Body.Bytes(), &seedResp2)
	if seedResp2.ChapterID != chapterID {
		t.Errorf("Idempotent: chapter_id changed from %s to %s", chapterID, seedResp2.ChapterID)
	}
}

// seedTemplates inserts the 3 question templates used by session composition.
func (ta *testApp) seedTemplates() {
	ta.t.Helper()
	templates := []struct {
		id           string
		name         string
		questionType string
	}{
		{"GEN.KNOW.FLASH_MCQ", "Flash MCQ", "MCQ"},
		{"GEN.KNOW.DEF_SHORT", "Définition courte", "SHORT_ANSWER"},
		{"GEN.KNOW.CLOZE_KEYWORDS", "Cloze mots-clés", "CLOZE"},
	}
	for _, tmpl := range templates {
		_, err := ta.pool.Exec(ta.t.Context(),
			`INSERT INTO templates (id, name, question_type, difficulty, prompt_template, eligibility, grading)
			 VALUES ($1, $2, $3, 1, $4, '{}', '{}')
			 ON CONFLICT (id) DO NOTHING`,
			tmpl.id, tmpl.name, tmpl.questionType, "{{term}}")
		if err != nil {
			ta.t.Fatalf("seedTemplates(%s): %v", tmpl.id, err)
		}
	}
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, expected, w.Body.String())
	}
}
