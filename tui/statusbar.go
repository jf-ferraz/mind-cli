package tui

import (
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
