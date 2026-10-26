package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/response"
	"database/sql/driver"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func TestRecentActivity(t *testing.T) {
	for _, tc := range []struct{ action, data, kind, want string }{
		{"POST /api/admin/function", `{"httpStatus":200}`, "success", "admin 保存云函数"},
		{"DELETE /api/hidden/user/:id", `{"httpStatus":403}`, "success", "admin 删除用户（失败）"},
		{"修改了用户资料", `{"success":true}`, "success", "admin 修改了用户资料"},
		{"试运行了云函数", `{"success":false}`, "success", "admin 试运行云函数（失败）"},
		{"POST /api/secret/system/api-routes/randomize-all", `{"success":true}`, "success", "admin 随机重定义了接口路径"},
		{"POST /api/admin/unknown", `{}`, "success", "admin 执行后台操作"},
		{"login", "", "primary", "admin 登录了客户端"},
		{"register", "", "primary", "admin 注册了账号"},
	} {
		if got := recentActivity("admin", tc.action, tc.data, tc.kind); got != tc.want {
			t.Errorf("%s: %s != %s", tc.action, got, tc.want)
		}
	}
}

func TestAuditApplicationOwnership(t *testing.T) {
	for _, tc := range []struct {
		route, path   string
		created, want int64
	}{
		{"/api/private/user", "/api/private/user", 0, 1000},
		{"/api/private/ver", "/api/private/ver", 0, 1000},
		{"/api/private/app/:id", "/api/private/app/2000", 0, 2000},
		{"/api/private/app/:id/mtls/gen-ca", "/api/private/app/2000/mtls/gen-ca", 0, 2000},
		{"/api/private/app", "/api/private/app", 3000, 3000},
		{"/api/private/app", "/api/private/app", 0, 0},
		{"/api/private/system/settings", "/api/private/system/settings", 0, 0},
		{"/api/private/admin", "/api/private/admin", 0, 0},
		{"/api/private/me/profile", "/api/private/me/profile", 0, 0},
	} {
		t.Run(tc.path, func(t *testing.T) {
			g := gin.New()
			g.POST(tc.route, func(c *gin.Context) {
				c.Set("appid", int64(1000))
				c.Set("audit.appid", tc.created)
				if got := auditAppID(c); got != tc.want {
					t.Fatalf("owner = %d, want %d", got, tc.want)
				}
			})
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", tc.path, nil))
		})
	}
}

type auditResultMatcher struct {
	code    int
	success bool
}

func (m auditResultMatcher) Match(value driver.Value) bool {
	s, ok := value.(string)
	if !ok || strings.Contains(s, "secret-value") {
		return false
	}
	var got struct {
		Code          int
		Success       bool
		Method, Route string
		HTTPStatus    int
	}
	return json.Unmarshal([]byte(s), &got) == nil && got.Code == m.code && got.Success == m.success && got.HTTPStatus == 200 && got.Method == "POST" && got.Route == "/api/private/function/:id/run"
}

func TestAuditUsesBusinessOutcomeWithoutSavingSecrets(t *testing.T) {
	for _, tc := range []struct {
		code         int
		scriptFailed bool
	}{{0, false}, {400, false}, {0, true}} {
		raw, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		d := sqlx.NewDb(raw, "sqlmock")
		mock.ExpectQuery("SELECT admin_log_enabled").WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(1))
		mock.ExpectExec("INSERT INTO .*logs").WithArgs(int64(0), "试运行了云函数", auditResultMatcher{tc.code, tc.code == 0 && !tc.scriptFailed}, "0", "admin", int64(1), sqlmock.AnyArg(), "admin", "", "", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		g := gin.New()
		g.Use(func(c *gin.Context) { c.Set("admin", &auth.Session{AdminID: 1, Username: "admin"}); c.Next() }, AdminAudit(d))
		g.POST("/api/private/function/:id/run", func(c *gin.Context) {
			c.Set("audit.executionFailed", tc.scriptFailed)
			if tc.code == 0 {
				response.OK(c, gin.H{"privateKey": "secret-value"})
			} else {
				response.Fail(c, tc.code, "secret-value")
			}
		})
		w := httptest.NewRecorder()
		g.ServeHTTP(w, httptest.NewRequest("POST", "/api/private/function/0/run", strings.NewReader(`{"password":"secret-value"}`)))
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		d.Close()
	}
}
