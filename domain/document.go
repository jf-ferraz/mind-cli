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
	Path    string    // Relative to project root
	AbsPath string    // Absolute path
	Zone    Zone      // Which zone it belongs to
	Name    string    // Filename without extension
	Size    int64     // Bytes
	ModTime time.Time // Last modification time
	IsStub  bool      // Detected by stub analysis
	Status  DocStatus // Inferred or from mind.toml
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
	Path            string
	Exists          bool
	IsStub          bool
	HasVision       bool
	HasDeliverables bool
	HasScope        bool
	GateResult      BriefGate
}
