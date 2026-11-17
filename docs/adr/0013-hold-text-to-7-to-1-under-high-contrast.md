# 0013 — Hold text to 7:1 under the desktop's high contrast

- **Status:** Accepted
- **Date:** 2026-11-16

## Context

[ADR-0011](0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)
applies the desktop's high contrast through MyGo's
`Theme.WithPreferences`, whose mix moves `Border` toward `Text`,
`TextMuted` 40% toward it and makes the focus ring and the scrollbar
opaque. The kit's measured pairs then still fall under 7:1 (WCAG 2 AAA):

- Frostyard (dark): the accent text on the pressed accent 5.75:1, the
  gradient's text at its end 4.65:1, the danger tone on `Background`,
  `Chrome` and `Panel` 5.30:1 to 6.51:1, the accent on `Panel` 6.90:1, the
  accent on its fill 6.00:1, the danger tone on its fill 5.41:1.
- Frostyard light: the accent text on the accent 5.53:1 and on its hover
  face 6.79:1, the gradient's text at its end 4.65:1, `Label` 5.36:1 to
  6.18:1, and every tone (accent, success, warning, danger) on
  `Background`, `Chrome`, `Panel` and on its fill, 4.80:1 to 6.86:1.

A user who turns high contrast on asks for more than AA: 4.5:1 is what
every pair already holds without it. The kit cannot reach 7:1 by changing
`Adapt`'s call alone: the values MyGo's mix produces are MyGo's.

## Decision

- Under the desktop's high contrast, the kit applies its own values after
  MyGo's mix, in `Adapt`: `Text`, `TextMuted`, the accent and its hover and
  pressed faces, `InactiveAccent`, `Focus`, the three tones, and two
  `ThemeColors` fields, the label in Frostyard light and the gradient's end
  in both themes. The spec lists them.
- `Adapt` also applies them over a Frostyard theme an app set itself, as
  the Pocket client does.
- `ColorsOf` still finds a Frostyard theme by its `Dark` and `Background`,
  and tells the high-contrast theme colors by the theme's `Text`, which
  neither the desktop's preferences nor `SoftSelection` change; `ToneFill`,
  `ToneLine` and `SoftSelection` follow from it.
- With them, every text pair of the spec's list holds 7:1 in both themes;
  the focus ring keeps its 3:1.
- Without high contrast nothing changes, and no exported declaration
  changes.

## Consequences

- The kit's text is readable at AAA for a user who asks for it, in both
  themes and in every app that calls `Adapt`, without a change in the apps.
- Under high contrast the accent and the tones are no longer the brand's
  values: a lighter accent and danger in Frostyard, a deeper accent and
  deeper tones in Frostyard light. The look differs from the brand there,
  on purpose.
- Under high contrast, `Adapt` replaces the text, accent and tone fields of
  a Frostyard theme an app set itself. An app that wants its own values
  there sets its theme in a scope after `Adapt`, as it already must for any
  theme of its own.
- An app that never calls `Adapt` gets only MyGo's mix, 4.5:1, so the kit's
  7:1 depends on that call, as the whole brand does.
- Two `ThemeColors` values per theme are new colors to keep in step with the
  spec; `frostyard_test.go` holds the measured ratios in the spec.

## Alternatives considered

- **Keep MyGo's mix alone:** the pairs above stay at 4.65:1 to 6.90:1,
  under the 7:1 the user's setting asks for.
- **A new exported API, such as `FrostyardHighContrastTheme(dark bool)`:**
  it changes the exported API and every consumer must adopt it, while
  `Adapt` can apply the values for them.
- **Change `Background` as the marker of the high-contrast theme:** it
  breaks the lookup by `Background` that `ColorsOf` and `SoftSelection`
  rely on, and it changes a color high contrast has no reason to change.

## References

- Shapes: [spec: the Frostyard themes](../specs/frostyard-theme.md#high-contrast),
  [design: theming](../design/frostui.md#theming),
  [plan: the Frostyard look](../plans/frostyard-look.md#later--ideas)
- Builds on: [ADR-0011](0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)
- Fixes: frostyard/frostui#32
