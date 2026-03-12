package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-70: Lock file location is mind.lock in project root.
func TestLockRepo_FR70_LockFileLocation(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock := &domain.LockFile{
		GeneratedAt: time.Now().UTC().Truncate(time.Second),
		Status:      domain.LockClean,
		Entries:     map[string]domain.LockEntry{},
	}
	if err := repo.Write(lock); err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(dir, "mind.lock")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("mind.lock should exist at %s: %v", expectedPath, err)
	}
}

// FR-71: Lock file JSON schema has all required fields.
func TestLockRepo_FR71_JSONSchema(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	now := time.Now().UTC().Truncate(time.Second)
	lock := &domain.LockFile{
		GeneratedAt: now,
		Status:      domain.LockStale,
		Stats: domain.LockStats{
			Total:      5,
			Changed:    1,
			Stale:      2,
			Missing:    0,
			Undeclared: 1,
			Clean:      2,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/requirements": {
				ID:          "doc:spec/requirements",
				Path:        "docs/spec/requirements.md",
				Hash:        "sha256:abc123def456",
				Size:        1024,
				ModTime:     now,
				Stale:       false,
				StaleReason: "",
				IsStub:      false,
				Status:      domain.EntryChanged,
			},
			"doc:spec/architecture": {
				ID:          "doc:spec/architecture",
				Path:        "docs/spec/architecture.md",
				Hash:        "sha256:def456ghi789",
				Size:        2048,
				ModTime:     now,
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/requirements",
				IsStub:      true,
				Status:      domain.EntryPresent,
			},
		},
	}

	if err := repo.Write(lock); err != nil {
		t.Fatal(err)
	}

	// Read raw JSON and verify structure
	data, err := os.ReadFile(filepath.Join(dir, "mind.lock"))
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("mind.lock is not valid JSON: %v", err)
	}

	// Verify top-level keys
	requiredKeys := []string{"generated_at", "status", "stats", "entries"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("mind.lock missing top-level key: %s", key)
		}
	}

	// Verify entries are keyed by document ID
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw["entries"], &entries); err != nil {
		t.Fatalf("entries is not a map: %v", err)
	}
	if _, ok := entries["doc:spec/requirements"]; !ok {
		t.Error("entries should be keyed by document ID")
	}

	// Verify entry has all required fields
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(entries["doc:spec/architecture"], &entry); err != nil {
		t.Fatal(err)
	}

	entryKeys := []string{"id", "path", "hash", "size", "mod_time", "stale", "stale_reason", "is_stub", "status"}
	for _, key := range entryKeys {
		if _, ok := entry[key]; !ok {
			t.Errorf("entry missing field: %s", key)
		}
	}

	// Verify stats has all required fields
	var stats map[string]json.RawMessage
	if err := json.Unmarshal(raw["stats"], &stats); err != nil {
		t.Fatal(err)
	}
	statsKeys := []string{"total", "changed", "stale", "missing", "undeclared", "clean"}
	for _, key := range statsKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("stats missing field: %s", key)
		}
	}
}

// FR-72: Round-trip produces byte-identical output.
func TestLockRepo_FR72_ByteIdenticalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	now := time.Now().UTC().Truncate(time.Second)
	lock := &domain.LockFile{
		GeneratedAt: now,
		Status:      domain.LockStale,
		Stats: domain.LockStats{
			Total: 3, Changed: 1, Stale: 1, Clean: 1,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/a": {
				ID: "doc:spec/a", Path: "docs/spec/a.md",
				Hash: "sha256:aaa", Size: 100, ModTime: now,
				Status: domain.EntryChanged,
			},
			"doc:spec/b": {
				ID: "doc:spec/b", Path: "docs/spec/b.md",
				Hash: "sha256:bbb", Size: 200, ModTime: now,
				Stale: true, StaleReason: "dependency changed",
				Status: domain.EntryPresent,
			},
			"doc:spec/c": {
				ID: "doc:spec/c", Path: "docs/spec/c.md",
				Hash: "sha256:ccc", Size: 300, ModTime: now,
				Status: domain.EntryUnchanged,
			},
		},
	}

	repo.Write(lock)

	// Read raw bytes
	lockPath := filepath.Join(dir, "mind.lock")
	bytes1, _ := os.ReadFile(lockPath)

	// Parse and re-write
	parsed, _ := repo.Read()
	repo.Write(parsed)

	// Read raw bytes again
	bytes2, _ := os.ReadFile(lockPath)

	if string(bytes1) != string(bytes2) {
		t.Errorf("round-trip produced different output:\n--- first ---\n%s\n--- second ---\n%s", bytes1, bytes2)
	}
}

// FR-73: Atomic write via temp file.
func TestLockRepo_FR73_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock := &domain.LockFile{
		GeneratedAt: time.Now().UTC(),
		Status:      domain.LockClean,
		Entries:     map[string]domain.LockEntry{},
	}

	if err := repo.Write(lock); err != nil {
		t.Fatal(err)
	}

	// Temp file should not remain
	tmpPath := filepath.Join(dir, "mind.lock.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("mind.lock.tmp should not exist after write")
	}

	// Lock file should exist and be valid JSON
	lockPath := filepath.Join(dir, "mind.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Error("mind.lock should contain valid JSON")
	}
}

// Read on corrupted JSON returns error.
func TestLockRepo_ReadCorruptedJSON(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "mind.lock")
	os.WriteFile(lockPath, []byte("{corrupted json"), 0644)

	repo := NewLockRepo(dir)
	_, err := repo.Read()
	if err == nil {
		t.Error("expected error for corrupted JSON")
	}
}

// Write overwrites existing lock file.
func TestLockRepo_WriteOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock1 := &domain.LockFile{
		GeneratedAt: time.Now().UTC().Truncate(time.Second),
		Status:      domain.LockClean,
		Stats:       domain.LockStats{Total: 1, Clean: 1},
		Entries:     map[string]domain.LockEntry{},
	}
	repo.Write(lock1)

	lock2 := &domain.LockFile{
		GeneratedAt: time.Now().UTC().Truncate(time.Second).Add(time.Minute),
		Status:      domain.LockStale,
		Stats:       domain.LockStats{Total: 3, Stale: 1, Clean: 2},
		Entries:     map[string]domain.LockEntry{},
	}
	repo.Write(lock2)

	read, _ := repo.Read()
	if read.Status != domain.LockStale {
		t.Errorf("Status = %q, want STALE (second write should overwrite)", read.Status)
	}
}

// Lock file ends with newline.
func TestLockRepo_EndsWithNewline(t *testing.T) {
	dir := t.TempDir()
	repo := NewLockRepo(dir)

	lock := &domain.LockFile{
		GeneratedAt: time.Now().UTC(),
		Status:      domain.LockClean,
		Entries:     map[string]domain.LockEntry{},
	}
	repo.Write(lock)

	data, _ := os.ReadFile(filepath.Join(dir, "mind.lock"))
	if len(data) == 0 {
		t.Fatal("lock file is empty")
	}
	if data[len(data)-1] != '\n' {
		t.Error("lock file should end with newline")
	}
}
