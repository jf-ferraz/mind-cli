package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// EmptyState renders a centered message for tabs with no data.
func EmptyState(msg string, width, height int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(width).
		Align(lipgloss.Center)

	// Vertical centering
	pad := (height - 3) / 2
	if pad < 0 {
		pad = 0
	}

	return strings.Repeat("\n", pad) + style.Render(msg)
}
