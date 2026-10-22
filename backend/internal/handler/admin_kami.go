package handler

import (
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"math/big"
	"strconv"
	"time"
)

const kamiSelect = "SELECT id,appid,group_id,group_name,card,status,use_user,use_time,remark,create_time,creator,create_ip,nickname,vip,ban_expire,bind_limit,region_limit,COALESCE(ip_list,'') AS ip_list,mi,region,last_ip,last_login_time,last_active_time FROM {kamis}"

func randomCard(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		n, e := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if e != nil {
			return "", e
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
func MountKami(g *gin.RouterGroup, d *sqlx.DB, clients ...*redis.Client) {
	var cache *redis.Client
	if len(clients) > 0 {
		cache = clients[0]
	}
	groups := repo.NewKamiGroupRepo(d)
	g.GET("/kami/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		page, size := pageParams(c)
		q := "appid=?"
		args := []any{a}
		if v := c.Query("status"); v != "" {
			q += " AND status=?"
			args = append(args, v)
		}
		if id, _ := strconv.ParseInt(c.Query("groupId"), 10, 64); id > 0 {
			q += " AND group_id=?"
			args = append(args, id)
		}
		if v := c.Query("keyword"); v != "" {
			q += " AND (card LIKE ? OR use_user LIKE ? OR remark LIKE ?)"
			args = append(args, "%"+v+"%", "%"+v+"%", "%"+v+"%")
		}
		var n int
		ctx := c.Request.Context()
		if dbError(c, d.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {kamis} WHERE "+q), args...)) {
			return
		}
		v := []repo.Kami{}
		args = append(args, size, (page-1)*size)
		if !dbError(c, d.SelectContext(ctx, &v, repo.SQL(kamiSelect+" WHERE "+q+" ORDER BY id DESC LIMIT ? OFFSET ?"), args...)) {
			response.OK(c, gin.H{"list": v, "total": n})
		}
	})
	g.POST("/kami/generate", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r kamiGenerateReq
		if c.ShouldBindJSON(&r) != nil || r.Count < 1 || r.Count > 1000 || r.Length < 8 || r.Length > 64 || len(r.Prefix)+r.Length > 128 {
			response.BadRequest(c, "数量需为 1-1000，长度需为 8-64")
			return
		}
		group, e := groups.GetByID(c.Request.Context(), a, r.GroupID)
		if dbError(c, e) {
			return
		}
		if group == nil {
			response.BadRequest(c, "卡密分组不存在")
			return
		}
		tx, e := d.BeginTxx(c.Request.Context(), nil)
		if dbError(c, e) {
			return
		}
		defer tx.Rollback()
		// Serialize generation with group deletion; never create orphaned cards.
		e = tx.GetContext(c.Request.Context(), group, repo.SQL("SELECT id,appid,name,type,value,remark FROM {kami_groups} WHERE appid=? AND id=? FOR UPDATE"), a, r.GroupID)
		if errors.Is(e, sql.ErrNoRows) {
			response.BadRequest(c, "卡密分组不存在")
			return
		}
		if dbError(c, e) {
			return
		}
		if group.Type != "vip" || group.Value <= 0 || group.Value > 315360000000 {
			response.BadRequest(c, "卡密分组类型或面值无效")
			return
		}
		cards := []string{}
		for i := 0; i < r.Count; i++ {
			card, err := randomCard(r.Length)
			if dbError(c, err) {
				return
			}
			card = r.Prefix + card
			_, e = tx.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {kamis} (appid,group_id,group_name,card,status,use_user,use_time,remark,create_time,creator,create_ip) VALUES (?,?,?,?,0,'',0,?,?,?,?)"), a, group.ID, group.Name, card, r.Remark, time.Now().Unix(), middleware.CurrentAdmin(c).Username, c.ClientIP())
			if dbError(c, e) {
				return
			}
			cards = append(cards, card)
		}
		if !dbError(c, tx.Commit()) {
			response.OK(c, gin.H{"cards": cards, "count": len(cards)})
		}
	})
	g.GET("/kami/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var v repo.Kami
		err := d.GetContext(c.Request.Context(), &v, repo.SQL(kamiSelect+" WHERE appid=? AND id=?"), a, id)
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, 404, "卡密不存在")
			return
		}
		if !dbError(c, err) {
			response.OK(c, v)
		}
	})
	g.PUT("/kami/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var r kamiUserUpdateReq
		if c.ShouldBindJSON(&r) != nil {
			response.BadRequest(c, "请求体格式错误")
			return
		}
		if err := validateAccountLimits(r.BindLimit, r.RegionLimit); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		v := map[string]any{}
		if r.Nickname != nil {
			v["nickname"] = *r.Nickname
		}
		if r.Vip != nil {
			v["vip"] = *r.Vip
		}
		if r.BanExpire != nil {
			v["ban_expire"] = *r.BanExpire
		}
		if r.BindLimit != nil {
			v["bind_limit"] = *r.BindLimit
		}
		if r.RegionLimit != nil {
			v["region_limit"] = *r.RegionLimit
		}
		if r.Mi != nil {
			v["mi"] = *r.Mi
		}
		if r.Remark != nil {
			v["remark"] = *r.Remark
		}
		err := repo.NewKamiRepo(d).UpdateAccount(c.Request.Context(), a, id, v, r.Frozen)
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, 404, "卡密不存在")
			return
		}
		if !dbError(c, err) {
			if ((r.Frozen != nil && *r.Frozen) || (r.BanExpire != nil && (*r.BanExpire < 0 || *r.BanExpire > time.Now().Unix()))) && dbError(c, clearUserPresence(c, cache, a, -id)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.DELETE("/kami/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		_, e := repo.NewKamiRepo(d).Delete(c.Request.Context(), a, id, false)
		if !dbError(c, e) {
			if dbError(c, clearUserPresence(c, cache, a, -id)) {
				return
			}
			response.OK(c, nil)
		}
	})
	g.POST("/kami/clean-used", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		ids, e := repo.NewKamiRepo(d).DeleteReturningIDs(c.Request.Context(), a, 0, true)
		if !dbError(c, e) {
			for i := range ids {
				ids[i] = -ids[i]
			}
			if dbError(c, clearUserPresence(c, cache, a, ids...)) {
				return
			}
			response.OK(c, gin.H{"affected": len(ids)})
		}
	})
	g.POST("/kami/:id/reward", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var r kamiRewardReq
		if c.ShouldBindJSON(&r) != nil || r.Seconds <= 0 || r.Seconds > 315360000 {
			response.BadRequest(c, "奖励秒数无效")
			return
		}
		_, e := d.ExecContext(c.Request.Context(), repo.SQL("UPDATE {kamis} SET vip=GREATEST(vip,?)+? WHERE appid=? AND id=?"), time.Now().Unix(), r.Seconds, a, id)
		if !dbError(c, e) {
			response.OK(c, nil)
		}
	})
	g.POST("/kami/:id/unbind-mi", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if ok && !dbError(c, repo.NewKamiRepo(d).UpdateBinding(c.Request.Context(), a, id, map[string]any{"mi": ""})) {
			if _, e := d.ExecContext(c.Request.Context(), repo.SQL("DELETE FROM {user_devices} WHERE uid=? AND EXISTS (SELECT 1 FROM {kamis} WHERE appid=? AND id=?)"), -id, a, id); dbError(c, e) {
				return
			}
			if dbError(c, clearUserPresence(c, cache, a, -id)) {
				return
			}
			response.OK(c, nil)
		}
	})
}
