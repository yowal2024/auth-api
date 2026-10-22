package handler

import (
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http/httptest"
	"testing"
)

func TestPartialAppUpdatePreservesSecurity(t *testing.T) {
	a := repo.App{BindDevice: true, BindRegion: true, DiffUpdate: true, RandomPackage: true, BruteForce: true, MtlsEnabled: true, Encryption: 7}
	var req appReq
	if err := json.Unmarshal([]byte(`{"name":"Renamed"}`), &req); err != nil {
		t.Fatal(err)
	}
	applyApp(&a, req)
	if !a.BindDevice || !a.BindRegion || !a.DiffUpdate || !a.RandomPackage || !a.BruteForce || !a.MtlsEnabled || a.Encryption != 7 {
		t.Fatal("security policy reset", a)
	}
	if err := json.Unmarshal([]byte(`{"name":"Renamed","bindDevice":false,"encryption":0}`), &req); err != nil {
		t.Fatal(err)
	}
	applyApp(&a, req)
	if a.BindDevice || a.Encryption != 0 {
		t.Fatal("explicit disabling ignored")
	}
}

func TestAfterTriggerDatabaseFailureDoesNotRewriteCommittedResponse(t *testing.T) {
	for _, trigger := range []string{"after_login", "after_register", "kami_use", "before_login"} {
		t.Run(trigger, func(t *testing.T) {
			raw, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer raw.Close()
			service := NewSDKService(sqlx.NewDb(raw, "sqlmock"), nil)
			mock.ExpectQuery("SELECT id,appid,name").WillReturnError(errors.New("connection lost"))
			g := gin.New()
			g.GET("/", func(c *gin.Context) {
				c.Set("sdkApp", &repo.App{ID: 1})
				if service.runTrigger(c, trigger, nil, nil) {
					response.OK(c, "committed")
				}
			})
			w := httptest.NewRecorder()
			g.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			var out struct {
				Code int
				Data any
			}
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal("multiple response bodies", w.Body.String())
			}
			if trigger == "before_login" {
				if out.Code != 500 {
					t.Fatal(out)
				}
			} else if out.Code != 0 || out.Data != "committed" {
				t.Fatal(out)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
