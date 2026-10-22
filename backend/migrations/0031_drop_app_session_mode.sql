-- ============================================================
-- 回滚 0030 里的 apps.session_mode 字段
-- ============================================================
-- 设计调整:会话密钥协商不再用 apps.session_mode 表达,改用
-- encrypts.algorithm = "session" 这一新算法值。原 ALTER 留下的字段需要 DROP 掉。
--
-- migrator 对 1091(Can't DROP)做了幂等跳过,所以即使 0030 在未应用此字段的
-- 全新环境跑过也安全。
ALTER TABLE `{apps}` DROP COLUMN `session_mode`;
