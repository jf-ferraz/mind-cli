package cmd

import "github.com/jf-ferraz/mind-cli/internal/repo/fs"

// resolveRoot finds the project root from --project flag or auto-detection.
func resolveRoot() (string, error) {
	if flagProject != "" {
		return fs.FindProjectRootFrom(flagProject)
	}
	return fs.FindProjectRoot()
}
