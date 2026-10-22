-- 合并字段:region_limit 复用为"IP 地点数上限",废弃独立的 ip_max / ip_distance_max。
-- ip_list(用户历史经纬度 JSON)保留。
ALTER TABLE `{users}`
  DROP COLUMN `ip_max`,
  DROP COLUMN `ip_distance_max`;
