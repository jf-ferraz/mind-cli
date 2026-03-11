# Domain Model

Phase 1 domain model for mind-cli Core CLI. Scoped to entities, business rules, state machines, and constraints required by Phase 1 commands. The full 24-entity model is documented in [BP-02: Domain Model](../blueprints/02-domain-model.md); this document extracts the Phase 1 subset with implementation-specific detail.

## Entities

### Core Entities

| Entity | Description | Key Attributes | Relationships |
|--------|-------------|----------------|---------------|
| **Project** | A detected Mind Framework project on disk. Root aggregate through which all operations flow. | Root (string, absolute path), Name (string, from mind.toml), Config (*Config, parsed manifest, nil if absent), Framework (string, version), DocsRoot (string), MindRoot (string) | Has one Config; contains Documents across 5 Zones; has zero or more Iterations; has zero or one active WorkflowState |
| **Config** | Parsed `mind.toml` manifest. Declares project identity, stack, document registry, governance rules, and profiles. | Manifest, ProjectMeta, Profiles, Documents (map[zone][name]DocEntry), Governance | Belongs to Project; references Documents via registry |
| **Document** | A markdown file in one of the 5 documentation zones. | Path (string, relative), AbsPath (string), Zone (Zone), Name (string), Size (int64), ModTime (time.Time), IsStub (bool), Status (DocStatus) | Belongs to one Zone; may be referenced by Config.Documents |
| **Zone** | One of 5 documentation areas that organize documents by purpose and volatility. | Value: spec, blueprints, state, iterations, knowledge | Contains Documents |
| **Brief** | The project brief with section detection for business context gate enforcement. | Path, Exists (bool), IsStub (bool), HasVision (bool), HasDeliverables (bool), HasScope (bool), GateResult (BriefGate) | Is a specialized Document in the spec zone |
| **Iteration** | A per-change tracking folder under `docs/iterations/` containing 5 artifacts. | Seq (int), Type (RequestType), Descriptor (string), DirName (string), Path (string), Artifacts ([]Artifact), Status (IterationStatus), CreatedAt (time.Time) | Contains Artifacts; belongs to iterations zone |
| **Artifact** | A single file within an iteration folder. | Name (string), Path (string), Exists (bool) | Belongs to one Iteration |
| **WorkflowState** | Persisted state of an in-progress AI workflow, parsed from `docs/state/workflow.md`. | Type (RequestType), Descriptor (string), IterationPath (string), Branch (string), LastAgent (string), RemainingChain ([]string), Session (int), TotalSessions (int), Artifacts ([]CompletedArtifact), DispatchLog ([]DispatchEntry) | References one Iteration; belongs to Project |
| **ValidationReport** | Aggregated results from running a validation suite. | Suite (string), Checks ([]CheckResult), Total (int), Passed (int), Failed (int), Warnings (int) | Contains CheckResults |
| **CheckResult** | The outcome of a single validation check. | ID (int), Name (string), Level (CheckLevel), Passed (bool), Message (string) | Belongs to one ValidationReport |
| **ProjectHealth** | Aggregate status object for `mind status` output. Combines project metadata, brief gate, zone health, workflow state, and diagnostics. | Project, Brief, Zones (map[Zone]ZoneHealth), Workflow (*WorkflowState), LastIteration (*Iteration), Warnings ([]string), Suggestions ([]string) | References Project, Brief, ZoneHealth map, WorkflowState, Iteration |
| **ZoneHealth** | Completeness metrics for a single documentation zone. | Zone (Zone), Total (int), Present (int), Stubs (int), Complete (int), Files ([]Document) | Belongs to ProjectHealth; references Documents |
| **Diagnostic** | An issue found by `mind doctor` with remediation advice. | Level (CheckLevel), Message (string), Fix (string), AutoFix (bool) | Produced by doctor command |

### Supporting Types

| Type | Kind | Values/Structure | Used By |
|------|------|------------------|---------|
| **DocStatus** | Enum (string) | `draft`, `active`, `complete`, `stub` | Document |
| **BriefGate** | Enum (string) | `BRIEF_PRESENT`, `BRIEF_STUB`, `BRIEF_MISSING` | Brief |
| **RequestType** | Enum (string) | `NEW_PROJECT`, `BUG_FIX`, `ENHANCEMENT`, `REFACTOR`, `COMPLEX_NEW` | Iteration, WorkflowState |
| **IterationStatus** | Enum (string) | `in_progress`, `complete`, `incomplete` | Iteration |
| **CheckLevel** | Enum (string) | `FAIL`, `WARN`, `INFO` | CheckResult |
| **Manifest** | Struct | Schema (string), Generation (int), Updated (time.Time), Invariants (map[string]bool) | Config |
| **ProjectMeta** | Struct | Name, Description, Type (string), Stack (StackConfig), Commands (CmdConfig) | Config |
| **StackConfig** | Struct | Language, Framework, Testing (string) | ProjectMeta |
| **CmdConfig** | Struct | Dev, Test, Lint, Typecheck, Build (string) | ProjectMeta |
| **Governance** | Struct | MaxRetries (int), ReviewPolicy, CommitPolicy, BranchStrategy (string) | Config |
| **Profiles** | Struct | Active ([]string) | Config |
| **DocEntry** | Struct | ID, Path, Zone, Status (string) | Config.Documents registry |
| **CompletedArtifact** | Struct | Agent, Output, Location (string) | WorkflowState |
| **DispatchEntry** | Struct | Agent, File, Model, Status (string), StartedAt (time.Time), Duration (time.Duration) | WorkflowState |
| **Suggestion** | Struct | Action, Reason, Command (string) | ProjectHealth |
| **OutputMode** | Enum (int) | `ModeInteractive` (0), `ModePlain` (1), `ModeJSON` (2) | Renderer (render layer, not domain) |

## Relationships

```
Project 1───1 Config
Project 1───* Document (via zone scan)
Project 1───* Iteration
Project 1───0..1 WorkflowState
Project 1───1 Brief (via spec/project-brief.md)

Config 1───* DocEntry (via documents map)
Config 1───1 Manifest
Config 1───1 ProjectMeta
Config 1───1 Governance
Config 1───1 Profiles

Document *───1 Zone
Document *───1 DocStatus

Iteration 1───* Artifact (exactly 5 expected)
Iteration *───1 RequestType
Iteration *───1 IterationStatus

Brief *───1 BriefGate
Brief is-a Document (specialized)

ValidationReport 1───* CheckResult
CheckResult *───1 CheckLevel

ProjectHealth 1───1 Project
ProjectHealth 1───1 Brief
ProjectHealth 1───* ZoneHealth
ProjectHealth 1───0..1 WorkflowState
ProjectHealth 1───0..1 Iteration (last)

ZoneHealth 1───* Document (files in zone)
```

## Business Rules

| ID | Rule | Entities | Invariant |
|----|------|----------|-----------|
| **BR-1** | A project MUST have a `.mind/` directory to be considered valid. Detection walks up from cwd. | Project | `Project.MindRoot` must exist as a directory on disk |
| **BR-2** | A document is classified as a stub if it contains only headings (`# ...`), HTML comments (`<!-- ... -->`), and template placeholder text with no substantive content. | Document | `Document.IsStub` is derived from content analysis, not metadata |
| **BR-3** | The business context gate classifies the project brief as BRIEF_PRESENT (has Vision + Key Deliverables + Scope with real content), BRIEF_STUB (file exists but is a stub), or BRIEF_MISSING (file does not exist). | Brief, BriefGate | `Brief.GateResult` is deterministically derived from file presence and section analysis |
| **BR-4** | The business context gate blocks NEW_PROJECT and COMPLEX_NEW request types when the brief is missing or a stub. | Brief, BriefGate, RequestType | When `Brief.GateResult != BRIEF_PRESENT` and RequestType is NEW_PROJECT or COMPLEX_NEW, the gate fails |
| **BR-5** | Iteration directory naming follows `{NNN}-{TYPE}-{slug}` where NNN is zero-padded to 3 digits, TYPE is one of the canonical RequestType values, and slug is kebab-case derived from the descriptor. | Iteration | `Iteration.DirName` must match `^\d{3}-[A-Z_]+-[a-z0-9]` |
| **BR-6** | Iteration sequence numbers are derived from the highest existing sequence + 1. Gaps in existing sequences are not filled (max + 1, not first-available). | Iteration | Next seq = max(existing) + 1; if none exist, seq = 1 |
| **BR-7** | Each iteration folder MUST contain exactly 5 expected artifact files: overview.md, changes.md, test-summary.md, validation.md, retrospective.md. | Iteration, Artifact | `len(ExpectedArtifacts) == 5`; IterationStatus is derived from artifact presence |
| **BR-8** | Iteration status is derived: `complete` when all 5 artifacts exist, `in_progress` when overview.md exists but others are missing, `incomplete` when overview.md is missing. | Iteration, IterationStatus | Status is computed, never stored |
| **BR-9** | WorkflowState is idle when the state is nil or the Type field is empty. | WorkflowState | `WorkflowState.IsIdle()` returns true when `s == nil \|\| s.Type == ""` |
| **BR-10** | mind.toml project name must be kebab-case, matching `^[a-z][a-z0-9-]*$`. | Config, ProjectMeta | Validated by `mind check config` |
| **BR-11** | mind.toml schema version must match `^mind/v\d+\.\d+$`. Unknown versions produce a warning, not an error. | Config, Manifest | Validated at parse time |
| **BR-12** | mind.toml generation must be >= 1 and is auto-incremented on every write. | Config, Manifest | `Manifest.Generation >= 1` |
| **BR-13** | Document paths in mind.toml must start with `docs/` and end with `.md`. | Config, DocEntry | `DocEntry.Path` matches `^docs/.*\.md$` |
| **BR-14** | Document IDs in mind.toml must match `^doc:[a-z]+/[a-z][a-z0-9-]*$`. | Config, DocEntry | `DocEntry.ID` format is validated |
| **BR-15** | Document zones in mind.toml must be one of the 5 valid zones: spec, blueprints, state, iterations, knowledge. | Config, DocEntry, Zone | `DocEntry.Zone` must match an AllZones value |
| **BR-16** | Slugification: lowercase input, replace non-alphanumeric characters with hyphens, strip leading/trailing hyphens. Result must be non-empty. | Iteration, Document | `Slugify()` is deterministic and idempotent |
| **BR-17** | ADR sequence numbers use `{NNN}` (3+ digit zero-padded). Next sequence = max(existing) + 1. No existing ADRs means start at 001. | Document (ADR) | Sequence derivation is consistent with BR-6 |
| **BR-18** | Blueprint sequence numbers use `{NN}` (2+ digit zero-padded). Creating a blueprint also appends an entry to INDEX.md. | Document (Blueprint) | INDEX.md must be updated atomically with blueprint creation |
| **BR-19** | Request type classification from natural language uses a priority chain: explicit prefix (create:, fix:, add:, refactor:, analyze:) > keyword matching > default (ENHANCEMENT). | RequestType | `Classify()` is deterministic for the same input |
| **BR-20** | Exit codes are deterministic: 0 = success, 1 = validation failure/issues found, 2 = runtime error (I/O failure), 3 = configuration error (bad mind.toml, not a project). | All commands | Exit code is never random; same inputs produce same exit code |
| **BR-21** | All check commands treat WARN-level failures as passing (exit 0) unless `--strict` is set, in which case they become FAIL (exit 1). | ValidationReport, CheckResult | `--strict` changes behavior only for WARN-level checks |
| **BR-22** | governance.max-retries must be in range 0-5. | Config, Governance | Validated by config checks |
| **BR-23** | `mind init` MUST NOT overwrite an existing `.mind/` directory. | Project | Checked before any writes |

## State Machines

### Iteration Lifecycle

```
                  +───────────────────+
                  │     CREATED       │   (mind create iteration)
                  +────────┬──────────+
                           │
                           │ overview.md created
                           v
                  +───────────────────+
                  │   IN_PROGRESS     │   (overview.md exists,
                  │                   │    other artifacts missing)
                  +────────┬──────────+
                           │
                           │ all 5 artifacts exist
                           v
                  +───────────────────+
                  │    COMPLETE       │   (all 5 artifacts present)
                  +───────────────────+

Note: INCOMPLETE is an error state — overview.md is missing,
      which means the iteration was created incorrectly.
```

**Transitions**:
- CREATED -> IN_PROGRESS: Automatic when `mind create iteration` writes overview.md
- IN_PROGRESS -> COMPLETE: When all 5 artifacts exist on disk (detected by scan)
- Any -> INCOMPLETE: overview.md is missing (error condition, not a normal transition)

**Note**: There is no explicit state field stored. IterationStatus is computed by scanning artifact presence on every read. This is a derived state, not a persisted state.

### Brief Gate Classification

```
                    ┌──────────────┐
           ┌───────┤  File Check  ├───────┐
           │       └──────────────┘       │
           │ file missing                 │ file exists
           v                              v
    ┌──────────────┐              ┌──────────────┐
    │ BRIEF_MISSING│              │  Stub Check  │
    └──────────────┘              └──────┬───────┘
                                   │           │
                              stub │           │ has content
                                   v           v
                            ┌────────────┐  ┌──────────────┐
                            │ BRIEF_STUB │  │  Section     │
                            └────────────┘  │  Analysis    │
                                            └──────┬───────┘
                                                   │
                                              all sections
                                              present
                                                   v
                                            ┌──────────────┐
                                            │BRIEF_PRESENT │
                                            └──────────────┘
```

**Classification rules**:
1. File does not exist -> BRIEF_MISSING
2. File exists but is a stub -> BRIEF_STUB
3. File exists with real content, check sections: Vision, Key Deliverables, Scope must all be present -> BRIEF_PRESENT

### Validation Check Lifecycle

```
    ┌──────────┐
    │  PENDING │   (check registered in suite)
    └────┬─────┘
         │
         │  Suite.Run() executes CheckFunc
         v
    ┌──────────┐
    │ EXECUTED │
    └────┬─────┘
         │
    ┌────┴────┐
    │         │
    v         v
 ┌──────┐ ┌──────┐
 │ PASS │ │ FAIL │  (Level determines severity: FAIL, WARN, INFO)
 └──────┘ └──────┘
```

**Note**: Checks are stateless functions. They are executed once per suite run. There is no retry or re-check within a single run. The `--strict` flag modifies whether WARN-level failures count toward the exit code, but does not change the check logic itself.

### Workflow State

```
    ┌──────┐
    │ IDLE │   (workflow.md empty or Type == "")
    └──┬───┘
       │
       │  preflight creates iteration + sets state
       │  (Phase 3 — not implemented in Phase 1)
       v
    ┌─────────┐
    │ RUNNING │   (Type set, agents dispatching)
    └────┬────┘
         │
         │  handoff completes + clears state
         │  (Phase 3)
         v
    ┌──────┐
    │ IDLE │
    └──────┘
```

**Phase 1 note**: Phase 1 only reads workflow state for display purposes (`mind workflow status`, `mind status`). State transitions (idle -> running -> idle) are triggered by Phase 3 commands (preflight, handoff). Phase 1 treats WorkflowState as read-only.

## Constraints

### Cross-Entity Constraints

| ID | Constraint | Entities Involved |
|----|-----------|-------------------|
| **XC-1** | Every document in `mind.toml [documents]` must reference one of the 5 valid zones. The zone in the table key must match the zone field. | Config, DocEntry, Zone |
| **XC-2** | The document ID format `doc:{zone}/{name}` must match the table key structure `documents.{zone}.{name}`. | Config, DocEntry |
| **XC-3** | All paths in `mind.toml [documents]` must point to files under `docs/` (enforced by path prefix check). | Config, DocEntry, Document |
| **XC-4** | Blueprint INDEX.md entries must correspond to actual blueprint files on disk (checked by `mind check refs`). | Document (Blueprint), Document (INDEX.md) |
| **XC-5** | `current.md` "Recent Changes" section references must point to valid iteration IDs (checked by `mind check refs`). | Document (current.md), Iteration |
| **XC-6** | `workflow.md` iteration path reference must point to a valid iteration directory when workflow is not idle (checked by `mind check refs`). | WorkflowState, Iteration |
| **XC-7** | `.claude/CLAUDE.md` internal path references must resolve to existing files (checked by `mind check refs`). | Document (.claude/CLAUDE.md), Project |
| **XC-8** | ADR, blueprint, and iteration sequence numbers should be contiguous (gaps produce WARN, not FAIL). | Document (ADR), Document (Blueprint), Iteration |
| **XC-9** | No orphan documents: files in `docs/` that match document patterns should be registered in `mind.toml [documents]` (checked by `mind check refs`, WARN level). | Config, Document |

### Domain Layer Purity Constraints

| ID | Constraint |
|----|-----------|
| **DC-1** | `domain/` package imports only Go standard library. No `os`, `filepath`, `io`, `net`, or third-party packages. |
| **DC-2** | Domain types are pure data structures with minimal behavior. Business logic involving I/O is in the service layer. |
| **DC-3** | All enums (Zone, DocStatus, BriefGate, RequestType, IterationStatus, CheckLevel) are typed string constants, not raw strings. |
| **DC-4** | `Slugify()` and `Classify()` are the only domain functions with logic. Both are pure (no side effects, deterministic). |

---

## Phase 1.5: Reconciliation Engine Domain Extensions

### New Entities

| Entity | Description | Key Attributes | Relationships |
|--------|-------------|----------------|---------------|
| **LockFile** | Persisted reconciliation state stored as `mind.lock` in the project root. Contains per-document tracking entries, aggregate stats, and overall project staleness status. | GeneratedAt (time.Time), Status (LockStatus), Stats (LockStats), Entries (map[string]LockEntry) | Contains LockEntries; has one LockStats; has one LockStatus |
| **LockEntry** | Per-document tracking entry within the lock file. Records the document's hash, filesystem metadata, staleness state, and stub classification. | ID (string, document ID), Path (string, relative path), Hash (string, `sha256:{hex}`), Size (int64), ModTime (time.Time), Stale (bool), StaleReason (string), IsStub (bool), Status (EntryStatus) | Belongs to LockFile; references a DocEntry via ID |
| **ReconcileResult** | Computed result from a reconciliation run. Ephemeral -- returned from the engine to the caller, not persisted directly. | Changed ([]string), Stale (map[string]string), Missing ([]string), Undeclared ([]string), Status (LockStatus), Stats (LockStats) | Contains LockStats; has one LockStatus |
| **GraphEdge** | A directed dependency between two documents as declared in `mind.toml [[graph]]`. Represents the assertion that a change in `from` may invalidate `to`. | From (string, document ID), To (string, document ID), Type (EdgeType) | Belongs to Config (via Graph field); participates in Graph |
| **Graph** | Directed graph of document dependencies. Built from `[[graph]]` entries on every reconciliation run. Not persisted -- the manifest is the source of truth. | Forward (map[string][]GraphEdge), Reverse (map[string][]GraphEdge), Nodes (map[string]bool) | Contains GraphEdges; built from Config.Graph |

### New Supporting Types

| Type | Kind | Values/Structure | Used By |
|------|------|------------------|---------|
| **EdgeType** | Enum (string) | `informs`, `requires`, `validates` | GraphEdge |
| **LockStatus** | Enum (string) | `CLEAN`, `STALE`, `DIRTY` | LockFile, ReconcileResult |
| **EntryStatus** | Enum (string) | `PRESENT`, `MISSING`, `CHANGED`, `UNCHANGED` | LockEntry |
| **LockStats** | Struct | Total (int), Changed (int), Stale (int), Missing (int), Undeclared (int), Clean (int) | LockFile, ReconcileResult |
| **ReconcileOpts** | Struct | Force (bool), CheckOnly (bool), GraphOnly (bool) | ReconciliationService |

### New Relationships

```
Config 1───* GraphEdge (via [[graph]] section, stored as Graph []GraphEdge)

LockFile 1───* LockEntry (via Entries map, keyed by document ID)
LockFile 1───1 LockStats
LockFile *───1 LockStatus

Graph 1───* GraphEdge (via Forward and Reverse maps)

ReconcileResult 1───1 LockStats
ReconcileResult *───1 LockStatus

ProjectHealth 1───0..1 LockFile (read from mind.lock for staleness panel)
```

### New Business Rules

| ID | Rule | Entities | Invariant |
|----|------|----------|-----------|
| **BR-24** | Hash computation uses SHA-256 of raw file bytes with no content normalization (no line-ending conversion, no whitespace stripping, no BOM removal). The hash format is `sha256:{64-character lowercase hex digest}`. | LockEntry | Hash is deterministic: identical file content always produces identical hash. |
| **BR-25** | The mtime fast-path skips hash computation when file mtime and size match the lock entry. This is a performance optimization, not a correctness mechanism. If mtime lies (e.g., `touch` without content change), the engine computes a hash and discovers content is identical -- a wasted hash but no incorrect result. | LockEntry | Fast-path produces false negatives (unnecessary rehash) but never false positives (missed change when mtime differs). |
| **BR-26** | Staleness propagates downstream only (in the direction of graph edges). If A has an edge to B and B changes, A is NOT stale. B's content derives from A, not the reverse. | Graph, LockEntry | Directionality matches semantic document dependency flow. |
| **BR-27** | Staleness propagation has a depth limit of 10 levels. Documents beyond depth 10 are not marked stale, and a warning is emitted. In practice, document chains rarely exceed 4-5 levels. | Graph | Prevents pathological runaway in misconfigured graphs. |
| **BR-28** | A document that changed (new hash differs from old hash) is fresh, not stale. Changed documents may make downstream documents stale, but they are not themselves stale. | LockEntry | Changed and stale are mutually exclusive states for a single entry. |
| **BR-29** | Cycles in the dependency graph are invalid. Cycle detection uses DFS. If a back edge is found, reconciliation aborts with an error reporting the full cycle path. This is governed by the `no-circular-dependencies` invariant in `mind.toml [manifest.invariants]`. | Graph, GraphEdge | The dependency graph is a DAG (directed acyclic graph). |
| **BR-30** | All document IDs in `[[graph]]` edge `from` and `to` fields must reference documents declared in `[documents]`. Undeclared references are errors, not warnings. | GraphEdge, Config, DocEntry | Graph edges are validated against the document registry before graph construction. |
| **BR-31** | Lock file writes are atomic: content is written to `mind.lock.tmp`, then renamed to `mind.lock`. This prevents corrupted lock files from partial writes or process interruption. | LockFile | No partially-written lock files exist on disk after any operation. |
| **BR-32** | Exit code 4 indicates staleness detection, used exclusively by `mind reconcile --check`. This is semantically distinct from validation failure (exit 1) and runtime error (exit 2). | ReconcileResult | Exit code mapping is deterministic. |
| **BR-33** | Lock file overall status priority: STALE (any document stale) > DIRTY (no stale, any missing) > CLEAN (no stale, no missing). | LockFile, LockStatus | Status is deterministically derived from entry states. |
| **BR-34** | The `is_stub` field in lock entries is computed by delegating to the existing `DocRepo.IsStub()` method during reconciliation. Stub detection logic is not reimplemented in the reconciliation engine. | LockEntry, Document | Single source of truth for stub classification (BR-2). |
| **BR-35** | All three edge types (informs, requires, validates) propagate staleness identically. The type distinction affects only the staleness reason message text, not the propagation algorithm. | EdgeType, Graph | Uniform propagation with differentiated reporting. |

### New Cross-Entity Constraints

| ID | Constraint | Entities Involved |
|----|-----------|-------------------|
| **XC-10** | Every document ID in `[[graph]]` `from` and `to` fields must exist in `[documents]`. This is checked before graph construction, not during propagation. | Config, GraphEdge, DocEntry |
| **XC-11** | Lock file entries must correspond 1:1 with documents declared in `mind.toml [documents]` at reconciliation time. Entries for documents removed from the manifest are pruned. Entries for newly added documents are created. | LockFile, LockEntry, Config |
| **XC-12** | The `is_stub` value in lock entries must match the result of `DocRepo.IsStub()` for the same path at reconciliation time. If stub status changes between reconciliation runs, the lock entry is updated. | LockEntry, Document |

### New State Machines

#### Lock File Lifecycle

```
                 ┌──────────────────────────────────────┐
                 │         No mind.lock exists           │
                 └──────────────────┬───────────────────┘
                                    │
                              mind reconcile
                                    │
                                    ▼
                 ┌──────────────────────────────────────┐
                 │     Lock file created (baseline)      │
                 │     All entries: stale = false         │
                 └──────────────────┬───────────────────┘
                                    │
                        ┌───────────┼───────────┐
                        │           │           │
                  mind reconcile    │    mind reconcile
                                    │       --force
                        │           │           │
                        ▼           │           ▼
                 ┌──────────────┐   │   ┌──────────────┐
                 │ Incremental  │   │   │ Full reset   │
                 │ update       │   │   │ (re-baseline)│
                 │ (fast path)  │   │   │              │
                 └──────────────┘   │   └──────────────┘
                                    │
                          mind reconcile --check
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ Read-only verify  │
                         │ Exit 0 or Exit 4  │
                         │ (no write)        │
                         └──────────────────┘
```

**Operations**:
- **Create** (first run): Hash all documents, create lock, no staleness possible.
- **Update** (subsequent runs): mtime fast-path, rehash changed files, propagate staleness, write lock.
- **Verify** (`--check`): Same as update but does not write lock. Exits 0 (clean) or 4 (stale).
- **Reset** (`--force`): Discard existing lock, rehash everything, clear all staleness.

#### Document Entry Status

```
                ┌─────────────┐
                │  (no entry)  │  First reconciliation
                └──────┬──────┘
                       │
                       ▼
              ┌──────────────────┐
              │    PRESENT       │  File exists, hash computed
              └────────┬─────────┘
                       │
           ┌───────────┼───────────┐
           │           │           │
    file deleted   content same   content differs
           │           │           │
           ▼           ▼           ▼
    ┌──────────┐ ┌───────────┐ ┌──────────┐
    │ MISSING  │ │ UNCHANGED │ │ CHANGED  │
    └──────────┘ └───────────┘ └──────────┘
```

**Note**: Entry status is computed on every reconciliation run. It is not a persisted state machine -- it is derived from comparing current filesystem state against the previous lock entry.

### Updated Domain Layer Purity Constraints

| ID | Constraint |
|----|-----------|
| **DC-3** (updated) | All enums (Zone, DocStatus, BriefGate, RequestType, IterationStatus, CheckLevel, **EdgeType, LockStatus, EntryStatus**) are typed string constants, not raw strings. |
| **DC-4** (updated) | `Slugify()`, `Classify()`, **and `BuildGraph()`** are the only domain functions with logic. All are pure (no side effects, deterministic). `BuildGraph()` constructs forward/reverse adjacency lists from a slice of GraphEdge. |

---

## Phase 2: TUI Dashboard Domain Extensions

### New Entities

| Entity | Description | Key Attributes | Relationships |
|--------|-------------|----------------|---------------|
| **QualityEntry** | A single convergence analysis result parsed from `quality-log.yml`. Represents one data point in the quality score trend. | Topic (string), Variant (string), Date (time.Time), Score (float64, 0.0-5.0), GatePass (bool), Dimensions ([]QualityDimension), Personas ([]string), OutputPath (string) | Contains QualityDimensions (exactly 6) |
| **QualityDimension** | A single dimension score within a convergence analysis. One of 6 standard dimensions. | Name (string), Value (int, 0-5) | Belongs to QualityEntry |

### New Supporting Types

| Type | Kind | Values/Structure | Used By |
|------|------|------------------|---------|
| **QualityDimensionName** | Constants (string) | `"rigor"`, `"coverage"`, `"actionability"`, `"objectivity"`, `"convergence"`, `"depth"` | QualityDimension |

### New Relationships

```
QualityEntry 1───6 QualityDimension (exactly 6 per entry)
```

### New Business Rules

| ID | Rule | Entities | Invariant |
|----|------|----------|-----------|
| **BR-36** | A QualityEntry MUST have an overall score in the range 0.0 to 5.0 inclusive. | QualityEntry | `0.0 <= Score <= 5.0` |
| **BR-37** | A QualityEntry passes Gate 0 when its overall score is >= 3.0. This mirrors the convergence analysis gate requirement. | QualityEntry | `GatePass == (Score >= 3.0)` |
| **BR-38** | QualityEntry dimension scores MUST each be in the range 0 to 5 inclusive. There are exactly 6 standard dimensions: rigor, coverage, actionability, objectivity, convergence, depth. | QualityDimension | `0 <= Value <= 5` and `len(Dimensions) == 6` |

### Presentation-Layer Types (Not in Domain)

The following types are used by the TUI but belong in the `tui/` package, not `domain/`:

| Type | Kind | Values | Used By |
|------|------|--------|---------|
| **TabID** | Enum (int) | `TabStatus` (0), `TabDocs` (1), `TabIterations` (2), `TabChecks` (3), `TabQuality` (4) | TUI app shell |
| **ViewState** | Enum (int) | `ViewLoading` (0), `ViewError` (1), `ViewEmpty` (2), `ViewReady` (3) | TUI tab models |

These are explicitly excluded from `domain/` per DC-1 (domain purity). They represent presentation concerns (which tab is active, whether data has loaded) not business concepts.
