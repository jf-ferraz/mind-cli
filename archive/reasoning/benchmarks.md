---
title: "AI Agent/Sub-Agent Workflow Exploration — Benchmarks & Source Cards"
date: "2026-02-24"
status: "draft"
scope: "software-development agent workflows"
---

## How to read this appendix

Each benchmark card is intentionally **operational**: it describes what a system *actually does* (topology, context/memory, tool/runtime, artifacts, governance, and observability), with **links to primary sources**.

For the main exploration document, these cards act as the evidence base for patterns and trade-offs.

### Rubric (applied per card)

- **Workflow topology**: single-agent, supervisor/worker, planner/executor, graph/state-machine, group-chat, etc.
- **Task decomposition & control**: delegation, retries, fallbacks, validation loops, HITL checkpoints.
- **Context/memory**: repo-map/indexing, summarization, persistent stores, context boundaries.
- **Tools/runtime**: tool handshake, permissioning, sandboxing (Docker/VM), command execution.
- **Artifacts & traceability**: file conventions, manifests, decision logs, mapping tasks→changes.
- **Governance & quality**: tests, guardrails/policies, approvals, failure recovery.
- **Dev workflow**: git discipline, PR creation/review automation, CI integration.
- **Extensibility**: plugins/hooks, MCP, adapters.
- **Observability**: logs, traces/spans, replay/time-travel.
- **Trade-offs**: what it optimizes for; where it breaks.

---

## Card: SWE-bench / SWE-bench Verified (benchmark + harness)

- **Type**: Benchmark + evaluation harness for “fix a real GitHub issue in a repo; judged by tests.”
- **Workflow topology**: Not a workflow itself; defines the *task shape* most SWE agents target.
- **Task decomposition**: Implied by evaluation constraints (agents must localize, edit, run tests).
- **Tools/runtime**: Containerized evaluation is standard; harness runs tests to judge success.
- **Governance/quality**: Verified subset adds human validation of issue statements/tests.
- **Key references**
  - SWE-bench repo: `https://github.com/swe-bench/swe-bench`
  - SWE-bench paper (ICLR 2024): `https://huggingface.co/papers/2310.06770`
  - SWE-bench Verified overview: `https://epoch.ai/benchmarks/swe-bench-verified`

---

## Card: SWE-agent (issue→code changes in sandbox; research-friendly)

- **Workflow topology**: Single agent loop with tool use (read/edit/run tests) inside sandbox.
- **Tools/runtime**: Strong emphasis on Docker images + reproducible environments.
- **Governance/quality**: Config-driven (YAML), cost limits; success judged by tests.
- **Dev workflow**: Oriented around issue URLs; commonly used in benchmark-style runs.
- **Key references**
  - Repo: `https://github.com/princeton-nlp/SWE-agent`
  - Docs (Docker install/config): `https://swe-agent.com/0.7/installation/docker`

---

## Card: Agentless (localize→repair baseline; “less agent, more pipeline”)

- **Workflow topology**: Pipeline: localization stage → candidate patch generation/filtering.
- **Core insight**: Strong performance can come from *structured, narrow* workflows vs general agents.
- **Trade-off**: Less flexible than tool-using agents but often cheaper/more reliable.
- **Key references**
  - Paper: `https://arxiv.org/html/2407.01489v1`

---

## Card: OpenHands (OpenDevin) (platform: SDK + CLI + UI; sandboxed execution)

- **Workflow topology**: Platform supports multiple agent styles; execution typically via agent server.
- **Tools/runtime**: Docker sandbox recommended; workspace mounting patterns for local repos.
- **Artifacts & traceability**: “Skills” system for repo-specific behavior; supports GitHub workflows.
- **Dev workflow**: Cloud GitHub integration + PR review workflows; @mention-based interaction.
- **Key references**
  - Repo: `https://github.com/OpenDevin/OpenDevin`
  - Docs: `https://docs.openhands.dev/`
  - Docker sandbox guide: `https://docs.openhands.dev/sdk/guides/agent-server/docker-sandbox`
  - Design principles: `https://docs.openhands.dev/sdk/arch/design`
  - PR review workflow: `https://docs.openhands.dev/sdk/guides/github-workflows/pr-review`

---

## Card: Aider (CLI; git-native + repo-map context compression)

- **Workflow topology**: Interactive single-agent pair-programming loop (human+agent) with git as control plane.
- **Context/memory**: “Repository map” built from symbols; graph-ranked to fit token budget.
- **Dev workflow**: Auto-commits; `/diff` and `/undo` operationalize safe iteration.
- **Key references**
  - Repo map docs: `https://aider.chat/docs/repomap.html`
  - Git integration docs: `https://aider.chat/docs/git.html`

---

## Card: Continue (IDE; explicit tool handshake + permission gating + MCP)

- **Workflow topology**: Single agent with structured tool-calling loop (“tool handshake”).
- **Governance**: Tool permission gate (manual vs automatic); Plan mode vs Agent mode tool sets.
- **Extensibility**: MCP servers provide additional tools beyond built-ins.
- **Key references**
  - Agent mode handshake: `https://docs.continue.dev/ide-extensions/agent/how-it-works`

---

## Card: Claude Code (CLI; subagents + tool allowlisting + MCP)

- **Workflow topology**: Primary agent with delegable **subagents** (separate context windows).
- **Governance**: `--allowedTools` allowlisting for safe automation boundaries.
- **Extensibility**: MCP server config and custom subagent definitions.
- **Key references**
  - Claude Code docs: `https://docs.anthropic.com/en/docs/claude-code/`
  - Subagents: `https://docs.anthropic.com/en/docs/claude-code/sub-agents`
  - CLI reference: `https://docs.anthropic.com/en/docs/claude-code/cli-reference`

---

## Card: OpenAI Codex CLI (CLI; skill system + universal container environment)

- **Workflow topology**: Local agent workflow in terminal; extensible via “skills”.
- **Extensibility**: Skill directories (scoped search paths) + progressive disclosure.
- **Runtime isolation**: “Universal” Docker image available as open source (codex-universal).
- **Key references**
  - Repo: `https://github.com/openai/codex`
  - Skills docs: `https://developers.openai.com/codex/skills`
  - Universal environment repo: `https://github.com/openai/codex-universal`
  - Cloud environments docs: `https://developers.openai.com/codex/cloud/environments/`

---

## Card: OpenAI Agents SDK (library; handoffs + guardrails + sessions + tracing)

- **Workflow topology**: Supports multi-agent via handoffs (handoff-as-tool pattern).
- **Governance**: Guardrails (input/output/tool) with parallel vs blocking modes.
- **Memory/state**: Sessions abstraction (e.g., RedisSession) for persistent conversation state.
- **Observability**: Built-in tracing (traces/spans) with pluggable processors; sensitive-data controls.
- **Key references**
  - Repo: `https://github.com/openai/openai-agents-python`
  - Tracing: `https://openai.github.io/openai-agents-python/tracing/`
  - Guardrails (Python ref): `https://openai.github.io/openai-agents-python/ref/guardrail/`
  - Sessions: `https://openai.github.io/openai-agents-python/sessions/`
  - Handoffs: `https://openai.github.io/openai-agents-python/ref/handoffs/`
  - RedisSession: `https://openai.github.io/openai-agents-python/ref/extensions/memory/redis_session/`

---

## Card: LangGraph (graph/state machine for agents; persistence, interrupts, time travel)

- **Workflow topology**: Explicit state machine/graph; nodes+edges; supports loops, branching, parallel super-steps.
- **State**: Typed state schemas + reducers; can separate input/output/internal state.
- **HITL**: `interrupt()` + `Command(resume=...)` for pause/resume workflows (requires persistence).
- **Durability/debug**: Checkpointers (e.g., Postgres) + **time travel** debugging via checkpoints/history.
- **Key references**
  - Multi-agent concepts: `https://langchain-ai.github.io/langgraph/concepts/multi_agent/`
  - Interrupts: `https://docs.langchain.com/oss/python/langgraph/interrupts`
  - Persistence how-to: `https://langchain-ai.github.io/langgraph/how-tos/persistence/`
  - Postgres checkpointer: `https://langchain-ai.github.io/langgraph/how-tos/persistence_postgres/`
  - Time travel concepts: `https://langchain-ai.github.io/langgraph/concepts/time-travel/`

---

## Card: AutoGen + Magentic-One (multi-agent conversation + orchestrator)

- **Workflow topology**: Conversational multi-agent; Magentic-One uses an orchestrator coordinating specialists.
- **Tools/runtime**: Includes tool agents for files, web, terminal, coding (in Magentic-One).
- **Benchmarking**: AutoGenBench emphasizes repetition, isolation (Docker), instrumentation.
- **Key references**
  - AutoGen repo: `https://github.com/microsoft/autogen`
  - Magentic-One report: `https://aka.ms/magentic-one-report`
  - Magentic-One (AutoGen docs): `https://microsoft.github.io/autogen/dev/user-guide/agentchat-user-guide/magentic-one.html`
  - AutoGenBench: `https://microsoft.github.io/autogen/blog/2024/01/25/AutoGenBench/`

---

## Card: CrewAI (crews vs flows; event-driven orchestration)

- **Workflow topology**: “Crews” for autonomous teams; “Flows” for structured event-driven orchestration.
- **State/control**: Flows provide explicit control flow (start/listen), shared state, and lifecycle events.
- **Key references**
  - Flows concept: `https://docs.crewai.com/concepts/flows`

---

## Card: LlamaIndex AgentWorkflow + CodeActAgent (handoffs + code-as-action)

- **Workflow topology**: Multi-agent with handoff constraints + shared context; async/event-driven.
- **Tools/runtime**: CodeActAgent generates code that uses provided functions; requires careful sandboxing.
- **Key references**
  - Multi-agent systems: `https://docs.llamaindex.ai/en/stable/understanding/agent/multi_agent/`
  - CodeActAgent: `https://docs.llamaindex.ai/en/stable/examples/agent/code_act_agent/`
  - CodeAct paper: `https://arxiv.org/abs/2402.01030`

---

## Card: Semantic Kernel Orchestration (patterns: sequential/concurrent/handoff/group/magentic)

- **Workflow topology**: Multiple orchestration patterns exposed with consistent APIs.
- **Key references**
  - Agent orchestration overview: `https://learn.microsoft.com/en-us/semantic-kernel/frameworks/agent/agent-orchestration/`
  - Magentic orchestration: `https://learn.microsoft.com/en-us/semantic-kernel/frameworks/agent/agent-orchestration/magentic`

---

## Card: MetaGPT / ChatDev (role-based “virtual software company” SDLC)

- **Workflow topology**: Role-based agents aligned to SDLC responsibilities; structured communication/SOPs.
- **Artifacts**: Tends to generate structured docs (PRDs, designs) as explicit intermediate outputs.
- **Trade-off**: Strong structure, but often heavy/complex; quality depends on discipline of interfaces and validation.
- **Key references**
  - MetaGPT paper: `https://arxiv.org/abs/2308.00352`
  - MetaGPT repo: `https://github.com/geekan/MetaGPT`
  - ChatDev paper: `https://arxiv.org/abs/2307.07924`
  - ChatDev repo: `https://github.com/OpenBMB/ChatDev`

---

## Card: MCP (Model Context Protocol) (tool ecosystem + boundaries + consent model)

- **Type**: Protocol standard for tools/resources/prompts; defines host/client/server model.
- **Governance**: Security principles emphasize explicit consent, tool safety, data privacy, and boundary “roots”.
- **Key references**
  - Spec: `https://modelcontextprotocol.io/specification/latest`
  - Roots concept: `https://modelcontextprotocol.io/docs/concepts/roots`
  - Security best practices (spec): `https://modelcontextprotocol.io/specification/2025-11-25/basic/security_best_practices`

---

## Card: GitHub-native agent automation (gh-aw + Patchwork + Jules)

### gh-aw (GitHub Agentic Workflows)

- **Workflow topology**: “Workflows as markdown” compiled to GitHub Actions; LLM processor is configurable.
- **Governance**: Security-first; read-only by default; controlled write outputs.
- **Key references**
  - Repo: `https://github.com/githubnext/gh-aw`
  - How they work: `https://github.github.io/gh-aw/introduction/how-they-work/`

### Patchwork

- **Workflow topology**: Scripted patchflows built from atomic “steps” and prompt templates.
- **Dev workflow**: Runs locally or in CI (GitHub Actions); oriented around PR review/auto-fix tasks.
- **Key references**
  - Repo: `https://github.com/patched-codes/patchwork`
  - Running patchflows: `https://docs.patched.codes/running-patchflow`

### Jules (Google Labs; GitHub Action integration)

- **Workflow topology**: Async “cloud VM” coding agent triggered by GitHub events/labels.
- **Key references**
  - GitHub action: `https://github.com/google-labs-code/jules-action`
  - Docs: `https://jules.google/docs/`

---

## Card: Durability / workflow engines (Temporal, DBOS, Prefect)

### Temporal AI cookbook (durable agent/MCP patterns)

- **Workflow topology**: External workflow engine ensures retries/idempotency and long-running state.
- **Key references**
  - AI cookbook: `https://docs.temporal.io/ai-cookbook`
  - Durable MCP example: `https://docs.temporal.io/ai-cookbook/hello-world-durable-mcp-server`

### DBOS

- **Workflow topology**: Library adds durable workflows/steps backed by Postgres (checkpoint outputs per step).
- **Key references**
  - How workflows work: `https://docs.dbos.dev/explanations/how-workflows-work`
  - Step tutorial (Python): `https://docs.dbos.dev/python/tutorials/step-tutorial`

### Prefect (durable execution for agent loops)

- **Workflow topology**: Python-native control flow + durable task caching; supports HITL.
- **Key references**
  - PydanticAI durable execution with Prefect: `https://ai.pydantic.dev/durable_execution/prefect/`

---

## Card: Observability stacks (Phoenix, Langfuse, Weave, LangSmith)

### OpenAI Agents SDK Tracing (baseline model)

- **Key references**
  - Tracing: `https://openai.github.io/openai-agents-python/tracing/`

### Phoenix (OTEL/OTLP)

- **Key references**
  - Phoenix tracing overview: `https://docs.arize.com/phoenix/tracing/llm-traces`

### Langfuse (OTEL backend)

- **Key references**
  - OTel support: `https://langfuse.com/docs/opentelemetry`

### Weave (ops/calls)

- **Key references**
  - Tracing docs: `https://docs.wandb.ai/weave/guides/tracking/tracing`

### LangSmith (LangGraph/LangChain tracing; distributed tracing)

- **Key references**
  - Trace LangGraph apps: `https://docs.smith.langchain.com/observability/how_to_guides/trace_with_langgraph`
  - Distributed tracing (agent server): `https://docs.langchain.com/langsmith/agent-server-distributed-tracing`
