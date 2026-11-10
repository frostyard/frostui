# 0007 — Embed a Lucide icon set as the last fallback

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

[ADR-0005](0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)
made an app off Linux embed every icon it draws, the kit's own included, or
show empty boxes, and rejected an icon set in the kit because it would set a
license and a look for every app. Frostmail now ships on Windows, and it and
Frostnotes draw the same names. Each app embedding the same files, and the
kit's 20 names besides, is repeated work with the same license question each
time. A desktop theme can also lack a name.

## Decision

- The kit embeds a set of symbolic icons from **Lucide**, under the ISC
  license (with Feather's MIT notice for the icons derived from it). The
  React apps of Frostmail, Frostnotes and lodge already use Lucide, so the
  native apps keep their look off Linux.
- `kitIconNames` (`iconset.go`) maps freedesktop names to Lucide's. `Icons.Get`
  looks in the app's embedded icons, then the desktop's themes, then this
  set: it is the only set off Linux and the fallback on Linux.
- Only mapped files are in `icons/`
  (`TestKitIconSetHoldsOnlyMappedIcons`), and every name the kit draws is
  mapped (`TestKitDrawsOnlyMappedIcons`).
- The upstream license stays beside the icons in `icons/LICENSE`, and `NOTICE`
  attributes them. Only MIT, ISC or Apache-2.0 icons may be added.

This supersedes ADR-0005's fallback for `NewIcons` off Linux ("it finds the
app's embedded icons alone") and its rejection of an icon set. The rest of
ADR-0005 stands.

## Consequences

- An app off Linux shows icons without embedding any, and an app's embedded
  icons still win.
- On Linux a name no theme has is a Lucide icon beside the theme's, instead
  of an empty box.
- Every binary that imports the kit carries about 32 KB of SVG.
- A new name needs a file and a map line; the guards catch a name the kit
  draws without one.

## Alternatives considered

- **Each app embeds its own:** the status quo, repeated in each app.
- **Phosphor (MIT):** a good set, but no frostyard app uses it.
- **Adwaita's symbolic icons:** under LGPL or CC-BY-SA, which the kit does
  not take.

## References

- Shapes: [design: frostui](../design/frostui.md#the-kits-icon-set)
- Supersedes in part: [ADR-0005](0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)
