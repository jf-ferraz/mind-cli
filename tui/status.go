package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/tui/components"
)

// StatusView is Tab 1: zone health, staleness, workflow, warnings, suggestions.
type StatusView struct {
	viewState ViewState
	health    *domain.ProjectHealth
	errMsg    string
	width     int
	height    int
}

// NewStatusView creates a StatusView.
func NewStatusView() StatusView {
	return StatusView{
		viewState: ViewLoading,
	}
}

// Init returns nil (data loads via app-level commands).
func (v StatusView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the Status tab.
func (v StatusView) Update(msg tea.Msg) (StatusView, tea.Cmd) {
	switch msg := msg.(type) {
	case healthLoadedMsg:
		if msg.health == nil {
			v.viewState = ViewEmpty
		} else {
			v.viewState = ViewReady
			v.health = msg.health
		}
	case healthErrorMsg:
		v.viewState = ViewError
		v.errMsg = msg.err.Error()
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	}
	return v, nil
}

// View renders the Status tab content.
func (v StatusView) View() string {
	switch v.viewState {
	case ViewLoading:
		return components.EmptyState("Loading project health...", v.width, v.height)
	case ViewError:
		return components.EmptyState("Error: "+v.errMsg+"\nPress r to retry", v.width, v.height)
	case ViewEmpty:
		return components.EmptyState("No project data available.", v.width, v.height)
	}

	h := v.health
	contentWidth := v.width - 2 // padding

	// Two-column layout at >= 80 cols
	if contentWidth >= MinWidth {
		leftWidth := contentWidth / 2
		rightWidth := contentWidth - leftWidth

		left := v.renderLeftColumn(leftWidth)
		right := v.renderRightColumn(rightWidth)

		leftCol := lipgloss.NewStyle().Width(leftWidth).Render(left)
		rightCol := lipgloss.NewStyle().Width(rightWidth).Render(right)

		_ = h
		return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	}

	// Single-column fallback
	return v.renderLeftColumn(contentWidth) + "\n" + v.renderRightColumn(contentWidth)
}

func (v StatusView) renderLeftColumn(width int) string {
	h := v.health
	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true)
	b.WriteString(heading.Render("Documentation Health"))
	b.WriteString("\n")

	for _, zone := range domain.AllZones {
		zh, ok := h.Zones[zone]
		if !ok {
			zh = domain.ZoneHealth{Zone: zone}
		}
		b.WriteString("  " + components.ZoneBar(zh, width) + "\n")
	}
	b.WriteString("\n")

	// Staleness
	if h.Staleness != nil && len(h.Staleness.Stale) > 0 {
		b.WriteString(components.Staleness(h.Staleness.Stale, width))
		b.WriteString("\n")
	}

	// Warnings
	if len(h.Warnings) > 0 {
		b.WriteString(components.Warnings(h.Warnings))
		b.WriteString("\n")
	}

	// Suggestions
	if len(h.Suggestions) > 0 {
		b.WriteString(components.Suggestions(h.Suggestions))
	}

	return b.String()
}

func (v StatusView) renderRightColumn(width int) string {
	h := v.health
	var b strings.Builder

	b.WriteString(components.WorkflowPanel(h.Workflow, h.LastIteration))
	b.WriteString("\n")
	b.WriteString(components.QuickActions())

	return b.String()
}
