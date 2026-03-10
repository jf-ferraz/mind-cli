# Documentation

Standards for project documentation within the agent framework.

## Hierarchy

Every word must earn its tokens. Documentation has two tiers:

**CLAUDE.md — Pure Index (~200 tokens max)**
- Tabular format only: `| Resource | When to Read |`
- No explanatory prose, no examples, no rationale
- Auto-loaded in every context — keep it minimal
- Points to where detailed information lives
- Content test: "Is this a pointer or an explanation?" Only pointers belong here.

**README.md — Invisible Knowledge**
- Architecture decisions, design rationale, invariants
- Loaded on demand — can be longer
- Content test: "Could a developer learn this by reading the source files?" If yes, don't document it here.
- Target audience: a competent developer encountering this project for the first time

## 5-Zone Documentation Structure

All project documentation lives under `docs/` in five semantic zones:

```
docs/
├── spec/                          # Zone 1: Stable specifications (versioned intent)
│   ├── project-brief.md         
#   Vision, deliverables, scope (filled by /discover)
│   ├── requirements.md            #   Living document — what the system must do
│   ├── domain-model.md            #   Entities, business rules, state machines, constraints
│   ├── architecture.md            #   Living document — how the system is structured
│   ├── api-contracts.md           #   External interfaces (if applicable)
│   └── decisions/                 #   Architecture Decision Records
│       └── _template.md
│
├── blueprints/                    # Zone 2: System-level planning artifacts
│   ├── INDEX.md                   #   Summary index — read this first, load docs on demand
│   └── {NN}-{descriptor}.md       #   Architecture blueprints, implementation plans, operational specs
│
├── state/                         # Zone 3: Volatile runtime state
│   ├── current.md                 #   Active state — tasks, issues, priorities
│   └── workflow.md                #   Workflow state for session management
│
├── iterations/                    # Zone 4: Append-only history
│   └── {NNN}-{type}-{descriptor}/ #   Per-change tracking
│       ├── overview.md
│       ├── changes.md
│       ├── test-summary.md
│       ├── validation.md
│       └── retrospective.md
│
└── knowledge/                     # Zone 5: Domain reference
    ├── glossary.md                #   Domain terminology
    ├── {topic}-spike.md           #   Spike findings (suffixed -spike.md)
    ├── {descriptor}-convergence.md#   Convergence analysis outputs from /analyze
    └── {topic}.md                 #   Other reference material
```

### Zone Semantics

| Zone | Content Type | Mutability | Owner |
|------|-------------|------------|-------|
| `spec/` | Specifications and design intent | Updated incrementally (append/modify, never regenerate) | Analyst, Architect, Discovery |
| `blueprints/` | System-level planning and reasoning | Stable — updated only at major phase transitions | Architect, Analyst |
| `state/` | Runtime context and active work | Overwritten freely — current truth only | Orchestrator |
| `iterations/` | Per-change tracking and history | Append-only — never modify completed iterations | All agents (within their iteration) |
| `knowledge/` | Domain reference material | Stable — updated when domain understanding changes | Discovery, Analyst |

### Blueprints vs. Spec

- **`spec/`** = "What the system **is** right now" — living documents, always current
- **`blueprints/`** = "What we **planned** to build and why" — stable, historical intent

When the system diverges from a blueprint, `spec/` is the source of truth. Blueprints provide context for understanding *why* decisions were made.

**Token efficiency**: Blueprints are large (20-50K each). Agents read `INDEX.md` first, then load only the specific blueprint needed. The orchestrator loads blueprints only at sprint initiation and re-planning, not during regular `BUG_FIX` or `ENHANCEMENT` workflows.

### Canonical URI Scheme

Every artifact can be referenced using the `doc:{zone}/{name}` URI scheme:

- `doc:spec/requirements` → `docs/spec/requirements.md`
- `doc:spec/requirements#FR-3` → specific section
- `doc:blueprints/04` → `docs/blueprints/04-operational-reference.md`
- `doc:iteration/003` → `docs/iterations/003-enhancement-dashboard/`
- `doc:knowledge/glossary` → `docs/knowledge/glossary.md`

In prose, use `@` shorthand: `@spec/requirements#FR-3`. This is readable in any markdown paragraph and survives file moves.

### Project Brief

`docs/spec/project-brief.md` is the upstream input for requirements:
- **Created by**: `/discover` command (interactive) or manually
- **Consumed by**: analyst (as primary context for requirements)
- **Updated**: rarely — only if the project's vision fundamentally shifts
- **Not a living document** — it's a snapshot of original intent

### Minimum Viable Business Context

The project brief must contain substantive content (not just template headings) for
the orchestrator's business context gate to pass:

| Section | Why It's Required |
|---------|-------------------|
| Vision | Without it, scope is undefined — analyst infers the system's purpose |
| Key Deliverables | Without it, "done" is undefined — scope creeps inevitably |
| Scope (In/Out) | Without it, every agent makes independent boundary assumptions |

A brief with only headings and HTML comments is classified as a **stub** and treated
as absent by the orchestrator and analyst.

For existing projects being onboarded: run `/discover` or manually fill the brief
before running `/workflow` for the first time. Subsequent BUG_FIX and REFACTOR
workflows do not require a brief.

### Domain Model

`docs/spec/domain-model.md` captures the project's domain knowledge:
- **Created by**: analyst (during NEW_PROJECT)
- **Consumed by**: architect, developer, tester, reviewer
- **Contains**: entities, business rules, state machines, constraints
- **Updated**: when enhancements introduce new domain concepts

### Invisible Knowledge

Knowledge NOT deducible from reading the code alone. Captured in `README.md` files **in the same directory as the affected code** (code-adjacent, not in a separate docs folder).

Categories:
- **Architecture decisions**: component relationships, data flow, module boundaries
- **Business rules**: domain constraints that shape implementation
- **Invariants**: properties that must hold but aren't enforced by types/compiler
- **Tradeoffs**: costs and benefits of chosen approaches
- **Performance characteristics**: non-obvious efficiency properties

**Self-contained principle**: Code-adjacent documentation must be self-contained. Do NOT reference external authoritative sources. If knowledge exists elsewhere, summarize it locally.

### Living Documents

`docs/spec/requirements.md` and `docs/spec/architecture.md` are updated incrementally:
- **Add** new sections when new features are built
- **Modify** existing sections when behavior changes
- **Never regenerate** the whole document — incremental updates preserve context
- **Version via git** — the document always represents current state

### Current State

`docs/state/current.md` is the single source of truth for "what's happening now":
```markdown
# Current State

## Active Work
{What's being worked on right now}

## Known Issues
{Bugs, limitations, tech debt items — with severity}

## Recent Changes
{Last 3-5 changes — brief, linked to iteration folders}

## Next Priorities
{What's coming next — helps with context when returning to the project}
```

### Workflow State

`docs/state/workflow.md` is the structured handoff for session management:
- Written by orchestrator at session boundaries
- Contains position, completed artifacts, key decisions, manifest delta, and context for next session
- Enables warm restart across sessions with minimal token cost

### Iteration Folders

Named `{NNN}-{type}-{descriptor}/` with zero-padded sequence numbers:
- `001-new-barcode-scanning/`
- `002-enhancement-reorder/`
- `003-bugfix-login-timeout/`

Each iteration contains up to 5 files (not all required for every type):

| File | Owner | Purpose |
|------|-------|---------|
| `overview.md` | Orchestrator | Scope, type, chain, traceability |
| `{analyst-output}.md` | Analyst | Varies by type (requirements, issue-analysis, delta, refactor-scope) |
| `changes.md` | Developer | What was implemented, commit hashes |
| `test-summary.md` | Tester | Test strategy, results, coverage |
| `validation.md` | Reviewer | Evidence-based review, sign-off |
| `retrospective.md` | Reviewer | Lessons learned, open items |

Lightweight enough that a bug fix generates 4-5 small files. Structured enough that a feature generates useful history.

## Trigger Quality Test

Before creating a new document, ask: "Will an agent or developer need to read this in the future?" If no, don't create it.

## Rules

1. **Use 5-zone structure.** `spec/` for specifications, `blueprints/` for planning artifacts, `state/` for runtime, `iterations/` for history, `knowledge/` for reference.
2. **No documentation in project root.** Everything under `docs/`.
3. **No scattered files.** One `current.md` for active state, not five task files.
4. **No temporal contamination.** Docs describe current state, not change history.
5. **No regeneration.** Incremental updates only on living documents.
6. **Every document has a reader.** If no one will read it, don't write it.
7. **Use canonical URIs.** `@spec/requirements#FR-3` for references.
8. **ADRs live in `spec/decisions/`.** Architecture Decision Records go in `docs/spec/decisions/` only.
9. **Spikes live in `knowledge/`.** Spike reports go in `docs/knowledge/` with `-spike.md` suffix.

## Zone Compliance

Agents must verify documentation placement before completing any workflow.

### Verification Checklist

1. All documentation files exist under `docs/`
2. Every doc is in one of the five canonical zones: `spec/`, `blueprints/`, `state/`, `iterations/`, `knowledge/`
3. ADRs are in `docs/spec/decisions/` (not `docs/adr/`, `docs/adrs/`, or `docs/architecture/`)
4. Spike reports are in `docs/knowledge/` with `-spike.md` suffix (not `docs/spikes/`)
5. No documentation files exist in project root or non-zone subdirectories
6. Blueprint documents have a corresponding entry in `docs/blueprints/INDEX.md`

### Legacy Paths (Non-Compliant)

The following paths are non-compliant and must be migrated:
- `docs/architecture/` — migrate to `docs/spec/` or `docs/knowledge/`
- `docs/adr/` — migrate to `docs/spec/decisions/`
- `docs/adrs/` — migrate to `docs/spec/decisions/`
- `docs/spikes/` — migrate to `docs/knowledge/` with `-spike.md` suffix
- `docs/current/` — migrate to `docs/spec/` or `docs/knowledge/`

---

> **See also:**
> - [../docs/reference/path-reference.md](../docs/reference/path-reference.md) — Canonical file inventory and cross-reference rules
> - [../docs/reference/scripts.md](../docs/reference/scripts.md) — `docs-gen.sh` and `validate-docs.sh` usage
> - [../docs/guides/documentation.guide.md](../docs/guides/documentation.guide.md) — Practical developer-facing guide for the 5-zone structure
