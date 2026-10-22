<script>
import { Button, Card, Tag, Link } from "@arco-design/web-vue";
import {
  IconSafe,
  IconArrowRight,
  IconUserGroup,
  IconCode,
  IconStorage,
  IconCheck,
  IconApps,
} from "@arco-design/web-vue/es/icon";
export default {
  name: "Home",
  components: {
    Button,
    Card,
    Tag,
    Link,
    IconSafe,
    IconArrowRight,
    IconUserGroup,
    IconCode,
    IconStorage,
    IconCheck,
    IconApps,
  },
  setup() {
    const features = [
      {
        number: "01",
        icon: IconUserGroup,
        title: "让每一次访问，都有据可依。",
        label: "用户与授权",
        text: "统一管理账号、用户分组与设备绑定。从卡密发放到授权续期，让访问权限清晰可控。",
        tags: ["账号认证", "卡密管理", "设备绑定"],
        details: [
          "按应用管理账号与分组，设置会员有效期和封禁状态",
          "批量生成卡密，管理充值时长与使用记录",
          "查看绑定设备，按需解绑并调整绑定数量",
        ],
      },
      {
        number: "02",
        icon: IconCode,
        title: "把业务逻辑，交给云端。",
        label: "云函数与扩展",
        text: "用云函数处理业务逻辑，通过扩展变量与文件管理分发应用所需的内容。",
        tags: ["云端函数", "扩展变量", "文件分发"],
        details: [
          "编辑和试运行云函数，查看执行结果",
          "分别维护应用配置与用户扩展信息",
          "上传扩展文件，按分类管理与下载",
        ],
      },
      {
        number: "03",
        icon: IconStorage,
        title: "每一次更新，都更进一步。",
        label: "版本与更新",
        text: "集中维护应用版本、更新说明与安装包，通过差异补丁分发版本之间的变化。",
        tags: ["版本管理", "差异更新", "更新公告"],
        details: [
          "维护版本号、更新说明与 APK 文件",
          "客户端查询版本，按条件获取差异补丁或完整包",
          "发布应用公告，集中查看用户与后台操作记录",
        ],
      },
    ];
    const layers = [
      {
        number: "01",
        title: "应用接入",
        subtitle: "发起请求",
        text: "客户端携带应用标识请求认证、配置或版本信息，进入对应应用的处理流程。",
        tags: ["账号登录", "卡密验证", "版本查询"],
      },
      {
        number: "02",
        title: "认证与策略",
        subtitle: "校验访问",
        text: "服务端按应用配置校验请求签名、登录会话、授权状态与设备条件。",
        tags: ["会话管理", "授权校验", "设备策略"],
      },
      {
        number: "03",
        title: "业务服务",
        subtitle: "处理业务",
        text: "通过认证后，由对应服务处理用户资料、云函数、扩展内容或应用更新。",
        tags: ["用户管理", "云端逻辑", "内容分发"],
      },
      {
        number: "04",
        title: "数据与记录",
        subtitle: "保存结果",
        text: "账号和业务记录持久保存，会话与缓存支持请求处理，操作日志供后续查询。",
        tags: ["业务数据", "会话缓存", "操作记录"],
      },
    ];
    return { features, layers, year: new Date().getFullYear() };
  },
};
</script>

<template>
  <div class="home-page">
    <a class="skip-link" href="#home-main">跳转到主要内容</a>
    <header class="site-header shell">
      <a class="brand" href="#home-main" aria-label="Auth 首页">
        <span class="brand-symbol"><IconSafe /></span>
        <span>Auth<span class="brand-dot">.</span></span>
        <span class="brand-caption">应用认证与管理</span>
      </a>
      <nav class="navigation" aria-label="主导航">
        <Link class="nav-feature" href="#capabilities">产品能力</Link>
        <Link class="nav-docs" href="#platform">平台架构</Link>
        <span class="nav-wordmark">BUILT FOR CONNECTION.</span>
      </nav>
    </header>

    <main id="home-main">
      <section class="hero shell" aria-labelledby="hero-title">
        <div class="hero-copy">
          <div class="eyebrow"><span></span> AUTH · APPLICATION PLATFORM</div>
          <h1 id="hero-title">
            连接你的应用，<br />掌控每一次<span class="blue-word"
              >访问<svg viewBox="0 0 200 14" aria-hidden="true">
                <path d="M3 10 Q100 -3 197 8" /></svg></span
            >。
          </h1>
          <p class="hero-description">
            从用户认证到版本更新，让应用管理回归简单。<br
              class="desktop-break"
            />一个工作台，连接账号、授权与云端能力。
          </p>
          <div class="hero-actions">
            <Button type="primary" size="large" href="#capabilities"
              >探索产品能力 <IconArrowRight
            /></Button>
            <Button type="secondary" size="large" href="#platform"
              ><template #icon><IconApps /></template>了解平台架构</Button
            >
          </div>
          <div class="hero-notes">
            <span><IconCheck /> 自有服务器部署</span
            ><span><IconCheck /> 多应用集中管理</span>
          </div>
        </div>

        <div
          class="platform-visual"
          role="img"
          aria-label="认证流程示意：客户端应用连接 Auth 认证服务，由服务管理身份、授权和设备，再连接云函数与版本更新。"
        >
          <div class="visual-grid"></div>
          <div class="visual-heading">
            <span class="visual-kicker">ONE PLATFORM. YOUR CONTROL.</span
            ><span class="diagram-label">架构示意</span>
          </div>
          <div class="client-node">
            <span class="node-icon"><IconApps /></span>
            <div>
              <strong>你的应用</strong><small>CLIENT APPLICATION</small>
            </div>
            <span class="node-dot"></span>
          </div>
          <div class="connector"><span>认证请求</span></div>
          <div class="auth-node">
            <div class="auth-node-heading">
              <span class="auth-symbol"><IconSafe /></span>
              <div>
                <strong>Auth<span>.</span></strong
                ><small>统一认证 · 集中管理</small>
              </div>
              <span class="node-badge">CORE</span>
            </div>
            <div class="auth-checks">
              <span><IconCheck /> 身份认证</span
              ><span><IconCheck /> 授权校验</span
              ><span><IconCheck /> 设备管理</span>
            </div>
          </div>
          <div class="branch" aria-hidden="true"><i></i><i></i></div>
          <div class="service-nodes">
            <div>
              <IconCode /><span>云端能力<small>CLOUD FUNCTIONS</small></span>
            </div>
            <div>
              <IconStorage /><span>版本分发<small>VERSION CONTROL</small></span>
            </div>
          </div>
          <div class="visual-footer">
            <span class="visual-line"></span> 从连接，到管理。<span
              class="visual-line"
            ></span>
          </div>
        </div>
      </section>

      <section
        id="capabilities"
        class="capabilities shell"
        aria-labelledby="capabilities-title"
      >
        <div class="section-heading">
          <div>
            <span class="section-kicker">BUILT FOR YOUR APPLICATIONS</span>
            <h2 id="capabilities-title">产品能力</h2>
          </div>
          <p>把分散的管理，汇集到同一个地方。</p>
        </div>
        <div class="feature-grid">
          <Card
            :bordered="false"
            :body-style="{ padding: 0 }"
            v-for="feature in features"
            :key="feature.number"
            class="feature"
          >
            <div class="feature-top">
              <component :is="feature.icon" /><span>{{ feature.number }}</span>
            </div>
            <p class="feature-label">{{ feature.label }}</p>
            <h3>{{ feature.title }}</h3>
            <p class="feature-description">{{ feature.text }}</p>
            <ul class="feature-details">
              <li v-for="detail in feature.details" :key="detail">
                <IconCheck aria-hidden="true" /><span>{{ detail }}</span>
              </li>
            </ul>
            <div class="feature-tags">
              <Tag
                v-for="tag in feature.tags"
                :key="tag"
                size="medium"
                color="arcoblue"
                >{{ tag }}</Tag
              >
            </div>
          </Card>
        </div>
      </section>
      <section
        id="platform"
        class="architecture shell"
        aria-labelledby="platform-title"
      >
        <div class="section-heading">
          <div>
            <span class="section-kicker">HOW AUTH WORKS</span>
            <h2 id="platform-title">平台架构</h2>
          </div>
          <p>从客户端请求到业务处理，一条清晰的服务链路。</p>
        </div>
        <ol class="architecture-flow">
          <li
            v-for="layer in layers"
            :key="layer.number"
            class="architecture-step"
          >
            <Card
              :bordered="false"
              :body-style="{ padding: 0 }"
              class="architecture-card"
            >
              <div class="step-heading">
                <span>{{ layer.number }}</span
                ><span>{{ layer.subtitle }}</span
                ><IconArrowRight aria-hidden="true" />
              </div>
              <h3>{{ layer.title }}</h3>
              <p>{{ layer.text }}</p>
              <div class="feature-tags">
                <Tag
                  v-for="tag in layer.tags"
                  :key="tag"
                  size="medium"
                  color="arcoblue"
                  >{{ tag }}</Tag
                >
              </div>
            </Card>
          </li>
        </ol>
        <div class="architecture-notes">
          <p>
            <IconApps aria-hidden="true" /><span
              ><strong>以应用为管理单位</strong
              >账号、授权配置与版本内容按应用维护。</span
            >
          </p>
          <p>
            <IconSafe aria-hidden="true" /><span
              ><strong>部署在自有服务器</strong
              >由你维护业务数据、文件与运行环境。</span
            >
          </p>
        </div>
      </section>
      <section class="workspace-banner shell" aria-label="应用管理理念">
        <div>
          <span class="banner-icon"><IconSafe /></span>
          <div>
            <h2>你的应用，你的管理空间。</h2>
            <p>连接、认证、分发，让每一个环节清晰可控。</p>
          </div>
        </div>
        <span class="banner-wordmark">YOUR APPS. YOUR RULES.</span>
      </section>
    </main>
    <footer class="site-footer shell">
      <span class="footer-brand">Auth<span>.</span></span
      ><span>© {{ year }} Auth · 应用认证与管理</span
      ><span class="footer-note">简单连接，有序管理。</span>
    </footer>
  </div>
</template>

<style scoped>
.home-page {
  --blue: rgb(var(--primary-6));
  --ink: var(--color-text-1);
  --muted: var(--color-text-2);
  min-height: 100vh;
  background: var(--color-bg-1);
  color: var(--ink);
  font-family: "Inter", "PingFang SC", "Microsoft YaHei", sans-serif;
  overflow: clip;
}
.home-page a {
  color: inherit;
  text-decoration: none;
}
.home-page a:focus-visible {
  outline: 3px solid #7ea3ff;
  outline-offset: 6px;
}
.shell {
  width: min(1320px, calc(100% - 64px));
  margin-inline: auto;
}
.skip-link {
  position: absolute;
  top: -80px;
  left: 24px;
  padding: 12px;
  z-index: 10;
  background: white;
}
.skip-link:focus {
  top: 12px;
}
.site-header {
  height: 72px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e6ebf6;
}
.brand {
  display: flex;
  align-items: center;
  gap: 11px;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -1px;
}
.brand-symbol {
  width: 36px;
  height: 40px;
  display: grid;
  place-items: center;
  color: white;
  background: var(--blue);
  border-radius: 11px 11px 16px 16px;
  font-size: 23px;
}
.brand-dot,
.footer-brand span {
  color: var(--blue);
}
.brand-caption {
  font-size: 12px;
  font-weight: 400;
  color: var(--muted);
  letter-spacing: 1px;
  margin-left: 15px;
  border-left: 1px solid #dce3f0;
  padding-left: 23px;
}
.navigation {
  display: flex;
  align-items: center;
  gap: 32px;
  font-size: 13px;
  font-weight: 500;
}
.navigation a:hover {
  color: var(--blue);
}
.nav-wordmark {
  font-size: 9px;
  letter-spacing: 1.2px;
  color: #7c8aa3;
}

.hero {
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  gap: 40px;
  align-items: center;
  padding-block: 30px 32px;
}
.eyebrow {
  display: flex;
  align-items: center;
  gap: 9px;
  color: var(--blue);
  font-size: 10px;
  letter-spacing: 1.7px;
  font-weight: 700;
  margin-bottom: 18px;
}
.eyebrow > span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--blue);
  box-shadow: 0 0 0 4px #e5edff;
}
h1 {
  font-size: clamp(36px, 3.8vw, 53px);
  line-height: 1.35;
  letter-spacing: -2px;
  margin: 0 0 18px;
  font-weight: 750;
}
.blue-word {
  position: relative;
  color: var(--blue);
  white-space: nowrap;
}
.blue-word svg {
  position: absolute;
  width: 100%;
  height: 10px;
  bottom: -8px;
  left: 0;
  fill: none;
  stroke: #86aaff;
  stroke-width: 4;
  stroke-linecap: round;
}
.hero-description {
  color: var(--muted);
  font-size: 15px;
  line-height: 2;
  margin-bottom: 18px;
}
.hero-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.hero-notes {
  display: flex;
  gap: 24px;
  margin-top: 16px;
  color: #6c7a93;
  font-size: 11px;
}
.hero-notes span {
  display: flex;
  align-items: center;
  gap: 6px;
}
.hero-notes svg {
  color: var(--blue);
}
.platform-visual {
  position: relative;
  isolation: isolate;
  border: 1px solid #dce6fc;
  border-radius: 22px;
  padding: 18px 24px 16px;
  background: linear-gradient(135deg, #edf3ff, #e0eaff);
  box-shadow: 0 20px 55px -25px #4170cb55;
}
.visual-grid {
  position: absolute;
  z-index: -1;
  inset: 0;
  border-radius: inherit;
  opacity: 0.5;
  background-image: radial-gradient(#9fb7e3 1px, transparent 1px);
  background-size: 18px 18px;
  mask-image: linear-gradient(black, transparent);
}
.visual-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 18px;
}
.visual-kicker {
  font-size: 8px;
  letter-spacing: 1.5px;
  color: #516f9e;
  font-weight: 600;
}
.diagram-label {
  font-size: 10px;
  color: #5c7399;
  padding: 4px 7px;
  background: #ffffff80;
  border: 1px solid #cfdcf3;
  border-radius: 4px;
  white-space: nowrap;
}
.client-node {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 76%;
  margin: 0 auto;
  padding: 11px;
  background: #ffffffed;
  border: 1px solid #dbe5f8;
  border-radius: 10px;
}
.node-icon {
  font-size: 24px;
  display: grid;
  place-items: center;
  color: var(--blue);
  width: 36px;
  height: 36px;
  background: #eef3ff;
  border-radius: 8px;
}
.client-node strong {
  display: block;
  font-size: 12px;
  font-weight: 600;
}
.client-node small {
  display: block;
  font-size: 7px;
  letter-spacing: 1px;
  color: #7d8ba4;
  margin-top: 5px;
}
.node-dot {
  width: 6px;
  height: 6px;
  background: #6791fa;
  margin-left: auto;
  border-radius: 50%;
}
.connector {
  height: 30px;
  width: 1px;
  background: #9bb6ef;
  margin: auto;
  position: relative;
}
.connector::after {
  content: "";
  position: absolute;
  bottom: 0;
  width: 6px;
  height: 6px;
  border-right: 1px solid #7095df;
  border-bottom: 1px solid #7095df;
  transform: translateX(-3px) rotate(45deg);
}
.connector span {
  position: absolute;
  top: 12px;
  left: 12px;
  color: #7189b4;
  font-size: 9px;
  white-space: nowrap;
}
.auth-node {
  background: linear-gradient(120deg, #2865fa, #164be0);
  border: 1px solid #3b73fb;
  border-radius: 12px;
  color: white;
  padding: 16px 20px;
  box-shadow: 0 12px 25px #245cf42b;
}
.auth-node-heading {
  display: flex;
  align-items: center;
  gap: 13px;
}
.auth-symbol {
  width: 43px;
  height: 47px;
  background: #ffffff1f;
  border: 1px solid #ffffff40;
  border-radius: 10px 10px 15px 15px;
  display: grid;
  place-items: center;
  font-size: 28px;
}
.auth-node strong {
  font-size: 29px;
  letter-spacing: -1px;
}
.auth-node strong span {
  color: #a7c4ff;
}
.auth-node small {
  display: block;
  margin-top: 2px;
  font-size: 10px;
  color: #e0eaff;
}
.node-badge {
  margin-left: auto;
  font-size: 8px;
  letter-spacing: 1px;
  padding: 5px 7px;
  border: 1px solid #ffffff40;
  border-radius: 4px;
  color: #d7e5ff;
}
.auth-checks {
  display: flex;
  justify-content: space-between;
  gap: 6px;
  margin-top: 14px;
  padding-top: 11px;
  border-top: 1px solid #ffffff26;
  font-size: 10px;
}
.auth-checks span {
  display: flex;
  align-items: center;
  gap: 4px;
}
.branch {
  width: 54%;
  height: 26px;
  margin: 0 auto;
  position: relative;
}
.branch::before {
  content: "";
  height: 13px;
  width: 1px;
  position: absolute;
  top: 0;
  left: 50%;
  background: #9bb6ef;
}
.branch::after {
  content: "";
  position: absolute;
  top: 13px;
  inset-inline: 0;
  bottom: 0;
  border: 1px solid #9bb6ef;
  border-bottom: 0;
  border-radius: 5px 5px 0 0;
}
.service-nodes {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.service-nodes > div {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  border: 1px solid #d6e1f6;
  border-radius: 9px;
  background: #ffffffce;
  font-size: 11px;
}
.service-nodes svg {
  font-size: 21px;
  color: var(--blue);
  flex-shrink: 0;
}
.service-nodes small {
  display: block;
  font-size: 7px;
  letter-spacing: 0.5px;
  color: #8090ad;
  margin-top: 5px;
}
.visual-footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 13px;
  margin-top: 14px;
  font-size: 9px;
  letter-spacing: 2px;
  color: #6d83a8;
}
.visual-line {
  width: 28px;
  height: 1px;
  background: #bdcde9;
}
.capabilities {
  padding-top: 0;
  scroll-margin-top: 14px;
}
.section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  margin-bottom: 18px;
}
.section-kicker {
  font-size: 9px;
  letter-spacing: 1.7px;
  color: var(--blue);
  font-weight: 600;
}
.section-heading h2 {
  font-size: 27px;
  font-weight: 650;
  margin: 8px 0 0;
  letter-spacing: -0.6px;
}
.section-heading > p {
  font-size: 12px;
  color: var(--muted);
  margin: 0 0 4px;
}
.feature-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border: 1px solid #e0e7f3;
  background: var(--color-bg-2);
  border-radius: 12px;
  overflow: hidden;
}
.feature {
  padding: 22px 24px;
  min-width: 0;
  border-radius: 0;
}
.feature :deep(.arco-card-body) {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.feature .feature-tags {
  margin-top: auto;
}
.feature + .feature {
  border-left: 1px solid #e6ecf5;
}
.feature-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.feature-top svg {
  font-size: 24px;
  color: var(--blue);
}
.feature-top > span {
  color: #afbdd6;
  font-size: 12px;
  font-family: monospace;
}
.feature-label {
  font-size: 12px;
  color: var(--blue);
  margin: 16px 0 7px;
}
.feature h3 {
  font-size: 17px;
  font-weight: 600;
  margin: 0;
  letter-spacing: -0.5px;
}
.feature-description {
  font-size: 13px;
  line-height: 1.95;
  color: var(--muted);
  margin: 10px 0 14px;
}
.feature-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}
.feature-details {
  list-style: none;
  padding: 0;
  margin: 0 0 16px;
  display: grid;
  gap: 8px;
}
.feature-details li {
  display: flex;
  gap: 7px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--color-text-2);
}
.feature-details svg {
  flex-shrink: 0;
  margin-top: 4px;
  color: var(--blue);
}
.architecture {
  margin-top: 28px;
  scroll-margin-top: 20px;
}
.architecture-flow {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border: 1px solid #dce6fa;
  border-radius: 12px;
  overflow: hidden;
  background: #edf3ff;
}
.architecture-step {
  padding: 20px;
  min-width: 0;
}
.architecture-step + .architecture-step {
  border-left: 1px solid #dce6fa;
}
.step-heading {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #637fae;
  font-size: 10px;
}
.step-heading > span:first-child {
  display: grid;
  place-items: center;
  width: 25px;
  height: 25px;
  border-radius: 6px;
  background: var(--blue);
  color: white;
  font-family: monospace;
}
.step-heading svg {
  margin-left: auto;
  font-size: 16px;
  color: var(--blue);
}
.architecture-step:last-child .step-heading svg {
  visibility: hidden;
}
.architecture-step h3 {
  font-size: 16px;
  margin: 14px 0 8px;
}
.architecture-card p {
  color: var(--color-text-2);
  font-size: 13px;
  line-height: 1.85;
  margin: 0 0 14px;
}
.architecture-notes {
  display: flex;
  gap: 30px;
  margin-top: 12px;
}
.architecture-notes p {
  display: flex;
  gap: 9px;
  align-items: flex-start;
  color: var(--muted);
  font-size: 11px;
  line-height: 1.8;
  margin: 0;
}
.architecture-notes svg {
  color: var(--blue);
  flex-shrink: 0;
  margin-top: 4px;
  font-size: 16px;
}
.architecture-notes strong {
  font-weight: 600;
  color: #334d75;
  margin-right: 9px;
}
@media (max-width: 1100px) {
  .architecture-flow {
    grid-template-columns: repeat(2, 1fr);
  }
  .architecture-step:nth-child(3) {
    border-left: 0;
  }
  .architecture-step:nth-child(n + 3) {
    border-top: 1px solid #dce6fa;
  }
}
@media (max-width: 540px) {
  .architecture-flow {
    grid-template-columns: 1fr;
  }
  .architecture-step {
    padding: 16px;
  }
  .architecture-step + .architecture-step {
    border-left: 0;
    border-top: 1px solid #dce6fa;
  }
  .step-heading svg {
    transform: rotate(90deg);
  }
  .architecture-notes {
    flex-direction: column;
    gap: 8px;
  }
}

.workspace-banner {
  margin-top: 20px;
  padding: 17px 24px;
  background: #eaf0ff;
  border: 1px solid #dce6ff;
  border-radius: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}
.workspace-banner > div {
  display: flex;
  align-items: center;
  gap: 18px;
}
.banner-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  color: var(--blue);
  font-size: 26px;
}
.workspace-banner h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
.workspace-banner p {
  margin: 7px 0 0;
  font-size: 11px;
  color: var(--muted);
}
.banner-wordmark {
  display: flex;
  align-items: center;
  gap: 28px;
  font-size: 12px;
  font-weight: 600;
  color: var(--blue);
  white-space: nowrap;
}
.site-footer {
  display: flex;
  align-items: center;
  gap: 23px;
  padding-block: 20px;
  margin-top: 4px;
  color: #7c8aa3;
  font-size: 10px;
}
.footer-brand {
  color: var(--ink);
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -1px;
}
.footer-note {
  margin-left: auto;
  letter-spacing: 1px;
}
@media (min-width: 1600px) {
  .hero {
    padding-block: 36px;
  }
}
@media (max-width: 1100px) {
  .shell {
    width: calc(100% - 48px);
  }
  .hero {
    gap: 30px;
  }
  .platform-visual {
    padding-inline: 20px;
  }
  .brand-caption {
    display: none;
  }
  .feature {
    padding: 20px;
  }
  .feature h3 {
    font-size: 15px;
  }
}
@media (max-width: 800px) {
  .hero {
    grid-template-columns: 1fr;
    padding-block: 26px;
    gap: 24px;
  }
  .hero-copy {
    text-align: center;
  }
  .eyebrow,
  .hero-actions,
  .hero-notes {
    justify-content: center;
  }
  h1 {
    font-size: 43px;
  }
  .platform-visual {
    width: min(470px, 100%);
    margin: auto;
    padding-inline: 30px;
  }
  .section-heading {
    display: block;
  }
  .section-heading > p {
    margin-top: 8px;
  }
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .feature {
    padding: 20px;
  }
  .feature + .feature {
    border-left: 0;
    border-top: 1px solid #e6ecf5;
  }
  .feature h3 {
    font-size: 18px;
  }
  .feature-label {
    margin-top: 16px;
  }
  .feature-description {
    max-width: 540px;
  }
  .workspace-banner {
    padding: 16px;
  }
}
@media (max-width: 540px) {
  .shell {
    width: calc(100% - 36px);
  }
  .site-header {
    height: 64px;
  }
  .brand {
    font-size: 25px;
    gap: 8px;
  }
  .brand-symbol {
    width: 30px;
    height: 34px;
    font-size: 20px;
  }
  .nav-wordmark {
    display: none;
  }
  .navigation {
    gap: 16px;
    font-size: 12px;
  }
  .nav-wordmark {
    font-size: 9px;
    letter-spacing: 1.2px;
    color: #7c8aa3;
  }
  .eyebrow {
    font-size: 8px;
    letter-spacing: 1.2px;
    margin-bottom: 16px;
  }
  h1 {
    font-size: clamp(29px, 8.5vw, 42px);
    letter-spacing: -1.4px;
  }
  .hero-description {
    font-size: 14px;
  }
  .desktop-break {
    display: none;
  }

  .hero-notes {
    gap: 17px;
    font-size: 10px;
  }
  .platform-visual {
    padding: 16px;
  }
  .visual-kicker {
    font-size: 7px;
    letter-spacing: 1px;
  }
  .client-node {
    width: 85%;
  }
  .auth-node {
    padding: 16px;
  }
  .service-nodes > div {
    padding: 10px 9px;
    gap: 7px;
  }
  .service-nodes small {
    font-size: 6px;
  }
  .section-heading h2 {
    font-size: 24px;
  }
  .workspace-banner {
    align-items: flex-start;
    flex-direction: column;
    gap: 14px;
  }
  .workspace-banner > div {
    gap: 10px;
  }
  .workspace-banner h2 {
    font-size: 15px;
  }
  .banner-wordmark {
    align-self: flex-end;
  }
  .site-footer {
    gap: 14px;
    flex-wrap: wrap;
    padding-block: 18px;
  }
  .footer-note {
    display: none;
  }
}

.hero-actions :deep(.arco-btn) {
  gap: 10px;
  height: 46px;
  padding: 0 22px;
  font-size: 13px;
}
.hero-actions :deep(.arco-btn-primary) {
  color: #fff;
}
.navigation :deep(.arco-link) {
  color: var(--color-text-2);
}
.architecture-card {
  background: transparent;
  height: 100%;
}
.architecture-card :deep(.arco-card-body) {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.architecture-card .feature-tags {
  margin-top: auto;
}
.architecture-card p {
  color: var(--muted);
}
:global(body[arco-theme="dark"] .home-page .platform-visual) {
  background: linear-gradient(135deg, #1c2944, #192237);
  border-color: var(--color-border-2);
}
:global(body[arco-theme="dark"] .home-page .client-node),
:global(body[arco-theme="dark"] .home-page .service-nodes > div),
:global(body[arco-theme="dark"] .home-page .diagram-label) {
  background: var(--color-bg-3);
  border-color: var(--color-border-2);
  color: var(--color-text-2);
}
:global(body[arco-theme="dark"] .home-page .architecture-flow),
:global(body[arco-theme="dark"] .home-page .workspace-banner) {
  background: var(--color-fill-1);
  border-color: var(--color-border-2);
}
:global(body[arco-theme="dark"] .home-page .feature-grid),
:global(body[arco-theme="dark"] .home-page .feature),
:global(body[arco-theme="dark"] .home-page .architecture-step),
:global(body[arco-theme="dark"] .home-page .site-header),
:global(body[arco-theme="dark"] .home-page .site-footer) {
  border-color: var(--color-border-2);
}
:global(body[arco-theme="dark"] .home-page .architecture-notes strong) {
  color: var(--color-text-2);
}
</style>
