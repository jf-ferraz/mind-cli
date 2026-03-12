package repo

import "testing"

// TestIsStubContent verifies FR-50 and BR-2: stub detection algorithm.
func TestIsStubContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		// FR-50: GIVEN a file with only headings, comments, placeholders THEN stub
		{
			name:    "headings and comment only is stub",
			content: "# Title\n## Section\n<!-- placeholder -->\n",
			want:    true,
		},
		// FR-50: GIVEN a file with real content THEN not stub (needs >2 real lines)
		{
			name:    "real content is not stub",
			content: "# Title\n\nThis project provides a REST API for user management.\nIt supports CRUD operations for users and roles.\nAuthentication is handled via JWT tokens.\n",
			want:    false,
		},
		// Empty file is stub
		{
			name:    "empty content is stub",
			content: "",
			want:    true,
		},
		// Only headings
		{
			name:    "only headings is stub",
			content: "# Title\n## Subtitle\n### Sub-subtitle\n",
			want:    true,
		},
		// Only HTML comments
		{
			name:    "only HTML comments is stub",
			content: "<!-- This is a comment -->\n<!-- Another comment -->\n",
			want:    true,
		},
		// Only blockquotes
		{
			name:    "only blockquotes is stub",
			content: "> This is a blockquote\n> Another line\n",
			want:    true,
		},
		// Table separators only (with heading)
		{
			name:    "table separator with heading",
			content: "# Table\n\n| Col1 | Col2 |\n|------|------|\n",
			want:    true,
		},
		// Placeholder rows
		{
			name:    "placeholder table rows",
			content: "# Table\n| Col1 | Col2 |\n|------|------|\n| <!-- name --> | <!-- value --> |\n",
			want:    true,
		},
		// Mixed boilerplate is still a stub
		{
			name:    "mixed boilerplate is stub",
			content: "# Title\n\n> Blockquote\n\n<!-- comment -->\n\n## Section\n",
			want:    true,
		},
		// Single real line + boilerplate is still a stub (<=2 real lines)
		{
			name:    "one real line is stub",
			content: "# Title\n\nOne real line of content.\n",
			want:    true,
		},
		// Two real lines is still a stub (<=2)
		{
			name:    "two real lines is stub",
			content: "# Title\n\nFirst real line.\nSecond real line.\n",
			want:    true,
		},
		// Three real lines is NOT a stub (>2)
		{
			name:    "three real lines is not stub",
			content: "# Title\n\nFirst real line.\nSecond real line.\nThird real line.\n",
			want:    false,
		},
		// Real-world stub brief template
		{
			name: "stub brief template is stub",
			content: `# Project Brief

## Vision

<!-- Describe the project vision -->

## Key Deliverables

<!-- List key deliverables -->

## Scope

### In Scope

<!-- What is in scope -->

### Out of Scope

<!-- What is out of scope -->

## Constraints

<!-- List constraints -->
`,
			want: true,
		},
		// Real-world filled brief is NOT a stub
		{
			name: "filled brief is not stub",
			content: `# Project Brief

## Vision

Build a CLI tool that replaces 9 bash scripts with a single Go binary.
The tool provides project health diagnostics and document generation.
It supports three output modes: interactive, plain, and JSON.

## Key Deliverables

- Core CLI binary with 20+ commands
- Validation engine with 28+ checks
- Document scaffolding for 6 artifact types

## Scope

### In Scope

- Domain types and business rules
- Repository and service layers
- CLI command handlers
`,
			want: false,
		},
		// Continuation comment (-->)
		{
			name:    "continuation comment marker",
			content: "# Title\n-->\n<!-- begin -->\n",
			want:    true,
		},
		// Table with colon alignment
		{
			name:    "table separator with colons",
			content: "# Table\n|:------|------:|\n",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStubContent([]byte(tt.content))
			if got != tt.want {
				t.Errorf("IsStubContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
