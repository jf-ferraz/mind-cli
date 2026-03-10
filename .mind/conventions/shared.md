# Shared Conventions

Cross-cutting patterns used by **both** the dev workflow (`agents/`) and the conversation workflow (`.github/agents/` + `conversation/`). These conventions are defined once here and referenced by both systems.

## Evidence Standards

All factual claims across both workflows must follow the **Source-or-Flag** mandate:

1. **Cite a source** — add it to an Evidence Registry or inline citation
2. **Flag explicitly** — mark with `[Unsourced assertion]` inline

There is no third option. Claims about adoption rates, performance characteristics, failure modes, industry trends, or "best practices" are factual claims and require sources.

**Canonical reference:** `conversation/skills/evidence-standards/SKILL.md`
**Applied in dev workflow by:** `agents/analyst.md` (requirement validation), `agents/reviewer.md` (claim verification)
**Applied in conversation workflow by:** All personas (Phase 2-4), moderator (Phase 5.0 Evidence Audit)

## Reasoning Chains

Multi-step conclusions must use explicit reasoning chains (≥ 2 steps):

```
evidence/premise → inference/rationale → conclusion/recommendation
```

Single-hop reasoning ("X therefore Y") is insufficient for high-confidence claims. The intermediate step makes the logic auditable and challengeable.

**Canonical reference:** `conversation/skills/reasoning-chains/SKILL.md`
**Applied in dev workflow by:** `agents/architect.md` (design decisions), `agents/reviewer.md` (issue analysis)
**Applied in conversation workflow by:** All personas (Phase 2 key arguments, Phase 4 position changes)

## Severity Classification

Issue severity uses three levels with clear boundaries:

| Level | Meaning | Blocking? |
|-------|---------|-----------|
| **MUST** | Production reliability, data safety, security | Yes |
| **SHOULD** | Project conformance, consistency, maintainability | No |
| **COULD** | Enhancement, polish, future improvement | No |

Dual-path verification is required before declaring MUST violations.

**Canonical reference:** `conventions/severity.md`
**Applied in dev workflow by:** All agents (issue classification)
**Applied in conversation workflow by:** Researcher Phase 3 (unsupported claim severity), moderator (quality threshold enforcement)

## Scope Discipline

Both workflows enforce explicit scope boundaries:

- **Dev workflow:** Agents stay within their role scope (analyst doesn't write code, developer doesn't redesign architecture)
- **Conversation workflow:** Personas use `[OUT OF SCOPE: {observation} — not pursued]` for tangential discoveries

Cross-system: When the conversation workflow is invoked by the dev orchestrator (Mode C / COMPLEX_NEW), the conversation scope is bounded by the orchestrator's request — the convergence analysis should not expand beyond the original design question.

**Canonical reference:** `conversation/skills/scope-discipline/SKILL.md`

## Intent Markers

Both workflows use inline intent markers to distinguish deliberate choices from oversights:

| Marker | Usage |
|--------|-------|
| `[DELIBERATE: {reason}]` | Conscious trade-off, not a gap |
| `:PERF:` | Performance-sensitive code path |
| `:UNSAFE:` | Known unsafe operation (justified) |
| `:SCHEMA:` | Schema-coupled code (migration-sensitive) |
| `:TEMP:` | Temporary solution with known expiry |

**Applied in dev workflow by:** Developer (code annotations), reviewer (trade-off acknowledgment)
**Applied in conversation workflow by:** All personas (evidence trade-offs, architectural trade-offs)

## Confidence Levels

Both workflows use calibrated confidence levels:

| Level | Evidence Requirement |
|-------|---------------------|
| **Very High** | ≥ 2 independent empirical sources |
| **High** | ≥ 2 independent sources (any type) |
| **Medium** | Single source — default ceiling for single-source claims |
| **Low** | No empirical evidence, theoretical or anecdotal only |

Claims must be marked when evidence is insufficient:
- `[Single-source — confidence capped at Medium]`
- `[No evidence — confidence: Low]`
- `[Unsourced assertion]`

**Canonical reference:** `conversation/skills/evidence-standards/SKILL.md` (Dual-Path Confidence Verification)
