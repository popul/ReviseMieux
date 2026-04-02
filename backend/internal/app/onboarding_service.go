package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// OnboardingStep represents the current step in the deterministic onboarding sequence (Z8-AC08).
type OnboardingStep string

const (
	StepAccountCreated   OnboardingStep = "account_created"    // (1)
	StepDemoAvailable    OnboardingStep = "demo_available"     // (2) demo chapter + empty state
	StepDemoSession      OnboardingStep = "demo_session"       // (3) optional demo evening_first
	StepFirstUpload      OnboardingStep = "first_upload"       // (4+5) tutorial → upload → pipeline
	StepFirstSession     OnboardingStep = "first_session"      // (6) evening_first on real chapter
	StepDebrief          OnboardingStep = "debrief"            // (7) debrief + parent invite
	StepOnboardingDone   OnboardingStep = "done"               // onboarding complete
)

// OnboardingStatus represents the current onboarding step for a user.
type OnboardingStatus struct {
	AccountCreated     bool           `json:"account_created"`
	DemoSessionDone    bool           `json:"demo_session_done"`
	FirstChapterReady  bool           `json:"first_chapter_ready"`
	HasDemoChapter     bool           `json:"has_demo_chapter"`
	DemoChapterID      *uuid.UUID     `json:"demo_chapter_id,omitempty"`
	CurrentStep        OnboardingStep `json:"current_step"`
	CompletedSteps     []OnboardingStep `json:"completed_steps"`
	Step1Label         string         `json:"step1_label"`
	Step2Label         string         `json:"step2_label"`
	Step3Label         string         `json:"step3_label"`
}

// DemoNotion is a pre-defined notion for the demo chapter.
type DemoNotion struct {
	Name      string
	SortOrder int
}

// demoNotions contains 3 notions for the demo chapter.
var demoNotions = []DemoNotion{
	{Name: "Masse volumique (ρ)", SortOrder: 1},
	{Name: "Densité et flottabilité", SortOrder: 2},
	{Name: "Conversions et applications", SortOrder: 3},
}

// DemoItem is a pre-defined item for the demo chapter.
type DemoItem struct {
	Term       string
	ItemType   chapter.ItemType
	Keywords   []string
	NotionIdx  int // index into demoNotions
}

// demoItems contains the 8 pre-generated items for the demo chapter (Z8-AC01).
var demoItems = []DemoItem{
	{Term: "La masse volumique est le rapport de la masse d'un corps sur son volume.", ItemType: chapter.ItemKnowledge, Keywords: []string{"masse volumique", "rapport", "masse", "volume"}, NotionIdx: 0},
	{Term: "L'unité SI de la masse volumique est le kilogramme par mètre cube (kg/m³).", ItemType: chapter.ItemKnowledge, Keywords: []string{"unité SI", "kg/m³", "masse volumique"}, NotionIdx: 0},
	{Term: "ρ = m / V", ItemType: chapter.ItemProcedure, Keywords: []string{"formule", "masse volumique", "ρ", "m", "V"}, NotionIdx: 0},
	{Term: "La densité d'un corps est le rapport de sa masse volumique sur celle de l'eau.", ItemType: chapter.ItemKnowledge, Keywords: []string{"densité", "masse volumique", "eau"}, NotionIdx: 1},
	{Term: "Un corps flotte si sa densité est inférieure à 1.", ItemType: chapter.ItemKnowledge, Keywords: []string{"flotte", "densité", "inférieure", "1"}, NotionIdx: 1},
	{Term: "La masse volumique de l'eau est 1000 kg/m³ (ou 1 g/cm³).", ItemType: chapter.ItemKnowledge, Keywords: []string{"eau", "1000 kg/m³", "1 g/cm³"}, NotionIdx: 1},
	{Term: "Pour convertir g/cm³ en kg/m³, on multiplie par 1000.", ItemType: chapter.ItemProcedure, Keywords: []string{"convertir", "g/cm³", "kg/m³", "1000"}, NotionIdx: 2},
	{Term: "La masse volumique permet d'identifier un matériau inconnu.", ItemType: chapter.ItemKnowledge, Keywords: []string{"identifier", "matériau", "masse volumique"}, NotionIdx: 2},
}

// OnboardingService handles first-use experience (Z8).
type OnboardingService struct {
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
	clock       event.Clock
	idGen       event.IDGenerator
}

// NewOnboardingService creates a new OnboardingService.
func NewOnboardingService(
	chapterRepo chapter.Repository,
	masteryRepo mastery.Repository,
	clock event.Clock,
	idGen event.IDGenerator,
) *OnboardingService {
	return &OnboardingService{
		chapterRepo: chapterRepo,
		masteryRepo: masteryRepo,
		clock:       clock,
		idGen:       idGen,
	}
}

// SeedDemoChapter creates the demo chapter with 8 pre-generated items (Z8-AC01).
// Returns the chapter and created items. Idempotent: if a demo chapter already exists, returns it.
func (s *OnboardingService) SeedDemoChapter(ctx context.Context, userID uuid.UUID) (*chapter.Chapter, []*chapter.Item, error) {
	// Check if demo chapter already exists
	chapters, err := s.chapterRepo.FindByUser(ctx, userID, true)
	if err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: find chapters: %w", err)
	}
	for _, ch := range chapters {
		if ch.IsDemo {
			items, err := s.chapterRepo.FindItemsByChapter(ctx, ch.ID, false)
			if err != nil {
				return nil, nil, fmt.Errorf("onboarding_service: find demo items: %w", err)
			}
			return ch, items, nil
		}
	}

	now := s.clock.Now()

	// Create demo chapter
	ch := chapter.NewChapter(s.idGen, userID, "Physique-Chimie", "5e", "Densité et masse volumique (démo)", now)
	ch.IsDemo = true

	// Create a revision for the demo chapter
	revID := s.idGen.New()
	rev := &chapter.Revision{
		ID:             revID,
		ChapterID:      ch.ID,
		RevisionNumber: 1,
		Status:         chapter.RevisionReady,
		CreatedAt:      now,
	}
	// Save chapter first (without revision link to avoid FK violation)
	if err := s.chapterRepo.Save(ctx, ch); err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: save chapter: %w", err)
	}
	// Save revision, then link it to the chapter
	if err := s.chapterRepo.SaveRevision(ctx, rev); err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: save revision: %w", err)
	}
	ch.CurrentRevisionID = &revID
	if err := s.chapterRepo.Save(ctx, ch); err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: link revision: %w", err)
	}

	// Create notions
	notionIDs := make([]uuid.UUID, len(demoNotions))
	for i, dn := range demoNotions {
		notionID := s.idGen.New()
		notionIDs[i] = notionID
		notion := &chapter.Notion{
			ID:        notionID,
			ChapterID: ch.ID,
			Name:      dn.Name,
			SortOrder: dn.SortOrder,
			CreatedAt: now,
		}
		if err := s.chapterRepo.SaveNotion(ctx, notion); err != nil {
			return nil, nil, fmt.Errorf("onboarding_service: save notion: %w", err)
		}
	}

	// Create items and masteries
	var items []*chapter.Item
	var masteries []*mastery.Mastery
	for _, di := range demoItems {
		term := di.Term
		notionID := notionIDs[di.NotionIdx]
		item := &chapter.Item{
			ID:                 s.idGen.New(),
			ChapterID:          ch.ID,
			RevisionID:         revID,
			NotionID:           &notionID,
			ItemType:           di.ItemType,
			Term:               &term,
			Confidence:         1.0, // demo items are trusted
			ValidationRequired: false,
			Keywords:           di.Keywords,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		items = append(items, item)

		m := mastery.NewMastery(s.idGen, userID, item.ID, now)
		masteries = append(masteries, m)
	}

	if err := s.chapterRepo.SaveItems(ctx, items); err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: save items: %w", err)
	}
	if err := s.masteryRepo.SaveAll(ctx, masteries); err != nil {
		return nil, nil, fmt.Errorf("onboarding_service: save masteries: %w", err)
	}

	return ch, items, nil
}

// ArchiveDemoIfNeeded archives the demo chapter when the first real chapter is ready (Z8-AC01).
func (s *OnboardingService) ArchiveDemoIfNeeded(ctx context.Context, userID uuid.UUID) error {
	chapters, err := s.chapterRepo.FindByUser(ctx, userID, true)
	if err != nil {
		return fmt.Errorf("onboarding_service: find chapters: %w", err)
	}

	var demoChapter *chapter.Chapter
	hasRealChapter := false
	for _, ch := range chapters {
		if ch.IsDemo && !ch.Archived {
			demoChapter = ch
		}
		if !ch.IsDemo && !ch.Archived {
			hasRealChapter = true
		}
	}

	if demoChapter != nil && hasRealChapter {
		now := s.clock.Now()
		demoChapter.Archived = true
		demoChapter.UpdatedAt = now
		if err := s.chapterRepo.Save(ctx, demoChapter); err != nil {
			return fmt.Errorf("onboarding_service: archive demo: %w", err)
		}
	}

	return nil
}

// GetOnboardingStatus returns the current onboarding state for a user (Z8-AC02, Z8-AC08).
// Z8-AC08: computes the deterministic onboarding step sequence.
func (s *OnboardingService) GetOnboardingStatus(ctx context.Context, userID uuid.UUID) (*OnboardingStatus, error) {
	chapters, err := s.chapterRepo.FindByUser(ctx, userID, true)
	if err != nil {
		return nil, fmt.Errorf("onboarding_service: find chapters: %w", err)
	}

	status := &OnboardingStatus{
		AccountCreated: true,
		Step1Label:     "Compte créé",
		Step2Label:     "Photographie ton premier cours",
		Step3Label:     "Ta première session de révision",
	}

	var hasRealChapterWithItems bool
	for _, ch := range chapters {
		if ch.IsDemo && !ch.Archived {
			status.HasDemoChapter = true
			id := ch.ID
			status.DemoChapterID = &id
		}
		if !ch.IsDemo && !ch.Archived {
			status.FirstChapterReady = true
			// Check if real chapter has items (pipeline completed)
			if ch.CurrentRevisionID != nil {
				items, err := s.chapterRepo.FindItemsByChapter(ctx, ch.ID, false)
				if err == nil && len(items) > 0 {
					hasRealChapterWithItems = true
				}
			}
		}
	}

	// Z8-AC08: Determine current step in the deterministic sequence
	status.CompletedSteps = []OnboardingStep{StepAccountCreated}
	status.CurrentStep = StepDemoAvailable

	if status.HasDemoChapter {
		status.CompletedSteps = append(status.CompletedSteps, StepDemoAvailable)
		status.CurrentStep = StepDemoSession
	}

	if status.DemoSessionDone {
		status.CompletedSteps = append(status.CompletedSteps, StepDemoSession)
		status.CurrentStep = StepFirstUpload
	}

	if status.FirstChapterReady && hasRealChapterWithItems {
		status.CompletedSteps = append(status.CompletedSteps, StepFirstUpload)
		status.CurrentStep = StepFirstSession
	}

	// If user already has sessions on real chapters, they're past first session
	if hasRealChapterWithItems {
		// Check for completed sessions on real chapters
		// For MVP, approximate: if items exist, user is at least at first_session
		// The mobile app will track actual session completion
	}

	return status, nil
}
