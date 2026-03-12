package validate

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestReconcileSuite_NilResult(t *testing.T) {
	report := ReconcileSuite(nil, false)

	if report.Suite != "reconcile" {
		t.Errorf("Suite = %q, want reconcile", report.Suite)
	}
	if report.Total != 1 {
		t.Errorf("Total = %d, want 1", report.Total)
	}
	if report.Warnings != 1 {
		t.Errorf("Warnings = %d, want 1", report.Warnings)
	}
}

func TestReconcileSuite_Clean(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockClean,
		Changed: []string{},
		Stale:   map[string]string{},
		Missing: []string{},
		Stats: domain.LockStats{
			Total: 3,
			Clean: 3,
		},
	}

	report := ReconcileSuite(result, false)

	// 1 cycle check + 1 "no missing documents" check = 2
	if report.Total != 2 {
		t.Errorf("Total = %d, want 2 (cycle + no-missing)", report.Total)
	}
	if report.Passed != 2 {
		t.Errorf("Passed = %d, want 2", report.Passed)
	}
	if report.Failed != 0 {
		t.Errorf("Failed = %d, want 0", report.Failed)
	}
}

func TestReconcileSuite_Stale(t *testing.T) {
	result := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/architecture": "dependency changed: doc:spec/requirements",
			"doc:spec/design":       "dependency changed: doc:spec/requirements",
		},
		Stats: domain.LockStats{
			Total: 3,
			Stale: 2,
			Clean: 1,
		},
	}

	report := ReconcileSuite(result, false)

	// 1 cycle check + 1 no-missing check + 2 stale checks = 4
	if report.Total != 4 {
		t.Errorf("Total = %d, want 4", report.Total)
	}
	if report.Passed != 2 {
		t.Errorf("Passed = %d, want 2", report.Passed)
	}
	if report.Warnings != 2 {
		t.Errorf("Warnings = %d, want 2", report.Warnings)
	}
	if report.Failed != 0 {
		t.Errorf("Failed = %d, want 0 (stale is WARN without strict)", report.Failed)
	}
}

func TestReconcileSuite_StaleStrict(t *testing.T) {
	result := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/architecture": "dependency changed",
		},
		Stats: domain.LockStats{
			Total: 2,
			Stale: 1,
			Clean: 1,
		},
	}

	report := ReconcileSuite(result, true)

	// 1 cycle + 1 no-missing + 1 stale = 3
	if report.Total != 3 {
		t.Errorf("Total = %d, want 3", report.Total)
	}
	if report.Failed != 1 {
		t.Errorf("Failed = %d, want 1 (strict promotes WARN to FAIL)", report.Failed)
	}
	if report.Warnings != 0 {
		t.Errorf("Warnings = %d, want 0 (strict mode)", report.Warnings)
	}
}
