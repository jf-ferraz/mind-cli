---
name: workflow
description: Entry point for the agent framework. Classifies the request and dispatches to the correct agent chain.
---

> **Execution:** Dispatch to the orchestrator agent at `.mind/agents/orchestrator.md`.

# /workflow

Start the agent workflow by describing what you need.

> **Have a vague idea?** Use `/discover "idea"` first to produce a project brief.

## Usage

```
/workflow "description of what you need"
```

## Examples

```
/workflow "Create a REST API for user management with authentication"
/workflow "Fix the 500 error on the /api/users endpoint"
/workflow "Add WebSocket support for real-time notifications"
/workflow "Refactor the data access layer to use repository pattern"
```

## What Happens

1. The **orchestrator** scans the workspace and classifies your request
2. For **COMPLEX_NEW**: conversation-moderator runs dialectical analysis first (Gate 0: quality ≥ 3.0/5.0)
3. Creates a git branch and iteration folder in `docs/iterations/`
4. Selects the minimal agent chain (3–6 agents depending on type)
5. Each agent runs sequentially, reading the previous agent's output
6. Quality gates check deliverables between phases
7. The **reviewer** provides final sign-off with evidence-based validation

## Classification Triggers

| Prefix | Classification |
|--------|---------------|
| `fix:` or `bug:` | BUG_FIX |
| `add:` or `feature:` | ENHANCEMENT |
| `refactor:` | REFACTOR |
| `create:` or `new:` | NEW_PROJECT |
| `analyze:` or `explore:` | COMPLEX_NEW |

## Agent Roles

| Agent | Role | File |
|-------|------|------|
| Orchestrator | Classify, route, manage gates | `.mind/agents/orchestrator.md` |
| Analyst | Scope requirements, domain model | `.mind/agents/analyst.md` |
| Architect | Structure, components, API contracts | `.mind/agents/architect.md` |
| Developer | Implement code following specs | `.mind/agents/developer.md` |
| Tester | Test strategy + implementation | `.mind/agents/tester.md` |
| Reviewer | Evidence-based review, sign-off | `.mind/agents/reviewer.md` |

## Session Management

For large workflows, the orchestrator splits sessions with a structured handoff in `docs/state/workflow.md`. Resume with:
```
/workflow "Resume the interrupted workflow"
```
