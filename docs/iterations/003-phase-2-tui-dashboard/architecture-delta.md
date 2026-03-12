# Architecture Delta: Phase 2 TUI Dashboard

**Iteration**: 003-phase-2-tui-dashboard
**Date**: 2026-03-11
**Phase**: 2
**Type**: COMPLEX_NEW -- new presentation layer (TUI) added alongside existing CLI, plus 4 SHOULD fixes

---

## Current Structure

Phase 1 + 1.5 delivers a 4-layer architecture (Presentation, Service, Domain, Infrastructure) with downward-only dependency flow. The domain layer is pure Go (zero external imports). All filesystem access passes through repository interfaces. The codebase has 246 passing tests with domain/ at 100% coverage.

Relevant portions for Phase 2:

```
Presentation Layer:
  cmd/root.go ──────> PersistentPreRunE wires repos + services into package-level vars
  cmd/status.go ────> projectSvc.AssembleHealth(), reconcileSvc.ReadStaleness()
  cmd/check.go ─────> validationSvc.RunDocs/RunRefs/RunConfig/RunAll()
  cmd/docs.go ──────> docRepo.ListByZone/ListAll(), direct filepath.WalkDir for search
  cmd/reconcile.go ─> reconcileSvc.Reconcile(), reconcileSvc.LoadGraph()
  internal/render/ ──> Renderer with 3 OutputModes (Interactive, Plain, JSON)

Service Layer:
  internal/service/project.go ──────> ProjectService.AssembleHealth()
  internal/service/validation.go ───> ValidationService.RunAll()
  internal/service/reconciliation.go > ReconciliationService.Reconcile(), ReadStaleness()
  internal/service/workflow.go ─────> WorkflowService.Status(), History()
  internal/service/doctor.go ───────> DoctorService.Run()
  internal/service/generate.go ─────> GenerateService.Create*()

Domain Layer:
  domain/ ──> pure types: Project, Config, Document, ZoneHealth, ProjectHealth,
              ValidationReport, CheckResult, WorkflowState, Iteration,
              LockFile, ReconcileResult, StalenessInfo

Infrastructure Layer:
  internal/repo/interfaces.go ──> DocRepo, IterationRepo, StateRepo, ConfigRepo,
                                   LockRepo, BriefRepo
  internal/repo/fs/ ────────────> Filesystem implementations
  internal/repo/mem/ ───────────> In-memory test implementations
```

### Current Issues Targeted by Phase 2

**Issue 1 (S-1)**: `cmd/reconcile.go` does not guard against `--check` and `--force` being set simultaneously. The api-contracts spec (line 768) documents them as mutually exclusive, but the code does not enforce this.

**Issue 2 (S-2)**: `ReconcileSuite` in `internal/validate/reconcile.go` checks for cycles (check 1) and stale documents (checks 2+), but does not check for documents declared in `mind.toml` that are missing from disk. This data is available in `ReconcileResult.Missing`.

**Issue 3 (Wiring)**: While Phase 1.5 centralized wiring into `PersistentPreRunE` in `cmd/root.go` (lines 42-79), this function is tightly coupled to the CLI's Cobra lifecycle. The TUI needs the same services but does not go through Cobra's command dispatch. A shared `buildDeps()` function is needed.

**Issue 4 (FR-91)**: `cmd/docs.go:runDocsSearch()` (lines 146-206) uses `filepath.WalkDir` and `os.Open` directly, bypassing `DocRepo`. The TUI's Documents tab needs the same search logic through the service layer.

---

## Proposed Changes

### SHOULD Fixes (FR-88 through FR-91)

These are resolved before TUI implementation per convergence recommendation R1 (90% HIGH confidence).

#### FR-88: Mutual Exclusion Guard

**File**: `cmd/reconcile.go`

**Change**: Add a guard at the top of `runReconcile()` before opts construction:

```go
func runReconcile(cmd *cobra.Command, args []string) error {
    if flagReconcileCheck && flagReconcileForce {
        fmt.Fprintln(os.Stderr, "Error: --check and --force are mutually exclusive")
        os.Exit(2)
        return nil
    }
    // ... existing code
}
```

**Rationale**: Exit code 2 is "runtime error" per the exit code scheme. This is a usage error, not a validation failure (1) or config error (3).

---

#### FR-89: Missing Documents Check in ReconcileSuite

**File**: `internal/validate/reconcile.go`

**Change**: Insert a check between the cycle check (check 1) and the stale document checks (checks 2+):

```go
// Check 2: No missing documents
if len(result.Missing) == 0 {
    report.Checks = append(report.Checks, domain.CheckResult{
        ID:     checkID,
        Name:   "No missing documents",
        Level:  domain.LevelWarn,
        Passed: true,
    })
    report.Passed++
} else {
    for _, id := range result.Missing {
        level := domain.LevelWarn
        if strict {
            level = domain.LevelFail
        }
        report.Checks = append(report.Checks, domain.CheckResult{
            ID:      checkID,
            Name:    fmt.Sprintf("Document exists: %s", id),
            Level:   level,
            Passed:  false,
            Message: fmt.Sprintf("declared in mind.toml but not found on disk: %s", id),
        })
        if level == domain.LevelFail {
            report.Failed++
        } else {
            report.Warnings++
        }
        checkID++
    }
}
checkID++
```

**Rationale**: Missing documents are WARN by default (not structurally broken, but incomplete). `--strict` promotes them to FAIL, matching the existing pattern for stale documents.

---

#### FR-90: Wiring Centralization via Deps Struct

**Files**: `cmd/root.go`, `cmd/tui_cmd.go` (new)

**Change**: Extract a `Deps` struct and `BuildDeps()` function from the current `PersistentPreRunE` body. Both CLI and TUI call `BuildDeps()` to get identical service wiring.

```go
// Deps holds all services and repositories. Constructed once, shared by CLI and TUI.
type Deps struct {
    ProjectRoot   string
    Renderer      *render.Renderer
    DocRepo       *fs.DocRepo
    IterRepo      *fs.IterationRepo
    BriefRepo     *fs.BriefRepo
    ConfigRepo    *fs.ConfigRepo
    LockRepo      *fs.LockRepo
    StateRepo     *fs.StateRepo
    ProjectSvc    *service.ProjectService
    ValidationSvc *service.ValidationService
    ReconcileSvc  *service.ReconciliationService
    DoctorSvc     *service.DoctorService
    WorkflowSvc   *service.WorkflowService
    GenerateSvc   *service.GenerateService
}

// BuildDeps constructs all repositories and services for a given project root.
// Renderer is nil when called from TUI (TUI uses Lip Gloss directly).
func BuildDeps(root string, r *render.Renderer) *Deps {
    docRepo := fs.NewDocRepo(root)
    iterRepo := fs.NewIterationRepo(root)
    stateRepo := fs.NewStateRepo(root)
    briefRepo := fs.NewBriefRepo(docRepo)
    configRepo := fs.NewConfigRepo(root)
    lockRepo := fs.NewLockRepo(root)

    return &Deps{
        ProjectRoot:   root,
        Renderer:      r,
        DocRepo:       docRepo,
        IterRepo:      iterRepo,
        BriefRepo:     briefRepo,
        ConfigRepo:    configRepo,
        LockRepo:      lockRepo,
        StateRepo:     stateRepo,
        ProjectSvc:    service.NewProjectService(docRepo, iterRepo, stateRepo, briefRepo),
        ValidationSvc: service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo),
        ReconcileSvc:  service.NewReconciliationService(configRepo, docRepo, lockRepo),
        DoctorSvc:     service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo, lockRepo),
        WorkflowSvc:   service.NewWorkflowService(stateRepo, iterRepo),
        GenerateSvc:   service.NewGenerateService(root),
    }
}
```

**PersistentPreRunE** then becomes:

```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if !requiresProject(cmd) {
        return nil
    }
    root, err := resolveRoot()
    if err != nil {
        if isNotProject(err) {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(3)
        }
        return err
    }
    mode := render.DetectMode(flagJSON, flagNoColor)
    r := render.New(mode, render.TermWidth())
    deps := BuildDeps(root, r)

    // Populate package-level variables for command handlers
    projectRoot = deps.ProjectRoot
    renderer = deps.Renderer
    docRepo = deps.DocRepo
    // ... etc
    return nil
}
```

**TUI** calls `BuildDeps(root, nil)` -- renderer is nil because the TUI uses Lip Gloss directly through Bubble Tea's `View()` method.

**Rationale**: This is the minimal change that eliminates dual wiring without restructuring every command handler. The package-level variables remain for backward compatibility with existing CLI handlers. The `Deps` struct enables the TUI to receive services through a single constructor call. Over time, command handlers can migrate to using `deps` directly, but that migration is not required for Phase 2.

---

#### FR-91: DocRepo Search Abstraction

**Files**: `internal/repo/interfaces.go`, `internal/repo/fs/doc_repo.go`, `internal/repo/mem/doc_repo.go`, `cmd/docs.go`

**Change**: Add a `Search` method to the `DocRepo` interface:

```go
// DocRepo reads and queries the 5-zone documentation structure.
type DocRepo interface {
    // ... existing methods ...

    // Search returns documents whose content matches the query string.
    // Search is case-insensitive substring matching across all .md files in docs/.
    // Each result includes matching lines with 1 line of context.
    Search(query string) (*domain.SearchResults, error)
}
```

`fs.DocRepo.Search()` moves the existing `filepath.WalkDir` logic from `cmd/docs.go:runDocsSearch()` into the repository implementation. `mem.DocRepo.Search()` provides an in-memory implementation for testing.

`cmd/docs.go:runDocsSearch()` becomes:

```go
func runDocsSearch(cmd *cobra.Command, args []string) error {
    results, err := docRepo.Search(args[0])
    if err != nil {
        return fmt.Errorf("search docs: %w", err)
    }
    fmt.Print(renderer.RenderSearchResults(results))
    return nil
}
```

**Rationale**: The TUI's Documents tab (FR-104) needs search through the service layer. Moving search into DocRepo follows the established pattern where all filesystem access goes through repository interfaces. The in-memory implementation enables search testing without filesystem access.

---

### New Components: TUI Package

#### `tui/` Package Structure

The TUI is a new presentation-layer package that sits alongside `cmd/` and `internal/render/`. It follows the Bubble Tea Elm architecture: each component is a `tea.Model` with `Init()`, `Update()`, and `View()` methods.

**File layout**:

| File | Responsibility | tea.Model? |
|------|---------------|------------|
| `tui/app.go` | Top-level model: chrome, tab delegation, service injection, global keys | Yes |
| `tui/status.go` | Tab 1: zone health bars, staleness, workflow, warnings, suggestions | Yes |
| `tui/docs.go` | Tab 2: document list, zone filter, search, preview pane | Yes |
| `tui/iterations.go` | Tab 3: iteration table, type filter, detail expander | Yes |
| `tui/checks.go` | Tab 4: accordion suites, live validation, spinner, check detail | Yes |
| `tui/quality.go` | Tab 5: score chart, latest analysis, empty state | Yes |
| `tui/help.go` | Help overlay: context-sensitive keybinding reference | Yes |
| `tui/statusbar.go` | Bottom status bar: key hints, cursor position | No (pure view function) |
| `tui/styles.go` | Lip Gloss theme: colors, borders, text styles | No (constants) |
| `tui/keys.go` | Key binding definitions: global, per-tab | No (constants) |
| `tui/messages.go` | Custom Bubble Tea message types for inter-component communication | No (types) |

**Component sub-package** (`tui/components/`):

Reusable view-rendering functions (not full `tea.Model` implementations). These are pure functions that take data and terminal dimensions and return styled strings.

| File | Renders | Data Input |
|------|---------|------------|
| `components/zone_bar.go` | Zone progress bar with label and fraction | `domain.ZoneHealth`, width |
| `components/staleness.go` | Stale document list with bullet markers | `map[string]string`, width |
| `components/workflow_panel.go` | Workflow state panel | `*domain.WorkflowState` |
| `components/warnings.go` | Warning list with prefix markers | `[]string` |
| `components/suggestions.go` | Suggestion list with prefix markers | `[]string` |
| `components/quick_actions.go` | Quick action key reference | (static) |
| `components/zone_filter.go` | Zone filter bar | `domain.Zone` (active), width |
| `components/type_filter.go` | Iteration type filter bar | `domain.RequestType` (active), width |
| `components/detail_expander.go` | Inline iteration artifact list | `domain.Iteration` |
| `components/suite_accordion.go` | Collapsible validation suite section | `domain.ValidationReport`, expanded bool |
| `components/check_detail.go` | Bordered check detail box | `domain.CheckResult` |
| `components/overall_summary.go` | Aggregated pass/fail/warn summary bar | totals |
| `components/score_chart.go` | ASCII line chart for quality scores | `[]domain.QualityEntry`, width, height |
| `components/latest_analysis.go` | Selected quality entry detail | `domain.QualityEntry`, width |
| `components/empty_state.go` | Centered empty state message | message string, width, height |

---

#### `tui/app.go` -- App Model

**Purpose**: Top-level Bubble Tea model. Owns the service references, active tab index, terminal dimensions, and child tab models. Delegates `Update()` messages to the active tab. Handles global key bindings.

**Constructor**:

```go
type App struct {
    deps       *cmd.Deps
    width      int
    height     int
    activeTab  TabID
    tabs       [5]tea.Model    // statusView, docsView, iterationsView, checksView, qualityView
    help       HelpModel
    showHelp   bool
    loaded     bool
}

// TabID is an int enum for the 5 tabs.
type TabID int

const (
    TabStatus     TabID = iota // 0
    TabDocs                    // 1
    TabIterations              // 2
    TabChecks                  // 3
    TabQuality                 // 4
)

func NewApp(deps *cmd.Deps) App {
    return App{
        deps:      deps,
        activeTab: TabStatus,
        tabs: [5]tea.Model{
            NewStatusView(deps),
            NewDocsView(deps),
            NewIterationsView(deps),
            NewChecksView(deps),
            NewQualityView(deps),
        },
        help: NewHelpModel(),
    }
}
```

**Init()**: Returns a `tea.Batch` of commands to load initial data:
- `loadProjectHealth` -- feeds StatusView and DocsView
- `loadIterations` -- feeds IterationsView
- `loadQualityEntries` -- feeds QualityView
- Validation is NOT loaded on init (lazy, per FR-122)

**Update() dispatch logic**:

```
1. tea.WindowSizeMsg → update width/height, propagate to all tabs
2. tea.KeyMsg:
   a. Ctrl+C → tea.Quit (always, even with modal open)
   b. showHelp == true → delegate to help.Update(); if '?' or Esc, close help
   c. '?' → toggle showHelp
   d. '1'-'5' → switch activeTab
   e. Tab/Shift+Tab → cycle activeTab
   f. 'q' → tea.Quit
   g. 'r' (and no text input focused) → re-dispatch initial data load commands
   h. Otherwise → delegate to tabs[activeTab].Update()
3. Custom messages (healthLoadedMsg, etc.) → route to appropriate tab(s)
```

**View()**: Composes the full screen:

```
titleBar(width, project, branch, version)
tabBar(width, activeTab)
separator(width)
tabs[activeTab].View()       // fills remaining height
statusBar(width, activeTab, tabState)
```

If `showHelp` is true, the help overlay is rendered on top of the content area.

If `width < 80 || height < 24`, render only the "Terminal too small" message.

---

#### `tui/status.go` -- StatusView Model

**Purpose**: Tab 1. Displays zone health bars, staleness panel, workflow state, warnings, suggestions, and quick actions in a two-column layout.

**Data**: `*domain.ProjectHealth` (received via `healthLoadedMsg`)

**State**:
- `viewState ViewState` -- Loading, Error, Empty, or Ready
- `health *domain.ProjectHealth`
- `errMsg string`
- `width, height int`

**View layout** (width >= 80):
```
Left column (50%):         Right column (50%):
  Documentation Health       Active Workflow
  [zone bars x5]             state, type, agent, branch
  Staleness                  Quick Actions
  [stale docs]               key-action list
  Warnings
  [warning list]
  Suggestions
  [suggestion list]
```

At width < 80: single column, panels stacked vertically.

**Service integration**:
- `deps.ProjectSvc.AssembleHealth(project)` via `healthLoadedMsg`
- `deps.ReconcileSvc.ReadStaleness(root)` via `healthLoadedMsg` (attached to health)

---

#### `tui/docs.go` -- DocsView Model

**Purpose**: Tab 2. Displays documents grouped by zone with filter, search, and preview.

**Data**: `[]domain.Document` (derived from `healthLoadedMsg` zone data)

**State**:
- `viewState ViewState`
- `documents []domain.Document` -- all documents
- `filtered []domain.Document` -- after zone filter + search applied
- `activeZone *domain.Zone` -- nil = all zones
- `searchInput textinput.Model` -- from bubbles
- `searchActive bool`
- `cursor int`
- `previewVisible bool`
- `previewContent string`
- `previewViewport viewport.Model` -- from bubbles
- `width, height int`

**Key interactions**:
- Zone filter keys (`a`, `s`, `b`, `t`, `i`, `k`) update `activeZone` and recompute `filtered`
- `/` activates `searchInput`; real-time filtering on keystroke
- `Esc` clears search and closes preview
- `Enter` on a document triggers `loadPreview` command (reads via `deps.DocRepo.Read()`, renders via Glamour)
- `e` suspends TUI, opens `$EDITOR`, resumes on exit

**Service integration**:
- Documents come from `ProjectHealth.Zones` (no separate load)
- Preview: `deps.DocRepo.Read(relPath)` + `glamour.Render()`
- Search: `deps.DocRepo.Search(query)` for inline filtering

---

#### `tui/iterations.go` -- IterationsView Model

**Purpose**: Tab 3. Displays iterations in a table with type filter and expandable detail.

**Data**: `[]domain.Iteration` (received via `iterationsLoadedMsg`)

**State**:
- `viewState ViewState`
- `iterations []domain.Iteration`
- `filtered []domain.Iteration`
- `activeType *domain.RequestType` -- nil = all types
- `cursor int`
- `expandedIndex int` -- -1 = none
- `width, height int`

**Key interactions**:
- Type filter keys (`a`, `n`, `e`, `b`, `r`) update `activeType` and recompute `filtered`
- `Enter` toggles `expandedIndex` for inline artifact detail
- Arrow/vim keys navigate rows

**Service integration**:
- `deps.IterRepo.List()` via `iterationsLoadedMsg`

---

#### `tui/checks.go` -- ChecksView Model

**Purpose**: Tab 4. Displays validation results as an accordion with live validation and spinner.

**Data**: `[]domain.ValidationReport` (received via `validationCompleteMsg`)

**State**:
- `viewState ViewState`
- `reports []domain.ValidationReport`
- `expandedSuites map[int]bool`
- `cursor int` -- position across all suite headers and check rows
- `detailVisible bool`
- `detailTarget int`
- `loading bool`
- `spinner spinner.Model` -- from bubbles
- `width, height int`

**Lazy loading**: Validation does NOT run on init. When the user first switches to Tab 4, the view dispatches `runValidation` command. On subsequent visits, cached results are shown unless `r` is pressed.

**Key interactions**:
- `Enter` on suite header toggles expand/collapse
- `Space` on a check row toggles detail pane
- `r` re-runs all suites (shows spinner)

**Service integration**:
- `deps.ValidationSvc.RunAll(root, false, reconcileResult)` via `validationCompleteMsg`
- `deps.ReconcileSvc.Reconcile(root, domain.ReconcileOpts{CheckOnly: true})` for reconcile result

---

#### `tui/quality.go` -- QualityView Model

**Purpose**: Tab 5. Displays quality score history as ASCII chart with analysis details, or empty state.

**Data**: `[]domain.QualityEntry` (received via `qualityLoadedMsg`)

**State**:
- `viewState ViewState`
- `entries []domain.QualityEntry`
- `selectedIndex int`
- `width, height int`

**Key interactions**:
- `Left`/`h` and `Right`/`l` navigate between data points on the chart
- `Enter` shows full analysis details for selected point

**Service integration**:
- `QualityRepo.ReadLog()` via `qualityLoadedMsg`

---

#### `tui/messages.go` -- Custom Message Types

```go
// Data loading messages
type healthLoadedMsg struct{ health *domain.ProjectHealth }
type healthErrorMsg struct{ err error }
type iterationsLoadedMsg struct{ iterations []domain.Iteration }
type iterationsErrorMsg struct{ err error }
type qualityLoadedMsg struct{ entries []domain.QualityEntry }
type qualityErrorMsg struct{ err error }
type validationCompleteMsg struct{ reports []domain.ValidationReport; reconcile *domain.ReconcileResult }
type validationErrorMsg struct{ err error }
type previewLoadedMsg struct{ content string }
type previewErrorMsg struct{ err error }

// UI state messages
type validationStartedMsg struct{}
```

---

#### `tui/styles.go` -- Theme Constants

All Lip Gloss styles defined in a single `Theme` struct for consistent styling across all tabs.

```go
type Theme struct {
    // Zone colors
    ZoneSpec       lipgloss.Style
    ZoneBlueprints lipgloss.Style
    ZoneState      lipgloss.Style
    ZoneIterations lipgloss.Style
    ZoneKnowledge  lipgloss.Style

    // Severity
    Pass    lipgloss.Style  // green
    Fail    lipgloss.Style  // red
    Warn    lipgloss.Style  // yellow
    Dim     lipgloss.Style  // gray

    // Chrome
    TitleBar   lipgloss.Style
    TabActive  lipgloss.Style  // bold + underline
    TabInactive lipgloss.Style // dim
    StatusBar  lipgloss.Style
    Separator  lipgloss.Style

    // Content
    Selected   lipgloss.Style  // reverse video
    Border     lipgloss.Style
    Heading    lipgloss.Style  // bold
    Subtle     lipgloss.Style  // dim text

    // Progress bars
    BarFilled  lipgloss.Style
    BarEmpty   lipgloss.Style
}
```

Zone colors per BP-05 Section 6: spec=blue (#5f87ff), blueprints=cyan (#5fd7ff), state=yellow (#ffd700), iterations=green (#5fd787), knowledge=magenta (#d75fd7).

---

#### `tui/keys.go` -- Key Binding Definitions

Uses `bubbles/key` for keybinding definitions with help text. Global bindings defined once; tab-specific bindings defined per tab model.

```go
type GlobalKeyMap struct {
    Quit      key.Binding
    ForceQuit key.Binding
    Help      key.Binding
    Refresh   key.Binding
    Tab1      key.Binding
    Tab2      key.Binding
    Tab3      key.Binding
    Tab4      key.Binding
    Tab5      key.Binding
    NextTab   key.Binding
    PrevTab   key.Binding
}
```

---

#### `cmd/tui_cmd.go` -- TUI Command Wiring

**Purpose**: Cobra command handler for `mind tui`. Resolves project root, builds deps, constructs App, runs Bubble Tea program.

```go
var tuiCmd = &cobra.Command{
    Use:   "tui",
    Short: "Launch interactive TUI dashboard",
    RunE:  runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
    root, err := resolveRoot()
    if err != nil {
        if isNotProject(err) {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(3)
        }
        return err
    }

    deps := BuildDeps(root, nil) // nil renderer -- TUI uses Lip Gloss directly
    app := tui.NewApp(deps)

    p := tea.NewProgram(app, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        return fmt.Errorf("TUI error: %w", err)
    }
    return nil
}
```

**Note**: `runTUI` does NOT go through `PersistentPreRunE` for service wiring. It calls `resolveRoot()` and `BuildDeps()` directly. The `requiresProject()` function in `root.go` must be updated to return `false` for the `tui` command so that `PersistentPreRunE` does not wire services redundantly. The TUI command handles its own wiring because it needs `tea.WithAltScreen()` and does not use the CLI renderer.

---

### New Domain Types

#### `domain/quality.go` -- QualityEntry and QualityDimension

```go
// QualityEntry represents a single convergence analysis result from quality-log.yml.
type QualityEntry struct {
    Topic      string             `json:"topic" yaml:"topic"`
    Variant    string             `json:"variant" yaml:"variant"`
    Date       time.Time          `json:"date" yaml:"date"`
    Score      float64            `json:"score" yaml:"score"`
    GatePass   bool               `json:"gate_pass" yaml:"gate_pass"`
    Dimensions []QualityDimension `json:"dimensions" yaml:"dimensions"`
    Personas   []string           `json:"personas" yaml:"personas"`
    OutputPath string             `json:"output_path" yaml:"output_path"`
}

// Validate checks BR-36, BR-37, BR-38.
func (e QualityEntry) Validate() error {
    if e.Score < 0.0 || e.Score > 5.0 {
        return fmt.Errorf("score %.2f outside valid range [0.0, 5.0]", e.Score)
    }
    if e.GatePass != (e.Score >= 3.0) {
        return fmt.Errorf("gate_pass=%v inconsistent with score %.2f", e.GatePass, e.Score)
    }
    if len(e.Dimensions) != 6 {
        return fmt.Errorf("expected 6 dimensions, got %d", len(e.Dimensions))
    }
    for _, d := range e.Dimensions {
        if d.Value < 0 || d.Value > 5 {
            return fmt.Errorf("dimension %s value %d outside valid range [0, 5]", d.Name, d.Value)
        }
    }
    return nil
}

// QualityDimension represents a single dimension score within a convergence analysis.
type QualityDimension struct {
    Name  string `json:"name" yaml:"name"`
    Value int    `json:"value" yaml:"value"`
}

// Standard quality dimension names.
const (
    DimRigor         = "rigor"
    DimCoverage      = "coverage"
    DimActionability = "actionability"
    DimObjectivity   = "objectivity"
    DimConvergence   = "convergence"
    DimDepth         = "depth"
)
```

**Domain purity**: `QualityEntry.Validate()` is a pure function using only `fmt` from stdlib. No external imports. This follows the DC-4 pattern established by `Slugify()` and `Classify()`.

---

### New Repository Interface

#### `QualityRepo` in `internal/repo/interfaces.go`

```go
// QualityRepo reads quality log data.
type QualityRepo interface {
    // ReadLog returns all quality entries from quality-log.yml, ordered by date.
    // Returns empty slice and nil error if the file does not exist.
    ReadLog() ([]domain.QualityEntry, error)
}
```

#### `internal/repo/fs/quality_repo.go`

Filesystem implementation. Reads `quality-log.yml` from the project root using `gopkg.in/yaml.v3` (or the existing TOML parser if quality-log is TOML). Returns `[]domain.QualityEntry` ordered by date.

If `quality-log.yml` does not exist, returns an empty slice (not an error). This enables the empty state in Tab 5.

#### `internal/repo/mem/quality_repo.go`

In-memory implementation backed by a `[]domain.QualityEntry` slice. Used for testing Tab 5 without filesystem access.

---

### Modified Components

#### `cmd/root.go` -- Wiring Centralization

| Aspect | Current | Proposed | Reason |
|--------|---------|----------|--------|
| Service construction | Inline in `PersistentPreRunE` | `BuildDeps()` function, called by both `PersistentPreRunE` and `cmd/tui_cmd.go` | FR-90: eliminate dual wiring |
| Package-level vars | 12 vars for repos/services | Retained for CLI backward compat; `Deps` struct added as the canonical grouping | Minimum-change migration |
| `requiresProject()` | Checks command name | Add `"tui"` to skip list | TUI handles its own wiring |

#### `cmd/reconcile.go` -- Mutual Exclusion

| Aspect | Current | Proposed | Reason |
|--------|---------|----------|--------|
| Flag validation | None | Guard at top of `runReconcile()` | FR-88 |

#### `internal/validate/reconcile.go` -- Missing Documents Check

| Aspect | Current | Proposed | Reason |
|--------|---------|----------|--------|
| Check count | 1 (cycle) + N (stale) | 1 (cycle) + M (missing) + N (stale) | FR-89 |

#### `internal/repo/interfaces.go` -- DocRepo Search Method

| Aspect | Current | Proposed | Reason |
|--------|---------|----------|--------|
| DocRepo methods | ListByZone, ListAll, Read, Exists, IsStub, IsDir | Add `Search(query string) (*domain.SearchResults, error)` | FR-91 |

#### `internal/repo/interfaces.go` -- QualityRepo Interface

| Aspect | Current | Proposed | Reason |
|--------|---------|----------|--------|
| Interfaces | DocRepo, IterationRepo, StateRepo, ConfigRepo, LockRepo, BriefRepo | Add `QualityRepo` | Tab 5 data source |

#### `go.mod` -- New Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/charmbracelet/bubbletea` | v1.2+ | TUI framework (Elm architecture) |
| `github.com/charmbracelet/bubbles` | v0.20+ | Standard TUI components (table, viewport, spinner, textinput, key) |
| `github.com/charmbracelet/glamour` | latest | Markdown rendering for document preview pane |
| `gopkg.in/yaml.v3` | v3.0+ | Parse `quality-log.yml` (only if not using existing TOML for quality log) |

**Note**: `lipgloss` is already an indirect dependency (used by `internal/render/`). It will become a direct dependency of `tui/`.

---

## Key Decisions

### Decision 1: TUI as a Peer Presentation Layer, Not a Wrapper

- **Choice**: The `tui/` package is a new presentation-layer package at the same level as `cmd/` and `internal/render/`. It accesses services through the same interfaces as the CLI. It does NOT wrap CLI commands or call CLI rendering functions.
- **Rationale**: The TUI has fundamentally different rendering requirements (Bubble Tea `View()` returning strings styled with Lip Gloss, not the `Renderer` with three output modes). Wrapping CLI output would mean parsing text to display in the TUI, which is fragile and defeats the purpose of structured domain types. Direct service access gives the TUI full control over data presentation.
- **Rejected alternatives**:
  - **TUI calls CLI commands and parses output**: Fragile, couples TUI to text formatting, prevents async data loading. Rejected.
  - **TUI extends the existing Renderer with a fourth mode**: Would require threading Bubble Tea concepts through the Renderer, violating its single-responsibility. The Renderer converts domain types to strings; the TUI converts domain types to interactive UI. These are different concerns. Rejected.
  - **TUI lives inside `cmd/` as another command**: This conflates the CLI presentation layer with the TUI presentation layer. Keeping them in separate packages enforces clean boundaries. Rejected.
- **Consequences**: Two presentation layers share services but have independent rendering. Adding a new domain type requires updating both CLI rendering and TUI views. This is acceptable because the presentation needs are genuinely different.
- **Convergence adoption**: Adopts convergence R2 (tab-by-tab implementation) and the unanimous agreement on service reuse from all four convergence personas.

---

### Decision 2: Per-Tab Delegated Model Architecture

- **Choice**: Each tab is an independent `tea.Model` with its own state, `Init()`, `Update()`, and `View()`. The top-level `App` model delegates messages to the active tab. Communication between `App` and tabs is via Bubble Tea messages only -- no direct state mutation.
- **Rationale**: BP-05 Section 3 specifies this pattern explicitly. All convergence personas agreed (unanimous). The pattern isolates state per tab, enabling: (a) independent development and testing per tab, (b) tab state preservation when switching tabs (FR-123), (c) lazy loading where only the active tab processes messages.
- **Rejected alternatives**:
  - **Monolithic model (single struct with all tab state)**: Leads to a 500+ line `Update()` function, makes testing impossible in isolation, breaks tab state preservation. Rejected.
  - **Separate Bubble Tea programs per tab**: Cannot share a single terminal, cannot switch tabs without restarting programs. Rejected.
  - **Shared mutable state between tabs**: Violates Elm architecture principles, creates race conditions with async commands. Rejected.
- **Consequences**: Tab models cannot directly read each other's state. Data sharing (e.g., DocsView getting documents from StatusView's health load) goes through the `App` model, which routes `healthLoadedMsg` to both StatusView and DocsView. This adds a small amount of message routing code to `App.Update()` but keeps tab boundaries clean.
- **Convergence adoption**: Directly implements convergence consensus area 1 ("Per-tab delegated model architecture").

---

### Decision 3: Components as Pure View Functions, Not tea.Models

- **Choice**: The `tui/components/` sub-package contains pure rendering functions (take data + dimensions, return styled strings), NOT `tea.Model` implementations. Only the 5 tab views and the help overlay are full `tea.Model`s.
- **Rationale**: BP-05 Section 3 lists 30+ components in the hierarchy, but most are pure rendering concerns: zone bars, staleness lists, warning panels, filter bars. Making each one a `tea.Model` would mean 30+ `Init/Update/View` implementations where `Init` returns nil, `Update` returns the model unchanged, and only `View` does work. This is boilerplate without benefit. Pure functions are simpler to test and compose.
- **Rejected alternatives**:
  - **Every component is a tea.Model**: BP-05 shows them in a tree suggesting model delegation. However, the tree describes the visual hierarchy, not necessarily the model hierarchy. The zone bar does not need to handle key events or produce commands. Making it a model adds complexity. Rejected for YAGNI.
  - **No components package**: Inline all rendering in tab views. This would make tab `View()` functions 200+ lines each and prevent reuse (zone bars appear in both StatusView and potentially other views). Rejected.
- **Consequences**: Tab `View()` functions call component rendering functions, passing them slices of domain data. The tab `Update()` function handles all key events for its children. This is simpler than delegating update to 5-10 child models per tab.
- **Convergence adoption**: Adapts convergence R3 recommendation. The convergence recommended using Bubbles components (which are `tea.Model`s) for standard UI elements like `textinput`, `spinner`, `viewport`, and `table`. Those are retained as models because they genuinely handle input. Custom visual components (zone bars, filter bars, etc.) are pure functions.

---

### Decision 4: Deps Struct for Wiring, Package Vars Retained

- **Choice**: Introduce a `Deps` struct and `BuildDeps()` function. The existing package-level variables in `cmd/root.go` are retained and populated from `Deps`. The TUI uses `Deps` directly.
- **Rationale**: The ideal is to eliminate all package-level variables and pass `Deps` to every command handler. However, that requires modifying every command handler in `cmd/` -- a 10+ file refactoring that creates merge conflicts and is not required for Phase 2. The `Deps` struct provides the canonical grouping; the package vars provide backward compatibility. The TUI (the only new consumer) uses `Deps` from day one.
- **Rejected alternatives**:
  - **Full migration to Deps in every handler**: High-risk refactoring touching every cmd/ file. Rejected for Phase 2 scope management.
  - **TUI gets its own separate wiring function**: Creates the dual-wiring problem the convergence identified (challenge A->B-1). Adding a service in Phase 3 would require updates in two places. Rejected.
  - **Pass Deps as Cobra context**: Cobra supports context values, but they require type assertions at every use site, adding boilerplate. Rejected.
- **Consequences**: Two wiring patterns coexist temporarily. CLI handlers use package vars, TUI uses `Deps`. This is pragmatic tech debt that can be resolved in a future refactoring iteration. The `BuildDeps()` function is the single source of truth for service construction.
- **Convergence adoption**: Implements convergence recommendation R1 item 3 (wiring centralization) while respecting convergence falsifiability condition ("if fix requires changes to more than 8 files, reassess scope").

---

### Decision 5: Glamour for Preview, Custom Chart for Quality

- **Choice**: Use `glamour` for the Docs tab preview pane (FR-105). Build a custom ASCII chart component for the Quality tab (FR-114) rather than using `asciigraph`.
- **Rationale**: Glamour is the standard Charm ecosystem library for terminal markdown rendering and integrates directly with Lip Gloss styles. For the chart, BP-05 specifies exact characters (`●`, `─`, `╭`, `╯`, `╰`, `╮`) that differ from `asciigraph`'s character set. A custom implementation of 150-200 lines matches the spec exactly and avoids an external dependency for a single-use component. The chart needs interactive navigation (data point selection with `←`/`→`) which `asciigraph` does not support.
- **Rejected alternatives**:
  - **No Glamour, render raw markdown text**: Provides a poor preview experience. The point of the preview is to see formatted content. Rejected.
  - **Use `asciigraph` for the chart**: Characters do not match BP-05 spec. No interactive navigation support. Would still need a custom wrapper. Rejected per convergence R3 guidance.
  - **Skip the chart entirely, show a table**: Loses the visual impact that motivates the Quality tab. The chart is SHOULD-level but straightforward to implement. Retained.
- **Consequences**: Glamour adds a dependency, but it is in the Charm ecosystem already used by the project. The custom chart is self-contained in `tui/components/score_chart.go` and testable as a pure function.
- **Convergence adoption**: Follows convergence R3 ("Evaluate `asciigraph` for Quality tab chart; if chart characters do not match BP-05 spec, build a custom chart component").

---

### Decision 6: ViewState Enum for Loading States

- **Choice**: Every tab model contains a `viewState ViewState` field that tracks its data loading status. The `ViewState` enum has 4 values: `ViewLoading`, `ViewError`, `ViewEmpty`, `ViewReady`. The `View()` function switches on this state to render the appropriate content.
- **Rationale**: FR-120 requires four distinct visual states per tab. A consistent enum prevents ad-hoc boolean flags (`loading`, `hasData`, `hasError`) that create combinatorial state explosion. The enum enforces exactly one state at a time.
- **Rejected alternatives**:
  - **Boolean flags per view**: `loading bool`, `hasError bool`, `isEmpty bool`. These can be in contradictory states (loading=true, hasError=true). Rejected.
  - **Single global loading state**: Not all tabs load at the same time. Checks tab loads lazily. Quality tab may be empty while Status is ready. Per-tab state is required. Rejected.
- **Consequences**: `ViewState` is a presentation-layer type defined in `tui/app.go`, NOT in `domain/`. This follows the requirements-delta guidance that TabID and ViewState are presentation concerns excluded from the domain layer.

---

### Decision 7: TUI Handles Its Own Project Detection

- **Choice**: The `mind tui` command calls `resolveRoot()` and `BuildDeps()` directly in `runTUI()`, bypassing `PersistentPreRunE`. The `requiresProject()` function returns `false` for the `tui` command.
- **Rationale**: `PersistentPreRunE` creates a `Renderer` and populates package-level variables that the TUI does not use. The TUI needs `tea.WithAltScreen()` before any output. Running `PersistentPreRunE` first would create unnecessary objects and could interfere with the alternate screen setup. Direct wiring in `runTUI()` is cleaner.
- **Rejected alternatives**:
  - **Let PersistentPreRunE wire, then pass to TUI**: Creates unnecessary Renderer, populates unused package vars. Rejected for clarity.
  - **Add a special mode flag to PersistentPreRunE**: Adds conditional logic to an already complex function. Rejected.
- **Consequences**: The TUI command is self-contained: resolve root, build deps, create app, run program. This is clear and debuggable.

---

### Decision 8: Quality Tab Launches with Empty State

- **Choice**: If `QualityRepo.ReadLog()` returns an empty slice (no `quality-log.yml` or empty file), Tab 5 renders the empty state message. No error, no crash. The chart and analysis detail are simply not shown.
- **Rationale**: `QualityService` may not be fully implemented when Phase 2 development starts. The empty state (FR-115) is a defined, specified behavior that provides a graceful degradation path. This matches convergence recommendation R4 ("launch Tab 5 with the empty state").
- **Rejected alternatives**:
  - **Block Phase 2 until QualityService is implemented**: Creates a sequential dependency that delays the entire TUI. Rejected.
  - **Hide Tab 5 when no quality data exists**: BP-08 acceptance criteria require all 5 tabs. The tab must exist. Rejected.
- **Consequences**: Tab 5 is usable on day one (shows a message) and becomes fully functional when quality data exists. The chart and analysis components are built but only render when data is available.

---

## Updated Dependency Matrix

```
cmd/tui_cmd.go ────────> tui/app.go
cmd/tui_cmd.go ────────> cmd.BuildDeps()
cmd/tui_cmd.go ────────> domain
cmd/tui_cmd.go ────────> bubbletea

tui/app.go ────────────> tui/status.go, tui/docs.go, tui/iterations.go,
                          tui/checks.go, tui/quality.go, tui/help.go
tui/app.go ────────────> tui/styles.go, tui/keys.go, tui/messages.go
tui/app.go ────────────> cmd.Deps (service injection)
tui/app.go ────────────> bubbletea, lipgloss

tui/status.go ─────────> tui/components/*
tui/status.go ─────────> domain (ProjectHealth, ZoneHealth, WorkflowState)
tui/status.go ─────────> lipgloss

tui/docs.go ───────────> tui/components/*
tui/docs.go ───────────> domain (Document, Zone, SearchResults)
tui/docs.go ───────────> bubbles/textinput, bubbles/viewport
tui/docs.go ───────────> glamour
tui/docs.go ───────────> lipgloss

tui/iterations.go ─────> tui/components/*
tui/iterations.go ─────> domain (Iteration, RequestType)
tui/iterations.go ─────> lipgloss

tui/checks.go ─────────> tui/components/*
tui/checks.go ─────────> domain (ValidationReport, CheckResult, ReconcileResult)
tui/checks.go ─────────> bubbles/spinner
tui/checks.go ─────────> lipgloss

tui/quality.go ─────────> tui/components/*
tui/quality.go ─────────> domain (QualityEntry, QualityDimension)
tui/quality.go ─────────> lipgloss

tui/components/* ──────> domain (various types)
tui/components/* ──────> lipgloss

cmd/root.go (BuildDeps) > internal/service/*, internal/repo/fs/*
cmd/*.go ──────────────> (unchanged from Phase 1.5)

internal/repo/interfaces.go > domain (adds QualityRepo, DocRepo.Search)
internal/repo/fs/quality_repo.go > domain, yaml.v3
internal/repo/mem/quality_repo.go > domain

domain/quality.go ─────> Go stdlib only (fmt, time)
```

**Layer enforcement**: The `tui/` package is at the presentation layer. It imports `domain` (down), `internal/service` via `Deps` (down through injection), and `internal/repo` interfaces (down). It never imports `cmd/` business logic. The `cmd.Deps` struct is a data container, not behavior.

---

## Testing Strategy

### Model State Testing with teatest

Each tab model receives specific `tea.Msg` values and the resulting model state is asserted. This tests the `Update()` function without rendering.

```go
// Example: ChecksView lazy loading
func TestChecksView_LazyLoad(t *testing.T) {
    deps := testDeps() // uses mem/ repos
    view := NewChecksView(deps)

    // Initially in loading state
    assert.Equal(t, ViewLoading, view.viewState)

    // Simulate first activation (tab switch to Checks)
    model, cmd := view.Update(tabActivatedMsg{tab: TabChecks})
    assert.NotNil(t, cmd) // should dispatch validation command

    // Simulate validation complete
    model, _ = model.Update(validationCompleteMsg{reports: testReports()})
    assert.Equal(t, ViewReady, model.(ChecksView).viewState)
    assert.Len(t, model.(ChecksView).reports, 4) // docs, refs, config, reconcile
}
```

**Target**: 3-5 state transition tests per tab model. Focus on:
- Loading lifecycle: Loading -> Ready, Loading -> Error, Loading -> Empty
- Key event handling: tab-specific keys produce correct state changes
- Data filtering: zone/type filters produce correct filtered lists
- Lazy loading: Checks tab loads only on first activation

### Golden File Tests

One golden file test per tab at 80x24 standard size. Uses `teatest` to capture `View()` output and compare against a `.golden` file.

```go
func TestStatusView_Golden80x24(t *testing.T) {
    view := newTestStatusView(testHealth(), 80, 24)
    got := view.View()
    golden.RequireEqual(t, got)
}
```

Golden files capture visual regression. They are stored in `tui/testdata/` and updated via `go test -update`.

**Target**: 5 golden files (one per tab) at 80x24. Plus 1 for the help overlay and 1 for the "terminal too small" message. Total: 7 golden files.

### Component Function Tests

Pure rendering functions in `tui/components/` are tested as unit tests with known inputs and expected string outputs.

```go
func TestZoneBar_FullWidth(t *testing.T) {
    zh := domain.ZoneHealth{Zone: domain.ZoneSpec, Total: 5, Complete: 4}
    got := ZoneBar(zh, 40)
    assert.Contains(t, got, "spec/")
    assert.Contains(t, got, "4/5")
}
```

### Coverage Target

- `tui/` `Update()` functions: >= 60% (NFR-15)
- `tui/` `View()` functions: excluded from coverage requirements (tested via golden files)
- `tui/components/`: >= 80% (pure functions, easy to test)
- `domain/quality.go`: 100% (matches domain/ coverage standard)

---

## Migration Path

The following 12-step sequence produces independently testable, compilable increments. Each step has a verification gate.

### Step 1: FR-88 -- Mutual Exclusion Guard

**Files changed**: `cmd/reconcile.go`
**Change**: Add 4-line guard at top of `runReconcile()`
**Test**: `TestReconcile_CheckAndForce_MutualExclusion` -- verify exit 2 and error message
**Gate**: `go test ./cmd/...` passes

### Step 2: FR-89 -- Missing Documents Check

**Files changed**: `internal/validate/reconcile.go`
**Change**: Add missing documents check between cycle check and stale checks
**Test**: `TestReconcileSuite_MissingDocuments` -- verify WARN-level check result for missing docs
**Gate**: `go test ./internal/validate/...` passes

### Step 3: FR-91 -- DocRepo Search Abstraction

**Files changed**: `internal/repo/interfaces.go`, `internal/repo/fs/doc_repo.go`, `internal/repo/mem/doc_repo.go`, `cmd/docs.go`
**Change**: Add `Search()` to DocRepo interface; move search logic from `cmd/docs.go` into `fs.DocRepo`; update `runDocsSearch()` to use `docRepo.Search()`
**Test**: `TestDocRepo_Search` with in-memory repo; verify identical results to current behavior
**Gate**: `go test ./internal/repo/...` and `go test ./cmd/...` pass; `mind docs search "test"` produces identical output

### Step 4: FR-90 -- Wiring Centralization

**Files changed**: `cmd/root.go` (add `Deps` struct, `BuildDeps()` function; update `PersistentPreRunE` to call it)
**Change**: Extract wiring into `BuildDeps()`; package vars populated from `Deps`
**Test**: All existing tests pass unchanged (backward compatible)
**Gate**: `go test ./...` passes with no behavioral changes

### Step 5: Domain Types + QualityRepo

**Files changed**: `domain/quality.go` (new), `internal/repo/interfaces.go`, `internal/repo/fs/quality_repo.go` (new), `internal/repo/mem/quality_repo.go` (new)
**Change**: Add `QualityEntry`, `QualityDimension`, `QualityRepo` interface, fs/mem implementations
**Test**: `TestQualityEntry_Validate` (BR-36/37/38); `TestQualityRepo_ReadLog` (fs, with fixture); `TestQualityRepo_Empty` (returns empty slice)
**Gate**: `go test ./domain/...` and `go test ./internal/repo/...` pass; `domain/purity_test.go` passes

### Step 6: TUI Foundation -- styles, keys, messages

**Files created**: `tui/styles.go`, `tui/keys.go`, `tui/messages.go`
**Change**: Theme constants, key binding definitions, custom message types
**Test**: Compile check (no runtime tests for constants)
**Gate**: `go build ./tui/...` succeeds

### Step 7: TUI App Shell + Status Tab

**Files created**: `tui/app.go`, `tui/status.go`, `tui/statusbar.go`, `tui/help.go`, `tui/components/zone_bar.go`, `tui/components/staleness.go`, `tui/components/workflow_panel.go`, `tui/components/warnings.go`, `tui/components/suggestions.go`, `tui/components/quick_actions.go`, `tui/components/empty_state.go`
**Files created**: `cmd/tui_cmd.go`
**Change**: Full app shell (title bar, tab bar, status bar, help overlay) + StatusView with all panels
**Test**: `TestApp_TabSwitch`, `TestApp_Quit`, `TestStatusView_HealthLoaded`, `TestStatusView_Golden80x24`
**Gate**: `mind tui` launches, shows Status tab with mock data, quit works; `go test ./tui/...` passes

### Step 8: Documents Tab

**Files created**: `tui/docs.go`, `tui/components/zone_filter.go`
**Change**: DocsView with document list, zone filter, search, preview pane
**Test**: `TestDocsView_ZoneFilter`, `TestDocsView_Search`, `TestDocsView_Preview`, `TestDocsView_Golden80x24`
**Gate**: Tab 2 renders document list, filter and search work; `go test ./tui/...` passes

### Step 9: Iterations Tab

**Files created**: `tui/iterations.go`, `tui/components/type_filter.go`, `tui/components/detail_expander.go`
**Change**: IterationsView with table, type filter, detail expander
**Test**: `TestIterationsView_TypeFilter`, `TestIterationsView_Expand`, `TestIterationsView_Golden80x24`
**Gate**: Tab 3 renders iteration table, filter and expand work; `go test ./tui/...` passes

### Step 10: Checks Tab

**Files created**: `tui/checks.go`, `tui/components/suite_accordion.go`, `tui/components/check_detail.go`, `tui/components/overall_summary.go`
**Change**: ChecksView with accordion, lazy loading, spinner, check detail, overall summary
**Test**: `TestChecksView_LazyLoad`, `TestChecksView_AccordionToggle`, `TestChecksView_ReRun`, `TestChecksView_Golden80x24`
**Gate**: Tab 4 runs validation on first visit, accordion works; `go test ./tui/...` passes

### Step 11: Quality Tab

**Files created**: `tui/quality.go`, `tui/components/score_chart.go`, `tui/components/latest_analysis.go`
**Change**: QualityView with chart, analysis detail, empty state
**Test**: `TestQualityView_EmptyState`, `TestQualityView_ChartRender`, `TestQualityView_PointNavigation`, `TestQualityView_Golden80x24`
**Gate**: Tab 5 shows empty state or chart depending on data; `go test ./tui/...` passes

### Step 12: Polish and Integration

**Change**: Terminal resize handling (FR-119), "too small" message (FR-118), tab state preservation verification (FR-123), terminal cleanup on quit/crash (FR-124), responsive column layout for Status tab, full help overlay with context-sensitive keys
**Test**: `TestApp_Resize`, `TestApp_TooSmall`, `TestApp_TabStatePreservation`, golden file for "too small" message
**Gate**: Full acceptance criteria from BP-08 Section 4. `go test ./...` passes. All 5 tabs render at 80x24. Tab switching preserves state. Quit restores terminal.

---

## Summary

Phase 2 adds 2 packages (`tui/`, `tui/components/`) with an estimated 10 tab/app files and 16 component files, plus 4 prerequisite fixes to existing code. The 4-layer architecture is preserved -- the TUI is a new presentation layer consuming existing service interfaces. Domain purity is maintained -- only `QualityEntry` and `QualityDimension` are added to `domain/`, with pure validation functions. The 12-step migration path ensures each change is independently testable and does not break existing functionality.
