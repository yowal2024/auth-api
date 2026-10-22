package handler

import (
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Reserve an attempt before password verification. A separate GET followed by
// an increment after verification lets concurrent guesses all pass the limit.
var reserveLoginAttempt = redis.NewScript(`
local n = tonumber(redis.call('GET', KEYS[1]) or '0')
if n >= 5 then return 0 end
n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('EXPIRE', KEYS[1], 300) end
return 1`)

func (s *SDKService) allowLoginAttempt(c *gin.Context, key string) bool {
	n, err := reserveLoginAttempt.Run(c.Request.Context(), s.rdb, []string{key}).Int()
	if err != nil {
		response.ServerError(c, "防爆破服务不可用")
		return false
	}
	if n == 0 {
		response.Fail(c, 1017, "请求过于频繁，请稍后重试")
		return false
	}
	return true
}
