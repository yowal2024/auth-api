package buildpool

import (
	"archive/zip"
	"auth-api/internal/buildpool/apkzip"
	"auth-api/internal/buildpool/axml"
	"auth-api/internal/buildpool/bsdiff"
	"auth-api/internal/buildpool/edge"
	"auth-api/internal/buildpool/v2sign"
	"auth-api/internal/config"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Stores are constructed per request; their instance mutex cannot bound total work.
var buildSlot = make(chan struct{}, 1)
var ErrBusy = fmt.Errorf("APK builder busy; retry later")

func readAPK(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (200<<20)+1))
	if len(b) > 200<<20 {
		return nil, fmt.Errorf("APK too large")
	}
	return b, err
}

func NewDiffStore(c config.CloneConfig) (*DiffStore, error) {
	if c.CanonicalDir == "" {
		c.CanonicalDir = "uploads/canonical"
	}
	if c.OutputDir == "" {
		c.OutputDir = "uploads/diffs"
	}
	if c.DiffTTLSec <= 0 {
		c.DiffTTLSec = 60
	}
	for _, p := range []string{c.CanonicalDir, c.OutputDir} {
		if e := os.MkdirAll(p, 0700); e != nil {
			return nil, e
		}
	}
	return &DiffStore{cfg: c}, nil
}
func ReadEntry(apk []byte, name string) ([]byte, error) {
	r, e := zip.NewReader(bytes.NewReader(apk), int64(len(apk)))
	if e != nil {
		return nil, e
	}
	for _, f := range r.File {
		if f.Name == name {
			limit := int64(128 << 20)
			if name == "AndroidManifest.xml" {
				limit = 8 << 20
			}
			if f.UncompressedSize64 > uint64(limit) {
				return nil, fmt.Errorf("entry too large")
			}
			in, e := f.Open()
			if e != nil {
				return nil, e
			}
			defer in.Close()
			b, err := io.ReadAll(io.LimitReader(in, limit+1))
			if int64(len(b)) > limit {
				return nil, fmt.Errorf("entry too large")
			}
			return b, err
		}
	}
	return nil, fmt.Errorf("%s missing", name)
}
func ExtractPackageName(apk []byte) (string, error) {
	manifest, e := ReadEntry(apk, "AndroidManifest.xml")
	if e != nil {
		return "", e
	}
	return axml.ReadOriginalPackage(manifest)
}
func randomPackage() (string, error) {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return "com.a" + hex.EncodeToString(b[:6]) + ".b" + hex.EncodeToString(b[6:]), nil
}
func sha(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func (s *DiffStore) GenerateFromBytes(apk []byte, key *v2sign.KeyAndCert) (Meta, error) {
	select {
	case buildSlot <- struct{}{}:
	default:
		return Meta{}, ErrBusy
	}
	defer func() { <-buildSlot }()
	return s.generateFromBytes(apk, key)
}
func (s *DiffStore) generateFromBytes(apk []byte, key *v2sign.KeyAndCert) (Meta, error) {
	if len(apk) > 200<<20 {
		return Meta{}, fmt.Errorf("APK too large")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var meta Meta
	old, e := ExtractPackageName(apk)
	if e != nil {
		return meta, e
	}
	pkg, e := randomPackage()
	if e != nil {
		return meta, e
	}
	manifest, e := ReadEntry(apk, "AndroidManifest.xml")
	if e != nil {
		return meta, e
	}
	classes := map[string]bool{}
	r, e := zip.NewReader(bytes.NewReader(apk), int64(len(apk)))
	if e != nil {
		return meta, e
	}
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "classes") && strings.HasSuffix(f.Name, ".dex") {
			b, e := ReadEntry(apk, f.Name)
			if e != nil {
				return meta, e
			}
			for k := range axml.DexClasses(b) {
				classes[k] = true
			}
		}
	}
	patched, e := axml.PatchManifest(manifest, old, pkg, classes)
	if e != nil {
		return meta, e
	}
	repacked, e := apkzip.Repack(apk, patched)
	if e != nil {
		return meta, e
	}
	if key == nil {
		key, e = v2sign.GenerateKeyAndCert()
		if e != nil {
			return meta, e
		}
	}
	clone, e := v2sign.SignV2(repacked, key)
	if e != nil {
		return meta, e
	}
	originalPairs, e := edge.Pairs(apk)
	if e != nil {
		return meta, e
	}
	if len(originalPairs) > 0 {
		clone, e = preservePairs(clone, originalPairs)
		if e != nil {
			return meta, e
		}
	}
	diff, e := bsdiff.Build(apk, clone)
	if e != nil {
		return meta, e
	}
	meta = Meta{NewPackage: pkg, OriginalSha256: sha(apk), CloneSha256: sha(clone)}
	meta.DiffName = meta.OriginalSha256 + "_" + meta.CloneSha256 + ".diff"
	if e = writeDiffAtomic(filepath.Join(s.cfg.OutputDir, meta.DiffName), diff); e != nil {
		return meta, e
	}
	return meta, nil
}
func (s *DiffStore) Generate(path string, key *v2sign.KeyAndCert) (Meta, error) {
	select {
	case buildSlot <- struct{}{}:
	default:
		return Meta{}, ErrBusy
	}
	defer func() { <-buildSlot }()
	b, e := readAPK(path)
	if e != nil {
		return Meta{}, e
	}
	return s.generateFromBytes(b, key)
}
func (s *DiffStore) GenerateUpgradeDiff(oldPath, newPath string) (Meta, error) {
	select {
	case buildSlot <- struct{}{}:
	default:
		return Meta{}, ErrBusy
	}
	defer func() { <-buildSlot }()
	old, e := readAPK(oldPath)
	if e != nil {
		return Meta{}, e
	}
	new, e := readAPK(newPath)
	if e != nil {
		return Meta{}, e
	}
	meta := Meta{OriginalSha256: sha(old), CloneSha256: sha(new)}
	if meta.OriginalSha256 == meta.CloneSha256 {
		return meta, fmt.Errorf("upgrade inputs are identical")
	}
	meta.DiffName = meta.OriginalSha256 + "_" + meta.CloneSha256 + ".diff"
	if _, err := s.DiffPath(meta.DiffName); err == nil {
		return meta, nil
	}
	diff, e := bsdiff.Build(old, new)
	if e != nil {
		return Meta{}, e
	}
	e = writeDiffAtomic(filepath.Join(s.cfg.OutputDir, meta.DiffName), diff)
	return meta, e
}
func (s *DiffStore) DiffPath(name string) (string, error) {
	if !regexp.MustCompile(`^[0-9a-f]{64}_[0-9a-f]{64}\.diff$`).MatchString(name) {
		return "", fmt.Errorf("invalid diff name")
	}
	path := filepath.Join(s.cfg.OutputDir, name)
	info, e := os.Lstat(path)
	if e != nil {
		return "", e
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("invalid diff file")
	}
	if time.Since(info.ModTime()) > time.Duration(s.cfg.DiffTTLSec)*time.Second {
		os.Remove(path)
		return "", os.ErrNotExist
	}
	return path, nil
}
func (s *DiffStore) SweepStale() error {
	entries, e := os.ReadDir(s.cfg.OutputDir)
	if e != nil {
		return e
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".diff") {
			continue
		}
		info, e := entry.Info()
		if e == nil && time.Since(info.ModTime()) > time.Duration(s.cfg.DiffTTLSec)*time.Second {
			_ = os.Remove(filepath.Join(s.cfg.OutputDir, entry.Name()))
		}
	}
	return nil
}
