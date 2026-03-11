# BP-01: System Architecture

> How is the system structured? What are the rules?

**Status**: Active
**Date**: 2026-03-11
**Cross-references**: [BP-02](03-architecture.md) for entity details, [BP-04](01-mind-cli.md) for command specifications

---

## 1. Design Principles

### P1: No AI in the CLI

The `mind` binary is strictly deterministic. Every command produces the same output given the same input. AI workflows (`/workflow`, `/discover`, `/analyze`) remain in Claude Code and agent contexts. The CLI manages everything *around* AI workflows: validation, scaffolding, state management, context assembly.

The sole exception is `mind run`, which *dispatches* agents via the `claude` CLI as an external process. The mind binary itself never calls AI APIs or makes non-deterministic decisions.

### P2: Single Binary, Zero-Config Defaults

The tool ships as one binary with no runtime dependencies. Running `mind status` in any directory containing `.mind/` works immediately. No configuration files, environment variables, or setup steps are required for basic operation. `mind.toml` adds project-specific tuning but is never mandatory.

### P3: Progressive Disclosure

Simple first, powerful later. A new user types `mind status` and gets an overview. They type `mind doctor` and get actionable diagnostics. Flags like `--strict`, `--json`, `--verbose`, and `--fix` unlock advanced behavior without cluttering the default experience. The TUI (`mind tui`) reveals the full depth for users who want to explore interactively.

### P4: Verb-Noun Command Structure

Commands follow `mind <verb> [noun]` consistently:

```
mind status                  # verb only (noun implied: project)
mind check docs              # verb + noun
mind create blueprint "..."  # verb + noun + argument
mind workflow clean --dry-run # verb + noun + flag
```

This maps naturally to how developers think about actions. Cobra enforces this structure and generates help text, completions, and man pages automatically.

### P5: Script Compatibility

The CLI must produce validation results identical to the existing bash scripts. If `validate-docs.sh` reports 15/17 pass, `mind check docs` must report 15/17 pass with the same check IDs and failure messages. This is not "roughly equivalent" --- it is exact behavioral parity verified by integration tests that run both implementations side by side.

### P6: Agent-Friendly Output

Every command that produces structured results supports `--json` for machine consumption. AI agents using MCP or processing CLI output can parse JSON reliably. The JSON schema is stable across patch versions. Breaking changes to JSON output require a major version bump.

---

## 2. Architectural Layers

The system uses a strict 4-layer architecture. Dependencies flow downward only. No layer may import from a layer above it.

```
+-------------------------------------------------------------------------+
|                          PRESENTATION LAYER                              |
|                                                                         |
|   +-------------+  +-------------+  +-------------+  +---------------+  |
|   | CLI (Cobra) |  | TUI (Tea)   |  | MCP Server  |  | JSON Output   |  |
|   | cmd/*.go    |  | tui/*.go    |  | mcp/*.go    |  | (--json flag) |  |
|   +------+------+  +------+------+  +------+------+  +-------+-------+  |
|          |                |                |                  |          |
+----------+----------------+----------------+------------------+----------+
           |                |                |                  |
           +--------+-------+--------+-------+------------------+
                    |                |
+-------------------+----------------+---------------------------------+
|                   v                v        SERVICE LAYER             |
|                                                                      |
|   +------------------+  +-------------------+  +------------------+  |
|   | ProjectService   |  | ValidationService |  | GenerateService  |  |
|   | (status, doctor) |  | (docs, refs, gate)|  | (ADR, blueprint) |  |
|   +------------------+  +-------------------+  +------------------+  |
|                                                                      |
|   +------------------+  +-------------------+  +------------------+  |
|   | WorkflowService  |  | QualityService    |  | SyncService      |  |
|   | (state, history) |  | (log, trends)     |  | (agents)         |  |
|   +------------------+  +-------------------+  +------------------+  |
|                                                                      |
|   +------------------+  +-------------------+                        |
|   | OrchestrationSvc |  | ReconcileSvc      |                        |
|   | (preflight, run) |  | (hash, lock)      |                        |
|   +--------+---------+  +--------+----------+                        |
|            |                     |                                    |
+------------+---------------------+------------------------------------+
             |                     |
             +----------+----------+
                        |
+-------------------+---+----------------------------------------------+
|                   v          DOMAIN LAYER                             |
|                                                                      |
|   +-----------+  +-----------+  +------------+  +-----------+        |
|   | Project   |  | Document  |  | Iteration  |  | Workflow  |        |
|   | Config    |  | Zone      |  | Artifact   |  | State     |        |
|   | Brief     |  | DocStatus |  | ReqType    |  | Chain     |        |
|   +-----------+  +-----------+  +------------+  +-----------+        |
|                                                                      |
|   +-----------+  +----------------+  +-----------+  +------------+   |
|   | Quality   |  | Validation     |  | AgentChain|  | LockFile   |   |
|   | Score     |  | Report, Check  |  | AgentRef  |  | DepEdge    |   |
|   +-----------+  +----------------+  +-----------+  +------------+   |
|                                                                      |
+-------------------+--------------------------------------------------+
                    |
+-------------------+--------------------------------------------------+
|                   v          INFRASTRUCTURE LAYER                     |
|                                                                      |
|   +--------------+  +-----------+  +------------+  +---------------+ |
|   | Filesystem   |  | Git       |  | Process    |  | Template      | |
|   | (repo/fs)    |  | (branch,  |  | (build,    |  | (load,render) | |
|   |              |  |  status)  |  |  test,lint) |  |               | |
|   +--------------+  +-----------+  +------------+  +---------------+ |
|                                                                      |
|   +--------------+  +-------------------+                            |
|   | fsnotify     |  | MCP Transport     |                            |
|   | (watch)      |  | (stdio JSON-RPC)  |                            |
|   +--------------+  +-------------------+                            |
+----------------------------------------------------------------------+
```

### 2.1 Presentation Layer

**Responsibility**: Accept user input, call a service method, render the result. Zero business logic.

The presentation layer has four surfaces:

- **CLI (Cobra)** --- `cmd/*.go` files. Each file registers one top-level command with its subcommands and flags. The `RunE` function extracts flags, resolves the service from context, calls it, and passes the result to a renderer. A command file should rarely exceed 60 lines.

- **TUI (Bubble Tea)** --- `tui/*.go` files. A full-screen interactive application following the Elm architecture (Model-View-Update). Each tab is its own Bubble Tea model. The top-level model delegates input to the active tab. Services are injected at construction time.

- **MCP Server** --- `internal/mcp/*.go`. Implements the Model Context Protocol over stdio using JSON-RPC 2.0. Exposes 16 tools that map directly to service methods. The MCP handler parses the JSON-RPC request, calls the corresponding service, and serializes the response. No business logic in the handler.

- **JSON Output** --- The `--json` flag on any command switches the renderer from interactive/plain to JSON. This is not a separate surface but a rendering mode. See Section 7.

**Layer rule**: Presentation may import Service and Domain. It must not import Infrastructure directly.

### 2.2 Service Layer

**Responsibility**: Orchestrate business logic by composing domain types and infrastructure repositories.

Each service is defined as an interface in `internal/service/` with one implementation. Services receive repository interfaces via constructor injection.

| Service | Responsibilities |
|---------|-----------------|
| **ProjectService** | Project detection, `status` health aggregation, `doctor` diagnostics, `init` scaffolding |
| **ValidationService** | Run validation suites (docs 17 checks, refs 11 checks, config, convergence 23 checks), deterministic gate, compose unified reports |
| **GenerateService** | Create ADR, blueprint, iteration, spike, convergence, brief from templates. Auto-number sequences. Update INDEX.md |
| **WorkflowService** | Read/write workflow state, list iteration history, clean stale state |
| **QualityService** | Extract scores from convergence files, append to quality log, compute trends |
| **SyncService** | Synchronize `.mind/conversation/agents/` to `.github/agents/`, detect drift |
| **OrchestrationService** | Pre-flight checks, full `mind run` pipeline (classify, gate, create iteration, dispatch agents, run gates, retry), handoff context assembly |
| **ReconciliationService** | Content hash tracking, lock file management, staleness detection across documents and their dependencies |

**Layer rule**: Services may import Domain and Infrastructure (via interfaces). They must not import Presentation. Services must not import other services directly --- shared logic belongs in the domain layer or is composed at the wiring level.

### 2.3 Domain Layer

**Responsibility**: Define the vocabulary of the system. Pure types, value objects, enumerations, constants, and business rules.

The domain layer has **zero imports beyond the standard library**. No third-party packages, no framework types, no IO. This makes domain types trivially testable and ensures the core concepts are decoupled from implementation details.

**Core types** (see BP-02 for complete field definitions):

| Type | Role |
|------|------|
| `Project` | Root aggregate. Represents a detected Mind Framework project |
| `Config` | Parsed `mind.toml` manifest |
| `Document` | A single documentation file with zone, status, stub detection |
| `Zone` | Enum: spec, blueprints, state, iterations, knowledge |
| `Iteration` | A per-change tracking folder with typed artifacts |
| `WorkflowState` | Persisted state of an in-progress AI workflow |
| `ValidationReport` | Aggregated results from a validation suite |
| `CheckResult` | Outcome of a single validation check (pass/fail, level, message) |
| `QualityScore` | Convergence analysis quality assessment (6 dimensions, overall score, gate pass/fail) |
| `Brief` | Parsed project brief with section detection and gate classification |
| `AgentChain` | Ordered sequence of agents for a given request type |
| `LockFile` | Content hashes for reconciliation tracking |
| `DependencyEdge` | Directed edge between documents for staleness propagation |

**Business rules encoded in the domain layer**:

- A project must contain `.mind/` to be valid.
- Documents with only headings, comments, and placeholders are classified as stubs.
- The business context gate blocks `NEW_PROJECT` and `COMPLEX_NEW` when the brief is missing or a stub.
- Quality gate 0 requires a convergence score >= 3.0/5.0.
- Maximum 2 retry loops per workflow (configurable via `mind.toml` governance).

**Layer rule**: Domain imports nothing outside `stdlib`. All other layers may import Domain.

### 2.4 Infrastructure Layer

**Responsibility**: Interact with the outside world. Filesystem, git, processes, network, templates.

All external interactions are abstracted behind repository interfaces defined in `internal/repo/`. The infrastructure layer provides concrete implementations.

| Component | Package | What it does |
|-----------|---------|--------------|
| **Filesystem** | `internal/repo/fs/` | Read/write files, list directories, detect stubs, search content. Implements `DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `QualityRepo`, `TemplateRepo` |
| **In-Memory** | `internal/repo/mem/` | In-memory implementations of all repository interfaces for unit and integration testing |
| **Git** | `internal/repo/fs/` | Branch creation, status detection, current branch name. Wraps `git` CLI |
| **Process** | `internal/repo/fs/` | Execute build, test, lint commands declared in `mind.toml [project.commands]`. Capture stdout/stderr and exit codes |
| **Template** | `internal/repo/fs/` | Load `.mind/docs/templates/*.md`, apply variable substitution (`{{.Title}}`, `{{.Date}}`, `{{.Seq}}`) |
| **fsnotify** | `internal/watch/` | Filesystem event monitoring for watch mode. Wraps `github.com/fsnotify/fsnotify` |
| **MCP Transport** | `internal/mcp/` | stdio-based JSON-RPC 2.0 transport for the MCP server |

**Layer rule**: Infrastructure may import Domain (for types it returns). It must not import Service or Presentation. Infrastructure implementations are always accessed through interfaces.

---

## 3. Component Map

### Go Package Structure

```
mind-cli/
|-- main.go                        Entry point, dependency wiring (Deps struct)
|-- cmd/                           Cobra commands (presentation)
|   |-- root.go                    Root command, global flags (--json, --no-color, --verbose)
|   |-- status.go                  mind status
|   |-- doctor.go                  mind doctor [--fix] [--json]
|   |-- init.go                    mind init [--name] [--with-github] [--from-existing]
|   |-- create.go                  mind create {adr,blueprint,iteration,spike,convergence,brief}
|   |-- docs.go                    mind docs {list,tree,open,stubs,search}
|   |-- check.go                   mind check {docs,refs,config,convergence,all}
|   |-- workflow.go                mind workflow {status,history,show,clean}
|   |-- sync.go                    mind sync {agents,status}
|   |-- quality.go                 mind quality {log,history,report}
|   |-- serve.go                   mind serve (MCP server)
|   |-- preflight.go               mind preflight "<request>"
|   |-- run.go                     mind run "<request>" [--dry-run]
|   |-- watch.go                   mind watch
|   |-- tui_cmd.go                 mind tui (launches Bubble Tea app)
|   |-- completion.go              mind completion {bash,zsh,fish}
|   +-- version.go                 mind version [--short]
|
|-- tui/                           Bubble Tea TUI application (presentation)
|   |-- app.go                     Top-level model, tab switching, service injection
|   |-- status.go                  Tab 1: project health dashboard
|   |-- docs.go                    Tab 2: document browser with zone filtering
|   |-- iterations.go              Tab 3: iteration timeline with type filtering
|   |-- checks.go                  Tab 4: live validation results
|   |-- quality.go                 Tab 5: convergence quality trends
|   |-- styles.go                  Lip Gloss style definitions (colors, borders, spacing)
|   +-- keys.go                    Key binding definitions
|
|-- domain/                        Entities, value objects, enums (domain)
|   |-- project.go                 Project, Config, Manifest, ProjectMeta, StackConfig
|   |-- document.go                Document, Zone, DocStatus, Brief, BriefGate
|   |-- iteration.go               Iteration, Artifact, RequestType, IterationStatus
|   |-- workflow.go                WorkflowState, DispatchEntry, CompletedArtifact
|   |-- validation.go              ValidationReport, CheckResult, CheckLevel
|   |-- quality.go                 QualityScore, QualityEntry
|   |-- health.go                  ProjectHealth, ZoneHealth
|   |-- agent.go                   AgentChain, AgentRef, Chains()
|   |-- reconcile.go               LockFile, DependencyEdge, ContentHash
|   +-- errors.go                  Sentinel and typed errors
|
|-- internal/
|   |-- service/                   Service interfaces and implementations
|   |   |-- interfaces.go          All service interfaces in one file
|   |   |-- project.go             ProjectService implementation
|   |   |-- validation.go          ValidationService implementation
|   |   |-- generate.go            GenerateService implementation
|   |   |-- workflow.go            WorkflowService implementation
|   |   |-- quality.go             QualityService implementation
|   |   |-- sync.go                SyncService implementation
|   |   |-- orchestration.go       OrchestrationService implementation
|   |   +-- reconciliation.go      ReconciliationService implementation
|   |
|   |-- repo/                      Repository interfaces
|   |   |-- interfaces.go          DocRepo, IterationRepo, StateRepo, ConfigRepo, etc.
|   |   |-- fs/                    Filesystem implementations
|   |   |   |-- doc_repo.go        FSDocRepo
|   |   |   |-- iteration_repo.go  FSIterationRepo
|   |   |   |-- state_repo.go      FSStateRepo
|   |   |   |-- config_repo.go     FSConfigRepo
|   |   |   |-- quality_repo.go    FSQualityRepo
|   |   |   |-- template_repo.go   FSTemplateRepo
|   |   |   |-- git.go             Git operations (branch, status)
|   |   |   +-- process.go         Command execution (build, test, lint)
|   |   +-- mem/                   In-memory implementations (testing)
|   |       |-- doc_repo.go        MemDocRepo
|   |       |-- iteration_repo.go  MemIterationRepo
|   |       +-- state_repo.go      MemStateRepo
|   |
|   |-- render/                    Output formatting (presentation support)
|   |   |-- render.go              Renderer interface, OutputMode, DetectMode()
|   |   |-- interactive.go         InteractiveRenderer (Lip Gloss)
|   |   |-- plain.go               PlainRenderer (no ANSI)
|   |   +-- json.go                JSONRenderer (structured output)
|   |
|   |-- validate/                  Validation suites
|   |   |-- check.go               Check, CheckFunc, Suite, Suite.Run()
|   |   |-- docs.go                17-check documentation suite
|   |   |-- refs.go                11-check cross-reference suite
|   |   |-- config.go              Conversation config validation
|   |   |-- convergence.go         23-check convergence output validation
|   |   +-- gate.go                Deterministic gate (build + test + lint)
|   |
|   |-- reconcile/                 Reconciliation engine
|   |   |-- hash.go                Content hashing (SHA-256 of normalized content)
|   |   |-- lock.go                Lock file read/write (.mind/lock.json)
|   |   +-- staleness.go           Dependency graph traversal, staleness propagation
|   |
|   |-- generate/                  Document generation
|   |   |-- template.go            Template loading and variable substitution
|   |   |-- sequence.go            Auto-numbering (scan existing, return next)
|   |   +-- slugify.go             Title to kebab-case slug conversion
|   |
|   |-- mcp/                       MCP server protocol
|   |   |-- server.go              JSON-RPC 2.0 handler, tool registry
|   |   |-- tools.go               16 tool definitions (name, schema, handler)
|   |   +-- transport.go           stdio reader/writer
|   |
|   |-- watch/                     Filesystem watcher
|   |   |-- watcher.go             fsnotify wrapper, debounce, dispatch loop
|   |   +-- handlers.go            Default file pattern handlers
|   |
|   |-- orchestrate/               Full orchestration pipeline
|   |   |-- pipeline.go            Step interface, Pipeline.Run(), PipelineState
|   |   |-- steps.go               Concrete steps: classify, brief-gate, create-iteration, etc.
|   |   |-- prompt.go              PromptBuilder for agent context assembly
|   |   +-- executor.go            AgentExecutor (wraps `claude` CLI invocation)
|   |
|   +-- project/                   Project detection
|       +-- detect.go              FindProjectRoot() --- walk up looking for .mind/
|
|-- go.mod
|-- go.sum
+-- Makefile
```

### Package Dependency Rules

Each package has explicit rules about what it may import:

| Package | May Import | Must Not Import |
|---------|-----------|-----------------|
| `domain/` | Go stdlib only | Everything else |
| `internal/repo/` | `domain/`, Go stdlib | `internal/service/`, `cmd/`, `tui/` |
| `internal/repo/fs/` | `domain/`, `internal/repo/` (interfaces), Go stdlib, go-toml, fsnotify | `internal/service/`, `cmd/`, `tui/` |
| `internal/repo/mem/` | `domain/`, `internal/repo/` (interfaces), Go stdlib | `internal/repo/fs/`, `internal/service/` |
| `internal/validate/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/reconcile/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/generate/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/render/` | `domain/`, Lip Gloss | `internal/service/`, `internal/repo/` |
| `internal/service/` | `domain/`, `internal/repo/` (interfaces), `internal/validate/`, `internal/reconcile/`, `internal/generate/` | `cmd/`, `tui/`, `internal/repo/fs/` |
| `internal/mcp/` | `domain/`, `internal/service/` (interfaces) | `cmd/`, `tui/`, `internal/repo/` |
| `internal/watch/` | `domain/`, `internal/service/` (interfaces), fsnotify | `cmd/`, `tui/` |
| `internal/orchestrate/` | `domain/`, `internal/service/` (interfaces), `internal/repo/` (interfaces) | `cmd/`, `tui/` |
| `internal/project/` | `domain/`, Go stdlib | Everything else |
| `cmd/` | `domain/`, `internal/service/` (interfaces), `internal/render/`, Cobra | `internal/repo/`, `tui/` |
| `tui/` | `domain/`, `internal/service/` (interfaces), Bubble Tea, Lip Gloss, Bubbles | `cmd/`, `internal/repo/` |
| `main.go` | Everything (wiring point) | N/A |

The key invariant: **`internal/repo/fs/` is only imported by `main.go`** during dependency wiring. All other packages interact with repositories through interfaces defined in `internal/repo/interfaces.go`.

---

## 4. Dependency Injection

### Constructor Injection Pattern

All dependencies are wired at startup in `main.go`. There is no global state, no `init()` functions, no service locators, no dependency injection framework.

The `Deps` struct holds all constructed services and is threaded through Cobra's context:

```go
// main.go

type Deps struct {
    Project       service.ProjectService
    Validation    service.ValidationService
    Generate      service.GenerateService
    Workflow      service.WorkflowService
    Quality       service.QualityService
    Sync          service.SyncService
    Orchestration service.OrchestrationService
    Reconcile     service.ReconciliationService
    Renderer      render.Renderer
}

func buildDeps(projectRoot string, outputMode render.OutputMode) *Deps {
    // Infrastructure (concrete implementations)
    docRepo       := fs.NewDocRepo(projectRoot)
    iterRepo      := fs.NewIterationRepo(projectRoot)
    stateRepo     := fs.NewStateRepo(projectRoot)
    configRepo    := fs.NewConfigRepo(projectRoot)
    qualityRepo   := fs.NewQualityRepo(projectRoot)
    templateRepo  := fs.NewTemplateRepo(projectRoot)
    gitOps        := fs.NewGitOps(projectRoot)
    processRunner := fs.NewProcessRunner(projectRoot)

    // Services (depend on interfaces, receive concrete implementations)
    projectSvc    := service.NewProjectService(docRepo, iterRepo, stateRepo, configRepo)
    validationSvc := service.NewValidationService(docRepo, iterRepo, configRepo)
    generateSvc   := service.NewGenerateService(templateRepo, iterRepo, docRepo)
    workflowSvc   := service.NewWorkflowService(stateRepo, iterRepo)
    qualitySvc    := service.NewQualityService(qualityRepo)
    syncSvc       := service.NewSyncService(docRepo)
    orchestSvc    := service.NewOrchestrationService(
        validationSvc, generateSvc, workflowSvc,
        docRepo, iterRepo, stateRepo, gitOps, processRunner,
    )
    reconcileSvc  := service.NewReconciliationService(docRepo, configRepo)

    // Renderer
    renderer := render.NewRenderer(outputMode, terminalWidth())

    return &Deps{
        Project:       projectSvc,
        Validation:    validationSvc,
        Generate:      generateSvc,
        Workflow:      workflowSvc,
        Quality:       qualitySvc,
        Sync:          syncSvc,
        Orchestration: orchestSvc,
        Reconcile:     reconcileSvc,
        Renderer:      renderer,
    }
}
```

### Context Threading

Cobra commands receive dependencies through `context.Context`:

```go
func main() {
    root := cmd.NewRootCmd()

    root.PersistentPreRunE = func(c *cobra.Command, args []string) error {
        projectRoot, err := project.FindProjectRoot()
        if err != nil {
            return err
        }
        outputMode := render.DetectMode(
            flagJSON(c), flagNoColor(c),
        )
        deps := buildDeps(projectRoot, outputMode)
        c.SetContext(context.WithValue(c.Context(), depsKey, deps))
        return nil
    }

    if err := root.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Commands extract what they need:

```go
// cmd/status.go

func runStatus(cmd *cobra.Command, args []string) error {
    deps := cmd.Context().Value(depsKey).(*Deps)
    health, err := deps.Project.Health(cmd.Context())
    if err != nil {
        return err
    }
    fmt.Print(deps.Renderer.RenderHealth(health))
    return nil
}
```

### Testing

Tests construct services with in-memory repositories:

```go
func TestProjectHealth(t *testing.T) {
    docRepo := mem.NewDocRepo()
    docRepo.AddFile("docs/spec/project-brief.md", briefContent)
    docRepo.AddFile("docs/spec/requirements.md", reqsContent)

    iterRepo := mem.NewIterationRepo()
    stateRepo := mem.NewStateRepo()
    configRepo := mem.NewConfigRepo()

    svc := service.NewProjectService(docRepo, iterRepo, stateRepo, configRepo)
    health, err := svc.Health(context.Background())

    assert.NoError(t, err)
    assert.Equal(t, 2, health.Zones[domain.ZoneSpec].Present)
}
```

No mocking frameworks. In-memory implementations provide full behavioral fidelity with zero filesystem access.

---

## 5. Error Handling Strategy

### Domain Layer: Sentinel and Typed Errors

```go
// domain/errors.go

// Sentinel errors (checked with errors.Is)
var (
    ErrNotProject   = errors.New("not a mind project: .mind/ not found")
    ErrBriefMissing = errors.New("project brief not found")
    ErrBriefStub    = errors.New("project brief is a stub")
    ErrNoIterations = errors.New("no iterations found")
    ErrNoWorkflow   = errors.New("no active workflow")
)

// Typed errors (checked with errors.As)
type ErrGateFailed struct {
    Gate    string   // "business-context", "deterministic", "quality-0"
    Failures []string // human-readable failure descriptions
}

func (e *ErrGateFailed) Error() string {
    return fmt.Sprintf("gate %s failed: %s", e.Gate, strings.Join(e.Failures, "; "))
}

type ErrCommandFailed struct {
    Command  string
    ExitCode int
    Stderr   string
}

func (e *ErrCommandFailed) Error() string {
    return fmt.Sprintf("command %q failed (exit %d): %s", e.Command, e.ExitCode, e.Stderr)
}

type ErrValidation struct {
    Suite    string
    Failed   int
    Warnings int
}

func (e *ErrValidation) Error() string {
    return fmt.Sprintf("validation %s: %d failures, %d warnings", e.Suite, e.Failed, e.Warnings)
}
```

### Infrastructure Layer: Wrap with Context

Infrastructure code wraps raw OS/IO errors with context using `fmt.Errorf`:

```go
func (r *FSDocRepo) Read(relPath string) ([]byte, error) {
    absPath := filepath.Join(r.projectRoot, relPath)
    data, err := os.ReadFile(absPath)
    if err != nil {
        return nil, fmt.Errorf("reading document %s: %w", relPath, err)
    }
    return data, nil
}
```

The `%w` verb preserves the original error for `errors.Is` / `errors.As` checks upstream.

### Service Layer: Map to Domain Errors

Services translate infrastructure errors into domain error types:

```go
func (s *projectService) Health(ctx context.Context) (*domain.ProjectHealth, error) {
    config, err := s.configRepo.ReadProjectConfig()
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // mind.toml missing is not fatal; use defaults
            config = domain.DefaultConfig()
        } else {
            return nil, fmt.Errorf("loading project config: %w", err)
        }
    }
    // ...
}
```

### Presentation Layer: User-Facing Messages and Exit Codes

The presentation layer converts errors to messages and exit codes. It never exposes raw error strings, stack traces, or file paths outside the project.

```go
// cmd/root.go

func handleError(err error) int {
    var gateFailed *domain.ErrGateFailed
    var cmdFailed  *domain.ErrCommandFailed
    var validation *domain.ErrValidation

    switch {
    case errors.Is(err, domain.ErrNotProject):
        fmt.Fprintln(os.Stderr, "Error: not a Mind project (no .mind/ found)")
        fmt.Fprintln(os.Stderr, "  Run 'mind init' to initialize")
        return 3 // config error

    case errors.As(err, &gateFailed):
        fmt.Fprintf(os.Stderr, "Gate failed: %s\n", gateFailed.Gate)
        for _, f := range gateFailed.Failures {
            fmt.Fprintf(os.Stderr, "  - %s\n", f)
        }
        return 1

    case errors.As(err, &cmdFailed):
        fmt.Fprintf(os.Stderr, "Command failed: %s (exit %d)\n",
            cmdFailed.Command, cmdFailed.ExitCode)
        return 1

    case errors.As(err, &validation):
        fmt.Fprintf(os.Stderr, "Validation failed: %s (%d failures)\n",
            validation.Suite, validation.Failed)
        return 1

    default:
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        return 2 // unknown
    }
}
```

### Exit Codes

| Code | Meaning | When |
|------|---------|------|
| `0` | Success | Command completed normally |
| `1` | Error | Known failure (validation failed, gate failed, command failed) |
| `2` | Unknown | Unexpected error (bugs, unhandled edge cases) |
| `3` | Config error | Not a project, mind.toml parse error, missing required config |
| `4` | Stale artifacts | Reconciliation detected stale documents (used by `mind check --strict`) |

---

## 6. Configuration Model

### Source: `mind.toml`

The primary configuration file lives at the project root alongside `.mind/`. It uses TOML format parsed by `github.com/pelletier/go-toml/v2`.

```toml
[manifest]
schema     = "mind/v2.0"
generation = 1
updated    = 2026-03-10T00:13:40Z

[project]
name        = "mind-cli"
description = "CLI and TUI for the Mind Agent Framework"
type        = "cli"

[project.stack]
language  = "go@1.23"
framework = "cobra+bubbletea"
testing   = "go-test"

[project.commands]
dev   = "go run ."
test  = "go test ./..."
lint  = "golangci-lint run ./..."
build = "go build -o mind ."

[governance]
max-retries     = 2
review-policy   = "evidence-based"
commit-policy   = "conventional"
branch-strategy = "type-descriptor"

[documents.spec.project-brief]
id     = "doc:spec/project-brief"
path   = "docs/spec/project-brief.md"
zone   = "spec"
status = "draft"
```

### Loading Sequence

```
1. Find project root
   - Start at cwd, walk up looking for .mind/ directory
   - Stop at filesystem root or MIND_ROOT override
   - Fail with ErrNotProject if not found

2. Read mind.toml
   - Path: {project_root}/mind.toml
   - If missing: use DefaultConfig() (all zero values + sensible defaults)
   - If malformed: return ErrConfigParse with line/column

3. Apply environment variable overrides
   - MIND_ROOT overrides detected project root
   - MIND_NO_COLOR forces plain output
   - MIND_JSON forces JSON output

4. Apply flag overrides
   - --json, --no-color, --verbose override env vars
```

### Override Precedence

From lowest to highest priority:

```
Defaults  <  mind.toml  <  Environment Variables  <  CLI Flags
```

Concrete example: output mode resolution:

```
Default:         Interactive (if TTY detected)
mind.toml:       (no output mode setting)
MIND_NO_COLOR=1: Plain
--json flag:     JSON (wins)
```

### Environment Variables

| Variable | Effect | Default |
|----------|--------|---------|
| `MIND_ROOT` | Override project root detection | Auto-detect (walk up for `.mind/`) |
| `MIND_NO_COLOR` | Disable colors (set to `1` or `true`) | Colors enabled when TTY |
| `MIND_JSON` | Force JSON output (set to `1` or `true`) | Interactive/plain based on TTY |
| `MIND_LOG_FILE` | Path for persistent log output | No log file |
| `MIND_VERBOSE` | Enable debug logging (set to `1` or `true`) | Disabled |

### Default Values

When `mind.toml` is missing or incomplete, the system uses sensible defaults:

| Setting | Default |
|---------|---------|
| `project.name` | Directory name |
| `project.type` | `"unknown"` |
| `governance.max-retries` | `2` |
| `governance.review-policy` | `"evidence-based"` |
| `governance.commit-policy` | `"conventional"` |
| `governance.branch-strategy` | `"type-descriptor"` |

---

## 7. Output Formatting

### Three Output Modes

```
                      +---------------------+
                      | Command produces    |
                      | domain result       |
                      +---------+-----------+
                                |
                    +-----------+-----------+
                    |                       |
              Is --json set?          Is TTY + no --no-color?
                    |                       |
               +----+----+           +------+------+
               |         |           |             |
              Yes        No         Yes            No
               |         |           |             |
         +-----v----+   |    +------v------+  +---v---------+
         | JSON      |   |    | Interactive |  | Plain       |
         | Renderer  |   |    | Renderer    |  | Renderer    |
         +-----------+   |    +-------------+  +-------------+
                         |
                   (fall through)
```

### Renderer Interface

```go
type Renderer interface {
    RenderHealth(h *domain.ProjectHealth) string
    RenderValidation(r *domain.ValidationReport) string
    RenderDoctor(diags []domain.Diagnostic) string
    RenderIterations(iters []domain.Iteration) string
    RenderDocList(docs []domain.Document) string
    RenderWorkflow(state *domain.WorkflowState) string
    RenderQualityTrend(entries []domain.QualityEntry) string
    RenderCreated(docType string, path string) string
    RenderError(err error) string
}
```

### Interactive Mode (default when TTY)

Uses Lip Gloss for styled output:
- Color-coded status indicators (green pass, red fail, yellow warning)
- Box-drawing characters for borders and tables
- Progress bars for zone completeness
- Bold headers, dimmed secondary text

### Plain Mode (piped or `--no-color`)

Clean text with no ANSI escape codes:
- `[PASS]`, `[FAIL]`, `[WARN]` prefixes instead of color
- ASCII table borders (`+---+---+`)
- No progress bars; use `3/5` fractions

### JSON Mode (`--json`)

Machine-readable structured output:
- Every top-level response has `{"ok": bool, "data": {...}, "error": string}`
- Validation results include `checks` array with `id`, `name`, `passed`, `level`, `message`
- Health includes `zones` map with per-zone completeness
- Exit code still reflects status (0 for success, 1 for failure)

### Detection Logic

```go
func DetectMode(jsonFlag bool, noColorFlag bool) OutputMode {
    if jsonFlag || os.Getenv("MIND_JSON") == "true" {
        return ModeJSON
    }
    if noColorFlag || os.Getenv("MIND_NO_COLOR") != "" {
        return ModePlain
    }
    if !isatty.IsTerminal(os.Stdout.Fd()) {
        return ModePlain
    }
    return ModeInteractive
}
```

---

## 8. Concurrency Model

### Design Principle

The CLI is predominantly single-threaded. Concurrency is introduced only where the problem demands it, and always through channels --- never shared mutable state.

### Watch Mode

Watch mode runs three goroutines:

```
Main goroutine          fsnotify goroutine          Dispatch goroutine
      |                       |                           |
      |  Start()              |                           |
      +------>  watcher.Start(ctx)                        |
      |                       |                           |
      |                  fsnotify.Events -->               |
      |                       |     debounce map          |
      |                       |     (200ms ticker)        |
      |                       |                           |
      |                       +-- Event{path, op} ------> |
      |                       |                    match handlers
      |                       |                    emit Action
      |                       |                           |
      |  <--- actions channel ----------------------------+
      |                       |                           |
 Render action                |                           |
 (run gate, log)              |                           |
```

Synchronization: buffered channels (capacity 64) for events and actions. The debounce map is local to the fsnotify goroutine --- no concurrent access. Context cancellation stops all goroutines cleanly.

### MCP Server

The MCP server is single-threaded. Stdio is inherently sequential (one request at a time). The server reads a JSON-RPC request from stdin, processes it synchronously, writes the response to stdout. No goroutines needed.

```
stdin --> ReadLine --> json.Unmarshal --> route to tool handler
                                              |
                                         call service method
                                              |
                                         json.Marshal --> stdout
```

### TUI (Bubble Tea)

Bubble Tea manages concurrency internally. The `Update` function runs on a single goroutine. Asynchronous work (loading health data, running validation) is dispatched via `tea.Cmd` functions that return `tea.Msg` values on completion.

```go
// Async command: runs in a separate goroutine managed by Bubble Tea
func (m Model) loadHealth() tea.Msg {
    health, err := m.projectSvc.Health(context.Background())
    if err != nil {
        return healthErrorMsg{err}
    }
    return healthLoadedMsg{health}
}
```

The Model is never accessed concurrently. Bubble Tea guarantees that `Update` is called sequentially.

### Orchestration Pipeline

The orchestration pipeline (`mind run`) executes steps sequentially. Each step completes before the next begins. This is deliberate: agent outputs feed into subsequent agent prompts.

```
Classify --> BriefGate --> CreateIteration --> Agent:analyst
    --> Gate:micro-a --> Agent:architect --> Agent:developer
    --> Gate:micro-b --> Agent:tester --> Gate:deterministic
    --> Agent:reviewer --> Finalize
```

Background gate execution (build/test/lint) uses `exec.CommandContext` with the pipeline context. If the user cancels (`Ctrl+C`), the context is cancelled, which terminates the running process.

Events are emitted to a channel for the TUI or plain-text progress display to consume.

### Rules

1. No `sync.Mutex` anywhere. If you need shared state, use a channel.
2. Every goroutine accepts a `context.Context` and stops on cancellation.
3. Buffered channels with explicit capacity (no unbounded queues).
4. The `domain/` package has zero concurrency primitives.

---

## 9. Logging and Observability

### Verbosity Levels

| Level | When | Output |
|-------|------|--------|
| Default | Normal operation | User-facing results only (to stdout) |
| `--verbose` | Debugging | Debug messages to stderr |
| `--log-file PATH` | Persistent logging | All levels written to file |

### Structured Log Format

Debug and log-file output uses a structured format:

```
[DEBUG] [project] detected root=/home/user/myproject
[DEBUG] [validate] running docs suite (17 checks, strict=false)
[INFO]  [validate] docs: 15/17 pass, 1 fail, 1 warn
[DEBUG] [reconcile] hash mismatch for docs/spec/requirements.md old=abc123 new=def456
[INFO]  [orchestrate] dispatching agent=analyst model=opus
[ERROR] [orchestrate] agent failed: exit code 1
```

### Components

Each subsystem logs under its own component tag:

| Component | Logs |
|-----------|------|
| `project` | Root detection, config loading, health assembly |
| `validate` | Check execution, suite results, strict mode behavior |
| `reconcile` | Hash computation, staleness detection, lock updates |
| `generate` | Template loading, variable substitution, file creation |
| `mcp` | Request/response, tool dispatch, transport events |
| `watch` | File events, debounce, handler matches, gate triggers |
| `orchestrate` | Pipeline steps, agent dispatch, gate results, retries |
| `sync` | Agent file comparison, diff detection, copy operations |
| `quality` | Score extraction, log append, trend computation |

### Implementation

Logging is implemented with Go's `log/slog` package (stdlib, Go 1.21+). A `slog.Handler` is configured at startup based on flags:

```go
func setupLogger(verbose bool, logFile string) *slog.Logger {
    var handler slog.Handler

    if logFile != "" {
        f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
        handler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
    } else if verbose {
        handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
    } else {
        handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
    }

    return slog.New(handler)
}
```

The logger is injected into services alongside repositories. No global `log` calls.

---

## 10. Security Considerations

### Path Traversal Prevention

All file paths are resolved relative to the project root. The system rejects any path that escapes the project boundary:

```go
func safePath(projectRoot, relPath string) (string, error) {
    absPath := filepath.Join(projectRoot, filepath.Clean(relPath))
    if !strings.HasPrefix(absPath, projectRoot) {
        return "", fmt.Errorf("path traversal blocked: %s", relPath)
    }
    return absPath, nil
}
```

This function is called by every repository method that accepts a path. Paths containing `..` that resolve outside the project root are rejected with an error.

### Command Execution Boundaries

Only commands declared in `mind.toml [project.commands]` are executed by the CLI:

```toml
[project.commands]
dev   = "go run ."
test  = "go test ./..."
lint  = "golangci-lint run ./..."
build = "go build -o mind ."
```

The `ProcessRunner` validates that a command name exists in the config before executing it. Arbitrary shell commands cannot be injected through flags or file contents.

```go
func (r *ProcessRunner) Run(cmdName string) error {
    cmdLine, ok := r.commands[cmdName]
    if !ok {
        return fmt.Errorf("unknown command %q: not declared in mind.toml [project.commands]", cmdName)
    }
    // Execute cmdLine with exec.Command
}
```

The orchestration pipeline (`mind run`) dispatches agents via the `claude` CLI. The agent file paths come from hardcoded domain constants (`domain.Chains()`), not from user input.

### No Secrets in Output

`mind.toml` is designed to contain no sensitive data. It holds project metadata, stack configuration, and governance policies. If a user puts secrets in `mind.toml`, that is a user error --- but the CLI does not echo config values to stdout in any mode.

The `--json` output includes document paths relative to the project root, never absolute paths. Validation messages reference file names, not full paths.

### File Permissions

The CLI respects existing file permissions. When creating new files (document generation, iteration scaffolding), it uses `0644` for files and `0755` for directories, matching standard project defaults. It never calls `chmod` on existing files.

### MCP Server: Stdio Only

The MCP server communicates exclusively over stdio (stdin/stdout). It does not open network sockets, listen on ports, or accept connections. The server process is started by the AI agent's host (e.g., Claude Code) and inherits its security context. There is no authentication mechanism because there is no network boundary.

### Input Validation

- **TOML parsing**: Malformed `mind.toml` produces a parse error with line/column, never a panic.
- **Template variables**: Variable substitution uses literal string replacement, not `text/template` or `html/template` evaluation. No code execution through template injection.
- **Search queries**: The `docs search` command uses literal string matching, not regex from user input. No ReDoS risk.
- **File content**: Markdown files are read as bytes. No interpretation or execution of embedded content.

---

## Cross-References

- **BP-02** (03-architecture.md): Complete entity field definitions, type relationships, all Go type declarations
- **BP-03** (02-ai-workflow-bridge.md): Four AI integration models, MCP tool definitions, orchestration pipeline details
- **BP-04** (01-mind-cli.md): Full command tree, command behavior specifications, TUI tab layouts, distribution plan
- **Domain Model** (docs/spec/domain-model.md): Business rules, state machines, entity constraints
- **Data Contracts** (docs/spec/api-contracts.md): JSON output schemas, MCP tool schemas, file format specifications

---

## Boundaries

### This Document Covers

- System structure (layers, packages, dependency rules)
- Design principles and their rationale
- Dependency injection and wiring pattern
- Error handling strategy across all layers
- Configuration loading and override precedence
- Output mode detection and rendering strategy
- Concurrency model and synchronization rules
- Logging and observability approach
- Security boundaries and input validation

### This Document Does NOT Cover

- Specific command behaviors and flag details (see BP-04: CLI Specification)
- Entity field lists and type declarations (see BP-02: Domain Model)
- File format schemas and JSON output contracts (see Data Contracts)
- TUI screen layouts and key bindings (see BP-04: CLI Specification)
- MCP tool definitions and JSON-RPC schemas (see BP-03: AI Workflow Bridge)
- Implementation timeline and phasing (see BP-04: CLI Specification)
