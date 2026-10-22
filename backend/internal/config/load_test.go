package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvironmentOverrides(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(p, []byte("server:\n  addr: ':8080'\nredis:\n  db: 2\nsession:\n  ttl_sec: 100\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AUTH_API_ADDR", "127.0.0.1:9000")
	t.Setenv("AUTH_API_REDIS_DB", "invalid")
	t.Setenv("AUTH_API_SESSION_TTL", "120")
	t.Setenv("AUTH_API_MODE", "")
	t.Setenv("AUTH_API_MYSQL_DSN", "")
	t.Setenv("AUTH_API_REDIS_ADDR", "")
	t.Setenv("AUTH_API_REDIS_PASSWORD", "")
	t.Setenv("AUTH_API_SESSION_ABSOLUTE_TTL", "")
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Server.Addr != "127.0.0.1:9000" || c.Redis.DB != 2 || c.Session.TTLSec != 120 {
		t.Fatal("incorrect overrides")
	}
}
