package middleware

import (
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"time"
)

// Reserve before expensive work; scripts make parallel requests share the limit.
func Limit(c *gin.Context, r *redis.Client, scope, identity string, limit int64, window time.Duration) bool {
	if r == nil {
		response.ServerError(c, "限流服务不可用")
		return false
	}
	app := int64(0)
	if a := CurrentApp(c); a != nil {
		app = a.ID
	}
	key := fmt.Sprintf("guard:%d:%s:%x", app, scope, sha256.Sum256([]byte(identity)))
	n, e := redis.NewScript(`local n=redis.call('INCR',KEYS[1]);if n==1 then redis.call('EXPIRE',KEYS[1],ARGV[1]) end;return n`).Run(c.Request.Context(), r, []string{key}, int64(window/time.Second)).Int64()
	if e != nil {
		response.ServerError(c, "限流服务不可用")
		return false
	}
	if n > limit {
		c.Set("security.kind", "rate_limit")
		c.Header("Retry-After", fmt.Sprint(int64(window/time.Second)))
		response.Fail(c, 1017, "请求过于频繁，请稍后重试")
		return false
	}
	return true
}

func SecurityEvents(d *sqlx.DB, r *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		code := c.GetInt(response.CodeContextKey)
		kind := c.GetString("security.kind")
		summary := "请求触发了频率限制"
		if kind == "" {
			switch code {
			case 1001:
				kind = "login_failed"
				summary = "登录验证失败"
			case 2001, 2002, 2003:
				kind = "card_rejected"
				summary = "卡密验证未通过"
			case 1017:
				kind = "rate_limit"
			case 1014, 1015, 1016:
				kind = "signature_rejected"
				summary = "请求签名或加密校验失败"
			case 1003, 1021:
				kind = "blocked"
				summary = "封禁规则拦截了请求"
			default:
				return
			}
		}
		app := int64(0)
		if a := CurrentApp(c); a != nil {
			app = a.ID
		}
		subject := c.GetString("security.subject")
		// Never persist raw passwords, card numbers, tokens, request bodies or device identifiers.
		if subject != "" {
			subject = fmt.Sprintf("%x", sha256.Sum256([]byte(subject)))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		dedupe := fmt.Sprintf("security:event:%x", sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s:%s", app, kind, c.ClientIP(), c.FullPath()))))
		ok, e := r.SetNX(ctx, dedupe, 1, time.Minute).Result()
		if e != nil || !ok {
			return
		}
		// A rotating-IP flood must not turn the event sink into an unbounded
		// database write stream. Count before insert, with an expiring global cap.
		budgetKey := fmt.Sprintf("security:event-budget:%d", time.Now().Unix()/60)
		count, err := rateScript.Run(ctx, r, []string{budgetKey}).Int64()
		if err != nil || count > 600 {
			return
		}
		_, e = d.ExecContext(ctx, repo.SQL("INSERT INTO {security_events} (appid,kind,summary,ip,subject,route,create_time) VALUES (?,?,?,?,?,?,?)"), app, kind, summary, c.ClientIP(), subject, c.FullPath(), time.Now().Unix())
		if e != nil {
			c.Error(e)
		}
	}
}

func PublicAdminLimit(r *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !Limit(c, r, "admin-public:"+c.FullPath(), c.ClientIP(), 30, time.Minute) {
			c.Abort()
			return
		}
		c.Next()
	}
}
