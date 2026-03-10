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
