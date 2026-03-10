# Blueprint: AI Workflow Bridge

> How the `mind` CLI/TUI connects to AI agent workflows — four integration models from lightweight to full orchestration.

**Status**: Proposal
**Date**: 2026-03-09
**Depends on**: [01-mind-cli.md](01-mind-cli.md)

---

## Problem

Blueprint 01 explicitly scoped the CLI as deterministic — "no AI in the CLI." That's correct for the core tool, but it leaves a gap: the user must manually switch between the CLI (project intelligence, validation, scaffolding) and Claude Code / Copilot Chat (AI workflows) with no coordination between them.

Real friction points:

1. **Context assembly is manual** — Before running `/workflow`, the user should check docs completeness, verify the brief, review recent iterations. This is 3-4 separate steps they must remember.
2. **Gate results don't flow back** — The deterministic gate (build/lint/test) runs inside Claude Code's context. If it fails, the AI retries — but the user has no visibility into what's happening.
3. **State is invisible** — Workflow state lives in `docs/state/workflow.md` as markdown. The user must read the file to know where they are.
4. **Post-workflow cleanup is forgotten** — After a workflow completes, `docs/state/current.md` should be updated, validation should run, the branch should be ready for PR. Nobody remembers all of these.
5. **No feedback loop** — The AI agents can't ask "is the documentation structure valid?" or "what iterations exist?" without reading files manually each time.

---

## Four Integration Models

Each model builds on the previous. They can be implemented incrementally.

```
Model A: Pre-Flight + Handoff       (CLI prepares, AI executes)
Model B: MCP Server                  (AI calls CLI as a tool)
Model C: Sidecar / Watch Mode        (CLI monitors AI in real-time)
Model D: Full Orchestration          (CLI drives the entire workflow)
```

---

## Model A: Pre-Flight + Handoff

**Complexity**: Low
**Value**: High
**Dependencies**: Blueprint 01 only

The CLI does all deterministic work before the AI starts, then hands off a prepared context package.

### Commands

```
mind preflight "create: REST API with JWT auth"
mind preflight --resume
mind handoff <iteration-id>
```

### Flow

```
User runs: mind preflight "create: REST API with JWT auth"

    ┌─────────────────────────────────────────────────────────┐
    │  1. Classify request                                     │
    │     → NEW_PROJECT (keywords: "create")                   │
    │                                                          │
    │  2. Business context gate                                │
    │     → Check docs/spec/project-brief.md                   │
    │     → Result: BRIEF_PRESENT (Vision ✓, Deliverables ✓,   │
    │       Scope ✓)                                           │
    │                                                          │
    │  3. Validate documentation                               │
    │     → Run 17-check validation                            │
    │     → Result: 15 pass, 0 fail, 2 warnings                │
    │                                                          │
    │  4. Create iteration                                     │
    │     → docs/iterations/007-NEW_PROJECT-rest-api/          │
    │     → 5 template files created                           │
    │     → overview.md populated with classification + scope   │
    │                                                          │
    │  5. Create git branch                                    │
    │     → git checkout -b new/rest-api                       │
    │                                                          │
    │  6. Assemble context package                             │
    │     → Read project-brief.md, requirements.md,            │
    │       architecture.md, recent iterations,                 │
    │       any relevant convergence docs                       │
    │     → Write docs/state/workflow.md with full state        │
    │                                                          │
    │  7. Generate prompt                                      │
    │     → Orchestrator instructions + context + iteration     │
    │     → Copy to clipboard / print / save to file            │
    └─────────────────────────────────────────────────────────┘

    Output:
    ╭─ Pre-Flight Complete ─────────────────────────────────╮
    │                                                        │
    │  Type: NEW_PROJECT                                     │
    │  Chain: analyst → architect → developer → tester       │
    │         → reviewer                                     │
    │  Branch: new/rest-api                                  │
    │  Iteration: docs/iterations/007-NEW_PROJECT-rest-api/  │
    │  Brief: ✓ present (Vision, Deliverables, Scope)        │
    │  Docs: 15/17 pass (2 warnings, 0 blockers)             │
    │                                                        │
    │  Context package ready. Open Claude Code and run:       │
    │  /workflow "create: REST API with JWT auth"             │
    │                                                        │
    │  Or paste the generated prompt (copied to clipboard).   │
    ╰────────────────────────────────────────────────────────╯
```

### What This Solves

- User never forgets to check the brief, validate docs, or create the iteration folder
- Classification happens before the AI starts (saves tokens — orchestrator can skip Step 1-4)
- If the brief is missing, the CLI tells the user immediately instead of burning an AI session
- Git branch is created consistently (`{type}/{descriptor}`)
- The AI starts with a clean, pre-validated environment

### `mind preflight --resume`

Checks `docs/state/workflow.md` for an interrupted workflow:

```
╭─ Resumable Workflow Found ─────────────────────────────────╮
│                                                              │
│  Type: NEW_PROJECT                                           │
│  Iteration: 007-NEW_PROJECT-rest-api                         │
│  Last Agent: architect (completed)                           │
│  Remaining: developer → tester → reviewer                    │
│  Branch: new/rest-api                                        │
│  Session: 1 of 2 (split after architect)                     │
│                                                              │
│  Completed Artifacts:                                        │
│    ✓ requirements.md (analyst)                               │
│    ✓ architecture.md (architect)                             │
│                                                              │
│  Resume in Claude Code:                                      │
│  /workflow                                                   │
╰──────────────────────────────────────────────────────────────╯
```

### `mind handoff <iteration-id>`

After a workflow completes, run post-workflow cleanup:

```
mind handoff 007

    1. Validate iteration completeness
       → overview.md ✓, changes.md ✓, test-summary.md ✓,
         validation.md ✓, retrospective.md ✓

    2. Run deterministic checks
       → cargo build ✓, cargo test ✓, cargo clippy ✓

    3. Update docs/state/current.md
       → Active Work: None
       → Recent Changes: 007-NEW_PROJECT-rest-api
       → Next Priorities: (prompt user)

    4. Clear workflow state
       → docs/state/workflow.md → idle

    5. Prepare for PR
       → Branch: new/rest-api (3 commits ahead of main)
       → Suggestion: mind pr (or gh pr create)
```

---

## Model B: MCP Server

**Complexity**: Medium
**Value**: Very High
**Dependencies**: Blueprint 01 + MCP protocol support

The CLI runs as an MCP (Model Context Protocol) server. Claude Code connects to it via `.mcp.json` and gets project intelligence as callable tools — the AI agents become aware of the framework's state without manually reading files.

### Configuration

```json
// .mcp.json (at project root)
{
  "mcpServers": {
    "mind": {
      "command": "mind",
      "args": ["serve"],
      "description": "Mind Framework project intelligence"
    }
  }
}
```

### Exposed Tools

```
┌─────────────────────────────────────────────────────────────────────────┐
│  MCP Tool                 │ Returns                                     │
├───────────────────────────┼─────────────────────────────────────────────┤
│  mind_status              │ Project health summary (JSON)               │
│  mind_doctor              │ Diagnostics with fix suggestions            │
│  mind_check_brief         │ Business context gate result                │
│  mind_validate_docs       │ 17-check validation results                 │
│  mind_validate_refs       │ 11-check cross-reference results            │
│  mind_list_iterations     │ All iterations with type/status/date        │
│  mind_show_iteration      │ Single iteration details + artifacts        │
│  mind_read_state          │ Current workflow state (parsed)             │
│  mind_update_state        │ Write workflow position/artifacts            │
│  mind_create_iteration    │ Create iteration folder + overview.md       │
│  mind_list_stubs          │ Documents that need content                 │
│  mind_check_gate          │ Run deterministic gate (build/lint/test)    │
│  mind_log_quality         │ Extract convergence scores to quality log   │
│  mind_search_docs         │ Full-text search across docs/               │
│  mind_read_config         │ Parse and return mind.toml manifest         │
│  mind_suggest_next        │ What should happen next, based on state     │
└───────────────────────────┴─────────────────────────────────────────────┘
```

### How Agents Use the Tools

#### Orchestrator (Step 1.5: Business Context Gate)

Before:
```
Read docs/spec/project-brief.md
Parse it manually
Determine if it's a stub
Check for Vision, Key Deliverables, Scope sections
```

After (with MCP):
```
Call mind_check_brief
→ { "status": "BRIEF_PRESENT", "sections": { "vision": true, "deliverables": true, "scope": true }, "is_stub": false }
```

The orchestrator gets a structured answer instead of parsing markdown heuristically.

#### Orchestrator (Step 4: Create Iteration)

Before:
```
Scan docs/iterations/ for existing folders
Calculate next sequence number
Create folder with 5 files
Fill overview.md template
```

After (with MCP):
```
Call mind_create_iteration { "type": "NEW_PROJECT", "descriptor": "rest-api", "request": "create: REST API with JWT auth" }
→ { "path": "docs/iterations/007-NEW_PROJECT-rest-api", "branch": "new/rest-api", "overview": "docs/iterations/007-NEW_PROJECT-rest-api/overview.md" }
```

Eliminates sequencing bugs and inconsistent folder creation across agents.

#### Developer (Before Committing)

```
Call mind_check_gate
→ { "build": { "pass": true, "command": "cargo build", "duration_ms": 4200 },
     "lint": { "pass": false, "command": "cargo clippy", "errors": 2, "output": "..." },
     "test": { "pass": true, "command": "cargo test", "passed": 142, "failed": 0 } }
```

The developer sees exactly what failed and can fix it before the deterministic gate runs formally.

#### Reviewer (Evidence Gathering)

```
Call mind_status
→ { "docs_health": { "spec": { "complete": 4, "total": 5 }, ... },
     "iterations": { "total": 7, "latest": "007-NEW_PROJECT-rest-api" },
     "warnings": ["domain-model.md is a stub"] }

Call mind_list_iterations { "type": "NEW_PROJECT" }
→ [{ "id": "007", "type": "NEW_PROJECT", "descriptor": "rest-api", "status": "in_progress", "artifacts": {...} }]
```

The reviewer can check project-level health without manually traversing the file tree.

#### Any Agent (Context Awareness)

```
Call mind_suggest_next
→ { "suggestion": "The analyst should proceed with requirements extraction.",
     "reason": "Business context gate passed. Project brief has substantive content.",
     "warnings": ["domain-model.md is a stub — analyst should create it"],
     "context": { "recent_iterations": [...], "open_issues": [...] } }
```

### Why MCP Is the Highest-Value Integration

1. **Zero workflow change** — Users keep using `/workflow`, `/discover`, `/analyze` exactly as before. The AI agents just get better tools.
2. **Structured over heuristic** — Agents stop parsing markdown files and start calling tools that return JSON. Fewer hallucinations, more reliable gates.
3. **State consistency** — One tool (`mind_update_state`) writes workflow state. No more inconsistent markdown formatting across agent implementations.
4. **Composable** — Any new validation check, diagnostic, or intelligence can be added as an MCP tool without modifying agent instructions.
5. **Platform-agnostic** — MCP works with Claude Code today, and the protocol is open. Other AI tools can connect to the same server.

### MCP Server Architecture

```go
// cmd/serve.go
package cmd

// mind serve — starts MCP server on stdio
// Implements the MCP protocol (JSON-RPC over stdio)
// Each tool maps to an internal/ package function

// Tool registration:
// tools := []mcp.Tool{
//     {Name: "mind_status",           Handler: status.Handle},
//     {Name: "mind_doctor",           Handler: doctor.Handle},
//     {Name: "mind_check_brief",      Handler: brief.Handle},
//     {Name: "mind_validate_docs",    Handler: validate.DocsHandle},
//     {Name: "mind_validate_refs",    Handler: validate.RefsHandle},
//     {Name: "mind_list_iterations",  Handler: iteration.ListHandle},
//     {Name: "mind_show_iteration",   Handler: iteration.ShowHandle},
//     {Name: "mind_read_state",       Handler: state.ReadHandle},
//     {Name: "mind_update_state",     Handler: state.UpdateHandle},
//     {Name: "mind_create_iteration", Handler: generate.IterationHandle},
//     {Name: "mind_list_stubs",       Handler: docs.StubsHandle},
//     {Name: "mind_check_gate",       Handler: gate.Handle},
//     {Name: "mind_log_quality",      Handler: quality.LogHandle},
//     {Name: "mind_search_docs",      Handler: docs.SearchHandle},
//     {Name: "mind_read_config",      Handler: config.Handle},
//     {Name: "mind_suggest_next",     Handler: suggest.Handle},
// }
```

The MCP server reuses the same `internal/` packages from the CLI. No duplication — `mind check docs` and `mind_validate_docs` call the same Go function.

---

## Model C: Sidecar / Watch Mode

**Complexity**: Medium
**Value**: High
**Dependencies**: Blueprint 01 + TUI (Phase 3)

The CLI runs alongside Claude Code, watching the filesystem for changes and providing real-time feedback. Think of it as a build monitor, but for the entire framework lifecycle.

### Command

```
mind watch [--tui]
```

### What It Watches

```
┌─────────────────────────────────────────────────────────────────────────┐
│  File Pattern                  │ Event        │ Action                  │
├────────────────────────────────┼──────────────┼─────────────────────────┤
│  docs/state/workflow.md        │ Modified     │ Parse state, update     │
│                                │              │ dashboard with current  │
│                                │              │ agent + progress        │
│                                │              │                         │
│  docs/iterations/*/overview.md │ Created      │ New iteration detected  │
│                                │              │ → show type + chain     │
│                                │              │                         │
│  docs/iterations/*/changes.md  │ Modified     │ Developer produced      │
│                                │              │ changes → run Micro-    │
│                                │              │ Gate B checks silently  │
│                                │              │                         │
│  docs/iterations/*/            │ Modified     │ Validation artifact     │
│    validation.md               │              │ → parse reviewer        │
│                                │              │   findings, show        │
│                                │              │   MUST/SHOULD/COULD     │
│                                │              │                         │
│  docs/spec/requirements.md     │ Modified     │ Analyst produced reqs   │
│                                │              │ → run Micro-Gate A      │
│                                │              │   checks silently       │
│                                │              │                         │
│  docs/spec/architecture.md     │ Modified     │ Architect produced      │
│                                │              │ design → show summary   │
│                                │              │                         │
│  docs/knowledge/*-             │ Created      │ Convergence analysis    │
│    convergence.md              │              │ complete → run 23-check │
│                                │              │ validation, show score  │
│                                │              │                         │
│  src/**/*                      │ Modified     │ Code changed → run      │
│                                │              │ build/test in background│
│                                │              │ show results            │
│                                │              │                         │
│  docs/spec/project-brief.md   │ Created/     │ Brief changed → re-run  │
│                                │ Modified     │ business context gate   │
│                                │              │ check, update status    │
└────────────────────────────────┴──────────────┴─────────────────────────┘
```

### Watch TUI

Split-screen: the left half is the watch dashboard, the right half is the user's terminal running Claude Code.

```
╭─ mind watch ─── iron-arch-v2 ─── new/rest-api ──────────────────────────────────╮
│                                                                                    │
│  Workflow: NEW_PROJECT (007-rest-api)                                              │
│  Chain: analyst ✓ → architect ✓ → [developer] → tester → reviewer                 │
│                                                                                    │
│  ┌─ Live Activity ────────────────────────────────────────────────────────────────┐ │
│  │  14:23:01  Developer writing src/routes/users.rs                               │ │
│  │  14:23:15  Developer writing src/routes/auth.rs                                │ │
│  │  14:23:32  Developer created src/middleware/jwt.rs                              │ │
│  │  14:24:01  changes.md updated — 4 files added                                  │ │
│  │  14:24:02  ▸ Micro-Gate B: checking...                                         │ │
│  │  14:24:03  ▸ Micro-Gate B: ✓ changes.md exists, 4 files on disk               │ │
│  │  14:24:10  Developer writing src/models/user.rs                                │ │
│  │  14:24:45  ▸ Background: cargo build ✓ (3.2s)                                  │ │
│  │  14:25:12  ▸ Background: cargo test ✓ 24 passed (4.1s)                         │ │
│  └────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                    │
│  ┌─ Pre-Gate Status ──────────────────────────────────────────────────────────────┐ │
│  │  Build: ✓ passing          Lint: ⏳ not run yet    Tests: ✓ 24/24 passing      │ │
│  │  Micro-Gate B: ✓ passing   Deterministic Gate: ready                           │ │
│  └────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                    │
│  Warnings: domain-model.md is still a stub                                         │
│                                                                                    │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

### Why Watch Mode Matters

1. **Visibility** — The user sees what the AI is doing in real-time without reading the Claude Code output stream.
2. **Early failure detection** — If `cargo build` breaks after the developer writes file 3 of 10, the user sees it immediately and can intervene.
3. **Gate confidence** — By the time the deterministic gate runs formally, the user already knows it will pass (or knows what's failing).
4. **Passive validation** — No manual step required. Validators run automatically as files change.

### Implementation

Use `fsnotify` (Go) or `notify` (Rust) for filesystem watching. Debounce rapid changes (AI writes multiple files quickly). Run build/test commands in background goroutines with output captured.

---

## Model D: Full Orchestration

**Complexity**: High
**Value**: Very High
**Dependencies**: Blueprint 01 + Claude Code CLI (`claude`) available on PATH

The CLI becomes the workflow orchestrator. It dispatches each agent as a separate Claude Code CLI invocation, runs quality gates between them, and manages state. The user drives the entire workflow from the terminal without opening Claude Code directly.

### Command

```
mind run "create: REST API with JWT auth"
mind run --resume
mind run --dry-run "fix: 500 error on /api/users"
```

### Flow

```
User runs: mind run "create: REST API with JWT auth"

    ┌──────────────────────────────────────────────────────────┐
    │  Step 1: Pre-flight (same as Model A)                    │
    │  → Classify: NEW_PROJECT                                 │
    │  → Business context gate: PASS                           │
    │  → Create iteration: 007-NEW_PROJECT-rest-api            │
    │  → Create branch: new/rest-api                           │
    │  → Assemble context package                              │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 2: Dispatch Analyst                                │
    │                                                          │
    │  Build prompt:                                           │
    │    - Load .mind/agents/analyst.md                        │
    │    - Inject: project-brief.md, iteration overview,       │
    │      existing requirements.md (if any)                   │
    │    - Inject: conventions (shared.md, documentation.md)   │
    │                                                          │
    │  Execute:                                                │
    │    echo "$prompt" | claude --model opus --print           │
    │              --allowedTools Read,Write,Edit,Grep,Glob     │
    │                                                          │
    │  Capture output → parse for artifacts                    │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 3: Micro-Gate A                                    │
    │                                                          │
    │  Check (deterministic, no AI):                           │
    │    ✓ requirements.md updated                             │
    │    ✓ GIVEN/WHEN/THEN criteria present                    │
    │    ✓ FR-N identifiers traceable                          │
    │    ✓ Scope boundary defined                              │
    │    ✗ Missing: success metrics                            │
    │                                                          │
    │  Result: RETRY (1 of 2)                                  │
    │  → Re-dispatch analyst with failure feedback             │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 4: Dispatch Architect                              │
    │  (same pattern: build prompt → claude --print → capture) │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 5: Session split decision                          │
    │                                                          │
    │  NEW_PROJECT auto-splits after architect.                │
    │  Save state to docs/state/workflow.md                    │
    │  Prompt user: "Continue to developer, or pause?"         │
    └──────────────────┬───────────────────────────────────────┘
                       │ (user continues)
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 6: Dispatch Developer                              │
    │  → Build prompt with requirements + architecture         │
    │  → claude --model sonnet --print --allowedTools ...      │
    │  → Micro-Gate B (deterministic check)                    │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 7: Dispatch Tester                                 │
    │  → Build prompt with requirements + changes.md           │
    │  → claude --model sonnet --print --allowedTools ...      │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 8: Deterministic Gate                              │
    │                                                          │
    │  Run (no AI, pure CLI):                                  │
    │    cargo build     → ✓                                   │
    │    cargo clippy     → ✓                                  │
    │    cargo test       → ✗ 2 failures                       │
    │                                                          │
    │  Result: RETRY → re-dispatch developer with test output  │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 9: Dispatch Reviewer                               │
    │  → Build prompt with full iteration context + git diff   │
    │  → claude --model opus --print --allowedTools ...        │
    │  → Parse validation.md for sign-off                      │
    └──────────────────┬───────────────────────────────────────┘
                       │
    ┌──────────────────▼───────────────────────────────────────┐
    │  Step 10: Post-workflow (same as mind handoff)           │
    │  → Update docs/state/current.md                          │
    │  → Clear workflow state                                  │
    │  → Print summary + PR suggestion                         │
    └──────────────────────────────────────────────────────────┘
```

### TUI for Full Orchestration

When running `mind run --tui "create: ..."`, the TUI shows the full pipeline progress:

```
╭─ mind run ─── NEW_PROJECT: rest-api ─────────────────────────────────────────────╮
│                                                                                    │
│  Pipeline Progress                                                                 │
│  ─────────────────                                                                 │
│  ✓ Pre-flight     classify, gate, iteration, branch          0.8s                  │
│  ✓ Analyst         requirements extracted (12 FR, 8 AC)       2m 14s               │
│  ✓ Micro-Gate A    6/6 checks pass                            0.2s                  │
│  ✓ Architect       architecture designed (4 components)       3m 01s               │
│  ▸ Developer       implementing... (src/routes/users.rs)      1m 32s               │
│  ○ Micro-Gate B    waiting                                                         │
│  ○ Tester          waiting                                                         │
│  ○ Det. Gate       waiting                                                         │
│  ○ Reviewer        waiting                                                         │
│                                                                                    │
│  ┌─ Developer Output (live) ──────────────────────────────────────────────────────┐ │
│  │  Creating src/routes/users.rs — CRUD endpoints for User entity                │ │
│  │  Creating src/routes/auth.rs — JWT authentication endpoints                   │ │
│  │  Creating src/middleware/jwt.rs — Token validation middleware                  │ │
│  │  Creating src/models/user.rs — User struct + database queries                 │ │
│  │  ...                                                                           │ │
│  └────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                    │
│  Tokens: 42,318 in / 18,204 out    Cost: ~$1.82    Elapsed: 6m 47s                │
│                                                                                    │
│  [p]ause  [s]kip agent  [a]bort  [d]etails  [l]og                                 │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

### `mind run --dry-run`

Simulates the workflow without calling the AI. Shows what would happen:

```
$ mind run --dry-run "fix: 500 error on /api/users"

Dry Run — no AI calls will be made

  Classification: BUG_FIX
  Chain: analyst → developer → tester → reviewer
  Business Context Gate: SKIPPED (BUG_FIX)
  Iteration: would create 008-BUG_FIX-fix-500-error/
  Branch: would create fix/fix-500-error

  Agent Dispatch Plan:
    1. Analyst  (opus)   — read codebase, trace error, define reproduction steps
    2. Developer (sonnet) — implement fix
    3. Tester   (sonnet) — write regression test
    4. Reviewer (opus)   — verify root cause addressed

  Quality Gates:
    Micro-Gate A — after analyst
    Micro-Gate B — after developer
    Deterministic Gate — before reviewer
      Commands: cargo build, cargo clippy, cargo test

  Estimated cost: ~$0.80-1.50 (4 agent calls)
```

### Agent Prompt Assembly

The CLI builds each agent's prompt from structured components:

```
┌─────────────────────────────────────────────────────┐
│  Agent Prompt (assembled by CLI)                     │
│                                                      │
│  1. System: Agent instructions                       │
│     ← .mind/agents/{agent}.md                        │
│                                                      │
│  2. Context: Project state                           │
│     ← docs/spec/project-brief.md                     │
│     ← docs/spec/requirements.md (if exists)          │
│     ← docs/spec/architecture.md (if exists)          │
│     ← docs/state/current.md                          │
│     ← docs/iterations/{current}/overview.md          │
│                                                      │
│  3. Context: Prior agent outputs (this workflow)     │
│     ← Previous agent's artifacts                     │
│     ← Gate results (if retry)                        │
│                                                      │
│  4. Context: Convergence (if COMPLEX_NEW)            │
│     ← docs/knowledge/{topic}-convergence.md          │
│                                                      │
│  5. Conventions                                      │
│     ← .mind/conventions/shared.md                    │
│     ← .mind/conventions/documentation.md             │
│                                                      │
│  6. Task: Specific instruction                       │
│     ← "Analyze the following request and produce     │
│        requirements: {user's original request}"      │
│                                                      │
│  7. Constraints                                      │
│     ← Model: {opus|sonnet} (from agent frontmatter)  │
│     ← Allowed tools: Read, Write, Edit, Grep, Glob   │
│     ← Output: Write to {iteration path}              │
└─────────────────────────────────────────────────────┘
```

### Claude Code CLI Integration

The `claude` CLI supports non-interactive use:

```bash
# Basic dispatch
echo "$prompt" | claude --print --model opus

# With tool restrictions
echo "$prompt" | claude --print --model sonnet \
  --allowedTools "Read,Write,Edit,Grep,Glob,Bash"

# With MCP server (Model B + Model D combined)
echo "$prompt" | claude --print --model opus \
  --mcp-server "mind:mind serve"

# Resume a conversation (multi-turn agent)
echo "$followup" | claude --print --resume --conversation-id "$conv_id"
```

### Retry Logic

```
  Agent dispatched
       │
       ▼
  Agent completes
       │
       ▼
  Quality gate runs (deterministic)
       │
   ┌───┴───┐
   │ Pass  │ Fail
   │       │
   │       ▼
   │   Retry count < 2?
   │       │
   │   ┌───┴───┐
   │   │ Yes   │ No
   │   │       │
   │   ▼       ▼
   │  Re-dispatch   Proceed with
   │  agent with    documented
   │  gate feedback concerns
   │       │          │
   └───────┴──────────┘
           │
           ▼
      Next agent
```

---

## Model Comparison

| Aspect | A: Pre-Flight | B: MCP Server | C: Sidecar | D: Orchestrator |
|--------|--------------|---------------|------------|-----------------|
| **Complexity** | Low | Medium | Medium | High |
| **User workflow change** | Minimal | None | None | Significant |
| **AI workflow change** | None | Agents get new tools | None | Agents run headless |
| **Real-time feedback** | No | No | Yes | Yes |
| **Token savings** | ~10-15% | ~5-10% | None | ~20-30% |
| **Gate reliability** | Same | Higher | Same | Highest |
| **State consistency** | Better | Much better | Same | Best |
| **Requires `claude` CLI** | No | No | No | Yes |
| **Works with Copilot** | Partially | Yes (if MCP) | Yes | No |

### Token Savings Explanation

- **Model A** saves tokens because the orchestrator skips classification, gate checks, and iteration creation (CLI already did it).
- **Model B** saves tokens because agents call `mind_validate_docs` (returns JSON) instead of reading + parsing markdown files.
- **Model D** saves the most because each agent runs in isolation with exactly the context it needs — no orchestrator agent consuming tokens between dispatches, and the CLI handles all gate logic deterministically.

---

## Recommended Implementation Order

```
Phase 1 (Blueprint 01)     CLI foundation: status, doctor, create, check, docs
Phase 2 (Model A)           Pre-flight + handoff commands
Phase 3 (Model B)           MCP server (mind serve)
Phase 4 (Model C)           Watch mode (mind watch)
Phase 5 (Model D)           Full orchestration (mind run)
```

**Model B is the inflection point** — once the MCP server exists, every subsequent model becomes easier because the tools are already built. Model D's prompt assembly and gate logic reuse Model B's internal functions.

### Why Not Start With Model D?

Model D is the most impressive but has the most risk:
- Depends on the `claude` CLI's non-interactive mode being stable and feature-complete
- Prompt assembly is complex — wrong context injection produces bad agent outputs
- Retry logic across process boundaries is error-prone
- Debugging is harder (which agent produced bad output? what context did it see?)

Starting with A→B→C→D means each model is validated before the next builds on it. The MCP server (Model B) is the safest high-value target — it improves AI workflows without changing them.

---

## New Commands Summary

These commands are added on top of Blueprint 01:

```
mind preflight "<request>"         Model A — prepare everything, hand off to AI
mind preflight --resume            Model A — check for resumable workflow
mind handoff <iteration-id>        Model A — post-workflow cleanup
mind serve                         Model B — start MCP server (stdio)
mind watch [--tui]                 Model C — filesystem watcher + dashboard
mind run "<request>"               Model D — full orchestrated workflow
mind run --resume                  Model D — resume interrupted orchestrated workflow
mind run --dry-run "<request>"     Model D — simulate without AI calls
mind run --tui "<request>"         Model D — orchestrated workflow with live TUI
```

---

## Open Questions

1. **Claude Code CLI stability** — Model D depends on `claude --print` being reliable for headless agent dispatch. This needs validation.
2. **MCP tool granularity** — Should `mind_validate_docs` return all 17 check results, or should each check be a separate tool? Fewer tools = simpler, but less composable.
3. **Cost tracking** — Model D can track token usage and cost per agent. Should this be stored in the iteration folder? In a separate cost log?
4. **Parallel agents** — Some agents in the chain are independent (e.g., tester could start while developer is still writing late files). Model D could dispatch them in parallel. Worth the complexity?
5. **Copilot Chat parity** — Models A, B, and C work with Copilot Chat. Model D is Claude Code-specific. Is that acceptable?
6. **Conversation analysis orchestration** — Model D could also orchestrate `/analyze` workflows: dispatch each persona as a separate `claude` invocation, run Gate 0 between phases. This would make convergence analysis more reliable but adds significant complexity.

---

> **See also:**
> - [01-mind-cli.md](01-mind-cli.md) — CLI/TUI foundation (prerequisite)
> - `../../agents/orchestrator.md` — Orchestrator logic this blueprint externalizes
> - `../../conversation/protocols/state-management/PROTOCOL.md` — State management protocol
> - `../../conversation/protocols/phase-routing/PROTOCOL.md` — Phase routing for conversation analysis
> - `../../docs/reference/quality-gates.md` — Gate definitions
