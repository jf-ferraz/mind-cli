// Package resolver implements the unified artifact resolution engine.
// It provides a two-layer resolution chain (project → global) with
// materialization support for thin-mode deployments.
package resolver

// ArtifactKind categorizes framework artifacts.
type ArtifactKind string

const (
	KindAgents      ArtifactKind = "agents"
	KindSkills      ArtifactKind = "skills"
	KindCommands    ArtifactKind = "commands"
	KindConventions ArtifactKind = "conventions"
)

// AllKinds returns all recognized artifact kinds in scan order.
func AllKinds() []ArtifactKind {
	return []ArtifactKind{KindAgents, KindSkills, KindCommands, KindConventions}
}

// ArtifactSource indicates where a resolved artifact was found.
type ArtifactSource string

const (
	SourceProject ArtifactSource = "project"
	SourceGlobal  ArtifactSource = "global"
)

// ResolvedArtifact is the result of resolving an artifact through the chain.
type ResolvedArtifact struct {
	Path     string         `json:"path"`     // Absolute file path
	Source   ArtifactSource `json:"source"`   // "project" or "global"
	Kind     ArtifactKind   `json:"kind"`     // Artifact type category
	Name     string         `json:"name"`     // Artifact filename
	Checksum string         `json:"checksum"` // SHA-256 hex digest
}
