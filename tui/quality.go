package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/tui/components"
)

// QualityView is Tab 5: convergence score chart and analysis detail.
type QualityView struct {
	viewState     ViewState
	entries       []domain.QualityEntry
	selectedIndex int
	width         int
	height        int
}

// NewQualityView creates a QualityView.
func NewQualityView() QualityView {
	return QualityView{
		viewState: ViewLoading,
	}
}

func (v QualityView) Init() tea.Cmd { return nil }

func (v QualityView) Update(msg tea.Msg) (QualityView, tea.Cmd) {
	switch msg := msg.(type) {
	case qualityLoadedMsg:
		v.entries = msg.entries
		if len(v.entries) == 0 {
			v.viewState = ViewEmpty
		} else {
			v.viewState = ViewReady
			v.selectedIndex = len(v.entries) - 1
		}

	case qualityErrorMsg:
		v.viewState = ViewError

	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		return v.handleKey(msg)
	}
	return v, nil
}

func (v QualityView) handleKey(msg tea.KeyMsg) (QualityView, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if v.selectedIndex > 0 {
			v.selectedIndex--
		}
	case "right", "l":
		if v.selectedIndex < len(v.entries)-1 {
			v.selectedIndex++
		}
	}
	return v, nil
}

func (v QualityView) View() string {
	switch v.viewState {
	case ViewLoading:
		return components.EmptyState("Loading quality data...", v.width, v.height)
	case ViewError:
		return components.EmptyState("Error loading quality data.\nPress r to retry", v.width, v.height)
	case ViewEmpty:
		return components.EmptyState(
			"No quality data.\nRun a convergence analysis and then\n`mind quality log <file>` to start tracking.",
			v.width, v.height)
	}

	var b strings.Builder

	// Chart
	b.WriteString(v.renderChart())
	b.WriteString("\n")

	// Selected entry detail
	if v.selectedIndex >= 0 && v.selectedIndex < len(v.entries) {
		b.WriteString(v.renderDetail(v.entries[v.selectedIndex]))
	}

	return b.String()
}

func (v QualityView) renderChart() string {
	chartWidth := v.width - 10
	if chartWidth < 20 {
		chartWidth = 20
	}
	chartHeight := 8

	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)
	b.WriteString(heading.Render("  Score History") + "\n\n")

	// Y-axis: 5.0 to 1.0
	for row := chartHeight; row >= 0; row-- {
		yVal := 1.0 + float64(row)/float64(chartHeight)*4.0
		b.WriteString(fmt.Sprintf("  %3.1f │", yVal))

		for i, entry := range v.entries {
			// Map score to row position
			pos := int(math.Round((entry.Score - 1.0) / 4.0 * float64(chartHeight)))
			spacing := chartWidth / (len(v.entries) + 1)
			if spacing < 2 {
				spacing = 2
			}

			char := " "
			if pos == row {
				if i == v.selectedIndex {
					char = theme.Pass.Render("●")
				} else {
					char = "●"
				}
			}

			// Gate 0 threshold line at 3.0
			gateRow := int(math.Round((3.0 - 1.0) / 4.0 * float64(chartHeight)))
			if row == gateRow && char == " " {
				char = theme.Dim.Render("┄")
			}

			padding := strings.Repeat(" ", spacing-1)
			b.WriteString(padding + char)
		}
		b.WriteString("\n")
	}

	// X-axis
	b.WriteString("       └")
	b.WriteString(strings.Repeat("─", chartWidth))
	b.WriteString("\n")

	// Date labels
	b.WriteString("        ")
	spacing := chartWidth / (len(v.entries) + 1)
	if spacing < 2 {
		spacing = 2
	}
	for _, entry := range v.entries {
		label := entry.Date.Format("01/02")
		pad := spacing - len(label)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(strings.Repeat(" ", pad) + label)
	}
	b.WriteString("\n")

	// Gate 0 legend
	b.WriteString("  " + theme.Dim.Render("┄┄┄ Gate 0 (3.0)") + "\n")

	return b.String()
}

func (v QualityView) renderDetail(entry domain.QualityEntry) string {
	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)

	gateLabel := theme.Pass.Render("PASS")
	if !entry.GatePass {
		gateLabel = theme.Fail.Render("FAIL")
	}

	b.WriteString(heading.Render("  Selected Analysis") + "\n")
	b.WriteString(fmt.Sprintf("  Topic: %s", entry.Topic))
	if entry.Variant != "" {
		b.WriteString(fmt.Sprintf(" (%s)", entry.Variant))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Score: %.1f  Gate: %s  Date: %s\n",
		entry.Score, gateLabel, entry.Date.Format("2006-01-02")))

	// Dimension bars
	if len(entry.Dimensions) > 0 {
		b.WriteString("\n")
		for _, dim := range entry.Dimensions {
			barWidth := 10
			filled := dim.Value * barWidth / 5
			bar := theme.Pass.Render(strings.Repeat("█", filled)) +
				theme.Dim.Render(strings.Repeat("░", barWidth-filled))
			b.WriteString(fmt.Sprintf("  %-14s %s %d/5\n", dim.Name, bar, dim.Value))
		}
	}

	if len(entry.Personas) > 0 {
		b.WriteString(fmt.Sprintf("\n  Personas: %s\n", strings.Join(entry.Personas, ", ")))
	}
	if entry.OutputPath != "" {
		b.WriteString(fmt.Sprintf("  Output:   %s\n", entry.OutputPath))
	}

	return b.String()
}
