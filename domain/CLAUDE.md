# domain/

| File | When to Read |
|------|-------------|
| `project.go` | Core types: `Project`, `Config`, `Manifest`, `ProjectMeta`, `StackConfig`, `CmdConfig`, `Governance`, `DocEntry` |
| `document.go` | `Document`, `DocStatus` enum, `Brief`, `BriefGate` enum |
| `zones.go` | `Zone` enum, `AllZones`, `ValidZone()`, `ZoneNames()` |
| `health.go` | Computed aggregates: `ProjectHealth`, `ZoneHealth`, `Diagnostic`, `DoctorReport`, `InitResult`, `CreateResult`, `DocumentList`, `StubList`, `SearchResults`, `VersionInfo` |
| `validation.go` | `CheckLevel` enum, `CheckResult`, `ValidationReport`, `UnifiedValidationReport` |
| `iteration.go` | `Iteration`, `IterationStatus` enum, `RequestType` enum, `Classify()`, `Slugify()` |
| `workflow.go` | `WorkflowState`, `CompletedArtifact`, `DispatchEntry` |
| `errors.go` | Sentinel errors: `ErrNotProject`, `ErrBriefMissing`, `ErrAlreadyInitialized`, `ErrAlreadyExists`; structured errors: `ErrGateFailed`, `ErrCommandFailed` |
| `purity_test.go` | NFR-4 enforcement — verifies zero external imports in domain package |
