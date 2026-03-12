package reconcile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

func setupTestProject(t *testing.T) (string, *domain.Config) {
	t.Helper()
	dir := t.TempDir()

	// Create docs structure
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)

	os.WriteFile(filepath.Join(docsDir, "requirements.md"), []byte("# Requirements\n\nFR-1: something"), 0644)
	os.WriteFile(filepath.Join(docsDir, "architecture.md"), []byte("# Architecture\n\nComponents"), 0644)
	os.WriteFile(filepath.Join(docsDir, "domain-model.md"), []byte("# Domain Model\n\nEntities"), 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"requirements": {
					ID:   "doc:spec/requirements",
					Path: "docs/spec/requirements.md",
					Zone: "spec",
				},
				"architecture": {
					ID:   "doc:spec/architecture",
					Path: "docs/spec/architecture.md",
					Zone: "spec",
				},
				"domain-model": {
					ID:   "doc:spec/domain-model",
					Path: "docs/spec/domain-model.md",
					Zone: "spec",
				},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/requirements", To: "doc:spec/architecture", Type: domain.EdgeInforms},
			{From: "doc:spec/requirements", To: "doc:spec/domain-model", Type: domain.EdgeInforms},
		},
	}

	return dir, cfg
}

func TestEngine_FirstRun(t *testing.T) {
	dir, cfg := setupTestProject(t)
	docRepo := mem.NewDocRepo()
	// Populate docRepo with files
	populateDocRepo(docRepo, dir)

	engine := NewEngine(docRepo)
	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// First run: all documents should be "changed" (no prior baseline)
	if len(result.Changed) != 3 {
		t.Errorf("Changed = %d, want 3", len(result.Changed))
	}

	// No staleness on first run (all documents are new)
	if len(result.Stale) != 0 {
		t.Errorf("Stale = %d, want 0", len(result.Stale))
	}

	// Status should be CLEAN (first run, all fresh)
	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", result.Status)
	}

	// Lock should have entries for all 3 documents
	if len(lock.Entries) != 3 {
		t.Errorf("Lock entries = %d, want 3", len(lock.Entries))
	}

	for _, entry := range lock.Entries {
		if entry.Hash == "" {
			t.Errorf("entry %s has no hash", entry.ID)
		}
		if entry.Stale {
			t.Errorf("entry %s should not be stale on first run", entry.ID)
		}
	}
}

func TestEngine_IncrementalChange(t *testing.T) {
	dir, cfg := setupTestProject(t)
	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)

	engine := NewEngine(docRepo)

	// First run
	_, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("First reconcile: %v", err)
	}

	// Modify requirements (change mtime and content)
	time.Sleep(10 * time.Millisecond)
	reqPath := filepath.Join(dir, "docs", "spec", "requirements.md")
	os.WriteFile(reqPath, []byte("# Requirements\n\nFR-1: something\nFR-2: new requirement"), 0644)

	// Second run
	result, _, err := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Second reconcile: %v", err)
	}

	// Requirements should be changed
	if len(result.Changed) != 1 {
		t.Errorf("Changed = %d, want 1", len(result.Changed))
	}
	if len(result.Changed) == 1 && result.Changed[0] != "doc:spec/requirements" {
		t.Errorf("Changed[0] = %q, want doc:spec/requirements", result.Changed[0])
	}

	// Architecture and domain-model should be stale
	if len(result.Stale) != 2 {
		t.Errorf("Stale = %d, want 2", len(result.Stale))
	}
	if _, ok := result.Stale["doc:spec/architecture"]; !ok {
		t.Error("architecture should be stale")
	}
	if _, ok := result.Stale["doc:spec/domain-model"]; !ok {
		t.Error("domain-model should be stale")
	}

	if result.Status != domain.LockStale {
		t.Errorf("Status = %q, want STALE", result.Status)
	}
}

func TestEngine_ForceMode(t *testing.T) {
	dir, cfg := setupTestProject(t)
	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)

	engine := NewEngine(docRepo)

	// Create a pre-existing lock with stale entries
	existingLock := &domain.LockFile{
		Status: domain.LockStale,
		Entries: map[string]domain.LockEntry{
			"doc:spec/architecture": {
				ID:          "doc:spec/architecture",
				Stale:       true,
				StaleReason: "old reason",
			},
		},
	}

	// Force mode: discard existing lock
	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{Force: true})
	if err != nil {
		t.Fatalf("Force reconcile: %v", err)
	}
	_ = existingLock // The service is responsible for passing nil lock when Force

	// All documents re-hashed, no staleness
	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", result.Status)
	}
	for _, entry := range lock.Entries {
		if entry.Stale {
			t.Errorf("entry %s should not be stale after force", entry.ID)
		}
	}
}

func TestEngine_CycleDetection(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			{From: "doc:spec/b", To: "doc:spec/a", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	engine := NewEngine(docRepo)

	_, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !containsStr(err.Error(), "circular dependency") {
		t.Errorf("error should mention circular dependency: %v", err)
	}
}

func TestEngine_MissingDocument(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	// b.md intentionally missing

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	engine := NewEngine(docRepo)

	result, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Missing) != 1 {
		t.Errorf("Missing = %d, want 1", len(result.Missing))
	}
	if result.Status != domain.LockDirty {
		t.Errorf("Status = %q, want DIRTY", result.Status)
	}
}

func TestEngine_NoGraph(t *testing.T) {
	dir, cfg := setupTestProject(t)
	cfg.Graph = nil // No graph edges

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)

	engine := NewEngine(docRepo)

	// First run
	_, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("First reconcile: %v", err)
	}

	// Modify requirements
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "docs", "spec", "requirements.md"), []byte("# Changed"), 0644)

	// Second run: change detected but no propagation (FR-64)
	result, _, err := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Second reconcile: %v", err)
	}

	if len(result.Changed) != 1 {
		t.Errorf("Changed = %d, want 1", len(result.Changed))
	}
	if len(result.Stale) != 0 {
		t.Errorf("Stale = %d, want 0 (no graph)", len(result.Stale))
	}
}

func TestEngine_UndeclaredEdges(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/nonexistent", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	engine := NewEngine(docRepo)

	_, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected error for undeclared edge reference")
	}
	if !containsStr(err.Error(), "doc:spec/nonexistent") {
		t.Errorf("error should mention undeclared doc: %v", err)
	}
}

// populateDocRepo adds files from disk into the in-memory doc repo.
func populateDocRepo(docRepo *mem.DocRepo, projectRoot string) {
	docsDir := filepath.Join(projectRoot, "docs")
	filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(projectRoot, path)
		content, _ := os.ReadFile(path)
		docRepo.Files[relPath] = content
		docRepo.Docs[relPath] = domain.Document{
			Path:    relPath,
			AbsPath: path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		return nil
	})
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
