# internal/validate/

| File | When to Read |
|------|-------------|
| `check.go` | Check framework: `Suite`, `Check`, `CheckFunc`, `CheckContext`; `Suite.Run()` executes checks and builds `ValidationReport` |
| `docs.go` | `DocsSuite()` — 17-check documentation validation (zone dirs, required files, naming, stubs, brief completeness) |
| `refs.go` | `RefsSuite()` — 11-check cross-reference validation (CLAUDE.md refs, INDEX.md links, mind.toml paths, sequencing) |
| `config.go` | `ConfigSuite()` — 10-check mind.toml schema validation (schema format, naming, doc entries, governance) |
