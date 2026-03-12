# Current State

## Active Work

None — iteration 005-phase-3-review-and-remediation is complete.

## Known Issues

- **SHOULD**: Preview pane uses raw content instead of Glamour rendering (S-2, tui/docs.go) — requires new glamour dependency
- **SHOULD**: 9 component files inlined into tab views instead of separate files (S-4)
- **SHOULD**: FR-88 (--check/--force exclusion) tested by code inspection only — no unit test (S-5)
- **SHOULD**: `ProjectService.DetectProject()` delegates to `fs.DetectProject()` — should use interface (reviewer S-1, iteration 004)
- **SHOULD**: 5 exported methods in fs/doc_repo.go lack GoDoc comments (NFR-8)
- **COULD**: DoctorService reimplements checks instead of delegating to ValidationService
- **COULD**: Graph rendering is flat adjacency list rather than rooted tree
- **COULD**: Quality tab uses fixed Y-axis scale (C-2)
- **COULD**: `mind preflight` prompt includes descriptor slug, not raw user request string in the /workflow suggestion

## Recent Changes

- **2026-03-12** — Phase 3 Review and Remediation complete (@iteration/005-remediation)
  - MCP `notifications/initialized` protocol fix (M-1, FR-140): server now returns nil for all `notifications/*` methods
  - Quality dimension constants aligned with conversation rubric (M-2, FR-141): 5 renames in `domain/quality.go`, parsing regex updated
  - Test coverage added for all Phase 3 packages (M-3): `internal/mcp` 80.3%, `internal/orchestrate` 81.2%, `internal/service/quality` 85%+
  - `HandoffService` extracted to `internal/orchestrate/handoff.go` (FR-146) — dead `PreflightService.Handoff()` stub removed
  - `StateRepo.AppendCurrentState()` added (FR-145) — layer violation in `cmd/handoff.go` resolved
  - Preflight blocks on hard doc failures (`docsReport.Failed > 0`, FR-147)
  - `branchAhead()` uses `mind.toml` `governance.default-branch` setting (FR-148)
  - COULD: `classify.go` adapter, stdlib cleanup, `mind preflight --json` support (FR-149–FR-151)
- **2026-03-12** — Phase 3 AI Bridge implemented (@iteration/005)
  - Model A: `mind preflight "<request>"` — 7-step pre-flight (classify, brief gate, validate docs, create iteration, git branch, write state, generate prompt)
  - Model A: `mind preflight --resume` — detect in-progress workflow from workflow.md
  - Model A: `mind handoff <iter-id>` — 5-step post-workflow (validate artifacts, run gate, update current.md, clear state, branch report)
  - Model B: `mind serve` — MCP server (JSON-RPC 2.0, stdio, 16 tools)
  - `.mcp.json` — Claude Code auto-discovery config
  - New packages: `internal/orchestrate/`, `internal/mcp/`
  - New domain type: `domain.GateResult` / `domain.GateCommandResult`
  - `StateRepo` extended with `WriteWorkflow()` (fs + mem implementations)
  - `WorkflowService` extended with `UpdateState()` and `Show()`
  - `ProjectService` extended with `ListStubs()`, `SearchDocs()`, `Config()`, `CheckBrief()`, `SuggestNext()`
  - `ValidationService` extended with `RunGate()` (executes build/lint/test from mind.toml)
  - New `QualityService` with `Log()` (parse convergence files → quality-log.yml)
  - Step 0 TUI fixes: S-1 (`$EDITOR` unset → error), S-3 (status bar cursor position)
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

- Phase 4: Watch + Orchestration (Model C/D)
- Fix remaining SHOULD/COULD items as bandwidth allows
