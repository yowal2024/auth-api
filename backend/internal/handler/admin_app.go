package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/geo"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewAdminAppHandler(r *repo.AppRepo, clients ...*redis.Client) *AdminAppHandler {
	h := &AdminAppHandler{apps: r}
	if len(clients) > 0 {
		h.rdb = clients[0]
	}
	return h
}

func validApp(c *gin.Context, v *repo.App) bool {
	if v.GeoProvider != "" && v.GeoProvider != "tencent" && v.GeoProvider != "amap" && v.GeoProvider != "dbip" {
		response.BadRequest(c, "IP 经纬度服务无效")
		return false
	}
	if (v.Type != "user" && v.Type != "kami") || (v.Mode != "free" && v.Mode != "vip") || v.UserTokenTTL < 1 || v.UserTokenTTL > 31536000 || v.Encryption < 0 || (v.Status != 0 && v.Status != 1) {
		response.BadRequest(c, "应用类型、模式、有效期或状态无效")
		return false
	}
	return true
}
func (h *AdminAppHandler) List(c *gin.Context) {
	v, e := h.apps.ListByApp(c.Request.Context())
	if !dbError(c, e) {
		if a := middleware.CurrentAdmin(c); a == nil || a.Role != "super" {
			items := make([]gin.H, 0, len(v))
			for _, app := range v {
				items = append(items, gin.H{"id": app.ID, "name": app.Name, "type": app.Type, "mode": app.Mode, "logo": app.Logo, "status": app.Status})
			}
			response.OK(c, items)
			return
		}
		response.OK(c, v)
	}
}
func applyApp(v *repo.App, r appReq) {
	v.Name = r.Name
	if r.Type != "" {
		v.Type = r.Type
	}
	if r.Mode != "" {
		v.Mode = r.Mode
	}
	v.Logo = r.Logo
	v.Pkg = r.Pkg
	if r.BindDevice != nil {
		v.BindDevice = *r.BindDevice
	}
	if r.BindRegion != nil {
		v.BindRegion = *r.BindRegion
	}
	if r.DiffUpdate != nil {
		v.DiffUpdate = *r.DiffUpdate
	}
	if r.RandomPackage != nil {
		v.RandomPackage = *r.RandomPackage
	}
	if r.Encryption != nil {
		v.Encryption = *r.Encryption
	}
	if r.BruteForce != nil {
		v.BruteForce = *r.BruteForce
	}
	if r.MtlsEnabled != nil {
		v.MtlsEnabled = *r.MtlsEnabled
	}
	if r.TokenKey != "" {
		v.TokenKey = r.TokenKey
	}
	if r.UserTokenTTL != nil {
		v.UserTokenTTL = *r.UserTokenTTL
	}
	if r.GeoProvider != nil {
		v.GeoProvider = *r.GeoProvider
	}
	if r.GeoKey != nil {
		v.GeoKey = *r.GeoKey
	}
	if r.IPDistanceMax != nil {
		v.IPDistanceMax = *r.IPDistanceMax
	}
	if r.Status != nil {
		v.Status = *r.Status
	}
	if v.Type == "" {
		v.Type = "user"
	}
	if v.Mode == "" {
		v.Mode = "free"
	}
}
func (h *AdminAppHandler) Create(c *gin.Context) {
	var r appReq
	if c.ShouldBindJSON(&r) != nil || r.Name == "" {
		response.BadRequest(c, "应用名称不能为空")
		return
	}
	key, e := auth.RandomToken()
	if dbError(c, e) {
		return
	}
	v := repo.App{TokenKey: key, Status: 1, CreateTime: time.Now().Unix(), UserTokenTTL: 86400, GeoProvider: geo.LocalProvider}
	applyApp(&v, r)
	if v.GeoProvider == "" {
		v.GeoProvider = geo.LocalProvider
	}
	if r.InheritFrom > 0 {
		source, err := h.apps.GetByID(c.Request.Context(), r.InheritFrom)
		if dbError(c, err) {
			return
		}
		if source == nil {
			response.BadRequest(c, "继承应用不存在")
			return
		}
		// Copy policy settings only. Each new app keeps its identity and token key.
		v.Mode, v.BindDevice, v.BindRegion = source.Mode, source.BindDevice, source.BindRegion
		v.DiffUpdate, v.RandomPackage, v.Encryption = source.DiffUpdate, source.RandomPackage, source.Encryption
		v.UserTokenTTL, v.GeoProvider, v.GeoKey = source.UserTokenTTL, source.GeoProvider, source.GeoKey
		v.IPDistanceMax, v.BruteForce = source.IPDistanceMax, source.BruteForce
		// mTLS requires a new app-specific CA and must be enabled after issuance.
		v.MtlsEnabled = false
	}
	if !validApp(c, &v) {
		return
	}
	if !h.validateRegionService(c, &v) {
		return
	}
	if r.InheritFrom <= 0 && v.Encryption != 0 {
		response.BadRequest(c, "请先创建应用，再配置该应用的加密方案")
		return
	}
	id, e := h.apps.CreateInherited(c.Request.Context(), &v, r.InheritFrom)
	if !dbError(c, e) {
		v.ID = id
		c.Set("audit.appid", id)
		response.OK(c, v)
	}
}
func (h *AdminAppHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var r appReq
	if c.ShouldBindJSON(&r) != nil || r.Name == "" {
		response.BadRequest(c, "应用名称不能为空")
		return
	}
	v, e := h.apps.GetByID(c.Request.Context(), id)
	if dbError(c, e) {
		return
	}
	if v == nil {
		response.Fail(c, 404, "应用不存在")
		return
	}
	applyApp(v, r)
	if !validApp(c, v) {
		return
	}
	if !h.validateRegionService(c, v) {
		return
	}
	if v.Encryption > 0 {
		ok, err := h.apps.HasEncryption(c.Request.Context(), id, v.Encryption)
		if dbError(c, err) {
			return
		}
		if !ok {
			response.BadRequest(c, "加密方案不属于当前应用")
			return
		}
	}
	err := h.apps.Update(c.Request.Context(), v)
	if errors.Is(err, repo.ErrEncryptionUnavailable) {
		response.BadRequest(c, err.Error())
		return
	}
	if !dbError(c, err) {
		response.OK(c, v)
	}
}
func (h *AdminAppHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if ok && !dbError(c, h.apps.Delete(c.Request.Context(), id)) {
		if h.rdb != nil {
			err := redis.NewScript(`
local members = redis.call('ZRANGE', KEYS[1], 0, -1)
for _, member in ipairs(members) do
  redis.call('ZREM', KEYS[2], member)
  redis.call('DEL', 'online:' .. member)
  redis.call('DEL', 'presence:sessions:' .. member)
end
return redis.call('DEL', KEYS[1])`).Run(c.Request.Context(), h.rdb, []string{fmt.Sprintf("presence:%d", id), "presence:0"}).Err()
			if dbError(c, err) {
				return
			}
		}
		response.OK(c, nil)
	}
}
