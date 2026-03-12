package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Suggestions renders a list of suggestion messages.
func Suggestions(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd7ff"))

	b.WriteString(heading.Render("Suggestions"))
	b.WriteString("\n")
	for _, s := range suggestions {
		b.WriteString(hint.Render("  → ") + s + "\n")
	}

	return b.String()
}
