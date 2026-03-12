package mem

import "github.com/jf-ferraz/mind-cli/domain"

// QualityRepo is an in-memory implementation of repo.QualityRepo for testing.
type QualityRepo struct {
	Entries []domain.QualityEntry
}

// NewQualityRepo creates an in-memory QualityRepo.
func NewQualityRepo() *QualityRepo {
	return &QualityRepo{}
}

// ReadLog returns stored quality entries.
func (r *QualityRepo) ReadLog() ([]domain.QualityEntry, error) {
	if r.Entries == nil {
		return []domain.QualityEntry{}, nil
	}
	return r.Entries, nil
}
