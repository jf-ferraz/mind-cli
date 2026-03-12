package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Inspect workflow state and iteration history",
}

var workflowStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current workflow state",
	RunE:  runWorkflowStatus,
}

var workflowHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "List all iterations chronologically",
	RunE:  runWorkflowHistory,
}

func init() {
	workflowCmd.AddCommand(workflowStatusCmd)
	workflowCmd.AddCommand(workflowHistoryCmd)
	rootCmd.AddCommand(workflowCmd)
}

func runWorkflowStatus(cmd *cobra.Command, args []string) error {
	ws, err := workflowSvc.Status()
	if err != nil {
		return fmt.Errorf("read workflow state: %w", err)
	}

	fmt.Print(renderer.RenderWorkflowStatus(ws))
	return nil
}

func runWorkflowHistory(cmd *cobra.Command, args []string) error {
	history, err := workflowSvc.History()
	if err != nil {
		return fmt.Errorf("list iterations: %w", err)
	}

	fmt.Print(renderer.RenderWorkflowHistory(history))
	return nil
}
