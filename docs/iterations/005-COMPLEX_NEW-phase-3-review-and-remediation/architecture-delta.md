# Architecture Delta — Phase 3 Review and Remediation

**Iteration**: 005-COMPLEX_NEW-phase-3-review-and-remediation
**Branch**: complex/phase-3-review-and-remediation
**Date**: 2026-03-12
**Architect**: architect agent
**Primary Input**: `docs/iterations/005-COMPLEX_NEW-phase-3-review-and-remediation/requirements-delta.md`
**Convergence Guidance**: `docs/knowledge/phase-3-review-convergence.md` (Recommendation 4 Option A)
**FR Range**: FR-145, FR-146, FR-148

---

## Current Structure

### internal/orchestrate/ — Current State

`internal/orchestrate/` contains one file: `preflight.go`. It defines `PreflightService` (7-step pre-flight) and `PromptBuilder` (prompt assembly). The package also declares two result types that belong to distinct flows:

- `PreflightResult` — output from `PreflightService.Run()`, correctly scoped to preflight
- `HandoffResult` — declared in `preflight.go` but belonging to a handoff flow that is not implemented here

`PreflightService` has a `Handoff()` method (lines 144–182) that always returns an error: `"iteration lookup requires IterationRepo — use HandoffService instead"`. This method is a broken public API stub. It references a service type (`HandoffService`) that does not exist. The actual handoff logic lives entirely in `cmd/handoff.go`.

The `findIteration()` method (lines 184–188) is a stub that always errors. The `updateCurrentState()` method (lines 190–193) is a no-op stub that silently does nothing.

### cmd/handoff.go — Current State

`cmd/handoff.go` implements the full 5-step handoff sequence inline in `runHandoff()`:

1. Iteration lookup via `workflowSvc.Show(iterID)`
2. Artifact validation loop (inline, not delegated)
3. `appendToCurrentState()` — a package-level function calling `os.ReadFile` and `os.WriteFile` directly on `docs/state/current.md`
4. `stateRepo.WriteWorkflow(nil)` (correct — uses StateRepo)
5. `currentGitBranch()` and `branchAhead()` — package-level git shell helpers

`branchAhead()` (line 164) hardcodes `"HEAD...main"` as the git rev-list range. It does not read from `mind.toml`.

`appendToCurrentState()` (lines 125–151) violates the 4-layer rule: the `cmd/` presentation layer calls `os.ReadFile`/`os.WriteFile` directly, bypassing `StateRepo`. Only `internal/repo/fs/` packages should perform raw filesystem I/O.

### StateRepo — Current State

`StateRepo` in `internal/repo/interfaces.go` has two methods:

```go
ReadWorkflow() (*domain.WorkflowState, error)
WriteWorkflow(state *domain.WorkflowState) error
```

No method exists for reading or writing `docs/state/current.md`. That file is accessed directly by `cmd/handoff.go`, a layer violation.

### domain/project.go — Governance Struct

The `Governance` struct contains: `MaxRetries`, `ReviewPolicy`, `CommitPolicy`, `BranchStrategy`. There is no `DefaultBranch` field. The `BranchStrategy` field describes the naming convention for feature branches (e.g., `"type-descriptor"`); it is semantically distinct from the comparison base for `git rev-list`.

---

## Proposed Changes

### New Components

| Component | Package | Responsibility | Dependencies | FR |
|-----------|---------|----------------|-------------|-----|
| `HandoffService` | `internal/orchestrate/handoff.go` | Encapsulates the 5-step handoff sequence: artifact validation, gate run, current.md update via StateRepo, state clear, branch report | `domain`, `repo.IterationRepo`, `repo.StateRepo`, `*service.ValidationService` | FR-146 |
| `HandoffResult` (moved) | `internal/orchestrate/handoff.go` | Result type for a completed handoff run — moved from `preflight.go` to co-locate with `HandoffService` | `domain` | FR-146 |

### Modified Components

| Component | Change | FR |
|-----------|--------|-----|
| `StateRepo` interface (`internal/repo/interfaces.go`) | Add `AppendCurrentState(iter *domain.Iteration) error` method | FR-145 |
| `internal/repo/fs/state_repo.go` | Implement `AppendCurrentState()` — encapsulate the `appendToCurrentState` logic currently in `cmd/handoff.go` | FR-145 |
| `internal/repo/mem/state_repo.go` | Implement `AppendCurrentState()` as a no-op that returns nil — sufficient for test use (tests verify behavior through fs impl or check the method is callable) | FR-145 |
| `internal/orchestrate/preflight.go` | Remove `Handoff()`, `findIteration()`, `updateCurrentState()`, and `HandoffResult` — these move to `handoff.go` | FR-146 |
| `domain/project.go` | Add `DefaultBranch string` field to `Governance` struct with TOML tag `default-branch` | FR-148 |
| `cmd/handoff.go` | Remove `appendToCurrentState()` and its `os` imports; delegate to `HandoffService`; read default branch from `Config.Governance.DefaultBranch` | FR-145, FR-146, FR-148 |
| `internal/deps/deps.go` | Add `HandoffSvc *orchestrate.HandoffService` field to `Deps` struct; wire in `Build()` | FR-146 |

---

## New Interfaces

### StateRepo.AppendCurrentState()

**Method signature**:

```go
AppendCurrentState(iter *domain.Iteration) error
```

**Parameter**: `iter *domain.Iteration` — the completed iteration. The implementation reads `iter.DirName`, `iter.Seq`, and uses `time.Now()` to produce the entry date. Passing nil is an error (implementations should return `fmt.Errorf("iter must not be nil")`).

**Behavior specification**:
- Reads `docs/state/current.md` from the project root
- Formats an entry: `- **{YYYY-MM-DD}** — {iter.DirName} completed (@iteration/{iter.Seq:03d})`
- Locates the `## Recent Changes` section header; inserts the entry immediately after the blank line that follows the header
- If the `## Recent Changes` section does not exist, appends it to the end of the file with the entry as its first item
- Writes the modified content back to disk
- If the file does not exist, returns a descriptive error (does not create the file — `current.md` is a required project document)

**fs implementation** (`internal/repo/fs/state_repo.go`): Direct `os.ReadFile`/`os.WriteFile` — identical logic to the current `appendToCurrentState()` in `cmd/handoff.go`, relocated to the correct layer.

**mem implementation** (`internal/repo/mem/state_repo.go`): Appends to an in-memory `CurrentStateEntries []string` slice for test assertion. No filesystem access. Returns nil on success.

**Rejected alternatives**:
- `AppendCurrentState(iterID, branch, summary string) error` — accepts flat strings instead of the domain type. Rejected because callers already hold `*domain.Iteration`; string decomposition at the call site would duplicate field access logic. The domain type is the correct abstraction boundary. Using a domain type is consistent with how `WriteWorkflow` accepts `*domain.WorkflowState`.
- `AppendCurrentState(entry domain.CurrentStateEntry) error` — introduce a new dedicated domain type. Rejected per minimum viable structure rule — no new domain type is needed when `*domain.Iteration` already carries all required fields. Adding a thin wrapper type would be premature abstraction.
- Keeping the logic in `cmd/handoff.go` and marking it `:TEMP:` — Recommendation 4 Option B from convergence analysis. Rejected because FR-145 explicitly targets this layer violation and the acceptance criterion for FR-145 states `cmd/handoff.go` must not import `os` for `current.md` operations. Option B does not satisfy the acceptance criterion.

---

## New Data

### HandoffResult

`HandoffResult` is moved from `internal/orchestrate/preflight.go` to `internal/orchestrate/handoff.go`. The existing type definition is extended to carry all information produced by the 5-step sequence:

```go
type HandoffResult struct {
    IterationID      string              // iteration directory name (e.g., "005-COMPLEX_NEW-...")
    ArtifactsPresent int
    ArtifactsTotal   int
    MissingArtifacts []string
    GateResult       *domain.GateResult
    StateCleared     bool
    Branch           string
    AheadBy          int
    Artifacts        []string            // names of all present artifact files
    Errors           []string            // non-fatal errors accumulated during steps 3–5
}
```

**Field notes**:
- `IterationID` replaces the current `IterationPath` — callers already have the full path through `*domain.Iteration`; the ID (directory name) is the display-relevant identifier
- `AheadBy` is new — currently `branchAhead()` only prints this to stdout; surfacing it in the result enables testability and future rendering changes
- `Artifacts` is new — the slice of artifact file names that are present, enabling renderers to display which artifacts passed
- `Errors` is new — non-fatal errors from steps 3, 4, 5 (current.md update failure, state clear failure) are currently printed inline; collecting them in the result enables `cmd/handoff.go` to decide how to present them without the service layer doing output

---

## Decisions

### Decision: HandoffService Constructor Signature

**Context**: FR-146 requires a `HandoffService` in `internal/orchestrate/`. The convergence analysis (Recommendation 4 Option A) specifies `IterationRepo` as a constructor dependency. The question is whether to accept explicit parameters or receive the `Deps` struct.

**Choice**: Explicit constructor parameters.

```go
func NewHandoffService(
    projectRoot   string,
    iterRepo      repo.IterationRepo,
    stateRepo     repo.StateRepo,
    validationSvc *service.ValidationService,
) *HandoffService
```

**Rationale**:
- `PreflightService` uses the same explicit-parameters pattern (`NewPreflightService` takes 6 explicit parameters). `HandoffService` should follow the same convention for consistency within the `internal/orchestrate/` package.
- The `Deps` struct is a wiring convenience defined at the `cmd/` boundary. Service-layer constructors should not depend on it — that would create an upward dependency from `internal/orchestrate/` to a `cmd/`-adjacent package, violating layer rules.
- Explicit parameters make the dependency graph visible at the call site in `internal/deps/deps.go` and enable direct construction in tests without building a `Deps` struct.
- `projectRoot` is passed explicitly (not via Deps) because it is needed for `validationSvc.RunGate(projectRoot)` and the git helper calls, matching the `PreflightService` constructor pattern.

**Rejected alternatives**:
- `NewHandoffService(d *deps.Deps)` — passing the full Deps struct. Rejected because `internal/orchestrate/` would need to import `internal/deps/`, which imports `internal/orchestrate/` (through `Build()` wiring), creating a circular dependency.
- `NewHandoffService(iterRepo repo.IterationRepo, stateRepo repo.StateRepo, validationSvc *service.ValidationService, projectRoot string)` (projectRoot last) — functionally equivalent but inconsistent with `PreflightService` where projectRoot is the first parameter. Rejected for consistency.
- Embedding `*PreflightService` in `HandoffService` — the two services share `stateRepo` and `validationSvc` but have distinct flows. Embedding would create implicit coupling between preflight and handoff. Rejected per convergence analysis note: "preflight is about starting work, handoff is about completing it."

---

### Decision: StateRepo.AppendCurrentState Signature

**Choice**: Accept `*domain.Iteration` as the parameter type.

```go
AppendCurrentState(iter *domain.Iteration) error
```

**Rationale**:
- The caller (`HandoffService`) already holds the `*domain.Iteration` retrieved during step 1. No extraction or conversion is needed at the call site.
- The method only needs `iter.DirName`, `iter.Seq`, and the current time to produce the entry. All of these are available on `*domain.Iteration` or can be derived locally (time.Now() is the repo's responsibility, not the caller's).
- Accepting the full domain type allows the implementation to evolve — if the entry format changes (e.g., adding `iter.Type` to the entry text), the interface signature does not need to change.
- This is consistent with `WriteWorkflow(*domain.WorkflowState)` — the existing StateRepo method accepts the full domain type, not decomposed fields.

**Rejected alternatives**:
- `AppendCurrentState(iterID, branch, summary string) error` — flat string parameters. Rejected because this decomposes a domain type at the boundary, requiring the caller to extract and format fields it may not know the formatting rules for. The branch field is not actually needed for the `current.md` entry format (the current `appendToCurrentState()` does not include the branch in the entry). Introducing a `branch` parameter would add a field the implementation ignores.
- `AppendCurrentState(seq int, dirName string, date time.Time) error` — pass only the specific fields used. Rejected because `time.Time` belongs in the infrastructure layer (the implementation calls `time.Now()`), not the interface signature. The entry date is a write-time concern, not a caller concern.

---

### Decision: mind.toml Default-Branch Governance Key

**Context**: FR-148 requires `branchAhead()` in `cmd/handoff.go` to read the comparison base from `mind.toml` instead of hardcoding `"main"`. The existing `Governance` struct has `BranchStrategy` (naming convention) but no `DefaultBranch` field.

**Choice**: Add `DefaultBranch string` to the `domain.Governance` struct with TOML tag `"default-branch"`.

**Key name in mind.toml**: `governance.default-branch`

**Example**:
```toml
[governance]
max-retries     = 2
review-policy   = "evidence-based"
commit-policy   = "conventional"
branch-strategy = "type-descriptor"
default-branch  = "main"
```

**Flow**:
1. `domain/project.go` — add `DefaultBranch string \`toml:"default-branch"\`` to `Governance`
2. `internal/repo/fs/config_repo.go` — no change needed; `go-toml/v2` reads the new field automatically via struct tag
3. `internal/deps/deps.go` `Build()` — no change; `ConfigRepo` already reads `mind.toml` into `*domain.Config`
4. `cmd/handoff.go` `runHandoff()` — reads `deps.ConfigRepo.ReadProjectConfig()`, extracts `cfg.Governance.DefaultBranch`; falls back to `"main"` when the field is empty string
5. `HandoffService.Run()` — accepts `defaultBranch string` as a parameter (passed by `cmd/handoff.go` after reading config), OR reads it from a `ConfigRepo` dependency

**`HandoffService` and the default branch**: The service receives `defaultBranch string` as a parameter to `Run(iterID, defaultBranch string) (*HandoffResult, error)`. This avoids giving `HandoffService` a `ConfigRepo` dependency solely for this one field. The caller (`cmd/handoff.go`) already reads config for other purposes (e.g., gate commands require config access via `ValidationService`), so reading `DefaultBranch` at the `cmd/` level before calling `Run()` is consistent with the thin-handler pattern.

**Fallback behavior**: When `cfg.Governance.DefaultBranch == ""` (field absent from mind.toml or empty), fall back to `"main"`. The fallback is implemented in `cmd/handoff.go` before passing the value to `HandoffService.Run()`.

**Rejected alternatives**:
- Reuse `BranchStrategy` — `BranchStrategy` describes the naming convention for new branches (e.g., `"type-descriptor"` produces `bugfix/slug`). It is semantically unrelated to the upstream comparison base for `git rev-list`. Rejected to avoid semantic overloading of an existing field.
- Add a top-level `[git]` section to `mind.toml` — introduces a new TOML section for a single key. Over-engineering for a single field. Rejected per minimum viable structure rule.
- Read from `.git/config` or `git symbolic-ref refs/remotes/origin/HEAD` — interrogating git configuration is more correct in theory but adds process execution complexity and failure modes. For a governance tool, explicit declaration in `mind.toml` is clearer and more portable. Rejected for operational simplicity.
- Treat absent `default-branch` as an error — too strict; most projects do not need to customize this. Rejected in favor of a `"main"` fallback.

---

## Migration Path

The migration from inline `cmd/handoff.go` logic to `HandoffService` delegation follows this sequence:

**Step 1 — StateRepo extension** (prerequisite for Steps 2 and 3):
- Add `AppendCurrentState(*domain.Iteration) error` to `internal/repo/interfaces.go`
- Implement in `internal/repo/fs/state_repo.go` by relocating the `appendToCurrentState()` body verbatim
- Implement in `internal/repo/mem/state_repo.go` as a no-op (returns nil, optionally records the call for test assertions)
- Verify existing tests pass (`go test ./internal/repo/...`)

**Step 2 — domain.Governance extension**:
- Add `DefaultBranch string \`toml:"default-branch"\`` to the `Governance` struct in `domain/project.go`
- No other domain changes required; existing tests should continue passing

**Step 3 — Create HandoffService** (`internal/orchestrate/handoff.go`):
- Define `HandoffService` struct and `NewHandoffService()` constructor
- Implement `Run(iterID, defaultBranch string) (*HandoffResult, error)` by extracting the 5-step logic from `cmd/handoff.go` — this is a mechanical lift-and-shift
- Replace `appendToCurrentState()` call with `s.stateRepo.AppendCurrentState(iter)`
- Replace hardcoded `"main"` with the `defaultBranch` parameter
- Move `HandoffResult` type definition here; update `AheadBy`, `Artifacts`, `Errors` fields

**Step 4 — Remove dead API from PreflightService** (`internal/orchestrate/preflight.go`):
- Delete `Handoff()`, `findIteration()`, `updateCurrentState()` methods
- Delete the `HandoffResult` type (now in `handoff.go`)
- Verify `preflight.go` still compiles

**Step 5 — Update cmd/handoff.go**:
- Remove `appendToCurrentState()` function and its `os`/`filepath`/`time` imports
- Add config read: `cfg, _ := deps.ConfigRepo.ReadProjectConfig(); defaultBranch := "main"; if cfg != nil && cfg.Governance.DefaultBranch != "" { defaultBranch = cfg.Governance.DefaultBranch }`
- Delegate to `deps.HandoffSvc.Run(iterID, defaultBranch)`
- Render `HandoffResult` (output behavior unchanged — same step labels, same field values)

**Step 6 — Wire HandoffService in deps.go**:
- Add `HandoffSvc *orchestrate.HandoffService` to the `Deps` struct
- Add `orchestrate.NewHandoffService(root, iterRepo, stateRepo, validationSvc)` in `Build()`

**Behavioral invariant**: At each step, `go test ./...` must pass. The handoff output (step labels, artifact counts, gate results, state messages, branch info) must be identical before and after the refactor. The migration introduces no functional changes — it is a pure relocation of existing logic.

---

## Out of Scope

The following are explicitly excluded from this architecture-delta:

- **FR-140** (MCP notifications/initialized): Purely a behavioral fix within `internal/mcp/server.go`. No new types, interfaces, or wiring decisions required.
- **FR-141** (Quality dimension constants): A constant rename in `domain/quality.go` and a regex update in `internal/service/quality.go`. No structural decisions required.
- **FR-142, FR-143, FR-144** (Test coverage): Test file additions do not require architectural decisions. Test structure follows the existing `*_test.go` pattern in `internal/service/` and `internal/orchestrate/`.
- **FR-147** (Preflight doc-failure blocking): A behavioral change within `PreflightService.Run()` — add an error check after step 3. No interface changes, no new types.
- **FR-149** (classify.go adapter): A thin file wrapping `domain.Classify()` and `domain.Slugify()`. No structural decisions.
- **FR-150** (splitOn/trimSpace replacement): A `strings` stdlib substitution. No structural decisions.
- **FR-151** (Renderer routing for preflight): Refactoring `cmd/preflight.go` to use `Renderer`. No interface changes; follows the existing Renderer pattern.
- **Phase 4** commands (`mind watch`, `mind run`).
- **PromptBuilder** filesystem access — accepted as tech debt per the `internal/reconcile/hash.go` architectural precedent (documented in convergence analysis).
- **New MCP tools** — the 16 registered tools are unchanged.
