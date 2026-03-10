# Blueprints Index

<!-- Summary index of system-level planning artifacts.
     The orchestrator reads this file first, then loads specific blueprints on demand.
     Token-efficient: agents never load all blueprints — only what's needed. -->

## Active Blueprints

<!-- Blueprints currently guiding implementation. -->

| # | Blueprint | Status | Summary |
|---|-----------|--------|---------|
| 01 | [Mind CLI & TUI](01-mind-cli.md) | Active | Go CLI/TUI for the Mind Agent Framework |
| 02 | [AI Workflow Bridge](02-ai-workflow-bridge.md) | Active | AI agent integration models (Pre-Flight, MCP, Sidecar, Orchestration) |
| 03 | [CLI/TUI Software Architecture](03-architecture.md) | Active | 4-layer architecture, design patterns, domain model |

## Completed Blueprints

<!-- Blueprints whose implementation is finished. Historical reference only. -->

## Blueprint Template

<!-- To create a new blueprint:
     1. Copy this structure to a new file: {NN}-{descriptor}.md
     2. Add an entry to the Active Blueprints table above
     3. The orchestrator will discover it via this INDEX.md -->
