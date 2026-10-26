-- Persist application ownership instead of mixing all administrator activity.
ALTER TABLE {logs}
  ADD COLUMN appid BIGINT NOT NULL DEFAULT 0 AFTER id,
  ADD INDEX idx_logs_app_type_time (appid, type, create_time, id);

-- Only recover ownership from recorded application routes. Unknown and global
-- historical events remain in scope 0; never assign them to the selected app.
UPDATE {logs}
SET appid = CASE
  WHEN JSON_VALID(data) THEN CASE
    WHEN JSON_UNQUOTE(JSON_EXTRACT(data, '$.route')) REGEXP '^/api/[^/]+/app/:id(/|$)'
      AND target REGEXP '^[1-9][0-9]{0,17}$'
      THEN CAST(target AS UNSIGNED)
    WHEN JSON_UNQUOTE(JSON_EXTRACT(data, '$.route')) REGEXP '^/api/[^/]+/(user|kami|notice|extend|extend-file|block|encrypt|function|ver|log)(/|$)'
      AND JSON_UNQUOTE(JSON_EXTRACT(data, '$.appId')) REGEXP '^[1-9][0-9]{0,17}$'
      THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(data, '$.appId')) AS UNSIGNED)
    ELSE 0 END
  ELSE 0 END
WHERE type = 'admin';
