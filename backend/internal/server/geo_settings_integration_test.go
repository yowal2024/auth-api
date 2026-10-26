package server

import (
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/migrations"
	"context"
	"fmt"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestGlobalGeoSettingsUpgrade(t *testing.T) {
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
	old := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() < "0047" {
			body, err := fs.ReadFile(migrations.Files, entry.Name())
			if err != nil {
				t.Fatal(err)
			}
			old[entry.Name()] = &fstest.MapFile{Data: body}
		}
	}
	if err := repo.InitTablePrefix("a_"); err != nil {
		t.Fatal(err)
	}
	type prior struct {
		provider, key string
		bound         bool
	}
	for _, tc := range []struct {
		name string
		apps []prior
		want repo.GeoSettings
	}{
		{"new installation", nil, repo.GeoSettings{Provider: "dbip"}},
		{"existing Tencent key", []prior{{"dbip", "", false}, {"tencent", "ExistingKey", true}}, repo.GeoSettings{Provider: "tencent", Key: "ExistingKey"}},
		{"shared disabled service", []prior{{"amap", "SharedKey", false}, {"amap", "SharedKey", false}}, repo.GeoSettings{Provider: "amap", Key: "SharedKey"}},
		{"conflicting services", []prior{{"tencent", "one", true}, {"amap", "two", true}}, repo.GeoSettings{}},
		{"case sensitive keys", []prior{{"tencent", "ApiKey", true}, {"tencent", "apikey", true}}, repo.GeoSettings{}},
		{"local ignores unused key", []prior{{"dbip", "old-key", true}, {"dbip", "", true}}, repo.GeoSettings{Provider: "dbip"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfg.DBName = fmt.Sprintf("geo_upgrade_%d", time.Now().UnixNano())
			if _, err := root.Exec("CREATE DATABASE `" + cfg.DBName + "` DEFAULT CHARACTER SET utf8mb4"); err != nil {
				t.Fatal(err)
			}
			defer root.Exec("DROP DATABASE `" + cfg.DBName + "`")
			migrator := install.Migrator{DSN: cfg.FormatDSN(), Prefix: "a_", Files: old}
			if _, err := migrator.Run(ctx); err != nil {
				t.Fatal(err)
			}
			db, err := sqlx.Connect("mysql", cfg.FormatDSN())
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			for i, app := range tc.apps {
				if _, err := db.Exec("INSERT INTO a_apps (name,geo_provider,geo_key,bind_region,ip_distance_max) VALUES (?,?,?,?,?)", fmt.Sprintf("App %d", i), app.provider, app.key, app.bound, 50+i); err != nil {
					t.Fatal(err)
				}
			}
			migrator.Files = migrations.Files
			applied, err := migrator.Run(ctx)
			if err != nil || len(applied) != 1 || applied[0] != "0047_system_geo.sql" {
				t.Fatalf("upgrade: %v %v", applied, err)
			}
			settings, err := repo.LoadGeoSettings(ctx, db)
			if err != nil || settings != tc.want {
				t.Fatalf("global configuration mismatch: provider=%q error=%v", settings.Provider, err)
			}
			var apps []repo.App
			if err := db.Select(&apps, "SELECT name,geo_provider,geo_key,bind_region,ip_distance_max FROM a_apps ORDER BY id"); err != nil {
				t.Fatal(err)
			}
			for i, app := range apps {
				if app.GeoProvider != tc.apps[i].provider || app.GeoKey != tc.apps[i].key || app.BindRegion != tc.apps[i].bound || app.IPDistanceMax != 50+i {
					t.Fatal("upgrade changed an application's existing configuration")
				}
				if settings.Provider == "" {
					effective := settings.ForApp(&app)
					if effective.Provider != app.GeoProvider || effective.Key != app.GeoKey {
						t.Fatal("conflicting application lost its existing service")
					}
				}
			}
			chosen := repo.GeoSettings{Provider: "amap", Key: "UnifiedKey"}
			if err := repo.SaveGeoSettings(ctx, db, chosen); err != nil {
				t.Fatal(err)
			}
			// Observe through the same transaction path used by SDK region checks.
			tx, err := db.BeginTxx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			current, err := repo.LoadGeoSettings(ctx, tx)
			if err != nil || current != chosen {
				t.Fatal("saved service was not immediately visible", err)
			}
			for _, app := range apps {
				if current.ForApp(&app) != chosen {
					t.Fatal("application ignored unified configuration")
				}
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if again, err := migrator.Run(ctx); err != nil || len(again) != 0 {
				t.Fatalf("migration replayed: %v %v", again, err)
			}
		})
	}
}

func testSystemGeoAPI(t *testing.T, db *sqlx.DB, token string, request func(string, string, any, string, bool) map[string]any) {
	t.Helper()
	ctx := context.Background()
	before, err := repo.LoadGeoSettings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := repo.SaveGeoSettings(ctx, db, before); err != nil {
			t.Error(err)
		}
	}()
	call := func(method, path string, body any, want int) any {
		t.Helper()
		result := request(method, "/api/admin"+path, body, token, false)
		if result["code"] != float64(want) {
			t.Fatalf("%s %s: %v", method, path, result)
		}
		return result["data"]
	}
	for _, body := range []map[string]any{
		{"provider": "dbip"}, {"provider": "tencent", "key": " "}, {"provider": "unknown", "key": "key"},
	} {
		call("POST", "/system/geo", body, 400)
		current, err := repo.LoadGeoSettings(ctx, db)
		if err != nil || current != before {
			t.Fatal("invalid settings changed the active service", err)
		}
	}
	call("POST", "/system/geo", map[string]any{"provider": "tencent", "key": "fixture-system-key"}, 0)
	global := call("GET", "/system/geo", nil, 0).(map[string]any)
	if global["provider"] != "tencent" || global["key"] != "fixture-system-key" || len(global["legacyApps"].([]any)) != 0 {
		t.Fatal("system configuration did not round trip")
	}
	ids := []int64{}
	for _, distance := range []int{25, 100} {
		// The app's legacy local provider is overridden by the configured system service.
		app := call("POST", "/app", map[string]any{"name": "Global geo", "bindRegion": true, "geoProvider": "dbip", "ipDistanceMax": distance}, 0).(map[string]any)
		id := int64(app["id"].(float64))
		ids = append(ids, id)
		defer call("DELETE", fmt.Sprintf("/app/%d", id), nil, 0)
	}
	call("POST", "/system/geo", map[string]any{"provider": "amap", "key": "another-system-key"}, 0)
	for i, id := range ids {
		app, err := repo.NewAppRepo(db).GetByID(ctx, id)
		if err != nil || !app.BindRegion || app.IPDistanceMax != []int{25, 100}[i] {
			t.Fatal("global service changed application binding rules", err)
		}
		effective, err := repo.NewAppRepo(db).GeoSettings(ctx, app)
		if err != nil || effective.Provider != "amap" || effective.Key != "another-system-key" {
			t.Fatal("application retained the old service", err)
		}
	}
}
