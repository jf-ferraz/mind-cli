package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// sampleConvergenceMarkdown is a realistic convergence document snippet
// using the snake_case dimension names that the scoreRe regex matches.
const sampleConvergenceMarkdown = `# Test Convergence Analysis

## Quality Scoring

| perspective_diversity | 4 | Four distinct personas with genuine philosophical conflict |
| evidence_quality | 4 | MCP spec citation and line-level code references |
| concession_depth | 3 | Two tracked concessions; positions genuinely revised |
| challenge_substantiveness | 4 | Specific protocol clause citations, not strawmen |
| synthesis_quality | 3 | Findings organized thematically with clear rationale |
| actionability | 4 | Each recommendation has concrete code location and fix |

**Overall Quality Score: 3.67 / 5.0**
`

// FR-144 / M-2 fix: parseConvergenceEntry must parse all 6 dimension names
// with non-zero values from a real convergence sample using snake_case names.
func TestParseConvergenceEntry_AllSixDimensions(t *testing.T) {
	entry, err := parseConvergenceEntry(sampleConvergenceMarkdown, "test-convergence.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	if len(entry.Dimensions) != 6 {
		t.Errorf("len(Dimensions) = %d, want 6", len(entry.Dimensions))
	}

	// All 6 dimensions must have Value > 0.
	for _, d := range entry.Dimensions {
		if d.Value == 0 {
			t.Errorf("dimension %q has Value=0, want > 0", d.Name)
		}
	}
}

// FR-144: parseConvergenceEntry extracts the correct dimension names.
func TestParseConvergenceEntry_DimensionNames(t *testing.T) {
	entry, err := parseConvergenceEntry(sampleConvergenceMarkdown, "test.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	wantNames := []string{
		domain.DimPerspectiveDiversity,
		domain.DimEvidenceQuality,
		domain.DimConcessionDepth,
		domain.DimChallengeSubstantiveness,
		domain.DimSynthesisQuality,
		domain.DimActionability,
	}

	foundNames := make(map[string]bool)
	for _, d := range entry.Dimensions {
		foundNames[d.Name] = true
	}

	for _, want := range wantNames {
		if !foundNames[want] {
			t.Errorf("dimension %q not found in parsed dimensions; got %v", want, foundNamesSlice(foundNames))
		}
	}
}

func foundNamesSlice(m map[string]bool) []string {
	var result []string
	for k := range m {
		result = append(result, k)
	}
	return result
}

// FR-144: parseConvergenceEntry extracts the overall score from "Overall Quality Score: N.N".
func TestParseConvergenceEntry_OverallScore(t *testing.T) {
	entry, err := parseConvergenceEntry(sampleConvergenceMarkdown, "test.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	// The sample has "Overall Quality Score: 3.67 / 5.0"
	if entry.Score <= 0 {
		t.Errorf("Score = %.2f, want > 0", entry.Score)
	}
	// Score should be close to 3.67 (overallRe matches "3.67").
	if entry.Score < 3.6 || entry.Score > 3.8 {
		t.Errorf("Score = %.2f, want ~3.67", entry.Score)
	}
}

// FR-144: parseConvergenceEntry sets GatePass=true when Score >= 3.0.
func TestParseConvergenceEntry_GatePass(t *testing.T) {
	entry, err := parseConvergenceEntry(sampleConvergenceMarkdown, "test.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	if !entry.GatePass {
		t.Errorf("GatePass = false for score %.2f, want true (score >= 3.0)", entry.Score)
	}
}

// FR-144: parseConvergenceEntry averages dimensions when no overall score line is present.
func TestParseConvergenceEntry_AveragedScoreWhenNoOverallLine(t *testing.T) {
	content := `| perspective_diversity | 4 | desc |
| evidence_quality | 4 | desc |
| concession_depth | 2 | desc |
| challenge_substantiveness | 4 | desc |
| synthesis_quality | 4 | desc |
| actionability | 4 | desc |
`
	entry, err := parseConvergenceEntry(content, "no-overall.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	// Average of [4,4,2,4,4,4] = 22/6 ≈ 3.67
	if entry.Score <= 0 {
		t.Errorf("Score = %.2f, want > 0 (computed from average)", entry.Score)
	}
}

// FR-144: parseConvergenceEntry fills missing dimensions with Value=0 to always produce 6.
func TestParseConvergenceEntry_MissingDimensionsFilled(t *testing.T) {
	// Only 3 dimensions present.
	content := `| perspective_diversity | 4 | desc |
| evidence_quality | 3 | desc |
| actionability | 5 | desc |
**Overall Quality Score: 4.0 / 5.0**
`
	entry, err := parseConvergenceEntry(content, "partial.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	if len(entry.Dimensions) != 6 {
		t.Errorf("len(Dimensions) = %d, want 6 (missing ones filled with 0)", len(entry.Dimensions))
	}
}

// FR-144: QualityService.Log() writes to quality-log.yml and returns a valid entry.
func TestQualityService_Log_WritesEntry(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write a real convergence file under projectRoot.
	convergenceRelPath := "docs/knowledge/test-convergence.md"
	convergenceAbsPath := filepath.Join(root, convergenceRelPath)
	if err := os.WriteFile(convergenceAbsPath, []byte(sampleConvergenceMarkdown), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	qualityRepo := mem.NewQualityRepo()
	svc := NewQualityService(root, qualityRepo)

	entry, err := svc.Log(convergenceRelPath, "phase-3-review", "convergence-v1")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	if entry == nil {
		t.Fatal("Log() returned nil entry")
	}
	if entry.Topic != "phase-3-review" {
		t.Errorf("Topic = %q, want 'phase-3-review'", entry.Topic)
	}
	if entry.Variant != "convergence-v1" {
		t.Errorf("Variant = %q, want 'convergence-v1'", entry.Variant)
	}
	if len(entry.Dimensions) != 6 {
		t.Errorf("len(Dimensions) = %d, want 6", len(entry.Dimensions))
	}
	if entry.Score <= 0 {
		t.Errorf("Score = %.2f, want > 0", entry.Score)
	}

	// All dimensions must have Value > 0 when the source doc uses rubric names.
	for _, d := range entry.Dimensions {
		if d.Value == 0 {
			t.Errorf("dimension %q has Value=0, want > 0 (M-2 regression check)", d.Name)
		}
	}
}

// FR-144: QualityService.Log() creates quality-log.yml file on disk.
func TestQualityService_Log_CreatesQualityLogFile(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	convergenceRelPath := "docs/knowledge/test-convergence.md"
	convergenceAbsPath := filepath.Join(root, convergenceRelPath)
	if err := os.WriteFile(convergenceAbsPath, []byte(sampleConvergenceMarkdown), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	qualityRepo := mem.NewQualityRepo()
	svc := NewQualityService(root, qualityRepo)

	_, err := svc.Log(convergenceRelPath, "test-topic", "")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	logPath := filepath.Join(root, "docs", "knowledge", "quality-log.yml")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("quality-log.yml not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("quality-log.yml is empty")
	}
	content := string(data)
	if !strings.Contains(content, "test-topic") {
		t.Errorf("quality-log.yml does not contain topic 'test-topic':\n%s", content)
	}
}

// FR-144: QualityService.Log() uses filename as topic when topic is empty.
func TestQualityService_Log_DefaultTopicFromFilename(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	convergenceRelPath := "docs/knowledge/my-analysis.md"
	convergenceAbsPath := filepath.Join(root, convergenceRelPath)
	if err := os.WriteFile(convergenceAbsPath, []byte(sampleConvergenceMarkdown), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	qualityRepo := mem.NewQualityRepo()
	svc := NewQualityService(root, qualityRepo)

	entry, err := svc.Log(convergenceRelPath, "", "")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	if entry.Topic != "my-analysis" {
		t.Errorf("Topic = %q, want 'my-analysis' (derived from filename)", entry.Topic)
	}
}

// FR-144: QualityService.Log() returns error for non-existent file.
func TestQualityService_Log_FileNotFound(t *testing.T) {
	root := t.TempDir()
	qualityRepo := mem.NewQualityRepo()
	svc := NewQualityService(root, qualityRepo)

	_, err := svc.Log("docs/knowledge/nonexistent.md", "topic", "")
	if err == nil {
		t.Error("Log() should return error for non-existent file")
	}
}

// FR-144: QualityEntry.Validate() passes for a valid entry with Score=3.5.
func TestQualityEntry_Validate_ValidScore(t *testing.T) {
	entry := domain.QualityEntry{
		Topic:    "test",
		Score:    3.5,
		GatePass: true,
		Date:     time.Now(),
		Dimensions: []domain.QualityDimension{
			{Name: domain.DimPerspectiveDiversity, Value: 4},
			{Name: domain.DimEvidenceQuality, Value: 3},
			{Name: domain.DimConcessionDepth, Value: 4},
			{Name: domain.DimChallengeSubstantiveness, Value: 3},
			{Name: domain.DimSynthesisQuality, Value: 3},
			{Name: domain.DimActionability, Value: 4},
		},
	}
	if err := entry.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil for valid entry", err)
	}
}

// FR-144: QualityEntry.Validate() fails when only 5 dimensions are present.
func TestQualityEntry_Validate_WrongDimensionCount(t *testing.T) {
	entry := domain.QualityEntry{
		Topic:    "test",
		Score:    4.0,
		GatePass: true,
		Date:     time.Now(),
		Dimensions: []domain.QualityDimension{
			{Name: domain.DimPerspectiveDiversity, Value: 4},
			{Name: domain.DimEvidenceQuality, Value: 4},
			{Name: domain.DimConcessionDepth, Value: 4},
			{Name: domain.DimChallengeSubstantiveness, Value: 4},
			{Name: domain.DimSynthesisQuality, Value: 4},
			// Missing DimActionability — only 5 dimensions.
		},
	}
	if err := entry.Validate(); err == nil {
		t.Error("Validate() should return error for 5 dimensions (want 6)")
	}
}

// FR-144: QualityService.ReadLog delegates to qualityRepo.ReadLog.
func TestQualityService_ReadLog_DelegatesToRepo(t *testing.T) {
	qualityRepo := mem.NewQualityRepo()
	qualityRepo.Entries = []domain.QualityEntry{
		{Topic: "a", Score: 4.0, GatePass: true},
		{Topic: "b", Score: 2.0, GatePass: false},
	}
	svc := NewQualityService(t.TempDir(), qualityRepo)

	entries, err := svc.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("ReadLog() = %d entries, want 2", len(entries))
	}
}

// FR-144: All 6 constant names are present in QualityService dimension map.
func TestQualityService_DimensionConstants_AllPresent(t *testing.T) {
	wantDims := []string{
		domain.DimPerspectiveDiversity,
		domain.DimEvidenceQuality,
		domain.DimConcessionDepth,
		domain.DimChallengeSubstantiveness,
		domain.DimSynthesisQuality,
		domain.DimActionability,
	}

	// Parse a document that uses all 6 dimension names.
	content := ""
	for _, d := range wantDims {
		content += "| " + d + " | 3 | description |\n"
	}
	content += "**Overall Quality Score: 3.0 / 5.0**\n"

	entry, err := parseConvergenceEntry(content, "all-dims.md")
	if err != nil {
		t.Fatalf("parseConvergenceEntry() error = %v", err)
	}

	parsedDims := make(map[string]int)
	for _, d := range entry.Dimensions {
		parsedDims[d.Name] = d.Value
	}

	for _, want := range wantDims {
		val, ok := parsedDims[want]
		if !ok {
			t.Errorf("dimension %q not found in parsed entry", want)
			continue
		}
		if val == 0 {
			t.Errorf("dimension %q has Value=0, want 3", want)
		}
	}
}
