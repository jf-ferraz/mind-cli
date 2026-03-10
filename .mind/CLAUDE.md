# Mind Agent Framework

> Framework root: `.mind/` — all paths below are relative to project root.

| Resource | When to Read |
|----------|-------------|
| `.mind/README.md` | Framework overview, architecture, design decisions, agent registry |
| `.mind/agents/orchestrator.md` | Starting any workflow — classifies requests, routes to agent chains |
| `.mind/agents/analyst.md` | Requirements analysis, context gathering, scope definition |
| `.mind/agents/architect.md` | System design, structural decisions (NEW_PROJECT or structural changes) |
| `.mind/agents/developer.md` | Implementation, code writing, incremental changes |
| `.mind/agents/tester.md` | Test strategy, test implementation, coverage verification |
| `.mind/agents/reviewer.md` | Code review, quality validation, final sign-off |
| `.mind/agents/discovery.md` | Interactive project exploration — vague ideas → structured briefs |
| `.mind/agents/technical-writer.md` | Documentation creation optimized for LLM consumption |
| `.mind/conversation/agents/moderator.md` | Conversation analysis — spawns personas, manages 7-phase dialectical workflow |
| `.mind/conversation/agents/persona*.md` | Specialist personas: architect, pragmatist, critic, researcher, generic |
| `.mind/conversation/README.md` | Conversation module overview — phases, entry points, configuration |
| `.mind/skills/` | Deep-dive reference — load on demand when agent needs detailed guidance |
| `.mind/conventions/` | Universal rules — code quality, documentation, git, severity, shared patterns |
| `docs/knowledge/` | Domain reference — convergence analyses (`*-convergence.md`), spikes (`*-spike.md`), and decision context |
| `.mind/CHANGELOG.md` | Framework version history — date-based entries tracking framework changes |

## Commands

| Command | Entry Point | What It Does |
|---------|-------------|-------------|
| `/analyze "topic"` | `commands/analyze.md` | Structured multi-persona analysis |
| `/discover "idea"` | `commands/discover.md` | Explore before building |
| `/workflow "description"` | `commands/workflow.md` | Build with full agent pipeline |
| `/init` | `commands/init.md` | Guided setup — detect missing docs, create what's needed |

## Request Types → Agent Chains

| Type | Trigger | Chain |
|------|---------|-------|
| `NEW_PROJECT` | No existing codebase | analyst → architect → developer → tester → reviewer |
| `BUG_FIX` | Fix, bug, error, broken, crash, regression, failing | analyst → developer → tester → reviewer |
| `ENHANCEMENT` | Add, extend, improve, integrate, support, feature | analyst → [architect]* → developer → tester → reviewer |
| `REFACTOR` | Refactor, clean, restructure, optimize, simplify, modernize | analyst → developer → reviewer |
| `COMPLEX_NEW` | `analyze:`, `explore:`, evaluate options, trade-offs, compare approaches; or 3+ components with architectural uncertainty | conversation-moderator → analyst → architect → developer → tester → reviewer |

*Architect activates only when structural changes are needed.

## Quality Gates

| Gate | Owner | Criteria |
|------|-------|----------|
| **Gate 0** *(COMPLEX_NEW only)* | conversation-moderator | Convergence quality ≥ 3.0/5.0, Decision Matrix ≥ 3 options, ≥ 3 Recommendations with confidence, Quality Rubric ≥ 3.0/5.0 |
| **Micro-Gate A** | analyst | Acceptance criteria in testable format, scope boundary defined (in + out), success metrics quantified, requirements traceable (FR-N), Assumptions section required when business context gap |
| **Micro-Gate B** | developer | `changes.md` exists, files exist on disk, changes within scope, requirements tracked with FR-N |
| **Deterministic Gate** | pre-reviewer | Build, lint, typecheck, and test commands pass |

- **Max 2 retry loops** — then proceed with documented concerns
