package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Warnings renders a list of warning messages.
func Warnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}

	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700"))

	b.WriteString(heading.Render("Warnings"))
	b.WriteString("\n")
	for _, w := range warnings {
		b.WriteString(warn.Render("  ⚠ ") + w + "\n")
	}

	return b.String()
}
