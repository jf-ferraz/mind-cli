package tui

import "testing"

func TestTabIDConstants(t *testing.T) {
	if TabStatus != 0 {
		t.Errorf("TabStatus = %d, want 0", TabStatus)
	}
	if TabDocs != 1 {
		t.Errorf("TabDocs = %d, want 1", TabDocs)
	}
	if TabIterations != 2 {
		t.Errorf("TabIterations = %d, want 2", TabIterations)
	}
	if TabChecks != 3 {
		t.Errorf("TabChecks = %d, want 3", TabChecks)
	}
	if TabQuality != 4 {
		t.Errorf("TabQuality = %d, want 4", TabQuality)
	}
}

func TestTabCount(t *testing.T) {
	if TabCount != 5 {
		t.Errorf("TabCount = %d, want 5", TabCount)
	}
}

func TestTabNames(t *testing.T) {
	if len(TabNames) != TabCount {
		t.Fatalf("TabNames length = %d, want %d", len(TabNames), TabCount)
	}

	expected := [TabCount]string{
		"1 Status",
		"2 Docs",
		"3 Iterations",
		"4 Check",
		"5 Quality",
	}
	for i, name := range expected {
		if TabNames[i] != name {
			t.Errorf("TabNames[%d] = %q, want %q", i, TabNames[i], name)
		}
	}
}

func TestViewStateConstants(t *testing.T) {
	if ViewLoading != 0 {
		t.Errorf("ViewLoading = %d, want 0", ViewLoading)
	}
	if ViewError != 1 {
		t.Errorf("ViewError = %d, want 1", ViewError)
	}
	if ViewEmpty != 2 {
		t.Errorf("ViewEmpty = %d, want 2", ViewEmpty)
	}
	if ViewReady != 3 {
		t.Errorf("ViewReady = %d, want 3", ViewReady)
	}
}

func TestMinDimensions(t *testing.T) {
	if MinWidth != 80 {
		t.Errorf("MinWidth = %d, want 80", MinWidth)
	}
	if MinHeight != 24 {
		t.Errorf("MinHeight = %d, want 24", MinHeight)
	}
}
