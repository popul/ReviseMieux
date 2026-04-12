package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
)

// renderQuestionFromItem generates a prompt and expected answer from an item and template.
func renderQuestionFromItem(item *chapter.Item, templateID string) (prompt string, expectedAnswer []byte) {
	term := ""
	if item.Term != nil {
		term = *item.Term
	}
	keywords := strings.Join(item.Keywords, ", ")

	isProcedure := item.ItemType == chapter.ItemProcedure

	switch templateID {
	case "GEN.KNOW.DEF_SHORT":
		if isProcedure {
			prompt = fmt.Sprintf("Quelle est la formule ou la méthode pour : %s ?", firstWords(term, 8))
		} else {
			prompt = fmt.Sprintf("Définis : %s", firstWords(term, 8))
		}
		expectedAnswer, _ = json.Marshal(map[string]string{"answer": term})
	case "GEN.KNOW.FLASH_MCQ":
		prompt = fmt.Sprintf("Vrai ou faux : %s", term)
		expectedAnswer, _ = json.Marshal(map[string]string{"answer": "Vrai"})
	case "GEN.KNOW.CLOZE_KEYWORDS":
		if isProcedure || len(keywords) < 3 {
			// Cloze doesn't work well with formulas or short items — use short answer instead
			prompt = fmt.Sprintf("Explique en une phrase : %s", firstWords(term, 8))
		} else {
			prompt = fmt.Sprintf("Complète : %s", maskedTerm(term))
		}
		expectedAnswer, _ = json.Marshal(map[string]string{"answer": term, "keywords": keywords})
	default:
		prompt = fmt.Sprintf("Qu'est-ce que : %s ?", firstWords(term, 8))
		expectedAnswer, _ = json.Marshal(map[string]string{"answer": term})
	}
	return
}

func firstWords(s string, n int) string {
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ") + "..."
}

func maskedTerm(s string) string {
	words := strings.Fields(s)
	for i := range words {
		if len(words[i]) > 4 && i%3 == 1 {
			words[i] = "____"
		}
	}
	return strings.Join(words, " ")
}

// scoreByKeywords scores a text answer by checking keyword overlap.
// Returns 1.0 if ≥70% of expected keywords found, 0.5 if ≥40%, 0.0 otherwise.
// stripAccents removes common French accents and Greek symbols for fuzzy matching.
func stripAccents(s string) string {
	replacer := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a", "ä", "a",
		"ù", "u", "û", "u", "ü", "u",
		"î", "i", "ï", "i",
		"ô", "o", "ö", "o",
		"ç", "c",
		"³", "3", "²", "2",
		"ρ", "rho", "Ω", "omega", "π", "pi", "Δ", "delta",
	)
	return replacer.Replace(s)
}

// stripPunctuation removes punctuation that interferes with keyword matching
// (apostrophes, periods, commas, etc.) and normalizes whitespace.
func stripPunctuation(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\'', '\u2019', '\u2018': // apostrophes → space
			b.WriteRune(' ')
		case '.', ',', ';', ':', '!', '?', '(', ')', '"':
			// drop punctuation
		default:
			b.WriteRune(r)
		}
	}
	// collapse multiple spaces
	return strings.Join(strings.Fields(b.String()), " ")
}

func scoreByKeywords(expectedAnswer, keywords, userAnswer string) float64 {
	normalize := func(s string) string {
		return stripPunctuation(stripAccents(strings.ToLower(strings.TrimSpace(s))))
	}
	answer := normalize(userAnswer)

	// Collect keywords from the keywords field (priority) — keep all, including 1-char
	var allKeywords []string
	if keywords != "" {
		for _, kw := range strings.Split(keywords, ",") {
			kw = normalize(kw)
			if kw != "" {
				allKeywords = append(allKeywords, kw)
			}
		}
	}
	// Also extract significant words from expected answer (>3 chars)
	for _, word := range strings.Fields(normalize(expectedAnswer)) {
		if len(word) > 3 {
			allKeywords = append(allKeywords, word)
		}
	}

	if len(allKeywords) == 0 {
		return 0.0
	}

	// Count matches
	matches := 0
	for _, kw := range allKeywords {
		if strings.Contains(answer, kw) {
			matches++
		}
	}

	ratio := float64(matches) / float64(len(allKeywords))
	if ratio >= 0.7 {
		return 1.0
	}
	if ratio >= 0.4 {
		return 0.5
	}
	return 0.0
}

// Difficulty-1 templates eligible for evening_first sessions.
var difficulty1Templates = []string{
	"GEN.KNOW.FLASH_MCQ",
	"GEN.KNOW.DEF_SHORT",
	"GEN.KNOW.CLOZE_KEYWORDS",
}

// SessionService handles session-related use cases.
type SessionService struct {
	sessionRepo session.Repository
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
	scorer      session.Scorer
}

// NewSessionService creates a new SessionService.
// The scorer parameter is optional (can be nil) — when nil, client-provided scores are used.
func NewSessionService(
	sessionRepo session.Repository,
	chapterRepo chapter.Repository,
	masteryRepo mastery.Repository,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
	scorer session.Scorer,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		chapterRepo: chapterRepo,
		masteryRepo: masteryRepo,
		publisher:   publisher,
		clock:       clock,
		idGen:       idGen,
		scorer:      scorer,
	}
}

// ProposeEveningFirst creates an evening_first session after pipeline completion.
// Z6-AC04: triggered immediately after pipeline produces items.
// Z6-AC05: 100% UNKNOWN items, difficulty 1 templates, 6-10 questions.
func (s *SessionService) ProposeEveningFirst(ctx context.Context, userID, chapterID uuid.UUID, itemIDs []uuid.UUID) (*session.Session, error) {
	now := s.clock.Now()

	if len(itemIDs) == 0 {
		return nil, session.ErrEmptyPool
	}

	sess := session.NewSession(s.idGen, userID, session.TypeEveningFirst, session.TriggerScheduled, now)
	sess.ChapterIDs = []uuid.UUID{chapterID}

	if err := s.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("session_service: save session: %w", err)
	}

	// Select items for the session: max 10 questions, min 6
	maxQuestions := 10
	if len(itemIDs) < maxQuestions {
		maxQuestions = len(itemIDs)
	}
	if maxQuestions < 6 {
		maxQuestions = len(itemIDs) // use all if fewer than 6
	}

	selectedItems := itemIDs[:maxQuestions]

	// Fetch items for rendering
	chapterItems, _ := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	itemMap := make(map[uuid.UUID]*chapter.Item)
	for _, it := range chapterItems {
		itemMap[it.ID] = it
	}

	// Generate questions using difficulty 1 templates (round-robin)
	for i, itemID := range selectedItems {
		templateID := difficulty1Templates[i%len(difficulty1Templates)]
		prompt, expectedAnswer := "", []byte(`{"answer":""}`)
		if item, ok := itemMap[itemID]; ok {
			prompt, expectedAnswer = renderQuestionFromItem(item, templateID)
		}
		q := &session.Question{
			ID:             s.idGen.New(),
			SessionID:      sess.ID,
			TemplateID:     templateID,
			ItemID:         itemID,
			RenderedPrompt: prompt,
			ExpectedAnswer: expectedAnswer,
			CreatedAt:      now,
		}
		if err := s.sessionRepo.SaveQuestion(ctx, q); err != nil {
			return nil, fmt.Errorf("session_service: save question: %w", err)
		}
	}

	return sess, nil
}

// ComposeDaily creates a daily session for a chapter.
// Z4-AC05: applies pack constraints (max_writing, must_include_doc).
// Z4-AC06: returns error if no items available (no empty session created).
func (s *SessionService) ComposeDaily(ctx context.Context, userID, chapterID uuid.UUID) (*session.Session, error) {
	now := s.clock.Now()

	// Get due items for this user
	dueMasteries, err := s.masteryRepo.FindDueByUser(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("session_service: find due: %w", err)
	}

	// Get chapter items with their types
	chapterItems, err := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	if err != nil {
		return nil, fmt.Errorf("session_service: find items: %w", err)
	}

	chapterItemMap := make(map[uuid.UUID]*chapter.Item)
	for _, item := range chapterItems {
		chapterItemMap[item.ID] = item
	}

	var eligibleMasteries []*mastery.Mastery
	for _, m := range dueMasteries {
		if _, ok := chapterItemMap[m.ItemID]; ok {
			eligibleMasteries = append(eligibleMasteries, m)
		}
	}

	if len(eligibleMasteries) == 0 {
		return nil, session.ErrEmptyPool
	}

	// Z4-AC05: Apply pack constraints
	pack := session.DefaultPackConstraints()
	selected := applyPackConstraints(eligibleMasteries, chapterItemMap, pack, 10)

	if len(selected) == 0 {
		return nil, session.ErrEmptyPool
	}

	sess := session.NewSession(s.idGen, userID, session.TypeDaily, session.TriggerManual, now)
	sess.ChapterIDs = []uuid.UUID{chapterID}

	if err := s.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("session_service: save session: %w", err)
	}

	for i, m := range selected {
		templateID := difficulty1Templates[i%len(difficulty1Templates)]
		item := chapterItemMap[m.ItemID]
		prompt, expectedAnswer := renderQuestionFromItem(item, templateID)
		q := &session.Question{
			ID:             s.idGen.New(),
			SessionID:      sess.ID,
			TemplateID:     templateID,
			ItemID:         m.ItemID,
			RenderedPrompt: prompt,
			ExpectedAnswer: expectedAnswer,
			CreatedAt:      now,
		}
		if err := s.sessionRepo.SaveQuestion(ctx, q); err != nil {
			return nil, fmt.Errorf("session_service: save question: %w", err)
		}
	}

	return sess, nil
}

// applyPackConstraints selects masteries respecting pack constraints (Z4-AC05).
// Returns at most maxItems masteries, ensuring:
// - At most pack.MaxWritingPerSession WRITING items
// - At least 1 DOCUMENT item if pack.SessionMustIncludeDoc and available
func applyPackConstraints(
	masteries []*mastery.Mastery,
	itemMap map[uuid.UUID]*chapter.Item,
	pack session.PackConstraints,
	maxItems int,
) []*mastery.Mastery {
	var docMasteries, writingMasteries, otherMasteries []*mastery.Mastery

	for _, m := range masteries {
		item, ok := itemMap[m.ItemID]
		if !ok {
			continue
		}
		switch item.ItemType {
		case chapter.ItemDocument:
			docMasteries = append(docMasteries, m)
		case chapter.ItemWriting:
			writingMasteries = append(writingMasteries, m)
		default:
			otherMasteries = append(otherMasteries, m)
		}
	}

	var selected []*mastery.Mastery

	// 1. Include exactly 1 document item if required and available
	if pack.SessionMustIncludeDoc && len(docMasteries) > 0 {
		selected = append(selected, docMasteries[0])
		docMasteries = docMasteries[1:]
	}

	// 2. Include writing items up to the cap
	writingCap := pack.MaxWritingPerSession
	for i := 0; i < len(writingMasteries) && i < writingCap && len(selected) < maxItems; i++ {
		selected = append(selected, writingMasteries[i])
	}

	// 3. Fill with other items (knowledge, procedure)
	for _, m := range otherMasteries {
		if len(selected) >= maxItems {
			break
		}
		selected = append(selected, m)
	}

	// 4. Fill remaining slots with extra doc items
	for _, m := range docMasteries {
		if len(selected) >= maxItems {
			break
		}
		selected = append(selected, m)
	}

	return selected
}

// ResumeSession allows resuming an in-progress session.
// Z4-AC04: resumes from last known question index.
func (s *SessionService) ResumeSession(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_service: find session: %w", err)
	}

	if err := sess.Resume(); err != nil {
		return nil, fmt.Errorf("session_service: %w", err)
	}

	return sess, nil
}

// GetSession retrieves a session by ID.
func (s *SessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	return s.sessionRepo.FindByID(ctx, sessionID)
}

// GetQuestions retrieves questions for a session.
func (s *SessionService) GetQuestions(ctx context.Context, sessionID uuid.UUID) ([]*session.Question, error) {
	return s.sessionRepo.FindQuestionsBySession(ctx, sessionID)
}

// SubmitAnswerResult is the result of submitting an answer.
type SubmitAnswerResult struct {
	AttemptID uuid.UUID
	Score     float64
	Feedback  *Feedback
}

// SubmitAnswer records an attempt and returns feedback (Z4-AC09).
func (s *SessionService) SubmitAnswer(ctx context.Context, sessionID, questionID, userID uuid.UUID, answerText string, score float64) (*SubmitAnswerResult, error) {
	now := s.clock.Now()

	// Find the question to get expected answer and template
	q, err := s.sessionRepo.FindQuestionByID(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("session_service: find question: %w", err)
	}

	// Count existing attempts to determine question rank in session
	questions, _ := s.sessionRepo.FindQuestionsBySession(ctx, sessionID)
	totalQ := len(questions)
	attempts, _ := s.sessionRepo.FindAttemptsBySession(ctx, sessionID)
	qNum := len(attempts) + 1

	// Auto-score text answers (non-MCQ questions)
	if !strings.Contains(q.TemplateID, "MCQ") {
		var expected struct {
			Answer   string `json:"answer"`
			Keywords string `json:"keywords"`
		}
		if json.Unmarshal(q.ExpectedAnswer, &expected) == nil && expected.Answer != "" {
			if s.scorer != nil {
				// LLM-based scoring
				result, err := s.scorer.ScoreAnswer(ctx, q.RenderedPrompt, expected.Answer, answerText)
				if err == nil {
					score = result.Score
					log.Printf("[scoring] q=%d/%d template=%s method=llm score=%.2f prompt=%q answer=%q",
						qNum, totalQ, q.TemplateID, score, q.RenderedPrompt, answerText)
				} else {
					log.Printf("[scoring] q=%d/%d template=%s method=llm error=%v, falling back to keywords",
						qNum, totalQ, q.TemplateID, err)
					score = scoreByKeywords(expected.Answer, expected.Keywords, answerText)
					log.Printf("[scoring] q=%d/%d template=%s method=keywords(fallback) score=%.2f prompt=%q answer=%q expected_keywords=%q",
						qNum, totalQ, q.TemplateID, score, q.RenderedPrompt, answerText, expected.Keywords)
				}
			} else {
				// Fallback: keyword matching when no LLM scorer
				score = scoreByKeywords(expected.Answer, expected.Keywords, answerText)
				log.Printf("[scoring] q=%d/%d template=%s method=keywords score=%.2f prompt=%q answer=%q",
					qNum, totalQ, q.TemplateID, score, q.RenderedPrompt, answerText)
			}
		}
	} else {
		log.Printf("[scoring] q=%d/%d template=%s method=auto(MCQ) score=%.2f",
			qNum, totalQ, q.TemplateID, score)
	}

	// Create and save the attempt
	attempt := session.NewAttempt(s.idGen, sessionID, questionID, userID, answerText, score, now)
	if err := s.sessionRepo.SaveAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("session_service: save attempt: %w", err)
	}

	// Generate feedback for incorrect answers
	var fb *Feedback
	if score < 0.7 && q.ExpectedAnswer != nil {
		answerBytes, _ := json.Marshal(answerText)
		f := GenerateFeedback(q.TemplateID, q.ExpectedAnswer, answerBytes)
		fb = &f
	}

	// Publish attempt event for mastery transition
	if err := s.publisher.Publish(ctx, event.AttemptRecorded{
		BaseEvent: event.BaseEvent{OccurredOn: now},
		UserID:    userID,
		ItemID:    q.ItemID,
		Score:     score,
	}); err != nil {
		return nil, fmt.Errorf("session_service: publish attempt: %w", err)
	}

	return &SubmitAnswerResult{
		AttemptID: attempt.ID,
		Score:     score,
		Feedback:  fb,
	}, nil
}

// ComposeMockExam creates a mock exam session, independent of any active daily session (Z4-AC07).
// CB1 has its own question pool and does not block or depend on daily sessions.
func (s *SessionService) ComposeMockExam(ctx context.Context, userID uuid.UUID, chapterIDs []uuid.UUID) (*session.Session, error) {
	now := s.clock.Now()

	if len(chapterIDs) == 0 {
		return nil, session.ErrEmptyPool
	}

	// Gather all items from all chapters
	var allItems []*chapter.Item
	for _, chID := range chapterIDs {
		items, err := s.chapterRepo.FindItemsByChapter(ctx, chID, false)
		if err != nil {
			return nil, fmt.Errorf("session_service: find items for chapter: %w", err)
		}
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		return nil, session.ErrEmptyPool
	}

	sess := session.NewSession(s.idGen, userID, session.TypeMockExam, session.TriggerManual, now)
	sess.ChapterIDs = chapterIDs

	if err := s.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("session_service: save mock exam: %w", err)
	}

	// Select up to 20 questions for a mock exam
	maxQuestions := 20
	if len(allItems) < maxQuestions {
		maxQuestions = len(allItems)
	}

	for i := 0; i < maxQuestions; i++ {
		templateID := difficulty1Templates[i%len(difficulty1Templates)]
		item := allItems[i]
		prompt, expectedAnswer := renderQuestionFromItem(item, templateID)
		q := &session.Question{
			ID:             s.idGen.New(),
			SessionID:      sess.ID,
			TemplateID:     templateID,
			ItemID:         item.ID,
			RenderedPrompt: prompt,
			ExpectedAnswer: expectedAnswer,
			CreatedAt:      now,
		}
		if err := s.sessionRepo.SaveQuestion(ctx, q); err != nil {
			return nil, fmt.Errorf("session_service: save mock exam question: %w", err)
		}
	}

	return sess, nil
}

// DebriefResult contains the computed debrief data for a completed session.
type DebriefResult struct {
	Score       float64
	Total       int
	Percentage  float64
	Transitions []TransitionInfo
}

// TransitionInfo describes a mastery state change for an item.
type TransitionInfo struct {
	ItemID   uuid.UUID
	ItemTerm string
	From     string
	To       string
}

// GetDebrief computes score statistics and real mastery transitions for a session.
func (s *SessionService) GetDebrief(ctx context.Context, sessionID uuid.UUID) (*DebriefResult, error) {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_service: get debrief: %w", err)
	}

	attempts, err := s.sessionRepo.FindAttemptsBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_service: find attempts: %w", err)
	}

	result := &DebriefResult{
		Transitions: []TransitionInfo{},
	}

	if len(attempts) == 0 {
		return result, nil
	}

	var totalScore float64
	for _, a := range attempts {
		totalScore += a.Score
	}

	result.Score = totalScore
	result.Total = len(attempts)
	result.Percentage = (totalScore / float64(len(attempts))) * 100

	// Compute real mastery transitions by looking at current mastery state
	// vs what it would have been before this session's attempts.
	questions, err := s.sessionRepo.FindQuestionsBySession(ctx, sessionID)
	if err != nil {
		//nolint:nilerr // degrade gracefully: return partial result rather than failing the whole summary
		return result, nil
	}

	// Build question lookup
	questionMap := make(map[uuid.UUID]*session.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Build item->term lookup from chapter items
	itemTerms := make(map[uuid.UUID]string)
	if len(sess.ChapterIDs) > 0 {
		for _, chID := range sess.ChapterIDs {
			items, err := s.chapterRepo.FindItemsByChapter(ctx, chID, false)
			if err != nil {
				continue
			}
			for _, item := range items {
				if item.Term != nil {
					itemTerms[item.ID] = *item.Term
				}
			}
		}
	}

	// For each attempt, compute what the transition was.
	// We infer from the score: success (>=0.7) or failure, and the current mastery state.
	seen := make(map[uuid.UUID]bool) // track items already processed
	for _, a := range attempts {
		q, ok := questionMap[a.QuestionID]
		if !ok {
			continue
		}
		if seen[q.ItemID] {
			continue // only report first transition per item
		}
		seen[q.ItemID] = true

		m, err := s.masteryRepo.FindByUserAndItem(ctx, a.UserID, q.ItemID)
		if err != nil {
			continue
		}

		// The current state IS the post-transition state (event handler already ran).
		// Infer the previous state from the transition rules.
		currentState := string(m.State)
		prevState := inferPreviousState(m, a.Score)

		if prevState != currentState {
			term := itemTerms[q.ItemID]
			if term == "" {
				term = q.ItemID.String()
			}
			result.Transitions = append(result.Transitions, TransitionInfo{
				ItemID:   q.ItemID,
				ItemTerm: term,
				From:     prevState,
				To:       currentState,
			})
		}
	}

	return result, nil
}

// inferPreviousState reverses the mastery state machine to determine
// what state the mastery was in before the last attempt.
func inferPreviousState(m *mastery.Mastery, score float64) string {
	success := score >= 0.7
	current := m.State

	switch current {
	case mastery.Fragile:
		if success {
			// Success led to FRAGILE → was UNKNOWN
			return string(mastery.Unknown)
		}
		// Failure kept FRAGILE, or regressed from OK
		if m.ConsecutiveFailures >= 1 {
			return string(mastery.OK)
		}
		return string(mastery.Fragile)
	case mastery.OK:
		if success {
			// Success led to OK → was FRAGILE
			return string(mastery.Fragile)
		}
		// Failure kept OK → was SOLID
		return string(mastery.Solid)
	case mastery.Solid:
		if success {
			// Success kept SOLID or led to SOLID → was OK
			if m.ConsecutiveSuccesses <= 2 {
				return string(mastery.OK)
			}
			return string(mastery.Solid)
		}
		return string(mastery.Solid) // shouldn't happen (SOLID+fail→OK)
	case mastery.Unknown:
		return string(mastery.Unknown) // failure on UNKNOWN stays UNKNOWN
	}
	return string(current)
}

// AvailableSessionTypes returns session types available based on schedule status.
// Z6-AC11: without schedule, daily/evening_first/mock_exam still available;
// pre_class requires schedule.
func (s *SessionService) AvailableSessionTypes(hasSchedule bool) []session.SessionType {
	types := []session.SessionType{
		session.TypeDaily,
		session.TypeEveningFirst,
		session.TypeMockExam,
		session.TypeDiagnostic,
	}
	if hasSchedule {
		types = append(types, session.TypePreClass)
	}
	return types
}
