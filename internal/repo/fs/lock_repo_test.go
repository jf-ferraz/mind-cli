package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestLockRepo_ReadMissing(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock, err := repo.Read()
	if err != nil {
		t.Fatalf("Read missing: %v", err)
	}
	if lock != nil {
		t.Error("expected nil for missing lock file")
	}
}

func TestLockRepo_ExistsFalse(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	if repo.Exists() {
		t.Error("Exists() should return false for missing lock")
	}
}

func TestLockRepo_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	now := time.Now().UTC().Truncate(time.Second)
	lock := &domain.LockFile{
		GeneratedAt: now,
		Status:      domain.LockClean,
		Stats: domain.LockStats{
			Total: 2,
			Clean: 2,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/requirements": {
				ID:      "doc:spec/requirements",
				Path:    "docs/spec/requirements.md",
				Hash:    "sha256:abc123",
				Size:    1024,
				ModTime: now,
				Status:  domain.EntryPresent,
			},
			"doc:spec/architecture": {
				ID:      "doc:spec/architecture",
				Path:    "docs/spec/architecture.md",
				Hash:    "sha256:def456",
				Size:    2048,
				ModTime: now,
				Status:  domain.EntryPresent,
			},
		},
	}

	if err := repo.Write(lock); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !repo.Exists() {
		t.Error("Exists() should return true after write")
	}

	read, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if read.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", read.Status)
	}
	if read.Stats.Total != 2 {
		t.Errorf("Stats.Total = %d, want 2", read.Stats.Total)
	}
	if len(read.Entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(read.Entries))
	}

	entry := read.Entries["doc:spec/requirements"]
	if entry.Hash != "sha256:abc123" {
		t.Errorf("entry hash = %q", entry.Hash)
	}
}

func TestLockRepo_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock := &domain.LockFile{
		GeneratedAt: time.Now().UTC(),
		Status:      domain.LockClean,
		Entries:     map[string]domain.LockEntry{},
	}

	if err := repo.Write(lock); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify temp file is cleaned up (renamed to mind.lock)
	tmpPath := filepath.Join(dir, "mind.lock.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("mind.lock.tmp should not exist after successful write")
	}

	// Verify mind.lock exists
	lockPath := filepath.Join(dir, "mind.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("mind.lock should exist: %v", err)
	}
}

func TestLockRepo_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	now := time.Now().UTC().Truncate(time.Second)
	lock := &domain.LockFile{
		GeneratedAt: now,
		Status:      domain.LockStale,
		Stats: domain.LockStats{
			Total:   3,
			Changed: 1,
			Stale:   1,
			Clean:   1,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/architecture": {
				ID:          "doc:spec/architecture",
				Path:        "docs/spec/architecture.md",
				Hash:        "sha256:aaa",
				Size:        100,
				ModTime:     now,
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/requirements",
				Status:      domain.EntryPresent,
			},
		},
	}

	if err := repo.Write(lock); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Read the raw bytes
	lockPath := filepath.Join(dir, "mind.lock")
	bytes1, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Read, then write back
	read, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if err := repo.Write(read); err != nil {
		t.Fatalf("Write back: %v", err)
	}

	// Read raw bytes again
	bytes2, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("ReadFile 2: %v", err)
	}

	// FR-72: round-trip should produce byte-identical output
	if string(bytes1) != string(bytes2) {
		t.Errorf("round-trip mismatch:\n--- first ---\n%s\n--- second ---\n%s", bytes1, bytes2)
	}
}
