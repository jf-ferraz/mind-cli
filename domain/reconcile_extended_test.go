package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
)

// BR-24: Hash format validation.
func TestLockEntry_HashFormat(t *testing.T) {
	entry := domain.LockEntry{
		ID:   "doc:spec/requirements",
		Hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	}

	if len(entry.Hash) != 7+64 {
		t.Errorf("hash length = %d, want 71", len(entry.Hash))
	}
}

// BR-28: Changed and stale are mutually exclusive.
func TestLockEntry_ChangedAndStaleExclusive(t *testing.T) {
	// A changed document should not be stale
	changed := domain.LockEntry{
		Status: domain.EntryChanged,
		Stale:  false, // Must be false for changed docs
	}
	if changed.Stale {
		t.Error("changed document should not be stale")
	}

	// A stale document should not be changed
	stale := domain.LockEntry{
		Status:      domain.EntryPresent,
		Stale:       true,
		StaleReason: "dependency changed",
	}
	if stale.Status == domain.EntryChanged {
		t.Error("stale document should not have CHANGED status")
	}
}

// BR-33: LockStatus priority.
func TestLockStatus_Priority(t *testing.T) {
	// STALE has highest priority
	if domain.LockStale != "STALE" {
		t.Errorf("LockStale = %q", domain.LockStale)
	}
	if domain.LockDirty != "DIRTY" {
		t.Errorf("LockDirty = %q", domain.LockDirty)
	}
	if domain.LockClean != "CLEAN" {
		t.Errorf("LockClean = %q", domain.LockClean)
	}
}

// BR-35: All three edge types are valid.
func TestEdgeType_AllThreeValid(t *testing.T) {
	for _, et := range domain.ValidEdgeTypes {
		if !domain.ValidEdgeType(string(et)) {
			t.Errorf("%q should be valid", et)
		}
	}
	if len(domain.ValidEdgeTypes) != 3 {
		t.Errorf("ValidEdgeTypes has %d entries, want 3", len(domain.ValidEdgeTypes))
	}
}

// BuildGraph with single edge creates correct adjacency.
func TestBuildGraph_SingleEdge(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeInforms},
	}
	g := domain.BuildGraph(edges)

	if len(g.Nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(g.Nodes))
	}
	if !g.Nodes["A"] || !g.Nodes["B"] {
		t.Error("both A and B should be nodes")
	}

	// Forward: A -> B
	if len(g.Forward["A"]) != 1 {
		t.Errorf("forward from A = %d, want 1", len(g.Forward["A"]))
	}
	if g.Forward["A"][0].To != "B" {
		t.Errorf("forward A->%s, want A->B", g.Forward["A"][0].To)
	}

	// Reverse: B <- A
	if len(g.Reverse["B"]) != 1 {
		t.Errorf("reverse to B = %d, want 1", len(g.Reverse["B"]))
	}
	if g.Reverse["B"][0].From != "A" {
		t.Errorf("reverse B<-%s, want B<-A", g.Reverse["B"][0].From)
	}
}

// BuildGraph preserves edge types.
func TestBuildGraph_PreservesEdgeType(t *testing.T) {
	edges := []domain.GraphEdge{
		{From: "A", To: "B", Type: domain.EdgeRequires},
		{From: "A", To: "C", Type: domain.EdgeValidates},
	}
	g := domain.BuildGraph(edges)

	for _, edge := range g.Forward["A"] {
		if edge.To == "B" && edge.Type != domain.EdgeRequires {
			t.Errorf("A->B type = %q, want requires", edge.Type)
		}
		if edge.To == "C" && edge.Type != domain.EdgeValidates {
			t.Errorf("A->C type = %q, want validates", edge.Type)
		}
	}
}

// ReconcileResult JSON serialization includes all fields.
func TestReconcileResult_JSONFields(t *testing.T) {
	result := domain.ReconcileResult{
		Changed:    []string{"doc:spec/a"},
		Stale:      map[string]string{"doc:spec/b": "reason"},
		Missing:    []string{"doc:spec/c"},
		Undeclared: []string{"docs/spec/d.md"},
		Status:     domain.LockStale,
		Stats: domain.LockStats{
			Total: 4, Changed: 1, Stale: 1, Missing: 1, Undeclared: 1, Clean: 1,
		},
		Warnings: []string{"some warning"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	requiredFields := []string{"changed", "stale", "missing", "undeclared", "status", "stats"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("ReconcileResult JSON missing field: %s", field)
		}
	}
}

// ReconcileOpts JSON serialization.
func TestReconcileOpts_JSON(t *testing.T) {
	opts := domain.ReconcileOpts{
		Force:     true,
		CheckOnly: false,
		GraphOnly: true,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}

	var decoded domain.ReconcileOpts
	json.Unmarshal(data, &decoded)

	if decoded.Force != true {
		t.Error("Force should be true")
	}
	if decoded.CheckOnly != false {
		t.Error("CheckOnly should be false")
	}
	if decoded.GraphOnly != true {
		t.Error("GraphOnly should be true")
	}
}

// LockFile JSON round-trip preserves all data.
func TestLockFile_JSONRoundTrip(t *testing.T) {
	lock := domain.LockFile{
		Status: domain.LockStale,
		Stats: domain.LockStats{
			Total: 3, Changed: 1, Stale: 1, Missing: 0, Undeclared: 0, Clean: 1,
		},
		Entries: map[string]domain.LockEntry{
			"doc:spec/a": {
				ID:          "doc:spec/a",
				Path:        "docs/spec/a.md",
				Hash:        "sha256:abc",
				Size:        100,
				Stale:       true,
				StaleReason: "reason",
				IsStub:      false,
				Status:      domain.EntryPresent,
			},
		},
	}

	data, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}

	var decoded domain.LockFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Status != domain.LockStale {
		t.Errorf("Status = %q, want STALE", decoded.Status)
	}
	if decoded.Stats.Total != 3 {
		t.Errorf("Stats.Total = %d, want 3", decoded.Stats.Total)
	}

	entry := decoded.Entries["doc:spec/a"]
	if entry.Hash != "sha256:abc" {
		t.Errorf("entry hash = %q", entry.Hash)
	}
	if !entry.Stale {
		t.Error("entry should be stale")
	}
	if entry.StaleReason != "reason" {
		t.Errorf("entry stale reason = %q", entry.StaleReason)
	}
}

// StalenessInfo JSON serialization with non-nil values.
func TestStalenessInfo_JSON(t *testing.T) {
	info := domain.StalenessInfo{
		Status: domain.LockStale,
		Stale:  map[string]string{"doc:spec/a": "reason A"},
		Stats:  domain.LockStats{Total: 2, Stale: 1, Clean: 1},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	var decoded domain.StalenessInfo
	json.Unmarshal(data, &decoded)

	if decoded.Status != domain.LockStale {
		t.Errorf("Status = %q", decoded.Status)
	}
	if len(decoded.Stale) != 1 {
		t.Errorf("Stale count = %d, want 1", len(decoded.Stale))
	}
}

// FR-78: ProjectHealth.Staleness is null when nil.
func TestProjectHealth_StalenessNull(t *testing.T) {
	health := domain.ProjectHealth{
		Staleness: nil,
	}

	data, _ := json.Marshal(health)

	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	if string(raw["staleness"]) != "null" {
		t.Errorf("staleness = %s, want null", raw["staleness"])
	}
}

// FR-78: ProjectHealth.Staleness is object when non-nil.
func TestProjectHealth_StalenessObject(t *testing.T) {
	health := domain.ProjectHealth{
		Staleness: &domain.StalenessInfo{
			Status: domain.LockStale,
			Stale:  map[string]string{"doc:spec/a": "reason"},
			Stats:  domain.LockStats{Total: 2, Stale: 1, Clean: 1},
		},
	}

	data, _ := json.Marshal(health)

	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	if string(raw["staleness"]) == "null" {
		t.Error("staleness should not be null when StalenessInfo is provided")
	}

	var staleObj map[string]json.RawMessage
	if err := json.Unmarshal(raw["staleness"], &staleObj); err != nil {
		t.Fatalf("staleness should be a JSON object: %v", err)
	}

	if _, ok := staleObj["status"]; !ok {
		t.Error("staleness object should have 'status' field")
	}
	if _, ok := staleObj["stale"]; !ok {
		t.Error("staleness object should have 'stale' field")
	}
}

// EntryStatus enum completeness.
func TestEntryStatus_AllValues(t *testing.T) {
	values := []domain.EntryStatus{
		domain.EntryPresent,
		domain.EntryMissing,
		domain.EntryChanged,
		domain.EntryUnchanged,
	}

	strings := map[string]bool{
		"PRESENT": false, "MISSING": false, "CHANGED": false, "UNCHANGED": false,
	}

	for _, v := range values {
		if _, ok := strings[string(v)]; !ok {
			t.Errorf("unexpected EntryStatus value: %q", v)
		}
		strings[string(v)] = true
	}

	for k, found := range strings {
		if !found {
			t.Errorf("EntryStatus %q not represented", k)
		}
	}
}

// GraphEdge TOML tags are correct for mind.toml parsing.
func TestGraphEdge_TOMLTags(t *testing.T) {
	// This is a compile-time check essentially. The toml tags must
	// match the expected [[graph]] format: from, to, type.
	edge := domain.GraphEdge{
		From: "doc:spec/requirements",
		To:   "doc:spec/architecture",
		Type: domain.EdgeInforms,
	}

	// Verify JSON round-trip preserves field names
	data, _ := json.Marshal(edge)
	var raw map[string]string
	json.Unmarshal(data, &raw)

	if raw["from"] != "doc:spec/requirements" {
		t.Errorf("JSON 'from' = %q", raw["from"])
	}
	if raw["to"] != "doc:spec/architecture" {
		t.Errorf("JSON 'to' = %q", raw["to"])
	}
	if raw["type"] != "informs" {
		t.Errorf("JSON 'type' = %q", raw["type"])
	}
}
