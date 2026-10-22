-- 卡密版用户:不另起表,直接把"用户态"字段挂到 kamis 上。
--   status=0  → 未激活(纯卡密池行)
--   status=1  → 已激活,本行同时充当 SDK 用户实体
--   status=2  → 冻结(发卡时即冻结,登录拒绝)
-- 字段说明同 users:vip/ban_expire/bind_limit/region_limit/ip_list/mi/region/nickname/last_ip/last_login_time
-- 复用 use_user 字段保存登录卡号(发卡时为空,激活时写 card),use_time 当作 register_time/激活时间。

ALTER TABLE `{kamis}`
  ADD COLUMN `nickname`         VARCHAR(64)  NOT NULL DEFAULT '',
  ADD COLUMN `vip`              INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN `ban_expire`       INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN `bind_limit`       INT          NOT NULL DEFAULT 0,
  ADD COLUMN `region_limit`     INT          NOT NULL DEFAULT 0,
  ADD COLUMN `ip_list`          TEXT         NULL,
  ADD COLUMN `mi`               VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN `region`           VARCHAR(64)  NOT NULL DEFAULT '',
  ADD COLUMN `last_ip`          VARCHAR(64)  NOT NULL DEFAULT '',
  ADD COLUMN `last_login_time`  INT UNSIGNED NOT NULL DEFAULT 0;
