package service

import (
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestAssembleHealth verifies ProjectService.AssembleHealth orchestration.
func TestAssembleHealth(t *testing.T) {
	t.Run("minimal project", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		iterRepo := mem.NewIterationRepo()
		stateRepo := mem.NewStateRepo()
		briefRepo := mem.NewBriefRepo()

		svc := NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		project := &domain.Project{Root: "/test", Name: "test-project"}

		health, err := svc.AssembleHealth(project)
		if err != nil {
			t.Fatalf("AssembleHealth() error = %v", err)
		}

		if health.Project.Name != "test-project" {
			t.Errorf("Project.Name = %q, want test-project", health.Project.Name)
		}
		// Brief is missing by default
		if health.Brief.GateResult != domain.BriefMissing {
			t.Errorf("Brief.GateResult = %q, want BRIEF_MISSING", health.Brief.GateResult)
		}
	})

	t.Run("healthy project with all data", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		iterRepo := mem.NewIterationRepo()
		stateRepo := mem.NewStateRepo()
		briefRepo := mem.NewBriefRepo()

		// Add some docs
		docRepo.Docs["docs/spec/requirements.md"] = domain.Document{
			Path: "docs/spec/requirements.md", Zone: domain.ZoneSpec, IsStub: false,
		}
		docRepo.Docs["docs/spec/architecture.md"] = domain.Document{
			Path: "docs/spec/architecture.md", Zone: domain.ZoneSpec, IsStub: true,
		}

		// Add brief
		briefRepo.Brief = &domain.Brief{
			Exists:     true,
			GateResult: domain.BriefPresent,
		}

		// Add iteration
		iterRepo.Iterations = []domain.Iteration{
			{Seq: 1, Type: domain.TypeNewProject, Descriptor: "core-cli"},
		}

		// Add workflow
		stateRepo.State = &domain.WorkflowState{
			Type:       domain.TypeNewProject,
			Descriptor: "core-cli",
			LastAgent:  "developer",
		}

		svc := NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		project := &domain.Project{
			Root: "/test",
			Name: "test-project",
			Config: &domain.Config{
				Project: domain.ProjectMeta{Name: "test-project"},
			},
		}

		health, err := svc.AssembleHealth(project)
		if err != nil {
			t.Fatalf("AssembleHealth() error = %v", err)
		}

		// Brief should be present
		if health.Brief.GateResult != domain.BriefPresent {
			t.Errorf("Brief.GateResult = %q, want BRIEF_PRESENT", health.Brief.GateResult)
		}

		// Zone health should reflect docs
		specZone, ok := health.Zones[domain.ZoneSpec]
		if !ok {
			t.Fatal("Missing spec zone in health")
		}
		if specZone.Total != 2 {
			t.Errorf("spec zone Total = %d, want 2", specZone.Total)
		}
		if specZone.Stubs != 1 {
			t.Errorf("spec zone Stubs = %d, want 1", specZone.Stubs)
		}
		if specZone.Complete != 1 {
			t.Errorf("spec zone Complete = %d, want 1", specZone.Complete)
		}

		// Last iteration
		if health.LastIteration == nil {
			t.Fatal("LastIteration should not be nil")
		}
		if health.LastIteration.Seq != 1 {
			t.Errorf("LastIteration.Seq = %d, want 1", health.LastIteration.Seq)
		}

		// Workflow
		if health.Workflow == nil {
			t.Fatal("Workflow should not be nil")
		}
		if health.Workflow.LastAgent != "developer" {
			t.Errorf("Workflow.LastAgent = %q, want developer", health.Workflow.LastAgent)
		}
	})

	t.Run("warnings for missing brief", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		iterRepo := mem.NewIterationRepo()
		stateRepo := mem.NewStateRepo()
		briefRepo := mem.NewBriefRepo() // Default: BRIEF_MISSING

		svc := NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		project := &domain.Project{Root: "/test", Name: "test"}

		health, _ := svc.AssembleHealth(project)

		// Should have warning about missing brief
		found := false
		for _, w := range health.Warnings {
			if w != "" {
				found = true
				break
			}
		}
		if !found && len(health.Warnings) == 0 {
			t.Error("Should have warnings for missing brief")
		}
	})

	t.Run("warnings for missing config", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		iterRepo := mem.NewIterationRepo()
		stateRepo := mem.NewStateRepo()
		briefRepo := mem.NewBriefRepo()
		briefRepo.Brief = &domain.Brief{Exists: true, GateResult: domain.BriefPresent}

		svc := NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		// Project with no Config
		project := &domain.Project{Root: "/test", Name: "test", Config: nil}

		health, _ := svc.AssembleHealth(project)

		hasConfigWarning := false
		for _, w := range health.Warnings {
			if w != "" {
				hasConfigWarning = true
			}
		}
		if !hasConfigWarning {
			t.Error("Should have warning for missing mind.toml")
		}
	})

	t.Run("suggestions for missing brief", func(t *testing.T) {
		docRepo := mem.NewDocRepo()
		iterRepo := mem.NewIterationRepo()
		stateRepo := mem.NewStateRepo()
		briefRepo := mem.NewBriefRepo()

		svc := NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		project := &domain.Project{Root: "/test", Name: "test"}

		health, _ := svc.AssembleHealth(project)
		if len(health.Suggestions) == 0 {
			t.Error("Should have suggestions for missing brief")
		}
	})
}
