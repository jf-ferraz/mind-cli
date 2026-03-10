# Conversation Analysis — Skill Index

> **Canonical source** for conversation persona quality techniques.
> Persona agent files reference these skills; the moderator auto-injects them per phase via `skill_injection` config.

## Skills

| Skill | Applies To | Phases | Sprint Origin |
|-------|-----------|--------|---------------|
| [Challenge Methodology](challenge-methodology/SKILL.md) | All personas — Phase 3 | Cross-Examination | S1 + S2 |
| [Evidence Standards](evidence-standards/SKILL.md) | All personas — Phase 2 | Opening Position | S6 |
| [Reasoning Chains](reasoning-chains/SKILL.md) | All personas — Phase 2 + Phase 4 | Opening Position, Rebuttal | S5 |
| [Decision Documentation](decision-documentation/SKILL.md) | All personas — Phase 2 | Opening Position | S4 |
| [Scope Discipline](scope-discipline/SKILL.md) | All personas — All phases | All | S3 |

## How Skills Are Used

1. **Auto-injection:** The moderator reads `skill_injection` rules from `conversation/config/extensions.yml` and appends the relevant skill content to each sub-agent prompt based on the current phase.
2. **Inline reference:** Persona agent files contain the same techniques inline (added in Sprint 1). The skills here are the **canonical source** — if there's a discrepancy, these files govern.
3. **Manual inclusion:** Users can reference individual skills when constructing custom persona prompts.

## Design Principles

- **One skill per concern** — each skill file covers exactly one quality technique
- **Phase-tagged** — every skill declares which phases it applies to
- **Additive** — skills enhance persona output quality without changing persona identity
- **Standalone** — each skill is self-contained and can be read independently
