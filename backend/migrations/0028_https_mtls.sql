-- ============================================================
-- HTTPS + per-app 双向认证(mTLS)
-- ============================================================
-- settings:HTTPS 服务端配置(单行全局)。https_key 存 AES-GCM 密文。
ALTER TABLE `{settings}`
  ADD COLUMN `https_enabled` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT 'HTTPS 监听开关',
  ADD COLUMN `https_addr`    VARCHAR(32) NOT NULL DEFAULT ':8443' COMMENT 'HTTPS 监听地址',
  ADD COLUMN `https_cert`    TEXT        NULL COMMENT '服务端证书 PEM(明文)',
  ADD COLUMN `https_key`     TEXT        NULL COMMENT '服务端私钥 PEM(AES-GCM 密文)';

-- apps:每应用一套 mTLS CA。mtls_ca_key 存 AES-GCM 密文,API 永不回吐。
ALTER TABLE `{apps}`
  ADD COLUMN `mtls_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'SDK 双向认证开关' AFTER `brute_force_enabled`,
  ADD COLUMN `mtls_ca_cert` TEXT       NULL COMMENT 'app CA 证书 PEM(明文)' AFTER `mtls_enabled`,
  ADD COLUMN `mtls_ca_key`  TEXT       NULL COMMENT 'app CA 私钥 PEM(AES-GCM 密文)' AFTER `mtls_ca_cert`;

-- mTLS 拒绝码
INSERT INTO `{api_codes}` (code, message, remark) VALUES
  (1019, '客户端证书校验失败', 'SDK mTLS:证书缺失/CN 与 appid 不符/非本 app CA 签发')
ON DUPLICATE KEY UPDATE message=VALUES(message), remark=VALUES(remark);
