# BP-02: Domain Model

> What are the core entities, their relationships, and the business rules?

**Status**: Active
**Date**: 2026-03-11
**Depends on**: [01-mind-cli.md](01-mind-cli.md), [03-architecture.md](03-architecture.md)

---

## 1. Entity Catalog

All entities live in the `domain/` package. The domain layer imports only the Go standard library — zero external dependencies. Entities are pure data structures with minimal behavior; business logic is orchestrated by the service layer.

---

### 1.1 Project

**Purpose**: Represents a detected Mind Framework project on disk. The root aggregate through which all operations flow.

**File**: `domain/project.go`

```go
type Project struct {
    Root      string  // Absolute path to project root (where .mind/ lives)
    Name      string  // From mind.toml [project].name
    Config    *Config // Parsed mind.toml manifest (nil if file doesn't exist)
    Framework string  // Framework version from .mind/CHANGELOG.md
    DocsRoot  string  // Root + "/docs"
    MindRoot  string  // Root + "/.mind"
}
```

| Field     | Type      | Description                                              |
|-----------|-----------|----------------------------------------------------------|
| Root      | string    | Absolute filesystem path where `.mind/` directory lives  |
| Name      | string    | Human-readable name, sourced from `mind.toml [project].name` |
| Config    | *Config   | Parsed mind.toml manifest; nil when the file is absent   |
| Framework | string    | Framework version string from `.mind/CHANGELOG.md`       |
| DocsRoot  | string    | Derived: `Root + "/docs"`                                |
| MindRoot  | string    | Derived: `Root + "/.mind"`                               |

---

### 1.2 Config

**Purpose**: Parsed representation of the `mind.toml` manifest. Central configuration that declares project identity, documents, governance rules, and dependency graph.

**File**: `domain/project.go`

```go
type Config struct {
    Manifest   Manifest                        `toml:"manifest"`
    Project    ProjectMeta                     `toml:"project"`
    Profiles   Profiles                        `toml:"profiles"`
    Documents  map[string]map[string]DocEntry  `toml:"documents"`
    Governance Governance                      `toml:"governance"`
}
```

| Field      | Type                              | Description                                                    |
|------------|-----------------------------------|----------------------------------------------------------------|
| Manifest   | Manifest                          | Schema version, generation counter, last-updated timestamp     |
| Project    | ProjectMeta                       | Name, description, type, stack configuration, build commands   |
| Profiles   | Profiles                          | List of active profile names                                   |
| Documents  | map[string]map[string]DocEntry    | Two-level map: zone name -> document key -> DocEntry           |
| Governance | Governance                        | Max retries, review policy, commit policy, branch strategy     |

#### Supporting Types

```go
type Manifest struct {
    Schema     string          `toml:"schema"`     // e.g. "mind/v2.0"
    Generation int             `toml:"generation"` // Monotonically increasing
    Updated    time.Time       `toml:"updated"`    // Last modification timestamp
    Invariants map[string]bool `toml:"invariants"` // Structural invariant flags
}

type ProjectMeta struct {
    Name        string      `toml:"name"`
    Description string      `toml:"description"`
    Type        string      `toml:"type"`        // e.g. "cli", "api", "library"
    Stack       StackConfig `toml:"stack"`
    Commands    CmdConfig   `toml:"commands"`
}

type StackConfig struct {
    Language  string `toml:"language"`  // e.g. "go@1.23"
    Framework string `toml:"framework"` // e.g. "cobra+bubbletea"
    Testing   string `toml:"testing"`   // e.g. "go-test"
}

type CmdConfig struct {
    Dev       string `toml:"dev"`
    Test      string `toml:"test"`
    Lint      string `toml:"lint"`
    Typecheck string `toml:"typecheck"`
    Build     string `toml:"build"`
}

type Profiles struct {
    Active []string `toml:"active"`
}

type DocEntry struct {
    ID     string `toml:"id"`     // e.g. "doc:spec/project-brief"
    Path   string `toml:"path"`   // Relative to project root
    Zone   string `toml:"zone"`   // Zone name
    Status string `toml:"status"` // draft, active, complete
}

type Governance struct {
    MaxRetries     int    `toml:"max-retries"`
    ReviewPolicy   string `toml:"review-policy"`   // e.g. "evidence-based"
    CommitPolicy   string `toml:"commit-policy"`   // e.g. "conventional"
    BranchStrategy string `toml:"branch-strategy"` // e.g. "type-descriptor"
}
```

---

### 1.3 Document

**Purpose**: A single documentation file within the project. Tracks filesystem metadata and stub classification for health reporting.

**File**: `domain/document.go`

```go
type Document struct {
    Path    string    // Relative to project root
    AbsPath string    // Absolute filesystem path
    Zone    Zone      // Which documentation zone it belongs to
    Name    string    // Filename without extension
    Size    int64     // File size in bytes
    ModTime time.Time // Last modification time
    IsStub  bool      // True if detected as a stub by content analysis
    Status  DocStatus // Inferred from content or declared in mind.toml
    Hash    string    // SHA-256 content hash for reconciliation
}
```

| Field   | Type      | Description                                                        |
|---------|-----------|--------------------------------------------------------------------|
| Path    | string    | Relative path from project root, e.g. `docs/spec/requirements.md` |
| AbsPath | string    | Absolute filesystem path                                           |
| Zone    | Zone      | The documentation zone this file belongs to                        |
| Name    | string    | Filename stem without `.md` extension                              |
| Size    | int64     | File size in bytes                                                 |
| ModTime | time.Time | Last filesystem modification time                                  |
| IsStub  | bool      | True when stub detection finds <=2 real content lines              |
| Status  | DocStatus | Lifecycle status: draft, active, complete, or stub                 |
| Hash    | string    | SHA-256 hex digest of file content, used by lock file reconciliation |

---

### 1.4 Brief

**Purpose**: Parsed project brief with section detection. Used by the business context gate to determine whether a workflow can proceed.

**File**: `domain/document.go`

```go
type Brief struct {
    Path            string
    Exists          bool
    IsStub          bool
    HasVision       bool
    HasDeliverables bool
    HasScope        bool
    GateResult      BriefGate
}
```

| Field           | Type      | Description                                                   |
|-----------------|-----------|---------------------------------------------------------------|
| Path            | string    | Path to `docs/spec/project-brief.md`                          |
| Exists          | bool      | True if the file exists on disk                               |
| IsStub          | bool      | True if the file exists but is a stub                         |
| HasVision       | bool      | True if a `## Vision` (or equivalent) section has content     |
| HasDeliverables | bool      | True if a `## Key Deliverables` section has content           |
| HasScope        | bool      | True if a `## Scope` section has content                      |
| GateResult      | BriefGate | Computed gate classification: PRESENT, STUB, or MISSING       |

---

### 1.5 Iteration

**Purpose**: Represents a single workflow iteration folder under `docs/iterations/`. Each iteration tracks one unit of work from request through completion.

**File**: `domain/iteration.go`

```go
type Iteration struct {
    Seq        int             // Sequence number (1, 2, 3...)
    Type       RequestType     // NEW_PROJECT, BUG_FIX, etc.
    Descriptor string          // Kebab-case slug, e.g. "rest-api"
    DirName    string          // Full directory name: "001-NEW_PROJECT-rest-api"
    Path       string          // Absolute path to iteration directory
    Artifacts  []Artifact      // Files in the iteration folder
    Status     IterationStatus // Derived from artifact presence
    CreatedAt  time.Time       // From overview.md modification time
}
```

| Field      | Type            | Description                                                           |
|------------|-----------------|-----------------------------------------------------------------------|
| Seq        | int             | Monotonically increasing sequence number, zero-padded to 3 digits on disk |
| Type       | RequestType     | The classification of the originating request                         |
| Descriptor | string          | Kebab-case slug derived from the request description                  |
| DirName    | string          | Directory name on disk, format: `{seq}-{TYPE}-{descriptor}`          |
| Path       | string          | Absolute filesystem path to the iteration directory                   |
| Artifacts  | []Artifact      | The expected 5 artifact files and their presence status               |
| Status     | IterationStatus | Auto-computed: complete if all 5 artifacts exist, incomplete otherwise |
| CreatedAt  | time.Time       | Derived from the modification time of `overview.md`                   |

---

### 1.6 Artifact

**Purpose**: A file within an iteration folder. Tracks whether the expected artifact has been produced.

**File**: `domain/iteration.go`

```go
type Artifact struct {
    Name   string // overview.md, changes.md, etc.
    Path   string // Absolute path
    Exists bool   // True if file exists on disk
}
```

| Field  | Type   | Description                                           |
|--------|--------|-------------------------------------------------------|
| Name   | string | Expected filename: overview.md, changes.md, test-summary.md, validation.md, or retrospective.md |
| Path   | string | Absolute filesystem path to the artifact              |
| Exists | bool   | True if the file exists on disk                       |

**Expected Artifacts** (defined as a package-level variable):

```go
var ExpectedArtifacts = []string{
    "overview.md",
    "changes.md",
    "test-summary.md",
    "validation.md",
    "retrospective.md",
}
```

---

### 1.7 WorkflowState

**Purpose**: Persisted state of an in-progress workflow. Serialized to `docs/state/workflow.md` as a markdown-wrapped JSON block. Enables pause/resume across sessions.

**File**: `domain/workflow.go`

```go
type WorkflowState struct {
    Type           RequestType         `json:"type"`
    Descriptor     string              `json:"descriptor"`
    IterationPath  string              `json:"iteration_path"`
    Branch         string              `json:"branch"`
    LastAgent      string              `json:"last_agent"`
    RemainingChain []string            `json:"remaining_chain"`
    Session        int                 `json:"session"`
    TotalSessions  int                 `json:"total_sessions"`
    Artifacts      []CompletedArtifact `json:"artifacts,omitempty"`
    DispatchLog    []DispatchEntry     `json:"dispatch_log,omitempty"`
    Decisions      []string            `json:"decisions,omitempty"`
    HandoffContext string              `json:"handoff_context,omitempty"`
}
```

| Field          | Type                | Description                                                    |
|----------------|---------------------|----------------------------------------------------------------|
| Type           | RequestType         | Classification of the originating request                      |
| Descriptor     | string              | Kebab-case slug for the workflow                               |
| IterationPath  | string              | Absolute path to the iteration directory                       |
| Branch         | string              | Git branch name, e.g. `new/rest-api`                          |
| LastAgent      | string              | Name of the last agent that completed                          |
| RemainingChain | []string            | Ordered list of agents still to run                            |
| Session        | int                 | Current session number (for split workflows)                   |
| TotalSessions  | int                 | Expected total sessions                                        |
| Artifacts      | []CompletedArtifact | Outputs produced by completed agents                           |
| DispatchLog    | []DispatchEntry     | Chronological log of all agent dispatches                      |
| Decisions      | []string            | Key decisions recorded during the workflow                     |
| HandoffContext | string              | Free-form context passed to the next session on resume         |

**Method**: `IsIdle() bool` — returns true if no workflow is in progress (`s == nil || s.Type == ""`).

---

### 1.8 CompletedArtifact

**Purpose**: Records an output file produced by a completed agent within a workflow.

**File**: `domain/workflow.go`

```go
type CompletedArtifact struct {
    Agent    string `json:"agent"`
    Output   string `json:"output"`
    Location string `json:"location"`
}
```

| Field    | Type   | Description                                        |
|----------|--------|----------------------------------------------------|
| Agent    | string | Agent name that produced this output (e.g. "analyst") |
| Output   | string | Description of what was produced                   |
| Location | string | File path where the output was written             |

---

### 1.9 DispatchEntry

**Purpose**: Log entry for a single agent dispatch. Captures timing and outcome for workflow observability.

**File**: `domain/workflow.go`

```go
type DispatchEntry struct {
    Agent     string        `json:"agent"`
    File      string        `json:"file"`
    Model     string        `json:"model"`
    Status    string        `json:"status"`
    StartedAt time.Time     `json:"started_at"`
    Duration  time.Duration `json:"duration"`
}
```

| Field     | Type          | Description                                                          |
|-----------|---------------|----------------------------------------------------------------------|
| Agent     | string        | Agent name (analyst, architect, developer, tester, reviewer, moderator) |
| File      | string        | Path to the agent's markdown definition file                         |
| Model     | string        | AI model used (opus, sonnet, haiku)                                  |
| Status    | string        | One of: `dispatched`, `completed`, `failed`, `retrying`              |
| StartedAt | time.Time     | When the dispatch began                                              |
| Duration  | time.Duration | How long the agent ran                                               |

---

### 1.10 ValidationReport

**Purpose**: Aggregated results from a validation suite. Produced by `mind check` commands and consumed by health reporting and gate enforcement.

**File**: `domain/validation.go`

```go
type ValidationReport struct {
    Suite    string        `json:"suite"`
    Checks   []CheckResult `json:"checks"`
    Total    int           `json:"total"`
    Passed   int           `json:"passed"`
    Failed   int           `json:"failed"`
    Warnings int           `json:"warnings"`
}
```

| Field    | Type          | Description                                                    |
|----------|---------------|----------------------------------------------------------------|
| Suite    | string        | Suite identifier: `"docs"`, `"refs"`, `"config"`, `"convergence"` |
| Checks   | []CheckResult | Individual check outcomes                                      |
| Total    | int           | Total number of checks executed                                |
| Passed   | int           | Number of checks that passed                                   |
| Failed   | int           | Number of checks that failed                                   |
| Warnings | int           | Number of checks that produced warnings                        |

**Method**: `Ok() bool` — returns `Failed == 0`.

---

### 1.11 CheckResult

**Purpose**: Outcome of a single validation check within a suite.

**File**: `domain/validation.go`

```go
type CheckResult struct {
    ID      int        `json:"id"`
    Name    string     `json:"name"`
    Level   CheckLevel `json:"level"`
    Passed  bool       `json:"passed"`
    Message string     `json:"message,omitempty"`
}
```

| Field   | Type       | Description                                            |
|---------|------------|--------------------------------------------------------|
| ID      | int        | Numeric identifier for the check within its suite      |
| Name    | string     | Human-readable check name                              |
| Level   | CheckLevel | Severity: FAIL, WARN, or INFO                         |
| Passed  | bool       | True if the check passed                               |
| Message | string     | Explanation when the check fails or warns              |

---

### 1.12 Diagnostic

**Purpose**: Deep diagnostic result from `mind doctor`. Each diagnostic identifies an issue, its severity, and optionally a fix.

**File**: `domain/health.go`

```go
type Diagnostic struct {
    Level   CheckLevel `json:"level"`
    Message string     `json:"message"`
    Fix     string     `json:"fix,omitempty"`
    AutoFix bool       `json:"auto_fix"`
}
```

| Field   | Type       | Description                                                      |
|---------|------------|------------------------------------------------------------------|
| Level   | CheckLevel | Severity of the diagnostic finding (FAIL, WARN, INFO)           |
| Message | string     | Description of what was found                                    |
| Fix     | string     | Suggested remediation action (human-readable command or step)    |
| AutoFix | bool       | True if `mind doctor --fix` can resolve this automatically       |

Note: The specification calls for a `Category` and `Check` field with a dedicated `DiagStatus` enum. The current implementation reuses `CheckLevel` for severity. The entity is expected to evolve to include category-based grouping (e.g., "framework", "documentation", "workflow", "config") and a dedicated status type as `mind doctor` gains deeper diagnostic capabilities.

---

### 1.13 ProjectHealth

**Purpose**: Aggregate status for `mind status`. Composes project metadata, brief status, per-zone health, workflow state, and actionable suggestions.

**File**: `domain/health.go`

```go
type ProjectHealth struct {
    Project       Project             `json:"project"`
    Brief         Brief               `json:"brief"`
    Zones         map[Zone]ZoneHealth `json:"zones"`
    Workflow      *WorkflowState      `json:"workflow,omitempty"`
    LastIteration *Iteration          `json:"last_iteration,omitempty"`
    Warnings      []string            `json:"warnings,omitempty"`
    Suggestions   []string            `json:"suggestions,omitempty"`
}
```

| Field         | Type                | Description                                            |
|---------------|---------------------|--------------------------------------------------------|
| Project       | Project             | Core project identity and paths                        |
| Brief         | Brief               | Business context gate status                           |
| Zones         | map[Zone]ZoneHealth | Per-zone documentation completeness                    |
| Workflow      | *WorkflowState      | Current workflow state (nil if idle)                   |
| LastIteration | *Iteration          | Most recent iteration (nil if none exist)              |
| Warnings      | []string            | Human-readable warnings for display                    |
| Suggestions   | []string            | Actionable next-step suggestions                       |

---

### 1.14 ZoneHealth

**Purpose**: Tracks completeness of a single documentation zone. Used by `mind status` and the TUI dashboard to render per-zone health bars.

**File**: `domain/health.go`

```go
type ZoneHealth struct {
    Zone     Zone       `json:"zone"`
    Total    int        `json:"total"`
    Present  int        `json:"present"`
    Stubs    int        `json:"stubs"`
    Complete int        `json:"complete"`
    Files    []Document `json:"files,omitempty"`
}
```

| Field    | Type       | Description                                     |
|----------|------------|-------------------------------------------------|
| Zone     | Zone       | Which zone this health report covers            |
| Total    | int        | Total documents declared in mind.toml for this zone |
| Present  | int        | Number of documents that exist on disk          |
| Stubs    | int        | Number of documents detected as stubs           |
| Complete | int        | Number of documents with status "complete"      |
| Files    | []Document | Full document details (omitted in summary views) |

---

### 1.15 Suggestion

**Purpose**: An actionable next step, used by `mind status` and `mind doctor` to guide the user.

**File**: `domain/health.go`

```go
type Suggestion struct {
    Action  string `json:"action"`
    Reason  string `json:"reason"`
    Command string `json:"command,omitempty"`
}
```

| Field   | Type   | Description                                              |
|---------|--------|----------------------------------------------------------|
| Action  | string | What should be done (e.g. "Fill project brief")          |
| Reason  | string | Why it matters (e.g. "Required for NEW_PROJECT workflows") |
| Command | string | Optional CLI command to execute the suggestion           |

---

### 1.16 QualityScore

**Purpose**: Convergence quality assessment. Computed from a convergence analysis document's scoring section.

**File**: `domain/quality.go` (planned)

```go
type QualityScore struct {
    Overall    float64        // Weighted average, 1.0-5.0 scale
    Dimensions map[string]int // Dimension name -> 1-5 score
    Gate0Pass  bool           // True if Overall >= 3.0
}
```

| Field      | Type           | Description                                                  |
|------------|----------------|--------------------------------------------------------------|
| Overall    | float64        | Weighted average score across all dimensions (1.0-5.0 scale) |
| Dimensions | map[string]int | Individual dimension scores (e.g. "depth": 4, "balance": 3) |
| Gate0Pass  | bool           | True when `Overall >= 3.0` (Gate 0 threshold)                |

---

### 1.17 QualityEntry

**Purpose**: A row in `quality-log.yml`. Persists quality scores over time for trend analysis.

**File**: `domain/quality.go` (planned)

```go
type QualityEntry struct {
    Date       string         // ISO date of the analysis
    Topic      string         // Analysis topic
    SessionID  string         // Unique session identifier
    Overall    float64        // Overall score
    Dimensions map[string]int // Per-dimension scores
    Gate0Pass  bool           // Whether Gate 0 passed
    Personas   []string       // Personas used in the analysis
    Variant    string         // Analysis variant (e.g. "default", "deep")
    OutputPath string         // Path to the convergence output file
}
```

| Field      | Type           | Description                                           |
|------------|----------------|-------------------------------------------------------|
| Date       | string         | ISO 8601 date when the analysis was performed         |
| Topic      | string         | Subject of the convergence analysis                   |
| SessionID  | string         | Unique identifier for the analysis session            |
| Overall    | float64        | Overall convergence quality score (1.0-5.0)           |
| Dimensions | map[string]int | Per-dimension scores keyed by dimension name          |
| Gate0Pass  | bool           | True if Overall >= 3.0                                |
| Personas   | []string       | List of persona names used in the analysis            |
| Variant    | string         | Analysis configuration variant                        |
| OutputPath | string         | Path to the convergence document that was scored      |

---

### 1.18 AgentChain

**Purpose**: Defines the ordered sequence of agents to dispatch for a given request type.

**File**: `domain/workflow.go` (planned)

```go
type AgentChain struct {
    Type   RequestType // Which request type this chain serves
    Agents []AgentRef  // Ordered sequence of agent references
}
```

| Field  | Type        | Description                                      |
|--------|-------------|--------------------------------------------------|
| Type   | RequestType | The request classification this chain handles    |
| Agents | []AgentRef  | Ordered list of agents to dispatch sequentially  |

---

### 1.19 AgentRef

**Purpose**: Reference to a single agent within a chain. Carries the agent's identity, location, preferred model, and optionality.

**File**: `domain/workflow.go` (planned)

```go
type AgentRef struct {
    Name     string // analyst, architect, developer, tester, reviewer, moderator
    File     string // Path to agent markdown definition
    Model    string // Preferred AI model: opus, sonnet, haiku
    Optional bool   // True if the agent can be skipped
}
```

| Field    | Type   | Description                                                      |
|----------|--------|------------------------------------------------------------------|
| Name     | string | Agent role name                                                  |
| File     | string | Path to agent markdown definition under `.mind/agents/`          |
| Model    | string | Preferred AI model for dispatch                                  |
| Optional | bool   | True if the chain can skip this agent (e.g. architect in ENHANCEMENT) |

---

### 1.20 LockFile

**Purpose**: Reconciliation state snapshot. Tracks content hashes for all declared documents to detect drift between mind.toml declarations and filesystem reality.

**File**: `domain/lock.go` (planned)

```go
type LockFile struct {
    SchemaVersion string               // Lock file format version
    GeneratedAt   time.Time            // When this lock file was computed
    ProjectHash   string               // SHA-256 of mind.toml content
    Entries       map[string]LockEntry // Keyed by document ID (e.g. "doc:spec/project-brief")
}
```

| Field         | Type                 | Description                                              |
|---------------|----------------------|----------------------------------------------------------|
| SchemaVersion | string               | Lock file format version for forward compatibility       |
| GeneratedAt   | time.Time            | Timestamp when the lock file was last regenerated        |
| ProjectHash   | string               | SHA-256 hash of mind.toml, detects manifest changes      |
| Entries       | map[string]LockEntry | Per-document entries keyed by document ID                |

---

### 1.21 LockEntry

**Purpose**: Per-document hash entry within the lock file. Enables staleness detection and dependency-aware invalidation.

**File**: `domain/lock.go` (planned)

```go
type LockEntry struct {
    ID          string   // Document ID, e.g. "doc:spec/project-brief"
    Path        string   // Relative path from project root
    Hash        string   // SHA-256 hex digest of file content
    Size        int64    // File size in bytes
    ModTime     time.Time // Last modification time
    Stale       bool     // True if content has changed since last lock
    StaleReason string   // Why this entry is stale
    DependsOn   []string // Document IDs this document depends on
}
```

| Field       | Type     | Description                                                       |
|-------------|----------|-------------------------------------------------------------------|
| ID          | string   | Stable document identifier matching mind.toml declarations        |
| Path        | string   | Relative filesystem path                                          |
| Hash        | string   | SHA-256 hex digest of the file content at lock time               |
| Size        | int64    | File size at lock time                                            |
| ModTime     | time.Time | Filesystem modification time at lock time                        |
| Stale       | bool     | True when current hash differs from locked hash                   |
| StaleReason | string   | Human-readable explanation of why the entry is stale              |
| DependsOn   | []string | Document IDs that this document depends on (for cascade invalidation) |

---

### 1.22 DependencyEdge

**Purpose**: Represents a directed relationship between two documents in the dependency graph declared in mind.toml.

**File**: `domain/lock.go` (planned)

```go
type DependencyEdge struct {
    From string // Source document ID
    To   string // Target document ID
    Type string // Relationship type: "informs", "requires", "validates"
}
```

| Field | Type   | Description                                                   |
|-------|--------|---------------------------------------------------------------|
| From  | string | Source document ID (the one that provides information)        |
| To    | string | Target document ID (the one that consumes information)        |
| Type  | string | `"informs"` (soft), `"requires"` (hard), or `"validates"` (check) |

---

### 1.23 Domain Errors

**Purpose**: Typed errors for domain-level failure conditions.

**File**: `domain/errors.go`

```go
// ErrNotProject signals that no .mind/ was found.
var ErrNotProject = errors.New("not a Mind project (no .mind/ directory found)")

// ErrBriefMissing signals a missing project brief for gate enforcement.
var ErrBriefMissing = errors.New(
    "project brief missing — run /discover or create docs/spec/project-brief.md",
)

// ErrGateFailed signals a quality gate failure.
type ErrGateFailed struct {
    Gate     string
    Failures []string
}

// ErrCommandFailed signals an external command failure.
type ErrCommandFailed struct {
    Command  string
    ExitCode int
    Output   string
}
```

---

### 1.24 Domain Functions

**Purpose**: Pure functions that operate on domain types with no I/O.

**File**: `domain/iteration.go`

```go
// Slugify converts a title to a kebab-case slug.
func Slugify(s string) string

// Classify determines the RequestType from a natural language description.
// Uses prefix matching (strongest signal) then keyword matching.
// Defaults to TypeEnhancement for ambiguous requests.
func Classify(request string) RequestType
```

`Classify` implements a deterministic heuristic:

1. **Prefix matching** (strongest): `create:` / `build:` -> NEW_PROJECT, `fix:` -> BUG_FIX, `add:` -> ENHANCEMENT, `refactor:` -> REFACTOR, `analyze:` / `explore:` -> COMPLEX_NEW.
2. **Keyword matching**: scans for domain-specific keywords (e.g. "bug", "error", "crash" -> BUG_FIX).
3. **Default**: ENHANCEMENT for ambiguous requests.

---

## 2. Value Objects & Enums

### 2.1 Zone

The five fixed documentation zones. Every document belongs to exactly one zone.

**File**: `domain/zones.go`

```go
type Zone string

const (
    ZoneSpec       Zone = "spec"       // Stable specifications (brief, requirements, architecture, domain model)
    ZoneBlueprints Zone = "blueprints" // Planning artifacts (INDEX.md + numbered blueprints)
    ZoneState      Zone = "state"      // Volatile runtime state (current.md, workflow.md)
    ZoneIterations Zone = "iterations" // Immutable history (per-change folders with 5 artifacts)
    ZoneKnowledge  Zone = "knowledge"  // Reference material (glossary, spikes, convergence analyses)
)

var AllZones = []Zone{ZoneSpec, ZoneBlueprints, ZoneState, ZoneIterations, ZoneKnowledge}
```

| Zone       | Mutability | Purpose                                      |
|------------|------------|----------------------------------------------|
| spec       | Stable     | Evolves slowly; core project specifications  |
| blueprints | Stable     | Planning-phase artifacts; move to completed  |
| state      | Volatile   | Changes frequently during workflows          |
| iterations | Immutable  | Historical record; never modified after completion |
| knowledge  | Stable     | Reference material; grows over project lifetime |

---

### 2.2 DocStatus

Lifecycle status of a document.

**File**: `domain/document.go`

```go
type DocStatus string

const (
    DocDraft    DocStatus = "draft"    // Initial state; content being developed
    DocActive   DocStatus = "active"   // In use and maintained
    DocComplete DocStatus = "complete" // Finalized; changes rare
    DocStub     DocStatus = "stub"     // Placeholder with no substantive content
)
```

---

### 2.3 BriefGate

Classification of the project brief for the business context gate.

**File**: `domain/document.go`

```go
type BriefGate string

const (
    BriefPresent BriefGate = "BRIEF_PRESENT" // Brief exists with Vision, Deliverables, and Scope
    BriefStub    BriefGate = "BRIEF_STUB"    // Brief exists but is a stub
    BriefMissing BriefGate = "BRIEF_MISSING" // Brief file does not exist
)
```

---

### 2.4 RequestType

Classification of a user's workflow request. Determines the agent chain, business context gate behavior, and iteration naming.

**File**: `domain/iteration.go`

```go
type RequestType string

const (
    TypeNewProject  RequestType = "NEW_PROJECT"  // Greenfield project or major component
    TypeBugFix      RequestType = "BUG_FIX"      // Defect repair
    TypeEnhancement RequestType = "ENHANCEMENT"  // Feature addition to existing codebase
    TypeRefactor    RequestType = "REFACTOR"      // Structural improvement without behavior change
    TypeComplexNew  RequestType = "COMPLEX_NEW"  // Requires multi-persona analysis before implementation
)
```

---

### 2.5 IterationStatus

Completeness state of an iteration, derived from artifact presence.

**File**: `domain/iteration.go`

```go
type IterationStatus string

const (
    IterInProgress IterationStatus = "in_progress" // Iteration created, artifacts being produced
    IterComplete   IterationStatus = "complete"     // All 5 expected artifacts present
    IterIncomplete IterationStatus = "incomplete"   // Some artifacts missing after workflow ended
)
```

---

### 2.6 CheckLevel

Severity level for validation checks and diagnostics.

**File**: `domain/validation.go`

```go
type CheckLevel string

const (
    LevelFail CheckLevel = "FAIL" // Blocks gate passage; must be resolved
    LevelWarn CheckLevel = "WARN" // Non-blocking; should be addressed
    LevelInfo CheckLevel = "INFO" // Informational; no action required
)
```

---

### 2.7 DiagStatus (planned)

Status for deep diagnostic results from `mind doctor`.

```go
type DiagStatus string

const (
    DiagPass DiagStatus = "pass" // Check passed cleanly
    DiagFail DiagStatus = "fail" // Issue found that needs resolution
    DiagWarn DiagStatus = "warn" // Potential issue that deserves attention
)
```

Note: The current implementation reuses `CheckLevel` for diagnostic severity. `DiagStatus` is planned for when `mind doctor` gains category-based diagnostic grouping.

---

### 2.8 EventOp (planned)

Filesystem event operations for watch mode (Model C).

```go
type EventOp string

const (
    EventCreated  EventOp = "Created"  // New file detected
    EventModified EventOp = "Modified" // Existing file changed
    EventDeleted  EventOp = "Deleted"  // File removed
)
```

---

### 2.9 OutputMode (planned)

Output rendering mode for CLI commands.

```go
type OutputMode string

const (
    OutputInteractive OutputMode = "Interactive" // Styled with Lip Gloss (TTY detected)
    OutputPlain       OutputMode = "Plain"       // Clean text, no ANSI codes
    OutputJSON        OutputMode = "JSON"        // Machine-readable JSON
)
```

---

### 2.10 GateStatus (planned)

Result of a quality gate evaluation.

```go
type GateStatus string

const (
    GatePassed  GateStatus = "passed"  // Gate requirements met
    GateFailed  GateStatus = "failed"  // Gate requirements not met
    GateSkipped GateStatus = "skipped" // Gate not applicable for this request type
    GatePending GateStatus = "pending" // Gate not yet evaluated
)
```

---

## 3. Entity Relationships

```
Project ──────┬── Config (mind.toml)
              │   ├── Manifest
              │   ├── ProjectMeta
              │   │   ├── StackConfig
              │   │   └── CmdConfig
              │   ├── Profiles
              │   ├── DocEntry[] (per-zone document declarations)
              │   ├── Governance
              │   └── DependencyEdge[] (document graph)
              │
              ├── Brief (docs/spec/project-brief.md)
              │   └── BriefGate
              │
              ├── Document[] ─── Zone
              │   ├── DocStatus
              │   └── Hash ─── LockEntry (reconciliation tracking)
              │
              ├── Iteration[] ─── Artifact[]
              │   ├── RequestType
              │   └── IterationStatus
              │
              ├── WorkflowState ─── DispatchEntry[]
              │   │                  └── CompletedArtifact[]
              │   └── AgentChain ─── AgentRef[]
              │
              └── ProjectHealth ─── ZoneHealth[]
                  ├── Suggestion[]
                  └── Diagnostic[]

LockFile ─── LockEntry[] ─── DependencyEdge[]
                              ├── From (doc ID)
                              └── To (doc ID)

ValidationReport ─── CheckResult[]
                     └── CheckLevel

AgentChain ─── AgentRef[]
               ├── Name (agent role)
               ├── Model (opus/sonnet/haiku)
               └── Optional (bool)

QualityScore ─── Dimensions (map)
                 └── Gate0Pass

QualityEntry[] ─── QualityScore (embedded scores)
                   └── OutputPath (link to convergence doc)
```

### Key Relationships Explained

| Relationship                  | Cardinality | Description                                                           |
|-------------------------------|-------------|-----------------------------------------------------------------------|
| Project -> Config             | 1:0..1      | Config is nil if mind.toml doesn't exist                              |
| Project -> Brief              | 1:1         | Always computed, even if file is missing (Exists=false)               |
| Project -> Document           | 1:N         | All documentation files discovered on disk                            |
| Project -> Iteration          | 1:N         | All iteration directories under docs/iterations/                      |
| Project -> WorkflowState      | 1:0..1      | Nil when no workflow is active                                        |
| Project -> ProjectHealth      | 1:1         | Computed on demand by the service layer                               |
| Iteration -> Artifact         | 1:5         | Exactly 5 expected artifacts per iteration                            |
| WorkflowState -> DispatchEntry | 1:N        | One entry per agent dispatch (including retries)                      |
| WorkflowState -> CompletedArtifact | 1:N    | One entry per successfully produced agent output                      |
| ValidationReport -> CheckResult | 1:N       | Variable number depending on suite                                    |
| LockFile -> LockEntry         | 1:N        | One entry per declared document                                       |
| LockEntry -> DependencyEdge   | N:M        | A document can depend on many and be depended upon by many            |
| AgentChain -> AgentRef        | 1:N        | 4-6 agents depending on request type                                  |
| Config -> DocEntry            | 1:N        | Two-level map: zone -> document key -> entry                          |

---

## 4. Business Rules

### BR-01: Project Detection

A directory is a valid Mind Framework project if and only if it contains a `.mind/` subdirectory. Detection walks upward from the current directory (like `git` finds `.git/`). If no `.mind/` is found, the domain error `ErrNotProject` is returned.

### BR-02: Stub Classification

A document is classified as a stub when it contains 2 or fewer "real content" lines. Real content excludes:
- Blank lines
- Markdown headings (`#`, `##`, etc.)
- HTML comments (`<!-- ... -->`)
- Template placeholders (lines containing only `<!-- ... -->`)

This is a heuristic applied to the file content, independent of the `status` field in mind.toml.

### BR-03: Business Context Gate — Blocking

The business context gate **blocks** workflows of type `NEW_PROJECT` and `COMPLEX_NEW` when the project brief is missing (`BRIEF_MISSING`) or is a stub (`BRIEF_STUB`). The workflow cannot proceed; the user must fill the brief first.

### BR-04: Business Context Gate — Warning

The business context gate **warns** (but does not block) `ENHANCEMENT` workflows when the brief is missing. The workflow proceeds with a warning that context may be insufficient.

### BR-05: Business Context Gate — Skip

The business context gate is **skipped** for `BUG_FIX` and `REFACTOR` request types. These workflows do not require business context to proceed.

### BR-06: Quality Gate 0

Convergence analysis quality Gate 0 requires an overall score of >= 3.0 out of 5.0. A score below this threshold indicates insufficient convergence quality and the analysis should be revised.

### BR-07: Agent Retry Limit

A maximum of 2 retry loops are permitted per agent per workflow. If an agent fails its quality gate after 2 retries, the workflow proceeds with documented concerns rather than retrying indefinitely. This is configured via `governance.max-retries` in mind.toml.

### BR-08: Iteration Sequence Numbering

Iteration sequence numbers are monotonically increasing integers, zero-padded to 3 digits on disk (e.g., `001`, `002`, `012`). The next sequence number is computed by scanning existing iteration directories and incrementing the highest found.

### BR-09: Iteration Directory Naming

Iteration directories follow the format `{seq}-{TYPE}-{descriptor}` where:
- `{seq}` is the 3-digit zero-padded sequence number
- `{TYPE}` is the RequestType value (e.g., `NEW_PROJECT`, `BUG_FIX`)
- `{descriptor}` is the kebab-case slug of the request description

Example: `007-NEW_PROJECT-rest-api`

### BR-10: Branch Naming

Git branches follow the `{type}/{descriptor}` pattern where `{type}` is a lowercase short form of the RequestType:
- NEW_PROJECT -> `new/`
- BUG_FIX -> `fix/`
- ENHANCEMENT -> `enhance/`
- REFACTOR -> `refactor/`
- COMPLEX_NEW -> `complex/`

Example: `new/rest-api`, `fix/auth-redirect`

### BR-11: ADR Naming

Architecture Decision Records follow the format `{seq}-{kebab-title}.md` where `{seq}` is a 3-digit zero-padded sequence number. Example: `001-use-postgresql.md`. ADRs live in `docs/spec/decisions/`.

### BR-12: Blueprint Naming

Blueprints follow the format `{seq}-{kebab-title}.md` where `{seq}` is a 2-digit zero-padded sequence number. Every blueprint must be registered in `docs/blueprints/INDEX.md`. Example: `01-mind-cli.md`.

### BR-13: Document Declaration Completeness

`mind.toml` must declare all documents that exist in `docs/`. An undeclared document on disk is an orphan and should be flagged by validation. A declared document missing from disk is flagged as missing.

### BR-14: Lock File Hash Consistency

When the lock file is in a clean state, every entry's `Hash` field must match the SHA-256 of the corresponding file's current content. A mismatch indicates drift and marks the entry as stale.

### BR-15: Staleness Propagation

Staleness propagates downstream only through the dependency graph. If document A depends on document B (via a `DependencyEdge`), and B's content changes, then A becomes stale. The reverse is not true: changing A does not make B stale. This follows the direction of the `informs` / `requires` / `validates` relationship types.

### BR-16: Workflow State Persistence

Workflow state is persisted in `docs/state/workflow.md` as a markdown file with an embedded JSON block. This enables both human readability and machine parsing. State must be written after each agent dispatch completes.

### BR-17: Iteration Completeness

A completed iteration must have all 5 expected artifacts present on disk:
1. `overview.md` — Scope, classification, and context
2. `changes.md` — What was changed and why
3. `test-summary.md` — Test results and coverage
4. `validation.md` — Review findings (MUST/SHOULD/COULD)
5. `retrospective.md` — Lessons learned

An iteration with fewer than 5 artifacts is classified as `incomplete`.

### BR-18: Zone Immutability Rules

Each zone has a defined mutability contract:
- **spec**: Stable. Evolves through deliberate updates only.
- **blueprints**: Stable. Active blueprints move to completed when done.
- **state**: Volatile. Changes frequently during active workflows.
- **iterations**: Immutable. Never modified after the iteration is marked complete.
- **knowledge**: Stable. Grows monotonically over the project lifetime.

### BR-19: Manifest Invariants

The `mind.toml` manifest declares structural invariants under `[manifest.invariants]`:
- `no-orphan-dependencies`: Every document ID referenced in a dependency edge must exist in the documents section.
- `no-circular-dependencies`: The dependency graph must be a DAG (directed acyclic graph).

---

## 5. Lifecycle State Machines

### 5.1 Document Lifecycle

Documents transition through statuses manually (via mind.toml `status` field) or by automatic stub detection.

```
                    ┌──────────────┐
                    │              │
          ┌────────▶│    draft     │◀────────┐
          │         │              │         │
          │         └──────┬───────┘         │
          │                │                 │
          │         (content added,          │
          │          status updated)         │
          │                │                 │
          │                ▼                 │
          │         ┌──────────────┐         │
          │         │              │         │
  (regressed)       │    active    │    (regressed)
          │         │              │         │
          │         └──────┬───────┘         │
          │                │                 │
          │         (finalized,              │
          │          status updated)         │
          │                │                 │
          │                ▼                 │
          │         ┌──────────────┐         │
          │         │              │         │
          └─────────│   complete   │─────────┘
                    │              │
                    └──────────────┘

  Note: "stub" is an orthogonal classification
  detected by content analysis, not a lifecycle
  state. A document can be status=draft AND
  IsStub=true simultaneously.
```

**Transitions**:
- `draft -> active`: Manual. Developer updates mind.toml status field when content is substantive.
- `active -> complete`: Manual. Developer updates mind.toml status field when content is finalized.
- `complete -> active` / `complete -> draft`: Manual regression when content needs rework.
- `stub` detection: Automatic. Applied by content analysis regardless of declared status.

---

### 5.2 Iteration Lifecycle

Iteration status is computed automatically from artifact presence.

```
    (mind create iteration / mind preflight)
                    │
                    ▼
            ┌──────────────┐
            │              │
            │  in_progress │
            │              │
            └──────┬───────┘
                   │
          (artifacts produced by agents)
                   │
         ┌─────────┴─────────┐
         │                   │
    (all 5 present)    (< 5 present,
         │              workflow ended)
         │                   │
         ▼                   ▼
  ┌──────────────┐   ┌──────────────┐
  │              │   │              │
  │   complete   │   │  incomplete  │
  │              │   │              │
  └──────────────┘   └──────────────┘
```

**Transitions**:
- `-> in_progress`: When the iteration directory is created with template files.
- `in_progress -> complete`: Automatic. All 5 expected artifacts exist on disk.
- `in_progress -> incomplete`: Automatic. Workflow ended but artifacts are missing.

---

### 5.3 Workflow Lifecycle

Workflow state tracks the full orchestration pipeline.

```
  ┌──────────────┐
  │              │
  │     idle     │◀──────────────────────────┐
  │              │                           │
  └──────┬───────┘                           │
         │                                   │
  (mind preflight / mind run)                │
         │                                   │
         ▼                                   │
  ┌──────────────┐                           │
  │              │                           │
  │   running    │◀─────┐                    │
  │              │      │                    │
  └──────┬───────┘      │                    │
         │              │                    │
  (agent completes)     │                    │
         │              │                    │
         ▼              │                    │
  ┌──────────────┐      │                    │
  │              │      │                    │
  │     gate     │──────┘                    │
  │              │  (gate passed,            │
  └──────┬───────┘   next agent)            │
         │                                   │
    ┌────┴────┐                              │
    │         │                              │
(all agents  (gate failed                   │
 complete)    after retries)                │
    │         │                              │
    ▼         ▼                              │
┌────────┐ ┌────────┐                       │
│        │ │        │                       │
│complete│ │ failed │                       │
│        │ │        │                       │
└────┬───┘ └────┬───┘                       │
     │          │                           │
     └──────────┴───────────────────────────┘
            (mind handoff / cleanup)
```

**Transitions**:
- `idle -> running`: Pre-flight succeeds, first agent dispatched.
- `running -> gate`: An agent completes, quality gate evaluates.
- `gate -> running`: Gate passes, next agent in chain dispatched.
- `gate -> running` (retry): Gate fails, retry count < 2, agent re-dispatched with feedback.
- `running -> completed`: Last agent finishes and final gate passes.
- `running -> failed`: Gate fails after max retries exhausted.
- `completed -> idle`: Post-workflow cleanup (mind handoff).
- `failed -> idle`: Post-workflow cleanup with documented failures.

---

### 5.4 Agent Dispatch Lifecycle

Individual agent dispatch within a workflow.

```
  ┌──────────────┐
  │              │
  │   pending    │
  │              │
  └──────┬───────┘
         │
  (CLI invokes claude)
         │
         ▼
  ┌──────────────┐
  │              │
  │  dispatched  │
  │              │
  └──────┬───────┘
         │
    ┌────┴────┐
    │         │
(success)  (error)
    │         │
    ▼         ▼
┌────────┐ ┌────────┐
│        │ │        │
│complete│ │ failed │
│        │ │        │
└────────┘ └───┬────┘
               │
          (retry count < 2?)
               │
          ┌────┴────┐
          │         │
        (yes)     (no)
          │         │
          ▼         ▼
   ┌──────────┐  ┌─────────────────┐
   │          │  │                 │
   │ retrying │  │ failed (final)  │
   │          │  │                 │
   └────┬─────┘  └─────────────────┘
        │
  (re-dispatched with
   gate feedback)
        │
        ▼
  ┌──────────────┐
  │              │
  │  dispatched  │  (back to dispatched)
  │              │
  └──────────────┘
```

**Transitions**:
- `pending -> dispatched`: Agent prompt assembled and sent to claude CLI.
- `dispatched -> completed`: Agent produces expected outputs successfully.
- `dispatched -> failed`: Agent errors or quality gate fails.
- `failed -> retrying`: Retry count < governance.max-retries (default 2).
- `retrying -> dispatched`: Re-dispatched with gate failure feedback appended to prompt.
- `failed -> failed (final)`: Retry count exhausted. Workflow proceeds with documented concerns.

---

## 6. Aggregate Boundaries

### 6.1 Project (Root Aggregate)

**Project** is the root aggregate. All operations require a valid project root as their entry point. The project aggregate owns:
- Config (mind.toml parse result)
- Brief (project brief analysis)
- Documents (filesystem scan results)
- Iterations (iteration directory scan)
- WorkflowState (current workflow)
- ProjectHealth (computed status)

**Invariant**: A Project always has a valid `Root` path pointing to a directory containing `.mind/`. Operations on a nil or invalid project return `ErrNotProject`.

### 6.2 Iteration (Secondary Aggregate)

**Iteration** is a self-contained aggregate representing one unit of work. It owns its 5 artifacts and can be validated independently of the broader project state. An iteration's status is derived entirely from its own artifact presence — no external data is needed.

**Invariant**: An iteration directory name always matches the format `{seq}-{TYPE}-{descriptor}`. The `Seq` is unique across all iterations in the project.

### 6.3 LockFile (Independent Aggregate)

**LockFile** is an independent aggregate that can be regenerated from the filesystem at any time. It holds no state that cannot be recomputed. Deleting the lock file and regenerating it produces an identical result if no files have changed.

**Invariant**: Every entry in the lock file references a document ID that exists in mind.toml. Orphan entries are pruned on regeneration.

### 6.4 ValidationReport (Value Object)

**ValidationReport** is a value object — immutable once computed. It has no identity beyond its contents. Two reports with identical check results are considered equal. Reports are never persisted; they are recomputed on each validation run.

### 6.5 QualityScore (Value Object)

**QualityScore** is a value object computed from a convergence analysis document. It is immutable once extracted. The `QualityEntry` persists a snapshot of the score in `quality-log.yml` for historical tracking.

---

## 7. Agent Chains

Agent chains define the ordered sequence of agents dispatched for each request type. The chain mapping is a core domain concept that drives workflow orchestration.

### 7.1 Chain Definitions

#### NEW_PROJECT

All agents required. Full specification-to-implementation pipeline.

```
analyst (opus) -> architect (opus) -> developer (sonnet) -> tester (sonnet) -> reviewer (opus)
```

#### BUG_FIX

Architect skipped. Bug fixes do not require architectural design.

```
analyst (opus) -> developer (sonnet) -> tester (sonnet) -> reviewer (opus)
```

#### ENHANCEMENT

Architect optional. May be needed for significant feature additions.

```
analyst (opus) -> architect? (opus) -> developer (sonnet) -> tester (sonnet) -> reviewer (opus)
```

#### REFACTOR

Tester skipped. Refactoring preserves behavior; existing tests validate correctness.

```
analyst (opus) -> developer (sonnet) -> reviewer (opus)
```

#### COMPLEX_NEW

Moderator prepares convergence analysis before the standard chain.

```
moderator (opus) -> analyst (opus) -> architect (opus) -> developer (sonnet) -> tester (sonnet) -> reviewer (opus)
```

### 7.2 Model Preferences

| Agent      | Preferred Model | Rationale                                               |
|------------|-----------------|---------------------------------------------------------|
| moderator  | opus            | Orchestrates multi-persona convergence; needs reasoning depth |
| analyst    | opus            | Requirements extraction demands nuanced understanding    |
| architect  | opus            | Design decisions require deep architectural reasoning    |
| developer  | sonnet          | Implementation is well-constrained by prior artifacts    |
| tester     | sonnet          | Test generation follows established patterns             |
| reviewer   | opus            | Evidence-based review requires judgment and synthesis    |

### 7.3 Chain Summary Table

| Request Type | Chain                                                    | Agents | Business Gate |
|-------------|----------------------------------------------------------|--------|---------------|
| NEW_PROJECT  | analyst -> architect -> developer -> tester -> reviewer  | 5      | Blocks        |
| BUG_FIX      | analyst -> developer -> tester -> reviewer               | 4      | Skipped       |
| ENHANCEMENT  | analyst -> architect? -> developer -> tester -> reviewer | 4-5    | Warns         |
| REFACTOR     | analyst -> developer -> reviewer                         | 3      | Skipped       |
| COMPLEX_NEW  | moderator -> analyst -> architect -> developer -> tester -> reviewer | 6 | Blocks |

### 7.4 Agent Definition File Locations

```
.mind/agents/analyst.md
.mind/agents/architect.md
.mind/agents/developer.md
.mind/agents/tester.md
.mind/agents/reviewer.md
.mind/agents/moderator.md
```

Each agent definition is a markdown file containing the agent's system prompt, responsibilities, expected inputs, expected outputs, and quality criteria. The CLI reads these files to assemble dispatch prompts in Model D (Full Orchestration).

---

## Exclusions

This document deliberately excludes:

- **Serialization formats**: How entities are encoded to/from TOML, JSON, or YAML is specified in the Data Contracts (see `docs/spec/api-contracts.md`).
- **Storage mechanics**: How entities are read from and written to the filesystem is an infrastructure concern (see `docs/spec/architecture.md` and `internal/repo/`).
- **Command behavior**: What each CLI command does with these entities is specified in the CLI Specification (see `docs/blueprints/01-mind-cli.md`).
- **AI integration**: How agents interact with these entities via MCP or orchestration is specified in the AI Workflow Bridge (see `docs/blueprints/02-ai-workflow-bridge.md`).

---

> **See also:**
> - [`domain/project.go`](../../domain/project.go) — Project, Config, and supporting types
> - [`domain/document.go`](../../domain/document.go) — Document, Brief, DocStatus, BriefGate
> - [`domain/zones.go`](../../domain/zones.go) — Zone enum and AllZones
> - [`domain/iteration.go`](../../domain/iteration.go) — Iteration, Artifact, RequestType, Classify, Slugify
> - [`domain/workflow.go`](../../domain/workflow.go) — WorkflowState, CompletedArtifact, DispatchEntry
> - [`domain/validation.go`](../../domain/validation.go) — ValidationReport, CheckResult, CheckLevel
> - [`domain/health.go`](../../domain/health.go) — ProjectHealth, ZoneHealth, Diagnostic, Suggestion
> - [`domain/errors.go`](../../domain/errors.go) — ErrNotProject, ErrBriefMissing, ErrGateFailed, ErrCommandFailed
> - [`docs/blueprints/01-mind-cli.md`](01-mind-cli.md) — CLI/TUI specification
> - [`docs/blueprints/03-architecture.md`](03-architecture.md) — System architecture with layered design
