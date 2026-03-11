# Blueprint: TUI Specification

> What does every screen look like and how does the user interact with it? Complete visual specification for the Bubble Tea TUI, covering all screen layouts, navigation, component hierarchy, state management, responsive design, theming, and variant TUI modes.

**Status**: Active
**Date**: 2026-03-11
**Depends on**: [01-mind-cli.md](01-mind-cli.md), [03-architecture.md](03-architecture.md)

---

## 1. Screen Layouts

All wireframes assume a standard 96-column terminal width. The outer chrome (title bar, tab bar, bottom border) is shared across all tabs.

### Shared Chrome

```
╭─ Mind Framework ─── {project-name} ── {branch} ──────────── v{version} ─╮
│                                                                            │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]  q:quit    │
│ ────────────────────────────────────────────────────────────────────────── │
│                                                                            │
│  {tab content area — variable height, scrollable}                          │
│                                                                            │
╰────────────────────────────────────────────────────────────────────────────╯
```

- Title bar: project name from `mind.toml`, current git branch, framework version.
- Tab bar: active tab rendered bold+underline, inactive tabs rendered dim. Number prefixes serve as keyboard shortcuts.
- Content area: fills remaining height. Scrollable via viewport when content exceeds terminal height.

---

### Tab 1: Status

Two-column layout. Left column: documentation health and diagnostics. Right column: workflow state and quick actions.

```
╭─ Mind Framework ─── mind-cli ── main ─────────────────────── v2026-03-09 ─╮
│                                                                             │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   q:quit    │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Documentation Health            │  Active Workflow                         │
│  ───────────────────             │  ───────────────                         │
│  spec/        ████████░░  4/5    │  State: running                         │
│  blueprints/  ██████████  3/3    │  Type: NEW_PROJECT                      │
│  state/       █████░░░░░  1/2    │  Agent: architect → developer           │
│  iterations/  ██████████  6/6    │  Branch: new/rest-api                   │
│  knowledge/   ██████░░░░  3/5    │                                          │
│                                   │  Quick Actions                          │
│  Staleness                       │  ──────────                              │
│  ──────────                      │  c  Create document                      │
│  2 stale (requirements changed)  │  d  Run doctor                          │
│  ● architecture.md               │  v  Validate all                         │
│  ● domain-model.md               │  r  Reconcile                           │
│                                   │  o  Open document                       │
│  Warnings (2)                    │                                          │
│  ────────────                    │                                          │
│  ⚠ domain-model.md is a stub    │                                          │
│  ⚠ glossary.md missing           │                                          │
│                                   │                                          │
│  Suggestions                     │                                          │
│  ───────────                     │                                          │
│  → Run 'mind doctor' for details │                                          │
│  → Fill domain-model.md          │                                          │
╰──────────────────────────────────────────────────────────────────────────── ╯
```

**Layout rules:**
- Left column occupies 50% of the content width, right column the remaining 50%.
- Progress bars scale proportionally to available column width (minimum 10 chars, maximum 20 chars).
- Staleness section appears only when stale documents exist. Stale documents are those whose upstream dependencies have been modified more recently.
- Warnings section appears only when warnings exist. Each warning uses `⚠` prefix.
- Suggestions section appears only when suggestions exist. Each uses `→` prefix.
- When workflow is idle, the right panel shows "State: idle" and the last completed iteration instead of the active chain.

---

### Tab 2: Documents

Full-width list of all documents grouped by zone, with status indicators, zone filter bar, and search capability.

```
╭─ Mind Framework ─── mind-cli ── main ─────────────────────── v2026-03-09 ─╮
│                                                                             │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   q:quit    │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Zone: [a All]  s Spec  b Blueprints  t State  i Iterations  k Knowledge   │
│  Search: /                                                                  │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  spec/                                                                      │
│  ├── project-brief.md          ✓ draft       2026-03-01     3.2 KB         │
│  ├── requirements.md           ✓ draft       2026-03-05     8.1 KB         │
│  ├── architecture.md           ✓ draft       2026-03-05     6.4 KB         │
│  ├── domain-model.md           ✗ stub        2026-02-28     0.4 KB         │
│  ├── api-contracts.md          ✗ stub        2026-03-01     0.6 KB         │
│  └── decisions/                                                             │
│      ├── 001-use-postgresql.md ✓ complete    2026-03-02     1.8 KB         │
│      └── 002-jwt-auth.md       ✓ complete    2026-03-04     2.1 KB         │
│                                                                             │
│  blueprints/                                                                │
│  ├── INDEX.md                  ✓ active      2026-03-10     0.9 KB         │
│  ├── 01-mind-cli.md            ✓ active      2026-03-09     4.2 KB         │
│  ├── 02-ai-workflow-bridge.md  ✓ active      2026-03-09     3.7 KB         │
│  └── 03-architecture.md        ✓ active      2026-03-09     6.1 KB         │
│                                                                             │
│  state/                                                                     │
│  ├── current.md                ✓ active      2026-03-10     1.1 KB         │
│  └── workflow.md               ✓ active      2026-03-10     0.8 KB         │
│                                                                             │
│  ↑↓ navigate  Enter preview  e edit  / search  Esc clear         3/22 docs │
╰──────────────────────────────────────────────────────────────────────────── ╯
```

**Layout rules:**
- Zone filter bar along the top. Active zone filter is rendered bold with brackets. Letter shortcuts shown inline.
- Search bar appears below the zone filter. When active (`/` pressed), a text input cursor appears and the list filters in real time.
- Documents listed in a tree structure per zone. Each row shows: filename, status indicator (`✓` for content, `✗` for stub/missing), status label (from `mind.toml` or inferred), modification date, file size.
- Currently selected row is highlighted with reverse video.
- Stub documents are rendered in the error color (red). Stale documents are rendered in the warning color (yellow).
- Bottom bar shows available keys and cursor position ("3/22 docs" means cursor on row 3 of 22 total).
- When a zone filter is active, only documents in that zone are shown.
- When a search is active, all zones are searched and results are flattened (no tree grouping).

**Preview pane:**
When the user presses Enter on a document, a right-side preview pane opens showing a markdown-rendered preview (via Glamour). The document list shrinks to 40% width, the preview pane takes 60%.

```
│  spec/                        │  # Project Brief                            │
│  ├── project-brief.md    ✓    │                                             │
│ >├── requirements.md     ✓    │  ## Vision                                  │
│  ├── architecture.md     ✓    │                                             │
│  ├── domain-model.md     ✗    │  A single-binary CLI and TUI tool that      │
│  └── api-contracts.md    ✗    │  provides unified project intelligence      │
│                               │  for the Mind Agent Framework...            │
│  blueprints/                  │                                             │
│  ├── INDEX.md            ✓    │  ## Target Users                            │
│  ...                          │  ...                                        │
│                               │                                             │
│  Esc close preview                  ↑↓ scroll preview  PgUp/PgDn page      │
```

---

### Tab 3: Iterations

Table layout with type filtering and expandable details.

```
╭─ Mind Framework ─── mind-cli ── main ─────────────────────── v2026-03-09 ─╮
│                                                                             │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   q:quit    │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Filter: [a All]  n NEW  e ENH  b BUG  r REFACTOR                         │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│   #    Type           Name                     Status       Date     Files  │
│  ──── ────────────── ──────────────────────── ────────── ────────── ─────── │
│  006  ENHANCEMENT    add-caching              ✓ complete  2026-03-08  5/5   │
│  005  BUG_FIX        fix-auth-redirect        ✓ complete  2026-03-07  5/5   │
│  004  REFACTOR       extract-repositories     ○ incomplete 2026-03-06  4/5  │
│  003  ENHANCEMENT    websocket-notifications  ✓ complete  2026-03-05  5/5   │
│  002  ENHANCEMENT    role-based-access        ✓ complete  2026-03-03  5/5   │
│  001  NEW_PROJECT    initial-api              ✓ complete  2026-03-01  5/5   │
│                                                                             │
│                                                                             │
│                                                                             │
│                                                                             │
│                                                                             │
│                                                                             │
│  ↑↓ navigate  Enter expand  o overview  v validation  n/e/b/r filter  6 it │
╰──────────────────────────────────────────────────────────────────────────── ╯
```

**Layout rules:**
- Filter bar along the top. Active filter is rendered bold with brackets.
- Table with fixed columns: `#` (4 chars), Type (14 chars), Name (24 chars, truncated with ellipsis), Status (10 chars), Date (10 chars), Files (7 chars).
- Selected row is highlighted with reverse video.
- Status indicators: `✓` complete (green), `▸` in_progress (yellow), `○` incomplete (dim).
- Type column is color-coded: NEW_PROJECT=blue, ENHANCEMENT=cyan, BUG_FIX=red, REFACTOR=magenta.
- Bottom bar shows available keys and total count.

**Expanded detail view:**
When the user presses Enter on an iteration, the row expands inline to show artifact status:

```
│  006  ENHANCEMENT    add-caching              ✓ complete  2026-03-08  5/5   │
│       ┌──────────────────────────────────────────────────────────────────┐  │
│       │  overview.md       ✓  2.1 KB    changes.md        ✓  1.8 KB    │  │
│       │  test-summary.md   ✓  3.4 KB    validation.md     ✓  2.9 KB    │  │
│       │  retrospective.md  ✓  1.2 KB                                    │  │
│       │  Branch: enhancement/add-caching                                │  │
│       └──────────────────────────────────────────────────────────────────┘  │
│  005  BUG_FIX        fix-auth-redirect        ✓ complete  2026-03-07  5/5   │
```

---

### Tab 4: Checks

Live validation results with expandable sections per validation suite.

```
╭─ Mind Framework ─── mind-cli ── main ─────────────────────── v2026-03-09 ─╮
│                                                                             │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   q:quit    │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Validation Results                                           r: re-run     │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  ▾ Documentation (17 checks)                    Pass: 15  Fail: 1  Warn: 1 │
│    ✓  [1]  docs/ directory exists                                           │
│    ✓  [2]  All 5 zone directories exist                                     │
│    ✓  [3]  Required spec files                                              │
│    ✓  [4]  decisions/ subdirectory                                          │
│    ✓  [5]  ADR naming convention                                            │
│    ✓  [6]  blueprints/INDEX.md                                              │
│    ✓  [7]  Blueprint → INDEX.md coverage                                    │
│    ✓  [8]  INDEX.md → file references                                       │
│    ✓  [9]  state/current.md                                                 │
│    ✓ [10]  state/workflow.md                                                │
│    ✓ [11]  knowledge/glossary.md                                            │
│    ✓ [12]  Iteration folder naming                                          │
│    ✓ [13]  Iterations have overview.md                                      │
│    ✓ [14]  Spike file naming                                                │
│    ✓ [15]  No legacy paths                                                  │
│    ✗ [16]  Stub detection — domain-model.md is a stub                       │
│    ⚠ [17]  Project brief completeness — missing Key Deliverables section    │
│                                                                             │
│  ▸ Cross-References (11 checks)                 Pass: 11  Fail: 0  Warn: 0 │
│                                                                             │
│  ▸ Config (4 checks)                            Pass: 4   Fail: 0  Warn: 0 │
│                                                                             │
│ ─────────────────────────────────────────────────────────────────────────── │
│  Overall: 30/32 pass    1 fail    1 warning                                 │
│                                                                             │
│  ↑↓ navigate  Enter expand/collapse  r re-run  Space toggle detail          │
╰──────────────────────────────────────────────────────────────────────────── ╯
```

**Layout rules:**
- Each validation suite is an accordion section. `▾` expanded, `▸` collapsed.
- Suite header shows: name, total checks in parentheses, and pass/fail/warn summary right-aligned.
- When expanded, each check is listed with its result indicator: `✓` pass (green), `✗` fail (red), `⚠` warning (yellow).
- Failed checks and warnings include an inline detail message after the em dash.
- Selected suite header is highlighted with reverse video.
- Space toggles a detail pane below a selected check (shows the full diagnostic message with fix suggestions).
- Overall summary bar at the bottom: total pass count, fail count, warning count.
- Validation runs automatically when the tab is first opened. Spinner shown while loading ("Running validation...").
- `r` re-runs all validation suites.

**Detail toggle (Space on a failed check):**

```
│    ✗ [16]  Stub detection — domain-model.md is a stub                       │
│            ┌─────────────────────────────────────────────────────────────┐  │
│            │  File: docs/spec/domain-model.md                           │  │
│            │  Issue: Contains only headings, comments, and placeholders │  │
│            │  Fix: Fill the Entities table and Business Rules section   │  │
│            │       or run '/discover' to generate domain context        │  │
│            └─────────────────────────────────────────────────────────────┘  │
```

---

### Tab 5: Quality

Convergence score trends over time with an ASCII chart and latest analysis details.

```
╭─ Mind Framework ─── mind-cli ── main ─────────────────────── v2026-03-09 ─╮
│                                                                             │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   q:quit    │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Overall Score History                                                      │
│  ─────────────────────                                                      │
│                                                                             │
│  5.0 ┤                                                                      │
│  4.5 ┤                                                                      │
│  4.0 ┤                                                    ╭──●              │
│  3.5 ┤                                ╭──●────●───────────╯                 │
│  3.0 ┤─ ─ ─ ─ ─ ─ ─ ─●──────●───────╯─ ─ ─ ─ ─ ─ ─ ─ ─  Gate 0          │
│  2.5 ┤            ╭───╯                                                     │
│  2.0 ┤       ●───╯                                                          │
│  1.5 ┤                                                                      │
│  1.0 ┤                                                                      │
│      └──────────────────────────────────────────────────────────────         │
│       Feb 25   Feb 27   Mar 01   Mar 03   Mar 05   Mar 07   Mar 09         │
│                                                                             │
│ ─────────────────────────────────────────────────────────────────────────── │
│                                                                             │
│  Latest Analysis                             Selected: ● Mar 09 (score 4.0) │
│  ───────────────                                                            │
│  Topic: auth-strategy    Variant: v2         Gate 0: ✓ PASS                 │
│                                                                             │
│  Dimensions                                                                 │
│  ──────────                                                                 │
│  Rigor          ████████░░  4/5              Personas: security-analyst,    │
│  Coverage       ████████░░  4/5                        api-designer,        │
│  Actionability  ██████████  5/5                        performance-eng      │
│  Objectivity    ██████░░░░  3/5                                             │
│  Convergence    ████████░░  4/5              Output: docs/knowledge/        │
│  Depth          ████████░░  4/5                auth-strategy-convergence.md │
│                                                                             │
│  ←→ scroll timeline  Enter details for point  ↑↓ scroll dimensions          │
╰──────────────────────────────────────────────────────────────────────────── ╯
```

**Layout rules:**
- Top half: ASCII line chart of overall convergence scores over time. Y-axis: 1.0 to 5.0. X-axis: dates. Gate 0 threshold (3.0) shown as a dashed horizontal line labeled "Gate 0".
- Data points rendered as `●`. Connected by lines (`─`, `╭`, `╯`, `╰`, `╮`).
- Chart width scales to available terminal width. Date labels spaced proportionally.
- Bottom half: details for the selected data point. Includes topic, variant, gate result, dimension scores as progress bars, personas used, and output path.
- When no quality data exists (no `quality-log.yml`), show an empty state: "No quality data. Run a convergence analysis and then `mind quality log <file>` to start tracking."
- Selected data point is indicated by the current `←→` position on the timeline.
- Dimension progress bars scale to available width, minimum 10 chars.

---

## 2. Navigation Model

### Global Keys

These keys work in every tab and are never overridden by tab-specific handlers.

| Key | Action | Context |
|-----|--------|---------|
| `1` | Switch to Status tab | Always |
| `2` | Switch to Docs tab | Always |
| `3` | Switch to Iterations tab | Always |
| `4` | Switch to Checks tab (triggers validation if not yet run) | Always |
| `5` | Switch to Quality tab | Always |
| `Tab` | Next tab (wraps from 5 to 1) | Always |
| `Shift+Tab` | Previous tab (wraps from 1 to 5) | Always |
| `q` | Quit application | When no modal overlay is open |
| `Ctrl+C` | Force quit | Always |
| `?` | Toggle help overlay | Always |
| `r` | Refresh all data (re-load from disk) | When no text input is focused |

### Tab-Specific Keys

#### Status Tab

| Key | Action |
|-----|--------|
| `c` | Create document — opens a sub-menu with document types (adr, blueprint, iteration, spike, convergence, brief) |
| `d` | Run doctor diagnostics — switches to Checks tab and runs full doctor |
| `v` | Validate all — switches to Checks tab and runs all validation suites |
| `r` | Reconcile — re-run staleness detection and update health |
| `o` | Open document — prompts for document selection, opens in `$EDITOR` |

#### Docs Tab

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up in document list |
| `↓` / `j` | Move cursor down in document list |
| `Enter` | Toggle preview pane for selected document |
| `e` | Open selected document in `$EDITOR` (suspends TUI, resumes on exit) |
| `/` | Activate search input — filters document list in real time |
| `Esc` | Clear search filter and close preview pane |
| `s` | Filter to spec zone |
| `b` | Filter to blueprints zone |
| `t` | Filter to state zone |
| `i` | Filter to iterations zone |
| `k` | Filter to knowledge zone (only when search is not focused) |
| `a` | Show all zones (clear filter) |
| `PgUp` | Page up in preview pane (when open) |
| `PgDn` | Page down in preview pane (when open) |

**Key disambiguation:** `k` serves as both "move up" (vim binding) and "knowledge zone filter". When the document list is focused and no zone filter input is active, `k` moves up. Zone filter shortcuts are processed only when the user has not typed a cursor key recently (debounced) or can be accessed via the zone filter bar. In practice, the zone shortcuts are single-press when the list has focus; vim-style up/down is the alternative to arrow keys. If this creates conflict, `K` (uppercase) activates the knowledge filter.

#### Iterations Tab

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up in iteration list |
| `↓` / `j` | Move cursor down in iteration list |
| `Enter` | Expand/collapse selected iteration details (artifact list) |
| `o` | Open overview.md for selected iteration in `$EDITOR` |
| `v` | Open validation.md for selected iteration in `$EDITOR` |
| `n` | Filter to NEW_PROJECT iterations |
| `e` | Filter to ENHANCEMENT iterations |
| `b` | Filter to BUG_FIX iterations |
| `r` | Filter to REFACTOR iterations |
| `a` | Show all types (clear filter) |

#### Checks Tab

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up (across suite headers and check rows) |
| `↓` / `j` | Move cursor down |
| `Enter` | Expand/collapse selected suite |
| `Space` | Toggle detail pane for selected check (shows full diagnostic) |
| `r` | Re-run all validation suites |

#### Quality Tab

| Key | Action |
|-----|--------|
| `←` / `h` | Select previous data point on timeline |
| `→` / `l` | Select next data point on timeline |
| `Enter` | Show full analysis details for selected data point |
| `↑` / `k` | Scroll up in dimension details (when details area overflows) |
| `↓` / `j` | Scroll down in dimension details |

### Focus Management

The TUI has three focus layers:

1. **Tab focus**: Which tab is active. Controlled by number keys and Tab/Shift+Tab. Tab switching always resets focus to layer 2 of the new tab.

2. **Component focus**: Within a tab, which component has focus. Managed per-tab:
   - Docs tab: document list (default) or preview pane (when open)
   - Checks tab: suite list (default) or expanded check list
   - Quality tab: timeline (default) or dimension list

3. **Modal overlay**: Help overlay (`?`) or create document sub-menu. When a modal is open, all keys except `Esc` and `?` are consumed by the modal. `Esc` closes the modal. `q` does not quit when a modal is open.

Focus transitions:
- `Enter` on Docs tab with preview closed: opens preview, focus stays on document list (arrow keys still navigate docs).
- `Tab` within Docs tab when preview is open: toggles focus between document list and preview pane (for scrolling the preview).
- `Esc` on any tab: dismisses the deepest active overlay/filter/input, then returns to default focus.

### Help Overlay

Pressing `?` shows a centered overlay (60x20 characters) listing all available keybindings for the current tab, plus global keys. The overlay renders with a border and title "Keyboard Shortcuts".

```
╭─ Keyboard Shortcuts ──────────────────────────────╮
│                                                     │
│  Global                                             │
│  1-5         Switch tab                             │
│  Tab         Next tab                               │
│  Shift+Tab   Previous tab                           │
│  q           Quit                                   │
│  ?           Toggle this help                       │
│  r           Refresh data                           │
│                                                     │
│  Status Tab                                         │
│  c           Create document                        │
│  d           Run doctor                             │
│  v           Validate all                           │
│  o           Open document                          │
│                                                     │
│                                Press ? or Esc to close │
╰─────────────────────────────────────────────────────╯
```

---

## 3. Component Hierarchy

The Bubble Tea model tree. Each node is a `tea.Model` with its own `Init`, `Update`, and `View` methods. The top-level `App` delegates messages to the active tab's model.

```
App (tui/app.go)
│
├── HeaderBar (tui/header.go)
│   Renders: project name, branch, version
│   Data: Project (static after init)
│
├── TabBar (tui/tabs.go)
│   Renders: [1 Status] [2 Docs] [3 Iterations] [4 Check] [5 Quality]
│   State: activeTab (Tab enum)
│   Keys: 1-5, Tab, Shift+Tab
│
├── StatusView (tui/status.go)
│   ├── HealthPanel (tui/components/health_panel.go)
│   │   ├── ZoneHealthBar (tui/components/zone_bar.go)
│   │   │   Renders: zone label + progress bar + fraction
│   │   │   Data: ZoneHealth
│   │   │   Repeated: 5 times (one per zone)
│   │   │
│   │   └── StalenessSection (tui/components/staleness.go)
│   │       Renders: stale document list with ● bullets
│   │       Data: []Document (filtered to stale)
│   │
│   ├── WorkflowPanel (tui/components/workflow_panel.go)
│   │   Renders: state, type, agent chain, branch
│   │   Data: *WorkflowState
│   │
│   ├── WarningsPanel (tui/components/warnings_panel.go)
│   │   Renders: warning list with ⚠ bullets
│   │   Data: []string (from ProjectHealth.Warnings)
│   │
│   ├── SuggestionsPanel (tui/components/suggestions_panel.go)
│   │   Renders: suggestion list with → bullets
│   │   Data: []string (from ProjectHealth.Suggestions)
│   │
│   └── QuickActionsPanel (tui/components/quick_actions.go)
│       Renders: key-action list
│       Static content
│
├── DocsView (tui/docs.go)
│   ├── ZoneFilter (tui/components/zone_filter.go)
│   │   Renders: [a All] s Spec b Blueprints t State i Iterations k Knowledge
│   │   State: activeZone (Zone | nil for all)
│   │   Keys: a, s, b, t, i, k
│   │
│   ├── SearchInput (bubbles/textinput)
│   │   Renders: "Search: " + input field
│   │   State: query string, focused bool
│   │   Keys: / to activate, Esc to clear
│   │
│   ├── DocsList (bubbles/list — customized)
│   │   Renders: tree-structured document list with status indicators
│   │   Data: []Document (filtered by zone + search)
│   │   State: cursor index, selected item
│   │   Keys: ↑↓jk navigate, Enter select, e edit
│   │
│   └── PreviewPane (bubbles/viewport)
│       Renders: Glamour-rendered markdown content
│       Data: rendered string (from selected document)
│       State: viewport scroll position, visible bool
│       Keys: ↑↓ scroll, PgUp/PgDn page, Esc close
│
├── IterationsView (tui/iterations.go)
│   ├── TypeFilter (tui/components/type_filter.go)
│   │   Renders: [a All] n NEW e ENH b BUG r REFACTOR
│   │   State: activeType (RequestType | nil for all)
│   │   Keys: a, n, e, b, r
│   │
│   ├── IterationTable (bubbles/table — customized)
│   │   Renders: table with columns: #, Type, Name, Status, Date, Files
│   │   Data: []Iteration (filtered by type)
│   │   State: cursor index, selected row
│   │   Keys: ↑↓jk navigate, Enter expand
│   │
│   └── DetailExpander (tui/components/detail_expander.go)
│       Renders: inline artifact list below expanded row
│       Data: Iteration.Artifacts
│       State: expandedIndex (int, -1 for none)
│
├── ChecksView (tui/checks.go)
│   ├── SuiteAccordion (tui/components/suite_accordion.go)
│   │   Renders: collapsible sections per validation suite
│   │   Data: []ValidationReport
│   │   State: expandedSuites (set of indices), cursor position
│   │   Keys: ↑↓jk navigate, Enter expand/collapse
│   │   Children:
│   │   ├── SuiteHeader (inline)
│   │   │   Renders: ▾/▸ icon + suite name + (N checks) + pass/fail/warn
│   │   └── CheckRow (inline)
│   │       Renders: ✓/✗/⚠ + [ID] + name + detail message
│   │
│   ├── CheckDetail (tui/components/check_detail.go)
│   │   Renders: bordered box with file, issue, and fix suggestion
│   │   Data: CheckResult (selected)
│   │   State: visible bool
│   │   Keys: Space toggle
│   │
│   ├── OverallSummary (tui/components/overall_summary.go)
│   │   Renders: "Overall: N/M pass  X fail  Y warning"
│   │   Data: aggregated from all ValidationReports
│   │
│   └── Spinner (bubbles/spinner)
│       Renders: "Running validation..." with animated spinner
│       State: visible during validation load
│
├── QualityView (tui/quality.go)
│   ├── ScoreChart (tui/components/score_chart.go)
│   │   Renders: ASCII line chart with Y-axis labels, data points, Gate 0 line
│   │   Data: []QualityEntry (time-ordered)
│   │   State: selectedPointIndex
│   │   Keys: ←→hl navigate points
│   │
│   ├── LatestAnalysis (tui/components/latest_analysis.go)
│   │   Renders: topic, variant, gate result, dimension bars, personas, path
│   │   Data: QualityEntry (selected point)
│   │   State: scrollOffset (for dimension list overflow)
│   │   Keys: ↑↓jk scroll
│   │
│   └── EmptyState (tui/components/empty_state.go)
│       Renders: "No quality data..." message
│       Visible: when QualityEntries is empty
│
├── HelpOverlay (tui/help.go)
│   Renders: centered modal with keybinding reference
│   State: visible bool, contextual keys based on activeTab
│   Keys: ? toggle, Esc close
│
└── StatusBar (tui/statusbar.go)
    Renders: contextual key hints for active tab + cursor position
    Data: derived from activeTab + component state
```

### Component Communication

Components communicate exclusively through Bubble Tea messages. A child never mutates parent state directly. The flow is:

1. User presses a key.
2. `App.Update` receives `tea.KeyMsg`.
3. If it is a global key, `App` handles it directly (tab switch, quit, help).
4. Otherwise, `App` delegates to the active tab's `Update`.
5. The tab may return a `tea.Cmd` that produces a message (e.g., `healthLoadedMsg`).
6. `App.Update` handles the message and distributes data to relevant child models.

---

## 4. State Management

### Data Requirements per View

| View | Domain Data | Source | Load Trigger |
|------|-------------|--------|--------------|
| StatusView | `ProjectHealth` | `ProjectService.Health()` | App init, manual refresh (`r`) |
| DocsView | `[]Document` | Derived from `ProjectHealth.Zones` | Propagated from StatusView load |
| DocsView (preview) | `string` (rendered markdown) | `DocRepo.Read()` + Glamour render | User presses Enter on a document |
| IterationsView | `[]Iteration` | `IterationRepo.List()` | App init, manual refresh |
| ChecksView | `[]ValidationReport` | `ValidationService.RunAll()` | First switch to tab 4, manual re-run |
| QualityView | `[]QualityEntry` | `QualityRepo.ReadLog()` | App init, manual refresh |

### View-Local State

Each view maintains its own UI state that is not shared:

| View | Local State |
|------|-------------|
| DocsView | `activeZone` (zone filter), `searchQuery` (string), `cursorIndex` (int), `previewVisible` (bool), `previewContent` (string), `previewScroll` (viewport.Model) |
| IterationsView | `activeType` (type filter), `cursorIndex` (int), `expandedIndex` (int, -1 = none) |
| ChecksView | `expandedSuites` (map[int]bool), `cursorPosition` (int), `detailVisible` (bool), `detailTarget` (int), `loading` (bool) |
| QualityView | `selectedPointIndex` (int), `dimensionScrollOffset` (int) |

### Data Loading Strategy

**On init:**
```
App.Init() → tea.Batch(
    loadProjectHealth,     // Fetches ProjectHealth, populates StatusView + DocsView
    loadIterations,        // Fetches []Iteration, populates IterationsView
    loadQualityEntries,    // Fetches []QualityEntry, populates QualityView
)
```

All three commands run concurrently as `tea.Cmd` functions. Each returns a typed message:
- `healthLoadedMsg { health *ProjectHealth }` or `healthErrorMsg { err error }`
- `iterationsLoadedMsg { iterations []Iteration }` or `iterationsErrorMsg { err error }`
- `qualityLoadedMsg { entries []QualityEntry }` or `qualityErrorMsg { err error }`

**Lazy loading (Checks tab):**
Validation is not run on init (it can be slow). It runs when the user first switches to tab 4 or presses `r`.
- `validationStartedMsg {}` — triggers spinner
- `validationCompleteMsg { reports []ValidationReport }` or `validationErrorMsg { err error }`

**Preview loading (Docs tab):**
When the user presses Enter on a document, a `tea.Cmd` reads and renders the file:
- `previewLoadedMsg { content string }` or `previewErrorMsg { err error }`

**Manual refresh (`r` key):**
Re-runs the same init commands. While loading, existing data remains visible. When new data arrives, it replaces the old data atomically.

### Loading States

Every view that loads data asynchronously displays one of three states:

| State | Visual | Condition |
|-------|--------|-----------|
| Loading | Spinner + "Loading..." message | Data command dispatched, no response yet |
| Error | Red error message + "Press r to retry" | Error message received |
| Empty | Dim message explaining what is missing | Data loaded but collection is empty |
| Ready | Normal view rendering | Data loaded, non-empty |

**Empty state messages:**
- StatusView: "No project detected. Run `mind init` to get started."
- DocsView: "No documents found in docs/. Run `mind create` to scaffold."
- IterationsView: "No iterations yet. Start a workflow to create one."
- ChecksView: "Press Enter or `r` to run validation."
- QualityView: "No quality data. Run a convergence analysis and then `mind quality log <file>` to start tracking."

### No Auto-Polling

The TUI does not automatically refresh data. All updates are user-initiated:
- `r` key refreshes all data
- Tab 4 lazy-loads on first visit
- Preview loads on Enter

This design is intentional: the TUI reads from the filesystem and should show a consistent snapshot. Auto-polling would cause visual flicker and potential race conditions with concurrent file writes from AI agents. For real-time monitoring during workflows, use `mind watch --tui` instead (Section 7).

---

## 5. Responsive Design

### Terminal Size Handling

The TUI listens for `tea.WindowSizeMsg` on every resize event and propagates the new dimensions to all child components.

| Width | Layout Strategy |
|-------|----------------|
| < 80 columns | **Narrow mode**: single-column layout for all tabs. Status tab stacks panels vertically (health above workflow). Docs tab hides preview pane. Iterations table truncates Name column aggressively. |
| 80-99 columns | **Standard mode**: two-column layout for Status tab. Other tabs use full width. Progress bars at 10 chars. |
| >= 100 columns | **Wide mode**: two-column layout for Status tab with wider columns. Progress bars at 20 chars. Docs preview pane available. Iteration names untruncated. |

| Height | Layout Strategy |
|--------|----------------|
| < 24 rows | **Compact mode**: tab bar and status bar consume 4 rows. Content area gets remaining rows. Scrollable via viewport. Warning shown if height < 16. |
| 24-40 rows | **Standard mode**: full chrome. Content fills remaining space. |
| > 40 rows | **Tall mode**: extra space added as padding. Content area does not stretch beyond useful size. |

### Minimum Size

- **Minimum supported**: 80x24 (standard terminal).
- **Below minimum**: if terminal is smaller than 80x24, the TUI displays a single centered message: "Terminal too small. Minimum: 80x24. Current: {w}x{h}." and does not render the full interface.

### Component Scaling Rules

| Component | Scaling Behavior |
|-----------|-----------------|
| Progress bars | Width = `min(20, max(10, (availableWidth - labelWidth - fractionWidth) / 2))` |
| Tab bar | Always full width. Tab labels truncate to numbers only if width < 60 |
| Tables | Fixed columns use minimum widths. Name/description columns absorb remaining space. Truncate with `…` when content exceeds column width. |
| Two-column split | Status tab: 50/50 split. Docs tab with preview: 40/60 split. Columns reflow to single column below 80 chars. |
| ASCII chart | Width scales to `availableWidth - 8` (leaving room for Y-axis labels and padding). Minimum 40 chars for the plot area. Below that, a text summary replaces the chart. |
| Borders | Outer border always present. Inner panel borders omitted in narrow mode. |

### Resize Event Handling

```
tea.WindowSizeMsg received
    │
    ├── Update App.width, App.height
    ├── Propagate to all tab models (each receives the new dimensions)
    ├── Each component recalculates its layout
    └── View re-renders with new dimensions on next frame
```

No debouncing is needed because Bubble Tea batches renders. Rapid resize events produce rapid `WindowSizeMsg` messages, but only the final rendered frame is painted to the terminal.

---

## 6. Color and Styling (Lip Gloss Theme)

All styles are defined in a single file (`tui/styles.go`) as a `Theme` struct. This enables future theme switching (light/dark/monochrome).

### Zone Colors

Each documentation zone has a distinct color used for its label, progress bar fill, and zone filter indicator.

| Zone | Color | Lip Gloss Value | Usage |
|------|-------|-----------------|-------|
| spec | Blue | `lipgloss.Color("63")` / `#5f87ff` | Zone labels, progress bar fill, filter indicator |
| blueprints | Cyan | `lipgloss.Color("81")` / `#5fd7ff` | Zone labels, progress bar fill, filter indicator |
| state | Yellow | `lipgloss.Color("220")` / `#ffd700` | Zone labels, progress bar fill, filter indicator |
| iterations | Green | `lipgloss.Color("78")` / `#5fd787` | Zone labels, progress bar fill, filter indicator |
| knowledge | Magenta | `lipgloss.Color("170")` / `#d75fd7` | Zone labels, progress bar fill, filter indicator |

### Severity Colors

Used across all tabs for status indicators.

| Severity | Color | Lip Gloss Value | Symbol |
|----------|-------|-----------------|--------|
| Pass / Success | Green | `lipgloss.Color("78")` / `#5fd787` | `✓` |
| Warning | Yellow | `lipgloss.Color("220")` / `#ffd700` | `⚠` |
| Fail / Error | Red | `lipgloss.Color("196")` / `#ff0000` | `✗` |
| Info | Blue | `lipgloss.Color("63")` / `#5f87ff` | `ℹ` |

### Status Colors

Applied to document and iteration status labels.

| Status | Foreground | Style | Symbol |
|--------|-----------|-------|--------|
| active | Green | Normal | `✓` |
| draft | Dim white | `Faint(true)` | `✓` |
| complete | Green | Bold | `✓` |
| stub | Red | Normal | `✗` |
| stale | Yellow | Normal | `●` |
| fresh | Green | Normal | `●` |

### Workflow Status Colors

| Status | Color | Symbol |
|--------|-------|--------|
| running | Green | `▸` |
| idle | Dim white | `○` |
| completed | Green+Bold | `✓` |
| failed | Red | `✗` |
| waiting | Dim white | `○` |
| retrying | Yellow | `↻` |

### UI Element Styles

| Element | Style Definition |
|---------|-----------------|
| Active tab | `Bold(true).Underline(true).Foreground(Color("255"))` |
| Inactive tab | `Faint(true).Foreground(Color("245"))` |
| Borders | `Border(lipgloss.RoundedBorder()).BorderForeground(Color("240"))` |
| Headings | `Bold(true).Foreground(Color("255"))` |
| Sub-headings | `Bold(true).Foreground(Color("250"))` |
| Selected row | `Reverse(true)` |
| Dim text | `Faint(true)` |
| Status bar (bottom) | `Background(Color("236")).Foreground(Color("245"))` |
| Help overlay border | `Border(lipgloss.DoubleBorder()).BorderForeground(Color("63"))` |

### Progress Bar Styles

| Segment | Rendering |
|---------|-----------|
| Filled | `█` in zone color (or green for generic bars) |
| Empty | `░` in `Color("238")` (dark gray) |
| Partial | `▓` in yellow (only used when the bar represents partial content, e.g., stubs counted separately) |

Example rendering at 10-char width for 3/5:
```
██████░░░░  3/5
```

### Monochrome Fallback

When the terminal does not support colors (detected via `lipgloss.HasDarkBackground()` failing or `NO_COLOR` environment variable set), all color styling is stripped. The interface relies entirely on symbols and layout:

| Colored Rendering | Monochrome Rendering |
|-------------------|---------------------|
| Green `✓` | Plain `✓` |
| Red `✗` | Plain `✗` |
| Yellow `⚠` | Plain `⚠` |
| Bold active tab | `[Tab Name]` (brackets indicate active) |
| Reverse selected row | `> ` prefix indicator |
| Colored progress bar `████░░` | ASCII progress bar `[###...]` |

---

## 7. Watch TUI Variant

The Watch TUI is a separate Bubble Tea application launched by `mind watch --tui`. It provides real-time filesystem monitoring during AI workflows.

### Layout

```
╭─ mind watch ─── mind-cli ── new/rest-api ─────────────────────────────────╮
│                                                                             │
│  Workflow: NEW_PROJECT (007-rest-api)                                      │
│  Chain: analyst ✓ → architect ✓ → [developer] → tester → reviewer         │
│                                                                             │
│  ┌─ Live Activity ───────────────────────────────────────────────────────┐  │
│  │  14:23:01  Developer writing src/routes/users.rs                      │  │
│  │  14:23:15  Developer writing src/routes/auth.rs                       │  │
│  │  14:23:32  Developer created src/middleware/jwt.rs                     │  │
│  │  14:24:01  changes.md updated — 4 files added                         │  │
│  │  14:24:02  ▸ Micro-Gate B: checking...                                │  │
│  │  14:24:03  ▸ Micro-Gate B: ✓ changes.md exists, 4 files on disk      │  │
│  │  14:24:10  Developer writing src/models/user.rs                       │  │
│  │  14:24:45  ▸ Background: cargo build ✓ (3.2s)                         │  │
│  │  14:25:12  ▸ Background: cargo test ✓ 24 passed (4.1s)                │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌─ Gate Status ─────────────────────────────────────────────────────────┐  │
│  │  Build: ✓ passing    Lint: ⏳ pending    Tests: ✓ 24/24 passing       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  Watching: docs/, src/    Changes: 12 files    Uptime: 6m 47s              │
│  [q]uit  [p]ause  [c]lear log  [f]ollow (auto-scroll: on)                 │
╰─────────────────────────────────────────────────────────────────────────── ╯
```

### Data Sources

| Data | Source | Update Trigger |
|------|--------|----------------|
| Workflow state | `docs/state/workflow.md` | fsnotify: file modified |
| Agent chain progress | Parsed from workflow.md `LastAgent` + `RemainingChain` | fsnotify: workflow.md modified |
| Live activity log | All fsnotify events from watched paths | Every filesystem event (debounced 100ms) |
| Gate status (Build) | Background `go build ./...` or project command | After source file changes (debounced 2s) |
| Gate status (Lint) | Background `golangci-lint run` or project command | After source file changes (debounced 5s) |
| Gate status (Tests) | Background `go test ./...` or project command | After source file changes (debounced 5s) |
| Change count | Count of unique files modified since watch started | Accumulated from fsnotify events |
| Uptime | Wall clock since watch started | `tea.Tick` every 1s |

### Watch Paths

The watcher monitors these paths via `fsnotify`:

| Path Pattern | Event Type | Action |
|-------------|-----------|--------|
| `docs/state/workflow.md` | Modify | Parse workflow state, update chain display |
| `docs/iterations/*/overview.md` | Create | Log "New iteration detected", update chain |
| `docs/iterations/*/changes.md` | Modify | Log "changes.md updated", run Micro-Gate B |
| `docs/iterations/*/validation.md` | Modify | Log "validation.md updated", parse findings |
| `docs/spec/requirements.md` | Modify | Log "requirements updated", run Micro-Gate A |
| `docs/spec/architecture.md` | Modify | Log "architecture updated" |
| `docs/knowledge/*-convergence.md` | Create | Log "convergence complete", run 23-check validation |
| `src/**/*` (or project source) | Modify/Create | Log file change, trigger background build/test |
| `docs/spec/project-brief.md` | Modify | Log "brief updated", re-run business context gate |

### Activity Log Behavior

- Maximum 500 entries in memory. Older entries are discarded (ring buffer).
- Each entry: `HH:MM:SS  {message}`.
- Gate check results are prefixed with `▸` to distinguish them from file events.
- Auto-scroll is on by default. New entries appear at the bottom and the viewport scrolls to keep the latest visible.
- `f` toggles auto-scroll (follow mode). When off, the user can scroll manually with `↑↓/PgUp/PgDn`. When a new entry arrives with follow off, a "New activity ↓" indicator appears at the bottom of the log.
- `c` clears the log.

### Background Command Execution

Build, lint, and test commands are read from `mind.toml` (`[project.commands]`). They run in background goroutines with output captured. Results are sent as Bubble Tea messages:

- `gateResultMsg { gate string, passed bool, output string, duration time.Duration }`

Debouncing prevents rapid re-triggers: after a source file change, the watcher waits 2 seconds of inactivity before starting a build. If another change arrives during the wait, the timer resets.

### Watch TUI Keys

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit watcher |
| `p` | Pause/resume filesystem watching |
| `c` | Clear activity log |
| `f` | Toggle follow mode (auto-scroll) |
| `↑` / `k` | Scroll activity log up (when follow is off) |
| `↓` / `j` | Scroll activity log down |
| `PgUp` | Page up in activity log |
| `PgDn` | Page down in activity log |

---

## 8. Orchestration TUI Variant

The Orchestration TUI is a separate Bubble Tea application launched by `mind run --tui "<request>"`. It visualizes the full agent pipeline in real time.

### Layout

```
╭─ mind run ─── NEW_PROJECT: rest-api ──────────────────────────────────────╮
│                                                                             │
│  Pipeline Progress                                                         │
│  ─────────────────                                                         │
│  ✓ Pre-flight       classify, gate, iteration, branch           0.8s       │
│  ✓ Analyst          requirements (12 FR, 8 AC)                   2m 14s    │
│  ✓ Micro-Gate A     6/6 checks pass                             0.2s       │
│  ▸ Architect        designing... (src/routes/users.rs)           1m 32s    │
│  ○ Developer        waiting                                                │
│  ○ Micro-Gate B     waiting                                                │
│  ○ Tester           waiting                                                │
│  ○ Det. Gate        waiting                                                │
│  ○ Reviewer         waiting                                                │
│                                                                             │
│  ┌─ Agent Output (live) ─────────────────────────────────────────────────┐  │
│  │  Designing component architecture for REST API...                     │  │
│  │  Creating src/routes/users.rs with CRUD endpoints...                  │  │
│  │  Creating src/routes/auth.rs with JWT login/refresh...                │  │
│  │  Creating src/middleware/jwt.rs for token validation...               │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  Tokens: 42,318 in / 18,204 out    Cost: ~$1.82    Elapsed: 6m 47s        │
│  [p]ause  [s]kip  [a]bort  [d]etails  [l]og                               │
╰─────────────────────────────────────────────────────────────────────────── ╯
```

### Pipeline Progress Display

Each step in the pipeline is a row with four columns:

| Column | Width | Content |
|--------|-------|---------|
| Status icon | 2 chars | `✓` (complete/green), `▸` (active/yellow), `○` (waiting/dim), `✗` (failed/red), `↻` (retrying/yellow) |
| Step name | 16 chars | Pre-flight, Analyst, Micro-Gate A, Architect, Developer, Micro-Gate B, Tester, Det. Gate, Reviewer |
| Description | Dynamic | Summary of what happened (complete) or what is happening (active) or "waiting" (pending) |
| Duration | 10 chars, right-aligned | Elapsed time for completed steps, running clock for active step, empty for waiting |

Step descriptions update live:
- **Active step**: shows current activity parsed from agent output (e.g., "designing... (src/routes/users.rs)"). Updates every time the agent writes to a file.
- **Completed step**: shows summary (e.g., "requirements (12 FR, 8 AC)", "6/6 checks pass").
- **Failed step**: shows failure reason (e.g., "2/6 checks failed — missing success metrics").
- **Retrying step**: shows "retrying (1/2) — {failure reason}".

### Agent Output Panel

The lower section shows the live output from the currently active agent (captured from `claude --print` stdout). This is a scrollable viewport.

- Maximum 200 lines in the buffer per agent. Older lines are discarded.
- Auto-scroll follows the latest output.
- When the pipeline advances to the next agent, the output panel clears and shows the new agent's output.

### Resource Tracking Bar

Below the agent output panel, a single-line summary shows:

```
Tokens: {in_total} in / {out_total} out    Cost: ~${estimated}    Elapsed: {time}
```

- Token counts accumulate across all agent dispatches.
- Cost is estimated from token counts using approximate model pricing.
- Elapsed time is wall clock since `mind run` started.

### Orchestration TUI Keys

| Key | Action |
|-----|--------|
| `p` | Pause pipeline — completes the current agent, then waits for user to press `p` again to resume |
| `s` | Skip current agent — marks it as skipped, advances to next step |
| `a` | Abort pipeline — sends interrupt to active agent, cleans up, exits |
| `d` | Toggle details view — shows the full prompt that was sent to the current agent |
| `l` | Toggle log view — shows the full raw output from all agents (scrollable history) |
| `q` / `Ctrl+C` | Abort and quit (same as `a` followed by exit) |
| `↑` / `k` | Scroll agent output up (when auto-scroll is off) |
| `↓` / `j` | Scroll agent output down |
| `f` | Toggle follow mode for agent output |

### Details View (d)

When `d` is pressed, the agent output panel is replaced with the full assembled prompt for the current agent:

```
│  ┌─ Prompt for: Architect (opus) ────────────────────────────────────────┐  │
│  │  System: .mind/agents/architect.md (482 lines)                        │  │
│  │  Context:                                                             │  │
│  │    docs/spec/project-brief.md (96 lines)                              │  │
│  │    docs/spec/requirements.md (218 lines)                              │  │
│  │    docs/iterations/007-NEW_PROJECT-rest-api/overview.md (34 lines)    │  │
│  │  Conventions:                                                         │  │
│  │    .mind/conventions/shared.md (45 lines)                             │  │
│  │    .mind/conventions/documentation.md (62 lines)                      │  │
│  │  Task: "Design architecture for REST API with JWT auth"               │  │
│  │  Model: opus    Allowed tools: Read, Write, Edit, Grep, Glob         │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
```

Press `d` again to return to the live agent output.

### Log View (l)

When `l` is pressed, the pipeline progress shrinks to a single summary line and the full panel becomes a scrollable log of all agent output from the entire run:

```
│  Pipeline: Pre-flight ✓ → Analyst ✓ → MG-A ✓ → [Architect] → ...          │
│  ┌─ Full Log ────────────────────────────────────────────────────────────┐  │
│  │  [Analyst 14:20:15] Analyzing request: "create: REST API with JWT"   │  │
│  │  [Analyst 14:20:18] Reading project-brief.md...                      │  │
│  │  [Analyst 14:21:32] Writing docs/spec/requirements.md...             │  │
│  │  [Analyst 14:22:29] Done. 12 functional requirements, 8 criteria.    │  │
│  │  [Micro-Gate A 14:22:30] Running 6 checks...                        │  │
│  │  [Micro-Gate A 14:22:30] ✓ 6/6 pass                                 │  │
│  │  [Architect 14:22:31] Designing component architecture...            │  │
│  │  [Architect 14:23:15] Creating src/routes/users.rs...                │  │
│  │  ...                                                                  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
```

Press `l` again to return to the normal view.

### State Recovery

If `mind run` is interrupted (Ctrl+C, terminal close, machine crash), the pipeline state is persisted to `docs/state/workflow.md` after each agent completes. Running `mind run --resume` reads this state and continues from the last completed step.

The Orchestration TUI checks for existing state on startup:

```
│  ┌─ Resumable Workflow Found ────────────────────────────────────────────┐  │
│  │  Type: NEW_PROJECT (007-rest-api)                                     │  │
│  │  Completed: Pre-flight ✓ → Analyst ✓ → Micro-Gate A ✓                │  │
│  │  Next: Architect                                                      │  │
│  │                                                                       │  │
│  │  [r]esume from Architect    [s]tart over    [q]uit                    │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
```

---

## 9. Accessibility

### Dual-Channel Status Indicators

Every status indicator uses both color AND a distinct symbol. Users who cannot perceive color differences can distinguish states by symbol alone.

| State | Symbol | Color | Meaning |
|-------|--------|-------|---------|
| Pass / Complete | `✓` | Green | Check passed, step complete, document has content |
| Fail / Error | `✗` | Red | Check failed, step failed, document is stub |
| Warning | `⚠` | Yellow | Non-blocking issue, degraded state |
| In Progress | `▸` | Yellow | Currently active, running |
| Waiting | `○` | Dim | Pending, not yet started |
| Active (dot) | `●` | Green/Yellow | Active item in a list (stale=yellow, fresh=green) |
| Retrying | `↻` | Yellow | Agent being re-dispatched after gate failure |
| Pending (clock) | `⏳` | Dim | Awaiting trigger (used in gate status) |
| Info | `ℹ` | Blue | Informational message |

### Numeric Counts with Visual Bars

Progress bars always include a numeric fraction alongside the visual representation:

```
████████░░  4/5
```

Never rely on the bar alone to convey quantity. The `4/5` is always present and positioned consistently after the bar.

### Tab Labels with Shortcut Numbers

Tab labels always include their shortcut number:

```
[1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]
```

The numbers are part of the label text, not just keybindings. A user who cannot see the styling still knows which number maps to which tab.

### Screen Reader Considerations

While Bubble Tea is a visual TUI and does not directly support screen readers, the following design choices improve compatibility with terminal-based assistive tools:

- All text content is plain UTF-8 (no box-drawing characters in data content, only in borders).
- Status messages are written as readable sentences ("domain-model.md is a stub") rather than icon-only indicators.
- Warning and error messages include the word "Warning" or "Error" in the text, not just color/symbol.
- The `--json` output mode (available on all CLI commands) provides a fully accessible alternative to the TUI.

### Monochrome Terminal Support

The TUI functions correctly on monochrome terminals:

- Detect monochrome via the `NO_COLOR` environment variable or `TERM=dumb`.
- All symbols carry meaning without color (see table above).
- Bold and reverse video are used for emphasis instead of color when in monochrome mode.
- Progress bars switch to ASCII: `[###..]  3/5` instead of `███░░  3/5`.
- Active tab uses `[brackets]` instead of color highlighting.
- Selected row uses `> ` prefix instead of reverse video.

### Keyboard-Only Navigation

The entire TUI is operable via keyboard. There is no mouse dependency:

- Every action has a keyboard shortcut.
- Tab order is logical and predictable (left-to-right, top-to-bottom).
- Focus is always visually indicated (reverse video, brackets, or `>` prefix).
- The `?` help overlay documents all available keys for the current context.
- No hidden actions — every possible interaction is documented in the status bar or help overlay.

---

## Excluded Content

This blueprint intentionally does NOT contain:

- **Bubble Tea implementation code**: The component hierarchy defines what each component does, not how it is coded. Implementation details (goroutines, channels, tea.Cmd internals) belong in the code.
- **Data formats and serialization**: How `ProjectHealth`, `ValidationReport`, etc. are serialized is defined in the Data Contracts specification, not here.
- **Service layer logic**: How data is loaded (repository calls, file parsing) is defined in [03-architecture.md](03-architecture.md).
- **CLI command definitions**: Command flags and arguments are defined in [01-mind-cli.md](01-mind-cli.md).

---

> **See also:**
> - [01-mind-cli.md](01-mind-cli.md) -- CLI command tree, TUI overview, and project structure
> - [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md) -- Watch mode and orchestration integration models
> - [03-architecture.md](03-architecture.md) -- MVU pattern, domain types, component code patterns
