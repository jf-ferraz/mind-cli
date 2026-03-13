// Package framework implements the framework lifecycle operations:
// install, status, diff, lock management, and doctor checks.
package framework

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/pelletier/go-toml/v2"
)

// DefaultGlobalDir returns the default global framework directory.
func DefaultGlobalDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mind")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mind")
}

// ArtifactKinds are the subdirectories of .mind/ that hold framework artifacts.
var ArtifactKinds = []string{"agents", "skills", "commands", "conventions"}

// InstallResult is returned by Install.
type InstallResult struct {
	Version       string `json:"version"`
	Source        string `json:"source"`
	ArtifactCount int    `json:"artifact_count"`
	Overwritten   bool   `json:"overwritten"`
}

// Install copies framework artifacts from source to the global directory.
// source must be a local path to a .mind/ directory or a directory containing
// the artifact kind subdirectories.
func Install(source string, globalDir string, force bool) (*InstallResult, error) {
	if globalDir == "" {
		globalDir = DefaultGlobalDir()
	}

	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("source not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source must be a directory: %s", source)
	}

	lockPath := filepath.Join(globalDir, "framework.lock")
	alreadyInstalled := false
	if _, err := os.Stat(lockPath); err == nil {
		alreadyInstalled = true
		if !force {
			return nil, fmt.Errorf("framework already installed at %s (use --force to overwrite)", globalDir)
		}
	}

	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create global directory: %w", err)
	}

	count := 0
	checksums := make(map[string]string)

	for _, kind := range ArtifactKinds {
		srcDir := filepath.Join(source, kind)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			continue
		}

		dstDir := filepath.Join(globalDir, kind)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create %s/: %w", kind, err)
		}

		err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(source, path)
			dstPath := filepath.Join(globalDir, relPath)

			if d.IsDir() {
				return os.MkdirAll(dstPath, 0755)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}

			hash := sha256.Sum256(data)
			checksums[relPath] = "sha256:" + hex.EncodeToString(hash[:])
			count++
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("copying %s/: %w", kind, err)
		}
	}

	version := detectVersion(source)

	lock := &FrameworkLock{
		Framework: LockFramework{
			Version:     version,
			Source:      source,
			InstalledAt: time.Now().UTC().Format(time.RFC3339),
		},
		Checksums: checksums,
	}
	if err := WriteLock(lockPath, lock); err != nil {
		return nil, fmt.Errorf("writing framework.lock: %w", err)
	}

	return &InstallResult{
		Version:       version,
		Source:        source,
		ArtifactCount: count,
		Overwritten:   alreadyInstalled,
	}, nil
}

func detectVersion(source string) string {
	parent := filepath.Dir(source)
	tomlPath := filepath.Join(parent, "mind.toml")
	data, err := os.ReadFile(tomlPath)
	if err == nil {
		var cfg struct {
			Framework *domain.FrameworkConfig `toml:"framework"`
		}
		if toml.Unmarshal(data, &cfg) == nil && cfg.Framework != nil && cfg.Framework.Version != "" {
			return cfg.Framework.Version
		}
	}
	now := time.Now()
	return fmt.Sprintf("%d.%02d.1", now.Year(), now.Month())
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
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func collectFiles(dir, base string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(base, path)
		parts := strings.SplitN(relPath, string(filepath.Separator), 2)
		if len(parts) < 2 {
			return nil
		}
		isKind := false
		for _, k := range ArtifactKinds {
			if parts[0] == k {
				isKind = true
				break
			}
		}
		if !isKind {
			return nil
		}
		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		result[relPath] = hash
		return nil
	})
	return result, err
}

// StatusResult is returned by Status.
type StatusResult struct {
	Installed   bool                  `json:"installed"`
	Version     string                `json:"version,omitempty"`
	Source      string                `json:"source,omitempty"`
	InstalledAt string                `json:"installed_at,omitempty"`
	Mode        domain.DeploymentMode `json:"mode,omitempty"`
	DriftFiles  []string              `json:"drift_files,omitempty"`
}

// Status checks the installed framework status and drift.
func Status(globalDir string, projectFramework *domain.FrameworkConfig) (*StatusResult, error) {
	if globalDir == "" {
		globalDir = DefaultGlobalDir()
	}

	lockPath := filepath.Join(globalDir, "framework.lock")
	lock, err := ReadLock(lockPath)
	if err != nil {
		return &StatusResult{Installed: false}, nil
	}

	result := &StatusResult{
		Installed:   true,
		Version:     lock.Framework.Version,
		Source:      lock.Framework.Source,
		InstalledAt: lock.Framework.InstalledAt,
		Mode:        projectFramework.DeploymentModeOrDefault(),
	}

	for relPath, expectedHash := range lock.Checksums {
		absPath := filepath.Join(globalDir, relPath)
		actualHash, err := hashFile(absPath)
		if err != nil || actualHash != expectedHash {
			result.DriftFiles = append(result.DriftFiles, relPath)
		}
	}
	sort.Strings(result.DriftFiles)

	return result, nil
}

// DiffEntry represents a single difference between project and global.
type DiffEntry struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// DiffResult is returned by Diff.
type DiffResult struct {
	Entries    []DiffEntry `json:"entries,omitempty"`
	HasDiff    bool        `json:"has_diff"`
	ProjectDir string      `json:"project_dir"`
	GlobalDir  string      `json:"global_dir"`
}

// Diff compares project .mind/ against global ~/.config/mind/.
func Diff(projectMindDir, globalDir string) (*DiffResult, error) {
	if globalDir == "" {
		globalDir = DefaultGlobalDir()
	}

	result := &DiffResult{
		ProjectDir: projectMindDir,
		GlobalDir:  globalDir,
	}

	globalFiles, err := collectFiles(globalDir, globalDir)
	if err != nil {
		return nil, fmt.Errorf("scanning global dir: %w", err)
	}

	projectFiles, err := collectFiles(projectMindDir, projectMindDir)
	if err != nil {
		return nil, fmt.Errorf("scanning project dir: %w", err)
	}

	for path, globalHash := range globalFiles {
		projHash, exists := projectFiles[path]
		if !exists {
			result.Entries = append(result.Entries, DiffEntry{Path: path, Status: "missing"})
		} else if projHash != globalHash {
			result.Entries = append(result.Entries, DiffEntry{Path: path, Status: "modified"})
		}
	}

	for path := range projectFiles {
		if _, exists := globalFiles[path]; !exists {
			result.Entries = append(result.Entries, DiffEntry{Path: path, Status: "extra"})
		}
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Path < result.Entries[j].Path
	})
	result.HasDiff = len(result.Entries) > 0

	return result, nil
}
