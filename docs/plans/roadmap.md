# Plan: Roadmap

This plan takes the kit from ChairLift's `internal/frostui` to a released
module that ChairLift, Frostmail and Frostnotes import
([ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md)).
[The Frostyard look](frostyard-look.md) is the plan that follows it.
Since then, two more apps import the module: the Firn GUI installer
(the `gui/` module of frostyard/firn) and the Pocket client
(frostyard/pocket).

Each phase's evidence was checked on 2026-11-12, against frostui's
`main` at 14ef2d6 (tag v0.6.0) and the apps' `main` branches. A phase is
marked done only when every part of its "Done when" holds by evidence that
was checked; a part that does not hold, or that nobody checked, keeps the
phase open, and its state paragraph says which part.

## Phase 1 — Extract the kit (done)

- Move chairlift's `internal/frostui` at 09e7184 to the module root, with
  `frostuitest`, and keep the exported API
  ([design](../design/frostui.md#moved-from-chairlift),
  [spec](../specs/api-stability.md)).
- Build for Linux, Windows, macOS and js/wasm with the freedesktop parts
  behind build tags
  ([ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md),
  [design](../design/frostui.md#platforms)).
- Guards, the command-center example and its snapshot, and the template's
  gates in `make verify` and CI.

**Done when:** `make ci` passes on branch `feat/extract`, and the API
extractor gives the same list for this module and chairlift's copy.

**Evidence:** frostyard/frostui#2 (`feat/extract`) merged with the `verify`
check (`make ci`) green; its description records that `api.txt` matches
chairlift's `internal/frostui` line for line.

## Phase 2 — First release (small, done)

- Merge to `main` with the `verify` check green, then the maintainer runs
  `make bump` for the first tag.

**Done when:** a `v0.x.y` tag exists and `go get github.com/frostyard/frostui@latest`
resolves it.

**Evidence:** tag v0.1.0 is frostyard/frostui#2's merge (5b6f83d), and the
Go module proxy resolves `@latest` (v0.6.0 at 14ef2d6).

## Phase 3 — ChairLift imports the module (small, done)

- Follow the steps in [design: moved from ChairLift](../design/frostui.md#moved-from-chairlift).

**Done when:** chairlift's `main` requires `github.com/frostyard/frostui`,
has no `internal/frostui`, and its `make ci` passes.

**Evidence:** frostyard/chairlift#282 (commit 0a2de1b) moved it onto the
module. Chairlift's `main` (94e89df) requires v0.5.2, has no
`internal/frostui`, and its `make ci` passes.

## Phase 4 — Frostmail and Frostnotes start from the shell (medium, done)

- Each app builds its window from `Shell`, `Jobs` and `Console`, as
  `examples/commandcenter` does, and embeds the icons it draws off Linux
  ([design: icons](../design/frostui.md#icons)).

**Done when:** both apps open the shell on Linux and Frostmail on Windows.

**Evidence:** frostyard/frostmail#223 and frostyard/frostnotes#145 moved
the apps onto the kit. Both build their windows from `PaneShell`, the frame
without the rail ([ADR-0006](../adr/0006-a-frame-without-the-rail.md)),
not from `Shell`; Frostmail keeps a `Console`, and neither app uses `Jobs`.
On the apps' `main` branches, CI starts the built app in a real window:
Frostnotes' `make desktop-smoke` under Xvfb (e090378), and Frostmail's
end-to-end tests under Xvfb on Linux and on Windows (9f5d030).

## Phase 5 — The frame and the layout (done)

The first half of v0.2.0's shared pieces, for Frostmail and Frostnotes:
`PaneShell` ([ADR-0006](../adr/0006-a-frame-without-the-rail.md),
[design](../design/frostui.md#the-frame-without-the-rail)),
[`Panes`](../design/frostui.md#panes),
[connection status](../design/frostui.md#connection-status),
[the action toast](../design/frostui.md#action-toast),
[the waker](../design/frostui.md#the-coalescing-waker) and
[the icon set](../design/frostui.md#the-kits-icon-set)
([ADR-0007](../adr/0007-embed-a-lucide-icon-set-as-the-last-fallback.md)).

**Done when:** `examples/threepane` shows all of them, and `make ci` passes
on branch `feat/frame-panes`.

**Evidence:** frostyard/frostui#4 (`feat/frame-panes`) merged with the
`verify` check (`make ci`) green. At its merge (b3143ce),
`examples/threepane` uses `PaneShell` with its `Panes`, `Connection`,
`ActionToast`, `NewWaker` and the embedded icons (`NewIcons`, `SetIcons`).

## Phase 6 — Commands, navigation and settings for v0.2.0 (medium)

- The second half: a command table, palette sources, a source list, a settings layout and
  an allow-listed opener, for Frostmail's and Frostnotes' desktop apps
  ([design: command table](../design/frostui.md#command-table),
  [palette sources](../design/frostui.md#palette-sources),
  [source list](../design/frostui.md#source-list),
  [settings layout](../design/frostui.md#settings-layout),
  [opener](../design/frostui.md#opener),
  [ADR-0008](../adr/0008-open-behind-allow-lists-and-let-editors-keep-their-keys.md)),
  additive to the API ([spec](../specs/api-stability.md)).

**Done when:** `examples/sourcelist` draws them all, `make ci` passes, and
`api.txt` only gains lines against v0.1.0.

**State:** open: `examples/sourcelist` does not use the `Opener`. The rest
holds. frostyard/frostui#5 merged with the `verify` check (`make ci`)
green, and tag v0.2.0 is its merge (a51e185). There, `examples/sourcelist`
draws the `CommandTable`, the palette sources, the `SourceList` and the
`Settings` pages. `opener_test.go` tests the `Opener`, which draws
nothing, but no example opens a link or a file through it. From v0.1.0 to
v0.2.0, `api.txt` gains 230 lines and loses none.

## Phase 7 — Native diff viewer

- Add a reusable read-only unified diff to the [building blocks](../design/frostui.md#building-blocks-rows-and-visualizations),
  additive to the [API](../specs/api-stability.md).

**Done when:** callers can display unified diffs with old/new line numbers,
wrapped and selectable text, virtualized rows, and accessible signed changes;
parser and headless rendering tests pass in `make ci`.

## Open questions

- Should `Adapt` leave MyGo's inactive-window accent alone on macOS, where
  fading it is the native look?

## References

- Implements: [design: frostui](../design/frostui.md),
  [spec: API stability](../specs/api-stability.md)
- Builds on: [ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md),
  [ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md),
  [ADR-0006](../adr/0006-a-frame-without-the-rail.md),
  [ADR-0007](../adr/0007-embed-a-lucide-icon-set-as-the-last-fallback.md),
  [ADR-0008](../adr/0008-open-behind-allow-lists-and-let-editors-keep-their-keys.md)
- Followed by: [the Frostyard look](frostyard-look.md)
