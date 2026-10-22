-- 版本表加列:canonical APK + 差异更新开关
ALTER TABLE {versions}
  ADD COLUMN is_canonical TINYINT NOT NULL DEFAULT 0 AFTER apk_url,
  ADD COLUMN canonical_path VARCHAR(255) NOT NULL DEFAULT '' AFTER is_canonical,
  ADD COLUMN clone_diff_enable TINYINT NOT NULL DEFAULT 0 AFTER canonical_path,
  ADD COLUMN original_package VARCHAR(128) NOT NULL DEFAULT '' AFTER clone_diff_enable;
