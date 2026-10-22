-- Permanent bans use -1. Preserve existing timestamps while allowing that value.
ALTER TABLE `{users}` MODIFY COLUMN `ban_expire` BIGINT NOT NULL DEFAULT 0;
ALTER TABLE `{kamis}` MODIFY COLUMN `ban_expire` BIGINT NOT NULL DEFAULT 0;
