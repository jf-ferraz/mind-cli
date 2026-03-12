package mem

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-145: MemStateRepo.AppendCurrentState records an iteration entry in memory.
func TestMemStateRepo_AppendCurrentState_RecordsEntry(t *testing.T) {
	repo := NewStateRepo()

	iter := &domain.Iteration{
		Seq:        1,
		Type:       domain.TypeEnhancement,
		Descriptor: "add-feature",
		DirName:    "001-ENHANCEMENT-add-feature",
		Status:     domain.IterComplete,
	}

	if err := repo.AppendCurrentState(iter); err != nil {
		t.Fatalf("AppendCurrentState() error = %v", err)
	}

	if len(repo.CurrentStateEntries) != 1 {
		t.Errorf("CurrentStateEntries = %d, want 1", len(repo.CurrentStateEntries))
	}
	if repo.CurrentStateEntries[0] != iter.DirName {
		t.Errorf("CurrentStateEntries[0] = %q, want %q", repo.CurrentStateEntries[0], iter.DirName)
	}
}

// FR-145: MemStateRepo.AppendCurrentState accumulates multiple entries.
func TestMemStateRepo_AppendCurrentState_MultipleEntries(t *testing.T) {
	repo := NewStateRepo()

	iters := []domain.Iteration{
		{Seq: 1, DirName: "001-ENHANCEMENT-alpha"},
		{Seq: 2, DirName: "002-BUG_FIX-beta"},
		{Seq: 3, DirName: "003-REFACTOR-gamma"},
	}

	for _, iter := range iters {
		iter := iter // capture range variable
		if err := repo.AppendCurrentState(&iter); err != nil {
			t.Fatalf("AppendCurrentState() error = %v for %q", err, iter.DirName)
		}
	}

	if len(repo.CurrentStateEntries) != 3 {
		t.Errorf("CurrentStateEntries = %d, want 3", len(repo.CurrentStateEntries))
	}
	for i, want := range []string{"001-ENHANCEMENT-alpha", "002-BUG_FIX-beta", "003-REFACTOR-gamma"} {
		if repo.CurrentStateEntries[i] != want {
			t.Errorf("CurrentStateEntries[%d] = %q, want %q", i, repo.CurrentStateEntries[i], want)
		}
	}
}

// FR-145: MemStateRepo.AppendCurrentState is a no-op for nil iteration (no error).
func TestMemStateRepo_AppendCurrentState_NilIteration(t *testing.T) {
	repo := NewStateRepo()

	if err := repo.AppendCurrentState(nil); err != nil {
		t.Errorf("AppendCurrentState(nil) error = %v, want nil", err)
	}

	if len(repo.CurrentStateEntries) != 0 {
		t.Errorf("CurrentStateEntries = %d, want 0 after nil append", len(repo.CurrentStateEntries))
	}
}

// FR-145: MemStateRepo.ReadWorkflow and WriteWorkflow still work after AppendCurrentState calls.
func TestMemStateRepo_AppendCurrentState_IndependentFromWorkflow(t *testing.T) {
	repo := NewStateRepo()

	// Write a workflow state.
	state := &domain.WorkflowState{
		Type:       domain.TypeBugFix,
		Descriptor: "fix-crash",
	}
	if err := repo.WriteWorkflow(state); err != nil {
		t.Fatalf("WriteWorkflow() error = %v", err)
	}

	// Append a current state entry.
	iter := &domain.Iteration{Seq: 1, DirName: "001-BUG_FIX-fix-crash"}
	if err := repo.AppendCurrentState(iter); err != nil {
		t.Fatalf("AppendCurrentState() error = %v", err)
	}

	// Workflow state should be unchanged.
	got, err := repo.ReadWorkflow()
	if err != nil {
		t.Fatalf("ReadWorkflow() error = %v", err)
	}
	if got == nil {
		t.Fatal("ReadWorkflow() returned nil after AppendCurrentState")
	}
	if got.Descriptor != "fix-crash" {
		t.Errorf("ReadWorkflow().Descriptor = %q, want 'fix-crash'", got.Descriptor)
	}

	// CurrentStateEntries should have exactly 1 entry.
	if len(repo.CurrentStateEntries) != 1 {
		t.Errorf("CurrentStateEntries = %d, want 1", len(repo.CurrentStateEntries))
	}
}
