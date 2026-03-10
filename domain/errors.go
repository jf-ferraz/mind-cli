package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotProject signals that no .mind/ was found.
var ErrNotProject = errors.New("not a Mind project (no .mind/ directory found)")

// ErrBriefMissing signals a missing project brief for gate enforcement.
var ErrBriefMissing = errors.New("project brief missing — run /discover or create docs/spec/project-brief.md")

// ErrGateFailed signals a quality gate failure.
type ErrGateFailed struct {
	Gate     string
	Failures []string
}

func (e *ErrGateFailed) Error() string {
	return fmt.Sprintf("gate %s failed: %s", e.Gate, strings.Join(e.Failures, "; "))
}

// ErrCommandFailed signals an external command failure.
type ErrCommandFailed struct {
	Command  string
	ExitCode int
	Output   string
}

func (e *ErrCommandFailed) Error() string {
	return fmt.Sprintf("command %q failed (exit %d): %s", e.Command, e.ExitCode, e.Output)
}
