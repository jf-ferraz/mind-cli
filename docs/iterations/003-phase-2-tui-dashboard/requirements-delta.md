# Requirements Delta: Phase 2 TUI Dashboard

- **Iteration**: 003-phase-2-tui-dashboard
- **Date**: 2026-03-11
- **Author**: Analyst (Claude Opus 4.6)
- **Convergence Input**: docs/knowledge/phase-2-tui-dashboard-convergence.md (4.00/5.0, PASS)

---

## Current State

Phase 1 (Core CLI) delivers 50 functional requirements (FR-1 through FR-50) covering project detection, output modes, 20+ commands with `--json` support, 17-check doc validation, 11-check ref validation, config validation, document scaffolding, and workflow state inspection.

Phase 1.5 (Reconciliation Engine) delivers 37 functional requirements (FR-51 through FR-87) covering SHA-256 hash tracking, dependency graph, staleness propagation, `mind.lock` lifecycle, and integration with `status`, `check all`, and `doctor`.

The codebase has 246 passing tests with domain/ at 100% coverage and validate/ at 90.7%. Four SHOULD-level findings remain open from Phase 1.5 review (S-1, S-2, wiring centralization, docs search abstraction).

## Desired State

Phase 2 adds the `mind tui` command -- a full-screen interactive Bubble Tea dashboard with 5 tabs (Status, Documents, Iterations, Checks, Quality). The TUI is a new presentation layer that consumes existing service interfaces without requiring new business logic.

Additionally, 4 critical SHOULD items are resolved before TUI implementation to ensure architectural cleanliness and data correctness in the TUI.

## Scope Boundary

### In-scope (Phase 2)

- Fix 4 SHOULD items: S-1 (flag exclusion), S-2 (missing docs check), wiring centralization, docs search abstraction
- `mind tui` command launching full-screen Bubble Tea application
- App shell: title bar, tab bar, status bar, global key handling
- Tab 1 (Status): zone health bars, staleness panel, workflow state, warnings, suggestions, quick actions
- Tab 2 (Documents): zone filter, search, document list with tree structure, preview pane (Glamour)
- Tab 3 (Iterations): type filter, iteration table, detail expander with artifact list
- Tab 4 (Checks): live validation, accordion suites, check detail, re-run, spinner
- Tab 5 (Quality): score chart or empty state, latest analysis detail
- Help overlay with context-sensitive keybindings
- Responsive layout (80x24 minimum, standard/wide modes)
- Clean quit with terminal state restoration
- Tab state preservation across tab switches
- Loading states for async data fetching
- New domain types: TabID, QualityEntry, QualityDimension

### Out-of-scope

- Watch TUI (`mind watch --tui`) -- Phase 3/4 per BP-05 Section 7
- Orchestration TUI (`mind run --tui`) -- Phase 3/4 per BP-05 Section 8
- MCP server (`mind serve`) -- Phase 3
- Monochrome fallback -- SHOULD enhancement within Phase 2, not gate requirement
- 3-tier responsive breakpoints (narrow < 80) -- SHOULD enhancement, not gate requirement
- Auto-polling / filesystem watching within TUI -- explicit design decision per BP-05 Section 4
- Mouse support
- QualityService implementation if not yet present (Tab 5 launches with empty state)
- S-3 (transitive reason strings) -- deferred per convergence recommendation
- `--project` flag rename -- deferred per convergence recommendation
- GoDoc gaps on existing methods -- deferred per convergence recommendation
- Shell completions -- Phase 5

---

## New Requirements

### SHOULD Fixes (Pre-TUI Prerequisites)

Per convergence recommendation R1 (90% HIGH confidence), these 4 items MUST be resolved before TUI implementation.

#### FR-88: --check and --force Mutual Exclusion [MUST]

`mind reconcile` MUST reject the combination of `--check` and `--force` flags with an error message and exit code 2.

**Acceptance Criteria:**

- GIVEN `mind reconcile --check --force` WHEN the command is invoked THEN it MUST exit with code 2 and print "Error: --check and --force are mutually exclusive" to stderr without modifying `mind.lock`.
- GIVEN `mind reconcile --check` alone WHEN invoked THEN behavior is unchanged from FR-52.
- GIVEN `mind reconcile --force` alone WHEN invoked THEN behavior is unchanged from FR-53.

**Traces**: S-1 from iteration 002 validation.md, api-contracts spec line 768.

---

#### FR-89: Missing Documents Check in ReconcileSuite [MUST]

`ReconcileSuite` MUST include a check for documents declared in `mind.toml [documents]` that are missing from disk. Each missing document MUST produce a WARN-level `CheckResult` entry in the reconcile suite.

**Acceptance Criteria:**

- GIVEN 5 documents declared in `mind.toml` with 1 missing from disk WHEN `mind check all` is run THEN the reconcile suite contains a WARN-level check result with message "declared in mind.toml but not found on disk: {document ID}" and the overall reconcile suite shows 1 warning.
- GIVEN 0 missing documents WHEN `mind check all` is run THEN no missing-document check results appear in the reconcile suite.
- GIVEN 2 missing documents and `--strict` WHEN `mind check all --strict` is run THEN exit code is 1 (WARN promoted to FAIL).

**Traces**: S-2 from iteration 002 validation.md, FR-79 specification.

---

#### FR-90: Wiring Centralization via buildDeps Pattern [MUST]

All repository and service construction MUST be centralized in a `buildDeps()` function (or equivalent) in `main.go` that is callable from both the CLI path (`PersistentPreRunE`) and the TUI initialization path. Command handlers MUST NOT construct repositories or services directly.

**Acceptance Criteria:**

- GIVEN the `mind tui` command WHEN it initializes THEN it receives all required services (ProjectService, ValidationService, ReconciliationService, WorkflowService, DocRepo, IterationRepo) through the same wiring function used by CLI commands.
- GIVEN a new command handler WHEN it needs a service THEN it accesses the service from a shared dependencies struct, not by constructing repositories inline.
- GIVEN `cmd/status.go` and `cmd/doctor.go` WHEN inspected THEN neither contains direct calls to `fs.NewDocRepo()`, `fs.NewConfigRepo()`, or similar constructors.

**Traces**: C-10 deviation from Phase 1 validation, wiring centralization acknowledged in Phase 1.5 architecture-delta.md.

---

#### FR-91: docs search DocRepo Abstraction [MUST]

`mind docs search` MUST perform file discovery and content reading through the `DocRepo` interface instead of direct `filepath.WalkDir` calls. This ensures the TUI Documents tab can reuse the same search logic through the service layer.

**Acceptance Criteria:**

- GIVEN `mind docs search "authentication"` WHEN run THEN results are identical to the current behavior but the implementation delegates to DocRepo for file discovery and content reading.
- GIVEN an in-memory DocRepo with test documents WHEN the search function is invoked THEN it returns matching results without filesystem access.
- GIVEN the Docs tab in the TUI WHEN the user presses `/` and types a search query THEN the same DocRepo-based search is used.

**Traces**: C-9 deviation noted in docs/state/current.md, convergence recommendation R1 item 4.

---

### TUI Command

#### FR-92: mind tui Command [MUST]

`mind tui` MUST launch a full-screen interactive Bubble Tea application that takes over the terminal. The application MUST initialize by loading project data through existing service interfaces and render a 5-tab dashboard.

**Acceptance Criteria:**

- GIVEN a valid Mind project WHEN `mind tui` is run in a terminal THEN a full-screen application launches with title bar, tab bar, content area, and status bar.
- GIVEN `mind tui` is running WHEN the user presses `q` THEN the application exits cleanly, restoring the terminal to its pre-launch state (cursor visible, input echoing, no alternate screen artifacts).
- GIVEN `mind tui` is running WHEN the user presses `Ctrl+C` THEN the application force-quits cleanly with terminal state restored.
- GIVEN a non-project directory WHEN `mind tui` is run THEN it exits with code 3 and message "not a Mind project".

---

### App Shell

#### FR-93: Title Bar [MUST]

The TUI MUST render a title bar showing: "Mind Framework" label, project name from `mind.toml`, current git branch (if available), and framework version. The title bar MUST span the full terminal width.

**Acceptance Criteria:**

- GIVEN project name "mind-cli" on branch "main" with version "v2026-03-09" WHEN the TUI renders THEN the title bar shows all four pieces of information.
- GIVEN no git repository (not in a git repo) WHEN the TUI renders THEN the branch field is omitted or shows "no branch".

---

#### FR-94: Tab Bar with Navigation [MUST]

The TUI MUST render a tab bar with 5 tabs labeled "[1 Status]", "[2 Docs]", "[3 Iterations]", "[4 Check]", "[5 Quality]". The active tab MUST be visually distinct (bold+underline). Tabs MUST be switchable via number keys (1-5), Tab (next, wrapping), and Shift+Tab (previous, wrapping).

**Acceptance Criteria:**

- GIVEN the TUI is on Tab 1 WHEN the user presses `3` THEN Tab 3 (Iterations) becomes active and its content renders.
- GIVEN the TUI is on Tab 5 WHEN the user presses `Tab` THEN Tab 1 (Status) becomes active (wrap).
- GIVEN the TUI is on Tab 1 WHEN the user presses `Shift+Tab` THEN Tab 5 (Quality) becomes active (wrap).
- GIVEN any tab is active WHEN the tab bar renders THEN the active tab is bold+underline and inactive tabs are dim.

---

#### FR-95: Global Key Bindings [MUST]

The following keys MUST work in every tab and MUST NOT be overridden by tab-specific handlers:

| Key | Action |
|-----|--------|
| `1`-`5` | Switch to corresponding tab |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `q` | Quit (when no modal overlay is open) |
| `Ctrl+C` | Force quit (always) |
| `?` | Toggle help overlay |
| `r` | Refresh all data (when no text input is focused) |

**Acceptance Criteria:**

- GIVEN the TUI is on Tab 2 with a document selected WHEN the user presses `?` THEN the help overlay appears.
- GIVEN the help overlay is open WHEN the user presses `q` THEN the overlay closes (does not quit the application).
- GIVEN the search input is focused on Tab 2 WHEN the user presses `r` THEN the character `r` is typed into the search field (not interpreted as refresh).
- GIVEN the TUI is on any tab with no overlay WHEN the user presses `q` THEN the application quits.

---

#### FR-96: Status Bar [MUST]

The TUI MUST render a bottom status bar showing context-sensitive key hints for the active tab and cursor position information where applicable.

**Acceptance Criteria:**

- GIVEN Tab 2 (Docs) is active with 22 documents and cursor on row 3 WHEN the status bar renders THEN it shows available keys and "3/22 docs".
- GIVEN Tab 1 (Status) is active WHEN the status bar renders THEN it shows quick action keys relevant to the Status tab.

---

### Tab 1: Status

#### FR-97: Status Tab Zone Health Bars [MUST]

Tab 1 MUST display zone health as progress bars for each of the 5 documentation zones (spec, blueprints, state, iterations, knowledge). Each bar MUST show the zone label, a visual progress indicator, and a numeric fraction (present/total).

**Acceptance Criteria:**

- GIVEN a project with 4/5 spec documents and 3/3 blueprints WHEN Tab 1 renders THEN it shows progress bars with the correct fractions for each zone.
- GIVEN a zone with 0 documents WHEN Tab 1 renders THEN the progress bar shows 0/0 or an empty state indicator.

---

#### FR-98: Status Tab Staleness Panel [SHOULD]

Tab 1 SHOULD display a staleness section when `mind.lock` exists and stale documents are detected. The section MUST list stale document names with `●` bullet markers.

**Acceptance Criteria:**

- GIVEN `mind.lock` with 2 stale documents WHEN Tab 1 renders THEN a "Staleness" section appears listing both documents.
- GIVEN no `mind.lock` exists WHEN Tab 1 renders THEN no staleness section appears.
- GIVEN `mind.lock` with 0 stale documents WHEN Tab 1 renders THEN no staleness section appears.

---

#### FR-99: Status Tab Workflow Panel [MUST]

Tab 1 MUST display the current workflow state. When a workflow is active, it MUST show: state (running), type, current agent position in the chain, and branch name. When idle, it MUST show "State: idle" and the last completed iteration if any.

**Acceptance Criteria:**

- GIVEN an active workflow with type NEW_PROJECT and last agent "architect" WHEN Tab 1 renders THEN the workflow panel shows running state with agent chain progress.
- GIVEN idle workflow and last iteration "006-ENHANCEMENT-add-caching" WHEN Tab 1 renders THEN the workflow panel shows "State: idle" and the last iteration summary.

---

#### FR-100: Status Tab Warnings and Suggestions [MUST]

Tab 1 MUST display warnings (with `⚠` prefix) and suggestions (with `→` prefix) from `ProjectHealth`. Each section MUST appear only when it has content.

**Acceptance Criteria:**

- GIVEN ProjectHealth with 2 warnings and 1 suggestion WHEN Tab 1 renders THEN both sections appear with their items.
- GIVEN ProjectHealth with 0 warnings WHEN Tab 1 renders THEN the warnings section is omitted entirely.

---

#### FR-101: Status Tab Two-Column Layout [MUST]

Tab 1 MUST use a two-column layout at terminal widths >= 80 columns. Left column: documentation health, staleness, warnings, suggestions. Right column: workflow state, quick actions. At widths < 80, columns SHOULD stack vertically.

**Acceptance Criteria:**

- GIVEN terminal width of 100 columns WHEN Tab 1 renders THEN left and right columns are side by side, each occupying approximately 50% of the content width.
- GIVEN terminal width of 78 columns WHEN Tab 1 renders THEN content stacks vertically (single column).

---

### Tab 2: Documents

#### FR-102: Documents Tab Document List [MUST]

Tab 2 MUST display all documents from `docs/` in a navigable list grouped by zone. Each entry MUST show: filename, status indicator (`✓` for content, `✗` for stub), status label, modification date, and file size. The cursor row MUST be visually highlighted.

**Acceptance Criteria:**

- GIVEN a project with 18 documents across 5 zones WHEN Tab 2 renders THEN all documents appear grouped under zone headers with status, date, and size columns.
- GIVEN the user presses `↓` or `j` WHEN on Tab 2 THEN the cursor moves to the next document row.
- GIVEN a stub document WHEN Tab 2 renders THEN it shows `✗` indicator and the row is styled in the error color.

---

#### FR-103: Documents Tab Zone Filter [MUST]

Tab 2 MUST provide a zone filter bar allowing the user to filter documents by zone. Filter shortcuts: `a` (all), `s` (spec), `b` (blueprints), `t` (state), `i` (iterations), `k` (knowledge -- when search is not focused). The active filter MUST be visually indicated.

**Acceptance Criteria:**

- GIVEN Tab 2 showing all documents WHEN the user presses `s` THEN only spec zone documents are displayed.
- GIVEN spec filter is active WHEN the user presses `a` THEN all zones are shown again.
- GIVEN a zone filter is active WHEN the tab bar renders THEN the active zone is indicated with brackets and bold styling.

---

#### FR-104: Documents Tab Search [MUST]

Tab 2 MUST support inline search activated by pressing `/`. When active, a text input cursor appears and the document list filters in real time by case-insensitive substring matching on filenames. `Esc` clears the search and restores the full list.

**Acceptance Criteria:**

- GIVEN Tab 2 with no search active WHEN the user presses `/` THEN the search input becomes focused with a visible cursor.
- GIVEN search input "brief" is active WHEN Tab 2 renders THEN only documents containing "brief" in their filename are displayed.
- GIVEN an active search WHEN the user presses `Esc` THEN the search is cleared and the full document list is restored.

---

#### FR-105: Documents Tab Preview Pane [SHOULD]

Tab 2 SHOULD support a preview pane activated by pressing `Enter` on a selected document. The preview MUST render the document content as markdown (via Glamour). When the preview is open, the document list shrinks to 40% width and the preview takes 60%.

**Acceptance Criteria:**

- GIVEN a document is selected on Tab 2 WHEN the user presses `Enter` THEN a preview pane opens showing rendered markdown content.
- GIVEN the preview pane is open WHEN the user presses `Esc` THEN the preview closes and the document list returns to full width.
- GIVEN the preview pane is open WHEN the user presses `↑`/`↓` or `PgUp`/`PgDn` in the preview THEN the preview content scrolls.

---

#### FR-106: Documents Tab Edit Action [SHOULD]

Tab 2 SHOULD support opening the selected document in `$EDITOR` by pressing `e`. The TUI MUST suspend (release the terminal), launch the editor, and resume when the editor exits.

**Acceptance Criteria:**

- GIVEN `$EDITOR` is set to "vim" and a document is selected WHEN the user presses `e` THEN vim opens with the selected document, and the TUI resumes after vim exits.
- GIVEN `$EDITOR` is not set WHEN the user presses `e` THEN an error message appears in the status bar suggesting to set `$EDITOR`.

---

### Tab 3: Iterations

#### FR-107: Iterations Tab Table [MUST]

Tab 3 MUST display iterations in a table with columns: sequence number (#), type, name, status, date, and artifact completeness (e.g., "5/5"). Rows MUST be navigable with `↑`/`↓` or `j`/`k`. The selected row MUST be highlighted. Status indicators: `✓` complete (green), `▸` in_progress (yellow), `○` incomplete (dim).

**Acceptance Criteria:**

- GIVEN 6 iterations exist WHEN Tab 3 renders THEN all 6 appear in a table with correct columns and data.
- GIVEN iteration 004 has 4/5 artifacts WHEN Tab 3 renders THEN its Files column shows "4/5" and status shows `○ incomplete`.
- GIVEN the user navigates with `j`/`k` WHEN on Tab 3 THEN the cursor moves between iteration rows.

---

#### FR-108: Iterations Tab Type Filter [MUST]

Tab 3 MUST provide a type filter bar: `a` (all), `n` (NEW_PROJECT), `e` (ENHANCEMENT), `b` (BUG_FIX), `r` (REFACTOR). The active filter MUST be visually indicated. Type column entries MUST be color-coded per type.

**Acceptance Criteria:**

- GIVEN Tab 3 showing all iterations WHEN the user presses `e` THEN only ENHANCEMENT iterations are displayed.
- GIVEN ENHANCEMENT filter is active WHEN the user presses `a` THEN all types are shown.

---

#### FR-109: Iterations Tab Detail Expander [SHOULD]

Tab 3 SHOULD support expanding a selected iteration by pressing `Enter` to show inline artifact details: each artifact's name, existence status, and file size. The expanded view MUST show within the table row below the selected iteration.

**Acceptance Criteria:**

- GIVEN iteration 006 is selected WHEN the user presses `Enter` THEN an inline detail view expands below the row showing all 5 artifact names with their status and size.
- GIVEN an expanded iteration WHEN the user presses `Enter` again THEN the detail view collapses.

---

### Tab 4: Checks

#### FR-110: Checks Tab Accordion Suites [MUST]

Tab 4 MUST display validation results as an accordion with one section per validation suite (docs, refs, config, reconcile). Each section header MUST show: expand/collapse indicator (`▾`/`▸`), suite name, check count, and pass/fail/warn summary. Sections MUST be expandable via `Enter`.

**Acceptance Criteria:**

- GIVEN validation has run with 17 doc checks, 11 ref checks, 4 config checks WHEN Tab 4 renders THEN 4 suite sections appear as accordion headers with correct counts.
- GIVEN the "Documentation" suite header is selected WHEN the user presses `Enter` THEN it expands to show all 17 individual check results.
- GIVEN an expanded suite WHEN the user presses `Enter` on the header THEN it collapses.

---

#### FR-111: Checks Tab Live Validation [MUST]

Tab 4 MUST run validation automatically when first activated (switched to). A loading indicator (spinner + "Running validation...") MUST display while validation is in progress. The user MUST be able to re-run validation by pressing `r`.

**Acceptance Criteria:**

- GIVEN the TUI launches on Tab 1 WHEN the user switches to Tab 4 for the first time THEN a spinner appears and validation runs asynchronously.
- GIVEN validation is complete WHEN Tab 4 renders THEN all suite results are displayed with no spinner.
- GIVEN validation results are displayed WHEN the user presses `r` THEN a spinner appears and validation re-runs.

---

#### FR-112: Checks Tab Check Detail [SHOULD]

Tab 4 SHOULD support toggling a detail pane for individual checks by pressing `Space`. The detail pane MUST show: file path, issue description, and fix suggestion (when available) in a bordered box below the check row.

**Acceptance Criteria:**

- GIVEN a failed check "[16] Stub detection" is selected WHEN the user presses `Space` THEN a detail pane appears showing the file path, issue, and fix suggestion.
- GIVEN the detail pane is open WHEN the user presses `Space` again THEN the detail pane closes.

---

#### FR-113: Checks Tab Overall Summary [MUST]

Tab 4 MUST display an overall summary bar at the bottom showing aggregated pass/fail/warn counts across all suites.

**Acceptance Criteria:**

- GIVEN 30/32 checks pass with 1 fail and 1 warning WHEN Tab 4 renders THEN the summary bar shows "Overall: 30/32 pass  1 fail  1 warning".

---

### Tab 5: Quality

#### FR-114: Quality Tab Score Chart [SHOULD]

Tab 5 SHOULD display an ASCII line chart of convergence score history when `quality-log.yml` exists and contains entries. The chart MUST show: Y-axis (1.0 to 5.0), X-axis (dates), data points (`●`), connecting lines, and a dashed Gate 0 threshold line at 3.0. The user MUST navigate between data points with `←`/`→` or `h`/`l`.

**Acceptance Criteria:**

- GIVEN `quality-log.yml` with 5 entries WHEN Tab 5 renders THEN an ASCII chart appears with 5 data points connected by lines and a "Gate 0" dashed line at 3.0.
- GIVEN the chart is displayed WHEN the user presses `→` THEN the next data point is selected and its details appear below.
- GIVEN terminal width of 80 columns WHEN the chart renders THEN it scales to fit within the available width.

---

#### FR-115: Quality Tab Empty State [MUST]

Tab 5 MUST display "No quality data. Run a convergence analysis and then `mind quality log <file>` to start tracking." when no `quality-log.yml` exists or when it contains no entries.

**Acceptance Criteria:**

- GIVEN no `quality-log.yml` file exists WHEN Tab 5 renders THEN the empty state message is displayed.
- GIVEN `quality-log.yml` exists but is empty WHEN Tab 5 renders THEN the empty state message is displayed.

---

#### FR-116: Quality Tab Latest Analysis Detail [SHOULD]

Tab 5 SHOULD display details for the selected data point below the chart: topic, variant, gate result, 6 dimension scores as progress bars, personas used, and output file path.

**Acceptance Criteria:**

- GIVEN data point "auth-strategy" with score 4.0 is selected WHEN Tab 5 renders THEN details show topic, variant, gate result (PASS/FAIL), and 6 dimension bars.
- GIVEN a data point with all 6 dimensions scored WHEN the detail area renders THEN each dimension shows a progress bar and numeric score.

---

### Help Overlay

#### FR-117: Help Overlay [MUST]

Pressing `?` MUST toggle a centered overlay (approximately 60x20 characters) listing all available keybindings for the current context: global keys and tab-specific keys for the active tab. The overlay MUST close on `?` or `Esc`. While the overlay is open, `q` MUST NOT quit the application.

**Acceptance Criteria:**

- GIVEN Tab 2 (Docs) is active WHEN the user presses `?` THEN the help overlay appears listing global keys and Docs-specific keys (zone filter shortcuts, `/` search, `e` edit, etc.).
- GIVEN the help overlay is open WHEN the user presses `Esc` THEN the overlay closes.
- GIVEN the help overlay is open WHEN the user presses `q` THEN the overlay closes (application does not quit).
- GIVEN the help overlay is closed WHEN the user presses `?` again from Tab 3 THEN the overlay shows global keys and Iterations-specific keys.

---

### Responsive Design

#### FR-118: Minimum Terminal Size [MUST]

The TUI MUST require a minimum terminal size of 80 columns by 24 rows. If the terminal is smaller, the TUI MUST display a centered message: "Terminal too small. Minimum: 80x24. Current: {w}x{h}." and MUST NOT render the full interface.

**Acceptance Criteria:**

- GIVEN a terminal of 79x24 WHEN `mind tui` renders THEN only the "too small" message appears.
- GIVEN a terminal of 80x24 WHEN `mind tui` renders THEN the full dashboard renders.

---

#### FR-119: Terminal Resize Handling [MUST]

The TUI MUST handle terminal resize events gracefully. On resize, all components MUST recalculate their layout and re-render without crashing. If the terminal is resized below the minimum, the "too small" message MUST appear. If resized back above the minimum, the full interface MUST resume.

**Acceptance Criteria:**

- GIVEN the TUI is running at 120x40 WHEN the terminal is resized to 80x24 THEN the layout adapts without crashing.
- GIVEN the TUI is running at 80x24 WHEN the terminal is resized to 70x20 THEN the "too small" message appears.
- GIVEN the "too small" message is showing WHEN the terminal is resized to 90x30 THEN the full dashboard resumes with all data intact.

---

### Data Loading

#### FR-120: Async Data Loading with Loading States [MUST]

The TUI MUST load data asynchronously on initialization. Each view MUST display one of four states: Loading (spinner + message), Error (message + "Press r to retry"), Empty (explanatory message), or Ready (normal rendering). Data loading MUST NOT block the UI thread.

**Acceptance Criteria:**

- GIVEN `mind tui` launches WHEN data is being fetched THEN a loading spinner appears on the active tab.
- GIVEN data loading fails (e.g., unreadable mind.toml) WHEN the Status tab renders THEN an error message appears with "Press r to retry".
- GIVEN no iterations exist WHEN Tab 3 renders THEN it shows "No iterations yet. Start a workflow to create one."

---

#### FR-121: Manual Refresh [MUST]

Pressing `r` (when no text input is focused) MUST re-load all project data from disk. While refreshing, existing data MUST remain visible. When new data arrives, it MUST replace the old data atomically.

**Acceptance Criteria:**

- GIVEN Tab 1 is displaying stale health data WHEN the user presses `r` THEN data reloads and the view updates with fresh data.
- GIVEN Tab 2 search input is focused WHEN the user presses `r` THEN the character `r` is typed (not interpreted as refresh).

---

#### FR-122: Lazy Loading for Checks Tab [MUST]

Validation MUST NOT run on TUI initialization. It MUST run lazily when the user first switches to Tab 4, or when the user presses `r` on Tab 4.

**Acceptance Criteria:**

- GIVEN `mind tui` launches on Tab 1 WHEN no user interaction occurs THEN no validation suites are executed.
- GIVEN the user switches to Tab 4 for the first time WHEN the tab renders THEN validation starts running with a spinner.
- GIVEN the user has visited Tab 4 once WHEN they switch away and back THEN the cached results are displayed (no re-run unless `r` is pressed).

---

### Tab State Preservation

#### FR-123: Tab State Preservation [MUST]

Switching tabs MUST preserve each tab's local state. When the user returns to a previously visited tab, the cursor position, scroll position, filter selection, expanded items, and search query MUST be restored.

**Acceptance Criteria:**

- GIVEN Tab 2 with cursor on row 5 and spec zone filter active WHEN the user switches to Tab 3 and back to Tab 2 THEN the cursor is on row 5 and the spec filter is still active.
- GIVEN Tab 3 with iteration 004 expanded WHEN the user switches to Tab 1 and back to Tab 3 THEN iteration 004 is still expanded.

---

### Terminal Cleanup

#### FR-124: Terminal State Restoration [MUST]

When the TUI exits (via `q`, `Ctrl+C`, or error), it MUST restore the terminal to its pre-launch state: cursor visible, input echoing enabled, alternate screen buffer exited, no residual ANSI state.

**Acceptance Criteria:**

- GIVEN the TUI is running WHEN the user presses `q` THEN the terminal returns to the normal prompt with cursor visible and input working.
- GIVEN the TUI crashes with a panic WHEN the error is caught THEN terminal state is still restored before the error is printed to stderr.

---

## Modified Requirements

### FR-49 (Exit Codes) -- Extended

**Change**: Add exit code 0 for successful `mind tui` launch-and-quit cycle. The TUI command uses the same exit code scheme: 0 (normal exit), 2 (runtime error during TUI), 3 (not a Mind project).

No change to exit codes 1 and 4.

### FR-6 (Output Modes) -- Clarification

**Change**: The TUI is a separate output path -- it does not use the `Renderer` or `OutputMode` enum. The TUI renders through Bubble Tea's `View()` method with Lip Gloss styling. `--json` is not applicable to `mind tui`.

---

## Unchanged Requirements

The following requirement groups MUST NOT change:

- **FR-1 through FR-5**: Project detection and configuration
- **FR-6 through FR-10**: Output modes (CLI path unchanged)
- **FR-11 through FR-13**: Status command (CLI path unchanged)
- **FR-14 through FR-19**: Init command
- **FR-20 through FR-23**: Doctor command
- **FR-24 through FR-31**: Create commands
- **FR-32 through FR-37**: Docs commands (CLI path unchanged)
- **FR-38 through FR-43**: Check commands (CLI path unchanged)
- **FR-44 through FR-48**: Workflow and version commands
- **FR-49**: Exit codes (extended, not replaced)
- **FR-50**: Stub detection
- **FR-51 through FR-87**: Reconciliation engine (Phase 1.5)
- **NFR-1 through NFR-11**: Non-functional requirements
- **C-1 through C-17**: Constraints
- **BR-1 through BR-35**: Business rules

---

## Domain Model Impact

### New Entities

| Entity | Description | Package |
|--------|-------------|---------|
| **QualityEntry** | A single convergence analysis result from `quality-log.yml`. Contains topic, variant, date, overall score, gate result, dimension scores, personas, and output path. | `domain/quality.go` |
| **QualityDimension** | A single dimension score within a QualityEntry (e.g., rigor=4, coverage=4). | `domain/quality.go` |

### New Supporting Types

| Type | Kind | Values | Used By |
|------|------|--------|---------|
| **TabID** | Enum (int) | `TabStatus` (0), `TabDocs` (1), `TabIterations` (2), `TabChecks` (3), `TabQuality` (4) | TUI app shell (presentation layer, not domain) |
| **ViewState** | Enum (int) | `ViewLoading` (0), `ViewError` (1), `ViewEmpty` (2), `ViewReady` (3) | TUI tab models (presentation layer, not domain) |

### New Business Rules

| ID | Rule | Entities | Invariant |
|----|------|----------|-----------|
| **BR-36** | A QualityEntry MUST have an overall score in the range 0.0 to 5.0 inclusive. | QualityEntry | `0.0 <= Score <= 5.0` |
| **BR-37** | A QualityEntry passes Gate 0 when its overall score is >= 3.0. | QualityEntry | `GatePass == (Score >= 3.0)` |
| **BR-38** | QualityEntry dimension scores MUST each be in the range 0 to 5 inclusive. There are exactly 6 dimensions: rigor, coverage, actionability, objectivity, convergence, depth. | QualityDimension | `0 <= Value <= 5`, `len(Dimensions) == 6` |

### Domain Layer Purity Note

`TabID` and `ViewState` are presentation-layer types that belong in the `tui/` package, not `domain/`. They are listed here for completeness but MUST NOT be added to `domain/`. Only `QualityEntry`, `QualityDimension`, and business rules BR-36 through BR-38 affect the domain layer.

---

## Structural Impact

### New Packages

| Package | Files | Responsibility |
|---------|-------|----------------|
| `tui/` | `app.go`, `status.go`, `docs.go`, `iterations.go`, `checks.go`, `quality.go`, `styles.go`, `keys.go`, `help.go`, `statusbar.go` | Bubble Tea TUI application -- presentation layer |
| `tui/components/` | `health_panel.go`, `zone_bar.go`, `staleness.go`, `workflow_panel.go`, `warnings_panel.go`, `suggestions_panel.go`, `quick_actions.go`, `zone_filter.go`, `type_filter.go`, `detail_expander.go`, `suite_accordion.go`, `check_detail.go`, `overall_summary.go`, `score_chart.go`, `latest_analysis.go`, `empty_state.go` | Reusable TUI components |

### New Files in Existing Packages

| Package | File | Contents |
|---------|------|----------|
| `domain/` | `quality.go` | `QualityEntry`, `QualityDimension` types, BR-36/37/38 validation |
| `internal/repo/` | `interfaces.go` (modified) | `QualityRepo` interface addition |
| `internal/repo/fs/` | `quality_repo.go` | Filesystem implementation of `QualityRepo` reading `quality-log.yml` |
| `internal/repo/mem/` | `quality_repo.go` | In-memory `QualityRepo` for testing |
| `cmd/` | `tui_cmd.go` | `mind tui` Cobra command, wiring to TUI app |

### Modified Files

| File | Change |
|------|--------|
| `cmd/reconcile.go` | FR-88: Add `--check`/`--force` mutual exclusion guard |
| `internal/validate/reconcile.go` | FR-89: Add missing documents check |
| `main.go` / `cmd/root.go` | FR-90: Centralize wiring via `buildDeps()` |
| `cmd/docs.go` | FR-91: Refactor search to use DocRepo |
| `internal/repo/interfaces.go` | Add `QualityRepo` interface |
| `go.mod` | Add `bubbletea`, `bubbles`, `glamour` dependencies |

### New Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/charmbracelet/bubbletea` | v1.2+ | TUI framework (Elm architecture) |
| `github.com/charmbracelet/bubbles` | v0.20+ | Standard TUI components (table, list, viewport, spinner, textinput) |
| `github.com/charmbracelet/glamour` | latest | Markdown rendering for document preview |

---

## Non-Functional Requirements (Phase 2 Additions)

- **NFR-12**: `mind tui` MUST launch (show first frame) in under 500ms for a project with 50 documents. [MUST]
- **NFR-13**: Tab switching MUST complete in under 50ms (no perceptible delay). [MUST]
- **NFR-14**: The TUI MUST use no more than 50MB of memory during normal operation with a 50-document project. [SHOULD]
- **NFR-15**: `tui/` package test coverage MUST be >= 60% for `Update()` functions. `View()` functions are excluded from coverage requirements. [SHOULD]

---

## Architectural Guidance for the Architect

Per convergence recommendation R3 (75% MEDIUM confidence):

1. Use Bubbles components (`bubbles/table`, `bubbles/list`, `bubbles/viewport`, `bubbles/spinner`, `bubbles/textinput`) for standard UI elements. Do not build custom components where Bubbles provides equivalents.
2. Use `glamour` for Docs tab preview pane markdown rendering.
3. Evaluate `asciigraph` for Quality tab chart. If chart characters do not match BP-05 spec, build a custom chart component (estimated 150-200 lines).
4. Use `teatest` for model state assertions. Target: 1 golden file test per tab view at 80x24, plus 3-5 state transition tests per tab model.
5. Follow MVU (Elm architecture) per BP-05 Section 3: `App` delegates to per-tab `tea.Model` implementations. Communication via Bubble Tea messages only.
6. Services injected into `App` at construction time, passed to tab models during initialization.

---

## Requirement Summary

| Category | FR Range | Count | Severity |
|----------|----------|-------|----------|
| SHOULD fixes (pre-TUI) | FR-88 -- FR-91 | 4 | MUST |
| TUI command | FR-92 | 1 | MUST |
| App shell | FR-93 -- FR-96 | 4 | MUST |
| Tab 1: Status | FR-97 -- FR-101 | 5 | 4 MUST, 1 SHOULD |
| Tab 2: Documents | FR-102 -- FR-106 | 5 | 3 MUST, 2 SHOULD |
| Tab 3: Iterations | FR-107 -- FR-109 | 3 | 2 MUST, 1 SHOULD |
| Tab 4: Checks | FR-110 -- FR-113 | 4 | 3 MUST, 1 SHOULD |
| Tab 5: Quality | FR-114 -- FR-116 | 3 | 1 MUST, 2 SHOULD |
| Help overlay | FR-117 | 1 | MUST |
| Responsive design | FR-118 -- FR-119 | 2 | MUST |
| Data loading | FR-120 -- FR-122 | 3 | MUST |
| Tab state | FR-123 | 1 | MUST |
| Terminal cleanup | FR-124 | 1 | MUST |
| **Total** | **FR-88 -- FR-124** | **37** | **27 MUST, 10 SHOULD** |
