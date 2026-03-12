# Retrospective

## What Went Well

1. **Convergence-driven scope was effective**: The pre-phase-3-cleanup convergence analysis (score 4.3/5.0) identified 10 concrete recommendations with weighted severity scores. This made scope definition mechanical rather than subjective. All 9 actionable recommendations were addressed, and the 1 deferred item (R10: DRY utilities) was explicitly scoped out with clear justification. The requirements-delta document traced each FR back to its convergence finding, making review traceability straightforward.

2. **Migration order minimized risk**: The architect's 9-step migration path (DiagnosticStatus first, then IsStubContent, then propagation fix, then Deps migration, then flag rename, then ExitError, then tests, then docs) kept the codebase buildable and test-passing at each step. The highest-risk change (FR-125 Deps interface migration) was positioned after simpler refactorings reduced the diff surface. All 8 implementation commits are atomic and individually revertible.

3. **ExitError pattern unlocked testability**: Replacing 29 `os.Exit()` calls with `ExitError` returns was the single most impactful change for codebase quality. It enabled 9 integration tests in `cmd/cmd_test.go` that validate the full command pipeline (PersistentPreRunE wiring, service calls, exit codes) without process forking. The `setupProject`/`executeWithRoot` test helpers establish a reusable pattern for Phase 3 command testing.

## What Could Improve

1. **cmd/ test assertions could be stronger**: Several exit-code tests (e.g., `TestCheckDocsExitCodePass`, `TestStatusExitCode`, `TestDoctorExitCode`) accept both exit 0 and exit 1 as valid outcomes because the minimal test project does not guarantee all checks pass. This pragmatic approach reduces test setup complexity but weakens regression detection. Future iterations should invest in fixture projects that produce deterministic pass/fail outcomes for specific check suites.

2. **DetectProject still couples service to fs**: `ProjectService.DetectProject()` delegates directly to `fs.DetectProject()`, meaning the service layer retains a hard import on the filesystem implementation. While the architecture permits this (services may import infrastructure), it means `ProjectService` cannot be tested with in-memory implementations for the detection path. A `ProjectDetector` interface would complete the decoupling. This was noted as a SHOULD finding in the validation report.

3. **Test count discrepancy in iteration documents**: The overview states a baseline of 374 tests, the changes.md states 389 tests baseline, and the test-summary states "Baseline 374" but "Final 408" with "19 + 15 = 34 new tests." The arithmetic from 374 + 34 = 408 is consistent, but changes.md referencing 389 creates confusion. The 389 figure likely represents the state after the developer's step 7 (before the tester's additions). Future iterations should maintain a single source of truth for test counts.

## Open Items

The following items from the convergence COULD list and known issues remain deferred:

- **DoctorService delegation to ValidationService**: Doctor reimplements 9 checks that overlap with validation suites. COULD priority, high churn for low Phase 3 impact.
- **GenerateService repository injection**: Direct filesystem I/O for file creation. Convergence consensus: acceptable for scaffolding operations.
- **TUI component extraction from tab views**: 9 component files inlined into tab views (S-4). Cosmetic, no Phase 3 impact.
- **DRY refactoring of artifact counting, filter bar, staleness map**: COULD priority, do opportunistically when touching files (R10).
- **TUI preview pane Glamour rendering**: S-2, cosmetic improvement.
- **TUI status bar cursor position**: S-3, cosmetic improvement.
- **TUI editor fallback behavior**: S-1, convergence consensus: current behavior is acceptable.
- **Graph rendering rooted tree**: Currently flat adjacency list (C-2 variant).
- **Quality tab fixed Y-axis scale**: C-2, cosmetic.
- **5 exported methods in fs/doc_repo.go lack GoDoc comments**: NFR-8.
- **FR-88 (--check/--force exclusion) tested by code inspection only**: S-5, no unit test.
