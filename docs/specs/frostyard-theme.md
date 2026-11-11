# Spec: The Frostyard themes

This contract governs the colors, metrics and font the kit draws with
([ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)):
the two themes `FrostyardTheme` returns, the `ThemeColors` that go with
them, and the contrast every pair keeps. The kit's components consume it;
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
desktop's preferences and `SoftSelection` leave as they are. For any other
theme it derives the colors the kit drew before ADR-0011: `Chrome` is
`Background` mixed 55% toward `Surface`, `Panel` 70%; a tone's fill is the
tone at 16% (`Border` at 60% for `Neutral`) and its line the tone at 35%;
the rest fall back to the theme's own fields.

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
- In both themes, with and without high contrast, these pairs hold 4.5:1
  (WCAG 2): `Text` and `TextMuted` on `Background`, `Chrome`, `Panel`,
  `PanelTo`, `Surface`, `SurfaceHover`, `SurfacePressed`, and on
  `SelectedFill` over `Background` and over `Chrome`; `SelectedText` on
  `SelectedFill` over both; `Text` on `Selection`; `AccentText` on
  `Accent`, `AccentHover` and `AccentPressed`; `OnGradient` on both
  gradient ends; `HoverText`, `Label`, `TagText` and each tone on
  `Background`, `Chrome` and `Panel`; each tone (`TagText` for `Neutral`)
  on its fill over `Panel`; `Text` on `Code`.
- `Focus` holds 3:1 on `Background`, `Chrome` and `Panel`.
- The weakest pairs, without high contrast: `OnGradient` at the gradient's
  end, 4.65:1 in both themes; in Frostyard light, `TextMuted` on
  `SurfacePressed` 4.64:1, the accent on `Chrome` 4.80:1, and the accent on
  its fill over `Panel` 4.97:1.
- `frostyard_test.go` checks every rule above.

## Derived artifacts

| Artifact | Derivation |
| --- | --- |
| The Frostyard Desktop design system | holds the same values for design work; `SurfaceHover`, `SurfacePressed`, `AccentHover`, `AccentPressed`, `Selection`, `Scrollbar`, `Fill(Neutral)` and the light theme's `AccentText` (white on the plain accent, where the design system's on-accent is the gradient's `OnGradient`) were chosen here and go back to it |

## References

- Motivated by: [ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)
- Fits into: [design: theming](../design/frostui.md#theming)
- Built by: [plan: the Frostyard look, phase 1](../plans/frostyard-look.md#phase-1--themes-palette-and-text-medium)
