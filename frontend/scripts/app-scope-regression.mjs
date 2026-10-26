import { chromium } from "playwright";
import { createServer } from "vite";
import assert from "node:assert/strict";

const server = await createServer({
  server: { host: "127.0.0.1", port: 18194, strictPort: true },
});
await server.listen();
let browser, heldRequest, holdPath;
const checks = [],
  errors = [],
  writes = [];
const apps = [
  { id: 4101, name: "Application A", type: "user", mode: "free", status: 1 },
  { id: 4102, name: "Application B", type: "user", mode: "free", status: 1 },
];
const versions = new Map(
  apps.map((app, i) => [
    app.id,
    [
      {
        id: i ? 902 : 501,
        appid: app.id,
        appRecordId: 1,
        versionName: `Version ${i ? "B" : "A"}`,
        versionKey: "stable",
        versionNo: "1.0.0",
        status: 1,
      },
    ],
  ]),
);
async function until(ready) {
  const deadline = Date.now() + 5000;
  while (!ready()) {
    assert.ok(Date.now() < deadline, "expected request did not arrive");
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
}
async function selectApp(page, id) {
  await page.evaluate(async (id) => {
    const { useAppStore } = await import("/src/stores/app.js");
    useAppStore().setCurrent(id);
  }, id);
}
async function navigate(page, path) {
  await page.evaluate(async (path) => {
    const { default: router } = await import("/src/router/index.js");
    await router.push(`/admin/${path}`);
  }, path);
}
async function releaseAndRender(page) {
  heldRequest.release();
  await heldRequest.done;
  heldRequest = undefined;
  await page.evaluate(
    () =>
      new Promise((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(resolve)),
      ),
  );
}
try {
  browser = await chromium.launch({
    executablePath:
      process.env.CHROME_PATH ||
      "C:/Program Files/Google/Chrome/Application/chrome.exe",
    headless: true,
  });
  const context = await browser.newContext();
  await context.addInitScript(() => {
    localStorage.setItem("auth_token", "test-token");
    if (!localStorage.getItem("auth_current_appid"))
      localStorage.setItem("auth_current_appid", "4101");
  });
  await context.route("**/api/**", async (route) => {
    const req = route.request(),
      url = new URL(req.url());
    if (!url.pathname.startsWith("/api/")) return route.continue();
    const scope = Number(req.headers()["x-app-id"]);
    const path = url.pathname;
    let finished;
    if (holdPath === path && scope === 4101) {
      holdPath = undefined;
      const done = new Promise((resolve) => {
        finished = resolve;
      });
      await new Promise((release) => {
        heldRequest = { release, done };
      });
    }
    let data = {};
    if (path === "/api/install/status")
      data = { installed: true, adminEntry: "admin" };
    else if (path === "/api/admin/me")
      data = { id: 1, username: "tester", role: "super" };
    else if (path === "/api/admin/app/list") data = apps;
    else if (path === "/api/admin/probe") data = { scope };
    else if (path === "/api/admin/ver/list") {
      assert.equal(
        Number(url.searchParams.get("appid")),
        scope,
        "view and request application disagree",
      );
      data = versions.get(scope) || [];
    } else if (path === "/api/admin/ver" && req.method() === "POST") {
      const body = req.postDataJSON();
      writes.push({ scope, method: "POST", id: body.id });
      data = { id: body.id };
    } else if (
      /\/api\/admin\/ver\/\d+$/.test(path) &&
      req.method() === "DELETE"
    ) {
      const id = Number(path.split("/").pop());
      writes.push({ scope, method: "DELETE", id });
      versions.set(
        scope,
        versions.get(scope).filter((row) => row.id !== id),
      );
    } else if (path === "/api/admin/dashboard") {
      data = {
        users: 1,
        usersToday: 0,
        kamiAvail: 0,
        kamiUsedToday: 0,
        recent: [
          {
            text: `Activity ${scope === 4101 ? "A" : "B"}`,
            time: "12:00",
            type: "success",
          },
        ],
        trend: { dates: ["10-27"], logins: [0], regs: [0] },
      };
    }
    try {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({ code: 0, data }),
      });
    } finally {
      finished?.();
    }
  });
  const first = await context.newPage(),
    second = await context.newPage();
  for (const page of [first, second]) {
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto("http://127.0.0.1:18194/admin/version");
    await page.getByRole("cell", { name: "Version A", exact: true }).waitFor();
  }
  await selectApp(second, 4102);
  await second.getByRole("cell", { name: "Version B", exact: true }).waitFor();
  assert.equal(
    await first.evaluate(() => localStorage.getItem("auth_current_appid")),
    "4102",
  );
  const scopes = await first.evaluate(async () => {
    const { request, rawRequest } = await import("/src/utils/request.js");
    return [
      (await request("/admin/probe")).scope,
      (await (await rawRequest("/admin/probe")).json()).data.scope,
    ];
  });
  assert.deepEqual(scopes, [4101, 4101]);
  assert.ok(
    await first
      .getByRole("cell", { name: "Version A", exact: true })
      .isVisible(),
  );
  checks.push(
    "two tabs retain separate application scope for normal requests and downloads",
  );

  const captured = await first.evaluate(async () => {
    const { request } = await import("/src/utils/request.js");
    const { useAppStore } = await import("/src/stores/app.js");
    const pending = request("/admin/probe");
    useAppStore().setCurrent(4102);
    return (await pending).scope;
  });
  assert.equal(captured, 4101);
  checks.push("request scope is fixed before an immediate application switch");

  holdPath = "/api/admin/ver/list";
  await selectApp(second, 4101);
  await until(() => heldRequest);
  await selectApp(second, 4102);
  await second.getByRole("cell", { name: "Version B", exact: true }).waitFor();
  await releaseAndRender(second);
  assert.equal(
    await second.getByRole("cell", { name: "Version A", exact: true }).count(),
    0,
  );
  checks.push(
    "late version response cannot replace the newly selected application's rows",
  );

  await navigate(second, "dashboard");
  await second.getByText("Activity B", { exact: true }).waitFor();
  holdPath = "/api/admin/dashboard";
  await selectApp(second, 4101);
  await until(() => heldRequest);
  await selectApp(second, 4102);
  await second.getByText("Activity B", { exact: true }).waitFor();
  await releaseAndRender(second);
  assert.equal(
    await second.getByText("Activity A", { exact: true }).count(),
    0,
  );
  checks.push(
    "late dashboard response cannot expose another application's recent activity",
  );

  await navigate(second, "version");
  const row = second.getByRole("row").filter({ hasText: "Version B" });
  await row.waitFor();
  assert.equal((await row.getByRole("cell").first().innerText()).trim(), "1");
  assert.ok(await second.getByText("应用内编号", { exact: true }).isVisible());
  await row.getByText("编辑", { exact: true }).click();
  await second
    .locator(".arco-modal:visible")
    .getByRole("button", { name: "提交", exact: true })
    .click();
  await until(() => writes.some((write) => write.method === "POST"));
  await second.locator(".arco-modal:visible").waitFor({ state: "hidden" });
  await row.getByText("删除", { exact: true }).click();
  await second
    .locator(".arco-popconfirm:visible")
    .getByRole("button", { name: "确定", exact: true })
    .click();
  await until(() => writes.some((write) => write.method === "DELETE"));
  assert.deepEqual(writes, [
    { scope: 4102, method: "POST", id: 902 },
    { scope: 4102, method: "DELETE", id: 902 },
  ]);
  checks.push(
    "tables show local number 1 while edits and deletes retain the original system ID",
  );
  assert.deepEqual(errors, []);
  console.log(JSON.stringify({ checks, errors }, null, 2));
} finally {
  heldRequest?.release();
  await browser?.close();
  await server.close();
}
