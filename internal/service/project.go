package service

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
)

// ProjectService orchestrates project detection and health assembly.
type ProjectService struct {
	docRepo   repo.DocRepo
	iterRepo  repo.IterationRepo
	stateRepo repo.StateRepo
	briefRepo repo.BriefRepo
}

// NewProjectService creates a ProjectService with injected dependencies.
func NewProjectService(
	docRepo repo.DocRepo,
	iterRepo repo.IterationRepo,
	stateRepo repo.StateRepo,
	briefRepo repo.BriefRepo,
) *ProjectService {
	return &ProjectService{
		docRepo:   docRepo,
		iterRepo:  iterRepo,
		stateRepo: stateRepo,
		briefRepo: briefRepo,
	}
}

// DetectProject builds a Project from the filesystem at the given root.
func (s *ProjectService) DetectProject(root string) (*domain.Project, error) {
	return fs.DetectProject(root)
}

// AssembleHealth builds the ProjectHealth aggregate from all data sources.
func (s *ProjectService) AssembleHealth(project *domain.Project) (*domain.ProjectHealth, error) {
	health := &domain.ProjectHealth{
		Project: *project,
		Zones:   make(map[domain.Zone]domain.ZoneHealth),
	}

	// Brief status
	if brief, err := s.briefRepo.ParseBrief(); err == nil {
		health.Brief = *brief
	}

	// Zone health
	for _, zone := range domain.AllZones {
		docs, err := s.docRepo.ListByZone(zone)
		if err != nil {
			continue
		}
		zh := domain.ZoneHealth{Zone: zone, Total: len(docs), Files: docs}
		for _, doc := range docs {
			zh.Present++
			if doc.IsStub {
				zh.Stubs++
			} else {
				zh.Complete++
			}
		}
		health.Zones[zone] = zh
	}

	// Workflow state
	if s.stateRepo != nil {
		if ws, err := s.stateRepo.ReadWorkflow(); err == nil && ws != nil {
			health.Workflow = ws
		}
	}

	// Last iteration
	iterations, err := s.iterRepo.List()
	if err == nil && len(iterations) > 0 {
		health.LastIteration = &iterations[0]
	}

	// Warnings
	if !health.Brief.Exists {
		health.Warnings = append(health.Warnings, "Project brief missing — run 'mind create brief' or create docs/spec/project-brief.md")
	} else if health.Brief.IsStub {
		health.Warnings = append(health.Warnings, "Project brief is a stub — fill in Vision, Key Deliverables, and Scope")
	}

	if project.Config == nil {
		health.Warnings = append(health.Warnings, "mind.toml not found — run 'mind init' or create mind.toml")
	}

	for _, zone := range domain.AllZones {
		zh, ok := health.Zones[zone]
		if !ok {
			continue
		}
		if zh.Stubs > 0 {
			health.Warnings = append(health.Warnings, fmt.Sprintf("%s/ has %d stub file(s)", zone, zh.Stubs))
		}
	}

	// Suggestions
	if !health.Brief.Exists {
		health.Suggestions = append(health.Suggestions, "Run: mind create brief")
	}
	if project.Config == nil {
		health.Suggestions = append(health.Suggestions, "Run: mind init")
	}

	return health, nil
}
