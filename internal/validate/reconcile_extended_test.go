package validate

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-79: ReconcileSuite with missing documents.
func TestReconcileSuite_MissingDocuments(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockDirty,
		Missing: []string{"doc:spec/requirements", "doc:spec/architecture"},
		Stale:   map[string]string{},
		Stats: domain.LockStats{
			Total:   5,
			Missing: 2,
			Clean:   3,
		},
	}

	report := ReconcileSuite(result, false)

	// 1 cycle check + 2 missing document checks = 3
	if report.Total != 3 {
		t.Errorf("Total = %d, want 3 (cycle + 2 missing)", report.Total)
	}
	if report.Passed != 1 {
		t.Errorf("Passed = %d, want 1", report.Passed)
	}
	if report.Warnings != 2 {
		t.Errorf("Warnings = %d, want 2", report.Warnings)
	}
}

// FR-79: Mixed stale and missing documents.
func TestReconcileSuite_MixedStaleAndMissing(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockStale,
		Missing: []string{"doc:spec/missing1"},
		Stale: map[string]string{
			"doc:spec/arch":   "dependency changed: doc:spec/req",
			"doc:spec/design": "dependency changed: doc:spec/req",
		},
		Stats: domain.LockStats{
			Total:   5,
			Missing: 1,
			Stale:   2,
			Clean:   2,
		},
	}

	report := ReconcileSuite(result, false)

	// 1 cycle + 1 missing doc check + 2 stale = 4
	if report.Total != 4 {
		t.Errorf("Total = %d, want 4", report.Total)
	}
	// 1 missing (WARN) + 2 stale (WARN) = 3
	if report.Warnings != 3 {
		t.Errorf("Warnings = %d, want 3 (1 missing + 2 stale docs are WARN)", report.Warnings)
	}
}

// FR-79: Strict mode promotes stale WARN to FAIL.
func TestReconcileSuite_StrictPromotesWarnToFail(t *testing.T) {
	result := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/arch":   "dependency changed",
			"doc:spec/design": "dependency changed",
			"doc:spec/model":  "dependency changed",
		},
		Stats: domain.LockStats{
			Total: 5,
			Stale: 3,
			Clean: 2,
		},
	}

	// Without strict: 1 cycle + 1 no-missing + 3 stale = 5
	normalReport := ReconcileSuite(result, false)
	if normalReport.Failed != 0 {
		t.Errorf("normal mode Failed = %d, want 0", normalReport.Failed)
	}
	if normalReport.Warnings != 3 {
		t.Errorf("normal mode Warnings = %d, want 3", normalReport.Warnings)
	}

	// With strict: 1 cycle + 1 no-missing + 3 stale(FAIL) = 5
	strictReport := ReconcileSuite(result, true)
	if strictReport.Failed != 3 {
		t.Errorf("strict mode Failed = %d, want 3", strictReport.Failed)
	}
	if strictReport.Warnings != 0 {
		t.Errorf("strict mode Warnings = %d, want 0", strictReport.Warnings)
	}
}

// Suite name is always "reconcile".
func TestReconcileSuite_Name(t *testing.T) {
	tests := []struct {
		name   string
		result *domain.ReconcileResult
	}{
		{"nil result", nil},
		{"empty result", &domain.ReconcileResult{Stale: map[string]string{}}},
		{"stale result", &domain.ReconcileResult{
			Stale: map[string]string{"doc:spec/a": "reason"},
			Stats: domain.LockStats{Total: 1, Stale: 1},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := ReconcileSuite(tt.result, false)
			if report.Suite != "reconcile" {
				t.Errorf("Suite = %q, want reconcile", report.Suite)
			}
		})
	}
}

// Cycle check always passes when we have a result (cycles abort before suite construction).
func TestReconcileSuite_CycleCheckAlwaysPasses(t *testing.T) {
	result := &domain.ReconcileResult{
		Status: domain.LockClean,
		Stale:  map[string]string{},
		Stats:  domain.LockStats{Total: 3, Clean: 3},
	}

	report := ReconcileSuite(result, false)

	if len(report.Checks) < 1 {
		t.Fatal("expected at least 1 check")
	}

	cycleCheck := report.Checks[0]
	if !strings.Contains(cycleCheck.Name, "cycle") {
		t.Errorf("first check name = %q, want to contain 'cycle'", cycleCheck.Name)
	}
	if !cycleCheck.Passed {
		t.Error("cycle check should pass when reconciliation succeeds")
	}
}

// Check reason messages appear in check output.
func TestReconcileSuite_StaleReasonInMessage(t *testing.T) {
	result := &domain.ReconcileResult{
		Status: domain.LockStale,
		Stale: map[string]string{
			"doc:spec/arch": "dependency changed: doc:spec/req (prerequisite changed)",
		},
		Stats: domain.LockStats{Total: 2, Stale: 1, Clean: 1},
	}

	report := ReconcileSuite(result, false)

	found := false
	for _, check := range report.Checks {
		if strings.Contains(check.Name, "doc:spec/arch") {
			found = true
			if !strings.Contains(check.Message, "prerequisite changed") {
				t.Errorf("stale check message = %q, want to contain reason", check.Message)
			}
		}
	}
	if !found {
		t.Error("expected a check for doc:spec/arch")
	}
}
