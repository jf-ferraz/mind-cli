package domain

import "testing"

// TestWorkflowStateIsIdle verifies BR-9: idle detection.
func TestWorkflowStateIsIdle(t *testing.T) {
	tests := []struct {
		name  string
		state *WorkflowState
		want  bool
	}{
		// BR-9: nil state is idle
		{name: "nil state is idle", state: nil, want: true},
		// BR-9: empty Type is idle
		{name: "empty type is idle", state: &WorkflowState{Type: ""}, want: true},
		// BR-9: non-empty Type is not idle
		{name: "NEW_PROJECT is not idle", state: &WorkflowState{Type: TypeNewProject}, want: false},
		{name: "BUG_FIX is not idle", state: &WorkflowState{Type: TypeBugFix}, want: false},
		{name: "ENHANCEMENT is not idle", state: &WorkflowState{Type: TypeEnhancement}, want: false},
		{name: "REFACTOR is not idle", state: &WorkflowState{Type: TypeRefactor}, want: false},
		{name: "COMPLEX_NEW is not idle", state: &WorkflowState{Type: TypeComplexNew}, want: false},
		// State with fields but empty Type is idle
		{
			name: "fields set but empty type is idle",
			state: &WorkflowState{
				Type:       "",
				Descriptor: "leftover",
				Branch:     "feature/old",
			},
			want: true,
		},
		// Full running workflow
		{
			name: "full running workflow",
			state: &WorkflowState{
				Type:           TypeNewProject,
				Descriptor:     "core-cli",
				IterationPath:  "docs/iterations/001-NEW_PROJECT-core-cli",
				Branch:         "feature/core-cli",
				LastAgent:      "architect",
				RemainingChain: []string{"developer", "tester", "reviewer"},
				Session:        1,
				TotalSessions:  5,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.IsIdle()
			if got != tt.want {
				t.Errorf("IsIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}
