package cmd

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
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
	root, err := resolveRoot()
	if err != nil {
		return err
	}

	iterRepo := fs.NewIterationRepo(root)
	iters, err := iterRepo.List()
	if err != nil {
		return fmt.Errorf("failed to list iterations: %w", err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderIterations(iters))
	return nil
}
