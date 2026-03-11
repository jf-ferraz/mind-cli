package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// FR-51: First run creates lock with all documents, all stale=false.
func TestEngine_FR51_FirstRunCreatesCleanLock(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)

	docs := map[string]string{
		"requirements":  "# Requirements\nFR-1: something",
		"architecture":  "# Architecture\nComponents",
		"domain-model":  "# Domain Model\nEntities",
		"api-contracts": "# API Contracts\nEndpoints",
		"project-brief": "# Brief\nVision",
	}
	for name, content := range docs {
		os.WriteFile(filepath.Join(docsDir, name+".md"), []byte(content), 0644)
	}

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{"spec": {}},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/requirements", To: "doc:spec/architecture", Type: domain.EdgeInforms},
			{From: "doc:spec/requirements", To: "doc:spec/domain-model", Type: domain.EdgeInforms},
			{From: "doc:spec/architecture", To: "doc:spec/api-contracts", Type: domain.EdgeRequires},
		},
	}
	for name := range docs {
		cfg.Documents["spec"][name] = domain.DocEntry{
			ID:   "doc:spec/" + name,
			Path: "docs/spec/" + name + ".md",
			Zone: "spec",
		}
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// 5 documents, all hashed
	if len(lock.Entries) != 5 {
		t.Errorf("lock entries = %d, want 5", len(lock.Entries))
	}
	for id, entry := range lock.Entries {
		if entry.Hash == "" {
			t.Errorf("entry %s has no hash", id)
		}
		if entry.Stale {
			t.Errorf("entry %s should not be stale on first run", id)
		}
		if !strings.HasPrefix(entry.Hash, "sha256:") {
			t.Errorf("entry %s hash has wrong prefix: %s", id, entry.Hash)
		}
	}

	// First run is CLEAN
	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN", result.Status)
	}
	if result.Stats.Total != 5 {
		t.Errorf("Stats.Total = %d, want 5", result.Stats.Total)
	}
}

// FR-53: Force mode re-hashes everything and clears staleness.
func TestEngine_FR53_ForceClearsStaleness(t *testing.T) {
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
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	// First run
	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	// Modify A to make B stale
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Modified"), 0644)

	// Normal run - B becomes stale
	result, lock, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
	if len(result.Stale) == 0 {
		t.Fatal("expected stale documents before force")
	}

	// Force mode - pass nil lock (service sets lock=nil for force)
	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{Force: true})
	if err != nil {
		t.Fatalf("Force reconcile: %v", err)
	}

	if result.Status != domain.LockClean {
		t.Errorf("Status = %q, want CLEAN after force", result.Status)
	}
	for id, entry := range lock.Entries {
		if entry.Stale {
			t.Errorf("entry %s should not be stale after force", id)
		}
		if entry.Hash == "" {
			t.Errorf("entry %s has no hash after force", id)
		}
	}
}

// FR-60: Graph construction with forward and reverse edges.
func TestEngine_FR60_GraphEdgeCounts(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "A", To: "C", Type: domain.EdgeRequires},
		{From: "B", To: "D", Type: domain.EdgeValidates},
		{From: "C", To: "D", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	if len(g.Nodes) != 4 {
		t.Errorf("nodes = %d, want 4", len(g.Nodes))
	}

	// Count total forward edges
	totalForward := 0
	for _, edges := range g.Forward {
		totalForward += len(edges)
	}
	if totalForward != 4 {
		t.Errorf("forward edges = %d, want 4", totalForward)
	}

	// Count total reverse edges
	totalReverse := 0
	for _, edges := range g.Reverse {
		totalReverse += len(edges)
	}
	if totalReverse != 4 {
		t.Errorf("reverse edges = %d, want 4", totalReverse)
	}
}

// FR-62: Cycle detection with full cycle path.
func TestEngine_FR62_CyclePathInError(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)
	os.WriteFile(filepath.Join(docsDir, "c.md"), []byte("# C"), 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
				"c": {ID: "doc:spec/c", Path: "docs/spec/c.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			{From: "doc:spec/b", To: "doc:spec/c", Type: domain.EdgeInforms},
			{From: "doc:spec/c", To: "doc:spec/a", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	engine := NewEngine(docRepo)

	_, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err == nil {
		t.Fatal("expected cycle error")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "circular dependency") {
		t.Errorf("error should mention circular dependency: %q", errMsg)
	}
	// Must contain all three nodes in the cycle path
	if !strings.Contains(errMsg, "doc:spec/a") {
		t.Errorf("cycle path should contain doc:spec/a: %q", errMsg)
	}
	if !strings.Contains(errMsg, "doc:spec/b") {
		t.Errorf("cycle path should contain doc:spec/b: %q", errMsg)
	}
	if !strings.Contains(errMsg, "doc:spec/c") {
		t.Errorf("cycle path should contain doc:spec/c: %q", errMsg)
	}
}

// FR-63: Undeclared document ID in graph produces error.
func TestEngine_FR63_UndeclaredDocInGraph(t *testing.T) {
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
		t.Fatal("expected error for undeclared doc in graph")
	}
	if !strings.Contains(err.Error(), "doc:spec/nonexistent") {
		t.Errorf("error should mention undeclared doc: %v", err)
	}
	if !strings.Contains(err.Error(), "graph references undeclared document") {
		t.Errorf("error message format wrong: %v", err)
	}
}

// FR-64: No graph entries still tracks documents.
func TestEngine_FR64_NoGraphStillTracksChanges(t *testing.T) {
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
		// No graph edges
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	// First run
	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	// Modify A
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Modified"), 0644)

	// Second run
	result, _, err := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// A changed, but no staleness propagation
	if len(result.Changed) != 1 {
		t.Errorf("Changed = %d, want 1", len(result.Changed))
	}
	if len(result.Stale) != 0 {
		t.Errorf("Stale = %d, want 0 (no graph)", len(result.Stale))
	}
}

// FR-65: Staleness propagates downstream only.
func TestEngine_FR65_DownstreamOnly(t *testing.T) {
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
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	// First run
	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	// Modify B (downstream). A should NOT be stale.
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B Modified"), 0644)

	result, _, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})

	if _, ok := result.Stale["doc:spec/a"]; ok {
		t.Error("A should NOT be stale (reverse propagation not allowed)")
	}
	if len(result.Changed) != 1 || result.Changed[0] != "doc:spec/b" {
		t.Errorf("Changed = %v, want [doc:spec/b]", result.Changed)
	}
}

// FR-66: Transitive staleness propagation A -> B -> C.
func TestEngine_FR66_TransitivePropagation(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)
	os.WriteFile(filepath.Join(docsDir, "c.md"), []byte("# C"), 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
				"c": {ID: "doc:spec/c", Path: "docs/spec/c.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			{From: "doc:spec/b", To: "doc:spec/c", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	// First run
	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	// Modify A
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Modified"), 0644)

	result, _, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})

	// B should be stale (direct)
	if _, ok := result.Stale["doc:spec/b"]; !ok {
		t.Error("B should be stale")
	}
	// C should be stale (transitive)
	if _, ok := result.Stale["doc:spec/c"]; !ok {
		t.Error("C should be stale (transitive)")
	}
	// C's reason should mention transitive
	if reason, ok := result.Stale["doc:spec/c"]; ok {
		if !strings.Contains(reason, "transitive") {
			t.Errorf("C reason should mention transitive: %q", reason)
		}
	}
}

// FR-68: Changed document is not stale.
func TestEngine_FR68_ChangedNotStale(t *testing.T) {
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
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	// Modify both A and B
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Modified"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B Modified"), 0644)

	result, _, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})

	// Both should be changed, neither stale
	if _, ok := result.Stale["doc:spec/a"]; ok {
		t.Error("A should NOT be stale (it changed)")
	}
	if _, ok := result.Stale["doc:spec/b"]; ok {
		t.Error("B should NOT be stale (it changed itself)")
	}
	if len(result.Changed) != 2 {
		t.Errorf("Changed = %d, want 2", len(result.Changed))
	}
}

// FR-69: Diamond graph - D marked stale only once.
func TestEngine_FR69_DiamondOnceOnly(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	for _, name := range []string{"a", "b", "c", "d"} {
		os.WriteFile(filepath.Join(docsDir, name+".md"), []byte("# "+strings.ToUpper(name)), 0644)
	}

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
				"c": {ID: "doc:spec/c", Path: "docs/spec/c.md", Zone: "spec"},
				"d": {ID: "doc:spec/d", Path: "docs/spec/d.md", Zone: "spec"},
			},
		},
		Graph: []domain.GraphEdge{
			{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			{From: "doc:spec/a", To: "doc:spec/c", Type: domain.EdgeInforms},
			{From: "doc:spec/b", To: "doc:spec/d", Type: domain.EdgeInforms},
			{From: "doc:spec/c", To: "doc:spec/d", Type: domain.EdgeInforms},
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Changed"), 0644)

	result, _, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})

	// D should be stale with exactly one reason
	reason, ok := result.Stale["doc:spec/d"]
	if !ok {
		t.Fatal("D should be stale")
	}
	if reason == "" {
		t.Error("D should have a reason")
	}

	// Count how many times D appears in the stale map (should be exactly 1 entry)
	count := 0
	for id := range result.Stale {
		if id == "doc:spec/d" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("D appeared %d times in stale map, want 1", count)
	}
}

// FR-74: First run with no lock produces CLEAN status.
func TestEngine_FR74_FirstRunClean(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(docsDir, fmt.Sprintf("doc%d.md", i)), []byte(fmt.Sprintf("# Doc %d\nContent", i)), 0644)
	}

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{"spec": {}},
	}
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("doc%d", i)
		cfg.Documents["spec"][name] = domain.DocEntry{
			ID:   fmt.Sprintf("doc:spec/%s", name),
			Path: fmt.Sprintf("docs/spec/%s.md", name),
			Zone: "spec",
		}
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if result.Status != domain.LockClean {
		t.Errorf("first run Status = %q, want CLEAN", result.Status)
	}
	if len(lock.Entries) != 5 {
		t.Errorf("lock entries = %d, want 5", len(lock.Entries))
	}
	for id, entry := range lock.Entries {
		if entry.Stale {
			t.Errorf("%s should not be stale on first run", id)
		}
	}
}

// FR-76/BR-33: Status priority STALE > DIRTY > CLEAN.
func TestEngine_FR76_StatusPriority(t *testing.T) {
	t.Run("CLEAN when no stale and no missing", func(t *testing.T) {
		dir, cfg := setupTestProject(t)
		docRepo := mem.NewDocRepo()
		populateDocRepo(docRepo, dir)
		engine := NewEngine(docRepo)

		result, _, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
		if result.Status != domain.LockClean {
			t.Errorf("Status = %q, want CLEAN", result.Status)
		}
	})

	t.Run("DIRTY when missing but no stale", func(t *testing.T) {
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
		populateDocRepo(docRepo, dir)
		engine := NewEngine(docRepo)

		result, _, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
		if result.Status != domain.LockDirty {
			t.Errorf("Status = %q, want DIRTY", result.Status)
		}
	})

	t.Run("STALE takes precedence over missing", func(t *testing.T) {
		dir := t.TempDir()
		docsDir := filepath.Join(dir, "docs", "spec")
		os.MkdirAll(docsDir, 0755)
		os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
		os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)
		// c.md missing

		cfg := &domain.Config{
			Documents: map[string]map[string]domain.DocEntry{
				"spec": {
					"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
					"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
					"c": {ID: "doc:spec/c", Path: "docs/spec/c.md", Zone: "spec"},
				},
			},
			Graph: []domain.GraphEdge{
				{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
			},
		}

		docRepo := mem.NewDocRepo()
		populateDocRepo(docRepo, dir)
		engine := NewEngine(docRepo)

		// First run
		_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

		// Modify A to make B stale (while C remains missing)
		time.Sleep(10 * time.Millisecond)
		os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A Changed"), 0644)

		result, _, _ := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})

		if result.Status != domain.LockStale {
			t.Errorf("Status = %q, want STALE (takes precedence over DIRTY)", result.Status)
		}
	})
}

// XC-11: Pruning entries for removed documents.
func TestEngine_XC11_PruneRemovedDocuments(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(docsDir, "b.md"), []byte("# B"), 0644)

	// Initial config with 2 docs
	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
				"b": {ID: "doc:spec/b", Path: "docs/spec/b.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	_, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if len(lock.Entries) != 2 {
		t.Fatalf("initial lock entries = %d, want 2", len(lock.Entries))
	}

	// Remove B from config
	cfg2 := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"a": {ID: "doc:spec/a", Path: "docs/spec/a.md", Zone: "spec"},
			},
		},
	}

	_, lock2, _ := engine.Reconcile(dir, cfg2, lock, domain.ReconcileOpts{})
	if len(lock2.Entries) != 1 {
		t.Errorf("pruned lock entries = %d, want 1", len(lock2.Entries))
	}
	if _, ok := lock2.Entries["doc:spec/b"]; ok {
		t.Error("B should have been pruned from lock")
	}
}

// Binary file detection produces warning.
func TestEngine_BinaryFileWarning(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)

	// Create a binary file (non-UTF8 content)
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	os.WriteFile(filepath.Join(docsDir, "binary.md"), binaryContent, 0644)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"binary": {ID: "doc:spec/binary", Path: "docs/spec/binary.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/binary.md"] = binaryContent
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true
	engine := NewEngine(docRepo)

	result, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// Binary file should still be hashed
	entry, ok := lock.Entries["doc:spec/binary"]
	if !ok {
		t.Fatal("binary doc should have lock entry")
	}
	if entry.Hash == "" {
		t.Error("binary file should have a hash")
	}

	// Should have a warning about binary file
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "binary file detected") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected binary file warning, got warnings: %v", result.Warnings)
	}
}

// Undeclared file detection (FR-85).
func TestEngine_FR85_UndeclaredFiles(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "declared.md"), []byte("# Declared"), 0644)
	os.WriteFile(filepath.Join(docsDir, "notes.md"), []byte("# Notes"), 0644) // undeclared

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"declared": {ID: "doc:spec/declared", Path: "docs/spec/declared.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	result, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(result.Undeclared) == 0 {
		t.Error("expected undeclared files to be detected")
	}

	found := false
	for _, u := range result.Undeclared {
		if strings.Contains(u, "notes.md") {
			found = true
		}
	}
	if !found {
		t.Errorf("notes.md should be in undeclared list, got: %v", result.Undeclared)
	}
}

// Missing document status (FR-76 DIRTY).
func TestEngine_MissingDocumentProducesDirty(t *testing.T) {
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "exists.md"), []byte("# Exists"), 0644)
	// missing.md intentionally not created

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"exists":  {ID: "doc:spec/exists", Path: "docs/spec/exists.md", Zone: "spec"},
				"missing": {ID: "doc:spec/missing", Path: "docs/spec/missing.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	populateDocRepo(docRepo, dir)
	engine := NewEngine(docRepo)

	result, lock, _ := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})

	if len(result.Missing) != 1 {
		t.Errorf("Missing = %d, want 1", len(result.Missing))
	}
	if result.Missing[0] != "doc:spec/missing" {
		t.Errorf("Missing[0] = %q, want doc:spec/missing", result.Missing[0])
	}
	if result.Status != domain.LockDirty {
		t.Errorf("Status = %q, want DIRTY", result.Status)
	}

	entry := lock.Entries["doc:spec/missing"]
	if entry.Status != domain.EntryMissing {
		t.Errorf("missing entry status = %q, want MISSING", entry.Status)
	}
}

// Symlink outside project root produces warning.
func TestEngine_SymlinkOutsideProjectRootWarning(t *testing.T) {
	projectDir := t.TempDir()
	externalDir := t.TempDir()

	docsDir := filepath.Join(projectDir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)

	// Create external file
	externalFile := filepath.Join(externalDir, "external.md")
	os.WriteFile(externalFile, []byte("# External"), 0644)

	// Create symlink to external file
	linkPath := filepath.Join(docsDir, "linked.md")
	os.Symlink(externalFile, linkPath)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {
				"linked": {ID: "doc:spec/linked", Path: "docs/spec/linked.md", Zone: "spec"},
			},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Files["docs/spec/linked.md"] = []byte("# External")
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true
	engine := NewEngine(docRepo)

	result, lock, err := engine.Reconcile(projectDir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// Should still be hashed
	entry := lock.Entries["doc:spec/linked"]
	if entry.Hash == "" {
		t.Error("symlink should still produce a hash")
	}

	// Should have warning about outside project root
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "symlink target outside project root") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected symlink outside warning, got: %v", result.Warnings)
	}
}
