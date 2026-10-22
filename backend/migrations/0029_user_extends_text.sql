-- SDK 远程配置存储复用 user_extends:e_key='_sdk_config',e_value 是整段 JSON 字符串。
-- 原 VARCHAR 长度不足以放完整客户端配置,统一拓到 TEXT。
ALTER TABLE `{user_extends}` MODIFY COLUMN `e_value` TEXT NOT NULL;

-- 让 sdkroute 启动时能挂上这两条新接口(api_routes 没行就不会注册)。
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('config',       '获取远程配置', 160),
  ('configSubmit', '提交远程配置', 170)
ON DUPLICATE KEY UPDATE name = VALUES(name);
