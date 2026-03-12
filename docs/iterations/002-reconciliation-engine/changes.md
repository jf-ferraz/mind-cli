# Changes

- **Iteration**: 002-reconciliation-engine
- **Date**: 2026-03-11

## Files Changed

### Step 1: Domain Types (`2c0554e`)
| File | Action | Description |
|------|--------|-------------|
| `domain/reconcile.go` | Created | EdgeType, LockStatus, EntryStatus enums; GraphEdge, Graph, LockFile, LockEntry, LockStats, ReconcileResult, ReconcileOpts, StalenessInfo types; BuildGraph() |
| `domain/reconcile_test.go` | Created | Tests for BuildGraph, enum values, JSON round-trips |
| `domain/project.go` | Modified | Added `Graph []GraphEdge` field to Config |
| `domain/health.go` | Modified | Added `Staleness *StalenessInfo` field to ProjectHealth |

### Step 2: Config Extension (`01641b9`)
| File | Action | Description |
|------|--------|-------------|
| `internal/validate/config.go` | Modified | Added checks 11-12: graph edge IDs and types validation |
| `internal/validate/config_test.go` | Modified | Tests for graph validation checks, updated suite count 10->12 |
| `internal/repo/fs/config_repo_test.go` | Created | Tests for [[graph]] TOML parsing |
| `internal/service/validation_test.go` | Modified | Updated expected check counts: RunConfig 10->12, RunAll 38->40 |

### Step 3: Hash Computation (`6dd8642`)
| File | Action | Description |
|------|--------|-------------|
| `internal/reconcile/hash.go` | Created | HashFile() SHA-256, NeedsRehash() mtime fast-path |
| `internal/reconcile/hash_test.go` | Created | Known content, empty file, not found, symlink, mtime tests |

### Step 4: Graph Operations (`9612e2d`)
| File | Action | Description |
|------|--------|-------------|
| `internal/reconcile/graph.go` | Created | DetectCycle() DFS, ValidateEdges() |
| `internal/reconcile/graph_test.go` | Created | Acyclic, cycle, self-loop, disconnected, edge validation tests |

### Step 5: Staleness Propagation (`4d3f006`)
| File | Action | Description |
|------|--------|-------------|
| `internal/reconcile/propagate.go` | Created | BFS PropagateDownstream() with MaxPropagationDepth=10 |
| `internal/reconcile/propagate_test.go` | Created | Linear chain, diamond, depth limit, edge type reason tests |

### Step 6: Lock Repository (`c46bddd`)
| File | Action | Description |
|------|--------|-------------|
| `internal/repo/interfaces.go` | Modified | Added LockRepo interface |
| `internal/repo/fs/lock_repo.go` | Created | Filesystem LockRepo with atomic writes (temp+rename) |
| `internal/repo/fs/lock_repo_test.go` | Created | Read missing, write+read, atomic write, round-trip tests |
| `internal/repo/mem/lock_repo.go` | Created | In-memory LockRepo with JSON deep copy |

### Step 7: Engine Orchestration (`2af717e`)
| File | Action | Description |
|------|--------|-------------|
| `internal/reconcile/engine.go` | Created | 6-phase reconciliation engine orchestrator |
| `internal/reconcile/engine_test.go` | Created | First run, incremental, force, cycle, missing, undeclared tests |

### Step 8: ReconciliationService (`724f07e`)
| File | Action | Description |
|------|--------|-------------|
| `internal/service/reconciliation.go` | Created | Service layer: Reconcile(), ReadStaleness() |
| `internal/service/reconciliation_test.go` | Created | Service reconcile, check-only, no config, staleness tests |

### Step 9: Centralized Wiring (`3ff5939`)
| File | Action | Description |
|------|--------|-------------|
| `cmd/root.go` | Modified | PersistentPreRunE with centralized repo/service construction |
| `cmd/check.go` | Modified | Simplified to use centralized validationSvc |
| `cmd/brief.go` | Modified | Simplified to use centralized briefRepo |
| `cmd/iterations.go` | Modified | Simplified to use centralized iterRepo |
| `cmd/docs.go` | Modified | Simplified to use centralized docRepo, renderer |
| `cmd/workflow.go` | Modified | Simplified to use centralized workflowSvc |
| `cmd/create.go` | Modified | Simplified to use centralized generateSvc |
| `cmd/doctor.go` | Modified | Simplified to use centralized doctorSvc |
| `cmd/status.go` | Modified | Simplified, added staleness panel via reconcileSvc |
| `internal/service/doctor.go` | Modified | Added lockRepo parameter to constructor |

### Step 10: Reconcile Command (`4429038`)
| File | Action | Description |
|------|--------|-------------|
| `cmd/reconcile.go` | Created | mind reconcile with --check, --force, --graph flags |
| `internal/render/render.go` | Modified | Added RenderReconcileResult(), RenderGraph() |
| `internal/service/reconciliation.go` | Modified | Added LoadGraph() method |

### Step 11: Integration (`06f6521`)
| File | Action | Description |
|------|--------|-------------|
| `internal/validate/reconcile.go` | Created | ReconcileSuite from pre-computed ReconcileResult |
| `internal/validate/reconcile_test.go` | Created | Nil result, clean, stale, strict mode tests |
| `internal/service/validation.go` | Modified | RunAll accepts variadic reconcile result |
| `internal/service/doctor.go` | Modified | Added checkStaleness() for FR-81 |
| `internal/render/render.go` | Modified | Staleness panel in renderHealthText() |
| `cmd/check.go` | Modified | runCheckAll passes reconcile result to RunAll |

### Step 12: Performance Benchmarks (`9c495b8`)
| File | Action | Description |
|------|--------|-------------|
| `internal/reconcile/bench_test.go` | Created | Full/incremental/force benchmarks, performance verification test |

## Summary

| Metric | Count |
|--------|-------|
| New files | 20 |
| Modified files | 18 |
| New FRs covered | 37 (FR-51 through FR-87) |
| New test files | 12 |
| Benchmark results | Full 50 docs: ~1.1ms (<200ms target), Incremental: ~0.45ms (<50ms target) |
