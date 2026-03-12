package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/deps"
)

// ToolDefinition describes a single MCP tool for tools/list.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is a JSON Schema object describing tool parameters.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property is a single JSON Schema property.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// AllToolDefinitions returns the 16 MCP tool schemas.
func AllToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "mind_status",
			Description: "Return a structured project health summary: documentation completeness per zone, active workflow state, iteration count, and warnings.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_doctor",
			Description: "Run deep diagnostics across the entire project and return structured findings with severity levels and suggested fixes.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_check_brief",
			Description: "Evaluate the business context gate. Parse project-brief.md and return whether it exists, is a stub, and which required sections are present.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_validate_docs",
			Description: "Run the 17-check documentation validation suite and return structured results for each check.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"strict": {Type: "boolean", Description: "Promote warnings to failures"},
				},
			},
		},
		{
			Name:        "mind_validate_refs",
			Description: "Run the 11-check cross-reference validation suite. Verify links between documents resolve correctly.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_list_iterations",
			Description: "Return a list of all iterations with sequence number, type, descriptor, status, date, and artifact completeness.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_show_iteration",
			Description: "Return detailed information about a single iteration including its overview content and artifact list.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "string", Description: "Iteration ID, directory name, or 3-digit sequence number"},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "mind_read_state",
			Description: "Read and parse docs/state/workflow.md, returning the current workflow state as structured data.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_update_state",
			Description: "Write workflow state to docs/state/workflow.md. Pass null or empty to reset to idle.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"type":            {Type: "string", Description: "Request type: NEW_PROJECT, BUG_FIX, ENHANCEMENT, REFACTOR, COMPLEX_NEW"},
					"descriptor":      {Type: "string", Description: "Short slug describing the workflow"},
					"iteration_path":  {Type: "string", Description: "Relative path to the iteration folder"},
					"branch":          {Type: "string", Description: "Git branch name"},
					"last_agent":      {Type: "string", Description: "Last completed agent name"},
					"remaining_chain": {Type: "string", Description: "Comma-separated remaining agents"},
					"session":         {Type: "string", Description: "Current session number"},
				},
			},
		},
		{
			Name:        "mind_create_iteration",
			Description: "Create a new iteration directory with the correct sequence number, type-based naming, and 5 template files.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"type": {Type: "string", Description: "Iteration type: new, enhancement, bugfix, refactor"},
					"name": {Type: "string", Description: "Short descriptor for the iteration"},
				},
				Required: []string{"type", "name"},
			},
		},
		{
			Name:        "mind_list_stubs",
			Description: "Scan all documentation zones and return a list of stub documents that need content.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_check_gate",
			Description: "Run the deterministic gate: execute build, lint, and test commands from mind.toml and return structured results.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_log_quality",
			Description: "Extract quality scores from a convergence analysis file and append them to docs/knowledge/quality-log.yml.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {Type: "string", Description: "Path to convergence analysis file (relative to project root)"},
					"topic":     {Type: "string", Description: "Topic override (optional)"},
					"variant":   {Type: "string", Description: "Variant override (optional)"},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "mind_search_docs",
			Description: "Full-text search across all files in docs/. Returns matching file paths, line numbers, and context.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {Type: "string", Description: "Search query string"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "mind_read_config",
			Description: "Parse mind.toml and return the project configuration as structured JSON.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mind_suggest_next",
			Description: "Analyze the current project state and suggest the next action based on brief, workflow, and iteration status.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
	}
}

// RegisterTools registers all 16 tool handlers on the server.
func RegisterTools(s *Server) {
	s.RegisterTool("mind_status", handleMindStatus)
	s.RegisterTool("mind_doctor", handleMindDoctor)
	s.RegisterTool("mind_check_brief", handleMindCheckBrief)
	s.RegisterTool("mind_validate_docs", handleMindValidateDocs)
	s.RegisterTool("mind_validate_refs", handleMindValidateRefs)
	s.RegisterTool("mind_list_iterations", handleMindListIterations)
	s.RegisterTool("mind_show_iteration", handleMindShowIteration)
	s.RegisterTool("mind_read_state", handleMindReadState)
	s.RegisterTool("mind_update_state", handleMindUpdateState)
	s.RegisterTool("mind_create_iteration", handleMindCreateIteration)
	s.RegisterTool("mind_list_stubs", handleMindListStubs)
	s.RegisterTool("mind_check_gate", handleMindCheckGate)
	s.RegisterTool("mind_log_quality", handleMindLogQuality)
	s.RegisterTool("mind_search_docs", handleMindSearchDocs)
	s.RegisterTool("mind_read_config", handleMindReadConfig)
	s.RegisterTool("mind_suggest_next", handleMindSuggestNext)
}

// --- Tool handlers ---

func handleMindStatus(d *deps.Deps, _ json.RawMessage) (any, error) {
	project, err := d.ProjectSvc.DetectProject(d.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("detect project: %w", err)
	}
	return d.ProjectSvc.AssembleHealth(project)
}

func handleMindDoctor(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.DoctorSvc.Run(false), nil
}

func handleMindCheckBrief(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.BriefRepo.ParseBrief()
}

func handleMindValidateDocs(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		Strict bool `json:"strict"`
	}
	if params != nil {
		_ = json.Unmarshal(params, &p)
	}
	report := d.ValidationSvc.RunDocs(d.ProjectRoot, p.Strict)
	return report, nil
}

func handleMindValidateRefs(d *deps.Deps, _ json.RawMessage) (any, error) {
	report := d.ValidationSvc.RunRefs(d.ProjectRoot)
	return report, nil
}

func handleMindListIterations(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.WorkflowSvc.History()
}

func handleMindShowIteration(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return d.WorkflowSvc.Show(p.ID)
}

func handleMindReadState(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.StateRepo.ReadWorkflow()
}

func handleMindUpdateState(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		Type           string `json:"type"`
		Descriptor     string `json:"descriptor"`
		IterationPath  string `json:"iteration_path"`
		Branch         string `json:"branch"`
		LastAgent      string `json:"last_agent"`
		RemainingChain string `json:"remaining_chain"`
		Session        int    `json:"session"`
	}
	if params != nil {
		_ = json.Unmarshal(params, &p)
	}

	if p.Type == "" {
		// Reset to idle
		if err := d.StateRepo.WriteWorkflow(nil); err != nil {
			return nil, err
		}
		return map[string]string{"status": "idle"}, nil
	}

	state := &domain.WorkflowState{
		Type:          domain.RequestType(p.Type),
		Descriptor:    p.Descriptor,
		IterationPath: p.IterationPath,
		Branch:        p.Branch,
		LastAgent:     p.LastAgent,
		Session:       p.Session,
	}
	if p.RemainingChain != "" {
		for _, a := range splitComma(p.RemainingChain) {
			state.RemainingChain = append(state.RemainingChain, a)
		}
	}

	if err := d.StateRepo.WriteWorkflow(state); err != nil {
		return nil, err
	}
	return map[string]string{"status": "updated"}, nil
}

func handleMindCreateIteration(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return d.GenerateSvc.CreateIteration(p.Type, p.Name)
}

func handleMindListStubs(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.ProjectSvc.ListStubs()
}

func handleMindCheckGate(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.ValidationSvc.RunGate(d.ProjectRoot), nil
}

func handleMindLogQuality(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		FilePath string `json:"file_path"`
		Topic    string `json:"topic"`
		Variant  string `json:"variant"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if d.QualitySvc == nil {
		return nil, fmt.Errorf("quality service not available")
	}
	return d.QualitySvc.Log(p.FilePath, p.Topic, p.Variant)
}

func handleMindSearchDocs(d *deps.Deps, params json.RawMessage) (any, error) {
	var p struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return d.ProjectSvc.SearchDocs(p.Query)
}

func handleMindReadConfig(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.ProjectSvc.Config()
}

func handleMindSuggestNext(d *deps.Deps, _ json.RawMessage) (any, error) {
	return d.ProjectSvc.SuggestNext(d.ProjectRoot)
}

func splitComma(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			result = append(result, t)
		}
	}
	return result
}
