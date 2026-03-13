package cmd

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/framework"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project health and documentation status",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	project, err := projectSvc.DetectProject(projectRoot)
	if err != nil {
		return err
	}

	health, err := projectSvc.AssembleHealth(project)
	if err != nil {
		return err
	}

	// Staleness panel: read existing lock data (FR-77, does NOT trigger reconciliation)
	staleness, err := reconcileSvc.ReadStaleness(projectRoot)
	if err == nil && staleness != nil {
		health.Staleness = staleness
	}

	// Framework panel: only if [framework] section exists in mind.toml.
	cfg, cfgErr := configRepo.ReadProjectConfig()
	if cfgErr == nil && cfg != nil && cfg.Framework != nil {
		fwStatus, fwErr := framework.Status("", cfg.Framework)
		if fwErr == nil && fwStatus.Installed {
			health.Framework = &domain.FrameworkStatus{
				Mode:       string(fwStatus.Mode),
				Version:    fwStatus.Version,
				DriftCount: len(fwStatus.DriftFiles),
			}
		}
	}

	fmt.Print(renderer.RenderHealth(health))

	// Exit code 1 if issues found
	hasIssues := false
	if health.Brief.GateResult == domain.BriefMissing {
		hasIssues = true
	}
	for _, zh := range health.Zones {
		if zh.Stubs > 0 {
			hasIssues = true
		}
	}
	if hasIssues {
		return exitQuiet(1)
	}

	return nil
}
