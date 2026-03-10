package fs

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

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

	return IsStubContent(content), nil
}

// IsStubContent checks if content is a stub (only template boilerplate).
func IsStubContent(content []byte) bool {
	realLines := 0
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isBoilerplateLine(line) {
			continue
		}
		realLines++
	}
	return realLines <= 2
}

func isBoilerplateLine(line string) bool {
	if line == "" {
		return true
	}
	if strings.HasPrefix(line, "#") {
		return true
	}
	if strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "-->") {
		return true
	}
	if strings.HasPrefix(line, ">") {
		return true
	}
	if isTableSeparator(line) {
		return true
	}
	if isPlaceholderRow(line) {
		return true
	}
	return false
}

func isTableSeparator(line string) bool {
	if !strings.HasPrefix(line, "|") {
		return false
	}
	cleaned := strings.ReplaceAll(line, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned == ""
}

func isPlaceholderRow(line string) bool {
	return strings.HasPrefix(line, "|") &&
		strings.Contains(line, "<!--") &&
		strings.Contains(line, "-->")
}
