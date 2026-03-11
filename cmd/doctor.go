package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var flagFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run full project diagnostics",
	Long:  "Checks framework installation, documentation structure, cross-references, config, and workflow state.",
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&flagFix, "fix", false, "Auto-fix resolvable issues")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	report := doctorSvc.Run(flagFix)

	fmt.Print(renderer.RenderDoctor(report))

	if report.Summary.Fail > 0 {
		os.Exit(1)
	}
	return nil
}
