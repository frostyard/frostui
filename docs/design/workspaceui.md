# workspaceui: the kit for NSGo apps

Living document. Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md).
Contracts: [specs/workspaceui.md](../specs/workspaceui.md).

## Overview

`workspaceui/` is a second Go module in this repository,
`github.com/frostyard/frostui/workspaceui`. It composes
[NSGo](https://github.com/nsgo-dev/nsgo)'s themed controls into the window
that frostyard's NSGo apps share. An app supplies destinations, content and
actions; the kit decides how they are laid out. The MyGo kit at the
repository's root ([frostui](frostui.md)) and this module share no code.

```
app (FrostMail, Frostnotes, PocketGo)
  └─ workspaceui ──► NSGo (core, base, ui, desktop, icons/lucide)

frostui (root module) ──► MyGo        no import in either direction
```

## Design

The module is one package:

| File | Parts |
| --- | --- |
| `shell.go` | `Shell`, the window: a resizable navigation sidebar (NSGo's `ui.SidebarShell`), the content, an optional `Inspector` at the right and an auxiliary panel beside the content, with a one-pane layout for narrow windows. `PaneWidths` and `OnWidthsChange` let an app keep the panes' widths. `Page`, `Row` and `Stack` lay content out. |
| `sidebar.go` | `BranchLabel` and `ChildLabel`, the content of a group row and of its indented children in the navigation. |
| `sections.go` | `Section` (a disclosure with a count), `Card`, `InspectorGroup` and `InspectorRow`. |
| `conversation.go` | `Conversation`, `Message` and `ConversationEntries`, over NSGo's chat view. |
| `explorer.go` | `FileExplorer`: a tree beside its preview. |
| `notice.go` | `Notice`: one dismissible message over the window. |
| `theme.go` | `Theme`, the Frostyard look as an NSGo theme with a dark and a light palette; `Apply`, which sets it and registers Inter; `Frame`, the window's root view, which sets the font family. `fonts/` holds Inter's upright variable file. |
| `colors.go` | `ColorsOf` and `Tone`: the Frostyard colors NSGo has no role for (the chrome, the panel, the chosen row's fill, the label, each tone's fill and line). |
| `appearance.go` | `AppearanceChoice`: the System, Light or Dark setting, as a segmented control. |
| `parts.go` | The parts NSGo has no control for, and the content of its rows: `NavLabel` (a sidebar row's icon, name and count), `SummaryRow` (a list row of a sender, a subject and a preview), `SoftRows` (the rows' hover and chosen fills), `Banner`, `EmptyState`, `DialogPanel` (a floating panel's opaque ground), `ConnectionItem` and `SearchField`. |

`workspaceui/gallery/` is a third module, an app: the standard window on
fixture mail, and a page for each family of controls (colors and type,
buttons and toolbars, fields and forms, rows and marks, dialogs and
menus, search and trees and popovers, settings). It is where a change of
the look is drawn and accepted before an app takes it. Its picture tours
keep 30 pictures in
`workspaceui/gallery/comparison/linux`, each view in both themes.

Every part reads its colors from the current theme, so it follows the
desktop's light and dark appearance. An app takes the look with one call
at its root and one view around its content:

```go
func root(cx *core.Scope) core.Node {
	_ = workspaceui.Apply(cx) // the theme, and Inter
	return workspaceui.Frame(workspaceui.Shell{...})
}
```

NSGo's controls then draw Frostyard's colors, since the palettes set
NSGo's own roles; a kit part or an app reads the other colors with
`ColorsOf(cx)`, inside a `StyleFunc`. `Theme.Sans` reaches only classes
(`font-sans`), so `Frame` sets the family for the views inside it. An app
that keeps the user's choice of appearance calls `cx.SetAppearance` after
`Apply`, and shows the choice with `AppearanceChoice`.

On Linux, `Shell` keeps the system's title bar: NSGo's GTK backend draws
no window controls and resizes no window under a hidden one (ADR-0014).

## Operational notes

- **The gallery on the web.** `make gallery-pictures-web` runs the tours
  in a headless Chromium (`NSGO_BROWSER` names one) and writes
  `workspaceui/gallery/comparison/web`, which git ignores: its 30 pictures
  are 2560 by 1640 and 6.9 MB. A tour is one test for each theme, since
  on the web target only the first app a test process starts writes its
  pictures.
- **The gallery.** `make gallery` opens it in a window. `make
  gallery-live` runs its tours in a GTK 4 window under Xvfb, and `make
  gallery-pictures` takes the kept pictures again; a change of the look
  commits its changed pictures. The targets set `GDK_BACKEND=x11`: without
  it GTK opens the windows on a Wayland desktop even under `xvfb-run`. Two
  takes of the pictures give the same bytes; a control that animates (a
  busy button's spinner) is left out of them for that reason. `go build`
  in the gallery's directory writes a binary named `gallery`, which git
  ignores.
- **Gates.** The Makefile's `MODULES` lists the three modules, and `make
  verify`, `make ci` and `make cross` run each step in each. The gallery's
  tours run headless there; no gate runs them in a real window. The module's
  tests run headless on NSGo's `nsgotest`, with no display.
- **Guards.** `workspaceui/imports_test.go` holds the module's imports and
  its one direct requirement. The root's guards (`internal/repocheck`)
  skip a directory with a go.mod of its own, so they do not read this
  module, and the root's pinned API list does not cover it.
- **Releases.** Tags are `workspaceui/vX.Y.Z`; the first is
  `workspaceui/v0.1.0`. The maintainer runs `make bump-workspaceui`, which
  takes the next version from svu, or the one `VERSION` names. `make bump`
  reads `v*` tags only.
- **Fonts.** `Apply` hands Inter's bytes to the platform
  (`Scope.RegisterFont`); a headless test registers nothing. The file adds
  about 0.9 MB to an app's binary, the web build's included.
- **After an NSGo bump**, read its release notes for `ui.SidebarShell`,
  the title bar on Linux and the theme's roles before merging.

## References

- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- Contracts: [specs/workspaceui.md](../specs/workspaceui.md)
- Built in: [the workspaceui plan](../plans/workspaceui.md), Phase 0
