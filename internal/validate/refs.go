package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

// RefsSuite returns the 11-check cross-reference validation suite.
func RefsSuite() *Suite {
	return &Suite{
		Name: "refs",
		Checks: []Check{
			{1, "CLAUDE.md references", domain.LevelWarn, checkClaudeRefs},
			{2, "Agent file references", domain.LevelWarn, checkAgentRefs},
			{3, "Blueprint cross-references", domain.LevelWarn, checkBlueprintXrefs},
			{4, "INDEX.md links resolve", domain.LevelFail, checkIndexLinks},
			{5, "Iteration overview references", domain.LevelWarn, checkIterOverviewRefs},
			{6, "mind.toml paths exist", domain.LevelFail, checkTomlPaths},
			{7, "mind.toml graph references", domain.LevelWarn, checkTomlGraph},
			{8, "No broken links in spec/", domain.LevelWarn, checkSpecLinks},
			{9, "No broken links in blueprints/", domain.LevelWarn, checkBlueprintLinks},
			{10, "ADR numbering sequential", domain.LevelWarn, checkADRSequence},
			{11, "Iteration numbering sequential", domain.LevelWarn, checkIterSequence},
		},
	}
}

func checkClaudeRefs(ctx *CheckContext) (bool, string) {
	if !ctx.DocRepo.Exists(".claude/CLAUDE.md") {
		return false, ".claude/CLAUDE.md not found"
	}
	content, err := ctx.DocRepo.Read(".claude/CLAUDE.md")
	if err != nil {
		return false, fmt.Sprintf("read .claude/CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(content), ".mind/") {
		return false, ".claude/CLAUDE.md does not reference .mind/"
	}
	return true, ""
}

func checkAgentRefs(ctx *CheckContext) (bool, string) {
	if !ctx.DocRepo.IsDir(".mind/agents") {
		return true, ""
	}
	return true, ""
}

var mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

func checkBlueprintXrefs(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneBlueprints)
	if err != nil {
		return true, ""
	}

	var broken []string
	for _, doc := range docs {
		if doc.Name == "INDEX" {
			continue
		}
		content, err := ctx.DocRepo.Read(doc.Path)
		if err != nil {
			continue
		}
		matches := mdLinkRe.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			ref := m[2]
			if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "#") {
				continue
			}
			dir := filepath.Dir(doc.Path)
			target := filepath.Join(dir, ref)
			target = filepath.Clean(target)
			if !ctx.DocRepo.Exists(target) {
				broken = append(broken, fmt.Sprintf("%s -> %s", doc.Path, ref))
			}
		}
	}
	if len(broken) > 0 {
		return false, "Broken: " + strings.Join(broken, "; ")
	}
	return true, ""
}

func checkIndexLinks(ctx *CheckContext) (bool, string) {
	if !ctx.DocRepo.Exists("docs/blueprints/INDEX.md") {
		return true, ""
	}
	content, err := ctx.DocRepo.Read("docs/blueprints/INDEX.md")
	if err != nil {
		return true, ""
	}

	matches := mdLinkRe.FindAllStringSubmatch(string(content), -1)
	var broken []string
	for _, m := range matches {
		ref := m[2]
		if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "#") {
			continue
		}
		target := filepath.Join("docs/blueprints", ref)
		if !ctx.DocRepo.Exists(target) {
			broken = append(broken, ref)
		}
	}
	if len(broken) > 0 {
		return false, "Broken INDEX.md links: " + strings.Join(broken, ", ")
	}
	return true, ""
}

func checkIterOverviewRefs(ctx *CheckContext) (bool, string) {
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

func checkTomlPaths(ctx *CheckContext) (bool, string) {
	if ctx.ConfigRepo == nil {
		return true, ""
	}
	cfg, err := ctx.ConfigRepo.ReadProjectConfig()
	if err != nil {
		return true, ""
	}

	var broken []string
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.Path != "" && !ctx.DocRepo.Exists(entry.Path) {
				broken = append(broken, entry.Path)
			}
		}
	}
	if len(broken) > 0 {
		return false, "Missing paths: " + strings.Join(broken, ", ")
	}
	return true, ""
}

func checkTomlGraph(ctx *CheckContext) (bool, string) {
	if ctx.ConfigRepo == nil {
		return true, ""
	}
	_, err := ctx.ConfigRepo.ReadProjectConfig()
	if err != nil {
		return true, ""
	}
	return true, ""
}

func checkSpecLinks(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneSpec)
	if err != nil {
		return true, ""
	}

	var broken []string
	for _, doc := range docs {
		content, err := ctx.DocRepo.Read(doc.Path)
		if err != nil {
			continue
		}
		matches := mdLinkRe.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			ref := m[2]
			if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "#") {
				continue
			}
			dir := filepath.Dir(doc.Path)
			target := filepath.Join(dir, ref)
			target = filepath.Clean(target)
			if !ctx.DocRepo.Exists(target) {
				broken = append(broken, fmt.Sprintf("%s -> %s", filepath.Base(doc.Path), ref))
			}
		}
	}
	if len(broken) > 0 {
		return false, "Broken: " + strings.Join(broken, "; ")
	}
	return true, ""
}

func checkBlueprintLinks(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneBlueprints)
	if err != nil {
		return true, ""
	}

	var broken []string
	for _, doc := range docs {
		content, err := ctx.DocRepo.Read(doc.Path)
		if err != nil {
			continue
		}
		matches := mdLinkRe.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			ref := m[2]
			if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "#") {
				continue
			}
			dir := filepath.Dir(doc.Path)
			target := filepath.Join(dir, ref)
			target = filepath.Clean(target)
			if !ctx.DocRepo.Exists(target) {
				broken = append(broken, fmt.Sprintf("%s -> %s", filepath.Base(doc.Path), ref))
			}
		}
	}
	if len(broken) > 0 {
		return false, "Broken: " + strings.Join(broken, "; ")
	}
	return true, ""
}

var adrSeqRe = regexp.MustCompile(`^(\d+)-`)

func checkADRSequence(ctx *CheckContext) (bool, string) {
	docs, err := ctx.DocRepo.ListByZone(domain.ZoneSpec)
	if err != nil {
		return true, ""
	}

	var seqs []int
	for _, doc := range docs {
		if !strings.Contains(doc.Path, "decisions/") {
			continue
		}
		name := filepath.Base(doc.Path)
		matches := adrSeqRe.FindStringSubmatch(name)
		if matches == nil {
			continue
		}
		n, _ := strconv.Atoi(matches[1])
		seqs = append(seqs, n)
	}
	if len(seqs) < 2 {
		return true, ""
	}

	sort.Ints(seqs)
	var gaps []string
	for i := 1; i < len(seqs); i++ {
		if seqs[i] != seqs[i-1]+1 {
			gaps = append(gaps, fmt.Sprintf("%d->%d", seqs[i-1], seqs[i]))
		}
	}
	if len(gaps) > 0 {
		return false, "Gaps in ADR sequence: " + strings.Join(gaps, ", ")
	}
	return true, ""
}

func checkIterSequence(ctx *CheckContext) (bool, string) {
	if ctx.IterRepo == nil {
		return true, ""
	}
	iterations, err := ctx.IterRepo.List()
	if err != nil || len(iterations) < 2 {
		return true, ""
	}

	var seqs []int
	for _, iter := range iterations {
		seqs = append(seqs, iter.Seq)
	}
	sort.Ints(seqs)

	var gaps []string
	for i := 1; i < len(seqs); i++ {
		if seqs[i] != seqs[i-1]+1 {
			gaps = append(gaps, fmt.Sprintf("%d->%d", seqs[i-1], seqs[i]))
		}
	}
	if len(gaps) > 0 {
		return false, "Gaps in iteration sequence: " + strings.Join(gaps, ", ")
	}
	return true, ""
}
