package handler

import (
	"auth-api/internal/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func SDKMetrics(r *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := middleware.CurrentApp(c).ID
		ctx := c.Request.Context()
		day := time.Now().Format("2006-01-02")
		_, e := r.TxPipelined(ctx, func(p redis.Pipeliner) error {
			for _, app := range []int64{0, id} {
				p.Incr(ctx, fmt.Sprintf("api:total:%d", app))
				k := fmt.Sprintf("api:day:%d:%s", app, day)
				p.Incr(ctx, k)
				p.Expire(ctx, k, 35*24*time.Hour)
			}
			return nil
		})
		if e != nil {
			_ = c.Error(e)
		}
		c.Next()
	}
}
