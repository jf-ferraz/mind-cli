package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/framework"
	"github.com/spf13/cobra"
)

var (
	flagFrameworkSource string
	flagFrameworkForce  bool
)

var frameworkCmd = &cobra.Command{
	Use:   "framework",
	Short: "Manage the mind framework installation",
	Long:  "Install, inspect, and compare the canonical mind framework artifacts.",
}

var frameworkInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install framework artifacts to global config",
	RunE:  runFrameworkInstall,
}

var frameworkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed framework version and drift status",
	RunE:  runFrameworkStatus,
}

var frameworkDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare project .mind/ against global framework",
	RunE:  runFrameworkDiff,
}

func init() {
	frameworkInstallCmd.Flags().StringVarP(&flagFrameworkSource, "source", "s", "", "Framework source (local path to .mind/ directory)")
	frameworkInstallCmd.Flags().BoolVar(&flagFrameworkForce, "force", false, "Overwrite existing installation")

	frameworkCmd.AddCommand(frameworkInstallCmd)
	frameworkCmd.AddCommand(frameworkStatusCmd)
	frameworkCmd.AddCommand(frameworkDiffCmd)
	rootCmd.AddCommand(frameworkCmd)
}

func runFrameworkInstall(cmd *cobra.Command, args []string) error {
	source := flagFrameworkSource
	if source == "" {
		source = filepath.Join(projectRoot, ".mind")
	}

	result, err := framework.Install(source, "", flagFrameworkForce)
	if err != nil {
		return exitConfig(err)
	}

	if flagJSON {
		out, merr := json.MarshalIndent(result, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal install result: %w", merr))
		}
		fmt.Println(string(out))
	} else {
		action := "installed"
		if result.Overwritten {
			action = "reinstalled"
		}
		fmt.Printf("Framework v%s %s (%d artifacts from %s)\n", result.Version, action, result.ArtifactCount, result.Source)
	}
	return nil
}

func runFrameworkStatus(cmd *cobra.Command, args []string) error {
	cfg, err := configRepo.ReadProjectConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read project config: %v\n", err)
	}
	var projFW *domain.FrameworkConfig
	if cfg != nil {
		projFW = cfg.Framework
	}

	result, err := framework.Status("", projFW)
	if err != nil {
		return exitRuntime(err)
	}

	if flagJSON {
		out, merr := json.MarshalIndent(result, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal status result: %w", merr))
		}
		fmt.Println(string(out))
	} else {
		if !result.Installed {
			fmt.Println("Framework: not installed")
			fmt.Println("Run: mind framework install --source <path-to-.mind/>")
			return nil
		}
		fmt.Printf("Framework: v%s (%s)\n", result.Version, result.Mode)
		fmt.Printf("Source:    %s\n", result.Source)
		fmt.Printf("Installed: %s\n", result.InstalledAt)
		if len(result.DriftFiles) == 0 {
			fmt.Println("Drift:     none")
		} else {
			fmt.Printf("Drift:     %d file(s) modified\n", len(result.DriftFiles))
			for _, f := range result.DriftFiles {
				fmt.Printf("  - %s\n", f)
			}
		}
	}
	return nil
}

func runFrameworkDiff(cmd *cobra.Command, args []string) error {
	projectMindDir := filepath.Join(projectRoot, ".mind")

	result, err := framework.Diff(projectMindDir, "")
	if err != nil {
		return exitRuntime(err)
	}

	if flagJSON {
		out, merr := json.MarshalIndent(result, "", "  ")
		if merr != nil {
			return exitRuntime(fmt.Errorf("marshal diff result: %w", merr))
		}
		fmt.Println(string(out))
	} else {
		if !result.HasDiff {
			fmt.Println("No differences found.")
			return nil
		}
		fmt.Printf("%d difference(s) found:\n", len(result.Entries))
		for _, e := range result.Entries {
			symbol := "?"
			switch e.Status {
			case "modified":
				symbol = "M"
			case "missing":
				symbol = "D"
			case "extra":
				symbol = "A"
			}
			fmt.Printf("  %s %s\n", symbol, e.Path)
		}
	}

	if result.HasDiff {
		return exitQuiet(1)
	}
	return nil
}
