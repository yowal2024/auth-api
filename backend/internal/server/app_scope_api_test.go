package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Exercise the actual API responses used by the application tables, including
// their original global IDs for edits and the permanent numbers for display.
func testApplicationRecordAPIs(t *testing.T, handler http.Handler, token string) {
	t.Helper()
	send := func(app int64, method, path, contentType string, body []byte) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/api/admin"+path, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-App-Id", fmt.Sprint(app))
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	decode := func(w *httptest.ResponseRecorder) any {
		t.Helper()
		var result struct {
			Code int
			Data any
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Code != 0 {
			t.Fatalf("API failed: %d %s", w.Code, w.Body.String())
		}
		return result.Data
	}
	call := func(app int64, method, path string, payload any) any {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		return decode(send(app, method, path, "application/json", body))
	}
	list := func(app int64, path string) []any {
		t.Helper()
		data := call(app, "GET", path, nil)
		if page, ok := data.(map[string]any); ok {
			return page["list"].([]any)
		}
		return data.([]any)
	}
	apps := []int64{}
	for _, name := range []string{"Number scope A", "Number scope B"} {
		app := int64(call(0, "POST", "/app", map[string]any{"name": name, "type": "user", "mode": "free"}).(map[string]any)["id"].(float64))
		apps = append(apps, app)
		defer call(app, "DELETE", fmt.Sprintf("/app/%d", app), nil)
	}
	paths := []string{"user/group", "kami/group", "notice", "extend", "block", "encrypt", "function", "ver", "user", "kami", "extend-file"}
	globalIDs := map[string][]float64{}
	for _, app := range apps {
		groups := map[string]any{}
		for _, item := range []struct {
			path string
			body map[string]any
		}{
			{"user/group", map[string]any{"name": "Users", "level": 1}},
			{"kami/group", map[string]any{"name": "Cards", "type": "vip", "value": 86400}},
			{"notice", map[string]any{"title": "Notice", "content": "Hello"}},
			{"extend", map[string]any{"key": "scope", "value": "value"}},
			{"block", map[string]any{"type": "mi", "value": "scope-device"}},
			{"encrypt", map[string]any{"name": "Policy", "algorithm": "none", "encoding": "base64"}},
			{"function", map[string]any{"name": "scope", "trigger": "manual", "code": "function handler(ctx){return {ok:true};}"}},
			{"ver", map[string]any{"versionName": "Draft", "versionKey": "stable", "versionNo": "1.0.0"}},
		} {
			saved := call(app, "POST", "/"+item.path, item.body).(map[string]any)
			groups[item.path] = saved["id"]
		}
		call(app, "POST", "/user", map[string]any{"username": "scopeuser", "password": "password123", "groupId": groups["user/group"]})
		call(app, "POST", "/kami/generate", map[string]any{"count": 2, "length": 16, "groupId": groups["kami/group"]})
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "scope.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write([]byte("application scoped file")); err != nil {
			t.Fatal(err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
		decode(send(app, "POST", "/extend-file/upload", writer.FormDataContentType(), body.Bytes()))
		for _, path := range paths {
			rows := list(app, "/"+path+"/list")
			want := 1
			if path == "kami" {
				want = 2
			}
			if len(rows) != want {
				t.Fatalf("%s app %d: got %d rows, want %d", path, app, len(rows), want)
			}
			seen := map[float64]bool{}
			for _, raw := range rows {
				row := raw.(map[string]any)
				number, ok := row["appRecordId"].(float64)
				if !ok || row["appid"] != float64(app) || number < 1 || number > float64(want) || seen[number] {
					t.Fatalf("%s app %d: incorrect scope or local number: %v", path, app, row)
				}
				seen[number] = true
				if number == 1 {
					globalIDs[path] = append(globalIDs[path], row["id"].(float64))
				}
			}
		}
		rows := list(app, "/user/list?searchField=appRecordId&keyword=1")
		if len(rows) != 1 || rows[0].(map[string]any)["username"] != "scopeuser" {
			t.Fatalf("application number user search failed: %v", rows)
		}
		csvResponse := send(app, "GET", "/user/export", "application/json", nil)
		records, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(csvResponse.Body.String(), "\ufeff"))).ReadAll()
		if err != nil || len(records) != 2 || len(records[0]) < 2 || records[0][0] != "应用内编号" || records[0][1] != "系统 ID" || records[1][0] != "1" || records[1][1] != fmt.Sprintf("%.0f", globalIDs["user"][len(globalIDs["user"])-1]) {
			t.Fatalf("export lost display or system ID: %v %v", records, err)
		}
	}
	for path, ids := range globalIDs {
		if len(ids) != 2 || ids[0] == ids[1] {
			t.Fatalf("global IDs overlapped for %s: %v", path, ids)
		}
	}
	firstID := globalIDs["ver"][0]
	call(apps[0], "POST", "/ver", map[string]any{"id": firstID, "versionName": "Updated", "versionKey": "stable", "versionNo": "2.0.0"})
	// Supplying another application's global ID must never delete its record.
	call(apps[1], "DELETE", fmt.Sprintf("/ver/%.0f", firstID), nil)
	rows := list(apps[0], "/ver/list")
	if len(rows) != 1 || rows[0].(map[string]any)["id"] != firstID || rows[0].(map[string]any)["appRecordId"] != float64(1) || rows[0].(map[string]any)["versionName"] != "Updated" {
		t.Fatalf("edit or foreign deletion changed identity: %v", rows)
	}
	call(apps[0], "DELETE", fmt.Sprintf("/ver/%.0f", firstID), nil)
	call(apps[0], "POST", "/ver", map[string]any{"versionName": "Next", "versionKey": "stable", "versionNo": "3.0.0"})
	for i, app := range apps {
		rows = list(app, "/ver/list")
		if len(rows) != 1 || rows[0].(map[string]any)["appRecordId"] != float64(2-i) {
			t.Fatalf("deletion renumbered or reused IDs: %v", rows)
		}
	}
}
