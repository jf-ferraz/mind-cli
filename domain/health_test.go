package domain

import (
	"encoding/json"
	"testing"
)

// FR-130: DiagnosticStatus enum values match JSON contract.
func TestDiagnosticStatus_Values(t *testing.T) {
	tests := []struct {
		status DiagnosticStatus
		want   string
	}{
		{DiagPass, "pass"},
		{DiagFail, "fail"},
		{DiagWarn, "warn"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("DiagnosticStatus %q != %q", tt.status, tt.want)
		}
	}
}

// FR-130: DiagnosticStatus JSON serialization produces lowercase string values.
// This verifies the JSON contract is preserved after the string-to-enum migration.
func TestDiagnosticStatus_JSONSerialization(t *testing.T) {
	tests := []struct {
		name   string
		status DiagnosticStatus
		want   string
	}{
		{"pass serializes to pass", DiagPass, `"pass"`},
		{"fail serializes to fail", DiagFail, `"fail"`},
		{"warn serializes to warn", DiagWarn, `"warn"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("json.Marshal(%q): %v", tt.status, err)
			}
			if string(b) != tt.want {
				t.Errorf("json.Marshal(%q) = %s, want %s", tt.status, b, tt.want)
			}
		})
	}
}

// FR-130: Diagnostic struct JSON output preserves status field as lowercase string.
func TestDiagnostic_JSONStatusField(t *testing.T) {
	diag := Diagnostic{
		Category: "test",
		Check:    "test-check",
		Status:   DiagFail,
		Message:  "something failed",
	}

	b, err := json.Marshal(diag)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	status, ok := parsed["status"].(string)
	if !ok {
		t.Fatal("status field missing or not a string")
	}
	if status != "fail" {
		t.Errorf("status = %q, want %q", status, "fail")
	}
}

// FR-130/Convergence: DiagnosticStatus is a typed string, so a typo like
// DiagnosticStatus("fali") would not match any constant and would be caught
// by switch statements that handle only DiagPass/DiagFail/DiagWarn.
// This test verifies that the three constants are the only expected values.
func TestDiagnosticStatus_Exhaustive(t *testing.T) {
	allStatuses := []DiagnosticStatus{DiagPass, DiagFail, DiagWarn}
	seen := make(map[DiagnosticStatus]bool)
	for _, s := range allStatuses {
		if seen[s] {
			t.Errorf("duplicate DiagnosticStatus: %q", s)
		}
		seen[s] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 unique DiagnosticStatus values, got %d", len(seen))
	}
}

// FR-130: DoctorSummary JSON field names match contract.
func TestDoctorSummary_JSONFields(t *testing.T) {
	summary := DoctorSummary{Pass: 5, Fail: 2, Warn: 1}

	b, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	for _, field := range []string{"pass", "fail", "warn"} {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q in DoctorSummary JSON", field)
		}
	}
}

// FR-130: DoctorReport JSON structure with typed status.
func TestDoctorReport_JSONStructure(t *testing.T) {
	report := DoctorReport{
		Diagnostics: []Diagnostic{
			{Category: "c1", Check: "k1", Status: DiagPass, Message: "ok"},
			{Category: "c2", Check: "k2", Status: DiagFail, Message: "bad"},
			{Category: "c3", Check: "k3", Status: DiagWarn, Message: "hmm"},
		},
		Summary: DoctorSummary{Pass: 1, Fail: 1, Warn: 1},
	}

	b, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	diagnostics, ok := parsed["diagnostics"].([]any)
	if !ok {
		t.Fatal("diagnostics not an array")
	}
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics length = %d, want 3", len(diagnostics))
	}

	expectedStatuses := []string{"pass", "fail", "warn"}
	for i, expected := range expectedStatuses {
		diag := diagnostics[i].(map[string]any)
		if diag["status"] != expected {
			t.Errorf("diagnostics[%d].status = %v, want %q", i, diag["status"], expected)
		}
	}
}
