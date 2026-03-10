package validate

import (
	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// CheckFunc receives the project context and returns pass/fail + message.
type CheckFunc func(ctx *CheckContext) (bool, string)

// Check is a single validation check.
type Check struct {
	ID    int
	Name  string
	Level domain.CheckLevel
	Fn    CheckFunc
}

// CheckContext provides everything a check might need.
type CheckContext struct {
	ProjectRoot string
	DocRepo     repo.DocRepo
	IterRepo    repo.IterationRepo
	BriefRepo   repo.BriefRepo
	Strict      bool
}

// Suite is an ordered list of checks.
type Suite struct {
	Name   string
	Checks []Check
}

// Run executes all checks and returns a report.
func (s *Suite) Run(ctx *CheckContext) domain.ValidationReport {
	report := domain.ValidationReport{
		Suite: s.Name,
		Total: len(s.Checks),
	}

	for _, check := range s.Checks {
		passed, msg := check.Fn(ctx)

		// In strict mode, promote warnings to failures for stub detection
		level := check.Level
		if ctx.Strict && check.ID == 16 && level == domain.LevelWarn {
			level = domain.LevelFail
		}

		result := domain.CheckResult{
			ID:      check.ID,
			Name:    check.Name,
			Level:   level,
			Passed:  passed,
			Message: msg,
		}
		report.Checks = append(report.Checks, result)

		if passed {
			report.Passed++
		} else if level == domain.LevelFail {
			report.Failed++
		} else {
			report.Warnings++
		}
	}

	return report
}
