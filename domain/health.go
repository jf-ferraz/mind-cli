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

// Diagnostic represents an issue found by `mind doctor`.
type Diagnostic struct {
	Level   CheckLevel `json:"level"`
	Message string     `json:"message"`
	Fix     string     `json:"fix,omitempty"`
	AutoFix bool       `json:"auto_fix"`
}

// Suggestion represents an actionable next step.
type Suggestion struct {
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	Command string `json:"command,omitempty"`
}
