package server

import (
	"auth-api/internal/auth"
	"auth-api/internal/backup"
	"auth-api/internal/buildpool"
	"auth-api/internal/config"
	"auth-api/internal/geo"
	"auth-api/internal/handler"
	"auth-api/internal/install"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync/atomic"
	"time"
)

type Swappable struct{ current atomic.Pointer[http.Handler] }

func (s *Swappable) Store(h http.Handler) { s.current.Store(&h) }
func (s *Swappable) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := s.current.Load()
	if h == nil {
		http.Error(w, "starting", 503)
		return
	}
	(*h).ServeHTTP(w, r)
}

type Options struct {
	Backup         *backup.Manager
	Config         *config.Config
	DB             *sqlx.DB
	Redis          *redis.Client
	Installer      *install.Handler
	Web            fs.FS
	OnChange       func() error
	OnTLS          func() error
	CanonicalAdmin bool
	LocalGeo       *geo.LocalDB
}

func New(o Options) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Recovery())
	_ = g.SetTrustedProxies(nil)
	g.Use(func(c *gin.Context) {
		limit := int64(2 << 20)
		if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			limit = 210 << 20
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		defer func() {
			if c.Request.MultipartForm != nil {
				_ = c.Request.MultipartForm.RemoveAll()
			}
		}()
		c.Next()
	})
	install.Mount(g.Group("/api/install"), o.Installer)
	if o.DB != nil && o.Redis != nil {
		settings := repo.NewSettingsRepo(o.DB)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if e := settings.Refresh(ctx); e != nil {
			return nil, e
		}
		if v := settings.Snapshot(); v.MysqlMaxConn > 0 {
			o.DB.SetMaxOpenConns(v.MysqlMaxConn)
			o.DB.SetMaxIdleConns(v.MysqlMaxIdle)
		}
		entry := settings.Snapshot().AdminEntry
		if entry == "" {
			entry = "admin"
		}
		if !install.ValidAdminEntry(entry) {
			return nil, http.ErrNotSupported
		}
		actualEntry := entry
		o.Installer.AdminEntryFunc = func() string {
			v, err := settings.Get(context.Background())
			if err == nil && v.AdminEntry != "" {
				return v.AdminEntry
			}
			return actualEntry
		}
		if o.CanonicalAdmin {
			entry = "admin"
		}
		admins := repo.NewAdminRepo(o.DB)
		ttl := o.Config.Session.TTLSec
		if v := settings.Snapshot().AdminTokenExpire; v > 0 {
			ttl = v
		}
		sessions := auth.NewSessionStore(o.Redis, time.Duration(ttl)*time.Second, time.Duration(o.Config.Session.AbsoluteTTLSec)*time.Second)
		sessions.TTLProvider = func() time.Duration { return time.Duration(settings.Snapshot().AdminTokenExpire) * time.Second }
		h := handler.NewAdminAuthHandler(admins, sessions, auth.NewCaptchaService(o.Redis))
		public := g.Group("/api/" + entry)
		public.POST("/login", middleware.SecurityEvents(o.DB, o.Redis), middleware.PublicAdminLimit(o.Redis), h.Login)
		public.GET("/captcha", middleware.PublicAdminLimit(o.Redis), h.Captcha)
		admin := public.Group("")
		admin.Use(middleware.AdminAuth(sessions, admins), middleware.AppID(), middleware.AdminPermissions())
		admin.Use(handler.AdminAudit(o.DB))
		handler.MountLocalGeo(admin, o.LocalGeo)
		admin.POST("/logout", h.Logout)
		admin.GET("/me", h.Me)
		me := handler.NewAdminMeHandler(admins, sessions)
		admin.PUT("/me/profile", me.Profile)
		admin.PUT("/me/password", me.Password)
		apps := handler.NewAdminAppHandler(repo.NewAppRepo(o.DB), o.Redis)
		apps.LocalGeo = o.LocalGeo
		admin.GET("/app/list", apps.List)
		admin.POST("/app", apps.Create)
		admin.PUT("/app/:id", apps.Update)
		admin.DELETE("/app/:id", apps.Delete)
		handler.MountContent(admin, o.DB, o.Redis)
		handler.MountUsers(admin, o.DB, o.Redis)
		handler.MountKami(admin, o.DB, o.Redis)
		handler.MountEncrypt(admin, o.DB)
		handler.MountFunctions(admin, o.DB)
		handler.MountKamiDevices(admin, o.DB, o.Redis)
		handler.MountUserSessions(admin, o.DB, o.Redis)
		handler.MountObservability(admin, o.DB, o.Redis)
		files := handler.NewFileService(o.DB, o.Config.Upload, o.Config.Clone)
		files.Mount(admin)
		files.MountDownloads(g)
		diffs, err := buildpool.NewDiffStore(o.Config.Clone)
		if err != nil {
			return nil, err
		}
		versions := &handler.VersionService{DB: o.DB, Files: files, Diffs: diffs, Config: o.Config.Clone}
		versions.Mount(admin)
		sdkSessions := auth.NewSDKSessionStore(o.Redis, 24*time.Hour)
		sdk := g.Group("/api/sdk")
		sdk.Use(middleware.SecurityEvents(o.DB, o.Redis), middleware.SDKApp(repo.NewAppRepo(o.DB), repo.NewEncryptRepo(o.DB)), middleware.SDKRateLimit(o.Redis), handler.SDKMetrics(o.Redis), middleware.SDKSign(o.Redis))
		sdk.POST("/handshake", handler.NewSDKHandshakeHandler(sdkSessions).Handle)
		business := sdk.Group("")
		business.Use(middleware.SDKCrypto(sdkSessions))
		service := handler.NewSDKService(o.DB, o.Redis)
		service.LocalGeo = o.LocalGeo
		service.SetCloneConfig(o.Config.Clone)
		service.SetUploadConfig(o.Config.Upload)
		service.Mount(business)
		rawSDK := g.Group("/api/sdk")
		rawSDK.Use(middleware.SecurityEvents(o.DB, o.Redis), middleware.SDKApp(repo.NewAppRepo(o.DB), repo.NewEncryptRepo(o.DB)), middleware.SDKRateLimit(o.Redis), handler.SDKMetrics(o.Redis))
		service.MountRaw(rawSDK, files, versions)
		super := admin.Group("")
		super.Use(middleware.RequireSuper())
		handler.MountAdmins(super, admins, sessions)
		handler.MountSecurityEvents(super, o.DB)
		handler.MountBackup(super, o.Backup)
		handler.MountGeoSettings(super, o.DB, o.LocalGeo)
		super.GET("/system/settings", func(c *gin.Context) { response.OK(c, settings.Snapshot()) })
		handler.MountSystem(super, o.DB, settings, o.OnChange)
		handler.MountCertificates(admin, super, o.DB, o.OnTLS)
	}
	if o.Web != nil {
		files := http.FileServer(http.FS(o.Web))
		g.NoRoute(func(c *gin.Context) {
			uploadURL := strings.TrimRight(o.Config.Upload.URLPrefix, "/")
			if uploadURL == "" {
				uploadURL = "/uploads"
			}
			if c.Request.URL.Path == uploadURL || strings.HasPrefix(c.Request.URL.Path, uploadURL+"/") {
				c.Status(404)
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, response.Envelope{Code: 404, Msg: "接口不存在"})
				return
			}
			if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
				c.Status(405)
				return
			}
			p := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
			if p == "." {
				p = "index.html"
			}
			if info, e := fs.Stat(o.Web, p); e == nil && !info.IsDir() {
				files.ServeHTTP(c.Writer, c.Request)
				return
			}
			b, e := fs.ReadFile(o.Web, "index.html")
			if e != nil {
				c.Status(404)
				return
			}
			c.Data(200, "text/html; charset=utf-8", b)
		})
	} else {
		g.NoRoute(func(c *gin.Context) {
			c.JSON(404, response.Envelope{Code: 404, Msg: "接口不存在"})
		})
	}
	return g, nil
}
