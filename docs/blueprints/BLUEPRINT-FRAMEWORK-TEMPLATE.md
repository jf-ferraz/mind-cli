# Blueprint Framework Template

> A meta-template documenting the 8-document blueprint composition for creating complete implementation specifications for any software project.

**Status**: Template
**Date**: 2026-03-11

---

## 1. Introduction

The Blueprint Framework is an 8-document composition that together provides a full-spectrum specification for guiding software implementation. Each document owns a clear, non-overlapping concern. Duplication is replaced by cross-references --- if a fact lives in one blueprint, every other blueprint points to it rather than restating it.

**Philosophy**: Separation of specification concerns mirrors separation of code concerns. Just as a well-architected system has modules with clear boundaries, a well-specified system has documents with clear boundaries.

**Goal**: Any line of implementation code should trace back to exactly one blueprint section. If a developer asks "where is this specified?", the answer is always one document, one section --- never "it's spread across three blueprints."

**Audience**: Development teams (human or AI-assisted) who need unambiguous implementation guidance. The framework assumes the team has already completed discovery and high-level design. Blueprints capture the *what* and *how* of implementation, not the *why* of the product (that belongs in a project brief or requirements document).

---

## 2. Document Composition

| # | Document | Concern | Answers |
|---|----------|---------|---------|
| 01 | System Architecture | Structural backbone | "How is the system structured? What are the rules?" |
| 02 | Domain Model | Entities, rules, invariants | "What are the core entities, relationships, and business rules?" |
| 03 | Data Contracts & Schemas | External formats | "What does every external format look like, field by field?" |
| 04 | Command/API Specification | Interface reference | "What does every command/endpoint do, exactly?" |
| 05 | UI Specification | User interface design | "What does every screen look like and how does the user interact?" |
| 06 | Core Engine/Algorithm | Key algorithmic subsystem | "How does the core algorithm/engine work, end to end?" |
| 07 | Integration Architecture | External system connections | "How do integrations work and connect to external systems?" |
| 08 | Implementation Roadmap | Delivery plan | "What do we build, in what order, and how do we verify it?" |

---

## 3. Document Dependency Graph

```
                    ┌──────────────┐
                    │  01-System   │
                    │ Architecture │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────────┐
              ▼            ▼                ▼
      ┌──────────┐  ┌────────────┐  ┌──────────────┐
      │ 02-Domain│  │ 03-Data    │  │ 06-Core      │
      │  Model   │  │ Contracts  │  │   Engine     │
      └────┬─────┘  └─────┬──────┘  └──────┬───────┘
           │               │                │
           ▼               ▼                ▼
      ┌──────────┐  ┌────────────┐  ┌──────────────┐
      │ 04-API   │  │ 05-UI      │  │ 07-Integra-  │
      │  Spec    │  │  Spec      │  │ tion Arch.   │
      └────┬─────┘  └─────┬──────┘  └──────┬───────┘
           └───────────────┼────────────────┘
                           ▼
                    ┌──────────────┐
                    │ 08-Implement │
                    │   Roadmap    │
                    └──────────────┘
```

**Reading order**: 01 first (everything depends on it), then 02/03/06 (second tier --- can be read in parallel), then 04/05/07 (third tier --- can be read in any order), then 08 last (synthesizes everything above into a delivery plan).

**Why this order matters**: System Architecture establishes the structural constraints that every other document must respect. The Domain Model, Data Contracts, and Core Engine define the *what* --- entities, formats, algorithms. The API Spec, UI Spec, and Integration Architecture define *how users and systems interact* with that core. The Roadmap sequences everything into buildable increments.

---

## 4. Per-Document Templates

Each template below includes:
- A document header with metadata
- Required sections with guidance on what belongs in each
- A "Does NOT contain" boundary to prevent overlap
- Example content stubs using `{placeholder}` notation

---

### BP-01: System Architecture

> How is the system structured? What are the rules?

```markdown
# BP-01: System Architecture

> How is the system structured? What are the rules?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Cross-references**: [BP-02](02-domain-model.md) for entity details, [BP-03](03-data-contracts.md) for schemas

---

## 1. Design Principles

### P1: {Principle Name}

{One paragraph explaining the principle and why it matters. Each principle is a
non-negotiable constraint that shapes every architectural decision below. Aim for
3-7 principles. Too few and you lack guidance; too many and nothing is prioritized.}

### P2: {Principle Name}

{...}

---

## 2. Architectural Layers

{ASCII diagram showing the layer stack, top to bottom. Each layer has a clear
responsibility and dependency direction (upper layers depend on lower layers, never
the reverse).}

```
┌─────────────────────────────────┐
│  Presentation / Interface Layer │  ← User-facing: CLI, API, UI
├─────────────────────────────────┤
│  Application / Orchestration    │  ← Use cases, workflows, coordination
├─────────────────────────────────┤
│  Domain / Business Logic        │  ← Entities, rules, invariants
├─────────────────────────────────┤
│  Infrastructure / Adapters      │  ← Storage, external services, I/O
└─────────────────────────────────┘
```

### Layer Rules

| Layer | May Depend On | Must Not Depend On |
|-------|---------------|-------------------|
| Presentation | Application, Domain | Infrastructure directly |
| Application | Domain | Presentation |
| Domain | Nothing | Any outer layer |
| Infrastructure | Domain (interfaces) | Application, Presentation |

---

## 3. Component Map

{Package/module structure showing every top-level component and its responsibility.
Each component has a single owner (one team or one developer).}

```
{project}/
├── cmd/              # Entry points
│   └── {binary}/     # Main binary
├── internal/
│   ├── {component}/  # {Responsibility}
│   ├── {component}/  # {Responsibility}
│   └── {component}/  # {Responsibility}
├── pkg/              # Public library code (if any)
└── config/           # Configuration defaults
```

| Component | Layer | Responsibility |
|-----------|-------|---------------|
| `{component}` | {Layer} | {One-line description} |

---

## 4. Dependency Injection

{How components are wired together at startup. Describe the initialization order,
dependency graph, and whether you use a DI container, manual wiring, or functional
composition.}

---

## 5. Error Handling Strategy

### Error Types

| Type | When | Example |
|------|------|---------|
| Validation Error | Input fails constraints | "{field} must be non-empty" |
| Domain Error | Business rule violated | "{entity} cannot transition from {state_a} to {state_b}" |
| Infrastructure Error | External system failure | "failed to read {file}: permission denied" |

### Propagation Rules

{How errors flow from origin to user. Do you wrap errors? Use error codes? Return
typed errors or strings?}

### User-Facing Conversion

{How internal errors become user-visible messages. What information is shown vs. hidden.}

---

## 6. Configuration Model

### Config File Format

{File name, format (TOML/YAML/JSON), location discovery order.}

### Loading Precedence

{Order of precedence: defaults → config file → environment variables → CLI flags.}

### Defaults

{Every configuration value and its default. Link to BP-03 for the complete schema.}

---

## 7. Output Formatting

| Mode | Trigger | Format |
|------|---------|--------|
| Human | Default | Styled text with colors, tables, alignment |
| Machine | `--json` / `Accept: application/json` | Stable JSON (schema in BP-03) |
| Quiet | `--quiet` / `-q` | Exit code only, no stdout |

---

## 8. Concurrency Model

{Threading/async model. Single-threaded? Thread pool? Event loop? Goroutine-based?
Describe synchronization primitives and any shared mutable state.}

---

## 9. Logging & Observability

| Level | When Used | Example |
|-------|-----------|---------|
| ERROR | Unrecoverable failure | "failed to parse {file}: {reason}" |
| WARN | Degraded but functional | "{feature} unavailable, using fallback" |
| INFO | Key lifecycle events | "loaded configuration from {path}" |
| DEBUG | Diagnostic detail | "resolved {entity} with {n} fields" |

---

## 10. Security Considerations

### Threat Model

{What are the trust boundaries? What inputs are untrusted?}

### Input Validation

{Validation strategy for all external inputs.}

### Permissions

{File system permissions, network access, credential handling.}
```

**Does NOT contain**: Specific API/command behaviors (BP-04), entity field lists (BP-02), file schemas (BP-03), UI layouts (BP-05), algorithm internals (BP-06).

---

### BP-02: Domain Model

> What are the core entities, relationships, and business rules?

```markdown
# BP-02: Domain Model

> What are the core entities, relationships, and business rules?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md)
**Cross-references**: [BP-03](03-data-contracts.md) for serialization formats

---

## 1. Entity Catalog

### {EntityName}

**Purpose**: {Why this entity exists — one sentence.}

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `{field}` | `string` | Yes | Non-empty, max 128 chars | {What it represents} |
| `{field}` | `{EnumType}` | Yes | See Value Objects | {What it represents} |
| `{field}` | `[]string` | No | Unique items | {What it represents} |

{Repeat for every entity in the domain.}

---

## 2. Value Objects & Enums

### {EnumName}

**Used by**: {EntityName}.{field}

| Value | Meaning |
|-------|---------|
| `{value_a}` | {Description} |
| `{value_b}` | {Description} |

{Repeat for every enum and value object.}

---

## 3. Entity Relationships

```
{EntityA} 1──* {EntityB}    "{EntityA} contains many {EntityB}s"
{EntityB} *──1 {EntityC}    "{EntityB} references one {EntityC}"
{EntityA} 1──1 {EntityD}    "{EntityA} has exactly one {EntityD}"
```

{ASCII diagram showing all relationships. Use 1──1, 1──*, *──* notation.
Include cardinality and a brief phrase describing each relationship.}

---

## 4. Business Rules

### BR-{NN}: {Rule Name}

**Applies to**: {EntityName}
**Rule**: {Precise statement of the invariant or constraint.}
**Enforced at**: {Where in the architecture this is checked — domain layer, validation layer, database constraint.}
**Violation behavior**: {What happens when the rule is broken — error type, message, recovery.}

{Repeat for every business rule. Number them for traceability.}

---

## 5. Lifecycle State Machines

### {EntityName} Lifecycle

```
               ┌─────────┐
               │ Created  │
               └────┬─────┘
                    │ {trigger}
                    ▼
               ┌─────────┐
          ┌────│  Active  │────┐
          │    └─────────┘    │
  {trigger}│               │{trigger}
          ▼                ▼
     ┌─────────┐     ┌──────────┐
     │ Paused  │     │ Completed│
     └────┬────┘     └──────────┘
          │ {trigger}
          ▼
     ┌─────────┐
     │  Active  │  (re-enters Active)
     └─────────┘
```

| From | To | Trigger | Guard Condition |
|------|----|---------|-----------------|
| Created | Active | {event} | {condition or "none"} |
| Active | Completed | {event} | {condition} |

{Repeat for every entity that has meaningful state transitions.}

---

## 6. Aggregate Boundaries

| Aggregate Root | Contains | Consistency Guarantee |
|----------------|----------|----------------------|
| {EntityA} | {EntityB}, {EntityC} | Transactional — all or nothing |
| {EntityD} | {EntityE} | Eventually consistent |

{Define which entities are always modified together (same transaction/operation)
and which can be independently updated.}
```

**Does NOT contain**: Serialization formats or file schemas (BP-03), storage/persistence implementation details, API behavior or command descriptions (BP-04).

---

### BP-03: Data Contracts & Schemas

> What does every external format look like, field by field?

```markdown
# BP-03: Data Contracts & Schemas

> What does every external format look like, field by field?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md), [BP-02](02-domain-model.md)
**Cross-references**: [BP-04](04-api-spec.md) for which commands produce which schemas

---

## 1. Configuration File Schema

**File**: `{config-file-name}`
**Format**: {TOML | YAML | JSON}
**Location**: {Discovery order — e.g., `./{file}`, `~/.config/{project}/{file}`}

### Top-Level Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `{field}` | `string` | No | `"{default}"` | {What it controls} |

### Section: `[{section}]`

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `{field}` | `{type}` | {Yes/No} | `{default}` | {What it controls} |

{Repeat for every section. Include validation rules where they differ from
the domain model constraints in BP-02.}

---

## 2. State/Lock File Schema

{If the project maintains state files (lock files, cache files, checkpoint files),
define them here. Same field-by-field format as above.}

**File**: `{state-file-name}`
**Format**: {format}
**Written by**: {Which operation creates/updates this file}
**Read by**: {Which operations consume this file}

---

## 3. API/CLI Output Schemas

### {OperationName} Output

**Produced by**: `{command}` or `{endpoint}`
**Format**: JSON

```json
{
  "{field}": "{type — string | number | boolean | object | array}",
  "{field}": {
    "{nested_field}": "{type}"
  }
}
```

| Field | Type | Always Present | Description |
|-------|------|----------------|-------------|
| `{field}` | `string` | Yes | {What it represents} |

{Repeat for every operation that produces structured output.}

---

## 4. Exit/Status Codes

| Code | Name | Meaning |
|------|------|---------|
| 0 | Success | Operation completed successfully |
| 1 | General Error | Unspecified failure |
| 2 | Usage Error | Invalid arguments or flags |
| {N} | {Name} | {Meaning} |

### Per-Operation Exit Codes

| Operation | Success | Partial | Failure |
|-----------|---------|---------|---------|
| `{command}` | 0 | {N} — {when} | 1 |

---

## 5. External Tool Schemas

{If the project exposes tools for external consumption (MCP tools, plugins,
extension points), define each tool's input/output schema here.}

### Tool: `{tool_name}`

**Description**: {What the tool does}

**Input Schema**:
```json
{
  "{param}": "{type}",
  "{param}": "{type}"
}
```

**Output Schema**:
```json
{
  "{field}": "{type}"
}
```

---

## 6. Log/Event Format

### Structured Log Entry

```json
{
  "timestamp": "ISO-8601",
  "level": "ERROR | WARN | INFO | DEBUG",
  "message": "string",
  "context": {
    "{key}": "{value}"
  }
}
```

{If the project emits events (webhooks, pub/sub messages), define those schemas
here as well.}

---

## 7. Versioning & Migration

**Schema version**: `{version}`
**Version field**: `{field_name}` in `{file}`
**Migration strategy**: {How old schemas are detected and upgraded — automatic migration, error with guidance, backwards compatibility window.}
```

**Does NOT contain**: Parsing implementation, field rationale or design justification (that belongs in decision records), API behavior descriptions (BP-04).

---

### BP-04: Command/API Specification

> What does every command/endpoint do, exactly?

```markdown
# BP-04: Command/API Specification

> What does every command/endpoint do, exactly?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md), [BP-02](02-domain-model.md)
**Cross-references**: [BP-03](03-data-contracts.md) for output schemas, [BP-05](05-ui-spec.md) for interactive views

---

## 1. Command/Endpoint Tree

```
{project}
├── {verb}                    # {Brief description}
│   ├── {noun}                # {Brief description}
│   └── {noun}                # {Brief description}
├── {verb} {noun}             # {Brief description}
└── {verb}                    # {Brief description}
```

{For REST APIs, use route tree instead:}
```
{base_url}
├── GET    /                  # {Brief description}
├── POST   /{resource}        # {Brief description}
├── GET    /{resource}/:id    # {Brief description}
└── DELETE /{resource}/:id    # {Brief description}
```

---

## 2. Global Parameters

{Flags/headers available on all operations.}

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--json` / `Accept: application/json` | boolean | false | Machine-readable output |
| `--verbose` / `-v` | boolean | false | Verbose output |
| `--quiet` / `-q` | boolean | false | Suppress non-error output |
| `{flag}` | `{type}` | `{default}` | {Description} |

---

## 3. Per-Operation Specification

### `{command}` / `{METHOD} {path}`

**Synopsis**: `{project} {verb} {noun} [args] [flags]`
**Purpose**: {One sentence — what this operation does and when you use it.}

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `{arg}` | `string` | Yes | {What it is} |

#### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--{flag}` | `-{f}` | `{type}` | `{default}` | {What it controls} |

#### Behavior

1. {Step 1 — what happens first}
2. {Step 2 — validation, processing, etc.}
3. {Step 3 — output generation}

#### Output

**Human mode**:
```
{Example human-readable output}
```

**JSON mode** (see BP-03 for full schema):
```json
{
  "{field}": "{example_value}"
}
```

#### Error Codes

| Exit/Status | Condition | Message |
|-------------|-----------|---------|
| 0 / 200 | Success | — |
| 1 / 400 | {Condition} | "{User-facing message}" |

#### Examples

```bash
# {Description of what this example demonstrates}
{project} {command} {args} {flags}
```

{Repeat this entire block for every operation.}

---

## 4. Completion/Discovery

### Auto-Completion

{How shell completion / API discovery works. What gets completed — commands,
arguments, flag values? What data sources feed completion?}

### Help System

{How `--help` / API documentation is structured and generated.}

---

## 5. Error Messages Catalog

| Code | Message | Remediation |
|------|---------|-------------|
| `{ERROR_CODE}` | "{User-facing message}" | {What the user should do} |

{Comprehensive list of all user-facing error messages. Each message should be
actionable — it tells the user what went wrong and how to fix it.}
```

**Does NOT contain**: Internal implementation details, JSON schema definitions (BP-03), domain entity field lists (BP-02), UI layout specifications (BP-05).

---

### BP-05: UI Specification

> What does every screen look like and how does the user interact?

```markdown
# BP-05: UI Specification

> What does every screen look like and how does the user interact?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md), [BP-02](02-domain-model.md)
**Cross-references**: [BP-04](04-api-spec.md) for data operations behind each view

---

## 1. Screen Layouts

### {ScreenName} View

**Entry point**: {How the user reaches this screen — URL, command, navigation action.}
**Purpose**: {What the user accomplishes on this screen.}

```
┌─────────────────────────────────────────────────┐
│  {Header / Navigation Bar}                      │
├─────────────────────────────────────────────────┤
│                                                 │
│  {Main Content Area}                            │
│                                                 │
│  ┌──────────────┐  ┌────────────────────────┐   │
│  │ {Sidebar /   │  │ {Detail Panel}         │   │
│  │  List Panel} │  │                        │   │
│  │              │  │                        │   │
│  └──────────────┘  └────────────────────────┘   │
│                                                 │
├─────────────────────────────────────────────────┤
│  {Status Bar / Footer}                          │
└─────────────────────────────────────────────────┘
```

{Repeat for every distinct screen/view in the application.}

---

## 2. Navigation Model

### Key Bindings / Interactions

| Input | Context | Action |
|-------|---------|--------|
| `{key}` / `{gesture}` | {Screen or global} | {What happens} |
| `{key}` / `{gesture}` | {Screen} | {What happens} |

### Navigation Flow

```
{ScreenA} ──{action}──▶ {ScreenB}
{ScreenB} ──{action}──▶ {ScreenC}
{ScreenB} ──{action}──▶ {ScreenA}  (back)
```

---

## 3. Component Hierarchy

```
Application
├── {TopLevelComponent}
│   ├── {ChildComponent}
│   └── {ChildComponent}
├── {TopLevelComponent}
│   ├── {ChildComponent}
│   └── {ChildComponent}
└── {SharedComponent}
```

| Component | Responsibility | Data Source |
|-----------|---------------|-------------|
| `{Component}` | {What it renders} | {Where it gets data — API call, prop, state} |

---

## 4. State Management

### Per-View Data Requirements

| View | Data Needed | Refresh Strategy |
|------|-------------|-----------------|
| {ScreenName} | {Entities/fields required} | {On mount / polling / event-driven} |

### Shared State

{What state is shared across views. How is it synchronized.}

---

## 5. Responsive Design

{If applicable. For terminal UIs, describe minimum terminal size and
how layouts adapt to different widths.}

| Breakpoint | Layout Change |
|------------|--------------|
| `< {width}` | {How layout adapts} |
| `>= {width}` | {Standard layout} |

---

## 6. Styling & Theme

### Color Scheme

| Element | Color / Style | Purpose |
|---------|--------------|---------|
| {Element} | {Color or ANSI code} | {Why this color} |

### Typography

{Font choices, sizes, emphasis rules. For TUIs: bold, dim, underline usage.}

---

## 7. Accessibility

- **Keyboard navigation**: {All functionality reachable via keyboard}
- **Screen reader support**: {ARIA labels, semantic structure, alt text}
- **Non-color indicators**: {Symbols, text labels, or patterns alongside color}
- **Contrast**: {Minimum contrast ratios or light/dark theme support}
```

**Does NOT contain**: Implementation code, data format definitions (BP-03), API/command behavior logic (BP-04), algorithm details (BP-06).

---

### BP-06: Core Engine/Algorithm

> How does the core algorithm/engine work, end to end?

```markdown
# BP-06: Core Engine/Algorithm

> How does the core algorithm/engine work, end to end?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md)
**Cross-references**: [BP-02](02-domain-model.md) for entity definitions, [BP-03](03-data-contracts.md) for input/output formats

---

## 1. Problem Statement

{Why this engine/algorithm exists. What problem it solves. Why a simpler approach
is insufficient. Keep this to 2-3 paragraphs.}

---

## 2. Input/Output Specification

### Inputs

| Input | Type | Source | Description |
|-------|------|--------|-------------|
| `{input}` | `{type}` | {Where it comes from} | {What it represents} |

### Outputs

| Output | Type | Consumer | Description |
|--------|------|----------|-------------|
| `{output}` | `{type}` | {Who uses it} | {What it represents} |

---

## 3. Algorithm Description

### Overview

{High-level description in 3-5 sentences.}

### Step-by-Step

```
FUNCTION {engine_name}({inputs}):
    // Phase 1: {Phase name}
    {step 1}
    {step 2}

    // Phase 2: {Phase name}
    FOR EACH {item} IN {collection}:
        {step 3}
        IF {condition}:
            {step 4}
        ELSE:
            {step 5}

    // Phase 3: {Phase name}
    {step 6}
    RETURN {output}
```

{Accompany pseudocode with prose explanations for non-obvious steps.
Pseudocode is for precision; prose is for understanding.}

---

## 4. Data Structures

### {StructureName}

**Purpose**: {Why this structure exists in the algorithm.}
**Representation**: {Array, tree, hash map, graph, etc.}

```
{StructureName} {
    {field}: {type}     // {role in algorithm}
    {field}: {type}     // {role in algorithm}
}
```

{Describe access patterns — how the algorithm reads, writes, and queries
this structure.}

---

## 5. Integration Points

| Consumer | Interface | Data Flow |
|----------|-----------|-----------|
| {Component from BP-01} | `{function/method signature}` | {Input → Engine → Output} |

{How other components in the system invoke the engine and consume its results.
Reference BP-01 component map for where these consumers live.}

---

## 6. Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Latency (p50) | `< {N}ms` | {How to measure} |
| Latency (p99) | `< {N}ms` | {How to measure} |
| Throughput | `{N} ops/sec` | {How to measure} |
| Memory | `< {N}MB` for {workload} | {How to measure} |

---

## 7. Edge Cases

| # | Scenario | Expected Behavior |
|---|----------|-------------------|
| 1 | {Edge case description} | {What the engine does} |
| 2 | {Edge case description} | {What the engine does} |
| 3 | Empty input | {What the engine does} |
| 4 | Maximum-size input | {What the engine does} |
| 5 | Malformed input | {What the engine does} |
```

**Does NOT contain**: JSON schema definitions (BP-03), CLI flag details (BP-04), UI layout specifications (BP-05), integration protocol details (BP-07).

---

### BP-07: Integration Architecture

> How do integrations work and connect to external systems?

```markdown
# BP-07: Integration Architecture

> How do integrations work and connect to external systems?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: [BP-01](01-system-architecture.md), [BP-06](06-core-engine.md)
**Cross-references**: [BP-03](03-data-contracts.md) for tool/event schemas, [BP-04](04-api-spec.md) for commands that trigger integrations

---

## 1. Integration Philosophy

{Design principles specific to how this project connects to external systems.
Examples: "all integrations are optional", "fail gracefully when unavailable",
"integrations are pluggable behind interfaces".}

---

## 2. Per-Integration Specification

### Integration: {IntegrationName}

**External system**: {What it connects to}
**Direction**: Inbound | Outbound | Bidirectional
**Required**: Yes | No (graceful degradation)

#### Protocol

{How communication happens — HTTP REST, gRPC, CLI subprocess, file system,
message queue, WebSocket, etc.}

#### Data Flow

```
{project} ──{data}──▶ {ExternalSystem}
{ExternalSystem} ──{response}──▶ {project}
```

#### Authentication

{How credentials are provided and managed. Reference BP-01 Security Considerations
for credential storage policy.}

#### Error Handling

| Failure Mode | Detection | Recovery |
|-------------|-----------|----------|
| {System unavailable} | {Timeout / error code} | {Retry / fallback / error message} |
| {Invalid response} | {Schema validation} | {Reject / partial accept / log warning} |

#### Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `{setting}` | `{type}` | `{default}` | {What it controls} |

{Repeat this entire block for every integration point.}

---

## 3. Shared Concerns

### Retry Strategy

{Default retry policy — max attempts, backoff strategy, which errors are retryable.}

### Timeout Policy

{Default and per-integration timeouts.}

### Circuit Breaker

{If applicable — when integrations are disabled after repeated failures.}

### Rate Limiting

{Outbound rate limits to respect external system constraints.}

---

## 4. Abstraction Layer

{How integrations are decoupled from core logic. Interface/adapter pattern?
Plugin system? Dependency injection? Show the boundary.}

```
┌───────────────────┐
│   Core Logic      │  Depends on interface, not implementation
│   (BP-01 Domain)  │
└────────┬──────────┘
         │ {IntegrationInterface}
         ▼
┌───────────────────┐
│   Adapter Layer   │  Implements interface per external system
├───────────────────┤
│ {AdapterA}        │ → {ExternalSystemA}
│ {AdapterB}        │ → {ExternalSystemB}
│ {MockAdapter}     │ → In-memory (testing)
└───────────────────┘
```

---

## 5. Incremental Delivery

| Phase | Integrations Included | Rationale |
|-------|----------------------|-----------|
| Phase 1 | {IntegrationA} | {Why this first — core dependency, highest value} |
| Phase 2 | {IntegrationB}, {IntegrationC} | {Why these next} |
| Phase N | {IntegrationD} | {Lower priority / optional} |
```

**Does NOT contain**: Tool schema definitions (BP-03), command/endpoint flags and behavior (BP-04), UI for integration configuration (BP-05).

---

### BP-08: Implementation Roadmap

> What do we build, in what order, and how do we verify it?

```markdown
# BP-08: Implementation Roadmap

> What do we build, in what order, and how do we verify it?

**Status**: Draft | Active | Final
**Date**: {YYYY-MM-DD}
**Depends on**: All preceding blueprints (01-07)

---

## 1. Phase Overview

```
Phase 1          Phase 2          Phase 3          Phase N
{Name}           {Name}           {Name}           {Name}
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ {scope}  │────▶│ {scope}  │────▶│ {scope}  │────▶│ {scope}  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
{duration}       {duration}       {duration}       {duration}
```

---

## 2. Per-Phase Detail

### Phase {N}: {Phase Name}

**Goal**: {One sentence — what is true when this phase is done.}
**Duration estimate**: {Time range}
**Prerequisites**: {Phase N-1 completion / external dependency}

#### Scope

| Deliverable | Blueprint Source | Description |
|-------------|----------------|-------------|
| `{package/module}` | BP-{NN} Section {N} | {What gets built} |
| `{feature}` | BP-{NN} Section {N} | {What gets built} |

#### Acceptance Criteria

- [ ] {Measurable criterion — not "works well" but "passes N tests" or "handles case X"}
- [ ] {Measurable criterion}
- [ ] {Measurable criterion}

#### Dependencies

{What from previous phases must be complete. What external dependencies are needed.}

{Repeat for every phase.}

---

## 3. Package/Module Structure

{Complete project layout at the end of all phases. Mark which phase introduces each
package. This is the target state.}

```
{project}/
├── cmd/                          # Phase 1
│   └── {binary}/
│       └── main.go
├── internal/
│   ├── {package_a}/              # Phase 1 — {responsibility}
│   ├── {package_b}/              # Phase 1 — {responsibility}
│   ├── {package_c}/              # Phase 2 — {responsibility}
│   └── {package_d}/              # Phase 3 — {responsibility}
├── pkg/                          # Phase 2
│   └── {public_package}/
├── config/                       # Phase 1
│   └── defaults.{ext}
└── test/
    ├── integration/              # Phase 1
    └── fixtures/                 # Phase 1
```

---

## 4. Testing Strategy

| Layer | Test Type | Tools | Coverage Target |
|-------|-----------|-------|-----------------|
| Domain | Unit tests | {framework} | {target}% |
| Application | Integration tests | {framework} | {target}% |
| Interface | End-to-end tests | {framework} | Key paths |
| Infrastructure | Contract tests | {framework} | All adapters |

### Testing Principles

{Testing philosophy — what gets tested, what does not. When to use mocks vs. real
dependencies. Test naming conventions.}

---

## 5. Build & CI

### Build

| Concern | Tool | Configuration |
|---------|------|---------------|
| Build | {tool} | {config file} |
| Lint | {tool} | {config file} |
| Format | {tool} | {config file} |
| Dependencies | {tool} | {config file} |

### CI Pipeline

```
Push/PR → Lint → Test → Build → {Additional steps}
```

### Release

{How releases are cut. Versioning scheme. Artifact distribution.}

---

## 6. Definition of Done

A phase is complete when ALL of the following are true:

- [ ] All acceptance criteria for the phase pass
- [ ] All tests pass (unit, integration, end-to-end as applicable)
- [ ] No lint errors or warnings
- [ ] Documentation updated (if user-facing changes)
- [ ] {Project-specific gate — e.g., "JSON output matches schema in BP-03"}
- [ ] {Project-specific gate}
```

**Does NOT contain**: Architectural decisions or rationale (BP-01), feature specifications (BP-02 through BP-07), code implementation.

---

## 5. Adapting for Different Project Types

The 8-document framework is designed to be universal, but the *emphasis* and *interpretation* of each document shifts depending on the project type. The structure stays the same; the content adapts.

### API / Backend Projects

| Document | Adaptation |
|----------|------------|
| BP-04 | Becomes **API Specification** — REST/GraphQL/gRPC endpoints replace CLI commands |
| BP-05 | Optional, or becomes **Admin Dashboard** spec if an admin UI exists |
| BP-06 | Could be Database/Query Engine, Event Processing Pipeline, Business Rule Engine |
| BP-07 | Covers external API integrations, message queues, third-party services, databases |

### Web Frontend Projects

| Document | Adaptation |
|----------|------------|
| BP-04 | Becomes **Route Specification** — pages and navigation replace commands |
| BP-05 | Critical and likely the largest document — page layouts, component library, design system |
| BP-06 | Could be State Management Engine, Rendering Pipeline, Offline Sync Engine |
| BP-07 | Covers API client layer, authentication flow, analytics, third-party widgets |

### Library / SDK Projects

| Document | Adaptation |
|----------|------------|
| BP-04 | Becomes **Public API Reference** — functions, classes, and types replace commands |
| BP-05 | Optional, or becomes Documentation Site spec |
| BP-06 | The core algorithm the library provides — this is often the largest document |
| BP-07 | Covers runtime integrations, plugin system, extension points |

### Data Pipeline Projects

| Document | Adaptation |
|----------|------------|
| BP-04 | Becomes **Pipeline Configuration Specification** — job definitions, DAG structure |
| BP-05 | Becomes **Monitoring Dashboard** spec |
| BP-06 | The data processing/transformation engine — typically the most complex document |
| BP-07 | Covers data source connectors, sink integrations, orchestrator integration |

### Mobile Application Projects

| Document | Adaptation |
|----------|------------|
| BP-04 | Becomes **Screen Action Specification** — user actions and their effects |
| BP-05 | Critical — screen designs, gesture interactions, platform-specific layouts |
| BP-06 | Could be Offline Sync Engine, Local Database Engine, Media Processing Pipeline |
| BP-07 | Covers backend API client, push notifications, platform services, analytics |

---

## 6. Guarantees

When all 8 documents are complete and internally consistent, the framework guarantees:

**No ambiguity during implementation.** Every entity, schema, operation, screen, and algorithm is specified to the level of detail needed to write code without guessing. If an implementer has to make a judgment call, a blueprint is incomplete.

**No overlap.** Each document owns a clear concern. The "Does NOT contain" boundary on every template prevents specification drift. If content appears in two places, one of them is wrong --- replace it with a cross-reference.

**Traceable coverage.** Any line of implementation code traces back to one blueprint section. During code review, you can ask "which blueprint section specifies this?" and get a single answer.

**Independent workstreams.** After BP-01 and BP-02 are stable, teams can work on BP-04, BP-05, BP-06, and BP-07 in parallel. The dependency graph makes this explicit --- no team blocks another unless they share a tier-2 dependency.

**Testable acceptance criteria.** BP-08 defines what "done" means for every phase, grounded in specifications from BP-01 through BP-07. Acceptance criteria are measurable, not subjective.

---

## 7. Anti-Patterns

### Duplicating content across blueprints

If the same field list appears in BP-02 and BP-03, they will inevitably diverge. The domain model (BP-02) defines what fields exist and their constraints. The data contracts (BP-03) define how those fields are serialized. BP-03 references BP-02 for the canonical field list --- it does not repeat it.

### Writing implementation code in blueprints

Blueprints are specifications, not codebases. Pseudocode in BP-06 is acceptable for algorithm clarity. Code examples in BP-04 are acceptable for illustrating usage. But if a blueprint contains copy-paste-ready implementation, it has gone too far and will become stale the moment the real code diverges.

### Making blueprints too detailed for later phases

Phase 1 specifications should be concrete enough to implement without questions. Phase 5 specifications can be high-level, with the understanding that they will be refined when Phase 5 approaches. Specifying distant phases in full detail wastes effort on decisions that will change.

### Treating blueprints as immutable

Blueprints are living documents. When an implementation decision invalidates a specification, update the blueprint. The blueprint should always reflect the current intended behavior, not the original plan. Track significant changes in decision records.

### Skipping the "Does NOT contain" boundaries

The boundaries exist to prevent overlap. When a contributor is unsure where something belongs, the boundaries provide the answer. Removing or ignoring them leads to the duplication anti-pattern above.

### Writing blueprints in isolation

Each blueprint should be reviewed against its dependencies. BP-04 should be reviewed against BP-02 (does every entity have CRUD operations?) and BP-03 (does every output reference a schema?). Cross-document consistency is the framework's primary value.
