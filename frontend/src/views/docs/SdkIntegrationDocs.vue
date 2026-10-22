<script>
import { Alert } from "@arco-design/web-vue";
import DocCode from "./DocCode.vue";
import DocTable from "./DocTable.vue";
import source from "./sdk-client-example.js?raw";
export default {
  components: { DocCode, DocTable, Alert },
  name: "SdkIntegrationDocs",
  setup() {
    const groups = [
      ["handshake", "POST", "否", "按 tokenKey 配置", "专用握手协议"],
      [
        "logon / register / ini",
        "POST",
        "否",
        "按 tokenKey 配置",
        "按加密方案配置",
      ],
      [
        "info / heartbeat / logout / updateInfo / userExtends / kamiTopup / cloudFunction",
        "POST",
        "是",
        "按 tokenKey 配置",
        "按加密方案配置",
      ],
      ["uploadAvatar", "POST multipart", "是", "否", "否"],
      ["clone / regen-sign", "GET / POST", "是", "否", "否"],
      [
        "diff/:name / extend-file/list / extend-file/:id/download",
        "GET",
        "否",
        "否",
        "否",
      ],
    ];
    return { source, groups };
  },
};
</script>
<template>
  <section class="sdk-guide">
    <h2>客户端 API 接入指南</h2>
    <p>
      本文覆盖当前后端的 17 个 SDK 接口。默认前缀为
      <code>/api/sdk</code
      >，具体接口参数和返回字段见左侧导航。后台自定义路径或错误码后，客户端必须同步配置；这里展示默认值。
    </p>
    <h3>1. 接口与鉴权</h3>
    <div class="scroll">
      <DocTable
        :headers="[
          '接口（省略前缀）',
          '方法',
          '登录 token',
          'HMAC 签名',
          'JSON 加密信封',
        ]"
        :rows="groups"
        :code-column="-1"
      />
    </div>
    <p>
      所有 SDK 接口均检查
      <code>X-App-Id</code>、应用是否启用、限流及按配置开启的 mTLS。需 token
      的接口还检查账户状态、设备绑定、地域和 VIP。APK
      换包、重置签名仅支持普通账号。
    </p>
    <p>
      业务请求使用客户端登录得到的 X-Token（优先）或 Authorization:
      Bearer，不使用后台管理员令牌。mTLS 必须在 TLS 连接上携带当前应用 CA
      签发的客户端证书，CN 为 app-{appid}；不是添加一个普通 HTTP 请求头。
    </p>
    <h3>2. 推荐调用顺序</h3>
    <ol>
      <li>
        从自己的应用配置取得服务地址、应用
        ID、tokenKey、加密方案以及实际接口路径。
      </li>
      <li>session 模式先握手，保存 AES 密钥和 sessionId；其他模式跳过。</li>
      <li>
        调用 ini 获取公告、应用扩展和版本。latestVersion 可能为
        null，更新补丁字段也可能缺失。
      </li>
      <li>
        普通账号调用 register 或 logon；卡密账号调用 logon 并传
        card。绑定设备时，登录 JSON 的 mi 与后续 X-Mi 保持一致。
      </li>
      <li>
        保存返回的 token，按需调用信息、扩展、充值、云函数等接口。空参数的 JSON
        请求统一传 {}，加密启用时仍需加密它。
      </li>
      <li>
        收到 1006 清除本地 token 并重新登录；1020 重新握手。退出时调用
        logout，成功后清除本地 token。
      </li>
    </ol>
    <p>
      登录 token 按应用的 userTokenTtl 滑动过期（创建时未配置则默认 24
      小时），sessionId 是另一套 24
      小时滑动过期会话。心跳会返回服务器时间；网络异常时不要把本地缓存的
      vipActive 当作本次服务端授权成功。
    </p>
    <Alert
      >在线人数按账号汇总最近 5
      分钟内的有效会话活动，短期令牌按较短有效期计算。一台客户端退出不会使其他会话下线；删除账号、改密码或解绑后，客户端在下一次请求或心跳收到失效响应，应清除本地登录状态。</Alert
    >
    <h3>3. 请求与响应约定</h3>
    <ul>
      <li>
        JSON 使用 UTF-8，字段名按接口表大小写。业务时间戳与 X-Ts 为 Unix
        秒；云函数 timeout、duration 则为毫秒。
      </li>
      <li>
        未启用加密时直接发送业务 JSON。启用时先加密业务 JSON，包装成
        {"data":"密文"}，最后对实际发送字节做 HMAC 签名。签名后不能再次格式化
        JSON。
      </li>
      <li>
        签名消息为 X-Ts + 换行 + X-Nonce + 换行 + 请求体；换行是一个 LF
        字节，不是字面量反斜杠 n。tokenKey 作为原始 UTF-8
        字符串使用，不自行解码成 hex。
      </li>
      <li>
        nonce 每次请求重新生成（最长 256 字节），在应用范围防重放。时间窗默认
        ±60 秒，可配置但最多 ±3600 秒。签名接口请求体最多 2 MiB。
      </li>
      <li>
        JSON 业务成功以 code=0 判断；HTTP 200
        也可能是失败。启用加密仅加密响应的非 null data，code/msg
        仍为明文；错误可能没有 data，不能一律解密。
      </li>
      <li>
        文件和补丁成功响应为二进制，不使用 JSON 信封。先检查 HTTP
        状态和附件头，再区分错误信封；JSON 文件本身也可能是合法附件。
      </li>
      <li>
        客户端下载到的相对 URL
        用服务端源站解析。自定义接口后使用实际映射；版本接口返回的默认 diffUrl
        也应核对下载路径配置。
      </li>
    </ul>
    <h3>4. 可以运行的 Node.js 示例</h3>
    <p>
      示例支持未加密及 AES-CBC 模式，包含
      HMAC、设备头、登录令牌和错误处理。保存为 .mjs 后使用 Node.js 18+ 运行；AES
      模式传入解码后的 aesKey Buffer。session、RSA、DES、RC4 和 mTLS
      需按对应配置另行接入，不能直接套用此示例。
    </p>
    <DocCode
      ><code>{{ source }}</code></DocCode
    >
    <h3>5. 常见接入问题</h3>
    <ul>
      <li>
        注册返回 400：账号要求 5–12 字节、密码 6–18 字节；卡密应用不支持注册。
      </li>
      <li>
        1014：核对 tokenKey、请求体字节、nonce
        是否复用；1016：核对设备时钟。自动重试也必须重新生成签名头。
      </li>
      <li>
        1015：检查加密算法、密钥、base64/hex 编码与 IV 前缀；AES/DES 使用 CBC +
        PKCS7。RSA 使用 PKCS#1 v1.5 分块（每块明文最多模数长度减 11 字节），不是
        OAEP。
      </li>
      <li>
        1006：令牌过期、跨应用、密码已更改或绑定设备已移除；绑定设备时缺 X-Mi
        或和登录 mi 不一致也会失败。
      </li>
      <li>
        1018：VIP 模式到期。持有有效令牌时充值和登出可豁免 VIP
        到期检查，但不豁免其他鉴权。
      </li>
      <li>
        1017：按退避策略稍后再试；充值、注册、重置 APK
        签名及有副作用的云函数不要盲目自动重试。当前没有通用幂等键协议。
      </li>
      <li>
        扩展变量值必须为字符串，读用户扩展也必须发送
        {}。客户端没有修改密码、查询/解绑设备、管理用户分组或删除扩展键的 SDK
        接口，应由后台管理。
      </li>
    </ul>
  </section>
</template>
<style scoped>
.sdk-guide {
  color: var(--color-text-1);
  line-height: 1.85;
  font-size: 13px;
}
h2 {
  margin: 4px 0 12px;
  font-size: 23px;
}
h3 {
  margin: 24px 0 10px;
  font-size: 17px;
}
p,
li {
  margin: 8px 0;
}
ul,
ol {
  padding-left: 23px;
}
.scroll,
pre {
  overflow-x: auto;
}
table {
  width: 100%;
  min-width: 620px;
  border-collapse: collapse;
}
td,
th {
  padding: 9px 12px;
  text-align: left;
  border: 1px solid var(--color-border-2);
}
th {
  background: var(--color-fill-2);
}
code {
  font-family: Consolas, monospace;
  font-size: 12px;
}
pre {
  padding: 16px;
  border-radius: 8px;
  background: #142139;
  color: #e4ecfc;
  line-height: 1.7;
}
</style>
