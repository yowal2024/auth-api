-- 用户级扩展 e_value 在 0029 已经从 VARCHAR(255) 升到 TEXT(64KB),用于 SDK _sdk_config 远程配置。
-- 但客户端配置 JSON 可能更大(规则集、偏移、皮肤数据等),64KB 仍可能不够,与 0033 把应用级 extends.e_value
-- 升 MEDIUMTEXT 的思路一致,这里也统一拓到 MEDIUMTEXT(16MB),后续再涨也不用迁了。
-- repo 始终显式传 e_value,无需 DEFAULT,沿用 NOT NULL。
ALTER TABLE `{user_extends}` MODIFY COLUMN `e_value` MEDIUMTEXT NOT NULL COMMENT '用户级扩展值,可放整段 JSON';
