package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/domain"
)

// ZoneColors maps zones to their theme colors.
var ZoneColors = map[domain.Zone]lipgloss.Color{
	domain.ZoneSpec:       lipgloss.Color("#5f87ff"),
	domain.ZoneBlueprints: lipgloss.Color("#5fd7ff"),
	domain.ZoneState:      lipgloss.Color("#ffd700"),
	domain.ZoneIterations: lipgloss.Color("#5fd787"),
	domain.ZoneKnowledge:  lipgloss.Color("#d75fd7"),
}

// ZoneBar renders a single zone progress bar with label and fraction.
func ZoneBar(zh domain.ZoneHealth, width int) string {
	label := fmt.Sprintf("%-12s", string(zh.Zone)+"/")

	barWidth := 10
	if width >= 100 {
		barWidth = 20
	}
	fraction := fmt.Sprintf(" %d/%d", zh.Complete, zh.Total)

	color, ok := ZoneColors[zh.Zone]
	if !ok {
		color = lipgloss.Color("#888888")
	}

	filled := 0
	if zh.Total > 0 {
		filled = int(float64(zh.Complete) / float64(zh.Total) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
	}

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", barWidth-filled))

	return label + bar + fraction
}
