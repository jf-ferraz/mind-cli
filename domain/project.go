package domain

import (
	"fmt"
	"regexp"
	"time"
)

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
	Framework  *FrameworkConfig               `toml:"framework,omitempty"`
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

// DeploymentMode represents the framework deployment mode.
type DeploymentMode string

const (
	// ModeStandalone means the project manages its own .mind/ artifacts.
	ModeStandalone DeploymentMode = "standalone"
	// ModeThin means the project resolves artifacts from global config.
	ModeThin DeploymentMode = "thin"
)

// FrameworkConfig declares a project's relationship to the shared framework.
// If nil (absent from mind.toml), the project is in standalone mode.
type FrameworkConfig struct {
	Version string `toml:"version"`
	Mode    string `toml:"mode,omitempty"`
}

// DeploymentModeOrDefault returns the deployment mode, defaulting to standalone.
func (fc *FrameworkConfig) DeploymentModeOrDefault() DeploymentMode {
	if fc == nil || fc.Mode == "" || fc.Mode == string(ModeStandalone) {
		return ModeStandalone
	}
	if fc.Mode == string(ModeThin) {
		return ModeThin
	}
	return ModeStandalone
}

// calVerPattern matches CalVer format YYYY.MM.N
var calVerPattern = regexp.MustCompile(`^\d{4}\.\d{2}\.\d+$`)

// ValidateFrameworkConfig checks that FrameworkConfig fields are valid.
// Returns nil if fc is nil (absent section is valid — standalone mode).
func ValidateFrameworkConfig(fc *FrameworkConfig) error {
	if fc == nil {
		return nil
	}
	if fc.Version == "" {
		return fmt.Errorf("framework.version is required when [framework] section is present")
	}
	if !calVerPattern.MatchString(fc.Version) {
		return fmt.Errorf("framework.version %q does not match CalVer format YYYY.MM.N", fc.Version)
	}
	if fc.Mode != "" && fc.Mode != string(ModeStandalone) && fc.Mode != string(ModeThin) {
		return fmt.Errorf("framework.mode %q is not valid; must be %q or %q", fc.Mode, ModeStandalone, ModeThin)
	}
	return nil
}
