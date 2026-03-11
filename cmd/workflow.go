package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
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
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	stateRepo := fs.NewStateRepo(root)
	iterRepo := fs.NewIterationRepo(root)

	svc := service.NewWorkflowService(stateRepo, iterRepo)
	ws, err := svc.Status()
	if err != nil {
		return fmt.Errorf("read workflow state: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderWorkflowStatus(ws))
	return nil
}

func runWorkflowHistory(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	stateRepo := fs.NewStateRepo(root)
	iterRepo := fs.NewIterationRepo(root)

	svc := service.NewWorkflowService(stateRepo, iterRepo)
	history, err := svc.History()
	if err != nil {
		return fmt.Errorf("list iterations: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderWorkflowHistory(history))
	return nil
}
