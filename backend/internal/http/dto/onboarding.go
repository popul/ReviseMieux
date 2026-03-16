package dto

// OnboardingStatusResponse represents the onboarding state for a user.
type OnboardingStatusResponse struct {
	AccountCreated    bool    `json:"account_created"`
	DemoSessionDone   bool    `json:"demo_session_done"`
	FirstChapterReady bool    `json:"first_chapter_ready"`
	HasDemoChapter    bool    `json:"has_demo_chapter"`
	DemoChapterID     *string `json:"demo_chapter_id,omitempty"`
	Step1Label        string  `json:"step1_label"`
	Step2Label        string  `json:"step2_label"`
	Step3Label        string  `json:"step3_label"`
}

// SeedDemoResponse is returned after seeding the demo chapter.
type SeedDemoResponse struct {
	ChapterID string `json:"chapter_id"`
	ItemCount int    `json:"item_count"`
	Message   string `json:"message"`
}

// PipelineProgressResponse represents the current pipeline progress (Z8-AC04).
type PipelineProgressResponse struct {
	RevisionID     string `json:"revision_id"`
	Status         string `json:"status"`
	TotalPages     int    `json:"total_pages"`
	ProcessedPages int    `json:"processed_pages"`
	FailedPages    int    `json:"failed_pages"`
	TotalItems     int    `json:"total_items"`
	Phase          string `json:"phase"`
	PhaseMessage   string `json:"phase_message"`
}
