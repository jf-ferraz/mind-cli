package resolver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Resolver resolves framework artifacts through a project → global chain.
type Resolver struct {
	projectDir string // project .mind/ directory
	globalDir  string // global ~/.config/mind/ directory
}

// New creates a Resolver with the given project and global directories.
func New(projectDir, globalDir string) *Resolver {
	return &Resolver{
		projectDir: projectDir,
		globalDir:  globalDir,
	}
}

// Resolve looks up a single artifact by kind and name.
// Resolution order: project .mind/{kind}/{name} → global ~/.config/mind/{kind}/{name} → error.
func (r *Resolver) Resolve(kind ArtifactKind, name string) (*ResolvedArtifact, error) {
	// Step 1: Check project layer
	projectPath := filepath.Join(r.projectDir, string(kind), name)
	if info, err := os.Stat(projectPath); err == nil && !info.IsDir() {
		checksum, err := hashFile(projectPath)
		if err != nil {
			return nil, fmt.Errorf("hashing project artifact %s/%s: %w", kind, name, err)
		}
		return &ResolvedArtifact{
			Path:     projectPath,
			Source:   SourceProject,
			Kind:     kind,
			Name:     name,
			Checksum: checksum,
		}, nil
	}

	// Step 2: Check global layer
	globalPath := filepath.Join(r.globalDir, string(kind), name)
	if info, err := os.Stat(globalPath); err == nil && !info.IsDir() {
		checksum, err := hashFile(globalPath)
		if err != nil {
			return nil, fmt.Errorf("hashing global artifact %s/%s: %w", kind, name, err)
		}
		return &ResolvedArtifact{
			Path:     globalPath,
			Source:   SourceGlobal,
			Kind:     kind,
			Name:     name,
			Checksum: checksum,
		}, nil
	}

	// Step 3: Not found
	return nil, fmt.Errorf("artifact not found: %s/%s", kind, name)
}

// List returns all artifacts of a given kind, with project files taking
// precedence over global files of the same name.
func (r *Resolver) List(kind ArtifactKind) ([]ResolvedArtifact, error) {
	seen := make(map[string]bool)
	var results []ResolvedArtifact

	// Scan project layer first (takes precedence)
	projectKindDir := filepath.Join(r.projectDir, string(kind))
	if entries, err := os.ReadDir(projectKindDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			seen[name] = true
			fullPath := filepath.Join(projectKindDir, name)
			checksum, err := hashFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("hashing project artifact %s/%s: %w", kind, name, err)
			}
			results = append(results, ResolvedArtifact{
				Path:     fullPath,
				Source:   SourceProject,
				Kind:     kind,
				Name:     name,
				Checksum: checksum,
			})
		}
	}

	// Scan global layer (only add if not shadowed)
	globalKindDir := filepath.Join(r.globalDir, string(kind))
	if entries, err := os.ReadDir(globalKindDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if seen[name] {
				continue // project override takes precedence
			}
			fullPath := filepath.Join(globalKindDir, name)
			checksum, err := hashFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("hashing global artifact %s/%s: %w", kind, name, err)
			}
			results = append(results, ResolvedArtifact{
				Path:     fullPath,
				Source:   SourceGlobal,
				Kind:     kind,
				Name:     name,
				Checksum: checksum,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results, nil
}

// ListAll returns all artifacts across all kinds.
func (r *Resolver) ListAll() ([]ResolvedArtifact, error) {
	var all []ResolvedArtifact
	for _, kind := range AllKinds() {
		artifacts, err := r.List(kind)
		if err != nil {
			return nil, err
		}
		all = append(all, artifacts...)
	}
	return all, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
