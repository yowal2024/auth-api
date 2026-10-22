-- ============================================================
-- 重构 api_routes / api_codes 为「逻辑 Key + 重写」模型
-- ============================================================
-- 路由:每行是一个逻辑 API (kamiTopup / vip / info / ...)
-- 重定义开启时,override_path 是客户端 SDK 实际请求的 URL
DROP TABLE IF EXISTS `{api_routes}`;
CREATE TABLE `{api_routes}` (
  `id`            BIGINT       NOT NULL AUTO_INCREMENT,
  `api_key`       VARCHAR(64)  NOT NULL,
  `name`          VARCHAR(128) NOT NULL DEFAULT '',
  `override_path` VARCHAR(255) NOT NULL DEFAULT '',
  `rewritten`     TINYINT(1)   NOT NULL DEFAULT 0,
  `sort_order`    INT          NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_api_key` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- 种默认逻辑 API(仅已实现 handler 的接口;占位项已在 0027 清理)
INSERT INTO `{api_routes}` (api_key, name, sort_order) VALUES
  ('kamiTopup',     '卡密充值',           10),
  ('info',          '获取信息',           40),
  ('logon',         '登录',               90),
  ('ini',           '获取应用配置',       120),
  ('register',      '注册',               130),
  ('logout',        '登出',               140),
  ('heartbeat',     '心跳',               150);

-- 状态码:override_code = 0 表示未重写,使用原 code
DROP TABLE IF EXISTS `{api_codes}`;
CREATE TABLE `{api_codes}` (
  `id`            BIGINT       NOT NULL AUTO_INCREMENT,
  `code`          INT          NOT NULL,
  `override_code` INT          NOT NULL DEFAULT 0,
  `message`       VARCHAR(255) NOT NULL DEFAULT '',
  `remark`        VARCHAR(255) NOT NULL DEFAULT '',
  `rewritten`     TINYINT(1)   NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

INSERT INTO `{api_codes}` (code, message, remark) VALUES
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
