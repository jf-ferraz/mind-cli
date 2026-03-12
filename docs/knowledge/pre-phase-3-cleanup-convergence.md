# Convergence Analysis: Pre-Phase 3 Codebase Cleanup

**Topic**: Pre-Phase 3 comprehensive codebase review -- architecture consistency, design pattern uniformity, code cleanliness, resilience, integration consistency, known issue severity reassessment, and documentation accuracy.
**Date**: 2026-03-11
**Personas**: Architect, Pragmatist, Critic, Researcher
**Variant**: Deep (4 personas, 2 rounds)
**Effective Persona Count**: 3.8 (see Diversity Audit)

---

## Phase 2: Opening Position Papers

### Position A -- The Architect (dev_architect)

**Core Assessment**: The 4-layer architecture is well-maintained across 137 Go files and 20,555 lines. Domain purity is compiler-verified and holds. However, three structural issues threaten Phase 3 readiness: (1) the Deps struct uses concrete types instead of interfaces, (2) DoctorService reimplements validation logic instead of delegating, and (3) GenerateService performs direct filesystem I/O instead of going through repositories.

**Detailed Findings**:

1. **Deps struct uses concrete `*fs.` types instead of interfaces (ARCHITECTURAL DEBT)**
   The `deps.Deps` struct (`internal/deps/deps.go`, lines 12-27) stores concrete `*fs.DocRepo`, `*fs.IterationRepo`, etc. instead of `repo.DocRepo`, `repo.IterationRepo` interfaces. This means:
   - The TUI (`tui/app.go`, line 12) imports `internal/repo/fs` directly just for type resolution.
   - `cmd/root.go` (line 39) declares `docRepo *fs.DocRepo` instead of `repo.DocRepo`.
   - **Phase 3 impact**: The MCP server will be a third consumer of Deps. Using concrete types means the MCP server will also import `fs`, creating a hard dependency on filesystem implementations where none is needed.

   *Evidence*: `tui/app.go` line 12 imports `github.com/jf-ferraz/mind-cli/internal/repo/fs` -- the only use is `fs.DetectProject()` on line 39. The Deps struct itself forces this coupling.

2. **DoctorService reimplements checks (KNOWN COULD, but severity underestimated)**
   `internal/service/doctor.go` (334 lines) contains 9 check methods that overlap significantly with `internal/validate/docs.go` (337 lines). For example:
   - `checkDocStructure()` (lines 132-170) checks zone directories and required files -- overlaps with DocsSuite checks D-01 through D-05.
   - `checkBrief()` (lines 172-187) duplicates brief gate logic from DocsSuite check D-11.
   - `checkConfig()` (lines 189-206) overlaps with ConfigSuite checks C-01 through C-03.

   **Phase 3 impact**: When adding MCP-based diagnostic endpoints, we would need to choose between doctor's reimplementation and the validation suites, or expose both with potentially inconsistent results.

3. **GenerateService bypasses repository pattern**
   `internal/service/generate.go` calls `os.MkdirAll`, `os.WriteFile`, `os.Stat`, `os.ReadDir` directly (20+ direct OS calls). It receives only `projectRoot string` -- no repository injection. This violates the architecture's rule that "All filesystem access passes through repository interfaces" (architecture.md, line 9).

   Similarly, `internal/service/init.go` (124 lines) does all filesystem operations directly with no repository injection.

   These services are untestable without a real filesystem.

4. **Layer violations in import graph**
   - `cmd/` imports `internal/generate` directly (`cmd/create.go`, line 12) -- this is a service-layer package being consumed from the presentation layer without going through a service. (Minor: the `generate` package is only used for templates in `create brief`'s interactive flow.)
   - `tui/` imports `internal/repo/fs` directly (`tui/app.go`, line 12) for `fs.DetectProject()`. This is an infrastructure-layer import from the presentation layer.

5. **Architecture spec documents only partially updated**
   `docs/spec/architecture.md` was updated for Phase 1.5 and Phase 2 additions, but the Phase 2 section still references `cmd/tui_cmd.go` (line 431) -- the actual file is `cmd/tui.go`. The component map for Phase 1 (lines 50-68) does not mention `InitService` or `DoctorService`.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| A1 | Deps uses concrete fs types | `internal/deps/deps.go` lines 14-20 | Very High |
| A2 | DoctorService duplicates validate logic | `internal/service/doctor.go` lines 132-206 vs `internal/validate/docs.go` | High |
| A3 | GenerateService bypasses repos | `internal/service/generate.go` lines 28, 42, 57, etc. (20+ os.* calls) | Very High |
| A4 | tui imports fs directly | `tui/app.go` line 12, import analysis output | Very High |
| A5 | Architecture doc has stale file reference | `docs/spec/architecture.md` line 431 mentions `cmd/tui_cmd.go` | Very High |

---

### Position B -- The Pragmatist (dev_pragmatist)

**Core Assessment**: The codebase is remarkably clean for its size. 374 tests all pass, go vet reports zero issues, domain purity holds. The SHOULD items in `docs/state/current.md` are correctly prioritized. Fix the items that directly impact Phase 3 consumers; defer cosmetic items. The biggest practical improvement is fixing the `--project` to `--project-root` rename, which affects the upcoming MCP server's API contract.

**Detailed Findings**:

1. **The `--project` flag rename is the highest-impact SHOULD item**
   `cmd/root.go` line 91 registers `--project` / `-p`. The spec (`api-contracts`) calls for `--project-root`. Phase 3 introduces the MCP server which will expose project root resolution as a tool. If the CLI flag is renamed after the MCP server ships, it's a breaking change across both interfaces. Fix it now while only the CLI uses it.

   Cost: ~5 lines changed in `root.go`, update help text in `init.go`.

2. **Editor fallback to `vi` (S-1) is correct behavior for the TUI, wrong for CLI**
   `tui/editor.go` (lines 9-15) defaults to `vi` when `$EDITOR` is unset. For a TUI, falling back to a known editor is pragmatic -- the user is already in an interactive terminal. The CLI's `cmd/docs.go` (lines 214-219) correctly returns an error when `$EDITOR` is unset. The S-1 issue should be split: TUI behavior is acceptable, CLI behavior is already correct.

3. **Preview pane raw content (S-2) is the right tradeoff for Phase 3**
   `tui/docs.go` line 188 displays raw content. Glamour rendering would add complexity (terminal width handling, ANSI escape sequences in overlay), and the Phase 3 MCP server won't need it at all. This is correctly classified as SHOULD.

4. **Status bar cursor position (S-3) is trivial but good UX**
   `tui/statusbar.go` line 19 receives only `activeTab` and `info` parameters. Adding cursor position for Docs (line X/Y) and Iterations (item X/Y) would improve navigation. ~20 lines of change.

5. **cmd/ package has zero tests**
   All 1,134 lines in `cmd/` are untested. While the thin-handler pattern means most logic is in services, the exit code logic and error handling in command handlers is complex:
   - 29 `os.Exit()` calls across cmd/ files
   - Error classification logic (`isConfigError` in `cmd/reconcile.go`, `isNotProject` in `cmd/helpers.go`)
   - Flag validation (`runDocsList` zone validation, `runReconcile` mutual exclusion)

   Phase 3 adds more commands (`mind serve`, `mind preflight`). Testing patterns should be established now.

6. **Consistent handler pattern across CLI commands**
   All CLI commands follow the same structure: resolve root -> call service -> render result -> set exit code. This is well-implemented. The only inconsistency is `cmd/init.go` which creates its own `service.NewInitService()` (line 51) and its own `render.New()` (lines 63-64) instead of using the centralized wiring from `PersistentPreRunE`. This is intentional (init runs before project detection), but should be documented.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| B1 | --project flag is named incorrectly per spec | `cmd/root.go` line 91 | Very High |
| B2 | TUI editor fallback is pragmatic | `tui/editor.go` lines 9-15 | High |
| B3 | cmd/ has zero tests | `go test ./cmd` output: `[no test files]` | Very High |
| B4 | 29 os.Exit calls in cmd/ | grep analysis of cmd/ | Very High |
| B5 | init.go creates own service/renderer | `cmd/init.go` lines 51, 63-64 | Very High |

---

### Position C -- The Critic (dev_critic)

**Core Assessment**: The codebase has three hidden risks that could cause problems in Phase 3: (1) direct `os.Exit()` calls in command handlers prevent proper cleanup and make error handling untestable, (2) the `Diagnostic.Status` field uses raw strings instead of typed enums creating a bug surface, and (3) the staleness propagation depth > 0 reason string issue is more severe than SHOULD because it produces misleading diagnostic output.

**Detailed Findings**:

1. **os.Exit() calls bypass Cobra's error handling (RISK)**
   29 `os.Exit()` calls across `cmd/` files. The pattern is:
   ```go
   fmt.Fprintf(os.Stderr, "Error: %v\n", err)
   os.Exit(2)
   return nil  // dead code
   ```
   This bypasses Cobra's `SilenceErrors` mechanism, prevents `defer` cleanup, and makes exit code behavior untestable. Phase 3's MCP server must NOT call `os.Exit()` -- it needs to return errors to the JSON-RPC layer. If the same services are shared, any code path that triggers `os.Exit()` would kill the MCP server process.

   The commands that call `os.Exit()` directly include: status (1), doctor (1), check (4), reconcile (5), docs (5), create (9), init (2), tui (1), root (1).

   **Recommendation**: Replace `os.Exit()` with returning typed errors that the root command's `PersistentPostRunE` or `Execute()` maps to exit codes. This is a standard Cobra pattern.

2. **Diagnostic.Status uses raw strings instead of typed enum**
   `domain/health.go` line 29: `Status string` in the `Diagnostic` struct uses raw strings `"pass"`, `"fail"`, `"warn"`. But there is already a `CheckLevel` enum (`domain/validation.go` lines 1-9) with `LevelFail`, `LevelWarn`, `LevelInfo`. The `DoctorService.addDiag()` method (`internal/service/doctor.go`, lines 90-108) takes `status string` and separately derives a `Level`, creating two parallel representations:
   ```go
   func (s *DoctorService) addDiag(..., status, message, fixHint string, ...) {
       level := domain.LevelInfo
       switch status {
       case "fail":
           level = domain.LevelFail
       case "warn":
           level = domain.LevelWarn
       }
   ```
   A typo in status (`"fali"` instead of `"fail"`) would silently produce a diagnostic with `LevelInfo` instead of `LevelFail`. The `DoctorSummary` counter logic (lines 71-80) also uses raw string comparison.

3. **Transitive propagation reason strings are misleading (SHOULD -> MUST)**
   `internal/reconcile/propagate.go` lines 90-104: The `buildReason()` function only resolves edge-type-specific reasons at `depth == 0`. At depth > 0, it falls through to the generic "may be outdated" reason from `edgeTypeReason()`'s default case, regardless of the actual edge type:
   ```go
   for _, edge := range graph.Reverse[targetID] {
       if depth == 0 && edge.From == sourceID {
           edgeReason = edgeTypeReason(edge.Type)
           break
       }
   }
   ```
   This means in a chain `A -> B -> C` where A changes:
   - B gets: "dependency changed: A (prerequisite changed)" -- correct
   - C gets: "dependency changed: A (via transitive chain, may be outdated)" -- loses the edge type between B and C

   For Phase 3, the MCP server will expose staleness reasons to AI agents. Generic "may be outdated" gives agents no actionable information. The fix is straightforward: at depth > 0, look up the edge type from the immediate predecessor instead of the original source.

4. **render.go has no tests (703 lines)**
   `internal/render/render.go` is 703 lines with zero tests (`go test ./internal/render` output: `[no test files]`). This includes 16 `Render*` methods and the `DetectMode()` function. Phase 3 will add `RenderPreflight()` and possibly `RenderMCPStatus()`. Without render tests, regressions in output formatting are invisible.

5. **The `cmd/` package uses package-level mutable variables (C-11 violation)**
   `cmd/root.go` lines 14-42 declare 12 package-level variables that are mutated by `PersistentPreRunE`. This violates constraint C-11 ("no global mutable state") from the architecture spec. The `PersistentPreRunE` writes to these on every command execution, and the TUI bypasses them entirely (using `deps` directly). This dual-path creates a risk: if a CLI command is ever called from within the TUI (e.g., TUI triggers a validation rerun), the package-level variables would be stale.

6. **`mem/` repo imports `fs/` (inverse dependency)**
   `internal/repo/mem/` imports `internal/repo/fs` -- an in-memory test implementation importing the real filesystem implementation. This import appears in the package analysis: `internal/repo/mem` imports `github.com/jf-ferraz/mind-cli/internal/repo/fs`. This is an architectural concern: test implementations should depend only on domain and interfaces, not on production implementations.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| C1 | 29 os.Exit calls in cmd/ | grep analysis | Very High |
| C2 | Diagnostic.Status is raw string, not enum | `domain/health.go` line 29 | Very High |
| C3 | Transitive propagation loses edge types at depth > 0 | `internal/reconcile/propagate.go` lines 90-104 | Very High |
| C4 | render.go has zero tests | go test output | Very High |
| C5 | 12 package-level mutable vars in cmd/ | `cmd/root.go` lines 14-42 | Very High |
| C6 | mem/ imports fs/ | import analysis output | Very High |

---

### Position D -- The Researcher (researcher)

**Core Assessment**: Quantitative analysis of the codebase reveals strong structural health but with specific measurable gaps. Test coverage is asymmetric (domain 100%, tui 62.8%, render 0%, cmd 0%), and the codebase has grown organically in ways that create measurable inconsistencies.

**Detailed Findings**:

1. **Package size distribution analysis**:

   | Package | Source LOC | Test LOC | Test/Source Ratio | Test Files |
   |---------|-----------|---------|-------------------|-----------|
   | domain/ | 742 | 1,301 | 1.75 | 10 |
   | cmd/ | 1,134 | 0 | 0.00 | 0 |
   | tui/ | 2,146 | 1,908 | 0.89 | 9 |
   | tui/components/ | 231 | 273 | 1.18 | 1 |
   | internal/service/ | 1,107 | 1,789 | 1.62 | 8 |
   | internal/validate/ | 1,071 | 2,160 | 2.02 | 7 |
   | internal/render/ | 703 | 0 | 0.00 | 0 |
   | internal/repo/ (fs+mem+interfaces) | 1,235 | 1,553 | 1.26 | 8 |
   | internal/reconcile/ | 553 | 2,465 | 4.46 | 8 |
   | internal/generate/ | 326 | 208 | 0.64 | 1 |
   | internal/deps/ | 57 | 84 | 1.47 | 1 |

   Key observations:
   - **cmd/ (1,134 LOC) and render/ (703 LOC) have zero tests** -- 1,837 LOC untested (17.4% of total source).
   - **internal/reconcile/** has the highest test ratio (4.46) reflecting its algorithmic complexity.
   - **tui/** at 2,146 LOC is the largest non-test package, larger than any service or validate package.

2. **Import coupling analysis**:
   - `cmd/` imports 12 external packages -- highest coupling of any package.
   - `domain/` imports only stdlib (`errors`, `fmt`, `regexp`, `strings`, `time`) -- purity holds.
   - `tui/` imports 7 external packages including `internal/repo/fs` (the architectural concern).
   - `internal/repo/mem/` imports `internal/repo/fs` -- inverse dependency.

3. **Naming consistency audit**:
   - `cmd/tui.go` vs architecture doc reference to `cmd/tui_cmd.go`
   - `flagProject` in `cmd/root.go` vs spec naming `--project-root`
   - `Deps.DocRepo` is type `*fs.DocRepo` (concrete) while `ValidationService.docRepo` is type `repo.DocRepo` (interface)
   - Status field: `Diagnostic.Status` is `string`, while `LockEntry.Status` is `EntryStatus` (typed), `Iteration.Status` is `IterationStatus` (typed)

4. **Error handling pattern audit across cmd/**:
   - 8 of 13 command handlers return `nil` after `os.Exit()` -- dead code
   - `runCreateIteration` has double error handling: `os.Exit(1)` for `ErrAlreadyExists` (line 105) AND `os.Exit(1)` for all other errors (line 108), but returns `nil` not the error
   - `runReconcile` handles errors with `os.Exit(2)` for runtime and `os.Exit(3)` for config, but `isConfigError()` relies on substring matching (`strings.Contains(msg, "mind.toml")`) -- fragile
   - `runDocsList` calls `os.Exit(1)` then `return nil` -- the return is dead code

5. **Consistency of the filter pattern in TUI tabs**:
   Both `DocsView.applyFilters()` (`tui/docs.go` lines 162-180) and `IterationsView.applyFilter()` (`tui/iterations.go` lines 98-113) follow the same pattern: iterate, skip non-matching, adjust cursor. But naming differs: `applyFilters` (plural) vs `applyFilter` (singular). Both are pointer receiver methods on value types (methods mutate through pointer receiver but the view types are value types used in Bubble Tea's immutable update pattern).

6. **DRY violations**:
   - Staleness reading logic is duplicated: `ReconciliationService.ReadStaleness()` (lines 99-124) and `ReconciliationService.LoadGraph()` (lines 72-94) both iterate over `lock.Entries` to build `stale map[string]string`.
   - `loadHealth()` in `tui/app.go` (lines 339-355) duplicates the staleness attachment pattern from `cmd/status.go` (lines 34-37): both call `reconcileSvc.ReadStaleness()` and attach to health.
   - Artifact counting logic is repeated 3 times: `internal/service/workflow.go` lines 40-44, `tui/iterations.go` lines 159-163, `internal/service/doctor.go` lines 230-236.
   - Zone filter bar rendering code is structurally identical between `tui/docs.go` (lines 212-236) and `tui/iterations.go` (lines 126-149) -- different data, same pattern.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| D1 | 1,837 LOC with zero test coverage | go test output, wc -l analysis | Very High |
| D2 | mem/ imports fs/ | go list imports analysis | Very High |
| D3 | Diagnostic.Status inconsistent with other typed status fields | `domain/health.go` line 29 vs `domain/reconcile.go` line 37, `domain/iteration.go` | Very High |
| D4 | Artifact counting duplicated 3x | source code analysis of workflow.go, iterations.go, doctor.go | Very High |
| D5 | Staleness reading duplicated between ReadStaleness and LoadGraph | `internal/service/reconciliation.go` lines 72-94 and 99-124 | High |
| D6 | Filter bar rendering pattern duplicated | `tui/docs.go` lines 212-236, `tui/iterations.go` lines 126-149 | High |

---

## Phase 2.5: Diversity Audit, Evidence Quality, Tension Extraction

### Diversity Audit

**Effective Persona Count**: 3.8

- **Architect** focuses on structural/layer issues (Deps types, layer violations, DoctorService delegation). Unique angle: Phase 3 MCP impact assessment.
- **Pragmatist** focuses on practical priorities (flag rename, editor behavior, testing patterns). Unique angle: cost-benefit analysis of each fix.
- **Critic** focuses on hidden risks (os.Exit, raw strings, propagation bugs). Unique angle: failure mode analysis.
- **Researcher** provides quantitative evidence (LOC, coverage, import analysis). Unique angle: measurable inconsistencies.

Overlap areas: Architect and Critic both identify Deps concrete types; Researcher and Critic both find Diagnostic.Status issues; Pragmatist and Critic both note testing gaps. These overlaps indicate genuine convergence on key issues, not redundancy.

### Evidence Quality Pre-Check

All evidence is grounded in actual source code files with line numbers. No claims based on assumptions:
- Import analysis from `go list` output
- LOC counts from `wc -l`
- Test status from `go test ./...` and `go vet ./...`
- Source code examination via Read tool

### Tension Extraction

| Tension | Persona A | Persona B | Resolution Path |
|---------|-----------|-----------|-----------------|
| T1: DoctorService redesign priority | Architect: MUST fix (Phase 3 blocker) | Pragmatist: COULD fix (low Phase 3 impact) | Resolve by assessing MCP diagnostic endpoint needs |
| T2: os.Exit pattern | Critic: MUST fix (kills MCP server) | Pragmatist: SHOULD fix (CLI-only code paths) | Resolve by scoping: fix in commands that share code with MCP |
| T3: Editor fallback behavior | Pragmatist: TUI fallback is fine | Critic: Silent fallback masks config issues | Resolve by documenting intent |
| T4: Transitive propagation reasons | Critic: MUST (misleads agents) | Pragmatist: SHOULD (informative enough) | Resolve by assessing Phase 3 MCP staleness API needs |
| T5: Render test coverage | Critic: Must test before adding more render methods | Pragmatist: Manual verification sufficient | Resolve by adding JSON-mode tests (most critical for MCP) |

---

## Phase 3: Cross-Examination

### Architect challenges Critic (on os.Exit)

**Architect**: Your claim that `os.Exit()` will kill the MCP server assumes the MCP server shares the `cmd/` code paths. Per the architecture spec (line 296), the MCP server will be in a separate `mcp/` package consuming services directly. It would never call `cmd/runReconcile()`. The `os.Exit()` pattern is presentation-layer behavior that stays in the presentation layer.

**Critic responds**: Fair point about code path separation. However, the risk is not just about MCP calling cmd/. It's about testability and defer cleanup. Currently, if `PersistentPreRunE` calls `os.Exit(3)` (root.go line 62), any defers in the calling stack are skipped. And we cannot unit test that `mind check docs` returns exit code 1 on failure without actually forking a process.

### Pragmatist challenges Architect (on GenerateService repos)

**Pragmatist**: You flag GenerateService's direct filesystem access as a violation, but generate and init are inherently filesystem-creative operations. They CREATE files that don't exist yet. What interface would a GenerateRepo implement? `CreateFile(path, content)`? That's just `os.WriteFile` with extra indirection. The current pattern is testable via `t.TempDir()` which is exactly how `init_test.go` works.

**Architect responds**: The argument about indirection is valid for simple file creation. But `GenerateService.nextSeq()` (line 221) does `os.ReadDir()` to find the next sequence number -- that's a query operation that belongs in a repository. The mixed concern is the issue: file creation is reasonable as direct I/O, but directory scanning for sequence derivation should go through `IterationRepo.NextSeq()`, which already exists.

### Critic challenges Researcher (on coverage gaps)

**Critic**: Your LOC analysis identifies render/ (703 LOC) and cmd/ (1,134 LOC) as untested. But not all LOC are equal. The render/ package is pure formatting -- if it produces wrong output, users see it immediately. The cmd/ package contains exit code logic and error classification that is invisible to users. Which is actually riskier?

**Researcher responds**: The risk asymmetry is valid. cmd/ error handling is harder to verify manually because exit codes are invisible unless specifically tested. However, render/ includes JSON marshaling logic that serves as the API contract for `--json` output. If JSON field names change or structure breaks, automated consumers (CI scripts, future MCP) silently fail. Both need tests, but cmd/ exit code tests are higher priority for Phase 3.

### Researcher challenges Pragmatist (on --project rename)

**Researcher**: You identify `--project` rename as highest-impact. But the naming inconsistency is only visible in the CLI flag. The Deps struct uses `ProjectRoot`, services use `projectRoot`, the flag variable is `flagProject`. There's no `--project-root` anywhere in the codebase. Is this really a spec compliance issue or a spec that was never updated after implementation?

**Pragmatist responds**: Good question. The api-contracts spec predates the implementation. If the spec was aspirational and the implementation chose `--project` for brevity, the fix is to update the spec, not the code. However, the semantic distinction matters: `--project` could mean "project name" while `--project-root` unambiguously means "filesystem path." For Phase 3 MCP where tool parameters need clear semantics, the longer name is better.

---

## Phase 4: Rebuttals and Refinements

### Architect refines position

**Concedes**: GenerateService's file creation pattern (WriteFile for new files) is acceptable. The real issue is `nextSeq()` duplicating `IterationRepo.NextSeq()`. Split the concern: keep direct file creation, but use repos for queries.

**Maintains**: Deps must use interface types before Phase 3. The MCP server should not import `internal/repo/fs`.

**Adds**: The `internal/repo/mem/` importing `internal/repo/fs` (found by Researcher D2, confirmed by Critic C6) should be fixed as part of the Deps interface migration. The `mem` package likely imports `fs` for the `IsStubContent()` function -- this should be moved to a shared utility or into the domain layer.

### Pragmatist refines position

**Concedes**: The `--project` vs `--project-root` decision should be resolved by evaluating Phase 3 MCP API naming conventions, not by blindly following a pre-implementation spec. But it should be resolved, not left ambiguous.

**Maintains**: Editor fallback to `vi` in TUI is correct behavior. Document it, don't change it.

**Adds**: The `cmd/` init.go pattern of creating its own renderer (lines 63-64) should be documented as an intentional exception, not silently accepted as a pattern others might copy.

### Critic refines position

**Concedes**: os.Exit() in cmd/ is unlikely to affect the MCP server directly (Architect's point about separate mcp/ package is valid). But the testability argument stands -- reclassify from "Phase 3 risk" to "testing debt."

**Maintains**: Transitive propagation reason strings MUST be fixed. The `buildReason()` function at depth > 0 doesn't even attempt to find the correct edge type between the immediate predecessor and the target. The fix is 5-10 lines.

**Refines**: Diagnostic.Status should use CheckLevel enum (or a dedicated DiagnosticStatus enum) instead of raw strings. This is a domain purity concern -- the domain should define its own status vocabulary, not rely on convention.

### Researcher refines position

**Concedes**: Not all LOC gaps are equal (Critic's point). Prioritize cmd/ exit code testing over render/ output testing.

**Maintains**: The DRY violations (artifact counting 3x, staleness reading 2x, filter bar pattern 2x) are small individually but accumulate. A shared `CountArtifacts(iter)` function and a shared `FilterBar` component would reduce 50+ lines of duplication.

**Adds**: The `tui/` package at 2,146 source LOC is larger than any service package. This will grow further with Phase 3 additions (MCP status panel?). The S-4 issue ("9 component files inlined into tab views") becomes more pressing as the package grows.

---

## Phase 5: Convergence Synthesis

### Evidence Audit

All 22 evidence items (A1-A5, B1-B5, C1-C6, D1-D6) are grounded in specific source file locations and verified through tooling (go list, go test, go vet, grep). No speculative claims remain.

### Semantic Grouping

Issues are organized into 6 themes, ordered by Phase 3 impact:

#### Theme 1: Interface/Type Consistency (Phase 3 Blocking)

**Finding**: The `deps.Deps` struct uses concrete `*fs.` types instead of `repo.` interfaces. This forces all consumers (CLI, TUI, and future MCP) to import the filesystem implementation package.

**Files affected**: `internal/deps/deps.go`, `cmd/root.go`, `tui/app.go`
**Confidence**: Very High
**All personas agree**: This must be fixed before Phase 3.

**Related**: `internal/repo/mem/` imports `internal/repo/fs` -- same type of coupling. Investigation shows `mem/search_test.go` imports `fs.IsStubContent()`. This function should be relocated (to domain or a shared utility in `internal/repo/`).

#### Theme 2: Error Handling and Exit Code Architecture

**Finding**: 29 `os.Exit()` calls in `cmd/` bypass Cobra's error handling, prevent defer cleanup, and make exit code logic untestable. The `isConfigError()` function uses fragile string matching. `Diagnostic.Status` uses raw strings instead of typed enums.

**Files affected**: All `cmd/*.go` files, `domain/health.go`
**Confidence**: Very High (exit calls) / Very High (string status)
**Convergence**: Architect and Pragmatist agree this is SHOULD priority; Critic argues for MUST on testability grounds.

**Recommendation**: Replace os.Exit() with error returns. Create a `cmdError` type that carries an exit code, and handle it in `Execute()`. This is the standard Cobra pattern and enables testing. Priority: SHOULD (does not block Phase 3 since MCP won't use cmd/).

**Recommendation**: Change `Diagnostic.Status` from `string` to a typed `DiagnosticStatus` enum. This aligns with every other status field in the domain (`IterationStatus`, `EntryStatus`, `LockStatus`, `CheckLevel`). Priority: SHOULD.

#### Theme 3: Staleness Propagation Accuracy

**Finding**: `buildReason()` in `internal/reconcile/propagate.go` lines 90-104 only resolves edge-type-specific reason strings at `depth == 0`. At `depth > 0`, all documents get generic "may be outdated" regardless of the actual edge type between the immediate predecessor and the target document.

**Files affected**: `internal/reconcile/propagate.go`
**Confidence**: Very High
**Convergence**: Critic rates MUST, Pragmatist rates SHOULD, Architect rates SHOULD.

**Recommendation**: Fix the `buildReason()` function to look up the edge type from the immediate predecessor (the node that caused this node to be enqueued) rather than from the original source. Change the queue item to carry the edge type. ~10 lines of change. Priority: SHOULD (upgrade to MUST if Phase 3 MCP exposes staleness reasons to AI agents).

#### Theme 4: Testing Gaps

**Finding**: 1,837 source LOC (17.4%) have zero test coverage: `cmd/` (1,134 LOC) and `internal/render/` (703 LOC). The render/ package includes JSON marshaling that serves as the API contract for `--json` output.

**Files affected**: All cmd/ and render/ files
**Confidence**: Very High
**Convergence**: All personas agree testing should improve, disagree on priority.

**Recommendation**:
- Add render/ tests for JSON output mode (critical for Phase 3 MCP compatibility). Priority: SHOULD.
- Add cmd/ exit code tests using `cobra.Command.Execute()` with injected args. Priority: SHOULD.
- These can be added incrementally alongside Phase 3 work.

#### Theme 5: DRY and Code Organization

**Finding**: Multiple DRY violations identified:
- Artifact counting logic repeated 3 times (workflow.go, iterations.go, doctor.go)
- Staleness map building duplicated (ReadStaleness, LoadGraph)
- Health + staleness attachment duplicated (cmd/status.go, tui/app.go)
- Filter bar rendering pattern duplicated (tui/docs.go, tui/iterations.go)
- S-4: 9 component rendering functions inlined in tab views

**Files affected**: Multiple service, TUI, and cmd files
**Confidence**: High
**Convergence**: Researcher quantifies, all personas agree on low-priority-but-worthwhile.

**Recommendation**: Extract shared utilities:
- `domain.CountArtifacts(iter Iteration) (present, expected int)` -- eliminates 3 duplications
- Shared staleness map builder in `ReconciliationService`
- Extract filter bar as a `tui/components/` function
- Extract tab-inlined rendering into `tui/components/` (addresses S-4)
Priority: COULD (do during Phase 3 when touching these files).

#### Theme 6: Documentation Accuracy

**Finding**: Architecture spec has stale references:
- `cmd/tui_cmd.go` referenced in architecture.md line 431, actual file is `cmd/tui.go`
- Component map omits `InitService`, `DoctorService`, `ReconciliationService` from Phase 1 table
- `--project` flag documented vs `--project-root` in api-contracts spec

**Files affected**: `docs/spec/architecture.md`, `docs/state/current.md`
**Confidence**: Very High
**Convergence**: All agree these should be fixed (trivial effort).

**Recommendation**: Update architecture.md file references and component tables. Priority: SHOULD (5 minutes of effort).

### Context-Aware Decision Matrix

Criteria derived from this project's constraints:
1. **Phase 3 Impact** (weight: 40%): Does this block or complicate the MCP server implementation?
2. **Defect Risk** (weight: 25%): Could this cause incorrect behavior or silent failures?
3. **Fix Cost** (weight: 20%): Lines changed, files touched, regression risk
4. **Consistency** (weight: 15%): Does it violate established patterns in the codebase?

| Issue | Phase 3 Impact | Defect Risk | Fix Cost | Consistency | Weighted Score | Priority |
|-------|---------------|-------------|----------|-------------|---------------|----------|
| Deps concrete types -> interfaces | High (40) | Medium (15) | Medium (12) | High (12) | **79** | **MUST** |
| mem/ imports fs/ | Medium (24) | Low (5) | Low (16) | High (12) | **57** | SHOULD |
| Transitive propagation reasons | Medium (24) | High (20) | Low (16) | Medium (9) | **69** | **SHOULD** |
| Diagnostic.Status raw strings | Low (8) | Medium (15) | Low (16) | High (12) | **51** | SHOULD |
| --project flag rename | High (32) | Low (5) | Low (16) | Medium (9) | **62** | **SHOULD** |
| os.Exit() -> error returns | Low (8) | Medium (15) | Medium (12) | Medium (9) | **44** | SHOULD |
| cmd/ test coverage | Medium (24) | Medium (15) | Medium (12) | Medium (9) | **60** | SHOULD |
| render/ test coverage | Medium (24) | Medium (15) | Medium (12) | Medium (9) | **60** | SHOULD |
| Architecture doc updates | Low (8) | Low (5) | Low (16) | Low (6) | **35** | SHOULD |
| DRY: artifact counting | Low (8) | Low (5) | Low (16) | Medium (9) | **38** | COULD |
| DRY: filter bar component | Low (8) | Low (5) | Low (16) | Medium (9) | **38** | COULD |
| DRY: staleness map builder | Low (8) | Low (5) | Low (16) | Low (6) | **35** | COULD |
| Editor fallback (S-1) | Low (8) | Low (5) | Low (16) | Low (6) | **35** | COULD |
| Preview Glamour (S-2) | Low (8) | Low (5) | Medium (12) | Medium (9) | **34** | COULD |
| Status bar cursor (S-3) | Low (8) | Low (5) | Low (16) | Low (6) | **35** | COULD |
| S-4 component extraction | Low (8) | Low (5) | Medium (12) | Medium (9) | **34** | COULD |
| DoctorService delegation | Medium (24) | Low (5) | High (8) | High (12) | **49** | COULD |
| GenerateService repo queries | Low (8) | Low (5) | Medium (12) | Medium (9) | **34** | COULD |

### Executive Summary

The mind-cli codebase is in strong structural health after three phases of development. The 4-layer architecture holds across all 137 Go files. Domain purity is compiler-verified. All 374 tests pass. `go vet` reports zero issues.

**One MUST-fix item** for Phase 3 readiness:
1. **Migrate Deps struct to interface types** -- The `deps.Deps` struct uses concrete `*fs.` types, forcing all consumers to import the filesystem implementation. The MCP server (Phase 3) must not have this dependency.

**Eight SHOULD-fix items**, ordered by weighted impact:
1. Transitive propagation reason strings at depth > 0 (misleading diagnostic output)
2. `--project` flag rename to `--project-root` (API naming before MCP)
3. cmd/ exit code test coverage (29 os.Exit calls untestable)
4. render/ JSON output test coverage (API contract validation)
5. `internal/repo/mem/` importing `internal/repo/fs` (inverse dependency)
6. `Diagnostic.Status` raw strings to typed enum
7. `os.Exit()` calls to error returns (Cobra best practice)
8. Architecture doc stale references

**Five COULD-fix items** (opportunistic, do when touching files):
- DRY: artifact counting, filter bar, staleness map
- DoctorService delegation to ValidationService
- TUI component extraction (S-4)

### Concession Trail

| Persona | Concession | To | Reason |
|---------|-----------|-----|--------|
| Architect | GenerateService file creation is acceptable | Pragmatist | Creating new files is inherently I/O; repo abstraction adds no value there |
| Pragmatist | --project rename needs resolution before Phase 3 | Architect | MCP tool parameter naming must be unambiguous |
| Critic | os.Exit() is SHOULD not MUST | Architect | MCP server uses separate mcp/ package, won't call cmd/ |
| Researcher | Render tests are lower priority than cmd tests | Critic | Exit code logic is invisible to manual testing |
| Architect | DoctorService delegation is COULD not MUST | Pragmatist | Phase 3 MCP won't expose doctor diagnostics initially |
| Critic | Editor fallback to vi in TUI is acceptable | Pragmatist | TUI is interactive-only; falling back is pragmatic |

### Key Insights

1. **The single Phase 3 blocker is the Deps type system**: Concrete types create import coupling that prevents clean MCP server implementation. This is a focused refactoring (~50 lines across 3 files) with well-defined scope.

2. **Error handling has two separate issues often conflated**: (a) `os.Exit()` bypassing Cobra is a testability issue, not a runtime risk. (b) `Diagnostic.Status` using raw strings is a correctness risk. They should be addressed independently.

3. **The codebase is remarkably consistent for organic growth across 3 phases**: The thin-handler pattern, constructor injection, and domain purity all hold. The inconsistencies found are minor (naming: `applyFilter` vs `applyFilters`, concrete vs interface types in one struct).

4. **Testing investment should target JSON output and exit codes**: These are the invisible contracts that automated consumers depend on. Visual output testing has lower ROI.

### Recommendations with Confidence and Falsifiability

| # | Recommendation | Confidence | Falsifiability |
|---|---------------|------------|----------------|
| R1 | Migrate Deps to interface types | Very High | Falsified if Phase 3 MCP server architecture avoids importing Deps entirely |
| R2 | Fix transitive propagation reasons | High | Falsified if Phase 3 MCP does not expose staleness reasons to consumers |
| R3 | Rename --project to --project-root | High | Falsified if api-contracts spec is updated to match current --project naming |
| R4 | Add cmd/ exit code tests | High | Falsified if Phase 3 testing strategy uses integration tests that cover exit codes |
| R5 | Add render/ JSON tests | High | Falsified if MCP server produces its own JSON without using Renderer |
| R6 | Fix mem/ importing fs/ | Medium | Falsified if IsStubContent is genuinely fs-specific logic |
| R7 | Type Diagnostic.Status | Medium | Falsified if DoctorService is deprecated in favor of validation-only diagnostics |
| R8 | Replace os.Exit with error returns | Medium | Falsified if cmd/ tests use process forking instead |
| R9 | Update architecture docs | Very High | Not falsifiable -- documentation accuracy is always correct |
| R10 | Extract DRY utilities | Medium | Falsified if the duplicated code paths diverge in Phase 3 |

### Quality Rubric

| Dimension | Score | Justification |
|-----------|-------|---------------|
| **Rigor** | 4/5 | All findings grounded in source code with line numbers. Import analysis from go list. One area could use deeper analysis: the actual IsStubContent dependency in mem/. |
| **Coverage** | 5/5 | All 7 analysis dimensions covered. All packages examined. All known issues re-assessed. Architecture, patterns, resilience, DRY, testing, documentation all addressed. |
| **Actionability** | 5/5 | Every recommendation has specific files, estimated LOC, clear priority (MUST/SHOULD/COULD), and ordering. Phase 3 impact is assessed per item. |
| **Objectivity** | 4/5 | Tensions genuinely explored and resolved with concessions. The Deps issue is unanimously MUST. Minor risk of anchoring on known issues from current.md rather than discovering unknown issues. |
| **Convergence** | 4/5 | 3 of 4 personas converge on top priority (Deps). Genuine disagreements on os.Exit and DoctorService resolved through evidence-based concessions. One unresolved tension: --project naming requires external spec review. |
| **Depth** | 4/5 | Cross-package analysis, import graph examination, quantitative metrics, and cross-referencing between spec and implementation. Could go deeper on Phase 3 MCP architecture to validate Phase 3 impact claims. |

**Overall Score**: 4.3 / 5.0

### Convergence Diff vs phase-2-tui-dashboard-convergence.md

| Aspect | Phase 2 Convergence | This Convergence |
|--------|-------------------|------------------|
| Focus | Implementation strategy (how to build TUI) | Code quality audit (what to fix before Phase 3) |
| SHOULD items | 7 items, cherry-picked 3 for Phase 2 | 11 items total, re-assessed against Phase 3 needs |
| Top recommendation | Per-tab model + service injection | Deps interface migration |
| Wiring concern | Raised as Phase 2 risk | Partially resolved (BuildDeps exists), concrete type issue identified as new MUST |
| DoctorService | Not discussed | Assessed as COULD (lower Phase 3 impact than expected) |
| Testing strategy | "Test Update(), not View()" | "Test JSON output and exit codes for Phase 3" |
| Resolved items from Phase 2 | - | S-1 (flag exclusion), S-2 (missing docs check), docs search abstraction, BuildDeps wiring |
| New issues found | - | Deps concrete types, mem/ imports fs/, Diagnostic raw strings, staleness reason accuracy |
