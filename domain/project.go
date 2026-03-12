package domain

import "time"

// Project represents a Mind Framework project detected on disk.
type Project struct {
	Root      string  `json:"root"`
	Name      string  `json:"name"`
	Config    *Config `json:"-"`
	Framework string  `json:"framework_version,omitempty"`
	DocsRoot  string  `json:"-"`
	MindRoot  string  `json:"-"`
}

// Config represents the parsed mind.toml manifest.
type Config struct {
	Manifest   Manifest                       `toml:"manifest"`
	Project    ProjectMeta                    `toml:"project"`
	Profiles   Profiles                       `toml:"profiles"`
	Documents  map[string]map[string]DocEntry `toml:"documents"`
	Governance Governance                     `toml:"governance"`
	Graph      []GraphEdge                    `toml:"graph"`
}

// Manifest tracks schema version and update time.
type Manifest struct {
	Schema     string          `toml:"schema"`
	Generation int             `toml:"generation"`
	Updated    time.Time       `toml:"updated"`
	Invariants map[string]bool `toml:"invariants"`
}

// ProjectMeta holds project-level metadata.
type ProjectMeta struct {
	Name        string      `toml:"name"`
	Description string      `toml:"description"`
	Type        string      `toml:"type"`
	Stack       StackConfig `toml:"stack"`
	Commands    CmdConfig   `toml:"commands"`
}

// StackConfig holds technology stack info.
type StackConfig struct {
	Language  string `toml:"language"`
	Framework string `toml:"framework"`
	Testing   string `toml:"testing"`
}

// CmdConfig holds shell commands for build/test/lint.
type CmdConfig struct {
	Dev       string `toml:"dev"`
	Test      string `toml:"test"`
	Lint      string `toml:"lint"`
	Typecheck string `toml:"typecheck"`
	Build     string `toml:"build"`
}

// Profiles holds active profile names.
type Profiles struct {
	Active []string `toml:"active"`
}

// DocEntry represents a document in mind.toml.
type DocEntry struct {
	ID     string `toml:"id"`
	Path   string `toml:"path"`
	Zone   string `toml:"zone"`
	Status string `toml:"status"`
}

// Governance holds project governance settings.
type Governance struct {
	MaxRetries     int    `toml:"max-retries"`
	ReviewPolicy   string `toml:"review-policy"`
	CommitPolicy   string `toml:"commit-policy"`
	BranchStrategy string `toml:"branch-strategy"`
	DefaultBranch  string `toml:"default-branch"`
}
