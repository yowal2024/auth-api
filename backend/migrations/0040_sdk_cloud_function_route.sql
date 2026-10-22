-- 新增 SDK 接口 cloudFunction:客户端按 name 手动调用 trigger=manual 且启用的云函数。
-- 服务端在 QuickJS 沙箱执行,注入 user/ctx.payload/app/ctx.ip,成功返回函数 return 对象。
-- sdkroute.Register 启动时按 api_routes 表注册路由,这里 INSERT 一行就能挂上。
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('cloudFunction', '云函数调用', 200)
ON DUPLICATE KEY UPDATE name = VALUES(name);
