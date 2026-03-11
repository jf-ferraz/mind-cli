# BP-09: Use Cases & Walkthroughs

> How do I actually use this thing? A practical, plain-language guide with real terminal examples for every capability of mind-cli.

**Status**: Active
**Date**: 2026-03-11
**Depends on**: [04-cli-specification.md](04-cli-specification.md), [05-tui-specification.md](05-tui-specification.md), [07-ai-workflow-integration.md](07-ai-workflow-integration.md)

---

## 1. Getting Started (First 5 Minutes)

### 1.1 "I just installed mind-cli, now what?"

You have the `mind` binary on your PATH. You open a terminal, navigate to your project directory, and wonder what to do first. Here is exactly what happens.

**Step 1: Check the current state**

```bash
$ cd ~/dev/projects/my-awesome-api
$ mind status
```

```
Error: Not a Mind project (no .mind/ directory found)
Tip: Run 'mind init' to set up the Mind Framework
```

This error is normal. The Mind Framework needs a `.mind/` directory in your project root to work. Think of `.mind/` as the brain of your project -- it holds agent definitions, conventions, and configuration that power everything else.

**Step 2: Initialize the framework**

```bash
$ mind init --name my-awesome-api
```

```
Initializing Mind Framework in ~/dev/projects/my-awesome-api/

  Created .mind/                         (framework root)
  Created docs/spec/                     (5 stub documents)
  Created docs/blueprints/INDEX.md       (blueprint registry)
  Created docs/state/current.md          (project state)
  Created docs/state/workflow.md         (workflow state)
  Created docs/knowledge/                (knowledge zone)
  Created mind.toml                      (project manifest)
  Created .claude/CLAUDE.md              (Claude Code adapter)

✓ Framework initialized: my-awesome-api
  Run 'mind status' to see project health
  Run 'mind create brief' to fill the project brief
```

What just happened? The `mind init` command created two important things:

1. **`.mind/` directory** -- This is where agent definitions (analyst, architect, developer, tester, reviewer), conventions, and conversation configurations live. You generally do not edit files in here directly.

2. **`docs/` directory with 5 zones** -- This is your project's documentation structure:
   - **spec/** -- Stable specifications: project brief, requirements, architecture, domain model, API contracts. These change slowly and are the source of truth.
   - **blueprints/** -- Planning documents: system-level design artifacts, architecture decision records. The `INDEX.md` file serves as the registry.
   - **state/** -- Volatile state: what you are working on right now (`current.md`) and the status of any in-progress AI workflow (`workflow.md`).
   - **iterations/** -- History: each change (feature, bugfix, refactor) gets its own numbered folder with 5 tracking files.
   - **knowledge/** -- Reference material: glossary, spike reports, convergence analyses.

3. **`mind.toml`** -- The project manifest. It declares your project name, tech stack, build commands, and a registry of every document. Think of it as `package.json` for your project's documentation.

**Step 3: See where you stand**

```bash
$ mind status
```

```
╭─ Mind Framework ─── my-awesome-api ──────────────────────────╮
│                                                               │
│  Project: my-awesome-api          Framework: v2026-03-09      │
│  Root: ~/dev/projects/my-awesome-api/                         │
│                                                               │
│  Documentation Health                                         │
│  ───────────────────                                          │
│  spec/        ░░░░░░░░░░  0/5   all stubs                    │
│  blueprints/  ██████████  1/1   INDEX ✓                       │
│  state/       ██████████  2/2   current ✓  workflow ✓         │
│  iterations/  ░░░░░░░░░░  0/0   none                         │
│  knowledge/   ░░░░░░░░░░  0/1   glossary ✗                   │
│                                                               │
│  Workflow: idle (no active workflow)                           │
│                                                               │
│  Warnings                                                     │
│  ────────                                                     │
│  ⚠ All spec documents are stubs — start with the project brief│
│  ⚠ No iterations found                                        │
│  ⚠ glossary.md missing                                        │
│                                                               │
│  Tip: Run 'mind create brief' or use /discover in Claude Code │
╰───────────────────────────────────────────────────────────────╯
```

The health dashboard tells you three things at a glance:

- **Progress bars** show how many documents in each zone have real content versus being empty stubs. Right now everything in `spec/` is a stub -- the files exist but contain only template headings.
- **Workflow state** shows whether an AI workflow is currently running. "idle" means nothing is in progress.
- **Warnings** tell you what needs attention. The most important one right now is that all your spec documents are stubs.

The exit code from `mind status` is `1` because there are issues. When everything is healthy, the exit code is `0`. This matters if you use mind-cli in CI (more on that in Section 4.2).

---

### 1.2 "How do I fill in my project brief?"

The project brief is the single most important document in the Mind Framework. It tells AI agents what your project is about, what it should deliver, and what is in or out of scope. Without it, agents have to guess -- and they will guess wrong.

You have two paths to fill it in.

**Path A: Interactive CLI**

```bash
$ mind create brief
```

```
Creating project brief: docs/spec/project-brief.md

Vision — What does this project do? (1-3 sentences)
> A REST API for managing user accounts, authentication, and role-based access
  control for a SaaS platform.

Key Deliverables — What are the concrete outputs? (comma-separated)
> REST API endpoints, JWT authentication, role-based authorization,
  user management CRUD, audit logging

Scope — What is IN scope?
> User registration, login, password reset, role assignment (admin/editor/viewer),
  session management, API key generation, audit trail for security events

Scope — What is explicitly OUT of scope?
> Frontend UI, email service, billing integration, social login (OAuth providers)

Constraints — Any technical or business constraints?
> Go 1.23, PostgreSQL, must pass SOC 2 security review, single-region deployment

✓ Created docs/spec/project-brief.md
  Business context gate: PASS (Vision ✓, Key Deliverables ✓, Scope ✓)
```

The "business context gate" at the end is a check that confirms your brief has the three required sections (Vision, Key Deliverables, Scope) with actual content. This gate is used later when you start AI workflows -- for NEW_PROJECT and COMPLEX_NEW request types, a missing or stub brief blocks the workflow entirely.

**Path B: Claude Code /discover**

If you use Claude Code, you can run the `/discover` command instead. This launches an interactive exploration where Claude asks you questions about your project and writes the brief for you. The result is the same file (`docs/spec/project-brief.md`), but the content tends to be more thorough because the AI can ask follow-up questions.

After either path, run `mind status` again and you will see the `spec/` progress bar move from `0/5` to `1/5`, and the warning about stubs will update to reflect one fewer stub.

---

### 1.3 "How do I know if my project is healthy?"

Three commands give you progressively deeper insight into your project's health.

**Level 1: Quick dashboard**

```bash
$ mind status
```

This is the fastest check. It shows documentation completeness per zone, active workflow state, and top-level warnings. Run this at the start of every work session.

**Level 2: Deep diagnostics**

```bash
$ mind doctor
```

```
Running diagnostics...

✓ Framework installed (.mind/ present)
✓ Claude Code adapter installed (.claude/ present)
✗ Copilot adapter not found (.github/agents/ missing)
  → Run: mind init --with-github

✓ Documentation structure (17/17 checks pass)
✗ 2 stub documents found:
  → docs/spec/domain-model.md — needs entity definitions
  → docs/spec/api-contracts.md — needs endpoint definitions
  Fix: Fill these files or run /discover to generate context

✓ Framework cross-references (11/11 checks pass)
✓ Conversation configs valid (4/4 files)

⚠ Project brief present but missing "Key Deliverables" section
  → The business context gate will warn on ENHANCEMENT workflows
  → Fix: Add a ## Key Deliverables section to docs/spec/project-brief.md

✓ No stale workflow state
✓ All iterations have overview.md

Summary: 9 pass, 1 fail, 2 warnings
Run 'mind doctor --fix' to auto-fix resolvable issues
```

The doctor command goes deeper than `status`. It runs every validator, cross-references results, and tells you exactly what is wrong and how to fix it. Each line has a specific remediation suggestion.

When you see issues that the doctor can fix automatically (like missing directories or `.gitkeep` files), run:

```bash
$ mind doctor --fix
```

```
Running diagnostics with auto-fix...

✓ Framework installed
✓ Claude Code adapter installed
+ Created .github/agents/ directory
+ Synced 5 agent definitions to .github/agents/

✓ Documentation structure (17/17 pass)
⚠ 2 stubs remain (cannot auto-fix — these need human content)

✓ Cross-references (11/11 pass)
✓ Conversation configs valid

Summary: All resolvable issues fixed. 2 stubs remain (manual action needed).
```

The `--fix` flag handles structural problems (missing directories, missing adapter files, naming convention violations). It cannot write substantive documentation content for you -- that requires either you or an AI agent.

**Level 3: Full validation**

```bash
$ mind check all
```

```
╭─ Documentation (17 checks) ──────────────────────╮
│  Pass: 14  Fail: 0  Warn: 3                       │
│  ⚠ [12] No iterations found                       │
│  ⚠ [15] knowledge/glossary.md missing              │
│  ⚠ [16] 2 stub documents found                     │
╰────────────────────────────────────────────────────╯

╭─ Cross-References (11 checks) ────────────────────╮
│  Pass: 11  Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

╭─ Conversation Config (4 files) ───────────────────╮
│  Pass: 4   Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

Overall: 29/32 pass, 0 fail, 3 warnings
```

The `check all` command runs three validation suites:
- **docs** (17 checks) -- zone directories, required files, stub detection, naming conventions
- **refs** (11 checks) -- internal links, registry consistency, sequence numbering
- **config** (variable) -- conversation YAML files in `.mind/conversation/`

Warnings are informational. Failures cause exit code `1`. If you want warnings treated as failures (useful in CI), add `--strict`:

```bash
$ mind check all --strict
```

Now those 3 warnings become failures and the exit code is `1`.

---

## 2. Day-to-Day Usage

### 2.1 "I want to check my documentation before starting work"

This is the most common workflow. You sit down to start a new feature, and you want to make sure the project documentation is consistent before you begin.

```bash
# Step 1: Quick health check
$ mind status
```

```
╭─ Mind Framework ─── my-awesome-api ──────────────────────────╮
│                                                               │
│  Documentation Health                                         │
│  ───────────────────                                          │
│  spec/        ████████░░  4/5   brief ✓  reqs ✓  arch ✓     │
│  blueprints/  ██████████  3/3   INDEX ✓  + 2 blueprints      │
│  state/       █████░░░░░  1/2   current ✓  workflow ✗        │
│  iterations/  ██████████  6/6   all complete                  │
│  knowledge/   ██████░░░░  3/5   glossary ✗  2 spikes         │
│                                                               │
│  Warnings                                                     │
│  ────────                                                     │
│  ⚠ domain-model.md is a stub (needs content)                  │
│  ⚠ glossary.md missing                                        │
│  ⚠ No workflow state saved                                    │
│                                                               │
│  Tip: Run 'mind doctor' for detailed diagnostics              │
╰───────────────────────────────────────────────────────────────╯
```

You see two documentation issues (a stub and a missing file) and one state warning. The "No workflow state saved" warning is normal when no AI workflow is in progress.

```bash
# Step 2: Get details and fixes
$ mind doctor
```

```
Running diagnostics...

✓ Framework installed (.mind/ present)
✓ Claude Code adapter installed (.claude/ present)
✓ Copilot adapter installed (.github/agents/ present)

✓ Documentation structure (17/17 checks pass)
✗ 1 stub document found:
  → docs/spec/domain-model.md — needs entity definitions
  Fix: Fill with entity definitions, or let the architect agent handle it

✓ Framework cross-references (11/11 checks pass)
✓ Conversation configs valid (4/4 files)
✓ Project brief complete (Vision ✓, Deliverables ✓, Scope ✓)
✓ No stale workflow state

Summary: 10 pass, 0 fail, 1 warning
```

The doctor tells you the only real issue is a stub document. You can either fill it in yourself or let it be handled by an AI workflow.

```bash
# Step 3: Run full validation to be thorough
$ mind check all
```

```
╭─ Documentation (17 checks) ──────────────────────╮
│  Pass: 16  Fail: 0  Warn: 1                       │
│  ⚠ [16] 1 stub document found (domain-model.md)   │
╰────────────────────────────────────────────────────╯

╭─ Cross-References (11 checks) ────────────────────╮
│  Pass: 11  Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

╭─ Conversation Config (4 files) ───────────────────╮
│  Pass: 4   Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

Overall: 31/32 pass, 0 fail, 1 warning
```

Everything passes. The one warning is a known issue you can live with. You are good to start working.

---

### 2.2 "I need to create a new architecture decision"

Your team just decided to use PostgreSQL for persistence. You want to record that decision formally so future developers (and AI agents) understand why.

```bash
$ mind create adr "Use PostgreSQL for persistence"
```

```
✓ Created docs/spec/decisions/003-use-postgresql.md
```

That is it. The command did several things automatically:

1. Found the `docs/spec/decisions/` directory (and would have created it if it did not exist)
2. Scanned existing ADRs (001 and 002 already existed) and assigned the next number: 003
3. Slugified the title into a filename: `003-use-postgresql.md`
4. Filled the file with the ADR template

Now open the file and fill it in:

```bash
$ mind docs open 003-use-postgresql
```

This opens the file in your `$EDITOR`. The template looks like this:

```markdown
# ADR-003: Use PostgreSQL for persistence

**Status**: Proposed
**Date**: 2026-03-11
**Deciders**: (who was involved in the decision)

## Context

(What is the problem? Why do we need to decide something?)

## Decision

(What did we decide? Be specific.)

## Consequences

### Positive
- (what this makes easier)

### Negative
- (what this makes harder)

### Neutral
- (side effects that are neither good nor bad)

## Alternatives Considered

### (Alternative 1)
- Rejected because: (why)

### (Alternative 2)
- Rejected because: (why)
```

Fill in each section with your reasoning. The key fields are:

- **Status**: Change from "Proposed" to "Accepted" once the team agrees
- **Context**: The problem you are solving
- **Decision**: What you chose and why
- **Consequences**: What this makes easier and harder
- **Alternatives Considered**: What else you evaluated and why you rejected it

After saving, verify the ADR appears in your documentation:

```bash
$ mind docs list --zone spec
```

```
spec/
  project-brief.md          ✓ complete    2026-03-01     3.2 KB
  requirements.md           ✓ complete    2026-03-05     8.1 KB
  architecture.md           ✓ complete    2026-03-05     6.4 KB
  domain-model.md           ⚠ stub        2026-02-28     0.4 KB
  api-contracts.md          ✓ complete    2026-03-07     2.8 KB
  decisions/
    001-use-event-sourcing.md  ✓ complete    2026-03-02     1.8 KB
    002-jwt-auth.md            ✓ complete    2026-03-04     2.1 KB
    003-use-postgresql.md      ✓ complete    2026-03-11     1.4 KB

Total: 8 documents (7 complete, 1 stub)
```

The new ADR shows up with a checkmark because you filled in the sections.

---

### 2.3 "I want to browse my documentation"

mind-cli gives you five ways to explore your project's documentation without leaving the terminal.

**List all documents**

```bash
$ mind docs list
```

```
spec/
  project-brief.md          ✓ complete    2026-03-01     3.2 KB
  requirements.md           ✓ complete    2026-03-05     8.1 KB
  architecture.md           ✓ complete    2026-03-05     6.4 KB
  domain-model.md           ⚠ stub        2026-02-28     0.4 KB
  decisions/
    001-use-event-sourcing.md  ✓ complete    2026-03-02     1.8 KB
    002-jwt-auth.md            ✓ complete    2026-03-04     2.1 KB

blueprints/
  INDEX.md                  ✓ complete    2026-03-05     0.9 KB
  01-system-design.md       ✓ complete    2026-03-03     4.2 KB
  02-api-design.md          ✓ complete    2026-03-04     3.7 KB

state/
  current.md                ✓ complete    2026-03-10     1.1 KB
  workflow.md               ✓ complete    2026-03-10     0.8 KB

iterations/
  001-NEW_PROJECT-initial/    ✓ complete  2026-03-01
  002-ENHANCEMENT-rbac/       ✓ complete  2026-03-03
  003-ENHANCEMENT-websocket/  ✓ complete  2026-03-05

knowledge/
  glossary.md               ⚠ stub        2026-02-28     0.4 KB
  auth-convergence.md       ✓ complete    2026-03-07     3.2 KB

Total: 18 documents (16 complete, 2 stubs)
```

**Filter by zone**

```bash
$ mind docs list --zone spec
```

Shows only the spec zone. Valid zone names are: `spec`, `blueprints`, `state`, `iterations`, `knowledge`.

**Visual tree**

```bash
$ mind docs tree
```

```
docs/
├── spec/
│   ├── project-brief.md          ✓
│   ├── requirements.md           ✓
│   ├── architecture.md           ✓
│   ├── domain-model.md           ⚠ stub
│   ├── api-contracts.md          ✓
│   └── decisions/
│       ├── 001-use-event-sourcing.md  ✓
│       └── 002-jwt-auth.md            ✓
├── blueprints/
│   ├── INDEX.md                  ✓
│   ├── 01-system-design.md       ✓
│   └── 02-api-design.md          ✓
├── state/
│   ├── current.md                ✓
│   └── workflow.md               ✓
├── iterations/
│   ├── 001-NEW_PROJECT-initial/
│   │   ├── overview.md           ✓
│   │   ├── changes.md            ✓
│   │   ├── test-summary.md       ✓
│   │   ├── validation.md         ✓
│   │   └── retrospective.md      ✓
│   └── ...
└── knowledge/
    ├── glossary.md               ⚠ stub
    └── auth-convergence.md       ✓
```

**Find stubs (documents that need content)**

```bash
$ mind docs stubs
```

```
Stub documents (need content):

  spec/domain-model.md         Needs: entity definitions, relationships, business rules
  knowledge/glossary.md        Needs: domain terminology

2 stubs found. Run '/discover' in Claude Code to generate content,
or edit the files directly.
```

**Search for content**

```bash
$ mind docs search "authentication"
```

```
Found 7 matches in 3 files:

docs/spec/requirements.md (4 matches)
  12: FR-3: The system SHALL authenticate users via JWT tokens
  45: FR-8: The system SHALL validate JWT expiration on every request
  67: FR-11: The system SHALL support multi-factor authentication
  89: AC-3: GIVEN a user with valid credentials WHEN they authenticate THEN a JWT is issued

docs/spec/architecture.md (2 matches)
  34: The authentication layer uses JWT with RS256 signing
  78: JWT tokens are validated in middleware before route handlers

docs/knowledge/auth-convergence.md (1 match)
  15: JWT was selected over session cookies for stateless API design
```

**Open a document in your editor**

```bash
$ mind docs open brief                           # Fuzzy match
$ mind docs open docs/spec/project-brief.md      # Full path
$ mind docs open doc:spec/project-brief           # Document ID
```

All three open the same file. The fuzzy match is the fastest for daily use -- just type enough of the filename to be unique.

---

### 2.4 "I want to validate everything before a PR"

Before you open a pull request, run the full validation suite to catch documentation drift, broken links, and structural issues.

```bash
$ mind check all --strict
```

```
╭─ Documentation (17 checks) ──────────────────────╮
│  Pass: 17  Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

╭─ Cross-References (11 checks) ────────────────────╮
│  Pass: 10  Fail: 1  Warn: 0                       │
│  ✗ [4] Broken link in architecture.md:45           │
│        → "decisions/003-caching.md" (not found)    │
╰────────────────────────────────────────────────────╯

╭─ Conversation Config (4 files) ───────────────────╮
│  Pass: 4   Fail: 0  Warn: 0                       │
╰────────────────────────────────────────────────────╯

Overall: 31/32 pass, 1 fail, 0 warnings
Exit code: 1 (failures present)
```

The cross-reference check found a broken link. You referenced an ADR in `architecture.md` that does not exist. Either create the ADR or fix the link.

The `--strict` flag means warnings are treated as errors. Without it, stub documents are warnings; with it, they are failures. Use `--strict` in CI to enforce complete documentation.

For machine-readable output (useful in CI pipelines), add `--json`:

```bash
$ mind check all --strict --json > results.json
```

This produces a JSON object with the full validation report, including each check ID, pass/fail/warn status, and detailed messages. You can parse this in CI to generate reports or block merges.

The three validation suites and their checks:

| Suite | Checks | What It Validates |
|-------|--------|-------------------|
| **docs** | 17 | Zone directories exist, required files present, stubs detected, mind.toml valid, naming conventions |
| **refs** | 11 | mind.toml paths resolve, no orphan docs, INDEX.md matches blueprints, no broken links, sequence numbers contiguous |
| **config** | variable | Conversation YAML syntax, required fields, referenced files exist |

---

### 2.5 "I want to track changes over time"

Every time you complete a piece of work (whether manually or through an AI workflow), the Mind Framework creates an iteration -- a numbered folder with 5 tracking files that capture what happened.

**See your project history**

```bash
$ mind workflow history
```

```
  #   Type          Name                    Status      Date         Files
  ─── ───────────── ─────────────────────── ─────────── ──────────── ─────
  006 ENHANCEMENT   add-caching             ✓ complete  2026-03-08   5/5
  005 BUG_FIX       fix-auth-redirect       ✓ complete  2026-03-07   5/5
  004 REFACTOR      extract-repositories    ✓ complete  2026-03-06   5/5
  003 ENHANCEMENT   websocket-notifications ✓ complete  2026-03-05   5/5
  002 ENHANCEMENT   role-based-access       ✓ complete  2026-03-03   5/5
  001 NEW_PROJECT   initial-api             ✓ complete  2026-03-01   5/5

6 iterations (6 complete, 0 incomplete)
```

Each iteration is a snapshot of one unit of work. The `Files` column shows artifact completeness -- `5/5` means all 5 tracking files (overview, changes, test summary, validation, retrospective) have content.

**Drill into a specific iteration**

```bash
$ mind workflow show 005
```

```
Iteration: 005-BUG_FIX-fix-auth-redirect

  Type: BUG_FIX
  Descriptor: fix-auth-redirect
  Created: 2026-03-07
  Status: complete

  Artifacts:
    ✓ overview.md          1.2 KB
    ✓ changes.md           3.8 KB
    ✓ test-summary.md      2.1 KB
    ✓ validation.md        1.8 KB
    ✓ retrospective.md     0.9 KB

  Reviewer Findings (from validation.md):
    MUST:   0
    SHOULD: 2
    COULD:  1
    Sign-off: APPROVED
```

This gives you the full picture of what happened during that iteration: what changed, what was tested, what the reviewer found, and whether it was approved.

**Create an iteration manually**

If you are tracking work that was not done through an AI workflow, you can create an iteration by hand:

```bash
$ mind create iteration enhancement "add rate limiting"
```

```
✓ Created docs/iterations/007-ENHANCEMENT-add-rate-limiting/
  ├── overview.md
  ├── changes.md
  ├── test-summary.md
  ├── validation.md
  └── retrospective.md
```

Fill in the files as you work. At minimum, fill in `overview.md` (what you did and why) and `changes.md` (which files changed).

---

## 3. Working with AI Workflows

The Mind Framework supports four integration models for working with AI. Each is progressively more automated. You do not need to use all of them -- pick the one that fits your workflow.

### 3.1 "I want to build a new feature with AI agents" (Model A: Pre-Flight)

Model A is the simplest integration. You run a command to prepare everything, then hand off to Claude Code. Here is the complete walkthrough.

**Step 1: Prepare everything with preflight**

You want to add JWT authentication to your API. Instead of manually creating branches, iteration folders, and assembling context, you run one command:

```bash
$ mind preflight "add: user authentication with JWT tokens"
```

```
╭─ Pre-Flight Complete ──────────────────────────────────────╮
│                                                             │
│  Type: ENHANCEMENT                                          │
│  Chain: analyst → developer → tester → reviewer             │
│  Branch: enhancement/user-authentication                    │
│  Iteration: docs/iterations/008-ENHANCEMENT-user-auth/      │
│  Brief: ✓ present (Vision, Deliverables, Scope)             │
│  Docs: 15/17 pass (2 warnings, 0 blockers)                  │
│                                                             │
│  Context package ready. Open Claude Code and run:            │
│  /workflow "add: user authentication with JWT tokens"        │
│                                                             │
│  Or paste the generated prompt (copied to clipboard).        │
╰─────────────────────────────────────────────────────────────╯
```

Here is what happened behind the scenes, step by step:

1. **Classified the request** -- The keyword "add" maps to ENHANCEMENT. Other keywords: "create"/"new" maps to NEW_PROJECT, "fix"/"bug" maps to BUG_FIX, "refactor"/"clean" maps to REFACTOR.

2. **Ran the business context gate** -- Checked that your project brief exists and has the required sections (Vision, Key Deliverables, Scope). For ENHANCEMENT requests, a missing brief is a warning. For NEW_PROJECT requests, it would be a blocker.

3. **Validated documentation** -- Ran the 17-check docs suite. Warnings are noted but do not block.

4. **Created the iteration** -- Made `docs/iterations/008-ENHANCEMENT-user-auth/` with 5 template files and populated `overview.md` with the classification, request, and agent chain.

5. **Created the git branch** -- Following the `type-descriptor` branch strategy from `mind.toml`, created `enhancement/user-authentication`.

6. **Assembled the context package** -- Read your project brief, requirements, architecture, domain model, the last 3 iterations, and any relevant convergence docs. Wrote the assembled context to `docs/state/workflow.md` so agents can find it.

7. **Generated the prompt** -- Built an orchestrator prompt from agent instructions, project context, iteration context, and conventions. Copied it to your clipboard.

**Step 2: Run the AI workflow in Claude Code**

Open Claude Code in the same project directory and run:

```
/workflow "add: user authentication with JWT tokens"
```

The orchestrator agent reads the pre-flight context from `docs/state/workflow.md`, skips its own classification and validation steps (because you already did them), and dispatches the agent chain:

- **Analyst** extracts requirements, writes acceptance criteria
- **Architect** designs the auth component (skipped for simple enhancements)
- **Developer** implements the solution, writes code, updates `changes.md`
- **Tester** writes and runs tests, fills `test-summary.md`
- **Reviewer** validates the implementation against requirements, writes `validation.md`

**Step 3: Clean up with handoff**

After the AI workflow finishes in Claude Code, come back to the terminal:

```bash
$ mind handoff 008
```

```
Handoff: 008-ENHANCEMENT-user-auth

  1. Iteration Completeness
     ✓ overview.md ✓ changes.md ✓ test-summary.md
     ✓ validation.md ✓ retrospective.md

  2. Deterministic Checks
     ✓ go build    (2.1s)
     ✓ go test     (4.8s, 142 passed)
     ✓ golangci-lint (1.3s)

  3. State Updated
     ✓ docs/state/current.md updated
     ✓ docs/state/workflow.md → idle

  Branch: enhancement/user-authentication (3 commits ahead of main)
  Suggestion: Create a pull request — gh pr create
```

The handoff command:
1. Verified all 5 iteration artifacts are present and have content
2. Ran your build, test, and lint commands to confirm the code is solid
3. Updated `docs/state/current.md` with the completed iteration
4. Cleared the workflow state back to idle
5. Told you the branch status so you can create a PR

**Why is this better than doing it manually?**

Without Model A, you would need to: check the brief yourself, run validation, create the iteration folder manually, number it correctly, create 5 template files, create a git branch with the right naming convention, assemble context documents for the AI, and clean everything up afterward. Preflight and handoff automate all of that into two commands.

---

### 3.2 "I want AI agents to have better tools" (Model B: MCP Server)

MCP stands for Model Context Protocol. It is a way for AI agents to call tools directly, getting structured data back instead of having to read and parse files. Think of it as an API that the AI can call to get project information.

**Setting it up**

Create (or edit) a `.mcp.json` file in your project root:

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

That is the entire setup. When Claude Code starts in your project directory, it reads `.mcp.json`, spawns `mind serve` as a background process, and connects to it over stdio (standard input/output -- no network involved, everything stays local).

**What changes for agents**

Before MCP, when an agent wanted to know the project's health, it had to:

1. Read `docs/state/current.md` (costs input tokens)
2. Scan `docs/spec/` for file presence (multiple file reads)
3. Check each zone for completeness (more file reads)
4. Parse `docs/state/workflow.md` for workflow state (another file read)
5. Reason about all of this scattered information (~100 tokens of reasoning)

After MCP, the agent calls one tool:

```
mind_status → {
  "project": "my-awesome-api",
  "zones": {
    "spec": { "total": 5, "complete": 4, "stubs": 1 },
    "blueprints": { "total": 3, "complete": 3, "stubs": 0 },
    ...
  },
  "workflow": { "state": "idle", "last_iteration": "006-ENHANCEMENT-add-caching" },
  "warnings": ["domain-model.md is a stub"]
}
```

One call, structured data, no file parsing, no heuristic errors. The MCP server exposes 16 tools total (see [BP-07](07-ai-workflow-integration.md) for the complete list), covering status, diagnostics, validation, iteration management, workflow state, quality logging, and more.

**You do not need to change your workflow.** Model B is invisible to you. You still use `/workflow`, `/discover`, and `/analyze` in Claude Code the same way. The agents are just smarter and faster because they get structured data instead of parsing raw markdown.

---

### 3.3 "I want to watch what AI is doing in real-time" (Model C: Watch)

When an AI workflow is running in Claude Code, you might want to see what is happening without interrupting it. The watch command monitors your project's filesystem and reports changes as they happen.

**Plain log mode**

Open a second terminal (or a tmux pane) and run:

```bash
$ mind watch
```

```
[14:23:01] Watching ~/dev/projects/my-awesome-api/
[14:23:15] docs/state/workflow.md updated: developer active (008-ENHANCEMENT-user-auth)
[14:23:32] src/middleware/jwt.go created
[14:23:45] src/routes/auth.go created
[14:24:01] docs/iterations/008-ENHANCEMENT-user-auth/changes.md updated
[14:24:02] Micro-Gate B: ✓ changes.md present, 4 files verified on disk
[14:24:45] Background: go build ✓ (2.1s)
[14:25:12] Background: go test ✓ 24 passed (4.8s)
[14:26:30] docs/state/workflow.md updated: tester active
[14:28:15] docs/iterations/008-ENHANCEMENT-user-auth/test-summary.md updated
[14:29:00] docs/state/workflow.md updated: reviewer active
[14:30:45] docs/iterations/008-ENHANCEMENT-user-auth/validation.md updated
[14:31:00] docs/state/workflow.md updated: idle (workflow complete)
```

Each line is timestamped and describes what happened. You can see:
- When the developer agent starts writing code
- Which files get created
- Whether background builds and tests pass
- When the workflow transitions between agents
- When the workflow completes

**TUI dashboard mode**

For a richer view, add `--tui`:

```bash
$ mind watch --tui
```

```
╭─ mind watch ─── my-awesome-api ─── enhancement/user-auth ──────────────╮
│                                                                          │
│  Workflow: ENHANCEMENT (008-user-auth)                                   │
│  Chain: analyst [done] → developer [done] → [tester] → reviewer         │
│                                                                          │
│  ╭─ Live Activity ──────────────────────────────────────────────────╮    │
│  │  14:26:30  Tester writing tests for JWT middleware               │    │
│  │  14:27:15  test-summary.md updated: 8 test cases                │    │
│  │  14:28:00  Tester writing tests for auth endpoints               │    │
│  │  14:28:30  test-summary.md updated: 12 test cases               │    │
│  │  14:28:45  Background: go test ✓ 36 passed (5.2s)              │    │
│  ╰──────────────────────────────────────────────────────────────────╯    │
│                                                                          │
│  ╭─ Gate Status ────────────────────────────────────────────────────╮    │
│  │  Build: [pass]          Lint: [pass]          Tests: 36/36 [pass]│    │
│  │  Micro-Gate A: [pass]   Micro-Gate B: [pass]  Det. Gate: ready  │    │
│  ╰──────────────────────────────────────────────────────────────────╯    │
│                                                                          │
│  Warnings: domain-model.md is still a stub                               │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
```

The TUI dashboard updates in real-time. The chain progress bar shows which agent is currently active. The gate status panel shows whether the code is in a good state. If a build or test fails, you see it immediately and can decide whether to intervene.

**When to intervene**

Watch mode is passive -- it never stops the AI from working. But it gives you signals:
- **Build fails** -- The developer wrote code that does not compile. Usually the AI will catch and fix this.
- **Tests fail** -- New code broke existing tests. If the AI does not fix it, you may want to step in.
- **Gate fails repeatedly** -- If a micro-gate fails and the agent retries but fails again, you might want to investigate.

Press `Ctrl+C` in the watch terminal to stop watching. This does not affect the AI workflow running in Claude Code.

---

### 3.4 "I want to run a complete AI workflow from the terminal" (Model D: Full Orchestration)

Model D is the most automated option. Instead of running preflight, switching to Claude Code, running the workflow, and running handoff, you type one command and the entire pipeline runs from your terminal.

**First, see what would happen (dry run)**

```bash
$ mind run --dry-run "fix: 500 error on /api/users endpoint"
```

```
Dry Run — no AI calls will be made

  Classification: BUG_FIX
  Chain: analyst → developer → tester → reviewer
  Business Context Gate: PASS
  Iteration: would create 009-BUG_FIX-fix-500-error-users/
  Branch: would create bugfix/fix-500-error-users

  Agent Dispatch Plan:
    1. Analyst  (opus)   — analyze the bug, produce requirements
    2. Developer (sonnet) — implement the fix
    3. Tester   (sonnet) — write regression tests
    4. Reviewer (opus)   — validate the fix

  Quality Gates:
    Micro-Gate A — after analyst
    Micro-Gate B — after developer
    Deterministic Gate — before reviewer
      Commands: go build -o mind ., go test ./..., golangci-lint run ./...

  Estimated cost: ~$1.50-3.00 (4 agent calls)
```

The dry run shows you exactly what will happen: which type it classified as, which agents will run, which models they use, what quality gates fire, and an estimated cost. No files are created, no branches are made, no AI calls happen.

**Then run for real**

```bash
$ mind run "fix: 500 error on /api/users endpoint"
```

```
Running: BUG_FIX — "fix: 500 error on /api/users endpoint"

  ✓ Pre-flight     classify, gate, iteration, branch          0.8s
  ✓ Analyst         bug analyzed (3 FR, 4 AC)                  1m 42s
  ✓ Micro-Gate A    6/6 checks pass                            0.2s
  ✓ Developer       fix implemented (2 files changed)          2m 18s
  ✓ Micro-Gate B    changes.md present, files verified         0.1s
  ✓ Tester          4 regression tests added                   1m 56s
  ✓ Det. Gate       build ✓ lint ✓ tests 148 pass              8.3s
  ✓ Reviewer        approved (0 MUST, 1 SHOULD)                1m 24s
  ✓ Handoff         iteration 009 complete, state cleared      0.5s

Complete in 8m 32s
  Branch: bugfix/fix-500-error-users (2 commits ahead of main)
  Suggestion: Create a pull request — gh pr create
```

The entire pipeline ran automatically:
1. Pre-flight: classified the request, checked the brief, validated docs, created the iteration and branch
2. Analyst: analyzed the bug and extracted requirements with acceptance criteria
3. Micro-Gate A: verified the requirements have the right structure (GIVEN/WHEN/THEN, FR identifiers)
4. Developer: implemented the fix, recorded changed files in `changes.md`
5. Micro-Gate B: confirmed `changes.md` exists and all listed files are on disk
6. Tester: wrote regression tests for the fix
7. Deterministic Gate: ran build, lint, and test commands -- all passed
8. Reviewer: validated the implementation against the requirements, approved
9. Handoff: verified all iteration artifacts, updated project state, cleared workflow

**With the TUI pipeline view**

For a richer visual experience during the run:

```bash
$ mind run --tui "fix: 500 error on /api/users endpoint"
```

This opens a full-screen TUI that shows the pipeline progress, live agent output, and gate results updating in real-time. Same behavior, better visibility.

**Resuming if something goes wrong**

If the run is interrupted (Ctrl+C, network issue, crash), the state is saved to `docs/state/workflow.md`. Resume where you left off:

```bash
$ mind run --resume
```

This reads the saved state, determines which agent was running, and picks up from there.

**Important requirements for Model D**

- The `claude` CLI must be installed and on your `$PATH`. If it is not, `mind run` will tell you with install instructions.
- Each agent runs as a separate `claude --print` process. The mind CLI is the dispatcher -- it builds prompts, pipes them to Claude, captures output, and runs gates between agents.
- If a gate fails, the agent is re-dispatched with the gate feedback (up to 2 retries, as configured in `mind.toml` `governance.max-retries`). After retries are exhausted, the pipeline proceeds with documented concerns.

---

## 4. Advanced Usage

### 4.1 "I want to track document staleness" (Reconciliation)

Documents in a project have dependencies. Your `architecture.md` depends on `requirements.md`. If the requirements change but the architecture document is not updated, the architecture is stale -- it describes a system that no longer matches the requirements.

The reconciliation engine makes this staleness visible and machine-detectable. For a deep dive into how the engine works, see [BP-06](06-reconciliation-engine.md).

**Initial setup**

The first time you run reconcile, it creates a `mind.lock` file with SHA-256 hashes of every registered document:

```bash
$ mind reconcile
```

```
Reconciliation Report

  Registry vs. Disk:
    ✓ 12 documents in sync

  State Freshness:
    ✓ current.md references latest iteration (006)

  Working Tree:
    ✓ No uncommitted doc changes

  Created mind.lock with 12 document hashes
```

The `mind.lock` file stores the content hash, modification time, and staleness status of every document declared in `mind.toml`. It also uses the `[[graph]]` edges in `mind.toml` to understand which documents depend on which.

**Later, after editing requirements.md**

Say you add a new functional requirement (FR-12) to `requirements.md`. Now run reconcile again:

```bash
$ mind reconcile
```

```
Reconciliation Report

  Registry vs. Disk:
    ✓ 12 documents in sync

  Staleness:
    ⚠ 3 documents are stale:
      ● architecture.md     (dependency changed: requirements.md)
      ● domain-model.md     (dependency changed: requirements.md)
      ● api-contracts.md    (dependency changed: requirements.md → architecture.md)

  State Freshness:
    ✓ current.md references latest iteration (006)

  Working Tree:
    ⚠ 1 uncommitted doc change: requirements.md

  Updated mind.lock
```

The engine detected that `requirements.md` changed (different SHA-256 hash), walked the dependency graph forward, and marked three downstream documents as stale:

- `architecture.md` directly depends on `requirements.md` (one hop)
- `domain-model.md` directly depends on `requirements.md` (one hop)
- `api-contracts.md` depends on `architecture.md` which depends on `requirements.md` (two hops -- staleness propagates transitively)

**Check in CI (read-only)**

For CI pipelines, use `--check` which does not write `mind.lock`:

```bash
$ mind reconcile --check
```

This exits with code `0` if everything is fresh, or code `1` if stale documents are detected. No files are modified -- it only reports.

**After updating all stale documents**

Once you have reviewed and updated `architecture.md`, `domain-model.md`, and `api-contracts.md` to account for the new requirement:

```bash
$ mind reconcile --force
```

```
Reconciliation Report

  Registry vs. Disk:
    ✓ 12 documents in sync

  Staleness:
    ✓ 0 stale documents (all fresh)

  Updated mind.lock (forced refresh)
```

The `--force` flag recomputes all hashes regardless of modification times and clears all staleness flags. Use it after you have genuinely updated the stale documents. Do not use it to silence warnings without actually reviewing the documents -- you would be lying to future developers (and AI agents) about the state of your documentation.

**When to use --force vs. updating documents**

| Situation | Action |
|-----------|--------|
| You updated the stale documents to reflect the upstream changes | `mind reconcile --force` to refresh all hashes |
| You reviewed the stale documents and confirmed they are still accurate | `mind reconcile --force` -- the review itself is the update |
| You just want the warnings to go away without reviewing anything | Do not use `--force`. Fix the documents first. |
| You are in CI and want to detect staleness | `mind reconcile --check` (read-only) |

---

### 4.2 "I want to use mind-cli in CI/CD"

mind-cli produces meaningful exit codes and supports JSON output, making it straightforward to integrate into CI pipelines.

**Basic documentation check**

```yaml
# .github/workflows/docs-check.yml
name: Documentation Check
on: [pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/jf-ferraz/mind-cli@latest
      - name: Validate documentation structure
        run: mind check all --strict --json > results.json
      - name: Check for stale documents
        run: mind reconcile --check
      - name: Verify agent sync
        run: mind sync agents --check
```

What each step catches:

- **`mind check all --strict`** -- All 17 docs checks + 11 refs checks + config checks. With `--strict`, stubs and missing optional files are treated as failures. Exit code `1` on any failure.
- **`mind reconcile --check`** -- Detects if documents have drifted from their dependencies. Exit code `1` if any document is stale. This catches the case where someone updated requirements but forgot to update architecture.
- **`mind sync agents --check`** -- Verifies that `.github/agents/` matches `.mind/agents/`. Exit code `1` if they are out of sync. This catches the case where someone updated an agent definition in `.mind/` but forgot to sync to Copilot.

**Making it a required check**

In your GitHub repository settings:
1. Go to Settings -> Branches -> Branch protection rules
2. Edit the rule for `main`
3. Under "Require status checks to pass before merging", add your workflow name
4. Now PRs cannot be merged if documentation checks fail

**Handling failures**

When `mind check all --strict` fails in CI:

```bash
# Run locally to see what failed
$ mind check all --strict

# Common fixes:
# - Stub document? Fill it with content.
# - Broken link? Fix the reference or create the target file.
# - Missing file? Create it with 'mind create' or 'mind doctor --fix'.
# - Naming convention? Rename the file to kebab-case.
```

When `mind reconcile --check` fails in CI:

```bash
# Run locally to see which documents are stale
$ mind reconcile

# Review and update each stale document
# Then force-refresh the hashes
$ mind reconcile --force
```

**Exit code reference**

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success, everything passes |
| `1` | Failures found (validation, staleness, sync issues) |
| `2` | User abort (Ctrl+C during interactive command) |
| `3` | Not a Mind project (`.mind/` directory not found) |

---

### 4.3 "I want to explore my project interactively" (TUI)

The full-screen TUI dashboard gives you a visual interface with 5 tabs. Launch it:

```bash
$ mind tui
```

**Tab 1: Status (press `1`)**

```
╭─ Mind Framework ─── my-awesome-api ─── main ────── v2026-03-09 ─╮
│                                                                   │
│  [1 Status]  [2 Docs]  [3 Iterations]  [4 Check]  [5 Quality]   │
│ ───────────────────────────────────────────────────────────────── │
│                                                                   │
│  Documentation Health          │  Active Workflow                  │
│  ───────────────────           │  ───────────────                  │
│  spec/        ████████░░ 4/5   │  State: idle                     │
│  blueprints/  ██████████ 3/3   │  Last: 006-ENHANCEMENT           │
│  state/       ██████████ 2/2   │        add-caching               │
│  iterations/  ██████████ 6/6   │                                   │
│  knowledge/   ██████░░░░ 3/5   │  Quick Actions                   │
│                                 │  ──────────                      │
│  Staleness                     │  c  Create document               │
│  ──────────                    │  d  Run doctor                    │
│  ✓ All documents fresh         │  v  Validate all                  │
│                                 │  r  Reconcile                    │
│  Warnings (1)                  │  o  Open document                 │
│  ────────────                  │                                   │
│  ⚠ domain-model.md is a stub  │                                   │
╰───────────────────────────────────────────────────────────────────╯
```

The Status tab is a two-column layout. The left column shows documentation health (progress bars per zone), staleness status, and warnings. The right column shows the active workflow (or "idle" with the last completed iteration) and quick-action keyboard shortcuts.

The quick actions let you trigger common commands without leaving the TUI: press `d` to run doctor, `v` to validate all, `r` to reconcile, `c` to create a document, or `o` to open a document in your editor.

**Tab 2: Documents (press `2`)**

```
│  Zone: [a All]  s Spec  b Blueprints  t State  i Iterations  k Knowledge │
│  Search: /                                                                │
│ ───────────────────────────────────────────────────────────────────────── │
│                                                                           │
│  spec/                                                                    │
│  ├── project-brief.md          ✓ draft       2026-03-01     3.2 KB       │
│  ├── requirements.md           ✓ draft       2026-03-05     8.1 KB       │
│  ├── architecture.md           ✓ draft       2026-03-05     6.4 KB       │
│  ├── domain-model.md           ✗ stub        2026-02-28     0.4 KB       │
│  └── decisions/                                                           │
│      ├── 001-use-postgresql.md ✓ complete    2026-03-02     1.8 KB       │
│      └── 002-jwt-auth.md       ✓ complete    2026-03-04     2.1 KB       │
│                                                                           │
│  blueprints/                                                              │
│  ├── INDEX.md                  ✓ active      2026-03-10     0.9 KB       │
│  ...                                                                      │
│                                                                           │
│  ↑↓ navigate  Enter preview  e edit  / search  Esc clear      3/18 docs  │
```

Browse all documents grouped by zone. Use the letter shortcuts along the top to filter by zone (`s` for spec, `b` for blueprints, etc.). Press `/` to search. Press `Enter` on any document to open a right-side preview pane that renders the markdown. Press `e` to open in your editor. Stub documents appear in red; stale documents appear in yellow.

**Tab 3: Iterations (press `3`)**

```
│  Filter: [a All]  n NEW  e ENH  b BUG  r REFACTOR                        │
│ ───────────────────────────────────────────────────────────────────────── │
│                                                                           │
│   #    Type           Name                     Status       Date   Files  │
│  ──── ────────────── ──────────────────────── ────────── ──────── ─────── │
│  006  ENHANCEMENT    add-caching              ✓ complete  03-08    5/5    │
│  005  BUG_FIX        fix-auth-redirect        ✓ complete  03-07    5/5    │
│  004  REFACTOR       extract-repositories     ✓ complete  03-06    5/5    │
│  003  ENHANCEMENT    websocket-notifications  ✓ complete  03-05    5/5    │
│  002  ENHANCEMENT    role-based-access        ✓ complete  03-03    5/5    │
│  001  NEW_PROJECT    initial-api              ✓ complete  03-01    5/5    │
│                                                                           │
│  ↑↓ navigate  Enter expand  o overview  v validation  filter: n/e/b/r    │
```

View your project's iteration history as a table. Filter by type using the letter shortcuts (`n` for NEW_PROJECT, `e` for ENHANCEMENT, etc.). Press `Enter` to expand an iteration inline and see its artifact details. Each iteration type is color-coded: NEW_PROJECT in blue, ENHANCEMENT in cyan, BUG_FIX in red, REFACTOR in magenta.

**Tab 4: Checks (press `4`)**

Run validation interactively. The tab shows the three validation suites (docs, refs, config) with their individual checks. You can select a suite and press `Enter` to run it. Results update in place as each check completes. Failed checks are highlighted and you can press `Enter` on a failure to see the detailed error message and suggested fix.

**Tab 5: Quality (press `5`)**

View convergence analysis quality trends. Shows a table of quality log entries with scores per dimension, overall scores, and Gate 0 results. If you have enough data points, an ASCII trend line shows whether quality is improving over time. This tab reads from `docs/knowledge/quality-log.yml`.

**Navigation**

| Key | Action |
|-----|--------|
| `1`-`5` | Switch between tabs |
| `q` or `Ctrl+C` | Quit the TUI |
| `Up`/`Down` | Navigate within a tab |
| `Enter` | Select/expand/preview |
| `Esc` | Close preview/clear filter |
| `/` | Start search (Tab 2) |
| `PgUp`/`PgDn` | Page through content |

---

### 4.4 "I want to track convergence analysis quality"

A convergence analysis is a multi-persona deep-dive on a topic (run via `/analyze` in Claude Code). It produces a document where different expert perspectives are synthesized into actionable recommendations with a quality rubric. mind-cli can track the quality of these analyses over time.

**Step 1: Run a convergence analysis in Claude Code**

```
/analyze "authentication strategy for the REST API"
```

This produces a file like `docs/knowledge/authentication-strategy-convergence.md` with persona sections, synthesis, and a quality rubric scoring 6 dimensions on a 1-5 scale.

**Step 2: Log the quality scores**

```bash
$ mind quality log docs/knowledge/authentication-strategy-convergence.md
```

```
Logged quality scores from: authentication-strategy-convergence.md

  Topic: Authentication Strategy
  Dimensions:
    Breadth of Perspectives:  4/5
    Depth of Analysis:        3/5
    Practical Applicability:  4/5
    Intellectual Rigor:       4/5
    Synthesis Quality:        3/5
    Actionable Outcomes:      4/5

  Overall: 3.7/5.0 — Gate 0: PASS
  Variant: (none)

✓ Appended to docs/knowledge/quality-log.yml
```

Gate 0 is the quality threshold: an overall score of 3.0 or higher passes. Below 3.0, the analysis likely needs revision. The quality entry is appended to `docs/knowledge/quality-log.yml`, which serves as the historical record.

**Step 3: View quality trends**

```bash
$ mind quality history
```

```
Quality Score History

  Date        Topic                    Overall  Gate 0
  ──────────  ───────────────────────  ───────  ──────
  2026-03-01  Authentication Strategy  2.3/5.0  FAIL
  2026-03-03  Authentication Strategy  3.2/5.0  PASS
  2026-03-05  Database Selection       3.5/5.0  PASS
  2026-03-07  Authentication Strategy  3.8/5.0  PASS

  Trend: 2.3 → 3.2 → 3.5 → 3.8 (improving ↑)
```

The history shows how your analysis quality has changed over time. In this example, the first authentication analysis scored below the Gate 0 threshold, was revised, and improved significantly.

**Step 4: Get the big picture**

```bash
$ mind quality report
```

```
Quality Report

  Topics Analyzed: 3
  Average Overall Score: 3.5/5.0
  Gate 0 Pass Rate: 2/3 (67%)

  Best Dimension:    Breadth of Perspectives (avg 4.2)
  Weakest Dimension: Synthesis Quality (avg 2.8)

  Topics Needing Attention:
    ⚠ Caching Strategy — latest score 2.5/5.0 (Gate 0: FAIL)

  Topics Passing:
    ✓ Authentication Strategy — 3.8/5.0
    ✓ Database Selection — 3.5/5.0
```

The report aggregates across all topics, identifies your strongest and weakest dimensions, and flags topics that have not yet passed Gate 0.

If you want to log a revised version of an analysis:

```bash
$ mind quality log docs/knowledge/authentication-strategy-convergence.md --variant "v2-revised"
```

The `--variant` flag lets you track multiple iterations of the same analysis without overwriting the original entry.

---

### 4.5 "I want to sync agents to GitHub Copilot"

If your team uses GitHub Copilot Chat alongside Claude Code, you can synchronize agent definitions so both platforms use the same persona definitions.

**Check the current state**

```bash
$ mind sync agents --check
```

```
Checking agent sync: .mind/agents/ → .github/agents/

  ✓ analyst.md        (in sync)
  ✗ architect.md      (out of sync — 3 lines differ)
  ✓ developer.md      (in sync)
  ✗ tester.md         (missing in target)
  ✓ reviewer.md       (in sync)

Result: 2 out of sync (exit code 1)
```

The `--check` flag reports differences without changing anything. Exit code `1` means the platforms are out of sync.

**Sync the agents**

```bash
$ mind sync agents
```

```
Syncing agents: .mind/agents/ → .github/agents/

  ✓ analyst.md        (up to date)
  ↻ architect.md      (updated — 3 lines changed)
  ✓ developer.md      (up to date)
  + tester.md         (new — copied)
  ✓ reviewer.md       (up to date)

Synced: 1 updated, 1 new, 3 unchanged
```

The canonical agent definitions live in `.mind/agents/`. The sync command copies them to `.github/agents/` so Copilot Chat picks them up. If `.github/agents/` does not exist, the sync command creates it.

Use `--check` in CI to ensure agents stay in sync:

```yaml
- name: Verify agent sync
  run: mind sync agents --check
```

---

## 5. Troubleshooting

### 5.1 "mind status shows warnings I don't understand"

Here is every warning message `mind status` can produce, what it means, and how to fix it.

| Warning | What It Means | How to Fix |
|---------|---------------|------------|
| "X is a stub" | The file exists but contains only template headings and placeholder text. No real content has been written. | Open the file and fill in the sections. For spec documents, run `/discover` in Claude Code to generate content. For the glossary, add domain terms as you encounter them. |
| "No workflow state saved" | No AI workflow is currently in progress. `docs/state/workflow.md` is idle or does not exist. | Nothing to fix -- this is normal when you are not in the middle of an AI workflow. |
| "N documents are stale" | Upstream dependencies changed and downstream documents may be outdated. For example, `requirements.md` changed but `architecture.md` was not updated to match. | Review each stale document. Update it if needed, then run `mind reconcile --force` to clear the staleness flags. |
| "glossary.md missing" | The `docs/knowledge/glossary.md` file does not exist. | Create it with `mind doctor --fix` (creates a stub) or write it manually. The glossary is optional but helpful for AI agents to understand domain terminology. |
| "No iterations found" | The `docs/iterations/` directory is empty. No work has been tracked yet. | This is normal for new projects. Iterations are created automatically by `mind preflight` or manually with `mind create iteration`. |
| "X is missing" (for required files) | A file that should exist in the documentation structure is absent. | Run `mind doctor --fix` to create stubs, or create the file manually. |
| "mind.toml missing" | The project manifest does not exist. | Run `mind init` to create it, or create it manually using the schema in [BP-03](03-data-contracts.md). |
| "N uncommitted doc changes" | Documentation files have been modified but not committed to git. | Commit or stash the changes. This is informational -- it does not affect functionality. |

---

### 5.2 "mind check fails but I don't know why"

When `mind check all` reports failures, the output tells you the check number and what failed. Here is how to read it and fix common issues.

**Reading the output**

```
╭─ Cross-References (11 checks) ────────────────────╮
│  Pass: 10  Fail: 1  Warn: 0                       │
│  ✗ [4] Broken link in architecture.md:45           │
│        → "decisions/003-caching.md" (not found)    │
╰────────────────────────────────────────────────────╯
```

The check number `[4]` is the cross-reference check for broken internal markdown links. The message tells you the exact file (`architecture.md`), line number (45), and what is broken (the link target does not exist).

**Common failures and fixes**

| Check | Failure | Fix |
|-------|---------|-----|
| docs [6] | mind.toml invalid TOML | Open `mind.toml` and fix the syntax error. Common issues: missing quotes around strings, unclosed brackets, or invalid dates. |
| docs [16] | Stub documents found | Fill the documents with real content. This is a warning by default, but becomes a failure with `--strict`. |
| docs [17] | Naming convention violation | Rename the file to kebab-case. Example: rename `My Blueprint.md` to `my-blueprint.md`. You can run `mind doctor --fix` to auto-rename. |
| refs [1] | mind.toml path does not exist | A document is registered in `mind.toml` but the file is missing from disk. Either create the file or remove the registry entry. |
| refs [2] | Orphan document | A file exists in `docs/` but is not registered in `mind.toml`. Add it to the `[documents]` section. |
| refs [3] | INDEX.md mismatch | `docs/blueprints/INDEX.md` lists a blueprint that does not exist, or a blueprint exists but is not listed. Update the INDEX. |
| refs [4] | Broken internal link | A markdown link `[text](path)` points to a file that does not exist. Fix the path or create the target file. |
| refs [5] | Iteration missing overview.md | An iteration directory exists but has no `overview.md`. Create the file. |
| config | Missing required field | A conversation YAML file is missing a required field. Check the error message for the field name and file path. |

**Strict mode**

Without `--strict`, stubs and some optional files are reported as warnings (exit code `0` if no failures). With `--strict`, they become failures (exit code `1`). Use `--strict` in CI to enforce complete documentation.

**JSON output for debugging**

```bash
$ mind check all --json | jq '.suites[] | select(.fail_count > 0)'
```

This filters the JSON output to show only suites with failures, making it easier to pinpoint issues in large projects.

---

### 5.3 "mind doctor says something is wrong"

The doctor command groups diagnostics into categories. Here is how to fix issues in each category.

**Framework (`✗ Framework not installed`)**

The `.mind/` directory is missing or incomplete.

```bash
# If the directory is completely missing:
$ mind init

# If the directory exists but is corrupted:
$ mind doctor --fix
```

**Adapters (`✗ Copilot adapter not found`)**

The `.github/agents/` directory is missing.

```bash
$ mind init --with-github
# Or:
$ mind sync agents
```

**Documentation structure (`✗ N checks fail`)**

Missing directories or required files.

```bash
$ mind doctor --fix
```

This creates missing directories, adds `.gitkeep` files, and creates stub documents from templates.

**Stubs (`✗ N stub documents found`)**

Files exist but contain only template content.

The doctor cannot fix stubs automatically because it cannot write meaningful content for you. Options:
- Fill them in manually
- Run `/discover` in Claude Code to generate content
- Leave them as stubs if you are not ready to fill them in (they will show as warnings)

**Cross-references (`✗ N checks fail`)**

Broken links or registry inconsistencies.

```bash
# Fix broken links in markdown files manually
# Then verify:
$ mind check refs
```

**Project brief (`⚠ Missing "Key Deliverables" section`)**

The brief exists but is incomplete.

```bash
$ mind docs open brief
# Add the missing section, then verify:
$ mind doctor
```

**Workflow state (`✗ Stale workflow state`)**

The workflow state references an iteration or branch that no longer exists.

```bash
$ mind workflow clean
```

---

### 5.4 "The MCP server isn't working"

If Claude Code is not picking up the MCP tools, check these things in order.

**1. Is `.mcp.json` formatted correctly?**

The file must be valid JSON and use the exact structure:

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

Common mistakes:
- Trailing commas after the last field (invalid JSON)
- Wrong key names (`servers` instead of `mcpServers`)
- `command` value is a full path instead of just `mind` (works, but make sure the path is correct)

**2. Is the `mind` binary on your PATH?**

```bash
$ which mind
/usr/local/bin/mind

$ mind version
mind version 0.1.0
```

If `which mind` returns nothing, the binary is not on your PATH. Either add it or use the full path in `.mcp.json`:

```json
{
  "mcpServers": {
    "mind": {
      "command": "/home/you/go/bin/mind",
      "args": ["serve"]
    }
  }
}
```

**3. Does `mind serve` start without errors?**

Test it manually:

```bash
$ mind serve
```

If the project is not a Mind project, you will see an error about `.mind/` not being found. The MCP server requires a valid project to serve.

The server communicates via stdin/stdout using JSON-RPC. When you run it manually, it will appear to hang (it is waiting for input). Press Ctrl+C to stop it.

**4. Restart Claude Code**

Claude Code reads `.mcp.json` at startup. If you added or changed the file after starting Claude Code, you need to restart Claude Code for it to detect the new configuration.

**5. Check Claude Code logs**

If the MCP server starts but tools are not appearing, check Claude Code's output for error messages related to MCP initialization or tool registration.

---

### 5.5 "mind run failed mid-workflow"

If `mind run` stops unexpectedly (crash, Ctrl+C, network issue), the workflow state is preserved in `docs/state/workflow.md`.

**Step 1: Check the current state**

```bash
$ mind workflow status
```

```
Workflow: active
  Type: ENHANCEMENT
  Iteration: docs/iterations/008-ENHANCEMENT-user-auth/
  Branch: enhancement/user-authentication
  Last Agent: developer (completed)
  Remaining: tester → reviewer
  Session: 1 of 1

  Completed Artifacts:
    ✓ requirements.md (analyst)
    ✓ architecture.md (architect)
    ✓ changes.md (developer)
```

This tells you exactly where the workflow stopped. The developer finished, but the tester never started.

**Step 2: Resume**

```bash
$ mind run --resume
```

This picks up from the tester and continues through the reviewer and handoff.

**Step 3: If resume does not work**

If the state is corrupted or the resume fails:

```bash
# Check what is in the iteration folder
$ mind workflow show 008

# If the code changes are good but the workflow state is broken:
$ mind workflow clean       # Reset workflow state to idle
$ mind handoff 008          # Manually run the handoff
```

**Step 4: Manual recovery**

In the worst case, if both resume and clean fail:

1. Check `docs/state/workflow.md` -- it is a markdown file you can read and edit
2. Check the iteration folder for artifacts that were created
3. Run your build/test/lint commands manually to verify the code is good
4. Reset the workflow state: edit `docs/state/workflow.md` and set it to idle
5. Create a PR from whatever state the branch is in

---

## 6. Cheat Sheet

```
ESSENTIALS
  mind status                          How is my project doing?
  mind doctor [--fix]                  What is wrong? (and fix it)
  mind check all [--strict]            Validate everything

CREATE THINGS
  mind create brief                    Start a project brief (interactive)
  mind create iteration new "X"        Start a new iteration
  mind create adr "X"                  Architecture decision record
  mind create blueprint "X"            Planning document
  mind create spike "X"                Technical spike report
  mind create convergence "X"          Convergence analysis template

BROWSE DOCS
  mind docs list [--zone spec]         List documents
  mind docs stubs                      What needs content?
  mind docs search "auth"              Find content
  mind docs tree                       Visual overview
  mind docs open brief                 Open in $EDITOR

TRACK STATE
  mind reconcile                       Update hashes, detect staleness
  mind reconcile --check               Read-only check (for CI)
  mind reconcile --force               Force-refresh all hashes
  mind workflow status                 Current workflow state
  mind workflow history                Past iterations
  mind workflow show 007               Iteration details
  mind workflow clean                  Clear stale workflow state

AI WORKFLOWS
  mind preflight "request"             Prepare for AI workflow (Model A)
  mind preflight --resume              Check for resumable workflow
  mind handoff 007                     Clean up after AI workflow (Model A)
  mind serve                           Start MCP server (Model B)
  mind watch [--tui]                   Monitor AI in real-time (Model C)
  mind run "request"                   Full AI orchestration (Model D)
  mind run --dry-run "request"         Preview what would happen
  mind run --resume                    Resume interrupted workflow

QUALITY
  mind quality log <file>              Log convergence quality scores
  mind quality history                 Score trends over time
  mind quality report                  Aggregate quality summary

SYNC & PLATFORM
  mind sync agents                     Sync agents to GitHub Copilot
  mind sync agents --check             Check sync status (for CI)

INTERACTIVE
  mind tui                             5-tab dashboard
                                         1 Status  2 Docs  3 Iterations
                                         4 Checks  5 Quality  q Quit

GLOBAL FLAGS
  --json / -j                          JSON output
  --no-color                           Disable colors
  --verbose / -v                       Debug logging
  --project-root <path>                Override project root

META
  mind version [--short]               Version and build info
  mind completion bash|zsh|fish        Shell completions
  mind help [command]                  Help for any command
```

---

## 7. Frequently Asked Questions

**Q: Do I need Claude Code to use mind-cli?**

No. The core CLI (status, check, create, docs, doctor, reconcile, workflow, tui, quality) works entirely standalone. You need Claude Code only for AI workflows -- specifically, `/workflow`, `/discover`, and `/analyze` (Models A and B). Model D (`mind run`) needs the `claude` CLI on your PATH but does not require the Claude Code application.

**Q: Does mind-cli send data to the internet?**

No. Everything runs locally on your machine. The MCP server communicates over stdio (standard input/output), not over the network. The `mind` binary never makes HTTP requests, never calls APIs, never phones home. The only exception is Model D (`mind run`), which pipes prompts to the `claude` CLI -- but the `claude` CLI handles the network communication, not `mind`.

**Q: Can I use mind-cli with languages other than Go?**

Yes. mind-cli is language-agnostic. It manages documentation structure, not code. The `mind.toml` `[project.stack]` section records your language and framework for display purposes, and `[project.commands]` records your build/test/lint commands. You can use mind-cli with Python, Rust, TypeScript, Java, or any other language. Set the appropriate commands in `mind.toml` and everything works.

**Q: What is the difference between `mind check` and `mind doctor`?**

`mind check` runs validation suites and reports pass/fail/warn for each check. It is objective and granular -- 17 docs checks, 11 refs checks, config checks. Think of it as a test suite for your documentation.

`mind doctor` runs all the same validators but goes further: it diagnoses root causes, groups related failures, and suggests specific fixes. It also has the `--fix` flag to auto-fix structural issues. Think of it as a doctor's visit -- not just "what is wrong" but "why is it wrong and how to fix it."

**Q: Why do I need mind.lock?**

`mind.lock` stores SHA-256 content hashes for every document registered in `mind.toml`. It serves two purposes:

1. **Change detection** -- By comparing the current file hash to the stored hash, the reconciliation engine knows exactly which documents changed since the last check.
2. **Staleness propagation** -- Using the dependency graph from `mind.toml`'s `[[graph]]` edges, it can determine which downstream documents are potentially outdated because an upstream dependency changed.

Without `mind.lock`, you would have no way to automatically detect that editing `requirements.md` might have made `architecture.md` outdated.

**Q: Can I use mind-cli without `.mind/`?**

Only three commands work without a project: `mind init` (to create one), `mind version`, and `mind help`. Everything else requires a `.mind/` directory. If you run a command without one, you get exit code `3` and a suggestion to run `mind init`.

**Q: How do I update mind-cli?**

```bash
$ go install github.com/jf-ferraz/mind-cli@latest
```

This downloads and installs the latest version. Your project's `.mind/` directory, `docs/`, and `mind.toml` are not affected by CLI updates -- they are project data, not CLI data.

**Q: What are the 5 request types?**

When you run `mind preflight` or `mind run`, the CLI classifies your request based on keywords:

| Type | Keywords | Agent Chain |
|------|----------|-------------|
| NEW_PROJECT | "create", "new" | analyst, architect, developer, tester, reviewer |
| BUG_FIX | "fix", "bug" | analyst, developer, tester, reviewer |
| ENHANCEMENT | "add", "improve", "update", "enhance" | analyst, developer, tester, reviewer |
| REFACTOR | "refactor", "clean", "reorganize" | analyst, developer, tester, reviewer |
| COMPLEX_NEW | "complex", "analyze" | analyst, architect, developer, tester, reviewer, moderator |

NEW_PROJECT and COMPLEX_NEW include the architect agent. COMPLEX_NEW adds a moderator for multi-perspective governance. The other types skip the architect for efficiency.

**Q: What are quality gates and why do they exist?**

Quality gates are deterministic checks that run between AI agents to catch problems before they propagate. There are three types:

- **Micro-Gate A** (after analyst) -- Checks that requirements have GIVEN/WHEN/THEN acceptance criteria, FR-N identifiers, and a scope boundary.
- **Micro-Gate B** (after developer) -- Checks that `changes.md` exists and all listed files are present on disk.
- **Deterministic Gate** (before reviewer) -- Runs the build, lint, and test commands from `mind.toml`. This is the hard gate -- if the code does not compile or tests fail, the developer is re-dispatched with the failure details.

Gates exist because AI agents make mistakes. Without gates, a developer agent that produces code with syntax errors would hand off to the tester, who would waste tokens trying to test broken code. Gates catch these issues early and trigger retries.

**Q: What happens if a gate fails?**

The agent that produced the failing output is re-dispatched with the gate feedback. The retry prompt includes the specific check failures and what needs to be fixed. The maximum number of retries is configured in `mind.toml` (`governance.max-retries`, default `2`). If retries are exhausted and the gate still fails, the pipeline continues with documented concerns -- the failure is recorded in the iteration's `validation.md` rather than blocking the workflow.

**Q: Can I use mind-cli with a monorepo?**

Each Mind Framework project is rooted at a `.mind/` directory. In a monorepo, you would initialize mind-cli in each subproject that needs its own documentation structure. The CLI walks up from your current directory looking for `.mind/`, so you just need to be inside the right subproject when running commands. Alternatively, use `--project-root` to specify which project to target.

**Q: What is the difference between Model A, B, C, and D?**

| Model | What It Does | Effort | Value |
|-------|-------------|--------|-------|
| **A: Pre-Flight** | Prepares everything before AI workflow, cleans up after | Low | High |
| **B: MCP Server** | Gives AI agents structured tools instead of file parsing | None (invisible) | Very High |
| **C: Watch** | Shows real-time progress while AI works | None (passive) | High |
| **D: Orchestration** | Runs the entire AI workflow from one terminal command | Low | Very High |

You can use them independently or together. Most users start with A (preflight + handoff), add B (MCP server) for better agent accuracy, and optionally use C (watch) for visibility. D (orchestration) is for users who want full automation.

**Q: How do I check the convergence quality of an analysis?**

Use the `mind check convergence` command with the path to the convergence file:

```bash
$ mind check convergence docs/knowledge/auth-convergence.md
```

This runs 23 checks across 4 categories (structure, persona quality, synthesis quality, quality rubric) and reports the overall score with a Gate 0 pass/fail result. See [BP-04](04-cli-specification.md) for the full list of checks.

**Q: What is COMPLEX_NEW and when do I use it?**

COMPLEX_NEW is a request type for projects that need multi-perspective analysis before implementation. It adds a moderator agent that coordinates the convergence analysis (multiple expert personas debating the approach). Use it when:
- The problem space is ambiguous and you need to explore options
- Multiple valid architectural approaches exist
- The decision has significant long-term consequences

Trigger it with keywords like "complex" or "analyze" in your request, or when the analyst determines the request warrants deeper analysis.

**Q: Can I customize the agent definitions?**

The agent definitions live in `.mind/agents/` as markdown files. You can edit them to change agent behavior, add domain-specific instructions, or adjust quality criteria. After editing, run `mind sync agents` if you want to keep Copilot agents in sync. Note that agent definitions are part of the framework -- changing them affects how all AI workflows behave for your project.

**Q: How do I handle a session split in Model D?**

For NEW_PROJECT and COMPLEX_NEW request types, Model D automatically splits the workflow after the architect completes. This is because implementation (developer, tester, reviewer) benefits from a human review of the requirements and architecture before proceeding. When a split occurs, `mind run` saves the state and prompts you to continue or pause. If you pause, you can resume later with `mind run --resume`. In non-interactive mode (piped or CI), the split auto-continues.

**Q: What is the `--from-existing` flag on `mind init`?**

If your project already has documentation in a `docs/` directory, `mind init --from-existing` scans for existing files, preserves them, and creates stubs only for files that are missing. It registers all discovered files in `mind.toml`. This is the safe way to adopt mind-cli in an existing project without losing any documentation you have already written.
