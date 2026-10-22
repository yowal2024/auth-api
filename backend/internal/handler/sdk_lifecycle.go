package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Bind sessions to the actual binding row, not just a reusable machine code.
func (s *SDKService) createUserSession(c *gin.Context, session *auth.UserSession, ttl time.Duration) (string, error) {
	if middleware.CurrentApp(c).BindDevice {
		id := session.UID
		if session.IsKami {
			id = -id
		}
		if err := s.db.GetContext(c.Request.Context(), &session.DeviceID, repo.SQL("SELECT id FROM {user_devices} WHERE uid=? AND mi=?"), id, session.Mi); err != nil {
			return "", err
		}
	}
	token, err := s.sessions.CreateWithTTL(c.Request.Context(), session, ttl)
	if err == nil {
		c.Set("presenceToken", token)
		c.Set("presenceDevice", session.DeviceID)
		c.Set("presenceTTL", session.TTLSec)
	}
	return token, err
}

func currentLogRegion(c *gin.Context) string {
	if u := currentUser(c); u != nil {
		return u.Region
	}
	return c.GetString("sdkLogRegion")
}

func (s *SDKService) touchDevice(c *gin.Context, uid int64, mi, region string) bool {
	if !middleware.CurrentApp(c).BindDevice {
		return true
	}
	_, err := s.db.ExecContext(c.Request.Context(), repo.SQL("UPDATE {user_devices} SET last_time=?,ip=?,region=? WHERE uid=? AND mi=?"), time.Now().Unix(), c.ClientIP(), region, uid, mi)
	return !dbError(c, err)
}

func clearUserPresence(c *gin.Context, r *redis.Client, appID int64, ids ...int64) error {
	if r == nil || len(ids) == 0 {
		return nil
	}
	ctx := c.Request.Context()
	store := auth.NewUserSessionStore(r, 24*time.Hour)
	for _, id := range ids {
		uid := id
		if uid < 0 {
			uid = -uid
		}
		if err := store.RevokeAll(ctx, appID, uid, id < 0); err != nil {
			return err
		}
	}
	_, err := r.TxPipelined(ctx, func(p redis.Pipeliner) error {
		for _, id := range ids {
			p.Del(ctx, fmt.Sprintf("presence:sessions:%d:%d", appID, id))
			p.Del(ctx, fmt.Sprintf("online:%d:%d", appID, id))
			member := fmt.Sprintf("%d:%d", appID, id)
			p.ZRem(ctx, fmt.Sprintf("presence:%d", appID), member)
			p.ZRem(ctx, "presence:0", member)
		}
		return nil
	})
	return err
}

func clearDevicePresence(c *gin.Context, r *redis.Client, appID, id, deviceID int64) error {
	if r == nil {
		return nil
	}
	uid := id
	if uid < 0 {
		uid = -uid
	}
	store := auth.NewUserSessionStore(r, 24*time.Hour)
	if err := store.RevokeDevice(c.Request.Context(), appID, uid, id < 0, deviceID); err != nil {
		return err
	}
	return changePresence(c, r, appID, id, "device", fmt.Sprintf("%d:", deviceID), 0)
}
