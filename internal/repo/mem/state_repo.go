package mem

import "github.com/jf-ferraz/mind-cli/domain"

// StateRepo is an in-memory implementation of repo.StateRepo for testing.
type StateRepo struct {
	State *domain.WorkflowState
}

// NewStateRepo creates an in-memory StateRepo.
func NewStateRepo() *StateRepo {
	return &StateRepo{}
}

// ReadWorkflow returns the stored workflow state.
func (r *StateRepo) ReadWorkflow() (*domain.WorkflowState, error) {
	return r.State, nil
}
