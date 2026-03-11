package mem

import (
	"encoding/json"

	"github.com/jf-ferraz/mind-cli/domain"
)

// LockRepo is an in-memory implementation of repo.LockRepo for testing.
type LockRepo struct {
	Lock *domain.LockFile
}

// NewLockRepo creates an in-memory LockRepo.
func NewLockRepo() *LockRepo {
	return &LockRepo{}
}

// Read returns a deep copy of the stored lock file.
// Returns nil, nil if no lock is stored (simulates missing mind.lock).
func (r *LockRepo) Read() (*domain.LockFile, error) {
	if r.Lock == nil {
		return nil, nil
	}
	return deepCopyLock(r.Lock), nil
}

// Write stores a deep copy of the lock file.
func (r *LockRepo) Write(lock *domain.LockFile) error {
	r.Lock = deepCopyLock(lock)
	return nil
}

// Exists returns true if a lock file is stored.
func (r *LockRepo) Exists() bool {
	return r.Lock != nil
}

// deepCopyLock creates a deep copy via JSON round-trip to prevent test mutation.
func deepCopyLock(lock *domain.LockFile) *domain.LockFile {
	data, err := json.Marshal(lock)
	if err != nil {
		return nil
	}
	var copy domain.LockFile
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil
	}
	return &copy
}
