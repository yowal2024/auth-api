package middleware

import (
	"auth-api/internal/auth"
	"auth-api/internal/response"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminPermissions(t *testing.T) {
	for _, tc := range []struct {
		method, route, query string
		allowed              bool
	}{
		{"GET", "app/list", "", true}, {"GET", "dashboard", "", true},
		{"PUT", "me/password", "", true}, {"POST", "upload/avatar", "", true},
		{"POST", "user", "", true}, {"DELETE", "user/:id/sessions/:sessionId", "", true},
		{"DELETE", "kami/:id/devices/:deviceId", "", true}, {"POST", "kami/generate", "", true},
		{"POST", "block", "", true}, {"POST", "notice", "", true},
		{"GET", "log/list", "?category=user", true}, {"GET", "log/actions", "?category=user", true},
		{"GET", "log/list", "", false}, {"GET", "log/list", "?category=user&type=system", false},
		{"POST", "log/clear", "?category=user", false}, {"POST", "app", "", false},
		{"PUT", "app/:id", "", false}, {"DELETE", "app/:id", "", false},
		{"POST", "app/:id/mtls/issue-cert", "", false}, {"POST", "app/:id/mtls/gen-ca", "", false},
		{"POST", "upload/app-logo", "", false}, {"GET", "encrypt/list", "", false},
		{"GET", "extend/list", "", false}, {"POST", "extend", "", false},
		{"GET", "extend-file/list", "", false}, {"POST", "extend-file/upload", "", false},
		{"GET", "ver/list", "", false}, {"POST", "function/0/run", "", false},
		{"GET", "system/backup/:backupId/download", "", false}, {"GET", "security-events", "", false},
		{"POST", "geo/local/update", "", false}, {"GET", "new-sensitive-feature", "", false},
	} {
		for _, role := range []string{"admin", "super", "unknown"} {
			g := gin.New()
			g.Use(func(c *gin.Context) { c.Set("admin", &auth.Session{Role: role}) }, AdminPermissions())
			g.Handle(tc.method, "/api/hidden/"+tc.route, func(c *gin.Context) { response.OK(c, nil) })
			w := httptest.NewRecorder()
			g.ServeHTTP(w, httptest.NewRequest(tc.method, "/api/hidden/"+tc.route+tc.query, nil))
			var out struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			allowed := role == "super" || (role == "admin" && tc.allowed)
			if (out.Code == 0) != allowed {
				t.Fatalf("%s %s %s: code %d", role, tc.method, tc.route, out.Code)
			}
		}
	}
}
