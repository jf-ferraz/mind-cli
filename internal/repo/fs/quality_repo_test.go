package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQualityRepo_ReadLog_FileNotExist(t *testing.T) {
	root := t.TempDir()
	repo := NewQualityRepo(root)

	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v, want nil for missing file", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0 for missing file", len(entries))
	}
}

func TestQualityRepo_ReadLog_EmptyFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "quality-log.yml"), []byte(""), 0o644)

	repo := NewQualityRepo(root)
	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v, want nil for empty file", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0 for empty file", len(entries))
	}
}

func TestQualityRepo_ReadLog_ValidEntries(t *testing.T) {
	root := t.TempDir()
	yaml := `- topic: auth-strategy
  variant: convergence-v1
  date: 2026-03-01T00:00:00Z
  score: 4.0
  gate_pass: true
  dimensions:
    - name: rigor
      value: 4
    - name: coverage
      value: 4
    - name: actionability
      value: 4
    - name: objectivity
      value: 4
    - name: convergence
      value: 4
    - name: depth
      value: 4
  personas:
    - moderator
    - analyst
  output_path: docs/knowledge/auth.md
- topic: reconciliation
  variant: convergence-v1
  date: 2026-03-05T00:00:00Z
  score: 4.33
  gate_pass: true
  dimensions:
    - name: rigor
      value: 5
    - name: coverage
      value: 4
    - name: actionability
      value: 4
    - name: objectivity
      value: 5
    - name: convergence
      value: 4
    - name: depth
      value: 4
  personas:
    - moderator
    - analyst
    - architect
  output_path: docs/knowledge/reconciliation.md
`
	os.WriteFile(filepath.Join(root, "quality-log.yml"), []byte(yaml), 0o644)

	repo := NewQualityRepo(root)
	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}

	// Entries should be sorted by date
	if entries[0].Topic != "auth-strategy" {
		t.Errorf("entries[0].Topic = %q, want auth-strategy (earlier date)", entries[0].Topic)
	}
	if entries[1].Topic != "reconciliation" {
		t.Errorf("entries[1].Topic = %q, want reconciliation (later date)", entries[1].Topic)
	}

	// Verify dimension count
	if len(entries[0].Dimensions) != 6 {
		t.Errorf("entries[0].Dimensions = %d, want 6", len(entries[0].Dimensions))
	}
}

func TestQualityRepo_ReadLog_DocsKnowledgePath(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge")
	os.MkdirAll(knowledgeDir, 0o755)

	yaml := `- topic: test
  date: 2026-03-01T00:00:00Z
  score: 3.0
  gate_pass: true
  dimensions:
    - {name: rigor, value: 3}
    - {name: coverage, value: 3}
    - {name: actionability, value: 3}
    - {name: objectivity, value: 3}
    - {name: convergence, value: 3}
    - {name: depth, value: 3}
`
	os.WriteFile(filepath.Join(knowledgeDir, "quality-log.yml"), []byte(yaml), 0o644)

	repo := NewQualityRepo(root)
	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("entries = %d, want 1 (from docs/knowledge/quality-log.yml)", len(entries))
	}
}

func TestQualityRepo_ReadLog_RootTakesPrecedence(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge")
	os.MkdirAll(knowledgeDir, 0o755)

	rootYaml := `- topic: root-entry
  date: 2026-03-01T00:00:00Z
  score: 4.0
  gate_pass: true
  dimensions:
    - {name: rigor, value: 4}
    - {name: coverage, value: 4}
    - {name: actionability, value: 4}
    - {name: objectivity, value: 4}
    - {name: convergence, value: 4}
    - {name: depth, value: 4}
`
	knowledgeYaml := `- topic: knowledge-entry
  date: 2026-03-01T00:00:00Z
  score: 3.0
  gate_pass: true
  dimensions:
    - {name: rigor, value: 3}
    - {name: coverage, value: 3}
    - {name: actionability, value: 3}
    - {name: objectivity, value: 3}
    - {name: convergence, value: 3}
    - {name: depth, value: 3}
`
	os.WriteFile(filepath.Join(root, "quality-log.yml"), []byte(rootYaml), 0o644)
	os.WriteFile(filepath.Join(knowledgeDir, "quality-log.yml"), []byte(knowledgeYaml), 0o644)

	repo := NewQualityRepo(root)
	entries, err := repo.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Topic != "root-entry" {
		t.Errorf("entries[0].Topic = %q, want root-entry (root path takes precedence)", entries[0].Topic)
	}
}

func TestQualityRepo_ReadLog_InvalidYAML(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "quality-log.yml"), []byte("invalid: [yaml:"), 0o644)

	repo := NewQualityRepo(root)
	_, err := repo.ReadLog()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
