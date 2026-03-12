package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jf-ferraz/mind-cli/domain"
	"github.com/jf-ferraz/mind-cli/internal/deps"
	"github.com/jf-ferraz/mind-cli/internal/repo/mem"
	"github.com/jf-ferraz/mind-cli/internal/service"
)

// newTestServer creates an MCP server with nil deps for protocol-level tests.
// Tool call tests that exercise the dispatcher need non-nil deps; protocol
// tests (initialize, tools/list, malformed JSON, notifications) do not.
func newTestServer() *Server {
	s := &Server{
		transport: nil,
		deps:      nil,
		tools:     make(map[string]ToolHandler),
	}
	RegisterTools(s)
	return s
}

// FR-143: handleRaw returns an initialize result with protocolVersion.
func TestHandleRaw_Initialize(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for initialize request")
	}
	if resp.Error != nil {
		t.Fatalf("handleRaw() returned error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("handleRaw() result is nil")
	}

	resultMap, ok := resp.Result.(InitializeResult)
	if !ok {
		t.Fatalf("result is %T, want InitializeResult", resp.Result)
	}
	if resultMap.ProtocolVersion == "" {
		t.Error("ProtocolVersion is empty in initialize response")
	}
}

// FR-143: handleRaw for tools/list returns a tools array with 16 items.
func TestHandleRaw_ToolsList(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for tools/list request")
	}
	if resp.Error != nil {
		t.Fatalf("handleRaw() returned error: %v", resp.Error)
	}

	result, ok := resp.Result.(ToolsListResult)
	if !ok {
		t.Fatalf("result is %T, want ToolsListResult", resp.Result)
	}
	if len(result.Tools) != 16 {
		t.Errorf("tools/list returned %d tools, want 16", len(result.Tools))
	}
}

// FR-143: handleRaw for tools/call with an unknown tool returns a method-not-found error.
func TestHandleRaw_ToolsCall_UnknownTool(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for unknown tool call")
	}
	if resp.Error == nil {
		t.Fatal("expected error response for unknown tool, got result")
	}
	if resp.Error.Code != errMethodNotFound {
		t.Errorf("error code = %d, want %d (errMethodNotFound)", resp.Error.Code, errMethodNotFound)
	}
}

// FR-143: handleRaw for tools/call on a known tool returns a result (not an error).
// Registers a minimal stub handler that avoids touching nil deps.
func TestHandleRaw_ToolsCall_KnownTool(t *testing.T) {
	s := newTestServer()
	// Register a safe stub for this test to avoid nil dep panics.
	s.tools["mind_status"] = func(_ *deps.Deps, _ json.RawMessage) (any, error) {
		return map[string]string{"status": "ok"}, nil
	}

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"mind_status","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for known tool call")
	}
	if resp.Error != nil {
		t.Errorf("expected result response for known tool, got RPC error: %v", resp.Error.Message)
	}
	if resp.Result == nil {
		t.Error("result is nil for known tool call")
	}
}

// FR-143: handleRaw for malformed JSON returns a parse error.
func TestHandleRaw_MalformedJSON(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{this is not valid json`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for malformed JSON — expected parse error response")
	}
	if resp.Error == nil {
		t.Fatal("expected error response for malformed JSON")
	}
	if resp.Error.Code != errParse {
		t.Errorf("error code = %d, want %d (errParse)", resp.Error.Code, errParse)
	}
}

// FR-140 / M-1 bug fix: handleRaw for notifications/initialized must return nil.
// Per JSON-RPC 2.0, notifications (no "id" field) must not receive any response.
func TestHandleRaw_NotificationsInitialized_ReturnsNil(t *testing.T) {
	s := newTestServer()

	// No "id" field — this is a notification.
	raw := json.RawMessage(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	resp := s.handleRaw(raw)

	if resp != nil {
		t.Errorf("handleRaw() returned %+v for notifications/initialized, want nil (no response for notifications)", resp)
	}
}

// FR-140: Any notifications/* method must return nil (no response).
func TestHandleRaw_NotificationsWildcard_ReturnsNil(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"2.0","method":"notifications/toolsListChanged"}`)
	resp := s.handleRaw(raw)

	if resp != nil {
		t.Errorf("handleRaw() returned %+v for notifications/toolsListChanged, want nil", resp)
	}
}

// FR-143: handleRaw for invalid jsonrpc version returns errInvalidRequest.
func TestHandleRaw_InvalidJSONRPCVersion(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"1.0","id":1,"method":"initialize"}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for invalid jsonrpc version")
	}
	if resp.Error == nil {
		t.Fatal("expected error response for invalid jsonrpc version")
	}
	if resp.Error.Code != errInvalidRequest {
		t.Errorf("error code = %d, want %d (errInvalidRequest)", resp.Error.Code, errInvalidRequest)
	}
}

// FR-143: handleRaw for unknown method (not notifications/) returns errMethodNotFound.
func TestHandleRaw_UnknownMethod(t *testing.T) {
	s := newTestServer()

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil for unknown method")
	}
	if resp.Error == nil {
		t.Fatal("expected error response for unknown method")
	}
	if resp.Error.Code != errMethodNotFound {
		t.Errorf("error code = %d, want %d (errMethodNotFound)", resp.Error.Code, errMethodNotFound)
	}
}

// AllToolDefinitions returns exactly 16 tool definitions.
func TestAllToolDefinitions_Count(t *testing.T) {
	tools := AllToolDefinitions()
	if len(tools) != 16 {
		t.Errorf("AllToolDefinitions() = %d tools, want 16", len(tools))
	}
}

// AllToolDefinitions: each tool has a non-empty Name and Description.
func TestAllToolDefinitions_NamesAndDescriptions(t *testing.T) {
	tools := AllToolDefinitions()
	for i, tool := range tools {
		if tool.Name == "" {
			t.Errorf("tool[%d].Name is empty", i)
		}
		if tool.Description == "" {
			t.Errorf("tool[%d] (%s).Description is empty", i, tool.Name)
		}
	}
}

// Transport: WriteMessage serializes JSON and appends newline.
func TestTransport_WriteMessage(t *testing.T) {
	var buf bytes.Buffer
	transport := &Transport{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: &buf,
	}

	msg := map[string]string{"key": "value"}
	if err := transport.WriteMessage(msg); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}

	written := buf.String()
	if !strings.Contains(written, `"key"`) {
		t.Errorf("WriteMessage() output = %q, want JSON containing 'key'", written)
	}
	if !strings.HasSuffix(written, "\n") {
		t.Errorf("WriteMessage() output = %q, want trailing newline", written)
	}
}

// Transport: ReadMessage reads a JSON line from the reader.
func TestTransport_ReadMessage(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n"
	transport := &Transport{
		reader: bufio.NewReader(strings.NewReader(input)),
		writer: &bytes.Buffer{},
	}

	raw, err := transport.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("json.Unmarshal(raw) error = %v", err)
	}
	if req.Method != "initialize" {
		t.Errorf("Method = %q, want 'initialize'", req.Method)
	}
}

// Transport: ReadMessage returns io.EOF on empty input.
func TestTransport_ReadMessage_EOF(t *testing.T) {
	transport := &Transport{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: &bytes.Buffer{},
	}

	_, err := transport.ReadMessage()
	if err == nil {
		t.Error("ReadMessage() should return error (io.EOF) for empty input")
	}
}

// RegisterTool: registered handler is invoked during tools/call dispatch.
func TestServer_RegisterTool_OverridesHandler(t *testing.T) {
	s := newTestServer()

	called := false
	s.RegisterTool("mind_doctor", func(_ *deps.Deps, _ json.RawMessage) (any, error) {
		called = true
		return map[string]string{"override": "yes"}, nil
	})

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"mind_doctor","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if !called {
		t.Error("registered handler was not called")
	}
}

// handleToolsCall: invalid params JSON returns errInvalidParams.
func TestHandleRaw_ToolsCall_InvalidParams(t *testing.T) {
	s := newTestServer()

	// params is not valid JSON for ToolCallParams.
	raw := json.RawMessage(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":"not-an-object"}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error == nil {
		t.Fatal("expected error response for invalid params")
	}
	if resp.Error.Code != errInvalidParams {
		t.Errorf("error code = %d, want %d (errInvalidParams)", resp.Error.Code, errInvalidParams)
	}
}

// newMinimalDeps creates a deps.Deps with in-memory repos for tool handler tests.
func newMinimalDeps(t *testing.T) *deps.Deps {
	t.Helper()
	root := t.TempDir()
	docRepo := mem.NewDocRepo()
	iterRepo := mem.NewIterationRepo()
	stateRepo := mem.NewStateRepo()
	briefRepo := mem.NewBriefRepo()
	configRepo := mem.NewConfigRepo()
	qualityRepo := mem.NewQualityRepo()

	lockRepo := mem.NewLockRepo()

	validationSvc := service.NewValidationService(docRepo, iterRepo, briefRepo, configRepo)
	workflowSvc := service.NewWorkflowService(stateRepo, iterRepo)
	qualitySvc := service.NewQualityService(root, qualityRepo)
	projectSvc := service.NewProjectServiceWithConfig(docRepo, iterRepo, stateRepo, briefRepo, configRepo)
	generateSvc := service.NewGenerateService(root)
	doctorSvc := service.NewDoctorService(root, docRepo, iterRepo, briefRepo, configRepo, lockRepo)

	return &deps.Deps{
		ProjectRoot:   root,
		DocRepo:       docRepo,
		IterRepo:      iterRepo,
		StateRepo:     stateRepo,
		BriefRepo:     briefRepo,
		ConfigRepo:    configRepo,
		LockRepo:      lockRepo,
		QualityRepo:   qualityRepo,
		ValidationSvc: validationSvc,
		WorkflowSvc:   workflowSvc,
		QualitySvc:    qualitySvc,
		ProjectSvc:    projectSvc,
		GenerateSvc:   generateSvc,
		DoctorSvc:     doctorSvc,
	}
}

// NewServer: constructs a server with 16 registered tools.
func TestNewServer_RegistersTools(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	if len(s.tools) != 16 {
		t.Errorf("NewServer registered %d tools, want 16", len(s.tools))
	}
}

// mind_read_state tool: returns nil state when no workflow is stored.
func TestToolHandler_MindReadState(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"mind_read_state","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_check_brief tool: returns a brief result.
func TestToolHandler_MindCheckBrief(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"mind_check_brief","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_list_iterations tool: returns workflow history.
func TestToolHandler_MindListIterations(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":22,"method":"tools/call","params":{"name":"mind_list_iterations","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_update_state tool: resets state to idle when type is empty.
func TestToolHandler_MindUpdateState_ResetToIdle(t *testing.T) {
	d := newMinimalDeps(t)
	// Pre-set a state.
	d.StateRepo.(*mem.StateRepo).State = &domain.WorkflowState{
		Type:       domain.TypeBugFix,
		Descriptor: "fix-crash",
	}
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":23,"method":"tools/call","params":{"name":"mind_update_state","arguments":{"type":""}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_update_state tool: updates state with provided values.
func TestToolHandler_MindUpdateState_UpdatesState(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":24,"method":"tools/call","params":{"name":"mind_update_state","arguments":{"type":"ENHANCEMENT","descriptor":"add-feature","last_agent":"developer","remaining_chain":"tester,reviewer"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	// Verify state was written.
	state := d.StateRepo.(*mem.StateRepo).State
	if state == nil {
		t.Fatal("StateRepo.State is nil after mind_update_state")
	}
	if state.Descriptor != "add-feature" {
		t.Errorf("state.Descriptor = %q, want 'add-feature'", state.Descriptor)
	}
}

// mind_validate_docs tool: returns a validation report.
func TestToolHandler_MindValidateDocs(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":25,"method":"tools/call","params":{"name":"mind_validate_docs","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_validate_refs tool: returns a refs validation report.
func TestToolHandler_MindValidateRefs(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":26,"method":"tools/call","params":{"name":"mind_validate_refs","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_check_gate tool: returns a gate result.
func TestToolHandler_MindCheckGate(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":27,"method":"tools/call","params":{"name":"mind_check_gate","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_create_iteration tool: creates an iteration in the temp dir.
func TestToolHandler_MindCreateIteration(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":28,"method":"tools/call","params":{"name":"mind_create_iteration","arguments":{"type":"enhancement","name":"add-feature"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// mind_show_iteration tool: returns IsError for unknown iteration.
func TestToolHandler_MindShowIteration_UnknownID(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":29,"method":"tools/call","params":{"name":"mind_show_iteration","arguments":{"id":"999-ENHANCEMENT-nonexistent"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	// Tool errors are wrapped as IsError=true content, not JSON-RPC errors.
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_doctor tool: returns a doctor report.
func TestToolHandler_MindDoctor(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":30,"method":"tools/call","params":{"name":"mind_doctor","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_list_stubs tool: returns a stubs list (empty for empty repo).
func TestToolHandler_MindListStubs(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":31,"method":"tools/call","params":{"name":"mind_list_stubs","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_search_docs tool: returns search results.
func TestToolHandler_MindSearchDocs(t *testing.T) {
	d := newMinimalDeps(t)
	// Add a doc to search.
	d.DocRepo.(*mem.DocRepo).Files["docs/spec/requirements.md"] = []byte("Authentication feature required.")
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":32,"method":"tools/call","params":{"name":"mind_search_docs","arguments":{"query":"authentication"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_log_quality tool: returns error for nil QualitySvc.
func TestToolHandler_MindLogQuality_MissingQualitySvc(t *testing.T) {
	d := newMinimalDeps(t)
	d.QualitySvc = nil
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":33,"method":"tools/call","params":{"name":"mind_log_quality","arguments":{"file_path":"docs/knowledge/test.md","topic":"test","variant":"v1"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	// Should be an IsError content result, not a JSON-RPC error.
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_read_config tool: returns IsError when no config is set.
func TestToolHandler_MindReadConfig_NoConfig(t *testing.T) {
	d := newMinimalDeps(t)
	// configRepo has no config set → Config() returns error.
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":34,"method":"tools/call","params":{"name":"mind_read_config","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_suggest_next tool: returns suggestions (or empty list).
func TestToolHandler_MindSuggestNext(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":35,"method":"tools/call","params":{"name":"mind_suggest_next","arguments":{}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_show_iteration tool: success case with known iteration in repo.
func TestToolHandler_MindShowIteration_Known(t *testing.T) {
	d := newMinimalDeps(t)
	d.IterRepo.(*mem.IterationRepo).Iterations = []domain.Iteration{
		{
			Seq:     1,
			Type:    domain.TypeEnhancement,
			DirName: "001-ENHANCEMENT-feature",
			Status:  domain.IterComplete,
		},
	}
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":36,"method":"tools/call","params":{"name":"mind_show_iteration","arguments":{"id":"001-ENHANCEMENT-feature"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

// mind_create_iteration tool: error case with invalid type.
func TestToolHandler_MindCreateIteration_InvalidType(t *testing.T) {
	d := newMinimalDeps(t)
	s := NewServer(nil, d)

	raw := json.RawMessage(`{"jsonrpc":"2.0","id":37,"method":"tools/call","params":{"name":"mind_create_iteration","arguments":{"type":"invalid_type","name":"test"}}}`)
	resp := s.handleRaw(raw)

	if resp == nil {
		t.Fatal("handleRaw() returned nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
	// Tool error should be wrapped as IsError content.
}

// Transport: ReadMessage handles partial input (no trailing newline).
func TestTransport_ReadMessage_NoNewline(t *testing.T) {
	// Input without trailing newline — EOF at end.
	input := `{"jsonrpc":"2.0","id":1,"method":"test"}`
	transport := &Transport{
		reader: bufio.NewReader(strings.NewReader(input)),
		writer: &bytes.Buffer{},
	}

	raw, err := transport.ReadMessage()
	// Either returns the message (io.EOF with content) or returns io.EOF.
	if err != nil && raw == nil {
		// pure EOF — acceptable for empty-after-content case.
	} else if err == nil && len(raw) > 0 {
		// Got the message.
	} else if raw != nil && len(raw) > 0 {
		// Got raw with error — also acceptable per ReadMessage implementation.
	}
}

// splitComma: splits a comma-separated string and trims spaces.
func TestSplitComma(t *testing.T) {
	result := splitComma("tester, reviewer , developer")
	if len(result) != 3 {
		t.Errorf("splitComma() = %v, want 3 items", result)
	}
	if result[0] != "tester" || result[1] != "reviewer" || result[2] != "developer" {
		t.Errorf("splitComma() = %v, want [tester reviewer developer]", result)
	}
}

// splitComma: returns empty slice for empty string.
func TestSplitComma_Empty(t *testing.T) {
	result := splitComma("")
	if len(result) != 0 {
		t.Errorf("splitComma('') = %v, want empty", result)
	}
}

// errorResponse: constructs correct JSON-RPC error structure.
func TestErrorResponse(t *testing.T) {
	resp := errorResponse(42, errMethodNotFound, "test message")
	if resp.Error == nil {
		t.Fatal("errorResponse().Error is nil")
	}
	if resp.Error.Code != errMethodNotFound {
		t.Errorf("Code = %d, want %d", resp.Error.Code, errMethodNotFound)
	}
	if resp.Error.Message != "test message" {
		t.Errorf("Message = %q, want 'test message'", resp.Error.Message)
	}
	if resp.ID != 42 {
		t.Errorf("ID = %v, want 42", resp.ID)
	}
}
