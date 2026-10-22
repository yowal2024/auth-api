<script>
import DocCode from "./DocCode.vue";
import DocTable from "./DocTable.vue";
import { cloudExamples } from "./cloud-examples.js";
export default {
  components: { DocCode, DocTable },
  name: "CloudFunctionDocs",
  setup() {
    const triggers = [
      [
        "manual",
        "手动调用",
        "客户端按函数名调用已启用的 manual 函数。后台试运行始终使用 manual 上下文，可测试草稿或禁用函数。",
      ],
      [
        "before_login",
        "登录前",
        "账号登录或卡密登录的前置检查；可返回拒绝结果。不是替代内置密码、卡密及授权校验。",
      ],
      [
        "after_login",
        "登录后",
        "登录状态更新后执行；返回值不决定放行，执行失败不回滚已完成的操作。",
      ],
      [
        "before_register",
        "注册前",
        "创建账号前执行；user 为 null，可拒绝注册。卡密应用不走注册流程。",
      ],
      [
        "after_register",
        "注册后",
        "账号创建后执行；返回值被忽略，失败不会撤销新账号。",
      ],
      [
        "kami_use",
        "卡密充值后",
        "账号使用充值卡且事务提交后执行；与卡密账号登录不同，不能用于阻止充值。",
      ],
    ];
    const context = [
      ["ctx.app", "当前应用：id、name、pkg、mode、type；不包含 tokenKey。"],
      [
        "ctx.user",
        "后台试运行、注册前和卡密登录前为 null。SDK 手动调用包含 id、username、mi、region、ip、isKami、appid；其他触发器提供当时的用户资料，字段不保证相同。",
      ],
      [
        "ctx.payload",
        "SDK 手动调用对应请求 param；后台试运行对应 payload；自动触发器的结构见下表。",
      ],
      [
        "ctx.trigger",
        "当前触发器名称。后台试运行固定为 manual，不会模拟完整登录或注册流程。",
      ],
      [
        "ctx.operator",
        "后台试运行时为管理员账号；SDK 调用与自动触发时为空字符串。",
      ],
      ["ctx.ip", "当前请求的客户端 IP。"],
    ];
    const payloads = [
      ["账号登录前 / 后", "Username、Password、Mi（注意大写）"],
      ["账号注册前 / 后", "Username、Password、Nickname、Mi（注意大写）"],
      ["卡密登录前", "username、card、mi、appid、isKami"],
      ["卡密登录后", "id、username、appid、isKami"],
      ["卡密充值后", "card、vip、groupName、addSec"],
    ];
    const apis = [
      [
        "console.log / info / warn / error(...values)",
        "输出试运行日志；最多 200 条，每条最多保留 4096 字节。SDK 响应不包含这些日志。",
      ],
      [
        "db.query(sql, ...args)",
        "同步查询并返回对象数组。只接受以 SELECT 加空格开头、且不含分号的单条 SQL；以只读事务执行，超过 1000 行报错。没有 db.exec、写入和自动表前缀替换。",
      ],
      [
        "http.get(url)",
        "同步 GET。成功得到 { ok, status, body }，body 是字符串；网络失败得到 { ok: false, error }。HTTP 状态小于 400 时 ok=true。",
      ],
      [
        "http.post(url, body)",
        "同步 POST，body 自动 JSON 编码，Content-Type 为 application/json；返回结构同 GET。当前没有自定义请求头参数。",
      ],
      [
        "crypto.md5 / sha1 / sha256(text)",
        "对字符串的 UTF-8 字节计算摘要，返回小写十六进制字符串。",
      ],
      [
        "crypto.base64Encode / base64Decode(text)",
        "标准 Base64 字符串编码 / 解码；非法 Base64 会抛出异常。",
      ],
    ];
    return { triggers, context, payloads, apis, examples: cloudExamples };
  },
};
</script>
<template>
  <section class="cloud-docs" aria-labelledby="cloud-doc-title">
    <h2 id="cloud-doc-title">云函数开发指南</h2>
    <p>
      当前运行引擎为 <strong>Goja</strong>。每次执行创建独立 JavaScript
      运行环境，通过
      <code>handler(ctx)</code>
      接收上下文并返回可序列化结果。全局变量不会跨请求保存。
    </p>
    <h3>1. 创建与试运行</h3>
    <ol>
      <li>
        选择应用，在「云函数」中新增函数，填写名称、触发器、代码和超时（毫秒）。
      </li>
      <li>
        先试运行确认结果，再保存并启用。相同触发器按函数 ID
        升序执行；建议同一应用内使用唯一函数名。
      </li>
      <li>
        编辑器试运行当前代码但不自动保存；列表试运行已保存代码。两个页面当前均传入空对象
        payload，并且 user 为 null。
      </li>
    </ol>
    <DocCode
      ><code>{{ examples.basic }}</code></DocCode
    >
    <p>
      上例空参数试运行返回 Hello, visitor。SDK 调用时把 name 放进 param
      即可传入自定义名字。
    </p>
    <h3>2. 上下文与参数</h3>
    <div class="table-scroll">
      <DocTable :headers="['字段', '说明']" :rows="context" :code-column="0" />
    </div>
    <p>
      app、user、payload、trigger、operator、ip 也可作为全局变量访问；<code
        >ctx</code
      >
      本身是 handler 的参数，不是全局对象。修改上下文对象不会自动保存到数据库。
    </p>
    <div class="table-scroll">
      <DocTable
        :headers="['自动触发场景', 'payload 字段（当前实现）']"
        :rows="payloads"
        :code-column="1"
      />
    </div>
    <p>
      账号登录、注册参数直接映射为 Go 字段名，和 SDK 手动传入的 JSON
      大小写不同。不要直接输出整个 payload，其中可能含有密码或卡密。
    </p>
    <h3>3. 触发器与拦截规则</h3>
    <div class="table-scroll">
      <DocTable
        :headers="['触发器', '时机', '行为']"
        :rows="triggers"
        :code-column="0"
      />
    </div>
    <p>
      仅 before_ 触发器检查拒绝结果：返回
      <code>{ allow: false, msg: "原因" }</code> 会拒绝请求；也可返回非零整数
      code 和 msg。返回普通值或 allow=true
      继续后续检查。前置脚本抛错或超时会使请求失败。后置触发器的返回值不会替换
      SDK 响应。
    </p>
    <DocCode
      ><code>{{ examples.reject }}</code></DocCode
    >
    <h3>4. 内置 API</h3>
    <div class="table-scroll">
      <DocTable
        :headers="['调用方式', '返回值与限制']"
        :rows="apis"
        :code-column="0"
      />
    </div>
    <h4>数据库参数查询</h4>
    <DocCode
      ><code>{{ examples.database }}</code></DocCode
    >
    <p>
      云函数的查看、编辑、删除与试运行仅限超级管理员。查询权限不会自动限定到当前应用，读取业务表时应显式添加 appid
      条件。只读不代表数据隔离，因此只让受信任管理员编写脚本。
    </p>
    <h4>访问公网 HTTP 服务</h4>
    <DocCode
      ><code>{{ examples.http }}</code></DocCode
    >
    <p>
      仅支持 HTTP /
      HTTPS；内网、回环及特殊用途地址被拒绝，重定向同样检查。请求共用脚本执行期限。响应体最多读取
      2 MiB，超过部分被截断，不能把它当作大文件下载接口。
    </p>
    <h4>摘要与编码</h4>
    <DocCode
      ><code>{{ examples.crypto }}</code></DocCode
    >
    <h3>5. 客户端调用</h3>
    <p>
      先创建并启用名为 hello、触发器为 manual
      的函数，再使用客户端登录令牌调用。下面展示未启用数据加密时的业务请求；启用加密时，按「加密信封」处理请求体，再按实际发送字节签名。
    </p>
    <DocCode
      ><code>{{ examples.request }}</code></DocCode
    >
    <p>
      需要遵循当前应用的设备绑定、签名、加密及 mTLS 配置；启用设备绑定时带
      X-Mi。接口路径可被重定义，应使用实际配置。SDK
      不接受管理员令牌代替用户令牌。
    </p>
    <h3>6. 返回结果与错误</h3>
    <DocCode
      ><code>{{ examples.result }}</code></DocCode
    >
    <p>
      SDK 手动函数返回的对象即使含有 code 或 allow，也只是
      data，不会自动改变外层状态。需向客户端返回业务错误时，用同步 handler
      抛出带“业务错误:”前缀的 Error：
    </p>
    <DocCode
      ><code>{{ examples.error }}</code></DocCode
    >
    <p>
      其他脚本错误默认返回 500「执行失败」；不存在、被禁用或不是 manual
      的函数返回 404。错误码可被系统重定义。排错时检查后台试运行的 ok 和
      error，不能只看外层 code=0。
    </p>
    <h3>7. 运行边界</h3>
    <ul>
      <li>
        新建页面默认超时 1000ms；未传或非正超时使用 3000ms；最多
        30000ms。脚本源码最多 1 MiB，调用栈上限 1024。
      </li>
      <li>
        没有 Node.js / 浏览器环境，不提供 require、import 模块加载、fetch、DOM
        或 setTimeout；HTTP 请求使用内置 http 对象。
      </li>
      <li>
        建议使用普通 function handler。async 函数仅在返回 Promise
        已完成时可取到结果；没有事件循环，等待中的 Promise
        会报“异步任务未完成”。
      </li>
      <li>
        没有独立堆内存硬上限；不要运行不可信脚本或构造大量对象。超时控制不等于内存隔离。
      </li>
    </ul>
  </section>
</template>
<style scoped>
.cloud-docs {
  color: var(--color-text-1);
  font-size: 13px;
  line-height: 1.85;
  padding: 4px 0 16px;
}
h2 {
  font-size: 23px;
  margin: 0 0 12px;
}
h3 {
  font-size: 17px;
  margin: 26px 0 10px;
}
h4 {
  font-size: 14px;
  margin: 18px 0 8px;
}
p {
  margin: 10px 0;
}
ol,
ul {
  padding-left: 23px;
}
li {
  margin: 5px 0;
}
code {
  font-family: Consolas, monospace;
  font-size: 12px;
  overflow-wrap: anywhere;
}
p code,
td code {
  color: rgb(var(--primary-6));
}
pre {
  margin: 12px 0;
  padding: 16px;
  border-radius: 8px;
  background: #142139;
  color: #e4ecfc;
  overflow-x: auto;
  line-height: 1.7;
}
pre code {
  overflow-wrap: normal;
}
.table-scroll {
  overflow-x: auto;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th,
td {
  border: 1px solid var(--color-border-2);
  padding: 9px 12px;
  text-align: left;
  vertical-align: top;
}
th {
  background: var(--color-fill-2);
  font-weight: 600;
}
td:first-child {
  min-width: 150px;
}
td:last-child {
  min-width: 220px;
}
</style>
