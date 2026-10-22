package buildpool

import (
	"auth-api/internal/config"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCleanupExpiryAndCancellation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "1000")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(dir, strings.Repeat("a", 64)+"_"+strings.Repeat("b", 64)+".diff")
	fresh := filepath.Join(dir, strings.Repeat("c", 64)+"_"+strings.Repeat("d", 64)+".diff")
	other := filepath.Join(dir, "keep.diff")
	for _, path := range []string{old, fresh, other} {
		if err := os.WriteFile(path, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(other, past, past); err != nil {
		t.Fatal(err)
	}
	cfg := config.CloneConfig{OutputDir: root, DiffTTLSec: 60}
	if err := SweepTree(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("expired diff remains", err)
	}
	for _, path := range []string{fresh, other} {
		if _, err := os.Stat(path); err != nil {
			t.Fatal("live or unrelated file removed", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { RunCleanup(ctx, cfg, nil); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not stop")
	}
}
