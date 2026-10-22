package handler

import (
	"auth-api/internal/middleware"
	"auth-api/internal/response"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// Cooldowns do not sleep in request goroutines. Reserve slots atomically so
// simultaneous guesses cannot all pass before the first failure is recorded.
func (s *SDKService) cardAttempt(c *gin.Context, card, identity string) bool {
	if len(card) > 256 {
		response.BadRequest(c, "卡密格式无效")
		return false
	}
	c.Set("security.subject", identity)
	if !middleware.Limit(c, s.rdb, "card-ip", c.ClientIP(), 20, time.Minute) {
		return false
	}
	if identity != "" && !middleware.Limit(c, s.rdb, "card-account", identity, 15, time.Minute) {
		return false
	}
	if !middleware.Limit(c, s.rdb, "card-value", strings.ToLower(strings.TrimSpace(card)), 20, time.Minute) {
		return false
	}
	app := middleware.CurrentApp(c).ID
	key := fmt.Sprintf("card:cooldown:%d:%x", app, sha256.Sum256([]byte(c.ClientIP())))
	allowed, e := redis.NewScript(`
 local t=redis.call('TIME');local now=tonumber(t[1])*1000+math.floor(tonumber(t[2])/1000)
 local next=tonumber(redis.call('HGET',KEYS[1],'next') or '0')
 if next>now then return 0 end
 local n=redis.call('HINCRBY',KEYS[1],'n',1)
 if n>5 then redis.call('HSET',KEYS[1],'next',now+math.min(30000,1000*2^math.min(n-6,5))) end
 redis.call('EXPIRE',KEYS[1],300);return 1
 `).Run(c.Request.Context(), s.rdb, []string{key}).Int()
	if e != nil {
		response.ServerError(c, "卡密验证服务不可用")
		return false
	}
	if allowed == 0 {
		c.Header("Retry-After", "30")
		response.Fail(c, 1017, "请求过于频繁，请稍后重试")
		return false
	}
	c.Set("cardGuardKey", key)
	return true
}
func (s *SDKService) cardSuccess(c *gin.Context) {
	if key := c.GetString("cardGuardKey"); key != "" {
		if e := s.rdb.Del(c.Request.Context(), key).Err(); e != nil {
			c.Error(e)
		}
	}
}
