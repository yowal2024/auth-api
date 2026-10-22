package handler

import (
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"time"
)

func MountEncrypt(g *gin.RouterGroup, d *sqlx.DB) {
	r := repo.NewEncryptRepo(d)
	g.GET("/encrypt/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v, e := r.ListByApp(c.Request.Context(), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	g.POST("/encrypt", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var v encryptReq
		if c.ShouldBindJSON(&v) != nil || v.Name == "" {
			response.BadRequest(c, "名称不能为空")
			return
		}
		allowed := map[string]bool{"aes": true, "des": true, "rc4": true, "rsa": true, "session": true, "none": true}
		if !allowed[v.Algorithm] || (v.Encoding != "base64" && v.Encoding != "hex") || v.TimeSkew < 0 || v.TimeSkew > 3600 {
			response.BadRequest(c, "加密参数无效")
			return
		}
		item := repo.Encrypt{ID: v.ID, AppID: a, Name: v.Name, Algorithm: v.Algorithm, Encoding: v.Encoding, SymKey: v.SymKey, TimeSkew: v.TimeSkew, SignMode: v.SignMode, ClientPrivateKey: v.ClientPrivateKey, ClientPublicKey: v.ClientPublicKey, ServerPrivateKey: v.ServerPrivateKey, ServerPublicKey: v.ServerPublicKey, CreateTime: time.Now().Unix()}
		var e error
		if item.ID > 0 {
			old, err := r.GetByID(c.Request.Context(), a, item.ID)
			if dbError(c, err) {
				return
			}
			if old == nil {
				response.Fail(c, 404, "加密方案不存在")
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
	g.DELETE("/encrypt/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		err := r.Delete(c.Request.Context(), a, id)
		if errors.Is(err, repo.ErrEncryptionInUse) {
			response.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, repo.ErrEncryptionUnavailable) {
			response.Fail(c, 404, err.Error())
			return
		}
		if !dbError(c, err) {
			response.OK(c, nil)
		}
	})
	g.GET("/encrypt/gen-rsa-key", func(c *gin.Context) {
		key, e := rsa.GenerateKey(rand.Reader, 2048)
		if dbError(c, e) {
			return
		}
		priv, e := x509.MarshalPKCS8PrivateKey(key)
		if dbError(c, e) {
			return
		}
		pub, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if dbError(c, e) {
			return
		}
		response.OK(c, gin.H{"privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv})), "publicKey": string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub}))})
	})
}
