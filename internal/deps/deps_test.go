package deps

import (
	"os"
	"path/filepath"
	"testing"
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
