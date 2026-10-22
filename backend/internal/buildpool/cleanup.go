package buildpool

import (
	"auth-api/internal/config"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var diffFilename = regexp.MustCompile("^[0-9a-f]{64}_[0-9a-f]{64}[.]diff$")

// SweepTree includes per-application directories and never follows symlinks.
func SweepTree(cfg config.CloneConfig) error {
	root := cfg.OutputDir
	if root == "" {
		root = "uploads/diffs"
	}
	ttl := cfg.DiffTTLSec
	if ttl <= 0 {
		ttl = 60
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() || !diffFilename.MatchString(entry.Name()) {
			return nil
		}
		info, err := entry.Info()
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if time.Since(info.ModTime()) > time.Duration(ttl)*time.Second {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		return nil
	})
}

func RunCleanup(ctx context.Context, cfg config.CloneConfig, report func(error)) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if err := SweepTree(cfg); err != nil && report != nil {
			report(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
