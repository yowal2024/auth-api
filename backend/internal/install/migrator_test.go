package install

import (
	"auth-api/internal/config"
	"auth-api/migrations"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSplitSQLQuotesAndComments(t *testing.T) {
	input := "-- comment;\nINSERT INTO t VALUES ('a;b', 'it''s;ok'); /* ; */ SELECT `a;b` FROM t; #skip\nSELECT 'x\\';y';"
	want := []string{"INSERT INTO t VALUES ('a;b', 'it''s;ok')", "SELECT `a;b` FROM t", "SELECT 'x\\';y'"}
	if got := SplitStatements(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("%q", got)
	}
}
func TestAllRecoveredMigrationsParse(t *testing.T) {
	entries, e := fs.ReadDir(migrations.Files, ".")
	if e != nil {
		t.Fatal(e)
	}
	n := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, e := fs.ReadFile(migrations.Files, entry.Name())
		if e != nil {
			t.Fatal(e)
		}
		if len(SplitStatements(string(raw))) == 0 {
			t.Fatal(entry.Name())
		}
		resolved := migrationReplacer("custom_").Replace(string(raw))
		if strings.Contains(resolved, "{users}") || strings.Contains(resolved, "{apps}") {
			t.Fatal("unresolved table token in", entry.Name())
		}
		n++
	}
	if n != 46 {
		t.Fatalf("want 46 migrations, got %d", n)
	}
}
func TestWriteConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	v := config.Config{Installed: true, MySQL: config.MySQLConfig{DSN: "u:p@tcp(localhost:3306)/test", TablePrefix: "a_"}}
	if e := WriteConfig(path, &v); e != nil {
		t.Fatal(e)
	}
	got, e := config.Load(path)
	if e != nil || got.MySQL.DSN != v.MySQL.DSN || !got.Installed {
		t.Fatalf("config mismatch %v", e)
	}
	if _, e = os.Stat(path); e != nil {
		t.Fatal(e)
	}
}
func TestAdminEntryValidation(t *testing.T) {
	for _, bad := range []string{"", "../admin", "install", "sdk", "a/b", "a?b"} {
		if ValidAdminEntry(bad) {
			t.Fatal(bad)
		}
	}
	if !ValidAdminEntry("private_admin-42") {
		t.Fatal("valid path rejected")
	}
}
