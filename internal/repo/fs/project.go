package fs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FindProjectRoot walks up from the current directory looking for .mind/.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".mind")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", domain.ErrNotProject
		}
		dir = parent
	}
}

// FindProjectRootFrom walks up from a given directory looking for .mind/.
func FindProjectRootFrom(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".mind")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", domain.ErrNotProject
		}
		dir = parent
	}
}

// DetectProject builds a Project from the filesystem.
func DetectProject(root string) (*domain.Project, error) {
	if _, err := os.Stat(filepath.Join(root, ".mind")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotProject
		}
		return nil, err
	}

	p := &domain.Project{
		Root:     root,
		DocsRoot: filepath.Join(root, "docs"),
		MindRoot: filepath.Join(root, ".mind"),
	}

	// Try to read project name from mind.toml
	confRepo := NewConfigRepo(root)
	if cfg, err := confRepo.ReadProjectConfig(); err == nil {
		p.Name = cfg.Project.Name
		p.Config = cfg
	}

	return p, nil
}
