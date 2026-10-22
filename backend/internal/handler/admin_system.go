package handler

import (
	"auth-api/internal/install"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
)

func MountSystem(g *gin.RouterGroup, d *sqlx.DB, r *repo.SettingsRepo, onChange func() error) {
	g.POST("/system/settings", func(c *gin.Context) {
		var v settingsReq
		if c.ShouldBindJSON(&v) != nil || !install.ValidAdminEntry(v.AdminEntry) || v.UploadLimit < 1 || v.UploadLimit > 200 || v.AdminTokenExpire < 60 || v.MysqlMaxConn < 1 || v.MysqlMaxIdle < 0 || v.MysqlMaxIdle > v.MysqlMaxConn {
			response.BadRequest(c, "设置参数无效")
			return
		}
		if v.LogLevel == "info" || v.LogLevel == "" {
			v.LogLevel = "all"
		}
		if v.LogLevel != "all" && v.LogLevel != "warning" && v.LogLevel != "error" {
			response.BadRequest(c, "日志级别无效")
			return
		}
		if v.MysqlMaxConn > 10000 || v.MysqlMaxIdle > 10000 {
			response.BadRequest(c, "连接池设置超出范围")
			return
		}
		s, e := r.Get(c.Request.Context())
		if dbError(c, e) {
			return
		}
		s.WebURL = v.WebURL
		s.UploadLimit = v.UploadLimit
		s.AdminTokenExpire = v.AdminTokenExpire
		s.AdminEntry = v.AdminEntry
		s.MysqlMaxConn = v.MysqlMaxConn
		s.MysqlMaxIdle = v.MysqlMaxIdle
		s.LogLevel = v.LogLevel
		boolInt := func(b bool) int {
			if b {
				return 1
			}
			return 0
		}
		if v.CacheEnabled != nil {
			s.CacheEnabled = boolInt(*v.CacheEnabled)
		}
		if v.UserLogEnabled != nil {
			s.UserLogEnabled = boolInt(*v.UserLogEnabled)
		}
		if v.AdminLogEnabled != nil {
			s.AdminLogEnabled = boolInt(*v.AdminLogEnabled)
		}
		if dbError(c, r.Save(c.Request.Context(), s)) {
			return
		}
		d.SetMaxOpenConns(s.MysqlMaxConn)
		d.SetMaxIdleConns(s.MysqlMaxIdle)
		if onChange != nil {
			if e = onChange(); e != nil {
				c.Set("audit.hotSwapFailed", true)
				response.OK(c, gin.H{"adminEntry": s.AdminEntry, "hotSwap": false, "hotSwapError": e.Error()})
				return
			}
		}
		response.OK(c, gin.H{"adminEntry": s.AdminEntry, "hotSwap": true})
	})
	for _, kind := range []string{"routes", "codes"} {
		kind := kind
		table := "api_" + kind
		base := "/system/api-" + kind
		list := func(c *gin.Context) {
			if kind == "routes" {
				v := []repo.APIRoute{}
				e := d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,api_key,name,override_path,rewritten,sort_order FROM {api_routes} ORDER BY sort_order,id"))
				if !dbError(c, e) {
					response.OK(c, v)
				}
			} else {
				v := []repo.APICode{}
				e := d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,code,override_code,message,remark,rewritten FROM {api_codes} ORDER BY id"))
				if !dbError(c, e) {
					response.OK(c, v)
				}
			}
		}
		g.GET(base, list)
		g.POST(base+"/save", func(c *gin.Context) {
			var body struct {
				Items []map[string]any `json:"items"`
			}
			if c.ShouldBindJSON(&body) != nil {
				response.BadRequest(c, "请求体格式错误")
				return
			}
			if e := saveOverrides(c.Request.Context(), d, kind, body.Items); e != nil {
				response.BadRequest(c, e.Error())
				return
			}
			if onChange != nil {
				if dbError(c, onChange()) {
					return
				}
			}
			response.OK(c, gin.H{"count": len(body.Items)})
		})
		for _, action := range []string{"override", "toggle"} {
			action := action
			g.POST(base+"/:id/"+action, func(c *gin.Context) {
				id, ok := parseIDParam(c, "id")
				if !ok {
					return
				}
				var body map[string]any
				if c.ShouldBindJSON(&body) != nil {
					response.BadRequest(c, "请求体格式错误")
					return
				}
				body["id"] = float64(id)
				if e := saveOverrides(c.Request.Context(), d, kind, []map[string]any{body}); e != nil {
					response.BadRequest(c, e.Error())
					return
				}
				if onChange != nil {
					if dbError(c, onChange()) {
						return
					}
				}
				response.OK(c, nil)
			})
		}
		g.POST(base+"/randomize-all", func(c *gin.Context) {
			ids := []int64{}
			if dbError(c, d.SelectContext(c.Request.Context(), &ids, repo.SQL("SELECT id FROM {"+table+"} ORDER BY id"))) {
				return
			}
			items := []map[string]any{}
			for i, id := range ids {
				b := make([]byte, 12)
				if _, e := rand.Read(b); dbError(c, e) {
					return
				}
				v := map[string]any{"id": float64(id), "rewritten": true}
				if kind == "routes" {
					v["overridePath"] = "/api/sdk/" + hex.EncodeToString(b)
				} else {
					v["overrideCode"] = float64(10000 + i*100 + int(b[0])%99)
				}
				items = append(items, v)
			}
			if e := saveOverrides(c.Request.Context(), d, kind, items); e != nil {
				response.BadRequest(c, e.Error())
				return
			}
			if onChange != nil {
				if dbError(c, onChange()) {
					return
				}
			}
			response.OK(c, gin.H{"count": len(items)})
		})
	}
}
func saveOverrides(ctx context.Context, d *sqlx.DB, kind string, items []map[string]any) error {
	tx, e := d.BeginTxx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, v := range items {
		id, ok := v["id"].(float64)
		if !ok || id <= 0 {
			return fmt.Errorf("无效 ID")
		}
		if kind == "routes" {
			var old repo.APIRoute
			if e = tx.GetContext(ctx, &old, repo.SQL("SELECT id,api_key,name,override_path,rewritten,sort_order FROM {api_routes} WHERE id=? FOR UPDATE"), int64(id)); e != nil {
				return e
			}
			if x, ok := v["overridePath"].(string); ok {
				x = strings.TrimSpace(x)
				if x != "" && !strings.HasPrefix(x, "/") {
					x = "/api/sdk/" + x
				}
				old.Override = x
			}
			if x, ok := v["rewritten"].(bool); ok {
				old.Rewritten = 0
				if x {
					old.Rewritten = 1
				}
			}
			if old.Rewritten == 1 {
				if !strings.HasPrefix(old.Override, "/api/sdk/") || strings.ContainsAny(old.Override, "?:*#\\ %\t\r\n") || strings.Contains(old.Override, "..") || len(old.Override) > 255 || len(old.Override) <= len("/api/sdk/") {
					return fmt.Errorf("重写路径必须是 /api/sdk/ 下的固定路径")
				}
				var n int
				if e = tx.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {api_routes} WHERE id<>? AND CONCAT('/api/sdk/',api_key)=?"), old.ID, old.Override); e != nil {
					return e
				}
				if n > 0 || strings.HasPrefix(old.Override, "/api/sdk/diff/") || strings.HasPrefix(old.Override, "/api/sdk/extend-file/") || old.Override == "/api/sdk/clone" || old.Override == "/api/sdk/regen-sign" || old.Override == "/api/sdk/uploadAvatar" {
					return fmt.Errorf("重写路径与现有接口冲突")
				}
				if e = tx.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {api_routes} WHERE id<>? AND rewritten=1 AND override_path=?"), old.ID, old.Override); e != nil {
					return e
				}
				if n > 0 {
					return fmt.Errorf("重写路径重复")
				}
			}
			_, e = tx.ExecContext(ctx, repo.SQL("UPDATE {api_routes} SET override_path=?,rewritten=? WHERE id=?"), old.Override, old.Rewritten, old.ID)
		} else {
			var old repo.APICode
			if e = tx.GetContext(ctx, &old, repo.SQL("SELECT id,code,override_code,message,remark,rewritten FROM {api_codes} WHERE id=? FOR UPDATE"), int64(id)); e != nil {
				return e
			}
			if x, ok := v["overrideCode"].(float64); ok {
				old.OverrideCode = int(x)
			}
			if x, ok := v["rewritten"].(bool); ok {
				old.Rewritten = 0
				if x {
					old.Rewritten = 1
				}
			}
			if old.Rewritten == 1 && old.OverrideCode < 0 {
				return fmt.Errorf("重写码必须为正整数")
			}
			if old.Rewritten == 1 {
				var n int
				if e = tx.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {api_codes} WHERE id<>? AND (code=? OR (rewritten=1 AND override_code=?))"), old.ID, old.OverrideCode, old.OverrideCode); e != nil {
					return e
				}
				if n > 0 {
					return fmt.Errorf("状态码与现有接口冲突")
				}
			}
			_, e = tx.ExecContext(ctx, repo.SQL("UPDATE {api_codes} SET override_code=?,rewritten=? WHERE id=?"), old.OverrideCode, old.Rewritten, old.ID)
		}
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}
func ParsePositive(s string, def int) int {
	v, e := strconv.Atoi(s)
	if e != nil || v < 1 {
		return def
	}
	return v
}
