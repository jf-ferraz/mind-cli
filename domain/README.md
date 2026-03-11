# domain/

## Purity Constraint

This package has zero external imports. Only Go standard library packages are allowed. Specifically banned: `os`, `path/filepath`, `io`, `net`, and all third-party packages (`github.com/`, `golang.org/`).

This constraint is enforced by `purity_test.go`, which parses every `.go` file's imports via `go/parser` and fails on any banned import. The test runs on every `go test ./...` invocation.

**Consequence**: Path strings in domain types (`Project.Root`, `Document.Path`, `Iteration.Path`) are opaque data. The domain never interprets paths -- that responsibility belongs to the infrastructure layer (`internal/repo/fs/`).

## Type Design

- Enums are typed string constants (`Zone`, `DocStatus`, `BriefGate`, `RequestType`, `IterationStatus`, `CheckLevel`), not raw strings or iota integers. This makes JSON output human-readable and `--json` contracts self-documenting.
- Sentinel errors (`ErrNotProject`, `ErrBriefMissing`) express domain concepts. Structured errors (`ErrGateFailed`, `ErrCommandFailed`) carry domain-relevant fields for rendering.
- JSON struct tags on output types (`ProjectHealth`, `ValidationReport`, etc.) define the `--json` serialization contract. The domain owns these contracts, not the renderer.

## Computed vs Persisted

All aggregates in `health.go` (`ProjectHealth`, `DoctorReport`, `DocumentList`, `SearchResults`, etc.) are computed on-demand from repository data. They are never persisted to disk. State machines (iteration lifecycle, brief gate, validation check) are derived from disk state, not stored.
