package server

import (
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/migrations"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestApplicationNumbersUpgradeAndConcurrency(t *testing.T) {
	dsn := os.Getenv("AUTH_RECOVERY_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set AUTH_RECOVERY_MYSQL_DSN for isolated integration test")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.DBName = ""
	root, err := sqlx.Connect("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	cfg.DBName = fmt.Sprintf("app_scope_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE DATABASE `" + cfg.DBName + "` DEFAULT CHARACTER SET utf8mb4"); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP DATABASE `" + cfg.DBName + "`")
	old := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() < "0045" {
			raw, err := fs.ReadFile(migrations.Files, entry.Name())
			if err != nil {
				t.Fatal(err)
			}
			old[entry.Name()] = &fstest.MapFile{Data: raw}
		}
	}
	m := install.Migrator{DSN: cfg.FormatDSN(), Prefix: "a_", Files: old}
	ctx := context.Background()
	if _, err = m.Run(ctx); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Connect("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = repo.InitTablePrefix("a_"); err != nil {
		t.Fatal(err)
	}
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(repo.SQL(q), args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("INSERT INTO {apps} (id,name) VALUES (9001,'scope-a'),(9002,'scope-b')")
	exec("INSERT INTO {users} (id,appid,username,ip_list) VALUES (31,9001,'old-a','[]'),(35,9002,'old-b','[]'),(41,9001,'old-a2','[]')")
	exec("INSERT INTO {versions} (id,appid,version_name,version_key,version_no) VALUES (71,9001,'old-a','stable','1'),(79,9002,'old-b','stable','1')")
	for i, row := range []struct {
		route, target string
		app           int64
	}{
		{"/api/admin/user", "", 9001},
		{"/api/private/ver", "", 9002},
		{"/api/private/me/profile", "", 9001},
		{"/api/private/app/:id", "9002", 9001},
	} {
		data, _ := json.Marshal(map[string]any{"route": row.route, "appId": row.app})
		exec("INSERT INTO {logs} (id,type,data,target) VALUES (?,'admin',?,?)", i+1, string(data), row.target)
	}
	exec("INSERT INTO {logs} (id,type,data) VALUES (5,'admin','not-json'),(6,'admin','{}')")
	m.Files = migrations.Files
	applied, err := m.Run(ctx)
	if err != nil || len(applied) != 2 {
		t.Fatalf("upgrade: %v %v", applied, err)
	}
	var scopes []int64
	if err = db.Select(&scopes, repo.SQL("SELECT appid FROM {logs} ORDER BY id")); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(scopes, []int64{9001, 9002, 0, 9002, 0, 0}) {
		t.Fatalf("historical log ownership: %v", scopes)
	}
	var numbers []struct{ AppID, RecordID, LocalID int64 }
	if err = db.Select(&numbers, repo.SQL("SELECT appid,record_id AS recordid,local_id AS localid FROM {app_record_ids} WHERE entity='users' ORDER BY record_id")); err != nil {
		t.Fatal(err)
	}
	if len(numbers) != 3 || numbers[0].LocalID != 1 || numbers[1].LocalID != 1 || numbers[2].LocalID != 2 {
		t.Fatalf("backfilled numbers: %+v", numbers)
	}
	var ids []int64
	if err = db.Select(&ids, repo.SQL("SELECT id FROM {users} ORDER BY id")); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []int64{31, 35, 41}) {
		t.Fatalf("primary IDs changed: %v", ids)
	}
	create := func(app int64) (int64, error) {
		result, err := repo.ExecAppInsert(ctx, db, app, "versions", "INSERT INTO {versions} (appid,version_name,version_key,version_no) VALUES (?,'concurrent','stable','2')", app)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	var wg sync.WaitGroup
	errors := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := create(9001); errors <- err }()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	var localIDs []int64
	if err = db.Select(&localIDs, repo.SQL("SELECT local_id FROM {app_record_ids} WHERE appid=9001 AND entity='versions' ORDER BY local_id")); err != nil {
		t.Fatal(err)
	}
	if len(localIDs) != 13 {
		t.Fatalf("wrong count: %v", localIDs)
	}
	for i, id := range localIDs {
		if id != int64(i+1) {
			t.Fatalf("non-unique/non-contiguous allocation: %v", localIDs)
		}
	}
	// Delete the highest visible number; the next creation must not reuse it.
	exec("DELETE FROM {versions} WHERE id=(SELECT record_id FROM {app_record_ids} WHERE appid=9001 AND entity='versions' AND local_id=13)")
	created, err := create(9001)
	if err != nil {
		t.Fatal(err)
	}
	var number int64
	if err = db.Get(&number, repo.SQL("SELECT local_id FROM {app_record_ids} WHERE appid=9001 AND entity='versions' AND record_id=?"), created); err != nil || number != 14 {
		t.Fatalf("deleted number reused: %d %v", number, err)
	}
	// A failed mapping rolls back the reserved numbers with its transaction.
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.AssignAppRecordIDs(ctx, tx, 9001, "versions", []int64{71}); err == nil {
		t.Fatal("duplicate mapping accepted")
	}
	if err = tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&number, repo.SQL("SELECT last_id FROM {app_record_counters} WHERE appid=9001 AND entity='versions'")); err != nil || number != 14 {
		t.Fatalf("failed transaction consumed a number: %d %v", number, err)
	}
	if err = db.Get(&number, repo.SQL("SELECT last_id FROM {app_record_counters} WHERE appid=9002 AND entity='versions'")); err != nil || number != 1 {
		t.Fatalf("another application's sequence changed: %d %v", number, err)
	}
	if again, err := m.Run(ctx); err != nil || len(again) != 0 {
		t.Fatalf("migration replay: %v %v", again, err)
	}
	t.Log("historical ownership and numbering, unchanged primary IDs, concurrent allocation, deletion and rollback verified")
}
