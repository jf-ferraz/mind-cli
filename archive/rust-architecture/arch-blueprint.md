# Mind Framework — Unified Architectural Blueprint

> **Version**: 1.0  
> **Date**: 2026-02-25  
> **Status**: Canonical — authoritative reference for backlog, planning, and execution  
> **Scope**: Architecture, framework design, operational layer, integration model  
> **Excludes**: Implementation guides, sprint plans, code-level decisions, MVP task decomposition

---

## A. Executive Summary

The Mind Framework is a declarative, agent-orchestrated software development framework that applies NixOS-inspired principles to knowledge artifact management. It provides a canonical, systematic approach to engineering workflows — from requirements analysis through implementation, testing, and review — with strong emphasis on structured documentation, iterative execution, and governance.

### Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| **TOML manifest (`mind.toml`)** as single source of truth | Unambiguous typing, section-navigable, Cargo/pyproject precedent; eliminates YAML's implicit coercion risks in a governance file |
| **JSON lock file (`mind.lock`)** for computed state | Pre-computed index with upstream hash tracking enables reactive staleness detection — the core NixOS innovation |
| **Reactive dependency graph** with transitive staleness | Change one artifact; the system computes the minimum rebuild set across the entire document graph |
| **Hybrid incremental implementation path** | Python MVP validates design with zero friction; Rust CLI follows after real-project validation; no wasted work |
| **MCP as primary integration surface** | One MCP server replaces N platform-specific adapters; works with Claude Code, Codex CLI, Gemini CLI, and any future MCP-compatible agent |
| **7 core agents, adaptive routing** | Lean agent set covers all workflow needs; token cost is proportional to request complexity (BUG_FIX loads 4 agents, not 7) |
| **Conversation module for COMPLEX_NEW** | Dialectical multi-persona analysis before implementation reduces architectural uncertainty; integrated as 5th workflow type with convergence quality gate |
| **4-zone documentation model** | Separates stable specs, volatile state, immutable history, and reference material — each with distinct lifecycle and ownership rules |

### What This Blueprint Enables

- Deterministic, reproducible development workflows across any coding agent CLI
- Full requirement traceability: business need → specification → implementation → test → validation
- Reactive project state management: declare intent, detect drift, converge automatically
- Gradual adoption from zero-config (Level 0) to full reconciliation (Level 3)
- Agent-agnostic operation through stable CLI and MCP interfaces

---

## B. Scope and Objectives

### In Scope

- Target architecture for the Mind Framework v2
- Manifest system design (`mind.toml`, `mind.lock`, URI scheme, dependency graph)
- Agent orchestration model (7 agents, 4 workflow types, quality gates)
- Operational layer (`.mind/` runtime, CLI commands, performance architecture)
- Integration model (CLI, MCP server, hooks, plugins)
- Framework structure and file organization
- Iteration lifecycle and governance model
- Technology stack direction (Python MVP → Rust CLI)

### Out of Scope

- Implementation-level code or build instructions
- Backlog decomposition or sprint planning
- MVP task breakdown (Phase 4 of project-definition.md)
- Success measurement framework and KPIs
- Specific CI/CD pipeline configurations

### Architectural Goals

1. **Declarative project governance** — the manifest declares desired state; agents converge toward it
2. **Token-efficient agent orchestration** — agents load only relevant context; total framework footprint < 2,500 lines
3. **Zero-barrier adoption** — framework works without manifest (L0); manifest adds capabilities progressively
4. **Platform independence** — operates identically across Claude Code, Codex CLI, Gemini CLI, and future agents
5. **Evidence-based quality** — deterministic gates (build/lint/test) before review; no self-assessed scores

### Non-Goals

- Replacing existing agent CLIs or their native orchestration
- Building a web dashboard or cloud-hosted service (deferred beyond Phase 3)
- Enforcing a specific programming language or technology stack on target projects
- Providing autonomous operation without human oversight at quality gates

---

## C. Architectural Principles and Guidelines

### Design Principles

| # | Principle | Rationale |
|---|-----------|-----------|
| P1 | **Declarative over imperative** | The manifest describes WHAT should exist. Agents determine HOW. |
| P2 | **Reactive dependency tracking** | Change one artifact; staleness propagates transitively through the graph. |
| P3 | **Single source of truth** | If it is not in `mind.toml`, agents do not know about it. |
| P4 | **Data, not code** | The manifest is pure data (TOML). Computed state lives in the lock file. No runtime dependency. |
| P5 | **Layered adoption** | Works at Level 0 with no manifest. Each level adds optional capability. |
| P6 | **Agents read, orchestrator writes** | Only the orchestrator modifies `mind.toml`. Agents consume it for context. |
| P7 | **Git is version control** | The manifest tracks current state. Full history lives in `git log -- mind.toml`. |
| P8 | **Add governance, not complexity** | Every addition must justify its token cost in every session that loads it. |

### Operating Principles

| Principle | Meaning |
|-----------|---------|
| **Filesystem is the API** | No database, no daemon — plain files that bash, jq, and Python can query |
| **Agents read, CLI writes** | Agents consume operational state; CLI commands and hooks produce it |
| **Committed vs. ephemeral** | `mind.toml` and `mind.lock` are committed; `.mind/` contents are local and disposable |
| **Progressive cost** | Operations scale with project complexity — a 5-file project costs almost nothing |
| **CLI is the state engine, agents are the decision engine** | The `mind` binary computes state. Agents decide what to do with that state. |

### Decision Rules and Guardrails

1. **Agent line count**: If any agent exceeds 300 lines, split into agent + loadable supplement
2. **Manifest nesting**: Maximum 3 levels in TOML structure; dotted keys for leaf properties
3. **Token budget**: Framework overhead per agent session must stay under 4,000 tokens at Level 3
4. **Retry limit**: Maximum 2 retries per workflow, with targeted feedback to specific agents
5. **Convention hierarchy** (conflict resolution):
   - User instruction (explicit override)
   - Project docs (`docs/spec/`, `docs/knowledge/`, project `CLAUDE.md`)
   - Codebase patterns (existing code conventions)
   - Framework conventions (`conventions/*.md`)
   - Best practices (community standards)

---

## D. Final Architecture Blueprint (Core Model)

### D.1 High-Level Architecture

The system operates on a manifest-lock-reconcile loop inspired by NixOS:

```
mind.toml (declared state) + mind.lock (actual state) → delta → agent dispatch → artifact production → lock regeneration
```

**Three primary layers**:

| Layer | Components | Responsibility |
|-------|-----------|----------------|
| **Governance Layer** | `mind.toml`, `[[graph]]`, `[governance]`, `[[generations]]` | Declares desired project state, policies, and relationships |
| **Reconciliation Layer** | `mind.lock`, delta computation, agent dispatch plan | Computes actual state, detects drift, determines minimum rebuild set |
| **Execution Layer** | Agent pipeline, quality gates, CLI/MCP tools, hooks | Produces artifacts that converge actual state toward declared state |

### D.2 Manifest System

**`mind.toml`** — Declarative manifest. Human-authored (project identity, stack) and orchestrator-managed (iterations, generations). Committed to git.

| Section | Purpose | Maintained By |
|---------|---------|---------------|
| `[manifest]` | Schema version, generation counter | Orchestrator |
| `[project]` | Identity, stack, commands | Human |
| `[profiles]` | Active profile bundles | Human |
| `[agents.*]` | Agent registry (path, role, loads, produces) | Framework |
| `[workflows.*]` | Agent chains per request type, gate placement | Framework |
| `[documents.*]` | Artifact registry with URIs, paths, owners, dependencies | Orchestrator |
| `[[graph]]` | Typed dependency edges between artifacts | Orchestrator |
| `[governance]` | Review policy, commit policy, gates, ADRs | Orchestrator + Human |
| `[operations]` | Runtime configuration: commands, capture, retention, infrastructure | Human |
| `[manifest.invariants]` | Self-validation rules (no orphan deps, no cycles, etc.) | Framework |
| `[[generations]]` | Strategic state history (monotonic counter) | Orchestrator |

**`mind.lock`** — Computed state snapshot. Auto-generated JSON. Committed to git. Never hand-edited.

| Field | Purpose |
|-------|---------|
| `resolved` | Per-artifact state: path, exists, hash, size, mtime, stale flag, upstream hashes |
| `warnings` | Missing artifacts, stale chains, orphan references |
| `completeness` | Requirement implementation percentages, iteration progress |
| `operations` | Last command results, infrastructure health, cache state |

### D.3 Canonical URI Scheme

Every artifact gets a stable, path-independent identifier.

| Prefix | Scope | Example |
|--------|-------|---------|
| `doc:spec/` | Stable specifications | `doc:spec/requirements`, `doc:spec/domain-model` |
| `doc:state/` | Volatile runtime state | `doc:state/current`, `doc:state/workflow` |
| `doc:iteration/` | Immutable history | `doc:iteration/003` |
| `doc:knowledge/` | Domain reference | `doc:knowledge/glossary` |
| `agent:` | Agent definitions | `agent:analyst`, `agent:architect` |
| `specialist:` | Specialist agents | `specialist:database` |
| `gate:` | Quality gates | `gate:micro-a`, `gate:deterministic` |
| `workflow:` | Workflow definitions | `workflow:enhancement` |

**Fragment addressing**: `doc:spec/requirements#FR-3` — enables full requirement traceability.

**`@`-shorthand in prose**: `@spec/requirements#FR-3` resolves to `doc:spec/requirements#FR-3` via the manifest registry. Survives file renames (only the manifest's `path` field changes).

### D.4 Reactive Dependency Graph

Explicit typed edges between artifacts, stored as `[[graph]]` entries in the manifest.

| Edge Type | Direction | Staleness Propagation |
|-----------|-----------|----------------------|
| `derives-from` | downstream → upstream | Yes — upstream change makes downstream stale |
| `implements` | iteration → requirement | Advisory — flagged but not auto-stale |
| `validates` | test → requirement | Advisory — flagged for review |
| `supersedes` | new → old | Old marked as superseded |
| `informs` | knowledge → spec | None (advisory only) |

**Staleness propagation**: Each resolved artifact records the hashes of its dependencies at the time it was last updated. When a dependency's hash changes, everything downstream becomes transitively stale.

### D.5 Reconciliation Engine

The reconciliation engine is the orchestrator's core behavior:

1. Read `mind.toml` (declared state)
2. Read `mind.lock` (actual state)
3. Compute delta: declared − actual = work
4. Identify missing and stale artifacts
5. Compute minimum agent dispatch plan in dependency order
6. Execute plan, updating lock after each agent completes

### D.6 Agent Orchestration

**7 core agents** with adaptive routing:

| Agent | Role | Loads | Produces |
|-------|------|-------|----------|
| **Orchestrator** | Dispatch, git, gates, reconciliation | Always | `doc:state/workflow`, iteration registration |
| **Analyst** | Requirements, domain model extraction | On-demand | `doc:spec/requirements`, `doc:spec/domain-model` |
| **Architect** | System design, API contracts | Conditional | `doc:spec/architecture`, `doc:spec/api-contracts` |
| **Developer** | Implementation, commit discipline | On-demand | `doc:iteration/changes` |
| **Tester** | Test strategy, domain model derivation | On-demand | — |
| **Reviewer** | Evidence-based review, deterministic gates | On-demand | `doc:iteration/validation` |
| **Discovery** | Product exploration, stakeholder mapping | On-demand | `doc:spec/project-brief` |
| **Conversation Moderator** | Dialectical analysis, convergence synthesis | Conditional | `analysis/conversation/*.md` |

> The conversation moderator and its specialist personas live in `.github/agents/` (Copilot Chat convention). They form the conversation module — invoked for `COMPLEX_NEW` requests or standalone via `/analyze`.

**5 workflow types**:

| Type | Trigger | Chain |
|------|---------|-------|
| `NEW_PROJECT` | No existing codebase | analyst → [specialist] → architect → developer → tester → reviewer |
| `BUG_FIX` | Fix, bug, error, broken | analyst → developer → tester → reviewer |
| `ENHANCEMENT` | Add, extend, improve | analyst → [architect if structural] → developer → tester → reviewer |
| `REFACTOR` | Refactor, clean, restructure | analyst → developer → reviewer |
| `COMPLEX_NEW` | Architectural uncertainty, 3+ components | conversation-moderator → analyst → architect → developer → tester → reviewer |

**Specialist injection**: Zero specialists ship by default. Project-specific specialists activate via keyword triggers and insert after a configured agent position.

### D.7 Quality Gate Architecture

| Gate | Type | When | Checks | On Failure |
|------|------|------|--------|------------|
| **Micro-Gate A** | Probabilistic | After analyst | Acceptance criteria present, scope defined, no ambiguous terms | Retry analyst |
| **Micro-Gate B** | Probabilistic | After developer | `changes.md` exists, files exist, scope adherence | Retry developer |
| **Deterministic** | Deterministic | Before reviewer | Build, lint, typecheck, test (from `[project.commands]`) | Return to developer |
| **Reviewer** | Evidence-based | After reviewer | MUST/SHOULD/COULD via git diff + test results + traceability | Max 2 retries, then proceed with documented concerns |

### D.8 Operational Layer

**`.mind/` runtime directory** — local, `.gitignore`d, disposable:

| Directory | Purpose | Retention |
|-----------|---------|-----------|
| `cache/summaries/` | Pre-computed document summaries (~200 tokens each) | Rebuilt on demand |
| `cache/hashes.json` | File hash cache for incremental lock sync | Rebuilt on demand |
| `logs/runs/` | Per-workflow JSONL event streams | Last 20 runs |
| `logs/gates/` | Structured gate result snapshots | Active + last completed iteration |
| `logs/audit.jsonl` | Append-only audit trail | Rotated at threshold |
| `outputs/{type}/` | Captured build/test/lint outputs with `latest` symlink | Last 5 per type |
| `tmp/` | Agent scratch (PLAN.md, WIP.md, LEARNINGS.md) | Deleted on workflow completion |
| `hooks/` | Generated git hook scripts | Regenerated on `mind init` |

### D.9 CLI Command Interface

| Command | Purpose | Performance Target |
|---------|---------|:--:|
| `mind init` | Scaffold `.mind/`, install hooks, generate first lock | < 0.5s |
| `mind lock` | Sync lock file (mtime fast path + SHA-256) | < 0.5s (50 artifacts) |
| `mind status` | Project state dashboard (human or JSON) | < 0.1s |
| `mind query` | Artifact lookup by URI or term | < 0.1s |
| `mind validate` | Manifest invariant checks | < 0.2s |
| `mind graph` | Dependency visualization (text, Mermaid, JSON) | < 0.1s |
| `mind gate` | Deterministic gate runner with structured capture | Dominated by gate commands |
| `mind clean` | Archive iterations, rotate logs, prune outputs | < 0.5s |
| `mind summarize` | Generate/regenerate document summaries | 1-5s |

**Output protocol**: stdout for content, stderr for diagnostics. Exit codes: 0 (success), 1 (failure), 2 (invalid args), 3 (manifest error), 4 (lock out of sync). All commands support `--json` for machine-readable output.

---

## E. Framework Structure

### E.1 Framework Repository Layout

```
mind-framework/
├── CLAUDE.md                       # Framework index (~200 tokens, pure routing)
├── MIND-FRAMEWORK.md               # Canonical specification
├── README.md                       # Framework documentation
├── install.sh                      # Install framework into project's .claude/
├── scaffold.sh                     # Bootstrap project structure + mind.toml
│
├── agents/                         # Core agents (7)
│   ├── orchestrator.md
│   ├── analyst.md
│   ├── architect.md
│   ├── developer.md
│   ├── tester.md
│   ├── reviewer.md
│   └── discovery.md
│
├── conventions/                    # Universal rules (7)
│   ├── CLAUDE.md
│   ├── shared.md                   # Cross-workflow patterns (evidence, reasoning, scope)
│   ├── code-quality.md
│   ├── documentation.md            # 4-zone model
│   ├── git-discipline.md           # Commit protocol, branch strategy, PR flow
│   ├── severity.md                 # MUST/SHOULD/COULD + intent markers
│   ├── temporal.md                 # Temporal contamination heuristic
│   └── backend-patterns.md         # Optional (activated by profile)
│
├── skills/                         # On-demand deep dives (4)
│   ├── CLAUDE.md
│   ├── planning/SKILL.md
│   ├── debugging/SKILL.md
│   ├── refactoring/SKILL.md
│   └── quality-review/SKILL.md
│
├── commands/                       # User entry points (3)
│   ├── analyze.md                  # /analyze — conversation analysis (Mode A/B)
│   ├── discover.md
│   └── workflow.md
│
├── specialists/                    # Optional domain specialists
│   ├── _contract.md
│   └── examples/
│       └── database-specialist.md
│
├── templates/                      # Reusable document templates
│   ├── domain-model.md
│   ├── api-contract.md
│   ├── iteration-overview.md
│   └── retrospective.md
│
├── bin/                            # CLI dispatcher
│   └── mind
│
├── conversation/                   # Conversation module (dialectical analysis)
│   ├── config/                     # conversation.yml, extensions.yml, personas.yml, quality.yml
│   ├── protocols/                  # Phase routing, evaluator-optimizer, approval gates, delegation
│   └── skills/                     # Evidence standards, reasoning chains, challenge methodology, etc.
│
├── .github/                        # Copilot Chat agents and prompts
│   ├── agents/                     # Conversation moderator + 4 specialist personas + generic persona
│   └── prompts/                    # analyze.prompt.md, analyze-documents.prompt.md
│
├── analysis/                       # Convergence output directory
│   └── conversation/
│
└── lib/                            # Shared libraries
    ├── mind_lock.py
    ├── mind_validate.py
    ├── mind_graph.py
    ├── mind_summarize.py
    └── common.sh
```

### E.2 Target Project Layout

```
project-root/
├── mind.toml                       # Declarative manifest
├── mind.lock                       # Computed state (auto-generated)
├── CLAUDE.md                       # Project routing table
├── .claude/                        # Installed framework
│   ├── agents/, conventions/, skills/, commands/
│   ├── specialists/, templates/
│   ├── bin/mind                    # CLI dispatcher
│   └── lib/                       # Python modules + Bash lib
│
├── docs/
│   ├── spec/                       # Zone 1: Stable specifications
│   │   ├── project-brief.md, requirements.md, domain-model.md
│   │   ├── architecture.md, api-contracts.md
│   │   └── decisions/              # ADRs
│   ├── state/                      # Zone 2: Volatile runtime state
│   │   ├── current.md, workflow.md
│   │   └── gate-results/           # Structured gate output capture
│   ├── iterations/                 # Zone 3: Immutable history (append-only)
│   │   └── {NNN}-{type}-{descriptor}/
│   └── knowledge/                  # Zone 4: Domain reference
│       ├── glossary.md
│       └── integrations.md
│
├── .mind/                          # Runtime state (.gitignored)
│   ├── cache/, logs/, outputs/, tmp/, hooks/
│
└── src/                            # Application source code
```

### E.3 Zone Architecture

| Zone | Mutability | Who Writes | Staleness Tracked |
|------|-----------|------------|:-:|
| **spec/** | Stable — updated after analysis | Analyst, Architect, Discovery | Yes |
| **state/** | Volatile — changes every session | Orchestrator | No (always current) |
| **iterations/** | Append-only — immutable once complete | All agents (within their iteration) | No (historical) |
| **knowledge/** | Stable — evolves with business understanding | Discovery, Analyst | Yes |

### E.4 Evolution Strategy

The framework evolves through **profiles** and **adoption levels**:

**Profiles** (NixOS module-like bundles):

| Profile | Activates | Use When |
|---------|-----------|----------|
| `backend-api` | backend-patterns convention, domain-model + api-contract templates, database specialist availability | Backend or fullstack projects |
| `event-driven` | Messaging patterns guidance, async awareness | Event-sourced projects |
| `minimal` | Core agents + conventions only | Small scripts, utilities |

**Adoption Levels**:

| Level | Capability Added | Tooling Required |
|:---:|-----------------|:-:|
| **L0** | Framework works as v1, no manifest | None |
| **L1** | Manifest with project + documents, human-maintained | None |
| **L2** | Agents consult manifest, registry-based dispatch, dependency graph | None |
| **L3** | Lock file with staleness detection, reconciliation engine, completeness metrics | `mind lock` script |

---

## F. Data Flows and Dependencies

### F.1 End-to-End Request Flow

```
User request
  → Orchestrator reads mind.toml + mind.lock
  → Computes delta (declared − actual)
  → Classifies request (NEW_PROJECT / BUG_FIX / ENHANCEMENT / REFACTOR)
  → Selects agent chain from [workflows.*]
  → Scans specialists for trigger matches
  → Initializes iteration (folder, branch, manifest registration, generation bump)
  → Dispatches agents sequentially with gates between phases
  → After each agent: lock regeneration captures new artifact state
  → Deterministic gates before reviewer (build/lint/typecheck/test)
  → Reviewer produces evidence-based validation
  → Completion: iteration status → complete, generation bump, final lock, PR summary
```

### F.2 Agent Context Loading Flow

Each agent reads only relevant manifest sections:

| Agent | Reads from mind.toml | Reads from mind.lock |
|-------|---------------------|---------------------|
| Orchestrator | `[project]`, `[workflows]`, `[agents]`, `[governance]`, `[[generations]]` | Full: staleness, completeness, warnings |
| Analyst | `[documents.spec.*]`, `[[graph]]` | Staleness of spec documents |
| Architect | `[documents.spec.*]`, `[project.stack]` | Staleness of architecture |
| Developer | `[project.commands]`, active iteration | — |
| Tester | Active iteration, `[documents.spec.domain-model]` | — |
| Reviewer | `[governance]`, active iteration | Staleness, completeness |

### F.3 Lock Sync Data Flow (Core Hot Path)

```
mind lock
  → Parse mind.toml → Manifest struct
  → Read previous mind.lock
  → For each declared artifact:
      → stat(path): exists? mtime? size?
      → If mtime + size unchanged → skip hash (fast path)
      → If changed → compute SHA-256 → compare against lock hash
      → If hash differs → mark CHANGED, propagate staleness downstream
  → Compute completeness metrics
  → Generate warnings
  → Write mind.lock (atomic: .tmp → rename)
  → Update .mind/cache/hashes.json
```

### F.4 Dependency and Coupling Considerations

| Dependency | Type | Risk | Mitigation |
|-----------|------|------|------------|
| Python 3.11+ (MVP) | Runtime | Moderate — not in minimal containers | Fallback to Python 3.7 + tomli; Rust CLI eliminates this |
| Git | Runtime | Low — universal in dev contexts | Graceful degradation: skip git ops if unavailable |
| Agent CLI (Claude Code, etc.) | Platform | Medium — CLIs evolve independently | MCP as stable integration surface; thin platform shims |
| TOML format | Design | Low — stable, widely adopted | Format locked at schema `mind/v2.0` |

---

## G. Iteration Lifecycle

### G.1 Iteration State Machine

```
[Created] → [Planning] → [InProgress] → [InReview] → [Complete]
                                ↑              ↓
                                └── NEEDS_REVISION (max 2)

[InProgress] → [Interrupted] → [InProgress]  (session boundary)
```

### G.2 Lifecycle Phases

| Phase | Inputs | Activities | Outputs | Checkpoint |
|-------|--------|-----------|---------|------------|
| **Planning** | User request, manifest context, lock state | Classification, chain selection, specialist scan, initialization | Iteration folder, branch, manifest registration | Generation bump |
| **Analysis** | Project brief, existing specs, domain context | Requirements extraction, domain model, acceptance criteria | requirements-delta.md, domain-model.md | Gate A |
| **Design** (conditional) | Requirements, domain model | Architecture decisions, API contracts | architecture.md, api-contracts.md | Session split (if needed) |
| **Implementation** | All specs, architecture | Code changes, commit at logical units | changes.md, source code | Gate B |
| **Verification** | Domain model, changes | Test derivation, coverage verification | Tests, test results | Deterministic gate |
| **Review** | All artifacts, git diff, test evidence | Evidence-based assessment, traceability | validation.md | Reviewer verdict |
| **Completion** | Reviewer approval | Status update, lock regen, PR summary | Final commit, generation bump | Workflow end |

### G.3 Session Handoff Protocol

When a workflow splits across sessions:

1. Orchestrator writes structured handoff to `docs/state/workflow.md` containing:
   - Position (type, descriptor, last agent, remaining chain)
   - Completed artifact locations
   - Key decisions made in this session
   - Manifest delta at session end
   - Context guidance for next session
2. Commit with message `wip: planning complete for {descriptor}`
3. On resume: orchestrator reads handoff → selectively loads referenced artifacts → skips completed agents

### G.4 Feedback Loop

Iteration outputs feed back into the framework:

- **Retrospectives** (optional per iteration) capture process improvements
- **ADRs** register decisions in the manifest governance section
- **Generation history** provides strategic changelog within the manifest
- **Gate results** (archived as `gate-summary.json` per iteration) enable quality trending

---

## H. Governance and Decision Model

### H.1 Ownership Boundaries

| Artifact | Owner | Write Authority |
|----------|-------|----------------|
| `mind.toml` — `[project]`, `[profiles]` | Human (project lead) | Direct edit |
| `mind.toml` — `[documents]`, `[[graph]]`, `[[generations]]` | Orchestrator | Automated during workflow |
| `mind.toml` — `[governance.decisions]` | Orchestrator + Human | ADRs proposed by agents, accepted by human |
| `mind.lock` | CLI tool (`mind lock`) | Fully automated, never hand-edited |
| Agent definitions | Framework maintainer | Updated via `install.sh --update` |
| Specialist definitions | Project team | Project-specific creation |
| Spec documents (Zone 1) | Assigned agent (per `owner` field) | During workflow execution |
| State documents (Zone 2) | Orchestrator | Overwritten per session |

### H.2 Decision-Making Points

| Decision | Who Decides | When | Record |
|----------|------------|------|--------|
| Request classification | Orchestrator | Workflow start | Iteration overview.md |
| Architect inclusion (ENHANCEMENT) | Orchestrator | After analyst, if structural change detected | Workflow state |
| Specialist injection | Orchestrator | After classification, via trigger matching | Iteration overview.md |
| Session split | Orchestrator | After architect (NEW_PROJECT) or configurable | `session-split-after` in workflow def |
| Gate pass/fail | Gate runner (deterministic) or Reviewer (probabilistic) | At configured gate positions | `.mind/logs/gates/` |
| Retry vs. proceed | Orchestrator | On gate failure, respecting max 2 retries | Documented in validation.md |
| Architecture decisions (ADRs) | Architect agent + Human approval | During design phase | `docs/spec/decisions/` |

### H.3 Architectural Change Control

- **Schema changes** to `mind.toml`: Require schema version bump (`manifest.schema`); existing projects continue working at prior version
- **Agent behavior changes**: Updated via `install.sh --update`; project-specific `CLAUDE.md` overrides preserved
- **New agent addition**: Rejected by design — 7 agents cover all needs; additional agents add permanent token cost
- **Profile changes**: New profiles can be added without breaking existing projects
- **Convention changes**: Updated centrally; all agents inherit automatically

---

## I. Risks, Gaps, and Validation Needs

### I.1 Confirmed Decisions

| Decision | Status | Source |
|----------|--------|--------|
| TOML format for manifest | Confirmed | Cross-document consensus |
| `doc:` URI scheme with `@` shorthand | Confirmed | Synthesis of Proposal A (shorthand) + Proposal B (formal URIs) |
| Lock file committed to git | Confirmed | All proposals agree |
| 7 core agents (unchanged count) | Confirmed | Analysis + improvement documents |
| 4-zone documentation model | Confirmed | All proposals agree |
| Evidence-based review (no numerical scores) | Confirmed | Preserved from v1; both benchmarks' scores rejected |
| Hybrid incremental implementation path | Confirmed | Reconciles early CLI rejection with operational complexity |
| Layered adoption (L0-L3) | Confirmed | Cross-document consensus |
| Conversation module as COMPLEX_NEW workflow | Confirmed | Panel-module integration — dialectical analysis for architecturally uncertain requests |

### I.2 Inferred Decisions (Require Validation)

| Decision | Basis | Validation Needed |
|----------|-------|-------------------|
| Python 3.11+ adequate for MVP performance | Operational-layer analysis | Run `mind lock` on 50+ artifact project; verify < 500ms |
| mtime fast path eliminates 95% of hash computations | Theoretical analysis | Measure in real workflows across multiple sessions |
| Context budgeting reduces token usage by 40%+ | Token estimation model | Compare agent token consumption v1 vs v2 on identical projects |
| MCP server can replace direct CLI invocation for agents | MCP protocol capability analysis | Test with all three target agent CLIs |
| `[operations]` section does not bloat manifest for simple projects | Design intent | Verify with Level 1 projects that never set operations |

### I.3 Open Questions

| Item | Why Unresolved | Impact | Validation Approach |
|------|---------------|--------|-------------------|
| Optimal `context-budget` defaults per agent | Requires empirical token measurement across project sizes | Medium — affects agent effectiveness | Run full workflows, measure per-agent token consumption |
| WASM plugin API stability | Plugin system is Phase 3+; API depends on real extension needs | Low — hooks cover 80% of extensibility | Defer until Phase 3; start with hook-only extensibility |
| Parallel agent dispatch feasibility | Dependency graph allows it theoretically; LLM session model may not | Medium — affects large project throughput | Test with agent CLIs that support concurrent tool calls |
| `mind.toml` growth limits for large monorepos | 100+ artifact registrations may exceed manageable size | Medium — affects scalability ceiling | Generate synthetic large manifest; measure parse time and readability |
| Rust vs Go as long-term CLI language | Both viable; Rust preferred but contributor pool is smaller | Medium — affects development velocity | Assess contributor availability after Phase 2 validation |

### I.4 Architectural Risks

| # | Risk | Probability | Impact | Mitigation |
|---|------|:-:|:-:|-----------|
| R-01 | Manifest becomes maintenance burden | Medium | Medium | Orchestrator maintains most sections automatically; human edits only `[project]` + `[profiles]` |
| R-02 | Agent CLIs build native manifest support, commoditizing the framework | Medium | High | Framework's value is the design (reactive graph, reconciliation), not just tooling; pivot to spec + reference implementation |
| R-03 | MCP protocol changes break integration | Low | Medium | Pin to MCP spec version; adapter layer isolates changes |
| R-04 | Domain model becomes over-engineered ritual | Medium | Medium | Template is optional; created only for NEW_PROJECT and structural ENHANCEMENT |
| R-05 | Python MVP performance insufficient for large projects | Medium | Low | Lock sync may exceed 500ms at 100+ artifacts; Rust CLI resolves this in Phase 3 |
| R-06 | Lock file diverges between team members | Low | Medium | Pre-commit hook runs `mind lock --verify`; CI validates lock freshness |
| R-07 | Rust learning curve limits contributors | Medium | High | Phase 1 in Python validates design first; Go identified as fallback |

---

## J. Final Blueprint Summary

### J.1 Architecture at a Glance

The Mind Framework v2 is a **declarative, agent-orchestrated development framework** built on three pillars:

1. **Governance Pillar**: `mind.toml` declares the desired project state — documents, dependencies, workflows, policies. The manifest is the single source of truth.

2. **Reconciliation Pillar**: `mind.lock` captures actual filesystem state. The orchestrator computes the delta and dispatches the minimum agent set to converge. This is `nixos-rebuild` for knowledge artifacts.

3. **Execution Pillar**: 7 agents execute in adaptive chains (4 workflow types), gated by deterministic and evidence-based quality checks. A CLI/MCP interface provides fast state queries and gate execution.

### J.2 Prioritized Focus Areas for Next Phase

| Priority | Area | Rationale |
|:--------:|------|-----------|
| **P0** | Agent prompt updates (manifest awareness, domain model, micro-gates, git integration) | Core value — agents become manifest-aware |
| **P0** | 4-zone documentation structure + documentation convention update | Foundation for all artifact management |
| **P0** | `mind.toml` schema finalization + scaffold integration | Enables Level 1 adoption |
| **P1** | `mind lock` (Python MVP, mtime fast path) | Enables Level 3 reconciliation |
| **P1** | `mind gate` (structured gate runner with output capture) | Enables deterministic quality gates |
| **P1** | `mind status` (project state dashboard) | Primary user-facing command |
| **P2** | Specialist mechanism + contract | Extensibility for domain-specific needs |
| **P2** | Git discipline convention update (commit protocol, branch strategy) | Formalizes git workflow |
| **P2** | Structured handoff in `workflow.md` | Enables reliable session resume |
| **P3** | Rust CLI rewrite | Performance and distribution improvements |
| **P3** | MCP server implementation | Agent-agnostic integration surface |
| **P4** | WASM plugin system | Advanced extensibility (deferred) |

---

## Appendix: Supporting Tables

### Table 1 — Consensus/Divergence Matrix

| Topic | Alignment | Final Decision | Notes |
|-------|:-:|------------|-------|
| Manifest format | Full | TOML | All proposals converge; YAML explicitly rejected |
| Lock file role | Full | Core (committed, auto-generated) | Early proposals had lock as optional; now core |
| URI scheme | Full | `doc:{zone}/{name}` + `@` shorthand in prose | Synthesis of formal (Proposal B) + ergonomic (Proposal A) |
| Agent count | Full | 7 (unchanged from v1) | Documenter/deployer/researcher explicitly rejected |
| Adoption levels | Full | L0-L3 progressive | All documents support layered adoption |
| Quality gates | Full | Deterministic before reviewer + micro-gates | Numerical scores rejected; evidence-based only |
| CLI tooling | **Resolved divergence** | Hybrid incremental (Python MVP → Rust) | Canonical spec rejected CLI; operational layer required it; hybrid reconciles both |
| `.mind/` directory | **Near-consensus** | Adopted for runtime state (`.gitignored`) | Operational-layer + implementation docs agree; canonical spec implied but didn't formalize |
| MCP integration | **Near-consensus** | Primary integration surface (Phase 3+) | Implementation docs agree; canonical spec mentions as future |
| WASM plugins | **Partial** | Deferred to Phase 4; hooks provide 80% value | One implementation doc proposes; others don't address |
| `[operations]` section | **Partial** | Adopted | Operational-layer proposes; canonical design doesn't include; unified here |

### Table 2 — Component & Responsibility Matrix

| Component | Responsibility | Inputs | Outputs | Key Dependencies |
|-----------|---------------|--------|---------|-----------------|
| `mind.toml` | Declare desired project state | Human edits + orchestrator updates | Configuration for all agents and tools | Git (versioning) |
| `mind.lock` | Capture actual filesystem state | `mind.toml` + filesystem scan | JSON snapshot with staleness, completeness | `mind.toml`, filesystem |
| Orchestrator | Request classification, dispatch, git workflow | User request, manifest, lock | Iteration creation, agent dispatch, PR summary | All agents, gates, git |
| Analyst | Requirements extraction, domain modeling | Project brief, existing specs | requirements.md, domain-model.md | Project brief |
| Architect | System design, API contracts | Requirements, domain model | architecture.md, api-contracts.md | Requirements, domain model |
| Developer | Code implementation, commit discipline | All specs, architecture | Source code, changes.md | Architecture, API contracts |
| Tester | Test strategy, coverage verification | Domain model, changes | Test code, test results | Domain model |
| Reviewer | Evidence-based quality validation | All artifacts, git diff, test results | validation.md | Gate results, git |
| `mind` CLI | State computation, gate execution | `mind.toml`, filesystem | `mind.lock`, gate results, status | Python 3.11+ (MVP) / Rust (final) |
| MCP Server | Agent-agnostic integration interface | MCP tool calls from agent CLIs | Structured JSON responses | Core engine |

### Table 3 — Data Flow & Dependency Matrix

| Flow | Source | Target | Purpose | Risk |
|------|--------|--------|---------|------|
| Manifest → Lock | `mind.toml` | `mind lock` process | Declare what should exist | Parse errors halt workflow |
| Filesystem → Lock | File system (stat + hash) | `mind.lock` | Capture what does exist | Large projects may slow sync |
| Lock → Orchestrator | `mind.lock` | Orchestrator agent | Determine stale/missing artifacts | Stale lock = incorrect dispatch |
| Orchestrator → Agent | Dispatch context | Each agent | Provide relevant context for task | Over-loading wastes tokens |
| Agent → Filesystem | Agent output | `docs/`, `src/` | Produce artifacts | File conflicts possible |
| Agent → Lock | Produced artifacts | Next `mind lock` sync | Update computed state | Requires re-sync after each agent |
| Gate → Developer | Gate failure | Developer retry | Fix build/lint/test errors | Max 2 retries, then proceed |
| Git → Orchestrator | Branch state, diff | Orchestrator decisions | Inform workflow operations | Detached HEAD in CI |

### Table 4 — Open Questions & Validation Table

| Item | Why Unresolved | Impact | Validation Needed |
|------|---------------|--------|-------------------|
| Context budget defaults | Empirical data required | Medium | Measure token consumption across 3+ real projects |
| Lock sync performance at scale | 100+ artifacts untested | Medium | Benchmark with synthetic large project |
| MCP tool completeness | Integration needs evolve with usage | Low | Test with real agent workflows |
| Session resume reliability | Handoff document format may need iteration | Medium | Test split workflows end-to-end |
| Profile activation mechanics | Convention/template loading path undefined at runtime | Low | Implement and validate with `backend-api` profile |
| Manifest growth management | No pruning strategy for `[[generations]]` | Low | Define max inline generations (suggest 10) |

---

*This document consolidates and supersedes all individual proposal documents in `documents/architecture/` for the purpose of architectural direction. Individual documents remain valid as detailed reference for their specific domains. This blueprint is the primary guideline for backlog creation, task breakdown, and implementation planning.*
