package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"regexp"
	"time"
)

func MountUserSessions(g *gin.RouterGroup, d *sqlx.DB, r *redis.Client) {
	store := auth.NewUserSessionStore(r, 24*time.Hour)
	for _, kind := range []string{"user", "kami"} {
		kind := kind
		table := "users"
		if kind == "kami" {
			table = "kamis"
		}
		check := func(c *gin.Context) (int64, int64, bool) {
			app, ok := requireAppID(c)
			if !ok {
				return 0, 0, false
			}
			id, ok := parseIDParam(c, "id")
			if !ok {
				return 0, 0, false
			}
			var n int
			if dbError(c, d.GetContext(c.Request.Context(), &n, repo.SQL("SELECT COUNT(*) FROM {"+table+"} WHERE appid=? AND id=?"), app, id)) {
				return 0, 0, false
			}
			if n == 0 {
				response.Fail(c, 404, "账号不存在")
				return 0, 0, false
			}
			return app, id, true
		}
		base := "/" + kind + "/:id/sessions"
		g.GET(base, func(c *gin.Context) {
			app, id, ok := check(c)
			if !ok {
				return
			}
			v, e := store.List(c.Request.Context(), app, id, kind == "kami")
			if !dbError(c, e) {
				response.OK(c, v)
			}
		})
		g.DELETE(base, func(c *gin.Context) {
			app, id, ok := check(c)
			if !ok {
				return
			}
			if kind == "kami" {
				id = -id
			}
			if !dbError(c, clearUserPresence(c, r, app, id)) {
				response.OK(c, nil)
			}
		})
		g.DELETE(base+"/:sessionId", func(c *gin.Context) {
			app, id, ok := check(c)
			if !ok {
				return
			}
			sid := c.Param("sessionId")
			if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(sid) {
				response.BadRequest(c, "会话标识无效")
				return
			}
			v, e := store.Revoke(c.Request.Context(), app, id, kind == "kami", sid)
			if errors.Is(e, auth.ErrSession) {
				response.Fail(c, 404, "会话已失效")
				return
			}
			if dbError(c, e) {
				return
			}
			if kind == "kami" {
				id = -id
			}
			if !dbError(c, changePresence(c, r, app, id, "logout", fmt.Sprintf("%d:%s", v.DeviceID, sid), 0)) {
				response.OK(c, nil)
			}
		})
	}
}
