package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/validate"
	"github.com/spf13/cobra"
)

var (
	flagStrict bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run documentation validation checks",
	Long:  "Runs the 17-check documentation validation suite (equivalent to validate-docs.sh).",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&flagStrict, "strict", false, "Promote warnings to failures (for CI)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot()
	if err != nil {
		return err
	}

	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)

	ctx := &validate.CheckContext{
		ProjectRoot: root,
		DocRepo:     docRepo,
		IterRepo:    iterRepo,
		BriefRepo:   briefRepo,
		Strict:      flagStrict,
	}

	suite := validate.DocsSuite()
	report := suite.Run(ctx)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}
