package repo

import "github.com/jf-ferraz/mind-cli/domain"

// DocRepo reads and queries the 5-zone documentation structure.
type DocRepo interface {
	// ListByZone returns all documents in a zone.
	ListByZone(zone domain.Zone) ([]domain.Document, error)

	// ListAll returns every document across all zones.
	ListAll() ([]domain.Document, error)

	// Read returns the content of a document.
	Read(relPath string) ([]byte, error)

	// Exists checks if a path exists relative to project root.
	Exists(relPath string) bool

	// IsStub detects if a document is a stub (template-only content).
	IsStub(relPath string) (bool, error)

	// IsDir checks if a path is a directory.
	IsDir(relPath string) bool
}

// IterationRepo manages iteration folders.
type IterationRepo interface {
	// List returns all iterations, newest first.
	List() ([]domain.Iteration, error)

	// NextSeq returns the next available sequence number.
	NextSeq() (int, error)

	// Create creates an iteration folder from templates.
	Create(reqType domain.RequestType, descriptor string) (*domain.Iteration, error)
}

// StateRepo reads and writes workflow state.
type StateRepo interface {
	// ReadWorkflow parses docs/state/workflow.md into structured state.
	ReadWorkflow() (*domain.WorkflowState, error)
}

// ConfigRepo reads project and framework configuration.
type ConfigRepo interface {
	// ReadProjectConfig parses mind.toml.
	ReadProjectConfig() (*domain.Config, error)

	// WriteProjectConfig writes mind.toml.
	WriteProjectConfig(cfg *domain.Config) error
}

// BriefRepo handles project brief parsing and validation.
type BriefRepo interface {
	// ParseBrief reads and analyzes the project brief.
	ParseBrief() (*domain.Brief, error)
}
