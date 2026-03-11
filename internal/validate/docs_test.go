package validate

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestDocsSuiteStructure verifies FR-38: 17 checks exist.
func TestDocsSuiteStructure(t *testing.T) {
	suite := DocsSuite()
	if suite.Name != "docs" {
		t.Errorf("Suite.Name = %q, want docs", suite.Name)
	}
	if len(suite.Checks) != 17 {
		t.Errorf("DocsSuite has %d checks, want 17", len(suite.Checks))
	}

	// Verify sequential IDs
	for i, check := range suite.Checks {
		expected := i + 1
		if check.ID != expected {
			t.Errorf("Check[%d].ID = %d, want %d", i, check.ID, expected)
		}
	}
}

// TestDocsSuiteAllPass verifies FR-38: all 17 checks pass for a well-formed project.
func TestDocsSuiteAllPass(t *testing.T) {
	docRepo := mem.NewDocRepo()
	briefRepo := mem.NewBriefRepo()
	iterRepo := mem.NewIterationRepo()

	// Set up well-formed project structure
	setupWellFormedDocs(docRepo)

	// Brief is present
	briefRepo.Brief = &domain.Brief{
		Exists:          true,
		IsStub:          false,
		HasVision:       true,
		HasDeliverables: true,
		HasScope:        true,
		GateResult:      domain.BriefPresent,
	}

	ctx := &CheckContext{
		ProjectRoot: "/test",
		DocRepo:     docRepo,
		IterRepo:    iterRepo,
		BriefRepo:   briefRepo,
	}

	suite := DocsSuite()
	report := suite.Run(ctx)

	if report.Failed > 0 {
		t.Errorf("Well-formed project should have 0 failures, got %d", report.Failed)
		for _, cr := range report.Checks {
			if !cr.Passed {
				t.Errorf("  FAIL: [%d] %s: %s", cr.ID, cr.Name, cr.Message)
			}
		}
	}
}

// TestDocsSuiteDocsDir verifies check 1: docs/ directory exists.
func TestDocsSuiteDocsDir(t *testing.T) {
	tests := []struct {
		name     string
		hasDir   bool
		wantPass bool
	}{
		{name: "docs/ exists", hasDir: true, wantPass: true},
		{name: "docs/ missing", hasDir: false, wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docRepo := mem.NewDocRepo()
			if tt.hasDir {
				docRepo.Dirs["docs"] = true
			}

			ctx := &CheckContext{DocRepo: docRepo}
			passed, _ := checkDocsDir(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkDocsDir() = %v, want %v", passed, tt.wantPass)
			}
		})
	}
}

// TestDocsSuiteZoneDirs verifies check 2: all 5 zone directories exist.
func TestDocsSuiteZoneDirs(t *testing.T) {
	t.Run("all zones present", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		for _, zone := range domain.AllZones {
			docRepo.Dirs["docs/"+string(zone)] = true
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkZoneDirs(ctx)
		if !passed {
			t.Error("checkZoneDirs should pass when all zones present")
		}
	})

	t.Run("missing zone", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		// Add all except one
		for _, zone := range domain.AllZones[:4] {
			docRepo.Dirs["docs/"+string(zone)] = true
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkZoneDirs(ctx)
		if passed {
			t.Error("checkZoneDirs should fail when a zone is missing")
		}
		if msg == "" {
			t.Error("should provide message about missing zone")
		}
	})
}

// TestDocsSuiteSpecFiles verifies check 3: required spec files.
func TestDocsSuiteSpecFiles(t *testing.T) {
	t.Run("all required files present", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\nReal content here.\nAnd more content.\nEven more.")
		docRepo.Files["docs/spec/requirements.md"] = []byte("# Requirements\n\nReal content.\nMore content.\nStill more.")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Architecture\n\nReal content.\nDesign details.\nMore info.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkSpecFiles(ctx)
		if !passed {
			t.Error("checkSpecFiles should pass when all files exist")
		}
	})

	t.Run("missing requirements", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\nContent.\nMore.\nStuff.")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch\n\nContent.\nMore.\nStuff.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkSpecFiles(ctx)
		if passed {
			t.Error("checkSpecFiles should fail when requirements.md is missing")
		}
		if msg == "" {
			t.Error("should mention missing file")
		}
	})
}

// TestDocsSuiteStubDetection verifies check 16: stub documents detected as warnings.
func TestDocsSuiteStubDetection(t *testing.T) {
	t.Run("no stubs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\nReal content.\nMore content.\nStill more.")
		docRepo.Files["docs/spec/requirements.md"] = []byte("# Reqs\n\nReal content.\nMore content.\nStill more.")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch\n\nReal content.\nMore content.\nStill more.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkStubs(ctx)
		if !passed {
			t.Error("checkStubs should pass when no key docs are stubs")
		}
	})

	t.Run("stub detected", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\n<!-- placeholder -->\n")
		docRepo.Files["docs/spec/requirements.md"] = []byte("# Reqs\n\nReal content.\nMore content.\nStill more.")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch\n\nReal content.\nMore content.\nStill more.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkStubs(ctx)
		if passed {
			t.Error("checkStubs should fail when a key doc is a stub")
		}
		if msg == "" {
			t.Error("should report which file is a stub")
		}
	})
}

// TestDocsSuiteBriefCompleteness verifies check 17: brief completeness.
func TestDocsSuiteBriefCompleteness(t *testing.T) {
	t.Run("brief present", func(t *testing.T) {
		briefRepo := mem.NewBriefRepo()
		briefRepo.Brief = &domain.Brief{
			Exists:          true,
			HasVision:       true,
			HasDeliverables: true,
			HasScope:        true,
			GateResult:      domain.BriefPresent,
		}
		ctx := &CheckContext{BriefRepo: briefRepo}
		passed, _ := checkBriefCompleteness(ctx)
		if !passed {
			t.Error("checkBriefCompleteness should pass when all sections present")
		}
	})

	t.Run("brief missing sections", func(t *testing.T) {
		briefRepo := mem.NewBriefRepo()
		briefRepo.Brief = &domain.Brief{
			Exists:          true,
			HasVision:       true,
			HasDeliverables: false,
			HasScope:        false,
			GateResult:      domain.BriefStub,
		}
		ctx := &CheckContext{BriefRepo: briefRepo}
		passed, msg := checkBriefCompleteness(ctx)
		if passed {
			t.Error("checkBriefCompleteness should fail when sections missing")
		}
		if msg == "" {
			t.Error("should list missing sections")
		}
	})

	t.Run("no brief repo", func(t *testing.T) {
		ctx := &CheckContext{BriefRepo: nil}
		passed, _ := checkBriefCompleteness(ctx)
		if !passed {
			t.Error("checkBriefCompleteness should pass when no BriefRepo (graceful)")
		}
	})
}

// TestDocsSuiteNoLegacyPaths verifies check 15: no legacy paths.
func TestDocsSuiteNoLegacyPaths(t *testing.T) {
	t.Run("no legacy paths", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkNoLegacyPaths(ctx)
		if !passed {
			t.Error("should pass when no legacy paths")
		}
	})

	t.Run("legacy path found", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Dirs["docs/adr"] = true
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkNoLegacyPaths(ctx)
		if passed {
			t.Error("should fail when legacy path exists")
		}
		if msg == "" {
			t.Error("should list legacy path")
		}
	})
}

// TestDocsSuiteIterationNaming verifies check 12: iteration naming.
func TestDocsSuiteIterationNaming(t *testing.T) {
	t.Run("no iterations is pass", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, _ := checkIterationNaming(ctx)
		if !passed {
			t.Error("should pass with no iterations")
		}
	})

	t.Run("well-named iterations pass", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{Seq: 1, DirName: "001-NEW_PROJECT-core-cli"},
			{Seq: 2, DirName: "002-ENHANCEMENT-add-caching"},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, _ := checkIterationNaming(ctx)
		if !passed {
			t.Error("should pass with well-named iterations")
		}
	})

	t.Run("nil iter repo is pass", func(t *testing.T) {
		ctx := &CheckContext{IterRepo: nil}
		passed, _ := checkIterationNaming(ctx)
		if !passed {
			t.Error("should pass when IterRepo is nil")
		}
	})
}

// TestDocsSuiteDecisionsDir verifies check 4: decisions/ subdirectory.
func TestDocsSuiteDecisionsDir(t *testing.T) {
	t.Run("decisions dir exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Dirs["docs/spec/decisions"] = true
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkDecisionsDir(ctx)
		if !passed {
			t.Error("should pass when decisions/ exists")
		}
	})

	t.Run("decisions dir missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkDecisionsDir(ctx)
		if passed {
			t.Error("should fail when decisions/ missing")
		}
		if msg == "" {
			t.Error("should provide message")
		}
	})
}

// TestDocsSuiteADRNaming verifies check 5: ADR naming convention.
func TestDocsSuiteADRNaming(t *testing.T) {
	t.Run("well-named ADRs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/001-auth.md"] = domain.Document{
			Path: "docs/spec/decisions/001-auth.md", Zone: domain.ZoneSpec,
		}
		docRepo.Docs["docs/spec/decisions/002-database.md"] = domain.Document{
			Path: "docs/spec/decisions/002-database.md", Zone: domain.ZoneSpec,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkADRNaming(ctx)
		if !passed {
			t.Error("should pass for well-named ADRs")
		}
	})

	t.Run("bad ADR name", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/auth-decision.md"] = domain.Document{
			Path: "docs/spec/decisions/auth-decision.md", Zone: domain.ZoneSpec,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkADRNaming(ctx)
		if passed {
			t.Error("should fail for bad ADR name (no number prefix)")
		}
		if msg == "" {
			t.Error("should report bad name")
		}
	})

	t.Run("template file ignored", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/_template.md"] = domain.Document{
			Path: "docs/spec/decisions/_template.md", Zone: domain.ZoneSpec,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkADRNaming(ctx)
		if !passed {
			t.Error("should pass: _template.md is excluded from naming check")
		}
	})
}

// TestDocsSuiteBlueprintsIndex verifies check 6: INDEX.md exists.
func TestDocsSuiteBlueprintsIndex(t *testing.T) {
	t.Run("INDEX.md exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Index")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintsIndex(ctx)
		if !passed {
			t.Error("should pass when INDEX.md exists")
		}
	})

	t.Run("INDEX.md missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkBlueprintsIndex(ctx)
		if passed {
			t.Error("should fail when INDEX.md missing")
		}
		if msg == "" {
			t.Error("should provide message")
		}
	})
}

// TestDocsSuiteBlueprintCoverage verifies check 7: blueprints in INDEX.md.
func TestDocsSuiteBlueprintCoverage(t *testing.T) {
	t.Run("all covered", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints,
		}
		docRepo.Docs["docs/blueprints/INDEX.md"] = domain.Document{
			Path: "docs/blueprints/INDEX.md", Zone: domain.ZoneBlueprints, Name: "INDEX",
		}
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-auth.md)")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintCoverage(ctx)
		if !passed {
			t.Error("should pass when all blueprints in INDEX.md")
		}
	})

	t.Run("uncovered blueprint", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints,
		}
		docRepo.Docs["docs/blueprints/02-api.md"] = domain.Document{
			Path: "docs/blueprints/02-api.md", Zone: domain.ZoneBlueprints,
		}
		docRepo.Docs["docs/blueprints/INDEX.md"] = domain.Document{
			Path: "docs/blueprints/INDEX.md", Zone: domain.ZoneBlueprints, Name: "INDEX",
		}
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-auth.md)")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkBlueprintCoverage(ctx)
		if passed {
			t.Error("should fail when blueprint not in INDEX.md")
		}
		if msg == "" {
			t.Error("should report uncovered blueprint")
		}
	})
}

// TestDocsSuiteIndexRefs verifies check 8: INDEX.md references resolve.
func TestDocsSuiteIndexRefs(t *testing.T) {
	t.Run("all refs resolve", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-auth.md)\n")
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("# Auth")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkIndexRefs(ctx)
		if !passed {
			t.Error("should pass when all refs resolve")
		}
	})

	t.Run("broken ref", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-missing.md)\n")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkIndexRefs(ctx)
		if passed {
			t.Error("should fail when ref does not resolve")
		}
		if msg == "" {
			t.Error("should report broken ref")
		}
	})
}

// TestDocsSuiteCurrentState verifies check 9: current.md exists.
func TestDocsSuiteCurrentState(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/state/current.md"] = []byte("# Current")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkCurrentState(ctx)
		if !passed {
			t.Error("should pass")
		}
	})

	t.Run("missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkCurrentState(ctx)
		if passed {
			t.Error("should fail")
		}
	})
}

// TestDocsSuiteWorkflowState verifies check 10: workflow.md exists.
func TestDocsSuiteWorkflowState(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkWorkflowState(ctx)
		if !passed {
			t.Error("should pass")
		}
	})

	t.Run("missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkWorkflowState(ctx)
		if passed {
			t.Error("should fail")
		}
	})
}

// TestDocsSuiteGlossary verifies check 11: glossary.md exists.
func TestDocsSuiteGlossary(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkGlossary(ctx)
		if !passed {
			t.Error("should pass")
		}
	})

	t.Run("missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkGlossary(ctx)
		if passed {
			t.Error("should fail")
		}
	})
}

// TestDocsSuiteIterationOverview verifies check 13: iterations have overview.md.
func TestDocsSuiteIterationOverview(t *testing.T) {
	t.Run("all have overview", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{
				Seq: 1, DirName: "001-NEW_PROJECT-init",
				Artifacts: []domain.Artifact{
					{Name: "overview.md", Exists: true},
					{Name: "changes.md", Exists: true},
					{Name: "test-summary.md", Exists: false},
					{Name: "validation.md", Exists: false},
					{Name: "retrospective.md", Exists: false},
				},
			},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, _ := checkIterationOverview(ctx)
		if !passed {
			t.Error("should pass when all iterations have overview.md")
		}
	})

	t.Run("missing overview", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{
				Seq: 1, DirName: "001-NEW_PROJECT-init",
				Artifacts: []domain.Artifact{
					{Name: "overview.md", Exists: false},
					{Name: "changes.md", Exists: true},
					{Name: "test-summary.md", Exists: false},
					{Name: "validation.md", Exists: false},
					{Name: "retrospective.md", Exists: false},
				},
			},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, msg := checkIterationOverview(ctx)
		if passed {
			t.Error("should fail when overview.md is missing")
		}
		if msg == "" {
			t.Error("should report which iteration lacks overview.md")
		}
	})

	t.Run("nil iter repo", func(t *testing.T) {
		ctx := &CheckContext{IterRepo: nil}
		passed, _ := checkIterationOverview(ctx)
		if !passed {
			t.Error("should pass when IterRepo is nil")
		}
	})
}

// TestDocsSuiteSpikeNaming verifies check 14: spike file naming convention.
func TestDocsSuiteSpikeNaming(t *testing.T) {
	t.Run("well-named spike", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/knowledge/redis-spike.md"] = domain.Document{
			Path: "docs/knowledge/redis-spike.md", Zone: domain.ZoneKnowledge,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkSpikeNaming(ctx)
		if !passed {
			t.Error("should pass for well-named spike")
		}
	})

	t.Run("badly named spike", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/knowledge/spike-redis.md"] = domain.Document{
			Path: "docs/knowledge/spike-redis.md", Zone: domain.ZoneKnowledge,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkSpikeNaming(ctx)
		if passed {
			t.Error("should fail for spike file not ending in -spike.md")
		}
		if msg == "" {
			t.Error("should report bad name")
		}
	})

	t.Run("glossary and convergence ignored", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/knowledge/glossary.md"] = domain.Document{
			Path: "docs/knowledge/glossary.md", Zone: domain.ZoneKnowledge,
		}
		docRepo.Docs["docs/knowledge/auth-convergence.md"] = domain.Document{
			Path: "docs/knowledge/auth-convergence.md", Zone: domain.ZoneKnowledge,
		}
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkSpikeNaming(ctx)
		if !passed {
			t.Error("should pass for non-spike files")
		}
	})
}

// setupWellFormedDocs populates a DocRepo with a valid project structure.
func setupWellFormedDocs(docRepo *mem.DocRepo) {
	// Directories
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Dirs["docs/spec/decisions"] = true

	// Required files with real content
	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\n## Vision\nBuild a CLI tool.\n\n## Key Deliverables\n- CLI binary\n\n## Scope\nPhase 1.")
	docRepo.Files["docs/spec/requirements.md"] = []byte("# Requirements\n\nFunctional requirements for the system.\nFR-1: The system shall...\nFR-2: The system shall...")
	docRepo.Files["docs/spec/architecture.md"] = []byte("# Architecture\n\n4-layer architecture with domain purity.\nPresentation layer uses Cobra.\nService layer orchestrates.")
	docRepo.Files["docs/state/current.md"] = []byte("# Current State\n\nActive work on Phase 1.\nCore CLI implementation.\nAll tests passing.")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\n\nCurrent workflow state.\nRunning core-cli iteration.\nTester agent active.")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\n\n| Term | Definition |\n|------|-----------|")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Blueprints\n\nIndex of all blueprints.\nManaged by mind create.")

	// Add corresponding docs entries
	docRepo.Docs["docs/spec/project-brief.md"] = domain.Document{
		Path: "docs/spec/project-brief.md", Zone: domain.ZoneSpec, Name: "project-brief",
	}
	docRepo.Docs["docs/spec/requirements.md"] = domain.Document{
		Path: "docs/spec/requirements.md", Zone: domain.ZoneSpec, Name: "requirements",
	}
	docRepo.Docs["docs/spec/architecture.md"] = domain.Document{
		Path: "docs/spec/architecture.md", Zone: domain.ZoneSpec, Name: "architecture",
	}
	docRepo.Docs["docs/blueprints/INDEX.md"] = domain.Document{
		Path: "docs/blueprints/INDEX.md", Zone: domain.ZoneBlueprints, Name: "INDEX",
	}
}
