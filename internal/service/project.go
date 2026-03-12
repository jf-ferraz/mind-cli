package service

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
)

// ProjectService orchestrates project detection and health assembly.
type ProjectService struct {
	docRepo    repo.DocRepo
	iterRepo   repo.IterationRepo
	stateRepo  repo.StateRepo
	briefRepo  repo.BriefRepo
	configRepo repo.ConfigRepo
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

// NewProjectServiceWithConfig creates a ProjectService including a configRepo.
func NewProjectServiceWithConfig(
	docRepo repo.DocRepo,
	iterRepo repo.IterationRepo,
	stateRepo repo.StateRepo,
	briefRepo repo.BriefRepo,
	configRepo repo.ConfigRepo,
) *ProjectService {
	return &ProjectService{
		docRepo:    docRepo,
		iterRepo:   iterRepo,
		stateRepo:  stateRepo,
		briefRepo:  briefRepo,
		configRepo: configRepo,
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

// ListStubs returns all stub documents across all zones.
func (s *ProjectService) ListStubs() (*domain.StubList, error) {
	docs, err := s.docRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("list docs: %w", err)
	}
	list := &domain.StubList{}
	for _, doc := range docs {
		if doc.IsStub {
			list.Stubs = append(list.Stubs, domain.StubEntry{
				Path: doc.Path,
				Zone: string(doc.Zone),
				Hint: fmt.Sprintf("Fill in content for %s", doc.Path),
			})
		}
	}
	list.Count = len(list.Stubs)
	return list, nil
}

// SearchDocs performs full-text search across all documentation.
func (s *ProjectService) SearchDocs(query string) (*domain.SearchResults, error) {
	return s.docRepo.Search(query)
}

// Config returns the parsed project configuration.
func (s *ProjectService) Config() (*domain.Config, error) {
	if s.configRepo == nil {
		return nil, fmt.Errorf("config repo not available")
	}
	return s.configRepo.ReadProjectConfig()
}

// CheckBrief returns the brief gate result.
func (s *ProjectService) CheckBrief() (*domain.Brief, error) {
	return s.briefRepo.ParseBrief()
}

// SuggestNext analyzes project state and returns the next recommended action.
func (s *ProjectService) SuggestNext(root string) ([]domain.Suggestion, error) {
	var suggestions []domain.Suggestion

	// Check brief
	brief, err := s.briefRepo.ParseBrief()
	if err == nil {
		switch brief.GateResult {
		case domain.BriefMissing:
			suggestions = append(suggestions, domain.Suggestion{
				Action:  "Create project brief",
				Reason:  "Brief is missing — required before starting AI workflows",
				Command: "mind create brief",
			})
			return suggestions, nil
		case domain.BriefStub:
			suggestions = append(suggestions, domain.Suggestion{
				Action:  "Fill in project brief",
				Reason:  "Brief is a stub — Vision, Deliverables, and Scope sections need content",
				Command: "mind create brief",
			})
		}
	}

	// Check workflow state
	if s.stateRepo != nil {
		if ws, err := s.stateRepo.ReadWorkflow(); err == nil && ws != nil && !ws.IsIdle() {
			suggestions = append(suggestions, domain.Suggestion{
				Action:  fmt.Sprintf("Resume workflow: %s", ws.Descriptor),
				Reason:  fmt.Sprintf("Workflow in progress — last agent: %s", ws.LastAgent),
				Command: "mind preflight --resume",
			})
			return suggestions, nil
		}
	}

	// Check for stale iterations
	iterations, err := s.iterRepo.List()
	if err == nil && len(iterations) > 0 {
		last := iterations[0]
		if last.Status == domain.IterInProgress {
			suggestions = append(suggestions, domain.Suggestion{
				Action:  fmt.Sprintf("Complete iteration %s", last.DirName),
				Reason:  "Last iteration is in progress",
				Command: fmt.Sprintf("mind handoff %s", last.DirName),
			})
			return suggestions, nil
		}
	}

	// Default: start a new workflow
	suggestions = append(suggestions, domain.Suggestion{
		Action:  "Start a new workflow",
		Reason:  "Project is in a clean state",
		Command: `mind preflight "<describe your request>"`,
	})

	return suggestions, nil
}
