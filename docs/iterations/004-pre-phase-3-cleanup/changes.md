# Pre-Phase 3 Cleanup -- Changes

## Commits

| Hash | Message | FRs |
|------|---------|-----|
| `9679482` | refactor: type DiagnosticStatus enum | FR-130 |
| `4c3553e` | refactor: move IsStubContent to internal/repo | FR-126 |
| `5f0dca5` | fix: transitive propagation edge-type reasons | FR-127 |
| `91b862a` | refactor: migrate Deps to interface types | FR-125, FR-137 |
| `6d150d8` | refactor: rename --project flag to --project-root | FR-128 |
| `5ea1e95` | refactor: replace os.Exit with ExitError returns | FR-129 |
| `0ae52fa` | test: add cmd/ exit code and render/ JSON tests | FR-131, FR-132 |
| `fb42dfb` | docs: update spec documents for pre-phase-3 cleanup | FR-133--FR-136 |

## Step 1: DiagnosticStatus Typed Enum (FR-130)

- `domain/health.go`: Added `DiagnosticStatus` string type and constants (`DiagPass`, `DiagFail`, `DiagWarn`). Changed `Diagnostic.Status` field from `string` to `DiagnosticStatus`.
- `internal/service/doctor.go`: Changed `addDiag` parameter from `status string` to `status domain.DiagnosticStatus`. Replaced ~30 raw string literals with typed constants. Updated summary counting and `applyFixes` to use typed constants.
- `internal/render/render.go`: Updated `renderDoctorText` switch cases and comparisons from raw strings to typed constants.

## Step 2: IsStubContent Relocation (FR-126)

- `internal/repo/stub.go` (new): `IsStubContent()` and helpers (`isBoilerplateLine`, `isTableSeparator`, `isPlaceholderRow`) moved from `fs/doc_repo.go`. Shared by both `fs/` and `mem/` packages, eliminating the `mem/` -> `fs/` inverse dependency.
- `internal/repo/stub_test.go` (new): 16 test cases moved from `fs/doc_repo_test.go` covering the stub detection algorithm.
- `internal/repo/fs/doc_repo.go`: `IsStub` method now delegates to `repo.IsStubContent(content)`. Local helpers removed.
- `internal/repo/mem/doc_repo.go`: Changed import from `internal/repo/fs` to `internal/repo`; calls `repo.IsStubContent(data)`.

## Step 3: Transitive Propagation Fix (FR-127)

- `internal/reconcile/propagate.go`: Fixed BFS propagation bug where edge type was reverse-looked-up at depth > 0 instead of carried on the queue item. Added `edgeType domain.EdgeType` field to `queueItem` struct. Changed `buildReason` signature to accept `edgeType` directly. Made `EdgeInforms` case explicit in `edgeTypeReason`.

## Step 4: Deps Interface Migration (FR-125, FR-137)

- `internal/deps/deps.go`: Changed all 7 repository fields from concrete `*fs.X` types to `repo.X` interface types.
- `internal/service/project.go`: Added `DetectProject(root string) (*domain.Project, error)` method delegating to `fs.DetectProject()`.
- `cmd/root.go`: Changed package-level vars `docRepo`, `iterRepo`, `briefRepo` from `*fs.X` to `repo.X`. Replaced `internal/repo/fs` import with `internal/repo`.
- `cmd/status.go`: Changed `fs.DetectProject(projectRoot)` to `projectSvc.DetectProject(projectRoot)`. Removed `internal/repo/fs` import.
- `tui/app.go`: Changed `fs.DetectProject` calls to `deps.ProjectSvc.DetectProject`. Removed `internal/repo/fs` import.

## Step 5: Flag Rename (FR-128)

- `cmd/root.go`: Renamed `--project` flag to `--project-root`. Variable name unchanged (`projectRoot`).

## Step 6: ExitError Migration (FR-129)

- `cmd/errors.go` (new): `ExitError` type with `Code`, `Err`, `Quiet` fields. Convenience constructors: `exitValidation`, `exitRuntime`, `exitConfig`, `exitStaleness`, `exitQuiet`.
- `cmd/root.go`: `Execute()` handles `ExitError` with `os.Exit(exitErr.Code)`. `PersistentPreRunE` returns `exitConfig(err)` instead of `os.Exit(3)`.
- `cmd/check.go`: All `os.Exit(1)` replaced with `return exitQuiet(1)`.
- `cmd/reconcile.go`: All `os.Exit` calls replaced with `exitRuntime`, `exitConfig`, `exitStaleness` returns.
- `cmd/doctor.go`: `os.Exit(1)` replaced with `return exitQuiet(1)`.
- `cmd/docs.go`: `os.Exit(1)` calls replaced with `exitValidation`/`exitQuiet` returns.
- `cmd/create.go`: `os.Exit` calls replaced with `exitValidation`/`exitRuntime` returns.
- `cmd/init.go`: `os.Exit(2)` calls replaced with `exitRuntime` returns.
- `cmd/tui.go`: `os.Exit(3)` replaced with `return exitConfig(err)`.
- `cmd/status.go`: `os.Exit(1)` replaced with `return exitQuiet(1)`.

## Step 7: Tests (FR-131, FR-132)

- `cmd/cmd_test.go` (new): 9 exit-code tests with `setupProject` and `executeWithRoot` helpers. Covers check-docs pass/fail, not-a-project, status, reconcile --check, doctor, check-refs, check-config, check-all.
- `internal/render/render_test.go` (new): 6 JSON output tests for `RenderHealth`, `RenderValidation`, `RenderReconcileResult`, `RenderDoctor`, plus `AllZones` coverage and `DiagnosticStatus` value tests.

## Step 8: Spec Document Updates (FR-133--FR-136)

- `docs/spec/architecture.md`: Fixed `cmd/tui_cmd.go` -> `cmd/tui.go`. Added InitService, DoctorService, ReconciliationService to Phase 1 Packages table.
- `docs/spec/requirements.md`: Updated overview to acknowledge all phases. Updated out-of-scope with "Delivered in later phases" subsection.
- `docs/spec/domain-model.md`: Added `DiagnosticStatus` to Supporting Types. Updated DC-3 with `DiagnosticStatus`. Updated Diagnostic entity attributes.
- `docs/state/current.md`: Removed resolved known issues. Added iteration 004 entry. Updated Active Work and Next Priorities.

## Summary

- **15 FRs implemented**: FR-125 through FR-139
- **4 new files**: `internal/repo/stub.go`, `cmd/errors.go`, `cmd/cmd_test.go`, `internal/render/render_test.go`
- **2 new test files**: `internal/repo/stub_test.go`, `cmd/cmd_test.go`
- **14 modified files**: across `domain/`, `internal/service/`, `internal/render/`, `internal/repo/fs/`, `internal/repo/mem/`, `internal/reconcile/`, `internal/deps/`, `cmd/`, `tui/`, `docs/`
- **389 tests passing** (15 new tests added)
- **8 commits**, each leaving the codebase buildable and test-passing
