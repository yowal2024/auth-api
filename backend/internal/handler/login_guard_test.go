package handler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestConcurrentLoginAttemptLimit(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer r.Close()
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 30 {
		wg.Go(func() {
			n, err := reserveLoginAttempt.Run(context.Background(), r, []string{"attempts"}).Int()
			if err != nil {
				t.Error(err)
				return
			}
			allowed.Add(int32(n))
		})
	}
	wg.Wait()
	if allowed.Load() != 5 {
		t.Fatalf("allowed %d parallel guesses", allowed.Load())
	}
	m.FastForward(301 * time.Second)
	if n, err := reserveLoginAttempt.Run(context.Background(), r, []string{"attempts"}).Int(); err != nil || n != 1 {
		t.Fatalf("lockout did not expire: %d %v", n, err)
	}
}
