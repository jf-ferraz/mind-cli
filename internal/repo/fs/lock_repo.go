package fs

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
)

// LockRepo implements repo.LockRepo using the filesystem.
type LockRepo struct {
	projectRoot string
}

// NewLockRepo creates a LockRepo.
func NewLockRepo(projectRoot string) *LockRepo {
	return &LockRepo{projectRoot: projectRoot}
}

func (r *LockRepo) lockPath() string {
	return filepath.Join(r.projectRoot, "mind.lock")
}

func (r *LockRepo) tmpPath() string {
	return filepath.Join(r.projectRoot, "mind.lock.tmp")
}

// Read loads mind.lock. Returns nil, nil if the file does not exist.
func (r *LockRepo) Read() (*domain.LockFile, error) {
	data, err := os.ReadFile(r.lockPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var lock domain.LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

// Write persists the lock file atomically: write to mind.lock.tmp, then rename (FR-73).
// Go's json.MarshalIndent sorts map keys, ensuring deterministic round-trips (FR-72).
func (r *LockRepo) Write(lock *domain.LockFile) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.WriteFile(r.tmpPath(), data, 0644); err != nil {
		return err
	}
	return os.Rename(r.tmpPath(), r.lockPath())
}

// Exists returns true if mind.lock exists on disk.
func (r *LockRepo) Exists() bool {
	_, err := os.Stat(r.lockPath())
	return err == nil
}
