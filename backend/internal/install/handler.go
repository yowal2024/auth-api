// Reconstructed installation flow; migration guards are deliberately stricter.
package install

import (
	"auth-api/internal/auth"
	"auth-api/internal/config"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func NewHandler(path string, cfg *config.Config, files fs.FS) *Handler {
	return &Handler{configPath: path, cfg: cfg, migrations: files}
}
func Mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/status", h.Status)
	g.POST("/test-mysql", h.TestMySQL)
	g.POST("/test-redis", h.TestRedis)
	g.POST("/save-config", h.SaveConfig)
	g.POST("/migrate", h.Migrate)
	g.POST("/create-admin", h.CreateAdmin)
	g.POST("/finalize", h.Finalize)
}
func (h *Handler) guardLocked(c *gin.Context) bool {
	if h.cfg.Installed {
		response.Fail(c, 403, "系统已安装,不允许重复安装")
		return false
	}
	return true
}
func (h *Handler) Status(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	entry := "admin"
	if h.AdminEntryFunc != nil {
		entry = h.AdminEntryFunc()
	}
	// Never advertise the hidden entry from a public discovery endpoint.
	// A visitor who already knows the entry may validate it for SPA routing.
	data := gin.H{"installed": h.cfg.Installed}
	if candidate := c.GetHeader("X-Admin-Entry"); candidate != "" && candidate == entry {
		data["adminEntry"] = candidate
	}
	c.Header("Cache-Control", "no-store")
	response.OK(c, data)
}
func BuildMySQLDSN(host string, port int, user, password, database string) string {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 3306
	}
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	cfg.User = user
	cfg.Passwd = password
	cfg.DBName = database
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.Timeout = 5 * time.Second
	return cfg.FormatDSN()
}
func mysqlDSN(v testMySQLReq) string {
	return BuildMySQLDSN(v.Host, v.Port, v.User, v.Password, v.Database)
}
func pingMySQL(ctx context.Context, dsn string) error {
	d, e := sql.Open("mysql", dsn)
	if e != nil {
		return e
	}
	defer d.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return d.PingContext(ctx)
}
func pingRedis(ctx context.Context, v testRedisReq) error {
	r := redis.NewClient(&redis.Options{Addr: v.Addr, Password: v.Password, DB: v.DB, DialTimeout: 5 * time.Second})
	defer r.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.Ping(ctx).Err()
}
func (h *Handler) TestMySQL(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	var v testMySQLReq
	if c.ShouldBindJSON(&v) != nil || v.User == "" || v.Database == "" {
		response.BadRequest(c, "用户名和数据库名必填")
		return
	}
	if e := pingMySQL(c.Request.Context(), mysqlDSN(v)); e != nil {
		response.BadRequest(c, "MySQL 连接失败，请检查地址、账号和数据库")
		return
	}
	response.OK(c, gin.H{"ok": true})
}
func (h *Handler) TestRedis(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	var v testRedisReq
	if c.ShouldBindJSON(&v) != nil || v.Addr == "" {
		response.BadRequest(c, "Redis 地址必填")
		return
	}
	if e := pingRedis(c.Request.Context(), v); e != nil {
		response.BadRequest(c, "Redis 连接失败")
		return
	}
	response.OK(c, gin.H{"ok": true})
}
func (h *Handler) SaveConfig(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	var v saveConfigReq
	if c.ShouldBindJSON(&v) != nil || v.MySQL.User == "" || v.MySQL.Database == "" || v.Redis.Addr == "" || repo.ValidatePrefix(v.MySQL.TablePrefix) != nil {
		response.BadRequest(c, "数据库配置不完整或表前缀非法")
		return
	}
	next := *h.cfg
	next.MySQL.DSN = mysqlDSN(v.MySQL)
	next.MySQL.TablePrefix = v.MySQL.TablePrefix
	next.Redis = config.RedisConfig{Addr: v.Redis.Addr, DB: v.Redis.DB, Password: v.Redis.Password}
	if v.Server != nil {
		next.Server.Addr = v.Server.Addr
		next.Server.Mode = v.Server.Mode
	}
	if e := WriteConfig(h.configPath, &next); e != nil {
		response.ServerError(c, "保存配置失败")
		return
	}
	*h.cfg = next
	response.OK(c, gin.H{"ok": true})
}
func (h *Handler) Migrate(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	if h.cfg.MySQL.DSN == "" {
		response.BadRequest(c, "尚未保存 MySQL 配置")
		return
	}
	v, e := (Migrator{h.cfg.MySQL.DSN, h.cfg.MySQL.TablePrefix, h.migrations}).Run(c.Request.Context())
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, gin.H{"applied": v})
}
func ValidAdminEntry(s string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`).MatchString(s) && s != "install" && s != "sdk"
}
func (h *Handler) CreateAdmin(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	var v createAdminReq
	if c.ShouldBindJSON(&v) != nil || len(v.Username) < 5 || len(v.Username) > 12 || len(v.Password) < 6 || len(v.Password) > 18 || !ValidAdminEntry(v.AdminEntry) {
		response.BadRequest(c, "账号需 5-12 位、密码需 6-18 位，入口仅限字母数字下划线或横线")
		return
	}
	if e := h.createAdmin(c.Request.Context(), v); e != nil {
		response.ServerError(c, "创建管理员失败（已有超级管理员时禁止覆盖）")
		return
	}
	response.OK(c, gin.H{"ok": true})
}
func (h *Handler) createAdmin(ctx context.Context, v createAdminReq) error {
	if e := repo.ValidatePrefix(h.cfg.MySQL.TablePrefix); e != nil {
		return e
	}
	d, e := sql.Open("mysql", h.cfg.MySQL.DSN)
	if e != nil {
		return e
	}
	defer d.Close()
	tx, e := d.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	p := h.cfg.MySQL.TablePrefix
	// Serialize initial admin creation through the single settings row.
	var sid int
	if e = tx.QueryRowContext(ctx, "SELECT id FROM `"+p+"settings` WHERE id=1 FOR UPDATE").Scan(&sid); e != nil {
		return e
	}
	var n int
	if e = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+p+"admins` WHERE role='super'").Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		return errors.New("super admin already exists")
	}
	hash, e := auth.HashPassword(v.Password)
	if e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO `"+p+"admins` (username,password,role,last_login,last_ip) VALUES (?,?,'super',0,'')", v.Username, hash); e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, "UPDATE `"+p+"settings` SET admin_entry=? WHERE id=1", v.AdminEntry); e != nil {
		return e
	}
	return tx.Commit()
}
func (h *Handler) Finalize(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.guardLocked(c) {
		return
	}
	if e := repo.ValidatePrefix(h.cfg.MySQL.TablePrefix); e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	d, e := sql.Open("mysql", h.cfg.MySQL.DSN)
	if e != nil {
		response.ServerError(c, "连接数据库失败")
		return
	}
	defer d.Close()
	var n int
	e = d.QueryRowContext(c.Request.Context(), "SELECT COUNT(*) FROM `"+h.cfg.MySQL.TablePrefix+"admins` WHERE role='super'").Scan(&n)
	if e != nil || n == 0 {
		response.BadRequest(c, "未检测到超级管理员账号,请先创建")
		return
	}
	if e = pingRedis(c.Request.Context(), testRedisReq{h.cfg.Redis.Addr, h.cfg.Redis.DB, h.cfg.Redis.Password}); e != nil {
		response.ServerError(c, "Redis 连接失败")
		return
	}
	next := *h.cfg
	next.Installed = true
	if e = WriteConfig(h.configPath, &next); e != nil {
		response.ServerError(c, "保存配置失败")
		return
	}
	*h.cfg = next
	restart := h.OnFinalize == nil
	if h.OnFinalize != nil {
		restart = h.OnFinalize() != nil
	}
	response.OK(c, gin.H{"ok": true, "restartRequired": restart})
}
func WriteConfig(path string, c *config.Config) error {
	b, e := yaml.Marshal(c)
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".config-*.yaml.tmp")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(b)
	}
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return fmt.Errorf("write config: %w", e)
	}
	return os.Rename(name, path)
}
