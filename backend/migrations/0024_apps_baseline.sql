-- 把示例应用清掉、appid 自增从 1000 起。
-- 全新库:0001/0002 已经不再 INSERT 示例应用 + AUTO_INCREMENT=1000,本迁移基本是 no-op。
-- 历史库:删掉早先 seed 进去的 id=1 示例应用,把自增基准对齐到 1000。
--
-- 只匹配未改过的示例应用(pkg='com.example.demo'),避免误删已被用户改名/复用的 id=1 应用。

DELETE FROM `{apps}` WHERE `id` = 1 AND `pkg` = 'com.example.demo';

ALTER TABLE `{apps}` AUTO_INCREMENT = 1000;
