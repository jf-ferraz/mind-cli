package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// TestConfigRepo_GraphParsing verifies that [[graph]] TOML entries are parsed into Config.Graph.
func TestConfigRepo_GraphParsing(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v1.0"
generation = 1

[project]
name = "test-project"
type = "cli"

[documents.spec.requirements]
id = "doc:spec/requirements"
path = "docs/spec/requirements.md"
zone = "spec"
status = "active"

[documents.spec.architecture]
id = "doc:spec/architecture"
path = "docs/spec/architecture.md"
zone = "spec"
status = "active"

[[graph]]
from = "doc:spec/requirements"
to   = "doc:spec/architecture"
type = "informs"

[[graph]]
from = "doc:spec/architecture"
to   = "doc:spec/requirements"
type = "validates"
`

	if err := os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644); err != nil {
		t.Fatal(err)
	}

	repo := NewConfigRepo(dir)
	cfg, err := repo.ReadProjectConfig()
	if err != nil {
		t.Fatalf("ReadProjectConfig: %v", err)
	}

	if len(cfg.Graph) != 2 {
		t.Fatalf("expected 2 graph edges, got %d", len(cfg.Graph))
	}

	edge0 := cfg.Graph[0]
	if edge0.From != "doc:spec/requirements" {
		t.Errorf("edge[0].From = %q", edge0.From)
	}
	if edge0.To != "doc:spec/architecture" {
		t.Errorf("edge[0].To = %q", edge0.To)
	}
	if edge0.Type != domain.EdgeInforms {
		t.Errorf("edge[0].Type = %q", edge0.Type)
	}

	edge1 := cfg.Graph[1]
	if edge1.From != "doc:spec/architecture" {
		t.Errorf("edge[1].From = %q", edge1.From)
	}
	if edge1.To != "doc:spec/requirements" {
		t.Errorf("edge[1].To = %q", edge1.To)
	}
	if edge1.Type != domain.EdgeValidates {
		t.Errorf("edge[1].Type = %q", edge1.Type)
	}
}

// TestConfigRepo_NoGraph verifies that configs without [[graph]] parse with empty Graph.
func TestConfigRepo_NoGraph(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[manifest]
schema = "mind/v1.0"
generation = 1

[project]
name = "test-project"
type = "cli"
`

	if err := os.WriteFile(filepath.Join(dir, "mind.toml"), []byte(tomlContent), 0644); err != nil {
		t.Fatal(err)
	}

	repo := NewConfigRepo(dir)
	cfg, err := repo.ReadProjectConfig()
	if err != nil {
		t.Fatalf("ReadProjectConfig: %v", err)
	}

	if len(cfg.Graph) != 0 {
		t.Errorf("expected 0 graph edges, got %d", len(cfg.Graph))
	}
}
