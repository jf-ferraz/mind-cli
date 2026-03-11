# Retrospective

- **Iteration**: 002-reconciliation-engine
- **Date**: 2026-03-11
- **Reviewer**: Claude Opus 4.6

## What Went Well

- **12-step implementation sequence** was well-chosen. Each step produced a testable increment. Domain types first, then pure algorithms, then I/O, then wiring -- this kept merge risk low and allowed each commit to build cleanly.
- **Domain purity maintained**. `domain/reconcile.go` has only a `time` import. The `BuildGraph()` function sits correctly in the domain layer as a pure transformation, following the `Slugify()`/`Classify()` precedent.
- **Wiring centralization** (Step 9) was executed at the right time. Adding `PersistentPreRunE` during Phase 1.5 avoided duplicating `LockRepo` and `ReconciliationService` construction across 4+ command handlers. Existing handlers were simplified in the same commit.
- **Test coverage is strong**: 82 new tests, all passing. Key algorithms (propagation, cycle detection, hash computation) are at or near 100% function coverage. Edge cases (empty files, symlinks, diamond graphs, depth limit boundaries) are explicitly tested.
- **Performance margins are wide**: full reconciliation at ~1.1ms vs 200ms target (180x headroom), incremental at ~0.45ms vs 50ms target (111x headroom).
- **Engine/Service separation** keeps algorithmic logic testable without filesystem mocking. The engine receives parsed data; the service handles I/O boundaries.

## What Could Improve

- **ReconcileSuite missing "missing documents" check** (S-2). The architecture delta and requirements both specify three check categories, but only two were implemented. This is a minor gap in the `mind check all` integration.
- **`--check` and `--force` mutual exclusivity not enforced** (S-1). The API contract specifies this, but no validation exists. The combination produces undocumented (though harmless) behavior.
- **Transitive edge-type reason strings** (S-3). Transitive propagation falls back to a generic "may be outdated" reason instead of preserving the edge type of the immediate upstream connection. The requirements are ambiguous on whether this is required for transitive cases, so the severity is debatable.
- **Render layer untested**. The `internal/render/` package has zero test coverage. This is pre-existing technical debt, not introduced by Phase 1.5, but the new `RenderReconcileResult()` and `RenderGraph()` methods add to the untested surface area.
- **cmd/ package untested**. Exit code behavior (FR-82), flag interaction validation, and command-level integration are verified by code inspection only. No integration test harness exists.

## Discovered Patterns

- **Pre-computed result projection** for validation suites. `ReconcileSuite` takes a `ReconcileResult` rather than running reconciliation itself. This pattern avoids redundant computation when the result is needed by multiple consumers (command handler for exit codes, renderer for output, validation suite for check all). Other expensive operations could follow the same pattern.
- **Package-level variables in `cmd/`** for centralized wiring via `PersistentPreRunE`. This is pragmatic for Cobra-based CLIs but creates implicit coupling. As the service count grows, consider a wiring struct or context-based injection.
- **Deep copy via JSON round-trip** in `mem.LockRepo`. Simple and correct, though it has a performance cost proportional to lock file size. Acceptable for test-only code.

## Open Items

1. **S-1**: Add `--check`/`--force` mutual exclusivity guard in `cmd/reconcile.go`.
2. **S-2**: Add missing documents check to `ReconcileSuite` in `internal/validate/reconcile.go`.
3. **S-3**: Consider preserving edge-type-specific reasons in transitive propagation in `internal/reconcile/propagate.go`.
4. **Deferred**: FR-54 (graph ASCII tree) rendering tests.
5. **Deferred**: FR-82 (exit code 4) integration tests.
6. **Deferred**: cmd/ package integration test harness.
