-- users.vip / kamis.vip 原本 INT UNSIGNED,上限 4294967295 (≈ 2106-02-07 UTC)。
-- 后台编辑用户/卡密时,选「永久」会传超出该上限的时间戳,触发 MySQL
-- "Out of range value for column 'vip' at row 1"。
-- Go 侧字段类型本来就是 int64,这里把列升到 BIGINT(signed)与之对齐,
-- 既能容纳"永久"语义下的远期时间戳,也消除 sqlx 在 unsigned/signed 之间的隐式转换。
-- 注意:COMMENT 内不要写分号,migrator splitStatements 是按 `;` 朴素切分,字符串内的分号会被当语句分隔。
ALTER TABLE `{users}` MODIFY COLUMN `vip` BIGINT NOT NULL DEFAULT 0 COMMENT 'VIP 到期 unix 秒,0 表示无 VIP';
ALTER TABLE `{kamis}` MODIFY COLUMN `vip` BIGINT NOT NULL DEFAULT 0 COMMENT 'VIP 到期 unix 秒,0 表示未激活';
