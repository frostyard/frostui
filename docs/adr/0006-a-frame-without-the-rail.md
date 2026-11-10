# 0006 — A frame without the rail, whose top-bar segments live in the panes' columns

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

`Shell` is a command center: a left rail of pages, and a page area that
scrolls, is padded and is at most 1,320 DIPs wide. Frostmail and Frostnotes
are porting their desktop apps to MyGo and frostui. Neither has pages: their
windows are edge-to-edge panes that do not scroll as a whole.

Frostmail's top bar is Mail.app's unified toolbar, in segments that follow
the widths of the panes below them. Frostnotes' is a bar with a centered
palette field. The orchestrating session's rules bind both: the sidebar
toggle at the far left of the bar, the same palette field in every app, the
main menu last before the window controls, the connection first in the
status bar, and Ctrl K opening the palette except inside a text editor.

MyGo's `ui.Split` has two panes and a fixed 40 DIP minimum, and its divider's
value is its own: it has no maximum, no N panes, and no collapsing.

## Decision

- **A second window type.** `PaneShell` with `PaneFrame`, beside `Shell` and
  `Frame`, which do not change. It draws no rail and does not scroll or pad
  its content.
- **`Panes` is the kit's own split,** built on MyGo's public element API: N
  panes with minimum and maximum widths and collapsible sides. Its separators
  are `RoleSeparator` elements with `Range`, and take the arrows, Home and End.
  It stores nothing: `PaneLayout` is a plain struct the app saves.
- **Segments live in the panes' columns.** Each `Pane` has a `Bar`, drawn at
  the top of the pane's own column, so a segment and its pane share one
  width by construction, in the same frame, at any window size. The
  separators run through the top bar.
- **The shared order is the frame's, not the app's.** The frame places the
  sidebar toggle first, and the palette field, the app's `End` items, the
  console's button and the main menu last, so an app cannot get them out of
  order.
- **The shell draws the toasts under the bar.** A toast at the top of the
  window lands on the toolbar button that raised it, and MyGo stops a toast's
  time while the pointer rests on it.

## Consequences

- An app's main window is one call, and the shared rules hold without
  review.
- A collapsed pane takes its segment with it; the toggle and the bar's end
  move to the first and the last pane shown.
- `Panes` reads the room from the last frame's layout, so it asks for one
  more frame when the room changes.
- `Shell` and `PaneShell` hold similar palette and status-bar code, kept
  apart so that `Shell` does not change.
- An app with a console reaches its command through `ConsoleCommand`, or gets
  it added to the palette.

## Alternatives considered

- **Options on `Shell` to drop the rail and the page scroll:** every field
  of `Frame` would then mean something different in each mode.
- **Segments in a separate bar, given the panes' widths:** the bar is built
  before the panes, so it would lag a drag or a narrowing window, or repeat
  the panes' arithmetic.
- **Nesting `ui.Split`:** it has no maximum width and no collapse, and its
  divider's range cannot be set.

## References

- Shapes: [design: frostui](../design/frostui.md#the-frame-without-the-rail)
- Builds on: [ADR-0004](0004-extract-the-ui-kit-into-its-own-module.md)
