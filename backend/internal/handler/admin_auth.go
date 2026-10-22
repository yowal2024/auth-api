package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func NewAdminAuthHandler(a *repo.AdminRepo, s *auth.SessionStore, cap *auth.CaptchaService) *AdminAuthHandler {
	return &AdminAuthHandler{a, s, cap}
}
func (h *AdminAuthHandler) Captcha(c *gin.Context) {
	v, e := h.captcha.Generate(c.Request.Context())
	if e != nil {
		response.ServerError(c, "验证码生成失败")
		return
	}
	response.OK(c, gin.H{"id": v.ID, "image": v.Image, "type": v.Type, "prompt": v.Prompt})
}
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var v loginReq
	if c.ShouldBindJSON(&v) != nil {
		response.BadRequest(c, "用户名或密码不能为空")
		return
	}
	ctx := c.Request.Context()
	ip := c.ClientIP()
	if h.captcha.NeedCaptcha(ctx, ip) && !h.captcha.Verify(ctx, v.CaptchaId, v.CaptchaCode) {
		response.FailData(c, 1001, "请输入正确的验证码", gin.H{"needCaptcha": true})
		return
	}
	a, e := h.admins.FindByUsername(ctx, v.Username)
	if e != nil {
		response.ServerError(c, "读取管理员失败")
		return
	}
	if a == nil || !auth.CheckPassword(a.Password, v.Password) {
		if e = h.captcha.RecordFailure(ctx, ip); e != nil {
			response.ServerError(c, "登录服务暂不可用")
			return
		}
		response.FailData(c, 1001, "用户名或密码错误", gin.H{"needCaptcha": h.captcha.NeedCaptcha(ctx, ip)})
		return
	}
	s := &auth.Session{AdminID: a.ID, Username: a.Username, Nickname: a.Nickname, Avatar: a.Avatar, Role: a.Role, Credential: auth.CredentialRevision(a.Password)}
	token, e := h.store.Create(ctx, s)
	if e != nil {
		log.Printf("admin session creation failed: %v", e)
		response.ServerError(c, "签发会话失败")
		return
	}
	_ = h.captcha.ResetFailure(ctx, ip)
	_ = h.admins.UpdateLastLogin(ctx, a.ID, time.Now().Unix(), ip)
	response.OK(c, loginResp{token, adminProfile(s)})
}
func adminProfile(a *auth.Session) map[string]any {
	return gin.H{"id": a.AdminID, "username": a.Username, "nickname": a.Nickname, "avatar": a.Avatar, "role": a.Role}
}
func (h *AdminAuthHandler) Logout(c *gin.Context) {
	if e := h.store.Delete(c.Request.Context(), c.GetString("adminToken")); e != nil {
		response.ServerError(c, "退出失败")
		return
	}
	response.OK(c, nil)
}
func (h *AdminAuthHandler) Me(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	if a == nil {
		response.Unauthorized(c)
		return
	}
	response.OK(c, adminProfile(a))
}
func NewAdminMeHandler(a *repo.AdminRepo, s *auth.SessionStore) *AdminMeHandler {
	return &AdminMeHandler{admins: a, store: s}
}
func (h *AdminMeHandler) Profile(c *gin.Context) {
	var v profileReq
	if c.ShouldBindJSON(&v) != nil || len(v.Username) < 5 || len(v.Username) > 12 {
		response.BadRequest(c, "账号长度需在 5-12 位")
		return
	}
	a := middleware.CurrentAdmin(c)
	if e := h.admins.UpdateProfile(c.Request.Context(), a.AdminID, v.Username, v.Nickname, v.Avatar); e != nil {
		response.BadRequest(c, "账号或昵称已存在，或更新失败")
		return
	}
	a.Username = v.Username
	a.Nickname = v.Nickname
	a.Avatar = v.Avatar
	if e := h.store.Update(c.Request.Context(), c.GetString("adminToken"), a); e != nil {
		response.ServerError(c, "更新会话失败")
		return
	}
	response.OK(c, adminProfile(a))
}
func (h *AdminMeHandler) Password(c *gin.Context) {
	var v passwordReq
	if c.ShouldBindJSON(&v) != nil || len(v.NewPassword) < 6 || len(v.NewPassword) > 18 {
		response.BadRequest(c, "密码长度需在 6-18 位")
		return
	}
	ctx := c.Request.Context()
	a, e := h.admins.GetByID(ctx, middleware.CurrentAdmin(c).AdminID)
	if e != nil || a == nil {
		response.ServerError(c, "读取管理员失败")
		return
	}
	if !auth.CheckPassword(a.Password, v.OldPassword) {
		response.BadRequest(c, "原密码错误")
		return
	}
	hash, e := auth.HashPassword(v.NewPassword)
	if e == nil {
		e = h.admins.UpdatePassword(ctx, a.ID, hash)
	}
	if e == nil {
		e = h.store.DeleteByAdmin(ctx, a.ID)
	}
	if e != nil {
		response.ServerError(c, "更新密码或撤销会话失败")
		return
	}
	response.OK(c, nil)
}
