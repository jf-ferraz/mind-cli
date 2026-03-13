package generate

import (
	"strings"
	"testing"
)

// TestADRTemplate verifies FR-24: ADR template content.
func TestADRTemplate(t *testing.T) {
	content := ADRTemplate("Use PostgreSQL", 3)

	if !strings.Contains(content, "3. Use PostgreSQL") {
		t.Error("ADR template should contain sequence and title")
	}
	if !strings.Contains(content, "**Status**: Proposed") {
		t.Error("ADR template should have Proposed status")
	}
	if !strings.Contains(content, "## Context") {
		t.Error("ADR template should have Context section")
	}
	if !strings.Contains(content, "## Decision") {
		t.Error("ADR template should have Decision section")
	}
	if !strings.Contains(content, "## Consequences") {
		t.Error("ADR template should have Consequences section")
	}
}

// TestBlueprintTemplate verifies FR-25: Blueprint template content.
func TestBlueprintTemplate(t *testing.T) {
	content := BlueprintTemplate("Auth System", 4)

	if !strings.Contains(content, "BP-04: Auth System") {
		t.Error("Blueprint template should contain BP-NN and title")
	}
	if !strings.Contains(content, "## Overview") {
		t.Error("Blueprint template should have Overview section")
	}
	if !strings.Contains(content, "## Design") {
		t.Error("Blueprint template should have Design section")
	}
}

// TestIterationTemplates verifies FR-26: iteration templates.
func TestIterationTemplates(t *testing.T) {
	t.Run("overview", func(t *testing.T) {
		content := IterationOverviewTemplate("core-cli", "NEW_PROJECT")
		if !strings.Contains(content, "core cli") {
			t.Error("Overview should contain descriptor (unhyphenated)")
		}
		if !strings.Contains(content, "NEW_PROJECT") {
			t.Error("Overview should contain request type")
		}
		if !strings.Contains(content, "## Scope") {
			t.Error("Overview should have Scope section")
		}
	})

	t.Run("changes", func(t *testing.T) {
		content := IterationChangesTemplate()
		if !strings.Contains(content, "# Changes") {
			t.Error("Changes template should have title")
		}
		if !strings.Contains(content, "| File |") {
			t.Error("Changes template should have table header")
		}
	})

	t.Run("test-summary", func(t *testing.T) {
		content := IterationTestSummaryTemplate()
		if !strings.Contains(content, "# Test Summary") {
			t.Error("Test summary template should have title")
		}
	})

	t.Run("validation", func(t *testing.T) {
		content := IterationValidationTemplate()
		if !strings.Contains(content, "# Validation") {
			t.Error("Validation template should have title")
		}
	})

	t.Run("retrospective", func(t *testing.T) {
		content := IterationRetrospectiveTemplate()
		if !strings.Contains(content, "# Retrospective") {
			t.Error("Retrospective template should have title")
		}
	})
}

// TestSpikeTemplate verifies FR-27: spike template content.
func TestSpikeTemplate(t *testing.T) {
	content := SpikeTemplate("Redis vs Memcached")
	if !strings.Contains(content, "Spike: Redis vs Memcached") {
		t.Error("Spike template should contain title")
	}
	if !strings.Contains(content, "## Question") {
		t.Error("Spike template should have Question section")
	}
	if !strings.Contains(content, "## Findings") {
		t.Error("Spike template should have Findings section")
	}
	if !strings.Contains(content, "## Recommendation") {
		t.Error("Spike template should have Recommendation section")
	}
}

// TestConvergenceTemplate verifies FR-28: convergence template content.
func TestConvergenceTemplate(t *testing.T) {
	content := ConvergenceTemplate("Auth Strategy")
	if !strings.Contains(content, "Convergence: Auth Strategy") {
		t.Error("Convergence template should contain title")
	}
	if !strings.Contains(content, "## Context") {
		t.Error("Convergence template should have Context section")
	}
	if !strings.Contains(content, "## Decision Matrix") {
		t.Error("Convergence template should have Decision Matrix section")
	}
	if !strings.Contains(content, "## Recommendation") {
		t.Error("Convergence template should have Recommendation section")
	}
}

// TestBriefTemplate verifies FR-29: filled brief passes gate.
func TestBriefTemplate(t *testing.T) {
	content := BriefTemplate(
		"Build a CLI tool",
		"- CLI binary\n- Validation engine",
		"Phase 1 features",
		"TUI, MCP server",
		"Go 1.23+, single binary",
	)

	if !strings.Contains(content, "## Vision") {
		t.Error("Brief template should have Vision section")
	}
	if !strings.Contains(content, "Build a CLI tool") {
		t.Error("Brief template should contain vision text")
	}
	if !strings.Contains(content, "## Key Deliverables") {
		t.Error("Brief template should have Key Deliverables section")
	}
	if !strings.Contains(content, "## Scope") {
		t.Error("Brief template should have Scope section")
	}
}

// TestStubBriefTemplate verifies the stub brief is actually a stub.
func TestStubBriefTemplate(t *testing.T) {
	content := StubBriefTemplate()
	if !strings.Contains(content, "## Vision") {
		t.Error("Stub brief should have Vision section")
	}
	if !strings.Contains(content, "<!-- ") {
		t.Error("Stub brief should contain HTML comment placeholders")
	}
}

// TestClaudeAdapterTemplate verifies FR-15: adapter references .mind/.
func TestClaudeAdapterTemplate(t *testing.T) {
	content := ClaudeAdapterTemplate()
	if !strings.Contains(content, ".mind/") {
		t.Error("Claude adapter should reference .mind/")
	}
	if !strings.Contains(content, ".mind/CLAUDE.md") {
		t.Error("Claude adapter should reference .mind/CLAUDE.md")
	}
}

// TestMindTomlTemplate verifies the generated mind.toml is valid.
func TestMindTomlTemplate(t *testing.T) {
	content := MindTomlTemplate("my-project", "")
	if !strings.Contains(content, `name = "my-project"`) {
		t.Error("mind.toml should contain project name")
	}
	if !strings.Contains(content, `schema = "mind/v1.0"`) {
		t.Error("mind.toml should have schema version")
	}
	if !strings.Contains(content, "generation = 1") {
		t.Error("mind.toml should have generation = 1")
	}
	if strings.Contains(content, "[framework]") {
		t.Error("mind.toml should NOT contain [framework] when version is empty")
	}
}

// TestMindTomlTemplate_WithFramework verifies [framework] section is included when version is provided.
func TestMindTomlTemplate_WithFramework(t *testing.T) {
	content := MindTomlTemplate("my-project", "2026.03.1")
	if !strings.Contains(content, `[framework]`) {
		t.Error("mind.toml should contain [framework] section")
	}
	if !strings.Contains(content, `version = "2026.03.1"`) {
		t.Error("mind.toml should contain framework version")
	}
	if !strings.Contains(content, `mode = "standalone"`) {
		t.Error("mind.toml should contain mode = standalone")
	}
}

// TestIndexEntry verifies blueprint index entry format.
func TestIndexEntry(t *testing.T) {
	entry := IndexEntry(4, "auth-system", "04-auth-system.md")
	if !strings.Contains(entry, "BP-04") {
		t.Error("Index entry should contain BP-NN prefix")
	}
	if !strings.Contains(entry, "04-auth-system.md") {
		t.Error("Index entry should contain filename")
	}
	if !strings.HasPrefix(entry, "- [") {
		t.Error("Index entry should be a markdown list item")
	}
}

// TestStubDocument verifies generic stub document.
func TestStubDocument(t *testing.T) {
	content := StubDocument("Requirements")
	if !strings.Contains(content, "# Requirements") {
		t.Error("Stub should have title")
	}
	if !strings.Contains(content, "<!-- ") {
		t.Error("Stub should have placeholder comment")
	}
}
