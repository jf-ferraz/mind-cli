# Architecture Delta: Pre-Phase 3 Cleanup

**Iteration**: 004-pre-phase-3-cleanup
**FR Range**: FR-125 through FR-139
**Base Architecture**: docs/spec/architecture.md (Phase 1 + 1.5 + 2)

---

## 1. FR-125 + FR-137: Deps Interface Migration

### Current State

`internal/deps/deps.go` declares 7 repository fields using concrete `*fs.X` types:

```go
DocRepo     *fs.DocRepo
IterRepo    *fs.IterationRepo
BriefRepo   *fs.BriefRepo
ConfigRepo  *fs.ConfigRepo
LockRepo    *fs.LockRepo
StateRepo   *fs.StateRepo
QualityRepo *fs.QualityRepo
```

This forces `tui/app.go` to import `internal/repo/fs` for two reasons:
1. Type resolution of the Deps struct fields (implicit, via the `deps` package importing `fs`).
2. Two direct calls to `fs.DetectProject()` (lines 39 and 341).

`cmd/root.go` also holds package-level variables with concrete `*fs.` types (`docRepo *fs.DocRepo`, `iterRepo *fs.IterationRepo`, `briefRepo *fs.BriefRepo`).

### Structural Changes

#### 1a. `internal/deps/deps.go`

Change the 7 repository fields from concrete to interface types:

| Field | Current Type | New Type |
|-------|-------------|----------|
| `DocRepo` | `*fs.DocRepo` | `repo.DocRepo` |
| `IterRepo` | `*fs.IterationRepo` | `repo.IterationRepo` |
| `BriefRepo` | `*fs.BriefRepo` | `repo.BriefRepo` |
| `ConfigRepo` | `*fs.ConfigRepo` | `repo.ConfigRepo` |
| `LockRepo` | `*fs.LockRepo` | `repo.LockRepo` |
| `StateRepo` | `*fs.StateRepo` | `repo.StateRepo` |
| `QualityRepo` | `*fs.QualityRepo` | `repo.QualityRepo` |

The import list changes from `internal/repo/fs` to `internal/repo`. The `Build()` function body is unchanged -- it still constructs `fs.NewDocRepo(root)` etc. -- but the struct fields it populates are now interface-typed. The `internal/repo/fs` import remains in `deps.go` because `Build()` calls `fs.New*()` constructors. This is correct: `deps.go` is the wiring site.

Service fields remain as concrete types (`*service.ProjectService`, etc.) since services are not interfaces in the current architecture. No change needed.

The `Build()` function signature remains unchanged: `func Build(root string, r *render.Renderer) *Deps`.

#### 1b. `cmd/root.go`

The 3 package-level repository variables change from concrete to interface types:

| Variable | Current Type | New Type |
|----------|-------------|----------|
| `docRepo` | `*fs.DocRepo` | `repo.DocRepo` |
| `iterRepo` | `*fs.IterationRepo` | `repo.IterationRepo` |
| `briefRepo` | `*fs.BriefRepo` | `repo.BriefRepo` |

This replaces the `internal/repo/fs` import with `internal/repo` in `root.go`. The `fs` import is still needed in `cmd/helpers.go` for `fs.FindProjectRoot()` and `fs.FindProjectRootFrom()`, so the `cmd` package retains the `fs` import (via `helpers.go`), but `root.go` itself no longer imports it.

#### 1c. `cmd/status.go`

`runStatus()` currently calls `fs.DetectProject(projectRoot)` directly (line 23). This call should move to `projectSvc` -- the `ProjectService` already has the concept of "detect project." Add a `DetectProject(root string) (*domain.Project, error)` method to `ProjectService` that delegates to `fs.DetectProject()`. The service already receives repos through constructor injection; the new method needs the project root string, which it can accept as a parameter.

Alternatively, since `fs.DetectProject()` creates a temporary `ConfigRepo` internally, and `ProjectService` already holds repos, the method can be implemented by reading the config through the injected `ConfigRepo` and constructing a `Project` value. However, this is a larger refactoring. The simpler approach: add a project-detection function to `ProjectService` that calls `fs.DetectProject()` internally (the service layer is allowed to import `fs`).

**Decision**: Keep `fs.DetectProject()` where it is. In `cmd/status.go`, replace the direct `fs.DetectProject()` call with a call through the `projectSvc`. Add `DetectProject(root string)` to `ProjectService`. This eliminates the `fs` import from `status.go`.

#### 1d. `tui/app.go`

Two calls to `fs.DetectProject()`:
1. `NewApp()` line 39: detects project name for the title bar.
2. `loadHealth()` line 341: detects project for health assembly.

Both should go through `deps.ProjectSvc.DetectProject()` after the method is added (see 1c above). This eliminates the `internal/repo/fs` import from `tui/app.go` entirely.

The import list changes: remove `"github.com/jf-ferraz/mind-cli/internal/repo/fs"`. No other import additions needed since `tui/app.go` already imports `deps`.

### Key Decision: ProjectService.DetectProject

- **Decided**: Add `DetectProject(root string) (*domain.Project, error)` to `ProjectService`.
- **Why**: `DetectProject` is project-level orchestration (validate `.mind/` exists, read config, build `Project` struct). This is service-layer responsibility. Moving it there decouples presentation layers (`cmd/`, `tui/`) from infrastructure (`fs/`).
- **Rejected**: Passing `DetectProject` as a function field in Deps. This would require a new function type and complicate the Deps struct for a single use case.
- **Rejected**: Moving `DetectProject` into a repo interface method. Project detection is not a repository concern -- it orchestrates across multiple repos (config reading, filesystem stat).
- **Consequences**: `ProjectService` gains a constructor parameter for `projectRoot string` (it needs the root to check `.mind/` existence), or it delegates to `fs.DetectProject()` internally. The latter is acceptable since service layer is allowed to import infrastructure.

---

## 2. FR-126: mem/ Inverse Dependency Removal

### Current State

`internal/repo/mem/doc_repo.go` imports `internal/repo/fs` solely for the `fs.IsStubContent()` function (line 75). This is an inverse dependency: the test-only in-memory implementation depends on the production filesystem implementation.

### Structural Change

Move `IsStubContent()` and its helper `isBoilerplateLine()` (and `isTableSeparator()`, `isPlaceholderRow()` if they exist as private helpers) out of `internal/repo/fs/doc_repo.go` into a shared location.

**Location choice**: `internal/repo/stub.go` (a new file at the `repo` package level).

**Why `internal/repo/` and not `domain/`**:
- `IsStubContent()` performs content analysis (scanning lines, checking prefixes). While it is a pure function (no I/O), stub detection is an implementation concern of the repository layer -- it defines what "stub" means when reading documents. The `domain/` package defines `Document.IsStub` as a boolean field; the logic for computing that field belongs to the repo layer.
- Placing it in `domain/` would work (the function is pure and uses only stdlib), but would add content-parsing logic to a package that currently contains only types, enums, and simple classification functions (`Slugify`, `Classify`, `BuildGraph`, `Validate`). This would weaken the cohesion of the domain package.

**Why not `internal/repo/shared/`**: Creating a sub-package adds a directory for two functions. The `internal/repo/` package already exists and contains `interfaces.go`. Adding `stub.go` alongside it is simpler and follows the Go idiom of keeping related code in the same package when the package is small.

**Migration**:
1. Create `internal/repo/stub.go` with `IsStubContent()`, `isBoilerplateLine()`, and related helpers (exported as needed).
2. In `internal/repo/fs/doc_repo.go`, change `IsStub()` method to call `repo.IsStubContent()`. The existing `IsStubContent` test in `fs/doc_repo_test.go` moves to `repo/stub_test.go` or remains in `fs/` calling the relocated function.
3. In `internal/repo/mem/doc_repo.go`, change the call from `fs.IsStubContent(data)` to `repo.IsStubContent(data)`. Remove the `internal/repo/fs` import.

**Consequence**: `fs/doc_repo.go` keeps a re-export or thin wrapper `IsStubContent` for backward compatibility if any external code references it. Given that `IsStubContent` is only called by `fs.DocRepo.IsStub()` and `mem.DocRepo.IsStub()`, a re-export is unnecessary -- both callers will be updated.

---

## 3. FR-129: os.Exit to Error Returns

### Current State

29 `os.Exit()` calls across `cmd/` files. The pattern is:
```go
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(2)
return nil // dead code
```

Exit codes used: 1 (validation failure), 2 (runtime error), 3 (config error / not a project), 4 (staleness).

### Structural Change

#### 3a. ExitError Type

Define a structured error type that carries an exit code:

```go
// ExitError wraps an error with a process exit code.
type ExitError struct {
    Code int
    Err  error
}
```

**Location**: `cmd/errors.go` (new file in `cmd/` package).

**Why `cmd/` and not `domain/`**: Exit codes are a CLI presentation concern. The domain defines error semantics (`ErrNotProject`, `ErrBriefMissing`); the `cmd/` layer maps those to OS exit codes. `ExitError` is used only by the `cmd/` package and the `Execute()` function.

**Why not a standalone package**: `ExitError` is consumed exclusively by `cmd/` handlers and the `Execute()` function. No other package needs it. Creating `internal/exitcode/` would be over-engineering.

Convenience constructors:

```go
func exitValidation(err error) *ExitError   { return &ExitError{Code: 1, Err: err} }
func exitRuntime(err error) *ExitError      { return &ExitError{Code: 2, Err: err} }
func exitConfig(err error) *ExitError       { return &ExitError{Code: 3, Err: err} }
func exitStaleness(err error) *ExitError    { return &ExitError{Code: 4, Err: err} }
```

`ExitError` implements the `error` interface via `func (e *ExitError) Error() string` and `func (e *ExitError) Unwrap() error`.

#### 3b. Execute() Exit Code Mapping

Modify `cmd/root.go` `Execute()` to extract exit codes from returned errors:

```go
func Execute() error {
    if err := rootCmd.Execute(); err != nil {
        var exitErr *ExitError
        if errors.As(err, &exitErr) {
            os.Exit(exitErr.Code)
        }
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1) // default for untyped errors
    }
    return nil
}
```

This is the **single** `os.Exit()` call site in the entire `cmd/` package (outside of test helpers). All command handlers return errors instead.

#### 3c. PersistentPreRunE Migration

The `os.Exit(3)` in `PersistentPreRunE` (root.go line 62) for `isNotProject` errors changes to:

```go
return exitConfig(err)
```

The error message printing (`fmt.Fprintln(os.Stderr, err)`) moves to `Execute()`, which prints before exiting.

#### 3d. Command Handler Migration Pattern

Each handler replaces `os.Exit(N)` + `return nil` with `return exitX(err)` or `return exitX(fmt.Errorf(...))`:

**Before**:
```go
if !report.Ok() {
    os.Exit(1)
}
return nil
```

**After**:
```go
if !report.Ok() {
    return exitValidation(fmt.Errorf("%d check(s) failed", report.Failed))
}
return nil
```

For handlers that print output before exiting (like `runCheckDocs` which calls `fmt.Print(renderer.RenderValidation(...))`), the output printing remains -- only the exit mechanism changes.

For `runReconcile`'s error classification (`isConfigError`), the pattern becomes:

**Before**:
```go
if isConfigError(err) {
    os.Exit(3)
}
os.Exit(2)
```

**After**:
```go
if isConfigError(err) {
    return exitConfig(err)
}
return exitRuntime(err)
```

#### 3e. Error Message Printing

Currently, some handlers print to stderr before calling `os.Exit()`. With the new pattern, `Execute()` handles stderr printing for all errors. To avoid double-printing, handlers should NOT print to stderr when returning an `ExitError` -- the error message is in the `ExitError.Err` field.

However, some commands print formatted output (via renderer) AND then exit with a non-zero code (e.g., `runCheckDocs` prints the validation report, then exits 1 if checks failed). In these cases, the output is already printed; the `ExitError` just carries the code and an error message for Cobra's error handling. `Execute()` should print the error message only for errors that were NOT already displayed to the user.

**Decision**: `ExitError` gets a `Quiet` field. When `Quiet` is true, `Execute()` exits with the code but does not print the error message (the handler already rendered output).

### Key Decision: ExitError Location

- **Decided**: `cmd/errors.go` in the `cmd` package.
- **Why**: Exit codes are presentation-layer semantics. Only `cmd/` maps domain errors to OS exit codes. The MCP server (Phase 3) will have its own error-to-JSON-RPC mapping in `mcp/`.
- **Rejected**: `domain/errors.go`. Exit codes are not domain concepts. The domain defines *what* went wrong; the presentation layer decides *how* to report it to the OS.
- **Rejected**: `internal/exitcode/` package. Only one consumer. Packages with one consumer are premature abstraction.
- **Consequences**: `cmd/` gains one file (`errors.go`, ~40 lines). All handlers import nothing new -- `ExitError` is in the same package.

---

## 4. FR-130: DiagnosticStatus Enum

### Current State

`domain/health.go` line 29: `Status string` in the `Diagnostic` struct. Raw string values `"pass"`, `"fail"`, `"warn"` are used in:
- `DoctorService.addDiag()` (passes raw strings)
- `DoctorService.Run()` summary counting (compares `d.Status` against `"pass"`, `"fail"`, `"warn"`)
- `DoctorService.applyFixes()` (compares `d.Status == "pass"`)
- `Renderer.renderDoctorText()` (switches on `d.Status` with `"fail"`, `"warn"`)

### Structural Change

#### 4a. Define DiagnosticStatus Enum

In `domain/health.go`, add:

```go
// DiagnosticStatus indicates the outcome of a doctor diagnostic check.
type DiagnosticStatus string

const (
    DiagPass DiagnosticStatus = "pass"
    DiagFail DiagnosticStatus = "fail"
    DiagWarn DiagnosticStatus = "warn"
)
```

This follows the existing pattern exactly:
- `IterationStatus string` with `IterInProgress`, `IterComplete`, `IterIncomplete` (lowercase values)
- `LockStatus string` with `LockClean`, `LockStale`, `LockDirty` (UPPERCASE values)
- `EntryStatus string` with `EntryPresent`, `EntryMissing`, `EntryChanged`, `EntryUnchanged` (UPPERCASE values)

The naming convention `DiagPass`/`DiagFail`/`DiagWarn` uses the `Diag` prefix to avoid collision with other enums (e.g., `LevelFail` in `CheckLevel`). The values are lowercase (`"pass"`, `"fail"`, `"warn"`) to match the existing JSON contract.

#### 4b. Change Diagnostic.Status Field Type

```go
type Diagnostic struct {
    // ...
    Status   DiagnosticStatus `json:"status"`
    // ...
}
```

JSON serialization is unchanged because `DiagnosticStatus` is a typed `string` -- `json.Marshal` produces the same `"pass"`, `"fail"`, `"warn"` values.

#### 4c. Derive Level from Status

The `Level CheckLevel` field on `Diagnostic` (currently JSON-excluded via `json:"-"`) can be derived from `DiagnosticStatus` rather than computed separately. The `addDiag()` method's switch statement becomes:

```go
func (s *DoctorService) addDiag(report *domain.DoctorReport, category, check string, status domain.DiagnosticStatus, message, fixHint string, autoFixable bool) {
    level := domain.LevelInfo
    switch status {
    case domain.DiagFail:
        level = domain.LevelFail
    case domain.DiagWarn:
        level = domain.LevelWarn
    }
    // ...
}
```

The `addDiag()` parameter changes from `status string` to `status domain.DiagnosticStatus`. All call sites in `doctor.go` change from passing `"pass"`, `"fail"`, `"warn"` string literals to `domain.DiagPass`, `domain.DiagFail`, `domain.DiagWarn`.

#### 4d. Update Summary Counting

```go
switch d.Status {
case domain.DiagPass:
    report.Summary.Pass++
case domain.DiagFail:
    report.Summary.Fail++
case domain.DiagWarn:
    report.Summary.Warn++
}
```

#### 4e. Update Renderer

`internal/render/render.go` `renderDoctorText()` changes from:
```go
switch d.Status {
case "fail": ...
case "warn": ...
}
```
to:
```go
switch d.Status {
case domain.DiagFail: ...
case domain.DiagWarn: ...
}
```

### Key Decision: Separate DiagnosticStatus vs. Reusing CheckLevel

- **Decided**: New `DiagnosticStatus` enum, separate from `CheckLevel`.
- **Why**: `CheckLevel` has values `FAIL`, `WARN`, `INFO` (uppercase). `Diagnostic.Status` has values `pass`, `fail`, `warn` (lowercase) -- different semantics and different JSON contract. `CheckLevel` indicates severity; `DiagnosticStatus` indicates outcome. A check can have level `WARN` but status `pass` (the check ran but found only warnings). Conflating them would require mapping logic everywhere.
- **Rejected**: Reusing `CheckLevel`. Different value sets (`INFO` has no `DiagnosticStatus` equivalent; `pass` has no `CheckLevel` equivalent). Different JSON serialization.
- **Consequences**: DC-3 gains one more enum. The `Level` field on `Diagnostic` remains for backward compatibility (some rendering code uses it). It could be removed later as a follow-up.

---

## 5. FR-127: Propagation Fix

### Current State

`internal/reconcile/propagate.go` `buildReason()` (line 90-104):

```go
for _, edge := range graph.Reverse[targetID] {
    if depth == 0 && edge.From == sourceID {
        edgeReason = edgeTypeReason(edge.Type)
        break
    }
}
```

At depth > 0, the loop body never executes because `depth == 0` is false. The edge reason falls back to `"may be outdated"` (the default from `edgeTypeReason`'s `default` case, which only handles `EdgeInforms` implicitly).

### Structural Change

#### 5a. Queue Item Carries Edge Type

The `queueItem` struct gains an `edgeType` field:

```go
type queueItem struct {
    nodeID   string
    sourceID string
    depth    int
    edgeType domain.EdgeType  // edge type from the immediate predecessor
}
```

When seeding the queue:
```go
for _, edge := range graph.Forward[changedID] {
    queue = append(queue, queueItem{
        nodeID:   edge.To,
        sourceID: changedID,
        depth:    0,
        edgeType: edge.Type,
    })
}
```

When enqueueing transitive dependents:
```go
for _, edge := range graph.Forward[item.nodeID] {
    queue = append(queue, queueItem{
        nodeID:   edge.To,
        sourceID: item.sourceID,
        depth:    item.depth + 1,
        edgeType: edge.Type,  // edge from item.nodeID to edge.To
    })
}
```

#### 5b. Simplify buildReason

`buildReason` no longer needs to search `graph.Reverse` to find the edge type -- it receives the edge type directly:

```go
func buildReason(sourceID string, edgeType domain.EdgeType, depth int) string {
    reason := edgeTypeReason(edgeType)
    if depth == 0 {
        return fmt.Sprintf("dependency changed: %s (%s)", sourceID, reason)
    }
    return fmt.Sprintf("dependency changed: %s (via transitive chain, %s)", sourceID, reason)
}
```

The call site changes from:
```go
reason := buildReason(graph, item.sourceID, item.nodeID, item.depth)
```
to:
```go
reason := buildReason(item.sourceID, item.edgeType, item.depth)
```

#### 5c. Fix edgeTypeReason Default Case

The `edgeTypeReason` function currently has `default: return "may be outdated"`. This default case matches `EdgeInforms` (which is `"informs"`). This should be explicit:

```go
func edgeTypeReason(edgeType domain.EdgeType) string {
    switch edgeType {
    case domain.EdgeRequires:
        return "prerequisite changed"
    case domain.EdgeValidates:
        return "needs re-validation"
    case domain.EdgeInforms:
        return "may be outdated"
    default:
        return "may be outdated"
    }
}
```

The `EdgeInforms` case is now explicit, and the `default` is retained as a safety net for unknown edge types (defensive programming).

### Key Decision: Edge Type on Queue vs. Reverse Lookup

- **Decided**: Carry edge type on the queue item.
- **Why**: At enqueue time, the edge being traversed is already known. Looking it up again from `graph.Reverse` at dequeue time is redundant work and the source of the current bug (the lookup condition `depth == 0` prevents finding the edge at depth > 0).
- **Rejected**: Fixing the `buildReason` reverse lookup to work at any depth. This would require searching for "which predecessor caused this node to be enqueued" -- information that is lost after enqueue. The queue item is the right place to carry this context.
- **Consequences**: ~10 lines changed. Queue item is one field larger (one `string` field). No performance impact.

---

## 6. FR-128: Flag Rename

### Current State

`cmd/root.go` line 91:
```go
rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "Path to project root (default: auto-detect)")
```

### Structural Change

Change the flag registration:
```go
rootCmd.PersistentFlags().StringVarP(&flagProject, "project-root", "p", "", "Path to project root (default: auto-detect)")
```

The Go variable `flagProject` remains unchanged (it is internal to the `cmd` package). Only the CLI-visible flag name changes.

**Decision**: Hard rename without backward-compatible alias.

**Why not alias with deprecation**: The `--project` flag is not in any released version (mind-cli has not been published). There are no external consumers to break. Adding a deprecated alias adds code that must be removed later. Clean break now is cheaper.

**Affected files**:
- `cmd/root.go`: Flag registration (1 line).
- `docs/spec/architecture.md`: References to `--project` in the component table and examples (FR-133 handles this).
- No test files reference `--project` directly (there are zero `cmd/` tests currently).

---

## 7. FR-131 + FR-132: Test Coverage Architecture

### FR-131: cmd/ Exit Code Tests

**Test file**: `cmd/cmd_test.go` (new file).

**Test strategy**: Use `cobra.Command.Execute()` with argument injection, not process forking. The `ExitError` type (from FR-129) enables this:

```go
func TestCheckDocsExitCode(t *testing.T) {
    // Set up in-memory deps
    // Execute command
    // Assert error type and code
    var exitErr *ExitError
    if errors.As(err, &exitErr) {
        assert(exitErr.Code == 1)
    }
}
```

**Challenge**: Command handlers currently read from package-level variables (`projectRoot`, `renderer`, etc.) populated by `PersistentPreRunE`. Tests must either:
1. Populate these variables directly before calling `Execute()`.
2. Bypass `PersistentPreRunE` by setting up a test root with `.mind/` and `mind.toml`.

**Decision**: Option 2 (test with real project structure using `t.TempDir()`). This matches existing test patterns in `internal/service/` tests. The `PersistentPreRunE` runs naturally, wires deps, and the handler executes. This validates the full command pipeline including wiring.

**Minimum test coverage**:
- `check docs` with passing checks -> exit 0
- `check docs` with failing checks -> exit 1
- `reconcile --check` with stale docs -> exit 4
- Command invoked outside a Mind project -> exit 3
- `status` with issues -> exit 1

### FR-132: render/ JSON Tests

**Test file**: `internal/render/render_test.go` (new file).

**Test strategy**: Construct domain type values, call `Render*()` in JSON mode, parse the output as JSON, and assert field presence and types.

```go
func TestRenderHealthJSON(t *testing.T) {
    r := render.New(render.ModeJSON, 80)
    health := &domain.ProjectHealth{...}
    output := r.RenderHealth(health)
    var parsed map[string]any
    json.Unmarshal([]byte(output), &parsed)
    // Assert expected fields exist
}
```

**Minimum test coverage**:
- `RenderHealth()` -> fields: `project`, `brief`, `zones`, `warnings`, `suggestions`
- `RenderValidation()` -> fields: `suite`, `checks`, `total`, `passed`, `failed`
- `RenderReconcileResult()` -> fields: `changed`, `stale`, `missing`, `status`, `stats`
- `RenderDoctor()` -> fields: `diagnostics`, `summary`

No structural changes to `render.go` itself -- only new test files.

---

## 8. FR-133 through FR-136: Documentation Updates

These are documentation-only changes. No architectural decisions needed. The developer will:

- **FR-133**: Fix `cmd/tui_cmd.go` -> `cmd/tui.go` in architecture.md; add `InitService`, `DoctorService`, `ReconciliationService` to Phase 1 component table; update `--project` -> `--project-root`.
- **FR-134**: Update requirements.md overview to acknowledge all phases.
- **FR-135**: Add `DiagnosticStatus` to DC-3 and the supporting types table in domain-model.md.
- **FR-136**: Update current.md to remove resolved issues and add iteration 004 entry.

---

## 9. Migration Path

The following order maintains a buildable codebase at each step. Each step produces a state where `go build ./...` and `go test ./...` pass.

### Step 1: FR-130 -- DiagnosticStatus Enum

**Why first**: Pure domain change with no cross-cutting dependencies. Adds the type, changes the field, updates all consumers (`doctor.go`, `render.go`). This is self-contained and touches the fewest files.

**Files touched**: `domain/health.go`, `internal/service/doctor.go`, `internal/render/render.go`.

### Step 2: FR-126 -- mem/ Inverse Dependency

**Why second**: Self-contained refactoring. Move `IsStubContent()` to `internal/repo/stub.go`, update imports in `fs/doc_repo.go` and `mem/doc_repo.go`.

**Files touched**: `internal/repo/stub.go` (new), `internal/repo/fs/doc_repo.go`, `internal/repo/mem/doc_repo.go`. Existing `fs/doc_repo_test.go` may need an import update.

### Step 3: FR-127 -- Propagation Fix

**Why third**: Isolated to one file. No interface changes, no import changes.

**Files touched**: `internal/reconcile/propagate.go`. Existing propagation tests will validate the fix.

### Step 4: FR-125 + FR-137 -- Deps Interface Migration

**Why after Steps 1-3**: This is the highest-risk change (touches `deps.go`, `root.go`, `status.go`, `tui/app.go`, and potentially other cmd/ files). Completing the simpler refactorings first reduces the diff size and merge conflict surface.

**Sub-steps**:
1. Add `DetectProject()` method to `ProjectService` (allows `status.go` and `tui/app.go` to stop calling `fs.DetectProject()` directly).
2. Change `deps.Deps` field types from `*fs.X` to `repo.X`.
3. Update `cmd/root.go` package-level variable types.
4. Update `tui/app.go` to use `deps.ProjectSvc.DetectProject()` instead of `fs.DetectProject()`.
5. Remove `internal/repo/fs` import from `tui/app.go`.

**Files touched**: `internal/deps/deps.go`, `cmd/root.go`, `cmd/status.go`, `tui/app.go`, `internal/service/project.go`.

### Step 5: FR-128 -- Flag Rename

**Why after Step 4**: The flag rename is trivial but must happen after Deps migration is stable. It also creates a checkpoint before the large FR-129 change.

**Files touched**: `cmd/root.go` (1 line).

### Step 6: FR-129 -- os.Exit to Error Returns

**Why after Step 5**: This is the largest change by file count (touches every `cmd/*.go` file). Having the Deps migration and flag rename already done means fewer merge conflicts.

**Depends on**: FR-125 (Deps migration must be stable), FR-128 (flag rename done).

**Sub-steps**:
1. Create `cmd/errors.go` with `ExitError` type and convenience constructors.
2. Modify `Execute()` in `root.go` to handle `ExitError`.
3. Migrate each command file: `check.go`, `reconcile.go`, `status.go`, `doctor.go`, `docs.go`, `create.go`, `init.go`, `tui.go`.
4. Remove `os.Exit` calls from `PersistentPreRunE`.

**Files touched**: `cmd/errors.go` (new), `cmd/root.go`, `cmd/check.go`, `cmd/reconcile.go`, `cmd/status.go`, `cmd/doctor.go`, `cmd/docs.go`, `cmd/create.go`, `cmd/init.go`, `cmd/tui.go`.

### Step 7: FR-131 + FR-132 -- Test Coverage

**Why after Step 6**: Tests for exit codes require the `ExitError` pattern from FR-129 to be in place. Render tests have no dependencies but are grouped here to batch test additions.

**Files touched**: `cmd/cmd_test.go` (new), `internal/render/render_test.go` (new).

### Step 8: FR-133 through FR-136 -- Documentation Updates

**Why last**: Documentation reflects the implemented state. Updating docs before implementation risks inconsistency.

**Files touched**: `docs/spec/architecture.md`, `docs/spec/requirements.md`, `docs/spec/domain-model.md`, `docs/state/current.md`.

### Step 9: FR-138 + FR-139 -- Verification

Final verification step: `go vet ./...`, `go build ./...`, `go test ./...`. Not a code change -- a validation gate.

---

## 10. Convergence Cross-Reference

| Convergence Recommendation | Adopted / Adapted / Deviated | Notes |
|---------------------------|------------------------------|-------|
| R1: Migrate Deps to interface types | **Adopted** | FR-125 + FR-137. Added `ProjectService.DetectProject()` to eliminate `tui/` -> `fs/` import. |
| R2: Fix transitive propagation reasons | **Adopted** | FR-127. Carried edge type on queue item instead of fixing reverse lookup. |
| R3: Rename --project to --project-root | **Adopted** | FR-128. Hard rename (no deprecation alias) since no released versions exist. |
| R4: Add cmd/ exit code tests | **Adopted** | FR-131. Using `t.TempDir()` with real project structure, not process forking. |
| R5: Add render/ JSON tests | **Adopted** | FR-132. JSON mode tests with field assertion. |
| R6: Fix mem/ importing fs/ | **Adopted** | FR-126. Moved `IsStubContent` to `internal/repo/stub.go`. Convergence suggested `domain/` or `internal/repo/shared/`; we chose `internal/repo/` (same package as interfaces.go) for simplicity. |
| R7: Type Diagnostic.Status | **Adopted** | FR-130. New `DiagnosticStatus` enum (not reusing `CheckLevel`). |
| R8: Replace os.Exit with error returns | **Adopted** | FR-129. `ExitError` type in `cmd/errors.go` with `Execute()` handling. Convergence suggested `cmdError`; we use `ExitError` for clarity. |
| R9: Update architecture docs | **Adopted** | FR-133, FR-134, FR-135, FR-136. |
| R10: Extract DRY utilities | **Deferred** | Out of scope per requirements-delta. COULD priority. |
| Concession: GenerateService file creation is acceptable | **Adopted** | Not in scope. |
| Concession: Editor fallback to vi in TUI is acceptable | **Adopted** | Not in scope. |
| Concession: DoctorService delegation is COULD | **Adopted** | Not in scope. |

---

## 11. Updated Dependency Matrix (Post-Cleanup)

Changes from current state:

```
# Removed dependencies
tui/app.go ──────────X──> internal/repo/fs      (FR-125/FR-137)
internal/repo/mem/ ──X──> internal/repo/fs       (FR-126)

# New dependencies
tui/app.go ──────────────> internal/deps          (unchanged, already exists)
internal/repo/fs/ ───────> internal/repo           (for IsStubContent)
internal/repo/mem/ ──────> internal/repo            (for IsStubContent)

# Modified (type changes, not import changes)
internal/deps/deps.go ───> internal/repo            (new: for interface types)
internal/deps/deps.go ───> internal/repo/fs          (retained: for constructors)
cmd/root.go ─────────────> internal/repo             (replaces internal/repo/fs for types)
```

All other dependencies remain unchanged.

---

## 12. New Files Summary

| File | Purpose | Lines (est.) |
|------|---------|-------------|
| `internal/repo/stub.go` | `IsStubContent()` and helpers moved from `fs/doc_repo.go` | ~40 |
| `cmd/errors.go` | `ExitError` type and convenience constructors | ~40 |
| `cmd/cmd_test.go` | Exit code tests for cmd/ handlers | ~150-200 |
| `internal/render/render_test.go` | JSON output mode tests | ~100-150 |
