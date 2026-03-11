package domain

import "time"

// EdgeType classifies the semantic relationship between two documents.
type EdgeType string

const (
	EdgeInforms   EdgeType = "informs"
	EdgeRequires  EdgeType = "requires"
	EdgeValidates EdgeType = "validates"
)

// ValidEdgeTypes contains all valid edge type values.
var ValidEdgeTypes = []EdgeType{EdgeInforms, EdgeRequires, EdgeValidates}

// ValidEdgeType returns true if the given string is a valid edge type.
func ValidEdgeType(s string) bool {
	for _, t := range ValidEdgeTypes {
		if string(t) == s {
			return true
		}
	}
	return false
}

// LockStatus represents the overall reconciliation state.
type LockStatus string

const (
	LockClean LockStatus = "CLEAN"
	LockStale LockStatus = "STALE"
	LockDirty LockStatus = "DIRTY"
)

// EntryStatus represents the state of a single document in a reconciliation run.
type EntryStatus string

const (
	EntryPresent   EntryStatus = "PRESENT"
	EntryMissing   EntryStatus = "MISSING"
	EntryChanged   EntryStatus = "CHANGED"
	EntryUnchanged EntryStatus = "UNCHANGED"
)

// GraphEdge is a directed dependency between two documents as declared in mind.toml [[graph]].
type GraphEdge struct {
	From string   `json:"from" toml:"from"`
	To   string   `json:"to" toml:"to"`
	Type EdgeType `json:"type" toml:"type"`
}

// Graph is a directed graph of document dependencies with forward and reverse adjacency lists.
type Graph struct {
	Forward map[string][]GraphEdge `json:"forward"`
	Reverse map[string][]GraphEdge `json:"reverse"`
	Nodes   map[string]bool        `json:"nodes"`
}

// BuildGraph constructs a directed graph from a slice of edges.
func BuildGraph(edges []GraphEdge) *Graph {
	g := &Graph{
		Forward: make(map[string][]GraphEdge),
		Reverse: make(map[string][]GraphEdge),
		Nodes:   make(map[string]bool),
	}
	for _, e := range edges {
		g.Forward[e.From] = append(g.Forward[e.From], e)
		g.Reverse[e.To] = append(g.Reverse[e.To], e)
		g.Nodes[e.From] = true
		g.Nodes[e.To] = true
	}
	return g
}

// LockStats holds aggregate counts from a reconciliation run.
type LockStats struct {
	Total      int `json:"total"`
	Changed    int `json:"changed"`
	Stale      int `json:"stale"`
	Missing    int `json:"missing"`
	Undeclared int `json:"undeclared"`
	Clean      int `json:"clean"`
}

// LockEntry tracks the reconciliation state of a single document.
type LockEntry struct {
	ID          string      `json:"id"`
	Path        string      `json:"path"`
	Hash        string      `json:"hash"`
	Size        int64       `json:"size"`
	ModTime     time.Time   `json:"mod_time"`
	Stale       bool        `json:"stale"`
	StaleReason string      `json:"stale_reason"`
	IsStub      bool        `json:"is_stub"`
	Status      EntryStatus `json:"status"`
}

// LockFile is the persisted reconciliation state (mind.lock).
type LockFile struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Status      LockStatus           `json:"status"`
	Stats       LockStats            `json:"stats"`
	Entries     map[string]LockEntry `json:"entries"`
}

// ReconcileResult is the ephemeral result of a reconciliation run.
type ReconcileResult struct {
	Changed    []string          `json:"changed"`
	Stale      map[string]string `json:"stale"`
	Missing    []string          `json:"missing"`
	Undeclared []string          `json:"undeclared"`
	Status     LockStatus        `json:"status"`
	Stats      LockStats         `json:"stats"`
	Warnings   []string          `json:"warnings,omitempty"`
}

// ReconcileOpts controls reconciliation behavior.
type ReconcileOpts struct {
	Force     bool `json:"force"`
	CheckOnly bool `json:"check_only"`
	GraphOnly bool `json:"graph_only"`
}

// StalenessInfo summarizes staleness for ProjectHealth integration.
type StalenessInfo struct {
	Status LockStatus        `json:"status"`
	Stale  map[string]string `json:"stale"`
	Stats  LockStats         `json:"stats"`
}
