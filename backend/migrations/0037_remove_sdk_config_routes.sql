-- 0036 引入 userExtends 接口(读 + 写一体)后,老的 config / configSubmit 两个接口下线。
-- 旧 ConfigManager 客户端把整段 JSON 存 user_extends._sdk_config 那行,改成按 key 拆开存。
-- 这里清掉 api_routes 表里的两行,sdkroute.Register 启动扫表时就不会再尝试挂老路由。
-- 旧 user_extends 表里残留的 e_key='_sdk_config' 行不做清理(管理员后台可见,有需要手动删)。
DELETE FROM `{api_routes}` WHERE api_key IN ('config', 'configSubmit');
