package benchmark

import (
	"context"
	"time"
)

// ProviderInfo holds metadata for any benchmark provider (used by Evaluate for cost computation).
type ProviderInfo interface {
	Name() string
	ModelID() string
	PricePerMInput() float64
	PricePerMOutput() float64
}

// Provider abstracts an LLM API for IDP benchmarking (text → structured items).
type Provider interface {
	ProviderInfo
	// StructureBlocks sends the structuration prompt and returns the raw response.
	StructureBlocks(ctx context.Context, systemPrompt, userPrompt string) (*Response, error)
}

// Response holds the raw LLM response with usage metadata.
type Response struct {
	RawJSON      []byte
	TokensInput  int
	TokensOutput int
	LatencyMs    int64
	ModelVersion string
}

// TestCase represents a single benchmark test case loaded from disk.
type TestCase struct {
	ID         string       `json:"id"`
	Subject    string       `json:"subject"`
	Level      string       `json:"level"`
	Topic      string       `json:"topic"`
	Difficulty string       `json:"difficulty"`
	Blocks     []OCRBlock   `json:"blocks"`
	Golden     GoldenOutput `json:"golden"`
	HasImages  bool         `json:"has_images"`
	ImagePaths []string     `json:"-"` // populated at load time for OCR benchmark
}

// OCRBlock mirrors chapter.OCRBlock for benchmark inputs.
type OCRBlock struct {
	Text       string  `json:"text" yaml:"text"`
	BlockType  string  `json:"block_type" yaml:"block_type"`
	Confidence float32 `json:"confidence" yaml:"confidence"`
}

// GoldenOutput is the human-validated expected output for a test case.
type GoldenOutput struct {
	Items   []GoldenItem `json:"items" yaml:"items"`
	Notions []string     `json:"notions" yaml:"notions"`
}

// GoldenItem is a single expected item in the golden output.
type GoldenItem struct {
	Type     string   `json:"type" yaml:"type"`
	Term     string   `json:"term" yaml:"term"`
	Keywords []string `json:"keywords" yaml:"keywords"`
	Steps    []string `json:"steps,omitempty" yaml:"steps,omitempty"`
	Notion   string   `json:"notion" yaml:"notion"`
}

// ParsedOutput is the LLM output parsed into structured form.
type ParsedOutput struct {
	Items   []ParsedItem `json:"items"`
	Notions []string     `json:"notions"`
}

// ParsedItem is a single item parsed from the LLM response.
type ParsedItem struct {
	Type       string   `json:"type"`
	Term       string   `json:"term"`
	Keywords   []string `json:"keywords"`
	Steps      []string `json:"steps,omitempty"`
	NotionName string   `json:"notion_name"`
	Confidence float32  `json:"confidence"`
}

// EvalResult holds all evaluation metrics for one (provider, test case) pair.
type EvalResult struct {
	CaseID    string    `json:"case_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`

	// Quality indicators (0-1, higher is better except HallucinationRate)
	CompletenessScore   float64 `json:"completeness_score"`   // Q1
	ClassificationScore float64 `json:"classification_score"` // Q2
	FidelityScore       float64 `json:"fidelity_score"`       // Q3
	KeywordScore        float64 `json:"keyword_score"`        // Q4
	HallucinationRate   float64 `json:"hallucination_rate"`   // Q5 (lower is better)
	NotionScore         float64 `json:"notion_score"`         // Q6
	SchemaCompliance    bool    `json:"schema_compliance"`    // Q7

	// Performance
	LatencyMs    int64   `json:"latency_ms"`
	TokensInput  int     `json:"tokens_input"`
	TokensOutput int     `json:"tokens_output"`
	CostUSD      float64 `json:"cost_usd"`
	CostPerItem  float64 `json:"cost_per_item"`
	ItemsFound   int     `json:"items_found"`

	// Composite
	QualityScore   float64 `json:"quality_score"`
	CompositeScore float64 `json:"composite_score"`

	// Error if the call failed
	Error string `json:"error,omitempty"`
}

// OCRProvider abstracts an OCR/vision API for benchmarking.
type OCRProvider interface {
	// Name returns the provider display name (e.g., "Anthropic").
	Name() string
	// ModelID returns the model identifier (e.g., "claude-sonnet-4-6").
	ModelID() string
	// PricePerMInput returns the cost in USD per 1M input tokens.
	PricePerMInput() float64
	// PricePerMOutput returns the cost in USD per 1M output tokens.
	PricePerMOutput() float64
	// ExtractBlocks sends images to the vision API and returns OCR blocks.
	ExtractBlocks(ctx context.Context, imagePaths []string, subject string) (*Response, error)
}

// E2EProvider abstracts a vision LLM API for end-to-end benchmarking (images → structured items).
type E2EProvider interface {
	Name() string
	ModelID() string
	PricePerMInput() float64
	PricePerMOutput() float64
	// StructureImages sends images directly to the VLM with the structuration prompt
	// and returns the structured items JSON.
	StructureImages(ctx context.Context, imagePaths []string, subject string) (*Response, error)
}

// HybridProvider abstracts a vision LLM API for hybrid benchmarking (images + OCR blocks → structured items).
type HybridProvider interface {
	ProviderInfo
	// StructureHybrid sends images + OCR text blocks to the VLM and returns structured items JSON.
	StructureHybrid(ctx context.Context, imagePaths []string, blocksJSON string, subject string) (*Response, error)
}

// OCREvalResult holds all evaluation metrics for one (provider, test case) OCR benchmark pair.
type OCREvalResult struct {
	CaseID    string    `json:"case_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`

	// Detection metrics
	BlocksExpected int     `json:"blocks_expected"`
	BlocksFound    int     `json:"blocks_found"`
	DetectionScore float64 `json:"detection_score"` // min(found,expected)/max(found,expected)

	// Text accuracy (average word overlap across matched blocks)
	TextAccuracy float64 `json:"text_accuracy"`

	// Block type classification accuracy
	TypeAccuracy float64 `json:"type_accuracy"`

	// Performance
	LatencyMs    int64   `json:"latency_ms"`
	TokensInput  int     `json:"tokens_input"`
	TokensOutput int     `json:"tokens_output"`
	CostUSD      float64 `json:"cost_usd"`

	// Composite
	QualityScore   float64 `json:"quality_score"`
	CompositeScore float64 `json:"composite_score"`

	// Error if the call failed
	Error string `json:"error,omitempty"`
}

// OCRRunSummary aggregates OCR results across all cases for one provider.
type OCRRunSummary struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`

	AvgDetectionScore float64 `json:"avg_detection_score"`
	AvgTextAccuracy   float64 `json:"avg_text_accuracy"`
	AvgTypeAccuracy   float64 `json:"avg_type_accuracy"`
	AvgLatencyMs      int64   `json:"avg_latency_ms"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	ErrorRate         float64 `json:"error_rate"`

	AvgQualityScore   float64 `json:"avg_quality_score"`
	AvgCompositeScore float64 `json:"avg_composite_score"`

	Results []OCREvalResult `json:"results"`
}

// RunSummary aggregates results across all cases for one provider.
type RunSummary struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`

	AvgCompletenessScore   float64 `json:"avg_completeness"`
	AvgClassificationScore float64 `json:"avg_classification"`
	AvgFidelityScore       float64 `json:"avg_fidelity"`
	AvgKeywordScore        float64 `json:"avg_keyword"`
	AvgHallucinationRate   float64 `json:"avg_hallucination_rate"`
	AvgLatencyMs           int64   `json:"avg_latency_ms"`
	TotalCostUSD           float64 `json:"total_cost_usd"`
	AvgCostPerItem         float64 `json:"avg_cost_per_item"`
	ErrorRate              float64 `json:"error_rate"`

	AvgQualityScore   float64 `json:"avg_quality_score"`
	AvgCompositeScore float64 `json:"avg_composite_score"`

	Results []EvalResult `json:"results"`
}
