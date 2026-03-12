package domain

import "time"

// GateCommandResult is the result of a single gate command execution.
type GateCommandResult struct {
	Name     string        `json:"name"`
	Command  string        `json:"command"`
	Pass     bool          `json:"pass"`
	Duration time.Duration `json:"duration_ns"`
	Stdout   string        `json:"stdout,omitempty"`
	Stderr   string        `json:"stderr,omitempty"`
	ExitCode int           `json:"exit_code"`
}

// GateResult is the aggregate result of running build/lint/test commands.
type GateResult struct {
	Commands []GateCommandResult `json:"commands"`
	Pass     bool                `json:"pass"`
	Total    int                 `json:"total"`
	Passed   int                 `json:"passed"`
}
