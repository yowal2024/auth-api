-- 取消"当前应用 / 全部应用"二选一,通知/加密方案都只绑当前 appid
ALTER TABLE `{notices}`  DROP COLUMN `target`;
ALTER TABLE `{encrypts}` DROP COLUMN `target`;
