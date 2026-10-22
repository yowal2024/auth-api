package middleware

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"auth-api/internal/sdkcrypto"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSDKSignTamperingReplayAndTimeWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer r.Close()
	g := gin.New()
	g.Use(func(c *gin.Context) { c.Set("sdkApp", &repo.App{ID: 1000, TokenKey: "key"}) })
	g.Use(SDKSign(r))
	g.POST("/test", func(c *gin.Context) { response.OK(c, true) })
	send := func(body, ts, nonce, sig string) float64 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("X-Ts", ts)
		req.Header.Set("X-Nonce", nonce)
		req.Header.Set("X-Sign", sig)
		g.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return out["code"].(float64)
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := `{"x":1}`
	sig := sdkcrypto.SignBytes("key", ts, "nonce", []byte(body))
	if send(body, ts, "nonce", sig) != 0 {
		t.Fatal("valid signature rejected")
	}
	if send(body, ts, "nonce", sig) != 1014 {
		t.Fatal("replay accepted")
	}
	if send(`{"x":2}`, ts, "new", sig) != 1014 {
		t.Fatal("tampering accepted")
	}
	old := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	if send(body, old, "old", sdkcrypto.SignBytes("key", old, "old", []byte(body))) != 1016 {
		t.Fatal("expired timestamp accepted")
	}
}
func TestSDKCryptoEnvelopeAndCrossAppSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer r.Close()
	store := auth.NewSDKSessionStore(r, time.Hour)
	key := make([]byte, 32)
	rand.Read(key)
	scheme := &repo.Encrypt{Algorithm: "session", Encoding: "base64", SymKey: hex.EncodeToString(key)}
	if e := store.Set(context.Background(), "session", &auth.SDKSession{AppID: 1, AESKey: scheme.SymKey}); e != nil {
		t.Fatal(e)
	}
	run := func(appid int64) *httptest.ResponseRecorder {
		g := gin.New()
		g.Use(func(c *gin.Context) { c.Set("sdkApp", &repo.App{ID: appid}); c.Set("sdkEncrypt", scheme) })
		g.Use(SDKCrypto(store))
		g.POST("/", func(c *gin.Context) {
			var v map[string]any
			if e := c.ShouldBindJSON(&v); e != nil {
				t.Fatal(e)
			}
			response.OK(c, v)
		})
		encrypted, e := sdkcrypto.EncryptPayload(scheme, []byte(`{"answer":42}`))
		if e != nil {
			t.Fatal(e)
		}
		body, _ := json.Marshal(map[string]string{"data": encrypted})
		req := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
		req.Header.Set("X-Session-Id", "session")
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)
		return w
	}
	var out struct {
		Code int
		Data string
	}
	w := run(1)
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil || out.Code != 0 {
		t.Fatalf("%s %v", w.Body.String(), e)
	}
	plain, e := sdkcrypto.DecryptPayload(scheme, out.Data)
	if e != nil || string(plain) != `{"answer":42}` {
		t.Fatalf("bad encrypted response %s %v", plain, e)
	}
	w = run(2)
	if e = json.Unmarshal(w.Body.Bytes(), &out); e != nil || out.Code != 1020 {
		t.Fatalf("cross-app accepted: %s", w.Body.String())
	}
}
