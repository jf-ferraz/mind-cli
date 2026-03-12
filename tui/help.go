package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModel renders a context-sensitive help overlay.
type HelpModel struct {
	width  int
	height int
}

// NewHelpModel creates a HelpModel.
func NewHelpModel() HelpModel {
	return HelpModel{}
}

// SetSize updates the overlay dimensions.
func (h *HelpModel) SetSize(w, h2 int) {
	h.width = w
	h.height = h2
}

// View renders the help overlay centered on screen.
func (h HelpModel) View(activeTab TabID) string {
	var b strings.Builder

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5f87ff")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cccccc"))
	headStyle := lipgloss.NewStyle().Bold(true).Underline(true)

	b.WriteString(headStyle.Render("Global Keys"))
	b.WriteString("\n\n")

	globals := []struct{ key, desc string }{
		{"1-5", "switch tab"},
		{"Tab/Shift+Tab", "cycle tabs"},
		{"r", "refresh data"},
		{"?", "toggle help"},
		{"q", "quit"},
		{"Ctrl+C", "force quit"},
	}
	for _, g := range globals {
		b.WriteString("  " + keyStyle.Render(padRight(g.key, 16)) + descStyle.Render(g.desc) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(headStyle.Render(TabNames[activeTab] + " Keys"))
	b.WriteString("\n\n")

	tabKeys := tabSpecificHelp(activeTab)
	for _, tk := range tabKeys {
		b.WriteString("  " + keyStyle.Render(padRight(tk.key, 16)) + descStyle.Render(tk.desc) + "\n")
	}

	content := b.String()

	// Overlay box: centered, with border
	overlayWidth := 50
	if h.width > 0 && overlayWidth > h.width-4 {
		overlayWidth = h.width - 4
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5f87ff")).
		Padding(1, 2).
		Width(overlayWidth)

	return boxStyle.Render(content)
}

type keyHelp struct{ key, desc string }

func tabSpecificHelp(tab TabID) []keyHelp {
	switch tab {
	case TabStatus:
		return []keyHelp{
			{"(no tab-specific keys)", ""},
		}
	case TabDocs:
		return []keyHelp{
			{"a/s/b/t/i/k", "zone filter"},
			{"/", "search"},
			{"Esc", "clear search/close preview"},
			{"Enter", "toggle preview"},
			{"e", "open in $EDITOR"},
			{"↑/↓ or j/k", "navigate"},
		}
	case TabIterations:
		return []keyHelp{
			{"a/n/e/b/r", "type filter"},
			{"Enter", "expand/collapse"},
			{"↑/↓ or j/k", "navigate"},
		}
	case TabChecks:
		return []keyHelp{
			{"Enter", "expand/collapse suite"},
			{"Space", "toggle check detail"},
			{"r", "re-run validation"},
			{"↑/↓ or j/k", "navigate"},
		}
	case TabQuality:
		return []keyHelp{
			{"←/→ or h/l", "select data point"},
			{"↑/↓ or j/k", "scroll details"},
		}
	}
	return nil
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
