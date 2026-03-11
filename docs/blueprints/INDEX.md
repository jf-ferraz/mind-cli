# Blueprints Index

<!-- Summary index of system-level planning artifacts.
     The orchestrator reads this file first, then loads specific blueprints on demand.
     Token-efficient: agents never load all blueprints — only what's needed. -->

## Active Blueprints

| # | Blueprint | Concern | Lines | Summary |
|---|-----------|---------|-------|---------|
| 01 | [System Architecture](01-system-architecture.md) | Structural backbone | 1066 | 4-layer architecture, design principles, DI, error handling, concurrency |
| 02 | [Domain Model](02-domain-model.md) | Entities & rules | 1459 | 24 entities, enums, business rules, state machines, agent chains |
| 03 | [Data Contracts](03-data-contracts.md) | External formats | 2310 | mind.toml/mind.lock schemas, JSON outputs, MCP tool schemas, exit codes |
| 04 | [CLI Specification](04-cli-specification.md) | Command reference | 2446 | 32 commands fully specified with behavior, output, examples |
| 05 | [TUI Specification](05-tui-specification.md) | Interface design | 1153 | 5-tab wireframes, navigation, styling, watch/orchestration TUI variants |
| 06 | [Reconciliation Engine](06-reconciliation-engine.md) | State tracking | 1042 | Hash computation, dependency graph, staleness propagation algorithm |
| 07 | [AI Workflow Integration](07-ai-workflow-integration.md) | AI bridge | 1419 | 4 integration models: Pre-Flight, MCP Server, Watch, Orchestration |
| 08 | [Implementation Roadmap](08-implementation-roadmap.md) | Delivery plan | 1320 | 6 phases, package structure, testing strategy, CI/CD |

## Document Dependency Graph

```
                    ┌──────────────┐
                    │  01-System   │
                    │ Architecture │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────────┐
              ▼            ▼                ▼
      ┌──────────┐  ┌────────────┐  ┌──────────────┐
      │ 02-Domain│  │ 03-Data    │  │ 06-Reconcile │
      │  Model   │  │ Contracts  │  │   Engine     │
      └────┬─────┘  └─────┬──────┘  └──────┬───────┘
           │               │                │
           ▼               ▼                ▼
      ┌──────────┐  ┌────────────┐  ┌──────────────┐
      │ 04-CLI   │  │ 05-TUI     │  │ 07-AI Work-  │
      │  Spec    │  │  Spec      │  │ flow Integr. │
      └────┬─────┘  └─────┬──────┘  └──────┬───────┘
           └───────────────┼────────────────┘
                           ▼
                    ┌──────────────┐
                    │ 08-Implement │
                    │   Roadmap    │
                    └──────────────┘
```

**Reading order**: 01 → 02 → 03 → 06 → {04, 05, 07} in any order → 08

## Superseded Blueprints

<!-- Original blueprints that were merged and superseded by the 8-document composition above. -->

| Original | Superseded By | Location |
|----------|--------------|----------|
| 01-mind-cli.md | BP-01 + BP-04 + BP-05 | [archive/blueprints-v1/](../../archive/blueprints-v1/01-mind-cli.md) |
| 02-ai-workflow-bridge.md | BP-07 | [archive/blueprints-v1/](../../archive/blueprints-v1/02-ai-workflow-bridge.md) |
| 03-architecture.md | BP-01 + BP-02 + BP-03 | [archive/blueprints-v1/](../../archive/blueprints-v1/03-architecture.md) |

## Template

The [Blueprint Framework Template](BLUEPRINT-FRAMEWORK-TEMPLATE.md) documents this 8-document composition for reuse in future projects.
