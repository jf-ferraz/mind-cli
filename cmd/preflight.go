package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jf-ferraz/mind-cli/internal/orchestrate"
	"github.com/spf13/cobra"
)

var flagPreflightResume bool

var preflightCmd = &cobra.Command{
	Use:   "preflight [request]",
	Short: "Run pre-flight checks and prepare an AI workflow",
	Long: `mind preflight runs the 7-step pre-flight sequence:
  1. Classify the request type
  2. Check the business context gate (brief)
  3. Validate documentation (17 checks)
  4. Create an iteration folder
  5. Create a git branch
  6. Assemble context package
  7. Generate orchestrator prompt

Use --resume to check for an in-progress workflow.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPreflight,
}

func init() {
	preflightCmd.Flags().BoolVar(&flagPreflightResume, "resume", false, "Check for in-progress workflow and display resumption info")
	rootCmd.AddCommand(preflightCmd)
}

func runPreflight(cmd *cobra.Command, args []string) error {
	if flagPreflightResume {
		return runPreflightResume()
	}

	if len(args) == 0 {
		return exitValidation(fmt.Errorf("request string required: mind preflight \"<describe your request>\""))
	}
	return runPreflightNew(args[0])
}

func runPreflightNew(request string) error {
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

	if flagJSON {
		out, merr := json.MarshalIndent(result, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal preflight result: %w", merr))
		}
		fmt.Println(string(out))
		return nil
	}

	fmt.Println(renderPreflightResult(result))
	return nil
}

func runPreflightResume() error {
	svc := orchestrate.NewPreflightService(
		projectRoot,
		briefRepo,
		stateRepo,
		validationSvc,
		generateSvc,
		nil,
	)

	state, err := svc.Resume()
	if err != nil {
		return exitRuntime(fmt.Errorf("read workflow state: %w", err))
	}

	if state == nil || state.IsIdle() {
		fmt.Println("No resumable workflow found.")
		iters, lerr := workflowSvc.History()
		if lerr == nil && iters != nil && len(iters.Iterations) > 0 {
			last := iters.Iterations[0]
			fmt.Printf("Last completed: %s (%s)\n", last.DirName, last.CreatedAt)
		}
		return nil
	}

	fmt.Printf("+-- Resumable Workflow Found -----------------------------------+\n")
	fmt.Printf("|\n")
	fmt.Printf("|  Type: %s\n", state.Type)
	fmt.Printf("|  Iteration: %s\n", state.IterationPath)
	fmt.Printf("|  Last Agent: %s\n", state.LastAgent)
	if len(state.RemainingChain) > 0 {
		fmt.Printf("|  Remaining: %s\n", strings.Join(state.RemainingChain, " → "))
	}
	fmt.Printf("|  Branch: %s\n", state.Branch)
	if state.TotalSessions > 1 {
		fmt.Printf("|  Session: %d of %d\n", state.Session, state.TotalSessions)
	}
	fmt.Printf("|\n")
	if len(state.Artifacts) > 0 {
		fmt.Printf("|  Completed Artifacts:\n")
		for _, a := range state.Artifacts {
			fmt.Printf("|    [done] %s (%s)\n", a.Output, a.Agent)
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("|  Resume in Claude Code:\n")
	fmt.Printf("|  /workflow --resume\n")
	fmt.Printf("+---------------------------------------------------------------+\n")
	return nil
}

func renderPreflightResult(result *orchestrate.PreflightResult) string {
	var sb strings.Builder

	sb.WriteString("+-- Pre-Flight Complete ----------------------------------------+\n")
	sb.WriteString("|\n")
	sb.WriteString(fmt.Sprintf("|  Type: %s\n", result.RequestType))

	chain := orchestrate.AgentChainFor(result.RequestType)
	if len(chain) > 0 {
		sb.WriteString(fmt.Sprintf("|  Chain: %s\n", strings.Join(chain, " → ")))
	}

	sb.WriteString(fmt.Sprintf("|  Branch: %s\n", result.Branch))
	sb.WriteString(fmt.Sprintf("|  Iteration: %s\n", result.IterationPath))

	briefStatus := string(result.BriefGate)
	sb.WriteString(fmt.Sprintf("|  Brief: %s\n", briefStatus))

	total := result.DocsReport.Total
	passed := result.DocsReport.Passed
	failed := result.DocsReport.Failed
	warned := result.DocsReport.Warnings
	sb.WriteString(fmt.Sprintf("|  Docs: %d/%d pass (%d warnings, %d blockers)\n", passed, total, warned, failed))

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("|\n|  WARNING: %s\n", w))
	}

	sb.WriteString("|\n")
	sb.WriteString("|  Context package ready. Open Claude Code and run:\n")
	sb.WriteString(fmt.Sprintf("|  /workflow %q\n", request(result)))
	sb.WriteString("|\n")

	if result.Prompt != "" {
		sb.WriteString("|  Orchestrator prompt generated.\n")
	}

	sb.WriteString("+---------------------------------------------------------------+\n")
	return sb.String()
}

func request(result *orchestrate.PreflightResult) string {
	return result.Descriptor
}
