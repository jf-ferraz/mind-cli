# Blueprint — Prompts, Scripts & Implementation Patches

> **⚠ ARCHIVED** — All 15 patches have been applied (Sprints 1–6). This file is preserved for historical reference only.
> **Location:** Moved to `docs/archive/conversation/` — no longer active reference.

> **Canonical reference:** [workflow-methodology.reference.md](../../reference/workflow-methodology.reference.md) — contains the v2 Features summary
> **Purpose:** Historical record of ready-to-apply patches for each v2 feature
> **Usage:** ~~Apply patches in order (Phase A → B → C → D).~~ All patches applied — see [ROADMAP.md](ROADMAP.md) for completion log.
> **Config note:** Patches below reference `.github/config/conversation-analysis.yml` — the actual config was implemented as 4 files in `conversation/config/`. See `conversation/README.md`.

---

## Table of Contents

1. [Phase A Patches](#phase-a-patches)
2. [Phase B Patches](#phase-b-patches)
3. [Phase C Patches](#phase-c-patches)
4. [Phase D Patches](#phase-d-patches)
5. [Complete v2 Config Template](#complete-v2-config-template)
6. [State File Template](#state-file-template)
7. [Moderator Instruction Blocks](#moderator-instruction-blocks)

---

## Phase A Patches

### Patch A1: State Management Config (conversation-analysis.yml)

Add after the `conversation:` section:

```yaml
# ─── STATE MANAGEMENT ─────────────────────────────────────────
state:
  enabled: true
  path: ".github/state"
  auto_cleanup: false             # If true, delete state dir after Phase 6
  resume_on_restart: true         # Check for existing state on moderator start
```

### Patch A2: Context Rules Config (conversation-analysis.yml)

Add after the `phases:` section:

```yaml
# ─── CONTEXT ISOLATION RULES ──────────────────────────────────
# Declarative rules controlling what each persona sees per phase.
# The moderator MUST consult these rules before constructing any
# sub-agent prompt. Only include specified context tokens.
#
# Context tokens:
#   topic_brief, persona_definition, own_position, other_positions,
#   all_positions, own_challenges, other_challenges, all_challenges,
#   challenges_received, own_rebuttal, all_rebuttals, tension_matrix,
#   diversity_audit, convergence_analysis, gap_identification, all
#
context_rules:
  phase_2_opening:
    receives: [topic_brief, persona_definition]
    blocked: [other_positions, all_challenges, all_rebuttals]
    rationale: "Independent position generation — no cross-contamination"

  phase_2_5_diversity_audit:
    actor: moderator
    receives: [all_positions]

  phase_2_7_tension_extraction:
    actor: moderator
    receives: [all_positions, diversity_audit]

  phase_3_cross_examination:
    receives: [topic_brief, persona_definition, other_positions]
    blocked: [own_position, all_challenges, all_rebuttals]
    rationale: "Engage with foreign ideas, not defend own"

  phase_4_rebuttal:
    receives: [persona_definition, own_position, challenges_received]
    blocked: [other_positions, other_challenges, all_rebuttals]
    rationale: "Focused defense/concession on challenges to own position"

  phase_5_convergence:
    actor: moderator
    receives: [all]

  phase_5_5_gap_fill:
    receives: [topic_brief, persona_definition, convergence_analysis, gap_identification]
    blocked: [all_positions, all_challenges, all_rebuttals]
    rationale: "Fresh perspective — knows conclusions but not individual arguments"
```

### Patch A3: Moderator State Management Block (conversation-moderator.md)

Add after the existing "## Configuration" section:

````markdown
## State Management (v2)

If `state.enabled: true` in the config:

### Session Start
1. Check for existing state file at `{state.path}/conversation-state.yml`
2. If exists and `session.status == "active"`:
   - Report: "Found interrupted session: {session.id}. Resuming from {current_phase}."
   - Read all existing phase outputs from state file paths
   - Resume from the first incomplete step
3. If no state file:
   - Create `.github/state/` directory
   - Initialize `conversation-state.yml` with session metadata
   - Set `session.status: "active"`

### Phase Boundaries
Before EVERY phase:
1. Read `conversation-state.yml`
2. Set current phase `status: "in_progress"` and `started_at`
3. Write updated state file

After EVERY sub-agent call:
1. Read `conversation-state.yml`
2. Append output entry to current phase's `outputs` list
3. Increment `metrics.sub_agent_calls`
4. Write updated state file

After EVERY phase completes:
1. Set phase `status: "completed"` and `completed_at`
2. Increment `metrics.phases_completed`
3. Write updated state file

### Phase Output Storage
Save each sub-agent's output as a separate file:
- Pattern: `.github/state/phase-{N}/{persona-id}-{output-type}.md`
- Record the file path in the state file's phase output entry
- The moderator reads these files (via `readFile`) when constructing subsequent prompts

### Session End
1. Set `session.status: "completed"` (or `"aborted"`)
2. Set `termination.reason`
3. Final state file write
````

### Patch A4: Moderator Context Rule Enforcement Block (conversation-moderator.md)

Add after the State Management block:

````markdown
## Context Rule Enforcement (v2)

If `context_rules` section exists in the config:

### Before Constructing ANY Sub-Agent Prompt
1. Read `context_rules` for the current phase
2. Resolve each token in `receives` to actual content:
   - `topic_brief` → read `.github/state/topic-brief.md`
   - `persona_definition` → read persona config from YAML
   - `own_position` → read `.github/state/phase-2/{persona-id}-position.md`
   - `other_positions` → read all `.github/state/phase-2/*.md` EXCEPT current persona
   - `challenges_received` → read challenges targeting this persona from Phase 3 outputs
   - `convergence_analysis` → read `.github/state/phase-5/convergence-analysis.md`
   - `all` → read all phase output files
3. Verify NO content matching `blocked` tokens is included
4. Construct sub-agent prompt using ONLY resolved `receives` content

### Audit Trail
For each sub-agent call, record in the state file:
```yaml
context_delivered:
  - token: "{token_name}"
    source: "{file_path}"
context_blocked:
  - token: "{token_name}"
    would_have_been: "{file_path}"
```
````

### Patch A5: Gitignore Addition

Add to project `.gitignore`:

```gitignore
# Conversation analysis ephemeral state
.github/state/
```

---

## Phase B Patches

### Patch B1: Phase Routing Config (conversation-analysis.yml)

Add after `context_rules:` section. This section coexists with the legacy `phases:` section — when present, `phase_routing` takes priority.

```yaml
# ─── PHASE ROUTING (v2 — directed graph) ──────────────────────
# When present, overrides the flat `phases` section.
# Each phase specifies its successor(s) with optional conditions.
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
          then: phase_2_opening
          max_retries: 1
        - id: "premature_consensus"
          when: "all_positions_agree == true"
          action: "inject_devil_advocate"
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
          then: phase_3_cross_examination
          side_effects:
            - set: challenge_depth
              to: "formal"
        - id: "critical_challenges_unresolved"
          when: "max_severity_challenge == 'critical' AND round < max_rounds"
          action: "repeat_round"
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
          then: null
        - id: "quality_needs_improvement"
          when: "quality_score < high_confidence_threshold AND iteration < max_iterations"
          action: "trigger_gap_fill"
          then: phase_5_5_gap_fill

  phase_5_5_gap_fill:
    enabled: true
    trigger: "conditional"
    next: phase_5_convergence

  phase_6_output:
    enabled: true
    output_path: "docs/analysis/conversation-analysis-output-{{date}}.md"
    next: null
```

### Patch B2: Termination Config (conversation-analysis.yml)

Add after `phase_routing:` section:

```yaml
# ─── TERMINATION CONDITIONS ────────────────────────────────────
termination:
  early_convergence:
    enabled: true
    check_after: [phase_2_5_diversity_audit]
    condition: "all_positions_agree == true AND effective_persona_count < min_personas"
    action: "skip_to_convergence"
    quality_flag: "premature_agreement"

  quality_abort:
    enabled: true
    check_after: [phase_5_convergence]
    conditions:
      - when: "quality_score < re_run_threshold"
        action: "abort"
      - when: "perspective_diversity_score == 1 AND concession_count == 0"
        action: "abort"

  token_budget:
    max_sub_agent_calls: 25
    warning_at: 20
    on_exceed: "graceful_degradation"

  stall_detection:
    enabled: true
    trigger: "quality_improvement < 0.2 across 2 consecutive iterations"
    action: "break_loop"
```

### Patch B3: Moderator Routing Block (conversation-moderator.md)

Add after the Context Rule Enforcement block:

````markdown
## Phase Routing (v2)

If `phase_routing` section exists in the config:

### After Each Phase Completes

1. **Termination check (priority 1):**
   - Read `termination` config
   - Check token budget: `metrics.sub_agent_calls` vs `termination.token_budget.max_sub_agent_calls`
   - Check phase-specific termination conditions
   - Check stall detection (if in an evaluator-optimizer loop)
   - If termination triggered: execute termination action and stop

2. **Routing evaluation (priority 2):**
   - Read `phase_routing[current_phase].next`
   - If `next` is a simple string: route to that phase
   - If `next` has `conditions`:
     - Evaluate each condition IN ORDER by reading state variables
     - First matching condition wins
     - Check loop bounds (`max_retries`, `max_rounds`) before routing backward
     - If bound exceeded: skip condition, evaluate next
     - If no condition matches: use `default` route
   - Execute the matching condition's `action` (if any)
   - Log the routing decision in `routing_log`

3. **Update state file** with routing decision and proceed to next phase

### Routing Actions

| Action | What to Do |
|--------|-----------|
| `inject_contrarian_persona` | Generate contrarian persona from presets, add to roster, re-run Phase 2 for new persona only |
| `inject_devil_advocate` | Use critic specialist to target consensus, add to roster, re-run Phase 2 |
| `escalate_challenge_depth` | Update `challenge_depth` in state (e.g., evidence → formal) |
| `repeat_round` | Increment `round` counter, re-run Phase 3 + 4 |
| `trigger_gap_fill` | Identify weakest quality dimension, route to Phase 5.5 |
| `abort_quality_failure` | Set session.status = "aborted", produce partial output |
| `skip_to_convergence` | Skip remaining phases, go directly to Phase 5 |

### Loop Prevention
Track retry counts per condition ID in the state file. When `max_retries` is reached for a condition, treat it as FALSE regardless of state values.
````

---

## Phase C Patches

### Patch C1: Evaluator-Optimizer Config (conversation-analysis.yml)

Add after `termination:` section:

```yaml
# ─── EVALUATOR-OPTIMIZER LOOP ──────────────────────────────────
evaluator_optimizer:
  enabled: true
  max_iterations: 2
  target_score: 3.6
  stall_threshold: 0.2
  stall_patience: 2
  optimization_priority:
    - perspective_diversity
    - concession_depth
    - evidence_quality
    - challenge_substantiveness
    - synthesis_quality
    - actionability
  dimension_limits:
    perspective_diversity:
      max_gap_fill_personas: 1
      preferred_persona: "critic"
    evidence_quality:
      max_gap_fill_personas: 1
      preferred_persona: "researcher"
    concession_depth:
      strategy: "re_run_aligned_pair"
      max_retries: 1
    challenge_substantiveness:
      max_gap_fill_personas: 1
      preferred_persona: "critic"
    synthesis_quality:
      strategy: "moderator_self_improve"
    actionability:
      strategy: "moderator_self_improve"
```

### Patch C2: Moderator Evaluator-Optimizer Block (conversation-moderator.md)

Add after the Phase Routing block:

````markdown
## Evaluator-Optimizer Loop (v2)

If `evaluator_optimizer.enabled: true`:

### Phase 5 — Evaluator Role

After producing the Convergence Analysis, immediately score with the Quality Rubric:

1. Score each of the 6 dimensions (1-5) with **evidence and justification**:
   ```
   #### {Dimension}: {score}/5
   **Evidence:** {cite specific outputs supporting this score}
   **Justification:** {reference the anchor descriptions from the rubric}
   ```

2. Calculate overall score (average of 6 dimensions)

3. Record in state file:
   ```yaml
   phases.phase_5_convergence.quality_scores:
     overall: {average}
     perspective_diversity: {score}
     evidence_quality: {score}
     concession_depth: {score}
     challenge_substantiveness: {score}
     synthesis_quality: {score}
     actionability: {score}
   ```

4. Append to iteration history:
   ```yaml
   phases.phase_5_convergence.history:
     - iteration: {N}
       quality_score: {overall}
       weakest_dimension: {name}
       weakest_score: {score}
       action: {what happens next}
   ```

5. **Exit decision:**
   - IF `overall >= evaluator_optimizer.target_score` → proceed to Phase 6
   - IF `iteration >= evaluator_optimizer.max_iterations` → Phase 6 with quality flag
   - IF stall detected → Phase 6 with stall flag
   - ELSE → route to Phase 5.5 (Optimizer) targeting weakest dimension

### Phase 5.5 — Optimizer Role

1. Read the weakest dimension from the evaluator's scoring
2. Look up `dimension_limits.{dimension}` in config
3. Select strategy:

   **For "moderator_self_improve" strategies (synthesis_quality, actionability):**
   - Re-run the relevant section of Phase 5 convergence with a targeted prompt
   - No sub-agent call needed — moderator refines its own output

   **For "re_run_aligned_pair" strategy (concession_depth):**
   - Identify the two most similar personas (from diversity audit similarity matrix)
   - Re-run Phase 3 cross-examination for just that pair with `challenge_depth: formal`
   - Add instruction: "You MUST find at least one substantive concession point"
   - Feed results into updated convergence input

   **For persona-based strategies (perspective_diversity, evidence_quality, challenge_substantiveness):**
   - Spawn the `preferred_persona` (or generate custom gap-fill persona)
   - Provide: topic brief, persona definition, current convergence analysis, gap identification
   - Collect output, save to `.github/state/phase-5.5/iter-{N}/`

4. Update convergence input with gap-fill results
5. Increment `phases.phase_5_convergence.iteration`
6. Route back to Phase 5 for re-evaluation
````

---

## Phase D Patches

### Patch D1: Approval Gates Config (conversation-analysis.yml)

```yaml
# ─── APPROVAL GATES ────────────────────────────────────────────
approval_gates:
  enabled: false
  gates:
    after_opening_positions:
      phase: phase_2_opening
      prompt: "Review opening positions before proceeding to cross-examination"
      options: [proceed, adjust_personas, inject_context, rerun, abort]
    after_rebuttal:
      phase: phase_4_rebuttal
      prompt: "Review challenges and rebuttals before convergence synthesis"
      options: [proceed, add_challenge, skip_to_output, another_round, abort]
    after_convergence:
      phase: phase_5_convergence
      prompt: "Review convergence analysis before final output assembly"
      options: [finalize, gap_fill, rerun_convergence, abort]
```

### Patch D2: Dynamic Selection Config (conversation-analysis.yml)

```yaml
# ─── DYNAMIC SPEAKER SELECTION ─────────────────────────────────
dynamic_selection:
  enabled: false
  applies_to: [phase_3_cross_examination, phase_4_rebuttal]
  strategy: "most_relevant"
  constraints:
    all_must_speak: true
    max_consecutive: 1
    selection_history: true
```

### Patch D3: Delegation Config (conversation-analysis.yml)

```yaml
# ─── MEDIATED DELEGATION ───────────────────────────────────────
delegation:
  enabled: false
  max_delegations_per_phase: 3
  max_delegations_per_persona: 1
  max_delegation_sub_agent_calls: 4
  approval: "auto"
  validation:
    must_state_reason: true
    must_state_priority: true
    reject_if_answerable_from_context: true
    reject_cross_phase_context: true
```

### Patch D4: Delegation Request Template (all persona agents)

Add to the "Phase Behaviors" section of each persona agent (`conversation-persona.md`, `conversation-persona-architect.md`, etc.):

````markdown
### Delegation Requests (optional, all phases)

If `delegation.enabled: true` in the workflow config, you may include delegation requests in your output when you need information from another perspective:

```markdown
### Delegation Requests

**Request:**
- **To:** {persona name or expertise description}
- **Question:** {specific, answerable question}
- **Why needed:** {how this strengthens your position or analysis}
- **Priority:** high | medium | low
- **Blocking:** yes | no
```

Rules:
- Maximum 1 request per phase (moderator enforces budget)
- Question must be specific and answerable in a short response
- Do NOT request information that will naturally emerge in the next phase
- Delegation responses you receive are mediated — the responder saw only your question, not your full position
````

### Patch D5: Moderator Gate/Selection/Delegation Block (conversation-moderator.md)

Add after the Evaluator-Optimizer block:

````markdown
## Approval Gates (v2)

If `approval_gates.enabled: true`:

After each phase listed in `approval_gates.gates`:
1. Generate a brief summary of the completed phase (position summary table, concession count, or quality score)
2. Present the gate prompt with options to the user
3. **STOP your response. Do not continue until the user responds.**
4. Process the user's choice and execute the corresponding action
5. Log the gate interaction in the state file under `gates`

## Dynamic Speaker Selection (v2)

If `dynamic_selection.enabled: true` and current phase is in `dynamic_selection.applies_to`:

Instead of fixed rotation, after each sub-agent completes within a phase:
1. Read the last sub-agent's output (brief summary)
2. List remaining personas who haven't spoken in this phase
3. Evaluate: "Whose expertise is most relevant to what was just said?"
4. Select that persona as next speaker
5. Ensure `all_must_speak` constraint is met by the end of the phase
6. Log selection reasoning in state file

## Mediated Delegation (v2)

If `delegation.enabled: true`:

After each sub-agent completes:
1. Scan their output for a "### Delegation Requests" section
2. For each request found:
   a. Validate: reason stated? priority stated? within budget? genuinely needed?
   b. If valid: spawn target persona with the specific question + relevant context
   c. Save response to `.github/state/delegations/del-{NNN}-response.md`
   d. Inject response into requesting persona's next phase prompt
3. Delegation sub-agent calls count toward `termination.token_budget`
4. Log all delegation activity in state file under `delegations`
````

---

## Complete v2 Config Template

This is the full YAML config with all v2 sections included, ready to use as a starting point:

```yaml
# ═══════════════════════════════════════════════════════════════
# Conversation Analysis Workflow v2 — Configuration
# ═══════════════════════════════════════════════════════════════
# Runtime: Yes
# Description: Single source of truth — personas, phases, routing,
#   quality loops, approval gates, and all v2 features
# ═══════════════════════════════════════════════════════════════

# ─── CONVERSATION PARAMETERS ──────────────────────────────────
conversation:
  topic: ""
  objective: ""
  constraints: []
  convergence_criteria: "weighted"

# ─── STATE MANAGEMENT (v2) ────────────────────────────────────
state:
  enabled: true
  path: ".github/state"
  auto_cleanup: false
  resume_on_restart: true

# ─── CONTEXT ISOLATION RULES (v2) ─────────────────────────────
context_rules:
  phase_2_opening:
    receives: [topic_brief, persona_definition]
    blocked: [other_positions, all_challenges, all_rebuttals]
  phase_2_5_diversity_audit:
    actor: moderator
    receives: [all_positions]
  phase_2_7_tension_extraction:
    actor: moderator
    receives: [all_positions, diversity_audit]
  phase_3_cross_examination:
    receives: [topic_brief, persona_definition, other_positions]
    blocked: [own_position, all_challenges, all_rebuttals]
  phase_4_rebuttal:
    receives: [persona_definition, own_position, challenges_received]
    blocked: [other_positions, other_challenges, all_rebuttals]
  phase_5_convergence:
    actor: moderator
    receives: [all]
  phase_5_5_gap_fill:
    receives: [topic_brief, persona_definition, convergence_analysis, gap_identification]
    blocked: [all_positions, all_challenges, all_rebuttals]

# ─── PHASE ROUTING (v2) ───────────────────────────────────────
phase_routing:
  phase_2_opening:
    enabled: true
    parallel: true
    next: phase_2_5_diversity_audit
  phase_2_5_diversity_audit:
    enabled: true
    similarity_threshold: 0.70
    next:
      default: phase_2_7_tension_extraction
      conditions:
        - id: "insufficient_diversity"
          when: "effective_persona_count < min_personas"
          action: "inject_contrarian_persona"
          then: phase_2_opening
          max_retries: 1
        - id: "premature_consensus"
          when: "all_positions_agree == true"
          action: "inject_devil_advocate"
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
          then: phase_3_cross_examination
        - id: "critical_challenges"
          when: "max_severity_challenge == 'critical' AND round < max_rounds"
          action: "repeat_round"
          then: phase_3_cross_examination
  phase_5_convergence:
    enabled: true
    next:
      default: phase_6_output
      conditions:
        - id: "quality_abort"
          when: "quality_score < re_run_threshold"
          action: "abort_quality_failure"
          then: null
        - id: "quality_improve"
          when: "quality_score < high_confidence_threshold AND iteration < max_iterations"
          action: "trigger_gap_fill"
          then: phase_5_5_gap_fill
  phase_5_5_gap_fill:
    enabled: true
    trigger: "conditional"
    next: phase_5_convergence
  phase_6_output:
    enabled: true
    next: null

# ─── MODERATION RULES ─────────────────────────────────────────
moderation:
  max_rounds: 2
  convergence_threshold: 0.70
  deadlock_protocol: "record"
  min_personas: 3
  max_personas: 5

# ─── TERMINATION CONDITIONS (v2) ──────────────────────────────
termination:
  early_convergence:
    enabled: true
    check_after: [phase_2_5_diversity_audit]
    condition: "all_positions_agree == true AND effective_persona_count < min_personas"
    action: "skip_to_convergence"
  quality_abort:
    enabled: true
    check_after: [phase_5_convergence]
    conditions:
      - when: "quality_score < re_run_threshold"
        action: "abort"
  token_budget:
    max_sub_agent_calls: 25
    warning_at: 20
    on_exceed: "graceful_degradation"
  stall_detection:
    enabled: true
    trigger: "quality_improvement < 0.2 across 2 consecutive iterations"
    action: "break_loop"

# ─── EVALUATOR-OPTIMIZER LOOP (v2) ────────────────────────────
evaluator_optimizer:
  enabled: true
  max_iterations: 2
  target_score: 3.6
  stall_threshold: 0.2
  stall_patience: 2
  optimization_priority:
    - perspective_diversity
    - concession_depth
    - evidence_quality
    - challenge_substantiveness
    - synthesis_quality
    - actionability

# ─── APPROVAL GATES (v2) ──────────────────────────────────────
approval_gates:
  enabled: false
  gates:
    after_opening_positions:
      phase: phase_2_opening
    after_rebuttal:
      phase: phase_4_rebuttal
    after_convergence:
      phase: phase_5_convergence

# ─── DYNAMIC SPEAKER SELECTION (v2) ───────────────────────────
dynamic_selection:
  enabled: false
  applies_to: [phase_3_cross_examination, phase_4_rebuttal]
  strategy: "most_relevant"
  constraints:
    all_must_speak: true
    max_consecutive: 1

# ─── MEDIATED DELEGATION (v2) ─────────────────────────────────
delegation:
  enabled: false
  max_delegations_per_phase: 3
  max_delegations_per_persona: 1
  max_delegation_sub_agent_calls: 4
  approval: "auto"

# ─── PERSONA LIBRARY ──────────────────────────────────────────
personas:
  architect:
    name: "The Architect"
    agent: conversation-persona-architect
    perspective: "Systems thinking — composability, scalability, separation of concerns"
    priorities: [Composability, Separation of concerns, Extensibility, Type safety, Maintainability]
    bias_disclosure: "Over-engineers when simplicity suffices"
    model: "Claude Sonnet 4 (copilot)"
    tools: [readFile, codebase, textSearch, fileSearch]

  pragmatist:
    name: "The Pragmatist"
    agent: conversation-persona-pragmatist
    perspective: "Shipping speed — minimal complexity, proven patterns, incremental delivery"
    priorities: [Time-to-value, Simplicity, Proven patterns, Incremental delivery, Cost]
    bias_disclosure: "Under-engineers when structure would prevent pain"
    model: "Claude Sonnet 4 (copilot)"
    tools: [readFile, codebase, textSearch]

  critic:
    name: "The Critic"
    agent: conversation-persona-critic
    perspective: "Devil's advocate — risk analysis, hidden assumptions, stress-testing"
    priorities: [Assumption exposure, Failure modes, Evidence scrutiny, Second-order effects, Reversibility]
    bias_disclosure: "Tends toward pessimism, may overweight unlikely failures"
    model: "Claude Sonnet 4 (copilot)"
    tools: [readFile, codebase, textSearch]

  researcher:
    name: "The Researcher"
    agent: conversation-persona-researcher
    perspective: "Evidence-first — benchmarks, empirical data, comparative analysis"
    priorities: [Evidence quality, Comparative analysis, Prior art, Quantitative grounding, Rigor]
    bias_disclosure: "Delays decisions waiting for more data"
    model: "Claude Sonnet 4 (copilot)"
    tools: [readFile, codebase, textSearch, fileSearch, fetch]

  custom:
    name: ""
    agent: conversation-persona
    perspective: ""
    priorities: []
    bias_disclosure: ""
    model: null
    tools: [readFile, codebase, textSearch]

# ─── VARIANT PRESETS ───────────────────────────────────────────
variants:
  quick:
    description: "2 personas, skip rebuttal"
    personas: [architect, pragmatist]
    token_cost: "~5-9 calls"
  deep:
    description: "4 personas, 2 rounds, evaluator-optimizer"
    personas: [architect, pragmatist, critic, researcher]
    max_rounds: 2
    token_cost: "~15-28 calls"
  devils_advocate:
    description: "2 personas asymmetric — propose + attack"
    personas: [pragmatist, critic]
    asymmetric: true
    token_cost: "~6-10 calls"
  panel_review:
    description: "N personas review independently"
    personas: [architect, pragmatist, critic, researcher]
    interaction: "parallel-independent"
    token_cost: "~5-9 calls"
  document_as_position:
    description: "Analyze pre-existing documents"
    personas: "auto-derive"
    token_cost: "~2-6 calls"

# ─── QUALITY RUBRIC ───────────────────────────────────────────
quality_rubric:
  dimensions:
    perspective_diversity: { weight: 1 }
    evidence_quality: { weight: 1 }
    concession_depth: { weight: 1 }
    challenge_substantiveness: { weight: 1 }
    synthesis_quality: { weight: 1 }
    actionability: { weight: 1 }
  thresholds:
    re_run: 2.0
    usable: 3.5
    high_confidence: 3.6

# ─── SELECTION GUIDE ──────────────────────────────────────────
selection_guide:
  technology_choice: [pragmatist, architect, researcher]
  architecture_design: [architect, critic, pragmatist]
  build_vs_buy: [pragmatist, custom]   # custom: economist
  go_no_go: [pragmatist, critic]
  risk_assessment: [critic, researcher, pragmatist]
```

---

## State File Template

Create this at `.github/state/conversation-state.yml` on session start:

```yaml
# Auto-generated by conversation-moderator. Do not edit during active session.
session:
  id: ""
  topic: ""
  variant: ""
  config_path: ".github/config/conversation-analysis.yml"
  started_at: ""
  updated_at: ""
  status: "active"

personas:
  configured: []
  effective: []
  removed: []

phases:
  phase_2_opening:      { status: "not_started", outputs: [] }
  phase_2_5_diversity:  { status: "not_started" }
  phase_2_7_tensions:   { status: "not_started" }
  phase_3_cross_exam:   { status: "not_started", round: 1, outputs: [] }
  phase_4_rebuttal:     { status: "not_started", round: 1, outputs: [] }
  phase_5_convergence:  { status: "not_started", iteration: 1, history: [] }
  phase_5_5_gap_fill:   { status: "not_started", iterations: [] }
  phase_6_output:       { status: "not_started" }

metrics:
  sub_agent_calls: 0
  sub_agent_budget: 25
  phases_completed: 0
  phases_skipped: 0
  total_concessions: 0
  quality_score: null

routing_log: []
gates: []
delegations: []

termination:
  reason: null
  triggered_at: null
  details: ""
```

---

## Moderator Instruction Blocks

### Summary: What to Add to conversation-moderator.md

The following blocks should be added to the moderator agent file, in this order:

1. **State Management** (Patch A3) — Read/write state file at phase boundaries
2. **Context Rule Enforcement** (Patch A4) — Construct prompts from declarative rules
3. **Phase Routing** (Patch B3) — Evaluate conditions and route after each phase
4. **Evaluator-Optimizer** (Patch C2) — Quality scoring and gap-fill loop in Phase 5/5.5
5. **Approval Gates / Dynamic Selection / Delegation** (Patch D5) — Opt-in features

Each block is gated by a config check (`if X.enabled: true`), so adding all blocks to the moderator is safe even when features are disabled — the moderator simply skips the inactive blocks.

### Moderator Decision Priority

When multiple v2 features apply at the same point, the moderator follows this priority:

```
1. Termination check     (G5 — highest priority, can stop everything)
2. Approval gate check   (G6 — user gets to decide before routing)
3. Routing evaluation    (G3 — conditions determine next phase)
4. Delegation processing (G8 — fulfilled before next phase starts)
5. Speaker selection     (G7 — determines sub-agent order within a phase)
6. Evaluator-optimizer   (G4 — operates within Phase 5/5.5 specifically)
```

---

*This blueprint provides all the implementation patches needed to upgrade from v1 to v2. Apply in phase order (A → B → C → D) for the smoothest migration.*
