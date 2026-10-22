-- SDK 终端用户独立的 APK 签名密钥(随机包名 clone 用同一把签)。
-- sign_pk8  : PKCS#8 DER 编码的 RSA 2048 私钥(~1.2KB)
-- sign_x509 : 自签 X.509 证书 DER(~1KB)
-- 客户端可在自己页面触发 /api/sdk/regen-sign 重新生成,服务端落库后续 clone 复用。
ALTER TABLE {users}
  ADD COLUMN sign_pk8  BLOB NULL COMMENT '每用户 APK 签名私钥 PKCS#8 DER',
  ADD COLUMN sign_x509 BLOB NULL COMMENT '每用户 APK 签名自签证书 X.509 DER';
