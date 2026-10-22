-- 将系统设置类的历史管理员日志单独归类为 type='system',与 app 操作日志区分。
UPDATE {logs} SET type = 'system'
WHERE type = 'admin'
  AND action IN (
    '修改系统设置',
    '保存 API 路由',
    'API 路由一键重定义',
    '保存 API 状态码',
    'API 状态码一键重定义'
  );
