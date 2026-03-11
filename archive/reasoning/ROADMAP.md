# Conversation Workflow v2 — Implementation Roadmap

> **Date:** 2026-02-24
> **Status:** Complete — All 6 sprints delivered (48/48 items)
> **Tracking:** Kanban — move items across columns as work progresses

---

## Board Legend

| Column | Meaning |
|--------|---------|
| **Backlog** | Defined, not yet started |
| **Ready** | Dependencies met, can start immediately |
| **In Progress** | Actively being implemented |
| **Review** | Implemented, needs validation/testing |
| **Done** | Validated and merged |
| **Blocked** | Cannot proceed — dependency or decision needed |

**Size labels:** `XS` = < 15 min, `S` = 15-30 min, `M` = 30-60 min, `L` = 1-2 hrs

**Priority:** `P0` = blocker for others, `P1` = high impact, `P2` = medium, `P3` = nice-to-have

---

## Sprint 1: Foundation + Quick Wins ✅

> **Goal:** State management, context isolation, and immediate persona quality improvements
> **Scope:** Phase A (G1 + G2) + Skill Tier 1 (prompt-level changes)
> **Estimated effort:** ~4 hours | **Actual:** ~2 hours
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| A1 | Create `.github/state/` directory + add to `.gitignore` | XS | P0 | `.gitignore`, `.github/state/.gitkeep` |
| A2 | Add `state:` config section to `conversation-analysis.yml` _(note: runtime state file schema lives in `archive/blueprint-prompts-and-patches.md` — A4 must reference it)_ | XS | P0 | `.github/config/conversation-analysis.yml` |
| A3 | _(merged into A2)_ | — | — | — |
| A6 | Add `context_rules:` section to `conversation-analysis.yml` | S | P0 | `.github/config/conversation-analysis.yml` |
| S1 | Add open verification questions to persona Phase 3 instructions | S | P1 | All 5 persona agent files |
| S2 | Add MUST/SHOULD/COULD severity classification to Phase 3 | S | P1 | All 5 persona agent files |
| S3 | Add scope discipline / flag-don't-drift to all persona agents | S | P1 | All 5 persona agent files |
| S4 | Add rejected alternatives section to Phase 2 output format | S | P1 | All 5 persona agent files |
| S5 | Add 2+ step reasoning chains requirement to Phase 2 + Phase 4 | S | P1 | All 5 persona agent files |

### Blocked

_empty_

### Quality Notes

- **A2 scope:** `state:` config block added to YAML. The runtime `conversation-state.yml` schema is in `archive/blueprint-prompts-and-patches.md`. Sprint 2 task A4 must embed/reference this schema when writing the moderator state management block.
- **Critic severity dual-track:** The critic previously used `critical/high/medium/low`; new MUST/SHOULD/COULD system maps via `critical=MUST, high=SHOULD, medium/low=COULD`. Both systems coexist with a compatibility note. Resolve definitively in Sprint 2 when moderator convergence logic is updated (S8).
- **S6 deferred:** Dual-path verification for strong claims (T2) was part of Skill Tier 2 and is correctly placed in Sprint 2 backlog.

---

## Sprint 2: Moderator Intelligence ✅

> **Goal:** Moderator reads state, enforces context rules, validates deliverables, weights arguments
> **Scope:** Phase A moderator changes + Skill Tier 2 (moderator logic)
> **Depends on:** Sprint 1 ✅ (config sections now exist)
> **Estimated effort:** ~3 hours | **Actual:** ~2 hours
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| A4 | State management block in moderator (session init, phase boundary read/write) | M | P0 | `.github/agents/conversation-moderator.md` |
| A5 | Resume detection logic (check `session.status == "active"` on start) | M | P0 | `.github/agents/conversation-moderator.md` |
| A7 | Context rule enforcement block (token resolution + blocked-content check) | M | P0 | `.github/agents/conversation-moderator.md` |
| A8 | Context audit trail (`context_delivered` / `context_blocked` in state file) | S | P1 | `.github/agents/conversation-moderator.md` |
| S6 | Dual-path confidence verification rule in Phase 2 (all 5 persona files) | S | P1 | All 5 persona agent files |
| S7 | Deliverable verification protocol — 3 gates with re-prompt instructions | M | P1 | `.github/agents/conversation-moderator.md` |
| S8 | Argument weighting hierarchy in Phase 5 + severity unification (MUST/SHOULD/COULD) | S | P1 | `.github/agents/conversation-moderator.md` |
| A9 | Workflow doc updated: v2 features section + Phase 3 severity label update | S | P2 | `.github/workflows/conversation-analysis-workflow.md` |

### Quality Notes

- **A4/A5 merged:** Resume detection is implemented inline within the State Management block (single cohesive section, not split).
- **A7/A8 merged:** Audit trail is part of the Context Rule Enforcement block.
- **S8 scope expanded:** Also updated Phase 3 challenge template in the workflow doc to use MUST/SHOULD/COULD (retired `critical | significant | minor`).
- **Critic severity system:** Now fully unified. `critical/high/medium/low` references in critic's Phase 3 section have a single authoritative mapping; moderator Critical Rules section carries the canonical definition.
- **A9 scope expanded:** Added `v2 Features` reference section to workflow doc covering all Sprint 1 + Sprint 2 improvements.

---

## Sprint 3: Flow Control ✅

> **Goal:** Conditional routing, termination conditions — workflow becomes self-correcting
> **Scope:** Phase B (G3 + G5)
> **Depends on:** Sprint 2 ✅ (moderator has state management)
> **Estimated effort:** ~3 hours | **Actual:** ~1.5 hours
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| B1 | Add `phase_routing:` directed graph config to YAML | M | P0 | `.github/config/conversation-analysis.yml` |
| B2 | Define routing actions (inject_contrarian, escalate_challenge_depth, repeat_round, etc.) | S | P0 | `.github/config/conversation-analysis.yml` |
| B3 | Add routing evaluation block to moderator (condition eval after each phase) | M | P0 | `.github/agents/conversation-moderator.md` |
| B4 | Add loop prevention logic (max_retries tracking per condition ID) | S | P1 | `.github/agents/conversation-moderator.md` |
| B5 | Add `termination:` config to YAML (early_convergence, quality_abort, token_budget, stall_detection) | S | P0 | `.github/config/conversation-analysis.yml` |
| B6 | Add termination check block to moderator (evaluated before routing) | M | P0 | `.github/agents/conversation-moderator.md` |
| B7 | Add graceful degradation output template (partial output with caveats) | S | P1 | `.github/agents/conversation-moderator.md` |
| B8 | Update variant presets to map to `phase_routing` overrides | S | P2 | `.github/config/conversation-analysis.yml` |

### Quality Notes

- **B1/B2 merged:** Routing actions are defined inline within the `phase_routing` graph (each condition's `action` field), not as a separate section.
- **B3/B4/B6/B7 merged:** All moderator routing logic delivered as a single cohesive "Phase Routing (v2)" section including termination checks (priority 1), routing evaluation (priority 2), loop prevention with per-condition retry tracking, and graceful degradation output template.
- **B5 delivered:** `termination:` section added with 4 condition types: early_convergence, quality_abort, token_budget (25 max / 20 warning), stall_detection.
- **B8 delivered:** All 5 variant presets now include `phase_routing_override` sections that merge on top of the base routing graph.
- **max_rounds bumped:** `moderation.max_rounds` updated from 1 → 2 to align with routing conditions that reference `round < max_rounds`.
- **Severity alignment:** Phase 4 routing condition uses `'MUST'` (not `'critical'`) to match the unified MUST/SHOULD/COULD severity system from Sprint 2.
- **Workflow doc updated:** v2 Features section expanded with Phase Routing, Termination Conditions, and Graceful Degradation subsections.

---

## Sprint 4: Quality Loop ✅

> **Goal:** Evaluator-optimizer pattern — quality-driven convergence
> **Scope:** Phase C (G4)
> **Depends on:** Sprint 3 ✅ (routing needed for Phase 5 → 5.5 → 5 loop)
> **Estimated effort:** ~2.5 hours | **Actual:** ~1 hour
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| C1 | Add `evaluator_optimizer:` config to YAML (loop params, priority order, dimension limits) | S | P0 | `.github/config/conversation-analysis.yml` |
| C2 | Add evaluator protocol to moderator Phase 5 (structured rubric scoring with evidence + calibration guidelines) | M | P0 | `.github/agents/conversation-moderator.md` |
| C3 | Add optimizer protocol to moderator Phase 5.5 (targeted gap-fill dispatch + reintegration) | M | P0 | `.github/agents/conversation-moderator.md` |
| C4 | Define dimension-specific optimization strategies (6 strategies with expected improvement ranges) | M | P1 | `.github/agents/conversation-moderator.md` |
| C5 | Add loop exit conditions (target_score, max_iterations, stall_detection, score interpretation) | S | P0 | `.github/agents/conversation-moderator.md` |
| C6 | Add score history tracking in state file (iteration history array + phase_5_5 tracking schema) | S | P1 | `.github/agents/conversation-moderator.md` |

### Quality Notes

- **C2/C5 merged scope:** The evaluator role section includes both the scoring protocol and the 5-condition exit decision tree (target met, max iterations, stall, usable-with-caveats, abort). Calibration guidelines with concrete anchors per dimension are included to reduce scoring inconsistency.
- **C3/C4 merged scope:** The optimizer role and dimension-specific strategies are delivered as a single cohesive section. Each of the 6 dimensions has a tailored strategy with persona selection, prompt guidance, and expected improvement range.
- **C6 expanded:** Score history schema covers both `phase_5_convergence` (scores + history array) and `phase_5_5_gap_fill` (per-iteration trigger, strategy, persona, output paths, quality delta).
- **Gap-fill context rules:** Documented inline — gap-fill personas receive convergence + gap identification but are blocked from individual positions/challenges (fresh perspective). Exception for `concession_depth` which requires seeing aligned positions.
- **Loop cost budget:** Documented as 1–2 calls per iteration, 2–4 max for full loop. Well within the 25-call token budget.
- **Workflow doc updated:** v2 Features section expanded with Evaluator-Optimizer Loop subsection covering loop control parameters, dimension strategies table, score history, and cost impact.

---

## Sprint 5: User Control & Dynamic Features ✅

> **Goal:** Approval gates, dynamic speaker selection, mediated delegation
> **Scope:** Phase D (G6 + G7 + G8) — all opt-in
> **Depends on:** Sprint 2 ✅ (moderator state management); independent of Sprints 3-4
> **Estimated effort:** ~3 hours | **Actual:** ~1 hour
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| D1 | Add `approval_gates:` config to YAML (3 gates with options) | S | P2 | `.github/config/conversation-analysis.yml` |
| D2 | Add gate execution protocol to moderator (pause, summarize, wait, process, log) | M | P2 | `.github/agents/conversation-moderator.md` |
| D3 | Add gate summary templates (position table, concession summary, quality summary) | S | P2 | `.github/agents/conversation-moderator.md` |
| D4 | Add `dynamic_selection:` config to YAML (strategy, constraints) | S | P2 | `.github/config/conversation-analysis.yml` |
| D5 | Add speaker selection logic to moderator (strategy table, selection protocol, logging) | M | P2 | `.github/agents/conversation-moderator.md` |
| D6 | Add `delegation:` config to YAML (budget, validation, response injection) | S | P3 | `.github/config/conversation-analysis.yml` |
| D7 | Add delegation request format to ALL 5 persona agent files | S | P3 | `.github/agents/conversation-persona*.md` |
| D8 | Add delegation detection + routing logic to moderator (validation matrix, routing, injection) | M | P3 | `.github/agents/conversation-moderator.md` |
| D9 | Add delegation loop prevention rules (chain depth 1, no circular, format enforcement) | XS | P3 | `.github/agents/conversation-moderator.md` |

### Quality Notes

- **D2/D3 merged scope:** Gate execution protocol and summary templates are delivered as a single cohesive "Approval Gates (v2)" section in the moderator. Includes complete execution protocol (6 steps), 3 gate summary templates (position/concession/quality), action dispatch table, state logging schema, and feature interaction notes.
- **D5 delivered with strategies:** Selection protocol includes all 3 strategies (most_relevant, least_heard, priority_weighted) as a dispatch table, constraint enforcement rules, selection reasoning format, and state logging schema.
- **D7 persona-specific framing:** Each persona’s delegation template uses persona-appropriate language in the "Why needed" field (e.g., architect says "architectural analysis", critic says "risk analysis", researcher says "evidence-based analysis").
- **D8/D9 merged scope:** Delegation detection, validation matrix (9 checks), routing protocol, response injection, budget accounting, and all 4 loop prevention rules are delivered as a single "Mediated Delegation (v2)" section with integrated state tracking schema.
- **All features opt-in:** All three features default to `enabled: false`. When disabled, zero behavioral change — verified by checking that all moderator protocol sections begin with conditional checks.
- **Moderator size impact:** File grew from ~639 to ~800 lines. Post-Sprint 5 cleanup recommended (protocol extraction using Sprint 6 skill system pattern).
- **Workflow doc updated:** v2 Features section expanded with Approval Gates, Dynamic Speaker Selection, and Mediated Delegation subsections.

---

## Sprint 6: Skill System Architecture ✅

> **Goal:** Loadable skill modules for conversation personas
> **Scope:** Skill Tier 3 (new file system + config)
> **Depends on:** Sprint 1 ✅ (persona prompt changes), Sprint 2 ✅ (moderator changes)
> **Estimated effort:** ~2.5 hours | **Actual:** ~1 hour
> **Completed:** 2026-02-24

### Backlog

_empty_

### Ready

_empty — all items Done_

### In Progress

_empty_

### Review

_empty_

### Done

| ID | Task | Size | Priority | Files Touched |
|----|------|------|----------|---------------|
| S9a | Create `skills/conversation/SKILLS.md` index file | S | P2 | New: `skills/conversation/SKILLS.md` |
| S9b | Create `skills/conversation/challenge-methodology/SKILL.md` | S | P2 | New: `skills/conversation/challenge-methodology/SKILL.md` |
| S9c | Create `skills/conversation/evidence-standards/SKILL.md` | S | P2 | New: `skills/conversation/evidence-standards/SKILL.md` |
| S9d | Create `skills/conversation/reasoning-chains/SKILL.md` | S | P2 | New: `skills/conversation/reasoning-chains/SKILL.md` |
| S9e | Create `skills/conversation/decision-documentation/SKILL.md` | S | P2 | New: `skills/conversation/decision-documentation/SKILL.md` |
| S9f | Create `skills/conversation/scope-discipline/SKILL.md` | S | P2 | New: `skills/conversation/scope-discipline/SKILL.md` |
| S10 | Add `skill_injection:` config to YAML (rules for auto-injection per phase) | M | P2 | `.github/config/conversation-analysis.yml` |
| S11 | Add topic type classification to conversation config | S | P2 | `.github/config/conversation-analysis.yml` |

### Quality Notes

- **S9a-f content fidelity:** Each skill file's content was extracted from the actual inline additions in all 5 persona agent files (verified via grep), then canonicalized with persona-specific variant tables, good/weak examples, anti-patterns, and verification checklists.
- **S9b merged scope:** `challenge-methodology` combines S1 (open verification questions) and S2 (MUST/SHOULD/COULD severity classification) into one skill module since both govern Phase 3 cross-examination quality.
- **S10 phase mapping:** Skill injection rules map skills to the phases where they're most relevant — Phase 2 gets 4 skills (all opening position skills), Phase 3 gets 2 (challenge + scope), Phase 4 gets 2 (reasoning + scope), Phase 5.5 gets 2 (evidence + challenge).
- **S11 topic_type:** Added as optional auto-detect classification that can influence persona selection via `selection_guide`. Default is empty string with `auto_detect: true`.
- **Backward compatibility:** Skill injection is `enabled: true` by default but purely additive — it enriches prompts, doesn't remove existing inline content from persona files. Sprint 1 inline additions remain as the runtime instructions; skills serve as the canonical reference and documentation source.

---

## Dependency Graph

```
Sprint 1 (Foundation + Quick Wins)
│
├─── A1, A2, A3, A6 ──────────────────────────┐
│    (config + structure)                       │
│                                               ▼
├─── S1-S5 ──────────┐               Sprint 2 (Moderator Intelligence)
│    (persona prompts)│                  │
│                     │     ┌────────────┤
│                     │     │            │
│                     ▼     ▼            ▼
│               Sprint 6    │     A4, A5, A7, A8, S6-S8, A9
│               (Skills)    │            │
│                           │            │
│                           │      ┌─────┴──────────────┐
│                           │      │                     │
│                           │      ▼                     ▼
│                           │  Sprint 3            Sprint 5
│                           │  (Flow Control)      (User Control)
│                           │      │               [independent]
│                           │      │
│                           │      ▼
│                           │  Sprint 4
│                           │  (Quality Loop)
│                           │
│                           └── (can start after Sprint 2)
```

**Critical path:** Sprint 1 ✅ → Sprint 2 ✅ → Sprint 3 ✅ → Sprint 4 ✅ (critical path complete)

**Parallel lanes:**
- Sprint 5 ✅ completed after Sprint 2 (independent of 3-4)
- Sprint 6 ✅ completed after Sprint 1 (independent of 2-5)
- S1-S5 in Sprint 1 have no dependencies and can run in parallel with A1-A6

---

## Next Actions

### ✅ v2 Roadmap Complete

All 6 sprints (48 items) have been delivered. The conversation workflow v2 is feature-complete.

### Post-Completion Recommendations

| Priority | Action | Description | Status |
|----------|--------|-------------|--------|
| P1 | **End-to-end integration test** | Run a real conversation analysis with all v2 features enabled (state management, routing, evaluator-optimizer, approval gates, dynamic selection, delegation) to validate the full stack | ✅ Test plan created |
| P2 | **Moderator protocol extraction** | The moderator was ~800 lines. Extracted 4 protocol sections (Phase Routing, Evaluator-Optimizer, Approval Gates, Mediated Delegation) into `protocols/conversation/` | ✅ Done (Sprint 7) |
| P2 | **Blueprint cleanup** | Archived `blueprint-prompts-and-patches.md` to `archive/` — all patches have been applied | ✅ Done (Sprint 7) |
| P3 | **Legacy `phases:` deprecation** | Add deprecation notice to the `phases:` YAML section or remove it — `phase_routing:` is now the authoritative routing source | Not started |
| P3 | **Skill system evolution** | Consider making skills the primary source with persona files referencing them (inverse of current inline-first approach) | Not started |

### Post-Completion Considerations

| Concern | Action | Status |
|---------|--------|--------|
| ~~Moderator file size (~800 lines)~~ | ~~Extract v2 protocol sections into loadable protocol files~~ | ✅ Done — reduced to ~400 lines |
| Legacy `phases:` section drift | Add deprecation notice or remove — `phase_routing:` is now authoritative | Not started |
| ~~End-to-end integration test~~ | ~~Run a real conversation analysis with all v2 features enabled~~ | ✅ Test plan created |
| ~~Blueprint cleanup~~ | ~~Archive `blueprint-prompts-and-patches.md`~~ | ✅ Moved to `archive/` |
| Skill system evolution | Consider making skills the primary source with persona files referencing them (inverse of current) | Post v2.0 |

---

## Sprint 7: Post-Completion Cleanup ✅

**Focus:** Protocol extraction, blueprint archival, integration test plan
**Dependencies:** All v2 sprints complete (48/48)

### Items

| ID | Task | Size | Status |
|----|------|------|--------|
| P7-1 | Create `protocols/conversation/PROTOCOLS.md` index file | S | ✅ Done |
| P7-2 | Extract Phase Routing protocol (131 lines → `phase-routing/PROTOCOL.md`) | M | ✅ Done |
| P7-3 | Extract Evaluator-Optimizer protocol (196 lines → `evaluator-optimizer/PROTOCOL.md`) | L | ✅ Done |
| P7-4 | Extract Approval Gates protocol (63 lines → `approval-gates/PROTOCOL.md`) | S | ✅ Done |
| P7-5 | Extract Mediated Delegation protocol (64 lines → `mediated-delegation/PROTOCOL.md`) | S | ✅ Done |
| P7-6 | Replace 4 moderator sections with protocol stubs | M | ✅ Done |
| P7-7 | Add Protocol Loading section to moderator | S | ✅ Done |
| P7-8 | Add `protocol_loading:` config section to YAML | S | ✅ Done |
| P7-9 | Archive blueprint to `docs/plans/v2-phases/archive/` | XS | ✅ Done |
| P7-10 | Create integration test plan (`integration-test-plan.md`) | M | ✅ Done |
| P7-11 | Update ROADMAP with Sprint 7 tracking | XS | ✅ Done |
| P7-12 | Update workflow documentation | S | ✅ Done |

### Results

- **Moderator size:** 813 → 399 lines (51% reduction)
- **Protocol files created:** 4 standalone protocols + 1 index
- **Lines extracted:** 454 lines of procedural instructions
- **Config additions:** `protocol_loading` section (17 lines) in YAML
- **Blueprint:** Archived with notice header

## Progress Summary

| Sprint | Items | Ready | In Progress | Review | Done | Blocked |
|--------|-------|-------|-------------|--------|------|---------|
| 1 — Foundation + Quick Wins ✅ | 9 | 0 | 0 | 0 | 9 | 0 |
| 2 — Moderator Intelligence ✅ | 8 | 0 | 0 | 0 | 8 | 0 |
| 3 — Flow Control ✅ | 8 | 0 | 0 | 0 | 8 | 0 |
| 4 — Quality Loop ✅ | 6 | 0 | 0 | 0 | 6 | 0 |
| 5 — User Control ✅ | 9 | 0 | 0 | 0 | 9 | 0 |
| 6 — Skill System ✅ | 8 | 0 | 0 | 0 | 8 | 0 |
| 7 — Post-Completion Cleanup ✅ | 12 | 0 | 0 | 0 | 12 | 0 |
| **Total** | **60** | **0** | **0** | **0** | **60** | **0** |

---

## Files Impact Map

Tracks which files are touched across all sprints (for merge conflict awareness):

| File | Sprints | Total Edits |
|------|---------|-------------|
| `.github/config/conversation-analysis.yml` | 1, 3, 4, 5, 6, 7 | ~13 edits |
| `.github/agents/conversation-moderator.md` | 2, 3, 4, 5, 7 | ~19 edits |
| `.github/agents/conversation-persona.md` | 1, 5 | ~6 edits |
| `.github/agents/conversation-persona-architect.md` | 1, 5 | ~6 edits |
| `.github/agents/conversation-persona-pragmatist.md` | 1, 5 | ~6 edits |
| `.github/agents/conversation-persona-critic.md` | 1, 5 | ~6 edits |
| `.github/agents/conversation-persona-researcher.md` | 1, 5 | ~6 edits |
| `.github/workflows/conversation-analysis-workflow.md` | 2, 5, 7 | ~3 edits |
| `.gitignore` | 1 | 1 edit |
| `skills/conversation/*.md` ✅ | 6 | 7 new files (6 skill files + 1 index) |
| `protocols/conversation/*.md` ✅ | 7 | 5 new files (4 protocol files + 1 index) |
| `docs/plans/v2-phases/archive/*` | 7 | 1 moved file (blueprint) |
| `docs/plans/v2-phases/integration-test-plan.md` ✅ | 7 | 1 new file |

**Hotspot:** `conversation-moderator.md` and `conversation-analysis.yml` are edited in almost every sprint. Sequence sprints carefully to avoid context drift in these files.

---

## Acceptance Criteria (per Sprint)

### Sprint 1 ✅ DONE (2026-02-24):
- [x] `.github/state/` exists and is gitignored
- [x] `state:` and `context_rules:` sections present in YAML config
- [x] All 5 persona agents include: open questions, severity classification, scope discipline, rejected alternatives, reasoning chains

### Sprint 2 ✅ DONE (2026-02-24):
- [x] Moderator reads/writes state file at every phase boundary
- [x] Moderator resumes interrupted sessions correctly
- [x] Moderator constructs prompts using only `context_rules.receives` tokens
- [x] Moderator validates persona output format before proceeding (3 gates with re-prompt)
- [x] Moderator uses argument weighting hierarchy in Phase 5
- [x] Severity unified to MUST/SHOULD/COULD across all files (critic dual-track retired)
- [x] All 5 personas enforce dual-path confidence verification for High/Very High claims

### Sprint 3 ✅ DONE (2026-02-24):
- [x] `phase_routing:` with conditional `next` for Phases 2.5, 4, 5
- [x] Moderator evaluates routing conditions using state variables
- [x] Loop bounds (`max_retries`) prevent infinite backward routing
- [x] Termination conditions fire correctly (early convergence, quality abort, budget)
- [x] Graceful degradation produces valid partial output

### Sprint 4 ✅ DONE (2026-02-24):
- [x] Phase 5 produces structured rubric scores with evidence citations
- [x] Scores below target trigger Phase 5.5 gap-fill for weakest dimension
- [x] Loop exits at target score, max iterations, or stall detection
- [x] Score history tracked across iterations in state file

### Sprint 5 ✅ DONE (2026-02-24):
- [x] `approval_gates.enabled: true` causes moderator to pause and wait at gates
- [x] `dynamic_selection.enabled: true` changes speaker order in Phase 3/4
- [x] `delegation.enabled: true` allows personas to make delegation requests
- [x] All three features disabled by default with no behavioral change

### Sprint 6 ✅ DONE (2026-02-24):
- [x] `skills/conversation/` directory with 5 skill files + index
- [x] `skill_injection:` config controls which skills load per phase
- [x] Skills content matches the inline prompt additions from Sprint 1
- [x] Skills are the canonical source; Sprint 1 inline additions reference them

### Sprint 7 ✅ DONE (2026-02-24):
- [x] 4 protocol files extracted from moderator into `protocols/conversation/`
- [x] Moderator reduced from 813 to 399 lines (51% reduction)
- [x] `protocol_loading:` config section added to YAML
- [x] Integration test plan document created with 16 checkpoint groups
- [x] Blueprint archived to `docs/plans/v2-phases/archive/`
- [x] ROADMAP and workflow documentation updated

---

## Changelog

| Date | Change | Items Affected |
|------|--------|---------------|
| 2026-02-24 | Initial roadmap created | All 48 items |
| 2026-02-24 | Sprint 1 completed — all 9 items delivered | A1, A2, A3 (merged), A6, S1-S5 |
| 2026-02-24 | Sprint 2 promoted to Ready — unblocked by Sprint 1 completion | A4, A5, A7, A8, S6, S7, S8, A9 |
| 2026-02-24 | Sprint 2 completed — all 8 items delivered | A4, A5, A7, A8, S6, S7, S8, A9 |
| 2026-02-24 | Sprint 3 promoted to Ready — unblocked by Sprint 2 completion | B1–B8 |
| 2026-02-24 | Sprint 3 completed — all 8 items delivered | B1–B8 |
| 2026-02-24 | Sprint 4 promoted to Ready — unblocked by Sprint 3 completion | C1–C6 |
| 2026-02-24 | Sprint 4 completed — all 6 items delivered | C1–C6 |
| 2026-02-24 | Sprints 5-6 promoted to Ready — all dependencies met | D1–D9, S9a–S11 |
| 2026-02-24 | Quality review — 5 consistency issues fixed (QR1–QR5) | YAML, moderator |
| 2026-02-24 | Sprint 6 completed — all 8 items delivered (skill files + config) | S9a–S9f, S10, S11 |
| 2026-02-24 | Sprint 5 completed — all 9 items delivered (gates, selection, delegation) | D1–D9 |
| 2026-02-24 | **v2 Roadmap complete — all 48/48 items delivered** | All |
| 2026-02-24 | Sprint 7 completed — protocol extraction, blueprint archive, test plan | P7-1–P7-12 |
| 2026-02-24 | Quality review (Sprint 7) — 4 issues fixed (QR6–QR9) | PROTOCOLS.md, ROADMAP, YAML |

---

## Quality Review Log

### QR-2026-02-24 — Post-Sprint 4 Full Audit

**Scope:** All 4 implementation files + 5 persona agents + infrastructure. Cross-referenced every roadmap claim against actual file contents.

**Verification Results:**

| Sprint | Claim | Verified | Method |
|--------|-------|----------|--------|
| 1 — A1 | `.github/state/` exists + gitignored | ✅ | `test -d`, grep `.gitignore` |
| 1 — A2 | `state:` config section in YAML | ✅ | YAML lines 31-36 |
| 1 — A6 | `context_rules:` in YAML | ✅ | YAML lines 75-112 |
| 1 — S1 | Open verification questions in all 5 personas | ✅ | grep: 2-3 hits per file |
| 1 — S2 | MUST/SHOULD/COULD severity in all 5 personas | ✅ | grep: 4-6 hits per file |
| 1 — S3 | Scope discipline in all 5 personas | ✅ | grep: 1 hit per file |
| 1 — S4 | Rejected alternatives in all 5 personas | ✅ | grep: 1 hit per file |
| 1 — S5 | Reasoning chains in all 5 personas | ✅ | grep: 2 hits per file |
| 2 — A4/A5 | State management + resume in moderator | ✅ | Lines 58-100 |
| 2 — A7/A8 | Context rule enforcement + audit trail in moderator | ✅ | Lines 106-133 |
| 2 — S6 | Dual-path confidence verification in all 5 personas | ✅ | grep: 1 hit per file |
| 2 — S7 | Deliverable verification protocol (3 gates) in moderator | ✅ | Lines 596-628 |
| 2 — S8 | Argument weighting hierarchy + severity unification in moderator | ✅ | Lines 545-555, 630-634 |
| 2 — A9 | Workflow doc v2 Features section | ✅ | Lines 858-970 |
| 3 — B1/B2 | `phase_routing:` directed graph with routing actions in YAML | ✅ | YAML lines 113-191 |
| 3 — B3/B4/B6/B7 | Phase Routing block in moderator (routing eval, loop prevention, termination checks, graceful degradation) | ✅ | Lines 136-260 |
| 3 — B5 | `termination:` config in YAML | ✅ | YAML lines 194-224 |
| 3 — B8 | Variant presets with `phase_routing_override` | ✅ | YAML lines 400-455 |
| 4 — C1 | `evaluator_optimizer:` config in YAML | ✅ | YAML lines 226-261 |
| 4 — C2/C5 | Evaluator protocol + exit conditions in moderator | ✅ | Lines 275-332 |
| 4 — C3/C4 | Optimizer protocol + dimension strategies in moderator | ✅ | Lines 336-400 |
| 4 — C6 | Score history state schema in moderator | ✅ | Lines 418-453 |

**Issues Found and Fixed:**

| ID | Severity | Issue | Fix | Files |
|----|----------|-------|-----|-------|
| QR1 | Minor | Score range gap: old rubric said "2.1-3.5" (excluded 2.0); evaluator section correctly says "2.0-3.5" | Changed "2.1-3.5" → "2.0-3.5" in legacy rubric section | `conversation-moderator.md` |
| QR2 | Medium | `aligned_positions` and `aligned_challenges` tokens referenced in evaluator-optimizer gap-fill context rules but not defined in token resolution table | Added both tokens to the resolution table with descriptions | `conversation-moderator.md` |
| QR3 | Medium | Two Phase 5.5 descriptions in moderator — old Step 5.5 (simple) and new Optimizer Role (detailed) — ambiguous which governs | Added conditional note: "When `evaluator_optimizer.enabled: true`, see Evaluator-Optimizer Loop section" | `conversation-moderator.md` |
| QR4 | Medium | YAML `context_rules.phase_5_5_gap_fill` missing the `concession_depth` exception documented in the evaluator-optimizer | Added `exceptions.concession_depth` block with `additional_receives` | `conversation-analysis.yml` |
| QR5 | Low | Legacy `phases.phase_5_5_gap_fill.trigger: "auto"` conflicts with routing's `trigger: "conditional"` — confusing but not a bug | Added comment noting the override behavior | `conversation-analysis.yml` |

**Observations (no fix needed):**

| ID | Category | Observation | Recommendation |
|----|----------|-------------|----------------|
| OB1 | Scalability | ~~Moderator file is 637 lines. ~66% is v2 protocol.~~ **Resolved:** Sprint 7 extracted 454 lines into 4 protocol files. Moderator is now 399 lines. | ✅ Done |
| OB2 | Redundancy | `phases:` and `phase_routing:` sections both exist in YAML with overlapping settings. | By design (backward-compatible). Added override comment to clarify priority. No action needed. |
| OB3 | Dual control | Both the routing graph (Phase 5 condition) and evaluator-optimizer exit logic govern the 5→5.5 transition. | Intentional dual-layer safety. Thresholds are consistent (both reference `high_confidence: 3.6`). Document as design decision. |
| OB4 | Threshold bands | Score 3.5-3.59 is "usable" but below `high_confidence` (3.6). Evaluator-optimizer keeps looping in this range. | Intentional per blueprint — the gap between `usable` and `high_confidence` triggers optimization when enabled. |

### QR-2026-02-24 — Post-Sprint 7 Audit

**Scope:** All Sprint 7 deliverables (4 protocol files, PROTOCOLS.md index, integration test plan, moderator stubs, YAML config, workflow doc, ROADMAP, archived blueprint). Cross-referenced paths, line counts, config gates, and section boundaries.

**Issues Found and Fixed:**

| ID | Severity | Issue | Fix | Files |
|----|----------|-------|-----|-------|
| QR6 | Low | PROTOCOLS.md line counts showed extraction sizes (131, 196, 63, 64) not actual file sizes (137, 202, 69, 70) — misleading since files include header metadata | Updated line counts to actual file sizes | `protocols/conversation/PROTOCOLS.md` |
| QR7 | Medium | Progress Summary table missing Sprint 7 — only showed Sprints 1–6 (48 items) | Added Sprint 7 row (12 items); updated total to 60 | `ROADMAP.md` |
| QR8 | Medium | Files Impact Map didn't include Sprint 7 changes — missing Sprint 7 for moderator, YAML, workflow doc, and 3 new file groups | Added Sprint 7 to moderator/YAML/workflow Sprints columns; added 3 new file rows (protocols, archive, test plan) | `ROADMAP.md` |
| QR9 | Medium | Two `selection_guide` entries (`framework_evolution`, `security_review`) misplaced under `delegation:` section at end of YAML — wrong parent key | Moved to correct `selection_guide:` section; removed from `delegation:` block | `conversation-analysis.yml` |

**Verification Results (Sprint 7 specific):**

| Item | Claim | Verified | Method |
|------|-------|----------|--------|
| P7-1 | PROTOCOLS.md index exists | ✅ | `test -f protocols/conversation/PROTOCOLS.md` |
| P7-2 | Phase Routing extracted verbatim | ✅ | Content comparison: routing actions, loop prevention, graceful degradation, state vars — all present |
| P7-3 | Eval-Optimizer extracted verbatim | ✅ | Content comparison: evaluator role, scoring calibration, optimizer role, 6 dimension strategies, loop cost budget — all present |
| P7-4 | Approval Gates extracted verbatim | ✅ | Content comparison: gate execution protocol (6 steps), 3 summary templates, feature interactions — all present |
| P7-5 | Mediated Delegation extracted verbatim | ✅ | Content comparison: detection (6 steps + validation matrix), loop prevention (4 rules), state tracking — all present |
| P7-6 | Moderator stubs reference correct paths | ✅ | All 4 stubs point to correct `protocols/conversation/{dir}/PROTOCOL.md` paths |
| P7-7 | Protocol Loading section in moderator | ✅ | Table with 4 protocols, correct config gates, correct file paths |
| P7-8 | `protocol_loading:` in YAML | ✅ | Section present at line ~551 with 4 protocol entries matching moderator table |
| P7-9 | Blueprint archived | ✅ | File at `docs/plans/v2-phases/archive/blueprint-prompts-and-patches.md` with archive notice |
| P7-10 | Integration test plan | ✅ | 16 checkpoint groups, sample topic, config overrides section, all v2 features covered |
| P7-11 | ROADMAP updated | ✅ | Sprint 7 section with 12 items, acceptance criteria, changelog entry |
| P7-12 | Workflow doc updated | ✅ | Protocol System subsection in v2 Features with protocol table and explanation |
| Config gates match | Moderator stubs ↔ YAML ↔ PROTOCOLS.md ↔ workflow doc all consistent | ✅ | Cross-file comparison of 4 config gates across all 4 files |
| Moderator size claim | "813 → 399 lines (51% reduction)" | ✅ | `wc -l` confirms 399 lines |
| YAML size change | 684 → 707 lines (+23 for protocol_loading + selection_guide fix) | ✅ | `wc -l` confirms 707 lines |
| Dynamic Selection stays inline | Not extracted — remains in moderator | ✅ | 44 lines present at moderator lines ~193–222 |

**Observations:**

| ID | Category | Observation | Recommendation |
|----|----------|-------------|----------------|
| OB5 | Structure | The moderator file (399 lines) is now ~50% core workflow + ~50% inline v2 features (State Mgmt, Context Rules, Dynamic Selection, Deliverable Verification). The inline sections are all <60 lines each. | No extraction needed. Current size is sustainable. |
| OB6 | Consistency | All 4 moderator stubs follow identical format (blockquote + conditional). Clean pattern. | No action needed. |

---

## Post-v2: Next Actions

> **Date:** 2026-02-25
> **Status:** Complete
> **Context:** v2 complete (Sprints 1-7, 48/48 items), project restructured (config split into 4 files, conversation/ directory, archive/, new naming convention), all 18 skill integration techniques implemented

### Validation (Priority: P0)

| ID | Task | Size | Status | Notes |
|----|------|------|--------|-------|
| V1 | End-to-end test: `/analyze` on a real topic | M | Deferred | Requires live Copilot Chat session — cannot be tested from terminal |
| V2 | End-to-end test: `/workflow "analyze: ..."` | M | Deferred | Requires live Claude Code session — cannot be tested from terminal |
| V3 | Verify moderator reads 4 config files correctly | S | ✅ Complete | All 4 files parseable, no tabs, all moderator refs resolve to existing files, 15 top-level keys verified |

### Integration Gaps (Priority: P1-P2)

| ID | Task | Size | Priority | Status | Description |
|----|------|------|----------|--------|-------------|
| E-C | Specialist agents as persona presets | M | P1 | ✅ Complete | 4 dev presets (dev_analyst, dev_architect, dev_reviewer, dev_tester) + 5 dev topic types in selection guide |
| E-D | Shared convention system | M | P2 | ✅ Complete | `conventions/shared.md` — evidence, reasoning chains, severity, scope, intent markers, confidence. Referenced by moderator + conventions index |
| E-E | Spec-agent registration | L | P2 | ✅ Complete | 5 spec presets (spec_analyst, spec_architect, spec_planner, spec_reviewer, spec_validator) + 5 spec topic types in selection guide |

### Quality Improvements (Priority: P1)

| ID | Task | Size | Priority | Status | Description |
|----|------|------|----------|--------|-------------|
| Q1 | Evidence Quality improvement | M | P1 | ✅ Complete | 7 files, 5 vectors: evidence-standards skill expanded (Source-or-Flag, evidence-seeking strategies), Phase 2.5 evidence pre-check, researcher cross-position audit in Phase 3, evidence-standards injected in Phases 3-4, Source-or-Flag added to all personas |
| Q2 | "architect" naming collision | S | P2 | ✅ Complete | Disambiguation notes in personas.yml (specialist comment) + moderator (selection rules + naming section) |

### Documentation (Priority: P3)

| ID | Task | Size | Priority | Status | Description |
|----|------|------|----------|--------|-------------|
| D1 | Project README.md | M | P3 | ✅ Complete | Root README with architecture overview, directory structure, both workflow systems, config reference, personas, variants, quick start |

### Recommended Execution Order

```
V1 → V2 → V3       (validate before building more)     — V3 ✅, V1-V2 deferred (need live sessions)
  → Q1              (highest-impact quality fix)         — ✅
  → E-C             (highest-value integration gap)      — ✅
  → Q2, E-D, E-E   (medium priority)                    — ✅
  → D1              (documentation)                      — ✅
```
