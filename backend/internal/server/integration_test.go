package server

import (
	"auth-api/internal/auth"
	"auth-api/internal/buildpool/bsdiff"
	"auth-api/internal/config"
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/internal/sdkcrypto"
	"auth-api/internal/tlsmgr"
	"auth-api/migrations"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Uses a newly created, uniquely named database. Never migrates an existing schema.
func TestRealDatabaseLifecycle(t *testing.T) {
	dsn := os.Getenv("AUTH_RECOVERY_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set AUTH_RECOVERY_MYSQL_DSN for isolated integration test")
	}
	mc, e := mysql.ParseDSN(dsn)
	if e != nil {
		t.Fatal(e)
	}
	mc.DBName = ""
	root, e := sqlx.Connect("mysql", mc.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	name := fmt.Sprintf("recovery_test_%d", time.Now().UnixNano())
	if _, e = root.Exec("CREATE DATABASE `" + name + "` DEFAULT CHARACTER SET utf8mb4"); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if _, e := root.Exec("DROP DATABASE `" + name + "`"); e != nil {
			t.Error(e)
		}
	}()
	mc.DBName = name
	mc.ParseTime = true
	m := install.Migrator{DSN: mc.FormatDSN(), Prefix: "a_", Files: migrations.Files}
	applied, e := m.Run(context.Background())
	if e != nil {
		t.Fatalf("migrations: %v", e)
	}
	if len(applied) != 44 {
		t.Fatalf("migrations=%d", len(applied))
	}
	if again, e := m.Run(context.Background()); e != nil || len(again) != 0 {
		t.Fatalf("idempotent migration: %v %v", again, e)
	}
	d, e := sqlx.Connect("mysql", mc.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	if e = repo.InitTablePrefix("a_"); e != nil {
		t.Fatal(e)
	}
	hash, e := auth.HashPassword("recovery-pass-123")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = repo.NewAdminRepo(d).Create(context.Background(), "admin", hash, "super"); e != nil {
		t.Fatal(e)
	}
	redisAddr := os.Getenv("AUTH_RECOVERY_REDIS_ADDR")
	options := &redis.Options{Addr: redisAddr, Password: os.Getenv("AUTH_RECOVERY_REDIS_PASSWORD")}
	if redisAddr == "" {
		options.Addr = miniredis.RunT(t).Addr()
	} else {
		// Real Redis tests must explicitly target a disposable dedicated instance.
		if os.Getenv("AUTH_RECOVERY_REDIS_DISPOSABLE") != "1" {
			t.Fatal("set AUTH_RECOVERY_REDIS_DISPOSABLE=1 for isolated Redis")
		}
		t.Log("using real Redis")
	}
	r := redis.NewClient(options)
	defer r.Close()
	if err := r.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}
	if redisAddr != "" {
		if err := r.FlushDB(context.Background()).Err(); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := r.FlushDB(context.Background()).Err(); err != nil {
				t.Error(err)
			}
		}()
	}
	temp := t.TempDir()
	t.Chdir(temp)
	cfg := &config.Config{Installed: true, Session: config.SessionConfig{TTLSec: 3600}, Upload: config.UploadConfig{Root: filepath.Join(temp, "uploads"), URLPrefix: "/uploads"}, Clone: config.CloneConfig{CanonicalDir: filepath.Join(temp, "custom-canonical"), OutputDir: filepath.Join(temp, "diffs"), DiffTTLSec: 60}}
	installer := install.NewHandler(filepath.Join(temp, "config.yml"), cfg, migrations.Files)
	wrapper := &RewriteHandler{DB: d}
	https := NewListenerManager("127.0.0.1:0", wrapper)
	defer https.Close()
	g, e := New(Options{Config: cfg, DB: d, Redis: r, Installer: installer, CanonicalAdmin: true, OnChange: func() error { return wrapper.Refresh(context.Background()) }, OnTLS: func() error { return https.Reload(d) }})
	if e != nil {
		t.Fatal(e)
	}
	wrapper.Next = g
	if e = https.Reload(d); e != nil {
		t.Fatal(e)
	}
	adminToken := ""
	key := ""
	appID := int64(0)
	seq := 0
	remoteIP := "192.0.2.1"
	request := func(method, path string, payload any, token string, sdk bool) map[string]any {
		t.Helper()
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(method, path, bytes.NewReader(b))
		req.RemoteAddr = net.JoinHostPort(remoteIP, "12345")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if appID > 0 {
			req.Header.Set("X-App-Id", strconv.FormatInt(appID, 10))
		}
		if sdk {
			seq++
			ts := strconv.FormatInt(time.Now().Unix(), 10)
			nonce := fmt.Sprint(seq)
			req.Header.Set("X-Ts", ts)
			req.Header.Set("X-Nonce", nonce)
			req.Header.Set("X-Sign", sdkcrypto.SignBytes(key, ts, nonce, b))
			req.Header.Set("X-Mi", "test-device")
		}
		w := httptest.NewRecorder()
		wrapper.ServeHTTP(w, req)
		var v map[string]any
		if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil {
			t.Fatalf("%s %s HTTP%d: %s", method, path, w.Code, w.Body.String())
		}
		t.Logf("%s %s code=%v", method, path, v["code"])
		return v
	}
	ok := func(v map[string]any) any {
		t.Helper()
		if v["code"] != float64(0) {
			t.Fatalf("API failure: %v", v)
		}
		return v["data"]
	}
	adminToken = ok(request("POST", "/api/admin/login", map[string]any{"username": "admin", "password": "recovery-pass-123"}, "", false)).(map[string]any)["token"].(string)
	app := ok(request("POST", "/api/admin/app", map[string]any{"name": "Integration", "type": "user", "mode": "free", "bindDevice": true}, adminToken, false)).(map[string]any)
	appID = int64(app["id"].(float64))
	key = app["tokenKey"].(string)
	t.Run("OrdinaryAdminOperationsAndConfigurationBoundary", func(t *testing.T) {
		created := ok(request("POST", "/api/admin/admin", map[string]any{"username": "operator1", "password": "test-pass-123", "role": "admin"}, adminToken, false)).(map[string]any)
		id := created["id"].(float64)
		defer func() { ok(request("DELETE", fmt.Sprintf("/api/admin/admin/%.0f", id), nil, adminToken, false)) }()
		login := ok(request("POST", "/api/admin/login", map[string]any{"username": "operator1", "password": "test-pass-123"}, "", false)).(map[string]any)
		operatorToken := login["token"].(string)
		apps := ok(request("GET", "/api/admin/app/list", nil, operatorToken, false)).([]any)
		for _, value := range apps {
			item := value.(map[string]any)
			for _, field := range []string{"tokenKey", "geoKey", "encryption", "userTokenTtl"} {
				if _, exists := item[field]; exists {
					t.Fatalf("operator received %s", field)
				}
			}
		}
		for _, endpoint := range []struct{ method, path string }{
			{"POST", "/app"}, {"PUT", fmt.Sprintf("/app/%d", appID)}, {"DELETE", fmt.Sprintf("/app/%d", appID)},
			{"POST", fmt.Sprintf("/app/%d/mtls/issue-cert", appID)}, {"POST", fmt.Sprintf("/app/%d/mtls/gen-ca", appID)},
			{"GET", "/encrypt/list"}, {"POST", "/encrypt"}, {"POST", "/extend"}, {"GET", "/extend-file/list"},
			{"POST", "/extend-file/upload"}, {"GET", "/ver/list"}, {"POST", "/function/0/run"},
			{"GET", "/system/settings"}, {"GET", "/admin/list"}, {"GET", "/log/list"}, {"POST", "/log/clear"},
		} {
			if v := request(endpoint.method, "/api/admin"+endpoint.path, map[string]any{}, operatorToken, false); v["code"] != float64(403) {
				t.Fatal("configuration accessible", endpoint, v)
			}
		}
		for _, path := range []string{"/user/list", "/kami/list", "/block/list", "/notice/list", "/log/list?category=user", "/dashboard"} {
			ok(request("GET", "/api/admin"+path, nil, operatorToken, false))
		}
		block := ok(request("POST", "/api/admin/block", map[string]any{"type": "ip", "value": "203.0.113.42", "reason": "permission test"}, operatorToken, false)).(map[string]any)
		ok(request("DELETE", fmt.Sprintf("/api/admin/block/%.0f", block["id"]), nil, operatorToken, false))
		// Existing tokens cannot retain elevated privileges after demotion.
		if _, err := d.Exec(repo.SQL("UPDATE {admins} SET role='super' WHERE id=?"), id); err != nil {
			t.Fatal(err)
		}
		ok(request("GET", "/api/admin/system/settings", nil, operatorToken, false))
		if _, err := d.Exec(repo.SQL("UPDATE {admins} SET role='admin' WHERE id=?"), id); err != nil {
			t.Fatal(err)
		}
		if v := request("GET", "/api/admin/system/settings", nil, operatorToken, false); v["code"] != float64(403) {
			t.Fatal("demoted token retained access", v)
		}
	})
	// Policy inheritance must remap app-scoped encryption rather than link across apps.
	encryption := ok(request("POST", "/api/admin/encrypt", map[string]any{"name": "Inherited", "algorithm": "aes", "encoding": "base64", "symKey": "1234567890abcdef", "timeSkew": 60, "signMode": "verify"}, adminToken, false)).(map[string]any)
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET encryption=? WHERE id=?"), encryption["id"], appID); e != nil {
		t.Fatal(e)
	}
	inherited := ok(request("POST", "/api/admin/app", map[string]any{"name": "Inherited app", "type": "user", "inheritFrom": appID}, adminToken, false)).(map[string]any)
	if inherited["tokenKey"] == key || inherited["bindDevice"] != true || inherited["encryption"] == encryption["id"] {
		t.Fatalf("inheritance: %v", inherited)
	}
	inheritedID := int64(inherited["id"].(float64))
	policy, err := repo.NewEncryptRepo(d).GetByID(context.Background(), inheritedID, int64(inherited["encryption"].(float64)))
	if err != nil || policy == nil || policy.SymKey != "1234567890abcdef" {
		t.Fatalf("policy scope: %v %v", policy, err)
	}
	ok(request("DELETE", fmt.Sprintf("/api/admin/app/%d", inheritedID), nil, adminToken, false))
	if result := request("DELETE", fmt.Sprintf("/api/admin/encrypt/%.0f", encryption["id"]), nil, adminToken, false); result["code"] != float64(400) {
		t.Fatal("active encryption deletion must be rejected", result)
	}
	if policy, err := repo.NewEncryptRepo(d).GetByID(context.Background(), appID, int64(encryption["id"].(float64))); err != nil || policy == nil {
		t.Fatal("active encryption was removed", err)
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET encryption=0 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	ok(request("DELETE", fmt.Sprintf("/api/admin/encrypt/%.0f", encryption["id"]), nil, adminToken, false))
	if result := request("DELETE", fmt.Sprintf("/api/admin/encrypt/%.0f", encryption["id"]), nil, adminToken, false); result["code"] != float64(404) {
		t.Fatal("missing encryption must return 404", result)
	}
	currentApp, err := repo.NewAppRepo(d).GetByID(context.Background(), appID)
	if err != nil {
		t.Fatal(err)
	}
	currentApp.Encryption = int64(encryption["id"].(float64))
	if err := repo.NewAppRepo(d).Update(context.Background(), currentApp); err != repo.ErrEncryptionUnavailable {
		t.Fatal("deleted policy must not be selected even after a stale handler check", err)
	}

	if _, e := os.Stat("data/master.key"); !os.IsNotExist(e) {
		t.Fatal("master key generated before enabling feature")
	}
	for _, path := range []string{"/me", "/app/list", "/notice/list", "/extend/list", "/user/list", "/user/group/list", "/kami/list", "/kami/group/list", "/encrypt/list", "/function/list", "/extend-file/list", "/extend-file/categories", "/ver/list", "/block/list", "/log/list", "/log/list?category=user", "/log/actions", "/dashboard", "/system/settings", "/system/api-routes", "/system/api-codes"} {
		ok(request("GET", "/api/admin"+path, nil, adminToken, false))
	}
	user := ok(request("POST", "/api/sdk/register", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)
	token := user["token"].(string)
	var registered struct {
		Active int64  `db:"last_active_time"`
		IP     string `db:"last_ip"`
	}
	if err := d.Get(&registered, repo.SQL("SELECT last_active_time,last_ip FROM {users} WHERE id=?"), user["user"].(map[string]any)["id"]); err != nil || registered.Active == 0 || registered.IP == "" {
		t.Fatalf("registration session must record activity before first heartbeat: %+v %v", registered, err)
	}
	if out := request("POST", "/api/sdk/register", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true); out["code"] != float64(1007) {
		t.Fatalf("duplicate registration: %v", out)
	}
	for _, path := range []string{"ini", "info", "heartbeat"} {
		ok(request("POST", "/api/sdk/"+path, map[string]any{}, token, true))
	}
	t.Run("ManagedSessionsAndHeartbeat", func(t *testing.T) {
		uid := int64(user["user"].(map[string]any)["id"].(float64))
		path := fmt.Sprintf("/api/admin/user/%d/sessions", uid)
		rows := ok(request("GET", path, nil, adminToken, false)).([]any)
		if len(rows) != 1 {
			t.Fatalf("sessions: %v", rows)
		}
		row := rows[0].(map[string]any)
		if row["lastHeartbeatAt"].(float64) <= 0 || row["id"] == token {
			t.Fatal("missing heartbeat or token exposed", row)
		}
		var active int64
		if e := d.Get(&active, repo.SQL("SELECT last_active_time FROM {users} WHERE id=?"), uid); e != nil || active == 0 {
			t.Fatal(active, e)
		}
		ok(request("DELETE", path+"/"+row["id"].(string), nil, adminToken, false))
		if v := request("POST", "/api/sdk/heartbeat", map[string]any{}, token, true); v["code"] != float64(1006) {
			t.Fatal("revoked session usable", v)
		}
		token = ok(request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)["token"].(string)
		ok(request("DELETE", path, nil, adminToken, false))
		if v := request("POST", "/api/sdk/heartbeat", map[string]any{}, token, true); v["code"] != float64(1006) {
			t.Fatal("revoke all failed", v)
		}
		token = ok(request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)["token"].(string)
		user["token"] = token
		// Unbinding one device must destroy its tokens, even if device binding is
		// subsequently disabled. A UI list of sessions must not retain those tokens.
		devices := ok(request("GET", fmt.Sprintf("/api/admin/user/%d/devices", uid), nil, adminToken, false)).([]any)
		did := devices[0].(map[string]any)["id"].(float64)
		ok(request("DELETE", fmt.Sprintf("/api/admin/user/%d/devices/%.0f", uid, did), nil, adminToken, false))
		if rows := ok(request("GET", path, nil, adminToken, false)).([]any); len(rows) != 0 {
			t.Fatal("unbound device still in sessions", rows)
		}
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET bind_device=0 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		if v := request("POST", "/api/sdk/heartbeat", map[string]any{}, token, true); v["code"] != float64(1006) {
			t.Fatal("unbound token revived after disabling binding", v)
		}
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET bind_device=1 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		token = ok(request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)["token"].(string)
		user["token"] = token
	})
	ok(request("POST", "/api/sdk/updateInfo", map[string]any{"nickname": "Updated"}, token, true))
	ok(request("POST", "/api/sdk/userExtends", map[string]any{"set": map[string]string{"theme": "dark"}}, token, true))
	t.Run("LegacyDuplicateExtensions", func(t *testing.T) {
		uid := int64(user["user"].(map[string]any)["id"].(float64))
		if _, err := d.Exec(repo.SQL("INSERT INTO {user_extends} (uid,e_key,e_value) VALUES (?,'theme','stale')"), uid); err != nil {
			t.Fatal(err)
		}
		out := ok(request("POST", "/api/sdk/userExtends", map[string]any{"set": map[string]string{"theme": "updated"}}, token, true)).(map[string]any)
		if out["extends"].(map[string]any)["theme"] != "updated" {
			t.Fatal("duplicate shadowed SDK update")
		}
		var n int
		if err := d.Get(&n, repo.SQL("SELECT COUNT(*) FROM {user_extends} WHERE uid=? AND e_key='theme' AND e_value<>'updated'"), uid); err != nil || n != 0 {
			t.Fatalf("stale duplicate: %d %v", n, err)
		}
		if out := request("POST", fmt.Sprintf("/api/admin/user/%d/extends", uid), map[string]any{"key": "theme", "value": "third"}, adminToken, false); out["code"] != float64(400) {
			t.Fatal("duplicate accepted", out)
		}
	})
	login := ok(request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)
	token = login["token"].(string)
	t.Run("SDKSecurity", func(t *testing.T) { testSDKSecurity(t, d, r, wrapper, appID, key, token) })
	t.Run("UsernameAliasesShareBruteForceLimit", func(t *testing.T) {
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET brute_force_enabled=1 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		defer func() {
			d.Exec(repo.SQL("UPDATE {apps} SET brute_force_enabled=0 WHERE id=?"), appID)
			uid := int64(user["user"].(map[string]any)["id"].(float64))
			r.Del(context.Background(), fmt.Sprintf("recovery:brute:uid:%d:%d", appID, uid))
		}()
		for i, name := range []string{"tester1", "TESTER1", "Tester1", "tEstEr1", "tester1 ", "tESteR1"} {
			out := request("POST", "/api/sdk/logon", map[string]any{"username": name, "password": "wrong-password", "mi": "test-device"}, "", true)
			want := float64(1001)
			if i == 5 {
				want = 1017
			}
			if out["code"] != want {
				t.Fatalf("alias %q bypass: %v", name, out)
			}
		}
	})
	listed := ok(request("GET", "/api/admin/user/list", nil, adminToken, false)).(map[string]any)["list"].([]any)
	if len(listed) != 1 || listed[0].(map[string]any)["online"] != true {
		t.Fatalf("online state: %v", listed)
	}
	if v := request("POST", "/api/admin/admin", map[string]any{"id": 1, "role": "admin"}, adminToken, false); v["code"] == float64(0) {
		t.Fatal("last super demoted")
	}
	secondAdmin := ok(request("POST", "/api/admin/admin", map[string]any{"username": "second", "password": "password123", "role": "super"}, adminToken, false)).(map[string]any)
	ok(request("POST", "/api/admin/admin", map[string]any{"id": secondAdmin["id"], "role": "admin"}, adminToken, false))
	ok(request("DELETE", fmt.Sprintf("/api/admin/admin/%.0f", secondAdmin["id"]), nil, adminToken, false))
	deny := ok(request("POST", "/api/admin/function", map[string]any{"name": "deny_login", "trigger": "before_login", "enabled": true, "code": "function handler(ctx){return {allow:false,msg:'denied'};}"}, adminToken, false)).(map[string]any)
	if v := request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true); v["code"] == float64(0) {
		t.Fatal("before trigger failed to reject")
	}
	ok(request("DELETE", fmt.Sprintf("/api/admin/function/%.0f", deny["id"]), nil, adminToken, false))
	after := ok(request("POST", "/api/admin/function", map[string]any{"name": "after_login", "trigger": "after_login", "enabled": true, "code": "function handler(ctx){return {allow:false,msg:'ignored'};}"}, adminToken, false)).(map[string]any)
	extraLogin := ok(request("POST", "/api/sdk/logon", map[string]any{"username": "tester1", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)
	ok(request("DELETE", fmt.Sprintf("/api/admin/function/%.0f", after["id"]), nil, adminToken, false))
	// Exercise writes as well as empty listing queries against the real schema.
	for _, item := range []struct {
		path    string
		payload map[string]any
	}{{"notice", map[string]any{"title": "Notice", "content": "Hello", "enabled": 1}}, {"extend", map[string]any{"key": "test_key", "value": "test_value"}}, {"block", map[string]any{"type": "mi", "value": "blocked-device", "reason": "test"}}, {"user/group", map[string]any{"name": "Test group", "level": 2}}, {"encrypt", map[string]any{"name": "AES", "algorithm": "aes", "encoding": "base64", "symKey": "1234567890abcdef", "timeSkew": 60, "signMode": "verify"}}} {
		saved := ok(request("POST", "/api/admin/"+item.path, item.payload, adminToken, false)).(map[string]any)
		id := saved["id"]
		item.payload["id"] = id
		ok(request("POST", "/api/admin/"+item.path, item.payload, adminToken, false))
		ok(request("DELETE", fmt.Sprintf("/api/admin/%s/%.0f", item.path, id), nil, adminToken, false))
	}
	postFile := func(path, name string, content []byte) map[string]any {
		t.Helper()
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, e := mw.CreateFormFile("file", name)
		if e != nil {
			t.Fatal(e)
		}
		part.Write(content)
		mw.WriteField("name", "Test file")
		mw.Close()
		req := httptest.NewRequest("POST", path, &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("X-App-Id", fmt.Sprint(appID))
		w := httptest.NewRecorder()
		wrapper.ServeHTTP(w, req)
		var out map[string]any
		if e = json.Unmarshal(w.Body.Bytes(), &out); e != nil {
			t.Fatal(w.Body.String())
		}
		return ok(out).(map[string]any)
	}
	file := postFile("/api/admin/extend-file/upload", "sample.txt", []byte("download-check"))
	t.Run("FileDownloadSecurity", func(t *testing.T) { testFileDownloads(t, d, cfg, wrapper, appID, adminToken) })
	fileID := file["id"].(float64)
	publicFile := httptest.NewRecorder()
	wrapper.ServeHTTP(publicFile, httptest.NewRequest("GET", file["downloadUrl"].(string), nil))
	if publicFile.Code != 200 || publicFile.Body.String() != "download-check" {
		t.Fatalf("public download URL: %d %s", publicFile.Code, publicFile.Body.String())
	}
	filePath := fmt.Sprintf("/api/sdk/extend-file/%.0f/download", fileID)
	req := httptest.NewRequest("GET", filePath, nil)
	req.Header.Set("X-App-Id", fmt.Sprint(appID))
	w := httptest.NewRecorder()
	wrapper.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() != "download-check" {
		t.Fatalf("download: %d %s", w.Code, w.Body.String())
	}
	fileList := ok(request("GET", "/api/sdk/extend-file/list", nil, "", false)).([]any)
	for _, item := range fileList {
		if path, ok := item.(map[string]any)["filepath"]; ok && path != "" {
			t.Fatal("server file path leaked", item)
		}
	}
	foreign := ok(request("POST", "/api/admin/app", map[string]any{"name": "Foreign download"}, adminToken, false)).(map[string]any)
	t.Run("LogFiltersAndScope", func(t *testing.T) {
		testLogScope(t, d, &appID, int64(foreign["id"].(float64)), adminToken, request)
	})
	firstApp := appID
	appID = int64(foreign["id"].(float64))
	if v := request("GET", filePath, nil, "", false); v["code"] != float64(404) {
		t.Fatal("cross-app file download", v)
	}
	appID = firstApp
	ok(request("DELETE", fmt.Sprintf("/api/admin/app/%.0f", foreign["id"]), nil, adminToken, false))
	ok(request("DELETE", fmt.Sprintf("/api/admin/extend-file/%.0f", fileID), nil, adminToken, false))
	if sdk := os.Getenv("AUTH_RECOVERY_ANDROID_SDK"); sdk != "" {
		manifest := filepath.Join(temp, "AndroidManifest.xml")
		apk := filepath.Join(temp, "input.apk")
		os.WriteFile(manifest, []byte(`<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.recovery.integration"><uses-sdk android:minSdkVersion="24"/><application android:hasCode="false"/></manifest>`), 0600)
		if out, e := exec.Command(filepath.Join(sdk, "build-tools", "36.0.0", "aapt2.exe"), "link", "--manifest", manifest, "-I", filepath.Join(sdk, "platforms", "android-37.0", "android.jar"), "-o", apk).CombinedOutput(); e != nil {
			t.Fatalf("aapt %s %v", out, e)
		}
		raw, e := os.ReadFile(apk)
		if e != nil {
			t.Fatal(e)
		}
		uploaded := postFile("/api/admin/ver/upload-apk", "test.apk", raw)
		fullAPK := httptest.NewRecorder()
		wrapper.ServeHTTP(fullAPK, httptest.NewRequest("GET", uploaded["url"].(string), nil))
		if fullAPK.Code != 404 {
			t.Fatal("unpublished canonical APK is downloadable")
		}
		if !strings.HasPrefix(uploaded["path"].(string), cfg.Clone.CanonicalDir) {
			t.Fatal("canonical directory ignored")
		}
		other := ok(request("POST", "/api/admin/app", map[string]any{"name": "Other app", "type": "user"}, adminToken, false)).(map[string]any)
		firstApp := appID
		appID = int64(other["id"].(float64))
		if out := request("POST", "/api/admin/ver", map[string]any{"versionName": "bad", "versionKey": "bad", "isCanonical": 1, "canonicalPath": uploaded["path"]}, adminToken, false); out["code"] == float64(0) {
			t.Fatal("cross-app canonical accepted")
		}
		appID = firstApp
		ok(request("DELETE", fmt.Sprintf("/api/admin/app/%.0f", other["id"]), nil, adminToken, false))
		ok(request("POST", "/api/admin/ver", map[string]any{"versionName": "1", "versionKey": "stable", "versionNo": "1", "isCanonical": 1, "canonicalPath": uploaded["path"], "cloneDiffEnable": 1}, adminToken, false))
		fullAPK = httptest.NewRecorder()
		wrapper.ServeHTTP(fullAPK, httptest.NewRequest("GET", uploaded["url"].(string), nil))
		if fullAPK.Code != 200 || !bytes.Equal(raw, fullAPK.Body.Bytes()) {
			t.Fatal("published APK mismatch")
		}
		update := ok(request("POST", "/api/sdk/ini", map[string]any{"versionKey": "old-channel", "versionNo": "1"}, "", true)).(map[string]any)["latestVersion"].(map[string]any)
		if update["hasUpdate"] != true || update["versionKey"] != "stable" {
			t.Fatalf("version key update: %v", update)
		}
		if _, e = d.Exec(repo.SQL("UPDATE {apps} SET random_package=1 WHERE id=?"), appID); e != nil {
			t.Fatal(e)
		}
		cloned := ok(request("GET", "/api/sdk/clone", nil, token, true)).(map[string]any)
		if cloned["newPackage"] == "com.recovery.integration" {
			t.Fatal("package not changed")
		}
		diffReq := httptest.NewRequest("GET", cloned["diffUrl"].(string), nil)
		diffReq.Header.Set("X-App-Id", fmt.Sprint(appID))
		dw := httptest.NewRecorder()
		wrapper.ServeHTTP(dw, diffReq)
		if dw.Code != 200 || !bytes.HasPrefix(dw.Body.Bytes(), []byte("DELTA1")) {
			t.Fatal("diff download failed")
		}
		cloneBytes, err := bsdiff.Apply(raw, dw.Body.Bytes())
		if err != nil || len(cloneBytes) == 0 {
			t.Fatal("downloaded clone patch cannot be applied", err)
		}
		// Exercise version-to-version diff via ini, including actual byte reconstruction.
		newRaw := append([]byte{}, raw...)
		// Change a manifest label by producing another valid APK.
		os.WriteFile(manifest, []byte(`<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.recovery.integration" android:versionCode="2"><uses-sdk android:minSdkVersion="24"/><application android:hasCode="false" android:label="version2"/></manifest>`), 0600)
		if out, err := exec.Command(filepath.Join(sdk, "build-tools", "36.0.0", "aapt2.exe"), "link", "--manifest", manifest, "-I", filepath.Join(sdk, "platforms", "android-37.0", "android.jar"), "-o", apk).CombinedOutput(); err != nil {
			t.Fatalf("second APK: %s %v", out, err)
		}
		newRaw, err = os.ReadFile(apk)
		if err != nil {
			t.Fatal(err)
		}
		newUpload := postFile("/api/admin/ver/upload-apk", "new.apk", newRaw)
		ok(request("POST", "/api/admin/ver", map[string]any{"versionName": "2", "versionKey": "stable", "versionNo": "2", "isCanonical": 1, "canonicalPath": newUpload["path"], "cloneDiffEnable": 1, "apkUrl": newUpload["url"]}, adminToken, false))
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET diff_update=1 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		update = ok(request("POST", "/api/sdk/ini", map[string]any{"versionKey": "stable", "versionNo": "1"}, "", true)).(map[string]any)["latestVersion"].(map[string]any)
		upgradeURL, exists := update["diffUrl"].(string)
		if !exists {
			t.Fatal("upgrade patch missing", update)
		}
		getDiff := func(app int64, url string) *httptest.ResponseRecorder {
			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("X-App-Id", fmt.Sprint(app))
			w := httptest.NewRecorder()
			wrapper.ServeHTTP(w, req)
			return w
		}
		upgrade := getDiff(appID, upgradeURL)
		patched, err := bsdiff.Apply(raw, upgrade.Body.Bytes())
		if err != nil || !bytes.Equal(patched, newRaw) {
			t.Fatal("upgrade patch mismatch", err)
		}
		otherApp := ok(request("POST", "/api/admin/app", map[string]any{"name": "Diff isolation"}, adminToken, false)).(map[string]any)
		if w := getDiff(int64(otherApp["id"].(float64)), upgradeURL); w.Code != 404 {
			t.Fatal("cross-app diff download accepted", w.Code)
		}
		ok(request("DELETE", fmt.Sprintf("/api/admin/app/%.0f", otherApp["id"]), nil, adminToken, false))
		for _, name := range []string{"bad.diff", "..%5Csecret", "%2e%2e%2fsecret"} {
			if w := getDiff(appID, "/api/sdk/diff/"+name); w.Code == 200 {
				t.Fatal("invalid diff path accepted", name)
			}
		}
		patchPath := filepath.Join(cfg.Clone.OutputDir, fmt.Sprint(appID), filepath.Base(upgradeURL))
		past := time.Now().Add(-time.Hour)
		if err := os.Chtimes(patchPath, past, past); err != nil {
			t.Fatal(err)
		}
		if w := getDiff(appID, upgradeURL); w.Code != 404 {
			t.Fatal("expired patch accepted")
		}
		if err := os.Remove(uploaded["path"].(string)); err != nil {
			t.Fatal(err)
		}
		fallback := ok(request("POST", "/api/sdk/ini", map[string]any{"versionKey": "stable", "versionNo": "1"}, "", true)).(map[string]any)["latestVersion"].(map[string]any)
		if fallback["diffUrl"] != nil || fallback["apkUrl"] != newUpload["url"] || fallback["hasUpdate"] != true {
			t.Fatal("missing base did not fall back to full APK", fallback)
		}
		info := ok(request("POST", "/api/sdk/info", map[string]any{}, token, true)).(map[string]any)
		if info["user"].(map[string]any)["hasSignKey"] != true {
			t.Fatal("signature key flag not refreshed")
		}
	}
	ok(request("POST", "/api/admin/function", map[string]any{"name": "echo", "trigger": "manual", "enabled": true, "code": "function handler(ctx) { return {echo: ctx.payload}; }", "timeout": 500}, adminToken, false))
	cloud := ok(request("POST", "/api/sdk/cloudFunction", map[string]any{"name": "echo", "param": map[string]any{"x": 1}}, token, true)).(map[string]any)
	if cloud["echo"].(map[string]any)["x"] != float64(1) {
		t.Fatalf("cloud payload %v", cloud)
	}
	broken := ok(request("POST", "/api/admin/function", map[string]any{"name": "broken", "trigger": "manual", "enabled": true, "code": "function handler(){throw new Error('broken');}"}, adminToken, false)).(map[string]any)
	if out := request("POST", "/api/sdk/cloudFunction", map[string]any{"name": "broken"}, token, true); out["code"] != float64(500) {
		t.Fatalf("cloud failure %v", out)
	}
	ok(request("DELETE", fmt.Sprintf("/api/admin/function/%.0f", broken["id"]), nil, adminToken, false))
	group, e := d.Exec(repo.SQL("INSERT INTO {kami_groups} (appid,name,type,value,remark) VALUES (?,'VIP','vip',3600,'')"), appID)
	if e != nil {
		t.Fatal(e)
	}
	gid, _ := group.LastInsertId()
	cards := ok(request("POST", "/api/admin/kami/generate", map[string]any{"groupId": gid, "count": 2, "length": 16}, adminToken, false)).(map[string]any)["cards"].([]any)
	ok(request("POST", "/api/sdk/kamiTopup", map[string]any{"card": cards[0]}, token, true))
	concurrentCards := ok(request("POST", "/api/admin/kami/generate", map[string]any{"groupId": gid, "count": 1, "length": 16}, adminToken, false)).(map[string]any)["cards"].([]any)
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(map[string]any{"card": concurrentCards[0]})
			req := httptest.NewRequest("POST", "/api/sdk/kamiTopup", bytes.NewReader(body))
			ts := fmt.Sprint(time.Now().Unix())
			nonce := fmt.Sprintf("concurrency-%d", i)
			for k, v := range map[string]string{"Content-Type": "application/json", "X-App-Id": fmt.Sprint(appID), "X-Mi": "test-device", "Authorization": "Bearer " + token, "X-Ts": ts, "X-Nonce": nonce, "X-Sign": sdkcrypto.SignBytes(key, ts, nonce, body)} {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			wrapper.ServeHTTP(w, req)
			var out struct{ Code int }
			if json.Unmarshal(w.Body.Bytes(), &out) != nil {
				codes <- 500
			} else {
				codes <- out.Code
			}
		}(i)
	}
	wg.Wait()
	close(codes)
	success := 0
	for code := range codes {
		if code == 0 {
			success++
		} else if code != 2002 {
			t.Fatalf("concurrent topup code %d", code)
		}
	}
	if success != 1 {
		t.Fatalf("card consumed %d times", success)
	}
	if v := request("POST", "/api/sdk/kamiTopup", map[string]any{"card": cards[0]}, token, true); v["code"] == float64(0) {
		t.Fatal("card reused")
	}
	ok(request("POST", "/api/sdk/logout", map[string]any{}, token, true))
	listed = ok(request("GET", "/api/admin/user/list", nil, adminToken, false)).(map[string]any)["list"].([]any)
	if listed[0].(map[string]any)["online"] != true {
		t.Fatal("logout removed other live sessions")
	}
	ok(request("POST", "/api/sdk/logout", map[string]any{}, user["token"].(string), true))
	ok(request("POST", "/api/sdk/logout", map[string]any{}, extraLogin["token"].(string), true))
	listed = ok(request("GET", "/api/admin/user/list", nil, adminToken, false)).(map[string]any)["list"].([]any)
	if listed[0].(map[string]any)["online"] != false {
		t.Fatal("logout left online flag")
	}
	if n, err := r.ZCard(context.Background(), fmt.Sprintf("presence:%d", appID)).Result(); err != nil || n != 0 {
		t.Fatalf("presence after logout: %d %v", n, err)
	}
	if v := request("POST", "/api/sdk/info", map[string]any{}, token, true); v["code"] != float64(1006) {
		t.Fatal("logged out token accepted")
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET type='kami',mode='vip' WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	if out := request("POST", "/api/sdk/logon", map[string]any{"card": cards[0], "mi": "test-device"}, "", true); out["code"] != float64(2002) {
		t.Fatal("consumed topup card accepted as account", out)
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET bind_region=1,geo_provider='tencent',geo_key='' WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	if v := request("POST", "/api/sdk/logon", map[string]any{"card": cards[1], "mi": "test-device"}, "", true); v["code"] == float64(0) {
		t.Fatal("geo failure accepted")
	}
	var cardStatus int
	if e = d.Get(&cardStatus, repo.SQL("SELECT status FROM {kamis} WHERE appid=? AND card=?"), appID, cards[1]); e != nil || cardStatus != 0 {
		t.Fatalf("failed geo consumed card: %d %v", cardStatus, e)
	}
	// A single SQL connection must suffice while geo resolution runs inside activation.
	d.SetMaxOpenConns(1)
	d.SetMaxIdleConns(1)
	if _, e = d.Exec(repo.SQL("UPDATE {settings} SET cache_enabled=1 WHERE id=1")); e != nil {
		t.Fatal(e)
	}
	if _, e = d.Exec(repo.SQL("INSERT INTO {ip_info} (ip,provider,lat,lng,nation,province,city,district,update_time) VALUES ('192.0.2.1','tencent',31,121,'CN','Test','City','',?)"), time.Now().Unix()); e != nil {
		t.Fatal(e)
	}
	cardLoginWithGeo := ok(request("POST", "/api/sdk/logon", map[string]any{"card": cards[1], "mi": "test-device"}, "", true)).(map[string]any)
	geoToken := cardLoginWithGeo["token"].(string)
	geoID := int64(cardLoginWithGeo["user"].(map[string]any)["id"].(float64))
	if _, e = d.Exec(repo.SQL("UPDATE {kamis} SET region_limit=1 WHERE id=?"), geoID); e != nil {
		t.Fatal(e)
	}
	if _, e = d.Exec(repo.SQL("INSERT INTO {ip_info} (ip,provider,lat,lng,nation,province,city,district,update_time) VALUES ('192.0.2.2','tencent',40,116,'CN','Far','City','',?)"), time.Now().Unix()); e != nil {
		t.Fatal(e)
	}
	remoteIP = "192.0.2.2"
	for _, path := range []string{"info", "heartbeat", "logon"} {
		if out := request("POST", "/api/sdk/"+path, map[string]any{"card": cards[1], "mi": "test-device"}, geoToken, true); out["code"] != float64(1009) {
			t.Fatalf("region limit bypass: %s %v", path, out)
		}
	}
	if _, e = d.Exec(repo.SQL("UPDATE {kamis} SET region_limit=2 WHERE id=?"), geoID); e != nil {
		t.Fatal(e)
	}
	ok(request("POST", "/api/sdk/logon", map[string]any{"card": cards[1], "mi": "test-device"}, "", true))
	var device repo.UserDevice
	if e = d.Get(&device, repo.SQL("SELECT id,uid,mi,device,os,ip,region,bind_time,last_time FROM {user_devices} WHERE uid=?"), -geoID); e != nil || device.IP != remoteIP || device.Region != "FarCity" {
		t.Fatalf("device IP/region did not update: %+v %v", device, e)
	}
	var count int
	if e = d.Get(&count, repo.SQL("SELECT COUNT(*) FROM {user_logs} WHERE uid=? AND action='login' AND ip=? AND region='FarCity'"), -geoID, remoteIP); e != nil || count != 1 {
		t.Fatalf("login region missing: %d %v", count, e)
	}
	remoteIP = "192.0.2.1"
	t.Run("UserRegionRecords", func(t *testing.T) {
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET type='user',mode='free' WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		defer func() {
			remoteIP = "192.0.2.1"
			if _, err := d.Exec(repo.SQL("UPDATE {apps} SET type='kami',mode='vip' WHERE id=?"), appID); err != nil {
				t.Error(err)
			}
		}()
		registered := ok(request("POST", "/api/sdk/register", map[string]any{"username": "geotester", "password": "test-password", "mi": "test-device"}, "", true)).(map[string]any)
		id := int64(registered["user"].(map[string]any)["id"].(float64))
		ut := registered["token"].(string)
		ok(request("PUT", fmt.Sprintf("/api/admin/user/%d", id), map[string]any{"regionLimit": 1}, adminToken, false))
		remoteIP = "192.0.2.2"
		if out := request("POST", "/api/sdk/heartbeat", map[string]any{}, ut, true); out["code"] != float64(1009) {
			t.Fatal("user region bypass", out)
		}
		ok(request("PUT", fmt.Sprintf("/api/admin/user/%d", id), map[string]any{"regionLimit": 2}, adminToken, false))
		ok(request("POST", "/api/sdk/logon", map[string]any{"username": "geotester", "password": "test-password", "mi": "test-device"}, "", true))
		row, err := repo.NewUserRepo(d).GetByID(context.Background(), appID, id)
		if err != nil || row.RegisterIP != "192.0.2.1" || row.LastIP != remoteIP || row.LastRegion != "FarCity" || row.Region != "FarCity" {
			t.Fatalf("user IP/region mismatch: %+v %v", row, err)
		}
		var locations []any
		if err := json.Unmarshal([]byte(row.IPList), &locations); err != nil || len(locations) != 2 {
			t.Fatalf("user IP list: %s %v", row.IPList, err)
		}
		if err := d.Get(&count, repo.SQL("SELECT COUNT(*) FROM {user_logs} WHERE uid=? AND region<>'' AND action IN ('login','register')"), id); err != nil || count != 2 {
			t.Fatalf("user region logs: %d %v", count, err)
		}
		ok(request("DELETE", fmt.Sprintf("/api/admin/user/%d", id), nil, adminToken, false))
	})
	ok(request("POST", "/api/sdk/logout", map[string]any{}, cardLoginWithGeo["token"].(string), true))
	d.SetMaxOpenConns(10)
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET bind_region=0 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	cardLogin := ok(request("POST", "/api/sdk/logon", map[string]any{"card": cards[1], "mi": "test-device"}, "", true)).(map[string]any)
	cardToken := cardLogin["token"].(string)
	t.Run("CardDeviceRevokesSession", func(t *testing.T) {
		id := int64(cardLogin["user"].(map[string]any)["id"].(float64))
		base := fmt.Sprintf("/api/admin/kami/%d", id)
		devices := ok(request("GET", base+"/devices", nil, adminToken, false)).([]any)
		ok(request("DELETE", fmt.Sprintf("%s/devices/%.0f", base, devices[0].(map[string]any)["id"]), nil, adminToken, false))
		if rows := ok(request("GET", base+"/sessions", nil, adminToken, false)).([]any); len(rows) != 0 {
			t.Fatal("unbound card device still in sessions", rows)
		}
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET bind_device=0 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		if v := request("POST", "/api/sdk/heartbeat", map[string]any{}, cardToken, true); v["code"] != float64(1006) {
			t.Fatal("unbound card token revived", v)
		}
		if _, err := d.Exec(repo.SQL("UPDATE {apps} SET bind_device=1 WHERE id=?"), appID); err != nil {
			t.Fatal(err)
		}
		cardToken = ok(request("POST", "/api/sdk/logon", map[string]any{"card": cards[1], "mi": "test-device"}, "", true)).(map[string]any)["token"].(string)
	})
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET brute_force_enabled=1 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 6; i++ {
		out := request("POST", "/api/sdk/logon", map[string]any{"card": "invalid-card", "mi": "test-device"}, "", true)
		want := float64(2001)
		if i == 5 {
			want = 1017
		}
		if out["code"] != want {
			t.Fatalf("card lockout %d: %v", i, out)
		}
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET brute_force_enabled=0 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	ok(request("POST", "/api/sdk/heartbeat", map[string]any{}, cardToken, true))
	if out := request("POST", "/api/sdk/updateInfo", map[string]any{"avatar": "/fake.png"}, cardToken, true); out["code"] != float64(400) {
		t.Fatal("card avatar accepted", out)
	}
	for _, path := range []string{"regen-sign", "uploadAvatar"} {
		if out := request("POST", "/api/sdk/"+path, map[string]any{}, cardToken, true); out["code"] != float64(400) {
			t.Fatal("card private feature accepted", path, out)
		}
	}
	ok(request("POST", "/api/sdk/userExtends", map[string]any{"set": map[string]string{"cardFlag": "yes"}}, cardToken, true))
	for _, path := range []string{"/log/list", "/log/list?category=user", "/dashboard"} {
		ok(request(http.MethodGet, "/api/admin"+path, nil, adminToken, false))
	}
	dashboard := ok(request("GET", "/api/admin/dashboard", nil, adminToken, false)).(map[string]any)
	if dashboard["apiTotal"].(float64) < 10 || dashboard["onlinePeakToday"].(float64) < 1 || len(dashboard["recent"].([]any)) == 0 {
		t.Fatalf("missing metrics %v", dashboard)
	}
	ok(request("POST", fmt.Sprintf("/api/admin/app/%d/mtls/gen-ca", appID), map[string]any{}, adminToken, false))
	issued := ok(request("POST", fmt.Sprintf("/api/admin/app/%d/mtls/issue-cert", appID), map[string]any{}, adminToken, false)).(map[string]any)
	if !strings.Contains(issued["cert"].(string), "BEGIN CERTIFICATE") {
		t.Fatal("client certificate missing")
	}
	if _, e := os.Stat("data/master.key"); e != nil {
		t.Fatal("master key not generated on enable")
	}
	ok(request("POST", "/api/admin/system/https/gen-cert", map[string]any{"host": "localhost"}, adminToken, false))
	tlsAddr := https.Address()
	ok(request("POST", "/api/admin/system/https", map[string]any{"enabled": true, "addr": ":8443"}, adminToken, false))
	if https.Address() != tlsAddr {
		t.Fatal("HTTPS changed the configured service port")
	}
	certificate := ok(request("GET", "/api/admin/system/https", nil, adminToken, false)).(map[string]any)["cert"].(string)
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(certificate)) {
		t.Fatal("invalid HTTPS certificate")
	}
	pair, e := tls.X509KeyPair([]byte(issued["cert"].(string)), []byte(issued["key"].(string)))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET mtls_enabled=1 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	foreignCA, foreignKey, err := tlsmgr.Generate("foreign", true)
	if err != nil {
		t.Fatal(err)
	}
	foreignCert, foreignPrivate, err := tlsmgr.IssueClient(foreignCA, foreignKey, appID+1)
	if err != nil {
		t.Fatal(err)
	}
	foreignPair, err := tls.X509KeyPair([]byte(foreignCert), []byte(foreignPrivate))
	if err != nil {
		t.Fatal(err)
	}
	for _, certCase := range []string{"missing", "valid", "foreign"} {
		tlsConfig := &tls.Config{RootCAs: pool, ServerName: "localhost", MinVersion: tls.VersionTLS12}
		if certCase == "valid" {
			tlsConfig.Certificates = []tls.Certificate{pair}
		}
		if certCase == "foreign" {
			tlsConfig.Certificates = []tls.Certificate{foreignPair}
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client := http.Client{Transport: transport, Timeout: 5 * time.Second}
		req, _ := http.NewRequest("GET", "https://"+tlsAddr+"/api/sdk/extend-file/list", nil)
		req.Header.Set("X-App-Id", fmt.Sprint(appID))
		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		var out struct{ Code int }
		e = json.NewDecoder(res.Body).Decode(&out)
		res.Body.Close()
		transport.CloseIdleConnections()
		if e != nil || (certCase == "valid" && out.Code != 0) || (certCase != "valid" && out.Code != 1019) {
			t.Fatalf("mTLS case=%s code=%d err=%v", certCase, out.Code, e)
		}
	}
	if _, e = d.Exec(repo.SQL("UPDATE {apps} SET mtls_enabled=0 WHERE id=?"), appID); e != nil {
		t.Fatal(e)
	}
	ok(request("POST", "/api/admin/system/https", map[string]any{"enabled": false}, adminToken, false))
	// The default SDK path must stop resolving after a rewrite.
	var codeID int64
	if e = d.Get(&codeID, repo.SQL("SELECT id FROM {api_codes} WHERE code=1006")); e != nil {
		t.Fatal(e)
	}
	ok(request("POST", "/api/admin/system/api-codes/save", map[string]any{"items": []any{map[string]any{"id": codeID, "overrideCode": 91006, "rewritten": true}}}, adminToken, false))
	if out := request("POST", "/api/sdk/info", map[string]any{}, "invalid", true); out["code"] != float64(91006) {
		t.Fatalf("code rewrite: %v", out)
	}
	ok(request("POST", "/api/admin/system/api-codes/save", map[string]any{"items": []any{map[string]any{"id": codeID, "overrideCode": 91006, "rewritten": false}}}, adminToken, false))
	var routeID int64
	if e = d.Get(&routeID, repo.SQL("SELECT id FROM {api_routes} WHERE api_key='ini'")); e != nil {
		t.Fatal(e)
	}
	patchRoute := func(path string) {
		ok(request("POST", "/api/admin/system/api-routes/save", map[string]any{"items": []any{map[string]any{"id": routeID, "overridePath": path, "rewritten": true}}}, adminToken, false))
	}
	patchRoute("/api/sdk/custom-ini")
	ok(request("POST", "/api/sdk/custom-ini", map[string]any{}, "", true))
	oldReq := httptest.NewRequest("POST", "/api/sdk/ini", nil)
	oldW := httptest.NewRecorder()
	wrapper.ServeHTTP(oldW, oldReq)
	if oldW.Code != 404 {
		t.Fatal("default path still enabled")
	}
	bad := request("POST", "/api/admin/system/api-routes/save", map[string]any{"items": []any{map[string]any{"id": routeID, "overridePath": "/api/sdk/logon", "rewritten": true}}}, adminToken, false)
	if bad["code"] == float64(0) {
		t.Fatal("route collision accepted")
	}
	settings := ok(request("GET", "/api/admin/system/settings", nil, adminToken, false)).(map[string]any)
	// This test bypasses the installer, whose finalization normally sets entry.
	settings["adminEntry"] = "admin"
	t.Run("SystemSettingsSwitchesAndLogLevels", func(t *testing.T) {
		original := map[string]any{}
		for k, v := range settings {
			original[k] = v
		}
		for _, k := range []string{"cacheEnabled", "userLogEnabled", "adminLogEnabled"} {
			original[k] = settings[k] == float64(1) || settings[k] == true
		}
		defer func() { ok(request("POST", "/api/admin/system/settings", original, adminToken, false)) }()
		for _, enabled := range []bool{false, true} {
			for _, k := range []string{"cacheEnabled", "userLogEnabled", "adminLogEnabled"} {
				settings[k] = enabled
			}
			settings["logLevel"] = "all"
			ok(request("POST", "/api/admin/system/settings", settings, adminToken, false))
			read := ok(request("GET", "/api/admin/system/settings", nil, adminToken, false)).(map[string]any)
			want := float64(0)
			if enabled {
				want = 1
			}
			for _, k := range []string{"cacheEnabled", "userLogEnabled", "adminLogEnabled"} {
				if read[k] != want {
					t.Fatal(k, read)
				}
			}
			var before, after int
			d.Get(&before, repo.SQL("SELECT COUNT(*) FROM {logs} WHERE type='admin'"))
			request("POST", "/api/admin/notice", map[string]any{}, adminToken, false)
			d.Get(&after, repo.SQL("SELECT COUNT(*) FROM {logs} WHERE type='admin'"))
			if (after > before) != enabled {
				t.Fatal("admin log switch ignored", before, after)
			}
		}
		count := func() int {
			var n int
			d.Get(&n, repo.SQL("SELECT COUNT(*) FROM {logs} WHERE type='system'"))
			return n
		}
		settings["logLevel"] = "warning"
		ok(request("POST", "/api/admin/system/settings", settings, adminToken, false))
		before := count()
		ok(request("POST", "/api/admin/system/settings", settings, adminToken, false))
		if count() != before {
			t.Fatal("warning recorded success")
		}
		bad := map[string]any{"logLevel": "bad"}
		request("POST", "/api/admin/system/settings", bad, adminToken, false)
		if count() != before+1 {
			t.Fatal("warning missing failed operation")
		}
		settings["logLevel"] = "error"
		ok(request("POST", "/api/admin/system/settings", settings, adminToken, false))
		before = count()
		request("POST", "/api/admin/system/settings", bad, adminToken, false)
		if count() != before {
			t.Fatal("error recorded validation failure")
		}
		result := ok(request("GET", "/api/admin/log/list?type=system", nil, adminToken, false)).(map[string]any)
		if result["total"].(float64) == 0 {
			t.Fatal("system logs missing")
		}
		for _, row := range result["list"].([]any) {
			if row.(map[string]any)["summary"] == "" {
				t.Fatal("missing readable summary")
			}
		}
	})
	settings["adminEntry"] = "private"
	settings["adminTokenExpire"] = float64(120)
	for _, k := range []string{"cacheEnabled", "userLogEnabled", "adminLogEnabled"} {
		settings[k] = true
	}
	ok(request("POST", "/api/admin/system/settings", settings, adminToken, false))
	ok(request("GET", "/api/private/me", nil, adminToken, false))
	if ttl, err := r.TTL(context.Background(), "sess:"+adminToken).Result(); err != nil || ttl <= 0 || ttl > 121*time.Second {
		t.Fatalf("admin TTL not updated: %v %v", ttl, err)
	}
	if os.Getenv("AUTH_RECOVERY_BROWSER") == "1" {
		_, filename, _, _ := runtime.Caller(0)
		frontend := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../frontend"))
		httpServer := httptest.NewServer(wrapper)
		defer httpServer.Close()
		command := exec.Command("node", "scripts/live-backend.mjs")
		command.Dir = frontend
		command.Env = append(os.Environ(), "API_PROXY_TARGET="+httpServer.URL, "TEST_APP_ID="+fmt.Sprint(appID))
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("browser: %v %s", e, out)
		} else {
			t.Log(string(out))
		}
	}
	var audit, ulog int
	// Deleting used cards must clean the negative UID namespace atomically.
	ok(request("POST", "/api/private/kami/clean-used", map[string]any{}, adminToken, false))
	if out := request("POST", "/api/sdk/info", map[string]any{}, cardToken, true); out["code"] != float64(1006) {
		t.Fatal("deleted card session accepted", out)
	}
	var orphan int
	if e = d.Get(&orphan, repo.SQL("SELECT COUNT(*) FROM {user_extends} WHERE uid<0")); e != nil || orphan != 0 {
		t.Fatalf("card variable cleanup %d %v", orphan, e)
	}
	if e = d.Get(&orphan, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid<0")); e != nil || orphan != 0 {
		t.Fatalf("card device cleanup %d %v", orphan, e)
	}
	if e = d.Get(&audit, repo.SQL("SELECT COUNT(*) FROM {logs}")); e != nil {
		t.Fatal(e)
	}
	if e = d.Get(&ulog, repo.SQL("SELECT COUNT(*) FROM {user_logs}")); e != nil {
		t.Fatal(e)
	}
	if audit == 0 || ulog == 0 {
		t.Fatalf("logs missing admin=%d user=%d", audit, ulog)
	}
	t.Logf("42 migrations, admin/SDK/card/function lifecycle, audit=%d userlogs=%d", audit, ulog)
}
