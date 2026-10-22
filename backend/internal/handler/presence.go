package handler

import (
	"auth-api/internal/middleware"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

// Session activity expires after five minutes or the token TTL, whichever is shorter.
// A single atomic operation keeps the user flag and both dashboard indexes consistent.
var sessionPresence = redis.NewScript(`
local now=tonumber(ARGV[1])
redis.call('ZREMRANGEBYSCORE',KEYS[1],'-inf',now)
if ARGV[2]=='touch' then
 redis.call('ZADD',KEYS[1],ARGV[4],ARGV[3])
elseif ARGV[2]=='logout' then
 redis.call('ZREM',KEYS[1],ARGV[3])
elseif ARGV[2]=='device' then
 for _, m in ipairs(redis.call('ZRANGE',KEYS[1],0,-1)) do
  if string.sub(m,1,string.len(ARGV[3]))==ARGV[3] then redis.call('ZREM',KEYS[1],m) end
 end
end
local last=redis.call('ZREVRANGE',KEYS[1],0,0,'WITHSCORES')
if #last==0 then
 redis.call('DEL',KEYS[1],KEYS[2])
 redis.call('ZREM',KEYS[3],ARGV[5])
 redis.call('ZREM',KEYS[4],ARGV[5])
else
 local ttl=math.max(1,math.ceil((tonumber(last[2])-now)*1000))
 redis.call('PEXPIRE',KEYS[1],ttl)
 redis.call('SET',KEYS[2],1,'PX',ttl)
 for i=3,4 do
  redis.call('ZREMRANGEBYSCORE',KEYS[i],'-inf',now)
  redis.call('ZADD',KEYS[i],last[2],ARGV[5])
  redis.call('EXPIRE',KEYS[i],600)
  if ARGV[2]=='touch' then
   local count=redis.call('ZCARD',KEYS[i])
   if count>tonumber(redis.call('GET',KEYS[i+2]) or '0') then redis.call('SET',KEYS[i+2],count,'EX',3024000) end
  end
 end
end
return 1`)

func changePresence(c *gin.Context, r *redis.Client, app, uid int64, action, member string, ttl int) error {
	if r == nil {
		return nil
	}
	now := time.Now()
	duration := 300
	if ttl > 0 && ttl < duration {
		duration = ttl
	}
	account := fmt.Sprintf("%d:%d", app, uid)
	day := now.Format("2006-01-02")
	return sessionPresence.Run(c.Request.Context(), r, []string{"presence:sessions:" + account, "online:" + account, fmt.Sprintf("presence:%d", app), "presence:0", fmt.Sprintf("online:peak:%d:%s", app, day), "online:peak:0:" + day}, float64(now.UnixMilli())/1000, action, member, float64(now.Add(time.Duration(duration)*time.Second).UnixMilli())/1000, account).Err()
}

func (s *SDKService) sessionActivity(c *gin.Context, uid int64, action string) error {
	token := c.GetString("sdkToken")
	device, ttl := c.GetInt64("presenceDevice"), c.GetInt("presenceTTL")
	if session := middleware.CurrentSDKUser(c); session != nil {
		device, ttl = session.DeviceID, session.TTLSec
	}
	if issued := c.GetString("presenceToken"); issued != "" {
		token = issued
	}
	if c.GetBool("sdkIsKami") {
		uid = -uid
	}
	return changePresence(c, s.rdb, middleware.CurrentApp(c).ID, uid, action, fmt.Sprintf("%d:%x", device, sha256.Sum256([]byte(token))), ttl)
}
