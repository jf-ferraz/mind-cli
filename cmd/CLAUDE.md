# cmd/

| File | When to Read |
|------|-------------|
| `root.go` | Global flags (`--json`, `--no-color`, `--project`), Cobra root initialization |
| `helpers.go` | Project root resolution, error classification helpers |
| `status.go` | `mind status` — project health and documentation status |
| `init.go` | `mind init` — project initialization with zone scaffolding |
| `doctor.go` | `mind doctor` — full project diagnostics with `--fix` |
| `create.go` | `mind create` — scaffold ADR, blueprint, iteration, spike, convergence, brief |
| `check.go` | `mind check` — validation suites (docs, refs, config, all) with `--strict` |
| `docs.go` | `mind docs` — list, tree, stubs, search, open documents |
| `workflow.go` | `mind workflow` — workflow state and iteration history |
| `brief.go` | `mind brief` — project brief status and gate result |
| `iterations.go` | `mind iterations` — list all iterations |
| `version.go` | `mind version` — build info with ldflags injection |
