// Run with Node's test runner and Playwright on the module path; see README.md.
// Riablo's four groups (#7221): every model in Review already belongs to
// an earlier sibling. These fixtures are serialized Handler traces from
// TestTraceKeepsACompletelySharedRuleGroup and
// TestTraceSharedGroupKeepsItsNestedRulePath, not a separately made UI shape.
// The real Routing page must keep Review, its three shared models only
// once, and the Review > Inner ancestors of a nested rule's chosen model.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");
const traces = require("./fixtures/routing-shared-groups.json");

const assets = path.resolve(__dirname, "../assets");
const sharedModels = [
  "claude-relay/claude-sonnet-5-5", "codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash:max",
  "deepseek/deepseek-flash", "claude-relay/claude-haiku-5-5", "codex-relay/gpt-6.1-terra",
  "openrouter/deepseek/deepseek-v4.1-flash-1017", "claude-relay/claude-opus-5-5", "codex-relay/gpt-6.2-astra",
];
const reviewModels = ["codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash:max", "claude-relay/claude-opus-5-5"];

async function open(t, engine, lang, width, fixture) {
  const browser = await (engine === "webkit" ? webkit.launch() : chromium.launch({ channel: "chromium" }));
  t.after(() => browser.close());
  const page = await browser.newPage({ viewport: { width, height: 1100 }, reducedMotion: "reduce" });
  page.setDefaultTimeout(6000);
  page.setDefaultNavigationTimeout(60000);
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  const route = structuredClone(traces[fixture]);
  // Only the clock is freshened so the recorded completed request remains
  // on the live stage during this browser check; all routing data is kept.
  route.time = new Date().toISOString();
  for (const attempt of route.tries) attempt.start = route.time;
  await page.route("**/*", async (request) => {
    const url = new URL(request.request().url());
    assert.equal(url.origin, "http://magpie.test", "no request may reach a real origin");
    const json = (data) => request.fulfill({ json: data });
    if (url.pathname === "/boot.js") return request.fulfill({ contentType: "text/javascript", body: `window.bootPrefs = {lang:"${lang}",theme:"light",web:true};` });
    if (url.pathname === "/wails/runtime.js") return request.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
    if (url.pathname === "/api/state") return json({ agents: [], profiles: [], settings: { lang, theme: "light" } });
    if (url.pathname === "/api/gateway/trace") {
      if (url.searchParams.get("wait")) return;
      return json({ mine: true, now: route.time, seq: route.seq, totals: { requests: 1, rerouted: 0, errors: 0 }, routes: [route] });
    }
    if (url.pathname === "/api/gateway/history") return json({ cut: false, days: [], routes: [] });
    if (url.pathname === "/api/groups") return json({ groups: [], pools: [] });
    if (url.pathname === "/api/providers") return json({ providers: [], excluded: [], gateway: { running: true, window: true } });
    if (url.pathname.startsWith("/api/")) return json({});
    const file = path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
    const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    await request.fulfill({ body: await fs.readFile(file), contentType });
  });
  await page.goto("http://magpie.test/?view=routing");
  await page.locator(".rt-accts > li.rt-sub").first().waitFor();
  const group = (name) => page.locator(".rt-accts > li.rt-sub", { has: page.locator("b", { hasText: new RegExp(`^${name}$`) }) });
  const models = () => page.locator(".rt-accts > li:not(.rt-sub):not(.rt-fold)").evaluateAll((rows) => rows.map((li) =>
    li.title.split("\n")[0] + "/" + li.querySelector("code.mdl").textContent));
  const screenshot = async () => {
    if (!process.env.ARTIFACT_DIR || lang !== "zh") return;
    await fs.mkdir(process.env.ARTIFACT_DIR, { recursive: true });
    await page.screenshot({ path: path.join(process.env.ARTIFACT_DIR, `7221-${fixture}-${engine}-${width}.png`), fullPage: true });
  };
  return { page, errors, group, models, screenshot };
}

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"])) {
  for (const lang of ["en", "zh", "zh-TW", "ja", "de"]) {
    for (const width of [1100, 420]) {
      test(`${engine} ${lang} ${width}px: the report's four groups retain completely shared Review without duplicate models`, async (t) => {
        const { page, errors, group, models, screenshot } = await open(t, engine, lang, width, "shared");
        assert.equal(await page.locator(".rt-what b").textContent(), "Task");
        assert.deepEqual((await page.locator(".rt-accts > li.rt-sub > b").allTextContents()).sort(), ["Review", "Task high", "Task low", "Task mid"]);
        assert.deepEqual((await models()).sort(), [...sharedModels].sort());
        assert.equal(new Set(await models()).size, 9);
        // A shared member later in the original flattened member order
        // still sits under its chosen heading, not the next sibling's.
        assert.deepEqual(await page.locator(".rt-accts > li").evaluateAll((rows) => rows.map((li) =>
          li.classList.contains("rt-sub") ? "group:" + li.querySelector("b").textContent : li.title.split("\n")[0] + "/" + li.querySelector("code.mdl").textContent)), [
          "group:Task mid", "claude-relay/claude-sonnet-5-5",
          "group:Review", "codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash:max", "claude-relay/claude-opus-5-5",
          "group:Task low", "deepseek/deepseek-flash", "claude-relay/claude-haiku-5-5", "codex-relay/gpt-6.1-terra", "openrouter/deepseek/deepseek-v4.1-flash-1017",
          "group:Task high", "codex-relay/gpt-6.2-astra",
        ]);
        assert.equal(await group("Review").getAttribute("aria-expanded"), "true");
        await screenshot();
        // Folding Review hides exactly its three selected shared members,
        // leaving the six sibling models, and reopening restores nine.
        await group("Review").click();
        await page.waitForFunction(() => document.querySelectorAll(".rt-accts > li:not(.rt-sub):not(.rt-fold)").length === 6);
        assert.deepEqual((await models()).sort(), sharedModels.filter((model) => !reviewModels.includes(model)).sort());
        await group("Review").click();
        await page.waitForFunction(() => document.querySelectorAll(".rt-accts > li:not(.rt-sub):not(.rt-fold)").length === 9);
        assert.deepEqual((await models()).sort(), [...sharedModels].sort());
        assert.deepEqual(errors, []);
      });

      test(`${engine} ${lang} ${width}px: a nested shared rule keeps Outer, Review and Inner over each unique model`, async (t) => {
        const { page, errors, group, models, screenshot } = await open(t, engine, lang, width, "nested");
        assert.equal(await page.locator(".rt-what b").textContent(), "Outer");
        assert.deepEqual(await page.locator(".rt-accts > li").evaluateAll((rows) => rows.map((li) => [
          li.classList.contains("rt-sub") ? li.querySelector("b").textContent : li.querySelector("code.mdl").textContent,
          li.style.getPropertyValue("--depth"),
        ])), [["Review", "0"], ["Inner", "1"], ["alpha", "2"], ["beta", "2"], ["gamma", "0"]]);
        assert.deepEqual(await models(), ["a/alpha", "b/beta", "c/gamma"]);
        await screenshot();
        await group("Inner").click();
        await page.waitForFunction(() => document.querySelectorAll(".rt-accts > li:not(.rt-sub):not(.rt-fold)").length === 1);
        assert.deepEqual(await models(), ["c/gamma"]);
        assert.equal(await group("Review").count(), 1);
        assert.equal(await group("Inner").count(), 1);
        await group("Review").click();
        await page.waitForFunction(() => document.querySelectorAll(".rt-accts > li.rt-sub").length === 1);
        assert.deepEqual(await models(), ["c/gamma"]);
        assert.deepEqual(errors, []);
      });
    }
  }
}
