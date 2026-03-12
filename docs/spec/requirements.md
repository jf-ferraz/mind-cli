# Requirements

## Overview

Phase 1 of mind-cli delivers a deterministic CLI that replaces the Mind Agent Framework's 9 bash scripts with a single Go binary. It provides project health diagnostics, 17-check documentation validation, 11-check cross-reference validation, document scaffolding for 6 artifact types, workflow state inspection, and three output modes (interactive, plain, JSON) -- all accessed through a unified `mind` command with `--json` support on every subcommand.

This document covers Phase 1 (Core CLI), Phase 1.5 (Reconciliation Engine), Phase 2 (TUI Dashboard), the pre-Phase 3 cleanup iteration, and the Phase 3 review and remediation iteration. Phase 1.5 requirements (FR-51 through FR-87) are appended below the Phase 1 section. Phase 2 requirements (FR-88 through FR-124) follow. Pre-Phase 3 cleanup requirements (FR-125 through FR-139) address interface consistency, type safety, test coverage, and documentation accuracy. Phase 3 review and remediation requirements (FR-140 through FR-151) address MCP protocol compliance, quality domain model alignment, test coverage for all Phase 3 packages, and architectural layer violation remediation.

## Scope Boundary

**In-scope (Phase 1)**:
- Domain types: Project, Config, Document, Zone, DocStatus, Brief, BriefGate, Iteration, WorkflowState, ValidationReport, CheckResult, Health
- Repository layer: DocRepo, IterationRepo, StateRepo, ConfigRepo, BriefRepo interfaces + filesystem implementations
- Service layer: ProjectService, ValidationService, GenerateService, WorkflowService
- Validation engine: 17 doc checks, 11 ref checks, config validation
- Rendering: Interactive (Lip Gloss), Plain, JSON output modes
- Document generation: ADR, blueprint, iteration, spike, convergence, brief templates
- Commands: status, init, doctor, create, docs, check, workflow, version, help
- Cross-cutting: --json flag, project root detection, mind.toml parsing, exit codes

**Out-of-scope**:
- MCP server, pre-flight, handoff (Phase 3)
- Watch mode, full orchestration (Phase 4)
- Shell completions (Phase 5)

**Delivered in later phases (requirements appended below)**:
- Reconciliation engine, mind.lock, staleness propagation (Phase 1.5, FR-51 through FR-87)
- TUI dashboard (Phase 2, FR-88 through FR-124)
- Pre-Phase 3 cleanup: interface migration, type safety, test coverage (FR-125 through FR-139)

## Functional Requirements

### Project Detection and Configuration

- **FR-1**: The CLI MUST detect a Mind Framework project by walking up from the current working directory looking for a `.mind/` directory. [MUST]
- **FR-2**: The CLI MUST accept a `--project-root` flag to override auto-detection and use an explicit project root path. [MUST]
- **FR-3**: The CLI MUST parse `mind.toml` with all sections: manifest, project, project.stack, project.commands, documents, governance, profiles. [MUST]
- **FR-4**: The CLI MUST return exit code 3 with an actionable error message when invoked outside a Mind project (no `.mind/` found) for commands that require a project. [MUST]
- **FR-5**: The CLI SHOULD operate in degraded mode when `mind.toml` is absent -- commands like `status` render what they can (zone scan, stub detection) with a warning that `mind.toml` is missing. [SHOULD]

### Output Modes

- **FR-6**: The CLI MUST auto-detect three output modes: interactive (stdout is a TTY), plain (stdout is piped or `--no-color`), JSON (`--json`). [MUST]
- **FR-7**: Every command that produces structured data MUST support `--json` output that conforms to the JSON schemas defined in BP-03. [MUST]
- **FR-8**: Interactive mode MUST use Lip Gloss styling with color and symbols (PASS, FAIL, WARN, SKIP). Output MUST NOT rely on color alone for meaning. [MUST]
- **FR-9**: Plain mode MUST produce clean text with no ANSI escape codes. [MUST]
- **FR-10**: JSON output MUST go to stdout. Errors and progress indicators MUST go to stderr. [MUST]

### Status Command

- **FR-11**: `mind status` MUST display: project name, root path, framework version, per-zone documentation health (present/total/stubs/complete counts for each of the 5 zones), workflow state (idle or active), last iteration summary, warnings list, and actionable suggestions. [MUST]
- **FR-12**: `mind status --json` MUST produce a `ProjectHealth` JSON object matching the BP-03 schema, including project metadata, brief gate result, zone health map, workflow state, last iteration, warnings array, and suggestions array. [MUST]
- **FR-13**: `mind status` MUST classify the project brief into one of three gate results: BRIEF_PRESENT, BRIEF_STUB, or BRIEF_MISSING. [MUST]

### Init Command

- **FR-14**: `mind init` MUST create the `.mind/` directory, `docs/` zone structure (5 zone directories), stub documents for required files (project-brief.md, requirements.md, architecture.md, domain-model.md, current.md, workflow.md, glossary.md, INDEX.md), and a `mind.toml` manifest with project metadata. [MUST]
- **FR-15**: `mind init` MUST create `.claude/CLAUDE.md` as an adapter that routes to `.mind/CLAUDE.md`. [MUST]
- **FR-16**: `mind init --name <name>` MUST use the provided name for `mind.toml [project].name`; otherwise it MUST fall back to the current directory name. [MUST]
- **FR-17**: `mind init --with-github` MUST additionally create `.github/agents/` with synced agent definitions. [SHOULD]
- **FR-18**: `mind init --from-existing` MUST detect existing `docs/` files, skip creating stubs for files that already exist, and register discovered files in `mind.toml`. [MUST]
- **FR-19**: `mind init` MUST abort with exit code 2 and a descriptive message when `.mind/` already exists. [MUST]

### Doctor Command

- **FR-20**: `mind doctor` MUST run all diagnostic checks: framework installation, adapter installations, 17-check doc validation, 11-check cross-reference validation, config validation, brief completeness, stub detection with severity, workflow state consistency, and iteration completeness. [MUST]
- **FR-21**: `mind doctor` MUST produce actionable remediation advice for every failing check (e.g., "Run: mind init --with-github"). [MUST]
- **FR-22**: `mind doctor --fix` MUST auto-fix resolvable issues: create missing directories, add `.gitkeep` files, create stub documents from templates, and fix naming convention violations. [MUST]
- **FR-23**: `mind doctor --fix` MUST report which fixes were applied and which failed (e.g., permission denied), exiting with code 1 on partial fix. [MUST]

### Create Commands

- **FR-24**: `mind create adr "<title>"` MUST create an auto-numbered ADR file in `docs/spec/decisions/` using the pattern `{NNN}-{slug}.md`, with sequence numbers derived from the highest existing ADR number + 1. [MUST]
- **FR-25**: `mind create blueprint "<title>"` MUST create an auto-numbered blueprint file in `docs/blueprints/` using the pattern `{NN}-{slug}.md` and append an entry to `docs/blueprints/INDEX.md`. [MUST]
- **FR-26**: `mind create iteration <type> "<name>"` MUST accept one of 4 types (new, enhancement, bugfix, refactor), map them to canonical names (NEW_PROJECT, ENHANCEMENT, BUG_FIX, REFACTOR), and create a directory `{NNN}-{TYPE}-{slug}/` with exactly 5 template files: overview.md, changes.md, test-summary.md, validation.md, retrospective.md. [MUST]
- **FR-27**: `mind create spike "<title>"` MUST create a spike report template at `docs/knowledge/{slug}-spike.md`. [MUST]
- **FR-28**: `mind create convergence "<title>"` MUST create a convergence analysis template at `docs/knowledge/{slug}-convergence.md`. [MUST]
- **FR-29**: `mind create brief` MUST launch interactive prompts (Vision, Key Deliverables, In Scope, Out of Scope, Constraints) and write a complete `docs/spec/project-brief.md` that passes the business context gate. [MUST]
- **FR-30**: All `create` commands MUST abort with an error when a file or directory with the target name already exists. [MUST]
- **FR-31**: Title slugification MUST: lowercase the input, replace spaces and non-alphanumeric characters with hyphens, strip leading and trailing hyphens. [MUST]

### Docs Commands

- **FR-32**: `mind docs list` MUST list all documents in `docs/` grouped by zone, showing path, stub status, modification date, and file size. [MUST]
- **FR-33**: `mind docs list --zone <zone>` MUST filter documents to only the specified zone. Invalid zone names MUST produce an error listing valid zones. [MUST]
- **FR-34**: `mind docs tree` MUST render a visual tree of all documentation files with zone-awareness and stub status annotations. [MUST]
- **FR-35**: `mind docs stubs` MUST list all documents classified as stubs, grouped by zone, with remediation hints. Exit code 0 when no stubs found, exit code 1 when stubs exist. [MUST]
- **FR-36**: `mind docs search "<query>"` SHOULD perform case-insensitive substring search across all `.md` files in `docs/`, returning matching lines with 1 line of context, grouped by file. [SHOULD]
- **FR-37**: `mind docs open <path-or-id>` SHOULD resolve a document by relative path, document ID (`doc:zone/name`), or fuzzy partial name match, and open it in `$EDITOR`. [SHOULD]

### Check Commands

- **FR-38**: `mind check docs` MUST run the 17-check documentation validation suite (zone directories exist, required files present, naming conventions, stub detection, brief completeness). [MUST]
- **FR-39**: `mind check docs --strict` MUST promote warnings to failures (stub documents become errors). [MUST]
- **FR-40**: `mind check refs` MUST run the 11-check cross-reference validation suite (mind.toml paths resolve, no orphan documents, INDEX.md matches blueprints, no broken internal markdown links, iteration overview.md presence, sequence number contiguity, state file references, .claude/CLAUDE.md references). [MUST]
- **FR-41**: `mind check config` MUST validate mind.toml against the schema rules defined in BP-03 (schema format, generation >= 1, project name kebab-case, valid zone names, document path format). [MUST]
- **FR-42**: `mind check all` MUST run all validation suites (docs, refs, config) and produce a unified `ValidationReport`. [MUST]
- **FR-43**: All check commands MUST exit with code 0 when all checks pass (warnings are acceptable) and exit code 1 when any check at FAIL level fails. [MUST]

### Workflow Commands

- **FR-44**: `mind workflow status` MUST parse `docs/state/workflow.md` and display the current workflow state: type, descriptor, last agent, remaining chain, session count. When idle, it MUST display "idle". [MUST]
- **FR-45**: `mind workflow history` MUST list all iterations chronologically with sequence number, type, descriptor, status, and artifact completeness ratio. [MUST]

### Version and Help

- **FR-46**: `mind version` MUST display version string, git commit SHA, build date, and Go version/platform. [MUST]
- **FR-47**: `mind version --short` SHOULD display only the version string. [SHOULD]
- **FR-48**: `mind help [command]` MUST display usage information for any command. Cobra auto-generates this. [MUST]

### Exit Codes

- **FR-49**: The CLI MUST use consistent exit codes across all commands: 0 (success), 1 (validation failure or issues found), 2 (runtime error), 3 (configuration error or not a Mind project). [MUST]

### Stub Detection

- **FR-50**: A document MUST be classified as a stub if it contains only headings, HTML comments, and template placeholder text with no substantive content. [MUST]

## Non-Functional Requirements

- **NFR-1**: `mind status` MUST complete in under 200ms for a project with 10-50 documents. [Measured by wall-clock time from command invocation to output completion]
- **NFR-2**: CLI startup time (root command initialization before subcommand execution) MUST be under 50ms. [Measured by running `mind version`]
- **NFR-3**: The compiled binary MUST be under 15MB for linux/amd64.
- **NFR-4**: The `domain/` package MUST have zero imports beyond the Go standard library. No `os`, `filepath`, `io`, or third-party packages.
- **NFR-5**: Test coverage MUST meet minimums: `domain/` >= 80%, `internal/validate/` >= 80%, overall >= 70%.
- **NFR-6**: The CLI MUST produce identical validation results to the existing bash scripts (`validate-docs.sh`, `validate-xrefs.sh`) for the same project state. Same check IDs, same pass/fail counts.
- **NFR-7**: All errors MUST be wrapped with context using `fmt.Errorf("operation: %w", err)`. No naked error returns.
- **NFR-8**: All exported functions MUST have GoDoc-style comments starting with the function name.
- **NFR-9**: The system MUST build and run on linux/amd64, linux/arm64, darwin/amd64, darwin/arm64.
- **NFR-10**: The system MUST use Go 1.23+ and follow `gofmt` formatting with zero exceptions.
- **NFR-11**: Every command MUST complete without network access. The CLI operates fully offline.

## Constraints

### Technology Constraints

- **C-1**: Language: Go 1.23+. No other language runtimes.
- **C-2**: CLI framework: Cobra v1.8+.
- **C-3**: TOML parsing: go-toml/v2 v2.2+.
- **C-4**: Styled output: Lip Gloss v0.10+.
- **C-5**: Single binary distribution with zero runtime dependencies.

### Architecture Constraints

- **C-6**: 4-layer architecture: Presentation (cmd/) -> Service (internal/service/) -> Domain (domain/) -> Infrastructure (internal/repo/). Dependencies flow down only.
- **C-7**: Domain layer has zero external imports. Pure Go structs, enums, validation functions.
- **C-8**: Presentation layer is thin: parse flags, call service, render result. No business logic in cmd/ files.
- **C-9**: All filesystem access goes through repository interfaces. Tests use in-memory implementations.
- **C-10**: No `init()` functions except for Cobra command registration in cmd/ files. All dependency wiring is explicit in main.go.
- **C-11**: No global mutable state. All state flows through constructor injection.
- **C-12**: No `panic()` in domain/ or internal/. Panics are reserved for truly unrecoverable startup failures in main.go.

### Business Rule Constraints

- **C-13**: The business context gate MUST block NEW_PROJECT and COMPLEX_NEW request types when the brief is missing or classified as a stub.
- **C-14**: A project is valid if and only if it contains a `.mind/` directory.
- **C-15**: Iteration directory naming MUST follow `{NNN}-{TYPE}-{slug}` pattern where NNN is zero-padded to 3 digits, TYPE is one of NEW_PROJECT/ENHANCEMENT/BUG_FIX/REFACTOR, and slug is kebab-case.
- **C-16**: mind.toml schema version MUST match `^mind/v\d+\.\d+$`. Unknown versions produce a warning, not an error.
- **C-17**: Project name in mind.toml MUST match `^[a-z][a-z0-9-]*$` (kebab-case).

## Acceptance Criteria

### Project Detection and Configuration

- **FR-1**: GIVEN a directory tree `/a/b/c/` where `/a/` contains `.mind/` WHEN `mind status` is run from `/a/b/c/` THEN the CLI detects `/a/` as the project root and shows status for that project.
- **FR-2**: GIVEN any directory WHEN `mind status --project-root /path/to/project` is run THEN the CLI uses `/path/to/project` as the project root without walking up.
- **FR-3**: GIVEN a valid `mind.toml` with all sections populated WHEN the CLI parses it THEN all fields are accessible: manifest.schema, manifest.generation, project.name, project.type, project.stack.language, project.commands.test, documents entries, governance.max-retries, profiles.active.
- **FR-4**: GIVEN a directory with no `.mind/` in the entire ancestor chain WHEN any project-requiring command is run THEN exit code is 3 and stderr contains "not a Mind project" and "Run 'mind init'".
- **FR-5**: GIVEN a project with `.mind/` but no `mind.toml` WHEN `mind status` is run THEN status renders zone health and stub detection with a warning "mind.toml not found".

### Output Modes

- **FR-6**: GIVEN stdout is a TTY WHEN `mind status` is run without flags THEN output contains ANSI color codes and styled formatting. GIVEN stdout is piped WHEN `mind status` is run THEN output contains no ANSI escape codes. GIVEN `--json` flag WHEN `mind status --json` is run THEN output is valid JSON.
- **FR-7**: GIVEN any command that produces structured data WHEN `--json` is passed THEN the output is a single valid JSON object parseable by `jq`.
- **FR-8**: GIVEN interactive mode WHEN validation results are shown THEN each check result uses both a symbol (checkmark, X, warning triangle) and a text label.
- **FR-9**: GIVEN `--no-color` flag or piped output WHEN any command runs THEN output contains zero ANSI escape sequences (no bytes in range 0x1B..0x9F followed by `[`).
- **FR-10**: GIVEN `mind check docs --json 2>/dev/null` WHEN run THEN stdout contains clean JSON with no progress text mixed in.

### Status Command

- **FR-11**: GIVEN a valid project with documents in all 5 zones WHEN `mind status` is run THEN output shows 5 zone lines each with present/total counts and a progress indicator.
- **FR-12**: GIVEN a valid project WHEN `mind status --json` is run THEN the JSON output contains keys: "project", "brief", "zones", "workflow", "last_iteration", "warnings", "suggestions" and each zone object contains "total", "present", "complete", "stubs".
- **FR-13**: GIVEN `docs/spec/project-brief.md` exists with Vision, Key Deliverables, and Scope sections filled WHEN `mind status` is run THEN brief gate shows BRIEF_PRESENT. GIVEN the file is a stub THEN gate shows BRIEF_STUB. GIVEN the file does not exist THEN gate shows BRIEF_MISSING.

### Init Command

- **FR-14**: GIVEN an empty directory WHEN `mind init` is run THEN the following exist: `.mind/`, `docs/spec/`, `docs/blueprints/`, `docs/state/`, `docs/iterations/`, `docs/knowledge/`, `docs/spec/project-brief.md`, `docs/state/current.md`, `docs/state/workflow.md`, `docs/blueprints/INDEX.md`, `mind.toml`.
- **FR-15**: GIVEN `mind init` completes THEN `.claude/CLAUDE.md` exists and contains a reference to `.mind/CLAUDE.md`.
- **FR-16**: GIVEN `mind init --name my-service` WHEN initialization completes THEN `mind.toml` contains `name = "my-service"`. GIVEN no `--name` flag and the current directory is named `cool-project` THEN `mind.toml` contains `name = "cool-project"`.
- **FR-17**: GIVEN `mind init --with-github` WHEN initialization completes THEN `.github/agents/` directory exists with agent definitions.
- **FR-18**: GIVEN a directory with existing `docs/spec/project-brief.md` containing real content WHEN `mind init --from-existing` is run THEN the existing file is preserved (not overwritten) and is registered in `mind.toml`.
- **FR-19**: GIVEN a directory already containing `.mind/` WHEN `mind init` is run THEN exit code is 2 and output says "already initialized".

### Doctor Command

- **FR-20**: GIVEN a project with mixed issues (missing adapter, stub documents, valid docs) WHEN `mind doctor` is run THEN output shows pass/fail/warning status for each diagnostic category with counts.
- **FR-21**: GIVEN a failing check "Copilot adapter not found" WHEN `mind doctor` reports it THEN the output includes "Run: mind init --with-github" as remediation.
- **FR-22**: GIVEN a project missing `docs/knowledge/` directory WHEN `mind doctor --fix` is run THEN the directory is created and the fix is reported in output.
- **FR-23**: GIVEN `mind doctor --fix` where one fix succeeds (create directory) and one fails (permission denied on a file) THEN output reports both the success and the failure, and exit code is 1.

### Create Commands

- **FR-24**: GIVEN `docs/spec/decisions/` contains `001-foo.md` and `002-bar.md` WHEN `mind create adr "Use PostgreSQL"` is run THEN `docs/spec/decisions/003-use-postgresql.md` is created with ADR template content including title, date, and "Proposed" status.
- **FR-25**: GIVEN `docs/blueprints/` contains `01-foo.md` and `03-bar.md` WHEN `mind create blueprint "Auth System"` is run THEN `docs/blueprints/04-auth-system.md` is created AND `docs/blueprints/INDEX.md` contains a new entry for it.
- **FR-26**: GIVEN `docs/iterations/` contains `001-NEW_PROJECT-initial/` WHEN `mind create iteration enhancement "add caching"` is run THEN `docs/iterations/002-ENHANCEMENT-add-caching/` is created containing exactly: overview.md, changes.md, test-summary.md, validation.md, retrospective.md.
- **FR-27**: GIVEN no existing spike files WHEN `mind create spike "Redis vs Memcached"` is run THEN `docs/knowledge/redis-vs-memcached-spike.md` is created with spike template content.
- **FR-28**: GIVEN no existing convergence files WHEN `mind create convergence "Auth Strategy"` is run THEN `docs/knowledge/auth-strategy-convergence.md` is created with convergence template content.
- **FR-29**: GIVEN interactive mode WHEN `mind create brief` is run and the user provides Vision, Key Deliverables, and Scope THEN `docs/spec/project-brief.md` is written and the business context gate returns BRIEF_PRESENT.
- **FR-30**: GIVEN `docs/spec/decisions/003-use-postgresql.md` already exists WHEN `mind create adr "Use PostgreSQL"` is run THEN the command aborts with an error (no overwrite).
- **FR-31**: GIVEN input "Use PostgreSQL (v15+)" WHEN slugified THEN result is "use-postgresql-v15".

### Docs Commands

- **FR-32**: GIVEN a project with 18 documents across 5 zones WHEN `mind docs list` is run THEN output shows all 18 documents grouped under their zone headers with status, date, and size columns.
- **FR-33**: GIVEN a project with documents in all zones WHEN `mind docs list --zone spec` is run THEN only spec zone documents appear. GIVEN `--zone invalid` THEN error message lists valid zone names: spec, blueprints, state, iterations, knowledge.
- **FR-34**: GIVEN a project with nested directories in docs/ WHEN `mind docs tree` is run THEN output uses tree characters (box-drawing) and annotates each file with stub status.
- **FR-35**: GIVEN a project with 3 stub documents WHEN `mind docs stubs` is run THEN all 3 are listed with their zone and path and exit code is 1. GIVEN no stubs THEN output says "No stubs found" and exit code is 0.
- **FR-36**: GIVEN documents containing the word "authentication" WHEN `mind docs search "authentication"` is run THEN matching lines are shown with file paths and line numbers, grouped by file.
- **FR-37**: GIVEN `$EDITOR` is set to `vim` WHEN `mind docs open brief` is run and "brief" uniquely matches `docs/spec/project-brief.md` THEN vim opens that file.

### Check Commands

- **FR-38**: GIVEN a valid project with all required files WHEN `mind check docs` is run THEN output shows 17 checks with pass/fail/warn status and a summary line.
- **FR-39**: GIVEN a project with 2 stub documents and no other failures WHEN `mind check docs` is run THEN exit code is 0 (stubs are warnings). WHEN `mind check docs --strict` is run THEN exit code is 1 (stubs become failures).
- **FR-40**: GIVEN a project where `mind.toml` references a non-existent file in `[documents]` WHEN `mind check refs` is run THEN check [1] "mind.toml paths resolve" fails with the non-existent path listed.
- **FR-41**: GIVEN a `mind.toml` with `project.name = "MY-PROJECT"` (uppercase) WHEN `mind check config` is run THEN a validation failure reports that name must be kebab-case.
- **FR-42**: GIVEN a project WHEN `mind check all --json` is run THEN the JSON output contains a "suites" array with entries for "docs", "refs", and "config", each containing their check results.
- **FR-43**: GIVEN all checks pass WHEN `mind check docs` is run THEN exit code is 0. GIVEN 1 FAIL-level check fails THEN exit code is 1.

### Workflow Commands

- **FR-44**: GIVEN `docs/state/workflow.md` contains workflow state with type=NEW_PROJECT, last_agent=architect WHEN `mind workflow status` is run THEN output shows type, descriptor, last agent, and remaining chain. GIVEN workflow.md is empty or idle THEN output shows "idle".
- **FR-45**: GIVEN 6 iterations exist WHEN `mind workflow history` is run THEN output lists all 6 with sequence numbers, types, descriptors, statuses, and artifact completeness ratios (e.g., "5/5").

### Version and Help

- **FR-46**: GIVEN the binary was built with ldflags setting Version, CommitSHA, BuildDate WHEN `mind version` is run THEN output contains all three values plus Go version and OS/architecture.
- **FR-47**: GIVEN `mind version --short` WHEN run THEN output is a single line containing only the version string.
- **FR-48**: GIVEN `mind help check` WHEN run THEN output shows the check command's usage, available subcommands, and flags.

### Exit Codes

- **FR-49**: GIVEN each exit scenario WHEN the command completes THEN: success returns 0, validation failure returns 1, runtime error returns 2, config/project error returns 3. Verified by `echo $?` after each command.

### Stub Detection

- **FR-50**: GIVEN a file containing only `# Title\n## Section\n<!-- placeholder -->\n` WHEN stub detection runs THEN it is classified as a stub. GIVEN a file containing `# Title\n\nThis project provides a REST API for...` THEN it is classified as not a stub.

---

## Phase 1.5: Reconciliation Engine

Phase 1.5 adds hash-based content tracking with staleness propagation through a dependency graph. It introduces the `mind reconcile` command, integrates staleness data into existing commands (`status`, `check all`, `doctor`), and manages the `mind.lock` file lifecycle. No new external dependencies -- SHA-256 is in Go's standard library.

### Reconcile Command

- **FR-51**: `mind reconcile` MUST compute SHA-256 hashes for all documents declared in `mind.toml [documents]`, build the dependency graph from `mind.toml [[graph]]`, detect changes by comparing hashes to `mind.lock`, propagate staleness downstream through the graph, and write the updated `mind.lock` file. [MUST]
- **FR-52**: `mind reconcile --check` MUST perform the same scan and staleness propagation as `mind reconcile` but MUST NOT write or modify `mind.lock`. It MUST exit with code 0 when all documents are clean and exit with code 4 when any documents are stale. [MUST]
- **FR-53**: `mind reconcile --force` MUST discard the existing `mind.lock` entirely, re-hash every declared document from scratch, clear all staleness flags, and write a new `mind.lock`. [MUST]
- **FR-54**: `mind reconcile --graph` MUST output an ASCII tree visualization of the dependency graph declared in `mind.toml [[graph]]`, showing document IDs as nodes and edge types as labels. Stale nodes MUST be visually annotated when staleness data exists in `mind.lock`. [MUST]
- **FR-55**: `mind reconcile` MUST support `--json` output producing a JSON object with changed document IDs, stale documents with reasons, missing documents, overall status (CLEAN/STALE/DIRTY), and stats. [MUST]
- **FR-56**: `mind reconcile` MUST require a valid `mind.toml` with a `[documents]` section. When `mind.toml` is missing or has no documents, the command MUST exit with code 3. [MUST]

### Hash Computation

- **FR-57**: The engine MUST compute SHA-256 hashes of raw file bytes with no content normalization. Hash format MUST be `sha256:{64-character lowercase hex digest}`. [MUST]
- **FR-58**: The engine MUST implement an mtime fast-path: skip hash computation when file mtime and size match the stored lock entry values. [MUST]
- **FR-59**: Hash edge cases: empty files produce the SHA-256 of empty input; binary files are hashed with a warning; symlinks are resolved to target content; files >10MB are hashed with a warning; unreadable files are marked MISSING with an error reason. [MUST]

### Dependency Graph

- **FR-60**: The engine MUST parse `[[graph]]` entries from `mind.toml` (from, to, type fields) and construct a directed adjacency list with forward and reverse edges. [MUST]
- **FR-61**: Three edge types MUST be supported: `informs` (produces "may be outdated"), `requires` (produces "prerequisite changed"), `validates` (produces "needs re-validation"). All three propagate staleness. [MUST]
- **FR-62**: The engine MUST detect cycles via DFS. Cycles MUST abort reconciliation with exit code 2 and report the full cycle path. [MUST]
- **FR-63**: All document IDs in `[[graph]]` entries MUST be validated against `[documents]`. Undeclared references MUST produce an error. [MUST]
- **FR-64**: When no `[[graph]]` entries exist, the engine MUST still hash and track documents without staleness propagation. [MUST]

### Staleness Propagation

- **FR-65**: Staleness MUST propagate downstream only (in the direction of graph edges). [MUST]
- **FR-66**: Staleness MUST propagate transitively with path information in the reason string. [MUST]
- **FR-67**: Staleness propagation MUST enforce a depth limit of 10 levels with a warning when reached. [MUST]
- **FR-68**: Documents that changed (new hash != old hash) MUST NOT be marked as stale. Changed and stale are mutually exclusive. [MUST]
- **FR-69**: Documents already marked stale via one path MUST NOT be re-processed via another path. [MUST]

### Lock File Lifecycle

- **FR-70**: The lock file MUST be located at `mind.lock` in the project root. [MUST]
- **FR-71**: The lock file MUST be JSON containing: generated_at, status (CLEAN/STALE/DIRTY), stats, and entries keyed by document ID with id, path, hash, size, mod_time, stale, stale_reason, is_stub, and status fields. [MUST]
- **FR-72**: The lock file MUST survive round-trip (read, parse, write produces byte-identical output). [MUST]
- **FR-73**: Lock file writes MUST be atomic (write to temp file, then rename). [MUST]
- **FR-74**: When `mind.lock` does not exist, reconciliation MUST treat it as first run with no staleness. [MUST]
- **FR-75**: Lock entries MUST include `is_stub` computed via `DocRepo.IsStub()`, not reimplemented. [MUST]
- **FR-76**: Lock file status: STALE when any stale, DIRTY when missing but no stale, CLEAN otherwise. [MUST]

### Integration with Existing Commands

- **FR-77**: `mind status` MUST display a staleness panel when `mind.lock` exists with stale documents. Omit when no lock file. Do not trigger reconciliation. [MUST]
- **FR-78**: `mind status --json` MUST include a `staleness` object when `mind.lock` exists, null otherwise. [MUST]
- **FR-79**: `mind check all` MUST include a ReconcileSuite with checks for cycle detection, missing documents, and stale documents (WARN normally, FAIL with --strict). [MUST]
- **FR-80**: `mind check all --json` MUST include a "reconcile" entry in the suites array. [MUST]
- **FR-81**: `mind doctor` MUST report stale documents as WARN-level findings with remediation suggesting `mind reconcile --force`. [MUST]

### Exit Codes (Extended)

- **FR-82**: Exit code 4 MUST represent staleness detection, used exclusively by `mind reconcile --check`. Exit codes 0, 1, 2, 3 are unchanged. [MUST]

### Config Extension

- **FR-83**: `mind.toml` MUST support `[[graph]]` array-of-tables with `from`, `to`, and `type` string fields. [MUST]
- **FR-84**: `mind check config` MUST validate `[[graph]]` entries: document ID format and valid edge types. Invalid entries MUST produce FAIL. [MUST]

### Undeclared File Detection

- **FR-85**: Reconciliation MUST detect files in `docs/` not declared in `mind.toml [documents]` and report them as warnings. Undeclared files MUST NOT participate in staleness propagation. [MUST]

### Performance

- **FR-86**: Full reconciliation MUST complete in under 200ms for 50 documents. [MUST]
- **FR-87**: Incremental reconciliation MUST complete in under 50ms for 50 documents with 1 change. [MUST]

### Phase 1.5 Acceptance Criteria

- **FR-51**: GIVEN a project with 5 documents and 3 graph edges and no `mind.lock` WHEN `mind reconcile` is run THEN `mind.lock` is created with all entries having stale=false.
- **FR-52**: GIVEN `mind.lock` exists and `requirements.md` has changed and `architecture.md` depends on it WHEN `mind reconcile --check` is run THEN exit code is 4, `architecture.md` is listed as stale, and `mind.lock` is unchanged.
- **FR-53**: GIVEN `mind.lock` with 2 stale entries WHEN `mind reconcile --force` is run THEN `mind.lock` is rewritten with all entries stale=false and fresh hashes.
- **FR-54**: GIVEN edges brief --> requirements --> architecture WHEN `mind reconcile --graph` is run THEN output shows an ASCII tree with the dependency hierarchy and edge type labels.
- **FR-55**: GIVEN 1 changed and 2 stale documents WHEN `mind reconcile --json` is run THEN JSON contains status "STALE" and stats.stale equals 2.
- **FR-56**: GIVEN a project with no `mind.toml` WHEN `mind reconcile` is run THEN exit code is 3 with "mind.toml required" in stderr.
- **FR-57**: GIVEN a file with `\r\n` line endings WHEN hashed THEN the hash includes the `\r` bytes (no normalization).
- **FR-58**: GIVEN 50 documents with 49 unchanged mtime WHEN `mind reconcile` runs THEN exactly 1 hash computation occurs.
- **FR-59**: GIVEN an empty file WHEN hashed THEN result is `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
- **FR-60**: GIVEN 4 `[[graph]]` entries WHEN the graph is built THEN it contains 4 forward edges, 4 reverse edges, and the correct node set.
- **FR-61**: GIVEN an edge A --(requires)--> B where A changes WHEN propagation runs THEN B's stale reason contains "prerequisite changed".
- **FR-62**: GIVEN edges A --> B --> C --> A WHEN `mind reconcile` runs THEN exit code is 2 with the full cycle path in stderr.
- **FR-63**: GIVEN a `[[graph]]` entry referencing `doc:spec/nonexistent` not in `[documents]` WHEN `mind reconcile` runs THEN exit code is 3 with error about undeclared document.
- **FR-64**: GIVEN 5 documents and zero `[[graph]]` entries WHEN 1 document changes between reconciliation runs THEN it is reported as changed with no stale documents.
- **FR-65**: GIVEN A --> B where B changes WHEN propagation runs THEN A is not stale.
- **FR-66**: GIVEN A --> B --> C where A changes WHEN propagation runs THEN both B and C are stale with transitive path in C's reason.
- **FR-67**: GIVEN a chain A1 --> A2 --> ... --> A12 where A1 changes WHEN propagation runs THEN A2-A11 are stale, A12 is not, and a depth limit warning is emitted.
- **FR-68**: GIVEN A --> B where both A and B changed WHEN propagation runs THEN neither is marked stale.
- **FR-69**: GIVEN a diamond A --> B, A --> C, B --> D, C --> D where A changes WHEN propagation runs THEN D is stale exactly once.
- **FR-70**: GIVEN project at `/path/to/project` WHEN `mind reconcile` runs THEN `mind.lock` is at `/path/to/project/mind.lock`.
- **FR-71**: GIVEN `mind.lock` exists WHEN parsed THEN it contains generated_at, status, stats, and entries with all specified fields.
- **FR-72**: GIVEN `mind.lock` content WHEN read, parsed, serialized THEN output is byte-identical to input.
- **FR-73**: GIVEN reconciliation writes `mind.lock` THEN it writes to `mind.lock.tmp` first and renames atomically.
- **FR-74**: GIVEN no `mind.lock` WHEN `mind reconcile` runs THEN lock is created with all stale=false and status CLEAN.
- **FR-75**: GIVEN 2 stub and 3 non-stub documents WHEN reconciliation runs THEN stub entries have is_stub=true.
- **FR-76**: GIVEN 0 stale and 0 missing THEN status is CLEAN. GIVEN 0 stale and 1 missing THEN status is DIRTY. GIVEN 2 stale THEN status is STALE.
- **FR-77**: GIVEN `mind.lock` with 2 stale entries WHEN `mind status` runs THEN output includes staleness section. GIVEN no `mind.lock` WHEN `mind status` runs THEN no staleness section.
- **FR-78**: GIVEN `mind.lock` with status STALE WHEN `mind status --json` runs THEN JSON contains staleness object. GIVEN no `mind.lock` THEN staleness is null.
- **FR-79**: GIVEN 2 stale documents WHEN `mind check all` runs THEN reconcile suite shows 2 WARN checks and exit is 0. WHEN `mind check all --strict` runs THEN exit is 1.
- **FR-80**: GIVEN `[[graph]]` entries exist WHEN `mind check all --json` runs THEN suites array contains a "reconcile" entry.
- **FR-81**: GIVEN `mind.lock` with 2 stale entries WHEN `mind doctor` runs THEN 2 WARN findings appear with "mind reconcile --force" remediation.
- **FR-82**: GIVEN stale documents WHEN `mind reconcile --check` runs THEN exit is 4. GIVEN no stale THEN exit is 0.
- **FR-83**: GIVEN `[[graph]]` with from/to/type fields WHEN config is parsed THEN Config.Graph contains correct GraphEdge structs.
- **FR-84**: GIVEN `[[graph]]` with `type = "invalid"` WHEN `mind check config` runs THEN a FAIL-level result reports invalid edge type.
- **FR-85**: GIVEN `docs/spec/notes.md` exists but is not in `[documents]` WHEN `mind reconcile` runs THEN a warning about the undeclared file appears.
- **FR-86**: GIVEN 50 documents and 40 edges WHEN `mind reconcile --force` runs THEN completion time is under 200ms.
- **FR-87**: GIVEN 50 documents with 1 changed WHEN `mind reconcile` runs THEN completion time is under 50ms.

---

## Phase 2: TUI Dashboard

Phase 2 adds the `mind tui` command -- a full-screen interactive Bubble Tea dashboard with 5 tabs (Status, Documents, Iterations, Checks, Quality). It also resolves 4 SHOULD items from Phase 1.5 that directly impact TUI quality. No changes to existing CLI commands. New dependencies: Bubble Tea, Bubbles, Glamour.

### SHOULD Fixes (Pre-TUI Prerequisites)

- **FR-88**: `mind reconcile` MUST reject the combination of `--check` and `--force` flags with an error message and exit code 2. [MUST]
- **FR-89**: `ReconcileSuite` MUST include a check for documents declared in `mind.toml [documents]` that are missing from disk. Each missing document MUST produce a WARN-level CheckResult entry. [MUST]
- **FR-90**: All repository and service construction MUST be centralized in a `buildDeps()` function (or equivalent) in `main.go` callable from both CLI and TUI paths. Command handlers MUST NOT construct repositories or services directly. [MUST]
- **FR-91**: `mind docs search` MUST perform file discovery and content reading through the DocRepo interface instead of direct `filepath.WalkDir` calls. [MUST]

### TUI Command

- **FR-92**: `mind tui` MUST launch a full-screen interactive Bubble Tea application that takes over the terminal, loads project data through existing service interfaces, and renders a 5-tab dashboard. [MUST]

### App Shell

- **FR-93**: The TUI MUST render a title bar showing: "Mind Framework" label, project name from `mind.toml`, current git branch (if available), and framework version. [MUST]
- **FR-94**: The TUI MUST render a tab bar with 5 tabs: [1 Status], [2 Docs], [3 Iterations], [4 Check], [5 Quality]. Active tab is bold+underline. Tabs switchable via number keys (1-5), Tab (next, wrapping), Shift+Tab (previous, wrapping). [MUST]
- **FR-95**: Global keys MUST work in every tab: 1-5 (switch tab), Tab/Shift+Tab (cycle), q (quit when no modal), Ctrl+C (force quit), ? (help overlay), r (refresh when no text input focused). [MUST]
- **FR-96**: The TUI MUST render a bottom status bar showing context-sensitive key hints for the active tab and cursor position information. [MUST]

### Tab 1: Status

- **FR-97**: Tab 1 MUST display zone health as progress bars for all 5 zones with zone label, visual progress, and numeric fraction (present/total). [MUST]
- **FR-98**: Tab 1 SHOULD display a staleness section when `mind.lock` exists with stale documents, listing stale document names with `●` bullets. [SHOULD]
- **FR-99**: Tab 1 MUST display workflow state: running (type, agent chain, branch) or idle (last iteration). [MUST]
- **FR-100**: Tab 1 MUST display warnings (`⚠` prefix) and suggestions (`→` prefix) from ProjectHealth. Sections appear only when they have content. [MUST]
- **FR-101**: Tab 1 MUST use two-column layout at widths >= 80 (left: health/staleness/warnings/suggestions, right: workflow/quick actions). SHOULD stack vertically below 80. [MUST]

### Tab 2: Documents

- **FR-102**: Tab 2 MUST display all documents in a navigable list grouped by zone with filename, status indicator, status label, modification date, and file size. Cursor row highlighted. [MUST]
- **FR-103**: Tab 2 MUST provide zone filter bar with shortcuts: a (all), s (spec), b (blueprints), t (state), i (iterations), k (knowledge). Active filter visually indicated. [MUST]
- **FR-104**: Tab 2 MUST support inline search via `/` with real-time case-insensitive filtering. Esc clears search. [MUST]
- **FR-105**: Tab 2 SHOULD support a preview pane (Enter to toggle) rendering markdown via Glamour. List shrinks to 40%, preview takes 60%. [SHOULD]
- **FR-106**: Tab 2 SHOULD support opening selected document in `$EDITOR` via `e`, suspending TUI during editing. [SHOULD]

### Tab 3: Iterations

- **FR-107**: Tab 3 MUST display iterations in a table with columns: #, type, name, status, date, files (artifact completeness). Navigable with arrow keys/j/k. [MUST]
- **FR-108**: Tab 3 MUST provide type filter bar: a (all), n (NEW_PROJECT), e (ENHANCEMENT), b (BUG_FIX), r (REFACTOR). [MUST]
- **FR-109**: Tab 3 SHOULD support expanding selected iteration (Enter) to show inline artifact details. [SHOULD]

### Tab 4: Checks

- **FR-110**: Tab 4 MUST display validation results as an accordion with sections per suite. Headers show expand/collapse indicator, suite name, check count, pass/fail/warn summary. [MUST]
- **FR-111**: Tab 4 MUST run validation automatically on first activation with loading spinner. Re-run via `r`. [MUST]
- **FR-112**: Tab 4 SHOULD support detail pane toggle (Space) for individual checks showing file, issue, and fix. [SHOULD]
- **FR-113**: Tab 4 MUST display overall summary bar with aggregated pass/fail/warn counts. [MUST]

### Tab 5: Quality

- **FR-114**: Tab 5 SHOULD display ASCII line chart of convergence scores with Y-axis (1.0-5.0), dates, data points, Gate 0 line. Navigate points with ←/→. [SHOULD]
- **FR-115**: Tab 5 MUST display empty state message when no quality-log.yml exists or is empty. [MUST]
- **FR-116**: Tab 5 SHOULD display selected data point details: topic, variant, gate result, 6 dimension bars, personas, output path. [SHOULD]

### Help Overlay

- **FR-117**: `?` MUST toggle a centered help overlay listing global and tab-specific keybindings. Closes on `?` or Esc. While open, `q` closes overlay (does not quit). [MUST]

### Responsive Design

- **FR-118**: The TUI MUST require minimum 80x24 terminal. Below minimum, display "Terminal too small" message instead of full interface. [MUST]
- **FR-119**: The TUI MUST handle terminal resize gracefully -- recalculate layout, re-render without crash, toggle "too small" message as needed. [MUST]

### Data Loading

- **FR-120**: The TUI MUST load data asynchronously. Each view MUST display Loading/Error/Empty/Ready states. Loading MUST NOT block the UI thread. [MUST]
- **FR-121**: `r` (when no text input focused) MUST re-load all data. Existing data remains visible during refresh. [MUST]
- **FR-122**: Validation MUST NOT run on init. It MUST run lazily on first Tab 4 activation or `r` on Tab 4. [MUST]

### Tab State and Terminal

- **FR-123**: Tab switching MUST preserve each tab's local state: cursor position, scroll position, filter selection, expanded items, search query. [MUST]
- **FR-124**: On exit (q, Ctrl+C, or error), the TUI MUST restore terminal state: cursor visible, input echoing, alternate screen exited. [MUST]

### Phase 2 Non-Functional Requirements

- **NFR-12**: `mind tui` MUST launch (first frame) in under 500ms for a 50-document project. [MUST]
- **NFR-13**: Tab switching MUST complete in under 50ms. [MUST]
- **NFR-14**: TUI SHOULD use no more than 50MB memory during normal operation. [SHOULD]
- **NFR-15**: `tui/` Update() test coverage SHOULD be >= 60%. View() excluded from coverage. [SHOULD]

### Phase 2 Acceptance Criteria

- **FR-88**: GIVEN `mind reconcile --check --force` WHEN invoked THEN exit code is 2 with "mutually exclusive" error on stderr.
- **FR-89**: GIVEN 5 documents declared with 1 missing from disk WHEN `mind check all` runs THEN reconcile suite contains a WARN for the missing document.
- **FR-90**: GIVEN `mind tui` WHEN it initializes THEN it receives all services through the same wiring function used by CLI commands.
- **FR-91**: GIVEN `mind docs search "auth"` WHEN run THEN results are produced through DocRepo, not direct filepath.WalkDir.
- **FR-92**: GIVEN a valid Mind project WHEN `mind tui` runs THEN a full-screen application launches. GIVEN `q` is pressed THEN it exits cleanly.
- **FR-93**: GIVEN project "mind-cli" on branch "main" WHEN the TUI renders THEN the title bar shows project name, branch, and version.
- **FR-94**: GIVEN Tab 1 is active WHEN user presses `3` THEN Tab 3 becomes active. GIVEN Tab 5 WHEN user presses Tab THEN Tab 1 activates (wrap).
- **FR-95**: GIVEN Tab 2 with search input focused WHEN user presses `r` THEN `r` is typed (not refresh). GIVEN help overlay open WHEN user presses `q` THEN overlay closes.
- **FR-96**: GIVEN Tab 2 with 22 documents and cursor on row 3 WHEN status bar renders THEN it shows "3/22 docs".
- **FR-97**: GIVEN 4/5 spec documents WHEN Tab 1 renders THEN spec zone progress bar shows correct fraction.
- **FR-98**: GIVEN `mind.lock` with 2 stale entries WHEN Tab 1 renders THEN staleness section appears. GIVEN no lock file THEN no staleness section.
- **FR-99**: GIVEN active workflow type NEW_PROJECT WHEN Tab 1 renders THEN workflow panel shows running state. GIVEN idle THEN shows "State: idle".
- **FR-100**: GIVEN 2 warnings and 1 suggestion WHEN Tab 1 renders THEN both sections appear. GIVEN 0 warnings THEN warnings section is omitted.
- **FR-101**: GIVEN width 100 WHEN Tab 1 renders THEN two columns side by side. GIVEN width 78 THEN single column.
- **FR-102**: GIVEN 18 documents WHEN Tab 2 renders THEN all appear grouped by zone. GIVEN stub document THEN shows `✗` indicator.
- **FR-103**: GIVEN all documents shown WHEN user presses `s` THEN only spec documents appear. GIVEN `a` THEN all zones shown.
- **FR-104**: GIVEN Tab 2 WHEN user presses `/` and types "brief" THEN only matching documents appear. GIVEN Esc THEN search clears.
- **FR-105**: GIVEN document selected WHEN Enter pressed THEN preview pane opens with rendered markdown. GIVEN Esc THEN preview closes.
- **FR-106**: GIVEN `$EDITOR=vim` WHEN `e` pressed THEN vim opens; TUI resumes after exit.
- **FR-107**: GIVEN 6 iterations WHEN Tab 3 renders THEN table shows all 6 with correct columns. GIVEN iteration with 4/5 artifacts THEN shows "4/5".
- **FR-108**: GIVEN Tab 3 WHEN user presses `e` THEN only ENHANCEMENT iterations shown. GIVEN `a` THEN all types shown.
- **FR-109**: GIVEN iteration selected WHEN Enter pressed THEN inline artifact details expand. GIVEN Enter again THEN collapses.
- **FR-110**: GIVEN 4 validation suites WHEN Tab 4 renders THEN 4 accordion headers appear. GIVEN Enter on header THEN expands to show checks.
- **FR-111**: GIVEN first visit to Tab 4 WHEN tab activates THEN spinner appears and validation runs. GIVEN `r` pressed THEN validation re-runs.
- **FR-112**: GIVEN failed check selected WHEN Space pressed THEN detail pane shows file, issue, fix.
- **FR-113**: GIVEN 30/32 pass, 1 fail, 1 warn WHEN Tab 4 renders THEN summary bar shows correct counts.
- **FR-114**: GIVEN 5 quality entries WHEN Tab 5 renders THEN ASCII chart with 5 points appears. GIVEN `→` THEN next point selected.
- **FR-115**: GIVEN no quality-log.yml WHEN Tab 5 renders THEN empty state message appears.
- **FR-116**: GIVEN selected data point WHEN Tab 5 renders THEN topic, variant, gate, and 6 dimension bars appear.
- **FR-117**: GIVEN Tab 2 active WHEN `?` pressed THEN help overlay shows Docs-specific keys. GIVEN Esc THEN overlay closes.
- **FR-118**: GIVEN terminal 79x24 WHEN TUI renders THEN "Terminal too small" message. GIVEN 80x24 THEN full dashboard.
- **FR-119**: GIVEN TUI at 120x40 WHEN resized to 80x24 THEN layout adapts. GIVEN resized to 70x20 THEN "too small" message. GIVEN back to 90x30 THEN dashboard resumes.
- **FR-120**: GIVEN TUI launches WHEN data loading THEN spinner shown. GIVEN load error THEN error message with retry hint.
- **FR-121**: GIVEN stale data on Tab 1 WHEN `r` pressed THEN data reloads. Existing data visible during refresh.
- **FR-122**: GIVEN TUI launches on Tab 1 WHEN no interaction THEN no validation runs. GIVEN switch to Tab 4 THEN validation starts.
- **FR-123**: GIVEN Tab 2 cursor on row 5 with spec filter WHEN switch to Tab 3 and back THEN cursor on row 5, spec filter active.
- **FR-124**: GIVEN TUI running WHEN `q` pressed THEN terminal restored (cursor visible, input echoing, alternate screen exited).

---

## Phase 2.5: Pre-Phase 3 Cleanup

Pre-Phase 3 cleanup addresses architectural debt, type safety gaps, testing coverage, and documentation accuracy identified by convergence analysis (score 4.3/5.0). One MUST-fix item blocks Phase 3 (MCP server): the Deps struct uses concrete filesystem types instead of repository interfaces. Eight SHOULD-fix items improve code quality, testability, and consistency. No new CLI commands. No new external dependencies.

### Interface/Type Consistency

- **FR-125**: The `deps.Deps` struct (`internal/deps/deps.go`) MUST declare all repository fields using `repo.` interface types (`repo.DocRepo`, `repo.IterationRepo`, `repo.BriefRepo`, `repo.ConfigRepo`, `repo.LockRepo`, `repo.StateRepo`, `repo.QualityRepo`) instead of concrete `*fs.` types. The `Build()` function MUST still construct `fs.` implementations but return them through interface fields. [MUST]
- **FR-126**: The `internal/repo/mem/` package MUST NOT import `internal/repo/fs`. Any shared functionality (such as `IsStubContent()`) MUST be relocated to a package that both `fs/` and `mem/` can import without creating inverse dependencies. [SHOULD]
- **FR-137**: After Deps migration, `tui/` MUST access repositories exclusively through `Deps` interface fields. Direct imports of `internal/repo/fs` from `tui/` MUST be removed. [SHOULD]

### Staleness Propagation Accuracy

- **FR-127**: The `buildReason()` function in `internal/reconcile/propagate.go` MUST produce edge-type-specific reason strings at ALL propagation depths. At depth > 0, the reason MUST reflect the edge type between the immediate predecessor and the target document, not the edge from the original source. [SHOULD]

### CLI Flag Rename

- **FR-128**: The `--project` global flag (`cmd/root.go`) MUST be renamed to `--project-root`. The short flag `-p` MAY be retained. All internal references MUST be updated. [SHOULD]

### Exit Code Architecture

- **FR-129**: All `cmd/` command handlers MUST return errors to Cobra instead of calling `os.Exit()` directly. A structured error type carrying an exit code MUST be used so `Execute()` maps errors to exit codes. Dead code after `os.Exit()` MUST be removed. [SHOULD]
- **FR-130**: `Diagnostic.Status` in `domain/health.go` MUST use a typed `DiagnosticStatus` enum instead of raw `string`. JSON serialization MUST produce unchanged values (`"pass"`, `"fail"`, `"warn"`). [SHOULD]

### Test Coverage

- **FR-131**: `cmd/` MUST have exit code tests using `cobra.Command.Execute()`. At minimum: `check docs`, `check refs`, `check config`, `check all`, `reconcile`, `status`, `doctor`. [SHOULD]
- **FR-132**: `internal/render/` MUST have JSON output tests for `RenderHealth()`, `RenderValidation()`, `RenderReconcileResult()`, `RenderDoctorReport()`. [SHOULD]

### Documentation Accuracy

- **FR-133**: `docs/spec/architecture.md` MUST fix stale references: `cmd/tui_cmd.go` to `cmd/tui.go`, add `InitService`/`DoctorService` to component map, update `--project` to `--project-root`. [SHOULD]
- **FR-134**: `docs/spec/requirements.md` overview MUST be updated to acknowledge multi-phase scope (Phase 1, 1.5, 2, 2.5). [SHOULD]
- **FR-135**: `docs/spec/domain-model.md` MUST include `DiagnosticStatus` in the supporting types table and DC-3 constraint. [SHOULD]
- **FR-136**: `docs/state/current.md` MUST remove resolved issues and add iteration 004 to Recent Changes. [SHOULD]

### Verification

- **FR-138**: After all changes, `go vet ./...` MUST report zero issues and `go build ./...` MUST succeed. [MUST]
- **FR-139**: All pre-existing 374 tests MUST pass. No test may be deleted. Tests may be modified only for interface or flag name changes. [MUST]

### Pre-Phase 3 Cleanup Acceptance Criteria

- **FR-125**: GIVEN `internal/deps/deps.go` WHEN inspected THEN all 7 repository fields use `repo.` interface types. GIVEN `tui/app.go` imports THEN `internal/repo/fs` does NOT appear. GIVEN `go build ./...` THEN success.
- **FR-126**: GIVEN `go list -f '{{.Imports}}' ./internal/repo/mem/` WHEN run THEN `internal/repo/fs` does NOT appear. GIVEN `go test ./...` THEN all pass.
- **FR-127**: GIVEN chain A --(requires)--> B --(informs)--> C where A changes WHEN propagation runs THEN B's reason contains "prerequisite changed" AND C's reason reflects the B-to-C `informs` edge type.
- **FR-128**: GIVEN `mind status --project-root /tmp/p` WHEN invoked THEN `/tmp/p` is used as root. GIVEN `mind --help` THEN `--project-root` appears in global flags.
- **FR-129**: GIVEN `cmd/` source files WHEN searched for `os.Exit(` THEN zero matches. GIVEN same error conditions THEN same exit codes produced via error returns.
- **FR-130**: GIVEN `domain/health.go` WHEN inspected THEN `Diagnostic.Status` is `DiagnosticStatus`. GIVEN `mind doctor --json` THEN status values unchanged.
- **FR-131**: GIVEN `go test ./cmd/` WHEN run THEN tests exist and pass. GIVEN check command with FAIL-level check THEN error carries exit code 1.
- **FR-132**: GIVEN `go test ./internal/render/` WHEN run THEN tests exist and pass. GIVEN JSON mode render THEN output is valid JSON with correct field names.
- **FR-133**: GIVEN `docs/spec/architecture.md` WHEN searched for `cmd/tui_cmd.go` THEN zero matches. GIVEN component map THEN `InitService` and `DoctorService` rows exist.
- **FR-134**: GIVEN `docs/spec/requirements.md` overview WHEN read THEN it acknowledges Phase 1, 1.5, 2, and 2.5 scope.
- **FR-135**: GIVEN `docs/spec/domain-model.md` supporting types WHEN inspected THEN `DiagnosticStatus` row exists. GIVEN DC-3 THEN it lists `DiagnosticStatus`.
- **FR-136**: GIVEN `docs/state/current.md` WHEN inspected THEN resolved SHOULD items are removed and iteration 004 entry exists.
- **FR-137**: GIVEN `tui/` imports WHEN inspected THEN `internal/repo/fs` does NOT appear. GIVEN `mind tui` THEN behavior unchanged.
- **FR-138**: GIVEN `go vet ./...` WHEN run THEN zero issues. GIVEN `go build ./...` THEN success. GIVEN `go test ./...` THEN all pass.
- **FR-139**: GIVEN `go test ./...` WHEN run THEN test count >= 374 + new tests. GIVEN any modified test THEN changes limited to type/flag updates.

---

## Phase 3 Review and Remediation (FR-140 through FR-151)

Phase 3 (iteration 005, implemented outside the Mind Framework process) delivered `mind preflight`, `mind handoff`, `mind serve` (MCP server), `internal/orchestrate/`, and `internal/mcp/`. A convergence analysis (`docs/knowledge/phase-3-review-convergence.md`) identified MUST, SHOULD, and COULD violations. This section defines requirements to remediate them. Traceable to findings M-1, M-2, M-3, S-1 through S-6, C-1, C-4.

### MCP Protocol Compliance

- **FR-140**: The MCP server (`internal/mcp/server.go`) MUST handle JSON-RPC 2.0 notifications (requests where `id` is absent or null) by returning `nil` from `handleRaw()` — producing no response on the wire. The `notifications/initialized` notification sent by MCP clients after the initialize handshake MUST be silently acknowledged (no response). Any method matching the `notifications/*` pattern MUST follow this rule. [MUST] _(traceable: M-1)_

### Quality Domain Model Alignment

- **FR-141**: The quality dimension constants in `domain/quality.go` MUST be renamed to match the six rubric dimension names defined in `.mind/conversation/config/quality.yml`: `perspective_diversity`, `evidence_quality`, `concession_depth`, `challenge_substantiveness`, `synthesis_quality`, `actionability`. The old constant names (`rigor`, `coverage`, `objectivity`, `convergence`, `depth`) MUST be removed. `internal/service/quality.go` parsing logic MUST be updated to recognize the new names so that all six dimensions parse with non-zero values from real convergence documents. [MUST] _(traceable: M-2)_

### Test Coverage for Phase 3 Packages

- **FR-142**: The `internal/orchestrate/` package MUST have a `preflight_test.go` file with unit tests covering: `PreflightService.Run()` for each `RequestType`, the brief gate blocking condition, the doc validation step (including hard-failure blocking when `docsReport.Failed > 0`), and `WorkflowState` write. Coverage on the package MUST be ≥ 80%. [MUST] _(traceable: M-3)_
- **FR-143**: The `internal/mcp/` package MUST have a `server_test.go` file with unit tests covering: `handleRaw()` for `initialize`, `tools/list`, `tools/call` (success and unknown tool), malformed JSON, and `notifications/initialized` (verifying `nil` return — no response written). Coverage on the package MUST be ≥ 80%. [MUST] _(traceable: M-3)_
- **FR-144**: `internal/service/quality.go` MUST have a `quality_test.go` file with unit tests covering `Log()` and the dimension-parsing regex, verified against at least one real convergence document sample that includes all six rubric dimension names. [MUST] _(traceable: M-3)_

### Layer Violation Remediation

- **FR-145**: `StateRepo` (`internal/repo/interfaces.go`) SHOULD be extended with a method for appending a completed iteration entry to `docs/state/current.md`. `cmd/handoff.go` SHOULD call this method instead of using `os.ReadFile`/`os.WriteFile` directly for `current.md`. After this change, `cmd/handoff.go` MUST NOT import `os` for reading or writing `current.md`. [SHOULD] _(traceable: S-1)_

### HandoffService Extraction

- **FR-146**: `PreflightService.Handoff()` and its always-erroring `findIteration()` stub SHOULD be removed from `internal/orchestrate/preflight.go`. A `HandoffService` SHOULD be introduced in `internal/orchestrate/` with proper `IterationRepo` constructor injection, encapsulating the 5-step handoff sequence. `cmd/handoff.go` SHOULD delegate to `HandoffService` rather than implementing steps inline. [SHOULD] _(traceable: S-2, S-3)_

### Preflight Doc-Failure Blocking

- **FR-147**: `PreflightService.Run()` SHOULD block (return a non-nil error) when `docsReport.Failed > 0` after step 3 (doc validation). The error MUST identify the failure count and instruct the user to run `mind check docs`. When `docsReport.Failed == 0` and warnings exist, preflight SHOULD proceed with warnings appended to `PreflightResult.Warnings`. [SHOULD] _(traceable: S-4)_

### Branch Comparison Portability

- **FR-148**: `cmd/handoff.go` `branchAhead()` SHOULD NOT hardcode `HEAD...main`. The comparison base SHOULD be read from `mind.toml` governance settings (e.g., a `default-branch` key) and fall back to `"main"` only when no setting is present. The string literal `"HEAD...main"` MUST NOT appear in the source file. [SHOULD] _(traceable: S-5)_

### Structural Conformance (COULD)

- **FR-149**: A `classify.go` file COULD be created in `internal/orchestrate/` as a thin adapter re-exporting `domain.Classify()` and `domain.Slugify()`, satisfying the BP-08 package structure expectation. [COULD] _(traceable: S-6)_
- **FR-150**: `splitOn()` and `trimSpace()` helper functions in `internal/mcp/tools.go` COULD be replaced with direct calls to `strings.Split()` and `strings.TrimSpace()`. [COULD] _(traceable: C-1)_
- **FR-151**: `renderPreflightResult()` in `cmd/preflight.go` COULD be refactored to use the `Renderer` type consistent with other commands, adding `--json` support to `mind preflight`. [COULD] _(traceable: C-4)_

### Phase 3 Remediation Acceptance Criteria

- **FR-140**: GIVEN an MCP client sends `notifications/initialized` (no `id` field) WHEN `handleRaw()` processes it THEN `nil` is returned and no bytes are written to the transport. GIVEN `go test ./internal/mcp/...` THEN all tests pass including the notification test.
- **FR-141**: GIVEN `domain/quality.go` WHEN inspected THEN exported constants are `DimPerspectiveDiversity`, `DimEvidenceQuality`, `DimConcessionDepth`, `DimChallengeSubstantiveness`, `DimSynthesisQuality`, `DimActionability` with snake_case string values. GIVEN `QualityService.Log()` run against a convergence document using rubric names THEN all 6 dimensions parse with `Value > 0`. GIVEN `go build ./...` THEN success.
- **FR-142**: GIVEN `go test ./internal/orchestrate/...` WHEN run THEN all tests pass. GIVEN a `TypeComplexNew` request with missing brief WHEN `Run()` is called THEN an error containing "brief gate BLOCKED" is returned. GIVEN coverage tool on `internal/orchestrate/` THEN ≥ 80%.
- **FR-143**: GIVEN `go test ./internal/mcp/...` WHEN run THEN all tests pass. GIVEN `notifications/initialized` input to `handleRaw()` THEN nil is returned. GIVEN malformed JSON THEN `errParse` response is returned. GIVEN coverage tool on `internal/mcp/` THEN ≥ 80%.
- **FR-144**: GIVEN `go test ./internal/service/...` WHEN run THEN all quality tests pass. GIVEN convergence document with all 6 rubric names WHEN parsed THEN `len(Dimensions) == 6` and all `Value > 0`.
- **FR-145**: GIVEN `cmd/handoff.go` source WHEN searched for `os.ReadFile` and `os.WriteFile` THEN zero matches for `current.md` operations. GIVEN `mind handoff <valid-iter-id>` THEN `current.md` "Recent Changes" contains the iteration entry.
- **FR-146**: GIVEN `internal/orchestrate/preflight.go` WHEN inspected THEN `PreflightService` has no `Handoff()` method. GIVEN `internal/orchestrate/` WHEN inspected THEN `HandoffService` type exists with `IterationRepo` constructor dependency. GIVEN `mind handoff <valid-iter-id>` THEN all 5 steps execute with output matching pre-refactor behavior.
- **FR-147**: GIVEN a project where `mind check docs` reports `Failed > 0` WHEN `mind preflight "add feature"` runs THEN exit is non-zero with an error message naming failure count and referencing `mind check docs`. GIVEN docs with zero failures but warnings WHEN preflight runs THEN proceeds with warnings in result.
- **FR-148**: GIVEN `mind.toml` with `default-branch = "develop"` WHEN `mind handoff` runs THEN `branchAhead()` compares against `develop`. GIVEN no `default-branch` setting THEN falls back to `"main"`. GIVEN `cmd/handoff.go` source THEN `"HEAD...main"` literal does NOT appear.
- **FR-149**: GIVEN `internal/orchestrate/` WHEN inspected THEN `classify.go` exists. GIVEN `classify.go` contents THEN it delegates to `domain.Classify()` and `domain.Slugify()` without duplicating logic.
- **FR-150**: GIVEN `internal/mcp/tools.go` WHEN searched for `splitOn` or `trimSpace` function definitions THEN zero matches. GIVEN `go test ./internal/mcp/...` THEN all tests pass.
- **FR-151**: GIVEN `mind preflight "add feature" --json` WHEN run THEN output is valid JSON. GIVEN `cmd/preflight.go` WHEN inspected THEN `renderPreflightResult()` uses the `Renderer` type.
