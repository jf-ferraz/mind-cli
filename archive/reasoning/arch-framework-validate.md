# Mind Framework — Final Unified Architectural Blueprint

Status: Authoritative design reference for next-phase backlog and planning  
Scope basis: `project-definition.md` (Phases 1-3 only; Phase 4 and MVP measurement prompt excluded) + all proposals in `documents/architecture/`

Terminology status legend:

- **[Confirmed]** explicitly supported across source proposals and/or prior decisions
- **[Inferred]** strongly implied by multi-document convergence but not fully codified
- **[Open]** unresolved; requires validation before architecture freeze

## A. Executive Summary

The Mind Framework target architecture is a **declarative, agent-orchestrated, and operationally traceable system** centered on a canonical manifest (`mind.toml`) and computed lock state (`mind.lock`). **[Confirmed]**

Key design decisions:

- Canonical model = `mind.toml` (intent) + `mind.lock` (resolved state) + canonical URI scheme (`doc:{zone}/{name}`) **[Confirmed]**
- Orchestrator is the reconciliation authority that computes deltas and coordinates agent workflows **[Confirmed]**
- Four-zone documentation topology structures framework knowledge and runtime state (`spec`, `state`, `iterations`, `knowledge`) **[Confirmed]**
- Operational layer (`.mind/`) manages cache, logs, outputs, and session continuity **[Confirmed]**
- Quality is enforced through deterministic gates and evidence-based review before closure **[Confirmed]**

This blueprint enables:

- consistent architectural decisions across iterations,
- faster and safer agent workflows through indexed context and dependency-aware updates,
- stronger governance and traceability from requirement to artifact to validation evidence.

## B. Scope and Objectives

### Scope Included

- Canonical architecture model and layer boundaries
- Operational layer design and dependency flows
- Governance and architectural decision model
- Iteration lifecycle and framework evolution rules
- Risks, gaps, and validation agenda

### Scope Excluded

- Implementation playbooks and build instructions
- Task decomposition, sprint planning, or execution sequencing
- Language/runtime selection finalization for implementation internals
- Code-level design decisions

### Architectural Goals

- Establish one coherent architecture reference for all future planning artifacts
- Minimize ambiguity in ownership, interfaces, and lifecycle transitions
- Keep system declarative, modular, auditable, and agent-agnostic
- Improve context efficiency, traceability, and operational reliability

### Non-Goals

- Defining delivery roadmap details
- Selecting final performance tuning tactics
- Designing plugin internals beyond architectural boundaries

## C. Architectural Principles and Guidelines

### Design Principles

- **Declarative first:** desired state is declared; runtime computes and reconciles actual state.
- **Single source of architectural truth:** manifest-centric architecture with explicit provenance.
- **Separation of concerns:** canonical model, orchestration, operations, and integrations remain bounded.
- **Traceability by design:** every key artifact is addressable and lineage-aware.
- **Incremental formalization:** architecture supports layered maturity without structural rewrites.

### Operating Principles

- Reconciliation over ad hoc mutation
- Deterministic quality gates before closure
- Structured session handoff and state persistence
- Explicit dependency modeling over implicit coupling

### Decision Rules and Guardrails

- Manifest schema changes require architecture decision record and compatibility review.
- New module/layer additions must declare boundaries, inputs/outputs, and dependency impact.
- Cross-layer coupling is allowed only via defined contracts (schema, URI, gate result format).
- Operational outputs are managed as lifecycle artifacts, not primary system-of-record documents.

### Standards and Boundaries

- Canonical identity standard: URI-based artifact references
- Documentation zoning standard: `spec/state/iterations/knowledge`
- Governance standard: explicit ownership + review checkpoints per lifecycle stage

## D. Final Architecture Blueprint (Core Model)

### D.1 High-Level Architecture View

1. **Canonical Model Layer**  

   `mind.toml`, `mind.lock`, URI registry, dependency graph metadata.

2. **Orchestration & Governance Layer**  

   Orchestrator, workflow chains, reconciliation engine, quality-gate policy, decision control.

3. **Operational Runtime Layer**  

   `.mind/` runtime state, indexing/cache services, gate outputs, logs, handoff artifacts.

4. **Integration & Extension Layer**  

   Git/GitHub integration, environment/container adapters, optional MCP/plugin/hook interfaces.

### D.2 Main Layers and Responsibilities

- **Canonical Model:** represent intended system state and artifact relationships.
- **Orchestration:** derive required actions from state deltas and coordinate agents.
- **Operational Runtime:** execute/support workflows, collect runtime evidence, preserve continuity.
- **Integration:** connect architecture to external tools without contaminating core model semantics.

### D.3 Operational Layer Design

Core operational capabilities:

- incremental lock synchronization (fast-path metadata checks + full refresh fallback),
- context budgeting and summary caching for agent efficiency,
- structured run/gate/audit outputs with retention controls,
- workflow/session handoff artifacts for multi-session continuity.

Operational layer is **supporting infrastructure**, not replacement for canonical architecture artifacts. **[Confirmed]**

### D.4 Component Boundaries

- **Manifest Authority Component:** owns schema, validation, and canonical declarations.
- **Lock Reconciliation Component:** owns computed state materialization and staleness propagation.
- **Workflow Orchestrator Component:** owns chain selection, role dispatch, and lifecycle transitions.
- **Quality Gate Component:** owns deterministic validation execution and result normalization.
- **Context Assembly Component:** owns artifact selection, token budget policies, and summary hydration.
- **Governance Component:** owns ADR policy, approval checkpoints, and change control.

### D.5 Interfaces and Interactions

- Manifest -> Reconciler: declarative state input
- Reconciler -> Orchestrator: delta and impact set
- Orchestrator -> Agents: scoped task context and required outputs
- Agents -> Runtime: produced artifacts, logs, and structured outcomes
- Runtime -> Quality Gate: deterministic validation evidence
- Governance -> Orchestrator: acceptance, rejection, or revision directives

### D.6 Architectural Patterns Used

- Declarative state + reconciliation
- Dependency graph propagation
- Layered architecture with contract boundaries
- Evidence-driven quality gating
- Iterative lifecycle state machine

## E. Framework Structure

### E.1 Final Framework Model

- **Canonical artifacts:** `mind.toml`, `mind.lock`, ADR records, governance rules
- **Documentation zones:**  

  - `spec/`: stable intent and architecture definitions  
  - `state/`: current operational/project state  
  - `iterations/`: append-only iteration history  
  - `knowledge/`: reusable domain and framework references

- **Operational runtime:** `.mind/` for caches, outputs, logs, temporary and handoff state

### E.2 Key Modules / Domains / Workstreams

- Canonical modeling and schema governance
- Orchestration policy and workflow semantics
- Operational indexing, caching, and traceability
- Quality and validation architecture
- Knowledge and context management
- Integrations and adapter boundaries

### E.3 Evolution Across Iterations

- **Level 0 -> Level 3 maturity** progression remains the architecture evolution model **[Confirmed]**
- Every iteration must update:

  - canonical state deltas,
  - decision records (if architecture-impacting),
  - validation evidence and lifecycle outputs.

- Structural evolution is additive and compatibility-aware; no hidden schema breaks.

## F. Data Flows and Dependencies

### F.1 End-to-End Flow Design

1. Planning inputs update canonical intent (`mind.toml` + architecture docs).
2. Reconciler computes declared-vs-actual delta into `mind.lock`.
3. Orchestrator selects workflow and dispatches scoped agent chain.
4. Agents produce artifacts + runtime outputs; dependency graph marks impacts.
5. Quality gates validate deterministic checks and record evidence.
6. Governance checkpoint accepts/revises; iteration outputs are archived.
7. Feedback loop updates architecture assumptions, rules, and priorities.

### F.2 Upstream / Downstream Dependencies

- Upstream: requirements, constraints, architecture decisions, domain knowledge
- Internal: manifest schema <-> lock reconciliation <-> orchestration policies <-> gates
- Downstream: backlog design, implementation planning, execution strategy, review practices

### F.3 Coupling and Sequencing Considerations

- Keep coupling **schema-contract based**; avoid direct runtime-to-governance shortcuts.
- Sequence critical path as: canonical update -> reconciliation -> orchestration -> validation -> governance.
- Treat integration mechanisms (CLI, MCP, hooks) as replaceable adapters behind stable architecture contracts.

## G. Iteration Lifecycle

### G.1 Iteration Lifecycle Model

1. **Planning Input Stage**  

   Inputs: goals, constraints, prior iteration outputs, unresolved risks.

2. **Design/Review Stage**  

   Outputs: architecture updates, boundary decisions, declared scope.

3. **Reconciliation/Dispatch Stage**  

   Outputs: computed state delta, selected workflow chain, context bundles.

4. **Validation Checkpoint Stage**  

   Outputs: gate results, evidence logs, architectural compliance signal.

5. **Iteration Output Stage**  

   Outputs: archived artifacts, decisions, residual risks, next-iteration proposals.

6. **Feedback Stage**  

   Updates terminology, principles, and governance rules based on observed outcomes.

### G.2 Validation Checkpoints

- Checkpoint A: canonical consistency and dependency integrity
- Checkpoint B: workflow completion and artifact traceability
- Checkpoint C: deterministic quality gate evidence
- Checkpoint D: governance acceptance or required revision

## H. Governance and Decision Model

### H.1 Ownership Boundaries

- **Framework Owner:** architecture direction, principle stewardship, final arbitration.
- **Architecture Authority:** schema, boundaries, compatibility, ADR quality.
- **Workflow/Operations Owner:** orchestration policies and runtime standards.
- **Domain Contributors:** content integrity within assigned domain artifacts.

### H.2 Decision-Making Points

- Manifest schema or boundary changes -> architecture review required
- Cross-layer dependency additions -> governance checkpoint required
- New operational capabilities affecting canonical contracts -> ADR + validation requirement
- Iteration closure -> evidence review + decision log update required

### H.3 Architectural Change Control

- Mandatory ADR for architecture-impacting change
- Backward compatibility review for canonical contracts
- Explicit migration note for changed assumptions or terminology
- No architectural acceptance without traceable validation evidence

## I. Risks, Gaps, and Validation Needs

### I.1 Confirmed Decisions

- Canonical manifest/lock model and URI-based artifact identity
- Four-zone document architecture and orchestrated agent chains
- Dependency-aware reconciliation and deterministic gate concept
- Operational runtime layer for cache/log/output/session continuity

### I.2 Inferred Decisions (Needs Formal Confirmation)

- `[operations]` should be treated as a governed manifest section with defaults
- Runtime implementation mechanism should remain pluggable behind stable contracts
- Integration-first strategy should prioritize adapter compatibility over platform lock-in

### I.3 Open Questions / Validation Required

- Required vs optional status of operational manifest blocks
- Minimum contract for external integrations (CLI/MCP/hooks) at architecture level
- Default context budget policy per agent role
- Long-term retention and pruning policy for high-volume runtime artifacts

### I.4 Key Risks and Mitigations

- **Schema drift risk:** enforce schema validation + ADR approval path.
- **Boundary erosion risk:** require interface contract declaration for cross-layer changes.
- **Operational noise risk:** enforce output retention and summary policies.
- **Governance bypass risk:** iteration closure blocked without checkpoint evidence.
- **Integration volatility risk:** maintain adapter abstraction and canonical contract stability.

## J. Final Blueprint Summary

The final architecture is a **manifest-centered, reconciliation-driven framework** with clear layer boundaries, deterministic validation, and explicit governance. It is designed to scale through iterative formalization while preserving traceability and operational efficiency.

### Prioritized Architecture Focus Areas (Next Phase)

1. Freeze canonical contracts (manifest, lock, URI, gate result schema).
2. Formalize operational manifest boundary (`[operations]`) and ownership.
3. Finalize context budgeting and dependency propagation policies.
4. Harden lifecycle checkpoints and governance acceptance criteria.
5. Define adapter contract baseline for integration neutrality.

---

## Supporting Tables

### 1) Consensus / Divergence Matrix

| Topic | Alignment Status | Final Decision | Notes |
| --- | --- | --- | --- |
| Canonical manifest + lock model | High consensus | Adopt as core architecture | `mind.toml` intent + `mind.lock` resolved state |
| URI artifact identity | High consensus | Standardize `doc:{zone}/{name}` | Use for traceable references |
| 4-zone document architecture | High consensus | Adopt as mandatory structure | `spec/state/iterations/knowledge` |
| Reconciliation + dependency graph | High consensus | Keep as orchestration core | Delta-driven updates |
| Deterministic gates | High consensus | Mandatory before closure | Evidence-based review |
| Operational runtime (`.mind/`) | Medium-high consensus | Adopt as operational standard | Keep decoupled from canonical records |
| `[operations]` manifest block | Divergent | Treat as governed extension (pending final validation) | Promote toward standard |
| Integration mode (CLI/MCP/hooks) | Divergent | Define stable adapter contract, keep mechanism-agnostic | Avoid premature lock-in |

### 2) Component & Responsibility Matrix

| Component / Layer | Responsibility | Inputs | Outputs | Dependencies |
| --- | --- | --- | --- | --- |
| Manifest Authority | Canonical declarations, schema governance | Requirements, decisions | Valid manifest state | Governance rules |
| Lock Reconciler | Compute resolved state and drift | Manifest, runtime metadata | Lock state, impact set | Artifact registry, dependency graph |
| Workflow Orchestrator | Select/execute agent chain | Delta set, lifecycle state | Scoped tasks, lifecycle transitions | Agent registry, policy rules |
| Context Assembly | Build efficient task context | Artifact graph, budget rules | Prioritized context bundle | Cache/index subsystem |
| Quality Gate Engine | Deterministic validation evidence | Artifacts, gate policy | Gate results, pass/fail evidence | Runtime outputs |
| Governance Control | Decision checkpoints, change control | ADRs, evidence, risk status | Approval/revision decisions | All architecture layers |

### 3) Data Flow & Dependency Matrix

| Flow / Dependency | Source | Target | Purpose | Risk / Notes |
| --- | --- | --- | --- | --- |
| Intent declaration | Planning/design artifacts | Manifest | Declare desired state | Risk: ambiguous scope |
| State reconciliation | Manifest + filesystem/runtime metadata | Lock state | Compute actual state and drift | Risk: stale metadata |
| Workflow dispatch | Reconciler | Orchestrator/agents | Execute required changes | Risk: context underloading |
| Artifact production | Agents | Docs/runtime outputs | Materialize iteration outcomes | Risk: inconsistent structure |
| Validation evidence | Runtime outputs | Gate engine | Verify quality/compliance | Risk: nondeterministic commands |
| Governance decision | Gate evidence + ADRs | Architecture state | Accept/revise/freeze changes | Risk: checkpoint bypass |
| Feedback incorporation | Iteration review | Principles/policies | Improve next iteration | Risk: undocumented learnings |

### 4) Open Questions / Validation Table

| Item | Why unresolved | Impact | Validation needed |
| --- | --- | --- | --- |
| Operational block standardization | Proposal mismatch on mandatory schema | Medium | Architecture review + schema test cases |
| Adapter contract baseline | Different integration emphasis across proposals | High | Contract definition workshop + compatibility criteria |
| Default context budgets | Mentioned broadly, not normalized | Medium | Empirical thresholds by role/workflow type |
| Runtime retention policy | Inconsistent detail across documents | Medium | Operational load simulation and governance policy |
| Cross-layer change escalation thresholds | Guardrails implied but not codified | Medium | Explicit change-control matrix and examples |
