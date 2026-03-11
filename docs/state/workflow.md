# Workflow State

## Position
- **Type**: NEW_PROJECT
- **Descriptor**: core-cli
- **Iteration**: docs/iterations/001-new-project-core-cli/
- **Branch**: feature/core-cli
- **Last Agent**: architect
- **Remaining Chain**: [developer, tester, reviewer]
- **Session**: 1 of 2

## Completed Artifacts
| Agent | Output | Location |
|-------|--------|----------|
| analyst | Requirements specification | docs/spec/requirements.md |
| analyst | Domain model | docs/spec/domain-model.md |
| architect | Architecture specification | docs/spec/architecture.md |
| architect | API contracts | docs/spec/api-contracts.md |

## Dispatch Log
| Agent | Agent File | Frontmatter Model | Task Model Param | Status |
|-------|-----------|-------------------|-----------------|--------|
| analyst | .mind/agents/analyst.md | claude-opus-4-6 | opus | completed |
| architect | .mind/agents/architect.md | claude-opus-4-6 | opus | completed |

## Key Decisions (This Session)
- 4-layer architecture with domain purity (zero external imports in domain/)
- Repository interfaces defined at consumer site, not in domain
- Validation engine as composable check framework with Suite.Run()
- Output modes via renderer pattern (interactive/plain/JSON)
- Constructor injection in main.go — no init(), no global state
- Exit code strategy: 0 success, 1 validation failure, 2 runtime error, 3 config error

## Context for Next Session
The developer should:
1. Start with `docs/spec/architecture.md` (component map and layer rules)
2. Reference `docs/spec/requirements.md` (50 FRs with acceptance criteria)
3. Reference `docs/spec/api-contracts.md` (JSON schemas and command contracts)
4. Reference `docs/spec/domain-model.md` (entities, business rules, state machines)
5. Honor the 4-layer architecture: cmd/ → internal/service/ → domain/ → internal/repo/
6. Centralize DI wiring in main.go (fix current cmd/ direct-wiring pattern)
7. Follow Phase 1 package structure from BP-08
8. Build in order: domain types → repo interfaces/implementations → services → validation engine → rendering → commands
