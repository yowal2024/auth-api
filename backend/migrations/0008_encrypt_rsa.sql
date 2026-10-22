-- 加密表加 RSA 4 个字段
ALTER TABLE `{encrypts}`
  ADD COLUMN `client_private_key` TEXT AFTER `sign_mode`,
  ADD COLUMN `client_public_key`  TEXT AFTER `client_private_key`,
  ADD COLUMN `server_private_key` TEXT AFTER `client_public_key`,
  ADD COLUMN `server_public_key`  TEXT AFTER `server_private_key`;
