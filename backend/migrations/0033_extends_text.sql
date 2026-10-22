-- 应用级扩展配置 e_value 原本 VARCHAR(255),前端管理后台已经在存大段 JSON(例如游戏偏移、
-- 规则集等),255 字符会被 MySQL 截断。统一拓到 MEDIUMTEXT(16MB),与 0029_user_extends_text
-- 把 user_extends.e_value 拓 TEXT 的思路一致,但这里量级更大,直接用 MEDIUMTEXT 避免后续再迁。
-- repo 始终显式传 e_value,无需 DEFAULT,沿用 NOT NULL。
ALTER TABLE `{extends}` MODIFY COLUMN `e_value` MEDIUMTEXT NOT NULL COMMENT '应用级扩展值,可放整段 JSON';
