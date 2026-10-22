-- 应用维度的 SDK 用户 token 过期时间(秒)。0 不推荐,默认 7 天 = 604800s。
ALTER TABLE `{apps}`
  ADD COLUMN `user_token_ttl` INT NOT NULL DEFAULT 604800 COMMENT 'SDK 用户 token 过期秒数,登录时颁发 + 每次访问滑动续期' AFTER `encryption`;
