# internal/service/

| File | When to Read |
|------|-------------|
| `project.go` | `ProjectService` — health assembly from docs, iterations, workflow, brief |
| `validation.go` | `ValidationService` — orchestrates doc/ref/config suites, unified reports |
| `generate.go` | `GenerateService` — ADR, blueprint, iteration, spike, convergence scaffolding with auto-sequencing |
| `workflow.go` | `WorkflowService` — workflow state reads, iteration history assembly |
| `init.go` | `InitService` — project initialization: directories, stubs, mind.toml, adapters |
| `doctor.go` | `DoctorService` — diagnostic checks across framework, docs, brief, config, iterations; auto-fix support |
