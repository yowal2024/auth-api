-- One geolocation service for all applications; binding rules stay per app.
ALTER TABLE {settings}
  ADD COLUMN geo_provider VARCHAR(16) NOT NULL DEFAULT 'dbip',
  ADD COLUMN geo_key VARCHAR(128) NOT NULL DEFAULT '';

-- Preserve the configured service when existing applications agree. Compare
-- keys case-sensitively. Multiple configurations keep working independently
-- until the administrator explicitly saves a unified system configuration.
UPDATE {settings} s
CROSS JOIN (
  SELECT COUNT(*) AS configurations, MIN(provider) AS provider, MIN(api_key) AS api_key
  FROM (
    SELECT CAST(CASE WHEN geo_provider='' THEN 'tencent' ELSE geo_provider END AS BINARY) AS provider,
           CAST(CASE WHEN geo_provider='dbip' THEN '' ELSE geo_key END AS BINARY) AS api_key
    FROM {apps}
    WHERE bind_region=1 OR (geo_provider IN ('tencent','amap','') AND TRIM(geo_key)<>'')
    GROUP BY provider, api_key
  ) existing
) legacy
SET s.geo_provider = CASE legacy.configurations WHEN 0 THEN 'dbip' WHEN 1 THEN legacy.provider ELSE '' END,
    s.geo_key = CASE WHEN legacy.configurations=1 THEN legacy.api_key ELSE '' END
WHERE s.id=1;
