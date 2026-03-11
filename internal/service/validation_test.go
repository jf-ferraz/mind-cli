package service

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestValidationServiceRunDocs verifies FR-38: RunDocs returns 17 checks.
func TestValidationServiceRunDocs(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunDocs("/test", false)

	if report.Suite != "docs" {
		t.Errorf("Suite = %q, want docs", report.Suite)
	}
	if report.Total != 17 {
		t.Errorf("Total = %d, want 17", report.Total)
	}
}

// TestValidationServiceRunRefs verifies FR-40: RunRefs returns 11 checks.
func TestValidationServiceRunRefs(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunRefs("/test")

	if report.Suite != "refs" {
		t.Errorf("Suite = %q, want refs", report.Suite)
	}
	if report.Total != 11 {
		t.Errorf("Total = %d, want 11", report.Total)
	}
}

// TestValidationServiceRunConfig verifies FR-41: RunConfig returns 10 checks.
func TestValidationServiceRunConfig(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunConfig("/test")

	if report.Suite != "config" {
		t.Errorf("Suite = %q, want config", report.Suite)
	}
	if report.Total != 10 {
		t.Errorf("Total = %d, want 10", report.Total)
	}
}

// TestValidationServiceRunAll verifies FR-42: RunAll produces unified report.
func TestValidationServiceRunAll(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunAll("/test", false)

	// FR-42: should have 3 suites
	if len(report.Suites) != 3 {
		t.Errorf("Suites count = %d, want 3", len(report.Suites))
	}

	// Verify suites are in order: docs, refs, config
	expectedNames := []string{"docs", "refs", "config"}
	for i, name := range expectedNames {
		if i < len(report.Suites) && report.Suites[i].Suite != name {
			t.Errorf("Suites[%d].Suite = %q, want %q", i, report.Suites[i].Suite, name)
		}
	}

	// Summary totals should match
	expectedTotal := 17 + 11 + 10 // docs + refs + config
	if report.Summary.Total != expectedTotal {
		t.Errorf("Summary.Total = %d, want %d", report.Summary.Total, expectedTotal)
	}

	// Summary counts should add up
	sumFromSuites := 0
	for _, s := range report.Suites {
		sumFromSuites += s.Passed + s.Failed + s.Warnings
	}
	summarySum := report.Summary.Passed + report.Summary.Failed + report.Summary.Warnings
	if sumFromSuites != summarySum {
		t.Errorf("Suite counts sum (%d) != summary counts sum (%d)", sumFromSuites, summarySum)
	}
}

// TestValidationServiceStrictMode verifies FR-39: strict propagation.
func TestValidationServiceStrictMode(t *testing.T) {
	docRepo := mem.NewDocRepo()
	// Set up a project with stubs (warn-level)
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/spec/requirements.md"] = []byte("# Reqs\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/state/current.md"] = []byte("# Current\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\n\n<!-- placeholder -->\n")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Index\n\n<!-- placeholder -->\n")

	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	// Without strict
	normalReport := svc.RunDocs("/test", false)
	// With strict
	strictReport := svc.RunDocs("/test", true)

	// Strict mode should have more failures (warnings promoted)
	if strictReport.Failed <= normalReport.Failed {
		t.Errorf("Strict mode failures (%d) should be > normal mode (%d)", strictReport.Failed, normalReport.Failed)
	}
}
