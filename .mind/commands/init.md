---
name: init
description: Guided setup — detect missing documentation and create what's needed.
---

> **Execution:** This is an interactive agent command — do NOT delegate to a sub-agent.
> Execute the steps below directly, using tool calls as needed.

# /init

Detect what documentation exists, what's missing, and interactively create missing components.

## Workflow

### Step 1: Run the documentation validator

Run `bash .mind/scripts/validate-docs.sh` against the project root. Capture and parse the output.

If `validate-docs.sh` is not found at `.mind/scripts/validate-docs.sh`, check if the framework was installed correctly. The script should exist if `install.sh` was used.

### Step 2: Report current state

Present a summary to the user grouped by zone:

| Zone | What to report |
|------|---------------|
| `spec/` | Which of the 3 required files exist vs. missing (`project-brief.md`, `requirements.md`, `architecture.md`), plus `decisions/` status |
| `blueprints/` | Whether `INDEX.md` exists, count of blueprint files |
| `state/` | Whether `current.md` and `workflow.md` exist |
| `iterations/` | Count and naming compliance |
| `knowledge/` | Whether `glossary.md` exists, spike/convergence file count |

Classify each component as **present**, **missing**, or **stub** (exists but has only template content).

### Step 3: Interactive creation

For each missing or stub component, ask the user whether to create it:

| Component | How to create |
|-----------|---------------|
| ADRs | `bash .mind/scripts/docs-gen.sh adr "<title>"` |
| Blueprints | `bash .mind/scripts/docs-gen.sh blueprint "<title>"` |
| Iterations | `bash .mind/scripts/docs-gen.sh iteration <type> "<descriptor>"` |
| Spikes | `bash .mind/scripts/docs-gen.sh spike "<title>"` |
| Convergence | `bash .mind/scripts/docs-gen.sh convergence "<title>"` |
| `project-brief.md` | Create directly using the structure from `scaffold.sh` (Vision, Target Users, Problem Statement, Key Deliverables, Success Metrics, Scope, Constraints, Technical Preferences, Core Domain Concepts, Open Questions) |
| `requirements.md` | Create directly using the structure from `scaffold.sh` (Overview, Functional Requirements, Non-Functional Requirements, Constraints, Acceptance Criteria) |
| `architecture.md` | Create directly using the structure from `scaffold.sh` (System Overview, Component Map, Data Model, API Contracts, Key Decisions, Boundaries) |
| `domain-model.md` | Create directly using the structure from `scaffold.sh` (Entities, Relationships, Business Rules, State Machines, Constraints) |
| `state/current.md` | Create with Active Work, Known Issues, Recent Changes, Next Priorities sections |
| `state/workflow.md` | Create with managed-by-orchestrator comment |
| `knowledge/glossary.md` | Create with Term/Definition table |
| `decisions/` directory | `mkdir -p docs/spec/decisions` |

**Do not batch-create everything silently.** Ask the user about each missing component or group them logically (e.g., "The spec/ zone is missing 3 files — create all?").

### Step 4: Re-validate

Run `bash .mind/scripts/validate-docs.sh` again and show the updated results. If all checks pass, confirm the setup is complete.

## Output

A validated 5-zone documentation structure with all required files present.

## Connection to Other Commands

- After `/init`, use `/discover` to fill in `project-brief.md` interactively
- After `/init`, use `/workflow` to start building with the full agent pipeline
- Run `bash .mind/scripts/validate-docs.sh` independently at any time to check compliance
- Run `bash .mind/scripts/docs-gen.sh list` to audit documentation inventory
