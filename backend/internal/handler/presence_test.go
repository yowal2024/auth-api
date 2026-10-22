package handler

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http/httptest"
	"testing"
)

func TestSessionPresence(t *testing.T) {
	r := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	defer r.Close()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)
	change := func(action, member string) {
		t.Helper()
		if err := changePresence(c, r, 1000, 7, action, member, 300); err != nil {
			t.Fatal(err)
		}
	}
	online := func(want int64) {
		t.Helper()
		n, err := r.ZCard(context.Background(), "presence:1000").Result()
		if err != nil || n != want {
			t.Fatalf("online=%d want=%d err=%v", n, want, err)
		}
	}
	change("touch", "1:a")
	change("touch", "1:b")
	change("touch", "2:c")
	online(1)
	change("logout", "1:a")
	online(1)
	change("device", "1:")
	online(1)
	change("logout", "2:c")
	online(0)
	change("touch", "1:a")
	if err := clearUserPresence(c, r, 1000, 7); err != nil {
		t.Fatal(err)
	}
	online(0)
	if n := r.Exists(context.Background(), "presence:sessions:1000:7", "online:1000:7").Val(); n != 0 {
		t.Fatal("presence keys survived reset")
	}
	change("touch", "2:c")
	online(1)
	change("device", "2:")
	online(0)
}
