package domain

import "testing"

// TestValidZone verifies FR-33 and BR-15: valid zone validation.
func TestValidZone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid zones
		{name: "spec is valid", input: "spec", want: true},
		{name: "blueprints is valid", input: "blueprints", want: true},
		{name: "state is valid", input: "state", want: true},
		{name: "iterations is valid", input: "iterations", want: true},
		{name: "knowledge is valid", input: "knowledge", want: true},
		// Invalid zones
		{name: "empty string", input: "", want: false},
		{name: "unknown zone", input: "invalid", want: false},
		{name: "uppercase", input: "SPEC", want: false},
		{name: "mixed case", input: "Spec", want: false},
		{name: "partial match", input: "spec/", want: false},
		{name: "close typo", input: "specs", want: false},
		{name: "plural", input: "knowledges", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidZone(tt.input)
			if got != tt.want {
				t.Errorf("ValidZone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestAllZones verifies that AllZones contains exactly 5 zones.
func TestAllZones(t *testing.T) {
	if len(AllZones) != 5 {
		t.Errorf("AllZones has %d zones, want 5", len(AllZones))
	}

	expected := map[Zone]bool{
		ZoneSpec: true, ZoneBlueprints: true, ZoneState: true,
		ZoneIterations: true, ZoneKnowledge: true,
	}
	for _, z := range AllZones {
		if !expected[z] {
			t.Errorf("Unexpected zone in AllZones: %q", z)
		}
	}
}

// TestZoneNames verifies ZoneNames returns string representations.
func TestZoneNames(t *testing.T) {
	names := ZoneNames()
	if len(names) != 5 {
		t.Errorf("ZoneNames() returned %d names, want 5", len(names))
	}

	// Verify all are non-empty strings
	for i, name := range names {
		if name == "" {
			t.Errorf("ZoneNames()[%d] is empty", i)
		}
	}

	// Verify correspondence with AllZones
	for i, z := range AllZones {
		if names[i] != string(z) {
			t.Errorf("ZoneNames()[%d] = %q, want %q", i, names[i], string(z))
		}
	}
}
