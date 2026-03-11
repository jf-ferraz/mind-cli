# Requirements

## Overview

Phase 1 of mind-cli delivers a deterministic CLI that replaces the Mind Agent Framework's 9 bash scripts with a single Go binary. It provides project health diagnostics, 17-check documentation validation, 11-check cross-reference validation, document scaffolding for 6 artifact types, workflow state inspection, and three output modes (interactive, plain, JSON) -- all accessed through a unified `mind` command with `--json` support on every subcommand.

This document covers Phase 1 (Core CLI) only. Reconciliation (Phase 1.5), TUI (Phase 2), MCP server (Phase 3), watch mode and orchestration (Phase 4) are explicitly out of scope.

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
- Reconciliation engine, mind.lock, staleness propagation (Phase 1.5)
- TUI dashboard (Phase 2)
- MCP server, pre-flight, handoff (Phase 3)
- Watch mode, full orchestration (Phase 4)
- Shell completions (Phase 5)

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
