# 0011 — Draw the frostyard brand, in a dark and a light theme

- **Status:** Accepted
- **Date:** 2026-11-11

## Context

The kit has no palette of its own. `Adapt` takes MyGo's `LightTheme` or
`DarkTheme`, which follow the desktop's accent, light or dark mode,
contrast and text size, and the kit derives `Chrome` and `Panel` from them
by mixing `Background` toward `Surface`
([design: theming](../design/frostui.md#theming)). The result is zinc grays,
the desktop's accent (blue when it has none), corners of 6 to 12 DIPs, the
system's UI font and a shadow under toasts: ChairLift, Frostmail and
Frostnotes draw MyGo's default look.

Several pairs the kit draws with MyGo's defaults miss WCAG 2 contrast: pill
text on its fill in the light theme (2.52:1 to 3.88:1), white on the dark
accent (3.68:1), muted text on a chosen row in light (3.73:1), and the light
focus ring (2.33:1) (frostyard/frostui#16). A rail badge in the `Accent`
tone vanishes on the selected item, accent on accent (frostyard/frostui#15).

frostyard has a brand, kept in frostyard/core as the `frostyard-design`
skill: near-black blue inks, ice and sky accents, Inter, 1px hairlines in
place of shadows, nearly square corners and uppercase mono labels. It is
dark only. Its Pilothouse kit applies it to an admin console with a
sidebar, cards, badges and tables.

The Frostyard Desktop design system maps the brand onto the kit's roles
(background, chrome, panel, surface, text, accent, selection, tones) and
adds a light theme. Every text pair in both themes holds 4.5:1, the focus
ring 3:1.

MyGo v0.9.0 has what the brand needs: `RegisterFont` for an embedded font,
`LinearGradient` fills, per-side `BorderWidth`, `LetterSpacing`, and
`Theme.WithPreferences`, which applies the desktop's high contrast and
text size, and its accent only when `Preferences.Accent` is set.

## Decision

- The kit draws two themes of its own, **Frostyard** (dark) and
  **Frostyard light**, and picks one from the desktop's dark or light
  preference, or from an app's own appearance setting.
- The accent is the brand's: sky in Frostyard, a deep sky in Frostyard
  light. The kit no longer applies the desktop's accent. It still applies
  the desktop's high contrast and text size, through `WithPreferences` with
  no accent.
- Roles that MyGo's `Theme` has no field for (chrome, panel and its
  gradient, selected fill and text, hover text, each tone's fill and line,
  the label and code colors) are the kit's, beside the theme.
- Text is set in Inter, embedded in the kit under the SIL Open Font License
  1.1, at sizes relative to the theme's `FontSize`, so that the desktop's
  text size applies to every kit style. Section titles and badges are
  uppercase monospace labels with letter spacing.
- Corners are nearly square: buttons 4 DIPs; pills, fields and banners 2;
  cards, the palette and toasts 0. Nothing casts a shadow: 1px hairlines
  separate regions.
- Pills and badges are 1px outlines in their tone's line color around the
  label, with no fill. The primary button is the brand's gradient from
  `#5ac9f7` to `#1584ca` with near-black text, in both themes.
- Every value lives in a spec that lands with the code, and a test computes
  the contrast of each text pair (4.5:1) and of the focus ring and
  meaningful marks (3:1) in both themes, high contrast included.
- Fonts under the SIL Open Font License 1.1 may be embedded, alongside the
  MIT, ISC and Apache-2.0 icons
  [ADR-0007](0007-embed-a-lucide-icon-set-as-the-last-fallback.md) allows,
  with the license beside the files and in `NOTICE`.

## Consequences

- ChairLift, Frostmail and Frostnotes share one identity with frostyard's
  site and images, in light and dark, and the contrast failures of MyGo's
  defaults go away for every app that uses the kit.
- The apps no longer take the desktop's accent color. A user who set one
  sees the brand's accent in these apps and theirs in the rest of the
  desktop.
- Kit text styles follow the desktop's text size, which they did not: a
  layout that fit at fixed sizes may need room at larger ones.
- Every app binary grows by the size of the embedded Inter font files.
- MyGo's own widgets used directly (`ui.Button`, `ui.SidebarItem`) draw
  with the theme's colors and radius but not every brand detail: a
  secondary button gets the brand's `surface` face, not the ghost outline
  the kit's button helpers draw. Apps move such calls to kit helpers to get
  the full look.
- Each app's literal font sizes and radii stop matching the kit until they
  move to kit styles. Frostmail has the most (148 `FontSize` literals and 27
  radii in its desktop packages).
- The examples' snapshots and the apps' screenshots change.
- The kit's look is no longer the desktop's: the
  [theming section](../design/frostui.md#theming) of the design doc is
  rewritten when the themes land.

## Alternatives considered

- **Keep following the desktop:** the apps stay generic, and the contrast
  failures remain MyGo's to fix.
- **Adopt a MyGo UI library (Breeze UI, Lobe MyGo, Onevra):** each is weeks
  old with one main maintainer; Lobe MyGo carries another company's design
  language; none has the kit's shell, palette, job queue or console, which
  would remain the kit's.
- **Frostyard dark only:** the apps follow the desktop's light or dark
  preference, and a dark-only look would ignore it.
- **The brand's accent only when the desktop has none:** the apps' identity
  would change from one machine to the next, and both accents would need
  testing for contrast.

## References

- Shapes: [design: theming](../design/frostui.md#theming),
  [spec: the Frostyard themes](../specs/frostyard-theme.md),
  [plan: the Frostyard look](../plans/frostyard-look.md)
- Builds on: [ADR-0007](0007-embed-a-lucide-icon-set-as-the-last-fallback.md)
  (embedded assets and their licenses)
- Fixes: frostyard/frostui#15, frostyard/frostui#16
- Brand: frostyard/core, `.agents/skills/frostyard-design/`
