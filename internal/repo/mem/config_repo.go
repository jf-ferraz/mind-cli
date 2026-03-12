package mem

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
)

// ConfigRepo is an in-memory implementation of repo.ConfigRepo for testing.
type ConfigRepo struct {
	Config *domain.Config
}

// NewConfigRepo creates an in-memory ConfigRepo.
func NewConfigRepo() *ConfigRepo {
	return &ConfigRepo{}
}

// ReadProjectConfig returns the stored config.
func (r *ConfigRepo) ReadProjectConfig() (*domain.Config, error) {
	if r.Config == nil {
		return nil, fmt.Errorf("mind.toml not found")
	}
	return r.Config, nil
}

// WriteProjectConfig stores the config.
func (r *ConfigRepo) WriteProjectConfig(cfg *domain.Config) error {
	r.Config = cfg
	return nil
}
