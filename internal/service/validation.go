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
// If reconcileResult is non-nil, the reconcile suite is included.
func (s *ValidationService) RunAll(projectRoot string, strict bool, reconcileResult ...*domain.ReconcileResult) domain.UnifiedValidationReport {
	docs := s.RunDocs(projectRoot, strict)
	refs := s.RunRefs(projectRoot)
	config := s.RunConfig(projectRoot)

	suites := []domain.ValidationReport{docs, refs, config}

	// Include reconcile suite if a result was provided
	if len(reconcileResult) > 0 && reconcileResult[0] != nil {
		reconcile := validate.ReconcileSuite(reconcileResult[0], strict)
		suites = append(suites, reconcile)
	}

	summary := domain.UnifiedValidationSummary{}
	for _, suite := range suites {
		summary.Total += suite.Total
		summary.Passed += suite.Passed
		summary.Failed += suite.Failed
		summary.Warnings += suite.Warnings
	}

	return domain.UnifiedValidationReport{
		Suites:  suites,
		Summary: summary,
	}
}
