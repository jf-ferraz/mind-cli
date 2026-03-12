package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI dashboard",
	Long: `Launch a full-screen interactive dashboard with 5 tabs:
Status, Documents, Iterations, Checks, and Quality.

The TUI provides a real-time view of project health, document browsing,
validation results, and convergence quality tracking.`,
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	deps := BuildDeps(root, nil)
	app := tui.NewApp(deps, Version)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
