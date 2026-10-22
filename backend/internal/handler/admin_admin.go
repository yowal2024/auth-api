package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
)

func MountAdmins(g *gin.RouterGroup, r *repo.AdminRepo, store *auth.SessionStore) {
	g.GET("/admin/list", func(c *gin.Context) {
		v, e := r.List(c.Request.Context())
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/admin", func(c *gin.Context) {
		var v adminReq
		if c.ShouldBindJSON(&v) != nil || (v.Role != "admin" && v.Role != "super") {
			response.BadRequest(c, "管理员参数无效")
			return
		}
		ctx := c.Request.Context()
		if v.Password != "" && (len(v.Password) < 6 || len(v.Password) > 18) {
			response.BadRequest(c, "密码长度需在 6-18 位")
			return
		}
		if v.ID > 0 {
			old, e := r.GetByID(ctx, v.ID)
			if dbError(c, e) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "管理员不存在")
				return
			}
			hash := ""
			if v.Password != "" {
				if len(v.Password) < 6 || len(v.Password) > 18 {
					response.BadRequest(c, "密码长度需在 6-18 位")
					return
				}
				hash, e = auth.HashPassword(v.Password)
				if dbError(c, e) {
					return
				}
			}
			if e = r.Manage(ctx, v.ID, v.Role, hash, false); e != nil {
				response.BadRequest(c, e.Error())
				return
			}
			if dbError(c, store.DeleteByAdmin(ctx, v.ID)) {
				return
			}
			response.OK(c, gin.H{"id": v.ID})
			return
		}
		if len(v.Username) < 5 || len(v.Username) > 12 || len(v.Password) < 6 || len(v.Password) > 18 {
			response.BadRequest(c, "账号或密码长度无效")
			return
		}
		hash, e := auth.HashPassword(v.Password)
		if dbError(c, e) {
			return
		}
		id, e := r.Create(ctx, v.Username, hash, v.Role)
		if !dbError(c, e) {
			response.OK(c, gin.H{"id": id})
		}
	})
	g.DELETE("/admin/:id", func(c *gin.Context) {
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		if id == middleware.CurrentAdmin(c).AdminID {
			response.BadRequest(c, "不能删除自己")
			return
		}
		a, e := r.GetByID(c.Request.Context(), id)
		if dbError(c, e) {
			return
		}
		if a == nil {
			response.Fail(c, 404, "管理员不存在")
			return
		}
		if e = r.Manage(c.Request.Context(), id, "", "", true); e != nil {
			response.BadRequest(c, e.Error())
			return
		}
		if !dbError(c, store.DeleteByAdmin(c.Request.Context(), id)) {
			response.OK(c, nil)
		}
	})
}
