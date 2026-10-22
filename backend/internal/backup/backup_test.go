package backup

import (
	"archive/tar"
	"auth-api/internal/config"
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/migrations"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupRestoreRealDatabase(t *testing.T) {
	dsn := os.Getenv("AUTH_RECOVERY_MYSQL_DSN")
	if dsn == "" {
		t.Skip("isolated MySQL integration")
	}
	mc, e := mysql.ParseDSN(dsn)
	if e != nil {
		t.Fatal(e)
	}
	mc.DBName = ""
	root, e := sqlx.Connect("mysql", mc.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	name := fmt.Sprintf("backup_test_%d", time.Now().UnixNano())
	restore := name + "_restore"
	for _, n := range []string{name, restore} {
		if _, e = root.Exec("CREATE DATABASE " + quoteID(n) + " CHARACTER SET utf8mb4"); e != nil {
			t.Fatal(e)
		}
		defer root.Exec("DROP DATABASE " + quoteID(n))
	}
	mc.DBName = name
	if _, e = (install.Migrator{DSN: mc.FormatDSN(), Prefix: "b_", Files: migrations.Files}).Run(context.Background()); e != nil {
		t.Fatal(e)
	}
	d, e := sqlx.Connect("mysql", mc.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	repo.InitTablePrefix("b_")
	if _, e = d.Exec("CREATE TABLE b_probe (id BIGINT PRIMARY KEY,body LONGBLOB,n INT,empty_value TEXT,optional TEXT)"); e != nil {
		t.Fatal(e)
	}
	probe := []byte{0, 39, 92, 10, 255}
	if _, e = d.Exec("INSERT INTO b_probe VALUES (1,?,123,'',NULL)", probe); e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	upload := filepath.Join(dir, "uploads")
	os.Mkdir(upload, 0700)
	os.WriteFile(filepath.Join(upload, "hello.txt"), []byte("upload fixture"), 0600)
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("test-config"), 0600)
	m := New(d, config.Config{MySQL: config.MySQLConfig{DSN: mc.FormatDSN(), TablePrefix: "b_"}, Upload: config.UploadConfig{Root: upload}}, cfgPath)
	defer m.Close()
	v, e := m.Settings(context.Background())
	if e != nil || v.Enabled {
		t.Fatal("must default off", v, e)
	}
	t.Run("InterruptedRecoveryWithSchedulingDisabled", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		d.SetMaxOpenConns(1)
		defer d.SetMaxOpenConns(0)
		if _, err := d.ExecContext(ctx, "INSERT INTO b_backup_runs(id,status,started_at) VALUES ('interrupted','running',1)"); err != nil {
			t.Fatal(err)
		}
		defer d.Exec("DELETE FROM b_backup_runs WHERE id='interrupted'")
		// Simulate another process still working: recovery must leave it alone.
		unlock, err := m.acquireLock(ctx)
		if err != nil {
			t.Fatal(err)
		}
		locked := true
		defer func() {
			if locked {
				unlock()
			}
		}()
		runs, err := m.Runs(ctx)
		if err != nil || len(runs) != 1 || runs[0].Status != "running" {
			t.Fatalf("active worker incorrectly recovered: %+v %v", runs, err)
		}
		unlock()
		locked = false
		runs, err = m.Runs(ctx)
		if err != nil || len(runs) != 1 || runs[0].Status != "failed" || runs[0].FinishedAt == 0 {
			t.Fatalf("interrupted backup stranded: %+v %v", runs, err)
		}
		policy, err := m.Settings(ctx)
		if err != nil || policy.Enabled {
			t.Fatal("recovery changed backup policy", policy, err)
		}
	})
	id := "20261013T120000-0123456789abcdef"
	size, e := m.build(context.Background(), id, v)
	if e != nil || size == 0 {
		t.Fatal(size, e)
	}
	f, e := os.Open(filepath.Join(m.Dir, id+".tar.gz"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	entries := map[string][]byte{}
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(tr)
		if e != nil {
			t.Fatal(e)
		}
		entries[h.Name] = b
	}
	if string(entries["files/uploads/hello.txt"]) != "upload fixture" || string(entries["runtime/config.yaml"]) != "test-config" {
		t.Fatal("missing runtime files")
	}
	mc.DBName = restore
	rd, e := sqlx.Connect("mysql", mc.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer rd.Close()
	conn, e := rd.Connx(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()
	for _, q := range install.SplitStatements(string(entries["database.sql"])) {
		if _, e = conn.ExecContext(context.Background(), q); e != nil {
			t.Fatalf("restore failed: %v, statement %.100s", e, q)
		}
	}
	var got []byte
	var num int
	var empty string
	var isnull bool
	if e = rd.QueryRow("SELECT body,n,empty_value,optional IS NULL FROM b_probe WHERE id=1").Scan(&got, &num, &empty, &isnull); e != nil {
		t.Fatal(e)
	}
	if string(got) != string(probe) || num != 123 || empty != "" || !isnull {
		t.Fatal("restore corrupted values")
	}
	var tables, restored int
	d.Get(&tables, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()")
	rd.Get(&restored, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()")
	if tables != restored {
		t.Fatal(tables, restored)
	}
	if _, e = m.Open(context.Background(), "../config.yaml"); e == nil {
		t.Fatal("traversal accepted")
	}
	if e = m.Save(context.Background(), Settings{IntervalHours: 0, KeepCount: 0}); e == nil {
		t.Fatal("invalid policy accepted")
	}
	// Retention removes only completed backups, preserving a failed run's predecessor.
	gz.Close()
	f.Close()
	for i := 0; i < 3; i++ {
		bid := fmt.Sprintf("20261013T12000%d-0123456789abcdef", i)
		os.WriteFile(filepath.Join(m.Dir, bid+".tar.gz"), []byte("x"), 0600)
		d.Exec("INSERT INTO b_backup_runs(id,status,started_at) VALUES (?,'complete',?)", bid, i+1)
	}
	if e = m.prune(context.Background(), 1); e != nil {
		t.Fatal(e)
	}
	runs, e := m.Runs(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	complete := 0
	for _, r := range runs {
		if r.Status == "complete" {
			complete++
		}
	}
	if complete != 1 {
		t.Fatal(runs)
	}
	if !strings.Contains(string(entries["manifest.json"]), "tablePrefix") {
		t.Fatal("missing manifest")
	}
	// Scheduling is opt-in, works with a one-connection application pool, and
	// launches at most one backup during a due interval.
	d.SetMaxOpenConns(1)
	v.IntervalHours = 1
	v.KeepCount = 2
	v.Enabled = false
	if e = m.Save(context.Background(), v); e != nil {
		t.Fatal(e)
	}
	m.tick()
	var before int
	d.Get(&before, "SELECT COUNT(*) FROM b_backup_runs")
	v.Enabled = true
	if e = m.Save(context.Background(), v); e != nil {
		t.Fatal(e)
	}
	m.tick()
	deadline := time.Now().Add(15 * time.Second)
	for {
		m.mu.Lock()
		busy := m.running
		m.mu.Unlock()
		if !busy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("scheduled backup stuck")
		}
		time.Sleep(20 * time.Millisecond)
	}
	var count int
	d.Get(&count, "SELECT COUNT(*) FROM b_backup_runs")
	if count != before+1 {
		t.Fatal("scheduler did not start", before, count)
	}
	var state string
	d.Get(&state, "SELECT status FROM b_backup_runs ORDER BY started_at DESC LIMIT 1")
	if state != "complete" {
		t.Fatal("scheduled backup failed", state)
	}
	m.tick()
	d.Get(&count, "SELECT COUNT(*) FROM b_backup_runs")
	if count != before+1 {
		t.Fatal("schedule duplicated")
	}
}
