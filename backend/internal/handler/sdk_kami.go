package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/geo"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func kamiUser(k *repo.Kami) *repo.User {
	return &repo.User{ID: k.ID, AppID: k.AppID, Username: k.Card, Nickname: k.Nickname, Vip: k.Vip, Mi: k.Mi, Region: k.Region, LastIP: k.LastIP, LastLoginTime: k.LastLoginTime, BindLimit: k.BindLimit, RegionLimit: k.RegionLimit, IPList: k.IPList, BanExpire: k.BanExpire, CreateTime: k.UseTime, GroupID: k.GroupID, GroupName: k.GroupName}
}
func (s *SDKService) kamiLogon(c *gin.Context) {
	a := middleware.CurrentApp(c)
	var r sdkKamiLogonReq
	if c.ShouldBindJSON(&r) != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if r.Card == "" {
		r.Card = r.Username
	}
	if r.Card == "" {
		response.BadRequest(c, "请输入卡密")
		return
	}
	if a.BindDevice && r.Mi == "" {
		response.Fail(c, 1011, "缺少机器码 mi")
		return
	}
	if s.blocked(c, r.Mi) {
		return
	}
	if !s.cardAttempt(c, r.Card, r.Mi) {
		return
	}
	failKey := fmt.Sprintf("recovery:brute:kami:%d:%s", a.ID, c.ClientIP())
	if a.BruteForce && !s.allowLoginAttempt(c, failKey) {
		return
	}
	if !s.runTrigger(c, "before_login", gin.H{"username": r.Card, "card": r.Card, "mi": r.Mi, "appid": a.ID, "isKami": true}, nil) {
		return
	}
	ctx := c.Request.Context()
	tx, e := s.db.BeginTxx(ctx, nil)
	if dbError(c, e) {
		return
	}
	defer tx.Rollback()
	var k repo.Kami
	if e = tx.GetContext(ctx, &k, repo.SQL(kamiSelect+" WHERE appid=? AND card=? FOR UPDATE"), a.ID, r.Card); e != nil {
		if !errors.Is(e, sql.ErrNoRows) {
			dbError(c, e)
			return
		}
		response.Fail(c, 2001, "卡密无效或不可用")
		return
	}
	if k.Status == 2 {
		response.Fail(c, 2003, "卡密无效或不可用")
		return
	}
	if k.Status == 1 && k.UseUser != k.Card {
		response.Fail(c, 2002, "卡密无效或不可用")
		return
	}
	now := time.Now().Unix()
	if k.Status == 0 {
		var group repo.KamiGroup
		if dbError(c, tx.GetContext(ctx, &group, repo.SQL("SELECT id,appid,name,type,value,remark FROM {kami_groups} WHERE appid=? AND id=?"), a.ID, k.GroupID)) {
			return
		}
		if group.Type != "vip" || group.Value <= 0 || group.Value > 315360000000 {
			response.BadRequest(c, "卡密面值无效")
			return
		}
		k.Status = 1
		k.UseTime = now
		k.UseUser = k.Card
		k.Vip = now + group.Value
		if _, e = tx.ExecContext(ctx, repo.SQL("UPDATE {kamis} SET status=1,use_user=?,use_time=?,vip=? WHERE appid=? AND id=? AND status=0"), k.Card, now, k.Vip, a.ID, k.ID); dbError(c, e) {
			return
		}
	}
	if !s.checkUser(c, kamiUser(&k), true) {
		return
	}
	if a.BindDevice {
		if e = bindKamiDevice(c, tx, &k, r.Mi); e != nil {
			response.Fail(c, 1008, e.Error())
			return
		}
	}
	if _, e = tx.ExecContext(ctx, repo.SQL("UPDATE {kamis} SET mi=?,last_ip=?,last_login_time=?,last_active_time=UNIX_TIMESTAMP() WHERE appid=? AND id=?"), r.Mi, c.ClientIP(), now, a.ID, k.ID); dbError(c, e) {
		return
	}
	regionUser := kamiUser(&k)
	if !s.checkRegionTx(c, regionUser, true, tx) {
		return
	}
	k.Region = regionUser.Region
	k.IPList = regionUser.IPList
	if dbError(c, tx.Commit()) {
		return
	}
	if a.BruteForce {
		if err := s.rdb.Del(ctx, failKey).Err(); err != nil {
			c.Error(err)
		}
	}
	s.cardSuccess(c)
	k.Mi = r.Mi
	u := kamiUser(&k)
	if !s.runTrigger(c, "after_login", gin.H{"id": k.ID, "username": k.Card, "isKami": true, "appid": a.ID}, u) {
		return
	}
	token, e := s.createUserSession(c, &auth.UserSession{UID: k.ID, AppID: a.ID, Username: k.Card, Mi: r.Mi, LastIP: c.ClientIP(), IsKami: true}, time.Duration(a.UserTokenTTL)*time.Second)
	if !dbError(c, e) {
		c.Set("sdkIsKami", true)
		s.online(c, k.ID)
		c.Set("sdkLogRegion", u.Region)
		if !s.touchDevice(c, -k.ID, r.Mi, u.Region) {
			return
		}
		s.writeUserLog(c, -k.ID, "login")
		profile := sdkUserProfile(u)
		profile["isKami"] = true
		profile["card"] = k.Card
		profile["regionLimit"] = k.RegionLimit
		response.OK(c, gin.H{"token": token, "user": profile, "serverTime": now})
	}
}

// Negative UID distinguishes card-account devices from ordinary users without changing schema.
func bindKamiDevice(c *gin.Context, tx *sqlx.Tx, k *repo.Kami, mi string) error {
	var n, exists int
	ctx := c.Request.Context()
	if e := tx.GetContext(ctx, &exists, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid=? AND mi=?"), -k.ID, mi); e != nil {
		return e
	}
	if exists == 0 {
		if e := tx.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid=?"), -k.ID); e != nil {
			return e
		}
		if n >= 1+k.BindLimit {
			return fmt.Errorf("设备数达上限")
		}
		_, e := tx.ExecContext(ctx, repo.SQL("INSERT INTO {user_devices} (uid,mi,device,os,ip,region,bind_time,last_time) VALUES (?,?,'','',?,'',?,?)"), -k.ID, mi, c.ClientIP(), time.Now().Unix(), time.Now().Unix())
		return e
	}
	_, e := tx.ExecContext(ctx, repo.SQL("UPDATE {user_devices} SET last_time=?,ip=? WHERE uid=? AND mi=?"), time.Now().Unix(), c.ClientIP(), -k.ID, mi)
	return e
}
func (s *SDKService) authorizeKami(c *gin.Context) {
	session := middleware.CurrentSDKUser(c)
	var k repo.Kami
	err := s.db.GetContext(c.Request.Context(), &k, repo.SQL(kamiSelect+" WHERE appid=? AND id=?"), session.AppID, session.UID)
	if errors.Is(err, sql.ErrNoRows) {
		response.Fail(c, 1006, "未登录或登录已过期")
		c.Abort()
		return
	}
	if dbError(c, err) {
		c.Abort()
		return
	}
	if k.Status != 1 {
		response.Fail(c, 2003, "卡密未激活或已冻结")
		c.Abort()
		return
	}
	if k.UseUser != k.Card {
		response.Fail(c, 1006, "卡密已用于充值，请重新登录")
		c.Abort()
		return
	}
	u := kamiUser(&k)
	if !s.checkUser(c, u, !stringsHasSuffix(c.FullPath(), "/kamiTopup") && !stringsHasSuffix(c.FullPath(), "/logout")) || s.blocked(c, session.Mi) {
		c.Abort()
		return
	}
	if middleware.CurrentApp(c).BindDevice {
		if c.GetHeader("X-Mi") == "" || c.GetHeader("X-Mi") != session.Mi {
			response.Fail(c, 1006, "设备不匹配")
			c.Abort()
			return
		}
		var n int
		if dbError(c, s.db.GetContext(c.Request.Context(), &n, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid=? AND mi=? AND id=?"), -k.ID, session.Mi, session.DeviceID)) {
			c.Abort()
			return
		}
		if n == 0 {
			_ = s.sessions.Delete(c.Request.Context(), c.GetString("sdkToken"))
			response.Fail(c, 1006, "设备已解绑")
			c.Abort()
			return
		}
	}
	if !s.checkRegion(c, u, true) {
		c.Abort()
		return
	}
	if !s.touchDevice(c, -k.ID, session.Mi, u.Region) {
		c.Abort()
		return
	}
	c.Set("sdkIsKami", true)
	c.Set("sdkUserRecord", u)
	s.online(c, k.ID)
	c.Next()
}
func (s *SDKService) checkRegion(c *gin.Context, u *repo.User, kami bool) bool {
	if !middleware.CurrentApp(c).BindRegion {
		return true
	}
	tx, e := s.db.BeginTxx(c.Request.Context(), nil)
	if dbError(c, e) {
		return false
	}
	defer tx.Rollback()
	if !s.checkRegionTx(c, u, kami, tx) {
		return false
	}
	return !dbError(c, tx.Commit())
}
func (s *SDKService) checkRegionTx(c *gin.Context, u *repo.User, kami bool, tx *sqlx.Tx) bool {
	app := middleware.CurrentApp(c)
	if !app.BindRegion {
		return true
	}
	var point geo.Point
	var e error
	if app.GeoProvider == geo.LocalProvider {
		// Local lookups are fast and must immediately use the newly installed database.
		point, e = s.LocalGeo.Lookup(c.ClientIP())
	} else {
		point, e = geo.Resolve(c.Request.Context(), tx, app.GeoProvider, app.GeoKey, c.ClientIP())
	}
	if e != nil {
		response.ServerError(c, "地域查询失败")
		return false
	}
	table := "users"
	if kami {
		table = "kamis"
	}
	ctx := c.Request.Context()
	var raw string
	if dbError(c, tx.GetContext(ctx, &raw, repo.SQL("SELECT COALESCE(ip_list,'') FROM {"+table+"} WHERE appid=? AND id=? FOR UPDATE"), u.AppID, u.ID)) {
		return false
	}
	entries := []geo.IPEntry{}
	if raw != "" {
		if e = json.Unmarshal([]byte(raw), &entries); e != nil {
			response.ServerError(c, "已保存的地域数据格式错误")
			return false
		}
	}
	entries, e = geo.Check(entries, point, c.ClientIP(), u.RegionLimit, app.IPDistanceMax)
	if e != nil {
		response.Fail(c, 1009, e.Error())
		return false
	}
	encoded, _ := json.Marshal(entries)
	u.IPList = string(encoded)
	u.Region = point.Province + point.City
	if _, e = tx.ExecContext(ctx, repo.SQL("UPDATE {"+table+"} SET ip_list=?,region=? WHERE appid=? AND id=?"), u.IPList, u.Region, u.AppID, u.ID); dbError(c, e) {
		return false
	}
	return true
}
func (s *SDKService) writeUserLog(c *gin.Context, uid int64, action string) {
	var enabled int
	if s.db.GetContext(c.Request.Context(), &enabled, repo.SQL("SELECT user_log_enabled FROM {settings} WHERE id=1")) != nil || enabled == 0 {
		return
	}
	_, e := s.db.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {user_logs} (uid,type,action,data,ip,region,create_time) VALUES (?,'sdk',?,'',?,?,?)"), uid, action, c.ClientIP(), currentLogRegion(c), time.Now().Unix())
	if e != nil {
		c.Error(e)
	}
}
func MountKamiDevices(g *gin.RouterGroup, d *sqlx.DB, clients ...*redis.Client) {
	var cache *redis.Client
	if len(clients) > 0 {
		cache = clients[0]
	}
	validate := func(c *gin.Context) (int64, bool) {
		a, ok := requireAppID(c)
		if !ok {
			return 0, false
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return 0, false
		}
		var n int
		if dbError(c, d.GetContext(c.Request.Context(), &n, repo.SQL("SELECT COUNT(*) FROM {kamis} WHERE appid=? AND id=?"), a, id)) {
			return 0, false
		}
		if n == 0 {
			response.Fail(c, 404, "卡密无效或不可用")
			return 0, false
		}
		return id, true
	}
	g.GET("/kami/:id/devices", func(c *gin.Context) {
		id, ok := validate(c)
		if !ok {
			return
		}
		v := []repo.UserDevice{}
		if !dbError(c, d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,uid,mi,device,os,ip,region,bind_time,last_time FROM {user_devices} WHERE uid=? ORDER BY id DESC"), -id)) {
			response.OK(c, v)
		}
	})
	for _, path := range []string{"/kami/:id/devices", "/kami/:id/devices/:deviceId"} {
		g.DELETE(path, func(c *gin.Context) {
			id, ok := validate(c)
			if !ok {
				return
			}
			q := "DELETE FROM {user_devices} WHERE uid=?"
			args := []any{-id}
			if c.Param("deviceId") != "" {
				did, ok := parseIDParam(c, "deviceId")
				if !ok {
					return
				}
				q += " AND id=?"
				args = append(args, did)
			}
			res, e := d.ExecContext(c.Request.Context(), repo.SQL(q), args...)
			if dbError(c, e) {
				return
			}
			n, e := res.RowsAffected()
			if !dbError(c, e) {
				app, _ := requireAppID(c)
				if c.Param("deviceId") == "" {
					if dbError(c, clearUserPresence(c, cache, app, -id)) {
						return
					}
				} else {
					did, _ := strconv.ParseInt(c.Param("deviceId"), 10, 64)
					if dbError(c, clearDevicePresence(c, cache, app, -id, did)) {
						return
					}
				}
				response.OK(c, gin.H{"affected": n})
			}
		})
	}
}
