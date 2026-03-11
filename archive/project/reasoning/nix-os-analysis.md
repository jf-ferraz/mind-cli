Summarized conversation historyThis is a genuinely brilliant intuition. The NixOS parallel is deeper than it might seem at first glance. Let me think through this properly.

## The Core Insight: NixOS → Mind Framework

What makes NixOS revolutionary isn't just "one config file." It's three properties working together:

1. **Declarative** — you describe *what should exist*, not *how to build it*
2. **Content-addressable** — every artifact is identified by its content, not its location
3. **Reproducible** — same manifest → same system state, always

Applied to the Mind Framework, this means: **the manifest doesn't just *describe* the project — it *defines* it.** The orchestrator reads the manifest, computes the delta between declared state and actual state, and dispatches agents to close the gap.

---

## Format Analysis

Let me evaluate each format against the five constraints that matter for this file:

| Constraint | Why It Matters |
|---|---|
| **LLM-parseable** | Claude reads this file to understand the entire system |
| **Human-editable** | Developers maintain and review it |
| **Machine-parseable** | Scripts (scaffold.sh, install.sh, orchestrator) consume it |
| **Comment-friendly** | Governance decisions need context — "why" not just "what" |
| **Diff-friendly** | Every change is version-controlled, diffs must be meaningful |

### TOML — Recommended

```toml
# Mind Framework Manifest
# Declares the complete system state — agents compute the delta

[manifest]
schema = "mind/v2"
generation = 12

[project]
name = "inventory-api"
domain = "manufacturing"
type = "backend"          # backend | frontend | fullstack | library

[documents.spec.requirements]
path = "docs/spec/requirements.md"
agent = "analyst"
status = "active"           # draft | active | stale | archived
depends-on = ["spec.project-brief"]
tags = ["core"]
```

**Pros**: Clean section headers, native comments, no indentation traps, excellent for medium nesting (which is what we need), every major language has a parser, diffs are line-oriented and clean, Cargo/Rust ecosystem proved it works for build systems.

**Cons**: Deep nesting (4+ levels) becomes `[a.b.c.d]` header soup — but we can flatten with dotted keys. Arrays-of-tables syntax (`[[array]]`) is less intuitive. Less familiar to LLMs than JSON/YAML in training data (though Claude handles it well).

### YAML — Strong Alternative

```yaml
# Mind Framework Manifest
manifest:
  schema: mind/v2
  generation: 12

project:
  name: inventory-api
  domain: manufacturing

documents:
  spec:
    requirements:
      path: docs/spec/requirements.md
      agent: analyst
      depends-on: [spec.project-brief]
```

**Pros**: Most human-readable for nested structures, native comments, anchors/aliases enable DRY (`&default-doc` / `*default-doc`), multiline strings for descriptions, dominant in DevOps/CI — widely familiar, Claude parses YAML perfectly.

**Cons**: Whitespace-sensitivity is a real footgun (one misindent breaks the file), the Norway problem (`NO` → `false`, `3.10` → `3.1`), spec is enormous and complex (1.2 has 10 collection types), merge-conflict resolution in indented structures is painful.

### JSON — Universal but Insufficient

```json
{
  "manifest": { "schema": "mind/v2", "generation": 12 },
  "project": { "name": "inventory-api" }
}
```

**Pros**: Universal parsing, strict/unambiguous, JSON Schema for validation, every LLM generates valid JSON natively, zero ambiguity.

**Cons**: **No comments** — fatal for a governance file. Verbose (closing braces, mandatory quotes on keys). No trailing commas. Poor human editability. Git diffs are noisy.

**JSONC** (JSON with Comments) partially fixes this but has inconsistent tooling support.

### Lua — Programmable Power

```lua
return {
  manifest = { schema = "mind/v2", generation = 12 },
  project = { name = "inventory-api" },
  documents = {
    spec = {
      requirements = {
        path = "docs/spec/requirements.md",
        -- Analyst produces this from project-brief
        agent = "analyst",
        depends_on = { "spec.project-brief" },
      }
    }
  }
}
```

**Pros**: Turing-complete — can compute derived values, conditional inclusion, inheritance. NixOS's power comes precisely from this (Nix is a functional language, not just a config format). Lightweight runtime. Trailing commas. Comments.

**Cons**: Adds a runtime dependency (luajit/lua). LLMs parse it less reliably than pure data formats. Harder to validate statically. "It's code, not config" — governance files shouldn't have side effects.

### KDL — Modern Document Language

```kdl
manifest schema="mind/v2" generation=12

project name="inventory-api" domain="manufacturing"

documents {
  spec {
    requirements path="docs/spec/requirements.md" agent="analyst" {
      depends-on "spec.project-brief"
      tags "core"
    }
  }
}
```

**Pros**: Designed specifically for document-like config (2021). Clean node-based syntax. Supports types. Both attributes and children. Comments.

**Cons**: Very new — limited tooling, limited LLM training data, no native parsers in many languages, adoption risk.

### Recommendation

**TOML as primary format**, with two supporting artifacts:

| File | Purpose | Format |
|---|---|---|
| `mind.toml` | Declarative manifest — what SHOULD exist | TOML |
| `mind.lock` | Computed state snapshot — what DOES exist | JSON (machine-generated, never hand-edited) |
| JSON Schema | Validation contract for both files | JSON Schema |

**Why TOML over YAML**: For a file that is the *single source of truth* for an entire system, the format must be **unambiguous**. YAML's whitespace sensitivity and type coercion create exactly the kind of subtle bugs that a governance file must never have. TOML's explicitness is a feature, not a bug. Cargo.toml proved that build-system manifests work beautifully in TOML.

**Why not Lua**: The NixOS analogy is tempting, but Nix's programmability is necessary because it manages 80,000+ packages with complex dependency resolution. Our manifest manages ~30-50 artifacts. Data > Code for this scale. If the framework ever needs computed derivations, we can add a `mind eval` command that processes TOML through a resolver — keeping the manifest pure data.

---

## Proposed Manifest Structure: `mind.toml`

```toml
# ╔══════════════════════════════════════════════════════════════╗
# ║  MIND MANIFEST — Single Source of Truth                      ║
# ║  This file declares the complete system state.               ║
# ║  Agents read it. Orchestrator enforces it. Git versions it.  ║
# ╚══════════════════════════════════════════════════════════════╝

# ─── MANIFEST METADATA ───────────────────────────────────────

[manifest]
schema    = "mind/v2.0"
generated = 2026-02-23T14:30:00Z   # Last generation timestamp
generation = 12                     # Monotonic counter (like NixOS generations)

# ─── PROJECT IDENTITY ────────────────────────────────────────

[project]
name        = "inventory-api"
description = "Manufacturing inventory management system"
domain      = "manufacturing"         # Business domain
type        = "backend"               # backend | frontend | fullstack | library | cli
created     = 2026-02-20

[project.stack]
# Declared, not enforced — agents adapt to the stack
language   = "csharp"
framework  = "dotnet-10"
database   = "postgresql"
# Optional extras
cache      = "redis"
messaging  = "rabbitmq"

# ─── FRAMEWORK CONFIGURATION ─────────────────────────────────

[framework]
version = "2.0.0"
path    = ".claude"                   # Where framework is installed

[framework.orchestration]
autonomy-level   = 3                  # 1-5 per ASDLC research
max-retries      = 2                  # Per quality gate
session-strategy = "split"            # single | split | manual
# split: orchestrator recommends session breaks for long workflows

[framework.quality-gates]
deterministic = true                  # Enable build/lint/type-check gates
micro-gates   = true                  # Enable per-agent verification (Gate A, Gate B)

# ─── AGENT REGISTRY ──────────────────────────────────────────
# Every agent the system can invoke. Core agents ship with the framework.
# Specialists are project-specific.

[agents.orchestrator]
id     = "agent:orchestrator"
path   = ".claude/agents/orchestrator.md"
role   = "dispatch"                   # dispatch | analysis | design | implementation | verification | exploration
loads  = "always"                     # always | on-demand | conditional

[agents.analyst]
id     = "agent:analyst"
path   = ".claude/agents/analyst.md"
role   = "analysis"
loads  = "on-demand"
produces = [
  "doc:spec/requirements",
  "doc:spec/domain-model",
  "doc:spec/issue-analysis",
  "doc:spec/requirements-delta",
]

[agents.architect]
id     = "agent:architect"
path   = ".claude/agents/architect.md"
role   = "design"
loads  = "conditional"                # Only for NEW_PROJECT or structural ENHANCEMENT
produces = [
  "doc:spec/architecture",
  "doc:spec/api-contracts",
]

[agents.developer]
id     = "agent:developer"
path   = ".claude/agents/developer.md"
role   = "implementation"
loads  = "on-demand"
produces = ["doc:iteration/changes"]

[agents.tester]
id     = "agent:tester"
path   = ".claude/agents/tester.md"
role   = "verification"
loads  = "on-demand"

[agents.reviewer]
id     = "agent:reviewer"
path   = ".claude/agents/reviewer.md"
role   = "verification"
loads  = "on-demand"
produces = ["doc:iteration/validation"]

[agents.discovery]
id     = "agent:discovery"
path   = ".claude/agents/discovery.md"
role   = "exploration"
loads  = "on-demand"
produces = ["doc:spec/project-brief"]

# ─── SPECIALIST AGENTS (project-specific) ────────────────────

[agents.database-specialist]
id       = "specialist:database"
path     = ".claude/specialists/database-specialist.md"
role     = "analysis"
loads    = "conditional"
triggers = ["database", "migration", "schema", "query", "index", "entity"]
inserts-after = "analyst"

# ─── CONVENTIONS ──────────────────────────────────────────────

[conventions]
# Active conventions for this project. Framework ships all; project activates what it needs.
active = [
  "code-quality",
  "documentation",
  "git-discipline",
  "severity",
  "temporal",
  "backend-patterns",    # Optional — activated because project.type = "backend"
]

# ─── SKILLS ───────────────────────────────────────────────────

[skills]
available = ["planning", "debugging", "refactoring", "quality-review"]
# Skills are always on-demand. Listed here for manifest completeness.

# ─── WORKFLOW DEFINITIONS ─────────────────────────────────────

[workflows.new-project]
chain = ["analyst", "architect", "developer", "tester", "reviewer"]
session-split-after = "architect"     # Recommend new session after planning phase
gates = { after-analyst = "micro-a", after-developer = "micro-b", before-reviewer = "deterministic" }

[workflows.bug-fix]
chain = ["analyst", "developer", "tester", "reviewer"]
gates = { after-developer = "micro-b", before-reviewer = "deterministic" }

[workflows.enhancement]
chain = ["analyst", "developer", "tester", "reviewer"]  # Architect added dynamically if structural
gates = { after-analyst = "micro-a", after-developer = "micro-b", before-reviewer = "deterministic" }

[workflows.refactor]
chain = ["analyst", "developer", "reviewer"]
gates = { after-developer = "micro-b", before-reviewer = "deterministic" }

# ─── DOCUMENT REGISTRY ────────────────────────────────────────
# The knowledge graph. Every artifact the system knows about.
# Canonical IDs use the mind:// URI scheme.
# Physical paths are resolved relative to project root.

# ── Zone 1: Specifications (stable) ──

[documents.spec.project-brief]
id       = "doc:spec/project-brief"
path     = "docs/spec/project-brief.md"
zone     = "spec"
status   = "active"                   # draft | active | stale | archived | superseded
owner    = "agent:discovery"          # Which agent creates/maintains this
tags     = ["core", "planning"]
hash     = "a3f2b1c8"                # Truncated SHA-256 of content (updated by mind.lock)

[documents.spec.requirements]
id          = "doc:spec/requirements"
path        = "docs/spec/requirements.md"
zone        = "spec"
status      = "active"
owner       = "agent:analyst"
depends-on  = ["doc:spec/project-brief"]        # Upstream dependencies
consumed-by = ["agent:architect", "agent:developer", "agent:tester"]
tags        = ["core", "planning"]

[documents.spec.domain-model]
id          = "doc:spec/domain-model"
path        = "docs/spec/domain-model.md"
zone        = "spec"
status      = "active"
owner       = "agent:analyst"          # Created by analyst, refined by architect
depends-on  = ["doc:spec/project-brief", "doc:spec/requirements"]
consumed-by = ["agent:architect", "agent:developer", "agent:tester"]
tags        = ["core", "domain", "data-model"]

[documents.spec.architecture]
id          = "doc:spec/architecture"
path        = "docs/spec/architecture.md"
zone        = "spec"
status      = "active"
owner       = "agent:architect"
depends-on  = ["doc:spec/requirements", "doc:spec/domain-model"]
consumed-by = ["agent:developer"]
tags        = ["core", "design"]

[documents.spec.api-contracts]
id          = "doc:spec/api-contracts"
path        = "docs/spec/api-contracts.md"
zone        = "spec"
status      = "draft"
owner       = "agent:architect"
depends-on  = ["doc:spec/domain-model", "doc:spec/architecture"]
consumed-by = ["agent:developer", "agent:tester"]
tags        = ["api", "contracts"]

# ── Zone 2: Runtime State (volatile) ──

[documents.state.current]
id     = "doc:state/current"
path   = "docs/state/current.md"
zone   = "state"
status = "active"
owner  = "agent:orchestrator"
tags   = ["runtime"]

[documents.state.workflow]
id     = "doc:state/workflow"
path   = "docs/state/workflow.md"
zone   = "state"
status = "active"
owner  = "agent:orchestrator"
tags   = ["runtime", "resume"]

[documents.state.backlog]
id     = "doc:state/backlog"
path   = "docs/state/backlog.md"
zone   = "state"
status = "active"
owner  = "agent:analyst"
tags   = ["planning", "runtime"]

# ── Zone 3: Iterations (append-only history) ──

[documents.iterations.003-bugfix-auth]
id          = "doc:iteration/003-bugfix-auth"
path        = "docs/iterations/003-bugfix-auth/"
zone        = "iteration"
status      = "archived"
type        = "bug-fix"
branch      = "bugfix/auth-500-error"
implements  = ["doc:spec/requirements#FR-12"]     # Traceability!
created     = 2026-02-22
artifacts   = ["overview.md", "changes.md", "validation.md"]

# ── Zone 4: Domain Knowledge (stable reference) ──

[documents.knowledge.glossary]
id     = "doc:knowledge/glossary"
path   = "docs/knowledge/glossary.md"
zone   = "knowledge"
status = "active"
owner  = "agent:discovery"
tags   = ["domain", "reference"]

[documents.knowledge.integrations]
id     = "doc:knowledge/integrations"
path   = "docs/knowledge/integrations.md"
zone   = "knowledge"
status = "draft"
tags   = ["external", "reference"]

# ─── DEPENDENCY GRAPH (computed edges) ─────────────────────────
# Explicit cross-cutting relationships that don't fit in per-document depends-on.

[[dependencies]]
from   = "doc:spec/requirements"
to     = "doc:iteration/003-bugfix-auth"
type   = "implemented-by"              # implemented-by | superseded-by | derived-from | validates

[[dependencies]]
from   = "doc:spec/domain-model"
to     = "doc:spec/api-contracts"
type   = "derived-from"

# ─── GOVERNANCE ────────────────────────────────────────────────

[governance]
autonomy-level = 3
review-policy  = "evidence-based"       # evidence-based | score-based | checklist
commit-policy  = "conventional"         # conventional | free-form | squash-only

[governance.gates]
# Named gate definitions referenced in workflow configs
[governance.gates.micro-a]
type        = "probabilistic"
description = "Requirements quality check"
checks      = ["acceptance-criteria-present", "scope-boundaries-defined", "no-ambiguous-terms"]

[governance.gates.micro-b]
type        = "probabilistic"
description = "Implementation completeness check"
checks      = ["changes-md-exists", "all-files-exist", "scope-adherence"]

[governance.gates.deterministic]
type     = "deterministic"
commands = ["build", "lint", "typecheck", "test"]
# Actual commands resolved from project.stack at runtime

[[governance.decisions]]
id       = "ADR-001"
title    = "Use PostgreSQL over MongoDB"
date     = 2026-02-20
status   = "accepted"                   # proposed | accepted | deprecated | superseded
document = "docs/spec/decisions/001-postgresql.md"

[[governance.decisions]]
id       = "ADR-002"
title    = "Event-driven inventory updates"
date     = 2026-02-21
status   = "accepted"
document = "docs/spec/decisions/002-event-driven.md"

# ─── GENERATIONS (state history) ──────────────────────────────
# Like NixOS generations — each meaningful state transition is recorded.
# Only the last N are kept in the manifest. Full history in git.

[[generations]]
number    = 12
timestamp = 2026-02-23T14:30:00Z
event     = "iteration-complete"
detail    = "003-bugfix-auth completed and reviewed"
hash      = "manifest:e4a7c2f1"         # Hash of manifest at this point

[[generations]]
number    = 11
timestamp = 2026-02-22T09:15:00Z
event     = "iteration-start"
detail    = "003-bugfix-auth created"
hash      = "manifest:b2d1a8e3"
```

---

## The Lock File: `mind.lock`

Inspired by `flake.lock` / `package-lock.json` — a machine-generated snapshot of the actual state:

```json
{
  "lockVersion": 1,
  "generatedAt": "2026-02-23T14:30:00Z",
  "generation": 12,
  "documents": {
    "doc:spec/project-brief": {
      "path": "docs/spec/project-brief.md",
      "exists": true,
      "hash": "sha256:a3f2b1c8e4d7f9a2b5c8e1d4f7a0b3c6",
      "size": 2847,
      "lastModified": "2026-02-20T10:00:00Z",
      "stale": false
    },
    "doc:spec/requirements": {
      "path": "docs/spec/requirements.md",
      "exists": true,
      "hash": "sha256:d4e7f0a3b6c9d2e5f8a1b4c7d0e3f6a9",
      "size": 5231,
      "lastModified": "2026-02-21T16:30:00Z",
      "stale": false,
      "upstreamHashes": {
        "doc:spec/project-brief": "sha256:a3f2b1c8e4d7f9a2b5c8e1d4f7a0b3c6"
      }
    },
    "doc:spec/api-contracts": {
      "path": "docs/spec/api-contracts.md",
      "exists": false,
      "stale": true,
      "reason": "declared but not yet created"
    }
  },
  "integrity": "sha256:full-manifest-hash-here"
}
```

**Key behaviors:**

1. **Staleness detection**: If `doc:spec/project-brief` hash changes, everything in its `consumed-by` chain is marked `stale: true` in the lock file. The orchestrator knows to re-run those agents.

2. **Existence verification**: Declared documents that don't exist on disk are flagged — the orchestrator knows what needs to be built.

3. **Upstream hash tracking**: Each document records the hashes of its dependencies *at the time it was last generated*. If upstream changes, downstream is stale.

4. **Never hand-edited**: Generated by `mind lock` (a command that scans the filesystem against the manifest).

---

## The Brilliant Indexing Idea: Canonical URI Scheme + Reactive Dependency Graph

### The `doc:` URI Scheme

Every artifact in the system gets a **canonical identifier** that is stable, semantic, and path-independent:

```
doc:{zone}/{name}              → Document artifact
doc:{zone}/{name}#{fragment}   → Specific section/requirement within a document
agent:{name}                   → Agent definition
specialist:{name}              → Specialist agent
convention:{name}              → Convention
skill:{name}                   → Skill
gate:{name}                    → Quality gate
workflow:{name}                → Workflow definition
adr:{number}                   → Architecture Decision Record
iteration:{descriptor}         → Full iteration bundle
```

**Why this matters:**

1. **Path independence** — If `docs/spec/requirements.md` moves to `docs/specifications/requirements.md`, only the manifest's `path` field changes. Every agent, every cross-reference, every dependency edge still uses `doc:spec/requirements`. Zero broken links.

2. **Fragment addressing** — `doc:spec/requirements#FR-12` points to a specific requirement. An iteration can declare `implements = ["doc:spec/requirements#FR-12", "doc:spec/requirements#FR-15"]`. This creates **requirement traceability** — from business need to implementation to test to validation.

3. **Cross-type references** — `agent:analyst` produces `doc:spec/requirements`. `gate:micro-a` validates `doc:spec/requirements`. `iteration:003-bugfix-auth` implements `doc:spec/requirements#FR-12`. The entire system is a navigable graph.

4. **Namespace scoping** — Agents can query the manifest by namespace:
   - "Give me everything in `doc:spec/*`" → all specifications
   - "What does `agent:analyst` produce?" → follow `produces` edges
   - "What's stale?" → check lock file hashes
   - "What implements FR-12?" → reverse-traverse `implements` edges

### The Reactive Dependency Graph

This is where NixOS's derivation model truly shines when adapted:

```
doc:spec/project-brief ──produces──▶ agent:discovery
         │
         ▼ depends-on
doc:spec/requirements ──produces──▶ agent:analyst
         │
    ┌────┴────┐
    ▼         ▼
doc:spec/   doc:spec/
domain-     architecture ──produces──▶ agent:architect
model           │
    │      ┌────┴────┐
    ▼      ▼         ▼
doc:spec/  (developer   (tester
api-        reads)       reads)
contracts
```

**When `doc:spec/project-brief` changes:**
1. Lock file detects hash mismatch
2. `doc:spec/requirements` is marked stale (depends on changed upstream)
3. Transitively: `doc:spec/domain-model`, `doc:spec/architecture`, `doc:spec/api-contracts` all stale
4. Orchestrator knows: re-run analyst → architect → developer pipeline
5. But `doc:knowledge/glossary` is NOT stale (no dependency edge)

**This is NixOS's `nixos-rebuild` for knowledge artifacts.** You change one input, and the system computes the minimum set of "rebuilds" (agent invocations) needed.

### Practical Implementation: The `mind` Commands

```bash
# Compute lock file from manifest + filesystem
mind lock

# Show what's stale (like `nix build --dry-run`)
mind status
# Output:
#   STALE  doc:spec/requirements     (upstream doc:spec/project-brief changed)
#   STALE  doc:spec/architecture     (transitive: requirements stale)
#   MISSING doc:spec/api-contracts   (declared but not created)
#   OK     doc:knowledge/glossary
#   OK     doc:state/current

# Show dependency graph
mind graph
# Output: ASCII/Mermaid dependency visualization

# Show what agents would run to resolve all staleness
mind plan
# Output:
#   1. agent:analyst → rebuild doc:spec/requirements, doc:spec/domain-model
#   2. agent:architect → rebuild doc:spec/architecture, doc:spec/api-contracts
#   3. agent:developer → implementation (stale specs affect code)

# Bump generation after significant state change
mind commit "iteration 003 complete"
# Appends to [[generations]], updates hashes, increments counter
```

These don't need to be real CLI tools — they can be **orchestrator behaviors**. When the orchestrator reads `mind.toml` + `mind.lock`, it effectively runs `mind status` + `mind plan` internally to decide what to do.

---

## Additional Ideas for Single Source of Truth

### 1. Template Inheritance

The manifest can declare document templates, ensuring every artifact follows the canonical structure:

```toml
[templates]
[templates.iteration-overview]
path = ".claude/templates/iteration-overview.md"
used-by = ["doc:iteration/*"]

[templates.domain-model]
path = ".claude/templates/domain-model.md"
used-by = ["doc:spec/domain-model"]
```

When the analyst creates `domain-model.md`, it reads the template from the manifest — not from a hardcoded path.

### 2. Project Profiles (like NixOS modules)

```toml
# Instead of configuring everything manually, activate a profile
[profiles]
active = ["backend-api", "event-driven"]

# Profile definitions (shipped with framework)
# backend-api activates: backend-patterns convention, database specialist, domain-model template
# event-driven activates: messaging conventions, async patterns
```

This is like NixOS modules — `services.nginx.enable = true` brings in all nginx configuration. `profiles.backend-api = true` activates all backend-relevant framework components.

### 3. Hooks (Declarative, not Imperative)

```toml
[hooks]
# Declarative triggers, not imperative scripts
pre-iteration  = ["verify-clean-tree", "create-branch"]
post-analyst   = ["gate:micro-a"]
post-developer = ["gate:micro-b", "gate:deterministic"]
post-reviewer  = ["commit-artifacts", "generate-pr-summary"]
```

### 4. Metrics & Observability

```toml
[metrics]
total-iterations = 12
avg-retries-per-iteration = 0.4
most-active-zones = ["spec", "iteration"]
last-stale-detection = 2026-02-22T09:15:00Z
```

The manifest becomes a living dashboard of project health.

### 5. Manifest Self-Validation

```toml
[manifest.invariants]
# Rules the manifest itself must satisfy
every-document-has-owner = true
every-iteration-has-validation = true
no-orphan-dependencies = true          # Every depends-on target must exist
no-circular-dependencies = true
```

Like NixOS's `assertions` — the system validates its own configuration before applying it.

---

## Design Recommendations for Scalability & Maintainability

1. **Start minimal, grow organically** — The manifest for a new project should be ~30 lines (project + core documents). It grows as the project grows. Don't require all sections upfront.

2. **Generated sections vs authored sections** — Clearly separate what humans write (project config, decisions, profiles) from what tools generate (hashes, generations, lock file). Use comments to mark the boundary.

3. **The lock file absorbs complexity** — Keep `mind.toml` clean and declarative. All computed state (hashes, staleness, existence checks) lives in `mind.lock`. This is the separation that makes `flake.nix` + `flake.lock` work.

4. **Agents read, orchestrator writes** — Only the orchestrator modifies `mind.toml` (new iterations, generation bumps). Agents read it for context. This prevents concurrent modification issues.

5. **Git is the version control** — Don't duplicate git's job. The manifest tracks current state + recent generations. Full history is in `git log -- mind.toml`. The lock file is .gitignore'd (like `node_modules`) or committed (like `package-lock.json`) — I'd recommend **committed**, so any session can verify state without running `mind lock`.

6. **Fragment IDs must be stable** — `doc:spec/requirements#FR-12` only works if requirement IDs are stable across edits. The analyst must assign deterministic IDs (FR-{sequential}) that persist across requirement updates. This is a convention, not a tool feature.

---

## Summary: The Vision

```
┌─────────────────────────────────────────────────────────┐
│                     mind.toml                           │
│              (Declarative Manifest)                      │
│                                                         │
│  "This is what the project SHOULD look like"            │
│                                                         │
│  project identity + agent registry + document graph     │
│  + workflows + governance + generations                 │
└───────────────────────┬─────────────────────────────────┘
                        │
                   ┌────▼────┐
                   │  mind   │
                   │  lock   │──── "This is what ACTUALLY exists"
                   └────┬────┘
                        │
                   ┌────▼────┐
                   │  delta  │──── "This is what needs to happen"
                   └────┬────┘
                        │
                   ┌────▼────────────┐
                   │  orchestrator   │
                   │  (NixOS-rebuild │──── Dispatches minimum agent set
                   │   for knowledge)│     to close the gap
                   └─────────────────┘
```

The canonical file isn't just a registry — it's a **declarative specification of the desired project state**. The orchestrator becomes a **reconciliation engine** that computes the delta between declared and actual, then dispatches the minimum set of agents to converge. Every artifact is addressable by canonical URI, every dependency is explicit, every state transition is a generation.

This is NixOS's core philosophy applied to software development knowledge: **declare what you want, let the system figure out how to get there.**