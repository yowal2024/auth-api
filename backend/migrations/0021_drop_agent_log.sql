-- 代理日志字段已下线(无代理场景),清理冗余列。
ALTER TABLE `{settings}` DROP COLUMN `agent_log_enabled`;
