# Requirements Delta: Pre-Phase 3 Cleanup

**Iteration**: 004-pre-phase-3-cleanup
**Type**: COMPLEX_NEW
**Source**: Convergence analysis (score 4.3/5.0, docs/knowledge/pre-phase-3-cleanup-convergence.md)
**FR Range**: FR-125 through FR-139

---

## Scope Boundary

### In-Scope

- Migrate `deps.Deps` struct from concrete `*fs.` types to `repo.` interface types
- Fix transitive staleness propagation to preserve edge-type-specific reason strings at depth > 0
- Rename `--project` CLI flag to `--project-root` per api-contracts spec (FR-2)
- Remove `internal/repo/mem/` import of `internal/repo/fs` (inverse dependency)
- Change `Diagnostic.Status` from raw `string` to a typed `DiagnosticStatus` enum
- Replace `os.Exit()` calls in `cmd/` with error returns following Cobra best practices
- Add test coverage for `cmd/` exit code paths
- Add test coverage for `internal/render/` JSON output mode
- Update `docs/spec/architecture.md` to fix stale file references and missing components

### Out-of-Scope

- DoctorService delegation to ValidationService (deferred -- COULD priority, high churn for low Phase 3 impact)
- GenerateService repository injection for file creation (deferred -- direct I/O for file creation is acceptable per convergence concession)
- TUI component extraction from tab views (S-4, deferred -- cosmetic, no Phase 3 impact)
- DRY refactoring of artifact counting, filter bar, staleness map (deferred -- COULD priority, do opportunistically when touching files)
- TUI preview pane Glamour rendering (S-2, deferred -- TUI cosmetic improvement)
- TUI status bar cursor position (S-3, deferred -- TUI cosmetic improvement)
- TUI editor fallback behavior (S-1, convergence consensus: current TUI behavior is acceptable)
- Phase 3 features (MCP server, preflight, handoff)
- Any changes to domain business rules or validation check logic
- Any changes to existing CLI command behavior beyond flag rename and error return pattern

---

## Success Metrics

- All 374 existing tests continue to pass (zero regressions)
- `deps.Deps` struct contains zero concrete `*fs.` type references
- `internal/repo/mem/` has zero imports from `internal/repo/fs`
- `Diagnostic.Status` field uses a typed enum, not raw `string`
- `cmd/` package contains zero direct `os.Exit()` calls (all exit code logic flows through Cobra error handling)
- `cmd/` package has at least 1 test file covering exit code behavior for `check`, `reconcile`, and `status` commands
- `internal/render/` has at least 1 test file covering JSON output mode for `RenderHealth`, `RenderValidation`, and `RenderReconcileResult`
- `--project-root` flag is functional; `--project` is removed or aliased for backward compatibility
- Transitive staleness propagation at depth > 0 includes edge-type-specific reason strings (not generic "may be outdated")
- `docs/spec/architecture.md` contains no stale file references

---

## Unchanged Requirements

The following MUST NOT change:

- All 374 existing tests pass without modification (test logic may be updated only if the test validates a changed interface, e.g., flag rename)
- Domain layer purity (DC-1): `domain/` imports only Go stdlib
- Existing CLI command behavior: all commands produce identical output for identical inputs (except `--project` flag rename and error message formatting changes from os.Exit removal)
- Exit code semantics: 0 (success), 1 (validation failure), 2 (runtime error), 3 (config error), 4 (staleness)
- Validation check IDs, names, and pass/fail logic are unchanged
- JSON output structure for `--json` mode is unchanged (field names, nesting)
- TUI behavior and rendering are unchanged (beyond any updates needed for Deps interface migration)

---

## Domain Model Impact

### Modified Entity: Diagnostic

The `Diagnostic` struct in `domain/health.go` gains a typed status field:

- `Status string` changes to `Status DiagnosticStatus`
- New enum type `DiagnosticStatus` with values: `DiagPass`, `DiagFail`, `DiagWarn`
- The `Level` field (currently `CheckLevel`, JSON-excluded) may be consolidated with or derived from `DiagnosticStatus`
- JSON serialization of `Diagnostic.Status` MUST remain unchanged (values `"pass"`, `"fail"`, `"warn"`)

### Updated Constraint: DC-3

DC-3 updated to include `DiagnosticStatus` in the typed enum list: Zone, DocStatus, BriefGate, RequestType, IterationStatus, CheckLevel, EdgeType, LockStatus, EntryStatus, DiagnosticStatus.

### No Other Domain Changes

No new entities, no new business rules, no changes to existing business rules (BR-1 through BR-38). The changes are confined to type safety improvements within existing entities.

---

## Functional Requirements

### Theme 1: Interface/Type Consistency (MUST -- Phase 3 Blocking)

- **FR-125**: The `deps.Deps` struct (`internal/deps/deps.go`) MUST declare all repository fields using `repo.` interface types (`repo.DocRepo`, `repo.IterationRepo`, `repo.BriefRepo`, `repo.ConfigRepo`, `repo.LockRepo`, `repo.StateRepo`, `repo.QualityRepo`) instead of concrete `*fs.` types. The `Build()` function MUST still construct `fs.` implementations but return them through interface fields. Service fields MAY remain as concrete service types since services are not interfaces in the current architecture. [MUST]

  **Acceptance Criteria**:
  - GIVEN `internal/deps/deps.go` WHEN inspected THEN all 7 repository fields use `repo.` interface types, not `*fs.` concrete types.
  - GIVEN `tui/app.go` WHEN its import list is inspected THEN it does NOT import `internal/repo/fs`.
  - GIVEN the full codebase WHEN `go build ./...` runs THEN compilation succeeds with zero errors.
  - GIVEN all 374+ existing tests WHEN `go test ./...` runs THEN all pass.

### Theme 2: Inverse Dependency Removal (SHOULD)

- **FR-126**: The `internal/repo/mem/` package MUST NOT import `internal/repo/fs`. Any shared functionality (such as `IsStubContent()`) MUST be relocated to a package that both `fs/` and `mem/` can import without creating circular or inverse dependencies. Candidates: `domain/` (if the function is pure with no I/O imports), `internal/repo/` (shared utility at the interface level), or a new `internal/repo/shared/` package. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `internal/repo/mem/` WHEN its import list is inspected via `go list -f '{{.Imports}}' ./internal/repo/mem/` THEN `internal/repo/fs` does NOT appear.
  - GIVEN `internal/repo/fs/` and `internal/repo/mem/` WHEN both use the shared stub detection function THEN they produce identical results for identical inputs.
  - GIVEN all existing tests WHEN `go test ./...` runs THEN all pass.

### Theme 3: Staleness Propagation Accuracy (SHOULD)

- **FR-127**: The `buildReason()` function in `internal/reconcile/propagate.go` MUST produce edge-type-specific reason strings at ALL propagation depths, not only at depth 0. At depth > 0, the reason string MUST reflect the edge type between the immediate predecessor and the target document (the edge that caused this node to be enqueued), not the edge type from the original source. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN a chain A --(requires)--> B --(informs)--> C where A changes WHEN propagation runs THEN B's reason contains "prerequisite changed" AND C's reason contains "may be outdated" (reflecting the B-to-C `informs` edge type, not a generic fallback).
  - GIVEN a chain A --(validates)--> B --(requires)--> C where A changes WHEN propagation runs THEN B's reason contains "needs re-validation" AND C's reason contains "prerequisite changed".
  - GIVEN depth > 0 propagation WHEN reason strings are inspected THEN none contain the generic "may be outdated" fallback from unresolved edge types; all contain the specific edge type reason from the immediate predecessor edge.

### Theme 4: CLI Flag Rename (SHOULD)

- **FR-128**: The `--project` global flag (`cmd/root.go`) MUST be renamed to `--project-root`. The short flag `-p` MAY be retained. All internal references to the flag value (`flagProject` variable, help text, error messages) MUST be updated to reflect the new name. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `mind status --project-root /tmp/myproject` WHEN invoked THEN the CLI uses `/tmp/myproject` as the project root.
  - GIVEN `mind status --project /tmp/myproject` WHEN invoked THEN the CLI rejects the flag as unknown (or, if backward compatibility is chosen, accepts it with a deprecation warning).
  - GIVEN `mind --help` WHEN inspected THEN the global flags section shows `--project-root` (not `--project`).
  - GIVEN `cmd/root.go` WHEN inspected THEN the flag registration uses `"project-root"` as the flag name.

### Theme 5: Exit Code Architecture (SHOULD)

- **FR-129**: All `cmd/` command handlers MUST return errors to Cobra instead of calling `os.Exit()` directly. A structured error type (e.g., `ExitError` carrying an exit code) MUST be used so that the root command's `Execute()` function (or a `PersistentPostRunE` / `SilenceErrors` handler) can map errors to the correct exit codes. Dead code (`return nil` after `os.Exit()`) MUST be removed. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `cmd/` package source files WHEN searched for `os.Exit(` THEN zero matches are found (excluding test files).
  - GIVEN a command that previously called `os.Exit(1)` on validation failure WHEN the same error condition occurs THEN the command returns an error and the process exits with code 1 (same behavior, different mechanism).
  - GIVEN all existing tests WHEN `go test ./...` runs THEN all pass.
  - GIVEN any command handler WHEN it encounters an error THEN it returns an error value (not nil), and any deferred cleanup functions execute normally.

- **FR-130**: The `Diagnostic.Status` field in `domain/health.go` MUST use a typed `DiagnosticStatus` enum instead of raw `string`. The enum MUST define constants for `"pass"`, `"fail"`, and `"warn"`. JSON serialization MUST produce the same string values as the current implementation. The `DoctorService.addDiag()` method and `DoctorSummary` counter logic MUST use the typed enum instead of string comparison. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `domain/health.go` WHEN inspected THEN `Diagnostic.Status` is typed `DiagnosticStatus`, not `string`.
  - GIVEN `domain/` package WHEN inspected THEN `DiagnosticStatus` is defined as a typed `string` constant with values `DiagPass`, `DiagFail`, `DiagWarn` (or equivalent naming following existing enum conventions).
  - GIVEN `internal/service/doctor.go` WHEN `addDiag()` is inspected THEN it accepts `DiagnosticStatus` (not raw `string`) for the status parameter.
  - GIVEN `mind doctor --json` WHEN output is inspected THEN the `status` field values are `"pass"`, `"fail"`, `"warn"` (unchanged JSON contract).
  - GIVEN DC-3 constraint WHEN reviewed THEN `DiagnosticStatus` is listed among the typed string constant enums.

### Theme 6: Test Coverage (SHOULD)

- **FR-131**: The `cmd/` package MUST have test coverage for exit code behavior. Tests MUST verify that commands return the correct exit codes for success (0), validation failure (1), runtime error (2), and configuration error (3) scenarios. Tests SHOULD use `cobra.Command.Execute()` with injected arguments rather than process forking. At minimum, the following commands MUST have exit code tests: `check docs`, `check refs`, `check config`, `check all`, `reconcile`, `status`, and `doctor`. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `cmd/` package WHEN `go test ./cmd/` runs THEN at least one test file exists and tests pass.
  - GIVEN a test for `mind check docs` with all checks passing WHEN executed THEN the command returns no error (exit 0).
  - GIVEN a test for `mind check docs` with a FAIL-level check WHEN executed THEN the command returns an error carrying exit code 1.
  - GIVEN a test for a command invoked outside a Mind project WHEN executed THEN the command returns an error carrying exit code 3.
  - GIVEN a test for `mind reconcile --check` with stale documents WHEN executed THEN the command returns an error carrying exit code 4.

- **FR-132**: The `internal/render/` package MUST have test coverage for JSON output mode. Tests MUST verify that `RenderHealth()`, `RenderValidation()`, `RenderReconcileResult()`, and `RenderDoctorReport()` produce valid JSON that matches the expected field names and structure defined by domain type JSON struct tags. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `internal/render/` package WHEN `go test ./internal/render/` runs THEN at least one test file exists and tests pass.
  - GIVEN a `ProjectHealth` value WHEN `RenderHealth()` is called in JSON mode THEN the output is valid JSON containing `"project"`, `"brief"`, `"zones"`, `"warnings"`, and `"suggestions"` fields.
  - GIVEN a `ValidationReport` value WHEN `RenderValidation()` is called in JSON mode THEN the output is valid JSON containing `"suite"`, `"checks"`, `"total"`, `"passed"`, `"failed"` fields.
  - GIVEN a `DoctorReport` value WHEN `RenderDoctorReport()` is called in JSON mode THEN the output is valid JSON containing `"diagnostics"` and `"summary"` fields.

### Theme 7: Documentation Accuracy (SHOULD)

- **FR-133**: `docs/spec/architecture.md` MUST be updated to fix the following stale references: (a) `cmd/tui_cmd.go` (line 431) MUST be corrected to `cmd/tui.go`; (b) the Phase 1 component map (lines 50-68) MUST include `InitService`, `DoctorService`, and `ReconciliationService`; (c) the `--project` flag reference MUST be updated to `--project-root` (after FR-128 is implemented). [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `docs/spec/architecture.md` WHEN searched for `cmd/tui_cmd.go` THEN zero matches are found.
  - GIVEN the Phase 2 Packages table in `docs/spec/architecture.md` WHEN inspected THEN the TUI command row references `cmd/tui.go`.
  - GIVEN the Phase 1 Packages table in `docs/spec/architecture.md` WHEN inspected THEN rows exist for `InitService` and `DoctorService` (at minimum).
  - GIVEN `docs/spec/architecture.md` WHEN searched for `--project` as a flag name THEN all instances read `--project-root`.

- **FR-134**: `docs/spec/requirements.md` overview section MUST be updated to reflect multi-phase scope. The current text states "This document covers Phase 1 (Core CLI) only" -- it MUST acknowledge Phase 1.5, Phase 2, and Phase 2.5 (cleanup) requirements that have been appended. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `docs/spec/requirements.md` WHEN the overview section is read THEN it acknowledges that the document covers Phase 1, Phase 1.5, Phase 2, and the pre-Phase 3 cleanup iteration.
  - GIVEN `docs/spec/requirements.md` WHEN the scope boundary section is read THEN it reflects the current scope including reconciliation, TUI, and cleanup.

- **FR-135**: `docs/spec/domain-model.md` MUST be updated to include `DiagnosticStatus` in the supporting types table and in constraint DC-3. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `docs/spec/domain-model.md` supporting types table WHEN inspected THEN a row exists for `DiagnosticStatus` with kind "Enum (string)" and values `pass`, `fail`, `warn`.
  - GIVEN DC-3 in `docs/spec/domain-model.md` WHEN inspected THEN `DiagnosticStatus` is listed among the typed string constant enums.

- **FR-136**: `docs/state/current.md` MUST be updated to reflect the resolution of SHOULD/COULD items addressed by this iteration. Items fixed by FR-125 through FR-135 MUST be removed from the Known Issues list. The Active Work and Recent Changes sections MUST be updated. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `docs/state/current.md` Known Issues WHEN inspected THEN items resolved by this iteration (transitive propagation, `--project` flag, `Diagnostic.Status`) are no longer listed.
  - GIVEN `docs/state/current.md` Recent Changes WHEN inspected THEN an entry for iteration 004 exists.

### Theme 8: Deps Build Signature Alignment (SHOULD)

- **FR-137**: After FR-125 migrates Deps to interface types, the `tui/app.go` file MUST access repositories and services exclusively through the `Deps` struct interface fields. Any direct imports of `internal/repo/fs` from `tui/` MUST be removed. The `fs.DetectProject()` call (if present in `tui/app.go`) MUST be moved to the command handler (`cmd/tui.go`) or into a service method. [SHOULD]

  **Acceptance Criteria**:
  - GIVEN `tui/` package WHEN its import list is inspected via `go list -f '{{.Imports}}' ./tui/` THEN `internal/repo/fs` does NOT appear.
  - GIVEN `tui/app.go` WHEN inspected THEN it accesses repos only through `deps.Deps` fields, never through direct `fs.` constructor calls.
  - GIVEN the TUI launches successfully WHEN `mind tui` is run against a valid project THEN behavior is identical to before the refactoring.

### Theme 9: Comprehensive Verification (MUST)

- **FR-138**: After all changes in FR-125 through FR-137 are applied, `go vet ./...` MUST report zero issues and `go build ./...` MUST succeed. [MUST]

  **Acceptance Criteria**:
  - GIVEN the completed codebase WHEN `go vet ./...` runs THEN zero issues are reported.
  - GIVEN the completed codebase WHEN `go build ./...` runs THEN compilation succeeds.
  - GIVEN the completed codebase WHEN `go test ./...` runs THEN all tests pass (including new tests from FR-131 and FR-132).

- **FR-139**: All pre-existing 374 tests MUST continue to pass after all changes. No existing test MAY be deleted. Existing tests MAY be modified only to accommodate interface changes (e.g., Deps field type changes) or the `--project` to `--project-root` rename. [MUST]

  **Acceptance Criteria**:
  - GIVEN `go test ./...` WHEN run THEN the total test count is >= 374 (existing) + new tests from FR-131 and FR-132.
  - GIVEN any existing test WHEN inspected THEN it has not been deleted.
  - GIVEN any existing test that was modified WHEN inspected THEN the modification is limited to type signature updates or flag name updates, not logic changes.

---

## Priority Summary

| Priority | FR IDs | Count | Description |
|----------|--------|-------|-------------|
| MUST | FR-125, FR-138, FR-139 | 3 | Deps interface migration, build verification, test regression guard |
| SHOULD | FR-126, FR-127, FR-128, FR-129, FR-130, FR-131, FR-132, FR-133, FR-134, FR-135, FR-136, FR-137 | 12 | Inverse deps, propagation fix, flag rename, exit codes, enum types, test coverage, doc updates |
| COULD | (none in this iteration) | 0 | DRY refactoring, DoctorService delegation, TUI component extraction deferred |

---

## Convergence Finding Traceability

| FR | Convergence Finding | Weighted Score | Theme |
|----|-------------------|----------------|-------|
| FR-125 | Deps struct uses concrete `*fs.` types instead of `repo.` interfaces | 79 | Interface/Type Consistency |
| FR-126 | `internal/repo/mem/` imports `internal/repo/fs` -- inverse dependency | 57 | Inverse Dependency Removal |
| FR-127 | Transitive propagation loses edge-type-specific reason strings at depth > 0 | 69 | Staleness Propagation Accuracy |
| FR-128 | `--project` flag should be `--project-root` per api-contracts spec | 62 | CLI Flag Rename |
| FR-129 | os.Exit() calls to error returns -- Cobra best practice / cmd/ exit code test coverage | 60 + 44 | Exit Code Architecture |
| FR-130 | `Diagnostic.Status` raw strings to typed enum | 51 | Exit Code Architecture |
| FR-131 | cmd/ exit code test coverage -- 29 os.Exit() calls untestable | 60 | Test Coverage |
| FR-132 | render/ JSON output test coverage -- API contract validation | 60 | Test Coverage |
| FR-133 | Architecture doc stale references | 35 | Documentation Accuracy |
| FR-134 | Architecture doc stale references (scope description) | 35 | Documentation Accuracy |
| FR-135 | Diagnostic.Status domain model update (DC-3) | 51 | Documentation Accuracy |
| FR-136 | Current state tracking accuracy | 35 | Documentation Accuracy |
| FR-137 | Deps concrete types -- tui/ import coupling | 79 | Interface/Type Consistency |
| FR-138 | Build verification (all themes) | -- | Comprehensive Verification |
| FR-139 | Test regression guard (all themes) | -- | Comprehensive Verification |
