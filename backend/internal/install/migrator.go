package install

import (
	"auth-api/internal/repo"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type Migrator struct {
	DSN, Prefix string
	Files       fs.FS
}

// SplitStatements understands SQL strings, identifiers, comments and escaped quotes.
func SplitStatements(s string) []string {
	out := []string{}
	var b strings.Builder
	var quote byte
	line, block := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		var n byte
		if i+1 < len(s) {
			n = s[i+1]
		}
		if line {
			if c == '\n' {
				line = false
				b.WriteByte('\n')
			}
			continue
		}
		if block {
			if c == '*' && n == '/' {
				block = false
				i++
				b.WriteByte(' ')
			}
			continue
		}
		if quote != 0 {
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				i++
				b.WriteByte(s[i])
				continue
			}
			if c == quote {
				if n == quote {
					i++
					b.WriteByte(n)
				} else {
					quote = 0
				}
			}
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			quote = c
			b.WriteByte(c)
			continue
		}
		if c == '#' || c == '-' && n == '-' && (i+2 == len(s) || s[i+2] <= ' ') {
			line = true
			if c == '-' {
				i++
			}
			continue
		}
		if c == '/' && n == '*' {
			block = true
			i++
			continue
		}
		if c == ';' {
			if t := strings.TrimSpace(b.String()); t != "" {
				out = append(out, t)
			}
			b.Reset()
		} else {
			b.WriteByte(c)
		}
	}
	if t := strings.TrimSpace(b.String()); t != "" {
		out = append(out, t)
	}
	return out
}
func (m Migrator) Run(ctx context.Context) ([]string, error) {
	if e := repo.ValidatePrefix(m.Prefix); e != nil {
		return nil, e
	}
	d, e := sql.Open("mysql", m.DSN)
	if e != nil {
		return nil, e
	}
	defer d.Close()
	conn, e := d.Conn(ctx)
	if e != nil {
		return nil, e
	}
	defer conn.Close()
	// MySQL DDL autocommits. Refuse an untracked nonempty schema instead of replaying DROP TABLE.
	lock := "auth-api-migrate:" + m.Prefix
	var locked int
	if e = conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 10)", lock).Scan(&locked); e != nil || locked != 1 {
		return nil, errors.New("cannot acquire migration lock")
	}
	defer conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lock)
	journal := m.Prefix + "_migrations"
	var tracked int
	e = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", journal).Scan(&tracked)
	if e != nil {
		return nil, e
	}
	if tracked == 0 {
		var n int
		e = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,?)=?", len(m.Prefix), m.Prefix).Scan(&n)
		if e != nil {
			return nil, e
		}
		if n > 0 {
			return nil, errors.New("数据库中已有同前缀表但没有迁移记录；拒绝重建，请使用空测试数据库")
		}
	}
	_, e = conn.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS `"+journal+"` (version VARCHAR(128) NOT NULL PRIMARY KEY, applied_at BIGINT NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	if e != nil {
		return nil, e
	}
	entries, e := fs.ReadDir(m.Files, ".")
	if e != nil {
		return nil, e
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	applied := []string{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := entry.Name()
		var ts int64
		e = conn.QueryRowContext(ctx, "SELECT applied_at FROM `"+journal+"` WHERE version=?", name).Scan(&ts)
		if e == nil {
			if ts < 0 {
				return applied, fmt.Errorf("迁移 %s 曾中断，需检查后修复记录，禁止自动重放", name)
			}
			continue
		}
		if !errors.Is(e, sql.ErrNoRows) {
			return applied, e
		}
		raw, e := fs.ReadFile(m.Files, name)
		if e != nil {
			return applied, e
		}
		_, e = conn.ExecContext(ctx, "INSERT INTO `"+journal+"` (version,applied_at) VALUES (?,-1)", name)
		if e != nil {
			return applied, e
		}
		replacement := migrationReplacer(m.Prefix)
		for i, q := range SplitStatements(replacement.Replace(string(raw))) {
			if _, e = conn.ExecContext(ctx, q); e != nil {
				var me *mysql.MySQLError
				if errors.As(e, &me) && (me.Number == 1060 || me.Number == 1061 || me.Number == 1091) {
					continue
				}
				return applied, fmt.Errorf("migration %s statement %d: %w", name, i+1, e)
			}
		}
		_, e = conn.ExecContext(ctx, "UPDATE `"+journal+"` SET applied_at=? WHERE version=?", time.Now().Unix(), name)
		if e != nil {
			return applied, e
		}
		applied = append(applied, name)
	}
	return applied, nil
}

func migrationReplacer(prefix string) *strings.Replacer {
	tables := []string{"admins", "apps", "users", "user_extends", "user_devices", "user_logs", "notices", "extends", "versions", "blocks", "kami_groups", "kamis", "functions", "encrypts", "logs", "settings", "api_routes", "api_codes", "user_groups", "ip_info", "extend_files", "security_events", "backup_settings", "backup_runs"}
	pairs := []string{}
	for _, table := range tables {
		pairs = append(pairs, "{"+table+"}", prefix+table)
	}
	return strings.NewReplacer(pairs...)
}
