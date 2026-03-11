package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
)

// setupBenchDocs creates n documents with content on disk and returns config and repos.
func setupBenchDocs(b *testing.B, n int) (string, *domain.Config, *mem.DocRepo) {
	b.Helper()

	dir := b.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		b.Fatal(err)
	}

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("doc%03d", i)
		relPath := fmt.Sprintf("docs/spec/%s.md", name)
		absPath := filepath.Join(dir, relPath)
		content := []byte(fmt.Sprintf("# Document %d\n\nContent for document number %d with enough text to be realistic.", i, i))

		if err := os.WriteFile(absPath, content, 0644); err != nil {
			b.Fatal(err)
		}

		cfg.Documents["spec"][name] = domain.DocEntry{
			ID:   fmt.Sprintf("doc:spec/%s", name),
			Path: relPath,
			Zone: "spec",
		}
		docRepo.Files[relPath] = content
	}

	return dir, cfg, docRepo
}

// setupBenchGraph creates a linear dependency chain: doc000 -> doc001 -> ... -> doc(n-1).
func setupBenchGraph(cfg *domain.Config, n int) {
	for i := 0; i < n-1; i++ {
		cfg.Graph = append(cfg.Graph, domain.GraphEdge{
			From: fmt.Sprintf("doc:spec/doc%03d", i),
			To:   fmt.Sprintf("doc:spec/doc%03d", i+1),
			Type: domain.EdgeInforms,
		})
	}
}

func BenchmarkReconcile_Full_50Docs(b *testing.B) {
	dir, cfg, docRepo := setupBenchDocs(b, 50)
	setupBenchGraph(cfg, 50)
	engine := NewEngine(docRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReconcile_Incremental_50Docs(b *testing.B) {
	dir, cfg, docRepo := setupBenchDocs(b, 50)
	setupBenchGraph(cfg, 50)
	engine := NewEngine(docRepo)

	// First run to create baseline lock
	_, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		b.Fatal(err)
	}

	// Modify one document
	modPath := filepath.Join(dir, "docs/spec/doc000.md")
	if err := os.WriteFile(modPath, []byte("# Modified Document 0\n\nUpdated content."), 0644); err != nil {
		b.Fatal(err)
	}
	// Ensure mtime differs
	future := time.Now().Add(time.Second)
	os.Chtimes(modPath, future, future)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReconcile_Full_100Docs(b *testing.B) {
	dir, cfg, docRepo := setupBenchDocs(b, 100)
	setupBenchGraph(cfg, 100)
	engine := NewEngine(docRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReconcile_Force_50Docs(b *testing.B) {
	dir, cfg, docRepo := setupBenchDocs(b, 50)
	setupBenchGraph(cfg, 50)
	engine := NewEngine(docRepo)

	// First run to create lock
	_, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{Force: true})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestBenchmark_Performance verifies reconciliation meets performance targets.
// Full reconciliation of 50 docs should be <200ms; incremental should be <50ms.
func TestBenchmark_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Setup 50 docs
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs", "spec")
	os.MkdirAll(docsDir, 0755)

	cfg := &domain.Config{
		Documents: map[string]map[string]domain.DocEntry{
			"spec": {},
		},
	}

	docRepo := mem.NewDocRepo()
	docRepo.Dirs["docs"] = true
	docRepo.Dirs["docs/spec"] = true

	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("doc%03d", i)
		relPath := fmt.Sprintf("docs/spec/%s.md", name)
		absPath := filepath.Join(dir, relPath)
		content := []byte(fmt.Sprintf("# Document %d\n\nContent for document number %d.", i, i))
		os.WriteFile(absPath, content, 0644)

		cfg.Documents["spec"][name] = domain.DocEntry{
			ID:   fmt.Sprintf("doc:spec/%s", name),
			Path: relPath,
			Zone: "spec",
		}
		docRepo.Files[relPath] = content

		if i < 49 {
			cfg.Graph = append(cfg.Graph, domain.GraphEdge{
				From: fmt.Sprintf("doc:spec/doc%03d", i),
				To:   fmt.Sprintf("doc:spec/doc%03d", i+1),
				Type: domain.EdgeInforms,
			})
		}
	}

	engine := NewEngine(docRepo)

	// Full reconciliation
	start := time.Now()
	_, lock, err := engine.Reconcile(dir, cfg, nil, domain.ReconcileOpts{})
	fullDuration := time.Since(start)
	if err != nil {
		t.Fatalf("full reconcile: %v", err)
	}

	t.Logf("Full reconciliation (50 docs): %v", fullDuration)
	if fullDuration > 200*time.Millisecond {
		t.Errorf("full reconciliation took %v, want <200ms", fullDuration)
	}

	// Modify one document
	modPath := filepath.Join(dir, "docs/spec/doc000.md")
	os.WriteFile(modPath, []byte("# Modified\n\nUpdated."), 0644)
	future := time.Now().Add(time.Second)
	os.Chtimes(modPath, future, future)

	// Incremental reconciliation
	start = time.Now()
	_, _, err = engine.Reconcile(dir, cfg, lock, domain.ReconcileOpts{})
	incrDuration := time.Since(start)
	if err != nil {
		t.Fatalf("incremental reconcile: %v", err)
	}

	t.Logf("Incremental reconciliation (50 docs, 1 change): %v", incrDuration)
	if incrDuration > 50*time.Millisecond {
		t.Errorf("incremental reconciliation took %v, want <50ms", incrDuration)
	}
}
