package geo

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func TestCacheSwitchUsesFreshCacheOnlyWhenEnabled(t *testing.T) {
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	d := sqlx.NewDb(raw, "sqlmock")
	m.ExpectQuery("SELECT cache_enabled").WillReturnRows(sqlmock.NewRows([]string{"cache_enabled"}).AddRow(1))
	m.ExpectQuery("SELECT lat,lng").WillReturnRows(sqlmock.NewRows([]string{"lat", "lng", "nation", "province", "city", "district", "update_time"}).AddRow(30, 120, "CN", "ZJ", "HZ", "", time.Now().Unix()))
	p, e := Resolve(context.Background(), d, "tencent", "", "1.2.3.4")
	if e != nil || p.Lat != 30 {
		t.Fatal(p, e)
	}
	m.ExpectQuery("SELECT cache_enabled").WillReturnRows(sqlmock.NewRows([]string{"cache_enabled"}).AddRow(0))
	if _, e = Resolve(context.Background(), d, "tencent", "", "1.2.3.4"); e == nil {
		t.Fatal("disabled cache must use provider (missing key)")
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
