-- 扩展文件管理:支持上传zip/data等文件作为扩展资源
CREATE TABLE `{extend_files}` (
  `id`          BIGINT        NOT NULL AUTO_INCREMENT,
  `appid`       BIGINT        NOT NULL DEFAULT 0 COMMENT '所属应用,0表示全局',
  `name`        VARCHAR(128)  NOT NULL COMMENT '文件显示名称',
  `filename`    VARCHAR(255)  NOT NULL COMMENT '存储的文件名',
  `filepath`    VARCHAR(512)  NOT NULL COMMENT '文件路径',
  `filesize`    BIGINT        NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `filetype`    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '文件类型(MIME或扩展名)',
  `category`    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '分类标签',
  `remark`      VARCHAR(255)  NOT NULL DEFAULT '',
  `version`     VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '版本号',
  `download_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '下载URL',
  `download_count` BIGINT      NOT NULL DEFAULT 0 COMMENT '下载次数',
  `created_at`  BIGINT        NOT NULL DEFAULT 0,
  `updated_at`  BIGINT        NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_appid_category` (`appid`, `category`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='扩展文件资源管理';
