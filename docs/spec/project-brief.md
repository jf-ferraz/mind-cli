# Project Brief

## Vision

A single-binary CLI and TUI tool that provides unified project intelligence for the Mind Agent Framework. It replaces 9 standalone bash scripts with a coherent interface, adds project health diagnostics, and bridges deterministic tooling with AI agent workflows via MCP server, watch mode, and full orchestration.

## Target Users

- **Framework users**: Developers who install the Mind Agent Framework into their projects and need project health visibility, document scaffolding, and validation without memorizing 9 script names.
- **AI workflow users**: Developers using Claude Code or Copilot Chat with the framework who need pre-flight checks, post-workflow cleanup, and real-time feedback during AI workflows.

## Problem Statement

The Mind Agent Framework has 9 bash scripts with inconsistent interfaces and no discovery mechanism. Users must read documentation to know the scripts exist. There is no project-level intelligence — no way to quickly answer "is my documentation complete?", "what's the workflow state?", or "what should I do next?". The gap between the deterministic CLI world and AI agent workflows requires manual context switching with no coordination.

## Key Deliverables

1. `mind` CLI binary — unified command interface for all framework operations
2. Interactive TUI dashboard — 5-tab interface showing project health, documents, iterations, validation, and quality trends
3. MCP server (`mind serve`) — exposes project intelligence as tools for AI agents
4. Watch mode (`mind watch`) — real-time filesystem monitoring with automatic validation
5. Full orchestration (`mind run`) — drives complete AI workflows from the terminal

## Success Metrics

- All 17 docs validation checks pass when ported from bash to Go
- All 11 cross-reference checks pass when ported from bash to Go
- CLI startup time under 50ms
- Single binary under 15MB
- `mind status` renders a complete dashboard in under 200ms

## Scope

### In Scope

- CLI commands: status, doctor, init, create, docs, check, workflow, quality, sync, tui, serve, preflight, run, watch
- TUI with 5 tabs: status, documents, iterations, checks, quality
- MCP server with 16 tools for AI agent integration
- Document generation (ADR, blueprint, iteration, spike, convergence, brief)
- Validation engine porting all bash script checks to native Go
- Watch mode with filesystem monitoring and automatic gate checks
- Full orchestration dispatching agents via `claude` CLI
- Shell completions for bash, zsh, fish
- Cross-platform builds (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)

### Out of Scope

- AI model calls from the CLI itself (except via `claude` CLI in orchestration mode)
- Claude Code or Copilot Chat plugin development
- Modifying the Mind Agent Framework itself (agents, skills, conventions)
- GUI or web interface
- Multi-project management (operates on one project at a time)

## Constraints

- **Language**: Go 1.23+
- **Dependencies**: Cobra (CLI), Bubble Tea + Lip Gloss (TUI), fsnotify (watch)
- **Distribution**: Single binary via `go install`, GitHub releases, AUR
- **Compatibility**: Must produce identical validation results to existing bash scripts

## Technical Preferences

- Go with Charm ecosystem (Bubble Tea, Lip Gloss, Bubbles, Glamour)
- Cobra for CLI argument parsing
- 4-layer architecture: Presentation → Service → Domain → Infrastructure
- Repository pattern for all filesystem access (enables in-memory testing)
- MVU (Elm architecture) for TUI state management

## Core Domain Concepts

### Entities

- **Project**: A directory containing `.mind/` with framework files
- **Document**: A markdown file in one of the 5 documentation zones
- **Zone**: One of spec, blueprints, state, iterations, knowledge
- **Iteration**: A per-change tracking folder under docs/iterations/
- **WorkflowState**: Persisted state of an in-progress AI workflow
- **ValidationReport**: Results from running a validation suite
- **QualityScore**: Convergence analysis quality assessment (6 dimensions)
- **Brief**: The project brief with business context gate classification
- **AgentChain**: Sequence of agents for a given request type

### Business Rules

- A project must have `.mind/` to be valid
- Documents with only headings, comments, and placeholders are classified as stubs
- The business context gate blocks NEW_PROJECT/COMPLEX_NEW when the brief is missing or a stub
- Quality gate 0 requires convergence score >= 3.0/5.0
- Maximum 2 retry loops per workflow

## Open Questions

- Should the MCP server return individual check results or aggregate reports?
- Should cost tracking be stored per-iteration or in a separate log?
- Should `mind run` support parallel agent dispatch for independent chain steps?
