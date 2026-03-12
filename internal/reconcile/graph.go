package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
)

// DetectCycle performs DFS on the graph and returns the cycle path if a cycle exists.
// Returns nil if the graph is acyclic.
func DetectCycle(g *domain.Graph) []string {
	visited := map[string]bool{}
	inStack := map[string]bool{}
	parent := map[string]string{}

	var cyclePath []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true

		for _, edge := range g.Forward[node] {
			if !visited[edge.To] {
				parent[edge.To] = node
				if dfs(edge.To) {
					return true
				}
			} else if inStack[edge.To] {
				// Build cycle path from edge.To back to edge.To
				cyclePath = []string{edge.To}
				cur := node
				for cur != edge.To {
					cyclePath = append([]string{cur}, cyclePath...)
					cur = parent[cur]
				}
				cyclePath = append([]string{edge.To}, cyclePath...)
				return true
			}
		}

		inStack[node] = false
		return false
	}

	// Sort nodes for deterministic traversal order
	nodes := make([]string, 0, len(g.Nodes))
	for n := range g.Nodes {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if !visited[node] {
			if dfs(node) {
				return cyclePath
			}
		}
	}
	return nil
}

// ValidateEdges checks that all document IDs in graph edges exist in the declared documents set.
// Returns an error listing all undeclared references.
func ValidateEdges(edges []domain.GraphEdge, declaredDocs map[string]bool) error {
	var undeclared []string
	seen := map[string]bool{}

	for _, edge := range edges {
		if !declaredDocs[edge.From] && !seen[edge.From] {
			undeclared = append(undeclared, edge.From)
			seen[edge.From] = true
		}
		if !declaredDocs[edge.To] && !seen[edge.To] {
			undeclared = append(undeclared, edge.To)
			seen[edge.To] = true
		}
	}

	if len(undeclared) > 0 {
		sort.Strings(undeclared)
		return fmt.Errorf("graph references undeclared document: %s", strings.Join(undeclared, ", "))
	}
	return nil
}
