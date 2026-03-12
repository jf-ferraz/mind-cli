# mind-cli

Single-binary Go CLI for the Mind Agent Framework. Provides project health diagnostics, documentation validation, document scaffolding, and workflow state inspection.

## Getting Started

### Prerequisites

- Go 1.23+

### Build

```bash
go build -o mind .
```

With version info:

```bash
go build -ldflags "-X github.com/jf-ferraz/mind-cli/cmd.Version=0.1.0 -X github.com/jf-ferraz/mind-cli/cmd.CommitSHA=$(git rev-parse --short HEAD) -X github.com/jf-ferraz/mind-cli/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o mind .
```

### Usage

```bash
mind init                          # Initialize a new project
mind status                        # Project health overview
mind doctor [--fix]                # Full diagnostics with optional auto-fix
mind check docs [--strict]         # 17-check documentation validation
mind check refs                    # 11-check cross-reference validation
mind check config                  # mind.toml schema validation
mind check all                     # Run all validation suites
mind docs list [--zone spec]       # List documents by zone
mind docs tree                     # Tree view with stub annotations
mind docs stubs                    # List stub documents
mind docs search "query"           # Search document content
mind docs open <path-or-id>        # Open in $EDITOR
mind create adr "Title"            # Auto-numbered ADR
mind create blueprint "Title"      # Auto-numbered blueprint + INDEX.md update
mind create iteration <type> <name># Iteration folder with 5 template files
mind create spike "Title"          # Spike report in knowledge/
mind create convergence "Title"    # Convergence analysis in knowledge/
mind create brief                  # Interactive project brief
mind brief                         # Brief status and gate result
mind iterations                    # List all iterations
mind workflow status               # Current workflow state
mind workflow history              # Iteration history
mind version [--short]             # Build information
```

Global flags: `--json`, `--no-color`, `--project <path>`

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Validation failure or issues found |
| 2 | Runtime error |
| 3 | Not a Mind project (no `.mind/` directory) |

## Documentation

Project documentation lives in `docs/` using the 5-zone structure:

- [`docs/spec/`](docs/spec/) — Specifications (requirements, architecture, domain model)
- [`docs/blueprints/`](docs/blueprints/) — System-level planning artifacts
- [`docs/state/`](docs/state/) — Active state and workflow tracking
- [`docs/iterations/`](docs/iterations/) — Per-change tracking and history
- [`docs/knowledge/`](docs/knowledge/) — Domain reference (glossary, integrations)

## Development

### Run Tests

```bash
go test ./...
```

### Architecture

4-layer architecture with downward-only dependency flow:

```
Presentation     cmd/ + internal/render/      CLI handlers, output formatting
Service          internal/service/             Business logic orchestration
Domain           domain/                       Pure types, enums, rules (zero external imports)
Infrastructure   internal/repo/                Repository interfaces + fs/mem implementations
```

Each layer depends only on the layers below it. The domain layer has zero external imports (enforced by `domain/purity_test.go`). All filesystem access passes through repository interfaces, enabling in-memory implementations for testing.

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/pelletier/go-toml/v2` | mind.toml parsing |
| `golang.org/x/term` | TTY detection, terminal width |

## License

MIT
