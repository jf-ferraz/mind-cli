# Mind Framework — Improvement Plan

> **Status:** ✅ Complete (48/52 tasks — 3 E2E deferred, 1 pre-commit on hold)
> **Scope:** Full ecosystem integration (Claude Code + Copilot Chat) + recommended framework improvements
> **Supersedes:** `docs/archive/conversation/conversation-framework-claude-integration.md` (incorporated into Phase 1)
> **Created:** 2026-02-25
>
> **⚠️ Path note:** Agent paths in this plan reflect the original implementation target (`agents/conversation-*.md`).
> These files were later reorganized to `agents/personas/conversation-*.md` — see `ARCHITECTURE.md` for current layout.

---

## Table of Contents

1. [Current State Assessment](#1-current-state-assessment)
2. [Phase 1 — Claude Code Agent Port](#2-phase-1--claude-code-agent-port-critical)
3. [Phase 2 — Tooling & Automation](#3-phase-2--tooling--automation)
4. [Phase 3 — Documentation Consolidation](#4-phase-3--documentation-consolidation)
5. [Phase 4 — Quality & Testing Infrastructure](#5-phase-4--quality--testing-infrastructure)
6. [Phase 5 — Feature Extensions](#6-phase-5--feature-extensions)
7. [Platform Coexistence Model](#7-platform-coexistence-model)
8. [Known Constraints & Mitigations](#8-known-constraints--mitigations)
9. [Verification Protocol](#9-verification-protocol)
10. [Timeline & Priority Matrix](#10-timeline--priority-matrix)

---

## 1. Current State Assessment

### What's Done

The conversation module has been fully integrated from the `panel-module` fork into the root framework:

| Component | Location | Status |
|-----------|----------|--------|
| Dev workflow agents (7) | `agents/` | Complete — orchestrator, analyst, architect, developer, tester, reviewer, discovery |
| Conversation configs (4) | `conversation/config/` | Complete — conversation.yml, personas.yml, quality.yml, extensions.yml |
| Conversation protocols (4) | `conversation/protocols/` | Complete — phase-routing, evaluator-optimizer, approval-gates, mediated-delegation |
| Conversation skills (5) | `conversation/skills/` | Complete — evidence-standards, reasoning-chains, challenge-methodology, decision-documentation, scope-discipline |
| Copilot Chat agents (6) | `.github/agents/` | Complete — moderator + 5 personas |
| Copilot Chat prompts (2) | `.github/prompts/` | Complete — analyze.prompt.md, analyze-documents.prompt.md |
| Commands (3) | `commands/` | Complete — workflow.md, discover.md, analyze.md |
| COMPLEX_NEW workflow type | `agents/orchestrator.md` | Complete — classification, routing, Gate 0 |
| Convergence awareness | All 7 pipeline agents | Complete — each reads `analysis/conversation/` when available |
| Cross-workflow conventions | `conventions/shared.md` | Complete — evidence, reasoning, severity, scope, intent markers |
| Architecture docs | 3 files updated | Complete — arch-framework, arch-blueprint, imp-architecture-v2 |
| Module routing files | `conversation/CLAUDE.md`, `.github/CLAUDE.md` | Complete |
| Output directory | `analysis/conversation/` | Complete — with README.md |
| State exclusion | `.gitignore` | Complete — `.github/state/` ignored |

### The Gap

**The conversation workflow agents exist only in Copilot Chat format (`.github/agents/`).** Claude Code discovers agents by scanning the `agents/` directory — it cannot invoke `.github/agents/` files. This means:

- `/analyze` in Claude Code cannot dispatch to `conversation-moderator` as a sub-agent
- COMPLEX_NEW routing in the orchestrator fails silently — Claude Code looks for `agents/conversation-moderator.md` which doesn't exist
- The entire conversation pipeline is unreachable from Claude Code

**Additionally**, several recommended improvements from the integration audit remain unimplemented: tooling automation, quality infrastructure, and feature extensions.

---

## 2. Phase 1 — Claude Code Agent Port (Critical)

> **Priority:** P0 — Blocking
> **Effort:** ~2 hours
> **Prerequisite:** None

### 1.1 Create `agents/conversation-moderator.md`

**Source:** `.github/agents/conversation-moderator.md` (540 lines)

**Action:** Copy the entire file. Replace ONLY the frontmatter block (lines 1–15).

Original Copilot frontmatter:
```yaml
---
description: "Orchestrates conversation analysis workflows by spawning persona sub-agents in sequence, managing context isolation between phases, and producing a final convergence synthesis."
tools:
  - agent
  - read
  - search/codebase
  - search/textSearch
  - search/fileSearch
agents:
  - conversation-persona
  - conversation-persona-architect
  - conversation-persona-pragmatist
  - conversation-persona-critic
  - conversation-persona-researcher
---
```

New Claude Code frontmatter:
```yaml
---
name: conversation-moderator
description: "Orchestrates conversation analysis workflows by spawning persona sub-agents in sequence, managing context isolation between phases, and producing a final convergence synthesis."
model: claude-opus-4-5
tools:
  - Task
  - Read
  - Write
  - Bash
---
```

**Key decisions:**
- `claude-opus-4-5` for the moderator — it performs the most cognitively demanding work (multi-phase management, convergence synthesis, 6-dimension quality scoring)
- `Task` tool replaces the `agents:` array — Claude Code spawns sub-agents by name via Task
- `Write` explicitly declared — Claude Code requires it (Copilot doesn't)
- Body copied verbatim — all phase behavior, state management, context rules, quality rubric are platform-agnostic

### 1.2 Create 5 Persona Agent Files

Each persona follows the same pattern: copy from `.github/agents/`, replace frontmatter only.

| File to Create | Source | Model | Tools | Special Notes |
|---|---|---|---|---|
| `agents/conversation-persona.md` | `.github/agents/conversation-persona.md` | `claude-sonnet-4-5` | Read, Write, Bash | Generic persona — accepts runtime config |
| `agents/conversation-persona-architect.md` | `.github/agents/conversation-persona-architect.md` | `claude-sonnet-4-5` | Read, Write, Bash | Systems thinking specialist |
| `agents/conversation-persona-pragmatist.md` | `.github/agents/conversation-persona-pragmatist.md` | `claude-sonnet-4-5` | Read, Write, Bash | Shipping speed specialist |
| `agents/conversation-persona-critic.md` | `.github/agents/conversation-persona-critic.md` | `claude-sonnet-4-5` | Read, Write, Bash | Devil's advocate specialist |
| `agents/conversation-persona-researcher.md` | `.github/agents/conversation-persona-researcher.md` | `claude-sonnet-4-5` | Read, Write, Bash, WebFetch | Evidence specialist — `fetch` → `WebFetch` |

**Frontmatter template for persona agents:**
```yaml
---
name: {filename-without-extension}
description: "{copy from .github/agents/ source}"
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

Add `WebFetch` to the researcher's tools list only. If `WebFetch` is unavailable (MCP not configured), the researcher can fall back to `Bash` with `curl`.

### 1.3 Add Dispatch Instruction to `commands/analyze.md`

**Insert immediately after the frontmatter closing `---`, before the `# /analyze` heading:**

```markdown
> **Execution:** Dispatch this entire request to the `conversation-moderator` sub-agent.
> Pass the full message content (topic, mode, and any file paths) as the sub-agent's input.
> The moderator will read `conversation/config/*.yml` and orchestrate all persona sub-agents.
> Do NOT execute the conversation workflow inline — always delegate to the sub-agent.
```

**Why this matters:** Without this instruction, Claude Code may execute the analysis inline in the current context rather than spawning the moderator as a sub-agent. This would bypass context isolation between personas and collapse the entire multi-agent architecture into a single token window.

### 1.4 End-to-End Verification

After creating all 7 files (6 agents + analyze.md update):

| Test | Command | Expected Result |
|------|---------|-----------------|
| Standalone analysis | `/analyze "GraphQL vs REST for API layer"` | Moderator spawns 4 personas via Task, convergence saved to `analysis/conversation/` |
| COMPLEX_NEW routing | `/workflow "analyze: design a notification system with 3 delivery channels"` | Orchestrator classifies as COMPLEX_NEW, dispatches moderator first, Gate 0 validates, analyst reads convergence |
| Copilot regression | `@conversation-moderator` in VS Code chat | Routes to `.github/agents/` (unchanged) |
| State file creation | After any `/analyze` run | `.github/state/conversation-state.yml` + phase output files created |

---

## 3. Phase 2 — Tooling & Automation

> **Priority:** P1 — High
> **Effort:** ~4 hours
> **Prerequisite:** Phase 1 complete

### 2.1 Agent Body Sync Script

**File:** `scripts/sync-conversation-agents.sh`

The instruction body of each conversation agent is identical across both platforms. Only frontmatter differs. This script automates synchronization:

```
For each .github/agents/conversation-*.md:
  1. Read the source file
  2. Extract the body (everything after the second ---)
  3. Read the Claude Code version's frontmatter from agents/conversation-*.md
  4. Combine: Claude Code frontmatter + source body
  5. Write to agents/conversation-*.md
  6. Report: "Synced {filename}: {N} lines from .github/agents/"
```

**Trigger:** Run manually after any conversation agent body update, or as a pre-commit hook.

**Primary source convention:** `.github/agents/` is always the authoritative source for instruction body content. `agents/` files are derived. This prevents drift.

### 2.2 Integration Validation Script

**File:** `scripts/validate-integration.sh`

Automated cross-reference checker:

```
Checks:
  1. Every path referenced in agents/*.md exists on disk
  2. Every path referenced in .github/agents/*.md exists on disk
  3. No stale "panel-module/" references anywhere
  4. Every agent declared in orchestrator chains has a corresponding .md file
  5. CLAUDE.md resource table entries all resolve to existing paths
  6. conversation/config/*.yml files are valid YAML (syntax check)
  7. Frontmatter in agents/ files: no Copilot tool names (readFile, codebase, textSearch, fileSearch, fetch)
  8. Frontmatter in .github/agents/ files: no Claude Code tool names (Task, Read, Write, Bash)
  9. Body diff between platform pairs: flag any divergence
```

**Output:** Pass/fail per check, with specific file:line for failures.

### 2.3 YAML Schema Validation

**File:** `scripts/validate-config.sh` (or integrated into 2.2)

Validates `conversation/config/*.yml` files against expected structure:

| Config File | Required Keys | Validation Rules |
|---|---|---|
| `conversation.yml` | phases, routing, context, termination, moderation | All phase IDs referenced in routing must exist in phases |
| `personas.yml` | specialists, presets, custom | Each specialist must have id, name, description, perspective, priorities |
| `quality.yml` | dimensions, thresholds, evaluator-optimizer | Each dimension must have name, weight (0-1), criteria |
| `extensions.yml` | skills, protocols, approval-gates, delegation | Each skill must reference a valid path in `conversation/skills/` |

Catches config errors before they surface at runtime (e.g., invalid phase names, missing thresholds, orphaned skill references).

---

## 4. Phase 3 — Documentation Consolidation

> **Priority:** P1 — High
> **Effort:** ~3 hours
> **Prerequisite:** Phase 1 complete

### 3.1 Top-Level Architecture Overview

**File:** `ARCHITECTURE.md` (project root)

A concise, single-document overview of the dual-platform system:

| Section | Content |
|---------|---------|
| System Overview | What the framework does, the two-workflow model |
| Agent Registry | All agents across both platforms with their roles |
| Workflow Types | NEW_PROJECT, BUG_FIX, ENHANCEMENT, REFACTOR, COMPLEX_NEW with chains |
| Quality Gates | Gate 0, Micro-Gate A/B, Deterministic Gate |
| Dual-Platform Model | How Claude Code and Copilot Chat agents coexist |
| Data Flow Diagram | ASCII/mermaid showing conversation → dev pipeline flow |
| Directory Map | Full tree with purpose annotations |
| Configuration | Where configs live and what they control |

Currently this information is spread across `arch-framework.md`, `arch-blueprint.md`, and `imp-architecture-v2.md`. The new file provides a single entry point.

### 3.2 Update Root `CLAUDE.md`

Add entries for the new Claude Code conversation agents:

```markdown
| `agents/conversation-moderator.md` | Conversation analysis — spawns personas, manages phases, produces convergence (Claude Code) |
| `agents/conversation-persona*.md` | Specialist personas for dialectical analysis (Claude Code) |
```

Update the existing `.github/agents/` entry to note it's the Copilot Chat version.

### 3.3 Update `conversation/README.md`

Add a "Platform Entry Points" section:

| Platform | Entry Point | Agent Location |
|----------|-------------|----------------|
| Claude Code | `/analyze` or COMPLEX_NEW routing | `agents/conversation-*.md` |
| Copilot Chat | `@conversation-moderator` or `@analyze` prompt | `.github/agents/conversation-*.md` |

Both platforms share: `conversation/config/`, `conversation/protocols/`, `conversation/skills/`, `.github/state/`, `analysis/conversation/`.

---

## 5. Phase 4 — Quality & Testing Infrastructure

> **Priority:** P2 — Medium
> **Effort:** ~6 hours
> **Prerequisite:** Phase 1 verified end-to-end

### 4.1 Convergence Quality Regression Tests

Create a test harness that validates convergence analysis output structure:

| Check | Assertion |
|-------|-----------|
| Executive Summary | Present, non-empty, contains architectural recommendation |
| Evidence Audit | Present, contains at least 1 classified claim |
| Decision Matrix | Present, ≥ 3 options scored, weighted criteria |
| Recommendations | ≥ 3 recommendations, each with confidence %, risk statement, falsifiability condition |
| Quality Rubric | 6 dimensions scored, overall score calculated |
| Output path | File exists at `analysis/conversation/{descriptor}-convergence.md` |

**Implementation:** Shell script or lightweight test framework that reads the convergence output and validates against the schema.

### 4.2 Gate 0 Dashboard

Track convergence quality scores across project iterations:

**File:** `analysis/conversation/quality-log.yml`

```yaml
entries:
  - date: "2026-02-25"
    topic: "GraphQL vs REST"
    descriptor: "api-layer-evaluation"
    overall_score: 3.8
    dimensions:
      perspective_diversity: 4.0
      evidence_quality: 3.5
      concession_depth: 3.8
      challenge_substantiveness: 4.0
      synthesis_quality: 3.8
      actionability: 3.7
    gate_0_pass: true
    personas: [architect, pragmatist, critic, researcher]
```

Provides visibility into analysis quality trends. Low scores across multiple analyses may indicate persona configuration issues (personas.yml) or phase timing problems (conversation.yml).

---

## 6. Phase 5 — Feature Extensions

> **Priority:** P3 — Low (incremental)
> **Effort:** Variable
> **Prerequisite:** Phases 1–4 complete and stable

### 5.1 Conversation History Persistence

**Directory:** `analysis/conversation/history/`

Currently the conversation state in `.github/state/` is ephemeral (gitignored). Add an option to persist full session transcripts for post-mortem and learning:

```
analysis/conversation/history/
└── {session-id}/
    ├── session-metadata.yml       ← Topic, personas, dates, quality score
    ├── phase-2-positions/         ← All position papers
    ├── phase-3-challenges/        ← All challenge documents
    ├── phase-4-rebuttals/         ← All rebuttals
    └── convergence.md             ← Final output (copy of the main convergence file)
```

Controlled by a `persist_history: true` flag in `conversation/config/conversation.yml`.

### 5.2 Custom Persona Presets from Specialists

Allow projects to define `specialists/*.md` files that auto-register as conversation personas (not just dev agents):

```yaml
# In specialists/domain-expert.md frontmatter
---
name: domain-expert
type: conversation-persona    # Signals this is a conversation addition
perspective: "Domain knowledge holder — specific industry or problem domain"
priorities: [domain accuracy, business rule enforcement, real-world constraints]
---
```

The moderator would scan `specialists/` at startup and include any `type: conversation-persona` files in the persona selection pool alongside the standard 4 specialists.

### 5.3 Convergence Diff

When re-running `/analyze` on a topic that already has a convergence analysis:

1. Detect existing `analysis/conversation/{topic}-convergence.md`
2. Run the new analysis normally
3. After Phase 6, produce a diff section showing:
   - Recommendations that changed (position shifts)
   - New evidence introduced
   - Decision Matrix score changes
   - Consensus points that held vs. shifted

Useful for iterative design exploration where the design space evolves over time.

### 5.4 State Collision Prevention

Both platform versions write to `.github/state/`. If both run simultaneously (unlikely but possible), state files collide.

**Solution:** Namespace state by session ID:
```
.github/state/{session-id}/
├── conversation-state.yml
├── topic-brief.md
├── phase-2/
└── ...
```

Session ID = `{topic-slug}-{YYYY-MM-DD}-{platform}` (e.g., `graphql-vs-rest-2026-02-25-claude`).

Requires a minor update to the moderator's state management section in both platform versions.

---

## 7. Platform Coexistence Model

### Architecture

```
┌─────────────────────┐    ┌─────────────────────┐
│   Copilot Chat       │    │    Claude Code        │
│   (VS Code)          │    │    (Terminal/IDE)      │
├─────────────────────┤    ├─────────────────────┤
│ @conversation-mod.   │    │ /analyze command       │
│ @analyze prompt      │    │ COMPLEX_NEW routing    │
├─────────────────────┤    ├─────────────────────┤
│ .github/agents/      │    │ agents/                │
│   conversation-*.md  │    │   conversation-*.md    │
│   (Copilot format)   │    │   (Claude format)      │
└────────┬────────────┘    └────────┬────────────┘
         │                          │
         └──────────┬───────────────┘
                    │
         ┌──────────▼───────────────┐
         │   Shared Runtime Layer    │
         ├──────────────────────────┤
         │ conversation/config/      │  ← 4 YAML config files
         │ conversation/protocols/   │  ← 4 protocol modules
         │ conversation/skills/      │  ← 5 skill modules
         │ conventions/shared.md     │  ← Cross-workflow standards
         │ .github/state/            │  ← Session state (ephemeral)
         │ analysis/conversation/    │  ← Convergence output (persistent)
         └──────────────────────────┘
```

### File Ownership Convention

| Directory | Owner | Updated By |
|-----------|-------|------------|
| `.github/agents/` | Copilot Chat platform | Manual edit (primary source for body content) |
| `agents/conversation-*.md` | Claude Code platform | Sync script from `.github/agents/` |
| `conversation/config/` | Both platforms | Manual edit (shared) |
| `conversation/protocols/` | Both platforms | Manual edit (shared) |
| `conversation/skills/` | Both platforms | Manual edit (shared) |
| `.github/state/` | Whichever platform is running | Auto-generated (gitignored) |
| `analysis/conversation/` | Whichever platform produces output | Auto-generated (version controlled) |

### Tool Name Mapping

| Copilot Chat | Claude Code | Function |
|---|---|---|
| `agent` | `Task` | Sub-agent spawning |
| `read` / `readFile` | `Read` | File reading |
| `codebase` | `Bash` (grep) | Semantic codebase search |
| `textSearch` / `search/textSearch` | `Bash` (grep) | Text pattern search |
| `fileSearch` / `search/fileSearch` | `Bash` (find) | File path search |
| `fetch` | `WebFetch` | HTTP requests |
| _(implicit)_ | `Write` | File writing |

---

## 8. Known Constraints & Mitigations

| # | Constraint | Impact | Mitigation | Risk |
|---|-----------|--------|------------|------|
| C1 | **Body duplication** — instruction body exists in 2 files per agent | Edit one, forget the other → drift | `.github/agents/` is primary source; `scripts/sync-conversation-agents.sh` propagates changes | Low — body changes infrequently |
| C2 | **WebFetch availability** — Researcher needs HTTP in Claude Code | Can't fetch external references if MCP not configured | `WebFetch` declared (silently ignored if unavailable); Researcher falls back to `Bash` + `curl`; Evidence Registry's `[Unsourced assertion]` handles unfetchable sources | Low |
| C3 | **Inline execution fallback** — Claude Code may execute /analyze inline | Loses context isolation, all personas in one token window, quality drops | Dispatch instruction in `commands/analyze.md` explicitly prevents this; moderator body has degraded "Inline Phase Execution" fallback | Medium |
| C4 | **State collision** — both platforms write to `.github/state/` | Concurrent sessions overwrite each other | Session ID namespace (Phase 5.4); in practice, concurrent multi-platform sessions are extremely unlikely | Very Low |
| C5 | **Model availability** — frontmatter specifies exact model IDs | Model names may change across releases | Use identifiers consistent with existing `agents/orchestrator.md`; update all frontmatter in batch when models change | Low |

---

## 9. Verification Protocol

### After Phase 1

| # | Check | Command/Action | Expected |
|---|-------|---------------|----------|
| V1 | Agent files exist | `ls agents/conversation-*.md` | 6 files |
| V2 | No Copilot tool names in agents/ | `grep -l 'readFile\|codebase\|textSearch\|fileSearch' agents/conversation-*.md` | Empty (no matches) |
| V3 | No Claude tool names in .github/agents/ | `grep -l '\bTask\b\|\bRead\b\|\bWrite\b\|\bBash\b' .github/agents/conversation-*.md` | Empty (no matches) |
| V4 | Frontmatter valid | Each file: `name:` matches filename, `model:` is valid, `tools:` uses platform names | Pass |
| V5 | Body parity | `diff <(sed '1,/^---$/d' .github/agents/conversation-moderator.md) <(sed '1,/^---$/d' agents/conversation-moderator.md)` | No differences |
| V6 | Dispatch instruction | `grep 'Dispatch this entire request' commands/analyze.md` | Match found |
| V7 | Standalone /analyze | `/analyze "test topic"` in Claude Code | Moderator invoked, convergence file created |
| V8 | COMPLEX_NEW routing | `/workflow "analyze: test topic"` in Claude Code | COMPLEX_NEW classified, moderator dispatched, Gate 0 checked |
| V9 | Copilot regression | `@conversation-moderator` in VS Code Chat | Routes to `.github/agents/` (unchanged) |

### After Phase 2

| # | Check | Command/Action | Expected |
|---|-------|---------------|----------|
| V10 | Sync script | `bash scripts/sync-conversation-agents.sh` | All 6 agents synced, zero diff |
| V11 | Validation script | `bash scripts/validate-integration.sh` | All checks pass |
| V12 | YAML validation | `bash scripts/validate-config.sh` | All 4 config files valid |

---

## 10. Timeline & Priority Matrix

```
Week 1                    Week 2                   Week 3                   Month 2+
┌───────────────────┐    ┌───────────────────┐    ┌───────────────────┐    ┌─────────────────┐
│ PHASE 1 (P0)      │    │ PHASE 2 (P1)      │    │ PHASE 4 (P2)      │    │ PHASE 5 (P3)    │
│ Claude Code Port   │    │ Sync script        │    │ Regression tests   │    │ History          │
│ • 6 agent files    │    │ Validation script  │    │ Gate 0 dashboard   │    │ Custom personas  │
│ • Dispatch instr.  │    │ YAML validation    │    │                   │    │ Convergence diff │
│ • E2E verification │    │                   │    │                   │    │ State namespacing│
│                   │    │ PHASE 3 (P1)      │    │                   │    │                 │
│                   │    │ ARCHITECTURE.md    │    │                   │    │                 │
│                   │    │ CLAUDE.md update   │    │                   │    │                 │
│                   │    │ README.md update   │    │                   │    │                 │
└───────────────────┘    └───────────────────┘    └───────────────────┘    └─────────────────┘
     BLOCKING                HIGH                     MEDIUM                   INCREMENTAL
```

### Decision Log

| Decision | Rationale | Date |
|----------|-----------|------|
| `.github/agents/` is primary source for body content | Copilot format was written first; Claude Code files are derived | 2026-02-25 |
| Moderator gets `claude-opus-4-5`, personas get `claude-sonnet-4-5` | Moderator work (multi-phase orchestration, synthesis, quality scoring) is more demanding | 2026-02-25 |
| Dispatch instruction over framework-level routing | Simpler, no Claude Code internals dependency, explicit behavioral contract | 2026-02-25 |
| Session state in `.github/state/` (gitignored) | Ephemeral by nature — convergence output in `analysis/conversation/` is what persists | 2026-02-25 |
| Parallel files over symlinks or dual-frontmatter | Symlinks break because frontmatter schemas are incompatible; dual-frontmatter is unsupported | 2026-02-25 |

---

*This plan incorporates all findings from the conversation-framework-claude-integration.md analysis, the previous integration audit (all gap items G1–G10), and the recommended actions from the cross-reference verification sessions.*
