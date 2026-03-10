---
name: doc-sync
description: Synchronizes docs across a repository. Use when user asks to sync docs.
---

# Doc Sync

Maintains the CLAUDE.md navigation hierarchy and README.md invisible knowledge docs across a repository. This skill is self-contained and performs all documentation work directly.

## Documentation Conventions

Read `.mind/conventions/documentation.md` for authoritative CLAUDE.md and README.md format specification.

## Scope Resolution

| User Request                                            | Scope                                     |
| ------------------------------------------------------- | ----------------------------------------- |
| "sync docs" / "update documentation" / no specific path | REPOSITORY-WIDE                           |
| "sync docs in src/validator/"                           | DIRECTORY: src/validator/ and descendants |
| "update CLAUDE.md for parser.py"                        | FILE: single file's parent directory      |

For REPOSITORY-WIDE scope, perform a full audit. For narrower scopes, operate only within the specified boundary.

## Workflow

### Phase 1: Discovery

Map directories requiring CLAUDE.md verification:

```bash
find . -type d \( -name .git -o -name node_modules -o -name __pycache__ -o -name .venv -o -name target -o -name dist -o -name build \) -prune -o -type d -print
```

For each directory in scope, record: Does CLAUDE.md exist? Does it have the required table-based index? What files/subdirectories need indexing?

### Phase 2: Audit

For each directory, check for drift: missing entries, stale entries (deleted files), misplaced content (architecture/design docs that belong in README.md), and whether README.md is warranted.

### Phase 3: Content Migration

If CLAUDE.md contains explanatory content, migrate it to README.md:
- **Move to README.md:** Architecture explanations, design decisions, component interactions, "why" explanations, invariants, constraints, dependency rationale
- **Keep in CLAUDE.md:** Build/test/deploy commands, regeneration/sync commands, operational procedures

**Test:** "Is this explaining WHY or telling HOW?" Explanatory → README.md. Operational → CLAUDE.md.

### Phase 4: Index Updates

1. Use appropriate template (ROOT or SUBDIRECTORY)
2. Populate tables with all files and subdirectories
3. Write action-oriented "When to read" triggers
4. If README.md exists, include it in the index

### Phase 5: Verification

1. Every directory in scope has CLAUDE.md
2. All CLAUDE.md files use pure table-based index format
3. No drift (files ↔ index entries match)
4. No misplaced explanatory content in CLAUDE.md
5. README.md exists wherever invisible knowledge was identified

## Exclusions

DO NOT create CLAUDE.md for: generated output dirs, vendored dependencies, .git/, IDE configs (unless project-specific), stub directories (only .gitkeep).

## Anti-Patterns

**Too vague:** `| config/ | Configuration | Working with configuration |`
**Correct:** `| config/ | YAML config parsing, env overrides | Adding config options, changing defaults |`

## Output Format

```
## Doc Sync Report
### Scope: [REPOSITORY-WIDE | directory path]
### Changes Made
- CREATED: [new CLAUDE.md files]
- UPDATED: [modified CLAUDE.md files]
- MIGRATED: [content moved from CLAUDE.md to README.md]
- FLAGGED: [issues requiring human decision]
### Verification
- Directories audited: [count]
- CLAUDE.md coverage: [count]/[total]
- Drift detected: [count] entries fixed
```

## Reference

For additional trigger pattern examples, see `references/trigger-patterns.md`.
