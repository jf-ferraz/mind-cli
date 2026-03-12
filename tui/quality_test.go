package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

func sampleEntries() []domain.QualityEntry {
	return []domain.QualityEntry{
		{
			Topic:    "auth-strategy",
			Variant:  "convergence-v1",
			Date:     time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Score:    4.0,
			GatePass: true,
			Dimensions: []domain.QualityDimension{
				{Name: "rigor", Value: 4},
				{Name: "coverage", Value: 4},
				{Name: "actionability", Value: 4},
				{Name: "objectivity", Value: 4},
				{Name: "convergence", Value: 4},
				{Name: "depth", Value: 4},
			},
			Personas:   []string{"moderator", "analyst"},
			OutputPath: "docs/knowledge/auth.md",
		},
		{
			Topic:    "reconciliation",
			Variant:  "convergence-v1",
			Date:     time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			Score:    4.33,
			GatePass: true,
			Dimensions: []domain.QualityDimension{
				{Name: "rigor", Value: 5},
				{Name: "coverage", Value: 4},
				{Name: "actionability", Value: 4},
				{Name: "objectivity", Value: 5},
				{Name: "convergence", Value: 4},
				{Name: "depth", Value: 4},
			},
			Personas:   []string{"moderator", "analyst", "architect"},
			OutputPath: "docs/knowledge/reconciliation.md",
		},
		{
			Topic:    "tui-dashboard",
			Variant:  "convergence-v1",
			Date:     time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			Score:    2.5,
			GatePass: false,
			Dimensions: []domain.QualityDimension{
				{Name: "rigor", Value: 3},
				{Name: "coverage", Value: 2},
				{Name: "actionability", Value: 3},
				{Name: "objectivity", Value: 2},
				{Name: "convergence", Value: 2},
				{Name: "depth", Value: 3},
			},
			Personas:   []string{"moderator"},
			OutputPath: "docs/knowledge/tui.md",
		},
	}
}

func TestQualityView_InitialState(t *testing.T) {
	v := NewQualityView()
	if v.viewState != ViewLoading {
		t.Errorf("initial viewState = %d, want ViewLoading", v.viewState)
	}
}

func TestQualityView_Loaded(t *testing.T) {
	v := NewQualityView()
	entries := sampleEntries()

	v, _ = v.Update(qualityLoadedMsg{entries: entries})

	if v.viewState != ViewReady {
		t.Errorf("viewState = %d, want ViewReady", v.viewState)
	}
	if len(v.entries) != 3 {
		t.Errorf("entries = %d, want 3", len(v.entries))
	}
	// Selected index should be last entry
	if v.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2 (last entry)", v.selectedIndex)
	}
}

// FR-115: Empty state for no quality data.
func TestQualityView_Empty(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(qualityLoadedMsg{entries: nil})

	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty", v.viewState)
	}
}

func TestQualityView_EmptySlice(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(qualityLoadedMsg{entries: []domain.QualityEntry{}})

	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty", v.viewState)
	}
}

func TestQualityView_Error(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(qualityErrorMsg{})

	if v.viewState != ViewError {
		t.Errorf("viewState = %d, want ViewError", v.viewState)
	}
}

// FR-114: Navigate with left/right or h/l.
func TestQualityView_Navigation(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(qualityLoadedMsg{entries: sampleEntries()})

	// Start at last (index 2)
	if v.selectedIndex != 2 {
		t.Fatalf("selectedIndex = %d, want 2", v.selectedIndex)
	}

	// Left
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if v.selectedIndex != 1 {
		t.Errorf("after left: selectedIndex = %d, want 1", v.selectedIndex)
	}

	// h
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if v.selectedIndex != 0 {
		t.Errorf("after h: selectedIndex = %d, want 0", v.selectedIndex)
	}

	// Left at leftmost -- no change
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if v.selectedIndex != 0 {
		t.Errorf("at min: selectedIndex = %d, want 0", v.selectedIndex)
	}

	// Right
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRight})
	if v.selectedIndex != 1 {
		t.Errorf("after right: selectedIndex = %d, want 1", v.selectedIndex)
	}

	// l
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if v.selectedIndex != 2 {
		t.Errorf("after l: selectedIndex = %d, want 2", v.selectedIndex)
	}

	// Right at rightmost -- no change
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRight})
	if v.selectedIndex != 2 {
		t.Errorf("at max: selectedIndex = %d, want 2", v.selectedIndex)
	}
}

func TestQualityView_WindowResize(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if v.width != 120 || v.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", v.width, v.height)
	}
}

// FR-115: Empty state view.
func TestQualityView_ViewEmpty(t *testing.T) {
	v := NewQualityView()
	v.viewState = ViewEmpty
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "No quality data") {
		t.Error("expected empty state message")
	}
	if !strings.Contains(output, "mind quality log") {
		t.Error("expected 'mind quality log' hint in empty state")
	}
}

// FR-114: Chart and detail in ready view.
func TestQualityView_ViewReady(t *testing.T) {
	v := NewQualityView()
	v, _ = v.Update(qualityLoadedMsg{entries: sampleEntries()})
	v.width = 100
	v.height = 40

	output := v.View()
	if !strings.Contains(output, "Score History") {
		t.Error("expected 'Score History' chart heading")
	}
	if !strings.Contains(output, "Selected Analysis") {
		t.Error("expected 'Selected Analysis' detail heading")
	}
	// The selected entry is the last one (tui-dashboard)
	if !strings.Contains(output, "tui-dashboard") {
		t.Error("expected selected entry topic in detail")
	}
}
