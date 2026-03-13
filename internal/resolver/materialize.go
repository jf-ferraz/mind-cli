package resolver

import (
"fmt"
"io"
"os"
"path/filepath"

"github.com/jf-ferraz/mind-cli/internal/framework"
)

// MaterializeResult is returned by Materialize.
type MaterializeResult struct {
	Version        string `json:"version"`
	TotalArtifacts int    `json:"total_artifacts"`
	Copied         int    `json:"copied"`
	ProjectKept    int    `json:"project_kept"`
}

// Materialize populates the project's .mind/ directory from resolved artifacts.
// Project-override files are preserved (never overwritten). Global artifacts are
// copied to the project. A .framework-manifest is written on success.
func (r *Resolver) Materialize(version string) (*MaterializeResult, error) {
manifest := NewManifest(version)
result := &MaterializeResult{Version: version}

for _, kind := range AllKinds() {
artifacts, err := r.List(kind)
if err != nil {
return nil, fmt.Errorf("listing %s: %w", kind, err)
}

for _, a := range artifacts {
relPath := filepath.Join(string(a.Kind), a.Name)
targetPath := filepath.Join(r.projectDir, string(a.Kind), a.Name)

if a.Source == SourceProject {
// Project override — already in place, just track
manifest.Add(relPath, SourceProject, a.Checksum)
result.ProjectKept++
} else {
// Global artifact — copy to project
if err := copyFile(a.Path, targetPath); err != nil {
return nil, fmt.Errorf("copying %s: %w", relPath, err)
}
// Compute checksum of the copied file
checksum, err := hashFile(targetPath)
if err != nil {
return nil, fmt.Errorf("hashing copied file %s: %w", relPath, err)
}
manifest.Add(relPath, SourceGlobal, checksum)
result.Copied++
}
result.TotalArtifacts++
}
}

if err := WriteManifest(r.projectDir, manifest); err != nil {
return nil, err
}

return result, nil
}

// UpdateResult is returned by Update.
type UpdateResult struct {
Version  string   `json:"version"`
Updated  []string `json:"updated"`
Added    []string `json:"added"`
Removed  []string `json:"removed"`
Kept     int      `json:"kept"`
LockPath string   `json:"-"`
}

// Update re-materializes the project's .mind/ from the global framework,
// updating only changed artifacts. Project overrides are preserved.
//
// Unlike Materialize, Update scans global artifacts directly and consults the
// manifest to distinguish genuine project overrides from previously-materialized
// global files (which also live in the project dir after materialization).
func (r *Resolver) Update(version string) (*UpdateResult, error) {
	// Read existing manifest to detect changes
	oldManifest, err := ReadManifest(r.projectDir)
	if err != nil {
		// No existing manifest — do a full materialize
		mResult, err := r.Materialize(version)
		if err != nil {
			return nil, err
		}
		return &UpdateResult{
			Version: mResult.Version,
			Added:   make([]string, mResult.Copied),
			Kept:    mResult.ProjectKept,
		}, nil
	}

	// Build lookup of old entries by path
	oldEntries := make(map[string]*ManifestEntry, len(oldManifest.Entries))
	for i := range oldManifest.Entries {
		oldEntries[oldManifest.Entries[i].Path] = &oldManifest.Entries[i]
	}

	newManifest := NewManifest(version)
	result := &UpdateResult{Version: version}

	// Track paths we process from global, so we can detect removals
	processed := make(map[string]bool)

	for _, kind := range AllKinds() {
		// Scan global artifacts directly — not via List(), which would show
		// previously-materialized files as SourceProject.
		globalKindDir := filepath.Join(r.globalDir, string(kind))
		globalEntries, _ := os.ReadDir(globalKindDir)

		for _, entry := range globalEntries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			relPath := filepath.Join(string(kind), name)
			globalPath := filepath.Join(globalKindDir, name)
			targetPath := filepath.Join(r.projectDir, string(kind), name)
			processed[relPath] = true

			// Check manifest to determine if project file is a genuine override
			oldEntry := oldEntries[relPath]
			if oldEntry != nil && oldEntry.Source == "project" {
				// Genuine project override — preserve it, re-track
				checksum, _ := hashFile(targetPath)
				newManifest.Add(relPath, SourceProject, checksum)
				result.Kept++
				continue
			}

			// Not in manifest yet OR was previously global — check for changes
			globalChecksum, err := hashFile(globalPath)
			if err != nil {
				return nil, fmt.Errorf("hashing global %s: %w", relPath, err)
			}

			if oldEntry != nil && oldEntry.Source == "global" && oldEntry.Checksum == globalChecksum {
				// Unchanged — keep existing, just re-track
				newManifest.Add(relPath, SourceGlobal, globalChecksum)
				result.Kept++
				continue
			}

			// New or changed — copy from global to project
			if err := copyFile(globalPath, targetPath); err != nil {
				return nil, fmt.Errorf("copying %s: %w", relPath, err)
			}
			checksum, err := hashFile(targetPath)
			if err != nil {
				return nil, fmt.Errorf("hashing %s: %w", relPath, err)
			}
			newManifest.Add(relPath, SourceGlobal, checksum)

			if oldEntry == nil {
				result.Added = append(result.Added, relPath)
			} else {
				result.Updated = append(result.Updated, relPath)
			}
		}

		// Also re-track project-only artifacts not in global
		projectKindDir := filepath.Join(r.projectDir, string(kind))
		projectEntries, _ := os.ReadDir(projectKindDir)
		for _, entry := range projectEntries {
			if entry.IsDir() {
				continue
			}
			relPath := filepath.Join(string(kind), entry.Name())
			if processed[relPath] {
				continue // Already handled above (exists in global)
			}
			// This file only exists in project — genuine project artifact
			oldEntry := oldEntries[relPath]
			if oldEntry != nil && oldEntry.Source == "project" {
				checksum, _ := hashFile(filepath.Join(projectKindDir, entry.Name()))
				newManifest.Add(relPath, SourceProject, checksum)
				result.Kept++
				processed[relPath] = true
			}
		}
	}

	// Detect removals: entries in old manifest that are no longer in global
	for path, entry := range oldEntries {
		if !processed[path] && entry.Source == "global" {
			// Global artifact removed from canonical — remove from project
			targetPath := filepath.Join(r.projectDir, path)
			os.Remove(targetPath)
			result.Removed = append(result.Removed, path)
		}
	}

if err := WriteManifest(r.projectDir, newManifest); err != nil {
return nil, err
}

// Update framework.lock if it exists in the global dir
lockPath := filepath.Join(r.globalDir, "framework.lock")
if _, err := os.Stat(lockPath); err == nil {
lock, err := framework.ReadLock(lockPath)
if err == nil {
result.LockPath = lockPath
// Update project's lock reference
			projectLockPath := filepath.Join(filepath.Dir(r.projectDir), "framework.lock")
			_ = framework.WriteLock(projectLockPath, lock)
		}
	}

	return result, nil
}

// copyFile copies src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, dst)
}
