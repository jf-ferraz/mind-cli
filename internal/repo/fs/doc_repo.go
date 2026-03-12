package fs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// Search performs case-insensitive substring search across all .md files in docs/.
// Returns matching lines with 1 line of context, grouped by file.
func (r *DocRepo) Search(query string) (*domain.SearchResults, error) {
	queryLower := strings.ToLower(query)
	results := &domain.SearchResults{Query: query}

	err := filepath.WalkDir(r.docsRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		relPath, _ := filepath.Rel(r.projectRoot, path)
		var matches []domain.SearchMatch
		var lines []string

		scanner := bufio.NewScanner(f)
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// DocRepo implements repo.DocRepo using the filesystem.
type DocRepo struct {
	projectRoot string
	docsRoot    string
}

// NewDocRepo creates a new filesystem-backed DocRepo.
func NewDocRepo(projectRoot string) *DocRepo {
	return &DocRepo{
		projectRoot: projectRoot,
		docsRoot:    filepath.Join(projectRoot, "docs"),
	}
}

func (r *DocRepo) ListByZone(zone domain.Zone) ([]domain.Document, error) {
	zoneDir := filepath.Join(r.docsRoot, string(zone))
	if _, err := os.Stat(zoneDir); os.IsNotExist(err) {
		return nil, nil
	}

	var docs []domain.Document
	err := filepath.WalkDir(zoneDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		relPath, _ := filepath.Rel(r.projectRoot, path)
		info, _ := d.Info()

		isStub, _ := r.IsStub(relPath)

		doc := domain.Document{
			Path:    relPath,
			AbsPath: path,
			Zone:    zone,
			Name:    strings.TrimSuffix(d.Name(), ".md"),
			IsStub:  isStub,
		}
		if info != nil {
			doc.Size = info.Size()
			doc.ModTime = info.ModTime()
		}
		if isStub {
			doc.Status = domain.DocStub
		} else {
			doc.Status = domain.DocComplete
		}

		docs = append(docs, doc)
		return nil
	})
	return docs, err
}

func (r *DocRepo) ListAll() ([]domain.Document, error) {
	var all []domain.Document
	for _, zone := range domain.AllZones {
		docs, err := r.ListByZone(zone)
		if err != nil {
			return nil, err
		}
		all = append(all, docs...)
	}
	return all, nil
}

func (r *DocRepo) Read(relPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(r.projectRoot, relPath))
}

func (r *DocRepo) Exists(relPath string) bool {
	_, err := os.Stat(filepath.Join(r.projectRoot, relPath))
	return err == nil
}

func (r *DocRepo) IsDir(relPath string) bool {
	info, err := os.Stat(filepath.Join(r.projectRoot, relPath))
	return err == nil && info.IsDir()
}

// IsStub implements the stub detection logic from validate-docs.sh.
// A stub is a file with only headings, HTML comments, empty lines,
// table separators, blockquotes, and placeholder rows.
func (r *DocRepo) IsStub(relPath string) (bool, error) {
	absPath := filepath.Join(r.projectRoot, relPath)

	info, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}
	if info.Size() == 0 {
		return true, nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return false, err
	}

	return repo.IsStubContent(content), nil
}
