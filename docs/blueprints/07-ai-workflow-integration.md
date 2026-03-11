# BP-07: AI Workflow Integration

> How do the 4 integration models work and connect to external AI systems?

**Status**: Active
**Date**: 2026-03-11
**Cross-references**: [BP-01](01-system-architecture.md) for architecture layers, [BP-02](02-domain-model.md) for entity definitions, [01-mind-cli](01-mind-cli.md) for command tree, [02-ai-workflow-bridge](02-ai-workflow-bridge.md) for original proposal

---

## 1. Integration Philosophy

### No AI in the CLI

The `mind` binary is strictly deterministic. Given the same filesystem state, every command produces the same output. The binary never calls AI APIs, never makes probabilistic decisions, never generates or interprets natural language. This is a hard constraint, not a guideline.

AI workflows run in Claude Code through the `/workflow`, `/discover`, and `/analyze` slash commands. These commands invoke the Mind Agent Framework's orchestrator, which coordinates persona-based agents (analyst, architect, developer, tester, reviewer) within Claude Code's runtime. The CLI has no knowledge of these agent definitions at runtime --- it reads them only as markdown files for validation and prompt assembly.

### What the CLI Provides

The CLI's role in the AI integration story is entirely supportive:

- **Context assembly** --- gather project-brief.md, requirements.md, architecture.md, recent iterations, and convergence docs into a structured package that agents can consume without redundant file reading
- **Validation** --- run the 17-check docs suite, 11-check refs suite, and 23-check convergence suite before, during, and after AI workflows
- **State tracking** --- read and write `docs/state/workflow.md` to persist workflow position, completed artifacts, and agent chain progress
- **Gate enforcement** --- execute deterministic gates (build, lint, test from `mind.toml [project.commands]`) and report structured pass/fail results
- **Agent dispatch coordination** --- in Model D only, invoke the `claude` CLI as an external process with assembled prompts and capture output

### The Model D Exception

Model D (`mind run`) invokes `claude` CLI as an external process. This is the sole point where the CLI interacts with an AI system. The interaction is strictly mechanical: the CLI builds a prompt string, pipes it to `claude --print`, and captures the stdout. The CLI does not interpret, evaluate, or modify the AI's output beyond parsing for artifact file paths. All intelligence remains in the agent definitions and Claude's model --- the CLI is a dispatcher, not a participant.

### Token Efficiency Principle

Each integration model progressively reduces redundant work that AI agents would otherwise perform. The cost of AI workflows is dominated by context tokens --- every file an agent reads consumes input tokens. By having the deterministic CLI handle classification, validation, iteration creation, and gate enforcement, the agent starts with structured results instead of raw files, and skips steps it would otherwise execute heuristically.

```
Model A: ~10-15% token savings    (orchestrator skips steps 1-4)
Model B: ~5-10% token savings     (agents get JSON instead of parsing markdown)
Model C: ~0% direct savings       (visibility, not optimization)
Model D: ~20-30% token savings    (each agent sees only its context, no orchestrator overhead)
```

---

## 2. Model A: Pre-Flight + Handoff

**Commands**: `mind preflight "<request>"`, `mind preflight --resume`, `mind handoff <iteration-id>`
**Complexity**: Low | **Value**: High

Model A wraps the deterministic steps that should happen before and after every AI workflow. It eliminates the 6-step manual checklist that users forget, ensures the project is in a valid state before tokens are spent, and automates post-workflow cleanup.

### Pre-Flight Flow

```
User runs: mind preflight "add JWT authentication to the REST API"

     +--------------------------------------------------------------+
     |  Step 1: CLASSIFY REQUEST                                     |
     |                                                               |
     |  Parse request string for keywords:                           |
     |    "create" / "new"        --> NEW_PROJECT                    |
     |    "fix" / "bug"           --> BUG_FIX                        |
     |    "add" / "improve" / "update" --> ENHANCEMENT               |
     |    "refactor" / "clean" / "reorganize" --> REFACTOR           |
     |    "complex" / "analyze"   --> COMPLEX_NEW                    |
     |                                                               |
     |  Result: ENHANCEMENT (keyword: "add")                         |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 2: BUSINESS CONTEXT GATE                                |
     |                                                               |
     |  Read docs/spec/project-brief.md                              |
     |  Evaluate based on RequestType:                               |
     |                                                               |
     |    NEW_PROJECT / COMPLEX_NEW:                                  |
     |      Brief missing or stub --> BLOCK (error, stop preflight)  |
     |                                                               |
     |    ENHANCEMENT:                                                |
     |      Brief missing --> WARN (continue with warning)           |
     |                                                               |
     |    BUG_FIX / REFACTOR:                                        |
     |      Brief check --> SKIP (not required)                      |
     |                                                               |
     |  Result: PASS (brief present, Vision + Deliverables + Scope)  |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 3: VALIDATE DOCUMENTATION                               |
     |                                                               |
     |  Run 17-check validation suite via ValidationService          |
     |  Report: pass / fail / warn counts                            |
     |  Non-blocking: warnings do not prevent preflight              |
     |                                                               |
     |  Result: 15 pass, 0 fail, 2 warnings                         |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 4: CREATE ITERATION                                     |
     |                                                               |
     |  Scan docs/iterations/ for highest sequence number            |
     |  Generate next: 007                                           |
     |  Create directory: docs/iterations/007-ENHANCEMENT-jwt-auth/  |
     |  Create 5 template files:                                     |
     |    overview.md, changes.md, test-summary.md,                  |
     |    validation.md, retrospective.md                            |
     |  Populate overview.md with:                                   |
     |    - classification (ENHANCEMENT)                             |
     |    - scope (from request string)                              |
     |    - agent chain (analyst -> developer -> tester -> reviewer) |
     |    - timestamp                                                |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 5: CREATE GIT BRANCH                                    |
     |                                                               |
     |  Branch name: {type}/{descriptor}                             |
     |  Following governance.branch-strategy = "type-descriptor"     |
     |                                                               |
     |  Execute: git checkout -b enhancement/jwt-auth                |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 6: ASSEMBLE CONTEXT PACKAGE                             |
     |                                                               |
     |  Read and aggregate:                                          |
     |    - docs/spec/project-brief.md                               |
     |    - docs/spec/requirements.md                                |
     |    - docs/spec/architecture.md                                |
     |    - docs/spec/domain-model.md                                |
     |    - Last 3 iteration overview.md files                       |
     |    - Relevant convergence docs (matched by topic keywords)    |
     |                                                               |
     |  Write assembled context to docs/state/workflow.md:           |
     |    - workflow_id, request_type, request_text                  |
     |    - iteration_path, branch_name                              |
     |    - agent_chain with status (pending/active/completed)       |
     |    - session_number, assembled_context summary                |
     +-------------------------------+------------------------------+
                                     |
     +-------------------------------v------------------------------+
     |  Step 7: GENERATE PROMPT                                      |
     |                                                               |
     |  Build orchestrator prompt:                                   |
     |    - Agent instructions from .mind/agents/orchestrator.md     |
     |    - Assembled project context (step 6)                       |
     |    - Iteration context (step 4)                               |
     |    - Conventions from .mind/conventions/                      |
     |    - User's original request                                  |
     |                                                               |
     |  Output options:                                              |
     |    - Copy to system clipboard (default if available)          |
     |    - Print to stdout (--print or no clipboard)                |
     |    - Save to file (--output <path>)                           |
     +--------------------------------------------------------------+
```

**Pre-Flight output**:

```
+-- Pre-Flight Complete -------------------------------------------+
|                                                                   |
|  Type: ENHANCEMENT                                                |
|  Chain: analyst -> developer -> tester -> reviewer                |
|  Branch: enhancement/jwt-auth                                     |
|  Iteration: docs/iterations/007-ENHANCEMENT-jwt-auth/             |
|  Brief: PASS (Vision, Deliverables, Scope present)                |
|  Docs: 15/17 pass (2 warnings, 0 blockers)                       |
|                                                                   |
|  Context package ready. Open Claude Code and run:                 |
|  /workflow "add JWT authentication to the REST API"               |
|                                                                   |
|  Or paste the generated prompt (copied to clipboard).             |
+-------------------------------------------------------------------+
```

### Resume Flow

`mind preflight --resume` reads `docs/state/workflow.md` and reports the current state of any in-progress workflow.

**If workflow in progress**:

```
+-- Resumable Workflow Found --------------------------------------+
|                                                                   |
|  Type: ENHANCEMENT                                                |
|  Iteration: 007-ENHANCEMENT-jwt-auth                              |
|  Last Agent: architect (completed)                                |
|  Remaining: developer -> tester -> reviewer                       |
|  Branch: enhancement/jwt-auth                                     |
|  Session: 1 of 2 (split after architect)                          |
|                                                                   |
|  Completed Artifacts:                                             |
|    [done] requirements.md (analyst)                               |
|    [done] architecture.md (architect)                             |
|                                                                   |
|  Resume in Claude Code:                                           |
|  /workflow --resume                                               |
+-------------------------------------------------------------------+
```

**If idle**:

```
No resumable workflow found.
Last completed: 006-ENHANCEMENT-add-caching (2026-03-08)
```

### Handoff Flow

`mind handoff <iteration-id>` runs post-workflow cleanup after an AI workflow completes.

```
mind handoff 007

Step 1: Validate iteration completeness
  Checking docs/iterations/007-ENHANCEMENT-jwt-auth/
    overview.md       [present]
    changes.md        [present]
    test-summary.md   [present]
    validation.md     [present]
    retrospective.md  [present]
  Result: 5/5 artifacts present

Step 2: Run deterministic checks
  Executing commands from mind.toml [project.commands]:
    go build -o mind .    [pass] (2.1s)
    golangci-lint run ./. [pass] (3.4s)
    go test ./...         [pass] (4.8s)
  Result: 3/3 pass

Step 3: Update docs/state/current.md
  Active Work: (cleared)
  Recent Changes: + 007-ENHANCEMENT-jwt-auth
  Next Priorities: (prompt user or leave unchanged)

Step 4: Clear workflow state
  docs/state/workflow.md -> idle

Step 5: Report branch status
  Branch: enhancement/jwt-auth
  Commits ahead of main: 3
  Suggestion: Create PR with 'gh pr create' or 'mind pr'
```

### What Model A Solves

| Problem | How Model A Solves It |
|---------|----------------------|
| User forgets to check brief before starting | Step 2 blocks or warns automatically |
| User forgets to validate docs | Step 3 runs validation suite |
| User creates iteration folder inconsistently | Step 4 uses GenerateService with auto-numbering |
| Classification happens inside AI (wastes tokens) | Step 1 classifies deterministically before AI starts |
| Git branch naming is inconsistent | Step 5 follows governance.branch-strategy |
| Post-workflow cleanup is forgotten | `mind handoff` automates all 5 cleanup steps |
| Resuming interrupted workflows is error-prone | `mind preflight --resume` reads persisted state |

**Token savings**: ~10-15%. The orchestrator agent can skip its own steps 1-4 (classify request, check brief, validate docs, create iteration) because the CLI already completed them. The assembled context package means the orchestrator reads one structured document instead of 4-6 raw files.

---

## 3. Model B: MCP Server

**Command**: `mind serve`
**Complexity**: Medium | **Value**: Very High

Model B exposes the CLI's capabilities as an MCP (Model Context Protocol) server. AI agents running in Claude Code call these tools via JSON-RPC instead of reading and parsing files manually. This is the highest-value integration because it improves every AI workflow without changing any user behavior.

### Transport and Protocol

**Transport**: stdio (JSON-RPC messages over stdin/stdout)
**Protocol**: MCP --- JSON-RPC 2.0 with tool registration, following the Model Context Protocol specification

The MCP server communicates via newline-delimited JSON-RPC 2.0 messages. stdin receives requests from the client (Claude Code), stdout sends responses. stderr is available for diagnostic logging but is never used for protocol messages.

### Server Lifecycle

```
1. User has .mcp.json in project root (or global config)
2. Claude Code starts, reads .mcp.json, spawns: mind serve
3. mind serve writes MCP initialization response to stdout:
   - protocol version
   - server name and version
   - supported capabilities (tools)

4. Client sends: tools/list
   Server responds: array of 16 tool definitions
     (name, description, input schema for each)

5. Client sends: tools/call { name: "mind_status", arguments: {} }
   Server executes: ProjectService.Health()
   Server responds: { content: [{ type: "text", text: "{...json...}" }] }

6. Repeat step 5 for any tool, any number of times

7. Client disconnects (EOF on stdin)
   Server exits cleanly
```

### Configuration

`.mcp.json` at the project root:

```json
{
  "mcpServers": {
    "mind": {
      "command": "mind",
      "args": ["serve"],
      "description": "Mind Framework project intelligence"
    }
  }
}
```

Claude Code discovers this file automatically and spawns the MCP server when a conversation starts. No additional setup is required from the user.

### 16 MCP Tools

#### Tool 1: mind_status

**Purpose**: Return a structured project health summary including documentation completeness per zone, active workflow state, iteration count, and warnings.

**When agents use it**: At the start of any workflow to understand the project's current state. The orchestrator calls this before making routing decisions. The reviewer calls this when gathering evidence for sign-off.

**What it replaces**: Manually reading `docs/state/current.md`, scanning `docs/spec/` for file presence, checking each zone for completeness, and parsing `docs/state/workflow.md` for workflow state. Without this tool, agents read 5-10 files and assemble a mental model; with it, they get one structured JSON response.

#### Tool 2: mind_doctor

**Purpose**: Run deep diagnostics across the entire project and return structured findings with severity levels and suggested fixes.

**When agents use it**: When the orchestrator detects problems during pre-flight and needs detailed information about what is wrong and how to fix it. Also used by the reviewer to verify project health post-workflow.

**What it replaces**: Running multiple validation scripts sequentially, cross-referencing results, and reasoning about which issues are blockers vs. warnings.

#### Tool 3: mind_check_brief

**Purpose**: Evaluate the business context gate. Parse `docs/spec/project-brief.md` and return a structured result indicating whether the brief exists, whether it is a stub, and which required sections (Vision, Key Deliverables, Scope) are present.

**When agents use it**: The orchestrator calls this at Step 1.5 (Business Context Gate) to decide whether the workflow can proceed for the given request type.

**What it replaces**: Reading `project-brief.md`, parsing its markdown structure, checking for specific headings, determining if the content is substantive or placeholder text. This is the most error-prone manual step --- agents frequently misclassify stubs as complete briefs.

#### Tool 4: mind_validate_docs

**Purpose**: Run the 17-check documentation validation suite and return structured results for each check.

**When agents use it**: Before starting any workflow (pre-flight validation). Also used by the reviewer to verify documentation completeness as part of sign-off.

**What it replaces**: Running `validate-docs.sh` or the equivalent manual checks --- verifying directory structure, required files, stub detection, naming conventions, and cross-references.

#### Tool 5: mind_validate_refs

**Purpose**: Run the 11-check cross-reference validation suite. Verify that links between documents resolve correctly, that referenced entities exist, and that dependency chains are intact.

**When agents use it**: After the architect modifies `architecture.md` or `domain-model.md`, to verify that cross-references to requirements, blueprints, and decisions still resolve.

**What it replaces**: Running `validate-integration.sh` or manually tracing links across the documentation tree.

#### Tool 6: mind_list_iterations

**Purpose**: Return a list of all iterations with their sequence number, type, descriptor, status, date, and artifact completeness.

**When agents use it**: The orchestrator lists iterations to determine the next sequence number and to understand recent project history. The reviewer lists iterations to find related past changes.

**What it replaces**: Scanning `docs/iterations/` directory, parsing folder names for type and descriptor, reading each `overview.md` to determine status, and checking artifact file presence.

#### Tool 7: mind_show_iteration

**Purpose**: Return detailed information about a single iteration, including its overview content, artifact list with presence/absence, and any validation findings.

**When agents use it**: Any agent that needs to understand the scope and artifacts of a specific iteration. The developer reads the current iteration to understand what has been planned. The reviewer reads the iteration to evaluate completeness.

**What it replaces**: Reading 5 files within an iteration directory (`overview.md`, `changes.md`, `test-summary.md`, `validation.md`, `retrospective.md`) and synthesizing them.

#### Tool 8: mind_read_state

**Purpose**: Read and parse `docs/state/workflow.md`, returning the current workflow state as structured data --- workflow ID, request type, current agent, agent chain with statuses, completed artifacts, session number.

**When agents use it**: Every agent checks state at the start of its turn to understand where it is in the pipeline. The orchestrator checks state to determine which agent to dispatch next.

**What it replaces**: Reading `docs/state/workflow.md` as raw markdown and parsing it heuristically. Different agents parse the same file differently, leading to state inconsistencies.

#### Tool 9: mind_update_state

**Purpose**: Write workflow state to `docs/state/workflow.md` with proper formatting and field validation. Accepts structured input (current agent, completed artifacts, status) and produces correctly formatted markdown.

**When agents use it**: After each agent completes its work, the orchestrator updates the state to reflect progress. On workflow completion, the state is transitioned to idle.

**What it replaces**: Manually formatting and writing markdown to `workflow.md`. This is the single most common source of state corruption --- agents write inconsistent markdown that other agents cannot parse reliably.

#### Tool 10: mind_create_iteration

**Purpose**: Create a new iteration directory with the correct sequence number, type-based naming, and 5 template files. Populate `overview.md` with the classification and scope.

**When agents use it**: The orchestrator calls this during pre-flight (Step 4) to create the iteration folder for the current workflow.

**What it replaces**: Manually scanning for the next sequence number, creating the directory, creating 5 files from templates, and populating the overview. Eliminates sequencing collisions and template inconsistencies.

#### Tool 11: mind_list_stubs

**Purpose**: Scan all documentation zones and return a list of documents that are stubs (contain only headings, comments, and placeholder text without substantive content).

**When agents use it**: The analyst uses this to identify what documentation needs to be created. The orchestrator uses it to warn about incomplete documentation before starting a workflow.

**What it replaces**: Reading every file in `docs/`, applying stub detection heuristics (line count, heading-only detection, placeholder pattern matching), and compiling a list.

#### Tool 12: mind_check_gate

**Purpose**: Run the deterministic gate --- execute build, lint, and test commands from `mind.toml [project.commands]` and return structured results for each command (pass/fail, duration, stdout, stderr).

**When agents use it**: The developer calls this before committing to verify that the code compiles, passes lint, and passes tests. The orchestrator calls this before dispatching the reviewer to ensure the deterministic gate passes.

**What it replaces**: Manually running build/lint/test commands, parsing their output, and determining pass/fail status. With structured JSON results, agents can programmatically check each gate component.

#### Tool 13: mind_log_quality

**Purpose**: Extract quality scores from a convergence analysis file and append them to `docs/knowledge/quality-log.yml`. Accepts a file path and optional topic/variant overrides.

**When agents use it**: After a convergence analysis completes (via `/analyze`), the orchestrator logs the quality scores for trend tracking.

**What it replaces**: Parsing convergence markdown for score tables, extracting 6 dimension scores and the overall score, formatting a YAML entry, and appending it to the log file.

#### Tool 14: mind_search_docs

**Purpose**: Full-text search across all files in `docs/`. Returns matching file paths, line numbers, and surrounding context for each match.

**When agents use it**: Any agent that needs to find where a concept, entity, or decision is documented. The analyst searches for existing requirements related to a new feature. The reviewer searches for architectural decisions relevant to a code change.

**What it replaces**: Manually reading files or using imprecise heuristics to locate documentation. Provides structured results instead of raw grep output.

#### Tool 15: mind_read_config

**Purpose**: Parse `mind.toml` and return the project configuration as structured JSON --- project metadata, stack configuration, commands, governance settings, document registry.

**When agents use it**: Any agent that needs project configuration. The orchestrator reads governance settings (max-retries, branch-strategy). The developer reads build commands. The reviewer reads project metadata for context.

**What it replaces**: Reading and parsing TOML manually. Agents frequently fail to parse TOML correctly; this tool returns pre-parsed JSON.

#### Tool 16: mind_suggest_next

**Purpose**: Analyze the current project state and suggest the next action. Considers: active workflow state, documentation completeness, iteration status, gate results, and recent changes.

**When agents use it**: The orchestrator calls this when it needs to determine the next step in a workflow. Any agent can call this for context-aware guidance.

**What it replaces**: The orchestrator's manual reasoning about "what should happen next" based on scattered state across multiple files. This tool encodes the decision logic deterministically.

### How Agents Use MCP Tools: Concrete Examples

**Orchestrator Step 1.5 --- Business Context Gate**:

```
Before MCP:
  1. Read docs/spec/project-brief.md          (400 tokens)
  2. Scan for "## Vision" heading             (heuristic, error-prone)
  3. Scan for "## Key Deliverables" heading   (heuristic, error-prone)
  4. Determine if content is substantive      (heuristic, error-prone)
  5. Decide: PASS / WARN / BLOCK             (~100 tokens reasoning)

After MCP:
  1. Call mind_check_brief
  2. Receive: { "status": "BRIEF_PRESENT",
                "sections": { "vision": true, "deliverables": true, "scope": true },
                "is_stub": false,
                "gate_result": "PASS" }
  3. Use structured result directly           (~20 tokens)

Savings: ~480 tokens, zero heuristic errors
```

**Orchestrator Step 4 --- Create Iteration**:

```
Before MCP:
  1. List docs/iterations/ contents           (variable tokens)
  2. Parse folder names for sequence numbers  (heuristic)
  3. Calculate next sequence number           (~50 tokens reasoning)
  4. Construct folder name                    (~30 tokens)
  5. Create directory                         (tool call)
  6. Create 5 template files                  (5 tool calls, ~200 tokens each)
  7. Populate overview.md                     (~300 tokens)

After MCP:
  1. Call mind_create_iteration {
       "type": "ENHANCEMENT",
       "descriptor": "jwt-auth",
       "request": "add JWT authentication to the REST API"
     }
  2. Receive: { "path": "docs/iterations/007-ENHANCEMENT-jwt-auth",
                "branch": "enhancement/jwt-auth",
                "overview": "docs/iterations/007-ENHANCEMENT-jwt-auth/overview.md",
                "artifacts": ["overview.md", "changes.md", "test-summary.md",
                              "validation.md", "retrospective.md"] }

Savings: ~1200 tokens, eliminates sequencing bugs and template inconsistencies
```

**Developer --- Before Committing**:

```
Before MCP:
  1. Run: go build -o mind .           (Bash tool call)
  2. Check exit code, parse output     (~100 tokens)
  3. Run: golangci-lint run ./...      (Bash tool call)
  4. Check exit code, parse output     (~100 tokens)
  5. Run: go test ./...                (Bash tool call)
  6. Check exit code, parse output     (~200 tokens)
  7. Decide: all pass or which failed  (~50 tokens reasoning)

After MCP:
  1. Call mind_check_gate
  2. Receive: {
       "passed": true,
       "commands": {
         "build": { "pass": true, "command": "go build -o mind .",
                     "duration_ms": 2100 },
         "lint":  { "pass": true, "command": "golangci-lint run ./...",
                     "duration_ms": 3400 },
         "test":  { "pass": true, "command": "go test ./...",
                     "duration_ms": 4800, "passed_count": 42, "failed_count": 0 }
       }
     }

Savings: ~450 tokens, one tool call instead of three
```

**Reviewer --- Evidence Gathering**:

```
1. Call mind_status
   -> Structured project health, docs completeness per zone, warnings

2. Call mind_list_iterations { "type": "ENHANCEMENT" }
   -> List of enhancement iterations with status, dates, artifact completeness

3. Call mind_show_iteration { "id": "007" }
   -> Full iteration details: overview, changes, test summary, validation

4. Call mind_check_gate
   -> Deterministic gate results for sign-off evidence

All four calls return structured JSON. The reviewer composes its
sign-off from structured data instead of reading ~15 raw files.
```

**Any Agent --- Context Awareness**:

```
Call mind_suggest_next
-> {
     "suggestion": "Developer should proceed with implementation.",
     "reason": "Analyst completed requirements extraction. Architecture is
                defined. Iteration 007 is active with overview populated.",
     "warnings": ["domain-model.md is a stub -- developer should create
                   entities as part of implementation"],
     "context": {
       "workflow_state": "active",
       "current_agent": "developer",
       "iteration": "007-ENHANCEMENT-jwt-auth",
       "recent_iterations": ["006-ENHANCEMENT-add-caching",
                             "005-BUG_FIX-fix-auth-redirect"]
     }
   }
```

### Why MCP Is the Highest-Value Integration

1. **Zero workflow change for users** --- Users continue using `/workflow`, `/discover`, and `/analyze` exactly as before. The agents become smarter tools, but the user's interface is unchanged. No new commands to learn, no new workflow to adopt.

2. **Structured over heuristic** --- Agents stop parsing markdown files and start receiving structured JSON. This eliminates an entire class of errors where agents misparse headings, miscategorize stubs, or miscount artifacts. A `mind_check_brief` response is unambiguous; parsing `project-brief.md` manually is not.

3. **State consistency** --- One tool (`mind_update_state`) writes workflow state using the same formatting logic every time. Without MCP, different agents write `workflow.md` with subtly different markdown formatting, and other agents fail to parse what was written. With MCP, the tool is the single source of truth for state serialization.

4. **Composable** --- New capabilities can be added as MCP tools without modifying any agent definitions. If a new validation suite is added (e.g., security checks), it becomes an MCP tool, and agents can call it without any instruction changes. The tool registry is open-ended.

5. **Platform-agnostic** --- MCP is an open protocol. It works with Claude Code today. If the project adds support for other AI platforms (Copilot, Cursor, Windsurf), they can connect to the same MCP server and access the same tools. The CLI does not need to know which client is connected.

### Architecture

```
+-----------------------------------------------------------+
|  Claude Code (client)                                      |
|                                                            |
|  Agent: "I need project status"                            |
|    |                                                       |
|    v                                                       |
|  MCP Client: tools/call { name: "mind_status" }           |
+------|----------------------------------------------------+
       | stdin (JSON-RPC)
       v
+-----------------------------------------------------------+
|  mind serve (MCP server process)                           |
|                                                            |
|  internal/mcp/server.go                                    |
|    - Parse JSON-RPC request                                |
|    - Route to tool handler by name                         |
|    - Tool handler calls service method                     |
|    - Serialize response as JSON-RPC                        |
|                                                            |
|  internal/mcp/tools.go                                     |
|    - 16 tool definitions (name, schema, handler function)  |
|    - Each handler is a thin adapter to a service method    |
|                                                            |
|  internal/service/*.go                                     |
|    - Same service implementations as CLI commands          |
|    - ProjectService.Health() called by both                |
|      'mind status' and 'mind_status' MCP tool              |
+------|----------------------------------------------------+
       | stdout (JSON-RPC)
       v
+-----------------------------------------------------------+
|  Claude Code (client)                                      |
|                                                            |
|  Agent receives: { "ok": true, "data": { ... } }          |
+-----------------------------------------------------------+
```

The MCP server reuses the exact same `internal/service/` packages as the CLI. There is no code duplication. `mind check docs` and `mind_validate_docs` call the same `ValidationService.ValidateDocs()` method. The only difference is the presentation layer: the CLI renders results as styled text or plain text; the MCP server serializes them as JSON.

---

## 4. Model C: Watch Mode

**Command**: `mind watch [--tui]`
**Complexity**: Medium | **Value**: High

Model C provides real-time visibility into what AI agents are doing by monitoring the filesystem for changes. While the user runs an AI workflow in Claude Code, `mind watch` runs in a separate terminal (or tmux pane) and shows live progress, background validation, and gate status.

### Filesystem Monitoring

Uses the `fsnotify` Go library for cross-platform filesystem event notification.

**Debounce strategy**: 300ms debounce window. When a file change event arrives, the watcher starts a 300ms timer. If additional events arrive within that window, the timer resets. When the timer fires, all accumulated events are batched into a single processing cycle. This handles the common case where AI agents write multiple files in rapid succession.

```
Events:                  |--A--|--B--|--C--|---------300ms---------|
                                                                   |
Processing:                                                   [A,B,C] handled
```

### Event-to-Action Mapping

```
+----------------------------------+-----------+-----------------------------------+
| File Pattern                     | Event     | Action                            |
+----------------------------------+-----------+-----------------------------------+
| docs/state/workflow.md           | Modified  | Parse workflow state, update       |
|                                  |           | dashboard with current agent,      |
|                                  |           | chain progress, session info       |
+----------------------------------+-----------+-----------------------------------+
| docs/iterations/*/overview.md    | Created   | New iteration detected. Log the   |
|                                  |           | type, descriptor, and agent chain. |
|                                  |           | Update iteration counter.          |
+----------------------------------+-----------+-----------------------------------+
| docs/iterations/*/changes.md     | Modified  | Developer produced changes. Run    |
|                                  |           | Micro-Gate B checks silently:      |
|                                  |           | verify changes.md exists and       |
|                                  |           | lists files, verify listed files   |
|                                  |           | exist on disk.                     |
+----------------------------------+-----------+-----------------------------------+
| docs/iterations/*/validation.md  | Modified  | Reviewer wrote findings. Parse     |
|                                  |           | for MUST/SHOULD/COULD categories.  |
|                                  |           | Display finding counts and any     |
|                                  |           | blockers (MUST items).             |
+----------------------------------+-----------+-----------------------------------+
| docs/spec/requirements.md        | Modified  | Analyst produced requirements.     |
|                                  |           | Run Micro-Gate A checks silently:  |
|                                  |           | verify GIVEN/WHEN/THEN present,    |
|                                  |           | FR-N identifiers traceable,        |
|                                  |           | scope boundary defined.            |
+----------------------------------+-----------+-----------------------------------+
| docs/spec/architecture.md        | Modified  | Architect modified design. Show    |
|                                  |           | change summary (sections modified).|
|                                  |           | Trigger reconciliation check to    |
|                                  |           | detect stale dependents.           |
+----------------------------------+-----------+-----------------------------------+
| docs/knowledge/*-convergence.md  | Created   | Convergence analysis complete.     |
|                                  |           | Run 23-check convergence           |
|                                  |           | validation. Show overall score     |
|                                  |           | and Gate 0 pass/fail.              |
+----------------------------------+-----------+-----------------------------------+
| docs/spec/project-brief.md       | Modified  | Brief changed. Re-run business     |
|                                  |           | context gate check. Update         |
|                                  |           | dashboard with new gate status.    |
+----------------------------------+-----------+-----------------------------------+
| src/**/*                         | Modified  | Source code changed. Run build     |
|                                  |           | and test commands in background.   |
|                                  |           | Show results when complete.        |
+----------------------------------+-----------+-----------------------------------+
```

### Background Command Execution

When source code changes trigger build/test commands, they run in goroutines with captured stdout/stderr. Results are piped to the activity log (stdout mode) or the TUI dashboard (TUI mode).

```go
// Simplified background execution model
go func() {
    result := processRunner.Run(ctx, "go build -o mind .")
    eventCh <- BuildResult{Command: "build", Result: result}
}()

go func() {
    result := processRunner.Run(ctx, "go test ./...")
    eventCh <- BuildResult{Command: "test", Result: result}
}()
```

Build and test commands run concurrently. If a new source change arrives while commands are running, the current run completes (not cancelled) and a new run is queued with a fresh debounce window.

### Watch Output Modes

**Plain mode** (`mind watch`):

```
[14:23:01] workflow.md updated: developer active (007-ENHANCEMENT-jwt-auth)
[14:23:15] src/middleware/jwt.go created
[14:23:32] src/routes/auth.go created
[14:24:01] changes.md updated: 4 files listed
[14:24:02] Micro-Gate B: checking...
[14:24:03] Micro-Gate B: PASS (changes.md present, 4 files verified on disk)
[14:24:45] Background: go build -o mind . [pass] (2.1s)
[14:25:12] Background: go test ./... [pass] 24 passed (4.8s)
[14:26:30] workflow.md updated: tester active
[14:28:15] test-summary.md updated: 8 test cases documented
[14:29:00] workflow.md updated: reviewer active
[14:30:45] validation.md updated: 0 MUST, 2 SHOULD, 1 COULD
[14:31:00] workflow.md updated: idle (workflow complete)
```

Events are timestamped and printed to stdout as a log stream. Suitable for piping, logging, or running in a background terminal.

**TUI mode** (`mind watch --tui`):

```
+-- mind watch --- mind-cli --- enhancement/jwt-auth -------------------------+
|                                                                              |
|  Workflow: ENHANCEMENT (007-jwt-auth)                                        |
|  Chain: analyst [done] -> developer [done] -> [tester] -> reviewer           |
|                                                                              |
|  +-- Live Activity -------------------------------------------------------+ |
|  |  14:26:30  Tester writing tests for JWT middleware                      | |
|  |  14:27:15  test-summary.md updated: 8 test cases                       | |
|  |  14:28:00  Tester writing tests for auth endpoints                      | |
|  |  14:28:30  test-summary.md updated: 12 test cases                      | |
|  |  14:28:45  Background: go test ./... [pass] 36 passed (5.2s)           | |
|  +------------------------------------------------------------------------+ |
|                                                                              |
|  +-- Gate Status ---------------------------------------------------------+ |
|  |  Build: [pass]          Lint: [pass]          Tests: 36/36 [pass]      | |
|  |  Micro-Gate A: [pass]   Micro-Gate B: [pass]  Det. Gate: ready         | |
|  +------------------------------------------------------------------------+ |
|                                                                              |
|  Warnings: domain-model.md is still a stub                                   |
|                                                                              |
+------------------------------------------------------------------------------+
```

The TUI dashboard updates in real-time as events arrive. It uses Bubble Tea's event-driven architecture: filesystem events are converted to Bubble Tea messages and dispatched to the model's `Update` method. The view re-renders on every state change.

### Watch Flow Diagram

```
+-------------------+       +-------------------+       +-------------------+
|  fsnotify         |       |  Debounce         |       |  Event Router     |
|  watcher          | ----> |  (300ms window)   | ----> |                   |
|  (goroutine)      |       |  (goroutine)      |       |  Match file path  |
+-------------------+       +-------------------+       |  to action handler|
                                                        +--------+----------+
                                                                 |
                              +----------------------------------+
                              |              |                   |
                              v              v                   v
                    +---------+----+ +-------+------+ +---------+--------+
                    | State Handler| | Gate Handler | | Build Handler    |
                    | Parse state, | | Run micro-   | | Run build/test   |
                    | update view  | | gate checks  | | in background    |
                    +---------+----+ +-------+------+ +---------+--------+
                              |              |                   |
                              v              v                   v
                    +---------+------------------------------------------+
                    |  Output Channel                                     |
                    |  (plain: print to stdout / TUI: send Msg to model) |
                    +----------------------------------------------------+
```

---

## 5. Model D: Full Orchestration

**Commands**: `mind run "<request>"`, `mind run --resume`, `mind run --dry-run "<request>"`, `mind run --tui "<request>"`
**Complexity**: High | **Value**: Very High

Model D is the fully automated pipeline. The CLI drives the entire workflow: classifies the request, creates the iteration, dispatches each agent as a separate `claude` CLI process, runs quality gates between agents, handles retries, and performs post-workflow cleanup. The user types one command and the entire agent chain executes.

### Pipeline

```
mind run "add JWT authentication to the REST API"

+================================================================+
|  STEP 1: PRE-FLIGHT (same as Model A)                          |
|                                                                 |
|  1. Classify: ENHANCEMENT                                       |
|  2. Business context gate: PASS                                  |
|  3. Validate docs: 15/17 pass                                   |
|  4. Create iteration: 007-ENHANCEMENT-jwt-auth                   |
|  5. Create branch: enhancement/jwt-auth                          |
|  6. Assemble context package                                     |
+================================+================================+
                                 |
+================================v================================+
|  STEP 2: DISPATCH ANALYST                                       |
|                                                                 |
|  a. Build prompt via PromptBuilder:                              |
|     - .mind/agents/analyst.md (agent instructions)               |
|     - project-brief.md, requirements.md, architecture.md         |
|     - iteration overview.md                                      |
|     - .mind/conventions/shared.md, documentation.md              |
|     - Task: "Analyze and extract requirements for: add JWT..."   |
|                                                                 |
|  b. Dispatch:                                                    |
|     echo "$prompt" | claude --print --model opus \               |
|       --allowedTools Read,Write,Edit,Grep,Glob                   |
|                                                                 |
|  c. Capture output                                               |
|                                                                 |
|  d. Run Micro-Gate A (deterministic):                            |
|     - requirements.md updated?                                   |
|     - GIVEN/WHEN/THEN criteria present?                          |
|     - FR-N identifiers traceable?                                |
|     - Scope boundary defined?                                    |
|     - Success metrics present?                                   |
|                                                                 |
|  e. Gate passed --> proceed to next agent                        |
|     Gate failed, retries < 2 --> re-dispatch with gate feedback  |
|     Gate failed, retries exhausted --> proceed with concerns      |
+================================+================================+
                                 |
+================================v================================+
|  STEP 3: DISPATCH ARCHITECT                                     |
|                                                                 |
|  Same pattern: build prompt -> dispatch -> capture -> gate       |
|  Prompt includes: analyst's requirements as prior agent output   |
+================================+================================+
                                 |
+================================v================================+
|  STEP 4: SESSION SPLIT DECISION                                 |
|                                                                 |
|  For NEW_PROJECT and COMPLEX_NEW:                                |
|    Auto-split after architect                                    |
|    Save state to docs/state/workflow.md                          |
|    Prompt user: "Requirements and architecture complete.          |
|                  Continue to implementation, or pause?"           |
|                                                                 |
|  For ENHANCEMENT, BUG_FIX, REFACTOR:                             |
|    No split, continue to developer                               |
+================================+================================+
                                 | (continue)
+================================v================================+
|  STEP 5: DISPATCH DEVELOPER                                     |
|                                                                 |
|  Prompt includes:                                                |
|    - requirements.md (analyst output)                            |
|    - architecture.md (architect output)                          |
|    - Iteration context                                           |
|  Dispatch: claude --print --model sonnet --allowedTools ...      |
|  Gate: Micro-Gate B (changes.md present, files exist)            |
+================================+================================+
                                 |
+================================v================================+
|  STEP 6: DISPATCH TESTER                                        |
|                                                                 |
|  Prompt includes:                                                |
|    - requirements.md (for acceptance criteria)                   |
|    - changes.md (developer output, file list)                    |
|  Dispatch: claude --print --model sonnet --allowedTools ...      |
+================================+================================+
                                 |
+================================v================================+
|  STEP 7: DETERMINISTIC GATE (no AI)                             |
|                                                                 |
|  Run commands from mind.toml [project.commands]:                 |
|    go build -o mind .       --> pass/fail                        |
|    golangci-lint run ./...  --> pass/fail                        |
|    go test ./...            --> pass/fail                        |
|                                                                 |
|  All pass --> proceed to reviewer                                |
|  Any fail, retries < 2 --> re-dispatch developer with failures   |
|  Any fail, retries exhausted --> proceed with documented concerns|
+================================+================================+
                                 |
+================================v================================+
|  STEP 8: DISPATCH REVIEWER                                      |
|                                                                 |
|  Prompt includes:                                                |
|    - Full iteration context (all artifacts)                      |
|    - requirements.md (acceptance criteria to verify)             |
|    - changes.md + git diff (what was implemented)                |
|    - Gate results (deterministic gate outcome)                   |
|  Dispatch: claude --print --model opus --allowedTools ...        |
|  Parse validation.md for sign-off decision                       |
+================================+================================+
                                 |
+================================v================================+
|  STEP 9: POST-WORKFLOW (same as mind handoff)                   |
|                                                                 |
|  1. Validate iteration completeness (5/5 artifacts)              |
|  2. Run deterministic checks (build/lint/test)                   |
|  3. Update docs/state/current.md                                 |
|  4. Clear workflow state (workflow.md -> idle)                    |
|  5. Report branch status, suggest PR creation                    |
+================================================================+
```

### Agent Prompt Assembly (PromptBuilder)

The PromptBuilder constructs each agent's prompt from 8 structured sections. The order is fixed. Each section is included only when applicable.

```
+--------------------------------------------------------------+
|  PROMPT STRUCTURE (assembled by PromptBuilder)                |
|                                                               |
|  Section 1: AGENT INSTRUCTIONS                                |
|  Source: .mind/agents/{agent}.md                              |
|  Always included. The agent's role, responsibilities,         |
|  output format, and quality criteria.                         |
|                                                               |
|  Section 2: PROJECT CONTEXT                                   |
|  Sources:                                                     |
|    - docs/spec/project-brief.md                               |
|    - docs/spec/requirements.md (if exists, non-stub)          |
|    - docs/spec/architecture.md (if exists, non-stub)          |
|    - docs/spec/domain-model.md (if exists, non-stub)          |
|  Included for all agents. Provides the project's identity,    |
|  requirements, architecture, and domain vocabulary.            |
|                                                               |
|  Section 3: ITERATION CONTEXT                                 |
|  Source: docs/iterations/{current}/overview.md                 |
|  Always included. The current iteration's classification,      |
|  scope, and objectives.                                       |
|                                                               |
|  Section 4: PRIOR AGENT OUTPUT                                |
|  Sources: Artifacts written by previous agents in this chain   |
|    - Analyst output: requirements.md                          |
|    - Architect output: architecture.md updates                |
|    - Developer output: changes.md, source files               |
|    - Tester output: test-summary.md                           |
|  Included only for agents after the first in the chain.       |
|  Each agent sees cumulative output from all prior agents.      |
|                                                               |
|  Section 5: CONVERGENCE CONTEXT                               |
|  Source: docs/knowledge/{topic}-convergence.md                 |
|  Included ONLY for COMPLEX_NEW request types. Provides the    |
|  multi-persona analysis that informed the project approach.    |
|                                                               |
|  Section 6: CONVENTIONS                                       |
|  Sources:                                                     |
|    - .mind/conventions/shared.md                              |
|    - .mind/conventions/documentation.md                       |
|  Always included. Project-wide coding and documentation        |
|  standards that all agents must follow.                        |
|                                                               |
|  Section 7: GATE FAILURE (retry only)                         |
|  Source: Previous gate result for this agent                   |
|  Included ONLY when re-dispatching after a gate failure.       |
|  Contains the specific check failures, expected vs. actual,    |
|  and what needs to be fixed.                                  |
|                                                               |
|  Section 8: TASK                                              |
|  Source: User's original request with agent-specific framing   |
|  Always included. The concrete instruction for this agent:     |
|    Analyst: "Analyze the following request and produce         |
|             requirements: {request}"                           |
|    Architect: "Design the architecture for: {request}"        |
|    Developer: "Implement the following: {request}"            |
|    Tester: "Write tests for: {request}"                       |
|    Reviewer: "Review the implementation of: {request}"        |
+--------------------------------------------------------------+
```

### Abstraction Layer for AI CLI

Model D's dependency on the `claude` CLI is isolated behind an interface:

```go
// internal/orchestrate/executor.go

type AgentExecutor interface {
    Run(ctx context.Context, model string, prompt string, allowedTools []string) (output []byte, err error)
}
```

**Default implementation** --- `ClaudeExecutor`:

```go
type ClaudeExecutor struct{}

func (e *ClaudeExecutor) Run(ctx context.Context, model string, prompt string, allowedTools []string) ([]byte, error) {
    args := []string{"--print", "--model", model}
    if len(allowedTools) > 0 {
        args = append(args, "--allowedTools", strings.Join(allowedTools, ","))
    }

    cmd := exec.CommandContext(ctx, "claude", args...)
    cmd.Stdin = strings.NewReader(prompt)

    return cmd.Output()
}
```

**Test implementation** --- `MockExecutor`:

```go
type MockExecutor struct {
    Responses map[string][]byte  // model -> canned response
    Calls     []ExecutorCall     // recorded calls for assertions
}

func (e *MockExecutor) Run(ctx context.Context, model string, prompt string, allowedTools []string) ([]byte, error) {
    e.Calls = append(e.Calls, ExecutorCall{Model: model, Prompt: prompt, Tools: allowedTools})
    if resp, ok := e.Responses[model]; ok {
        return resp, nil
    }
    return nil, fmt.Errorf("no mock response for model %s", model)
}
```

This abstraction ensures:

- Model D is not coupled to `claude` CLI specifics. If the CLI flags change, only `ClaudeExecutor` changes.
- Tests run without any AI. The full pipeline can be tested with `MockExecutor` returning canned responses for each agent.
- Future flexibility: the interface could be implemented by an API client (Anthropic API), a different AI CLI, or a batch processing system.

### Retry Logic

```
  Agent dispatched
       |
       v
  Agent completes
       |
       v
  Quality gate runs (deterministic, no AI)
       |
   +---+---+
   |       |
  PASS    FAIL
   |       |
   |       v
   |   retries < max? (governance.max-retries, default 2)
   |       |
   |   +---+---+
   |   |       |
   |  YES      NO
   |   |       |
   |   v       v
   |  Re-dispatch agent       Proceed with
   |  with gate feedback      documented concerns
   |  appended as Section 7   in iteration overview
   |       |                       |
   +-------+-----------+----------+
                       |
                       v
                  Next agent in chain
```

**On gate failure (retries remaining)**: The same agent is re-dispatched with Section 7 (GATE FAILURE) appended to its prompt. This section contains the specific failures --- which checks failed, what was expected, what was found. The agent can see exactly what to fix.

**On gate failure (retries exhausted)**: The pipeline continues to the next agent. The failures are documented in the iteration's `overview.md` under a "Known Concerns" section. The reviewer will see these concerns and factor them into the sign-off decision.

**On agent error (process failure)**: If the `claude` process exits with a non-zero code or times out, retry once. If the second attempt also fails, abort the pipeline and save state for `mind run --resume`.

**Timeouts**: Configurable per-agent, default 10 minutes. The `context.Context` passed to `AgentExecutor.Run` has a deadline set from the timeout value.

### Session Splitting

For `NEW_PROJECT` and `COMPLEX_NEW` request types, the pipeline automatically splits after the architect agent. This is because:

1. Requirements + architecture form a natural review checkpoint
2. The implementation phase (developer + tester + reviewer) may take significantly longer
3. The user should review the design before committing to implementation

```
mind run "create: REST API with JWT auth"

  [Pre-flight] -> [Analyst] -> [Architect]
                                    |
                              AUTO-SPLIT
                                    |
  Save state to workflow.md         |
  Display: "Session 1 complete.     |
            Requirements and        |
            architecture ready.     |
            Continue? [y/n/later]"  |
                                    |
  If "y": [Developer] -> [Tester] -> [Det. Gate] -> [Reviewer] -> [Handoff]
  If "n" or "later": Exit. Resume with 'mind run --resume'
```

For `ENHANCEMENT`, `BUG_FIX`, and `REFACTOR`, no split occurs --- the pipeline runs end-to-end.

### Dry Run

`mind run --dry-run "<request>"` executes classification, gate checks, and iteration creation, but skips all AI dispatches. It shows what would happen:

```
$ mind run --dry-run "add JWT authentication to the REST API"

Dry Run -- no AI calls will be made

  Classification: ENHANCEMENT
  Chain: analyst -> developer -> tester -> reviewer
  Business Context Gate: PASS (brief present with Vision, Deliverables, Scope)
  Documentation Validation: 15/17 pass (2 warnings, 0 blockers)
  Iteration: would create 007-ENHANCEMENT-jwt-auth/
  Branch: would create enhancement/jwt-auth

  Agent Dispatch Plan:
    1. Analyst   (opus)   -- analyze request, extract requirements
    2. Developer (sonnet) -- implement JWT authentication
    3. Tester    (sonnet) -- write tests for auth endpoints and middleware
    4. Reviewer  (opus)   -- verify requirements met, review implementation

  Quality Gates:
    Micro-Gate A -- after analyst (requirements structure)
    Micro-Gate B -- after developer (changes.md and files)
    Deterministic Gate -- before reviewer
      Commands: go build -o mind ., golangci-lint run ./..., go test ./...

  Estimated tokens: ~40,000-60,000 (4 agent dispatches)
  Estimated cost: ~$0.80-1.50
```

The dry run creates nothing on disk. It is purely informational. Useful for verifying that classification is correct, the agent chain is appropriate, and gates will not block before committing to a full run.

### TUI Mode

`mind run --tui "<request>"` shows the pipeline in a full-screen TUI dashboard:

```
+-- mind run --- ENHANCEMENT: jwt-auth ------------------------------------+
|                                                                          |
|  Pipeline Progress                                                       |
|  -----------------                                                       |
|  [done] Pre-flight     classify, gate, iteration, branch     0.8s       |
|  [done] Analyst        requirements extracted (8 FR, 6 AC)    2m 14s    |
|  [done] Micro-Gate A   5/5 checks pass                        0.2s      |
|  [>>>]  Developer      implementing... (src/middleware/jwt.go) 1m 32s   |
|  [ ]    Micro-Gate B   waiting                                           |
|  [ ]    Tester         waiting                                           |
|  [ ]    Det. Gate      waiting                                           |
|  [ ]    Reviewer       waiting                                           |
|                                                                          |
|  +-- Developer Output (live) ------------------------------------------+ |
|  |  Creating src/middleware/jwt.go -- Token validation middleware       | |
|  |  Creating src/routes/auth.go -- JWT authentication endpoints        | |
|  |  Creating src/models/claims.go -- JWT claims struct                 | |
|  |  Updating src/routes/users.go -- Adding auth middleware             | |
|  |  ...                                                                | |
|  +---------------------------------------------------------------------+ |
|                                                                          |
|  Tokens: 42,318 in / 18,204 out    Cost: ~$1.82    Elapsed: 6m 47s     |
|                                                                          |
|  [p]ause  [s]kip agent  [a]bort  [d]etails  [l]og                      |
+--------------------------------------------------------------------------+
```

The TUI provides interactive controls:

- **p (pause)**: Save state and exit after the current agent completes
- **s (skip)**: Skip the current agent and proceed to the next (with documented skip)
- **a (abort)**: Save state and exit immediately
- **d (details)**: Show full prompt and output for the selected agent
- **l (log)**: Show raw log of all events

### Cost Tracking

Model D tracks token usage and estimated cost per agent dispatch:

- **Input tokens**: Estimated heuristically at ~4 characters per token from the prompt size
- **Output tokens**: Estimated from the captured stdout size
- **Cost**: Calculated from model-specific per-token pricing

Cost data is stored in the iteration's `overview.md` under a "Cost Summary" section:

```markdown
## Cost Summary

| Agent | Model | Input Tokens | Output Tokens | Est. Cost | Duration |
|-------|-------|-------------|---------------|-----------|----------|
| Analyst | opus | 12,400 | 4,200 | $0.52 | 2m 14s |
| Developer | sonnet | 18,300 | 8,900 | $0.68 | 3m 45s |
| Tester | sonnet | 14,200 | 6,100 | $0.42 | 2m 30s |
| Reviewer | opus | 22,100 | 3,800 | $0.71 | 1m 55s |
| **Total** | | **67,000** | **23,000** | **$2.33** | **10m 24s** |
```

---

## 6. Shared Concerns

These concerns span multiple integration models. They use the same domain types and service implementations regardless of which model is active.

### Iteration State Transitions

All models create and manage iterations using the same `IterationRepo` interface and `GenerateService`:

- **Model A**: `mind preflight` creates the iteration in Step 4. `mind handoff` validates completeness.
- **Model B**: `mind_create_iteration` MCP tool calls `GenerateService.CreateIteration()`.
- **Model C**: Watch mode detects iteration creation and tracks artifact presence.
- **Model D**: Pre-flight step creates the iteration. Pipeline tracks which artifacts each agent produces.

The `Iteration` domain type and its status transitions are identical across all models:

```
Created --> In Progress --> Complete
                |
                +--> Abandoned (if workflow aborted)
```

### Quality Log Entries

Gates in Models A and D log quality metrics via `QualityService`:

- **Model A**: `mind handoff` logs the final gate result (build/lint/test pass/fail).
- **Model B**: `mind_log_quality` MCP tool calls `QualityService.LogEntry()`.
- **Model D**: Each gate execution (micro-gates and deterministic gate) logs results.

All entries use the same `QualityEntry` domain type and append to `docs/knowledge/quality-log.yml`.

### Token Estimation

A shared heuristic estimates token counts for pre-dispatch cost estimation:

```
Estimated tokens = len(text) / 4
```

This ~4 characters per token ratio is a rough approximation for English text. It is used for:

- Model A: Estimating the size of the assembled context package
- Model D: Pre-dispatch cost estimation and the dry-run cost projection
- Cost tracking: When exact token counts are unavailable from the `claude` CLI output

The heuristic is never used for billing or precise measurement. It provides order-of-magnitude estimates.

### Workflow State Persistence

All models read and write `docs/state/workflow.md` via the `StateRepo` interface:

```
+------------------+     +------------------+     +------------------+
| Model A          |     | Model B          |     | Model D          |
| preflight writes |     | mind_update_state|     | Pipeline writes  |
| initial state    |     | tool writes/reads|     | after each agent |
+--------+---------+     +--------+---------+     +--------+---------+
         |                        |                        |
         +------------------------+------------------------+
                                  |
                    +-------------v--------------+
                    | StateRepo (interface)       |
                    | ReadState() -> WorkflowState|
                    | WriteState(WorkflowState)   |
                    +-------------+--------------+
                                  |
                    +-------------v--------------+
                    | FSStateRepo (implementation)|
                    | Reads/writes workflow.md    |
                    | with consistent formatting  |
                    +----------------------------+
```

The `WorkflowState` domain type captures:

- `WorkflowID`: Unique identifier for this workflow run
- `RequestType`: Classification (NEW_PROJECT, ENHANCEMENT, BUG_FIX, REFACTOR, COMPLEX_NEW)
- `RequestText`: Original user request
- `IterationPath`: Path to the iteration directory
- `BranchName`: Git branch for this workflow
- `AgentChain`: Ordered list of agents with per-agent status (pending, active, completed, failed)
- `CurrentAgent`: Which agent is currently active (or empty if idle)
- `SessionNumber`: For split workflows, which session this is
- `CompletedArtifacts`: Map of agent name to list of artifact paths produced

---

## 7. Abstraction Layer

The `claude` CLI dependency in Model D is the only external AI dependency in the entire system. It is abstracted behind a single interface to ensure testability, portability, and future flexibility.

```go
// internal/orchestrate/executor.go

// AgentExecutor dispatches a prompt to an AI model and returns the output.
// The implementation decides how to communicate with the AI system.
type AgentExecutor interface {
    Run(ctx context.Context, model string, prompt string, allowedTools []string) ([]byte, error)
}

// ClaudeExecutor dispatches via the claude CLI (default for production).
// Invokes: echo "$prompt" | claude --print --model {model} --allowedTools {tools}
type ClaudeExecutor struct{}

// MockExecutor returns canned responses (for testing).
// Records all calls for assertion in tests.
type MockExecutor struct {
    Responses map[string][]byte
    Calls     []ExecutorCall
}
```

### Why This Abstraction Exists

1. **Model D is not coupled to `claude` CLI specifics**. If Anthropic changes the CLI flags, argument format, or output structure, only `ClaudeExecutor` needs updating. The pipeline, prompt builder, gate logic, and retry logic are unaffected.

2. **Tests run without AI**. The full orchestration pipeline --- classify, create iteration, dispatch agents, run gates, retry, handoff --- can be integration-tested using `MockExecutor` with deterministic canned responses. No API calls, no costs, no flakiness.

3. **Future portability**. The interface could be implemented by:
   - An API client calling the Anthropic Messages API directly
   - A different AI CLI (e.g., a hypothetical `openai` or `gemini` CLI)
   - A batch processing system that queues prompts for async execution
   - A local model runner for offline development

The interface intentionally mirrors the lowest common denominator: model selection, prompt text, and tool restrictions. Implementations handle transport, authentication, and protocol specifics.

---

## 8. Incremental Delivery

### Why This Order: Model A, Then B, Then C, Then D

```
+-------------------------------------------------------------------+
|  Model A: Pre-Flight + Handoff                                     |
|  Complexity: Low                                                    |
|  Dependencies: CLI foundation (BP-01) only                          |
|  Value: High --- immediate workflow improvement                     |
|                                                                    |
|  Delivers: preflight, handoff commands                              |
|  No external dependencies, no protocol implementation,              |
|  no process management. Pure filesystem + git operations.           |
+================================+==================================+
                                 |
+================================v==================================+
|  Model B: MCP Server                          *** INFLECTION ***   |
|  Complexity: Medium                                                 |
|  Dependencies: Model A + MCP protocol (JSON-RPC 2.0)               |
|  Value: Very High --- highest value-to-effort ratio                 |
|                                                                    |
|  Delivers: mind serve, 16 MCP tools                                 |
|  Requires: JSON-RPC parser/serializer, tool registry,              |
|  stdio transport. All tool handlers reuse existing services.        |
|                                                                    |
|  THIS IS THE INFLECTION POINT:                                      |
|  Once the 16 tools exist as service methods with JSON in/out,       |
|  Models C and D become simpler:                                     |
|    - Model C handlers call the same service methods                 |
|    - Model D gate checks call the same service methods              |
|    - No new business logic needed, only new orchestration           |
+================================+==================================+
                                 |
+================================v==================================+
|  Model C: Watch Mode                                                |
|  Complexity: Medium                                                 |
|  Dependencies: Model B services + fsnotify                          |
|  Value: High --- real-time visibility during workflows              |
|                                                                    |
|  Delivers: mind watch, mind watch --tui                              |
|  Requires: fsnotify integration, debounce logic, event routing,    |
|  background command execution. All handlers call existing           |
|  service methods (validation, gate checks, state parsing).          |
|  fsnotify is already needed for reconciliation (BP-01).             |
+================================+==================================+
                                 |
+================================v==================================+
|  Model D: Full Orchestration                                        |
|  Complexity: High                                                   |
|  Dependencies: Models A+B+C + claude CLI on PATH                    |
|  Value: Very High --- fully automated workflows                     |
|                                                                    |
|  Delivers: mind run, mind run --resume, mind run --dry-run,         |
|            mind run --tui                                            |
|  Requires: PromptBuilder, AgentExecutor, pipeline state machine,   |
|  retry logic, session splitting, cost tracking. Depends on          |
|  claude CLI stability for non-interactive dispatch.                 |
+===================================================================+
```

### Each Model Is Independently Valuable

The project can ship after any model and still provide meaningful value:

| Ship After | What Users Get |
|------------|---------------|
| **Model A only** | Pre-flight checklist automation, post-workflow cleanup, classification, context assembly. Users still run `/workflow` in Claude Code manually. |
| **Models A+B** | Everything above, plus AI agents get structured tools instead of parsing files. Workflows become more reliable without any user workflow change. |
| **Models A+B+C** | Everything above, plus real-time visibility into what agents are doing. Users see build failures, gate results, and progress in a live dashboard. |
| **Models A+B+C+D** | Everything above, plus fully automated pipeline. One command drives the entire workflow from classification to PR-ready branch. |

### Risk Mitigation

**Model A** has no external dependencies and no risk of API instability. It is pure filesystem and git operations. Ship first.

**Model B** depends on the MCP protocol specification, which is stable and open. The stdio transport is simple (newline-delimited JSON). The risk is low.

**Model C** depends on `fsnotify`, a mature and well-tested Go library. The debounce and event routing logic is straightforward. Medium risk from cross-platform filesystem notification quirks.

**Model D** carries the highest risk because it depends on the `claude` CLI's non-interactive mode being stable, reliable, and feature-complete. The `--print` flag, `--allowedTools` flag, and stdin prompt piping must work correctly. If the CLI changes its interface, `ClaudeExecutor` must be updated. The abstraction layer (Section 7) mitigates this --- the pipeline logic is decoupled from CLI specifics. Start Model D last to give the `claude` CLI time to stabilize.

---

> **See also:**
> - [BP-01](01-system-architecture.md) --- Architecture layers, design principles, package structure
> - [BP-02](02-domain-model.md) --- Domain entities (WorkflowState, Iteration, AgentChain, ValidationReport)
> - [01-mind-cli](01-mind-cli.md) --- Original CLI proposal with command tree and TUI design
> - [02-ai-workflow-bridge](02-ai-workflow-bridge.md) --- Original proposal for the 4 integration models
> - [03-architecture](03-architecture.md) --- Software architecture with service interfaces and DI patterns
