# Plan: Roadmap

This plan takes the kit from ChairLift's `internal/frostui` to a released
module that ChairLift, Frostmail and Frostnotes import
([ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md)). It is
the only plan of this repository.

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

## Phase 2 — First release (small)

- Merge to `main` with the `verify` check green, then the maintainer runs
  `make bump` for the first tag.

**Done when:** a `v0.x.y` tag exists and `go get github.com/frostyard/frostui@latest`
resolves it.

## Phase 3 — ChairLift imports the module (small)

- Follow the steps in [design: moved from ChairLift](../design/frostui.md#moved-from-chairlift).

**Done when:** chairlift's `main` requires `github.com/frostyard/frostui`,
has no `internal/frostui`, and its `make ci` passes.

## Phase 4 — Frostmail and Frostnotes start from the shell (medium)

- Each app builds its window from `Shell`, `Jobs` and `Console`, as
  `examples/commandcenter` does, and embeds the icons it draws off Linux
  ([design: icons](../design/frostui.md#icons)).

**Done when:** both apps open the shell on Linux and Frostmail on Windows.

## Open questions

- Should the kit ship a small icon set for its own 20 icon names, so apps
  off Linux need not embed them?
- Should `Adapt` leave MyGo's inactive-window accent alone on macOS, where
  fading it is the native look?
