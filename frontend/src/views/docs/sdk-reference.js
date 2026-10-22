// Field names and limits follow the current Go SDK handlers.
const profile = [
  "user",
  "object",
  "用户资料，见「用户返回字段」；卡密账号增加 card、isKami、regionLimit。",
];
const serverTime = ["serverTime", "integer", "服务器 Unix 时间戳（秒）。"];
const mi = [
  "mi / JSON",
  "string",
  "条件必填",
  "启用设备绑定时必填；后续鉴权接口通过 X-Mi 发送同一机器码。",
];
export const sdkDetails = {
  logon: {
    params: [
      [
        "username / JSON",
        "string",
        "账号版必填",
        "账号名称；卡密版也接受此字段作为 card 的后备值。",
      ],
      ["password / JSON", "string", "账号版必填", "账号密码；卡密版不需要。"],
      [
        "card / JSON",
        "string",
        "卡密版必填其一",
        "与 username 二选一；两者都有时优先 card。已作为充值券消费的卡号不能再登录（2002）；冻结返回 2003。",
      ],
      mi,
    ],
    returns: [
      ["token", "string", "当前应用客户端登录令牌，不能用于管理接口。"],
      profile,
      serverTime,
    ],
  },
  register: {
    params: [
      [
        "username / JSON",
        "string",
        "是",
        "5–12 个 UTF-8 字节，建议使用 ASCII 账号。",
      ],
      ["password / JSON", "string", "是", "6–18 个 UTF-8 字节。"],
      [
        "nickname / JSON",
        "string",
        "否",
        "昵称，省略为空；建议不超过 32 字，与修改资料保持一致。",
      ],
      mi,
    ],
    returns: [
      ["token", "string", "注册成功即返回登录令牌；不会自动赠送 VIP。"],
      profile,
    ],
  },
  logout: {
    returns: [
      ["ok", "boolean", "true 表示当前令牌已注销；其他设备或会话保持有效。"],
    ],
  },
  heartbeat: {
    returns: [["ok", "boolean", "true 表示本次鉴权成功。"], serverTime],
  },
  info: { returns: [profile, serverTime] },
  updateInfo: {
    params: [
      [
        "nickname / JSON",
        "string",
        "至少提供一项",
        "最多 32 个 Unicode 字符；空字符串清空；省略或 null 不修改。",
      ],
      [
        "avatar / JSON",
        "string",
        "至少提供一项",
        "URL 最多 255 个 UTF-8 字节；空字符串清空；卡密账号传此项返回 400。",
      ],
    ],
    returns: [profile, serverTime],
  },
  userExtends: {
    params: [
      [
        "set / JSON",
        "object<string,string>",
        "否",
        "省略时读取；写入键名 1–64 个 UTF-8 字节且不能全为空白，值必须为字符串；同名覆盖。空字符串是值，不代表删除。",
      ],
    ],
    reqExamples: [
      { label: "读取", code: "{}" },
      {
        label: "写入多个键",
        code: JSON.stringify(
          {
            set: {
              theme: "blue",
              preferences: JSON.stringify({ sound: true }),
            },
          },
          null,
          2,
        ),
      },
    ],
    returns: [
      [
        "extends",
        "object<string,string>",
        "当前账号全部扩展键值；无数据时 {}。普通账号和卡密账号互相隔离。",
      ],
    ],
    notes: [
      "这些值允许当前客户端修改，不要将会员、角色或授权开关等可信权限保存在这里。当前没有客户端删除键接口。",
    ],
  },
  kamiTopup: {
    params: [
      [
        "card / JSON",
        "string",
        "是",
        "同一应用下未使用、未冻结的充值卡；仅接受 vip 类型且面值大于零。",
      ],
    ],
    returns: [
      ["vip", "integer", "max(原到期时间, 当前时间) + 卡密秒数。"],
      ["vipActive", "boolean", "新到期时间是否晚于当前时间。"],
      ["groupName", "string", "充值卡分组名称，不是用户分组。"],
      ["addSec", "integer", "本次充值增加的秒数。"],
    ],
    notes: [
      "支持普通账号和已登录的卡密账号。VIP 过期但仍有有效 token 时可充值或登出；过期后重新登录可能被 VIP 模式拒绝。",
      "卡密模式中，充值使用另一张未用卡；消费后不能再用该充值卡直接登录。冻结、解冻不会恢复已消费卡的充值次数。",
      "请求超时后先查询账户或联系后台核实，避免直接换一张卡重试造成重复充值。",
    ],
  },
  uploadAvatar: {
    params: [
      [
        "file / multipart",
        "binary",
        "是",
        "PNG/JPEG/GIF/WEBP；以实际内容识别，最多 5 MiB。让 HTTP 库生成 multipart boundary。",
      ],
    ],
    returns: [
      [
        "avatar",
        "string",
        "已写入当前用户的头像 URL；相对地址需拼接服务端源站。",
      ],
      profile,
      serverTime,
    ],
    notes: [
      "不签名、不加密；仍要求应用校验、限流、登录鉴权，以及按配置启用的设备、地域、VIP 和 mTLS 检查。",
    ],
  },
  cloudFunction: {
    params: [
      ["name / JSON", "string", "是", "当前应用内已启用的 manual 函数名。"],
      [
        "param / JSON",
        "任意 JSON",
        "否",
        "传入 ctx.payload；对象推荐传 {}。省略时为 null。",
      ],
    ],
    returns: [
      [
        "data 本身",
        "任意 JSON",
        "handler 的返回值；不包裹 result、logs、duration。",
      ],
    ],
  },
  ini: {
    params: [
      [
        "versionKey / JSON",
        "string",
        "否",
        "客户端已安装版本的索引，与 versionNo 配合比较。",
      ],
      [
        "versionNo / JSON",
        "string",
        "否",
        "客户端已安装版本号；省略则无法生成升级差分。",
      ],
    ],
    returns: [
      [
        "notices",
        "array<object>",
        "启用的公告：id、title、content、author、updateTime（秒）。",
      ],
      ["extends", "object<string,string>", "当前应用的扩展变量。"],
      [
        "latestVersion",
        "object | null",
        "无已发布版本时为 null；按 create_time、id 倒序选最新已发布版本，不按请求的 versionKey 筛选。",
      ],
      [
        "latestVersion.versionName / versionKey / versionNo",
        "string",
        "版本名称、索引和版本号。",
      ],
      [
        "latestVersion.apkUrl / url / changelog",
        "string",
        "APK 地址、外部地址及更新说明；可能为空。",
      ],
      [
        "latestVersion.hasUpdate",
        "boolean",
        "versionKey 或 versionNo 任一不相同；不是语义版本的大小比较。",
      ],
      [
        "latestVersion.diffUrl / diffOldSha / diffNewSha",
        "string（可选）",
        "可生成升级补丁时提供；旧包和新包的 SHA-256。缺失时使用完整下载。",
      ],
    ],
    notes: [
      "差分还要求应用开启 diff_update、新版本开启 cloneDiffEnable 且为 canonical、有旧版本记录、文件可读并成功生成；生成失败可能仍返回正常版本信息但不含 diffUrl。",
    ],
  },
  clone: {
    params: [
      [
        "versionKey / query",
        "string",
        "否",
        "筛选版本索引；选择该索引下最新已发布的 canonical APK，省略则从所有索引中选择。",
      ],
    ],
    returns: [
      ["diffUrl", "string", "短期有效的 DELTA1 补丁地址。"],
      ["newPackage", "string", "生成的新包名。"],
      [
        "originalSha256 / cloneSha256",
        "string",
        "输入 canonical APK / 还原后 APK 的 SHA-256。",
      ],
      [
        "versionName / versionKey / versionNo",
        "string",
        "此次使用的 canonical 版本。",
      ],
    ],
    notes: [
      "仅普通账号支持；卡密版返回 400。需开启随机包名，否则 1005；无可用 canonical 返回 1004；并发生成繁忙返回 1017。",
      "每次调用可能生成新的包名和补丁。验证输入、输出 SHA-256 后再安装；补丁不是直接可安装 APK。",
    ],
  },
  regen: {
    returns: [
      ["regenerated", "boolean", "true 表示密钥已重建。"],
      ["certBytes", "integer", "证书 DER 编码的字节数；不返回私钥或证书内容。"],
    ],
    notes: [
      "仅普通账号支持；卡密版返回 400。该调用改变后续 APK 签名，不要作为心跳或自动重试请求。",
    ],
  },
  diffdl: {
    params: [
      [
        "name / path",
        "string",
        "是",
        "使用 ini 或 clone 返回的 diffUrl；不要自行拼装不同应用的名称。",
      ],
    ],
    returns: [
      [
        "HTTP body",
        "binary",
        "成功为 DELTA1 字节，无 JSON 信封；失效可直接 HTTP 404，鉴权错误仍可能是 JSON。",
      ],
    ],
    notes: [
      "不需要登录 token，但要求 X-App-Id、应用启用、限流及按配置启用的 mTLS。下载后先核验旧包哈希，应用补丁后核验新包哈希。",
    ],
  },
};
const raw = {
  method: "GET",
  format: "JSON",
  token: "否",
  crypto: "否",
  sign: "否",
  charset: "UTF-8",
};
export const sdkExtraEndpoints = {
  handshakeApi: {
    title: "会话握手",
    info: {
      path: "/api/sdk/handshake",
      method: "POST",
      format: "JSON",
      token: "否",
      crypto: "握手专用 RSA / AES",
      sign: "按应用配置",
      charset: "UTF-8",
    },
    intro:
      "仅 session 模式需要，完整密钥流程见左侧「握手 (session)」。握手不需要登录 token，成功后仍需登录。",
    params: [
      [
        "key / JSON",
        "string",
        "是",
        "随机 32 字节 AES 密钥经服务器公钥 RSA/PKCS1v15 加密，再按配置 base64 或 hex 编码。",
      ],
    ],
    reqExamples: [{ code: '{ "key": "<RSA 加密并编码后的 AES 密钥>" }' }],
    returns: [
      [
        "data",
        "string",
        "AES-CBC 加密并编码的 JSON；以此次生成的 AES 密钥解出 sessionId。",
      ],
      [
        "解密后 sessionId",
        "string",
        "后续 JSON 业务请求的 X-Session-Id；24 小时滑动过期，和登录 token 独立。",
      ],
    ],
    success: '{ "code": 0, "msg": "ok", "data": "<编码后的 AES 密文>" }',
    fail: '{ "code": 1015, "msg": "握手密钥格式错误" }',
    failNote:
      "400 未启用 session 或参数格式错误；1014 签名错误；1015 密钥无效；1016 时钟偏差。",
  },
  fileList: {
    title: "扩展文件列表",
    info: { ...raw, path: "/api/sdk/extend-file/list" },
    intro:
      "列出当前应用和全局（appid=0）的扩展文件。返回数组，无分页；不需要客户端登录令牌。",
    params: [
      ["category / query", "string", "否", "分类精确匹配。"],
      [
        "keyword / query",
        "string",
        "否",
        "匹配名称、备注或版本；使用 SQL LIKE，% 和 _ 具有通配含义。",
      ],
    ],
    reqExamples: [
      {
        code: "GET /api/sdk/extend-file/list?category=config&keyword=1.0\nX-App-Id: <应用 ID>",
      },
    ],
    returns: [
      ["data", "array<object>", "无文件时 []。按 appid、createdAt、id 倒序。"],
      ["[].id / appid", "integer", "文件 ID 与所属应用，appid=0 表示全局。"],
      [
        "[].name / filename / category / remark / version",
        "string",
        "展示名、原文件名、分类、备注、版本。",
      ],
      [
        "[].filesize / filetype",
        "integer / string",
        "文件字节数与 MIME 类型。",
      ],
      [
        "[].downloadUrl",
        "string",
        "上传文件的静态 URL；与带 X-App-Id 的 SDK 下载接口是两种访问方式。",
      ],
      [
        "[].downloadCount / createdAt / updatedAt",
        "integer",
        "SDK/后台下载接口计数、创建和修改时间（Unix 秒）；静态 URL 下载不增加此计数。",
      ],
      ["[].filepath", "string", "SDK 固定返回空字符串，不暴露服务器文件路径。"],
    ],
    success:
      '{ "code": 0, "msg": "ok", "data": [{ "id": 1, "appid": 1000, "name": "配置文件", "filename": "config.json", "filepath": "", "filesize": 128, "filetype": "application/json", "category": "config", "remark": "客户端配置", "version": "1.0", "downloadUrl": "/uploads/extend_files/example.json", "downloadCount": 0, "createdAt": 1733200000, "updatedAt": 1733200000 }] }',
    notes: [
      "此接口和下载接口不按用户权限过滤。静态 downloadUrl 不经过 SDK 用户鉴权，但会检查文件记录及所属应用是否有效；应用停用、删除或文件记录删除后不可继续下载。不应在扩展文件中放置密码、私钥或用户专属秘密。",
    ],
  },
  fileDownload: {
    title: "扩展文件下载",
    info: {
      ...raw,
      path: "/api/sdk/extend-file/:id/download",
      format: "文件附件",
      charset: "二进制",
    },
    intro: "按文件 ID 下载当前应用或全局文件；成功时下载计数加一。",
    params: [
      [
        "id / path",
        "integer",
        "是",
        "文件列表返回的正整数 ID；其他应用私有文件不可访问。",
      ],
    ],
    reqExamples: [
      { code: "GET /api/sdk/extend-file/1/download\nX-App-Id: <应用 ID>" },
    ],
    returns: [
      [
        "HTTP body",
        "binary",
        "成功时为附件原始内容，以 Content-Disposition 识别；错误可能是 HTTP 200 的 JSON 信封。",
      ],
    ],
    success:
      'HTTP/1.1 200 OK\nContent-Disposition: attachment; filename="config.json"\n\n<文件原始字节>',
    fail: '{ "code": 404, "msg": "文件不存在" }',
    failNote:
      "400 非法 ID；404 文件不存在、已丢失或不属于当前应用/全局。下载 JSON 文件时不能仅凭 Content-Type 判断为错误。",
  },
};
export const sdkProfileFields = [
  ["id", "integer", "当前应用内账号记录 ID。"],
  ["username", "string", "普通账号名称；卡密账号为卡号。"],
  ["nickname / avatar", "string", "昵称和头像 URL；avatar 可能为空。"],
  ["vip", "integer", "VIP 到期时间（Unix 秒）。"],
  ["vipActive", "boolean", "到期时间是否晚于现在，不等价于全部授权检查通过。"],
  ["mi / region", "string", "最近记录的机器码 / 地域。"],
  [
    "banExpire",
    "integer",
    "0 未封禁，负数永久封禁，正数为封禁结束时间（秒）。封禁账号通常已在鉴权阶段被拒绝。",
  ],
  ["bindLimit", "integer", "额外设备名额；允许绑定数量为 1 + bindLimit。"],
  [
    "groupId / groupName / groupLevel",
    "integer / string / integer",
    "用户分组 ID、名称、级别；卡密账号 groupLevel 可为 0。",
  ],
  [
    "hasSignKey",
    "boolean",
    "是否已有 APK 独立签名密钥；不是客户端 HMAC tokenKey。",
  ],
  [
    "isKami / card / regionLimit",
    "boolean / string / integer",
    "卡密账号额外返回：true、卡号、地域限制数量。普通账号不返回这三项。",
  ],
];
