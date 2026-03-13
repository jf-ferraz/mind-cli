package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jf-ferraz/mind-cli/internal/orchestrate"
	"github.com/spf13/cobra"
)

var (
	flagAnalyzeOutput bool
	flagAnalyzeModel  string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [topic]",
	Short: "Run a dialectical conversation analysis",
	Long: `mind analyze explores architectural options, evaluates trade-offs,
and produces a convergence analysis with ranked recommendations.

This is the terminal equivalent of /analyze in Claude Code.

Examples:
  mind analyze "Should we use GraphQL or REST for the API layer?"
  mind analyze "What's the best caching strategy?" --model opus
  mind analyze "Monolith vs microservices" --output`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().BoolVarP(&flagAnalyzeOutput, "output", "o", false, "Output prompt only (do not launch claude)")
	analyzeCmd.Flags().StringVar(&flagAnalyzeModel, "model", "", "Model override (e.g. opus, sonnet)")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	topic := args[0]

	promptBuilder := orchestrate.NewPromptBuilder(projectRoot)
	systemPrompt, err := promptBuilder.BuildAnalyze(topic)
	if err != nil {
		return exitRuntime(fmt.Errorf("build analyze prompt: %w", err))
	}

	runner := orchestrate.NewRunner()
	mode := orchestrate.RunModeAuto
	if flagAnalyzeOutput {
		mode = orchestrate.RunModeOutput
	}

	cfg := orchestrate.RunConfig{
		SystemPrompt: systemPrompt,
		Request:      fmt.Sprintf("Analyze: %s", topic),
		ProjectRoot:  projectRoot,
		Mode:         mode,
		Model:        flagAnalyzeModel,
	}

	if flagJSON {
		out := map[string]any{
			"topic":       topic,
			"claude_found": runner.HasClaude(),
			"output_mode": mode == orchestrate.RunModeOutput,
			"prompt":      systemPrompt,
		}
		data, merr := json.MarshalIndent(out, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal result: %w", merr))
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("+-- Conversation Analysis -------------------------------------+\n")
	fmt.Printf("|  Topic: %s\n", topic)
	fmt.Printf("+--------------------------------------------------------------+\n\n")

	runResult, err := runner.Run(cfg)
	if err != nil {
		return exitRuntime(fmt.Errorf("claude exited with error: %w", err))
	}

	if !runResult.Launched {
		fmt.Print(orchestrate.FormatPromptOutput(cfg.SystemPrompt, cfg.Request, runner.HasClaude()))
	}

	return nil
}
