package orchestrate

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// HandoffResult holds the output from a completed handoff run.
type HandoffResult struct {
	IterationID      string
	ArtifactsPresent int
	ArtifactsTotal   int
	MissingArtifacts []string
	GateResult       *domain.GateResult
	StateCleared     bool
	Branch           string
	AheadBy          int
	Artifacts        []string
	Errors           []string
}

// HandoffService encapsulates the 5-step handoff sequence.
type HandoffService struct {
	projectRoot   string
	iterRepo      repo.IterationRepo
	stateRepo     repo.StateRepo
	validationSvc *service.ValidationService
}

// NewHandoffService creates a HandoffService.
func NewHandoffService(
	projectRoot string,
	iterRepo repo.IterationRepo,
	stateRepo repo.StateRepo,
	validationSvc *service.ValidationService,
) *HandoffService {
	return &HandoffService{
		projectRoot:   projectRoot,
		iterRepo:      iterRepo,
		stateRepo:     stateRepo,
		validationSvc: validationSvc,
	}
}

// Run executes all 5 handoff steps for the given iteration ID.
// defaultBranch is the branch to compare against for git rev-list (e.g., "main").
func (s *HandoffService) Run(iterID, defaultBranch string) (*HandoffResult, error) {
	result := &HandoffResult{}

	// Step 1: Look up iteration
	iters, err := s.iterRepo.List()
	if err != nil {
		return nil, fmt.Errorf("list iterations: %w", err)
	}
	var iter *domain.Iteration
	for i := range iters {
		if iters[i].DirName == iterID {
			iter = &iters[i]
			break
		}
	}
	if iter == nil {
		return nil, fmt.Errorf("iteration %q not found", iterID)
	}
	result.IterationID = iter.DirName

	// Step 2: Validate artifacts
	result.ArtifactsTotal = len(domain.ExpectedArtifacts)
	for _, a := range iter.Artifacts {
		if a.Exists {
			result.ArtifactsPresent++
			result.Artifacts = append(result.Artifacts, a.Name)
		} else {
			result.MissingArtifacts = append(result.MissingArtifacts, a.Name)
		}
	}

	// Step 3: Run deterministic gate
	result.GateResult = s.validationSvc.RunGate(s.projectRoot)

	// Step 4: Append to docs/state/current.md
	if err := s.stateRepo.AppendCurrentState(iter); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("update current.md: %v", err))
	}

	// Step 5: Clear workflow state
	if err := s.stateRepo.WriteWorkflow(nil); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("clear workflow state: %v", err))
	} else {
		result.StateCleared = true
	}

	// Report branch status
	result.Branch = currentGitBranch(s.projectRoot)
	result.AheadBy = branchAhead(s.projectRoot, defaultBranch)

	return result, nil
}

func currentGitBranch(dir string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func branchAhead(dir, defaultBranch string) int {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD..."+defaultBranch)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}
