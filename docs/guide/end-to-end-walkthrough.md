# Mind Framework — End-to-End Walkthrough

> **Audience**: Developers using the Mind Framework for the first time.
> **Date**: 2026-03-13
> **CLI Version**: Built from `github.com/jf-ferraz/mind-cli` (Go 1.24+)

---

## Table of Contents

1. [What is Mind?](#1-what-is-mind)
2. [Architecture Overview](#2-architecture-overview)
3. [Installing the CLI](#3-installing-the-cli)
4. [Global Setup (One-Time)](#4-global-setup-one-time)
5. [Scenario A: Brand-New Project (No Codebase)](#5-scenario-a-brand-new-project-no-codebase)
6. [Scenario B: Existing Project Without Mind](#6-scenario-b-existing-project-without-mind)
7. [Scenario C: Existing Project With Mind Already Installed](#7-scenario-c-existing-project-with-mind-already-installed)
8. [Running Your First Workflow](#8-running-your-first-workflow)
9. [CLI Command Reference](#9-cli-command-reference)
10. [The mind.toml Configuration File](#10-the-mindtoml-configuration-file)
11. [The 5-Zone Documentation Structure](#11-the-5-zone-documentation-structure)
12. [Framework Commands (Slash Commands in Claude Code)](#12-framework-commands-slash-commands-in-claude-code)
13. [Agent Roles and the Workflow Pipeline](#13-agent-roles-and-the-workflow-pipeline)
14. [Quality Gates](#14-quality-gates)
15. [MCP Server Integration](#15-mcp-server-integration)
16. [Common Operations Cheat Sheet](#16-common-operations-cheat-sheet)
17. [Troubleshooting](#17-troubleshooting)

---

## 1. What is Mind?

Mind is a project intelligence framework that sits alongside your codebase. It has two parts:

- **`mind` CLI** — A Go binary you install on your machine. It manages project health, documentation, validation, framework artifacts, and an MCP server for AI integration.
- **`.mind/` directory** — A set of markdown agents, commands, conventions, and skills that orchestrate AI-assisted development workflows through Claude Code (or other AI tools).

Think of it this way: the CLI is the **toolbox**, and the `.mind/` directory is the **instruction manual** that tells AI agents how to work on your project.

---

## 2. Architecture Overview

```
Your Machine
├── ~/.config/mind/                  ← Global config (one per machine)
│   ├── config.toml                  ← Global preferences
│   ├── projects.toml                ← Project registry (@aliases)
│   ├── framework.lock               ← Installed framework version + checksums
│   ├── agents/                      ← Canonical agent definitions
│   ├── commands/                    ← Canonical command definitions
│   ├── conventions/                 ← Canonical conventions
│   ├── conversation/                ← Dialectical analysis system
│   ├── skills/                      ← Deep-dive skill guides
│   ├── docs/                        ← Framework documentation/templates
│   ├── platform/                    ← Platform integrations
│   ├── scripts/                     ← Utility scripts
│   ├── CLAUDE.md                    ← Framework routing hub
│   └── README.md                    ← Framework overview
│
├── ~/dev/projects/my-project/       ← Your project
│   ├── mind.toml                    ← Project manifest (config)
│   ├── .mind/                       ← Framework artifacts (local copy)
│   │   ├── agents/                  ← Agent definitions (can override global)
│   │   ├── commands/                ← Command definitions
│   │   ├── conventions/             ← Code quality conventions
│   │   ├── conversation/            ← Analysis system
│   │   ├── skills/                  ← Skill guides
│   │   ├── CLAUDE.md                ← Routing hub for this project
│   │   └── .framework-manifest     ← Tracks what came from global vs project
│   ├── .claude/
│   │   └── CLAUDE.md               ← Claude Code adapter (points to .mind/)
│   ├── .mcp.json                    ← MCP server auto-discovery
│   ├── docs/                        ← 5-zone documentation
│   │   ├── spec/                    ← Requirements, architecture, domain model
│   │   ├── blueprints/              ← Planning artifacts
│   │   ├── state/                   ← Workflow state, current status
│   │   ├── iterations/              ← Change tracking per iteration
│   │   └── knowledge/               ← Glossary, spikes, convergence analyses
│   └── src/                         ← Your actual code
```

**Key concept: Two-layer resolution.** When the framework needs an agent definition (say `analyst.md`), it looks in your project's `.mind/agents/` first. If not found, it falls back to the global `~/.config/mind/agents/`. This means you can override any framework artifact per-project.

---

## 3. Installing the CLI

### Prerequisites

- **Go 1.24+** installed ([https://go.dev/dl/](https://go.dev/dl/))
- **Git** installed
- A terminal (Linux, macOS, or WSL on Windows)

### Option A: Install from source with Makefile (recommended)

```bash
# Clone the repository
git clone https://github.com/jf-ferraz/mind-cli.git
cd mind-cli

# Build and install to $GOPATH/bin (binary named 'mind', with version info)
make install

# Verify it works
mind version
```

You should see output like:

```
mind v0.3.1 (<commit>) built 2026-03-13T00:00:00Z linux/amd64
```

### Option B: Build without installing globally

```bash
git clone https://github.com/jf-ferraz/mind-cli.git
cd mind-cli

# Build a binary in the current directory (with version info)
make build

# Move it somewhere in your PATH
sudo mv mind /usr/local/bin/
# or
mv mind ~/bin/   # if ~/bin is in your PATH

# Verify
mind version
```

### Option C: go install from remote

```bash
go install github.com/jf-ferraz/mind-cli@v0.3.1

# Note: go install names the binary 'mind-cli' (from the module name).
# Create a symlink if you want the 'mind' command:
ln -s "$(go env GOPATH)/bin/mind-cli" "$(go env GOPATH)/bin/mind"
# Fish shell: ln -s (go env GOPATH)/bin/mind-cli (go env GOPATH)/bin/mind

# Verify
mind version
# Version shows 'dev (unknown)' — this is expected without ldflags injection.
# Use Option A for full version info.
```

### Shell Completion (Optional)

Generate shell completions for a better experience:

```bash
# Bash
mind completion bash > ~/.bash_completion.d/mind

# Zsh
mind completion zsh > "${fpath[1]}/_mind"

# Fish
mind completion fish > ~/.config/fish/completions/mind.fish
```

---

## 4. Global Setup (One-Time)

After installing the CLI, you need to install the framework artifacts globally. This is a **one-time setup per machine**.

### Step 1: Get the framework source

The canonical framework lives in the `mind` repository (separate from `mind-cli`). You need a local copy:

```bash
git clone -b develop https://github.com/jf-ferraz/.mind.git ~/dev/projects/mind
```

The framework artifacts are in `~/dev/projects/mind/.mind/`.

### Step 2: Install the framework globally

```bash
mind framework install --source ~/dev/projects/mind
```

This command:
1. Detects the `.mind/` subdirectory inside the source path
2. Copies all artifact kinds (agents, commands, conventions, skills, conversation, docs, platform, scripts) to `~/.config/mind/`
3. Copies root files (CLAUDE.md, README.md) to `~/.config/mind/`
4. Computes SHA-256 checksums for every file
5. Writes `~/.config/mind/framework.lock` with version, source path, and checksums

Expected output:

```
Framework installed
  version:   2026.03.1
  artifacts: 42
  location:  ~/.config/mind/
```

### Step 3: Verify the installation

```bash
# Check status
mind framework status

# Expected output:
# Framework Status
#   installed:  yes
#   version:    2026.03.1
#   mode:       standalone
#   source:     /home/you/dev/projects/mind/.mind
#   drift:      0 files
```

```bash
# Run diagnostics
mind doctor

# Look for: "framework_installed: PASS"
```

### Step 4: Set up the global config (optional)

The first time you run a mind command, `~/.config/mind/config.toml` is created with sensible defaults. You can customize it:

```bash
# View current config
mind config show

# Edit in your $EDITOR
mind config edit

# Validate it
mind config validate
```

### Step 5: Register your projects (optional)

The project registry lets you reference projects by alias (e.g., `@my-app`):

```bash
# Register a project
mind registry add my-app ~/dev/projects/my-app

# List all registered projects
mind registry list

# Resolve an alias
mind registry resolve @my-app
# Output: /home/you/dev/projects/my-app

# Check all registered paths exist
mind registry check
```

---

## 5. Scenario A: Brand-New Project (No Codebase)

You're starting a project from scratch. No code, no files, nothing.

### Step 1: Create a directory and initialize git

```bash
mkdir ~/dev/projects/my-new-app
cd ~/dev/projects/my-new-app
git init
```

### Step 2: Initialize the Mind project

```bash
mind init --name my-new-app
```

This creates:

```
my-new-app/
├── mind.toml                    ← Project manifest
├── .mind/                       ← Empty (framework not materialized yet)
├── .claude/
│   └── CLAUDE.md               ← Claude Code adapter
├── docs/
│   ├── spec/
│   │   ├── project-brief.md    ← Stub (needs filling)
│   │   ├── requirements.md     ← Stub
│   │   ├── architecture.md     ← Stub
│   │   ├── domain-model.md     ← Stub
│   │   └── decisions/          ← Empty (for ADRs)
│   ├── blueprints/
│   │   └── INDEX.md            ← Stub
│   ├── state/
│   │   ├── current.md          ← Stub
│   │   └── workflow.md         ← Stub
│   ├── iterations/             ← Empty
│   └── knowledge/
│       └── glossary.md         ← Stub
```

**What happened automatically**: If the framework is installed globally, `mind init` detects the framework version from `~/.config/mind/framework.lock` and includes a `[framework]` section in `mind.toml`:

```toml
[framework]
version = "2026.03.1"
mode = "standalone"
```

### Step 3: Materialize the framework

The `.mind/` directory is empty after init. You need to populate it with framework artifacts:

```bash
mind framework materialize
```

This copies all agents, commands, conventions, skills, conversation files, and root files from `~/.config/mind/` into your project's `.mind/` directory. Project-specific overrides (if any existed) are preserved.

Expected output:

```
Materialized framework
  version:    2026.03.1
  total:      42
  copied:     42
  kept:       0
```

### Step 4: Set up the MCP server (for Claude Code integration)

Create `.mcp.json` in your project root:

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

### Step 5: Configure mind.toml

Edit `mind.toml` to match your project. At minimum, fill in:

```toml
[project]
name = "my-new-app"
description = "A REST API for managing tasks"
type = "api"

[project.stack]
language = "go"
framework = "net/http"
testing = "go test"

[project.commands]
test = "go test ./..."
lint = "golangci-lint run"
build = "go build -o my-new-app ."
```

The `[project.commands]` section is important — these commands are used by the **deterministic gate** to verify your code compiles and passes tests before the reviewer agent runs.

### Step 6: Fill the project brief

The project brief is the most important document for AI-assisted workflows. Without it, the analyst agent has to guess your intent.

**Option A — Interactive CLI:**

```bash
mind create brief
```

This prompts you for: Vision, Key Deliverables, In Scope, Out of Scope, and Constraints.

**Option B — Use the `/discover` command in Claude Code:**

Open Claude Code in your project directory and run:

```
/discover "A task management REST API with user authentication and team collaboration"
```

The discovery agent asks 5-8 targeted questions and produces `docs/spec/project-brief.md`.

**Option C — Write it manually:**

Edit `docs/spec/project-brief.md` directly. At minimum, include:

```markdown
# Project Brief

## Vision
A REST API for managing tasks with user authentication.

## Key Deliverables
- User registration and login (JWT)
- CRUD operations for tasks
- Team-based task sharing

## Scope

### In Scope
- REST API endpoints
- PostgreSQL persistence
- JWT authentication

### Out of Scope
- Frontend/UI
- Email notifications
- Mobile app

## Constraints
- Go 1.24+
- Single binary deployment
- Must run in Docker
```

### Step 7: Initial commit

```bash
git add -A
git commit -m "docs: initialize mind framework project"
```

### Step 8: Verify everything

```bash
# Project health
mind status

# Documentation structure
mind docs tree

# Run all validation checks
mind check all

# Full diagnostics
mind doctor
```

You're ready to start development. Jump to [Section 8: Running Your First Workflow](#8-running-your-first-workflow).

---

## 6. Scenario B: Existing Project Without Mind

You have an existing codebase — say a Go API with 50 files — and you want to add Mind to it.

### Step 1: Navigate to your project

```bash
cd ~/dev/projects/existing-api
```

### Step 2: Initialize with --from-existing

The `--from-existing` flag tells mind to preserve any documentation files you already have:

```bash
mind init --from-existing
```

This:
- Creates `.mind/`, `docs/` zones, `.claude/CLAUDE.md`
- **Preserves** any existing files (e.g., if you already have `docs/spec/architecture.md`, it won't be overwritten)
- Creates stubs only for files that don't exist yet
- Auto-detects framework version from global installation

Example output:

```
Initialized project: existing-api
  created:   12 files
  preserved: 3 files (docs/spec/architecture.md, docs/spec/requirements.md, mind.toml)
```

### Step 3: Materialize the framework

```bash
mind framework materialize
```

### Step 4: Configure mind.toml

Since `--from-existing` preserves an existing `mind.toml` (if one exists), you may need to add or update sections. Open it and ensure it has:

```toml
[manifest]
schema = "mind/v1.0"
generation = 1

[project]
name = "existing-api"
description = "Production REST API for order management"
type = "api"

[project.stack]
language = "go"
framework = "gin"
testing = "go test"

[project.commands]
test = "go test ./..."
lint = "golangci-lint run ./..."
build = "go build -o api ./cmd/api"

[framework]
version = "2026.03.1"
mode = "standalone"

[governance]
max-retries = 2
default-branch = "main"

[profiles]
active = []
```

### Step 5: Register your documents

If you have existing documentation, register it in `mind.toml` so the reconciliation engine can track changes:

```toml
[documents.spec.requirements]
id = "doc:spec/requirements"
path = "docs/spec/requirements.md"
zone = "spec"
status = "active"

[documents.spec.architecture]
id = "doc:spec/architecture"
path = "docs/spec/architecture.md"
zone = "spec"
status = "active"

[[graph]]
from = "doc:spec/requirements"
to = "doc:spec/architecture"
type = "informs"
```

### Step 6: Run reconciliation

The reconciliation engine computes hashes for all registered documents and tracks changes:

```bash
# First run — creates mind.lock with baseline hashes
mind reconcile

# Check mode — verify without modifying (good for CI)
mind reconcile --check

# Force mode — reset all hashes (useful after bulk edits)
mind reconcile --force

# View dependency graph
mind reconcile --graph
```

### Step 7: Fill the project brief

Even for existing projects, the brief helps AI agents understand your project's purpose:

```bash
mind create brief
```

Or use `/discover` in Claude Code to generate one interactively.

### Step 8: Set up MCP and commit

```bash
# Create .mcp.json (see Scenario A, Step 4)

# Add to .gitignore if needed
echo "mind.lock" >> .gitignore  # Optional: lock file can be committed or ignored

git add -A
git commit -m "docs: add mind framework to existing project"
```

### Step 9: Verify

```bash
mind status
mind check all
mind doctor
```

---

## 7. Scenario C: Existing Project With Mind Already Installed

Your project already has `.mind/` and `mind.toml`. Maybe a teammate set it up, or you're returning to a project after some time. Here's how to get up to speed and keep things current.

### Step 1: Verify your global framework is installed

```bash
mind framework status
```

If it says "not installed":

```bash
mind framework install --source /path/to/mind-repo
```

### Step 2: Check for framework drift

Compare your project's `.mind/` against the global canonical:

```bash
mind framework diff
```

Output shows differences:

```
M  agents/analyst.md          ← Modified (project override or outdated)
D  conversation/agents/moderator.md  ← Missing from project
A  agents/custom-agent.md     ← Extra (project-specific)
```

- **M** = modified (content differs between project and global)
- **D** = deleted/missing (in global but not in project)
- **A** = added (in project but not in global)

### Step 3: Update to latest framework

If the global framework has been updated (new agents, convention changes):

```bash
mind framework update
```

This:
- Detects which global artifacts have changed
- Copies only changed artifacts to your project
- **Preserves all project overrides** (files you've customized)
- Removes artifacts that no longer exist in global
- Updates the `.framework-manifest`

Output:

```
Framework updated
  version:  2026.03.2
  added:    2 (conversation/agents/new-persona.md, skills/security/SKILL.md)
  updated:  3 (agents/analyst.md, agents/reviewer.md, conventions/shared.md)
  removed:  0
  kept:     37 (including 5 project overrides)
```

### Step 4: Re-install global framework (when mind repo updates)

If you've pulled new changes in the mind framework repository:

```bash
cd ~/dev/projects/mind
git pull

# Re-install globally (requires --force since already installed)
mind framework install --source . --force

# Then update your project
cd ~/dev/projects/my-project
mind framework update
```

### Step 5: Run diagnostics

```bash
mind doctor
mind status
mind check all
```

---

## 8. Running Your First Workflow

This is where Mind shines. The framework orchestrates AI agents to analyze, design, implement, test, and review code changes.

### Prerequisites

- Project initialized with Mind (Scenario A, B, or C)
- Project brief filled (`docs/spec/project-brief.md` is not a stub)
- Claude Code installed and configured
- MCP server configured (`.mcp.json` in project root)

### Understanding the Workflow Types

| When you say... | Classification | What happens |
|---|---|---|
| "fix the login 500 error" | **BUG_FIX** | analyst → developer → tester → reviewer |
| "add pagination to the API" | **ENHANCEMENT** | analyst → [architect] → developer → tester → reviewer |
| "refactor the data layer" | **REFACTOR** | analyst → developer → reviewer |
| "create a new notification service" | **NEW_PROJECT** | analyst → architect → developer → tester → reviewer |
| "analyze: should we use GraphQL or REST?" | **COMPLEX_NEW** | conversation → analyst → architect → developer → tester → reviewer |

### The Two Approaches

#### Approach 1: Full CLI-Assisted Workflow (Recommended)

**Step 1 — Pre-flight:**

```bash
mind preflight "add pagination to the user listing API endpoint"
```

This runs 7 checks automatically:
1. Classifies the request → `ENHANCEMENT`
2. Checks the project brief → PASS/WARN
3. Validates documentation (17 checks) → PASS/FAIL count
4. Creates an iteration folder → `docs/iterations/006-enhancement-pagination/`
5. Creates a git branch → `enhancement/pagination`
6. Writes workflow state → `docs/state/workflow.md`
7. Generates an orchestrator prompt

Output:

```
┌─ Pre-flight ─────────────────────────────────────────────┐
│ Type:      ENHANCEMENT                                    │
│ Chain:     analyst → developer → tester → reviewer        │
│ Branch:    enhancement/pagination                         │
│ Iteration: docs/iterations/006-enhancement-pagination/    │
│ Brief:     PASS                                           │
│ Docs:      14 pass, 3 warn                                │
│                                                           │
│ Run /workflow in Claude Code to start.                    │
└───────────────────────────────────────────────────────────┘
```

**Step 2 — Execute the workflow in Claude Code:**

Open Claude Code in your project and type:

```
/workflow "add pagination to the user listing API endpoint"
```

The orchestrator takes over:
- Dispatches each agent in sequence
- Runs quality gates between agents
- Creates iteration artifacts (changes.md, test-summary.md, validation.md)
- The reviewer signs off with APPROVED / APPROVED_WITH_NOTES / NEEDS_REVISION

**Step 3 — Post-workflow handoff:**

After Claude Code finishes the workflow:

```bash
mind handoff 006-enhancement-pagination
```

This runs 5 post-workflow steps:
1. Validates iteration artifacts (overview, changes, test-summary, validation, retrospective)
2. Runs deterministic gate (build + lint + test from `mind.toml`)
3. Updates `docs/state/current.md`
4. Clears workflow state
5. Reports branch status (commits ahead of main)

**Step 4 — Review and merge:**

```bash
# Review the changes
git log --oneline enhancement/pagination

# Merge when satisfied
git checkout main
git merge enhancement/pagination --no-ff
```

#### Approach 2: Direct Claude Code Workflow

Skip the CLI pre-flight and handoff. Just open Claude Code and run:

```
/workflow "add pagination to the user listing API endpoint"
```

The orchestrator inside Claude Code handles everything — classification, iteration creation, branch creation, agent dispatch, and quality gates. The CLI pre-flight/handoff adds extra validation and state tracking, but is not strictly required.

### Resuming an Interrupted Workflow

If a workflow was interrupted (Claude Code session ended, timeout, etc.):

```bash
# Check for in-progress workflow
mind preflight --resume
```

Output:

```
┌─ Resumable Workflow ─────────────────────────────────────┐
│ Type:       ENHANCEMENT                                   │
│ Iteration:  006-enhancement-pagination                    │
│ Last Agent: developer                                     │
│ Remaining:  tester → reviewer                             │
│ Branch:     enhancement/pagination                        │
│                                                           │
│ Run /workflow "Resume the interrupted workflow"           │
│ in Claude Code to continue.                               │
└───────────────────────────────────────────────────────────┘
```

Then in Claude Code:

```
/workflow "Resume the interrupted workflow"
```

---

## 9. CLI Command Reference

### Global Flags (available on all commands)

| Flag | Short | Description |
|---|---|---|
| `--json` | `-j` | Output in JSON format |
| `--no-color` | | Disable colored output |
| `--project-root` | `-p` | Explicit project root path (auto-detects if omitted) |

### Exit Codes

| Code | Meaning | Example |
|---|---|---|
| 0 | Success | Command completed successfully |
| 1 | Validation failure | Check failed, missing data, malformed input |
| 2 | Runtime error | I/O failure, unexpected error |
| 3 | Configuration error | Project not found, not initialized, schema violation |
| 4 | Staleness detected | `mind reconcile --check` found stale documents |

### Commands At a Glance

```
mind
├── init                          Initialize a new project
│   ├── --name, -n <name>         Project name (default: dir name)
│   ├── --with-github             Create .github/agents/ adapter
│   └── --from-existing           Preserve existing docs
│
├── status                        Show project health
├── doctor                        Run full diagnostics
│   └── --fix                     Auto-fix resolvable issues
│
├── check                         Run validation checks
│   ├── docs                      17-check documentation suite
│   │   └── --strict              Promote warnings to failures
│   ├── refs                      11-check cross-reference suite
│   ├── config                    Validate mind.toml schema
│   └── all                       Run all suites
│       └── --strict              Promote warnings to failures
│
├── create                        Create artifacts
│   ├── adr <title>               Auto-numbered ADR
│   ├── blueprint <title>         Auto-numbered blueprint
│   ├── iteration <type> <name>   Iteration folder (new|enhancement|bugfix|refactor)
│   ├── spike <title>             Spike report template
│   ├── convergence <title>       Convergence analysis template
│   └── brief                     Interactive project brief
│
├── docs                          Manage documentation
│   ├── list                      List all documents by zone
│   │   └── --zone <zone>         Filter by zone
│   ├── tree                      Show doc tree with stub markers
│   ├── stubs                     List stub documents only
│   ├── search <query>            Full-text search
│   └── open <path-or-id>         Open in $EDITOR
│
├── workflow                      Inspect workflow state
│   ├── status                    Show current workflow state
│   └── history                   List all iterations
│
├── brief                         Show project brief status
├── iterations (iter, iters)      List all iterations
│
├── reconcile                     Document hash reconciliation
│   ├── --check                   Read-only verification
│   ├── --force                   Re-hash everything
│   └── --graph                   Show dependency graph
│
├── preflight <request>           Pre-flight checks + iteration setup
│   └── --resume                  Check for in-progress workflow
│
├── handoff <iteration-id>        Post-workflow validation + state update
│
├── framework                     Manage framework installation
│   ├── install                   Install to global config
│   │   ├── --source, -s <path>   Framework source directory
│   │   └── --force               Overwrite existing
│   ├── status                    Show version and drift
│   ├── diff                      Compare project vs global
│   ├── materialize               Populate .mind/ from global
│   └── update                    Re-materialize changed artifacts
│
├── config                        Global configuration
│   ├── show                      Display config
│   ├── edit                      Open in $EDITOR
│   ├── path                      Print config file path
│   └── validate                  Validate config
│
├── registry                      Project registry
│   ├── list                      List all projects
│   ├── add <alias> <path>        Register a project
│   ├── remove <alias>            Remove a project
│   ├── resolve <@alias>          Resolve alias to path
│   └── check                     Validate all paths
│
├── tui                           Interactive dashboard (5 tabs)
├── serve                         Start MCP server (JSON-RPC 2.0 stdio)
├── version                       Print version info
│   └── --short                   Version string only
└── completion                    Shell completions (bash|zsh|fish)
```

---

## 10. The mind.toml Configuration File

This is your project's manifest. It lives at the project root.

### Complete Annotated Example

```toml
# ── Manifest ─────────────────────────────────────────────
# Schema version and generation counter. Do not edit generation
# manually — it auto-increments on every write.

[manifest]
schema = "mind/v1.0"           # Required. Format: mind/vN.N
generation = 1                  # Required. Starts at 1, auto-incremented
updated = 2026-03-13T10:30:00Z  # Set automatically

# ── Project ──────────────────────────────────────────────
# Project identity and metadata.

[project]
name = "my-api"                 # Required. Kebab-case: ^[a-z][a-z0-9-]*$
description = "Order management REST API"
type = "api"                    # cli | api | library | webapp | service

[project.stack]
language = "go"                 # Programming language
framework = "gin"               # Web/app framework (freeform)
testing = "go test"             # Test framework

# These commands are used by the deterministic gate.
# If a command is empty, that step is skipped.
[project.commands]
dev = "go run ./cmd/api"        # Local dev startup
test = "go test ./..."          # Test suite (USED BY GATE)
lint = "golangci-lint run"      # Linting (USED BY GATE)
typecheck = "go vet ./..."      # Type checking (USED BY GATE)
build = "go build -o api ."    # Build command (USED BY GATE)

# ── Framework ────────────────────────────────────────────
# Optional. Auto-detected by mind init if framework is installed.

[framework]
version = "2026.03.1"          # CalVer. Must match installed global version.
mode = "standalone"             # standalone | thin

# ── Governance ───────────────────────────────────────────
# Workflow rules and policies.

[governance]
max-retries = 2                 # Max retry loops across entire workflow (0-5)
review-policy = "evidence-based"
commit-policy = "conventional"
branch-strategy = "type-descriptor"
default-branch = "main"         # Used by preflight/handoff for branch comparison

# ── Profiles ─────────────────────────────────────────────
# Reserved for future use.

[profiles]
active = []

# ── Documents ────────────────────────────────────────────
# Register documents for the reconciliation engine.
# Format: [documents.{zone}.{name}]

[documents.spec.requirements]
id = "doc:spec/requirements"
path = "docs/spec/requirements.md"
zone = "spec"
status = "active"               # draft | active | complete

[documents.spec.architecture]
id = "doc:spec/architecture"
path = "docs/spec/architecture.md"
zone = "spec"
status = "active"

[documents.spec.domain-model]
id = "doc:spec/domain-model"
path = "docs/spec/domain-model.md"
zone = "spec"
status = "draft"

# ── Dependency Graph ─────────────────────────────────────
# Edges between documents. Changes propagate staleness.
# Types: informs | requires | validates

[[graph]]
from = "doc:spec/requirements"
to = "doc:spec/architecture"
type = "informs"

[[graph]]
from = "doc:spec/architecture"
to = "doc:spec/domain-model"
type = "requires"
```

### Validation Rules

Run `mind check config` to validate your `mind.toml`. The checks:

| # | Check | Fails If |
|---|---|---|
| 1 | File exists and parses | mind.toml missing or invalid TOML |
| 2 | Schema format | `manifest.schema` doesn't match `mind/vN.N` |
| 3 | Generation | `manifest.generation` < 1 |
| 4 | Project name | Not kebab-case |
| 5 | Project type | Not one of the 5 valid types (or empty) |
| 6 | Document IDs | Don't match `doc:{zone}/{name}` |
| 7 | Document paths | Don't start with `docs/` or end with `.md` |
| 8 | Document zones | Not one of the 5 valid zones |
| 9 | Document status | Not `draft`, `active`, or `complete` |
| 10 | Max retries (warn) | Outside 0-5 range |
| 11 | Graph edge IDs | Reference undeclared documents |
| 12 | Graph edge types | Not `informs`, `requires`, or `validates` |

---

## 11. The 5-Zone Documentation Structure

Mind organizes all project documentation into 5 zones. Each zone has a specific purpose and lifecycle:

```
docs/
├── spec/                  ← SPECIFICATIONS (updated incrementally)
│   ├── project-brief.md   ← Vision, scope, constraints
│   ├── requirements.md    ← Functional requirements (FR-N IDs)
│   ├── architecture.md    ← System design, component map
│   ├── domain-model.md    ← Entities, business rules (BR-N IDs)
│   └── decisions/         ← Architecture Decision Records (ADR-NNN)
│
├── blueprints/            ← PLANNING ARTIFACTS (stable, historical)
│   ├── INDEX.md           ← Auto-managed index
│   └── 01-system-arch.md  ← Blueprint documents (BP-NN)
│
├── state/                 ← RUNTIME CONTEXT (overwritten freely)
│   ├── current.md         ← Active work, known issues, next priorities
│   └── workflow.md        ← In-progress workflow state
│
├── iterations/            ← CHANGE TRACKING (append-only)
│   ├── 001-new-core-cli/
│   │   ├── overview.md    ← What was done, scope, type
│   │   ├── changes.md     ← Files modified/created
│   │   ├── test-summary.md ← Test results, coverage
│   │   ├── validation.md   ← Review findings
│   │   └── retrospective.md ← Lessons learned
│   └── 002-enhancement-pagination/
│       └── ...
│
└── knowledge/             ← DOMAIN REFERENCE (updated when understanding changes)
    ├── glossary.md        ← Domain terms
    ├── auth-convergence.md ← Analysis output
    └── redis-spike.md     ← Investigation results
```

### Zone Rules

| Zone | Can Overwrite? | Created By | Purpose |
|---|---|---|---|
| `spec/` | Updated incrementally | Analyst, Architect | Living specifications |
| `blueprints/` | Append-only (new files) | Architect, You | Historical planning intent |
| `state/` | Freely overwritten | Orchestrator, Handoff | Current runtime context |
| `iterations/` | Append-only (new folders) | Orchestrator, Developer, Tester, Reviewer | Change history |
| `knowledge/` | Updated as understanding changes | Conversation module, You | Domain knowledge |

### Creating Documents

```bash
# Create an ADR (auto-numbered)
mind create adr "Use PostgreSQL for persistence"
# Creates: docs/spec/decisions/001-use-postgresql-for-persistence.md

# Create a blueprint
mind create blueprint "Authentication System"
# Creates: docs/blueprints/01-authentication-system.md

# Create a spike report
mind create spike "Redis vs Memcached for caching"
# Creates: docs/knowledge/redis-vs-memcached-for-caching-spike.md

# Create a convergence analysis template
mind create convergence "API Design Strategy"
# Creates: docs/knowledge/api-design-strategy-convergence.md

# Create an iteration manually
mind create iteration enhancement "user-pagination"
# Creates: docs/iterations/007-enhancement-user-pagination/ with 5 template files
```

### Browsing Documents

```bash
# Tree view with stub markers
mind docs tree

# List by zone
mind docs list --zone spec

# Find stub documents (incomplete)
mind docs stubs

# Search across all docs
mind docs search "authentication"

# Open a document in your editor
mind docs open docs/spec/requirements.md
mind docs open architecture    # fuzzy match
```

---

## 12. Framework Commands (Slash Commands in Claude Code)

When you open Claude Code in a Mind project, you get 4 slash commands:

### `/workflow "description"`

The main entry point. Classifies your request and dispatches the appropriate agent chain.

```
/workflow "fix the 500 error on the /api/users endpoint"
/workflow "add WebSocket support for real-time notifications"
/workflow "refactor the data access layer to use repository pattern"
/workflow "create a new notification microservice"
```

**Classification prefixes** (optional, helps the orchestrator):

| Prefix | Forces Classification |
|---|---|
| `fix:` or `bug:` | BUG_FIX |
| `add:` or `feature:` | ENHANCEMENT |
| `refactor:` | REFACTOR |
| `create:` or `new:` | NEW_PROJECT |
| `analyze:` or `explore:` | COMPLEX_NEW |

### `/discover "idea"`

Interactive exploration for vague ideas. Use this **before** `/workflow` when you don't have clear requirements yet.

```
/discover "I want to build something that helps developers track tech debt"
```

The discovery agent asks 5-8 targeted questions, then produces `docs/spec/project-brief.md`.

### `/analyze "topic"`

Structured dialectical analysis. Spawns 3-4 AI personas that debate a design question from different perspectives, then synthesizes a convergence document with ranked recommendations.

```
/analyze "Should we use GraphQL or REST for our API?"
/analyze "Monolith vs microservices for our scaling needs"
```

Output: `docs/knowledge/{topic}-convergence.md` with:
- Decision matrix (3+ options scored)
- Recommendations with confidence levels
- Evidence registry
- Quality rubric score

### `/init`

Guided setup — detects missing documentation and helps create what's needed. Different from `mind init` (the CLI command). This slash command runs validation scripts and interactively fills gaps.

---

## 13. Agent Roles and the Workflow Pipeline

### The Agent Chain

When you run `/workflow`, agents execute in sequence. Each agent reads the previous agent's output and produces its own artifact.

```
┌───────────┐   ┌───────────┐   ┌───────────┐   ┌──────────┐   ┌──────────┐
│  Analyst  │──→│ Architect │──→│ Developer │──→│  Tester  │──→│ Reviewer │
│           │   │ (if needed│   │           │   │          │   │          │
│ Scope &   │   │  for new  │   │ Implement │   │ Write &  │   │ Evidence │
│ Require-  │   │  struct)  │   │ code      │   │ run      │   │ based    │
│ ments     │   │           │   │           │   │ tests    │   │ review   │
└───────────┘   └───────────┘   └───────────┘   └──────────┘   └──────────┘
     │                │                │               │              │
     ▼                ▼                ▼               ▼              ▼
  Micro-          (optional)      Micro-          Deterministic   validation.md
  Gate A                          Gate B            Gate         retrospective.md
```

### Agent Details

| Agent | Model | What It Does | Output |
|---|---|---|---|
| **Analyst** | Opus | Reads brief + code. Defines scope, requirements (FR-N), acceptance criteria (GIVEN/WHEN/THEN), success metrics. | `requirements-delta.md` or `issue-analysis.md` |
| **Architect** | Opus | Designs structure. Component map, data model, API contracts, key decisions with rationale. | `architecture-delta.md` or `architecture.md` |
| **Developer** | Sonnet | Implements code. Follows spec, writes conventional commits, stays within scope. | Actual code changes + `changes.md` |
| **Tester** | Sonnet | Writes tests from acceptance criteria and business rules. Runs test suite. | Test files + `test-summary.md` |
| **Reviewer** | Opus | Evidence-based review. Dual-path verification for MUST findings. Signs off or requests revision. | `validation.md` + `retrospective.md` |

### Chains by Request Type

| Type | Chain | When |
|---|---|---|
| BUG_FIX | analyst → developer → tester → reviewer | Fixing defects |
| ENHANCEMENT | analyst → [architect] → developer → tester → reviewer | Adding features (architect only if structural) |
| REFACTOR | analyst → developer → reviewer | Restructuring without behavior change |
| NEW_PROJECT | analyst → architect → developer → tester → reviewer | Building from scratch |
| COMPLEX_NEW | conversation → analyst → architect → developer → tester → reviewer | Major decisions needing analysis first |

### The Conversation Module (COMPLEX_NEW only)

For `COMPLEX_NEW` requests, a full dialectical analysis runs before the dev workflow:

1. **Persona selection**: 3-4 AI personas chosen (architect, pragmatist, critic, researcher)
2. **Opening positions**: Each persona writes an independent position paper
3. **Diversity audit**: Moderator checks for duplicate perspectives
4. **Cross-examination**: Personas challenge each other with evidence-based questions
5. **Rebuttal**: Personas concede, rebut, or partially accept challenges
6. **Convergence synthesis**: Decision matrix, ranked recommendations, evidence audit
7. **Quality scoring**: 6-dimension rubric (target ≥ 3.6/5.0)

The convergence output feeds into the analyst and architect as context.

---

## 14. Quality Gates

Quality gates are checkpoints that run between agents to catch problems early.

### Gate 0 — After Conversation Analysis (COMPLEX_NEW only)

| Check | Requirement |
|---|---|
| Executive Summary | Contains clear architectural recommendation |
| Decision Matrix | ≥ 3 options scored |
| Recommendations | ≥ 3 with confidence levels |
| Quality Rubric | Score ≥ 3.0/5.0 |

If it fails: conversation-moderator retries once.

### Micro-Gate A — After Analyst

| Check | Requirement |
|---|---|
| Acceptance criteria | Exist and use GIVEN/WHEN/THEN format |
| Scope boundary | In-scope AND out-of-scope explicitly defined |
| Success metrics | Quantified (no vague "fast" or "user-friendly") |
| Requirements | Traceable with FR-N or AC-N identifiers |
| Assumptions | Present if project brief was missing |

If it fails: analyst retries once with specific failing checks.

### Micro-Gate B — After Developer

| Check | Requirement |
|---|---|
| changes.md | Exists in iteration folder |
| Files on disk | All files referenced in changes.md exist |
| Scope compliance | No files outside analyst's scope modified |
| Requirement coverage | Each FR-N has a corresponding change or deferral justification |

If it fails: developer retries once.

### Deterministic Gate — Before Reviewer

Runs the commands from `[project.commands]` in mind.toml:

```bash
# Each non-empty command is executed:
go build -o api .          # build
golangci-lint run          # lint
go vet ./...               # typecheck
go test ./...              # test
```

All must exit 0. If any fails: developer retries once.

**Total retry budget: 2 across the entire workflow.** After retries, proceed to reviewer with documented concerns.

---

## 15. MCP Server Integration

The `mind serve` command starts an MCP server that gives Claude Code access to 16 tools for project intelligence.

### Setup

Create `.mcp.json` in your project root:

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

Claude Code auto-discovers this file and connects to the MCP server.

### Available MCP Tools

| Tool | Description |
|---|---|
| `mind_status` | Project health report |
| `mind_doctor` | Full diagnostics |
| `mind_check_brief` | Project brief status and completeness |
| `mind_check_gate` | Run deterministic gate (build/lint/test) |
| `mind_validate_docs` | 17-check documentation validation |
| `mind_validate_refs` | 11-check cross-reference validation |
| `mind_list_iterations` | List all iterations |
| `mind_show_iteration` | Show a specific iteration's artifacts |
| `mind_create_iteration` | Create a new iteration folder |
| `mind_read_state` | Read workflow or current state |
| `mind_update_state` | Update workflow or current state |
| `mind_list_stubs` | Find incomplete documents |
| `mind_search_docs` | Full-text document search |
| `mind_read_config` | Read mind.toml configuration |
| `mind_log_quality` | Parse convergence files into quality log |
| `mind_suggest_next` | Suggest next priorities |

---

## 16. Common Operations Cheat Sheet

### Daily Development

```bash
# Check project health
mind status

# See what needs attention
mind doctor

# Check for stale documents
mind reconcile --check

# Open a doc for editing
mind docs open architecture
```

### Before Starting Work

```bash
# Pre-flight a new task
mind preflight "add rate limiting to the API"

# Then in Claude Code:
/workflow "add rate limiting to the API"
```

### After Completing Work

```bash
# Hand off the completed iteration
mind handoff 007-enhancement-rate-limiting

# Verify everything
mind check all
```

### Framework Maintenance

```bash
# Check for framework updates
mind framework diff

# Update framework artifacts
mind framework update

# Reinstall framework (after pulling new mind repo changes)
mind framework install --source ~/dev/projects/mind --force
mind framework update
```

### Documentation Management

```bash
# Create artifacts
mind create adr "Switch to Redis for sessions"
mind create spike "WebSocket scaling options"

# Browse
mind docs tree
mind docs stubs
mind docs search "authentication"

# Reconcile (detect/propagate document changes)
mind reconcile
```

### TUI Dashboard

```bash
# Launch interactive dashboard
mind tui
```

The TUI has 5 tabs:
1. **Status** — Project health overview
2. **Documents** — Browse and inspect docs
3. **Iterations** — Iteration history
4. **Checks** — Validation results
5. **Quality** — Convergence analysis quality tracking

---

## 17. Troubleshooting

### "Project not found" / "not initialized"

```bash
# Make sure you're in a directory with .mind/
ls -la .mind/

# If missing, initialize
mind init

# Or specify the project root explicitly
mind status --project-root /path/to/my-project
```

### "Framework not installed"

```bash
# Install the framework globally
mind framework install --source /path/to/mind-repo

# Verify
mind framework status
```

### "Framework already installed (use --force to overwrite)"

```bash
# Re-install with --force
mind framework install --source /path/to/mind-repo --force
```

### mind.toml validation failures

```bash
# See what's wrong
mind check config

# Common fixes:
# - project.name must be kebab-case (lowercase, hyphens only)
# - schema must be "mind/v1.0"
# - generation must be >= 1
# - document IDs must be "doc:zone/name" format
```

### "No project brief" warnings

```bash
# Check brief status
mind brief

# Create one interactively
mind create brief

# Or use discovery in Claude Code
# /discover "your project idea"
```

### Stale documents detected

```bash
# See what's stale
mind reconcile --check

# View the dependency graph
mind reconcile --graph

# Force re-hash everything (nuclear option)
mind reconcile --force
```

### MCP server not connecting

1. Verify `.mcp.json` exists in your project root
2. Verify `mind` is in your PATH: `which mind`
3. Test the server manually: `mind serve` (should wait for JSON-RPC input)
4. Check Claude Code's MCP panel for connection errors

### Framework diff shows unexpected differences

```bash
# See all differences
mind framework diff

# M entries = modified (you customized them, or global updated)
# D entries = missing (global has it, your project doesn't)
# A entries = extra (you added project-specific files)

# To re-sync with global:
mind framework update
```

### Workflow stuck / interrupted

```bash
# Check for in-progress workflow
mind preflight --resume

# If it shows a resumable workflow, continue in Claude Code:
# /workflow "Resume the interrupted workflow"

# If you want to abandon it:
# Clear docs/state/workflow.md manually or start a new workflow
```
