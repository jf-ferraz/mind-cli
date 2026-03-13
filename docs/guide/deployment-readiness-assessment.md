# Deployment Readiness Assessment — mind-cli

> **Date**: 2026-03-13 (revised)
> **Assessor**: Automated (assessment-only, no code changes)
> **CLI Version**: v0.3.0 — built from `main` @ `c8cd0ff`
> **Go Version**: go1.26.1
> **Tag**: `v0.3.0` pushed to `origin`

---

## Executive Summary

**Overall Verdict: PASS — 0 blockers, 2 caveats, 1 recommendation**

The CLI builds, all 19 test packages pass, and all commands execute without crashing. The v0.3.0 tag is pushed to GitHub and `go install github.com/jf-ferraz/mind-cli@v0.3.0` works. Global commands (`config`, `registry`, `framework install/status`) work correctly outside a project directory. Two caveats remain: the `go install` binary is named `mind-cli` (not `mind`), and version injection requires the Makefile. Both are documented.

---

## 1 — Build & Distribution Readiness

| Check | Verdict | Evidence |
|-------|---------|----------|
| `go build -o mind .` | **PASS** | Builds cleanly via `make build` |
| `go install github.com/jf-ferraz/mind-cli@v0.3.0` | **PASS** | Installs successfully; binary named `mind-cli` (documented caveat) |
| `go test ./...` | **PASS** | 19/19 packages pass (69 test files) |
| Version injection via ldflags | **PASS** | `make build` and `make install` inject version, commit SHA, and build date |
| Makefile | **PASS** | Targets: `build`, `install`, `test`, `vet`, `clean`, `help` |

### Binary Name Caveat

`go install github.com/jf-ferraz/mind-cli@v0.3.0` produces `~/go/bin/mind-cli`, but the CLI self-identifies as `mind` and all documentation refers to `mind`. The README documents both install paths:

1. **`make install`** — produces `mind` in `$GOPATH/bin` with version injection
2. **`go install @v0.3.0`** — produces `mind-cli`; tester can symlink: `ln -s ~/go/bin/mind-cli ~/go/bin/mind`

---

## 2 — Repository Readiness

| Check | Verdict | Evidence |
|-------|---------|----------|
| GitHub remote configured | **PASS** | `origin → https://github.com/jf-ferraz/mind-cli.git` |
| Repo publicly accessible | **PASS** | `git ls-remote` succeeds without auth |
| `mind` (framework) repo accessible | **PASS** | `https://github.com/jf-ferraz/mind.git` also public |
| Main branch up to date | **PASS** | `origin/main` matches local `main` at `c8cd0ff` |
| v0.3.0 tag pushed | **PASS** | `go install @v0.3.0` resolves correctly |
| `.gitignore` coverage | **PASS** | `/mind` (build output), `.mind/`, `.claude/`, `archive/`, `prompt.txt` all covered |
| No binary in working tree | **PASS** | Removed via `make clean` |
| Go dependencies | **PASS** | Not vendored; `go.sum` present; `go install @v0.3.0` fetches cleanly |
| Docs tracked | **PASS** | `docs/guide/end-to-end-walkthrough.md` and `deployment-readiness-assessment.md` both committed |
| LICENSE | **PASS** | MIT license file present and tracked |
| Stale local branches | **NOTE** | 16 local branches from earlier phases — no impact on testers or contributors (remote only has `main`) |

---

## 3 — Installation Documentation

| Check | Verdict | Evidence |
|-------|---------|----------|
| README.md install section | **PASS** | Quick Start for Testers (7 steps), build from source, `go install` |
| End-to-end walkthrough | **PASS** | `docs/guide/end-to-end-walkthrough.md` — 1,538 lines, 17 sections |
| Walkthrough: install flow | **PASS** | Three install options documented with clear trade-offs |
| Walkthrough: global setup | **PASS** | Clone mind repo → `framework install --source` → verify |
| Prerequisites | **PASS** | README and walkthrough both state Go 1.24+ (matches `go.mod` requirement of 1.24.2) |
| Platform support | **NOTE** | Linux/macOS/WSL documented. No Windows native guidance. |
| Shell completion | **PASS** | Documented for bash/zsh/fish |

---

## 4 — Global Configuration Bootstrap

| Check | Verdict | Evidence |
|-------|---------|----------|
| `framework install --source <path>` | **PASS** | Works correctly from any directory; copies 145 artifacts to `~/.config/mind/` |
| `config show` without project | **PASS** | Works from any directory |
| `registry list` without project | **PASS** | Works from any directory |
| `framework status` without project | **PASS** | Works from any directory |
| Framework requires second repo clone | **NOTE** | Tester must clone `github.com/jf-ferraz/mind.git` separately — documented in README Quick Start |
| `config.toml` auto-created | **PASS** | Created on first access with sensible defaults |
| `projects.toml` auto-created | **PASS** | Created when `registry add` is first called |
| `framework.lock` written | **PASS** | SHA-256 checksums, version, source path all recorded |

Global commands correctly exempt from project requirement via `requiresProject()` in `cmd/root.go:124`: `config`, `registry`, `framework install`, and `framework status` all work outside a project directory.

---

## 5 — First-Run Experience

| Check | Verdict | Evidence |
|-------|---------|----------|
| `mind init` on empty dir | **PASS** | Creates `mind.toml`, `.claude/CLAUDE.md`, `docs/` with 5 zones, 8 stub files |
| `mind status` on fresh project | **PASS** | Shows health dashboard, warns about stubs, exits 1 (correct) |
| `mind doctor` on fresh project | **PASS** | 11 pass, 0 fail, 7 warnings — all actionable |
| `mind check all` on fresh project | **PASS** | 40 pass, 0 fail, 1 warning (stubs) |
| Error message for non-project dir | **PASS** | "not a Mind project (no .mind/ directory found)" with exit code 3 |
| `mind framework materialize` | **PASS** | Copies 145 artifacts from global to project `.mind/` |
| `mind framework update` | **PASS** | Detects no changes, reports "already up to date" |
| `mind framework diff` | **PASS** | Shows differences between project and global |
| `create brief` (interactive) | **PASS** | Launches interactive prompt — works, requires TTY |

---

## 6 — Command Verification Matrix

### Command Count

20 top-level commands (including `help` and `completion`) with 30 subcommands across 7 parent commands.

### Full Matrix

| Command | No Project | With Project | Notes |
|---------|-----------|-------------|-------|
| `version` | PASS | PASS | Always works |
| `version --short` | PASS | PASS | Shows version string |
| `help` | PASS | PASS | Always works |
| `completion bash/zsh/fish` | PASS | PASS | Always works |
| `init` | PASS | PASS | Creates project structure |
| `init --with-github` | PASS | PASS | Adds `.github/agents/` |
| `init --from-existing` | PASS | PASS | Preserves existing docs |
| `status` | exit 3 | PASS | |
| `doctor` | exit 3 | PASS | |
| `doctor --fix` | exit 3 | PASS | |
| `brief` | exit 3 | PASS | |
| `check docs` | exit 3 | PASS | 17-check suite |
| `check docs --strict` | exit 3 | PASS | |
| `check refs` | exit 3 | PASS | 11-check suite |
| `check config` | exit 3 | PASS | 12-check suite |
| `check all` | exit 3 | PASS | Aggregates all suites |
| `docs list` | exit 3 | PASS | |
| `docs list --zone spec` | exit 3 | PASS | |
| `docs tree` | exit 3 | PASS | |
| `docs stubs` | exit 3 | PASS | Exits 1 if stubs found (correct) |
| `docs search "query"` | exit 3 | PASS | |
| `docs open <path>` | exit 3 | PASS | Opens `$EDITOR` |
| `create adr "Title"` | exit 3 | PASS | Auto-numbered |
| `create blueprint "Title"` | exit 3 | PASS | Updates INDEX.md |
| `create iteration <type> <name>` | exit 3 | PASS | Types: new, enhancement, bugfix, refactor |
| `create spike "Title"` | exit 3 | PASS | |
| `create convergence "Title"` | exit 3 | PASS | |
| `create brief` | exit 3 | PASS | Interactive TTY |
| `iterations` | exit 3 | PASS | |
| `workflow status` | exit 3 | PASS | |
| `workflow history` | exit 3 | PASS | |
| `preflight "request"` | exit 3 | PASS | Creates iteration + branch |
| `handoff <iter-id>` | exit 3 | PASS | Requires valid iteration ID |
| `reconcile` | exit 3 | PASS | Requires `[documents]` in mind.toml |
| `reconcile --check` | exit 3 | PASS | Exit 4 if stale |
| `reconcile --graph` | exit 3 | PASS | Requires `[[graph]]` in mind.toml |
| `framework install` | PASS | PASS | Global command — works anywhere |
| `framework status` | PASS | PASS | Global command — works anywhere |
| `framework diff` | exit 3 | PASS | Needs project (compares project vs global) |
| `framework materialize` | exit 3 | PASS | Needs project (writes to .mind/) |
| `framework update` | exit 3 | PASS | Needs project (writes to .mind/) |
| `config show` | PASS | PASS | Global command — works anywhere |
| `config path` | PASS | PASS | Global command — works anywhere |
| `config edit` | PASS | PASS | Global command — works anywhere |
| `config validate` | PASS | PASS | Global command — works anywhere |
| `registry list` | PASS | PASS | Global command — works anywhere |
| `registry add` | PASS | PASS | Global command — works anywhere |
| `registry remove` | PASS | PASS | Global command — works anywhere |
| `registry resolve` | PASS | PASS | Global command — works anywhere |
| `registry check` | PASS | PASS | Global command — works anywhere |
| `serve` | PASS | PASS | MCP server (JSON-RPC 2.0 over stdio) |
| `tui` | exit 3 | PASS | BubbleTea dashboard |

### Summary

- **0 commands crash** — all exit gracefully
- **0 commands are unwired** — all registered commands have implementations
- **12 global commands** work without a project directory (config 4, registry 5, framework install/status, serve)

---

## 7 — Caveats and Recommendations

### CAVEATS (known, documented, not blocking)

| # | Issue | Mitigation |
|---|-------|------------|
| C1 | `go install` produces binary named `mind-cli` | README documents both install paths; `make install` produces `mind` |
| C2 | `go install` cannot inject version via ldflags | README documents `make build`/`make install` for version injection |

### RECOMMENDATION

| # | Improvement | Priority |
|---|-------------|----------|
| R1 | Clean up 16 stale local branches | Low — no impact on testers, remote only has `main` |

---

## 8 — Phase C Verification Results (v0.3.0)

All 10 verification steps executed and passed:

| Step | Commands | Result |
|------|----------|--------|
| C1 | `go install @v0.3.0`, `mind-cli version` | **PASS** — binary installs, all commands present |
| C2 | `framework install --source`, `framework status` | **PASS** — v2026.03.1, 145 artifacts |
| C3 | `mkdir && git init && mind init` | **PASS** — mind.toml, docs/, .claude/ created |
| C4 | `framework materialize`, `framework diff` | **PASS** — 145 artifacts, no differences |
| C5 | `doctor`, `status`, `check docs/refs/config/all` | **PASS** — 11/0/7 doctor, 40/0/1 check all |
| C6 | `registry add/list/check/resolve/remove` | **PASS** — all 5 operations |
| C7 | `config show/path/validate` | **PASS** — all 3 operations |
| C8 | `docs list/tree/stubs` | **PASS** — all 3 operations |
| C9 | `create adr`, `create iteration` | **PASS** — auto-numbered, templates generated |
| C10 | `framework update` | **PASS** — up to date |

---

## Appendix: Test Environment

| Item | Value |
|------|-------|
| OS | Linux (Arch) |
| Go | go1.26.1 |
| Shell | zsh |
| mind-cli tag | v0.3.0 (`c8cd0ff`, in sync with origin) |
| mind framework | v2026.03.1 (145 artifacts) |
| Test project | `/tmp/mind-test` (fresh `mind init` + `framework materialize`) |
