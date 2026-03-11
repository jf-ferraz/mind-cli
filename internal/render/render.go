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

// RenderDoctor formats a DoctorReport for display.
func (r *Renderer) RenderDoctor(report *domain.DoctorReport) string {
	if r.mode == ModeJSON {
		return jsonMarshal(report)
	}
	return r.renderDoctorText(report)
}

// RenderInitResult formats an InitResult for display.
func (r *Renderer) RenderInitResult(result *domain.InitResult) string {
	if r.mode == ModeJSON {
		return jsonMarshal(result)
	}
	return r.renderInitResultText(result)
}

// RenderCreateResult formats a CreateResult for display.
func (r *Renderer) RenderCreateResult(result *domain.CreateResult) string {
	if r.mode == ModeJSON {
		return jsonMarshal(result)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Created: %s\n", result.Path)
	if result.IndexUpdated {
		fmt.Fprintln(&b, "INDEX.md updated")
	}
	return b.String()
}

// RenderCreateIterationResult formats a CreateIterationResult for display.
func (r *Renderer) RenderCreateIterationResult(result *domain.CreateIterationResult) string {
	if r.mode == ModeJSON {
		return jsonMarshal(result)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Created iteration: %s\n", result.Path)
	fmt.Fprintf(&b, "  Type: %s\n", result.Type)
	fmt.Fprintf(&b, "  Files: %s\n", strings.Join(result.Files, ", "))
	return b.String()
}

// RenderDocumentList formats a DocumentList for display.
func (r *Renderer) RenderDocumentList(list *domain.DocumentList) string {
	if r.mode == ModeJSON {
		return jsonMarshal(list)
	}
	return r.renderDocumentListText(list)
}

// RenderDocTree formats a docs tree for display.
func (r *Renderer) RenderDocTree(docs []domain.Document) string {
	if r.mode == ModeJSON {
		return jsonMarshal(docs)
	}
	return r.renderDocTreeText(docs)
}

// RenderStubList formats a StubList for display.
func (r *Renderer) RenderStubList(list *domain.StubList) string {
	if r.mode == ModeJSON {
		return jsonMarshal(list)
	}
	return r.renderStubListText(list)
}

// RenderSearchResults formats SearchResults for display.
func (r *Renderer) RenderSearchResults(results *domain.SearchResults) string {
	if r.mode == ModeJSON {
		return jsonMarshal(results)
	}
	return r.renderSearchResultsText(results)
}

// RenderUnifiedValidation formats a UnifiedValidationReport for display.
func (r *Renderer) RenderUnifiedValidation(report *domain.UnifiedValidationReport) string {
	if r.mode == ModeJSON {
		return jsonMarshal(report)
	}
	return r.renderUnifiedValidationText(report)
}

// RenderWorkflowStatus formats a WorkflowState for display.
func (r *Renderer) RenderWorkflowStatus(ws *domain.WorkflowState) string {
	if r.mode == ModeJSON {
		if ws == nil {
			return jsonMarshal(map[string]string{"state": "idle"})
		}
		return jsonMarshal(ws)
	}
	return r.renderWorkflowStatusText(ws)
}

// RenderWorkflowHistory formats a WorkflowHistory for display.
func (r *Renderer) RenderWorkflowHistory(history *domain.WorkflowHistory) string {
	if r.mode == ModeJSON {
		return jsonMarshal(history)
	}
	return r.renderWorkflowHistoryText(history)
}

// RenderVersionInfo formats a VersionInfo for display.
func (r *Renderer) RenderVersionInfo(info *domain.VersionInfo) string {
	if r.mode == ModeJSON {
		return jsonMarshal(info)
	}
	return fmt.Sprintf("mind %s (%s) built %s %s/%s\n", info.Version, info.Commit, info.BuildDate, info.OS, info.Arch)
}

func (r *Renderer) renderDoctorText(report *domain.DoctorReport) string {
	var b strings.Builder

	fmt.Fprintln(&b, "=== Doctor Report ===\n")

	for _, d := range report.Diagnostics {
		icon := "✓"
		switch d.Status {
		case "fail":
			icon = "✗"
		case "warn":
			icon = "⚠"
		}
		fmt.Fprintf(&b, "%s [%s] %s: %s\n", icon, d.Category, d.Check, d.Message)
		if d.Fix != "" && d.Status != "pass" {
			fmt.Fprintf(&b, "  Fix: %s\n", d.Fix)
		}
	}

	fmt.Fprintf(&b, "\nPass: %d | Fail: %d | Warn: %d\n", report.Summary.Pass, report.Summary.Fail, report.Summary.Warn)

	if len(report.FixesApplied) > 0 {
		fmt.Fprintln(&b, "\nFixes Applied:")
		for _, fix := range report.FixesApplied {
			fmt.Fprintf(&b, "  ✓ %s\n", fix)
		}
	}

	return b.String()
}

func (r *Renderer) renderInitResultText(result *domain.InitResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Initialized Mind project: %s\n", result.ProjectName)
	fmt.Fprintf(&b, "Root: %s\n\n", result.Root)

	if len(result.FilesCreated) > 0 {
		fmt.Fprintln(&b, "Created:")
		for _, f := range result.FilesCreated {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}

	if len(result.ExistingPreserved) > 0 {
		fmt.Fprintln(&b, "\nPreserved (existing):")
		for _, f := range result.ExistingPreserved {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}

	return b.String()
}

func (r *Renderer) renderDocumentListText(list *domain.DocumentList) string {
	var b strings.Builder

	currentZone := domain.Zone("")
	for _, doc := range list.Documents {
		if doc.Zone != currentZone {
			if currentZone != "" {
				fmt.Fprintln(&b)
			}
			currentZone = doc.Zone
			fmt.Fprintf(&b, "%s/\n", string(doc.Zone))
			fmt.Fprintln(&b, strings.Repeat("─", 40))
		}
		stub := ""
		if doc.IsStub {
			stub = " [stub]"
		}
		fmt.Fprintf(&b, "  %-40s %6d bytes  %s%s\n",
			doc.Path,
			doc.Size,
			doc.ModTime.Format("2006-01-02"),
			stub)
	}

	fmt.Fprintf(&b, "\nTotal: %d documents\n", list.Total)
	return b.String()
}

func (r *Renderer) renderDocTreeText(docs []domain.Document) string {
	var b strings.Builder

	fmt.Fprintln(&b, "docs/")

	// Group by zone
	byZone := make(map[domain.Zone][]domain.Document)
	for _, doc := range docs {
		byZone[doc.Zone] = append(byZone[doc.Zone], doc)
	}

	for i, zone := range domain.AllZones {
		zoneDocs := byZone[zone]
		isLastZone := i == len(domain.AllZones)-1
		prefix := "├── "
		childPrefix := "│   "
		if isLastZone {
			prefix = "└── "
			childPrefix = "    "
		}
		fmt.Fprintf(&b, "%s%s/\n", prefix, string(zone))

		for j, doc := range zoneDocs {
			isLast := j == len(zoneDocs)-1
			filePrefix := childPrefix + "├── "
			if isLast {
				filePrefix = childPrefix + "└── "
			}
			stub := ""
			if doc.IsStub {
				stub = " [stub]"
			}
			name := doc.Name + ".md"
			fmt.Fprintf(&b, "%s%s%s\n", filePrefix, name, stub)
		}
	}

	return b.String()
}

func (r *Renderer) renderStubListText(list *domain.StubList) string {
	if list.Count == 0 {
		return "No stubs found.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d stub(s):\n\n", list.Count)

	for _, stub := range list.Stubs {
		fmt.Fprintf(&b, "  [%s] %s\n", stub.Zone, stub.Path)
		if stub.Hint != "" {
			fmt.Fprintf(&b, "    Hint: %s\n", stub.Hint)
		}
	}

	return b.String()
}

func (r *Renderer) renderSearchResultsText(results *domain.SearchResults) string {
	if results.TotalMatches == 0 {
		return fmt.Sprintf("No matches found for %q.\n", results.Query)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d match(es) in %d file(s) for %q:\n\n", results.TotalMatches, results.FilesMatched, results.Query)

	for _, file := range results.Results {
		fmt.Fprintf(&b, "%s:\n", file.Path)
		for _, match := range file.Matches {
			if match.ContextBefore != "" {
				fmt.Fprintf(&b, "  %d: %s\n", match.Line-1, match.ContextBefore)
			}
			fmt.Fprintf(&b, "  %d: %s\n", match.Line, match.Text)
			if match.ContextAfter != "" {
				fmt.Fprintf(&b, "  %d: %s\n", match.Line+1, match.ContextAfter)
			}
			fmt.Fprintln(&b)
		}
	}

	return b.String()
}

func (r *Renderer) renderUnifiedValidationText(report *domain.UnifiedValidationReport) string {
	var b strings.Builder

	for _, suite := range report.Suites {
		b.WriteString(r.renderValidationText(&suite))
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "=== Summary ===")
	fmt.Fprintf(&b, "Total: %d | Pass: %d | Fail: %d | Warn: %d\n",
		report.Summary.Total, report.Summary.Passed, report.Summary.Failed, report.Summary.Warnings)

	if report.Summary.Failed == 0 {
		fmt.Fprintln(&b, "✓ All suites passed.")
	} else {
		fmt.Fprintf(&b, "✗ %d check(s) failed across all suites.\n", report.Summary.Failed)
	}

	return b.String()
}

func (r *Renderer) renderWorkflowStatusText(ws *domain.WorkflowState) string {
	if ws == nil || ws.IsIdle() {
		return "Workflow: idle\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Workflow: %s\n", ws.Type)
	fmt.Fprintf(&b, "  Descriptor: %s\n", ws.Descriptor)
	fmt.Fprintf(&b, "  Last agent: %s\n", ws.LastAgent)
	if len(ws.RemainingChain) > 0 {
		fmt.Fprintf(&b, "  Remaining:  %s\n", strings.Join(ws.RemainingChain, " → "))
	}
	if ws.Session > 0 {
		fmt.Fprintf(&b, "  Session:    %d/%d\n", ws.Session, ws.TotalSessions)
	}
	return b.String()
}

func (r *Renderer) renderWorkflowHistoryText(history *domain.WorkflowHistory) string {
	if history.Total == 0 {
		return "No iterations found.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-5s %-14s %-30s %-12s %s\n", "#", "Type", "Descriptor", "Status", "Artifacts")
	fmt.Fprintln(&b, strings.Repeat("─", 75))

	for _, iter := range history.Iterations {
		fmt.Fprintf(&b, "%03d   %-14s %-30s %-12s %d/%d\n",
			iter.Seq,
			string(iter.Type),
			iter.Descriptor,
			string(iter.Status),
			iter.Artifacts.Present,
			iter.Artifacts.Expected,
		)
	}

	fmt.Fprintf(&b, "\nTotal: %d iteration(s)\n", history.Total)
	return b.String()
}

func jsonMarshal(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data) + "\n"
}
