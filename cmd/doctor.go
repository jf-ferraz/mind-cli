package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
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
	root, err := resolveRoot()
	if err != nil {
		if isNotProject(err) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return err
	}

	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)
	configRepo := fs.NewConfigRepo(root)

	svc := service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo)
	report := svc.Run(flagFix)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderDoctor(report))

	if report.Summary.Fail > 0 {
		os.Exit(1)
	}
	return nil
}
