# BP-00: Project Guidelines

> The north star document. Read this first before touching the codebase or any other blueprint.

**Status**: Active
**Date**: 2026-03-11
**Cross-references**: All blueprints (BP-01 through BP-08), [mind.toml](../../mind.toml), [project-brief.md](../spec/project-brief.md)

---

## 1. Project Overview

### What mind-cli Is

mind-cli is a single-binary CLI and TUI tool for the [Mind Agent Framework](../spec/project-brief.md). It provides unified project intelligence: validation, health diagnostics, document management, reconciliation, and AI workflow integration --- all through one command, `mind`.

It ships as a Go binary with zero runtime dependencies. Run `mind status` in any project directory containing `.mind/` and you get a complete project health dashboard. No configuration, no API keys, no setup.

### Why It Exists

The Mind Agent Framework currently ships 9 bash scripts (`validate-docs.sh`, `validate-xrefs.sh`, `scaffold-iteration.sh`, etc.) with inconsistent interfaces, no discoverability, and no coordination. A developer has to memorize script names, interpret raw output, and manually bridge between deterministic checks and AI agent workflows.

mind-cli solves three problems:

1. **Unified interface** --- 32 commands behind one binary replace 9 scripts and add capabilities the scripts never had (doctor, reconciliation, TUI dashboard).
2. **Project intelligence** --- answers questions the scripts can't: "Is my documentation complete?", "What's stale?", "What should I do next?", "Is the project healthy enough to start an AI workflow?"
3. **AI workflow bridge** --- connects deterministic tooling with AI agents through 4 integration models, from lightweight pre-flight checks (Model A) to full orchestration that drives entire agent chains from the terminal (Model D).

### Who It's For

- **Framework users**: Developers who install the Mind Agent Framework and want project health visibility without memorizing script names. They run `mind status`, `mind doctor`, and `mind check all`.
- **AI workflow users**: Developers using Claude Code or other AI assistants with the framework who need pre-flight validation, real-time workflow monitoring, and post-workflow cleanup. They use `mind preflight`, `mind watch`, and `mind run`.
- **CI/CD pipelines**: Automated systems that run `mind check all --json` and gate deployments on validation results. They consume structured JSON output and use exit codes.

### What It Is NOT

- **Not an AI.** mind-cli never calls AI APIs or makes non-deterministic decisions. It dispatches agents via `claude` CLI in orchestration mode, but the binary itself is strictly deterministic.
- **Not a Claude Code replacement.** Claude Code runs agents. mind-cli manages everything *around* agent workflows: validation, scaffolding, state management, context assembly.
- **Not a framework installer.** mind-cli assumes the Mind Agent Framework is already present. `mind init` scaffolds the documentation structure, not the framework itself.
- **Not a multi-project manager.** It operates on one project at a time. The working directory determines the project scope.

### The "No AI in CLI" Philosophy

This is the single most important design constraint. The `mind` binary is deterministic: same inputs produce same outputs, every time. This means:

- No API keys required. No network calls for core functionality.
- Every command can be tested with standard Go testing. No mocks for AI services.
- The tool works offline and in air-gapped environments.
- Output is predictable. CI pipelines can rely on it.

AI capabilities live in Claude Code, agents, and the MCP server clients. mind-cli provides the *infrastructure* that makes AI workflows reliable, but it does not *contain* AI. The sole exception is `mind run`, which dispatches agents as external processes via the `claude` CLI --- but mind-cli itself never interprets or generates AI responses.

See [BP-01 Section 1: Design Principles](01-system-architecture.md) for the full rationale.

### Vision

mind-cli is the project intelligence layer that makes AI agent workflows reliable, visible, and efficient. When a developer opens a Mind Framework project, `mind status` is the first command they run. When an AI agent needs to understand project state, it calls the MCP server. When a workflow fails, the reconciliation engine shows exactly which documents are stale and why.

---

## 2. Quick Reference Map

### Blueprints

| I need to understand... | Read this |
|------------------------|-----------|
| How the system is structured | [BP-01: System Architecture](01-system-architecture.md) |
| What the core entities are | [BP-02: Domain Model](02-domain-model.md) |
| What file formats look like | [BP-03: Data Contracts](03-data-contracts.md) |
| What a specific command does | [BP-04: CLI Specification](04-cli-specification.md) |
| How the TUI screens work | [BP-05: TUI Specification](05-tui-specification.md) |
| How staleness tracking works | [BP-06: Reconciliation Engine](06-reconciliation-engine.md) |
| How AI integration models work | [BP-07: AI Workflow Integration](07-ai-workflow-integration.md) |
| What to build next | [BP-08: Implementation Roadmap](08-implementation-roadmap.md) |
| How blueprints are organized | [Blueprints INDEX](INDEX.md) |

**Reading order**: BP-01 first (architecture), then BP-02 (domain) and BP-03 (contracts), then BP-06 (reconciliation), then BP-04/BP-05/BP-07 in any order, and finally BP-08 (roadmap). See [INDEX.md](INDEX.md) for the full dependency graph.

### Project Files

| File | Purpose |
|------|---------|
| [mind.toml](../../mind.toml) | Project manifest --- identity, stack, document registry, governance rules |
| [CLAUDE.md](../../CLAUDE.md) | Claude Code routing table --- resource index for AI agents |
| [docs/spec/project-brief.md](../spec/project-brief.md) | Vision, scope, deliverables, success metrics |
| [docs/spec/requirements.md](../spec/requirements.md) | Functional and non-functional requirements |
| [docs/spec/architecture.md](../spec/architecture.md) | High-level architecture summary (spec zone) |
| [docs/spec/domain-model.md](../spec/domain-model.md) | Entity definitions and business rules (spec zone) |
| [docs/state/current.md](../state/current.md) | Active work, known issues, next priorities |
| [docs/state/workflow.md](../state/workflow.md) | Current AI workflow state |
| [docs/knowledge/glossary.md](../knowledge/glossary.md) | Domain terminology reference |

---

## 3. Technical Standards

### 3.1 Go Conventions

**Language version**: Go 1.23 minimum. Use language features up to 1.23 (range-over-func, enhanced type inference). Do not use features from unreleased versions.

**Project layout** (defined in [BP-01](01-system-architecture.md) and [BP-08](08-implementation-roadmap.md)):

```
cmd/           # Cobra command definitions (presentation layer)
domain/        # Pure domain types, enums, business rules (zero external imports)
internal/      # Service and infrastructure packages
  validate/    # Validation engine
  reconcile/   # Reconciliation engine
  repo/        # Repository interfaces and implementations
  render/      # Output formatting (text, JSON, table)
tui/           # Bubble Tea TUI components
mcp/           # MCP server implementation
```

**Formatting**: Use `gofmt`. No exceptions. Never commit unformatted Go code. Configure your editor to format on save.

**Linting**: Use `golangci-lint` with the project's linter configuration. Run `golangci-lint run ./...` before every commit. Treat lint warnings as errors.

**Error handling**:

```go
// Do this: wrap with context
if err != nil {
    return fmt.Errorf("loading mind.toml: %w", err)
}

// Not this: naked return
if err != nil {
    return err
}

// Not this: swallowed error
result, _ := riskyOperation()
```

Always wrap errors with what you were doing when the error occurred. Use `%w` for wrapping so callers can inspect with `errors.Is` and `errors.As`. Never silently discard errors unless you have a documented reason.

**Naming**:

- Exported identifiers: `PascalCase` (`ProjectService`, `ValidateDocuments`)
- Unexported identifiers: `camelCase` (`parseManifest`, `checkBrief`)
- Acronyms stay uppercase: `ID`, `URL`, `MCP`, `JSON`, `TUI`, `CLI`
- Package names: singular, lowercase, descriptive of what they provide (`validate`, `reconcile`, `render`)

**Packages**: Small and focused. A package is named for what it *provides*, not what it *contains*. `validate` provides validation, not "a bunch of validation-related things." If a package needs a `utils` suffix, it's too broad --- split it.

**Interfaces**: Define where consumed, not where implemented. Keep them small.

```go
// Do this: small interface, defined where it's used
type DocumentReader interface {
    ReadDocument(ctx context.Context, path string) (*domain.Document, error)
}

// Not this: large interface defined in the implementation package
type DocumentRepository interface {
    Read(...)
    Write(...)
    Delete(...)
    List(...)
    Search(...)
    Count(...)
    // 15 more methods...
}
```

**Initialization**: Zero use of `init()` functions. All initialization happens explicitly in `main.go`. No global mutable state. Everything flows through constructor injection.

**Context**: Use `context.Context` for cancellation and timeouts. Pass it as the first parameter to any function that performs I/O or might take significant time.

### 3.2 Architecture Rules

The system follows a strict 4-layer architecture defined in [BP-01: System Architecture](01-system-architecture.md).

```
Presentation  →  Service  →  Domain  →  Infrastructure
(cmd/, tui/,     (internal/   (domain/)   (internal/repo/,
 mcp/)            service/)                internal/fs/)
```

**Rules that are never broken**:

1. **Dependencies flow DOWN only.** `cmd/` imports `internal/service/`. `internal/service/` imports `domain/`. Never the reverse.
2. **Domain has ZERO imports beyond stdlib.** No `os`, no `filepath`, no `io`, no third-party packages. Domain types are pure Go: structs, enums, validation functions, business rules.
3. **Presentation is THIN.** A command handler parses flags, calls a service method, and renders the result. If you're writing business logic in `cmd/`, stop and move it to a service.
4. **Services orchestrate.** They compose domain types with infrastructure implementations. Business logic that involves multiple entities or I/O lives here.
5. **Infrastructure is SWAPPABLE.** Filesystem access goes through repository interfaces. Tests use in-memory implementations from `internal/repo/mem/`. Production uses real filesystem implementations.

### 3.3 Testing Standards

Coverage targets (enforced in CI):

| Package | Minimum |
|---------|---------|
| `domain/` | 80% |
| `internal/validate/` | 80% |
| `internal/reconcile/` | 80% |
| Overall | 70% |

**Test naming**: `TestFunctionName_Scenario_Expected`

```go
func TestCheckBrief_MissingFile_ReturnsBriefMissing(t *testing.T) { ... }
func TestComputeHash_EmptyFile_ReturnsZeroHash(t *testing.T) { ... }
func TestPropagateStaleness_TransitiveDep_MarksDownstream(t *testing.T) { ... }
```

**Table-driven tests** for anything with multiple input/output combinations:

```go
func TestClassifyDocument(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected domain.DocStatus
    }{
        {"empty file", "", domain.StatusMissing},
        {"headings only", "# Title\n## Section\n", domain.StatusStub},
        {"real content", "# Title\n\nThis document describes...", domain.StatusDraft},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ClassifyDocument(tt.content)
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

**Golden file tests** for CLI output: store expected output in `testdata/` files. Compare actual output against golden files. Update golden files with a `-update` flag when output intentionally changes.

**Filesystem tests**: Always use `t.TempDir()`. Never read from or write to real project directories. Never depend on the working directory.

**Repository mocks**: Use in-memory implementations in `internal/repo/mem/`. These are real implementations, not mocks --- they behave like the real thing but store data in maps instead of the filesystem.

**Non-negotiable rules**:

- No `time.Sleep` in tests. If a test needs timing, use channels, `sync.WaitGroup`, or test clocks.
- No flaky tests. A test that fails intermittently is a broken test. Fix it or delete it.
- Tests are documentation. A reader should understand the expected behavior from the test name and structure alone.

### 3.4 Git & Commit Standards

**Commit format**: [Conventional Commits](https://www.conventionalcommits.org/).

```
feat: add staleness propagation to reconciliation engine
fix: check-docs crash when mind.toml has no documents section
refactor: extract hash computation into dedicated package
test: add golden file tests for mind status output
docs: update BP-06 with transitive staleness algorithm
chore: bump golangci-lint to v1.62
perf: use mtime fast-path to skip unchanged files in reconcile
```

**Branch naming**: `{type}/{descriptor}`

```
feat/reconcile-engine
fix/check-docs-crash
refactor/extract-hash-package
test/golden-file-status
```

**Pull requests**:

- One logical change per PR. A PR that adds reconciliation should not also refactor error handling.
- Title is descriptive (same style as commit messages).
- Description links to the relevant blueprint section.
- All checks must pass before merge.

**Never commit**:

- Generated binaries (the `mind` binary)
- `.env` files or any file containing secrets
- IDE-specific files (`.idea/`, `.vscode/settings.json` with personal paths)
- OS artifacts (`.DS_Store`, `Thumbs.db`)

**Note**: `mind.lock` is a generated file but is intentionally committed for CI verification. This is a deliberate choice documented in [BP-03](03-data-contracts.md).

### 3.5 Documentation Standards

**Code comments**: Explain *why*, never *what*. If the code needs a comment explaining what it does, the code should be clearer.

```go
// Do this: explains the non-obvious "why"
// SHA-256 chosen over xxhash for reproducibility across Go versions.
// The performance cost is ~2ms for a typical document, acceptable for our use case.
hash := sha256.Sum256(content)

// Not this: restates the code
// Compute the SHA-256 hash of the content
hash := sha256.Sum256(content)
```

**Package docs**: Every package's primary file (the one matching the package name) must have a package-level doc comment explaining what the package provides and how to use it.

**Exported functions**: GoDoc-style comments. Start with the function name.

```go
// ValidateDocuments runs all 17 documentation checks against the project
// at the given root path and returns a ValidationReport.
func ValidateDocuments(ctx context.Context, root string) (*ValidationReport, error) {
```

**Blueprint maintenance**: Blueprints are living documents, not historical artifacts. When implementation diverges from a blueprint, **update the blueprint**. A blueprint that doesn't match the code is worse than no blueprint at all.

**CLAUDE.md**: Keep `CLAUDE.md` and `.claude/CLAUDE.md` in sync with actual project state. These files are the routing table for AI agents --- stale routes cause wasted tokens and wrong behavior.

---

## 4. Golden Standards

These describe what excellence looks like for mind-cli. Every feature should be measured against these standards.

### 4.1 Command Design

Every command follows these principles (detailed specifications in [BP-04](04-cli-specification.md)):

- **Zero arguments work.** `mind status`, `mind doctor`, `mind check all` --- the most common invocation takes no arguments. Sensible defaults everywhere.
- **`--json` on everything.** Every command that produces structured data supports `--json`. Agents and scripts rely on this.
- **Actionable errors.** Every error message includes what went wrong AND what to do about it.

```
# Do this:
Error: mind.toml not found in current directory or parent directories.
  Run 'mind init' to create a new Mind Framework project, or
  change to a directory containing a Mind Framework project.

# Not this:
Error: file not found
```

- **Fast.** Every command completes in under 200ms for typical projects (10-50 documents). The exception is `mind run`, which dispatches AI agents and reports progress.
- **Self-sufficient help.** `mind help <command>` tells you everything you need to know. If a user has to leave the terminal to understand a command, the help text failed.
- **Meaningful exit codes** (defined in [BP-03](03-data-contracts.md)):

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Validation failure (checks found issues) |
| 2 | Runtime error (crash, I/O failure) |
| 3 | Configuration error (bad mind.toml, missing project) |
| 4 | Workflow error (agent dispatch failure, gate failure) |

### 4.2 Output Quality

- **Color + symbols.** Interactive output uses both color AND symbols (`PASS`, `FAIL`, `WARN`, `SKIP`). Never rely on color alone --- accessibility matters.
- **Progress indicators.** Anything that takes more than 500ms gets a spinner or progress bar.
- **Respect the terminal.** Default output is concise. Use `--verbose` for detail. Do not dump walls of text.
- **Streams go to the right place.** Data goes to stdout. Errors go to stderr. Progress indicators go to stderr. This means `mind check all --json 2>/dev/null` gives you clean JSON.
- **JSON schema compliance.** JSON output matches the documented schemas in [BP-03](03-data-contracts.md) exactly. No surprise fields, no missing fields, no type changes.

### 4.3 Reconciliation Quality

The reconciliation engine (detailed in [BP-06](06-reconciliation-engine.md)) is the most technically demanding component. Standards are strict:

- **Deterministic hashing.** The same file content produces the same SHA-256 hash on every platform, every time. Normalize line endings before hashing.
- **Correct staleness propagation.** If A depends on B and B changes, A is stale. If C depends on A, C is also stale (transitive). No false positives (marking fresh docs as stale). No false negatives (missing genuinely stale docs).
- **Atomic lock file writes.** Write to a temporary file, then rename. Never write directly to `mind.lock` --- a crash during write must not corrupt the lock file.
- **Hash is truth, mtime is optimization.** The mtime fast-path skips hash computation when a file's modification time hasn't changed. But if mtime says "unchanged" and hash says "changed," hash wins. The fast-path is a performance optimization, not a correctness mechanism.

### 4.4 MCP Server Quality

The MCP server (detailed in [BP-07](07-ai-workflow-integration.md)) is how AI agents interact with project intelligence. Standards:

- **Structured JSON responses.** Tools return parsed, structured data. Never raw text or markdown blobs. An agent should be able to use the response without string parsing.
- **Stable tool names.** Renaming a tool breaks every agent configuration that references it. Tool names are part of the public API. Treat them with the same care as command names.
- **Precise descriptions.** Agents use tool descriptions to decide when to call them. A vague description ("get project info") causes agents to call the wrong tool. Be specific ("Returns the project's current validation status, including pass/fail counts for all check suites and a list of failing checks with descriptions").
- **Graceful error handling.** Malformed input produces a proper JSON-RPC error response, not a crash. Bad parameters produce descriptive error messages, not stack traces.

### 4.5 TUI Quality

The TUI (detailed in [BP-05](05-tui-specification.md)) uses Bubble Tea with the MVU (Model-View-Update) architecture:

- **Minimum size: 80x24.** The TUI must render correctly at the smallest common terminal size. Degrade gracefully at smaller sizes (truncate, don't crash).
- **Handle resizes.** Terminal resize events must not crash the application or corrupt the display. Redraw cleanly.
- **Keyboard-first.** All interactions are available via keyboard. Mouse support is optional, keyboard support is mandatory.
- **Clean exit.** `q` and `Ctrl+C` exit cleanly, restoring terminal state. A crash that leaves the terminal in raw mode is unacceptable.
- **Loading states.** Async operations show a spinner or "Loading..." indicator. An unresponsive-looking TUI is a broken TUI.

---

## 5. Anti-Patterns

Things that have gone wrong in similar projects, or patterns that will hurt mind-cli if introduced.

### 5.1 Architecture Anti-Patterns

**God service.** Don't put all logic in `ProjectService`. The service layer has focused services: `ValidationService`, `ReconciliationService`, `DocumentService`, `WorkflowService`. Each owns a clear responsibility.

```go
// Not this:
type ProjectService struct { /* 30 methods covering everything */ }

// Do this:
type ValidationService struct { ... }  // runs checks, produces reports
type ReconcileService struct { ... }   // hash tracking, staleness
type DocumentService struct { ... }    // CRUD for documents, scaffolding
type WorkflowService struct { ... }    // pre-flight, handoff, orchestration
```

**Domain I/O.** Don't import `os`, `filepath`, or `io` in the `domain/` package. Domain types describe *what* things are and *what rules* govern them. They never touch the filesystem.

```go
// domain/document.go --- this is correct:
type Document struct {
    ID       DocumentID
    Path     string
    Zone     Zone
    Status   DocStatus
    Hash     string
}

func (d Document) IsStub() bool {
    return d.Status == StatusStub
}

// domain/document.go --- this is WRONG:
func (d Document) Load() ([]byte, error) {
    return os.ReadFile(d.Path) // NO: domain must not do I/O
}
```

**Presentation logic.** Don't compute business results in `cmd/` files. Command handlers parse flags, call a service, and render the result. That's it.

```go
// cmd/check.go --- correct:
func runCheckDocs(cmd *cobra.Command, args []string) error {
    report, err := svc.ValidateDocuments(cmd.Context(), projectRoot)
    if err != nil {
        return err
    }
    return renderer.RenderValidationReport(report, outputFormat)
}

// cmd/check.go --- WRONG:
func runCheckDocs(cmd *cobra.Command, args []string) error {
    files, _ := os.ReadDir(projectRoot + "/docs/spec")
    for _, f := range files {
        content, _ := os.ReadFile(f.Name())
        if len(content) < 100 {
            fmt.Println("WARN: stub detected:", f.Name())
        }
        // 50 more lines of inline business logic...
    }
}
```

**Implicit dependencies.** Don't use package-level variables or `init()` for dependency wiring. All dependencies are created and wired explicitly in `main.go`.

**Interface pollution.** Don't create interfaces "just in case." Create them when you have 2+ implementations (real filesystem + in-memory for tests) or when a consumer needs to be decoupled from a specific implementation.

### 5.2 Code Anti-Patterns

**Swallowed errors.** Never discard an error unless you can articulate why it doesn't matter, and leave a comment explaining the reasoning.

```go
// Acceptable: documented reason
_ = tempFile.Close() // best-effort cleanup; file will be removed by t.TempDir()

// Unacceptable: silent discard
result, _ := parseManifest(data)
```

**Panic in library code.** Never use `panic()` in `domain/` or `internal/`. Panics are reserved for truly unrecoverable startup failures in `main.go` (e.g., failing to wire dependencies). Everything else returns an error.

**String typing.** Don't use raw strings where a typed constant communicates intent.

```go
// Do this:
type Zone string
const (
    ZoneSpec       Zone = "spec"
    ZoneBlueprints Zone = "blueprints"
    ZoneState      Zone = "state"
)

// Not this:
func getDocuments(zone string) { ... } // any typo compiles fine
```

**Deep nesting.** If you're 4+ levels deep in `if`/`for` blocks, refactor. Use early returns, extract helper functions, or restructure the logic.

```go
// Do this: early returns
func processDocument(doc *Document) error {
    if doc == nil {
        return ErrNilDocument
    }
    if doc.IsStub() {
        return nil // nothing to process
    }
    return doc.Validate()
}

// Not this: deep nesting
func processDocument(doc *Document) error {
    if doc != nil {
        if !doc.IsStub() {
            if err := doc.Validate(); err != nil {
                return err
            }
        }
    }
    return nil
}
```

**Over-abstraction.** Three lines of direct code beats a `DocumentProcessorFactoryBuilder`. Abstraction should remove duplication or enable testing, not add layers for their own sake.

**Test-only exports.** Don't export functions solely for testing. Use `_test.go` files in the same package, or create internal test helpers. If you must expose something for cross-package testing, put it in an `internal/testutil/` package.

### 5.3 Process Anti-Patterns

**Big bang PRs.** Don't submit 50-file PRs. Break work into logical increments. A PR that adds "the entire reconciliation engine" should instead be 3-4 PRs: hash computation, dependency graph, staleness propagation, integration with status/doctor.

**Blueprint drift.** Don't implement features that diverge from blueprints without updating the blueprint first. If you discover a better approach during implementation, update the blueprint, then implement. Blueprints that don't match the code erode trust in all project documentation.

**Premature optimization.** Don't optimize before profiling. The mtime fast-path in reconciliation ([BP-06](06-reconciliation-engine.md)) is justified by measurements showing 80%+ of files are unchanged between runs. Most optimizations aren't justified --- prove the bottleneck first.

**Feature creep.** Don't add commands not specified in [BP-04](04-cli-specification.md). If a new command is needed, add it to the blueprint first with full specification (behavior, flags, output format, exit codes, examples), then implement. The blueprint is the contract.

**Skipping tests.** Don't merge code without tests. If a deadline forces it, create a tracking issue immediately and address it in the next PR. Untested code is a liability that compounds.

---

## 6. Goals to Pursue

### 6.1 Short-Term (Phase 1--2)

These are the goals for [Phase 1: Core CLI](08-implementation-roadmap.md) and [Phase 1.5: Reconciliation](08-implementation-roadmap.md):

- **Replace the bash scripts.** Ship a CLI that is genuinely faster and more informative than the 9 scripts it replaces. Users should prefer `mind check docs` over `validate-docs.sh` within a week of trying it.
- **Achieve validation parity.** Identical results between bash scripts and Go implementation. Same check IDs, same pass/fail counts, same failure messages. Verified by integration tests that run both implementations side by side.
- **Make `mind status` the default first command.** When developers open a Mind Framework project, `mind status` should be the first thing they run. It should be fast enough (<200ms) and informative enough (health score, document counts, stale docs, workflow state) to be worth the habit.
- **Build reconciliation early.** The reconciliation engine ([BP-06](06-reconciliation-engine.md)) differentiates mind-cli from a simple script wrapper. Hash-based staleness detection with transitive propagation is the foundation for project intelligence. It must be solid before the TUI or MCP server builds on top of it.

### 6.2 Medium-Term (Phase 3--4)

These are the goals for [Phase 3: AI Bridge A+B](08-implementation-roadmap.md) and [Phase 4: AI Bridge C+D](08-implementation-roadmap.md):

- **Make MCP the standard agent interface.** AI agents should interact with project state through structured JSON tools ([BP-07](07-ai-workflow-integration.md)), not by parsing markdown files. The MCP server should be the default way agents get project context.
- **Reduce token waste.** Pre-flight context assembly ([BP-07 Model A](07-ai-workflow-integration.md)) should reduce wasted tokens by 10-30% by giving agents exactly the context they need instead of letting them explore the filesystem.
- **Provide real-time visibility.** Watch mode ([BP-07 Model C](07-ai-workflow-integration.md)) should give developers real-time visibility into AI workflows. They should never need to wonder "what is the AI doing?" or "did the last agent pass the gate?"
- **Make `mind run` reliable.** Full orchestration ([BP-07 Model D](07-ai-workflow-integration.md)) should be reliable enough for unattended execution. Deterministic gates between agents catch failures early. Retry logic (max 2 retries per agent) handles transient issues.

### 6.3 Long-Term

- **Reference implementation.** Make mind-cli the reference for project intelligence tooling in AI-assisted development. The patterns (reconciliation, pre-flight checks, deterministic gates, MCP integration) should be reusable beyond the Mind Framework.
- **Community.** Build a community around the Mind Agent Framework. mind-cli is the on-ramp --- it's what people install first and interact with daily.
- **Multi-project support.** Explore managing a workspace of Mind Framework projects (monorepo support, cross-project dependencies). This is explicitly out of scope for v1 but is a natural evolution.
- **Plugin system.** Consider an extension mechanism for custom validators, generators, and integrations. The interface-based architecture ([BP-01](01-system-architecture.md)) makes this feasible, but it should not be built until there are concrete use cases.
- **Stay lean.** Resist the urge to absorb everything into the CLI. The core should remain focused on project intelligence. Features that don't serve validation, reconciliation, or AI workflow integration belong elsewhere.

---

## 7. Decision Log

Key architectural decisions and their rationale. When ADRs exist in `docs/spec/decisions/`, they are referenced here.

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| Language | Go 1.23 | Best TUI ecosystem (Bubble Tea), single-binary distribution, fast compilation, no runtime dependencies |
| CLI framework | Cobra | Industry standard for Go CLIs, auto-generated completions and man pages, used by kubectl/gh/hugo |
| TUI framework | Bubble Tea + Lip Gloss | Best-in-class terminal UI in Go, MVU architecture matches our state management needs, rich component library (Bubbles) |
| Architecture | 4-layer ([BP-01](01-system-architecture.md)) | Clean separation of concerns, domain stays pure and testable, infrastructure is swappable via interfaces |
| Reconciliation | SHA-256 + dependency graph ([BP-06](06-reconciliation-engine.md)) | Inspired by NixOS, provides transitive staleness detection, makes document dependencies visible and verifiable |
| AI integration | 4 incremental models ([BP-07](07-ai-workflow-integration.md)) | Each model is independently valuable, Model B (MCP server) is the inflection point, risk managed by incremental delivery rather than big-bang integration |
| State formats | TOML + JSON + Markdown | Each format chosen for its strength: TOML for human-edited config (mind.toml), JSON for machine-generated state (mind.lock), Markdown for human-readable state (workflow.md) |
| Distribution | GoReleaser | Cross-platform builds (linux/darwin, amd64/arm64), AUR and Homebrew support, GitHub releases with checksums |
| No AI in CLI | Hard constraint | Determinism, testability, offline operation, no API keys required, CI-friendly. AI capabilities live in Claude Code and agent contexts. See [BP-01 Section 1](01-system-architecture.md). |
| Dependency injection | Constructor injection in main.go | No framework, no init(), no global state. Services receive their dependencies explicitly. Enables in-memory testing without mocks. |
| Error strategy | Wrapped errors with context ([BP-01](01-system-architecture.md)) | Every error includes what operation failed. `%w` wrapping enables `errors.Is`/`errors.As` inspection. User-facing errors include remediation steps. |

---

## 8. Glossary of Project Terms

Quick reference for terminology used across all blueprints. See also [docs/knowledge/glossary.md](../knowledge/glossary.md).

| Term | Definition |
|------|-----------|
| **Mind Framework** | The Claude Code-based agent system with 7 core agents, 4-zone documentation model, and structured workflows. mind-cli is its companion tooling. |
| **Zone** | One of 5 documentation areas: **spec** (stable specifications), **blueprints** (planning artifacts), **state** (volatile runtime state), **iterations** (immutable per-change history), **knowledge** (reference material). |
| **Iteration** | A per-change tracking folder under `docs/iterations/` containing 5 artifacts: plan, progress, completion, review-checklist, and lessons-learned. |
| **Agent Chain** | The sequence of agents dispatched for a request type. Example: analyst -> architect -> developer -> tester -> reviewer. Defined in [BP-02](02-domain-model.md). |
| **Reconciliation** | The process of comparing declared documents (in mind.toml) against actual filesystem state (in mind.lock) to detect missing, orphaned, or stale documents. See [BP-06](06-reconciliation-engine.md). |
| **Staleness** | The condition where a downstream document hasn't been updated after an upstream dependency changed. Detected by comparing hashes in mind.lock against current file hashes. |
| **Transitive Staleness** | When staleness propagates through dependency chains: if A depends on B and B is stale, A is transitively stale even if A's own content hasn't changed. |
| **Business Context Gate** | The pre-workflow check ensuring `project-brief.md` is present and substantive (not a stub). Blocks `NEW_PROJECT` and `COMPLEX_NEW` request types. |
| **Deterministic Gate** | Build/lint/test commands run between agents to verify code quality. Defined per-project in mind.toml. |
| **Micro-Gate** | Lightweight checks run after specific agents (e.g., after analyst, verify requirements format). Faster than full deterministic gates. |
| **Pre-Flight** | Preparation steps before an AI workflow: validate documents, create iteration folder, assemble context, check business context gate. See [BP-07 Model A](07-ai-workflow-integration.md). |
| **Handoff** | Cleanup steps after an AI workflow: validate created artifacts, update state documents, clear workflow.md, record metrics. |
| **MCP** | Model Context Protocol. A JSON-RPC-based protocol for exposing tools to AI models. mind-cli implements an MCP server via `mind serve`. See [BP-07 Model B](07-ai-workflow-integration.md). |
| **Stub** | A document containing only template scaffolding (headings, HTML comments, placeholder text) with no substantive content. Detected by content classification. |
| **Convergence** | A multi-persona analysis process that produces a scored synthesis document. Used by the `/analyze` command in the Mind Framework. |
| **mind.toml** | The project manifest file. Declares project identity, stack, document registry, governance rules, and profiles. See [BP-03](03-data-contracts.md). |
| **mind.lock** | Machine-generated state file. Contains SHA-256 hashes, dependency edges, and timestamps for reconciliation. See [BP-03](03-data-contracts.md) and [BP-06](06-reconciliation-engine.md). |
