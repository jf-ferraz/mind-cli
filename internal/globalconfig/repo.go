package globalconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// GlobalConfig represents the ~/.config/mind/config.toml contents.
type GlobalConfig struct {
	Editor      string `toml:"editor,omitempty" json:"editor,omitempty"`
	DefaultMode string `toml:"default_mode,omitempty" json:"default_mode,omitempty"`
	LogLevel    string `toml:"log_level,omitempty" json:"log_level,omitempty"`
}

// DefaultGlobalConfig returns a config with sensible defaults.
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		DefaultMode: "standalone",
		LogLevel:    "info",
	}
}

var validLogLevels = map[string]bool{
	"debug": true, "info": true, "warn": true, "error": true,
}

var validModes = map[string]bool{
	"standalone": true, "thin": true,
}

// Validate checks that config fields contain valid values.
func (c *GlobalConfig) Validate() error {
	if c.LogLevel != "" && !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level %q: must be debug, info, warn, or error", c.LogLevel)
	}
	if c.DefaultMode != "" && !validModes[c.DefaultMode] {
		return fmt.Errorf("invalid default_mode %q: must be standalone or thin", c.DefaultMode)
	}
	return nil
}

// GlobalDir returns the global mind config directory, respecting XDG_CONFIG_HOME.
func GlobalDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mind")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mind")
}

// GlobalConfigRepo manages reading/writing ~/.config/mind/config.toml.
type GlobalConfigRepo struct {
	dir string
}

// NewGlobalConfigRepo creates a repo rooted at the given directory.
func NewGlobalConfigRepo(dir string) *GlobalConfigRepo {
	if dir == "" {
		dir = GlobalDir()
	}
	return &GlobalConfigRepo{dir: dir}
}

// Dir returns the directory managed by this repo.
func (r *GlobalConfigRepo) Dir() string { return r.dir }

// ConfigPath returns the full path to config.toml.
func (r *GlobalConfigRepo) ConfigPath() string {
	return filepath.Join(r.dir, "config.toml")
}

// Read loads the config from disk. If the file doesn't exist, returns defaults.
func (r *GlobalConfigRepo) Read() (*GlobalConfig, error) {
	path := r.ConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config.toml: %w", err)
	}
	cfg := DefaultGlobalConfig()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config.toml: %w", err)
	}
	return cfg, nil
}

// Write saves the config to disk atomically.
func (r *GlobalConfigRepo) Write(cfg *GlobalConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config.toml: %w", err)
	}
	path := r.ConfigPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing config.toml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming config.toml: %w", err)
	}
	return nil
}

// EnsureExists creates the config dir and a default config.toml if missing.
func (r *GlobalConfigRepo) EnsureExists() error {
	path := r.ConfigPath()
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return r.Write(DefaultGlobalConfig())
}
