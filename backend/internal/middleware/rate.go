package middleware

import (
	"auth-api/internal/response"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

var rateScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('EXPIRE', KEYS[1], 60) end
return n`)

// Applies even when application-level brute-force checking is disabled.
// Gin's trusted proxy list is empty, so spoofed forwarding headers cannot reset it.
func SDKRateLimit(r *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := int64(300)
		switch c.FullPath() {
		case "/api/sdk/clone", "/api/sdk/regen-sign":
			limit = 6
		case "/api/sdk/handshake", "/api/sdk/uploadAvatar":
			limit = 30
		case "/api/sdk/kamiTopup":
			limit = 20
		case "/api/sdk/register":
			limit = 10
		case "/api/sdk/logon":
			limit = 60
		}
		switch c.FullPath() {
		case "/api/sdk/logon", "/api/sdk/register", "/api/sdk/kamiTopup":
			if !Limit(c, r, "credential-app", "all", 600, time.Minute) {
				c.Abort()
				return
			}
		}
		digest := sha256.Sum256([]byte(c.ClientIP() + "\n" + c.FullPath()))
		key := fmt.Sprintf("sdk:rate:%d:%x", CurrentApp(c).ID, digest)
		n, err := rateScript.Run(c.Request.Context(), r, []string{key}).Int64()
		if err != nil {
			response.ServerError(c, "限流服务不可用")
			c.Abort()
			return
		}
		if n > limit {
			c.Header("Retry-After", "60")
			response.Fail(c, 1017, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}
		c.Next()
	}
}
