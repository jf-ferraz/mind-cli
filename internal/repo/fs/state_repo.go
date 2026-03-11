package fs

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

// StateRepo implements repo.StateRepo using the filesystem.
type StateRepo struct {
	projectRoot string
}

// NewStateRepo creates a StateRepo.
func NewStateRepo(projectRoot string) *StateRepo {
	return &StateRepo{projectRoot: projectRoot}
}

// ReadWorkflow parses docs/state/workflow.md into structured state.
// The workflow file uses a markdown-based format with key-value pairs in a table
// or as YAML-like frontmatter within the markdown.
func (r *StateRepo) ReadWorkflow() (*domain.WorkflowState, error) {
	path := filepath.Join(r.projectRoot, "docs", "state", "workflow.md")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	state := &domain.WorkflowState{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse table rows: | Key | Value |
		if strings.HasPrefix(line, "|") && !strings.Contains(line, "---") {
			parts := strings.Split(line, "|")
			if len(parts) < 3 {
				continue
			}
			key := strings.TrimSpace(parts[1])
			val := strings.TrimSpace(parts[2])

			key = strings.ToLower(key)
			switch key {
			case "type":
				state.Type = domain.RequestType(val)
			case "descriptor", "request":
				state.Descriptor = val
			case "iteration_path", "iteration path", "iteration":
				state.IterationPath = val
			case "branch":
				state.Branch = val
			case "last_agent", "last agent":
				state.LastAgent = val
			case "remaining_chain", "remaining chain", "remaining":
				chain := strings.Split(val, ",")
				for i := range chain {
					chain[i] = strings.TrimSpace(chain[i])
				}
				if len(chain) > 0 && chain[0] != "" {
					state.RemainingChain = chain
				}
			case "session":
				state.Session, _ = strconv.Atoi(val)
			case "total_sessions", "total sessions":
				state.TotalSessions, _ = strconv.Atoi(val)
			}
		}

		// Parse YAML-like key: value lines
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "-") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(strings.ToLower(parts[0]))
			val := strings.TrimSpace(parts[1])

			switch key {
			case "type":
				state.Type = domain.RequestType(val)
			case "descriptor", "request":
				state.Descriptor = val
			case "iteration_path", "iteration path":
				state.IterationPath = val
			case "branch":
				state.Branch = val
			case "last_agent", "last agent":
				state.LastAgent = val
			case "remaining_chain", "remaining chain":
				chain := strings.Split(val, ",")
				for i := range chain {
					chain[i] = strings.TrimSpace(chain[i])
				}
				if len(chain) > 0 && chain[0] != "" {
					state.RemainingChain = chain
				}
			case "session":
				state.Session, _ = strconv.Atoi(val)
			case "total_sessions", "total sessions":
				state.TotalSessions, _ = strconv.Atoi(val)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if state.IsIdle() {
		return nil, nil
	}

	return state, nil
}
