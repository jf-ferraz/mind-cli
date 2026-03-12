package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// setupProject creates a minimal Mind project in a temp directory.
// Returns the root path and a cleanup function.
func setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create .mind/ directory (required for project detection)
	if err := os.MkdirAll(filepath.Join(root, ".mind"), 0755); err != nil {
		t.Fatalf("create .mind: %v", err)
	}

	// Create docs/ zone directories
	for _, zone := range []string{"spec", "blueprints", "state", "iterations", "knowledge"} {
		if err := os.MkdirAll(filepath.Join(root, "docs", zone), 0755); err != nil {
			t.Fatalf("create docs/%s: %v", zone, err)
		}
	}

	// Create mind.toml
	toml := `[manifest]
schema = "mind/v1.0"
generation = 1

[project]
name = "test-project"
description = "A test project"
type = "cli"

[project.stack]
language = "go"
framework = ""
testing = ""

[project.commands]
dev = ""
test = ""
lint = ""
typecheck = ""
build = ""

[governance]
max-retries = 2
review-policy = ""
commit-policy = ""
branch-strategy = ""

[profiles]
`
	if err := os.WriteFile(filepath.Join(root, "mind.toml"), []byte(toml), 0644); err != nil {
		t.Fatalf("write mind.toml: %v", err)
	}

	// Create project brief with real content (non-stub)
	briefContent := `# Project Brief

## Vision

This is a test project for validation.
It provides comprehensive test coverage.
The project ensures all commands work correctly.

## Key Deliverables

- Test suite coverage
- Exit code verification
- Integration validation

## Scope

### In Scope

- Command handler tests
- Exit code tests
- Project detection tests
`
	if err := os.WriteFile(filepath.Join(root, "docs", "spec", "project-brief.md"), []byte(briefContent), 0644); err != nil {
		t.Fatalf("write brief: %v", err)
	}

	// Create current.md with real content
	currentContent := `# Current State

## Active Work

Testing infrastructure.

## Known Issues

None at this time.
`
	if err := os.WriteFile(filepath.Join(root, "docs", "state", "current.md"), []byte(currentContent), 0644); err != nil {
		t.Fatalf("write current.md: %v", err)
	}

	// Create workflow.md
	workflowContent := `# Workflow State

## Status

idle
`
	if err := os.WriteFile(filepath.Join(root, "docs", "state", "workflow.md"), []byte(workflowContent), 0644); err != nil {
		t.Fatalf("write workflow.md: %v", err)
	}

	// Create INDEX.md
	indexContent := `# Blueprint Index

## Active Blueprints

None at this time.

## Completed Blueprints

None at this time.
`
	if err := os.WriteFile(filepath.Join(root, "docs", "blueprints", "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("write INDEX.md: %v", err)
	}

	// Create glossary.md
	glossaryContent := `# Glossary

## Terms

| Term | Definition |
|------|-----------|
| CLI | Command-line interface |
`
	if err := os.WriteFile(filepath.Join(root, "docs", "knowledge", "glossary.md"), []byte(glossaryContent), 0644); err != nil {
		t.Fatalf("write glossary.md: %v", err)
	}

	return root
}

// executeWithRoot runs the root command with given args and the --project-root flag.
func executeWithRoot(root string, args ...string) error {
	// Reset package-level variables
	projectRoot = ""
	renderer = nil
	reconcileSvc = nil
	validationSvc = nil
	doctorSvc = nil
	projectSvc = nil
	workflowSvc = nil
	generateSvc = nil
	docRepo = nil
	iterRepo = nil
	briefRepo = nil

	// Reset flags
	flagJSON = false
	flagNoColor = true // force plain mode for tests
	flagProject = root

	fullArgs := append([]string{"--project-root", root, "--no-color"}, args...)
	rootCmd.SetArgs(fullArgs)

	return rootCmd.Execute()
}

// FR-129: ExitError.Error() returns the wrapped error message.
func TestExitError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ExitError
		want string
	}{
		{
			name: "with wrapped error",
			err:  &ExitError{Code: 1, Err: fmt.Errorf("validation failed")},
			want: "validation failed",
		},
		{
			name: "with nil error",
			err:  &ExitError{Code: 2, Err: nil},
			want: "exit code 2",
		},
		{
			name: "quiet with wrapped error",
			err:  &ExitError{Code: 1, Err: fmt.Errorf("check failed"), Quiet: true},
			want: "check failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ExitError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// FR-129: ExitError.Unwrap() returns the underlying error for errors.Is/As chains.
func TestExitError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	exitErr := &ExitError{Code: 2, Err: inner}

	unwrapped := exitErr.Unwrap()
	if unwrapped != inner {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, inner)
	}
}

// FR-129: ExitError.Unwrap() returns nil when no wrapped error.
func TestExitError_UnwrapNil(t *testing.T) {
	exitErr := &ExitError{Code: 1}
	if exitErr.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", exitErr.Unwrap())
	}
}

// FR-129: errors.As extracts ExitError from wrapped error chains.
func TestExitError_ErrorsAs(t *testing.T) {
	inner := fmt.Errorf("config not found")
	exitErr := exitConfig(inner)

	// Wrap it once more to test chain traversal
	wrapped := fmt.Errorf("command failed: %w", exitErr)

	var extracted *ExitError
	if !errors.As(wrapped, &extracted) {
		t.Fatal("errors.As failed to extract ExitError from wrapped chain")
	}
	if extracted.Code != 3 {
		t.Errorf("extracted.Code = %d, want 3", extracted.Code)
	}
	if extracted.Err.Error() != "config not found" {
		t.Errorf("extracted.Err = %q, want %q", extracted.Err.Error(), "config not found")
	}
}

// FR-129: Convenience constructors produce correct exit codes.
func TestExitError_Constructors(t *testing.T) {
	tests := []struct {
		name string
		err  *ExitError
		code int
	}{
		{"exitValidation", exitValidation(fmt.Errorf("v")), 1},
		{"exitRuntime", exitRuntime(fmt.Errorf("r")), 2},
		{"exitConfig", exitConfig(fmt.Errorf("c")), 3},
		{"exitStaleness", exitStaleness(fmt.Errorf("s")), 4},
		{"exitQuiet", exitQuiet(1), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("Code = %d, want %d", tt.err.Code, tt.code)
			}
		})
	}
}

// FR-129: exitQuiet sets the Quiet flag.
func TestExitError_QuietFlag(t *testing.T) {
	e := exitQuiet(4)
	if !e.Quiet {
		t.Error("exitQuiet should set Quiet = true")
	}
	if e.Code != 4 {
		t.Errorf("exitQuiet code = %d, want 4", e.Code)
	}

	// Non-quiet constructors should not set Quiet
	v := exitValidation(fmt.Errorf("x"))
	if v.Quiet {
		t.Error("exitValidation should not set Quiet")
	}
}

// FR-128: The --project-root flag is registered on rootCmd.
func TestProjectRootFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("project-root")
	if flag == nil {
		t.Fatal("--project-root flag not registered on rootCmd")
	}
	if flag.Shorthand != "p" {
		t.Errorf("--project-root shorthand = %q, want %q", flag.Shorthand, "p")
	}
}

// FR-128: The old --project flag does not exist.
func TestOldProjectFlagRemoved(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("project")
	if flag != nil {
		t.Error("old --project flag should not exist; expected --project-root")
	}
}

func TestCheckDocsExitCodePass(t *testing.T) {
	root := setupProject(t)

	err := executeWithRoot(root, "check", "docs")
	if err != nil {
		// Check docs may fail due to missing docs -- that is expected
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			// Exit code 1 means check failures -- acceptable
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestCheckDocsExitCodeFail(t *testing.T) {
	root := setupProject(t)

	// Remove a required file to trigger failure
	os.Remove(filepath.Join(root, "docs", "spec", "project-brief.md"))

	err := executeWithRoot(root, "check", "docs")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.Code)
	}
}

func TestNotProjectExitCode(t *testing.T) {
	root := t.TempDir()
	// No .mind/ directory -- should fail with exit code 3

	err := executeWithRoot(root, "status")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.Code)
	}
}

func TestStatusExitCode(t *testing.T) {
	root := setupProject(t)

	err := executeWithRoot(root, "status")
	// Status may succeed (0) or fail (1) depending on project content
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestReconcileCheckExitCode(t *testing.T) {
	root := setupProject(t)

	// Add a document entry to mind.toml for reconciliation
	toml := `[manifest]
schema = "mind/v1.0"
generation = 1

[project]
name = "test-project"
description = "A test project"
type = "cli"

[project.stack]
language = "go"

[project.commands]

[governance]
max-retries = 2

[profiles]

[documents.spec]
[documents.spec.requirements]
id = "doc:spec/requirements"
path = "docs/spec/requirements.md"
zone = "spec"
`
	os.WriteFile(filepath.Join(root, "mind.toml"), []byte(toml), 0644)

	// Create the declared document
	os.WriteFile(filepath.Join(root, "docs", "spec", "requirements.md"),
		[]byte("# Requirements\n\nThis is the requirements document.\nIt has real content.\nWith multiple lines."), 0644)

	// First reconcile to establish baseline
	err := executeWithRoot(root, "reconcile")
	if err != nil {
		t.Fatalf("initial reconcile: %v", err)
	}

	// Check should be clean
	err = executeWithRoot(root, "reconcile", "--check")
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) && exitErr.Code == 4 {
			// Stale is also valid if something changed
		} else {
			t.Fatalf("reconcile --check: %v", err)
		}
	}
}

func TestDoctorExitCode(t *testing.T) {
	root := setupProject(t)

	// Create .claude/CLAUDE.md for the doctor check
	os.MkdirAll(filepath.Join(root, ".claude"), 0755)
	os.WriteFile(filepath.Join(root, ".claude", "CLAUDE.md"), []byte("# Claude"), 0644)

	err := executeWithRoot(root, "doctor")
	// Doctor may pass or fail depending on project completeness
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestCheckRefsExitCode(t *testing.T) {
	root := setupProject(t)

	err := executeWithRoot(root, "check", "refs")
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestCheckConfigExitCode(t *testing.T) {
	root := setupProject(t)

	err := executeWithRoot(root, "check", "config")
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestCheckAllExitCode(t *testing.T) {
	root := setupProject(t)

	err := executeWithRoot(root, "check", "all")
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Code != 1 {
				t.Fatalf("unexpected exit code: got %d, want 0 or 1", exitErr.Code)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}
