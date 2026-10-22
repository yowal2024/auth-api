package buildpool

import (
	"archive/zip"
	"auth-api/internal/buildpool/bsdiff"
	"auth-api/internal/config"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestManifestInflationLimit(t *testing.T) {
	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	f, err := w.Create("AndroidManifest.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(bytes.Repeat([]byte("x"), (8<<20)+1)); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = ExtractPackageName(archive.Bytes()); err == nil || !strings.Contains(err.Error(), "entry too large") {
		t.Fatalf("oversized compressed manifest accepted: %v", err)
	}
}

func TestUpgradeDiffIsolationExpiryAndBusy(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewDiffStore(config.CloneConfig{CanonicalDir: dir, OutputDir: filepath.Join(dir, "app1"), DiffTTLSec: 60})
	b, _ := NewDiffStore(config.CloneConfig{CanonicalDir: dir, OutputDir: filepath.Join(dir, "app2"), DiffTTLSec: 60})
	oldPath, newPath := filepath.Join(dir, "old.apk"), filepath.Join(dir, "new.apk")
	old, new := []byte("old fixture"), []byte("new fixture, longer")
	if err := os.WriteFile(oldPath, old, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, new, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GenerateUpgradeDiff(oldPath, oldPath); err == nil {
		t.Fatal("identical inputs must not produce an upgrade patch")
	}
	meta, err := a.GenerateUpgradeDiff(oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	path, err := a.DiffPath(meta.DiffName)
	if err != nil {
		t.Fatal(err)
	}
	patch, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := bsdiff.Apply(old, patch)
	if err != nil || !bytes.Equal(got, new) || meta.OriginalSha256 != sha(old) || meta.CloneSha256 != sha(got) {
		t.Fatal("upgrade mismatch", err)
	}
	if _, err := b.DiffPath(meta.DiffName); err == nil {
		t.Fatal("cross-app diff accepted")
	}
	for _, name := range []string{"../" + meta.DiffName, "..\\" + meta.DiffName, "/" + meta.DiffName, strings.ReplaceAll(meta.DiffName, "a", "A"), "bad.diff"} {
		if name != meta.DiffName {
			if _, err := a.DiffPath(name); err == nil {
				t.Fatal("invalid name", name)
			}
		}
	}
	buildSlot <- struct{}{}
	_, err = b.GenerateUpgradeDiff(oldPath, newPath)
	<-buildSlot
	if err != ErrBusy {
		t.Fatal("per-request stores bypassed concurrency limit", err)
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	if _, err := a.DiffPath(meta.DiffName); err == nil {
		t.Fatal("expired diff accepted")
	}
}
