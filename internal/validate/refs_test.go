package validate

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestRefsSuiteStructure verifies FR-40: 11 checks exist.
func TestRefsSuiteStructure(t *testing.T) {
	suite := RefsSuite()
	if suite.Name != "refs" {
		t.Errorf("Suite.Name = %q, want refs", suite.Name)
	}
	if len(suite.Checks) != 11 {
		t.Errorf("RefsSuite has %d checks, want 11", len(suite.Checks))
	}
}

// TestRefsCheckClaudeRefs verifies check 1: .claude/CLAUDE.md references .mind/.
func TestRefsCheckClaudeRefs(t *testing.T) {
	t.Run("CLAUDE.md exists and references .mind/", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files[".claude/CLAUDE.md"] = []byte("Read `.mind/CLAUDE.md` for the full index.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkClaudeRefs(ctx)
		if !passed {
			t.Error("should pass when .claude/CLAUDE.md references .mind/")
		}
	})

	t.Run("CLAUDE.md missing", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkClaudeRefs(ctx)
		if passed {
			t.Error("should fail when .claude/CLAUDE.md is missing")
		}
		if msg == "" {
			t.Error("should provide error message")
		}
	})

	t.Run("CLAUDE.md exists but no .mind/ reference", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files[".claude/CLAUDE.md"] = []byte("# Claude Code\n\nSome content without reference.")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkClaudeRefs(ctx)
		if passed {
			t.Error("should fail when no .mind/ reference")
		}
		if msg == "" {
			t.Error("should report the issue")
		}
	})
}

// TestRefsCheckTomlPaths verifies check 6: mind.toml paths resolve to existing files.
func TestRefsCheckTomlPaths(t *testing.T) {
	t.Run("all paths resolve", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/requirements.md"] = []byte("# Reqs")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch")

		configRepo := mem.NewConfigRepo()
		configRepo.Config = &domain.Config{
			Documents: map[string]map[string]domain.DocEntry{
				"spec": {
					"requirements": {Path: "docs/spec/requirements.md"},
					"architecture": {Path: "docs/spec/architecture.md"},
				},
			},
		}

		ctx := &CheckContext{DocRepo: docRepo, ConfigRepo: configRepo}
		passed, _ := checkTomlPaths(ctx)
		if !passed {
			t.Error("should pass when all paths resolve")
		}
	})

	// FR-40 acceptance: GIVEN mind.toml references non-existent file THEN fails
	t.Run("broken path", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/spec/requirements.md"] = []byte("# Reqs")

		configRepo := mem.NewConfigRepo()
		configRepo.Config = &domain.Config{
			Documents: map[string]map[string]domain.DocEntry{
				"spec": {
					"requirements": {Path: "docs/spec/requirements.md"},
					"missing":      {Path: "docs/spec/nonexistent.md"},
				},
			},
		}

		ctx := &CheckContext{DocRepo: docRepo, ConfigRepo: configRepo}
		passed, msg := checkTomlPaths(ctx)
		if passed {
			t.Error("should fail when a path does not resolve")
		}
		if msg == "" {
			t.Error("should list the broken path")
		}
	})

	t.Run("no config repo", func(t *testing.T) {
		ctx := &CheckContext{ConfigRepo: nil, DocRepo: mem.NewDocRepo()}
		passed, _ := checkTomlPaths(ctx)
		if !passed {
			t.Error("should pass when no ConfigRepo (graceful)")
		}
	})
}

// TestRefsCheckIndexLinks verifies check 4: INDEX.md links resolve.
func TestRefsCheckIndexLinks(t *testing.T) {
	t.Run("all links resolve", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-auth.md)\n- [BP-02](02-api.md)\n")
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("# Auth")
		docRepo.Files["docs/blueprints/02-api.md"] = []byte("# API")

		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkIndexLinks(ctx)
		if !passed {
			t.Error("should pass when all INDEX.md links resolve")
		}
	})

	t.Run("broken link", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [BP-01](01-auth.md)\n- [BP-02](02-missing.md)\n")
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("# Auth")

		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkIndexLinks(ctx)
		if passed {
			t.Error("should fail when INDEX.md has broken links")
		}
		if msg == "" {
			t.Error("should report broken link")
		}
	})

	t.Run("no INDEX.md", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkIndexLinks(ctx)
		if !passed {
			t.Error("should pass when INDEX.md does not exist (graceful)")
		}
	})

	t.Run("external links ignored", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Files["docs/blueprints/INDEX.md"] = []byte("- [External](https://example.com)\n- [Anchor](#section)\n")

		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkIndexLinks(ctx)
		if !passed {
			t.Error("should pass when only external/anchor links present")
		}
	})
}

// TestRefsCheckAgentRefs verifies check 2: agent file references.
func TestRefsCheckAgentRefs(t *testing.T) {
	t.Run("no agents dir", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkAgentRefs(ctx)
		if !passed {
			t.Error("should pass when no agents dir")
		}
	})

	t.Run("agents dir exists", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Dirs[".mind/agents"] = true
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkAgentRefs(ctx)
		if !passed {
			t.Error("should pass when agents dir exists")
		}
	})
}

// TestRefsCheckBlueprintXrefs verifies check 3: blueprint cross-references.
func TestRefsCheckBlueprintXrefs(t *testing.T) {
	t.Run("no blueprints", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintXrefs(ctx)
		if !passed {
			t.Error("should pass when no blueprints")
		}
	})

	t.Run("valid xrefs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints, Name: "01-auth",
		}
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("See [other](02-api.md) and [external](https://example.com)")
		docRepo.Files["docs/blueprints/02-api.md"] = []byte("# API")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintXrefs(ctx)
		if !passed {
			t.Error("should pass when all xrefs resolve")
		}
	})

	t.Run("broken xrefs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints, Name: "01-auth",
		}
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("See [missing](99-missing.md)")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkBlueprintXrefs(ctx)
		if passed {
			t.Error("should fail when xref does not resolve")
		}
		if msg == "" {
			t.Error("should report broken xref")
		}
	})
}

// TestRefsCheckIterOverviewRefs verifies check 5: iteration overview references.
func TestRefsCheckIterOverviewRefs(t *testing.T) {
	t.Run("nil iter repo", func(t *testing.T) {
		ctx := &CheckContext{IterRepo: nil}
		passed, _ := checkIterOverviewRefs(ctx)
		if !passed {
			t.Error("should pass when no IterRepo")
		}
	})

	t.Run("all have overview", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{
				Seq: 1, DirName: "001-NEW_PROJECT-init",
				Artifacts: []domain.Artifact{{Name: "overview.md", Exists: true}},
			},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, _ := checkIterOverviewRefs(ctx)
		if !passed {
			t.Error("should pass when all have overview")
		}
	})

	t.Run("missing overview", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{
				Seq: 1, DirName: "001-NEW_PROJECT-init",
				Artifacts: []domain.Artifact{{Name: "overview.md", Exists: false}},
			},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, msg := checkIterOverviewRefs(ctx)
		if passed {
			t.Error("should fail when overview missing")
		}
		if msg == "" {
			t.Error("should report missing overview")
		}
	})
}

// TestRefsCheckTomlGraph verifies check 7: mind.toml graph references.
func TestRefsCheckTomlGraph(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		configRepo := mem.NewConfigRepo()
		configRepo.Config = &domain.Config{}
		ctx := &CheckContext{ConfigRepo: configRepo}
		passed, _ := checkTomlGraph(ctx)
		if !passed {
			t.Error("should pass with valid config")
		}
	})

	t.Run("no config", func(t *testing.T) {
		ctx := &CheckContext{ConfigRepo: nil}
		passed, _ := checkTomlGraph(ctx)
		if !passed {
			t.Error("should pass when no config")
		}
	})
}

// TestRefsCheckSpecLinks verifies check 8: no broken links in spec/.
func TestRefsCheckSpecLinks(t *testing.T) {
	t.Run("no spec docs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkSpecLinks(ctx)
		if !passed {
			t.Error("should pass when no spec docs")
		}
	})

	t.Run("valid links", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/requirements.md"] = domain.Document{
			Path: "docs/spec/requirements.md", Zone: domain.ZoneSpec,
		}
		docRepo.Files["docs/spec/requirements.md"] = []byte("See [arch](architecture.md)")
		docRepo.Files["docs/spec/architecture.md"] = []byte("# Arch")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkSpecLinks(ctx)
		if !passed {
			t.Error("should pass when all links resolve")
		}
	})

	t.Run("broken link", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/requirements.md"] = domain.Document{
			Path: "docs/spec/requirements.md", Zone: domain.ZoneSpec,
		}
		docRepo.Files["docs/spec/requirements.md"] = []byte("See [missing](missing.md)")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkSpecLinks(ctx)
		if passed {
			t.Error("should fail when link broken")
		}
		if msg == "" {
			t.Error("should report broken link")
		}
	})
}

// TestRefsCheckBlueprintLinks verifies check 9: no broken links in blueprints/.
func TestRefsCheckBlueprintLinks(t *testing.T) {
	t.Run("no blueprints", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintLinks(ctx)
		if !passed {
			t.Error("should pass when no blueprints")
		}
	})

	t.Run("valid links", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints,
		}
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("See [other](02-api.md) and [external](https://example.com)")
		docRepo.Files["docs/blueprints/02-api.md"] = []byte("# API")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkBlueprintLinks(ctx)
		if !passed {
			t.Error("should pass when all links resolve")
		}
	})

	t.Run("broken link", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/blueprints/01-auth.md"] = domain.Document{
			Path: "docs/blueprints/01-auth.md", Zone: domain.ZoneBlueprints,
		}
		docRepo.Files["docs/blueprints/01-auth.md"] = []byte("See [missing](99-missing.md)")
		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkBlueprintLinks(ctx)
		if passed {
			t.Error("should fail when link broken")
		}
		if msg == "" {
			t.Error("should report broken link")
		}
	})
}

// TestRefsCheckADRSequence verifies check 10: ADR numbering sequential.
func TestRefsCheckADRSequence(t *testing.T) {
	t.Run("sequential ADRs", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/001-auth.md"] = domain.Document{Path: "docs/spec/decisions/001-auth.md", Zone: domain.ZoneSpec}
		docRepo.Docs["docs/spec/decisions/002-db.md"] = domain.Document{Path: "docs/spec/decisions/002-db.md", Zone: domain.ZoneSpec}
		docRepo.Docs["docs/spec/decisions/003-api.md"] = domain.Document{Path: "docs/spec/decisions/003-api.md", Zone: domain.ZoneSpec}

		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkADRSequence(ctx)
		if !passed {
			t.Error("should pass for sequential ADRs")
		}
	})

	t.Run("gap in ADR sequence", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/001-auth.md"] = domain.Document{Path: "docs/spec/decisions/001-auth.md", Zone: domain.ZoneSpec}
		docRepo.Docs["docs/spec/decisions/003-api.md"] = domain.Document{Path: "docs/spec/decisions/003-api.md", Zone: domain.ZoneSpec}

		ctx := &CheckContext{DocRepo: docRepo}
		passed, msg := checkADRSequence(ctx)
		if passed {
			t.Error("should fail for non-sequential ADRs")
		}
		if msg == "" {
			t.Error("should report gap")
		}
	})

	t.Run("single ADR passes", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		docRepo.Docs["docs/spec/decisions/001-auth.md"] = domain.Document{Path: "docs/spec/decisions/001-auth.md", Zone: domain.ZoneSpec}

		ctx := &CheckContext{DocRepo: docRepo}
		passed, _ := checkADRSequence(ctx)
		if !passed {
			t.Error("should pass for single ADR")
		}
	})
}

// TestRefsCheckIterSequence verifies check 11: iteration numbering sequential.
func TestRefsCheckIterSequence(t *testing.T) {
	t.Run("sequential iterations", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{Seq: 1, DirName: "001-NEW_PROJECT-init"},
			{Seq: 2, DirName: "002-ENHANCEMENT-feature"},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, _ := checkIterSequence(ctx)
		if !passed {
			t.Error("should pass for sequential iterations")
		}
	})

	t.Run("gap in iteration sequence", func(t *testing.T) {
		iterRepo := mem.NewIterationRepo()
		iterRepo.Iterations = []domain.Iteration{
			{Seq: 1, DirName: "001-NEW_PROJECT-init"},
			{Seq: 3, DirName: "003-ENHANCEMENT-feature"},
		}
		ctx := &CheckContext{IterRepo: iterRepo}
		passed, msg := checkIterSequence(ctx)
		if passed {
			t.Error("should fail for non-sequential iterations")
		}
		if msg == "" {
			t.Error("should report gap")
		}
	})

	t.Run("nil iter repo passes", func(t *testing.T) {
		ctx := &CheckContext{IterRepo: nil}
		passed, _ := checkIterSequence(ctx)
		if !passed {
			t.Error("should pass when no IterRepo")
		}
	})
}
