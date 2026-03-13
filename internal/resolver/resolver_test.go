package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDirs creates temporary project and global dirs with test artifacts.
func setupTestDirs(t *testing.T) (projectDir, globalDir string) {
	t.Helper()
	projectDir = t.TempDir()
	globalDir = t.TempDir()
	return projectDir, globalDir
}

// writeArtifact writes a file at dir/kind/name with the given content.
func writeArtifact(t *testing.T, dir string, kind ArtifactKind, name, content string) {
	t.Helper()
	kindDir := filepath.Join(dir, string(kind))
	if err := os.MkdirAll(kindDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kindDir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// --- Resolve tests ---

func TestResolve_ProjectOnly_ReturnsProjectFile(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "explorer.md", "project agent content")

	r := New(projectDir, globalDir)
	result, err := r.Resolve(KindAgents, "explorer.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != SourceProject {
		t.Errorf("expected source %q, got %q", SourceProject, result.Source)
	}
	if result.Kind != KindAgents {
		t.Errorf("expected kind %q, got %q", KindAgents, result.Kind)
	}
	if result.Name != "explorer.md" {
		t.Errorf("expected name %q, got %q", "explorer.md", result.Name)
	}
	if result.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestResolve_GlobalOnly_ReturnsGlobalFile(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "explorer.md", "global agent content")

	r := New(projectDir, globalDir)
	result, err := r.Resolve(KindAgents, "explorer.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != SourceGlobal {
		t.Errorf("expected source %q, got %q", SourceGlobal, result.Source)
	}
}

func TestResolve_ProjectOverride_ReturnsProjectFile(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "explorer.md", "project override")
	writeArtifact(t, globalDir, KindAgents, "explorer.md", "global original")

	r := New(projectDir, globalDir)
	result, err := r.Resolve(KindAgents, "explorer.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != SourceProject {
		t.Errorf("expected project to take precedence, got source %q", result.Source)
	}
	if !strings.Contains(result.Path, projectDir) {
		t.Errorf("expected path in project dir, got %q", result.Path)
	}
}

func TestResolve_NotFound_ReturnsError(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)

	r := New(projectDir, globalDir)
	_, err := r.Resolve(KindAgents, "nonexistent.md")
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestResolve_ChecksumConsistent(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "test.md", "deterministic content")

	r := New(projectDir, globalDir)
	r1, _ := r.Resolve(KindAgents, "test.md")
	r2, _ := r.Resolve(KindAgents, "test.md")
	if r1.Checksum != r2.Checksum {
		t.Error("same file should produce same checksum")
	}
}

// --- List tests ---

func TestList_EmptyKind_ReturnsEmpty(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)

	r := New(projectDir, globalDir)
	results, err := r.List(KindAgents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestList_UnionWithProjectPrecedence(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "shared.md", "project version")
	writeArtifact(t, projectDir, KindAgents, "project-only.md", "only in project")
	writeArtifact(t, globalDir, KindAgents, "shared.md", "global version")
	writeArtifact(t, globalDir, KindAgents, "global-only.md", "only in global")

	r := New(projectDir, globalDir)
	results, err := r.List(KindAgents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results (union), got %d", len(results))
	}

	// Check by name → source
	byName := make(map[string]ArtifactSource)
	for _, r := range results {
		byName[r.Name] = r.Source
	}

	if byName["shared.md"] != SourceProject {
		t.Errorf("shared.md should be SourceProject (override), got %q", byName["shared.md"])
	}
	if byName["project-only.md"] != SourceProject {
		t.Errorf("project-only.md should be SourceProject, got %q", byName["project-only.md"])
	}
	if byName["global-only.md"] != SourceGlobal {
		t.Errorf("global-only.md should be SourceGlobal, got %q", byName["global-only.md"])
	}
}

func TestList_SortedByName(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "charlie.md", "c")
	writeArtifact(t, globalDir, KindAgents, "alpha.md", "a")
	writeArtifact(t, globalDir, KindAgents, "bravo.md", "b")

	r := New(projectDir, globalDir)
	results, err := r.List(KindAgents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Name != "alpha.md" || results[1].Name != "bravo.md" || results[2].Name != "charlie.md" {
		t.Errorf("expected sorted order, got: %s, %s, %s", results[0].Name, results[1].Name, results[2].Name)
	}
}

func TestListAll_MultipleKinds(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "agent1.md", "a1")
	writeArtifact(t, globalDir, KindConventions, "conv1.md", "c1")
	writeArtifact(t, projectDir, KindCommands, "cmd1.md", "m1")

	r := New(projectDir, globalDir)
	results, err := r.ListAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 artifacts across all kinds, got %d", len(results))
	}
}

// --- Materialize tests ---

func TestMaterialize_CopiesGlobalPreservesProject(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "override.md", "project override content")
	writeArtifact(t, globalDir, KindAgents, "override.md", "global version — should NOT overwrite")
	writeArtifact(t, globalDir, KindAgents, "global-agent.md", "global agent content")
	writeArtifact(t, globalDir, KindConventions, "shared.md", "global convention")

	r := New(projectDir, globalDir)
	result, err := r.Materialize("2026.03.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Version != "2026.03.1" {
		t.Errorf("expected version 2026.03.1, got %s", result.Version)
	}
	if result.TotalArtifacts != 3 {
		t.Errorf("expected 3 total artifacts, got %d", result.TotalArtifacts)
	}
	if result.ProjectKept != 1 {
		t.Errorf("expected 1 project override kept, got %d", result.ProjectKept)
	}
	if result.Copied != 2 {
		t.Errorf("expected 2 global artifacts copied, got %d", result.Copied)
	}

	// Verify project override was NOT overwritten
	content, _ := os.ReadFile(filepath.Join(projectDir, "agents", "override.md"))
	if string(content) != "project override content" {
		t.Errorf("project override was overwritten: %s", string(content))
	}

	// Verify global artifact was copied
	content, _ = os.ReadFile(filepath.Join(projectDir, "agents", "global-agent.md"))
	if string(content) != "global agent content" {
		t.Errorf("global agent not copied correctly: %s", string(content))
	}

	// Verify manifest was written
	manifest, err := ReadManifest(projectDir)
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	if manifest.Manifest.Version != "2026.03.1" {
		t.Errorf("manifest version wrong: %s", manifest.Manifest.Version)
	}
	if len(manifest.Entries) != 3 {
		t.Errorf("expected 3 manifest entries, got %d", len(manifest.Entries))
	}

	// Verify entry sources
	for _, e := range manifest.Entries {
		if e.Path == "agents/override.md" && e.Source != "project" {
			t.Errorf("override.md should have source 'project', got %q", e.Source)
		}
	}
}

// --- Manifest tests ---

func TestManifest_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest("2026.03.1")
	m.Add("agents/test.md", SourceGlobal, "abc123")
	m.Add("conventions/shared.md", SourceProject, "def456")

	if err := WriteManifest(dir, m); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	read, err := ReadManifest(dir)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if read.Manifest.Version != "2026.03.1" {
		t.Errorf("version mismatch: %s", read.Manifest.Version)
	}
	if len(read.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(read.Entries))
	}
	if read.Entries[0].Path != "agents/test.md" {
		t.Errorf("first entry path: %s", read.Entries[0].Path)
	}
	if read.Entries[0].Source != "global" {
		t.Errorf("first entry source: %s", read.Entries[0].Source)
	}
}

func TestManifest_FindEntry(t *testing.T) {
	m := NewManifest("2026.03.1")
	m.Add("agents/test.md", SourceGlobal, "abc123")

	found := m.FindEntry("agents/test.md")
	if found == nil {
		t.Fatal("expected to find entry")
	}
	if found.Checksum != "abc123" {
		t.Errorf("wrong checksum: %s", found.Checksum)
	}

	notFound := m.FindEntry("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent entry")
	}
}

// --- Update tests ---

func TestUpdate_DetectsChanges(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "agent1.md", "original content")
	writeArtifact(t, globalDir, KindConventions, "conv1.md", "convention content")

	r := New(projectDir, globalDir)

	// Initial materialize
	_, err := r.Materialize("2026.03.1")
	if err != nil {
		t.Fatalf("initial materialize: %v", err)
	}

	// Modify global artifact
	writeArtifact(t, globalDir, KindAgents, "agent1.md", "UPDATED content")
	// Add new global artifact
	writeArtifact(t, globalDir, KindAgents, "agent2.md", "new agent")

	result, err := r.Update("2026.03.2")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d: %v", len(result.Updated), result.Updated)
	}
	if len(result.Added) != 1 {
		t.Errorf("expected 1 added, got %d: %v", len(result.Added), result.Added)
	}
	if result.Kept != 1 {
		t.Errorf("expected 1 kept (unchanged conv1.md), got %d", result.Kept)
	}

	// Verify updated file content
	content, _ := os.ReadFile(filepath.Join(projectDir, "agents", "agent1.md"))
	if string(content) != "UPDATED content" {
		t.Errorf("agent1.md not updated: %s", string(content))
	}
}

func TestUpdate_PreservesProjectOverrides(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, projectDir, KindAgents, "custom.md", "my custom agent")
	writeArtifact(t, globalDir, KindAgents, "custom.md", "global version")
	writeArtifact(t, globalDir, KindAgents, "standard.md", "standard content")

	r := New(projectDir, globalDir)

	_, err := r.Materialize("2026.03.1")
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}

	// Update global
	writeArtifact(t, globalDir, KindAgents, "standard.md", "updated standard")

	result, err := r.Update("2026.03.2")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	// Project override should still be kept
	content, _ := os.ReadFile(filepath.Join(projectDir, "agents", "custom.md"))
	if string(content) != "my custom agent" {
		t.Errorf("project override was modified: %s", string(content))
	}

	if result.Kept < 1 {
		t.Errorf("expected at least 1 kept (project override), got %d", result.Kept)
	}
}

func TestUpdate_DetectsRemovals(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "keep.md", "keep this")
	writeArtifact(t, globalDir, KindAgents, "remove-me.md", "will be removed")

	r := New(projectDir, globalDir)

	_, err := r.Materialize("2026.03.1")
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}

	// Remove artifact from global
	os.Remove(filepath.Join(globalDir, "agents", "remove-me.md"))

	result, err := r.Update("2026.03.2")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Errorf("expected 1 removed, got %d: %v", len(result.Removed), result.Removed)
	}

	// Verify file was actually removed from project
	if _, err := os.Stat(filepath.Join(projectDir, "agents", "remove-me.md")); !os.IsNotExist(err) {
		t.Error("removed artifact still exists in project")
	}
}

func TestUpdate_NoManifest_FallsBackToMaterialize(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindAgents, "agent.md", "content")

	r := New(projectDir, globalDir)
	result, err := r.Update("2026.03.1")
	if err != nil {
		t.Fatalf("update without manifest: %v", err)
	}

	if result.Version != "2026.03.1" {
		t.Errorf("expected version 2026.03.1, got %s", result.Version)
	}

	// Manifest should exist after update
	_, err = ReadManifest(projectDir)
	if err != nil {
		t.Errorf("manifest should exist after update: %v", err)
	}
}

// --- Edge case tests for coverage ---

func TestResolve_DirDoesNotExist_FallsThrough(t *testing.T) {
	// Project dir has no kind subdirectories at all
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindSkills, "skill.md", "from global")

	r := New(projectDir, globalDir)
	result, err := r.Resolve(KindSkills, "skill.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != SourceGlobal {
		t.Errorf("expected SourceGlobal, got %q", result.Source)
	}
}

func TestManifest_ReadNonExistent_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadManifest(dir)
	if err == nil {
		t.Fatal("expected error reading non-existent manifest")
	}
}

func TestMaterialize_CreatesKindDirs(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	writeArtifact(t, globalDir, KindSkills, "skill.md", "skill content")
	writeArtifact(t, globalDir, KindCommands, "cmd.md", "command content")

	r := New(projectDir, globalDir)
	result, err := r.Materialize("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Copied != 2 {
		t.Errorf("expected 2 copied, got %d", result.Copied)
	}
	// Verify directories were created
	for _, kind := range []string{"skills", "commands"} {
		if _, err := os.Stat(filepath.Join(projectDir, kind)); os.IsNotExist(err) {
			t.Errorf("expected %s directory to be created", kind)
		}
	}
}

func TestListAll_EmptyDirs_ReturnsEmpty(t *testing.T) {
	projectDir, globalDir := setupTestDirs(t)
	r := New(projectDir, globalDir)
	results, err := r.ListAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0, got %d", len(results))
	}
}

// --- AllKinds test ---

func TestAllKinds_ReturnsFourKinds(t *testing.T) {
	kinds := AllKinds()
	if len(kinds) != 4 {
		t.Errorf("expected 4 kinds, got %d", len(kinds))
	}
	expected := map[ArtifactKind]bool{
		KindAgents: true, KindSkills: true, KindCommands: true, KindConventions: true,
	}
	for _, k := range kinds {
		if !expected[k] {
			t.Errorf("unexpected kind: %s", k)
		}
	}
}
