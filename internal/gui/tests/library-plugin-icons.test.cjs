// #6377: generated plugin MCP rows load their checked cached picture, while
// an explicit Library picture wins and an ordinary server keeps its glyph
// Actual DOM and image requests, including an MCP-only mutation response,
// in both engines and all five languages, with every request intercepted
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");
const { reach } = require("./reader.cjs");

const assets = path.resolve(__dirname, "../assets");
const HOME = "/Users/fixture";
const cached = "8f9aae64d410eb9a.png";
const override = "https://example.org/assigned.png";
const PNG = Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGNgYGAAAAAEAAH2FzhVAAAAAElFTkSuQmCC", "base64");
const agents = [{ id: "claude", name: "Claude Code", icon: "", mcp: `${HOME}/.claude.json` }];
const servers = (st) => [
  { name: "inherited", transport: "http", url: "http://127.0.0.1:3425/mcp/magpie/tools/search", plugin: "@example/search", icon: "file:" + cached, agents: st.on },
  { name: "assigned", transport: "http", url: "http://127.0.0.1:3425/mcp/magpie/tools/assigned", plugin: "@example/search", icon: override, agents: [] },
  { name: "ordinary", transport: "http", url: "https://example.org/mcp", agents: [] },
  { name: "unsafe", transport: "http", url: "https://example.org/unsafe", icon: "file:../providers.json", agents: [] },
];
const mcp = (st) => ({ agents, servers: servers(st), foundServers: [], projects: [], plugins: [], mcpHub: false });

function serve(lang, st) {
  return async (route) => {
    const req = route.request(), url = new URL(req.url());
    const json = (data) => route.fulfill({ json: data });
    if (url.host !== "magpie.test") { st.external.push(req.url()); return route.abort(); }
    if (url.pathname === "/boot.js") return route.fulfill({ contentType: "text/javascript", body: `window.bootPrefs={lang:${JSON.stringify(lang)},theme:"light",web:true};` });
    if (url.pathname === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window={};" });
    if (url.pathname === "/api/state") return json({ agents: [], profiles: [], settings: { lang, theme: "light" } });
    if (url.pathname === "/api/plugins") return json({ plugins: [] });
    if (url.pathname === "/api/library") return json({
      ...mcp(st), dir: `${HOME}/.magpie/library`, backups: `${HOME}/.magpie/backups`, home: HOME,
      instructions: { agents: [], sets: [] }, skills: [], foundSkills: [], problems: [],
    });
    if (url.pathname === "/api/library/servers/agents") {
      const body = req.postDataJSON();
      st.posts.push(body);
      st.on = body.agents;
      return json({ part: "mcp", mcp: mcp(st), problems: [], result: { changed: [] } });
    }
    if (url.pathname.startsWith("/api/icons/") || url.pathname === "/api/library/icon") {
      st.images.push(url.pathname + url.search);
      if (url.pathname === "/api/icons/" + cached || url.searchParams.get("u") === override) {
        return route.fulfill({ body: PNG, contentType: "image/png" });
      }
      return route.fulfill({ status: 404, body: "" });
    }
    if (url.pathname === "/api/library/mcp/check") return json({ servers: {} });
    if (url.pathname === "/api/groups") return json({ groups: [] });
    if (url.pathname === "/api/providers") return json({ providers: [], gateway: { running: true } });
    if (url.pathname === "/api/usage/quotas") return json([]);
    if (url.pathname.startsWith("/api/")) return json({});
    const file = url.pathname === "/plugins/library/client.js" && process.env.MAGPIE_LIBRARY_ICON_CLIENT
      ? process.env.MAGPIE_LIBRARY_ICON_CLIENT
      : path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
    const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    const body = await fs.readFile(file).catch(() => null);
    return body ? route.fulfill({ body, contentType }) : route.fulfill({ status: 404, body: "" });
  };
}

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"])) {
  test(`${engine}: Library plugin icons use the checked local cache`, async (t) => {
    assert(["chromium", "webkit"].includes(engine));
    const browser = await (engine === "webkit" ? webkit.launch({ headless: true }) : chromium.launch({ headless: true, channel: process.env.PLAYWRIGHT_CHANNEL || "chromium" }));
    t.after(() => browser.close());
    for (const lang of ["en", "zh", "zh-TW", "ja", "de"]) await t.test(lang, async (subtest) => {
      const st = { on: [], posts: [], images: [], external: [] }, errors = [];
      const context = await browser.newContext({ viewport: { width: 980, height: 900 }, reducedMotion: "reduce" });
      subtest.after(() => context.close());
      await context.addInitScript(() => localStorage.setItem("magpie.libTab", "mcp"));
      const page = await context.newPage();
      page.setDefaultTimeout(5000);
      page.setDefaultNavigationTimeout(60000);
      page.on("pageerror", (e) => errors.push(e.message));
      await page.route("**/*", serve(lang, st));
      await page.goto("http://magpie.test/?view=library");
      const row = (name) => page.locator(`.lib-server[data-server="${name}"]`);
      const loaded = async (name, src) => {
        const item = row(name);
        await reach(page, item);
        await page.waitForFunction((name) => {
          const r = document.querySelector(`.lib-server[data-server="${name}"]`);
          const i = r?.querySelector(".lib-logo img");
          return r && ((!i && r.querySelector(":scope > .lib-glyph")) || i?.complete);
        }, name);
        assert.equal(await item.locator(".lib-logo img").count(), 1, `${name}: the row lost its picture`);
        const image = item.locator(".lib-logo img");
        assert.equal(await image.getAttribute("src"), src);
        assert.equal(await image.evaluate((i) => i.naturalWidth), 1, `${name}: the picture did not load`);
      };
      await loaded("inherited", "/api/icons/" + cached);
      await loaded("assigned", "/api/library/icon?u=" + encodeURIComponent(override));
      assert.equal(await row("ordinary").locator("img").count(), 0);
      await reach(page, row("unsafe"));
      await page.waitForFunction(() => !document.querySelector('.lib-server[data-server="unsafe"] img'));

      const chip = row("inherited").locator('[data-agent="claude"]');
      await reach(page, chip);
      const answered = page.waitForResponse((r) => r.url().endsWith("/servers/agents"));
      await chip.click();
      await answered;
      await page.waitForFunction(() => document.querySelector('.lib-server[data-server="inherited"] [data-agent="claude"]')?.getAttribute("aria-pressed") === "true");
      await loaded("inherited", "/api/icons/" + cached);
      await loaded("assigned", "/api/library/icon?u=" + encodeURIComponent(override));
      assert.deepEqual(st.posts, [{ name: "inherited", agents: ["claude"] }]);
      assert(st.images.includes("/api/icons/" + cached), st.images.join(" | "));
      assert(!st.images.some((u) => u.startsWith("/api/library/icon") && new URL(u, "http://magpie.test").searchParams.get("u")?.startsWith("file:")), st.images.join(" | "));
      assert.deepEqual(st.external, []);
      assert.deepEqual(errors, []);
      if (process.env.ARTIFACT_DIR && lang === "en") {
        await fs.mkdir(process.env.ARTIFACT_DIR, { recursive: true });
        await reach(page, row("inherited"));
        await page.screenshot({ path: path.join(process.env.ARTIFACT_DIR, `6377-library-icons-${engine}-en.png`) });
      }
    });
  });
}
