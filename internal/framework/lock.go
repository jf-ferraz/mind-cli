package framework

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// FrameworkLock represents the framework.lock file.
type FrameworkLock struct {
	Framework LockFramework     `toml:"framework"`
	Checksums map[string]string `toml:"checksums"`
}

// LockFramework holds the [framework] section of framework.lock.
type LockFramework struct {
	Version     string `toml:"version"`
	Source      string `toml:"source"`
	InstalledAt string `toml:"installed_at"`
}

// ReadLock parses a framework.lock file.
func ReadLock(path string) (*FrameworkLock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading framework.lock: %w", err)
	}

	var lock FrameworkLock
	if err := toml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parsing framework.lock: %w", err)
	}
	return &lock, nil
}

// WriteLock writes a framework.lock file atomically.
func WriteLock(path string, lock *FrameworkLock) error {
	data, err := toml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("marshaling framework.lock: %w", err)
	}

	// Write to temp then rename for atomicity
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing framework.lock: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming framework.lock: %w", err)
	}
	return nil
}
