package cmd

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
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
	root, err := resolveRoot()
	if err != nil {
		return err
	}

	project, err := fs.DetectProject(root)
	if err != nil {
		return err
	}

	docRepo := fs.NewDocRepo(root)
	iterRepo := fs.NewIterationRepo(root)
	briefRepo := fs.NewBriefRepo(docRepo)

	health := &domain.ProjectHealth{
		Project: *project,
		Zones:   make(map[domain.Zone]domain.ZoneHealth),
	}

	// Brief status
	if brief, err := briefRepo.ParseBrief(); err == nil {
		health.Brief = *brief
	}

	// Zone health
	for _, zone := range domain.AllZones {
		docs, err := docRepo.ListByZone(zone)
		if err != nil {
			continue
		}
		zh := domain.ZoneHealth{Zone: zone, Total: len(docs)}
		for _, doc := range docs {
			if doc.IsStub {
				zh.Stubs++
			} else {
				zh.Complete++
			}
			zh.Present++
		}
		health.Zones[zone] = zh
	}

	// Last iteration
	iterations, err := iterRepo.List()
	if err == nil && len(iterations) > 0 {
		health.LastIteration = &iterations[0]
	}

	// Warnings
	if !health.Brief.Exists {
		health.Warnings = append(health.Warnings, "Project brief missing — run /discover or create docs/spec/project-brief.md")
	} else if health.Brief.IsStub {
		health.Warnings = append(health.Warnings, "Project brief is a stub — fill in Vision, Key Deliverables, and Scope")
	}

	for zone, zh := range health.Zones {
		if zh.Stubs > 0 {
			health.Warnings = append(health.Warnings, fmt.Sprintf("%s/ has %d stub file(s)", zone, zh.Stubs))
		}
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderHealth(health))
	return nil
}
