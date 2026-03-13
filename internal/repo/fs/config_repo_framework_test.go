package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRepo_FrameworkPresent(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v2.0"
generation = 1

[project]
name = "test-project"
type = "cli"

[framework]
version = "2026.03.1"
mode = "standalone"
`
	os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644)

	repo := NewConfigRepo(dir)
	cfg, err := repo.ReadProjectConfig()
	if err != nil {
		t.Fatalf("ReadProjectConfig: %v", err)
	}
	if cfg.Framework == nil {
		t.Fatal("expected Framework to be non-nil")
	}
	if cfg.Framework.Version != "2026.03.1" {
		t.Errorf("Version = %q, want 2026.03.1", cfg.Framework.Version)
	}
	if cfg.Framework.Mode != "standalone" {
		t.Errorf("Mode = %q, want standalone", cfg.Framework.Mode)
	}
}

func TestConfigRepo_FrameworkAbsent(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v2.0"
generation = 1

[project]
name = "test-project"
type = "cli"
`
	os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644)

	repo := NewConfigRepo(dir)
	cfg, err := repo.ReadProjectConfig()
	if err != nil {
		t.Fatalf("ReadProjectConfig: %v", err)
	}
	if cfg.Framework != nil {
		t.Errorf("expected Framework to be nil when section absent, got %+v", cfg.Framework)
	}
}

func TestConfigRepo_FrameworkInvalidVersion(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v2.0"
generation = 1

[project]
name = "test-project"
type = "cli"

[framework]
version = "bad-version"
`
	os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644)

	repo := NewConfigRepo(dir)
	_, err := repo.ReadProjectConfig()
	if err == nil {
		t.Fatal("expected error for invalid framework version")
	}
}

func TestConfigRepo_FrameworkMissingVersion(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v2.0"
generation = 1

[project]
name = "test-project"
type = "cli"

[framework]
mode = "standalone"
`
	os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644)

	repo := NewConfigRepo(dir)
	_, err := repo.ReadProjectConfig()
	if err == nil {
		t.Fatal("expected error for missing framework version")
	}
}
