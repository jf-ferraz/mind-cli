package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"golang.org/x/term"
)

// OutputMode determines how results are displayed.
type OutputMode int

const (
	ModeInteractive OutputMode = iota
	ModePlain
	ModeJSON
)

// DetectMode checks TTY and flags.
func DetectMode(jsonFlag bool, noColorFlag bool) OutputMode {
	if jsonFlag {
		return ModeJSON
	}
	if noColorFlag || !term.IsTerminal(int(os.Stdout.Fd())) {
		return ModePlain
	}
	return ModeInteractive
}

// TermWidth returns the terminal width or 80 as default.
func TermWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// Renderer formats domain results for display.
type Renderer struct {
	mode  OutputMode
	width int
}

// New creates a Renderer.
func New(mode OutputMode, width int) *Renderer {
	return &Renderer{mode: mode, width: width}
}

// RenderHealth formats a ProjectHealth for display.
func (r *Renderer) RenderHealth(h *domain.ProjectHealth) string {
	if r.mode == ModeJSON {
		return jsonMarshal(h)
	}
	return r.renderHealthText(h)
}

// RenderValidation formats a ValidationReport for display.
func (r *Renderer) RenderValidation(report *domain.ValidationReport) string {
	if r.mode == ModeJSON {
		return jsonMarshal(report)
	}
	return r.renderValidationText(report)
}

// RenderBrief formats a Brief for display.
func (r *Renderer) RenderBrief(b *domain.Brief) string {
	if r.mode == ModeJSON {
		return jsonMarshal(b)
	}
	return r.renderBriefText(b)
}

// RenderIterations formats an iteration list for display.
func (r *Renderer) RenderIterations(iters []domain.Iteration) string {
	if r.mode == ModeJSON {
		return jsonMarshal(iters)
	}
	return r.renderIterationsText(iters)
}

func (r *Renderer) renderHealthText(h *domain.ProjectHealth) string {
	var b strings.Builder

	name := h.Project.Name
	if name == "" {
		name = "(unnamed)"
	}

	fmt.Fprintf(&b, "Project: %s\n", name)
	fmt.Fprintf(&b, "Root:    %s\n\n", h.Project.Root)

	fmt.Fprintln(&b, "Documentation Health")
	fmt.Fprintln(&b, strings.Repeat("─", 50))

	for _, zone := range domain.AllZones {
		zh, ok := h.Zones[zone]
		if !ok {
			continue
		}
		bar := progressBar(zh.Complete, zh.Total, 20)
		fmt.Fprintf(&b, "  %-14s %s  %d/%d", string(zone)+"/", bar, zh.Complete, zh.Total)
		if zh.Stubs > 0 {
			fmt.Fprintf(&b, "  (%d stubs)", zh.Stubs)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b)
	if h.Workflow == nil || h.Workflow.IsIdle() {
		fmt.Fprintln(&b, "Workflow: idle")
	} else {
		fmt.Fprintf(&b, "Workflow: %s (%s)\n", h.Workflow.Type, h.Workflow.Descriptor)
		fmt.Fprintf(&b, "  Last agent: %s\n", h.Workflow.LastAgent)
		if len(h.Workflow.RemainingChain) > 0 {
			fmt.Fprintf(&b, "  Remaining:  %s\n", strings.Join(h.Workflow.RemainingChain, " → "))
		}
	}

	if h.LastIteration != nil {
		fmt.Fprintf(&b, "\nLast iteration: %s (%s)\n", h.LastIteration.DirName, h.LastIteration.Status)
	}

	if len(h.Warnings) > 0 {
		fmt.Fprintln(&b, "\nWarnings")
		fmt.Fprintln(&b, strings.Repeat("─", 50))
		for _, w := range h.Warnings {
			fmt.Fprintf(&b, "  ⚠ %s\n", w)
		}
	}

	return b.String()
}

func (r *Renderer) renderValidationText(report *domain.ValidationReport) string {
	var b strings.Builder

	title := report.Suite
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}
	fmt.Fprintf(&b, "=== %s Validation ===\n\n", title)

	for _, check := range report.Checks {
		icon := "✓"
		if !check.Passed {
			if check.Level == domain.LevelFail {
				icon = "✗"
			} else {
				icon = "⚠"
			}
		}
		fmt.Fprintf(&b, "[%d/%d] %s %s\n", check.ID, report.Total, icon, check.Name)
		if !check.Passed && check.Message != "" {
			fmt.Fprintf(&b, "       → %s\n", check.Message)
		}
	}

	fmt.Fprintf(&b, "\nPass: %d | Fail: %d | Warn: %d\n", report.Passed, report.Failed, report.Warnings)
	if report.Ok() {
		fmt.Fprintln(&b, "✓ All checks passed.")
	} else {
		fmt.Fprintf(&b, "✗ %d check(s) failed.\n", report.Failed)
	}

	return b.String()
}

func (r *Renderer) renderBriefText(b *domain.Brief) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Brief: %s\n", b.Path)
	fmt.Fprintf(&sb, "  Exists:       %v\n", b.Exists)
	fmt.Fprintf(&sb, "  Is stub:      %v\n", b.IsStub)
	fmt.Fprintf(&sb, "  Vision:       %v\n", b.HasVision)
	fmt.Fprintf(&sb, "  Deliverables: %v\n", b.HasDeliverables)
	fmt.Fprintf(&sb, "  Scope:        %v\n", b.HasScope)
	fmt.Fprintf(&sb, "  Gate result:  %s\n", b.GateResult)

	return sb.String()
}

func (r *Renderer) renderIterationsText(iters []domain.Iteration) string {
	if len(iters) == 0 {
		return "No iterations found.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-5s %-14s %-30s %-12s %s\n", "#", "Type", "Name", "Status", "Files")
	fmt.Fprintln(&b, strings.Repeat("─", 75))

	for _, iter := range iters {
		present := 0
		for _, a := range iter.Artifacts {
			if a.Exists {
				present++
			}
		}
		fmt.Fprintf(&b, "%03d   %-14s %-30s %-12s %d/%d\n",
			iter.Seq,
			string(iter.Type),
			iter.Descriptor,
			string(iter.Status),
			present,
			len(domain.ExpectedArtifacts),
		)
	}

	return b.String()
}

func progressBar(complete, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(float64(complete) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func jsonMarshal(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data) + "\n"
}
