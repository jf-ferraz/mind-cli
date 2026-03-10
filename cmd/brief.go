package cmd

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
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
	root, err := resolveRoot()
	if err != nil {
		return err
	}

	docRepo := fs.NewDocRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)
	brief, err := briefRepo.ParseBrief()
	if err != nil {
		return fmt.Errorf("failed to parse brief: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderBrief(brief))
	return nil
}
