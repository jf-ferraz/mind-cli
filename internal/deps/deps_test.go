package deps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// FR-90: Build creates all expected services and repos.
func TestBuild_AllFieldsPopulated(t *testing.T) {
	// Create a minimal project directory structure
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mind"), 0o755)
	os.MkdirAll(filepath.Join(root, "docs", "spec"), 0o755)
	os.WriteFile(filepath.Join(root, "mind.toml"), []byte("[manifest]\nschema = \"1.0\"\n"), 0o644)

	d := Build(root, nil)

	if d == nil {
		t.Fatal("Build returned nil")
	}

	// Repos
	if d.DocRepo == nil {
		t.Error("DocRepo is nil")
	}
	if d.IterRepo == nil {
		t.Error("IterRepo is nil")
	}
	if d.BriefRepo == nil {
		t.Error("BriefRepo is nil")
	}
	if d.ConfigRepo == nil {
		t.Error("ConfigRepo is nil")
	}
	if d.LockRepo == nil {
		t.Error("LockRepo is nil")
	}
	if d.StateRepo == nil {
		t.Error("StateRepo is nil")
	}
	if d.QualityRepo == nil {
		t.Error("QualityRepo is nil")
	}

	// Services
	if d.ProjectSvc == nil {
		t.Error("ProjectSvc is nil")
	}
	if d.ValidationSvc == nil {
		t.Error("ValidationSvc is nil")
	}
	if d.ReconcileSvc == nil {
		t.Error("ReconcileSvc is nil")
	}
	if d.DoctorSvc == nil {
		t.Error("DoctorSvc is nil")
	}
	if d.WorkflowSvc == nil {
		t.Error("WorkflowSvc is nil")
	}
	if d.GenerateSvc == nil {
		t.Error("GenerateSvc is nil")
	}

	// Project root preserved
	if d.ProjectRoot != root {
		t.Errorf("ProjectRoot = %q, want %q", d.ProjectRoot, root)
	}
}

// FR-125: Build returns Deps with repository fields that satisfy repo interfaces.
// This test explicitly asserts the interface types, strengthening the compile-time
// guarantee that Deps uses repo.X interfaces (not concrete *fs.X types).
func TestBuild_RepoFieldsSatisfyInterfaces(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mind"), 0o755)

	d := Build(root, nil)
	if d == nil {
		t.Fatal("Build returned nil")
	}

	// Assert each repo field satisfies its interface.
	// These assignments would fail to compile if Deps used concrete *fs.X types
	// and the concrete types did not implement the interfaces.
	var _ repo.DocRepo = d.DocRepo
	var _ repo.IterationRepo = d.IterRepo
	var _ repo.BriefRepo = d.BriefRepo
	var _ repo.ConfigRepo = d.ConfigRepo
	var _ repo.LockRepo = d.LockRepo
	var _ repo.StateRepo = d.StateRepo
	var _ repo.QualityRepo = d.QualityRepo
}

// FR-90: Build with nil renderer is valid (TUI path).
func TestBuild_NilRenderer(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mind"), 0o755)

	d := Build(root, nil)
	if d == nil {
		t.Fatal("Build returned nil")
	}
	if d.Renderer != nil {
		t.Error("expected nil Renderer for TUI path")
	}
}
