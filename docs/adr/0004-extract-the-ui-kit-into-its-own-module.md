# 0004 — Extract the UI kit into its own module

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

frostyard/chairlift draws its window with MyGo (`github.com/egoist/mygo`)
through a kit of its own, `internal/frostui`: the command-center shell, the
command palette, the console, a serial job queue, dialogs, cards, rows,
pipeline and gauges, desktop integration and a test helper package,
`frostuitest` (chairlift ADR-0017, design `frostui.md`). Frostmail and
Frostnotes are MyGo desktop apps that need the same shell and pieces. Go
does not let another module import a package under `internal/`. The kit
already imports only the standard library, MyGo and purego, and ChairLift
guards that rule with `TestFrostUIImportsNoChairLift`.

## Decision

- **One module, two packages.** `github.com/frostyard/frostui` holds the
  kit at its root as package `frostui`, and `frostuitest` as a subpackage.
  The code is chairlift's `internal/frostui` at commit 09e7184. go.mod
  takes chairlift's `go` line (1.27.3) and its MyGo (v0.8.3) and purego
  (v0.11.1) versions.
- **Scope.** The kit holds what any frostyard desktop app on MyGo can use:
  the shell, palette, console, job queue, dialogs, building blocks, rows,
  wizard parts, toasts, visualizations, desktop integration and the
  UI-thread test loop. App code stays in the apps: a command tracer, page
  logic, navigation inventories, configuration and translations. The kit's
  default strings are English and name no app.
- **Imports.** Every file of the kit, tests included, imports only the
  standard library, MyGo, purego and the kit's own packages;
  `frostuitest` imports the standard library only. go.mod requires MyGo
  and purego directly and nothing else. `internal/repocheck` holds the
  guards and `examples/` the example programs; the kit imports neither.
- **API.** The exported identifiers and signatures are chairlift's,
  unchanged, so ChairLift moves by changing its import paths.
  `internal/repocheck/testdata/api.txt` pins the exported API. The module
  is versioned with semver tags `vX.Y.Z` that svu derives from
  Conventional Commits (`.svu.yml`, `make bump`). An addition is a minor
  version. A removal or a changed signature breaks callers: it needs a
  major version, or a minor one while the module is v0, and a note in the
  design doc.
- **License.** The module is GPL-3.0-or-later, the license of the
  ChairLift code it comes from.

## Consequences

- Frostmail, Frostnotes and ChairLift import one kit and get its fixes
  and tests in one place; a change to the kit is reviewed once.
- ChairLift replaces `github.com/frostyard/chairlift/internal/frostui`
  with `github.com/frostyard/frostui` and deletes its copy. Its own
  import guard keeps its intent in this repository as
  `TestKitImportsOnlyMyGoAndPurego`.
- The Pipeline list's accessible name is "Pipeline", not ChairLift's
  "Update run"; an app names its own run by chaining `Label` on the
  element `Pipeline` returns.
- A breaking change to the kit is a release of its own, followed by a
  change in each app. `TestExportedAPIIsUnchanged` makes every API change
  visible in review.
- An app that imports the kit is a combined work under the GPL: it must
  carry a GPL-compatible license.
- MyGo is young and has one main author. The kit is the one place that
  calls it for the apps' shells, and a MyGo bump is read in its source
  before it merges.

## Alternatives considered

- **Copy the kit into each app:** each copy drifts, and a fix lands three
  times.
- **A nested module inside the chairlift repository:** couples the kit's
  releases and reviews to ChairLift's, and its tags to ChairLift's tag
  scheme.
- **Split the kit into several packages (shell, jobs, widgets):** renames
  every identifier ChairLift uses, for no gain while one team owns all
  three apps.

## References

- Shapes: [design: frostui](../design/frostui.md), [spec: API stability](../specs/api-stability.md), [plan: roadmap](../plans/roadmap.md), `AGENTS.md`
- Builds on: [ADR-0001](0001-record-architecture-decisions.md), [ADR-0003](0003-verify-gate-risk-paths-and-corrections.md)
- Source: frostyard/chairlift ADR-0017, `docs/adr/0017-mygo-command-center-window.md`, and its design `docs/design/frostui.md`, at commit 09e7184
