package pragoai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mcpTestInput struct {
	Query string `json:"query" ai:"Hledaný text"`
	Limit int64  `json:"limit,omitempty"`
}

func newMCPTestServer(t *testing.T) *MCPServer {
	t.Helper()
	server := NewMCPServer("test", "1.0")
	server.Instructions = "instructions"

	echo, err := NewMCPTool("echo", "Echo", func(ctx context.Context, input mcpTestInput) (any, error) {
		return input, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	server.AddTool(echo.ReadOnly())

	text, err := NewMCPTool("text", "Text", func(ctx context.Context, input struct{}) (any, error) {
		return "hello", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	server.AddTool(text)

	failing, err := NewMCPTool("failing", "Failing", func(ctx context.Context, input struct{}) (any, error) {
		return nil, errors.New("something failed")
	})
	if err != nil {
		t.Fatal(err)
	}
	server.AddTool(failing)

	panicking, err := NewMCPTool("panicking", "Panicking", func(ctx context.Context, input struct{}) (any, error) {
		panic("boom")
	})
	if err != nil {
		t.Fatal(err)
	}
	server.AddTool(panicking)

	return server
}

func mcpTestCall(t *testing.T, server *MCPServer, body string) (int, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Body.Len() == 0 {
		return recorder.Code, nil
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("can't parse response %q: %s", recorder.Body.String(), err)
	}
	return recorder.Code, response
}

func mcpTestResult(t *testing.T, response map[string]any) map[string]any {
	t.Helper()
	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatalf("response has no result: %v", response)
	}
	return result
}

func mcpTestErrorCode(t *testing.T, response map[string]any) float64 {
	t.Helper()
	rpcErr, ok := response["error"].(map[string]any)
	if !ok {
		t.Fatalf("response has no error: %v", response)
	}
	return rpcErr["code"].(float64)
}

func mcpTestText(t *testing.T, result map[string]any) string {
	t.Helper()
	content := result["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func TestMCPInitialize(t *testing.T) {
	server := newMCPTestServer(t)

	code, response := mcpTestCall(t, server, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`)
	if code != http.StatusOK {
		t.Fatalf("unexpected status %d", code)
	}
	if response["id"].(float64) != 1 {
		t.Fatalf("unexpected id %v", response["id"])
	}
	result := mcpTestResult(t, response)
	if result["protocolVersion"] != "2025-06-18" {
		t.Fatalf("unexpected protocol version %v", result["protocolVersion"])
	}
	if result["instructions"] != "instructions" {
		t.Fatalf("unexpected instructions %v", result["instructions"])
	}
	if result["serverInfo"].(map[string]any)["name"] != "test" {
		t.Fatalf("unexpected server info %v", result["serverInfo"])
	}

	_, response = mcpTestCall(t, server, `{"jsonrpc":"2.0","id":"a","method":"initialize","params":{"protocolVersion":"1999-01-01"}}`)
	if response["id"] != "a" {
		t.Fatalf("unexpected id %v", response["id"])
	}
	if version := mcpTestResult(t, response)["protocolVersion"]; version != mcpProtocolVersions[0] {
		t.Fatalf("unexpected protocol version %v", version)
	}
}

func TestMCPListTools(t *testing.T) {
	server := newMCPTestServer(t)

	_, response := mcpTestCall(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools := mcpTestResult(t, response)["tools"].([]any)
	if len(tools) != 4 {
		t.Fatalf("unexpected tools count %d", len(tools))
	}

	echo := tools[0].(map[string]any)
	if echo["name"] != "echo" {
		t.Fatalf("unexpected tool name %v", echo["name"])
	}
	if echo["annotations"].(map[string]any)["readOnlyHint"] != true {
		t.Fatalf("echo should be read only: %v", echo)
	}
	schema := echo["inputSchema"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if properties["query"].(map[string]any)["description"] != "Hledaný text" {
		t.Fatalf("unexpected query schema %v", properties["query"])
	}
	required := schema["required"].([]any)
	if len(required) != 1 || required[0] != "query" {
		t.Fatalf("unexpected required %v", required)
	}

	text := tools[1].(map[string]any)
	if _, ok := text["annotations"]; ok {
		t.Fatalf("text should not have annotations: %v", text)
	}
	if _, ok := text["inputSchema"].(map[string]any)["required"]; ok {
		t.Fatalf("empty struct schema should not have required: %v", text["inputSchema"])
	}
}

func TestMCPCallTool(t *testing.T) {
	server := newMCPTestServer(t)

	_, response := mcpTestCall(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"query":"abc","limit":5}}}`)
	result := mcpTestResult(t, response)
	if result["isError"] != nil {
		t.Fatalf("unexpected error result %v", result)
	}
	structured := result["structuredContent"].(map[string]any)
	if structured["query"] != "abc" || structured["limit"].(float64) != 5 {
		t.Fatalf("unexpected structured content %v", structured)
	}
	if text := mcpTestText(t, result); text != `{"query":"abc","limit":5}` {
		t.Fatalf("unexpected text %q", text)
	}

	_, response = mcpTestCall(t, server, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"text"}}`)
	result = mcpTestResult(t, response)
	if text := mcpTestText(t, result); text != "hello" {
		t.Fatalf("unexpected text %q", text)
	}
	if _, ok := result["structuredContent"]; ok {
		t.Fatalf("text result should not have structured content: %v", result)
	}

	for _, tc := range []struct{ name, arguments, text string }{
		{"failing", `{}`, "something failed"},
		{"panicking", `{}`, "boom"},
		{"echo", `{"query":5}`, "invalid arguments"},
	} {
		_, response = mcpTestCall(t, server, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"`+tc.name+`","arguments":`+tc.arguments+`}}`)
		result = mcpTestResult(t, response)
		if result["isError"] != true {
			t.Fatalf("%s: expected error result, got %v", tc.name, result)
		}
		if text := mcpTestText(t, result); !strings.Contains(text, tc.text) {
			t.Fatalf("%s: unexpected error text %q", tc.name, text)
		}
	}

	_, response = mcpTestCall(t, server, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown"}}`)
	if code := mcpTestErrorCode(t, response); code != mcpErrorInvalidParams {
		t.Fatalf("unexpected error code %v", code)
	}
}

func TestMCPProtocol(t *testing.T) {
	server := newMCPTestServer(t)

	_, response := mcpTestCall(t, server, `{"jsonrpc":"2.0","id":1,"method":"ping"}`)
	if len(mcpTestResult(t, response)) != 0 {
		t.Fatalf("unexpected ping result %v", response)
	}

	_, response = mcpTestCall(t, server, `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`)
	if code := mcpTestErrorCode(t, response); code != mcpErrorMethodNotFound {
		t.Fatalf("unexpected error code %v", code)
	}

	_, response = mcpTestCall(t, server, `{not json`)
	if code := mcpTestErrorCode(t, response); code != mcpErrorParse {
		t.Fatalf("unexpected error code %v", code)
	}
	if response["id"] != nil {
		t.Fatalf("parse error should have null id: %v", response)
	}

	_, response = mcpTestCall(t, server, `[{"jsonrpc":"2.0","id":1,"method":"ping"}]`)
	if code := mcpTestErrorCode(t, response); code != mcpErrorInvalidRequest {
		t.Fatalf("unexpected error code %v", code)
	}

	code, response := mcpTestCall(t, server, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if code != http.StatusAccepted || response != nil {
		t.Fatalf("unexpected notification response %d %v", code, response)
	}

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected GET status %d", recorder.Code)
	}
}

func TestNewMCPToolRequiresStruct(t *testing.T) {
	_, err := NewMCPTool("bad", "Bad", func(ctx context.Context, input string) (any, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("expected error for non-struct input")
	}
}
