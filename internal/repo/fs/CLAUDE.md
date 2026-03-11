# internal/repo/fs/

| File | When to Read |
|------|-------------|
| `project.go` | `FindProjectRoot()`, `FindProjectRootFrom()` — walk-up `.mind/` detection; `DetectProject()` — project assembly with config loading |
| `doc_repo.go` | `DocRepo` — filesystem-backed document queries: list by zone, list all, read, exists, stub detection |
| `config_repo.go` | `ConfigRepo` — mind.toml parsing and writing via go-toml/v2 |
| `brief_repo.go` | `BriefRepo` — project brief section analysis (vision, deliverables, scope) |
| `iteration_repo.go` | `IterationRepo` — iteration folder scanning, artifact detection, sequence derivation |
| `state_repo.go` | `StateRepo` — workflow state parsing from docs/state/workflow.md |
