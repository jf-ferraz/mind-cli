package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

// DocsSuite returns the 17-check documentation validation suite.
func DocsSuite() *Suite {
	return &Suite{
		Name: "docs",
		Checks: []Check{
			{1, "docs/ directory exists", domain.LevelFail, checkDocsDir},
			{2, "All 5 zone directories exist", domain.LevelFail, checkZoneDirs},
			{3, "Required spec files", domain.LevelFail, checkSpecFiles},
			{4, "decisions/ subdirectory", domain.LevelWarn, checkDecisionsDir},
			{5, "ADR naming convention", domain.LevelWarn, checkADRNaming},
			{6, "blueprints/INDEX.md", domain.LevelFail, checkBlueprintsIndex},
			{7, "Blueprint → INDEX.md coverage", domain.LevelWarn, checkBlueprintCoverage},
			{8, "INDEX.md → file references", domain.LevelFail, checkIndexRefs},
			{9, "state/current.md", domain.LevelFail, checkCurrentState},
			{10, "state/workflow.md", domain.LevelWarn, checkWorkflowState},
			{11, "knowledge/glossary.md", domain.LevelWarn, checkGlossary},
			{12, "Iteration folder naming", domain.LevelWarn, checkIterationNaming},
			{13, "Iterations have overview.md", domain.LevelWarn, checkIterationOverview},
			{14, "Spike file naming", domain.LevelWarn, checkSpikeNaming},
			{15, "No legacy paths", domain.LevelFail, checkNoLegacyPaths},
			{16, "Stub detection", domain.LevelWarn, checkStubs},
			{17, "Project brief completeness", domain.LevelWarn, checkBriefCompleteness},
		},
	}
}

func checkDocsDir(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.IsDir("docs") {
		return true, ""
	}
	return false, "docs/ directory not found"
}

func checkZoneDirs(ctx *CheckContext) (bool, string) {
	var missing []string
	for _, zone := range domain.AllZones {
		if !ctx.DocRepo.IsDir(filepath.Join("docs", string(zone))) {
			missing = append(missing, string(zone))
		}
	}
	if len(missing) > 0 {
		return false, "Missing zones: " + strings.Join(missing, ", ")
	}
	return true, ""
}

func checkSpecFiles(ctx *CheckContext) (bool, string) {
	required := []string{
		"docs/spec/project-brief.md",
		"docs/spec/requirements.md",
		"docs/spec/architecture.md",
	}
	var missing []string
	for _, f := range required {
		if !ctx.DocRepo.Exists(f) {
			missing = append(missing, filepath.Base(f))
		}
	}
	if len(missing) > 0 {
		return false, "Missing: " + strings.Join(missing, ", ")
	}
	return true, ""
}

func checkDecisionsDir(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.IsDir("docs/spec/decisions") {
		return true, ""
	}
	return false, "docs/spec/decisions/ not found"
}

var adrNameRe = regexp.MustCompile(`^\d+-[a-z0-9].*\.md$`)

func checkADRNaming(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneSpec)
	if err != nil {
		return true, "" // Can't check, pass silently
	}

	var bad []string
	for _, doc := range docs {
		// Only check files in decisions/
		if !strings.Contains(doc.Path, "decisions/") {
			continue
		}
		name := filepath.Base(doc.Path)
		if name == "_template.md" {
			continue
		}
		if !adrNameRe.MatchString(name) {
			bad = append(bad, name)
		}
	}
	if len(bad) > 0 {
		return false, "Bad ADR names: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkBlueprintsIndex(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.Exists("docs/blueprints/INDEX.md") {
		return true, ""
	}
	return false, "docs/blueprints/INDEX.md not found"
}

var blueprintNameRe = regexp.MustCompile(`^\d+.*\.md$`)

func checkBlueprintCoverage(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneBlueprints)
	if err != nil {
		return true, ""
	}

	indexContent, err := ctx.DocRepo.Read("docs/blueprints/INDEX.md")
	if err != nil {
		return true, ""
	}

	var uncovered []string
	for _, doc := range docs {
		name := filepath.Base(doc.Path)
		if name == "INDEX.md" {
			continue
		}
		if !blueprintNameRe.MatchString(name) {
			continue
		}
		if !strings.Contains(string(indexContent), name) {
			uncovered = append(uncovered, name)
		}
	}
	if len(uncovered) > 0 {
		return false, "Not in INDEX.md: " + strings.Join(uncovered, ", ")
	}
	return true, ""
}

var linkRefRe = regexp.MustCompile(`\((\d[\da-z-]*\.md)\)`)

func checkIndexRefs(ctx *CheckContext) (bool, string) {
	content, err := ctx.DocRepo.Read("docs/blueprints/INDEX.md")
	if err != nil {
		return true, ""
	}

	matches := linkRefRe.FindAllStringSubmatch(string(content), -1)
	var bad []string
	for _, m := range matches {
		ref := m[1]
		if !ctx.DocRepo.Exists(filepath.Join("docs/blueprints", ref)) {
			bad = append(bad, ref)
		}
	}
	if len(bad) > 0 {
		return false, strings.Join(bad, ", ")
	}
	return true, ""
}

func checkCurrentState(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.Exists("docs/state/current.md") {
		return true, ""
	}
	return false, "docs/state/current.md not found"
}

func checkWorkflowState(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.Exists("docs/state/workflow.md") {
		return true, ""
	}
	return false, "docs/state/workflow.md not found"
}

func checkGlossary(ctx *CheckContext) (bool, string) {
	if ctx.DocRepo.Exists("docs/knowledge/glossary.md") {
		return true, ""
	}
	return false, "docs/knowledge/glossary.md not found"
}

var iterNameRe = regexp.MustCompile(`^\d{3}-[A-Z_]+-[a-z0-9]`)

func checkIterationNaming(ctx *CheckContext) (bool, string) {
	if ctx.IterRepo == nil {
		return true, ""
	}
	iterations, err := ctx.IterRepo.List()
	if err != nil || len(iterations) == 0 {
		return true, ""
	}

	var bad []string
	for _, iter := range iterations {
		if !iterNameRe.MatchString(iter.DirName) {
			bad = append(bad, iter.DirName)
		}
	}
	if len(bad) > 0 {
		return false, "Bad names: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkIterationOverview(ctx *CheckContext) (bool, string) {
	if ctx.IterRepo == nil {
		return true, ""
	}
	iterations, err := ctx.IterRepo.List()
	if err != nil || len(iterations) == 0 {
		return true, ""
	}

	var missing []string
	for _, iter := range iterations {
		hasOverview := false
		for _, a := range iter.Artifacts {
			if a.Name == "overview.md" && a.Exists {
				hasOverview = true
				break
			}
		}
		if !hasOverview {
			missing = append(missing, iter.DirName)
		}
	}
	if len(missing) > 0 {
		return false, "Missing overview.md: " + strings.Join(missing, ", ")
	}
	return true, ""
}

func checkSpikeNaming(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneKnowledge)
	if err != nil {
		return true, ""
	}

	var bad []string
	for _, doc := range docs {
		name := filepath.Base(doc.Path)
		if name == "glossary.md" || name == "quality-log.yml" {
			continue
		}
		if strings.HasSuffix(name, "-convergence.md") {
			continue
		}
		// If it looks like a spike but doesn't have the suffix
		if strings.Contains(name, "spike") && !strings.HasSuffix(name, "-spike.md") {
			bad = append(bad, name)
		}
	}
	if len(bad) > 0 {
		return false, "Bad spike names: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkNoLegacyPaths(ctx *CheckContext) (bool, string) {
	legacy := []string{
		"docs/adr", "docs/adrs", "docs/spikes",
		"docs/architecture", "docs/current",
	}
	var found []string
	for _, p := range legacy {
		if ctx.DocRepo.IsDir(p) {
			found = append(found, p)
		}
	}
	if len(found) > 0 {
		return false, "Legacy paths: " + strings.Join(found, ", ")
	}
	return true, ""
}

func checkStubs(ctx *CheckContext) (bool, string) {
	keyDocs := []string{
		"docs/spec/project-brief.md",
		"docs/spec/requirements.md",
		"docs/spec/architecture.md",
	}

	var stubs []string
	for _, path := range keyDocs {
		if !ctx.DocRepo.Exists(path) {
			continue
		}
		isStub, err := ctx.DocRepo.IsStub(path)
		if err != nil {
			continue
		}
		if isStub {
			stubs = append(stubs, path)
		}
	}
	if len(stubs) > 0 {
		return false, strings.Join(stubs, " ")
	}
	return true, ""
}

func checkBriefCompleteness(ctx *CheckContext) (bool, string) {
	if ctx.BriefRepo == nil {
		return true, ""
	}
	brief, err := ctx.BriefRepo.ParseBrief()
	if err != nil || !brief.Exists || brief.IsStub {
		return true, "" // Caught by other checks
	}

	var missing []string
	if !brief.HasVision {
		missing = append(missing, "Vision")
	}
	if !brief.HasDeliverables {
		missing = append(missing, "Key Deliverables")
	}
	if !brief.HasScope {
		missing = append(missing, "Scope")
	}
	if len(missing) > 0 {
		msg := fmt.Sprintf("Missing sections: %s", strings.Join(missing, ", "))
		return false, msg
	}
	return true, ""
}
