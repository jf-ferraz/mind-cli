# internal/repo/

| File/Package | When to Read |
|-------------|-------------|
| `interfaces.go` | Repository contracts: `DocRepo`, `IterationRepo`, `StateRepo`, `ConfigRepo`, `BriefRepo` |
| `fs/` | Filesystem implementations — real I/O for production use |
| `mem/` | In-memory implementations — deterministic testing without filesystem |
