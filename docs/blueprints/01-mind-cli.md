# Blueprint: `mind` CLI & TUI

> A unified command-line tool that replaces 9 standalone scripts, adds project intelligence, and provides an interactive TUI dashboard.

**Status**: Proposal
**Date**: 2026-03-09

---

## Problem

The framework currently has 9 bash scripts with inconsistent interfaces, no discovery mechanism (users must read docs to know they exist), and no project-level intelligence. A user cloning a project with `.mind/` installed has no quick way to answer:

- Is my documentation complete?
- What's the current workflow state?
- What should I do next?
- What issues exist in my project?

Scripts solve individual tasks but don't compose into a coherent developer experience.

---

## Language Recommendation: Go

### Why Go

| Criterion | Go | Rust | Python | Bash |
|-----------|----|----|--------|------|
| **Single binary** | Yes | Yes | No (runtime) | No (interpreter) |
| **TUI ecosystem** | Bubble Tea (best-in-class) | ratatui (excellent) | Textual (good) | dialog/whiptail (basic) |
| **CLI framework** | Cobra (industry standard) | clap (excellent) | typer/click (excellent) | getopt (fragile) |
| **Build time** | ~3s | ~30s+ | N/A | N/A |
| **Cross-platform** | Trivial | Trivial | Needs venv | POSIX only |
| **Learning curve** | Low | Medium-high | Low | N/A |
| **Startup time** | ~5ms | ~2ms | ~200ms | ~50ms |
| **Distribution** | `go install` + releases | releases | pip/pipx | copy |

**Go wins because:**

1. **Charm ecosystem** — [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles) produce the best terminal UIs in any language. Used by `gh`, `glow`, `soft-serve`.
2. **Cobra** — Powers `kubectl`, `hugo`, `gh`. Auto-generates shell completions, man pages, and help text.
3. **Single binary distribution** — `go install github.com/jf-ferraz/mind-cli@latest`, AUR package, or download from GitHub releases. No runtime dependencies in the target project.
4. **Fast builds** — Iterate quickly during development.
5. **Framework-agnostic** — The CLI shouldn't be tied to the project's language. Go ensures the tool works regardless of whether the project is Rust, Node, Python, or anything else.

**Rust is the runner-up** — ratatui is excellent and the user already knows Rust. The trade-off is slower iteration during development (compile times) and more boilerplate for CLI argument handling. If the maintainer prefers Rust, the architecture below translates directly — swap Cobra→clap, Bubble Tea→ratatui.

### Key Dependencies

```
github.com/spf13/cobra           # CLI framework
github.com/charmbracelet/bubbletea  # TUI framework
github.com/charmbracelet/lipgloss   # TUI styling
github.com/charmbracelet/bubbles    # TUI components (tables, spinners, text input)
github.com/charmbracelet/glamour    # Markdown rendering in terminal
gopkg.in/yaml.v3                    # YAML parsing (for conversation config)
github.com/pelletier/go-toml/v2     # TOML parsing (for mind.toml)
```

---

## Command Tree

### Design Principles

1. **Verb-noun structure** — `mind <verb> [noun]` (create, check, show, list)
2. **Smart defaults** — `mind status` needs zero arguments. Most commands auto-detect project root.
3. **Progressive disclosure** — Simple commands first, flags for power users
4. **Script compatibility** — Wraps existing bash scripts initially, replaces logic incrementally
5. **No AI dependency** — The CLI is deterministic. AI workflows stay in Claude Code / Copilot Chat.

### Full Command Tree

```
mind
│
├── status                              Project health dashboard (default TUI if terminal is interactive)
│
├── init [--name NAME] [--with-github]  Initialize framework in current directory
│   └── --from-existing                 Detect existing docs and preserve them
│
├── doctor                              Deep diagnostics with actionable fix suggestions
│   ├── --fix                           Auto-fix what can be fixed (create missing dirs, fix naming)
│   └── --json                          Machine-readable output
│
├── create                              Create framework artifacts
│   ├── adr "<title>"                   Architecture Decision Record (auto-numbered)
│   ├── blueprint "<title>"             Blueprint + INDEX.md update (auto-numbered)
│   ├── iteration <type> "<name>"       Iteration folder with 5 files (auto-numbered)
│   │   └── types: new, enhancement, bugfix, refactor
│   ├── spike "<title>"                 Technical spike report
│   ├── convergence "<title>"           Convergence analysis template
│   └── brief                           Interactive project brief creation (guided prompts)
│
├── docs                                Document management
│   ├── list [--zone ZONE]              List documents grouped by zone
│   │   └── zones: spec, blueprints, state, iterations, knowledge
│   ├── tree                            Visual tree of all documentation
│   ├── open <path-or-id>               Open document in $EDITOR
│   ├── stubs                           List all stub documents that need content
│   └── search "<query>"                Full-text search across docs/
│
├── check                               Validation suite
│   ├── docs [--strict]                 5-zone documentation (17 checks)
│   ├── refs                            Framework cross-references (11 checks)
│   ├── config                          Conversation YAML configs
│   ├── convergence <file>              Convergence output validation (23 checks)
│   ├── all [--strict]                  Run everything, unified report
│   └── --json                          Machine-readable output (all subcommands)
│
├── workflow                            Workflow state management
│   ├── status                          Current workflow state (from docs/state/workflow.md)
│   ├── history                         Past iterations chronologically
│   ├── show <iteration-id>             Show iteration details (overview + validation)
│   └── clean [--dry-run]               Remove stale workflow state
│
├── sync                                Platform synchronization
│   ├── agents [--check]                Sync .mind/conversation/agents/ → .github/agents/
│   └── status                          Show sync state (diffs between platforms)
│
├── quality                             Quality tracking
│   ├── log <convergence-file>          Extract scores and append to quality-log.yml
│   │   └── --topic, --variant          Override topic name and variant
│   ├── history                         Show quality score trends
│   └── report                          Summary report of all convergence analyses
│
├── tui                                 Launch interactive TUI (full-screen)
│
├── completion <shell>                  Generate shell completions (bash, zsh, fish)
│
├── help [command]                      Help for any command
│
└── version [--short]                   Version and build info
```

### Command Details

#### `mind status` — The Entry Point

The most important command. Shows project health at a glance:

```
╭─ Mind Framework ─────────────────────────────────────────────────────────╮
│                                                                          │
│  Project: iron-arch-v2              Framework: v2026-03-09               │
│  Root: ~/.config/iron/              Branch: feature/add-caching          │
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

Non-interactive (piped/CI): outputs plain text. Interactive terminal: styled with Lip Gloss.

#### `mind doctor` — Diagnostics

Goes deeper than `status`. Runs all validators, cross-references results, and produces actionable suggestions:

```
$ mind doctor

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

With `--fix`: creates missing directories, adds `.gitkeep` files, creates stub documents from templates, fixes naming conventions.

#### `mind create brief` — Guided Brief Creation

Interactive terminal prompts (using Bubble Tea text input):

```
$ mind create brief

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

#### `mind check all` — Unified Validation

```
$ mind check all --strict

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

---

## TUI Design

### Main Dashboard (`mind tui`)

Full-screen interactive interface with tab navigation:

```
╭─ Mind Framework ─── iron-arch-v2 ── main ────────────────────────── v2026-03-09 ─╮
│                                                                                    │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]       q:quit ?:help │
│ ─────────────────────────────────────────────────────────────────────────────────── │
│                                                                                    │
│  Documentation Health                    │  Active Workflow                         │
│  ───────────────────                     │  ───────────────                         │
│  spec/        ████████░░  4/5            │  State: idle                             │
│  blueprints/  ██████████  3/3            │  Last: 006-ENH-add-caching               │
│  state/       █████░░░░░  1/2            │  Completed: 2026-03-08                   │
│  iterations/  ██████████  6/6            │                                          │
│  knowledge/   ██████░░░░  3/5            │  Quick Actions                           │
│                                          │  ──────────                               │
│  Warnings (2)                            │  c  Create document                      │
│  ────────────                            │  d  Run doctor                           │
│  ⚠ domain-model.md is a stub            │  v  Validate all                         │
│  ⚠ glossary.md missing                  │  o  Open document                        │
│                                          │  s  Sync agents                          │
│                                          │                                          │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

### Tab 2: Documents Browser

```
╭─ Documents ──────────────────────────────────────────────────────────────────────╮
│                                                                                    │
│  Zone: [all] spec  blueprints  state  iterations  knowledge        /: search       │
│ ─────────────────────────────────────────────────────────────────────────────────── │
│                                                                                    │
│  spec/                                                                             │
│  ├── project-brief.md          ✓ complete    2026-03-01     3.2 KB                │
│  ├── requirements.md           ✓ complete    2026-03-05     8.1 KB                │
│  ├── architecture.md           ✓ complete    2026-03-05     6.4 KB                │
│  ├── domain-model.md           ⚠ stub        2026-02-28     0.4 KB                │
│  └── decisions/                                                                    │
│      ├── 001-use-postgresql.md ✓ complete    2026-03-02     1.8 KB                │
│      └── 002-jwt-auth.md       ✓ complete    2026-03-04     2.1 KB                │
│                                                                                    │
│  blueprints/                                                                       │
│  ├── INDEX.md                  ✓ complete    2026-03-05     0.9 KB                │
│  ├── 01-api-design.md          ✓ complete    2026-03-03     4.2 KB                │
│  └── 02-auth-system.md         ✓ complete    2026-03-04     3.7 KB                │
│                                                                                    │
│  ↑↓: navigate   enter: preview   e: edit   n: new   d: delete          1/18 files │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

### Tab 3: Iterations Timeline

```
╭─ Iterations ─────────────────────────────────────────────────────────────────────╮
│                                                                                    │
│  Filter: [all] NEW_PROJECT  ENHANCEMENT  BUG_FIX  REFACTOR         /: search       │
│ ─────────────────────────────────────────────────────────────────────────────────── │
│                                                                                    │
│  #   Type          Name                    Status      Date         Files           │
│  ─── ───────────── ─────────────────────── ─────────── ──────────── ─────           │
│  006 ENHANCEMENT   add-caching             ✓ complete  2026-03-08   5/5             │
│  005 BUG_FIX       fix-auth-redirect       ✓ complete  2026-03-07   5/5             │
│  004 REFACTOR      extract-repositories    ✓ complete  2026-03-06   4/5             │
│  003 ENHANCEMENT   websocket-notifications ✓ complete  2026-03-05   5/5             │
│  002 ENHANCEMENT   role-based-access       ✓ complete  2026-03-03   5/5             │
│  001 NEW_PROJECT   initial-api             ✓ complete  2026-03-01   5/5             │
│                                                                                    │
│                                                                                    │
│  ↑↓: navigate   enter: details   o: open overview   v: open validation             │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

### Tab 4: Validation Results

Runs all validators and displays results interactively. Expandable sections per check.

### Tab 5: Quality Trends

Displays convergence quality scores over time (if quality-log.yml exists):

```
╭─ Quality Trends ─────────────────────────────────────────────────────────────────╮
│                                                                                    │
│  Overall Score History                                                             │
│                                                                                    │
│  4.0 ┤                                              ╭──●                          │
│  3.5 ┤                              ╭──●────●───────╯                             │
│  3.0 ┤─ ─ ─ ─ ─ ─ ─●──────●───────╯─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  Gate 0 threshold   │
│  2.5 ┤          ╭───╯                                                              │
│  2.0 ┤     ●───╯                                                                   │
│      └────────────────────────────────────────────────────────────────              │
│       Feb 25   Feb 27   Mar 01   Mar 03   Mar 05   Mar 07   Mar 09                │
│                                                                                    │
│  Latest: auth-strategy-convergence.md  Score: 3.8/5.0  Gate 0: PASS               │
╰────────────────────────────────────────────────────────────────────────────────────╯
```

---

## Architecture

### Project Structure

```
mind-cli/
├── cmd/                        Cobra command definitions
│   ├── root.go                 Root command, global flags
│   ├── status.go               mind status
│   ├── doctor.go               mind doctor [--fix] [--json]
│   ├── init.go                 mind init
│   ├── create.go               mind create <type>
│   ├── docs.go                 mind docs {list,tree,open,stubs,search}
│   ├── check.go                mind check {docs,refs,config,convergence,all}
│   ├── workflow.go             mind workflow {status,history,show,clean}
│   ├── sync.go                 mind sync {agents}
│   ├── quality.go              mind quality {log,history,report}
│   ├── tui.go                  mind tui (launches Bubble Tea app)
│   ├── completion.go           mind completion <shell>
│   └── version.go              mind version
│
├── internal/
│   ├── project/                Project detection and context
│   │   ├── detect.go           Find .mind/ root, read mind.toml
│   │   ├── config.go           Parse mind.toml manifest
│   │   └── state.go            Read/write workflow state
│   │
│   ├── docs/                   Document management
│   │   ├── zones.go            5-zone enumeration and file listing
│   │   ├── stub.go             Stub detection (port of is_stub)
│   │   ├── brief.go            Project brief parsing and validation
│   │   ├── iteration.go        Iteration parsing and listing
│   │   └── search.go           Full-text search across docs/
│   │
│   ├── validate/               Validation engine
│   │   ├── docs.go             17-check docs validator (port of validate-docs.sh)
│   │   ├── refs.go             11-check cross-reference validator
│   │   ├── config.go           YAML config validator
│   │   ├── convergence.go      23-check convergence validator
│   │   └── report.go           Unified report builder
│   │
│   ├── generate/               Document generation
│   │   ├── template.go         Template loading and substitution
│   │   ├── sequence.go         Auto-numbering (port of next_seq)
│   │   ├── slugify.go          Title → slug conversion
│   │   └── types.go            Document type definitions
│   │
│   ├── quality/                Quality tracking
│   │   ├── extract.go          Score extraction from convergence files
│   │   ├── log.go              quality-log.yml read/write
│   │   └── trends.go           Score trend analysis
│   │
│   └── sync/                   Platform synchronization
│       ├── agents.go           Agent body sync logic
│       └── diff.go             Content diff detection
│
├── tui/                        Bubble Tea TUI application
│   ├── app.go                  Main TUI model (tab switching)
│   ├── status.go               Status tab
│   ├── docs.go                 Document browser tab
│   ├── iterations.go           Iteration timeline tab
│   ├── checks.go               Validation results tab
│   ├── quality.go              Quality trends tab
│   ├── styles.go               Lip Gloss style definitions
│   └── keys.go                 Key bindings
│
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── README.md
```

### Design Decisions

**1. Port scripts to native Go, don't shell out**

Initially, the CLI could wrap bash scripts (`exec.Command("bash", "scripts/validate-docs.sh")`). But this creates a runtime dependency on the scripts being present and adds startup overhead. Instead, port the validation logic to Go directly. The scripts are 100-700 lines each — straightforward to translate.

Keep the bash scripts for backward compatibility. The CLI becomes the primary interface; scripts remain for CI/CD and users who prefer them.

**2. Project root detection**

Walk up from the current directory looking for `.mind/` (similar to how `git` finds `.git/`). Cache the root path for the session. All commands operate relative to this root.

```go
func FindProjectRoot() (string, error) {
    dir, _ := os.Getwd()
    for {
        if _, err := os.Stat(filepath.Join(dir, ".mind")); err == nil {
            return dir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            return "", errors.New("not a Mind project (no .mind/ found)")
        }
        dir = parent
    }
}
```

**3. Output modes**

Every command supports three output modes:
- **Interactive** (default when TTY detected): Styled with Lip Gloss colors, progress bars, tables
- **Plain** (piped or `--no-color`): Clean text, no ANSI codes
- **JSON** (`--json`): Machine-readable for scripting and CI

**4. No AI in the CLI**

The CLI is deterministic. It manages files, validates structure, and displays state. AI workflows (`/workflow`, `/discover`, `/analyze`) stay in Claude Code / Copilot Chat. The CLI is the companion tool — it handles everything around the AI workflows.

The only exception is `mind create brief`, which uses interactive terminal prompts (not AI) to guide the user through filling out a project brief.

---

## Distribution

### Install Methods

```bash
# Go install (requires Go toolchain)
go install github.com/jf-ferraz/mind-cli@latest

# Homebrew (macOS/Linux)
brew install jf-ferraz/tap/mind

# AUR (Arch Linux)
yay -S mind-cli

# Download binary from GitHub releases
curl -fsSL https://github.com/jf-ferraz/mind-cli/releases/latest/download/mind-linux-amd64 -o /usr/local/bin/mind
chmod +x /usr/local/bin/mind

# From source
git clone https://github.com/jf-ferraz/mind-cli
cd mind-cli && make install
```

### Release Automation

Use [GoReleaser](https://goreleaser.com/) for cross-compilation and release packaging:

```yaml
# .goreleaser.yml
builds:
  - binary: mind
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    name_template: "mind-{{ .Os }}-{{ .Arch }}"

aurs:
  - name: mind-cli-bin
    homepage: https://github.com/jf-ferraz/mind-cli
```

---

## Implementation Phases

### Phase 1: Core CLI (1-2 weeks)

Minimal viable CLI that replaces the most-used scripts.

```
mind status         → Project health summary (plain text)
mind init           → Wraps scaffold.sh + install.sh
mind create *       → Wraps docs-gen.sh (adr, blueprint, iteration, spike, convergence)
mind check docs     → Wraps validate-docs.sh
mind check refs     → Wraps validate-integration.sh
mind check config   → Wraps validate-config.sh
mind check all      → Runs all three
mind version
mind help
mind completion
```

**Strategy**: Shell out to existing bash scripts in Phase 1. This ships fast and proves the command structure. Port to native Go in Phase 2.

### Phase 2: Native Validators + Doctor (1-2 weeks)

Port bash validation logic to Go. Add project intelligence.

```
mind doctor          → Deep diagnostics with fix suggestions
mind doctor --fix    → Auto-fix resolvable issues
mind docs list       → Native document listing with zone grouping
mind docs stubs      → List all stub documents
mind docs tree       → Visual tree view
mind workflow status → Parse docs/state/workflow.md
mind workflow history → List iterations chronologically
mind check --json    → Machine-readable output for CI
```

### Phase 3: Interactive TUI (1-2 weeks)

Full Bubble Tea TUI with tab navigation.

```
mind tui             → Full-screen dashboard
  Tab 1: Status      → Health bars, warnings, workflow state
  Tab 2: Documents   → Browseable file tree with stub detection
  Tab 3: Iterations  → Timeline with type filtering
  Tab 4: Validation  → Live validation results
```

### Phase 4: Advanced Features (1-2 weeks)

```
mind create brief     → Interactive project brief creation with prompts
mind docs search      → Full-text search
mind docs open        → Open in $EDITOR with fuzzy finding
mind quality log      → Native score extraction
mind quality history  → Score trends (ASCII chart)
mind sync agents      → Native agent sync
mind check convergence → Native convergence validation
```

### Phase 5: Polish (1 week)

- Shell completions for bash, zsh, fish
- Man page generation
- AUR package
- GoReleaser CI/CD
- Integration tests

---

## Commands Critique: Original vs. Recommended

The user's original examples and why they were adjusted:

| Original | Issue | Recommended |
|----------|-------|-------------|
| `mind create docs` | "docs" is too vague — create what? | `mind create <type>` with explicit types: `adr`, `blueprint`, `iteration`, `spike`, `convergence`, `brief` |
| `mind create docs --product-requirements` | Requirements are generated by the analyst agent, not manually created | `mind create brief` (the brief is what humans fill; requirements come from AI) |
| `mind create docs --project-debrief` | "Debrief" isn't a framework concept | `mind create brief` (project brief is the correct term) |
| `mind create docs --blueprint architecture` | Reversed verb-noun | `mind create blueprint "Architecture Design"` |
| `mind status` | Good as-is | `mind status` — kept, made the hero command |
| `mind help` | Standard | `mind help [command]` — Cobra provides this automatically |
| `mind project --status` | Redundant with `mind status` | Dropped — `mind status` covers it |

### Additional Commands Not in Original Request

| Command | Why It's Valuable |
|---------|-------------------|
| `mind doctor` | Goes beyond status — diagnoses root causes and suggests fixes |
| `mind doctor --fix` | Reduces friction — auto-creates missing dirs and stubs |
| `mind docs stubs` | Quick answer to "what needs attention?" |
| `mind docs search` | Find content without knowing where it lives |
| `mind check all` | One command for CI gates |
| `mind check --json` | Enables CI/CD integration |
| `mind workflow history` | "What happened on this project?" at a glance |
| `mind quality history` | Track convergence quality improvement over time |
| `mind tui` | The full experience for interactive exploration |
| `mind completion` | Shell completions for productivity |

---

## What This Tool Does NOT Do

1. **Does not replace AI workflows** — `/workflow`, `/discover`, `/analyze` remain Claude Code slash commands. The CLI manages the surrounding infrastructure.
2. **Does not call AI APIs** — Zero AI dependency. Every command is deterministic.
3. **Does not manage Claude Code or Copilot settings** — Platform configuration stays in `.claude/` and `.github/`.
4. **Does not modify agent definitions** — Agents are markdown files edited by hand or by AI. The CLI only reads them for validation and sync.

---

## Repository Structure

```
github.com/jf-ferraz/mind-cli       # The CLI tool (Go)
github.com/jf-ferraz/.mind          # The framework (markdown + scripts)
```

The CLI is a separate repository. It reads `.mind/` from whatever project it's run in, but doesn't ship with the framework. Users install the framework (`install.sh`) and the CLI (`go install`) independently.

---

## Alternative: Rust Implementation

If Rust is preferred over Go, the architecture maps directly:

| Go | Rust |
|----|------|
| Cobra | clap (derive) |
| Bubble Tea | ratatui + crossterm |
| Lip Gloss | ratatui::style |
| Bubbles | tui-rs widgets |
| Glamour | termimad |
| gopkg.in/yaml.v3 | serde_yaml |
| go-toml | toml (serde) |
| GoReleaser | cargo-dist or cross |

The Rust equivalent would produce a smaller binary (~3MB vs ~8MB) with faster startup, at the cost of longer compile times and more boilerplate for the TUI event loop.

---

> **See also:**
> - `README.md` — Framework overview
> - `docs/reference/scripts.md` — Current script reference (what the CLI replaces)
> - `docs/guides/use-cases.md` — Workflow scenarios the CLI supports
