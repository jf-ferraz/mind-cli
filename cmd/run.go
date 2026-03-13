package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jf-ferraz/mind-cli/internal/orchestrate"
	"github.com/spf13/cobra"
)

var (
	flagRunOutput bool
	flagRunModel  string
)

var workflowRunCmd = &cobra.Command{
	Use:   "run [request]",
	Short: "Execute a full AI workflow from the terminal",
	Long: `mind workflow run executes the complete agent pipeline:
  1. Classify the request type
  2. Run pre-flight checks (brief gate, doc validation)
  3. Create iteration folder and git branch
  4. Generate orchestrator prompt
  5. Launch claude CLI with the prompt (or output it with --output)

If claude CLI is not installed, falls back to prompt output automatically.

Examples:
  mind workflow run "add user authentication"
  mind workflow run "fix the login bug" --model sonnet
  mind workflow run "refactor the API layer" --output`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkflowRun,
}

func init() {
	workflowRunCmd.Flags().BoolVarP(&flagRunOutput, "output", "o", false, "Output prompt only (do not launch claude)")
	workflowRunCmd.Flags().StringVar(&flagRunModel, "model", "", "Model override (e.g. opus, sonnet)")
	workflowCmd.AddCommand(workflowRunCmd)
}

func runWorkflowRun(cmd *cobra.Command, args []string) error {
	request := args[0]

	// Step 1-7: Run preflight
	promptBuilder := orchestrate.NewPromptBuilder(projectRoot)
	svc := orchestrate.NewPreflightService(
		projectRoot,
		briefRepo,
		stateRepo,
		validationSvc,
		generateSvc,
		promptBuilder,
	)

	result, err := svc.Run(request)
	if err != nil {
		return exitValidation(err)
	}

	// Display preflight summary
	if !flagJSON {
		fmt.Println(renderRunPreflightSummary(result))
	}

	if result.Prompt == "" {
		return exitRuntime(fmt.Errorf("prompt generation failed — check warnings above"))
	}

	// Determine run mode
	runner := orchestrate.NewRunner()
	mode := orchestrate.RunModeAuto
	if flagRunOutput {
		mode = orchestrate.RunModeOutput
	}

	cfg := orchestrate.RunConfig{
		SystemPrompt: result.Prompt,
		Request:      fmt.Sprintf("Execute the %s workflow for: %s", result.RequestType, request),
		ProjectRoot:  projectRoot,
		Mode:         mode,
		Model:        flagRunModel,
	}

	if flagJSON {
		return runWorkflowRunJSON(runner, cfg, result)
	}

	runResult, err := runner.Run(cfg)
	if err != nil {
		return exitRuntime(fmt.Errorf("claude exited with error: %w", err))
	}

	if !runResult.Launched {
		fmt.Print(orchestrate.FormatPromptOutput(cfg.SystemPrompt, cfg.Request, runner.HasClaude()))
	}

	return nil
}

func runWorkflowRunJSON(runner *orchestrate.Runner, cfg orchestrate.RunConfig, preflight *orchestrate.PreflightResult) error {
	out := map[string]any{
		"request_type":   preflight.RequestType,
		"descriptor":     preflight.Descriptor,
		"branch":         preflight.Branch,
		"iteration_path": preflight.IterationPath,
		"brief_gate":     preflight.BriefGate,
		"claude_found":   runner.HasClaude(),
		"output_mode":    cfg.Mode == orchestrate.RunModeOutput,
		"prompt":         cfg.SystemPrompt,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return exitRuntime(fmt.Errorf("marshal result: %w", err))
	}
	fmt.Println(string(data))
	return nil
}

func renderRunPreflightSummary(result *orchestrate.PreflightResult) string {
	var sb strings.Builder

	chain := orchestrate.AgentChainFor(result.RequestType)

	sb.WriteString("+-- Workflow Run -----------------------------------------------+\n")
	sb.WriteString("|\n")
	sb.WriteString(fmt.Sprintf("|  Type: %s\n", result.RequestType))
	if len(chain) > 0 {
		sb.WriteString(fmt.Sprintf("|  Chain: %s\n", strings.Join(chain, " → ")))
	}
	sb.WriteString(fmt.Sprintf("|  Branch: %s\n", result.Branch))
	sb.WriteString(fmt.Sprintf("|  Iteration: %s\n", result.IterationPath))
	sb.WriteString(fmt.Sprintf("|  Brief: %s\n", result.BriefGate))

	total := result.DocsReport.Total
	passed := result.DocsReport.Passed
	sb.WriteString(fmt.Sprintf("|  Docs: %d/%d pass\n", passed, total))

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("|  WARNING: %s\n", w))
	}

	sb.WriteString("|\n")
	sb.WriteString("+---------------------------------------------------------------+\n")
	return sb.String()
}
