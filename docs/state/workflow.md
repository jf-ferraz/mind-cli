# Workflow State

## Position
- **Type**: COMPLEX_NEW
- **Descriptor**: reconciliation-engine
- **Iteration**: docs/iterations/002-reconciliation-engine/
- **Branch**: feature/reconciliation-engine
- **Last Agent**: architect
- **Remaining Chain**: [developer, tester, reviewer]
- **Session**: 1 of 2

## Completed Artifacts
| Agent | Output | Location |
|-------|--------|----------|
| conversation-moderator | Convergence analysis (4.33/5.0) | docs/knowledge/reconciliation-engine-convergence.md |
| analyst | Requirements delta (37 FRs: FR-51–FR-87) | docs/iterations/002-reconciliation-engine/requirements-delta.md |
| analyst | Requirements update | docs/spec/requirements.md |
| analyst | Domain model update | docs/spec/domain-model.md |
| architect | Architecture delta (6 decisions, 12-step migration) | docs/iterations/002-reconciliation-engine/architecture-delta.md |
| architect | Architecture spec update | docs/spec/architecture.md |
| architect | API contracts update | docs/spec/api-contracts.md |

## Dispatch Log
| Agent | Agent File | Frontmatter Model | Task Model Param | Status |
|-------|-----------|-------------------|-----------------|--------|
| conversation-moderator | .mind/conversation/agents/moderator.md | claude-opus-4-6 | opus | completed |
| analyst | .mind/agents/analyst.md | claude-opus-4-6 | opus | completed |
| architect | .mind/agents/architect.md | claude-opus-4-6 | opus | completed |

## Key Decisions (This Session)
- 5 blueprint inconsistencies resolved: no hash normalization, all edge types propagate, hash.go allowed direct I/O, is_stub in lock, exit code 4
- 12-step implementation sequence: domain → config → hash → graph → propagate → lock_repo → engine → service → wiring → cmd → integration → perf
- Wiring centralization via PersistentPreRunE in cmd/root.go (step 9)
- ReconcileSuite integrates with check framework using pre-computed results
- Engine orchestrates 6 phases, service handles I/O boundaries
- Reconcile types in single domain/reconcile.go file

## Context for Next Session
The developer should:
1. Start with `docs/iterations/002-reconciliation-engine/architecture-delta.md` — the 12-step migration path (sections "Step 1" through "Step 12")
2. Reference `docs/iterations/002-reconciliation-engine/requirements-delta.md` for FR-51 through FR-87 acceptance criteria
3. Reference `docs/spec/api-contracts.md` Phase 1.5 section for JSON schemas (mind.lock, reconcile output, staleness info)
4. Reference `docs/blueprints/06-reconciliation-engine.md` for algorithm pseudocode (hash, graph, propagation, engine)
5. Follow the existing patterns in `domain/`, `internal/repo/`, `internal/service/`, `internal/validate/`, `cmd/`
6. Key architectural decisions to honor:
   - Domain purity: domain/reconcile.go has zero external imports
   - hash.go accepts absolute paths, does direct os.Open() — this is a deliberate exception
   - ReconcileSuite uses pre-computed ReconcileResult projected into CheckResults
   - Wiring centralization in step 9 via PersistentPreRunE
   - Engine phases match BP-06 Section 6 pseudocode
   - Atomic lock writes via temp file + rename
7. All 37 FRs (FR-51–FR-87) must be addressed with corresponding changes
8. Update docs/iterations/002-reconciliation-engine/changes.md as files are created/modified
