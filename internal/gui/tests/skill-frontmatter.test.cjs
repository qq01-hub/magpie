// A skill with malformed YAML stays visible by its folder name, warns in the
// Library, in skills found in agents and in the source picker, and never trusts
// partly parsed fields. Both engines and every Library language, with an
// HTML-like parser error kept out of the DOM.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");
const { strings } = require("./strings.cjs");

const assets = path.resolve(__dirname, "../assets");
const HOME = "/Users/me";
const BAD = "<img src=x onerror=alert(1)>";
const agent = { id: "claude", name: "Claude Code", icon: "", skills: `${HOME}/.claude/skills` };
const base = () => ({
  dir: `${HOME}/.magpie/library`, backups: `${HOME}/.magpie/backups`, home: HOME,
  agents: [agent], instructions: { agents: [], sets: [] }, servers: [], skills: [
    { name: "broken", description: "", frontmatterError: BAD, agents: ["claude"] },
    { name: "valid", description: "Valid description", manual: true, agents: ["claude"] },
  ], foundServers: [], foundSkills: [
    { name: "own-broken", description: "", frontmatterError: BAD, agents: ["claude"] },
  ], newSkills: [], projects: [],
});
const probe = { source: "acme/skills", kind: "github", candidates: [
  { path: "skills/listed-a", name: "listed-a", description: "Listed A", have: false },
  { path: "skills/listed-b", name: "listed-b", description: "Listed B", have: false },
  { path: "skills/space skill", name: "space-skill", displayName: "space skill", description: "", frontmatterError: BAD, have: false },
] };
const marketItems = [
  { id: "acme/skills/listed-a", name: "listed-a", source: "acme/skills", skillId: "listed-a", description: "Listed A", installs: 200, icon: "" },
  { id: "acme/skills/listed-b", name: "listed-b", source: "acme/skills", skillId: "listed-b", description: "Listed B", installs: 100, icon: "" },
];

let I18N;
function tr(lang, key) {
  if (lang === "en") return key;
  assert.equal(typeof I18N[lang]?.[key], "string", `${lang} has no ${JSON.stringify(key)}`);
  return I18N[lang][key];
}

function server(lang, posts, probeGate) {
  const installed = [];
  const current = () => ({
    ...base(), skills: [...base().skills, ...installed.map((name) => ({ name, description: "", agents: ["claude"] }))],
  });
  return async (route) => {
    const req = route.request(), url = new URL(req.url());
    if (url.pathname === "/boot.js") return route.fulfill({ contentType: "text/javascript", body: `window.bootPrefs = {lang:${JSON.stringify(lang)},theme:"light",web:true};` });
    if (url.pathname === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
    if (url.pathname === "/api/state") return route.fulfill({ json: { agents: [], profiles: [], settings: { lang, theme: "light" } } });
    if (url.pathname === "/api/library") return route.fulfill({ json: current() });
    if (url.pathname === "/api/library/market/skills") return route.fulfill({ json: { items: structuredClone(marketItems) } });
    if (url.pathname === "/api/library/skills/probe") {
      await probeGate;
      return route.fulfill({ json: probe });
    }
    if (url.pathname === "/api/library/skills/install") {
      const body = req.postDataJSON();
      posts.push(body);
      const names = body.paths.map((p) => p.split("/").pop().replaceAll(" ", "-"));
      installed.push(...names);
      return route.fulfill({ json: { ...current(), result: { changed: ["claude"], problems: [], installed: names } } });
    }
    if (url.pathname === "/api/plugins") return route.fulfill({ json: { plugins: [] } });
    if (url.pathname === "/api/usage/quotas") return route.fulfill({ json: [] });
    if (url.pathname === "/api/groups") return route.fulfill({ json: { groups: [] } });
    if (url.pathname === "/api/providers") return route.fulfill({ json: { providers: [], gateway: { running: true } } });
    if (url.pathname.startsWith("/api/")) return route.fulfill({ json: {} });
    const file = path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
    const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    try { await route.fulfill({ body: await fs.readFile(file), contentType }); }
    catch { await route.fulfill({ status: 404, body: "" }); }
  };
}

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"])) {
  test(engine + ": malformed skill front matter stays visible and is explained", async (t) => {
    assert(["chromium", "webkit"].includes(engine), "BROWSER must be chromium or webkit");
    I18N = strings().I18N;
    const browser = await (engine === "webkit" ? webkit.launch() : chromium.launch({ channel: "chromium" }));
    const errors = [];
    t.after(async () => {
      if (errors.length) console.log(errors);
      await browser.close();
    });

    const languages = process.env.FRONTMATTER_LANG ? [process.env.FRONTMATTER_LANG] : ["en", "zh", "zh-TW", "ja", "de"];
    for (const lang of languages) {
      await t.test(lang, async () => {
        const posts = [];
        let releaseProbe;
        const probeGate = new Promise((resolve) => { releaseProbe = resolve; });
        const width = lang === "zh" || lang === "zh-TW" ? 440 : 980;
        const ctx = await browser.newContext({ viewport: { width, height: 800 } });
        await ctx.addInitScript(() => { try { localStorage.setItem("magpie.libTab", "skills"); } catch {} });
        const page = await ctx.newPage();
        page.setDefaultTimeout(5000);
        page.setDefaultNavigationTimeout(60000);
        page.on("pageerror", (e) => errors.push(e.message));
        await page.route("**/*", server(lang, posts, probeGate));
        await page.goto("http://magpie.test/");
        await page.locator('button[data-view="library"]').click();

        const warn = tr(lang, "Front matter unreadable");
        const why = tr(lang, "SKILL.md front matter could not be read; its folder name is shown and its metadata is ignored");
        const broken = page.locator("#view-library .lib-skill").filter({ hasText: "broken" });
        await broken.waitFor();
        assert.equal((await broken.locator(".lib-tag.warn").textContent()).trim(), warn);
        assert.equal(await broken.locator(".lib-tag.warn").getAttribute("title"), why);
        assert.equal(await broken.locator(".lib-manual").count(), 0, "partially parsed manual metadata is ignored");
        const valid = page.locator("#view-library .lib-skill").filter({ hasText: "valid" });
        assert.equal((await valid.locator(".sub").textContent()).trim(), "Valid description");
        assert.equal(await valid.locator(".lib-tag.warn").count(), 0, "valid neighboring metadata stays intact");

        const own = page.locator("#view-library .lib-found").filter({ hasText: "own-broken" });
        assert.equal((await own.locator(".lib-tag.warn").textContent()).trim(), warn);
        assert.equal(await own.locator(".lib-manual").count(), 0);
        assert.equal(await page.locator('#view-library img[src="x"]').count(), 0, "error contents are not interpreted as markup");
        assert.equal((await page.locator("#view-library").innerText()).includes(BAD), false, "raw parser errors are not copied into page text");

        // Before the repository fetch completes, its market-listed skills
        // still have names; the later response replaces them with candidates.
        const repo = page.locator('#view-library .mk-card[data-id="repo:acme/skills"]');
        await repo.click();
        const sheet = page.locator("#modal .mk-reposheet");
        const listed = sheet.locator(".lib-pick");
        await listed.first().waitFor();
        assert.deepEqual(await listed.locator(".name").allTextContents(), ["listed-a", "listed-b"]);
        assert.equal(await listed.locator(".lib-tag.warn").count(), 0);
        releaseProbe();
        await page.waitForFunction(() => document.querySelectorAll("#modal .mk-reposheet .lib-pick").length === 3);
        assert.deepEqual(await listed.locator(".name").evaluateAll((els) => els.map((el) => el.firstChild.textContent.trim())), ["listed-a", "listed-b", "space skill"]);
        await sheet.locator(".bar .primary").click();
        await sheet.waitFor({ state: "detached" });

        const source = page.locator('#view-library input[data-lib="source"]');
        await source.fill("acme/skills");
        await source.press("Enter");
        const candidate = page.locator("#view-library .lib-pick").filter({ hasText: "space skill" });
        await candidate.waitFor();
        assert.equal((await candidate.locator(".name").evaluate((el) => el.firstChild.textContent)).trim(), "space skill");
        assert.equal((await candidate.locator(".lib-tag.warn").textContent()).trim(), warn);
        assert.equal(await candidate.locator(".lib-tag.warn").getAttribute("title"), why);
        assert.equal(await candidate.locator("input").isChecked(), true);
        await page.locator("#view-library .lib-pick").filter({ hasText: "Listed A" }).locator("input").uncheck();
        await page.locator("#view-library .lib-pick").filter({ hasText: "Listed B" }).locator("input").uncheck();
        await page.locator("#view-library .lib-probefoot button.primary").click();
        await page.waitForFunction(() => [...document.querySelectorAll("#view-library .lib-skill .name")].some((el) => el.firstChild.textContent.trim() === "space-skill"));
        assert.deepEqual(posts.map((p) => p.paths), [["skills/listed-a", "skills/listed-b"], ["skills/space skill"]], "the picker submits candidate paths, not display labels");
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth), true, "the warning stays within the viewport");
        await ctx.close();
      });
    }

    assert.deepEqual(errors, []);
  });
}
