# Current State

## Active Work

None — iteration 001-core-cli is complete.

## Known Issues

- **SHOULD**: `--project` flag should be `--project-root` per api-contracts spec
- **SHOULD**: `docs search` bypasses DocRepo abstraction (C-9 deviation)
- **SHOULD**: 5 exported methods in fs/doc_repo.go lack GoDoc comments (NFR-8)
- **SHOULD**: Repo wiring in command handlers instead of main.go (C-10 deviation, acknowledged)
- **COULD**: DoctorService reimplements checks instead of delegating to ValidationService

## Recent Changes

- **2026-03-11** — Phase 1 Core CLI implemented (@iteration/001)
  - 50 FRs implemented across 20+ commands
  - 395 tests, all passing
  - domain/ 100% coverage, validate/ 90.7%
  - Binary size 7.1MB, domain purity maintained

## Next Priorities

- Fix SHOULD items from reviewer (flag rename, GoDoc, search abstraction)
- Centralize repo wiring in main.go before Phase 2
- Phase 1.5: Reconciliation engine (mind.lock, staleness propagation)
