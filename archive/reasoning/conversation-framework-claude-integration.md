# Conversation Framework → Claude Code Integration

> **⚠️ ARCHIVED — Integration Complete**
> All steps in this plan have been executed. Agent files now live in `agents/personas/conversation-*.md`.
> For the current architecture, see [`ARCHITECTURE.md`](../../ARCHITECTURE.md).

> **Document type:** Integration plan and implementation guide (historical)
> **Scope:** Porting the GitHub Copilot Chat conversation workflow to run natively in Claude Code
> **Status:** ✅ Complete — archived
> **Output target:** Originally `agents/conversation-moderator.md` + 5 persona agents → now `agents/personas/`

---

## Table of Contents

1. [Context & Background](#1-context--background)
2. [System Architecture Overview](#2-system-architecture-overview)
3. [The Gap: Why Copilot Agents Don't Work in Claude Code](#3-the-gap-why-copilot-agents-dont-work-in-claude-code)
4. [Platform Comparison](#4-platform-comparison)
5. [Integration Strategy](#5-integration-strategy)
6. [Step-by-Step Implementation Plan](#6-step-by-step-implementation-plan)
   - [Step 1 — Port `conversation-moderator.md`](#step-1--port-conversation-moderatormd)
   - [Step 2 — Port the 5 Persona Agents](#step-2--port-the-5-persona-agents)
   - [Step 3 — Update the `/analyze` Command](#step-3--update-the-analyze-command)
   - [Step 4 — Verify State Path Alignment](#step-4--verify-state-path-alignment)
   - [Step 5 — Verify Orchestrator Integration](#step-5--verify-orchestrator-integration)
7. [Frontmatter Adaptation Reference](#7-frontmatter-adaptation-reference)
8. [Tool Name Mapping](#8-tool-name-mapping)
9. [File-by-File Creation Guide](#9-file-by-file-creation-guide)
10. [State Management Alignment](#10-state-management-alignment)
11. [Dual-Platform Coexistence Model](#11-dual-platform-coexistence-model)
12. [Cross-System Data Flow](#12-cross-system-data-flow)
13. [Verification Checklist](#13-verification-checklist)
14. [Known Constraints & Trade-offs](#14-known-constraints--trade-offs)

---

## 1. Context & Background

The `panel-module` implements a **two-workflow framework**:

| Workflow | Purpose | Platform | Location |
|----------|---------|----------|----------|
| **Dev Workflow** | Software development pipeline (analyse → design → implement → test → review) | Claude Code | `agents/` + `commands/` |
| **Conversation Workflow** | Structured multi-persona dialectic that produces evidence-grounded architectural recommendations | GitHub Copilot Chat | `.github/agents/` + `.github/prompts/` |

Both workflows are designed to integrate: the dev orchestrator automatically invokes the conversation moderator for `COMPLEX_NEW` requests (architectural uncertainty detected) before dispatching the development agent chain. However, as of this writing, the conversation workflow agents exist **only** in Copilot Chat format — meaning that integration path is broken when running purely inside Claude Code.

This document describes the complete plan to port the conversation workflow to Claude Code so that both entry points work natively:

- **Copilot Chat entry:** `@conversation-moderator` → `.github/agents/conversation-moderator.md`
- **Claude Code entry:** `/analyze` command or `COMPLEX_NEW` orchestrator routing → `agents/conversation-moderator.md`

The two runtime paths share the same configuration files (`conversation/config/*.yml`), the same output directory (`analysis/conversation/`), and the same conventions (`conventions/shared.md`). The only difference is the platform-specific agent format.

---

## 2. System Architecture Overview

```
panel-module/
│
├── agents/                          ← Claude Code dev workflow agents (COMPLETE)
│   ├── orchestrator.md              ← Routes COMPLEX_NEW → conversation-moderator
│   ├── analyst.md
│   ├── architect.md
│   ├── developer.md
│   ├── tester.md
│   ├── reviewer.md
│   ├── discovery.md
│   │
│   ├── conversation-moderator.md    ← MISSING: needs to be created (Claude Code port)
│   ├── conversation-persona.md      ← MISSING: needs to be created (Claude Code port)
│   ├── conversation-persona-architect.md    ← MISSING
│   ├── conversation-persona-pragmatist.md   ← MISSING
│   ├── conversation-persona-critic.md       ← MISSING
│   └── conversation-persona-researcher.md   ← MISSING
│
├── .github/
│   ├── agents/                      ← Copilot Chat conversation workflow (COMPLETE)
│   │   ├── conversation-moderator.md
│   │   ├── conversation-persona.md
│   │   ├── conversation-persona-architect.md
│   │   ├── conversation-persona-pragmatist.md
│   │   ├── conversation-persona-critic.md
│   │   └── conversation-persona-researcher.md
│   ├── prompts/                     ← Copilot Chat prompts (COMPLETE)
│   │   ├── analyze.prompt.md
│   │   └── analyze-documents.prompt.md
│   └── state/                       ← Shared session state (used by both platforms)
│       └── .gitkeep
│
├── commands/                        ← Claude Code slash commands
│   ├── workflow.md                  ← /workflow — dev pipeline trigger
│   ├── discover.md                  ← /discover — codebase exploration
│   └── analyze.md                   ← /analyze — NEEDS UPDATE: explicit moderator dispatch
│
├── conversation/                    ← Shared conversation workflow runtime (used by BOTH platforms)
│   ├── config/
│   │   ├── conversation.yml         ← Phases, routing, context isolation, termination
│   │   ├── personas.yml             ← Persona library, presets, variants, selection guide
│   │   ├── quality.yml              ← Quality rubric, evaluator-optimizer
│   │   └── extensions.yml           ← Skills, protocols, approval gates, delegation
│   ├── protocols/                   ← Moderator protocol modules (loaded at runtime)
│   └── skills/                      ← Persona skill modules (injected by moderator)
│
├── conventions/                     ← Shared standards (both workflows)
│   └── shared.md                    ← Evidence, severity, scope, reasoning standards
│
└── analysis/conversation/           ← Convergence analysis output (shared by both platforms)
```

### Data Flow Summary

```
[Claude Code]                            [Copilot Chat]
     │                                        │
/analyze command                    @conversation-moderator
or COMPLEX_NEW routing                         │
     │                                         │
     ▼                                         ▼
agents/conversation-moderator.md    .github/agents/conversation-moderator.md
     │                                         │
     ├── reads ──────────────────────── conversation/config/*.yml
     ├── reads ──────────────────────── conversation/protocols/*.md
     ├── reads ──────────────────────── conversation/skills/*.md
     ├── reads ──────────────────────── conventions/shared.md
     ├── writes ─────────────────────── .github/state/conversation-state.yml
     │                                  .github/state/phase-*/*.md
     └── writes ─────────────────────── analysis/conversation/*-convergence.md
                                                │
                                   [Dev Workflow reads this]
                               agents/orchestrator.md (Gate 0)
                               agents/analyst.md (prior context)
                               agents/architect.md (prior context)
```

---

## 3. The Gap: Why Copilot Agents Don't Work in Claude Code

GitHub Copilot Chat and Claude Code use **different agent runtime contracts**. A `.github/agents/*.md` file cannot be directly invoked by Claude Code as a sub-agent, and vice versa.

### Contract Incompatibilities

**1. Agent Discovery Location**

Claude Code discovers sub-agents by scanning the `agents/` directory at the project root. It does **not** scan `.github/agents/`. When the orchestrator's `COMPLEX_NEW` chain calls `conversation-moderator`, Claude Code looks for `agents/conversation-moderator.md` — which does not exist.

**2. Frontmatter Schema**

Each platform uses a different frontmatter schema for declaring the agent's identity, model, and tools:

```yaml
# Copilot Chat format (.github/agents/)
---
description: "..."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
  - fileSearch
agents:
  - conversation-persona
  - conversation-persona-architect
---
```

```yaml
# Claude Code format (agents/)
---
name: conversation-moderator
description: "..."
model: claude-sonnet-4-5
tools:
  - Task
  - Read
  - Write
  - Bash
---
```

Key differences:
- `model` uses platform-specific identifiers (`Claude Sonnet 4 (copilot)` vs `claude-sonnet-4-5`)
- Tool names are different (Copilot uses camelCase API names; Claude Code uses capability names)
- Copilot declares an `agents:` array listing sub-agents by name; Claude Code discovers sub-agents by file presence in `agents/` and invokes them via the `Task` tool

**3. Sub-Agent Invocation Model**

- **Copilot Chat:** The moderator spawns persona sub-agents by referencing agent names declared in its `agents:` frontmatter array. The Copilot runtime resolves these names to `.github/agents/*.md` files.
- **Claude Code:** Sub-agents are spawned using the `Task` tool (or equivalent `agent` invocation). The runtime resolves agent names by scanning `agents/*.md` files and matching the `name:` frontmatter field.

**4. Tool API Surface**

Copilot Chat tools (`readFile`, `codebase`, `textSearch`, `fileSearch`, `fetch`) map to different tool names in Claude Code (`Read`, `Write`, `Bash`, `WebFetch`). An agent file referencing `readFile` in its tools list will not have that tool available when run inside Claude Code.

**5. Prompt Entry Points**

Copilot Chat prompts live in `.github/prompts/*.prompt.md` and support `${input:name:description}` variable interpolation. Claude Code slash commands live in `commands/*.md` and receive free-form message content — there is no built-in variable interpolation syntax.

### What This Means in Practice

When a user runs `/workflow` with a `COMPLEX_NEW` request in Claude Code:

1. The orchestrator correctly identifies the request as `COMPLEX_NEW`
2. It routes to `conversation-moderator` as the first agent in the chain
3. Claude Code looks for `agents/conversation-moderator.md` — **not found**
4. The workflow fails silently or errors, depending on Claude Code's error handling
5. The `COMPLEX_NEW` chain never executes; the architectural analysis is skipped

When a user runs `/analyze` directly in Claude Code:

1. The command describes what should happen
2. It mentions invoking `conversation-moderator`
3. Claude Code cannot resolve this to an agent file
4. Claude Code attempts to execute the analysis inline in the current context instead
5. The multi-agent sub-agent architecture is lost; all personas run in the same token window

---

## 4. Platform Comparison

### Comprehensive Feature Matrix

| Feature | GitHub Copilot Chat | Claude Code | Integration Notes |
|---------|--------------------|-----------|--------------------|
| **Agent file location** | `.github/agents/` | `agents/` | Create parallel files in `agents/` |
| **Agent discovery** | Scan `.github/agents/` | Scan `agents/` by `name:` field | Both directories can coexist |
| **Invocation trigger** | `@agent-name` in chat UI | `/command` or orchestrator routing | Keep both entry points |
| **Sub-agent spawning** | `agents:` array in frontmatter | `Task` tool in instructions | Replace `agents:` array with `Task` calls |
| **Model identifier** | `Claude Sonnet 4 (copilot)` | `claude-sonnet-4-5` | Update per file |
| **Model for moderator** | `Claude Sonnet 4 (copilot)` | `claude-opus-4-5` (recommended) | Moderator benefits from more capacity |
| **File read tool** | `readFile` | `Read` | Rename in frontmatter |
| **Codebase search** | `codebase` | `Bash` (grep) | Rename in frontmatter |
| **Text search** | `textSearch` | `Bash` (grep) | Rename in frontmatter |
| **File search** | `fileSearch` | `Bash` (find) | Rename in frontmatter |
| **Web fetch** | `fetch` | `WebFetch` or `Bash` (curl) | Rename in frontmatter (Researcher only) |
| **File write** | Not declared (implicit) | `Write` | Add to frontmatter |
| **Prompt entry point** | `.github/prompts/*.prompt.md` | `commands/*.md` | Keep both; update `commands/analyze.md` |
| **Input variable syntax** | `${input:name:description}` | Free-form message content | No change needed to commands |
| **State file writes** | `.github/state/` | `.github/state/` | Same path — no change needed |
| **Config files** | `conversation/config/*.yml` | `conversation/config/*.yml` | Same path — no change needed |
| **Output directory** | `analysis/conversation/` | `analysis/conversation/` | Same path — no change needed |
| **Agent instructions** | Platform-agnostic markdown | Platform-agnostic markdown | Copy verbatim — zero changes |

### What Changes vs. What Stays the Same

**Changes (frontmatter only):**
- `model:` value
- `tools:` list (names change, not capabilities)
- Remove `agents:` array
- Add `name:` field (required by Claude Code for agent resolution)

**Stays identical (instruction body — copy verbatim):**
- All phase behavior descriptions (Phase 2, 3, 4, 5)
- Output format specifications
- Evidence Registry format
- Critical Rules
- Delegation request format
- Persona identity and bias disclosures
- Challenge construction tables
- Confidence verification rules
- Scope discipline and intent marker rules

The instruction body is **100% platform-agnostic** because it describes what to think and how to structure outputs — not how to call platform APIs.

---

## 5. Integration Strategy

### Core Principle: Parallel Port, Not Migration

The goal is **not** to replace the Copilot Chat agents — it is to create a second set of agent files in `agents/` that Claude Code can discover and invoke. Both sets of files will:

- Share the same instruction body (copy verbatim)
- Read the same config files (`conversation/config/*.yml`)
- Write to the same state directory (`.github/state/`)
- Write to the same output directory (`analysis/conversation/`)

The only difference is the frontmatter (platform-specific runtime declarations).

### Why Not Symlinks or Includes?

- **Symlinks** would create a single file that both platforms try to parse — but the frontmatter schema is incompatible. A file with `model: Claude Sonnet 4 (copilot)` will be rejected or misinterpreted by Claude Code's model resolution.
- **Includes/imports** are not supported by either platform's agent format.
- **Dual frontmatter** (trying to satisfy both schemas in one file) is not a supported pattern.

The cleanest, most maintainable solution is **two sets of files** with the same body and platform-appropriate frontmatter. When the instruction body needs to change, update both files (or automate the sync with a script).

### Scope of Changes

| Area | Action | Effort |
|------|--------|--------|
| `agents/conversation-moderator.md` | Create (port from `.github/agents/`) | Medium — frontmatter + minor tool-ref edits |
| `agents/conversation-persona.md` | Create (port from `.github/agents/`) | Low — frontmatter only |
| `agents/conversation-persona-architect.md` | Create (port from `.github/agents/`) | Low — frontmatter only |
| `agents/conversation-persona-pragmatist.md` | Create (port from `.github/agents/`) | Low — frontmatter only |
| `agents/conversation-persona-critic.md` | Create (port from `.github/agents/`) | Low — frontmatter only |
| `agents/conversation-persona-researcher.md` | Create (port from `.github/agents/`) | Low — frontmatter + `fetch` → `WebFetch` |
| `commands/analyze.md` | Update — add explicit moderator dispatch header | Very low — 3 lines |
| `agents/orchestrator.md` | No change needed | — |
| `conversation/config/*.yml` | No change needed | — |
| `.github/agents/*.md` | No change needed (kept for Copilot Chat) | — |
| `.github/prompts/*.md` | No change needed (kept for Copilot Chat) | — |

---

## 6. Step-by-Step Implementation Plan

### Step 1 — Port `conversation-moderator.md`

**Source:** `.github/agents/conversation-moderator.md`
**Destination:** `agents/conversation-moderator.md`

#### 1.1 Replace the Frontmatter

The original Copilot frontmatter:

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

Replace with the Claude Code frontmatter:

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

**Why `claude-opus-4-5` for the moderator?**
The moderator performs the most cognitively demanding work in the system: reading 4 YAML config files at startup, managing multi-phase context isolation, running diversity audits, synthesizing 4-5 full position papers into a convergence analysis, and scoring on a 6-dimension quality rubric. This is the correct agent to allocate more model capacity to.

The persona agents do focused single-phase work (generate one position paper, or cross-examine one set of documents) and are well-suited to `claude-sonnet-4-5`.

#### 1.2 Remove the `agents:` Array

The `agents:` frontmatter array is a Copilot Chat-specific declaration that tells the Copilot runtime which sub-agents this agent can invoke. Claude Code does not use this field — it discovers available sub-agents by scanning `agents/*.md` files and matches them via the `name:` field.

**Remove the entire `agents:` block.** The moderator's instruction body already contains the sub-agent invocation logic (it calls personas by name using the `Task` tool). Once the persona agent files exist in `agents/`, they will be automatically discoverable.

#### 1.3 Update Tool References in the Body

Scan the instruction body for any inline references to Copilot tool names and update them:

| Find | Replace |
|------|---------|
| `readFile` (in prose/instructions) | `Read` |
| `fileSearch` (in prose/instructions) | `Bash` |
| `textSearch` (in prose/instructions) | `Bash` |

The instruction body of the moderator primarily references these tools when describing how to load config files and state files. The core workflow logic (phase management, context isolation, synthesis) is tool-agnostic.

**Important:** The `context_rule` token resolution table in the moderator body references `.github/state/` file paths — these are filesystem paths, not tool names. Do NOT change these.

#### 1.4 Copy the Entire Body Verbatim

Everything from the `> **Runtime:** Yes` line onward copies without modification:

- Configuration loading instructions
- Protocol loading (v2) — all 4 protocols
- State management (v2) — session start, phase boundaries, output storage, session end
- Context rule enforcement (v2) — token resolution table, audit trail
- Phase routing, evaluator-optimizer, approval gates (all reference external protocol files)
- Dynamic speaker selection
- Mediated delegation
- Input modes (Mode A, B, C)
- Inline phase execution instructions
- Full workflow execution (Steps 1-6, including all sub-steps)
- Conversation quality rubric
- Deliverable verification protocol (gates 2→3, 3→4, 4→5)
- Critical rules

**Nothing in the body is platform-specific.** The body describes intent, not API calls.

---

### Step 2 — Port the 5 Persona Agents

Each of the 5 persona files requires the same transformation: **replace the frontmatter, copy the body verbatim.**

#### 2.1 `agents/conversation-persona.md`

**Source:** `.github/agents/conversation-persona.md`

**Original frontmatter:**
```yaml
---
description: "A configurable conversation participant that adopts a specified persona..."
tools:
  - readFile
  - codebase
  - textSearch
---
```

**New frontmatter:**
```yaml
---
name: conversation-persona
description: "A configurable conversation participant that adopts a specified persona, perspective, and priorities to generate position papers, challenges, and rebuttals in a structured conversation analysis workflow."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

#### 2.2 `agents/conversation-persona-architect.md`

**Source:** `.github/agents/conversation-persona-architect.md`

**Original frontmatter:**
```yaml
---
description: "Systems-thinking conversation persona focused on architecture, scalability, composability, and long-term technical sustainability. Optimizes for elegant abstractions and extensible designs."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
  - fileSearch
---
```

**New frontmatter:**
```yaml
---
name: conversation-persona-architect
description: "Systems-thinking conversation persona focused on architecture, scalability, composability, and long-term technical sustainability. Optimizes for elegant abstractions and extensible designs."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

#### 2.3 `agents/conversation-persona-pragmatist.md`

**Source:** `.github/agents/conversation-persona-pragmatist.md`

**Original frontmatter:**
```yaml
---
description: "Pragmatic conversation persona focused on shipping speed, minimal complexity, real-world constraints, and incremental delivery. Optimizes for getting things done with proven tools."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
---
```

**New frontmatter:**
```yaml
---
name: conversation-persona-pragmatist
description: "Pragmatic conversation persona focused on shipping speed, minimal complexity, real-world constraints, and incremental delivery. Optimizes for getting things done with proven tools."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

#### 2.4 `agents/conversation-persona-critic.md`

**Source:** `.github/agents/conversation-persona-critic.md`

**Original frontmatter:**
```yaml
---
description: "Devil's advocate conversation persona focused on fault-finding, risk analysis, hidden assumptions, and stress-testing claims. Optimizes for identifying what could go wrong."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
---
```

**New frontmatter:**
```yaml
---
name: conversation-persona-critic
description: "Devil's advocate conversation persona focused on fault-finding, risk analysis, hidden assumptions, and stress-testing claims. Optimizes for identifying what could go wrong."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

#### 2.5 `agents/conversation-persona-researcher.md`

**Source:** `.github/agents/conversation-persona-researcher.md`

The Researcher is the only persona with a unique tool requirement: `fetch`, used to retrieve external references and benchmarks. In Claude Code this maps to `WebFetch`.

**Original frontmatter:**
```yaml
---
description: "Evidence-first conversation persona focused on literature review, benchmarks, empirical data, and comparative analysis. Optimizes for factual grounding and quantitative rigor."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
  - fileSearch
  - fetch
---
```

**New frontmatter:**
```yaml
---
name: conversation-persona-researcher
description: "Evidence-first conversation persona focused on literature review, benchmarks, empirical data, and comparative analysis. Optimizes for factual grounding and quantitative rigor."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
  - WebFetch
---
```

**Note on `WebFetch`:** If `WebFetch` is not available in the target Claude Code environment (e.g., MCP not configured), `Bash` with `curl` is a functional fallback. The Researcher can issue `curl` commands to retrieve external URLs. Update the fallback note in the Researcher's body if needed.

---

### Step 3 — Update the `/analyze` Command

**File:** `commands/analyze.md`

The existing file is well-structured and describes the workflow correctly. The only addition needed is an explicit **dispatch instruction** at the very top of the command body, before the usage description.

This instruction tells Claude Code to hand off execution to the `conversation-moderator` sub-agent rather than attempting to execute the analysis inline in the current context.

**Add the following block immediately after the frontmatter, before the `# /analyze` heading:**

```markdown
> **Execution:** Dispatch this entire request to the `conversation-moderator` sub-agent.
> Pass the full message content (topic, mode, file paths) as the sub-agent's input.
> The moderator will read configuration files and orchestrate all persona sub-agents.
> Do NOT execute the conversation workflow inline — always delegate to the sub-agent.
```

**Why this matters:**
Without this instruction, Claude Code may interpret the `/analyze` command as a directive to perform the analysis directly in the current context. This bypasses the multi-agent architecture entirely — all personas would run in the same token window, context isolation would be impossible, and the sub-agent spawn budget (`max_sub_agent_calls: 25`) would be meaningless.

The dispatch instruction is a single-line behavioral contract that ensures Claude Code always routes through the proper agent hierarchy.

**No other changes to `commands/analyze.md` are needed.** The Mode A / Mode B examples, the output table, the connection to `/workflow`, and the configuration references are all correct.

---

### Step 4 — Verify State Path Alignment

The conversation workflow writes session state to `.github/state/conversation-state.yml` and phase outputs to `.github/state/phase-{N}/*.md`. This path is configured in `conversation/config/conversation.yml`:

```yaml
state:
  enabled: true
  path: ".github/state"
  auto_cleanup: false
  resume_on_restart: true
```

**This path works identically for both Claude Code and Copilot Chat.** Both platforms run in the context of the project workspace and have read/write access to the filesystem. No change is required.

**Verify the following:**

1. `.github/state/` directory exists (it does — `.gitkeep` is present)
2. The `Write` tool is declared in the moderator's frontmatter (it is, in the new Claude Code frontmatter above)
3. The `Bash` tool is available for directory creation (`mkdir -p .github/state/phase-2/`) if subdirectories don't yet exist

**State file paths referenced in the moderator body** (context rule token resolution table):

| Token | Path | Notes |
|-------|------|-------|
| `topic_brief` | `.github/state/topic-brief.md` | Written by moderator at start |
| `own_position` | `.github/state/phase-2/{persona-id}-position.md` | Written after Phase 2 |
| `other_positions` | `.github/state/phase-2/*.md` (excluding own) | Read during Phase 3 |
| `challenges_received` | `.github/state/phase-3/{persona-id}-challenges.md` | Written after Phase 3 |
| `convergence_analysis` | `.github/state/phase-5/convergence-analysis.md` | Written after Phase 5 |
| `gap_identification` | `.github/state/phase-5/gap-identification.md` | Written after Phase 5 |
| `aligned_positions` | `.github/state/phase-2/{id1}-position.md` + `{id2}-position.md` | Used in Phase 5.5 `concession_depth` optimization |

All paths are relative to the project root. No adjustment needed.

---

### Step 5 — Verify Orchestrator Integration

**File:** `agents/orchestrator.md`

**No changes are required to this file.** The orchestrator already has the correct `COMPLEX_NEW` routing logic:

```
COMPLEX_NEW:
conversation-moderator → analyst → architect → developer → tester → reviewer
```

And Gate 0 validation:

```
Gate 0 — After Analysis (conversation-moderator):
- Verify Executive Summary with a clear architectural recommendation
- Verify Decision Matrix with ≥ 3 options scored
- Verify ≥ 3 Recommendations with confidence levels
- Verify Quality Rubric score ≥ 3.0/5.0
- Save output to analysis/conversation/{descriptor}-convergence.md
```

Once `agents/conversation-moderator.md` exists, the orchestrator's `COMPLEX_NEW` chain will work natively in Claude Code. The orchestrator resolves `conversation-moderator` by scanning `agents/` for a file with `name: conversation-moderator` in its frontmatter — exactly what the new file provides.

**Also verify the downstream context loading** — the orchestrator passes convergence analysis context to analyst and architect via `overview.md`:

```markdown
## Prior Analysis Context
- **Source**: analysis/conversation/{filename}
- **Key Recommendations**: {numbered list}
- **Decision Matrix Winner**: {top-scored architecture}
```

This logic (Step 1, "Check `analysis/conversation/`" in the orchestrator) is already in place and does not depend on which platform ran the conversation workflow. Whether the analysis was produced by the Copilot Chat moderator or the Claude Code moderator, it lands in the same `analysis/conversation/` directory and is picked up identically.

---

## 7. Frontmatter Adaptation Reference

Complete before/after comparison for all 6 files.

### Moderator

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-moderator.md)
---
description: "Orchestrates conversation analysis workflows by spawning persona sub-agents..."
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

# AFTER (Claude Code — agents/conversation-moderator.md)
---
name: conversation-moderator
description: "Orchestrates conversation analysis workflows by spawning persona sub-agents..."
model: claude-opus-4-5
tools:
  - Task
  - Read
  - Write
  - Bash
---
```

### Generic Persona

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-persona.md)
---
description: "A configurable conversation participant that adopts a specified persona..."
tools:
  - readFile
  - codebase
  - textSearch
---

# AFTER (Claude Code — agents/conversation-persona.md)
---
name: conversation-persona
description: "A configurable conversation participant that adopts a specified persona, perspective, and priorities to generate position papers, challenges, and rebuttals in a structured conversation analysis workflow."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

### Architect Persona

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-persona-architect.md)
---
description: "Systems-thinking conversation persona focused on architecture, scalability, composability, and long-term technical sustainability."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
  - fileSearch
---

# AFTER (Claude Code — agents/conversation-persona-architect.md)
---
name: conversation-persona-architect
description: "Systems-thinking conversation persona focused on architecture, scalability, composability, and long-term technical sustainability. Optimizes for elegant abstractions and extensible designs."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

### Pragmatist Persona

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-persona-pragmatist.md)
---
description: "Pragmatic conversation persona focused on shipping speed, minimal complexity, real-world constraints, and incremental delivery."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
---

# AFTER (Claude Code — agents/conversation-persona-pragmatist.md)
---
name: conversation-persona-pragmatist
description: "Pragmatic conversation persona focused on shipping speed, minimal complexity, real-world constraints, and incremental delivery. Optimizes for getting things done with proven tools."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

### Critic Persona

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-persona-critic.md)
---
description: "Devil's advocate conversation persona focused on fault-finding, risk analysis, hidden assumptions, and stress-testing claims."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
---

# AFTER (Claude Code — agents/conversation-persona-critic.md)
---
name: conversation-persona-critic
description: "Devil's advocate conversation persona focused on fault-finding, risk analysis, hidden assumptions, and stress-testing claims. Optimizes for identifying what could go wrong."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
---
```

### Researcher Persona

```yaml
# BEFORE (Copilot Chat — .github/agents/conversation-persona-researcher.md)
---
description: "Evidence-first conversation persona focused on literature review, benchmarks, empirical data, and comparative analysis."
model: Claude Sonnet 4 (copilot)
tools:
  - readFile
  - codebase
  - textSearch
  - fileSearch
  - fetch
---

# AFTER (Claude Code — agents/conversation-persona-researcher.md)
---
name: conversation-persona-researcher
description: "Evidence-first conversation persona focused on literature review, benchmarks, empirical data, and comparative analysis. Optimizes for factual grounding and quantitative rigor."
model: claude-sonnet-4-5
tools:
  - Read
  - Write
  - Bash
  - WebFetch
---
```

---

## 8. Tool Name Mapping

Complete mapping of every Copilot Chat tool declaration to its Claude Code equivalent.

| Copilot Chat Tool | Claude Code Tool | Functional Equivalent | Notes |
|---|---|---|---|
| `agent` | `Task` | Sub-agent spawning | The `Task` tool is how Claude Code spawns sub-agents by name |
| `read` | `Read` | File reading | Same capability, different casing convention |
| `readFile` | `Read` | File reading | Copilot uses camelCase; Claude Code uses PascalCase |
| `codebase` | `Bash` (grep) | Semantic codebase search | Claude Code uses grep via Bash for text search |
| `search/codebase` | `Bash` (grep) | Codebase search | Slash-prefixed variant used in moderator frontmatter |
| `textSearch` | `Bash` (grep) | Text pattern search | `grep -r "pattern" .` |
| `search/textSearch` | `Bash` (grep) | Text pattern search | Slash-prefixed variant |
| `fileSearch` | `Bash` (find) | File name/path search | `find . -name "pattern"` |
| `search/fileSearch` | `Bash` (find) | File name/path search | Slash-prefixed variant |
| `fetch` | `WebFetch` | HTTP GET requests | Used by Researcher to retrieve external URLs and benchmarks |
| _(implicit write)_ | `Write` | File writing | Copilot doesn't require explicit write declaration; Claude Code does |

### Notes on Semantic Gaps

**`codebase` vs `Bash` (grep):**
The Copilot `codebase` tool performs semantic search with language-aware understanding (it understands function names, class hierarchies, etc.). Claude Code's `Bash` with `grep` is purely textual. For the conversation agents, this difference is immaterial — the personas primarily search for configuration files by path, not for code symbols. The moderator loads config files by known path (`conversation/config/conversation.yml`), not by semantic search.

**`fetch` vs `WebFetch`:**
If the Claude Code environment has MCP configured with a web fetch server, `WebFetch` is available as a first-class tool. If not, the Researcher persona can use `Bash` with `curl -s "URL"` as a fallback. In practice, the Researcher uses external references primarily during Phase 2 (Opening Position), when it needs to verify empirical claims. The Evidence Registry format already accounts for cases where evidence cannot be fetched (`[Unsourced assertion]` marker).

**`Write` declaration:**
Copilot Chat agents write files implicitly — no tool declaration is needed. Claude Code requires `Write` to be explicitly declared in the agent's frontmatter for the agent to have permission to write files. The moderator writes to `.github/state/` and `analysis/conversation/`. The persona agents should also declare `Write` in case they need to save intermediate outputs.

---

## 9. File-by-File Creation Guide

This section provides the exact recipe for creating each of the 6 new files. All files follow the same pattern:

1. Copy the source file from `.github/agents/`
2. Replace the frontmatter block (between the `---` delimiters)
3. Keep everything else identical

### Creation Checklist

| # | File to Create | Source | Frontmatter Changes | Body Changes |
|---|---|---|---|---|
| 1 | `agents/conversation-moderator.md` | `.github/agents/conversation-moderator.md` | Replace entire frontmatter; add `name:`; change model; update tools; remove `agents:` array | None (copy verbatim from `> **Runtime:**` onward) |
| 2 | `agents/conversation-persona.md` | `.github/agents/conversation-persona.md` | Replace entire frontmatter; add `name:`; add `model:`; update tools | None |
| 3 | `agents/conversation-persona-architect.md` | `.github/agents/conversation-persona-architect.md` | Replace entire frontmatter; add `name:`; update model string; update tools | None |
| 4 | `agents/conversation-persona-pragmatist.md` | `.github/agents/conversation-persona-pragmatist.md` | Replace entire frontmatter; add `name:`; update model string; update tools | None |
| 5 | `agents/conversation-persona-critic.md` | `.github/agents/conversation-persona-critic.md` | Replace entire frontmatter; add `name:`; update model string; update tools | None |
| 6 | `agents/conversation-persona-researcher.md` | `.github/agents/conversation-persona-researcher.md` | Replace entire frontmatter; add `name:`; update model string; update tools; replace `fetch` with `WebFetch` | None |

### Post-Creation Verification per File

After creating each file, verify:

- [ ] Frontmatter opens and closes with `---` on its own line
- [ ] `name:` field matches the filename without `.md` extension
- [ ] `model:` field uses a valid Claude Code model identifier
- [ ] `tools:` list contains only Claude Code tool names (no Copilot tools)
- [ ] No `agents:` array present
- [ ] Body starts immediately after the closing `---` (or after a blank line)
- [ ] The `> **Runtime:** Yes` line is present and intact
- [ ] No Copilot tool names appear anywhere in the body prose (check for `readFile`, `codebase`, `textSearch`, `fileSearch`, `fetch`)

### Update `commands/analyze.md`

After creating the agent files, add the dispatch instruction to `commands/analyze.md`.

**Exact insertion point:** Immediately after the closing `---` of the frontmatter, before the `# /analyze` heading.

**Text to insert:**

```markdown
> **Execution:** Dispatch this entire request to the `conversation-moderator` sub-agent.
> Pass the full message content (topic, mode, and any file paths) as the sub-agent's input.
> The moderator will read `conversation/config/*.yml` and orchestrate all persona sub-agents.
> Do NOT execute the conversation workflow inline — always delegate to the sub-agent.
```

---

## 10. State Management Alignment

The conversation workflow's state management is already designed to be platform-agnostic. This section documents the complete state schema and confirms there are no changes needed.

### Session State File

**Path:** `.github/state/conversation-state.yml`
**Writer:** `conversation-moderator` (both platforms)
**Reader:** `conversation-moderator` (session resume), `agents/orchestrator.md` (does not read — it reads `analysis/conversation/` instead)

```yaml
# Full session state schema
session:
  id: "{topic-slug}-{YYYY-MM-DD}"
  topic: "{topic}"
  variant: "{variant}"          # standard | quick | deep | devils_advocate | panel_review | document_as_position
  status: "active"              # active | completed | aborted
  started_at: "{ISO-8601}"

current_phase: "phase_2_opening"

phases_completed: []            # List of phase IDs that have status: completed

phases:
  phase_2_opening:
    status: "pending"           # pending | in_progress | completed | skipped
    started_at: null
    completed_at: null
    outputs: []                 # [{path: ".github/state/phase-2/{id}-position.md", persona: "{id}", status: "completed"}]

  phase_2_5_diversity_audit:
    status: "pending"
    outputs: []

  phase_2_7_tension_extraction:
    status: "pending"
    outputs: []

  phase_3_cross_examination:
    status: "pending"
    outputs: []

  phase_4_rebuttal:
    status: "pending"
    outputs: []

  phase_5_convergence:
    status: "pending"
    outputs: []
    quality_score: null         # Filled after Phase 5

  phase_5_5_gap_fill:
    status: "pending"
    outputs: []

  phase_6_output:
    status: "pending"
    outputs: []

metrics:
  sub_agent_calls: 0
  phases_completed: 0
  quality_iterations: 0

termination:
  reason: null                  # null | quality_abort | token_budget_exceeded | stall_detected | user_abort | completed

context_audit: []               # Per sub-agent call: [{phase, persona, context_delivered, context_blocked}]
```

### Phase Output Directory Structure

```
.github/state/
├── conversation-state.yml          ← Session tracker (written throughout)
├── topic-brief.md                  ← Written at session start by moderator
├── phase-2/
│   ├── architect-position.md
│   ├── pragmatist-position.md
│   ├── critic-position.md
│   └── researcher-position.md
├── phase-2-5/
│   └── diversity-audit.md
├── phase-2-7/
│   └── tension-matrix.md
├── phase-3/
│   ├── architect-challenges.md     ← Challenges authored BY architect (targeting others)
│   ├── pragmatist-challenges.md
│   ├── critic-challenges.md
│   └── researcher-challenges.md
├── phase-4/
│   ├── architect-rebuttal.md
│   ├── pragmatist-rebuttal.md
│   ├── critic-rebuttal.md
│   └── researcher-rebuttal.md
├── phase-5/
│   ├── convergence-analysis.md
│   ├── gap-identification.md       ← Written when quality score triggers gap-fill
│   └── quality-score.md
└── phase-5-5/
    └── gap-fill-position.md        ← One per gap-fill persona spawned
```

### Resume Behavior

Both the Copilot Chat and Claude Code moderators implement the same resume logic (defined in the moderator instruction body):

1. At startup, check for `.github/state/conversation-state.yml`
2. If `session.status == "active"`: report the interrupted session to the user and resume from the first phase where `status != "completed"`
3. If no state file: initialize fresh session

This means a session started in Copilot Chat can theoretically be resumed by the Claude Code moderator, and vice versa — as long as the state file and phase output files are present. The moderator body is identical across both platforms; only the frontmatter differs.

---

## 11. Dual-Platform Coexistence Model

After the integration is complete, the framework operates as a true dual-platform system.

### Entry Points

| Entry Point | Platform | Agent Invoked | Path |
|---|---|---|---|
| `@conversation-moderator` in VS Code chat | Copilot Chat | `.github/agents/conversation-moderator.md` | Copilot resolves by scanning `.github/agents/` |
| `/analyze` slash command | Claude Code | `agents/conversation-moderator.md` | Claude Code routes via dispatch instruction in `commands/analyze.md` |
| `COMPLEX_NEW` orchestrator routing | Claude Code | `agents/conversation-moderator.md` | Orchestrator chains `conversation-moderator` as first agent |
| `@analyze` prompt | Copilot Chat | `.github/agents/conversation-moderator.md` | Copilot routes via `agent:` in `.github/prompts/analyze.prompt.md` |
| `@analyze-documents` prompt | Copilot Chat | `.github/agents/conversation-moderator.md` | Copilot routes via `agent:` in `.github/prompts/analyze-documents.prompt.md` |

### Shared Runtime Resources

All entry points — regardless of platform — consume and produce to the same locations:

| Resource | Path | Direction |
|---|---|---|
| Core config | `conversation/config/conversation.yml` | Read |
| Persona config | `conversation/config/personas.yml` | Read |
| Quality config | `conversation/config/quality.yml` | Read |
| Extensions config | `conversation/config/extensions.yml` | Read |
| Protocol modules | `conversation/protocols/*/PROTOCOL.md` | Read |
| Skill modules | `conversation/skills/*/SKILL.md` | Read |
| Shared conventions | `conventions/shared.md` | Read |
| Session state | `.github/state/conversation-state.yml` | Read + Write |
| Phase outputs | `.github/state/phase-*/` | Read + Write |
| Convergence analysis | `analysis/conversation/*-convergence.md` | Write |

### Maintenance Implications

Because both platform versions share the same instruction body, **any change to the workflow logic must be applied to both files**. There are two strategies for managing this:

**Option A — Manual sync (current recommendation):**
When the instruction body needs updating (e.g., a new phase, a protocol change, a rubric dimension update), apply the change to both `.github/agents/conversation-moderator.md` and `agents/conversation-moderator.md`. The frontmatter is the only section that differs.

Suggested workflow for body updates:
1. Make the change in `.github/agents/conversation-moderator.md` (primary source)
2. Copy the changed section(s) to `agents/conversation-moderator.md`
3. Verify the frontmatter in the Claude Code file was not inadvertently overwritten

**Option B — Sync script (future automation):**
A shell script (`scripts/sync-conversation-agents.sh`) could automate this by:
1. Reading each `.github/agents/conversation-*.md` file
2. Extracting the body (everything after the second `---`)
3. Prepending the Claude Code-specific frontmatter
4. Writing to `agents/conversation-*.md`

This script would be run manually whenever the Copilot Chat files change, or as a pre-commit hook.

---

## 12. Cross-System Data Flow

This section traces the complete data flow for two key scenarios.

### Scenario A: User Runs `/analyze` in Claude Code (Mode A)

```
User types: /analyze "Should we use GraphQL or REST for the API layer?"
                │
                ▼
        commands/analyze.md
        (dispatch instruction routes to sub-agent)
                │
                ▼
        agents/conversation-moderator.md          [Claude Code runtime]
                │
                ├── Reads conversation/config/conversation.yml
                ├── Reads conversation/config/personas.yml
                ├── Reads conversation/config/quality.yml
                ├── Reads conversation/config/extensions.yml
                ├── Reads conversation/protocols/phase-routing/PROTOCOL.md
                ├── Reads conversation/protocols/evaluator-optimizer/PROTOCOL.md
                │
                ├── Writes .github/state/conversation-state.yml  (session init)
                ├── Writes .github/state/topic-brief.md
                │
                │── Phase 2: Spawns 4 persona sub-agents via Task tool
                │       │
                │       ├── Task → agents/conversation-persona-architect.md
                │       │     └── Writes .github/state/phase-2/architect-position.md
                │       ├── Task → agents/conversation-persona-pragmatist.md
                │       │     └── Writes .github/state/phase-2/pragmatist-position.md
                │       ├── Task → agents/conversation-persona-critic.md
                │       │     └── Writes .github/state/phase-2/critic-position.md
                │       └── Task → agents/conversation-persona-researcher.md
                │             └── Writes .github/state/phase-2/researcher-position.md
                │
                ├── Phase 2.5: Diversity audit (inline — moderator)
                │     └── Writes .github/state/phase-2-5/diversity-audit.md
                │
                ├── Phase 2.7: Tension extraction (inline — moderator)
                │     └── Writes .github/state/phase-2-7/tension-matrix.md
                │
                ├── Phase 3: Spawns 4 cross-examination sub-agents via Task tool
                │       (each receives OTHER personas' positions, not their own)
                │       └── Writes .github/state/phase-3/{persona}-challenges.md
                │
                ├── Phase 4: Spawns 4 rebuttal sub-agents via Task tool
                │       (each receives own position + challenges directed at them)
                │       └── Writes .github/state/phase-4/{persona}-rebuttal.md
                │
                ├── Phase 5: Convergence synthesis (inline — moderator)
                │     ├── Writes .github/state/phase-5/convergence-analysis.md
                │     └── Scores quality rubric (6 dimensions)
                │
                ├── Phase 5.5 (if quality < 3.6): Gap-fill
                │     └── Task → agents/conversation-persona.md  (gap-fill persona)
                │
                └── Phase 6: Output assembly
                      └── Writes analysis/conversation/{topic-slug}-convergence.md
                                          │
                                          └── (Available for future /workflow invocations)
```

### Scenario B: COMPLEX_NEW Request Routes Through Both Systems

```
User types: /workflow "Design a real-time collaborative editor"
                │
                ▼
        commands/workflow.md
                │
                ▼
        agents/orchestrator.md
                │
                ├── Classifies as COMPLEX_NEW (3+ components, architectural uncertainty)
                ├── Checks analysis/conversation/ — no prior analysis found
                ├── Creates docs/iterations/complex-new-realtime-editor/overview.md
                │
                │── Gate 0: Dispat
ches conversation-moderator (Mode C)
                │       │
                │       ▼
                │   agents/conversation-moderator.md
                │       │
                │       ├── Mode C: reads docs/project-brief.md, docs/requirements.md
                │       ├── Sets topic_type: architecture_design
                │       ├── Runs full 6-phase workflow (Phases 2 → 6)
                │       └── Writes analysis/conversation/complex-new-realtime-editor-convergence.md
                │
                ├── Gate 0 validation:
                │       ├── Checks Executive Summary ✓
                │       ├── Checks Decision Matrix (≥3 options) ✓
                │       ├── Checks Recommendations (≥3 with confidence) ✓
                │       └── Checks Quality Rubric score (≥3.0/5.0) ✓
                │
                ├── Updates overview.md with Prior Analysis Context section
                │
                ├── Dispatches: agents/analyst.md
                │       └── Reads docs/iterations/*/overview.md (includes analysis context)
                │
                ├── Dispatches: agents/architect.md
                │       └── Reads analysis/conversation/*-convergence.md directly
                │
                ├── Dispatches: agents/developer.md
                ├── Dispatches: agents/tester.md
                └── Dispatches: agents/reviewer.md
```

---

## 13. Verification Checklist

Use this checklist to confirm the integration is complete and working after creating the files.

### File Existence

- [ ] `agents/conversation-moderator.md` exists
- [ ] `agents/conversation-persona.md` exists
- [ ] `agents/conversation-persona-architect.md` exists
- [ ] `agents/conversation-persona-pragmatist.md` exists
- [ ] `agents/conversation-persona-critic.md` exists
- [ ] `agents/conversation-persona-researcher.md` exists
- [ ] `commands/analyze.md` has the dispatch instruction added

### Frontmatter Validation (per file)

For each of the 6 new agent files:

- [ ] `name:` field matches filename (without `.md`)
- [ ] `model:` is `claude-opus-4-5` (moderator) or `claude-sonnet-4-5` (personas)
- [ ] `tools:` list contains only: `Task`, `Read`, `Write`, `Bash` (+ `WebFetch` for Researcher)
- [ ] No `agents:` array present
- [ ] No Copilot tool names: `readFile`, `codebase`, `textSearch`, `fileSearch`, `fetch`
- [ ] No Copilot model name: `Claude Sonnet 4 (copilot)`
- [ ] Frontmatter opens and closes with `---` on its own line

### Body Integrity (per file)

- [ ] Body is identical to the corresponding `.github/agents/` source file
- [ ] `> **Runtime:** Yes` line is present
- [ ] Phase behavior sections are all present (Phase 2, 3, 4 for personas)
- [ ] Evidence Registry format table is present
- [ ] Critical Rules section is present
- [ ] For moderator: all 6 workflow steps (Steps 1–6) are present
- [ ] For moderator: State Management section is present
- [ ] For moderator: Context Rule Enforcement section is present

### Integration Path Validation

- [ ] Run `/workflow` with a request that triggers `COMPLEX_NEW` classification
  - Confirm the orchestrator dispatches to `conversation-moderator`
  - Confirm a convergence analysis file is created in `analysis/conversation/`
  - Confirm the orchestrator reads it and passes context to analyst + architect

- [ ] Run `/analyze "test topic"` directly
  - Confirm it dispatches to `conversation-moderator` (not inline execution)
  - Confirm sub-agent Task calls are made for each persona
  - Confirm output is saved to `analysis/conversation/`

- [ ] Confirm `.github/state/` is writable by the moderator
  - Run a short analysis and check that `conversation-state.yml` was created
  - Check that `topic-brief.md` was written

### Copilot Chat Regression

- [ ] Open Copilot Chat in VS Code
- [ ] Invoke `@conversation-moderator` — confirm it routes to `.github/agents/` (not `agents/`)
- [ ] Invoke `@analyze` prompt — confirm it still works via `.github/prompts/analyze.prompt.md`
- [ ] Confirm `.github/agents/` files are unchanged

---

## 14. Known Constraints & Trade-offs

### Constraint 1: Body Duplication

**Problem:** The instruction body of each agent exists in two files — one in `.github/agents/` and one in `agents/`. Any update to the workflow logic must be applied twice.

**Mitigation:** 
- Treat `.github/agents/` as the **primary source** for instruction body content
- When updating, always start with the Copilot file, then apply the same change to the Claude Code file
- A future sync script (see Section 11, Option B) can automate this
- The frontmatter is the only intentionally divergent section; flag it clearly in a comment at the top of each file

**Risk level:** Low. The instruction body changes infrequently (sprint-level changes, not request-by-request). The frontmatter never changes after initial setup.

### Constraint 2: `WebFetch` Availability

**Problem:** The Researcher persona's `fetch` tool maps to `WebFetch` in Claude Code. `WebFetch` requires an MCP server configured in the Claude Code environment. If MCP is not configured, the tool is unavailable.

**Mitigation:**
- Declare `WebFetch` in the frontmatter — Claude Code will silently ignore unavailable tools rather than erroring
- The Researcher can fall back to `Bash` with `curl -s "URL"` for HTTP GET requests
- The Evidence Registry's `[Unsourced assertion]` mechanism already handles cases where external sources cannot be fetched
- Most conversation analyses focus on design decisions where the relevant evidence is in the workspace (not on the web)

**Risk level:** Low. The Researcher's core value is in structuring evidence and applying rigor, not in live web browsing. The majority of sessions will not require `WebFetch`.

### Constraint 3: Inline Execution Fallback

**Problem:** If Claude Code fails to resolve `conversation-moderator` as a sub-agent (e.g., the file has a frontmatter parsing error), it may fall back to executing the workflow inline in the current context.

**Consequences of inline execution:**
- Context isolation between phases is impossible — all personas see all outputs
- Token budget grows linearly with phase count instead of being distributed across sub-agents
- The `max_sub_agent_calls` termination condition becomes meaningless
- Quality rubric scores for Perspective Diversity and Concession Depth will be artificially low (no genuine independent generation)

**Mitigation:**
- The dispatch instruction in `commands/analyze.md` explicitly warns against inline execution
- The moderator instruction body includes an "Inline Phase Execution" section that describes how to degrade gracefully if sub-agent spawning is unavailable — this is a documented fallback, not a failure
- Always verify agent file frontmatter after creation (see Section 13 checklist)

**Risk level:** Medium. If sub-agent spawning fails, the output is still useful but significantly lower quality. The quality rubric will catch this (low Perspective Diversity score will trigger gap-fill, which may also fail if sub-agents are unavailable).

### Constraint 4: State Collision Between Platforms

**Problem:** Both platforms write to the same `.github/state/` directory. If a Copilot Chat session and a Claude Code session run simultaneously on the same project, they will overwrite each other's `conversation-state.yml`.

**Mitigation:**
- In practice, a developer is unlikely to run both platforms simultaneously on the same conversation topic
- The state file includes `session.id` (topic slug + date) — the resume logic checks `session.status == "active"` before assuming ownership of the file
- A future enhancement could namespace state files by session ID: `.github/state/{session-id}/conversation-state.yml`

**Risk level:** Very low. Concurrent multi-platform sessions on the same project are an edge case. Document it as a known limitation.

### Constraint 5: Model Availability

**Problem:** The frontmatter specifies `claude-opus-4-5` for the moderator and `claude-sonnet-4-5` for personas. These model identifiers must match exactly what the Claude Code runtime accepts. Model availability and naming conventions may change as Anthropic releases updates.

**Mitigation:**
- Use the model identifiers from the existing `agents/orchestrator.md` file as the reference (it specifies `claude-opus-4-1-5` — verify whether this is the correct current identifier for the opus tier in your environment)
- If a specified model is unavailable, Claude Code will typically fall back to the default model — check Claude Code release notes for the current valid identifiers
- Keep model identifiers consistent across all dev workflow agents and conversation agents

**Risk level:** Low. Model naming is a configuration detail that can be updated in the frontmatter without touching the instruction body.

---

*Document complete. All sections 1–14 are covered. Proceed to Section 6 for the implementation sequence, and use Section 13 as the verification checklist after creating all files.*