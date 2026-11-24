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

The module is one package of 6 files:

| File | Parts |
| --- | --- |
| `shell.go` | `Shell`, the window: a resizable navigation sidebar (NSGo's `ui.SidebarShell`), the content, an optional `Inspector` at the right and an auxiliary panel beside the content, with a one-pane layout for narrow windows. `PaneWidths` and `OnWidthsChange` let an app keep the panes' widths. `Page`, `Row` and `Stack` lay content out. |
| `sidebar.go` | `BranchLabel` and `ChildLabel`, the content of a group row and of its indented children in the navigation. |
| `sections.go` | `Section` (a disclosure with a count), `Card`, `InspectorGroup` and `InspectorRow`. |
| `conversation.go` | `Conversation`, `Message` and `ConversationEntries`, over NSGo's chat view. |
| `explorer.go` | `FileExplorer`: a tree beside its preview. |
| `notice.go` | `Notice`: one dismissible message over the window. |

Every part reads its colors from the current theme, so it follows the
desktop's light and dark appearance. The module sets no theme of its own
yet: an app draws NSGo's default palette until
[the plan](../plans/workspaceui.md)'s Phase 1.

On Linux, `Shell` keeps the system's title bar: NSGo's GTK backend draws
no window controls and resizes no window under a hidden one (ADR-0014).

## Operational notes

- **Gates.** The Makefile's `MODULES` lists both modules, and `make
  verify`, `make ci` and `make cross` run each step in each. The module's
  tests run headless on NSGo's `nsgotest`, with no display.
- **Guards.** `workspaceui/imports_test.go` holds the module's imports and
  its one direct requirement. The root's guards (`internal/repocheck`)
  skip a directory with a go.mod of its own, so they do not read this
  module, and the root's pinned API list does not cover it.
- **Releases.** Tags are `workspaceui/vX.Y.Z`. The maintainer runs `make
  bump-workspaceui`; the first tag is named (`VERSION=workspaceui/v0.1.0`),
  since svu has no tag to count from. `make bump` reads `v*` tags only.
- **After an NSGo bump**, read its release notes for `ui.SidebarShell`,
  the title bar on Linux and the theme's roles before merging.

## References

- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- Contracts: [specs/workspaceui.md](../specs/workspaceui.md)
- Built in: [the workspaceui plan](../plans/workspaceui.md), Phase 0
