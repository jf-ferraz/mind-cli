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

	sb.WriteString("---\n\n")
	sb.WriteString("## Task\n\n")
	sb.WriteString(fmt.Sprintf("Execute the agent chain for the following request:\n\n> %s\n\n", request))
	sb.WriteString("Begin with the first agent in the chain. Use `/workflow` in Claude Code to start.\n")

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
