# Reconciliation Engine

- **Type**: COMPLEX_NEW
- **Request**: Analyze current project status and next recommended actions. After convergence, proceed with next required implementations as described in all blueprints documents — Phase 1.5: Reconciliation Engine.
- **Agent Chain**: conversation-moderator → analyst → architect → developer → tester → reviewer
- **Branch**: feature/reconciliation-engine
- **Created**: 2026-03-11

## Scope

Phase 1.5 delivers the reconciliation engine: hash-based content tracking with staleness propagation through a dependency graph. This includes the `mind reconcile` command (with `--check`, `--force`, `--graph` flags), integration with existing `mind status`, `mind check all`, and `mind doctor` commands, and the `mind.lock` file lifecycle. No new external dependencies — SHA-256 is in Go's standard library.

## Prior Analysis Context
- **Source**: docs/knowledge/reconciliation-engine-convergence.md
- **Key Recommendations**:
  1. Resolve 5 blueprint inconsistencies before writing code (95% HIGH)
  2. Implement in 12-step sequence: domain → config → hash → graph → propagate → lock_repo → engine → service → wiring → cmd → integration → perf (85% HIGH)
  3. Fix wiring centralization during step 9 — concurrent with Phase 1.5 (90% HIGH)
  4. Defer all other tech debt past Phase 1.5 (80% MEDIUM)
  5. Use ReconcileSuite for `mind check all` integration (75% MEDIUM)
- **Decision Matrix Winner**: Option C (Blueprint-first + Test-driven) — scored 4.20/5.00

## Requirement Traceability

| Req ID | Description | Analyst | Architect | Developer | Reviewer |
|--------|-------------|---------|-----------|-----------|----------|
| FR-51 | `mind reconcile` core command (hash, graph, propagate, write lock) | ✓ | | | |
| FR-52 | `mind reconcile --check` (read-only verify, exit 0 or 4) | ✓ | | | |
| FR-53 | `mind reconcile --force` (discard lock, re-hash, clear staleness) | ✓ | | | |
| FR-54 | `mind reconcile --graph` (ASCII tree visualization) | ✓ | | | |
| FR-55 | `mind reconcile --json` (JSON output) | ✓ | | | |
| FR-56 | `mind reconcile` requires valid mind.toml | ✓ | | | |
| FR-57 | SHA-256 hash of raw bytes, no normalization | ✓ | | | |
| FR-58 | mtime fast-path optimization | ✓ | | | |
| FR-59 | Hash edge cases (empty, binary, symlink, >10MB, unreadable) | ✓ | | | |
| FR-60 | Dependency graph construction from [[graph]] | ✓ | | | |
| FR-61 | Three edge types with differentiated messages | ✓ | | | |
| FR-62 | Cycle detection with full path reporting | ✓ | | | |
| FR-63 | Graph edge validation against [documents] | ✓ | | | |
| FR-64 | No-graph mode (hash tracking without propagation) | ✓ | | | |
| FR-65 | Downstream-only staleness propagation | ✓ | | | |
| FR-66 | Transitive propagation with path info | ✓ | | | |
| FR-67 | Depth limit of 10 with warning | ✓ | | | |
| FR-68 | Changed documents are fresh, not stale | ✓ | | | |
| FR-69 | No duplicate processing via multiple paths | ✓ | | | |
| FR-70 | Lock file at project root as mind.lock | ✓ | | | |
| FR-71 | Lock file JSON schema (entries, stats, status) | ✓ | | | |
| FR-72 | Lock file round-trip correctness | ✓ | | | |
| FR-73 | Atomic lock file writes (temp + rename) | ✓ | | | |
| FR-74 | First-run behavior (no prior lock) | ✓ | | | |
| FR-75 | is_stub via DocRepo.IsStub() delegation | ✓ | | | |
| FR-76 | Lock status derivation (CLEAN/STALE/DIRTY) | ✓ | | | |
| FR-77 | `mind status` staleness panel (read-only) | ✓ | | | |
| FR-78 | `mind status --json` staleness object | ✓ | | | |
| FR-79 | `mind check all` ReconcileSuite integration | ✓ | | | |
| FR-80 | `mind check all --json` reconcile suite entry | ✓ | | | |
| FR-81 | `mind doctor` stale document findings | ✓ | | | |
| FR-82 | Exit code 4 for staleness | ✓ | | | |
| FR-83 | mind.toml [[graph]] section support | ✓ | | | |
| FR-84 | Config validation for [[graph]] entries | ✓ | | | |
| FR-85 | Undeclared file detection | ✓ | | | |
| FR-86 | Full reconciliation <200ms for 50 docs | ✓ | | | |
| FR-87 | Incremental reconciliation <50ms for 50 docs | ✓ | | | |
