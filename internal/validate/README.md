# internal/validate/

## Check Framework

Validation uses a `Suite` containing ordered `Check` items. Each check has an ID, name, severity level, and a `CheckFunc` that receives a `CheckContext` and returns pass/fail with a message.

Suites are stateless: `Suite.Run()` executes all checks sequentially, collects results into a `domain.ValidationReport`.

## Suite Composition

Three suites cover the validation surface:

| Suite | Checks | File | Scope |
|-------|--------|------|-------|
| `DocsSuite()` | 17 | `docs.go` | Zone structure, required files, naming conventions, stubs, brief completeness |
| `RefsSuite()` | 11 | `refs.go` | CLAUDE.md references, INDEX.md links, mind.toml paths, cross-references, sequence gaps |
| `ConfigSuite()` | 10 | `config.go` | mind.toml schema format, project name, doc entries, governance settings |

`mind check all` runs all three suites and merges results into a `UnifiedValidationReport`.

## Strict Mode

The `--strict` flag promotes `WARN`-level checks to `FAIL`. This is handled in `Suite.Run()` by checking `ctx.Strict` before classifying results. Check functions do not need to know about strict mode.

## Adding a Check

1. Write a `CheckFunc` in the appropriate suite file
2. Append a `Check` to the suite's `Checks` slice with the next sequential ID
3. Choose a `CheckLevel` (`LevelFail`, `LevelWarn`, `LevelInfo`)

Check IDs are stable -- existing IDs must not be renumbered. New checks get the next available ID in their suite.
