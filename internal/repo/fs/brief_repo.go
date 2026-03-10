package fs

import (
	"regexp"

	"github.com/jf-ferraz/mind-cli/domain"
)

// BriefRepo implements repo.BriefRepo using the filesystem.
type BriefRepo struct {
	docRepo *DocRepo
}

// NewBriefRepo creates a BriefRepo.
func NewBriefRepo(docRepo *DocRepo) *BriefRepo {
	return &BriefRepo{docRepo: docRepo}
}

var (
	visionRe       = regexp.MustCompile(`(?im)^##\s+.*vision`)
	deliverablesRe = regexp.MustCompile(`(?im)^##\s+.*key\s+deliverables`)
	scopeRe        = regexp.MustCompile(`(?im)^##\s+.*scope`)
)

// ParseBrief reads and analyzes the project brief.
func (r *BriefRepo) ParseBrief() (*domain.Brief, error) {
	briefPath := "docs/spec/project-brief.md"

	brief := &domain.Brief{
		Path: briefPath,
	}

	if !r.docRepo.Exists(briefPath) {
		brief.Exists = false
		brief.GateResult = domain.BriefMissing
		return brief, nil
	}

	brief.Exists = true

	isStub, err := r.docRepo.IsStub(briefPath)
	if err != nil {
		return nil, err
	}
	brief.IsStub = isStub

	if isStub {
		brief.GateResult = domain.BriefStub
		return brief, nil
	}

	content, err := r.docRepo.Read(briefPath)
	if err != nil {
		return nil, err
	}

	brief.HasVision = visionRe.Match(content)
	brief.HasDeliverables = deliverablesRe.Match(content)
	brief.HasScope = scopeRe.Match(content)

	if brief.HasVision && brief.HasDeliverables && brief.HasScope {
		brief.GateResult = domain.BriefPresent
	} else {
		// Has content but missing required sections — still a stub for gate purposes
		brief.GateResult = domain.BriefStub
	}

	return brief, nil
}
