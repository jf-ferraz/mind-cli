package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/jf-ferraz/mind-cli/domain"
)

// HashPrefix is the type tag prepended to all SHA-256 hashes.
const HashPrefix = "sha256:"

// EmptyHash is the SHA-256 hash of empty input.
const EmptyHash = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// HashFile computes the SHA-256 hash of a file's raw content.
// The returned hash uses the format "sha256:{64-char hex digest}".
// Accepts an absolute path. Returns an error if the file cannot be opened or read.
func HashFile(absPath string) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", absPath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash %s: %w", absPath, err)
	}

	return HashPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// NeedsRehash returns true if the file needs its hash recomputed.
// The mtime fast-path: if the lock entry exists and both mtime and size match,
// hash computation can be skipped (the stored hash is reused).
func NeedsRehash(entry *domain.LockEntry, info os.FileInfo) bool {
	if entry == nil {
		return true
	}
	if entry.Hash == "" {
		return true
	}
	return !info.ModTime().Equal(entry.ModTime) || info.Size() != entry.Size
}
