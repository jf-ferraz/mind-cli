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
	flagJSON    bool
	flagNoColor bool
	flagProject string
)

// Centralized wiring: package-level variables populated by PersistentPreRunE.
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
		projectRoot = root

		// Create renderer
		mode := render.DetectMode(flagJSON, flagNoColor)
		renderer = render.New(mode, render.TermWidth())

		// Create repositories
		docRepo = fs.NewDocRepo(root)
		iterRepo = fs.NewIterationRepo(root)
		stateRepo := fs.NewStateRepo(root)
		briefRepo = fs.NewBriefRepo(docRepo)
		configRepo := fs.NewConfigRepo(root)
		lockRepo := fs.NewLockRepo(root)

		// Create services
		reconcileSvc = service.NewReconciliationService(configRepo, docRepo, lockRepo)
		validationSvc = service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
		doctorSvc = service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo, lockRepo)
		projectSvc = service.NewProjectService(docRepo, iterRepo, stateRepo, briefRepo)
		workflowSvc = service.NewWorkflowService(stateRepo, iterRepo)
		generateSvc = service.NewGenerateService(root)

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
// Commands like version, help, init, and completion do not require a project.
func requiresProject(cmd *cobra.Command) bool {
	name := cmd.Name()

	// Commands that do not require a project
	switch name {
	case "version", "help", "init", "completion",
		"mind": // root command itself
		return false
	}

	// Check parent commands (e.g., "mind help status" should not require project)
	if cmd.Parent() != nil && cmd.Parent().Name() == "help" {
		return false
	}

	return true
}
