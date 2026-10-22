package middleware

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func AdminAuth(store *auth.SessionStore, admins *repo.AdminRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		s, e := store.Get(c.Request.Context(), token)
		if e != nil || s == nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// Recheck persisted role so deleted/demoted accounts cannot retain privileges.
		a, e := admins.GetByID(c.Request.Context(), s.AdminID)
		if e != nil {
			response.ServerError(c, "读取管理员失败")
			c.Abort()
			return
		}
		if a == nil || s.Credential != auth.CredentialRevision(a.Password) {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		s.Role = a.Role
		s.Username = a.Username
		s.Nickname = a.Nickname
		s.Avatar = a.Avatar
		c.Set("admin", s)
		c.Set("adminToken", token)
		c.Next()
	}
}
func CurrentAdmin(c *gin.Context) *auth.Session {
	v, _ := c.Get("admin")
	a, _ := v.(*auth.Session)
	return a
}
func RequireSuper() gin.HandlerFunc {
	return func(c *gin.Context) {
		a := CurrentAdmin(c)
		if a == nil || a.Role != "super" {
			response.Fail(c, 403, "需要超级管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// Ordinary administrators operate accounts; all other management endpoints
// require super privileges by default, including newly added endpoints.
func AdminPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		a := CurrentAdmin(c)
		if a != nil && a.Role == "super" {
			c.Next()
			return
		}
		parts := strings.SplitN(strings.TrimPrefix(c.FullPath(), "/"), "/", 3)
		allowed := false
		if a != nil && a.Role == "admin" && len(parts) == 3 {
			route := parts[2]
			switch route {
			case "me", "me/profile", "me/password", "logout", "upload/avatar":
				allowed = true
			case "app/list", "dashboard":
				allowed = c.Request.Method == "GET"
			case "log/list", "log/actions":
				allowed = c.Request.Method == "GET" && c.Query("category") == "user" && c.Query("type") == ""
			default:
				for _, family := range []string{"user", "kami", "block", "notice"} {
					if route == family || strings.HasPrefix(route, family+"/") {
						allowed = true
						break
					}
				}
			}
		}
		if !allowed {
			response.Fail(c, 403, "需要超级管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
func AppID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.GetHeader("X-App-Id"), 10, 64)
		c.Set("appid", id)
		c.Next()
	}
}
