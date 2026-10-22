package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http/httptest"
	"testing"
)

func TestSystemAuditLevelAndIndependentSwitch(t *testing.T) {
	for _, tc := range []struct {
		level         string
		code, enabled int
		want          bool
	}{
		{"all", 0, 0, true}, {"info", 0, 1, true}, {"warning", 0, 1, false}, {"warning", 400, 0, true}, {"error", 400, 1, false}, {"error", 500, 0, true},
	} {
		t.Run(tc.level+string(rune('A'+tc.code/100)), func(t *testing.T) {
			raw, m, e := sqlmock.New()
			if e != nil {
				t.Fatal(e)
			}
			defer raw.Close()
			d := sqlx.NewDb(raw, "sqlmock")
			m.ExpectQuery("SELECT admin_log_enabled").WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(tc.enabled))
			m.ExpectQuery("SELECT log_level").WillReturnRows(sqlmock.NewRows([]string{"log_level"}).AddRow(tc.level))
			if tc.want {
				m.ExpectExec("INSERT INTO .*logs.*VALUES \\('system'").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			g := gin.New()
			g.Use(func(c *gin.Context) { c.Set("admin", &auth.Session{AdminID: 1, Username: "admin"}) }, AdminAudit(d))
			g.POST("/api/private/system/settings", func(c *gin.Context) {
				if tc.code == 0 {
					response.OK(c, nil)
				} else {
					response.Fail(c, tc.code, "test")
				}
			})
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/private/system/settings", nil))
			if e = m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestUserLogSwitch(t *testing.T) {
	for _, enabled := range []int{0, 1} {
		raw, m, e := sqlmock.New()
		if e != nil {
			t.Fatal(e)
		}
		d := sqlx.NewDb(raw, "sqlmock")
		m.ExpectQuery("SELECT user_log_enabled").WillReturnRows(sqlmock.NewRows([]string{"user_log_enabled"}).AddRow(enabled))
		if enabled == 1 {
			m.ExpectExec("INSERT INTO .*user_logs").WillReturnResult(sqlmock.NewResult(1, 1))
		}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Set("sdkUserRecord", &repo.User{ID: 1})
		NewSDKService(d, nil).writeUserLog(c, 1, "login")
		if e = m.ExpectationsWereMet(); e != nil {
			t.Error(e)
		}
		d.Close()
	}
}
