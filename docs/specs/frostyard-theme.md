# Spec: The Frostyard themes

This contract governs the colors, metrics and font the kit draws with
([ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md),
[ADR-0013](../adr/0013-hold-text-to-7-to-1-under-high-contrast.md)):
the two themes `FrostyardTheme` returns, the `ThemeColors` that go with
them, the values the desktop's high contrast adds, and the contrast every
pair keeps. The kit's components consume it;
apps consume it through `Adapt`, `FrostyardTheme`, `ColorsOf`, `ToneFill`,
`ToneLine` and the text styles. [Design: theming](../design/frostui.md#theming)
shows where it fits.

## Interface

### `FrostyardTheme(dark bool) *ui.Theme`

| Field | Frostyard (`dark`) | Frostyard light |
| --- | --- | --- |
| `Background` | `#06111d` | `#f5f9fc` |
| `Surface` | `#0d2a41` | `#dfeaf2` |
| `SurfaceHover` | `#1a364c` (surface 6% toward text) | `#d2dde5` (same) |
| `SurfacePressed` | `#284358` (surface 12% toward text) | `#c5d0d8` (same) |
| `Border` | `rgba(162,222,255,0.18)` | `rgba(6,17,29,0.14)` |
| `Text` | `#edf8ff` | `#06111d` |
| `TextMuted` | `#b6d0df` | `#3f5a6b` |
| `Accent`, `InactiveAccent`, `Focus` | `#47b8ef` | `#0b6ea6` |
| `AccentHover` | `#68c5f2` (accent 20% toward text) | `#0a6091` (accent 15% toward text) |
| `AccentPressed` | `#3a97c5` (accent 20% toward background) | `#0a527d` (accent 30% toward text) |
| `AccentText` | `#02131f` | `#ffffff` |
| `Success` | `#7edafa` | `#076377` |
| `Warning` | `#e0b15c` | `#8a5300` |
| `Danger` | `#e07a72` | `#b0322b` |
| `Selection` | `rgba(71,184,239,0.35)` | `rgba(11,110,166,0.25)` |
| `Scrollbar` | `rgba(162,222,255,0.35)` | `rgba(6,17,29,0.32)` |
| `Inverse`, `InverseText` | unset: `Text` on `Background` turned over, as MyGo draws by default | same |
| `FocusRingWidth`, `FocusRingOffset` | 2, 1 | 2, 1 |
| `Radius`, `Spacing` | 4, 4 | 4, 4 |
| `FontSize` | MyGo's `LightTheme().FontSize` (14 on Linux and Windows, 13 on macOS) | same |
| `Font` | `FontFamily` ("Inter") | same |

### `ThemeColors`, from `ColorsOf(t)`

| Field | Frostyard | Frostyard light |
| --- | --- | --- |
| `Chrome` | `#081b2d` | `#e8f0f6` |
| `Panel` | `#0d263b` | `#ffffff` |
| `PanelFrom`, `PanelTo` | `#0f3048`, `#091c2d` | `#ffffff`, `#eef5fa` |
| `SelectedFill` | `rgba(71,184,239,0.08)` | `rgba(11,110,166,0.10)` |
| `SelectedText`, `HoverText` | `#aee9ff` | `#064a72` |
| `GradientFrom`, `GradientTo` | `#5ac9f7`, `#1584ca` | same |
| `OnGradient` | `#02131f` | `#02131f` |
| `Label` | `#82b6d0` | `#3b6680` |
| `TagText` | `#b9e9fb` | `#24485e` |
| `Code` | `#061826` | `#ffffff` |
| `LineStrong` | `#2c6380` | `#2c6380` |
| `Fill(Neutral)` | `rgba(162,222,255,0.10)` | `rgba(6,17,29,0.06)` |
| `Fill(Accent)` | `rgba(71,184,239,0.08)` | `rgba(11,110,166,0.08)` |
| `Fill(Success)` | `#0e3247` | `#ddf1f6` |
| `Fill(Warning)` | `rgba(51,41,20,0.2)` | `#f8ecd6` |
| `Fill(Danger)` | `#3a181e` | `#f9e2df` |
| `Line(Neutral)` | `rgba(166,223,250,0.22)` | `rgba(6,17,29,0.22)` |
| `Line(Accent)`, `Line(Success)` | `rgba(71,184,239,0.3)` | `rgba(11,110,166,0.4)`, `rgba(7,99,119,0.4)` |
| `Line(Warning)` | `rgba(224,177,92,0.33)` | `rgba(138,83,0,0.4)` |
| `Line(Danger)` | `rgba(224,122,114,0.33)` | `rgba(176,50,43,0.4)` |

`ColorsOf` finds these by the theme's `Dark` and `Background`, which the
desktop's preferences and `SoftSelection` leave as they are, and by its
`Text`, which tells the high-contrast colors below. For any other
theme it derives the colors the kit drew before ADR-0011: `Chrome` is
`Background` mixed 55% toward `Surface`, `Panel` 70%; a tone's fill is the
tone at 16% (`Border` at 60% for `Neutral`) and its line the tone at 35%;
the rest fall back to the theme's own fields.

### High contrast

With the desktop's high contrast on, MyGo's `WithPreferences` moves
`Border` toward `Text`, moves `TextMuted` 40% toward it, and makes `Focus`
and `Scrollbar` opaque; several text pairs still fall under 7:1 after that
mix, so `Adapt` sets the kit's own values over it
([ADR-0013](../adr/0013-hold-text-to-7-to-1-under-high-contrast.md)).

| Field | Frostyard | Frostyard light |
| --- | --- | --- |
| `Text` | `#ffffff` | `#000000` |
| `TextMuted` | `#cce0ec` | `#283d4c` |
| `Accent`, `InactiveAccent`, `Focus` | `#71c8f3` | `#08517b` |
| `AccentHover` | `#8dd3f5` (accent 20% toward text) | `#074569` (accent 15% toward text) |
| `AccentPressed` | `#66b6de` (accent 10% toward background) | `#063956` (accent 30% toward text) |
| `Success` | `#7edafa` (the brand's) | `#065465` |
| `Warning` | `#e0b15c` (the brand's) | `#704300` |
| `Danger` | `#e89e98` | `#862621` |
| `ColorsOf(t).GradientTo` | `#3baae3` | `#3baae3` |
| `ColorsOf(t).Label` | `#82b6d0` (the brand's) | `#2f5266` |

Every other field keeps what the brand's theme has, or what the desktop's
mix gives it: `Background`, `Surface`, `SurfaceHover`, `SurfacePressed`,
`Border`, `Selection`, `Scrollbar`, `AccentText`, the `ThemeColors`'
`Chrome`, `Panel`, both `Panel` ends, `SelectedFill`, `SelectedText`,
`HoverText`, `GradientFrom`, `OnGradient`, `TagText`, `Code`, `LineStrong`,
the fills and the lines.

Measured in both themes, the ratio of every pair of the list below:

| Pair | Frostyard, MyGo's mix | Frostyard, the kit's | Frostyard light, MyGo's mix | Frostyard light, the kit's |
| --- | ---: | ---: | ---: | ---: |
| text on background | 17.61:1 | 18.99:1 | 17.94:1 | 19.84:1 |
| muted text on background | 13.97:1 | 13.97:1 | 10.65:1 | 10.65:1 |
| text on chrome | 16.16:1 | 17.43:1 | 16.48:1 | 18.23:1 |
| muted text on chrome | 12.81:1 | 12.81:1 | 9.79:1 | 9.79:1 |
| text on panel | 14.34:1 | 15.47:1 | 18.99:1 | 21.00:1 |
| muted text on panel | 11.37:1 | 11.37:1 | 11.28:1 | 11.28:1 |
| text on panel gradient end | 16.02:1 | 17.28:1 | 17.25:1 | 19.08:1 |
| muted text on panel gradient end | 12.70:1 | 12.70:1 | 10.24:1 | 10.24:1 |
| text on surface | 13.67:1 | 14.74:1 | 15.55:1 | 17.19:1 |
| muted text on surface | 10.84:1 | 10.84:1 | 9.23:1 | 9.23:1 |
| text on surface hover | 11.62:1 | 12.53:1 | 13.76:1 | 15.21:1 |
| muted text on surface hover | 9.21:1 | 9.21:1 | 8.17:1 | 8.17:1 |
| text on surface pressed | 9.57:1 | 10.32:1 | 12.11:1 | 13.39:1 |
| muted text on surface pressed | 7.59:1 | 7.59:1 | 7.19:1 | 7.19:1 |
| text on selected fill over background | 15.73:1 | 16.97:1 | 15.60:1 | 17.25:1 |
| muted text on selected fill over background | 12.47:1 | 12.47:1 | 9.26:1 | 9.26:1 |
| selected text over background | 12.84:1 | 12.84:1 | 7.74:1 | 7.74:1 |
| text on selected fill over chrome | 14.19:1 | 15.30:1 | 14.41:1 | 15.93:1 |
| muted text on selected fill over chrome | 11.25:1 | 11.25:1 | 8.56:1 | 8.56:1 |
| selected text over chrome | 11.58:1 | 11.58:1 | 7.14:1 | 7.14:1 |
| text on selection | 8.65:1 | 9.33:1 | 12.53:1 | 13.85:1 |
| accent text on accent | 8.40:1 | 10.09:1 | 5.53:1 | 8.49:1 |
| accent text on accent hover | 9.73:1 | 11.45:1 | 6.79:1 | 10.17:1 |
| accent text on accent pressed | 5.75:1 | 8.35:1 | 8.35:1 | 12.16:1 |
| on-gradient text at its start | 9.98:1 | 9.98:1 | 9.98:1 | 9.98:1 |
| on-gradient text at its end | 4.65:1 | 7.21:1 | 4.65:1 | 7.21:1 |
| hover text on background | 14.38:1 | 14.38:1 | 8.90:1 | 8.90:1 |
| label on background | 8.64:1 | 8.64:1 | 5.84:1 | 7.87:1 |
| tag text on background | 14.57:1 | 14.57:1 | 9.17:1 | 9.17:1 |
| focus ring on background | 8.48:1 | 10.18:1 | 5.23:1 | 8.02:1 |
| accent tone on background | 8.48:1 | 10.18:1 | 5.23:1 | 8.02:1 |
| success tone on background | 12.03:1 | 12.03:1 | 6.48:1 | 8.05:1 |
| warning tone on background | 9.60:1 | 9.60:1 | 5.98:1 | 7.96:1 |
| danger tone on background | 6.51:1 | 8.84:1 | 5.94:1 | 8.58:1 |
| hover text on chrome | 13.19:1 | 13.19:1 | 8.17:1 | 8.17:1 |
| label on chrome | 7.93:1 | 7.93:1 | 5.36:1 | 7.23:1 |
| tag text on chrome | 13.37:1 | 13.37:1 | 8.42:1 | 8.42:1 |
| focus ring on chrome | 7.78:1 | 9.34:1 | 4.80:1 | 7.37:1 |
| accent tone on chrome | 7.78:1 | 9.34:1 | 4.80:1 | 7.37:1 |
| success tone on chrome | 11.03:1 | 11.03:1 | 5.96:1 | 7.40:1 |
| warning tone on chrome | 8.81:1 | 8.81:1 | 5.49:1 | 7.31:1 |
| danger tone on chrome | 5.97:1 | 8.11:1 | 5.46:1 | 7.88:1 |
| hover text on panel | 11.71:1 | 11.71:1 | 9.42:1 | 9.42:1 |
| label on panel | 7.04:1 | 7.04:1 | 6.18:1 | 8.33:1 |
| tag text on panel | 11.87:1 | 11.87:1 | 9.71:1 | 9.71:1 |
| focus ring on panel | 6.90:1 | 8.29:1 | 5.53:1 | 8.49:1 |
| accent tone on panel | 6.90:1 | 8.29:1 | 5.53:1 | 8.49:1 |
| success tone on panel | 9.79:1 | 9.79:1 | 6.86:1 | 8.52:1 |
| warning tone on panel | 7.82:1 | 7.82:1 | 6.33:1 | 8.43:1 |
| danger tone on panel | 5.30:1 | 7.20:1 | 6.29:1 | 9.08:1 |
| neutral tag text on its fill over panel | 9.22:1 | 9.22:1 | 8.58:1 | 8.58:1 |
| accent tone on its fill over panel | 6.00:1 | 7.21:1 | 4.97:1 | 7.63:1 |
| success tone on its fill over panel | 8.49:1 | 8.49:1 | 5.88:1 | 7.30:1 |
| warning tone on its fill over panel | 7.75:1 | 7.75:1 | 5.41:1 | 7.21:1 |
| danger tone on its fill over panel | 5.41:1 | 7.35:1 | 5.09:1 | 7.34:1 |
| text on code | 16.69:1 | 18.00:1 | 18.99:1 | 21.00:1 |

### Text

`TextSize(c, size)` is `size × theme FontSize ÷ LightTheme().FontSize`.
Every kit text style uses it: `Title` 26 bold, `Heading` 14.5 bold, `Strong`
13.5 bold, `Body` 13.5, `Muted` 12.5, `Caption` 11.5, `Mono` 12.5 in
"monospace", `Toned` 12.5. `Label` draws its text in capitals, 9.5,
weight 700, letter spacing 0.14 of its size, in "monospace" and the `Label`
color, and keeps the text as written as its accessible name.

The fonts are Inter 4.001, upright and italic variable files, registered
under `FontFamily` on the first `Adapt`.

## Rules

- `Adapt` sets `FrostyardTheme(t.Dark).WithPreferences(p)`, where `t` is the
  theme MyGo picked and `p` the desktop's preferences with `Accent` cleared,
  unless the theme is already a Frostyard theme.
- The desktop's high contrast and text scale apply to both themes; its
  accent never does.
- In both themes, without high contrast, and with MyGo's mix alone (the
  desktop's high contrast), these pairs hold 4.5:1 (WCAG 2): `Text` and
  `TextMuted` on `Background`, `Chrome`, `Panel`, `PanelTo`, `Surface`,
  `SurfaceHover`, `SurfacePressed`, and on `SelectedFill` over `Background`
  and over `Chrome`; `SelectedText` on `SelectedFill` over both; `Text` on
  `Selection`; `AccentText` on `Accent`, `AccentHover` and `AccentPressed`;
  `OnGradient` on both gradient ends; `HoverText`, `Label`, `TagText` and
  each tone on `Background`, `Chrome` and `Panel`; each tone (`TagText` for
  `Neutral`) on its fill over `Panel`; `Text` on `Code`.
- With high contrast on, `Adapt` sets the kit's high-contrast values, also
  over a Frostyard theme an app set itself, and every text pair of that
  list holds 7:1 (WCAG 2 AAA).
- `Focus` holds 3:1 on `Background`, `Chrome` and `Panel`, in every case.
- The weakest pairs, without high contrast: `OnGradient` at the gradient's
  end, 4.65:1 in both themes; in Frostyard light, `TextMuted` on
  `SurfacePressed` 4.64:1, the accent on `Chrome` 4.80:1, and the accent on
  its fill over `Panel` 4.97:1.
- The weakest pairs, with the kit's high contrast: in Frostyard, `Label`
  on `Panel` 7.04:1, which the high-contrast values leave as the brand's;
  in Frostyard light, `SelectedText` on `SelectedFill` over `Chrome`
  7.14:1.
- `frostyard_test.go` checks every rule above, and
  `TestSpecListsHighContrastRatios` holds the table above to the values
  the kit sets.

## Derived artifacts

| Artifact | Derivation |
| --- | --- |
| The Frostyard Desktop design system | holds the same values for design work; `SurfaceHover`, `SurfacePressed`, `AccentHover`, `AccentPressed`, `Selection`, `Scrollbar`, `Fill(Neutral)` and the light theme's `AccentText` (white on the plain accent, where the design system's on-accent is the gradient's `OnGradient`) were chosen here and go back to it |

## References

- Motivated by: [ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md),
  [ADR-0013](../adr/0013-hold-text-to-7-to-1-under-high-contrast.md)
- Fits into: [design: theming](../design/frostui.md#theming)
- Built by: [plan: the Frostyard look, phase 1](../plans/frostyard-look.md#phase-1--themes-palette-and-text-done)
