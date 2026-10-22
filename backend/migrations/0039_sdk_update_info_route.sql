-- 新增 SDK 接口 updateInfo:当前登录用户自助修改昵称 / 头像。
-- 普通用户改 nickname + avatar,卡密用户只改 nickname。
-- sdkroute.Register 启动时按 api_routes 表注册路由,这里 INSERT 一行就能挂上。
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('updateInfo', '修改资料', 190)
ON DUPLICATE KEY UPDATE name = VALUES(name);
