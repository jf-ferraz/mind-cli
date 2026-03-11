package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

var (
	schemaRe  = regexp.MustCompile(`^mind/v\d+\.\d+$`)
	kebabRe   = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	docIDRe   = regexp.MustCompile(`^doc:[a-z]+/[a-z][a-z0-9-]*$`)
	docPathRe = regexp.MustCompile(`^docs/.*\.md$`)
)

var validProjectTypes = map[string]bool{
	"cli": true, "api": true, "library": true, "webapp": true, "service": true,
}

// ConfigSuite returns the config validation suite for mind.toml.
func ConfigSuite() *Suite {
	return &Suite{
		Name: "config",
		Checks: []Check{
			{1, "mind.toml exists and is valid TOML", domain.LevelFail, checkTomlExists},
			{2, "manifest.schema format", domain.LevelFail, checkSchemaFormat},
			{3, "manifest.generation >= 1", domain.LevelFail, checkGeneration},
			{4, "project.name is kebab-case", domain.LevelFail, checkProjectName},
			{5, "project.type is valid", domain.LevelFail, checkProjectType},
			{6, "document IDs are valid", domain.LevelFail, checkDocIDs},
			{7, "document paths are valid", domain.LevelFail, checkDocPaths},
			{8, "document zones are valid", domain.LevelFail, checkDocZones},
			{9, "document statuses are valid", domain.LevelFail, checkDocStatuses},
			{10, "governance.max-retries in range", domain.LevelWarn, checkMaxRetries},
		},
	}
}

func checkTomlExists(ctx *CheckContext) (bool, string) {
	if ctx.ConfigRepo == nil {
		return false, "mind.toml not found"
	}
	_, err := ctx.ConfigRepo.ReadProjectConfig()
	if err != nil {
		return false, fmt.Sprintf("mind.toml parse error: %v", err)
	}
	return true, ""
}

func checkSchemaFormat(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	if cfg.Manifest.Schema == "" {
		return false, "manifest.schema is empty"
	}
	if !schemaRe.MatchString(cfg.Manifest.Schema) {
		return false, fmt.Sprintf("manifest.schema %q does not match mind/vN.N", cfg.Manifest.Schema)
	}
	return true, ""
}

func checkGeneration(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	if cfg.Manifest.Generation < 1 {
		return false, fmt.Sprintf("manifest.generation is %d, must be >= 1", cfg.Manifest.Generation)
	}
	return true, ""
}

func checkProjectName(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	if cfg.Project.Name == "" {
		return false, "project.name is empty"
	}
	if !kebabRe.MatchString(cfg.Project.Name) {
		return false, fmt.Sprintf("project.name %q is not kebab-case", cfg.Project.Name)
	}
	return true, ""
}

func checkProjectType(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	if cfg.Project.Type == "" {
		return true, ""
	}
	if !validProjectTypes[cfg.Project.Type] {
		valid := make([]string, 0, len(validProjectTypes))
		for k := range validProjectTypes {
			valid = append(valid, k)
		}
		return false, fmt.Sprintf("project.type %q is not valid (expected: %s)", cfg.Project.Type, strings.Join(valid, ", "))
	}
	return true, ""
}

func checkDocIDs(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	var bad []string
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.ID != "" && !docIDRe.MatchString(entry.ID) {
				bad = append(bad, entry.ID)
			}
		}
	}
	if len(bad) > 0 {
		return false, "Invalid doc IDs: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkDocPaths(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	var bad []string
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.Path != "" && !docPathRe.MatchString(entry.Path) {
				bad = append(bad, entry.Path)
			}
		}
	}
	if len(bad) > 0 {
		return false, "Invalid doc paths: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkDocZones(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	var bad []string
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.Zone != "" && !domain.ValidZone(entry.Zone) {
				bad = append(bad, entry.Zone)
			}
		}
	}
	if len(bad) > 0 {
		return false, "Invalid zones: " + strings.Join(bad, ", ")
	}
	return true, ""
}

var validDocStatuses = map[string]bool{
	"draft": true, "active": true, "complete": true,
}

func checkDocStatuses(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	var bad []string
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.Status != "" && !validDocStatuses[entry.Status] {
				bad = append(bad, entry.Status)
			}
		}
	}
	if len(bad) > 0 {
		return false, "Invalid statuses: " + strings.Join(bad, ", ")
	}
	return true, ""
}

func checkMaxRetries(ctx *CheckContext) (bool, string) {
	cfg := getConfig(ctx)
	if cfg == nil {
		return true, ""
	}
	if cfg.Governance.MaxRetries < 0 || cfg.Governance.MaxRetries > 5 {
		return false, fmt.Sprintf("governance.max-retries is %d, must be 0-5", cfg.Governance.MaxRetries)
	}
	return true, ""
}

func getConfig(ctx *CheckContext) *domain.Config {
	if ctx.ConfigRepo == nil {
		return nil
	}
	cfg, err := ctx.ConfigRepo.ReadProjectConfig()
	if err != nil {
		return nil
	}
	return cfg
}
