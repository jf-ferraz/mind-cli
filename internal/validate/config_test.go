package validate

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// TestConfigSuiteStructure verifies FR-41: config validation suite has 12 checks (10 Phase 1 + 2 graph).
func TestConfigSuiteStructure(t *testing.T) {
	suite := ConfigSuite()
	if suite.Name != "config" {
		t.Errorf("Suite.Name = %q, want config", suite.Name)
	}
	if len(suite.Checks) != 12 {
		t.Errorf("ConfigSuite has %d checks, want 12", len(suite.Checks))
	}
}

// TestConfigSuiteAllPass verifies all checks pass for a well-formed config.
func TestConfigSuiteAllPass(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Manifest: domain.Manifest{
			Schema:     "mind/v1.0",
			Generation: 1,
		},
		Project: domain.ProjectMeta{
			Name: "my-project",
			Type: "cli",
		},
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"requirements": {
					ID:     "doc:spec/requirements",
					Path:   "docs/spec/requirements.md",
					Zone:   "spec",
					Status: "active",
				},
			},
		},
		Governance: domain.Governance{
			MaxRetries: 2,
		},
	}

	ctx := &CheckContext{ConfigRepo: configRepo}
	suite := ConfigSuite()
	report := suite.Run(ctx)

	if report.Failed > 0 {
		t.Errorf("Well-formed config should have 0 failures, got %d", report.Failed)
		for _, cr := range report.Checks {
			if !cr.Passed {
				t.Errorf("  FAIL: [%d] %s: %s", cr.ID, cr.Name, cr.Message)
			}
		}
	}
}

// TestCheckSchemaFormat verifies BR-11: schema version must match ^mind/v\d+\.\d+$.
func TestCheckSchemaFormat(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		wantPass bool
	}{
		{name: "valid mind/v1.0", schema: "mind/v1.0", wantPass: true},
		{name: "valid mind/v2.3", schema: "mind/v2.3", wantPass: true},
		{name: "valid mind/v10.99", schema: "mind/v10.99", wantPass: true},
		{name: "empty schema", schema: "", wantPass: false},
		{name: "no prefix", schema: "v1.0", wantPass: false},
		{name: "wrong prefix", schema: "mind/1.0", wantPass: false},
		{name: "no minor", schema: "mind/v1", wantPass: false},
		{name: "extra parts", schema: "mind/v1.0.0", wantPass: false},
		{name: "garbage", schema: "garbage", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: tt.schema, Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkSchemaFormat(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkSchemaFormat(schema=%q) = %v, want %v", tt.schema, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckGeneration verifies BR-12: generation must be >= 1.
func TestCheckGeneration(t *testing.T) {
	tests := []struct {
		name       string
		generation int
		wantPass   bool
	}{
		{name: "zero fails", generation: 0, wantPass: false},
		{name: "negative fails", generation: -1, wantPass: false},
		{name: "one passes", generation: 1, wantPass: true},
		{name: "large passes", generation: 100, wantPass: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: tt.generation},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkGeneration(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkGeneration(gen=%d) = %v, want %v", tt.generation, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckProjectName verifies BR-10 and FR-41: project name must be kebab-case.
func TestCheckProjectName(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		wantPass bool
	}{
		{name: "valid kebab", projName: "my-project", wantPass: true},
		{name: "single word", projName: "myproject", wantPass: true},
		{name: "with numbers", projName: "project123", wantPass: true},
		{name: "valid with hyphens", projName: "my-cool-project", wantPass: true},
		{name: "empty name", projName: "", wantPass: false},
		// FR-41 acceptance: uppercase should fail
		{name: "uppercase", projName: "MY-PROJECT", wantPass: false},
		{name: "mixed case", projName: "My-Project", wantPass: false},
		{name: "starts with number", projName: "1project", wantPass: false},
		{name: "starts with hyphen", projName: "-project", wantPass: false},
		{name: "has underscore", projName: "my_project", wantPass: false},
		{name: "has space", projName: "my project", wantPass: false},
		{name: "has dot", projName: "my.project", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: tt.projName, Type: "cli"},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkProjectName(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkProjectName(name=%q) = %v, want %v", tt.projName, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckProjectType verifies project type validation.
func TestCheckProjectType(t *testing.T) {
	tests := []struct {
		name     string
		projType string
		wantPass bool
	}{
		{name: "cli valid", projType: "cli", wantPass: true},
		{name: "api valid", projType: "api", wantPass: true},
		{name: "library valid", projType: "library", wantPass: true},
		{name: "webapp valid", projType: "webapp", wantPass: true},
		{name: "service valid", projType: "service", wantPass: true},
		{name: "empty passes (optional)", projType: "", wantPass: true},
		{name: "invalid type", projType: "microservice", wantPass: false},
		{name: "uppercase", projType: "CLI", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: tt.projType},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkProjectType(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkProjectType(type=%q) = %v, want %v", tt.projType, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckDocIDs verifies BR-14: doc ID format ^doc:[a-z]+/[a-z][a-z0-9-]*$.
func TestCheckDocIDs(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantPass bool
	}{
		{name: "valid id", id: "doc:spec/requirements", wantPass: true},
		{name: "valid with hyphen", id: "doc:spec/project-brief", wantPass: true},
		{name: "valid with number", id: "doc:spec/domain-model2", wantPass: true},
		{name: "empty is valid (optional)", id: "", wantPass: true},
		{name: "missing doc: prefix", id: "spec/requirements", wantPass: false},
		{name: "uppercase zone", id: "doc:Spec/requirements", wantPass: false},
		{name: "starts with number", id: "doc:spec/1requirements", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
				Documents: map[string]map[string]domain.DocEntry{
					"spec": {"test": {ID: tt.id, Path: "docs/spec/test.md", Zone: "spec", Status: "active"}},
				},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkDocIDs(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkDocIDs(id=%q) = %v, want %v", tt.id, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckDocPaths verifies BR-13: paths must start with docs/ and end with .md.
func TestCheckDocPaths(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPass bool
	}{
		{name: "valid path", path: "docs/spec/requirements.md", wantPass: true},
		{name: "nested path", path: "docs/spec/decisions/001-foo.md", wantPass: true},
		{name: "empty is valid (optional)", path: "", wantPass: true},
		{name: "missing docs/ prefix", path: "spec/requirements.md", wantPass: false},
		{name: "not .md suffix", path: "docs/spec/requirements.txt", wantPass: false},
		{name: "absolute path", path: "/docs/spec/requirements.md", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
				Documents: map[string]map[string]domain.DocEntry{
					"spec": {"test": {ID: "doc:spec/test", Path: tt.path, Zone: "spec", Status: "active"}},
				},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkDocPaths(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkDocPaths(path=%q) = %v, want %v", tt.path, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckDocZones verifies BR-15: zones must be valid.
func TestCheckDocZones(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		wantPass bool
	}{
		{name: "spec valid", zone: "spec", wantPass: true},
		{name: "blueprints valid", zone: "blueprints", wantPass: true},
		{name: "empty is valid (optional)", zone: "", wantPass: true},
		{name: "invalid zone", zone: "config", wantPass: false},
		{name: "uppercase", zone: "SPEC", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
				Documents: map[string]map[string]domain.DocEntry{
					"spec": {"test": {ID: "doc:spec/test", Path: "docs/spec/test.md", Zone: tt.zone, Status: "active"}},
				},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkDocZones(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkDocZones(zone=%q) = %v, want %v", tt.zone, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckDocStatuses verifies document status values.
func TestCheckDocStatuses(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		wantPass bool
	}{
		{name: "draft valid", status: "draft", wantPass: true},
		{name: "active valid", status: "active", wantPass: true},
		{name: "complete valid", status: "complete", wantPass: true},
		{name: "empty is valid (optional)", status: "", wantPass: true},
		{name: "stub is invalid in config", status: "stub", wantPass: false},
		{name: "unknown", status: "archived", wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
				Documents: map[string]map[string]domain.DocEntry{
					"spec": {"test": {ID: "doc:spec/test", Path: "docs/spec/test.md", Zone: "spec", Status: tt.status}},
				},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkDocStatuses(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkDocStatuses(status=%q) = %v, want %v", tt.status, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckMaxRetries verifies BR-22: governance.max-retries in range 0-5.
func TestCheckMaxRetries(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		wantPass   bool
	}{
		{name: "zero is valid", maxRetries: 0, wantPass: true},
		{name: "five is valid", maxRetries: 5, wantPass: true},
		{name: "three is valid", maxRetries: 3, wantPass: true},
		{name: "negative is invalid", maxRetries: -1, wantPass: false},
		{name: "six is invalid", maxRetries: 6, wantPass: false},
		{name: "large is invalid", maxRetries: 100, wantPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			configRepo.Config = &domain.Config{
				Manifest:   domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:    domain.ProjectMeta{Name: "test", Type: "cli"},
				Governance: domain.Governance{MaxRetries: tt.maxRetries},
			}
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkMaxRetries(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkMaxRetries(retries=%d) = %v, want %v", tt.maxRetries, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckGraphEdgeIDs verifies FR-84: graph edge IDs match doc ID format.
func TestCheckGraphEdgeIDs(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		wantPass bool
	}{
		{name: "valid edges", from: "doc:spec/requirements", to: "doc:spec/architecture", wantPass: true},
		{name: "invalid from", from: "bad-id", to: "doc:spec/architecture", wantPass: false},
		{name: "invalid to", from: "doc:spec/requirements", to: "not-a-doc-id", wantPass: false},
		{name: "no graph", from: "", to: "", wantPass: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			cfg := &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
			}
			if tt.from != "" || tt.to != "" {
				cfg.Graph = []domain.GraphEdge{
					{From: tt.from, To: tt.to, Type: domain.EdgeInforms},
				}
			}
			configRepo.Config = cfg
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, _ := checkGraphEdgeIDs(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkGraphEdgeIDs(from=%q, to=%q) = %v, want %v", tt.from, tt.to, passed, tt.wantPass)
			}
		})
	}
}

// TestCheckGraphEdgeTypes verifies FR-84: graph edge types must be valid.
func TestCheckGraphEdgeTypes(t *testing.T) {
	tests := []struct {
		name     string
		edgeType domain.EdgeType
		wantPass bool
	}{
		{name: "informs", edgeType: domain.EdgeInforms, wantPass: true},
		{name: "requires", edgeType: domain.EdgeRequires, wantPass: true},
		{name: "validates", edgeType: domain.EdgeValidates, wantPass: true},
		{name: "invalid type", edgeType: "depends", wantPass: false},
		{name: "empty type", edgeType: "", wantPass: true}, // empty = not set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := mem.NewConfigRepo()
			cfg := &domain.Config{
				Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
				Project:  domain.ProjectMeta{Name: "test", Type: "cli"},
			}
			if tt.edgeType != "" {
				cfg.Graph = []domain.GraphEdge{
					{From: "doc:spec/requirements", To: "doc:spec/architecture", Type: tt.edgeType},
				}
			}
			configRepo.Config = cfg
			ctx := &CheckContext{ConfigRepo: configRepo}
			passed, msg := checkGraphEdgeTypes(ctx)
			if passed != tt.wantPass {
				t.Errorf("checkGraphEdgeTypes(type=%q) = %v, want %v (msg: %s)", tt.edgeType, passed, tt.wantPass, msg)
			}
		})
	}
}

// TestCheckTomlExists verifies check 1: mind.toml existence.
func TestCheckTomlExists(t *testing.T) {
	t.Run("no config repo", func(t *testing.T) {
		ctx := &CheckContext{ConfigRepo: nil}
		passed, msg := checkTomlExists(ctx)
		if passed {
			t.Error("should fail when ConfigRepo is nil")
		}
		if !strings.Contains(msg, "not found") {
			t.Errorf("message = %q, should mention not found", msg)
		}
	})

	t.Run("valid config", func(t *testing.T) {
		configRepo := mem.NewConfigRepo()
		configRepo.Config = &domain.Config{
			Manifest: domain.Manifest{Schema: "mind/v1.0", Generation: 1},
			Project:  domain.ProjectMeta{Name: "test"},
		}
		ctx := &CheckContext{ConfigRepo: configRepo}
		passed, _ := checkTomlExists(ctx)
		if !passed {
			t.Error("should pass when config is valid")
		}
	})
}
