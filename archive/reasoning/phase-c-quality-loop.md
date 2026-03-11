# Phase C: Quality Loop — Evaluator-Optimizer Pattern

> **Parent:** [conversation-workflow-v2-improvements.md](../2026-02-24-conversation-workflow-v2-improvements.md)
> **Implements:** G4 (Evaluator-Optimizer Loop)
> **Depends on:** Phase A (G1 State File), Phase B (G3 Conditional Routing)
> **Priority:** Medium — upgrades gap-fill from single-pass to iterative refinement

---

## Table of Contents

1. [Overview](#overview)
2. [The Evaluator-Optimizer Pattern](#the-evaluator-optimizer-pattern)
3. [Evaluator: Quality Rubric Scoring Protocol](#evaluator-quality-rubric-scoring-protocol)
4. [Optimizer: Targeted Gap-Fill Protocol](#optimizer-targeted-gap-fill-protocol)
5. [Loop Control and Exit Conditions](#loop-control-and-exit-conditions)
6. [Dimension-Specific Optimization Strategies](#dimension-specific-optimization-strategies)
7. [Implementation Sequence](#implementation-sequence)
8. [Validation Criteria](#validation-criteria)

---

## Overview

The current workflow has a primitive quality improvement mechanism:

```
Current (v1):
  Phase 5: Convergence → score with rubric
  Phase 5.5: Gap-Fill → single-pass correction (if triggered)
  Phase 6: Output
```

This is **open-loop** — Phase 5.5 fires once, and whatever it produces goes to output regardless of whether the gap was actually filled. There's no verification that the optimization worked.

The Evaluator-Optimizer pattern closes the loop:

```
Target (v2):
  Phase 5: Convergence → score with rubric (EVALUATOR)
    │
    ├── Score >= target → Phase 6 (exit loop)
    ├── Score < target AND iteration < max → Phase 5.5 (OPTIMIZER)
    │     │
    │     └── Target weakest dimension → generate corrective input
    │           │
    │           └── Return to Phase 5 with updated data (re-evaluate)
    │
    └── Score < target AND iteration >= max → Phase 6 (exit with caveats)
```

### Benchmark Reference

**LangGraph Evaluator-Optimizer:** An LLM evaluator scores the output against criteria. If below threshold, an optimizer LLM receives the evaluation feedback and produces a refined version. The loop repeats until the score threshold is met or max iterations reached. The key insight: the optimizer receives the **evaluation feedback** (what's wrong), not just the original input.

**Our adaptation:** The moderator plays both evaluator and optimizer roles. The quality rubric serves as the evaluation function. Gap-fill personas serve as the optimization mechanism. The evaluation feedback (which dimension is weakest and why) guides which gap-fill persona to spawn and what they should address.

---

## The Evaluator-Optimizer Pattern

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                EVALUATOR-OPTIMIZER LOOP                   │
│                                                           │
│  ┌─────────────────────┐                                  │
│  │    Phase 5:          │                                  │
│  │    EVALUATOR         │◄──────────────────────┐         │
│  │    (Convergence +    │                       │         │
│  │     Rubric Scoring)  │                       │         │
│  └──────────┬───────────┘                       │         │
│             │                                   │         │
│        Score >= target?                         │         │
│        ┌────┴────┐                              │         │
│       YES       NO                              │         │
│        │         │                              │         │
│        │    iteration < max?                    │         │
│        │    ┌────┴────┐                         │         │
│        │   YES       NO                         │         │
│        │    │         │                          │         │
│        │    ▼         │                          │         │
│        │  ┌──────────────────┐                  │         │
│        │  │   Phase 5.5:      │                  │         │
│        │  │   OPTIMIZER       │──────────────────┘         │
│        │  │   (Targeted       │   (returns to evaluator   │
│        │  │    Gap-Fill)      │    with new data)         │
│        │  └──────────────────┘                            │
│        │         │                                         │
│        ▼         ▼                                         │
│  ┌─────────────────────┐                                  │
│  │    Phase 6:          │                                  │
│  │    OUTPUT            │                                  │
│  │    (+ quality flags) │                                  │
│  └─────────────────────┘                                  │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### State Tracking

The loop state is tracked in the state file (G1):

```yaml
# In conversation-state.yml
phases:
  phase_5_convergence:
    status: "completed"
    iteration: 2                    # Current iteration (starts at 1)
    history:                        # Score history across iterations
      - iteration: 1
        quality_score: 2.8
        weakest_dimension: "perspective_diversity"
        weakest_score: 2
        action: "triggered_gap_fill"
      - iteration: 2
        quality_score: 3.7
        weakest_dimension: "evidence_quality"
        weakest_score: 3
        action: "passed_threshold"
    quality_scores:                 # Latest scores
      overall: 3.7
      perspective_diversity: 4
      evidence_quality: 3
      concession_depth: 4
      challenge_substantiveness: 4
      synthesis_quality: 4
      actionability: 3

  phase_5_5_gap_fill:
    status: "completed"
    iterations:
      - iteration: 1
        trigger: "perspective_diversity < 3"
        personas_spawned: ["devil_advocate_for_consensus"]
        output_paths: [".github/state/phase-5.5/iter-1/devil-advocate-position.md"]
        quality_improvement: 0.9    # Score delta from gap-fill
```

---

## Evaluator: Quality Rubric Scoring Protocol

### The 6-Dimension Rubric

The evaluator (moderator in Phase 5) scores the conversation across 6 dimensions. Each dimension has a 1-5 scale with anchored descriptions.

| Dimension | What It Measures | 1 (Poor) | 3 (Adequate) | 5 (Excellent) |
|-----------|-----------------|----------|--------------|---------------|
| **Perspective Diversity** | Did personas genuinely disagree? | All agree from start | Moderate tension, 1+ contrarian | Genuine philosophical conflict |
| **Evidence Quality** | Is argumentation backed by data? | Opinions only | Mix of empirical + anecdotal | Multiple independent empirical sources |
| **Concession Depth** | Did positions actually change? | Zero concessions | Adjustments on minor points | Fundamental recommendation revision |
| **Challenge Substantiveness** | Were challenges meaningful? | Superficial or absent | Challenges identify real weaknesses | Counter-evidence + alternatives provided |
| **Synthesis Quality** | Did convergence produce insight? | Lists all positions | Identifies consensus + disagreements | Emergent insights beyond any position |
| **Actionability** | Can the reader act on this? | Vague direction | Clear recommendation | Phased plan with falsifiability |

### Scoring Protocol

The moderator scores each dimension using this structured reasoning:

```markdown
### Quality Rubric Evaluation (Iteration {N})

#### 1. Perspective Diversity: {score}/5
**Evidence:** {cite specific examples of agreement/disagreement from position papers}
**Justification:** {why this score, referencing the anchor descriptions}

#### 2. Evidence Quality: {score}/5
**Evidence:** {cite specific evidence claims from the Evidence Registry}
**Justification:** {assessment of empirical vs. anecdotal ratio}

#### 3. Concession Depth: {score}/5
**Evidence:** {cite from Concession Trail — specific position changes}
**Justification:** {were concessions substantive or superficial}

#### 4. Challenge Substantiveness: {score}/5
**Evidence:** {cite from Phase 3 challenges — specific counter-arguments}
**Justification:** {did challenges provide counter-evidence or just disagree}

#### 5. Synthesis Quality: {score}/5
**Evidence:** {cite specific emergent insights from convergence analysis}
**Justification:** {are insights novel or just summaries of positions}

#### 6. Actionability: {score}/5
**Evidence:** {cite recommendations — specificity and conditional detail}
**Justification:** {can a reader act on these without further analysis}

#### Overall Score: {average}/5
#### Weakest Dimension: {dimension} ({score})
#### Optimization Target: {what to address in gap-fill, if triggered}
```

### Score Interpretation

| Score Range | Label | Action |
|-------------|-------|--------|
| **< 2.0** | Unusable | If first iteration: trigger gap-fill. If second iteration: abort (conversation is structurally broken). |
| **2.0 – 3.5** | Usable with caveats | Trigger gap-fill if iterations remain. Otherwise, proceed with quality flag. |
| **3.6 – 5.0** | High confidence | Proceed to Phase 6. No gap-fill needed. |

### Scoring Calibration Guidelines

To reduce scoring inconsistency across sessions:

**Perspective Diversity:**
- Score 1: Zero challenges in Phase 3 address a genuine disagreement
- Score 3: At least one "critical" severity challenge exists
- Score 5: At least two personas have irreconcilable core values on the topic

**Evidence Quality:**
- Score 1: No Evidence Registry entries across all position papers
- Score 3: ≥ 50% of key arguments have Evidence Registry entries with "moderate" or better strength
- Score 5: ≥ 80% of key arguments backed by "strong" empirical evidence with replication

**Concession Depth:**
- Score 1: Concession Trail in Phase 5 is empty
- Score 3: At least 2 concessions, at least 1 affecting a "Key Argument" (not just a minor point)
- Score 5: At least 1 persona changed their primary recommendation based on challenges

**Challenge Substantiveness:**
- Score 1: Challenges are "I disagree because" with no counter-evidence
- Score 3: Challenges cite specific weaknesses with reasoning
- Score 5: Challenges provide alternative solutions or counter-data, not just criticism

**Synthesis Quality:**
- Score 1: Convergence analysis is a list of "Persona A said X, Persona B said Y"
- Score 3: Convergence identifies non-obvious patterns across positions
- Score 5: At least 1 key insight that no individual persona articulated

**Actionability:**
- Score 1: Recommendation is "more research needed" or equivalent
- Score 3: Recommendation specifies an approach with rationale
- Score 5: Recommendation includes phased implementation, decision criteria, and falsifiability conditions

---

## Optimizer: Targeted Gap-Fill Protocol

### Gap Identification

The optimizer's first task is identifying **what to fix**. It reads the evaluator's rubric scores and targets the weakest dimension(s).

```
FUNCTION identify_gaps(quality_scores, config):
  1. Sort dimensions by score (ascending)
  2. FOR EACH dimension with score < config.evaluator_optimizer.target_score:
     → Add to gap_list with:
       - dimension name
       - current score
       - target score
       - improvement_priority (lower score = higher priority)
  3. RETURN top N gaps (where N = min(2, len(gap_list)))
     → Cap at 2 to prevent scope explosion per iteration
```

### Gap-Fill Persona Generation

For each identified gap, the optimizer generates a targeted persona:

| Weak Dimension | Gap-Fill Persona Strategy |
|----------------|--------------------------|
| **Perspective Diversity** | Spawn a persona whose priorities directly conflict with the consensus position. Use one from the preset library (critic, devil's advocate) or generate a custom contrarian. |
| **Evidence Quality** | Spawn the Researcher persona (or a custom "Evidence Auditor") with explicit instructions to find empirical data supporting OR contradicting the primary recommendation. |
| **Concession Depth** | Re-run Phase 3-4 for the two most aligned personas with `challenge_depth: formal` and an explicit instruction: "You MUST find at least one substantive point to concede." |
| **Challenge Substantiveness** | Spawn the Critic persona targeting the PRIMARY recommendation specifically. Instructions must demand counter-evidence, not just objections. |
| **Synthesis Quality** | Re-run Phase 5 convergence with an additional prompt: "Identify at least one emergent insight that none of the individual positions articulated. Look for patterns across the concession trail." |
| **Actionability** | Re-run Phase 5 convergence with an additional prompt: "For each recommendation, specify: (a) implementation phases, (b) decision criteria for proceeding/stopping, (c) falsifiability condition." |

### Gap-Fill Execution

```
FUNCTION execute_gap_fill(gaps, state, config):
  FOR EACH gap IN gaps:
    1. Determine persona and strategy from the table above
    2. Construct gap-fill prompt:
       - Persona definition
       - The CURRENT convergence analysis (so gap-fill persona knows what exists)
       - The SPECIFIC gap being addressed
       - Phase 2 Opening Position format (for new perspective gaps)
       - OR targeted critique format (for depth/quality gaps)
    3. Spawn sub-agent
    4. Save output to .github/state/phase-5.5/iter-{N}/{persona}-output.md
    5. Update state file

  RETURN gap_fill_outputs
```

### Gap-Fill Context Rules

Gap-fill personas receive tailored context (per G2 context rules):

```yaml
context_rules:
  phase_5_5_gap_fill:
    receives:
      - topic_brief
      - persona_definition
      - convergence_analysis          # Current Phase 5 output
      - gap_identification            # Which dimension is weak and why
    blocked:
      - individual_positions          # Don't bias with existing positions
      - individual_challenges         # Fresh perspective required
      - individual_rebuttals
    rationale: >
      Gap-fill personas need to know WHAT the conversation concluded
      (to address gaps) but NOT the individual argumentation (to bring
      genuinely new perspective).
    exception:
      dimension: "concession_depth"
      additional_receives:
        - aligned_positions           # For concession-depth, must see the two
        - aligned_challenges          # most similar positions to challenge them
```

### Reintegration

After gap-fill outputs are generated, they're reintegrated into the convergence:

```
FUNCTION reintegrate_gap_fill(gap_outputs, existing_convergence, state):
  1. READ existing convergence analysis from state
  2. FOR EACH gap_output:
     → Append as an addendum section:
       "## Gap-Fill: {dimension} (Iteration {N})"
       "{gap-fill persona output}"
  3. WRITE updated convergence input back to state
  4. INCREMENT state.phases.phase_5_convergence.iteration
  5. Route back to Phase 5 for re-evaluation
```

The moderator re-runs Phase 5 convergence with the gap-fill outputs included as additional input. This produces a new rubric score. If the score now meets the threshold, the loop exits.

---

## Loop Control and Exit Conditions

### Exit Decision Matrix

| Condition | Action | Output Flag |
|-----------|--------|-------------|
| `quality_score >= target_score` | Exit loop → Phase 6 | None (high confidence) |
| `iteration >= max_iterations AND quality_score >= usable_threshold` | Exit loop → Phase 6 | `"max_iterations_reached"` |
| `iteration >= max_iterations AND quality_score < usable_threshold` | Exit loop → Phase 6 | `"below_usable_threshold"` |
| `quality_improvement < 0.2 for 2 iterations` | Exit loop (stall) → Phase 6 | `"stall_detected"` |
| `quality_score < re_run_threshold` | Abort session | `"quality_failure"` |
| `sub_agent_calls >= budget` | Exit loop (budget) → Phase 6 | `"budget_exceeded"` |

### Diminishing Returns Detection

Track quality score history across iterations:

```yaml
# In state file
phases:
  phase_5_convergence:
    history:
      - iteration: 1
        quality_score: 2.8
      - iteration: 2
        quality_score: 3.0    # Improvement: 0.2
      - iteration: 3
        quality_score: 3.1    # Improvement: 0.1 — diminishing returns
```

If the improvement between consecutive iterations drops below 0.2 for two consecutive iterations, the loop is stalling. Continuing will waste tokens without meaningful quality gain.

### Maximum Token Cost Per Loop

| Iteration Step | Sub-Agent Calls |
|---------------|-----------------|
| Phase 5 evaluation | 0 (moderator does it) |
| Phase 5.5 gap-fill | 1-2 (per identified gap) |
| Phase 5 re-evaluation | 0 (moderator does it) |
| **Per iteration total** | **1-2 calls** |
| **Max loop cost (2 iterations)** | **2-4 calls** |

The loop is lightweight — each iteration adds 1-2 sub-agent calls. With `max_iterations: 2`, the maximum additional cost is 4 sub-agent calls. This is well within the 25-call budget.

---

## Dimension-Specific Optimization Strategies

### Strategy: Perspective Diversity (Score < 3)

**Diagnosis:** Positions are too similar. The diversity audit may have caught this, but insufficient correction was applied.

**Optimization:**
1. Identify the consensus position (recommendation shared by ≥ 2 personas)
2. Generate a contrarian persona definition:
   ```yaml
   gap_persona:
     name: "The Contrarian (Gap-Fill)"
     perspective: "Directly oppose {consensus recommendation}. Find the strongest case AGAINST it."
     priorities:
       - Identifying fatal flaws in {consensus}
       - Finding superior alternatives not yet considered
       - Stress-testing the key evidence supporting {consensus}
   ```
3. Spawn with Phase 2 template — generate an independent position opposing the consensus
4. Spawn follow-up: brief cross-examination between contrarian and strongest consensus advocate

**Expected improvement:** +1 to +2 points on perspective diversity

### Strategy: Evidence Quality (Score < 3)

**Diagnosis:** Arguments are opinion-based or cite weak/anecdotal evidence.

**Optimization:**
1. Extract all claims from the convergence analysis that lack strong evidence
2. Spawn the Researcher persona (specialist) with targeted instructions:
   ```
   Evaluate the following claims. For each, either:
   a) Find supporting empirical evidence (with source, methodology, strength)
   b) Find contradicting empirical evidence
   c) Declare the claim unfalsifiable or untestable
   ```
3. Feed evidence findings back into convergence as an Evidence Appendix

**Expected improvement:** +1 to +1.5 points on evidence quality

### Strategy: Concession Depth (Score < 3)

**Diagnosis:** No persona changed their mind. The debate was theater.

**Optimization:**
1. Identify the two most aligned personas (highest similarity from diversity audit)
2. Re-run a targeted Phase 3 for just these two personas with escalated rules:
   ```
   challenge_depth: formal
   Special instruction: "You MUST identify at least one point where
   the other position is genuinely stronger than yours. Concede it
   explicitly and update your recommendation if warranted."
   ```
3. Feed concession results into convergence

**Expected improvement:** +1 to +2 points on concession depth. This is the hardest dimension to improve because it requires genuine position change, which LLMs tend to avoid.

### Strategy: Challenge Substantiveness (Score < 3)

**Diagnosis:** Challenges were "I disagree" without counter-evidence or alternatives.

**Optimization:**
1. Spawn the Critic persona with instructions to attack the primary recommendation
2. Require format: "For each challenge, you MUST provide: (a) counter-evidence citing specific data, (b) at least one alternative approach, (c) a concrete scenario where the recommendation fails"
3. Feed substantive challenges into convergence

**Expected improvement:** +1 to +1.5 points on challenge substantiveness

### Strategy: Synthesis Quality (Score < 3)

**Diagnosis:** Convergence analysis just lists positions without producing new insight.

**Optimization:**
1. Re-run Phase 5 with additional prompting:
   ```
   Focus specifically on:
   a) What do ALL personas' concessions have in common? (meta-pattern)
   b) Where did a challenge from one persona strengthen a DIFFERENT
      persona's position? (cross-pollination)
   c) What option does NO persona recommend, but which the full
      conversation arc implicitly supports? (emergent recommendation)
   ```
2. This is moderator self-improvement — no additional sub-agents needed

**Expected improvement:** +1 point on synthesis quality

### Strategy: Actionability (Score < 3)

**Diagnosis:** Recommendations are vague or lack implementation detail.

**Optimization:**
1. Re-run Phase 5 recommendation section with requirements:
   ```
   Each recommendation must include:
   a) Phased implementation plan (Phase 1: ..., Phase 2: ...)
   b) Decision gates between phases (proceed if ..., stop if ...)
   c) Falsifiability condition (this recommendation is wrong if ...)
   d) Resource requirements (time, cost, expertise)
   e) Rollback plan (if this fails, the next-best option is ...)
   ```
2. This is moderator self-improvement — no additional sub-agents needed

**Expected improvement:** +1 to +1.5 points on actionability

---

## Configuration

```yaml
# ─── EVALUATOR-OPTIMIZER LOOP ──────────────────────────────────
evaluator_optimizer:
  enabled: true
  max_iterations: 2               # Maximum optimization loops
  target_score: 3.6               # Quality rubric threshold to exit
  stall_threshold: 0.2            # Minimum improvement per iteration
  stall_patience: 2               # Consecutive stalled iterations before exit

  # Which dimensions to optimize, in priority order
  # Only the top 2 weakest dimensions are addressed per iteration
  optimization_priority:
    - perspective_diversity
    - concession_depth
    - evidence_quality
    - challenge_substantiveness
    - synthesis_quality
    - actionability

  # Per-dimension optimization limits
  dimension_limits:
    perspective_diversity:
      max_gap_fill_personas: 1    # Max personas to spawn for this dimension
      preferred_persona: "critic"  # Specialist to prefer for this gap
    evidence_quality:
      max_gap_fill_personas: 1
      preferred_persona: "researcher"
    concession_depth:
      strategy: "re_run_aligned_pair"  # Special strategy — doesn't spawn new persona
      max_retries: 1
    challenge_substantiveness:
      max_gap_fill_personas: 1
      preferred_persona: "critic"
    synthesis_quality:
      strategy: "moderator_self_improve"  # No sub-agent needed
    actionability:
      strategy: "moderator_self_improve"
```

---

## Implementation Sequence

### Step 1: Verify Phase A and B prerequisites

- G1 state file supports iteration tracking and score history
- G3 routing supports Phase 5 → 5.5 → 5 conditional loop

### Step 2: Add evaluator_optimizer config to YAML

Add the full configuration section above to `conversation-analysis.yml`.

### Step 3: Update moderator agent — Evaluator role

Add to `conversation-moderator.md` Phase 5 instructions:

```markdown
### Phase 5 Evaluator Protocol

After producing the Convergence Analysis, immediately score it using the
Quality Rubric (6 dimensions, 1-5 each).

1. Score each dimension with structured evidence and justification
2. Calculate overall score (average)
3. Record scores in state file at phases.phase_5_convergence.quality_scores
4. Append to state file phases.phase_5_convergence.history

EXIT CHECK:
- IF overall_score >= evaluator_optimizer.target_score → proceed to Phase 6
- IF iteration >= evaluator_optimizer.max_iterations → proceed to Phase 6 with quality flag
- IF stall detected (improvement < stall_threshold for stall_patience iterations) → proceed to Phase 6 with stall flag
- ELSE → route to Phase 5.5 with weakest dimension identification
```

### Step 4: Update moderator agent — Optimizer role

Add to `conversation-moderator.md` Phase 5.5 instructions:

```markdown
### Phase 5.5 Optimizer Protocol

You receive: weakest dimension identification from the evaluator.

1. Read dimension_limits from config for the weakest dimension
2. Select strategy:
   - "moderator_self_improve" → re-run convergence section with targeted prompt
   - "re_run_aligned_pair" → re-run cross-exam for the two most similar personas
   - Default → spawn gap-fill persona from preferred_persona or generate custom
3. Execute strategy
4. Save outputs to .github/state/phase-5.5/iter-{N}/
5. Reintegrate outputs into Phase 5 input
6. Route back to Phase 5 for re-evaluation
```

### Step 5: Update workflow documentation

Add "Evaluator-Optimizer Loop" section to `conversation-analysis-workflow.md` with:
- Architecture diagram
- Score history tracking
- Dimension-specific strategies
- Exit conditions

### Step 6: Validate

Test scenarios:
- Conversation with score 4.0 on first pass → loop never triggers → Phase 6 directly
- Conversation with score 2.5 → gap-fill runs → re-score at 3.8 → exit loop
- Conversation with score 2.5 → gap-fill → score 2.7 → gap-fill → score 2.9 → exit with caveats
- Conversation with score 1.5 → abort (below re_run_threshold)

---

## Validation Criteria

| Criterion | Test |
|-----------|------|
| Rubric scores recorded in state file | quality_scores populated after Phase 5 |
| Score history tracks across iterations | history array grows with each iteration |
| Gap-fill targets weakest dimension | gap identification matches lowest-scored dimension |
| Loop exits at target score | Score ≥ 3.6 → no Phase 5.5 triggered |
| Loop exits at max iterations | After 2 iterations, proceeds to Phase 6 regardless |
| Stall detection works | Two iterations with < 0.2 improvement → exit with stall flag |
| Dimension strategies applied correctly | The right persona/strategy used per dimension |
| Reintegration doesn't overwrite | Gap-fill outputs appended, not replacing existing convergence |
| Output flags set correctly | Phase 6 output includes quality flags when applicable |
| Token budget respected during loop | Loop respects G5 token budget — exits if budget exceeded |

---

*Phase C closes the quality-assurance loop. Combined with Phase A and B, the workflow now self-evaluates and self-corrects until quality is sufficient or resources are exhausted.*
