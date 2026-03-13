# mind-cli

Single-binary Go CLI for the Mind Agent Framework. Provides project health diagnostics, documentation validation, document scaffolding, framework lifecycle management, and AI workflow orchestration.

## Quick Start for Testers

Mind requires two repositories: this CLI and the [.mind](https://github.com/jf-ferraz/.mind) framework.

```bash
# 1. Install the CLI
go install github.com/jf-ferraz/mind-cli@v0.3.1
mv "$(go env GOPATH)/bin/mind-cli" "$(go env GOPATH)/bin/mind"
# Fish shell: mv (go env GOPATH)/bin/mind-cli (go env GOPATH)/bin/mind

# 2. Clone the framework source (needed for framework install)
git clone -b develop https://github.com/jf-ferraz/.mind.git ~/dev/mind

# 3. Install framework artifacts globally (works from any directory)
mind framework install --source ~/dev/mind

# 4. Create a test project
mkdir /tmp/my-project && cd /tmp/my-project
git init && echo "# My Project" > README.md && git add . && git commit -m "init"

# 5. Initialize Mind in the project
mind init

# 6. Populate framework artifacts
mind framework materialize

# 7. Verify everything works
mind status
mind doctor
mind check all
```

## Getting Started

### Prerequisites

- Go 1.24+
- Git

### Install

```bash
go install github.com/jf-ferraz/mind-cli@latest

# Note: go install names the binary 'mind-cli' (from the module name).
# Rename it so all commands work as documented:
mv "$(go env GOPATH)/bin/mind-cli" "$(go env GOPATH)/bin/mind"
# Fish shell: mv (go env GOPATH)/bin/mind-cli (go env GOPATH)/bin/mind
```

Or build from source (recommended — produces `mind` with version info):

```bash
git clone https://github.com/jf-ferraz/mind-cli.git
cd mind-cli
make build        # builds ./mind with version info
make install      # installs to $GOPATH/bin
```

### Verify

```bash
mind version
```

## Commands

### Project Setup

```bash
mind init [--name NAME]               # Initialize a new Mind project
mind init --from-existing             # Initialize in existing project (preserves docs)
mind init --with-github               # Also create .github/agents/ adapter
```

### Project Health

```bash
mind status                           # Project health and documentation status
mind doctor [--fix]                   # Full diagnostics with optional auto-fix
mind brief                            # Project brief status and gate result
```

### Validation

```bash
mind check docs [--strict]            # 17-check documentation validation suite
mind check refs                       # 11-check cross-reference validation suite
mind check config                     # mind.toml schema validation
mind check all [--strict]             # Run all validation suites
```

### Documentation

```bash
mind docs list [--zone ZONE]          # List documents by zone
mind docs tree                        # Tree view with stub annotations
mind docs stubs                       # List stub (incomplete) documents
mind docs search "query"              # Full-text search across documents
mind docs open <path-or-id>           # Open document in $EDITOR
```

### Scaffolding

```bash
mind create adr "Title"               # Auto-numbered Architecture Decision Record
mind create blueprint "Title"         # Auto-numbered blueprint + INDEX.md update
mind create iteration <type> <name>   # Iteration folder with 5 template files
mind create spike "Title"             # Spike report in knowledge/
mind create convergence "Title"       # Convergence analysis template in knowledge/
mind create brief                     # Interactive project brief creation
```

Types for `create iteration`: `new`, `enhancement`, `bugfix`, `refactor`.

### Workflow & Iterations

```bash
mind workflow status                  # Current workflow state
mind workflow history                 # List all iterations chronologically
mind iterations                       # List all iterations (alias: iter, iters)
mind preflight "request description"  # Pre-flight checks + iteration/branch setup
mind preflight --resume               # Check for in-progress workflow
mind handoff <iteration-id>           # Post-workflow validation + state update
```

### Reconciliation

```bash
mind reconcile                        # Hash documents, detect changes, propagate staleness
mind reconcile --check                # Read-only verification (exit 4 if stale)
mind reconcile --force                # Re-hash everything, clear staleness
mind reconcile --graph                # Show ASCII dependency graph
```

### Framework Lifecycle

```bash
mind framework install --source PATH  # Install framework to ~/.config/mind/
mind framework install --force        # Overwrite existing installation
mind framework status                 # Show version, mode, source, drift count
mind framework diff                   # Compare project .mind/ vs global framework
mind framework materialize            # Populate project .mind/ from global (preserves overrides)
mind framework update                 # Re-materialize only changed artifacts
```

### Global Configuration

```bash
mind config show                      # Display ~/.config/mind/config.toml
mind config edit                      # Open config in $EDITOR
mind config path                      # Print config file path
mind config validate                  # Validate config schema
```

### Project Registry

```bash
mind registry list                    # List all registered projects
mind registry add <alias> <path>      # Register a project
mind registry remove <alias>          # Remove a project
mind registry resolve <@alias>        # Resolve @alias to absolute path
mind registry check                   # Validate all registered paths exist
```

### Other

```bash
mind tui                              # Interactive 5-tab TUI dashboard
mind serve                            # Start MCP server (JSON-RPC 2.0, stdio)
mind version [--short]                # Print version and build info
mind completion bash|zsh|fish         # Generate shell completions
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--no-color` | | Disable colored output |
| `--project-root` | `-p` | Path to project root (default: auto-detect) |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Validation failure or issues found |
| 2 | Runtime error |
| 3 | Configuration error (not a Mind project, schema violation) |
| 4 | Staleness detected (`reconcile --check`) |

## Framework Architecture

Mind has two components:

- **`mind` CLI** (this repo) — Go binary with 42 commands for project management
- **`.mind/` directory** ([.mind repo](https://github.com/jf-ferraz/.mind)) — Markdown agents, conventions, and skills for AI-assisted workflows

The CLI installs framework artifacts globally (`~/.config/mind/`) and materializes them into each project's `.mind/` directory. Projects can override any artifact locally.

```
~/.config/mind/              ← Global framework (shared across projects)
├── agents/                  ← Agent definitions (analyst, architect, etc.)
├── commands/                ← Slash commands (/workflow, /discover, etc.)
├── conventions/             ← Code quality rules
├── conversation/            ← Dialectical analysis system
├── skills/                  ← Deep-dive guides (debugging, planning, etc.)
├── framework.lock           ← Version + SHA-256 checksums
└── projects.toml            ← Project registry

~/project/.mind/             ← Project-local copy (can override globals)
~/project/mind.toml          ← Project manifest
~/project/.claude/CLAUDE.md  ← Claude Code adapter
```

## Development

### Run Tests

```bash
make test          # or: go test ./...
```

### Architecture

4-layer architecture with downward-only dependency flow:

```
Presentation     cmd/ + internal/render/      CLI handlers, output formatting
Service          internal/service/             Business logic orchestration
Domain           domain/                       Pure types, enums, rules (zero imports)
Infrastructure   internal/repo/                Repository interfaces + fs/mem impls
```

Each layer depends only on the layers below it. The domain layer has zero external imports (enforced by `domain/purity_test.go`).

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/pelletier/go-toml/v2` | TOML parsing |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/bubbles` | TUI components |
| `github.com/charmbracelet/glamour` | Markdown rendering |
| `golang.org/x/term` | TTY detection |

## License

MIT — see [LICENSE](LICENSE).
