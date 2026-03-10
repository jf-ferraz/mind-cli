package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagNoColor bool
	flagProject string
)

var rootCmd = &cobra.Command{
	Use:   "mind",
	Short: "Mind Framework CLI — project intelligence at your fingertips",
	Long: `mind is the command-line companion for the Mind Agent Framework.
It inspects documentation structure, validates quality gates,
manages iterations, and bridges AI agent workflows.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "Path to project root (default: auto-detect)")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
