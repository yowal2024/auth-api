package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSDKLoginCorePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	d := sqlx.NewDb(raw, "sqlmock")
	mr := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer r.Close()
	repo.InitTablePrefix("a_")
	service := NewSDKService(d, r)
	g := gin.New()
	g.Use(func(c *gin.Context) {
		c.Set("sdkApp", &repo.App{ID: 1000, Type: "user", Mode: "free", Status: 1, UserTokenTTL: 600})
	})
	g.POST("/logon", service.Logon)
	hash, e := auth.HashPassword("password")
	if e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("SELECT COALESCE.*FROM a_blocks").WithArgs(int64(1000), "192.0.2.1", "device").WillReturnRows(sqlmock.NewRows([]string{"n"}).AddRow(0))
	m.ExpectQuery("SELECT id FROM a_users").WithArgs(int64(1000), "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	m.ExpectQuery("FROM a_users u LEFT JOIN a_user_groups").WithArgs(int64(1000), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "appid", "username", "password", "vip", "ban_expire"}).AddRow(7, 1000, "alice", hash, 0, 0))
	m.ExpectQuery("FROM a_functions WHERE appid").WithArgs(int64(1000), "before_login").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	m.ExpectExec("UPDATE a_users SET `mi`").WithArgs("device", int64(1000), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE a_users SET last_ip").WithArgs("192.0.2.1", "", sqlmock.AnyArg(), int64(1000), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectQuery("FROM a_functions WHERE appid").WithArgs(int64(1000), "after_login").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	m.ExpectQuery("SELECT user_log_enabled").WillReturnRows(sqlmock.NewRows([]string{"user_log_enabled"}).AddRow(0))
	req := httptest.NewRequest("POST", "/logon", strings.NewReader(`{"username":"alice","password":"password","mi":"device"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	var out struct {
		Code int
		Data struct {
			Token string
			User  struct{ ID int64 }
		}
	}
	if e = json.Unmarshal(w.Body.Bytes(), &out); e != nil || out.Code != 0 || len(out.Data.Token) != 64 || out.Data.User.ID != 7 {
		t.Fatalf("response: %s %v", w.Body.String(), e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestTopupRollsBackWhenCardAlreadyUsed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	d := sqlx.NewDb(raw, "sqlmock")
	repo.InitTablePrefix("a_")
	mr := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer r.Close()
	service := NewSDKService(d, r)
	g := gin.New()
	g.POST("/", func(c *gin.Context) {
		c.Set("sdkApp", &repo.App{ID: 1000})
		c.Set("sdkUserRecord", &repo.User{ID: 7, AppID: 1000, Username: "alice"})
		service.KamiTopup(c)
	})
	m.ExpectBegin()
	m.ExpectQuery("FROM a_kamis WHERE appid").WithArgs(int64(1000), "USED").WillReturnRows(sqlmock.NewRows([]string{"id", "appid", "status"}).AddRow(8, 1000, 1))
	m.ExpectRollback()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"card":"USED"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	var out struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.Code != 2002 {
		t.Fatal(w.Body.String())
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
