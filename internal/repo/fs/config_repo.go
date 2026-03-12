package fs

import (
	"os"
	"path/filepath"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/pelletier/go-toml/v2"
)

// ConfigRepo implements repo.ConfigRepo using the filesystem.
type ConfigRepo struct {
	projectRoot string
}

// NewConfigRepo creates a ConfigRepo.
func NewConfigRepo(projectRoot string) *ConfigRepo {
	return &ConfigRepo{projectRoot: projectRoot}
}

// ReadProjectConfig parses mind.toml.
func (r *ConfigRepo) ReadProjectConfig() (*domain.Config, error) {
	path := filepath.Join(r.projectRoot, "mind.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// WriteProjectConfig writes mind.toml.
func (r *ConfigRepo) WriteProjectConfig(cfg *domain.Config) error {
	path := filepath.Join(r.projectRoot, "mind.toml")
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
