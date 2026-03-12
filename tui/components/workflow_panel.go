package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jf-ferraz/mind-cli/domain"
)

// WorkflowPanel renders the workflow state section.
func WorkflowPanel(ws *domain.WorkflowState, lastIter *domain.Iteration) string {
	var b strings.Builder
	heading := lipgloss.NewStyle().Bold(true)

	b.WriteString(heading.Render("Workflow"))
	b.WriteString("\n")

	if ws == nil || ws.IsIdle() {
		b.WriteString("  State: idle\n")
		if lastIter != nil {
			b.WriteString(fmt.Sprintf("  Last:  %s (%s)\n", lastIter.DirName, lastIter.Status))
		}
	} else {
		active := lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787"))
		b.WriteString("  State: " + active.Render("running") + "\n")
		b.WriteString(fmt.Sprintf("  Type:  %s\n", ws.Type))
		b.WriteString(fmt.Sprintf("  Agent: %s\n", ws.LastAgent))
		if ws.Branch != "" {
			b.WriteString(fmt.Sprintf("  Branch: %s\n", ws.Branch))
		}
		if len(ws.RemainingChain) > 0 {
			b.WriteString(fmt.Sprintf("  Next:  %s\n", strings.Join(ws.RemainingChain, " → ")))
		}
	}

	return b.String()
}
