package reconcile

import (
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestDetectCycle_Acyclic(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeInforms},
		{From: "A", To: "C", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle != nil {
		t.Errorf("expected no cycle, got %v", cycle)
	}
}

func TestDetectCycle_SimpleCycle(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeInforms},
		{From: "C", To: "A", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle == nil {
		t.Fatal("expected a cycle")
	}

	// Cycle should contain at least A, B, C
	str := strings.Join(cycle, " -> ")
	if !strings.Contains(str, "A") || !strings.Contains(str, "B") || !strings.Contains(str, "C") {
		t.Errorf("cycle path should contain A, B, C: %s", str)
	}

	// Cycle should start and end with the same node
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node: %v", cycle)
	}
}

func TestDetectCycle_SelfLoop(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "A", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle == nil {
		t.Fatal("expected a cycle for self-loop")
	}
}

func TestDetectCycle_DisconnectedWithCycle(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "C", To: "D", Type: domain.EdgeInforms},
		{From: "D", To: "C", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	cycle := DetectCycle(g)
	if cycle == nil {
		t.Fatal("expected a cycle in disconnected component")
	}

	str := strings.Join(cycle, " -> ")
	if !strings.Contains(str, "C") || !strings.Contains(str, "D") {
		t.Errorf("cycle should be in C-D component: %s", str)
	}
}

func TestDetectCycle_EmptyGraph(t *testing.T) {
	g := domain.BuildGraph(nil)
	cycle := DetectCycle(g)
	if cycle != nil {
		t.Errorf("expected no cycle for empty graph, got %v", cycle)
	}
}

func TestValidateEdges_AllDeclared(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/req", To: "doc:spec/arch", Type: domain.EdgeInforms},
		{From: "doc:spec/arch", To: "doc:spec/domain", Type: domain.EdgeInforms},
	}
	declared := map[string]bool{
		"doc:spec/req":    true,
		"doc:spec/arch":   true,
		"doc:spec/domain": true,
	}

	err := ValidateEdges(edges, declared)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateEdges_UndeclaredFrom(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/nonexistent", To: "doc:spec/arch", Type: domain.EdgeInforms},
	}
	declared := map[string]bool{
		"doc:spec/arch": true,
	}

	err := ValidateEdges(edges, declared)
	if err == nil {
		t.Fatal("expected error for undeclared from")
	}
	if !strings.Contains(err.Error(), "doc:spec/nonexistent") {
		t.Errorf("error should mention undeclared doc: %v", err)
	}
}

func TestValidateEdges_UndeclaredTo(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/req", To: "doc:spec/nonexistent", Type: domain.EdgeInforms},
	}
	declared := map[string]bool{
		"doc:spec/req": true,
	}

	err := ValidateEdges(edges, declared)
	if err == nil {
		t.Fatal("expected error for undeclared to")
	}
	if !strings.Contains(err.Error(), "doc:spec/nonexistent") {
		t.Errorf("error should mention undeclared doc: %v", err)
	}
}

func TestValidateEdges_EmptyEdges(t *testing.T) {
	err := ValidateEdges(nil, map[string]bool{"doc:spec/req": true})
	if err != nil {
		t.Errorf("expected no error for empty edges, got: %v", err)
	}
}
