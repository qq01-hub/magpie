import test from "node:test";
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { mkdtemp, rm, symlink } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

test("credits: a linked command reports its usage instead of silently doing nothing", async () => {
  const dir = await mkdtemp(join(tmpdir(), "site entry # "));
  try {
    const site = fileURLToPath(new URL("./", import.meta.url));
    const linked = join(dir, "linked site");
    await symlink(site, linked, process.platform === "win32" ? "junction" : "dir");
    const result = spawnSync(process.execPath, [join(linked, "credits-cli.mjs")], { encoding: "utf8", timeout: 10_000 });
    assert.ifError(result.error);
    assert.equal(result.status, 1);
    assert.match(result.stderr, /usage: credits-cli\.mjs/);
  } finally {
    await rm(dir, { recursive: true, force: true });
  }
});

test("site commands: importing with no script argument has no command side effects", () => {
  for (const file of ["search-index.mjs", "llms-full.mjs", "docs-index.mjs", "credits-cli.mjs"]) {
    const url = pathToFileURL(fileURLToPath(new URL(file, import.meta.url))).href;
    const out = execFileSync(process.execPath, ["--input-type=module", "-e", `await import(${JSON.stringify(url)}); console.log('imported')`], { encoding: "utf8", timeout: 10_000 });
    assert.equal(out.trim(), "imported");
  }
});
