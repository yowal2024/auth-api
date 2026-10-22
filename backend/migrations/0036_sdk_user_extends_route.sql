-- 新增 SDK 接口 userExtends:一次返回当前登录用户在后台「用户编辑 → 扩展项」里配置的所有 key/value。
-- 与 ini 接口的 data.extends(应用级)对称,这里是用户级、按 uid 隔离。
-- sdkroute.Register 启动时按 api_routes 表注册路由,这里 INSERT 一行就能挂上。
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('userExtends', '获取用户扩展', 180)
ON DUPLICATE KEY UPDATE name = VALUES(name);
