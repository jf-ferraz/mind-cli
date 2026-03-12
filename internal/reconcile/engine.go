package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/repo"
)

// Engine orchestrates the 6-phase reconciliation algorithm.
type Engine struct {
	docRepo repo.DocRepo
}

// NewEngine creates a reconciliation engine.
func NewEngine(docRepo repo.DocRepo) *Engine {
	return &Engine{docRepo: docRepo}
}

// Reconcile runs the 6-phase reconciliation algorithm.
// Config and lock are passed by the caller (service handles I/O boundaries).
func (e *Engine) Reconcile(projectRoot string, cfg *domain.Config, lock *domain.LockFile, opts domain.ReconcileOpts) (*domain.ReconcileResult, *domain.LockFile, error) {
	// Ensure lock has an entries map
	if lock == nil {
		lock = &domain.LockFile{
			Entries: make(map[string]domain.LockEntry),
		}
	}
	if lock.Entries == nil {
		lock.Entries = make(map[string]domain.LockEntry)
	}

	// Phase 2: Build and validate graph
	graph := domain.BuildGraph(cfg.Graph)

	declaredDocs := buildDeclaredDocs(cfg)
	if len(cfg.Graph) > 0 {
		if err := ValidateEdges(cfg.Graph, declaredDocs); err != nil {
			return nil, nil, err
		}
	}

	cycle := DetectCycle(graph)
	if cycle != nil {
		return nil, nil, fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
	}

	// Phase 3: Scan filesystem and compute hashes
	var changed []string
	var missing []string
	var warnings []string

	for docID := range declaredDocs {
		entry := e.scanDocument(projectRoot, cfg, docID, lock, &warnings)
		lock.Entries[docID] = entry

		switch entry.Status {
		case domain.EntryMissing:
			missing = append(missing, docID)
		case domain.EntryChanged:
			changed = append(changed, docID)
		}
	}

	// Phase 4: Detect undeclared files
	undeclared := e.detectUndeclared(cfg, declaredDocs)

	// Phase 5: Propagate staleness
	changedSet := make(map[string]bool, len(changed))
	for _, id := range changed {
		changedSet[id] = true
	}

	staleMap, propWarnings := PropagateDownstream(graph, changed, changedSet)
	warnings = append(warnings, propWarnings...)

	// Apply staleness to lock entries
	for docID, reason := range staleMap {
		if entry, ok := lock.Entries[docID]; ok {
			entry.Stale = true
			entry.StaleReason = reason
			lock.Entries[docID] = entry
		}
	}

	// Clear staleness on changed documents (they are fresh per BR-28)
	for _, docID := range changed {
		if entry, ok := lock.Entries[docID]; ok {
			entry.Stale = false
			entry.StaleReason = ""
			lock.Entries[docID] = entry
		}
	}

	// Clear staleness on unchanged documents not in staleMap
	for docID, entry := range lock.Entries {
		if !changedSet[docID] && staleMap[docID] == "" {
			entry.Stale = false
			entry.StaleReason = ""
			lock.Entries[docID] = entry
		}
	}

	// Prune lock entries for documents no longer in config (XC-11)
	for docID := range lock.Entries {
		if !declaredDocs[docID] {
			delete(lock.Entries, docID)
		}
	}

	// Phase 6: Compute stats and finalize
	stats := domain.LockStats{
		Total:      len(declaredDocs),
		Changed:    len(changed),
		Stale:      len(staleMap),
		Missing:    len(missing),
		Undeclared: len(undeclared),
	}
	stats.Clean = stats.Total - stats.Changed - stats.Stale - stats.Missing

	var status domain.LockStatus
	switch {
	case len(staleMap) > 0:
		status = domain.LockStale
	case len(missing) > 0:
		status = domain.LockDirty
	default:
		status = domain.LockClean
	}

	lock.GeneratedAt = time.Now().UTC()
	lock.Status = status
	lock.Stats = stats

	result := &domain.ReconcileResult{
		Changed:    changed,
		Stale:      staleMap,
		Missing:    missing,
		Undeclared: undeclared,
		Status:     status,
		Stats:      stats,
		Warnings:   warnings,
	}

	// Ensure non-nil slices and maps for JSON output
	if result.Changed == nil {
		result.Changed = []string{}
	}
	if result.Stale == nil {
		result.Stale = map[string]string{}
	}
	if result.Missing == nil {
		result.Missing = []string{}
	}
	if result.Undeclared == nil {
		result.Undeclared = []string{}
	}

	return result, lock, nil
}

// scanDocument examines a single document and returns its lock entry.
func (e *Engine) scanDocument(projectRoot string, cfg *domain.Config, docID string, lock *domain.LockFile, warnings *[]string) domain.LockEntry {
	docPath := findDocPath(cfg, docID)
	absPath := filepath.Join(projectRoot, docPath)

	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		// Check if symlink target is outside project root
		if resolved != absPath {
			absRoot, _ := filepath.Abs(projectRoot)
			if !strings.HasPrefix(resolved, absRoot) {
				*warnings = append(*warnings, fmt.Sprintf("warning: symlink target outside project root: %s -> %s", docPath, resolved))
			}
		}
		absPath = resolved
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return domain.LockEntry{
			ID:     docID,
			Path:   docPath,
			Status: domain.EntryMissing,
		}
	}

	// Check for large files
	if stat.Size() > 10*1024*1024 {
		*warnings = append(*warnings, fmt.Sprintf("warning: large file (%d bytes): %s", stat.Size(), docPath))
	}

	// Stub detection via DocRepo (BR-34)
	isStub, _ := e.docRepo.IsStub(docPath)

	prevEntry := lock.Entries[docID]

	// Mtime fast-path (FR-58)
	if !NeedsRehash(&prevEntry, stat) {
		return domain.LockEntry{
			ID:      docID,
			Path:    docPath,
			Hash:    prevEntry.Hash,
			Size:    stat.Size(),
			ModTime: stat.ModTime(),
			IsStub:  isStub,
			Status:  domain.EntryUnchanged,
		}
	}

	// Compute hash
	hash, err := HashFile(absPath)
	if err != nil {
		*warnings = append(*warnings, fmt.Sprintf("warning: hash failed for %s: %v", docPath, err))
		return domain.LockEntry{
			ID:     docID,
			Path:   docPath,
			Status: domain.EntryMissing,
		}
	}

	// Binary detection warning
	if isBinaryFile(absPath) {
		*warnings = append(*warnings, fmt.Sprintf("warning: binary file detected: %s", docPath))
	}

	// Determine if content changed
	status := domain.EntryUnchanged
	if prevEntry.Hash == "" || hash != prevEntry.Hash {
		status = domain.EntryChanged
	}

	return domain.LockEntry{
		ID:      docID,
		Path:    docPath,
		Hash:    hash,
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		IsStub:  isStub,
		Status:  status,
	}
}

// detectUndeclared finds files in docs/ that are not declared in mind.toml.
func (e *Engine) detectUndeclared(cfg *domain.Config, declaredDocs map[string]bool) []string {
	allDocs, err := e.docRepo.ListAll()
	if err != nil {
		return nil
	}

	// Build set of declared paths
	declaredPaths := make(map[string]bool)
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			declaredPaths[entry.Path] = true
		}
	}

	var undeclared []string
	for _, doc := range allDocs {
		if !declaredPaths[doc.Path] {
			undeclared = append(undeclared, doc.Path)
		}
	}
	return undeclared
}

// buildDeclaredDocs creates a set of document IDs from the config.
func buildDeclaredDocs(cfg *domain.Config) map[string]bool {
	docs := make(map[string]bool)
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.ID != "" {
				docs[entry.ID] = true
			}
		}
	}
	return docs
}

// findDocPath looks up the path for a document ID in the config.
func findDocPath(cfg *domain.Config, docID string) string {
	for _, zone := range cfg.Documents {
		for _, entry := range zone {
			if entry.ID == docID {
				return entry.Path
			}
		}
	}
	return ""
}

// isBinaryFile checks if a file contains non-UTF8 content (simple heuristic).
func isBinaryFile(absPath string) bool {
	f, err := os.Open(absPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if n == 0 {
		return false
	}

	return !utf8.Valid(buf[:n])
}
