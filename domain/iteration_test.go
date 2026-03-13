package domain

import "testing"

// TestSlugify verifies FR-31 acceptance criteria and BR-16 rules.
func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// FR-31: GIVEN input "Use PostgreSQL (v15+)" WHEN slugified THEN result is "use-postgresql-v15"
		{name: "FR-31 acceptance criteria", input: "Use PostgreSQL (v15+)", want: "use-postgresql-v15"},
		// BR-16: lowercase input
		{name: "lowercase", input: "HELLO", want: "hello"},
		// BR-16: replace spaces with hyphens
		{name: "spaces to hyphens", input: "hello world", want: "hello-world"},
		// BR-16: replace non-alphanumeric with hyphens
		{name: "non-alphanumeric", input: "hello@world!!", want: "hello-world"},
		// BR-16: strip leading/trailing hyphens
		{name: "leading trailing hyphens", input: "  hello  ", want: "hello"},
		// BR-16: collapse multiple hyphens
		{name: "collapse multi hyphens", input: "a---b", want: "a-b"},
		// Simple input
		{name: "already slug", input: "my-project", want: "my-project"},
		// Numbers preserved
		{name: "numbers preserved", input: "v2 release", want: "v2-release"},
		// Mixed case and special chars
		{name: "mixed case special", input: "Add Caching (Redis)", want: "add-caching-redis"},
		// Single word
		{name: "single word", input: "auth", want: "auth"},
		// BR-16: idempotent
		{name: "idempotent", input: "use-postgresql-v15", want: "use-postgresql-v15"},
		// Accented / unicode → stripped to hyphens
		{name: "unicode chars", input: "café résumé", want: "caf-r-sum"},
		// Only special chars → empty (edge case)
		{name: "only special chars", input: "!!!", want: ""},
		// Empty input
		{name: "empty input", input: "", want: ""},
		// Leading special chars
		{name: "leading special", input: "---hello", want: "hello"},
		// Trailing special chars
		{name: "trailing special", input: "hello---", want: "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestSlugifyIdempotent verifies BR-16: Slugify is idempotent.
func TestSlugifyIdempotent(t *testing.T) {
	inputs := []string{
		"Use PostgreSQL (v15+)",
		"Hello World",
		"my-project",
		"A B C",
	}
	for _, input := range inputs {
		first := Slugify(input)
		second := Slugify(first)
		if first != second {
			t.Errorf("Slugify not idempotent: Slugify(%q) = %q, Slugify(%q) = %q", input, first, first, second)
		}
	}
}

// TestClassify verifies BR-19 classification rules.
func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RequestType
	}{
		// Explicit prefix matching (strongest signal)
		{name: "create: prefix", input: "create: new user service", want: TypeNewProject},
		{name: "build: prefix", input: "build: authentication module", want: TypeNewProject},
		{name: "fix: prefix", input: "fix: login button broken", want: TypeBugFix},
		{name: "add: prefix", input: "add: dark mode support", want: TypeEnhancement},
		{name: "refactor: prefix", input: "refactor: database layer", want: TypeRefactor},
		{name: "analyze: prefix", input: "analyze: performance bottleneck", want: TypeComplexNew},
		{name: "explore: prefix", input: "explore: new architecture", want: TypeComplexNew},
		// Explicit prefix - diagnose
		{name: "diagnose: prefix", input: "diagnose: intermittent test failure", want: TypeDiagnose},
		{name: "investigate: prefix", input: "investigate: memory leak in prod", want: TypeDiagnose},
		// Keyword matching - diagnose keywords (before bug keywords)
		{name: "diagnose keyword", input: "need to diagnose the issue", want: TypeDiagnose},
		{name: "investigate keyword", input: "please investigate the failure", want: TypeDiagnose},
		{name: "troubleshoot keyword", input: "troubleshoot network timeout", want: TypeDiagnose},
		{name: "root cause keyword", input: "find the root cause of this", want: TypeDiagnose},
		{name: "why is keyword", input: "why is the server slow", want: TypeDiagnose},
		{name: "intermittent keyword", input: "intermittent failures in CI", want: TypeDiagnose},
		// Keyword matching - bug keywords
		{name: "fix keyword", input: "the login is broken", want: TypeBugFix},
		{name: "bug keyword", input: "bug in payment processing", want: TypeBugFix},
		{name: "error keyword", input: "error handling improvements", want: TypeBugFix},
		{name: "crash keyword", input: "app crash on startup", want: TypeBugFix},
		{name: "regression keyword", input: "regression in tests", want: TypeBugFix},
		{name: "failing keyword", input: "tests are failing", want: TypeBugFix},
		// Keyword matching - refactor keywords
		{name: "refactor keyword", input: "we should refactor the code", want: TypeRefactor},
		{name: "clean keyword", input: "clean up old code", want: TypeRefactor},
		{name: "restructure keyword", input: "restructure modules", want: TypeRefactor},
		{name: "optimize keyword", input: "optimize query performance", want: TypeRefactor},
		{name: "simplify keyword", input: "simplify the interface", want: TypeRefactor},
		{name: "modernize keyword", input: "modernize the codebase", want: TypeRefactor},
		// Keyword matching - enhancement keywords
		{name: "add keyword", input: "pagination for the list", want: TypeEnhancement},
		{name: "feature keyword", input: "new feature for dashboard", want: TypeEnhancement},
		{name: "extend keyword", input: "extend API endpoints", want: TypeEnhancement},
		{name: "improve keyword", input: "improve search results", want: TypeEnhancement},
		{name: "integrate keyword", input: "integrate third-party API", want: TypeEnhancement},
		{name: "support keyword", input: "support for dark mode", want: TypeEnhancement},
		// Keyword matching - new project keywords
		{name: "create keyword", input: "let's create a new service", want: TypeNewProject},
		{name: "build keyword", input: "i want to build something", want: TypeNewProject},
		{name: "new project keyword", input: "this is a new project", want: TypeNewProject},
		{name: "scaffold keyword", input: "scaffold the boilerplate", want: TypeNewProject},
		// Default: ENHANCEMENT
		{name: "ambiguous defaults to enhancement", input: "update the docs", want: TypeEnhancement},
		{name: "empty defaults to enhancement", input: "", want: TypeEnhancement},
		// Case insensitivity
		{name: "case insensitive prefix", input: "Fix: uppercase prefix", want: TypeBugFix},
		{name: "case insensitive keyword", input: "There is a BUG", want: TypeBugFix},
		// Priority: prefix > keyword
		{name: "prefix overrides keyword", input: "add: fix the bug system", want: TypeEnhancement},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.input)
			if got != tt.want {
				t.Errorf("Classify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestClassifyDeterministic verifies BR-19: Classify is deterministic.
func TestClassifyDeterministic(t *testing.T) {
	inputs := []string{
		"fix: broken login",
		"add caching layer",
		"refactor the database",
		"something ambiguous",
		"diagnose: the test regression",
	}
	for _, input := range inputs {
		first := Classify(input)
		second := Classify(input)
		if first != second {
			t.Errorf("Classify not deterministic: Classify(%q) first=%q, second=%q", input, first, second)
		}
	}
}

// TestExpectedArtifacts verifies BR-7: exactly 5 expected artifacts.
func TestExpectedArtifacts(t *testing.T) {
	if len(ExpectedArtifacts) != 5 {
		t.Errorf("ExpectedArtifacts has %d items, want 5", len(ExpectedArtifacts))
	}

	expected := map[string]bool{
		"overview.md":      true,
		"changes.md":       true,
		"test-summary.md":  true,
		"validation.md":    true,
		"retrospective.md": true,
	}
	for _, a := range ExpectedArtifacts {
		if !expected[a] {
			t.Errorf("Unexpected artifact: %q", a)
		}
	}
}
