# Blueprint: CLI/TUI Software Architecture

> Complete architecture specification with design patterns, interface contracts, data flow, and implementation-ready Go code.

**Status**: Proposal
**Date**: 2026-03-09
**Depends on**: [01-mind-cli.md](01-mind-cli.md), [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md)

---

## Architecture Overview

### Layered Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Presentation Layer                                  │
│                                                                                  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│   │  CLI (Cobra)  │  │  TUI (Tea)   │  │  MCP Server  │  │  JSON Output     │   │
│   │  cmd/*.go     │  │  tui/*.go    │  │  mcp/*.go    │  │  (--json flag)   │   │
│   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────────┘   │
│          │                 │                  │                  │               │
└──────────┼─────────────────┼──────────────────┼──────────────────┼───────────────┘
           │                 │                  │                  │
           └─────────┬───────┴──────────┬───────┴──────────────────┘
                     │                  │
┌────────────────────┼──────────────────┼─────────────────────────────────────────┐
│                    ▼                  ▼        Service Layer                      │
│                                                                                  │
│   ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────────┐  │
│   │  ProjectService   │  │  WorkflowService  │  │  OrchestrationService       │  │
│   │  Status, Doctor   │  │  State, History   │  │  Preflight, Run, Handoff    │  │
│   └────────┬─────────┘  └────────┬─────────┘  └──────────────┬───────────────┘  │
│            │                     │                            │                   │
│   ┌────────┴─────────┐  ┌───────┴──────────┐  ┌─────────────┴───────────────┐  │
│   │  ValidationService│  │  GenerateService  │  │  SyncService               │  │
│   │  Docs, Refs, Gate │  │  Create, Template │  │  Agents, Platforms         │  │
│   └────────┬─────────┘  └────────┬─────────┘  └─────────────┬───────────────┘  │
│            │                     │                            │                   │
└────────────┼─────────────────────┼────────────────────────────┼──────────────────┘
             │                     │                            │
             └──────────┬──────────┴────────────────────────────┘
                        │
┌───────────────────────┼─────────────────────────────────────────────────────────┐
│                       ▼                Domain Layer                               │
│                                                                                  │
│   ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌────────────┐  │
│   │  Project   │  │  Document  │  │  Iteration │  │  Workflow  │  │  Quality   │  │
│   │  Config    │  │  Zone      │  │  Gate      │  │  State     │  │  Score     │  │
│   │  Brief     │  │  Stub      │  │  Chain     │  │  Chain     │  │  Rubric    │  │
│   └───────────┘  └───────────┘  └───────────┘  └───────────┘  └────────────┘  │
│                                                                                  │
└───────────────────────┬─────────────────────────────────────────────────────────┘
                        │
┌───────────────────────┼─────────────────────────────────────────────────────────┐
│                       ▼             Infrastructure Layer                          │
│                                                                                  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│   │  Filesystem   │  │  Git          │  │  Process      │  │  Template        │   │
│   │  (read/write) │  │  (branch,     │  │  (build, test │  │  (load, render)  │   │
│   │               │  │   status)     │  │   lint, exec) │  │                  │   │
│   └──────────────┘  └──────────────┘  └──────────────┘  └──────────────────┘   │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Layer Rules

| Rule | Description |
|------|-------------|
| **Dependencies flow down** | Presentation → Service → Domain → Infrastructure. Never upward. |
| **Domain has zero imports** | Domain types import only the standard library. No framework, no IO. |
| **Service orchestrates** | Services compose domain types and infrastructure. Business logic lives here. |
| **Presentation is thin** | CLI commands parse flags, call a service, render the result. No logic. |
| **Infrastructure is swappable** | Filesystem access goes through interfaces. Tests use in-memory implementations. |

---

## Domain Model

### Core Types

```go
// domain/project.go

// Project represents a Mind Framework project detected on disk.
type Project struct {
    Root       string     // Absolute path to project root (where .mind/ lives)
    Name       string     // From mind.toml [project].name
    Config     *Config    // Parsed mind.toml (nil if file doesn't exist)
    Framework  string     // Framework version from .mind/CHANGELOG.md
    DocsRoot   string     // Root + "/docs"
    MindRoot   string     // Root + "/.mind"
}

// Config represents the parsed mind.toml manifest.
type Config struct {
    Manifest   Manifest              `toml:"manifest"`
    Project    ProjectMeta           `toml:"project"`
    Profiles   Profiles              `toml:"profiles"`
    Documents  map[string]DocEntry   `toml:"documents"`
    Governance Governance            `toml:"governance"`
}

type Manifest struct {
    Schema     string `toml:"schema"`
    Generation int    `toml:"generation"`
    Updated    string `toml:"updated"`
}

type ProjectMeta struct {
    Name        string       `toml:"name"`
    Description string       `toml:"description"`
    Type        string       `toml:"type"`
    Stack       StackConfig  `toml:"stack"`
    Commands    CmdConfig    `toml:"commands"`
}

type StackConfig struct {
    Language  string `toml:"language"`
    Framework string `toml:"framework"`
    Testing   string `toml:"testing"`
}

type CmdConfig struct {
    Dev       string `toml:"dev"`
    Test      string `toml:"test"`
    Lint      string `toml:"lint"`
    Typecheck string `toml:"typecheck"`
    Build     string `toml:"build"`
}

type Governance struct {
    MaxRetries     int    `toml:"max-retries"`
    ReviewPolicy   string `toml:"review-policy"`
    CommitPolicy   string `toml:"commit-policy"`
    BranchStrategy string `toml:"branch-strategy"`
}
```

```go
// domain/document.go

// Zone represents one of the 5 documentation zones.
type Zone string

const (
    ZoneSpec       Zone = "spec"
    ZoneBlueprints Zone = "blueprints"
    ZoneState      Zone = "state"
    ZoneIterations Zone = "iterations"
    ZoneKnowledge  Zone = "knowledge"
)

// Document represents a single documentation file.
type Document struct {
    Path     string        // Relative to project root
    AbsPath  string        // Absolute path
    Zone     Zone
    Name     string        // Filename without extension
    Size     int64         // Bytes
    ModTime  time.Time
    IsStub   bool          // Detected by stub analysis
    Status   DocStatus     // From mind.toml or inferred
}

type DocStatus string

const (
    DocDraft    DocStatus = "draft"
    DocActive   DocStatus = "active"
    DocComplete DocStatus = "complete"
    DocStub     DocStatus = "stub"
)

// Brief represents a parsed project brief with section detection.
type Brief struct {
    Path           string
    Exists         bool
    IsStub         bool
    HasVision      bool
    HasDeliverables bool
    HasScope       bool
    GateResult     BriefGate
}

type BriefGate string

const (
    BriefPresent BriefGate = "BRIEF_PRESENT"
    BriefStub    BriefGate = "BRIEF_STUB"
    BriefMissing BriefGate = "BRIEF_MISSING"
)
```

```go
// domain/iteration.go

// RequestType classifies a user's workflow request.
type RequestType string

const (
    TypeNewProject  RequestType = "NEW_PROJECT"
    TypeBugFix      RequestType = "BUG_FIX"
    TypeEnhancement RequestType = "ENHANCEMENT"
    TypeRefactor    RequestType = "REFACTOR"
    TypeComplexNew  RequestType = "COMPLEX_NEW"
)

// Iteration represents a single workflow iteration.
type Iteration struct {
    Seq        int            // Sequence number (1, 2, 3...)
    Type       RequestType
    Descriptor string         // Kebab-case slug
    DirName    string         // Full directory name: "001-NEW_PROJECT-rest-api"
    Path       string         // Absolute path to iteration directory
    Artifacts  []Artifact     // Files in the iteration folder
    Status     IterationStatus
    CreatedAt  time.Time
}

type IterationStatus string

const (
    IterInProgress IterationStatus = "in_progress"
    IterComplete   IterationStatus = "complete"
    IterIncomplete IterationStatus = "incomplete"  // Missing artifacts
)

type Artifact struct {
    Name   string  // overview.md, changes.md, etc.
    Path   string
    Exists bool
}

// AgentChain defines the sequence of agents for a request type.
type AgentChain struct {
    Type   RequestType
    Agents []AgentRef
}

type AgentRef struct {
    Name     string  // analyst, architect, developer, tester, reviewer
    File     string  // .mind/agents/analyst.md
    Model    string  // opus, sonnet, haiku
    Optional bool    // true for architect in ENHANCEMENT
}

// Chains returns the agent chain for a given request type.
func Chains() map[RequestType]AgentChain {
    return map[RequestType]AgentChain{
        TypeNewProject: {TypeNewProject, []AgentRef{
            {"analyst", "agents/analyst.md", "opus", false},
            {"architect", "agents/architect.md", "opus", false},
            {"developer", "agents/developer.md", "sonnet", false},
            {"tester", "agents/tester.md", "sonnet", false},
            {"reviewer", "agents/reviewer.md", "opus", false},
        }},
        TypeBugFix: {TypeBugFix, []AgentRef{
            {"analyst", "agents/analyst.md", "opus", false},
            {"developer", "agents/developer.md", "sonnet", false},
            {"tester", "agents/tester.md", "sonnet", false},
            {"reviewer", "agents/reviewer.md", "opus", false},
        }},
        TypeEnhancement: {TypeEnhancement, []AgentRef{
            {"analyst", "agents/analyst.md", "opus", false},
            {"architect", "agents/architect.md", "opus", true},
            {"developer", "agents/developer.md", "sonnet", false},
            {"tester", "agents/tester.md", "sonnet", false},
            {"reviewer", "agents/reviewer.md", "opus", false},
        }},
        TypeRefactor: {TypeRefactor, []AgentRef{
            {"analyst", "agents/analyst.md", "opus", false},
            {"developer", "agents/developer.md", "sonnet", false},
            {"reviewer", "agents/reviewer.md", "opus", false},
        }},
        TypeComplexNew: {TypeComplexNew, []AgentRef{
            {"moderator", "conversation/agents/moderator.md", "opus", false},
            {"analyst", "agents/analyst.md", "opus", false},
            {"architect", "agents/architect.md", "opus", false},
            {"developer", "agents/developer.md", "sonnet", false},
            {"tester", "agents/tester.md", "sonnet", false},
            {"reviewer", "agents/reviewer.md", "opus", false},
        }},
    }
}
```

```go
// domain/workflow.go

// WorkflowState represents the persisted state of an in-progress workflow.
type WorkflowState struct {
    Type           RequestType
    Descriptor     string
    IterationPath  string
    Branch         string
    LastAgent      string
    RemainingChain []string
    Session        int
    TotalSessions  int
    Artifacts      []CompletedArtifact
    DispatchLog    []DispatchEntry
    Decisions      []string
    HandoffContext string
}

type CompletedArtifact struct {
    Agent    string
    Output   string
    Location string
}

type DispatchEntry struct {
    Agent     string
    File      string
    Model     string
    Status    string  // dispatched, completed, failed, retrying
    StartedAt time.Time
    Duration  time.Duration
}

// IsIdle returns true if no workflow is in progress.
func (s *WorkflowState) IsIdle() bool {
    return s == nil || s.Type == ""
}
```

```go
// domain/validation.go

// CheckResult represents the outcome of a single validation check.
type CheckResult struct {
    ID       int
    Name     string
    Level    CheckLevel  // Fail, Warn, Info
    Passed   bool
    Message  string      // Human-readable detail (empty if passed)
}

type CheckLevel string

const (
    LevelFail CheckLevel = "FAIL"
    LevelWarn CheckLevel = "WARN"
    LevelInfo CheckLevel = "INFO"
)

// ValidationReport aggregates results from a validation suite.
type ValidationReport struct {
    Suite    string         // "docs", "refs", "config", "convergence"
    Checks   []CheckResult
    Total    int
    Passed   int
    Failed   int
    Warnings int
}

func (r *ValidationReport) Ok() bool { return r.Failed == 0 }
```

```go
// domain/quality.go

// QualityScore represents a convergence analysis quality assessment.
type QualityScore struct {
    Overall    float64
    Dimensions map[string]int  // dimension name → score (1-5)
    Gate0Pass  bool            // overall >= 3.0
}

// QualityEntry represents a row in quality-log.yml.
type QualityEntry struct {
    Date        string             `yaml:"date"`
    Topic       string             `yaml:"topic"`
    SessionID   string             `yaml:"session_id"`
    Overall     float64            `yaml:"overall_score"`
    Dimensions  map[string]int     `yaml:"dimensions"`
    Gate0Pass   bool               `yaml:"gate_0_pass"`
    Personas    []string           `yaml:"personas_used"`
    Variant     string             `yaml:"variant"`
    OutputPath  string             `yaml:"output_path"`
}
```

```go
// domain/health.go

// ProjectHealth is the aggregate status shown by `mind status`.
type ProjectHealth struct {
    Project       Project
    Brief         Brief
    Zones         map[Zone]ZoneHealth
    Workflow      *WorkflowState
    LastIteration *Iteration
    Warnings      []string
    Suggestions   []string
}

// ZoneHealth tracks completeness of a single documentation zone.
type ZoneHealth struct {
    Zone     Zone
    Total    int       // Total expected documents
    Present  int       // Documents that exist
    Stubs    int       // Documents that are stubs
    Complete int       // Documents with real content
    Files    []Document
}
```

### Type Relationships

```
Project ──────┬── Config (mind.toml)
              ├── Brief (project-brief.md)
              ├── Document[] ─── Zone
              ├── Iteration[] ─── Artifact[]
              │                    └── RequestType
              ├── WorkflowState ─── DispatchEntry[]
              │                      └── CompletedArtifact[]
              └── ProjectHealth ─── ZoneHealth[]
                                     └── QualityScore

ValidationReport ─── CheckResult[]

AgentChain ─── AgentRef[]
```

---

## Design Patterns

### 1. Elm Architecture (Model-View-Update) — TUI

The TUI uses Bubble Tea's MVU pattern. The entire UI state is a single immutable model. User input produces messages. An update function transforms the model. A view function renders it.

```go
// tui/app.go

type Tab int

const (
    TabStatus Tab = iota
    TabDocs
    TabIterations
    TabChecks
    TabQuality
)

// Model is the top-level TUI state.
type Model struct {
    // Navigation
    activeTab   Tab
    width       int
    height      int

    // Tab models (each tab is its own Bubble Tea model)
    status      StatusModel
    docs        DocsModel
    iterations  IterationsModel
    checks      ChecksModel
    quality     QualityModel

    // Shared state
    project     domain.Project
    health      *domain.ProjectHealth
    err         error

    // Services (injected)
    projectSvc  service.ProjectService
    validateSvc service.ValidationService
}

// Messages
type tabSwitchMsg Tab
type healthLoadedMsg struct{ health *domain.ProjectHealth }
type healthErrorMsg struct{ err error }
type validationCompleteMsg struct{ report domain.ValidationReport }
type windowSizeMsg tea.WindowSizeMsg

// Init loads project health on startup.
func (m Model) Init() tea.Cmd {
    return m.loadHealth
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "1":
            m.activeTab = TabStatus
        case "2":
            m.activeTab = TabDocs
        case "3":
            m.activeTab = TabIterations
        case "4":
            m.activeTab = TabChecks
            return m, m.runValidation
        case "5":
            m.activeTab = TabQuality
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

    case healthLoadedMsg:
        m.health = msg.health
        m.status = m.status.SetHealth(msg.health)
        m.docs = m.docs.SetDocuments(msg.health)
        m.iterations = m.iterations.SetIterations(msg.health)

    case healthErrorMsg:
        m.err = msg.err

    case validationCompleteMsg:
        m.checks = m.checks.SetReport(msg.report)
    }

    // Delegate to active tab
    switch m.activeTab {
    case TabStatus:
        m.status, cmd = m.status.Update(msg)
    case TabDocs:
        m.docs, cmd = m.docs.Update(msg)
    // ...
    }

    return m, cmd
}

// View renders the active tab with the tab bar.
func (m Model) View() string {
    tabs := m.renderTabBar()
    var content string
    switch m.activeTab {
    case TabStatus:
        content = m.status.View()
    case TabDocs:
        content = m.docs.View()
    case TabIterations:
        content = m.iterations.View()
    case TabChecks:
        content = m.checks.View()
    case TabQuality:
        content = m.quality.View()
    }
    return lipgloss.JoinVertical(lipgloss.Left, tabs, content)
}
```

Each tab is its own Bubble Tea model with its own `Update` and `View`. The top-level model delegates.

```go
// tui/status.go

type StatusModel struct {
    health    *domain.ProjectHealth
    viewport  viewport.Model  // Scrollable content
}

func (m StatusModel) View() string {
    if m.health == nil {
        return "Loading..."
    }

    var b strings.Builder

    // Documentation health bars
    for _, zone := range domain.AllZones {
        zh := m.health.Zones[zone]
        bar := renderProgressBar(zh.Complete, zh.Total, 20)
        fmt.Fprintf(&b, "  %-14s %s  %d/%d\n", zone+"/", bar, zh.Complete, zh.Total)
    }

    // Workflow state
    b.WriteString("\n")
    if m.health.Workflow.IsIdle() {
        b.WriteString("  Workflow: idle\n")
    } else {
        fmt.Fprintf(&b, "  Workflow: %s (%s)\n",
            m.health.Workflow.Type,
            m.health.Workflow.Descriptor)
        fmt.Fprintf(&b, "  Agent: %s → next: %s\n",
            m.health.Workflow.LastAgent,
            m.health.Workflow.RemainingChain[0])
    }

    // Warnings
    if len(m.health.Warnings) > 0 {
        b.WriteString("\n  Warnings\n")
        for _, w := range m.health.Warnings {
            fmt.Fprintf(&b, "  ⚠ %s\n", w)
        }
    }

    return b.String()
}
```

### 2. Strategy Pattern — Output Rendering

Every command produces a domain result. The presentation layer selects a renderer based on the output mode (interactive, plain, JSON).

```go
// internal/render/render.go

// Renderer formats a result for display.
type Renderer interface {
    RenderHealth(h *domain.ProjectHealth) string
    RenderValidation(r *domain.ValidationReport) string
    RenderIterations(iters []domain.Iteration) string
    RenderDoctor(diags []domain.Diagnostic) string
    // ... one method per result type
}

// InteractiveRenderer uses Lip Gloss styling.
type InteractiveRenderer struct {
    width int
    styles Styles
}

// PlainRenderer outputs clean text with no ANSI codes.
type PlainRenderer struct{}

// JSONRenderer outputs structured JSON.
type JSONRenderer struct {
    pretty bool
}

// NewRenderer selects the appropriate renderer.
func NewRenderer(mode OutputMode, width int) Renderer {
    switch mode {
    case ModeInteractive:
        return &InteractiveRenderer{width: width, styles: DefaultStyles()}
    case ModePlain:
        return &PlainRenderer{}
    case ModeJSON:
        return &JSONRenderer{pretty: true}
    default:
        return &PlainRenderer{}
    }
}

// OutputMode is determined at startup.
type OutputMode int

const (
    ModeInteractive OutputMode = iota
    ModePlain
    ModeJSON
)

// DetectMode checks TTY and flags.
func DetectMode(jsonFlag bool, noColorFlag bool) OutputMode {
    if jsonFlag {
        return ModeJSON
    }
    if noColorFlag || !isatty.IsTerminal(os.Stdout.Fd()) {
        return ModePlain
    }
    return ModeInteractive
}
```

Usage in a command:

```go
// cmd/status.go

func runStatus(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    svc := ctx.Value(ctxProjectSvc).(service.ProjectService)
    renderer := ctx.Value(ctxRenderer).(render.Renderer)

    health, err := svc.Health(ctx)
    if err != nil {
        return err
    }

    fmt.Print(renderer.RenderHealth(health))
    return nil
}
```

### 3. Repository Pattern — Data Access

All file system access goes through repository interfaces. This makes the domain and service layers testable with in-memory fakes.

```go
// internal/repo/interfaces.go

// DocRepo reads and queries the 5-zone documentation structure.
type DocRepo interface {
    // ListByZone returns all documents in a zone.
    ListByZone(zone domain.Zone) ([]domain.Document, error)

    // ListAll returns every document across all zones.
    ListAll() ([]domain.Document, error)

    // Read returns the content of a document.
    Read(relPath string) ([]byte, error)

    // Exists checks if a file exists.
    Exists(relPath string) bool

    // IsStub detects if a document is a stub (template-only content).
    IsStub(relPath string) (bool, error)

    // Search performs full-text search across all documents.
    Search(query string) ([]SearchResult, error)
}

// IterationRepo manages iteration folders.
type IterationRepo interface {
    // List returns all iterations, newest first.
    List() ([]domain.Iteration, error)

    // Get returns a single iteration by sequence number.
    Get(seq int) (*domain.Iteration, error)

    // NextSeq returns the next available sequence number.
    NextSeq() (int, error)

    // Create creates an iteration folder from templates.
    Create(reqType domain.RequestType, descriptor string) (*domain.Iteration, error)
}

// StateRepo reads and writes workflow state.
type StateRepo interface {
    // ReadWorkflow parses docs/state/workflow.md.
    ReadWorkflow() (*domain.WorkflowState, error)

    // WriteWorkflow serializes state to docs/state/workflow.md.
    WriteWorkflow(state *domain.WorkflowState) error

    // ReadCurrent parses docs/state/current.md.
    ReadCurrent() (*domain.CurrentState, error)

    // WriteCurrent updates docs/state/current.md.
    WriteCurrent(state *domain.CurrentState) error
}

// ConfigRepo reads project and framework configuration.
type ConfigRepo interface {
    // ReadProjectConfig parses mind.toml.
    ReadProjectConfig() (*domain.Config, error)

    // ReadConversationConfig parses conversation/config/*.yml.
    ReadConversationConfig() (*domain.ConversationConfig, error)

    // ReadQualityConfig parses conversation/config/quality.yml.
    ReadQualityConfig() (*domain.QualityConfig, error)
}

// TemplateRepo loads and renders document templates.
type TemplateRepo interface {
    // Load reads a template file from .mind/docs/templates/.
    Load(name string) ([]byte, error)

    // Render applies substitution markers to a template.
    Render(template []byte, vars TemplateVars) []byte
}

type TemplateVars struct {
    Title string
    Date  string
    Seq   string
    Type  string
}

// QualityRepo reads and writes quality tracking data.
type QualityRepo interface {
    // ReadLog parses docs/knowledge/quality-log.yml.
    ReadLog() ([]domain.QualityEntry, error)

    // AppendLog adds an entry to quality-log.yml.
    AppendLog(entry domain.QualityEntry) error

    // ExtractScores parses quality scores from a convergence file.
    ExtractScores(convergencePath string) (*domain.QualityScore, error)
}
```

Filesystem implementation:

```go
// internal/repo/fs/doc_repo.go

type FSDocRepo struct {
    projectRoot string
    docsRoot    string  // projectRoot + "/docs"
}

func NewDocRepo(projectRoot string) *FSDocRepo {
    return &FSDocRepo{
        projectRoot: projectRoot,
        docsRoot:    filepath.Join(projectRoot, "docs"),
    }
}

func (r *FSDocRepo) IsStub(relPath string) (bool, error) {
    absPath := filepath.Join(r.projectRoot, relPath)

    info, err := os.Stat(absPath)
    if err != nil {
        return false, err
    }
    if info.Size() == 0 {
        return true, nil
    }

    content, err := os.ReadFile(absPath)
    if err != nil {
        return false, err
    }

    realLines := 0
    scanner := bufio.NewScanner(bytes.NewReader(content))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" { continue }                              // blank
        if strings.HasPrefix(line, "#") { continue }            // heading
        if strings.HasPrefix(line, "<!--") { continue }         // HTML comment open
        if strings.HasPrefix(line, "-->") { continue }          // HTML comment close
        if strings.HasPrefix(line, ">") { continue }            // blockquote
        if isTableSeparator(line) { continue }                  // |---|---|
        if isPlaceholderRow(line) { continue }                  // | <!-- --> |
        realLines++
    }

    return realLines <= 2, nil
}

func isTableSeparator(line string) bool {
    if !strings.HasPrefix(line, "|") { return false }
    cleaned := strings.ReplaceAll(line, "|", "")
    cleaned = strings.ReplaceAll(cleaned, "-", "")
    cleaned = strings.ReplaceAll(cleaned, ":", "")
    cleaned = strings.TrimSpace(cleaned)
    return cleaned == ""
}

func isPlaceholderRow(line string) bool {
    return strings.HasPrefix(line, "|") &&
        strings.Contains(line, "<!--") &&
        strings.Contains(line, "-->")
}
```

In-memory implementation for tests:

```go
// internal/repo/mem/doc_repo.go

type MemDocRepo struct {
    files map[string][]byte  // relPath → content
}

func NewMemDocRepo() *MemDocRepo {
    return &MemDocRepo{files: make(map[string][]byte)}
}

func (r *MemDocRepo) AddFile(relPath string, content string) {
    r.files[relPath] = []byte(content)
}

func (r *MemDocRepo) Exists(relPath string) bool {
    _, ok := r.files[relPath]
    return ok
}

func (r *MemDocRepo) IsStub(relPath string) (bool, error) {
    content, ok := r.files[relPath]
    if !ok {
        return false, os.ErrNotExist
    }
    // Same logic as FSDocRepo, operating on content bytes
    return stubDetect(content), nil
}
```

### 4. Chain of Responsibility — Validation Pipeline

Validators are composable checks. Each check is independent. The pipeline collects all results.

```go
// internal/validate/check.go

// Check is a single validation check function.
type Check struct {
    ID    int
    Name  string
    Level domain.CheckLevel
    Fn    CheckFunc
}

// CheckFunc receives the project context and returns pass/fail + message.
type CheckFunc func(ctx *CheckContext) (bool, string)

// CheckContext provides everything a check might need.
type CheckContext struct {
    Project  domain.Project
    DocRepo  repo.DocRepo
    IterRepo repo.IterationRepo
    Strict   bool
}

// Suite is an ordered list of checks.
type Suite struct {
    Name   string
    Checks []Check
}

// Run executes all checks and returns a report.
func (s *Suite) Run(ctx *CheckContext) domain.ValidationReport {
    report := domain.ValidationReport{
        Suite: s.Name,
        Total: len(s.Checks),
    }

    for _, check := range s.Checks {
        passed, msg := check.Fn(ctx)
        result := domain.CheckResult{
            ID:      check.ID,
            Name:    check.Name,
            Level:   check.Level,
            Passed:  passed,
            Message: msg,
        }
        report.Checks = append(report.Checks, result)

        if passed {
            report.Passed++
        } else if check.Level == domain.LevelFail {
            report.Failed++
        } else {
            report.Warnings++
        }
    }

    return report
}
```

Registering checks:

```go
// internal/validate/docs.go

func DocsSuite() *Suite {
    return &Suite{
        Name: "docs",
        Checks: []Check{
            {1, "docs/ directory exists", domain.LevelFail, checkDocsDir},
            {2, "All 5 zone directories exist", domain.LevelFail, checkZoneDirs},
            {3, "Required spec files", domain.LevelFail, checkSpecFiles},
            {4, "decisions/ subdirectory", domain.LevelWarn, checkDecisionsDir},
            {5, "ADR naming convention", domain.LevelWarn, checkADRNaming},
            {6, "blueprints/INDEX.md", domain.LevelFail, checkBlueprintsIndex},
            {7, "Blueprint → INDEX.md coverage", domain.LevelWarn, checkBlueprintCoverage},
            {8, "INDEX.md → file references", domain.LevelFail, checkIndexRefs},
            {9, "state/current.md", domain.LevelFail, checkCurrentState},
            {10, "state/workflow.md", domain.LevelWarn, checkWorkflowState},
            {11, "knowledge/glossary.md", domain.LevelWarn, checkGlossary},
            {12, "Iteration folder naming", domain.LevelWarn, checkIterationNaming},
            {13, "Iterations have overview.md", domain.LevelWarn, checkIterationOverview},
            {14, "Spike file naming", domain.LevelWarn, checkSpikeNaming},
            {15, "No legacy paths", domain.LevelFail, checkNoLegacyPaths},
            {16, "Stub detection", domain.LevelWarn, checkStubs},  // LevelFail in strict
            {17, "Project brief completeness", domain.LevelWarn, checkBriefCompleteness},
        },
    }
}

func checkDocsDir(ctx *CheckContext) (bool, string) {
    if ctx.DocRepo.Exists("docs") {
        return true, ""
    }
    return false, "docs/ directory not found"
}

func checkSpecFiles(ctx *CheckContext) (bool, string) {
    required := []string{
        "docs/spec/project-brief.md",
        "docs/spec/requirements.md",
        "docs/spec/architecture.md",
    }
    var missing []string
    for _, f := range required {
        if !ctx.DocRepo.Exists(f) {
            missing = append(missing, filepath.Base(f))
        }
    }
    if len(missing) > 0 {
        return false, "Missing: " + strings.Join(missing, ", ")
    }
    return true, ""
}

func checkBriefCompleteness(ctx *CheckContext) (bool, string) {
    briefPath := "docs/spec/project-brief.md"
    if !ctx.DocRepo.Exists(briefPath) {
        return true, "" // Caught by check 3
    }
    isStub, _ := ctx.DocRepo.IsStub(briefPath)
    if isStub {
        return true, "" // Caught by check 16
    }

    content, err := ctx.DocRepo.Read(briefPath)
    if err != nil {
        return false, err.Error()
    }

    sections := []string{"Vision", "Key Deliverables", "Scope"}
    var missing []string
    for _, s := range sections {
        pattern := regexp.MustCompile(`(?i)^##\s+.*` + regexp.QuoteMeta(s))
        if !pattern.Match(content) {
            missing = append(missing, s)
        }
    }

    if len(missing) > 0 {
        return false, "Missing sections: " + strings.Join(missing, ", ")
    }
    return true, ""
}
```

### 5. Builder Pattern — Context Assembly

For Model D (full orchestration), agent prompts are assembled from multiple sources. A builder ensures all required pieces are included.

```go
// internal/orchestrate/prompt.go

// PromptBuilder assembles an agent prompt from structured components.
type PromptBuilder struct {
    sections []promptSection
}

type promptSection struct {
    header  string
    content string
}

func NewPromptBuilder() *PromptBuilder {
    return &PromptBuilder{}
}

func (b *PromptBuilder) AgentInstructions(agentFile string, content []byte) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "AGENT INSTRUCTIONS",
        content: string(content),
    })
    return b
}

func (b *PromptBuilder) ProjectContext(brief, requirements, architecture []byte) *PromptBuilder {
    var parts []string
    if brief != nil {
        parts = append(parts, "### Project Brief\n"+string(brief))
    }
    if requirements != nil {
        parts = append(parts, "### Requirements\n"+string(requirements))
    }
    if architecture != nil {
        parts = append(parts, "### Architecture\n"+string(architecture))
    }
    if len(parts) > 0 {
        b.sections = append(b.sections, promptSection{
            header:  "PROJECT CONTEXT",
            content: strings.Join(parts, "\n\n"),
        })
    }
    return b
}

func (b *PromptBuilder) IterationContext(overview []byte) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "ITERATION CONTEXT",
        content: string(overview),
    })
    return b
}

func (b *PromptBuilder) PriorAgentOutput(agentName string, output []byte) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "OUTPUT FROM " + strings.ToUpper(agentName),
        content: string(output),
    })
    return b
}

func (b *PromptBuilder) ConvergenceContext(convergence []byte) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "CONVERGENCE ANALYSIS (COMPLEX_NEW)",
        content: string(convergence),
    })
    return b
}

func (b *PromptBuilder) Conventions(shared, documentation []byte) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "CONVENTIONS",
        content: string(shared) + "\n\n" + string(documentation),
    })
    return b
}

func (b *PromptBuilder) GateFailure(gateName string, results domain.ValidationReport) *PromptBuilder {
    var failures []string
    for _, c := range results.Checks {
        if !c.Passed {
            failures = append(failures, fmt.Sprintf("- %s: %s", c.Name, c.Message))
        }
    }
    b.sections = append(b.sections, promptSection{
        header:  "RETRY — " + gateName + " FAILED",
        content: "Fix these issues:\n" + strings.Join(failures, "\n"),
    })
    return b
}

func (b *PromptBuilder) Task(instruction string) *PromptBuilder {
    b.sections = append(b.sections, promptSection{
        header:  "TASK",
        content: instruction,
    })
    return b
}

// Build produces the final prompt string.
func (b *PromptBuilder) Build() string {
    var sb strings.Builder
    for _, s := range b.sections {
        sb.WriteString("## ")
        sb.WriteString(s.header)
        sb.WriteString("\n\n")
        sb.WriteString(s.content)
        sb.WriteString("\n\n---\n\n")
    }
    return sb.String()
}
```

### 6. Observer Pattern — Watch Mode

Filesystem events flow through a channel. Handlers react to specific file patterns.

```go
// internal/watch/watcher.go

// Event represents a filesystem change relevant to the framework.
type Event struct {
    Path      string
    Op        EventOp
    Timestamp time.Time
}

type EventOp int

const (
    OpCreated EventOp = iota
    OpModified
    OpDeleted
)

// Handler reacts to a specific file pattern.
type Handler struct {
    Pattern  glob.Glob      // e.g., "docs/iterations/*/changes.md"
    Name     string
    OnEvent  func(Event) Action
}

// Action is what the watcher should do in response.
type Action struct {
    Log     string           // Message to display in the activity log
    RunGate string           // Gate to run ("micro-gate-b", "deterministic", "")
    RunCmd  string           // Shell command to run in background ("cargo build", "")
}

// Watcher monitors the project filesystem and dispatches to handlers.
type Watcher struct {
    root     string
    handlers []Handler
    events   chan Event
    actions  chan Action
    fsw      *fsnotify.Watcher
}

func New(root string, handlers []Handler) (*Watcher, error) {
    fsw, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    return &Watcher{
        root:     root,
        handlers: handlers,
        events:   make(chan Event, 64),
        actions:  make(chan Action, 64),
        fsw:      fsw,
    }, nil
}

// Start begins watching. Returns the actions channel for the TUI to consume.
func (w *Watcher) Start(ctx context.Context) <-chan Action {
    // Add recursive watches
    filepath.WalkDir(w.root, func(path string, d fs.DirEntry, err error) error {
        if d != nil && d.IsDir() {
            w.fsw.Add(path)
        }
        return nil
    })

    // Debounce + dispatch loop
    go func() {
        debounce := make(map[string]time.Time)
        ticker := time.NewTicker(200 * time.Millisecond)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case ev := <-w.fsw.Events:
                relPath, _ := filepath.Rel(w.root, ev.Name)
                debounce[relPath] = time.Now()
            case <-ticker.C:
                now := time.Now()
                for path, t := range debounce {
                    if now.Sub(t) > 300*time.Millisecond {
                        delete(debounce, path)
                        w.dispatch(Event{Path: path, Op: OpModified, Timestamp: now})
                    }
                }
            }
        }
    }()

    return w.actions
}

func (w *Watcher) dispatch(event Event) {
    for _, h := range w.handlers {
        if h.Pattern.Match(event.Path) {
            action := h.OnEvent(event)
            if action.Log != "" {
                w.actions <- action
            }
        }
    }
}
```

Default handlers:

```go
// internal/watch/handlers.go

func DefaultHandlers(validateSvc service.ValidationService) []Handler {
    return []Handler{
        {
            Pattern: glob.MustCompile("docs/state/workflow.md"),
            Name:    "workflow-state",
            OnEvent: func(e Event) Action {
                return Action{Log: "Workflow state updated"}
            },
        },
        {
            Pattern: glob.MustCompile("docs/iterations/*/changes.md"),
            Name:    "developer-changes",
            OnEvent: func(e Event) Action {
                return Action{
                    Log:     "Developer changes detected — running Micro-Gate B",
                    RunGate: "micro-gate-b",
                }
            },
        },
        {
            Pattern: glob.MustCompile("docs/spec/requirements.md"),
            Name:    "analyst-requirements",
            OnEvent: func(e Event) Action {
                return Action{
                    Log:     "Requirements updated — running Micro-Gate A",
                    RunGate: "micro-gate-a",
                }
            },
        },
        {
            Pattern: glob.MustCompile("docs/knowledge/*-convergence.md"),
            Name:    "convergence-output",
            OnEvent: func(e Event) Action {
                return Action{
                    Log:     "Convergence analysis detected — validating",
                    RunGate: "convergence",
                }
            },
        },
        {
            Pattern: glob.MustCompile("src/**/*"),
            Name:    "source-change",
            OnEvent: func(e Event) Action {
                return Action{
                    Log:    fmt.Sprintf("Source changed: %s", e.Path),
                    RunCmd: "cargo build",  // Detected from mind.toml
                }
            },
        },
    }
}
```

### 7. Pipeline Pattern — Orchestration Engine

Model D's orchestration is a pipeline of steps. Each step is a function with a uniform interface.

```go
// internal/orchestrate/pipeline.go

// Step represents one stage in the orchestration pipeline.
type Step interface {
    Name() string
    Execute(ctx context.Context, state *PipelineState) error
}

// PipelineState carries data between steps.
type PipelineState struct {
    Request        string
    RequestType    domain.RequestType
    Iteration      *domain.Iteration
    Branch         string
    BriefGate      domain.BriefGate
    AgentChain     domain.AgentChain
    CurrentAgent   int
    CompletedSteps []string
    Retries        map[string]int  // agent name → retry count
    Artifacts      map[string][]byte
    DryRun         bool
    Aborted        bool

    // Events channel for TUI updates
    Events chan<- PipelineEvent
}

type PipelineEvent struct {
    Step      string
    Status    string  // "started", "completed", "failed", "retrying"
    Message   string
    Timestamp time.Time
}

func (s *PipelineState) Emit(step, status, msg string) {
    if s.Events != nil {
        s.Events <- PipelineEvent{
            Step:      step,
            Status:    status,
            Message:   msg,
            Timestamp: time.Now(),
        }
    }
}

// Pipeline executes steps in order with retry logic.
type Pipeline struct {
    steps []Step
}

func NewPipeline(steps ...Step) *Pipeline {
    return &Pipeline{steps: steps}
}

func (p *Pipeline) Run(ctx context.Context, state *PipelineState) error {
    for _, step := range p.steps {
        if state.Aborted {
            return fmt.Errorf("pipeline aborted at step: %s", step.Name())
        }
        state.Emit(step.Name(), "started", "")

        if err := step.Execute(ctx, state); err != nil {
            state.Emit(step.Name(), "failed", err.Error())
            return fmt.Errorf("step %s failed: %w", step.Name(), err)
        }

        state.CompletedSteps = append(state.CompletedSteps, step.Name())
        state.Emit(step.Name(), "completed", "")
    }
    return nil
}

// Concrete steps

type ClassifyStep struct{}
func (s *ClassifyStep) Name() string { return "classify" }
func (s *ClassifyStep) Execute(ctx context.Context, state *PipelineState) error {
    state.RequestType = classify(state.Request)
    state.AgentChain = domain.Chains()[state.RequestType]
    return nil
}

type BriefGateStep struct {
    docRepo repo.DocRepo
}
func (s *BriefGateStep) Name() string { return "brief-gate" }
func (s *BriefGateStep) Execute(ctx context.Context, state *PipelineState) error {
    brief := parseBrief(s.docRepo)
    state.BriefGate = brief.GateResult

    switch state.RequestType {
    case domain.TypeBugFix, domain.TypeRefactor:
        return nil // Skip gate
    case domain.TypeNewProject, domain.TypeComplexNew:
        if brief.GateResult != domain.BriefPresent {
            state.Emit("brief-gate", "blocked",
                "No project brief. Run /discover or fill docs/spec/project-brief.md")
            return fmt.Errorf("business context gate: %s", brief.GateResult)
        }
    case domain.TypeEnhancement:
        if brief.GateResult != domain.BriefPresent {
            state.Emit("brief-gate", "warning", "No project brief found")
        }
    }
    return nil
}

type CreateIterationStep struct {
    iterRepo repo.IterationRepo
}
func (s *CreateIterationStep) Name() string { return "create-iteration" }
func (s *CreateIterationStep) Execute(ctx context.Context, state *PipelineState) error {
    slug := domain.Slugify(state.Request)
    iter, err := s.iterRepo.Create(state.RequestType, slug)
    if err != nil {
        return err
    }
    state.Iteration = iter
    state.Emit("create-iteration", "completed",
        fmt.Sprintf("Created %s", iter.DirName))
    return nil
}

type DispatchAgentStep struct {
    agent    domain.AgentRef
    executor AgentExecutor
}
func (s *DispatchAgentStep) Name() string { return "agent:" + s.agent.Name }
func (s *DispatchAgentStep) Execute(ctx context.Context, state *PipelineState) error {
    if state.DryRun {
        state.Emit(s.Name(), "dry-run",
            fmt.Sprintf("Would dispatch %s (model: %s)", s.agent.Name, s.agent.Model))
        return nil
    }

    prompt := buildPromptForAgent(s.agent, state)
    output, err := s.executor.Run(ctx, s.agent.Model, prompt)
    if err != nil {
        return err
    }
    state.Artifacts[s.agent.Name] = output
    return nil
}

type QualityGateStep struct {
    name      string
    validator func(*PipelineState) *domain.ValidationReport
    maxRetry  int
    retryStep Step
}
func (s *QualityGateStep) Name() string { return "gate:" + s.name }
func (s *QualityGateStep) Execute(ctx context.Context, state *PipelineState) error {
    report := s.validator(state)
    if report.Ok() {
        return nil
    }

    retries := state.Retries[s.name]
    if retries >= s.maxRetry {
        state.Emit(s.Name(), "warning",
            fmt.Sprintf("Gate failed after %d retries, proceeding with concerns", s.maxRetry))
        return nil // Proceed with documented concerns
    }

    state.Retries[s.name] = retries + 1
    state.Emit(s.Name(), "retrying",
        fmt.Sprintf("Retry %d/%d", retries+1, s.maxRetry))
    return s.retryStep.Execute(ctx, state)
}

type DeterministicGateStep struct {
    commands CmdConfig
}
func (s *DeterministicGateStep) Name() string { return "gate:deterministic" }
func (s *DeterministicGateStep) Execute(ctx context.Context, state *PipelineState) error {
    if state.DryRun {
        state.Emit(s.Name(), "dry-run", "Would run: build, lint, test")
        return nil
    }
    cmds := []struct{ name, cmd string }{
        {"build", s.commands.Build},
        {"lint", s.commands.Lint},
        {"test", s.commands.Test},
    }
    for _, c := range cmds {
        if c.cmd == "" { continue }
        state.Emit(s.Name(), "running", c.name)
        if err := exec.CommandContext(ctx, "sh", "-c", c.cmd).Run(); err != nil {
            return fmt.Errorf("%s failed: %w", c.name, err)
        }
    }
    return nil
}
```

Building the full pipeline:

```go
// internal/orchestrate/build.go

func BuildPipeline(
    reqType domain.RequestType,
    docRepo repo.DocRepo,
    iterRepo repo.IterationRepo,
    executor AgentExecutor,
    commands CmdConfig,
) *Pipeline {
    chain := domain.Chains()[reqType]
    steps := []Step{
        &ClassifyStep{},
        &BriefGateStep{docRepo: docRepo},
        &CreateIterationStep{iterRepo: iterRepo},
    }

    for i, agent := range chain.Agents {
        // Dispatch agent
        steps = append(steps, &DispatchAgentStep{
            agent:    agent,
            executor: executor,
        })

        // Quality gate after specific agents
        switch agent.Name {
        case "analyst":
            steps = append(steps, &QualityGateStep{
                name:      "micro-gate-a",
                validator: validateMicroGateA,
                maxRetry:  1,
                retryStep: &DispatchAgentStep{agent: agent, executor: executor},
            })
        case "developer":
            steps = append(steps, &QualityGateStep{
                name:      "micro-gate-b",
                validator: validateMicroGateB,
                maxRetry:  1,
                retryStep: &DispatchAgentStep{agent: agent, executor: executor},
            })
        case "tester":
            // Deterministic gate runs after tester, before reviewer
            if i < len(chain.Agents)-1 && chain.Agents[i+1].Name == "reviewer" {
                steps = append(steps, &DeterministicGateStep{commands: commands})
            }
        }
    }

    return NewPipeline(steps...)
}
```

### 8. Adapter Pattern — MCP Server

The MCP server adapts internal service calls to the MCP JSON-RPC protocol.

```go
// internal/mcp/server.go

// Tool wraps a service call as an MCP tool.
type Tool struct {
    Name        string
    Description string
    Schema      json.RawMessage  // JSON Schema for input
    Handler     func(ctx context.Context, input json.RawMessage) (any, error)
}

// Server implements the MCP protocol over stdio.
type Server struct {
    tools   map[string]Tool
    scanner *bufio.Scanner
    encoder *json.Encoder
}

func NewServer(tools []Tool) *Server {
    toolMap := make(map[string]Tool)
    for _, t := range tools {
        toolMap[t.Name] = t
    }
    return &Server{
        tools:   toolMap,
        scanner: bufio.NewScanner(os.Stdin),
        encoder: json.NewEncoder(os.Stdout),
    }
}

// RegisterTools wires services into MCP tools.
func RegisterTools(
    projectSvc service.ProjectService,
    validateSvc service.ValidationService,
    iterSvc service.IterationService,
    stateSvc service.WorkflowService,
    generateSvc service.GenerateService,
    qualitySvc service.QualityService,
) []Tool {
    return []Tool{
        {
            Name:        "mind_status",
            Description: "Project health summary: documentation completeness, workflow state, warnings",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                return projectSvc.Health(ctx)
            },
        },
        {
            Name:        "mind_check_brief",
            Description: "Check business context gate: is project-brief.md present and non-stub?",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                return projectSvc.CheckBrief(ctx)
            },
        },
        {
            Name:        "mind_validate_docs",
            Description: "Run 17-check documentation structure validation",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                var opts struct{ Strict bool `json:"strict"` }
                json.Unmarshal(input, &opts)
                return validateSvc.ValidateDocs(ctx, opts.Strict)
            },
        },
        {
            Name:        "mind_create_iteration",
            Description: "Create a new iteration folder with template files",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                var req struct {
                    Type       string `json:"type"`
                    Descriptor string `json:"descriptor"`
                }
                json.Unmarshal(input, &req)
                return iterSvc.Create(ctx, domain.RequestType(req.Type), req.Descriptor)
            },
        },
        {
            Name:        "mind_check_gate",
            Description: "Run deterministic gate: execute build, lint, and test commands",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                return validateSvc.RunDeterministicGate(ctx)
            },
        },
        {
            Name:        "mind_suggest_next",
            Description: "Suggest the next action based on current project state",
            Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
                return projectSvc.SuggestNext(ctx)
            },
        },
        // ... remaining tools
    }
}
```

---

## Service Layer

Services orchestrate domain logic. Each service has a clear responsibility.

```go
// internal/service/interfaces.go

type ProjectService interface {
    // Detect finds the project root and loads context.
    Detect(ctx context.Context) (*domain.Project, error)

    // Health computes aggregate project health.
    Health(ctx context.Context) (*domain.ProjectHealth, error)

    // CheckBrief evaluates the business context gate.
    CheckBrief(ctx context.Context) (*domain.Brief, error)

    // Doctor runs deep diagnostics.
    Doctor(ctx context.Context) ([]domain.Diagnostic, error)

    // Fix auto-fixes resolvable diagnostics.
    Fix(ctx context.Context, diags []domain.Diagnostic) ([]domain.FixResult, error)

    // SuggestNext proposes what should happen next.
    SuggestNext(ctx context.Context) (*domain.Suggestion, error)
}

type ValidationService interface {
    // ValidateDocs runs the 17-check docs suite.
    ValidateDocs(ctx context.Context, strict bool) (*domain.ValidationReport, error)

    // ValidateRefs runs the 11-check cross-reference suite.
    ValidateRefs(ctx context.Context) (*domain.ValidationReport, error)

    // ValidateConfig runs YAML config validation.
    ValidateConfig(ctx context.Context) (*domain.ValidationReport, error)

    // ValidateConvergence runs the 23-check convergence suite.
    ValidateConvergence(ctx context.Context, path string) (*domain.ValidationReport, error)

    // ValidateAll runs every suite and returns a combined report.
    ValidateAll(ctx context.Context, strict bool) ([]domain.ValidationReport, error)

    // RunDeterministicGate executes build/lint/test commands.
    RunDeterministicGate(ctx context.Context) (*domain.GateResult, error)
}

type GenerateService interface {
    // CreateADR generates an ADR with auto-sequencing.
    CreateADR(ctx context.Context, title string) (string, error)

    // CreateBlueprint generates a blueprint + INDEX.md row.
    CreateBlueprint(ctx context.Context, title string) (string, error)

    // CreateIteration generates an iteration folder with 5 files.
    CreateIteration(ctx context.Context, reqType domain.RequestType, descriptor string) (string, error)

    // CreateSpike generates a spike report.
    CreateSpike(ctx context.Context, title string) (string, error)

    // CreateConvergence generates a convergence template.
    CreateConvergence(ctx context.Context, title string) (string, error)

    // CreateBrief runs interactive prompts to create a project brief.
    CreateBrief(ctx context.Context) (string, error)
}

type WorkflowService interface {
    // Status returns current workflow state.
    Status(ctx context.Context) (*domain.WorkflowState, error)

    // History returns past iterations.
    History(ctx context.Context) ([]domain.Iteration, error)

    // Clean removes stale workflow state.
    Clean(ctx context.Context, dryRun bool) error
}

type QualityService interface {
    // Log extracts scores from a convergence file and appends to quality-log.yml.
    Log(ctx context.Context, path string, topic string, variant string) error

    // History returns quality score entries over time.
    History(ctx context.Context) ([]domain.QualityEntry, error)
}

type SyncService interface {
    // SyncAgents syncs conversation agents to Copilot format.
    SyncAgents(ctx context.Context, check bool) (*domain.SyncReport, error)
}
```

### Service Implementation (Example)

```go
// internal/service/project.go

type projectService struct {
    docRepo   repo.DocRepo
    iterRepo  repo.IterationRepo
    stateRepo repo.StateRepo
    confRepo  repo.ConfigRepo
}

func NewProjectService(
    docRepo repo.DocRepo,
    iterRepo repo.IterationRepo,
    stateRepo repo.StateRepo,
    confRepo repo.ConfigRepo,
) ProjectService {
    return &projectService{docRepo, iterRepo, stateRepo, confRepo}
}

func (s *projectService) Health(ctx context.Context) (*domain.ProjectHealth, error) {
    project, err := s.Detect(ctx)
    if err != nil {
        return nil, err
    }

    brief, _ := s.CheckBrief(ctx)
    workflow, _ := s.stateRepo.ReadWorkflow()
    iterations, _ := s.iterRepo.List()

    zones := make(map[domain.Zone]domain.ZoneHealth)
    for _, zone := range domain.AllZones {
        docs, _ := s.docRepo.ListByZone(zone)
        zh := domain.ZoneHealth{Zone: zone, Total: len(docs)}
        for _, doc := range docs {
            zh.Present++
            if doc.IsStub {
                zh.Stubs++
            } else {
                zh.Complete++
            }
            zh.Files = append(zh.Files, doc)
        }
        zones[zone] = zh
    }

    var warnings []string
    if brief != nil && brief.GateResult != domain.BriefPresent {
        warnings = append(warnings, "Project brief is missing or a stub")
    }
    for zone, zh := range zones {
        if zh.Stubs > 0 {
            for _, f := range zh.Files {
                if f.IsStub {
                    warnings = append(warnings, f.Name+" is a stub ("+string(zone)+"/)")
                }
            }
        }
    }

    var lastIter *domain.Iteration
    if len(iterations) > 0 {
        lastIter = &iterations[0]
    }

    return &domain.ProjectHealth{
        Project:       *project,
        Brief:         *brief,
        Zones:         zones,
        Workflow:       workflow,
        LastIteration: lastIter,
        Warnings:      warnings,
    }, nil
}
```

---

## Error Handling

### Error Types

```go
// domain/errors.go

// ErrNotProject signals that no .mind/ was found.
var ErrNotProject = errors.New("not a Mind project (no .mind/ directory found)")

// ErrBriefMissing signals a missing project brief for gate enforcement.
var ErrBriefMissing = errors.New("project brief missing — run /discover or create docs/spec/project-brief.md")

// ErrGateFailed signals a quality gate failure.
type ErrGateFailed struct {
    Gate     string
    Failures []string
}

func (e *ErrGateFailed) Error() string {
    return fmt.Sprintf("gate %s failed: %s", e.Gate, strings.Join(e.Failures, "; "))
}

// ErrCommandFailed signals an external command failure.
type ErrCommandFailed struct {
    Command string
    ExitCode int
    Output   string
}

func (e *ErrCommandFailed) Error() string {
    return fmt.Sprintf("command %q failed (exit %d): %s", e.Command, e.ExitCode, e.Output)
}
```

### Error Propagation Rules

| Layer | Error Strategy |
|-------|---------------|
| **Infrastructure** | Return raw errors wrapped with `fmt.Errorf("reading %s: %w", path, err)` |
| **Domain** | Define sentinel errors and typed errors. No wrapping. |
| **Service** | Wrap infrastructure errors with domain context. Return domain error types. |
| **Presentation** | Convert errors to user-facing messages. Set exit codes. Never expose stack traces. |

```go
// cmd/root.go

func handleError(err error) {
    switch {
    case errors.Is(err, domain.ErrNotProject):
        fmt.Fprintln(os.Stderr, "Error: Not a Mind project. Run 'mind init' to set up.")
        os.Exit(1)
    case errors.Is(err, domain.ErrBriefMissing):
        fmt.Fprintln(os.Stderr, "Error:", err)
        fmt.Fprintln(os.Stderr, "Tip: Run '/discover' in Claude Code to create a project brief.")
        os.Exit(1)
    default:
        var gateFailed *domain.ErrGateFailed
        if errors.As(err, &gateFailed) {
            fmt.Fprintf(os.Stderr, "Gate %s failed:\n", gateFailed.Gate)
            for _, f := range gateFailed.Failures {
                fmt.Fprintf(os.Stderr, "  ✗ %s\n", f)
            }
            os.Exit(1)
        }
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(1)
    }
}
```

---

## Dependency Injection

### Application Bootstrap

All dependencies are wired at startup in `main.go`. No global state, no init() functions, no service locators.

```go
// main.go

func main() {
    root, err := project.FindRoot()
    if err != nil {
        // Some commands (init, version, help) work without a project root
        root = ""
    }

    // Infrastructure
    docRepo := fs.NewDocRepo(root)
    iterRepo := fs.NewIterationRepo(root)
    stateRepo := fs.NewStateRepo(root)
    confRepo := fs.NewConfigRepo(root)
    templateRepo := fs.NewTemplateRepo(root)
    qualityRepo := fs.NewQualityRepo(root)

    // Services
    projectSvc := service.NewProjectService(docRepo, iterRepo, stateRepo, confRepo)
    validateSvc := service.NewValidationService(docRepo, iterRepo, confRepo)
    generateSvc := service.NewGenerateService(iterRepo, templateRepo, docRepo)
    workflowSvc := service.NewWorkflowService(stateRepo, iterRepo)
    qualitySvc := service.NewQualityService(qualityRepo)
    syncSvc := service.NewSyncService(docRepo)

    // Output
    mode := render.DetectMode(globalFlags.json, globalFlags.noColor)
    renderer := render.NewRenderer(mode, termWidth())

    // Wire into Cobra
    app := cmd.NewApp(cmd.Deps{
        ProjectSvc:  projectSvc,
        ValidateSvc: validateSvc,
        GenerateSvc: generateSvc,
        WorkflowSvc: workflowSvc,
        QualitySvc:  qualitySvc,
        SyncSvc:     syncSvc,
        Renderer:    renderer,
    })

    if err := app.Execute(); err != nil {
        handleError(err)
    }
}
```

### Context Propagation

Services receive `context.Context` for cancellation. Dependencies are passed via struct fields (constructor injection), not context values.

```go
// cmd/app.go

type Deps struct {
    ProjectSvc  service.ProjectService
    ValidateSvc service.ValidationService
    GenerateSvc service.GenerateService
    WorkflowSvc service.WorkflowService
    QualitySvc  service.QualityService
    SyncSvc     service.SyncService
    Renderer    render.Renderer
}

func NewApp(deps Deps) *cobra.Command {
    root := &cobra.Command{Use: "mind"}

    root.AddCommand(
        newStatusCmd(deps),
        newDoctorCmd(deps),
        newCreateCmd(deps),
        newDocsCmd(deps),
        newCheckCmd(deps),
        newWorkflowCmd(deps),
        newQualityCmd(deps),
        newSyncCmd(deps),
        newTuiCmd(deps),
        newServeCmd(deps),     // MCP server
        newPreflightCmd(deps), // Model A
        newRunCmd(deps),       // Model D
        newWatchCmd(deps),     // Model C
    )

    return root
}

func newStatusCmd(deps Deps) *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Project health dashboard",
        RunE: func(cmd *cobra.Command, args []string) error {
            health, err := deps.ProjectSvc.Health(cmd.Context())
            if err != nil {
                return err
            }
            fmt.Print(deps.Renderer.RenderHealth(health))
            return nil
        },
    }
}
```

---

## Concurrency Model

### Goroutine Ownership

| Component | Goroutines | Communication |
|-----------|-----------|---------------|
| **CLI commands** | Single goroutine (main) | Direct return values |
| **TUI** | 1 main + N background (data loading, validation) | Bubble Tea messages (tea.Cmd) |
| **Watch mode** | 1 fsnotify + 1 debounce + N background commands | Channels (Event, Action) |
| **MCP server** | 1 stdio reader + 1 per request | Channel per request (request/response) |
| **Orchestration** | 1 pipeline + 1 per agent dispatch | PipelineEvent channel |

### Rules

1. **Never share mutable state between goroutines.** Pass messages through channels.
2. **All background work is cancellable** via `context.Context`.
3. **Background commands (build, test) are debounced** — 300ms window.
4. **TUI uses `tea.Cmd` exclusively** for async work. Never block the Update loop.

```go
// tui/app.go — async pattern

func (m Model) loadHealth() tea.Msg {
    health, err := m.projectSvc.Health(context.Background())
    if err != nil {
        return healthErrorMsg{err}
    }
    return healthLoadedMsg{health}
}

func (m Model) runValidation() tea.Msg {
    report, err := m.validateSvc.ValidateAll(context.Background(), false)
    if err != nil {
        return validationErrorMsg{err}
    }
    return validationCompleteMsg{report}
}
```

---

## Testing Strategy

### Test Pyramid

```
                 ┌───────────┐
                 │    E2E     │   5-10 tests: full CLI invocations
                 │  (golden)  │   against real filesystem
                 ├───────────┤
                 │Integration │   20-30 tests: service + repo
                 │(in-memory) │   with MemDocRepo/MemIterRepo
                 ├───────────┤
                 │   Unit     │   100+ tests: domain logic,
                 │            │   validation checks, slugify,
                 │            │   stub detection, classification
                 └───────────┘
```

### Unit Tests

Domain logic tests with zero dependencies:

```go
// domain/iteration_test.go

func TestClassify(t *testing.T) {
    tests := []struct {
        input    string
        expected RequestType
    }{
        {"create: REST API", TypeNewProject},
        {"build a CLI tool", TypeNewProject},
        {"fix: 500 error on /api", TypeBugFix},
        {"bug in auth", TypeBugFix},
        {"add: WebSocket support", TypeEnhancement},
        {"feature: dark mode", TypeEnhancement},
        {"refactor: repository pattern", TypeRefactor},
        {"clean up database layer", TypeRefactor},
        {"analyze: CRDT vs OT", TypeComplexNew},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got := Classify(tt.input)
            if got != tt.expected {
                t.Errorf("Classify(%q) = %s, want %s", tt.input, got, tt.expected)
            }
        })
    }
}

func TestSlugify(t *testing.T) {
    tests := []struct{ input, expected string }{
        {"Chose PostgreSQL", "chose-postgresql"},
        {"fix: 500 Error on /api/users", "fix-500-error-on-api-users"},
        {"  Leading Spaces  ", "leading-spaces"},
        {"Special!@#$chars", "special-chars"},
    }
    for _, tt := range tests {
        if got := Slugify(tt.input); got != tt.expected {
            t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.expected)
        }
    }
}
```

### Integration Tests

Test services with in-memory repos:

```go
// internal/service/project_test.go

func TestHealthDetectsStubs(t *testing.T) {
    docRepo := mem.NewMemDocRepo()
    docRepo.AddFile("docs/spec/project-brief.md", "# Project Brief\n## Vision\n<!-- placeholder -->")
    docRepo.AddFile("docs/spec/requirements.md", "# Requirements\n\nFR-1: Users can log in\n...")
    // ... add more files

    svc := service.NewProjectService(docRepo, mem.NewMemIterRepo(), ...)
    health, err := svc.Health(context.Background())
    require.NoError(t, err)
    assert.True(t, health.Brief.IsStub)
    assert.Equal(t, domain.BriefStub, health.Brief.GateResult)
}
```

### Golden File Tests (E2E)

Test full CLI output against saved "golden" files:

```go
// cmd/status_test.go

func TestStatusOutput(t *testing.T) {
    // Set up a temp directory with a known project structure
    dir := setupTestProject(t)

    // Run the CLI
    out, err := exec.Command("mind", "status", "--no-color").Output()
    require.NoError(t, err)

    // Compare against golden file
    golden := filepath.Join("testdata", "status.golden")
    if *update {
        os.WriteFile(golden, out, 0644)
    }
    expected, _ := os.ReadFile(golden)
    assert.Equal(t, string(expected), string(out))
}
```

### Validation Check Tests

Each validation check is independently testable:

```go
// internal/validate/docs_test.go

func TestCheckBriefCompleteness(t *testing.T) {
    tests := []struct {
        name    string
        content string
        pass    bool
    }{
        {
            name:    "all sections present",
            content: "# Brief\n## Vision\nBuild X\n## Key Deliverables\nY\n## Scope\nZ",
            pass:    true,
        },
        {
            name:    "missing scope",
            content: "# Brief\n## Vision\nBuild X\n## Key Deliverables\nY",
            pass:    false,
        },
        {
            name:    "stub file",
            content: "# Brief\n## Vision\n<!-- placeholder -->\n## Key Deliverables\n<!-- fill -->",
            pass:    true, // Caught by stub check, not this check
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := mem.NewMemDocRepo()
            repo.AddFile("docs/spec/project-brief.md", tt.content)
            ctx := &CheckContext{DocRepo: repo}

            passed, _ := checkBriefCompleteness(ctx)
            assert.Equal(t, tt.pass, passed)
        })
    }
}
```

---

## Configuration

### Precedence

```
1. CLI flags          (highest priority)    --strict, --json, --no-color
2. Environment vars                         MIND_ROOT, MIND_NO_COLOR
3. mind.toml                                [project.commands], [governance]
4. Auto-detection                           TTY, terminal width, project root
5. Defaults           (lowest priority)     plain output, 80 cols, 2 retries
```

### Global Flags

```go
// cmd/root.go

type GlobalFlags struct {
    JSON    bool   // --json
    NoColor bool   // --no-color
    Root    string // --root (override project root detection)
    Verbose bool   // --verbose
}

func addGlobalFlags(cmd *cobra.Command) {
    cmd.PersistentFlags().BoolVar(&flags.JSON, "json", false, "Output as JSON")
    cmd.PersistentFlags().BoolVar(&flags.NoColor, "no-color", false, "Disable colored output")
    cmd.PersistentFlags().StringVar(&flags.Root, "root", "", "Override project root detection")
    cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
}
```

---

## Styles (Lip Gloss)

```go
// tui/styles.go

type Styles struct {
    Title     lipgloss.Style
    Subtitle  lipgloss.Style
    Success   lipgloss.Style
    Warning   lipgloss.Style
    Error     lipgloss.Style
    Muted     lipgloss.Style
    TabActive lipgloss.Style
    TabNormal lipgloss.Style
    Box       lipgloss.Style
    Bar       lipgloss.Style
    BarFill   lipgloss.Style
}

func DefaultStyles() Styles {
    return Styles{
        Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
        Subtitle:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
        Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
        Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
        Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
        Muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
        TabActive: lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("12")),
        TabNormal: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
        Box: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("8")).
            Padding(1, 2),
        Bar:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
        BarFill: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
    }
}

func renderProgressBar(complete, total, width int) string {
    if total == 0 { return strings.Repeat("░", width) }
    filled := int(float64(complete) / float64(total) * float64(width))
    return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
```

---

## Project Structure (Complete)

```
mind-cli/
├── main.go                         Entry point + DI wiring
├── go.mod
├── go.sum
├── Makefile                        Build, test, lint, release
├── .goreleaser.yml                 Cross-compilation + AUR
│
├── cmd/                            Presentation: Cobra commands
│   ├── app.go                      Command tree + Deps struct
│   ├── root.go                     Global flags, error handler
│   ├── status.go                   mind status
│   ├── doctor.go                   mind doctor [--fix]
│   ├── init_cmd.go                 mind init [--name]
│   ├── create.go                   mind create {adr,blueprint,...}
│   ├── docs.go                     mind docs {list,tree,stubs,search,open}
│   ├── check.go                    mind check {docs,refs,config,convergence,all}
│   ├── workflow.go                 mind workflow {status,history,show,clean}
│   ├── quality.go                  mind quality {log,history,report}
│   ├── sync.go                     mind sync {agents}
│   ├── tui_cmd.go                  mind tui
│   ├── serve.go                    mind serve (MCP server)
│   ├── preflight.go                mind preflight (Model A)
│   ├── run.go                      mind run (Model D)
│   ├── watch.go                    mind watch (Model C)
│   ├── completion.go               mind completion {bash,zsh,fish}
│   └── version.go                  mind version
│
├── domain/                         Domain types (zero external deps)
│   ├── project.go                  Project, Config, Manifest
│   ├── document.go                 Document, Zone, Brief, BriefGate
│   ├── iteration.go                Iteration, RequestType, AgentChain, AgentRef
│   ├── workflow.go                 WorkflowState, DispatchEntry
│   ├── validation.go               CheckResult, ValidationReport
│   ├── quality.go                  QualityScore, QualityEntry
│   ├── health.go                   ProjectHealth, ZoneHealth
│   ├── diagnostic.go               Diagnostic, FixResult, Suggestion
│   ├── errors.go                   Sentinel + typed errors
│   └── classify.go                 Request classification logic
│
├── internal/
│   ├── repo/                       Repository interfaces
│   │   └── interfaces.go           DocRepo, IterationRepo, StateRepo, ...
│   │
│   ├── repo/fs/                    Filesystem implementations
│   │   ├── doc_repo.go
│   │   ├── iteration_repo.go
│   │   ├── state_repo.go
│   │   ├── config_repo.go
│   │   ├── template_repo.go
│   │   └── quality_repo.go
│   │
│   ├── repo/mem/                   In-memory implementations (tests)
│   │   ├── doc_repo.go
│   │   ├── iteration_repo.go
│   │   └── state_repo.go
│   │
│   ├── service/                    Business logic
│   │   ├── interfaces.go           Service interfaces
│   │   ├── project.go              ProjectService impl
│   │   ├── validation.go           ValidationService impl
│   │   ├── generate.go             GenerateService impl
│   │   ├── workflow.go             WorkflowService impl
│   │   ├── quality.go              QualityService impl
│   │   └── sync.go                 SyncService impl
│   │
│   ├── validate/                   Validation engine
│   │   ├── check.go                Check, Suite, Run()
│   │   ├── docs.go                 17 docs checks
│   │   ├── refs.go                 11 cross-ref checks
│   │   ├── config.go               YAML config checks
│   │   └── convergence.go          23 convergence checks
│   │
│   ├── render/                     Output renderers
│   │   ├── render.go               Renderer interface + DetectMode
│   │   ├── interactive.go          Lip Gloss styled output
│   │   ├── plain.go                Plain text output
│   │   └── json.go                 JSON output
│   │
│   ├── orchestrate/                Model D: full orchestration
│   │   ├── pipeline.go             Pipeline, Step, PipelineState
│   │   ├── build.go                BuildPipeline()
│   │   ├── prompt.go               PromptBuilder
│   │   ├── classify.go             Request classification
│   │   └── executor.go             AgentExecutor (claude CLI wrapper)
│   │
│   ├── watch/                      Model C: filesystem watcher
│   │   ├── watcher.go              Watcher, Event, Action
│   │   └── handlers.go             DefaultHandlers
│   │
│   └── mcp/                        Model B: MCP server
│       ├── server.go               MCP JSON-RPC protocol
│       └── tools.go                RegisterTools()
│
├── tui/                            Bubble Tea TUI application
│   ├── app.go                      Model (MVU root)
│   ├── status.go                   Tab 1: status dashboard
│   ├── docs.go                     Tab 2: document browser
│   ├── iterations.go               Tab 3: iteration timeline
│   ├── checks.go                   Tab 4: validation results
│   ├── quality.go                  Tab 5: quality trends
│   ├── watch.go                    Watch mode TUI (Model C)
│   ├── pipeline.go                 Pipeline TUI (Model D)
│   ├── styles.go                   Lip Gloss style definitions
│   ├── keys.go                     Key bindings
│   └── components/                 Reusable UI components
│       ├── progressbar.go          Zone health bars
│       ├── table.go                Styled table wrapper
│       ├── tree.go                 File tree renderer
│       └── chart.go                ASCII quality score chart
│
└── testdata/                       Golden files for E2E tests
    ├── status.golden
    ├── doctor.golden
    ├── check-docs.golden
    └── projects/                   Fixture projects
        ├── complete/               Project with all docs filled
        ├── minimal/                Project with only required files
        └── stubs/                  Project with stub documents
```

---

## Build System

```makefile
# Makefile

BINARY = mind
VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -v -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

golden-update:
	go test ./cmd/... -update

clean:
	rm -f $(BINARY) coverage.out

release:
	goreleaser release --clean
```

---

## Extension Points

### Adding a New Validation Check

1. Write the check function in `internal/validate/docs.go`
2. Add it to the `DocsSuite()` check list
3. Write a test in `docs_test.go`
4. The check automatically appears in CLI, TUI, MCP, and JSON output

### Adding a New Document Type

1. Add template to `.mind/docs/templates/`
2. Add a case to `GenerateService.Create*()`
3. Add a subcommand to `cmd/create.go`
4. The TUI document browser auto-discovers it

### Adding a New TUI Tab

1. Create `tui/newtab.go` with a model implementing `Update()` + `View()`
2. Add the tab constant to `tui/app.go`
3. Wire it into the top-level `Model` struct
4. Add the key binding to the tab bar

### Adding a New MCP Tool

1. Add a method to the relevant service interface
2. Implement it in the service
3. Add a `Tool{}` entry in `mcp/tools.go`
4. The tool auto-registers when the MCP server starts

---

> **See also:**
> - [01-mind-cli.md](01-mind-cli.md) — Command tree, distribution, implementation phases
> - [02-ai-workflow-bridge.md](02-ai-workflow-bridge.md) — AI integration models (A, B, C, D)
> - `../../scripts/validate-docs.sh` — Original validation logic being ported
> - `../../scripts/docs-gen.sh` — Original generation logic being ported
> - `../../conversation/config/quality.yml` — Quality rubric schema
