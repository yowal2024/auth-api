package handler

import (
	"auth-api/internal/geo"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"strings"
)

func MountGeoSettings(admin *gin.RouterGroup, db *sqlx.DB, local *geo.LocalDB) {
	admin.GET("/system/geo", func(c *gin.Context) {
		settings, err := repo.LoadGeoSettings(c.Request.Context(), db)
		if dbError(c, err) {
			return
		}
		legacy := []repo.LegacyGeoSetting{}
		if settings.Provider == "" {
			legacy, err = repo.LegacyGeoSettings(c.Request.Context(), db)
			if dbError(c, err) {
				return
			}
		}
		response.OK(c, gin.H{"provider": settings.Provider, "key": settings.Key, "legacyApps": legacy})
	})
	admin.POST("/system/geo", func(c *gin.Context) {
		var settings repo.GeoSettings
		if c.ShouldBindJSON(&settings) != nil || (settings.Provider != geo.LocalProvider && settings.Provider != "tencent" && settings.Provider != "amap") {
			response.BadRequest(c, "请选择有效的 IP 定位服务")
			return
		}
		settings.Key = strings.TrimSpace(settings.Key)
		if settings.Provider == geo.LocalProvider {
			if !local.Status().Ready {
				response.BadRequest(c, "请先安装本地 IP 库")
				return
			}
			settings.Key = ""
		} else if settings.Key == "" || len(settings.Key) > 128 {
			response.BadRequest(c, "请填写有效的服务 KEY（最多 128 字节）")
			return
		}
		if !dbError(c, repo.SaveGeoSettings(c.Request.Context(), db, settings)) {
			response.OK(c, settings)
		}
	})
}

func (h *AdminAppHandler) validateRegionService(c *gin.Context, app *repo.App) bool {
	if !app.BindRegion {
		return true
	}
	settings, err := h.apps.GeoSettings(c.Request.Context(), app)
	if dbError(c, err) {
		return false
	}
	if settings.Provider == geo.LocalProvider && !h.LocalGeo.Status().Ready {
		response.BadRequest(c, "请先在系统设置 → IP 定位服务中安装本地 IP 库")
		return false
	}
	if settings.Provider != geo.LocalProvider && strings.TrimSpace(settings.Key) == "" {
		response.BadRequest(c, "请先在系统设置 → IP 定位服务中配置服务 KEY")
		return false
	}
	return true
}

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
