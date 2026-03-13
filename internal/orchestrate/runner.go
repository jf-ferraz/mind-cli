package orchestrate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunMode controls how the runner executes a prompt.
type RunMode int

const (
	// RunModeAuto launches claude CLI if available, falls back to prompt output.
	RunModeAuto RunMode = iota
	// RunModeOutput always outputs the prompt without launching claude.
	RunModeOutput
)

// RunConfig holds configuration for executing a workflow prompt.
type RunConfig struct {
	SystemPrompt string  // orchestrator context (passed as --system-prompt)
	Request      string  // user's request (passed as positional arg)
	ProjectRoot  string  // working directory for the subprocess
	Mode         RunMode // auto or output-only
	Model        string  // optional model override (e.g. "opus", "sonnet")
}

// RunResult holds the outcome of a run attempt.
type RunResult struct {
	Launched   bool   // true if claude CLI was launched
	ClaudePath string // path to claude binary ("" if not found)
	Prompt     string // the full prompt (system + request) for output mode
}

// Runner detects and launches the claude CLI, or falls back to prompt output.
type Runner struct {
	claudePath string
}

// NewRunner creates a Runner, detecting claude CLI on PATH.
func NewRunner() *Runner {
	path, _ := exec.LookPath("claude")
	return &Runner{claudePath: path}
}

// HasClaude returns true if the claude CLI was found.
func (r *Runner) HasClaude() bool {
	return r.claudePath != ""
}

// ClaudePath returns the detected path to the claude binary.
func (r *Runner) ClaudePath() string {
	return r.claudePath
}

// Run executes the prompt via claude CLI or returns it for manual use.
func (r *Runner) Run(cfg RunConfig) (*RunResult, error) {
	result := &RunResult{
		ClaudePath: r.claudePath,
		Prompt:     cfg.SystemPrompt,
	}

	if cfg.Mode == RunModeOutput || r.claudePath == "" {
		return result, nil
	}

	args := r.buildArgs(cfg)
	cmd := exec.Command(r.claudePath, args...)
	cmd.Dir = cfg.ProjectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	result.Launched = true
	return result, cmd.Run()
}

// buildArgs constructs the claude CLI arguments.
func (r *Runner) buildArgs(cfg RunConfig) []string {
	var args []string

	if cfg.SystemPrompt != "" {
		args = append(args, "--system-prompt", cfg.SystemPrompt)
	}

	// Add MCP config if .mcp.json exists in the project root.
	mcpPath := filepath.Join(cfg.ProjectRoot, ".mcp.json")
	if _, err := os.Stat(mcpPath); err == nil {
		args = append(args, "--mcp-config", mcpPath)
	}

	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	// The request becomes the initial user message.
	if cfg.Request != "" {
		args = append(args, cfg.Request)
	}

	return args
}

// FormatPromptOutput formats the prompt for terminal display when not launching claude.
func FormatPromptOutput(systemPrompt, request string, claudeAvailable bool) string {
	var out string

	if !claudeAvailable {
		out += "NOTE: claude CLI not found — outputting prompt for manual use.\n"
		out += "Install Claude Code (https://claude.ai/code) to enable automatic launch.\n\n"
	}

	out += "--- SYSTEM PROMPT ---\n"
	out += systemPrompt
	out += "\n--- END SYSTEM PROMPT ---\n\n"

	if request != "" {
		out += "--- REQUEST ---\n"
		out += request
		out += "\n--- END REQUEST ---\n"
	}

	out += fmt.Sprintf("\nTo use manually:\n  claude --system-prompt '<paste system prompt>' '%s'\n", request)

	return out
}
