package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

func sampleHealthMsg() healthLoadedMsg {
	return healthLoadedMsg{
		health: &domain.ProjectHealth{
			Zones: map[domain.Zone]domain.ZoneHealth{
				domain.ZoneSpec: {
					Zone:  domain.ZoneSpec,
					Total: 3,
					Files: []domain.Document{
						{Name: "architecture", Zone: domain.ZoneSpec, Path: "docs/spec/architecture.md", Size: 1000},
						{Name: "requirements", Zone: domain.ZoneSpec, Path: "docs/spec/requirements.md", Size: 2000},
						{Name: "domain-model", Zone: domain.ZoneSpec, Path: "docs/spec/domain-model.md", Size: 1500, IsStub: true},
					},
				},
				domain.ZoneBlueprints: {
					Zone:  domain.ZoneBlueprints,
					Total: 1,
					Files: []domain.Document{
						{Name: "INDEX", Zone: domain.ZoneBlueprints, Path: "docs/blueprints/INDEX.md", Size: 500},
					},
				},
				domain.ZoneState: {
					Zone:  domain.ZoneState,
					Total: 2,
					Files: []domain.Document{
						{Name: "current", Zone: domain.ZoneState, Path: "docs/state/current.md", Size: 800},
						{Name: "workflow", Zone: domain.ZoneState, Path: "docs/state/workflow.md", Size: 400},
					},
				},
			},
		},
	}
}

func TestDocsView_InitialState(t *testing.T) {
	v := NewDocsView(nil)
	if v.viewState != ViewLoading {
		t.Errorf("initial viewState = %d, want ViewLoading", v.viewState)
	}
}

func TestDocsView_HealthLoaded(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	if v.viewState != ViewReady {
		t.Errorf("viewState = %d, want ViewReady", v.viewState)
	}
	// 3 spec + 1 blueprints + 2 state = 6
	if len(v.documents) != 6 {
		t.Errorf("documents = %d, want 6", len(v.documents))
	}
	if len(v.filtered) != 6 {
		t.Errorf("filtered = %d, want 6", len(v.filtered))
	}
}

func TestDocsView_HealthNil(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(healthLoadedMsg{health: nil})

	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty", v.viewState)
	}
}

func TestDocsView_HealthError(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(healthErrorMsg{})

	if v.viewState != ViewError {
		t.Errorf("viewState = %d, want ViewError", v.viewState)
	}
}

// FR-103: Zone filter.
func TestDocsView_ZoneFilter(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	// Filter to spec
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if len(v.filtered) != 3 {
		t.Errorf("spec filter: %d filtered, want 3", len(v.filtered))
	}
	for _, doc := range v.filtered {
		if doc.Zone != domain.ZoneSpec {
			t.Errorf("filtered doc zone = %q, want spec", doc.Zone)
		}
	}

	// Filter to blueprints
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if len(v.filtered) != 1 {
		t.Errorf("blueprints filter: %d filtered, want 1", len(v.filtered))
	}

	// Filter to state
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if len(v.filtered) != 2 {
		t.Errorf("state filter: %d filtered, want 2", len(v.filtered))
	}

	// Back to all
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(v.filtered) != 6 {
		t.Errorf("all filter: %d filtered, want 6", len(v.filtered))
	}

	// Filter to knowledge (none in sample)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if len(v.filtered) != 0 {
		t.Errorf("knowledge filter: %d filtered, want 0", len(v.filtered))
	}

	// Filter to iterations (none in sample)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if len(v.filtered) != 0 {
		t.Errorf("iterations filter: %d filtered, want 0", len(v.filtered))
	}
}

// FR-104: Search mode.
func TestDocsView_SearchActivation(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	if v.searchActive {
		t.Error("search should not be active initially")
	}

	// Activate search with /
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !v.searchActive {
		t.Error("search should be active after /")
	}

	// Type "arch" while in search mode
	for _, r := range "arch" {
		v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if v.searchQuery != "arch" {
		t.Errorf("searchQuery = %q, want 'arch'", v.searchQuery)
	}
	if len(v.filtered) != 1 {
		t.Errorf("filtered with 'arch' = %d, want 1", len(v.filtered))
	}

	// Backspace
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if v.searchQuery != "arc" {
		t.Errorf("after backspace: searchQuery = %q, want 'arc'", v.searchQuery)
	}

	// Esc to clear search
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if v.searchActive {
		t.Error("search should be inactive after Esc")
	}
	if v.searchQuery != "" {
		t.Errorf("searchQuery = %q, want '' after Esc", v.searchQuery)
	}
	if len(v.filtered) != 6 {
		t.Errorf("after Esc: filtered = %d, want 6 (all)", len(v.filtered))
	}
}

// FR-104: Search is case-insensitive on filenames.
func TestDocsView_SearchCaseInsensitive(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	// Activate search and type "ARCH" (uppercase)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "ARCH" {
		v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Should match "architecture"
	if len(v.filtered) != 1 {
		t.Errorf("uppercase search: filtered = %d, want 1", len(v.filtered))
	}
}

// Navigation in normal mode.
func TestDocsView_Navigation(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	if v.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", v.cursor)
	}

	// Down
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyDown})
	if v.cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", v.cursor)
	}

	// j
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.cursor != 2 {
		t.Errorf("after j: cursor = %d, want 2", v.cursor)
	}

	// Up
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyUp})
	if v.cursor != 1 {
		t.Errorf("after up: cursor = %d, want 1", v.cursor)
	}
}

// Cursor bounds when filter reduces list.
func TestDocsView_CursorResetOnFilter(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	// Move cursor to position 5 (last of 6)
	for i := 0; i < 5; i++ {
		v, _ = v.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if v.cursor != 5 {
		t.Fatalf("cursor = %d, want 5", v.cursor)
	}

	// Filter to blueprints (1 doc)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if v.cursor >= len(v.filtered) {
		t.Errorf("cursor %d >= filtered length %d", v.cursor, len(v.filtered))
	}
}

func TestDocsView_WindowResize(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if v.width != 100 {
		t.Errorf("width = %d, want 100", v.width)
	}
}

func TestDocsView_ViewLoading(t *testing.T) {
	v := NewDocsView(nil)
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "Loading") {
		t.Error("expected loading message")
	}
}

// Preview loaded message.
func TestDocsView_PreviewLoaded(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(previewLoadedMsg{content: "# Architecture\n\nContent here."})

	if !v.previewVisible {
		t.Error("preview should be visible after previewLoadedMsg")
	}
	if v.previewContent != "# Architecture\n\nContent here." {
		t.Errorf("previewContent = %q", v.previewContent)
	}
}

// Preview error.
func TestDocsView_PreviewError(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(previewErrorMsg{err: fmt.Errorf("read error")})

	if !v.previewVisible {
		t.Error("preview should be visible even on error")
	}
	if !strings.Contains(v.previewContent, "read error") {
		t.Errorf("previewContent should contain error, got: %q", v.previewContent)
	}
}

// Esc closes preview.
func TestDocsView_PreviewCloseOnEsc(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())
	v.previewVisible = true

	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if v.previewVisible {
		t.Error("preview should be closed after Esc")
	}
}

// Enter in search mode deactivates search but keeps query.
func TestDocsView_SearchEnterDeactivates(t *testing.T) {
	v := NewDocsView(nil)
	v, _ = v.Update(sampleHealthMsg())

	// Activate and type
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.searchActive {
		t.Error("search should be deactivated after Enter")
	}
	// Query stays, unlike Esc which clears it
	if v.searchQuery != "a" {
		t.Errorf("searchQuery = %q, want 'a' (preserved after Enter)", v.searchQuery)
	}
}
