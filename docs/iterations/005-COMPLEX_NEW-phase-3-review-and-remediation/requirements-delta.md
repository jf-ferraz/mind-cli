# Requirements Delta — Phase 3 Review and Remediation

**Iteration**: 005-COMPLEX_NEW-phase-3-review-and-remediation
**Type**: COMPLEX_NEW (treated as ENHANCEMENT for artifact purposes)
**Branch**: complex/phase-3-review-and-remediation
**Date**: 2026-03-12
**Analyst**: analyst agent
**Primary Input**: `docs/knowledge/phase-3-review-convergence.md`
**FR Range**: FR-140 through FR-148

---

## Current State

### What Phase 3 Delivered

Phase 3 (iteration 005, implemented outside the Mind Framework process) delivered:

- `mind preflight "<request>"` — 7-step pre-flight command: classify request, run brief gate, validate docs, create iteration folder, create git branch, write workflow state, generate prompt
- `mind preflight --resume` — detect and display an in-progress workflow from `docs/state/workflow.md`
- `mind handoff <iter-id>` — 5-step post-workflow command: validate iteration artifacts, run deterministic gate, update `current.md`, clear workflow state, report branch status
- `mind serve` — MCP server over stdio (JSON-RPC 2.0), 16 registered tools for Claude Code integration
- `.mcp.json` — Claude Code auto-discovery configuration
- New packages: `internal/orchestrate/`, `internal/mcp/`
- New domain types: `domain/quality.go` (`QualityEntry`, `QualityDimension`, dimension constants), `domain/gate.go` (`GateResult`, `GateCommandResult`)
- `StateRepo` extended with `WriteWorkflow()` (fs + mem implementations)
- `WorkflowService` extended with `UpdateState()` and `Show()`
- `QualityService` with `Log()` for parsing convergence files → `quality-log.yml`

### Known Problems (Convergence Findings)

The convergence analysis (`docs/knowledge/phase-3-review-convergence.md`) identified the following defects with a weighted score of 2.0/5.0 (below the 3.0 gate threshold):

**MUST Fix:**

- **M-1** (`internal/mcp/server.go:108`): The MCP `notifications/initialized` notification triggers the `default:` dispatch branch, which calls `errorResponse(req.ID, errMethodNotFound, ...)`. Per JSON-RPC 2.0, a notification (no `id` field) must not receive any response. The current behavior is a protocol violation that puts AC-14 (Claude Code integration) at risk.
- **M-2** (`domain/quality.go:47–52`, `internal/service/quality.go:78–97`): Quality dimension constants (`rigor`, `coverage`, `objectivity`, `convergence`, `depth`) do not match the conversation workflow rubric names (`perspective_diversity`, `evidence_quality`, `concession_depth`, `challenge_substantiveness`, `synthesis_quality`, `actionability`). Only `actionability` overlaps. Every quality log entry written via `mind_log_quality` silently produces five zero-dimension scores and an incorrect total score, corrupting `quality-log.yml`.
- **M-3** (all Phase 3 packages): Zero test coverage exists for `internal/orchestrate/`, `internal/mcp/`, `internal/service/quality.go`, `cmd/preflight.go`, `cmd/handoff.go`, and `cmd/serve.go`. All 17 BP-08 acceptance criteria are unverifiable. This violates the project quality standard established in Phases 1, 1.5, and 2 (service-layer packages have test files).

**SHOULD Fix:**

- **S-1** (`cmd/handoff.go:125–151`): `appendToCurrentState()` calls `os.ReadFile` and `os.WriteFile` directly on `docs/state/current.md`, bypassing `StateRepo`. This is a presentation-layer violation of the 4-layer architecture rules (only `internal/repo/fs/` should perform raw filesystem I/O).
- **S-2** (`internal/orchestrate/preflight.go:144–188`): `PreflightService.Handoff()` is an exported method that always returns `fmt.Errorf("iteration lookup requires IterationRepo — use HandoffService instead")`. It is a dead public API — the actual handoff logic lives in `cmd/handoff.go`. This naming implies a `HandoffService` that does not exist, misleading future maintainers.
- **S-3** (`internal/orchestrate/preflight.go:190–193`): `PreflightService.updateCurrentState()` is a no-op stub — the function body is empty. Handoff step 3 (update `current.md`) is advertised by the command but silently does nothing at the service layer.
- **S-4** (`internal/orchestrate/preflight.go:78`): Preflight step 3 runs docs validation but treats all outcomes, including hard failures (`docsReport.Failed > 0`), as non-blocking. BP-08 acceptance criterion AC-3 states preflight should block on blocking documentation failures. BP-07 line 97 says warnings are non-blocking; reconciled interpretation is: block on `Failed > 0`, proceed with warning for warnings only.
- **S-5** (`cmd/handoff.go:165`): `branchAhead()` hardcodes `HEAD...main` as the git rev-list range. This breaks when the repository's default branch is not `main`. The `mind.toml` governance section is not consulted.

**COULD Fix:**

- **S-6** (`internal/orchestrate/`): BP-08 specifies `classify.go` as a separate file in `internal/orchestrate/`. The classification logic is correctly placed in `domain/` (`domain.Classify()`), making the spec expectation architecturally moot, but the package structure expectation is formally unmet.
- **C-1** (`internal/mcp/tools.go:336–357`): `splitOn()` and `trimSpace()` re-implement `strings.Split()` and `strings.TrimSpace()` from the Go standard library.
- **C-4** (`cmd/preflight.go:117–156`): `renderPreflightResult()` uses raw `strings.Builder` and `fmt.Printf` instead of routing through the `Renderer`. Unlike other commands, `mind preflight` does not support `--json` output.

---

## Desired State

After remediation, the Phase 3 AI Bridge implementation MUST meet the following conditions:

1. **Protocol-correct MCP server**: The server dispatches `notifications/*` (and other JSON-RPC notifications — messages with null/absent `id`) by returning `nil` (no response written). Claude Code connects, exchanges `initialize`/`notifications/initialized`, and calls tools without protocol errors.

2. **Accurate quality dimension model**: `domain/quality.go` constants match the six names defined in `.mind/conversation/config/quality.yml`. `QualityService.Log()` correctly parses convergence documents and produces quality log entries where all six dimension scores are non-zero when the source document uses the rubric.

3. **Test coverage at project standard**: All Phase 3 service-layer packages (`internal/orchestrate/`, `internal/mcp/`, `internal/service/quality.go`) have test files. Coverage on new Phase 3 code is ≥ 80%. The `go test ./...` command passes with no failures.

4. **No layer violations in cmd/handoff.go**: `current.md` updates are performed through `StateRepo` (or an extended repo interface), not via direct `os.ReadFile`/`os.WriteFile` calls in the presentation layer.

5. **No dead public API on PreflightService**: Either a `HandoffService` exists in `internal/orchestrate/` and `cmd/handoff.go` delegates to it, or `PreflightService.Handoff()` is removed and the `cmd/` implementation is explicitly marked as the intended location with an architecture note.

6. **Preflight blocks on hard doc failures**: `mind preflight` exits non-zero with an actionable error message when `docsReport.Failed > 0`. Warnings remain non-blocking.

7. **Branch comparison uses configured default branch**: `branchAhead()` reads the default branch from `mind.toml` governance settings or falls back to a configurable default rather than hardcoding `main`.

---

## New Requirements (FR-140 onward)

### Phase 3 Remediation Requirements

#### MCP Protocol Compliance

- **FR-140**: The MCP server (`internal/mcp/server.go`) MUST handle JSON-RPC 2.0 notifications (requests where `id` is absent or null) by returning `nil` from `handleRaw()` — producing no response on the wire. The `notifications/initialized` notification sent by MCP clients after the initialize handshake MUST be silently acknowledged (no response). Any method matching the `notifications/*` pattern MUST follow this rule. [MUST]

#### Quality Domain Model Alignment

- **FR-141**: The quality dimension constants in `domain/quality.go` MUST be renamed to match the six rubric dimension names defined in `.mind/conversation/config/quality.yml`: `perspective_diversity`, `evidence_quality`, `concession_depth`, `challenge_substantiveness`, `synthesis_quality`, `actionability`. The old constant names (`rigor`, `coverage`, `objectivity`, `convergence`, `depth`) MUST be removed. `internal/service/quality.go` parsing logic MUST be updated to recognize the new names so that all six dimensions parse with non-zero values from real convergence documents. [MUST]

#### Test Coverage

- **FR-142**: The `internal/orchestrate/` package MUST have a `preflight_test.go` file with unit tests covering: `PreflightService.Run()` for each `RequestType`, the brief gate blocking condition, the doc validation step (including the hard-failure blocking behavior when `docsReport.Failed > 0` after FR-145 is implemented), and `WorkflowState` write. [MUST]
- **FR-143**: The `internal/mcp/` package MUST have a `server_test.go` file with unit tests covering: `handleRaw()` for `initialize`, `tools/list`, `tools/call` (success and unknown tool), malformed JSON, and `notifications/initialized` (verifying `nil` return — no response written). [MUST]
- **FR-144**: `internal/service/quality.go` MUST have a `quality_test.go` file with unit tests covering `Log()` and the dimension-parsing regex, verified against at least one real convergence document sample that includes all six rubric dimension names. [MUST]

#### Layer Violation Remediation

- **FR-145**: `StateRepo` (`internal/repo/interfaces.go`) SHOULD be extended with a method for appending a completed iteration entry to `docs/state/current.md`. `cmd/handoff.go` SHOULD call this method instead of using `os.ReadFile`/`os.WriteFile` directly. After this change, `cmd/handoff.go` MUST NOT import `os` for the purpose of reading or writing `current.md`. [SHOULD]

#### HandoffService Extraction

- **FR-146**: `PreflightService.Handoff()` and its always-erroring `findIteration()` stub SHOULD be removed from `internal/orchestrate/preflight.go`. A `HandoffService` SHOULD be introduced in `internal/orchestrate/` with proper `IterationRepo` constructor injection, encapsulating the 5-step handoff sequence (artifact validation, gate run, current.md update, state clear, branch report). `cmd/handoff.go` SHOULD delegate to `HandoffService` instead of implementing the logic inline. [SHOULD]

#### Preflight Doc-Failure Blocking

- **FR-147**: `PreflightService.Run()` SHOULD block (return a non-nil error) when `docsReport.Failed > 0` after step 3 (doc validation). The error message MUST identify the number of failing checks and direct the user to run `mind check docs` for details. When `docsReport.Failed == 0` and warnings exist, preflight SHOULD proceed and add the warning count to `PreflightResult.Warnings`. [SHOULD]

#### Branch Comparison Portability

- **FR-148**: `cmd/handoff.go` `branchAhead()` SHOULD NOT hardcode `HEAD...main`. The comparison base SHOULD be read from `mind.toml` governance settings (e.g., a `default-branch` key) and fall back to `"main"` only when no governance setting is present. [SHOULD]

#### Structural Placeholder (COULD)

- **FR-149**: A `classify.go` file COULD be created in `internal/orchestrate/` as a thin adapter that re-exports `domain.Classify()` and `domain.Slugify()`, satisfying the BP-08 package structure expectation without moving any logic. [COULD]
- **FR-150**: `splitOn()` and `trimSpace()` helper functions in `internal/mcp/tools.go` COULD be replaced with direct calls to `strings.Split()` and `strings.TrimSpace()` from the Go standard library. [COULD]
- **FR-151**: `renderPreflightResult()` in `cmd/preflight.go` COULD be refactored to use the `Renderer` type consistent with other commands, adding `--json` support to `mind preflight`. [COULD]

---

## Unchanged Requirements

The following requirements MUST continue to be satisfied after all changes in this iteration. No existing test may be deleted.

- **FR-1 through FR-139**: All Phase 1, 1.5, 2, and pre-Phase 3 cleanup requirements remain in force.
- **FR-138**: `go vet ./...` MUST report zero issues and `go build ./...` MUST succeed. This constraint applies throughout remediation.
- **FR-139**: All pre-existing tests MUST continue to pass. The test count must not decrease. The target after this iteration is ≥ 374 existing tests + new Phase 3 tests.
- Exit code contracts (FR-49 and extensions) are unchanged.
- All `--json` behaviors for existing commands are unchanged.
- TUI tab behavior and Renderer output are unchanged.
- Reconciliation engine behavior is unchanged.
- `QualityEntry.Validate()` business rules (BR-36, BR-37, BR-38 in `domain/quality.go`) — score range [0.0, 5.0], gate threshold 3.0, exactly 6 dimensions — remain unchanged. Only the dimension name constants change.

---

## Domain Model Impact

### domain/quality.go Constant Rename (M-2, FR-141)

This is a **breaking change to exported constants**. The following constants will be renamed:

| Old Constant | Old Value | New Constant | New Value |
|---|---|---|---|
| `DimRigor` | `"rigor"` | `DimPerspectiveDiversity` | `"perspective_diversity"` |
| `DimCoverage` | `"coverage"` | `DimEvidenceQuality` | `"evidence_quality"` |
| `DimObjectivity` | `"objectivity"` | `DimConcessionDepth` | `"concession_depth"` |
| `DimConvergence` | `"convergence"` | `DimChallengeSubstantiveness` | `"challenge_substantiveness"` |
| `DimDepth` | `"depth"` | `DimSynthesisQuality` | `"synthesis_quality"` |
| `DimActionability` | `"actionability"` | `DimActionability` | `"actionability"` (unchanged) |

**Impact scope**:
- `domain/quality.go` — constant definitions
- `internal/service/quality.go` — parsing regex and any constant references
- `tui/quality.go` — any rendering that references dimension names
- `internal/repo/fs/quality_repo.go` and `internal/repo/mem/quality_repo.go` — any constant references
- Any existing test that asserts dimension names

The `QualityEntry.Validate()` logic (checks `len(e.Dimensions) == 6`, range [0,5] per dimension) does not reference constant values and is unaffected.

No existing `quality-log.yml` data is retroactively corrected by this change — historical entries written with old dimension names will remain with those names. Only new entries written after the fix will use the correct names.

---

## Structural Impact

This iteration requires **architect activation**. The following structural decisions are not specified here and must be designed by the architect:

1. **HandoffService** (FR-146): A new service type in `internal/orchestrate/` requires architect decisions on constructor signature, interface for `IterationRepo`, and the division of responsibilities between `PreflightService` and `HandoffService`.

2. **StateRepo extension** (FR-145): Adding an `AppendCurrentState()` (or equivalent) method to `StateRepo` requires architect decisions on method signature, the domain type accepted as input, and the `mem/` test implementation behavior.

3. **mind.toml governance key for default branch** (FR-148): If a new governance key is added, the architect must specify the key name, where it is parsed in `ConfigRepo`, and how it flows to `cmd/handoff.go`.

All other findings (FR-140, FR-141, FR-142, FR-143, FR-144, FR-149, FR-150, FR-151) are self-contained changes that do not require new types, interfaces, or architectural decisions.

---

## Acceptance Criteria

### FR-140 (MCP Notification Handling)

- GIVEN an MCP client sends a `notifications/initialized` JSON-RPC message (with `id` absent or null) WHEN `handleRaw()` processes it THEN the return value is `nil` and no bytes are written to the transport.
- GIVEN `internal/mcp/server_test.go` tests `notifications/initialized` handling WHEN `go test ./internal/mcp/...` runs THEN the test passes and no response message is written.
- GIVEN the MCP server is running via `mind serve` and Claude Code connects WHEN Claude Code sends `notifications/initialized` after the initialize handshake THEN no error is logged by Claude Code and `tools/list` responds correctly.

### FR-141 (Quality Dimension Alignment)

- GIVEN `domain/quality.go` WHEN inspected THEN the exported constants are `DimPerspectiveDiversity`, `DimEvidenceQuality`, `DimConcessionDepth`, `DimChallengeSubstantiveness`, `DimSynthesisQuality`, `DimActionability` with the corresponding snake_case string values. The old constants (`DimRigor`, `DimCoverage`, `DimObjectivity`, `DimConvergence`, `DimDepth`) do NOT appear.
- GIVEN `internal/service/quality.go` WHEN `Log()` is called on an existing convergence document that uses the rubric names (e.g., `docs/knowledge/phase-3-review-convergence.md`) THEN all 6 dimensions parse with `Value > 0` and the `Score` field equals the document's stated score.
- GIVEN `go build ./...` WHEN run after the rename THEN build succeeds with zero errors.

### FR-142 (Orchestrate Package Tests)

- GIVEN `internal/orchestrate/preflight_test.go` exists WHEN `go test ./internal/orchestrate/...` runs THEN all tests pass.
- GIVEN a `TypeBugFix` request WHEN `PreflightService.Run()` is called with a valid project state THEN `PreflightResult.RequestType == domain.TypeBugFix` and no error is returned.
- GIVEN a `TypeComplexNew` request with a missing brief WHEN `PreflightService.Run()` is called THEN an error is returned containing "brief gate BLOCKED".
- GIVEN `go tool cover` on `internal/orchestrate/` WHEN measured THEN coverage is ≥ 80%.

### FR-143 (MCP Server Tests)

- GIVEN `internal/mcp/server_test.go` exists WHEN `go test ./internal/mcp/...` runs THEN all tests pass.
- GIVEN a valid `initialize` request WHEN `handleRaw()` is called THEN a response with `result.protocolVersion` is returned.
- GIVEN a `tools/call` request for an unknown tool WHEN `handleRaw()` is called THEN an error response with `errMethodNotFound` code is returned.
- GIVEN a malformed JSON input WHEN `handleRaw()` is called THEN an error response with `errParse` code is returned.
- GIVEN a `notifications/initialized` message WHEN `handleRaw()` is called THEN `nil` is returned (no response).
- GIVEN `go tool cover` on `internal/mcp/` WHEN measured THEN coverage is ≥ 80%.

### FR-144 (Quality Service Tests)

- GIVEN `internal/service/quality_test.go` exists WHEN `go test ./internal/service/...` runs THEN all tests pass.
- GIVEN a convergence markdown sample containing all six rubric dimension names with non-zero scores WHEN `QualityService.Log()` processes it THEN the resulting `QualityEntry` has `len(Dimensions) == 6` and each `Dimension.Value > 0`.
- GIVEN a `QualityEntry` where `Score = 3.5` and `GatePass = true` WHEN `entry.Validate()` is called THEN no error is returned.

### FR-145 (StateRepo Layer Violation Fix)

- GIVEN `cmd/handoff.go` source file WHEN searched for `os.ReadFile` and `os.WriteFile` THEN zero matches appear (for the purpose of reading/writing `current.md`).
- GIVEN `mind handoff <valid-iter-id>` on a project with a `docs/state/current.md` WHEN run THEN "Recent Changes" in `current.md` contains the iteration entry and the operation succeeds.
- GIVEN `go test ./internal/repo/...` WHEN run with the new `StateRepo` method THEN tests pass including the `mem/` implementation.

### FR-146 (HandoffService Extraction)

- GIVEN `internal/orchestrate/preflight.go` WHEN inspected THEN `PreflightService` has no `Handoff()` method and no `findIteration()` method.
- GIVEN `internal/orchestrate/` WHEN inspected THEN a `HandoffService` type exists with an `IterationRepo` constructor dependency.
- GIVEN `cmd/handoff.go` WHEN inspected THEN handoff logic delegates to `HandoffService` methods rather than implementing steps inline.
- GIVEN `mind handoff <valid-iter-id>` WHEN run on a project with an in-progress iteration THEN all 5 handoff steps execute and the output matches the pre-refactor behavior.

### FR-147 (Preflight Doc-Failure Blocking)

- GIVEN a project where `mind check docs` would report `Failed > 0` WHEN `mind preflight "add feature"` is run THEN the command exits non-zero with an error message that includes the failure count and instructs the user to run `mind check docs`.
- GIVEN a project where `mind check docs` reports zero failures but has warnings WHEN `mind preflight "add feature"` is run THEN the command proceeds past step 3 and the warning count appears in `PreflightResult.Warnings`.
- GIVEN a project with fully passing docs WHEN `mind preflight "add feature"` is run THEN step 3 completes with no blocking error.

### FR-148 (Branch Comparison Portability)

- GIVEN `mind.toml` governance section contains `default-branch = "develop"` WHEN `mind handoff <iter-id>` runs THEN `branchAhead()` compares against `develop`, not `main`.
- GIVEN `mind.toml` has no `default-branch` governance setting WHEN `mind handoff <iter-id>` runs THEN `branchAhead()` falls back to `"main"` as the comparison base.
- GIVEN `cmd/handoff.go` source WHEN inspected THEN the string literal `"HEAD...main"` does NOT appear.

---

## Out of Scope

The following are explicitly excluded from this iteration:

- **Phase 4 implementation**: `mind watch`, `mind run`, and any orchestration beyond preflight/handoff/serve.
- **MCP tool additions or removals**: The existing 16 tools registered in `internal/mcp/tools.go` are out of scope. No new tools are added. Tool behavior is not changed except as required by test coverage (FR-143).
- **C-2 (prompt.go filesystem access)**: The `PromptBuilder.readFile()` direct `os.ReadFile` usage in `internal/orchestrate/prompt.go` is accepted as tech debt per the `internal/reconcile/hash.go` architectural precedent. No interface change required.
- **C-3 (duration JSON serialization)**: `domain/gate.go` `time.Duration` serialization as nanoseconds is not addressed. The JSON contract for `GateCommandResult` is unchanged.
- **C-5 (FR-N traceability comments)**: Adding traceability comments to Phase 3 source files is not a required deliverable for this iteration.
- **Historical quality-log.yml data correction**: Entries already written with old dimension names are not retroactively corrected.
- **New `mind` subcommands**: No new commands beyond the existing `preflight`, `handoff`, `serve` surface.
- **TUI quality tab behavior changes**: The quality tab (`tui/quality.go`) rendering behavior is not changed by this iteration. If dimension constant references are updated for the rename, only the constant identifiers change — not the display logic.
- **Shell completions, GoReleaser, CI/CD** (Phase 5 scope).
- **Performance optimization** of any Phase 3 package.
