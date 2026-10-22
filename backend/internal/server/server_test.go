package server

import (
	"auth-api/internal/auth"
	"auth-api/internal/config"
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/migrations"
	"context"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRegisteredRouteInventory(t *testing.T) {
	g, m, _ := fixture(t)
	routes := g.Routes()
	// These routes have success and rejection cases in the database lifecycle tests.
	sdkRoutes := map[string]bool{}
	for _, route := range []string{"POST logon", "POST register", "POST ini", "POST info", "POST heartbeat", "POST logout", "POST updateInfo", "POST userExtends", "POST kamiTopup", "POST cloudFunction", "POST handshake", "POST uploadAvatar", "POST regen-sign", "GET clone", "GET diff/:name", "GET extend-file/list", "GET extend-file/:id/download"} {
		sdkRoutes[route] = true
	}
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/api/sdk/") {
			key := route.Method + " " + strings.TrimPrefix(route.Path, "/api/sdk/")
			if !sdkRoutes[key] {
				t.Errorf("SDK route needs coverage: %s", key)
			}
			delete(sdkRoutes, key)
		}
	}
	if len(sdkRoutes) > 0 {
		t.Fatal("covered SDK routes disappeared", sdkRoutes)
	}
	items := make([]map[string]string, 0, len(routes))
	for _, r := range routes {
		items = append(items, map[string]string{"method": r.Method, "path": r.Path})
	}
	if p := os.Getenv("AUTH_RECOVERY_ROUTE_REPORT"); p != "" {
		b, e := json.MarshalIndent(items, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("registered routes: %d", len(items))
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func fixture(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, *redis.Client) {
	t.Helper()
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	d := sqlx.NewDb(raw, "sqlmock")
	t.Cleanup(func() { d.Close() })
	mr := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { r.Close() })
	cfg := &config.Config{Installed: true, Session: config.SessionConfig{TTLSec: 3600}}
	temp := t.TempDir()
	cfg.Upload.Root = filepath.Join(temp, "uploads")
	cfg.Clone = config.CloneConfig{CanonicalDir: filepath.Join(temp, "canonical"), OutputDir: filepath.Join(temp, "diffs")}
	h := install.NewHandler("unused", cfg, migrations.Files)
	if e = repo.InitTablePrefix("a_"); e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("SELECT id,web_url").WillReturnRows(sqlmock.NewRows([]string{"id", "admin_entry", "admin_token_expire"}).AddRow(1, "private", 3600))
	engine, e := New(Options{Config: cfg, DB: d, Redis: r, Installer: h})
	if e != nil {
		t.Fatal(e)
	}
	return engine, m, r
}
func call(t *testing.T, g *gin.Engine, method, path, body, token string) map[string]any {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	g.ServeHTTP(w, req)
	var v map[string]any
	if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil {
		t.Fatalf("invalid JSON %s: %v", w.Body.String(), e)
	}
	return v
}
func adminRows(hash, role string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "password", "role", "last_login", "last_ip"}).AddRow(1, "admin", "", "", hash, role, 0, "")
}
func TestLoginMeLogoutThroughRouter(t *testing.T) {
	g, m, _ := fixture(t)
	hash, e := auth.HashPassword("correct123")
	if e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("FROM a_admins WHERE username").WithArgs("admin").WillReturnRows(adminRows(hash, "super"))
	m.ExpectExec("UPDATE a_admins SET last_login").WithArgs(sqlmock.AnyArg(), "192.0.2.1", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	v := call(t, g, "POST", "/api/private/login", `{"username":"admin","password":"correct123"}`, "")
	if v["code"] != float64(0) {
		t.Fatalf("login: %v", v)
	}
	token := v["data"].(map[string]any)["token"].(string)
	m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows(hash, "super"))
	if v = call(t, g, "GET", "/api/private/me", "", token); v["code"] != float64(0) {
		t.Fatal(v)
	}
	m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows(hash, "super"))
	if v = call(t, g, "POST", "/api/private/logout", "{}", token); v["code"] != float64(0) {
		t.Fatal(v)
	}
	if v = call(t, g, "GET", "/api/private/me", "", token); v["code"] != float64(401) {
		t.Fatal("logged-out token accepted")
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestAdminAuthorizationAndAppHeader(t *testing.T) {
	g, m, r := fixture(t)
	token, e := auth.NewSessionStore(r, time.Hour, 0).Create(context.Background(), &auth.Session{AdminID: 1, Role: "super", Credential: auth.CredentialRevision("")})
	if e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows("", "admin"))
	if v := call(t, g, "GET", "/api/private/admin/list", "", token); v["code"] != float64(403) {
		t.Fatal("demoted account retained privilege", v)
	}
	m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows("", "admin"))
	if v := call(t, g, "GET", "/api/private/user/list", "", token); v["code"] != float64(400) {
		t.Fatal("missing app header accepted", v)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestInstalledEndpointsLocked(t *testing.T) {
	g, m, _ := fixture(t)
	for _, endpoint := range []string{"test-mysql", "test-redis", "save-config", "migrate", "create-admin", "finalize"} {
		v := call(t, g, "POST", "/api/install/"+endpoint, "{}", "")
		if v["code"] != float64(403) {
			t.Fatalf("%s: %v", endpoint, v)
		}
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestCloudFunctionManagementRequiresSuper(t *testing.T) {
	g, m, r := fixture(t)
	token, err := auth.NewSessionStore(r, time.Hour, 0).Create(context.Background(), &auth.Session{AdminID: 1, Credential: auth.CredentialRevision("hash"), Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ method, path string }{
		{"GET", "/function/list"}, {"GET", "/function/1"}, {"POST", "/function"},
		{"DELETE", "/function/1"}, {"POST", "/function/0/run"}, {"POST", "/function/1/run"},
	} {
		m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows("hash", "admin"))
		out := call(t, g, route.method, "/api/private"+route.path, `{}`, token)
		if out["code"] != float64(403) {
			t.Fatalf("%s accepted: %v", route.path, out)
		}
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdminPasswordRevisionInvalidatesStaleAndLegacySessions(t *testing.T) {
	g, m, r := fixture(t)
	for _, credential := range []string{auth.CredentialRevision("old-hash"), ""} {
		token, err := auth.NewSessionStore(r, time.Hour, 0).Create(context.Background(), &auth.Session{AdminID: 1, Credential: credential, Role: "super"})
		if err != nil {
			t.Fatal(err)
		}
		// Simulate revocation cache failure or an old-password login completing after a reset.
		m.ExpectQuery("FROM a_admins WHERE id").WithArgs(int64(1)).WillReturnRows(adminRows("new-hash", "super"))
		if out := call(t, g, "GET", "/api/private/me", "", token); out["code"] != float64(401) {
			t.Fatalf("stale credential accepted: %v", out)
		}
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
