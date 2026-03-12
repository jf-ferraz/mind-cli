# Current State

## Active Work

None — iteration 003-phase-2-tui-dashboard is complete.

## Known Issues

- **SHOULD**: Editor defaults to `vi` instead of returning error when `$EDITOR` unset (S-1, tui/editor.go)
- **SHOULD**: Preview pane uses raw content instead of Glamour rendering (S-2, tui/docs.go)
- **SHOULD**: Status bar lacks cursor position info for lists (S-3, tui/statusbar.go)
- **SHOULD**: 9 component files inlined into tab views instead of separate files (S-4)
- **SHOULD**: FR-88 (--check/--force exclusion) tested by code inspection only — no unit test (S-5)
- **SHOULD**: Transitive propagation loses edge-type-specific reason strings at depth > 0 (S-3 from Phase 1.5)
- **SHOULD**: `--project` flag should be `--project-root` per api-contracts spec
- **SHOULD**: 5 exported methods in fs/doc_repo.go lack GoDoc comments (NFR-8)
- **COULD**: DoctorService reimplements checks instead of delegating to ValidationService
- **COULD**: Graph rendering is flat adjacency list rather than rooted tree
- **COULD**: Quality tab uses fixed Y-axis scale (C-2)

## Recent Changes

- **2026-03-11** — Phase 2 TUI Dashboard implemented (@iteration/003)
  - 37 FRs implemented (FR-88–FR-124) across 26 new files, 12 modified files
  - 374 tests, all passing (128 new tests added by tester)
  - 4 SHOULD fixes: flag exclusion, missing docs check, BuildDeps wiring, search abstraction
  - `mind tui` command with 5-tab Bubble Tea dashboard
  - New deps: bubbletea, bubbles, glamour
  - domain/ 100% coverage, tui/components/ 96.3%, tui/ 62.8%
- **2026-03-11** — Phase 1.5 Reconciliation Engine implemented (@iteration/002)
  - 37 FRs (FR-51–FR-87), 246 tests, reconciliation engine
- **2026-03-11** — Phase 1 Core CLI implemented (@iteration/001)
  - 50 FRs, 395 tests, full CLI command surface

## Next Priorities

- Fix SHOULD items from Phase 2 reviewer (Glamour preview, editor error, status bar)
- Fix remaining deferred SHOULD items (transitive reasons, flag rename, GoDoc)
- Phase 3: AI Bridge (Pre-Flight + MCP server)
