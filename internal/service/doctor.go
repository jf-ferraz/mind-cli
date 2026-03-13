package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/generate"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// DoctorService runs full project diagnostics.
type DoctorService struct {
	projectRoot string
	docRepo     repo.DocRepo
	iterRepo    repo.IterationRepo
	briefRepo   repo.BriefRepo
	configRepo  repo.ConfigRepo
	lockRepo    repo.LockRepo
}

// NewDoctorService creates a DoctorService.
func NewDoctorService(
	projectRoot string,
	docRepo repo.DocRepo,
	iterRepo repo.IterationRepo,
	briefRepo repo.BriefRepo,
	configRepo repo.ConfigRepo,
	lockRepo repo.LockRepo,
) *DoctorService {
	return &DoctorService{
		projectRoot: projectRoot,
		docRepo:     docRepo,
		iterRepo:    iterRepo,
		briefRepo:   briefRepo,
		configRepo:  configRepo,
		lockRepo:    lockRepo,
	}
}

// Run executes all diagnostic checks and returns a report.
func (s *DoctorService) Run(fix bool) *domain.DoctorReport {
	report := &domain.DoctorReport{}

	// Framework checks
	s.checkFramework(report)

	// Adapter checks
	s.checkAdapters(report)

	// Doc structure checks
	s.checkDocStructure(report)

	// Brief check
	s.checkBrief(report)

	// Config check
	s.checkConfig(report)

	// Workflow check
	s.checkWorkflow(report)

	// Iteration checks
	s.checkIterations(report)

	// Staleness checks (FR-81)
	s.checkStaleness(report)

	// MCP and claude CLI checks (G6)
	s.checkMCP(report)
	s.checkClaudeCLI(report)

	// Count summary
	for _, d := range report.Diagnostics {
		switch d.Status {
		case domain.DiagPass:
			report.Summary.Pass++
		case domain.DiagFail:
			report.Summary.Fail++
		case domain.DiagWarn:
			report.Summary.Warn++
		}
	}

	// Auto-fix if requested
	if fix {
		s.applyFixes(report)
	}

	return report
}

func (s *DoctorService) addDiag(report *domain.DoctorReport, category, check string, status domain.DiagnosticStatus, message, fixHint string, autoFixable bool) {
	level := domain.LevelInfo
	switch status {
	case domain.DiagFail:
		level = domain.LevelFail
	case domain.DiagWarn:
		level = domain.LevelWarn
	}

	report.Diagnostics = append(report.Diagnostics, domain.Diagnostic{
		Category: category,
		Check:    check,
		Status:   status,
		Level:    level,
		Message:  message,
		Fix:      fixHint,
		AutoFix:  autoFixable,
	})
}

func (s *DoctorService) checkFramework(report *domain.DoctorReport) {
	if s.docRepo.IsDir(".mind") {
		s.addDiag(report, "framework", ".mind/ directory", domain.DiagPass, ".mind/ directory exists", "", false)
	} else {
		s.addDiag(report, "framework", ".mind/ directory", domain.DiagFail, ".mind/ directory missing", "Run: mind init", false)
	}
}

func (s *DoctorService) checkAdapters(report *domain.DoctorReport) {
	if s.docRepo.Exists(".claude/CLAUDE.md") {
		s.addDiag(report, "framework", "Claude adapter", domain.DiagPass, ".claude/CLAUDE.md exists", "", false)
	} else {
		s.addDiag(report, "framework", "Claude adapter", domain.DiagFail, ".claude/CLAUDE.md missing", "Run: mind init", true)
	}

	if s.docRepo.IsDir(".github/agents") {
		s.addDiag(report, "framework", "GitHub agents", domain.DiagPass, ".github/agents/ exists", "", false)
	} else {
		s.addDiag(report, "framework", "GitHub agents", domain.DiagWarn, ".github/agents/ not found", "Run: mind init --with-github", false)
	}
}

func (s *DoctorService) checkDocStructure(report *domain.DoctorReport) {
	if !s.docRepo.IsDir("docs") {
		s.addDiag(report, "docs", "docs/ directory", domain.DiagFail, "docs/ directory missing", "Run: mind init", true)
		return
	}
	s.addDiag(report, "docs", "docs/ directory", domain.DiagPass, "docs/ directory exists", "", false)

	zones := domain.AllZones
	for _, zone := range zones {
		zoneDir := filepath.Join("docs", string(zone))
		if s.docRepo.IsDir(zoneDir) {
			s.addDiag(report, "docs", string(zone)+" zone", domain.DiagPass, zoneDir+" exists", "", false)
		} else {
			s.addDiag(report, "docs", string(zone)+" zone", domain.DiagFail, zoneDir+" missing", "Create directory: "+zoneDir, true)
		}
	}

	// Required files
	requiredFiles := map[string]string{
		"docs/spec/project-brief.md": "Project brief",
		"docs/state/current.md":      "Current state",
		"docs/state/workflow.md":     "Workflow state",
		"docs/blueprints/INDEX.md":   "Blueprint index",
		"docs/knowledge/glossary.md": "Glossary",
	}

	for path, name := range requiredFiles {
		if s.docRepo.Exists(path) {
			isStub, _ := s.docRepo.IsStub(path)
			if isStub {
				s.addDiag(report, "docs", name, domain.DiagWarn, path+" is a stub", "Fill in content for "+path, false)
			} else {
				s.addDiag(report, "docs", name, domain.DiagPass, path+" exists with content", "", false)
			}
		} else {
			s.addDiag(report, "docs", name, domain.DiagFail, path+" missing", "Create file: "+path, true)
		}
	}
}

func (s *DoctorService) checkBrief(report *domain.DoctorReport) {
	brief, err := s.briefRepo.ParseBrief()
	if err != nil {
		s.addDiag(report, "brief", "Brief analysis", domain.DiagFail, fmt.Sprintf("parse error: %v", err), "", false)
		return
	}

	switch brief.GateResult {
	case domain.BriefPresent:
		s.addDiag(report, "brief", "Brief gate", domain.DiagPass, "Brief has all required sections", "", false)
	case domain.BriefStub:
		s.addDiag(report, "brief", "Brief gate", domain.DiagWarn, "Brief is a stub or missing required sections", "Fill in Vision, Key Deliverables, and Scope sections", false)
	case domain.BriefMissing:
		s.addDiag(report, "brief", "Brief gate", domain.DiagFail, "Project brief missing", "Run: mind create brief", false)
	}
}

func (s *DoctorService) checkConfig(report *domain.DoctorReport) {
	if s.configRepo == nil {
		s.addDiag(report, "config", "mind.toml", domain.DiagFail, "mind.toml not found", "Run: mind init", false)
		return
	}

	cfg, err := s.configRepo.ReadProjectConfig()
	if err != nil {
		s.addDiag(report, "config", "mind.toml", domain.DiagFail, fmt.Sprintf("parse error: %v", err), "Fix mind.toml syntax", false)
		return
	}

	s.addDiag(report, "config", "mind.toml", domain.DiagPass, "mind.toml is valid", "", false)

	if cfg.Project.Name == "" {
		s.addDiag(report, "config", "Project name", domain.DiagWarn, "project.name is empty", "Set project.name in mind.toml", false)
	}
}

func (s *DoctorService) checkWorkflow(report *domain.DoctorReport) {
	if !s.docRepo.Exists("docs/state/workflow.md") {
		s.addDiag(report, "workflow", "Workflow file", domain.DiagWarn, "docs/state/workflow.md missing", "Create file: docs/state/workflow.md", true)
		return
	}
	s.addDiag(report, "workflow", "Workflow file", domain.DiagPass, "docs/state/workflow.md exists", "", false)
}

func (s *DoctorService) checkIterations(report *domain.DoctorReport) {
	iterations, err := s.iterRepo.List()
	if err != nil {
		return
	}

	if len(iterations) == 0 {
		s.addDiag(report, "iterations", "Iteration count", domain.DiagPass, "No iterations (OK for new projects)", "", false)
		return
	}

	s.addDiag(report, "iterations", "Iteration count", domain.DiagPass, fmt.Sprintf("%d iteration(s) found", len(iterations)), "", false)

	for _, iter := range iterations {
		hasOverview := false
		complete := 0
		for _, a := range iter.Artifacts {
			if a.Exists {
				complete++
			}
			if a.Name == "overview.md" && a.Exists {
				hasOverview = true
			}
		}
		if !hasOverview {
			s.addDiag(report, "iterations", iter.DirName, domain.DiagFail, "Missing overview.md", "Create overview.md in "+iter.DirName, false)
		} else if complete < len(domain.ExpectedArtifacts) {
			s.addDiag(report, "iterations", iter.DirName, domain.DiagWarn,
				fmt.Sprintf("%d/%d artifacts present", complete, len(domain.ExpectedArtifacts)),
				"Complete remaining artifacts", false)
		}
	}
}

func (s *DoctorService) checkStaleness(report *domain.DoctorReport) {
	if s.lockRepo == nil || !s.lockRepo.Exists() {
		return
	}

	lock, err := s.lockRepo.Read()
	if err != nil || lock == nil {
		return
	}

	for id, entry := range lock.Entries {
		if entry.Stale {
			reason := entry.StaleReason
			if reason == "" {
				reason = "document is stale"
			}
			s.addDiag(report, "staleness", id, domain.DiagWarn,
				fmt.Sprintf("%s: %s", id, reason),
				"Review and update this document, then run 'mind reconcile --force'",
				false)
		}
	}
}

func (s *DoctorService) checkMCP(report *domain.DoctorReport) {
	mcpPath := filepath.Join(s.projectRoot, ".mcp.json")
	if _, err := os.Stat(mcpPath); err == nil {
		s.addDiag(report, "integration", "MCP config", domain.DiagPass, ".mcp.json exists — Claude Code can discover mind MCP tools", "", false)
	} else {
		s.addDiag(report, "integration", "MCP config", domain.DiagWarn,
			".mcp.json missing — Claude Code cannot discover mind MCP tools automatically",
			"Run: mind init (or create .mcp.json manually with {\"mcpServers\":{\"mind\":{\"command\":\"mind\",\"args\":[\"serve\"]}}})",
			true)
	}
}

func (s *DoctorService) checkClaudeCLI(report *domain.DoctorReport) {
	path, err := exec.LookPath("claude")
	if err != nil {
		s.addDiag(report, "integration", "Claude CLI", domain.DiagWarn,
			"claude CLI not found — mind workflow run will output prompts instead of launching automatically",
			"Install Claude Code: https://claude.ai/code",
			false)
		return
	}
	s.addDiag(report, "integration", "Claude CLI", domain.DiagPass, fmt.Sprintf("claude CLI found at %s", path), "", false)
}

func (s *DoctorService) applyFixes(report *domain.DoctorReport) {
	for _, d := range report.Diagnostics {
		if !d.AutoFix || d.Status == domain.DiagPass {
			continue
		}

		switch {
		case d.Category == "docs" && d.Fix != "" && len(d.Fix) > 0:
			s.tryFixDoc(report, d)
		case d.Category == "framework" && d.Check == "Claude adapter":
			s.tryFixClaudeAdapter(report)
		case d.Category == "integration" && d.Check == "MCP config":
			s.tryFixMCPConfig(report)
		}
	}
}

func (s *DoctorService) tryFixDoc(report *domain.DoctorReport, d domain.Diagnostic) {
	// Try to create missing directories
	for _, zone := range domain.AllZones {
		zoneDir := filepath.Join("docs", string(zone))
		if d.Check == string(zone)+" zone" && !s.docRepo.IsDir(zoneDir) {
			absDir := filepath.Join(s.projectRoot, zoneDir)
			if err := os.MkdirAll(absDir, 0755); err == nil {
				report.FixesApplied = append(report.FixesApplied, "Created "+zoneDir)
			}
			return
		}
	}

	// Try to create missing stub files
	stubs := map[string]string{
		"docs/spec/project-brief.md": generate.StubBriefTemplate(),
		"docs/state/current.md":      generate.CurrentStub(),
		"docs/state/workflow.md":     generate.WorkflowStub(),
		"docs/blueprints/INDEX.md":   generate.IndexStub(),
		"docs/knowledge/glossary.md": generate.GlossaryStub(),
	}

	for path, content := range stubs {
		if !s.docRepo.Exists(path) {
			absPath := filepath.Join(s.projectRoot, path)
			dir := filepath.Dir(absPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				continue
			}
			if err := os.WriteFile(absPath, []byte(content), 0644); err == nil {
				report.FixesApplied = append(report.FixesApplied, "Created "+path)
			}
		}
	}
}

func (s *DoctorService) tryFixMCPConfig(report *domain.DoctorReport) {
	mcpPath := filepath.Join(s.projectRoot, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte(generate.MCPConfigTemplate()), 0644); err == nil {
		report.FixesApplied = append(report.FixesApplied, "Created .mcp.json")
	}
}

func (s *DoctorService) tryFixClaudeAdapter(report *domain.DoctorReport) {
	absPath := filepath.Join(s.projectRoot, ".claude", "CLAUDE.md")
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	if err := os.WriteFile(absPath, []byte(generate.ClaudeAdapterTemplate()), 0644); err == nil {
		report.FixesApplied = append(report.FixesApplied, "Created .claude/CLAUDE.md")
	}
}
