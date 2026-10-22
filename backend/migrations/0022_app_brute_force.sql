-- 登录防爆破开关:同一用户名 5 分钟内连续失败 5 次,锁 5 分钟。
-- 阈值/锁定时长固定在代码常量里,管理员只控开关。
ALTER TABLE `{apps}`
  ADD COLUMN `brute_force_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '登录防爆破: 1=启用 5次/5分钟锁' AFTER `ip_distance_max`;

-- 防爆破触发码 + token 链路新增的拒绝码
INSERT INTO `{api_codes}` (code, message, remark) VALUES
  (1017, '登录失败次数过多,请稍后再试', '同一用户名 5 分钟内失败 ≥5 次触发'),
  (1018, 'VIP 已到期,请续费后再使用',     'mode=vip 的应用,token 中间件拦截过期会员')
ON DUPLICATE KEY UPDATE message = VALUES(message), remark = VALUES(remark);
