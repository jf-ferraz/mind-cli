# Sprint Blueprint — [Sprint Title]

> **Source:** [Link to improvement plan or parent document]
> **Created:** YYYY-MM-DD
> **Last Updated:** YYYY-MM-DD

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

- **Phase:** Sequential integer (maps to sprint phases)
- **Sequence:** Two-digit zero-padded task number within the phase
- Example: `P1-T03` = Phase 1, Task 3

### Parallelization Markers

Use these markers in the "Depends On" column to signal agent spawning strategy:

| Marker | Meaning |
|--------|---------|
| `—` | No dependencies — can start immediately |
| `P1-T01` | Blocked on a specific task |
| `Phase 1 ✅` | Blocked on entire phase completion |
| `‖ P1-T01, P1-T02` | Can run in parallel with listed tasks |

**Concurrency rule:** Tasks with `—` or `‖` markers within the same phase can be dispatched as concurrent sub-agents. Tasks with explicit `P{n}-T{nn}` dependencies must wait.

---

## Phase Summary

| Phase | Title | Priority | Tasks | Prereqs | Est. Effort |
|-------|-------|----------|-------|---------|-------------|
| **1** | [Phase 1 Title] | P0 | — | None | — |
| **2** | [Phase 2 Title] | P1 | — | Phase 1 ✅ | — |
| | | | **— total** | | |

---

## Phase 1 — [Phase Title]

> **Priority:** P0 — Blocking
> **Target:** [Timeframe]
> **Prerequisite:** None
> **Acceptance:** [What must be true for this phase to be complete]

### Implementation Tasks

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P1-T01 | ⬜ | `IMPL` | [Task name] | [What to do] | [Output file/artifact] | — |
| P1-T02 | ⬜ | `IMPL` | [Task name] | [What to do] | [Output file/artifact] | — |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P1-T03 | ⬜ | `VERIFY` | [Check name] | [What to verify] | [Command to run] | P1-T01, P1-T02 |

### Gate Checkpoint

**Gate condition:** [What must pass before Phase 2 begins]
**Acceptance criteria:**
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

---

## Phase 2 — [Phase Title]

> **Priority:** P1 — High
> **Target:** [Timeframe]
> **Prerequisite:** Phase 1 all tasks ✅
> **Acceptance:** [What must be true for this phase to be complete]

### Implementation Tasks

| ID | Status | Type | Task | Description | Deliverable | Depends On |
|----|--------|------|------|-------------|-------------|------------|
| P2-T01 | ⬜ | `IMPL` | [Task name] | [What to do] | [Output file/artifact] | Phase 1 ✅ |

### Verification Tasks

| ID | Status | Type | Task | Description | Command / Action | Depends On |
|----|--------|------|------|-------------|-----------------|------------|
| P2-T02 | ⬜ | `VERIFY` | [Check name] | [What to verify] | [Command to run] | P2-T01 |

### Gate Checkpoint

**Gate condition:** [What must pass]
**Acceptance criteria:**
- [ ] [Criterion 1]

---

<!-- Repeat Phase sections as needed -->

---

## Cross-Phase Dependencies

```
Phase 1 (P0)              Phase 2 (P1)
┌────────────┐           ┌────────────┐
│ P1-T01…T02 │── gate ──▶│ P2-T01     │
│ (core)     │           │ (build on) │
└────────────┘           └────────────┘
```

### Dependency Rules

| Rule | Constraint |
|------|-----------|
| **D1** | [Phase N] cannot start until [condition] |
| **D2** | [Phases X and Y] can run in parallel |
| **D3** | [Task] requires [other task] for propagation |

---

## Metrics

### Progress Tracking

| Phase | Backlog | Ready | In Progress | Done | Blocked | On Hold | Total |
|-------|---------|-------|-------------|------|---------|---------|-------|
| 1 | — | — | — | — | — | — | — |
| 2 | — | — | — | — | — | — | — |
| **Total** | **—** | **—** | **—** | **—** | **—** | **—** | **—** |

### Burndown

| Date | Total | Done | Remaining | % Complete |
|------|-------|------|-----------|------------|
| YYYY-MM-DD | — | 0 | — | 0% |

---

## Task Index (Sorted by ID)

| ID | Phase | Type | Task | Priority | Status |
|----|-------|------|------|----------|--------|
| P1-T01 | 1 | `IMPL` | [Task name] | P0 | ⬜ |
| P1-T02 | 1 | `IMPL` | [Task name] | P0 | ⬜ |
| P1-T03 | 1 | `VERIFY` | [Check name] | P0 | ⬜ |
| P2-T01 | 2 | `IMPL` | [Task name] | P1 | ⬜ |
| P2-T02 | 2 | `VERIFY` | [Check name] | P1 | ⬜ |

---

## Changelog

| Date | Author | Change |
|------|--------|--------|
| YYYY-MM-DD | — | Initial blueprint created from [source]. |
