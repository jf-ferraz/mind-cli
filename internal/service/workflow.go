package service

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// WorkflowService reads workflow state and iteration history.
type WorkflowService struct {
	stateRepo repo.StateRepo
	iterRepo  repo.IterationRepo
}

// NewWorkflowService creates a WorkflowService.
func NewWorkflowService(stateRepo repo.StateRepo, iterRepo repo.IterationRepo) *WorkflowService {
	return &WorkflowService{
		stateRepo: stateRepo,
		iterRepo:  iterRepo,
	}
}

// Status returns the current workflow state.
func (s *WorkflowService) Status() (*domain.WorkflowState, error) {
	return s.stateRepo.ReadWorkflow()
}

// UpdateState persists a new workflow state.
func (s *WorkflowService) UpdateState(state *domain.WorkflowState) error {
	return s.stateRepo.WriteWorkflow(state)
}

// Show returns detailed information about a single iteration by sequence ID or dir name prefix.
func (s *WorkflowService) Show(id string) (*domain.Iteration, error) {
	iterations, err := s.iterRepo.List()
	if err != nil {
		return nil, fmt.Errorf("list iterations: %w", err)
	}
	for i, iter := range iterations {
		if iter.DirName == id || fmt.Sprintf("%03d", iter.Seq) == id {
			return &iterations[i], nil
		}
	}
	return nil, fmt.Errorf("iteration %q not found", id)
}

// History returns all iterations as a summary list.
func (s *WorkflowService) History() (*domain.WorkflowHistory, error) {
	iterations, err := s.iterRepo.List()
	if err != nil {
		return nil, err
	}

	history := &domain.WorkflowHistory{
		Total: len(iterations),
	}

	for _, iter := range iterations {
		present := 0
		for _, a := range iter.Artifacts {
			if a.Exists {
				present++
			}
		}

		summary := domain.IterationSummary{
			Seq:        iter.Seq,
			Type:       iter.Type,
			Descriptor: iter.Descriptor,
			DirName:    iter.DirName,
			Status:     iter.Status,
			CreatedAt:  iter.CreatedAt.Format("2006-01-02"),
			Artifacts: domain.ArtifactCount{
				Present:  present,
				Expected: len(domain.ExpectedArtifacts),
			},
		}
		history.Iterations = append(history.Iterations, summary)
	}

	return history, nil
}
