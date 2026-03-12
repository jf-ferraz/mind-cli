package fs

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/jf-ferraz/mind-cli/domain"
	"gopkg.in/yaml.v3"
)

// QualityRepo reads quality log entries from quality-log.yml.
type QualityRepo struct {
	projectRoot string
}

// NewQualityRepo creates a filesystem-backed QualityRepo.
func NewQualityRepo(projectRoot string) *QualityRepo {
	return &QualityRepo{projectRoot: projectRoot}
}

// ReadLog returns all quality entries ordered by date.
// Returns empty slice if the file does not exist.
func (r *QualityRepo) ReadLog() ([]domain.QualityEntry, error) {
	// Check both root and docs/knowledge locations
	paths := []string{
		filepath.Join(r.projectRoot, "quality-log.yml"),
		filepath.Join(r.projectRoot, "docs", "knowledge", "quality-log.yml"),
	}

	var data []byte
	var readErr error
	for _, p := range paths {
		data, readErr = os.ReadFile(p)
		if readErr == nil {
			break
		}
	}

	if readErr != nil {
		if os.IsNotExist(readErr) {
			return []domain.QualityEntry{}, nil
		}
		return nil, readErr
	}

	if len(data) == 0 {
		return []domain.QualityEntry{}, nil
	}

	var entries []domain.QualityEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})

	return entries, nil
}
