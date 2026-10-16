package prago

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPEndpoint(t *testing.T) {
	resource := prepareResource(t)
	app := resource.app

	itemA := &ResourceStruct{Name: "A", Count: 1}
	Must(app.Create(itemA))
	Must(app.Create(&ResourceStruct{Name: "B", Count: 2}))

	admin := &user{Name: "Admin", Email: "admin@example.com", Role: sysadminRoleName}
	Must(app.Create(admin))
	adminKey := app.createSessionKey(admin, true)
	adminCookieKey := app.createSessionKey(admin, false)

	basic := &user{Name: "Basic", Email: "basic@example.com"}
	Must(app.Create(basic))
	basicKey := app.createSessionKey(basic, true)

	call := func(t *testing.T, header http.Header, cookieKey string, body string) (int, map[string]any) {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/admin/api/mcp", strings.NewReader(body))
		request.Header = header
		request.Header.Set("Content-Type", "application/json")
		if cookieKey != "" {
			request.AddCookie(&http.Cookie{Name: app.getLoginCookieID(), Value: cookieKey})
		}
		recorder := httptest.NewRecorder()
		(&server{app}).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			return recorder.Code, nil
		}
		var response map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("can't parse response %q: %s", recorder.Body.String(), err)
		}
		return recorder.Code, response
	}

	apiKeyHeader := func(key string) http.Header {
		header := http.Header{}
		header.Set(apiHTTPHeader, key)
		return header
	}

	callTool := func(t *testing.T, key, name, arguments string) map[string]any {
		t.Helper()
		code, response := call(t, apiKeyHeader(key), "", fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, name, arguments))
		if code != http.StatusOK {
			t.Fatalf("unexpected status %d", code)
		}
		result, ok := response["result"].(map[string]any)
		if !ok {
			t.Fatalf("response has no result: %v", response)
		}
		return result
	}

	toolsListBody := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`

	t.Run("authentication", func(t *testing.T) {
		if code, _ := call(t, http.Header{}, "", toolsListBody); code != http.StatusUnauthorized {
			t.Fatalf("request without key: unexpected status %d", code)
		}
		if code, _ := call(t, http.Header{}, adminCookieKey, toolsListBody); code != http.StatusUnauthorized {
			t.Fatalf("request with cookie session: unexpected status %d", code)
		}
		if code, _ := call(t, apiKeyHeader("PAK_invalid"), "", toolsListBody); code != http.StatusUnauthorized {
			t.Fatalf("request with invalid key: unexpected status %d", code)
		}

		code, response := call(t, http.Header{"Authorization": []string{"Bearer " + adminKey}}, "", toolsListBody)
		if code != http.StatusOK {
			t.Fatalf("request with bearer key: unexpected status %d", code)
		}
		tools := response["result"].(map[string]any)["tools"].([]any)
		var names []string
		for _, tool := range tools {
			names = append(names, tool.(map[string]any)["name"].(string))
		}
		if strings.Join(names, ",") != "search,list_resources,list_items,get_item" {
			t.Fatalf("unexpected tools %v", names)
		}
	})

	t.Run("list_resources", func(t *testing.T) {
		result := callTool(t, adminKey, "list_resources", `{}`)
		resources := map[string]map[string]any{}
		for _, resource := range result["structuredContent"].(map[string]any)["resources"].([]any) {
			resources[resource.(map[string]any)["id"].(string)] = resource.(map[string]any)
		}
		if resources[resource.id] == nil {
			t.Fatalf("missing resource %s in %v", resource.id, resources)
		}
		if resources[app.sessionsResource.id] != nil {
			t.Fatal("sessions should not be listed")
		}

		result = callTool(t, basicKey, "list_resources", `{}`)
		for _, item := range result["structuredContent"].(map[string]any)["resources"].([]any) {
			if item.(map[string]any)["id"] == resource.id {
				t.Fatalf("basic user should not see %s", resource.id)
			}
		}
	})

	t.Run("list_items", func(t *testing.T) {
		result := callTool(t, adminKey, "list_items", fmt.Sprintf(`{"resource":%q,"order":"count","order_desc":true,"limit":1}`, resource.id))
		content := result["structuredContent"].(map[string]any)
		if content["total"].(float64) != 2 {
			t.Fatalf("unexpected total %v", content["total"])
		}
		items := content["items"].([]any)
		if len(items) != 1 || items[0].(map[string]any)["name"] != "B" {
			t.Fatalf("unexpected items %v", items)
		}

		result = callTool(t, adminKey, "list_items", fmt.Sprintf(`{"resource":%q,"filters":[{"field":"count","value":"1"}]}`, resource.id))
		content = result["structuredContent"].(map[string]any)
		items = content["items"].([]any)
		if content["total"].(float64) != 1 || len(items) != 1 || items[0].(map[string]any)["name"] != "A" {
			t.Fatalf("unexpected filtered result %v", content)
		}

		for _, arguments := range []string{
			fmt.Sprintf(`{"resource":%q}`, app.sessionsResource.id),
			fmt.Sprintf(`{"resource":%q,"order":"nonexistent"}`, resource.id),
			fmt.Sprintf(`{"resource":%q,"filters":[{"field":"count","value":"abc"}]}`, resource.id),
		} {
			if result := callTool(t, adminKey, "list_items", arguments); result["isError"] != true {
				t.Fatalf("%s: expected error, got %v", arguments, result)
			}
		}
	})

	t.Run("get_item", func(t *testing.T) {
		result := callTool(t, adminKey, "get_item", fmt.Sprintf(`{"resource":%q,"id":%d}`, resource.id, itemA.ID))
		content := result["structuredContent"].(map[string]any)
		if content["name"] != "A" || content["fields"].(map[string]any)["count"].(float64) != 1 {
			t.Fatalf("unexpected item %v", content)
		}
		if !strings.HasPrefix(content["url"].(string), "http") {
			t.Fatalf("url should be absolute: %v", content["url"])
		}

		if result := callTool(t, basicKey, "get_item", fmt.Sprintf(`{"resource":%q,"id":%d}`, resource.id, itemA.ID)); result["isError"] != true {
			t.Fatalf("basic user should not get item: %v", result)
		}
		if result := callTool(t, adminKey, "get_item", fmt.Sprintf(`{"resource":%q,"id":999999}`, resource.id)); result["isError"] != true {
			t.Fatalf("expected not found error: %v", result)
		}
	})

	t.Run("info page", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/admin/_aimcp", nil)
		request.AddCookie(&http.Cookie{Name: app.getLoginCookieID(), Value: adminCookieKey})
		recorder := httptest.NewRecorder()
		(&server{app}).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("unexpected status %d", recorder.Code)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "/admin/api/mcp") || !strings.Contains(body, "list_items") {
			t.Fatal("info page should contain endpoint and tools")
		}
	})

	t.Run("search", func(t *testing.T) {
		result := callTool(t, adminKey, "search", `{"query":"resource"}`)
		if result["isError"] == true {
			t.Fatalf("unexpected error %v", result)
		}
		if _, ok := result["structuredContent"].(map[string]any)["hits"]; !ok {
			t.Fatalf("missing hits %v", result)
		}
	})
}
