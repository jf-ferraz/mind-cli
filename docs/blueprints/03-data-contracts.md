# Blueprint: Data Contracts & Schemas

> What does every external format look like, field by field? Complete schemas for all file formats, JSON outputs, MCP tool interfaces, and structured data the mind-cli reads or writes.

**Status**: Active
**Date**: 2026-03-11
**Depends on**: [01-mind-cli.md](01-mind-cli.md), [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md), [03-architecture.md](03-architecture.md)

---

## 1. `mind.toml` Schema

The project manifest. Every Mind Framework project has exactly one `mind.toml` at the project root. The CLI reads it on every invocation; writes happen only during `mind init` and when the reconciliation engine bumps `generation`.

### Complete Schema

```toml
# ─── Manifest Metadata ───────────────────────────────────────────────────────
[manifest]
schema     = "mind/v2.0"                 # required, string
                                          # Format: "mind/vX.Y" where X is major, Y is minor
                                          # Validation: must match ^mind/v\d+\.\d+$
                                          # The CLI reads any version >= mind/v1.0
                                          # Unknown versions produce a warning, not an error

generation = 1                            # required, int, >= 1
                                          # Auto-incremented on every write to mind.toml
                                          # Used by mind.lock to detect manifest changes

updated    = 2026-03-10T00:13:40Z         # required, TOML datetime (RFC 3339)
                                          # Set automatically on every write
                                          # Used by mind.lock staleness detection

# ─── Manifest Invariants ─────────────────────────────────────────────────────
[manifest.invariants]
every-document-has-owner   = false        # optional, bool, default: false
                                          # When true, every [documents.*] entry must have
                                          # an "owner" field (reserved for future use)

no-orphan-dependencies     = true         # optional, bool, default: true
                                          # When true, every [[graph]] "from"/"to" must
                                          # reference a document ID that exists in [documents]

no-circular-dependencies   = true         # optional, bool, default: true
                                          # When true, the [[graph]] edges must form a DAG
                                          # Validated by `mind check config`

# ─── Project Identity ────────────────────────────────────────────────────────
[project]
name        = "my-project"                # required, string
                                          # Format: kebab-case, ^[a-z][a-z0-9-]*$
                                          # Used in TUI header, status output, branch naming

description = "A short description"       # optional, string, default: ""
                                          # Displayed in `mind status` and `mind_status` MCP tool

type        = "cli"                       # required, string
                                          # One of: "cli", "api", "library", "webapp", "service"
                                          # Affects agent chain defaults and preflight heuristics

# ─── Stack Configuration ─────────────────────────────────────────────────────
[project.stack]
language  = "go@1.23"                     # required, string
                                          # Format: "{language}@{version}"
                                          # Validation: must match ^[a-z]+@.+$
                                          # The version portion is free-form (e.g., "3.12", "1.23", "21")
                                          # Used by `mind check gate` to select default commands

framework = "cobra+bubbletea"             # optional, string, default: ""
                                          # Free-form, used for display only

testing   = "go-test"                     # optional, string, default: ""
                                          # Free-form, used for display only

# ─── Build/Dev Commands ──────────────────────────────────────────────────────
[project.commands]
dev       = "go run ."                    # optional, string, default: ""
                                          # Shell command for development mode
                                          # Not used by any gate — informational only

test      = "go test ./..."               # optional, string, default: ""
                                          # Shell command for running tests
                                          # Used by deterministic gate (`mind check gate`)

lint      = "golangci-lint run ./..."     # optional, string, default: ""
                                          # Shell command for linting
                                          # Used by deterministic gate (`mind check gate`)

typecheck = ""                            # optional, string, default: ""
                                          # Shell command for type checking
                                          # Used by deterministic gate if non-empty

build     = "go build -o mind ."          # optional, string, default: ""
                                          # Shell command for building
                                          # Used by deterministic gate (`mind check gate`)

# ─── Profiles ────────────────────────────────────────────────────────────────
[profiles]
active = []                               # optional, string array, default: []
                                          # Active profile names — reserved for future use
                                          # Profiles modify CLI behavior (e.g., "strict", "ci")
                                          # Each entry: ^[a-z][a-z0-9-]*$

# ─── Document Registry ───────────────────────────────────────────────────────
# Format: [documents.{zone}.{name}]
# The zone and name segments form the document ID: "doc:{zone}/{name}"
# Every document the framework tracks must have an entry here.

[documents.spec.project-brief]
id     = "doc:spec/project-brief"         # required, string
                                          # Format: "doc:{zone}/{name}"
                                          # Must match the table key: documents.{zone}.{name}

path   = "docs/spec/project-brief.md"    # required, string
                                          # Relative path from project root
                                          # Must start with "docs/"
                                          # Must end with ".md"

zone   = "spec"                           # required, string
                                          # One of: "spec", "blueprints", "state",
                                          #         "iterations", "knowledge"
                                          # Must match the {zone} in the table key

status = "draft"                          # required, string
                                          # One of: "draft", "active", "complete"
                                          # "draft" — initial state, content not yet reviewed
                                          # "active" — content is current and authoritative
                                          # "complete" — content is finalized, no further changes expected

# Additional document entries follow the same pattern:
# [documents.spec.requirements]
# [documents.spec.architecture]
# [documents.spec.domain-model]
# [documents.blueprints.index]
# [documents.state.current]
# [documents.state.workflow]
# [documents.knowledge.glossary]
# ... (as many as the project needs)

# ─── Governance ──────────────────────────────────────────────────────────────
[governance]
max-retries     = 2                       # optional, int, default: 2
                                          # Maximum retry loops per agent in a workflow
                                          # Range: 0-5
                                          # 0 means no retries — gate failures proceed with concerns

review-policy   = "evidence-based"        # optional, string, default: "evidence-based"
                                          # One of: "evidence-based", "checklist", "none"
                                          # Controls the reviewer agent's evaluation approach

commit-policy   = "conventional"          # optional, string, default: "conventional"
                                          # One of: "conventional", "freeform"
                                          # "conventional" enforces Conventional Commits format

branch-strategy = "type-descriptor"       # optional, string, default: "type-descriptor"
                                          # One of: "type-descriptor", "flat", "none"
                                          # "type-descriptor" → {type}/{descriptor}
                                          #   e.g., new/rest-api, fix/auth-redirect
                                          # "flat" → {descriptor} only
                                          # "none" → no automatic branch creation

# ─── Dependency Graph ────────────────────────────────────────────────────────
# Array of tables — each entry is a directed edge between two documents.
# Used by the reconciliation engine to determine staleness propagation.

[[graph]]
from = "doc:spec/requirements"            # required, string
                                          # Document ID of the source (the informing document)
                                          # Must reference a document ID in [documents]

to   = "doc:spec/architecture"            # required, string
                                          # Document ID of the target (the informed document)
                                          # Must reference a document ID in [documents]

type = "informs"                          # required, string
                                          # One of: "informs", "requires", "validates"
                                          # "informs"   — source content shapes target content
                                          # "requires"  — target cannot be written without source
                                          # "validates" — source is used to verify target correctness

# Additional edges:
# [[graph]]
# from = "doc:spec/architecture"
# to   = "doc:spec/domain-model"
# type = "informs"
```

### Validation Rules Summary

| Field | Rule | Error Level |
|-------|------|-------------|
| `manifest.schema` | Must match `^mind/v\d+\.\d+$` | FAIL |
| `manifest.generation` | Must be >= 1 | FAIL |
| `manifest.updated` | Must be valid RFC 3339 | FAIL |
| `project.name` | Must match `^[a-z][a-z0-9-]*$` | FAIL |
| `project.type` | Must be one of allowed values | FAIL |
| `project.stack.language` | Must match `^[a-z]+@.+$` | FAIL |
| `documents.*.id` | Must match `^doc:[a-z]+/[a-z][a-z0-9-]*$` | FAIL |
| `documents.*.path` | Must start with `docs/` and end with `.md` | FAIL |
| `documents.*.zone` | Must be a valid zone | FAIL |
| `documents.*.status` | Must be one of allowed values | FAIL |
| `governance.max-retries` | Must be 0-5 | WARN |
| `graph[].from`, `graph[].to` | Must reference existing document IDs | FAIL (if invariant enabled) |
| `graph[]` | Must be acyclic | FAIL (if invariant enabled) |

---

## 2. `mind.lock` Schema

JSON lock file written by the reconciliation engine (`mind reconcile`). Located at `mind.lock` in the project root. Used to detect staleness, track file hashes, and verify document graph integrity without re-reading every file.

```json
{
  "schema_version": "1.0",
  "generated_at": "2026-03-11T14:30:00Z",
  "project_hash": "sha256:a1b2c3d4e5f6...",
  "manifest_generation": 3,
  "entries": {
    "doc:spec/project-brief": {
      "id": "doc:spec/project-brief",
      "path": "docs/spec/project-brief.md",
      "zone": "spec",
      "status": "active",
      "hash": "sha256:def456abc789...",
      "size": 3200,
      "mod_time": "2026-03-10T15:00:00Z",
      "is_stub": false,
      "stale": false,
      "stale_reason": "",
      "depends_on": ["doc:spec/requirements"]
    },
    "doc:spec/requirements": {
      "id": "doc:spec/requirements",
      "path": "docs/spec/requirements.md",
      "zone": "spec",
      "status": "active",
      "hash": "sha256:789abc123def...",
      "size": 8100,
      "mod_time": "2026-03-09T10:00:00Z",
      "is_stub": false,
      "stale": false,
      "stale_reason": "",
      "depends_on": []
    },
    "doc:spec/architecture": {
      "id": "doc:spec/architecture",
      "path": "docs/spec/architecture.md",
      "zone": "spec",
      "status": "active",
      "hash": "sha256:abc123def456...",
      "size": 6400,
      "mod_time": "2026-03-09T12:00:00Z",
      "is_stub": false,
      "stale": true,
      "stale_reason": "dependency changed: doc:spec/requirements",
      "depends_on": ["doc:spec/requirements"]
    }
  },
  "dependency_graph": [
    {
      "from": "doc:spec/requirements",
      "to": "doc:spec/architecture",
      "type": "informs"
    },
    {
      "from": "doc:spec/architecture",
      "to": "doc:spec/domain-model",
      "type": "informs"
    }
  ],
  "stats": {
    "total_documents": 10,
    "stale_documents": 2,
    "missing_documents": 0,
    "undeclared_files": 1
  }
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | string | Lock file schema version. Currently `"1.0"`. |
| `generated_at` | string (RFC 3339) | Timestamp when the lock file was last written. |
| `project_hash` | string | SHA-256 of `mind.toml` content. Format: `"sha256:{hex}"`. Used to detect manifest changes since last reconcile. |
| `manifest_generation` | int | The `manifest.generation` value from `mind.toml` at reconcile time. |
| `entries` | object | Map of document ID to entry object. Keys match `documents.*.id` from `mind.toml`. |
| `entries.*.id` | string | Document ID. Same as the map key. |
| `entries.*.path` | string | Relative file path from project root. |
| `entries.*.zone` | string | Documentation zone. |
| `entries.*.status` | string | Document status from `mind.toml`. |
| `entries.*.hash` | string | SHA-256 of file content. Format: `"sha256:{hex}"`. Empty string if file does not exist. |
| `entries.*.size` | int | File size in bytes. `0` if file does not exist. |
| `entries.*.mod_time` | string (RFC 3339) | File modification time. Empty string if file does not exist. |
| `entries.*.is_stub` | bool | Whether the file was detected as a stub at reconcile time. |
| `entries.*.stale` | bool | Whether the document is stale (a dependency changed since last reconcile). |
| `entries.*.stale_reason` | string | Human-readable reason for staleness. Empty if not stale. |
| `entries.*.depends_on` | string array | Document IDs this entry depends on (derived from `[[graph]]` edges where this document is the `to`). |
| `dependency_graph` | array | Copy of `[[graph]]` edges from `mind.toml` at reconcile time. |
| `dependency_graph[].from` | string | Source document ID. |
| `dependency_graph[].to` | string | Target document ID. |
| `dependency_graph[].type` | string | Edge type: `"informs"`, `"requires"`, or `"validates"`. |
| `stats.total_documents` | int | Total documents registered in `mind.toml`. |
| `stats.stale_documents` | int | Documents with `stale: true`. |
| `stats.missing_documents` | int | Documents registered in `mind.toml` but not present on disk. |
| `stats.undeclared_files` | int | Files in `docs/` that match document patterns but are not registered in `mind.toml`. |

### Staleness Algorithm

A document is marked stale when:

1. Its own `hash` differs from the hash recorded in the previous lock file (content changed).
2. Any document in its `depends_on` list has a newer `mod_time` than this document's `mod_time`.
3. The `manifest_generation` increased and the document's entry in `mind.toml` changed (status, path, or zone modified).

Staleness propagates transitively through `"informs"` and `"requires"` edges but not through `"validates"` edges.

---

## 3. JSON Output Schemas

Every CLI command that supports `--json` produces a JSON object on stdout. The top-level structure is always a single JSON object (never an array, never streaming NDJSON). All timestamps are RFC 3339. All paths are relative to the project root unless noted otherwise.

### `mind status --json`

```json
{
  "project": {
    "name": "mind-cli",
    "root": "/home/user/projects/mind-cli",
    "description": "CLI and TUI for the Mind Agent Framework",
    "type": "cli",
    "framework_version": "v2026-03-09",
    "stack": {
      "language": "go@1.23",
      "framework": "cobra+bubbletea",
      "testing": "go-test"
    }
  },
  "brief": {
    "exists": true,
    "is_stub": false,
    "gate": "BRIEF_PRESENT",
    "sections": {
      "vision": true,
      "key_deliverables": true,
      "scope": true,
      "constraints": true,
      "success_metrics": true
    }
  },
  "zones": {
    "spec": {
      "total": 5,
      "present": 5,
      "complete": 4,
      "stubs": 1,
      "files": [
        {
          "path": "docs/spec/project-brief.md",
          "name": "project-brief",
          "status": "active",
          "is_stub": false,
          "size": 3200,
          "mod_time": "2026-03-10T15:00:00Z"
        }
      ]
    },
    "blueprints": {
      "total": 4,
      "present": 4,
      "complete": 4,
      "stubs": 0,
      "files": []
    },
    "state": {
      "total": 2,
      "present": 2,
      "complete": 1,
      "stubs": 1,
      "files": []
    },
    "iterations": {
      "total": 6,
      "present": 6,
      "complete": 6,
      "stubs": 0,
      "files": []
    },
    "knowledge": {
      "total": 3,
      "present": 3,
      "complete": 2,
      "stubs": 1,
      "files": []
    }
  },
  "workflow": {
    "state": "running",
    "type": "NEW_PROJECT",
    "descriptor": "rest-api",
    "last_agent": "architect",
    "remaining_chain": ["developer", "tester", "reviewer"],
    "session": 1,
    "total_sessions": 2
  },
  "last_iteration": {
    "seq": 7,
    "type": "NEW_PROJECT",
    "descriptor": "rest-api",
    "dir_name": "007-NEW_PROJECT-rest-api",
    "status": "in_progress",
    "created_at": "2026-03-11T10:00:00Z"
  },
  "warnings": [
    "domain-model.md is a stub",
    "glossary.md is a stub"
  ],
  "suggestions": [
    "Fill docs/spec/domain-model.md with entity definitions",
    "Fill docs/knowledge/glossary.md with domain terms"
  ]
}
```

**Field notes:**

- `project.root` is the only absolute path in any JSON output.
- `workflow` is `null` when no workflow is active (idle state).
- `last_iteration` is `null` when no iterations exist.
- `zones.*.files` contains every file in that zone. The array may be large for the `iterations` zone.
- `brief.gate` is one of: `"BRIEF_PRESENT"`, `"BRIEF_STUB"`, `"BRIEF_MISSING"`.

### `mind check all --json`

```json
{
  "suites": [
    {
      "name": "docs",
      "total": 17,
      "passed": 15,
      "failed": 1,
      "warnings": 1,
      "checks": [
        {
          "id": 1,
          "name": "docs/ directory exists",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 2,
          "name": "All 5 zone directories exist",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 3,
          "name": "Required spec files",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 4,
          "name": "decisions/ subdirectory",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 5,
          "name": "ADR naming convention",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 6,
          "name": "blueprints/INDEX.md",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 7,
          "name": "Blueprint to INDEX.md coverage",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 8,
          "name": "INDEX.md to file references",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 9,
          "name": "state/current.md",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 10,
          "name": "state/workflow.md",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 11,
          "name": "knowledge/glossary.md",
          "level": "WARN",
          "passed": false,
          "message": "glossary.md not found"
        },
        {
          "id": 12,
          "name": "Iteration folder naming",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 13,
          "name": "Iterations have overview.md",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 14,
          "name": "Spike file naming",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 15,
          "name": "No legacy paths",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 16,
          "name": "Stub detection",
          "level": "WARN",
          "passed": false,
          "message": "2 stubs found: domain-model.md, glossary.md"
        },
        {
          "id": 17,
          "name": "Project brief completeness",
          "level": "WARN",
          "passed": true,
          "message": ""
        }
      ]
    },
    {
      "name": "refs",
      "total": 11,
      "passed": 11,
      "failed": 0,
      "warnings": 0,
      "checks": [
        {
          "id": 1,
          "name": "CLAUDE.md references valid documents",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 2,
          "name": "Agent files reference valid conventions",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 3,
          "name": "Blueprint cross-references resolve",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 4,
          "name": "INDEX.md links resolve to files",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 5,
          "name": "Iteration overview references valid docs",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 6,
          "name": "mind.toml document paths exist on disk",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 7,
          "name": "mind.toml graph references valid document IDs",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 8,
          "name": "No broken markdown links in spec/",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 9,
          "name": "No broken markdown links in blueprints/",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 10,
          "name": "ADR numbering is sequential",
          "level": "WARN",
          "passed": true,
          "message": ""
        },
        {
          "id": 11,
          "name": "Iteration numbering is sequential",
          "level": "WARN",
          "passed": true,
          "message": ""
        }
      ]
    },
    {
      "name": "config",
      "total": 4,
      "passed": 4,
      "failed": 0,
      "warnings": 0,
      "checks": [
        {
          "id": 1,
          "name": "mind.toml is valid TOML",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 2,
          "name": "mind.toml has required fields",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 3,
          "name": "mind.toml document IDs are consistent",
          "level": "FAIL",
          "passed": true,
          "message": ""
        },
        {
          "id": 4,
          "name": "mind.toml graph has no cycles",
          "level": "FAIL",
          "passed": true,
          "message": ""
        }
      ]
    }
  ],
  "summary": {
    "total": 32,
    "passed": 30,
    "failed": 0,
    "warnings": 2
  }
}
```

**Field notes:**

- `suites` is always an array of exactly three entries in the order: `docs`, `refs`, `config`.
- When `--strict` is passed, `level` for WARN checks is promoted to `"FAIL"` and failures increase accordingly.
- `checks[].message` is empty string when the check passes.

### `mind doctor --json`

```json
{
  "diagnostics": [
    {
      "category": "framework",
      "check": "Framework installed",
      "status": "pass",
      "message": ".mind/ directory found",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "framework",
      "check": "Claude Code adapter installed",
      "status": "pass",
      "message": ".claude/ directory found",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "framework",
      "check": "Copilot adapter installed",
      "status": "fail",
      "message": ".github/agents/ directory not found",
      "fix": "Run: mind init --with-github",
      "auto_fixable": true
    },
    {
      "category": "docs",
      "check": "Documentation structure",
      "status": "pass",
      "message": "17/17 checks pass",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "docs",
      "check": "Stub documents",
      "status": "warn",
      "message": "2 stubs found: domain-model.md, glossary.md",
      "fix": "Fill these files or run /discover to generate context",
      "auto_fixable": false
    },
    {
      "category": "refs",
      "check": "Framework cross-references",
      "status": "pass",
      "message": "11/11 checks pass",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "config",
      "check": "Conversation configs valid",
      "status": "pass",
      "message": "4/4 files validated",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "brief",
      "check": "Project brief completeness",
      "status": "warn",
      "message": "Missing section: Key Deliverables",
      "fix": "Add a ## Key Deliverables section to docs/spec/project-brief.md",
      "auto_fixable": false
    },
    {
      "category": "workflow",
      "check": "No stale workflow state",
      "status": "pass",
      "message": "",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "iterations",
      "check": "All iterations have overview.md",
      "status": "pass",
      "message": "6/6 iterations complete",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "naming",
      "check": "File naming conventions",
      "status": "pass",
      "message": "",
      "fix": "",
      "auto_fixable": false
    },
    {
      "category": "sync",
      "check": "Agent sync status",
      "status": "warn",
      "message": "2 agents out of sync with .github/agents/",
      "fix": "Run: mind sync agents",
      "auto_fixable": true
    }
  ],
  "summary": {
    "pass": 9,
    "fail": 1,
    "warn": 2
  }
}
```

**Field notes:**

- `diagnostics[].status` is one of: `"pass"`, `"fail"`, `"warn"`.
- `diagnostics[].fix` is a non-empty string only when the diagnostic has an actionable fix suggestion.
- `diagnostics[].auto_fixable` is `true` only when `mind doctor --fix` can resolve the issue automatically.
- `diagnostics[].category` is one of: `"framework"`, `"docs"`, `"refs"`, `"config"`, `"brief"`, `"workflow"`, `"iterations"`, `"naming"`, `"sync"`.

### `mind docs list --json`

```json
{
  "documents": [
    {
      "path": "docs/spec/project-brief.md",
      "zone": "spec",
      "name": "project-brief",
      "id": "doc:spec/project-brief",
      "status": "active",
      "is_stub": false,
      "size": 3200,
      "mod_time": "2026-03-10T15:00:00Z"
    },
    {
      "path": "docs/spec/requirements.md",
      "zone": "spec",
      "name": "requirements",
      "id": "doc:spec/requirements",
      "status": "active",
      "is_stub": false,
      "size": 8100,
      "mod_time": "2026-03-09T10:00:00Z"
    },
    {
      "path": "docs/spec/architecture.md",
      "zone": "spec",
      "name": "architecture",
      "id": "doc:spec/architecture",
      "status": "active",
      "is_stub": false,
      "size": 6400,
      "mod_time": "2026-03-09T12:00:00Z"
    },
    {
      "path": "docs/spec/domain-model.md",
      "zone": "spec",
      "name": "domain-model",
      "id": "doc:spec/domain-model",
      "status": "draft",
      "is_stub": true,
      "size": 420,
      "mod_time": "2026-03-09T00:13:40Z"
    }
  ],
  "by_zone": {
    "spec": 5,
    "blueprints": 4,
    "state": 2,
    "iterations": 6,
    "knowledge": 3
  },
  "total": 20
}
```

**Field notes:**

- When `--zone ZONE` is provided, `documents` is filtered to that zone only. `by_zone` still shows all zones.
- Documents are sorted by zone (spec, blueprints, state, iterations, knowledge), then by name alphabetically within each zone.
- `total` is the count of all documents across all zones, regardless of any zone filter.

### `mind workflow status --json`

```json
{
  "state": "running",
  "type": "NEW_PROJECT",
  "descriptor": "rest-api",
  "iteration_path": "docs/iterations/007-NEW_PROJECT-rest-api",
  "branch": "new/rest-api",
  "last_agent": "architect",
  "remaining_chain": ["developer", "tester", "reviewer"],
  "session": 1,
  "total_sessions": 2,
  "artifacts": [
    {
      "agent": "analyst",
      "output": "requirements.md",
      "location": "docs/spec/requirements.md"
    },
    {
      "agent": "architect",
      "output": "architecture.md",
      "location": "docs/spec/architecture.md"
    }
  ],
  "dispatch_log": [
    {
      "agent": "analyst",
      "file": "agents/analyst.md",
      "model": "opus",
      "status": "completed",
      "started_at": "2026-03-11T10:05:00Z",
      "duration_ms": 134000
    },
    {
      "agent": "architect",
      "file": "agents/architect.md",
      "model": "opus",
      "status": "completed",
      "started_at": "2026-03-11T10:07:14Z",
      "duration_ms": 181000
    }
  ]
}
```

**Field notes:**

- When the workflow is idle, the output is:
  ```json
  {
    "state": "idle",
    "type": null,
    "descriptor": null,
    "iteration_path": null,
    "branch": null,
    "last_agent": null,
    "remaining_chain": [],
    "session": 0,
    "total_sessions": 0,
    "artifacts": [],
    "dispatch_log": []
  }
  ```
- `state` is one of: `"running"`, `"idle"`, `"paused"`.
- `dispatch_log[].status` is one of: `"dispatched"`, `"completed"`, `"failed"`, `"retrying"`.
- `dispatch_log[].duration_ms` is the wall-clock duration in milliseconds. `0` if the agent has not completed.

### `mind preflight --json`

```json
{
  "type": "NEW_PROJECT",
  "descriptor": "rest-api",
  "chain": ["analyst", "architect", "developer", "tester", "reviewer"],
  "branch": "new/rest-api",
  "iteration": "docs/iterations/007-NEW_PROJECT-rest-api",
  "brief_gate": "BRIEF_PRESENT",
  "brief_sections": {
    "vision": true,
    "key_deliverables": true,
    "scope": true
  },
  "docs_validation": {
    "total": 17,
    "passed": 15,
    "failed": 0,
    "warnings": 2
  },
  "context_files": [
    "docs/spec/project-brief.md",
    "docs/spec/requirements.md",
    "docs/spec/architecture.md",
    "docs/state/current.md"
  ],
  "warnings": [
    "domain-model.md is a stub — analyst should populate it",
    "glossary.md is a stub"
  ],
  "prompt_length": 4820,
  "prompt_preview": "## AGENT INSTRUCTIONS\n\nYou are the analyst agent..."
}
```

**Field notes:**

- `prompt_preview` is the first 200 characters of the assembled prompt. The full prompt is written to `docs/state/workflow.md` and optionally copied to the clipboard.
- `brief_gate` is one of: `"BRIEF_PRESENT"`, `"BRIEF_STUB"`, `"BRIEF_MISSING"`, `"BRIEF_SKIPPED"` (skipped for BUG_FIX type).
- `context_files` lists all files that will be included in the agent prompt context, in the order they appear.

### `mind iterations list --json`

```json
{
  "iterations": [
    {
      "seq": 7,
      "type": "NEW_PROJECT",
      "descriptor": "rest-api",
      "dir_name": "007-NEW_PROJECT-rest-api",
      "path": "docs/iterations/007-NEW_PROJECT-rest-api",
      "status": "in_progress",
      "created_at": "2026-03-11T10:00:00Z",
      "artifacts": [
        {"name": "overview.md", "exists": true, "path": "docs/iterations/007-NEW_PROJECT-rest-api/overview.md"},
        {"name": "changes.md", "exists": false, "path": "docs/iterations/007-NEW_PROJECT-rest-api/changes.md"},
        {"name": "test-summary.md", "exists": false, "path": "docs/iterations/007-NEW_PROJECT-rest-api/test-summary.md"},
        {"name": "validation.md", "exists": false, "path": "docs/iterations/007-NEW_PROJECT-rest-api/validation.md"},
        {"name": "retrospective.md", "exists": false, "path": "docs/iterations/007-NEW_PROJECT-rest-api/retrospective.md"}
      ]
    },
    {
      "seq": 6,
      "type": "ENHANCEMENT",
      "descriptor": "add-caching",
      "dir_name": "006-ENHANCEMENT-add-caching",
      "path": "docs/iterations/006-ENHANCEMENT-add-caching",
      "status": "complete",
      "created_at": "2026-03-08T09:00:00Z",
      "artifacts": [
        {"name": "overview.md", "exists": true, "path": "docs/iterations/006-ENHANCEMENT-add-caching/overview.md"},
        {"name": "changes.md", "exists": true, "path": "docs/iterations/006-ENHANCEMENT-add-caching/changes.md"},
        {"name": "test-summary.md", "exists": true, "path": "docs/iterations/006-ENHANCEMENT-add-caching/test-summary.md"},
        {"name": "validation.md", "exists": true, "path": "docs/iterations/006-ENHANCEMENT-add-caching/validation.md"},
        {"name": "retrospective.md", "exists": true, "path": "docs/iterations/006-ENHANCEMENT-add-caching/retrospective.md"}
      ]
    }
  ],
  "total": 7,
  "by_type": {
    "NEW_PROJECT": 1,
    "ENHANCEMENT": 3,
    "BUG_FIX": 2,
    "REFACTOR": 1
  }
}
```

**Field notes:**

- Iterations are sorted by `seq` descending (newest first).
- `status` is one of: `"in_progress"`, `"complete"`, `"incomplete"`.
- `"complete"` means all 5 expected artifacts exist. `"incomplete"` means some artifacts are missing but the workflow is not active.
- `artifacts` always contains exactly 5 entries in the canonical order.

### `mind quality history --json`

```json
{
  "entries": [
    {
      "date": "2026-03-08",
      "topic": "auth-strategy",
      "session_id": "conv-003",
      "overall_score": 3.8,
      "dimensions": {
        "depth_of_analysis": 4,
        "breadth_of_perspectives": 3,
        "quality_of_synthesis": 4,
        "actionability": 4,
        "intellectual_rigor": 4,
        "creative_insight": 3
      },
      "gate_0_pass": true,
      "personas_used": ["systems-architect", "security-engineer", "dx-advocate"],
      "variant": "v2",
      "output_path": "docs/knowledge/auth-strategy-convergence.md"
    },
    {
      "date": "2026-03-05",
      "topic": "caching-strategy",
      "session_id": "conv-002",
      "overall_score": 3.2,
      "dimensions": {
        "depth_of_analysis": 3,
        "breadth_of_perspectives": 3,
        "quality_of_synthesis": 4,
        "actionability": 3,
        "intellectual_rigor": 3,
        "creative_insight": 3
      },
      "gate_0_pass": true,
      "personas_used": ["systems-architect", "performance-engineer", "backend-engineer"],
      "variant": "v2",
      "output_path": "docs/knowledge/caching-strategy-convergence.md"
    }
  ],
  "total": 2,
  "average_score": 3.5,
  "gate_0_pass_rate": 1.0,
  "trend": "improving"
}
```

**Field notes:**

- Entries are sorted by `date` descending (newest first).
- `trend` is one of: `"improving"`, `"stable"`, `"declining"`, `"insufficient_data"`. Requires at least 3 entries to compute; otherwise returns `"insufficient_data"`.
- `gate_0_pass_rate` is a float between 0.0 and 1.0 representing the ratio of entries where `gate_0_pass` is `true`.
- `dimensions` values are integers from 1-5.

### `mind sync agents --json`

```json
{
  "agents": [
    {
      "name": "analyst",
      "source": ".mind/agents/analyst.md",
      "targets": [
        {
          "platform": "claude",
          "path": ".claude/agents/analyst.md",
          "status": "synced",
          "diff_lines": 0
        },
        {
          "platform": "github",
          "path": ".github/agents/analyst.md",
          "status": "out_of_sync",
          "diff_lines": 12
        }
      ]
    },
    {
      "name": "architect",
      "source": ".mind/agents/architect.md",
      "targets": [
        {
          "platform": "claude",
          "path": ".claude/agents/architect.md",
          "status": "synced",
          "diff_lines": 0
        },
        {
          "platform": "github",
          "path": ".github/agents/architect.md",
          "status": "missing",
          "diff_lines": 0
        }
      ]
    }
  ],
  "summary": {
    "total_agents": 5,
    "synced": 3,
    "out_of_sync": 1,
    "missing": 1
  },
  "action_taken": "dry_run"
}
```

**Field notes:**

- `agents[].targets[].status` is one of: `"synced"`, `"out_of_sync"`, `"missing"`, `"error"`.
- `action_taken` is one of: `"synced"` (changes were written), `"dry_run"` (with `--check` flag, no changes written), `"none"` (everything already in sync).
- `diff_lines` is the number of lines that differ between source and target. `0` when synced or missing.

### `mind reconcile --json`

```json
{
  "manifest_changed": true,
  "previous_generation": 2,
  "current_generation": 3,
  "documents": {
    "total": 10,
    "scanned": 10,
    "stale": 2,
    "missing": 0,
    "undeclared": 1,
    "hash_changed": 3
  },
  "stale_documents": [
    {
      "id": "doc:spec/architecture",
      "path": "docs/spec/architecture.md",
      "reason": "dependency changed: doc:spec/requirements"
    },
    {
      "id": "doc:spec/domain-model",
      "path": "docs/spec/domain-model.md",
      "reason": "dependency changed: doc:spec/architecture (transitive)"
    }
  ],
  "undeclared_files": [
    "docs/spec/api-contracts.md"
  ],
  "lock_file_written": true,
  "lock_file_path": "mind.lock"
}
```

**Field notes:**

- `manifest_changed` is `true` when `mind.toml` has been modified since the last reconcile (detected by comparing `project_hash` in the existing lock file).
- `undeclared_files` lists files in `docs/` that match the `*.md` pattern and are in a valid zone directory but are not registered in `mind.toml`.
- `lock_file_written` is `false` when running with `--dry-run`.

---

## 4. Exit Codes

All commands follow a consistent exit code scheme. Scripts and CI systems can rely on these codes.

| Code | Constant | Meaning | Description |
|------|----------|---------|-------------|
| 0 | `ExitSuccess` | Success | Command completed without errors or validation failures. |
| 1 | `ExitValidationFailure` | Validation failure | One or more checks failed, a gate was not met, or a command-specific operation failed in an expected way. |
| 2 | `ExitUnexpectedError` | Unexpected error | An unhandled error occurred (file I/O failure, parse error, internal bug). |
| 3 | `ExitConfigError` | Configuration error | `mind.toml` is missing, malformed, or has invalid values. Project root not found. |
| 4 | `ExitStaleArtifacts` | Stale artifacts | Reconciliation detected stale documents. Only used by `reconcile` and `status --strict`. |

### Per-Command Exit Code Map

| Command | 0 | 1 | 2 | 3 | 4 |
|---------|---|---|---|---|---|
| `mind status` | Health displayed | -- | Unexpected error | No project found | Stale (with `--strict`) |
| `mind status --strict` | Health displayed, no stale | -- | Unexpected error | No project found | Stale artifacts detected |
| `mind doctor` | All diagnostics pass | Failures found | Unexpected error | No project found | -- |
| `mind doctor --fix` | All issues resolved | Some issues remain | Unexpected error | No project found | -- |
| `mind init` | Project initialized | -- | Unexpected error | -- | -- |
| `mind create *` | Artifact created | -- | Unexpected error | No project found | -- |
| `mind check docs` | All checks pass | Failures found | Unexpected error | No project found | -- |
| `mind check refs` | All checks pass | Failures found | Unexpected error | No project found | -- |
| `mind check config` | All checks pass | Failures found | Unexpected error | Invalid mind.toml | -- |
| `mind check all` | All checks pass | Failures found | Unexpected error | No project found | -- |
| `mind workflow status` | State displayed | -- | Unexpected error | No project found | -- |
| `mind preflight` | Pre-flight complete | Gate not met | Unexpected error | No project found | -- |
| `mind run` | Workflow complete | Agent/gate failure | Unexpected error | No project found | -- |
| `mind handoff` | Handoff complete | Incomplete iteration | Unexpected error | No project found | -- |
| `mind reconcile` | No stale artifacts | -- | Unexpected error | No project found | Stale artifacts detected |
| `mind sync agents` | All synced | Sync failures | Unexpected error | No project found | -- |
| `mind docs list` | Documents listed | -- | Unexpected error | No project found | -- |
| `mind quality log` | Score logged | Extraction failed | Unexpected error | No project found | -- |
| `mind quality history` | History displayed | -- | Unexpected error | No project found | -- |
| `mind serve` | Server stopped cleanly | -- | Unexpected error | No project found | -- |
| `mind watch` | Watcher stopped cleanly | -- | Unexpected error | No project found | -- |
| `mind version` | Version displayed | -- | -- | -- | -- |
| `mind completion` | Completions generated | -- | Unexpected error | -- | -- |

### Exit Code in JSON Output

When `--json` is used, the exit code is still set on the process. The JSON output does not include an explicit exit code field. The caller should check both the process exit code and the JSON content.

---

## 5. MCP Tool Schemas

The MCP server (`mind serve`) exposes 16 tools over stdio using the Model Context Protocol. Each tool is defined here with its name, description, JSON Schema for `inputSchema`, and output format.

All tools return results as JSON-encoded strings in the MCP `content` field with `type: "text"`. The JSON structure matches the corresponding `--json` CLI output where applicable.

### `mind_status`

```json
{
  "name": "mind_status",
  "description": "Get project health summary including documentation completeness, workflow state, and actionable warnings",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as `mind status --json`.

### `mind_doctor`

```json
{
  "name": "mind_doctor",
  "description": "Run deep diagnostics with actionable fix suggestions for all project health aspects",
  "inputSchema": {
    "type": "object",
    "properties": {
      "fix": {
        "type": "boolean",
        "description": "Auto-fix resolvable issues (create missing dirs, fix naming)",
        "default": false
      }
    },
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as `mind doctor --json`.

### `mind_check_brief`

```json
{
  "name": "mind_check_brief",
  "description": "Evaluate the project brief for the business context gate. Returns gate classification and section presence.",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "gate": "BRIEF_PRESENT",
  "exists": true,
  "is_stub": false,
  "path": "docs/spec/project-brief.md",
  "sections": {
    "vision": true,
    "key_deliverables": true,
    "scope": true,
    "constraints": true,
    "success_metrics": false
  },
  "word_count": 482,
  "warnings": [
    "Missing section: Success Metrics"
  ]
}
```

### `mind_validate_docs`

```json
{
  "name": "mind_validate_docs",
  "description": "Run 17-check documentation structure validation across all 5 zones",
  "inputSchema": {
    "type": "object",
    "properties": {
      "strict": {
        "type": "boolean",
        "description": "Promote warnings to failures",
        "default": false
      }
    },
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as a single suite entry from `mind check all --json` (the `docs` suite object).

### `mind_validate_refs`

```json
{
  "name": "mind_validate_refs",
  "description": "Run 11-check cross-reference validation for document links and mind.toml consistency",
  "inputSchema": {
    "type": "object",
    "properties": {
      "strict": {
        "type": "boolean",
        "description": "Promote warnings to failures",
        "default": false
      }
    },
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as a single suite entry from `mind check all --json` (the `refs` suite object).

### `mind_list_iterations`

```json
{
  "name": "mind_list_iterations",
  "description": "List all workflow iterations with type, status, and artifact completeness",
  "inputSchema": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "description": "Filter by request type",
        "enum": ["NEW_PROJECT", "BUG_FIX", "ENHANCEMENT", "REFACTOR", "COMPLEX_NEW"]
      },
      "status": {
        "type": "string",
        "description": "Filter by iteration status",
        "enum": ["in_progress", "complete", "incomplete"]
      },
      "limit": {
        "type": "integer",
        "description": "Maximum number of iterations to return (newest first)",
        "default": 50,
        "minimum": 1,
        "maximum": 200
      }
    },
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as `mind iterations list --json`.

### `mind_show_iteration`

```json
{
  "name": "mind_show_iteration",
  "description": "Show details for a single iteration including all artifacts and their content summaries",
  "inputSchema": {
    "type": "object",
    "properties": {
      "seq": {
        "type": "integer",
        "description": "Iteration sequence number (e.g., 7 for iteration 007)",
        "minimum": 1
      }
    },
    "required": ["seq"],
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "seq": 7,
  "type": "NEW_PROJECT",
  "descriptor": "rest-api",
  "dir_name": "007-NEW_PROJECT-rest-api",
  "path": "docs/iterations/007-NEW_PROJECT-rest-api",
  "status": "in_progress",
  "created_at": "2026-03-11T10:00:00Z",
  "artifacts": [
    {
      "name": "overview.md",
      "exists": true,
      "path": "docs/iterations/007-NEW_PROJECT-rest-api/overview.md",
      "size": 1200,
      "mod_time": "2026-03-11T10:00:00Z"
    },
    {
      "name": "changes.md",
      "exists": true,
      "path": "docs/iterations/007-NEW_PROJECT-rest-api/changes.md",
      "size": 3400,
      "mod_time": "2026-03-11T10:15:00Z"
    },
    {
      "name": "test-summary.md",
      "exists": false,
      "path": "docs/iterations/007-NEW_PROJECT-rest-api/test-summary.md",
      "size": 0,
      "mod_time": ""
    },
    {
      "name": "validation.md",
      "exists": false,
      "path": "docs/iterations/007-NEW_PROJECT-rest-api/validation.md",
      "size": 0,
      "mod_time": ""
    },
    {
      "name": "retrospective.md",
      "exists": false,
      "path": "docs/iterations/007-NEW_PROJECT-rest-api/retrospective.md",
      "size": 0,
      "mod_time": ""
    }
  ],
  "overview_content": "# Iteration 007: NEW_PROJECT — rest-api\n\n..."
}
```

### `mind_read_state`

```json
{
  "name": "mind_read_state",
  "description": "Read the current workflow state from docs/state/workflow.md (parsed, not raw markdown)",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**Output**: Same JSON structure as `mind workflow status --json`.

### `mind_update_state`

```json
{
  "name": "mind_update_state",
  "description": "Write workflow state to docs/state/workflow.md. Used by agents to record their position in the chain.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "last_agent": {
        "type": "string",
        "description": "Name of the agent that just completed"
      },
      "status": {
        "type": "string",
        "description": "Agent completion status",
        "enum": ["completed", "failed", "retrying"]
      },
      "artifact": {
        "type": "object",
        "description": "Artifact produced by the agent",
        "properties": {
          "output": {
            "type": "string",
            "description": "Output filename (e.g., requirements.md)"
          },
          "location": {
            "type": "string",
            "description": "Relative path where the artifact was written"
          }
        },
        "required": ["output", "location"]
      },
      "decisions": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Key decisions made during this agent's execution"
      },
      "handoff_context": {
        "type": "string",
        "description": "Context notes for the next agent in the chain"
      }
    },
    "required": ["last_agent", "status"],
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "success": true,
  "state": "running",
  "last_agent": "architect",
  "remaining_chain": ["developer", "tester", "reviewer"],
  "workflow_file": "docs/state/workflow.md"
}
```

### `mind_create_iteration`

```json
{
  "name": "mind_create_iteration",
  "description": "Create a new iteration folder with all 5 template files and auto-numbered sequence",
  "inputSchema": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "description": "Request type classification",
        "enum": ["NEW_PROJECT", "BUG_FIX", "ENHANCEMENT", "REFACTOR", "COMPLEX_NEW"]
      },
      "descriptor": {
        "type": "string",
        "description": "Kebab-case descriptor (e.g., rest-api, fix-auth-redirect)"
      },
      "request": {
        "type": "string",
        "description": "Original user request text to include in overview.md"
      }
    },
    "required": ["type", "descriptor"],
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "seq": 7,
  "dir_name": "007-NEW_PROJECT-rest-api",
  "path": "docs/iterations/007-NEW_PROJECT-rest-api",
  "branch": "new/rest-api",
  "files_created": [
    "docs/iterations/007-NEW_PROJECT-rest-api/overview.md",
    "docs/iterations/007-NEW_PROJECT-rest-api/changes.md",
    "docs/iterations/007-NEW_PROJECT-rest-api/test-summary.md",
    "docs/iterations/007-NEW_PROJECT-rest-api/validation.md",
    "docs/iterations/007-NEW_PROJECT-rest-api/retrospective.md"
  ]
}
```

### `mind_list_stubs`

```json
{
  "name": "mind_list_stubs",
  "description": "List all documents that are stubs (template-only content that needs to be filled)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "zone": {
        "type": "string",
        "description": "Filter by documentation zone",
        "enum": ["spec", "blueprints", "state", "iterations", "knowledge"]
      }
    },
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "stubs": [
    {
      "path": "docs/spec/domain-model.md",
      "zone": "spec",
      "name": "domain-model",
      "id": "doc:spec/domain-model",
      "size": 420,
      "mod_time": "2026-03-09T00:13:40Z",
      "expected_content": "Entity definitions, relationships, business rules, state machines"
    },
    {
      "path": "docs/knowledge/glossary.md",
      "zone": "knowledge",
      "name": "glossary",
      "id": "doc:knowledge/glossary",
      "size": 180,
      "mod_time": "2026-03-09T00:13:40Z",
      "expected_content": "Domain terminology definitions"
    }
  ],
  "total": 2
}
```

### `mind_check_gate`

```json
{
  "name": "mind_check_gate",
  "description": "Run the deterministic quality gate: execute build, lint, and test commands from mind.toml",
  "inputSchema": {
    "type": "object",
    "properties": {
      "commands": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Specific commands to run. If empty, runs all configured commands (build, lint, test, typecheck)."
      }
    },
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "results": [
    {
      "name": "build",
      "command": "go build -o mind .",
      "passed": true,
      "exit_code": 0,
      "duration_ms": 3200,
      "stdout": "",
      "stderr": ""
    },
    {
      "name": "lint",
      "command": "golangci-lint run ./...",
      "passed": false,
      "exit_code": 1,
      "duration_ms": 8400,
      "stdout": "",
      "stderr": "internal/validate/docs.go:42:5: unused variable 'x' (deadcode)\n"
    },
    {
      "name": "test",
      "command": "go test ./...",
      "passed": true,
      "exit_code": 0,
      "duration_ms": 12300,
      "stdout": "ok  \tgithub.com/jf-ferraz/mind-cli/internal/validate\t0.842s\n",
      "stderr": ""
    }
  ],
  "overall_passed": false,
  "total": 3,
  "passed": 2,
  "failed": 1,
  "total_duration_ms": 23900
}
```

### `mind_log_quality`

```json
{
  "name": "mind_log_quality",
  "description": "Extract quality scores from a convergence analysis file and append to quality-log.yml",
  "inputSchema": {
    "type": "object",
    "properties": {
      "convergence_path": {
        "type": "string",
        "description": "Relative path to the convergence analysis file"
      },
      "topic": {
        "type": "string",
        "description": "Override topic name (default: extracted from filename)"
      },
      "variant": {
        "type": "string",
        "description": "Override variant label (default: extracted from file content)"
      }
    },
    "required": ["convergence_path"],
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "logged": true,
  "entry": {
    "date": "2026-03-11",
    "topic": "auth-strategy",
    "session_id": "conv-004",
    "overall_score": 3.8,
    "dimensions": {
      "depth_of_analysis": 4,
      "breadth_of_perspectives": 3,
      "quality_of_synthesis": 4,
      "actionability": 4,
      "intellectual_rigor": 4,
      "creative_insight": 3
    },
    "gate_0_pass": true,
    "personas_used": ["systems-architect", "security-engineer", "dx-advocate"],
    "variant": "v2",
    "output_path": "docs/knowledge/auth-strategy-convergence.md"
  },
  "quality_log_path": "docs/knowledge/quality-log.yml"
}
```

### `mind_search_docs`

```json
{
  "name": "mind_search_docs",
  "description": "Full-text search across all documentation files in docs/",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query (case-insensitive substring match)"
      },
      "zone": {
        "type": "string",
        "description": "Restrict search to a specific zone",
        "enum": ["spec", "blueprints", "state", "iterations", "knowledge"]
      },
      "limit": {
        "type": "integer",
        "description": "Maximum number of results to return",
        "default": 20,
        "minimum": 1,
        "maximum": 100
      }
    },
    "required": ["query"],
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "results": [
    {
      "path": "docs/spec/architecture.md",
      "zone": "spec",
      "matches": [
        {
          "line_number": 42,
          "line_content": "The authentication layer uses JWT tokens for stateless auth.",
          "context_before": "## Authentication",
          "context_after": "Tokens are signed with RS256 and expire after 24 hours."
        }
      ],
      "total_matches": 3
    }
  ],
  "total_results": 1,
  "query": "JWT"
}
```

### `mind_read_config`

```json
{
  "name": "mind_read_config",
  "description": "Parse and return the mind.toml manifest as structured JSON",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "manifest": {
    "schema": "mind/v2.0",
    "generation": 3,
    "updated": "2026-03-10T00:13:40Z",
    "invariants": {
      "every_document_has_owner": false,
      "no_orphan_dependencies": true,
      "no_circular_dependencies": true
    }
  },
  "project": {
    "name": "mind-cli",
    "description": "CLI and TUI for the Mind Agent Framework",
    "type": "cli",
    "stack": {
      "language": "go@1.23",
      "framework": "cobra+bubbletea",
      "testing": "go-test"
    },
    "commands": {
      "dev": "go run .",
      "test": "go test ./...",
      "lint": "golangci-lint run ./...",
      "typecheck": "",
      "build": "go build -o mind ."
    }
  },
  "profiles": {
    "active": []
  },
  "documents": {
    "doc:spec/project-brief": {
      "id": "doc:spec/project-brief",
      "path": "docs/spec/project-brief.md",
      "zone": "spec",
      "status": "active"
    }
  },
  "governance": {
    "max_retries": 2,
    "review_policy": "evidence-based",
    "commit_policy": "conventional",
    "branch_strategy": "type-descriptor"
  },
  "graph": [
    {
      "from": "doc:spec/requirements",
      "to": "doc:spec/architecture",
      "type": "informs"
    }
  ]
}
```

### `mind_suggest_next`

```json
{
  "name": "mind_suggest_next",
  "description": "Analyze project state and suggest what should happen next. Considers documentation gaps, workflow state, and recent changes.",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**Output**:

```json
{
  "suggestion": "The analyst should proceed with requirements extraction.",
  "reason": "Business context gate passed. Project brief has substantive content. No existing requirements document.",
  "priority": "high",
  "action_type": "workflow",
  "warnings": [
    "domain-model.md is a stub — analyst should create it during requirements phase"
  ],
  "context": {
    "workflow_state": "running",
    "last_agent": "none",
    "brief_gate": "BRIEF_PRESENT",
    "stubs_count": 2,
    "recent_iterations": [
      {
        "seq": 6,
        "type": "ENHANCEMENT",
        "descriptor": "add-caching",
        "status": "complete"
      }
    ]
  }
}
```

**Field notes:**

- `priority` is one of: `"high"`, `"medium"`, `"low"`.
- `action_type` is one of: `"workflow"` (continue or start a workflow), `"fix"` (fix a project issue), `"document"` (fill a stub or create a missing document), `"review"` (review existing content).

---

## 6. Quality Log Format

The quality log tracks convergence analysis scores over time. Located at `docs/knowledge/quality-log.yml`. YAML format with a single top-level `entries` array.

### Schema

```yaml
# docs/knowledge/quality-log.yml
#
# Appended by: mind quality log <convergence-file>
# Read by:     mind quality history, mind quality report, mind_log_quality MCP tool
# Format:      YAML (not JSON) for human readability

entries:
  - date: "2026-03-08"                    # required, string, format: YYYY-MM-DD
                                           # Date the convergence analysis was completed

    topic: "auth-strategy"                 # required, string, kebab-case
                                           # Topic of the convergence analysis
                                           # Default: extracted from filename
                                           #   e.g., auth-strategy-convergence.md → "auth-strategy"

    session_id: "conv-003"                 # required, string
                                           # Unique session identifier
                                           # Format: "conv-{NNN}" auto-incremented

    overall_score: 3.8                     # required, float, range: 1.0-5.0
                                           # Weighted average of 6 dimensions
                                           # Extracted from convergence file's quality assessment section

    dimensions:                            # required, object
      depth_of_analysis: 4                 # required, int, range: 1-5
      breadth_of_perspectives: 3           # required, int, range: 1-5
      quality_of_synthesis: 4              # required, int, range: 1-5
      actionability: 4                     # required, int, range: 1-5
      intellectual_rigor: 4                # required, int, range: 1-5
      creative_insight: 3                  # required, int, range: 1-5

    gate_0_pass: true                      # required, bool
                                           # true when overall_score >= 3.0

    personas_used:                         # required, string array
      - "systems-architect"                # Persona identifiers from the convergence analysis
      - "security-engineer"                # Minimum 3 personas per analysis
      - "dx-advocate"

    variant: "v2"                          # required, string
                                           # Convergence protocol variant
                                           # Current values: "v1", "v2"

    output_path: "docs/knowledge/auth-strategy-convergence.md"
                                           # required, string
                                           # Relative path to the convergence output file

  - date: "2026-03-05"
    topic: "caching-strategy"
    session_id: "conv-002"
    overall_score: 3.2
    dimensions:
      depth_of_analysis: 3
      breadth_of_perspectives: 3
      quality_of_synthesis: 4
      actionability: 3
      intellectual_rigor: 3
      creative_insight: 3
    gate_0_pass: true
    personas_used:
      - "systems-architect"
      - "performance-engineer"
      - "backend-engineer"
    variant: "v2"
    output_path: "docs/knowledge/caching-strategy-convergence.md"
```

### Extraction Rules

When `mind quality log` parses a convergence file, it looks for:

1. **Overall score**: A line matching `Overall.*Score.*:\s*(\d+\.?\d*)/5` (case-insensitive).
2. **Dimension scores**: Lines matching `{dimension_name}.*:\s*(\d)/5` within 20 lines of the overall score.
3. **Personas**: A section matching `## Persona` or `## Participants` containing list items (`- **name**` or `- name`).
4. **Variant**: A line matching `[Vv]ariant.*:\s*(v\d+)` or inferred from file structure.

If extraction fails for any required field, the command exits with code 1 and reports which fields could not be extracted.

---

## 7. Iteration Manifest Format

Each iteration lives in a directory under `docs/iterations/`. The directory name follows a strict pattern. It contains exactly 5 markdown files.

### Directory Naming

```
docs/iterations/{SEQ}-{TYPE}-{descriptor}/
```

| Component | Format | Example |
|-----------|--------|---------|
| `{SEQ}` | Zero-padded 3-digit integer | `007` |
| `{TYPE}` | Uppercase request type | `NEW_PROJECT`, `BUG_FIX`, `ENHANCEMENT`, `REFACTOR`, `COMPLEX_NEW` |
| `{descriptor}` | Kebab-case slug | `rest-api`, `fix-auth-redirect` |

Full example: `docs/iterations/007-NEW_PROJECT-rest-api/`

### Artifact Files

| File | Created By | Purpose |
|------|-----------|---------|
| `overview.md` | CLI (preflight) or orchestrator | Iteration metadata, request classification, scope |
| `changes.md` | Developer agent | Files created/modified, implementation notes |
| `test-summary.md` | Tester agent | Test results, coverage, regressions |
| `validation.md` | Reviewer agent | Review findings (MUST/SHOULD/COULD), sign-off |
| `retrospective.md` | Reviewer agent | Lessons learned, process improvements |

### `overview.md` Structure

```markdown
# Iteration {SEQ}: {TYPE} -- {descriptor}

## Request

{Original user request text}

## Classification

- **Type**: {TYPE}
- **Agent Chain**: {agent1} -> {agent2} -> ... -> {agentN}
- **Branch**: {branch-name}
- **Created**: {YYYY-MM-DD}

## Scope

{Brief description of what this iteration covers}

## Business Context

- **Brief Gate**: {BRIEF_PRESENT | BRIEF_STUB | BRIEF_MISSING | BRIEF_SKIPPED}
- **Docs Validation**: {passed}/{total} checks pass

## Agent Outputs

| Agent | Status | Output | Location |
|-------|--------|--------|----------|
| analyst | completed | requirements.md | docs/spec/requirements.md |
| architect | completed | architecture.md | docs/spec/architecture.md |
| developer | in_progress | -- | -- |
| tester | pending | -- | -- |
| reviewer | pending | -- | -- |

## Decisions

- {Decision 1 made during this iteration}
- {Decision 2}

## Handoff Context

{Notes from the last completed agent for the next agent}
```

### `changes.md` Structure

```markdown
# Changes -- Iteration {SEQ}

## Files Created

| File | Purpose |
|------|---------|
| src/routes/users.go | CRUD endpoints for User entity |
| src/middleware/auth.go | JWT authentication middleware |

## Files Modified

| File | Change |
|------|--------|
| go.mod | Added jwt-go dependency |

## Implementation Notes

{Developer's notes on implementation decisions, trade-offs, and known limitations}
```

### `test-summary.md` Structure

```markdown
# Test Summary -- Iteration {SEQ}

## Test Results

- **Total**: {N}
- **Passed**: {N}
- **Failed**: {N}
- **Skipped**: {N}

## Coverage

{Coverage percentage or summary}

## New Tests

| Test | File | Covers |
|------|------|--------|
| TestUserCreate | internal/user/user_test.go | FR-1: User creation |

## Regressions

{Any regressions found, or "None"}
```

### `validation.md` Structure

```markdown
# Validation -- Iteration {SEQ}

## Review Summary

**Reviewer**: {reviewer agent}
**Date**: {YYYY-MM-DD}
**Verdict**: {APPROVED | APPROVED_WITH_CONCERNS | NEEDS_REVISION}

## Findings

### MUST (blocking)

- {Finding that must be addressed before merge}

### SHOULD (recommended)

- {Finding that should be addressed but is not blocking}

### COULD (suggestions)

- {Nice-to-have improvements}

## Gate Results

| Gate | Result |
|------|--------|
| Build | PASS |
| Lint | PASS |
| Test | PASS |
| Micro-Gate A | PASS |
| Micro-Gate B | PASS |
| Deterministic Gate | PASS |

## Sign-Off

{Reviewer's final assessment and recommendation}
```

### `retrospective.md` Structure

```markdown
# Retrospective -- Iteration {SEQ}

## What Went Well

- {Positive outcome 1}

## What Could Improve

- {Area for improvement 1}

## Process Notes

- {Observation about the workflow, agent behavior, or framework usage}

## Recommendations

- {Specific recommendation for future iterations}
```

---

## 8. Versioning & Migration

### Schema Version Format

| Artifact | Version Format | Current | Location |
|----------|---------------|---------|----------|
| `mind.toml` | `mind/vX.Y` (semver-like) | `mind/v2.0` | `manifest.schema` field |
| `mind.lock` | `X.Y` (simple) | `1.0` | `schema_version` field |
| MCP protocol | MCP specification version | `2024-11-05` | MCP `protocolVersion` in init |
| Quality log | Implicit (no version field) | v1 | Determined by field presence |

### Backward Compatibility

The CLI reads older schema versions with the following behavior:

| Scenario | Behavior |
|----------|----------|
| `mind.toml` with `mind/v1.x` schema | Reads successfully. Missing fields use defaults. Emits a deprecation warning recommending `mind migrate`. |
| `mind.toml` with unknown `mind/v3.x` schema | Reads successfully. Unknown fields are preserved. Emits a warning: "mind.toml uses schema mind/v3.x which is newer than this CLI version." |
| `mind.lock` with schema `1.0` | Reads successfully. This is the only version. |
| `mind.lock` with unknown schema | Deletes the lock file and regenerates. Emits a warning. |
| Missing `mind.toml` fields | Each field has a documented default value. Missing fields are filled with defaults at parse time. |
| Extra `mind.toml` fields | Preserved on read. If the CLI writes the file, unknown fields are retained (round-trip safe). |

### Forward Compatibility

| Rule | Description |
|------|-------------|
| Unknown TOML keys preserved | When mind-cli reads and writes `mind.toml`, any keys it does not recognize are kept in place. This allows newer schema versions to add fields without older CLI versions stripping them. |
| Unknown JSON keys ignored | When reading `mind.lock`, unknown keys are silently ignored (standard `json.Unmarshal` behavior in Go with struct tags). |
| YAML append-only | `quality-log.yml` is append-only. New fields added to entries do not break older readers — they are simply ignored. |

### Migration Strategy

Future `mind migrate` command (not in initial release):

```
mind migrate [--dry-run]
```

Behavior:

1. Read `mind.toml` and detect the current schema version.
2. Apply migration steps sequentially from current to latest:
   - `mind/v1.0` to `mind/v1.1`: Add `[governance]` section with defaults.
   - `mind/v1.1` to `mind/v2.0`: Restructure `[documents]` from flat to nested. Add `[manifest.invariants]`. Add `[[graph]]`.
3. Bump `manifest.schema` to the latest version.
4. Increment `manifest.generation`.
5. Update `manifest.updated`.
6. With `--dry-run`: print the diff without writing.
7. Without `--dry-run`: write the file and report changes.

Migration is idempotent. Running `mind migrate` on an already-current schema produces no changes.

### Round-Trip Safety

The CLI uses `github.com/pelletier/go-toml/v2` with its document-preserving API to ensure:

- Comments are retained.
- Key ordering is preserved.
- Unknown fields survive read-write cycles.
- Whitespace and formatting are maintained as closely as possible.

This is critical because users may hand-edit `mind.toml` with comments explaining their choices. The CLI must not strip those comments on write.

---

## Scope Boundaries

**This document defines**: The exact shape of every external data format the CLI reads, writes, or exposes. Field names, types, validation rules, enumerations, and example values.

**This document does not define**:

- How schemas are parsed (implementation detail -- see [03-architecture.md](03-architecture.md) for Go type definitions and repository patterns).
- Why fields exist or what business rules they encode (domain knowledge -- see `docs/spec/domain-model.md`).
- How commands use these schemas internally (control flow -- see [01-mind-cli.md](01-mind-cli.md) for command descriptions and [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md) for workflow integration).

---

> **See also:**
> - [01-mind-cli.md](01-mind-cli.md) -- CLI command tree and TUI design
> - [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md) -- MCP server and workflow integration models
> - [03-architecture.md](03-architecture.md) -- Go type definitions, design patterns, service interfaces
> - `docs/spec/domain-model.md` -- Domain entities and business rules
> - `docs/spec/api-contracts.md` -- Higher-level API contract summary
