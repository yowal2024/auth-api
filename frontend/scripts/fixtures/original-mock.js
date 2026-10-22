// Recovered demonstration fixtures; used only by local browser tests.
const o = () => Math.floor(Date.now() / 1e3);
const w = e => new Promise(i => setTimeout(i, e));
const a = (e, i) => Math.floor(Math.random() * (i - e + 1)) + e;
const c = e => e[a(0, e.length - 1)];
const k = ["广东省", "北京市", "上海市", "浙江省", "四川省", "日本", "美国", "新加坡"];
const G = () => [{
  id: 1,
  appid: 1,
  name: "普通会员",
  level: 0,
  remark: "默认分组",
  createTime: o() - 90 * 86400
}, {
  id: 2,
  appid: 1,
  name: "黄金会员",
  level: 50,
  remark: "中级会员,享部分高级功能",
  createTime: o() - 60 * 86400
}, {
  id: 3,
  appid: 1,
  name: "钻石会员",
  level: 100,
  remark: "顶级会员,享全部功能",
  createTime: o() - 30 * 86400
}];
const L = [{
  id: 1,
  name: "普通会员",
  level: 0
}, {
  id: 2,
  name: "黄金会员",
  level: 50
}, {
  id: 3,
  name: "钻石会员",
  level: 100
}];
const I = () => {
  const e = [];
  for (let i = 1; i <= 87; i++) {
    const n = i % 5 === 0,
      t = L[i % 3];
    e.push({
      id: i,
      groupId: t.id,
      groupName: t.name,
      groupLevel: t.level,
      username: `user_${1e3 + i}`,
      nickname: "",
      avatar: "",
      vip: i % 3 === 0 ? o() + a(3600, 30 * 86400) : 0,
      mi: i % 2 === 0 ? Array.from({
        length: 32
      }, () => "0123456789abcdef"[a(0, 15)]).join("") : "",
      region: c(k),
      registerIp: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
      lastIp: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
      lastRegion: c(k),
      lastLoginTime: o() - (n ? a(0, 600) : a(3600, 30 * 86400)),
      online: n,
      bindLimit: 0,
      regionLimit: 0,
      banExpire: 0,
      status: i % 13 === 0 ? 0 : 1,
      createTime: o() - a(0, 90 * 86400)
    });
  }
  return e;
};
const A = () => [{
  id: 1,
  name: "默认应用",
  type: "user",
  mode: "free",
  logo: "",
  pkg: "com.example.default",
  tokenKey: "a1b2c3d4e5f60718" + "cafebabe".repeat(2),
  bindDevice: !1,
  bindRegion: !1,
  diffUpdate: !1,
  randomPackage: !1,
  encryption: 0,
  status: 1,
  createTime: o() - 120 * 86400
}, {
  id: 2,
  name: "Demo 安卓客户端",
  type: "user",
  mode: "free",
  logo: "",
  pkg: "com.example.android",
  tokenKey: "11223344556677880011223344556677",
  bindDevice: !0,
  bindRegion: !1,
  diffUpdate: !0,
  randomPackage: !0,
  encryption: 1,
  status: 1,
  createTime: o() - 60 * 86400
}, {
  id: 3,
  name: "时光荚",
  type: "kami",
  mode: "kami",
  logo: "",
  pkg: "com.test.sandbox",
  tokenKey: "abcd1234efgh5678ijkl9012mnop3456",
  bindDevice: !1,
  bindRegion: !1,
  diffUpdate: !1,
  randomPackage: !1,
  encryption: 0,
  status: 1,
  createTime: o() - 20 * 86400
}];
const D = () => [{
  id: 1,
  appid: 1,
  title: "欢迎使用 Auth 平台",
  content: "本系统支持账密、卡密两种登录方式，欢迎使用。",
  enabled: !0,
  author: "admin",
  views: 128,
  target: "current",
  updateTime: o() - 86400
}, {
  id: 2,
  appid: 1,
  title: "维护通知",
  content: "2026-06-01 凌晨 2:00 - 4:00 系统维护，期间服务暂停。",
  enabled: !0,
  author: "admin",
  views: 42,
  target: "all",
  updateTime: o() - 3600
}];
const U = () => [{
  id: 1,
  appid: 1,
  key: "share_url",
  value: "https://share.example.com",
  remark: "推广链接",
  updateTime: o() - 7200
}, {
  id: 2,
  appid: 1,
  key: "kf_qq",
  value: "12345678",
  remark: "客服 QQ",
  updateTime: o() - 86400
}, {
  id: 3,
  appid: 1,
  key: "ad_switch",
  value: "on",
  remark: "广告开关",
  updateTime: o() - 3600
}];
const j = () => [{
  id: 1,
  appid: 1,
  versionName: "安卓版",
  versionKey: "android",
  versionNo: "1.2.0",
  status: 1,
  url: "https://dl.example.com/app-1.2.0.apk",
  apkName: "",
  apkSize: 0,
  apkUrl: "",
  isCanonical: 0,
  canonicalPath: "",
  cloneDiffEnable: 0,
  originalPackage: "",
  changelog: `- 修复登录闪退
- 新增暗色模式`,
  createTime: o() - 3 * 86400
}, {
  id: 2,
  appid: 1,
  versionName: "iOS版",
  versionKey: "ios",
  versionNo: "1.3.0",
  status: 1,
  url: "https://dl.example.com/app-1.3.0.ipa",
  apkName: "",
  apkSize: 0,
  apkUrl: "",
  isCanonical: 0,
  canonicalPath: "",
  cloneDiffEnable: 0,
  originalPackage: "",
  changelog: "- 抢先体验新设置面板",
  createTime: o() - 86400
}];
const R = () => [{
  id: 1,
  appid: 1,
  type: "ip",
  value: "203.0.113.42",
  reason: "多账号撞库",
  createTime: o() - 2 * 86400
}, {
  id: 2,
  appid: 1,
  type: "mi",
  value: "deadbeef" + "cafe".repeat(6),
  reason: "反编译尝试",
  createTime: o() - 86400
}, {
  id: 3,
  appid: 1,
  type: "region",
  value: "朝鲜",
  reason: "不支持地区",
  createTime: o() - 7 * 86400
}];
const M = () => [{
  id: 1,
  appid: 1,
  name: "月卡",
  type: "vip",
  value: 30 * 86400,
  remark: "30 天 VIP"
}, {
  id: 2,
  appid: 1,
  name: "季卡",
  type: "vip",
  value: 90 * 86400,
  remark: "90 天 VIP"
}, {
  id: 3,
  appid: 1,
  name: "年卡",
  type: "vip",
  value: 365 * 86400,
  remark: "365 天 VIP"
}];
const X = () => {
  const e = [];
  for (let i = 1; i <= 120; i++) {
    const n = i % 4 === 0;
    e.push({
      id: i,
      appid: 1,
      groupId: a(1, 3),
      groupName: ["月卡", "季卡", "年卡"][a(0, 2)],
      card: "AUTH-" + Array.from({
        length: 16
      }, () => "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"[a(0, 31)]).join(""),
      status: n ? 1 : 0,
      useUser: n ? `user_${1e3 + a(1, 80)}` : "",
      useTime: n ? o() - a(60, 7 * 86400) : 0,
      createTime: o() - a(0, 30 * 86400)
    });
  }
  return e;
};
const C = () => [{
  id: 1,
  appid: 1,
  name: "登录前 - 限制免费用户登录次数",
  trigger: "before_login",
  code: `async function handler(ctx) {
  const today = await db.count('login_log', {
    uid: ctx.user.id,
    time__gte: Math.floor(Date.now() / 1000 / 86400) * 86400
  })
  if (ctx.user.vip < Math.floor(Date.now() / 1000) && today >= 10) {
    return { allow: false, msg: '免费用户每天最多登录 10 次，请开通 VIP' }
  }
  return { allow: true }
}`,
  remark: "免费用户每日 10 次登录限制",
  timeout: 1e3,
  enabled: !0,
  lastRun: o() - 600
}, {
  id: 2,
  appid: 1,
  name: "注册后 - 通知 Webhook",
  trigger: "after_register",
  code: `async function handler(ctx) {
  await http.post('https://hook.example.com/notify', {
    text: '新用户注册：' + ctx.user.username
  })
}`,
  remark: "",
  timeout: 3e3,
  enabled: !0,
  lastRun: o() - 86400
}, {
  id: 3,
  appid: 1,
  name: "卡密使用 - 通知群组",
  trigger: "kami_use",
  code: `async function handler(ctx) {
  await http.post('https://hook.example.com/notify', {
    text: '卡密 ' + ctx.kami.card + ' 已被 ' + ctx.user.username + ' 使用'
  })
}`,
  remark: "Webhook 通知",
  timeout: 2e3,
  enabled: !1,
  lastRun: 0
}];
const z = () => [{
  id: 1,
  username: "admin",
  role: "super",
  lastLogin: o() - 600,
  lastIp: "127.0.0.1"
}, {
  id: 2,
  username: "ops",
  role: "normal",
  lastLogin: o() - 3 * 86400,
  lastIp: "10.0.0.5"
}];
const K = () => {
  const e = [],
    i = [["通知公告添加", ""], ["通知公告删除", ""], ["版本添加", "mid: 0, name: 安卓版, ver_key: android, ver_major: 1, ver_minor: 0, ver_patch: 2, ver_state: on, ver_new_url: https://yowal.lanzoue.com/iqw0W3py27tg, ver_off_msg: , ver_new_content:"], ["应用编辑", "app_key: 9a9431df4a863d9ca1a4799c01fbf15a, app_logo: , app_mode: n, app_name: 时光荚, app_state: on, app_off_msg:"], ["添加应用", ""], ["修改用户", "uid: 1023, vip_expire: 1748390400, status: 1"], ["生成卡密", "group_id: 1, count: 50, prefix: AUTH-"], ["修改系统设置", "admin_token_expire: 86400, cache_enabled: true"], ["清理缓存", ""]],
    n = {
      login: ["登录", "注册", "退出登录"],
      kami: ["使用卡密", "卡密充值"],
      func: ["触发云函数"]
    };
  for (let u = 1; u <= 60; u++) {
    const m = c(i);
    e.push({
      id: u,
      type: "admin",
      action: m[0],
      data: m[1],
      target: "",
      operator: "admin",
      uid: 0,
      ip: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
      region: c(k),
      createTime: o() - a(0, 30 * 86400)
    });
  }
  let t = 61;
  const d = ["匿名用户", "", "小明", "阿强", "路人甲", "测试号", "匿名用户"];
  return Object.entries(n).forEach(([u, m]) => {
    for (let p = 0; p < 50; p++) {
      const g = 1e9 + a(1, 999999999),
        N = a(1, 9999);
      e.push({
        id: t++,
        type: u,
        action: c(m),
        data: u === "kami" ? "card: AUTH-XXXXXXX, group: 月卡" : "",
        target: "",
        operator: String(g),
        uid: g,
        username: `user_${1e3 + N % 100}`,
        nickname: c(d),
        avatar: "",
        ip: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
        region: c(k),
        createTime: o() - a(0, 30 * 86400)
      });
    }
  }), e.sort((u, m) => m.createTime - u.createTime);
};
const q = () => {
  const e = {},
    i = ["qq", "phone", "email", "invite_code", "channel", "level", "points"];
  return [3, 6, 9, 12, 15, 18, 21, 24].forEach(n => {
    e[n] = i.slice(0, a(2, 5)).map((t, d) => ({
      id: n * 10 + d,
      key: t,
      value: t === "qq" ? String(a(1e7, 999999999)) : t === "phone" ? "138" + a(1e7, 99999999) : t === "email" ? `user${n}@example.com` : t === "invite_code" ? "INV" + a(1e3, 9999) : t === "level" ? String(a(1, 99)) : t === "points" ? String(a(0, 1e4)) : c(["官网", "安卓市场", "苹果商店", "推广链接"])
    }));
  }), e;
};
const W = () => {
  const e = {};
  return [2, 4, 6, 8, 10, 12, 14, 16, 18, 20].forEach(i => {
    const n = a(1, 3);
    e[i] = Array.from({
      length: n
    }, (t, d) => ({
      id: i * 100 + d,
      mi: Array.from({
        length: 32
      }, () => "0123456789abcdef"[a(0, 15)]).join(""),
      device: c(["小米 14 Pro", "华为 Mate60", "iPhone 15", "OPPO Find X7", "vivo X100", "Windows PC", "MacBook Pro"]),
      os: c(["Android 14", "HarmonyOS 4", "iOS 17.4", "Windows 11", "macOS 14"]),
      ip: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
      region: c(k),
      bindTime: o() - a(60, 30 * 86400),
      lastTime: o() - a(0, 7 * 86400)
    }));
  }), e;
};
const H = () => {
  const e = {},
    i = [["login", "登录", ""], ["login", "登录失败", "reason: 密码错误"], ["login", "退出登录", ""], ["kami", "使用卡密", "card: AUTH-XXXXXX, group: 月卡"], ["kami", "使用卡密", "card: AUTH-YYYYYY, group: 季卡"], ["func", "触发云函数", "name: 登录前限制"], ["user", "修改密码", ""], ["user", "修改昵称", "old: 匿名用户, new: 小明"]];
  return [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 15, 18, 21, 24].forEach(n => {
    const t = a(8, 20);
    e[n] = Array.from({
      length: t
    }, (d, u) => {
      const m = c(i);
      return {
        id: n * 1e3 + u,
        type: m[0],
        action: m[1],
        data: m[2],
        ip: `${a(1, 223)}.${a(0, 255)}.${a(0, 255)}.${a(0, 255)}`,
        region: c(k),
        createTime: o() - a(60, 30 * 86400)
      };
    }).sort((d, u) => u.createTime - d.createTime);
  }), e;
};
const V = () => ({
  webUrl: "https://user.yowal.cn",
  uploadLimit: 2,
  adminTokenExpire: 86400,
  adminEntry: "admin",
  mysqlMaxConn: 150,
  mysqlMaxIdle: 20,
  cacheEnabled: !1,
  userLogEnabled: !0,
  adminLogEnabled: !0,
  logLevel: "all"
});
const Y = () => [{
  id: 1,
  method: "POST",
  path: "/api/user/login",
  name: "用户登录",
  group: "user",
  enabled: !0
}, {
  id: 2,
  method: "POST",
  path: "/api/user/register",
  name: "用户注册",
  group: "user",
  enabled: !0
}, {
  id: 3,
  method: "POST",
  path: "/api/user/logout",
  name: "用户登出",
  group: "user",
  enabled: !0
}, {
  id: 4,
  method: "GET",
  path: "/api/user/info",
  name: "获取用户信息",
  group: "user",
  enabled: !0
}, {
  id: 5,
  method: "POST",
  path: "/api/kami/use",
  name: "使用卡密",
  group: "kami",
  enabled: !0
}, {
  id: 6,
  method: "POST",
  path: "/api/kami/verify",
  name: "验证卡密",
  group: "kami",
  enabled: !0
}, {
  id: 7,
  method: "GET",
  path: "/api/app/version",
  name: "获取应用版本",
  group: "app",
  enabled: !0
}, {
  id: 8,
  method: "GET",
  path: "/api/app/notice",
  name: "获取应用公告",
  group: "app",
  enabled: !0
}];
const B = () => [{
  id: 1,
  code: 0,
  message: "操作成功",
  remark: "通用成功响应"
}, {
  id: 2,
  code: 400,
  message: "参数错误",
  remark: "请求参数缺失或不合法"
}, {
  id: 3,
  code: 401,
  message: "未授权",
  remark: "Token 无效或已过期"
}, {
  id: 4,
  code: 403,
  message: "禁止访问",
  remark: "权限不足"
}, {
  id: 5,
  code: 404,
  message: "资源不存在",
  remark: "请求的资源未找到"
}, {
  id: 6,
  code: 429,
  message: "请求过于频繁",
  remark: "触发限流"
}, {
  id: 7,
  code: 500,
  message: "服务器内部错误",
  remark: "后端异常"
}, {
  id: 8,
  code: 1001,
  message: "账号或密码错误",
  remark: "登录验证失败"
}, {
  id: 9,
  code: 1002,
  message: "账号已被封禁",
  remark: "用户状态异常"
}, {
  id: 10,
  code: 2001,
  message: "卡密不存在",
  remark: "卡密校验失败"
}, {
  id: 11,
  code: 2002,
  message: "卡密已被使用",
  remark: "卡密已激活"
}];
const F = () => [{
  id: 1,
  appid: 1,
  name: "默认方案",
  algorithm: "aes",
  encoding: "base64",
  symKey: "1234567890abcdef",
  timeSkew: 60,
  signMode: "verify",
  target: "current",
  createTime: o() - 86400
}, {
  id: 2,
  appid: 1,
  name: "调试方案",
  algorithm: "rsa",
  timeSkew: 120,
  signMode: "none",
  target: "current",
  createTime: o() - 3600
}];
const r = {
  userGroups: G(),
  users: I(),
  apps: A(),
  notices: D(),
  extends_: U(),
  versions: j(),
  blocks: R(),
  kamiGroups: M(),
  kamis: X(),
  functions: C(),
  admins: z(),
  logs: K(),
  settings: V(),
  encrypts: F(),
  apiRoutes: Y(),
  apiCodes: B(),
  userExtends: q(),
  userDevices: W(),
  userLogs: H()
};
const s = e => ({
  code: 0,
  msg: "ok",
  data: e
});
const l = e => Math.max(0, ...e.map(i => i.id)) + 1;
const T = (e, i = 1, n = 20) => {
  const t = (i - 1) * n;
  return {
    list: e.slice(t, t + n),
    total: e.length
  };
};
const f = (e, i, n) => {
  let t = e;
  if (i.keyword) {
    const d = String(i.keyword).toLowerCase();
    t = t.filter(u => n.some(m => String(u[m] || "").toLowerCase().includes(d)));
  }
  return i.status !== void 0 && i.status !== null && i.status !== "" && (t = t.filter(d => d.status === Number(i.status))), i.appid && (t = t.filter(d => !d.appid || d.appid === Number(i.appid))), t;
};
const Q = [["POST", "/admin/login", ({
  data: e
}) => e.username === "admin" && e.password === "123456" ? s({
  token: "mock-token-" + Date.now(),
  profile: {
    id: 1,
    username: "admin",
    role: "super"
  }
}) : {
  code: 401,
  msg: "账号或密码错误"
}], ["POST", "/admin/logout", () => s(null)], ["GET", "/admin/me", () => s({
  id: 1,
  username: "admin",
  role: "super"
})], ["GET", "/admin/dashboard", ({
  params: e
}) => {
  const i = Math.max(1, Number(e?.range) || 7),
    n = [],
    t = [],
    d = [];
  for (let u = i - 1; u >= 0; u--) {
    const m = new Date();
    m.setDate(m.getDate() - u), n.push(`${m.getMonth() + 1}/${m.getDate()}`), t.push(a(80, 320)), d.push(a(5, 50));
  }
  return s({
    users: r.users.length,
    usersToday: 7,
    apps: r.apps.length,
    appsActive: r.apps.filter(u => u.status === 1).length,
    kamiAvail: r.kamis.filter(u => u.status === 0).length,
    kamiUsedToday: 4,
    loginsToday: t[t.length - 1],
    loginSuccessRate: 97,
    onlineNow: a(120, 480),
    onlinePeakToday: a(500, 900),
    trend: {
      dates: n,
      logins: t,
      regs: d
    },
    recent: [{
      time: "刚刚",
      type: "primary",
      text: "admin 修改用户 user_1023 的 VIP 到期时间"
    }, {
      time: "10 分钟前",
      type: "success",
      text: "user_1042 登录成功"
    }, {
      time: "30 分钟前",
      type: "warning",
      text: "云函数「登录前限制」拦截 user_1099"
    }, {
      time: "1 小时前",
      type: "",
      text: "admin 生成 50 个月卡卡密"
    }]
  });
}], ["GET", "/admin/user/group/list", () => s(r.userGroups)], ["POST", "/admin/user/group", ({
  data: e
}) => {
  const i = String(e.name || "").trim();
  if (!i) return {
    code: 1,
    msg: "分组名称不能为空"
  };
  const n = Math.max(0, Number(e.level) || 0);
  if (e.id) {
    const d = r.userGroups.find(u => u.id === e.id);
    return d && Object.assign(d, {
      name: i,
      level: n,
      remark: e.remark || ""
    }), s(d);
  }
  const t = {
    id: l(r.userGroups),
    appid: 1,
    name: i,
    level: n,
    remark: e.remark || "",
    createTime: o()
  };
  return r.userGroups.push(t), s(t);
}], ["DELETE", /^\/admin\/user\/group\/(\d+)$/, ({
  match: e
}) => {
  const i = Number(e[1]);
  return r.users.some(n => n.groupId === i) ? {
    code: 1008,
    msg: "该分组下还有用户,无法删除"
  } : (r.userGroups = r.userGroups.filter(n => n.id !== i), s(null));
}], ["GET", "/admin/user/list", ({
  params: e
}) => {
  let i = r.users;
  if (e.keyword) {
    const n = String(e.keyword).toLowerCase(),
      t = e.searchField || "username",
      d = t === "fuzzy" ? ["username", "nickname", "id"] : [t];
    i = i.filter(u => d.some(m => String(u[m] ?? "").toLowerCase().includes(n)));
  }
  if (e.status !== void 0 && e.status !== "" && e.status !== null && (i = i.filter(n => n.status === Number(e.status))), e.userType) {
    const n = o();
    e.userType === "vip" ? i = i.filter(t => t.vip > n) : e.userType === "free" && (i = i.filter(t => !t.vip || t.vip <= n));
  }
  if (e.groupId) {
    const n = Number(e.groupId);
    i = i.filter(t => t.groupId === n);
  }
  return s(T(i, e.page, e.size));
}], ["GET", /^\/admin\/user\/(\d+)$/, ({
  match: e
}) => {
  const i = r.users.find(n => n.id === Number(e[1]));
  return s(i);
}], ["POST", "/admin/user", ({
  data: e
}) => {
  if (!e.groupId) return {
    code: 1,
    msg: "请选择用户分组"
  };
  const i = r.userGroups.find(t => t.id === Number(e.groupId));
  if (!i) return {
    code: 1004,
    msg: "用户分组不存在"
  };
  const n = {
    id: l(r.users),
    groupId: i.id,
    groupName: i.name,
    groupLevel: i.level,
    username: e.username,
    mi: "",
    region: "",
    vip: 0,
    status: e.status ?? 1,
    createTime: o()
  };
  return r.users.unshift(n), s(n);
}], ["PUT", /^\/admin\/user\/(\d+)$/, ({
  match: e,
  data: i
}) => {
  const n = r.users.find(p => p.id === Number(e[1]));
  if (!n) return s(null);
  if (i.groupId !== void 0) {
    const p = r.userGroups.find(g => g.id === Number(i.groupId));
    if (!p) return {
      code: 1004,
      msg: "用户分组不存在"
    };
    n.groupId = p.id, n.groupName = p.name, n.groupLevel = p.level;
  }
  const {
    groupId: t,
    groupName: d,
    groupLevel: u,
    ...m
  } = i;
  return Object.assign(n, m), s(n);
}], ["DELETE", /^\/admin\/user\/(\d+)$/, ({
  match: e
}) => (r.users = r.users.filter(i => i.id !== Number(e[1])), s(null))], ["POST", "/admin/user/batch-delete", ({
  data: e
}) => {
  const i = (e.ids || []).map(Number);
  return r.users = r.users.filter(n => !i.includes(n.id)), s({
    affected: i.length
  });
}], ["POST", /^\/admin\/user\/(\d+)\/password$/, () => s(null)], ["POST", /^\/admin\/user\/(\d+)\/unbind-mi$/, ({
  match: e
}) => {
  const i = r.users.find(n => n.id === Number(e[1]));
  return i && (i.mi = ""), s(null);
}], ["GET", /^\/admin\/user\/(\d+)\/extends$/, ({
  match: e
}) => {
  const i = Number(e[1]);
  return s(r.userExtends[i] || []);
}], ["POST", /^\/admin\/user\/(\d+)\/extends$/, ({
  match: e,
  data: i
}) => {
  const n = Number(e[1]);
  if (Array.isArray(i.list)) {
    const d = r.userExtends[n] || [];
    let u = Math.max(0, ...d.map(m => m.id));
    return r.userExtends[n] = i.list.map(m => ({
      id: m.id || ++u,
      key: m.key,
      value: m.value
    })), s(null);
  }
  r.userExtends[n] || (r.userExtends[n] = []);
  const t = r.userExtends[n];
  if (i.id) {
    const d = t.find(u => u.id === i.id);
    d && Object.assign(d, i);
  } else t.push({
    id: l(t),
    key: i.key,
    value: i.value
  });
  return s(null);
}], ["DELETE", /^\/admin\/user\/(\d+)\/extends\/(\d+)$/, ({
  match: e
}) => {
  const i = Number(e[1]),
    n = Number(e[2]);
  return r.userExtends[i] && (r.userExtends[i] = r.userExtends[i].filter(t => t.id !== n)), s(null);
}], ["GET", /^\/admin\/user\/(\d+)\/devices$/, ({
  match: e
}) => {
  const i = Number(e[1]);
  return s(r.userDevices[i] || []);
}], ["DELETE", /^\/admin\/user\/(\d+)\/devices\/(\d+)$/, ({
  match: e
}) => {
  const i = Number(e[1]),
    n = Number(e[2]);
  return r.userDevices[i] && (r.userDevices[i] = r.userDevices[i].filter(t => t.id !== n)), s(null);
}], ["GET", /^\/admin\/user\/(\d+)\/logs$/, ({
  match: e,
  params: i
}) => {
  const n = Number(e[1]),
    t = r.userLogs[n] || [];
  return s(T(t, i.page, i.size));
}], ["POST", "/admin/user/reward", ({
  data: e
}) => {
  const i = Number(e.value || 0) * (e.unit === "year" ? 365 : e.unit === "month" ? 30 : 1) * 86400,
    n = e.target === "all" ? r.users : r.users.filter(t => (e.userIds || []).includes(t.id));
  return n.forEach(t => {
    t.vip = Math.max(t.vip || 0, o()) + i;
  }), s({
    affected: n.length
  });
}], ["GET", "/admin/app/list", () => s(r.apps)], ["POST", "/admin/app", ({
  data: e
}) => {
  const i = {
    ...e,
    id: l(r.apps),
    createTime: o(),
    tokenKey: e.tokenKey || Array.from({
      length: 32
    }, () => "0123456789abcdef"[a(0, 15)]).join("")
  };
  return r.apps.push(i), s(i);
}], ["PUT", /^\/admin\/app\/(\d+)$/, ({
  match: e,
  data: i
}) => {
  const n = r.apps.find(t => t.id === Number(e[1]));
  return n && Object.assign(n, i), s(n);
}], ["DELETE", /^\/admin\/app\/(\d+)$/, ({
  match: e
}) => (r.apps = r.apps.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/notice/list", ({
  params: e
}) => s(f(r.notices, e, ["title"]))], ["POST", "/admin/notice", ({
  data: e
}) => {
  if (e.id) {
    const i = r.notices.find(n => n.id === e.id);
    i && Object.assign(i, e, {
      updateTime: o()
    });
  } else r.notices.unshift({
    author: "admin",
    views: 0,
    target: "current",
    enabled: !0,
    ...e,
    id: l(r.notices),
    updateTime: o()
  });
  return s(null);
}], ["DELETE", /^\/admin\/notice\/(\d+)$/, ({
  match: e
}) => (r.notices = r.notices.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/extend/list", ({
  params: e
}) => s(f(r.extends_, e, ["key", "value", "remark"]))], ["POST", "/admin/extend", ({
  data: e
}) => {
  if (e.id) {
    const i = r.extends_.find(n => n.id === e.id);
    i && Object.assign(i, e, {
      updateTime: o()
    });
  } else r.extends_.unshift({
    ...e,
    id: l(r.extends_),
    updateTime: o()
  });
  return s(null);
}], ["DELETE", /^\/admin\/extend\/(\d+)$/, ({
  match: e
}) => (r.extends_ = r.extends_.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/ver/list", ({
  params: e
}) => {
  const i = f(r.versions, e, ["versionName", "versionKey"]);
  return s(i);
}], ["POST", "/admin/ver", ({
  data: e
}) => {
  if (e.id) {
    const i = r.versions.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.versions.unshift({
    ...e,
    id: l(r.versions),
    createTime: o()
  });
  return s(null);
}], ["DELETE", /^\/admin\/ver\/(\d+)$/, ({
  match: e
}) => (r.versions = r.versions.filter(i => i.id !== Number(e[1])), s(null))], ["POST", "/admin/ver/upload-apk", () => {
  const e = Math.random().toString(16).slice(2, 18).padEnd(64, "0");
  return s({
    path: `./uploads/canonical/${e}.apk`,
    sha256: e,
    originalPackage: "me.kernel.deltaforce",
    size: 5242880,
    url: `/uploads/canonical/${e}.apk`,
    name: "canonical.apk"
  });
}], ["GET", "/admin/encrypt/list", ({
  params: e
}) => s(f(r.encrypts, e, ["name"]))], ["POST", "/admin/encrypt", ({
  data: e
}) => {
  if (e.id) {
    const i = r.encrypts.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.encrypts.unshift({
    ...e,
    id: l(r.encrypts),
    createTime: o()
  });
  return s(null);
}], ["DELETE", /^\/admin\/encrypt\/(\d+)$/, ({
  match: e
}) => (r.encrypts = r.encrypts.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/block/list", ({
  params: e
}) => {
  let i = f(r.blocks, e, ["value", "reason"]);
  return e.type && (i = i.filter(n => n.type === e.type)), s(i);
}], ["POST", "/admin/block", ({
  data: e
}) => (r.blocks.unshift({
  ...e,
  id: l(r.blocks),
  createTime: o()
}), s(null))], ["DELETE", /^\/admin\/block\/(\d+)$/, ({
  match: e
}) => (r.blocks = r.blocks.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/kami/group/list", () => s(r.kamiGroups)], ["POST", "/admin/kami/group", ({
  data: e
}) => {
  if (e.id) {
    const i = r.kamiGroups.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.kamiGroups.push({
    ...e,
    id: l(r.kamiGroups)
  });
  return s(null);
}], ["DELETE", /^\/admin\/kami\/group\/(\d+)$/, ({
  match: e
}) => (r.kamiGroups = r.kamiGroups.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/kami/list", ({
  params: e
}) => {
  let i = f(r.kamis, e, ["card", "useUser"]);
  return e.groupId && (i = i.filter(n => n.groupId === Number(e.groupId))), s(T(i, e.page, e.size));
}], ["POST", "/admin/kami/generate", ({
  data: e
}) => {
  const i = r.kamiGroups.find(n => n.id === e.groupId);
  for (let n = 0; n < e.count; n++) {
    const t = (e.prefix || "") + Array.from({
      length: e.length || 16
    }, () => "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"[a(0, 31)]).join("");
    r.kamis.unshift({
      id: l(r.kamis),
      appid: e.appid,
      groupId: e.groupId,
      groupName: i?.name || "",
      card: t,
      status: 0,
      useUser: "",
      useTime: 0,
      createTime: o()
    });
  }
  return s(null);
}], ["DELETE", /^\/admin\/kami\/(\d+)$/, ({
  match: e
}) => (r.kamis = r.kamis.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/kami/export", () => s({
  url: "mock://export.csv"
})], ["GET", "/admin/function/list", ({
  params: e
}) => {
  let i = f(r.functions, e, ["name", "remark"]);
  return e.trigger && (i = i.filter(n => n.trigger === e.trigger)), s(i);
}], ["GET", /^\/admin\/function\/(\d+)$/, ({
  match: e
}) => {
  const i = Number(e[1]);
  return s(r.functions.find(n => n.id === i));
}], ["POST", "/admin/function", ({
  data: e
}) => {
  if (e.id) {
    const i = r.functions.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.functions.unshift({
    ...e,
    id: l(r.functions),
    lastRun: 0
  });
  return s(null);
}], ["DELETE", /^\/admin\/function\/(\d+)$/, ({
  match: e
}) => (r.functions = r.functions.filter(i => i.id !== Number(e[1])), s(null))], ["POST", /^\/admin\/function\/(\d+)\/run$/, () => s({
  result: {
    ok: !0,
    msg: "hello from cloud function (mock)"
  },
  durationMs: a(3, 80)
})], ["GET", "/admin/admin/list", () => s(r.admins)], ["POST", "/admin/admin", ({
  data: e
}) => {
  if (e.id) {
    const i = r.admins.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.admins.push({
    ...e,
    id: l(r.admins),
    lastLogin: 0,
    lastIp: ""
  });
  return s(null);
}], ["DELETE", /^\/admin\/admin\/(\d+)$/, ({
  match: e
}) => (r.admins = r.admins.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/system/settings", () => s(r.settings)], ["POST", "/admin/system/settings", ({
  data: e
}) => (Object.assign(r.settings, e), s(null))], ["GET", "/admin/system/api-routes", ({
  params: e
}) => s(f(r.apiRoutes, e, ["path", "name"]))], ["POST", "/admin/system/api-routes", ({
  data: e
}) => {
  if (e.id) {
    const i = r.apiRoutes.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.apiRoutes.push({
    ...e,
    id: l(r.apiRoutes)
  });
  return s(null);
}], ["DELETE", /^\/admin\/system\/api-routes\/(\d+)$/, ({
  match: e
}) => (r.apiRoutes = r.apiRoutes.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/system/api-codes", ({
  params: e
}) => s(f(r.apiCodes, e, ["message", "remark"]))], ["POST", "/admin/system/api-codes", ({
  data: e
}) => {
  if (e.id) {
    const i = r.apiCodes.find(n => n.id === e.id);
    i && Object.assign(i, e);
  } else r.apiCodes.push({
    ...e,
    id: l(r.apiCodes)
  });
  return s(null);
}], ["DELETE", /^\/admin\/system\/api-codes\/(\d+)$/, ({
  match: e
}) => (r.apiCodes = r.apiCodes.filter(i => i.id !== Number(e[1])), s(null))], ["GET", "/admin/log/list", ({
  params: e
}) => {
  let i = f(r.logs, e, ["action", "data", "operator", "ip", "username", "nickname"]);
  return e.type && (i = i.filter(n => n.type === e.type)), e.category === "user" ? i = i.filter(n => n.type !== "admin") : e.category === "admin" && (i = i.filter(n => n.type === "admin")), e.action && (i = i.filter(n => n.action === e.action)), e.start && (i = i.filter(n => n.createTime >= Number(e.start))), e.end && (i = i.filter(n => n.createTime <= Number(e.end))), s(T(i, e.page, e.size));
}], ["GET", "/admin/log/actions", ({
  params: e
}) => {
  let i = r.logs;
  e.category === "user" ? i = i.filter(t => t.type !== "admin") : e.category === "admin" && (i = i.filter(t => t.type === "admin"));
  const n = new Set(i.map(t => t.action));
  return s([...n]);
}], ["POST", "/admin/log/clear", ({
  data: e
}) => {
  const i = e.ids || [];
  return i.length ? r.logs = r.logs.filter(n => !i.includes(n.id)) : e.category === "admin" ? r.logs = r.logs.filter(n => n.type !== "admin") : e.category === "user" ? r.logs = r.logs.filter(n => n.type === "admin") : e.type ? r.logs = r.logs.filter(n => n.type !== e.type) : r.logs = [], s(null);
}]];
async function Z(e, i = {}) {
  await w(120 + Math.random() * 180);
  const n = (i.method || "GET").toUpperCase(),
    t = i.params || {},
    d = i.data || {};
  for (const [u, m, p] of Q) if (u === n) if (typeof m == "string") {
    if (m === e) return E(p({
      params: t,
      data: d
    }));
  } else {
    const g = m.exec(e);
    if (g) return E(p({
      params: t,
      data: d,
      match: g
    }));
  }
  return console.warn(`[mock] no route matched: ${n} ${e}`), Promise.reject(new Error(`接口未实现: ${e}`));
}
const E = e => e?.code === 0 || e?.code === 200 ? Promise.resolve(e.data) : Promise.reject(new Error(e?.msg || "fail"));
export { Z as mockRequest };
