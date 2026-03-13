package globalconfig

import (
"os"
"path/filepath"
"testing"
)

// --- Config tests (P1-T13/T14) ---

func TestRead_DefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	cfg, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.DefaultMode != "standalone" {
		t.Errorf("expected default_mode=standalone, got %s", cfg.DefaultMode)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log_level=info, got %s", cfg.LogLevel)
	}
}

func TestWriteAndRead_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	cfg := &GlobalConfig{
		Editor:      "nvim",
		DefaultMode: "thin",
		LogLevel:    "debug",
	}
	if err := repo.Write(cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	loaded, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if loaded.Editor != "nvim" {
		t.Errorf("editor: got %s, want nvim", loaded.Editor)
	}
	if loaded.DefaultMode != "thin" {
		t.Errorf("default_mode: got %s, want thin", loaded.DefaultMode)
	}
	if loaded.LogLevel != "debug" {
		t.Errorf("log_level: got %s, want debug", loaded.LogLevel)
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &GlobalConfig{LogLevel: "verbose"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid log_level")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	cfg := &GlobalConfig{DefaultMode: "hybrid"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid default_mode")
	}
}

func TestWrite_RejectsInvalid(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	cfg := &GlobalConfig{LogLevel: "badlevel"}
	if err := repo.Write(cfg); err == nil {
		t.Error("expected Write to fail with invalid config")
	}
}

func TestShowConfig(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	text, err := repo.ShowConfig()
	if err != nil {
		t.Fatalf("ShowConfig: %v", err)
	}
	if text == "" {
		t.Error("expected non-empty TOML output")
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	if err := repo.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig: %v", err)
	}
}

func TestEnsureExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	if err := repo.EnsureExists(); err != nil {
		t.Fatalf("EnsureExists: %v", err)
	}
	if _, err := os.Stat(repo.ConfigPath()); err != nil {
		t.Errorf("config.toml not created: %v", err)
	}
	// Second call is a no-op.
	if err := repo.EnsureExists(); err != nil {
		t.Fatalf("second EnsureExists: %v", err)
	}
}

func TestEditorCommand_Fallback(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	t.Setenv("EDITOR", "")
	editor, err := repo.EditorCommand()
	if err != nil {
		t.Fatalf("EditorCommand: %v", err)
	}
	if editor != "vi" {
		t.Errorf("expected vi fallback, got %s", editor)
	}
}

func TestEditorCommand_FromEnv(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	t.Setenv("EDITOR", "code")
	editor, err := repo.EditorCommand()
	if err != nil {
		t.Fatalf("EditorCommand: %v", err)
	}
	if editor != "code" {
		t.Errorf("expected code, got %s", editor)
	}
}

func TestEditorCommand_FromConfig(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	repo.Write(&GlobalConfig{Editor: "emacs", DefaultMode: "standalone", LogLevel: "info"})
	t.Setenv("EDITOR", "nvim")
	editor, err := repo.EditorCommand()
	if err != nil {
		t.Fatalf("EditorCommand: %v", err)
	}
	if editor != "emacs" {
		t.Errorf("expected emacs (config takes priority), got %s", editor)
	}
}

// --- Registry tests (P1-T16/T17) ---

func TestRegistryAdd_And_List(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	if err := repo.RegistryAdd("my-project", projDir); err != nil {
		t.Fatalf("RegistryAdd: %v", err)
	}
	items, err := repo.RegistryList()
	if err != nil {
		t.Fatalf("RegistryList: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Alias != "my-project" {
		t.Errorf("alias: got %s, want my-project", items[0].Alias)
	}
}

func TestRegistryAdd_InvalidAlias(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	if err := repo.RegistryAdd("MyProject", projDir); err == nil {
		t.Error("expected error for non-kebab-case alias")
	}
}

func TestRegistryAdd_NonexistentPath(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	if err := repo.RegistryAdd("test", "/nonexistent/path/abc"); err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestRegistryAdd_Duplicate(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	if err := repo.RegistryAdd("test", projDir); err == nil {
		t.Error("expected error for duplicate alias")
	}
}

func TestRegistryRemove(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	if err := repo.RegistryRemove("test"); err != nil {
		t.Fatalf("RegistryRemove: %v", err)
	}
	items, _ := repo.RegistryList()
	if len(items) != 0 {
		t.Errorf("expected 0 items after remove, got %d", len(items))
	}
}

func TestRegistryRemove_NotFound(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	if err := repo.RegistryRemove("nonexistent"); err == nil {
		t.Error("expected error for nonexistent alias")
	}
}

func TestRegistryResolve(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	resolved, err := repo.RegistryResolve("@test")
	if err != nil {
		t.Fatalf("RegistryResolve: %v", err)
	}
	abs, _ := filepath.Abs(projDir)
	if resolved != abs {
		t.Errorf("resolved: got %s, want %s", resolved, abs)
	}
}

func TestRegistryResolve_WithoutAtSign(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	resolved, err := repo.RegistryResolve("test")
	if err != nil {
		t.Fatalf("RegistryResolve: %v", err)
	}
	abs, _ := filepath.Abs(projDir)
	if resolved != abs {
		t.Errorf("resolved: got %s, want %s", resolved, abs)
	}
}

func TestRegistryResolve_NotFound(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	_, err := repo.RegistryResolve("@missing")
	if err == nil {
		t.Error("expected error for missing alias")
	}
}

func TestRegistryCheck_AllExist(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	results, err := repo.RegistryCheck()
	if err != nil {
		t.Fatalf("RegistryCheck: %v", err)
	}
	if len(results) != 1 || !results[0].Exists {
		t.Errorf("expected path to exist: %+v", results)
	}
}

func TestRegistryCheck_MissingPath(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)
	projDir := t.TempDir()
	repo.RegistryAdd("test", projDir)
	os.RemoveAll(projDir)
	results, err := repo.RegistryCheck()
	if err != nil {
		t.Fatalf("RegistryCheck: %v", err)
	}
	if len(results) != 1 || results[0].Exists {
		t.Errorf("expected path to not exist: %+v", results)
	}
	if results[0].Error == "" {
		t.Error("expected error message for missing path")
	}
}

// --- Integration: combined config + registry workflow ---

func TestIntegration_ConfigAndRegistry(t *testing.T) {
	dir := t.TempDir()
	repo := NewGlobalConfigRepo(dir)

	// Step 1: Create default config.
	if err := repo.EnsureExists(); err != nil {
		t.Fatalf("EnsureExists: %v", err)
	}

	// Step 2: Show config.
	text, err := repo.ShowConfig()
	if err != nil {
		t.Fatalf("ShowConfig: %v", err)
	}
	if text == "" {
		t.Error("expected non-empty config output")
	}

	// Step 3: Validate config.
	if err := repo.ValidateConfig(); err != nil {
		t.Fatalf("ValidateConfig: %v", err)
	}

	// Step 4: Add projects.
	proj1 := t.TempDir()
	proj2 := t.TempDir()
	if err := repo.RegistryAdd("project-a", proj1); err != nil {
		t.Fatalf("RegistryAdd project-a: %v", err)
	}
	if err := repo.RegistryAdd("project-b", proj2); err != nil {
		t.Fatalf("RegistryAdd project-b: %v", err)
	}

	// Step 5: List projects.
	items, err := repo.RegistryList()
	if err != nil {
		t.Fatalf("RegistryList: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Step 6: Resolve alias.
	resolved, err := repo.RegistryResolve("@project-a")
	if err != nil {
		t.Fatalf("RegistryResolve: %v", err)
	}
	abs1, _ := filepath.Abs(proj1)
	if resolved != abs1 {
		t.Errorf("resolved: got %s, want %s", resolved, abs1)
	}

	// Step 7: Check all paths.
	results, err := repo.RegistryCheck()
	if err != nil {
		t.Fatalf("RegistryCheck: %v", err)
	}
	for _, r := range results {
		if !r.Exists {
			t.Errorf("path %s should exist", r.Path)
		}
	}

	// Step 8: Remove project.
	if err := repo.RegistryRemove("project-b"); err != nil {
		t.Fatalf("RegistryRemove: %v", err)
	}
	items, _ = repo.RegistryList()
	if len(items) != 1 {
		t.Errorf("expected 1 item after remove, got %d", len(items))
	}
}
