package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// TestParseBrief verifies BR-3: brief gate classification.
func TestParseBrief(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(root string) // Creates files before test
		wantGate   domain.BriefGate
		wantExists bool
		wantStub   bool
	}{
		{
			name:       "BRIEF_MISSING when file does not exist",
			setup:      func(root string) { /* no file */ },
			wantGate:   domain.BriefMissing,
			wantExists: false,
			wantStub:   false,
		},
		{
			name: "BRIEF_STUB when file is a stub template",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Vision

<!-- Describe the project vision -->

## Key Deliverables

<!-- List key deliverables -->

## Scope

<!-- What is in scope -->
`)
			},
			wantGate:   domain.BriefStub,
			wantExists: true,
			wantStub:   true,
		},
		{
			name: "BRIEF_PRESENT when all sections have real content",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Vision

Build a comprehensive CLI tool for project management.
The tool replaces legacy bash scripts with a single binary.
It provides health diagnostics and document generation.

## Key Deliverables

- Core CLI with 20+ commands
- Validation engine with 28+ checks
- Document scaffolding for 6 types

## Scope

### In Scope

- Domain types and business rules
- Repository and service layers
- Validation suites
`)
			},
			wantGate:   domain.BriefPresent,
			wantExists: true,
			wantStub:   false,
		},
		{
			name: "BRIEF_STUB when content exists but missing Vision section",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Key Deliverables

- Core CLI with 20+ commands
- Validation engine with 28+ checks
- Document scaffolding for 6 types

## Scope

### In Scope

- Domain types and business rules
- Repository and service layers
- Validation suites
`)
			},
			wantGate:   domain.BriefStub,
			wantExists: true,
			wantStub:   false,
		},
		{
			name: "BRIEF_STUB when content exists but missing Key Deliverables section",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Vision

Build a comprehensive CLI tool for project management.
The tool replaces legacy bash scripts with a single binary.
It provides health diagnostics and document generation.

## Scope

### In Scope

- Domain types and business rules
- Repository and service layers
- Validation suites
`)
			},
			wantGate:   domain.BriefStub,
			wantExists: true,
			wantStub:   false,
		},
		{
			name: "BRIEF_STUB when content exists but missing Scope section",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Vision

Build a comprehensive CLI tool for project management.
The tool replaces legacy bash scripts with a single binary.
It provides health diagnostics and document generation.

## Key Deliverables

- Core CLI with 20+ commands
- Validation engine with 28+ checks
- Document scaffolding for 6 types
`)
			},
			wantGate:   domain.BriefStub,
			wantExists: true,
			wantStub:   false,
		},
		{
			name: "BRIEF_PRESENT with case-insensitive section matching",
			setup: func(root string) {
				writeFile(t, root, "docs/spec/project-brief.md", `# Project Brief

## Project Vision

A comprehensive approach to CLI tooling.
This replaces the legacy scripts with a modern binary.
Supports multiple output formats for different contexts.

## Key Deliverables

- CLI binary
- Validation engine
- Document templates

## Project Scope

Everything in Phase 1.
Core domain types.
Repository implementations.
`)
			},
			wantGate:   domain.BriefPresent,
			wantExists: true,
			wantStub:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			// Always create the docs/spec directory
			os.MkdirAll(filepath.Join(root, "docs", "spec"), 0755)
			tt.setup(root)

			docRepo := NewDocRepo(root)
			briefRepo := NewBriefRepo(docRepo)
			brief, err := briefRepo.ParseBrief()
			if err != nil {
				t.Fatalf("ParseBrief() error = %v", err)
			}

			if brief.GateResult != tt.wantGate {
				t.Errorf("GateResult = %q, want %q", brief.GateResult, tt.wantGate)
			}
			if brief.Exists != tt.wantExists {
				t.Errorf("Exists = %v, want %v", brief.Exists, tt.wantExists)
			}
			if brief.IsStub != tt.wantStub {
				t.Errorf("IsStub = %v, want %v", brief.IsStub, tt.wantStub)
			}
		})
	}
}

func writeFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(root, relPath)
	os.MkdirAll(filepath.Dir(absPath), 0755)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", relPath, err)
	}
}
