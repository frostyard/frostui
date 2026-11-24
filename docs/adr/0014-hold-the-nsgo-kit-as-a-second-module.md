# 0014 — Hold the NSGo kit as a second module, and give it the Frostyard look

- **Status:** Accepted
- **Date:** 2026-11-24

## Context

frostyard's apps are moving from MyGo to
[NSGo](https://github.com/nsgo-dev/nsgo). Three have an NSGo client beside
their MyGo one, and the three share `workspaceui`: a shell with a
navigation sidebar and two side panels, a file explorer, a conversation
view, a notice and a few surfaces. It is 6 files, a module of its own in
frostyard/pocketgo (`desktop/workspaceui/v0.1.0-rc.1`), and it sets no
theme. Each client draws NSGo's stock grey and blue, and the three differ
in their buttons, dialogs, notices and paddings.

The Frostyard look is this repository's: the two themes and their measured
contrast ([ADR-0011](0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md),
[ADR-0013](0013-hold-text-to-7-to-1-under-high-contrast.md),
[the themes' spec](../specs/frostyard-theme.md)) and Inter (`fonts/`). The
kit at the repository's root draws with MyGo, and its imports are held to
MyGo and purego
([ADR-0004](0004-extract-the-ui-kit-into-its-own-module.md)).

NSGo v0.3.0 sets what a kit on it can change:

- A theme sets every color role, named roles of the kit's own, and the
  fonts. NSGo's `ui` package fixes its corners (8 for a control, 12 for a
  card, 4 for a small part), its heights and its type scale.
- Under a hidden title bar, the GTK backend draws no window controls and
  has no call that resizes the window from its edges.

The repository has 1 series of tags, `vX.Y.Z`, and `make bump` reads the
newest tag with svu. Go reads a nested module's versions from tags named
`<directory>/vX.Y.Z`.

## Decision

1. **A second module.** `workspaceui/` is a Go module of its own,
   `github.com/frostyard/frostui/workspaceui`, beside the MyGo kit. Its
   package keeps the name `workspaceui`. It starts as the 6 files of
   `desktop/workspaceui/v0.1.0-rc.1`, unchanged but for the module path.
2. **Its imports.** The module imports the standard library and NSGo only,
   and its go.mod requires NSGo directly and nothing else. It never imports
   MyGo or the root module, and the root module never imports it. A test in
   the module holds this.
3. **Its gates.** Every gate of `make verify` and `make ci` runs in both
   modules: tidy, vet, lint, the tests (headless, as js/wasm and under the
   race detector) and the cross-build for every target.
4. **Its tags.** The module has a tag series of its own,
   `workspaceui/vX.Y.Z`, from the Conventional Commits that touched
   `workspaceui/`. The maintainer runs `make bump-workspaceui`. `make bump`
   reads only `v*` tags.
5. **The look.** The kit draws Frostyard's colors in a dark and a light
   theme that follow the desktop, set in Inter, from the values of the
   themes' spec. It keeps NSGo's corners: Frostyard's nearly square corners
   (ADR-0011) stay the MyGo kit's.
6. **The title bar.** On Linux the kit keeps the system's title bar and
   draws its toolbar under it, until NSGo can resize an undecorated GTK
   window and show its controls. On Windows and macOS the shell's header is
   the title bar.
7. **The selected row.** A chosen row takes the accent at a low alpha with
   rounded corners and keeps its own text colors, as `SoftSelection` does
   in the MyGo kit.

Points 5 to 7 are not built yet;
[the plan](../plans/workspaceui.md) orders them.

## Consequences

- The three apps take the look by a version bump, and a change of the look
  is a change here.
- The repository holds two kits with two import rules and two tag series.
  A doc, a guard or a rule says which module it is about.
- The root's pinned API list (`internal/repocheck/testdata/api.txt`) does
  not cover the new module. Its API changes freely while it is v0 and the
  parts of the plan land; a pin comes with its first stable shape.
- A commit under `workspaceui/` also counts toward the root's next version
  (`.svu.yml` reads every directory), so the root's version can move with
  no change to the MyGo kit.
- NSGo and what it requires enter the repository's `go.sum` files, and
  Dependabot watches a second go.mod.
- The MyGo kit and the NSGo kit differ in their corners. An app that ships
  both clients shows both.
- On Linux a window has two rows at its top until NSGo closes its gap.

## Alternatives considered

- **Leave the kit in frostyard/pocketgo:** a change of the shared look
  would pass one app's gates and releases, away from the look's spec, its
  contrast table and its font.
- **A new repository:** a whole repository baseline to build and keep for 6
  files.
- **Move the root module to NSGo:** ChairLift, the Firn installer and the
  MyGo clients still import the MyGo kit.
- **One module with both kits:** every app would require both frameworks.
- **Draw every control from NSGo's `base` to keep square corners:** the kit
  would own each control's behaviour for a difference of a few points.
- **Draw window controls and resize edges in the kit on Linux:** it copies
  what the framework should do, on the system where a wrong resize edge
  leaves a window the user cannot size.
- **A full accent fill for the chosen row:** every text and mark of the row
  needs a second color, and a row's own marks (a flag, an unread dot) lose
  theirs.

## References

- Shapes: [design/workspaceui.md](../design/workspaceui.md),
  [specs/workspaceui.md](../specs/workspaceui.md),
  [plans/workspaceui.md](../plans/workspaceui.md); and
  [design/frostui.md](../design/frostui.md), the MyGo kit, which keeps its
  module, its import rule and its tag series
- Builds on: [ADR-0004](0004-extract-the-ui-kit-into-its-own-module.md),
  [ADR-0005](0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md),
  [ADR-0011](0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md),
  [ADR-0013](0013-hold-text-to-7-to-1-under-high-contrast.md)
- frostyard/frostmail: ADR-0062 and plan 0016, where the owner chose the
  look, the kit's home, the title bar and the selected row (2026-11-24)
- frostyard/pocketgo: ADR-0068, which made the module
