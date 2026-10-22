package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

func userIndex(app, uid int64, kami bool) string {
	if kami {
		uid = -uid
	}
	return fmt.Sprintf("usess:index:%d:%d", app, uid)
}
func sessionID(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (s *UserSessionStore) createIndexed(ctx context.Context, token string, v *UserSession, b []byte, ttl time.Duration) error {
	return redis.NewScript(`
 local v=cjson.decode(ARGV[1])
 v.revision=redis.call('GET',KEYS[3]) or ''
 redis.call('SET',KEYS[1],cjson.encode(v),'PX',ARGV[2])
 redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',ARGV[4])
 redis.call('ZADD',KEYS[2],ARGV[5],ARGV[3])
 for _,t in ipairs(redis.call('ZRANGE',KEYS[2],0,-1001)) do redis.call('DEL','usess:'..t);redis.call('ZREM',KEYS[2],t) end
 if redis.call('PTTL',KEYS[2]) < tonumber(ARGV[2]) then redis.call('PEXPIRE',KEYS[2],ARGV[2]) end
 return 1`).Run(ctx, s.rdb, []string{"usess:" + token, userIndex(v.AppID, v.UID, v.IsKami), userIndex(v.AppID, v.UID, v.IsKami) + ":revision"}, b, ttl.Milliseconds(), token, time.Now().UnixMilli(), time.Now().Add(ttl).UnixMilli()).Err()
}
func (s *UserSessionStore) indexSession(ctx context.Context, token string, v *UserSession, ttl time.Duration) error {
	n, e := redis.NewScript(`
 if not redis.call('GET',KEYS[1]) then return 0 end
 if (redis.call('GET',KEYS[3]) or '') ~= ARGV[5] then redis.call('DEL',KEYS[1]); return 0 end
 redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',ARGV[1])
 redis.call('ZADD',KEYS[2],ARGV[2],ARGV[3])
 if redis.call('PTTL',KEYS[2]) < tonumber(ARGV[4]) then redis.call('PEXPIRE',KEYS[2],ARGV[4]) end
 return 1`).Run(ctx, s.rdb, []string{"usess:" + token, userIndex(v.AppID, v.UID, v.IsKami), userIndex(v.AppID, v.UID, v.IsKami) + ":revision"}, time.Now().UnixMilli(), time.Now().Add(ttl).UnixMilli(), token, ttl.Milliseconds(), v.Revision).Int()
	if e == nil && n == 0 {
		return ErrSession
	}
	return e
}

type SessionInfo struct {
	ID              string `json:"id"`
	Mi              string `json:"mi"`
	LoginIP         string `json:"loginIp"`
	LastIP          string `json:"lastIp"`
	IssuedAt        int64  `json:"issuedAt"`
	LastActiveAt    int64  `json:"lastActiveAt"`
	LastHeartbeatAt int64  `json:"lastHeartbeatAt"`
	ExpiresAt       int64  `json:"expiresAt"`
}

func (s *UserSessionStore) List(ctx context.Context, app, uid int64, kami bool) ([]SessionInfo, error) {
	key := userIndex(app, uid, kami)
	if e := s.rdb.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprint(time.Now().UnixMilli())).Err(); e != nil {
		return nil, e
	}
	tokens, e := s.rdb.ZRevRange(ctx, key, 0, 999).Result()
	if e != nil {
		return nil, e
	}
	out := []SessionInfo{}
	for _, token := range tokens {
		b, err := s.rdb.Get(ctx, "usess:"+token).Bytes()
		if err == redis.Nil {
			s.rdb.ZRem(ctx, key, token)
			continue
		}
		if err != nil {
			return nil, err
		}
		var v UserSession
		if json.Unmarshal(b, &v) != nil || v.AppID != app || v.UID != uid || v.IsKami != kami {
			continue
		}
		ttl, err := s.rdb.TTL(ctx, "usess:"+token).Result()
		if err != nil {
			return nil, err
		}
		if ttl <= 0 {
			continue
		}
		out = append(out, SessionInfo{sessionID(token), v.Mi, v.LoginIP, v.LastIP, v.IssuedAt, v.LastActiveAt, v.LastHeartbeatAt, time.Now().Add(ttl).Unix()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActiveAt > out[j].LastActiveAt })
	return out, nil
}
func (s *UserSessionStore) Revoke(ctx context.Context, app, uid int64, kami bool, id string) (*UserSession, error) {
	key := userIndex(app, uid, kami)
	tokens, e := s.rdb.ZRange(ctx, key, 0, -1).Result()
	if e != nil {
		return nil, e
	}
	for _, token := range tokens {
		if sessionID(token) != id {
			continue
		}
		b, e := redis.NewScript(`
   local raw=redis.call('GET',KEYS[1]);if not raw then return '' end
   local v=cjson.decode(raw)
   if v.appid~=tonumber(ARGV[1]) or v.uid~=tonumber(ARGV[2]) or (v.isKami or false)~=(ARGV[3]=='true') then return '' end
   redis.call('DEL',KEYS[1]);redis.call('ZREM',KEYS[2],ARGV[4]);return raw
  `).Run(ctx, s.rdb, []string{"usess:" + token, key}, app, uid, fmt.Sprint(kami), token).Text()
		if e != nil {
			return nil, e
		}
		if b == "" {
			return nil, ErrSession
		}
		var v UserSession
		if e = json.Unmarshal([]byte(b), &v); e != nil {
			return nil, e
		}
		return &v, nil
	}
	return nil, ErrSession
}
func (s *UserSessionStore) RevokeAll(ctx context.Context, app, uid int64, kami bool) error {
	revision, e := RandomToken()
	if e != nil {
		return e
	}
	key := userIndex(app, uid, kami)
	return redis.NewScript(`
 redis.call('SET',KEYS[2],ARGV[1])
 for _,t in ipairs(redis.call('ZRANGE',KEYS[1],0,-1)) do redis.call('DEL','usess:'..t) end
 redis.call('DEL',KEYS[1]);return 1`).Run(ctx, s.rdb, []string{key, key + ":revision"}, revision).Err()
}

// Revoke only sessions bound to this concrete device row; other devices stay logged in.
func (s *UserSessionStore) RevokeDevice(ctx context.Context, app, uid int64, kami bool, deviceID int64) error {
	return redis.NewScript(`
 for _,t in ipairs(redis.call('ZRANGE',KEYS[1],0,-1)) do
  local key='usess:'..t
  local raw=redis.call('GET',key)
  if not raw then redis.call('ZREM',KEYS[1],t)
  else
   local v=cjson.decode(raw)
   if v.appid==tonumber(ARGV[1]) and v.uid==tonumber(ARGV[2]) and (v.isKami or false)==(ARGV[3]=='true') and v.deviceId==tonumber(ARGV[4]) then
    redis.call('DEL',key);redis.call('ZREM',KEYS[1],t)
   end
  end
 end
 return 1`).Run(ctx, s.rdb, []string{userIndex(app, uid, kami)}, app, uid, fmt.Sprint(kami), deviceID).Err()
}
func (s *UserSessionStore) Touch(ctx context.Context, token, ip string, heartbeat bool) error {
	n, e := redis.NewScript(`
 local raw=redis.call('GET',KEYS[1]);if not raw then return 0 end
 local v=cjson.decode(raw);v.lastActiveAt=tonumber(ARGV[1]);v.lastIp=ARGV[2]
 if ARGV[3]=='true' then v.lastHeartbeatAt=tonumber(ARGV[1]) end
 redis.call('SET',KEYS[1],cjson.encode(v),'KEEPTTL','XX');return 1
 `).Run(ctx, s.rdb, []string{"usess:" + token}, time.Now().Unix(), ip, fmt.Sprint(heartbeat)).Int()
	if e == nil && n == 0 {
		return ErrSession
	}
	return e
}
