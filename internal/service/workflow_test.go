package service

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestWorkflowServiceStatus verifies FR-44: workflow status retrieval.
func TestWorkflowServiceStatus(t *testing.T) {
	t.Run("idle workflow", func(t *testing.T) {
		stateRepo := mem.NewStateRepo()
		iterRepo := mem.NewIterationRepo()
		svc := NewWorkflowService(stateRepo, iterRepo)

		state, err := svc.Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		// nil state means idle
		if state != nil && !state.IsIdle() {
			t.Error("Expected idle workflow")
		}
	})

	t.Run("active workflow", func(t *testing.T) {
		stateRepo := mem.NewStateRepo()
		stateRepo.State = &domain.WorkflowState{
			Type:           domain.TypeNewProject,
			Descriptor:     "core-cli",
			LastAgent:      "architect",
			RemainingChain: []string{"developer", "tester"},
		}
		iterRepo := mem.NewIterationRepo()
		svc := NewWorkflowService(stateRepo, iterRepo)

		state, err := svc.Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if state.IsIdle() {
			t.Error("Expected active workflow, got idle")
		}
		if state.LastAgent != "architect" {
			t.Errorf("LastAgent = %q, want architect", state.LastAgent)
		}
	})
}

// TestWorkflowServiceHistory verifies FR-45: iteration history.
func TestWorkflowServiceHistory(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		stateRepo := mem.NewStateRepo()
		iterRepo := mem.NewIterationRepo()
		svc := NewWorkflowService(stateRepo, iterRepo)

		history, err := svc.History()
		if err != nil {
			t.Fatalf("History() error = %v", err)
		}
		if history.Total != 0 {
			t.Errorf("Total = %d, want 0", history.Total)
		}
	})

	t.Run("multiple iterations", func(t *testing.T) {
		stateRepo := mem.NewStateRepo()
		iterRepo := mem.NewIterationRepo()

		// Add iterations with different states
		iterRepo.Iterations = []domain.Iteration{
			{
				Seq:        1,
				Type:       domain.TypeNewProject,
				Descriptor: "init",
				DirName:    "001-NEW_PROJECT-init",
				Status:     domain.IterComplete,
				Artifacts: []domain.Artifact{
					{Name: "overview.md", Exists: true},
					{Name: "changes.md", Exists: true},
					{Name: "test-summary.md", Exists: true},
					{Name: "validation.md", Exists: true},
					{Name: "retrospective.md", Exists: true},
				},
			},
			{
				Seq:        2,
				Type:       domain.TypeEnhancement,
				Descriptor: "caching",
				DirName:    "002-ENHANCEMENT-caching",
				Status:     domain.IterInProgress,
				Artifacts: []domain.Artifact{
					{Name: "overview.md", Exists: true},
					{Name: "changes.md", Exists: false},
					{Name: "test-summary.md", Exists: false},
					{Name: "validation.md", Exists: false},
					{Name: "retrospective.md", Exists: false},
				},
			},
		}

		svc := NewWorkflowService(stateRepo, iterRepo)
		history, err := svc.History()
		if err != nil {
			t.Fatalf("History() error = %v", err)
		}

		if history.Total != 2 {
			t.Errorf("Total = %d, want 2", history.Total)
		}

		// Verify artifact counts
		for _, iter := range history.Iterations {
			if iter.Artifacts.Expected != 5 {
				t.Errorf("iter %d: Artifacts.Expected = %d, want 5", iter.Seq, iter.Artifacts.Expected)
			}
			if iter.Seq == 1 && iter.Artifacts.Present != 5 {
				t.Errorf("iter 1: Artifacts.Present = %d, want 5", iter.Artifacts.Present)
			}
			if iter.Seq == 2 && iter.Artifacts.Present != 1 {
				t.Errorf("iter 2: Artifacts.Present = %d, want 1", iter.Artifacts.Present)
			}
		}
	})
}
