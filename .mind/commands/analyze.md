---
name: analyze
description: Run a dialectical conversation analysis to explore a design space before committing to implementation.
---

> **Execution:** Dispatch this entire request to the `conversation-moderator` sub-agent.
> Read `.mind/conversation/agents/moderator.md` for the agent definition.
> Pass the full message content (topic, mode, and any file paths) as the sub-agent's input.
> The moderator will read `.mind/conversation/config/*.yml` and orchestrate all persona sub-agents.
> Do NOT execute the conversation workflow inline — always delegate to the sub-agent.

# /analyze

Explore architectural options, evaluate trade-offs, and produce a convergence analysis with ranked recommendations — before writing any code.

> **Ready to build?** After reviewing the analysis, use `/workflow` to implement the winning recommendation. The orchestrator will automatically read the convergence analysis as context.

## Usage

**Mode A — Generate fresh positions:**
```
/analyze "topic or design question"
```

**Mode B — Analyze existing documents:**
```
/analyze docs/file1.md docs/file2.md
```

## Examples

```
/analyze "Should we use GraphQL or REST for the API layer?"
/analyze "What's the best architecture for a real-time collaborative editor?"
/analyze docs/exploration/option-a.md docs/exploration/option-b.md
```

## What Happens

1. The **conversation-moderator** receives your topic or documents
2. **Mode A**: Spawns 3–5 persona sub-agents to generate independent position papers
3. **Mode B**: Treats each document as a position paper, auto-derives personas
4. **Diversity Audit**: Flags duplicate positions, reports effective persona count
5. **Tension Extraction**: Builds disagreement matrix across positions
6. **Cross-Examination**: Each persona challenges opposing positions with evidence
7. **Rebuttal & Refinement**: Personas concede, rebut, or partially accept
8. **Convergence Analysis**: Evidence audit, decision matrix, consensus map, ranked recommendations
9. **Quality Scoring**: 6-dimension rubric (target ≥ 3.6/5.0)
10. Output saved to `docs/knowledge/{descriptor}-convergence.md`

## Output

| Section | Purpose |
|---------|---------|
| Executive Summary | What was decided, what remains open |
| Evidence Audit | Every empirical claim classified by evidence tier |
| Decision Matrix | Context-aware weighted criteria, all options scored |
| Recommendations | Each with confidence %, risk statement, falsifiability condition |
| Quality Rubric | Scores on all 6 dimensions with justifications |

## Connection to /workflow

- **Before development**: Run `/analyze` → then `/workflow` to implement. The orchestrator reads the convergence analysis automatically.
- **During development**: If classified as `COMPLEX_NEW`, the orchestrator auto-triggers conversation analysis.
- **Standalone**: Use `/analyze` purely for decision-making without building.

## Configuration

Configured in `.mind/conversation/config/`. Key settings: `variants`, `quality_rubric`, `evaluator_optimizer`, `evidence_audit`, `context_aware_criteria`.
