package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
)

// PromptBuilder assembles agent orchestrator prompts from project context.
type PromptBuilder struct {
	projectRoot string
}

// NewPromptBuilder creates a PromptBuilder.
func NewPromptBuilder(projectRoot string) *PromptBuilder {
	return &PromptBuilder{projectRoot: projectRoot}
}

// Build assembles an orchestrator prompt for the given request.
func (b *PromptBuilder) Build(request string, reqType domain.RequestType, iterationPath string) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Mind Framework — AI Workflow Orchestrator\n\n")
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Request**: %s\n", request))
	sb.WriteString(fmt.Sprintf("**Type**: %s\n", reqType))
	sb.WriteString(fmt.Sprintf("**Iteration**: %s\n\n", iterationPath))

	// Agent chain
	chain := agentChain(reqType)
	sb.WriteString(fmt.Sprintf("**Agent Chain**: %s\n\n", strings.Join(chain, " → ")))

	sb.WriteString("---\n\n")

	// Project context
	sb.WriteString("## Project Context\n\n")

	contextFiles := []string{
		"docs/spec/project-brief.md",
		"docs/spec/requirements.md",
		"docs/spec/architecture.md",
	}
	for _, relPath := range contextFiles {
		content := b.readFile(relPath)
		if content != "" {
			sb.WriteString(fmt.Sprintf("### %s\n\n", relPath))
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	// Recent iterations (last 3)
	sb.WriteString("## Recent Iterations\n\n")
	recentIters := b.recentIterationOverviews(3)
	if len(recentIters) == 0 {
		sb.WriteString("No previous iterations.\n\n")
	} else {
		for _, entry := range recentIters {
			sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", entry[0], entry[1]))
		}
	}

	// Conventions
	conventionsDir := filepath.Join(b.projectRoot, ".mind", "conventions")
	if entries, err := os.ReadDir(conventionsDir); err == nil {
		sb.WriteString("## Conventions\n\n")
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				content := b.readFile(filepath.Join(".mind", "conventions", e.Name()))
				if content != "" {
					sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", e.Name(), content))
				}
			}
		}
	}

	// Orchestrator instructions
	orchInstr := b.readFile(".mind/agents/orchestrator.md")
	if orchInstr != "" {
		sb.WriteString("## Orchestrator Instructions\n\n")
		sb.WriteString(orchInstr)
		sb.WriteString("\n\n")
	}

	// MCP tools available (G4)
	sb.WriteString("## Available MCP Tools\n\n")
	sb.WriteString("The following mind_* tools are available via MCP for state tracking and quality gates:\n\n")
	sb.WriteString("| Tool | Purpose |\n")
	sb.WriteString("|------|---------|\n")
	sb.WriteString("| `mind_status` | Project health summary: docs completeness, workflow state, warnings |\n")
	sb.WriteString("| `mind_doctor` | Deep diagnostics with severity levels and suggested fixes |\n")
	sb.WriteString("| `mind_check_brief` | Business context gate — brief existence and section check |\n")
	sb.WriteString("| `mind_validate_docs` | 17-check documentation validation suite |\n")
	sb.WriteString("| `mind_validate_refs` | 11-check cross-reference validation |\n")
	sb.WriteString("| `mind_list_iterations` | All iterations with status and artifact completeness |\n")
	sb.WriteString("| `mind_show_iteration` | Detailed iteration info (overview + artifacts) |\n")
	sb.WriteString("| `mind_read_state` | Read current workflow state from docs/state/workflow.md |\n")
	sb.WriteString("| `mind_update_state` | Write workflow state (track agent progress) |\n")
	sb.WriteString("| `mind_create_iteration` | Create iteration directory with templates |\n")
	sb.WriteString("| `mind_list_stubs` | List stub documents needing content |\n")
	sb.WriteString("| `mind_check_gate` | Run deterministic gate (build/lint/test) |\n")
	sb.WriteString("| `mind_log_quality` | Log quality scores from convergence analysis |\n")
	sb.WriteString("| `mind_search_docs` | Full-text search across docs/ |\n")
	sb.WriteString("| `mind_read_config` | Read mind.toml as structured JSON |\n")
	sb.WriteString("| `mind_suggest_next` | Suggest next action based on project state |\n\n")
	sb.WriteString("Use `mind_read_state` and `mind_update_state` to track agent chain progress.\n")
	sb.WriteString("Use `mind_check_gate` before the reviewer agent to run the deterministic gate.\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("## Task\n\n")
	sb.WriteString(fmt.Sprintf("Execute the agent chain for the following request:\n\n> %s\n\n", request))
	sb.WriteString("Begin with the first agent in the chain. Follow the orchestrator instructions above.\n")

	return sb.String(), nil
}

// BuildAnalyze assembles a system prompt for the conversation analysis workflow.
func (b *PromptBuilder) BuildAnalyze(topic string) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Mind Framework — Conversation Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Topic**: %s\n\n", topic))

	sb.WriteString("---\n\n")

	// Moderator agent instructions
	moderator := b.readFile(".mind/conversation/agents/moderator.md")
	if moderator != "" {
		sb.WriteString("## Moderator Instructions\n\n")
		sb.WriteString(moderator)
		sb.WriteString("\n\n")
	}

	// Conversation configuration files
	configFiles := []string{
		".mind/conversation/config/conversation.yml",
		".mind/conversation/config/personas.yml",
		".mind/conversation/config/quality.yml",
		".mind/conversation/config/extensions.yml",
	}
	for _, relPath := range configFiles {
		content := b.readFile(relPath)
		if content != "" {
			sb.WriteString(fmt.Sprintf("### %s\n\n```yaml\n%s\n```\n\n", filepath.Base(relPath), content))
		}
	}

	// Project context (brief + requirements for Mode C awareness)
	for _, relPath := range []string{"docs/spec/project-brief.md", "docs/spec/requirements.md"} {
		content := b.readFile(relPath)
		if content != "" {
			sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", relPath, content))
		}
	}

	// Conventions
	conventionsDir := filepath.Join(b.projectRoot, ".mind", "conventions")
	if entries, err := os.ReadDir(conventionsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				content := b.readFile(filepath.Join(".mind", "conventions", e.Name()))
				if content != "" {
					sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", e.Name(), content))
				}
			}
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## Task\n\n")
	sb.WriteString(fmt.Sprintf("Run a dialectical conversation analysis on the following topic:\n\n> %s\n\n", topic))
	sb.WriteString("Follow the moderator instructions above. Produce a convergence analysis saved to docs/knowledge/.\n")

	return sb.String(), nil
}

// BuildDiscover assembles a system prompt for the interactive discovery workflow.
func (b *PromptBuilder) BuildDiscover(idea string) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Mind Framework — Project Discovery\n\n")
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Idea**: %s\n\n", idea))

	sb.WriteString("---\n\n")

	// Discovery agent instructions
	discovery := b.readFile(".mind/agents/discovery.md")
	if discovery != "" {
		sb.WriteString("## Discovery Agent Instructions\n\n")
		sb.WriteString(discovery)
		sb.WriteString("\n\n")
	}

	// Existing brief (for update-not-replace behavior)
	brief := b.readFile("docs/spec/project-brief.md")
	if brief != "" {
		sb.WriteString("## Existing Project Brief\n\n")
		sb.WriteString(brief)
		sb.WriteString("\n\n")
	}

	// Existing requirements (for awareness check)
	reqs := b.readFile("docs/spec/requirements.md")
	if reqs != "" {
		sb.WriteString("## Existing Requirements\n\n")
		sb.WriteString(reqs)
		sb.WriteString("\n\n")
	}

	// Conventions
	conventionsDir := filepath.Join(b.projectRoot, ".mind", "conventions")
	if entries, err := os.ReadDir(conventionsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				content := b.readFile(filepath.Join(".mind", "conventions", e.Name()))
				if content != "" {
					sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", e.Name(), content))
				}
			}
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## Task\n\n")
	sb.WriteString(fmt.Sprintf("Run interactive project discovery for the following idea:\n\n> %s\n\n", idea))
	sb.WriteString("Follow the discovery agent instructions above. Ask targeted questions and produce docs/spec/project-brief.md.\n")

	return sb.String(), nil
}

func (b *PromptBuilder) readFile(relPath string) string {
	absPath := filepath.Join(b.projectRoot, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}
	return string(data)
}

func (b *PromptBuilder) recentIterationOverviews(n int) [][2]string {
	iterDir := filepath.Join(b.projectRoot, "docs", "iterations")
	entries, err := os.ReadDir(iterDir)
	if err != nil {
		return nil
	}

	// Collect directories in reverse order (newest first)
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	// Reverse to get newest first
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	var results [][2]string
	for _, dir := range dirs {
		if len(results) >= n {
			break
		}
		overviewPath := filepath.Join(iterDir, dir, "overview.md")
		data, err := os.ReadFile(overviewPath)
		if err != nil {
			continue
		}
		results = append(results, [2]string{dir, string(data)})
	}
	return results
}
