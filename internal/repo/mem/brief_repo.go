package mem

import "github.com/jf-ferraz/mind-cli/domain"

// BriefRepo is an in-memory implementation of repo.BriefRepo for testing.
type BriefRepo struct {
	Brief *domain.Brief
}

// NewBriefRepo creates an in-memory BriefRepo.
func NewBriefRepo() *BriefRepo {
	return &BriefRepo{}
}

// ParseBrief returns the stored brief.
func (r *BriefRepo) ParseBrief() (*domain.Brief, error) {
	if r.Brief == nil {
		return &domain.Brief{
			GateResult: domain.BriefMissing,
		}, nil
	}
	return r.Brief, nil
}
