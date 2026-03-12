package cmd

import (
	"fmt"
	"runtime"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/render"
	"github.com/spf13/cobra"
)

// Set via ldflags at build time.
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var flagShort bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run:   runVersion,
}

func init() {
	versionCmd.Flags().BoolVar(&flagShort, "short", false, "Print version string only")
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	if flagShort {
		fmt.Println(Version)
		return
	}

	info := &domain.VersionInfo{
		Version:   Version,
		Commit:    CommitSHA,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	mode := render.DetectMode(flagJSON, flagNoColor)
	r := render.New(mode, render.TermWidth())
	fmt.Print(r.RenderVersionInfo(info))
}
