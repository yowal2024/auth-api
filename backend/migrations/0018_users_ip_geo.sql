-- 用户维度的 IP 经纬度绑定。
-- ip_max:允许的历史地理位置数,0=不限;超过则新位置登录需要重新校验
-- ip_distance_max:用户级距离阈值(km),0 时回退 apps.ip_distance_max
-- ip_list:JSON 数组,元素 {ip, lat, lng, time},按 ip_max 滚动保留
ALTER TABLE `{users}`
  ADD COLUMN `ip_max`          INT  NOT NULL DEFAULT 0  COMMENT 'IP 地理位置数上限,0=不限' AFTER `region_limit`,
  ADD COLUMN `ip_distance_max` INT  NOT NULL DEFAULT 0  COMMENT '用户级距离阈值(km),0=用 app 默认' AFTER `ip_max`,
  ADD COLUMN `ip_list`         TEXT NOT NULL            COMMENT '历史 IP 经纬度 JSON 列表' AFTER `ip_distance_max`;
