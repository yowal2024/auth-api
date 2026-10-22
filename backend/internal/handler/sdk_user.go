package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/config"
	"auth-api/internal/geo"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type SDKService struct {
	db           *sqlx.DB
	rdb          *redis.Client
	sessions     *auth.UserSessionStore
	cloneConfig  config.CloneConfig
	uploadConfig config.UploadConfig
	LocalGeo     *geo.LocalDB
}

func NewSDKService(d *sqlx.DB, r *redis.Client) *SDKService {
	return &SDKService{db: d, rdb: r, sessions: auth.NewUserSessionStore(r, 24*time.Hour)}
}
func (s *SDKService) Mount(g *gin.RouterGroup) {
	g.POST("/logon", s.Logon)
	g.POST("/register", s.Register)
	g.POST("/ini", s.Ini)
	protected := g.Group("")
	protected.Use(middleware.SDKToken(s.sessions), s.authorize)
	protected.POST("/info", s.Info)
	protected.POST("/heartbeat", s.Heartbeat)
	protected.POST("/logout", s.Logout)
	protected.POST("/updateInfo", s.UpdateInfo)
	protected.POST("/userExtends", s.UserExtends)
	protected.POST("/kamiTopup", s.KamiTopup)
	protected.POST("/cloudFunction", s.CloudFunction)
}

func (s *SDKService) blocked(c *gin.Context, mi string) bool {
	a := middleware.CurrentApp(c)
	var code int
	e := s.db.GetContext(c.Request.Context(), &code, repo.SQL("SELECT COALESCE(MIN(CASE WHEN type='ip' THEN 1003 ELSE 1021 END),0) FROM {blocks} WHERE appid=? AND ((type='ip' AND value=?) OR (type='mi' AND value=?))"), a.ID, c.ClientIP(), mi)
	if dbError(c, e) {
		return true
	}
	if code > 0 {
		response.Fail(c, code, "IP 或机器码已被封禁")
		return true
	}
	return false
}
func (s *SDKService) Logon(c *gin.Context) {
	a := middleware.CurrentApp(c)
	if a.Type == "kami" {
		s.kamiLogon(c)
		return
	}
	var req sdkLogonReq
	if c.ShouldBindJSON(&req) != nil || req.Username == "" || req.Password == "" {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if s.blocked(c, req.Mi) {
		return
	}
	var id int64
	e := s.db.GetContext(c.Request.Context(), &id, repo.SQL("SELECT id FROM {users} WHERE appid=? AND username=? LIMIT 1"), a.ID, req.Username)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		dbError(c, e)
		return
	}
	// Use the matched database identity, not spelling: MySQL may consider case,
	// accents and trailing spaces equivalent. Unknown accounts share an IP limit.
	failKey := fmt.Sprintf("recovery:brute:uid:%d:%d", a.ID, id)
	if errors.Is(e, sql.ErrNoRows) {
		failKey = fmt.Sprintf("recovery:brute:unknown:%d:%x", a.ID, sha256.Sum256([]byte(c.ClientIP())))
	}
	c.Set("security.subject", fmt.Sprint(id))
	if !middleware.Limit(c, s.rdb, "login-account", failKey, 30, time.Minute) {
		return
	}
	if a.BruteForce && !s.allowLoginAttempt(c, failKey) {
		return
	}
	var u *repo.User
	if e == nil {
		u, e = repo.NewUserRepo(s.db).GetByID(c.Request.Context(), a.ID, id)
		if dbError(c, e) {
			return
		}
	}
	if u == nil || !auth.CheckPassword(u.Password, req.Password) {
		response.Fail(c, 1001, "账号或密码错误")
		return
	}
	if !s.runTrigger(c, "before_login", req, u) {
		return
	}
	if !s.checkUser(c, u, true) || !s.checkRegion(c, u, false) {
		return
	}
	if a.BindDevice {
		if req.Mi == "" {
			response.Fail(c, 1011, "缺少机器码 mi,请检查 SDK 配置")
			return
		}
		if e = s.bindDevice(c.Request.Context(), u, req.Mi, c.ClientIP()); e != nil {
			response.Fail(c, 1008, e.Error())
			return
		}
	}
	if dbError(c, repo.NewUserRepo(s.db).UpdatePartial(c.Request.Context(), a.ID, u.ID, map[string]any{"mi": req.Mi})) {
		return
	}
	_, e = s.db.ExecContext(c.Request.Context(), repo.SQL("UPDATE {users} SET last_ip=?,last_region=?,last_login_time=?,last_active_time=UNIX_TIMESTAMP() WHERE appid=? AND id=?"), c.ClientIP(), u.Region, time.Now().Unix(), a.ID, u.ID)
	if dbError(c, e) {
		return
	}
	u.Mi = req.Mi
	if a.RandomPackage && !u.HasSignKey {
		c.Set("sdkUserRecord", u)
		v := VersionService{DB: s.db}
		if _, e = v.signKey(c, false); dbError(c, e) {
			return
		}
		u.HasSignKey = true
	}
	if !s.runTrigger(c, "after_login", req, u) {
		return
	}
	token, e := s.createUserSession(c, &auth.UserSession{Credential: auth.CredentialRevision(u.Password), UID: u.ID, AppID: a.ID, Username: u.Username, Mi: req.Mi, LastIP: c.ClientIP()}, time.Duration(a.UserTokenTTL)*time.Second)
	if dbError(c, e) {
		return
	}
	c.Set("sdkLogRegion", u.Region)
	s.writeUserLog(c, u.ID, "login")
	if !s.touchDevice(c, u.ID, req.Mi, u.Region) {
		return
	}
	if a.BruteForce {
		_ = s.rdb.Del(c.Request.Context(), failKey).Err()
	}
	s.online(c, u.ID)
	response.OK(c, gin.H{"token": token, "user": sdkUserProfile(u), "serverTime": time.Now().Unix()})
}
func (s *SDKService) checkUser(c *gin.Context, u *repo.User, enforceVIP bool) bool {
	now := time.Now().Unix()
	if u.BanExpire < 0 || u.BanExpire > now {
		response.Fail(c, 1002, "账号已被封禁")
		return false
	}
	if enforceVIP && middleware.CurrentApp(c).Mode == "vip" && u.Vip <= now {
		response.Fail(c, 1018, "VIP 已到期,请续费后再使用")
		return false
	}
	return true
}
func (s *SDKService) bindDevice(ctx context.Context, u *repo.User, mi, ip string) error {
	tx, e := s.db.BeginTxx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var limit int
	if e = tx.GetContext(ctx, &limit, repo.SQL("SELECT bind_limit FROM {users} WHERE appid=? AND id=? FOR UPDATE"), u.AppID, u.ID); e != nil {
		return e
	}
	var id int64
	e = tx.GetContext(ctx, &id, repo.SQL("SELECT id FROM {user_devices} WHERE uid=? AND mi=?"), u.ID, mi)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return e
	}
	if errors.Is(e, sql.ErrNoRows) {
		var n int
		if e = tx.GetContext(ctx, &n, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid=?"), u.ID); e != nil {
			return e
		}
		if n >= 1+limit {
			return errors.New("设备数达上限,请先在后台清理或调整上限")
		}
		_, e = tx.ExecContext(ctx, repo.SQL("INSERT INTO {user_devices} (uid,mi,device,os,ip,region,bind_time,last_time) VALUES (?,?,'','',?,'',?,?)"), u.ID, mi, ip, time.Now().Unix(), time.Now().Unix())
	} else {
		_, e = tx.ExecContext(ctx, repo.SQL("UPDATE {user_devices} SET last_time=?,ip=? WHERE uid=? AND id=?"), time.Now().Unix(), ip, u.ID, id)
	}
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (s *SDKService) Register(c *gin.Context) {
	a := middleware.CurrentApp(c)
	if a.Type == "kami" {
		response.BadRequest(c, "卡密版应用不支持注册,请使用卡密登录")
		return
	}
	var r sdkRegisterReq
	if c.ShouldBindJSON(&r) != nil || len(r.Username) < 5 || len(r.Username) > 12 || len(r.Password) < 6 || len(r.Password) > 18 {
		response.BadRequest(c, "账号或密码长度无效")
		return
	}
	if a.BindDevice && r.Mi == "" {
		response.Fail(c, 1011, "缺少机器码 mi,请检查 SDK 配置")
		return
	}
	if s.blocked(c, r.Mi) {
		return
	}
	if !s.runTrigger(c, "before_register", r, nil) {
		return
	}
	hash, e := auth.HashPassword(r.Password)
	if dbError(c, e) {
		return
	}
	u := &repo.User{AppID: a.ID, Username: r.Username, Password: hash, Nickname: r.Nickname, Mi: r.Mi, IPList: "[]", RegisterIP: c.ClientIP(), CreateTime: time.Now().Unix()}
	u.ID, e = repo.NewUserRepo(s.db).CreateWith(c.Request.Context(), u, func(tx *sqlx.Tx) error {
		if !s.checkRegionTx(c, u, false, tx) {
			return errors.New("region check failed")
		}
		// Registration returns a logged-in session, so it is also account activity.
		if _, err := tx.ExecContext(c.Request.Context(), repo.SQL("UPDATE {users} SET last_ip=?,last_region=?,last_active_time=? WHERE appid=? AND id=?"), c.ClientIP(), u.Region, time.Now().Unix(), a.ID, u.ID); err != nil {
			return err
		}
		if a.BindDevice {
			_, err := tx.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {user_devices} (uid,mi,device,os,ip,region,bind_time,last_time) VALUES (?,?,'','',?,?,?,?)"), u.ID, r.Mi, c.ClientIP(), u.Region, time.Now().Unix(), time.Now().Unix())
			return err
		}
		return nil
	})
	if e != nil && c.Writer.Written() {
		return
	}
	if errors.Is(e, repo.ErrUsernameExists) {
		response.Fail(c, 1007, e.Error())
		return
	}
	if dbError(c, e) {
		return
	}
	if !s.runTrigger(c, "after_register", r, u) {
		return
	}
	token, e := s.createUserSession(c, &auth.UserSession{Credential: auth.CredentialRevision(u.Password), UID: u.ID, AppID: a.ID, Username: u.Username, Mi: r.Mi, LastIP: c.ClientIP()}, time.Duration(a.UserTokenTTL)*time.Second)
	if !dbError(c, e) {
		c.Set("sdkLogRegion", u.Region)
		s.writeUserLog(c, u.ID, "register")
		s.online(c, u.ID)
		response.OK(c, gin.H{"token": token, "user": sdkUserProfile(u)})
	}
}
func sdkUserProfile(u *repo.User) gin.H {
	return gin.H{"id": u.ID, "username": u.Username, "nickname": u.Nickname, "avatar": u.Avatar, "vip": u.Vip, "vipActive": u.Vip > time.Now().Unix(), "mi": u.Mi, "region": u.Region, "banExpire": u.BanExpire, "bindLimit": u.BindLimit, "groupId": u.GroupID, "groupName": u.GroupName, "groupLevel": u.GroupLevel, "hasSignKey": u.HasSignKey}
}
func (s *SDKService) online(c *gin.Context, id int64) {
	token := c.GetString("sdkToken")
	if issued := c.GetString("presenceToken"); issued != "" {
		token = issued
	}
	if token != "" {
		if err := s.sessions.Touch(c.Request.Context(), token, c.ClientIP(), false); err != nil {
			c.Error(err)
		}
	}
	if err := s.sessionActivity(c, id, "touch"); err != nil {
		c.Error(err)
	}
}
func (s *SDKService) authorize(c *gin.Context) {
	session := middleware.CurrentSDKUser(c)
	if session.IsKami != (middleware.CurrentApp(c).Type == "kami") {
		response.Fail(c, 1006, "应用账号类型已变更，请重新登录")
		c.Abort()
		return
	}
	if session.IsKami {
		s.authorizeKami(c)
		return
	}
	u, e := repo.NewUserRepo(s.db).GetByID(c.Request.Context(), session.AppID, session.UID)
	if dbError(c, e) {
		c.Abort()
		return
	}
	if u == nil || session.Credential != auth.CredentialRevision(u.Password) {
		response.Fail(c, 1006, "未登录或登录已过期")
		c.Abort()
		return
	}
	vipRequired := !stringsHasSuffix(c.FullPath(), "/kamiTopup") && !stringsHasSuffix(c.FullPath(), "/logout")
	if !s.checkUser(c, u, vipRequired) || s.blocked(c, session.Mi) {
		c.Abort()
		return
	}
	if middleware.CurrentApp(c).BindDevice {
		mi := c.GetHeader("X-Mi")
		if mi == "" || mi != session.Mi {
			response.Fail(c, 1006, "设备不匹配,请重新登录")
			c.Abort()
			return
		}
		var n int
		e = s.db.GetContext(c.Request.Context(), &n, repo.SQL("SELECT COUNT(*) FROM {user_devices} WHERE uid=? AND mi=? AND id=?"), u.ID, mi, session.DeviceID)
		if dbError(c, e) {
			c.Abort()
			return
		}
		if n == 0 {
			_ = s.sessions.Delete(c.Request.Context(), c.GetString("sdkToken"))
			response.Fail(c, 1006, "设备已解绑,请重新登录")
			c.Abort()
			return
		}
	}
	if !s.checkRegion(c, u, false) {
		c.Abort()
		return
	}
	if !s.touchDevice(c, u.ID, session.Mi, u.Region) {
		c.Abort()
		return
	}
	c.Set("sdkUserRecord", u)
	s.online(c, u.ID)
	c.Next()
}
func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
func currentUser(c *gin.Context) *repo.User {
	v, _ := c.Get("sdkUserRecord")
	u, _ := v.(*repo.User)
	return u
}
func (s *SDKService) Info(c *gin.Context) {
	profile := sdkUserProfile(currentUser(c))
	if c.GetBool("sdkIsKami") {
		profile["isKami"] = true
		profile["card"] = currentUser(c).Username
		profile["regionLimit"] = currentUser(c).RegionLimit
	}
	response.OK(c, gin.H{"user": profile, "serverTime": time.Now().Unix()})
}
func (s *SDKService) Heartbeat(c *gin.Context) {
	u := currentUser(c)
	table := "users"
	if c.GetBool("sdkIsKami") {
		table = "kamis"
	}
	if dbError(c, s.sessions.Touch(c.Request.Context(), c.GetString("sdkToken"), c.ClientIP(), true)) {
		return
	}
	_, err := s.db.ExecContext(c.Request.Context(), repo.SQL("UPDATE {"+table+"} SET last_active_time=GREATEST(last_active_time,?),last_ip=? WHERE appid=? AND id=?"), time.Now().Unix(), c.ClientIP(), u.AppID, u.ID)
	if dbError(c, err) {
		return
	}
	response.OK(c, gin.H{"ok": true, "serverTime": time.Now().Unix()})
}
func (s *SDKService) Logout(c *gin.Context) {
	if dbError(c, s.sessions.Delete(c.Request.Context(), c.GetString("sdkToken"))) {
		return
	}
	id := currentUser(c).ID
	if c.GetBool("sdkIsKami") {
		id = -id
	}
	if dbError(c, s.sessionActivity(c, currentUser(c).ID, "logout")) {
		return
	}
	s.writeUserLog(c, id, "logout")
	response.OK(c, gin.H{"ok": true})
}
func (s *SDKService) UpdateInfo(c *gin.Context) {
	var r sdkUpdateInfoReq
	if c.ShouldBindJSON(&r) != nil || (r.Nickname == nil && r.Avatar == nil) {
		response.BadRequest(c, "未提供要修改的字段")
		return
	}
	v := map[string]any{}
	u := currentUser(c)
	if c.GetBool("sdkIsKami") && r.Avatar != nil {
		response.BadRequest(c, "卡密版账号不支持头像")
		return
	}
	if r.Nickname != nil {
		if len([]rune(*r.Nickname)) > 32 {
			response.BadRequest(c, "昵称过长(最多 32 字)")
			return
		}
		v["nickname"] = *r.Nickname
		u.Nickname = *r.Nickname
	}
	if r.Avatar != nil {
		if len(*r.Avatar) > 255 {
			response.BadRequest(c, "头像地址过长")
			return
		}
		v["avatar"] = *r.Avatar
		u.Avatar = *r.Avatar
	}
	var err error
	if c.GetBool("sdkIsKami") {
		delete(v, "avatar")
		err = repo.NewKamiRepo(s.db).UpdateBinding(c.Request.Context(), u.AppID, u.ID, v)
	} else {
		err = repo.NewUserRepo(s.db).UpdatePartial(c.Request.Context(), u.AppID, u.ID, v)
	}
	if !dbError(c, err) {
		s.Info(c)
	}
}
func (s *SDKService) UserExtends(c *gin.Context) {
	copyUser := *currentUser(c)
	u := &copyUser
	table := "users"
	if c.GetBool("sdkIsKami") {
		table = "kamis"
	}
	recordID := u.ID
	if c.GetBool("sdkIsKami") {
		u.ID = -u.ID
	}
	var r sdkUserExtendsReq
	if c.ShouldBindJSON(&r) != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	ctx := c.Request.Context()
	tx, e := s.db.BeginTxx(ctx, nil)
	if dbError(c, e) {
		return
	}
	defer tx.Rollback()
	var id int64
	if dbError(c, tx.GetContext(ctx, &id, repo.SQL("SELECT id FROM {"+table+"} WHERE appid=? AND id=? FOR UPDATE"), u.AppID, recordID)) {
		return
	}
	for k, v := range r.Set {
		if strings.TrimSpace(k) == "" || len(k) > 64 {
			response.BadRequest(c, "变量名无效")
			return
		}
		var eid int64
		e = tx.GetContext(ctx, &eid, repo.SQL("SELECT id FROM {user_extends} WHERE uid=? AND e_key=? ORDER BY id LIMIT 1"), u.ID, k)
		if errors.Is(e, sql.ErrNoRows) {
			_, e = tx.ExecContext(ctx, repo.SQL("INSERT INTO {user_extends} (uid,e_key,e_value) VALUES (?,?,?)"), u.ID, k, v)
		} else if e == nil {
			// Historical duplicate keys must not shadow a successful SDK write.
			_, e = tx.ExecContext(ctx, repo.SQL("UPDATE {user_extends} SET e_value=? WHERE uid=? AND e_key=?"), v, u.ID, k)
		}
		if dbError(c, e) {
			return
		}
	}
	rows := []repo.UserExtend{}
	if dbError(c, tx.SelectContext(ctx, &rows, repo.SQL("SELECT id,uid,e_key,e_value FROM {user_extends} WHERE uid=? ORDER BY id"), u.ID)) {
		return
	}
	if dbError(c, tx.Commit()) {
		return
	}
	values := map[string]string{}
	for _, v := range rows {
		values[v.Key] = v.Value
	}
	response.OK(c, gin.H{"extends": values})
}
func (s *SDKService) KamiTopup(c *gin.Context) {
	var r sdkKamiReq
	if c.ShouldBindJSON(&r) != nil || r.Card == "" {
		response.BadRequest(c, "请求参数错误")
		return
	}
	u := currentUser(c)
	if !s.cardAttempt(c, r.Card, fmt.Sprintf("%t:%d", c.GetBool("sdkIsKami"), u.ID)) {
		return
	}
	table := "users"
	if c.GetBool("sdkIsKami") {
		table = "kamis"
	}
	ctx := c.Request.Context()
	tx, e := s.db.BeginTxx(ctx, nil)
	if dbError(c, e) {
		return
	}
	defer tx.Rollback()
	var card repo.Kami
	e = tx.GetContext(ctx, &card, repo.SQL(kamiSelect+" WHERE appid=? AND card=? FOR UPDATE"), u.AppID, r.Card)
	if errors.Is(e, sql.ErrNoRows) {
		response.Fail(c, 2001, "卡密无效或不可用")
		return
	}
	if dbError(c, e) {
		return
	}
	if card.Status != 0 {
		response.Fail(c, 2002, "卡密无效或不可用")
		return
	}
	var group repo.KamiGroup
	if dbError(c, tx.GetContext(ctx, &group, repo.SQL("SELECT id,appid,name,type,value,remark FROM {kami_groups} WHERE appid=? AND id=?"), u.AppID, card.GroupID)) {
		return
	}
	if group.Type != "vip" || group.Value <= 0 || group.Value > 315360000000 {
		response.BadRequest(c, "卡密类型或面值不支持")
		return
	}
	now := time.Now().Unix()
	res, e := tx.ExecContext(ctx, repo.SQL("UPDATE {kamis} SET status=1,use_user=?,use_time=? WHERE appid=? AND id=? AND status=0"), u.Username, now, u.AppID, card.ID)
	if dbError(c, e) {
		return
	}
	n, e := res.RowsAffected()
	if dbError(c, e) {
		return
	}
	if n != 1 {
		response.Fail(c, 2002, "卡密已被使用")
		return
	}
	if _, e = tx.ExecContext(ctx, repo.SQL("UPDATE {"+table+"} SET vip=GREATEST(vip,?)+? WHERE appid=? AND id=?"), now, group.Value, u.AppID, u.ID); dbError(c, e) {
		return
	}
	var vip int64
	if dbError(c, tx.GetContext(ctx, &vip, repo.SQL("SELECT vip FROM {"+table+"} WHERE appid=? AND id=?"), u.AppID, u.ID)) {
		return
	}
	if !dbError(c, tx.Commit()) {
		s.cardSuccess(c)
		u.Vip = vip
		logID := u.ID
		if c.GetBool("sdkIsKami") {
			logID = -logID
		}
		s.writeUserLog(c, logID, "kamiTopup")
		_ = s.runTrigger(c, "kami_use", gin.H{"card": r.Card, "vip": vip, "groupName": group.Name, "addSec": group.Value}, u)
		response.OK(c, gin.H{"vip": vip, "vipActive": vip > now, "groupName": group.Name, "addSec": group.Value})
	}
}
func (s *SDKService) Ini(c *gin.Context) {
	a := middleware.CurrentApp(c)
	notices, e := repo.NewNoticeRepo(s.db).ListByApp(c.Request.Context(), a.ID)
	if dbError(c, e) {
		return
	}
	items := []gin.H{}
	for _, n := range notices {
		if n.Enabled == 1 {
			items = append(items, gin.H{"id": n.ID, "title": n.Title, "content": n.Content, "author": n.Author, "updateTime": n.UpdateTime})
		}
	}
	exts, e := repo.NewExtendRepo(s.db).ListByApp(c.Request.Context(), a.ID)
	if dbError(c, e) {
		return
	}
	values := map[string]string{}
	for _, v := range exts {
		values[v.Key] = v.Value
	}
	latest, ok := s.latestVersion(c)
	if !ok {
		return
	}
	response.OK(c, gin.H{"notices": items, "extends": values, "latestVersion": latest})
}

func (s *SDKService) SetCloneConfig(c config.CloneConfig)   { s.cloneConfig = c }
func (s *SDKService) SetUploadConfig(c config.UploadConfig) { s.uploadConfig = c }
