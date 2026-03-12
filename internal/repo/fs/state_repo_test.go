package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-145: FsStateRepo.AppendCurrentState appends an iteration entry to docs/state/current.md.
func TestFsStateRepo_AppendCurrentState_AppendsEntry(t *testing.T) {
	root := t.TempDir()

	// Create the docs/state directory and current.md.
	stateDir := filepath.Join(root, "docs", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	currentPath := filepath.Join(stateDir, "current.md")
	initialContent := "# Current State\n\n## Recent Changes\n\n(no entries yet)\n"
	if err := os.WriteFile(currentPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	repo := NewStateRepo(root)

	iter := &domain.Iteration{
		Seq:        5,
		Type:       domain.TypeEnhancement,
		Descriptor: "add-json-export",
		DirName:    "005-ENHANCEMENT-add-json-export",
		Status:     domain.IterComplete,
	}

	if err := repo.AppendCurrentState(iter); err != nil {
		t.Fatalf("AppendCurrentState() error = %v", err)
	}

	// Verify current.md was updated.
	data, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("ReadFile current.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "005-ENHANCEMENT-add-json-export") {
		t.Errorf("current.md does not contain iteration dir name:\n%s", content)
	}
}

// FR-145: FsStateRepo.AppendCurrentState creates a Recent Changes section when absent.
func TestFsStateRepo_AppendCurrentState_CreatesSection(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "docs", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	currentPath := filepath.Join(stateDir, "current.md")
	// No "Recent Changes" section.
	if err := os.WriteFile(currentPath, []byte("# Current State\n\nSome content.\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	repo := NewStateRepo(root)

	iter := &domain.Iteration{
		Seq:        3,
		Type:       domain.TypeBugFix,
		Descriptor: "fix-crash",
		DirName:    "003-BUG_FIX-fix-crash",
		Status:     domain.IterComplete,
	}

	if err := repo.AppendCurrentState(iter); err != nil {
		t.Fatalf("AppendCurrentState() error = %v", err)
	}

	data, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Recent Changes") {
		t.Errorf("current.md does not contain 'Recent Changes' section:\n%s", content)
	}
	if !strings.Contains(content, "003-BUG_FIX-fix-crash") {
		t.Errorf("current.md does not contain iteration dir name:\n%s", content)
	}
}

// FR-145: FsStateRepo.AppendCurrentState returns error when current.md does not exist.
func TestFsStateRepo_AppendCurrentState_MissingFile(t *testing.T) {
	root := t.TempDir()
	// Do NOT create docs/state/current.md.
	repo := NewStateRepo(root)

	iter := &domain.Iteration{
		Seq:     1,
		DirName: "001-ENHANCEMENT-test",
	}

	err := repo.AppendCurrentState(iter)
	if err == nil {
		t.Error("AppendCurrentState() should return error when current.md does not exist")
	}
}

// FR-145: FsStateRepo.AppendCurrentState returns error for nil iteration.
func TestFsStateRepo_AppendCurrentState_NilIteration(t *testing.T) {
	root := t.TempDir()
	repo := NewStateRepo(root)

	err := repo.AppendCurrentState(nil)
	if err == nil {
		t.Error("AppendCurrentState(nil) should return error")
	}
}
