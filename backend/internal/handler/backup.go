package handler

import (
	"auth-api/internal/backup"
	"auth-api/internal/response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func MountBackup(g *gin.RouterGroup, m *backup.Manager) {
	if m == nil {
		return
	}
	g.GET("/system/backup", func(c *gin.Context) {
		v, e := m.Settings(c.Request.Context())
		if dbError(c, e) {
			return
		}
		runs, e := m.Runs(c.Request.Context())
		if !dbError(c, e) {
			response.OK(c, gin.H{"settings": v, "runs": runs})
		}
	})
	g.POST("/system/backup/settings", func(c *gin.Context) {
		var v backup.Settings
		if c.ShouldBindJSON(&v) != nil {
			response.BadRequest(c, "备份参数无效")
			return
		}
		if e := m.Save(c.Request.Context(), v); e != nil {
			response.BadRequest(c, "备份设置保存失败：间隔需 1–720 小时，保留份数需 1–30")
			return
		}
		response.OK(c, nil)
	})
	g.POST("/system/backup/run", func(c *gin.Context) {
		id, e := m.Trigger(c.Request.Context())
		if errors.Is(e, backup.ErrBusy) {
			response.Fail(c, 409, e.Error())
			return
		}
		if !dbError(c, e) {
			response.OK(c, gin.H{"id": id})
		}
	})
	g.GET("/system/backup/:backupId/download", func(c *gin.Context) {
		f, e := m.Open(c.Request.Context(), c.Param("backupId"))
		if e != nil {
			response.Fail(c, 404, "备份文件不存在或尚未完成")
			return
		}
		defer f.Close()
		info, e := f.Stat()
		if dbError(c, e) {
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Disposition", `attachment; filename="`+c.Param("backupId")+`.tar.gz"`)
		c.Header("Content-Type", "application/gzip")
		http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), f)
	})
}
