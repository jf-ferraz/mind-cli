# ASDLC Framework for Claude Code
## Complete Workflow Structure — From Planning to Governance

> **Purpose:** A production-ready, reusable directory structure for agentic software development inside the Claude Code environment, covering the full lifecycle: planning, execution, standardization, documentation, and governance.

---

## Design Principles

This framework synthesizes three bodies of knowledge into a single, actionable structure:

1. **ASDLC theory** (Spec-Driven Development, Context/Agents/Gates layers, autonomy levels)
2. **Claude Code native primitives** (CLAUDE.md, skills, agents, commands, hooks, MCP, plugins)
3. **Industry patterns** (`.agent` directory paradigm, multi-agent topologies, MCP integration)

The key insight: Claude Code already provides the scaffolding (skills, agents, commands, hooks). The framework's job is to **organize domain knowledge, governance rules, and workflow orchestration** around those primitives so that every session starts with full context and ends with auditable output.

---

## Complete Directory Tree

```
project-root/
│
├── CLAUDE.md                          # ① Constitution — project-wide laws
├── .mcp.json                          # ② Project-scoped MCP servers
│
├── .claude/                           # ③ Claude Code Runtime Layer
│   ├── settings.json                  #    Hooks, permissions, env vars
│   ├── agents/                        #    Subagent definitions
│   │   ├── architect.md               #    → System design & ADR creation
│   │   ├── implementer.md             #    → Code generation specialist
│   │   ├── reviewer.md                #    → Adversarial code review (Critic)
│   │   ├── researcher.md              #    → Codebase exploration (uses Explore)
│   │   ├── tester.md                  #    → Test generation & validation
│   │   ├── documenter.md              #    → Auto-documentation writer
│   │   └── deployer.md                #    → CI/CD and release tasks
│   │
│   ├── skills/                        #    Domain capabilities
│   │   ├── planning/                  #    ASDLC Phase 0-1
│   │   │   ├── SKILL.md               #    → Spec-driven planning orchestrator
│   │   │   ├── scripts/
│   │   │   │   └── extract-requirements.sh
│   │   │   └── references/
│   │   │       └── spec-template.md
│   │   │
│   │   ├── implementation/            #    ASDLC Phase 4
│   │   │   ├── SKILL.md               #    → Code generation with standards
│   │   │   ├── scripts/
│   │   │   │   └── scaffold-component.sh
│   │   │   └── references/
│   │   │       └── code-patterns.md
│   │   │
│   │   ├── review-gate/               #    ASDLC Phase 5 — Quality Gate
│   │   │   ├── SKILL.md               #    → Adversarial review loop
│   │   │   ├── scripts/
│   │   │   │   ├── run-linter.sh
│   │   │   │   └── complexity-check.sh
│   │   │   └── references/
│   │   │       └── review-checklist.md
│   │   │
│   │   ├── testing/                   #    ASDLC Phase 5
│   │   │   ├── SKILL.md               #    → Test strategy & execution
│   │   │   ├── scripts/
│   │   │   │   └── run-tests.sh
│   │   │   └── references/
│   │   │       └── testing-standards.md
│   │   │
│   │   ├── documentation/             #    ASDLC Phase 7
│   │   │   ├── SKILL.md               #    → Auto-doc generation
│   │   │   └── references/
│   │   │       └── doc-style-guide.md
│   │   │
│   │   └── deployment/                #    ASDLC Phase 6
│   │       ├── SKILL.md               #    → Release & deploy procedures
│   │       └── scripts/
│   │           └── pre-deploy-check.sh
│   │
│   └── commands/                      #    User-triggered slash commands
│       ├── plan.md                    #    /plan  → Kick off spec-driven planning
│       ├── implement.md               #    /implement → Execute from spec
│       ├── review.md                  #    /review → Trigger review gate
│       ├── test.md                    #    /test → Run test strategy
│       ├── document.md                #    /document → Generate docs
│       ├── deploy.md                  #    /deploy → Pre-deploy checks
│       ├── status.md                  #    /status → Project health dashboard
│       └── new-feature.md             #    /new-feature → Full lifecycle wizard
│
├── .agent/                            #    ④ Spec & Knowledge Layer (source of truth)
│   │
│   ├── spec/                          #    Blueprints — versionable intent
│   │   ├── requirements.md            #    User stories, PRD, acceptance criteria
│   │   ├── design.md                  #    System design, API contracts, data models
│   │   ├── tasks.md                   #    Backlog: State (how it works) vs Delta (what changes)
│   │   └── current-sprint.md          #    Active work items for this cycle
│   │
│   ├── wiki/                          #    Static knowledge — immutable foundation
│   │   ├── architecture.md            #    ADRs, system diagrams, tech stack rationale
│   │   ├── domain.md                  #    Business rules, glossary, domain model
│   │   ├── conventions.md             #    Naming, file organization, commit format
│   │   └── security.md               #    Auth flows, secrets policy, compliance
│   │
│   ├── links/                         #    Dynamic atlas — external integrations
│   │   ├── resources.md               #    Figma URLs, Jira boards, dashboards
│   │   ├── api-catalog.md             #    External API docs and endpoints
│   │   └── mcp-registry.md            #    Available MCP servers and capabilities
│   │
│   ├── templates/                     #    Reusable templates for agents
│   │   ├── feature-spec.md            #    Template for new feature specifications
│   │   ├── adr-template.md            #    Architecture Decision Record format
│   │   ├── bug-report.md              #    Structured bug report template
│   │   ├── pr-description.md          #    Pull request description template
│   │   └── retrospective.md           #    Sprint retrospective template
│   │
│   └── governance/                    #    Quality gates & audit trail
│       ├── gates.md                   #    Defined quality gates (deterministic + probabilistic)
│       ├── autonomy-levels.md         #    SAE-inspired autonomy taxonomy for this project
│       ├── review-log.md              #    Append-only log of gate pass/fail decisions
│       └── metrics.md                 #    Quality KPIs: complexity, coverage, drift
│
├── docs/                              #    ⑤ Human-facing documentation
│   ├── README.md                      #    Project overview
│   ├── CONTRIBUTING.md                #    Contribution guidelines
│   ├── CHANGELOG.md                   #    Release history
│   ├── setup.md                       #    Developer onboarding
│   └── runbooks/                      #    Operational procedures
│       ├── incident-response.md
│       └── rollback.md
│
└── src/                               #    ⑥ Source code (your application)
    └── ...                            #    (Organized per your stack)
```

---

## Layer-by-Layer Breakdown

### ① CLAUDE.md — The Constitution

This is the single most important file. It loads automatically at every session start. Keep it lean (under 100 instructions) and point elsewhere for details.

```markdown
# Project: [Name]

[One-line description of what this system does]

## Stack
- Language: TypeScript 5.x / Python 3.12
- Framework: [e.g., Next.js 15, FastAPI]
- Database: [e.g., PostgreSQL via Prisma]
- Testing: [e.g., Vitest, Playwright]

## Commands
- `npm run dev` — Start dev server (port 3000)
- `npm run test` — Run unit tests
- `npm run lint` — Lint + type-check
- `npm run build` — Production build

## Architecture
- `src/` — Application source code
- `.agent/spec/` — Requirements and design specs (@.agent/spec/requirements.md)
- `.agent/wiki/` — Architecture decisions and domain knowledge
- `.agent/governance/` — Quality gates and audit logs

## Workflow Rules
1. ALWAYS read `.agent/spec/tasks.md` before starting implementation work
2. ALWAYS run the /review command after completing any feature
3. NEVER modify `.agent/wiki/architecture.md` without creating an ADR first
4. NEVER skip quality gates defined in `.agent/governance/gates.md`
5. Commit messages follow Conventional Commits: `type(scope): description`

## Agent Coordination
- Use the `researcher` agent for codebase exploration (keeps main context clean)
- Use the `reviewer` agent for adversarial code review after implementation
- Complex features: /plan first → /implement second (separate sessions)

## Gotchas
- [Project-specific warnings, quirky modules, things that break easily]
- See @.agent/wiki/security.md before touching auth
```

### ② .mcp.json — Project-Scoped MCP Servers

Declares which external tool servers are available for this project. Agents use these to interact with external systems without hardcoded integrations.

```json
{
  "mcpServers": {
    "github": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" }
    },
    "postgres": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": { "DATABASE_URL": "${DATABASE_URL}" }
    },
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "./src", "./.agent"]
    }
  }
}
```

### ③ .claude/ — Claude Code Runtime Layer

This is where Claude Code's native primitives live. The framework maps ASDLC lifecycle phases onto skills, agents, and commands.

#### Agents (`.claude/agents/`)

Each agent is a focused subagent with a defined role, invoked either by other skills or directly by the user. They run in separate context windows and report back summaries.

**Example: `reviewer.md`** (The Critic — ASDLC Phase 5 Gate)
```markdown
---
name: reviewer
description: Adversarial code reviewer. Invoked after implementation to check quality gates.
model: claude-sonnet-4-5-20250929
allowed-tools: Read, Grep, Glob, Bash(npm run lint), Bash(npm run test)
---

You are an adversarial code reviewer. Your job is to FIND PROBLEMS, not approve code.

## Review Protocol
1. Read `.agent/governance/gates.md` for quality gate criteria
2. Read `.agent/spec/requirements.md` for acceptance criteria
3. Review all changed files against the spec
4. Check: type safety, error handling, edge cases, security, tests
5. Output a structured verdict: PASS / FAIL with specific line references

## Rules
- You have NO creation bias — you did not write this code
- Flag any deviation from `.agent/wiki/conventions.md`
- Check cyclomatic complexity (flag functions > 10)
- Verify test coverage for new code paths
- Append results to `.agent/governance/review-log.md`

If the code FAILS, output specific fix instructions for the implementer agent.
```

**Example: `architect.md`** (ASDLC Phase 2)
```markdown
---
name: architect
description: System design specialist. Creates and validates architecture decisions.
model: claude-sonnet-4-5-20250929
allowed-tools: Read, Grep, Glob, WebFetch
---

You are the system architect. You make structural decisions and document them as ADRs.

## Process
1. Read `.agent/wiki/architecture.md` for existing decisions
2. Read `.agent/spec/requirements.md` for the current need
3. Propose a design using the template in `.agent/templates/adr-template.md`
4. Consider: scalability, maintainability, security, team capability
5. Output the ADR to be appended to `.agent/wiki/architecture.md`

## Constraints
- NEVER contradict existing ADRs without explicit justification
- ALWAYS consider the current tech stack defined in CLAUDE.md
- Prefer composition over inheritance, simplicity over cleverness
```

#### Skills (`.claude/skills/`)

Skills are auto-discovered capabilities with bundled scripts and references. Claude loads them on demand when relevant.

**Example: `planning/SKILL.md`** (ASDLC Phases 0-1)
```markdown
---
name: planning
description: >
  Spec-driven planning for new features and projects. Use when the user says
  "plan", "spec", "requirements", "design", "scope", "new feature", "PRD",
  or "what should we build". Enforces the principle: no spec, no build.
allowed-tools: Read, Grep, Glob, Write, Edit
---

# Spec-Driven Planning

## Principle
No specification, no construction. Every feature MUST have a written spec
before implementation begins.

## Process
1. Read `.agent/spec/requirements.md` for existing context
2. Interview the user about the feature (use deep, non-obvious questions)
3. Write the spec using the template at `.agent/templates/feature-spec.md`
4. Save to `.agent/spec/` with a descriptive filename
5. Update `.agent/spec/tasks.md` with new work items
6. Define acceptance criteria that are testable and unambiguous

## Output Format
The spec must contain:
- **Problem statement** — what user pain this solves
- **Acceptance criteria** — testable conditions for "done"
- **Technical approach** — how it integrates with existing architecture
- **Risks** — what could go wrong and mitigation strategies
- **Out of scope** — what this feature explicitly does NOT do

## References
- Feature spec template: `references/spec-template.md`
- Existing requirements: `.agent/spec/requirements.md`
- Architecture constraints: `.agent/wiki/architecture.md`
```

**Example: `review-gate/SKILL.md`** (ASDLC Phase 5 — The Gate)
```markdown
---
name: review-gate
description: >
  Adversarial quality gate. Use after implementation to validate code quality.
  Triggers on: "review", "check quality", "gate", "validate", "is this ready".
  Runs deterministic checks (lint, tests) and probabilistic review (critic agent).
allowed-tools: Read, Grep, Glob, Bash(npm run lint), Bash(npm run test)
agent: reviewer
context: fork
---

# Quality Gate — Review Protocol

## Gate Types

### Deterministic Gates (must all pass)
1. `npm run lint` — zero errors
2. `npm run test` — all tests pass
3. Type check — `npx tsc --noEmit` passes
4. Run `scripts/complexity-check.sh` — no function exceeds threshold

### Probabilistic Gates (agent-reviewed)
1. Spawn the `reviewer` agent for adversarial code review
2. Check adherence to `.agent/spec/` acceptance criteria
3. Verify conventions from `.agent/wiki/conventions.md`

## Verdict
- ALL deterministic gates must PASS
- Probabilistic review produces PASS / CONDITIONAL / FAIL
- Results appended to `.agent/governance/review-log.md`
- If FAIL: return specific remediation instructions

## Scripts
- `scripts/run-linter.sh` — wrapper for project linter
- `scripts/complexity-check.sh` — cyclomatic complexity scanner
```

#### Commands (`.claude/commands/`)

These are user-triggered `/slash` commands — the entry points for each lifecycle phase.

**Example: `new-feature.md`** (Full Lifecycle Wizard)
```markdown
---
description: >
  Full feature lifecycle: plan → implement → review → test → document.
  Use for complete end-to-end feature development.
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, Skill
---

# New Feature Lifecycle

You are orchestrating the complete ASDLC lifecycle for a new feature.

## Phase 0-1: Planning
1. Invoke the `planning` skill to create a spec
2. Confirm the spec with the user before proceeding
3. **STOP HERE** — recommend starting a fresh session for implementation

## Phase 2-4: Implementation (if continuing)
1. Read the spec from `.agent/spec/`
2. Invoke the `implementation` skill
3. Create a feature branch: `feat/<feature-name>`

## Phase 5: Quality Gates
1. Invoke the `review-gate` skill
2. If FAIL → iterate with `implementer` agent
3. If PASS → proceed to documentation

## Phase 6: Documentation
1. Invoke the `documentation` skill
2. Update CHANGELOG.md

## Phase 7: Completion
1. Update `.agent/spec/tasks.md` (mark complete)
2. Provide a summary of all changes made

Feature description: $ARGUMENTS
```

**Example: `status.md`** (Project Health)
```markdown
---
description: Show project health dashboard — specs, gates, coverage, open tasks.
allowed-tools: Read, Grep, Glob, Bash(npm run test -- --coverage)
---

# Project Status Dashboard

Compile and report:

1. **Open tasks**: Read `.agent/spec/tasks.md`, count open vs completed
2. **Sprint progress**: Read `.agent/spec/current-sprint.md`
3. **Quality gates**: Read `.agent/governance/review-log.md`, show last 5 entries
4. **Test coverage**: Run `npm run test -- --coverage` and summarize
5. **Architecture debt**: Grep `.agent/wiki/architecture.md` for open TODOs

Present as a concise, scannable summary.
```

#### Settings (`.claude/settings.json`)

Hooks enforce governance automatically — no human discipline required.

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "[ \"$(git branch --show-current)\" != \"main\" ] || { echo '{\"block\": true, \"message\": \"Cannot edit on main branch. Create a feature branch first.\"}' >&2; exit 2; }",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "echo '{\"feedback\": \"Remember: run /review before merging.\"}'",
            "timeout": 5
          }
        ]
      }
    ]
  },
  "permissions": {
    "allow": [
      "Read(.agent/**)",
      "Read(docs/**)",
      "Read(src/**)",
      "Bash(npm run lint)",
      "Bash(npm run test *)",
      "Bash(npm run build)",
      "Bash(git *)"
    ],
    "deny": [
      "Read(.env)",
      "Read(.env.*)",
      "Read(**/secrets/**)",
      "Bash(rm -rf *)",
      "Bash(curl *)"
    ]
  }
}
```

### ④ .agent/ — Spec & Knowledge Layer

This is the project's **long-term memory** — the machine-readable API that agents consume. It maps directly to the `.agent` directory paradigm from the research.

**Key files explained:**

| File | Purpose | Who writes it | Who reads it |
|---|---|---|---|
| `spec/requirements.md` | What to build | Human + planning skill | All agents |
| `spec/design.md` | How to build it | architect agent | implementer agent |
| `spec/tasks.md` | What's left to do | Human + agents | /status, /plan |
| `wiki/architecture.md` | Why decisions were made | architect agent | All agents |
| `wiki/domain.md` | Business rules & glossary | Human | All agents |
| `wiki/conventions.md` | Code standards | Human | reviewer, implementer |
| `governance/gates.md` | Quality criteria | Human | review-gate skill |
| `governance/review-log.md` | Audit trail | reviewer agent | /status, humans |
| `templates/*` | Reusable formats | Human | planning, architect |

### ⑤ docs/ — Human-Facing Documentation

Standard project documentation that lives outside the agent layer. The `documentation` skill generates and updates these files.

### ⑥ src/ — Application Source Code

Your actual application code. The framework is stack-agnostic — organize `src/` however your technology demands.

---

## ASDLC Phase Mapping

| ASDLC Phase | Command | Skill | Agent(s) | Key Files |
|---|---|---|---|---|
| 0 — Preparation | `/plan` | `planning` | researcher | `spec/requirements.md` |
| 1 — Scope | `/plan` | `planning` | architect | `spec/design.md`, `spec/tasks.md` |
| 2 — Agent Design | — | — | — | `wiki/architecture.md` (ADRs) |
| 3 — Simulation | — | `implementation` | architect | `spec/design.md` |
| 4 — Implementation | `/implement` | `implementation` | implementer | `src/` |
| 5 — Gates | `/review`, `/test` | `review-gate`, `testing` | reviewer, tester | `governance/review-log.md` |
| 6 — Deploy | `/deploy` | `deployment` | deployer | `docs/CHANGELOG.md` |
| 7 — Governance | `/status` | `documentation` | documenter | `governance/metrics.md` |

---

## Workflow Patterns

### Pattern A: Sequential Pipeline (Simple Features)
```
/plan → /implement → /review → /test → /document → /deploy
```
Each command triggers the next phase. Best for small, well-understood features.

### Pattern B: Supervisor + Critic Loop (Complex Features)
```
/new-feature
  ├── planning skill (creates spec)
  ├── architect agent (validates design)
  ├── implementer agent (writes code)
  ├── reviewer agent (adversarial review) ← LOOP until PASS
  ├── tester agent (generates + runs tests)
  └── documenter agent (updates docs)
```

### Pattern C: Parallel Exploration (Research Phase)
```
User: "Investigate how we should handle real-time notifications"

Claude spawns in parallel:
  ├── researcher agent → explores codebase for existing patterns
  ├── architect agent → evaluates WebSocket vs SSE vs polling
  └── Main context receives summaries → user decides
```

---

## Getting Started

### 1. Initialize the structure
```bash
mkdir -p .agent/{spec,wiki,links,templates,governance}
mkdir -p .claude/{agents,commands,skills}
mkdir -p docs/runbooks
```

### 2. Create your CLAUDE.md
Run `/init` in Claude Code, then refine with the template from Section ① above.

### 3. Populate the knowledge layer
Start with these files (even if minimal):
- `.agent/spec/requirements.md` — what you're building
- `.agent/wiki/architecture.md` — key decisions so far
- `.agent/wiki/conventions.md` — your code standards
- `.agent/governance/gates.md` — what "done" means

### 4. Add skills incrementally
Don't create all skills at once. Start with `planning` and `review-gate`, then add others as friction points emerge.

### 5. Configure MCP servers
Add external integrations to `.mcp.json` as needed (GitHub, database, monitoring).

---

## Scaling Considerations

**Monorepos:** Place a root `CLAUDE.md` with global rules, then add subdirectory `CLAUDE.md` files for each package/service. Skills and agents at root apply everywhere; package-level `.claude/` adds specialization.

**Team collaboration:** Commit `.agent/`, `.claude/commands/`, `.claude/skills/`, and `CLAUDE.md` to version control. Use `.claude/agents/` for shared agent definitions. Personal preferences go in `~/.claude/CLAUDE.md` (global, not committed).

**Context management:** Use the `researcher` agent (Explore mode) for codebase exploration — it runs in a separate context window and returns summaries, keeping your main context clean for implementation. Use `/compact` with custom instructions to preserve critical context during long sessions.

---

*Framework version: 1.0 — February 2026*
*Based on ASDLC principles, Claude Code primitives, and the .agent directory standard.*
