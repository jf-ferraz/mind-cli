# mind-cli — TODO & Ideas

> **Last updated**: 2026-03-13
> **Status**: Living document — items added as ideas surface, checked off as delivered

---

## 1. CLI Initialization & Recovery

| # | Item | Priority | Status |
|---|------|----------|--------|
| 1.1 | **`mind init --repair`** — Add a repair/force mode that fills in missing files (mind.toml, doc stubs) without requiring `.mind/` to be absent. Currently `init` refuses to run if `.mind/` exists, leaving partially initialized projects stuck with no recovery path other than deleting `.mind/` and re-initializing. | High | TODO |
| 1.2 | **`doctor --fix` coverage gap** — `doctor --fix` cannot regenerate `mind.toml` or spec docs like `architecture.md`, `domain-model.md`. It only creates files in its check list (e.g., `glossary.md`). Expand fix capabilities to regenerate all expected stubs and the manifest. | High | TODO |
| 1.3 | **`mind init` auto-materialize** — If framework is globally installed (`~/.config/mind/` exists), auto-run `framework materialize` after init so the user doesn't need a separate step. | Low | TODO |

---

## 2. mind.toml — Documentation

| # | Item | Priority | Status |
|---|------|----------|--------|
| 2.1 | **Comprehensive mind.toml reference** — Write a full reference document covering every section, every parameter, valid values, defaults, and examples. Should serve as both onboarding guide and day-to-day reference. Cover: `[manifest]`, `[project]`, `[stack]`, `[governance]`, `[[documents]]`, `[[graph]]`, `[profiles.*]`, `[framework]`. | High | TODO |

---

## 3. mind.toml — Parameter Exploration & Design

| # | Item | Priority | Status |
|---|------|----------|--------|
| 3.1 | **Parameter audit** — Review all current mind.toml parameters and identify gaps. Specifically investigate: | Medium | TODO |
|     | - **Git policies**: branch protection rules, commit message conventions (conventional commits?), merge strategies (squash/rebase/merge), required reviewers, signed commits | | |
|     | - **Analyze/Conversation parameters**: dialectical analysis config, persona weighting, convergence thresholds, max rounds, synthesis strategy | | |
|     | - **What else could the manifest drive?** Build commands, test commands, lint commands, deploy targets, environment variables, secret references | | |
| 3.2 | **Multi-library / polyglot projects** — Explore scenarios where a project has multiple libraries, secondary languages, or mixed stacks. How should `[project]` and `[stack]` evolve? Consider: | Medium | TODO |
|     | - `stack.secondary_languages = ["Python", "TypeScript"]` | | |
|     | - `[[stack.libraries]]` with name, version, purpose | | |
|     | - Monorepo support: per-module mind.toml or `[[modules]]` section | | |
|     | - Workspace-level vs package-level configuration | | |
| 3.3 | **Profiles deep dive** — Explore what profiles can control and how they compose. Consider: | Medium | TODO |
|     | - Environment-specific profiles: `[profiles.dev]`, `[profiles.staging]`, `[profiles.prod]` | | |
|     | - Team-specific profiles: `[profiles.backend-team]`, `[profiles.frontend-team]` | | |
|     | - CI vs local profiles: different governance strictness, different validation suites | | |
|     | - Profile inheritance/composition: can profiles extend other profiles? | | |
|     | - Profile activation: CLI flag (`--profile dev`), env var, auto-detect? | | |

---

## 4. mind.toml — Naming & Semantics

| # | Item | Priority | Status |
|---|------|----------|--------|
| 4.1 | **`branch-strategy` rename** — Current name conflates two concepts: | Medium | TODO |
|     | - **Branch naming format** (e.g., `feature/X`, `fix/X`, `feat/TICKET-123-description`) — this is what `branch-strategy` currently controls | | |
|     | - **Branch strategy/workflow model** (e.g., gitflow, trunk-based, GitHub flow) — a higher-level concern | | |
|     | Consider splitting into: `branch-format` or `branch-style` for naming patterns, and `branch-strategy` for the workflow model. Or pick one clear name if only one concept is needed now. | | |

---

## 5. Project Registry

| # | Item | Priority | Status |
|---|------|----------|--------|
| 5.1 | **`description` field in registry entries** — Add an optional `description` parameter to `projects.toml` entries so users can annotate what each registered project is for. Currently only stores `alias → path`. Example: | Medium | TODO |
|     | ```toml | | |
|     | [projects.mind-cli] | | |
|     | path = "~/dev/projects/mind-cli" | | |
|     | description = "CLI companion for the Mind Agent Framework" | | |
|     | ``` | | |
|     | Show description in `mind registry list` output. | | |

---

## 6. Domain Model

| # | Item | Priority | Status |
|---|------|----------|--------|
| 6.1 | **Expand `project.type` values** — Current valid types may not cover all real-world cases. Add/adjust to include: `engine`, `backend`, `frontend`, `library`, `monorepo`, `cli`, `service`, `sdk`, `api`, `plugin`, `framework`. Review `domain/project.go` for the enum and validation in `internal/validate/`. | Medium | TODO |

---

## 7. Validation & Integrity

| # | Item | Priority | Status |
|---|------|----------|--------|
| 7.1 | **Debug ID/path consistency across refs and config** — Audit how document IDs and paths are used across: | High | TODO |
|     | - `mind.toml` `[[documents]]` entries (id, path, zone, status) | | |
|     | - `mind.toml` `[[graph]]` edges (from/to reference document IDs) | | |
|     | - Refs validation suite (11 checks in `internal/validate/`) | | |
|     | - Config validation suite (12 checks) | | |
|     | - Reconciliation engine (hash tracking uses paths) | | |
|     | Ensure: IDs are unique, paths resolve to real files, graph edges reference valid IDs, no orphaned references, no silent mismatches between config and filesystem. | | |

---

## 8. Build & Distribution

| # | Item | Priority | Status |
|---|------|----------|--------|
| 8.1 | **Binary name mismatch** — `go install` produces `mind-cli` not `mind`. Options: (a) keep documenting the symlink workaround, (b) add a `cmd/mind/main.go` wrapper, (c) Goreleaser handles it. | Medium | TODO |
| 8.2 | **Goreleaser** — Cross-platform binary releases via GitHub Releases. Eliminates binary name issue, provides `.tar.gz`/`.zip` with correct name, generates checksums. | Medium | TODO |
| 8.3 | **GitHub Actions CI** — Automated test + build + vet on PR. Catch regressions before merge. | Medium | TODO |
| 8.4 | **Version via `debug.ReadBuildInfo`** — `go install` users see `dev (unknown)`. Explore reading Go module version from build info as fallback when ldflags aren't set. | Low | TODO |

---

## 9. Housekeeping

| # | Item | Priority | Status |
|---|------|----------|--------|
| 9.1 | **Clean up stale local branches** — 16 local branches from earlier phases. All merged. Clean up: `complex/`, `enhancement/`, `feat/`, `feature/`, `fix/`, `phase-2/`, `refactor/`, `test-run`. | Low | TODO |
| 9.2 | **Windows native support** — No Windows guidance in docs. Linux/macOS/WSL only currently. | Low | TODO |
| 9.3 | **`mind framework install` auto-detect** — Without `--source`, detect framework location from registry or well-known paths. | Low | TODO |
