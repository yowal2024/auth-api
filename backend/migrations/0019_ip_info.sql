-- IP → 经纬度缓存,跨 app 共享,避免重复调用第三方 geo 接口。
-- provider:tencent / amap,同一 IP 在不同 provider 下分别缓存
-- update_time:写入时戳,可按业务需要做 TTL 清理
CREATE TABLE IF NOT EXISTS `{ip_info}` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT,
  `ip`          VARCHAR(64)  NOT NULL                COMMENT '客户端 IP',
  `provider`    VARCHAR(16)  NOT NULL DEFAULT 'tencent' COMMENT 'geo 服务商',
  `lat`         DECIMAL(10,6) NOT NULL DEFAULT 0    COMMENT '纬度',
  `lng`         DECIMAL(10,6) NOT NULL DEFAULT 0    COMMENT '经度',
  `nation`      VARCHAR(32)  NOT NULL DEFAULT ''     COMMENT '国家',
  `province`    VARCHAR(32)  NOT NULL DEFAULT ''     COMMENT '省',
  `city`        VARCHAR(32)  NOT NULL DEFAULT ''     COMMENT '市',
  `district`    VARCHAR(32)  NOT NULL DEFAULT ''     COMMENT '区',
  `update_time` INT UNSIGNED NOT NULL DEFAULT 0      COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_ip_provider` (`ip`, `provider`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='IP 经纬度缓存';
