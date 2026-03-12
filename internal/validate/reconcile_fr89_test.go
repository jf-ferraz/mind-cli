package validate

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-89: Missing documents produce WARN-level checks.
func TestReconcileSuite_FR89_MissingDocWarn(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockDirty,
		Missing: []string{"doc:spec/requirements"},
		Stale:   map[string]string{},
		Stats: domain.LockStats{
			Total:   5,
			Missing: 1,
			Clean:   4,
		},
	}

	report := ReconcileSuite(result, false)

	// Find the missing-doc check
	var found bool
	for _, check := range report.Checks {
		if strings.Contains(check.Name, "Document exists") && strings.Contains(check.Name, "doc:spec/requirements") {
			found = true
			if check.Passed {
				t.Error("missing doc check should not pass")
			}
			if check.Level != domain.LevelWarn {
				t.Errorf("missing doc level = %q, want WARN", check.Level)
			}
			if !strings.Contains(check.Message, "declared in mind.toml but not found on disk") {
				t.Errorf("missing doc message = %q, want 'declared in mind.toml' phrasing", check.Message)
			}
		}
	}
	if !found {
		t.Error("expected a 'Document exists' check for the missing document")
	}
}

// FR-89: Missing documents with --strict produce FAIL.
func TestReconcileSuite_FR89_MissingDocStrict(t *testing.T) {
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

	report := ReconcileSuite(result, true)

	// In strict mode, missing should be FAIL
	for _, check := range report.Checks {
		if strings.Contains(check.Name, "Document exists") {
			if check.Level != domain.LevelFail {
				t.Errorf("strict mode: missing doc level = %q, want FAIL", check.Level)
			}
		}
	}
	if report.Failed != 2 {
		t.Errorf("strict mode: Failed = %d, want 2", report.Failed)
	}
	if report.Warnings != 0 {
		t.Errorf("strict mode: Warnings = %d, want 0", report.Warnings)
	}
}

// FR-89: No missing documents produces a passing check.
func TestReconcileSuite_FR89_NoMissing(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockClean,
		Missing: []string{},
		Stale:   map[string]string{},
		Stats:   domain.LockStats{Total: 5, Clean: 5},
	}

	report := ReconcileSuite(result, false)

	var found bool
	for _, check := range report.Checks {
		if check.Name == "No missing documents" {
			found = true
			if !check.Passed {
				t.Error("'No missing documents' check should pass when no docs are missing")
			}
		}
	}
	if !found {
		t.Error("expected 'No missing documents' check when Missing is empty")
	}
}

// FR-89: Suite includes both cycle check and missing check.
func TestReconcileSuite_FR89_CheckOrdering(t *testing.T) {
	result := &domain.ReconcileResult{
		Status:  domain.LockClean,
		Missing: []string{},
		Stale:   map[string]string{},
		Stats:   domain.LockStats{Total: 3, Clean: 3},
	}

	report := ReconcileSuite(result, false)

	if len(report.Checks) < 2 {
		t.Fatalf("expected at least 2 checks (cycle + no-missing), got %d", len(report.Checks))
	}

	if !strings.Contains(report.Checks[0].Name, "cycle") {
		t.Errorf("first check = %q, expected cycle check", report.Checks[0].Name)
	}
	if !strings.Contains(report.Checks[1].Name, "missing") {
		t.Errorf("second check = %q, expected missing documents check", report.Checks[1].Name)
	}
}
