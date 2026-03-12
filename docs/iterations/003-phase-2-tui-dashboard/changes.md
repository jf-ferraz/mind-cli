# Phase 2: TUI Dashboard -- Changes

## SHOULD Fixes (Steps 1-4)

### FR-88: Mutual Exclusion Guard
- `cmd/reconcile.go`: Added 4-line guard at top of `runReconcile()` to reject `--check` + `--force` together (exit 2).

### FR-89: Missing Documents Check
- `internal/validate/reconcile.go`: Inserted "No missing documents" check between cycle check and stale checks. Produces WARN per missing doc (FAIL with `--strict`).
- `internal/validate/reconcile_test.go`, `reconcile_extended_test.go`: Updated all test expectations to account for the additional check (+1 total in each report).

### FR-90: Wiring Centralization (BuildDeps)
- `internal/deps/deps.go` (new): Extracted `Deps` struct and `Build()` function into shared package. Breaks the import cycle between `cmd` and `tui`.
- `cmd/root.go`: Replaced inline `Deps` struct and `BuildDeps()` with type alias (`Deps = deps.Deps`) and wrapper function delegating to `deps.Build()`. Package-level variables still populated for backward compatibility.

### FR-91: DocRepo Search Abstraction
- `internal/repo/interfaces.go`: Added `Search(query string) (*domain.SearchResults, error)` to DocRepo interface; added `QualityRepo` interface.
- `internal/repo/fs/doc_repo.go`: Implemented `Search()` -- moved `filepath.WalkDir` logic from `cmd/docs.go`.
- `internal/repo/mem/doc_repo.go`: Implemented `Search()` for in-memory testing.
- `cmd/docs.go`: Simplified `runDocsSearch` to delegate to `docRepo.Search()`.

## Domain Types (Step 5)

- `domain/quality.go` (new): `QualityEntry` struct with `Validate()` for BR-36/37/38, `QualityDimension` struct, dimension name constants.
- `internal/repo/fs/quality_repo.go` (new): Reads `quality-log.yml` from root or `docs/knowledge/`. Returns empty slice if file absent.
- `internal/repo/mem/quality_repo.go` (new): In-memory `QualityRepo` for testing.

## TUI Foundation (Step 6)

- `tui/styles.go` (new): `Theme` struct with zone colors, severity styles, chrome styles, content styles, iteration type colors. `DefaultTheme()` per BP-05 Section 6. Package-level `theme` variable.
- `tui/keys.go` (new): `GlobalKeyMap` (quit, force-quit, help, refresh, tab 1-5, next/prev tab), `NavigationKeyMap` (up/down).
- `tui/messages.go` (new): Custom `tea.Msg` types for async data loading (health, iterations, quality, validation, preview), `tabActivatedMsg` for lazy loading.
- `tui/types.go` (new): `TabID` enum (0-4), `TabCount`, `TabNames`, `ViewState` enum (Loading/Error/Empty/Ready), `MinWidth`/`MinHeight` constants.

## TUI App Shell + Status Tab (Step 7)

- `tui/app.go` (new): Top-level Bubble Tea model. Title bar (project name, git branch, version), tab bar, separator, content area, status bar, help overlay. Handles `WindowSizeMsg`, global keys, tab switching with `tabActivatedMsg`, async data loading (`loadHealth`, `loadIterations`, `loadQuality`). "Too small" terminal guard. `overlayOnScreen` for help overlay.
- `tui/status.go` (new): Tab 1 -- two-column layout (zone bars + staleness + warnings + suggestions on left, workflow panel + quick actions on right). Single-column fallback below MinWidth.
- `tui/statusbar.go` (new): Context-sensitive key hints per tab.
- `tui/help.go` (new): Context-sensitive help overlay with global keys and tab-specific keys.
- `tui/components/empty_state.go` (new): Centered empty state message.
- `tui/components/zone_bar.go` (new): Zone progress bar with colored fill and fraction label.
- `tui/components/staleness.go` (new): Stale document list with bullet markers.
- `tui/components/workflow_panel.go` (new): Workflow state panel (idle/running, last iteration, agent chain).
- `tui/components/warnings.go` (new): Warning list.
- `tui/components/suggestions.go` (new): Suggestion list.
- `tui/components/quick_actions.go` (new): Quick action key reference for Status tab.

## Documents Tab (Step 8)

- `tui/docs.go` (new): Tab 2 -- document list with zone filter bar (a/s/b/t/i/k), search mode (/), preview pane (Enter), editor integration (e via $EDITOR). Split-pane preview (40/60). Zone grouping with headers.

## Iterations Tab (Step 9)

- `tui/iterations.go` (new): Tab 3 -- iteration table with type filter bar (a/n/e/b/r for all/new/enhance/bugfix/refactor), expand/collapse detail (Enter), artifact presence indicators, type-colored labels.

## Checks Tab (Step 10)

- `tui/checks.go` (new): Tab 4 -- accordion validation suites with lazy loading (FR-122). Spinner while running. Suite expand/collapse (Enter), check detail toggle (Space), re-run validation (r). Overall summary with pass/fail/warning counts.

## Quality Tab (Step 11)

- `tui/quality.go` (new): Tab 5 -- ASCII score chart with Y-axis (1.0-5.0), Gate 0 threshold line at 3.0, data point navigation (left/right or h/l). Selected analysis detail with dimension bars, personas, output path.

## Polish and Integration (Step 12)

- `tui/editor.go` (new): `editorCmd` helper returning `exec.Cmd` for `$EDITOR`.
- `tui/util.go` (new): `readFile` helper.
- `cmd/tui.go` (new): `mind tui` Cobra command. Resolves project root, builds deps with nil renderer, launches Bubble Tea with `tea.WithAltScreen()`.
- Terminal resize: `WindowSizeMsg` propagated to all 5 tabs in `App.Update()`.
- "Too small" message: Rendered when terminal is below 80x24 (FR-118).
- Tab state preservation: Tab models kept in `App` struct, not recreated on switch (FR-123).
- Terminal cleanup: `tea.WithAltScreen()` ensures terminal restored on quit (FR-124).

## Reviewer Fixes

### M-1: Global `r` key intercepts tab-specific handlers (FR-108, FR-111/FR-112)
- `tui/app.go`: Added tab exemption in global `Refresh` key handler — when active tab is `TabIterations` (Tab 3) or `TabChecks` (Tab 4), `r` is not handled globally and falls through to the tab's `Update()` method. This allows `r` to toggle the REFACTOR type filter on Iterations and to re-run validation on Checks.

## Architecture Notes

- Import cycle resolved by extracting `Deps` to `internal/deps/` package. `cmd` imports `internal/deps` (via type alias), `tui` imports `internal/deps` directly. Neither `tui` nor `cmd` imports the other.
- Some component files from the architecture delta (zone_filter, type_filter, detail_expander, suite_accordion, check_detail, overall_summary, score_chart, latest_analysis) were inlined into their respective tab views for simplicity. The functionality is identical.
- The `k` key in the Docs tab is reserved for the knowledge zone filter; vim-style up navigation uses `up` arrow only in that tab.
