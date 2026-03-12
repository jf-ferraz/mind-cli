# internal/repo/

## Repository Pattern

Interfaces are defined in `interfaces.go`, close to where they are consumed by services and validation checks. Implementations live in sub-packages:

- `fs/` — Filesystem implementations for production. These are the only packages that call `os.ReadFile`, `os.Stat`, `filepath.Walk`, etc.
- `mem/` — In-memory implementations for tests. Same interfaces, map/slice-backed storage.

## Interface-at-Consumer-Site

Go idiom: interfaces are defined where they are consumed, not where they are implemented. The five repository interfaces (`DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `BriefRepo`) are small and focused. They live in a shared `interfaces.go` because Phase 1's repo count is manageable. If interface count exceeds 10, split into per-consumer interfaces.

## Dependency Direction

Repositories return domain types. They convert filesystem state (file existence, content parsing, directory scanning) into domain objects. The infrastructure layer depends on the domain layer, never the reverse.

The `fs/` package is the only place that imports `github.com/pelletier/go-toml/v2` (for `mind.toml` parsing in `ConfigRepo`). All other third-party dependencies are confined to the presentation layer (`cobra`, `x/term`).
