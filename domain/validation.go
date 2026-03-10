package domain

// CheckLevel indicates the severity of a validation check.
type CheckLevel string

const (
	LevelFail CheckLevel = "FAIL"
	LevelWarn CheckLevel = "WARN"
	LevelInfo CheckLevel = "INFO"
)

// CheckResult represents the outcome of a single validation check.
type CheckResult struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Level   CheckLevel `json:"level"`
	Passed  bool       `json:"passed"`
	Message string     `json:"message,omitempty"`
}

// ValidationReport aggregates results from a validation suite.
type ValidationReport struct {
	Suite    string        `json:"suite"`
	Checks   []CheckResult `json:"checks"`
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Warnings int           `json:"warnings"`
}

// Ok returns true if no checks failed.
func (r *ValidationReport) Ok() bool { return r.Failed == 0 }
