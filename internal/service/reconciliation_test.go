package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

func TestReconciliationService_Reconcile(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "requirements.md"), []byte("# Requirements\n\nContent"), 0644)

	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"requirements": {
					ID:   "doc:spec/requirements",
					Path: "docs/spec/requirements.md",
					Zone: "spec",
				},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/requirements.md"] = []byte("# Requirements\n\nContent")
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true

	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	result, err := svc.Reconcile(dir, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", result.Status)
	}
	if result.Stats.Total != 1 {
		t.Errorf("Stats.Total = %d, want 1", result.Stats.Total)
	}

	// Lock should have been persisted
	if !lockRepo.Exists() {
		t.Error("lock should exist after reconcile")
	}
}

func TestReconciliationService_CheckOnly(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)

	configRepo := mem.NewConfigRepo()
	configRepo.Config = &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/a.md"] = []byte("# A")
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true

	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, err := svc.Reconcile(dir, domain.ReconcileOpts{CheckOnly: true})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// Lock should NOT have been written (check-only mode, FR-52)
	if lockRepo.Exists() {
		t.Error("lock should not exist after check-only reconcile")
	}
}

func TestReconciliationService_NoConfig(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	// No config set
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	_, err := svc.Reconcile("/test", domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

func TestReconciliationService_ReadStaleness_NoLock(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	svc := NewReconciliationService(configRepo, docRepo, lockRepo)

	info, err := svc.ReadStaleness("/test")
	if err != nil {
		t.Fatalf("ReadStaleness: %v", err)
	}
	if info != nil {
		t.Error("expected nil staleness info when no lock exists")
	}
}

func TestReconciliationService_ReadStaleness_WithLock(t *testing.T) {
	configRepo := mem.NewConfigRepo()
	docRepo := mem.NewDocRepo()
	lockRepo := mem.NewLockRepo()

	lockRepo.Lock = &domain.LockFile{
		GeneratedAt: time.Now().UTC(),
		Status:      domain.LockStale,
		Stats: domain.LockStats{
			Total: 3,
			Stale: 1,
			Clean: 2,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/architecture": {
				ID:          "doc:spec/architecture",
				Stale:       true,
				StaleReason: "dependency changed: doc:spec/requirements",
			},
			"doc:spec/requirements": {
				ID: "doc:spec/requirements",
			},
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
	if len(info.Stale) != 1 {
		t.Errorf("Stale count = %d, want 1", len(info.Stale))
	}
	if _, ok := info.Stale["doc:spec/architecture"]; !ok {
		t.Error("architecture should be in stale map")
	}
}
