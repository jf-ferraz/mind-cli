package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// FR-52: Check-only mode does not write lock file.
func TestReconciliationService_FR52_CheckOnlyNoWrite(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)

	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/a.md"] = []byte("# A")
	docRepo.Files["docs/spec/b.md"] = []byte("# B")
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	// First reconcile to create baseline
	_, err := svc.Reconcile(dir, domain.ReconcileOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !lockRepo.Exists() {
		t.Fatal("lock should exist after first reconcile")
	}

	// Record lock state
	lockBefore, _ := lockRepo.Read()
	generatedBefore := lockBefore.GeneratedAt

	// Check-only mode
	result, err := svc.Reconcile(dir, domain.ReconcileOpts{CheckOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	// Lock should not have been modified
	lockAfter, _ := lockRepo.Read()
	if !lockAfter.GeneratedAt.Equal(generatedBefore) {
		t.Error("check-only mode should not modify lock file")
	}

	// Result should still be computed correctly
	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", result.Status)
	}
}

// FR-53: Force mode re-hashes everything, clears staleness.
func TestReconciliationService_FR53_ForceClearsStaleness(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)

	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/a.md"] = []byte("# A")
	docRepo.Files["docs/spec/b.md"] = []byte("# B")
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	// First reconcile
	_, _ = svc.Reconcile(dir, domain.ReconcileOpts{})

	// Modify A to create staleness
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Modified"), 0644)
	docRepo.Files["docs/spec/a.md"] = []byte("# A Modified")

	// Normal reconcile - B should be stale
	result, _ := svc.Reconcile(dir, domain.ReconcileOpts{})
	if result.Status != domain.LockStale {
		t.Fatalf("expected STALE before force, got %q", result.Status)
	}

	// Force reconcile - clear staleness
	result, err := svc.Reconcile(dir, domain.ReconcileOpts{Force: true})
	if err != nil {
		t.Fatalf("Force reconcile: %v", err)
	}

	if result.Status != domain.LockClean {
		t.Errorf("Status = %q after force, want CLEAN", result.Status)
	}
}

// FR-56: Missing config produces error.
func TestReconciliationService_FR56_MissingConfig(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	// No config set
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, err := svc.Reconcile("/test", domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
	if !contains(err.Error(), "mind.toml required") {
		t.Errorf("error message = %q, want to contain 'mind.toml required'", err.Error())
	}
}

// FR-56: Config with no documents section produces error.
func TestReconciliationService_FR56_NoDocuments(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		// No Documents section
	}
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, err := svc.Reconcile("/test", domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected error when no documents section")
	}
	if !contains(err.Error(), "mind.toml required") {
		t.Errorf("error message = %q, want to contain 'mind.toml required'", err.Error())
	}
}

// Service LoadGraph method.
func TestReconciliationService_LoadGraph(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			{From: "doc:spec/b", To: "doc:spec/c", Type: domain.EdgeRequires},
		},
	}

	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	graph, stale, err := svc.LoadGraph("/test")
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	if graph == nil {
		t.Fatal("graph should not be nil")
	}
	if len(graph.Nodes) != 3 {
		t.Errorf("nodes = %d, want 3", len(graph.Nodes))
	}
	if stale != nil {
		t.Errorf("stale should be nil when no lock exists, got %v", stale)
	}
}

// LoadGraph with existing lock file and staleness data.
func TestReconciliationService_LoadGraph_WithStaleness(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()
	lockRepo.Lock = &domain.LockFile{
		Status: domain.LockStale,
		Entries: map[string]domain.LockEntry{
			"doc:spec/b": {
				ID:          "doc:spec/b",
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/a",
			},
			"doc:spec/a": {
				ID: "doc:spec/a",
			},
		},
	}

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, stale, err := svc.LoadGraph("/test")
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	if stale == nil {
		t.Fatal("stale should not be nil when lock exists with stale entries")
	}
	if _, ok := stale["doc:spec/b"]; !ok {
		t.Error("doc:spec/b should be in stale map")
	}
}

// LoadGraph with no config.
func TestReconciliationService_LoadGraph_NoConfig(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	// No config
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, _, err := svc.LoadGraph("/test")
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

// ReadStaleness with stale entries returns correct info.
func TestReconciliationService_ReadStaleness_StaleEntries(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	now := time.Now().UTC()
	lockRepo.Lock = &domain.LockFile{
		GeneratedAt: now,
		Status:      domain.LockStale,
		Stats: domain.LockStats{
			Total: 3,
			Stale: 2,
			Clean: 1,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/a": {ID: "doc:spec/a", Stale: true, StaleReason: "reason A"},
			"doc:spec/b": {ID: "doc:spec/b", Stale: true, StaleReason: "reason B"},
			"doc:spec/c": {ID: "doc:spec/c", Stale: false},
		},
	}

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	info, err := svc.ReadStaleness("/test")
	if err != nil {
		t.Fatalf("ReadStaleness: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil staleness info")
	}
	if info.Status != domain.LockStale {
		t.Errorf("Status = %q, want STALE", info.Status)
	}
	if len(info.Stale) != 2 {
		t.Errorf("Stale count = %d, want 2", len(info.Stale))
	}
	if info.Stats.Total != 3 {
		t.Errorf("Stats.Total = %d, want 3", info.Stats.Total)
	}
}

// ReadStaleness with CLEAN lock returns info with empty stale map.
func TestReconciliationService_ReadStaleness_CleanLock(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	lockRepo.Lock = &domain.LockFile{
		Status: domain.LockClean,
		Stats:  domain.LockStats{Total: 3, Clean: 3},
		Entries: map[string]domain.LockEntry{
			"doc:spec/a": {ID: "doc:spec/a"},
			"doc:spec/b": {ID: "doc:spec/b"},
			"doc:spec/c": {ID: "doc:spec/c"},
		},
	}

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	info, err := svc.ReadStaleness("/test")
	if err != nil {
		t.Fatal(err)
	}
	if info.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", info.Status)
	}
	if len(info.Stale) != 0 {
		t.Errorf("Stale count = %d, want 0", len(info.Stale))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
