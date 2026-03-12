# Retrospective

- **Iteration**: 001-new-project-core-cli
- **Date**: 2026-03-11

## What Went Well

- Domain purity at 100% coverage with zero external imports establishes a strong foundation for all future phases.
- The validation engine design (Suite/Check/CheckFunc) is clean and composable -- adding checks is trivial.
- Test suite of 395 tests with strong traceability to business rules (BR-1 through BR-23) and acceptance criteria.
- In-memory repository implementations enable fast, deterministic testing without filesystem dependencies.
- All 20+ commands implemented with consistent --json support and exit code handling.

## What Could Improve

- Flag naming (`--project` vs `--project-root`) should be validated against api-contracts early in the developer phase to prevent spec drift.
- The `docs search` command bypasses the DocRepo abstraction. A `SearchRepo` or `DocRepo.Search()` method would maintain C-9 compliance.
- GoDoc gaps on concrete repo implementations should be caught by a lint step (e.g., `golangci-lint` with `revive` or `godot`).
- DoctorService would benefit from delegating to the existing validation suites rather than reimplementing checks, reducing divergence risk.
- Centralizing repo wiring in main.go (as the architecture doc recommends) would reduce duplication across command handlers.

## Discovered Patterns

- The thin-handler pattern (resolve root -> create repos -> call service -> render) is consistent and easy to review.
- JSON output contracts owned by domain struct tags is effective -- render layer just calls `json.MarshalIndent`.
- Strict mode as a flag on the check context (promoting WARN to FAIL) is a clean design that avoids touching check logic.

## Open Items

- **SHOULD-1**: Rename `--project` flag to `--project-root` per api-contracts spec. Consider whether walk-up behavior should be disabled when flag is explicit.
- **SHOULD-2**: Move search filesystem logic into DocRepo or a dedicated SearchRepo.
- **SHOULD-3/4**: Add GoDoc to 5 methods in `fs/doc_repo.go` and 2 Error() methods in `domain/errors.go`.
- **COULD-3**: Consider having DoctorService delegate to ValidationService for check execution.
- **COULD-4**: Use `time.Time` for `IterationSummary.CreatedAt` to get RFC 3339 in JSON output.
- **Tech debt**: Centralize repo wiring in main.go before Phase 2 adds TUI complexity.
