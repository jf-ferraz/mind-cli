# API Contracts

Phase 1 CLI interface contracts for mind-cli. Since mind-cli is a CLI tool (not a REST API), these contracts define command interfaces, JSON output schemas, exit codes, and file format contracts. Distilled from [BP-03: Data Contracts](../blueprints/03-data-contracts.md) and [BP-04: CLI Specification](../blueprints/04-cli-specification.md), scoped to Phase 1 commands only.

## 1. Global Flags

Every command inherits these flags via `rootCmd.PersistentFlags()`:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | `false` | Output in JSON format (structured data to stdout, errors to stderr) |
| `--no-color` | | bool | `false` | Disable ANSI color codes in output |
| `--project-root` | `-p` | string | auto-detect | Override project root path (skip `.mind/` walk-up detection) |

**Output mode auto-detection** (FR-6):
1. `--json` flag present: JSON mode. Structured JSON to stdout, errors/progress to stderr.
2. `--no-color` flag present OR stdout is not a TTY (piped): Plain mode. Clean text, no ANSI codes.
3. Otherwise: Interactive mode. Lip Gloss styling with color and symbols.

## 2. Command Interface Contracts

### 2.1 `mind status`

| Property | Value |
|----------|-------|
| Synopsis | `mind status [flags]` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `ProjectHealth` object |
| Exit codes | 0 (healthy), 1 (issues found), 3 (not a project) |

**Behavior**: Detect project root, parse `mind.toml`, scan docs/ for zone health, run stub detection, classify brief gate, read workflow state, find latest iteration, assemble warnings and suggestions, render output.

**Degraded mode** (FR-5): When `mind.toml` is absent, status renders zone scan and stub detection with a warning. Config-dependent fields are omitted.

---

### 2.2 `mind init`

| Property | Value |
|----------|-------|
| Synopsis | `mind init [flags]` |
| Arguments | None |
| Flags | `--name` (string), `--with-github` (bool), `--from-existing` (bool) |
| Requires project | No (creates one) |
| JSON output | `InitResult` object |
| Exit codes | 0 (created), 1 (I/O failure), 2 (already initialized) |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--name` | `-n` | string | directory name | Project name for mind.toml `[project].name` |
| `--with-github` | | bool | `false` | Create `.github/agents/` adapter |
| `--from-existing` | | bool | `false` | Detect existing docs/, preserve and register |

**Behavior**: Guard that `.mind/` does not exist (exit 2 if it does). Determine project name. Create `.mind/`, `docs/` with 5 zone directories, stub documents, `mind.toml`, `.claude/CLAUDE.md`. If `--with-github`: create `.github/agents/`. If `--from-existing`: scan for existing docs, skip stubs where files exist, register discovered files.

**Created file list** (FR-14):
- `.mind/` directory
- `docs/spec/`, `docs/blueprints/`, `docs/state/`, `docs/iterations/`, `docs/knowledge/`
- `docs/spec/project-brief.md`, `docs/spec/requirements.md`, `docs/spec/architecture.md`, `docs/spec/domain-model.md`
- `docs/state/current.md`, `docs/state/workflow.md`
- `docs/blueprints/INDEX.md`
- `docs/knowledge/glossary.md`
- `mind.toml`
- `.claude/CLAUDE.md`

---

### 2.3 `mind doctor`

| Property | Value |
|----------|-------|
| Synopsis | `mind doctor [flags]` |
| Arguments | None |
| Flags | `--fix` (bool) |
| Requires project | Yes |
| JSON output | `DoctorReport` object |
| Exit codes | 0 (all pass), 1 (failures or partial fix), 3 (not a project) |

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--fix` | | bool | `false` | Auto-fix resolvable issues |

**Diagnostic categories**: framework, docs, refs, config, brief, workflow, iterations, naming.

**Auto-fixable issues** (`--fix`): Create missing directories, add `.gitkeep` files, create stub documents from templates, fix naming convention violations.

---

### 2.4 `mind create adr`

| Property | Value |
|----------|-------|
| Synopsis | `mind create adr "<title>"` |
| Arguments | `title` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `CreateResult` object |
| Exit codes | 0 (created), 1 (failed or target exists), 3 (not a project) |

**Behavior**: Auto-number using `{NNN}` (3-digit zero-padded). Next sequence = max(existing) + 1. Slugify title. Create `docs/spec/decisions/{NNN}-{slug}.md` with ADR template. Abort if target file exists (FR-30).

---

### 2.5 `mind create blueprint`

| Property | Value |
|----------|-------|
| Synopsis | `mind create blueprint "<title>"` |
| Arguments | `title` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `CreateResult` object (includes `index_updated: true`) |
| Exit codes | 0 (created), 1 (failed or target exists), 3 (not a project) |

**Behavior**: Auto-number using `{NN}` (2-digit zero-padded). Slugify title. Create `docs/blueprints/{NN}-{slug}.md`. Append entry to `docs/blueprints/INDEX.md` (create INDEX.md if missing). Abort if target exists.

---

### 2.6 `mind create iteration`

| Property | Value |
|----------|-------|
| Synopsis | `mind create iteration <type> "<name>"` |
| Arguments | `type` (required: new/enhancement/bugfix/refactor), `name` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `CreateIterationResult` object |
| Exit codes | 0 (created), 1 (failed, invalid type, or target exists), 3 (not a project) |

**Type mapping**:
| Input | Canonical |
|-------|-----------|
| `new` | `NEW_PROJECT` |
| `enhancement` | `ENHANCEMENT` |
| `bugfix` | `BUG_FIX` |
| `refactor` | `REFACTOR` |

**Behavior**: Auto-number using `{NNN}`. Slugify name. Create `docs/iterations/{NNN}-{TYPE}-{slug}/` with exactly 5 files: `overview.md`, `changes.md`, `test-summary.md`, `validation.md`, `retrospective.md`.

---

### 2.7 `mind create spike`

| Property | Value |
|----------|-------|
| Synopsis | `mind create spike "<title>"` |
| Arguments | `title` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `CreateResult` object |
| Exit codes | 0 (created), 1 (failed or target exists), 3 (not a project) |

**Behavior**: Slugify title. Create `docs/knowledge/{slug}-spike.md`. Ensure `docs/knowledge/` exists. Abort if target exists.

---

### 2.8 `mind create convergence`

| Property | Value |
|----------|-------|
| Synopsis | `mind create convergence "<title>"` |
| Arguments | `title` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `CreateResult` object |
| Exit codes | 0 (created), 1 (failed or target exists), 3 (not a project) |

**Behavior**: Slugify title. Create `docs/knowledge/{slug}-convergence.md`. Ensure `docs/knowledge/` exists. Abort if target exists.

---

### 2.9 `mind create brief`

| Property | Value |
|----------|-------|
| Synopsis | `mind create brief` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | Not supported (interactive-only command; `--json` prints error) |
| Exit codes | 0 (created, gate passes), 1 (created but gate fails, or user cancelled), 3 (not a project) |

**Behavior**: Interactive prompts for Vision, Key Deliverables, In Scope, Out of Scope, Constraints. Write `docs/spec/project-brief.md`. Run brief gate validation on result. Requires TTY -- if stdin is piped, exit with error suggesting direct file editing.

---

### 2.10 `mind docs list`

| Property | Value |
|----------|-------|
| Synopsis | `mind docs list [--zone ZONE]` |
| Arguments | None |
| Flags | `--zone` (string: spec/blueprints/state/iterations/knowledge) |
| Requires project | Yes |
| JSON output | `DocumentList` object |
| Exit codes | 0 (success), 1 (invalid zone), 3 (not a project) |

---

### 2.11 `mind docs tree`

| Property | Value |
|----------|-------|
| Synopsis | `mind docs tree` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | Nested tree object |
| Exit codes | 0 (success), 3 (not a project) |

---

### 2.12 `mind docs stubs`

| Property | Value |
|----------|-------|
| Synopsis | `mind docs stubs` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | Array of stub objects |
| Exit codes | 0 (no stubs), 1 (stubs found), 3 (not a project) |

---

### 2.13 `mind docs search`

| Property | Value |
|----------|-------|
| Synopsis | `mind docs search "<query>"` |
| Arguments | `query` (required, string) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | Search results array |
| Exit codes | 0 (success), 3 (not a project) |

**Behavior**: Case-insensitive substring search across all `.md` files in `docs/`. Returns matching lines with 1 line of context, grouped by file.

---

### 2.14 `mind docs open`

| Property | Value |
|----------|-------|
| Synopsis | `mind docs open <path-or-id>` |
| Arguments | `path-or-id` (required: file path, `doc:zone/name` ID, or fuzzy name) |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `{ "path": "...", "abs_path": "..." }` (does not open editor in JSON mode) |
| Exit codes | 0 (editor opened), 1 (not found, ambiguous, no $EDITOR), 3 (not a project) |

---

### 2.15 `mind check docs`

| Property | Value |
|----------|-------|
| Synopsis | `mind check docs [--strict]` |
| Arguments | None |
| Flags | `--strict` (bool: promote WARN to FAIL) |
| Requires project | Yes |
| JSON output | `ValidationReport` object |
| Exit codes | 0 (all pass, warnings acceptable), 1 (FAIL-level check failed), 3 (not a project) |

**17 checks**: (1) docs/ exists, (2) 5 zone dirs exist, (3) required spec files, (4) decisions/ dir, (5) ADR naming, (6) INDEX.md exists, (7) blueprint coverage, (8) INDEX.md refs, (9) current.md, (10) workflow.md, (11) glossary.md, (12) iteration naming, (13) iteration overview.md, (14) spike naming, (15) no legacy paths, (16) stub detection, (17) brief completeness.

---

### 2.16 `mind check refs`

| Property | Value |
|----------|-------|
| Synopsis | `mind check refs` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `ValidationReport` object |
| Exit codes | 0 (all pass), 1 (FAIL-level check failed), 3 (not a project) |

**11 checks**: (1) CLAUDE.md references, (2) agent file references, (3) blueprint cross-references, (4) INDEX.md links, (5) iteration overview references, (6) mind.toml paths exist, (7) mind.toml graph references, (8) no broken links in spec/, (9) no broken links in blueprints/, (10) ADR numbering sequential, (11) iteration numbering sequential.

---

### 2.17 `mind check config`

| Property | Value |
|----------|-------|
| Synopsis | `mind check config` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `ValidationReport` object |
| Exit codes | 0 (valid), 1 (invalid), 3 (not a project) |

**Checks**: Valid TOML, required fields present, document IDs consistent, graph acyclic.

---

### 2.18 `mind check all`

| Property | Value |
|----------|-------|
| Synopsis | `mind check all [--strict]` |
| Arguments | None |
| Flags | `--strict` (bool) |
| Requires project | Yes |
| JSON output | `UnifiedValidationReport` object (contains `suites` array) |
| Exit codes | 0 (all pass), 1 (any FAIL), 3 (not a project) |

---

### 2.19 `mind workflow status`

| Property | Value |
|----------|-------|
| Synopsis | `mind workflow status` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `WorkflowStatus` object |
| Exit codes | 0 (success), 3 (not a project) |

---

### 2.20 `mind workflow history`

| Property | Value |
|----------|-------|
| Synopsis | `mind workflow history` |
| Arguments | None |
| Flags | Global only |
| Requires project | Yes |
| JSON output | `WorkflowHistory` object |
| Exit codes | 0 (success), 3 (not a project) |

---

### 2.21 `mind version`

| Property | Value |
|----------|-------|
| Synopsis | `mind version [--short]` |
| Arguments | None |
| Flags | `--short` (bool: version string only) |
| Requires project | No |
| JSON output | `VersionInfo` object |
| Exit codes | 0 (always) |

---

### 2.22 `mind help`

| Property | Value |
|----------|-------|
| Synopsis | `mind help [command]` |
| Arguments | `command` (optional) |
| Requires project | No |
| JSON output | N/A (Cobra auto-generated) |
| Exit codes | 0 (always) |

## 3. JSON Output Schemas

All JSON output is a single JSON object on stdout (never an array, never NDJSON). Timestamps are RFC 3339. Paths are relative to project root unless noted.

### 3.1 ProjectHealth (`mind status --json`)

```json
{
  "project": {
    "name": "string",
    "root": "string (absolute path)",
    "description": "string",
    "type": "string",
    "framework_version": "string",
    "stack": {
      "language": "string",
      "framework": "string",
      "testing": "string"
    }
  },
  "brief": {
    "exists": "bool",
    "is_stub": "bool",
    "gate": "BRIEF_PRESENT | BRIEF_STUB | BRIEF_MISSING",
    "sections": {
      "vision": "bool",
      "key_deliverables": "bool",
      "scope": "bool"
    }
  },
  "zones": {
    "<zone_name>": {
      "total": "int",
      "present": "int",
      "complete": "int",
      "stubs": "int",
      "files": [
        {
          "path": "string (relative)",
          "name": "string",
          "status": "draft | active | complete | stub",
          "is_stub": "bool",
          "size": "int (bytes)",
          "mod_time": "string (RFC 3339)"
        }
      ]
    }
  },
  "workflow": "WorkflowState object | null",
  "last_iteration": {
    "seq": "int",
    "type": "NEW_PROJECT | BUG_FIX | ENHANCEMENT | REFACTOR",
    "descriptor": "string",
    "dir_name": "string",
    "status": "in_progress | complete | incomplete",
    "created_at": "string (RFC 3339)"
  },
  "warnings": ["string"],
  "suggestions": ["string"]
}
```

**Notes**:
- `project.root` is the only absolute path in any JSON output.
- `workflow` is `null` when idle.
- `last_iteration` is `null` when no iterations exist.
- `zones` keys are the 5 zone names: `spec`, `blueprints`, `state`, `iterations`, `knowledge`.

### 3.2 ValidationReport (`mind check docs|refs|config --json`)

```json
{
  "suite": "docs | refs | config",
  "total": "int",
  "passed": "int",
  "failed": "int",
  "warnings": "int",
  "checks": [
    {
      "id": "int",
      "name": "string",
      "level": "FAIL | WARN | INFO",
      "passed": "bool",
      "message": "string (empty when passed)"
    }
  ]
}
```

### 3.3 UnifiedValidationReport (`mind check all --json`)

```json
{
  "suites": [
    {
      "name": "docs | refs | config",
      "total": "int",
      "passed": "int",
      "failed": "int",
      "warnings": "int",
      "checks": ["...same as ValidationReport.checks"]
    }
  ],
  "summary": {
    "total": "int",
    "passed": "int",
    "failed": "int",
    "warnings": "int"
  }
}
```

**Notes**:
- `suites` is always an array with entries in order: `docs`, `refs`, `config`.
- When `--strict` is passed, WARN-level checks are promoted to FAIL and counts shift accordingly.

### 3.4 DoctorReport (`mind doctor --json`)

```json
{
  "diagnostics": [
    {
      "category": "framework | docs | refs | config | brief | workflow | iterations | naming",
      "check": "string",
      "status": "pass | fail | warn",
      "message": "string",
      "fix": "string (empty if no fix)",
      "auto_fixable": "bool"
    }
  ],
  "summary": {
    "pass": "int",
    "fail": "int",
    "warn": "int"
  },
  "fixes_applied": ["string (only present with --fix)"]
}
```

### 3.5 InitResult (`mind init --json`)

```json
{
  "project_name": "string",
  "root": "string (absolute path)",
  "files_created": ["string (relative path)"],
  "from_existing": "bool",
  "existing_preserved": ["string (relative path, only with --from-existing)"]
}
```

### 3.6 CreateResult (`mind create adr|blueprint|spike|convergence --json`)

```json
{
  "path": "string (relative)",
  "seq": "int (only for adr and blueprint)",
  "title": "string",
  "index_updated": "bool (only for blueprint)"
}
```

### 3.7 CreateIterationResult (`mind create iteration --json`)

```json
{
  "path": "string (relative, directory)",
  "seq": "int",
  "type": "NEW_PROJECT | BUG_FIX | ENHANCEMENT | REFACTOR",
  "descriptor": "string",
  "files": ["overview.md", "changes.md", "test-summary.md", "validation.md", "retrospective.md"]
}
```

### 3.8 DocumentList (`mind docs list --json`)

```json
{
  "documents": [
    {
      "path": "string (relative)",
      "zone": "string",
      "name": "string",
      "id": "doc:<zone>/<name>",
      "status": "draft | active | complete | stub",
      "is_stub": "bool",
      "size": "int (bytes)",
      "mod_time": "string (RFC 3339)"
    }
  ],
  "by_zone": {
    "<zone_name>": "int (count)"
  },
  "total": "int"
}
```

**Notes**:
- When `--zone` is provided, `documents` is filtered but `by_zone` and `total` reflect all zones.
- Documents are sorted by zone order, then alphabetically within each zone.

### 3.9 StubList (`mind docs stubs --json`)

```json
{
  "stubs": [
    {
      "path": "string (relative)",
      "zone": "string",
      "hint": "string (remediation suggestion)"
    }
  ],
  "count": "int"
}
```

### 3.10 SearchResults (`mind docs search --json`)

```json
{
  "query": "string",
  "results": [
    {
      "path": "string (relative)",
      "matches": [
        {
          "line": "int",
          "text": "string",
          "context_before": "string",
          "context_after": "string"
        }
      ]
    }
  ],
  "total_matches": "int",
  "files_matched": "int"
}
```

### 3.11 WorkflowStatus (`mind workflow status --json`)

```json
{
  "state": "idle | running",
  "type": "NEW_PROJECT | BUG_FIX | ENHANCEMENT | REFACTOR | (empty when idle)",
  "descriptor": "string",
  "iteration_path": "string (relative)",
  "branch": "string",
  "last_agent": "string",
  "remaining_chain": ["string"],
  "session": "int",
  "total_sessions": "int"
}
```

**Note**: When idle, `state` is `"idle"` and all other fields are empty/zero.

### 3.12 WorkflowHistory (`mind workflow history --json`)

```json
{
  "iterations": [
    {
      "seq": "int",
      "type": "string",
      "descriptor": "string",
      "dir_name": "string",
      "status": "in_progress | complete | incomplete",
      "created_at": "string (RFC 3339)",
      "artifacts": {
        "present": "int",
        "expected": 5
      }
    }
  ],
  "total": "int"
}
```

### 3.13 VersionInfo (`mind version --json`)

```json
{
  "version": "string",
  "commit": "string (SHA)",
  "build_date": "string",
  "go_version": "string",
  "os": "string",
  "arch": "string"
}
```

## 4. Exit Codes

| Code | Meaning | Used By | Example |
|------|---------|---------|---------|
| 0 | Success | All commands | Checks pass, file created, status healthy |
| 1 | Validation failure / issues found | `check *`, `doctor`, `docs stubs` | A FAIL-level check did not pass; stubs exist; doctor found failures |
| 2 | Runtime error | `init` | `.mind/` already exists; I/O failure; permission denied |
| 3 | Configuration error / not a project | All project-requiring commands | No `.mind/` directory found in ancestry; invalid mind.toml |

**Rules**:
- Exit code is deterministic: same inputs produce same exit code (BR-20).
- WARN-level failures do NOT affect exit code unless `--strict` is set (BR-21).
- `mind version` and `mind help` always exit 0.
- When `--fix` is used with `mind doctor` and some fixes fail, exit code is 1.

## 5. File Format Contracts

### 5.1 `mind.toml` (read by Phase 1, written by `mind init`)

Full schema is in [BP-03 Section 1](../blueprints/03-data-contracts.md). Phase 1 validation rules:

| Field | Rule | Error Level |
|-------|------|-------------|
| `manifest.schema` | Must match `^mind/v\d+\.\d+$` | FAIL |
| `manifest.generation` | Must be >= 1 | FAIL |
| `project.name` | Must match `^[a-z][a-z0-9-]*$` (kebab-case) | FAIL |
| `project.type` | Must be one of: cli, api, library, webapp, service | FAIL |
| `documents.*.id` | Must match `^doc:[a-z]+/[a-z][a-z0-9-]*$` | FAIL |
| `documents.*.path` | Must start with `docs/` and end with `.md` | FAIL |
| `documents.*.zone` | Must be one of: spec, blueprints, state, iterations, knowledge | FAIL |
| `documents.*.status` | Must be one of: draft, active, complete | FAIL |
| `governance.max-retries` | Must be 0-5 | WARN |

### 5.2 Iteration Directory Structure

Pattern: `docs/iterations/{NNN}-{TYPE}-{slug}/`

```
{NNN}-{TYPE}-{slug}/
  overview.md           # Created on iteration creation (always present)
  changes.md            # Filled during development
  test-summary.md       # Filled during testing
  validation.md         # Filled during review
  retrospective.md      # Filled after completion
```

- `{NNN}`: 3-digit zero-padded sequence number (001, 002, ...)
- `{TYPE}`: One of NEW_PROJECT, ENHANCEMENT, BUG_FIX, REFACTOR
- `{slug}`: Kebab-case slugified from the descriptor

### 5.3 Slugification Algorithm (BR-16)

```
Input:  "Use PostgreSQL (v15+)"
Step 1: Lowercase             -> "use postgresql (v15+)"
Step 2: Replace non-alnum     -> "use-postgresql--v15-"
Step 3: Collapse multi-hyphen -> "use-postgresql-v15-"
Step 4: Strip leading/trailing -> "use-postgresql-v15"
```

The algorithm is implemented in `domain.Slugify()` and must be deterministic and idempotent.

### 5.4 Stub Detection Algorithm (FR-50, BR-2)

A document is classified as a stub if it contains no more than 2 lines of substantive content. Lines classified as boilerplate (not substantive):
- Empty lines
- Markdown headings (`# ...`, `## ...`)
- HTML comments (`<!-- ... -->`, `-->`)
- Blockquote lines (`> ...`)
- Table separators (`|---|---|`)
- Table rows containing only HTML comment placeholders

Implementation: `internal/repo/fs.IsStubContent()`.

### 5.5 ADR File Naming

Pattern: `docs/spec/decisions/{NNN}-{slug}.md`
- `{NNN}`: 3+ digit zero-padded (001, 002, ..., 999)
- Sequence: max(existing ADR numbers) + 1

### 5.6 Blueprint File Naming

Pattern: `docs/blueprints/{NN}-{slug}.md`
- `{NN}`: 2+ digit zero-padded (01, 02, ..., 99)
- Sequence: max(existing blueprint numbers) + 1
- Side effect: entry appended to `docs/blueprints/INDEX.md`
