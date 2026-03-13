package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/framework"
	"github.com/jf-ferraz/mind-cli/internal/generate"
)

// InitService handles project initialization.
type InitService struct{}

// NewInitService creates an InitService.
func NewInitService() *InitService {
	return &InitService{}
}

// Init creates a new Mind Framework project in the given directory.
func (s *InitService) Init(root, name string, withGitHub, fromExisting bool) (*domain.InitResult, error) {
	mindDir := filepath.Join(root, ".mind")
	if _, err := os.Stat(mindDir); err == nil {
		return nil, domain.ErrAlreadyInitialized
	}

	if name == "" {
		name = filepath.Base(root)
	}

	result := &domain.InitResult{
		ProjectName:  name,
		Root:         root,
		FromExisting: fromExisting,
	}

	// Directories to create
	dirs := []string{
		".mind",
		"docs/spec",
		"docs/spec/decisions",
		"docs/blueprints",
		"docs/state",
		"docs/iterations",
		"docs/knowledge",
		".claude",
	}

	for _, dir := range dirs {
		absDir := filepath.Join(root, dir)
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Stub documents with their content generators
	stubs := map[string]string{
		"docs/spec/project-brief.md": generate.StubBriefTemplate(),
		"docs/spec/requirements.md":  generate.StubDocument("Requirements"),
		"docs/spec/architecture.md":  generate.StubDocument("Architecture"),
		"docs/spec/domain-model.md":  generate.StubDocument("Domain Model"),
		"docs/state/current.md":      generate.CurrentStub(),
		"docs/state/workflow.md":     generate.WorkflowStub(),
		"docs/blueprints/INDEX.md":   generate.IndexStub(),
		"docs/knowledge/glossary.md": generate.GlossaryStub(),
	}

	for relPath, content := range stubs {
		absPath := filepath.Join(root, relPath)

		if fromExisting {
			if _, err := os.Stat(absPath); err == nil {
				result.ExistingPreserved = append(result.ExistingPreserved, relPath)
				continue
			}
		}

		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", relPath, err)
		}
		result.FilesCreated = append(result.FilesCreated, relPath)
	}

	// Detect installed framework version
	frameworkVersion := ""
	globalDir := framework.DefaultGlobalDir()
	lockPath := filepath.Join(globalDir, "framework.lock")
	if lock, err := framework.ReadLock(lockPath); err == nil {
		frameworkVersion = lock.Framework.Version
	}

	// mind.toml
	tomlPath := filepath.Join(root, "mind.toml")
	tomlContent := generate.MindTomlTemplate(name, frameworkVersion)
	if fromExisting {
		if _, err := os.Stat(tomlPath); err == nil {
			result.ExistingPreserved = append(result.ExistingPreserved, "mind.toml")
		} else {
			if err := os.WriteFile(tomlPath, []byte(tomlContent), 0644); err != nil {
				return nil, fmt.Errorf("write mind.toml: %w", err)
			}
			result.FilesCreated = append(result.FilesCreated, "mind.toml")
		}
	} else {
		if err := os.WriteFile(tomlPath, []byte(tomlContent), 0644); err != nil {
			return nil, fmt.Errorf("write mind.toml: %w", err)
		}
		result.FilesCreated = append(result.FilesCreated, "mind.toml")
	}

	// .claude/CLAUDE.md adapter
	claudePath := filepath.Join(root, ".claude", "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(generate.ClaudeAdapterTemplate()), 0644); err != nil {
		return nil, fmt.Errorf("write .claude/CLAUDE.md: %w", err)
	}
	result.FilesCreated = append(result.FilesCreated, ".claude/CLAUDE.md")

	// .github/agents/ if requested
	if withGitHub {
		agentsDir := filepath.Join(root, ".github", "agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return nil, fmt.Errorf("create .github/agents: %w", err)
		}
		gitkeep := filepath.Join(agentsDir, ".gitkeep")
		if err := os.WriteFile(gitkeep, []byte(""), 0644); err != nil {
			return nil, fmt.Errorf("write .gitkeep: %w", err)
		}
		result.FilesCreated = append(result.FilesCreated, ".github/agents/.gitkeep")
	}

	return result, nil
}
