package repo

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestBatchDeleteScopesChildDeletionToOwnedUsers(t *testing.T) {
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	d := sqlx.NewDb(raw, "sqlmock")
	InitTablePrefix("a_")
	r := NewUserRepo(d)
	m.ExpectBegin()
	m.ExpectQuery(`SELECT id FROM a_users WHERE appid=\? AND id IN \(\?, \?\) FOR UPDATE`).WithArgs(int64(1000), int64(7), int64(8)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	for _, table := range []string{"user_extends", "user_devices", "user_logs"} {
		m.ExpectExec("DELETE FROM a_" + table + ` WHERE uid IN \(\?\)`).WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	m.ExpectExec(`DELETE FROM a_users WHERE appid=\? AND id IN \(\?\)`).WithArgs(int64(1000), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	n, e := r.BatchDelete(context.Background(), 1000, []int64{7, 8})
	if e != nil || n != 1 {
		t.Fatalf("delete %d %v", n, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestPartialUpdateRejectsIdentifiers(t *testing.T) {
	raw, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	r := NewUserRepo(sqlx.NewDb(raw, "sqlmock"))
	if e = r.UpdatePartial(context.Background(), 1, 2, map[string]any{"vip=0; DROP TABLE users": "x"}); e == nil {
		t.Fatal("untrusted field accepted")
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestPrefixValidation(t *testing.T) {
	for _, p := range []string{"a_; DROP", "a`", "a-", "a.b"} {
		if ValidatePrefix(p) == nil {
			t.Fatal(p)
		}
	}
	if e := InitTablePrefix("tenant_"); e != nil {
		t.Fatal(e)
	}
	if got := SQL("SELECT * FROM {users} WHERE username=?"); got != "SELECT * FROM tenant_users WHERE username=?" {
		t.Fatal(got)
	}
	InitTablePrefix("a_")
}
