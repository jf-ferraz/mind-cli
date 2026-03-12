package fs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// AppendCurrentState appends a completed iteration entry to docs/state/current.md.
func (r *StateRepo) AppendCurrentState(iter *domain.Iteration) error {
	if iter == nil {
		return fmt.Errorf("iter must not be nil")
	}
	currentPath := filepath.Join(r.projectRoot, "docs", "state", "current.md")
	data, err := os.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("docs/state/current.md does not exist: %w", err)
		}
		return err
	}

	content := string(data)
	entry := fmt.Sprintf("- **%s** — %s completed (@iteration/%03d)\n",
		currentDate(),
		iter.DirName,
		iter.Seq,
	)

	const marker = "## Recent Changes\n"
	if idx := strings.Index(content, marker); idx >= 0 {
		insertAt := idx + len(marker) + 1 // after the blank line following the header
		if insertAt > len(content) {
			insertAt = len(content)
		}
		content = content[:insertAt] + entry + content[insertAt:]
	} else {
		content += "\n## Recent Changes\n\n" + entry
	}

	return os.WriteFile(currentPath, []byte(content), 0644)
}

// WriteWorkflow persists workflow state to docs/state/workflow.md.
// Passing nil writes an idle marker.
func (r *StateRepo) WriteWorkflow(state *domain.WorkflowState) error {
	path := filepath.Join(r.projectRoot, "docs", "state", "workflow.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	var b strings.Builder
	b.WriteString("# Workflow State\n\n")

	if state == nil || state.IsIdle() {
		b.WriteString("| Key | Value |\n")
		b.WriteString("|-----|-------|\n")
		b.WriteString("| type | |\n")
		b.WriteString("| descriptor | |\n")
		return os.WriteFile(path, []byte(b.String()), 0644)
	}

	b.WriteString("| Key | Value |\n")
	b.WriteString("|-----|-------|\n")
	b.WriteString(fmt.Sprintf("| type | %s |\n", state.Type))
	b.WriteString(fmt.Sprintf("| descriptor | %s |\n", state.Descriptor))
	b.WriteString(fmt.Sprintf("| iteration_path | %s |\n", state.IterationPath))
	b.WriteString(fmt.Sprintf("| branch | %s |\n", state.Branch))
	b.WriteString(fmt.Sprintf("| last_agent | %s |\n", state.LastAgent))
	b.WriteString(fmt.Sprintf("| remaining_chain | %s |\n", strings.Join(state.RemainingChain, ", ")))
	b.WriteString(fmt.Sprintf("| session | %d |\n", state.Session))
	b.WriteString(fmt.Sprintf("| total_sessions | %d |\n", state.TotalSessions))

	if state.HandoffContext != "" {
		b.WriteString("\n## Handoff Context\n\n")
		b.WriteString(state.HandoffContext + "\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func currentDate() string {
	return time.Now().Format("2006-01-02")
}
