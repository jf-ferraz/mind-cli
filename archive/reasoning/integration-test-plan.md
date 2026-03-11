# v2 Integration Test Plan

> **Purpose:** Verify all v2 features work together in a single end-to-end conversation analysis run.
> **Status:** Ready for manual execution
> **Scope:** All 48 v2 items across Sprints 1–6 + Sprint 7 protocol extraction

---

## Test Topic Brief

Use this sample topic to exercise the full feature surface:

```
Should a notification system be built as a standalone microservice or integrated into the existing monolith?

Consider: real-time requirements, team expertise, infrastructure cost, deployment complexity, scaling patterns, failure isolation, and migration risk.
```

This topic is effective because it:
- Has genuinely strong arguments on both sides (exercises concession depth)
- Spans technical and organizational dimensions (exercises persona diversity)
- Has well-known tradeoffs with empirical data (exercises evidence quality)
- Admits phased/hybrid solutions (exercises synthesis quality + actionability)

---

## Config Overrides

Enable all opt-in features for maximum coverage. Apply these overrides at workflow start:

```yaml
# All opt-in features ON
evaluator_optimizer:
  enabled: true
  target_score: 3.5          # Low enough to exit in 1–2 iterations
  max_iterations: 2

approval_gates:
  enabled: true               # Will pause at configured gates

dynamic_selection:
  enabled: true

delegation:
  enabled: true

# Use the thorough variant for full routing exercise
variants:
  selected: "thorough"

# Skills ON (default)
skill_injection:
  enabled: true

# Protocols ON (default after Sprint 7)
protocol_loading:
  enabled: true
```

---

## Phase-by-Phase Verification Checklist

### CP-1: Configuration Loading

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 1.1 | Config parsed | Sprint 1 | State file created with all 18 top-level config sections loaded |
| 1.2 | Variant applied | Sprint 2 | If variant selected, merged overrides visible in effective config |
| 1.3 | Protocol loading | Sprint 7 | Moderator reads `protocol_loading` section, loads 4 protocol files |
| 1.4 | Skill index loaded | Sprint 6 | `skills/conversation/SKILLS.md` parsed, skill rules available |

### CP-2: Phase 1 — Opening (Persona Generation)

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 2.1 | Topic type detected | Sprint 2 | `topic_type.detected` set in state (e.g., "architecture_decision") |
| 2.2 | Personas generated | Sprint 1 | ≥ `min_personas` personas created with name, perspective, priorities |
| 2.3 | Skill injection | Sprint 6 | Phase 1–tagged skills injected into persona prompts (check for scope-discipline markers) |

### CP-3: Phase 2 — Position Development

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 3.1 | Positions produced | Sprint 1 | Each persona outputs a structured position paper |
| 3.2 | Context rules | Sprint 2 | Phase 2 personas receive only `topic_brief` + `persona_definition` (not other positions) |
| 3.3 | Dynamic selection | Sprint 5 | Speaker order logged in `selection_log` with strategy + rationale |
| 3.4 | Delegation scan | Sprint 5 | Each output scanned for `### Delegation Requests` heading |

### CP-4: Phase 2.5 — Diversity Audit

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 4.1 | Diversity audit runs | Sprint 3 | `effective_persona_count` and `all_positions_agree` set in state |
| 4.2 | Premature agreement | Sprint 3 | If all agree, routing injects contrarian persona |

### CP-5: Approval Gate 1

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 5.1 | Gate triggers | Sprint 5 | Position Summary table presented, workflow pauses |
| 5.2 | User choice processed | Sprint 5 | Gate interaction logged in state under `gates` |
| 5.3 | Dynamic selection info | Sprint 5 | Gate summary includes speaker order if `dynamic_selection.enabled` |

### CP-6: Phase 3 — Cross-Examination

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 6.1 | Challenges issued | Sprint 1 | Each persona issues challenges with severity levels (MUST/SHOULD/COULD) |
| 6.2 | Challenge depth | Sprint 3 | `challenge_depth` from state used (assertion/evidence/formal) |
| 6.3 | Skill injection | Sprint 6 | Phase 3–tagged skills injected (challenge-methodology, evidence-standards) |
| 6.4 | Dynamic selection | Sprint 5 | Next speaker selected based on strategy, not fixed order |

### CP-7: Phase 4 — Rebuttal & Concession

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 7.1 | Rebuttals produced | Sprint 1 | Each persona responds to challenges with rebuttals + concessions |
| 7.2 | Concession tracking | Sprint 2 | `concession_count` updated in state, Concession Trail maintained |
| 7.3 | Routing evaluation | Sprint 3 | After Phase 4, routing checks conditions (e.g., zero concessions → repeat round) |
| 7.4 | Loop prevention | Sprint 3 | Retry counts tracked in `routing_log`, max_retries enforced |
| 7.5 | Delegation responses | Sprint 5 | Any approved delegation responses injected into rebuttal prompts |

### CP-8: Approval Gate 2

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 8.1 | Gate triggers | Sprint 5 | Concession Summary presented, workflow pauses |
| 8.2 | Delegation info | Sprint 5 | Gate summary includes delegation status if `delegation.enabled` |

### CP-9: Phase 5 — Convergence & Evaluation

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 9.1 | Convergence produced | Sprint 1 | Synthesis document with Key Insights, Recommendations |
| 9.2 | Quality scoring | Sprint 4 | All 6 dimensions scored (1–5) with evidence + justification |
| 9.3 | Calibration anchors | Sprint 4 | Scores reference calibration table descriptions |
| 9.4 | Score history | Sprint 4 | `phases.phase_5_convergence.history` has iteration entry |
| 9.5 | Exit decision | Sprint 4 | Correct exit path taken (pass/gap-fill/max-iter/stall) |

### CP-10: Phase 5.5 — Gap-Fill (if triggered)

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 10.1 | Optimizer dispatches | Sprint 4 | Weakest dimension identified, matching strategy executed |
| 10.2 | Gap-fill context rules | Sprint 4 | Gap-fill persona receives only permitted context |
| 10.3 | Reintegration | Sprint 4 | Gap-fill output appended to convergence as addendum |
| 10.4 | Re-evaluation | Sprint 4 | Phase 5 re-runs with updated input, new scores recorded |
| 10.5 | Stall detection | Sprint 3 | If improvement < threshold, loop breaks with quality flag |

### CP-11: Approval Gate 3

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 11.1 | Gate triggers | Sprint 5 | Quality Summary presented (score, strongest/weakest dimension) |
| 11.2 | User can inject challenge | Sprint 5 | `inject_challenge` option available and functional |

### CP-12: Phase 6 — Final Output

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 12.1 | Output produced | Sprint 1 | Final analysis document with all required sections |
| 12.2 | Quality flags | Sprint 4 | Any quality flags from eval-optimizer visible in output metadata |
| 12.3 | Deliverable verification | Sprint 2 | All required deliverable sections present per config |

### CP-13: Termination & Budget

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 13.1 | Budget tracking | Sprint 3 | `metrics.sub_agent_calls` incremented for each call |
| 13.2 | Budget warning | Sprint 3 | Warning logged at `warning_at` threshold |
| 13.3 | Budget includes delegation | Sprint 5 | Delegation calls counted in budget |

### CP-14: State File Integrity

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 14.1 | All state sections | Sprint 2 | `conversation-state.yml` has: personas, phases, metrics, routing_log, gates, selection_log, delegations |
| 14.2 | Routing log complete | Sprint 3 | Every phase transition logged with condition ID + action |
| 14.3 | Phase outputs saved | Sprint 1 | All phase outputs saved to `.github/state/phase-{N}/` |

### CP-15: Protocol System Integrity

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 15.1 | Phase routing loaded | Sprint 7 | Moderator follows phase-routing protocol from extracted file |
| 15.2 | Eval-optimizer loaded | Sprint 7 | Moderator follows eval-optimizer protocol from extracted file |
| 15.3 | Approval gates loaded | Sprint 7 | Moderator follows approval-gates protocol from extracted file |
| 15.4 | Mediated delegation loaded | Sprint 7 | Moderator follows mediated-delegation protocol from extracted file |
| 15.5 | Behavior unchanged | Sprint 7 | Identical workflow behavior to pre-extraction moderator |

### CP-16: Cross-Feature Interactions

| # | Check | Feature | Pass Criteria |
|---|-------|---------|---------------|
| 16.1 | Gates + dynamic selection | Sprint 5 | Gate summaries reflect dynamic speaker ordering |
| 16.2 | Gates + delegation | Sprint 5 | Gate summaries include delegation status |
| 16.3 | Routing + eval-optimizer | Sprint 3+4 | Routing conditions correctly reference quality scores |
| 16.4 | Skills + phases | Sprint 6 | Correct skills injected per phase tag (not leaked across phases) |
| 16.5 | Delegation + budget | Sprint 5 | Delegation calls reduce remaining budget for all features |
| 16.6 | Variant + routing | Sprint 2+3 | Variant routing overrides correctly merged with base graph |

---

## Execution Notes

- **Duration:** A full run with all features enabled uses ~15–20 sub-agent calls
- **Manual gates:** The 3 approval gates require user interaction — select "continue" unless testing gate-specific actions
- **Expected quality:** With the microservice topic, expect quality scores of 3.0–4.0 on first evaluation
- **Gap-fill likelihood:** ~60% chance the evaluator triggers at least one gap-fill iteration (depends on persona quality)

## Known Limitations

- This test plan is a manual checklist, not automated tests
- LLM non-determinism means exact scores and concession counts vary between runs
- The plan verifies feature presence, not output quality (that requires human judgment)
- Protocol extraction (Sprint 7) should produce identical behavior — any deviation indicates a regression
