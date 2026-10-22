-- Default new applications to the local database; preserve configured providers.
ALTER TABLE `{apps}`
  MODIFY COLUMN `geo_provider` VARCHAR(16) NOT NULL DEFAULT 'dbip' COMMENT 'IP geolocation provider: dbip / tencent / amap';

UPDATE `{apps}` SET `geo_provider` = 'dbip'
WHERE `geo_provider` IN ('', 'tencent')
  AND `bind_region` = 0 AND TRIM(`geo_key`) = '';
