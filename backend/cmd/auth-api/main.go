package main

import (
	"auth-api/internal/backup"
	"auth-api/internal/buildpool"
	"auth-api/internal/config"
	"auth-api/internal/db"
	"auth-api/internal/geo"
	"auth-api/internal/install"
	"auth-api/internal/server"
	"auth-api/internal/webassets"
	"auth-api/migrations"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configFlag := flag.String("config", "", "configuration path (defaults to configs/config.yaml or config.yaml beside the executable, then the working directory)")
	addrFlag := flag.String("addr", "", "override listen address")
	portFlag := flag.Int("port", 0, "listen on :port; overrides -addr and config server.addr (e.g. -port 9000)")
	migrate := flag.Bool("migrate", false, "apply pending database migrations and exit (installed services also migrate automatically on startup)")
	flag.Parse()
	configPath, err := filepath.Abs(resolveConfigPath(*configFlag))
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	cfg, err := config.Load(configPath)
	if errors.Is(err, os.ErrNotExist) {
		cfg = &config.Config{
			Server:  config.ServerConfig{Addr: ":8080", Mode: "prod"},
			MySQL:   config.MySQLConfig{MaxOpen: 50, MaxIdle: 10, ConnLifetime: 3600, TablePrefix: "a_"},
			Session: config.SessionConfig{TTLSec: 604800},
			Upload:  config.UploadConfig{Root: "./uploads", URLPrefix: "/uploads"},
			Clone:   config.CloneConfig{CanonicalDir: "./uploads/canonical", OutputDir: "./uploads/diffs", DiffTTLSec: 60},
			Log:     config.LogConfig{Level: "info"},
		}
		if err = os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
			return err
		}
		if err = install.WriteConfig(configPath, cfg); err != nil {
			return err
		}
		cfg, err = config.Load(configPath)
	}
	if err != nil {
		return err
	}
	runtimeDir := runtimeDirForConfig(configPath)
	if err = os.Chdir(runtimeDir); err != nil {
		return fmt.Errorf("change runtime directory to %s: %w", runtimeDir, err)
	}
	log.Printf("using config %s (runtime directory %s)", configPath, runtimeDir)
	if *migrate {
		return migrateDatabase(cfg.MySQL)
	}
	listenAddr, err := resolveListenAddress(cfg.Server.Addr, *addrFlag, *portFlag)
	if err != nil {
		return err
	}
	var web fs.FS
	if cfg.Web.Dist != "" {
		web = os.DirFS(cfg.Web.Dist)
	} else {
		web = webassets.FS()
		if web == nil {
			if _, err := os.Stat("webdist/index.html"); err == nil {
				web = os.DirFS("webdist")
			}
		}
	}
	switcher := &server.Swappable{}
	listener := server.NewListenerManager(listenAddr, switcher)
	geoDir := cfg.Geo.LocalDir
	if geoDir == "" {
		geoDir = filepath.Join(runtimeDir, "data", "geo")
	}
	localGeo := geo.NewLocalDB(geoDir)
	defer localGeo.Close()
	installer := install.NewHandler(configPath, cfg, migrations.Files)
	var closers []func()
	var mu sync.Mutex
	defer func() {
		listener.Close()
		mu.Lock()
		closeFunctions := append([]func(){}, closers...)
		mu.Unlock()
		for _, close := range closeFunctions {
			close()
		}
	}()
	buildFull := func() error {
		snapshot := *cfg
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		d, err := db.Open(ctx, snapshot.MySQL)
		if err != nil {
			return err
		}
		r := redis.NewClient(&redis.Options{Addr: snapshot.Redis.Addr, Password: snapshot.Redis.Password, DB: snapshot.Redis.DB})
		ready := false
		defer func() {
			if !ready {
				r.Close()
				d.Close()
			}
		}()
		if err = r.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("connect Redis: %w", err)
		}
		rewriter := &server.RewriteHandler{DB: d}
		refresh := func() error { return rewriter.Refresh(context.Background()) }
		reloadTLS := func() error { return listener.Reload(d) }
		backups := backup.New(d, snapshot, configPath)
		backups.Dir = filepath.Join(runtimeDir, "backups")
		engine, err := server.New(server.Options{Config: &snapshot, DB: d, Redis: r, Installer: installer, Web: web, LocalGeo: localGeo, Backup: backups, CanonicalAdmin: true, OnChange: refresh, OnTLS: reloadTLS})
		if err != nil {
			return err
		}
		rewriter.Next = engine
		if err = rewriter.Refresh(ctx); err != nil {
			return err
		}
		if err = reloadTLS(); err != nil {
			return err
		}
		switcher.Store(rewriter)
		cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
		cleanupDone := make(chan struct{})
		go func() {
			defer close(cleanupDone)
			buildpool.RunCleanup(cleanupCtx, snapshot.Clone, func(err error) { log.Printf("diff cleanup: %v", err) })
		}()
		backups.Start()
		mu.Lock()
		closers = append(closers, func() {
			cleanupCancel()
			<-cleanupDone
			backups.Close()
			r.Close()
			d.Close()
		})
		mu.Unlock()
		ready = true
		return nil
	}
	installer.OnFinalize = buildFull
	if cfg.Installed {
		// Upgrade before handlers, background jobs or the listener can use the new schema.
		if err = migrateDatabase(cfg.MySQL); err != nil {
			return err
		}
		if err = buildFull(); err != nil {
			return err
		}
	} else {
		engine, err := server.New(server.Options{Config: cfg, Installer: installer, Web: web})
		if err != nil {
			return err
		}
		switcher.Store(engine)
		if err = listener.Reload(nil); err != nil {
			return err
		}
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(done)
	select {
	case <-done:
		return nil
	case err := <-listener.Errors():
		return err
	}
}

func migrateDatabase(cfg config.MySQLConfig) error {
	log.Print("checking database migrations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	applied, err := (install.Migrator{DSN: cfg.DSN, Prefix: cfg.TablePrefix, Files: migrations.Files}).Run(ctx)
	if err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	log.Printf("database migrations applied: %v", applied)
	return nil
}

func resolveConfigPath(configPath string) string {
	executablePath, _ := os.Executable()
	workingDir, _ := os.Getwd()
	return resolveConfigPathFrom(configPath, executablePath, workingDir)
}

func resolveConfigPathFrom(configPath, executablePath, workingDir string) string {
	if workingDir == "" {
		workingDir = "."
	}
	if configPath != "" {
		if filepath.IsAbs(configPath) {
			return filepath.Clean(configPath)
		}
		return filepath.Clean(filepath.Join(workingDir, configPath))
	}
	if executablePath == "" {
		executablePath = filepath.Join(workingDir, "auth-api")
	}
	executableDir := filepath.Dir(executablePath)
	candidates := []string{
		filepath.Join(executableDir, "configs", "config.yaml"),
		filepath.Join(executableDir, "config.yaml"),
		filepath.Join(workingDir, "configs", "config.yaml"),
		filepath.Join(workingDir, "config.yaml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); !errors.Is(err, os.ErrNotExist) {
			// Report unreadable configurations instead of silently starting a new installation.
			return filepath.Clean(candidate)
		}
	}
	return filepath.Join(executableDir, "configs", "config.yaml")
}

func runtimeDirForConfig(configPath string) string {
	configDir := filepath.Dir(configPath)
	if filepath.Base(filepath.Clean(configDir)) == "configs" {
		return filepath.Dir(configDir)
	}
	return configDir
}

func resolveListenAddress(configured, override string, port int) (string, error) {
	if port < 0 || port > 65535 {
		return "", fmt.Errorf("invalid listen port %d: must be between 1 and 65535", port)
	}
	if port > 0 {
		// The original -port flag binds all interfaces. Keep it highest priority.
		return ":" + strconv.Itoa(port), nil
	}
	if override != "" {
		return override, nil
	}
	if configured != "" {
		return configured, nil
	}
	return ":8080", nil
}
