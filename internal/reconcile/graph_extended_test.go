package reconcile

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-62: Self-loop cycle is detected.
func TestDetectCycle_SelfLoopPath(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "A", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle == nil {
		t.Fatal("expected cycle for self-loop")
	}
	// Should start and end with same node
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node: %v", cycle)
	}
}

// FR-62: Complex cycle with multiple possible paths.
func TestDetectCycle_ComplexCycle(t *testing.T) {
	// A -> B, B -> C, C -> D, D -> B (cycle: B -> C -> D -> B)
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeInforms},
		{From: "C", To: "D", Type: domain.EdgeInforms},
		{From: "D", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle == nil {
		t.Fatal("expected cycle")
	}

	// Cycle should include B, C, D
	str := strings.Join(cycle, " -> ")
	if !strings.Contains(str, "B") || !strings.Contains(str, "C") || !strings.Contains(str, "D") {
		t.Errorf("cycle should contain B, C, D: %s", str)
	}
}

// Acyclic diamond graph should not have cycle.
func TestDetectCycle_DiamondAcyclic(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "A", To: "C", Type: domain.EdgeInforms},
		{From: "B", To: "D", Type: domain.EdgeInforms},
		{From: "C", To: "D", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle != nil {
		t.Errorf("expected no cycle for diamond graph, got %v", cycle)
	}
}

// Long acyclic chain should not have cycle.
func TestDetectCycle_LongAcyclicChain(t *testing.T) {
	var edges []domain.GraphEdge
	for i := 0; i < 20; i++ {
		from := string(rune('A' + i))
		to := string(rune('A' + i + 1))
		edges = append(edges, domain.GraphEdge{From: from, To: to, Type: domain.EdgeInforms})
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle != nil {
		t.Errorf("expected no cycle for linear chain, got %v", cycle)
	}
}

// FR-63: ValidateEdges reports all undeclared docs, not just the first.
func TestValidateEdges_MultipleUndeclared(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/x", To: "doc:spec/y", Type: domain.EdgeInforms},
		{From: "doc:spec/y", To: "doc:spec/z", Type: domain.EdgeInforms},
	}
	declared := map[string]bool{
		"doc:spec/y": true,
	}

	err := ValidateEdges(edges, declared)
	if err == nil {
		t.Fatal("expected error for undeclared docs")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "doc:spec/x") {
		t.Errorf("error should mention doc:spec/x: %v", err)
	}
	if !strings.Contains(errMsg, "doc:spec/z") {
		t.Errorf("error should mention doc:spec/z: %v", err)
	}
}

// ValidateEdges with all docs declared passes.
func TestValidateEdges_AllDeclaredPasses(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/a", To: "doc:spec/b", Type: domain.EdgeInforms},
		{From: "doc:spec/b", To: "doc:spec/c", Type: domain.EdgeRequires},
		{From: "doc:spec/c", To: "doc:spec/d", Type: domain.EdgeValidates},
	}
	declared := map[string]bool{
		"doc:spec/a": true,
		"doc:spec/b": true,
		"doc:spec/c": true,
		"doc:spec/d": true,
	}

	err := ValidateEdges(edges, declared)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// Duplicate undeclared references are reported once.
func TestValidateEdges_DuplicateUndeclaredOnce(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/ghost", To: "doc:spec/a", Type: domain.EdgeInforms},
		{From: "doc:spec/ghost", To: "doc:spec/b", Type: domain.EdgeInforms},
	}
	declared := map[string]bool{
		"doc:spec/a": true,
		"doc:spec/b": true,
	}

	err := ValidateEdges(edges, declared)
	if err == nil {
		t.Fatal("expected error")
	}

	// "ghost" should appear once, not twice
	errMsg := err.Error()
	count := strings.Count(errMsg, "doc:spec/ghost")
	if count != 1 {
		t.Errorf("doc:spec/ghost appears %d times in error, want 1", count)
	}
}

// DetectCycle is deterministic (sorted node order).
func TestDetectCycle_Deterministic(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "C", To: "A", Type: domain.EdgeInforms},
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	// Run multiple times and verify same result
	var firstCycle []string
	for i := 0; i < 10; i++ {
		cycle := DetectCycle(g)
		if cycle == nil {
			t.Fatal("expected cycle")
		}
		if firstCycle == nil {
			firstCycle = cycle
		} else {
			if strings.Join(cycle, ",") != strings.Join(firstCycle, ",") {
				t.Errorf("non-deterministic cycle detection: run %d got %v, run 0 got %v", i, cycle, firstCycle)
			}
		}
	}
}
