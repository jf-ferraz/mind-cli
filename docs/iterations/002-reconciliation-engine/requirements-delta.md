# Requirements Delta: Phase 1.5 Reconciliation Engine

**Iteration**: 002-reconciliation-engine
**Date**: 2026-03-11
**Phase**: 1.5
**Type**: COMPLEX_NEW (new subsystem added to existing codebase)

---

## Current State (Phase 1)

Phase 1 delivers a deterministic CLI with:
- Project detection, `mind.toml` parsing, 5-zone documentation structure
- 17-check doc validation, 11-check ref validation, config validation
- Document scaffolding for 6 artifact types
- Workflow state inspection (read-only)
- Three output modes: interactive, plain, JSON
- Exit codes: 0 (success), 1 (validation failure), 2 (runtime error), 3 (config error)
- Commands: status, init, doctor, create, docs, check, workflow, version, help

Phase 1 has **no temporal coherence detection**. It can verify that documents exist and have structure, but cannot answer whether a document is outdated relative to its dependencies. There is no `mind.lock`, no dependency graph, no hash tracking, and no staleness propagation.

## Desired State (Phase 1.5)

Phase 1.5 adds the reconciliation engine: a subsystem that computes SHA-256 hashes of document content, builds a directed dependency graph from `mind.toml [[graph]]` declarations, propagates staleness downstream when source documents change, and persists state in `mind.lock`.

After Phase 1.5, the system can answer three questions that Phase 1 cannot:
1. **What changed?** -- Which documents have different content since the last reconciliation.
2. **What is stale?** -- Which documents depend on something that changed and may need updating.
3. **What is the dependency chain?** -- Why a document is stale, traced back to the root change.

This integrates into existing commands (`status`, `check all`, `doctor`) and introduces a new command (`mind reconcile`) with `--check`, `--force`, and `--graph` flags.

---

## Scope Boundary

### In Scope (Phase 1.5)

- `mind reconcile` command with `--check`, `--force`, `--graph` flags
- SHA-256 hash computation with mtime fast-path optimization
- Dependency graph construction from `mind.toml [[graph]]` section
- Cycle detection with full cycle path reporting
- BFS staleness propagation with depth limit of 10
- `mind.lock` lifecycle: create, update, verify (read-only), reset
- Atomic lock file writes (write-to-temp then rename)
- Integration with `mind status` (staleness panel, read-only)
- Integration with `mind check all` (ReconcileSuite)
- Integration with `mind doctor` (stale findings with remediation)
- Exit code 4 for staleness detection
- Hash edge cases: empty files, binary files, symlinks, files >10MB, unreadable files
- Performance targets: full reconciliation <200ms, incremental <50ms for 50 docs
- `mind.toml` extension: `[[graph]]` section parsing
- Domain types: LockFile, LockEntry, ReconcileResult, GraphEdge, EdgeType, Graph
- New package: `internal/reconcile/` (hash, graph, propagate, engine)
- New repository: LockRepo (interface + fs + mem implementations)
- New service: ReconciliationService
- New validation suite: ReconcileSuite

### Out of Scope (Phase 1.5)

- Content normalization (line endings, whitespace, BOM) -- raw bytes only per BP-06 and convergence recommendation 1.1
- Git-based change detection -- filesystem-only per offline constraint (NFR-11)
- Automatic staleness resolution (running agents to update stale docs)
- MCP server integration with staleness data (Phase 3)
- TUI staleness panel (Phase 2)
- Watch mode automatic reconciliation (Phase 4)
- `mind reconcile --format dot` (DOT graph output) -- ASCII tree only in Phase 1.5
- Cross-project dependency tracking
- `--project` flag rename to `--project-root` (deferred tech debt, per convergence recommendation 4)
- GoDoc comment additions for existing Phase 1 code (deferred, per convergence recommendation 4)
- `docs search` DocRepo abstraction fix (deferred, per convergence recommendation 4)
- DoctorService full delegation refactor (deferred, per convergence recommendation 4)

---

## New Requirements

### Reconcile Command

- **FR-51**: `mind reconcile` MUST compute SHA-256 hashes for all documents declared in `mind.toml [documents]`, build the dependency graph from `mind.toml [[graph]]`, detect changes by comparing hashes to `mind.lock`, propagate staleness downstream through the graph, and write the updated `mind.lock` file. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 5 documents declared in `mind.toml [documents]` and 3 edges declared in `[[graph]]` and no existing `mind.lock`
  WHEN `mind reconcile` is run
  THEN `mind.lock` is created containing entries for all 5 documents with SHA-256 hashes, sizes, modification times, and stale=false for all entries.

- **FR-52**: `mind reconcile --check` MUST perform the same scan and staleness propagation as `mind reconcile` but MUST NOT write or modify `mind.lock`. It MUST exit with code 0 when all documents are clean (no staleness) and exit with code 4 when any documents are stale. [MUST]

  **Acceptance Criteria**:
  GIVEN a project where `mind.lock` exists and `requirements.md` has been modified since the last reconciliation and `architecture.md` depends on `requirements.md`
  WHEN `mind reconcile --check` is run
  THEN exit code is 4, output lists `architecture.md` as stale, and `mind.lock` is unchanged (same content and mtime as before the command ran).

- **FR-53**: `mind reconcile --force` MUST discard the existing `mind.lock` entirely, re-hash every declared document from scratch, clear all staleness flags, and write a new `mind.lock`. [MUST]

  **Acceptance Criteria**:
  GIVEN a project where `mind.lock` contains 2 entries marked as stale
  WHEN `mind reconcile --force` is run
  THEN `mind.lock` is rewritten with all entries having stale=false and fresh SHA-256 hashes matching current file content.

- **FR-54**: `mind reconcile --graph` MUST output an ASCII tree visualization of the dependency graph declared in `mind.toml [[graph]]`. The tree MUST show document IDs as nodes and edge types as labels. When staleness data exists in `mind.lock`, stale nodes MUST be visually annotated. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with edges `project-brief --(informs)--> requirements --(informs)--> architecture`
  WHEN `mind reconcile --graph` is run
  THEN output contains an ASCII tree with `doc:spec/project-brief` at the root, `doc:spec/requirements` as a child labeled `[informs]`, and `doc:spec/architecture` as a grandchild labeled `[informs]`.

- **FR-55**: `mind reconcile` MUST support `--json` output that produces a JSON object containing: changed document IDs, stale documents with reasons, missing documents, overall status (CLEAN/STALE/DIRTY), and stats (total, changed, stale, missing, clean counts). [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 1 changed document and 2 stale documents
  WHEN `mind reconcile --json` is run
  THEN stdout contains a valid JSON object with keys "changed", "stale", "missing", "status", "stats" where status is "STALE" and stats.stale equals 2.

- **FR-56**: `mind reconcile` MUST require a valid `mind.toml` with a `[documents]` section. When `mind.toml` is missing or has no documents section, the command MUST exit with code 3 and an actionable error message. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with `.mind/` but no `mind.toml`
  WHEN `mind reconcile` is run
  THEN exit code is 3 and stderr contains "mind.toml required for reconciliation".

### Hash Computation

- **FR-57**: The engine MUST compute SHA-256 hashes of raw file bytes with no content normalization (no line-ending conversion, no whitespace stripping, no BOM removal). The hash format MUST be `sha256:{64-character lowercase hex digest}`. [MUST]

  **Acceptance Criteria**:
  GIVEN a file containing exactly the bytes `Hello\r\nWorld\n`
  WHEN the hash is computed
  THEN the result is `sha256:` followed by the SHA-256 hex digest of those exact bytes, including the `\r`.

- **FR-58**: The engine MUST implement an mtime fast-path: before computing a hash, compare the file's modification time and size against the values stored in the lock entry. When both mtime and size are unchanged, the engine MUST skip hash computation and reuse the stored hash. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 50 documents where 49 have unchanged mtime/size and 1 has changed mtime
  WHEN `mind reconcile` is run
  THEN exactly 1 SHA-256 hash computation occurs (not 50).

- **FR-59**: The engine MUST handle hash computation edge cases as follows: empty files produce the SHA-256 hash of empty input (`sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`); binary files are hashed normally with a warning logged to stderr; symlinks are resolved to their target and the target content is hashed; files larger than 10MB are hashed normally with a warning logged to stderr; unreadable files are marked as MISSING with an error reason, not treated as a hash failure. [MUST]

  **Acceptance Criteria**:
  GIVEN an empty file declared in `mind.toml [documents]`
  WHEN reconciliation runs
  THEN the lock entry hash equals `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

  GIVEN a binary file declared in `mind.toml [documents]`
  WHEN reconciliation runs
  THEN the lock entry contains a valid SHA-256 hash AND stderr contains `warning: binary file detected:`.

  GIVEN a symlink declared in `mind.toml [documents]` pointing to a file within the project root
  WHEN reconciliation runs
  THEN the lock entry hash matches the SHA-256 of the symlink target's content.

  GIVEN a symlink declared in `mind.toml [documents]` pointing to a file outside the project root
  WHEN reconciliation runs
  THEN the lock entry hash matches the SHA-256 of the target's content AND stderr contains `warning: symlink target outside project root:`.

  GIVEN a file larger than 10MB declared in `mind.toml [documents]`
  WHEN reconciliation runs
  THEN the lock entry contains a valid SHA-256 hash AND stderr contains `warning: large file`.

  GIVEN a file with no read permission declared in `mind.toml [documents]`
  WHEN reconciliation runs
  THEN the lock entry status is MISSING with an error reason describing the permission failure.

### Dependency Graph

- **FR-60**: The engine MUST parse `[[graph]]` entries from `mind.toml`, where each entry has `from` (document ID), `to` (document ID), and `type` (edge type). The engine MUST construct a directed graph as an adjacency list with both forward edges (from -> to) and reverse edges (to -> from). [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` with 4 `[[graph]]` entries defining edges between 5 documents
  WHEN the graph is constructed
  THEN the graph contains exactly 4 forward edges, 4 reverse edges, and 5 nodes.

- **FR-61**: The engine MUST support three edge types: `informs` (upstream content informs downstream), `requires` (upstream is a prerequisite for downstream), and `validates` (upstream validates downstream correctness). All three edge types MUST propagate staleness. The edge type MUST affect the staleness reason message: `informs` produces "may be outdated", `requires` produces "prerequisite changed", `validates` produces "needs re-validation". [MUST]

  **Acceptance Criteria**:
  GIVEN a graph with edge A --(requires)--> B and A changes
  WHEN staleness propagation runs
  THEN B is marked stale with a reason containing "prerequisite changed".

  GIVEN a graph with edge A --(validates)--> B and A changes
  WHEN staleness propagation runs
  THEN B is marked stale with a reason containing "needs re-validation".

  GIVEN a graph with edge A --(informs)--> B and A changes
  WHEN staleness propagation runs
  THEN B is marked stale with a reason containing "may be outdated".

- **FR-62**: The engine MUST detect cycles in the dependency graph using depth-first search. When a cycle is detected, reconciliation MUST abort with exit code 2 and an error message that includes the full cycle path (e.g., "circular dependency detected: A -> B -> C -> A"). [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` with edges A --> B, B --> C, C --> A
  WHEN `mind reconcile` is run
  THEN exit code is 2 and stderr contains "circular dependency detected" followed by the full cycle path including all three nodes.

- **FR-63**: The engine MUST validate that all document IDs referenced in `[[graph]]` entries exist in the `[documents]` section of `mind.toml`. References to undeclared document IDs MUST produce an error before graph construction proceeds. [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` with a `[[graph]]` entry referencing `doc:spec/nonexistent` which is not in `[documents]`
  WHEN `mind reconcile` is run
  THEN exit code is 3 and stderr contains "graph references undeclared document: doc:spec/nonexistent".

- **FR-64**: When `mind.toml` contains no `[[graph]]` entries, the engine MUST still hash and track all declared documents. No staleness propagation occurs. This is a valid configuration for projects that need hash-based change detection without dependency tracking. [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` with 5 documents in `[documents]` and zero `[[graph]]` entries
  WHEN `mind reconcile` is run twice (with one document changed between runs)
  THEN the changed document is reported as changed, no documents are reported as stale, and `mind.lock` is updated with the new hash.

### Staleness Propagation

- **FR-65**: The engine MUST propagate staleness downstream only (in the direction of graph edges). When document A has an edge to document B and A changes, B MUST be marked as stale. When B changes, A MUST NOT be marked as stale. [MUST]

  **Acceptance Criteria**:
  GIVEN a graph with edge A --(informs)--> B where B changes (but A does not)
  WHEN staleness propagation runs
  THEN A is NOT marked as stale and B is marked as changed (not stale, because it changed itself).

- **FR-66**: The engine MUST propagate staleness transitively. When the graph contains A --> B --> C and A changes, both B and C MUST be marked as stale. The stale reason for C MUST include the transitive path information. [MUST]

  **Acceptance Criteria**:
  GIVEN a graph with edges A --> B --> C where A changes
  WHEN staleness propagation runs
  THEN B is marked stale with reason referencing A, AND C is marked stale with reason referencing A and indicating transitive propagation.

- **FR-67**: The engine MUST enforce a staleness propagation depth limit of 10 levels. When the limit is reached, propagation MUST stop and a warning MUST be logged to stderr. Documents beyond the depth limit are not marked as stale. [MUST]

  **Acceptance Criteria**:
  GIVEN a linear dependency chain A1 --> A2 --> ... --> A12 where A1 changes
  WHEN staleness propagation runs
  THEN A2 through A11 are marked stale, A12 is NOT marked stale, and stderr contains a warning about the depth limit being reached.

- **FR-68**: A document that is itself changed (new hash differs from lock hash) MUST NOT be marked as stale. Changed documents are fresh -- they may make others stale, but they are not themselves stale. [MUST]

  **Acceptance Criteria**:
  GIVEN a graph A --> B where both A and B have changed content
  WHEN staleness propagation runs
  THEN A is marked as changed (not stale), B is marked as changed (not stale), and neither appears in the stale list.

- **FR-69**: A document that is already marked as stale via one path MUST NOT be re-processed via another path. The first staleness reason is retained. [MUST]

  **Acceptance Criteria**:
  GIVEN a diamond graph where A --> B, A --> C, B --> D, C --> D and A changes
  WHEN staleness propagation runs
  THEN D is marked stale exactly once (not twice), with the reason from whichever path reached D first.

### Lock File Lifecycle

- **FR-70**: The lock file MUST be located at `mind.lock` in the project root directory. [MUST]

  **Acceptance Criteria**:
  GIVEN a project rooted at `/path/to/project`
  WHEN `mind reconcile` is run
  THEN the lock file is written to `/path/to/project/mind.lock`.

- **FR-71**: The lock file MUST be JSON format containing: a `generated_at` timestamp (RFC 3339), an overall `status` string (CLEAN, STALE, or DIRTY), a `stats` object with counts (total, changed, stale, missing, clean), and an `entries` map keyed by document ID where each entry contains: id, path, hash, size, mod_time, stale (boolean), stale_reason (string), is_stub (boolean), and status (PRESENT, MISSING, CHANGED, UNCHANGED). [MUST]

  **Acceptance Criteria**:
  GIVEN `mind reconcile` has run
  WHEN `mind.lock` is read and parsed as JSON
  THEN it contains all specified top-level keys, entries are keyed by document ID, and each entry has all specified fields with correct types.

- **FR-72**: The lock file MUST survive round-trip: reading `mind.lock`, parsing it, and writing it back without any intermediate changes MUST produce byte-identical output. [MUST]

  **Acceptance Criteria**:
  GIVEN an existing `mind.lock` with known content
  WHEN the lock file is read, parsed into a LockFile struct, serialized back to JSON, and written
  THEN the output bytes are identical to the input bytes.

- **FR-73**: The lock file MUST be written atomically: write to a temporary file (`mind.lock.tmp`) then rename to `mind.lock`. This prevents corrupted lock files from partial writes. [MUST]

  **Acceptance Criteria**:
  GIVEN a reconciliation that writes `mind.lock`
  WHEN the write operation occurs
  THEN at no point does `mind.lock` contain partial or invalid JSON (verified by checking that the write uses temp-file-then-rename pattern).

- **FR-74**: When `mind.lock` does not exist, `mind reconcile` MUST treat this as a first run: hash all documents, create the lock file, and report no staleness (no baseline to compare against). [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 5 declared documents and no `mind.lock`
  WHEN `mind reconcile` is run
  THEN `mind.lock` is created with 5 entries, all with stale=false, and the overall status is CLEAN.

- **FR-75**: Each lock entry MUST include an `is_stub` boolean field computed by delegating to the existing `DocRepo.IsStub()` method during reconciliation. The engine MUST NOT reimplement stub detection logic. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 2 stub documents and 3 non-stub documents
  WHEN `mind reconcile` is run
  THEN the 2 lock entries for stub documents have `is_stub: true` and the 3 other entries have `is_stub: false`.

- **FR-76**: The lock file overall status MUST be computed as follows: STALE when any document is stale, DIRTY when no documents are stale but any are missing, CLEAN when no documents are stale and none are missing. [MUST]

  **Acceptance Criteria**:
  GIVEN a reconciliation result with 0 stale and 0 missing documents
  WHEN the lock file is written
  THEN status is "CLEAN".

  GIVEN a reconciliation result with 0 stale but 1 missing document
  WHEN the lock file is written
  THEN status is "DIRTY".

  GIVEN a reconciliation result with 2 stale documents
  WHEN the lock file is written
  THEN status is "STALE".

### Integration with Existing Commands

- **FR-77**: `mind status` MUST display a staleness panel when `mind.lock` exists and contains stale documents. The panel MUST show the count of stale documents and list each stale document with its reason. When `mind.lock` does not exist, the staleness panel MUST be omitted. `mind status` MUST NOT trigger reconciliation -- it reads existing lock data only. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with `mind.lock` containing 2 stale entries
  WHEN `mind status` is run
  THEN output includes a staleness section showing "2 stale documents" and listing both with their reasons.

  GIVEN a project with no `mind.lock`
  WHEN `mind status` is run
  THEN output does NOT include a staleness section and no error is produced.

- **FR-78**: `mind status --json` MUST include a `staleness` object in the JSON output when `mind.lock` exists. The staleness object MUST contain: status (CLEAN/STALE/DIRTY), stale documents with reasons, and stats. When `mind.lock` does not exist, the `staleness` key MUST be null. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with `mind.lock` containing status "STALE" with 2 stale documents
  WHEN `mind status --json` is run
  THEN the JSON output contains a "staleness" object with "status": "STALE" and a "stale" map with 2 entries.

  GIVEN a project with no `mind.lock`
  WHEN `mind status --json` is run
  THEN the JSON output contains "staleness": null.

- **FR-79**: `mind check all` MUST include a ReconcileSuite in its unified validation report alongside the existing docs, refs, and config suites. The ReconcileSuite MUST project reconciliation results into CheckResult entries: one check for cycle detection (FAIL if cycle exists), one check for missing documents (WARN per missing document), and one check per stale document (WARN in normal mode, FAIL with `--strict`). [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 2 stale documents and no cycles
  WHEN `mind check all` is run
  THEN the output includes a "reconcile" suite section with the cycle check passing and 2 stale document checks at WARN level, and exit code is 0.

  GIVEN the same project
  WHEN `mind check all --strict` is run
  THEN the 2 stale document checks are at FAIL level and exit code is 1.

- **FR-80**: `mind check all --json` MUST include a "reconcile" entry in the "suites" array of the JSON output. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with `mind.toml` containing `[[graph]]` entries
  WHEN `mind check all --json` is run
  THEN the JSON output contains a "suites" array with an entry where "name" is "reconcile" containing check results.

- **FR-81**: `mind doctor` MUST report stale documents as diagnostic findings when `mind.lock` exists. Each stale document MUST produce a WARN-level finding with the stale reason and the remediation text "Review and update this document, then run 'mind reconcile --force'". [MUST]

  **Acceptance Criteria**:
  GIVEN a project with `mind.lock` containing 2 stale entries
  WHEN `mind doctor` is run
  THEN output includes 2 WARN-level findings for stale documents, each with the specific stale reason and "mind reconcile --force" in the remediation text.

### Exit Codes

- **FR-82**: The CLI MUST add exit code 4 to represent staleness detection. Exit code 4 is returned exclusively by `mind reconcile --check` when stale documents exist. All other exit code semantics (0, 1, 2, 3) remain unchanged. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with stale documents
  WHEN `mind reconcile --check` is run
  THEN exit code is 4.

  GIVEN a project with no stale documents
  WHEN `mind reconcile --check` is run
  THEN exit code is 0.

  GIVEN a project with stale documents
  WHEN `mind check all` is run (not `--strict`)
  THEN exit code is 0 (staleness produces WARN, not FAIL in `check all`).

### Config Extension

- **FR-83**: `mind.toml` MUST support a `[[graph]]` array-of-tables section where each entry has three required string fields: `from` (source document ID), `to` (target document ID), and `type` (edge type: "informs", "requires", or "validates"). [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` containing:
  ```toml
  [[graph]]
  from = "doc:spec/requirements"
  to   = "doc:spec/architecture"
  type = "informs"
  ```
  WHEN the config is parsed
  THEN `Config.Graph` contains one GraphEdge with From="doc:spec/requirements", To="doc:spec/architecture", Type="informs".

- **FR-84**: `mind check config` MUST validate `[[graph]]` entries: each `from` and `to` value MUST match the document ID format `^doc:[a-z]+/[a-z][a-z0-9-]*$`, and each `type` value MUST be one of "informs", "requires", "validates". Invalid entries MUST produce a FAIL-level check result. [MUST]

  **Acceptance Criteria**:
  GIVEN a `mind.toml` with a `[[graph]]` entry where `type = "invalid-type"`
  WHEN `mind check config` is run
  THEN a FAIL-level check reports that the edge type is invalid, listing valid types.

### Undeclared File Detection

- **FR-85**: During reconciliation, the engine MUST scan the `docs/` directory recursively and identify files that exist on disk but are not declared in `mind.toml [documents]`. These undeclared files MUST be reported as warnings in the reconciliation output. Undeclared files MUST NOT participate in staleness propagation. [MUST]

  **Acceptance Criteria**:
  GIVEN a project where `docs/spec/notes.md` exists on disk but is not in `mind.toml [documents]`
  WHEN `mind reconcile` is run
  THEN the output includes a warning about `docs/spec/notes.md` being undeclared, and the file is not included in any staleness propagation.

### Performance

- **FR-86**: Full reconciliation (hashing all documents, building graph, propagating staleness, writing lock file) MUST complete in under 200ms for a project with 50 documents. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 50 declared documents and 40 graph edges
  WHEN `mind reconcile --force` is run
  THEN wall-clock time from invocation to completion is under 200ms.

- **FR-87**: Incremental reconciliation (mtime fast-path, rehashing only changed files) MUST complete in under 50ms for a project with 50 documents where 1 document has changed. [MUST]

  **Acceptance Criteria**:
  GIVEN a project with 50 documents, existing `mind.lock`, and 1 document with changed mtime
  WHEN `mind reconcile` is run
  THEN wall-clock time from invocation to completion is under 50ms.

---

## Modified Requirements

### FR-49 (Exit Codes) -- Extended

**Previous**: Exit codes 0 (success), 1 (validation failure), 2 (runtime error), 3 (config error).

**Updated**: Exit codes 0 (success), 1 (validation failure or issues found), 2 (runtime error), 3 (configuration error or not a Mind project), **4 (staleness detected, used exclusively by `mind reconcile --check`)**.

This is an additive change. Existing exit code semantics are unchanged. Exit code 4 is new.

### FR-42 (Check All) -- Extended

**Previous**: `mind check all` runs docs, refs, and config suites.

**Updated**: `mind check all` runs docs, refs, config, **and reconcile** suites and produces a unified `ValidationReport`.

The reconcile suite is added as a fourth suite. Existing suite behavior is unchanged.

### FR-11 (Status Display) -- Extended

**Previous**: `mind status` displays project health without staleness information.

**Updated**: `mind status` displays project health **plus a staleness panel when `mind.lock` exists**. When `mind.lock` does not exist, behavior is identical to Phase 1.

### FR-12 (Status JSON) -- Extended

**Previous**: `mind status --json` produces ProjectHealth JSON without staleness.

**Updated**: `mind status --json` produces ProjectHealth JSON **with an additional `staleness` key** (object when `mind.lock` exists, null when it does not).

### FR-20 (Doctor Diagnostics) -- Extended

**Previous**: `mind doctor` runs framework, adapter, doc, ref, config, brief, stub, workflow, and iteration checks.

**Updated**: `mind doctor` runs all previous checks **plus staleness diagnostics** when `mind.lock` exists. When `mind.lock` does not exist, behavior is identical to Phase 1.

---

## Unchanged Requirements

The following Phase 1 requirements MUST NOT change in behavior or interface:

- **FR-1 through FR-5**: Project detection and configuration (unchanged)
- **FR-6 through FR-10**: Output modes (unchanged; new command follows same patterns)
- **FR-13**: Brief gate classification (unchanged)
- **FR-14 through FR-19**: Init command (unchanged)
- **FR-21 through FR-23**: Doctor remediation and --fix behavior (unchanged for existing checks)
- **FR-24 through FR-31**: Create commands (unchanged)
- **FR-32 through FR-37**: Docs commands (unchanged)
- **FR-38 through FR-41**: Check docs, refs, config commands (unchanged)
- **FR-43**: Check exit codes for docs/refs/config (unchanged; new suite follows same pattern)
- **FR-44 through FR-45**: Workflow commands (unchanged)
- **FR-46 through FR-48**: Version and help (unchanged)
- **FR-50**: Stub detection logic (unchanged; reused by reconciliation via DocRepo.IsStub())

All non-functional requirements (NFR-1 through NFR-11) remain unchanged. Phase 1.5 adds new performance targets (FR-86, FR-87) that do not conflict with existing ones.

All constraints (C-1 through C-17) remain unchanged. Phase 1.5 operates within the existing 4-layer architecture, uses no new external dependencies (SHA-256 is Go stdlib), and follows constructor injection patterns.

All business rules (BR-1 through BR-23) remain unchanged. Phase 1.5 adds new business rules (see Domain Model Impact below).

---

## Domain Model Impact

### New Entities

| Entity | Description | Key Attributes | Module |
|--------|-------------|----------------|--------|
| **LockFile** | Persisted reconciliation state. Contains all lock entries, stats, and overall status. | GeneratedAt (time.Time), Status (LockStatus), Stats (LockStats), Entries (map[string]LockEntry) | `domain/` |
| **LockEntry** | Per-document tracking entry within the lock file. | ID (string), Path (string), Hash (string), Size (int64), ModTime (time.Time), Stale (bool), StaleReason (string), IsStub (bool), Status (EntryStatus) | `domain/` |
| **ReconcileResult** | Computed result from a reconciliation run. Not persisted -- returned to caller. | Changed ([]string), Stale (map[string]string), Missing ([]string), Undeclared ([]string), Status (LockStatus), Stats (LockStats) | `domain/` |
| **GraphEdge** | A directed dependency between two documents as declared in `mind.toml [[graph]]`. | From (string), To (string), Type (EdgeType) | `domain/` |
| **Graph** | Directed graph of document dependencies. Contains forward edges, reverse edges, and node set. | Forward (map[string][]GraphEdge), Reverse (map[string][]GraphEdge), Nodes (map[string]bool) | `domain/` |

### New Supporting Types

| Type | Kind | Values/Structure | Used By |
|------|------|------------------|---------|
| **EdgeType** | Enum (string) | `informs`, `requires`, `validates` | GraphEdge |
| **LockStatus** | Enum (string) | `CLEAN`, `STALE`, `DIRTY` | LockFile, ReconcileResult |
| **EntryStatus** | Enum (string) | `PRESENT`, `MISSING`, `CHANGED`, `UNCHANGED` | LockEntry |
| **LockStats** | Struct | Total (int), Changed (int), Stale (int), Missing (int), Undeclared (int), Clean (int) | LockFile, ReconcileResult |
| **ReconcileOpts** | Struct | Force (bool), CheckOnly (bool), GraphOnly (bool) | ReconciliationService |

### New Business Rules

| ID | Rule | Entities | Invariant |
|----|------|----------|-----------|
| **BR-24** | Hash computation uses SHA-256 of raw file bytes with no content normalization. The hash format is `sha256:{64-char hex}`. | LockEntry | Hash is deterministic for identical file content |
| **BR-25** | The mtime fast-path skips hash computation when file mtime and size match the lock entry. This is an optimization; correctness does not depend on it. | LockEntry | Fast-path may produce false negatives (rehash when content unchanged) but never false positives (skip when content changed and mtime differs) |
| **BR-26** | Staleness propagates downstream only. If A --> B and B changes, A is NOT stale. | Graph, LockEntry | Directionality matches semantic document flow |
| **BR-27** | Staleness propagation has a depth limit of 10. Documents beyond depth 10 in a dependency chain are not marked stale. A warning is emitted. | Graph | Prevents runaway propagation in pathological graphs |
| **BR-28** | A document that changed (new hash != old hash) is fresh, not stale. It may make downstream documents stale, but it is not itself stale. | LockEntry | Changed and stale are mutually exclusive states for a single entry |
| **BR-29** | Cycles in the dependency graph are invalid and cause reconciliation to abort. | Graph | No back edges in the directed graph |
| **BR-30** | All document IDs in `[[graph]]` edges must reference documents declared in `[documents]`. Undeclared references are errors. | GraphEdge, Config | Graph edges validated against document registry |
| **BR-31** | Lock file writes are atomic: write to temp file, then rename. | LockFile | No partially-written lock files on disk |
| **BR-32** | Exit code 4 indicates staleness, used exclusively by `mind reconcile --check`. | ReconcileResult | Distinct from validation failure (exit 1) and runtime error (exit 2) |
| **BR-33** | Lock file overall status is STALE > DIRTY > CLEAN (STALE takes precedence over DIRTY). | LockFile, LockStatus | Status priority is deterministic |
| **BR-34** | The `is_stub` field in lock entries is computed by delegating to `DocRepo.IsStub()`. Stub detection logic is not reimplemented. | LockEntry | Single source of truth for stub classification |
| **BR-35** | All three edge types (informs, requires, validates) propagate staleness. The distinction affects reporting messages only, not propagation behavior. | EdgeType, Graph | Uniform propagation, differentiated messaging |

### New Relationships

```
Config 1───* GraphEdge (via [[graph]] section)

LockFile 1───* LockEntry (via entries map)
LockFile 1───1 LockStats
LockFile 1───1 LockStatus

Graph 1───* GraphEdge (via forward/reverse maps)

ReconcileResult 1───1 LockStats
ReconcileResult 1───1 LockStatus

ProjectHealth 1───0..1 LockFile (via mind.lock read, for staleness panel)
```

### New Cross-Entity Constraints

| ID | Constraint | Entities Involved |
|----|-----------|-------------------|
| **XC-10** | Every document ID in `[[graph]]` `from` and `to` fields must exist in `[documents]`. | Config, GraphEdge, DocEntry |
| **XC-11** | Lock file entries must correspond 1:1 with documents declared in `mind.toml [documents]`. Extra lock entries (from removed documents) are pruned on reconciliation. | LockFile, LockEntry, Config |
| **XC-12** | The `is_stub` value in lock entries must be consistent with the result of `DocRepo.IsStub()` for the same path at reconciliation time. | LockEntry, Document |

---

## Structural Impact

Phase 1.5 requires the following new modules, services, and repositories. The specific architecture and API design is the architect's responsibility; this section identifies what is needed.

### New Packages

| Package | Purpose |
|---------|---------|
| `internal/reconcile/` | Reconciliation engine: hash computation, graph construction, staleness propagation, engine orchestration |

### New Files (Anticipated)

| File | Purpose |
|------|---------|
| `internal/reconcile/hash.go` | SHA-256 computation, mtime fast-path, edge case handling. Allowed to use `os.Open()` directly (per convergence recommendation 1.3). |
| `internal/reconcile/graph.go` | Adjacency list construction, cycle detection |
| `internal/reconcile/propagate.go` | BFS/DFS downstream staleness propagation with depth limit |
| `internal/reconcile/engine.go` | Top-level 6-phase orchestration (load, scan, detect undeclared, propagate, write, report) |
| `internal/repo/fs/lock_repo.go` | Filesystem LockRepo: read/write mind.lock as JSON |
| `internal/repo/mem/lock_repo.go` | In-memory LockRepo for testing |
| `internal/service/reconciliation.go` | ReconciliationService with constructor injection |
| `cmd/reconcile.go` | `mind reconcile` command with flags |
| `internal/validate/reconcile_suite.go` | ReconcileSuite for `mind check all` integration |

### Extended Files

| File | Change |
|------|--------|
| `domain/reconcile.go` (new) | LockFile, LockEntry, ReconcileResult, GraphEdge, EdgeType, Graph, LockStatus, EntryStatus, LockStats |
| `domain/project.go` | Add `Graph []GraphEdge` field to `Config` struct |
| `internal/repo/interfaces.go` | Add `LockRepo` interface |
| `cmd/status.go` | Add staleness panel when mind.lock exists |
| `cmd/check.go` | Add ReconcileSuite to `mind check all` |
| `cmd/doctor.go` | Add staleness diagnostics |
| `main.go` (or `cmd/root.go`) | Centralize repo wiring (convergence recommendation 3) |

### Service Dependencies

The architect MUST define the precise service API and dependency injection structure. The following is the anticipated dependency pattern:

- `ReconciliationService` depends on: `ConfigRepo`, `DocRepo`, `LockRepo`
- `ReconcileSuite` depends on: `ReconciliationService` (or its results)
- Modified `cmd/status.go` depends on: `LockRepo` (read-only)
- Modified `cmd/doctor.go` depends on: `LockRepo` (read-only) or `ReconciliationService`

---

## Traceability

| Requirement | Source |
|-------------|--------|
| FR-51 through FR-56 | BP-06 Section 7 (Integration Points), BP-08 Section 3 (Phase 1.5 Scope) |
| FR-57 through FR-59 | BP-06 Section 2 (Hash Computation), Convergence Rec. 1.1 (no normalization) |
| FR-60 through FR-64 | BP-06 Section 3 (Dependency Graph) |
| FR-65 through FR-69 | BP-06 Section 4 (Staleness Propagation Algorithm) |
| FR-70 through FR-76 | BP-06 Section 5 (Lock File Lifecycle), Convergence Rec. 1.4 (is_stub via DocRepo) |
| FR-77 through FR-81 | BP-06 Section 7 (Integration Points), BP-08 Section 3 (Integrations table) |
| FR-82 | BP-06 Section 7, Convergence Rec. 1.5 (exit code 4) |
| FR-83 through FR-84 | BP-06 Section 3 (Edge Declaration in mind.toml), Convergence Rec. 2 step 2 |
| FR-85 | BP-06 Section 6 Phase 4 (Detect Undeclared) |
| FR-86 through FR-87 | BP-06 Section 8 (Performance Targets) |
