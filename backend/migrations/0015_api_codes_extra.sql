-- ============================================================
-- SDK 业务接口新增的错误码,补进 api_codes(后台"状态码管理"可见)
-- 注意:1004/1005 已被 sdk_version 占用(未找到可用版本 / 未启用随机包名),不能复用
-- ============================================================
INSERT INTO `{api_codes}` (code, message, remark) VALUES
  (1003, 'IP/设备/区域被封禁',   '黑名单命中'),
  (1006, '未登录或 token 失效',  'SDK 用户 token 校验失败'),
  (1007, '用户名已被占用',       '注册冲突'),
  (1008, '设备数达上限',         'bind_limit 命中'),
  (1014, '签名校验失败',         'HMAC sign mismatch'),
  (1015, '加密数据格式错误',     'AES envelope 解析失败'),
  (1016, '请求时间窗超出',       'X-Ts 与服务端时间差超出 time_skew')
ON DUPLICATE KEY UPDATE message = VALUES(message), remark = VALUES(remark);
