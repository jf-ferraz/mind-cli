package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

func TestChecksView_InitialState(t *testing.T) {
	v := NewChecksView(nil)
	if v.viewState != ViewLoading {
		t.Errorf("initial viewState = %d, want ViewLoading", v.viewState)
	}
	if v.activated {
		t.Error("should not be activated initially")
	}
}

// FR-122: Lazy loading -- validation runs only when tab first activated.
func TestChecksView_LazyLoading(t *testing.T) {
	v := NewChecksView(nil)

	// Activation of a different tab should not trigger validation
	v2, _ := v.Update(tabActivatedMsg{tab: TabStatus})
	if v2.activated {
		t.Error("should not activate for non-checks tab")
	}
	if v2.loading {
		t.Error("should not load for non-checks tab")
	}
}

// FR-111: Validation complete populates report.
func TestChecksView_ValidationComplete(t *testing.T) {
	v := NewChecksView(nil)
	report := domain.UnifiedValidationReport{
		Suites: []domain.ValidationReport{
			{Suite: "docs", Total: 17, Passed: 15, Failed: 1, Warnings: 1},
			{Suite: "refs", Total: 11, Passed: 11},
			{Suite: "config", Total: 12, Passed: 12},
		},
		Summary: domain.UnifiedValidationSummary{
			Total: 40, Passed: 38, Failed: 1, Warnings: 1,
		},
	}

	v, _ = v.Update(validationCompleteMsg{report: report})

	if v.viewState != ViewReady {
		t.Errorf("viewState = %d, want ViewReady", v.viewState)
	}
	if v.loading {
		t.Error("loading should be false after completion")
	}
	if v.report == nil {
		t.Fatal("report is nil")
	}
	if len(v.report.Suites) != 3 {
		t.Errorf("suites = %d, want 3", len(v.report.Suites))
	}
}

func TestChecksView_ValidationEmpty(t *testing.T) {
	v := NewChecksView(nil)
	report := domain.UnifiedValidationReport{
		Summary: domain.UnifiedValidationSummary{Total: 0},
	}

	v, _ = v.Update(validationCompleteMsg{report: report})
	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty for zero checks", v.viewState)
	}
}

func TestChecksView_ValidationError(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(validationErrorMsg{})

	if v.viewState != ViewError {
		t.Errorf("viewState = %d, want ViewError", v.viewState)
	}
	if v.loading {
		t.Error("loading should be false after error")
	}
}

// FR-110: Expand/collapse suite accordion.
func TestChecksView_ExpandCollapseSuite(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(validationCompleteMsg{
		report: domain.UnifiedValidationReport{
			Suites: []domain.ValidationReport{
				{
					Suite: "docs", Total: 3, Passed: 3,
					Checks: []domain.CheckResult{
						{ID: 1, Name: "Check 1", Passed: true},
						{ID: 2, Name: "Check 2", Passed: true},
						{ID: 3, Name: "Check 3", Passed: true},
					},
				},
				{Suite: "refs", Total: 2, Passed: 2,
					Checks: []domain.CheckResult{
						{ID: 1, Name: "Ref 1", Passed: true},
						{ID: 2, Name: "Ref 2", Passed: true},
					},
				},
			},
			Summary: domain.UnifiedValidationSummary{Total: 5, Passed: 5},
		},
	})

	// Cursor at 0 = first suite header
	if v.expandedSuites[0] {
		t.Error("suite 0 should not be expanded initially")
	}

	// Press Enter to expand
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !v.expandedSuites[0] {
		t.Error("suite 0 should be expanded after Enter")
	}

	// Press Enter again to collapse
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.expandedSuites[0] {
		t.Error("suite 0 should be collapsed after 2nd Enter")
	}
}

// FR-112: Check detail toggle.
func TestChecksView_DetailToggle(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(validationCompleteMsg{
		report: domain.UnifiedValidationReport{
			Suites: []domain.ValidationReport{
				{Suite: "docs", Total: 1, Passed: 1, Checks: []domain.CheckResult{
					{ID: 1, Name: "Check", Passed: true},
				}},
			},
			Summary: domain.UnifiedValidationSummary{Total: 1, Passed: 1},
		},
	})

	if v.detailVisible {
		t.Error("detail should not be visible initially")
	}

	// Toggle detail with Space
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !v.detailVisible {
		t.Error("detail should be visible after Space")
	}
	if v.detailTarget != v.cursor {
		t.Errorf("detailTarget = %d, want %d (cursor)", v.detailTarget, v.cursor)
	}

	// Toggle off
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeySpace})
	if v.detailVisible {
		t.Error("detail should be hidden after 2nd Space")
	}
}

// Navigation with j/k.
func TestChecksView_Navigation(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(validationCompleteMsg{
		report: domain.UnifiedValidationReport{
			Suites: []domain.ValidationReport{
				{Suite: "docs", Total: 1, Checks: []domain.CheckResult{{ID: 1}}},
				{Suite: "refs", Total: 1, Checks: []domain.CheckResult{{ID: 1}}},
			},
			Summary: domain.UnifiedValidationSummary{Total: 2},
		},
	})

	if v.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", v.cursor)
	}

	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", v.cursor)
	}

	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", v.cursor)
	}

	// Up at 0 stays at 0
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyUp})
	if v.cursor != 0 {
		t.Errorf("up at top: cursor = %d, want 0", v.cursor)
	}
}

func TestChecksView_WindowResize(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if v.width != 100 {
		t.Errorf("width = %d, want 100", v.width)
	}
}

// FR-113: Overall summary in view.
func TestChecksView_ViewShowsSummary(t *testing.T) {
	v := NewChecksView(nil)
	v, _ = v.Update(validationCompleteMsg{
		report: domain.UnifiedValidationReport{
			Suites: []domain.ValidationReport{
				{
					Suite: "docs", Total: 17, Passed: 15, Failed: 1, Warnings: 1,
					Checks: []domain.CheckResult{
						{ID: 1, Name: "Zone dirs", Passed: true},
					},
				},
			},
			Summary: domain.UnifiedValidationSummary{Total: 17, Passed: 15, Failed: 1, Warnings: 1},
		},
	})
	v.width = 100
	v.height = 40

	output := v.View()
	if !strings.Contains(output, "Overall") {
		t.Error("expected 'Overall' summary in output")
	}
	if !strings.Contains(output, "15/17 pass") {
		t.Error("expected '15/17 pass' in summary")
	}
}

// ViewLoading (not yet activated).
func TestChecksView_ViewBeforeActivation(t *testing.T) {
	v := NewChecksView(nil)
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "Switch to this tab") {
		t.Error("expected activation prompt in view before first activation")
	}
}
