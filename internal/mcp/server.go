package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jf-ferraz/mind-cli/internal/deps"
)

// JSON-RPC 2.0 error codes.
const (
	errParse          = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
	errInternal       = -32603
)

// Request is a JSON-RPC 2.0 request object.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response object.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error is a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server dispatches JSON-RPC 2.0 requests to registered tool handlers.
type Server struct {
	transport *Transport
	deps      *deps.Deps
	tools     map[string]ToolHandler
}

// ToolHandler is a function that handles a single MCP tool call.
type ToolHandler func(d *deps.Deps, params json.RawMessage) (any, error)

// NewServer creates an MCP server with the given deps.
func NewServer(transport *Transport, d *deps.Deps) *Server {
	s := &Server{
		transport: transport,
		deps:      d,
		tools:     make(map[string]ToolHandler),
	}
	RegisterTools(s)
	return s
}

// RegisterTool registers a handler for the given tool name.
func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.tools[name] = handler
}

// Run enters the main request-dispatch loop. Blocks until stdin closes.
func (s *Server) Run() error {
	for {
		raw, err := s.transport.ReadMessage()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}
		if len(raw) == 0 {
			continue
		}

		resp := s.handleRaw(raw)
		if resp != nil {
			if err := s.transport.WriteMessage(resp); err != nil {
				return fmt.Errorf("write response: %w", err)
			}
		}
	}
}

func (s *Server) handleRaw(raw json.RawMessage) *Response {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return errorResponse(nil, errParse, "parse error: "+err.Error())
	}

	if req.JSONRPC != "2.0" {
		return errorResponse(req.ID, errInvalidRequest, "jsonrpc must be \"2.0\"")
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		// Per JSON-RPC 2.0, notifications have no id field — must not send a response.
		if strings.HasPrefix(req.Method, "notifications/") {
			return nil
		}
		return errorResponse(req.ID, errMethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
	}
}

// InitializeParams holds the client's initialize request params.
type InitializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// InitializeResult is sent back on initialize.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}

// ServerInfo identifies this MCP server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities declares what this server supports.
type Capabilities struct {
	Tools *struct{} `json:"tools,omitempty"`
}

func (s *Server) handleInitialize(req Request) *Response {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "mind",
			Version: "1.0.0",
		},
		Capabilities: Capabilities{
			Tools: &struct{}{},
		},
	}
	return &Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

// ToolsListResult contains the tool definitions.
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

func (s *Server) handleToolsList(req Request) *Response {
	result := ToolsListResult{Tools: AllToolDefinitions()}
	return &Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

// ToolCallParams holds the params for a tools/call request.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolCallResult wraps tool output in the MCP content array format.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock is a single piece of tool output.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *Server) handleToolsCall(req Request) *Response {
	var params ToolCallParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errorResponse(req.ID, errInvalidParams, "invalid params: "+err.Error())
		}
	}

	handler, ok := s.tools[params.Name]
	if !ok {
		return errorResponse(req.ID, errMethodNotFound, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	output, err := handler(s.deps, params.Arguments)
	if err != nil {
		result := ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			IsError: true,
		}
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: result}
	}

	// Serialize output to JSON text
	text, merr := json.Marshal(output)
	if merr != nil {
		text = []byte(fmt.Sprintf("%v", output))
	}

	result := ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}
	return &Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func errorResponse(id any, code int, message string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &Error{Code: code, Message: message},
	}
}
