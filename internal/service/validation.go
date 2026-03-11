package service

import (
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/validate"
)

// ValidationService orchestrates validation suite execution.
type ValidationService struct {
	docRepo    repo.DocRepo
	iterRepo   repo.IterationRepo
	briefRepo  repo.BriefRepo
	configRepo repo.ConfigRepo
}

// NewValidationService creates a ValidationService.
func NewValidationService(
	docRepo repo.DocRepo,
	iterRepo repo.IterationRepo,
	briefRepo repo.BriefRepo,
	configRepo repo.ConfigRepo,
) *ValidationService {
	return &ValidationService{
		docRepo:    docRepo,
		iterRepo:   iterRepo,
		briefRepo:  briefRepo,
		configRepo: configRepo,
	}
}

// RunDocs executes the 17-check documentation validation suite.
func (s *ValidationService) RunDocs(projectRoot string, strict bool) domain.ValidationReport {
	ctx := &validate.CheckContext{
		ProjectRoot: projectRoot,
		DocRepo:     s.docRepo,
		IterRepo:    s.iterRepo,
		BriefRepo:   s.briefRepo,
		Strict:      strict,
	}
	suite := validate.DocsSuite()
	return suite.Run(ctx)
}

// RunRefs executes the 11-check cross-reference validation suite.
func (s *ValidationService) RunRefs(projectRoot string) domain.ValidationReport {
	ctx := &validate.CheckContext{
		ProjectRoot: projectRoot,
		DocRepo:     s.docRepo,
		IterRepo:    s.iterRepo,
		BriefRepo:   s.briefRepo,
		ConfigRepo:  s.configRepo,
	}
	suite := validate.RefsSuite()
	return suite.Run(ctx)
}

// RunConfig executes the config validation suite.
func (s *ValidationService) RunConfig(projectRoot string) domain.ValidationReport {
	ctx := &validate.CheckContext{
		ProjectRoot: projectRoot,
		DocRepo:     s.docRepo,
		ConfigRepo:  s.configRepo,
	}
	suite := validate.ConfigSuite()
	return suite.Run(ctx)
}

// RunAll executes all validation suites and returns a unified report.
func (s *ValidationService) RunAll(projectRoot string, strict bool) domain.UnifiedValidationReport {
	docs := s.RunDocs(projectRoot, strict)
	refs := s.RunRefs(projectRoot)
	config := s.RunConfig(projectRoot)

	return domain.UnifiedValidationReport{
		Suites: []domain.ValidationReport{docs, refs, config},
		Summary: domain.UnifiedValidationSummary{
			Total:    docs.Total + refs.Total + config.Total,
			Passed:   docs.Passed + refs.Passed + config.Passed,
			Failed:   docs.Failed + refs.Failed + config.Failed,
			Warnings: docs.Warnings + refs.Warnings + config.Warnings,
		},
	}
}
