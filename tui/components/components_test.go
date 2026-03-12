package components

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// --- ZoneBar tests ---

func TestZoneBar_FullHealth(t *testing.T) {
	zh := domain.ZoneHealth{
		Zone:     domain.ZoneSpec,
		Total:    5,
		Complete: 5,
	}
	result := ZoneBar(zh, 80)

	if !strings.Contains(result, "spec/") {
		t.Errorf("expected zone label 'spec/', got: %s", result)
	}
	if !strings.Contains(result, "5/5") {
		t.Errorf("expected fraction '5/5', got: %s", result)
	}
}

func TestZoneBar_PartialHealth(t *testing.T) {
	zh := domain.ZoneHealth{
		Zone:     domain.ZoneBlueprints,
		Total:    3,
		Complete: 1,
	}
	result := ZoneBar(zh, 80)

	if !strings.Contains(result, "1/3") {
		t.Errorf("expected fraction '1/3', got: %s", result)
	}
}

func TestZoneBar_ZeroTotal(t *testing.T) {
	zh := domain.ZoneHealth{
		Zone:     domain.ZoneKnowledge,
		Total:    0,
		Complete: 0,
	}
	result := ZoneBar(zh, 80)

	if !strings.Contains(result, "0/0") {
		t.Errorf("expected fraction '0/0', got: %s", result)
	}
}

func TestZoneBar_WideTerminal(t *testing.T) {
	zh := domain.ZoneHealth{
		Zone:     domain.ZoneSpec,
		Total:    5,
		Complete: 5,
	}
	// Width >= 100 should use wider bar
	result := ZoneBar(zh, 100)
	// Just verify it renders without panicking and has content
	if result == "" {
		t.Error("expected non-empty output for wide terminal")
	}
}

func TestZoneBar_AllZones(t *testing.T) {
	for _, zone := range domain.AllZones {
		zh := domain.ZoneHealth{
			Zone:     zone,
			Total:    3,
			Complete: 2,
		}
		result := ZoneBar(zh, 80)
		if !strings.Contains(result, string(zone)+"/") {
			t.Errorf("zone %s: expected label '%s/', got: %s", zone, zone, result)
		}
	}
}

// --- Staleness tests ---

func TestStaleness_Empty(t *testing.T) {
	result := Staleness(nil, 80)
	if result != "" {
		t.Errorf("expected empty string for no stale docs, got: %q", result)
	}

	result = Staleness(map[string]string{}, 80)
	if result != "" {
		t.Errorf("expected empty string for empty stale map, got: %q", result)
	}
}

func TestStaleness_WithDocuments(t *testing.T) {
	stale := map[string]string{
		"doc:spec/architecture": "dependency changed",
		"doc:spec/requirements": "direct change",
	}
	result := Staleness(stale, 80)

	if !strings.Contains(result, "Staleness") {
		t.Error("expected 'Staleness' heading")
	}
	if !strings.Contains(result, "2 stale") {
		t.Error("expected '2 stale' count")
	}
	if !strings.Contains(result, "doc:spec/architecture") {
		t.Error("expected stale document ID in output")
	}
	if !strings.Contains(result, "doc:spec/requirements") {
		t.Error("expected stale document ID in output")
	}
}

func TestStaleness_Deterministic(t *testing.T) {
	stale := map[string]string{
		"doc:spec/z-last":  "reason1",
		"doc:spec/a-first": "reason2",
		"doc:spec/m-mid":   "reason3",
	}

	// Run multiple times to verify sort order is stable
	result1 := Staleness(stale, 80)
	result2 := Staleness(stale, 80)
	if result1 != result2 {
		t.Error("staleness output is not deterministic across calls")
	}

	// Verify sorted order
	aIdx := strings.Index(result1, "a-first")
	mIdx := strings.Index(result1, "m-mid")
	zIdx := strings.Index(result1, "z-last")
	if aIdx > mIdx || mIdx > zIdx {
		t.Error("stale documents should be sorted alphabetically")
	}
}

// --- WorkflowPanel tests ---

func TestWorkflowPanel_Idle(t *testing.T) {
	result := WorkflowPanel(nil, nil)

	if !strings.Contains(result, "Workflow") {
		t.Error("expected 'Workflow' heading")
	}
	if !strings.Contains(result, "idle") {
		t.Error("expected 'idle' state")
	}
}

func TestWorkflowPanel_IdleWithLastIteration(t *testing.T) {
	lastIter := &domain.Iteration{
		DirName: "001-NEW_PROJECT-initial-setup",
		Status:  domain.IterComplete,
	}
	result := WorkflowPanel(nil, lastIter)

	if !strings.Contains(result, "idle") {
		t.Error("expected 'idle' state")
	}
	if !strings.Contains(result, "001-NEW_PROJECT-initial-setup") {
		t.Error("expected last iteration name in output")
	}
}

func TestWorkflowPanel_Running(t *testing.T) {
	ws := &domain.WorkflowState{
		Type:           domain.TypeEnhancement,
		LastAgent:      "architect",
		Branch:         "feature/add-caching",
		RemainingChain: []string{"developer", "tester"},
	}
	result := WorkflowPanel(ws, nil)

	if !strings.Contains(result, "running") {
		t.Error("expected 'running' state")
	}
	if !strings.Contains(result, string(domain.TypeEnhancement)) {
		t.Error("expected workflow type")
	}
	if !strings.Contains(result, "architect") {
		t.Error("expected current agent")
	}
	if !strings.Contains(result, "feature/add-caching") {
		t.Error("expected branch name")
	}
}

// --- Warnings tests ---

func TestWarnings_Empty(t *testing.T) {
	result := Warnings(nil)
	if result != "" {
		t.Errorf("expected empty for no warnings, got: %q", result)
	}

	result = Warnings([]string{})
	if result != "" {
		t.Errorf("expected empty for empty slice, got: %q", result)
	}
}

func TestWarnings_WithItems(t *testing.T) {
	warnings := []string{"Missing project brief", "No mind.toml found"}
	result := Warnings(warnings)

	if !strings.Contains(result, "Warnings") {
		t.Error("expected 'Warnings' heading")
	}
	if !strings.Contains(result, "Missing project brief") {
		t.Error("expected warning message in output")
	}
	if !strings.Contains(result, "No mind.toml found") {
		t.Error("expected warning message in output")
	}
}

// --- Suggestions tests ---

func TestSuggestions_Empty(t *testing.T) {
	result := Suggestions(nil)
	if result != "" {
		t.Errorf("expected empty for no suggestions, got: %q", result)
	}
}

func TestSuggestions_WithItems(t *testing.T) {
	suggestions := []string{"Run mind init", "Create a project brief"}
	result := Suggestions(suggestions)

	if !strings.Contains(result, "Suggestions") {
		t.Error("expected 'Suggestions' heading")
	}
	if !strings.Contains(result, "Run mind init") {
		t.Error("expected suggestion in output")
	}
}

// --- QuickActions tests ---

func TestQuickActions_RendersAllActions(t *testing.T) {
	result := QuickActions()

	if !strings.Contains(result, "Quick Actions") {
		t.Error("expected 'Quick Actions' heading")
	}

	expectedKeys := []string{"r", "2", "4", "?", "q"}
	for _, key := range expectedKeys {
		if !strings.Contains(result, key) {
			t.Errorf("expected key '%s' in quick actions", key)
		}
	}
}

// --- EmptyState tests ---

func TestEmptyState_ContainsMessage(t *testing.T) {
	result := EmptyState("No data available.", 80, 24)
	if !strings.Contains(result, "No data available.") {
		t.Error("expected message in empty state output")
	}
}

func TestEmptyState_SmallDimensions(t *testing.T) {
	// Should not panic with very small dimensions
	result := EmptyState("Test", 10, 3)
	if !strings.Contains(result, "Test") {
		t.Error("expected message even with small dimensions")
	}
}
