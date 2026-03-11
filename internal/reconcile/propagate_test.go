package reconcile

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestPropagateDownstream_LinearChain(t *testing.T) {
	// A --> B --> C; A changes
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "B", To: "C", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	changedIDs := []string{"A"}
	changedSet := map[string]bool{"A": true}

	staleMap, warnings := PropagateDownstream(g, changedIDs, changedSet)

	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %v", warnings)
	}

	if _, ok := staleMap["B"]; !ok {
		t.Error("B should be stale")
	}
	if _, ok := staleMap["C"]; !ok {
		t.Error("C should be stale (transitive)")
	}
	if _, ok := staleMap["A"]; ok {
		t.Error("A should NOT be stale (it changed)")
	}

	// B reason should reference A
	if !strings.Contains(staleMap["B"], "A") {
		t.Errorf("B reason should reference A: %q", staleMap["B"])
	}
	// C reason should mention transitive
	if !strings.Contains(staleMap["C"], "transitive") {
		t.Errorf("C reason should mention transitive: %q", staleMap["C"])
	}
}

func TestPropagateDownstream_DiamondGraph(t *testing.T) {
	// A --> B, A --> C, B --> D, C --> D; A changes
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
		{From: "A", To: "C", Type: domain.EdgeInforms},
		{From: "B", To: "D", Type: domain.EdgeInforms},
		{From: "C", To: "D", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	changedIDs := []string{"A"}
	changedSet := map[string]bool{"A": true}

	staleMap, _ := PropagateDownstream(g, changedIDs, changedSet)

	for _, id := range []string{"B", "C", "D"} {
		if _, ok := staleMap[id]; !ok {
			t.Errorf("%s should be stale", id)
		}
	}

	// D should be marked stale exactly once (FR-69)
	if staleMap["D"] == "" {
		t.Error("D should have a reason")
	}
}

func TestPropagateDownstream_ChangedNotStale(t *testing.T) {
	// A --> B; both A and B change (FR-68)
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	changedIDs := []string{"A", "B"}
	changedSet := map[string]bool{"A": true, "B": true}

	staleMap, _ := PropagateDownstream(g, changedIDs, changedSet)

	if _, ok := staleMap["A"]; ok {
		t.Error("A should NOT be stale (it changed)")
	}
	if _, ok := staleMap["B"]; ok {
		t.Error("B should NOT be stale (it changed itself)")
	}
}

func TestPropagateDownstream_NoGraph(t *testing.T) {
	g := domain.BuildGraph(nil)

	staleMap, warnings := PropagateDownstream(g, []string{"A"}, map[string]bool{"A": true})

	if len(staleMap) != 0 {
		t.Errorf("expected no stale docs, got %d", len(staleMap))
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %d", len(warnings))
	}
}

func TestPropagateDownstream_ReverseNotPropagated(t *testing.T) {
	// A --> B; B changes. A should NOT be stale (FR-65).
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	staleMap, _ := PropagateDownstream(g, []string{"B"}, map[string]bool{"B": true})

	if _, ok := staleMap["A"]; ok {
		t.Error("A should NOT be stale (reverse propagation)")
	}
}

func TestPropagateDownstream_EdgeTypeReasons(t *testing.T) {
	tests := []struct {
		edgeType       domain.EdgeType
		expectedReason string
	}{
		{domain.EdgeInforms, "may be outdated"},
		{domain.EdgeRequires, "prerequisite changed"},
		{domain.EdgeValidates, "needs re-validation"},
	}

	for _, tt := range tests {
		t.Run(string(tt.edgeType), func(t *testing.T) {
			edges := []domain.GraphEdge{
				{From: "A", To: "B", Type: tt.edgeType},
			}
			g := domain.BuildGraph(edges)

			staleMap, _ := PropagateDownstream(g, []string{"A"}, map[string]bool{"A": true})

			reason := staleMap["B"]
			if !strings.Contains(reason, tt.expectedReason) {
				t.Errorf("reason %q should contain %q", reason, tt.expectedReason)
			}
		})
	}
}

func TestPropagateDownstream_DepthLimit(t *testing.T) {
	// Build chain: n01 --> n02 --> ... --> n12 (11 edges, 12 nodes)
	names := make([]string, 13)
	for i := range names {
		names[i] = fmt.Sprintf("doc:spec/n%02d", i+1)
	}

	var edges []domain.GraphEdge
	for i := 0; i < 11; i++ {
		edges = append(edges, domain.GraphEdge{From: names[i], To: names[i+1], Type: domain.EdgeInforms})
	}
	g := domain.BuildGraph(edges)

	changedIDs := []string{names[0]}
	changedSet := map[string]bool{names[0]: true}

	staleMap, warnings := PropagateDownstream(g, changedIDs, changedSet)

	// names[1] through names[10] should be stale (depth 0-9)
	for i := 1; i <= 10; i++ {
		if _, ok := staleMap[names[i]]; !ok {
			t.Errorf("%s (depth %d) should be stale", names[i], i-1)
		}
	}

	// names[11] should NOT be stale (depth 10, at limit)
	if _, ok := staleMap[names[11]]; ok {
		t.Errorf("%s should NOT be stale (depth limit reached)", names[11])
	}

	if len(warnings) == 0 {
		t.Error("expected depth limit warning")
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
