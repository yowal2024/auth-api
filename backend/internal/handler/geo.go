package handler

import (
	"auth-api/internal/geo"
	"auth-api/internal/middleware"
	"auth-api/internal/response"
	"errors"
	"github.com/gin-gonic/gin"
)

func MountLocalGeo(admin *gin.RouterGroup, local *geo.LocalDB) {
	admin.GET("/geo/local", func(c *gin.Context) { response.OK(c, local.Status()) })
	admin.GET("/geo/local/lookup", func(c *gin.Context) {
		p, err := local.Lookup(c.Query("ip"))
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		response.OK(c, gin.H{"lat": p.Lat, "lng": p.Lng, "nation": p.Nation, "province": p.Province, "city": p.City})
	})
	admin.POST("/geo/local/update", middleware.RequireSuper(), func(c *gin.Context) {
		status, err := local.Update(c.Request.Context())
		if errors.Is(err, geo.ErrUpdateBusy) {
			response.Fail(c, 409, err.Error())
			return
		}
		if err != nil {
			response.ServerError(c, err.Error())
			return
		}
		response.OK(c, status)
	})
}
