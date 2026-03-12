package cmd

import (
	"errors"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/fs"
)

// resolveRoot finds the project root from --project flag or auto-detection.
func resolveRoot() (string, error) {
	if flagProject != "" {
		return fs.FindProjectRootFrom(flagProject)
	}
	return fs.FindProjectRoot()
}

// isNotProject returns true if the error indicates no project was found.
func isNotProject(err error) bool {
	return errors.Is(err, domain.ErrNotProject)
}
