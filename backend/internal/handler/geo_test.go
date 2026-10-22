package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/geo"
	"auth-api/internal/repo"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLocalGeoManagementPermissionAndErrors(t *testing.T) {
	l := geo.NewLocalDB(t.TempDir())
	defer l.Close()
	g := gin.New()
	admin := g.Group("/api/admin")
	admin.Use(func(c *gin.Context) { c.Set("admin", &auth.Session{Role: "admin"}) })
	MountLocalGeo(admin, l)
	for _, tc := range []struct {
		method, path string
		code         int
	}{
		{"GET", "/geo/local", 0},
		{"GET", "/geo/local/lookup?ip=127.0.0.1", 400},
		{"GET", "/geo/local/lookup?ip=1.2.3.4", 400},
		{"POST", "/geo/local/update", 403},
	} {
		w := httptest.NewRecorder()
		g.ServeHTTP(w, httptest.NewRequest(tc.method, "/api/admin"+tc.path, nil))
		var result struct{ Code int }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Code != tc.code {
			t.Fatal(tc, w.Body.String(), err)
		}
	}
}

func TestAppCannotEnableMissingLocalGeo(t *testing.T) {
	g := gin.New()
	h := NewAdminAppHandler(nil)
	g.POST("/app", h.Create)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/app", bytes.NewBufferString(`{"name":"Test","type":"user","mode":"free","bindRegion":true,"geoProvider":"dbip"}`))
	r.Header.Set("Content-Type", "application/json")
	g.ServeHTTP(w, r)
	var result struct{ Code int }
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Code != 400 {
		t.Fatal(w.Body.String(), err)
	}
}

func TestGeoProviderValidation(t *testing.T) {
	for _, provider := range []string{"", "tencent", "amap", "dbip", "http://localhost"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		ok := validApp(c, &repo.App{Type: "user", Mode: "free", Status: 1, UserTokenTTL: 60, GeoProvider: provider})
		if ok != (provider != "http://localhost") {
			t.Fatal(provider, ok)
		}
	}
}

func TestLocalRegionBindingFailsClosedWithoutDatabase(t *testing.T) {
	for _, kami := range []bool{false, true} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/sdk/logon", nil)
		c.Request.RemoteAddr = "1.2.3.4:1234"
		c.Set("sdkApp", &repo.App{ID: 1, BindRegion: true, GeoProvider: geo.LocalProvider})
		service := &SDKService{}
		// A nil transaction ensures the local branch does not query SQL cache or fall back online.
		if service.checkRegionTx(c, &repo.User{ID: 1, AppID: 1}, kami, nil) {
			t.Fatal("missing local database bypassed region binding", kami)
		}
		var result struct{ Code int }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Code != 500 {
			t.Fatal(w.Body.String(), err)
		}
	}
}
