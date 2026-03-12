package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// QuickActions renders a key reference panel for the Status tab.
func QuickActions() string {
	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("#5f87ff"))

	b.WriteString(heading.Render("Quick Actions"))
	b.WriteString("\n")
	actions := []struct {
		k, desc string
	}{
		{"r", "refresh data"},
		{"2", "browse documents"},
		{"4", "run checks"},
		{"?", "help"},
		{"q", "quit"},
	}

	for _, a := range actions {
		b.WriteString("  " + key.Render(a.k) + "  " + a.desc + "\n")
	}

	return b.String()
}
