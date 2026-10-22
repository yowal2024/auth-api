package server

import (
	"auth-api/internal/auth"
	"auth-api/internal/repo"
	"auth-api/internal/sdkcrypto"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testSDKSecurity(t *testing.T, d *sqlx.DB, r *redis.Client, handler http.Handler, appID int64, key, token string) {
	ctx := context.Background()
	execSQL := func(query string, args ...any) {
		t.Helper()
		if _, err := d.Exec(repo.SQL(query), args...); err != nil {
			t.Fatal(err)
		}
	}
	sequence := 0
	request := func(method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
		t.Helper()
		sequence++
		ts, nonce := fmt.Sprint(time.Now().Unix()), fmt.Sprintf("security-%d", sequence)
		req := httptest.NewRequest(method, "/api/sdk/"+path, bytes.NewReader(body))
		for k, v := range map[string]string{"Content-Type": "application/json", "X-App-Id": fmt.Sprint(appID), "X-Token": token, "X-Mi": "test-device", "X-Ts": ts, "X-Nonce": nonce, "X-Sign": sdkcrypto.SignBytes(key, ts, nonce, body)} {
			req.Header.Set(k, v)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	expect := func(w *httptest.ResponseRecorder, code int) any {
		t.Helper()
		var out struct {
			Code int
			Data any
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || out.Code != code {
			t.Fatalf("expected code %d, HTTP %d: %s", code, w.Code, w.Body.String())
		}
		return out.Data
	}
	body := []byte(`{}`)
	other := &repo.App{Name: "Isolation", Type: "user", Mode: "free", TokenKey: key, Status: 1, UserTokenTTL: 3600}
	otherID, err := repo.NewAppRepo(d).Create(ctx, other)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := repo.NewAppRepo(d).Delete(ctx, otherID); err != nil {
			t.Error(err)
		}
	}()

	t.Run("TokenDeviceAndPolicyBypasses", func(t *testing.T) {
		for _, path := range []string{"info", "heartbeat", "logout", "updateInfo", "userExtends", "kamiTopup", "cloudFunction"} {
			expect(request("POST", path, body, map[string]string{"X-Token": ""}), 1006)
		}
		for _, v := range []struct{ method, path string }{{"POST", "uploadAvatar"}, {"GET", "clone"}, {"POST", "regen-sign"}} {
			expect(request(v.method, v.path, body, map[string]string{"X-Token": ""}), 1006)
		}
		expect(request("POST", "info", body, map[string]string{"X-App-Id": fmt.Sprint(otherID)}), 1006)
		expect(request("POST", "info", body, map[string]string{"X-Mi": "different"}), 1006)
		expect(request("POST", "info", body, map[string]string{"X-App-Id": ""}), 1010)
		execSQL("UPDATE {apps} SET status=0 WHERE id=?", appID)
		expect(request("POST", "heartbeat", body, nil), 403)
		execSQL("UPDATE {apps} SET status=1,type='kami' WHERE id=?", appID)
		expect(request("POST", "info", body, nil), 1006)
		execSQL("UPDATE {apps} SET type='user',mode='vip' WHERE id=?", appID)
		expect(request("POST", "heartbeat", body, nil), 1018)
		execSQL("UPDATE {apps} SET mode='free' WHERE id=?", appID)
		execSQL("UPDATE {users} SET ban_expire=? WHERE appid=?", time.Now().Add(time.Hour).Unix(), appID)
		expect(request("POST", "heartbeat", body, nil), 1002)
		execSQL("UPDATE {users} SET ban_expire=0 WHERE appid=?", appID)
		for _, block := range []struct {
			kind, value string
			code        int
		}{{"mi", "test-device", 1021}, {"ip", "192.0.2.1", 1003}} {
			execSQL("INSERT INTO {blocks} (appid,type,value,reason,create_time) VALUES (?,?,?,'security test',0)", appID, block.kind, block.value)
			expect(request("POST", "heartbeat", body, nil), block.code)
			execSQL("DELETE FROM {blocks} WHERE appid=? AND type=? AND value=?", appID, block.kind, block.value)
		}
		var hash string
		if err := d.Get(&hash, repo.SQL("SELECT password FROM {users} WHERE appid=?"), appID); err != nil {
			t.Fatal(err)
		}
		execSQL("UPDATE {users} SET password='reset-hash' WHERE appid=?", appID)
		expect(request("POST", "info", body, nil), 1006)
		execSQL("UPDATE {users} SET password=? WHERE appid=?", hash, appID)
		store := auth.NewUserSessionStore(r, time.Hour)
		expiring, err := store.CreateWithTTL(ctx, &auth.UserSession{UID: 1, AppID: appID}, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.PExpireAt(ctx, "usess:"+expiring, time.Now().Add(-time.Second)).Err(); err != nil {
			t.Fatal(err)
		}
		expect(request("POST", "info", body, map[string]string{"X-Token": expiring}), 1006)
	})
	t.Run("SignatureReplayTamperAndTimeWindow", func(t *testing.T) {
		expect(request("POST", "info", body, map[string]string{"X-Sign": "bad"}), 1014)
		ts, nonce := fmt.Sprint(time.Now().Unix()), "explicit-replay"
		headers := map[string]string{"X-Ts": ts, "X-Nonce": nonce, "X-Sign": sdkcrypto.SignBytes(key, ts, nonce, body)}
		expect(request("POST", "info", body, headers), 0)
		expect(request("POST", "info", body, headers), 1014)
		headers["X-Nonce"] = "tampered"
		headers["X-Sign"] = sdkcrypto.SignBytes(key, ts, "tampered", body)
		expect(request("POST", "info", []byte(`{"modified":true}`), headers), 1014)
		ts = fmt.Sprint(time.Now().Add(-2 * time.Hour).Unix())
		headers = map[string]string{"X-Ts": ts, "X-Nonce": "stale", "X-Sign": sdkcrypto.SignBytes(key, ts, "stale", body)}
		expect(request("POST", "info", body, headers), 1016)
	})
	t.Run("RawFilesAvatarAndSignatureRotation", func(t *testing.T) {
		expect(request("GET", "extend-file/list", nil, nil), 0)
		expect(request("GET", "extend-file/999999999/download", nil, nil), 404)
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		p, err := mw.CreateFormFile("file", "avatar.png")
		if err != nil {
			t.Fatal(err)
		}
		// Valid 1x1 PNG, including signature checked by upload sniffing.
		png, _ := hex.DecodeString("89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000b49444154789c636000020000050001a5f645400000000049454e44ae426082")
		p.Write(png)
		mw.Close()
		expect(request("POST", "uploadAvatar", b.Bytes(), map[string]string{"Content-Type": mw.FormDataContentType()}), 0)
		expect(request("POST", "uploadAvatar", body, nil), 400)
		expect(request("POST", "regen-sign", body, nil), 0)
		var old []byte
		if err := d.Get(&old, repo.SQL("SELECT sign_x509 FROM {users} WHERE appid=?"), appID); err != nil {
			t.Fatal(err)
		}
		expect(request("POST", "regen-sign", body, nil), 0)
		var current []byte
		if err := d.Get(&current, repo.SQL("SELECT sign_x509 FROM {users} WHERE appid=?"), appID); err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(old, current) {
			t.Fatal("signing certificate did not rotate")
		}
	})
	t.Run("EncryptionAndHandshakeEndToEnd", func(t *testing.T) {
		private, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		der, err := x509.MarshalPKIXPublicKey(&private.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		pub := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
		priv := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)}))
		for _, algorithm := range []string{"aes", "des", "rc4", "rsa", "session"} {
			t.Run(algorithm, func(t *testing.T) {
				policy := &repo.Encrypt{AppID: appID, Name: "e2e", Algorithm: algorithm, Encoding: "base64", SymKey: "1234567890abcdef", TimeSkew: 60, ServerPrivateKey: priv, ServerPublicKey: pub, ClientPublicKey: pub, ClientPrivateKey: priv}
				if algorithm == "des" {
					policy.SymKey = "12345678"
				}
				id, err := repo.NewEncryptRepo(d).Create(ctx, policy)
				if err != nil {
					t.Fatal(err)
				}
				execSQL("UPDATE {apps} SET encryption=? WHERE id=?", id, appID)
				defer func() {
					execSQL("UPDATE {apps} SET encryption=0 WHERE id=?", appID)
					if err := repo.NewEncryptRepo(d).Delete(ctx, appID, id); err != nil {
						t.Error(err)
					}
				}()
				headers := map[string]string{}
				if algorithm == "session" {
					expect(request("POST", "handshake", []byte(`{"key":"invalid"}`), nil), 1015)
					symmetric := bytes.Repeat([]byte{42}, 32)
					wrapped, err := sdkcrypto.RSAEncrypt(pub, symmetric)
					if err != nil {
						t.Fatal(err)
					}
					payload, _ := json.Marshal(map[string]string{"key": sdkcrypto.EncodeBlob(wrapped, "base64")})
					answer := expect(request("POST", "handshake", payload, nil), 0).(string)
					wire, err := sdkcrypto.DecodeBlob(answer, "base64")
					if err != nil {
						t.Fatal(err)
					}
					plain, err := sdkcrypto.DecryptAESCBC(symmetric, wire)
					if err != nil {
						t.Fatal(err)
					}
					var session struct{ SessionID string }
					if err := json.Unmarshal(plain, &session); err != nil || session.SessionID == "" {
						t.Fatal("invalid session", err)
					}
					headers["X-Session-Id"] = session.SessionID
					policy.SymKey = hex.EncodeToString(symmetric)
					expect(request("POST", "info", body, nil), 1020)
					store := auth.NewSDKSessionStore(r, time.Hour)
					if err := store.Set(ctx, "foreign-session", &auth.SDKSession{AppID: otherID, AESKey: policy.SymKey}); err != nil {
						t.Fatal(err)
					}
					expect(request("POST", "info", body, map[string]string{"X-Session-Id": "foreign-session"}), 1020)
				}
				encoded, err := sdkcrypto.EncryptPayload(policy, body)
				if err != nil {
					t.Fatal(err)
				}
				payload, _ := json.Marshal(map[string]string{"data": encoded})
				answer := expect(request("POST", "info", payload, headers), 0).(string)
				plain, err := sdkcrypto.DecryptPayload(policy, answer)
				if err != nil {
					t.Fatal(err)
				}
				var profile map[string]any
				if err := json.Unmarshal(plain, &profile); err != nil || profile["user"] == nil {
					t.Fatal("encrypted response", err)
				}
				expect(request("POST", "info", []byte(`{"data":"broken"}`), headers), 1015)
				if algorithm == "session" {
					if err := r.Del(ctx, "sdksess:"+headers["X-Session-Id"]).Err(); err != nil {
						t.Fatal(err)
					}
					expect(request("POST", "info", payload, headers), 1020)
				}
			})
		}
	})
}
