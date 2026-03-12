# Retrospective

**Iteration**: 005-COMPLEX_NEW-phase-3-review-and-remediation
**Date**: 2026-03-12

---

## What Went Well

- **All three MUST fixes were cleanly isolated**: M-1 (protocol), M-2 (domain constants), and M-3 (test coverage) were each addressed in discrete commits with FR references, making the remediation traceable and reviewable.
- **HandoffService extraction was complete**: The refactor from inline `cmd/handoff.go` logic to `HandoffService` with `IterationRepo` injection followed the architecture-delta spec precisely — constructor signature, `HandoffResult` field set, and `StateRepo` delegation all match the architect's design.
- **Layer violation fully resolved**: `cmd/handoff.go` no longer imports `os`; the `AppendCurrentState` logic now lives at the correct layer (`internal/repo/fs/`) with a clean interface contract and both `mem/` and `fs/` implementations tested.
- **Coverage targets met or exceeded**: `internal/mcp` at 80.3%, `internal/orchestrate` at 81.2%, `quality.go` at 85–100% per function. The test files also cover HandoffService directly, verifying FR-146 behavior.
- **COULD items were all addressed**: FR-149 (`classify.go` adapter), FR-150 (`strings` stdlib replacement), and FR-151 (`--json` flag) were completed as a single `chore:` commit without noise.

## What Could Improve

- **`TestPreflightService_Run_DocWarningsNonBlocking` assertion is incomplete**: The test proves warnings are non-blocking but does not assert that `DocWarnings > 0` when the validation stub returns warnings. A stronger test would inject a stub that returns `Warnings: 1` and verify `result.DocWarnings == 1`.
- **FR-151 Renderer routing is partial**: `cmd/preflight.go` handles `--json` inline rather than through `Renderer`. The developer flagged this deferred item correctly — full routing would require adding a method to `internal/render/render.go` that imports `internal/orchestrate`, which is structurally sound but was deprioritized.
- **Branch prefix `complex/` vs `complex-new/`**: The orchestrator created the branch manually as `complex/phase-3-review-and-remediation` before the `buildBranchName()` function was available. Future COMPLEX_NEW iterations will use `complex-new/` per the current `buildBranchName()` logic.

## Discovered Patterns

- **Two-stage convergence input**: This iteration was the first to treat a convergence document as the primary input for both the analyst (requirements) and architect (structural decisions). The quality of requirements-delta.md and architecture-delta.md was noticeably higher than earlier iterations because the convergence had resolved ambiguities (e.g., Option A vs Option B for HandoffService) before any code was written.
- **Stub no-op vs record for mem implementations**: The mem `StateRepo.AppendCurrentState()` records entries to `CurrentStateEntries []string` rather than silently dropping them. This pattern (record for assertions, not just return nil) produces better testability and was correctly chosen here.

## Open Items

- `cmd/preflight.go` Renderer routing (FR-151 partial) — carry forward to Phase 4 polish.
- `TestPreflightService_Run_DocWarningsNonBlocking` positive assertion gap — low priority; code is correct.
- `cmd/handoff_test.go` integration test for the full command was noted in the convergence document but not required by FR-142 scope. Not blocking.
