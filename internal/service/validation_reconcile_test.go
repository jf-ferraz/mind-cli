package service

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// FR-42/FR-79: RunAll with reconcile result includes 4 suites.
func TestValidationService_RunAll_WithReconcileResult(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	reconcileResult := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/arch": "dependency changed: doc:spec/req",
		},
		Stats: domain.LockStats{
			Total: 3, Stale: 1, Clean: 2,
		},
	}

	report := svc.RunAll("/test", false, reconcileResult)

	// Should have 4 suites: docs, refs, config, reconcile
	if len(report.Suites) != 4 {
		t.Errorf("Suites count = %d, want 4", len(report.Suites))
	}

	// Verify suite names in order
	expectedNames := []string{"docs", "refs", "config", "reconcile"}
	for i, name := range expectedNames {
		if i < len(report.Suites) && report.Suites[i].Suite != name {
			t.Errorf("Suites[%d].Suite = %q, want %q", i, report.Suites[i].Suite, name)
		}
	}

	// Reconcile suite should have checks
	reconcileSuite := report.Suites[3]
	if reconcileSuite.Total < 1 {
		t.Errorf("reconcile suite Total = %d, want >= 1", reconcileSuite.Total)
	}

	// Summary should include reconcile suite counts
	expectedTotal := 17 + 11 + 12 + reconcileSuite.Total // docs + refs + config + reconcile
	if report.Summary.Total != expectedTotal {
		t.Errorf("Summary.Total = %d, want %d", report.Summary.Total, expectedTotal)
	}
}

// FR-42: RunAll without reconcile result includes only 3 suites.
func TestValidationService_RunAll_WithoutReconcileResult(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	report := svc.RunAll("/test", false)

	if len(report.Suites) != 3 {
		t.Errorf("Suites count = %d, want 3 (without reconcile)", len(report.Suites))
	}
}

// FR-42: RunAll with nil reconcile result includes only 3 suites.
func TestValidationService_RunAll_WithNilReconcileResult(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	report := svc.RunAll("/test", false, nil)

	if len(report.Suites) != 3 {
		t.Errorf("Suites count = %d, want 3 (nil reconcile result)", len(report.Suites))
	}
}

// FR-79: RunAll strict mode promotes stale warnings to failures.
func TestValidationService_RunAll_StrictWithReconcile(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	reconcileResult := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/arch":  "dependency changed",
			"doc:spec/model": "dependency changed",
		},
		Stats: domain.LockStats{
			Total: 3, Stale: 2, Clean: 1,
		},
	}

	normalReport := svc.RunAll("/test", false, reconcileResult)
	strictReport := svc.RunAll("/test", true, reconcileResult)

	// In strict mode, the reconcile suite stale checks become FAIL instead of WARN
	normalReconcile := normalReport.Suites[3]
	strictReconcile := strictReport.Suites[3]

	if strictReconcile.Failed <= normalReconcile.Failed {
		t.Errorf("strict reconcile Failed (%d) should be > normal (%d)",
			strictReconcile.Failed, normalReconcile.Failed)
	}
}

// FR-82: Summary counts are consistent with suite details.
func TestValidationService_RunAll_SummaryCounts(t *testing.T) {
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()

	svc := NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	reconcileResult := &domain.ReconcileResult{
		Status: domain.LockClean,
		Stale:  map[string]string{},
		Stats:  domain.LockStats{Total: 3, Clean: 3},
	}

	report := svc.RunAll("/test", false, reconcileResult)

	// Verify summary totals match sum of suite totals
	totalFromSuites := 0
	passedFromSuites := 0
	failedFromSuites := 0
	warningsFromSuites := 0

	for _, suite := range report.Suites {
		totalFromSuites += suite.Total
		passedFromSuites += suite.Passed
		failedFromSuites += suite.Failed
		warningsFromSuites += suite.Warnings
	}

	if report.Summary.Total != totalFromSuites {
		t.Errorf("Summary.Total = %d, suite sum = %d", report.Summary.Total, totalFromSuites)
	}
	if report.Summary.Passed != passedFromSuites {
		t.Errorf("Summary.Passed = %d, suite sum = %d", report.Summary.Passed, passedFromSuites)
	}
	if report.Summary.Failed != failedFromSuites {
		t.Errorf("Summary.Failed = %d, suite sum = %d", report.Summary.Failed, failedFromSuites)
	}
	if report.Summary.Warnings != warningsFromSuites {
		t.Errorf("Summary.Warnings = %d, suite sum = %d", report.Summary.Warnings, warningsFromSuites)
	}
}
