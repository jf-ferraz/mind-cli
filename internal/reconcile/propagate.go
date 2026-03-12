package reconcile

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
)

// MaxPropagationDepth limits BFS staleness traversal to prevent runaway propagation.
const MaxPropagationDepth = 10

// edgeTypeReason maps edge types to human-readable staleness reason fragments.
func edgeTypeReason(edgeType domain.EdgeType) string {
	switch edgeType {
	case domain.EdgeRequires:
		return "prerequisite changed"
	case domain.EdgeValidates:
		return "needs re-validation"
	case domain.EdgeInforms:
		return "may be outdated"
	default:
		return "may be outdated"
	}
}

// PropagateDownstream performs BFS staleness propagation through the graph.
// Starting from changedIDs, it marks downstream dependents as stale with
// edge-type-specific reason messages. Documents in changedSet are skipped
// (they are fresh, not stale). Returns a stale map and any depth-limit warnings.
func PropagateDownstream(graph *domain.Graph, changedIDs []string, changedSet map[string]bool) (staleMap map[string]string, warnings []string) {
	staleMap = make(map[string]string)

	type queueItem struct {
		nodeID   string
		sourceID string
		depth    int
		edgeType domain.EdgeType // edge type from the immediate predecessor
	}

	queue := make([]queueItem, 0, len(changedIDs)*2)

	// Seed the BFS queue with direct dependents of changed documents
	for _, changedID := range changedIDs {
		for _, edge := range graph.Forward[changedID] {
			queue = append(queue, queueItem{
				nodeID:   edge.To,
				sourceID: changedID,
				depth:    0,
				edgeType: edge.Type,
			})
		}
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= MaxPropagationDepth {
			warnings = append(warnings, fmt.Sprintf(
				"staleness propagation depth limit (%d) reached at %s",
				MaxPropagationDepth, item.nodeID,
			))
			continue
		}

		// Skip documents that changed themselves (they are fresh)
		if changedSet[item.nodeID] {
			continue
		}

		// Skip documents already marked stale (first path wins per FR-69)
		if _, exists := staleMap[item.nodeID]; exists {
			continue
		}

		// Build reason using the edge type carried on the queue item
		reason := buildReason(item.sourceID, item.edgeType, item.depth)
		staleMap[item.nodeID] = reason

		// Enqueue downstream dependents for transitive propagation
		for _, edge := range graph.Forward[item.nodeID] {
			queue = append(queue, queueItem{
				nodeID:   edge.To,
				sourceID: item.sourceID,
				depth:    item.depth + 1,
				edgeType: edge.Type,
			})
		}
	}

	return staleMap, warnings
}

// buildReason constructs a staleness reason string using the edge type from the immediate predecessor.
func buildReason(sourceID string, edgeType domain.EdgeType, depth int) string {
	reason := edgeTypeReason(edgeType)
	if depth == 0 {
		return fmt.Sprintf("dependency changed: %s (%s)", sourceID, reason)
	}
	return fmt.Sprintf("dependency changed: %s (via transitive chain, %s)", sourceID, reason)
}
