package reconcile

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// FR-61: Edge type "informs" produces "may be outdated".
func TestPropagateDownstream_FR61_InformsReason(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	staleMap, _ := PropagateDownstream(g, []string{"A"}, map[string]bool{"A": true})

	reason := staleMap["B"]
	if !strings.Contains(reason, "may be outdated") {
		t.Errorf("informs reason = %q, want to contain 'may be outdated'", reason)
	}
}

// FR-61: Edge type "requires" produces "prerequisite changed".
func TestPropagateDownstream_FR61_RequiresReason(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeRequires},
	}
	g := domain.BuildGraph(edges)

	staleMap, _ := PropagateDownstream(g, []string{"A"}, map[string]bool{"A": true})

	reason := staleMap["B"]
	if !strings.Contains(reason, "prerequisite changed") {
		t.Errorf("requires reason = %q, want to contain 'prerequisite changed'", reason)
	}
}

// FR-61: Edge type "validates" produces "needs re-validation".
func TestPropagateDownstream_FR61_ValidatesReason(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeValidates},
	}
	g := domain.BuildGraph(edges)

	staleMap, _ := PropagateDownstream(g, []string{"A"}, map[string]bool{"A": true})

	reason := staleMap["B"]
	if !strings.Contains(reason, "needs re-validation") {
		t.Errorf("validates reason = %q, want to contain 'needs re-validation'", reason)
	}
}

// FR-67: Depth limit at exactly 10.
func TestPropagateDownstream_FR67_DepthLimitExact(t *testing.T) {
	// Build chain: n00 -> n01 -> ... -> n12 (13 nodes)
	// A change at n00 should propagate to n01..n10 (depth 0..9) but NOT to n11 (depth 10) or n12 (depth 11).
	names := make([]string, 13)
	for i := range names {
		names[i] = fmt.Sprintf("n%02d", i)
	}

	var edges []domain.GraphEdge
	for i := 0; i < 12; i++ {
		edges = append(edges, domain.GraphEdge{From: names[i], To: names[i+1], Type: domain.EdgeInforms})
	}
	g := domain.BuildGraph(edges)

	staleMap, warnings := PropagateDownstream(g, []string{names[0]}, map[string]bool{names[0]: true})

	// nodes[1] through nodes[10] should be stale (depth 0-9)
	for i := 1; i <= 10; i++ {
		if _, ok := staleMap[names[i]]; !ok {
			t.Errorf("%s (depth %d) should be stale", names[i], i-1)
		}
	}

	// nodes[11] should NOT be stale (depth 10, at the limit)
	if _, ok := staleMap[names[11]]; ok {
		t.Errorf("%s (depth 10) should NOT be stale (depth limit reached)", names[11])
	}

	// nodes[12] should NOT be stale (depth 11, beyond limit)
	if _, ok := staleMap[names[12]]; ok {
		t.Errorf("%s (depth 11) should NOT be stale (beyond depth limit)", names[12])
	}

	// Should have depth limit warnings
	if len(warnings) == 0 {
		t.Error("expected depth limit warnings")
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "depth limit") {
			found = true
		}
	}
	if !found {
		t.Errorf("warnings should mention depth limit: %v", warnings)
	}
}

// FR-67: No warning when chain is within limit.
func TestPropagateDownstream_FR67_NoWarningWithinLimit(t *testing.T) {
	// Chain of 5: within limit
	var edges []domain.GraphEdge
	for i := 0; i < 5; i++ {
		edges = append(edges, domain.GraphEdge{
			From: fmt.Sprintf("n%d", i),
			To:   fmt.Sprintf("n%d", i+1),
			Type: domain.EdgeInforms,
		})
	}
	g := domain.BuildGraph(edges)

	_, warnings := PropagateDownstream(g, []string{"n0"}, map[string]bool{"n0": true})
	if len(warnings) != 0 {
		t.Errorf("no depth limit warnings expected for short chain, got: %v", warnings)
	}
}

// Multiple changed sources propagate correctly.
func TestPropagateDownstream_MultipleChangedSources(t *testing.T) {
	// A -> C, B -> C; both A and B change
	edges := []domain.GraphEdge{
		{From: "A", To: "C", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeRequires},
	}
	g := domain.BuildGraph(edges)

	staleMap, _ := PropagateDownstream(g, []string{"A", "B"}, map[string]bool{"A": true, "B": true})

	// C should be stale (first path wins per FR-69)
	if _, ok := staleMap["C"]; !ok {
		t.Error("C should be stale")
	}
}

// No changed docs means no propagation.
func TestPropagateDownstream_NoChanges(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	staleMap, warnings := PropagateDownstream(g, []string{}, map[string]bool{})
	if len(staleMap) != 0 {
		t.Errorf("expected 0 stale with no changes, got %d", len(staleMap))
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings with no changes, got %d", len(warnings))
	}
}

// edgeTypeReason maps correctly.
func TestEdgeTypeReason(t *testing.T) {
	tests := []struct {
		edgeType domain.EdgeType
		expected string
	}{
		{domain.EdgeInforms, "may be outdated"},
		{domain.EdgeRequires, "prerequisite changed"},
		{domain.EdgeValidates, "needs re-validation"},
		{domain.EdgeType("unknown"), "may be outdated"}, // default case
	}
	for _, tt := range tests {
		got := edgeTypeReason(tt.edgeType)
		if got != tt.expected {
			t.Errorf("edgeTypeReason(%q) = %q, want %q", tt.edgeType, got, tt.expected)
		}
	}
}

// MaxPropagationDepth constant.
func TestMaxPropagationDepth(t *testing.T) {
	if MaxPropagationDepth != 10 {
		t.Errorf("MaxPropagationDepth = %d, want 10", MaxPropagationDepth)
	}
}
