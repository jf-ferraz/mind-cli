# Blueprint: CLI Specification

> What does every command do, exactly? Complete behavioral specification for all `mind` CLI commands.

**Status**: Active
**Date**: 2026-03-11
**Depends on**: [01-mind-cli.md](01-mind-cli.md), [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md), [03-architecture.md](03-architecture.md)

---

## 1. Command Tree

```
mind
├── status                              # Project health dashboard
├── init [--name] [--with-github]       # Initialize framework
│   └── --from-existing                 # Detect existing docs
├── doctor [--fix] [--json]             # Deep diagnostics
├── create                              # Create framework artifacts
│   ├── adr "<title>"
│   ├── blueprint "<title>"
│   ├── iteration <type> "<name>"
│   ├── spike "<title>"
│   ├── convergence "<title>"
│   └── brief
├── docs                                # Document management
│   ├── list [--zone ZONE]
│   ├── tree
│   ├── open <path-or-id>
│   ├── stubs
│   └── search "<query>"
├── check                               # Validation suite
│   ├── docs [--strict]
│   ├── refs
│   ├── config
│   ├── convergence <file>
│   └── all [--strict]
├── workflow                            # Workflow state
│   ├── status
│   ├── history
│   ├── show <iteration-id>
│   └── clean [--dry-run]
├── sync                                # Platform sync
│   └── agents [--check]
├── quality                             # Quality tracking
│   ├── log <convergence-file> [--topic] [--variant]
│   ├── history
│   └── report
├── reconcile [--force]                 # Reconciliation engine
├── tui                                 # Interactive TUI
├── preflight "<request>"               # Pre-flight check (Model A)
├── preflight --resume                  # Resume interrupted workflow
├── handoff <iteration-id>              # Post-workflow cleanup (Model A)
├── serve                               # MCP server (Model B)
├── watch [--tui]                       # Filesystem watcher (Model C)
├── run "<request>"                     # Full orchestration (Model D)
│   ├── --resume
│   ├── --dry-run
│   └── --tui
├── completion <shell>                  # Shell completions
├── version [--short]                   # Version info
└── help [command]                      # Help
```

---

## 2. Global Flags

These flags are available on every command via `rootCmd.PersistentFlags()`:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | `false` | Output in JSON format (suppresses styled output) |
| `--no-color` | | bool | `false` | Disable ANSI color codes in output |
| `--verbose` | `-v` | bool | `false` | Enable debug logging to stderr |
| `--project-root` | | string | auto-detect | Override project root path (skip `.mind/` walk-up) |
| `--log-file` | | string | `""` | Write debug logs to a file instead of stderr |

**Auto-detection behavior**: When `--project-root` is not specified, the CLI walks up from `$PWD` looking for a `.mind/` directory. If none is found and the command requires a project, exit code 3 is returned.

**Output mode selection**: The CLI detects three output modes automatically:
1. **Interactive** (default when stdout is a TTY): Styled with Lip Gloss colors, progress bars, box drawing
2. **Plain** (stdout is piped or `--no-color`): Clean text, no ANSI escape codes
3. **JSON** (`--json`): Machine-readable JSON, one object per invocation

---

## 3. Per-Command Specifications

### `mind status`

**Synopsis**: `mind status [flags]`

**Description**: Show the project health dashboard. This is the hero command -- the most common entry point. When the CLI is invoked with no subcommand and stdout is a TTY, `mind status` runs by default.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root (walk up from `$PWD` looking for `.mind/`)
2. Parse `mind.toml` for project metadata (name, stack, framework version)
3. Scan `docs/` for document completeness per zone (spec, blueprints, state, iterations, knowledge)
4. Run stub detection on each document -- classify as complete or stub
5. Parse `docs/spec/project-brief.md` for business context gate (Vision, Key Deliverables, Scope sections)
6. Read `docs/state/workflow.md` for active workflow state
7. Find the latest iteration directory in `docs/iterations/`
8. If `mind.lock` exists, run reconciliation staleness check
9. Aggregate warnings (stubs, missing files, stale state) and suggestions
10. Render output based on mode (interactive/plain/JSON)

**Output (Interactive)**:
```
╭─ Mind Framework ─────────────────────────────────────────────────────────╮
│                                                                          │
│  Project: mind-cli                    Framework: v2026-03-09             │
│  Root: ~/dev/projects/mind-cli/       Branch: main                       │
│                                                                          │
│  Documentation Health                                                    │
│  ───────────────────                                                     │
│  spec/         ████████░░  4/5   brief ✓  reqs ✓  arch ✓  domain ✗     │
│  blueprints/   ██████████  3/3   INDEX ✓  + 2 blueprints               │
│  state/        █████░░░░░  1/2   current ✓  workflow ✗                  │
│  iterations/   ██████████  6/6   all complete                           │
│  knowledge/    ██████░░░░  3/5   glossary ✗  2 spikes  1 convergence   │
│                                                                          │
│  Workflow: idle (no active workflow)                                      │
│  Last: 006-ENHANCEMENT-add-caching (completed 2026-03-08)               │
│                                                                          │
│  Warnings                                                                │
│  ────────                                                                │
│  ⚠ domain-model.md is a stub (needs content)                            │
│  ⚠ glossary.md missing                                                   │
│  ⚠ No workflow state saved                                               │
│                                                                          │
│  Tip: Run 'mind doctor' for detailed diagnostics                         │
╰──────────────────────────────────────────────────────────────────────────╯
```

**Output (Plain)**: Same content without box drawing, colors, or progress bars. Zone health rendered as `spec/ 4/5 [########..]`.

**Output (JSON)**: A `ProjectHealth` object. See BP-03, `domain/health.go` for the schema: project metadata, zone health map, workflow state, last iteration, warnings array, and suggestions array.

**Exit Codes**:
- `0` -- Project is healthy (no failures)
- `1` -- Issues found (stubs, missing documents, stale state)
- `3` -- Not a Mind project (`.mind/` not found)

**Edge Cases**:
- No `mind.toml` present: Status still renders what it can (zone scan, stub detection), with a warning that `mind.toml` is missing.
- Empty `docs/` directory: All zones show 0/N, suggestions include running `mind init` or `mind doctor --fix`.
- No iterations exist: Iterations section shows "none" rather than an error.

**Examples**:
```bash
mind status              # Interactive dashboard
mind status --json       # JSON output for scripting
mind status | head -5    # Plain text (pipe detected)
```

---

### `mind init`

**Synopsis**: `mind init [flags]`

**Description**: Initialize the Mind Agent Framework in the current directory. Creates `.mind/`, `docs/` zone structure, `mind.toml`, adapter files for Claude Code (`.claude/`), and optionally GitHub Copilot (`.github/agents/`).

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--name` | `-n` | string | directory name | Project name for `mind.toml` |
| `--with-github` | | bool | `false` | Also create `.github/agents/` adapter |
| `--from-existing` | | bool | `false` | Detect existing `docs/` content and preserve it |

**Behavior**:
1. Check if `.mind/` already exists -- if so, abort with "already initialized" message
2. Determine project name: `--name` flag, or fall back to the current directory name
3. Create `.mind/` directory with framework boilerplate (agents, conventions, conversation configs)
4. Create `docs/` zone structure: `docs/spec/`, `docs/blueprints/`, `docs/state/`, `docs/iterations/`, `docs/knowledge/`
5. Create stub documents for each zone (project-brief.md, requirements.md, architecture.md, domain-model.md, current.md, workflow.md, glossary.md, blueprints/INDEX.md)
6. Generate `mind.toml` with project metadata, document registry, and default governance settings
7. Create `.claude/CLAUDE.md` adapter that routes to `.mind/CLAUDE.md`
8. If `--with-github`: create `.github/agents/` with synced agent definitions
9. If `--from-existing`: scan for existing `docs/` files, skip creating stubs for files that already exist, register discovered files in `mind.toml`
10. Print summary of created files

**Output (Interactive)**:
```
Initializing Mind Framework in ~/dev/projects/my-app/

  Created .mind/                         (framework root)
  Created docs/spec/                     (5 stub documents)
  Created docs/blueprints/INDEX.md       (blueprint registry)
  Created docs/state/current.md          (project state)
  Created docs/state/workflow.md         (workflow state)
  Created docs/knowledge/                (knowledge zone)
  Created mind.toml                      (project manifest)
  Created .claude/CLAUDE.md              (Claude Code adapter)

✓ Framework initialized: my-app
  Run 'mind status' to see project health
  Run 'mind create brief' to fill the project brief
```

**Output (JSON)**: Object with `project_name`, `root`, `files_created` (string array), `from_existing` (bool), `existing_preserved` (string array if applicable).

**Exit Codes**:
- `0` -- Initialization successful
- `1` -- Initialization failed (permissions, disk full)
- `2` -- Already initialized (`.mind/` exists)

**Edge Cases**:
- `.mind/` already exists: Exit with code 2 and message suggesting `mind doctor --fix` if something is broken.
- `--from-existing` with no existing docs: Behaves identically to normal init.
- `--from-existing` with partial docs: Preserves existing files, creates stubs only for missing ones, registers all in `mind.toml`.
- Non-writable directory: Exit code 1 with permission error.

**Examples**:
```bash
mind init                              # Initialize with auto-detected name
mind init --name my-service            # Initialize with explicit name
mind init --with-github                # Include GitHub Copilot adapter
mind init --from-existing              # Preserve existing documentation
```

---

### `mind doctor`

**Synopsis**: `mind doctor [flags]`

**Description**: Run deep diagnostics on the project and produce actionable fix suggestions. Goes beyond `status` by running all validators, cross-referencing results, and identifying root causes.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--fix` | | bool | `false` | Auto-fix resolvable issues (create missing dirs, stubs, fix naming) |
| `--json` | `-j` | bool | `false` | Machine-readable output |

**Behavior**:
1. Detect project root
2. Check framework installation (`.mind/` present, structure intact)
3. Check adapter installations (`.claude/` present, `.github/agents/` present)
4. Run 17-check documentation validation (same engine as `mind check docs`)
5. Run 11-check cross-reference validation (same engine as `mind check refs`)
6. Validate conversation YAML configs (same engine as `mind check config`)
7. Check project brief for completeness (all required sections present with content)
8. Check for stub documents and classify severity
9. Check workflow state for consistency (no orphan state, no stale locks)
10. Check iteration completeness (all iterations have expected artifacts)
11. Aggregate results into pass/fail/warning counts with remediation advice
12. If `--fix`: create missing directories, add `.gitkeep` files, create stub documents from templates, fix naming convention violations (e.g., rename `My Blueprint.md` to `my-blueprint.md`)

**Output (Interactive)**:
```
Running diagnostics...

✓ Framework installed (.mind/ present)
✓ Claude Code adapter installed (.claude/ present)
✗ Copilot adapter not found (.github/agents/ missing)
  → Run: mind init --with-github

✓ Documentation structure (17/17 checks pass)
✗ 2 stub documents found:
  → docs/spec/domain-model.md — needs entity definitions
  → docs/knowledge/glossary.md — needs domain terms
  Fix: Fill these files or run /discover to generate context

✓ Framework cross-references (11/11 checks pass)
✓ Conversation configs valid (4/4 files)

⚠ Project brief present but missing "Key Deliverables" section
  → The business context gate will warn on ENHANCEMENT workflows
  → Fix: Add a ## Key Deliverables section to docs/spec/project-brief.md

✓ No stale workflow state
✓ All iterations have overview.md

Summary: 9 pass, 1 fail, 2 warnings
Run 'mind doctor --fix' to auto-fix resolvable issues
```

**Output (JSON)**: Object with `checks` (array of `{name, passed, level, message, fix}`), `summary` (`{pass, fail, warn}`), `fixes_applied` (array, only present with `--fix`).

**Exit Codes**:
- `0` -- All checks pass
- `1` -- Failures found
- `3` -- Not a Mind project

**Edge Cases**:
- `--fix` with no fixable issues: Prints "Nothing to fix" and exits 0.
- `--fix` where some fixes fail (e.g., permission denied): Reports partial fix, exits 1 with details.
- Run outside a project with `--fix`: Suggests `mind init` instead.

**Examples**:
```bash
mind doctor                # Run diagnostics
mind doctor --fix          # Auto-fix resolvable issues
mind doctor --json         # Machine-readable output for CI
mind doctor --fix --json   # Fix and report results as JSON
```

---

### `mind create adr`

**Synopsis**: `mind create adr "<title>" [flags]`

**Description**: Create a new Architecture Decision Record with auto-numbering. Generates a file in `docs/spec/decisions/` using the ADR template.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `title` | Yes | The ADR title (e.g., "Use PostgreSQL for persistence") |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Ensure `docs/spec/decisions/` exists (create if not)
3. Scan existing ADRs for the next sequence number (e.g., if `001-*.md` and `002-*.md` exist, next is `003`)
4. Slugify the title: lowercase, replace spaces with hyphens, strip special characters
5. Generate filename: `{NNN}-{slug}.md` (e.g., `003-use-postgresql.md`)
6. Render ADR template with title, date, and status "Proposed"
7. Write file to disk
8. Print confirmation with the file path

**Output (Interactive)**:
```
✓ Created docs/spec/decisions/003-use-postgresql.md
```

**Output (JSON)**: `{ "path": "docs/spec/decisions/003-use-postgresql.md", "seq": 3, "title": "Use PostgreSQL for persistence" }`

**Exit Codes**:
- `0` -- ADR created
- `1` -- Creation failed
- `3` -- Not a Mind project

**Edge Cases**:
- No existing ADRs: Starts at sequence 001.
- Title with special characters: Stripped during slugification (`"Use PostgreSQL (v15+)"` becomes `use-postgresql-v15`).
- Title argument missing: Cobra reports the required argument error.
- Gap in sequence numbers (001, 003 exist): Next is 004 (max + 1), not 002.

**Examples**:
```bash
mind create adr "Use PostgreSQL for persistence"
mind create adr "Adopt event sourcing for audit trail"
mind create adr "Use JWT for API authentication" --json
```

---

### `mind create blueprint`

**Synopsis**: `mind create blueprint "<title>" [flags]`

**Description**: Create a new blueprint document with auto-numbering and update `docs/blueprints/INDEX.md` with a registry entry.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `title` | Yes | The blueprint title (e.g., "Authentication System") |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Scan `docs/blueprints/` for the next sequence number (ignore `INDEX.md`)
3. Slugify the title to a kebab-case descriptor
4. Generate filename: `{NN}-{slug}.md` (e.g., `04-authentication-system.md`)
5. Render blueprint template with title, date, status "Proposal", and placeholder sections
6. Write blueprint file
7. Append entry to `docs/blueprints/INDEX.md` Active Blueprints table
8. Print confirmation

**Output (Interactive)**:
```
✓ Created docs/blueprints/04-authentication-system.md
✓ Updated docs/blueprints/INDEX.md
```

**Output (JSON)**: `{ "path": "docs/blueprints/04-authentication-system.md", "seq": 4, "title": "Authentication System", "index_updated": true }`

**Exit Codes**:
- `0` -- Blueprint created
- `1` -- Creation failed
- `3` -- Not a Mind project

**Edge Cases**:
- `INDEX.md` missing: Create it from template before appending.
- `INDEX.md` has no Active Blueprints table: Append the table header and row.

**Examples**:
```bash
mind create blueprint "Authentication System"
mind create blueprint "Data Migration Strategy"
```

---

### `mind create iteration`

**Synopsis**: `mind create iteration <type> "<name>" [flags]`

**Description**: Create a new iteration tracking folder with 5 template files (overview.md, changes.md, test-summary.md, validation.md, retrospective.md) and auto-numbering.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `type` | Yes | Iteration type: `new`, `enhancement`, `bugfix`, `refactor` |
| `name` | Yes | Descriptive name (e.g., "add user authentication") |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Validate the type argument against allowed values (`new`, `enhancement`, `bugfix`, `refactor`)
3. Map short types to canonical names: `new` -> `NEW_PROJECT`, `enhancement` -> `ENHANCEMENT`, `bugfix` -> `BUG_FIX`, `refactor` -> `REFACTOR`
4. Scan `docs/iterations/` for the next sequence number
5. Slugify the name to a kebab-case descriptor
6. Generate directory name: `{NNN}-{TYPE}-{slug}` (e.g., `007-NEW_PROJECT-rest-api`)
7. Create the directory
8. Render and write 5 template files: `overview.md`, `changes.md`, `test-summary.md`, `validation.md`, `retrospective.md`
9. Populate `overview.md` with type, name, date, and placeholder sections
10. Print confirmation with directory path

**Output (Interactive)**:
```
✓ Created docs/iterations/007-NEW_PROJECT-rest-api/
  ├── overview.md
  ├── changes.md
  ├── test-summary.md
  ├── validation.md
  └── retrospective.md
```

**Output (JSON)**: `{ "path": "docs/iterations/007-NEW_PROJECT-rest-api", "seq": 7, "type": "NEW_PROJECT", "descriptor": "rest-api", "files": ["overview.md", "changes.md", "test-summary.md", "validation.md", "retrospective.md"] }`

**Exit Codes**:
- `0` -- Iteration created
- `1` -- Creation failed or invalid type
- `3` -- Not a Mind project

**Edge Cases**:
- Invalid type: Exit code 1 with message listing valid types.
- No existing iterations: Starts at sequence 001.
- Directory already exists with the same name: Abort with message suggesting a different name.

**Examples**:
```bash
mind create iteration new "REST API with auth"
mind create iteration enhancement "add caching layer"
mind create iteration bugfix "fix 500 on user endpoint"
mind create iteration refactor "extract repository pattern"
```

---

### `mind create spike`

**Synopsis**: `mind create spike "<title>" [flags]`

**Description**: Create a technical spike report template in `docs/knowledge/`.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `title` | Yes | The spike title (e.g., "Evaluate Redis vs Memcached") |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Ensure `docs/knowledge/` exists
3. Slugify the title
4. Generate filename: `{slug}-spike.md` (e.g., `evaluate-redis-vs-memcached-spike.md`)
5. Render spike template with title, date, hypothesis, methodology, findings, and conclusion sections
6. Write file
7. Print confirmation

**Output (Interactive)**:
```
✓ Created docs/knowledge/evaluate-redis-vs-memcached-spike.md
```

**Output (JSON)**: `{ "path": "docs/knowledge/evaluate-redis-vs-memcached-spike.md", "title": "Evaluate Redis vs Memcached" }`

**Exit Codes**:
- `0` -- Spike created
- `1` -- Creation failed
- `3` -- Not a Mind project

**Edge Cases**:
- File already exists with the same slug: Abort with message.
- `docs/knowledge/` does not exist: Create it.

**Examples**:
```bash
mind create spike "Evaluate Redis vs Memcached"
mind create spike "WebSocket library comparison"
```

---

### `mind create convergence`

**Synopsis**: `mind create convergence "<title>" [flags]`

**Description**: Create a convergence analysis template in `docs/knowledge/`.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `title` | Yes | The convergence topic (e.g., "Authentication strategy") |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Ensure `docs/knowledge/` exists
3. Slugify the title
4. Generate filename: `{slug}-convergence.md` (e.g., `authentication-strategy-convergence.md`)
5. Render convergence template with title, date, persona sections, synthesis, and quality rubric
6. Write file
7. Print confirmation

**Output (Interactive)**:
```
✓ Created docs/knowledge/authentication-strategy-convergence.md
```

**Output (JSON)**: `{ "path": "docs/knowledge/authentication-strategy-convergence.md", "title": "Authentication strategy" }`

**Exit Codes**:
- `0` -- Convergence document created
- `1` -- Creation failed
- `3` -- Not a Mind project

**Edge Cases**:
- File already exists: Abort with message.
- `docs/knowledge/` does not exist: Create it.

**Examples**:
```bash
mind create convergence "Authentication strategy"
mind create convergence "Database selection"
```

---

### `mind create brief`

**Synopsis**: `mind create brief [flags]`

**Description**: Interactive guided creation of `docs/spec/project-brief.md`. Prompts the user for Vision, Key Deliverables, Scope (in/out), and Constraints, then writes a complete brief that passes the business context gate.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Check if `docs/spec/project-brief.md` already exists and has non-stub content -- if so, confirm overwrite
3. Launch interactive prompts (Bubble Tea text input):
   - Vision (1-3 sentences)
   - Key Deliverables (comma-separated list)
   - In Scope (free text)
   - Out of Scope (free text)
   - Constraints (free text, optional)
4. Render the brief template with user responses
5. Write to `docs/spec/project-brief.md`
6. Run business context gate validation on the written file
7. Print confirmation with gate result

**Output (Interactive)**:
```
Creating project brief: docs/spec/project-brief.md

Vision — What does this project do? (1-3 sentences)
> A declarative system configuration manager for Arch Linux

Key Deliverables — What are the concrete outputs? (comma-separated)
> CLI tool, TOML config DSL, Lua scripting, TUI interface

Scope — What is IN scope?
> Package management, service management, dotfile sync, snapshots

Scope — What is explicitly OUT of scope?
> Multi-distro support, GUI, cloud deployment

Constraints — Any technical or business constraints?
> Must work offline, Rust only, single binary

✓ Created docs/spec/project-brief.md
  Business context gate: PASS (Vision ✓, Key Deliverables ✓, Scope ✓)
```

**Output (JSON)**: Not applicable -- this command requires interactive input. If `--json` is passed, print an error directing the user to run without `--json`.

**Exit Codes**:
- `0` -- Brief created, gate passes
- `1` -- Brief created but gate check failed (missing section content)
- `3` -- Not a Mind project

**Edge Cases**:
- Existing non-stub brief: Prompt for overwrite confirmation before proceeding.
- User cancels (Ctrl+C): No file written, exit 1.
- Stdin is not a TTY (piped): Error message suggesting direct file editing instead.

**Examples**:
```bash
mind create brief              # Interactive guided creation
```

---

### `mind docs list`

**Synopsis**: `mind docs list [--zone ZONE] [flags]`

**Description**: List all documents in the project, grouped by documentation zone. Includes file size, modification date, and stub status.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--zone` | `-z` | string | `""` (all) | Filter to a specific zone: spec, blueprints, state, iterations, knowledge |

**Behavior**:
1. Detect project root
2. Scan `docs/` recursively for `.md` files
3. Classify each file into its zone based on path
4. Run stub detection on each file
5. Sort by zone, then by path within zone
6. If `--zone` is specified, filter to only that zone
7. Render list with columns: path, status (complete/stub), mod date, size

**Output (Interactive)**:
```
spec/
  project-brief.md          ✓ complete    2026-03-01     3.2 KB
  requirements.md           ✓ complete    2026-03-05     8.1 KB
  architecture.md           ✓ complete    2026-03-05     6.4 KB
  domain-model.md           ⚠ stub        2026-02-28     0.4 KB
  decisions/
    001-use-postgresql.md   ✓ complete    2026-03-02     1.8 KB
    002-jwt-auth.md         ✓ complete    2026-03-04     2.1 KB

blueprints/
  INDEX.md                  ✓ complete    2026-03-05     0.9 KB
  01-mind-cli.md            ✓ complete    2026-03-03     4.2 KB
  02-ai-workflow-bridge.md  ✓ complete    2026-03-04     3.7 KB

Total: 18 documents (16 complete, 2 stubs)
```

**Output (JSON)**: Array of document objects with `path`, `zone`, `name`, `status`, `is_stub`, `size_bytes`, `mod_time`.

**Exit Codes**:
- `0` -- Success
- `1` -- Invalid zone name
- `3` -- Not a Mind project

**Edge Cases**:
- Empty zone: Zone header rendered with "(empty)" note.
- Invalid `--zone` value: Error listing valid zone names.
- No documents at all: Print "No documents found. Run 'mind init' or 'mind doctor --fix'."

**Examples**:
```bash
mind docs list                   # All zones
mind docs list --zone spec       # Only spec zone
mind docs list --zone iterations # Only iterations
mind docs list --json            # JSON for scripting
```

---

### `mind docs tree`

**Synopsis**: `mind docs tree [flags]`

**Description**: Display a visual tree of all documentation files, similar to the `tree` command but aware of zones and stub status.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Scan `docs/` recursively
3. Build a tree structure respecting directory nesting
4. Annotate each file with stub status
5. Render as an indented tree with `├──`, `└──`, `│` connectors

**Output (Interactive)**:
```
docs/
├── spec/
│   ├── project-brief.md          ✓
│   ├── requirements.md           ✓
│   ├── architecture.md           ✓
│   ├── domain-model.md           ⚠ stub
│   ├── api-contracts.md          ⚠ stub
│   └── decisions/
│       ├── 001-use-postgresql.md ✓
│       └── 002-jwt-auth.md       ✓
├── blueprints/
│   ├── INDEX.md                  ✓
│   ├── 01-mind-cli.md            ✓
│   └── 02-ai-workflow-bridge.md  ✓
├── state/
│   ├── current.md                ✓
│   └── workflow.md               ✓
├── iterations/
│   ├── 001-NEW_PROJECT-initial/
│   │   ├── overview.md           ✓
│   │   ├── changes.md            ✓
│   │   ├── test-summary.md       ✓
│   │   ├── validation.md         ✓
│   │   └── retrospective.md      ✓
│   └── ...
└── knowledge/
    ├── glossary.md               ⚠ stub
    └── auth-convergence.md       ✓
```

**Output (JSON)**: Nested object mirroring directory structure, each file node containing `name`, `is_stub`, `status`.

**Exit Codes**:
- `0` -- Success
- `3` -- Not a Mind project

**Edge Cases**:
- No `docs/` directory: Print message suggesting `mind init`.
- Very deep nesting: Tree truncates at 6 levels with "..." indicator.

**Examples**:
```bash
mind docs tree               # Visual tree
mind docs tree --json        # Structured JSON tree
mind docs tree --no-color    # Plain text tree
```

---

### `mind docs open`

**Synopsis**: `mind docs open <path-or-id> [flags]`

**Description**: Open a document in `$EDITOR`. Accepts a relative path, a document ID from `mind.toml` (e.g., `doc:spec/project-brief`), or a fuzzy partial match.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `path-or-id` | Yes | File path relative to project root, document ID, or partial name |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Resolve the argument:
   a. If it starts with `doc:`, look up in `mind.toml` document registry
   b. If it contains `/`, treat as a relative path from project root
   c. Otherwise, fuzzy-match against all document names (case-insensitive substring)
3. If fuzzy match yields multiple results, list them and ask the user to be more specific
4. If exactly one match, open the file with `$EDITOR` (fall back to `$VISUAL`, then `vi`)
5. Wait for editor to close

**Output (Interactive)**: Opens the editor. If disambiguation is needed:
```
Multiple matches for "brief":
  1. docs/spec/project-brief.md
  2. docs/knowledge/api-brief-spike.md

Specify a more precise path or use the full document ID.
```

**Output (JSON)**: `{ "path": "docs/spec/project-brief.md", "abs_path": "/home/user/project/docs/spec/project-brief.md" }` (does not open editor in JSON mode)

**Exit Codes**:
- `0` -- Editor opened and closed
- `1` -- Document not found, ambiguous match, or no `$EDITOR` set
- `3` -- Not a Mind project

**Edge Cases**:
- `$EDITOR` not set and `$VISUAL` not set: Error with message to set `$EDITOR`.
- Document ID not found in `mind.toml`: Fall back to fuzzy path match.
- Ambiguous match in non-interactive mode (piped): List matches and exit 1.

**Examples**:
```bash
mind docs open docs/spec/project-brief.md    # By path
mind docs open doc:spec/project-brief        # By document ID
mind docs open brief                         # Fuzzy match
mind docs open glossary                      # Fuzzy match
```

---

### `mind docs stubs`

**Synopsis**: `mind docs stubs [flags]`

**Description**: List all documents classified as stubs (contain only headings, HTML comments, and placeholder text with no substantive content).

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Scan all documents in `docs/`
3. Run stub detection on each (a document is a stub if it contains only headings, HTML comments, and template placeholders)
4. List stubs grouped by zone with remediation hints

**Output (Interactive)**:
```
Stub documents (need content):

  spec/domain-model.md         Needs: entity definitions, relationships, business rules
  spec/api-contracts.md        Needs: endpoint definitions, request/response shapes
  knowledge/glossary.md        Needs: domain terminology

3 stubs found. Run '/discover' in Claude Code to generate content,
or edit the files directly.
```

**Output (JSON)**: Array of `{ "path": "...", "zone": "...", "hint": "..." }`.

**Exit Codes**:
- `0` -- No stubs found (everything has content)
- `1` -- Stubs found
- `3` -- Not a Mind project

**Edge Cases**:
- No documents at all: Message suggesting `mind init`.
- All documents are stubs: List all with a note about `mind create brief` as the starting point.

**Examples**:
```bash
mind docs stubs              # List stubs
mind docs stubs --json       # JSON for scripting
```

---

### `mind docs search`

**Synopsis**: `mind docs search "<query>" [flags]`

**Description**: Full-text search across all documents in `docs/`. Shows matching lines with context.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `query` | Yes | Search string (case-insensitive substring match) |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Walk all `.md` files in `docs/`
3. Search for the query string (case-insensitive) in file contents
4. Collect matching lines with 1 line of context above and below
5. Group results by file
6. Sort by relevance (number of matches per file, then alphabetically)

**Output (Interactive)**:
```
Found 7 matches in 3 files:

docs/spec/requirements.md (4 matches)
  12: FR-3: The system SHALL authenticate users via JWT tokens
  45: FR-8: The system SHALL validate JWT expiration on every request
  ...

docs/spec/architecture.md (2 matches)
  34: The authentication layer uses JWT with RS256 signing
  78: JWT tokens are validated in middleware before route handlers

docs/knowledge/auth-convergence.md (1 match)
  15: JWT was selected over session cookies for stateless API design
```

**Output (JSON)**: Array of `{ "file": "...", "matches": [{ "line": N, "text": "...", "context_before": "...", "context_after": "..." }] }`.

**Exit Codes**:
- `0` -- Matches found
- `1` -- No matches found
- `3` -- Not a Mind project

**Edge Cases**:
- Empty query string: Error message.
- No `.md` files in `docs/`: Message suggesting `mind init`.
- Very large number of matches: Truncate output at 50 matches with a count of remaining.

**Examples**:
```bash
mind docs search "authentication"
mind docs search "FR-" --json        # Find all functional requirements
mind docs search "TODO"
```

---

### `mind check docs`

**Synopsis**: `mind check docs [--strict] [flags]`

**Description**: Run the 17-check documentation validation suite. Verifies zone structure, required files, content completeness, and naming conventions.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--strict` | `-s` | bool | `false` | Treat warnings as failures (stubs become errors) |

**Behavior**:
1. Detect project root
2. Run 17 checks:
   - [1-5] Zone directories exist (spec, blueprints, state, iterations, knowledge)
   - [6] `mind.toml` exists and is valid TOML
   - [7] `project-brief.md` exists
   - [8] `requirements.md` exists
   - [9] `architecture.md` exists
   - [10] `domain-model.md` exists
   - [11] `blueprints/INDEX.md` exists
   - [12] At least one iteration exists (warn, not fail)
   - [13] `current.md` exists in state/
   - [14] `workflow.md` exists in state/
   - [15] `glossary.md` exists in knowledge/ (warn)
   - [16] All spec files have non-stub content (warn, or fail if `--strict`)
   - [17] Naming conventions: files are kebab-case, iterations match `NNN-TYPE-descriptor` pattern
3. Build a `ValidationReport` with pass/fail/warn per check
4. Render report

**Output (Interactive)**:
```
Documentation Validation (17 checks)

  ✓ [1]  spec/ directory exists
  ✓ [2]  blueprints/ directory exists
  ✓ [3]  state/ directory exists
  ✓ [4]  iterations/ directory exists
  ✓ [5]  knowledge/ directory exists
  ✓ [6]  mind.toml valid
  ✓ [7]  project-brief.md exists
  ✓ [8]  requirements.md exists
  ✓ [9]  architecture.md exists
  ✓ [10] domain-model.md exists
  ✓ [11] blueprints/INDEX.md exists
  ⚠ [12] No iterations found
  ✓ [13] state/current.md exists
  ✓ [14] state/workflow.md exists
  ⚠ [15] knowledge/glossary.md missing
  ⚠ [16] 2 stub documents found (domain-model.md, api-contracts.md)
  ✓ [17] Naming conventions pass

Result: 14 pass, 0 fail, 3 warnings
```

**Output (JSON)**: `ValidationReport` object. See BP-03, `domain/validation.go` for the schema.

**Exit Codes**:
- `0` -- All checks pass (warnings are okay)
- `1` -- Failures found (or warnings with `--strict`)
- `3` -- Not a Mind project

**Edge Cases**:
- `mind.toml` is malformed TOML: Check [6] fails with parse error message.
- `--strict` with stubs: Stubs become failures rather than warnings.

**Examples**:
```bash
mind check docs              # Standard validation
mind check docs --strict     # Stubs are errors
mind check docs --json       # JSON report
```

---

### `mind check refs`

**Synopsis**: `mind check refs [flags]`

**Description**: Run the 11-check cross-reference validation suite. Verifies that internal document links, iteration references, and `mind.toml` registry entries are consistent.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Run 11 checks:
   - [1] All paths in `mind.toml` `[documents]` point to existing files
   - [2] No orphan documents (files in `docs/` not registered in `mind.toml`)
   - [3] `blueprints/INDEX.md` entries match actual blueprint files
   - [4] No broken internal markdown links (`[text](path)` where path does not exist)
   - [5] All iteration directories contain an `overview.md`
   - [6] ADR sequence numbers are contiguous (warn if gaps)
   - [7] Blueprint sequence numbers are contiguous (warn if gaps)
   - [8] Iteration sequence numbers are contiguous (warn if gaps)
   - [9] `current.md` "Recent Changes" references valid iteration IDs
   - [10] `workflow.md` references a valid iteration path (if non-idle)
   - [11] `.claude/CLAUDE.md` references valid paths
3. Build `ValidationReport` and render

**Output (Interactive)**:
```
Cross-Reference Validation (11 checks)

  ✓ [1]  mind.toml paths resolve
  ✓ [2]  No orphan documents
  ✓ [3]  INDEX.md matches blueprints
  ✗ [4]  Broken link in architecture.md:45 → "decisions/003-caching.md" (not found)
  ✓ [5]  All iterations have overview.md
  ✓ [6]  ADR sequences contiguous
  ✓ [7]  Blueprint sequences contiguous
  ✓ [8]  Iteration sequences contiguous
  ✓ [9]  current.md references valid
  ✓ [10] workflow.md references valid
  ✓ [11] .claude/CLAUDE.md references valid

Result: 10 pass, 1 fail, 0 warnings
```

**Output (JSON)**: `ValidationReport` object.

**Exit Codes**:
- `0` -- All checks pass
- `1` -- Failures found
- `3` -- Not a Mind project

**Edge Cases**:
- No `mind.toml`: Check [1] and [2] fail, remaining checks still run.
- Broken link in a stub document: Reported but won't block unless `--strict` is used at the `check all` level.

**Examples**:
```bash
mind check refs              # Cross-reference validation
mind check refs --json       # JSON report
```

---

### `mind check config`

**Synopsis**: `mind check config [flags]`

**Description**: Validate conversation YAML configuration files in `.mind/conversation/`. Checks for valid YAML syntax, required fields, and schema conformance.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Scan `.mind/conversation/` for YAML files (`.yml`, `.yaml`)
3. For each file:
   - Parse YAML syntax (report parse errors with line numbers)
   - Validate required fields per schema
   - Check that referenced agent files exist
   - Check that referenced protocol files exist
4. Build `ValidationReport` and render

**Output (Interactive)**:
```
Conversation Config Validation

  ✓ .mind/conversation/analysis/config.yml       (valid, 3 personas)
  ✓ .mind/conversation/workflow/config.yml        (valid, 5 agents)
  ✓ .mind/conversation/protocols/state.yml        (valid)
  ✗ .mind/conversation/protocols/routing.yml      (missing required field: phases)

Result: 3 pass, 1 fail
```

**Output (JSON)**: `ValidationReport` object.

**Exit Codes**:
- `0` -- All configs valid
- `1` -- Validation failures
- `3` -- Not a Mind project

**Edge Cases**:
- No YAML files found: Pass with message "No conversation configs found."
- Binary file in directory: Skip gracefully.
- Deeply nested YAML: Parse up to 10 levels deep.

**Examples**:
```bash
mind check config            # Validate conversation configs
mind check config --json     # JSON report
```

---

### `mind check convergence`

**Synopsis**: `mind check convergence <file> [flags]`

**Description**: Run the 23-check convergence output validation suite on a specific convergence analysis file. Verifies persona sections, synthesis quality, quality rubric, and structural completeness.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Path to the convergence file (relative or absolute) |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Resolve the file path (relative to project root or absolute)
3. Run 23 checks across 4 categories:
   - **Structure** [1-6]: Required sections present (Persona headers, Synthesis, Recommendations, Quality Rubric)
   - **Persona quality** [7-14]: Each persona has substantive content (not stub), unique perspective, evidence-based arguments
   - **Synthesis quality** [15-19]: Synthesis references all personas, identifies agreements/disagreements, has actionable recommendations
   - **Quality rubric** [20-23]: Scores present in all 6 dimensions, overall score calculated, Gate 0 threshold evaluated
4. Build `ValidationReport` and render

**Output (Interactive)**:
```
Convergence Validation: docs/knowledge/auth-convergence.md (23 checks)

  Structure (6 checks)         ✓ 6 pass
  Persona Quality (8 checks)   ✓ 7 pass, 1 warn
  Synthesis Quality (5 checks) ✓ 5 pass
  Quality Rubric (4 checks)    ✓ 4 pass

  ⚠ [11] Persona "security-engineer" has thin analysis (< 200 words)

  Overall: 3.8/5.0 — Gate 0: PASS (threshold: 3.0)

Result: 22 pass, 0 fail, 1 warning
```

**Output (JSON)**: `ValidationReport` object with additional `quality_score` field.

**Exit Codes**:
- `0` -- All checks pass, Gate 0 passes
- `1` -- Failures found or Gate 0 fails
- `3` -- Not a Mind project

**Edge Cases**:
- File not found: Error with message.
- File is not a convergence document (no quality rubric section): Structural checks fail.
- Quality rubric has unparseable scores: Check fails with parse error detail.

**Examples**:
```bash
mind check convergence docs/knowledge/auth-convergence.md
mind check convergence docs/knowledge/db-convergence.md --json
```

---

### `mind check all`

**Synopsis**: `mind check all [--strict] [flags]`

**Description**: Run all validation suites (docs, refs, config) and produce a unified report. Does not include convergence validation (that requires a specific file argument).

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--strict` | `-s` | bool | `false` | Treat warnings as failures across all suites |

**Behavior**:
1. Detect project root
2. Run `check docs` (17 checks)
3. Run `check refs` (11 checks)
4. Run `check config` (variable checks based on files found)
5. Aggregate all results into a unified report
6. Render with per-suite sections and overall summary

**Output (Interactive)**:
```
╭─ Documentation (17 checks) ──────────────────────╮
│  Pass: 15  Fail: 1  Warn: 1                      │
│  ✗ [16] domain-model.md is a stub (STRICT)        │
│  ⚠ [12] No iterations found                       │
╰───────────────────────────────────────────────────╯

╭─ Cross-References (11 checks) ───────────────────╮
│  Pass: 11  Fail: 0  Warn: 0                      │
╰───────────────────────────────────────────────────╯

╭─ Conversation Config (4 files) ──────────────────╮
│  Pass: 4   Fail: 0  Warn: 0                      │
╰───────────────────────────────────────────────────╯

Overall: 30/32 pass, 1 fail, 1 warning
Exit code: 1 (failures present)
```

**Output (JSON)**: Object with `suites` (array of `ValidationReport` per suite) and `overall` summary.

**Exit Codes**:
- `0` -- All suites pass
- `1` -- Any suite has failures (or warnings with `--strict`)
- `3` -- Not a Mind project

**Edge Cases**:
- One suite fails to run (e.g., missing `.mind/conversation/`): That suite is skipped with a warning, others still run.
- `--strict`: Warnings in any suite become failures in the overall report.

**Examples**:
```bash
mind check all               # Run everything
mind check all --strict      # Warnings are errors
mind check all --json        # Unified JSON report
```

---

### `mind workflow status`

**Synopsis**: `mind workflow status [flags]`

**Description**: Show the current workflow state. Reads `docs/state/workflow.md` and displays whether a workflow is active, which agent ran last, and what remains.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Read and parse `docs/state/workflow.md`
3. If idle (no active workflow), display idle state with last completed iteration
4. If active, display: type, iteration path, branch, last agent, remaining chain, session info, completed artifacts

**Output (Interactive)**:
```
Workflow: idle (no active workflow)
Last completed: 006-ENHANCEMENT-add-caching (2026-03-08)
```

Or, if a workflow is active:
```
Workflow: active
  Type: NEW_PROJECT
  Iteration: docs/iterations/007-NEW_PROJECT-rest-api/
  Branch: new/rest-api
  Last Agent: architect (completed)
  Remaining: developer → tester → reviewer
  Session: 1 of 2

  Completed Artifacts:
    ✓ requirements.md (analyst)
    ✓ architecture.md (architect)
```

**Output (JSON)**: `WorkflowState` object. See BP-03, `domain/workflow.go` for the schema.

**Exit Codes**:
- `0` -- Success (idle or active)
- `3` -- Not a Mind project

**Edge Cases**:
- `workflow.md` does not exist: Report as idle.
- `workflow.md` is malformed: Warning about parse error, report as idle.

**Examples**:
```bash
mind workflow status         # Show current state
mind workflow status --json  # JSON for scripting
```

---

### `mind workflow history`

**Synopsis**: `mind workflow history [flags]`

**Description**: List all past iterations chronologically. Shows type, name, status, date, and artifact count.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Scan `docs/iterations/` for iteration directories
3. Parse each directory name for sequence, type, and descriptor
4. Check artifact completeness (5 expected files)
5. Read `overview.md` for creation date
6. Sort by sequence number descending (most recent first)
7. Render table

**Output (Interactive)**:
```
  #   Type          Name                    Status      Date         Files
  ─── ───────────── ─────────────────────── ─────────── ──────────── ─────
  006 ENHANCEMENT   add-caching             ✓ complete  2026-03-08   5/5
  005 BUG_FIX       fix-auth-redirect       ✓ complete  2026-03-07   5/5
  004 REFACTOR      extract-repositories    ✓ complete  2026-03-06   4/5
  003 ENHANCEMENT   websocket-notifications ✓ complete  2026-03-05   5/5
  002 ENHANCEMENT   role-based-access       ✓ complete  2026-03-03   5/5
  001 NEW_PROJECT   initial-api             ✓ complete  2026-03-01   5/5

6 iterations (5 complete, 1 incomplete)
```

**Output (JSON)**: Array of `Iteration` objects.

**Exit Codes**:
- `0` -- Success
- `3` -- Not a Mind project

**Edge Cases**:
- No iterations: Print "No iterations found."
- Malformed directory name: Warn and skip.

**Examples**:
```bash
mind workflow history         # Chronological list
mind workflow history --json  # JSON output
```

---

### `mind workflow show`

**Synopsis**: `mind workflow show <iteration-id> [flags]`

**Description**: Show detailed information about a specific iteration, including its overview, artifacts, and validation status.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `iteration-id` | Yes | Sequence number (e.g., `007`) or full directory name |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Resolve the iteration: if numeric, look up `docs/iterations/NNN-*`; if full name, resolve directly
3. Parse the iteration directory for artifacts
4. Read `overview.md` for metadata
5. Check artifact completeness
6. If `validation.md` exists, parse reviewer findings
7. Display iteration details

**Output (Interactive)**:
```
Iteration: 007-NEW_PROJECT-rest-api

  Type: NEW_PROJECT
  Descriptor: rest-api
  Created: 2026-03-09
  Status: complete

  Artifacts:
    ✓ overview.md          1.2 KB
    ✓ changes.md           3.8 KB
    ✓ test-summary.md      2.1 KB
    ✓ validation.md        1.8 KB
    ✓ retrospective.md     0.9 KB

  Reviewer Findings (from validation.md):
    MUST:   0
    SHOULD: 2
    COULD:  1
    Sign-off: APPROVED
```

**Output (JSON)**: `Iteration` object with additional `reviewer_findings` field.

**Exit Codes**:
- `0` -- Success
- `1` -- Iteration not found
- `3` -- Not a Mind project

**Edge Cases**:
- Iteration ID not found: List available IDs in error message.
- Partial match (multiple iterations match `007`): Should not happen if directory naming is correct; if it does, show first match.
- `validation.md` missing or unparseable: Show artifact as present but skip reviewer findings section.

**Examples**:
```bash
mind workflow show 007                      # By sequence number
mind workflow show 007-NEW_PROJECT-rest-api # By full name
mind workflow show 7                        # Leading zeros optional
mind workflow show 007 --json               # JSON details
```

---

### `mind workflow clean`

**Synopsis**: `mind workflow clean [--dry-run] [flags]`

**Description**: Remove stale workflow state. Resets `docs/state/workflow.md` to idle if no active workflow should exist (e.g., the referenced iteration is complete or the branch no longer exists).

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--dry-run` | | bool | `false` | Show what would be cleaned without making changes |

**Behavior**:
1. Detect project root
2. Parse `docs/state/workflow.md`
3. If already idle, report "Nothing to clean"
4. If active, check for staleness:
   - Referenced iteration directory exists?
   - Referenced branch exists in git?
   - Last agent update older than 24 hours?
5. If stale (or forced with user confirmation), reset `workflow.md` to idle state
6. Report changes

**Output (Interactive)**:
```
Stale workflow detected:
  Iteration: 007-NEW_PROJECT-rest-api
  Branch: new/rest-api (exists)
  Last update: 2026-03-08 (3 days ago)

Cleaned: workflow.md reset to idle
```

**Output (JSON)**: `{ "was_stale": true, "iteration": "...", "action": "cleaned" }` or `{ "was_stale": false, "action": "none" }`.

**Exit Codes**:
- `0` -- Cleaned or nothing to clean
- `1` -- Error during cleanup
- `3` -- Not a Mind project

**Edge Cases**:
- `workflow.md` does not exist: Nothing to clean, exit 0.
- `--dry-run`: Report what would happen without modifying files.
- Active workflow is not stale: Warn and ask for confirmation (skip in non-interactive mode).

**Examples**:
```bash
mind workflow clean              # Clean stale state
mind workflow clean --dry-run    # Preview cleanup
```

---

### `mind sync agents`

**Synopsis**: `mind sync agents [--check] [flags]`

**Description**: Synchronize agent definitions from `.mind/conversation/agents/` (or `.mind/agents/`) to `.github/agents/`. Ensures Copilot Chat agents match the canonical definitions.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--check` | | bool | `false` | Only check for differences, do not sync (exit 1 if out of sync) |

**Behavior**:
1. Detect project root
2. Scan source agent files in `.mind/`
3. Scan target agent files in `.github/agents/`
4. Diff each pair (content comparison)
5. If `--check`: report differences and exit
6. If not `--check`: copy updated files from source to target
7. Report sync results

**Output (Interactive)**:
```
Syncing agents: .mind/agents/ → .github/agents/

  ✓ analyst.md        (up to date)
  ↻ architect.md      (updated — 3 lines changed)
  ✓ developer.md      (up to date)
  + tester.md         (new — copied)
  ✓ reviewer.md       (up to date)

Synced: 1 updated, 1 new, 3 unchanged
```

**Output (JSON)**: Object with `synced` array of `{ "file": "...", "action": "updated|new|unchanged", "diff_lines": N }`.

**Exit Codes**:
- `0` -- Sync complete (or `--check` and everything is in sync)
- `1` -- `--check` and differences found, or sync error
- `3` -- Not a Mind project

**Edge Cases**:
- `.github/agents/` does not exist: Create it during sync.
- No source agents found: Error message.
- `--check` mode in CI: Exit 1 when out of sync, useful as a CI gate.
- Target has files not in source (extra agents): Warn but do not delete.

**Examples**:
```bash
mind sync agents             # Sync agent definitions
mind sync agents --check     # Check only (CI mode)
mind sync agents --json      # JSON sync report
```

---

### `mind quality log`

**Synopsis**: `mind quality log <convergence-file> [flags]`

**Description**: Extract quality scores from a convergence analysis file and append an entry to `docs/knowledge/quality-log.yml`.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `convergence-file` | Yes | Path to the convergence document |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--topic` | `-t` | string | auto-detect | Override the topic name (defaults to filename-derived topic) |
| `--variant` | | string | `""` | Label this as a variant analysis (e.g., "v2", "revised") |

**Behavior**:
1. Detect project root
2. Read and parse the convergence file
3. Locate the Quality Rubric section
4. Extract scores for each of the 6 dimensions
5. Calculate the overall score (average of dimensions)
6. Determine Gate 0 result (overall >= 3.0 passes)
7. Build a `QualityEntry` with date, topic, scores, gate result, personas used, variant
8. Append to `docs/knowledge/quality-log.yml` (create if it does not exist)
9. Print confirmation with score summary

**Output (Interactive)**:
```
Logged quality scores from: auth-convergence.md

  Topic: Authentication Strategy
  Dimensions:
    Breadth of Perspectives:  4/5
    Depth of Analysis:        3/5
    Practical Applicability:  4/5
    Intellectual Rigor:       4/5
    Synthesis Quality:        3/5
    Actionable Outcomes:      4/5

  Overall: 3.7/5.0 — Gate 0: PASS
  Variant: (none)

✓ Appended to docs/knowledge/quality-log.yml
```

**Output (JSON)**: The `QualityEntry` object that was appended.

**Exit Codes**:
- `0` -- Successfully logged
- `1` -- Parse error (no quality rubric found, unparseable scores)
- `3` -- Not a Mind project

**Edge Cases**:
- Quality rubric section missing: Error with message pointing to the expected section header.
- Scores not in expected format: Error listing what was found vs. what was expected.
- `quality-log.yml` does not exist: Create it with a header comment and the first entry.
- Duplicate entry (same file already logged): Warn and skip unless `--variant` provides a different label.

**Examples**:
```bash
mind quality log docs/knowledge/auth-convergence.md
mind quality log docs/knowledge/db-convergence.md --topic "Database Selection"
mind quality log docs/knowledge/auth-convergence.md --variant "v2-revised"
```

---

### `mind quality history`

**Synopsis**: `mind quality history [flags]`

**Description**: Show quality score trends over time from `docs/knowledge/quality-log.yml`.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Read and parse `docs/knowledge/quality-log.yml`
3. Sort entries chronologically
4. Display as a table with date, topic, overall score, gate result
5. In interactive mode, render an ASCII trend line showing score progression

**Output (Interactive)**:
```
Quality Score History

  Date        Topic                    Overall  Gate 0
  ──────────  ───────────────────────  ───────  ──────
  2026-03-01  Authentication Strategy  2.3/5.0  FAIL
  2026-03-03  Authentication Strategy  3.2/5.0  PASS
  2026-03-05  Database Selection       3.5/5.0  PASS
  2026-03-07  Authentication Strategy  3.8/5.0  PASS

  Trend: 2.3 → 3.2 → 3.5 → 3.8 (improving ↑)
```

**Output (JSON)**: Array of `QualityEntry` objects.

**Exit Codes**:
- `0` -- Success
- `1` -- `quality-log.yml` not found or empty
- `3` -- Not a Mind project

**Edge Cases**:
- `quality-log.yml` does not exist: Message suggesting `mind quality log <file>`.
- Only one entry: Show it without trend line.
- Entries for multiple topics: Group by topic in table.

**Examples**:
```bash
mind quality history          # Show trends
mind quality history --json   # JSON array
```

---

### `mind quality report`

**Synopsis**: `mind quality report [flags]`

**Description**: Generate a summary report of all convergence analyses. Aggregates scores by topic, shows best/worst dimensions, and identifies topics that have not yet passed Gate 0.

**Arguments**: None

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Read `docs/knowledge/quality-log.yml`
3. Group entries by topic
4. For each topic, find the latest score
5. Calculate aggregate statistics: average overall, worst dimension across topics, best dimension
6. Identify topics with Gate 0 failures
7. Render summary report

**Output (Interactive)**:
```
Quality Report

  Topics Analyzed: 3
  Average Overall Score: 3.5/5.0
  Gate 0 Pass Rate: 2/3 (67%)

  Best Dimension:    Breadth of Perspectives (avg 4.2)
  Weakest Dimension: Synthesis Quality (avg 2.8)

  Topics Needing Attention:
    ⚠ Caching Strategy — latest score 2.5/5.0 (Gate 0: FAIL)

  Topics Passing:
    ✓ Authentication Strategy — 3.8/5.0
    ✓ Database Selection — 3.5/5.0
```

**Output (JSON)**: Object with `topics_count`, `average_overall`, `gate_pass_rate`, `best_dimension`, `weakest_dimension`, `topics` array.

**Exit Codes**:
- `0` -- Success
- `1` -- `quality-log.yml` not found or empty
- `3` -- Not a Mind project

**Edge Cases**:
- No quality data: Message suggesting convergence analysis workflow.
- Single entry: Report still renders with that entry.

**Examples**:
```bash
mind quality report          # Summary report
mind quality report --json   # JSON report
```

---

### `mind reconcile`

**Synopsis**: `mind reconcile [--force] [flags]`

**Description**: Run the reconciliation engine that detects staleness between documentation, code, and workflow state. Checks if documents have drifted from their `mind.toml` registry, if code has changed without a matching iteration, and if the current state file is up to date.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | `-f` | bool | `false` | Force reconciliation even if `mind.lock` is fresh |

**Behavior**:
1. Detect project root
2. Read `mind.lock` for last reconciliation timestamp (if exists)
3. If lock is fresh (< 1 hour) and `--force` is not set, skip with "Already reconciled recently"
4. Compare `mind.toml` document registry against actual files on disk:
   - Files in registry but missing from disk
   - Files on disk but missing from registry
5. Check `docs/state/current.md` "Recent Changes" against actual latest iteration
6. Check git status for uncommitted documentation changes
7. Write results to `mind.lock` with timestamp
8. Report findings with suggested actions

**Output (Interactive)**:
```
Reconciliation Report

  Registry vs. Disk:
    ✓ 12 documents in sync
    ✗ 1 missing: docs/spec/api-contracts.md (in registry, not on disk)
    ⚠ 1 unregistered: docs/knowledge/caching-spike.md

  State Freshness:
    ✓ current.md references latest iteration (006)

  Working Tree:
    ⚠ 2 uncommitted doc changes: requirements.md, architecture.md

  Updated mind.lock (next check in 1 hour)
```

**Output (JSON)**: Object with `registry_sync` (`{in_sync, missing, unregistered}`), `state_freshness` (bool), `uncommitted_docs` (array), `lock_updated` (timestamp).

**Exit Codes**:
- `0` -- Everything in sync
- `1` -- Drift detected
- `3` -- Not a Mind project

**Edge Cases**:
- No `mind.lock`: First reconciliation, always runs.
- `--force`: Ignores lock freshness.
- No git repository: Skip working tree check.

**Examples**:
```bash
mind reconcile               # Run reconciliation
mind reconcile --force       # Force even if recently reconciled
mind reconcile --json        # JSON report
```

---

### `mind tui`

**Synopsis**: `mind tui [flags]`

**Description**: Launch the full-screen interactive TUI dashboard. Provides 5 tabs: Status, Documents, Iterations, Checks, Quality. Requires an interactive terminal.

**Arguments**: None

**Flags**: Global flags only (except `--json` has no effect).

**Behavior**:
1. Detect project root
2. Check that stdout is a TTY (abort if piped)
3. Initialize Bubble Tea application with 5 tab models
4. Load project health data asynchronously
5. Render the TUI with tab navigation (number keys 1-5)
6. Handle keyboard input: `q`/`Ctrl+C` to quit, tab-specific keys for navigation
7. On quit, restore terminal state cleanly

**Output**: Full-screen TUI application (see BP-01 for detailed mockups of each tab).

**Exit Codes**:
- `0` -- Normal exit (user quit)
- `1` -- Error during startup
- `3` -- Not a Mind project

**Edge Cases**:
- Not a TTY (piped): Error message: "TUI requires an interactive terminal. Use 'mind status' for non-interactive output."
- Terminal too small (< 80x24): Render with truncated layout and a size warning.
- Project data fails to load: Show error in the TUI rather than crashing.

**Examples**:
```bash
mind tui                     # Launch full-screen dashboard
```

---

### `mind preflight`

**Synopsis**: `mind preflight "<request>" [flags]` or `mind preflight --resume [flags]`

**Description**: Prepare everything for an AI workflow before handing off to Claude Code. Classifies the request, runs the business context gate, creates the iteration folder, creates the git branch, and assembles a context package. This is Model A of the AI Workflow Bridge (see BP-02).

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `request` | Yes (unless `--resume`) | The workflow request (e.g., "create: REST API with JWT auth") |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resume` | | bool | `false` | Check for a resumable interrupted workflow instead of starting a new one |

**Behavior (new workflow)**:
1. Detect project root
2. Classify the request by keywords into a `RequestType` (NEW_PROJECT, BUG_FIX, ENHANCEMENT, REFACTOR)
3. Run business context gate: check `project-brief.md` for Vision, Key Deliverables, Scope
4. If gate blocks (e.g., NEW_PROJECT with missing brief): abort with remediation advice
5. Run documentation validation (17 checks) -- report warnings but don't block
6. Create iteration folder with 5 template files (via `create iteration` logic)
7. Populate `overview.md` with classification, request, scope, and chain
8. Create git branch: `{type}/{descriptor}` (e.g., `new/rest-api`)
9. Assemble context package: read relevant spec docs, recent iterations, convergence docs
10. Write workflow state to `docs/state/workflow.md`
11. Generate a prompt summary (copy to clipboard if `xclip`/`pbcopy` available)
12. Print pre-flight summary with next steps

**Behavior (--resume)**:
1. Detect project root
2. Read `docs/state/workflow.md`
3. If idle: "No interrupted workflow found"
4. If active: display resumption details (last agent, remaining chain, completed artifacts)

**Output (Interactive, new workflow)**:
```
╭─ Pre-Flight Complete ─────────────────────────────────╮
│                                                        │
│  Type: NEW_PROJECT                                     │
│  Chain: analyst → architect → developer → tester       │
│         → reviewer                                     │
│  Branch: new/rest-api                                  │
│  Iteration: docs/iterations/007-NEW_PROJECT-rest-api/  │
│  Brief: ✓ present (Vision, Deliverables, Scope)        │
│  Docs: 15/17 pass (2 warnings, 0 blockers)             │
│                                                        │
│  Context package ready. Open Claude Code and run:       │
│  /workflow "create: REST API with JWT auth"             │
│                                                        │
│  Or paste the generated prompt (copied to clipboard).   │
╰────────────────────────────────────────────────────────╯
```

**Output (Interactive, --resume)**:
```
╭─ Resumable Workflow Found ─────────────────────────────╮
│                                                          │
│  Type: NEW_PROJECT                                       │
│  Iteration: 007-NEW_PROJECT-rest-api                     │
│  Last Agent: architect (completed)                       │
│  Remaining: developer → tester → reviewer                │
│  Branch: new/rest-api                                    │
│  Session: 1 of 2 (split after architect)                 │
│                                                          │
│  Completed Artifacts:                                    │
│    ✓ requirements.md (analyst)                           │
│    ✓ architecture.md (architect)                         │
│                                                          │
│  Resume in Claude Code:                                  │
│  /workflow                                               │
╰──────────────────────────────────────────────────────────╯
```

**Output (JSON)**: Object with `type`, `chain`, `branch`, `iteration_path`, `brief_gate`, `docs_validation`, `context_files` (array), `clipboard` (bool).

**Exit Codes**:
- `0` -- Pre-flight complete
- `1` -- Gate blocked (brief missing for NEW_PROJECT), or no resumable workflow (with `--resume`)
- `3` -- Not a Mind project

**Edge Cases**:
- Request classification is ambiguous: Use the most common type (ENHANCEMENT) and note the ambiguity in the output.
- Brief is a stub for NEW_PROJECT: Block with message to run `mind create brief` or `/discover`.
- Git branch already exists: Append a numeric suffix (e.g., `new/rest-api-2`).
- Workflow already active: Error suggesting `mind preflight --resume` or `mind workflow clean`.
- No git: Skip branch creation, warn.

**Examples**:
```bash
mind preflight "create: REST API with JWT auth"
mind preflight "fix: 500 error on /api/users"
mind preflight "enhance: add caching to user queries"
mind preflight --resume
mind preflight "create: new service" --json
```

---

### `mind handoff`

**Synopsis**: `mind handoff <iteration-id> [flags]`

**Description**: Post-workflow cleanup after an AI workflow completes. Validates iteration completeness, runs deterministic checks (build/lint/test), updates project state, and clears workflow state. This is the companion to `mind preflight` in Model A.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `iteration-id` | Yes | Sequence number or full directory name of the completed iteration |

**Flags**: Global flags only.

**Behavior**:
1. Detect project root
2. Resolve the iteration (same logic as `workflow show`)
3. Validate iteration artifact completeness (all 5 files present with content)
4. Run deterministic checks using commands from `mind.toml` `[project.commands]`:
   - `build` command (e.g., `go build -o mind .`)
   - `lint` command (e.g., `golangci-lint run ./...`)
   - `test` command (e.g., `go test ./...`)
5. Update `docs/state/current.md`:
   - "Active Work" -> None
   - "Recent Changes" -> Add the completed iteration
   - "Next Priorities" -> Prompt the user (interactive) or leave as-is (non-interactive)
6. Clear `docs/state/workflow.md` to idle
7. Report summary with PR suggestion

**Output (Interactive)**:
```
Handoff: 007-NEW_PROJECT-rest-api

  1. Iteration Completeness
     ✓ overview.md ✓ changes.md ✓ test-summary.md
     ✓ validation.md ✓ retrospective.md

  2. Deterministic Checks
     ✓ go build    (2.1s)
     ✓ go test     (4.8s, 142 passed)
     ✓ golangci-lint (1.3s)

  3. State Updated
     ✓ docs/state/current.md updated
     ✓ docs/state/workflow.md → idle

  Branch: new/rest-api (5 commits ahead of main)
  Suggestion: Create a pull request — gh pr create
```

**Output (JSON)**: Object with `iteration`, `artifacts` (completeness), `checks` (build/lint/test results), `state_updated` (bool), `branch_info`.

**Exit Codes**:
- `0` -- Handoff complete, all checks pass
- `1` -- Handoff complete but deterministic checks failed (state still updated)
- `3` -- Not a Mind project

**Edge Cases**:
- Iteration not found: Error with list of available IDs.
- Missing artifacts: Warn but continue handoff (don't block state update).
- Build/lint/test commands not configured in `mind.toml`: Skip deterministic checks with a warning.
- No git: Skip branch info.

**Examples**:
```bash
mind handoff 007
mind handoff 007-NEW_PROJECT-rest-api
mind handoff 7 --json
```

---

### `mind serve`

**Synopsis**: `mind serve [flags]`

**Description**: Start the MCP (Model Context Protocol) server on stdio. Claude Code and other MCP clients connect to this server to access project intelligence as callable tools. This is Model B of the AI Workflow Bridge (see BP-02).

**Arguments**: None

**Flags**: Global flags only (except `--json` and `--no-color` have no effect).

**Behavior**:
1. Detect project root
2. Initialize all service layer components
3. Register MCP tools (16 tools -- see BP-02 for the full list):
   - `mind_status`, `mind_doctor`, `mind_check_brief`
   - `mind_validate_docs`, `mind_validate_refs`
   - `mind_list_iterations`, `mind_show_iteration`
   - `mind_read_state`, `mind_update_state`
   - `mind_create_iteration`, `mind_list_stubs`
   - `mind_check_gate`, `mind_log_quality`
   - `mind_search_docs`, `mind_read_config`
   - `mind_suggest_next`
4. Start JSON-RPC over stdio (MCP protocol)
5. Handle tool calls by delegating to the same `internal/` service packages used by CLI commands
6. Run until the client disconnects or the process is killed

**Output**: JSON-RPC messages on stdout (MCP protocol). No human-readable output -- this command is designed for machine consumption.

**Configuration**: Clients configure the server in `.mcp.json`:
```json
{
  "mcpServers": {
    "mind": {
      "command": "mind",
      "args": ["serve"],
      "description": "Mind Framework project intelligence"
    }
  }
}
```

**Exit Codes**:
- `0` -- Normal shutdown (client disconnected)
- `1` -- Startup error (project not found, initialization failure)
- `3` -- Not a Mind project

**Edge Cases**:
- Run in an interactive terminal: Works, but produces JSON-RPC output that is not human-readable. No special handling.
- Client sends invalid JSON-RPC: Respond with standard JSON-RPC error.
- Project state changes while server is running: Tools re-read from disk on each call (no caching of stale data).

**Examples**:
```bash
mind serve                   # Start MCP server (typically not run directly)
# Configured in .mcp.json for Claude Code:
# { "mcpServers": { "mind": { "command": "mind", "args": ["serve"] } } }
```

---

### `mind watch`

**Synopsis**: `mind watch [--tui] [flags]`

**Description**: Start a filesystem watcher that monitors the project for changes and provides real-time feedback. Automatically runs validation, build checks, and gate monitoring as files change. This is Model C of the AI Workflow Bridge (see BP-02).

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--tui` | | bool | `false` | Launch a live TUI dashboard instead of log-style output |

**Behavior**:
1. Detect project root
2. Initialize filesystem watcher (fsnotify) on `docs/`, `src/` (or language-appropriate source dirs), `.mind/`
3. Start event loop:
   - `docs/state/workflow.md` changed -> Parse workflow state, update display
   - `docs/iterations/*/overview.md` created -> New iteration detected, show type + chain
   - `docs/iterations/*/changes.md` modified -> Run Micro-Gate B checks silently
   - `docs/spec/requirements.md` modified -> Run Micro-Gate A checks silently
   - `docs/spec/architecture.md` modified -> Show summary
   - `docs/knowledge/*-convergence.md` created -> Run 23-check convergence validation
   - Source files modified -> Run build/test in background, show results
   - `docs/spec/project-brief.md` changed -> Re-run business context gate
4. Debounce rapid changes (100ms window)
5. Run until Ctrl+C

**Output (log-style)**:
```
[14:23:01] Watching ~/dev/projects/my-app/
[14:23:15] docs/spec/requirements.md modified — Micro-Gate A: 6/6 pass
[14:23:32] src/routes/users.go created
[14:24:01] docs/iterations/007-NEW_PROJECT-rest-api/changes.md modified
[14:24:02] Micro-Gate B: ✓ changes.md exists, 4 files on disk
[14:24:45] Background: go build ✓ (2.1s)
[14:25:12] Background: go test ✓ 24 passed (4.8s)
```

**Output (--tui)**: Full-screen live dashboard (see BP-02 for mockup) showing workflow chain progress, live activity log, and pre-gate status.

**Exit Codes**:
- `0` -- Normal exit (Ctrl+C)
- `1` -- Startup error (watcher initialization failure)
- `3` -- Not a Mind project

**Edge Cases**:
- Source directory does not exist: Watch only `docs/` and `.mind/`.
- Build/test command not configured: Skip code checks, watch only docs.
- Very rapid file changes (AI writing many files): Debounce ensures checks run once per burst.
- File deleted: Log the deletion, do not crash.

**Examples**:
```bash
mind watch                   # Log-style watcher
mind watch --tui             # Full-screen live dashboard
```

---

### `mind run`

**Synopsis**: `mind run "<request>" [flags]`

**Description**: Full AI workflow orchestration. Performs pre-flight, dispatches each agent in the chain as a separate `claude` CLI invocation, runs quality gates between agents, and performs post-workflow cleanup. This is Model D of the AI Workflow Bridge (see BP-02). Requires `claude` CLI on `$PATH`.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `request` | Yes (unless `--resume`) | The workflow request (e.g., "create: REST API with JWT auth") |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resume` | | bool | `false` | Resume an interrupted orchestrated workflow |
| `--dry-run` | | bool | `false` | Simulate the workflow without making AI calls |
| `--tui` | | bool | `false` | Show a live TUI pipeline dashboard during execution |

**Behavior (standard run)**:
1. Check that `claude` CLI is on `$PATH` (abort if not)
2. Detect project root
3. Run pre-flight (same as `mind preflight`): classify, gate, create iteration, create branch
4. For each agent in the chain:
   a. Build the prompt: load agent instructions from `.mind/agents/{agent}.md`, inject project context (brief, requirements, architecture, iteration overview, prior agent outputs, conventions)
   b. Dispatch: pipe the prompt to `claude --print --model {model} --allowedTools Read,Write,Edit,Grep,Glob,Bash`
   c. Capture output
5. After each agent, run the appropriate quality gate:
   - After analyst: Micro-Gate A (requirements structure checks)
   - After developer: Micro-Gate B (changes.md, files on disk)
   - Before reviewer: Deterministic Gate (build/lint/test)
6. If a gate fails and retry count < 2: re-dispatch the agent with gate feedback
7. If a gate fails and retry count >= 2: proceed with documented concerns
8. Handle session splits (NEW_PROJECT auto-splits after architect): save state, prompt user
9. After all agents complete: run post-workflow cleanup (same as `mind handoff`)
10. Print final summary with cost estimate and PR suggestion

**Behavior (--resume)**:
1. Read `docs/state/workflow.md`
2. Determine the next agent in the chain
3. Resume dispatch loop from step 4

**Behavior (--dry-run)**:
1. Run pre-flight classification and gate checks
2. Print what would happen: agent chain, models, gate sequence, estimated cost
3. Do not create files, branches, or make AI calls

**Output (Interactive, standard)**:
```
Running: NEW_PROJECT — "create: REST API with JWT auth"

  ✓ Pre-flight     classify, gate, iteration, branch          0.8s
  ✓ Analyst         requirements extracted (12 FR, 8 AC)       2m 14s
  ✓ Micro-Gate A    6/6 checks pass                            0.2s
  ✓ Architect       architecture designed (4 components)       3m 01s
  ● Developer       implementing...                            1m 32s
```

**Output (Interactive, --dry-run)**:
```
Dry Run — no AI calls will be made

  Classification: NEW_PROJECT
  Chain: analyst → architect → developer → tester → reviewer
  Business Context Gate: PASS
  Iteration: would create 008-NEW_PROJECT-rest-api/
  Branch: would create new/rest-api

  Agent Dispatch Plan:
    1. Analyst  (opus)   — analyze request, produce requirements
    2. Architect (opus)  — design system architecture
    3. Developer (sonnet) — implement the solution
    4. Tester   (sonnet) — write tests
    5. Reviewer (opus)   — validate implementation

  Quality Gates:
    Micro-Gate A — after analyst
    Micro-Gate B — after developer
    Deterministic Gate — before reviewer
      Commands: go build -o mind ., go test ./..., golangci-lint run ./...

  Estimated cost: ~$1.50-3.00 (5 agent calls)
```

**Output (JSON)**: Object with `type`, `chain`, `dispatch_log` (array of `{agent, model, status, duration_ms}`), `gate_results`, `iteration`, `branch`, `cost_estimate`.

**Exit Codes**:
- `0` -- Workflow complete, all gates pass
- `1` -- Workflow complete with gate failures or errors
- `2` -- Workflow aborted by user
- `3` -- Not a Mind project

**Edge Cases**:
- `claude` not on `$PATH`: Error with install instructions.
- Agent dispatch fails (claude CLI error): Save state, report error, allow `--resume`.
- User presses Ctrl+C during agent dispatch: Save state for resume, clean up gracefully.
- `--resume` with no interrupted workflow: Error message.
- Session split: Prompt user to continue or pause; if non-interactive, auto-continue.

**Examples**:
```bash
mind run "create: REST API with JWT auth"
mind run "fix: 500 error on /api/users"
mind run --dry-run "enhance: add caching layer"
mind run --tui "create: user management module"
mind run --resume
```

---

### `mind completion`

**Synopsis**: `mind completion <shell> [flags]`

**Description**: Generate shell completion scripts. Outputs the script to stdout for sourcing into the shell configuration.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `shell` | Yes | Target shell: `bash`, `zsh`, `fish` |

**Flags**: Global flags only.

**Behavior**:
1. Validate the shell argument
2. Use Cobra's built-in completion generation for the specified shell
3. Output the completion script to stdout

**Output**: Shell completion script (source-able).

**Exit Codes**:
- `0` -- Script generated
- `1` -- Invalid shell argument

**Edge Cases**:
- Invalid shell name: Error listing valid options (bash, zsh, fish).
- Piped to file: Works as expected (stdout contains the script).

**Examples**:
```bash
mind completion bash > /etc/bash_completion.d/mind
mind completion zsh > "${fpath[1]}/_mind"
mind completion fish > ~/.config/fish/completions/mind.fish

# Or source directly:
source <(mind completion bash)
eval "$(mind completion zsh)"
```

---

### `mind version`

**Synopsis**: `mind version [--short] [flags]`

**Description**: Show version and build information.

**Arguments**: None

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--short` | | bool | `false` | Print only the version number (e.g., `0.1.0`) |

**Behavior**:
1. Read version from build-time variables (set by GoReleaser or `go build -ldflags`)
2. If `--short`: print only the semver string
3. Otherwise: print version, git commit, build date, Go version, OS/arch

**Output (Interactive)**:
```
mind version 0.1.0
  Commit:  a1b2c3d
  Date:    2026-03-10
  Go:      go1.23.0
  OS/Arch: linux/amd64
```

**Output (--short)**: `0.1.0`

**Output (JSON)**: `{ "version": "0.1.0", "commit": "a1b2c3d", "date": "2026-03-10", "go": "go1.23.0", "os": "linux", "arch": "amd64" }`

**Exit Codes**:
- `0` -- Always

**Edge Cases**:
- Development build (no ldflags set): Version shows `dev`, commit shows `unknown`.

**Examples**:
```bash
mind version              # Full version info
mind version --short      # Just the version number
mind version --json       # JSON for scripting
```

---

### `mind help`

**Synopsis**: `mind help [command] [flags]`

**Description**: Show help for any command. Provided automatically by Cobra.

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `command` | No | Command or subcommand to get help for |

**Flags**: None (Cobra handles this).

**Behavior**:
1. If no argument: show root help with all top-level commands
2. If argument provided: show help for that specific command, including synopsis, description, flags, and examples
3. Cobra auto-generates this from command definitions

**Output**: Standard Cobra help text.

**Exit Codes**:
- `0` -- Always

**Edge Cases**:
- Unknown command: Cobra suggests similar commands ("Did you mean ...?").

**Examples**:
```bash
mind help                    # Root help
mind help create             # Help for 'create' and its subcommands
mind help create iteration   # Help for 'create iteration' specifically
mind help check              # Help for 'check' and its subcommands
```

---

## 4. Shell Completion Specification

Cobra generates shell completions automatically for subcommands and flags. The following custom completions are registered:

### Custom Argument Completions

| Command | Position | Completes To |
|---------|----------|--------------|
| `mind create iteration <type>` | 1st arg | `new`, `enhancement`, `bugfix`, `refactor` |
| `mind create iteration <type> <name>` | 2nd arg | No completion (free text) |
| `mind workflow show <id>` | 1st arg | Existing iteration sequence numbers (e.g., `001`, `002`, `007`) |
| `mind handoff <id>` | 1st arg | Existing iteration sequence numbers |
| `mind completion <shell>` | 1st arg | `bash`, `zsh`, `fish` |
| `mind docs open <path-or-id>` | 1st arg | Filesystem paths under `docs/` + document IDs from `mind.toml` |
| `mind check convergence <file>` | 1st arg | Files matching `docs/knowledge/*-convergence.md` |
| `mind quality log <file>` | 1st arg | Files matching `docs/knowledge/*-convergence.md` |
| `mind help <command>` | 1st arg | All command names |

### Custom Flag Completions

| Flag | Completes To |
|------|--------------|
| `--zone` | `spec`, `blueprints`, `state`, `iterations`, `knowledge` |
| `--project-root` | Filesystem directories |
| `--log-file` | Filesystem paths |

### Subcommand Completion

All parent commands complete to their child subcommands:

- `mind create` -> `adr`, `blueprint`, `iteration`, `spike`, `convergence`, `brief`
- `mind docs` -> `list`, `tree`, `open`, `stubs`, `search`
- `mind check` -> `docs`, `refs`, `config`, `convergence`, `all`
- `mind workflow` -> `status`, `history`, `show`, `clean`
- `mind sync` -> `agents`
- `mind quality` -> `log`, `history`, `report`

### Implementation

```go
// In each command's init function:
createIterationCmd.RegisterFlagCompletionFunc("zone", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"spec", "blueprints", "state", "iterations", "knowledge"}, cobra.ShellCompDirectiveNoFileComp
})

createIterationCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    if len(args) == 0 {
        return []string{"new", "enhancement", "bugfix", "refactor"}, cobra.ShellCompDirectiveNoFileComp
    }
    return nil, cobra.ShellCompDirectiveNoFileComp
}
```

---

## 5. Error Messages Catalog

All user-facing error messages follow the pattern: **what happened**, **why it matters**, **what to do about it**.

| Error | Context | Remediation |
|-------|---------|-------------|
| `Not a Mind project (no .mind/ directory found)` | Any command run outside a project directory | Run `mind init` to initialize, or `cd` to a project directory |
| `Already initialized (.mind/ exists)` | `mind init` in an existing project | Use `mind doctor --fix` to repair, or remove `.mind/` to re-initialize |
| `Project brief missing: docs/spec/project-brief.md` | `preflight` (NEW_PROJECT/COMPLEX_NEW), `check docs` | Run `mind create brief` for guided creation, or `/discover` in Claude Code |
| `Project brief is a stub (no substantive content)` | `preflight` (NEW_PROJECT/COMPLEX_NEW), `check docs --strict` | Fill the brief with Vision, Key Deliverables, and Scope sections |
| `Business context gate: BLOCKED — brief required for {type}` | `preflight`, `run` with NEW_PROJECT or COMPLEX_NEW | Create or fill the project brief before running this workflow type |
| `mind.toml not found` | `status`, `check`, `reconcile`, `sync` | Run `mind init` to generate, or create `mind.toml` manually |
| `mind.toml parse error: {detail} at line {N}` | Any command that reads `mind.toml` | Fix the TOML syntax error at the indicated line |
| `Invalid iteration type: "{value}" (expected: new, enhancement, bugfix, refactor)` | `create iteration` | Use one of the valid types |
| `Iteration {id} not found` | `workflow show`, `handoff` | Run `mind workflow history` to see available iterations |
| `Workflow already active — cannot start a new pre-flight` | `preflight` when a workflow is in progress | Run `mind preflight --resume` or `mind workflow clean` first |
| `No interrupted workflow found` | `preflight --resume`, `run --resume` | Start a new workflow with `mind preflight "<request>"` |
| `claude CLI not found on $PATH` | `run` | Install Claude Code CLI: https://docs.anthropic.com/claude-code |
| `$EDITOR not set` | `docs open` | Set the `$EDITOR` environment variable (e.g., `export EDITOR=vim`) |
| `Document not found: {path}` | `docs open`, `check convergence`, `quality log` | Check the path and try `mind docs list` to find the correct file |
| `Ambiguous match for "{query}" — multiple documents found` | `docs open` with fuzzy match | Provide a more specific path or use the full document ID |
| `TUI requires an interactive terminal` | `tui` when stdout is not a TTY | Use `mind status` for non-interactive output, or run in a terminal |
| `quality-log.yml not found` | `quality history`, `quality report` | Run `mind quality log <convergence-file>` to create the first entry |
| `No quality rubric found in {file}` | `quality log` on a non-convergence file | Ensure the file has a `## Quality Rubric` section with dimension scores |
| `File already exists: {path}` | `create adr`, `create blueprint`, `create spike`, `create convergence` | Choose a different title or remove the existing file |
| `Permission denied: {path}` | Any write operation | Check file permissions and directory ownership |
| `Agent dispatch failed: {detail}` | `run` during agent execution | Check `claude` CLI status, review the error, and `mind run --resume` |
| `Deterministic gate failed: {command} exited {code}` | `handoff`, `run` during gate check | Fix build/lint/test failures, then re-run handoff or resume the workflow |
| `Reconciliation skipped (lock is fresh)` | `reconcile` without `--force` | Use `--force` to override, or wait for the lock to expire (1 hour) |
| `Invalid zone: "{value}" (expected: spec, blueprints, state, iterations, knowledge)` | `docs list --zone` | Use one of the valid zone names |
| `Stale workflow detected but not confirmed` | `workflow clean` on non-stale active workflow (non-interactive) | Use `--dry-run` to inspect, or confirm interactively |
| `Configuration error: {detail}` | `check config` finding invalid YAML | Fix the YAML syntax error in the indicated file |

---

## Cross-References

- **BP-01** ([01-mind-cli.md](01-mind-cli.md)) -- Command tree design, TUI mockups, project structure, implementation phases
- **BP-02** ([02-ai-workflow-bridge.md](02-ai-workflow-bridge.md)) -- Integration models A-D, MCP tool definitions, orchestration flow
- **BP-03** ([03-architecture.md](03-architecture.md)) -- 4-layer architecture, domain model, JSON output rendering, service contracts
- **Data Contracts** (`docs/spec/api-contracts.md`) -- JSON schema definitions for all `--json` outputs
- **Domain Model** (`docs/spec/domain-model.md`) -- Entity definitions referenced by exit codes and validation checks

---

> **What this document does NOT contain:**
> - Internal implementation details (module structure, Go code) -- see BP-03
> - JSON schema definitions (field-level contracts) -- see Data Contracts
> - TUI visual design details (mockups, key bindings) -- see BP-01
> - MCP tool input/output schemas -- see BP-02
