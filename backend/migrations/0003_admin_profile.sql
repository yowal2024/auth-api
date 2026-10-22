-- 管理员个人资料字段
ALTER TABLE `{admins}`
  ADD COLUMN `nickname` VARCHAR(64)  NOT NULL DEFAULT '' AFTER `username`,
  ADD COLUMN `avatar`   VARCHAR(255) NOT NULL DEFAULT '' AFTER `nickname`;

UPDATE `{admins}` SET `nickname` = '超级管理员' WHERE `id` = 1 AND `nickname` = '';
