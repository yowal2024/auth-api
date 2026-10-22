package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

func MountUsers(g *gin.RouterGroup, d *sqlx.DB, clients ...*redis.Client) {
	var cache *redis.Client
	if len(clients) > 0 {
		cache = clients[0]
	}
	users := repo.NewUserRepo(d)
	groups := repo.NewUserGroupRepo(d)
	detail := func(c *gin.Context) (*repo.User, bool) {
		a, ok := requireAppID(c)
		if !ok {
			return nil, false
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return nil, false
		}
		v, e := users.GetByID(c.Request.Context(), a, id)
		if dbError(c, e) {
			return nil, false
		}
		if v == nil {
			response.Fail(c, 404, "用户不存在")
			return nil, false
		}
		return v, true
	}
	g.GET("/user/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		page, size := pageParams(c)
		gid, _ := strconv.ParseInt(c.Query("groupId"), 10, 64)
		v, total, e := users.List(c.Request.Context(), a, repo.UserFilter{Page: page, Size: size, GroupID: gid, Field: c.Query("searchField"), Keyword: c.Query("keyword"), Status: c.Query("status"), UserType: c.Query("userType")})
		if dbError(c, e) {
			return
		}
		items := []userListItem{}
		online := make([]any, len(v))
		if len(clients) > 0 && clients[0] != nil && len(v) > 0 {
			keys := make([]string, len(v))
			for i := range v {
				keys[i] = "online:" + strconv.FormatInt(a, 10) + ":" + strconv.FormatInt(v[i].ID, 10)
			}
			var err error
			online, err = clients[0].MGet(c.Request.Context(), keys...).Result()
			if dbError(c, err) {
				return
			}
		}
		for i := range v {
			status := 1
			if v[i].BanExpire > time.Now().Unix() || v[i].BanExpire < 0 {
				status = 0
			}
			items = append(items, userListItem{User: &v[i], Status: status, Online: online[i] != nil})
		}
		response.OK(c, gin.H{"list": items, "total": total})
	})
	g.GET("/user/:id", func(c *gin.Context) {
		if v, ok := detail(c); ok {
			response.OK(c, v)
		}
	})
	g.POST("/user", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r userCreateReq
		if c.ShouldBindJSON(&r) != nil || len(r.Username) < 5 || len(r.Username) > 12 || len(r.Password) < 6 || len(r.Password) > 18 || r.GroupID <= 0 {
			response.BadRequest(c, "账号需 5-12 位，密码需 6-18 位，并选择用户分组")
			return
		}
		group, e := groups.GetByID(c.Request.Context(), a, r.GroupID)
		if dbError(c, e) {
			return
		}
		if group == nil {
			response.BadRequest(c, "用户分组不存在")
			return
		}
		hash, e := auth.HashPassword(r.Password)
		if dbError(c, e) {
			return
		}
		v := repo.User{AppID: a, GroupID: r.GroupID, Username: r.Username, Password: hash, Nickname: r.Nickname, RegisterIP: c.ClientIP(), CreateTime: time.Now().Unix(), IPList: "[]"}
		id, e := users.Create(c.Request.Context(), &v)
		if !dbError(c, e) {
			v.ID = id
			response.OK(c, v)
		}
	})
	g.PUT("/user/:id", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		var r userUpdateReq
		if c.ShouldBindJSON(&r) != nil {
			response.BadRequest(c, "请求体格式错误")
			return
		}
		if err := validateAccountLimits(r.BindLimit, r.RegionLimit); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if r.IPList != nil && (r.ClearIPList == nil || !*r.ClearIPList) {
			if err := validateIPList(*r.IPList); err != nil {
				response.BadRequest(c, err.Error())
				return
			}
		}
		v := map[string]any{}
		if r.Vip != nil {
			v["vip"] = *r.Vip
		}
		if r.Nickname != nil {
			v["nickname"] = *r.Nickname
		}
		if r.Avatar != nil {
			v["avatar"] = *r.Avatar
		}
		if r.BindLimit != nil {
			v["bind_limit"] = *r.BindLimit
		}
		if r.RegionLimit != nil {
			v["region_limit"] = *r.RegionLimit
		}
		if r.IPList != nil {
			v["ip_list"] = *r.IPList
		}
		if r.ClearIPList != nil && *r.ClearIPList {
			v["ip_list"] = "[]"
		}
		if r.BanExpire != nil {
			v["ban_expire"] = *r.BanExpire
		}
		if r.GroupID != nil {
			group, e := groups.GetByID(c.Request.Context(), u.AppID, *r.GroupID)
			if dbError(c, e) {
				return
			}
			if group == nil {
				response.BadRequest(c, "用户分组不存在")
				return
			}
			v["group_id"] = *r.GroupID
		}
		if r.Password != nil && *r.Password != "" {
			if len(*r.Password) < 6 || len(*r.Password) > 18 {
				response.BadRequest(c, "密码长度需在 6-18 位")
				return
			}
			hash, e := auth.HashPassword(*r.Password)
			if dbError(c, e) {
				return
			}
			v["password"] = hash
		}
		if !dbError(c, users.UpdatePartial(c.Request.Context(), u.AppID, u.ID, v)) {
			if (v["password"] != nil || (r.BanExpire != nil && (*r.BanExpire < 0 || *r.BanExpire > time.Now().Unix()))) && dbError(c, clearUserPresence(c, cache, u.AppID, u.ID)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.DELETE("/user/:id", func(c *gin.Context) {
		if u, ok := detail(c); ok && !dbError(c, users.Delete(c.Request.Context(), u.AppID, u.ID)) {
			if dbError(c, clearUserPresence(c, cache, u.AppID, u.ID)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.POST("/user/batch-delete", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r batchDeleteReq
		if c.ShouldBindJSON(&r) != nil || len(r.IDs) > 1000 {
			response.BadRequest(c, "无效的用户列表")
			return
		}
		n, e := users.BatchDelete(c.Request.Context(), a, r.IDs)
		if !dbError(c, e) {
			if dbError(c, clearUserPresence(c, cache, a, r.IDs...)) {
				return
			}
			response.OK(c, gin.H{"affected": n})
		}
	})
	g.POST("/user/reward", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r rewardReq
		if c.ShouldBindJSON(&r) != nil || r.Value <= 0 || r.Value > 1000000 {
			response.BadRequest(c, "奖励数值无效")
			return
		}
		if (r.Target != "all" && r.Target != "vip" && r.Target != "selected") || (r.Target == "selected" && (len(r.UserIDs) == 0 || len(r.UserIDs) > 1000)) {
			response.BadRequest(c, "请选择有效的奖励范围和用户")
			return
		}
		units := map[string]int64{"second": 1, "minute": 60, "hour": 3600, "day": 86400, "month": 2592000, "year": 31536000}
		unit := units[r.Unit]
		if unit == 0 {
			response.BadRequest(c, "不支持的时间单位")
			return
		}
		n, e := users.Reward(c.Request.Context(), a, r.Value*unit, r.Target, r.UserIDs)
		if !dbError(c, e) {
			response.OK(c, gin.H{"affected": n})
		}
	})
	g.POST("/user/:id/unbind-mi", func(c *gin.Context) {
		if u, ok := detail(c); ok && !dbError(c, users.UpdatePartial(c.Request.Context(), u.AppID, u.ID, map[string]any{"mi": ""})) {
			if _, e := d.ExecContext(c.Request.Context(), repo.SQL("DELETE FROM {user_devices} WHERE uid=?"), u.ID); dbError(c, e) {
				return
			}
			if dbError(c, clearUserPresence(c, cache, u.AppID, u.ID)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.POST("/user/:id/password", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		var v resetPwdReq
		if c.ShouldBindJSON(&v) != nil || len(v.Password) < 6 || len(v.Password) > 18 {
			response.BadRequest(c, "密码长度需在 6-18 位")
			return
		}
		hash, e := auth.HashPassword(v.Password)
		if !dbError(c, e) && !dbError(c, users.UpdatePartial(c.Request.Context(), u.AppID, u.ID, map[string]any{"password": hash})) {
			if dbError(c, clearUserPresence(c, cache, u.AppID, u.ID)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.GET("/user/:id/extends", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		v := []repo.UserExtend{}
		e := d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,uid,e_key,e_value FROM {user_extends} WHERE uid=? ORDER BY id"), u.ID)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/user/:id/extends", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		var r extendsSaveReq
		if c.ShouldBindJSON(&r) != nil {
			response.BadRequest(c, "请求体格式错误")
			return
		}
		items := r.List
		if items == nil {
			items = []extendItem{{ID: r.ID, Key: r.Key, Value: r.Value}}
		}
		for _, v := range items {
			if strings.TrimSpace(v.Key) == "" || len(v.Key) > 64 || v.ID < 0 {
				response.BadRequest(c, "变量名不能为空且不能超过 64 字节，变量 ID 必须有效")
				return
			}
		}
		tx, e := d.BeginTxx(c.Request.Context(), nil)
		if dbError(c, e) {
			return
		}
		defer tx.Rollback()
		// Serialize with SDK writes and user deletion; prevent orphaned variables.
		var owner int64
		e = tx.GetContext(c.Request.Context(), &owner, repo.SQL("SELECT id FROM {users} WHERE appid=? AND id=? FOR UPDATE"), u.AppID, u.ID)
		if errors.Is(e, sql.ErrNoRows) {
			response.Fail(c, 404, "用户不存在")
			return
		}
		if dbError(c, e) {
			return
		}
		for _, v := range items {
			var id int64
			if v.ID > 0 {
				e = tx.GetContext(c.Request.Context(), &id, repo.SQL("SELECT id FROM {user_extends} WHERE uid=? AND id=?"), u.ID, v.ID)
				if errors.Is(e, sql.ErrNoRows) {
					response.Fail(c, 404, "扩展变量不存在")
					return
				}
				if dbError(c, e) {
					return
				}
			}
			e = tx.GetContext(c.Request.Context(), &id, repo.SQL("SELECT id FROM {user_extends} WHERE uid=? AND e_key=? AND id<>? LIMIT 1"), u.ID, v.Key, v.ID)
			if e == nil {
				response.BadRequest(c, "变量名已存在，请修改已有变量")
				return
			}
			if !errors.Is(e, sql.ErrNoRows) && dbError(c, e) {
				return
			}
			if v.ID > 0 {
				_, e = tx.ExecContext(c.Request.Context(), repo.SQL("UPDATE {user_extends} SET e_key=?,e_value=? WHERE uid=? AND id=?"), v.Key, v.Value, u.ID, v.ID)
			} else {
				_, e = tx.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {user_extends} (uid,e_key,e_value) VALUES (?,?,?)"), u.ID, v.Key, v.Value)
			}
			if dbError(c, e) {
				return
			}
		}
		if !dbError(c, tx.Commit()) {
			response.OK(c, nil)
		}
	})
	g.DELETE("/user/:id/extends/:eid", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "eid")
		if !ok {
			return
		}
		_, e := d.ExecContext(c.Request.Context(), repo.SQL("DELETE FROM {user_extends} WHERE uid=? AND id=?"), u.ID, id)
		if !dbError(c, e) {
			response.OK(c, nil)
		}
	})
	g.GET("/user/:id/devices", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		v := []repo.UserDevice{}
		e := d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,uid,mi,device,os,ip,region,bind_time,last_time FROM {user_devices} WHERE uid=? ORDER BY id DESC"), u.ID)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	for _, path := range []string{"/user/:id/devices", "/user/:id/devices/:deviceId"} {
		g.DELETE(path, func(c *gin.Context) {
			u, ok := detail(c)
			if !ok {
				return
			}
			q := "DELETE FROM {user_devices} WHERE uid=?"
			args := []any{u.ID}
			if c.Param("deviceId") != "" {
				id, valid := parseIDParam(c, "deviceId")
				if !valid {
					return
				}
				q += " AND id=?"
				args = append(args, id)
			}
			res, e := d.ExecContext(c.Request.Context(), repo.SQL(q), args...)
			if dbError(c, e) {
				return
			}
			n, e := res.RowsAffected()
			if !dbError(c, e) {
				if c.Param("deviceId") == "" {
					if dbError(c, clearUserPresence(c, cache, u.AppID, u.ID)) {
						return
					}
				} else {
					did, _ := strconv.ParseInt(c.Param("deviceId"), 10, 64)
					if dbError(c, clearDevicePresence(c, cache, u.AppID, u.ID, did)) {
						return
					}
				}
				response.OK(c, gin.H{"affected": n})
			}
		})
	}
	g.GET("/user/:id/logs", func(c *gin.Context) {
		u, ok := detail(c)
		if !ok {
			return
		}
		page, size := pageParams(c)
		v := []repo.UserLog{}
		var n int
		if dbError(c, d.GetContext(c.Request.Context(), &n, repo.SQL("SELECT COUNT(*) FROM {user_logs} WHERE uid=?"), u.ID)) {
			return
		}
		e := d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT id,uid,type,action,COALESCE(data,'') AS data,ip,region,create_time FROM {user_logs} WHERE uid=? ORDER BY id DESC LIMIT ? OFFSET ?"), u.ID, size, (page-1)*size)
		if !dbError(c, e) {
			response.OK(c, gin.H{"list": v, "total": n})
		}
	})
}
