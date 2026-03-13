package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Manifest tracks which files in .mind/ are materialized vs. project-authored.
// Format: TOML per BP-03 §2.2.
type Manifest struct {
	Manifest ManifestMeta    `toml:"manifest"`
	Entries  []ManifestEntry `toml:"entries"`
}

// ManifestMeta holds the [manifest] section metadata.
type ManifestMeta struct {
	Version       string `toml:"version"`
	MaterializedAt string `toml:"materialized_at"`
}

// ManifestEntry tracks a single file in .mind/.
type ManifestEntry struct {
	Path     string `toml:"path"`     // Relative to .mind/
	Source   string `toml:"source"`   // "global" or "project"
	Checksum string `toml:"checksum"` // SHA-256 hex digest
}

// ManifestFileName is the name of the manifest file within .mind/.
const ManifestFileName = ".framework-manifest"

// NewManifest creates a new manifest with the given version.
func NewManifest(version string) *Manifest {
	return &Manifest{
		Manifest: ManifestMeta{
			Version:       version,
			MaterializedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// Add appends an entry to the manifest.
func (m *Manifest) Add(path string, source ArtifactSource, checksum string) {
	m.Entries = append(m.Entries, ManifestEntry{
		Path:     path,
		Source:   string(source),
		Checksum: checksum,
	})
}

// FindEntry returns the manifest entry for a given path, or nil.
func (m *Manifest) FindEntry(path string) *ManifestEntry {
	for i := range m.Entries {
		if m.Entries[i].Path == path {
			return &m.Entries[i]
		}
	}
	return nil
}

// ReadManifest parses a .framework-manifest file.
func ReadManifest(mindDir string) (*Manifest, error) {
	path := filepath.Join(mindDir, ManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", ManifestFileName, err)
	}
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ManifestFileName, err)
	}
	return &m, nil
}

// WriteManifest writes the manifest atomically.
func WriteManifest(mindDir string, m *Manifest) error {
	data, err := toml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshaling %s: %w", ManifestFileName, err)
	}
	path := filepath.Join(mindDir, ManifestFileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", ManifestFileName, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming %s: %w", ManifestFileName, err)
	}
	return nil
}
