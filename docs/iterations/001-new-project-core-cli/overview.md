# Core CLI

- **Type**: NEW_PROJECT
- **Request**: Using all blueprints as reference, start implementations following provided guidelines and requirements
- **Agent Chain**: analyst → architect → developer → tester → reviewer
- **Branch**: feature/core-cli
- **Created**: 2026-03-11

## Scope

Phase 1 of the implementation roadmap (BP-08): Build the Core CLI that replaces bash scripts and provides the full command surface for deterministic project management. Includes domain types, repository layer, service layer, validation engine, rendering, document generation, and 20+ CLI commands with --json support.

## Requirement Traceability

| Req ID | Description | Analyst | Architect | Developer | Reviewer |
|--------|-------------|---------|-----------|-----------|----------|
| FR-1 | Project root auto-detection (walk up for .mind/) | ✓ | ✓ | | |
| FR-2 | --project-root flag override | ✓ | ✓ | | |
| FR-3 | mind.toml full parsing (all sections) | ✓ | ✓ | | |
| FR-4 | Exit code 3 when not a Mind project | ✓ | ✓ | | |
| FR-5 | Degraded mode when mind.toml absent | ✓ | ✓ | | |
| FR-6 | Auto-detect output mode (interactive/plain/JSON) | ✓ | ✓ | | |
| FR-7 | --json on every structured command | ✓ | ✓ | | |
| FR-8 | Interactive mode with Lip Gloss styling | ✓ | ✓ | | |
| FR-9 | Plain mode with no ANSI codes | ✓ | ✓ | | |
| FR-10 | Stdout for data, stderr for errors/progress | ✓ | ✓ | | |
| FR-11 | mind status dashboard (zones, workflow, warnings) | ✓ | ✓ | | |
| FR-12 | mind status --json (ProjectHealth schema) | ✓ | ✓ | | |
| FR-13 | Brief gate classification (PRESENT/STUB/MISSING) | ✓ | ✓ | | |
| FR-14 | mind init (create .mind/, docs/, mind.toml) | ✓ | ✓ | | |
| FR-15 | mind init creates .claude/CLAUDE.md adapter | ✓ | ✓ | | |
| FR-16 | mind init --name flag | ✓ | ✓ | | |
| FR-17 | mind init --with-github flag | ✓ | ✓ | | |
| FR-18 | mind init --from-existing flag | ✓ | ✓ | | |
| FR-19 | mind init abort if .mind/ exists (exit 2) | ✓ | ✓ | | |
| FR-20 | mind doctor full diagnostics | ✓ | ✓ | | |
| FR-21 | Actionable remediation in doctor output | ✓ | ✓ | | |
| FR-22 | mind doctor --fix auto-remediation | ✓ | ✓ | | |
| FR-23 | Partial fix reporting with exit code 1 | ✓ | ✓ | | |
| FR-24 | mind create adr (auto-numbered) | ✓ | ✓ | | |
| FR-25 | mind create blueprint (auto-numbered + INDEX.md) | ✓ | ✓ | | |
| FR-26 | mind create iteration (4 types, 5 artifacts) | ✓ | ✓ | | |
| FR-27 | mind create spike | ✓ | ✓ | | |
| FR-28 | mind create convergence | ✓ | ✓ | | |
| FR-29 | mind create brief (interactive) | ✓ | ✓ | | |
| FR-30 | Create commands abort on existing target | ✓ | ✓ | | |
| FR-31 | Title slugification | ✓ | ✓ | | |
| FR-32 | mind docs list (all zones) | ✓ | ✓ | | |
| FR-33 | mind docs list --zone filter | ✓ | ✓ | | |
| FR-34 | mind docs tree (visual tree) | ✓ | ✓ | | |
| FR-35 | mind docs stubs | ✓ | ✓ | | |
| FR-36 | mind docs search (full-text) | ✓ | ✓ | | |
| FR-37 | mind docs open (fuzzy match + $EDITOR) | ✓ | ✓ | | |
| FR-38 | mind check docs (17 checks) | ✓ | ✓ | | |
| FR-39 | mind check docs --strict | ✓ | ✓ | | |
| FR-40 | mind check refs (11 checks) | ✓ | ✓ | | |
| FR-41 | mind check config (mind.toml schema validation) | ✓ | ✓ | | |
| FR-42 | mind check all (unified report) | ✓ | ✓ | | |
| FR-43 | Check exit codes (0 pass, 1 fail) | ✓ | ✓ | | |
| FR-44 | mind workflow status | ✓ | ✓ | | |
| FR-45 | mind workflow history | ✓ | ✓ | | |
| FR-46 | mind version (full info) | ✓ | ✓ | | |
| FR-47 | mind version --short | ✓ | ✓ | | |
| FR-48 | mind help [command] | ✓ | ✓ | | |
| FR-49 | Consistent exit codes (0/1/2/3) | ✓ | ✓ | | |
| FR-50 | Stub detection algorithm | ✓ | ✓ | | |
| NFR-1 | mind status < 200ms for 10-50 docs | ✓ | ✓ | | |
| NFR-2 | CLI startup < 50ms | ✓ | ✓ | | |
| NFR-3 | Binary < 15MB (linux/amd64) | ✓ | ✓ | | |
| NFR-4 | domain/ zero external imports | ✓ | ✓ | | |
| NFR-5 | Test coverage: domain/ 80%, validate/ 80%, overall 70% | ✓ | ✓ | | |
| NFR-6 | Validation parity with bash scripts | ✓ | ✓ | | |
