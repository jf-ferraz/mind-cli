package service

import (
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
