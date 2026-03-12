package render

import (
	"encoding/json"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

func jsonRenderer() *Renderer {
	return New(ModeJSON, 80)
}

func TestRenderHealthJSON(t *testing.T) {
	r := jsonRenderer()
	health := &domain.ProjectHealth{
		Project: domain.Project{Name: "test", Root: "/tmp/test"},
		Brief:   domain.Brief{Exists: true, GateResult: domain.BriefPresent},
		Zones: map[domain.Zone]domain.ZoneHealth{
			domain.ZoneSpec: {Zone: domain.ZoneSpec, Total: 3, Present: 3, Complete: 2, Stubs: 1},
		},
		Warnings:    []string{"test warning"},
		Suggestions: []string{"test suggestion"},
	}

	output := r.RenderHealth(health)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	requiredFields := []string{"project", "brief", "zones", "warnings", "suggestions"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q in JSON output", field)
		}
	}

	// Verify project sub-fields
	project, ok := parsed["project"].(map[string]any)
	if !ok {
		t.Fatal("project field is not an object")
	}
	if name, ok := project["name"].(string); !ok || name != "test" {
		t.Errorf("project.name = %v, want %q", project["name"], "test")
	}
}

func TestRenderValidationJSON(t *testing.T) {
	r := jsonRenderer()
	report := &domain.ValidationReport{
		Suite:  "docs",
		Total:  3,
		Passed: 2,
		Failed: 1,
		Checks: []domain.CheckResult{
			{ID: 1, Name: "check-1", Passed: true, Level: domain.LevelFail},
			{ID: 2, Name: "check-2", Passed: true, Level: domain.LevelFail},
			{ID: 3, Name: "check-3", Passed: false, Level: domain.LevelFail, Message: "failed"},
		},
	}

	output := r.RenderValidation(report)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	requiredFields := []string{"suite", "checks", "total", "passed", "failed"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q in JSON output", field)
		}
	}

	if suite, ok := parsed["suite"].(string); !ok || suite != "docs" {
		t.Errorf("suite = %v, want %q", parsed["suite"], "docs")
	}

	if total, ok := parsed["total"].(float64); !ok || int(total) != 3 {
		t.Errorf("total = %v, want 3", parsed["total"])
	}
}

func TestRenderReconcileResultJSON(t *testing.T) {
	r := jsonRenderer()
	result := &domain.ReconcileResult{
		Changed: []string{"doc:spec/a"},
		Stale:   map[string]string{"doc:spec/b": "dependency changed"},
		Missing: []string{"doc:spec/c"},
		Status:  domain.LockStale,
		Stats: domain.LockStats{
			Total:   3,
			Changed: 1,
			Stale:   1,
			Missing: 1,
			Clean:   0,
		},
	}

	output := r.RenderReconcileResult(result)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	requiredFields := []string{"changed", "stale", "missing", "status", "stats"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q in JSON output", field)
		}
	}

	if status, ok := parsed["status"].(string); !ok || status != string(domain.LockStale) {
		t.Errorf("status = %v, want %q", parsed["status"], domain.LockStale)
	}
}

func TestRenderDoctorJSON(t *testing.T) {
	r := jsonRenderer()
	report := &domain.DoctorReport{
		Diagnostics: []domain.Diagnostic{
			{
				Category: "framework",
				Check:    ".mind/ directory",
				Status:   domain.DiagPass,
				Message:  ".mind/ directory exists",
			},
			{
				Category: "docs",
				Check:    "spec zone",
				Status:   domain.DiagFail,
				Message:  "docs/spec missing",
				Fix:      "Create directory: docs/spec",
				AutoFix:  true,
			},
			{
				Category: "framework",
				Check:    "GitHub agents",
				Status:   domain.DiagWarn,
				Message:  ".github/agents/ not found",
			},
		},
		Summary: domain.DoctorSummary{Pass: 1, Fail: 1, Warn: 1},
	}

	output := r.RenderDoctor(report)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	requiredFields := []string{"diagnostics", "summary"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q in JSON output", field)
		}
	}

	// Verify diagnostics array
	diagnostics, ok := parsed["diagnostics"].([]any)
	if !ok {
		t.Fatal("diagnostics field is not an array")
	}
	if len(diagnostics) != 3 {
		t.Errorf("diagnostics length = %d, want 3", len(diagnostics))
	}

	// Verify first diagnostic has expected status
	if len(diagnostics) > 0 {
		diag, ok := diagnostics[0].(map[string]any)
		if !ok {
			t.Fatal("diagnostic entry is not an object")
		}
		if status, ok := diag["status"].(string); !ok || status != "pass" {
			t.Errorf("diagnostics[0].status = %v, want %q", diag["status"], "pass")
		}
	}

	// Verify summary sub-fields
	summary, ok := parsed["summary"].(map[string]any)
	if !ok {
		t.Fatal("summary field is not an object")
	}
	if pass, ok := summary["pass"].(float64); !ok || int(pass) != 1 {
		t.Errorf("summary.pass = %v, want 1", summary["pass"])
	}
}

func TestRenderHealthJSONContainsAllZones(t *testing.T) {
	r := jsonRenderer()
	zones := make(map[domain.Zone]domain.ZoneHealth)
	for _, z := range domain.AllZones {
		zones[z] = domain.ZoneHealth{Zone: z, Total: 1, Present: 1, Complete: 1}
	}
	health := &domain.ProjectHealth{
		Project: domain.Project{Name: "multi-zone"},
		Zones:   zones,
	}

	output := r.RenderHealth(health)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	zonesMap, ok := parsed["zones"].(map[string]any)
	if !ok {
		t.Fatal("zones field is not an object")
	}

	for _, z := range domain.AllZones {
		if _, ok := zonesMap[string(z)]; !ok {
			t.Errorf("missing zone %q in JSON output", z)
		}
	}
}

func TestRenderDoctorJSONStatusValues(t *testing.T) {
	r := jsonRenderer()

	statuses := []domain.DiagnosticStatus{domain.DiagPass, domain.DiagFail, domain.DiagWarn}
	for _, s := range statuses {
		report := &domain.DoctorReport{
			Diagnostics: []domain.Diagnostic{
				{Category: "test", Check: "test", Status: s, Message: "msg"},
			},
			Summary: domain.DoctorSummary{},
		}
		output := r.RenderDoctor(report)

		var parsed map[string]any
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("invalid JSON for status %q: %v", s, err)
		}

		diagnostics := parsed["diagnostics"].([]any)
		diag := diagnostics[0].(map[string]any)
		if diag["status"] != string(s) {
			t.Errorf("status = %v, want %q", diag["status"], s)
		}
	}
}
