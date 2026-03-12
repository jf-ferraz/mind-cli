package validate

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// TestSuiteRun verifies Suite.Run() mechanics: counting, pass/fail/warn.
func TestSuiteRun(t *testing.T) {
	suite := &Suite{
		Name: "test",
		Checks: []Check{
			{1, "always pass", domain.LevelFail, func(ctx *CheckContext) (bool, string) {
				return true, ""
			}},
			{2, "always fail", domain.LevelFail, func(ctx *CheckContext) (bool, string) {
				return false, "something failed"
			}},
			{3, "always warn", domain.LevelWarn, func(ctx *CheckContext) (bool, string) {
				return false, "warning"
			}},
		},
	}

	ctx := &CheckContext{}
	report := suite.Run(ctx)

	if report.Suite != "test" {
		t.Errorf("Suite = %q, want test", report.Suite)
	}
	if report.Total != 3 {
		t.Errorf("Total = %d, want 3", report.Total)
	}
	if report.Passed != 1 {
		t.Errorf("Passed = %d, want 1", report.Passed)
	}
	if report.Failed != 1 {
		t.Errorf("Failed = %d, want 1", report.Failed)
	}
	if report.Warnings != 1 {
		t.Errorf("Warnings = %d, want 1", report.Warnings)
	}
	if report.Ok() {
		t.Error("report.Ok() should be false when there are failures")
	}
}

// TestSuiteRunStrict verifies FR-39 and BR-21: strict mode promotes WARN to FAIL.
func TestSuiteRunStrict(t *testing.T) {
	suite := &Suite{
		Name: "test-strict",
		Checks: []Check{
			{1, "pass check", domain.LevelFail, func(ctx *CheckContext) (bool, string) {
				return true, ""
			}},
			{2, "warn check", domain.LevelWarn, func(ctx *CheckContext) (bool, string) {
				return false, "warning becomes failure"
			}},
			{3, "another warn", domain.LevelWarn, func(ctx *CheckContext) (bool, string) {
				return false, "also becomes failure"
			}},
		},
	}

	// Without strict: warnings stay warnings
	ctx := &CheckContext{Strict: false}
	report := suite.Run(ctx)
	if report.Failed != 0 {
		t.Errorf("Non-strict: Failed = %d, want 0", report.Failed)
	}
	if report.Warnings != 2 {
		t.Errorf("Non-strict: Warnings = %d, want 2", report.Warnings)
	}
	if !report.Ok() {
		t.Error("Non-strict: report.Ok() should be true when no FAIL-level failures")
	}

	// With strict: warnings become failures
	ctx = &CheckContext{Strict: true}
	report = suite.Run(ctx)
	if report.Failed != 2 {
		t.Errorf("Strict: Failed = %d, want 2", report.Failed)
	}
	if report.Warnings != 0 {
		t.Errorf("Strict: Warnings = %d, want 0", report.Warnings)
	}
	if report.Ok() {
		t.Error("Strict: report.Ok() should be false when promoted warnings exist")
	}

	// Verify the check level is changed in the result
	for _, check := range report.Checks {
		if check.ID == 2 || check.ID == 3 {
			if check.Level != domain.LevelFail {
				t.Errorf("Strict: Check %d level = %q, want FAIL", check.ID, check.Level)
			}
		}
	}
}

// TestSuiteRunEmpty verifies that running an empty suite produces a valid report.
func TestSuiteRunEmpty(t *testing.T) {
	suite := &Suite{Name: "empty"}
	ctx := &CheckContext{}
	report := suite.Run(ctx)

	if report.Total != 0 {
		t.Errorf("Total = %d, want 0", report.Total)
	}
	if !report.Ok() {
		t.Error("Empty suite should be Ok()")
	}
}

// TestSuiteRunAllPass verifies all checks passing.
func TestSuiteRunAllPass(t *testing.T) {
	suite := &Suite{
		Name: "all-pass",
		Checks: []Check{
			{1, "check1", domain.LevelFail, func(ctx *CheckContext) (bool, string) {
				return true, ""
			}},
			{2, "check2", domain.LevelWarn, func(ctx *CheckContext) (bool, string) {
				return true, ""
			}},
			{3, "check3", domain.LevelInfo, func(ctx *CheckContext) (bool, string) {
				return true, ""
			}},
		},
	}

	ctx := &CheckContext{}
	report := suite.Run(ctx)

	if report.Total != 3 {
		t.Errorf("Total = %d, want 3", report.Total)
	}
	if report.Passed != 3 {
		t.Errorf("Passed = %d, want 3", report.Passed)
	}
	if report.Failed != 0 {
		t.Errorf("Failed = %d, want 0", report.Failed)
	}
	if report.Warnings != 0 {
		t.Errorf("Warnings = %d, want 0", report.Warnings)
	}
	if !report.Ok() {
		t.Error("All-pass suite should be Ok()")
	}
}

// TestCheckResultPreservesMetadata verifies check results carry ID, Name, Level, Message.
func TestCheckResultPreservesMetadata(t *testing.T) {
	suite := &Suite{
		Name: "meta",
		Checks: []Check{
			{42, "specific check", domain.LevelWarn, func(ctx *CheckContext) (bool, string) {
				return false, "detailed message"
			}},
		},
	}

	ctx := &CheckContext{}
	report := suite.Run(ctx)

	if len(report.Checks) != 1 {
		t.Fatalf("Expected 1 check result, got %d", len(report.Checks))
	}

	cr := report.Checks[0]
	if cr.ID != 42 {
		t.Errorf("ID = %d, want 42", cr.ID)
	}
	if cr.Name != "specific check" {
		t.Errorf("Name = %q, want 'specific check'", cr.Name)
	}
	if cr.Level != domain.LevelWarn {
		t.Errorf("Level = %q, want WARN", cr.Level)
	}
	if cr.Passed {
		t.Error("Passed should be false")
	}
	if cr.Message != "detailed message" {
		t.Errorf("Message = %q, want 'detailed message'", cr.Message)
	}
}
