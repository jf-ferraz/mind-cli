package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// TestIterationRepoList verifies BR-5, BR-8: parsing iteration directories
// and computing status from artifact presence.
func TestIterationRepoList(t *testing.T) {
	root := t.TempDir()
	iterDir := filepath.Join(root, "docs", "iterations")

	// Create iteration directories with varying artifacts
	// Complete iteration: all 5 artifacts
	mkIterDir(t, iterDir, "001-NEW_PROJECT-core-cli", domain.ExpectedArtifacts)
	// In-progress iteration: only overview.md (and maybe one other, but not all)
	mkIterDir(t, iterDir, "002-ENHANCEMENT-add-caching", []string{"overview.md", "changes.md"})
	// Non-matching directory (should be skipped)
	os.MkdirAll(filepath.Join(iterDir, "not-an-iteration"), 0755)

	repo := NewIterationRepo(root)
	iterations, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(iterations) != 2 {
		t.Fatalf("List() returned %d iterations, want 2", len(iterations))
	}

	// Should be sorted newest first
	if iterations[0].Seq != 2 {
		t.Errorf("First iteration Seq = %d, want 2 (newest first)", iterations[0].Seq)
	}
	if iterations[1].Seq != 1 {
		t.Errorf("Second iteration Seq = %d, want 1", iterations[1].Seq)
	}

	// Check first iteration (seq=2, incomplete because not all artifacts)
	iter2 := iterations[0]
	if iter2.Type != domain.TypeEnhancement {
		t.Errorf("iter2.Type = %q, want %q", iter2.Type, domain.TypeEnhancement)
	}
	if iter2.Descriptor != "add-caching" {
		t.Errorf("iter2.Descriptor = %q, want %q", iter2.Descriptor, "add-caching")
	}
	if iter2.DirName != "002-ENHANCEMENT-add-caching" {
		t.Errorf("iter2.DirName = %q, want %q", iter2.DirName, "002-ENHANCEMENT-add-caching")
	}

	// Check second iteration (seq=1, complete because all artifacts)
	iter1 := iterations[1]
	if iter1.Type != domain.TypeNewProject {
		t.Errorf("iter1.Type = %q, want %q", iter1.Type, domain.TypeNewProject)
	}
	if iter1.Status != domain.IterComplete {
		t.Errorf("iter1.Status = %q, want %q (all artifacts present)", iter1.Status, domain.IterComplete)
	}

	// Verify BR-7: each iteration has 5 artifacts checked
	for _, iter := range iterations {
		if len(iter.Artifacts) != 5 {
			t.Errorf("iter %d has %d artifacts, want 5", iter.Seq, len(iter.Artifacts))
		}
	}
}

// TestIterationRepoNextSeq verifies BR-6: next seq = max(existing) + 1.
func TestIterationRepoNextSeq(t *testing.T) {
	tests := []struct {
		name    string
		dirs    []string // iteration directory names to create
		wantSeq int
	}{
		{
			name:    "no iterations returns 1",
			dirs:    nil,
			wantSeq: 1,
		},
		{
			name:    "after one iteration returns 2",
			dirs:    []string{"001-NEW_PROJECT-init"},
			wantSeq: 2,
		},
		{
			name:    "gap: 001 and 003 returns 4 (not 2)",
			dirs:    []string{"001-NEW_PROJECT-init", "003-ENHANCEMENT-feature"},
			wantSeq: 4,
		},
		{
			name:    "sequential: 001, 002, 003 returns 4",
			dirs:    []string{"001-NEW_PROJECT-init", "002-BUG_FIX-patch", "003-ENHANCEMENT-feature"},
			wantSeq: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			iterDir := filepath.Join(root, "docs", "iterations")
			os.MkdirAll(iterDir, 0755)

			for _, dir := range tt.dirs {
				os.MkdirAll(filepath.Join(iterDir, dir), 0755)
			}

			repo := NewIterationRepo(root)
			seq, err := repo.NextSeq()
			if err != nil {
				t.Fatalf("NextSeq() error = %v", err)
			}
			if seq != tt.wantSeq {
				t.Errorf("NextSeq() = %d, want %d", seq, tt.wantSeq)
			}
		})
	}
}

// TestIterationRepoListEmpty verifies behavior with no iterations directory.
func TestIterationRepoListEmpty(t *testing.T) {
	root := t.TempDir()
	repo := NewIterationRepo(root)

	iterations, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if iterations != nil {
		t.Errorf("List() = %v, want nil for non-existent dir", iterations)
	}
}

// TestIterationStatusDerivation verifies BR-8: status is derived from artifacts.
func TestIterationStatusDerivation(t *testing.T) {
	tests := []struct {
		name       string
		artifacts  []string // files to create in iteration dir
		wantStatus domain.IterationStatus
	}{
		{
			name:       "all 5 artifacts = complete",
			artifacts:  domain.ExpectedArtifacts,
			wantStatus: domain.IterComplete,
		},
		{
			name:       "some artifacts = incomplete",
			artifacts:  []string{"overview.md", "changes.md"},
			wantStatus: domain.IterIncomplete,
		},
		{
			name:       "no artifacts = in_progress",
			artifacts:  nil,
			wantStatus: domain.IterInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			iterDir := filepath.Join(root, "docs", "iterations")
			mkIterDir(t, iterDir, "001-NEW_PROJECT-test", tt.artifacts)

			repo := NewIterationRepo(root)
			iterations, err := repo.List()
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(iterations) != 1 {
				t.Fatalf("List() returned %d, want 1", len(iterations))
			}
			if iterations[0].Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", iterations[0].Status, tt.wantStatus)
			}
		})
	}
}

func mkIterDir(t *testing.T, iterDir, name string, files []string) {
	t.Helper()
	dirPath := filepath.Join(iterDir, name)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("mkIterDir(%s): %v", name, err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dirPath, f), []byte("# "+f+"\n"), 0644); err != nil {
			t.Fatalf("mkIterDir(%s/%s): %v", name, f, err)
		}
	}
}
