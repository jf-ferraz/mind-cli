// Package orchestrate implements the Model A pre-flight and handoff flows.
package orchestrate

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// PreflightResult holds the output from a successful pre-flight run.
type PreflightResult struct {
	RequestType   domain.RequestType
	Descriptor    string
	Branch        string
	IterationPath string
	BriefGate     domain.BriefGate
	DocsReport    domain.ValidationReport
	Prompt        string
	Warnings      []string
}

// PreflightService runs the 7-step pre-flight sequence.
type PreflightService struct {
	projectRoot   string
	briefRepo     repo.BriefRepo
	stateRepo     repo.StateRepo
	validationSvc *service.ValidationService
	generateSvc   *service.GenerateService
	promptBuilder *PromptBuilder
}

// NewPreflightService creates a PreflightService.
func NewPreflightService(
	projectRoot string,
	briefRepo repo.BriefRepo,
	stateRepo repo.StateRepo,
	validationSvc *service.ValidationService,
	generateSvc *service.GenerateService,
	promptBuilder *PromptBuilder,
) *PreflightService {
	return &PreflightService{
		projectRoot:   projectRoot,
		briefRepo:     briefRepo,
		stateRepo:     stateRepo,
		validationSvc: validationSvc,
		generateSvc:   generateSvc,
		promptBuilder: promptBuilder,
	}
}

// Run executes all 7 pre-flight steps for the given request.
// Returns a PreflightResult on success, or an error on blockers.
func (s *PreflightService) Run(request string) (*PreflightResult, error) {
	result := &PreflightResult{}

	// Step 1: Classify request
	reqType := domain.Classify(request)
	result.RequestType = reqType
	result.Descriptor = domain.Slugify(request)
	if len(result.Descriptor) > 40 {
		result.Descriptor = result.Descriptor[:40]
	}

	// Step 2: Business context gate
	briefGate, warn, err := s.runBriefGate(reqType)
	if err != nil {
		return nil, fmt.Errorf("business context gate: %w", err)
	}
	result.BriefGate = briefGate
	if warn != "" {
		result.Warnings = append(result.Warnings, warn)
	}

	// Step 3: Validate documentation (non-blocking)
	docsReport := s.validationSvc.RunDocs(s.projectRoot, false)
	result.DocsReport = docsReport

	// Step 4: Create iteration folder
	typeStr := typeToString(reqType)
	iterResult, err := s.generateSvc.CreateIteration(typeStr, result.Descriptor)
	if err != nil {
		return nil, fmt.Errorf("create iteration: %w", err)
	}
	result.IterationPath = iterResult.Path

	// Step 5: Create git branch
	branchName := buildBranchName(reqType, result.Descriptor)
	result.Branch = branchName
	if err := createGitBranch(s.projectRoot, branchName); err != nil {
		// Non-fatal: log as warning
		result.Warnings = append(result.Warnings, fmt.Sprintf("git branch creation failed: %v", err))
	}

	// Step 6: Assemble context package + write workflow state
	state := &domain.WorkflowState{
		Type:           reqType,
		Descriptor:     result.Descriptor,
		IterationPath:  result.IterationPath,
		Branch:         result.Branch,
		LastAgent:      "",
		RemainingChain: agentChain(reqType),
		Session:        1,
		TotalSessions:  1,
	}
	if err := s.stateRepo.WriteWorkflow(state); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("write workflow state: %v", err))
	}

	// Step 7: Generate prompt
	if s.promptBuilder != nil {
		prompt, err := s.promptBuilder.Build(request, reqType, result.IterationPath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("generate prompt: %v", err))
		} else {
			result.Prompt = prompt
		}
	}

	return result, nil
}

// Resume reads docs/state/workflow.md and returns current in-progress state.
// Returns nil, nil if no workflow is in progress.
func (s *PreflightService) Resume() (*domain.WorkflowState, error) {
	return s.stateRepo.ReadWorkflow()
}


func (s *PreflightService) runBriefGate(reqType domain.RequestType) (domain.BriefGate, string, error) {
	brief, err := s.briefRepo.ParseBrief()
	if err != nil {
		return domain.BriefMissing, "", fmt.Errorf("parse brief: %w", err)
	}

	gate := brief.GateResult

	switch reqType {
	case domain.TypeNewProject, domain.TypeComplexNew:
		if gate == domain.BriefMissing || gate == domain.BriefStub {
			return gate, "", fmt.Errorf(
				"brief gate BLOCKED for %s: brief is %s — fill in Vision, Key Deliverables, and Scope before starting",
				reqType, gate,
			)
		}
	case domain.TypeEnhancement:
		if gate == domain.BriefMissing {
			return gate, "brief is missing — proceeding with warning (required for NEW_PROJECT)", nil
		}
	case domain.TypeBugFix, domain.TypeRefactor:
		// SKIP — not required
	}

	return gate, "", nil
}

func typeToString(t domain.RequestType) string {
	switch t {
	case domain.TypeNewProject:
		return "new"
	case domain.TypeEnhancement:
		return "enhancement"
	case domain.TypeBugFix:
		return "bugfix"
	case domain.TypeRefactor:
		return "refactor"
	case domain.TypeComplexNew:
		return "new"
	default:
		return "enhancement"
	}
}

func buildBranchName(reqType domain.RequestType, descriptor string) string {
	prefix := strings.ToLower(string(reqType))
	prefix = strings.ReplaceAll(prefix, "_", "-")
	return fmt.Sprintf("%s/%s", prefix, descriptor)
}

// AgentChainFor returns the agent chain for the given request type.
func AgentChainFor(reqType domain.RequestType) []string {
	return agentChain(reqType)
}

func agentChain(reqType domain.RequestType) []string {
	switch reqType {
	case domain.TypeNewProject, domain.TypeComplexNew:
		return []string{"analyst", "architect", "developer", "tester", "reviewer"}
	case domain.TypeEnhancement:
		return []string{"analyst", "developer", "tester", "reviewer"}
	case domain.TypeBugFix:
		return []string{"developer", "tester", "reviewer"}
	case domain.TypeRefactor:
		return []string{"developer", "tester", "reviewer"}
	default:
		return []string{"analyst", "developer", "tester", "reviewer"}
	}
}

func createGitBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func currentBranch(dir string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
