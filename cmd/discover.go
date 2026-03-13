package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jf-ferraz/mind-cli/internal/orchestrate"
	"github.com/spf13/cobra"
)

var (
	flagDiscoverOutput bool
	flagDiscoverModel  string
)

var discoverCmd = &cobra.Command{
	Use:   "discover [idea]",
	Short: "Run interactive project discovery",
	Long: `mind discover explores and defines a project idea through
targeted questions, producing a structured project brief.

This is the terminal equivalent of /discover in Claude Code.

Examples:
  mind discover "inventory management system for small warehouses"
  mind discover "real-time collaboration feature" --model sonnet
  mind discover "monitoring dashboard for microservices" --output`,
	Args: cobra.ExactArgs(1),
	RunE: runDiscover,
}

func init() {
	discoverCmd.Flags().BoolVarP(&flagDiscoverOutput, "output", "o", false, "Output prompt only (do not launch claude)")
	discoverCmd.Flags().StringVar(&flagDiscoverModel, "model", "", "Model override (e.g. opus, sonnet)")
	rootCmd.AddCommand(discoverCmd)
}

func runDiscover(cmd *cobra.Command, args []string) error {
	idea := args[0]

	promptBuilder := orchestrate.NewPromptBuilder(projectRoot)
	systemPrompt, err := promptBuilder.BuildDiscover(idea)
	if err != nil {
		return exitRuntime(fmt.Errorf("build discover prompt: %w", err))
	}

	runner := orchestrate.NewRunner()
	mode := orchestrate.RunModeAuto
	if flagDiscoverOutput {
		mode = orchestrate.RunModeOutput
	}

	cfg := orchestrate.RunConfig{
		SystemPrompt: systemPrompt,
		Request:      fmt.Sprintf("Discover: %s", idea),
		ProjectRoot:  projectRoot,
		Mode:         mode,
		Model:        flagDiscoverModel,
	}

	if flagJSON {
		out := map[string]any{
			"idea":        idea,
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

	fmt.Printf("+-- Project Discovery -----------------------------------------+\n")
	fmt.Printf("|  Idea: %s\n", idea)
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
