package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	flagInitName         string
	flagInitWithGitHub   bool
	flagInitFromExisting bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Mind Framework project",
	Long:  "Creates .mind/, docs/ zone structure, mind.toml, and adapter files.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVarP(&flagInitName, "name", "n", "", "Project name (default: directory name)")
	initCmd.Flags().BoolVar(&flagInitWithGitHub, "with-github", false, "Create .github/agents/ adapter")
	initCmd.Flags().BoolVar(&flagInitFromExisting, "from-existing", false, "Detect and preserve existing docs/")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	root := flagProject
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
	} else {
		var err error
		root, err = filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("resolve root: %w", err)
		}
	}

	svc := service.NewInitService()
	result, err := svc.Init(root, flagInitName, flagInitWithGitHub, flagInitFromExisting)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyInitialized) {
			return exitRuntime(fmt.Errorf("project already initialized (.mind/ exists)"))
		}
		return exitRuntime(err)
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderInitResult(result))
	return nil
}
