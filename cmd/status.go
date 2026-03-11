package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project health and documentation status",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	project, err := fs.DetectProject(root)
	if err != nil {
		return err
	}

	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	stateRepo := fs.NewStateRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)

	svc := service.NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
	health, err := svc.AssembleHealth(project)
	if err != nil {
		return err
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderHealth(health))

	// Exit code 1 if issues found
	hasIssues := false
	if health.Brief.GateResult == domain.BriefMissing {
		hasIssues = true
	}
	for _, zh := range health.Zones {
		if zh.Stubs > 0 {
			hasIssues = true
		}
	}
	if hasIssues {
		os.Exit(1)
	}

	return nil
}
