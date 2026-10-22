package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

// Delete after Redis GET has returned but before the caller can refresh TTL.
type revokeAfterRead struct {
	redis.Hook
	db  *miniredis.Miniredis
	key string
}

func (h revokeAfterRead) DialHook(next redis.DialHook) redis.DialHook { return next }
func (h revokeAfterRead) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h revokeAfterRead) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if cmd.Name() == "get" && len(cmd.Args()) > 1 && cmd.Args()[1] == h.key {
			h.db.Del(h.key)
		}
		return err
	}
}

func TestSessionRevokedBetweenReadAndRefresh(t *testing.T) {
	for _, kind := range []string{"admin", "user"} {
		t.Run(kind, func(t *testing.T) {
			m, r := redisTest(t)
			ctx := context.Background()
			var token, prefix string
			var err error
			var get func() error
			if kind == "admin" {
				s := NewSessionStore(r, time.Hour, 0)
				token, err = s.Create(ctx, &Session{AdminID: 1})
				prefix = "sess:"
				get = func() error { _, err := s.Get(ctx, token); return err }
			} else {
				s := NewUserSessionStore(r, time.Hour)
				token, err = s.CreateWithTTL(ctx, &UserSession{UID: 1}, time.Hour)
				prefix = "usess:"
				get = func() error { _, err := s.Get(ctx, token); return err }
			}
			if err != nil {
				t.Fatal(err)
			}
			r.AddHook(revokeAfterRead{db: m, key: prefix + token})
			if err := get(); !errors.Is(err, ErrSession) {
				t.Fatalf("revoked token accepted: %v", err)
			}
		})
	}
}

func redisTest(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	t.Cleanup(func() { r.Close() })
	return m, r
}
func TestSessionLifecycle(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, 2*time.Hour)
	v := &Session{AdminID: 7, Username: "admin", Role: "super"}
	token, e := s.Create(ctx, v)
	if e != nil || len(token) != 64 {
		t.Fatalf("create: %v", e)
	}
	m.FastForward(50 * time.Minute)
	got, e := s.Get(ctx, token)
	if e != nil || got.AdminID != 7 {
		t.Fatalf("get: %v", e)
	}
	if m.TTL("sess:"+token) < 59*time.Minute {
		t.Fatal("sliding expiration not refreshed")
	}
	if m.TTL(adminIndex(7)) != m.TTL("sess:"+token) {
		t.Fatal("session and index expiration diverged")
	}
	got.Nickname = "updated"
	if e = s.Update(ctx, token, got); e != nil {
		t.Fatal(e)
	}
	got, e = s.Get(ctx, token)
	if e != nil || got.Nickname != "updated" {
		t.Fatal("profile was not persisted")
	}
	if e = s.DeleteByAdmin(ctx, 7); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, token); e == nil {
		t.Fatal("revoked token survived")
	}
}

func TestShorterTTLDoesNotLoseRevocationIndex(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, 0)
	old, err := s.Create(ctx, &Session{AdminID: 9})
	if err != nil {
		t.Fatal(err)
	}
	s.TTLProvider = func() time.Duration { return time.Minute }
	current, err := s.Create(ctx, &Session{AdminID: 9})
	if err != nil {
		t.Fatal(err)
	}
	if m.Exists("sess:" + old) {
		t.Fatal("new login retained the old long-lived session")
	}
	if m.TTL(adminIndex(9)) != time.Minute || m.TTL("sess:"+current) != time.Minute {
		t.Fatal("new session and index did not use the changed TTL")
	}
	m.FastForward(2 * time.Minute)
	if err := s.DeleteByAdmin(ctx, 9); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, old); err == nil {
		t.Fatal("older token survived revocation")
	}
}

func TestAdminSessionCreateReplacesPriorIndexes(t *testing.T) {
	for _, kind := range []string{"string", "set"} {
		for _, persistent := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/persistent=%t", kind, persistent), func(t *testing.T) {
				m, r := redisTest(t)
				ctx := context.Background()
				s := NewSessionStore(r, time.Hour, 0)
				legacyToken := "legacy-admin-token"
				legacyData := `{"AdminID":7,"Username":"admin"}`
				if err := r.Set(ctx, "sess:"+legacyToken, legacyData, 2*time.Hour).Err(); err != nil {
					t.Fatal(err)
				}
				indexTTL := 2 * time.Hour
				if persistent {
					indexTTL = 0
				}
				if err := r.Set(ctx, adminIndex(7), legacyToken, indexTTL).Err(); err != nil {
					t.Fatal(err)
				}
				previous := []string{legacyToken}
				if kind == "set" {
					r.Del(ctx, adminIndex(7))
					r.SAdd(ctx, adminIndex(7), legacyToken, "second-old-token")
					r.Set(ctx, "sess:second-old-token", legacyData, 2*time.Hour)
					previous = append(previous, "second-old-token")
					if !persistent {
						r.Expire(ctx, adminIndex(7), indexTTL)
					}
				}

				token, err := s.Create(ctx, &Session{AdminID: 7, Username: "admin"})
				if err != nil {
					t.Fatalf("login with an existing %s index failed: %v", kind, err)
				}
				if got := r.Type(ctx, adminIndex(7)).Val(); got != "string" {
					t.Fatalf("index type = %s, want string", got)
				}
				if r.Get(ctx, adminIndex(7)).Val() != token {
					t.Fatal("index does not point to the new token")
				}
				for _, old := range previous {
					if m.Exists("sess:" + old) {
						t.Fatal("new login retained a previous session")
					}
				}
				if _, err := s.Get(ctx, token); err != nil {
					t.Fatalf("new token rejected: %v", err)
				}
				if got := m.TTL(adminIndex(7)); got != time.Hour {
					t.Fatalf("new index TTL = %v, want 1h", got)
				}
				if err := s.DeleteByAdmin(ctx, 7); err != nil {
					t.Fatal(err)
				}
				if remaining := r.Exists(ctx, "sess:"+legacyToken, "sess:"+token, adminIndex(7)).Val(); remaining != 0 {
					t.Fatal("revocation left a legacy or new session behind")
				}
			})
		}
	}
}

func TestAdminSingleSessionAndAccountIsolation(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, 0)
	first, err := s.Create(ctx, &Session{AdminID: 7})
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.Create(ctx, &Session{AdminID: 8})
	if err != nil {
		t.Fatal(err)
	}
	latest, err := s.Create(ctx, &Session{AdminID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, first); !errors.Is(err, ErrSession) {
		t.Fatalf("old login accepted: %v", err)
	}
	if err := s.Update(ctx, first, &Session{AdminID: 7}); !errors.Is(err, ErrSession) {
		t.Fatalf("old login updated: %v", err)
	}
	if err := s.Delete(ctx, first); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{latest, other} {
		if _, err := s.Get(ctx, token); err != nil {
			t.Fatalf("active login rejected: %v", err)
		}
	}
	if err := s.DeleteByAdmin(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, other); err != nil {
		t.Fatalf("other account revoked: %v", err)
	}
	if m.Exists("sess:"+first) || m.Exists("sess:"+latest) {
		t.Fatal("revoked tokens survived")
	}
}

func TestAdminConcurrentLoginsKeepOneSession(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, 0)
	type result struct {
		token string
		err   error
	}
	results := make(chan result, 32)
	start := make(chan struct{})
	for i := 0; i < cap(results); i++ {
		go func() {
			<-start
			token, err := s.Create(ctx, &Session{AdminID: 7})
			results <- result{token, err}
		}()
	}
	close(start)
	tokens := make([]string, cap(results))
	for i := range tokens {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		tokens[i] = result.token
	}
	current := r.Get(ctx, adminIndex(7)).Val()
	valid := 0
	for _, token := range tokens {
		_, err := s.Get(ctx, token)
		if token == current {
			if err != nil {
				t.Fatalf("current token rejected: %v", err)
			}
			valid++
		} else if !errors.Is(err, ErrSession) || m.Exists("sess:"+token) {
			t.Fatalf("concurrent login retained an old token: %v", err)
		}
	}
	if valid != 1 || len(m.Keys()) != 2 {
		t.Fatal("expected exactly one session and its index")
	}
}

func TestAdminSessionRejectsNonCurrentIndex(t *testing.T) {
	for _, kind := range []string{"missing", "different-token", "set"} {
		t.Run(kind, func(t *testing.T) {
			_, r := redisTest(t)
			ctx := context.Background()
			s := NewSessionStore(r, time.Hour, 0)
			token, err := s.Create(ctx, &Session{AdminID: 7})
			if err != nil {
				t.Fatal(err)
			}
			r.Del(ctx, adminIndex(7))
			switch kind {
			case "different-token":
				r.Set(ctx, adminIndex(7), "new-token", time.Hour)
			case "set":
				r.SAdd(ctx, adminIndex(7), token, "another-token")
			}
			if _, err := s.Get(ctx, token); !errors.Is(err, ErrSession) {
				t.Fatalf("non-current token accepted: %v", err)
			}
		})
	}
}

type afterSessionCommand struct {
	redis.Hook
	once  sync.Once
	match func(redis.Cmder) bool
	after func()
}

func (h *afterSessionCommand) DialHook(next redis.DialHook) redis.DialHook { return next }
func (h *afterSessionCommand) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h *afterSessionCommand) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if err == nil && h.match(cmd) {
			h.once.Do(h.after)
		}
		return err
	}
}

func TestAdminReloginDuringReadOrUpdate(t *testing.T) {
	for _, operation := range []string{"read", "update"} {
		t.Run(operation, func(t *testing.T) {
			m, r := redisTest(t)
			ctx := context.Background()
			s := NewSessionStore(r, time.Hour, 0)
			old, err := s.Create(ctx, &Session{AdminID: 7})
			if err != nil {
				t.Fatal(err)
			}
			otherClient := redis.NewClient(&redis.Options{Addr: m.Addr()})
			defer otherClient.Close()
			otherStore := NewSessionStore(otherClient, time.Hour, 0)
			var latest string
			var loginErr error
			r.AddHook(&afterSessionCommand{
				match: func(cmd redis.Cmder) bool {
					if operation == "read" {
						return cmd.Name() == "get" && cmd.Args()[1] == "sess:"+old
					}
					return (cmd.Name() == "eval" || cmd.Name() == "evalsha") && len(cmd.Args()) > 3 && cmd.Args()[3] == "sess:"+old
				},
				after: func() { latest, loginErr = otherStore.Create(ctx, &Session{AdminID: 7}) },
			})
			if operation == "read" {
				_, err = s.Get(ctx, old)
			} else {
				err = s.Update(ctx, old, &Session{AdminID: 7, Nickname: "stale"})
			}
			if latest == "" || loginErr != nil {
				t.Fatalf("interleaved login failed: %v", loginErr)
			}
			if !errors.Is(err, ErrSession) {
				t.Fatalf("stale %s accepted: %v", operation, err)
			}
			if m.Exists("sess:" + old) {
				t.Fatal("old session was resurrected")
			}
			if _, err := otherStore.Get(ctx, latest); err != nil {
				t.Fatalf("new session rejected: %v", err)
			}
		})
	}
}

func TestFailedAdminSessionCreationKeepsCurrentLogin(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, 0)
	current, err := s.Create(ctx, &Session{AdminID: 7})
	if err != nil {
		t.Fatal(err)
	}
	s.TTLProvider = func() time.Duration { return time.Nanosecond }
	if _, err := s.Create(ctx, &Session{AdminID: 7}); err == nil {
		t.Fatal("invalid Redis TTL was accepted")
	}
	s.TTLProvider = nil
	if _, err := s.Get(ctx, current); err != nil {
		t.Fatalf("failed login revoked current session: %v", err)
	}
	if len(m.Keys()) != 2 {
		t.Fatal("failed login left an orphan token")
	}
}

func TestAdminSessionLegacyStringIndexRevocation(t *testing.T) {
	for _, mode := range []string{"logout", "logout-other-session", "revoke-admin"} {
		t.Run(mode, func(t *testing.T) {
			_, r := redisTest(t)
			ctx := context.Background()
			s := NewSessionStore(r, time.Hour, 0)
			legacyToken := "legacy-admin-token"
			data, err := json.Marshal(&Session{AdminID: 7, Username: "admin"})
			if err != nil {
				t.Fatal(err)
			}
			if err := r.Set(ctx, "sess:"+legacyToken, data, time.Hour).Err(); err != nil {
				t.Fatal(err)
			}
			indexedToken := legacyToken
			if mode == "logout-other-session" {
				indexedToken = "other-admin-token"
			}
			if err := r.Set(ctx, adminIndex(7), indexedToken, time.Hour).Err(); err != nil {
				t.Fatal(err)
			}
			if mode == "revoke-admin" {
				err = s.DeleteByAdmin(ctx, 7)
			} else {
				err = s.Delete(ctx, legacyToken)
			}
			if err != nil {
				t.Fatalf("legacy %s failed: %v", mode, err)
			}
			if r.Exists(ctx, "sess:"+legacyToken).Val() != 0 {
				t.Fatal("revoked token survived")
			}
			if mode == "logout-other-session" {
				if r.Get(ctx, adminIndex(7)).Val() != indexedToken {
					t.Fatal("logout removed another session's legacy index")
				}
			} else if r.Exists(ctx, adminIndex(7)).Val() != 0 {
				t.Fatal("revoked legacy index survived")
			}
		})
	}
}

func TestAdminSessionCreateRejectsUnexpectedIndexWithoutIssuingToken(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	if err := r.HSet(ctx, adminIndex(7), "unrelated", "data").Err(); err != nil {
		t.Fatal(err)
	}
	s := NewSessionStore(r, time.Hour, 0)
	if _, err := s.Create(ctx, &Session{AdminID: 7}); err == nil {
		t.Fatal("unexpected index type was accepted")
	}
	if keys := m.Keys(); len(keys) != 1 || keys[0] != adminIndex(7) {
		t.Fatal("failed login left an untracked token")
	}
	if r.HGet(ctx, adminIndex(7), "unrelated").Val() != "data" {
		t.Fatal("unexpected index data was overwritten")
	}
}

func TestAbsoluteExpiryAndMissingUpdate(t *testing.T) {
	_, r := redisTest(t)
	ctx := context.Background()
	s := NewSessionStore(r, time.Hour, time.Hour)
	token, e := s.Create(ctx, &Session{AdminID: 1})
	if e != nil {
		t.Fatal(e)
	}
	expired := Session{AdminID: 1, IssuedAt: time.Now().Add(-2 * time.Hour).Unix()}
	b, _ := json.Marshal(expired)
	r.Set(ctx, "sess:"+token, b, time.Hour)
	if _, e = s.Get(ctx, token); e == nil {
		t.Fatal("absolute expiry not enforced")
	}
	if e = s.Update(ctx, "missing", &Session{AdminID: 2}); e == nil {
		t.Fatal("update created a missing session")
	}
}
func TestUserAndSDKIsolationExpiry(t *testing.T) {
	m, r := redisTest(t)
	ctx := context.Background()
	users := NewUserSessionStore(r, time.Hour)
	token, e := users.CreateWithTTL(ctx, &UserSession{UID: 9, AppID: 1000}, 10*time.Minute)
	if e != nil {
		t.Fatal(e)
	}
	m.FastForward(9 * time.Minute)
	v, e := users.Get(ctx, token)
	if e != nil || v.AppID != 1000 || m.TTL("usess:"+token) < 9*time.Minute {
		t.Fatal("per-app session TTL failed")
	}
	sdk := NewSDKSessionStore(r, time.Hour)
	if _, e = sdk.Get(ctx, token); e == nil {
		t.Fatal("user token accepted as handshake token")
	}
	m.FastForward(11 * time.Minute)
	if _, e = users.Get(ctx, token); e == nil {
		t.Fatal("expired user token accepted")
	}
}
func TestCaptchaSingleUseAndThreshold(t *testing.T) {
	_, r := redisTest(t)
	ctx := context.Background()
	s := NewCaptchaService(r)
	if s.NeedCaptcha(ctx, "ip") {
		t.Fatal("unexpected captcha")
	}
	for i := 0; i < 3; i++ {
		if e := s.RecordFailure(ctx, "ip"); e != nil {
			t.Fatal(e)
		}
	}
	if !s.NeedCaptcha(ctx, "ip") {
		t.Fatal("missing captcha threshold")
	}
	r.Set(ctx, "captcha:id", "abcd", time.Minute)
	if !s.Verify(ctx, "id", "ABCD") || s.Verify(ctx, "id", "abcd") {
		t.Fatal("single-use verification failed")
	}
	v, e := s.Generate(ctx)
	if e != nil || v.Image == "" || v.ID == "" {
		t.Fatalf("image generation failed: %v", e)
	}
}
