package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	flagStrict bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run validation checks",
	Long:  "Run documentation, cross-reference, and config validation checks.",
}

var checkDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Run the 17-check documentation validation suite",
	RunE:  runCheckDocs,
}

var checkRefsCmd = &cobra.Command{
	Use:   "refs",
	Short: "Run the 11-check cross-reference validation suite",
	RunE:  runCheckRefs,
}

var checkConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate mind.toml schema",
	RunE:  runCheckConfig,
}

var checkAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all validation suites (docs, refs, config)",
	RunE:  runCheckAll,
}

func init() {
	checkDocsCmd.Flags().BoolVar(&flagStrict, "strict", false, "Promote warnings to failures")
	checkAllCmd.Flags().BoolVar(&flagStrict, "strict", false, "Promote warnings to failures")

	checkCmd.AddCommand(checkDocsCmd)
	checkCmd.AddCommand(checkRefsCmd)
	checkCmd.AddCommand(checkConfigCmd)
	checkCmd.AddCommand(checkAllCmd)
	rootCmd.AddCommand(checkCmd)
}

func runCheckDocs(cmd *cobra.Command, args []string) error {
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

	svc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunDocs(root, flagStrict)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckRefs(cmd *cobra.Command, args []string) error {
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

	svc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunRefs(root)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckConfig(cmd *cobra.Command, args []string) error {
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

	svc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunConfig(root)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckAll(cmd *cobra.Command, args []string) error {
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

	svc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	report := svc.RunAll(root, flagStrict)

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderUnifiedValidation(&report))

	if report.Summary.Failed > 0 {
		os.Exit(1)
	}
	return nil
}
