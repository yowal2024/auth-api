<!-- Reconstructed editable Vue source; see README for recovery limits. -->
<script>
import DocCode from "./DocCode.vue";
import DocTable from "./DocTable.vue";
import {
  Button,
  Link,
  Descriptions,
  DescriptionsItem,
  Input,
  Menu,
  MenuItem,
  MenuItemGroup,
  Tag,
  Breadcrumb,
  BreadcrumbItem,
  Empty,
  Card,
  Alert,
  Table,
  Tabs,
  TableColumn,
  TabPane,
} from "@arco-design/web-vue";
import { ref, computed, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import SdkIntegrationDocs from "./SdkIntegrationDocs.vue";
import {
  sdkDetails,
  sdkExtraEndpoints,
  sdkProfileFields,
} from "./sdk-reference.js";
import CloudFunctionDocs from "./CloudFunctionDocs.vue";
import { cloudExamples } from "./cloud-examples.js";
import { entryPath } from "@/utils/route.js";
const he = `sign = HMAC_SHA256( token_key, X-Ts + "\\n" + X-Nonce + "\\n" + raw_body )
X-Sign = hex(sign)   // 小写 hex`;
const Te =
  '{ "code": 0, "msg": "ok", "data": "<encode( AES-CBC( X, {\\"sessionId\\":\\"<hex>\\"} ) )>" }';
const Se = `{
  "code": 0, "msg": "ok",
  "data": {
    "token": "<登录态 token>",
    "serverTime": 1733200000,
    "user": {
      "id": 1, "username": "alice", "nickname": "Alice", "avatar": "",
      "vip": 1735000000, "vipActive": true,
      "mi": "<机器码>", "region": "广东深圳",
      "banExpire": 0, "bindLimit": 2,
      "groupId": 0, "groupName": "", "groupLevel": 0,
      "hasSignKey": false
    }
  }
}`;
const sampleUser = JSON.parse(Se).data.user;
const Pe = JSON.stringify(
  { code: 0, msg: "ok", data: { token: "<登录态 token>", user: sampleUser } },
  null,
  2,
);
const xe =
  '{ "code": 0, "msg": "ok", "data": { "extends": { "key1": "v1", "key2": "v2" } } }';
const Ae = `{
  "code": 0, "msg": "ok",
  "data": { "vip": 1735000000, "vipActive": true, "groupName": "月卡", "addSec": 2592000 }
}`;
const Le = JSON.stringify(
  { code: 0, msg: "ok", data: { user: sampleUser, serverTime: 1733200000 } },
  null,
  2,
);
const Ie = Le;
const we = JSON.stringify(
  {
    code: 0,
    msg: "ok",
    data: {
      avatar: "/uploads/avatars/example.png",
      user: { ...sampleUser, avatar: "/uploads/avatars/example.png" },
      serverTime: 1733200000,
    },
  },
  null,
  2,
);
const Ee = `{
  "code": 0, "msg": "ok",
  "data": {
    "notices": [
      { "id": 1, "title": "公告", "content": "...", "author": "admin", "updateTime": 1733200000 }
    ],
    "extends": { "someKey": "someValue" },
    "latestVersion": {
      "versionName": "安卓版", "versionKey": "android", "versionNo": "1.0.1",
      "apkUrl": "https://.../app.apk", "url": "https://...",
      "changelog": "修复若干问题", "hasUpdate": true,

      "diffUrl": "/api/sdk/diff/<old>_<new>.diff",
      "diffOldSha": "<sha256>", "diffNewSha": "<sha256>"
    }
  }
}`;
const Oe = `{
  "code": 0, "msg": "ok",
  "data": {
    "diffUrl": "/api/sdk/diff/<originalSha>_<cloneSha>.diff",
    "newPackage": "com.gv35rr.pm7pn4.f55e56t",
    "originalSha256": "<canonical apk sha256>",
    "cloneSha256": "<还原后应得的 sha256>",
    "versionName": "安卓版", "versionKey": "android", "versionNo": "1.0.0"
  }
}`;
const Ne =
  '{ "code": 0, "msg": "ok", "data": { "regenerated": true, "certBytes": 1234 } }';
const Xe = `HTTP/1.1 200 OK
Content-Type: application/octet-stream

<DELTA1 补丁二进制>`;
const Ce = `偏移   大小   字段
0      8      magic = "DELTA1\\0\\0"
8      8      newSize     (u64 LE) — 还原后 APK 大小
16     8      ctrlLen     (u64 LE) — zlib 压缩后的 control 流长度
24     8      diffLen     (u64 LE) — zlib 压缩后的 diff   流长度
32     8      extraLen    (u64 LE) — zlib 压缩后的 extra  流长度
40     ...    zlib(control) || zlib(diff) || zlib(extra)`;
const qe = `import "auth-api/internal/buildpool/bsdiff"

func patch(oldAPK, diffBlob []byte) ([]byte, error) {
    return bsdiff.Apply(oldAPK, diffBlob)
}`;
const Be = `// patch: DELTA1 字节;old: 本地 canonical APK 字节
fun applyDelta1(old: ByteArray, patch: ByteArray): ByteArray {
    require(patch.size >= 40 && patch.copyOfRange(0, 8).contentEquals(
        byteArrayOf(0x44, 0x45, 0x4c, 0x54, 0x41, 0x31, 0, 0)
    )) { "not DELTA1" }

    val bb = ByteBuffer.wrap(patch).order(ByteOrder.LITTLE_ENDIAN)
    val newSize  = bb.getLong(8).toInt()
    val ctrlLen  = bb.getLong(16).toInt()
    val diffLen  = bb.getLong(24).toInt()
    val extraLen = bb.getLong(32).toInt()

    fun inflate(src: ByteArray): ByteArray {
        val inf = Inflater(); inf.setInput(src)
        val out = ByteArrayOutputStream(); val buf = ByteArray(64 * 1024)
        while (!inf.finished()) { val n = inf.inflate(buf); if (n == 0) break; out.write(buf, 0, n) }
        inf.end(); return out.toByteArray()
    }

    val ctrl  = inflate(patch.copyOfRange(40, 40 + ctrlLen))
    val diff  = inflate(patch.copyOfRange(40 + ctrlLen, 40 + ctrlLen + diffLen))
    val extra = inflate(patch.copyOfRange(40 + ctrlLen + diffLen, 40 + ctrlLen + diffLen + extraLen))

    fun offt(b: ByteArray, o: Int): Long {
        var v = 0L
        for (i in 0..6) v = v or ((b[o + i].toLong() and 0xff) shl (i * 8))
        v = v or ((b[o + 7].toLong() and 0x7f) shl 56)
        return if (b[o + 7].toInt() and 0x80 != 0) -v else v
    }

    val out = ByteArray(newSize)
    var newPos = 0; var oldPos = 0; var cp = 0; var dp = 0; var ep = 0
    while (newPos < newSize) {
        val x = offt(ctrl, cp).toInt(); val y = offt(ctrl, cp + 8).toInt()
        val z = offt(ctrl, cp + 16).toInt(); cp += 24
        for (i in 0 until x) {
            val ob = if (oldPos + i in 0 until old.size) old[oldPos + i] else 0
            out[newPos + i] = (ob + diff[dp + i]).toByte()
        }
        dp += x; newPos += x; oldPos += x
        System.arraycopy(extra, ep, out, newPos, y)
        ep += y; newPos += y; oldPos += z
    }
    return out
}`;
const Ke = `1. (可选) algorithm=session → POST /handshake 换 sessionId
2. POST /ini                 → 公告、扩展、最新版本;若有 diffUrl 走差量更新
3. POST /logon               → 拿 token + user(校验 vipActive / banExpire)
4. 业务期间任意带 token 的调用都会自动续在线;需要时 POST /heartbeat 校时
5. (可选) GET /clone          → 换随机包名;GET /diff/:name 下补丁 → bspatch → 安装
6. POST /logout              → 退出`;
export default {
  name: "Documentation",
  components: {
    Button,
    Link,
    Descriptions,
    DescriptionsItem,
    Input,
    Menu,
    MenuItem,
    MenuItemGroup,
    Tag,
    Breadcrumb,
    BreadcrumbItem,
    Empty,
    Card,
    DocCode,
    DocTable,
    CloudFunctionDocs,
    SdkIntegrationDocs,
    Alert,
    Table,
    TableColumn,
    Tabs,
    TabPane,
  },
  setup(ze) {
    const route = useRoute();
    const router = useRouter();
    const contentRef = ref(null);
    const search = ref("");
    const navOpen = ref(false);
    const outline = ref([]);
    const activeHeading = ref("");
    function updateActiveHeading() {
      const top = contentRef.value?.getBoundingClientRect().top || 0;
      const passed = outline.value.filter(
        (item) => item.el.getBoundingClientRect().top <= top + 100,
      );
      activeHeading.value = (passed.at(-1) || outline.value[0])?.id || "";
    }
    function jumpTo(item) {
      const pane = contentRef.value;
      pane.scrollTo({
        top:
          pane.scrollTop +
          item.el.getBoundingClientRect().top -
          pane.getBoundingClientRect().top -
          24,
      });
      activeHeading.value = item.id;
      item.el.focus({ preventScroll: true });
    }
    const r = ref("overview");
    const b = (g) => {
      navOpen.value = false;
      router.push({ query: { ...route.query, section: g } });
      contentRef.value && (contentRef.value.scrollTop = 0);
    };
    const tableRows = [
      {
        h: "X-App-Id",
        when: "所有 SDK 请求",
        desc: "应用 id(整数)。缺失/非法 → 1010",
      },
      {
        h: "X-Ts",
        when: "JSON 业务接口 / 握手，且 app 配了 token_key 时",
        desc: "Unix 秒级时间戳,校验时间窗(默认 ±60s)",
      },
      {
        h: "X-Nonce",
        when: "同上",
        desc: "随机串,与 X-Ts 一起参与签名,防重放",
      },
      {
        h: "X-Sign",
        when: "同上",
        desc: "HMAC-SHA256 签名,见签名一节",
      },
      {
        h: "X-Session-Id",
        when: "加密模式 = session 的业务接口",
        desc: "握手拿到的 sessionId;缺失 → 1020",
      },
      {
        h: "X-Token / Authorization",
        when: "需 token 的接口",
        desc: "X-Token: <token> 或 Authorization: Bearer <token>,二者任一",
      },
      {
        h: "X-Mi",
        when: "bind_device 开启时的需 token 接口",
        desc: "机器码必须与当前 token 登录时的 mi 一致，且设备绑定仍有效；否则返回 1006",
      },
    ];
    const tableRows2 = [
      {
        algo: "空 / 未配",
        mode: "不加密,明文透传",
        key: "—",
      },
      {
        algo: "aes / des / rc4",
        mode: "对称,所有客户端共用一把",
        key: "sym_key",
      },
      {
        algo: "rsa",
        mode: "非对称",
        key: "解请求用 server_private_key,加响应用 client_public_key",
      },
      {
        algo: "session",
        mode: "握手,每客户端一把临时 AES",
        key: "握手产物(见握手一节)",
      },
    ];
    const tableRows3 = [
      {
        code: 403,
        msg: "应用停用或前置云函数拒绝",
        cause: "应用不存在 / 未启用；或脚本 allow=false",
      },
      {
        code: 404,
        msg: "资源不存在",
        cause: "文件、云函数不存在；补丁可能直接 HTTP 404",
      },
      {
        code: 1099,
        msg: "云函数业务错误",
        cause: "脚本抛出业务错误:前缀的异常",
      },
      {
        code: 0,
        msg: "成功",
        cause: "—",
      },
      {
        code: 400,
        msg: "参数错误",
        cause: "body 解析失败 / 必填缺失",
      },
      {
        code: 500,
        msg: "服务器内部错误",
        cause: "未预期异常",
      },
      {
        code: 1001,
        msg: "账号或密码错误",
        cause: "登录校验失败",
      },
      {
        code: 1002,
        msg: "账号已被封禁",
        cause: "ban_expire < 0 永久封禁，或 ban_expire > now",
      },
      {
        code: 1003,
        msg: "IP 已被封禁",
        cause: "命中 IP 黑名单",
      },
      {
        code: 1004,
        msg: "未找到可用版本",
        cause: "app 无 status=1 版本",
      },
      {
        code: 1005,
        msg: "当前版本未启用随机包名",
        cause: "应用未开启随机包名",
      },
      {
        code: 1006,
        msg: "未登录或 token 失效",
        cause: "缺 token / 过期 / 跨 app",
      },
      {
        code: 1007,
        msg: "用户名已被占用",
        cause: "注册重名",
      },
      {
        code: 1008,
        msg: "设备数达上限",
        cause: "超 1 + bind_limit",
      },
      {
        code: 1009,
        msg: "IP 地点数超出限制",
        cause: "超 region_limit",
      },
      {
        code: 1010,
        msg: "X-App-Id 缺失或非法",
        cause: "没传 / 非整数",
      },
      {
        code: 1011,
        msg: "缺少机器码 mi",
        cause: "bind_device 开启却没传 mi",
      },
      {
        code: 1014,
        msg: "签名校验失败",
        cause: "X-Sign 不符 / 缺签名头 / nonce 重复",
      },
      {
        code: 1015,
        msg: "加密数据格式错误",
        cause: "信封解密失败 / handshake key 非法",
      },
      {
        code: 1016,
        msg: "请求时间窗超出",
        cause: "X-Ts 偏差超 time_skew",
      },
      {
        code: 1017,
        msg: "请求过于频繁 / 生成繁忙",
        cause: "接口限流、防爆破或补丁并发已满",
      },
      {
        code: 1018,
        msg: "VIP 已到期",
        cause: "app.mode=vip 且 VIP 过期",
      },
      {
        code: 1019,
        msg: "mTLS 证书缺失或不匹配",
        cause: "证书链、用途或 CN 不满足要求",
      },
      {
        code: 1020,
        msg: "session 失效",
        cause: "缺 X-Session-Id / 过期 → 重新握手",
      },
      {
        code: 1021,
        msg: "机器码已被封禁",
        cause: "命中 mi 黑名单",
      },
      {
        code: 2001,
        msg: "卡密不存在",
        cause: "—",
      },
      {
        code: 2002,
        msg: "卡密已被使用",
        cause: "充值时卡密已使用或冻结；已消费的充值卡不能作为卡密账号登录",
      },
      {
        code: 2003,
        msg: "卡密已冻结",
        cause: "status=2",
      },
    ];
    const N = {
      ...sdkExtraEndpoints,
      logon: {
        title: "登录",
        info: {
          path: "/api/sdk/logon",
          method: "POST",
          format: "JSON",
          token: "否",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "用户版(<code>app.type=user</code>)用账号密码;卡密版(<code>app.type=kami</code>)用卡号、无密码,首次登录即「激活」,按卡密分组算 VIP 时长。<code>bind_device</code> 开启时 <code>mi</code> 必填(缺失 → 1011)。",
        reqExamples: [
          {
            label: "用户版请求",
            code: '{ "username": "alice", "password": "123456", "mi": "<机器码>" }',
          },
          {
            label: "卡密版请求(卡号放 card 或 username 均可)",
            code: '{ "card": "KAMI-XXXX", "mi": "<机器码>" }',
          },
        ],
        notes: [
          "卡密版 <code>user</code> 含 <code>card</code>、<code>isKami:true</code>、<code>regionLimit</code>,公共字段仍存在,可能为空或默认值。",
          "普通账号登录且开启随机包名时会生成缺失的 APK 签名密钥；<code>/clone</code> 也会按需生成。",
        ],
        success: Se,
        fail: '{ "code": 1001, "msg": "账号或密码错误", "data": null }',
        failNote:
          "其它:1002 封禁 / 1003 IP 封禁 / 1008 设备超限 / 1009 地点超限 / 1011 缺机器码 / 1017 失败过多 / 1018 VIP 过期 / 1021 机器码封禁。",
      },
      register: {
        title: "注册",
        info: {
          path: "/api/sdk/register",
          method: "POST",
          format: "JSON",
          token: "否",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "仅用户版自助注册。成功返回 <code>{ token, user }</code>,结构同登录。",
        reqExamples: [
          {
            code: '{ "username": "bob01", "password": "123456", "nickname": "Bob", "mi": "<机器码>" }',
          },
        ],
        success: Pe,
        fail: '{ "code": 1007, "msg": "用户名已被占用", "data": null }',
        failNote:
          "用户名占用 → 1007;卡密版调用 → 400；账号 5–12 字节、密码 6–18 字节。",
      },
      logout: {
        title: "登出",
        info: {
          path: "/api/sdk/logout",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "注销当前 token，其他会话保持有效；最后一个活跃会话退出后账号显示离线。返回后清除本地令牌。JSON 请求可传 {}，开启加密时必须发送信封。",
        success: '{ "code": 0, "msg": "ok", "data": { "ok": true } }',
        fail: '{ "code": 1006, "msg": "未登录或 token 失效", "data": null }',
      },
      heartbeat: {
        title: "心跳",
        info: {
          path: "/api/sdk/heartbeat",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "传 {}。通过鉴权的请求会更新在线状态；心跳返回服务器时间。登录 token 按应用配置滑动续期，在线统计按会话最近 5 分钟活动汇总，短于 5 分钟的令牌按令牌有效期计算；不等同于 TCP 连接状态。",
        success:
          '{ "code": 0, "msg": "ok", "data": { "ok": true, "serverTime": 1733200000 } }',
        fail: '{ "code": 1006, "msg": "未登录或 token 失效", "data": null }',
      },
      info: {
        title: "用户信息",
        info: {
          path: "/api/sdk/info",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "传 {}。拉取当前用户信息；封禁或 VIP 校验未通过时会直接返回错误。<code>user</code> 结构同登录。",
        success: Le,
        fail: '{ "code": 1006, "msg": "未登录或 token 失效", "data": null }',
      },
      userExtends: {
        title: "用户扩展",
        info: {
          path: "/api/sdk/userExtends",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "读写合一,按当前用户隔离。纯读:必须传 JSON <code>{}</code>;写:按 key upsert。不论读写都返回该用户全量 <code>extends</code>。",
        reqExamples: [
          {
            label: "写入",
            code: '{ "set": { "key1": "v1" } }',
          },
        ],
        success: xe,
        fail: '{ "code": 1006, "msg": "未登录或 token 失效", "data": null }',
      },
      kamiTopup: {
        title: "卡密充值",
        info: {
          path: "/api/sdk/kamiTopup",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro: "给<b>当前登录账号</b>用一张卡密加 VIP 时长。",
        reqExamples: [
          {
            code: '{ "card": "KAMI-XXXX" }',
          },
        ],
        success: Ae,
        fail: '{ "code": 2001, "msg": "卡密不存在", "data": null }',
        failNote:
          "2001 充值卡不存在 / 2002 充值卡已使用或冻结 / 400 卡密类型或面值不支持。",
      },
      updateInfo: {
        title: "修改资料",
        info: {
          path: "/api/sdk/updateInfo",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "当前登录用户自助改昵称 / 头像。字段都可选,缺省的不动;头像传 <b>URL 字符串</b>(客户端自行上传图床,或先调 <code>/uploadAvatar</code> 直传二进制拿 URL)。",
        reqExamples: [
          {
            code: '{ "nickname": "新昵称", "avatar": "https://cdn.example.com/a.png" }',
          },
        ],
        notes: [
          "<code>nickname</code>:普通用户 / 卡密用户都可改,最长 32 字。",
          "<code>avatar</code>:仅普通用户可改；卡密版传入返回 400。最长 255 个 UTF-8 字节。",
          "返回最新 <code>{ user, serverTime }</code>(结构同登录),无需再拉 <code>/info</code>。",
        ],
        success: Ie,
        fail: '{ "code": 400, "msg": "未提供可更新字段", "data": null }',
        failNote: "400 未提供字段 / 过长;1006 未登录。",
      },
      uploadAvatar: {
        title: "上传头像",
        info: {
          path: "/api/sdk/uploadAvatar",
          method: "POST",
          format: "multipart/form-data",
          token: "是",
          crypto: "否",
          sign: "否",
          charset: "UTF-8",
        },
        intro:
          "直传头像<b>二进制</b>,服务端落盘后把 URL 写入当前用户 <code>avatar</code>,一次完成「上传 + 绑定」。<b>走 multipart,不套加密信封、不参与签名</b>；仍执行应用、限流、token 及按配置启用的设备、地域、VIP 和 mTLS 检查。",
        reqExamples: [
          {
            label: "请求(multipart)",
            code: `POST /api/sdk/uploadAvatar
X-App-Id: 5
X-Token: <token>
Content-Type: multipart/form-data; boundary=...

字段名 file:头像文件(jpg/png/gif/webp,≤5 MiB)`,
          },
        ],
        success: we,
        fail: '{ "code": 400, "msg": "卡密版账号不支持头像", "data": null }',
        failNote:
          "400 未上传 / 超 5 MiB / 格式不支持;1006 未登录;400 卡密版不支持头像。",
      },
      cloudFunction: {
        title: "调用云函数",
        info: {
          path: "/api/sdk/cloudFunction",
          method: "POST",
          format: "JSON",
          token: "是",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "按名称调用当前应用内已启用的 manual 函数。脚本编写、触发器和内置 API 见左侧「云函数开发指南」。",
        reqExamples: [{ label: "请求", code: cloudExamples.request }],
        notes: [
          "<code>name</code> 为函数名称；<code>param</code> 作为脚本的 <code>ctx.payload</code>。",
          "需要客户端登录令牌，不能使用管理员令牌；设备绑定、签名、加密、mTLS 规则仍然生效。",
          "成功时 <code>data</code> 为 handler 返回值，不包含后台试运行的日志、耗时和 ok 字段。",
        ],
        success:
          '{ "code": 0, "msg": "ok", "data": { "message": "Hello, Alice", "appId": 1000 } }',
        fail: '{ "code": 1099, "msg": "自定义业务错误" }',
        failNote:
          "404 函数不存在、禁用或不是 manual；500 执行失败；1099 为脚本抛出的业务错误。错误码可重定义。",
      },
      ini: {
        title: "初始化",
        info: {
          path: "/api/sdk/ini",
          method: "POST",
          format: "JSON",
          token: "否",
          crypto: "按应用配置",
          sign: "按应用配置",
          charset: "UTF-8",
        },
        intro:
          "SDK 启动拉一次:公告 + 应用级扩展 + 最新版本 + 是否需要更新 / 差量更新。可选带当前版本。",
        reqExamples: [
          {
            code: '{ "versionKey": "android", "versionNo": "1.0.0" }',
          },
        ],
        notes: [
          "<code>hasUpdate</code>:不带版本号默认 <code>true</code>(首次);带了则 <code>versionKey</code>/<code>versionNo</code> 任一不一致即 <code>true</code>。",
          "<code>diffUrl</code> 三件套<b>仅当</b>:有更新 + app 开启 <code>diff_update</code> + 新版本是 canonical + 能反查到旧 canonical 时才出现。",
          "版本通过 <code>versionKey</code> + <code>versionNo</code> 组合识别,不靠 APK 的 sha256(随机包名后本地 sha 已变)。",
        ],
        success: Ee,
      },
      clone: {
        title: "换包名",
        info: {
          path: "/api/sdk/clone",
          method: "GET",
          format: "—(无 body)",
          token: "是",
          crypto: "否",
          sign: "否",
          charset: "UTF-8",
        },
        intro:
          "无请求体。每次调用都基于当前最新 canonical 生成一份<b>新的</b>随机包名 clone,与版本是否更新无关。机制说明见左侧「差量更新机制」。",
        reqExamples: [
          {
            label: "请求",
            code: `GET /api/sdk/clone
X-App-Id: 5
X-Token: <token>`,
          },
        ],
        success: Oe,
        fail: '{ "code": 1004, "msg": "未找到可用版本", "data": null }',
        failNote: "1004 无可用版本 / 1005 未启用随机包名 / 1006 未登录。",
      },
      regen: {
        title: "重置签名",
        info: {
          path: "/api/sdk/regen-sign",
          method: "POST",
          format: "—(无 body)",
          token: "是",
          crypto: "否",
          sign: "否",
          charset: "UTF-8",
        },
        intro:
          "强制重生成本用户的 RSA 2048 + 自签证书(CN = 登录账号)。之后 <code>/clone</code> 拿到的 APK 用新证书签,因签名不一致客户端需<b>全量重装</b>。",
        success: Ne,
        fail: '{ "code": 1006, "msg": "未登录或 token 失效", "data": null }',
      },
      diffdl: {
        title: "下补丁",
        info: {
          path: "/api/sdk/diff/:name",
          method: "GET",
          format: "application/octet-stream",
          token: "否",
          crypto: "否",
          sign: "否",
          charset: "—(二进制)",
        },
        intro:
          "下载 bspatch 补丁字节。补丁<b>只走这个端点,不走 <code>/uploads/</code> 静态路由</b>。继承 <code>X-App-Id</code> + mTLS,不挂 token。补丁 <b>TTL 默认 60 秒，可由服务端配置</b>,过期 / 非法文件名统一 404。",
        reqExamples: [
          {
            label: "请求",
            code: `GET /api/sdk/diff/<originalSha>_<cloneSha>.diff
X-App-Id: 5`,
          },
        ],
        success: Xe,
        fail: "404 Not Found  —— 补丁过期 / 文件名非法",
        failNote: "mTLS 开启时无证书 / CN 不符 → 1019。",
      },
    };
    const a = computed(() => {
      const g = N[r.value];
      return g
        ? {
            key: r.value,
            ...g,
            ...sdkDetails[r.value],
          }
        : null;
    });
    const X = [
      {
        group: "接入说明",
        items: [
          { key: "client", title: "客户端接入指南" },
          { key: "profile", title: "用户返回字段" },
          { key: "cloud", title: "云函数开发指南" },
          {
            key: "overview",
            title: "概览",
          },
          {
            key: "headers",
            title: "请求头",
          },
          {
            key: "sign",
            title: "签名",
          },
          {
            key: "crypto",
            title: "加密信封",
          },
          {
            key: "handshake",
            title: "握手 (session)",
          },
          {
            key: "diff",
            title: "差量更新机制",
          },
          {
            key: "errors",
            title: "错误码",
          },
          {
            key: "flow",
            title: "启动流程",
          },
        ],
      },
      {
        group: "业务接口",
        items: [
          { key: "handshakeApi", title: "会话握手", method: "POST" },
          { key: "cloudFunction", title: "调用云函数", method: "POST" },
          {
            key: "logon",
            title: "登录",
            method: "POST",
          },
          {
            key: "register",
            title: "注册",
            method: "POST",
          },
          {
            key: "logout",
            title: "登出",
            method: "POST",
          },
          {
            key: "heartbeat",
            title: "心跳",
            method: "POST",
          },
          {
            key: "info",
            title: "用户信息",
            method: "POST",
          },
          {
            key: "userExtends",
            title: "用户扩展",
            method: "POST",
          },
          {
            key: "kamiTopup",
            title: "卡密充值",
            method: "POST",
          },
          {
            key: "updateInfo",
            title: "修改资料",
            method: "POST",
          },
          {
            key: "uploadAvatar",
            title: "上传头像",
            method: "POST",
          },
          {
            key: "ini",
            title: "初始化",
            method: "POST",
          },
        ],
      },
      {
        group: "文件接口",
        items: [
          { key: "fileList", title: "扩展文件列表", method: "GET" },
          { key: "fileDownload", title: "扩展文件下载", method: "GET" },
        ],
      },
      {
        group: "升级接口",
        items: [
          {
            key: "clone",
            title: "换包名",
            method: "GET",
          },
          {
            key: "regen",
            title: "重置签名",
            method: "POST",
          },
          {
            key: "diffdl",
            title: "下补丁",
            method: "GET",
          },
        ],
      },
    ];
    const overviewIndex = X[0].items.findIndex(
      (item) => item.key === "overview",
    );
    X[0].items.unshift(...X[0].items.splice(overviewIndex, 1));
    const allItems = X.flatMap((group) =>
      group.items.map((item) => ({ ...item, group: group.group })),
    );
    const currentItem = computed(() =>
      allItems.find((item) => item.key === r.value),
    );
    const visibleGroups = computed(() => {
      const query = search.value.trim().toLowerCase();
      return X.map((group) => ({
        ...group,
        items: group.items.filter((item) =>
          [item.title, item.key, item.method, N[item.key]?.info.path]
            .join(" ")
            .toLowerCase()
            .includes(query),
        ),
      })).filter((group) => group.items.length);
    });
    const neighbors = computed(() => {
      const i = allItems.findIndex((item) => item.key === r.value);
      return { previous: allItems[i - 1], next: allItems[i + 1] };
    });
    const sections = new Set(
      X.flatMap((group) => group.items.map((item) => item.key)),
    );
    watch(
      () => route.query.section,
      async (section, previousSection) => {
        r.value = sections.has(section) ? section : "overview";
        await nextTick();
        if (!contentRef.value) return;
        contentRef.value.scrollTop = 0;
        if (
          previousSection !== undefined &&
          window.matchMedia("(max-width: 760px)").matches
        )
          window.scrollTo(0, 0);
        outline.value = [...contentRef.value.querySelectorAll("h3")]
          .filter((el) => el.getClientRects().length)
          .map((el, i) => {
            el.id = `doc-${r.value}-${i}`;
            el.tabIndex = -1;
            return { id: el.id, title: el.textContent.trim(), el };
          });
        updateActiveHeading();
      },
      { immediate: true },
    );
    return {
      search,
      navOpen,
      outline,
      activeHeading,
      visibleGroups,
      currentItem,
      neighbors,
      updateActiveHeading,
      jumpTo,
      r,
      b,
      tableRows,
      tableRows2,
      tableRows3,
      N,
      a,
      X,
      sdkProfileFields,
      contentRef,
      entryPath,
      he,
      Te,
      Ce,
      qe,
      Be,
      Ke,
    };
  },
};
</script>
<template>
  <div class="docs-page">
    <header class="docs-topbar">
      <div class="docs-brand">
        <span class="docs-brand-icon" aria-hidden="true">{ }</span>
        <div>
          <h1>开发文档</h1>
          <span>AUTH API · DEVELOPER CENTER</span>
        </div>
      </div>
      <div class="topbar-right">
        <span class="version-pill"><i></i> 17 个客户端接口</span
        ><Button
          class="back-link"
          type="text"
          @click="$router.push(entryPath('/app'))"
        >
          返回应用 <span aria-hidden="true">↗</span>
        </Button>
      </div>
    </header>
    <Card
      class="docs-card"
      :bordered="true"
      :body-style="{ padding: 0, height: '100%' }"
    >
      <div class="docs-layout">
        <aside class="docs-nav" aria-label="文档导航">
          <div class="nav-heading">
            <span>文档目录</span
            ><Button
              class="mobile-menu"
              type="text"
              :aria-expanded="navOpen"
              aria-controls="docs-navigation"
              @click="navOpen = !navOpen"
            >
              {{ navOpen ? "收起目录 −" : "展开目录 +" }}</Button
            ><span class="nav-count">GUIDES & API</span>
          </div>
          <div class="nav-search">
            <Input
              v-model="search"
              :input-attrs="{ 'aria-label': '搜索文档' }"
              placeholder="搜索接口、路径或标题"
              @input="navOpen = true"
            >
              <template #suffix
                ><Button
                  v-if="search"
                  type="text"
                  size="mini"
                  aria-label="清空搜索"
                  @click="search = ''"
                  >×</Button
                ></template
              >
            </Input>
          </div>
          <nav
            id="docs-navigation"
            class="nav-scroll"
            :class="{ expanded: navOpen }"
          >
            <Menu
              v-if="visibleGroups.length"
              :selected-keys="[r]"
              @menu-item-click="b"
            >
              <MenuItemGroup
                v-for="n in visibleGroups"
                :key="n.group"
                :title="n.group"
              >
                <MenuItem
                  v-for="p in n.items"
                  :key="p.key"
                  class="nav-item"
                  :aria-current="r === p.key ? 'page' : undefined"
                >
                  <Tag
                    v-if="p.method"
                    size="small"
                    :color="p.method === 'GET' ? 'arcoblue' : 'green'"
                    >{{ p.method }}</Tag
                  >
                  <span class="nav-item-title">{{ p.title }}</span>
                </MenuItem>
              </MenuItemGroup>
            </Menu>
            <Empty v-else class="nav-empty" description="没有找到相关文档"
              ><Button type="text" @click="search = ''">清空搜索</Button></Empty
            >
          </nav>
          <div class="nav-foot">
            <span class="nav-foot-dot"></span>默认接口前缀 <code>/api/sdk</code>
          </div>
        </aside>
        <div class="reading-area">
          <div class="reading-toolbar">
            <Breadcrumb
              ><BreadcrumbItem>文档</BreadcrumbItem
              ><BreadcrumbItem>{{ currentItem?.group }}</BreadcrumbItem
              ><BreadcrumbItem>{{
                currentItem?.title
              }}</BreadcrumbItem></Breadcrumb
            >
            <Tag size="small" color="arcoblue" class="reading-label">{{
              a ? "API 参考" : "开发指南"
            }}</Tag>
          </div>
          <div class="reading-columns">
            <div
              ref="contentRef"
              class="docs-content"
              @scroll.passive="updateActiveHeading"
            >
              <SdkIntegrationDocs v-if="r === 'client'" />
              <section v-if="r === 'profile'" class="panel">
                <h2 class="doc-h2">用户返回字段</h2>
                <p>
                  登录、注册、用户信息和修改资料共用下表。所有时间戳均为 Unix
                  秒；缺省字符串可能为空。卡密账号也返回公共字段，avatar、groupLevel、hasSignKey
                  等可为空或默认值。
                </p>
                <div class="field-scroll">
                  <DocTable
                    :headers="['字段', '类型', '含义']"
                    :rows="sdkProfileFields"
                    :code-column="0"
                  />
                </div>
                <p>
                  vipActive 等字段供展示；业务是否放行由服务端鉴权决定。user
                  中不会返回密码、APK 私钥或应用 tokenKey。
                </p>
              </section>
              <CloudFunctionDocs v-if="r === 'cloud'" />
              <section v-show="r === 'overview'" class="panel">
                <div class="overview-label">BUILD WITH AUTH API</div>
                <h2 class="doc-h2 overview-title">
                  从接入到上线，每一步都有据可查。
                </h2>
                <p class="overview-description">
                  客户端鉴权、应用接口与云函数开发说明，集中在这里。
                </p>
                <div class="quick-links">
                  <Button type="text" @click="b('client')">
                    <span class="quick-icon">01</span
                    ><strong>开始接入 <span>↗</span></strong
                    ><small>调用流程与完整代码示例</small>
                  </Button>
                  <Button type="text" @click="b('logon')">
                    <span class="quick-icon">02</span
                    ><strong>客户端 API <span>↗</span></strong
                    ><small>参数、鉴权与返回数据</small>
                  </Button>
                  <Button type="text" @click="b('cloud')">
                    <span class="quick-icon">03</span
                    ><strong>云函数开发 <span>↗</span></strong
                    ><small>触发器与运行环境</small>
                  </Button>
                </div>
                <h3 class="doc-h3">基础约定</h3>
                <p>
                  首次接入请先阅读
                  <Link @click="b('client')">客户端 API 接入指南</Link
                  >，包含完整接口分类、调用顺序与可运行的签名示例。用户资料字段见
                  <Link @click="b('profile')">用户返回字段</Link>。
                </p>
                <p>
                  所有 SDK 接口都在
                  <code>/api/sdk</code>
                  前缀下,
                  <b>
                    GET 用于 /clone、/diff/:name、/extend-file/list 和
                    /extend-file/:id/download；其余均为 POST
                  </b>
                  。 请求统一要求
                  <code>X-App-Id</code>
                  头,
                  <b>不需要 admin 登录鉴权</b>
                  。业务接口的 path 可被后台「接口管理」改写,文档里是默认 path。
                </p>
                <p>
                  JSON 接口统一响应信封:
                  <code>{ code, msg, data }</code>
                  。
                  <code>code=0</code>
                  为成功；HTTP 200 不代表业务成功，错误可能不含
                  data。文件与补丁成功时直接返回字节。非 0 见
                  <Link @click="(n) => b('errors')">错误码</Link>
                  。启用加密时
                  <code>data</code>
                  字段是密文字符串(见
                  <Link @click="(n) => b('crypto')">加密信封</Link>
                  )。
                </p>
                <Alert
                  type="warning"
                  :style="{
                    'margin-top': '12px',
                  }"
                >
                  <template #default>
                    安全判定一律以服务端返回的
                    <code>code</code>
                    为准(VIP / 封禁 / 设备 / IP 全在后端判)。 客户端不要本地比对
                    <code>vip</code>
                    /
                    <code>banExpire</code>
                    时间戳来放行——那只用于展示。
                  </template>
                </Alert>
              </section>
              <section v-show="r === 'headers'" class="panel">
                <h2 class="doc-h2">请求头</h2>
                <Table
                  :data="tableRows"
                  :pagination="!1"
                  :bordered="{
                    wrapper: !0,
                    cell: !0,
                  }"
                  size="small"
                >
                  <template #columns>
                    <TableColumn title="头" data-index="h" :width="220">
                      <template #cell="{ record: n }">
                        <code>{{ n.h }}</code>
                      </template>
                    </TableColumn>
                    <TableColumn
                      title="何时必带"
                      data-index="when"
                      :width="240"
                    ></TableColumn>
                    <TableColumn title="说明" data-index="desc"></TableColumn>
                  </template>
                </Table>
                <Alert
                  type="info"
                  :style="{
                    'margin-top': '12px',
                  }"
                >
                  <template #default>
                    <b>mTLS</b>
                    :app 开启
                    <code>mtls_enabled</code>
                    后,所有
                    <code>/api/sdk/*</code>
                    请求必须在 TLS 层带客户端证书, 且证书
                    <code>CN == &quot;app-{appid}&quot;</code>
                    ,否则 1019。证书由后台「应用 → mTLS」签发。
                  </template>
                </Alert>
              </section>
              <section v-show="r === 'sign'" class="panel">
                <h2 class="doc-h2">签名</h2>
                <p>
                  只要 app 配了
                  <code>token_key</code>
                  (后台「应用 → 设置」),所有业务接口(
                  <b>含握手</b>
                  )都要带签名,
                  <b>与加密是否开启无关</b>
                  。
                  <code>/clone</code>
                  、
                  <code>/regen-sign</code>
                  、
                  <code>/diff/:name</code>
                  、<code>/uploadAvatar</code> 和两个
                  <code>/extend-file</code> 接口不走签名。
                </p>
                <DocCode
                  ><code>{{ he }}</code></DocCode
                >
                <ul>
                  <li>
                    <code>raw_body</code>
                    是
                    <b>实际发出的请求体字节</b>
                    。启用加密时 body 已是密文信封,
                    <b>签名算的是密文</b>
                    (先加密再签名)。
                  </li>
                  <li>
                    同一应用内每次请求使用新 nonce，最长 256 字节；复用 nonce
                    返回 1014。
                  </li>
                  <li>
                    <code>X-Ts</code>
                    与服务器时间偏差超
                    <code>time_skew</code>
                    (默认 ±60s，上限 ±3600s)→ 1016。
                  </li>
                  <li>
                    app 未配
                    <code>token_key</code>
                    时跳过签名校验(老应用兼容),新接入建议配上。
                  </li>
                </ul>
              </section>
              <section v-show="r === 'crypto'" class="panel">
                <h2 class="doc-h2">加密信封</h2>
                <p>
                  加密由后台「应用 → 加密配置」决定,
                  <code>algorithm</code>
                  字段定模式:
                </p>
                <Table
                  :data="tableRows2"
                  :pagination="!1"
                  :bordered="{
                    wrapper: !0,
                    cell: !0,
                  }"
                  size="small"
                >
                  <template #columns>
                    <TableColumn
                      title="algorithm"
                      data-index="algo"
                      :width="160"
                    >
                      <template #cell="{ record: n }">
                        <code>{{ n.algo }}</code>
                      </template>
                    </TableColumn>
                    <TableColumn
                      title="模式"
                      data-index="mode"
                      :width="200"
                    ></TableColumn>
                    <TableColumn
                      title="密钥来源"
                      data-index="key"
                    ></TableColumn>
                  </template>
                </Table>
                <h3 class="doc-h3">
                  信封形状(请求 / 响应一致,只有业务体被加密)
                </h3>
                <ul>
                  <li>
                    请求体:
                    <code
                      >{&quot;data&quot;:
                      &quot;&lt;编码后的密文&gt;&quot;}</code
                    >
                  </li>
                  <li>
                    响应体:
                    <code>
                      {&quot;code&quot;:0,&quot;msg&quot;:&quot;ok&quot;,&quot;data&quot;:&quot;&lt;编码后的密文&gt;&quot;}
                    </code>
                    (只有
                    <code>data</code>
                    是密文)
                  </li>
                </ul>
                <p>
                  <b>编码</b>
                  <code>encoding</code>
                  =
                  <code>base64</code>
                  (默认)或
                  <code>hex</code>
                  ,请求/响应/握手三处一致。
                </p>
                <ul>
                  <li>
                    <b>AES</b>
                    :CBC + PKCS7,密文 =
                    <code>IV(16) || ciphertext</code>
                    ,
                    <code>sym_key</code>
                    支持 hex/base64/raw
                  </li>
                  <li>
                    <b>DES</b>
                    :CBC + PKCS7,密文 =
                    <code>IV(8) || ciphertext</code>
                  </li>
                  <li>
                    <b>RC4</b>
                    :流式,无 IV
                  </li>
                </ul>
                <Alert type="info">
                  <template #default>
                    不加密时直接发明文 JSON,响应
                    <code>data</code>
                    也是明文对象,本节可整段跳过。
                  </template>
                </Alert>
              </section>
              <section v-show="r === 'handshake'" class="panel">
                <h2 class="doc-h2">握手(session 模式)</h2>
                <p>
                  仅当
                  <code>algorithm = session</code>
                  时需要。启动时先握手换
                  <code>sessionId</code>
                  ,之后业务请求带
                  <code>X-Session-Id</code>
                  ,服务端用 per-session 的 AES key 解请求 / 加响应。session 在
                  Redis 里
                  <b>24h 滑动过期</b>
                  。
                </p>
                <h3 class="doc-h3">
                  握手请求 POST /api/sdk/handshake（按配置签名，请求不套 data
                  信封）
                </h3>
                <ol>
                  <li>
                    客户端随机生成 32 字节 AES key
                    <code>X</code>
                    。
                  </li>
                  <li>
                    用 app 的
                    <code>server_public_key</code>
                    对
                    <code>X</code>
                    做
                    <b>RSA/PKCS1v15</b>
                    加密,按 encoding 编码:
                  </li>
                </ol>
                <DocCode
                  ><code
                    >{ &quot;key&quot;: &quot;&lt;encode(RSA(X))&gt;&quot;
                    }</code
                  ></DocCode
                >
                <h3 class="doc-h3">握手响应</h3>
                <DocCode
                  ><code>{{ Te }}</code></DocCode
                >
                <ul>
                  <li>
                    <code>data</code>
                    是用刚才那把
                    <code>X</code>
                    做 AES-CBC(
                    <code>IV(16)||cipher</code>
                    )加密的
                    <code>{&quot;sessionId&quot;:&quot;...&quot;}</code>
                    ,再按 encoding 编码。
                  </li>
                  <li>
                    客户端用本地
                    <code>X</code>
                    解出
                    <code>sessionId</code>
                    保存,之后业务请求头带
                    <code>X-Session-Id</code>
                    ,body 用
                    <code>X</code>
                    走 AES 信封。
                  </li>
                  <li>
                    session 失效(过期 / Redis 会话丢失 / 跨 app)→ 业务接口返回
                    <b>1020</b>
                    ,客户端据此重新握手。
                  </li>
                </ul>
              </section>
              <section v-show="r === 'diff'" class="panel">
                <h2 class="doc-h2">差量更新机制</h2>
                <p>
                  服务端基于管理员上传的
                  <b>canonical APK</b>
                  按需生成「随机包名 clone」,以
                  <code>canonical → clone</code>
                  的
                  <b>DELTA1 bsdiff 补丁</b>
                  下发(几 KB vs 完整 APK 几 MB)。客户端用本地 canonical 字节 +
                  补丁跑 bspatch 还原 clone,直接安装换包名。
                  相关接口见左侧「升级接口」下的
                  <Link @click="(n) => b('clone')">换包名</Link>
                  、
                  <Link @click="(n) => b('regen')">重置签名</Link>
                  、
                  <Link @click="(n) => b('diffdl')">下补丁</Link>
                  。
                </p>
                <h3 class="doc-h3">DELTA1 补丁格式</h3>
                <DocCode
                  ><code>{{ Ce }}</code></DocCode
                >
                <p>
                  control 流由 24B 三元组
                  <code>(x, y, z)</code>
                  (bsdiff offtin 8 字节有符号整数,LE)组成:
                </p>
                <ul>
                  <li>
                    <code>x</code>
                    — 从
                    <code>old[oldPos..]</code>
                    拷 x 字节,
                    <b>逐字节加上 diff 流对应字节</b>
                  </li>
                  <li>
                    <code>y</code>
                    — 从 extra 流追加 y 字节
                  </li>
                  <li>
                    <code>z</code>
                    —
                    <code>oldPos += z</code>
                    (有符号)
                  </li>
                </ul>
                <p>
                  三流加和直到产出
                  <code>newSize</code>
                  字节。zlib 是带头/尾的标准 zlib(
                  <code>wbits=15</code>
                  ),Kotlin
                  <code>Inflater()</code>
                  默认即吃。
                </p>
                <h3 class="doc-h3">Go 参考实现</h3>
                <DocCode
                  ><code>{{ qe }}</code></DocCode
                >
                <p>
                  仓库自带
                  <code>cmd/bspatch</code>
                  命令行工具,可作调试 + 参考:
                </p>
                <DocCode
                  ><code>{{
                    `go build -o bspatch ./cmd/bspatch
./bspatch canonical.apk patch.diff clone.apk`
                  }}</code></DocCode
                >
                <h3 class="doc-h3">Kotlin DELTA1 参考实现</h3>
                <p>
                  下列代码用于说明格式；正式客户端还应限制包大小、校验各段边界和控制记录，并拒绝截断或非法压缩流。DELTA1
                  不能直接交给标准 BSDIFF40 补丁工具。
                </p>
                <DocCode
                  ><code>{{ Be }}</code></DocCode
                >
                <h3 class="doc-h3">还原后自检 + 安装</h3>
                <ol>
                  <li>
                    算还原结果 SHA-256,与接口返回的
                    <code>cloneSha256</code>
                    (或
                    <code>/ini</code>
                    的
                    <code>diffNewSha</code>
                    )比对。
                  </li>
                  <li>
                    合成的 clone APK 已带 V2 签名,可
                    <code>apksigner verify</code>
                    或交给系统装时自动校验。
                  </li>
                  <li>
                    用
                    <code>Intent.ACTION_VIEW</code>
                    /
                    <code>PackageInstaller</code>
                    安装;
                    <code>FileProvider</code>
                    + targetSdk≥24 用
                    <code>content://</code>
                    URI。
                  </li>
                </ol>
                <Alert type="info">
                  <template #default>
                    <b>差量更新前提</b>
                    :bspatch 只对
                    <b>字节完全一致</b>
                    的源有效。客户端若开了随机包名,本地 APK 已不是 canonical,
                    不能直接 patch 升级——需保留一份 canonical
                    原始字节,或在启用差量更新时关掉随机包名。
                  </template>
                </Alert>
              </section>
              <section v-show="r === 'errors'" class="panel">
                <h2 class="doc-h2">错误码</h2>
                <Table
                  :data="tableRows3"
                  :pagination="!1"
                  :bordered="{
                    wrapper: !0,
                    cell: !0,
                  }"
                  size="small"
                >
                  <template #columns>
                    <TableColumn
                      title="code"
                      data-index="code"
                      :width="70"
                    ></TableColumn>
                    <TableColumn
                      title="含义"
                      data-index="msg"
                      :width="260"
                    ></TableColumn>
                    <TableColumn
                      title="触发场景"
                      data-index="cause"
                    ></TableColumn>
                  </template>
                </Table>
              </section>
              <section v-show="r === 'flow'" class="panel">
                <h2 class="doc-h2">启动流程</h2>
                <DocCode
                  ><code>{{ Ke }}</code></DocCode
                >
              </section>
              <template v-if="a">
                <section :key="0" class="panel">
                  <header class="panel-head">
                    <h2 class="panel-title">
                      <Tag
                        :color="a.info.method === 'GET' ? 'arcoblue' : 'green'"
                        >{{ a.info.method }}</Tag
                      >
                      {{ " " + a.title }}
                    </h2>
                  </header>
                  <div class="endpoint-strip">
                    <Tag
                      :color="a.info.method === 'GET' ? 'arcoblue' : 'green'"
                      >{{ a.info.method }}</Tag
                    ><code>{{ a.info.path }}</code>
                  </div>
                  <Descriptions
                    class="api-facts"
                    bordered
                    size="small"
                    :column="{ xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 2 }"
                  >
                    <DescriptionsItem label="请求格式">{{
                      a.info.format
                    }}</DescriptionsItem>
                    <DescriptionsItem label="登录令牌">{{
                      a.info.token
                    }}</DescriptionsItem>
                    <DescriptionsItem label="数据加密">{{
                      a.info.crypto
                    }}</DescriptionsItem>
                    <DescriptionsItem label="请求签名">{{
                      a.info.sign
                    }}</DescriptionsItem>
                  </Descriptions>
                  <template v-if="a.intro">
                    <p :key="0" class="intro" :innerHTML="a.intro"></p>
                  </template>
                  <template v-else></template>
                  <h3 class="doc-h3">参数说明</h3>
                  <div v-if="a.params?.length" class="field-scroll">
                    <DocTable
                      :headers="['参数 / 位置', '类型', '必填', '说明']"
                      :rows="a.params"
                      :code-column="0"
                    />
                  </div>
                  <p v-else>
                    无业务参数。JSON 业务接口建议传 {}；启用加密时仍需把 {}
                    加密后发送信封。GET、multipart 和原始接口遵循上方数据格式。
                  </p>
                  <template v-if="a.reqExamples?.length">
                    <h3 class="doc-h3">请求参数</h3>
                    <template v-for="(n, p) in a.reqExamples" :key="p">
                      <div>
                        <template v-if="n.label">
                          <p :key="0" class="ex-label">{{ n.label }}</p>
                        </template>
                        <template v-else></template>
                        <DocCode
                          ><code>{{ n.code }}</code></DocCode
                        >
                      </div>
                    </template>
                  </template>
                  <template v-else></template>
                  <template v-if="a.notes?.length">
                    <ul :key="2" class="notes">
                      <template v-for="(n, p) in a.notes" :key="p"
                        ><li :innerHTML="n"></li
                      ></template>
                    </ul>
                  </template>
                  <template v-else></template>
                  <h3 class="doc-h3">返回字段（data）</h3>
                  <div class="field-scroll">
                    <DocTable
                      :headers="['字段', '类型', '说明']"
                      :rows="a.returns"
                      :code-column="0"
                    />
                  </div>
                  <h3 class="doc-h3">响应示例</h3>
                  <Tabs type="rounded" size="small">
                    <template #default>
                      <TabPane key="ok" title="成功">
                        <template #default>
                          <DocCode
                            ><code>{{ a.success }}</code></DocCode
                          >
                        </template>
                      </TabPane>
                      <template v-if="a.fail">
                        <TabPane key="fail" title="失败">
                          <template #default>
                            <DocCode
                              ><code>{{ a.fail }}</code></DocCode
                            >
                            <template v-if="a.failNote">
                              <p
                                :key="0"
                                class="fail-note"
                                :innerHTML="a.failNote"
                              ></p>
                            </template>
                            <template v-else></template>
                          </template>
                        </TabPane>
                      </template>
                      <template v-else></template>
                    </template>
                  </Tabs>
                </section>
              </template>
              <template v-else></template>
              <footer class="doc-pagination">
                <Button
                  v-if="neighbors.previous"
                  type="text"
                  @click="b(neighbors.previous.key)"
                >
                  <small>← 上一篇</small
                  ><strong>{{ neighbors.previous.title }}</strong></Button
                ><Button
                  v-if="neighbors.next"
                  type="text"
                  class="next-page"
                  @click="b(neighbors.next.key)"
                >
                  <small>下一篇 →</small
                  ><strong>{{ neighbors.next.title }}</strong>
                </Button>
              </footer>
            </div>
            <aside class="docs-outline" aria-label="本页目录">
              <span class="outline-label">本页内容</span>
              <Menu
                v-if="outline.length"
                :selected-keys="[activeHeading]"
                @menu-item-click="
                  (key) => jumpTo(outline.find((item) => item.id === key))
                "
              >
                <MenuItem
                  v-for="item in outline"
                  :key="item.id"
                  class="outline-item"
                  >{{ item.title }}</MenuItem
                >
              </Menu>
              <p v-else>{{ currentItem?.title }}</p>
              <div class="outline-help">
                对接遇到问题？<Button type="text" @click="b('errors')">
                  查看错误码说明 ↗
                </Button>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </Card>
  </div>
</template>
<style scoped>
.docs-page {
  --doc-blue: rgb(var(--primary-6));
  --doc-tint: var(--color-fill-1);
  --doc-line: var(--color-border-2);
  height: calc(100dvh - 56px);
  min-height: 500px;
  max-width: 1800px;
  margin: auto;
  padding: 0 24px 20px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  color: var(--color-text-1);
}
button,
input {
  font: inherit;
}
button {
  cursor: pointer;
}
button:focus-visible,
input:focus-visible {
  outline: 2px solid var(--doc-blue);
  outline-offset: 3px;
}
.docs-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  min-height: 86px;
  flex-shrink: 0;
}
.docs-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}
.docs-brand-icon {
  width: 40px;
  height: 40px;
  background: var(--doc-blue);
  color: #fff;
  display: grid;
  place-items: center;
  border-radius: 11px;
  font:
    600 21px Consolas,
    monospace;
  box-shadow: 0 5px 14px #2563eb22;
}
.docs-brand h1 {
  font-size: 19px;
  margin: 0 0 4px;
  font-weight: 650;
  letter-spacing: 0.3px;
}
.docs-brand div > span {
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 1.4px;
  color: var(--color-text-3);
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 24px;
}
.version-pill {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  color: var(--color-text-3);
}
.version-pill i,
.nav-foot-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3b82f6;
}
.back-link {
  padding: 7px 11px;
  border: 1px solid var(--doc-line);
  border-radius: 7px;
  background: var(--color-bg-2);
  color: var(--color-text-2);
  font-size: 12px;
}
.back-link span {
  color: var(--doc-blue);
  margin-left: 10px;
}
.docs-card {
  flex: 1;
  min-height: 0;
  background: var(--color-bg-2);
  border: 1px solid var(--doc-line);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 24px #14294b04;
}
.docs-layout {
  display: flex;
  height: 100%;
}
.docs-nav {
  --color-menu-light-bg: transparent;
  width: 242px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--doc-tint);
  border-right: 1px solid var(--doc-line);
}
.nav-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 18px 13px;
  font-size: 13px;
  font-weight: 600;
}
.nav-count {
  color: var(--color-text-4);
  font-size: 8px;
  letter-spacing: 1px;
}
.nav-search {
  margin: 0 14px 12px;
}

.nav-scroll {
  overflow: auto;
  min-height: 0;
  padding: 0 10px 18px;
  flex: 1;
}

.nav-item-title {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.method-tag {
  flex-shrink: 0;
  display: inline-block;
  font-family: Consolas, monospace;
  font-size: 9px;
  font-weight: 700;
  padding: 2px 5px;
  border-radius: 4px;
  letter-spacing: 0.2px;
}
.method-tag.post {
  background: #10b98112;
  color: #06865f;
}
.method-tag.get {
  background: #2563eb12;
  color: #3975e3;
}
.method-tag.big {
  font-size: 11px;
  margin-right: 10px;
  padding: 4px 8px;
}
.nav-empty {
  padding: 28px 8px;
  text-align: center;
  color: var(--color-text-3);
  font-size: 12px;
}
.nav-empty button {
  display: block;
  margin: 12px auto 0;
  border: 0;
  background: transparent;
  color: var(--doc-blue);
}
.nav-foot {
  padding: 14px 18px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  color: var(--color-text-3);
  border-top: 1px solid var(--doc-line);
}
.nav-foot code {
  font-size: 10px;
}
.mobile-menu {
  display: none;
}
.reading-area {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.reading-toolbar {
  min-height: 53px;
  flex-shrink: 0;
  padding: 0 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--doc-line);
  font-size: 11px;
  color: var(--color-text-3);
}
.toolbar-divider {
  margin: 0 8px;
  color: var(--color-text-4);
}
.reading-toolbar strong {
  font-weight: 500;
  color: var(--color-text-2);
}
.reading-label {
  font-size: 12px;
  white-space: nowrap;
}
.reading-columns {
  flex: 1;
  display: flex;
  min-height: 0;
}
.docs-content {
  min-width: 0;
  flex: 1;
  padding: 28px 30px 20px;
  overflow-y: auto;
  overscroll-behavior: contain;
  scroll-padding-top: 24px;
  scrollbar-width: thin;
}
.docs-outline {
  width: 166px;
  margin: 28px 18px 20px 0;
  padding-left: 18px;
  border-left: 1px solid var(--doc-line);
  overflow-y: auto;
  flex-shrink: 0;
  font-size: 11px;
}
.outline-label {
  display: block;
  color: var(--color-text-3);
  font-weight: 600;
  margin-bottom: 13px;
}

.docs-outline p {
  color: var(--color-text-3);
}
.outline-help {
  margin-top: 26px;
  padding-top: 17px;
  border-top: 1px solid var(--doc-line);
  color: var(--color-text-4);
  line-height: 1.8;
}
.outline-help button {
  display: block;
  margin-top: 6px;
  border: 0;
  background: transparent;
  padding: 0;
  color: var(--doc-blue);
  font-size: 11px;
}
.panel {
  font-size: 14px;
  padding-bottom: 8px;
}
.doc-h2,
.panel-title {
  font-size: 26px;
  font-weight: 650;
  line-height: 1.5;
  color: var(--color-text-1);
  margin: 0 0 14px;
  letter-spacing: -0.5px;
}
.panel-title {
  display: flex;
  align-items: center;
}
.doc-h3 {
  font-size: 15px;
  margin: 28px 0 12px;
  color: var(--color-text-1);
  font-weight: 650;
}
.panel p,
.panel ul,
.panel ol {
  font-size: 14px;
  color: var(--color-text-2);
  line-height: 1.95;
  margin: 10px 0;
}
.panel ul,
.panel ol {
  padding-left: 22px;
}
.panel li {
  margin: 5px 0;
}
.panel a {
  color: var(--doc-blue);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: #2563eb40;
  text-underline-offset: 3px;
}
.panel code {
  color: var(--doc-blue);
  background: var(--doc-tint);
  padding: 2px 4px;
  border-radius: 4px;
  font:
    12px Consolas,
    monospace;
  overflow-wrap: anywhere;
}
.panel :deep(.doc-code code) {
  color: #dfe9fa;
  background: transparent;
  padding: 0;
  overflow-wrap: normal;
}
.overview-label {
  font:
    10px Consolas,
    monospace;
  letter-spacing: 1.5px;
  color: var(--doc-blue);
  margin-bottom: 12px;
}
.overview-title {
  max-width: 550px;
  font-size: 29px;
  letter-spacing: -1px;
  line-height: 1.5;
}
.panel .overview-description {
  color: var(--color-text-2);
  font-size: 14px;
  margin-bottom: 20px;
}
.quick-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 20px 0 28px;
}
.quick-links :deep(.arco-btn) {
  height: auto;
  white-space: normal;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  background: var(--color-bg-2);
  border: 1px solid var(--doc-line);
  border-radius: 9px;
  padding: 16px;
  text-align: left;
  color: var(--color-text-1);
  transition:
    border-color 0.15s,
    background 0.15s;
}
.quick-links :deep(.arco-btn):hover {
  border-color: #2563eb66;
  background: var(--doc-tint);
}
.quick-icon {
  font:
    11px Consolas,
    monospace;
  color: var(--doc-blue);
}
.quick-links strong {
  font-size: 13px;
  display: flex;
  justify-content: space-between;
}
.quick-links strong span {
  color: var(--doc-blue);
  font-weight: 400;
}
.quick-links small {
  color: var(--color-text-3);
  font-size: 11px;
  line-height: 1.6;
}
.endpoint-strip {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  border: 1px solid #2563eb26;
  background: var(--doc-tint);
  border-radius: 8px;
  margin: 12px 0 16px;
}
.endpoint-strip code {
  color: var(--color-text-1);
  background: transparent;
  padding: 0;
  font-size: 13px;
}
.endpoint-strip .method-tag {
  font-size: 11px;
  padding: 3px 7px;
}

.info-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.field-table {
  min-width: 480px;
}
.info-table td,
.info-table th {
  padding: 12px 14px;
  text-align: left;
  vertical-align: top;
  line-height: 1.85;
  border-bottom: 1px solid var(--doc-line);
}
.info-table th {
  background: var(--doc-tint);
  color: var(--color-text-2);
  font-weight: 600;
  font-size: 11px;
  white-space: nowrap;
}
.info-table td {
  color: var(--color-text-2);
}
.info-table tbody tr:last-child td {
  border-bottom: 0;
}
.info-table tbody tr:hover {
  background: #2563eb04;
}
.info-table td:first-child {
  min-width: 110px;
  overflow-wrap: anywhere;
}
.info-table td:last-child {
  min-width: 170px;
}
.ex-label {
  font-size: 11px;
  color: var(--color-text-3);
  margin: 14px 0 6px;
}
.fail-note {
  color: var(--color-text-3);
  font-size: 12px;
}
.doc-pagination {
  display: flex;
  gap: 14px;
  margin-top: 28px;
  padding: 22px 0 0;
  border-top: 1px solid var(--doc-line);
}
.doc-pagination :deep(.arco-btn) {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
  height: auto;
  white-space: normal;
  background: transparent;
  border: 1px solid var(--doc-line);
  border-radius: 8px;
  padding: 12px 15px;
  text-align: left;
  min-width: 0;
  width: 50%;
  color: var(--color-text-1);
}
.doc-pagination :deep(.arco-btn):hover {
  border-color: #2563eb55;
  background: var(--doc-tint);
}
.doc-pagination small {
  display: block;
  font-size: 12px;
  color: var(--color-text-3);
  margin: 0;
}
.doc-pagination strong {
  font-size: 13px;
  font-weight: 500;
}
.doc-pagination .next-page {
  text-align: right;
  margin-left: auto;
}
.docs-content :deep(.cloud-docs),
.docs-content :deep(.sdk-guide) {
  font-size: 14px;
  line-height: 1.95;
}
.docs-content :deep(.cloud-docs h2),
.docs-content :deep(.sdk-guide h2) {
  font-size: 26px;
  font-weight: 650;
  margin: 0 0 16px;
}
.docs-content :deep(.cloud-docs h3),
.docs-content :deep(.sdk-guide h3) {
  font-size: 15px;
  margin: 28px 0 12px;
}
.docs-content :deep(.cloud-docs p),
.docs-content :deep(.sdk-guide p),
.docs-content :deep(.cloud-docs li),
.docs-content :deep(.sdk-guide li) {
  color: var(--color-text-2);
}
.docs-content :deep(.arco-table) {
  font-size: 13px;
}
.docs-content :deep(.arco-table-container) {
  overflow: auto;
}
.docs-content :deep(.arco-alert) {
  border-radius: 8px;
  font-size: 12px;
}
.docs-content :deep(.arco-tabs-nav) {
  margin-top: 12px;
}
.docs-content :deep(h3:focus) {
  outline: none;
}
:global(body[arco-theme="dark"] .docs-page) {
  --doc-blue: #80aaff;
  --doc-tint: #729cff0b;
}
.docs-content :deep(.cloud-docs p code),
.docs-content :deep(.cloud-docs td code),
.docs-content :deep(.sdk-guide p code) {
  color: var(--doc-blue);
}
:global(body[arco-theme="dark"] .method-tag.post) {
  color: #67d9b3;
}
@media (max-width: 1250px) {
  .docs-outline {
    display: none;
  }
  .docs-content {
    padding: 24px;
  }
  .docs-nav {
    width: 225px;
  }
}
@media (max-width: 760px) {
  .docs-page {
    padding: 0 12px 16px;
    height: auto;
    min-height: calc(100dvh - 56px);
  }
  .docs-topbar {
    min-height: 78px;
  }
  .docs-brand {
    gap: 9px;
  }
  .docs-brand h1 {
    font-size: 17px;
  }
  .docs-brand-icon {
    width: 34px;
    height: 34px;
    font-size: 18px;
  }
  .docs-brand div > span {
    font-size: 7px;
    letter-spacing: 0.7px;
  }
  .version-pill {
    display: none;
  }
  .topbar-right {
    gap: 0;
  }
  .back-link {
    padding: 6px 8px;
    font-size: 11px;
  }
  .docs-card {
    flex: none;
    overflow: visible;
  }
  .docs-layout {
    flex-direction: column;
  }
  .docs-nav {
    width: 100%;
    border-right: 0;
    border-bottom: 1px solid var(--doc-line);
    border-radius: 12px 12px 0 0;
  }
  .nav-heading {
    padding: 14px 16px 10px;
  }
  .nav-count,
  .nav-foot {
    display: none;
  }
  .mobile-menu {
    display: block;
    font-size: 11px;
    background: transparent;
    border: 0;
    color: var(--doc-blue);
  }
  .nav-search {
    margin-bottom: 12px;
  }
  .nav-scroll {
    display: none;
  }
  .nav-scroll.expanded {
    display: block;
    max-height: 42dvh;
    flex: none;
  }
  .reading-toolbar {
    padding: 12px 16px;
    min-height: 0;
    font-size: 10px;
    line-height: 1.8;
  }
  .reading-label {
    display: none;
  }
  .toolbar-divider {
    margin: 0 4px;
  }
  .reading-columns {
    display: block;
  }
  .docs-content {
    overflow: visible;
    padding: 22px 16px;
  }
  .doc-h2,
  .panel-title,
  .docs-content :deep(.cloud-docs h2),
  .docs-content :deep(.sdk-guide h2) {
    font-size: 23px;
  }
  .overview-title {
    font-size: 25px;
  }
  .quick-links {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .quick-links :deep(.arco-btn) {
    height: auto;
    white-space: normal;
    padding: 12px;
    display: grid;
    grid-template-columns: 24px 1fr;
    gap: 5px 8px;
  }
  .quick-icon {
    grid-row: 1 / 3;
    align-self: center;
  }
  .quick-links small {
    grid-column: 2;
  }

  .endpoint-strip {
    padding: 12px;
    gap: 8px;
  }
  .endpoint-strip code {
    font-size: 12px;
  }
}
@media (prefers-reduced-motion: reduce) {
  * {
    transition: none !important;
    scroll-behavior: auto !important;
  }
}

.nav-scroll :deep(.arco-menu) {
  background: transparent;
}
.nav-scroll :deep(.arco-menu-inner) {
  padding: 0;
}
.nav-scroll :deep(.arco-menu-item) {
  padding: 0 10px;
  display: flex;
  align-items: center;
  min-height: 38px;
  margin-bottom: 2px;
  font-size: 13px;
}
.nav-scroll :deep(.arco-menu-group-title) {
  font-size: 12px;
  line-height: 32px;
  padding: 8px 10px 0;
}
.docs-nav :deep(.arco-menu-selected),
.docs-outline :deep(.arco-menu-selected) {
  background: var(--color-primary-light-1);
}
.nav-scroll :deep(.arco-menu-indent-list) {
  display: none;
}
.nav-item :deep(.arco-menu-item-inner) {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  min-width: 0;
}
.nav-item :deep(.arco-tag) {
  flex-shrink: 0;
}
.docs-outline :deep(.arco-menu-inner) {
  padding: 0;
}
.docs-outline :deep(.arco-menu-item) {
  padding: 7px 8px;
  line-height: 1.6;
  height: auto;
  white-space: normal;
}
.reading-toolbar :deep(.arco-breadcrumb) {
  font-size: 12px;
}

.api-facts {
  margin-bottom: 18px;
}
.panel-title > .arco-tag {
  margin-right: 10px;
}
</style>
