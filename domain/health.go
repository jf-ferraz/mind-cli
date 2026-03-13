package domain

// ProjectHealth is the aggregate status shown by `mind status`.
type ProjectHealth struct {
	Project       Project             `json:"project"`
	Brief         Brief               `json:"brief"`
	Zones         map[Zone]ZoneHealth `json:"zones"`
	Workflow      *WorkflowState      `json:"workflow,omitempty"`
	LastIteration *Iteration          `json:"last_iteration,omitempty"`
	Warnings      []string            `json:"warnings,omitempty"`
	Suggestions   []string            `json:"suggestions,omitempty"`
	Staleness     *StalenessInfo      `json:"staleness"`
	Framework     *FrameworkStatus    `json:"framework,omitempty"`
}

// FrameworkStatus is the framework section shown in `mind status`.
type FrameworkStatus struct {
	Mode       string `json:"mode"`
	Version    string `json:"version"`
	DriftCount int    `json:"drift_count"`
}

// ZoneHealth tracks completeness of a single documentation zone.
type ZoneHealth struct {
	Zone     Zone       `json:"zone"`
	Total    int        `json:"total"`
	Present  int        `json:"present"`
	Stubs    int        `json:"stubs"`
	Complete int        `json:"complete"`
	Files    []Document `json:"files,omitempty"`
}

// DiagnosticStatus indicates the outcome of a doctor diagnostic check.
type DiagnosticStatus string

const (
	DiagPass DiagnosticStatus = "pass"
	DiagFail DiagnosticStatus = "fail"
	DiagWarn DiagnosticStatus = "warn"
)

// Diagnostic represents an issue found by `mind doctor`.
type Diagnostic struct {
	Category string           `json:"category"`
	Check    string           `json:"check"`
	Status   DiagnosticStatus `json:"status"`
	Level    CheckLevel       `json:"-"`
	Message  string           `json:"message"`
	Fix      string           `json:"fix,omitempty"`
	AutoFix  bool             `json:"auto_fixable"`
}

// DoctorReport aggregates diagnostics from `mind doctor`.
type DoctorReport struct {
	Diagnostics  []Diagnostic  `json:"diagnostics"`
	Summary      DoctorSummary `json:"summary"`
	FixesApplied []string      `json:"fixes_applied,omitempty"`
}

// DoctorSummary counts diagnostic outcomes.
type DoctorSummary struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
	Warn int `json:"warn"`
}

// InitResult represents the output of `mind init`.
type InitResult struct {
	ProjectName       string   `json:"project_name"`
	Root              string   `json:"root"`
	FilesCreated      []string `json:"files_created"`
	FromExisting      bool     `json:"from_existing"`
	ExistingPreserved []string `json:"existing_preserved,omitempty"`
}

// CreateResult represents the output of `mind create adr|blueprint|spike|convergence`.
type CreateResult struct {
	Path         string `json:"path"`
	Seq          int    `json:"seq,omitempty"`
	Title        string `json:"title"`
	IndexUpdated bool   `json:"index_updated,omitempty"`
}

// CreateIterationResult represents the output of `mind create iteration`.
type CreateIterationResult struct {
	Path       string      `json:"path"`
	Seq        int         `json:"seq"`
	Type       RequestType `json:"type"`
	Descriptor string      `json:"descriptor"`
	Files      []string    `json:"files"`
}

// Suggestion represents an actionable next step.
type Suggestion struct {
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	Command string `json:"command,omitempty"`
}

// DocumentList represents the output of `mind docs list`.
type DocumentList struct {
	Documents []Document     `json:"documents"`
	ByZone    map[string]int `json:"by_zone"`
	Total     int            `json:"total"`
}

// StubEntry represents a stub document with remediation hint.
type StubEntry struct {
	Path string `json:"path"`
	Zone string `json:"zone"`
	Hint string `json:"hint"`
}

// StubList represents the output of `mind docs stubs`.
type StubList struct {
	Stubs []StubEntry `json:"stubs"`
	Count int         `json:"count"`
}

// SearchMatch represents a single line match in a search result.
type SearchMatch struct {
	Line          int    `json:"line"`
	Text          string `json:"text"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
}

// SearchFileResult groups matches within a single file.
type SearchFileResult struct {
	Path    string        `json:"path"`
	Matches []SearchMatch `json:"matches"`
}

// SearchResults represents the output of `mind docs search`.
type SearchResults struct {
	Query        string             `json:"query"`
	Results      []SearchFileResult `json:"results"`
	TotalMatches int                `json:"total_matches"`
	FilesMatched int                `json:"files_matched"`
}

// UnifiedValidationReport aggregates multiple validation suites.
type UnifiedValidationReport struct {
	Suites  []ValidationReport       `json:"suites"`
	Summary UnifiedValidationSummary `json:"summary"`
}

// UnifiedValidationSummary aggregates counts across suites.
type UnifiedValidationSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Warnings int `json:"warnings"`
}

// WorkflowHistory represents the output of `mind workflow history`.
type WorkflowHistory struct {
	Iterations []IterationSummary `json:"iterations"`
	Total      int                `json:"total"`
}

// IterationSummary is a compact iteration representation for history.
type IterationSummary struct {
	Seq        int             `json:"seq"`
	Type       RequestType     `json:"type"`
	Descriptor string          `json:"descriptor"`
	DirName    string          `json:"dir_name"`
	Status     IterationStatus `json:"status"`
	CreatedAt  string          `json:"created_at"`
	Artifacts  ArtifactCount   `json:"artifacts"`
}

// ArtifactCount tracks present vs expected artifacts.
type ArtifactCount struct {
	Present  int `json:"present"`
	Expected int `json:"expected"`
}

// VersionInfo represents the output of `mind version --json`.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}
