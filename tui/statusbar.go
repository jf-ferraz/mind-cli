package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabHints provides context-sensitive key hints per tab.
var tabHints = [TabCount]string{
	"r:refresh  ?:help  q:quit",
	"a/s/b/t/i/k:zone  /:search  Enter:preview  e:edit  ?:help  q:quit",
	"a/n/e/b/r:filter  Enter:expand  ?:help  q:quit",
	"Enter:expand  Space:detail  r:rerun  ?:help  q:quit",
	"←/→:navigate  ?:help  q:quit",
}

// renderStatusBar produces the bottom status bar.
// info is shown on the right side (e.g. "3/15" cursor position).
func renderStatusBar(width int, activeTab TabID, info string) string {
	hints := tabHints[activeTab]

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#aaaaaa")).
		Width(width)

	left := " " + hints
	right := info + " "

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return style.Render(left + strings.Repeat(" ", gap) + right)
}

// cursorInfo returns a "cursor/total" string for list views, or "" if not applicable.
func (v DocsView) CursorInfo() string {
	if v.viewState != ViewReady || len(v.filtered) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", v.cursor+1, len(v.filtered))
}

// CursorInfo returns a "cursor/total" string for the iterations list.
func (v IterationsView) CursorInfo() string {
	if v.viewState != ViewReady || len(v.filtered) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", v.cursor+1, len(v.filtered))
}
