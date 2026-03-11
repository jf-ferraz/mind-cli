package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-57: Hash of raw bytes with no normalization. \r\n must be preserved.
func TestHashFile_RawBytesNoNormalization(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.md")
	content := []byte("Hello\r\nWorld\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}

	// Compute expected hash of the exact bytes including \r
	h := sha256.Sum256(content)
	expected := HashPrefix + hex.EncodeToString(h[:])

	if hash != expected {
		t.Errorf("hash = %q, want %q (raw bytes including \\r)", hash, expected)
	}
}

// FR-57: Hash format must be sha256:{64-char lowercase hex}.
func TestHashFile_FormatCorrect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "format.md")
	os.WriteFile(path, []byte("test content"), 0644)

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}

	if len(hash) != 7+64 { // "sha256:" + 64 hex chars
		t.Errorf("hash length = %d, want %d (sha256: prefix + 64 hex)", len(hash), 7+64)
	}

	prefix := hash[:7]
	if prefix != "sha256:" {
		t.Errorf("hash prefix = %q, want sha256:", prefix)
	}

	// Verify hex portion is lowercase
	hexPart := hash[7:]
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("hash hex contains non-lowercase-hex char: %c in %s", c, hexPart)
			break
		}
	}
}

// FR-59: Empty file produces known SHA-256 hash.
func TestHashFile_EmptyFileKnownHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	os.WriteFile(path, []byte{}, 0644)

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}

	expectedEmpty := "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expectedEmpty {
		t.Errorf("empty file hash = %q, want %q", hash, expectedEmpty)
	}
	if hash != EmptyHash {
		t.Errorf("hash != EmptyHash constant")
	}
}

// FR-57: Hash is deterministic for identical content.
func TestHashFile_DeterministicForIdenticalContent(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "file1.md")
	path2 := filepath.Join(dir, "file2.md")
	content := []byte("identical content\n")
	os.WriteFile(path1, content, 0644)
	os.WriteFile(path2, content, 0644)

	hash1, err := HashFile(path1)
	if err != nil {
		t.Fatal(err)
	}
	hash2, err := HashFile(path2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Errorf("identical content produced different hashes: %q != %q", hash1, hash2)
	}
}

// FR-57: Different content produces different hashes.
func TestHashFile_DifferentContentDifferentHash(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "file1.md")
	path2 := filepath.Join(dir, "file2.md")
	os.WriteFile(path1, []byte("content A"), 0644)
	os.WriteFile(path2, []byte("content B"), 0644)

	hash1, _ := HashFile(path1)
	hash2, _ := HashFile(path2)

	if hash1 == hash2 {
		t.Error("different content produced same hash")
	}
}

// FR-59: Unreadable file returns error.
func TestHashFile_Unreadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noperm.md")
	os.WriteFile(path, []byte("content"), 0644)
	os.Chmod(path, 0000)
	defer os.Chmod(path, 0644) // Cleanup

	_, err := HashFile(path)
	if err == nil {
		t.Error("expected error for unreadable file")
	}
}

// BR-24: Hash computation is pure SHA-256 of raw bytes.
func TestHashFile_MatchesManualSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known.md")
	content := []byte("# Requirements\n\nFR-1: something")
	os.WriteFile(path, content, 0644)

	hash, err := HashFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Manually compute
	h := sha256.Sum256(content)
	expected := "sha256:" + hex.EncodeToString(h[:])

	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}
}

// FR-59: Symlink resolves to target content.
func TestHashFile_SymlinkResolvesToTarget(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	os.MkdirAll(subdir, 0755)

	target := filepath.Join(subdir, "target.md")
	link := filepath.Join(dir, "link.md")
	content := []byte("target content for symlink test")
	os.WriteFile(target, content, 0644)
	os.Symlink(target, link)

	hashTarget, _ := HashFile(target)
	hashLink, _ := HashFile(link)

	if hashTarget != hashLink {
		t.Errorf("symlink hash %q != target hash %q", hashLink, hashTarget)
	}
}

// FR-25/BR-25: NeedsRehash - false negatives are OK, false positives are not.
func TestNeedsRehash_MtimeTouchSameContent(t *testing.T) {
	// Touch a file (change mtime, same content) - NeedsRehash should return true
	// because mtime differs. This is a false negative (unnecessary rehash) which is acceptable.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := []byte("unchanged content")
	os.WriteFile(path, content, 0644)
	info1, _ := os.Stat(path)

	// "Touch" by rewriting same content
	os.WriteFile(path, content, 0644)
	info2, _ := os.Stat(path)

	entry := &domain.LockEntry{
		Hash:    "sha256:abc123",
		ModTime: info1.ModTime(),
		Size:    info1.Size(),
	}

	// If mtime changed, NeedsRehash should return true (false negative is OK)
	if !info1.ModTime().Equal(info2.ModTime()) {
		if !NeedsRehash(entry, info2) {
			t.Error("different mtime should trigger rehash (false negative prevention)")
		}
	}
}
