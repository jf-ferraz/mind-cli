package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var briefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Show project brief status and completeness",
	RunE:  runBrief,
}

func init() {
	rootCmd.AddCommand(briefCmd)
}

func runBrief(cmd *cobra.Command, args []string) error {
	brief, err := briefRepo.ParseBrief()
	if err != nil {
		return fmt.Errorf("failed to parse brief: %w", err)
	}

	fmt.Print(renderer.RenderBrief(brief))
	return nil
}
