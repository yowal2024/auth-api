-- ============================================================
-- SDK 业务接口的查询索引优化 + user_devices 唯一键
-- ============================================================

-- blocks 黑名单查询走 (appid, type, value) 三列定位,加复合索引
ALTER TABLE `{blocks}` ADD INDEX `idx_block_lookup` (`appid`, `type`, `value`);

-- 同一用户同一设备 ID 只允许一条绑定记录,防并发重复插入
ALTER TABLE `{user_devices}` ADD UNIQUE KEY `uniq_uid_mi` (`uid`, `mi`);
