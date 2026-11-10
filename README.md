<!-- readme-kind: org -->
# agentic-template

A GitHub template for repositories that LLM agents build and maintain, for maintainers and teams who want reliable, reviewable agent work. It provides no application code; it provides portable agent instructions, a documentation structure with templates, a verify gate, and a starting point for skills.

**Status:** in production. New repositories copy it from `main`; it has no versioned releases.

**Documentation:** [docs/README.md](docs/README.md)

## What is here

| Path | What it is |
| --- | --- |
| `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md` | Canonical agent instructions and tool-facing symlinks |
| `.agents/skills/`, `.claude/skills` | Canonical shared skills and the Claude skills symlink |
| `docs/` | ADRs, designs, specs, plans and documentation indexes |
| `Makefile` | The verify, check and CI gate |
| `mise.toml`, `mise.lock` | Pinned tool versions |
| `scripts/` | Template and corrections validation scripts |
| `test/` | Node.js fixtures and gate tests |
| `.memory/` | The append-only corrections inbox |
| `.github/` | Risk paths, dependabot settings, pull request guidance and workflows |
| `LICENSE`, `SECURITY.md` | License text and private vulnerability reporting |

## Use

1. Click **Use this template** on GitHub, or run `gh repo create myproject --template bketelsen/agentic-template`.
2. In `AGENTS.md`, replace the title and intro, then fill in the commented placeholder sections as the project takes shape. Put project checks behind `make verify` in the `Makefile`; CI runs the gate.
3. Set the dates in `docs/adr/0001`, `0002` and `0003` to today; these ADRs are pre-accepted because this template implements them.
4. Set the copyright holder and year in `LICENSE`. Keep MIT unless the project inherits GPL-3.0.
5. Replace the repository URL and supported-versions section in `SECURITY.md` with your own values.
6. Add a `.github/dependabot.yml` entry for each package ecosystem that the project adds: `gomod`, `npm`, `cargo`, `docker`. Give each entry 1 group with `patterns: ["*"]` and no `applies-to`, like the `github-actions` entry (core [ADR-0075](https://github.com/frostyard/core/blob/main/docs/adr/0075-batch-routine-pull-requests.md)).
7. For a released Go project, add `.svu.yml` and a `bump` target that runs `svu next`; use the Go and golangci-lint versions of core's `frostyard-go-repo` skill.
8. For a user-facing tool, add a `site/` documentation site (core [ADR-0070](https://github.com/frostyard/core/blob/main/docs/adr/0070-build-tool-docs-sites-with-materialx-on-github-pages.md)).
9. Write the first real ADR, `0004`, for the first significant choice — language, framework or storage — and continue for later choices.
10. Write `README.md` for the new project from core's [frostyard-readme skill](https://github.com/frostyard/core/blob/main/.agents/skills/frostyard-readme/SKILL.md) and [core ADR-0072](https://github.com/frostyard/core/blob/main/docs/adr/0072-write-every-readme-to-one-standard.md). Run its checker, `check-readme.mjs`, until it prints `ok`. Update [docs/design/agent-ready-repository.md](docs/design/agent-ready-repository.md) when the project changes its structure.
11. Complete the repository admin steps for the labels, ruleset, private vulnerability reporting and risk paths in [docs/new-repository.md](docs/new-repository.md).

The full checklist is core's [repository baseline spec](https://github.com/frostyard/core/blob/main/docs/specs/repository-baseline.md).

## How it works

The template makes agents use one canonical instruction surface, category-based documentation, reusable templates and skills, two-way cross-links, live conventions and a verify gate. It keeps the mechanism in `AGENTS.md`, docs and code, so each agent sees the same repository. For the full design, read [agent-ready repository](docs/design/agent-ready-repository.md).

## Develop

```sh
mise install
make ci
```

`make verify` needs only Node.js 22 or later and `make`; it uses no package manager and no third-party dependency. Clone with `core.symlinks=true` on Windows.

## Where to go next

- [AGENTS.md](AGENTS.md): the rules for agents and contributors.
- [docs/README.md](docs/README.md): the index of decisions, designs, specs and plans.
- [SECURITY.md](SECURITY.md): report a vulnerability privately.
- [LICENSE](LICENSE): MIT.
