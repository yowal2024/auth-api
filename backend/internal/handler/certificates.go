package handler

import (
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"auth-api/internal/secretkey"
	"auth-api/internal/tlsmgr"
	"crypto/tls"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"os"
	"sync"
)

var keyMu sync.Mutex

func masterKey(existing bool) ([]byte, error) {
	keyMu.Lock()
	defer keyMu.Unlock()
	if existing {
		b, e := os.ReadFile("data/master.key")
		if e != nil {
			return nil, e
		}
		return b, nil
	}
	return secretkey.Load("data/master.key")
}
func MountCertificates(admin, super *gin.RouterGroup, d *sqlx.DB, onTLS func() error) {
	admin.GET("/app/:id/mtls", func(c *gin.Context) {
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var cert string
		e := d.GetContext(c.Request.Context(), &cert, repo.SQL("SELECT COALESCE(mtls_ca_cert,'') FROM {apps} WHERE id=?"), id)
		if !dbError(c, e) {
			response.OK(c, gin.H{"hasCA": cert != "", "fingerprint": tlsmgr.Fingerprint(cert)})
		}
	})
	admin.POST("/app/:id/mtls/gen-ca", func(c *gin.Context) {
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		a, e := repo.NewAppRepo(d).GetByID(c.Request.Context(), id)
		if dbError(c, e) {
			return
		}
		if a == nil {
			response.Fail(c, 404, "应用不存在")
			return
		}
		cert, key, e := tlsmgr.Generate(a.Name, true)
		if dbError(c, e) {
			return
		}
		master, e := masterKey(false)
		if dbError(c, e) {
			return
		}
		encrypted, e := secretkey.Encrypt(master, []byte(key))
		if dbError(c, e) {
			return
		}
		_, e = d.ExecContext(c.Request.Context(), repo.SQL("UPDATE {apps} SET mtls_ca_cert=?,mtls_ca_key=? WHERE id=?"), cert, encrypted, id)
		if !dbError(c, e) {
			response.OK(c, gin.H{"caCert": cert, "fingerprint": tlsmgr.Fingerprint(cert)})
		}
	})
	admin.POST("/app/:id/mtls/issue-cert", func(c *gin.Context) {
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var row struct {
			Cert string `db:"cert"`
			Key  string `db:"key"`
		}
		e := d.GetContext(c.Request.Context(), &row, repo.SQL("SELECT COALESCE(mtls_ca_cert,'') AS cert,COALESCE(mtls_ca_key,'') AS `key` FROM {apps} WHERE id=?"), id)
		if dbError(c, e) {
			return
		}
		if row.Cert == "" {
			response.BadRequest(c, "该应用尚未生成 CA,请先生成 CA")
			return
		}
		master, e := masterKey(true)
		if dbError(c, e) {
			return
		}
		key, e := secretkey.Decrypt(master, row.Key)
		if dbError(c, e) {
			return
		}
		cert, priv, e := tlsmgr.IssueClient(row.Cert, string(key), id)
		if !dbError(c, e) {
			response.OK(c, gin.H{"cert": cert, "key": priv, "fingerprint": tlsmgr.Fingerprint(cert)})
		}
	})
	super.GET("/system/https", func(c *gin.Context) {
		v, e := repo.NewSettingsRepo(d).Get(c.Request.Context())
		if !dbError(c, e) {
			response.OK(c, gin.H{"enabled": v.HTTPSEnabled == 1, "addr": v.HTTPSAddr, "hasCert": v.HTTPSCert != "", "cert": v.HTTPSCert, "fingerprint": tlsmgr.Fingerprint(v.HTTPSCert)})
		}
	})
	savePair := func(c *gin.Context, cert, key string) bool {
		if _, e := tls.X509KeyPair([]byte(cert), []byte(key)); e != nil {
			response.BadRequest(c, "证书或私钥无效")
			return false
		}
		master, e := masterKey(false)
		if dbError(c, e) {
			return false
		}
		enc, e := secretkey.Encrypt(master, []byte(key))
		if dbError(c, e) {
			return false
		}
		_, e = d.ExecContext(c.Request.Context(), repo.SQL("UPDATE {settings} SET https_cert=?,https_key=? WHERE id=1"), cert, enc)
		return !dbError(c, e)
	}
	super.POST("/system/https/gen-cert", func(c *gin.Context) {
		var v httpsGenReq
		if c.ShouldBindJSON(&v) != nil || v.Host == "" {
			response.BadRequest(c, "主机名不能为空")
			return
		}
		cert, key, e := tlsmgr.Generate(v.Host, false)
		if dbError(c, e) || !savePair(c, cert, key) {
			return
		}
		if onTLS != nil {
			if e = onTLS(); dbError(c, e) {
				return
			}
		}
		response.OK(c, gin.H{"cert": cert, "fingerprint": tlsmgr.Fingerprint(cert)})
	})
	super.POST("/system/https", func(c *gin.Context) {
		var v httpsSaveReq
		if c.ShouldBindJSON(&v) != nil {
			response.BadRequest(c, "请求体格式错误")
			return
		}
		if v.Cert != nil || v.Key != nil {
			if v.Cert == nil || v.Key == nil {
				response.BadRequest(c, "请同时提供证书与私钥")
				return
			}
			if !savePair(c, *v.Cert, *v.Key) {
				return
			}
		}
		current, e := repo.NewSettingsRepo(d).Get(c.Request.Context())
		if dbError(c, e) {
			return
		}
		enabled := current.HTTPSEnabled
		if v.Enabled != nil {
			enabled = 0
			if *v.Enabled {
				enabled = 1
			}
		}
		addr := v.Addr
		if addr == "" {
			addr = current.HTTPSAddr
		}
		if addr == "" {
			addr = ":8443"
		}
		if enabled == 1 && current.HTTPSCert == "" {
			response.BadRequest(c, "请先生成或上传服务端证书再开启 HTTPS")
			return
		}
		_, e = d.ExecContext(c.Request.Context(), repo.SQL("UPDATE {settings} SET https_enabled=?,https_addr=? WHERE id=1"), enabled, addr)
		if dbError(c, e) {
			return
		}
		if onTLS != nil {
			if e = onTLS(); e != nil {
				c.Set("audit.hotSwapFailed", true)
				response.OK(c, gin.H{"hotSwap": false, "hotSwapError": e.Error()})
				return
			}
		}
		response.OK(c, gin.H{"hotSwap": true})
	})
}
