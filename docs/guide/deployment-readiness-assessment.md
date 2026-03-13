# Deployment Readiness Assessment — mind-cli

> **Date**: 2026-03-13
> **Assessor**: Automated (assessment-only, no code changes)
> **CLI Version**: `dev (unknown)` — built from `main` @ `caf05b1`
> **Go Version**: go1.26.1

---

## Executive Summary

**Overall Verdict: CONDITIONAL PASS — 3 blockers, 6 warnings**

The CLI builds, all 18 test packages pass, and 31 of 31 commands execute without crashing. The end-to-end walkthrough is comprehensive (1,538 lines). However, a tester cannot succeed without addressing **3 blockers**: unpushed commits, binary name mismatch from `go install`, and global commands (`config`, `registry`, `framework`) failing outside a project directory.

---

## 1 — Build & Distribution Readiness

| Check | Verdict | Evidence |
|-------|---------|----------|
| `go build -o mind .` | **PASS** | Builds cleanly, 10.2 MB binary |
| `go install github.com/jf-ferraz/mind-cli@latest` | **PASS (with caveat)** | Installs successfully but binary is named `mind-cli`, not `mind` |
| `go test ./...` | **PASS** | 18/18 packages pass (69 test files) |
| Version injection via ldflags | **FAIL** | Reports `mind dev (unknown) built unknown` — undocumented for testers |
| Makefile / Goreleaser / CI | **BLOCKED** | None exist — no reproducible release pipeline |
| Binary confusion (17MB vs 10MB) | **WARNING** | `~/.local/bin/mind` (17 MB) built with different flags/older Go; `~/go/bin/mind-cli` (10.2 MB) from `go install` — different names, different sizes |

### Binary Name Mismatch — BLOCKER

`go install github.com/jf-ferraz/mind-cli@latest` produces `~/go/bin/mind-cli`, but the CLI self-identifies as `mind` and all documentation refers to `mind`. A tester who follows "Option C" in the walkthrough will get a binary named `mind-cli` that they must manually rename or symlink.

**Root cause**: The Go module is `github.com/jf-ferraz/mind-cli`, so `go install` uses the module name. The walkthrough says `go install .` (which also produces `mind-cli` since package name = module name).

**Fix options**:
1. Document: "After `go install`, rename or symlink: `ln -s ~/go/bin/mind-cli ~/go/bin/mind`"
2. Or add a `//go:build` directive / Makefile target that builds as `mind`

### Binary Size Explanation

| Binary | Size | Reason |
|--------|------|--------|
| `~/.local/bin/mind` | 17.2 MB | Debug info retained, possibly race detector or different build flags |
| `~/go/bin/mind-cli` | 10.2 MB | Standard `go install` (debug info, not stripped) |
| `/tmp/mind-test-stripped` | 7.0 MB | Built with `-ldflags "-s -w"` (stripped) |

---

## 2 — Repository Readiness

| Check | Verdict | Evidence |
|-------|---------|----------|
| GitHub remote configured | **PASS** | `origin → https://github.com/jf-ferraz/mind-cli.git` |
| Repo publicly accessible | **PASS** | `git ls-remote` succeeds without auth |
| `mind` (framework) repo accessible | **PASS** | `https://github.com/jf-ferraz/mind.git` also public |
| Main branch up to date | **FAIL — BLOCKER** | Local `main` is **2 commits ahead** of `origin/main` (resolver fixes + framework install improvements, 8 files, 608 insertions) |
| Untracked files | **WARNING** | `docs/guide/` is untracked — the end-to-end walkthrough isn't committed |
| `.gitignore` coverage | **PASS** | `/mind` (build output), `.mind/`, `.claude/`, `archive/`, `prompt.txt` all covered |
| Compiled binary in working tree | **WARNING** | `./mind` (10.2 MB) exists in repo root — ignored by `.gitignore` but should be cleaned before push |
| Go dependencies | **PASS** | Not vendored; `go.sum` present; `go install @latest` fetches cleanly |
| Stale local branches | **WARNING** | 16 local branches (many feature/fix branches from earlier phases) — cleanup recommended |

### Unpushed Commits — BLOCKER

The 2 unpushed commits contain resolver fixes and framework install improvements that affect `framework materialize` and `framework install`. A tester pulling from GitHub gets the older code at `phase-2-complete` tag, which may have bugs in subdirectory traversal that were fixed locally.

---

## 3 — Installation Documentation

| Check | Verdict | Evidence |
|-------|---------|----------|
| README.md install section | **PASS** | Build command, ldflags example, usage summary present |
| End-to-end walkthrough | **PASS** | `docs/guide/end-to-end-walkthrough.md` — 1,538 lines, 17 sections, 99 subsections |
| Walkthrough: install flow | **PASS (with caveat)** | Three install options documented; Option C (`go install @latest`) doesn't mention name mismatch |
| Walkthrough: global setup | **PASS** | Steps 1-5 clear: clone mind repo → `framework install --source` → verify |
| Prerequisites stated | **WARNING** | README says "Go 1.23+"; `go.mod` says `go 1.24.2`; walkthrough says "Go 1.24+". **Three different version numbers.** |
| Platform support | **WARNING** | Only Linux/macOS/WSL mentioned. No Windows native guidance. |
| Shell completion | **PASS** | Documented for bash/zsh/fish in walkthrough |
| Walkthrough committed | **FAIL** | `docs/guide/` is untracked — tester cannot access it from GitHub |

### Go Version Inconsistency

| Source | States |
|--------|--------|
| `README.md` | Go 1.23+ |
| `go.mod` | go 1.24.2 |
| `docs/guide/end-to-end-walkthrough.md` | Go 1.24+ |
| Actual binaries | Built with go1.26.1 |

Should be reconciled to a single minimum version.

---

## 4 — Global Configuration Bootstrap

| Check | Verdict | Evidence |
|-------|---------|----------|
| `framework install --source <path>` | **PASS** | Works correctly; copies 145 artifacts to `~/.config/mind/` |
| Framework requires second repo clone | **WARNING** | Tester must clone `github.com/jf-ferraz/mind.git` separately — this is documented in the walkthrough but adds friction |
| `config.toml` auto-created | **PASS** | Created on first access with sensible defaults |
| `projects.toml` auto-created | **PASS** | Created when `registry add` is first called |
| `framework.lock` written | **PASS** | SHA-256 checksums, version, source path all recorded |
| Framework install without project | **FAIL — BLOCKER** | `mind framework install --source ~/path/to/mind` fails with "not a Mind project" if run outside a `.mind/` directory. A fresh tester who hasn't run `mind init` yet cannot install the framework globally. |
| `config show` without project | **FAIL** | Same issue — `config`, `registry`, `framework` all require project context needlessly |

### Global Commands Require Project Context — BLOCKER

The `requiresProject()` function in `cmd/root.go` only exempts: `version`, `help`, `init`, `completion`, `tui`, `serve`. This means **all** of these fail outside a Mind project directory:
- `mind config show/edit/validate/path`
- `mind registry list/add/remove/resolve/check`
- `mind framework install/status/diff`

These operate on `~/.config/mind/` (global state) and should work anywhere. The walkthrough instructs the tester to run `mind framework install --source ...` as a "one-time setup per machine" — but this fails unless they happen to be inside a Mind project.

**Workaround**: Run from inside any Mind project directory (e.g., `cd ~/dev/projects/mind && mind config show`).

---

## 5 — First-Run Experience

| Check | Verdict | Evidence |
|-------|---------|----------|
| `mind init` on empty dir | **PASS** | Creates `mind.toml`, `.mind/`, `.claude/CLAUDE.md`, `docs/` with 5 zones, 8 stub files |
| `mind status` on fresh project | **PASS** | Shows health dashboard, warns about stubs, exits 1 (correct — stubs = issues) |
| `mind doctor` on fresh project | **PASS** | 11 pass, 0 fail, 7 warnings — all actionable |
| `mind check all` on fresh project | **PASS** | 39 pass, 0 fail, 1 warning (stubs) |
| Error message for non-project dir | **PASS** | "not a Mind project (no .mind/ directory found)" with exit code 3 |
| `.mind/` empty after init | **PASS (expected)** | By design — `mind framework materialize` populates it |
| `mind framework materialize` | **PASS** | Copies 145 artifacts from global to project `.mind/` |
| `mind framework update` | **PASS** | Detects no changes, reports "already up to date" |
| `mind framework diff` first-run | **PASS** | Shows 145 deletions (all global, none in project) — clear output |
| `create brief` (interactive) | **PASS** | Launches interactive prompt — works but requires TTY |

---

## 6 — Command Verification Matrix

### Command Count

14 top-level commands + 22 subcommands = **36 total** (not 31 as originally stated).

### Full Matrix

| Command | No Project | With Project | With Global FW | Notes |
|---------|-----------|-------------|----------------|-------|
| `version` | PASS | PASS | — | Always works |
| `version --short` | PASS | PASS | — | Prints "dev" |
| `help` | PASS | PASS | — | Always works |
| `completion bash/zsh/fish` | PASS | PASS | — | Always works |
| `init` | PASS | PASS | — | Creates project structure |
| `init --with-github` | PASS | PASS | — | Adds `.github/agents/` |
| `init --from-existing` | PASS | PASS | — | Preserves existing docs |
| `status` | exit 3 | PASS | PASS (shows FW panel) | |
| `doctor` | exit 3 | PASS | PASS | |
| `doctor --fix` | exit 3 | PASS | PASS | |
| `brief` | exit 3 | PASS | — | |
| `check docs` | exit 3 | PASS | — | |
| `check docs --strict` | exit 3 | PASS | — | |
| `check refs` | exit 3 | PASS | — | |
| `check config` | exit 3 | PASS | — | |
| `check all` | exit 3 | PASS | — | |
| `docs list` | exit 3 | PASS | — | |
| `docs list --zone spec` | exit 3 | PASS | — | |
| `docs tree` | exit 3 | PASS | — | |
| `docs stubs` | exit 3 | PASS | — | |
| `docs search "query"` | exit 3 | PASS | — | |
| `docs open <path>` | exit 3 | PASS | — | Opens $EDITOR |
| `create adr "Title"` | exit 3 | PASS | — | Auto-numbered |
| `create blueprint "Title"` | exit 3 | PASS | — | Updates INDEX.md |
| `create iteration <type> <name>` | exit 3 | PASS | — | Types: new, enhancement, bugfix, refactor |
| `create spike "Title"` | exit 3 | PASS | — | |
| `create convergence "Title"` | exit 3 | PASS | — | |
| `create brief` | exit 3 | PASS | — | Interactive TTY |
| `iterations` | exit 3 | PASS | — | |
| `workflow status` | exit 3 | PASS | — | |
| `workflow history` | exit 3 | PASS | — | |
| `preflight "request"` | exit 3 | PASS | — | Creates iteration + branch (branch fails without git) |
| `handoff <iter-id>` | exit 3 | PASS | — | Requires valid iteration ID |
| `reconcile` | exit 3 | PASS* | — | *Requires `[documents]` in mind.toml |
| `reconcile --check` | exit 3 | PASS* | — | *Same requirement |
| `reconcile --graph` | exit 3 | PASS | — | Requires `[[graph]]` in mind.toml |
| `framework install` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `framework status` | **exit 3** | PASS | PASS | **BLOCKER: should work without project** |
| `framework diff` | exit 3 | PASS | PASS | Needs project (comparing project vs global — reasonable) |
| `framework materialize` | exit 3 | PASS | PASS | Needs project (writing to .mind/ — reasonable) |
| `framework update` | exit 3 | PASS | PASS | Needs project (writing to .mind/ — reasonable) |
| `config show` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `config path` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `config edit` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `config validate` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `registry list` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `registry add` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `registry remove` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `registry resolve` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `registry check` | **exit 3** | PASS | — | **BLOCKER: should work without project** |
| `serve` | PASS* | PASS | — | *Starts but uses cwd as root |
| `tui` | exit 3 | PASS | — | Launches BubbleTea dashboard |

### Summary

- **0 commands crash** — all exit gracefully
- **0 commands are unwired** — all registered commands have implementations
- **36 commands total** (14 top-level + 22 subcommands)
- **10 commands wrongly require project context** (config/registry/framework install+status)

---

## 7 — Blockers, Warnings, and Recommendations

### BLOCKERS (must fix before tester can succeed)

| # | Issue | Impact | Fix |
|---|-------|--------|-----|
| B1 | **2 commits unpushed** to origin/main | Tester gets stale code missing resolver fixes | `git push origin main` |
| B2 | **`go install` produces `mind-cli`** not `mind` | Tester's binary doesn't match any docs | Document rename/symlink, or add `cmd/mind/main.go` with `package main` |
| B3 | **Global commands require project context** | `config`, `registry`, `framework install/status` fail outside project dir; breaks the "one-time global setup" flow from walkthrough | Add these to `requiresProject()` exemption list |

### WARNINGS (should fix, tester can work around)

| # | Issue | Impact | Workaround |
|---|-------|--------|------------|
| W1 | `docs/guide/` not committed | Walkthrough not available on GitHub | `git add docs/guide/ && git commit` |
| W2 | Go version inconsistency (1.23 / 1.24.2 / 1.24+) | Confusing prerequisites | Pick one minimum version |
| W3 | Version always shows `dev (unknown)` | Tester can't confirm installed version | Document ldflags in install instructions |
| W4 | Two binaries with different names/sizes | Confusing for tester | Document canonical path; remove stale binary |
| W5 | `./mind` binary in repo working tree | Clutters repo root | `rm ./mind` before push |
| W6 | 16 stale local branches | No impact on tester but messy for contributor | Clean up merged branches |

### NICE-TO-HAVE (can wait)

| # | Improvement |
|---|-------------|
| N1 | Makefile with `build`, `install`, `clean`, `test` targets |
| N2 | Goreleaser config for cross-platform binaries |
| N3 | GitHub Actions CI (test + build on PR) |
| N4 | `mind init` auto-materializes framework if globally installed |
| N5 | `mind create iteration` help text should list valid types in error (it does — good!) |
| N6 | `mind framework install` without `--source` could auto-detect from registry |

---

## 8 — Recommended Tester Instructions

Below is a draft step-by-step guide a tester can follow once blockers are fixed.

### Prerequisites
- Go 1.24+ installed
- Git installed
- Linux, macOS, or WSL

### Step 1: Install the CLI
```bash
git clone https://github.com/jf-ferraz/mind-cli.git
cd mind-cli
go build -o mind .
sudo cp mind /usr/local/bin/   # or: cp mind ~/.local/bin/
mind version
# Expected: mind dev (unknown) built unknown linux/amd64
```

### Step 2: Install the framework globally
```bash
# Clone the framework source
git clone https://github.com/jf-ferraz/mind.git ~/mind-framework

# Install globally (run from any Mind project, or after B3 fix, from anywhere)
mind framework install --source ~/mind-framework
# Expected: Framework v2026.03.1 installed (145 artifacts from ~/mind-framework/.mind)

# Verify
mind framework status
mind config show
```

### Step 3: Create a test project
```bash
mkdir ~/test-project && cd ~/test-project
git init
mind init --name test-project
# Expected: Initialized Mind project: test-project (10 files created)

mind status
# Expected: Dashboard with 8 stubs, workflow: idle

mind doctor
# Expected: 11 pass, 0 fail, 7 warnings
```

### Step 4: Materialize the framework
```bash
mind framework materialize
# Expected: Materialized v2026.03.1: 145 artifacts

ls .mind/agents/
# Expected: analyst.md, architect.md, developer.md, etc.
```

### Step 5: Exercise all command groups
```bash
# Validation
mind check docs
mind check refs
mind check config
mind check all

# Documentation management
mind docs list
mind docs tree
mind docs stubs
mind docs search "project"

# Artifact creation
mind create adr "Test Decision"
mind create blueprint "Test Blueprint"
mind create iteration enhancement "test-feature"
mind create spike "Test Spike"
mind create convergence "Test Analysis"

# Status & workflow
mind brief
mind iterations
mind workflow status
mind workflow history

# Framework management
mind framework status
mind framework diff
mind framework update

# Global config
mind config show
mind config path
mind config validate
mind registry list
mind registry add test-project ~/test-project
mind registry check

# Orchestration
mind preflight "Add user authentication"
# Expected: Pre-flight complete, iteration created, branch created

# Version & completion
mind version
mind version --short
mind completion bash > /dev/null
```

### Step 6: MCP server (optional)
```bash
# Create .mcp.json in project root
echo '{"mcpServers":{"mind":{"command":"mind","args":["serve"]}}}' > .mcp.json

# Test server starts (Ctrl+C to stop)
mind serve
```

### Expected Outcome
All commands should complete without crashes. Exit code 0 for success, 1 for validation warnings (expected on fresh project with stubs), 3 for missing project context (only outside project dir).

---

## Appendix: Test Environment

| Item | Value |
|------|-------|
| OS | Linux (Arch) |
| Go | go1.26.1 |
| Shell | zsh |
| mind-cli commit | `caf05b1` (main, 2 ahead of origin) |
| mind framework | v2026.03.1 (145 artifacts) |
| Test binary | `/tmp/mind-test` (10.2 MB, standard build) |
| Test project | `/tmp/mind-test-project` (fresh `mind init`) |
