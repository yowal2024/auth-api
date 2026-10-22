-- 卡密生成时记录创建人(管理员用户名)和创建 IP,便于审计与抽屉展示。
ALTER TABLE `{kamis}`
  ADD COLUMN `creator`   VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN `create_ip` VARCHAR(64) NOT NULL DEFAULT '';
