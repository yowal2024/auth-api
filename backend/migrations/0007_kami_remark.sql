-- 卡密增加备注字段(同一批生成共用一个备注)
ALTER TABLE `{kamis}`
  ADD COLUMN `remark` VARCHAR(255) NOT NULL DEFAULT '' AFTER `use_time`;
