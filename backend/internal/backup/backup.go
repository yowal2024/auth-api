package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"auth-api/internal/config"
	"auth-api/internal/repo"
	"github.com/jmoiron/sqlx"
)

type Settings struct {
	Enabled       bool `db:"enabled" json:"enabled"`
	IntervalHours int  `db:"interval_hours" json:"intervalHours"`
	KeepCount     int  `db:"keep_count" json:"keepCount"`
	IncludeFiles  bool `db:"include_files" json:"includeFiles"`
}
type Run struct {
	ID         string `db:"id" json:"id"`
	Status     string `db:"status" json:"status"`
	StartedAt  int64  `db:"started_at" json:"startedAt"`
	FinishedAt int64  `db:"finished_at" json:"finishedAt"`
	Size       int64  `db:"size" json:"size"`
	Error      string `db:"error" json:"error"`
}
type Manager struct {
	DB              *sqlx.DB
	Config          config.Config
	ConfigPath, Dir string
	mu              sync.Mutex
	running         bool
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

var ErrBusy = errors.New("已有备份正在执行")

func New(d *sqlx.DB, cfg config.Config, path string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{DB: d, Config: cfg, ConfigPath: path, Dir: filepath.Join(filepath.Dir(path), "backups"), ctx: ctx, cancel: cancel}
}
func (m *Manager) Settings(ctx context.Context) (Settings, error) {
	var v Settings
	e := m.DB.GetContext(ctx, &v, repo.SQL("SELECT enabled,interval_hours,keep_count,include_files FROM {backup_settings} WHERE id=1"))
	return v, e
}
func (m *Manager) Save(ctx context.Context, v Settings) error {
	if v.IntervalHours < 1 || v.IntervalHours > 720 || v.KeepCount < 1 || v.KeepCount > 30 {
		return errors.New("备份间隔需为 1–720 小时，保留份数需为 1–30")
	}
	_, e := m.DB.ExecContext(ctx, repo.SQL("UPDATE {backup_settings} SET enabled=?,interval_hours=?,keep_count=?,include_files=? WHERE id=1"), v.Enabled, v.IntervalHours, v.KeepCount, v.IncludeFiles)
	return e
}
func (m *Manager) Runs(ctx context.Context) ([]Run, error) {
	v := []Run{}
	e := m.DB.SelectContext(ctx, &v, repo.SQL("SELECT id,status,started_at,finished_at,size,error FROM {backup_runs} ORDER BY started_at DESC,id DESC LIMIT 100"))
	if e == nil {
		for _, run := range v {
			if run.Status == "running" {
				if e = m.recoverInterrupted(ctx); e != nil {
					return nil, e
				}
				e = m.DB.SelectContext(ctx, &v, repo.SQL("SELECT id,status,started_at,finished_at,size,error FROM {backup_runs} ORDER BY started_at DESC,id DESC LIMIT 100"))
				break
			}
		}
	}
	return v, e
}

// A persisted running row is only stale if no process holds the worker lock.
// Recover on reads as well as startup so disabled scheduling cannot strand the UI.
func (m *Manager) recoverInterrupted(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return nil
	}
	unlock, err := m.acquireLock(ctx)
	if errors.Is(err, ErrBusy) {
		return nil
	}
	if err != nil {
		return err
	}
	defer unlock()
	return m.markInterrupted(ctx)
}

func (m *Manager) markInterrupted(ctx context.Context) error {
	_, err := m.DB.ExecContext(ctx, repo.SQL("UPDATE {backup_runs} SET status='failed',finished_at=?,error='服务中断，备份未完成' WHERE status='running'"), time.Now().Unix())
	return err
}

func (m *Manager) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		if err := m.recoverInterrupted(ctx); err != nil {
			log.Printf("backup recovery: %v", err)
		}
		cancel()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.tick()
			}
		}
	}()
}
func (m *Manager) Close() { m.mu.Lock(); m.cancel(); m.mu.Unlock(); m.wg.Wait() }
func (m *Manager) tick() {
	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()
	// Security events have a bounded retention independent of ordinary audit switches.
	_, _ = m.DB.ExecContext(ctx, repo.SQL("DELETE FROM {security_events} WHERE create_time<? LIMIT 10000"), time.Now().Add(-30*24*time.Hour).Unix())
	v, e := m.Settings(ctx)
	if e != nil || !v.Enabled {
		return
	}
	var last int64
	if m.DB.GetContext(ctx, &last, repo.SQL("SELECT COALESCE(MAX(started_at),0) FROM {backup_runs}")) != nil {
		return
	}
	if time.Now().Unix()-last >= int64(v.IntervalHours)*3600 {
		_, _ = m.Trigger(ctx)
	}
}
func (m *Manager) acquireLock(ctx context.Context) (func(), error) {
	// Dedicated MySQL lock also serializes backup workers across server processes.
	lockDB, e := sqlx.Open("mysql", m.Config.MySQL.DSN)
	if e != nil {
		return nil, e
	}
	conn, e := lockDB.Connx(ctx)
	if e != nil {
		lockDB.Close()
		return nil, e
	}
	lock := "auth-backup:" + m.Config.MySQL.TablePrefix
	var got int
	if e = conn.GetContext(ctx, &got, "SELECT GET_LOCK(?,0)", lock); e != nil || got != 1 {
		conn.Close()
		lockDB.Close()
		if e != nil {
			return nil, e
		}
		return nil, ErrBusy
	}
	unlock := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn.ExecContext(cleanup, "SELECT RELEASE_LOCK(?)", lock)
		conn.Close()
		lockDB.Close()
	}
	return unlock, nil
}

func (m *Manager) Trigger(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ctx.Err() != nil {
		return "", m.ctx.Err()
	}
	if m.running {
		return "", ErrBusy
	}
	unlock, e := m.acquireLock(ctx)
	if e != nil {
		return "", e
	}
	v, e := m.Settings(ctx)
	if e != nil {
		unlock()
		return "", e
	}
	b := make([]byte, 8)
	if _, e = rand.Read(b); e != nil {
		unlock()
		return "", e
	}
	id := time.Now().UTC().Format("20060102T150405") + "-" + hex.EncodeToString(b)
	// Any unfinished older job belonged to an interrupted process (we hold the lock).
	e = m.markInterrupted(ctx)
	if e != nil {
		unlock()
		return "", e
	}
	_, e = m.DB.ExecContext(ctx, repo.SQL("INSERT INTO {backup_runs}(id,status,started_at) VALUES (?,'running',?)"), id, time.Now().Unix())
	if e != nil {
		unlock()
		return "", e
	}
	m.running = true
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer unlock()
		defer func() { m.mu.Lock(); m.running = false; m.mu.Unlock() }()
		runCtx, cancel := context.WithTimeout(m.ctx, 30*time.Minute)
		defer cancel()
		size, err := m.build(runCtx, id, v)
		status, msg := "complete", ""
		if err != nil {
			log.Printf("backup %s failed: %v", id, err)
			status = "failed"
			msg = "备份失败，请检查磁盘空间、文件权限或数据库连接"
			if errors.Is(err, context.Canceled) {
				msg = "服务中断，备份未完成"
			}
		}
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer saveCancel()
		_, saveErr := m.DB.ExecContext(saveCtx, repo.SQL("UPDATE {backup_runs} SET status=?,finished_at=?,size=?,error=? WHERE id=?"), status, time.Now().Unix(), size, msg, id)
		if saveErr != nil {
			log.Printf("backup %s status update failed: %v", id, saveErr)
		}
		if err == nil && saveErr == nil {
			if pruneErr := m.prune(saveCtx, v.KeepCount); pruneErr != nil {
				log.Printf("backup retention failed: %v", pruneErr)
			}
		}
	}()
	return id, nil
}

var validID = regexp.MustCompile(`^\d{8}T\d{6}-[a-f0-9]{16}$`)

func (m *Manager) Open(ctx context.Context, id string) (*os.File, error) {
	if !validID.MatchString(id) {
		return nil, fs.ErrNotExist
	}
	var status string
	if e := m.DB.GetContext(ctx, &status, repo.SQL("SELECT status FROM {backup_runs} WHERE id=?"), id); e != nil {
		return nil, e
	}
	if status != "complete" {
		return nil, fs.ErrNotExist
	}
	path := filepath.Join(m.Dir, id+".tar.gz")
	st, e := os.Lstat(path)
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, fs.ErrPermission
	}
	return os.Open(path)
}
func (m *Manager) prune(ctx context.Context, keep int) error {
	rows := []string{}
	if e := m.DB.SelectContext(ctx, &rows, repo.SQL("SELECT id FROM {backup_runs} WHERE status='complete' ORDER BY started_at DESC,id DESC LIMIT 100000 OFFSET ?"), keep); e != nil {
		return e
	}
	for _, id := range rows {
		if !validID.MatchString(id) {
			continue
		}
		e := os.Remove(filepath.Join(m.Dir, id+".tar.gz"))
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if _, e = m.DB.ExecContext(ctx, repo.SQL("UPDATE {backup_runs} SET status='expired' WHERE id=?"), id); e != nil {
			return e
		}
	}
	_, e := m.DB.ExecContext(ctx, repo.SQL("DELETE FROM {backup_runs} WHERE status IN ('expired','failed') AND started_at<?"), time.Now().Add(-90*24*time.Hour).Unix())
	return e
}
func quoteID(v string) string { return "`" + strings.ReplaceAll(v, "`", "``") + "`" }
func (m *Manager) dump(ctx context.Context, w io.Writer) error {
	tx, e := m.DB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if e != nil {
		return e
	}
	defer tx.Rollback()
	tables := []string{}
	if e = tx.SelectContext(ctx, &tables, "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND table_type='BASE TABLE' AND LEFT(table_name,?)=? ORDER BY table_name", len(m.Config.MySQL.TablePrefix), m.Config.MySQL.TablePrefix); e != nil {
		return e
	}
	if _, e = io.WriteString(w, "SET NAMES utf8mb4;\nSET FOREIGN_KEY_CHECKS=0;\nSET time_zone='+00:00';\n"); e != nil {
		return e
	}
	// One consistent snapshot for data. Run schema migrations separately from backups.
	for _, table := range tables {
		var name, ddl string
		if e = tx.QueryRowxContext(ctx, "SHOW CREATE TABLE "+quoteID(table)).Scan(&name, &ddl); e != nil {
			return e
		}
		if _, e = fmt.Fprintf(w, "DROP TABLE IF EXISTS %s;\n%s;\n", quoteID(table), ddl); e != nil {
			return e
		}
		rows, e := tx.QueryxContext(ctx, "SELECT * FROM "+quoteID(table))
		if e != nil {
			return e
		}
		cols, e := rows.Columns()
		if e != nil {
			rows.Close()
			return e
		}
		types, e := rows.ColumnTypes()
		if e != nil {
			rows.Close()
			return e
		}
		quoted := make([]string, len(cols))
		for i, c := range cols {
			quoted[i] = quoteID(c)
		}
		values := make([]sql.RawBytes, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		for rows.Next() {
			if e = rows.Scan(dest...); e != nil {
				rows.Close()
				return e
			}
			if _, e = fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (", quoteID(table), strings.Join(quoted, ",")); e != nil {
				rows.Close()
				return e
			}
			for i, v := range values {
				if i > 0 {
					if _, e = io.WriteString(w, ","); e != nil {
						break
					}
				}
				if v == nil {
					_, e = io.WriteString(w, "NULL")
				} else {
					switch types[i].DatabaseTypeName() {
					case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "BIGINT", "DECIMAL", "FLOAT", "DOUBLE", "YEAR":
						_, e = fmt.Fprintf(w, "CONVERT(X'%x' USING ascii)", []byte(v))
					default:
						_, e = fmt.Fprintf(w, "X'%x'", []byte(v))
					}
				}
				if e != nil {
					break
				}
			}
			if e == nil {
				_, e = io.WriteString(w, ");\n")
			}
			if e != nil {
				rows.Close()
				return e
			}
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
	}
	if _, e = io.WriteString(w, "SET FOREIGN_KEY_CHECKS=1;\n"); e != nil {
		return e
	}
	return tx.Commit()
}

type boundedWriter struct {
	w    io.Writer
	left int64
	ctx  context.Context
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	if e := w.ctx.Err(); e != nil {
		return 0, e
	}
	if int64(len(p)) > w.left {
		return 0, errors.New("backup size limit exceeded")
	}
	n, e := w.w.Write(p)
	w.left -= int64(n)
	return n, e
}
func (m *Manager) build(ctx context.Context, id string, v Settings) (int64, error) {
	if e := os.MkdirAll(m.Dir, 0700); e != nil {
		return 0, e
	}
	temp, e := os.MkdirTemp(m.Dir, ".work-")
	if e != nil {
		return 0, e
	}
	defer os.RemoveAll(temp)
	dumpPath := filepath.Join(temp, "database.sql")
	f, e := os.OpenFile(dumpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return 0, e
	}
	e = m.dump(ctx, &boundedWriter{f, 8 << 30, ctx})
	closeErr := f.Close()
	if e != nil {
		return 0, e
	}
	if closeErr != nil {
		return 0, closeErr
	}
	archive := filepath.Join(temp, "backup.tar.gz")
	out, e := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return 0, e
	}
	defer out.Close()
	gz := gzip.NewWriter(&boundedWriter{out, 10 << 30, ctx})
	tw := tar.NewWriter(gz)
	budget := int64(10 << 30)
	add := func(source, name string) error {
		info, e := os.Lstat(source)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return errors.New("backup refuses nonregular file")
		}
		if info.Size() > budget {
			return errors.New("backup exceeds size limit")
		}
		budget -= info.Size()
		in, e := os.Open(source)
		if e != nil {
			return e
		}
		defer in.Close()
		opened, e := in.Stat()
		if e != nil {
			return e
		}
		if !os.SameFile(info, opened) {
			return errors.New("backup file changed")
		}
		if e = tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: info.Size(), ModTime: info.ModTime()}); e != nil {
			return e
		}
		_, e = io.CopyN(tw, &contextReader{ctx, in}, info.Size())
		return e
	}
	if e = add(dumpPath, "database.sql"); e != nil {
		return 0, e
	}
	if e = add(m.ConfigPath, "runtime/config.yaml"); e != nil {
		return 0, e
	}
	if _, e = os.Lstat("data/master.key"); e == nil {
		if e = add("data/master.key", "runtime/data/master.key"); e != nil {
			return 0, e
		}
	} else if !os.IsNotExist(e) {
		return 0, e
	}
	manifest := map[string]any{"version": 1, "createdAt": time.Now().Unix(), "tablePrefix": m.Config.MySQL.TablePrefix, "includeFiles": v.IncludeFiles, "uploadRoot": m.Config.Upload.Root, "canonicalDir": m.Config.Clone.CanonicalDir, "notes": "SQL snapshot is consistent. Uploaded files are copied during backup; pause file mutations for exact consistency. Redis sessions and regenerated diffs are excluded."}
	if v.IncludeFiles {
		for _, root := range []struct{ source, name string }{{m.Config.Upload.Root, "files/uploads"}, {m.Config.Clone.CanonicalDir, "files/canonical"}} {
			if root.source == "" {
				continue
			}
			if _, e = os.Stat(root.source); os.IsNotExist(e) {
				continue
			} else if e != nil {
				return 0, e
			}
			abs, e := filepath.Abs(root.source)
			if e != nil {
				return 0, e
			}
			backupAbs, _ := filepath.Abs(m.Dir)
			diffAbs, _ := filepath.Abs(m.Config.Clone.OutputDir)
			e = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if e := ctx.Err(); e != nil {
					return e
				}
				if d.Type()&os.ModeSymlink != 0 {
					return errors.New("backup refuses symbolic links")
				}
				if d.IsDir() {
					if p == backupAbs || (m.Config.Clone.OutputDir != "" && p == diffAbs) {
						return filepath.SkipDir
					}
					return nil
				}
				rel, err := filepath.Rel(abs, p)
				if err != nil {
					return err
				}
				return add(p, root.name+"/"+filepath.ToSlash(rel))
			})
			if e != nil {
				return 0, e
			}
		}
	}
	raw, e := json.MarshalIndent(manifest, "", "  ")
	if e != nil {
		return 0, e
	}
	if e = tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(raw))}); e != nil {
		return 0, e
	}
	if _, e = tw.Write(raw); e != nil {
		return 0, e
	}
	if e = tw.Close(); e != nil {
		return 0, e
	}
	if e = gz.Close(); e != nil {
		return 0, e
	}
	if e = out.Sync(); e != nil {
		return 0, e
	}
	info, e := out.Stat()
	if e != nil {
		return 0, e
	}
	if e = out.Close(); e != nil {
		return 0, e
	}
	if e = ctx.Err(); e != nil {
		return 0, e
	}
	if e = os.Rename(archive, filepath.Join(m.Dir, id+".tar.gz")); e != nil {
		return 0, e
	}
	return info.Size(), nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if e := r.ctx.Err(); e != nil {
		return 0, e
	}
	return r.r.Read(p)
}
