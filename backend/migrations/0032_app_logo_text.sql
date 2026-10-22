-- logo 字段原本是 VARCHAR(255),前端把图片转 base64 dataURL 存(data:image/png;base64,xxxx...),
-- 哪怕 5KB 小图 base64 也 7KB+,被截断。改成 MEDIUMTEXT(16MB)允许任意合理大小的 logo。
-- MySQL TEXT 类型不能有默认值,改成 NULL 允许,repo SELECT 加 COALESCE 兜底空串。
ALTER TABLE `{apps}` MODIFY COLUMN `logo` MEDIUMTEXT NULL COMMENT 'app logo dataURL(base64) 或外链 URL';
