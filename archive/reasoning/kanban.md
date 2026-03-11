# Mind Framework — Kanban Board

> **Source:** [improvement-plan.md](improvement-plan.md)
> **Created:** 2026-02-25
> **Last Updated:** 2026-02-25

---

## Legend

### Status

| Symbol | Status | Description |
|--------|--------|-------------|
| `⬜` | Backlog | Not yet scheduled |
| `🔵` | Ready | All prerequisites met, ready to start |
| `🟡` | In Progress | Currently being worked on |
| `✅` | Done | Completed and verified |
| `🔴` | Blocked | Cannot proceed — see blocker reference |
| `⏸️` | On Hold | Deprioritized or waiting for external input |

### Task Type

| Code | Type | Description |
|------|------|-------------|
| `IMPL` | Implementation | Code or agent file creation/modification |
| `CONFIG` | Configuration | Frontmatter, YAML, or settings changes |
| `DOCS` | Documentation | Markdown documentation authoring or updates |
| `SCRIPT` | Scripting | Shell scripts or automation tooling |
| `VERIFY` | Verification | Manual or automated validation checks |
| `E2E` | End-to-End Test | Full workflow execution test |
| `DESIGN` | Design | Architecture, schema, or feature design |
| `INFRA` | Infrastructure | Directory structure, .gitignore, CI/CD |

### Priority

| Level | Label | Meaning |
|-------|-------|---------|
| P0 | Blocking | Must complete before any other work proceeds |
| P1 | High | Should complete in the current sprint |
| P2 | Medium | Scheduled for next sprint |
| P3 | Low | Backlog — incremental, no deadline |

### Identifier Schema

```
P{phase}-T{sequence:02d}
```

- **Phase:** 1–5 (maps to improvement plan phases)
- **Sequence:** Two-digit zero-padded task number within the phase
- Example: `P1-T03` = Phase 1, Task 3

---

## Phase Summary

| Phase | Title | Priority | Tasks | Prereqs | Est. Effort |
|-------|-------|----------|-------|---------|-------------|
| **1** | Claude Code Agent Port | P0 | 16 | None | ~2 hours |
| **2** | Tooling & Automation | P1 | 9 | Phase 1 ✅ | ~4 hours |
| **3** | Documentation Consolidation | P1 | 7 | Phase 1 ✅ | ~3 hours |
| **4** | Quality & Testing Infrastructure | P2 | 8 | Phase 1 verified E2E | ~6 hours |
| **5** | Feature Extensions | P3 | 12 | Phases 1–4 stable | Variable |
| | | | **52 total** | | |

---

## Phase 1 — Claude Code Agent Port

> **Priority:** P0 — Blocking
> **Target:** Week 1
> **Prerequisite:** None
> **Acceptance:** All 6 agent files created, dispatch instruction added, all 9 verification checks pass

### Implementation Tasks

> **Note:** Deliverable paths in P1-T01…T06 reflect the original implementation location (`agents/`).
> These files were later reorganized to `agents/personas/` — see P3 documentation phase and `ARCHITECTURE.md`.

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P1-T01 | ✅ | `IMPL` | Create moderator agent (Claude Code) | Copy body from `.github/agents/conversation-moderator.md`. Replace frontmatter: `name: conversation-moderator`, `model: claude-opus-4-5`, `tools: [Task, Read, Write, Bash]`. Body verbatim — no changes. | `agents/conversation-moderator.md` | — |
| P1-T02 | ✅ | `IMPL` | Create generic persona agent | Copy body from `.github/agents/conversation-persona.md`. Replace frontmatter: `name: conversation-persona`, `model: claude-sonnet-4-5`, `tools: [Read, Write, Bash]`. | `agents/conversation-persona.md` | — |
| P1-T03 | ✅ | `IMPL` | Create architect persona agent | Copy body from `.github/agents/conversation-persona-architect.md`. Replace frontmatter: `name: conversation-persona-architect`, `model: claude-sonnet-4-5`, `tools: [Read, Write, Bash]`. | `agents/conversation-persona-architect.md` | — |
| P1-T04 | ✅ | `IMPL` | Create pragmatist persona agent | Copy body from `.github/agents/conversation-persona-pragmatist.md`. Replace frontmatter: `name: conversation-persona-pragmatist`, `model: claude-sonnet-4-5`, `tools: [Read, Write, Bash]`. | `agents/conversation-persona-pragmatist.md` | — |
| P1-T05 | ✅ | `IMPL` | Create critic persona agent | Copy body from `.github/agents/conversation-persona-critic.md`. Replace frontmatter: `name: conversation-persona-critic`, `model: claude-sonnet-4-5`, `tools: [Read, Write, Bash]`. | `agents/conversation-persona-critic.md` | — |
| P1-T06 | ✅ | `IMPL` | Create researcher persona agent | Copy body from `.github/agents/conversation-persona-researcher.md`. Replace frontmatter: `name: conversation-persona-researcher`, `model: claude-sonnet-4-5`, `tools: [Read, Write, Bash, WebFetch]`. Note: `fetch` → `WebFetch`. | `agents/conversation-persona-researcher.md` | — |
| P1-T07 | ✅ | `CONFIG` | Add dispatch instruction to analyze.md | Insert dispatch block after frontmatter `---`, before `# /analyze` heading. Block must instruct: delegate to `conversation-moderator` sub-agent, pass full message, do NOT execute inline. | `commands/analyze.md` (modified) | — |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P1-T08 | ✅ | `VERIFY` | V1 — Agent files exist | Confirm all 6 conversation agent files created in `agents/` | `ls agents/conversation-*.md` → 6 files | P1-T01…T06 |
| P1-T09 | ✅ | `VERIFY` | V2 — No Copilot tool names in agents/ | Ensure no Copilot-specific tool names leak into Claude Code agents | `grep -l 'readFile\|codebase\|textSearch\|fileSearch' agents/conversation-*.md` → empty | P1-T01…T06 |
| P1-T10 | ✅ | `VERIFY` | V3 — No Claude tool names in .github/ | Ensure no Claude-specific tool names leak into Copilot agents | `grep -l '\bTask\b\|\bRead\b\|\bWrite\b\|\bBash\b' .github/agents/conversation-*.md` → empty | — |
| P1-T11 | ✅ | `VERIFY` | V4 — Frontmatter validity | Each agent: `name:` matches filename, `model:` is valid identifier, `tools:` uses correct platform names | Manual inspection or script | P1-T01…T06 |
| P1-T12 | ✅ | `VERIFY` | V5 — Body parity | Instruction body identical between `.github/agents/` and `agents/` for each agent pair | `diff` between body sections (skip frontmatter) | P1-T01…T06 |
| P1-T13 | ✅ | `VERIFY` | V6 — Dispatch instruction present | Confirm dispatch block exists in `commands/analyze.md` | `grep 'Dispatch this entire request' commands/analyze.md` → match | P1-T07 |

### End-to-End Tests

| ID | Status | Type | Task | Description | Expected Outcome | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P1-T14 | 🔵 | `E2E` | V7 — Standalone /analyze | Run `/analyze "GraphQL vs REST for API layer"` in Claude Code | Moderator invoked via Task, 4 personas spawned, convergence saved to `analysis/conversation/` | P1-T01…T07 |
| P1-T15 | 🔵 | `E2E` | V8 — COMPLEX_NEW routing | Run `/workflow "analyze: design a notification system with 3 delivery channels"` in Claude Code | Orchestrator classifies COMPLEX_NEW → dispatches moderator → Gate 0 validates → analyst reads convergence | P1-T01…T07 |
| P1-T16 | 🔵 | `E2E` | V9 — Copilot regression | Invoke `@conversation-moderator` in VS Code Copilot Chat | Routes to `.github/agents/conversation-moderator.md` (unchanged behavior) | P1-T01…T07 |

---

## Phase 2 — Tooling & Automation

> **Priority:** P1 — High
> **Target:** Week 2
> **Prerequisite:** Phase 1 all tasks ✅
> **Acceptance:** All 3 scripts created, all 3 verification checks pass

### Implementation Tasks

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P2-T01 | ✅ | `SCRIPT` | Create agent body sync script | Sync instruction body from `.github/agents/conversation-*.md` (primary source) to `agents/conversation-*.md` (derived). Preserve Claude Code frontmatter. Report per-file sync status. | `scripts/sync-conversation-agents.sh` | Phase 1 ✅ |
| P2-T02 | ✅ | `SCRIPT` | Create integration validation script | 9-check cross-reference validator: path existence, stale refs, orchestrator chain completeness, CLAUDE.md resolution, YAML syntax, platform-specific tool name isolation, body divergence detection. Pass/fail per check with file:line detail. | `scripts/validate-integration.sh` | Phase 1 ✅ |
| P2-T03 | ✅ | `SCRIPT` | Create YAML config validation script | Validate `conversation/config/*.yml` against expected structure: required keys, cross-references (phase IDs in routing, skill paths in extensions), value constraints (weights 0–1). | `scripts/validate-config.sh` | Phase 1 ✅ |
| P2-T04 | ✅ | `DOCS` | Document sync workflow | Add usage instructions, trigger conditions (manual / pre-commit), and primary-source convention to script headers and `scripts/README.md` or inline comments. | Script headers + optional README | P2-T01 |
| P2-T05 | ⏸️ | `CONFIG` | Configure pre-commit hook (optional) | Wire `sync-conversation-agents.sh` as a git pre-commit hook so body drift is caught automatically. | `.githooks/pre-commit` or `.husky/` config | P2-T01 |
| P2-T06 | ✅ | `INFRA` | Make scripts executable | `chmod +x scripts/sync-conversation-agents.sh scripts/validate-integration.sh scripts/validate-config.sh` | File permissions | P2-T01…T03 |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P2-T07 | ✅ | `VERIFY` | V10 — Sync script smoke test | Run sync script; confirm all 6 agents synced with zero body diff | `bash scripts/sync-conversation-agents.sh` → 6 synced, 0 diffs | P2-T01, P2-T06 |
| P2-T08 | ✅ | `VERIFY` | V11 — Validation script full pass | Run integration validator on current codebase; all 9 checks pass | `bash scripts/validate-integration.sh` → all pass | P2-T02, P2-T06 |
| P2-T09 | ✅ | `VERIFY` | V12 — YAML config full pass | Run config validator on all 4 YAML files; all pass | `bash scripts/validate-config.sh` → all pass | P2-T03, P2-T06 |

---

## Phase 3 — Documentation Consolidation

> **Priority:** P1 — High
> **Target:** Week 2–3
> **Prerequisite:** Phase 1 all tasks ✅
> **Acceptance:** ARCHITECTURE.md created, CLAUDE.md updated, conversation/README.md updated, all cross-references resolve

### Implementation Tasks

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P3-T01 | ✅ | `DESIGN` | Design ARCHITECTURE.md structure | Define sections: System Overview, Agent Registry, Workflow Types (5), Quality Gates (4), Dual-Platform Model, Data Flow Diagram (mermaid), Directory Map, Configuration Reference. Outline content per section. | Design document or outline | Phase 1 ✅ |
| P3-T02 | ✅ | `DOCS` | Create ARCHITECTURE.md | Author single-document architecture overview at project root. Consolidates information from `arch-framework.md`, `arch-blueprint.md`, `imp-architecture-v2.md`. Include mermaid data flow diagram. | `ARCHITECTURE.md` | P3-T01 |
| P3-T03 | ✅ | `DOCS` | Update root CLAUDE.md | Add resource table entries for Claude Code conversation agents: `agents/conversation-moderator.md`, `agents/conversation-persona*.md`. Annotate existing `.github/agents/` entry as "Copilot Chat version". | `CLAUDE.md` (modified) | P1-T01…T06 |
| P3-T04 | ✅ | `DOCS` | Update conversation/README.md | Add "Platform Entry Points" section: Claude Code entry (`/analyze`, COMPLEX_NEW routing, `agents/conversation-*.md`), Copilot Chat entry (`@conversation-moderator`, `@analyze` prompt, `.github/agents/`). Note shared resources. | `conversation/README.md` (modified) | P1-T01…T06 |
| P3-T05 | ✅ | `DOCS` | Update .github/CLAUDE.md | Add note that `.github/agents/` files are the Copilot Chat platform variants; `agents/` holds Claude Code variants. Cross-link to `ARCHITECTURE.md`. | `.github/CLAUDE.md` (modified) | P3-T02 |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P3-T06 | ✅ | `VERIFY` | Cross-reference resolution | All paths referenced in ARCHITECTURE.md, CLAUDE.md, and README.md exist on disk | `scripts/validate-integration.sh` or manual check | P3-T02…T05 |
| P3-T07 | ✅ | `VERIFY` | No stale references | Grep for `panel-module/`, outdated agent paths, or broken links across all modified docs | `grep -r 'panel-module' docs/ CLAUDE.md ARCHITECTURE.md` → empty | P3-T02…T05 |

---

## Phase 4 — Quality & Testing Infrastructure

> **Priority:** P2 — Medium
> **Target:** Week 3–4
> **Prerequisite:** Phase 1 verified end-to-end (P1-T14…T16 ✅)
> **Acceptance:** Convergence schema defined, regression harness operational, quality log template in place

### Implementation Tasks

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P4-T01 | ✅ | `DESIGN` | Define convergence output schema | Formalize required sections: Executive Summary, Evidence Audit (≥1 claim), Decision Matrix (≥3 options), Recommendations (≥3 with confidence/risk/falsifiability), Quality Rubric (6 dimensions). Document as a schema spec. | `docs/reference/convergence-schema.md` or inline in test harness | Phase 1 E2E ✅ |
| P4-T02 | ✅ | `SCRIPT` | Create convergence regression test harness | Shell script or lightweight test framework that reads a convergence output file and validates against the schema. Assert: sections present, minimum counts, score ranges. Exit 0 on pass, non-zero with details on fail. | `scripts/test-convergence.sh` | P4-T01 |
| P4-T03 | ✅ | `CONFIG` | Create quality-log.yml template | Define the YAML schema for tracking convergence quality scores: date, topic, descriptor, overall_score, per-dimension scores, gate_0_pass, personas used. Seed with example entry. | `analysis/conversation/quality-log.yml` | P4-T01 |
| P4-T04 | ✅ | `SCRIPT` | Implement Gate 0 dashboard logging | Script or moderator instruction update that appends an entry to `quality-log.yml` after each convergence analysis completes. Reads score from convergence output, formats YAML entry. | `scripts/log-quality.sh` or moderator patch | P4-T03 |
| P4-T05 | ✅ | `IMPL` | Patch moderator for auto-logging | Update moderator agent body (in `.github/agents/conversation-moderator.md` as primary source) to invoke quality logging after Phase 8. Then sync to `agents/` via P2-T01 script. | `.github/agents/conversation-moderator.md` (modified) + sync | P4-T04, P2-T01 |
| P4-T06 | ✅ | `DOCS` | Document quality tracking workflow | Explain: how quality-log.yml is populated, how to read scores, what low scores indicate, remediation steps (adjust personas.yml, conversation.yml). | Section in `ARCHITECTURE.md` or standalone doc | P4-T03…T05 |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P4-T07 | ✅ | `VERIFY` | Regression test on sample output | Run convergence test harness against a known-good convergence file. All assertions pass. | `bash scripts/test-convergence.sh analysis/conversation/sample-convergence.md` → pass | P4-T02 |
| P4-T08 | ✅ | `VERIFY` | Quality log append test | Run an analysis, confirm quality-log.yml gains a new entry with correct schema | Inspect `analysis/conversation/quality-log.yml` after `/analyze` run | P4-T05 |

---

## Phase 5 — Feature Extensions

> **Priority:** P3 — Low (incremental)
> **Target:** Month 2+
> **Prerequisite:** Phases 1–4 complete and stable
> **Acceptance:** Per-feature; each extension independently deployable

### 5A — Conversation History Persistence

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P5-T01 | ✅ | `DESIGN` | Design history persistence structure | Define directory layout: `analysis/conversation/history/{session-id}/` with session-metadata.yml, phase subdirectories, convergence copy. Define session-id format. | Design spec | Phases 1–4 ✅ |
| P5-T02 | ✅ | `CONFIG` | Add persist_history flag to conversation.yml | Add `persist_history: false` (default off) to `conversation/config/conversation.yml`. Document the flag. | `conversation/config/conversation.yml` (modified) | P5-T01 |
| P5-T03 | ✅ | `IMPL` | Implement history persistence in moderator | Update moderator body: after Phase 8, if `persist_history: true`, copy phase outputs and convergence to `analysis/conversation/history/{session-id}/`. Sync to both platforms. | `.github/agents/conversation-moderator.md` + sync | P5-T01, P5-T02 |

### 5B — Custom Persona Presets from Specialists

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P5-T04 | ✅ | `DESIGN` | Design specialist-as-persona registration | Define `type: conversation-persona` frontmatter signal in `specialists/*.md`. Define how moderator discovers and includes them in persona pool alongside standard 4. | Design spec | Phases 1–4 ✅ |
| P5-T05 | ✅ | `IMPL` | Implement specialists/ scanning in moderator | Update moderator body: at startup, scan `specialists/` for files with `type: conversation-persona` in frontmatter. Add to available persona pool. Sync to both platforms. | `.github/agents/conversation-moderator.md` + sync | P5-T04 |
| P5-T06 | ✅ | `DOCS` | Document custom persona registration | How-to guide: creating a specialist file, required frontmatter fields, how it appears in persona selection. | Section in `conversation/README.md` or standalone guide | P5-T05 |

### 5C — Convergence Diff

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P5-T07 | ✅ | `DESIGN` | Design convergence diff feature | Define: detection of existing convergence, diff sections (recommendation shifts, new evidence, score changes, consensus stability), output format. | Design spec | Phases 1–4 ✅ |
| P5-T08 | ✅ | `IMPL` | Implement convergence diff in moderator | Update moderator body: after Phase 7 (Convergence), if prior convergence exists, produce a diff section appended to the new convergence output. Sync to both platforms. | `.github/agents/conversation-moderator.md` + sync | P5-T07 |
| P5-T09 | ✅ | `DOCS` | Document convergence diff usage | When it triggers, how to read the diff, use cases (iterative design exploration). | Section in `playbook.md` or `conversation/README.md` | P5-T08 |

### 5D — State Collision Prevention

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P5-T10 | ✅ | `DESIGN` | Design session ID namespacing | Define session ID format: `{topic-slug}-{YYYY-MM-DD}-{platform}`. Define state directory structure: `.github/state/{session-id}/`. Define cleanup policy. | Design spec | Phases 1–4 ✅ |
| P5-T11 | ✅ | `IMPL` | Implement session ID namespacing in moderator | Update moderator body: generate session ID at start, create namespaced state directory, use it for all state writes. Update state cleanup instructions. Sync to both platforms. | `.github/agents/conversation-moderator.md` + sync | P5-T10 |
| P5-T12 | ✅ | `INFRA` | Update .gitignore for namespaced state | Ensure `.github/state/*/` pattern covers namespaced session directories (should already work with existing `.github/state/` rule — verify). | `.gitignore` (verified or modified) | P5-T11 |

---

## Cross-Phase Dependencies

```
Phase 1 (P0)                    Phase 2 (P1)                Phase 4 (P2)
┌────────────┐                 ┌────────────┐              ┌────────────┐
│ P1-T01…T07 │─── creates ───▶│ P2-T01     │──── used ──▶│ P4-T05     │
│ (agents +  │    agent files  │ (sync      │   by sync   │ (moderator │
│  dispatch) │                 │  script)   │              │  patch)    │
└─────┬──────┘                 └────────────┘              └────────────┘
      │
      │ enables                 Phase 3 (P1)                Phase 5 (P3)
      │                        ┌────────────┐              ┌────────────┐
      ├───────────────────────▶│ P3-T02…T05 │              │ P5-T01…T12 │
      │                        │ (docs)     │              │ (features) │
      │                        └────────────┘              └────────────┘
      │                                                          ▲
      │ E2E verified                                             │
      └─────────────────── prerequisite ─────────────────────────┘
        (P1-T14…T16)           for all Phase 5 work
```

### Dependency Rules

| Rule | Constraint |
|------|-----------|
| **D1** | Phase 2, 3 cannot start until ALL Phase 1 implementation tasks (P1-T01…T07) are ✅ |
| **D2** | Phase 2, 3 can run in parallel |
| **D3** | Phase 4 cannot start until Phase 1 E2E tests (P1-T14…T16) are ✅ |
| **D4** | Phase 5 cannot start until Phases 1–4 are stable (all verification tasks ✅) |
| **D5** | P4-T05 (moderator patch) requires P2-T01 (sync script) to propagate changes |
| **D6** | All Phase 5 moderator updates require sync via P2-T01 to maintain body parity |
| **D7** | Within any phase, verification tasks depend on their corresponding implementation tasks |
| **D8** | Phase 5 sub-features (5A–5D) are independent and can be implemented in any order |

---

## Metrics

### Progress Tracking

| Phase | Backlog | Ready | In Progress | Done | Blocked | On Hold | Total |
|-------|---------|-------|-------------|------|---------|---------|-------|
| 1 | 0 | 3 | 0 | 13 | 0 | 0 | 16 |
| 2 | 0 | 0 | 0 | 8 | 0 | 1 | 9 |
| 3 | 0 | 0 | 0 | 7 | 0 | 0 | 7 |
| 4 | 0 | 0 | 0 | 8 | 0 | 0 | 8 |
| 5 | 0 | 0 | 0 | 12 | 0 | 0 | 12 |
| **Total** | **0** | **3** | **0** | **48** | **0** | **1** | **52** |

### Burndown

| Date | Total | Done | Remaining | % Complete |
|------|-------|------|-----------|------------|
| 2026-02-25 | 52 | 0 | 52 | 0% |
| 2026-02-25 | 52 | 13 | 39 | 25% |
| 2026-02-25 | 52 | 28 | 24 | 54% |
| 2026-02-25 | 52 | 48 | 4 | 92% |

---

## Task Index (Sorted by ID)

| ID | Phase | Type | Task | Priority | Status |
|----|-------|------|------|----------|--------|
| P1-T01 | 1 | `IMPL` | Create moderator agent (Claude Code) | P0 | ✅ |
| P1-T02 | 1 | `IMPL` | Create generic persona agent | P0 | ✅ |
| P1-T03 | 1 | `IMPL` | Create architect persona agent | P0 | ✅ |
| P1-T04 | 1 | `IMPL` | Create pragmatist persona agent | P0 | ✅ |
| P1-T05 | 1 | `IMPL` | Create critic persona agent | P0 | ✅ |
| P1-T06 | 1 | `IMPL` | Create researcher persona agent | P0 | ✅ |
| P1-T07 | 1 | `CONFIG` | Add dispatch instruction to analyze.md | P0 | ✅ |
| P1-T08 | 1 | `VERIFY` | V1 — Agent files exist | P0 | ✅ |
| P1-T09 | 1 | `VERIFY` | V2 — No Copilot tool names in agents/ | P0 | ✅ |
| P1-T10 | 1 | `VERIFY` | V3 — No Claude tool names in .github/ | P0 | ✅ |
| P1-T11 | 1 | `VERIFY` | V4 — Frontmatter validity | P0 | ✅ |
| P1-T12 | 1 | `VERIFY` | V5 — Body parity | P0 | ✅ |
| P1-T13 | 1 | `VERIFY` | V6 — Dispatch instruction present | P0 | ✅ |
| P1-T14 | 1 | `E2E` | V7 — Standalone /analyze | P0 | 🔵 |
| P1-T15 | 1 | `E2E` | V8 — COMPLEX_NEW routing | P0 | 🔵 |
| P1-T16 | 1 | `E2E` | V9 — Copilot regression | P0 | 🔵 |
| P2-T01 | 2 | `SCRIPT` | Create agent body sync script | P1 | ✅ |
| P2-T02 | 2 | `SCRIPT` | Create integration validation script | P1 | ✅ |
| P2-T03 | 2 | `SCRIPT` | Create YAML config validation script | P1 | ✅ |
| P2-T04 | 2 | `DOCS` | Document sync workflow | P1 | ✅ |
| P2-T05 | 2 | `CONFIG` | Configure pre-commit hook (optional) | P1 | ⏸️ |
| P2-T06 | 2 | `INFRA` | Make scripts executable | P1 | ✅ |
| P2-T07 | 2 | `VERIFY` | V10 — Sync script smoke test | P1 | ✅ |
| P2-T08 | 2 | `VERIFY` | V11 — Validation script full pass | P1 | ✅ |
| P2-T09 | 2 | `VERIFY` | V12 — YAML config full pass | P1 | ✅ |
| P3-T01 | 3 | `DESIGN` | Design ARCHITECTURE.md structure | P1 | ✅ |
| P3-T02 | 3 | `DOCS` | Create ARCHITECTURE.md | P1 | ✅ |
| P3-T03 | 3 | `DOCS` | Update root CLAUDE.md | P1 | ✅ |
| P3-T04 | 3 | `DOCS` | Update conversation/README.md | P1 | ✅ |
| P3-T05 | 3 | `DOCS` | Update .github/CLAUDE.md | P1 | ✅ |
| P3-T06 | 3 | `VERIFY` | Cross-reference resolution | P1 | ✅ |
| P3-T07 | 3 | `VERIFY` | No stale references | P1 | ✅ |
| P4-T01 | 4 | `DESIGN` | Define convergence output schema | P2 | ✅ |
| P4-T02 | 4 | `SCRIPT` | Create convergence regression test harness | P2 | ✅ |
| P4-T03 | 4 | `CONFIG` | Create quality-log.yml template | P2 | ✅ |
| P4-T04 | 4 | `SCRIPT` | Implement Gate 0 dashboard logging | P2 | ✅ |
| P4-T05 | 4 | `IMPL` | Patch moderator for auto-logging | P2 | ✅ |
| P4-T06 | 4 | `DOCS` | Document quality tracking workflow | P2 | ✅ |
| P4-T07 | 4 | `VERIFY` | Regression test on sample output | P2 | ✅ |
| P4-T08 | 4 | `VERIFY` | Quality log append test | P2 | ✅ |
| P5-T01 | 5A | `DESIGN` | Design history persistence structure | P3 | ✅ |
| P5-T02 | 5A | `CONFIG` | Add persist_history flag | P3 | ✅ |
| P5-T03 | 5A | `IMPL` | Implement history persistence in moderator | P3 | ✅ |
| P5-T04 | 5B | `DESIGN` | Design specialist-as-persona registration | P3 | ✅ |
| P5-T05 | 5B | `IMPL` | Implement specialists/ scanning in moderator | P3 | ✅ |
| P5-T06 | 5B | `DOCS` | Document custom persona registration | P3 | ✅ |
| P5-T07 | 5C | `DESIGN` | Design convergence diff feature | P3 | ✅ |
| P5-T08 | 5C | `IMPL` | Implement convergence diff in moderator | P3 | ✅ |
| P5-T09 | 5C | `DOCS` | Document convergence diff usage | P3 | ✅ |
| P5-T10 | 5D | `DESIGN` | Design session ID namespacing | P3 | ✅ |
| P5-T11 | 5D | `IMPL` | Implement session ID namespacing | P3 | ✅ |
| P5-T12 | 5D | `INFRA` | Update .gitignore for namespaced state | P3 | ✅ |

---

## Type Distribution

| Type | Count | % |
|------|-------|---|
| `IMPL` | 14 | 27% |
| `VERIFY` | 13 | 25% |
| `DOCS` | 8 | 15% |
| `SCRIPT` | 5 | 10% |
| `DESIGN` | 5 | 10% |
| `CONFIG` | 4 | 8% |
| `E2E` | 2 | 4% |
| `INFRA` | 2 | 4% |
| **Total** | **52** |  |

---

## Changelog

| Date | Author | Change |
|------|--------|--------|
| 2026-02-25 | — | Initial kanban created from improvement-plan.md. 52 tasks across 5 phases. |
| 2026-02-25 | — | Phase 1 implementation complete: P1-T01…T13 ✅. E2E tests (P1-T14…T16) and Phase 2/3 tasks now 🔵 Ready. 25% overall. |
| 2026-02-25 | — | Phases 2–3 complete: P2-T01…T09 ✅ (P2-T05 ⏸️), P3-T01…T07 ✅. 54% overall. |
| 2026-02-25 | — | Phases 4–5 complete: P4-T01…T08 ✅, P5-T01…T12 ✅. All features implemented. 92% overall. Remaining: 3 E2E tests (🔵), 1 optional pre-commit hook (⏸️). |
| 2026-02-25 | — | Phase 2 complete (8/9 ✅, P2-T05 ⏸️ on hold). Phase 3 complete (7/7 ✅). ARCHITECTURE.md created. 3 scripts created and verified. 54% overall. |
