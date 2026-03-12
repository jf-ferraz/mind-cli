package tui

import "os"

// readFile is a simple file reader (best-effort, returns error on failure).
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
