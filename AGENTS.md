# frostui

frostui is frostyard's desktop UI kit, a Go library on
[MyGo](https://github.com/egoist/mygo) (`github.com/egoist/mygo/ui`): the
command-center shell (top bar with a command palette, left rail, console,
status bar), a serial job queue, dialogs, cards, rows and visualizations,
and desktop integration. ChairLift, Frostmail and Frostnotes build their
windows from it. It came out of ChairLift's `internal/frostui`
([ADR-0004](docs/adr/0004-extract-the-ui-kit-into-its-own-module.md)). Start
at [docs/design/frostui.md](docs/design/frostui.md), then
[docs/README.md](docs/README.md).

This file (`AGENTS.md`) is the CANONICAL agent instructions — `CLAUDE.md`,
`GEMINI.md`, `.cursorrules`, and `.github/copilot-instructions.md` are symlinks
to it, and `.claude/skills` symlinks to `.agents/skills/`
([ADR-0002](docs/adr/0002-agent-portable-instruction-surface.md)). Edit only
the canonical paths; keep content tool-agnostic.

## Build, test, lint

- `mise install` — the pinned golangci-lint and Node (`mise.toml`,
  checksums in `mise.lock`). Go comes from go.mod's `go` line, the only Go
  pin. Change a pin only with `mise lock`, in a commit of its own.
- `make test` — every test, headless (`env -u DISPLAY -u WAYLAND_DISPLAY`,
  `CGO_ENABLED=0`). MyGo's Tester needs Pango and cairo, no display, no GTK.
- `make verify` — the gate that leaves `git status` unchanged: `go mod tidy
  -diff`, `go vet`, the gofmt check, the pinned golangci-lint (as linux,
  windows, darwin and js/wasm), the template gate (`scripts/check-template.mjs`
  and `test/*.test.mjs`, Node 22 or later), then `make test`.
- `make check` — `make fmt`, then `make verify`.
- `make ci` — what CI's `verify` job runs: `make verify`, the corrections
  check against `BASE` (`make ci BASE=origin/main`), `make test-wasm` (the
  tests as js/wasm under Node), `make race` and `make cross` (build and vet
  every `CROSS_TARGETS` pair). Run it before you push.
- `make snapshot` — the examples' frames as PNGs in `build/snapshots`
  (`FROSTUI_SNAPSHOT_DIR`); copy `commandcenter.png`, `threepane.png` and
  `sourcelist.png` to `docs/images/` to refresh the docs' pictures.
- `go run ./examples/commandcenter`, `go run ./examples/threepane` and
  `go run ./examples/sourcelist` — the example windows (need GTK 3 and
  libepoxy on Linux).

## Skills (follow these for common tasks)

Step-by-step procedures live in [.agents/skills/](.agents/skills/); follow
them rather than improvising, whichever agent you are. frostyard/core syncs
its skills there (core ADR-0026); start a repo-local one from
`.agents/skills/TEMPLATE/SKILL.md`.

- *(no repo-local skill yet)*

## Code conventions (live — the code exists)

- **Imports.** Every file of the kit, tests included, imports only the
  standard library, MyGo (`github.com/egoist/mygo/...`), purego
  (`github.com/ebitengine/purego`) and the kit's own packages; `frostuitest`
  imports the standard library only. The kit never imports `internal/` or
  `examples/`. go.mod requires MyGo and purego directly and nothing else.
  Guards: `internal/repocheck.TestKitImportsOnlyMyGoAndPurego` (with
  fixtures) and `TestModuleRequiresOnlyMyGoAndPurego`.
- **Cross-platform builds.** The module builds and vets with
  `CGO_ENABLED=0` for linux, windows and darwin on amd64 and arm64, and for
  js/wasm ([ADR-0005](docs/adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)).
  A freedesktop or Linux-only call goes in a `*_linux.go` file, with a
  fallback in a `*_other.go` file (`//go:build !linux`); the exported
  function stays in the untagged file (`appid.go` + `appid_linux.go` +
  `appid_other.go` is the model). A platform file exports nothing. Guards:
  `TestKitCrossCompiles`, `TestPlatformFilesExportNothing`,
  `TestCrossTargetsAreListedEverywhere` (the test's, the Makefile's and
  CI's targets agree).
- **API stability and semver.** `internal/repocheck/testdata/api.txt` pins
  every exported declaration (`TestExportedAPIIsUnchanged`). Adding is a
  minor change (`feat:`). Removing or changing a line breaks every app: it
  needs a major version, or a minor one while the module is v0, a `feat!:`
  commit or `BREAKING CHANGE:` footer, and an entry under "API changes" in
  `docs/design/frostui.md` ([spec](docs/specs/api-stability.md)). Rewrite
  the file with `go test ./internal/repocheck -run
  TestExportedAPIIsUnchanged -update` in the same commit, never by hand.
- **No app-specific code.** Nothing here names or serves one app: no
  ChairLift, Frostmail or Frostnotes wording in code, comments or default
  strings, no app configuration, no command tracer. A helper any app could
  use belongs here; one that is about one app stays in it.
- **UI thread.** Views and every model (`Shell`, `PaneShell`, `Panes`,
  `Jobs`, `Palette`, `Dialogs`, `ConsoleView`, `CommandTable`, `SourceList`,
  `Settings`) belong to the UI thread; only `Console.Add`, `Addf`, `Writer`,
  `Progress.Set`, `Progress.Log`, `Waker.Wake`, `PaletteQuery.Context` and
  `Opener` are safe from any goroutine. Workers hand results back through
  the post function the app gives `NewJobs` (MyGo's `Window.Update`), which
  rebuilds the view on every call: coalesce bursts (`Waker`).
- **MyGo's rules.** MyGo resets the theme every build pass, so `Adapt` runs
  first in every frame (`Shell.View` does it). State (`ListState`, dialog
  and palette state, parsed SVGs) lives in structs, never in a view
  function; a `ListState` belongs to one window. Never create a `ui.Handle`
  in a view or with `ui.Local`: its first `Bind` registers it with the
  window, nothing takes it out until the window closes, and every build
  pass walks them all, so a handle made per build, or per mount (a row
  scrolling back into view, a section opening), leaks. Keep handles in
  structs, one per control: a handle stays registered until the window
  closes even when the kit drops it, so use a bounded set of stable handle
  identities, never one handle per item of a set that can grow or churn
  without limit (`internal/repocheck.TestKitMakesNoHandleInAView`, with
  fixtures). Never load a UI library in an `init` function or a
  package-level variable.
  After a MyGo bump, read its source in `GOMODCACHE` (the theme reset,
  `Window.Update`, the Tester, `Handle` and its registration) before
  merging.
- **Dependency API adoption.** Use the pinned MyGo API explicitly. Do not
  probe for future dependency methods with type assertions: a later bump must
  not silently change the kit's behavior. Adopt new methods in the bump PR,
  with tests for the kit's existing semantics (including idempotent changes).
- **Tests run headless.** Views are tested with `ui.NewTester`; a test that
  stands in for the UI thread posts to a `frostuitest.Loop`. No test
  needs a display, a network or a real desktop app; `LaunchDesktopApp` is
  tested against a temporary desktop entry, and `Opener` through its
  `Launch` and `LaunchFile` fields.
- Run `make verify` before you call any change done, and `make ci` before
  you push. CI job `verify` runs `make ci`, which also checks that
  `.memory/corrections.jsonl` is append-only against the PR base.

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
L2, which declares tier 3. Here go.mod, go.sum, the guards
(`internal/repocheck/**`, with `api.txt`), `.golangci.yml` and
`appid_linux.go` are L2, and `.svu.yml` and the tool pins are L1. Add a path
when you add a file of that kind (signing, auth, publish, secrets,
destructive operations, CI, native code), with a `why` that names its
category. The gate requires schema version 2.

**Vendored files.** `scripts/lib/risk-paths.mjs` and
`scripts/check-corrections.mjs` are copies from frostyard/core; never edit them
by hand (`make verify` checks their hashes).

## Repository boundary

- **Releases.** Tags are `vMAJOR.MINOR.PATCH`; svu derives the next one from
  the Conventional Commits since the last tag (`.svu.yml`: prefix `v`, `v0:
  true`). The maintainer (Brian) runs `make bump`. Agents never run it,
  create or push a tag, or cut a release; an agent that finds a release due
  proposes `svu next` and asks.
- **A library, not a product.** No binaries are published: no GoReleaser,
  no packages, no `dev` prerelease. `examples/` holds programs that show
  the kit, never an app.
- **License.** MIT, as Frostmail and Frostnotes, which import the kit. The
  kit's code came from ChairLift (GPL-3.0-or-later), whose author relicensed
  it. Code copied in must be MIT-compatible: never copy from a GPL source.
  The icons in `icons/` are Lucide's, under ISC: keep `icons/LICENSE` and
  `NOTICE`, and add only MIT, ISC or Apache-2.0 icons, never GPL, LGPL or
  CC-BY-SA ([ADR-0007](docs/adr/0007-embed-a-lucide-icon-set-as-the-last-fallback.md)).
  The fonts in `fonts/` are Inter, under the SIL Open Font License 1.1:
  keep `fonts/OFL.txt` and `NOTICE`, and add only OFL-1.1 fonts
  ([ADR-0011](docs/adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)).
- No secrets, credentials or personal data anywhere in the repository.

## Releases and dependencies

The facts that recurring repository checks read. The release rules are the
**Releases** bullet of [Repository boundary](#repository-boundary); they are
not repeated here.

- Tag format: `vMAJOR.MINOR.PATCH` (for example `v0.6.0`), 1 series at the
  repository root.
- How a release starts: `make bump` tags the next version and pushes the
  tag. No release workflow runs and no GitHub release is made: the module is
  a library, and the Go module proxy serves the tag.
- Who tags: the maintainer. Agents never tag, push a tag or release.
- Version rule: `svu next` from the Conventional Commits since the last tag
  (`.svu.yml`).
- Internal dependencies: none. go.mod requires only MyGo
  (`github.com/egoist/mygo`) and purego, which frostyard does not own; after
  a MyGo bump, follow **MyGo's rules** under Code conventions.
- Who uses this repository: frostyard/chairlift, frostyard/frostmail,
  frostyard/frostnotes, frostyard/firn (`gui/`) and frostyard/pocket, each
  through its `go.mod`.

## Documentation rules (enforced)

Docs live in `docs/` in four categories. **Every new doc starts from its
category's `TEMPLATE.md`** and follows its structure:

- `docs/adr/` — why we decided. Immutable once Accepted; reversals are new
  ADRs that mark the old one Superseded. This includes ADRs 0001 to 0003
  that came with the template: never edit their decision or consequences.
  A new org-wide rule goes in a core ADR listed in `docs/org-adrs.md`; a new
  project rule goes in a new ADR.
- `docs/design/` — how it fits together. Living; updated in place to match
  reality. `docs/design/frostui.md` is the entry point.
- `docs/specs/` — exact contracts. Change only alongside implementing code.
- `docs/plans/` — order of work. Phases with "Done when" outcomes.

After a change to the kit, update `docs/design/frostui.md`, `README.md` and
this file where they describe it; check a current-state claim against the
source, never against an old document.

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
