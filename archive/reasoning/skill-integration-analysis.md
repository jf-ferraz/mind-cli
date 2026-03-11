# Deep Analysis: Improving Conversation Workflow Technicity Through Agent Skills

**Date:** 2026-02-24
**Status:** Complete — all 18 techniques implemented (2026-02-25)
**Scope:** Cross-pollination of blackbox agent/skill patterns into the conversation analysis workflow

---

## Executive Summary

The blackbox `agents/` and `skills/` directories contain a mature set of patterns for structuring agent behavior — open verification questions, dual-path verification, evidence-based assessment, scope violation detection, skills-as-modules, intent markers, and decision documentation with reasoning chains. The conversation analysis workflow currently uses none of these. This document maps each technique to specific conversation workflow improvements and proposes a concrete integration plan.

---

## Table of Contents

1. [Source Inventory](#source-inventory)
2. [Gap Matrix](#gap-matrix)
3. [Technique Deep Dives](#technique-deep-dives)
4. [Integration Architecture](#integration-architecture)
5. [Persona Skill Injection Model](#persona-skill-injection-model)
6. [Moderator Skill Injection Model](#moderator-skill-injection-model)
7. [New Conversation Skills](#new-conversation-skills)
8. [Implementation Priority](#implementation-priority)

---

## Source Inventory

### Blackbox Core Agents (7)

| Agent | Key Technique | Transferable To |
|-------|--------------|-----------------|
| **Orchestrator** | Request type classification (NEW_PROJECT/BUG_FIX/ENHANCEMENT/REFACTOR), quality gates with retry limits (max 2), workflow state resume via `docs/current.md`, deliverable verification between handoffs, specialist scanning | **Moderator** — same orchestration needs |
| **Analyst** | Open verification questions (70% vs 17% accuracy), per-type artifacts, context-first protocol ("read before writing"), scope explicitness (in/out), incremental doc updates | **All personas** — question quality in challenges |
| **Architect** | Convention hierarchy (user > docs > codebase > best practices), rejected alternatives documentation, minimum viable structure, "no tech-specific prescriptions" | **Architect persona** — decision documentation |
| **Developer** | Spec classification (detailed vs freeform → different latitude), scope violation self-monitoring, flag-don't-fix for out-of-scope discoveries, batch operations, incremental over regenerative | **All personas** — scope discipline in phases |
| **Tester** | Test type hierarchy (integration > property-based > unit), "test behavior not implementation", per-type testing strategy, independent tests, coverage verification protocol | **Moderator** — quality verification approach |
| **Reviewer** | Evidence-based assessment (no self-scoring), MUST/SHOULD/COULD priority, dual-path verification for blocking findings, temporal contamination check, intent markers (:PERF: :UNSAFE: :SCHEMA: :TEMP:) | **Moderator** — convergence assessment |
| **Discovery** | Interactive exploration, 5-8 targeted questions in batches of 2-3, open over closed, synthesize then confirm | **Moderator** — Phase 1 Topic Brief elicitation |

### Blackbox Skills (4)

| Skill | Key Technique | Transferable To |
|-------|--------------|-----------------|
| **Debugging** | 5-step protocol (reproduce → evidence → hypothesis → verify → fix), open verification questions, clean exit invariant, 10+ observations before hypothesis, 3 hypothesis-verify cycles then escalate | **Critic persona** — challenge methodology |
| **Planning** | Three-document model (PLAN/WIP/LEARNINGS), decision log with 2+ step reasoning chains, rejected alternatives, invisible knowledge capture, known-good increments | **Moderator** — phase tracking; **All personas** — reasoning documentation |
| **Quality Review** | 6 cognitive questions (names/types → structure → idioms → DRY → docs/tests → cross-file patterns), MUST/SHOULD/COULD severity | **Moderator** — quality rubric methodology |
| **Refactoring** | Priority classification (Critical/High/Nice/Skip), DRY = knowledge not code, semantic over structural abstraction, golden rule (baseline before changing) | **All personas** — argument prioritization |

---

## Gap Matrix

| # | Technique | Present in Blackbox Agents | Present in Conversation Workflow | Gap Severity |
|---|-----------|---------------------------|--------------------------------|-------------|
| T1 | Open verification questions | Analyst, Debugging skill | **No** — challenges use closed adversarial framing | **High** |
| T2 | Dual-path verification | Reviewer | **No** — claims verified in one direction only | **High** |
| T3 | Convention hierarchy for decisions | Architect | **No** — no structured priority for conflicting arguments | **Medium** |
| T4 | Scope violation self-monitoring | Developer | **No** — personas drift freely without self-check | **High** |
| T5 | Evidence-based assessment (no self-scoring) | Reviewer | **Partial** — rubric exists but scoring is self-reported | **Medium** |
| T6 | Context-first protocol | All agents | **Partial** — moderator reads config, personas don't read prior context systematically | **Low** |
| T7 | Quality gates with retry limits | Orchestrator | **Partial** — gap-fill exists but no structured gate/retry model | **Medium** |
| T8 | Skills-as-modules | Skills directory | **No** — no loadable skill system for conversation agents | **High** |
| T9 | Intent markers | Reviewer | **No** — no way to mark deliberate choices in positions | **Medium** |
| T10 | Decision documentation (rejected alternatives + reasoning chains) | Architect, Planning skill | **Partial** — positions have recommendations but no rejected alternatives section | **High** |
| T11 | Clean exit invariants | Debugging skill | **No** — no post-condition checklist per phase | **Medium** |
| T12 | Per-type behavior classification | Orchestrator, Analyst, Developer, Tester | **No** — conversation workflow has one mode regardless of topic type | **Medium** |
| T13 | Deliverable verification between handoffs | Orchestrator | **No** — moderator doesn't validate persona output before proceeding | **High** |
| T14 | Flag-don't-fix | Developer | **No** — personas either address issues or ignore them entirely | **Low** |
| T15 | MUST/SHOULD/COULD severity classification | Reviewer, Quality Review skill | **No** — all challenges treated as equal weight | **High** |
| T16 | Reasoning chains (2+ steps) | Planning skill | **No** — arguments are stated, not chain-reasoned | **High** |
| T17 | Hypothesis-verify cycles | Debugging skill | **No** — claims are asserted, not tested against counter-evidence | **Medium** |
| T18 | Semantic over structural abstraction | Refactoring skill | **No** — synthesis often lists positions structurally, not by meaning | **Low** |

---

## Technique Deep Dives

### T1: Open Verification Questions in Cross-Examination

**Source:** Analyst agent — "Use open verification questions when analyzing problems: 'What happens when...' (70% accuracy) over 'Is this the cause?' (17% accuracy)"

**Current problem:** Phase 3 (Cross-Examination) produces challenges that are often closed adversarial assertions: "Your recommendation of X is wrong because Y." This invites defensive rebuttal rather than genuine exploration.

**Proposed change:** Add to persona Phase 3 instructions:

```markdown
### Challenge Construction Protocol

Frame challenges as **open verification questions**, not closed assertions:

| Instead of (closed, 17% accuracy) | Use (open, 70% accuracy) |
|----------------------------------|--------------------------|
| "Your assumption about X is wrong" | "What happens to your recommendation when X doesn't hold?" |
| "You didn't consider Y" | "How does Y interact with your proposed approach?" |
| "Z contradicts your claim" | "What would your model predict about Z, and how does that compare to the observed data?" |

Each challenge MUST contain:
1. An open question (starts with What/How/Which/When)
2. A specific scenario or data point that motivates the question
3. Your assessment of severity: MUST-ADDRESS / SHOULD-ADDRESS / COULD-CONSIDER

Do NOT frame challenges as "you are wrong because..." — frame them as "what happens when..."
```

**Impact:** Higher-quality challenges that expose genuine weaknesses instead of triggering defensive posturing. The severity classification (T15) is bundled here.

### T2: Dual-Path Verification for Key Claims

**Source:** Reviewer agent — "Before declaring a MUST violation, verify through both paths: Forward ('The code does X → this leads to problem Y') and Backward ('Problem Y would require condition Z → does the code have condition Z?')."

**Current problem:** Position papers make claims that are verified in only one direction. "gRPC is faster than REST because of binary serialization" — but is binary serialization the bottleneck in the specific scenario being discussed?

**Proposed change:** Add to persona Phase 2 (position generation) and Phase 3 (challenges):

```markdown
### Claim Verification Protocol (required for "strong" evidence claims)

For every claim you rate as "strong" in the Evidence Registry, verify through dual paths:

1. **Forward path:** "Evidence E supports claim C because mechanism M"
   - E: [specific evidence]
   - M: [causal mechanism]
   - C: [your claim]

2. **Backward path:** "Claim C would require condition K to hold. Does K hold in our context?"
   - C: [your claim]
   - K: [necessary condition]
   - Holds? [yes with evidence / no / unknown]

If only the forward path confirms but the backward path is "unknown" or "no",
downgrade the claim from "strong" to "moderate" and note the gap.
```

**Impact:** Eliminates confident-sounding claims that don't survive scrutiny. Forces personas to reality-check their own arguments before the cross-examination phase.

### T3: Convention Hierarchy for Conflicting Arguments

**Source:** Architect agent — "When making design decisions, follow this priority: 1. User instruction 2. Project documentation 3. Codebase patterns 4. General best practices"

**Current problem:** When personas cite conflicting evidence or priorities, the moderator has no structured framework for weighting them during convergence. All arguments are treated as equal in principle.

**Proposed change:** Add to moderator's Phase 5 (Convergence) instructions:

```markdown
### Argument Weighting Hierarchy

When conflicting claims or recommendations arise, weight by this priority:

1. **User-stated constraints** — If the user specified a constraint, no argument can override it
2. **Empirical evidence (strong)** — Claims backed by replicated, quantitative data
3. **Empirical evidence (moderate)** — Claims backed by credible single-study data
4. **Domain-specific reasoning** — Arguments from relevant domain expertise with causal mechanism
5. **Theoretical reasoning** — Arguments from first principles without empirical backing
6. **Anecdotal evidence** — Case studies, individual experiences, "in my experience"
7. **General best practice** — Industry conventions without context-specific justification

When two claims at the same level conflict, note the conflict as an unresolved tension.
When a lower-level claim contradicts a higher-level one, the higher level prevails
(with the contradiction noted as a risk factor).
```

**Impact:** Convergence synthesis becomes principled, not arbitrary. The reasoning behind "why recommendation A over B" is traceable to a declared hierarchy.

### T4: Scope Violation Self-Monitoring

**Source:** Developer agent — "Monitor yourself. If you find yourself creating a new module the architect didn't specify → stop and flag."

**Current problem:** Personas drift outside their assigned perspective. The Architect starts making shipping-speed arguments (Pragmatist territory). The Critic starts proposing solutions (Architect territory). This reduces effective diversity.

**Proposed change:** Add to ALL persona agents:

```markdown
### Scope Discipline

You have a defined perspective and priorities. Monitor yourself for scope drift:

| If you find yourself... | It means... | Do this instead... |
|------------------------|-------------|-------------------|
| Proposing solutions to problems you raised | Drifting from Critic → Architect | State the problem. Let others propose solutions. |
| Arguing about timeline/cost from a design persona | Drifting from Architect → Pragmatist | State the design trade-off. Let Pragmatist assess cost. |
| Citing evidence without analyzing methodology | Drifting from any → Researcher territory | Flag: "Evidence needed. Delegation request to Researcher." |
| Making risk arguments from a solutions-focused persona | Drifting from Pragmatist → Critic | Flag the risk briefly, then return to your perspective. |

**Flag-don't-drift:** If a critical point falls outside your perspective, note it in a
`### Out-of-Scope Observations` section at the end of your output. Don't develop the argument.
```

**Impact:** Maintains diversity throughout the conversation. Prevents the common failure mode where all personas converge on similar reasoning styles by Phase 4.

### T10: Decision Documentation with Rejected Alternatives

**Source:** Architect agent — "Every significant design choice requires: What was decided, Why, What was rejected, Consequences. Rejected alternatives are more valuable than chosen ones."

**Current problem:** Position papers state recommendations but don't document what alternatives were considered and rejected. The convergence synthesis can't assess whether a recommendation was the best of many options or the only option considered.

**Proposed change:** Add to persona Phase 2 output format:

```markdown
### Revised Phase 2 Output Format

1. Thesis Statement (2-3 sentences)
2. Key Arguments (3-5, each with evidence/reasoning and confidence level)
3. **Alternatives Considered** (NEW — mandatory)
   For each rejected alternative:
   - What: {the alternative approach}
   - Why rejected: {specific reason, with 2+ step reasoning chain}
   - What it would have been better for: {scenario where this IS the right choice}
4. Assumptions (explicit)
5. Risks & Weaknesses (self-aware)
6. Recommendation
7. **Reasoning Chain** (NEW — mandatory)
   - Premise → Implication → Conclusion (minimum 2 reasoning steps)
   - Each step labeled as: empirical / theoretical / assumed
8. Priority Ranking (if applicable)
9. Evidence Registry (mandatory)
10. Out-of-Scope Observations (from T4)
```

**Impact:** Richer positions that the moderator can synthesize more effectively. Rejected alternatives from different personas often surface as recommended approaches of other personas — the moderator can identify these overlaps during convergence.

### T13: Deliverable Verification Between Phase Handoffs

**Source:** Orchestrator agent — "Between agents, verify the previous agent completed its deliverables before proceeding."

**Current problem:** The moderator spawns personas and processes whatever comes back, without validating that the output meets the expected format. A persona that produces a 2-sentence position paper (missing Evidence Registry, assumptions, alternatives) gets the same treatment as a thorough one.

**Proposed change:** Add to moderator's phase transition logic:

```markdown
### Output Validation Protocol (after every sub-agent call)

Before proceeding to the next phase, verify each persona output contains:

**Phase 2 (Opening Position):**
- [ ] Thesis statement present (2-3 sentences)
- [ ] At least 3 key arguments with evidence/reasoning
- [ ] Alternatives Considered section with ≥1 rejected alternative
- [ ] Reasoning chain with ≥2 steps
- [ ] Evidence Registry with ≥1 entry
- [ ] Assumptions stated
- [ ] Risks & Weaknesses stated

**Phase 3 (Cross-Examination):**
- [ ] At least 1 open verification question per position examined
- [ ] Severity classification on each challenge (MUST/SHOULD/COULD)
- [ ] At least 1 agreement identified per position (intellectual honesty check)
- [ ] Counter-evidence or alternative cited for MUST-ADDRESS challenges

**Phase 4 (Rebuttal):**
- [ ] Every MUST-ADDRESS challenge responded to (concede, rebut, or partial)
- [ ] Concessions Log present (even if empty — must state "No concessions" explicitly)
- [ ] Updated Recommendation reflects any concessions made
- [ ] Reasoning chain for any position changes

**If validation fails:** Re-prompt the persona with: "Your output is missing: {list}.
Please complete the following sections: {specific sections}." Max 1 retry per persona.
```

**Impact:** Consistent output quality across personas. Prevents garbage-in/garbage-out where a weak Phase 2 output produces meaningless Phase 3 challenges produces empty Phase 4 rebuttals.

### T15: MUST/SHOULD/COULD Severity Classification

**Source:** Reviewer + Quality Review skill — consistent severity classification across all quality assessments.

**Current problem:** All challenges in Phase 3 are treated as equal weight. A fundamental logical flaw in a recommendation gets the same treatment as a stylistic quibble about terminology. The rebutting persona wastes effort on low-severity challenges while potentially rushing through critical ones.

**Proposed change:** Already bundled into T1 (challenge construction protocol). Additionally, add to moderator's convergence:

```markdown
### Challenge Severity in Convergence

When synthesizing the conversation, weight challenges by their assigned severity:

- **MUST-ADDRESS** challenges that went unrebutted → high-impact finding, must appear
  in Executive Summary and affect recommendation confidence
- **SHOULD-ADDRESS** challenges unrebutted → medium-impact, appears in Key Insights
- **COULD-CONSIDER** challenges unrebutted → low-impact, appears in appendix only

This prevents the convergence from treating 12 minor quibbles as
equivalent to 1 fundamental architectural flaw.
```

### T16: Reasoning Chains (2+ Steps)

**Source:** Planning skill — "Each rationale must contain at least 2 reasoning steps. Single-step rationales are insufficient."

**Current problem:** Arguments in position papers are often stated as conclusions without showing the reasoning path: "We should use PostgreSQL because it supports JSONB." This is a single-step rationale that can't be effectively challenged.

**Proposed change:** Require explicit reasoning chains:

```markdown
### Reasoning Chain Format (required for all Key Arguments)

Each argument must show a reasoning chain with ≥2 steps:

**Insufficient (1 step):**
> "PostgreSQL because it supports JSONB"

**Sufficient (2+ steps):**
> "Our data model requires flexible schema for user-defined fields
> → JSONB columns provide schema flexibility within a relational database
> → PostgreSQL's JSONB has GIN indexing for query performance on flexible fields
> → Therefore PostgreSQL satisfies both the flexibility requirement and the
>   performance requirement without introducing a second database technology"

Label each step: [empirical] [theoretical] [assumed]

Reasoning chains enable precise challenges — a challenger can attack step 2
without disputing step 1, leading to more targeted cross-examination.
```

---

## Integration Architecture

### Where Techniques Map to Workflow Phases

```
                    ┌─────────────────────────────────┐
                    │       CONVERSATION WORKFLOW       │
                    └─────────────────────────────────┘

Phase 1: Topic Brief
  └─ T6:  Context-first (moderator reads docs first)
  └─ T12: Topic type classification (arch/eval/strategy/risk)

Phase 2: Opening Positions
  └─ T2:  Dual-path verification on strong claims
  └─ T4:  Scope violation self-monitoring
  └─ T10: Rejected alternatives + reasoning chains
  └─ T16: 2+ step reasoning chains (mandatory)
  └─ T9:  Intent markers on deliberate trade-offs

Phase 2.5: Diversity Audit
  └─ T11: Clean exit invariant (audit checklist)
  └─ T13: Deliverable verification (output format check)

Phase 3: Cross-Examination
  └─ T1:  Open verification questions (70% accuracy)
  └─ T15: MUST/SHOULD/COULD severity classification
  └─ T4:  Scope violation self-monitoring
  └─ T17: Hypothesis-verify structure for challenges

Phase 4: Rebuttal
  └─ T2:  Dual-path on rebuttal arguments
  └─ T14: Flag-don't-fix for out-of-scope discoveries
  └─ T4:  Scope discipline

Phase 5: Convergence
  └─ T3:  Convention hierarchy for argument weighting
  └─ T5:  Evidence-based assessment (concrete findings)
  └─ T15: Severity-weighted synthesis
  └─ T18: Semantic grouping over structural listing

Phase 5.5: Gap-Fill
  └─ T7:  Quality gate with retry limit
  └─ T17: Targeted hypothesis-verify for gaps

Phase 6: Output
  └─ T11: Clean exit invariant (output completeness checklist)
  └─ T13: Final deliverable verification
```

---

## Persona Skill Injection Model

Inspired by the blackbox `skills/` architecture — on-demand methodology modules that augment agent prompts without replacing them — apply the same pattern to conversation personas.

### What is a Conversation Skill?

A conversation skill is a **loadable methodology block** that the moderator injects into a persona's prompt when the conversation context calls for it. Skills don't change who the persona IS — they change HOW the persona operates for a specific phase.

```
skills/conversation/
├── SKILLS.md                        # Index — when to load each skill
├── challenge-methodology/SKILL.md   # Open questions, severity classification
├── evidence-standards/SKILL.md      # Dual-path verification, evidence registry rules
├── reasoning-chains/SKILL.md        # 2+ step chains, labeled steps
├── decision-documentation/SKILL.md  # Rejected alternatives, consequences
└── scope-discipline/SKILL.md        # Self-monitoring, flag-don't-drift
```

### Skill Index (SKILLS.md)

```markdown
# Conversation Skills

| Skill | When to Load |
|-------|-------------|
| `challenge-methodology/SKILL.md` | Phase 3 — injected into all persona prompts for cross-examination |
| `evidence-standards/SKILL.md` | Phase 2 — injected when topic requires empirical rigor |
| `reasoning-chains/SKILL.md` | Phase 2 + Phase 4 — injected for all position/rebuttal generation |
| `decision-documentation/SKILL.md` | Phase 2 — injected for architecture/technology decisions |
| `scope-discipline/SKILL.md` | All phases — always injected to maintain persona diversity |
```

### Integration with Moderator

The moderator decides which skills to inject based on topic type and phase:

```yaml
# Addition to conversation-analysis.yml
skill_injection:
  enabled: true
  skills_path: "skills/conversation"

  # Auto-injection rules
  rules:
    - skill: "scope-discipline"
      inject_in: [phase_2, phase_3, phase_4]
      condition: "always"

    - skill: "challenge-methodology"
      inject_in: [phase_3]
      condition: "always"

    - skill: "reasoning-chains"
      inject_in: [phase_2, phase_4]
      condition: "always"

    - skill: "evidence-standards"
      inject_in: [phase_2]
      condition: "topic_type in ['technology_evaluation', 'architecture_decision']"

    - skill: "decision-documentation"
      inject_in: [phase_2]
      condition: "topic_type in ['architecture_decision', 'build_vs_buy', 'go_no_go']"
```

When the moderator constructs a sub-agent prompt, it reads the relevant skill files and **appends their content after the phase instructions**. The persona sees its normal phase behavior instructions plus the skill methodology.

---

## Moderator Skill Injection Model

The moderator itself benefits from skill injection at specific phases:

### Convergence Methodology Skill

Loaded during Phase 5, gives the moderator structured methods for synthesis:

```markdown
# Skill: Convergence Methodology

## Argument Weighting Hierarchy
1. User-stated constraints (absolute)
2. Strong empirical evidence (replicated, quantitative)
3. Moderate empirical evidence (credible single-study)
4. Domain-specific reasoning with causal mechanism
5. Theoretical reasoning from first principles
6. Anecdotal evidence (case studies, experience)
7. General best practice (industry convention)

## Severity-Weighted Synthesis
- Count unrebutted MUST-ADDRESS challenges per recommendation
- A recommendation with ≥2 unrebutted MUST-ADDRESS challenges
  gets confidence downgraded by at least 20 percentage points
- A recommendation where all MUST-ADDRESS challenges were conceded
  gets confidence upgraded

## Semantic Grouping
Group findings by MEANING, not by PERSONA:
- Bad: "The Architect said X, the Pragmatist said Y, the Critic said Z"
- Good: "On the question of scalability, three perspectives emerged: {synthesis}"

## Reasoning Chain Audit
For each recommendation in the output:
- Trace the reasoning chain back through positions → challenges → rebuttals
- If any step in the chain was unrebutted, note it
- If the chain includes an [assumed] step, note the assumption as a risk
```

### Quality Gate Skill

Loaded during Phase 2.5 (Diversity Audit) and Phase 5 (Rubric Scoring):

```markdown
# Skill: Quality Gate Methodology

## Deliverable Validation Checklist
See T13 above — full checklist per phase.

## Assessment Protocol
Borrow the Reviewer's evidence-based approach:
- NO self-scoring without evidence citations
- Every rubric dimension score must cite specific output passages
- Dual-path verification for any "5/5" or "1/5" score:
  Forward: "This deserves a 5 because {evidence}"
  Backward: "A score of 5 requires {criteria}. Does the output meet {criteria}?"
```

---

## New Conversation Skills (5 skill files)

### 1. challenge-methodology/SKILL.md

```markdown
# Challenge Methodology

## Open Verification Questions
Frame every challenge as an open question, not a closed assertion.
- "What happens to your recommendation when X doesn't hold?"
- "How does Y interact with your proposed approach?"
- "What would your model predict for scenario Z?"

Open questions (70% diagnostic accuracy) outperform closed assertions
(17% accuracy) because they force the defender to explore, not defend.

## Severity Classification
Classify each challenge:
- **MUST-ADDRESS**: Fundamental flaw that invalidates the recommendation
  if not resolved. The defender must concede, rebut with counter-evidence,
  or revise their recommendation.
- **SHOULD-ADDRESS**: Significant weakness that reduces confidence but
  doesn't invalidate. The defender should respond but may acknowledge
  without revising.
- **COULD-CONSIDER**: Minor point or improvement suggestion. The defender
  may note it for completeness without detailed response.

## Challenge Structure
For each challenge:
1. The open question (mandatory)
2. The specific scenario/evidence motivating the question
3. Your severity classification with justification
4. What you'd need to see in the rebuttal to consider the issue resolved

## Intellectual Honesty
- If a position is genuinely strong in an area, say so
- Don't manufacture disagreement for the sake of appearing adversarial
- "I find no weakness in this argument" is a valid challenge output
```

### 2. evidence-standards/SKILL.md

```markdown
# Evidence Standards

## Dual-Path Verification (required for "strong" claims)
1. Forward: "Evidence E → mechanism M → claim C"
2. Backward: "Claim C requires condition K. Does K hold here?"
If only forward confirms, downgrade to "moderate."

## Evidence Registry Rules
Every empirical claim must appear in the Evidence Registry.
Registry fields: Claim, Source, Type (empirical/theoretical/anecdotal),
Strength (strong/moderate/weak), Replication (replicated/single/unreplicated).

## Evidence Hierarchy
When citing conflicting evidence:
1. Replicated empirical > single-study empirical
2. Controlled experiment > observational study
3. Large N > small N
4. Peer-reviewed > preprint > blog post
5. Primary source > secondary citation

## Claim Strength Rules
- "Strong" requires: replicated empirical evidence directly supporting the claim
- "Moderate" requires: credible single-study evidence OR strong theoretical reasoning
- "Weak" = everything else (including "industry best practice" without data)
```

### 3. reasoning-chains/SKILL.md

```markdown
# Reasoning Chains

## Minimum Chain Length
Every Key Argument and every Recommendation must include a reasoning
chain with ≥2 steps. Single-step rationales are insufficient.

## Chain Format
> Step 1 [empirical]: {observation from data}
> → Step 2 [theoretical]: {implication derived from step 1}
> → Step 3 [assumed]: {conclusion requiring this assumption}
> ∴ Recommendation: {your recommendation}

## Step Labels
- [empirical]: Based on quantitative data or observed behavior
- [theoretical]: Derived from established principles or formal analysis
- [assumed]: Requires an assumption that has not been empirically verified

## Why Chains Matter
Reasoning chains enable:
- **Precise challenges**: Attack step 2 without disputing step 1
- **Concession specificity**: "I concede step 3's assumption is weak,
  but steps 1-2 still hold, so my recommendation changes to..."
- **Convergence tracing**: Moderator can trace why recommendations differ
  (they diverge at step N in the chain)
```

### 4. decision-documentation/SKILL.md

```markdown
# Decision Documentation

## Rejected Alternatives (mandatory)
For each recommendation, document at least 1 alternative you considered
and rejected:

| Alternative | Why Rejected | When It WOULD Be Right |
|-------------|-------------|----------------------|
| {approach} | {Premise → Implication → Conclusion, ≥2 steps} | {scenario} |

## Consequences Section
For your recommended approach, document:
- What it makes **easier** (intended benefits)
- What it makes **harder** (known trade-offs)
- What it **rules out** (options foreclosed by this choice)

## Invisible Knowledge
Note knowledge that is NOT deducible from the recommendation alone:
- Domain constraints that shaped the recommendation
- Trade-offs that were accepted and their cost
- Invariants that must hold for the recommendation to remain valid

## Why Rejected Alternatives Are Valuable
In convergence, the moderator often finds that Persona A's rejected
alternative is Persona B's recommendation. This overlap is the richest
source of productive tension and reveals the true decision axes.
```

### 5. scope-discipline/SKILL.md

```markdown
# Scope Discipline

## Self-Monitoring
You have a defined perspective and priorities. Before writing any argument:
- Is this argument rooted in MY perspective's priorities?
- Or am I drifting into another persona's territory?

## Drift Detection
| If you find yourself... | Diagnosis | Action |
|------------------------|-----------|--------|
| Proposing solutions (from critic/researcher) | Architect drift | State problem, not solution |
| Arguing cost/timeline (from architect/researcher) | Pragmatist drift | State trade-off, not cost |
| Citing evidence without analyzing methodology (from any) | Researcher drift | Flag as delegation request |
| Finding fault without constructive tension (from any) | Critic drift | Check if it reveals genuine tension |

## Flag-Don't-Drift Protocol
If a critical insight falls outside your perspective:
1. Note it in `### Out-of-Scope Observations` at end of output
2. Do NOT develop the argument — one sentence maximum
3. Return to your assigned perspective

This maintains persona diversity throughout the conversation.
If everyone drifts toward the same balanced-middle reasoning,
the dialectic loses its adversarial value.

## Out-of-Scope Observations Format
```
### Out-of-Scope Observations
- [Pragmatist territory] The migration timeline for this approach may exceed 6 months
- [Researcher territory] No benchmarks found for this specific load pattern
```
```

---

## Implementation Priority

### Tier 1: Highest Impact, Lowest Effort (Prompt-level changes)

These are text additions to existing agent markdown files. No structural changes.

| Technique | Target File(s) | Effort |
|-----------|----------------|--------|
| T1: Open verification questions | `conversation-persona*.md` Phase 3 section | Small |
| T4: Scope discipline | `conversation-persona*.md` all phases | Small |
| T10: Rejected alternatives | `conversation-persona*.md` Phase 2 format | Small |
| T15: MUST/SHOULD/COULD severity | `conversation-persona*.md` Phase 3 section | Small |
| T16: Reasoning chains | `conversation-persona*.md` Phase 2 + 4 format | Small |

### Tier 2: High Impact, Medium Effort (Moderator logic changes)

| Technique | Target File(s) | Effort |
|-----------|----------------|--------|
| T3: Argument weighting hierarchy | `conversation-moderator.md` Phase 5 | Medium |
| T13: Deliverable verification | `conversation-moderator.md` phase transitions | Medium |
| T2: Dual-path verification | `conversation-persona*.md` Phase 2 format | Medium |

### Tier 3: Medium Impact, Higher Effort (New file system)

| Technique | Target File(s) | Effort |
|-----------|----------------|--------|
| T8: Skills-as-modules | New `skills/conversation/` directory + 5 skill files | High |
| T12: Topic type classification | `conversation-analysis.yml` + moderator | Medium |

### Recommended Order

```
1. Tier 1 (all 5 together) — single editing session, immediate quality improvement
2. T13 (deliverable verification) — prevents bad output from contaminating later phases
3. T3 (argument weighting) — most impactful moderator change
4. T8 (skills system) — architectural improvement enabling future extensibility
```

---

## Expected Quality Impact

| Quality Rubric Dimension | Primary Techniques | Expected Improvement |
|--------------------------|-------------------|---------------------|
| **Perspective Diversity** | T4 (scope discipline), T10 (rejected alternatives) | +1 to +1.5 — personas stay differentiated, alternatives surface cross-persona overlap |
| **Evidence Quality** | T2 (dual-path), T16 (reasoning chains) | +1 to +1.5 — claims are reality-checked before publication |
| **Concession Depth** | T1 (open questions), T15 (severity classification) | +0.5 to +1 — MUST-ADDRESS challenges force substantive response |
| **Challenge Substantiveness** | T1 (open questions), T15 (severity), T17 (hypothesis-verify) | +1 to +2 — the biggest single improvement area |
| **Synthesis Quality** | T3 (argument weighting), T18 (semantic grouping) | +0.5 to +1 — principled weighting replaces arbitrary synthesis |
| **Actionability** | T10 (rejected alternatives + consequences), T16 (reasoning chains) | +0.5 to +1 — consequences and trade-offs inform action |

The single highest-impact change is **T1 + T15 combined** (open verification questions with severity classification) applied to Phase 3. This transforms the weakest link in most conversation analyses — the cross-examination phase — from superficial adversarialism to diagnostic investigation.

---

*This analysis maps blackbox agent techniques to conversation workflow improvements. The techniques are battle-tested in the blackbox software development workflow and transfer naturally to the dialectic analysis domain.*
