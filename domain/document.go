package domain

import "time"

// DocStatus represents the status of a document.
type DocStatus string

const (
	DocDraft    DocStatus = "draft"
	DocActive   DocStatus = "active"
	DocComplete DocStatus = "complete"
	DocStub     DocStatus = "stub"
)

// Document represents a single documentation file.
type Document struct {
	Path    string    `json:"path"`
	AbsPath string    `json:"-"`
	Zone    Zone      `json:"zone"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsStub  bool      `json:"is_stub"`
	Status  DocStatus `json:"status"`
}

// BriefGate classifies the project brief for the business context gate.
type BriefGate string

const (
	BriefPresent BriefGate = "BRIEF_PRESENT"
	BriefStub    BriefGate = "BRIEF_STUB"
	BriefMissing BriefGate = "BRIEF_MISSING"
)

// Brief represents a parsed project brief with section detection.
type Brief struct {
	Path            string    `json:"path"`
	Exists          bool      `json:"exists"`
	IsStub          bool      `json:"is_stub"`
	HasVision       bool      `json:"has_vision"`
	HasDeliverables bool      `json:"has_deliverables"`
	HasScope        bool      `json:"has_scope"`
	GateResult      BriefGate `json:"gate"`
}
