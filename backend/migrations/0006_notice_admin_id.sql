-- 公告改为引用管理员 ID,列表显示当前昵称(管理员改名后同步)
ALTER TABLE `{notices}`
  ADD COLUMN `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `enabled`,
  ADD KEY `idx_admin_id` (`admin_id`);

-- 历史数据回填:用 author 同时匹配 admins.username / admins.nickname
UPDATE `{notices}` n
LEFT JOIN `{admins}` a
  ON a.username = n.author OR (a.nickname <> '' AND a.nickname = n.author)
SET n.admin_id = IFNULL(a.id, 0);
