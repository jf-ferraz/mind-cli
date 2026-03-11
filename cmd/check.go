package cmd

import (
	"fmt"
	"os"

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
	report := validationSvc.RunDocs(projectRoot, flagStrict)

	fmt.Print(renderer.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckRefs(cmd *cobra.Command, args []string) error {
	report := validationSvc.RunRefs(projectRoot)

	fmt.Print(renderer.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckConfig(cmd *cobra.Command, args []string) error {
	report := validationSvc.RunConfig(projectRoot)

	fmt.Print(renderer.RenderValidation(&report))

	if !report.Ok() {
		os.Exit(1)
	}
	return nil
}

func runCheckAll(cmd *cobra.Command, args []string) error {
	report := validationSvc.RunAll(projectRoot, flagStrict)

	fmt.Print(renderer.RenderUnifiedValidation(&report))

	if report.Summary.Failed > 0 {
		os.Exit(1)
	}
	return nil
}
