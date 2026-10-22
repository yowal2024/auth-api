-- 初始化 settings 单行(若不存在)
INSERT INTO `{settings}`
  (id, web_url, upload_limit, admin_token_expire, admin_entry,
   mysql_max_conn, mysql_max_idle, cache_enabled, user_log_enabled,
   agent_log_enabled, admin_log_enabled, log_level)
SELECT 1, '', 2, 604800, 'admin', 150, 20, 0, 1, 0, 1, 'all'
FROM dual
WHERE NOT EXISTS (SELECT 1 FROM `{settings}` WHERE id = 1);

-- 种 api_codes (前端 mock 的默认列表)
INSERT IGNORE INTO `{api_codes}` (code, message, remark) VALUES
  (0,    '操作成功',       '通用成功响应'),
  (400,  '参数错误',       '请求参数缺失或不合法'),
  (401,  '未授权',         'Token 无效或已过期'),
  (403,  '禁止访问',       '权限不足'),
  (404,  '资源不存在',     '请求的资源未找到'),
  (429,  '请求过于频繁',   '触发限流'),
  (500,  '服务器内部错误', '后端异常'),
  (1001, '账号或密码错误', '登录验证失败'),
  (1002, '账号已被封禁',   '用户状态异常'),
  (2001, '卡密不存在',     '卡密校验失败'),
  (2002, '卡密已被使用',   '卡密已激活');

-- 种 api_routes
INSERT IGNORE INTO `{api_routes}` (method, path, name, `group`, enabled) VALUES
  ('POST', '/api/user/login',    '用户登录',     'user', 1),
  ('POST', '/api/user/register', '用户注册',     'user', 1),
  ('POST', '/api/user/logout',   '用户登出',     'user', 1),
  ('GET',  '/api/user/info',     '获取用户信息', 'user', 1),
  ('POST', '/api/kami/use',      '使用卡密',     'kami', 1),
  ('POST', '/api/kami/verify',   '验证卡密',     'kami', 1),
  ('GET',  '/api/app/version',   '获取应用版本', 'app',  1),
  ('GET',  '/api/app/notice',    '获取应用公告', 'app',  1);
