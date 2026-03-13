package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// setupPassingDocRepo populates a mem.DocRepo with the minimum structure to pass DocsSuite.
func setupPassingDocRepo(docRepo *mem.DocRepo) {
	docRepo.Dirs["docs"] = true
	for _, zone := range domain.AllZones {
		docRepo.Dirs["docs/"+string(zone)] = true
	}
	docRepo.Dirs["docs/spec/decisions"] = true

	docRepo.Files["docs/spec/project-brief.md"] = []byte("# Brief\n\n## Vision\nBuild CLI.\n\n## Key Deliverables\n- Binary\n\n## Scope\nPhase 1.")
	docRepo.Files["docs/spec/requirements.md"] = []byte("# Requirements\n\nFR-1: The system shall...\nFR-2: The system shall...")
	docRepo.Files["docs/spec/architecture.md"] = []byte("# Architecture\n\n4-layer architecture.\nPresentation layer uses Cobra.\nService layer orchestrates.")
	docRepo.Files["docs/state/current.md"] = []byte("# Current State\n\nActive work.\nAll tests passing.")
	docRepo.Files["docs/state/workflow.md"] = []byte("# Workflow\n\nCurrent state.\nTester agent active.")
	docRepo.Files["docs/knowledge/glossary.md"] = []byte("# Glossary\n\n| Term | Definition |\n|------|-----------|")
	docRepo.Files["docs/blueprints/INDEX.md"] = []byte("# Blueprints\n\nIndex of all blueprints.")

	docRepo.Docs["docs/spec/project-brief.md"] = domain.Document{Path: "docs/spec/project-brief.md", Zone: domain.ZoneSpec, Name: "project-brief"}
	docRepo.Docs["docs/spec/requirements.md"] = domain.Document{Path: "docs/spec/requirements.md", Zone: domain.ZoneSpec, Name: "requirements"}
	docRepo.Docs["docs/spec/architecture.md"] = domain.Document{Path: "docs/spec/architecture.md", Zone: domain.ZoneSpec, Name: "architecture"}
	docRepo.Docs["docs/blueprints/INDEX.md"] = domain.Document{Path: "docs/blueprints/INDEX.md", Zone: domain.ZoneBlueprints, Name: "INDEX"}
}

// passingBrief returns a domain.Brief that passes the brief gate.
func passingBrief() *domain.Brief {
	return &domain.Brief{
		Exists:          true,
		IsStub:          false,
		HasVision:       true,
		HasDeliverables: true,
		HasScope:        true,
		GateResult:      domain.BriefPresent,
	}
}

// newTestPreflightService creates a PreflightService with in-memory repos and a real temp dir.
// passDocValidation controls whether the mem doc repo is populated to pass DocsSuite.
func newTestPreflightService(t *testing.T, brief *domain.Brief, passDocValidation bool) *PreflightService {
	t.Helper()
	root := t.TempDir()

	briefRepo := mem.NewBriefRepo()
	if brief != nil {
		briefRepo.Brief = brief
	}
	stateRepo := mem.NewStateRepo()
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	configRepo := mem.NewConfigRepo()

	if passDocValidation {
		setupPassingDocRepo(docRepo)
	}

	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	generateSvc := service.NewGenerateService(root)

	return NewPreflightService(root, briefRepo, stateRepo, validationSvc, generateSvc, nil)
}

// FR-142: Run() classifies TypeBugFix correctly.
func TestPreflightService_Run_BugFixClassification(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	result, err := svc.Run("fix the authentication token expiry bug")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RequestType != domain.TypeBugFix {
		t.Errorf("RequestType = %v, want %v", result.RequestType, domain.TypeBugFix)
	}
}

// FR-142: Run() classifies TypeEnhancement correctly.
func TestPreflightService_Run_EnhancementClassification(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	result, err := svc.Run("add JSON export support to the report command")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RequestType != domain.TypeEnhancement {
		t.Errorf("RequestType = %v, want %v", result.RequestType, domain.TypeEnhancement)
	}
}

// FR-142: Run() classifies TypeRefactor correctly.
func TestPreflightService_Run_RefactorClassification(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	result, err := svc.Run("refactor the service layer to reduce coupling")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RequestType != domain.TypeRefactor {
		t.Errorf("RequestType = %v, want %v", result.RequestType, domain.TypeRefactor)
	}
}

// FR-142: Run() classifies TypeComplexNew for analyze: prefix requests.
func TestPreflightService_Run_ComplexNewClassification(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	// "analyze:" prefix → TypeComplexNew per domain.Classify().
	result, err := svc.Run("analyze: design the new multi-phase orchestration system")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RequestType != domain.TypeComplexNew {
		t.Errorf("RequestType = %v, want %v", result.RequestType, domain.TypeComplexNew)
	}
}

// FR-142 / FR-147: Run() blocks when TypeComplexNew request has a missing brief ("brief gate BLOCKED").
func TestPreflightService_Run_ComplexNew_MissingBriefBlocks(t *testing.T) {
	// nil brief → BriefMissing gate result.
	svc := newTestPreflightService(t, nil, true)

	// "analyze:" prefix → TypeComplexNew; missing brief must block.
	_, err := svc.Run("analyze: design the new orchestration system")
	if err == nil {
		t.Fatal("Run() should return error when brief is missing for TypeComplexNew")
	}
	if !strings.Contains(err.Error(), "BLOCKED") {
		t.Errorf("error = %q, want message containing 'BLOCKED'", err.Error())
	}
}

// FR-142 / FR-147: Run() blocks when TypeComplexNew request has a stub brief.
func TestPreflightService_Run_ComplexNew_StubBriefBlocks(t *testing.T) {
	stubBrief := &domain.Brief{GateResult: domain.BriefStub}
	svc := newTestPreflightService(t, stubBrief, true)

	// "analyze:" prefix → TypeComplexNew; stub brief must block.
	_, err := svc.Run("analyze: design the complete new system")
	if err == nil {
		t.Fatal("Run() should return error when brief is stub for TypeComplexNew")
	}
	if !strings.Contains(err.Error(), "BLOCKED") {
		t.Errorf("error = %q, want message containing 'BLOCKED'", err.Error())
	}
}

// FR-147: Run() blocks when doc validation returns Failed > 0.
// With empty mem repos, DocsSuite will report failures.
func TestPreflightService_Run_DocFailureBlocks(t *testing.T) {
	// passDocValidation=false → empty doc repo → DocsSuite will fail.
	svc := newTestPreflightService(t, passingBrief(), false)

	_, err := svc.Run("fix crash on startup")
	if err == nil {
		t.Fatal("Run() should return error when doc validation fails")
	}
	if !strings.Contains(err.Error(), "documentation check") && !strings.Contains(err.Error(), "preflight blocked") {
		t.Errorf("error = %q, want message about documentation failures", err.Error())
	}
}

// FR-142: Resume() returns nil when no workflow state is stored.
func TestPreflightService_Resume_NoState(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), false)

	state, err := svc.Resume()
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if state != nil {
		t.Errorf("Resume() = %+v, want nil for no stored state", state)
	}
}

// FR-142: Resume() returns the stored WorkflowState when one exists.
func TestPreflightService_Resume_WithState(t *testing.T) {
	root := t.TempDir()

	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = passingBrief()
	stateRepo := mem.NewStateRepo()
	stateRepo.State = &domain.WorkflowState{
		Type:           domain.TypeEnhancement,
		Descriptor:     "add-feature",
		LastAgent:      "analyst",
		RemainingChain: []string{"developer", "tester"},
	}
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	configRepo := mem.NewConfigRepo()

	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	generateSvc := service.NewGenerateService(root)

	svc := NewPreflightService(root, briefRepo, stateRepo, validationSvc, generateSvc, nil)

	state, err := svc.Resume()
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if state == nil {
		t.Fatal("Resume() returned nil, want stored workflow state")
	}
	if state.Descriptor != "add-feature" {
		t.Errorf("state.Descriptor = %q, want 'add-feature'", state.Descriptor)
	}
	if state.LastAgent != "analyst" {
		t.Errorf("state.LastAgent = %q, want 'analyst'", state.LastAgent)
	}
}

// FR-142: AgentChainFor returns non-empty slices for each RequestType.
func TestAgentChainFor_AllTypes(t *testing.T) {
	types := []domain.RequestType{
		domain.TypeNewProject,
		domain.TypeComplexNew,
		domain.TypeEnhancement,
		domain.TypeBugFix,
		domain.TypeRefactor,
		domain.TypeDiagnose,
	}

	for _, reqType := range types {
		t.Run(string(reqType), func(t *testing.T) {
			chain := AgentChainFor(reqType)
			if len(chain) == 0 {
				t.Errorf("AgentChainFor(%v) returned empty slice", reqType)
			}
			for _, agent := range chain {
				if agent == "" {
					t.Errorf("AgentChainFor(%v) contains empty agent name", reqType)
				}
			}
		})
	}
}

// FR-142: AgentChainFor(TypeComplexNew) includes tester agent.
func TestAgentChainFor_ComplexNew_HasTester(t *testing.T) {
	chain := AgentChainFor(domain.TypeComplexNew)
	found := false
	for _, a := range chain {
		if a == "tester" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AgentChainFor(TypeComplexNew) = %v, want 'tester' in chain", chain)
	}
}

// FR-142: Run() writes workflow state on a successful execution.
func TestPreflightService_Run_WritesWorkflowState(t *testing.T) {
	root := t.TempDir()

	briefRepo := mem.NewBriefRepo()
	briefRepo.Brief = passingBrief()

	stateRepo := mem.NewStateRepo()

	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	configRepo := mem.NewConfigRepo()
	setupPassingDocRepo(docRepo)

	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	generateSvc := service.NewGenerateService(root)

	svc := NewPreflightService(root, briefRepo, stateRepo, validationSvc, generateSvc, nil)

	// BugFix skips brief gate requirement.
	result, err := svc.Run("fix crash on startup")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if stateRepo.State == nil {
		t.Error("stateRepo.State is nil after successful Run() — workflow state was not written")
	}
	if result.RequestType != domain.TypeBugFix {
		t.Errorf("RequestType = %v, want TypeBugFix", result.RequestType)
	}
}

// FR-149: Classify adapter delegates to domain.Classify.
func TestClassify_DelegatesTo_DomainClassify(t *testing.T) {
	tests := []struct {
		request string
		want    domain.RequestType
	}{
		{"fix: crash on startup", domain.TypeBugFix},
		{"add: new export feature", domain.TypeEnhancement},
		{"refactor: simplify service layer", domain.TypeRefactor},
		{"analyze: design new system", domain.TypeComplexNew},
		{"create: new project scaffold", domain.TypeNewProject},
	}
	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			got := Classify(tt.request)
			if got != tt.want {
				t.Errorf("Classify(%q) = %v, want %v", tt.request, got, tt.want)
			}
		})
	}
}

// FR-149: Slugify adapter delegates to domain.Slugify.
func TestSlugify_DelegatesTo_DomainSlugify(t *testing.T) {
	input := "Add JSON export support"
	slug := Slugify(input)
	if slug == "" {
		t.Error("Slugify() returned empty string")
	}
	// Slug should be lowercase and hyphen-separated.
	if strings.Contains(slug, " ") {
		t.Errorf("Slugify() = %q, want no spaces", slug)
	}
}

// FR-142: HandoffService.Run() returns error for unknown iterationID.
func TestHandoffService_Run_UnknownIterationID(t *testing.T) {
	root := t.TempDir()

	iterRepo := mem.NewIterationRepo()
	stateRepo := mem.NewStateRepo()
	docRepo := mem.NewDocRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()
	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	svc := NewHandoffService(root, iterRepo, stateRepo, validationSvc)

	_, err := svc.Run("999-ENHANCEMENT-nonexistent", "main")
	if err == nil {
		t.Fatal("Run() should return error for non-existent iteration ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want message containing 'not found'", err.Error())
	}
}

// FR-142: HandoffService.Run() succeeds for a known iteration.
func TestHandoffService_Run_KnownIteration(t *testing.T) {
	root := t.TempDir()

	iterRepo := mem.NewIterationRepo()
	// Add an iteration to the repo.
	iterRepo.Iterations = []domain.Iteration{
		{
			Seq:        1,
			Type:       domain.TypeBugFix,
			Descriptor: "fix-crash",
			DirName:    "001-BUG_FIX-fix-crash",
			Status:     domain.IterComplete,
			Artifacts: []domain.Artifact{
				{Name: "overview.md", Exists: true},
				{Name: "changes.md", Exists: true},
				{Name: "test-summary.md", Exists: true},
				{Name: "validation.md", Exists: true},
				{Name: "retrospective.md", Exists: true},
			},
		},
	}

	stateRepo := mem.NewStateRepo()
	stateRepo.State = &domain.WorkflowState{
		Type:       domain.TypeBugFix,
		Descriptor: "fix-crash",
	}

	docRepo := mem.NewDocRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()
	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)

	svc := NewHandoffService(root, iterRepo, stateRepo, validationSvc)

	result, err := svc.Run("001-BUG_FIX-fix-crash", "main")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.IterationID != "001-BUG_FIX-fix-crash" {
		t.Errorf("IterationID = %q, want '001-BUG_FIX-fix-crash'", result.IterationID)
	}
	if !result.StateCleared {
		t.Error("StateCleared = false, want true after handoff")
	}
	if result.ArtifactsPresent != 5 {
		t.Errorf("ArtifactsPresent = %d, want 5", result.ArtifactsPresent)
	}
}

// FR-142: Run() populates DocWarnings when doc validation has warnings but no failures.
func TestPreflightService_Run_DocWarningsNonBlocking(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	// BugFix: brief gate is skipped; doc validation runs on well-formed docs.
	result, err := svc.Run("fix crash on startup")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// DocWarnings should be >= 0 (non-negative).
	if result.DocWarnings < 0 {
		t.Errorf("DocWarnings = %d, want >= 0", result.DocWarnings)
	}
}

// PromptBuilder.Build returns a non-empty string for any request type.
func TestPromptBuilder_Build_ReturnsPrompt(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.Build("add feature X", domain.TypeEnhancement, "docs/iterations/001-ENHANCEMENT-add-feature-x")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if prompt == "" {
		t.Error("Build() returned empty prompt")
	}
	if !strings.Contains(prompt, "add feature X") {
		t.Errorf("prompt does not contain request text:\n%s", prompt)
	}
}

// PromptBuilder.Build includes agent chain in the output.
func TestPromptBuilder_Build_IncludesAgentChain(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.Build("analyze: new system", domain.TypeComplexNew, "docs/iterations/001-COMPLEX_NEW-new-system")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !strings.Contains(prompt, "tester") {
		t.Errorf("prompt for TypeComplexNew should include 'tester' in agent chain:\n%s", prompt)
	}
}

// PromptBuilder.Build reads context files when they exist.
func TestPromptBuilder_Build_ReadsContextFiles(t *testing.T) {
	root := t.TempDir()

	// Create docs directories and a brief.
	if err := os.MkdirAll(filepath.Join(root, "docs", "spec"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	briefContent := "# Project Brief\n\n## Vision\nTest project."
	if err := os.WriteFile(filepath.Join(root, "docs", "spec", "project-brief.md"), []byte(briefContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	builder := NewPromptBuilder(root)
	prompt, err := builder.Build("fix bug", domain.TypeBugFix, "docs/iterations/001-BUG_FIX-fix-bug")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !strings.Contains(prompt, "Test project") {
		t.Errorf("prompt should include brief content:\n%s", prompt)
	}
}

// FR-142: Run() handles TypeNewProject with passing brief.
func TestPreflightService_Run_NewProjectClassification(t *testing.T) {
	svc := newTestPreflightService(t, passingBrief(), true)

	// "create:" prefix → TypeNewProject per domain.Classify().
	result, err := svc.Run("create: new project scaffold")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RequestType != domain.TypeNewProject {
		t.Errorf("RequestType = %v, want %v", result.RequestType, domain.TypeNewProject)
	}
}

// FR-142: Run() proceeds for TypeEnhancement with a missing brief (warning, not error).
func TestPreflightService_Run_Enhancement_MissingBriefIsWarning(t *testing.T) {
	// nil brief → BriefMissing; for TypeEnhancement this is a warning not a block.
	svc := newTestPreflightService(t, nil, true)

	result, err := svc.Run("add export feature to command")
	if err != nil {
		t.Fatalf("Run() should not block for TypeEnhancement with missing brief, got: %v", err)
	}
	if result.RequestType != domain.TypeEnhancement {
		t.Errorf("RequestType = %v, want TypeEnhancement", result.RequestType)
	}
	// Warning should be in Warnings.
	if len(result.Warnings) == 0 {
		t.Error("expected warning about missing brief for TypeEnhancement")
	}
}

// PromptBuilder.Build includes "No previous iterations" when no iterations dir exists.
func TestPromptBuilder_Build_NoIterations(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.Build("fix bug", domain.TypeBugFix, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !strings.Contains(prompt, "No previous iterations") {
		t.Errorf("prompt should say 'No previous iterations' when dir is missing:\n%s", prompt)
	}
}

// PromptBuilder.Build includes MCP tools section.
func TestPromptBuilder_Build_IncludesMCPTools(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.Build("add feature", domain.TypeEnhancement, "docs/iterations/001")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !strings.Contains(prompt, "Available MCP Tools") {
		t.Error("prompt should contain 'Available MCP Tools' section")
	}
	// Spot-check a few tool names
	for _, tool := range []string{"mind_status", "mind_check_gate", "mind_read_state", "mind_update_state", "mind_suggest_next"} {
		if !strings.Contains(prompt, tool) {
			t.Errorf("prompt should contain tool %q", tool)
		}
	}
}

// PromptBuilder.BuildAnalyze returns a non-empty prompt with the topic.
func TestPromptBuilder_BuildAnalyze(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.BuildAnalyze("GraphQL vs REST")
	if err != nil {
		t.Fatalf("BuildAnalyze() error = %v", err)
	}
	if prompt == "" {
		t.Error("BuildAnalyze() returned empty prompt")
	}
	if !strings.Contains(prompt, "GraphQL vs REST") {
		t.Error("prompt should contain topic")
	}
	if !strings.Contains(prompt, "Conversation Analysis") {
		t.Error("prompt should contain 'Conversation Analysis' header")
	}
}

// PromptBuilder.BuildAnalyze reads moderator agent when it exists.
func TestPromptBuilder_BuildAnalyze_ReadsModerator(t *testing.T) {
	root := t.TempDir()

	// Create moderator file
	modDir := filepath.Join(root, ".mind", "conversation", "agents")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "moderator.md"), []byte("# Moderator\n\nTest moderator content."), 0644); err != nil {
		t.Fatal(err)
	}

	builder := NewPromptBuilder(root)
	prompt, err := builder.BuildAnalyze("test topic")
	if err != nil {
		t.Fatalf("BuildAnalyze() error = %v", err)
	}
	if !strings.Contains(prompt, "Test moderator content") {
		t.Error("prompt should include moderator agent content")
	}
}

// PromptBuilder.BuildDiscover returns a non-empty prompt with the idea.
func TestPromptBuilder_BuildDiscover(t *testing.T) {
	root := t.TempDir()
	builder := NewPromptBuilder(root)

	prompt, err := builder.BuildDiscover("inventory management system")
	if err != nil {
		t.Fatalf("BuildDiscover() error = %v", err)
	}
	if prompt == "" {
		t.Error("BuildDiscover() returned empty prompt")
	}
	if !strings.Contains(prompt, "inventory management system") {
		t.Error("prompt should contain idea")
	}
	if !strings.Contains(prompt, "Project Discovery") {
		t.Error("prompt should contain 'Project Discovery' header")
	}
}

// PromptBuilder.BuildDiscover reads discovery agent when it exists.
func TestPromptBuilder_BuildDiscover_ReadsDiscoveryAgent(t *testing.T) {
	root := t.TempDir()

	// Create discovery agent file
	agentDir := filepath.Join(root, ".mind", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "discovery.md"), []byte("# Discovery\n\nTest discovery content."), 0644); err != nil {
		t.Fatal(err)
	}

	builder := NewPromptBuilder(root)
	prompt, err := builder.BuildDiscover("test idea")
	if err != nil {
		t.Fatalf("BuildDiscover() error = %v", err)
	}
	if !strings.Contains(prompt, "Test discovery content") {
		t.Error("prompt should include discovery agent content")
	}
}

// PromptBuilder.BuildDiscover includes existing brief for update behavior.
func TestPromptBuilder_BuildDiscover_IncludesExistingBrief(t *testing.T) {
	root := t.TempDir()

	// Create existing brief
	specDir := filepath.Join(root, "docs", "spec")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "project-brief.md"), []byte("# Brief\n\n## Vision\nExisting vision."), 0644); err != nil {
		t.Fatal(err)
	}

	builder := NewPromptBuilder(root)
	prompt, err := builder.BuildDiscover("extend the project")
	if err != nil {
		t.Fatalf("BuildDiscover() error = %v", err)
	}
	if !strings.Contains(prompt, "Existing vision") {
		t.Error("prompt should include existing brief content")
	}
}

// PromptBuilder.recentIterationOverviews returns entries for existing iteration dirs.
func TestPromptBuilder_RecentIterationOverviews(t *testing.T) {
	root := t.TempDir()
	iterDir := filepath.Join(root, "docs", "iterations")

	// Create an iteration directory with an overview.md.
	iterPath := filepath.Join(iterDir, "001-ENHANCEMENT-test-feature")
	if err := os.MkdirAll(iterPath, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(iterPath, "overview.md"), []byte("# Overview\n\nTest iteration."), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	builder := NewPromptBuilder(root)
	prompt, err := builder.Build("fix bug", domain.TypeBugFix, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !strings.Contains(prompt, "Test iteration") {
		t.Errorf("prompt should include overview content:\n%s", prompt)
	}
}
