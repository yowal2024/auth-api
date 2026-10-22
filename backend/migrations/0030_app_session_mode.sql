-- ============================================================
-- SDK 会话密钥协商(handshake + per-session AES)
-- ============================================================
-- 设计调整:不再用 apps.session_mode 独立字段,改用 encrypts.algorithm = "session"
-- 表达"该 app 启用 handshake 模式"。algorithm=session 的 encrypt 记录仍要填
-- server_public_key / server_private_key 这对 RSA,专门给 handshake 用。
-- 本 migration 仅注册 handshake 接口 + session 失效错误码,字段相关已迁移到 0031。

-- 把 handshake 接口加进 api_routes,sdkroute 启动时按表挂载
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('handshake', 'SDK 会话密钥协商', 180)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- session 失效错误码
INSERT INTO `{api_codes}` (code, message, remark) VALUES
  (1020, 'session 失效', 'SDK 会话密钥过期/不存在,客户端需重新 handshake')
ON DUPLICATE KEY UPDATE message=VALUES(message), remark=VALUES(remark);
