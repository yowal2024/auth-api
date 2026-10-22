package middleware

import (
	"auth-api/internal/repo"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExpensiveSDKRateLimitCannotBeBypassedWithForwardedIP(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer r.Close()
	g := gin.New()
	g.SetTrustedProxies(nil)
	g.Use(func(c *gin.Context) { c.Set("sdkApp", &repo.App{ID: 1}) }, SDKRateLimit(r))
	g.POST("/api/sdk/regen-sign", func(c *gin.Context) { c.JSON(200, gin.H{"code": 0}) })
	call := func() int {
		req := httptest.NewRequest("POST", "/api/sdk/regen-sign", nil)
		req.Header.Set("X-Forwarded-For", "8.8.8.8")
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)
		var out struct{ Code int }
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.Code
	}
	for i := 0; i < 6; i++ {
		if code := call(); code != 0 {
			t.Fatal(code)
		}
	}
	if code := call(); code != 1017 {
		t.Fatal("rate bypass", code)
	}
	m.FastForward(time.Minute)
	if code := call(); code != 0 {
		t.Fatal("rate did not expire", code)
	}
}
