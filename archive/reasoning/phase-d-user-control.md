# Phase D: User Control — Approval Gates, Dynamic Selection & Delegation

> **Parent:** [conversation-workflow-v2-improvements.md](../2026-02-24-conversation-workflow-v2-improvements.md)
> **Implements:** G6 (Human-in-the-Loop Gates) + G7 (Dynamic Speaker Selection) + G8 (Agent-to-Agent Delegation)
> **Depends on:** Phase A (G1 State File), Phase B (G3 Conditional Routing) — recommended but not strictly required
> **Priority:** Lower — all features are opt-in with defaults matching current behavior

---

## Table of Contents

1. [Overview](#overview)
2. [G6: Human-in-the-Loop Approval Gates](#g6-human-in-the-loop-approval-gates)
3. [G7: Dynamic Speaker Selection](#g7-dynamic-speaker-selection)
4. [G8: Mediated Agent-to-Agent Delegation](#g8-mediated-agent-to-agent-delegation)
5. [Feature Interactions](#feature-interactions)
6. [Implementation Sequence](#implementation-sequence)
7. [Validation Criteria](#validation-criteria)

---

## Overview

Phase D adds three **opt-in** capabilities that enhance user control and agent autonomy. All three default to disabled, preserving the current fully-autonomous workflow behavior.

| Feature | Default | Enables |
|---------|---------|---------|
| **G6: Approval Gates** | `disabled` | User checkpoints between phases — approve, redirect, or abort |
| **G7: Dynamic Selection** | `disabled` | Moderator selects next speaker based on conversation context |
| **G8: Mediated Delegation** | `disabled` | Personas can request input from other personas via moderator |

These features are independent — any combination can be enabled. They're designed for users who want more control or more natural conversation dynamics without requiring platform evolution.

### Design Philosophy

All three features follow the same pattern: **the moderator acts as a proxy**.

- G6: Moderator pauses and asks the user (proxy for a human-in-the-loop infra component)
- G7: Moderator evaluates conversation context and selects the next speaker (proxy for a dedicated selector agent)
- G8: Moderator detects delegation requests and routes them (proxy for an inter-agent communication bus)

This is deliberate. VS Code Copilot custom agents don't support native HITL gates, selector agents, or inter-agent messaging. The moderator simulates all three using its existing tools (context reasoning, sub-agent spawning, file read/write).

---

## G6: Human-in-the-Loop Approval Gates

### Rationale

The current workflow runs end-to-end without user intervention. This is efficient for well-configured conversations but leaves no structured way to:

- Review opening positions before investing in cross-examination
- Redirect the conversation if positions are off-track
- Inject additional context discovered during the conversation
- Abort early based on human judgment

### Gate Placement

Gates are placed at natural phase boundaries where user input is most valuable:

```
Phase 2 ──→ ◆ GATE 1 ──→ Phase 2.5 ──→ Phase 3 ──→ Phase 4 ──→ ◆ GATE 2 ──→ Phase 5 ──→ ◆ GATE 3 ──→ Phase 6
             │                                                      │                        │
             ▼                                                      ▼                        ▼
        Review positions                                    Review challenges           Review convergence
        before debate                                       before synthesis            before output
```

### Gate Configuration

```yaml
# ─── APPROVAL GATES ────────────────────────────────────────────
approval_gates:
  enabled: false                    # Default: fully autonomous

  # Gate definitions
  gates:

    after_opening_positions:
      phase: phase_2_opening
      prompt: |
        ## Opening Positions Review

        The following personas have generated their opening positions:
        {summary_of_positions}

        **Options:**
        1. **Proceed** — Continue to diversity audit and cross-examination
        2. **Adjust personas** — Add, remove, or reconfigure a persona before continuing
        3. **Inject context** — Provide additional information for all personas
        4. **Rerun** — Regenerate positions (with optional guidance)
        5. **Abort** — Stop the conversation here
      options:
        proceed:
          label: "Proceed to cross-examination"
          action: continue
        adjust_personas:
          label: "Adjust persona configuration"
          action: reconfigure
          requires_input: true        # User must specify what to change
        inject_context:
          label: "Add context to the Topic Brief"
          action: update_topic_brief
          requires_input: true
        rerun:
          label: "Regenerate opening positions"
          action: rerun_phase
          allows_guidance: true       # User can add guidance for the re-run
        abort:
          label: "Stop conversation"
          action: terminate
          termination_reason: "user_abort_after_positions"

    after_rebuttal:
      phase: phase_4_rebuttal
      prompt: |
        ## Challenge & Rebuttal Review

        Cross-examination and rebuttal are complete.
        {summary_of_concessions_and_remaining_disagreements}

        **Options:**
        1. **Proceed** — Continue to moderator convergence synthesis
        2. **Add challenge** — Inject a specific challenge for the moderator to consider
        3. **Skip to output** — Produce output from current data (skip convergence)
        4. **Another round** — Run another cross-exam + rebuttal cycle
        5. **Abort**
      options:
        proceed:
          label: "Proceed to convergence"
          action: continue
        add_challenge:
          label: "Inject a specific challenge"
          action: inject_challenge
          requires_input: true
        skip_rebuttal:
          label: "Skip to output assembly"
          action: skip_to_phase
          target: phase_6_output
        another_round:
          label: "Run another round of debate"
          action: route_to_phase
          target: phase_3_cross_examination
        abort:
          label: "Stop conversation"
          action: terminate

    after_convergence:
      phase: phase_5_convergence
      prompt: |
        ## Convergence Review

        The moderator has produced the convergence analysis.
        Quality score: {quality_score}/5
        {quality_dimension_summary}

        **Options:**
        1. **Finalize** — Assemble final output document
        2. **Gap-fill** — Run gap-fill to address weak dimensions
        3. **Rerun convergence** — Re-synthesize with different emphasis
        4. **Abort**
      options:
        finalize:
          label: "Assemble final output"
          action: continue
        gap_fill:
          label: "Run gap-fill optimization"
          action: route_to_phase
          target: phase_5_5_gap_fill
        rerun_convergence:
          label: "Re-run convergence synthesis"
          action: rerun_phase
          allows_guidance: true
        abort:
          label: "Stop conversation"
          action: terminate
```

### Gate Execution Protocol

When the moderator reaches a gate:

```
FUNCTION execute_gate(gate_config, state):
  1. Generate gate summary:
     → Read completed phase outputs from state
     → Produce the {summary} sections referenced in the prompt template
     → Format the prompt with options

  2. Present to user:
     → Output the gate prompt as a message to the user
     → STOP and WAIT for user response

  3. Process user response:
     → Parse the user's choice (may be option number, label, or free text)
     → Map to an action from gate_config.options

  4. Execute action:
     - continue: proceed to next phase per routing
     - reconfigure: read user's configuration changes, update state
     - update_topic_brief: append user's new context to topic brief
     - rerun_phase: re-run current phase (with optional user guidance)
     - inject_challenge: add challenge to Phase 5 inputs
     - skip_to_phase: route directly to specified phase
     - route_to_phase: route to specified phase
     - terminate: set session.status = "aborted", produce partial output

  5. Log gate interaction in state:
     → Record: gate_id, user choice, action executed, timestamp
```

### Gate Summaries

The quality of the gate experience depends on useful summaries. The moderator should produce brief, decision-relevant summaries:

**Position Summary Template (Gate 1):**
```markdown
| Persona | Core Recommendation | Key Argument | Confidence |
|---------|-------------------|-------------|------------|
| {name} | {1-sentence recommendation} | {strongest argument} | {high/med/low} |
```

**Concession Summary Template (Gate 2):**
```markdown
**Concessions:** {N} total across {M} personas
**Key concessions:** {top 2-3 most significant position changes}
**Remaining disagreements:** {core unresolved tensions}
```

**Quality Summary Template (Gate 3):**
```markdown
**Overall:** {score}/5 ({label: high confidence / usable / below threshold})
**Strongest:** {best dimension} ({score})
**Weakest:** {worst dimension} ({score}) — {1-sentence explanation}
```

### State Tracking for Gates

```yaml
# Addition to state file
gates:
  - gate_id: "after_opening_positions"
    triggered_at: "2026-02-24T14:50:00Z"
    user_choice: "proceed"
    action_executed: "continue"
    user_input: null

  - gate_id: "after_rebuttal"
    triggered_at: "2026-02-24T15:15:00Z"
    user_choice: "add_challenge"
    action_executed: "inject_challenge"
    user_input: "Consider the maintenance cost of microservices for a 2-person team"
```

---

## G7: Dynamic Speaker Selection

### Rationale

In a real panel discussion, the most relevant person speaks next — not whoever's next in the rotation. If The Critic raises a challenge about technical risk, The Architect (who can address scalability concerns) should respond before The Pragmatist (who might just say "let's not over-engineer").

The current workflow uses fixed rotation: all personas execute in the same order every phase. This misses opportunities for natural dialectic flow.

### Selection Strategy

The moderator acts as a **selector** — after each sub-agent completes within a phase, it evaluates who should go next.

```yaml
# ─── DYNAMIC SPEAKER SELECTION ─────────────────────────────────
dynamic_selection:
  enabled: false                    # Default: fixed rotation

  # Which phases use dynamic selection
  applies_to:
    - phase_3_cross_examination
    - phase_4_rebuttal

  # Selection strategy
  strategy: "most_relevant"         # most_relevant | least_heard | priority_weighted

  # Selection reasoning template (moderator uses this to decide)
  selection_prompt: |
    Given the conversation so far in this phase:

    **Last speaker:** {last_persona_name}
    **Last output summary:** {1-2 sentence summary of what they said}
    **Remaining speakers:** {list of personas who haven't spoken in this phase}

    Select the next speaker based on:
    1. **Relevance:** Whose expertise is most directly engaged by the last output?
    2. **Balance:** Who has been least represented so far?
    3. **Tension:** Whose perspective would create the most productive challenge?

    **Selection:** {chosen persona} because {1-sentence reason}

  # Constraints
  constraints:
    all_must_speak: true            # Every persona must speak once per phase
    max_consecutive: 1              # A persona can't speak twice in a row
    selection_history: true         # Log selection reasoning in state
```

### Selection Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| **most_relevant** | Select the persona whose expertise is most engaged by the current conversation state | Focused, deep debates |
| **least_heard** | Select the persona who has spoken least recently | Ensuring balanced representation |
| **priority_weighted** | Select based on persona priority weights (configurable per topic) | Weighted panel reviews |

### Selection Algorithm

```
FUNCTION select_next_speaker(phase, state, config):
  1. spoken = personas who have already produced output in this phase
  2. remaining = all_personas - spoken
  3. IF len(remaining) == 0: RETURN null (phase complete)
  4. IF len(remaining) == 1: RETURN remaining[0] (no choice needed)

  5. SWITCH config.dynamic_selection.strategy:

     CASE "most_relevant":
       → Read last sub-agent output (the most recent Phase output)
       → For each remaining persona, evaluate:
         "How directly does {persona.perspective} engage with {last_output_topics}?"
       → Select highest relevance

     CASE "least_heard":
       → Count total words/outputs per persona across all phases
       → Select persona with lowest count

     CASE "priority_weighted":
       → Read persona priority weights from config
       → Select highest-weight remaining persona

  6. Log selection in state:
     selection_log[phase].append({
       selected: persona_id,
       reason: "...",
       remaining_options: [...],
       timestamp: "..."
     })

  7. RETURN selected persona
```

### Phase 3 Dynamic Selection Example

Without dynamic selection (fixed rotation):
```
Architect challenges → Pragmatist challenges → Critic challenges
```

With dynamic selection:
```
1. Architect challenges Pragmatist and Critic positions
   → Architect raises concern about "tight coupling in the proposed microservices"
2. Moderator selects: Pragmatist (directly relevant — they proposed microservices)
   → Pragmatist challenges Architect and Critic positions,
     including a defense of their microservices approach
3. Moderator selects: Critic (tension — hasn't addressed risk yet)
   → Critic challenges Architect and Pragmatist positions
```

The conversation flows more naturally because each speaker's output informs who responds next.

### Constraints and Guardrails

| Constraint | Purpose |
|-----------|---------|
| `all_must_speak: true` | Prevents selection bias from excluding quiet personas |
| `max_consecutive: 1` | Prevents two-persona ping-pong that excludes others |
| `selection_history: true` | Audit trail of selection reasoning |

---

## G8: Mediated Agent-to-Agent Delegation

### Rationale

In the current workflow, personas operate in isolation within each phase. If The Architect realizes they need empirical data to support their position, they can't ask The Researcher — they must proceed without it. The missing data becomes a weakness in their position that only gets exposed (and never filled) during cross-examination.

Delegation allows personas to express needs the moderator can fulfill.

### Delegation Request Format

Personas include structured delegation requests in their output (any phase):

```markdown
### Delegation Requests

> These requests are for the moderator. They will be routed to the
> appropriate persona and the response will be provided before the
> next phase begins.

**Request 1:**
- **To:** The Researcher (or: "anyone with empirical data expertise")
- **Question:** "What are the measured latency percentiles for gRPC vs REST in production systems with >10k QPS?"
- **Why needed:** "My architectural recommendation depends on gRPC having <5ms p99 latency advantage. Without data, this is an assumption."
- **Priority:** high | medium | low
- **Blocking:** yes | no (does this block my position or just improve it?)

**Request 2:**
- **To:** The Pragmatist
- **Question:** "What's the realistic migration timeline from REST to gRPC for a 50k LOC codebase?"
- **Why needed:** "Need to assess if the architectural benefits justify the migration cost."
- **Priority:** medium
- **Blocking:** no
```

### Delegation Protocol

```
┌────────────────────────────────────────────────────────────┐
│                   DELEGATION PROTOCOL                       │
│                                                              │
│  1. Persona completes phase output                           │
│     └── Output includes "Delegation Requests" section        │
│                                                              │
│  2. Moderator detects delegation requests                    │
│     └── Parse: target persona, question, priority, blocking  │
│                                                              │
│  3. Moderator evaluates requests                             │
│     ├── Is target persona available?                         │
│     ├── Is request within scope of the conversation?         │
│     ├── Would fulfilling this exceed token budget?           │
│     └── Is this a legitimate information gap or just stalling?│
│                                                              │
│  4. Moderator routes approved requests                       │
│     └── Spawn target persona with:                           │
│         - Original persona definition                        │
│         - The specific question                              │
│         - Relevant context (topic brief + own position)      │
│         - Instruction: "Answer this specific question only"  │
│                                                              │
│  5. Moderator collects responses                             │
│     └── Save to .github/state/delegations/                   │
│                                                              │
│  6. Moderator injects responses                              │
│     └── Requesting persona receives delegation response      │
│         as additional context in their NEXT phase prompt     │
│                                                              │
└────────────────────────────────────────────────────────────┘
```

### Configuration

```yaml
# ─── MEDIATED DELEGATION ───────────────────────────────────────
delegation:
  enabled: false                    # Default: no delegation

  # Budget limits
  max_delegations_per_phase: 3      # Cap total delegations per phase
  max_delegations_per_persona: 1    # Cap delegations per persona per phase
  max_delegation_sub_agent_calls: 4 # Total sub-agent budget for all delegations

  # Approval mode
  approval: "auto"                  # auto | moderator_review
  # auto: fulfill all valid requests within budget
  # moderator_review: moderator evaluates each request before fulfilling

  # Request validation rules
  validation:
    must_state_reason: true         # "Why needed" field must be present
    must_state_priority: true
    reject_if_answerable_from_context: true  # Reject if the question can be answered from existing phase outputs
    reject_cross_phase_context: true        # Reject requests for information that will naturally appear in the next phase

  # Response injection
  response_injection:
    target_phase: "next"            # Inject into the requesting persona's next phase prompt
    format: |
      ### Delegation Response (from {responder_name})
      **Your question:** {original_question}
      **Response:** {response_content}
      **Note:** This was a mediated response - {responder_name} answered your
      specific question without seeing your full position.
```

### Delegation Decision Matrix

The moderator evaluates each delegation request:

| Check | Approve | Reject | Reason |
|-------|---------|--------|--------|
| Target persona exists | ✔ | | |
| Target persona unavailable | | ✔ | "Target persona not in this conversation" |
| Question in scope | ✔ | | |
| Question out of scope | | ✔ | "Question falls outside conversation scope" |
| Within delegation budget | ✔ | | |
| Budget exceeded | | ✔ | "Delegation budget exhausted" |
| Genuine information gap | ✔ | | |
| Answerable from existing data | | ✔ | "This information is available in {persona}'s Phase 2 output" |
| Will be naturally addressed next phase | | ✔ | "Cross-examination will naturally surface this" |

### Preventing Delegation Loops

A delegation request from Persona A to Persona B could trigger a delegation request from B back to A, creating an infinite loop.

**Prevention rules:**
1. Delegation responses MUST NOT contain delegation requests (format enforcement)
2. A persona cannot delegate to someone who has a pending delegation TO them
3. Maximum delegation chain depth: 1 (no delegation-of-delegation)

### State Tracking

```yaml
# Addition to state file
delegations:
  - id: "del-001"
    requesting_persona: "architect"
    target_persona: "researcher"
    phase: "phase_2_opening"
    question: "gRPC vs REST latency percentiles at >10k QPS"
    priority: "high"
    blocking: false
    status: "fulfilled"              # pending | approved | rejected | fulfilled
    response_path: ".github/state/delegations/del-001-response.md"
    injected_in_phase: "phase_3_cross_examination"
    sub_agent_call_used: true
    timestamp: "2026-02-24T14:52:00Z"
```

---

## Feature Interactions

### G6 + G7: Gates with Dynamic Selection

When both are enabled, gates show selection history to help the user understand the conversation flow:

```markdown
## Challenge & Rebuttal Review

**Speaker order (dynamically selected):**
1. Architect — selected first (raised coupling concern)
2. Critic — selected second (most tension with Architect on risk)
3. Pragmatist — selected last (balanced perspective after debate)

**Selection rationale visible in state file.**
```

### G6 + G8: Gates with Delegation

Gates can show pending delegation requests, allowing the user to approve/reject them:

```markdown
## Opening Positions Review

**Pending delegation requests:**
- Architect → Researcher: "gRPC latency data" (priority: high)
- Pragmatist → Critic: "Risk assessment for monolith approach" (priority: medium)

**Options:**
1. Proceed (fulfill approved delegations before cross-examination)
2. Proceed (skip delegations)
3. Review delegations individually
```

### G7 + G8: Dynamic Selection informed by Delegations

When a delegation request arrives, the moderator can use it as a signal for speaker selection:

```
Architect's output includes delegation request to Researcher
→ Dynamic selection: prioritize Researcher as next speaker
   (they already need to run; delegation can be naturally addressed)
```

---

## Implementation Sequence

### Step 1: G6 — Approval Gates (lowest risk)

1. Add `approval_gates` section to `conversation-analysis.yml`
2. Add gate execution protocol to `conversation-moderator.md`
3. Add gate interaction logging to state file schema
4. Test: enable gates, verify moderator pauses and waits correctly

### Step 2: G7 — Dynamic Selection

1. Add `dynamic_selection` section to `conversation-analysis.yml`
2. Add selection logic to `conversation-moderator.md` (Phase 3 and 4)
3. Add selection history logging to state file schema
4. Test: enable dynamic selection, verify speaker order varies by conversation

### Step 3: G8 — Mediated Delegation

1. Add `delegation` section to `conversation-analysis.yml`
2. Add delegation request format to persona prompts (all persona agents)
3. Add delegation detection and routing to `conversation-moderator.md`
4. Add delegation state tracking to state file schema
5. Test: persona generates delegation request, moderator routes and injects response

### Step 4: Integration testing

- Test all three features together
- Verify budget accounting includes delegation sub-agent calls
- Verify gates show delegation and selection information
- Verify state file captures all interaction data

---

## Validation Criteria

### G6 Pass Criteria

| Criterion | Test |
|-----------|------|
| Gate pauses execution | Moderator stops and waits for user input at each gate |
| Summaries are useful | Gate summaries contain decision-relevant information |
| All options work | Each option (proceed, adjust, inject, rerun, abort) executes correctly |
| State logs gate choices | Gate interactions recorded with user choice and action |
| Disabled by default | No gates fire when `approval_gates.enabled: false` |
| Injected context propagates | Context added at Gate 1 appears in subsequent phase prompts |

### G7 Pass Criteria

| Criterion | Test |
|-----------|------|
| Speaker order varies | Two runs of the same topic produce different speaker orders |
| All personas speak | `all_must_speak` constraint ensures no persona is skipped |
| Selection reasoning logged | State file contains selection rationale per choice |
| Strategies produce different orders | `most_relevant` vs `least_heard` produce different sequences |
| Disabled by default | Fixed rotation used when `dynamic_selection.enabled: false` |

### G8 Pass Criteria

| Criterion | Test |
|-----------|------|
| Delegation requests detected | Moderator parses delegation sections from persona output |
| Valid requests fulfilled | Target persona spawned with correct question |
| Invalid requests rejected | Requests that fail validation are rejected with reason |
| Responses injected | Requesting persona receives delegation response in next phase |
| Budget respected | Delegation sub-agent calls count toward token budget |
| Loops prevented | Delegation responses never contain delegation requests |
| Disabled by default | No delegation processing when `delegation.enabled: false` |

---

*Phase D adds sophistication. Each feature is independently useful and opt-in, making adoption incremental and risk-free.*
