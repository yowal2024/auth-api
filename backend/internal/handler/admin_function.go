package handler

import (
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"auth-api/internal/sandbox"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

var triggers = map[string]bool{"manual": true, "before_login": true, "after_login": true, "before_register": true, "after_register": true, "kami_use": true}

func MountFunctions(g *gin.RouterGroup, d *sqlx.DB) {
	// Script authors can query the database and initiate outbound requests.
	// Treat code management as a server-level privilege, including dry runs.
	g = g.Group("", middleware.RequireSuper())
	r := repo.NewFunctionRepo(d)
	apps := repo.NewAppRepo(d)
	g.GET("/function/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := r.ListByApp(c.Request.Context(), a)
		if dbError(c, e) {
			return
		}
		out := []repo.Function{}
		for _, f := range v {
			if (c.Query("trigger") == "" || f.Trigger == c.Query("trigger")) && (c.Query("keyword") == "" || strings.Contains(f.Name, c.Query("keyword"))) {
				out = append(out, f)
			}
		}
		response.OK(c, out)
	})
	g.GET("/function/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		v, e := r.GetByID(c.Request.Context(), a, id)
		if dbError(c, e) {
			return
		}
		if v == nil {
			response.Fail(c, 404, "云函数不存在")
			return
		}
		response.OK(c, v)
	})
	g.POST("/function", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var v fnReq
		if c.ShouldBindJSON(&v) != nil || v.Name == "" || !triggers[v.Trigger] || len(v.Code) > 1<<20 {
			response.BadRequest(c, "云函数参数无效")
			return
		}
		enabled := 0
		if v.Enabled == true || v.Enabled == float64(1) {
			enabled = 1
		}
		if v.Timeout <= 0 {
			v.Timeout = 3000
		}
		if v.Timeout > 30000 {
			response.BadRequest(c, "超时上限为 30000ms")
			return
		}
		item := repo.Function{ID: v.ID, AppID: a, Name: v.Name, Trigger: v.Trigger, Code: v.Code, Remark: v.Remark, Timeout: v.Timeout, Enabled: enabled}
		var e error
		if v.ID > 0 {
			old, err := r.GetByID(c.Request.Context(), a, v.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "云函数不存在")
				return
			}
			e = r.Update(c.Request.Context(), &item)
		} else {
			item.ID, e = r.Create(c.Request.Context(), &item)
		}
		if !dbError(c, e) {
			response.OK(c, item)
		}
	})
	g.DELETE("/function/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if ok && !dbError(c, r.Delete(c.Request.Context(), a, id)) {
			response.OK(c, nil)
		}
	})
	g.POST("/function/:id/run", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, e := strconv.ParseInt(c.Param("id"), 10, 64)
		if e != nil || id < 0 {
			response.BadRequest(c, "无效的 ID")
			return
		}
		var v fnRunReq
		if c.ShouldBindJSON(&v) != nil {
			response.BadRequest(c, "请求参数错误")
			return
		}
		if id > 0 {
			f, err := r.GetByID(c.Request.Context(), a, id)
			if dbError(c, err) {
				return
			}
			if f == nil {
				response.Fail(c, 404, "云函数不存在")
				return
			}
			if v.Code == "" {
				v.Code = f.Code
			}
			if v.Timeout <= 0 {
				v.Timeout = f.Timeout
			}
		}
		app, e := apps.GetByID(c.Request.Context(), a)
		if dbError(c, e) {
			return
		}
		out := sandbox.Run(c.Request.Context(), sandbox.RunOptions{Code: v.Code, Timeout: time.Duration(v.Timeout) * time.Millisecond, Payload: v.Payload, App: app, DB: d, Operator: middleware.CurrentAdmin(c).Username, Trigger: "manual", IP: c.ClientIP()})
		if id > 0 {
			_, e = d.ExecContext(c.Request.Context(), repo.SQL("UPDATE {functions} SET last_run=? WHERE appid=? AND id=?"), time.Now().Unix(), a, id)
			if dbError(c, e) {
				return
			}
		}
		c.Set("audit.executionFailed", !out.OK)
		response.OK(c, out)
	})
}
func (s *SDKService) runTrigger(c *gin.Context, name string, payload, user any) bool {
	a := middleware.CurrentApp(c)
	items := []repo.Function{}
	e := s.db.SelectContext(c.Request.Context(), &items, repo.SQL("SELECT id,appid,name,`trigger`,code,remark,timeout,enabled,last_run FROM {functions} WHERE appid=? AND `trigger`=? AND enabled=1 ORDER BY id"), a.ID, name)
	if e != nil && !strings.HasPrefix(name, "before_") {
		_ = c.Error(e)
		return true
	}
	if dbError(c, e) {
		return false
	}
	for _, f := range items {
		r := sandbox.Run(c.Request.Context(), sandbox.RunOptions{Code: f.Code, Timeout: time.Duration(f.Timeout) * time.Millisecond, Payload: payload, User: user, App: a, DB: s.db, Trigger: name, IP: c.ClientIP()})
		_, updateErr := s.db.ExecContext(c.Request.Context(), repo.SQL("UPDATE {functions} SET last_run=? WHERE appid=? AND id=?"), time.Now().Unix(), a.ID, f.ID)
		if updateErr != nil {
			_ = c.Error(updateErr)
		}
		// Original runTrigger calls evalReject only for before_ triggers.
		if !strings.HasPrefix(name, "before_") {
			if !r.OK {
				_ = c.Error(fmt.Errorf("trigger %s: %s", name, r.Error))
			}
			continue
		}
		if !r.OK {
			response.ServerError(c, "云函数执行失败: "+r.Error)
			return false
		}
		if v, ok := r.Result.(map[string]any); ok {
			if rejected, ok := v["allow"].(bool); ok && !rejected {
				response.Fail(c, 403, fmt.Sprint(v["msg"]))
				return false
			}
			if code, ok := v["code"]; ok && fmt.Sprint(code) != "0" {
				n, _ := strconv.Atoi(fmt.Sprint(code))
				if n == 0 {
					n = 403
				}
				response.Fail(c, n, fmt.Sprint(v["msg"]))
				return false
			}
		}
	}
	return true
}
func (s *SDKService) CloudFunction(c *gin.Context) {
	var v sdkCloudFunctionReq
	if c.ShouldBindJSON(&v) != nil || v.Name == "" {
		response.BadRequest(c, "函数名不能为空")
		return
	}
	var f repo.Function
	e := s.db.GetContext(c.Request.Context(), &f, repo.SQL("SELECT id,appid,name,`trigger`,code,remark,timeout,enabled,last_run FROM {functions} WHERE appid=? AND name=? AND `trigger`='manual' AND enabled=1 ORDER BY id LIMIT 1"), middleware.CurrentApp(c).ID, v.Name)
	if errors.Is(e, sql.ErrNoRows) {
		response.Fail(c, 404, "云函数不存在")
		return
	}
	if dbError(c, e) {
		return
	}
	u := currentUser(c)
	user := gin.H{"id": u.ID, "username": u.Username, "mi": u.Mi, "region": u.Region, "ip": c.ClientIP(), "isKami": c.GetBool("sdkIsKami"), "appid": u.AppID}
	out := sandbox.Run(c.Request.Context(), sandbox.RunOptions{Code: f.Code, Timeout: time.Duration(f.Timeout) * time.Millisecond, Payload: v.Param, User: user, App: middleware.CurrentApp(c), DB: s.db, Trigger: "manual", IP: c.ClientIP()})
	if _, err := s.db.ExecContext(c.Request.Context(), repo.SQL("UPDATE {functions} SET last_run=? WHERE appid=? AND id=?"), time.Now().Unix(), f.AppID, f.ID); err != nil {
		c.Error(err)
	}
	if !out.OK {
		if strings.HasPrefix(out.Error, "业务错误:") {
			response.Fail(c, 1099, strings.TrimSpace(strings.TrimPrefix(out.Error, "业务错误:")))
		} else {
			response.ServerError(c, "执行失败")
		}
		return
	}
	response.OK(c, out.Result)
}
