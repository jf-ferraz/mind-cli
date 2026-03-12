package validate

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
)

// ReconcileSuite creates a ValidationReport from a pre-computed ReconcileResult.
// Unlike DocsSuite/RefsSuite/ConfigSuite, this does not use CheckFunc because
// reconciliation is expensive and the result is needed elsewhere (rendering, exit codes).
// The suite projects the result into Check entries for unified reporting.
//
// Checks produced:
//   - Check 1: "No cycle in dependency graph" (PASS -- cycles would have aborted reconciliation)
//   - Check 2+: "Document not stale: {doc ID}" per stale document (WARN, FAIL with strict)
func ReconcileSuite(result *domain.ReconcileResult, strict bool) domain.ValidationReport {
	report := domain.ValidationReport{
		Suite: "reconcile",
	}

	// If result is nil (reconciliation could not run), report a single warning
	if result == nil {
		report.Total = 1
		report.Checks = append(report.Checks, domain.CheckResult{
			ID:      1,
			Name:    "Reconciliation available",
			Level:   domain.LevelWarn,
			Passed:  false,
			Message: "reconciliation could not run (missing config or documents)",
		})
		report.Warnings++
		return report
	}

	checkID := 1

	// Check 1: No cycle (if we got a result, cycles were not detected)
	report.Checks = append(report.Checks, domain.CheckResult{
		ID:     checkID,
		Name:   "No cycle in dependency graph",
		Level:  domain.LevelFail,
		Passed: true,
	})
	report.Passed++
	checkID++

	// Check 2: No missing documents
	if len(result.Missing) == 0 {
		report.Checks = append(report.Checks, domain.CheckResult{
			ID:     checkID,
			Name:   "No missing documents",
			Level:  domain.LevelWarn,
			Passed: true,
		})
		report.Passed++
	} else {
		for _, id := range result.Missing {
			level := domain.LevelWarn
			if strict {
				level = domain.LevelFail
			}
			report.Checks = append(report.Checks, domain.CheckResult{
				ID:      checkID,
				Name:    fmt.Sprintf("Document exists: %s", id),
				Level:   level,
				Passed:  false,
				Message: fmt.Sprintf("declared in mind.toml but not found on disk: %s", id),
			})
			if level == domain.LevelFail {
				report.Failed++
			} else {
				report.Warnings++
			}
			checkID++
		}
	}
	checkID++

	// Check per stale document
	for id, reason := range result.Stale {
		level := domain.LevelWarn
		if strict {
			level = domain.LevelFail
		}

		report.Checks = append(report.Checks, domain.CheckResult{
			ID:      checkID,
			Name:    fmt.Sprintf("Document not stale: %s", id),
			Level:   level,
			Passed:  false,
			Message: reason,
		})
		if level == domain.LevelFail {
			report.Failed++
		} else {
			report.Warnings++
		}
		checkID++
	}

	report.Total = len(report.Checks)
	return report
}
