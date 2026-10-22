package handler

import (
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func MountContent(g *gin.RouterGroup, d *sqlx.DB, clients ...*redis.Client) {
	var cache *redis.Client
	if len(clients) > 0 {
		cache = clients[0]
	}
	notices := repo.NewNoticeRepo(d)
	extends := repo.NewExtendRepo(d)
	blocks := repo.NewBlockRepo(d)
	ug := repo.NewUserGroupRepo(d)
	kg := repo.NewKamiGroupRepo(d)
	g.GET("/notice/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := notices.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/notice", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r noticeSaveReq
		if c.ShouldBindJSON(&r) != nil || r.Title == "" {
			response.BadRequest(c, "标题不能为空")
			return
		}
		v := repo.Notice{ID: r.ID, AppID: a, Title: r.Title, Content: r.Content, Enabled: 1, AdminID: middleware.CurrentAdmin(c).AdminID, Author: middleware.CurrentAdmin(c).Username, UpdateTime: time.Now().Unix()}
		if r.Enabled != nil {
			v.Enabled = *r.Enabled
		}
		var e error
		if r.ID > 0 {
			old, err := notices.GetByID(c.Request.Context(), a, r.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "公告不存在")
				return
			}
			e = notices.Update(c.Request.Context(), &v)
		} else {
			v.ID, e = notices.Create(c.Request.Context(), &v)
		}
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.DELETE("/notice/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if ok && !dbError(c, notices.Delete(c.Request.Context(), a, id)) {
			response.OK(c, nil)
		}
	})
	g.GET("/extend/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := extends.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			keyword := c.Query("keyword")
			if keyword != "" {
				filtered := []repo.Extend{}
				for _, x := range v {
					if strings.Contains(x.Key, keyword) || strings.Contains(x.Value, keyword) || strings.Contains(x.Remark, keyword) {
						filtered = append(filtered, x)
					}
				}
				v = filtered
			}
			response.OK(c, v)
		}
	})
	g.POST("/extend", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r extendSaveReq
		if c.ShouldBindJSON(&r) != nil || r.Key == "" {
			response.BadRequest(c, "变量名不能为空")
			return
		}
		ctx := c.Request.Context()
		var n int
		if dbError(c, d.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {extends} WHERE appid=? AND e_key=? AND id<>?"), a, r.Key, r.ID)) {
			return
		}
		if n > 0 {
			response.BadRequest(c, "变量名已存在")
			return
		}
		v := repo.Extend{ID: r.ID, AppID: a, Key: r.Key, Value: r.Value, Remark: r.Remark, UpdateTime: time.Now().Unix()}
		var e error
		if r.ID > 0 {
			old, err := extends.GetByID(ctx, a, r.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "扩展项不存在")
				return
			}
			e = extends.Update(ctx, &v)
		} else {
			v.ID, e = extends.Create(ctx, &v)
		}
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.DELETE("/extend/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if ok && !dbError(c, extends.Delete(c.Request.Context(), a, id)) {
			response.OK(c, nil)
		}
	})
	g.GET("/block/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := blocks.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/block", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r blockAddReq
		if c.ShouldBindJSON(&r) != nil || r.Value == "" || (r.Type != "ip" && r.Type != "mi" && r.Type != "region") {
			response.BadRequest(c, "封禁类型或值无效")
			return
		}
		v := repo.Block{AppID: a, Type: r.Type, Value: r.Value, Reason: r.Reason, CreateTime: time.Now().Unix()}
		id, e := blocks.Create(c.Request.Context(), &v)
		if !dbError(c, e) {
			v.ID = id
			response.OK(c, v)
		}
	})
	g.DELETE("/block/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if ok && !dbError(c, blocks.Delete(c.Request.Context(), a, id)) {
			response.OK(c, nil)
		}
	})
	g.GET("/user/group/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := ug.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/user/group", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r userGroupReq
		if c.ShouldBindJSON(&r) != nil || r.Name == "" {
			response.BadRequest(c, "分组名称不能为空")
			return
		}
		v := repo.UserGroup{ID: r.ID, AppID: a, Name: r.Name, Level: r.Level, Remark: r.Remark, CreateTime: time.Now().Unix()}
		var e error
		if r.ID > 0 {
			old, err := ug.GetByID(c.Request.Context(), a, r.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "分组不存在")
				return
			}
			e = ug.Update(c.Request.Context(), &v)
		} else {
			v.ID, e = ug.Create(c.Request.Context(), &v)
		}
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.DELETE("/user/group/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		n, e := ug.CountUsers(c.Request.Context(), a, id)
		if dbError(c, e) {
			return
		}
		if n > 0 {
			response.BadRequest(c, "分组下仍有用户")
			return
		}
		if !dbError(c, ug.Delete(c.Request.Context(), a, id)) {
			response.OK(c, nil)
		}
	})
	g.GET("/kami/group/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := kg.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/kami/group", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r kamiGroupReq
		if c.ShouldBindJSON(&r) != nil || strings.TrimSpace(r.Name) == "" || r.Type != "vip" || r.Value <= 0 || r.Value > 315360000000 {
			response.BadRequest(c, "分组参数无效")
			return
		}
		v := repo.KamiGroup{ID: r.ID, AppID: a, Name: r.Name, Type: r.Type, Value: r.Value, Remark: r.Remark}
		var e error
		if r.ID > 0 {
			old, err := kg.GetByID(c.Request.Context(), a, r.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "分组不存在")
				return
			}
			e = kg.Update(c.Request.Context(), &v)
		} else {
			v.ID, e = kg.Create(c.Request.Context(), &v)
		}
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.DELETE("/kami/group/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		ids, err := kg.DeleteReturningIDs(c.Request.Context(), a, id)
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, 404, "分组不存在")
			return
		}
		if !dbError(c, err) {
			if dbError(c, clearUserPresence(c, cache, a, ids...)) {
				return
			}
			response.OK(c, nil)
		}
	})
}
