package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/middleware"
	"auth-api/internal/response"
	"auth-api/internal/sdkcrypto"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"time"
)

func NewSDKHandshakeHandler(store *auth.SDKSessionStore) *SDKHandshakeHandler {
	return &SDKHandshakeHandler{sessStore: store}
}
func (h *SDKHandshakeHandler) Handle(c *gin.Context) {
	app, en := middleware.CurrentApp(c), middleware.CurrentEncryption(c)
	if app == nil || en == nil || en.Algorithm != "session" {
		response.BadRequest(c, "应用未配置 session 加密")
		return
	}
	var v handshakeRequest
	if c.ShouldBindJSON(&v) != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	b, e := sdkcrypto.DecodeBlob(v.Key, en.Encoding)
	if e != nil {
		response.Fail(c, 1015, "握手密钥格式错误")
		return
	}
	key, e := sdkcrypto.RSADecrypt(en.ServerPrivateKey, b)
	if e != nil || len(key) != 32 {
		response.Fail(c, 1015, "握手密钥格式错误")
		return
	}
	id, e := auth.RandomToken()
	if dbError(c, e) {
		return
	}
	if e = h.sessStore.Set(c.Request.Context(), id, &auth.SDKSession{AppID: app.ID, AESKey: hex.EncodeToString(key), CreateTime: time.Now().Unix()}); e != nil {
		response.ServerError(c, "保存握手会话失败")
		return
	}
	payload, _ := json.Marshal(gin.H{"sessionId": id})
	encrypted, e := sdkcrypto.EncryptAESCBC(key, payload)
	if dbError(c, e) {
		return
	}
	response.OK(c, sdkcrypto.EncodeBlob(encrypted, en.Encoding))
}
