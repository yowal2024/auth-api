package handler

import (
	"auth-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (s *SDKService) MountRaw(g *gin.RouterGroup, files *FileService, versions *VersionService) {
	g.GET("/diff/:name", versions.Diff)
	g.GET("/extend-file/list", files.List)
	g.GET("/extend-file/:id/download", files.Download)
	private := g.Group("")
	private.Use(middleware.SDKToken(s.sessions), s.authorize)
	private.POST("/uploadAvatar", files.SDKAvatar)
	private.GET("/clone", versions.Clone)
	private.POST("/regen-sign", versions.RegenerateSign)
}
