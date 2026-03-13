package resolver

import (
	"fmt"
	"io"
	"io/fs"
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

	// Copy root files (CLAUDE.md, README.md) from global if not overridden
	for _, rootFile := range RootFiles {
		globalPath := filepath.Join(r.globalDir, rootFile)
		projectPath := filepath.Join(r.projectDir, rootFile)

		// Check if project has its own version (override)
		if _, err := os.Stat(projectPath); err == nil {
			// Project has this root file — treat as override
			checksum, _ := hashFile(projectPath)
			manifest.Add(rootFile, SourceProject, checksum)
			result.ProjectKept++
			result.TotalArtifacts++
			continue
		}

		// Check if global has this root file
		if _, err := os.Stat(globalPath); err == nil {
			if err := copyFile(globalPath, projectPath); err != nil {
				return nil, fmt.Errorf("copying root file %s: %w", rootFile, err)
			}
			checksum, _ := hashFile(projectPath)
			manifest.Add(rootFile, SourceGlobal, checksum)
			result.Copied++
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
		// Scan global artifacts recursively — not via List(), which would show
		// previously-materialized files as SourceProject.
		globalKindDir := filepath.Join(r.globalDir, string(kind))
		if _, statErr := os.Stat(globalKindDir); statErr == nil {
			if walkErr := filepath.WalkDir(globalKindDir, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil || d.IsDir() {
					return walkErr
				}
				name, _ := filepath.Rel(globalKindDir, path)
				relPath := filepath.Join(string(kind), name)
				targetPath := filepath.Join(r.projectDir, string(kind), name)
				processed[relPath] = true

				// Check manifest to determine if project file is a genuine override
				oldEntry := oldEntries[relPath]
				if oldEntry != nil && oldEntry.Source == "project" {
					// Genuine project override — preserve it, re-track
					checksum, _ := hashFile(targetPath)
					newManifest.Add(relPath, SourceProject, checksum)
					result.Kept++
					return nil
				}

				// Not in manifest yet OR was previously global — check for changes
				globalChecksum, herr := hashFile(path)
				if herr != nil {
					return fmt.Errorf("hashing global %s: %w", relPath, herr)
				}

				if oldEntry != nil && oldEntry.Source == "global" && oldEntry.Checksum == globalChecksum {
					// Unchanged — keep existing, just re-track
					newManifest.Add(relPath, SourceGlobal, globalChecksum)
					result.Kept++
					return nil
				}

				// New or changed — copy from global to project
				if cerr := copyFile(path, targetPath); cerr != nil {
					return fmt.Errorf("copying %s: %w", relPath, cerr)
				}
				checksum, herr := hashFile(targetPath)
				if herr != nil {
					return fmt.Errorf("hashing %s: %w", relPath, herr)
				}
				newManifest.Add(relPath, SourceGlobal, checksum)

				if oldEntry == nil {
					result.Added = append(result.Added, relPath)
				} else {
					result.Updated = append(result.Updated, relPath)
				}
				return nil
			}); walkErr != nil {
				return nil, fmt.Errorf("scanning global %s: %w", kind, walkErr)
			}
		}

		// Also re-track project-only artifacts not in global (recursive)
		projectKindDir := filepath.Join(r.projectDir, string(kind))
		if _, statErr := os.Stat(projectKindDir); statErr == nil {
			if walkErr := filepath.WalkDir(projectKindDir, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil || d.IsDir() {
					return walkErr
				}
				name, _ := filepath.Rel(projectKindDir, path)
				relPath := filepath.Join(string(kind), name)
				if processed[relPath] {
					return nil // Already handled above (exists in global)
				}
				// This file only exists in project — genuine project artifact
				oldEntry := oldEntries[relPath]
				if oldEntry != nil && oldEntry.Source == "project" {
					checksum, _ := hashFile(path)
					newManifest.Add(relPath, SourceProject, checksum)
					result.Kept++
					processed[relPath] = true
				}
				return nil
			}); walkErr != nil {
				return nil, fmt.Errorf("scanning project %s: %w", kind, walkErr)
			}
		}
	}

	// Handle root files (CLAUDE.md, README.md)
	for _, rootFile := range RootFiles {
		globalPath := filepath.Join(r.globalDir, rootFile)
		targetPath := filepath.Join(r.projectDir, rootFile)
		processed[rootFile] = true

		oldEntry := oldEntries[rootFile]
		if oldEntry != nil && oldEntry.Source == "project" {
			checksum, _ := hashFile(targetPath)
			newManifest.Add(rootFile, SourceProject, checksum)
			result.Kept++
			continue
		}

		if _, err := os.Stat(globalPath); err != nil {
			continue // root file doesn't exist in global
		}

		globalChecksum, err := hashFile(globalPath)
		if err != nil {
			return nil, fmt.Errorf("hashing global root file %s: %w", rootFile, err)
		}

		if oldEntry != nil && oldEntry.Source == "global" && oldEntry.Checksum == globalChecksum {
			newManifest.Add(rootFile, SourceGlobal, globalChecksum)
			result.Kept++
			continue
		}

		if err := copyFile(globalPath, targetPath); err != nil {
			return nil, fmt.Errorf("copying root file %s: %w", rootFile, err)
		}
		checksum, err := hashFile(targetPath)
		if err != nil {
			return nil, fmt.Errorf("hashing copied root file %s: %w", rootFile, err)
		}
		newManifest.Add(rootFile, SourceGlobal, checksum)
		if oldEntry == nil {
			result.Added = append(result.Added, rootFile)
		} else {
			result.Updated = append(result.Updated, rootFile)
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
