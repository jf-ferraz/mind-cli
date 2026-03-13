package globalconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// RegistryEntry represents a single project in projects.toml.
type RegistryEntry struct {
	Path string `toml:"path" json:"path"`
}

// Registry is the top-level structure of projects.toml.
type Registry struct {
	Projects map[string]RegistryEntry `toml:"projects" json:"projects"`
}

var kebabCasePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// RegistryPath returns the path to projects.toml.
func (r *GlobalConfigRepo) RegistryPath() string {
	return filepath.Join(r.dir, "projects.toml")
}

// ReadRegistry loads projects.toml. Returns empty registry if file doesn't exist.
func (r *GlobalConfigRepo) ReadRegistry() (*Registry, error) {
	path := r.RegistryPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Registry{Projects: make(map[string]RegistryEntry)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading projects.toml: %w", err)
	}
	var reg Registry
	if err := toml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing projects.toml: %w", err)
	}
	if reg.Projects == nil {
		reg.Projects = make(map[string]RegistryEntry)
	}
	return &reg, nil
}

// WriteRegistry saves projects.toml atomically.
func (r *GlobalConfigRepo) WriteRegistry(reg *Registry) error {
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := toml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshaling projects.toml: %w", err)
	}
	path := r.RegistryPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing projects.toml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming projects.toml: %w", err)
	}
	return nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// RegistryListItem represents a project entry for display.
type RegistryListItem struct {
	Alias string `json:"alias"`
	Path  string `json:"path"`
}

// RegistryList returns all registered projects sorted by alias.
func (r *GlobalConfigRepo) RegistryList() ([]RegistryListItem, error) {
	reg, err := r.ReadRegistry()
	if err != nil {
		return nil, err
	}
	items := make([]RegistryListItem, 0, len(reg.Projects))
	for alias, entry := range reg.Projects {
		items = append(items, RegistryListItem{Alias: alias, Path: entry.Path})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Alias < items[j].Alias
	})
	return items, nil
}

// RegistryAdd adds a project with the given alias and path.
func (r *GlobalConfigRepo) RegistryAdd(alias, path string) error {
	if !kebabCasePattern.MatchString(alias) {
		return fmt.Errorf("invalid alias %q: must be kebab-case (e.g. my-project)", alias)
	}
	expanded := expandPath(path)
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}
	reg, err := r.ReadRegistry()
	if err != nil {
		return err
	}
	if _, exists := reg.Projects[alias]; exists {
		return fmt.Errorf("alias %q already registered (use remove first)", alias)
	}
	reg.Projects[alias] = RegistryEntry{Path: absPath}
	return r.WriteRegistry(reg)
}

// RegistryRemove removes a project by alias.
func (r *GlobalConfigRepo) RegistryRemove(alias string) error {
	reg, err := r.ReadRegistry()
	if err != nil {
		return err
	}
	if _, exists := reg.Projects[alias]; !exists {
		return fmt.Errorf("alias %q not found in registry", alias)
	}
	delete(reg.Projects, alias)
	return r.WriteRegistry(reg)
}

// RegistryResolve resolves an @alias to its absolute path.
func (r *GlobalConfigRepo) RegistryResolve(aliasRef string) (string, error) {
	alias := strings.TrimPrefix(aliasRef, "@")
	reg, err := r.ReadRegistry()
	if err != nil {
		return "", err
	}
	entry, exists := reg.Projects[alias]
	if !exists {
		return "", fmt.Errorf("alias %q not found in registry", alias)
	}
	return expandPath(entry.Path), nil
}

// RegistryCheckResult describes the status of a single registry entry.
type RegistryCheckResult struct {
	Alias  string `json:"alias"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// RegistryCheck validates that all registered paths exist on disk.
func (r *GlobalConfigRepo) RegistryCheck() ([]RegistryCheckResult, error) {
	reg, err := r.ReadRegistry()
	if err != nil {
		return nil, err
	}
	results := make([]RegistryCheckResult, 0, len(reg.Projects))
	for alias, entry := range reg.Projects {
		expanded := expandPath(entry.Path)
		result := RegistryCheckResult{Alias: alias, Path: expanded}
		info, serr := os.Stat(expanded)
		if serr != nil {
			result.Exists = false
			result.Error = fmt.Sprintf("path does not exist: %s", expanded)
		} else if !info.IsDir() {
			result.Exists = false
			result.Error = fmt.Sprintf("path is not a directory: %s", expanded)
		} else {
			result.Exists = true
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Alias < results[j].Alias
	})
	return results, nil
}
