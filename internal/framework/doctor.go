package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
)

// DoctorCheck represents a single framework doctor diagnostic.
type DoctorCheck struct {
	Check   string
	Status  domain.DiagnosticStatus
	Message string
	Fix     string
}

// RunDoctorChecks performs framework-specific diagnostics.
func RunDoctorChecks(projectRoot string, cfg *domain.Config) []DoctorCheck {
	var checks []DoctorCheck

	globalDir := DefaultGlobalDir()
	lockPath := filepath.Join(globalDir, "framework.lock")

	// Check 1: framework.lock exists
	lock, err := ReadLock(lockPath)
	if err != nil || lock == nil {
		checks = append(checks, DoctorCheck{
			Check:   "Framework installed",
			Status:  domain.DiagWarn,
			Message: "Framework not installed globally",
			Fix:     "Run: mind framework install --source <path-to-.mind/>",
		})
		return checks
	}

	checks = append(checks, DoctorCheck{
		Check:   "Framework installed",
		Status:  domain.DiagPass,
		Message: fmt.Sprintf("Framework v%s installed", lock.Framework.Version),
	})

	// Check 2: version match if project has [framework] section
	if cfg != nil && cfg.Framework != nil {
		if cfg.Framework.Version != lock.Framework.Version {
			checks = append(checks, DoctorCheck{
				Check:   "Framework version match",
				Status:  domain.DiagFail,
				Message: fmt.Sprintf("Project expects v%s, installed v%s", cfg.Framework.Version, lock.Framework.Version),
				Fix:     fmt.Sprintf("Run: mind framework install --source <v%s-source>", cfg.Framework.Version),
			})
		} else {
			checks = append(checks, DoctorCheck{
				Check:   "Framework version match",
				Status:  domain.DiagPass,
				Message: fmt.Sprintf("Project and installed versions match: v%s", lock.Framework.Version),
			})
		}
	}

	// Check 3: drift detection
	driftCount := 0
	for relPath, expectedHash := range lock.Checksums {
		absPath := filepath.Join(globalDir, relPath)
		actualHash, err := hashFile(absPath)
		if err != nil {
			driftCount++
			continue
		}
		if actualHash != expectedHash {
			driftCount++
		}
	}

	if driftCount == 0 {
		checks = append(checks, DoctorCheck{
			Check:   "Framework integrity",
			Status:  domain.DiagPass,
			Message: "No drift detected — all checksums match",
		})
	} else {
		checks = append(checks, DoctorCheck{
			Check:   "Framework integrity",
			Status:  domain.DiagWarn,
			Message: fmt.Sprintf("%d file(s) modified since installation", driftCount),
			Fix:     "Run: mind framework install --force to restore originals",
		})
	}

	// Check 4: project .mind/ directory exists
	mindDir := filepath.Join(projectRoot, ".mind")
	if _, err := os.Stat(mindDir); os.IsNotExist(err) {
		checks = append(checks, DoctorCheck{
			Check:   "Project .mind/ directory",
			Status:  domain.DiagFail,
			Message: ".mind/ directory missing",
			Fix:     "Run: mind init",
		})
	} else {
		checks = append(checks, DoctorCheck{
			Check:   "Project .mind/ directory",
			Status:  domain.DiagPass,
			Message: ".mind/ directory exists",
		})
	}

	return checks
}
