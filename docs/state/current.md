# Current State

## Active Work

None — iteration 004-pre-phase-3-cleanup is complete.

## Known Issues

- **SHOULD**: Editor defaults to `vi` instead of returning error when `$EDITOR` unset (S-1, tui/editor.go)
- **SHOULD**: Preview pane uses raw content instead of Glamour rendering (S-2, tui/docs.go)
- **SHOULD**: Status bar lacks cursor position info for lists (S-3, tui/statusbar.go)
- **SHOULD**: 9 component files inlined into tab views instead of separate files (S-4)
- **SHOULD**: FR-88 (--check/--force exclusion) tested by code inspection only — no unit test (S-5)
- **SHOULD**: `ProjectService.DetectProject()` delegates to `fs.DetectProject()` — should use interface (reviewer S-1, iteration 004)
- **SHOULD**: 5 exported methods in fs/doc_repo.go lack GoDoc comments (NFR-8)
- **COULD**: DoctorService reimplements checks instead of delegating to ValidationService
- **COULD**: Graph rendering is flat adjacency list rather than rooted tree
- **COULD**: Quality tab uses fixed Y-axis scale (C-2)

## Recent Changes

- **2026-03-11** — Pre-Phase 3 Cleanup implemented (@iteration/004)
  - 15 FRs implemented (FR-125–FR-139) across 4 new files, 14 modified files
  - Deps struct migrated from concrete `*fs.` types to `repo.` interfaces (FR-125, FR-137)
  - `IsStubContent()` moved to `internal/repo/` to eliminate mem/ -> fs/ inverse dependency (FR-126)
  - Transitive staleness propagation edge-type reasons fixed at all depths (FR-127)
  - `--project` flag renamed to `--project-root` (FR-128)
  - `os.Exit()` calls replaced with `ExitError` returns in all cmd/ handlers (FR-129)
  - `DiagnosticStatus` typed enum added (FR-130)
  - cmd/ exit code tests and render/ JSON tests added (FR-131, FR-132)
  - Spec docs updated (FR-133–FR-136)
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

- Fix remaining SHOULD items (Glamour preview, editor error, status bar)
- Phase 3: AI Bridge (Pre-Flight + MCP server)
