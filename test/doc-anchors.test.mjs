// Section anchors in cross-links (AGENTS.md, "Cross-linking is mandatory"): a
// link to another doc's heading must name a heading that exists. A heading
// that is reworded takes its anchor with it, and nothing else notices.
import assert from "node:assert/strict";
import { after, test } from "node:test";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { basename, dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const temporary = [];

after(() => {
  for (const directory of temporary) rmSync(directory, { recursive: true, force: true });
});

function fixture(...files) {
  const root = mkdtempSync(join(tmpdir(), "agentic-template-anchors-"));
  temporary.push(root);
  for (const [name, text] of files) {
    mkdirSync(dirname(join(root, name)), { recursive: true });
    writeFileSync(join(root, name), text, "utf8");
  }
  return root;
}

const FENCE = /^ {0,3}(`{3,}|~{3,})(.*)$/;

// The lines outside fenced code blocks (``` or ~~~), joined again: a heading
// or a link inside one is an example, not a heading or a link.
export function unfenced(text) {
  const kept = [];
  let fence = null;
  for (const line of text.split("\n")) {
    const match = FENCE.exec(line);
    if (fence) {
      if (match && match[1][0] === fence.char && match[1].length >= fence.length && match[2].trim() === "") {
        fence = null;
      }
      continue;
    }
    if (match && !(match[1][0] === "`" && match[2].includes("`"))) {
      fence = { char: match[1][0], length: match[1].length };
      continue;
    }
    kept.push(line);
  }
  return kept.join("\n");
}

// A heading's text as GitHub's anchor for it: link text kept, emphasis markup
// dropped, other punctuation gone, each space a hyphen (never collapsed, so
// "Phase 5 — The frame" gives "phase-5--the-frame").
export function slug(heading) {
  let text = heading.replace(/\[([^\]]*)\]\([^)]*\)/g, "$1");
  text = text.replace(/`/g, "").toLowerCase();
  text = text.replace(/[^\p{L}\p{N} _-]/gu, "");
  return text.replace(/ /g, "-");
}

// Every anchor a Markdown text offers: one for each ATX heading, numbered
// after a repeat ("a", "a-1"), plus any explicit <a id> or <a name>.
export function headingAnchors(text) {
  const seen = new Set();
  for (const [, name] of unfenced(text).matchAll(/<a\s+(?:id|name)="([^"]*)"/g)) {
    seen.add(name);
  }
  for (const line of unfenced(text).split("\n")) {
    const match = /^ {0,3}#{1,6}\s+(.*?)\s*#*\s*$/.exec(line);
    if (!match) continue;
    const base = slug(match[1]);
    let candidate = base;
    for (let n = 1; seen.has(candidate); n++) candidate = `${base}-${n}`;
    seen.add(candidate);
  }
  return [...seen];
}

const LINK = /\[[^\]]*\]\(\s*([^()\s]+)(?:\s+(?:"[^"]*"|'[^']*'))?\s*\)/g;

// The link targets in a text that promise a fragment: relative (no scheme, no
// leading slash) and holding a #. Links in code are examples, not links.
export function anchorTargets(text) {
  const targets = [];
  for (const line of unfenced(text).replace(/`[^`\n]*`/g, "").split("\n")) {
    for (const [, target] of line.matchAll(LINK)) {
      if (!target.includes("#")) continue;
      if (/^[a-z][a-z0-9+.-]*:/i.test(target) || target.startsWith("/")) continue;
      targets.push(target);
    }
  }
  return targets;
}

// README.md and every *.md under docs/, a category's TEMPLATE.md left out.
export function docFiles(root) {
  const files = [];
  const readme = join(root, "README.md");
  if (existsSync(readme)) files.push(readme);
  const walk = (directory) => {
    if (!existsSync(directory)) return;
    for (const entry of readdirSync(directory, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
      const path = join(directory, entry.name);
      if (entry.isDirectory()) walk(path);
      else if (entry.name.endsWith(".md") && basename(path) !== "TEMPLATE.md") files.push(path);
    }
  };
  walk(join(root, "docs"));
  return files;
}

// Returns { checked, missing }: checked counts the anchor links compared with
// a file that exists, missing lists each "file: target" that resolves to no
// anchor (or to no file).
export function anchorReport(root) {
  const anchors = new Map();
  const anchorsOf = (path) => {
    if (!anchors.has(path)) anchors.set(path, new Set(headingAnchors(readFileSync(path, "utf8"))));
    return anchors.get(path);
  };
  const report = { checked: 0, missing: [] };
  for (const file of docFiles(root)) {
    const from = relative(root, file);
    for (const target of anchorTargets(readFileSync(file, "utf8"))) {
      const hash = target.indexOf("#");
      const path = target.slice(0, hash);
      if (path !== "" && !path.toLowerCase().endsWith(".md")) continue;
      const destination = path === "" ? file : resolve(dirname(file), path);
      const link = `${from}: ${target}`;
      if (!existsSync(destination)) {
        report.missing.push(link);
        continue;
      }
      report.checked++;
      let fragment = target.slice(hash + 1);
      try {
        fragment = decodeURIComponent(fragment);
      } catch {
        // A malformed escape stays as written and fails below.
      }
      if (!anchorsOf(destination).has(fragment)) report.missing.push(link);
    }
  }
  return report;
}

test("every anchor link in the docs names a heading that exists", (t) => {
  const { checked, missing } = anchorReport(repo);
  t.diagnostic(`${checked} anchor links checked`);
  assert.ok(checked > 0, "no anchor link was checked");
  assert.deepEqual(missing, []);
});

test("an anchor link to a heading that no longer exists is reported", () => {
  const root = fixture(
    ["docs/a.md", "See [b, gone](b.md#gone) and [b, kept](b.md#kept).\n"],
    ["docs/b.md", "# B\n\n## Kept\n"],
  );
  assert.deepEqual(anchorReport(root).missing, ["docs/a.md: b.md#gone"]);
});

test("a heading's anchor is GitHub's: punctuation dropped, spaces kept as hyphens", () => {
  assert.equal(slug("Phase 2 — Components in the brand (done)"), "phase-2--components-in-the-brand-done");
  assert.equal(slug("Phase 6 — Commands, navigation and settings for v0.2.0 (medium)"),
    "phase-6--commands-navigation-and-settings-for-v020-medium");
  assert.equal(slug("`Shell.View` and [ADR-0006](x.md)"), "shellview-and-adr-0006");
  assert.deepEqual(headingAnchors("# T\n\n## A\n\n## A\n"), ["t", "a", "a-1"]);
  assert.deepEqual(headingAnchors('## Kept\n\n<a name="legacy"></a>\n'), ["legacy", "kept"]);
});

test("a link inside a fenced code block is not checked", () => {
  const root = fixture(
    ["docs/a.md", "## Kept\n\n```md\n[gone](b.md#gone)\n```\n\n~~~\n[gone](a.md#gone)\n~~~\n\n[kept](b.md#kept)\n"],
    ["docs/b.md", "## Kept\n"],
  );
  assert.deepEqual(anchorReport(root), { checked: 1, missing: [] });
});

test("external, absolute, non-Markdown and same-file links", () => {
  const root = fixture([
    "docs/a.md",
    [
      "## Kept",
      "",
      "[web](https://example.com/x.md#gone)",
      "[root](/docs/b.md#gone)",
      "[image](pic.png#gone)",
      "[same](#kept)",
      "[same, gone](#gone)",
      "[other](../docs/a.md#kept)",
      "`[code](a.md#gone)`",
      "",
    ].join("\n"),
  ]);
  assert.deepEqual(anchorReport(root), { checked: 3, missing: ["docs/a.md: #gone"] });
});
