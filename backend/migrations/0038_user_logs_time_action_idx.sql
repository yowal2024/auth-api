-- dashboard 趋势查询 (TrendByDate) 用到 WHERE create_time BETWEEN ? AND ? AND action IN (?)
-- 现有索引只有 (id) 主键和 (uid, create_time)，按时间 + action 过滤要全表扫。
-- 加复合索引让 user_logs 几十万行后趋势接口仍能毫秒返回。
ALTER TABLE `{user_logs}` ADD INDEX `idx_time_action` (`create_time`, `action`);
