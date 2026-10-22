-- 用户分组(会员等级)
CREATE TABLE IF NOT EXISTS {user_groups} (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  appid       BIGINT NOT NULL,
  name        VARCHAR(64) NOT NULL,
  level       INT NOT NULL DEFAULT 0,
  remark      VARCHAR(255) NOT NULL DEFAULT '',
  create_time INT UNSIGNED NOT NULL DEFAULT 0,
  KEY idx_appid (appid),
  KEY idx_level (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- 用户关联分组
ALTER TABLE {users} ADD COLUMN group_id BIGINT NOT NULL DEFAULT 0 AFTER appid;
ALTER TABLE {users} ADD KEY idx_group (group_id);
