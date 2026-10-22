// Reconstructed from token.go, usertoken.go and sdksessionstore.go assembly.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

var ErrSession = errors.New("session expired or invalid")

// Bind a token to the password revision without copying the password hash into Redis.
func CredentialRevision(hash string) string {
	v := sha256.Sum256([]byte(hash))
	return hex.EncodeToString(v[:])
}

func HashPassword(password string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), e
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
func RandomToken() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return hex.EncodeToString(b), nil
}
func NewSessionStore(r *redis.Client, ttl, absolute time.Duration) *SessionStore {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	return &SessionStore{rdb: r, ttl: ttl, absoluteTTL: absolute}
}
func (s *SessionStore) currentTTL() time.Duration {
	if s.TTLProvider != nil {
		if v := s.TTLProvider(); v > 0 {
			return v
		}
	}
	return s.ttl
}
func adminIndex(id int64) string { return "sess:admin:" + strconv.FormatInt(id, 10) }
func (s *SessionStore) Create(ctx context.Context, v *Session) (string, error) {
	token, e := RandomToken()
	if e != nil {
		return "", e
	}
	v.IssuedAt = time.Now().Unix()
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	// Match the original single-session string index. Also revoke all tokens
	// from the set index used by the previous reconstructed release.
	ttl := s.currentTTL()
	if s.absoluteTTL > 0 && s.absoluteTTL < ttl {
		ttl = s.absoluteTTL
	}
	_, e = redis.NewScript(`
local kind = redis.call('TYPE', KEYS[2]).ok
local previous = {}
if kind == 'string' then
  previous = {redis.call('GET', KEYS[2])}
elseif kind == 'set' then
  previous = redis.call('SMEMBERS', KEYS[2])
end
if kind ~= 'none' and kind ~= 'string' and kind ~= 'set' then
  return redis.error_reply('WRONGTYPE invalid admin session index')
end
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
redis.call('SET', KEYS[2], ARGV[3], 'PX', ARGV[2])
for _, token in ipairs(previous) do
  if token ~= ARGV[3] then redis.call('DEL', 'sess:' .. token) end
end
return 1`).Run(ctx, s.rdb, []string{"sess:" + token, adminIndex(v.AdminID)}, b, ttl.Milliseconds(), token).Result()
	return token, e
}
func (s *SessionStore) Get(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, ErrSession
	}
	b, e := s.rdb.Get(ctx, "sess:"+token).Bytes()
	if errors.Is(e, redis.Nil) {
		return nil, ErrSession
	}
	if e != nil {
		return nil, e
	}
	var v Session
	if e = json.Unmarshal(b, &v); e != nil {
		return nil, e
	}
	ttl := s.currentTTL()
	if s.absoluteTTL > 0 {
		left := time.Until(time.Unix(v.IssuedAt, 0).Add(s.absoluteTTL))
		if left <= 0 {
			_ = s.Delete(ctx, token)
			return nil, ErrSession
		}
		if left < ttl {
			ttl = left
		}
	}
	// Validate and refresh both keys atomically. A request that read its token
	// just before a new login must not revive or accept the revoked session.
	alive, e := redis.NewScript(`
if redis.call('TYPE', KEYS[2]).ok ~= 'string' then return 0 end
if redis.call('GET', KEYS[2]) ~= ARGV[1] then return 0 end
if redis.call('PEXPIRE', KEYS[1], ARGV[2]) == 0 then return 0 end
redis.call('PEXPIRE', KEYS[2], ARGV[2])
return 1`).Run(ctx, s.rdb, []string{"sess:" + token, adminIndex(v.AdminID)}, token, ttl.Milliseconds()).Int64()
	if e != nil {
		return nil, e
	}
	if alive != 1 {
		return nil, ErrSession
	}
	return &v, nil
}
func (s *SessionStore) Update(ctx context.Context, token string, v *Session) error {
	current, e := s.Get(ctx, token)
	if e != nil {
		return e
	}
	if current.AdminID != v.AdminID {
		return ErrSession
	}
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	updated, e := redis.NewScript(`
if redis.call('TYPE', KEYS[2]).ok ~= 'string' then return 0 end
if redis.call('GET', KEYS[2]) ~= ARGV[1] then return 0 end
if not redis.call('SET', KEYS[1], ARGV[2], 'XX', 'KEEPTTL') then return 0 end
return 1`).Run(ctx, s.rdb, []string{"sess:" + token, adminIndex(v.AdminID)}, token, b).Int64()
	if e != nil {
		return e
	}
	if updated != 1 {
		return ErrSession
	}
	return nil
}
func (s *SessionStore) Delete(ctx context.Context, token string) error {
	b, e := s.rdb.Get(ctx, "sess:"+token).Bytes()
	if e != nil && !errors.Is(e, redis.Nil) {
		return e
	}
	var v Session
	_ = json.Unmarshal(b, &v)
	if v.AdminID <= 0 {
		return s.rdb.Del(ctx, "sess:"+token).Err()
	}
	return redis.NewScript(`
local kind = redis.call('TYPE', KEYS[2]).ok
if kind ~= 'none' and kind ~= 'string' and kind ~= 'set' then
  return redis.error_reply('WRONGTYPE invalid admin session index')
end
redis.call('DEL', KEYS[1])
if kind == 'string' then
  if redis.call('GET', KEYS[2]) == ARGV[1] then redis.call('DEL', KEYS[2]) end
elseif kind == 'set' then
  redis.call('SREM', KEYS[2], ARGV[1])
end
return 1`).Run(ctx, s.rdb, []string{"sess:" + token, adminIndex(v.AdminID)}, token).Err()
}
func (s *SessionStore) DeleteByAdmin(ctx context.Context, id int64) error {
	return redis.NewScript(`
local kind = redis.call('TYPE', KEYS[1]).ok
local tokens = {}
if kind == 'string' then
  tokens = {redis.call('GET', KEYS[1])}
elseif kind == 'set' then
  tokens = redis.call('SMEMBERS', KEYS[1])
elseif kind ~= 'none' then
  return redis.error_reply('WRONGTYPE invalid admin session index')
end
for _, token in ipairs(tokens) do redis.call('DEL', 'sess:' .. token) end
return redis.call('DEL', KEYS[1])`).Run(ctx, s.rdb, []string{adminIndex(id)}).Err()
}
func NewUserSessionStore(r *redis.Client, ttl time.Duration) *UserSessionStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &UserSessionStore{r, ttl}
}
func (s *UserSessionStore) CreateWithTTL(ctx context.Context, v *UserSession, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = s.ttl
	}
	token, e := RandomToken()
	if e != nil {
		return "", e
	}
	v.IssuedAt = time.Now().Unix()
	v.TTLSec = int(ttl / time.Second)
	v.LastActiveAt = v.IssuedAt
	v.LoginIP = v.LastIP
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	return token, s.createIndexed(ctx, token, v, b, ttl)
}
func (s *UserSessionStore) Get(ctx context.Context, token string) (*UserSession, error) {
	if token == "" {
		return nil, ErrSession
	}
	b, e := s.rdb.Get(ctx, "usess:"+token).Bytes()
	if errors.Is(e, redis.Nil) {
		return nil, ErrSession
	}
	if e != nil {
		return nil, e
	}
	var v UserSession
	if e = json.Unmarshal(b, &v); e != nil {
		return nil, e
	}
	ttl := s.ttl
	if v.TTLSec > 0 {
		ttl = time.Duration(v.TTLSec) * time.Second
	}
	alive, e := s.rdb.Expire(ctx, "usess:"+token, ttl).Result()
	if e != nil {
		return nil, e
	}
	if !alive {
		return nil, ErrSession
	}
	if e = s.indexSession(ctx, token, &v, ttl); e != nil {
		return nil, e
	}
	return &v, nil
}
func (s *UserSessionStore) Delete(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, "usess:"+token).Err()
}
func NewSDKSessionStore(r *redis.Client, ttl time.Duration) *SDKSessionStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &SDKSessionStore{r, ttl}
}
func (s *SDKSessionStore) Set(ctx context.Context, id string, v *SDKSession) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return s.rdb.Set(ctx, "sdksess:"+id, b, s.ttl).Err()
}
func (s *SDKSessionStore) Get(ctx context.Context, id string) (*SDKSession, error) {
	b, e := s.rdb.Get(ctx, "sdksess:"+id).Bytes()
	if errors.Is(e, redis.Nil) {
		return nil, ErrSession
	}
	if e != nil {
		return nil, e
	}
	var v SDKSession
	e = json.Unmarshal(b, &v)
	return &v, e
}
func (s *SDKSessionStore) RefreshTTL(ctx context.Context, id string) error {
	return s.rdb.Expire(ctx, "sdksess:"+id, s.ttl).Err()
}
