# mind-cli

| Resource | When to Read |
|----------|-------------|
| `mind.toml` | Project manifest — identity, stack, document registry, profiles |
| `cmd/` | CLI command handlers (Cobra) — presentation layer |
| `domain/` | Pure domain types, enums, business rules — zero external imports |
| `internal/service/` | Business logic orchestration — service layer |
| `internal/validate/` | Validation check framework — 17 doc + 11 ref + 10 config checks |
| `internal/generate/` | Document template rendering |
| `internal/render/` | Output formatting — interactive, plain, JSON modes |
| `internal/repo/` | Repository interfaces and implementations (fs, mem) |
| `docs/spec/architecture.md` | 4-layer architecture, key decisions, dependency matrix |
| `docs/spec/requirements.md` | Functional and non-functional requirements |
| `docs/spec/domain-model.md` | Entities, business rules, constraints |
| `docs/state/current.md` | Active state, known issues, priorities |
| `docs/blueprints/INDEX.md` | System-level planning artifacts |
