package framework

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// Helper: create a minimal .mind/ source tree with artifact files.
func setupSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	for _, kind := range ArtifactKinds {
		dir := filepath.Join(src, kind)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sample.md"), []byte("# "+kind+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Write mind.toml alongside source so detectVersion can find it.
	toml := `[framework]
version = "2025.01.1"
mode = "standalone"
`
	if err := os.WriteFile(filepath.Join(filepath.Dir(src), "mind.toml"), []byte(toml), 0644); err != nil {
		// In TempDir, the parent might not be writable — place inside src's parent.
		// detectVersion looks at filepath.Dir(source)/mind.toml, which is the temp root.
		t.Logf("could not write mind.toml next to source: %v (will use generated CalVer)", err)
	}
	return src
}

func TestInstall_Fresh(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	result, err := Install(src, globalDir, false)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	if result.ArtifactCount != len(ArtifactKinds) {
		t.Errorf("expected %d artifacts, got %d", len(ArtifactKinds), result.ArtifactCount)
	}
	if result.Overwritten {
		t.Error("expected Overwritten=false for fresh install")
	}
	if result.Source != src {
		t.Errorf("expected source=%s, got %s", src, result.Source)
	}

	// Verify framework.lock was created.
	lockPath := filepath.Join(globalDir, "framework.lock")
	lock, err := ReadLock(lockPath)
	if err != nil {
		t.Fatalf("ReadLock: %v", err)
	}
	if len(lock.Checksums) != len(ArtifactKinds) {
		t.Errorf("expected %d checksums, got %d", len(ArtifactKinds), len(lock.Checksums))
	}
}

func TestInstall_BlocksWithoutForce(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("first install: %v", err)
	}

	_, err := Install(src, globalDir, false)
	if err == nil {
		t.Fatal("expected error on re-install without --force")
	}
}

func TestInstall_ForceOverwrite(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("first install: %v", err)
	}

	result, err := Install(src, globalDir, true)
	if err != nil {
		t.Fatalf("force install: %v", err)
	}
	if !result.Overwritten {
		t.Error("expected Overwritten=true on force reinstall")
	}
}

func TestStatus_NotInstalled(t *testing.T) {
	globalDir := filepath.Join(t.TempDir(), "empty")

	result, err := Status(globalDir, nil)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if result.Installed {
		t.Error("expected Installed=false when no lock file")
	}
}

func TestStatus_Installed_NoDrift(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	fw := &domain.FrameworkConfig{Version: "2025.01.1", Mode: "standalone"}
	result, err := Status(globalDir, fw)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !result.Installed {
		t.Error("expected Installed=true")
	}
	if len(result.DriftFiles) != 0 {
		t.Errorf("expected no drift, got %d files: %v", len(result.DriftFiles), result.DriftFiles)
	}
}

func TestStatus_DetectsDrift(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Modify one installed file to create drift.
	driftFile := filepath.Join(globalDir, "agents", "sample.md")
	if err := os.WriteFile(driftFile, []byte("modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Status(globalDir, nil)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(result.DriftFiles) != 1 {
		t.Errorf("expected 1 drift file, got %d: %v", len(result.DriftFiles), result.DriftFiles)
	}
}

func TestDiff_NoDifferences(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Use same source as "project .mind/" — should be identical.
	result, err := Diff(src, globalDir)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if result.HasDiff {
		t.Errorf("expected no diff, got %d entries: %v", len(result.Entries), result.Entries)
	}
}

func TestDiff_DetectsModified(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Create a project .mind/ with a modified file.
	projMind := filepath.Join(t.TempDir(), "project-mind")
	for _, kind := range ArtifactKinds {
		dir := filepath.Join(projMind, kind)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "sample.md"), []byte("# "+kind+"\n"), 0644)
	}
	// Modify one file in the project copy.
	os.WriteFile(filepath.Join(projMind, "agents", "sample.md"), []byte("different content\n"), 0644)

	result, err := Diff(projMind, globalDir)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !result.HasDiff {
		t.Fatal("expected HasDiff=true")
	}

	found := false
	for _, e := range result.Entries {
		if e.Path == filepath.Join("agents", "sample.md") && e.Status == "modified" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'modified' entry for agents/sample.md, got: %v", result.Entries)
	}
}

func TestDiff_DetectsExtra(t *testing.T) {
	src := setupSource(t)
	globalDir := filepath.Join(t.TempDir(), "global")

	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Create project .mind/ with an extra file.
	projMind := filepath.Join(t.TempDir(), "project-mind")
	for _, kind := range ArtifactKinds {
		dir := filepath.Join(projMind, kind)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "sample.md"), []byte("# "+kind+"\n"), 0644)
	}
	os.WriteFile(filepath.Join(projMind, "agents", "custom.md"), []byte("custom agent\n"), 0644)

	result, err := Diff(projMind, globalDir)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !result.HasDiff {
		t.Fatal("expected HasDiff=true")
	}

	found := false
	for _, e := range result.Entries {
		if e.Path == filepath.Join("agents", "custom.md") && e.Status == "extra" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'extra' entry for agents/custom.md, got: %v", result.Entries)
	}
}

func TestLock_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "framework.lock")

	original := &FrameworkLock{
		Framework: LockFramework{
			Version:     "2025.01.1",
			Source:      "/some/path/.mind",
			InstalledAt: "2025-01-15T10:30:00Z",
		},
		Checksums: map[string]string{
			"agents/sample.md": "sha256:abc123",
			"skills/tools.md":  "sha256:def456",
		},
	}

	if err := WriteLock(lockPath, original); err != nil {
		t.Fatalf("WriteLock: %v", err)
	}

	loaded, err := ReadLock(lockPath)
	if err != nil {
		t.Fatalf("ReadLock: %v", err)
	}

	if loaded.Framework.Version != original.Framework.Version {
		t.Errorf("version mismatch: got %s, want %s", loaded.Framework.Version, original.Framework.Version)
	}
	if loaded.Framework.Source != original.Framework.Source {
		t.Errorf("source mismatch: got %s, want %s", loaded.Framework.Source, original.Framework.Source)
	}
	if len(loaded.Checksums) != len(original.Checksums) {
		t.Errorf("checksums count: got %d, want %d", len(loaded.Checksums), len(original.Checksums))
	}
	for k, v := range original.Checksums {
		if loaded.Checksums[k] != v {
			t.Errorf("checksum %s: got %s, want %s", k, loaded.Checksums[k], v)
		}
	}
}

func TestDoctorChecks_NotInstalled(t *testing.T) {
	projectRoot := t.TempDir()
	// Ensure global dir doesn't have framework.lock
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))

	checks := RunDoctorChecks(projectRoot, nil)
	if len(checks) == 0 {
		t.Fatal("expected at least one check")
	}
	if checks[0].Status != domain.DiagWarn {
		t.Errorf("expected first check to be warn, got %s", checks[0].Status)
	}
}

func TestDoctorChecks_InstalledAllPass(t *testing.T) {
	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	globalDir := filepath.Join(xdg, "mind")

	src := setupSource(t)
	if _, err := Install(src, globalDir, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	projectRoot := t.TempDir()
	os.MkdirAll(filepath.Join(projectRoot, ".mind"), 0755)

	cfg := &domain.Config{
		Framework: &domain.FrameworkConfig{Version: "2025.01.1", Mode: "standalone"},
	}

	checks := RunDoctorChecks(projectRoot, cfg)
	for _, c := range checks {
		if c.Status == domain.DiagFail {
			t.Errorf("check %q failed: %s", c.Check, c.Message)
		}
	}
}

func TestInstall_EmptySource(t *testing.T) {
	// Source with no artifact directories — should succeed with 0 artifacts.
	src := t.TempDir()
	globalDir := filepath.Join(t.TempDir(), "global")

	result, err := Install(src, globalDir, false)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if result.ArtifactCount != 0 {
		t.Errorf("expected 0 artifacts from empty source, got %d", result.ArtifactCount)
	}
}
