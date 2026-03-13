package globalconfig

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// ShowConfig returns the current global config as TOML text.
func (r *GlobalConfigRepo) ShowConfig() (string, error) {
	cfg, err := r.Read()
	if err != nil {
		return "", err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshaling config: %w", err)
	}
	return string(data), nil
}

// EditorCommand returns the editor to use for config editing.
// Priority: config editor field, then $EDITOR env, then "vi".
func (r *GlobalConfigRepo) EditorCommand() (string, error) {
	cfg, err := r.Read()
	if err != nil {
		return "", err
	}
	if cfg.Editor != "" {
		return cfg.Editor, nil
	}
	if env := os.Getenv("EDITOR"); env != "" {
		return env, nil
	}
	return "vi", nil
}

// ValidateConfig reads and validates the current config.
func (r *GlobalConfigRepo) ValidateConfig() error {
	cfg, err := r.Read()
	if err != nil {
		return err
	}
	return cfg.Validate()
}
