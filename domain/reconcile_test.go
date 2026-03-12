package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

func TestBuildGraph_Empty(t *testing.T) {
	g := domain.BuildGraph(nil)
	if len(g.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes))
	}
	if len(g.Forward) != 0 {
		t.Errorf("expected 0 forward edges, got %d", len(g.Forward))
	}
}

func TestBuildGraph_BasicEdges(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "doc:spec/requirements", To: "doc:spec/architecture", Type: domain.EdgeInforms},
		{From: "doc:spec/architecture", To: "doc:spec/domain-model", Type: domain.EdgeInforms},
		{From: "doc:spec/requirements", To: "doc:spec/domain-model", Type: domain.EdgeInforms},
	}

	g := domain.BuildGraph(edges)

	// 3 unique nodes
	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}

	// Forward edges
	if len(g.Forward["doc:spec/requirements"]) != 2 {
		t.Errorf("expected 2 forward edges from requirements, got %d", len(g.Forward["doc:spec/requirements"]))
	}
	if len(g.Forward["doc:spec/architecture"]) != 1 {
		t.Errorf("expected 1 forward edge from architecture, got %d", len(g.Forward["doc:spec/architecture"]))
	}

	// Reverse edges
	if len(g.Reverse["doc:spec/architecture"]) != 1 {
		t.Errorf("expected 1 reverse edge to architecture, got %d", len(g.Reverse["doc:spec/architecture"]))
	}
	if len(g.Reverse["doc:spec/domain-model"]) != 2 {
		t.Errorf("expected 2 reverse edges to domain-model, got %d", len(g.Reverse["doc:spec/domain-model"]))
	}
}

func TestEdgeType_Values(t *testing.T) {
	if string(domain.EdgeInforms) != "informs" {
		t.Errorf("EdgeInforms = %q", domain.EdgeInforms)
	}
	if string(domain.EdgeRequires) != "requires" {
		t.Errorf("EdgeRequires = %q", domain.EdgeRequires)
	}
	if string(domain.EdgeValidates) != "validates" {
		t.Errorf("EdgeValidates = %q", domain.EdgeValidates)
	}
}

func TestValidEdgeType(t *testing.T) {
	for _, et := range []string{"informs", "requires", "validates"} {
		if !domain.ValidEdgeType(et) {
			t.Errorf("expected %q to be valid", et)
		}
	}
	for _, bad := range []string{"", "depends", "blocks", "INFORMS"} {
		if domain.ValidEdgeType(bad) {
			t.Errorf("expected %q to be invalid", bad)
		}
	}
}

func TestLockStatus_Values(t *testing.T) {
	if string(domain.LockClean) != "CLEAN" {
		t.Errorf("LockClean = %q", domain.LockClean)
	}
	if string(domain.LockStale) != "STALE" {
		t.Errorf("LockStale = %q", domain.LockStale)
	}
	if string(domain.LockDirty) != "DIRTY" {
		t.Errorf("LockDirty = %q", domain.LockDirty)
	}
}

func TestEntryStatus_Values(t *testing.T) {
	if string(domain.EntryPresent) != "PRESENT" {
		t.Errorf("EntryPresent = %q", domain.EntryPresent)
	}
	if string(domain.EntryMissing) != "MISSING" {
		t.Errorf("EntryMissing = %q", domain.EntryMissing)
	}
	if string(domain.EntryChanged) != "CHANGED" {
		t.Errorf("EntryChanged = %q", domain.EntryChanged)
	}
	if string(domain.EntryUnchanged) != "UNCHANGED" {
		t.Errorf("EntryUnchanged = %q", domain.EntryUnchanged)
	}
}

func TestReconcileResult_JSONRoundTrip(t *testing.T) {
	result := domain.ReconcileResult{
		Changed:    []string{"doc:spec/requirements"},
		Stale:      map[string]string{"doc:spec/architecture": "dependency changed: doc:spec/requirements"},
		Missing:    []string{},
		Undeclared: []string{"docs/spec/notes.md"},
		Status:     domain.LockStale,
		Stats: domain.LockStats{
			Total:      5,
			Changed:    1,
			Stale:      1,
			Missing:    0,
			Undeclared: 1,
			Clean:      3,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded domain.ReconcileResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Status != domain.LockStale {
		t.Errorf("status = %q, want STALE", decoded.Status)
	}
	if len(decoded.Changed) != 1 {
		t.Errorf("changed = %d, want 1", len(decoded.Changed))
	}
	if decoded.Stats.Clean != 3 {
		t.Errorf("stats.clean = %d, want 3", decoded.Stats.Clean)
	}
}

func TestStalenessInfo_NilJSON(t *testing.T) {
	health := domain.ProjectHealth{
		Staleness: nil,
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// staleness should be null in JSON
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(raw["staleness"]) != "null" {
		t.Errorf("staleness = %s, want null", string(raw["staleness"]))
	}
}
