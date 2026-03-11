# Architecture

Phase 1 architecture for mind-cli Core CLI. Distilled from [BP-01: System Architecture](../blueprints/01-system-architecture.md) and validated against the scaffolded codebase. This document covers only what Phase 1 builds; extension points indicate where later phases connect.

## System Overview

mind-cli is a deterministic, single-binary CLI that provides project health diagnostics, documentation validation, document scaffolding, and workflow state inspection for the Mind Agent Framework. It replaces 9 bash scripts with a unified `mind` command.

The system follows a strict 4-layer architecture with downward-only dependency flow. The domain layer is pure Go (zero external imports). All filesystem access passes through repository interfaces, enabling in-memory implementations for testing.

Phase 1 delivers: `status`, `init`, `doctor`, `create` (6 artifact types), `docs` (5 subcommands), `check` (4 subcommands), `workflow` (2 subcommands), `version`, and `help` -- with `--json` support on every structured output command.

```
                     ┌─────────────────────────────────┐
                     │          User / CI / Agent       │
                     └──────────────┬──────────────────┘
                                    │
                     ┌──────────────v──────────────────┐
                     │      Presentation Layer          │
                     │  cmd/ (Cobra commands)           │
                     │  internal/render/ (formatting)   │
                     └──────────────┬──────────────────┘
                                    │ calls
                     ┌──────────────v──────────────────┐
                     │        Service Layer             │
                     │  internal/service/ (orchestrate) │
                     │  internal/validate/ (checks)     │
                     │  internal/generate/ (templates)  │
                     └──────────────┬──────────────────┘
                                    │ uses
                     ┌──────────────v──────────────────┐
                     │        Domain Layer              │
                     │  domain/ (types, enums, rules)   │
                     └──────────────┬──────────────────┘
                                    │ implemented by
                     ┌──────────────v──────────────────┐
                     │      Infrastructure Layer        │
                     │  internal/repo/interfaces.go     │
                     │  internal/repo/fs/ (filesystem)  │
                     │  internal/repo/mem/ (testing)    │
                     └─────────────────────────────────┘
```

## Component Map

### Phase 1 Packages

| Component | Package | Responsibility | Dependencies |
|-----------|---------|----------------|-------------|
| Root command | `cmd/root.go` | Global flags (`--json`, `--no-color`, `--project-root`), Cobra root initialization | `cobra` |
| Status command | `cmd/status.go` | Parse flags, assemble `ProjectHealth`, render output | `domain`, `internal/repo/fs`, `internal/render` |
| Init command | `cmd/init.go` | Create `.mind/`, `docs/`, `mind.toml`, adapter files | `domain`, `internal/generate`, `internal/repo/fs` |
| Doctor command | `cmd/doctor.go` | Run all diagnostic checks, produce remediation advice, auto-fix | `domain`, `internal/validate`, `internal/render`, `internal/repo/fs` |
| Create commands | `cmd/create_*.go` | Scaffold ADRs, blueprints, iterations, spikes, convergence, brief | `domain`, `internal/generate`, `internal/repo/fs` |
| Docs commands | `cmd/docs_*.go` | List, tree, stubs, search, open documents | `domain`, `internal/repo/fs`, `internal/render` |
| Check commands | `cmd/check_*.go` | Run validation suites (docs, refs, config, all) | `domain`, `internal/validate`, `internal/render` |
| Workflow commands | `cmd/workflow_*.go` | Display workflow state and iteration history | `domain`, `internal/repo/fs`, `internal/render` |
| Version command | `cmd/version.go` | Print build info (version, SHA, date, platform) | Go stdlib only |
| Helpers | `cmd/helpers.go` | `resolveRoot()` project detection delegation | `internal/repo/fs` |
| Domain types | `domain/` | Pure types: `Project`, `Config`, `Document`, `Zone`, `Brief`, `Iteration`, `WorkflowState`, `ValidationReport`, `CheckResult`, `ProjectHealth`, `Diagnostic` | Go stdlib only (NFR-4) |
| Domain logic | `domain/iteration.go` | `Slugify()`, `Classify()` -- pure deterministic functions | Go stdlib only |
| Domain errors | `domain/errors.go` | Sentinel errors: `ErrNotProject`, `ErrBriefMissing`, `ErrGateFailed`, `ErrCommandFailed` | Go stdlib only |
| Validation engine | `internal/validate/` | Check framework: `Suite`, `Check`, `CheckFunc`, `CheckContext`. Suite runners: `DocsSuite()`, `RefsSuite()`, `ConfigSuite()` | `domain`, `internal/repo` |
| Renderer | `internal/render/` | Output formatting: `Renderer`, `DetectMode()`, `TermWidth()`. Three modes: interactive (Lip Gloss), plain, JSON | `domain`, `golang.org/x/term` |
| Document generator | `internal/generate/` | Template rendering for ADRs, blueprints, iterations, spikes, convergence, briefs. Sequence number derivation. | `domain` |
| Repository interfaces | `internal/repo/interfaces.go` | `DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `BriefRepo` | `domain` |
| FS implementations | `internal/repo/fs/` | Real filesystem implementations: `DocRepo`, `IterationRepo`, `ConfigRepo`, `BriefRepo`, `FindProjectRoot()`, `DetectProject()` | `domain`, `go-toml/v2` |
| In-memory implementations | `internal/repo/mem/` | Test-only implementations backed by maps | `domain` |
| Project detection | `internal/repo/fs/project.go` | Walk-up `.mind/` detection, `DetectProject()` with config loading | `domain`, `internal/repo/fs` |
| Service: Project | `internal/service/project.go` | Orchestrate project detection, health assembly, config loading | `domain`, `internal/repo` |
| Service: Validation | `internal/service/validation.go` | Orchestrate doc/ref/config validation suites, unified reports | `domain`, `internal/validate`, `internal/repo` |
| Service: Generate | `internal/service/generate.go` | Orchestrate document scaffolding, sequence derivation, INDEX.md updates | `domain`, `internal/generate`, `internal/repo` |
| Service: Workflow | `internal/service/workflow.go` | Read workflow state, list iteration history | `domain`, `internal/repo` |

### Dependency Matrix

Arrows indicate "depends on" (import direction). Lower layers never import upper layers.

```
cmd/* ──────────────> internal/service/*
cmd/* ──────────────> internal/render
cmd/* ──────────────> domain
cmd/* ──────────────> internal/repo/fs (wiring only, in main.go ideally)

internal/service/* ─> domain
internal/service/* ─> internal/validate
internal/service/* ─> internal/generate
internal/service/* ─> internal/repo (interfaces)

internal/validate/ ─> domain
internal/validate/ ─> internal/repo (interfaces)

internal/generate/ ─> domain

internal/render/ ──-> domain

internal/repo/fs/ ──> domain
internal/repo/fs/ ──> go-toml/v2

internal/repo/mem/ ─> domain

domain/ ───────────-> (Go stdlib only)
```

## Layer Rules

### Layer 1: Presentation (`cmd/`, `internal/render/`)

**Responsibility**: Parse CLI flags, call service methods, format output.

**Rules**:
- Command handlers follow the pattern: resolve root -> create repos -> call service -> render result -> set exit code.
- No business logic. If you're computing whether a check passes, that belongs in `internal/validate/`.
- `init()` functions are permitted only for Cobra command registration (`rootCmd.AddCommand()`).
- The renderer receives domain types and produces strings. It never calls repositories or services.

**Phase 1 example** (actual scaffolded code in `cmd/status.go`):
```go
func runStatus(cmd *cobra.Command, args []string) error {
    root, err := resolveRoot()           // parse flags
    project, err := fs.DetectProject(root) // call infra
    // ... assemble health via repos ...
    r := render.New(mode, render.TermWidth())
    fmt.Print(r.RenderHealth(health))    // render
    return nil
}
```

### Layer 2: Service (`internal/service/`)

**Responsibility**: Orchestrate domain types with repository implementations. Business logic involving multiple entities or I/O coordination lives here.

**Rules**:
- Services receive repositories through constructor injection. No global state.
- Services operate on domain types and return domain types.
- Services never format output -- that is the renderer's job.
- Each service owns a focused responsibility: `ProjectService` (health), `ValidationService` (checks), `GenerateService` (scaffolding), `WorkflowService` (state).

**Phase 1 note**: The current scaffolded code wires repos directly in command handlers rather than through service constructors. Phase 1 implementation should migrate to proper service injection via `main.go`, per BP-01 Section 2.3 and C-10/C-11.

### Layer 3: Domain (`domain/`)

**Responsibility**: Define the nouns (types) and business rules. Pure Go, zero side effects.

**Rules**:
- Zero imports beyond Go stdlib. No `os`, `filepath`, `io`, or third-party packages. (NFR-4, DC-1)
- Enums are typed string constants (`Zone`, `DocStatus`, `BriefGate`, `RequestType`, `IterationStatus`, `CheckLevel`), never raw strings. (DC-3)
- Functions are pure: `Slugify()` and `Classify()` take inputs, return outputs, no side effects. (DC-4)
- Error sentinels are defined here (`ErrNotProject`, `ErrBriefMissing`) because they express domain concepts. Structured errors (`ErrGateFailed`, `ErrCommandFailed`) carry domain-relevant fields.
- JSON struct tags are defined on types that appear in `--json` output (`ProjectHealth`, `ValidationReport`, etc.) to ensure serialization contracts are domain-owned.

### Layer 4: Infrastructure (`internal/repo/`)

**Responsibility**: Implement repository interfaces with real I/O or in-memory storage.

**Rules**:
- Interfaces are defined in `internal/repo/interfaces.go`, consumed by services and validation checks.
- `internal/repo/fs/` contains real filesystem implementations. These are the only packages that call `os.ReadFile`, `os.Stat`, `filepath.Walk`, etc.
- `internal/repo/mem/` contains in-memory implementations for tests. Same interface, different storage.
- Repositories return domain types. They never return raw `[]byte` to callers that don't expect it (exception: `DocRepo.Read()` for content analysis).

## Data Model

The Phase 1 data model is fully specified in [docs/spec/domain-model.md](domain-model.md). Key relationships relevant to the architecture:

### Aggregate: Project

`Project` is the root aggregate. All operations begin by detecting a project (walking up for `.mind/`) and optionally loading its `Config` from `mind.toml`.

```
Project
 ├── Config (0..1, parsed from mind.toml)
 │    ├── Manifest (schema, generation, updated)
 │    ├── ProjectMeta (name, type, stack, commands)
 │    ├── Governance (max-retries, policies)
 │    ├── Profiles (active list)
 │    └── Documents (zone -> name -> DocEntry registry)
 ├── Documents (0..*, scanned from docs/ across 5 zones)
 ├── Brief (0..1, specialized Document with section analysis)
 ├── Iterations (0..*, scanned from docs/iterations/)
 │    └── Artifacts (exactly 5 expected per iteration)
 └── WorkflowState (0..1, parsed from docs/state/workflow.md)
```

### Computed Aggregates

These are assembled on-demand, not persisted:

- **ProjectHealth**: Assembled by `mind status`. Combines Project, Brief, ZoneHealth map, WorkflowState, last Iteration, warnings, suggestions.
- **ValidationReport**: Assembled by `mind check *`. Contains CheckResults with pass/fail/message per check.
- **Diagnostic**: Assembled by `mind doctor`. Each has level, message, fix suggestion, auto-fixability.

### State Machines

Three state machines operate in Phase 1. All are derived (computed from disk state), never persisted:

1. **Iteration Lifecycle**: CREATED -> IN_PROGRESS -> COMPLETE (derived from artifact presence scan)
2. **Brief Gate**: BRIEF_MISSING -> BRIEF_STUB -> BRIEF_PRESENT (derived from file existence + section analysis)
3. **Validation Check**: PENDING -> EXECUTED -> PASS/FAIL (stateless function execution)

WorkflowState is read-only in Phase 1 (state transitions are Phase 3).

## Key Decisions

### Decision: 4-Layer Architecture with Domain Purity

- **Choice**: Strict 4-layer architecture where the domain layer has zero external imports.
- **Rationale**: Domain purity enables testing business rules without filesystem mocks, ensures the domain model is portable, and enforces separation of concerns through compiler-checked import boundaries. Go's package system naturally enforces this -- if `domain/` imports `os`, the build fails the import constraint.
- **Rejected alternatives**:
  - **2-layer (handlers + models)**: Simpler but business logic migrates into handlers, making testing require full CLI setup. Rejected because validation engine complexity justifies a dedicated service layer.
  - **Hexagonal / ports-and-adapters with DI framework**: Correct conceptually but over-engineered for a CLI tool. Go's simplicity favors explicit constructor injection over framework-based DI. Rejected because the framework overhead exceeds the benefit for a single-binary CLI.
  - **Domain imports `filepath` for path manipulation**: Tempting because paths are central to the domain. Rejected because it creates a slippery slope -- once you import `filepath`, `os` follows. Path strings in domain types are opaque data; interpretation happens in infrastructure.
- **Consequences**: Makes domain testing trivial (no mocks needed). Makes adding new infrastructure implementations easy (e.g., in-memory for tests). Adds a thin translation layer where repos convert filesystem state to domain types.

### Decision: Repository Interfaces Defined at Consumer Site

- **Choice**: Repository interfaces are defined in `internal/repo/interfaces.go`, close to where they are consumed. Implementations live in sub-packages (`fs/`, `mem/`).
- **Rationale**: Go idiom says "define interfaces where they are consumed, not where they are implemented." Keeps interfaces small and focused on what consumers actually need. Avoids "god interface" anti-pattern where one interface has 20 methods.
- **Rejected alternatives**:
  - **Interfaces in domain/**: Would force the domain package to know about I/O concepts. Rejected per DC-1 (domain purity).
  - **Interfaces per-consumer (each service defines its own)**: Would create interface duplication and fragmentation. Rejected because Phase 1's five repos are small enough to share a single interfaces file. Revisit if interface count exceeds 10.
  - **No interfaces, concrete types everywhere**: Would make testing require real filesystem. Rejected because validation logic has 28+ checks that need fast, deterministic test execution.
- **Consequences**: Tests use `internal/repo/mem/` implementations. Adding a new data source (e.g., git-based iteration detection) means implementing the existing interface without changing consumers.

### Decision: Validation Engine as Check Framework

- **Choice**: Validation uses a `Suite` containing ordered `Check` items, each with an ID, name, level, and `CheckFunc`. Suites are stateless: execute all checks, collect results into a `ValidationReport`.
- **Rationale**: The bash scripts (`validate-docs.sh`, `validate-xrefs.sh`) use numbered checks that produce pass/fail output. The Go implementation must produce identical check IDs and pass/fail results (NFR-6). A check framework makes this deterministic, testable, and composable.
- **Rejected alternatives**:
  - **Ad-hoc validation functions**: Each command runs its own checks inline. Rejected because it prevents `mind check all` from aggregating results and prevents `mind doctor` from reusing check logic.
  - **Declarative schema validation (e.g., JSON Schema for mind.toml)**: Only works for config validation. Document structure checks and cross-reference checks require procedural logic. Rejected as insufficient for the full validation surface.
  - **Parallel check execution**: Checks run concurrently for speed. Rejected for Phase 1 because check order matters for early-exit scenarios (if `docs/` doesn't exist, later checks are meaningless) and because sequential execution is fast enough for 10-50 documents (NFR-1).
- **Consequences**: Adding a new check = defining a `CheckFunc` and adding it to a suite. `--strict` modifies level interpretation without touching check logic. Parity with bash scripts is verifiable by comparing check IDs and results.

### Decision: Output Modes via Renderer Pattern

- **Choice**: A `Renderer` type with three modes (interactive, plain, JSON) selected by `DetectMode()`. Each render method (`RenderHealth`, `RenderValidation`, etc.) dispatches to mode-specific formatting.
- **Rationale**: Every command must support `--json` (FR-7). Terminal detection (TTY vs pipe) must auto-select between interactive and plain modes (FR-6). Centralizing this in a renderer prevents scattered `if flagJSON` checks in every command handler.
- **Rejected alternatives**:
  - **Each command handles its own formatting**: Leads to inconsistent output, duplicated JSON marshaling code, and format-specific bugs. Rejected for maintainability.
  - **Template-based rendering**: Use Go templates for text output. Rejected because the formatting logic (progress bars, box drawing, column alignment) is imperative, not template-friendly. Templates add complexity without reducing code.
  - **Separate formatter per command**: `StatusFormatter`, `CheckFormatter`, etc. Viable but Phase 1 has only 4-5 output shapes. Premature to split until there are 10+. Rejected for now; revisit in Phase 2 when TUI rendering adds complexity.
- **Consequences**: Adding a new output shape = adding a `Render*` method to `Renderer`. Lip Gloss styling is isolated to interactive mode methods. JSON mode always uses `json.MarshalIndent` on domain types, so JSON contracts are owned by struct tags.

### Decision: Constructor Injection in main.go

- **Choice**: All dependency wiring happens explicitly in `main.go`. Services receive their repository implementations through constructors. No `init()` functions for wiring. No global mutable state.
- **Rationale**: Makes the dependency graph visible in one place. No hidden initialization order. Easy to swap implementations for testing. Follows the constraint C-10 (no init() except Cobra registration) and C-11 (no global mutable state).
- **Rejected alternatives**:
  - **DI framework (wire, dig, fx)**: Adds a code generation or reflection step. Overkill for a CLI with ~5 services and ~5 repos. Rejected for simplicity.
  - **Package-level singletons**: `var defaultDocRepo = fs.NewDocRepo(...)`. Breaks testability, creates hidden coupling, violates C-11. Rejected.
  - **Service locator pattern**: Pass a `Container` that services query for dependencies. Hides the dependency graph, makes refactoring risky. Rejected.
- **Consequences**: `main.go` grows as services are added. This is acceptable -- a 50-line `main.go` is readable and debuggable. The current scaffolded code wires repos in command handlers; Phase 1 implementation should centralize this.

### Decision: Exit Code Strategy

- **Choice**: Four exit codes: 0 (success), 1 (validation failure / issues found), 2 (runtime error), 3 (configuration error / not a Mind project). Deterministic mapping per FR-49.
- **Rationale**: CI pipelines and scripts depend on exit codes. A consistent, documented scheme prevents misinterpretation. The bash scripts use similar codes, enabling parity testing.
- **Rejected alternatives**:
  - **Unix convention (0 = success, 1 = everything else)**: Loses information. A CI pipeline can't distinguish "checks failed" from "crash" from "wrong directory". Rejected.
  - **Granular codes (one per error type)**: Too many codes to document and maintain. Diminishing returns after 4-5 distinct categories. Rejected.
  - **Exit code 2 for "already initialized" on init**: The current scaffolded spec uses exit 2 for `mind init` when `.mind/` exists (FR-19). This reuses the "runtime error" code. Acceptable because it is a recoverable operational error, not a validation failure.
- **Consequences**: Every command handler must map errors to the correct exit code. Helper functions in `cmd/` translate domain errors (`ErrNotProject` -> exit 3) to exit codes.

## Boundaries

### In Scope (Phase 1)

- Domain types: `Project`, `Config`, `Document`, `Zone`, `DocStatus`, `Brief`, `BriefGate`, `Iteration`, `WorkflowState`, `ValidationReport`, `CheckResult`, `ProjectHealth`, `Diagnostic`, and all supporting types
- Repository layer: `DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `BriefRepo` interfaces + filesystem implementations + in-memory test implementations
- Service layer: `ProjectService`, `ValidationService`, `GenerateService`, `WorkflowService`
- Validation engine: 17-check doc suite, 11-check ref suite, config validation suite
- Rendering: Interactive (Lip Gloss), Plain, JSON output modes
- Document generation: ADR, blueprint, iteration, spike, convergence, brief templates
- Commands: `status`, `init`, `doctor`, `create` (6 sub-commands), `docs` (5 sub-commands), `check` (4 sub-commands), `workflow` (2 sub-commands), `version`, `help`
- Cross-cutting: `--json` flag, `--no-color` flag, `--project-root` flag, project root auto-detection, `mind.toml` parsing, exit codes

### Out of Scope (Phase 1.5+)

- **Phase 1.5 -- Reconciliation**: `mind.lock`, SHA-256 hash tracking, dependency graph evaluation, staleness propagation, `mind reconcile` command. Extension point: `ReconcileService` will consume `DocRepo` + `ConfigRepo` and produce a `LockFile` domain type.
- **Phase 2 -- TUI**: `mind tui` with 5-tab Bubble Tea interface. Extension point: TUI components will consume the same services as `cmd/` handlers, rendering to Bubble Tea models instead of strings.
- **Phase 3 -- AI Bridge (A+B)**: `mind preflight`, `mind handoff`, `mind serve` (MCP server). Extension point: `WorkflowService` will gain write capabilities; `MCP` package will expose service methods as JSON-RPC tools.
- **Phase 4 -- AI Bridge (C+D)**: `mind watch`, `mind run`. Extension point: orchestration service will compose `WorkflowService` + `ValidationService` + external process dispatch.
- **Phase 5 -- Polish**: Shell completions, GoReleaser configuration, CI/CD pipeline.

### Extension Points

| Extension Point | Where | What Plugs In |
|----------------|-------|---------------|
| New validation suite | `internal/validate/` | Add a new `*Suite()` function (e.g., `ReconcileSuite()`) following the existing `DocsSuite()` pattern |
| New repository | `internal/repo/interfaces.go` + `fs/` + `mem/` | Add interface + two implementations (e.g., `LockRepo` for Phase 1.5) |
| New service | `internal/service/` | Add service struct with constructor injection (e.g., `ReconcileService` for Phase 1.5) |
| New command | `cmd/` | Add Cobra command file, wire to service in init(), follow thin-handler pattern |
| New render shape | `internal/render/` | Add `Render*()` method for new domain output type |
| New output mode | `internal/render/` | Extend `OutputMode` enum (e.g., for TUI rendering in Phase 2) |
| MCP tool | `mcp/` (Phase 3) | New package consuming same services as `cmd/`, exposing via JSON-RPC |

---

## Phase 1.5: Reconciliation Engine

Phase 1.5 adds hash-based content tracking with staleness propagation through a dependency graph. This section documents the architectural extensions. For full design rationale, see [architecture-delta.md](../iterations/002-reconciliation-engine/architecture-delta.md).

### Phase 1.5 Packages

| Component | Package | Responsibility | Dependencies |
|-----------|---------|----------------|-------------|
| Reconcile domain types | `domain/reconcile.go` | `LockFile`, `LockEntry`, `ReconcileResult`, `GraphEdge`, `EdgeType`, `Graph`, `LockStatus`, `EntryStatus`, `LockStats`, `ReconcileOpts`, `StalenessInfo`, `BuildGraph()` | Go stdlib only |
| Hash computation | `internal/reconcile/hash.go` | SHA-256 of raw file bytes, mtime fast-path, edge case handling (empty, binary, symlinks, large, unreadable) | `domain`, `os`, `crypto/sha256`, `io` |
| Graph operations | `internal/reconcile/graph.go` | Cycle detection (DFS), edge validation against document registry | `domain` |
| Staleness propagation | `internal/reconcile/propagate.go` | BFS downstream propagation with depth limit 10, edge-type-specific reason messages | `domain` |
| Engine orchestration | `internal/reconcile/engine.go` | 6-phase reconciliation: load, graph, scan, detect undeclared, propagate, report | `domain`, `internal/reconcile/*`, `internal/repo` (DocRepo) |
| Lock repository (fs) | `internal/repo/fs/lock_repo.go` | Read/write `mind.lock` as JSON with atomic writes (temp + rename) | `domain`, `os`, `encoding/json` |
| Lock repository (mem) | `internal/repo/mem/lock_repo.go` | In-memory `LockRepo` for testing | `domain` |
| Reconciliation service | `internal/service/reconciliation.go` | `ReconciliationService`: orchestrate config loading, lock I/O, engine execution | `domain`, `internal/reconcile`, `internal/repo` |
| Reconcile suite | `internal/validate/reconcile.go` | `ReconcileSuite()`: project reconciliation results into check framework for `mind check all` | `domain`, `internal/validate` |
| Reconcile command | `cmd/reconcile.go` | `mind reconcile` with `--check`, `--force`, `--graph` flags | `domain`, `internal/service`, `internal/render` |

### Updated Dependency Matrix (Phase 1.5 additions)

```
cmd/reconcile.go ──────> internal/service/reconciliation
cmd/reconcile.go ──────> internal/render
cmd/reconcile.go ──────> domain

cmd/status.go ─────────> internal/service/reconciliation (ReadStaleness)
cmd/check.go ──────────> internal/service/validation (extended with reconcile)
cmd/doctor.go ─────────> internal/service/doctor (extended with staleness)

internal/service/reconciliation ──> internal/reconcile/engine
internal/service/reconciliation ──> internal/repo (ConfigRepo, DocRepo, LockRepo)
internal/service/reconciliation ──> domain

internal/service/validation ──────> internal/validate/reconcile (ReconcileSuite)

internal/reconcile/engine ────────> internal/reconcile/hash
internal/reconcile/engine ────────> internal/reconcile/graph
internal/reconcile/engine ────────> internal/reconcile/propagate
internal/reconcile/engine ────────> internal/repo (DocRepo interface)
internal/reconcile/engine ────────> domain

internal/reconcile/hash ──────────> domain
internal/reconcile/hash ──────────> os, crypto/sha256, io (direct I/O)

internal/reconcile/graph ─────────> domain
internal/reconcile/propagate ─────> domain

internal/validate/reconcile ──────> domain

internal/repo/fs/lock_repo ──────> domain
internal/repo/mem/lock_repo ─────> domain

domain/reconcile.go ──────────────> Go stdlib only
```

### Extension Points Activated by Phase 1.5

| Extension Point | How It Was Used |
|----------------|----------------|
| New validation suite | `ReconcileSuite()` added in `internal/validate/reconcile.go`, following the `DocsSuite()` pattern |
| New repository | `LockRepo` interface added to `interfaces.go`, with `fs/lock_repo.go` and `mem/lock_repo.go` implementations |
| New service | `ReconciliationService` added in `internal/service/reconciliation.go` with constructor injection |
| New command | `cmd/reconcile.go` added with thin-handler pattern, wired to `ReconciliationService` |
| New render shape | `RenderReconcileResult()` and `RenderGraph()` added to `Renderer` |

### Phase 1.5 Key Decisions

#### Decision: hash.go Performs Direct Filesystem I/O

- **Choice**: `internal/reconcile/hash.go` uses `os.Open()` directly to stream file content through SHA-256. It does not use `DocRepo.Read()`.
- **Rationale**: Hash computation requires streaming I/O (constant memory) rather than loading entire file contents into memory. The existing `DocRepo.Read()` returns `[]byte`, which would require loading the full file. For files up to 10MB, this is wasteful. Streaming uses O(1) memory regardless of file size.
- **Rejected alternatives**:
  - **Add `DocRepo.OpenReader()` streaming interface**: Over-engineering for a single consumer. Rejected.
  - **Use `DocRepo.Read()` then hash the bytes**: Defeats streaming purpose, wastes memory. Rejected.
- **Consequences**: `internal/reconcile/hash.go` imports `os` and `io`. This is acceptable because `internal/reconcile/` is a service-layer package, not the domain layer. The hash function accepts an absolute path as a parameter, making it a testable utility.

#### Decision: Exit Code 4 for Staleness Detection

- **Choice**: Add exit code 4 to represent staleness, used exclusively by `mind reconcile --check`.
- **Rationale**: Staleness is semantically distinct from validation failure (exit 1), runtime error (exit 2), and configuration error (exit 3). CI pipelines need to differentiate "documentation has validation errors" from "documentation is out of date relative to dependencies." The existing exit code 1 means "something failed structurally"; exit code 4 means "everything is structurally sound but temporally stale."
- **Rejected alternatives**:
  - **Reuse exit code 1**: Loses the semantic distinction between "broken" and "stale." CI pipelines cannot differentiate. Rejected.
  - **Use exit code 5+**: No benefit over 4. The next available code is 4. Rejected for simplicity.
- **Consequences**: Exit code documentation is extended. The exit code helper in `cmd/` must handle the new code. Only `mind reconcile --check` returns exit 4; no other command uses it.

#### Decision: Centralized Wiring via PersistentPreRunE

- **Choice**: Centralize repository and service construction in `rootCmd.PersistentPreRunE` during Phase 1.5.
- **Rationale**: Phase 1.5 adds 4 integration points (reconcile, status, check, doctor) that all need `LockRepo` and `ReconciliationService`. The current Phase 1 pattern of creating repos in each command handler would duplicate wiring in 4+ places. Centralizing during Phase 1.5 is the cheapest point because those files are already being modified.
- **Rejected alternatives**:
  - **Keep per-handler wiring**: More duplication, harder to maintain. Rejected.
  - **Fix wiring before Phase 1.5**: Separate refactoring iteration touching the same files. Merge conflict risk. Rejected.
  - **Defer past Phase 1.5**: Every new integration point repeats the anti-pattern. Rejected.
- **Consequences**: `cmd/root.go` gains ~40 lines of wiring code. All existing command handlers are simplified by removing inline repo construction. Commands that do not require a project use a guard annotation to skip wiring.

#### Decision: ReconcileSuite Takes Pre-computed Result

- **Choice**: `ReconcileSuite(result *ReconcileResult)` takes a pre-computed result rather than running reconciliation internally.
- **Rationale**: Reconciliation is expensive (file I/O, hashing). The result is needed by the command handler (for exit code determination), the renderer (for output), and the suite (for check projection). Computing it once and sharing is more efficient than running it inside the suite.
- **Rejected alternatives**:
  - **Suite runs reconciliation internally**: Would cause redundant I/O. The command handler also needs the result. Rejected.
  - **Dedicated reconciliation panel (not suite)**: Would break the established check framework pattern and require special-case rendering. Rejected.
- **Consequences**: `ReconcileSuite()` has a different signature from other suite constructors. `ValidationService.RunAll()` must receive or compute the reconcile result before calling the suite.
