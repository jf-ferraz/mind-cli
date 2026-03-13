package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRunner_DetectsClaude(t *testing.T) {
	runner := NewRunner()
	// On this machine claude is installed; just verify no panic and consistent results.
	if runner.HasClaude() {
		if runner.ClaudePath() == "" {
			t.Error("HasClaude() is true but ClaudePath() is empty")
		}
	} else {
		if runner.ClaudePath() != "" {
			t.Error("HasClaude() is false but ClaudePath() is non-empty")
		}
	}
}

func TestRunner_Run_OutputMode_ReturnsPrompt(t *testing.T) {
	runner := NewRunner()

	cfg := RunConfig{
		SystemPrompt: "test system prompt",
		Request:      "test request",
		ProjectRoot:  t.TempDir(),
		Mode:         RunModeOutput,
	}

	result, err := runner.Run(cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Launched {
		t.Error("Run() in output mode should not launch claude")
	}
	if result.Prompt != "test system prompt" {
		t.Errorf("Prompt = %q, want 'test system prompt'", result.Prompt)
	}
}

func TestRunner_buildArgs_SystemPrompt(t *testing.T) {
	runner := &Runner{claudePath: "/usr/bin/claude"}

	cfg := RunConfig{
		SystemPrompt: "my system prompt",
		Request:      "my request",
		ProjectRoot:  "/tmp/project",
	}

	args := runner.buildArgs(cfg)

	// Should contain --system-prompt
	found := false
	for i, arg := range args {
		if arg == "--system-prompt" && i+1 < len(args) && args[i+1] == "my system prompt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("buildArgs() = %v, want --system-prompt 'my system prompt'", args)
	}

	// Last arg should be the request
	if args[len(args)-1] != "my request" {
		t.Errorf("last arg = %q, want 'my request'", args[len(args)-1])
	}
}

func TestRunner_buildArgs_MCPConfig(t *testing.T) {
	root := t.TempDir()
	runner := &Runner{claudePath: "/usr/bin/claude"}

	// Without .mcp.json
	cfg := RunConfig{
		Request:     "test",
		ProjectRoot: root,
	}
	args := runner.buildArgs(cfg)
	for _, arg := range args {
		if arg == "--mcp-config" {
			t.Error("buildArgs() should not include --mcp-config when .mcp.json is missing")
		}
	}

	// With .mcp.json
	mcpPath := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte(`{"mcpServers":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	args = runner.buildArgs(cfg)
	foundMCP := false
	for i, arg := range args {
		if arg == "--mcp-config" && i+1 < len(args) && args[i+1] == mcpPath {
			foundMCP = true
			break
		}
	}
	if !foundMCP {
		t.Errorf("buildArgs() = %v, want --mcp-config %s", args, mcpPath)
	}
}

func TestRunner_buildArgs_Model(t *testing.T) {
	runner := &Runner{claudePath: "/usr/bin/claude"}

	cfg := RunConfig{
		Request:     "test",
		ProjectRoot: t.TempDir(),
		Model:       "opus",
	}

	args := runner.buildArgs(cfg)
	found := false
	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) && args[i+1] == "opus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("buildArgs() = %v, want --model opus", args)
	}
}

func TestFormatPromptOutput_NoClaude(t *testing.T) {
	out := FormatPromptOutput("system", "request", false)
	if !strings.Contains(out, "claude CLI not found") {
		t.Error("output should mention claude CLI not found")
	}
	if !strings.Contains(out, "system") {
		t.Error("output should contain system prompt")
	}
	if !strings.Contains(out, "request") {
		t.Error("output should contain request")
	}
}

func TestFormatPromptOutput_WithClaude(t *testing.T) {
	out := FormatPromptOutput("system", "request", true)
	if strings.Contains(out, "claude CLI not found") {
		t.Error("output should not mention missing claude when available")
	}
}
