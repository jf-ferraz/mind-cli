package reconcile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestHashFile_KnownContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("Hello\r\nWorld\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}

	// Verify prefix
	if hash[:7] != "sha256:" {
		t.Errorf("hash prefix = %q, want sha256:", hash[:7])
	}

	// Verify determinism
	hash2, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile second call: %v", err)
	}
	if hash != hash2 {
		t.Errorf("non-deterministic: %q != %q", hash, hash2)
	}
}

func TestHashFile_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}

	if hash != EmptyHash {
		t.Errorf("empty file hash = %q, want %q", hash, EmptyHash)
	}
}

func TestHashFile_NotFound(t *testing.T) {
	_, err := HashFile("/nonexistent/path/to/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestHashFile_Symlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.md")
	link := filepath.Join(dir, "link.md")

	content := []byte("symlink target content")
	if err := os.WriteFile(target, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	hashTarget, err := HashFile(target)
	if err != nil {
		t.Fatalf("HashFile(target): %v", err)
	}
	hashLink, err := HashFile(link)
	if err != nil {
		t.Fatalf("HashFile(link): %v", err)
	}

	if hashTarget != hashLink {
		t.Errorf("symlink hash (%q) != target hash (%q)", hashLink, hashTarget)
	}
}

func TestNeedsRehash_NilEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("content"), 0644)
	info, _ := os.Stat(path)

	if !NeedsRehash(nil, info) {
		t.Error("nil entry should need rehash")
	}
}

func TestNeedsRehash_EmptyHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("content"), 0644)
	info, _ := os.Stat(path)

	entry := &domain.LockEntry{
		Hash:    "",
		ModTime: info.ModTime(),
		Size:    info.Size(),
	}

	if !NeedsRehash(entry, info) {
		t.Error("empty hash should need rehash")
	}
}

func TestNeedsRehash_MtimeMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("content"), 0644)
	info, _ := os.Stat(path)

	entry := &domain.LockEntry{
		Hash:    "sha256:abc123",
		ModTime: info.ModTime(),
		Size:    info.Size(),
	}

	if NeedsRehash(entry, info) {
		t.Error("matching mtime+size should not need rehash")
	}
}

func TestNeedsRehash_MtimeDiffers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("content"), 0644)
	info, _ := os.Stat(path)

	entry := &domain.LockEntry{
		Hash:    "sha256:abc123",
		ModTime: info.ModTime().Add(-time.Hour),
		Size:    info.Size(),
	}

	if !NeedsRehash(entry, info) {
		t.Error("different mtime should need rehash")
	}
}

func TestNeedsRehash_SizeDiffers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("content"), 0644)
	info, _ := os.Stat(path)

	entry := &domain.LockEntry{
		Hash:    "sha256:abc123",
		ModTime: info.ModTime(),
		Size:    info.Size() + 100,
	}

	if !NeedsRehash(entry, info) {
		t.Error("different size should need rehash")
	}
}
