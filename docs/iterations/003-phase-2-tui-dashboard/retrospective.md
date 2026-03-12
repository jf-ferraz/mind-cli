# Retrospective: Phase 2 TUI Dashboard

- **Iteration**: 003-phase-2-tui-dashboard
- **Date**: 2026-03-11
- **Verdict**: APPROVED_WITH_NOTES
- **Tests**: 374 passing (128 new), 0 failures

---

## What Went Well

1. **Clean layer separation**: The TUI package (`tui/`) is a pure presentation layer that consumes existing services without modifying business logic. Zero changes to the service layer were needed for the 5-tab dashboard.

2. **BuildDeps centralization (FR-90)**: Extracting `internal/deps/deps.go` resolved the import cycle cleanly and provides a single wiring point for both CLI and TUI paths. This was a prerequisite that paid off immediately.

3. **DocRepo search abstraction (FR-91)**: Moving search into the repository interface followed the established pattern and enabled testability. The in-memory implementation allows search testing without filesystem access.

4. **Component testing strategy**: Pure view functions in `tui/components/` achieved 96.3% coverage. Message-based tab model testing (construct, send message, verify state) proved effective for Update() logic at 62.8% coverage.

5. **Domain model purity preserved**: QualityEntry and QualityDimension added to `domain/`. TabID and ViewState correctly placed in `tui/` per DC-1. Purity test continues to pass.

## What Could Be Improved

1. **Global vs. tab-specific key routing**: The `r` key conflict (M-1) reveals a gap in the key dispatch design. FR-95 defines `r` as a global refresh key, while FR-108 and FR-111/FR-112 assign `r` tab-specific meanings (REFACTOR filter, validation re-run). The specification itself has this conflict. The implementation resolved it in favor of global, but the tab-specific behaviors are broken. A whitelist/exemption approach per-tab would resolve this.

2. **Integration-level key tests**: Tab model tests call `.Update()` directly, bypassing the App's key dispatch. This is correct for unit testing but misses integration issues like the `r` key conflict. Adding App-level integration tests that send keys and verify tab state changes would catch such routing bugs.

3. **Glamour integration deferred**: FR-105 specifies Glamour for preview rendering but the implementation displays raw markdown. The dependency is in `go.mod` but unused in `tui/docs.go`. This should be wired in.

4. **Status bar cursor position**: The `info` parameter in `renderStatusBar` is always empty. Wiring per-tab cursor information (e.g., "3/22 docs") would complete FR-96.

5. **Component inlining**: 9 of 16 planned component files were inlined into tab views. While functionally equivalent, this reduces reusability and deviates from the architecture delta. Future refactoring could extract them if reuse patterns emerge.

## Metrics

| Metric | Value |
|--------|-------|
| New code (non-test) | ~2,800 lines across 20 files |
| New test code | ~2,300 lines across 14 test files |
| New tests | 128 top-level test functions |
| MUST findings | 1 |
| SHOULD findings | 5 |
| COULD findings | 3 |
| Requirements delivered | 35/37 (95%) |

## Recommendations for Next Iteration

1. **Fix M-1 first**: The `r` key routing fix is a 4-line change. Should be the first commit of any subsequent work on this branch.
2. **Add Glamour rendering**: Wire the existing dependency for preview pane markdown rendering.
3. **Wire status bar cursor info**: Pass cursor position from active tab to the status bar.
4. **Add App-level integration tests**: Test key routing through the full App.Update() dispatch chain.
