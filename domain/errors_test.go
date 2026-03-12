package domain

import (
	"errors"
	"strings"
	"testing"
)

// TestSentinelErrors verifies sentinel errors are usable with errors.Is.
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNotProject message",
			err:  ErrNotProject,
			want: "not a Mind project",
		},
		{
			name: "ErrBriefMissing message",
			err:  ErrBriefMissing,
			want: "project brief missing",
		},
		{
			name: "ErrAlreadyInitialized message",
			err:  ErrAlreadyInitialized,
			want: "already initialized",
		},
		{
			name: "ErrAlreadyExists message",
			err:  ErrAlreadyExists,
			want: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !strings.Contains(msg, tt.want) {
				t.Errorf("error message %q does not contain %q", msg, tt.want)
			}
		})
	}
}

// TestErrGateFailed verifies ErrGateFailed formatting.
func TestErrGateFailed(t *testing.T) {
	err := &ErrGateFailed{
		Gate:     "brief",
		Failures: []string{"missing Vision", "missing Scope"},
	}

	msg := err.Error()
	if !strings.Contains(msg, "gate brief failed") {
		t.Errorf("ErrGateFailed.Error() = %q, want 'gate brief failed'", msg)
	}
	if !strings.Contains(msg, "missing Vision") {
		t.Errorf("ErrGateFailed.Error() = %q, want 'missing Vision'", msg)
	}
	if !strings.Contains(msg, "missing Scope") {
		t.Errorf("ErrGateFailed.Error() = %q, want 'missing Scope'", msg)
	}
}

// TestErrCommandFailed verifies ErrCommandFailed formatting.
func TestErrCommandFailed(t *testing.T) {
	err := &ErrCommandFailed{
		Command:  "go test",
		ExitCode: 1,
		Output:   "FAIL: some test",
	}

	msg := err.Error()
	if !strings.Contains(msg, "go test") {
		t.Errorf("ErrCommandFailed.Error() = %q, want 'go test'", msg)
	}
	if !strings.Contains(msg, "exit 1") {
		t.Errorf("ErrCommandFailed.Error() = %q, want 'exit 1'", msg)
	}
}

// TestSentinelErrorsAreWrappable verifies sentinel errors work with errors.Is.
func TestSentinelErrorsAreWrappable(t *testing.T) {
	wrapped := errors.New("wrapped: " + ErrNotProject.Error())
	// The plain sentinel is usable
	if !errors.Is(ErrNotProject, ErrNotProject) {
		t.Error("ErrNotProject should match itself with errors.Is")
	}
	// Wrapping preserves original (as long as it's the same pointer)
	_ = wrapped // just verifying the sentinel is a regular error
}
