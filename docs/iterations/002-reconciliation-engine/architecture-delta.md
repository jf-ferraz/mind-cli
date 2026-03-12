# Architecture Delta: Phase 1.5 Reconciliation Engine

**Iteration**: 002-reconciliation-engine
**Date**: 2026-03-11
**Phase**: 1.5
**Type**: ENHANCEMENT (structural) -- new subsystem added to existing 4-layer architecture

---

## Current Structure

Phase 1 delivers a 4-layer architecture (Presentation, Service, Domain, Infrastructure) with downward-only dependency flow. The domain layer is pure Go (zero external imports). All filesystem access passes through repository interfaces.

Relevant portions for Phase 1.5:

```
cmd/status.go ──> service.ProjectService ──> repo.DocRepo, repo.StateRepo, repo.BriefRepo
cmd/check.go  ──> service.ValidationService ──> validate.DocsSuite/RefsSuite/ConfigSuite
cmd/doctor.go ──> service.DoctorService ──> repo.DocRepo, repo.IterRepo, repo.BriefRepo
domain/       ──> pure types: Project, Config, Document, ValidationReport, CheckResult
internal/repo/interfaces.go ──> DocRepo, IterationRepo, StateRepo, ConfigRepo, BriefRepo
internal/validate/check.go ──> Suite, Check, CheckFunc, CheckContext framework
```

**Current wiring pattern**: Each command handler in `cmd/` creates its own repository instances inline (e.g., `fs.NewDocRepo(root)`, `fs.NewConfigRepo(root)`). Services receive repos through constructors, but the construction happens redundantly in every command handler. This is tech debt identified in Phase 1 and targeted for cleanup in this phase.

**Current `Config` struct**: Has no `Graph` field. `mind.toml` parsing does not handle `[[graph]]` sections.

**Current `ProjectHealth`**: Has no staleness data. The `mind status` command has no awareness of `mind.lock`.

**Current validation suites**: Three suites (docs, refs, config) in `internal/validate/`. `ValidationService.RunAll()` composes them into a `UnifiedValidationReport`.

---

## Proposed Changes

### New Components

#### `domain/reconcile.go` -- Reconciliation Domain Types

**Responsibility**: Define all pure domain types for the reconciliation subsystem. Zero external imports (NFR-4 compliance).

**Types defined**:
- `LockFile` -- persisted reconciliation state with entries, stats, status, timestamp
- `LockEntry` -- per-document tracking entry (hash, size, mod_time, stale flag, stub flag)
- `ReconcileResult` -- ephemeral result from a reconciliation run (changed, stale, missing, undeclared)
- `GraphEdge` -- directed dependency between two document IDs with an edge type
- `EdgeType` -- enum: `informs`, `requires`, `validates`
- `LockStatus` -- enum: `CLEAN`, `STALE`, `DIRTY`
- `EntryStatus` -- enum: `PRESENT`, `MISSING`, `CHANGED`, `UNCHANGED`
- `LockStats` -- aggregate counts (total, changed, stale, missing, undeclared, clean)
- `ReconcileOpts` -- options struct for reconciliation (Force, CheckOnly, GraphOnly)
- `Graph` -- directed adjacency list (Forward, Reverse, Nodes) with `BuildGraph()` constructor
- `StalenessInfo` -- summary for `ProjectHealth` integration (status, stale map, stats)

**Integration points**:
- `Config.Graph` field references `[]GraphEdge`
- `ProjectHealth.Staleness` field references `*StalenessInfo`
- JSON struct tags on all types for `--json` serialization

**Key design note**: `BuildGraph()` is a pure function (takes `[]GraphEdge`, returns `*Graph`). It lives in the domain layer because it operates only on domain types with no I/O. This follows the DC-4 pattern established by `Slugify()` and `Classify()`.

---

#### `internal/reconcile/hash.go` -- Hash Computation

**Responsibility**: SHA-256 computation of file content with mtime fast-path optimization and edge case handling.

**Public API**:
```
HashFile(absPath string) (string, error)
NeedsRehash(entry *domain.LockEntry, info os.FileInfo) bool
```

**Integration points**:
- Called by `engine.go` during the scan phase
- Returns hash strings in `sha256:{hex}` format
- Accepts absolute file paths as parameters

**I/O justification**: This file performs direct `os.Open()` for streaming hash computation. This is architecturally justified as infrastructure-adjacent code per convergence recommendation 1.3. The alternative -- routing through `DocRepo.Read()` -- would require loading entire file contents into memory, defeating the purpose of streaming hash computation. The function accepts an absolute path as a parameter, making it a pure utility that does not depend on project structure.

---

#### `internal/reconcile/graph.go` -- Graph Operations

**Responsibility**: Cycle detection via DFS on the domain `Graph` type. The `Graph` construction itself (`BuildGraph()`) lives in `domain/reconcile.go`.

**Public API**:
```
DetectCycle(g *domain.Graph) []string   // returns cycle path or nil
ValidateEdges(edges []domain.GraphEdge, declaredDocs map[string]bool) error
```

**Integration points**:
- Called by `engine.go` during the graph validation phase
- `DetectCycle` returns the cycle path for error reporting
- `ValidateEdges` checks that all `from`/`to` IDs exist in the document registry (XC-10)

**Design note**: `BuildGraph()` is in `domain/` because it is a pure transformation (slice of edges to adjacency list). `DetectCycle()` is in `internal/reconcile/` because it is algorithmic logic that could hypothetically require non-domain imports in the future, and because it is not a business rule but an operational check. `ValidateEdges()` is here because it checks edge IDs against a set, which is validation logic not domain logic.

---

#### `internal/reconcile/propagate.go` -- Staleness Propagation

**Responsibility**: BFS downstream staleness propagation with depth limit of 10. Applies edge-type-specific reason messages.

**Public API**:
```
PropagateDownstream(graph *domain.Graph, changedIDs []string, changedSet map[string]bool) (staleMap map[string]string, warnings []string)
```

**Integration points**:
- Called by `engine.go` after change detection
- Returns a stale map (document ID to reason string) and any depth-limit warnings
- Reason strings include edge type semantics: "may be outdated" (informs), "prerequisite changed" (requires), "needs re-validation" (validates)

**Design note**: Returns the stale map rather than mutating lock entries, keeping propagation as a pure computation over graph structure. The engine is responsible for applying stale state to lock entries.

---

#### `internal/reconcile/engine.go` -- Reconciliation Orchestrator

**Responsibility**: Full 6-phase orchestration of a reconciliation run as defined in BP-06 Section 6.

**Public API**:
```
type Engine struct { ... }
func NewEngine(docRepo repo.DocRepo) *Engine
func (e *Engine) Reconcile(projectRoot string, cfg *domain.Config, lock *domain.LockFile, opts domain.ReconcileOpts) (*domain.ReconcileResult, *domain.LockFile, error)
```

**Six phases**:
1. **Load** -- receive config and lock file (loading is done by caller)
2. **Graph** -- build graph from config edges, validate edges, detect cycles
3. **Scan** -- stat files, mtime fast-path, hash changed files, detect missing
4. **Detect undeclared** -- scan `docs/` for files not in config (uses `DocRepo.ListAll()`)
5. **Propagate** -- BFS from changed documents through graph
6. **Report** -- compute stats, determine overall status, return result and updated lock

**Dependencies**: `repo.DocRepo` (for `IsStub()` and `ListAll()`). The engine does not depend on `LockRepo` -- loading and saving the lock file is the service layer's responsibility. The engine does not depend on `ConfigRepo` -- config is passed as a parameter.

**Integration points**:
- Called by `ReconciliationService`
- Returns both the `ReconcileResult` (for rendering) and the updated `LockFile` (for persistence)

---

#### `internal/repo/fs/lock_repo.go` -- Filesystem Lock Repository

**Responsibility**: Read and write `mind.lock` as JSON with atomic writes.

**Implements**: `repo.LockRepo` interface

**Public API**:
```
type LockRepo struct { ... }
func NewLockRepo(projectRoot string) *LockRepo
func (r *LockRepo) Read() (*domain.LockFile, error)    // returns nil, nil if file not found
func (r *LockRepo) Write(lock *domain.LockFile) error   // atomic: write to .tmp, rename
func (r *LockRepo) Exists() bool
```

**Atomic write implementation**: Write to `mind.lock.tmp`, then `os.Rename()` to `mind.lock`. Per FR-73, this prevents corrupted lock files from partial writes.

**Round-trip correctness**: JSON serialization uses `json.MarshalIndent` with consistent field ordering via struct tags and `json.Encoder` settings to ensure byte-identical round-trips (FR-72). Map iteration order is handled by sorting entry keys before serialization.

---

#### `internal/repo/mem/lock_repo.go` -- In-Memory Lock Repository

**Responsibility**: Test-only `LockRepo` implementation backed by an in-memory `*domain.LockFile`.

**Implements**: `repo.LockRepo` interface

**Design**: Stores a deep copy of the lock file to prevent test mutation issues.

---

#### `internal/service/reconciliation.go` -- ReconciliationService

**Responsibility**: Orchestrate reconciliation workflows. Coordinates config loading, lock file I/O, engine execution, and lock file persistence.

**Public API**:
```
type ReconciliationService struct { ... }
func NewReconciliationService(configRepo repo.ConfigRepo, docRepo repo.DocRepo, lockRepo repo.LockRepo) *ReconciliationService
func (s *ReconciliationService) Reconcile(projectRoot string, opts domain.ReconcileOpts) (*domain.ReconcileResult, error)
func (s *ReconciliationService) ReadStaleness(projectRoot string) (*domain.StalenessInfo, error)
```

**Reconcile flow**:
1. Load config via `ConfigRepo`
2. Validate config has `Documents` section (exit 3 if not)
3. Load lock via `LockRepo` (nil if not found = first run)
4. If `opts.Force`: create empty lock
5. Create engine, call `engine.Reconcile()`
6. If not `opts.CheckOnly`: write lock via `LockRepo`
7. Return result

**ReadStaleness flow** (used by `mind status`):
1. Check if lock exists via `LockRepo.Exists()`
2. If not: return nil (no staleness panel)
3. Load lock via `LockRepo.Read()`
4. Return `StalenessInfo` extracted from lock

**Dependencies**: `repo.ConfigRepo`, `repo.DocRepo`, `repo.LockRepo`

**Integration points**:
- Called by `cmd/reconcile.go` for the `mind reconcile` command
- `ReadStaleness()` called by `cmd/status.go` for the staleness panel
- Results consumed by `ReconcileSuite` for `mind check all` integration
- Results consumed by `DoctorService` for stale findings

---

#### `internal/validate/reconcile.go` -- ReconcileSuite

**Responsibility**: Project reconciliation results into the `Suite`/`Check` framework for `mind check all` integration.

**Public API**:
```
func ReconcileSuite(result *domain.ReconcileResult) *Suite
```

**Design**: Unlike `DocsSuite()`, `RefsSuite()`, and `ConfigSuite()` which are stateless (they run checks by querying repos), `ReconcileSuite()` takes a pre-computed `ReconcileResult` and projects it into `Check` entries. This is because running reconciliation is expensive and the result is needed elsewhere (rendering, exit code determination). The suite is constructed from the result rather than re-running reconciliation.

**Check projection**:
- Check 1: "No circular dependencies" -- FAIL level, passes unless cycle was detected (cycle detection happens in the engine; if we reach suite construction, no cycle exists, so this always passes in the projected suite)
- Check 2: "No missing documents" -- WARN per missing document (one check per missing doc)
- Check N+2: "Document not stale: {doc ID}" -- WARN per stale document (FAIL with `--strict`)

**Integration**: `ValidationService.RunAll()` calls `ReconcileSuite()` and appends the result to the unified report.

---

#### `cmd/reconcile.go` -- Reconcile Command

**Responsibility**: Parse flags, call `ReconciliationService`, render results, set exit code.

**Flags**:
- `--check` (bool): read-only verification, exit 0 or 4
- `--force` (bool): discard lock, re-hash everything
- `--graph` (bool): ASCII tree visualization of dependency graph

**Exit codes**: 0 (clean), 2 (cycle or runtime error), 3 (no config), 4 (stale, `--check` only)

**Integration**: Follows the thin-handler pattern established by all Phase 1 commands. Creates service, calls it, renders result, sets exit code.

---

### Modified Components

#### `domain/project.go` -- Config Extension

**Change**: Add `Graph []GraphEdge` field to the `Config` struct.

```go
type Config struct {
    Manifest   Manifest                       `toml:"manifest"`
    Project    ProjectMeta                    `toml:"project"`
    Profiles   Profiles                       `toml:"profiles"`
    Documents  map[string]map[string]DocEntry `toml:"documents"`
    Governance Governance                     `toml:"governance"`
    Graph      []GraphEdge                    `toml:"graph"`   // NEW
}
```

**Rationale**: `GraphEdge` is a domain type defined in `domain/reconcile.go`. Adding it to `Config` is a one-field addition. The TOML tag `"graph"` maps to the `[[graph]]` array-of-tables in `mind.toml`. `go-toml/v2` handles array-of-tables deserialization automatically when the struct field is a slice.

**Impact**: `ConfigRepo.ReadProjectConfig()` automatically picks up `[[graph]]` entries because `go-toml/v2` deserializes based on struct tags. No changes needed to `config_repo.go` for basic parsing. Any `[[graph]]` entries in `mind.toml` will populate `Config.Graph`.

---

#### `domain/health.go` -- ProjectHealth Extension

**Change**: Add `Staleness *StalenessInfo` field to `ProjectHealth`.

```go
type ProjectHealth struct {
    Project       Project             `json:"project"`
    Brief         Brief               `json:"brief"`
    Zones         map[Zone]ZoneHealth `json:"zones"`
    Workflow      *WorkflowState      `json:"workflow,omitempty"`
    LastIteration *Iteration          `json:"last_iteration,omitempty"`
    Warnings      []string            `json:"warnings,omitempty"`
    Suggestions   []string            `json:"suggestions,omitempty"`
    Staleness     *StalenessInfo      `json:"staleness"`           // NEW
}
```

**JSON behavior**: When `mind.lock` does not exist, `Staleness` is nil and serializes as `"staleness": null` (FR-78). When it exists, `StalenessInfo` contains status, stale map, and stats.

---

#### `internal/repo/interfaces.go` -- LockRepo Interface

**Change**: Add `LockRepo` interface.

```go
// LockRepo manages the mind.lock reconciliation state file.
type LockRepo interface {
    // Read loads mind.lock. Returns nil, nil if file does not exist.
    Read() (*domain.LockFile, error)

    // Write persists the lock file atomically (write to temp, rename).
    Write(lock *domain.LockFile) error

    // Exists returns true if mind.lock exists on disk.
    Exists() bool
}
```

**Rationale**: Follows the existing pattern where all repository interfaces live in `interfaces.go`. The interface is minimal (3 methods) per the existing Go idiom of small interfaces.

---

#### `internal/repo/fs/config_repo.go` -- Graph Parsing

**Change**: No code change needed. The `go-toml/v2` library automatically deserializes `[[graph]]` entries into `Config.Graph` because the `GraphEdge` struct will have proper `toml` tags. The TOML tag on `Config.Graph` is `toml:"graph"`, and `GraphEdge` fields have `toml:"from"`, `toml:"to"`, `toml:"type"`.

**Verification**: The developer should write a test confirming that `[[graph]]` TOML entries are correctly parsed into `Config.Graph` with the expected `GraphEdge` values.

---

#### `cmd/status.go` -- Staleness Panel

**Change**: After assembling `ProjectHealth`, check for lock file existence. If it exists, populate `health.Staleness` from `ReconciliationService.ReadStaleness()`. The staleness panel is rendered by the existing `Renderer.RenderHealth()` method (which must be updated to handle the new `Staleness` field).

**Key constraint**: `mind status` MUST NOT trigger reconciliation. It only reads existing lock data (FR-77). This is why `ReadStaleness()` is a separate method from `Reconcile()`.

---

#### `cmd/check.go` -- ReconcileSuite Integration

**Change**: `runCheckAll()` must run reconciliation (or read existing results) and include `ReconcileSuite` in the unified report.

**Approach**: In `runCheckAll()`, create a `ReconciliationService`, run reconciliation with `CheckOnly: true` to get a `ReconcileResult`, then pass it to `validate.ReconcileSuite()` to get a `ValidationReport`, and include it in the `UnifiedValidationReport`.

**Graceful degradation**: If `mind.toml` has no `[documents]` section or `[[graph]]` section, the reconcile suite reports 0 checks (empty suite). This prevents `mind check all` from failing on projects that do not use reconciliation.

---

#### `cmd/doctor.go` -- Stale Findings

**Change**: After existing checks, if `mind.lock` exists, read lock entries and produce WARN-level diagnostics for each stale document.

**Approach**: Add a `checkStaleness()` method to `DoctorService`. This method reads the lock file (via `LockRepo`), iterates stale entries, and appends `Diagnostic` entries with category "staleness", the stale reason as the message, and "Review and update this document, then run 'mind reconcile --force'" as the fix.

**Integration**: `DoctorService` needs `LockRepo` added to its constructor. This is a one-field addition to the struct.

---

#### `internal/render/render.go` -- Reconciliation Rendering

**Changes**:
1. **Update `renderHealthText()`**: Add staleness panel section after the workflow section. Show stale count and list each stale document with its reason. Only shown when `health.Staleness` is non-nil.
2. **Add `RenderReconcileResult()`**: New render method for `mind reconcile` output. Shows changed, stale, missing, undeclared documents with stats and overall status.
3. **Add `RenderGraph()`**: New render method for `mind reconcile --graph` output. Renders ASCII tree of dependency graph with stale annotations.

**JSON mode**: All new render methods use `jsonMarshal()` when mode is `ModeJSON`, following the established pattern.

---

#### `main.go` or `cmd/root.go` -- Centralized Wiring

**Change**: Centralize repository construction and service wiring. This addresses the tech debt identified in Phase 1 where each command handler creates its own repos.

**Approach**: Use `PersistentPreRunE` on `rootCmd` to:
1. Resolve project root (if applicable -- skip for commands that don't require a project)
2. Create all repository instances once
3. Create all service instances once
4. Store them in package-level variables accessible to command handlers

**Implementation detail**: Commands that do not require a project (version, help, init) bypass wiring by checking `cmd.Annotations["requires-project"]` or similar mechanism.

**Rationale**: Phase 1.5 adds 4 integration points (reconcile, status, check, doctor) that all need `LockRepo` and `ReconciliationService`. Without centralization, wiring code would be duplicated in 4+ places. Fixing this during Phase 1.5 is the cheapest point (per convergence recommendation 3) because those files are already being modified.

---

### New Interfaces

#### LockRepo

Defined in `internal/repo/interfaces.go`:

```go
type LockRepo interface {
    Read() (*domain.LockFile, error)
    Write(lock *domain.LockFile) error
    Exists() bool
}
```

**Read()**: Returns the parsed `LockFile` from `mind.lock`. Returns `nil, nil` when the file does not exist (first run). Returns `nil, error` on parse failures.

**Write()**: Atomically writes the lock file (temp file + rename). Caller is responsible for populating all fields including `GeneratedAt`.

**Exists()**: Fast existence check without parsing. Used by `mind status` to decide whether to show the staleness panel.

#### ReconciliationService Contract

Defined in `internal/service/reconciliation.go`:

```go
type ReconciliationService struct {
    configRepo repo.ConfigRepo
    docRepo    repo.DocRepo
    lockRepo   repo.LockRepo
}

func NewReconciliationService(configRepo repo.ConfigRepo, docRepo repo.DocRepo, lockRepo repo.LockRepo) *ReconciliationService

func (s *ReconciliationService) Reconcile(projectRoot string, opts domain.ReconcileOpts) (*domain.ReconcileResult, error)

func (s *ReconciliationService) ReadStaleness(projectRoot string) (*domain.StalenessInfo, error)
```

**Reconcile()**: Full reconciliation. Returns `ReconcileResult` for rendering and reporting. Handles lock file persistence internally (load, pass to engine, save).

**ReadStaleness()**: Read-only lock file inspection. Returns nil when no lock exists. Used by `mind status` and `mind doctor`.

---

## Key Decisions

### Decision 1: Reconcile Types in `domain/reconcile.go`

**What**: All reconciliation domain types (`LockFile`, `LockEntry`, `ReconcileResult`, `GraphEdge`, `EdgeType`, `Graph`, `LockStatus`, `EntryStatus`, `LockStats`, `ReconcileOpts`, `StalenessInfo`) live in a single new file `domain/reconcile.go`.

**Rationale**: Follows the existing pattern where domain types are organized by feature area (`project.go`, `document.go`, `iteration.go`, `workflow.go`, `health.go`, `validation.go`). A single file is appropriate because these types form a cohesive unit with strong internal relationships. The file will be approximately 150-200 lines, well within Go's norms for a single file.

**Rejected alternatives**:
- **Split across multiple domain files** (e.g., `domain/lock.go`, `domain/graph.go`): Would fragment closely related types and make the domain model harder to understand at a glance. Rejected because the types have tight coupling -- `LockFile` contains `LockEntry` and `LockStats`, `ReconcileResult` contains `LockStats` -- and belong together.
- **Types in `internal/reconcile/`**: Would break domain purity. The reconcile package is a service-layer concern; the types it operates on are domain concepts. Rejected per DC-1.

**Consequences**: `domain/reconcile.go` becomes the canonical location for all Phase 1.5 domain types. `domain/project.go` gains a single new field (`Config.Graph`). `domain/health.go` gains a single new field (`ProjectHealth.Staleness`).

**Cross-reference**: Convergence recommendation 1.4 (is_stub in lock entries computed via DocRepo).

---

### Decision 2: `internal/reconcile/` Depends on Repo Interfaces, Not Paths

**What**: `internal/reconcile/engine.go` accepts a `repo.DocRepo` interface (for `IsStub()` and `ListAll()`), while `internal/reconcile/hash.go` accepts absolute file paths directly.

**Rationale**: The engine needs `DocRepo` for two operations: stub detection (BR-34 requires using `DocRepo.IsStub()` rather than reimplementing) and undeclared file detection (comparing `DocRepo.ListAll()` against config entries). However, hash computation is a streaming I/O operation that does not fit the `DocRepo.Read() []byte` interface -- streaming is essential for performance with large files. The hash function accepts an absolute path, making it infrastructure-adjacent code.

**Rejected alternatives**:
- **Engine accepts only paths, no repo interfaces**: Would require reimplementing stub detection logic, violating BR-34 (single source of truth). Rejected.
- **Add streaming interface to DocRepo**: Over-engineering for one use case. The streaming hash is the only consumer. Adding `DocRepo.OpenReader()` would pollute the interface for all implementations. Rejected.
- **hash.go uses DocRepo.Read()**: Would load entire file content into memory before hashing. For a 10MB file, this means 10MB allocation + 10MB hash buffer. Streaming uses constant memory. Rejected for performance.

**Consequences**: `internal/reconcile/hash.go` imports `os` and `io` (filesystem packages). This is acceptable because `internal/reconcile/` is a service-layer package, not the domain layer. The engine's testability is preserved because its `DocRepo` dependency is an interface (testable with `mem.DocRepo`), and hash functions can be tested with temporary files.

**Cross-reference**: Convergence recommendation 1.3 (hash.go allowed direct os.Open()).

---

### Decision 3: ReconcileSuite Integrates with Existing Check Framework

**What**: `ReconcileSuite()` is a function in `internal/validate/` that takes a pre-computed `ReconcileResult` and produces a `Suite` following the existing `DocsSuite()`/`RefsSuite()`/`ConfigSuite()` pattern.

**Rationale**: Using the existing `Suite`/`Check` framework means `mind check all` can include reconciliation results without special-casing. The `UnifiedValidationReport` JSON schema naturally includes the reconcile suite. The renderer's existing `renderUnifiedValidationText()` renders it without modification.

**Rejected alternatives**:
- **Dedicated reconciliation section in check output**: Would require parallel rendering logic, special JSON schema handling, and break the established pattern. Rejected for consistency.
- **ReconcileSuite runs reconciliation internally**: Would couple the validation framework to the reconciliation engine and cause redundant reconciliation runs. The result is needed by the command handler for exit code determination and by the renderer for output. Rejected because pre-computing the result and projecting it into checks is cheaper and cleaner.

**Consequences**: `ReconcileSuite()` has a different signature from other suite constructors (it takes a `ReconcileResult` parameter rather than being zero-arg). This is an acceptable divergence because the reconciliation result is expensive to compute and must be shared. The `ValidationService.RunAll()` method gains a `reconcileResult` parameter.

**Cross-reference**: Convergence recommendation 5 (ReconcileSuite for mind check all).

---

### Decision 4: Wiring Centralization via `PersistentPreRunE` in `cmd/root.go`

**What**: Centralize all repository construction and service wiring in `rootCmd.PersistentPreRunE`. Commands access pre-built services through package-level variables in the `cmd` package.

**Rationale**: Phase 1.5 adds 4 integration points that all need the same repositories and services. Without centralization, each command handler creates repos redundantly (the current Phase 1 pattern). Centralizing during Phase 1.5 (not before, not after) is optimal because the files are already being modified for integration.

**Rejected alternatives**:
- **Keep wiring in each command handler**: Would add `LockRepo` and `ReconciliationService` creation to 4+ handlers. More duplication, harder to maintain, more error-prone. Rejected.
- **Wiring in `main.go`**: Would require `main.go` to know about all commands and pass services through `cobra.Command.Context()` or similar mechanism. Cobra's `PersistentPreRunE` is the idiomatic place for cross-cutting setup. Rejected because `PersistentPreRunE` is cleaner.
- **Fix wiring before starting Phase 1.5**: Would create a separate pull request / iteration for a refactor that touches the same files Phase 1.5 modifies. Merge conflicts, wasted effort. Rejected per convergence recommendation 3.
- **Defer wiring past Phase 1.5**: Every Phase 1.5 integration point would repeat the anti-pattern, making the eventual fix harder. Rejected.

**Consequences**: `cmd/root.go` gains a `PersistentPreRunE` function and package-level service variables. Existing command handlers are simplified (remove inline repo creation, use pre-built services). Commands that do not require a project (version, help, init) use a Cobra annotation or pre-run guard to skip wiring.

**Cross-reference**: Convergence recommendation 3 (wiring centralization during step 9).

---

### Decision 5: Engine Orchestrates 6 Phases, Service Handles I/O Boundaries

**What**: `internal/reconcile/engine.go` orchestrates the 6-phase algorithm from BP-06 (load, graph, scan, detect undeclared, propagate, report). The `ReconciliationService` handles the I/O boundaries: loading config, loading/saving lock files.

**Rationale**: Separating I/O boundary management (service) from algorithmic orchestration (engine) follows the existing separation between `service/validation.go` (which creates `CheckContext` and calls `Suite.Run()`) and `validate/check.go` (which executes checks). The engine is more testable because it receives parsed data rather than reading files itself.

**Rejected alternatives**:
- **Engine handles everything including I/O**: Would make the engine hard to test (requires real or mock filesystem for config and lock files). Would duplicate repository usage patterns. Rejected.
- **Service is the orchestrator, no engine**: Would put algorithmic logic (cycle detection, propagation, hash comparison) in the service layer, which is meant for coordination, not computation. Rejected per layer rules.

**Consequences**: The engine has a slightly unusual constructor -- it takes `DocRepo` for stub detection and undeclared file scanning but receives the config and lock file as parameters. This is a pragmatic split: config and lock I/O is service-level coordination; stub detection and file scanning is infrastructure access needed during the algorithm.

---

### Decision 6: Renderer Handles New Staleness Output Shapes

**What**: Add three new render methods: `RenderReconcileResult()` for `mind reconcile` output, `RenderGraph()` for `mind reconcile --graph`, and update `renderHealthText()` for the staleness panel.

**Rationale**: Follows the existing pattern where each domain output type has a corresponding `Render*()` method on the `Renderer`. The renderer remains the single place where output formatting decisions are made.

**New output shapes**:
- **Reconcile result**: Changed list, stale list with reasons, missing list, undeclared warnings, stats summary, overall status
- **Graph visualization**: ASCII tree using box-drawing characters, document IDs as nodes, edge types as labels, stale annotations
- **Staleness panel in status**: Stale count + per-document stale reasons, inserted after workflow section

**Rejected alternatives**:
- **Separate formatter structs per output type**: Premature; the renderer has only 12 render methods after Phase 1.5. Revisit if count exceeds 20. Rejected.
- **Graph rendering in the reconcile package**: Would violate the layer rule that rendering belongs in the presentation layer. Rejected.

**Consequences**: `render.go` grows by approximately 100-150 lines. The `Renderer` struct gains 2 new public methods. `renderHealthText()` gains approximately 15 lines for the staleness panel.

---

## Migration Path

The 12-step sequence below progresses from foundation to integration. Each step builds on the previous, is independently testable, and does not break existing behavior until the final integration steps which extend (not modify) existing outputs.

### Step 1: Domain Types
Add `domain/reconcile.go` with all new types, enums, and `BuildGraph()`. Add `Graph []GraphEdge` field to `Config` in `domain/project.go`. Add `Staleness *StalenessInfo` field to `ProjectHealth` in `domain/health.go`. Run `domain/purity_test.go` to verify zero external imports.

**Files changed**: `domain/reconcile.go` (new), `domain/project.go` (modified), `domain/health.go` (modified)
**Tests**: Unit tests for `BuildGraph()`, enum string values, JSON serialization round-trips

### Step 2: Config Extension
Add `toml` tags to `GraphEdge` so `go-toml/v2` parses `[[graph]]` entries into `Config.Graph`. Write tests verifying `[[graph]]` TOML parsing. Extend `ConfigSuite()` with graph entry validation checks (FR-84).

**Files changed**: `domain/reconcile.go` (add toml tags if not already present), `internal/validate/config.go` (add graph validation checks)
**Tests**: Config parsing test with `[[graph]]` entries, config validation test for invalid edge types

### Step 3: Hash Computation
Create `internal/reconcile/hash.go` with `HashFile()` and `NeedsRehash()`. Handle edge cases: empty files, binary detection, symlinks, large files, unreadable files.

**Files changed**: `internal/reconcile/hash.go` (new)
**Tests**: Hash of known content, empty file hash, mtime fast-path behavior, symlink resolution, error handling

### Step 4: Graph Operations
Create `internal/reconcile/graph.go` with `DetectCycle()` and `ValidateEdges()`.

**Files changed**: `internal/reconcile/graph.go` (new)
**Tests**: Acyclic graph, simple cycle, complex cycle with full path, edge validation against document set

### Step 5: Staleness Propagation
Create `internal/reconcile/propagate.go` with `PropagateDownstream()`. Test depth limit, diamond graphs, transitive paths, edge type reason messages.

**Files changed**: `internal/reconcile/propagate.go` (new)
**Tests**: Linear chain, diamond graph, depth limit at 10, changed documents not marked stale, edge type reason strings

### Step 6: Lock Repository
Add `LockRepo` interface to `internal/repo/interfaces.go`. Create `internal/repo/fs/lock_repo.go` and `internal/repo/mem/lock_repo.go`. Test atomic writes, round-trip correctness, missing file behavior.

**Files changed**: `internal/repo/interfaces.go` (modified), `internal/repo/fs/lock_repo.go` (new), `internal/repo/mem/lock_repo.go` (new)
**Tests**: Round-trip (read/write/read byte-identical), atomic write (temp file then rename), missing file returns nil, JSON schema compliance

### Step 7: Engine Orchestration
Create `internal/reconcile/engine.go`. Integrate hash, graph, propagate modules into 6-phase orchestration. Test full reconciliation scenarios: first run, incremental update, force reset.

**Files changed**: `internal/reconcile/engine.go` (new)
**Tests**: First-run (all clean), incremental with one change, force mode, cycle detection abort, missing documents, undeclared files

### Step 8: ReconciliationService
Create `internal/service/reconciliation.go`. Wire engine with repos. Test Reconcile and ReadStaleness workflows.

**Files changed**: `internal/service/reconciliation.go` (new)
**Tests**: Service-level reconciliation, read-only check mode, force mode, graceful handling when no config

### Step 9: Centralized Wiring
Refactor `cmd/root.go` to add `PersistentPreRunE` with centralized repo and service construction. Update all existing command handlers to use centralized services instead of inline repo creation. Add `LockRepo` and `ReconciliationService` to the centralized wiring.

**Files changed**: `cmd/root.go` (modified), `cmd/status.go` (simplified), `cmd/check.go` (simplified), `cmd/doctor.go` (simplified), all other `cmd/*.go` files (simplified)
**Tests**: Existing integration tests must still pass. No behavioral changes, only wiring refactoring.

### Step 10: Reconcile Command
Create `cmd/reconcile.go` with `--check`, `--force`, `--graph` flags. Add `RenderReconcileResult()` and `RenderGraph()` to renderer.

**Files changed**: `cmd/reconcile.go` (new), `internal/render/render.go` (modified)
**Tests**: Command-level tests for each flag combination, exit code verification, JSON output validation

### Step 11: Integration with Existing Commands
Modify `cmd/status.go` to include staleness panel. Modify `cmd/check.go` to include ReconcileSuite. Modify `cmd/doctor.go` to include stale findings. Create `internal/validate/reconcile.go` with ReconcileSuite. Update `renderHealthText()` for staleness panel.

**Files changed**: `cmd/status.go` (modified), `cmd/check.go` (modified), `cmd/doctor.go` (modified), `internal/validate/reconcile.go` (new), `internal/render/render.go` (modified), `internal/service/validation.go` (modified), `internal/service/doctor.go` (modified)
**Tests**: Status with and without lock file, check all with reconcile suite, doctor with stale findings, JSON output compliance

### Step 12: Performance Benchmarks
Add benchmark tests for full and incremental reconciliation. Verify against targets: full <200ms for 50 docs, incremental <50ms for 50 docs with 1 change.

**Files changed**: `internal/reconcile/engine_bench_test.go` (new)
**Tests**: Benchmark with synthetic 50-document project, timing assertions

---

## Dependency Flow After Phase 1.5

```
cmd/reconcile.go ──────> internal/service/reconciliation.go
cmd/status.go ─────────> internal/service/reconciliation.go (ReadStaleness)
cmd/check.go ──────────> internal/service/validation.go (extended)
cmd/doctor.go ─────────> internal/service/doctor.go (extended)
cmd/* ─────────────────> internal/render/ (new render methods)

internal/service/reconciliation.go ─> internal/reconcile/engine.go
internal/service/reconciliation.go ─> internal/repo (interfaces: ConfigRepo, DocRepo, LockRepo)
internal/service/validation.go ────> internal/validate/reconcile.go (ReconcileSuite)
internal/service/doctor.go ────────> internal/repo (interfaces: LockRepo)

internal/reconcile/engine.go ──────> internal/reconcile/hash.go
internal/reconcile/engine.go ──────> internal/reconcile/graph.go
internal/reconcile/engine.go ──────> internal/reconcile/propagate.go
internal/reconcile/engine.go ──────> internal/repo (interfaces: DocRepo)
internal/reconcile/engine.go ──────> domain

internal/reconcile/hash.go ────────> domain (LockEntry for NeedsRehash)
internal/reconcile/hash.go ────────> os, crypto/sha256, io (direct I/O)
internal/reconcile/graph.go ───────> domain (Graph, GraphEdge)
internal/reconcile/propagate.go ───> domain (Graph, GraphEdge)

internal/validate/reconcile.go ────> domain (ReconcileResult, CheckResult)

internal/repo/fs/lock_repo.go ────> domain (LockFile)
internal/repo/mem/lock_repo.go ───> domain (LockFile)

domain/reconcile.go ───────────────> Go stdlib only (time, encoding/json tags)
```

All dependency arrows point inward/downward. The domain layer remains pure Go. The new `internal/reconcile/` package sits at the service layer alongside `internal/service/` and `internal/validate/`. No circular dependencies are introduced.
