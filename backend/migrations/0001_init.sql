-- auth-api 初始化 schema
-- 编码: utf8mb4_unicode_ci,所有时间戳为 INT UNSIGNED (Unix 秒)
-- 在 MySQL 8 上执行,前先 CREATE DATABASE auth DEFAULT CHARSET=utf8mb4;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 管理员
-- ============================================================
DROP TABLE IF EXISTS `{admins}`;
CREATE TABLE `{admins}` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT,
  `username`    VARCHAR(64)  NOT NULL,
  `password`    VARCHAR(128) NOT NULL,
  `role`        VARCHAR(32)  NOT NULL DEFAULT 'admin',
  `last_login`  INT UNSIGNED NOT NULL DEFAULT 0,
  `last_ip`     VARCHAR(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 应用
-- ============================================================
DROP TABLE IF EXISTS `{apps}`;
CREATE TABLE `{apps}` (
  `id`             BIGINT       NOT NULL AUTO_INCREMENT,
  `name`           VARCHAR(128) NOT NULL,
  `type`           VARCHAR(32)  NOT NULL DEFAULT 'user',
  `mode`           VARCHAR(32)  NOT NULL DEFAULT 'free',
  `logo`           VARCHAR(255) NOT NULL DEFAULT '',
  `pkg`            VARCHAR(128) NOT NULL DEFAULT '',
  `token_key`      VARCHAR(64)  NOT NULL DEFAULT '',
  `bind_device`    TINYINT(1)   NOT NULL DEFAULT 0,
  `bind_region`    TINYINT(1)   NOT NULL DEFAULT 0,
  `diff_update`    TINYINT(1)   NOT NULL DEFAULT 0,
  `random_package` TINYINT(1)   NOT NULL DEFAULT 0,
  `encryption`     BIGINT       NOT NULL DEFAULT 0,
  `status`         TINYINT(1)   NOT NULL DEFAULT 1,
  `create_time`    INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_create` (`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 用户
-- ============================================================
DROP TABLE IF EXISTS `{users}`;
CREATE TABLE `{users}` (
  `id`               BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`            BIGINT       NOT NULL DEFAULT 0,
  `username`         VARCHAR(64)  NOT NULL,
  `password`         VARCHAR(128) NOT NULL DEFAULT '',
  `nickname`         VARCHAR(64)  NOT NULL DEFAULT '',
  `avatar`           VARCHAR(255) NOT NULL DEFAULT '',
  `vip`              INT UNSIGNED NOT NULL DEFAULT 0,
  `mi`               VARCHAR(128) NOT NULL DEFAULT '',
  `region`           VARCHAR(64)  NOT NULL DEFAULT '',
  `register_ip`      VARCHAR(64)  NOT NULL DEFAULT '',
  `last_ip`          VARCHAR(64)  NOT NULL DEFAULT '',
  `last_region`      VARCHAR(64)  NOT NULL DEFAULT '',
  `last_login_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  `bind_limit`       INT          NOT NULL DEFAULT 0,
  `region_limit`     INT          NOT NULL DEFAULT 0,
  `ban_expire`       INT UNSIGNED NOT NULL DEFAULT 0,
  `create_time`      INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_username` (`username`),
  KEY `idx_create` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- 用户拓展信息 (key/value)
DROP TABLE IF EXISTS `{user_extends}`;
CREATE TABLE `{user_extends}` (
  `id`       BIGINT       NOT NULL AUTO_INCREMENT,
  `uid`      BIGINT       NOT NULL,
  `e_key`    VARCHAR(64)  NOT NULL,
  `e_value`  VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- 用户设备
DROP TABLE IF EXISTS `{user_devices}`;
CREATE TABLE `{user_devices}` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `uid`        BIGINT       NOT NULL,
  `mi`         VARCHAR(128) NOT NULL,
  `device`     VARCHAR(128) NOT NULL DEFAULT '',
  `os`         VARCHAR(64)  NOT NULL DEFAULT '',
  `ip`         VARCHAR(64)  NOT NULL DEFAULT '',
  `region`     VARCHAR(64)  NOT NULL DEFAULT '',
  `bind_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  `last_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- 用户日志
DROP TABLE IF EXISTS `{user_logs}`;
CREATE TABLE `{user_logs}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `uid`          BIGINT       NOT NULL,
  `type`         VARCHAR(32)  NOT NULL DEFAULT '',
  `action`       VARCHAR(64)  NOT NULL DEFAULT '',
  `data`         TEXT,
  `ip`           VARCHAR(64)  NOT NULL DEFAULT '',
  `region`       VARCHAR(64)  NOT NULL DEFAULT '',
  `create_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_uid_time` (`uid`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 通知
-- ============================================================
DROP TABLE IF EXISTS `{notices}`;
CREATE TABLE `{notices}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`        BIGINT       NOT NULL DEFAULT 0,
  `title`        VARCHAR(255) NOT NULL,
  `content`      TEXT,
  `enabled`      TINYINT(1)   NOT NULL DEFAULT 1,
  `author`       VARCHAR(64)  NOT NULL DEFAULT '',
  `views`        INT UNSIGNED NOT NULL DEFAULT 0,
  `update_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 应用级拓展信息
-- ============================================================
DROP TABLE IF EXISTS `{extends}`;
CREATE TABLE `{extends}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`        BIGINT       NOT NULL DEFAULT 0,
  `e_key`        VARCHAR(128) NOT NULL,
  `e_value`      VARCHAR(255) NOT NULL DEFAULT '',
  `remark`       VARCHAR(255) NOT NULL DEFAULT '',
  `update_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 版本
-- ============================================================
DROP TABLE IF EXISTS `{versions}`;
CREATE TABLE `{versions}` (
  `id`            BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`         BIGINT       NOT NULL DEFAULT 0,
  `version_name`  VARCHAR(64)  NOT NULL DEFAULT '',
  `version_key`   VARCHAR(64)  NOT NULL DEFAULT '',
  `version_no`    VARCHAR(32)  NOT NULL DEFAULT '0.0.0',
  `status`        TINYINT(1)   NOT NULL DEFAULT 1,
  `url`           VARCHAR(255) NOT NULL DEFAULT '',
  `apk_name`      VARCHAR(255) NOT NULL DEFAULT '',
  `apk_size`      BIGINT       NOT NULL DEFAULT 0,
  `apk_url`       VARCHAR(255) NOT NULL DEFAULT '',
  `changelog`     TEXT,
  `create_time`   INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_appid_key` (`appid`, `version_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 封禁
-- ============================================================
DROP TABLE IF EXISTS `{blocks}`;
CREATE TABLE `{blocks}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`        BIGINT       NOT NULL DEFAULT 0,
  `type`         VARCHAR(16)  NOT NULL,   -- ip / mi / region
  `value`        VARCHAR(255) NOT NULL,
  `reason`       VARCHAR(255) NOT NULL DEFAULT '',
  `create_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid_type` (`appid`, `type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 卡密组 + 卡密
-- ============================================================
DROP TABLE IF EXISTS `{kami_groups}`;
CREATE TABLE `{kami_groups}` (
  `id`       BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`    BIGINT       NOT NULL DEFAULT 0,
  `name`     VARCHAR(128) NOT NULL,
  `type`     VARCHAR(32)  NOT NULL DEFAULT 'vip',
  `value`    BIGINT       NOT NULL DEFAULT 0,
  `remark`   VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

DROP TABLE IF EXISTS `{kamis}`;
CREATE TABLE `{kamis}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`        BIGINT       NOT NULL DEFAULT 0,
  `group_id`     BIGINT       NOT NULL DEFAULT 0,
  `group_name`   VARCHAR(128) NOT NULL DEFAULT '',
  `card`         VARCHAR(128) NOT NULL,
  `status`       TINYINT(1)   NOT NULL DEFAULT 0,  -- 0 未用 / 1 已用 / 2 冻结
  `use_user`     VARCHAR(64)  NOT NULL DEFAULT '',
  `use_time`     INT UNSIGNED NOT NULL DEFAULT 0,
  `create_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_card` (`card`),
  KEY `idx_appid_status` (`appid`, `status`),
  KEY `idx_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 云函数
-- ============================================================
DROP TABLE IF EXISTS `{functions}`;
CREATE TABLE `{functions}` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`      BIGINT       NOT NULL DEFAULT 0,
  `name`       VARCHAR(128) NOT NULL,
  `trigger`    VARCHAR(64)  NOT NULL DEFAULT 'manual',
  `code`       MEDIUMTEXT,
  `remark`     VARCHAR(255) NOT NULL DEFAULT '',
  `timeout`    INT          NOT NULL DEFAULT 3000,
  `enabled`    TINYINT(1)   NOT NULL DEFAULT 1,
  `last_run`   INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid_trigger` (`appid`, `trigger`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 加密方案
-- ============================================================
DROP TABLE IF EXISTS `{encrypts}`;
CREATE TABLE `{encrypts}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `appid`        BIGINT       NOT NULL DEFAULT 0,
  `name`         VARCHAR(128) NOT NULL,
  `algorithm`    VARCHAR(32)  NOT NULL DEFAULT 'aes',
  `encoding`     VARCHAR(32)  NOT NULL DEFAULT 'base64',
  `sym_key`      VARCHAR(255) NOT NULL DEFAULT '',
  `time_skew`    INT          NOT NULL DEFAULT 30,
  `sign_mode`    VARCHAR(32)  NOT NULL DEFAULT 'none',
  `create_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 操作日志 (管理后台)
-- ============================================================
DROP TABLE IF EXISTS `{logs}`;
CREATE TABLE `{logs}` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `type`         VARCHAR(32)  NOT NULL DEFAULT '',
  `action`       VARCHAR(64)  NOT NULL DEFAULT '',
  `data`         TEXT,
  `target`       VARCHAR(128) NOT NULL DEFAULT '',
  `operator`     VARCHAR(64)  NOT NULL DEFAULT '',
  `uid`          BIGINT       NOT NULL DEFAULT 0,
  `ip`           VARCHAR(64)  NOT NULL DEFAULT '',
  `region`       VARCHAR(64)  NOT NULL DEFAULT '',
  `username`     VARCHAR(64)  NOT NULL DEFAULT '',
  `nickname`     VARCHAR(64)  NOT NULL DEFAULT '',
  `avatar`       VARCHAR(255) NOT NULL DEFAULT '',
  `create_time`  INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_type_time` (`type`, `create_time`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 全局设置 (单行)
-- ============================================================
DROP TABLE IF EXISTS `{settings}`;
CREATE TABLE `{settings}` (
  `id`                  BIGINT       NOT NULL AUTO_INCREMENT,
  `web_url`             VARCHAR(255) NOT NULL DEFAULT '',
  `upload_limit`        INT          NOT NULL DEFAULT 10,
  `admin_token_expire`  INT          NOT NULL DEFAULT 604800,
  `admin_entry`         VARCHAR(64)  NOT NULL DEFAULT '',
  `mysql_max_conn`      INT          NOT NULL DEFAULT 50,
  `mysql_max_idle`      INT          NOT NULL DEFAULT 10,
  `cache_enabled`       TINYINT(1)   NOT NULL DEFAULT 1,
  `user_log_enabled`    TINYINT(1)   NOT NULL DEFAULT 1,
  `agent_log_enabled`   TINYINT(1)   NOT NULL DEFAULT 1,
  `admin_log_enabled`   TINYINT(1)   NOT NULL DEFAULT 1,
  `log_level`           VARCHAR(16)  NOT NULL DEFAULT 'info',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

-- ============================================================
-- API 路由 / 错误码 (运行时可配)
-- ============================================================
DROP TABLE IF EXISTS `{api_routes}`;
CREATE TABLE `{api_routes}` (
  `id`       BIGINT       NOT NULL AUTO_INCREMENT,
  `method`   VARCHAR(16)  NOT NULL,
  `path`     VARCHAR(255) NOT NULL,
  `name`     VARCHAR(128) NOT NULL DEFAULT '',
  `group`    VARCHAR(64)  NOT NULL DEFAULT '',
  `enabled`  TINYINT(1)   NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  -- 旧 schema 的 (method, path) 唯一索引会超 InnoDB 767 字节前缀(MySQL 5.6 默认 Antelope 行格式)。
  -- 这张表在 0010_api_rewrite.sql 里会被完全 DROP + 重建为 api_key 模型,这里的索引是过渡值,留 (method) 即可。
  KEY `idx_method` (`method`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

DROP TABLE IF EXISTS `{api_codes}`;
CREATE TABLE `{api_codes}` (
  `id`       BIGINT       NOT NULL AUTO_INCREMENT,
  `code`     INT          NOT NULL,
  `message`  VARCHAR(255) NOT NULL DEFAULT '',
  `remark`   VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;
