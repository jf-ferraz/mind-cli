package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var iterationsCmd = &cobra.Command{
	Use:     "iterations",
	Aliases: []string{"iter", "iters"},
	Short:   "List all iterations",
	RunE:    runIterations,
}

func init() {
	rootCmd.AddCommand(iterationsCmd)
}

func runIterations(cmd *cobra.Command, args []string) error {
	iters, err := iterRepo.List()
	if err != nil {
		return fmt.Errorf("failed to list iterations: %w", err)
	}

	fmt.Print(renderer.RenderIterations(iters))
	return nil
}
