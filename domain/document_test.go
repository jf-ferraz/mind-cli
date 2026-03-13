package domain

import "testing"

// TestBriefGateConstants verifies BriefGate string values match spec.
func TestBriefGateConstants(t *testing.T) {
	tests := []struct {
		name string
		gate BriefGate
		want string
	}{
		{name: "BRIEF_PRESENT", gate: BriefPresent, want: "BRIEF_PRESENT"},
		{name: "BRIEF_STUB", gate: BriefStub, want: "BRIEF_STUB"},
		{name: "BRIEF_MISSING", gate: BriefMissing, want: "BRIEF_MISSING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.gate) != tt.want {
				t.Errorf("BriefGate = %q, want %q", string(tt.gate), tt.want)
			}
		})
	}
}

// TestDocStatusConstants verifies DocStatus string values.
func TestDocStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status DocStatus
		want   string
	}{
		{name: "draft", status: DocDraft, want: "draft"},
		{name: "active", status: DocActive, want: "active"},
		{name: "complete", status: DocComplete, want: "complete"},
		{name: "stub", status: DocStub, want: "stub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("DocStatus = %q, want %q", string(tt.status), tt.want)
			}
		})
	}
}

// TestRequestTypeConstants verifies RequestType string values.
func TestRequestTypeConstants(t *testing.T) {
	tests := []struct {
		name    string
		reqType RequestType
		want    string
	}{
		{name: "NEW_PROJECT", reqType: TypeNewProject, want: "NEW_PROJECT"},
		{name: "BUG_FIX", reqType: TypeBugFix, want: "BUG_FIX"},
		{name: "ENHANCEMENT", reqType: TypeEnhancement, want: "ENHANCEMENT"},
		{name: "REFACTOR", reqType: TypeRefactor, want: "REFACTOR"},
		{name: "COMPLEX_NEW", reqType: TypeComplexNew, want: "COMPLEX_NEW"},
		{name: "DIAGNOSE", reqType: TypeDiagnose, want: "DIAGNOSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.reqType) != tt.want {
				t.Errorf("RequestType = %q, want %q", string(tt.reqType), tt.want)
			}
		})
	}
}

// TestIterationStatusConstants verifies IterationStatus string values.
func TestIterationStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status IterationStatus
		want   string
	}{
		{name: "in_progress", status: IterInProgress, want: "in_progress"},
		{name: "complete", status: IterComplete, want: "complete"},
		{name: "incomplete", status: IterIncomplete, want: "incomplete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("IterationStatus = %q, want %q", string(tt.status), tt.want)
			}
		})
	}
}
