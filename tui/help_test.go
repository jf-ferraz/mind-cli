package tui

import (
	"strings"
	"testing"
)

func TestHelpModel_ViewGlobalKeys(t *testing.T) {
	h := NewHelpModel()
	h.SetSize(80, 40)

	output := h.View(TabStatus)

	if !strings.Contains(output, "Global Keys") {
		t.Error("expected 'Global Keys' heading")
	}
	// Check global key entries
	globalKeys := []string{"1-5", "Tab/Shift+Tab", "quit", "Ctrl+C"}
	for _, key := range globalKeys {
		if !strings.Contains(output, key) {
			t.Errorf("expected global key '%s' in help output", key)
		}
	}
}

// FR-117: Tab-specific keys change per active tab.
func TestHelpModel_TabSpecificKeys(t *testing.T) {
	h := NewHelpModel()
	h.SetSize(80, 40)

	tests := []struct {
		tab      TabID
		expected []string
	}{
		{TabDocs, []string{"a/s/b/t/i/k", "search", "preview", "$EDITOR"}},
		{TabIterations, []string{"a/n/e/b/r", "expand"}},
		{TabChecks, []string{"expand", "detail", "re-run"}},
		{TabQuality, []string{"select data point"}},
	}

	for _, tt := range tests {
		output := h.View(tt.tab)
		if !strings.Contains(output, TabNames[tt.tab]) {
			t.Errorf("tab %d: expected tab name '%s' in help", tt.tab, TabNames[tt.tab])
		}
		for _, exp := range tt.expected {
			if !strings.Contains(output, exp) {
				t.Errorf("tab %d: expected '%s' in help output", tt.tab, exp)
			}
		}
	}
}

func TestTabSpecificHelp_AllTabs(t *testing.T) {
	for i := 0; i < TabCount; i++ {
		keys := tabSpecificHelp(TabID(i))
		if keys == nil && TabID(i) != TabStatus {
			t.Errorf("tabSpecificHelp(%d) returned nil", i)
		}
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		n        int
		expected string
	}{
		{"abc", 5, "abc  "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
		{"", 3, "   "},
	}
	for _, tt := range tests {
		result := padRight(tt.input, tt.n)
		if result != tt.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.n, result, tt.expected)
		}
	}
}
