# BP-08: Implementation Roadmap

> What do we build, in what order, and how do we verify it works?

**Status**: Active
**Date**: 2026-03-11
**Cross-references**: [BP-01](01-system-architecture.md) for architecture, [BP-02](02-domain-model.md) for entities, [BP-04](01-mind-cli.md) for command specifications, [BP-05](02-ai-workflow-bridge.md) for AI integration models

---

## 1. Phase Overview

Five phases deliver mind-cli from scaffolded skeleton to production-ready tool. Each phase is independently shippable and produces a usable artifact. Phases are sequential --- later phases depend on earlier ones --- but work within a phase can be parallelized across packages.

```
Phase     Name                  Delivers                                          Depends On
------    --------------------  ------------------------------------------------  ----------
1         Core CLI              Foundation commands, validation engine, rendering  (none)
1.5       Reconciliation        Hash tracking, staleness propagation, mind.lock   Phase 1
2         TUI Dashboard         5-tab interactive interface                        Phase 1.5
3         AI Bridge (A+B)       Pre-flight, handoff, MCP server                   Phase 2
4         AI Bridge (C+D)       Watch mode, full orchestration                    Phase 3
5         Polish                Completions, releases, CI/CD, docs                Phase 4
```

The numbering reflects dependency, not importance. Phase 1.5 exists because reconciliation is a self-contained engine that must integrate into status, check, and doctor before the TUI can display staleness data.

---

## 2. Phase 1: Core CLI

**Goal**: A usable CLI that replaces the most-used bash scripts and provides the full command surface for deterministic project management.

### Scope

**Commands delivered**:

| Command | Description |
|---------|-------------|
| `mind status` | Project health dashboard (plain text + JSON) |
| `mind init` | Initialize Mind Framework in current directory |
| `mind doctor` | Deep diagnostics with actionable fix suggestions |
| `mind doctor --fix` | Auto-fix resolvable issues (create dirs, stubs, fix naming) |
| `mind create adr "<title>"` | Create auto-numbered ADR |
| `mind create blueprint "<title>"` | Create auto-numbered blueprint + update INDEX.md |
| `mind create iteration <type> "<name>"` | Create iteration folder with 5 template files |
| `mind create spike "<title>"` | Create technical spike report |
| `mind create convergence "<title>"` | Create convergence analysis template |
| `mind create brief` | Interactive project brief creation (guided prompts) |
| `mind check docs` | 17-check documentation validation |
| `mind check refs` | 11-check cross-reference validation |
| `mind check config` | YAML config validation |
| `mind check all` | Unified validation across all suites |
| `mind docs list [--zone ZONE]` | List documents grouped by zone |
| `mind docs stubs` | Find stub documents that need content |
| `mind docs tree` | Visual tree view of all documentation |
| `mind workflow status` | Show current workflow state |
| `mind workflow history` | List past iterations chronologically |
| `mind version [--short]` | Version and build info |
| `mind help [command]` | Help for all commands (Cobra auto-generated) |

**Cross-cutting concerns**:

- All commands support `--json` output mode (BP-01 Section 7)
- Project root auto-detection: walk up from cwd looking for `.mind/` directory
- `mind.toml` parsing with all sections: manifest, project, project.stack, project.commands, documents, governance, profiles
- Three output modes: interactive (Lip Gloss styled), plain (no ANSI), JSON (machine-readable)
- Exit codes per BP-01: 0 success, 1 failure, 2 unknown, 3 config error, 4 stale artifacts
- Graceful "not a Mind project" error when `.mind/` is missing

### Packages to Implement

**Domain layer** (`domain/`):

| File | Contents |
|------|----------|
| `project.go` | `Project`, `Config`, `Manifest`, `ProjectMeta`, `StackConfig`, `GovernanceConfig` |
| `document.go` | `Document`, `Zone`, `DocStatus`, `Brief`, `BriefGate`, `BriefClassification` |
| `iteration.go` | `Iteration`, `Artifact`, `RequestType`, `IterationStatus`, `IterationDirName()` |
| `workflow.go` | `WorkflowState`, `DispatchEntry`, `CompletedArtifact` |
| `validation.go` | `ValidationReport`, `CheckResult`, `CheckLevel`, `Diagnostic` |
| `quality.go` | `QualityScore`, `QualityEntry` |
| `health.go` | `ProjectHealth`, `ZoneHealth` |
| `errors.go` | Sentinel errors (`ErrNotProject`, `ErrConfigParse`, `ErrBriefMissing`, `ErrStubDocument`, etc.), typed error wrappers |

**Repository layer** (`internal/repo/`):

| File | Contents |
|------|----------|
| `interfaces.go` | `DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `TemplateRepo`, `QualityRepo` |
| `fs/doc_repo.go` | `FSDocRepo` --- list files by zone, read content, detect stubs, search |
| `fs/iteration_repo.go` | `FSIterationRepo` --- list iterations, read overview, count artifacts |
| `fs/state_repo.go` | `FSStateRepo` --- read/write `docs/state/workflow.md` and `docs/state/current.md` |
| `fs/config_repo.go` | `FSConfigRepo` --- parse `mind.toml` via go-toml/v2 |
| `fs/template_repo.go` | `FSTemplateRepo` --- load `.mind/docs/templates/*.md`, apply variable substitution |
| `fs/project.go` | `FindProjectRoot()` --- walk up for `.mind/` |

**Service layer** (`internal/service/`):

| File | Contents |
|------|----------|
| `interfaces.go` | All service interfaces (`ProjectService`, `ValidationService`, `GenerateService`, `WorkflowService`) |
| `project.go` | `ProjectService` implementation: `Health()`, `Doctor()`, `DoctorFix()`, `Init()` |
| `validation.go` | `ValidationService` implementation: `CheckDocs()`, `CheckRefs()`, `CheckConfig()`, `CheckAll()` |
| `generate.go` | `GenerateService` implementation: `CreateADR()`, `CreateBlueprint()`, `CreateIteration()`, `CreateSpike()`, `CreateConvergence()`, `CreateBrief()` |
| `workflow.go` | `WorkflowService` implementation: `Status()`, `History()`, `Show()`, `Clean()` |

**Validation engine** (`internal/validate/`):

| File | Contents |
|------|----------|
| `check.go` | `Check` struct, `CheckFunc` type, `Suite` struct, `Suite.Run()` framework |
| `docs.go` | 17-check documentation suite (zone existence, required files, stub detection, naming conventions) |
| `refs.go` | 11-check cross-reference suite (internal links, blueprint INDEX, iteration references) |
| `config.go` | Config validation (YAML syntax, required fields, schema version) |
| `report.go` | `ReportBuilder` --- aggregate suite results into unified `ValidationReport` |

**Rendering** (`internal/render/`):

| File | Contents |
|------|----------|
| `render.go` | `Renderer` interface, `OutputMode` enum, `DetectMode()` (TTY check, --json flag, --no-color flag) |
| `interactive.go` | `InteractiveRenderer` --- Lip Gloss styled output (progress bars, boxes, colored check marks) |
| `plain.go` | `PlainRenderer` --- clean text, no ANSI codes |
| `json.go` | `JSONRenderer` --- structured JSON output |

**Document generation** (`internal/generate/`):

| File | Contents |
|------|----------|
| `template.go` | Template loading from `.mind/docs/templates/`, `{{.Title}}` / `{{.Date}}` / `{{.Seq}}` substitution |
| `sequence.go` | Auto-numbering: scan existing files, parse numeric prefixes, return next sequence |
| `slugify.go` | Title to kebab-case: lowercase, replace spaces with hyphens, strip non-alphanumeric |

**Project detection** (`internal/project/`):

| File | Contents |
|------|----------|
| `detect.go` | `FindProjectRoot()` --- walk up from cwd looking for `.mind/` directory |

**Command layer** (`cmd/`):

| File | Contents |
|------|----------|
| `root.go` | Root command, global flags (`--json`, `--no-color`, `--verbose`), PersistentPreRunE for project detection and dependency wiring |
| `status.go` | `mind status` --- call `ProjectService.Health()`, render result |
| `init.go` | `mind init` --- call `ProjectService.Init()` |
| `doctor.go` | `mind doctor` / `mind doctor --fix` |
| `create.go` | `mind create {adr,blueprint,iteration,spike,convergence,brief}` |
| `docs.go` | `mind docs {list,tree,stubs}` |
| `check.go` | `mind check {docs,refs,config,all}` |
| `workflow.go` | `mind workflow {status,history}` |
| `version.go` | `mind version [--short]` |

### Acceptance Criteria

- [ ] `mind status` shows zone health (5 zones with present/total counts), workflow state, and warnings for a valid project
- [ ] `mind status --json` produces valid JSON matching the `ProjectHealth` schema from BP-02
- [ ] `mind doctor` reports diagnostics with actionable fix suggestions (e.g., "Run: mind init --with-github")
- [ ] `mind doctor --fix` creates missing directories, adds `.gitkeep` files, creates stub documents from templates
- [ ] `mind create iteration new "rest-api"` creates correctly numbered iteration folder (e.g., `007-NEW_PROJECT-rest-api/`) with 5 files: `overview.md`, `changes.md`, `test-summary.md`, `validation.md`, `retrospective.md`
- [ ] `mind create adr "Use PostgreSQL"` creates `docs/spec/decisions/NNN-use-postgresql.md` with correct template
- [ ] `mind check docs` runs all 17 checks and reports pass/fail/warn per check
- [ ] `mind check refs` runs all 11 checks and reports pass/fail/warn per check
- [ ] `mind check all --json` produces combined `ValidationReport` as JSON
- [ ] Project root detection works from nested directories (e.g., running from `src/pkg/` finds root `../../`)
- [ ] `mind.toml` parsed correctly with all sections: manifest, project, project.stack, project.commands, documents, governance
- [ ] Exit codes match BP-01 specification: 0 success, 1 failure, 3 config error
- [ ] All commands gracefully handle "not a Mind project" with exit code 3 and helpful message
- [ ] `mind docs stubs` lists documents classified as stubs with their zone and path
- [ ] `mind docs tree` renders a visual tree matching the project's documentation structure
- [ ] `mind version` shows version, commit hash, build date, Go version
- [ ] Output mode auto-detection: styled when TTY, plain when piped, JSON when `--json`
- [ ] `go test ./...` passes with >70% coverage on `domain/` and `internal/validate/`

### Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/spf13/cobra` | v1.8+ | CLI framework |
| `github.com/pelletier/go-toml/v2` | v2.2+ | TOML parsing for mind.toml |
| `github.com/charmbracelet/lipgloss` | v0.10+ | Styled terminal output |

---

## 3. Phase 1.5: Reconciliation Engine

**Goal**: Hash-based content tracking integrated into existing commands. Detects when documents are out of date with respect to their dependencies.

### Scope

**Commands delivered**:

| Command | Description |
|---------|-------------|
| `mind reconcile` | Compute hashes, build dependency graph, update `mind.lock` |
| `mind reconcile --check` | Verify without writing --- exits 0 on clean, exits 4 on stale (CI mode) |
| `mind reconcile --force` | Reset all hashes, re-compute everything |
| `mind reconcile --graph` | Visualize dependency graph as ASCII |

**Integrations with Phase 1 commands**:

| Existing Command | New Behavior |
|------------------|-------------|
| `mind status` | Shows staleness count panel when `mind.lock` exists (e.g., "3 stale documents") |
| `mind check all` | Includes staleness check in unified validation report |
| `mind doctor` | Reports stale documents as diagnostic findings with "Run: mind reconcile" suggestion |

**Data model**:

The reconciliation engine reads the `[[graph]]` section from `mind.toml` to build a directed dependency graph between documents. It computes SHA-256 hashes of document content, stores them in `mind.lock`, and propagates staleness through the graph when a source document changes.

```
mind.toml [[graph]] defines:
  docs/spec/project-brief.md  -->  docs/spec/requirements.md
  docs/spec/requirements.md   -->  docs/spec/architecture.md
  docs/spec/architecture.md   -->  docs/spec/domain-model.md

When project-brief.md changes:
  requirements.md  = STALE (direct dependency)
  architecture.md  = STALE (transitive)
  domain-model.md  = STALE (transitive)
```

### Packages to Implement

**Reconciliation engine** (`internal/reconcile/`):

| File | Contents |
|------|----------|
| `hash.go` | SHA-256 computation over normalized file content (strip trailing whitespace, normalize line endings). mtime fast-path: skip hash if file mtime unchanged since last run |
| `graph.go` | Build directed graph from `mind.toml [[graph]]` edges. Topological sort. Cycle detection with cycle path reporting |
| `propagate.go` | Staleness propagation: BFS/DFS from changed nodes, mark all reachable downstream nodes as stale |
| `engine.go` | Top-level `Reconcile()` orchestration: load lock file, compute hashes, build graph, detect changes, propagate staleness, write lock file |

**Repository extension** (`internal/repo/fs/`):

| File | Contents |
|------|----------|
| `lock_repo.go` | `FSLockRepo` --- read/write `mind.lock` (JSON format with per-document entries: path, SHA-256 hash, mtime, stale flag) |

**Service extension** (`internal/service/`):

| File | Contents |
|------|----------|
| `reconciliation.go` | `ReconciliationService` implementation: `Reconcile()`, `Check()`, `Force()`, `Graph()` |

**Command** (`cmd/`):

| File | Contents |
|------|----------|
| `reconcile.go` | `mind reconcile` with `--check`, `--force`, `--graph` flags |

### Acceptance Criteria

- [ ] `mind reconcile` creates `mind.lock` with correct SHA-256 hashes for all documents declared in `mind.toml [documents]`
- [ ] Changing a file's content and re-running `mind reconcile` detects the change and updates the hash
- [ ] Dependency graph propagates staleness downstream: if A depends on B and B changes, A is marked stale
- [ ] Transitive propagation works: if A depends on B depends on C, and C changes, both A and B are marked stale
- [ ] `mind reconcile --check` exits 0 on clean state, exits 4 when stale documents exist
- [ ] `mind reconcile --force` clears all staleness flags and re-computes every hash
- [ ] `mind reconcile --graph` renders an ASCII visualization of the dependency graph with stale nodes highlighted
- [ ] `mind status` shows a staleness count panel when `mind.lock` exists (e.g., "Staleness: 3 documents need update")
- [ ] `mind check all` includes a reconciliation check in its unified report
- [ ] `mind doctor` reports stale documents as findings with severity and fix suggestion
- [ ] mtime fast-path skips hash computation for unchanged files (verified by timing: second run measurably faster)
- [ ] Cycle detection reports an error with the full cycle path (e.g., "cycle: A -> B -> C -> A")
- [ ] Reconciliation of 50 documents completes in <200ms
- [ ] `mind.lock` format is valid JSON and survives round-trip (read, write, read produces identical output)
- [ ] `go test ./internal/reconcile/...` passes with >80% coverage

### Dependencies

Phase 1 must be complete. No new external dependencies --- SHA-256 is in Go's standard library (`crypto/sha256`).

---

## 4. Phase 2: TUI Dashboard

**Goal**: Full-screen interactive TUI with 5 tabs that surfaces all project intelligence from Phases 1 and 1.5.

### Scope

**Command delivered**:

| Command | Description |
|---------|-------------|
| `mind tui` | Launch full-screen interactive dashboard |

**Tab layout**:

```
[1 Status]  [2 Docs]  [3 Iterations]  [4 Checks]  [5 Quality]
```

**Tab 1 --- Status**:
- Zone health bars (progress indicators per zone: spec, blueprints, state, iterations, knowledge)
- Workflow state panel (idle / active with agent chain progress)
- Staleness panel (count of stale documents from reconciliation, top stale items)
- Warnings list (stub documents, missing files, stale artifacts)
- Quick action keys: `c` create, `d` doctor, `v` validate, `o` open, `s` sync

**Tab 2 --- Documents**:
- Zone filter bar (all, spec, blueprints, state, iterations, knowledge)
- Browseable document list with columns: path, status (complete/stub/stale), date, size
- Inline search (`/` to filter)
- Document preview panel (first 20 lines of selected document)
- Actions: `enter` preview, `e` edit in $EDITOR, `n` new document

**Tab 3 --- Iterations**:
- Type filter bar (all, NEW_PROJECT, ENHANCEMENT, BUG_FIX, REFACTOR)
- Chronological iteration list with columns: number, type, name, status, date, artifact completeness (e.g., 5/5)
- Expandable detail view showing overview.md content and artifact list
- Actions: `enter` expand, `o` open overview, `v` open validation

**Tab 4 --- Checks**:
- Live validation results (runs all suites on tab activation)
- Expandable suite sections: docs (17 checks), refs (11 checks), config, reconciliation
- Per-check detail: check ID, description, pass/fail/warn, message
- Summary bar: total pass/fail/warn counts
- Action: `r` re-run validation

**Tab 5 --- Quality**:
- Convergence score trend chart (ASCII line chart)
- Score history table: date, topic, variant, overall score, gate pass/fail
- Latest convergence detail: 6 dimension scores
- Only renders if `quality-log.yml` exists; shows "No quality data" otherwise

**Navigation**:
- Tab switching: number keys (1-5), Tab/Shift+Tab to cycle
- Standard: `q` or Ctrl+C to quit, `?` for help overlay
- Responsive layout: adapts to terminal width (minimum 80x24, optimal 120+)

### Packages to Implement

**TUI application** (`tui/`):

| File | Contents |
|------|----------|
| `app.go` | Top-level Bubble Tea `Model` --- manages active tab, delegates `Init()`, `Update()`, `View()` to active tab model. Services injected at construction |
| `status.go` | Status tab model --- calls `ProjectService.Health()` and `ReconciliationService.Check()`, renders zone bars and panels |
| `docs.go` | Documents tab model --- calls `DocRepo` (via service) for file listing, implements zone filtering and search |
| `iterations.go` | Iterations tab model --- calls `WorkflowService.History()`, implements type filtering and detail expansion |
| `checks.go` | Checks tab model --- calls `ValidationService.CheckAll()`, renders expandable suite results |
| `quality.go` | Quality tab model --- calls `QualityService.History()`, renders ASCII trend chart |
| `styles.go` | Lip Gloss theme definitions: colors, borders, spacing, tab indicators, progress bar styles |
| `keys.go` | Key binding definitions: tab navigation, per-tab actions, global keys (quit, help) |

### Acceptance Criteria

- [ ] `mind tui` launches a full-screen interactive interface
- [ ] All 5 tabs render correctly at 80x24 minimum terminal size
- [ ] Tab switching works with number keys (1-5) and Tab/Shift+Tab cycling
- [ ] Status tab shows zone health progress bars, workflow state, staleness panel, and warnings
- [ ] Documents tab supports zone filtering, inline search, and document preview
- [ ] Iterations tab supports type filtering and shows artifact completeness ratio
- [ ] Checks tab runs validation suites on activation and shows expandable results per suite
- [ ] Quality tab shows score trend chart when `quality-log.yml` exists
- [ ] Quality tab shows "No quality data" message when `quality-log.yml` is absent
- [ ] Terminal resize handled gracefully (layout adapts without crash)
- [ ] `q` and Ctrl+C quit cleanly (restores terminal state)
- [ ] Responsive layout renders usably at both 80-col and 120-col widths
- [ ] Tab state is preserved when switching away and back (e.g., scroll position, selected item)
- [ ] Loading states shown while data is fetched (spinner or "Loading..." text)

### Dependencies

Phase 1.5 must be complete (staleness data needed for Status tab and Checks tab).

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/charmbracelet/bubbletea` | v1.2+ | TUI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | v0.10+ | TUI styling (already from Phase 1) |
| `github.com/charmbracelet/bubbles` | v0.20+ | TUI components (tables, spinners, text input, viewport) |

---

## 5. Phase 3: AI Bridge --- Pre-Flight + MCP

**Goal**: Connect the CLI to AI workflows via two integration models. Model A (pre-flight/handoff) prepares and cleans up around AI sessions. Model B (MCP server) gives AI agents direct access to project intelligence as callable tools.

### Scope

**Commands delivered**:

| Command | Description |
|---------|-------------|
| `mind preflight "<request>"` | Classify request, run business context gate, create iteration, create branch, assemble context, generate prompt |
| `mind preflight --resume` | Detect and display interrupted workflow for resumption |
| `mind handoff <iteration-id>` | Validate iteration artifacts, run deterministic gate, update state, clear workflow |
| `mind serve` | Start MCP server on stdio (JSON-RPC 2.0 protocol) |

**Pre-flight flow** (Model A):
1. Classify the user's request into a `RequestType` (NEW_PROJECT, ENHANCEMENT, BUG_FIX, REFACTOR, COMPLEX_NEW)
2. Run business context gate (check brief presence and content for NEW_PROJECT/COMPLEX_NEW)
3. Run documentation validation (17 checks, fail on blockers)
4. Create iteration folder with overview.md populated from classification
5. Create git branch using naming convention from `mind.toml` governance
6. Assemble context package (read relevant spec docs, recent iterations, convergence docs)
7. Generate orchestrator prompt (copy to clipboard or print)

**MCP server** (Model B):
- 16 tools exposed via JSON-RPC 2.0 over stdio
- Each tool maps directly to a service method --- no duplication of logic
- `.mcp.json` configuration file for Claude Code integration

**MCP tool inventory**:

| Tool | Service Method | Returns |
|------|---------------|---------|
| `mind_status` | `ProjectService.Health()` | Project health summary (JSON) |
| `mind_doctor` | `ProjectService.Doctor()` | Diagnostics with fix suggestions |
| `mind_check_brief` | `ProjectService.CheckBrief()` | Business context gate result |
| `mind_validate_docs` | `ValidationService.CheckDocs()` | 17-check validation results |
| `mind_validate_refs` | `ValidationService.CheckRefs()` | 11-check cross-reference results |
| `mind_list_iterations` | `WorkflowService.History()` | All iterations with type/status/date |
| `mind_show_iteration` | `WorkflowService.Show()` | Single iteration details + artifacts |
| `mind_read_state` | `WorkflowService.Status()` | Current workflow state (parsed) |
| `mind_update_state` | `WorkflowService.UpdateState()` | Write workflow position/artifacts |
| `mind_create_iteration` | `GenerateService.CreateIteration()` | Create iteration folder + overview.md |
| `mind_list_stubs` | `ProjectService.ListStubs()` | Documents that need content |
| `mind_check_gate` | `ValidationService.RunGate()` | Deterministic gate result (build/lint/test) |
| `mind_log_quality` | `QualityService.Log()` | Extract convergence scores to quality log |
| `mind_search_docs` | `ProjectService.SearchDocs()` | Full-text search across docs/ |
| `mind_read_config` | `ProjectService.Config()` | Parsed mind.toml manifest |
| `mind_suggest_next` | `ProjectService.SuggestNext()` | Next action suggestion based on state |

### Packages to Implement

**Orchestration** (`internal/orchestrate/`):

| File | Contents |
|------|----------|
| `classify.go` | Request classification engine: keyword matching + heuristics to determine `RequestType` |
| `preflight.go` | Pre-flight logic: gate check, iteration creation, branch creation, context assembly |
| `prompt.go` | `PromptBuilder` --- assemble agent prompts from project context, conventions, prior artifacts |

**MCP server** (`internal/mcp/`):

| File | Contents |
|------|----------|
| `server.go` | JSON-RPC 2.0 protocol handler: read requests from stdin, dispatch to tools, write responses to stdout |
| `tools.go` | Tool registration: 16 tool definitions with name, description, input schema, handler function |
| `transport.go` | stdio transport: buffered reader/writer, message framing |

**Commands** (`cmd/`):

| File | Contents |
|------|----------|
| `preflight.go` | `mind preflight` with positional request arg and `--resume` flag |
| `handoff.go` | `mind handoff <iteration-id>` |
| `serve.go` | `mind serve` --- start MCP server, block until stdin closes |

### Acceptance Criteria

- [ ] `mind preflight "create: REST API"` classifies as NEW_PROJECT, runs gate, creates iteration folder, creates git branch
- [ ] `mind preflight` blocks with actionable error when brief is missing for NEW_PROJECT
- [ ] `mind preflight` blocks with actionable error when documentation has blocking failures
- [ ] `mind preflight --resume` detects interrupted workflow from `docs/state/workflow.md` and displays resumption info
- [ ] `mind handoff 007` validates iteration 007 artifacts (overview, changes, test-summary, validation, retrospective), runs deterministic gate, updates `docs/state/current.md`, clears workflow state
- [ ] `mind handoff` reports missing artifacts with specific file names
- [ ] `mind serve` starts and responds correctly to MCP `initialize` request
- [ ] `mind serve` responds to `tools/list` with all 16 tool definitions including input schemas
- [ ] All 16 MCP tools return correct JSON responses when called
- [ ] MCP `mind_validate_docs` returns identical results to `mind check docs --json`
- [ ] MCP `mind_create_iteration` creates iteration folder and returns path
- [ ] MCP `mind_check_gate` executes build/lint/test commands from `mind.toml [project.commands]` and returns structured results
- [ ] MCP `mind_update_state` writes workflow state and returns confirmation
- [ ] Claude Code connects to MCP server via `.mcp.json` and can call tools successfully
- [ ] MCP server handles malformed JSON-RPC requests gracefully (returns proper error response)
- [ ] MCP server handles unknown tool names gracefully (returns "method not found" error)
- [ ] Pre-flight classification matches expected types: "create" -> NEW_PROJECT, "fix" -> BUG_FIX, "refactor" -> REFACTOR, "add/enhance" -> ENHANCEMENT

### Dependencies

Phase 2 must be complete (TUI integration for preflight output). No new external dependencies for MCP --- JSON-RPC is hand-implemented over stdio using `encoding/json`.

---

## 6. Phase 4: AI Bridge --- Watch + Orchestration

**Goal**: Real-time filesystem monitoring (Model C) and full AI workflow automation via the `claude` CLI (Model D).

### Scope

**Commands delivered**:

| Command | Description |
|---------|-------------|
| `mind watch` | Filesystem watcher with event logging to stdout |
| `mind watch --tui` | Watch mode with live TUI dashboard |
| `mind run "<request>"` | Full orchestration pipeline: pre-flight through reviewer |
| `mind run --resume` | Resume interrupted orchestration from saved state |
| `mind run --dry-run "<request>"` | Simulate pipeline without AI calls |
| `mind run --tui "<request>"` | Orchestration with live TUI progress view |

**Watch mode** (Model C):
- Monitor filesystem for changes relevant to Mind Framework workflows
- Debounce rapid changes (300ms window --- AI agents write multiple files quickly)
- Trigger appropriate actions on file patterns:

| File Pattern | Action |
|-------------|--------|
| `docs/state/workflow.md` | Parse state, log agent progress |
| `docs/iterations/*/overview.md` | New iteration detected, log type + chain |
| `docs/iterations/*/changes.md` | Developer changes detected, run micro-gate B checks |
| `docs/iterations/*/validation.md` | Reviewer findings detected, parse MUST/SHOULD/COULD |
| `docs/spec/requirements.md` | Analyst output detected, run micro-gate A checks |
| `docs/spec/architecture.md` | Architect output detected, log summary |
| `docs/knowledge/*-convergence.md` | Convergence complete, run 23-check validation |
| `docs/spec/project-brief.md` | Brief changed, re-run business context gate |
| `src/**/*` | Code changed, run build/test in background |

**Full orchestration** (Model D):
- Execute the complete agent pipeline by dispatching each agent as a separate `claude` CLI invocation
- Pipeline steps: pre-flight -> classify -> gate -> create iteration -> [analyst -> micro-gate A -> architect -> session split -> developer -> micro-gate B -> tester -> deterministic gate -> reviewer] -> handoff
- Quality gate enforcement between agents with retry logic (max 2 retries per agent, configurable via `mind.toml` governance.max-retries)
- Session splitting: NEW_PROJECT auto-splits after architect, prompts user to continue or pause
- State persistence to `docs/state/workflow.md` at each step (enables `--resume`)
- Agent prompt assembly via `PromptBuilder`: load agent instructions, inject project context, inject prior agent outputs, inject conventions

### Packages to Implement

**Filesystem watcher** (`internal/watch/`):

| File | Contents |
|------|----------|
| `watcher.go` | fsnotify wrapper with debounce (300ms), file pattern matching, dispatch loop |
| `handlers.go` | Pattern-to-action handlers: parse workflow state changes, trigger gate checks, run background build/test |

**Orchestration extension** (`internal/orchestrate/`):

| File | Contents |
|------|----------|
| `pipeline.go` | `Step` interface, `Pipeline` struct with `Run()` and `Resume()`, `PipelineState` for persistence |
| `steps.go` | Concrete step implementations: `ClassifyStep`, `BriefGateStep`, `CreateIterationStep`, `DispatchAgentStep`, `GateStep`, `SessionSplitStep`, `HandoffStep` |
| `executor.go` | `AgentExecutor` interface + `ClaudeExecutor` implementation: invoke `claude --print --model {model} --allowedTools {tools}` as subprocess, capture output |

**TUI extensions** (`tui/`):

| File | Contents |
|------|----------|
| `watch.go` | Watch mode TUI model: live activity log, pre-gate status panel, workflow progress bar |
| `run.go` | Orchestration TUI model: pipeline progress view, live agent output panel, token/cost counter |

**Commands** (`cmd/`):

| File | Contents |
|------|----------|
| `watch.go` | `mind watch` with `--tui` flag |
| `run.go` | `mind run` with `--resume`, `--dry-run`, `--tui` flags |

### Acceptance Criteria

- [ ] `mind watch` detects file changes in the project and logs events to stdout
- [ ] Watch mode debounces rapid changes within a 300ms window (multiple writes produce one event)
- [ ] Watch mode triggers appropriate gate checks on file changes (e.g., micro-gate B when `changes.md` updates)
- [ ] Watch mode runs build/test in background when source files change
- [ ] `mind watch --tui` shows live activity log, workflow progress, and pre-gate status panel
- [ ] `mind run --dry-run "fix: bug"` shows planned pipeline without making any AI calls or filesystem changes
- [ ] `mind run --dry-run` output includes: classification, agent chain, gate plan, estimated cost range
- [ ] `mind run "create: feature"` dispatches agents sequentially via `claude` CLI
- [ ] Each agent receives correctly assembled prompt with project context, conventions, and prior agent outputs
- [ ] Quality gates run between agents: micro-gate A after analyst, micro-gate B after developer, deterministic gate before reviewer
- [ ] Gate failure triggers retry: agent is re-dispatched with gate feedback included in prompt
- [ ] Retry count respects `mind.toml` governance.max-retries (default 2)
- [ ] After max retries, pipeline proceeds with documented concerns rather than blocking
- [ ] Session splitting prompts user after architect for NEW_PROJECT (with continue/pause choice)
- [ ] `mind run --resume` reads state from `docs/state/workflow.md` and continues from the last completed step
- [ ] `mind run --tui "create: feature"` shows pipeline progress with live agent output, token count, and elapsed time
- [ ] Orchestration handles `claude` CLI errors gracefully: non-zero exit, timeout, missing binary
- [ ] Pipeline state is saved to `docs/state/workflow.md` after each step completes

### Dependencies

Phase 3 must be complete (preflight and MCP server provide the foundation).

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/fsnotify/fsnotify` | v1.7+ | Filesystem event monitoring |

External runtime requirement: `claude` CLI must be on PATH for `mind run`.

---

## 7. Phase 5: Polish

**Goal**: Production-ready release with cross-platform distribution, shell integrations, and remaining feature commands.

### Scope

**Commands delivered**:

| Command | Description |
|---------|-------------|
| `mind completion bash\|zsh\|fish` | Generate shell completions (Cobra built-in) |
| `mind docs search "<query>"` | Full-text search across docs/ with context |
| `mind docs open <path-or-id>` | Open document in $EDITOR with fuzzy finding |
| `mind quality log <convergence-file>` | Extract scores, append to quality-log.yml |
| `mind quality history` | Show quality score trends |
| `mind quality report` | Summary report of all convergence analyses |
| `mind sync agents [--check]` | Synchronize `.mind/conversation/agents/` to `.github/agents/` |
| `mind check convergence <file>` | 23-check convergence validation |

**Distribution**:
- GoReleaser configuration for cross-platform builds (linux/darwin + amd64/arm64, windows/amd64)
- GitHub Actions CI/CD pipeline (test, lint, build on push; release on tag)
- AUR package (`mind-cli-bin`)
- Homebrew formula (`jf-ferraz/tap/mind`)
- Man page generation via Cobra

**Service extensions** (`internal/service/`):

| File | Contents |
|------|----------|
| `quality.go` | `QualityService`: `Log()`, `History()`, `Report()` |
| `sync.go` | `SyncService`: `SyncAgents()`, `CheckSync()` |

**Validation extension** (`internal/validate/`):

| File | Contents |
|------|----------|
| `convergence.go` | 23-check convergence output validation suite |

**Commands** (`cmd/`):

| File | Contents |
|------|----------|
| `completion.go` | `mind completion {bash,zsh,fish}` |
| `quality.go` | `mind quality {log,history,report}` |
| `sync.go` | `mind sync agents` |

### Acceptance Criteria

- [ ] `mind completion bash` generates valid bash completions (verified by sourcing and tab-completing)
- [ ] `mind completion zsh` generates valid zsh completions
- [ ] `mind completion fish` generates valid fish completions
- [ ] `mind docs search "query"` returns matching documents with surrounding context lines
- [ ] `mind docs search` with no matches returns empty result (not an error)
- [ ] `mind docs open project-brief` opens `docs/spec/project-brief.md` in `$EDITOR`
- [ ] `mind docs open` with ambiguous input shows fuzzy-matched candidates
- [ ] `mind quality log convergence.md` extracts 6 dimension scores + overall score, appends to `quality-log.yml`
- [ ] `mind quality history` shows score trend data (date, topic, score, gate pass/fail)
- [ ] `mind sync agents` copies agent files from `.mind/conversation/agents/` to `.github/agents/`
- [ ] `mind sync agents --check` reports drift without making changes
- [ ] `mind check convergence file.md` runs 23 checks and reports results
- [ ] GoReleaser builds for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] Release archives: tar.gz for linux/darwin, zip for windows
- [ ] SHA-256 checksums generated for all release artifacts
- [ ] CI pipeline runs `go test`, `golangci-lint`, and `go build` on every push to main and on PRs
- [ ] CI release workflow triggers on tag push and runs GoReleaser
- [ ] Binary size < 15MB for all platforms
- [ ] Startup time < 50ms (measured with `hyperfine --warmup 3 './mind version'`)
- [ ] `go vet ./...` passes with zero warnings
- [ ] `golangci-lint run ./...` passes with zero warnings
- [ ] Man pages generated via `cobra/doc` and included in release archives

### Dependencies

Phase 4 must be complete.

| Dependency | Version | Purpose |
|------------|---------|---------|
| `gopkg.in/yaml.v3` | v3.0+ | YAML parsing for quality-log.yml |

Build-time tools:

| Tool | Purpose |
|------|---------|
| GoReleaser | Cross-platform builds, release packaging |
| golangci-lint | Linter aggregation |
| GitHub Actions | CI/CD pipeline |

---

## 8. Package Structure

Complete Go package layout showing every file, its owner phase, and import relationships.

```
mind-cli/
|
|-- main.go                              [Phase 1]  Entry point, Deps struct, dependency wiring
|                                                    Imports: cmd, internal/service, internal/repo/fs,
|                                                             internal/render, internal/project
|
|-- cmd/                                 Cobra commands (Presentation Layer)
|   |-- root.go                          [Phase 1]  Root command, global flags (--json, --no-color,
|   |                                                --verbose), PersistentPreRunE, error handler
|   |-- status.go                        [Phase 1]  mind status
|   |-- init.go                          [Phase 1]  mind init [--name] [--with-github] [--from-existing]
|   |-- doctor.go                        [Phase 1]  mind doctor [--fix]
|   |-- create.go                        [Phase 1]  mind create {adr,blueprint,iteration,spike,
|   |                                                convergence,brief}
|   |-- docs.go                          [Phase 1]  mind docs {list,tree,stubs}
|   |                                    [Phase 5]  mind docs {open,search} added
|   |-- check.go                         [Phase 1]  mind check {docs,refs,config,all}
|   |                                    [Phase 5]  mind check convergence added
|   |-- workflow.go                      [Phase 1]  mind workflow {status,history,show,clean}
|   |-- version.go                       [Phase 1]  mind version [--short]
|   |-- reconcile.go                     [Phase 1.5] mind reconcile [--check] [--force] [--graph]
|   |-- tui_cmd.go                       [Phase 2]  mind tui (launches Bubble Tea app)
|   |-- preflight.go                     [Phase 3]  mind preflight "<request>" [--resume]
|   |-- handoff.go                       [Phase 3]  mind handoff <iteration-id>
|   |-- serve.go                         [Phase 3]  mind serve (MCP server)
|   |-- watch.go                         [Phase 4]  mind watch [--tui]
|   |-- run.go                           [Phase 4]  mind run "<request>" [--resume] [--dry-run] [--tui]
|   |-- quality.go                       [Phase 5]  mind quality {log,history,report}
|   |-- sync.go                          [Phase 5]  mind sync agents [--check]
|   +-- completion.go                    [Phase 5]  mind completion {bash,zsh,fish}
|
|-- domain/                              Domain entities (zero external imports, Go stdlib only)
|   |-- project.go                       [Phase 1]  Project, Config, Manifest, ProjectMeta,
|   |                                                StackConfig, GovernanceConfig
|   |-- document.go                      [Phase 1]  Document, Zone, DocStatus, Brief, BriefGate,
|   |                                                BriefClassification
|   |-- iteration.go                     [Phase 1]  Iteration, Artifact, RequestType,
|   |                                                IterationStatus, IterationDirName()
|   |-- workflow.go                      [Phase 1]  WorkflowState, DispatchEntry, CompletedArtifact
|   |-- validation.go                    [Phase 1]  ValidationReport, CheckResult, CheckLevel,
|   |                                                Diagnostic
|   |-- quality.go                       [Phase 1]  QualityScore, QualityEntry
|   |-- health.go                        [Phase 1]  ProjectHealth, ZoneHealth
|   |-- reconcile.go                     [Phase 1.5] LockFile, LockEntry, DependencyEdge, ContentHash
|   |-- agent.go                         [Phase 3]  AgentChain, AgentRef, Chains()
|   +-- errors.go                        [Phase 1]  Sentinel errors, typed error wrappers
|
|-- internal/
|   |-- service/                         Service interfaces & implementations
|   |   |-- interfaces.go               [Phase 1]  All service interfaces
|   |   |                                [Phase 1.5] ReconciliationService added
|   |   |                                [Phase 3]  OrchestrationService added
|   |   |-- project.go                  [Phase 1]  ProjectService: Health(), Doctor(), DoctorFix(),
|   |   |                                            Init(), ListStubs(), SearchDocs(), Config(),
|   |   |                                            SuggestNext(), CheckBrief()
|   |   |-- validation.go              [Phase 1]  ValidationService: CheckDocs(), CheckRefs(),
|   |   |                                            CheckConfig(), CheckAll(), RunGate()
|   |   |-- generate.go                [Phase 1]  GenerateService: CreateADR(), CreateBlueprint(),
|   |   |                                            CreateIteration(), CreateSpike(),
|   |   |                                            CreateConvergence(), CreateBrief()
|   |   |-- workflow.go                [Phase 1]  WorkflowService: Status(), History(), Show(),
|   |   |                                            Clean(), UpdateState()
|   |   |-- reconciliation.go          [Phase 1.5] ReconciliationService: Reconcile(), Check(),
|   |   |                                            Force(), Graph()
|   |   |-- quality.go                 [Phase 5]  QualityService: Log(), History(), Report()
|   |   |-- sync.go                    [Phase 5]  SyncService: SyncAgents(), CheckSync()
|   |   +-- orchestration.go           [Phase 3]  OrchestrationService: Preflight(), Resume(),
|   |                                    [Phase 4]    Handoff(), RunPipeline(), DryRun()
|   |
|   |-- repo/                           Repository interfaces
|   |   |-- interfaces.go              [Phase 1]  DocRepo, IterationRepo, StateRepo, ConfigRepo,
|   |   |                                            TemplateRepo, QualityRepo
|   |   |                                [Phase 1.5] LockRepo added
|   |   |-- fs/                          Filesystem implementations
|   |   |   |-- doc_repo.go            [Phase 1]  FSDocRepo
|   |   |   |-- iteration_repo.go      [Phase 1]  FSIterationRepo
|   |   |   |-- state_repo.go          [Phase 1]  FSStateRepo
|   |   |   |-- config_repo.go         [Phase 1]  FSConfigRepo
|   |   |   |-- template_repo.go       [Phase 1]  FSTemplateRepo
|   |   |   |-- project.go             [Phase 1]  FindProjectRoot()
|   |   |   |-- quality_repo.go        [Phase 5]  FSQualityRepo
|   |   |   |-- lock_repo.go           [Phase 1.5] FSLockRepo
|   |   |   |-- git.go                 [Phase 3]  GitOps: CreateBranch(), CurrentBranch(), Status()
|   |   |   +-- process.go             [Phase 3]  ProcessRunner: Run() for build/test/lint
|   |   +-- mem/                         In-memory implementations (testing)
|   |       |-- doc_repo.go            [Phase 1]  MemDocRepo
|   |       |-- iteration_repo.go      [Phase 1]  MemIterationRepo
|   |       +-- state_repo.go          [Phase 1]  MemStateRepo
|   |
|   |-- validate/                        Validation suites
|   |   |-- check.go                   [Phase 1]  Check, CheckFunc, Suite, Suite.Run()
|   |   |-- docs.go                    [Phase 1]  17-check documentation suite
|   |   |-- refs.go                    [Phase 1]  11-check cross-reference suite
|   |   |-- config.go                  [Phase 1]  Config validation
|   |   |-- convergence.go            [Phase 5]  23-check convergence output validation
|   |   |-- gate.go                    [Phase 3]  Deterministic gate (build + test + lint)
|   |   +-- report.go                  [Phase 1]  ReportBuilder
|   |
|   |-- reconcile/                       Reconciliation engine
|   |   |-- hash.go                    [Phase 1.5] SHA-256 computation, mtime fast-path
|   |   |-- graph.go                   [Phase 1.5] Dependency graph, topological sort, cycle detection
|   |   |-- propagate.go              [Phase 1.5] Staleness propagation (BFS from changed nodes)
|   |   +-- engine.go                  [Phase 1.5] Top-level Reconcile() orchestration
|   |
|   |-- render/                          Output rendering
|   |   |-- render.go                  [Phase 1]  Renderer interface, OutputMode, DetectMode()
|   |   |-- interactive.go            [Phase 1]  InteractiveRenderer (Lip Gloss styled)
|   |   |-- plain.go                   [Phase 1]  PlainRenderer (no ANSI codes)
|   |   +-- json.go                    [Phase 1]  JSONRenderer
|   |
|   |-- generate/                        Document generation
|   |   |-- template.go               [Phase 1]  Template loading, variable substitution
|   |   |-- sequence.go               [Phase 1]  Auto-numbering (scan, parse, next)
|   |   +-- slugify.go                [Phase 1]  Title -> kebab-case
|   |
|   |-- mcp/                            MCP server
|   |   |-- server.go                  [Phase 3]  JSON-RPC 2.0 protocol handler
|   |   |-- tools.go                   [Phase 3]  16 tool definitions + handlers
|   |   +-- transport.go              [Phase 3]  stdio transport (buffered reader/writer)
|   |
|   |-- watch/                           Filesystem watcher
|   |   |-- watcher.go                [Phase 4]  fsnotify wrapper, debounce (300ms), dispatch loop
|   |   +-- handlers.go               [Phase 4]  File pattern -> action handlers
|   |
|   |-- orchestrate/                     Full orchestration pipeline
|   |   |-- classify.go               [Phase 3]  Request classification (keyword -> RequestType)
|   |   |-- preflight.go              [Phase 3]  Pre-flight logic (gate, iteration, branch, context)
|   |   |-- prompt.go                  [Phase 3]  PromptBuilder (context assembly for agents)
|   |   |-- pipeline.go               [Phase 4]  Step interface, Pipeline.Run(), Pipeline.Resume()
|   |   |-- steps.go                   [Phase 4]  Concrete steps (ClassifyStep, DispatchAgentStep,
|   |   |                                            GateStep, SessionSplitStep, HandoffStep)
|   |   +-- executor.go               [Phase 4]  AgentExecutor interface, ClaudeExecutor
|   |                                               (wraps `claude` CLI invocation)
|   |
|   +-- project/                         Project detection
|       +-- detect.go                  [Phase 1]  FindProjectRoot() --- walk up for .mind/
|
|-- tui/                                 Bubble Tea TUI (Presentation Layer)
|   |-- app.go                          [Phase 2]  Main model, tab switching, service injection
|   |-- status.go                       [Phase 2]  Tab 1: project health dashboard
|   |-- docs.go                         [Phase 2]  Tab 2: document browser with zone filtering
|   |-- iterations.go                   [Phase 2]  Tab 3: iteration timeline with type filtering
|   |-- checks.go                       [Phase 2]  Tab 4: live validation results
|   |-- quality.go                      [Phase 2]  Tab 5: convergence quality trends
|   |-- watch.go                        [Phase 4]  Watch mode TUI (live activity, gate status)
|   |-- run.go                          [Phase 4]  Orchestration TUI (pipeline progress, agent output)
|   |-- styles.go                       [Phase 2]  Lip Gloss theme definitions
|   +-- keys.go                         [Phase 2]  Key binding definitions
|
|-- go.mod                              [Phase 1]
|-- go.sum                              [Phase 1]
|-- Makefile                            [Phase 1]
|-- .goreleaser.yml                     [Phase 5]
+-- .github/workflows/ci.yml           [Phase 5]
```

### Import Dependency Rules

Every package has strict rules about what it may import. Violations should be caught by CI (e.g., `depguard` linter rule).

| Package | May Import | Must Not Import |
|---------|-----------|-----------------|
| `domain/` | Go stdlib only | Everything else |
| `internal/repo/interfaces.go` | `domain/` | Everything else |
| `internal/repo/fs/` | `domain/`, `internal/repo/` (interfaces), go-toml, fsnotify | `internal/service/`, `cmd/`, `tui/` |
| `internal/repo/mem/` | `domain/`, `internal/repo/` (interfaces) | `internal/repo/fs/`, `internal/service/` |
| `internal/validate/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/reconcile/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/generate/` | `domain/`, `internal/repo/` (interfaces) | `internal/service/`, `cmd/`, `tui/` |
| `internal/render/` | `domain/`, lipgloss | `internal/service/`, `internal/repo/` |
| `internal/service/` | `domain/`, `internal/repo/` (interfaces), `internal/validate/`, `internal/reconcile/`, `internal/generate/` | `cmd/`, `tui/`, `internal/repo/fs/` |
| `internal/mcp/` | `domain/`, `internal/service/` (interfaces) | `cmd/`, `tui/`, `internal/repo/` |
| `internal/watch/` | `domain/`, `internal/service/` (interfaces), fsnotify | `cmd/`, `tui/` |
| `internal/orchestrate/` | `domain/`, `internal/service/` (interfaces), `internal/repo/` (interfaces) | `cmd/`, `tui/` |
| `internal/project/` | `domain/`, Go stdlib | Everything else |
| `cmd/` | `domain/`, `internal/service/` (interfaces), `internal/render/`, cobra | `internal/repo/`, `tui/` |
| `tui/` | `domain/`, `internal/service/` (interfaces), bubbletea, lipgloss, bubbles | `cmd/`, `internal/repo/` |
| `main.go` | Everything (sole wiring point) | N/A |

**Key invariant**: `internal/repo/fs/` is only imported by `main.go` during dependency wiring. All other packages interact with repositories through interfaces defined in `internal/repo/interfaces.go`. This makes every service testable with in-memory implementations.

---

## 9. Testing Strategy

### Per-Layer Approach

Testing follows the architectural layers. Each layer has an appropriate testing strategy that maximizes coverage while minimizing brittleness.

#### Domain Layer --- Unit Tests

Pure unit tests with zero external dependencies. The domain package imports only Go stdlib, so tests are fast and deterministic.

**What to test**:
- Business rule functions (e.g., `BriefGate()` classification, `IsStub()` detection)
- Value object constructors and validation (e.g., `RequestType.Validate()`)
- State machine transitions (e.g., `WorkflowState` progression)
- Helper functions (e.g., `IterationDirName()`, `Zone.String()`)
- Error construction and sentinel error matching

**Example tests**:
```
TestIterationDirName                      "007", NEW_PROJECT, "rest-api" -> "007-NEW_PROJECT-rest-api"
TestBriefGate_MissingBrief                nil Brief -> BRIEF_MISSING
TestBriefGate_StubBrief                   stub Brief -> BRIEF_STUB
TestBriefGate_CompleteBrief               complete Brief -> BRIEF_PRESENT
TestRequestTypeFromKeyword_Create         "create" -> NEW_PROJECT
TestRequestTypeFromKeyword_Fix            "fix" -> BUG_FIX
TestIsStub_OnlyHeadingsAndComments        headings-only content -> true
TestIsStub_SubstantiveContent             real content -> false
TestZoneFromPath_Spec                     "docs/spec/foo.md" -> ZoneSpec
TestCycleDetection_NoCycle                acyclic graph -> nil
TestCycleDetection_DirectCycle            A->B->A -> error with path
TestCycleDetection_TransitiveCycle        A->B->C->A -> error with path
```

**Coverage target**: >80%

#### Service Layer --- Unit + Integration Tests

Unit tests inject `MemRepo` implementations (from `internal/repo/mem/`). Integration tests use real filesystem with temp directories.

**Unit tests** (fast, no I/O):
```
TestProjectHealth_AllZonesPresent         MemDocRepo with all zones -> healthy report
TestProjectHealth_MissingSpec             MemDocRepo missing spec zone -> degraded health
TestValidateDocsAllChecks_ValidProject    MemDocRepo valid project -> 17/17 pass
TestValidateDocsAllChecks_MissingFiles    MemDocRepo missing files -> specific failures
TestGenerateIteration_CorrectSequence     MemIterationRepo with 6 existing -> creates 007
TestReconcile_PropagatesStaleness         MemDocRepo with changed file -> downstream stale
```

**Integration tests** (slower, real filesystem):
```
TestProjectHealth_RealProject             temp dir with Mind structure -> correct health
TestDoctorFix_CreatesMissingDirs          temp dir missing dirs -> dirs created
TestCreateIteration_FilesOnDisk           temp dir -> iteration folder with 5 files
```

**Coverage target**: >70% (unit), >50% (integration)

#### Validation Suites --- Golden File Tests

Create fixture directories under `testdata/` with known documentation structures. Run validation suites and compare output to golden files.

**Fixture directories**:
```
testdata/
|-- valid-project/                       All docs present, no stubs -> 17/17 pass
|-- missing-docs/                        Missing required files -> specific failures
|-- stub-project/                        Stub documents -> warnings in non-strict, failures in strict
|-- broken-refs/                         Invalid cross-references -> refs suite failures
|-- invalid-config/                      Malformed YAML configs -> config suite failures
|-- stale-project/                       Outdated lock file -> reconciliation failures
+-- cycle-graph/                         Circular dependencies -> cycle detection error
```

**Golden file workflow**:
1. Run suite against fixture: `ValidationService.CheckDocs(fixtureDir)`
2. Serialize result to JSON
3. Compare against `testdata/{fixture}.golden.json`
4. Update golden files with `-update` flag when behavior intentionally changes

**Coverage target**: 100% of defined checks (every check ID exercised by at least one fixture)

#### Repository Layer --- Integration Tests

Repository implementations interact with the real filesystem. Tests create temp directories, populate them with fixtures, and verify read/write operations.

```
TestFSDocRepo_ListByZone                  temp dir with docs -> correct zone grouping
TestFSDocRepo_DetectStub                  temp dir with stub file -> IsStub() returns true
TestFSIterationRepo_Create                temp dir -> creates folder with correct naming
TestFSIterationRepo_ListChronological     temp dir with iterations -> correct ordering
TestFSConfigRepo_ParseMindToml            temp dir with mind.toml -> correct Config struct
TestFSLockRepo_RoundTrip                  write lock -> read lock -> identical content
TestFSLockRepo_InvalidJSON                corrupt lock file -> graceful error
```

**Coverage target**: >70%

#### CLI Layer --- Golden File Tests

Run commands via `cobra.Command.Execute()` (not as subprocess). Capture stdout/stderr and compare to golden files. Test all three output modes.

```
TestStatusCommand_Interactive             valid project -> styled output matches golden
TestStatusCommand_Plain                   valid project, no-color -> plain output matches golden
TestStatusCommand_JSON                    valid project, --json -> JSON matches golden
TestDoctorCommand_WithIssues              broken project -> diagnostic output matches golden
TestCheckAllCommand_Strict                stub project, --strict -> failures in output
TestCreateIterationCommand_Args           "new", "rest-api" -> correct folder created
TestVersionCommand_Short                  --short -> version string only
TestNotProjectError                       empty dir -> exit code 3, "not a Mind project" message
```

**Coverage target**: >60% (focused on argument parsing and output rendering, not business logic)

#### TUI Layer --- Model Unit Tests

Test `Update()` and `View()` functions with mock messages. Do not test visual rendering (Bubble Tea handles that).

```
TestAppModel_TabSwitching                 key "2" -> active tab changes to docs
TestAppModel_TabCycling                   Tab key -> cycles through tabs
TestStatusTab_RendersZoneBars             Health data -> View() contains zone names and counts
TestDocsTab_ZoneFilter                    filter to "spec" -> only spec docs in View()
TestDocsTab_Search                        search "brief" -> filtered list
TestChecksTab_ExpandSuite                 expand docs suite -> individual checks visible
TestQualityTab_NoData                     empty history -> "No quality data" in View()
```

**Coverage target**: >50% (focused on state transitions and data rendering, not layout)

#### Reconciliation --- Unit + Integration Tests

```
TestHashComputation_KnownInput            "hello\n" -> expected SHA-256 hex string
TestHashComputation_NormalizesLineEndings CRLF input -> same hash as LF input
TestHashComputation_StripsTrailingSpace   trailing spaces -> same hash as without
TestGraphBuild_FromEdges                  3 edges -> correct adjacency list
TestGraphBuild_CycleDetection             cycle in edges -> error with path
TestTopologicalSort_CorrectOrder          A->B->C -> [C, B, A]
TestStalePropagation_DirectDependency     A changes, B depends on A -> B stale
TestStalePropagation_Transitive           A changes, A->B->C -> B and C stale
TestStalePropagation_DiamondGraph         A->B, A->C, B->D, C->D -> all stale
TestMtimeFastPath_UnchangedFile           same mtime -> hash not recomputed
TestReconcileEngine_EndToEnd              temp dir with files and graph -> correct lock file
```

**Coverage target**: >80%

#### MCP Server --- Protocol Conformance Tests

```
TestMCPInitialize                         init request -> correct capabilities response
TestMCPToolsList                          list request -> 16 tools with schemas
TestMCPToolCall_Status                    mind_status call -> correct JSON response
TestMCPToolCall_ValidateDocs              mind_validate_docs -> matches check docs --json
TestMCPToolCall_CreateIteration           mind_create_iteration -> iteration created, path returned
TestMCPToolCall_UnknownTool               unknown_tool -> method not found error
TestMCPMalformedRequest                   invalid JSON -> parse error response
TestMCPBatchRequest                       batch of 3 calls -> 3 responses
```

**Coverage target**: >70%

#### E2E Tests --- Full Command Invocation

Run the `mind` binary as a subprocess against real project fixtures. Verify both stdout output and filesystem side effects.

```
TestE2E_InitThenStatus                    mind init -> mind status -> healthy project
TestE2E_InitThenCreateThenCheck           mind init -> mind create iteration -> mind check all -> pass
TestE2E_ReconcileDetectsChange            mind reconcile -> modify file -> mind reconcile --check -> exit 4
TestE2E_DoctorFixCreatesFiles             mind doctor --fix -> missing files created
TestE2E_JSONOutputIsValid                 mind status --json | jq . -> valid JSON
TestE2E_NotProjectError                   mind status in empty dir -> exit 3
```

**Coverage target**: Critical paths only (not percentage-based)

---

## 10. Build & CI

### Makefile

```makefile
.PHONY: build test lint coverage install clean

# Build binary with version info embedded
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o mind .

test:
	go test ./...

lint:
	golangci-lint run ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "---"
	@go tool cover -func=coverage.out | grep total:

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f mind coverage.out

# Development helpers
test-verbose:
	go test -v ./...

test-race:
	go test -race ./...

test-update-golden:
	go test ./... -update
```

### GoReleaser Configuration (`.goreleaser.yml`)

```yaml
version: 2

builds:
  - binary: mind
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "mind-{{ .Os }}-{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"

aurs:
  - name: mind-cli-bin
    homepage: "https://github.com/jf-ferraz/mind-cli"
    description: "CLI and TUI for the Mind Agent Framework"
    license: "MIT"
    maintainers:
      - "jf-ferraz"
    private_key: "{{ .Env.AUR_KEY }}"
    git_url: "ssh://aur@aur.archlinux.org/mind-cli-bin.git"
    package: |-
      install -Dm755 mind "${pkgdir}/usr/bin/mind"
      install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/mind-cli/LICENSE"

brews:
  - repository:
      owner: jf-ferraz
      name: homebrew-tap
    homepage: "https://github.com/jf-ferraz/mind-cli"
    description: "CLI and TUI for the Mind Agent Framework"
    license: "MIT"
    install: |-
      bin.install "mind"
```

### CI Pipeline (`.github/workflows/ci.yml`)

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out | grep total:

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o mind .
```

### Release Workflow (`.github/workflows/release.yml`)

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AUR_KEY: ${{ secrets.AUR_KEY }}
```

---

## 11. Definition of Done

Each phase has a completion checklist. ALL items must pass before the next phase begins.

### Phase 1: Core CLI

- [ ] All Phase 1 acceptance criteria (Section 2) pass
- [ ] `go test ./...` passes with >70% coverage on `domain/` and `internal/validate/`
- [ ] `go vet ./...` reports zero warnings
- [ ] `golangci-lint run ./...` passes clean
- [ ] All commands produce correct exit codes per BP-01 (0, 1, 3)
- [ ] `--json` output parses as valid JSON for every command that supports it
- [ ] "Not a Mind project" error works from any directory without `.mind/`
- [ ] Commands tested against at least 2 fixture projects (valid, incomplete)
- [ ] No data races (`go test -race ./...` passes)

### Phase 1.5: Reconciliation

- [ ] All Phase 1.5 acceptance criteria (Section 3) pass
- [ ] `go test ./internal/reconcile/...` passes with >80% coverage
- [ ] Integration with `mind status`, `mind check all`, and `mind doctor` verified
- [ ] Performance benchmark: 50 documents reconciled in <200ms
- [ ] Cycle detection tested with at least 3 graph topologies (no cycle, direct cycle, transitive cycle)
- [ ] `mind.lock` format documented and round-trip tested
- [ ] mtime fast-path verified (second run measurably faster than first)

### Phase 2: TUI Dashboard

- [ ] All Phase 2 acceptance criteria (Section 4) pass
- [ ] All 5 tabs render correctly at 80x24 terminal size (manual verification)
- [ ] All 5 tabs render correctly at 120x40 terminal size (manual verification)
- [ ] Terminal resize tested: no panic, no garbled output
- [ ] Tab state preserved across switches (scroll position, selection)
- [ ] Clean exit tested: terminal state restored correctly
- [ ] TUI model unit tests pass for all tab models

### Phase 3: AI Bridge (A+B)

- [ ] All Phase 3 acceptance criteria (Section 5) pass
- [ ] MCP server tested with Claude Code via `.mcp.json` configuration
- [ ] All 16 MCP tools return correct JSON responses
- [ ] Pre-flight/handoff workflows verified end-to-end with a real Mind project
- [ ] Classification accuracy tested: at least 5 request examples per type
- [ ] Business context gate blocks correctly for missing/stub briefs
- [ ] MCP error handling tested: malformed requests, unknown tools, internal errors

### Phase 4: AI Bridge (C+D)

- [ ] All Phase 4 acceptance criteria (Section 6) pass
- [ ] Watch mode tested with simulated AI workflow (files created/modified in sequence)
- [ ] Debounce verified: rapid file writes produce single event
- [ ] Full orchestration tested with `claude` CLI against a real project
- [ ] Retry logic tested: gate failure triggers re-dispatch with feedback
- [ ] Resume tested: interrupt mid-pipeline, resume from saved state
- [ ] Session splitting tested: pause after architect, resume to developer
- [ ] Error handling tested: `claude` CLI not found, non-zero exit, timeout

### Phase 5: Polish

- [ ] All Phase 5 acceptance criteria (Section 7) pass
- [ ] Cross-platform builds verified: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] CI pipeline green on main branch
- [ ] Release workflow tested: tag push produces GitHub release with artifacts
- [ ] Binary size < 15MB for all platforms
- [ ] Startup time < 50ms (measured with `hyperfine --warmup 3 './mind version'`)
- [ ] Shell completions verified: bash, zsh, fish
- [ ] `go vet ./...` and `golangci-lint run ./...` pass with zero warnings
- [ ] All E2E tests pass against release binary

---

## Cross-Reference Summary

This roadmap is the execution plan. It deliberately excludes:

- **Architectural decisions** --- see [BP-01: System Architecture](01-system-architecture.md)
- **Entity definitions and business rules** --- see [BP-02: Domain Model](02-domain-model.md)
- **Interface contracts and data flow** --- see [BP-03: Software Architecture](03-architecture.md)
- **Command specifications and UX** --- see [BP-04: Mind CLI & TUI](01-mind-cli.md)
- **AI integration model design** --- see [BP-05: AI Workflow Bridge](02-ai-workflow-bridge.md)

The roadmap references these blueprints for detail but does not duplicate their content. If a specification changes in another blueprint, this roadmap's acceptance criteria may need updating.
