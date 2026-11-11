# 0009 — The top bar's end belongs to the window, and the palette field turns compact

- **Status:** Accepted
- **Date:** 2026-11-11

## Context

[ADR-0006](0006-a-frame-without-the-rail.md) puts each pane's top-bar
segment in the pane's column, and the frame's end of the bar (the palette
field, the app's `End` items, the console's button and the main menu) in
the segment of the last pane shown, before the window controls. That
segment clips.

A narrow last pane has no room for the end. In a window 1,200 DIPs wide
with 110 DIPs of window controls at the right, a right pane 200 DIPs wide
leaves 66 DIPs between its padding and the controls. The console's button
and the main menu alone take 108. The field went under the window controls
and the main menu out of the window, where no one could click it
([frostyard/frostui#11](https://github.com/frostyard/frostui/issues/11), B1
of the cross-review of #4). Frostnotes has such a pane: a right sidebar 200
to 600 DIPs wide. ADR-0006 requires a reachable main menu in every app.

The bar not split into segments, and `Shell`'s bar, overflowed the same
way in narrow windows. Their trailing side could shrink below its controls,
and their field did not give way.

## Decision

- **The end of a bar of segments is the window's, not the last pane's.**
  `PaneShell` draws it once, above the segments and their separators, at its
  own width, before the window controls. The part of each segment under it
  is padding, so the panes' `Bar` items give way, the last pane's first. It
  reaches over the segments before the last when the last is too narrow.
  The segments stay in the panes' columns.
- **The trailing controls never shrink.** `End`, the console's button and
  the main menu keep their size in every bar, and so do the sidebar toggle
  and `Shell`'s brand. Only the palette field and the apps' items (`Bar`,
  `Start`) give way.
- **The palette field turns compact.** It shrinks to a least width: 160
  DIPs in a bar of segments, where it fits the room in the last segment, and
  180 in the other bars. With less room it is drawn as an icon button
  with the field's name ("Search or run a command"), its tooltip ("Command
  palette (Ctrl K)") and its key for assistive technology. Clicking it
  opens the palette, and Ctrl K still does.
- **Right to left, the end is at the left,** by the window controls there.

## Consequences

- The main menu, the console's button and `End` stay inside the window,
  clear of its controls, at any pane width. In a narrow window the apps'
  bar items and the field give way instead.
- The field stays in the same place in the shared order in every app. Only
  its width changes, and below the least width its look.
- The bar reads its trailing controls' and its end's widths from the last
  frame and asks for one more frame when they change, as `Panes` does with
  its room.
- A pane's `Bar` can lose its end, or all of it, under the bar's end; the
  separators do not show in the bar there.
- The end comes after the panes in the order Tab moves the focus and
  assistive technology reads. Before, it came before the last pane's
  content.

## Alternatives considered

- **`Shrink(0)` on the trailing controls inside the last segment:** the
  segment clips, and a narrow pane has less room than the controls take, so
  they still leave the window.
- **Moving the palette field to the widest segment:** the field would sit
  in a different place in each app and jump between segments as panes are
  resized, against "the same palette field in every app" (ADR-0006).
- **Raising the last pane's minimum to fit the end:** it overrides the
  app's `Min` and `Max`, and Frostnotes' sidebar could not be 200 DIPs.
- **Letting the last column draw outside its clip:** the separator beside
  it is above the column, and raising the column would cover half of the
  separator's grip.

## References

- Shapes: [design: frostui](../design/frostui.md#the-frame-without-the-rail), [the command-center shell](../design/frostui.md#the-command-center-shell)
- Builds on: [ADR-0006](0006-a-frame-without-the-rail.md)
