package domain

// Zone represents one of the 5 documentation zones.
type Zone string

const (
	ZoneSpec       Zone = "spec"
	ZoneBlueprints Zone = "blueprints"
	ZoneState      Zone = "state"
	ZoneIterations Zone = "iterations"
	ZoneKnowledge  Zone = "knowledge"
)

// AllZones returns all zones in display order.
var AllZones = []Zone{ZoneSpec, ZoneBlueprints, ZoneState, ZoneIterations, ZoneKnowledge}

// ValidZone returns true if the given string is a valid zone name.
func ValidZone(s string) bool {
	for _, z := range AllZones {
		if string(z) == s {
			return true
		}
	}
	return false
}

// ZoneNames returns zone names as strings for error messages.
func ZoneNames() []string {
	names := make([]string, len(AllZones))
	for i, z := range AllZones {
		names[i] = string(z)
	}
	return names
}
