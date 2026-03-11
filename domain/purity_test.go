package domain_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDomainPurity verifies NFR-4: domain/ has zero external imports.
// Only Go standard library imports are allowed. Specifically banned:
// os, filepath, io, net, and any third-party packages.
func TestDomainPurity(t *testing.T) {
	// Banned import prefixes for the domain package
	banned := []string{
		"os",
		"path/filepath",
		"io",
		"net",
		"github.com/",
		"golang.org/",
	}

	fset := token.NewFileSet()
	domainDir := "." // Since this test is in the domain package

	entries, err := os.ReadDir(domainDir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", domainDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		// Skip test files themselves
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(domainDir, entry.Name())
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("Parse %s: %v", path, err)
			continue
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, b := range banned {
				if importPath == b || strings.HasPrefix(importPath, b) {
					t.Errorf("NFR-4 violation: %s imports %q (banned: %s)", entry.Name(), importPath, b)
				}
			}
		}
	}
}
