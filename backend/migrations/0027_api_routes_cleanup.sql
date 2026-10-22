-- 清理 api_routes 里没有 handler 的占位项(wxloginQuery / setPhone / rmPhone /
-- order / pay / messageList / messageReply / vip),只保留实际可用的 SDK 接口。
DELETE FROM `{api_routes}` WHERE api_key IN (
  'wxloginQuery', 'setPhone', 'rmPhone',
  'order', 'pay', 'messageList', 'messageReply', 'vip'
);
