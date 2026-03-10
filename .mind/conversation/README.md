# Conversation Module

A structured dialectical analysis system that produces convergence analyses through multi-persona debate. Designed to explore design spaces, evaluate trade-offs, and surface hidden assumptions before implementation begins.

## How It Works

The conversation-moderator orchestrates 3–5 specialist personas through a phased workflow:

1. **Opening Positions** (Phase 2) — Each persona independently generates a position paper
2. **Diversity Audit & Tension Extraction** (Phase 2.5) — Flags duplicate positions, reports effective persona count, builds matrix of disagreements
3. **Cross-Examination** (Phase 3) — Personas challenge opposing positions with evidence
4. **Rebuttal & Refinement** (Phase 4) — Personas respond: conceding, rebutting, or partially accepting
5. **Convergence Synthesis** (Phase 5) — Evidence audit, decision matrix, consensus map, ranked recommendations
6. **Quality Scoring** — 6-dimension rubric (target ≥ 3.6/5.0)
7. **Output Assembly** (Phase 6) — Final document saved to `docs/knowledge/`

## Entry Points

| Method | Usage |
|--------|-------|
| `/analyze "topic"` | Standalone analysis — explore trade-offs without building |
| `/workflow "analyze: topic"` | Integrated — orchestrator auto-triggers for COMPLEX_NEW requests |
| Mode B | `/analyze docs/file1.md docs/file2.md` — treat existing documents as positions |

## Directory Structure

```
conversation/
├── config/
│   ├── conversation.yml    ← Phases, routing, context rules, termination
│   ├── personas.yml        ← Persona library, presets, variants, selection guide
│   ├── quality.yml         ← 6-dimension rubric, evaluator-optimizer loop
│   └── extensions.yml      ← Skill injection, protocols, gates, delegation
├── protocols/
│   ├── PROTOCOLS.md        ← Protocol index
│   ├── phase-routing/      ← Directed graph phase transitions
│   ├── evaluator-optimizer/ ← Closed-loop quality refinement
│   ├── approval-gates/     ← Human-in-the-loop checkpoints
│   └── mediated-delegation/ ← Agent-to-agent delegation via moderator
└── skills/
    ├── SKILLS.md           ← Skill index
    ├── evidence-standards/ ← Source-or-Flag, confidence verification
    ├── reasoning-chains/   ← Premise → rationale → conclusion
    ├── challenge-methodology/ ← Open verification questions, severity
    ├── decision-documentation/ ← Rejected alternatives with reasoning
    └── scope-discipline/   ← Out-of-scope flagging
```

## Platform Entry Points

| Platform | Entry Point | Agent Location |
|----------|-------------|----------------|
| Claude Code | `/analyze "topic"` or COMPLEX_NEW routing | `.mind/conversation/agents/` (primary source) |
| Copilot Chat | `@conversation-moderator` or `@analyze` prompt | `.github/agents/conversation-*.md` (synced) |

Both platforms share: `.mind/conversation/config/`, `.mind/conversation/protocols/`, `.mind/conversation/skills/`, `.github/state/`, `docs/knowledge/`.

See `.mind/README.md` for the framework architecture overview.

## Related Files

- **Agents (primary)**: `.mind/conversation/agents/moderator.md` + 5 persona agents
- **Agents (Copilot)**: `.github/agents/conversation-*.md` (synced from primary via `.mind/scripts/sync-agents.sh`)
- **Prompts**: `.github/prompts/analyze.prompt.md`, `.github/prompts/analyze-documents.prompt.md`
- **Output**: `docs/knowledge/` — convergence analyses consumed by dev workflow
- **Shared conventions**: `.mind/conventions/shared.md` — evidence, reasoning, severity patterns
- **Command docs**: `.mind/commands/analyze.md`
- **Architecture**: `.mind/README.md` — framework overview

## Custom Persona Registration

You can extend the persona pool beyond the built-in 4 specialists by creating custom persona files in `conversation/specialists/`:

1. Create a markdown file (e.g., `conversation/specialists/security-auditor.md`)
2. Add the required frontmatter:
   ```yaml
   ---
   type: conversation-persona
   name: "The Security Auditor"
   perspective: "Attack surface and threat modeling..."
   priorities: [Attack surface minimization, Authentication strength, ...]
   bias_disclosure: "May over-prioritize security at the expense of DX."
   agent_mapping: conversation-persona
   ---
   ```
3. The moderator auto-discovers these at session start and adds them to the available pool
4. Reference by filename slug in variant presets or persona selection: `security_auditor`

Required frontmatter fields: `type`, `name`, `perspective`, `priorities`, `bias_disclosure`, `agent_mapping`.

## Convergence Diff

When you run a conversation analysis on a topic that has a prior convergence file in `docs/knowledge/`, the moderator automatically produces a **Convergence Diff** section appended to the new output. This tracks:

| Category | What it shows |
|----------|---------------|
| Recommendation Shifts | Which recommendations changed direction or priority |
| New Evidence | Evidence in this run not present in the prior |
| Score Changes | Per-dimension quality score delta (↑/↓/=) |
| Consensus Stability | Which consensus points held vs. shifted |
| Persona Differences | Different composition, new perspectives introduced |

The diff includes a trajectory verdict: **STABILIZING**, **SHIFTING**, or **DIVERGING**. This is useful for iterative design exploration — run the same topic multiple times as understanding deepens.

## History Persistence

Enable `history.persist_history: true` in `conversation/config/conversation.yml` to save full conversation artifacts after each analysis:

```
docs/knowledge/history/{session-id}/
├── session-metadata.yml   ← Topic, personas, scores, timestamps
├── convergence.md         ← Copy of final convergence output
├── quality-scores.yml     ← Per-dimension rubric scores
├── state-snapshot.yml     ← Final state file
└── phases/                ← Position papers, challenges, rebuttals
```

The `history.include` list in `conversation.yml` controls which artifact types are persisted. History files are git-ignored by default.

## Session ID Namespacing

Each conversation session generates a unique ID in the format `{topic-slug}-{YYYY-MM-DD}-{platform}` (e.g., `graphql-vs-rest-2026-02-25-copilot`). This prevents state collisions when:

- Multiple analyses run concurrently
- The same topic is analyzed on both Claude Code and Copilot Chat
- Re-running an analysis on the same day

Configure in `conversation/config/conversation.yml` under `session_id:` (format, state directory, cleanup policy, TTL).

## Integration with Dev Workflow

The conversation module integrates with the main agent workflow at two points:

1. **COMPLEX_NEW classification** — The orchestrator auto-dispatches conversation-moderator before the dev pipeline. Gate 0 validates convergence quality ≥ 3.0/5.0.
2. **Downstream consumption** — Analyst and architect read the convergence analysis from `docs/knowledge/` via the iteration overview's "Prior Analysis Context" section. Developer, tester, and reviewer also reference it.
