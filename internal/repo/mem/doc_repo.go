package mem

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
)

// DocRepo is an in-memory implementation of repo.DocRepo for testing.
type DocRepo struct {
	Docs  map[string]domain.Document
	Files map[string][]byte
	Dirs  map[string]bool
}

// NewDocRepo creates an in-memory DocRepo.
func NewDocRepo() *DocRepo {
	return &DocRepo{
		Docs:  make(map[string]domain.Document),
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

// ListByZone returns all documents in a zone.
func (r *DocRepo) ListByZone(zone domain.Zone) ([]domain.Document, error) {
	var result []domain.Document
	prefix := fmt.Sprintf("docs/%s/", string(zone))
	for path, doc := range r.Docs {
		if strings.HasPrefix(path, prefix) {
			result = append(result, doc)
		}
	}
	return result, nil
}

// ListAll returns all documents.
func (r *DocRepo) ListAll() ([]domain.Document, error) {
	var result []domain.Document
	for _, doc := range r.Docs {
		result = append(result, doc)
	}
	return result, nil
}

// Read returns file content.
func (r *DocRepo) Read(relPath string) ([]byte, error) {
	data, ok := r.Files[relPath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", relPath)
	}
	return data, nil
}

// Exists checks if a file exists.
func (r *DocRepo) Exists(relPath string) bool {
	_, ok := r.Files[relPath]
	if ok {
		return true
	}
	_, ok = r.Dirs[relPath]
	return ok
}

// IsStub checks if file content is a stub.
func (r *DocRepo) IsStub(relPath string) (bool, error) {
	data, ok := r.Files[relPath]
	if !ok {
		return false, fmt.Errorf("file not found: %s", relPath)
	}
	return fs.IsStubContent(data), nil
}

// IsDir checks if a path is a directory.
func (r *DocRepo) IsDir(relPath string) bool {
	return r.Dirs[relPath]
}

// Search performs case-insensitive substring search across all stored files.
func (r *DocRepo) Search(query string) (*domain.SearchResults, error) {
	queryLower := strings.ToLower(query)
	results := &domain.SearchResults{Query: query}

	for relPath, content := range r.Files {
		if !strings.HasSuffix(relPath, ".md") {
			continue
		}
		var matches []domain.SearchMatch
		var lines []string

		scanner := bufio.NewScanner(bytes.NewReader(content))
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				match := domain.SearchMatch{
					Line: i + 1,
					Text: line,
				}
				if i > 0 {
					match.ContextBefore = lines[i-1]
				}
				if i < len(lines)-1 {
					match.ContextAfter = lines[i+1]
				}
				matches = append(matches, match)
			}
		}

		if len(matches) > 0 {
			results.Results = append(results.Results, domain.SearchFileResult{
				Path:    relPath,
				Matches: matches,
			})
			results.TotalMatches += len(matches)
			results.FilesMatched++
		}
	}

	return results, nil
}
