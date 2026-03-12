# Convergence Analysis: Reconciliation Engine Implementation Strategy

**Topic**: Phase 1.5 Reconciliation Engine -- optimal implementation approach, sequencing, and integration strategy for adding hash-based staleness tracking to the existing Phase 1 codebase
**Date**: 2026-03-11
**Personas**: Architect, Pragmatist, Critic
**Variant**: Standard
**Effective Persona Count**: 2.7 (see Diversity Audit)

---

## Phase 2: Opening Positions

### Position A -- The Architect

**Core Recommendation**: Fix the repo wiring tech debt first (1 focused session), then implement Phase 1.5 bottom-up following the 4-layer architecture strictly. The reconciliation engine is a new domain concept that must integrate cleanly at every layer; attempting to add it atop a wiring pattern the architecture itself flags as wrong invites compounding structural debt.

**Detailed Position**:

1. **Tech debt sequencing**: Fix SHOULD items *before* Phase 1.5, prioritized by coupling impact:
   - **(P0) Centralize repo wiring in main.go.** Every command handler currently creates its own repos (`fs.NewDocRepo(root)`, `fs.NewIterationRepo(root)`, etc.). Phase 1.5 adds `LockRepo` and `ReconciliationService` as new dependencies. If wiring stays in handlers, every integration point (`status.go`, `check.go`, `doctor.go`, `reconcile.go`) must independently construct these. This is the difference between modifying 1 file vs. 4+ files for every repo addition going forward.
   - **(P1) Rename `--project` to `--project-root`.** Trivial rename but the longer it lives, the more scripts and muscle memory depend on the wrong name. Phase 1.5 is a natural API break point.
   - **(P2) DoctorService delegation to ValidationService.** Phase 1.5 adds reconciliation diagnostics to doctor. If DoctorService already delegates to ValidationService, adding a reconciliation check is one method call. If it reimplements, we need to duplicate reconciliation check logic in DoctorService.
   - **(Defer) GoDoc comments, search abstraction.** These are hygiene items with zero bearing on Phase 1.5 correctness.

2. **Implementation sequence within Phase 1.5** (bottom-up, layer by layer):
   - Step 1: Domain types (`LockFile`, `LockEntry`, `ReconcileResult`, `Edge`, `EdgeType`, `Graph` -- all pure, stdlib-only)
   - Step 2: `internal/reconcile/hash.go` (SHA-256, mtime fast-path)
   - Step 3: `internal/reconcile/graph.go` (adjacency list, cycle detection)
   - Step 4: `internal/reconcile/propagate.go` (staleness BFS with depth limit)
   - Step 5: `internal/reconcile/engine.go` (orchestration of phases 1-6 from BP-06)
   - Step 6: `internal/repo/fs/lock_repo.go` + `internal/repo/mem/lock_repo.go` + interface addition
   - Step 7: `internal/service/reconciliation.go`
   - Step 8: `cmd/reconcile.go`
   - Step 9: Integration -- modify `status.go`, `check.go`, `doctor.go`
   - Step 10: `mind.toml` schema extension for `[[graph]]` edges

3. **Integration strategy**: The reconciliation engine is a self-contained subsystem consumed by existing commands via its service layer. The `ReconciliationService` should follow the same constructor injection pattern as `ProjectService`, `ValidationService`, etc. Existing commands call service methods and render results -- the reconciliation engine fits this pattern exactly.

4. **Blueprint adaptation needed**:
   - BP-06 shows `internal/reconcile/hash.go` doing `os.Open()` directly. This violates the repo pattern. Hash computation should either go through `DocRepo.Read()` or the reconcile package should accept an `io.Reader` / file path resolver interface.
   - BP-03 `mind.lock` schema includes `depends_on` per entry, which is redundant with the `dependency_graph` array. The implementation should normalize: store the graph once, compute `depends_on` at read time.
   - The `[[graph]]` TOML section needs a corresponding domain type (`GraphEdge`) and `Config` struct field. The current `domain.Config` does not have a `Graph` field.
   - BP-06 Section 4 propagation pseudocode uses the `sourceID` for the reason string at all depths, but the reason should track the *root* changed document, not the intermediate node in transitive propagation. The implementation should preserve the original root.

**Evidence Registry**:
- [E-A1] `cmd/status.go` lines 37-44: repo wiring pattern repeated in every handler
- [E-A2] `cmd/check.go` lines 68-72, 97-99, 124-128, 152-157: identical repo construction 4 times in one file
- [E-A3] Architecture doc Section "Constructor Injection in main.go": "All dependency wiring happens explicitly in main.go"
- [E-A4] `docs/state/current.md`: "SHOULD: Repo wiring in command handlers instead of main.go (C-10 deviation, acknowledged)"
- [E-A5] BP-06 Section 2: `HashFile()` calls `os.Open()` directly, bypassing repo pattern
- [E-A6] `internal/repo/interfaces.go`: 5 repo interfaces, none for lock file operations
- [E-A7] Retrospective: "DoctorService would benefit from delegating to the existing validation suites"
- [E-A8] `domain/project.go`: Config struct would need a `Graph []GraphEdge` field

**Risk Assessment**:
- Risk: Fixing wiring before Phase 1.5 may introduce regressions in 395 passing tests. *Mitigation*: The wiring change is purely structural (move `fs.New*()` calls from handlers to main.go), no logic changes. Tests that use `mem/` repos are unaffected.
- Risk: Bottom-up approach means integration issues surface late. *Mitigation*: Integration is step 9 of 10; each prior step produces independently testable units. Bottom-up is actually lower risk for a well-specified system.
- Risk: Overly strict layer adherence may slow implementation. *Mitigation*: Phase 1 proved that layer discipline pays off (domain at 100% coverage, validate at 90.7%). The cost is manageable for a known-complexity system.

---

### Position B -- The Pragmatist

**Core Recommendation**: Interleave tech debt fixes with Phase 1.5 implementation, tackling each fix at the natural point where it would otherwise cause friction. Start with a vertical slice that delivers `mind reconcile` end-to-end within 2-3 sessions, then widen to integrations.

**Detailed Position**:

1. **Tech debt sequencing**: Fix each item *when it creates friction*, not as a separate pre-pass:
   - Wiring centralization: Fix when implementing `cmd/reconcile.go` (you need to add a new repo and service anyway -- refactor wiring as part of that commit).
   - `--project` rename: Do it in the same commit that adds reconcile command flags. One flag-related commit is simpler to review than two.
   - DoctorService delegation: Fix when adding reconciliation diagnostics to doctor. The alternative (reimplementing reconcile checks in DoctorService) will be obviously wrong at that moment.
   - GoDoc comments: Batch at the end with a lint pass.
   - Search abstraction: Not relevant to Phase 1.5 at all. Defer entirely.

2. **Implementation sequence within Phase 1.5** (vertical slice first):
   - Step 1: Domain types + `internal/reconcile/hash.go` + basic tests (immediate feedback)
   - Step 2: `internal/reconcile/graph.go` + `internal/repo/fs/lock_repo.go` (need I/O to test end-to-end)
   - Step 3: `internal/reconcile/engine.go` + `internal/service/reconciliation.go` (first working `Reconcile()`)
   - Step 4: `cmd/reconcile.go` -- deliver `mind reconcile` as working command (milestone: demo-able)
   - Step 5: Fix repo wiring centralization while adding ReconciliationService to existing commands
   - Step 6: `internal/reconcile/propagate.go` + staleness tests
   - Step 7: Integration with `mind status`, `mind check all`, `mind doctor`
   - Step 8: `mind reconcile --check`, `--force`, `--graph` flags
   - Step 9: Performance validation and edge cases

3. **Integration strategy**: Same as architect (service layer with constructor injection), but pragmatically accept that the initial `cmd/reconcile.go` can wire its own repos. Centralize in step 5 when integrating with existing commands -- that is the natural refactoring moment, not before.

4. **Blueprint adaptation needed**:
   - Accept BP-06's hash.go using `os.Open()` directly. The `internal/reconcile/` package is infrastructure-adjacent. Creating a `HashRepo` interface for the sake of purity would add an abstraction nobody will ever swap. SHA-256 of file content is not going to have an in-memory implementation that differs meaningfully from the real one.
   - The `mind.lock` `depends_on` field per entry is useful for quick lookups without graph traversal. Keep it despite the redundancy -- it costs ~200 bytes and saves a graph lookup on every `mind status` call.
   - Add `[[graph]]` parsing to `ConfigRepo` by extending the existing TOML parser. This is a config concern, not a domain concern -- `ConfigRepo.ReadProjectConfig()` already handles complex nested TOML.

**Evidence Registry**:
- [E-B1] Phase 1 delivered 60+ files in a single iteration. Separating tech debt into a pre-pass adds coordination overhead for a single developer.
- [E-B2] `cmd/status.go`, `cmd/doctor.go`, `cmd/check.go`: Each handler follows the same resolve-root/create-repos/call-service/render pattern. The pattern is correct even if the wiring location is suboptimal.
- [E-B3] BP-08 Section 3 acceptance criteria: 14 acceptance criteria, none of which require wiring centralization first.
- [E-B4] Retrospective "Discovered Patterns": "The thin-handler pattern is consistent and easy to review" -- this works for reconcile too.
- [E-B5] BP-06 Section 8 performance targets: <200ms full reconcile, <50ms incremental -- these are generous for Go stdlib. No need for premature optimization.
- [E-B6] `internal/reconcile/` is a new package. There is no existing code to refactor -- it is all greenfield, so the implementation sequence within reconcile/ is flexible.
- [E-B7] `domain/project.go`: Config struct uses `go-toml/v2` tags. Adding Graph field is a one-line struct change + parser extension.

**Risk Assessment**:
- Risk: Interleaving debt fixes may produce messy commit history. *Mitigation*: Single developer, feature branch. Commit history can be cleaned up before merge.
- Risk: Vertical slice may produce an initial implementation that needs significant rework at integration time. *Mitigation*: The service layer interface is well-defined. `ReconciliationService.Reconcile()` signature is stable regardless of wiring approach.
- Risk: Deferring wiring fix may mean Phase 1.5 ships with the same anti-pattern. *Mitigation*: Step 5 explicitly addresses this. The integration step *is* the fix point.

---

### Position C -- The Critic

**Core Recommendation**: Neither the architect's clean-room nor the pragmatist's interleaved approach addresses the fundamental risk: the blueprints (BP-06, BP-08, BP-03) contain internal inconsistencies that will surface as implementation bugs if followed literally. The first task must be reconciling the blueprints themselves, then implementing with an emphasis on testing the integration points that bridge Phase 1 and Phase 1.5.

**Detailed Position**:

1. **Tech debt sequencing**: Fix only the wiring centralization, and do it first. Everything else is noise:
   - Wiring centralization is the only SHOULD item that directly impacts Phase 1.5 implementation cost. Fix it.
   - `--project` rename is cosmetic. It can ship in any commit between now and Phase 2. It does not affect reconciliation.
   - DoctorService delegation is a COULD item. The current doctor works. The Phase 1.5 doctor integration can add reconciliation checks the same way DoctorService adds all its other checks -- as a new `checkReconciliation()` method. Refactoring DoctorService to delegate is a Phase 2 concern when TUI adds another consumer.
   - GoDoc and search abstraction: irrelevant to Phase 1.5.

2. **Implementation sequence within Phase 1.5** (test-first, integration-focused):
   - Step 1: Resolve blueprint inconsistencies (document decisions, do not just discover them mid-implementation)
   - Step 2: Domain types with comprehensive tests (property-based if feasible)
   - Step 3: `internal/reconcile/graph.go` with cycle detection -- this is the most algorithmically complex piece and the most likely source of bugs. Test it first, in isolation.
   - Step 4: `internal/reconcile/hash.go` -- straightforward but must handle edge cases (empty files, binary files, symlinks). Each edge case from BP-06 Section 9 needs a test.
   - Step 5: `internal/reconcile/propagate.go` -- the transitive propagation algorithm must be tested with graphs of various topologies (linear chain, diamond, wide fan-out, depth-limit boundary)
   - Step 6: `internal/repo/fs/lock_repo.go` + `mem/lock_repo.go` -- the lock file round-trip (write then read produces identical structure) is a critical correctness property
   - Step 7: `internal/reconcile/engine.go` -- integration test that wires all prior components
   - Step 8: `internal/service/reconciliation.go` + `cmd/reconcile.go`
   - Step 9: Integration with existing commands (status, check, doctor)
   - Step 10: Performance benchmarks against BP-06 Section 8 targets

3. **Integration strategy**: Use the existing `internal/validate/` check framework for the `mind check all` integration. Define a `ReconcileSuite()` that follows the `DocsSuite()`/`RefsSuite()` pattern. This leverages the proven framework rather than building a parallel check mechanism. For `mind status`, add an optional `StalenessPanel` to `ProjectHealth` (nil when no lock file exists). For `mind doctor`, add a `checkReconciliation()` method following the existing DoctorService pattern.

4. **Blueprint inconsistencies that must be resolved before implementation**:
   - **BP-06 vs. BP-03 on staleness propagation through `validates` edges**: BP-06 Section 3 says all three edge types propagate staleness. BP-03 Section 2 "Staleness Algorithm" says "Staleness propagates transitively through `informs` and `requires` edges but not through `validates` edges." These directly contradict. Resolution needed.
   - **BP-06 hash.go vs. architecture layer rules**: BP-06 shows `HashFile()` calling `os.Open()` in `internal/reconcile/hash.go`. Architecture doc says only `internal/repo/fs/` calls `os.ReadFile`, `os.Stat`, etc. The reconcile package is in the service layer (between service and repo), not in infrastructure. Where does it sit in the 4-layer model?
   - **BP-08 hash.go description mentions normalization**: BP-08 Section 3 package list says hash.go does "SHA-256 computation over normalized file content (strip trailing whitespace, normalize line endings)." BP-06 Section 2 explicitly says "No normalization, no line-ending conversion, no BOM stripping." Direct contradiction.
   - **BP-03 `mind.lock` field `is_stub`**: The lock entry includes an `is_stub` field. This duplicates stub detection from `mind check docs`. The reconciliation engine should not reimplement stub detection -- it should store only hash/mtime/staleness data. Stub detection is a separate concern.
   - **Exit code 4**: BP-08 acceptance criteria says `mind reconcile --check` exits 4 on stale. Architecture doc defines 4 exit codes (0, 1, 2, 3). Exit code 4 is an undocumented extension. Needs explicit decision: is this a new exit code or should stale map to exit 1?

**Evidence Registry**:
- [E-C1] BP-06 Section 3: "All three edge types propagate staleness"
- [E-C2] BP-03 Section 2 "Staleness Algorithm": "Staleness propagates transitively through `informs` and `requires` edges but not through `validates` edges"
- [E-C3] BP-08 Section 3 hash.go: "normalized file content (strip trailing whitespace, normalize line endings)"
- [E-C4] BP-06 Section 2: "No normalization, no line-ending conversion, no BOM stripping"
- [E-C5] Architecture doc Layer 4 rules: "These are the only packages that call os.ReadFile, os.Stat, filepath.Walk"
- [E-C6] BP-06 Section 2 Go implementation sketch: `func HashFile(path string)` calls `os.Open(path)`
- [E-C7] BP-03 `mind.lock` schema: `is_stub` field in lock entries
- [E-C8] Architecture doc "Exit Code Strategy": "Four exit codes: 0, 1, 2, 3"
- [E-C9] BP-08 acceptance criteria: "exits 4 when stale documents exist"
- [E-C10] `internal/validate/check.go`: Suite/Check/CheckFunc framework would naturally host a ReconcileSuite

**Risk Assessment**:
- Risk: Blueprint reconciliation delays implementation start. *Mitigation*: 5 inconsistencies can be resolved in a 30-minute decision document. The alternative (discovering them during implementation) costs more in rework.
- Risk: Test-first approach may over-invest in edge cases that never occur in practice. *Mitigation*: BP-06 Section 9 enumerates exactly 10 edge cases. Each needs at most 1 test. This is bounded, not open-ended.
- Risk: Integration via existing check framework may not fit reconciliation's output shape. *Mitigation*: The check framework is generic (`CheckFunc` returns pass/fail). Reconciliation results can be projected into this shape: "check: no stale documents" pass/fail.

---

## Phase 2.5: Diversity Audit + Tension Extraction

### Effective Persona Count: 2.7

The three personas produce substantively different positions, but the architect and pragmatist share more common ground than either shares with the critic. Specifically:

- **Architect and Pragmatist** agree on: 4-layer architecture adherence, service layer as integration point, domain types as starting point, constructor injection pattern. They disagree primarily on *timing* of tech debt fixes and *sequencing* within Phase 1.5.
- **Architect and Critic** agree on: wiring centralization should happen before Phase 1.5 integration, bottom-up-ish implementation order, need for blueprint adaptation. They disagree on whether hash.go's I/O is a layer violation and on scope of pre-work.
- **Pragmatist and Critic** agree on: only wiring centralization matters among tech debt items, DoctorService delegation is not urgent. They disagree on implementation sequencing (vertical slice vs. test-first integration-focused) and on the severity of blueprint inconsistencies.

The effective persona count is 2.7 (not 3.0) because the architect-pragmatist pair has ~40% position overlap, while the critic is genuinely orthogonal on the blueprint inconsistency axis.

### Tension Matrix

| Tension | Personas | Severity | Nature |
|---------|----------|----------|--------|
| T1: Pre-fix vs. interleave tech debt | Architect vs. Pragmatist | SHOULD | Sequencing |
| T2: Layer purity of hash.go I/O | Architect vs. Pragmatist + Critic | SHOULD | Architecture |
| T3: Blueprint reconciliation as prerequisite | Critic vs. Architect + Pragmatist | MUST | Correctness |
| T4: Vertical slice vs. bottom-up | Pragmatist vs. Architect + Critic | COULD | Methodology |
| T5: `depends_on` redundancy in lock file | Architect vs. Pragmatist | COULD | Data design |
| T6: DoctorService refactoring timing | Architect vs. Pragmatist + Critic | COULD | Scope |
| T7: `is_stub` in lock entries | Critic (unique) | SHOULD | Separation of concerns |
| T8: Exit code 4 as new code | Critic (unique) | MUST | Contract |

---

## Phase 3: Cross-Examination

### Architect challenges Pragmatist

**Challenge P-1 (MUST)**: Your vertical slice approach delivers a working `mind reconcile` early but defers repo wiring centralization to step 5. Between steps 1-4, the new `cmd/reconcile.go` will wire its own repos -- *adding* to the very anti-pattern you claim to fix later. Evidence [E-A2] shows `check.go` already has 4 copies of identical repo construction. Every step that adds another copy makes the eventual refactoring larger and riskier. How do you prevent step 5 from becoming a "big bang" refactor?

**Challenge P-2 (SHOULD)**: You accept BP-06's `HashFile()` calling `os.Open()` directly in `internal/reconcile/hash.go`, arguing that a `HashRepo` interface is an abstraction nobody will swap. But the architecture doc [E-A3] explicitly states: "These are the only packages that call os.ReadFile, os.Stat, filepath.Walk" about `internal/repo/fs/`. Placing `os.Open()` in `internal/reconcile/` creates a *second* package that does filesystem I/O outside the repo layer. This is not a theoretical concern -- it breaks the promise that makes in-memory testing work. How do you test `HashFile()` in CI without touching the filesystem?

**Challenge P-3 (COULD)**: Your step ordering puts `propagate.go` (step 6) *after* `engine.go` + service (step 3). How does `engine.go` work without propagation? Do you stub it? If so, you are writing throwaway code in step 3 that gets replaced in step 6 -- the very waste that bottom-up avoids.

### Architect challenges Critic

**Challenge C-1 (SHOULD)**: You recommend resolving blueprint inconsistencies as step 1 before any code. For a single-developer project, written decisions are overhead. The developer who resolves the inconsistencies *is* the developer who implements the code. A decision captured as a code comment or test assertion is more durable than a decision captured as a separate document. Why document decisions separately when the code *is* the decision?

**Challenge C-2 (COULD)**: You propose using the existing `Suite`/`Check`/`CheckFunc` framework for reconciliation integration with `mind check all`. But the reconciliation engine produces a `ReconcileResult` with `Changed`, `Stale`, `Missing`, `Stats` -- a much richer output than a simple pass/fail `CheckResult`. Projecting this into `CheckResult` loses information. Would a dedicated reconciliation panel (separate from the check suites) be more appropriate, as BP-06 Section 7 shows?

### Pragmatist challenges Architect

**Challenge A-1 (SHOULD)**: You want to fix wiring centralization *before* starting Phase 1.5. This means modifying `main.go` and all 12 command files in `cmd/` to pass repos through a different mechanism -- with no functional change to the user. That is a refactoring-only commit touching 12+ files, which must be reviewed, tested, and merged before Phase 1.5 even starts. For a single developer on a feature branch, this is pure overhead. The natural moment to centralize is when adding a new dependency (ReconciliationService) forces you to touch those files anyway.

**Challenge A-2 (MUST)**: Your step 10 (last step) is adding `[[graph]]` edge parsing to `mind.toml`. But steps 3-5 (graph.go, propagate.go, engine.go) all depend on graph edges as input. Where do the edges come from in steps 3-9 if the parser does not exist yet? Your bottom-up approach has a dependency inversion: the highest-level concern (config parsing) is required by the lowest-level components (graph construction).

**Challenge A-3 (COULD)**: You recommend fixing `--project` rename as P1 priority and DoctorService delegation as P2, claiming they reduce Phase 1.5 friction. The flag rename has zero impact on Phase 1.5 (reconcile command does not use `--project`; it uses the resolved root). DoctorService delegation also has zero impact -- Phase 1.5 adds `checkReconciliation()` as a new method regardless of delegation pattern. Your priority ordering optimizes for architectural purity, not Phase 1.5 delivery.

### Pragmatist challenges Critic

**Challenge C-3 (SHOULD)**: You identify 5 blueprint inconsistencies and demand they be resolved before implementation. But two of them (T3: `validates` edge propagation, T8: exit code 4) are *design decisions*, not bugs. The blueprints present options; the implementation must choose one. These are resolved by implementing the chosen behavior and documenting it in code. Framing them as "inconsistencies that block implementation" inflates their severity.

**Challenge C-4 (COULD)**: Your test-first, integration-focused approach puts graph cycle detection as step 3 (before hash.go). This means you cannot test the full reconciliation pipeline until step 7 (engine.go). The pragmatist's vertical slice delivers a testable pipeline 4 steps earlier. For a single developer who needs feedback loops, which approach reduces risk more?

### Critic challenges Architect

**Challenge A-4 (MUST)**: You describe hash.go's `os.Open()` as a layer violation and propose using `DocRepo.Read()` or an `io.Reader` interface instead. But `DocRepo.Read()` returns `[]byte` -- the entire file content loaded into memory. For hash computation, streaming through `io.Copy(h, f)` is the correct approach (handles large files without memory pressure). The repo pattern as designed is incompatible with streaming hashing. Your solution (use DocRepo.Read) is worse than the problem (hash.go calls os.Open). How do you resolve this without redesigning the repo interface?

**Challenge A-5 (SHOULD)**: You list 10 implementation steps but do not mention where `mind.toml` schema extension for `[[graph]]` edges happens relative to the domain type definition. The `Config` struct in `domain/project.go` needs a `Graph` field. The `ConfigRepo` parser needs to handle `[[graph]]`. The `internal/reconcile/graph.go` needs to consume these edges. This is a cross-cutting concern that does not fit cleanly into your bottom-up layer-by-layer sequence. Where does it go?

### Critic challenges Pragmatist

**Challenge P-4 (MUST)**: You dismiss the blueprint inconsistency about hash normalization (BP-08 says normalize, BP-06 says do not normalize) as something to resolve during implementation. This is a correctness decision that affects *every hash ever computed*. If you implement with normalization and later switch to raw, every lock file becomes invalid. If you implement without normalization and later add it, same problem. This decision must be explicit and irrevocable before any hash is computed. Treating it as an implementation detail is the definition of accidental complexity.

**Challenge P-5 (SHOULD)**: Your vertical slice delivers a working `mind reconcile` by step 4, but without staleness propagation (step 6). A reconcile command that computes hashes but does not propagate staleness is a half-feature. It updates `mind.lock` with hashes but cannot answer the core question: "which documents are stale?" This is not a useful milestone -- it is a demo that misleads about feature completeness.

---

## Phase 4: Rebuttal & Refinement

### Architect responds to challenges

**A-1 response (Partial Accept)**: The pragmatist is correct that fixing wiring before Phase 1.5 has zero user-visible benefit and touches 12+ files purely for structural improvement. *Concession*: I accept that the wiring fix should happen *during* Phase 1.5, specifically at the integration step (not before, not after). However, I insist it happens before the `cmd/reconcile.go` commit, not after -- the reconcile command should be the first handler to use centralized wiring, establishing the pattern for the integration changes to status/check/doctor.

**A-2 response (Concede)**: Valid point. My sequencing placed `[[graph]]` parsing last (step 10) while graph construction (step 3) depends on it. *Concession*: Graph edge parsing from `mind.toml` should move to step 2, immediately after domain types. The domain types include `GraphEdge`; the config parser extension follows naturally. Steps 3-5 then have real input data. Revised sequence:
- Step 1: Domain types
- Step 2: Config extension (`[[graph]]` parsing)
- Step 3: graph.go (consumes parsed edges)
- (rest unchanged)

**A-3 response (Rebut)**: The `--project` rename has low Phase 1.5 impact, agreed. But DoctorService delegation *does* matter: if doctor already delegates to ValidationService, adding reconciliation checks to `mind check all` automatically surfaces them in `mind doctor` via delegation. Without delegation, we add reconciliation checks in two places. However, I accept the pragmatist's framing that this is a COULD, not a SHOULD, for Phase 1.5 timeline.

**A-4 response (Partial Accept)**: The critic correctly identifies that `DocRepo.Read()` returns `[]byte`, which is incompatible with streaming. *Concession*: hash.go should be allowed to do filesystem I/O directly, but it should accept a path resolver or root parameter rather than constructing paths itself. The function signature should be `HashFile(absPath string) (string, error)` and the caller (engine.go or service) resolves the path from the project root + document path. This keeps the I/O in reconcile/ but prevents it from knowing about project structure.

**A-5 response (Concede)**: My revised sequence (A-2 response) addresses this. `[[graph]]` parsing moves to step 2. The cross-cutting nature is acknowledged.

### Pragmatist responds to challenges

**P-1 response (Partial Accept)**: The architect is right that steps 1-4 would add another handler with inline repo wiring. *Concession*: I accept that `cmd/reconcile.go` should not ship with inline wiring if wiring centralization is planned for step 5. Compromise: implement `reconcile.go` in step 4 using a `setupReconcileService(root string)` helper function that can be migrated to centralized wiring in step 5 without changing the handler. This avoids adding to the anti-pattern while maintaining the vertical slice timeline.

**P-2 response (Rebut)**: The architecture doc's claim that "only repo/fs/ calls os.ReadFile" is an aspiration documented in Phase 1, before reconciliation existed. Phase 1.5 introduces a new package (`internal/reconcile/`) that is infrastructure-adjacent -- it is a computational engine that operates on files. Forcing it through a repo interface adds complexity without testability benefit. The hash computation *can* be tested with real temp files in tests (as Go's testing package makes trivial). An `io.Reader`-based interface is a reasonable middle ground -- accept a reader, not a path -- but I would not route through DocRepo.

**P-3 response (Concede)**: Valid. Delivering engine.go (step 3) before propagate.go (step 6) means engine.go is incomplete. *Concession*: Move propagate.go to step 3 (before engine.go). Revised sequence:
- Step 1: Domain types + hash.go
- Step 2: graph.go + lock_repo.go
- Step 3: propagate.go
- Step 4: engine.go + service
- Step 5: cmd/reconcile.go (working command)
- Step 6: Wiring centralization
- Step 7: Integration with status/check/doctor
- Step 8: Flags (--check, --force, --graph)
- Step 9: Performance + edge cases

**P-4 response (Accept)**: The critic is right. Hash normalization is a one-way door. *Concession*: The normalization decision must be resolved before step 1 (hash.go implementation). My position: follow BP-06 (no normalization). Raw bytes is simpler, deterministic, and the false-positive cost is low. Document this decision as a code comment in hash.go. This is a 5-minute decision, not a blocking document.

**P-5 response (Partial Accept)**: Fair point that reconcile-without-propagation is a half-feature. However, it is a *useful* half-feature: it creates `mind.lock` with hashes and detects changed files. Staleness propagation adds "why" but the "what changed" question is already answered. *Concession*: With the revised sequence (P-3 concession), propagation moves before the first working command, making this objection moot.

### Critic responds to challenges

**C-1 response (Partial Accept)**: The architect is right that for a single developer, code *is* documentation. *Concession*: Blueprint inconsistencies can be resolved as code comments and test assertions rather than a separate decision document. However, the 5 inconsistencies must still be *identified and resolved explicitly* before the affected code is written. A comment like `// Decision: no normalization (BP-06 Section 2 overrides BP-08 Section 3)` takes 10 seconds and prevents future confusion.

**C-2 response (Rebut)**: The `CheckResult` limitation is real, but the check framework is extensible. A `ReconcileSuite()` can produce multiple `CheckResult` entries: one per-document staleness check ("doc:spec/architecture is stale: pass/fail"), one for cycle detection, one for missing documents. The detailed `ReconcileResult` is used by `cmd/reconcile.go` for rich output. The projected check results are used by `mind check all` for unified reporting. Both consumers get what they need. BP-06 Section 7 shows a dedicated panel for `mind check all`, which is what a ReconcileSuite section would render as.

**C-3 response (Partial Accept)**: The pragmatist is right that `validates` edge propagation and exit code 4 are design decisions, not bugs. *Concession*: Reclassify these as "design decisions requiring explicit choice" rather than "inconsistencies that block implementation." However, they still must be decided before the affected code is written. They are not optional decisions -- the code must behave one way or the other.

**C-4 response (Rebut)**: The pragmatist conflates "testable pipeline" with "working CLI command." My approach has a testable pipeline at step 7 (engine.go as integration test), 2 steps after the pragmatist's revised step 5. But my earlier steps (3, 4, 5, 6) each produce independently tested, correct components. The pragmatist's vertical slice produces a testable pipeline earlier but with less confidence in individual component correctness. For a reconciliation engine where correctness is the primary concern (wrong staleness propagation is worse than no staleness propagation), component-level confidence matters more than pipeline-level speed.

---

## Phase 5: Convergence Synthesis

### 1. Executive Summary

The reconciliation engine implementation should begin with resolving 5 identified blueprint inconsistencies as inline decisions (code comments, not documents), fix repo wiring centralization concurrently with the first integration step, and follow a component-first sequence that delivers domain types and config parsing first, then computational core (hash, graph, propagation), then service and command layers. Tech debt items other than wiring centralization should be addressed opportunistically. The blueprints are highly prescriptive and largely correct, but the hash normalization contradiction and `validates` edge propagation semantics must be resolved before coding begins.

### 2. Convergence Map

**Full Consensus (all 3 personas agree)**:
- Domain types are the starting point (pure Go, testable, foundational)
- `ReconciliationService` follows constructor injection pattern like existing services
- Wiring centralization is the only tech debt item that materially affects Phase 1.5
- GoDoc and search abstraction can be deferred past Phase 1.5
- `internal/reconcile/` is a new package (not modification of existing packages)
- `mind.toml` needs `[[graph]]` section and `Config` struct needs `Graph` field
- Lock file round-trip correctness is a critical property to test
- BP-06 Section 9 edge cases each need at least one test

**Majority Consensus (2 of 3 agree)**:
- hash.go may do filesystem I/O directly (Pragmatist + Critic vs. Architect; Architect conceded with path-resolver compromise)
- Blueprint inconsistencies should be resolved as code comments, not separate documents (Architect + Pragmatist vs. Critic; Critic conceded)
- Wiring centralization should happen during Phase 1.5, not before (Pragmatist + Critic vs. Architect; Architect conceded)
- DoctorService delegation is a COULD, not a prerequisite (Pragmatist + Critic vs. Architect; Architect conceded)
- Propagation code should precede engine orchestration code (Architect + Critic vs. Pragmatist; Pragmatist conceded)

**Unresolved Disagreements**:
- **ReconcileSuite vs. dedicated panel for `mind check all`**: Critic favors projecting into CheckResult via ReconcileSuite; Architect favors a dedicated reconciliation section. Both approaches work. *Recommended resolution*: Use ReconcileSuite for the check framework (maintains pattern consistency) but allow the renderer to display reconciliation results with richer formatting than standard checks.
- **`depends_on` field in lock entries**: Architect considers it redundant; Pragmatist considers it useful for quick lookups. *Recommended resolution*: Include it. Storage cost is negligible and it simplifies the `mind status` read path.
- **Vertical slice vs. bottom-up within reconcile/**: After concessions, the sequences converged significantly. The remaining difference is whether hash.go or graph.go comes first. *Recommended resolution*: hash.go first (simpler, provides immediate feedback).

### 3. Decision Matrix

| Criterion | Weight | Option A: Pre-fix + Bottom-up | Option B: Interleaved + Vertical | Option C: Blueprint-first + Test-driven |
|-----------|--------|-------------------------------|-----------------------------------|-----------------------------------------|
| **Delivery speed** | 25% | 3/5 -- Extra pre-pass slows start | 5/5 -- Fastest to working command | 3/5 -- Blueprint reconciliation adds upfront cost |
| **Correctness confidence** | 30% | 4/5 -- Layer discipline catches integration bugs | 3/5 -- Late propagation integration risks | 5/5 -- Test-first validates each component |
| **Architecture alignment** | 15% | 5/5 -- Strict 4-layer adherence | 3/5 -- Pragmatic deviations (hash.go I/O) | 4/5 -- Mostly strict, hash.go pragmatic |
| **Maintainability** | 15% | 5/5 -- Clean wiring from the start | 4/5 -- Wiring cleaned during integration | 4/5 -- Clean but no wiring focus |
| **Risk management** | 15% | 3/5 -- Bottom-up delays integration discovery | 4/5 -- Early pipeline reduces unknowns | 5/5 -- Blueprint inconsistencies caught first |
| **Weighted Score** | 100% | **3.85** | **3.85** | **4.20** |

**Winner**: Option C (Blueprint-first + Test-driven) wins on weighted score, primarily due to its emphasis on correctness confidence and risk management. However, the practical recommendation is a hybrid that takes the best elements from each position (see Recommendations below).

### 4. Key Insights

**Insight 1: Blueprint inconsistencies are real and consequential.** The 5 inconsistencies identified by the Critic are not theoretical -- they affect hash computation correctness (normalization), staleness semantics (validates propagation), data contract stability (is_stub in lock), and API contract (exit code 4). Discovering these during implementation costs more than discovering them during analysis. This is the most valuable contribution of the dialectical process.

**Insight 2: The wiring centralization is the only tech debt item that compounds.** Each new command or service integration adds another instance of the wiring anti-pattern. Phase 1.5 adds 4 integration points (reconcile, status, check, doctor). Fixing wiring during integration is not just desirable -- it is the cheapest point to do it, because those files must be modified anyway.

**Insight 3: hash.go's filesystem I/O is architecturally justified as infrastructure-adjacent.** The 4-layer model was designed for Phase 1 where all I/O is document reading. Reconciliation introduces a new I/O pattern (streaming hash computation) that does not fit the existing `DocRepo.Read() []byte` interface. The pragmatic resolution -- allow hash.go to accept an absolute path and do its own I/O -- is correct. The architectural fix (adding a streaming interface to DocRepo) is over-engineering for one use case.

**Insight 4: The implementation sequences converged after cross-examination.** After concessions, all three personas agree on approximately this order: domain types, config parsing, graph, hash, propagation, engine, service, command, integration. The disagreements were about granularity and testing emphasis, not fundamental ordering.

**Insight 5: `mind.lock` `is_stub` field crosses concern boundaries.** Stub detection is a documentation validation concern. Lock file is a staleness tracking concern. Including stub status in the lock file couples two independent subsystems. However, it is useful for `mind status` to show staleness and stub status from a single file read. The pragmatic resolution: include it but compute it from `DocRepo.IsStub()` during reconciliation, do not reimplement stub detection.

### 5. Concession Trail

| ID | Persona | Concession | Triggered By |
|----|---------|-----------|--------------|
| CON-1 | Architect | Wiring fix should happen during Phase 1.5 (at integration step), not before | Pragmatist challenge A-1 |
| CON-2 | Architect | `[[graph]]` parsing moved to step 2 (was step 10) | Pragmatist challenge A-2 |
| CON-3 | Architect | `--project` rename is low priority for Phase 1.5 | Pragmatist challenge A-3 |
| CON-4 | Architect | hash.go may do filesystem I/O with path parameter (not DocRepo.Read) | Critic challenge A-4 |
| CON-5 | Pragmatist | `cmd/reconcile.go` should use helper function, not inline wiring | Architect challenge P-1 |
| CON-6 | Pragmatist | propagate.go must precede engine.go in sequence | Architect challenge P-3 |
| CON-7 | Pragmatist | Hash normalization decision must be explicit before implementation | Critic challenge P-4 |
| CON-8 | Critic | Blueprint inconsistencies resolved as code comments, not separate documents | Architect challenge C-1 |
| CON-9 | Critic | `validates` propagation and exit code 4 are design decisions, not bugs | Pragmatist challenge C-3 |

### 6. Recommendations

#### Recommendation 1: Resolve the 5 blueprint inconsistencies before writing code
**Confidence**: 95% (HIGH)
**Falsifiability**: If implementation proceeds without resolving these and no rework occurs, this recommendation was wrong.

Decisions to make:
1. **Hash normalization**: Follow BP-06 -- no normalization, raw bytes. Simpler, deterministic.
2. **`validates` edge propagation**: Follow BP-06 -- all edge types propagate. The distinction matters for *reporting* (message text), not for propagation logic. Simpler algorithm, easier to restrict later if needed.
3. **hash.go I/O**: Allow direct `os.Open()` in `internal/reconcile/hash.go`. Accept absolute path as parameter. This is infrastructure-adjacent code, not service-layer logic.
4. **`is_stub` in lock entries**: Include it. Compute via `DocRepo.IsStub()` during reconciliation, do not reimplement stub detection.
5. **Exit code 4**: Accept as new exit code for staleness. Document in architecture doc alongside existing 0/1/2/3. This is a meaningful semantic distinction (stale artifacts vs. validation failure vs. runtime error).

#### Recommendation 2: Implement Phase 1.5 in this sequence
**Confidence**: 85% (HIGH)
**Falsifiability**: If a different sequence proves faster or catches bugs earlier in practice, this was wrong.

1. Domain types: `LockFile`, `LockEntry`, `ReconcileResult`, `GraphEdge`, `EdgeType` in `domain/`
2. Config extension: Add `Graph []GraphEdge` to `Config`, extend `ConfigRepo` TOML parsing for `[[graph]]`
3. `internal/reconcile/hash.go`: SHA-256, mtime fast-path, edge case handling
4. `internal/reconcile/graph.go`: Adjacency list, cycle detection, topological awareness
5. `internal/reconcile/propagate.go`: BFS downstream propagation with depth limit 10
6. `internal/repo/fs/lock_repo.go` + `internal/repo/mem/lock_repo.go` + interface in `interfaces.go`
7. `internal/reconcile/engine.go`: Full 6-phase orchestration from BP-06
8. `internal/service/reconciliation.go`: `ReconciliationService` with constructor injection
9. Centralize repo wiring in `main.go` (or root.go `PersistentPreRunE`) -- fix tech debt here
10. `cmd/reconcile.go`: `mind reconcile` with `--check`, `--force`, `--graph`
11. Integration: modify `cmd/status.go` (staleness panel), `cmd/check.go` (ReconcileSuite), `cmd/doctor.go` (stale findings)
12. Performance benchmarks against BP-06 Section 8 targets

#### Recommendation 3: Fix wiring centralization during step 9 (concurrent with Phase 1.5)
**Confidence**: 90% (HIGH)
**Falsifiability**: If wiring centralization during Phase 1.5 causes integration regressions that a pre-fix would have prevented, this was wrong.

Do not fix wiring before Phase 1.5. Do not defer it past Phase 1.5. Step 9 is the natural point: the new `ReconciliationService` must be wired, and existing commands must gain access to it. Centralizing all wiring at this moment minimizes total file modifications.

#### Recommendation 4: Defer all other tech debt past Phase 1.5
**Confidence**: 80% (MEDIUM)
**Falsifiability**: If deferred items cause bugs or friction during Phase 1.5, this was wrong.

- `--project` rename: cosmetic, no Phase 1.5 impact
- GoDoc comments: hygiene, no Phase 1.5 impact
- `docs search` abstraction: unrelated to reconciliation
- DoctorService delegation: add `checkReconciliation()` method following existing pattern; full delegation refactor can wait for Phase 2

#### Recommendation 5: Use ReconcileSuite for `mind check all` integration
**Confidence**: 75% (MEDIUM)
**Falsifiability**: If the CheckResult projection loses information that users need from `mind check all`, a dedicated panel would have been better.

Define `ReconcileSuite()` in `internal/validate/` following the `DocsSuite()`/`RefsSuite()` pattern. Project reconciliation results into check results: one check per stale document, one for cycle detection, one for missing documents. Allow the renderer to format the reconciliation section with richer output (staleness counts, dependency chains) than standard check results.

### 7. Meta-Analysis

#### Quality Rubric

| Dimension | Score | Justification |
|-----------|-------|---------------|
| **Dialectical Tension** | 4/5 | Genuine disagreements on sequencing, layer purity, and blueprint handling. The architect-pragmatist tension on wiring timing produced real concessions. Slight dock because all three personas share the same implementation domain, limiting paradigmatic diversity. |
| **Evidence Grounding** | 5/5 | All positions cite specific code files, line numbers, blueprint sections, and retrospective findings. No position relies on general principles without concrete evidence. |
| **Concession Quality** | 4/5 | 9 concessions tracked. Most are substantive (architect conceding on wiring timing, pragmatist conceding on propagation ordering). Two are minor (architect on flag rename priority, critic on document format). No persona refused to concede when evidence warranted it. |
| **Recommendation Actionability** | 5/5 | Each recommendation specifies what to do, when, and in what order. The 12-step implementation sequence is directly executable. Blueprint inconsistency resolutions are specific decisions, not vague principles. |
| **Falsifiability** | 4/5 | Each recommendation includes a falsifiability condition. Some conditions are slightly vague ("if rework occurs" -- how much rework?), but most are concrete and testable. |
| **Coverage** | 4/5 | All aspects of the topic are addressed: tech debt timing, implementation sequence, integration strategy, blueprint adaptation, performance, testing. Minor gap: no discussion of branch strategy (should Phase 1.5 be one branch or multiple?) or how to handle `mind.toml` migration for existing users who lack `[[graph]]` sections. |

**Overall Quality Score**: 4.33 / 5.00

### 8. Evidence Audit Summary

| Evidence ID | Source | Used By | Verified |
|-------------|--------|---------|----------|
| E-A1 | `cmd/status.go` lines 37-44 | Architect: wiring anti-pattern | Yes -- read file, confirmed repo construction in handler |
| E-A2 | `cmd/check.go` lines 68-72, 97-99, 124-128, 152-157 | Architect: 4x duplication | Yes -- read file, confirmed identical patterns |
| E-A3 | Architecture doc, "Constructor Injection in main.go" | Architect: wiring should be centralized | Yes -- architecture.md section confirmed |
| E-A4 | `docs/state/current.md` | Architect: acknowledged SHOULD item | Yes -- current.md read |
| E-A5 | BP-06 Section 2 Go sketch | Architect: hash.go I/O concern | Yes -- BP-06 read, HashFile uses os.Open |
| E-A6 | `internal/repo/interfaces.go` | Architect: no LockRepo exists | Yes -- interfaces.go read, 5 repos defined |
| E-A7 | Retrospective | Architect: DoctorService delegation | Yes -- retrospective.md read |
| E-A8 | `domain/project.go` | Architect: Config needs Graph field | Yes -- domain files read |
| E-B1 | Phase 1 delivery scope | Pragmatist: single-iteration delivery | Yes -- 60+ files confirmed via glob |
| E-B2 | `cmd/status.go`, `cmd/doctor.go` | Pragmatist: handler pattern consistency | Yes -- files read |
| E-B3 | BP-08 Section 3 acceptance criteria | Pragmatist: no wiring prerequisite | Yes -- BP-08 read |
| E-B4 | Retrospective "Discovered Patterns" | Pragmatist: thin-handler works | Yes -- retrospective.md read |
| E-B5 | BP-06 Section 8 | Pragmatist: generous perf targets | Yes -- BP-06 read |
| E-B6 | `internal/reconcile/` | Pragmatist: greenfield package | Yes -- no existing package found |
| E-B7 | `domain/project.go` | Pragmatist: Config struct extensible | Yes -- domain files read |
| E-C1 | BP-06 Section 3 | Critic: all edges propagate | Yes -- BP-06 Section 3 table confirmed |
| E-C2 | BP-03 Section 2 "Staleness Algorithm" | Critic: validates edges excluded | Yes -- BP-03 read, contradiction confirmed |
| E-C3 | BP-08 Section 3 package table | Critic: normalized content | Yes -- BP-08 hash.go description confirmed |
| E-C4 | BP-06 Section 2 | Critic: no normalization | Yes -- BP-06 Section 2 text confirmed |
| E-C5 | Architecture doc Layer 4 rules | Critic: repo-only I/O | Yes -- architecture.md read |
| E-C6 | BP-06 Section 2 Go sketch | Critic: HashFile calls os.Open | Yes -- BP-06 read |
| E-C7 | BP-03 `mind.lock` schema | Critic: is_stub field | Yes -- BP-03 lock schema read |
| E-C8 | Architecture doc "Exit Code Strategy" | Critic: 4 exit codes only | Yes -- architecture.md read |
| E-C9 | BP-08 acceptance criteria | Critic: exit code 4 for stale | Yes -- BP-08 Section 3 read |
| E-C10 | `internal/validate/check.go` | Critic: Suite/Check framework | Yes -- validate package confirmed via glob |

**All 25 evidence items verified against source material.**

---

*Generated by Conversation Analysis Moderator -- Mode C (Orchestrator-Invoked)*
*Analysis type: architecture_design | Standard variant | 3 personas*
