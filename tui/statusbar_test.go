package tui

import (
	"strings"
	"testing"
)

func TestRenderStatusBar_AllTabs(t *testing.T) {
	for i := 0; i < TabCount; i++ {
		output := renderStatusBar(80, TabID(i), "")
		if output == "" {
			t.Errorf("tab %d: empty status bar", i)
		}
	}
}

// FR-96: Status bar shows context-sensitive hints.
func TestRenderStatusBar_TabSpecificHints(t *testing.T) {
	tests := []struct {
		tab      TabID
		expected []string
	}{
		{TabStatus, []string{"refresh", "help", "quit"}},
		{TabDocs, []string{"zone", "search", "preview", "edit", "help", "quit"}},
		{TabIterations, []string{"filter", "expand", "help", "quit"}},
		{TabChecks, []string{"expand", "detail", "rerun", "help", "quit"}},
		{TabQuality, []string{"navigate", "help", "quit"}},
	}

	for _, tt := range tests {
		output := renderStatusBar(100, tt.tab, "")
		for _, hint := range tt.expected {
			if !strings.Contains(output, hint) {
				t.Errorf("tab %d: expected hint '%s' in status bar", tt.tab, hint)
			}
		}
	}
}

// FR-96: Status bar shows info text when provided.
func TestRenderStatusBar_WithInfo(t *testing.T) {
	output := renderStatusBar(100, TabDocs, "3/22 docs")
	if !strings.Contains(output, "3/22 docs") {
		t.Error("expected info text '3/22 docs' in status bar")
	}
}

func TestTabHints_AllPopulated(t *testing.T) {
	for i := 0; i < TabCount; i++ {
		if tabHints[i] == "" {
			t.Errorf("tabHints[%d] is empty", i)
		}
	}
}
