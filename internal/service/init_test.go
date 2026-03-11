package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// TestInitService verifies FR-14 through FR-19.
func TestInitService(t *testing.T) {
	t.Run("FR-14: creates full project structure", func(t *testing.T) {
		root := t.TempDir()
		svc := NewInitService()

		result, err := svc.Init(root, "test-project", false, false)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Verify directories exist
		expectedDirs := []string{
			".mind",
			"docs/spec",
			"docs/blueprints",
			"docs/state",
			"docs/iterations",
			"docs/knowledge",
			".claude",
		}
		for _, dir := range expectedDirs {
			info, err := os.Stat(filepath.Join(root, dir))
			if err != nil {
				t.Errorf("Expected directory %q to exist: %v", dir, err)
				continue
			}
			if !info.IsDir() {
				t.Errorf("%q should be a directory", dir)
			}
		}

		// Verify stub files exist
		expectedFiles := []string{
			"docs/spec/project-brief.md",
			"docs/spec/requirements.md",
			"docs/spec/architecture.md",
			"docs/spec/domain-model.md",
			"docs/state/current.md",
			"docs/state/workflow.md",
			"docs/blueprints/INDEX.md",
			"docs/knowledge/glossary.md",
			"mind.toml",
			".claude/CLAUDE.md",
		}
		for _, file := range expectedFiles {
			if _, err := os.Stat(filepath.Join(root, file)); err != nil {
				t.Errorf("Expected file %q to exist: %v", file, err)
			}
		}

		// Verify result
		if result.ProjectName != "test-project" {
			t.Errorf("ProjectName = %q, want test-project", result.ProjectName)
		}
		if result.Root != root {
			t.Errorf("Root = %q, want %q", result.Root, root)
		}
		if len(result.FilesCreated) == 0 {
			t.Error("FilesCreated should not be empty")
		}
	})

	t.Run("FR-15: creates .claude/CLAUDE.md adapter", func(t *testing.T) {
		root := t.TempDir()
		svc := NewInitService()

		_, err := svc.Init(root, "test", false, false)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		claudePath := filepath.Join(root, ".claude", "CLAUDE.md")
		content, err := os.ReadFile(claudePath)
		if err != nil {
			t.Fatalf("Read .claude/CLAUDE.md: %v", err)
		}
		if len(content) == 0 {
			t.Error(".claude/CLAUDE.md should not be empty")
		}
	})

	t.Run("FR-16: --name flag overrides directory name", func(t *testing.T) {
		root := t.TempDir()
		svc := NewInitService()

		result, err := svc.Init(root, "my-service", false, false)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		if result.ProjectName != "my-service" {
			t.Errorf("ProjectName = %q, want my-service", result.ProjectName)
		}
	})

	t.Run("FR-16: falls back to directory name", func(t *testing.T) {
		root := t.TempDir()
		svc := NewInitService()

		result, err := svc.Init(root, "", false, false)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		expected := filepath.Base(root)
		if result.ProjectName != expected {
			t.Errorf("ProjectName = %q, want %q (dir name)", result.ProjectName, expected)
		}
	})

	t.Run("FR-17: --with-github creates .github/agents/", func(t *testing.T) {
		root := t.TempDir()
		svc := NewInitService()

		result, err := svc.Init(root, "test", true, false)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		agentsDir := filepath.Join(root, ".github", "agents")
		info, err := os.Stat(agentsDir)
		if err != nil {
			t.Fatalf(".github/agents/ should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error(".github/agents/ should be a directory")
		}

		// Should have .gitkeep
		if _, err := os.Stat(filepath.Join(agentsDir, ".gitkeep")); err != nil {
			t.Error(".github/agents/.gitkeep should exist")
		}

		// Result should mention it
		found := false
		for _, f := range result.FilesCreated {
			if f == ".github/agents/.gitkeep" {
				found = true
				break
			}
		}
		if !found {
			t.Error("FilesCreated should include .github/agents/.gitkeep")
		}
	})

	t.Run("FR-18: --from-existing preserves existing files", func(t *testing.T) {
		root := t.TempDir()

		// Create existing file first
		os.MkdirAll(filepath.Join(root, "docs", "spec"), 0755)
		existingContent := "# My Custom Brief\n\nReal content here.\n"
		os.WriteFile(filepath.Join(root, "docs", "spec", "project-brief.md"), []byte(existingContent), 0644)

		svc := NewInitService()
		result, err := svc.Init(root, "test", false, true)
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Existing file should be preserved
		content, _ := os.ReadFile(filepath.Join(root, "docs", "spec", "project-brief.md"))
		if string(content) != existingContent {
			t.Error("Existing file should be preserved, not overwritten")
		}

		if result.FromExisting != true {
			t.Error("FromExisting should be true")
		}

		// Should report preserved files
		found := false
		for _, p := range result.ExistingPreserved {
			if p == "docs/spec/project-brief.md" {
				found = true
				break
			}
		}
		if !found {
			t.Error("ExistingPreserved should include the existing brief")
		}
	})

	t.Run("FR-19: abort if .mind/ already exists", func(t *testing.T) {
		root := t.TempDir()
		os.MkdirAll(filepath.Join(root, ".mind"), 0755)

		svc := NewInitService()
		_, err := svc.Init(root, "test", false, false)
		if err == nil {
			t.Fatal("Init() should fail when .mind/ exists")
		}
		if !errors.Is(err, domain.ErrAlreadyInitialized) {
			t.Errorf("error should be ErrAlreadyInitialized, got %v", err)
		}
	})
}
