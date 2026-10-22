-- users.status 字段冗余,完全由 ban_expire 决定:
--   ban_expire > now() 视为冻结,否则正常
-- 删掉它,后端在 List 时动态计算状态返回给前端
ALTER TABLE `{users}` DROP COLUMN `status`;
