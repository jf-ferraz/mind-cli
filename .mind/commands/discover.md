---
name: discover
description: Interactive project discovery. Explores vision, constraints, and deliverables before formal requirements.
---

> **Execution:** Dispatch to the discovery agent at `.mind/agents/discovery.md`.

# /discover

Explore and define a project idea before building. The discovery agent asks targeted questions and produces a structured project brief.

## Usage

```
/discover "your project idea"
```

## Examples

```
/discover "I want to build an inventory management system for small warehouses"
/discover "I need a tool that monitors our microservices and alerts on failures"
/discover "A mobile-friendly dashboard for tracking sales metrics"
```

## What Happens

1. The **discovery** agent reads your description and identifies knowledge gaps
2. It asks 5–8 targeted questions in small batches (2–3 at a time)
3. You answer conversationally — no formal format needed
4. It extracts domain concepts: core entities, business rules, key workflows
5. It synthesizes your answers into `docs/spec/project-brief.md`
6. You review and confirm the brief
7. When ready, run `/workflow` to start building

## When to Use

- **Before `/workflow`** — when you have an idea but haven't defined specifics
- **New projects** — to extract vision, users, deliverables, boundaries
- **Major features** — to clarify scope before formal requirements analysis

## When to Skip

- You already have clear, detailed requirements
- The task is a bug fix, refactor, or small enhancement

> **Architectural uncertainty?** Run `/analyze "design question"` to explore options before committing.

## After Discovery

**Direct to implementation:**
```
/workflow "Build the system described in the project brief"
```

**Explore trade-offs first:**
```
/analyze "What architecture best fits the requirements in the project brief?"
```
