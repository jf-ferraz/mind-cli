package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

func sampleIterations() []domain.Iteration {
	return []domain.Iteration{
		{
			Seq:        1,
			Type:       domain.TypeNewProject,
			Descriptor: "initial-setup",
			DirName:    "001-NEW_PROJECT-initial-setup",
			Status:     domain.IterComplete,
			Artifacts: []domain.Artifact{
				{Name: "requirements-delta.md", Exists: true},
				{Name: "architecture-delta.md", Exists: true},
				{Name: "changes.md", Exists: true},
				{Name: "validation.md", Exists: true},
				{Name: "test-summary.md", Exists: true},
			},
		},
		{
			Seq:        2,
			Type:       domain.TypeEnhancement,
			Descriptor: "add-caching",
			DirName:    "002-ENHANCEMENT-add-caching",
			Status:     domain.IterInProgress,
			Artifacts: []domain.Artifact{
				{Name: "requirements-delta.md", Exists: true},
				{Name: "architecture-delta.md", Exists: true},
				{Name: "changes.md", Exists: false},
				{Name: "validation.md", Exists: false},
				{Name: "test-summary.md", Exists: false},
			},
		},
		{
			Seq:        3,
			Type:       domain.TypeBugFix,
			Descriptor: "fix-crash",
			DirName:    "003-BUG_FIX-fix-crash",
			Status:     domain.IterIncomplete,
			Artifacts: []domain.Artifact{
				{Name: "requirements-delta.md", Exists: false},
				{Name: "architecture-delta.md", Exists: false},
				{Name: "changes.md", Exists: false},
				{Name: "validation.md", Exists: false},
				{Name: "test-summary.md", Exists: false},
			},
		},
	}
}

func TestIterationsView_InitialState(t *testing.T) {
	v := NewIterationsView()
	if v.viewState != ViewLoading {
		t.Errorf("initial viewState = %d, want ViewLoading", v.viewState)
	}
	if v.expandedIndex != -1 {
		t.Errorf("initial expandedIndex = %d, want -1", v.expandedIndex)
	}
}

func TestIterationsView_IterationsLoaded(t *testing.T) {
	v := NewIterationsView()
	iters := sampleIterations()

	v, _ = v.Update(iterationsLoadedMsg{iterations: iters})

	if v.viewState != ViewReady {
		t.Errorf("viewState = %d, want ViewReady", v.viewState)
	}
	if len(v.filtered) != 3 {
		t.Errorf("filtered = %d, want 3", len(v.filtered))
	}
}

func TestIterationsView_EmptyIterations(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsLoadedMsg{iterations: nil})

	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty", v.viewState)
	}
}

func TestIterationsView_Error(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsErrorMsg{})

	if v.viewState != ViewError {
		t.Errorf("viewState = %d, want ViewError", v.viewState)
	}
}

// FR-108: Type filter tests.
func TestIterationsView_TypeFilter(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsLoadedMsg{iterations: sampleIterations()})

	// Filter to ENHANCEMENT
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if len(v.filtered) != 1 {
		t.Errorf("ENHANCEMENT filter: %d filtered, want 1", len(v.filtered))
	}
	if v.filtered[0].Type != domain.TypeEnhancement {
		t.Errorf("filtered type = %q, want ENHANCEMENT", v.filtered[0].Type)
	}

	// Filter to BUG_FIX
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if len(v.filtered) != 1 {
		t.Errorf("BUG_FIX filter: %d filtered, want 1", len(v.filtered))
	}

	// Back to all
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(v.filtered) != 3 {
		t.Errorf("all filter: %d filtered, want 3", len(v.filtered))
	}

	// NEW_PROJECT
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if len(v.filtered) != 1 {
		t.Errorf("NEW_PROJECT filter: %d filtered, want 1", len(v.filtered))
	}

	// REFACTOR (none exist)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if len(v.filtered) != 0 {
		t.Errorf("REFACTOR filter: %d filtered, want 0", len(v.filtered))
	}
}

// FR-107: Navigation with j/k and cursor bounds.
func TestIterationsView_Navigation(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsLoadedMsg{iterations: sampleIterations()})

	// Initial cursor at 0
	if v.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", v.cursor)
	}

	// Move down
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", v.cursor)
	}

	// Move down again
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.cursor != 2 {
		t.Errorf("after 2nd j: cursor = %d, want 2", v.cursor)
	}

	// Move down at bottom -- should not exceed bounds
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.cursor != 2 {
		t.Errorf("cursor should stay at max: %d, want 2", v.cursor)
	}

	// Move up
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.cursor != 1 {
		t.Errorf("after k: cursor = %d, want 1", v.cursor)
	}

	// Move up with arrow
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyUp})
	if v.cursor != 0 {
		t.Errorf("after up: cursor = %d, want 0", v.cursor)
	}

	// Move up at top
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyUp})
	if v.cursor != 0 {
		t.Errorf("cursor should stay at 0: %d", v.cursor)
	}
}

// FR-109: Expand/collapse detail.
func TestIterationsView_ExpandCollapse(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsLoadedMsg{iterations: sampleIterations()})

	if v.expandedIndex != -1 {
		t.Errorf("initial expandedIndex = %d, want -1", v.expandedIndex)
	}

	// Enter to expand
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.expandedIndex != 0 {
		t.Errorf("after Enter: expandedIndex = %d, want 0", v.expandedIndex)
	}

	// Enter again to collapse
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.expandedIndex != -1 {
		t.Errorf("after 2nd Enter: expandedIndex = %d, want -1", v.expandedIndex)
	}
}

func TestIterationsView_WindowResize(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	if v.width != 100 {
		t.Errorf("width = %d, want 100", v.width)
	}
}

func TestIterationsView_ViewEmpty(t *testing.T) {
	v := NewIterationsView()
	v.viewState = ViewEmpty
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "No iterations") {
		t.Error("expected empty state message")
	}
}

// FR-108: Filter resets expanded state and adjusts cursor.
func TestIterationsView_FilterResetsExpanded(t *testing.T) {
	v := NewIterationsView()
	v, _ = v.Update(iterationsLoadedMsg{iterations: sampleIterations()})

	// Expand iteration 0
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.expandedIndex != 0 {
		t.Fatal("expected expanded index 0")
	}

	// Apply filter
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if v.expandedIndex != -1 {
		t.Errorf("expandedIndex should reset to -1 after filter, got %d", v.expandedIndex)
	}
}
