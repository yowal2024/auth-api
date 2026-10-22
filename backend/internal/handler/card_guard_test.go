package handler

import (
	"auth-api/internal/repo"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCardCooldownResistsParallelGuessesAndRecovers(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer r.Close()
	s := NewSDKService(nil, r)
	var success atomic.Int64
	g := gin.New()
	g.SetTrustedProxies(nil)
	g.POST("/", func(c *gin.Context) {
		c.Set("sdkApp", &repo.App{ID: 1})
		if s.cardAttempt(c, "guess", "device") {
			success.Add(1)
		}
	})
	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/", nil)
			req.Header.Set("X-Forwarded-For", "8.8.8.8")
			g.ServeHTTP(httptest.NewRecorder(), req)
		}()
	}
	wg.Wait()
	if success.Load() > 6 || success.Load() == 0 {
		t.Fatal("parallel bypass", success.Load())
	}
	m.FastForward(6 * time.Minute)
	g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil))
	if success.Load() != 7 {
		t.Fatal("cooldown did not recover", success.Load())
	}
}
