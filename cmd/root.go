package cmd

import (
	"fmt"
	"os"

	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
	"github.com/jf-ferraz/mind-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagNoColor bool
	flagProject string
)

// Deps is a type alias for the shared deps.Deps struct.
// Kept here for backward compatibility with existing CLI code.
type Deps = deps.Deps

// BuildDeps delegates to deps.Build for constructing all repositories and services.
func BuildDeps(root string, r *render.Renderer) *Deps {
	return deps.Build(root, r)
}

// Package-level variables for CLI command handlers (populated from Deps).
var (
	projectRoot   string
	renderer      *render.Renderer
	reconcileSvc  *service.ReconciliationService
	validationSvc *service.ValidationService
	doctorSvc     *service.DoctorService
	projectSvc    *service.ProjectService
	workflowSvc   *service.WorkflowService
	generateSvc   *service.GenerateService
	docRepo       *fs.DocRepo
	iterRepo      *fs.IterationRepo
	briefRepo     *fs.BriefRepo
)

var rootCmd = &cobra.Command{
	Use:   "mind",
	Short: "Mind Framework CLI — project intelligence at your fingertips",
	Long: `mind is the command-line companion for the Mind Agent Framework.
It inspects documentation structure, validates quality gates,
manages iterations, and bridges AI agent workflows.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip wiring for commands that do not require a project
		if !requiresProject(cmd) {
			return nil
		}

		root, err := resolveRoot()
		if err != nil {
			if isNotProject(err) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(3)
			}
			return err
		}

		mode := render.DetectMode(flagJSON, flagNoColor)
		r := render.New(mode, render.TermWidth())
		deps := BuildDeps(root, r)

		// Populate package-level variables for CLI command handlers
		projectRoot = deps.ProjectRoot
		renderer = deps.Renderer
		docRepo = deps.DocRepo
		iterRepo = deps.IterRepo
		briefRepo = deps.BriefRepo
		reconcileSvc = deps.ReconcileSvc
		validationSvc = deps.ValidationSvc
		doctorSvc = deps.DoctorSvc
		projectSvc = deps.ProjectSvc
		workflowSvc = deps.WorkflowSvc
		generateSvc = deps.GenerateSvc

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagJSON, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "Path to project root (default: auto-detect)")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// requiresProject returns true if the command needs project wiring.
// Commands like version, help, init, completion, and tui do not require
// standard CLI wiring (tui handles its own wiring via BuildDeps).
func requiresProject(cmd *cobra.Command) bool {
	name := cmd.Name()

	switch name {
	case "version", "help", "init", "completion", "tui",
		"mind": // root command itself
		return false
	}

	// Check parent commands (e.g., "mind help status" should not require project)
	if cmd.Parent() != nil && cmd.Parent().Name() == "help" {
		return false
	}

	return true
}
