# Conversation Analysis Protocols

> **Purpose:** Extracted v2 protocol modules for the conversation moderator agent.
> **Source:** Originally inline in `conversation/agents/moderator.md` (Sprint 7 extraction).
> **Pattern:** One protocol per directory, each with a `PROTOCOL.md` file.

---

## Protocol Matrix

| ID | Protocol | Config Gate | Applies To | Lines | Sprint Origin |
|----|----------|-------------|------------|-------|---------------|
| P1 | [Phase Routing](phase-routing/PROTOCOL.md) | `phase_routing` section exists | After every phase | 137 | Sprint 3 |
| P2 | [Evaluator-Optimizer Loop](evaluator-optimizer/PROTOCOL.md) | `evaluator_optimizer.enabled: true` | Phase 5 → 5.5 loop | 202 | Sprint 4 |
| P3 | [Approval Gates](approval-gates/PROTOCOL.md) | `approval_gates.enabled: true` | After Phases 2, 4, 5 | 69 | Sprint 5 |
| P4 | [Mediated Delegation](mediated-delegation/PROTOCOL.md) | `delegation.enabled: true` | After every sub-agent | 70 | Sprint 5 |

---

## How Protocols Work

### Relationship to Skills

| Concept | Skills (`conversation/skills/`) | Protocols (`conversation/protocols/`) |
|---------|--------------------------------|--------------------------------------|
| **Content type** | Quality techniques injected into persona prompts | Procedural instructions for the moderator |
| **Consumed by** | Personas (via moderator injection) | Moderator directly |
| **Trigger** | Phase-based (`skill_injection.rules`) | Config-gated (feature enabled/disabled) |
| **Side effects** | None — additive content only | State file writes, routing decisions, workflow control |
| **Pattern** | `SKILL.md` per directory | `PROTOCOL.md` per directory |

### Loading Behavior

1. At workflow start, the moderator reads `protocol_loading` from `conversation/config/extensions.yml`
2. When a protocol's config gate evaluates to true, the moderator reads the corresponding `PROTOCOL.md`
3. The protocol instructions are followed **in addition to** the moderator's core workflow execution steps
4. Protocols reference config values from `conversation/config/*.yml` — they don't embed config

### Protocols That Remain Inline

These v2 sections are small enough to stay in the moderator file:

| Section | Lines | Reason |
|---------|-------|--------|
| State Management | 56 | Foundational — every other feature depends on it |
| Context Rule Enforcement | 35 | Tightly coupled to prompt construction in every phase |
| Dynamic Speaker Selection | 44 | Small; interleaves with Phase 3/4 workflow steps |
| Deliverable Verification Protocol | 29 | Tightly coupled to phase transition gates |

---

## Adding New Protocols

1. Create a new directory under `conversation/protocols/`
2. Add a `PROTOCOL.md` with the protocol instructions
3. Add the config gate to `conversation/config/extensions.yml`
4. Add a reference stub in the moderator file
5. Update this index table
