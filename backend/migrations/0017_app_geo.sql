-- 应用维度的 IP 经纬度绑定配置。
-- geo_provider: tencent / amap;geo_key: 对应平台的应用 key
-- ip_distance_max: 默认距离阈值(km),用户级未配置时回退此值
ALTER TABLE `{apps}`
  ADD COLUMN `geo_provider`    VARCHAR(16)  NOT NULL DEFAULT 'tencent' COMMENT 'IP 地理位置查询服务: tencent / amap' AFTER `bind_region`,
  ADD COLUMN `geo_key`         VARCHAR(128) NOT NULL DEFAULT ''        COMMENT '上述 provider 的应用 key'              AFTER `geo_provider`,
  ADD COLUMN `ip_distance_max` INT          NOT NULL DEFAULT 50        COMMENT 'IP 距离阈值(km),小于此值视为同地点'   AFTER `geo_key`;
