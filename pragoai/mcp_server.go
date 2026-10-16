package pragoai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
)

// mcpProtocolVersions jsou podporované verze MCP protokolu, od nejnovější.
var mcpProtocolVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26"}

const mcpMaxRequestSize = 10 << 20

const (
	mcpErrorParse          = -32700
	mcpErrorInvalidRequest = -32600
	mcpErrorMethodNotFound = -32601
	mcpErrorInvalidParams  = -32602
)

// MCPServer je bezstavový MCP server (transport Streamable HTTP), který
// nabízí jen nástroje (tools) a odpovídá jako application/json, bez SSE.
// Server se dá sestavit pro každý HTTP požadavek zvlášť, např. jen
// s nástroji, ke kterým má přihlášený uživatel přístup.
type MCPServer struct {
	Name         string
	Title        string
	Version      string
	Instructions string

	tools []*MCPTool
}

// MCPTool je nástroj MCP serveru. Vytváří se přes NewMCPTool a dá se
// sdílet mezi více servery.
type MCPTool struct {
	name        string
	description string
	inputSchema map[string]any
	readOnly    bool
	call        func(ctx context.Context, arguments json.RawMessage) (any, error)
}

func NewMCPServer(name, version string) *MCPServer {
	return &MCPServer{
		Name:    name,
		Version: version,
	}
}

// NewMCPTool vytvoří nástroj, jehož vstupem je struktura T. JSON schéma
// vstupu se odvodí z T stejně jako u Get (tagy json, ai a prago-name).
// Handler dostane kontext HTTP požadavku. Když vrátí string, pošle se
// klientovi jako text, jinak jako JSON. Chyba nebo panic z handleru se
// klientovi vrátí jako chyba nástroje (isError).
func NewMCPTool[T any](name, description string, handler func(ctx context.Context, input T) (any, error)) (*MCPTool, error) {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("pragoai: MCP tool %s: input must be a struct, got %s", name, t)
	}
	schema, err := structSchema(t)
	if err != nil {
		return nil, fmt.Errorf("pragoai: MCP tool %s: %w", name, err)
	}

	return &MCPTool{
		name:        name,
		description: description,
		inputSchema: schema,
		call: func(ctx context.Context, arguments json.RawMessage) (any, error) {
			var input T
			if len(arguments) > 0 && !bytes.Equal(arguments, []byte("null")) {
				if err := json.Unmarshal(arguments, &input); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
			}
			return handler(ctx, input)
		},
	}, nil
}

func (tool *MCPTool) Name() string        { return tool.name }
func (tool *MCPTool) Description() string { return tool.description }

// ReadOnly označí nástroj, že nic nemění (anotace readOnlyHint).
func (tool *MCPTool) ReadOnly() *MCPTool {
	tool.readOnly = true
	return tool
}

func (server *MCPServer) AddTool(tool *MCPTool) *MCPServer {
	server.tools = append(server.tools, tool)
	return server
}

func (server *MCPServer) getTool(name string) *MCPTool {
	for _, tool := range server.tools {
		if tool.name == name {
			return tool
		}
	}
	return nil
}

type mcpMessage struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (server *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, mcpMaxRequestSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body = bytes.TrimSpace(body)
	if len(body) > 0 && body[0] == '[' {
		writeMCPResponse(w, nil, nil, &mcpError{Code: mcpErrorInvalidRequest, Message: "batch requests are not supported"})
		return
	}

	var message mcpMessage
	if err := json.Unmarshal(body, &message); err != nil {
		writeMCPResponse(w, nil, nil, &mcpError{Code: mcpErrorParse, Message: err.Error()})
		return
	}

	// notifikace a odpovědi klienta nevyžadují odpověď
	if len(message.ID) == 0 || bytes.Equal(message.ID, []byte("null")) || message.Method == "" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	result, rpcErr := server.handle(r.Context(), message.Method, message.Params)
	writeMCPResponse(w, message.ID, result, rpcErr)
}

func writeMCPResponse(w http.ResponseWriter, id json.RawMessage, result any, rpcErr *mcpError) {
	if id == nil {
		id = json.RawMessage("null")
	}
	data, err := json.Marshal(mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   rpcErr,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (server *MCPServer) handle(ctx context.Context, method string, params json.RawMessage) (any, *mcpError) {
	switch method {
	case "initialize":
		return server.initialize(params)
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return server.listTools(), nil
	case "tools/call":
		return server.callTool(ctx, params)
	default:
		return nil, &mcpError{Code: mcpErrorMethodNotFound, Message: "method not found: " + method}
	}
}

func (server *MCPServer) initialize(params json.RawMessage) (any, *mcpError) {
	var request struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, &mcpError{Code: mcpErrorInvalidParams, Message: err.Error()}
		}
	}

	protocolVersion := mcpProtocolVersions[0]
	if slices.Contains(mcpProtocolVersions, request.ProtocolVersion) {
		protocolVersion = request.ProtocolVersion
	}

	serverInfo := map[string]any{
		"name":    server.Name,
		"version": server.Version,
	}
	if server.Title != "" {
		serverInfo["title"] = server.Title
	}

	ret := map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities": map[string]any{
			"tools": map[string]any{"listChanged": false},
		},
		"serverInfo": serverInfo,
	}
	if server.Instructions != "" {
		ret["instructions"] = server.Instructions
	}
	return ret, nil
}

func (server *MCPServer) listTools() any {
	tools := []map[string]any{}
	for _, tool := range server.tools {
		item := map[string]any{
			"name":        tool.name,
			"description": tool.description,
			"inputSchema": tool.inputSchema,
		}
		if tool.readOnly {
			item["annotations"] = map[string]any{"readOnlyHint": true}
		}
		tools = append(tools, item)
	}
	return map[string]any{"tools": tools}
}

func (server *MCPServer) callTool(ctx context.Context, params json.RawMessage) (any, *mcpError) {
	var request struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, &mcpError{Code: mcpErrorInvalidParams, Message: err.Error()}
	}

	tool := server.getTool(request.Name)
	if tool == nil {
		return nil, &mcpError{Code: mcpErrorInvalidParams, Message: "unknown tool: " + request.Name}
	}

	value, err := tool.execute(ctx, request.Arguments)
	if err != nil {
		return mcpTextResult(err.Error(), true), nil
	}

	if text, ok := value.(string); ok {
		return mcpTextResult(text, false), nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return mcpTextResult(err.Error(), true), nil
	}
	ret := mcpTextResult(string(data), false)
	if len(data) > 0 && data[0] == '{' {
		ret["structuredContent"] = json.RawMessage(data)
	}
	return ret, nil
}

func (tool *MCPTool) execute(ctx context.Context, arguments json.RawMessage) (value any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	return tool.call(ctx, arguments)
}

func mcpTextResult(text string, isError bool) map[string]any {
	ret := map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
	if isError {
		ret["isError"] = true
	}
	return ret
}
