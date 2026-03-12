package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// FR-91: DocRepo.Search performs case-insensitive substring search.
func TestDocRepo_Search_BasicMatch(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs", "spec")
	os.MkdirAll(docsDir, 0o755)

	os.WriteFile(filepath.Join(docsDir, "architecture.md"), []byte("# Architecture\n\nThis document describes the authentication layer.\n"), 0o644)
	os.WriteFile(filepath.Join(docsDir, "requirements.md"), []byte("# Requirements\n\nThe system must handle authentication requests.\n"), 0o644)

	repo := NewDocRepo(root)
	results, err := repo.Search("authentication")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.TotalMatches != 2 {
		t.Errorf("TotalMatches = %d, want 2", results.TotalMatches)
	}
	if results.FilesMatched != 2 {
		t.Errorf("FilesMatched = %d, want 2", results.FilesMatched)
	}
}

// FR-91: Search is case-insensitive.
func TestDocRepo_Search_CaseInsensitive(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs", "spec")
	os.MkdirAll(docsDir, 0o755)

	os.WriteFile(filepath.Join(docsDir, "test.md"), []byte("# Test\n\nAuthentication is important.\nAUTHENTICATION layer.\nauthentication module.\n"), 0o644)

	repo := NewDocRepo(root)
	results, err := repo.Search("authentication")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.TotalMatches != 3 {
		t.Errorf("TotalMatches = %d, want 3 (case-insensitive match)", results.TotalMatches)
	}
}

// FR-91: Search provides context lines.
func TestDocRepo_Search_ContextLines(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs", "spec")
	os.MkdirAll(docsDir, 0o755)

	content := "Line 1\nLine 2\nTarget line\nLine 4\nLine 5\n"
	os.WriteFile(filepath.Join(docsDir, "context.md"), []byte(content), 0o644)

	repo := NewDocRepo(root)
	results, err := repo.Search("Target")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.TotalMatches != 1 {
		t.Fatalf("TotalMatches = %d, want 1", results.TotalMatches)
	}

	match := results.Results[0].Matches[0]
	if match.Line != 3 {
		t.Errorf("match.Line = %d, want 3", match.Line)
	}
	if match.ContextBefore != "Line 2" {
		t.Errorf("ContextBefore = %q, want 'Line 2'", match.ContextBefore)
	}
	if match.ContextAfter != "Line 4" {
		t.Errorf("ContextAfter = %q, want 'Line 4'", match.ContextAfter)
	}
}

// FR-91: Search ignores non-.md files.
func TestDocRepo_Search_IgnoresNonMarkdown(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs", "spec")
	os.MkdirAll(docsDir, 0o755)

	os.WriteFile(filepath.Join(docsDir, "test.md"), []byte("authentication\n"), 0o644)
	os.WriteFile(filepath.Join(docsDir, "test.txt"), []byte("authentication\n"), 0o644)
	os.WriteFile(filepath.Join(docsDir, "test.go"), []byte("authentication\n"), 0o644)

	repo := NewDocRepo(root)
	results, err := repo.Search("authentication")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.FilesMatched != 1 {
		t.Errorf("FilesMatched = %d, want 1 (only .md files)", results.FilesMatched)
	}
}

// FR-91: Search with no matches returns empty results.
func TestDocRepo_Search_NoMatches(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs", "spec")
	os.MkdirAll(docsDir, 0o755)

	os.WriteFile(filepath.Join(docsDir, "test.md"), []byte("# Test\n\nNo matching content here.\n"), 0o644)

	repo := NewDocRepo(root)
	results, err := repo.Search("nonexistent-query-xyz")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.TotalMatches != 0 {
		t.Errorf("TotalMatches = %d, want 0", results.TotalMatches)
	}
	if results.FilesMatched != 0 {
		t.Errorf("FilesMatched = %d, want 0", results.FilesMatched)
	}
}

// FR-91: Search preserves query in results.
func TestDocRepo_Search_QueryPreserved(t *testing.T) {
	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	os.MkdirAll(docsDir, 0o755)

	repo := NewDocRepo(root)
	results, err := repo.Search("my-query")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if results.Query != "my-query" {
		t.Errorf("Query = %q, want 'my-query'", results.Query)
	}
}
