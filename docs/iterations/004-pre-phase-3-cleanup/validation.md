# Validation Report

## Summary
- **Type**: COMPLEX_NEW
- **Status**: APPROVED
- **Implements**: FR-125 through FR-139
- **MUST findings**: 0
- **SHOULD findings**: 1
- **COULD findings**: 2

## Deterministic Gate Results
- Build: PASS (`go build ./...`)
- Vet: PASS (`go vet ./...` -- zero issues)
- Tests: PASS (408 passed, 0 failed)

## MUST Findings

No blocking issues found.

## SHOULD Findings

1. **[internal/service/project.go:36] DetectProject delegates to fs.DetectProject without interface abstraction**
   - **Observation**: `ProjectService.DetectProject()` directly calls `fs.DetectProject(root)`, importing `internal/repo/fs` from the service layer. While the architecture allows service-to-infrastructure imports, this means the service layer retains a hard dependency on the filesystem implementation for project detection. The convergence recommendation (R1) was to decouple presentation layers from `fs/`; this is achieved for `cmd/` and `tui/`, but the service layer still has the coupling.
   - **Recommendation**: For Phase 3, consider extracting a `ProjectDetector` interface in `internal/repo/` that both `fs/` and `mem/` implement. This would make `ProjectService` fully testable with in-memory implementations. Low urgency since the current approach works correctly and the architecture permits service-to-infrastructure imports.

## COULD Findings

1. **[cmd/cmd_test.go:303-319] TestCheckDocsExitCodePass allows both exit 0 and exit 1**
   - **Observation**: The test accepts both outcomes ("may fail due to missing docs"). While pragmatic for test setup complexity, this weakens the assertion. A test that always passes regardless of outcome provides limited regression value.
   - **Suggestion**: Create a project fixture that guarantees all 17 doc checks pass, then assert exit 0. Alternatively, split into two tests: one for a fully valid project (exit 0) and one with a known failure (exit 1).

2. **[cmd/root.go:96-108] Execute() retains two os.Exit call sites**
   - **Observation**: `Execute()` has `os.Exit(exitErr.Code)` on line 102 and `os.Exit(1)` on line 105. This is correct per the architecture (single exit point), but it means the `Execute()` function itself is not unit-testable in isolation since `os.Exit` terminates the process.
   - **Suggestion**: For Phase 3 testing, consider extracting exit code logic into a testable function that returns an int, with `Execute()` as the thin wrapper that calls `os.Exit`. Low priority since cmd/ tests work around this by calling `rootCmd.Execute()` directly.

## Requirement Traceability

| Requirement | Implementation | Test | Status |
|-------------|---------------|------|--------|
| FR-125 | `internal/deps/deps.go`: 7 repo fields changed from `*fs.X` to `repo.X` | `deps_test.go:TestBuild_RepoFieldsSatisfyInterfaces` + compilation | PASS |
| FR-126 | `internal/repo/stub.go`: `IsStubContent()` relocated; `mem/doc_repo.go` imports `repo` not `fs` | `stub_test.go:TestIsStubContent` (16 cases); `go list` confirms no `fs` import in `mem/` | PASS |
| FR-127 | `internal/reconcile/propagate.go`: `edgeType` field on `queueItem`; `buildReason` uses carried edge type | 4 FR-127 tests in `propagate_test.go` covering mixed edge types at all depths | PASS |
| FR-128 | `cmd/root.go:91`: flag name changed to `"project-root"` | `cmd_test.go:TestProjectRootFlagRegistered`, `TestOldProjectFlagRemoved` | PASS |
| FR-129 | `cmd/errors.go`: `ExitError` type; all `os.Exit` in handlers replaced with `exitX()` returns; only `Execute()` calls `os.Exit` | 6 ExitError tests + 9 exit-code integration tests in `cmd_test.go` | PASS |
| FR-130 | `domain/health.go`: `DiagnosticStatus` type with `DiagPass`, `DiagFail`, `DiagWarn`; `Diagnostic.Status` field typed | 6 tests in `health_test.go` covering values, JSON serialization, exhaustiveness | PASS |
| FR-131 | `cmd/cmd_test.go`: 9 exit-code tests covering check-docs, check-refs, check-config, check-all, status, reconcile, doctor, not-a-project | Tests exist and pass | PASS |
| FR-132 | `internal/render/render_test.go`: 6 JSON tests covering `RenderHealth`, `RenderValidation`, `RenderReconcileResult`, `RenderDoctor`, zone coverage, status values | Tests exist and pass | PASS |
| FR-133 | `docs/spec/architecture.md`: `cmd/tui_cmd.go` replaced with `cmd/tui.go`; `--project` replaced with `--project-root`; `InitService`, `DoctorService`, `ReconciliationService` added to component table | Verified by inspection: zero matches for `cmd/tui_cmd.go`, zero matches for `--project[^-]` | PASS |
| FR-134 | `docs/spec/requirements.md:7`: overview acknowledges Phase 1, 1.5, 2, and pre-Phase 3; scope boundary updated with "Delivered in later phases" subsection | Verified by inspection at lines 1-30 | PASS |
| FR-135 | `docs/spec/domain-model.md:44`: `DiagnosticStatus` row added to Supporting Types; DC-3 updated at lines 246 and 388 | Verified by inspection | PASS |
| FR-136 | `docs/state/current.md`: resolved issues removed; iteration 004 entry added; Active Work and Next Priorities updated | Verified by inspection | PASS |
| FR-137 | `tui/app.go`: `fs.DetectProject` calls replaced with `deps.ProjectSvc.DetectProject`; `internal/repo/fs` import removed | `go list` confirms zero `fs` imports from `tui/` | PASS |
| FR-138 | `go vet ./...` PASS; `go build ./...` PASS; `go test ./...` PASS (408 tests) | Deterministic gate results | PASS |
| FR-139 | 408 tests >= 374 baseline; no existing tests deleted; modified tests limited to interface/flag updates | Test count verified; changes.md documents test-only modifications | PASS |

## Domain Model Compliance

- **DiagnosticStatus enum**: Defined in `domain/health.go:25-32` as `type DiagnosticStatus string` with constants `DiagPass = "pass"`, `DiagFail = "fail"`, `DiagWarn = "warn"`. Follows existing enum pattern (e.g., `IterationStatus`, `LockStatus`).
- **DC-3 constraint**: `docs/spec/domain-model.md` DC-3 updated at both line 246 (Phase 1 section) and line 388 (Phase 1.5 section) to include `DiagnosticStatus` in the typed string constant enum list.
- **Diagnostic entity**: `docs/spec/domain-model.md:23` Diagnostic row updated to show `Status (DiagnosticStatus)` instead of `Status (string)`.
- **JSON contract preserved**: `DiagnosticStatus` is a typed `string`, so `json.Marshal` produces `"pass"`, `"fail"`, `"warn"` unchanged. Verified by `TestDiagnosticStatus_JSONSerialization` and `TestDiagnostic_JSONStatusField`.

## Convergence Alignment

| Recommendation | Status | Evidence |
|---------------|--------|----------|
| R1: Migrate Deps to interface types | IMPLEMENTED | FR-125 + FR-137. `deps.go` uses `repo.X` interfaces; `tui/app.go` no longer imports `fs`. |
| R2: Fix transitive propagation reasons | IMPLEMENTED | FR-127. Edge type carried on queue item; 4 dedicated tests verify mixed-edge chains at all depths. |
| R3: Rename --project to --project-root | IMPLEMENTED | FR-128. Hard rename (no deprecation alias). 2 flag tests confirm. |
| R4: Add cmd/ exit code tests | IMPLEMENTED | FR-131. 9 integration tests in `cmd_test.go`. |
| R5: Add render/ JSON tests | IMPLEMENTED | FR-132. 6 JSON output tests in `render_test.go`. |
| R6: Fix mem/ importing fs/ | IMPLEMENTED | FR-126. `IsStubContent` moved to `internal/repo/stub.go`. `go list` confirms zero `fs` imports in `mem/`. |
| R7: Type Diagnostic.Status | IMPLEMENTED | FR-130. New `DiagnosticStatus` enum. 6 dedicated tests. |
| R8: Replace os.Exit with error returns | IMPLEMENTED | FR-129. `ExitError` type in `cmd/errors.go`. Only `Execute()` calls `os.Exit`. 15 tests (6 unit + 9 integration). |
| R9: Update architecture docs | IMPLEMENTED | FR-133 through FR-136. All 4 spec documents updated. |
| R10: Extract DRY utilities | DEFERRED | Out of scope per requirements-delta. COULD priority. Artifact counting, filter bar, staleness map patterns remain in tab views. |

All 9 actionable convergence recommendations (R1-R9) are implemented. R10 was explicitly deferred as COULD priority per the scope boundary defined in requirements-delta.md.

## Git Discipline
- **Commits**: 12 (2 planning/WIP + 8 implementation + 2 testing/docs)
- **Atomic**: Yes -- each commit addresses a specific FR or theme and leaves the codebase buildable
- **Convention**: Yes -- prefixes follow established pattern (`refactor:`, `fix:`, `test:`, `docs:`, `wip:`)
- **FR references**: All commits include FR ID references in the message (e.g., `(FR-130)`, `(FR-125, FR-137)`)
- **Diff size**: 3,118 insertions, 220 deletions across 37 files -- proportional to a 15-FR cleanup iteration

## Quality Review

### 1. Names and Types Express Intent
- `DiagnosticStatus` with `DiagPass`/`DiagFail`/`DiagWarn` follows the existing naming convention (`IterationStatus`/`IterInProgress`). Clear.
- `ExitError` with `Code`, `Err`, `Quiet` fields -- self-documenting. Convenience constructors (`exitValidation`, `exitRuntime`, `exitConfig`, `exitStaleness`) map directly to exit code semantics.
- `IsStubContent` at `internal/repo/stub.go` -- well-named, clear package placement.
- No vague names, no misleading types.

### 2. Well-Structured
- `ExitError` is 30 lines in its own file -- focused and minimal.
- `propagate.go` `buildReason` simplified from a graph-lookup function to a 5-line direct formatting function.
- `ProjectService.DetectProject` is a thin delegation method (1 line body) -- appropriate for its purpose of decoupling presentation from infrastructure.
- No functions exceeding 50 lines added. No deep nesting.

### 3. Idiomatic Go
- `ExitError` implements `error` via `Error() string` and `Unwrap() error` -- standard Go error wrapping pattern.
- `errors.As` used correctly in `Execute()` for exit code extraction.
- Typed string constants for `DiagnosticStatus` follow the Go enum idiom used throughout the codebase.
- No non-idiomatic patterns observed.

### 4. DRY and Consistent
- `ExitError` convenience constructors avoid repeating `&ExitError{Code: N, Err: err}` patterns across 10+ files.
- `IsStubContent` shared between `fs/` and `mem/` -- eliminates the duplicated dependency.
- All cmd handlers follow the same post-migration pattern: render output, then return `exitQuiet(N)` or `nil`.

### 5. Documented and Tested
- All new exported types have GoDoc comments: `ExitError`, `DiagnosticStatus`, `DiagPass`/`DiagFail`/`DiagWarn`, `IsStubContent`.
- 34 new tests added (19 by tester + 15 by developer). All critical paths covered.
- `cmd/cmd_test.go` establishes a `setupProject`/`executeWithRoot` testing pattern reusable for Phase 3 command tests.

### 6. Cross-File Patterns Clean
- The `exitQuiet(N)` pattern is used consistently across `check.go`, `status.go`, `doctor.go`, and `docs.go` for commands that render output before exiting.
- The `exitConfig`/`exitRuntime` pattern is used consistently in `reconcile.go`, `root.go`, and `tui.go` for error classification.
- No dead exports introduced. No implicit contracts.

### Clean
- Domain layer purity maintained: `domain/` imports only Go stdlib (DC-1 verified).
- Exit code semantics unchanged: 0 (success), 1 (validation), 2 (runtime), 3 (config), 4 (staleness).
- JSON output structure unchanged (verified by 6 render tests).
- TUI behavior unchanged (confirmed by `tui/app.go` using `deps.ProjectSvc.DetectProject` transparently).

## Sign-off

**APPROVED**. All 15 functional requirements (FR-125 through FR-139) are implemented and verified. The 3 MUST requirements (FR-125, FR-138, FR-139) pass with strong evidence. All 12 SHOULD requirements are implemented. Zero MUST findings, 1 SHOULD finding (service-layer fs import retained in DetectProject delegation -- acceptable per architecture rules), 2 COULD suggestions for future improvement. The codebase is cleaner, more testable, and well-positioned for Phase 3. All 9 actionable convergence recommendations are addressed. Git discipline is strong with atomic, well-labeled commits.
