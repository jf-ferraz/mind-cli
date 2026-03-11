# Test Summary

- **Iteration**: 002-reconciliation-engine
- **Date**: 2026-03-11
- **Tester**: Claude Opus 4.6

## Baseline

All existing tests passed before any new tests were added.

## Coverage

| Package | Before | After | Notes |
|---------|--------|-------|-------|
| `domain/` | 100.0% | 100.0% | Maintained full coverage |
| `internal/reconcile/` | 92.2% | 95.4% | +3.2% from edge case and FR tests |
| `internal/service/` | 55.2% | 76.5% | +21.3% from LoadGraph, doctor, RunAll tests |
| `internal/repo/fs/` | 46.0% | 46.4% | +0.4% from lock repo schema/round-trip tests |
| `internal/validate/` | 91.1% | 91.1% | Maintained via extended reconcile suite tests |
| **Overall** | **47.2%** | **51.5%** | +4.3% net improvement |

### Key Function Coverage

| Function | Coverage |
|----------|----------|
| `reconcile.Reconcile()` (engine) | 96.9% |
| `reconcile.PropagateDownstream()` | 100% |
| `reconcile.DetectCycle()` | 100% |
| `reconcile.ValidateEdges()` | 100% |
| `reconcile.HashFile()` | 87.5% |
| `reconcile.NeedsRehash()` | 100% |
| `reconcile.edgeTypeReason()` | 100% |
| `reconcile.buildReason()` | 100% |
| `service.ReconciliationService.Reconcile()` | 84.2% |
| `service.ReconciliationService.LoadGraph()` | 100% (was 0%) |
| `service.ReconciliationService.ReadStaleness()` | 83.3% |
| `service.DoctorService.Run()` | 94.1% |
| `validate.ReconcileSuite()` | 100% |
| `fs.LockRepo.Read/Write/Exists` | 88.9%/71.4%/100% |

## Test Counts

| Category | Files Added | Tests Added |
|----------|-------------|-------------|
| Hash computation | 1 | 10 |
| Graph operations | 1 | 7 |
| Staleness propagation | 1 | 8 |
| Engine orchestration | 1 | 15 |
| Lock repository | 1 | 6 |
| Domain types | 1 | 13 |
| Reconcile suite | 1 | 6 |
| Service: reconciliation | 1 | 8 |
| Service: validation+reconcile | 1 | 5 |
| Service: doctor staleness | 1 | 4 |
| **Total** | **10** | **82** |

All 246 tests (existing + new) pass. Zero failures.

## Acceptance Criteria Coverage

### Reconcile Command (FR-51 through FR-56)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-51 | First run creates lock with all docs, stale=false | `TestEngine_FR51_FirstRunCreatesCleanLock` | PASS |
| FR-52 | --check does not write lock, exit 4 on stale | `TestReconciliationService_FR52_CheckOnlyNoWrite` | PASS |
| FR-53 | --force clears all staleness | `TestEngine_FR53_ForceClearsStaleness`, `TestReconciliationService_FR53_ForceClearsStaleness` | PASS |
| FR-54 | --graph ASCII tree | Not unit-testable (render layer, no test files) | DEFERRED |
| FR-55 | --json output schema | `TestReconcileResult_JSONFields` | PASS |
| FR-56 | Missing config exits 3 | `TestReconciliationService_FR56_MissingConfig`, `TestReconciliationService_FR56_NoDocuments` | PASS |

### Hash Computation (FR-57 through FR-59)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-57 | SHA-256 raw bytes, sha256: prefix, no normalization | `TestHashFile_RawBytesNoNormalization`, `TestHashFile_FormatCorrect`, `TestHashFile_MatchesManualSHA256` | PASS |
| FR-58 | mtime fast-path | `TestNeedsRehash_MtimeMatch`, `TestNeedsRehash_MtimeDiffers`, `TestNeedsRehash_SizeDiffers`, `TestNeedsRehash_NilEntry`, `TestNeedsRehash_EmptyHash` | PASS |
| FR-59 | Empty file hash, binary warning, symlink, unreadable | `TestHashFile_EmptyFileKnownHash`, `TestEngine_BinaryFileWarning`, `TestHashFile_SymlinkResolvesToTarget`, `TestHashFile_Unreadable` | PASS |

### Dependency Graph (FR-60 through FR-64)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-60 | Graph with forward/reverse edges | `TestEngine_FR60_GraphEdgeCounts`, `TestBuildGraph_BasicEdges`, `TestBuildGraph_SingleEdge` | PASS |
| FR-61 | Edge type reason messages | `TestPropagateDownstream_FR61_InformsReason`, `_RequiresReason`, `_ValidatesReason` | PASS |
| FR-62 | Cycle detection with full path | `TestEngine_FR62_CyclePathInError`, `TestDetectCycle_SimpleCycle`, `TestDetectCycle_SelfLoopPath`, `TestDetectCycle_ComplexCycle` | PASS |
| FR-63 | Undeclared doc ID in graph | `TestEngine_FR63_UndeclaredDocInGraph`, `TestValidateEdges_UndeclaredFrom`, `TestValidateEdges_UndeclaredTo` | PASS |
| FR-64 | No graph still tracks changes | `TestEngine_FR64_NoGraphStillTracksChanges`, `TestEngine_NoGraph` | PASS |

### Staleness Propagation (FR-65 through FR-69)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-65 | Downstream only | `TestEngine_FR65_DownstreamOnly`, `TestPropagateDownstream_ReverseNotPropagated` | PASS |
| FR-66 | Transitive propagation | `TestEngine_FR66_TransitivePropagation`, `TestPropagateDownstream_LinearChain` | PASS |
| FR-67 | Depth limit 10 | `TestPropagateDownstream_FR67_DepthLimitExact`, `TestPropagateDownstream_DepthLimit` | PASS |
| FR-68 | Changed not stale | `TestEngine_FR68_ChangedNotStale`, `TestPropagateDownstream_ChangedNotStale` | PASS |
| FR-69 | Diamond graph, stale once | `TestEngine_FR69_DiamondOnceOnly`, `TestPropagateDownstream_DiamondGraph` | PASS |

### Lock File (FR-70 through FR-76)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-70 | Lock at mind.lock | `TestLockRepo_FR70_LockFileLocation` | PASS |
| FR-71 | JSON schema fields | `TestLockRepo_FR71_JSONSchema` | PASS |
| FR-72 | Byte-identical round-trip | `TestLockRepo_FR72_ByteIdenticalRoundTrip`, `TestLockRepo_RoundTrip` | PASS |
| FR-73 | Atomic write | `TestLockRepo_FR73_AtomicWrite`, `TestLockRepo_AtomicWrite` | PASS |
| FR-74 | First run clean | `TestEngine_FR74_FirstRunClean`, `TestEngine_FirstRun` | PASS |
| FR-75 | is_stub via DocRepo | Verified via engine scan (BR-34 delegation) | PASS |
| FR-76 | Status priority | `TestEngine_FR76_StatusPriority` (CLEAN/DIRTY/STALE sub-tests) | PASS |

### Integration (FR-77 through FR-81)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-77 | status staleness panel (read-only) | `TestReconciliationService_ReadStaleness_*` | PASS |
| FR-78 | status --json staleness null/object | `TestProjectHealth_StalenessNull`, `TestProjectHealth_StalenessObject` | PASS |
| FR-79 | check all includes reconcile suite | `TestValidationService_RunAll_WithReconcileResult` | PASS |
| FR-80 | check all --json reconcile suite | Verified via RunAll JSON serialization | PASS |
| FR-81 | doctor stale findings | `TestDoctorService_FR81_StaleDiagnostics`, `_NoLock`, `_NilLockRepo`, `_CleanLock` | PASS |

### Exit Codes and Config (FR-82 through FR-85)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-82 | Exit code 4 (staleness) | Integration-level (cmd layer, not unit-testable) | DEFERRED |
| FR-83 | [[graph]] TOML parsing | `TestConfigRepo_GraphParsing` (in config_repo_test.go) | PASS |
| FR-84 | Graph edge validation | `TestCheckGraphEdgeIDs`, `TestCheckGraphEdgeTypes` | PASS |
| FR-85 | Undeclared file detection | `TestEngine_FR85_UndeclaredFiles` | PASS |

### Performance (FR-86, FR-87)

| FR | Criteria | Test | Status |
|----|----------|------|--------|
| FR-86 | Full <200ms for 50 docs | `TestBenchmark_Performance` (full: ~1.1ms) | PASS |
| FR-87 | Incremental <50ms for 50 docs | `TestBenchmark_Performance` (incr: ~0.45ms) | PASS |

## Business Rules Coverage

| BR | Rule | Tests | Status |
|----|------|-------|--------|
| BR-24 | SHA-256 raw bytes, sha256:{hex} format | `TestHashFile_RawBytesNoNormalization`, `TestHashFile_FormatCorrect`, `TestHashFile_MatchesManualSHA256` | PASS |
| BR-25 | mtime fast-path: false negatives OK, no false positives | `TestNeedsRehash_*`, `TestNeedsRehash_MtimeTouchSameContent` | PASS |
| BR-26 | Downstream-only propagation | `TestEngine_FR65_DownstreamOnly`, `TestPropagateDownstream_ReverseNotPropagated` | PASS |
| BR-27 | Depth limit 10 | `TestPropagateDownstream_FR67_DepthLimitExact`, `TestMaxPropagationDepth` | PASS |
| BR-28 | Changed != stale | `TestEngine_FR68_ChangedNotStale`, `TestPropagateDownstream_ChangedNotStale` | PASS |
| BR-29 | Cycles invalid, abort with path | `TestEngine_CycleDetection`, `TestEngine_FR62_CyclePathInError`, `TestDetectCycle_*` | PASS |
| BR-30 | Undeclared graph refs are errors | `TestEngine_UndeclaredEdges`, `TestEngine_FR63_UndeclaredDocInGraph`, `TestValidateEdges_*` | PASS |
| BR-31 | Atomic lock writes | `TestLockRepo_FR73_AtomicWrite`, `TestLockRepo_AtomicWrite` | PASS |
| BR-32 | Exit code 4 for --check | Integration-level (cmd layer) | DEFERRED |
| BR-33 | Status priority STALE>DIRTY>CLEAN | `TestEngine_FR76_StatusPriority` | PASS |
| BR-34 | is_stub via DocRepo.IsStub() | Engine delegates to docRepo (verified in engine tests) | PASS |
| BR-35 | All edge types propagate, differentiated messages | `TestPropagateDownstream_EdgeTypeReasons`, `TestEdgeTypeReason` | PASS |

## Edge Cases Tested

- Empty file hash (known constant)
- Binary file detection and warning
- Symlink resolution (target content hashed)
- Symlink outside project root warning
- Unreadable file error handling
- Self-loop cycle detection
- Complex cycles with multiple possible paths
- Diamond graph (stale once, not twice)
- Long acyclic chains (no false cycle detection)
- Depth limit at exact boundary (10)
- Multiple changed sources propagation
- No graph edges (hash-only tracking)
- Missing documents (DIRTY status)
- Pruning removed documents from lock (XC-11)
- Corrupted JSON lock file
- Lock file overwrite behavior
- Lock file trailing newline
- Nil reconcile result in suite
- Nil lock repo in doctor
- Empty documents section in config

## Deferred Items

1. **FR-54 (--graph ASCII tree)**: Rendering is in `internal/render/` which has no test files. Testing render output is presentation-level verification that would require either capturing stdout or refactoring the render layer for testability.

2. **FR-82 (Exit code 4)**: Exit codes are set in `cmd/reconcile.go` which uses Cobra command handlers. Testing exit codes requires either integration tests that execute the CLI binary or refactoring to extract exit code logic into testable functions.

3. **cmd/ package**: Zero coverage because Cobra command handlers require integration-level testing (executing the binary, checking stdout/stderr/exit codes). This is a pre-existing gap, not specific to Phase 1.5.

## Regression

All existing Phase 1 tests continue to pass:
- 17-check docs validation suite
- 11-check refs validation suite
- 10+2 check config validation suite (extended with graph checks)
- Service tests: RunDocs, RunRefs, RunConfig, RunAll
- Domain purity test (zero external imports)
- All generate, fs repo, and iteration tests

The test count for `RunAll` was updated from 38 to 40 (adding 2 graph config checks) by the developer in Step 2. Our new tests verify the further extension to 4 suites when reconcile result is provided.
