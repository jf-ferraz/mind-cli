package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Staleness renders a list of stale documents with bullet markers.
func Staleness(stale map[string]string, width int) string {
	if len(stale) == 0 {
		return ""
	}

	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700"))

	b.WriteString(heading.Render("Staleness"))
	b.WriteString(fmt.Sprintf(" (%d stale)\n", len(stale)))

	// Sort for deterministic output
	ids := make([]string, 0, len(stale))
	for id := range stale {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		b.WriteString(warn.Render("  ● ") + id + "\n")
	}

	return b.String()
}
