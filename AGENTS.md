# Project Name

<!-- One paragraph: what this project is and who it serves. Point at
docs/README.md as the entry point for everything else. Keep this file under
~200 lines — it is loaded into every agent session; long files dilute the
rules that matter. -->

One or two sentences describing the project. Start at
[docs/README.md](docs/README.md).

This file (`AGENTS.md`) is the CANONICAL agent instructions — `CLAUDE.md`,
`GEMINI.md`, `.cursorrules`, and `.github/copilot-instructions.md` are symlinks
to it, and `.claude/skills` symlinks to `.agents/skills/`
([ADR-0002](docs/adr/0002-agent-portable-instruction-surface.md)). Edit only
the canonical paths; keep content tool-agnostic.

## Skills (follow these for common tasks)

Step-by-step procedures live in [.agents/skills/](.agents/skills/); follow
them rather than improvising, whichever agent you are:

<!-- One bullet per skill: **When to use it** → [.agents/skills/<name>/SKILL.md].
Add a skill whenever you find yourself re-explaining a multi-step procedure.
Start from .agents/skills/TEMPLATE/SKILL.md. -->

- *(none yet — add the first one when a procedure repeats)*

## Code conventions (live — the code exists)

<!-- The most important section. Rules here must describe the code AS IT IS,
not aspirations — an agent that follows a stale rule produces broken work.
Graduate a rule into this list only when the code enforcing or exemplifying
it has landed; until then it lives in a design doc as intent.

Write rules imperatively and concretely, each with enough mechanism to be
followed without asking ("Storage only via db.Open(slug, migrations)" — not
"use the database layer"). Point at one canonical example in the code for
every structural rule. Rules that remove a degree of freedom are the
valuable ones: every choice an agent doesn't have to make is a failure mode
removed. -->

- *(add rules as the code that enforces them lands)*
- Run `make verify` before you call any change done. CI job `verify` runs
  `make ci`, which also checks that `.memory/corrections.jsonl` is append-only
  against the PR base; run `make ci BASE=origin/main` for the same check locally.
- `mise.toml` and `mise.lock` pin every tool the gate runs (ADR-0043 in core);
  run `mise install` to get those versions. Change a pin only with `mise lock`,
  in a commit of its own.

**Test reliability.** No sleeps for timing (wait on a condition or inject a
clock). No network and no wall clock in unit tests (inject them). A flaky test
gets an issue, then a fix or a quarantine (skip with the issue link) in the
same PR. Never retry a job or a test until it is green.

**Blocker to guard.** When a review (agent cross-review, owner review, or a
human) finds a blocking point, the PR that fixes it also adds a guard: a
targeted test that fails without the fix (preferred), or an `AGENTS.md` rule
when no test can catch it. The same PR appends 1 line to
`.memory/corrections.jsonl` (`evidence`: the PR and the review;
`promoted_to`: the test or the rule). See `.memory/README.md`.

**Risk paths.** `.github/risk-paths.json` is a level list (schema version 2):
every glob carries a level, and a PR that touches one gets that level label from
`.github/workflows/risk-paths.yml` (`level:L1` read every line, `level:L2` read
the diff, `level:L3` skim, `level:L4` title and CI), plus `risk:high` at L1 and
L2, which declares tier 3. Core computes the level from the matched globs, the
`default_level` and the org floors, and it never lowers a level label that the
PR already has. The same caller removes `ready-for-human` when the head moves
after the label. Add a path when you add a file of that kind (signing, auth,
publish, secrets, destructive operations, CI), with a `why` that names its
category. The gate requires schema version 2.

**Vendored files.** `scripts/lib/risk-paths.mjs` and
`scripts/check-corrections.mjs` are copies from frostyard/core; never edit them
by hand (`make verify` checks their hashes).

## Repository boundary

<!-- What does NOT belong in this repo (secrets, personal data, generated
files that are actually build outputs, apps that belong elsewhere)? How are
releases cut? Delete the section if genuinely not applicable. -->

## Documentation rules (enforced)

Docs live in `docs/` in four categories. **Every new doc starts from its
category's `TEMPLATE.md`** and follows its structure:

- `docs/adr/` — why we decided. Immutable once Accepted; reversals are new
  ADRs that mark the old one Superseded. This includes ADRs 0001 to 0003 that
  come with the template: their `YYYY-MM-DD` date is a placeholder, not a
  draft status. A new org-wide rule goes in a core ADR listed in
  `docs/org-adrs.md`; a new project rule goes in a new ADR.
- `docs/design/` — how it fits together. Living; updated in place to match
  reality.
- `docs/specs/` — exact contracts. Change only alongside implementing code.
- `docs/plans/` — order of work. Phases with "Done when" outcomes.

### Cross-linking is mandatory

A doc without its required links is incomplete — do not finish a docs change
until they exist, in both directions:

- **ADR** → links every design doc/spec it shapes, and prior ADRs it builds on.
- **Design doc** → links the ADR(s) providing its rationale, the spec(s)
  pinning its contracts, and the roadmap phase that builds it.
- **Spec** → links its motivating ADR(s) and the design doc showing where it
  fits.
- **Plan** → every phase links the design docs/specs it implements; resolved
  open questions become ADRs.

When you touch a doc, verify its links still hold (targets exist, section
anchors valid) and add the back-links on the targets. Use relative paths.

### Housekeeping

- New doc ⇒ add a line to the index in [docs/README.md](docs/README.md).
- New significant decision ⇒ new ADR *first*, then update the affected design
  docs/specs in the same change.
- Convert relative dates ("next weekend") to absolute dates in all docs.
