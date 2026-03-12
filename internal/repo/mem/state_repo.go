package mem

import "github.com/jf-ferraz/mind-cli/domain"

// StateRepo is an in-memory implementation of repo.StateRepo for testing.
type StateRepo struct {
	State                *domain.WorkflowState
	CurrentStateEntries  []string
}

// NewStateRepo creates an in-memory StateRepo.
func NewStateRepo() *StateRepo {
	return &StateRepo{}
}

// ReadWorkflow returns the stored workflow state.
func (r *StateRepo) ReadWorkflow() (*domain.WorkflowState, error) {
	return r.State, nil
}

// WriteWorkflow stores the workflow state in memory.
func (r *StateRepo) WriteWorkflow(state *domain.WorkflowState) error {
	r.State = state
	return nil
}

// AppendCurrentState records the iteration entry in memory (no-op for filesystem).
func (r *StateRepo) AppendCurrentState(iter *domain.Iteration) error {
	if iter == nil {
		return nil
	}
	r.CurrentStateEntries = append(r.CurrentStateEntries, iter.DirName)
	return nil
}
