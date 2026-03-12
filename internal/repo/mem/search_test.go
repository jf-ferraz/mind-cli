package mem

import (
	"testing"
)

// FR-91: In-memory DocRepo.Search returns matching results without filesystem access.
func TestDocRepo_Search_InMemory(t *testing.T) {
	repo := NewDocRepo()
	repo.Files["docs/spec/architecture.md"] = []byte("# Architecture\n\nAuthentication layer design.\n")
	repo.Files["docs/spec/requirements.md"] = []byte("# Requirements\n\nMust handle authentication.\n")
	repo.Files["docs/state/current.md"] = []byte("# Current State\n\nNo auth content here.\n")

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

// FR-91: In-memory search is case-insensitive.
func TestDocRepo_Search_CaseInsensitive(t *testing.T) {
	repo := NewDocRepo()
	repo.Files["docs/spec/test.md"] = []byte("AUTHENTICATION\nauthentication\nAuthentication\n")

	results, err := repo.Search("authentication")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if results.TotalMatches != 3 {
		t.Errorf("TotalMatches = %d, want 3", results.TotalMatches)
	}
}

// FR-91: In-memory search ignores non-.md files.
func TestDocRepo_Search_IgnoresNonMarkdown(t *testing.T) {
	repo := NewDocRepo()
	repo.Files["docs/spec/test.md"] = []byte("authentication\n")
	repo.Files["docs/spec/test.txt"] = []byte("authentication\n")
	repo.Files["docs/spec/test.go"] = []byte("authentication\n")

	results, err := repo.Search("authentication")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if results.FilesMatched != 1 {
		t.Errorf("FilesMatched = %d, want 1", results.FilesMatched)
	}
}

// FR-91: In-memory search provides context lines.
func TestDocRepo_Search_ContextLines(t *testing.T) {
	repo := NewDocRepo()
	repo.Files["docs/spec/ctx.md"] = []byte("Before line\nTarget line\nAfter line\n")

	results, err := repo.Search("Target")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if results.TotalMatches != 1 {
		t.Fatalf("TotalMatches = %d, want 1", results.TotalMatches)
	}

	match := results.Results[0].Matches[0]
	if match.ContextBefore != "Before line" {
		t.Errorf("ContextBefore = %q, want 'Before line'", match.ContextBefore)
	}
	if match.ContextAfter != "After line" {
		t.Errorf("ContextAfter = %q, want 'After line'", match.ContextAfter)
	}
}

// In-memory QualityRepo returns empty when no entries set.
func TestQualityRepo_ReadLog_Empty(t *testing.T) {
	repo := NewQualityRepo()
	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0", len(entries))
	}
}
