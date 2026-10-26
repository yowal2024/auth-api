package server

import (
	"auth-api/internal/config"
	"auth-api/internal/repo"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

func testFileDownloads(t *testing.T, db *sqlx.DB, cfg *config.Config, handler http.Handler, app int64, token string) {
	t.Helper()
	request := func(method, path string, appID int64, body string, extra map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("X-App-Id", fmt.Sprint(appID))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		for key, value := range extra {
			req.Header.Set(key, value)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(repo.SQL(query), args...); err != nil {
			t.Fatal(err)
		}
	}
	code := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		var out struct{ Code int }
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || out.Code != want {
			t.Fatalf("want %d: %s", want, w.Body.String())
		}
	}
	content := []byte("{\n  \"code\": 0, \"data\": [1, 2], \"number\": 9007199254740993\n}\n")
	name := "download-security.json"
	path := filepath.Join(cfg.Upload.Root, "extend_files", name)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	res, err := db.Exec(repo.SQL("INSERT INTO {extend_files} (appid,name,filename,filepath,filesize,filetype,category,remark,version,download_url,created_at,updated_at) VALUES (?,'security',?,?,?,'application/json','','','','',0,0)"), app, name, path, len(content))
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	defer exec("DELETE FROM {extend_files} WHERE id=?", id)
	public := cfg.Upload.URLPrefix + "/extend_files/" + name
	sdk := fmt.Sprintf("/api/sdk/extend-file/%d/download", id)
	admin := fmt.Sprintf("/api/admin/extend-file/%d/download", id)
	for _, url := range []string{public, sdk, admin} {
		w := request("GET", url, app, "", nil)
		if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), content) {
			t.Fatalf("download bytes changed %s: %d %s", url, w.Code, w.Body.String())
		}
		if w.Header().Get("Content-Type") != "application/octet-stream" || w.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment;") {
			t.Fatalf("unsafe attachment headers: %v", w.Header())
		}
		w = request("GET", url, app, "", map[string]string{"Range": "bytes=0-9"})
		if w.Code != 206 || !bytes.Equal(w.Body.Bytes(), content[:10]) {
			t.Fatalf("range mismatch %s: %d %s", url, w.Code, w.Body.String())
		}
	}
	w := request("HEAD", public, app, "", nil)
	if w.Code != 200 || w.Body.Len() != 0 || w.Header().Get("Content-Length") != fmt.Sprint(len(content)) {
		t.Fatal("HEAD mismatch", w)
	}
	for _, method := range []string{"PUT", "DELETE"} {
		code(request(method, fmt.Sprintf("/api/admin/extend-file/%d", id), app+999999, `{"name":"foreign"}`, nil), 404)
	}
	code(request("GET", admin, app+999999, "", nil), 404)
	// Legacy unassigned files must not appear in every selected application.
	exec("UPDATE {extend_files} SET appid=0,category='unassigned-legacy' WHERE id=?", id)
	for _, url := range []string{sdk, admin} {
		code(request("GET", url, app, "", nil), 404)
	}
	for _, url := range []string{"/api/admin/extend-file/list?category=unassigned-legacy", "/api/admin/extend-file/categories", "/api/sdk/extend-file/list?category=unassigned-legacy"} {
		w := request("GET", url, app, "", nil)
		code(w, 0)
		if strings.Contains(w.Body.String(), "unassigned-legacy") || strings.Contains(w.Body.String(), name) {
			t.Fatalf("legacy global file leaked into application: %s", w.Body.String())
		}
	}
	if w := request("GET", public, app, "", nil); w.Code != 200 || !bytes.Equal(w.Body.Bytes(), content) {
		t.Fatal("legacy public download changed", w.Code)
	}
	exec("UPDATE {extend_files} SET appid=?,category='' WHERE id=?", app, id)
	exec("UPDATE {apps} SET status=0 WHERE id=?", app)
	if request("GET", public, app, "", nil).Code != 404 {
		t.Fatal("disabled app public file still available")
	}
	exec("UPDATE {apps} SET status=1 WHERE id=?", app)
	exec("DELETE FROM {extend_files} WHERE id=?", id)
	if request("GET", public, app, "", nil).Code != 404 {
		t.Fatal("orphan file still available")
	}
	// A leftover record cannot publish files for an application that no longer exists.
	exec("INSERT INTO {extend_files} (appid,name,filename,filepath) VALUES (?,'orphan',?,?)", app+999999, name, path)
	defer exec("DELETE FROM {extend_files} WHERE appid=?", app+999999)
	if request("GET", public, app, "", nil).Code != 404 {
		t.Fatal("deleted application file accessible")
	}

	canonical := filepath.Join(cfg.Clone.CanonicalDir, fmt.Sprint(app), "security.apk")
	if err := os.MkdirAll(filepath.Dir(canonical), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(canonical, []byte("fixture APK bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(canonical)
	url := fmt.Sprintf("%s/canonical/%d/security.apk", cfg.Upload.URLPrefix, app)
	if request("GET", url, app, "", nil).Code != 404 {
		t.Fatal("unpublished APK accessible")
	}
	res, err = db.Exec(repo.SQL("INSERT INTO {versions} (appid,version_name,version_key,is_canonical,canonical_path,status) VALUES (?,'security','security',1,?,1)"), app, canonical)
	if err != nil {
		t.Fatal(err)
	}
	vid, _ := res.LastInsertId()
	defer exec("DELETE FROM {versions} WHERE id=?", vid)
	if request("GET", url, app, "", nil).Code != 200 {
		t.Fatal("published APK unavailable")
	}
	exec("UPDATE {versions} SET status=0 WHERE id=?", vid)
	if request("GET", url, app, "", nil).Code != 404 {
		t.Fatal("disabled APK accessible")
	}
	exec("DELETE FROM {versions} WHERE id=?", vid)
	if request("GET", url, app, "", nil).Code != 404 {
		t.Fatal("deleted APK accessible")
	}
	for _, url := range []string{"/uploads/avatars/", "/uploads/app_logos/", "/uploads/extend_files/", "/uploads/avatars/%2e%2e%2fconfig.yaml", "/uploads/extend_files/%2e%2e%5cconfig.yaml"} {
		if w := request("GET", url, app, "", nil); w.Code != 404 {
			t.Fatalf("directory/traversal: %s => %d", url, w.Code)
		}
	}
}
