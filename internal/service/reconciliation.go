package service

import (
	"fmt"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/reconcile"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// ReconciliationService orchestrates reconciliation workflows.
// Coordinates config loading, lock file I/O, engine execution, and lock persistence.
type ReconciliationService struct {
	configRepo repo.ConfigRepo
	docRepo    repo.DocRepo
	lockRepo   repo.LockRepo
}

// NewReconciliationService creates a ReconciliationService with injected dependencies.
func NewReconciliationService(configRepo repo.ConfigRepo, docRepo repo.DocRepo, lockRepo repo.LockRepo) *ReconciliationService {
	return &ReconciliationService{
		configRepo: configRepo,
		docRepo:    docRepo,
		lockRepo:   lockRepo,
	}
}

// Reconcile runs a full reconciliation: load config, load/create lock, run engine, persist lock.
func (s *ReconciliationService) Reconcile(projectRoot string, opts domain.ReconcileOpts) (*domain.ReconcileResult, error) {
	// Load config
	cfg, err := s.configRepo.ReadProjectConfig()
	if err != nil {
		return nil, fmt.Errorf("mind.toml required for reconciliation: %w", err)
	}

	// Validate config has documents
	if len(cfg.Documents) == 0 {
		return nil, fmt.Errorf("mind.toml required for reconciliation: no [documents] section")
	}

	// Load or create lock
	var lock *domain.LockFile
	if opts.Force {
		lock = nil // Engine treats nil as empty lock
	} else {
		lock, err = s.lockRepo.Read()
		if err != nil {
			return nil, fmt.Errorf("read mind.lock: %w", err)
		}
		// nil is fine -- first run
	}

	// Run engine
	engine := reconcile.NewEngine(s.docRepo)
	result, updatedLock, err := engine.Reconcile(projectRoot, cfg, lock, opts)
	if err != nil {
		return nil, err
	}

	// Persist lock (unless check-only mode)
	if !opts.CheckOnly {
		if err := s.lockRepo.Write(updatedLock); err != nil {
			return nil, fmt.Errorf("write mind.lock: %w", err)
		}
	}

	return result, nil
}

// LoadGraph loads the dependency graph from config and annotates with staleness from the lock file.
// Returns the graph and a map of stale document IDs to reasons (may be nil).
func (s *ReconciliationService) LoadGraph(projectRoot string) (*domain.Graph, map[string]string, error) {
	cfg, err := s.configRepo.ReadProjectConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("mind.toml required for reconciliation: %w", err)
	}

	graph := domain.BuildGraph(cfg.Graph)

	// Try to load staleness data from existing lock
	var stale map[string]string
	if s.lockRepo.Exists() {
		lock, err := s.lockRepo.Read()
		if err == nil && lock != nil {
			stale = make(map[string]string)
			for id, entry := range lock.Entries {
				if entry.Stale {
					stale[id] = entry.StaleReason
				}
			}
		}
	}

	return graph, stale, nil
}

// ReadStaleness reads existing lock data for the staleness panel (mind status).
// Returns nil when no lock file exists. Does NOT trigger reconciliation (FR-77).
func (s *ReconciliationService) ReadStaleness(projectRoot string) (*domain.StalenessInfo, error) {
	if !s.lockRepo.Exists() {
		return nil, nil
	}

	lock, err := s.lockRepo.Read()
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, nil
	}

	stale := make(map[string]string)
	for id, entry := range lock.Entries {
		if entry.Stale {
			stale[id] = entry.StaleReason
		}
	}

	return &domain.StalenessInfo{
		Status: lock.Status,
		Stale:  stale,
		Stats:  lock.Stats,
	}, nil
}
