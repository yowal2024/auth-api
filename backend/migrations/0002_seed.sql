-- 默认数据
-- 默认管理员行由安装向导 CreateAdmin 写入,这里不再 seed。
-- 默认 settings 1 行(admin_entry 由安装向导填入)。

SET NAMES utf8mb4;

-- 默认设置 (单行)
INSERT INTO `{settings}` (
  `id`, `web_url`, `upload_limit`, `admin_token_expire`, `admin_entry`,
  `mysql_max_conn`, `mysql_max_idle`,
  `cache_enabled`, `user_log_enabled`, `agent_log_enabled`, `admin_log_enabled`,
  `log_level`
) VALUES (
  1, '', 10, 604800, '',
  50, 10,
  1, 1, 1, 1,
  'info'
) ON DUPLICATE KEY UPDATE `id` = `id`;
