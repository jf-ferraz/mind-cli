package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var handoffCmd = &cobra.Command{
	Use:   "handoff <iteration-id>",
	Short: "Validate iteration artifacts, run deterministic gate, and update state",
	Long: `mind handoff runs the 5-step post-workflow sequence:
  1. Validate 5 iteration artifacts (overview, changes, test-summary, validation, retrospective)
  2. Run deterministic gate (build/lint/test from mind.toml)
  3. Update docs/state/current.md
  4. Clear docs/state/workflow.md → idle
  5. Report branch status`,
	Args: cobra.ExactArgs(1),
	RunE: runHandoff,
}

func init() {
	rootCmd.AddCommand(handoffCmd)
}

func runHandoff(cmd *cobra.Command, args []string) error {
	iterID := args[0]

	// Read default branch from config, fall back to "main"
	defaultBranch := "main"
	cfg, _ := configRepo.ReadProjectConfig()
	if cfg != nil && cfg.Governance.DefaultBranch != "" {
		defaultBranch = cfg.Governance.DefaultBranch
	}

	result, err := handoffSvc.Run(iterID, defaultBranch)
	if err != nil {
		return exitValidation(fmt.Errorf("iteration %q not found: %w", iterID, err))
	}

	// Step 1: Validate iteration completeness
	fmt.Printf("Step 1: Validate iteration completeness\n")
	fmt.Printf("  Checking %s/\n", result.IterationID)
	for _, a := range result.Artifacts {
		fmt.Printf("    %-20s [present]\n", a)
	}
	for _, a := range result.MissingArtifacts {
		fmt.Printf("    %-20s [MISSING]\n", a)
	}
	fmt.Printf("  Result: %d/%d artifacts present\n\n", result.ArtifactsPresent, result.ArtifactsTotal)

	if len(result.MissingArtifacts) > 0 {
		fmt.Printf("  Missing artifacts:\n")
		for _, m := range result.MissingArtifacts {
			fmt.Printf("    - %s\n", m)
		}
		fmt.Println()
	}

	// Step 2: Run deterministic checks
	fmt.Printf("Step 2: Run deterministic checks\n")
	gateResult := result.GateResult
	if gateResult == nil || gateResult.Total == 0 {
		fmt.Printf("  No commands configured in mind.toml [project.commands]\n\n")
	} else {
		for _, cr := range gateResult.Commands {
			status := "pass"
			if !cr.Pass {
				status = "FAIL"
			}
			fmt.Printf("    %-30s [%s] (%.1fs)\n", cr.Command, status, cr.Duration.Seconds())
			if !cr.Pass && cr.Stderr != "" {
				lines := strings.Split(cr.Stderr, "\n")
				n := len(lines)
				if n > 5 {
					n = 5
				}
				for _, line := range lines[:n] {
					if line != "" {
						fmt.Printf("      %s\n", line)
					}
				}
			}
		}
		fmt.Printf("  Result: %d/%d pass\n\n", gateResult.Passed, gateResult.Total)
	}

	// Step 3: Update docs/state/current.md
	fmt.Printf("Step 3: Update docs/state/current.md\n")
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			if strings.HasPrefix(e, "update current.md:") {
				fmt.Printf("  WARNING: could not update current.md: %s\n\n", strings.TrimPrefix(e, "update current.md: "))
			}
		}
	} else {
		fmt.Printf("  Recent Changes: + %s\n\n", result.IterationID)
	}

	// Step 4: Clear workflow state
	fmt.Printf("Step 4: Clear workflow state\n")
	clearErr := false
	for _, e := range result.Errors {
		if strings.HasPrefix(e, "clear workflow state:") {
			fmt.Printf("  WARNING: could not clear workflow state: %s\n\n", strings.TrimPrefix(e, "clear workflow state: "))
			clearErr = true
		}
	}
	if !clearErr {
		fmt.Printf("  docs/state/workflow.md → idle\n\n")
	}

	// Step 5: Report branch status
	fmt.Printf("Step 5: Report branch status\n")
	if result.Branch != "" {
		fmt.Printf("  Branch: %s\n", result.Branch)
		if result.AheadBy > 0 {
			fmt.Printf("  Commits ahead of %s: %d\n", defaultBranch, result.AheadBy)
		}
		fmt.Printf("  Suggestion: Create PR with 'gh pr create'\n")
	}

	if gateResult != nil && !gateResult.Pass && gateResult.Total > 0 {
		fmt.Printf("\n  Gate FAILED — review failures before creating PR\n")
		return exitValidation(fmt.Errorf("deterministic gate failed: %d/%d pass", gateResult.Passed, gateResult.Total))
	}

	return nil
}
