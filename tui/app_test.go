package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-118: Terminal too small renders warning.
func TestApp_TooSmall(t *testing.T) {
	a := App{width: 79, height: 24}
	output := a.View()
	if !strings.Contains(output, "Terminal too small") {
		t.Error("expected 'Terminal too small' message")
	}
	if !strings.Contains(output, "79x24") {
		t.Error("expected current dimensions in message")
	}
	if !strings.Contains(output, "80x24") {
		t.Error("expected minimum dimensions in message")
	}
}

func TestApp_TooShort(t *testing.T) {
	a := App{width: 80, height: 23}
	output := a.View()
	if !strings.Contains(output, "Terminal too small") {
		t.Error("expected 'Terminal too small' message for height < 24")
	}
}

func TestApp_MinimalValid(t *testing.T) {
	a := App{width: 80, height: 24, activeTab: TabStatus}
	// Set status to have some data so it doesn't panic
	a.status = NewStatusView()
	a.status.viewState = ViewLoading
	a.status.width = 80
	a.status.height = 20

	output := a.View()
	// Should not show "too small"
	if strings.Contains(output, "Terminal too small") {
		t.Error("80x24 should be valid, not too small")
	}
}

// FR-94: Tab switching with number keys.
func TestApp_TabSwitching(t *testing.T) {
	a := App{activeTab: TabStatus, width: 100, height: 40}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	tests := []struct {
		key     tea.KeyMsg
		wantTab TabID
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, TabDocs},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}, TabIterations},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}, TabChecks},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}, TabQuality},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}, TabStatus},
	}

	for _, tt := range tests {
		model, _ := a.Update(tt.key)
		a = model.(App)
		if a.activeTab != tt.wantTab {
			t.Errorf("after key %q: activeTab = %d, want %d", tt.key.String(), a.activeTab, tt.wantTab)
		}
	}
}

// FR-94: Tab wrapping.
func TestApp_TabWrap(t *testing.T) {
	a := App{activeTab: TabQuality, width: 100, height: 40}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	// Tab from last tab wraps to first
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = model.(App)
	if a.activeTab != TabStatus {
		t.Errorf("Tab from Quality: activeTab = %d, want TabStatus (0)", a.activeTab)
	}

	// Shift+Tab from first tab wraps to last
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	a = model.(App)
	if a.activeTab != TabQuality {
		t.Errorf("Shift+Tab from Status: activeTab = %d, want TabQuality (4)", a.activeTab)
	}
}

// FR-95: Ctrl+C always force-quits.
func TestApp_ForceQuit(t *testing.T) {
	a := App{width: 100, height: 40}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C should produce a quit command")
	}
}

// FR-95: q quits when no overlay.
func TestApp_QuitNoOverlay(t *testing.T) {
	a := App{width: 100, height: 40, showHelp: false}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q should produce a quit command when help is closed")
	}
}

// FR-117: Help overlay.
func TestApp_HelpToggle(t *testing.T) {
	a := App{width: 100, height: 40, showHelp: false}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	// Open help with ?
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	a = model.(App)
	if !a.showHelp {
		t.Error("? should open help overlay")
	}

	// q closes help (does not quit)
	model, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	a = model.(App)
	if a.showHelp {
		t.Error("q should close help overlay")
	}
	if cmd != nil {
		t.Error("q should NOT quit when help overlay is open")
	}
}

// FR-117: Esc closes help.
func TestApp_HelpCloseEsc(t *testing.T) {
	a := App{width: 100, height: 40, showHelp: true}
	a.status = NewStatusView()
	a.help = NewHelpModel()

	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyEsc})
	a = model.(App)
	if a.showHelp {
		t.Error("Esc should close help overlay")
	}
}

// FR-117: ? closes help when already open.
func TestApp_HelpCloseQuestion(t *testing.T) {
	a := App{width: 100, height: 40, showHelp: true}
	a.status = NewStatusView()
	a.help = NewHelpModel()

	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	a = model.(App)
	if a.showHelp {
		t.Error("? should close help when already open")
	}
}

// FR-119: WindowSizeMsg updates dimensions.
func TestApp_WindowResize(t *testing.T) {
	a := App{}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	model, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	a = model.(App)
	if a.width != 120 || a.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", a.width, a.height)
	}
}

// FR-123: Tab state preservation -- tabs are not recreated on switch.
func TestApp_TabStatePreserved(t *testing.T) {
	a := App{activeTab: TabIterations, width: 100, height: 40}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	// Load iterations and navigate
	a.iters, _ = a.iters.Update(iterationsLoadedMsg{iterations: sampleIterations()})
	a.iters, _ = a.iters.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	savedCursor := a.iters.cursor

	// Switch to Status
	model, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	a = model.(App)

	// Switch back to Iterations
	model, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	a = model.(App)

	if a.iters.cursor != savedCursor {
		t.Errorf("cursor = %d, want %d (preserved across tab switch)", a.iters.cursor, savedCursor)
	}
}

// Data message routing.
func TestApp_HealthMessageRouting(t *testing.T) {
	a := App{width: 100, height: 40}
	a.status = NewStatusView()
	a.docs = NewDocsView(nil)
	a.iters = NewIterationsView()
	a.checks = NewChecksView(nil)
	a.quality = NewQualityView()
	a.help = NewHelpModel()

	health := &domain.ProjectHealth{
		Zones: map[domain.Zone]domain.ZoneHealth{
			domain.ZoneSpec: {Zone: domain.ZoneSpec, Total: 3, Complete: 3},
		},
	}

	model, _ := a.Update(healthLoadedMsg{health: health})
	a = model.(App)

	if !a.loaded {
		t.Error("loaded should be true after healthLoadedMsg")
	}
	if a.status.viewState != ViewReady {
		t.Error("status should be ViewReady after health loaded")
	}
}

// overlayOnScreen utility.
func TestOverlayOnScreen(t *testing.T) {
	bg := "Hello World\nSecond Line"
	overlay := "XX"
	result := overlayOnScreen(bg, overlay, 2, 0)

	lines := strings.Split(result, "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	if !strings.Contains(lines[0], "XX") {
		t.Errorf("expected overlay 'XX' in first line, got: %q", lines[0])
	}
}

func TestOverlayOnScreen_OutOfBounds(t *testing.T) {
	bg := "Hello"
	overlay := "XX"
	// Place overlay out of bounds (y=5 when bg has only 1 line)
	result := overlayOnScreen(bg, overlay, 0, 5)
	// Should not panic or change content
	if result != bg {
		t.Errorf("out-of-bounds overlay should not change background")
	}
}

// detectBranch with no git dir returns empty.
func TestDetectBranch_NoGit(t *testing.T) {
	branch := detectBranch(t.TempDir())
	if branch != "" {
		t.Errorf("branch = %q, want '' for non-git dir", branch)
	}
}
