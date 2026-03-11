# Current State

## Active Work

None — iteration 002-reconciliation-engine is complete.

## Known Issues

- **SHOULD**: `--check` and `--force` flags not enforced as mutually exclusive in `cmd/reconcile.go` (S-1)
- **SHOULD**: `ReconcileSuite` omits "missing documents" check specified in FR-79 (S-2)
- **SHOULD**: Transitive propagation loses edge-type-specific reason strings at depth > 0 (S-3)
- **SHOULD**: `--project` flag should be `--project-root` per api-contracts spec
- **SHOULD**: `docs search` bypasses DocRepo abstraction (C-9 deviation)
- **SHOULD**: 5 exported methods in fs/doc_repo.go lack GoDoc comments (NFR-8)
- **SHOULD**: Repo wiring in command handlers instead of main.go (C-10 deviation, acknowledged — partially fixed by wiring centralization in Phase 1.5)
- **COULD**: DoctorService reimplements checks instead of delegating to ValidationService
- **COULD**: Graph rendering is flat adjacency list rather than rooted tree (C-1)
- **COULD**: `isConfigError` uses string matching rather than typed errors (C-2)

## Recent Changes

- **2026-03-11** — Phase 1.5 Reconciliation Engine implemented (@iteration/002)
  - 37 FRs implemented (FR-51–FR-87) across 20 new files, 18 modified files
  - 246 tests, all passing (82 new tests added by tester)
  - Reconciliation engine: hash, graph, propagation, lock lifecycle
  - `mind reconcile` command with `--check`, `--force`, `--graph` flags
  - Integration with `mind status`, `mind check all`, `mind doctor`
  - Wiring centralization via PersistentPreRunE in cmd/root.go
  - Performance: full <200ms (actual ~1.1ms), incremental <50ms (actual ~0.45ms)
- **2026-03-11** — Phase 1 Core CLI implemented (@iteration/001)
  - 50 FRs implemented across 20+ commands
  - 395 tests, all passing
  - domain/ 100% coverage, validate/ 90.7%

## Next Priorities

- Fix SHOULD items S-1, S-2, S-3 from Phase 1.5 reviewer
- Fix remaining Phase 1 SHOULD items (flag rename, GoDoc, search abstraction)
- Phase 2: TUI dashboard, interactive mode enhancements
