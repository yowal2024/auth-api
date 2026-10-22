// Protocol reconstructed using original disassembly and embedded SDK documentation.
package middleware

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"auth-api/internal/sdkcrypto"
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"strconv"
	"strings"
	"time"
)

func SDKApp(apps *repo.AppRepo, encrypts *repo.EncryptRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, e := strconv.ParseInt(c.GetHeader("X-App-Id"), 10, 64)
		if e != nil || id <= 0 {
			response.Fail(c, 1010, "请求头 X-App-Id 缺失或非法")
			c.Abort()
			return
		}
		app, e := apps.GetByID(c.Request.Context(), id)
		if e != nil {
			response.ServerError(c, "读取应用失败")
			c.Abort()
			return
		}
		if app == nil || app.Status != 1 {
			response.Fail(c, 403, "应用不存在或已停用")
			c.Abort()
			return
		}
		if app.MtlsEnabled {
			certPEM, err := apps.CACertificate(c.Request.Context(), id)
			if err != nil || c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
				response.Fail(c, 1019, "缺少客户端证书")
				c.Abort()
				return
			}
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM([]byte(certPEM))
			peer := c.Request.TLS.PeerCertificates[0]
			if _, err = peer.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); err != nil || peer.Subject.CommonName != fmt.Sprintf("app-%d", id) {
				response.Fail(c, 1019, "客户端证书与应用不匹配")
				c.Abort()
				return
			}
		}
		c.Set("sdkApp", app)
		c.Set("appid", id)
		if app.Encryption > 0 {
			en, e := encrypts.GetByID(c.Request.Context(), id, app.Encryption)
			if e != nil || en == nil {
				response.ServerError(c, "应用加密方案不可用")
				c.Abort()
				return
			}
			c.Set("sdkEncrypt", en)
		}
		c.Next()
	}
}
func CurrentApp(c *gin.Context) *repo.App { v, _ := c.Get("sdkApp"); a, _ := v.(*repo.App); return a }
func CurrentEncryption(c *gin.Context) *repo.Encrypt {
	v, _ := c.Get("sdkEncrypt")
	a, _ := v.(*repo.Encrypt)
	return a
}
func SDKSign(r *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := CurrentApp(c)
		if a == nil {
			response.Fail(c, 1010, "应用无效")
			c.Abort()
			return
		}
		if a.TokenKey == "" {
			c.Next()
			return
		}
		raw, e := io.ReadAll(io.LimitReader(c.Request.Body, (2<<20)+1))
		if e != nil || len(raw) > 2<<20 {
			response.BadRequest(c, "请求体过大或无效")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		ts, nonce, sig := c.GetHeader("X-Ts"), c.GetHeader("X-Nonce"), c.GetHeader("X-Sign")
		timestamp, e := strconv.ParseInt(ts, 10, 64)
		if e != nil || nonce == "" || len(nonce) > 256 || !sdkcrypto.VerifySign(sdkcrypto.SignBytes(a.TokenKey, ts, nonce, raw), sig) {
			response.Fail(c, 1014, "签名校验失败")
			c.Abort()
			return
		}
		skew := int64(60)
		if en := CurrentEncryption(c); en != nil && en.TimeSkew > 0 {
			skew = int64(en.TimeSkew)
		}
		if skew > 3600 {
			skew = 3600
		}
		now := time.Now().Unix()
		if timestamp < now-skew || timestamp > now+skew {
			response.Fail(c, 1016, "请求时间窗超出")
			c.Abort()
			return
		}
		digest := sha256.Sum256([]byte(nonce))
		key := "sdk:nonce:" + strconv.FormatInt(a.ID, 10) + ":" + hex.EncodeToString(digest[:])
		ok, e := r.SetNX(c.Request.Context(), key, 1, time.Duration(2*skew+1)*time.Second).Result()
		if e != nil {
			response.ServerError(c, "签名服务不可用")
			c.Abort()
			return
		}
		if !ok {
			response.Fail(c, 1014, "请求重复")
			c.Abort()
			return
		}
		c.Next()
	}
}

type payloadWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *payloadWriter) Write(b []byte) (int, error)       { return w.body.Write(b) }
func (w *payloadWriter) WriteString(s string) (int, error) { return w.body.WriteString(s) }
func (w *payloadWriter) WriteHeader(n int)                 { w.status = n }
func (w *payloadWriter) WriteHeaderNow()                   {}
func (w *payloadWriter) Status() int {
	if w.status == 0 {
		return 200
	}
	return w.status
}
func (w *payloadWriter) Size() int     { return w.body.Len() }
func (w *payloadWriter) Written() bool { return w.body.Len() > 0 }
func SDKCrypto(store *auth.SDKSessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		en := CurrentEncryption(c)
		if en == nil || en.Algorithm == "none" {
			c.Next()
			return
		}
		local := *en
		if en.Algorithm == "session" {
			id := c.GetHeader("X-Session-Id")
			session, e := store.Get(c.Request.Context(), id)
			if e != nil || session == nil || session.AppID != CurrentApp(c).ID {
				response.Fail(c, 1020, "session 失效")
				c.Abort()
				return
			}
			local.SymKey = session.AESKey
			if e = store.RefreshTTL(c.Request.Context(), id); e != nil {
				response.ServerError(c, "session 服务不可用")
				c.Abort()
				return
			}
		}
		var envelope struct {
			Data string `json:"data"`
		}
		if json.NewDecoder(io.LimitReader(c.Request.Body, 2<<20)).Decode(&envelope) != nil {
			response.Fail(c, 1015, "加密数据格式错误")
			c.Abort()
			return
		}
		plain, e := sdkcrypto.DecryptPayload(&local, envelope.Data)
		if e != nil || !json.Valid(plain) {
			response.Fail(c, 1015, "加密数据格式错误")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Request.ContentLength = int64(len(plain))
		c.Request.Header.Set("Content-Type", "application/json")
		original := c.Writer
		buffer := &payloadWriter{ResponseWriter: original}
		c.Writer = buffer
		c.Next()
		c.Writer = original
		var out map[string]json.RawMessage
		if e = json.Unmarshal(buffer.body.Bytes(), &out); e != nil {
			response.ServerError(c, "响应编码失败")
			return
		}
		if data, ok := out["data"]; ok && string(data) != "null" {
			encoded, err := sdkcrypto.EncryptPayload(&local, data)
			if err != nil {
				response.ServerError(c, "响应加密失败")
				return
			}
			out["data"], _ = json.Marshal(encoded)
		}
		c.JSON(buffer.Status(), out)
	}
}
func SDKToken(store *auth.UserSessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Token")
		if token == "" {
			token = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		u, e := store.Get(c.Request.Context(), token)
		if e != nil || u == nil || u.AppID != CurrentApp(c).ID {
			response.Fail(c, 1006, "未登录或登录已过期")
			c.Abort()
			return
		}
		c.Set("sdkUser", u)
		c.Set("sdkToken", token)
		c.Next()
	}
}
func CurrentSDKUser(c *gin.Context) *auth.UserSession {
	v, _ := c.Get("sdkUser")
	u, _ := v.(*auth.UserSession)
	return u
}
