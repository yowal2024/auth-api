ALTER TABLE `{users}` ADD COLUMN last_active_time BIGINT NOT NULL DEFAULT 0;
ALTER TABLE `{kamis}` ADD COLUMN last_active_time BIGINT NOT NULL DEFAULT 0;
UPDATE `{users}` SET last_active_time=last_login_time;
UPDATE `{kamis}` SET last_active_time=last_login_time;
CREATE TABLE `{security_events}` (
 id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY, appid BIGINT NOT NULL DEFAULT 0,
 kind VARCHAR(64) NOT NULL, summary VARCHAR(255) NOT NULL, ip VARCHAR(64) NOT NULL DEFAULT '',
 subject VARCHAR(80) NOT NULL DEFAULT '', route VARCHAR(128) NOT NULL DEFAULT '',
 create_time BIGINT NOT NULL, INDEX time_idx(create_time), INDEX app_time_idx(appid,create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE `{backup_settings}` (
 id INT NOT NULL PRIMARY KEY, enabled BOOLEAN NOT NULL DEFAULT FALSE,
 interval_hours INT NOT NULL DEFAULT 24, keep_count INT NOT NULL DEFAULT 7,
 include_files BOOLEAN NOT NULL DEFAULT TRUE
) ENGINE=InnoDB;
INSERT INTO `{backup_settings}` (id) VALUES (1);
CREATE TABLE `{backup_runs}` (
 id VARCHAR(64) NOT NULL PRIMARY KEY, status VARCHAR(16) NOT NULL,
 started_at BIGINT NOT NULL, finished_at BIGINT NOT NULL DEFAULT 0,
 size BIGINT NOT NULL DEFAULT 0, error VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB;
