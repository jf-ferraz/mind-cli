package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/tui/components"
)

// ChecksView is Tab 4: accordion validation suites with lazy loading.
type ChecksView struct {
	deps           *deps.Deps
	viewState      ViewState
	report         *domain.UnifiedValidationReport
	expandedSuites map[int]bool
	cursor         int
	detailVisible  bool
	detailTarget   int
	loading        bool
	activated      bool
	spinner        spinner.Model
	width          int
	height         int
}

// NewChecksView creates a ChecksView.
func NewChecksView(deps *deps.Deps) ChecksView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return ChecksView{
		deps:           deps,
		viewState:      ViewLoading,
		expandedSuites: make(map[int]bool),
		spinner:        sp,
	}
}

func (v ChecksView) Init() tea.Cmd { return nil }

func (v ChecksView) Update(msg tea.Msg) (ChecksView, tea.Cmd) {
	switch msg := msg.(type) {
	case tabActivatedMsg:
		if msg.tab == TabChecks && !v.activated {
			v.activated = true
			v.loading = true
			return v, tea.Batch(v.spinner.Tick, v.runValidation())
		}

	case validationCompleteMsg:
		v.loading = false
		v.report = &msg.report
		if v.report.Summary.Total == 0 {
			v.viewState = ViewEmpty
		} else {
			v.viewState = ViewReady
		}

	case validationErrorMsg:
		v.loading = false
		v.viewState = ViewError

	case spinner.TickMsg:
		if v.loading {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			return v, cmd
		}

	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		return v.handleKey(msg)
	}
	return v, nil
}

func (v ChecksView) handleKey(msg tea.KeyMsg) (ChecksView, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		v.cursor++
	case "enter":
		// Toggle suite expand/collapse
		suiteIdx := v.suiteIndexForCursor()
		if suiteIdx >= 0 {
			v.expandedSuites[suiteIdx] = !v.expandedSuites[suiteIdx]
		}
	case " ":
		// Toggle check detail
		v.detailVisible = !v.detailVisible
		v.detailTarget = v.cursor
	case "r":
		v.loading = true
		v.activated = true
		return v, tea.Batch(v.spinner.Tick, v.runValidation())
	}
	return v, nil
}

func (v ChecksView) suiteIndexForCursor() int {
	if v.report == nil {
		return -1
	}
	pos := 0
	for i, suite := range v.report.Suites {
		if pos == v.cursor {
			return i
		}
		pos++
		if v.expandedSuites[i] {
			pos += len(suite.Checks)
		}
	}
	return -1
}

func (v ChecksView) runValidation() tea.Cmd {
	return func() tea.Msg {
		// Run reconcile in check-only mode for the reconcile suite
		var reconcileResult *domain.ReconcileResult
		result, err := v.deps.ReconcileSvc.Reconcile(v.deps.ProjectRoot, domain.ReconcileOpts{CheckOnly: true})
		if err == nil {
			reconcileResult = result
		}

		report := v.deps.ValidationSvc.RunAll(v.deps.ProjectRoot, false, reconcileResult)
		return validationCompleteMsg{report: report}
	}
}

func (v ChecksView) View() string {
	if v.loading {
		return components.EmptyState(v.spinner.View()+" Running validation...", v.width, v.height)
	}

	switch v.viewState {
	case ViewLoading:
		return components.EmptyState("Switch to this tab to run validation.", v.width, v.height)
	case ViewError:
		return components.EmptyState("Error running validation.\nPress r to retry", v.width, v.height)
	case ViewEmpty:
		return components.EmptyState("No validation checks found.", v.width, v.height)
	}

	var b strings.Builder
	pos := 0
	for i, suite := range v.report.Suites {
		// Suite header
		expandIcon := "▸"
		if v.expandedSuites[i] {
			expandIcon = "▾"
		}
		suiteName := strings.ToUpper(suite.Suite[:1]) + suite.Suite[1:]
		summary := fmt.Sprintf("(%d checks: %d pass, %d fail, %d warn)",
			suite.Total, suite.Passed, suite.Failed, suite.Warnings)

		header := fmt.Sprintf("  %s %s %s", expandIcon, suiteName, summary)
		if pos == v.cursor {
			b.WriteString(theme.Selected.Render(header) + "\n")
		} else {
			b.WriteString(theme.Heading.Render(header) + "\n")
		}
		pos++

		// Expanded checks
		if v.expandedSuites[i] {
			for _, check := range suite.Checks {
				icon := theme.Pass.Render("✓")
				if !check.Passed {
					if check.Level == domain.LevelFail {
						icon = theme.Fail.Render("✗")
					} else {
						icon = theme.Warn.Render("⚠")
					}
				}
				line := fmt.Sprintf("    %s [%d] %s", icon, check.ID, check.Name)
				if pos == v.cursor {
					b.WriteString(theme.Selected.Render(line) + "\n")
				} else {
					b.WriteString(line + "\n")
				}

				// Detail pane
				if v.detailVisible && v.detailTarget == pos && !check.Passed && check.Message != "" {
					detail := theme.Border.Render(
						fmt.Sprintf("  Level: %s\n  %s", check.Level, check.Message))
					b.WriteString(detail + "\n")
				}
				pos++
			}
		}
	}

	// Overall summary
	b.WriteString("\n")
	s := v.report.Summary
	summaryLine := fmt.Sprintf("  Overall: %d/%d pass  %d fail  %d warning",
		s.Passed, s.Total, s.Failed, s.Warnings)
	if s.Failed == 0 {
		b.WriteString(theme.Pass.Render(summaryLine) + "\n")
	} else {
		b.WriteString(theme.Fail.Render(summaryLine) + "\n")
	}

	return b.String()
}
