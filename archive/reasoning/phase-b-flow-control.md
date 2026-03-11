# Phase B: Flow Control — Conditional Routing & Early Termination

> **Parent:** [conversation-workflow-v2-improvements.md](../2026-02-24-conversation-workflow-v2-improvements.md)
> **Implements:** G3 (Conditional Phase Routing) + G5 (Early Termination Conditions)
> **Depends on:** Phase A (G1 State File required for condition evaluation)
> **Priority:** High — transforms fixed pipeline into self-correcting graph

---

## Table of Contents

1. [Overview](#overview)
2. [G3: Conditional Phase Routing](#g3-conditional-phase-routing)
3. [G5: Early Termination Conditions](#g5-early-termination-conditions)
4. [Interaction Between G3 and G5](#interaction-between-g3-and-g5)
5. [Implementation Sequence](#implementation-sequence)
6. [Validation Criteria](#validation-criteria)

---

## Overview

Phase B transforms the conversation from a fixed pipeline to an adaptive graph. Instead of "execute all configured phases in order," the moderator now evaluates conditions at each phase boundary and routes accordingly.

**Current behavior (v1):**
```
Phase 2 → 2.5 → 2.7 → 3 → 4 → 5 → 5.5 → 6
(always runs this exact sequence)
```

**Target behavior (v2):**
```
Phase 2 → 2.5 → [CONDITION: diversity sufficient?]
  YES → 3 → 4 → [CONDITION: concessions occurred?]
    YES → 5 → [CONDITION: quality threshold met?]
      YES → 6
      NO → 5.5 → 5 (loop, max 2)
    NO → [CONDITION: under max_rounds?]
      YES → 3 (repeat with escalated challenge depth)
      NO → 5 (proceed with quality warning)
  NO → [ACTION: inject devil's advocate] → 2 (re-run)
```

### Why This Matters

The three most common quality failures in the current workflow are:

1. **Zero-concession conversations** — Personas generate positions, challenge each other, but nobody actually changes their mind. The rebuttal phase produces "I acknowledge your point but maintain my position" for every challenge. Without conditional routing, this goes straight to convergence, producing a low-quality synthesis.

2. **Premature consensus** — All personas arrive at the same recommendation in Phase 2. The diversity audit catches this, but currently just flags it. With conditional routing, the moderator can automatically inject a contrarian persona and re-run.

3. **Quality below threshold** — Phase 5 convergence produces a synthesis that scores below the usable threshold on the quality rubric. Currently, Phase 5.5 gap-fill is a single-pass attempt. With conditional routing, the gap-fill can iterate.

---

## G3: Conditional Phase Routing

### Condition Grammar

Conditions are natural language expressions that the moderator evaluates by reading the state file and phase outputs. They are designed to be unambiguous and testable.

#### State Variables

These variables are available for condition evaluation. The moderator reads them from the state file (G1):

| Variable | Type | Source | Example |
|----------|------|--------|---------|
| `effective_persona_count` | integer | Phase 2.5 diversity audit | `3` |
| `min_personas` | integer | Config: moderation.min_personas | `3` |
| `similarity_max` | float | Phase 2.5 highest pairwise similarity | `0.72` |
| `similarity_threshold` | float | Config: phase_2_5.similarity_threshold | `0.70` |
| `all_positions_agree` | boolean | Phase 2.5 — all positions share same recommendation | `false` |
| `round` | integer | Current cross-exam/rebuttal round number | `1` |
| `max_rounds` | integer | Config: moderation.max_rounds | `2` |
| `concession_count` | integer | Phase 4 — total concessions across all personas | `0` |
| `max_severity_challenge` | enum | Phase 3 — highest severity challenge issued | `"critical"` |
| `challenge_depth` | enum | Current challenge depth setting | `"evidence"` |
| `quality_score` | float | Phase 5 — overall quality rubric score | `2.8` |
| `quality_dimensions` | map | Phase 5 — per-dimension quality scores | `{perspective_diversity: 4, ...}` |
| `re_run_threshold` | float | Config: quality_rubric.thresholds.re_run | `2.0` |
| `high_confidence_threshold` | float | Config: quality_rubric.thresholds.high_confidence | `3.6` |
| `sub_agent_calls` | integer | State: metrics.sub_agent_calls | `14` |
| `sub_agent_budget` | integer | Config: termination.token_budget.max_sub_agent_calls | `25` |
| `iteration` | integer | Evaluator-optimizer loop iteration (Phase C) | `1` |
| `max_iterations` | integer | Config: evaluator_optimizer.max_iterations | `2` |

#### Condition Syntax

Conditions use simple comparison expressions. The moderator evaluates them by reading state values and applying the comparison.

```
<variable> <operator> <value>
<condition> AND <condition>
<condition> OR <condition>
```

Operators: `==`, `!=`, `<`, `>`, `<=`, `>=`

Examples:
```
effective_persona_count < min_personas
all_positions_agree == true
concession_count == 0 AND round < max_rounds
quality_score >= high_confidence_threshold
max_severity_challenge == "critical" AND round < max_rounds
sub_agent_calls >= sub_agent_budget
```

> **Evaluation note:** The moderator evaluates these as natural language reasoning, not code execution. Conditions must be simple enough that an LLM can reliably evaluate them. Complex boolean logic (nested AND/OR) should be avoided — decompose into multiple conditions with priority ordering instead.

### Phase Routing Configuration

Replace the flat `phases` config with a directed graph:

```yaml
# ─── PHASE ROUTING (replaces flat phase list) ─────────────────
phase_routing:

  phase_2_opening:
    enabled: true
    parallel: true
    next: phase_2_5_diversity_audit

  phase_2_5_diversity_audit:
    enabled: true
    similarity_threshold: 0.70
    evidence_overlap_threshold: 0.50
    next:
      default: phase_2_7_tension_extraction
      conditions:
        - id: "insufficient_diversity"
          when: "effective_persona_count < min_personas"
          action: "inject_contrarian_persona"
          description: "Too few effective personas — inject a devil's advocate and re-run opening positions for the new persona only"
          then: phase_2_opening
          max_retries: 1            # Prevent infinite loops

        - id: "premature_consensus"
          when: "all_positions_agree == true"
          action: "inject_devil_advocate"
          description: "All positions agree — add an explicit contrarian persona targeting the consensus recommendation"
          then: phase_2_opening
          max_retries: 1

  phase_2_7_tension_extraction:
    enabled: true
    next: phase_3_cross_examination

  phase_3_cross_examination:
    enabled: true
    challenge_depth: "evidence"
    next: phase_4_rebuttal

  phase_4_rebuttal:
    enabled: true
    next:
      default: phase_5_convergence
      conditions:
        - id: "zero_concessions"
          when: "concession_count == 0 AND round < max_rounds"
          action: "escalate_challenge_depth"
          description: "No concessions occurred — escalate challenge depth and run another round"
          then: phase_3_cross_examination
          side_effects:
            - set: challenge_depth
              to: "formal"            # Escalate from evidence → formal

        - id: "critical_challenges_unresolved"
          when: "max_severity_challenge == 'critical' AND round < max_rounds"
          action: "repeat_round"
          description: "Critical challenges remain — run another cross-exam + rebuttal round"
          then: phase_3_cross_examination

  phase_5_convergence:
    enabled: true
    quality_rubric: true
    falsifiability: true
    next:
      default: phase_6_output
      conditions:
        - id: "quality_below_rerun"
          when: "quality_score < re_run_threshold"
          action: "abort_quality_failure"
          description: "Quality too low to salvage — abort with quality failure"
          then: null                  # Terminates session (G5)

        - id: "quality_below_usable"
          when: "quality_score < high_confidence_threshold AND iteration < max_iterations"
          action: "trigger_gap_fill"
          description: "Quality below high-confidence — run gap-fill optimization"
          then: phase_5_5_gap_fill

  phase_5_5_gap_fill:
    enabled: true
    trigger: "conditional"            # Only runs when routed to
    next: phase_5_convergence         # Loop back for re-evaluation

  phase_6_output:
    enabled: true
    output_path: "docs/analysis/conversation-analysis-output-{{date}}.md"
    next: null                        # Terminal node
```

### Routing Actions

Each condition specifies an `action` that the moderator executes before routing:

| Action | Description | Moderator Behavior |
|--------|-------------|-------------------|
| `inject_contrarian_persona` | Add a new persona optimized for disagreement | Select or generate a persona from the preset library whose priorities conflict with the consensus. Add to persona roster. Run Phase 2 for this persona only. |
| `inject_devil_advocate` | Add explicit devil's advocate | Configure a persona with the sole objective of attacking the consensus recommendation. Use the `critic` specialist or generate a custom contrarian. |
| `escalate_challenge_depth` | Increase rigor of challenges | Change `challenge_depth` from `assertion` → `evidence` → `formal`. Log the escalation in routing_log. |
| `repeat_round` | Re-run cross-exam + rebuttal | Increment `round` counter. Re-run Phase 3 and Phase 4 with existing personas and updated challenge depth. |
| `trigger_gap_fill` | Run Phase 5.5 gap-fill | Identify weakest quality dimension. Generate targeted gap-fill personas. |
| `abort_quality_failure` | Terminate with failure | Set `session.status = "aborted"`, `termination.reason = "quality_failure"`. Produce partial output with caveats. |

### Loop Prevention

Every conditional route that loops backward MUST have a bound:

```yaml
# Each condition that creates a backward edge must specify ONE of:
max_retries: 1          # Maximum times this condition can trigger a loop
# OR
max_rounds: 2           # Global round limit (from moderation config)
# OR
max_iterations: 2       # From evaluator_optimizer config (Phase C)
```

The moderator tracks loop counts in the state file and refuses to route backward once the bound is reached, falling through to the `default` route instead.

### Routing Decision Log

Every routing decision is recorded in the state file:

```yaml
routing_log:
  - phase: "phase_2_5_diversity_audit"
    timestamp: "2026-02-24T14:46:00Z"
    conditions_evaluated:
      - id: "insufficient_diversity"
        variable_values: {effective_persona_count: 3, min_personas: 3}
        result: false
      - id: "premature_consensus"
        variable_values: {all_positions_agree: false}
        result: false
    route_taken: "default → phase_2_7_tension_extraction"

  - phase: "phase_4_rebuttal"
    timestamp: "2026-02-24T15:10:00Z"
    conditions_evaluated:
      - id: "zero_concessions"
        variable_values: {concession_count: 0, round: 1, max_rounds: 2}
        result: true
    route_taken: "zero_concessions → phase_3_cross_examination"
    action_executed: "escalate_challenge_depth (evidence → formal)"
```

### Routing Evaluation Algorithm

```
FUNCTION evaluate_routing(phase_id, state, config):
  1. route_config = config.phase_routing[phase_id]

  2. IF route_config.next is a string (not a map):
     → RETURN route_config.next  # Unconditional — always goes to this phase

  3. FOR EACH condition IN route_config.next.conditions (in order):
     a. Read variable values from state file
     b. Evaluate condition.when using variable values
     c. IF condition evaluates to TRUE:
        - Check loop bounds (max_retries, max_rounds)
        - IF bound exceeded:
          → LOG "Condition {id} would trigger but bound exceeded"
          → CONTINUE to next condition
        - Execute condition.action
        - Log to routing_log
        - RETURN condition.then
     d. IF condition evaluates to FALSE:
        → LOG condition evaluation result
        → CONTINUE to next condition

  4. No condition matched → RETURN route_config.next.default
```

> **Priority rule:** Conditions are evaluated **in order**. The first matching condition wins. Design condition lists with the most specific/critical conditions first.

---

## G5: Early Termination Conditions

### Termination Types

| Type | Trigger | Behavior |
|------|---------|----------|
| **Early convergence** | All positions agree before cross-examination | Skip to Phase 5 with quality warning |
| **Quality abort** | Quality score below re-run threshold | Abort — produce partial output with failure explanation |
| **Token budget** | Sub-agent call count exceeds budget | Graceful degradation — complete current phase, skip to convergence |
| **User abort** | User explicitly requests stop | Save state, produce partial output from completed phases |
| **Stall detection** | Two consecutive loops with no quality improvement | Break loop, proceed with best available output |

### Termination Configuration

```yaml
# ─── TERMINATION CONDITIONS ────────────────────────────────────
termination:

  # Early convergence — all positions agree, skip debate phases
  early_convergence:
    enabled: true
    check_after:
      - phase_2_5_diversity_audit
    condition: "all_positions_agree == true AND effective_persona_count < min_personas"
    action: "skip_to_convergence"
    quality_flag: "premature_agreement"
    note: |
      Premature agreement is a quality concern, not a success signal.
      The convergence analysis should note that no adversarial pressure
      was applied and confidence should be downgraded accordingly.

  # Quality abort — conversation is unrecoverable
  quality_abort:
    enabled: true
    check_after:
      - phase_5_convergence
    conditions:
      - when: "quality_score < re_run_threshold"
        action: "abort"
        reason: "Quality below minimum threshold ({quality_score} < {re_run_threshold})"
      - when: "perspective_diversity_score == 1 AND concession_count == 0"
        action: "abort"
        reason: "No diversity and no concessions — conversation produced no dialectic value"

  # Token budget — hard cap on resource usage
  token_budget:
    max_sub_agent_calls: 25
    warning_at: 20
    on_exceed: "graceful_degradation"  # graceful_degradation | hard_stop
    graceful_degradation_behavior: |
      1. Complete the current sub-agent call
      2. Skip all remaining optional phases
      3. If Phase 5 not yet completed, run Phase 5 with available data
      4. Produce output with "budget_exceeded" flag

  # Stall detection — loop is not improving quality
  stall_detection:
    enabled: true
    trigger: "quality_improvement < 0.2 across 2 consecutive iterations"
    action: "break_loop"
    note: "Diminishing returns detected — further iteration unlikely to improve output"
```

### Termination Evaluation Points

The moderator evaluates termination conditions at specific checkpoints:

```
                     ┌─────────────────────┐
                     │   TERMINATION CHECK  │
                     │   (at each ◆ node)   │
                     └─────────────────────┘

Phase 2 ──◆──→ Phase 2.5 ──◆──→ Phase 2.7 ──→ Phase 3 ──→ Phase 4 ──◆──→ Phase 5 ──◆──→ Phase 6
          │                 │                                          │              │
          │                 │                                          │              │
          ▼                 ▼                                          ▼              ▼
     token budget    early convergence                           token budget   quality abort
                     premature consensus                         stall detect   token budget
                                                                                stall detect
```

The moderator runs this check **after each phase completes, before routing to the next phase**:

```
FUNCTION check_termination(state, config):
  1. Check token budget:
     IF state.metrics.sub_agent_calls >= config.termination.token_budget.warning_at:
       → LOG warning: "Approaching token budget ({current}/{max})"
     IF state.metrics.sub_agent_calls >= config.termination.token_budget.max_sub_agent_calls:
       → Execute termination.token_budget.on_exceed behavior
       → RETURN TERMINATE

  2. Check phase-specific termination conditions:
     FOR EACH termination_rule IN config.termination:
       IF current_phase IN termination_rule.check_after:
         → Evaluate termination_rule.condition against state
         → IF TRUE: Execute termination_rule.action
         → RETURN TERMINATE or SKIP

  3. Check stall detection (if in a loop):
     IF state.phases.phase_5_convergence.iteration > 1:
       → Compare quality_score with previous iteration
       → IF improvement < stall_threshold:
         → RETURN BREAK_LOOP

  4. No termination triggered → RETURN CONTINUE
```

### Graceful Degradation Output

When termination occurs before Phase 6, the moderator produces a partial output:

```markdown
# Conversation Analysis: {topic}

> ⚠️ **Partial Output** — Session terminated: {termination.reason}
> Completed phases: {list of completed phases}
> Skipped phases: {list of skipped phases}

## Termination Details
{termination.details}

## Available Analysis
{whatever convergence/synthesis was possible from completed phases}

## Caveats
- This analysis did not complete the full dialectic process
- {specific caveats based on what was skipped}
- Confidence should be rated lower than a fully completed analysis

## Recommendation
{if enough data exists for a recommendation, provide it with appropriate confidence downgrade}
{if not, state "Insufficient dialectic data for a recommendation"}
```

### Token Budget Tracking

The moderator increments `metrics.sub_agent_calls` each time it invokes a persona sub-agent. The budget check happens before each invocation:

```
FUNCTION before_sub_agent_call(state, config):
  current = state.metrics.sub_agent_calls
  budget = config.termination.token_budget.max_sub_agent_calls
  warning = config.termination.token_budget.warning_at

  IF current >= budget:
    → Trigger graceful degradation
    → DO NOT spawn sub-agent
    → RETURN BLOCKED

  IF current >= warning:
    → Log: "Token budget warning: {current}/{budget} calls used"

  → Increment state.metrics.sub_agent_calls
  → RETURN ALLOWED
```

---

## Interaction Between G3 and G5

G3 (routing) and G5 (termination) both evaluate conditions after each phase. They interact in this priority order:

```
1. FIRST: Check G5 termination conditions
   → If TERMINATE: stop immediately, no routing
   → If SKIP: jump to specified phase, skip routing evaluation

2. THEN: Check G3 routing conditions
   → If condition matches: route to specified phase
   → If no match: take default route

3. ALWAYS: Log both evaluations in the state file
```

Termination takes priority over routing. If the token budget is exceeded, no routing condition can override that.

### Example: Complete Flow with Both Active

```
Phase 2 completes
  → G5 check: token_budget OK (6/25) → CONTINUE
  → G3 check: (no conditions on Phase 2) → default route to Phase 2.5

Phase 2.5 completes (effective_persona_count = 2, min_personas = 3)
  → G5 check: early_convergence? all_positions_agree = false → CONTINUE
  → G5 check: token_budget OK (6/25) → CONTINUE
  → G3 check: insufficient_diversity? 2 < 3 = TRUE
    → Action: inject_contrarian_persona
    → Route: phase_2_opening (re-run for new persona only)

Phase 2 re-run completes (now 3 effective personas)
  → G5 check: token_budget OK (7/25) → CONTINUE
  → G3 check: (no conditions) → default route to Phase 2.5

Phase 2.5 re-check completes (effective_persona_count = 3)
  → G5 check: CONTINUE
  → G3 check: insufficient_diversity? 3 >= 3 = FALSE
  → G3 check: premature_consensus? false → FALSE
  → G3: default route → Phase 2.7

... [phases 2.7 through 4] ...

Phase 4 completes (concession_count = 0, round = 1, max_rounds = 2)
  → G5 check: token_budget OK (17/25) → CONTINUE
  → G3 check: zero_concessions? 0 == 0 AND 1 < 2 = TRUE
    → Action: escalate_challenge_depth (evidence → formal)
    → Route: phase_3_cross_examination (round 2)

Phase 3 round 2 completes
  → G5 check: token_budget WARNING (20/25) → LOG WARNING, CONTINUE
  → G3 check: (no conditions) → default route to Phase 4

Phase 4 round 2 completes (concession_count = 2)
  → G5 check: token_budget OK (23/25) → CONTINUE
  → G3 check: zero_concessions? 2 != 0 → FALSE
  → G3 check: critical_challenges? max_severity = "significant" → FALSE
  → G3: default route → Phase 5

Phase 5 completes (quality_score = 3.2)
  → G5 check: quality_abort? 3.2 >= 2.0 → CONTINUE
  → G5 check: token_budget OK (24/25) → CONTINUE
  → G3 check: quality_below_rerun? 3.2 >= 2.0 → FALSE
  → G3 check: quality_below_usable? 3.2 < 3.6 AND 1 < 2 = TRUE
    → Action: trigger_gap_fill
    → Route: phase_5_5_gap_fill

Phase 5.5 completes
  → G5 check: token_budget EXCEEDED (26/25) → GRACEFUL DEGRADATION
    → Skip remaining phases
    → Route directly to Phase 6 with available data
    → Flag output with "budget_exceeded"
```

---

## Implementation Sequence

### Step 1: Ensure Phase A is complete

G3 and G5 depend on the state file (G1) for:
- Variable values for condition evaluation
- Loop counter tracking
- Routing decision logging
- Sub-agent call counting

### Step 2: Add phase_routing to YAML config

Replace the flat `phases` section with the `phase_routing` directed graph structure. Maintain backward compatibility by keeping the old `phases` section as a fallback.

```yaml
# ─── PHASE CONFIGURATION (legacy — used when phase_routing is absent) ───
phases:
  # ... existing flat config ...

# ─── PHASE ROUTING (v2 — used when present, overrides phases) ──────────
phase_routing:
  # ... directed graph config ...
```

### Step 3: Add termination config to YAML

Add the `termination` section to `conversation-analysis.yml`.

### Step 4: Update moderator agent instructions

Add to `conversation-moderator.md`:

1. **Routing evaluation block** — After each phase, evaluate routing conditions
2. **Termination check block** — Before each sub-agent call, check termination conditions
3. **Loop tracking block** — Maintain round/iteration counters in state
4. **Priority rules** — Termination checks before routing checks

### Step 5: Update variant presets

Each variant preset should map to a phase_routing configuration:

```yaml
variants:
  quick:
    phase_routing_override:
      phase_3_cross_examination:
        next: phase_5_convergence     # Skip rebuttal
      phase_4_rebuttal:
        enabled: false
```

### Step 6: Validate with test scenarios

Run conversations designed to trigger each condition:
- Near-identical positions → triggers diversity injection
- Zero concessions → triggers round escalation
- Budget exhaustion → triggers graceful degradation
- Low quality score → triggers gap-fill loop

---

## Validation Criteria

### G3 Pass Criteria

| Criterion | Test |
|-----------|------|
| Conditions evaluated after each phase | Routing log shows evaluation entries for every phase transition |
| Correct variable resolution | Variable values in routing log match state file values |
| First matching condition wins | When multiple conditions match, first one's route is taken |
| Loop bounds respected | Backward routes never exceed max_retries/max_rounds |
| Actions executed correctly | Side effects (e.g., challenge_depth escalation) reflected in state |
| Default route used when no condition matches | Routing log shows "default → {phase}" when all conditions are false |

### G5 Pass Criteria

| Criterion | Test |
|-----------|------|
| Token budget tracked correctly | metrics.sub_agent_calls matches actual call count |
| Budget warning fires at threshold | Log message appears when warning_at is reached |
| Budget exceeded triggers degradation | Session produces partial output when budget is exceeded |
| Quality abort fires correctly | Session aborts when quality_score < re_run_threshold |
| Stall detection breaks loops | Loop exits when quality improvement < threshold for 2 iterations |
| Partial output is well-formed | Terminated sessions produce valid output with caveats |

---

*Phase B makes the workflow intelligent. Combined with Phase A's foundation, the conversation can now adapt to its own quality signals.*
